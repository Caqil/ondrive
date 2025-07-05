import { NextRequest } from 'next/server';
import { File } from '@/models/File';
import { Folder } from '@/models/Folder';
import { connectDB } from '@/lib/db';
import { fileUploadSchema } from '@/lib/validations/file';
import { apiRateLimiter, createRateLimitMiddleware } from '@/lib/rate-limit';
import { getServerSession } from 'next-auth';
import { authOptions } from '@/lib/auth';
import { getClientIP, formatHeaders, getStorageProvider } from '@/lib/api-utils';
import { generateSafeFilename } from '@/lib/file-utils';
import { generateToken, generateFileChecksum } from '@/lib/crypto';
import { getUserStorageQuota } from '@/lib/permissions';
import type { BaseResponse, FileUploadResponse, User } from '@/types';
import crypto from 'crypto';

// In-memory upload session storage (use Redis in production)
const uploadSessions = new Map<string, any>();

// Make sessions globally accessible for other routes
(globalThis as any).uploadSessions = uploadSessions;

// POST /api/upload - Initialize upload session
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
    const { name, size, mimeType, folderId, tags, labels } = fileUploadSchema.parse(body);

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

    // Check user quota and file size limits
    const userFiles = await File.aggregate([
      {
        $match: {
          owner: user._id,
          isTrashed: false
        }
      },
      {
        $group: {
          _id: null,
          totalSize: { $sum: '$size' },
          totalFiles: { $sum: 1 }
        }
      }
    ]);

    const currentUsage = userFiles[0] || { totalSize: 0, totalFiles: 0 };
    const userQuota = getUserStorageQuota(user);
    const maxFileSize = 750 * 1024 * 1024 * 1024; // 750GB max file size

    if (currentUsage.totalSize + size > userQuota) {
      return Response.json({
        success: false,
        error: 'File size exceeds available quota'
      }, { status: 413 });
    }

    if (size > maxFileSize) {
      return Response.json({
        success: false,
        error: 'File size exceeds maximum allowed size'
      }, { status: 413 });
    }

    // Generate unique identifiers
    const uploadId = crypto.randomUUID();
    const extension = name.split('.').pop() || '';
    const safeName = generateSafeFilename(name);
    const uniqueKey = `uploads/${user._id}/${folderId}/${Date.now()}-${safeName}`;

    // Get storage provider configuration
    const storageConfig = await getStorageProvider();
    const chunkSize = 10 * 1024 * 1024; // 10MB chunks
    const maxChunks = Math.ceil(size / chunkSize);

    // Determine upload strategy
    const useChunkedUpload = size > chunkSize;

    // Create file record in pending state
    const fileData = {
      name: name.replace(/\.[^/.]+$/, ''), // Remove extension from display name
      originalName: name,
      mimeType,
      size,
      path: uniqueKey,
      key: uniqueKey,
      extension: extension.toLowerCase(),
      checksum: '', // Will be set after upload completion
      metadata: {},
      folder: folderId,
      owner: user._id,
      team: user.currentTeam,
      visibility: folder.visibility || 'private',
      tags: tags || [],
      labels: labels || [],
      storageProvider: storageConfig.provider,
      processingStatus: 'pending',
      version: 1,
      versionHistory: [],
      isLatestVersion: true,
      isStarred: false,
      isTrashed: false,
      shareCount: 0,
      downloadCount: 0,
      viewCount: 0,
      syncStatus: 'pending'
    };

    const file = new File(fileData);
    await file.save();

    // Create upload session
    const uploadSession = {
      uploadId,
      fileId: file._id.toString(),
      userId: user._id.toString(),
      size,
      uploadedChunks: new Set<number>(),
      isComplete: false,
      useChunkedUpload,
      totalChunks: useChunkedUpload ? maxChunks : 1,
      chunkSize,
      key: uniqueKey,
      fileName: name,
      mimeType,
      createdAt: new Date(),
      expiresAt: new Date(Date.now() + 24 * 60 * 60 * 1000), // 24 hours
      chunks: new Map<number, Buffer>() // Store chunks temporarily
    };

    // Store session (use Redis in production)
    uploadSessions.set(uploadId, uploadSession);

    // Generate response
    let uploadResponse: FileUploadResponse;

    if (useChunkedUpload) {
      uploadResponse = {
        uploadUrl: `/api/upload/chunk`,
        uploadId,
        fileId: file._id,
        chunkSize,
        maxChunks
      };
    } else {
      // Generate direct upload URL
      const uploadUrlResponse = await storageConfig.getSignedUrl(uniqueKey, {
        operation: 'upload',
        contentType: mimeType,
        contentLength: size,
        expiresIn: 3600 // 1 hour
      });

      // Extract URL from response (handle both string and object returns)
      const signedUrl = typeof uploadUrlResponse === 'string' 
        ? uploadUrlResponse 
        : uploadUrlResponse.uploadUrl || `/api/upload/direct?key=${encodeURIComponent(uniqueKey)}`;

      uploadResponse = {
        uploadUrl: signedUrl,
        uploadId,
        fileId: file._id
      };
    }

    const response: BaseResponse<FileUploadResponse> = {
      success: true,
      data: uploadResponse
    };

    return Response.json(response, {
      headers: formatHeaders(rateLimitResult.headers)
    });

  } catch (error: any) {
    console.error('Upload initialization error:', error);
    
    if (error.name === 'ZodError') {
      return Response.json({
        success: false,
        error: 'Invalid request data',
        details: error.errors
      }, { status: 400 });
    }

    return Response.json({
      success: false,
      error: 'Failed to initialize upload'
    }, { status: 500 });
  }
}

// GET /api/upload - Get upload session status
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
    const uploadSession = uploadSessions.get(uploadId);

    if (!uploadSession) {
      return Response.json({
        success: false,
        error: 'Upload session not found or expired'
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
      fileId: string;
      isComplete: boolean;
      uploadedChunks: number[];
      totalChunks: number;
      progress: number;
    }> = {
      success: true,
      data: {
        uploadId: uploadSession.uploadId,
        fileId: uploadSession.fileId,
        isComplete: uploadSession.isComplete,
        uploadedChunks: Array.from(uploadSession.uploadedChunks),
        totalChunks: uploadSession.totalChunks,
        progress: uploadSession.uploadedChunks.size / uploadSession.totalChunks * 100
      }
    };

    return Response.json(response, {
      headers: formatHeaders(rateLimitResult.headers)
    });

  } catch (error: any) {
    console.error('Get upload status error:', error);
    
    return Response.json({
      success: false,
      error: 'Failed to get upload status'
    }, { status: 500 });
  }
}