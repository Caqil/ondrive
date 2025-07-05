import { NextRequest } from 'next/server';
import { File } from '@/models/File';
import { connectDB } from '@/lib/db';
import { updateFileSchema } from '@/lib/validations/file';
import { apiRateLimiter, createRateLimitMiddleware } from '@/lib/rate-limit';
import { getServerSession } from 'next-auth';
import { authOptions } from '@/lib/auth';
import { getClientIP, formatHeaders } from '@/lib/api-utils';
import type { BaseResponse, File as FileType, User } from '@/types';

// GET /api/client/files/[fileId] - Get file details
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
    })
    .populate('owner', 'name email avatar')
    .populate('folder', 'name path')
    .lean();

    if (!file) {
      return Response.json({
        success: false,
        error: 'File not found'
      }, { status: 404 });
    }

    // Update view count and last accessed
    await File.findByIdAndUpdate(fileId, {
      $inc: { viewCount: 1 },
      lastAccessedAt: new Date()
    });

    const response: BaseResponse<FileType> = {
      success: true,
      data: file as unknown as FileType
    };

    return Response.json(response, {
      headers: formatHeaders(rateLimitResult.headers)
    });
  } catch (error) {
    console.error('File get error:', error);
    return Response.json({
      success: false,
      error: 'Failed to fetch file'
    }, { status: 500 });
  }
}

// PATCH /api/client/files/[fileId] - Update file
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
    
    const validatedData = updateFileSchema.parse(body);

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

    const updatedFile = await File.findByIdAndUpdate(
      fileId,
      { ...validatedData, updatedAt: new Date() },
      { new: true, runValidators: true }
    )
    .populate('owner', 'name email avatar')
    .populate('folder', 'name path')
    .lean();

    const response: BaseResponse<FileType> = {
      success: true,
      data: updatedFile as unknown as FileType
    };

    return Response.json(response, {
      headers: formatHeaders(rateLimitResult.headers)
    });
  } catch (error: any) {
    console.error('File update error:', error);
    
    if (error.name === 'ZodError') {
      return Response.json({
        success: false,
        error: 'Invalid file data',
        details: error.errors
      }, { status: 400 });
    }

    return Response.json({
      success: false,
      error: 'Failed to update file'
    }, { status: 500 });
  }
}

// DELETE /api/client/files/[fileId] - Delete file (soft delete)
export async function DELETE(
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
      owner: user._id
    });

    if (!file) {
      return Response.json({
        success: false,
        error: 'File not found'
      }, { status: 404 });
    }

    // Soft delete - move to trash
    await File.findByIdAndUpdate(fileId, {
      isTrashed: true,
      trashedAt: new Date(),
      trashedBy: user._id
    });

    const response: BaseResponse<{ deleted: boolean }> = {
      success: true,
      data: { deleted: true }
    };

    return Response.json(response, {
      headers: formatHeaders(rateLimitResult.headers)
    });
  } catch (error) {
    console.error('File delete error:', error);
    return Response.json({
      success: false,
      error: 'Failed to delete file'
    }, { status: 500 });
  }
}