import { NextRequest } from 'next/server';
import { File } from '@/models/File';
import { connectDB } from '@/lib/db';
import { abortUploadSchema } from '@/lib/validations/file';
import { apiRateLimiter, createRateLimitMiddleware } from '@/lib/rate-limit';
import { getServerSession } from 'next-auth';
import { authOptions } from '@/lib/auth';
import { getClientIP, formatHeaders } from '@/lib/api-utils';
import type { BaseResponse, User } from '@/types';

// POST /api/upload/abort - Abort upload and cleanup
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
    const { uploadId } = abortUploadSchema.parse(body);

    // Get upload session
    const uploadSessions = (globalThis as any).uploadSessions || new Map();
    const uploadSession = uploadSessions.get(uploadId);

    if (!uploadSession) {
      return Response.json({
        success: false,
        error: 'Upload session not found'
      }, { status: 404 });
    }

    // Verify ownership
    if (uploadSession.userId !== user._id.toString()) {
      return Response.json({
        success: false,
        error: 'Access denied'
      }, { status: 403 });
    }

    // Clean up file record if it exists
    if (uploadSession.fileId) {
      try {
        await File.findByIdAndDelete(uploadSession.fileId);
      } catch (error) {
        console.warn('Failed to delete file record:', error);
      }
    }

    // Clean up upload session and chunks
    uploadSessions.delete(uploadId);

    const response: BaseResponse<{
      aborted: boolean;
      uploadId: string;
      cleanedUp: boolean;
    }> = {
      success: true,
      data: {
        aborted: true,
        uploadId,
        cleanedUp: true
      }
    };

    return Response.json(response, {
      headers: formatHeaders(rateLimitResult.headers)
    });

  } catch (error: any) {
    console.error('Abort upload error:', error);
    
    if (error.name === 'ZodError') {
      return Response.json({
        success: false,
        error: 'Invalid request data',
        details: error.errors
      }, { status: 400 });
    }

    return Response.json({
      success: false,
      error: 'Failed to abort upload'
    }, { status: 500 });
  }
}

// DELETE /api/upload/abort - Alternative abort method
export async function DELETE(request: NextRequest) {
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
    const uploadId = url.searchParams.get('uploadId');

    if (!uploadId) {
      return Response.json({
        success: false,
        error: 'Upload ID is required'
      }, { status: 400 });
    }

    // Get upload session
    const uploadSessions = (globalThis as any).uploadSessions || new Map();
    const uploadSession = uploadSessions.get(uploadId);

    if (!uploadSession) {
      return Response.json({
        success: false,
        error: 'Upload session not found'
      }, { status: 404 });
    }

    // Verify ownership
    if (uploadSession.userId !== user._id.toString()) {
      return Response.json({
        success: false,
        error: 'Access denied'
      }, { status: 403 });
    }

    // Clean up file record if it exists
    if (uploadSession.fileId) {
      try {
        await connectDB();
        await File.findByIdAndDelete(uploadSession.fileId);
      } catch (error) {
        console.warn('Failed to delete file record:', error);
      }
    }

    // Clean up upload session and chunks
    uploadSessions.delete(uploadId);

    const response: BaseResponse<{
      deleted: boolean;
      uploadId: string;
    }> = {
      success: true,
      data: {
        deleted: true,
        uploadId
      }
    };

    return Response.json(response, {
      headers: formatHeaders(rateLimitResult.headers)
    });

  } catch (error: any) {
    console.error('Delete upload session error:', error);
    
    return Response.json({
      success: false,
      error: 'Failed to delete upload session'
    }, { status: 500 });
  }
}