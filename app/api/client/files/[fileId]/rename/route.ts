import { NextRequest } from 'next/server';
import { File } from '@/models/File';
import { connectDB } from '@/lib/db';
import { apiRateLimiter, createRateLimitMiddleware } from '@/lib/rate-limit';
import { getServerSession } from 'next-auth';
import { authOptions } from '@/lib/auth';
import { getClientIP, formatHeaders } from '@/lib/api-utils';
import { z } from 'zod';
import type { BaseResponse, File as FileType, User } from '@/types';

const renameFileSchema = z.object({
  name: z.string().min(1, 'Name is required').max(255, 'Name too long')
});

// PATCH /api/client/files/[fileId]/rename - Rename file
export async function PATCH(
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
    const body = await request.json();
    
    await connectDB();
    
    const { name } = renameFileSchema.parse(body);

    const file = await File.findOne({
      _id: fileId,
      owner: user._id
    });

    if (!file) {
      return Response.json({
        success: false,
        error: 'File not found'
      }, { status: 404 });
    }

    // Update file name
    const renamedFile = await File.findByIdAndUpdate(
      fileId,
      {
        name,
        updatedAt: new Date()
      },
      { new: true, runValidators: true }
    )
    .populate('owner', 'name email avatar')
    .populate('folder', 'name path')
    .lean();

    const response: BaseResponse<FileType> = {
      success: true,
      data: renamedFile as unknown as FileType
    };

    return Response.json(response, {
      headers: formatHeaders(rateLimitResult.headers)
    });
  } catch (error: any) {
    console.error('File rename error:', error);
    
    if (error.name === 'ZodError') {
      return Response.json({
        success: false,
        error: 'Invalid name',
        details: error.errors
      }, { status: 400 });
    }

    return Response.json({
      success: false,
      error: 'Failed to rename file'
    }, { status: 500 });
  }
}
