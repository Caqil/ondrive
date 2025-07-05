import { NextRequest } from 'next/server';
import { File } from '@/models/File';
import { Folder } from '@/models/Folder';
import { connectDB } from '@/lib/db';
import { apiRateLimiter, createRateLimitMiddleware } from '@/lib/rate-limit';
import { getServerSession } from 'next-auth';
import { authOptions } from '@/lib/auth';
import { getClientIP, formatHeaders, getStorageProvider } from '@/lib/api-utils';
import { canAccessResource } from '@/lib/permissions';
import type { BaseResponse, User } from '@/types';
import archiver from 'archiver';
import fs from 'fs/promises';
import path from 'path';

// POST /api/download/zip - Create ZIP archive from folders/files
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
    
    const { 
      fileIds = [], 
      folderIds = [], 
      name,
      includeSubfolders = true 
    } = body;

    // Validate that at least one item is selected
    if (fileIds.length === 0 && folderIds.length === 0) {
      return Response.json({
        success: false,
        error: 'At least one file or folder must be selected'
      }, { status: 400 });
    }

    // Collect all files to include
    const filesToInclude: any[] = [];
    let totalSize = 0;

    // Add individual files
    if (fileIds.length > 0) {
      const files = await File.find({
        _id: { $in: fileIds },
        isTrashed: false,
        processingStatus: 'completed'
      })
      .populate('folder', 'name path')
      .lean();

      for (const file of files) {
        if (canAccessResource(user, file as any, 'read')) {
          filesToInclude.push({
            ...file,
            archivePath: file.originalName,
            type: 'file'
          });
          totalSize += file.size;
        }
      }
    }

    // Add files from folders
    if (folderIds.length > 0) {
      for (const folderId of folderIds) {
        // Verify folder access
        const folderQuery = await Folder.findOne({
          _id: folderId,
          isTrashed: false
        });

        const folder = folderQuery as any;

        if (!folder || !canAccessResource(user, folder, 'read')) {
          continue;
        }

        // Get all files in folder (and subfolders if enabled)
        const folderFiles = await getFolderFiles(folderId, includeSubfolders, user);
        
        for (const file of folderFiles) {
          filesToInclude.push({
            ...file,
            archivePath: file.relativePath,
            type: 'file'
          });
          totalSize += file.size;
        }
      }
    }

    if (filesToInclude.length === 0) {
      return Response.json({
        success: false,
        error: 'No accessible files found'
      }, { status: 404 });
    }

    // Check size limit
    const maxZipSize = 10 * 1024 * 1024 * 1024; // 10GB limit
    if (totalSize > maxZipSize) {
      return Response.json({
        success: false,
        error: `Total size exceeds ZIP download limit of ${formatFileSize(maxZipSize)}`
      }, { status: 413 });
    }

    // Generate ZIP filename
    const timestamp = new Date().toISOString().slice(0, 19).replace(/[:.]/g, '-');
    const zipName = name ? `${name}.zip` : `download-${timestamp}.zip`;

    // Create ZIP archive
    const archive = archiver('zip', {
      zlib: { level: 6 } // Compression level
    });

    // Create readable stream for response
    const { readable, writable } = new TransformStream();
    const writer = writable.getWriter();

    // Handle archive events
    let archiveError: Error | null = null;

    archive.on('error', (err) => {
      console.error('Archive error:', err);
      archiveError = err;
      writer.abort(err);
    });

    archive.on('end', () => {
      if (!archiveError) {
        writer.close();
      }
    });

    archive.on('warning', (err) => {
      console.warn('Archive warning:', err);
    });

    // Pipe archive to response stream
    const { Writable } = require('stream');
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
    let filesAdded = 0;
    
    try {
      for (const fileItem of filesToInclude) {
        if (storageConfig.provider === 'local') {
          const uploadDir = process.env.UPLOAD_DIR || './uploads';
          const filePath = path.join(uploadDir, fileItem.key);
          
          try {
            await fs.access(filePath);
            
            // Use the archive path to maintain folder structure
            archive.file(filePath, { name: fileItem.archivePath });
            filesAdded++;
            
          } catch (fsError) {
            console.warn(`File not accessible: ${fileItem.originalName}`, fsError);
            
            // Add error placeholder
            archive.append(
              `Error: File "${fileItem.originalName}" could not be accessed\nPath: ${fileItem.archivePath}\nSize: ${formatFileSize(fileItem.size)}`,
              { name: `_ERRORS/${fileItem.archivePath}.error.txt` }
            );
          }
        } else {
          // For cloud storage, add placeholder (implement actual cloud download in production)
          archive.append(
            `Cloud Storage File\nName: ${fileItem.originalName}\nSize: ${formatFileSize(fileItem.size)}\nNote: Download from cloud storage not implemented in this demo`,
            { name: `_CLOUD/${fileItem.archivePath}.info.txt` }
          );
          filesAdded++;
        }
      }

      // Add manifest file
      const manifest = {
        created: new Date().toISOString(),
        totalFiles: filesToInclude.length,
        filesAdded,
        totalSize,
        user: {
          id: user._id,
          name: user.name,
          email: user.email
        },
        files: filesToInclude.map(f => ({
          name: f.originalName,
          path: f.archivePath,
          size: f.size,
          type: f.mimeType,
          modified: f.updatedAt
        }))
      };

      archive.append(JSON.stringify(manifest, null, 2), { name: 'manifest.json' });

      // Finalize archive
      await archive.finalize();

      // Update download statistics
      const fileIds = filesToInclude.map(f => f._id);
      if (fileIds.length > 0) {
        await File.updateMany(
          { _id: { $in: fileIds } },
          {
            $inc: { downloadCount: 1 },
            $set: { lastAccessedAt: new Date() }
          }
        );
      }

    } catch (archiveError) {
      console.error('Archive creation error:', archiveError);
      return Response.json({
        success: false,
        error: 'Failed to create ZIP archive'
      }, { status: 500 });
    }

    // Prepare response headers
    const headers = new Headers();
    headers.set('Content-Type', 'application/zip');
    headers.set('Content-Disposition', `attachment; filename="${encodeURIComponent(zipName)}"`);
    headers.set('Cache-Control', 'no-cache, no-store, must-revalidate');
    headers.set('X-Total-Files', filesToInclude.length.toString());
    headers.set('X-Files-Added', filesAdded.toString());
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
    console.error('ZIP download error:', error);
    
    return Response.json({
      success: false,
      error: 'Failed to create ZIP download'
    }, { status: 500 });
  }
}

