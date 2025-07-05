import { NextRequest } from 'next/server';
import { File } from '@/models/File';
import { Share } from '@/models/Share';
import { connectDB } from '@/lib/db';
import { apiRateLimiter, createRateLimitMiddleware } from '@/lib/rate-limit';
import { getServerSession } from 'next-auth';
import { authOptions } from '@/lib/auth';
import { getClientIP, formatHeaders, getStorageProvider } from '@/lib/api-utils';
import { canAccessResource, canAccessSharedResource } from '@/lib/permissions';
import type { BaseResponse, User } from '@/types';
import fs from 'fs/promises';
import path from 'path';

// GET /api/download/[fileId] - Download single file
export async function GET(
  request: NextRequest,
  { params }: { params: { fileId: string } }
) {
  try {
    const session = await getServerSession(authOptions);
    const user = session?.user as User;
    
    const clientIP = getClientIP(request);
    const rateLimitResult = createRateLimitMiddleware(apiRateLimiter)(clientIP);
    
    await connectDB();
    const { fileId } = params;
    const url = new URL(request.url);
    
    // Check for share token (for public downloads)
    const shareToken = url.searchParams.get('token');
    const download = url.searchParams.get('download') === 'true';
    const inline = url.searchParams.get('inline') === 'true';

    // Find file
    const fileQuery = await File.findOne({
      _id: fileId,
      isTrashed: false,
      processingStatus: 'completed'
    })
    .populate('owner', 'name email avatar')
    .populate('folder', 'name path');

    const file = fileQuery as any;

    if (!file) {
      return Response.json({
        success: false,
        error: 'File not found'
      }, { status: 404 });
    }

    // Check access permissions
    let hasAccess = false;
    let shareInfo: any = null;

    if (shareToken) {
      // Check share access
      const shareQuery = await Share.findOne({
        token: shareToken,
        resource: fileId,
        resourceType: 'file',
        isActive: true,
        isRevoked: false
      });

      const share = shareQuery as any;

      if (share) {
        // Check if share allows downloads
        if (!share.allowDownload) {
          return Response.json({
            success: false,
            error: 'Download not allowed for this share'
          }, { status: 403 });
        }

        // Check share permissions
        hasAccess = canAccessSharedResource(user, share, 'download');
        shareInfo = share;
      }
    } else if (user) {
      // Check direct file access
      hasAccess = canAccessResource(user, file, 'read');
    }

    if (!hasAccess) {
      return Response.json({
        success: false,
        error: user ? 'Access denied' : 'Authentication required'
      }, { status: user ? 403 : 401 });
    }

    // Get storage provider and file location
    const storageConfig = await getStorageProvider();
    let fileStream: ReadableStream | null = null;

    try {
      if (storageConfig.provider === 'local') {
        // Local file system
        const uploadDir = process.env.UPLOAD_DIR || './uploads';
        const filePath = path.join(uploadDir, file.key);

        try {
          await fs.access(filePath);
          const fileBuffer = await fs.readFile(filePath);
          
          fileStream = new ReadableStream({
            start(controller) {
              controller.enqueue(fileBuffer);
              controller.close();
            }
          });

        } catch (fsError) {
          console.error('File access error:', fsError);
          return Response.json({
            success: false,
            error: 'File not accessible'
          }, { status: 404 });
        }

      } else {
        // Cloud storage - generate signed download URL
        const downloadUrl = await storageConfig.getSignedUrl(file.key, {
          operation: 'download',
          expiresIn: 3600,
          responseContentDisposition: download 
            ? `attachment; filename="${encodeURIComponent(file.originalName)}"` 
            : undefined
        });

        // Redirect to signed URL for cloud storage
        return Response.redirect(typeof downloadUrl === 'string' ? downloadUrl : downloadUrl.toString());
      }

    } catch (storageError) {
      console.error('Storage access error:', storageError);
      return Response.json({
        success: false,
        error: 'Failed to access file storage'
      }, { status: 500 });
    }

    if (!fileStream) {
      return Response.json({
        success: false,
        error: 'Failed to create file stream'
      }, { status: 500 });
    }

    // Update download statistics
    try {
      await File.findByIdAndUpdate(fileId, {
        $inc: { downloadCount: 1 },
        $set: { lastAccessedAt: new Date() }
      });

      // Log share access if this is a shared download
      if (shareInfo) {
        await Share.findByIdAndUpdate(shareInfo._id, {
          $inc: { accessCount: 1 },
          $set: { lastAccessedAt: new Date() },
          $push: {
            accessLog: {
              ip: clientIP,
              userAgent: request.headers.get('user-agent') || 'Unknown',
              userId: user?._id,
              email: user?.email,
              accessedAt: new Date(),
              action: 'download'
            }
          }
        });
      }

    } catch (updateError) {
      console.warn('Failed to update download statistics:', updateError);
    }

    // Prepare response headers
    const headers = new Headers();
    
    // Content type
    headers.set('Content-Type', file.mimeType || 'application/octet-stream');
    headers.set('Content-Length', file.size.toString());
    
    // Content disposition
    if (download) {
      headers.set('Content-Disposition', `attachment; filename="${encodeURIComponent(file.originalName)}"`);
    } else if (inline) {
      headers.set('Content-Disposition', `inline; filename="${encodeURIComponent(file.originalName)}"`);
    }
    
    // Cache headers
    headers.set('Cache-Control', 'private, max-age=3600');
    headers.set('ETag', `"${file.checksum}"`);
    
    // Security headers
    headers.set('X-Content-Type-Options', 'nosniff');
    
    // Rate limit headers
    Object.entries(formatHeaders(rateLimitResult.headers)).forEach(([key, value]) => {
      headers.set(key, value);
    });

    // Return file stream
    return new Response(fileStream, {
      status: 200,
      headers
    });

  } catch (error: any) {
    console.error('File download error:', error);
    
    return Response.json({
      success: false,
      error: 'Failed to download file'
    }, { status: 500 });
  }
}

