import { NextRequest } from 'next/server';
import { File } from '@/models/File';
import { connectDB } from '@/lib/db';
import { apiRateLimiter, createRateLimitMiddleware } from '@/lib/rate-limit';
import { getServerSession } from 'next-auth';
import { authOptions } from '@/lib/auth';
import { getClientIP, formatHeaders, getStorageProvider } from '@/lib/api-utils';
import type { User } from '@/types';

// GET /api/client/files/[fileId]/preview - Get file preview
export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ fileId: string }> }
) {
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
    const { fileId } = await params;
    
    await connectDB();

    const file = await File.findOne({
      _id: fileId,
      $or: [
        { owner: user._id },
        { visibility: 'public' },
        { team: user.currentTeam, visibility: 'team' }
      ]
    });

    if (!file) {
      return Response.json({
        success: false,
        error: 'File not found'
      }, { status: 404 });
    }

    if (file.isTrashed) {
      return Response.json({
        success: false,
        error: 'File is in trash'
      }, { status: 403 });
    }

    const storageProvider = await getStorageProvider(file.storageProvider);
    let previewUrl = file.previewUrl;

    // Generate preview URL if not exists
    if (!previewUrl && file.url) {
      previewUrl = await storageProvider.getSignedUrl(file.key, {
        operation: 'preview',
        expiresIn: 3600
      });
    }

    // Update view count
    await File.findByIdAndUpdate(fileId, {
      $inc: { viewCount: 1 },
      lastAccessedAt: new Date()
    });

    return Response.json({
      success: true,
      data: {
        previewUrl,
        thumbnailUrl: file.thumbnailUrl,
        metadata: file.metadata,
        mimeType: file.mimeType,
        size: file.size,
        ocrText: file.ocrText
      }
    }, {
      headers: formatHeaders(rateLimitResult.headers)
    });
  } catch (error) {
    console.error('File preview error:', error);
    return Response.json({
      success: false,
      error: 'Failed to generate preview'
    }, { status: 500 });
  }
}
