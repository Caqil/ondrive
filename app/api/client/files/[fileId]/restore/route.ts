import { NextRequest } from 'next/server';
import { File } from '@/models/File';
import { connectDB } from '@/lib/db';
import { apiRateLimiter, createRateLimitMiddleware } from '@/lib/rate-limit';
import { getServerSession } from 'next-auth';
import { authOptions } from '@/lib/auth';
import { getClientIP, formatHeaders } from '@/lib/api-utils';
import type { BaseResponse, File as FileType, User } from '@/types';

// POST /api/client/files/[fileId]/restore - Restore file from trash
export async function POST(
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
      owner: user._id,
      isTrashed: true
    });

    if (!file) {
      return Response.json({
        success: false,
        error: 'File not found in trash'
      }, { status: 404 });
    }

    // Restore file
    const restoredFile = await File.findByIdAndUpdate(
      fileId,
      {
        isTrashed: false,
        trashedAt: null,
        trashedBy: null,
        updatedAt: new Date()
      },
      { new: true, runValidators: true }
    )
    .populate('owner', 'name email avatar')
    .populate('folder', 'name path')
    .lean();

    const response: BaseResponse<FileType> = {
      success: true,
      data: restoredFile as unknown as FileType
    };

    return Response.json(response, {
      headers: formatHeaders(rateLimitResult.headers)
    });
  } catch (error) {
    console.error('File restore error:', error);
    return Response.json({
      success: false,
      error: 'Failed to restore file'
    }, { status: 500 });
  }
}