// HEAD /api/download/[fileId] - Get file metadata for download
export async function HEAD(
  request: NextRequest,
  { params }: { params: { fileId: string } }
) {
  try {
    const session = await getServerSession(authOptions);
    const user = session?.user as User;
    
    const clientIP = getClientIP(request);
    const rateLimitResult = createRateLimitMiddleware(apiRateLimiter)(clientIP);
    
    await connectDB();
    const { fileId } = params;
    const url = new URL(request.url);
    
    const shareToken = url.searchParams.get('token');

    // Find file
    const fileQuery = await File.findOne({
      _id: fileId,
      isTrashed: false,
      processingStatus: 'completed'
    });

    const file = fileQuery as any;

    if (!file) {
      return new Response(null, { status: 404 });
    }

    // Check access permissions
    let hasAccess = false;

    if (shareToken) {
      const shareQuery = await Share.findOne({
        token: shareToken,
        resource: fileId,
        resourceType: 'file',
        isActive: true,
        isRevoked: false
      });

      const share = shareQuery as any;
      if (share && share.allowDownload) {
        hasAccess = canAccessSharedResource(user, share, 'download');
      }
    } else if (user) {
      hasAccess = canAccessResource(user, file, 'read');
    }

    if (!hasAccess) {
      return new Response(null, { status: user ? 403 : 401 });
    }

    // Prepare headers with file metadata
    const headers = new Headers();
    headers.set('Content-Type', file.mimeType || 'application/octet-stream');
    headers.set('Content-Length', file.size.toString());
    headers.set('Last-Modified', file.updatedAt.toUTCString());
    headers.set('ETag', `"${file.checksum}"`);
    headers.set('Accept-Ranges', 'bytes');

    // Rate limit headers
    Object.entries(formatHeaders(rateLimitResult.headers)).forEach(([key, value]) => {
      headers.set(key, value);
    });

    return new Response(null, {
      status: 200,
      headers
    });

  } catch (error: any) {
    console.error('File HEAD request error:', error);
    return new Response(null, { status: 500 });
  }
}