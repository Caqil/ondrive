import { NextRequest } from 'next/server';
import { File } from '@/models/File';
import { Folder } from '@/models/Folder';
import { connectDB } from '@/lib/db';
import { apiRateLimiter, createRateLimitMiddleware } from '@/lib/rate-limit';
import { getServerSession } from 'next-auth';
import { authOptions } from '@/lib/auth';
import { getClientIP, formatHeaders } from '@/lib/api-utils';
import { z } from 'zod';
import type { BaseResponse, File as FileType, User } from '@/types';

const moveFileSchema = z.object({
  targetFolderId: z.string().min(1, 'Target folder ID is required')
});

// POST /api/client/files/[fileId]/move - Move file to another folder
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
    const body = await request.json();
    
    await connectDB();
    
    const { targetFolderId } = moveFileSchema.parse(body);

    // Verify file ownership
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

    // Verify target folder ownership
    const targetFolder = await Folder.findOne({
      _id: targetFolderId,
      owner: user._id
    });

    if (!targetFolder) {
      return Response.json({
        success: false,
        error: 'Target folder not found'
      }, { status: 404 });
    }

    // Move file
    const movedFile = await File.findByIdAndUpdate(
      fileId,
      {
        folder: targetFolderId,
        updatedAt: new Date()
      },
      { new: true, runValidators: true }
    )
    .populate('owner', 'name email avatar')
    .populate('folder', 'name path')
    .lean();

    const response: BaseResponse<FileType> = {
      success: true,
      data: movedFile as unknown as FileType
    };

    return Response.json(response, {
      headers: formatHeaders(rateLimitResult.headers)
    });
  } catch (error: any) {
    console.error('File move error:', error);
    
    if (error.name === 'ZodError') {
      return Response.json({
        success: false,
        error: 'Invalid request data',
        details: error.errors
      }, { status: 400 });
    }

    return Response.json({
      success: false,
      error: 'Failed to move file'
    }, { status: 500 });
  }
}
