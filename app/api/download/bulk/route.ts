import { NextRequest } from 'next/server';
import { File } from '@/models/File';
import { Folder } from '@/models/Folder';
import { connectDB } from '@/lib/db';
import { downloadRequestSchema } from '@/lib/validations/file';
import { apiRateLimiter, createRateLimitMiddleware } from '@/lib/rate-limit';
import { getServerSession } from 'next-auth';
import { authOptions } from '@/lib/auth';
import { getClientIP, formatHeaders, getStorageProvider } from '@/lib/api-utils';
import { canAccessResource } from '@/lib/permissions';
import type { BaseResponse, User } from '@/types';
import archiver from 'archiver';
import fs from 'fs/promises';
import path from 'path';
import { Writable } from 'stream';
// POST /api/download/bulk - Create bulk download
export async function POST(request: NextRequest) {
  try {
    const session = await getServerSession(authOptions);
    if (!session?.user) {
      return Response.json({
        success: false,
        error: 'Authentication required'
      }, { status: 401 });
    }

    const user = session.user as User;
    const clientIP = getClientIP(request);
    const rateLimitResult = createRateLimitMiddleware(apiRateLimiter)(clientIP);
    
    await connectDB();
    const body = await request.json();
    
    // Validate request data
    const { fileIds, format = 'zip' } = downloadRequestSchema.parse(body);

    // Find files and verify access
    const filesQuery = await File.find({
      _id: { $in: fileIds },
      isTrashed: false,
      processingStatus: 'completed'
    })
    .populate('folder', 'name path')
    .lean();

    const files = filesQuery as any[];

    if (files.length === 0) {
      return Response.json({
        success: false,
        error: 'No files found'
      }, { status: 404 });
    }

    // Check access permissions for each file
    const accessibleFiles: any[] = [];
    for (const file of files) {
      if (canAccessResource(user, file as any, 'read')) {
        accessibleFiles.push(file);
      }
    }

    if (accessibleFiles.length === 0) {
      return Response.json({
        success: false,
        error: 'No accessible files found'
      }, { status: 403 });
    }

    // Calculate total size
    const totalSize = accessibleFiles.reduce((sum, file) => sum + file.size, 0);
    const maxBulkSize = 5 * 1024 * 1024 * 1024; // 5GB limit

    if (totalSize > maxBulkSize) {
      return Response.json({
        success: false,
        error: `Total size exceeds bulk download limit of ${formatFileSize(maxBulkSize)}`
      }, { status: 413 });
    }

    // Create archive based on format
    let archive: archiver.Archiver;
    let mimeType: string;
    let fileExtension: string;

    switch (format) {
      case 'zip':
        archive = archiver('zip', { zlib: { level: 6 } });
        mimeType = 'application/zip';
        fileExtension = 'zip';
        break;
      case 'tar':
        archive = archiver('tar');
        mimeType = 'application/x-tar';
        fileExtension = 'tar';
        break;
      default:
        return Response.json({
          success: false,
          error: 'Unsupported archive format'
        }, { status: 400 });
    }

    // Generate archive filename
    const timestamp = new Date().toISOString().slice(0, 19).replace(/[:.]/g, '-');
    const archiveName = `bulk-download-${timestamp}.${fileExtension}`;

    // Create readable stream for response
    const { readable, writable } = new TransformStream();
    const writer = writable.getWriter();

    // Handle archive events
    archive.on('error', (err) => {
      console.error('Archive error:', err);
      writer.abort(err);
    });

    archive.on('end', () => {
      writer.close();
    });



    // Pipe archive to response stream
    archive.pipe(new Writable({
      write(chunk, encoding, callback) {
        writer.write(chunk);
        callback();
      },
      final(callback) {
        writer.close();
        callback();
      },
      destroy(error, callback) {
        writer.abort(error);
        callback(error);
      }
    }));

    // Add files to archive
    const storageConfig = await getStorageProvider();
    
    try {
      for (const file of accessibleFiles) {
        if (storageConfig.provider === 'local') {
          // Local file system
          const uploadDir = process.env.UPLOAD_DIR || './uploads';
          const filePath = path.join(uploadDir, file.key);
          
          try {
            await fs.access(filePath);
            
            // Generate safe filename for archive
            const safeName = generateArchiveFileName(file as any);
            archive.file(filePath, { name: safeName });
            
          } catch (fsError) {
            console.warn(`File not accessible: ${file.originalName}`, fsError);
            // Add a placeholder file indicating the error
            archive.append(
              `Error: File "${file.originalName}" could not be accessed`,
              { name: `ERROR-${file.originalName}.txt` }
            );
          }
        } else {
          // For cloud storage, we'd need to download first
          // This is a simplified version - in production, implement streaming from cloud storage
          console.warn('Cloud storage bulk download not fully implemented');
          archive.append(
            `Note: "${file.originalName}" requires cloud storage download`,
            { name: `NOTE-${file.originalName}.txt` }
          );
        }
      }

      // Finalize archive
      archive.finalize();

      // Update download statistics
      await File.updateMany(
        { _id: { $in: accessibleFiles.map(f => f._id) } },
        {
          $inc: { downloadCount: 1 },
          $set: { lastAccessedAt: new Date() }
        }
      );

    } catch (archiveError) {
      console.error('Archive creation error:', archiveError);
      return Response.json({
        success: false,
        error: 'Failed to create archive'
      }, { status: 500 });
    }

    // Prepare response headers
    const headers = new Headers();
    headers.set('Content-Type', mimeType);
    headers.set('Content-Disposition', `attachment; filename="${encodeURIComponent(archiveName)}"`);
    headers.set('Cache-Control', 'no-cache, no-store, must-revalidate');
    headers.set('X-Total-Files', accessibleFiles.length.toString());
    headers.set('X-Total-Size', totalSize.toString());

    // Rate limit headers
    Object.entries(formatHeaders(rateLimitResult.headers)).forEach(([key, value]) => {
      headers.set(key, value);
    });

    return new Response(readable, {
      status: 200,
      headers
    });

  } catch (error: any) {
    console.error('Bulk download error:', error);
    
    if (error.name === 'ZodError') {
      return Response.json({
        success: false,
        error: 'Invalid request data',
        details: error.errors
      }, { status: 400 });
    }

    return Response.json({
      success: false,
      error: 'Failed to create bulk download'
    }, { status: 500 });
  }
}

