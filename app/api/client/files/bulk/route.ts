import { NextRequest } from 'next/server';
import { File } from '@/models/File';
import { Folder } from '@/models/Folder';
import { connectDB } from '@/lib/db';
import { bulkFileOperationSchema } from '@/lib/validations/file';
import { apiRateLimiter, createRateLimitMiddleware } from '@/lib/rate-limit';
import { getServerSession } from 'next-auth';
import { authOptions } from '@/lib/auth';
import { getClientIP, formatHeaders } from '@/lib/api-utils';
import type { BaseResponse, User } from '@/types';

// POST /api/client/files/bulk - Bulk copy operation
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
    
    const { fileIds, action, targetFolderId } = bulkFileOperationSchema.parse(body);

    if (action !== 'copy') {
      return Response.json({
        success: false,
        error: 'Invalid action for POST method'
      }, { status: 400 });
    }

    // Verify target folder
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

    // Verify file access
    const files = await File.find({
      _id: { $in: fileIds },
      $or: [
        { owner: user._id },
        { visibility: 'public' },
        { team: user.currentTeam, visibility: 'team' }
      ]
    });

    if (files.length !== fileIds.length) {
      return Response.json({
        success: false,
        error: 'Some files not found'
      }, { status: 404 });
    }

    // Copy files
    const copiedFiles: string[] = [];
    for (const file of files) {
      const newKey = `${user._id}/${targetFolderId}/${Date.now()}-${file.originalName}`;
      
      const copiedFile = new File({
        name: `Copy of ${file.name}`,
        originalName: file.originalName,
        mimeType: file.mimeType,
        size: file.size,
        path: newKey,
        key: newKey,
        extension: file.extension,
        checksum: file.checksum,
        metadata: file.metadata,
        folder: targetFolderId,
        owner: user._id,
        team: user.currentTeam,
        visibility: file.visibility,
        tags: file.tags,
        labels: file.labels,
        storageProvider: file.storageProvider,
        version: 1,
        versionHistory: [],
        shareCount: 0,
        downloadCount: 0,
        viewCount: 0,
        processingStatus: 'completed'
      });

      await copiedFile.save();
      copiedFiles.push(copiedFile._id.toString());
    }

    const response: BaseResponse<{ copiedFiles: string[] }> = {
      success: true,
      data: { copiedFiles }
    };

    return Response.json(response, {
      headers: formatHeaders(rateLimitResult.headers)
    });
  } catch (error: any) {
    console.error('Bulk copy error:', error);
    
    if (error.name === 'ZodError') {
      return Response.json({
        success: false,
        error: 'Invalid request data',
        details: error.errors
      }, { status: 400 });
    }

    return Response.json({
      success: false,
      error: 'Failed to copy files'
    }, { status: 500 });
  }
}

// PATCH /api/client/files/bulk - Bulk move operation
export async function PATCH(request: NextRequest) {
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
    
    const { fileIds, action, targetFolderId } = bulkFileOperationSchema.parse(body);

    if (action !== 'move') {
      return Response.json({
        success: false,
        error: 'Invalid action for PATCH method'
      }, { status: 400 });
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

    // Verify file ownership
    const files = await File.find({
      _id: { $in: fileIds },
      owner: user._id
    });

    if (files.length !== fileIds.length) {
      return Response.json({
        success: false,
        error: 'Some files not found'
      }, { status: 404 });
    }

    // Move files
    await File.updateMany(
      { _id: { $in: fileIds } },
      { 
        folder: targetFolderId,
        updatedAt: new Date()
      }
    );

    const response: BaseResponse<{ movedFiles: string[] }> = {
      success: true,
      data: { movedFiles: fileIds }
    };

    return Response.json(response, {
      headers: formatHeaders(rateLimitResult.headers)
    });
  } catch (error: any) {
    console.error('Bulk move error:', error);
    
    if (error.name === 'ZodError') {
      return Response.json({
        success: false,
        error: 'Invalid request data',
        details: error.errors
      }, { status: 400 });
    }

    return Response.json({
      success: false,
      error: 'Failed to move files'
    }, { status: 500 });
  }
}

// DELETE /api/client/files/bulk - Bulk delete operation
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
    
    await connectDB();
    const body = await request.json();
    
    const { fileIds } = bulkFileOperationSchema.parse(body);

    // Verify file ownership
    const files = await File.find({
      _id: { $in: fileIds },
      owner: user._id
    });

    if (files.length !== fileIds.length) {
      return Response.json({
        success: false,
        error: 'Some files not found'
      }, { status: 404 });
    }

    // Soft delete - move to trash
    await File.updateMany(
      { _id: { $in: fileIds } },
      {
        isTrashed: true,
        trashedAt: new Date(),
        trashedBy: user._id
      }
    );

    const response: BaseResponse<{ deletedFiles: string[] }> = {
      success: true,
      data: { deletedFiles: fileIds }
    };

    return Response.json(response, {
      headers: formatHeaders(rateLimitResult.headers)
    });
  } catch (error: any) {
    console.error('Bulk delete error:', error);
    
    if (error.name === 'ZodError') {
      return Response.json({
        success: false,
        error: 'Invalid request data',
        details: error.errors
      }, { status: 400 });
    }

    return Response.json({
      success: false,
      error: 'Failed to delete files'
    }, { status: 500 });
  }
}
