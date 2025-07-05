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
import { generateFileChecksum, generateToken } from '@/lib/crypto';
import { getUserStorageQuota } from '@/lib/permissions';
import type { BaseResponse, FileUploadResponse, User } from '@/types';
import crypto from 'crypto';

// POST /api/client/files/upload - Create upload session
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

    // Check if folder exists and user has access
    const folder = await Folder.findOne({
      _id: folderId,
      $or: [
        { owner: user._id },
        { team: user.currentTeam, visibility: { $in: ['team', 'public'] } }
      ],
      isTrashed: false
    });

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
    const maxFileSize = 500 * 1024 * 1024; // 500MB max file size

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

    // Generate unique file key and upload ID
    const uploadId = crypto.randomUUID();
    const extension = name.split('.').pop() || '';
    const uniqueKey = `${user._id}/${folderId}/${Date.now()}-${generateSafeFilename(name)}`;

    // Get storage provider configuration
    const storageConfig = await getStorageProvider();
    const chunkSize = 5 * 1024 * 1024; // 5MB chunks
    const maxChunks = Math.ceil(size / chunkSize);

    // Determine if chunked upload is needed
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
      visibility: folder.visibility,
      tags: tags || [],
      labels: labels || [],
      storageProvider: storageConfig.provider,
      processingStatus: 'pending',
      version: 1,
      versionHistory: [],
      isLatestVersion: true
    };

    const file = new File(fileData);
    await file.save();

    // Generate signed URL for upload
    let uploadResponse: FileUploadResponse;

    if (useChunkedUpload) {
      // Prepare chunked upload configuration
      uploadResponse = {
        uploadUrl: `/api/client/files/upload/chunk`,
        uploadId,
        fileId: file._id,
        chunkSize,
        maxChunks
      };
    } else {
      // Generate direct upload URL
      const signedUrl = await storageConfig.getSignedUrl(uniqueKey, {
        operation: 'upload',
        contentType: mimeType,
        contentLength: size,
        expiresIn: 3600 // 1 hour
      });

      uploadResponse = {
        uploadUrl: signedUrl as string,
        uploadId,
        fileId: file._id
      };
    }

    // Store upload session for tracking
    const uploadSession = {
      uploadId,
      fileId: file._id,
      userId: user._id,
      size,
      uploadedChunks: [],
      isComplete: false,
      expiresAt: new Date(Date.now() + 24 * 60 * 60 * 1000), // 24 hours
      chunkedUpload: useChunkedUpload,
      totalChunks: useChunkedUpload ? maxChunks : 1
    };

    // In a real implementation, store this in Redis or a temporary collection
    // For now, we'll store in file metadata
    await File.findByIdAndUpdate(file._id, {
      $set: {
        'metadata.uploadSession': uploadSession
      }
    });

    const response: BaseResponse<FileUploadResponse> = {
      success: true,
      data: uploadResponse
    };

    return Response.json(response, {
      headers: formatHeaders(rateLimitResult.headers)
    });

  } catch (error: any) {
    console.error('File upload initialization error:', error);
    
    if (error.name === 'ZodError') {
      return Response.json({
        success: false,
        error: 'Invalid request data',
        details: error.errors
      }, { status: 400 });
    }

    return Response.json({
      success: false,
      error: 'Failed to initialize file upload'
    }, { status: 500 });
  }
}

// PUT /api/client/files/upload - Complete direct upload
export async function PUT(request: NextRequest) {
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

    const { searchParams } = new URL(request.url);
    const uploadId = searchParams.get('uploadId');
    
    if (!uploadId) {
      return Response.json({
        success: false,
        error: 'Upload ID is required'
      }, { status: 400 });
    }

    // Find file by upload ID
    const file = await File.findOne({
      'metadata.uploadSession.uploadId': uploadId,
      owner: user._id
    });

    if (!file) {
      return Response.json({
        success: false,
        error: 'Upload session not found'
      }, { status: 404 });
    }

    // Get uploaded file from request
    const contentLength = request.headers.get('content-length');
    if (!contentLength || parseInt(contentLength) !== file.size) {
      return Response.json({
        success: false,
        error: 'Content length mismatch'
      }, { status: 400 });
    }

    // Calculate checksum of uploaded content
    const buffer = await request.arrayBuffer();
    const checksum = generateFileChecksum(Buffer.from(buffer));

    // Update file with final details
    const storageConfig = await getStorageProvider();
    const previewUrl = await storageConfig.getSignedUrl(file.key, {
      operation: 'preview',
      expiresIn: 3600
    });

    await File.findByIdAndUpdate(file._id, {
      $set: {
        checksum,
        url: previewUrl,
        previewUrl,
        processingStatus: 'completed',
        'metadata.uploadSession.isComplete': true,
        'metadata.completedAt': new Date()
      },
      $unset: {
        'metadata.uploadSession': 1
      }
    });

    // Update folder statistics
    await Folder.findByIdAndUpdate(file.folder, {
      $inc: {
        fileCount: 1,
        totalSize: file.size
      }
    });

    const response: BaseResponse<{ fileId: string; url?: string }> = {
      success: true,
      data: {
        fileId: file._id.toString(),
        url: typeof previewUrl === 'string' ? previewUrl : undefined
      }
    };

    return Response.json(response, {
      headers: formatHeaders(rateLimitResult.headers)
    });

  } catch (error: any) {
    console.error('Direct upload completion error:', error);
    
    return Response.json({
      success: false,
      error: 'Failed to complete upload'
    }, { status: 500 });
  }
}