import { NextRequest } from 'next/server';
import { File } from '@/models/File';
import { StorageProvider } from '@/models/StorageProvider';
import { connectDB } from '@/lib/db';
import { downloadRateLimiter, createRateLimitMiddleware } from '@/lib/rate-limit';
import { getServerSession } from 'next-auth';
import { authOptions } from '@/lib/auth';
import { getClientIP, formatHeaders } from '@/lib/api-utils';
import type { User } from '@/types';

// Mock storage provider implementation
const getStorageProvider = async (providerType?: string) => {
  const provider = await StorageProvider.findOne({
    $or: [
      { type: providerType, isActive: true },
      { isDefault: true, isActive: true }
    ]
  }).lean();

  return {
    provider: (provider && typeof provider === 'object' && 'type' in provider ? (provider as any).type : 'local'),
    getSignedUrl: async (key: string, options: any) => {
      // Implementation would vary by provider
      return `https://storage.example.com/${key}?expires=${Date.now() + (options.expiresIn * 1000)}`;
    }
  };
};

// GET /api/client/files/[fileId]/download - Download file
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
    const rateLimitResult = createRateLimitMiddleware(downloadRateLimiter)(clientIP);
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

    // Get signed download URL from storage provider
    const storageProvider = await getStorageProvider(file.storageProvider);
    const downloadUrl = await storageProvider.getSignedUrl(file.key, {
      operation: 'download',
      expiresIn: 3600, // 1 hour
      filename: file.originalName
    });

    // Update download count
    await File.findByIdAndUpdate(fileId, {
      $inc: { downloadCount: 1 },
      lastAccessedAt: new Date()
    });

    return Response.json({
      success: true,
      data: {
        downloadUrl,
        filename: file.originalName,
        size: file.size,
        mimeType: file.mimeType
      }
    }, {
      headers: formatHeaders(rateLimitResult.headers)
    });
  } catch (error) {
    console.error('File download error:', error);
    return Response.json({
      success: false,
      error: 'Failed to generate download URL'
    }, { status: 500 });
  }
}