import { NextRequest } from 'next/server';
import { File } from '@/models/File';
import { connectDB } from '@/lib/db';
import { fileSearchSchema } from '@/lib/validations/file';
import { apiRateLimiter, createRateLimitMiddleware } from '@/lib/rate-limit';
import { getServerSession } from 'next-auth';
import { authOptions } from '@/lib/auth';
import { getClientIP, formatHeaders } from '@/lib/api-utils';
import type { PaginatedResponse, File as FileType, User } from '@/types';

// GET /api/client/files - List files with pagination and filters
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
    
    await connectDB();
    const { searchParams } = new URL(request.url);
    
    // Parse and validate query parameters
    const queryParams = Object.fromEntries(searchParams.entries());
    const validatedParams = fileSearchSchema.parse(queryParams);
    
    const {
      page = 1,
      limit = 20,
      folderId,
      mimeType,
      size,
      dateRange,
      tags,
      owner,
      isStarred,
      isTrashed = false,
      sort = 'updatedAt',
      order = 'desc'
    } = validatedParams;

    // Build query based on user permissions
    const query: any = {
      isTrashed,
      $or: [
        { owner: user._id },
        { visibility: 'public' },
        { team: user.currentTeam, visibility: 'team' }
      ]
    };

    if (folderId) query.folder = folderId;
    if (mimeType) query.mimeType = new RegExp(mimeType, 'i');
    if (size?.min || size?.max) {
      query.size = {};
      if (size.min) query.size.$gte = size.min;
      if (size.max) query.size.$lte = size.max;
    }
    if (dateRange?.start || dateRange?.end) {
      query.createdAt = {};
      if (dateRange.start) query.createdAt.$gte = new Date(dateRange.start);
      if (dateRange.end) query.createdAt.$lte = new Date(dateRange.end);
    }
    if (tags?.length) query.tags = { $in: tags };
    if (owner) query.owner = owner;
    if (isStarred !== undefined) query.isStarred = isStarred;

    // Execute query with pagination
    const skip = (page - 1) * limit;
    const [files, total] = await Promise.all([
      File.find(query)
        .sort({ [sort]: order === 'asc' ? 1 : -1 })
        .skip(skip)
        .limit(limit)
        .populate('owner', 'name email avatar')
        .populate('folder', 'name path')
        .lean(),
      File.countDocuments(query)
    ]);

    const response: PaginatedResponse<FileType> = {
      success: true,
      data: files as unknown as FileType[],
      pagination: {
        page,
        limit,
        total,
        totalPages: Math.ceil(total / limit),
        hasNext: page < Math.ceil(total / limit),
        hasPrev: page > 1
      }
    };

    return Response.json(response, {
      headers: formatHeaders(rateLimitResult.headers)
    });
  } catch (error: any) {
    console.error('Files list error:', error);
    
    if (error.name === 'ZodError') {
      return Response.json({
        success: false,
        error: 'Invalid query parameters',
        details: error.errors
      }, { status: 400 });
    }

    return Response.json({
      success: false,
      error: 'Failed to fetch files'
    }, { status: 500 });
  }
}