// GET /api/download/zip - Get ZIP download info
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
    const fileIds = url.searchParams.get('fileIds')?.split(',') || [];
    const folderIds = url.searchParams.get('folderIds')?.split(',') || [];

    if (fileIds.length === 0 && folderIds.length === 0) {
      return Response.json({
        success: false,
        error: 'At least one file or folder ID must be provided'
      }, { status: 400 });
    }

    await connectDB();

    // Calculate estimated size and file count
    let totalSize = 0;
    let totalFiles = 0;

    // Count files
    if (fileIds.length > 0) {
      const files = await File.find({
        _id: { $in: fileIds },
        isTrashed: false,
        processingStatus: 'completed'
      });

      for (const file of files) {
        if (canAccessResource(user, file as any, 'read')) {
          totalSize += file.size;
          totalFiles++;
        }
      }
    }

    // Count folder contents
    if (folderIds.length > 0) {
      for (const folderId of folderIds) {
        const folderQuery = await Folder.findOne({
          _id: folderId,
          isTrashed: false
        });

        const folder = folderQuery as any;

        if (folder && canAccessResource(user, folder, 'read')) {
          const folderStats = await getFolderStats(folderId, user);
          totalSize += folderStats.size;
          totalFiles += folderStats.files;
        }
      }
    }

    const response: BaseResponse<{
      estimatedSize: number;
      estimatedSizeFormatted: string;
      totalFiles: number;
      canDownload: boolean;
      maxSizeLimit: number;
      maxSizeLimitFormatted: string;
    }> = {
      success: true,
      data: {
        estimatedSize: totalSize,
        estimatedSizeFormatted: formatFileSize(totalSize),
        totalFiles,
        canDownload: totalSize <= 10 * 1024 * 1024 * 1024, // 10GB limit
        maxSizeLimit: 10 * 1024 * 1024 * 1024,
        maxSizeLimitFormatted: formatFileSize(10 * 1024 * 1024 * 1024)
      }
    };

    return Response.json(response, {
      headers: formatHeaders(rateLimitResult.headers)
    });

  } catch (error: any) {
    console.error('Get ZIP info error:', error);
    
    return Response.json({
      success: false,
      error: 'Failed to get ZIP info'
    }, { status: 500 });
  }
}

// Helper function to get all files in a folder recursively
interface FileType {
  [key: string]: any;
}

async function getFolderFiles(folderId: string, includeSubfolders: boolean, user: User): Promise<FileType[]> {
  const files: FileType[] = [];
  
  // Get files directly in this folder
  const directFiles = await File.find({
    folder: folderId,
    isTrashed: false,
    processingStatus: 'completed'
  })
  .populate('folder', 'name path')
  .lean();

  for (const file of directFiles) {
    if (canAccessResource(user, file as any, 'read')) {
      const folder = file.folder as any;
      const relativePath = folder?.name ? `${folder.name}/${file.originalName}` : file.originalName;
      
      files.push({
        ...file,
        relativePath
      });
    }
  }

  // Get subfolders if enabled
  if (includeSubfolders) {
    const subfolders = await Folder.find({
      parent: folderId,
      isTrashed: false
    });

    for (const subfolder of subfolders) {
      if (canAccessResource(user, subfolder as any, 'read')) {
        const subfolderFiles = await getFolderFiles(subfolder._id.toString(), true, user);
        
        for (const subFile of subfolderFiles) {
          files.push({
            ...subFile,
            relativePath: `${subfolder.name}/${subFile.relativePath}`
          });
        }
      }
    }
  }

  return files;
}

// Helper function to get folder statistics
async function getFolderStats(folderId: string, user: User): Promise<{ size: number; files: number }> {
  let totalSize = 0;
  let totalFiles = 0;

  // Get direct files
  const files = await File.find({
    folder: folderId,
    isTrashed: false,
    processingStatus: 'completed'
  });

  for (const file of files) {
    if (canAccessResource(user, file as any, 'read')) {
      totalSize += file.size;
      totalFiles++;
    }
  }

  // Get subfolders
  const subfolders = await Folder.find({
    parent: folderId,
    isTrashed: false
  });

  for (const subfolder of subfolders) {
    if (canAccessResource(user, subfolder as any, 'read')) {
      const subStats = await getFolderStats(subfolder._id.toString(), user);
      totalSize += subStats.size;
      totalFiles += subStats.files;
    }
  }

  return { size: totalSize, files: totalFiles };
}

// Helper function to format file size
function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 Bytes';
  
  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}