// GET /api/download/bulk - Get bulk download status (for async downloads)
export async function GET(request: NextRequest) {
  try {
    const session = await getServerSession(authOptions);
    if (!session?.user) {
      return Response.json({
        success: false,
        error: 'Authentication required'
      }, { status: 401 });
    }

    const user = session.user as User;
    const clientIP = getClientIP(request);
    const rateLimitResult = createRateLimitMiddleware(apiRateLimiter)(clientIP);
    
    const url = new URL(request.url);
    const downloadId = url.searchParams.get('downloadId');

    if (!downloadId) {
      return Response.json({
        success: false,
        error: 'Download ID is required'
      }, { status: 400 });
    }

    // In a real implementation, you'd track download progress in Redis/database
    // For now, return a simple status
    const response: BaseResponse<{
      downloadId: string;
      status: 'pending' | 'processing' | 'completed' | 'failed';
      progress: number;
      downloadUrl?: string;
    }> = {
      success: true,
      data: {
        downloadId,
        status: 'completed',
        progress: 100,
        downloadUrl: `/api/download/bulk/${downloadId}`
      }
    };

    return Response.json(response, {
      headers: formatHeaders(rateLimitResult.headers)
    });

  } catch (error: any) {
    console.error('Get bulk download status error:', error);
    
    return Response.json({
      success: false,
      error: 'Failed to get download status'
    }, { status: 500 });
  }
}

// Helper function to generate safe archive filenames
function generateArchiveFileName(file: any): string {
  const folder = file.folder;
  const folderPath = folder?.path ? folder.path.replace(/^\//, '') : '';
  
  if (folderPath) {
    return `${folderPath}/${file.originalName}`;
  }
  
  return file.originalName;
}

// Helper function to format file size
function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 Bytes';
  
  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}