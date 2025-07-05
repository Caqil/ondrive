import { NextRequest } from 'next/server';
import { File } from '@/models/File';
import { Folder } from '@/models/Folder';
import { connectDB } from '@/lib/db';
import { apiRateLimiter, createRateLimitMiddleware } from '@/lib/rate-limit';
import { getServerSession } from 'next-auth';
import { authOptions } from '@/lib/auth';
import { getClientIP, formatHeaders, getStorageProvider } from '@/lib/api-utils';
import { z } from 'zod';
import type { BaseResponse, File as FileType, User } from '@/types';

const copyFileSchema = z.object({
  targetFolderId: z.string().min(1, 'Target folder ID is required'),
  name: z.string().optional()
});

// POST /api/client/files/[fileId]/copy - Copy file to another folder
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
    
    const { targetFolderId, name } = copyFileSchema.parse(body);

    // Verify file access
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

    // Copy file in storage would happen here
    const storageProvider = await getStorageProvider(file.storageProvider);
    const newKey = `${user._id}/${targetFolderId}/${Date.now()}-${file.originalName}`;
    
    // Create new file record
    const newFileName = name || `Copy of ${file.name}`;
    const copiedFile = new File({
      name: newFileName,
      originalName: newFileName,
      mimeType: file.mimeType,
      size: file.size,
      path: file.path,
      key: newKey,
      url: file.url,
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
      parentVersion: null,
      isLatestVersion: true,
      shareCount: 0,
      downloadCount: 0,
      viewCount: 0,
      processingStatus: 'completed'
    });

    await copiedFile.save();

    const populatedFile = await File.findById(copiedFile._id)
      .populate('owner', 'name email avatar')
      .populate('folder', 'name path')
      .lean();

    const response: BaseResponse<FileType> = {
      success: true,
      data: populatedFile as unknown as FileType
    };

    return Response.json(response, {
      headers: formatHeaders(rateLimitResult.headers)
    });
  } catch (error: any) {
    console.error('File copy error:', error);
    
    if (error.name === 'ZodError') {
      return Response.json({
        success: false,
        error: 'Invalid request data',
        details: error.errors
      }, { status: 400 });
    }

    return Response.json({
      success: false,
      error: 'Failed to copy file'
    }, { status: 500 });
  }
}