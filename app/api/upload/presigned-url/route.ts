import { NextRequest } from 'next/server';
import { Folder } from '@/models/Folder';
import { connectDB } from '@/lib/db';
import { apiRateLimiter, createRateLimitMiddleware } from '@/lib/rate-limit';
import { getServerSession } from 'next-auth';
import { authOptions } from '@/lib/auth';
import { getClientIP, formatHeaders, getStorageProvider } from '@/lib/api-utils';
import { generateSafeFilename } from '@/lib/file-utils';
import { getUserStorageQuota } from '@/lib/permissions';
import type { BaseResponse, User } from '@/types';
import crypto from 'crypto';

// POST /api/upload/presigned-url - Generate presigned URL for direct upload
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
      fileName,
      fileSize,
      mimeType,
      folderId,
      expiresIn = 3600 // Default 1 hour
    } = body;

    // Validate required fields
    if (!fileName || !fileSize || !mimeType || !folderId) {
      return Response.json({
        success: false,
        error: 'Missing required fields: fileName, fileSize, mimeType, folderId'
      }, { status: 400 });
    }

    // Verify folder exists and user has access
    const folderQuery = await Folder.findOne({
      _id: folderId,
      $or: [
        { owner: user._id },
        { team: user.currentTeam, visibility: { $in: ['team', 'public'] } }
      ],
      isTrashed: false
    });

    const folder = folderQuery as any;

    if (!folder) {
      return Response.json({
        success: false,
        error: 'Folder not found or access denied'
      }, { status: 404 });
    }

    // Check file size limits
    const maxFileSize = 750 * 1024 * 1024 * 1024; // 750GB
    if (fileSize > maxFileSize) {
      return Response.json({
        success: false,
        error: 'File size exceeds maximum allowed size'
      }, { status: 413 });
    }

    // Check user quota
    const userQuota = getUserStorageQuota(user);
    if (fileSize > userQuota) {
      return Response.json({
        success: false,
        error: 'File size exceeds available quota'
      }, { status: 413 });
    }

    // Generate unique file key
    const safeName = generateSafeFilename(fileName);
    const uniqueKey = `uploads/${user._id}/${folderId}/${Date.now()}-${safeName}`;

    // Get storage provider
    const storageConfig = await getStorageProvider();

    // Generate presigned URL based on storage provider
    let presignedUrl: string;
    let uploadFields: Record<string, string> = {};

    try {
      if (storageConfig.provider === 'local') {
        // For local storage, return our own upload endpoint
        presignedUrl = `/api/upload/direct?key=${encodeURIComponent(uniqueKey)}`;
      } else {
        // For cloud storage providers
        const signedResult = await storageConfig.getSignedUrl(uniqueKey, {
          operation: 'upload',
          contentType: mimeType,
          contentLength: fileSize,
          expiresIn
        });

        if (typeof signedResult === 'string') {
          presignedUrl = signedResult;
        } else {
          presignedUrl = (signedResult as any).uploadUrl || (signedResult as any).url || signedResult;
          uploadFields = (signedResult as any).fields || {};
        }
      }

    } catch (error) {
      console.error('Failed to generate presigned URL:', error);
      return Response.json({
        success: false,
        error: 'Failed to generate upload URL'
      }, { status: 500 });
    }

    // Generate upload session ID for tracking
    const uploadId = crypto.randomUUID();

    // Store basic session info (without chunks since this is direct upload)
    const uploadSessions = (globalThis as any).uploadSessions || new Map();
    (globalThis as any).uploadSessions = uploadSessions;

    const uploadSession = {
      uploadId,
      userId: user._id.toString(),
      fileName,
      fileSize,
      mimeType,
      folderId,
      key: uniqueKey,
      isDirect: true,
      createdAt: new Date(),
      expiresAt: new Date(Date.now() + expiresIn * 1000)
    };

    uploadSessions.set(uploadId, uploadSession);

    const response: BaseResponse<{
      uploadUrl: string;
      uploadId: string;
      key: string;
      fields?: Record<string, string>;
      expiresIn: number;
      maxFileSize: number;
    }> = {
      success: true,
      data: {
        uploadUrl: presignedUrl,
        uploadId,
        key: uniqueKey,
        fields: uploadFields,
        expiresIn,
        maxFileSize
      }
    };

    return Response.json(response, {
      headers: formatHeaders(rateLimitResult.headers)
    });

  } catch (error: any) {
    console.error('Generate presigned URL error:', error);
    
    return Response.json({
      success: false,
      error: 'Failed to generate presigned URL'
    }, { status: 500 });
  }
}

// GET /api/upload/presigned-url - Get upload URL info
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

    // Check if session has expired
    if (new Date() > uploadSession.expiresAt) {
      uploadSessions.delete(uploadId);
      return Response.json({
        success: false,
        error: 'Upload session has expired'
      }, { status: 410 });
    }

    const response: BaseResponse<{
      uploadId: string;
      fileName: string;
      fileSize: number;
      mimeType: string;
      key: string;
      isDirect: boolean;
      expiresAt: string;
    }> = {
      success: true,
      data: {
        uploadId: uploadSession.uploadId,
        fileName: uploadSession.fileName,
        fileSize: uploadSession.fileSize,
        mimeType: uploadSession.mimeType,
        key: uploadSession.key,
        isDirect: uploadSession.isDirect,
        expiresAt: uploadSession.expiresAt.toISOString()
      }
    };

    return Response.json(response, {
      headers: formatHeaders(rateLimitResult.headers)
    });

  } catch (error: any) {
    console.error('Get presigned URL info error:', error);
    
    return Response.json({
      success: false,
      error: 'Failed to get upload info'
    }, { status: 500 });
  }
}