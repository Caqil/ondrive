import { NextRequest } from 'next/server';
import { Folder } from '@/models/Folder';
import { connectDB } from '@/lib/db';
import { createFolderSchema } from '@/lib/validations/file';
import { apiRateLimiter, createRateLimitMiddleware } from '@/lib/rate-limit';
import { getServerSession } from 'next-auth';
import { authOptions } from '@/lib/auth';
import { getClientIP, formatHeaders } from '@/lib/api-utils';
import { generateSafeFilename } from '@/lib/file-utils';
import type { BaseResponse, Folder as FolderType, User } from '@/types';

// GET /api/client/folders - List folders with pagination and filters
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
    const url = new URL(request.url);
    
    const parentId = url.searchParams.get('parentId');
    const page = parseInt(url.searchParams.get('page') || '1');
    const limit = parseInt(url.searchParams.get('limit') || '50');
    const sort = url.searchParams.get('sort') || 'name';
    const order = url.searchParams.get('order') || 'asc';
    const includeTrash = url.searchParams.get('includeTrash') === 'true';

    // Validate parameters
    if (limit > 100) {
      return Response.json({
        success: false,
        error: 'Limit cannot exceed 100'
      }, { status: 400 });
    }

    // Build query based on user permissions
    const query: any = {
      isTrashed: includeTrash ? { $in: [true, false] } : false,
      $or: [
        { owner: user._id },
        { visibility: 'public' },
        { team: user.currentTeam, visibility: 'team' }
      ]
    };

    // Filter by parent folder
    if (parentId) {
      query.parent = parentId;
    } else {
      query.parent = null; // Root level folders
    }

    // Execute query with pagination
    const skip = (page - 1) * limit;
    const sortOrder = order === 'desc' ? -1 : 1;
    
    const [folders, total] = await Promise.all([
      Folder.find(query)
        .populate('owner', 'name email avatar')
        .populate('parent', 'name path')
        .sort({ [sort]: sortOrder })
        .skip(skip)
        .limit(limit)
        .lean(),
      Folder.countDocuments(query)
    ]);

    const response: BaseResponse<{
      folders: any[];
      pagination: {
        page: number;
        limit: number;
        total: number;
        pages: number;
      };
    }> = {
      success: true,
      data: {
        folders,
        pagination: {
          page,
          limit,
          total,
          pages: Math.ceil(total / limit)
        }
      }
    };

    return Response.json(response, {
      headers: formatHeaders(rateLimitResult.headers)
    });

  } catch (error: any) {
    console.error('List folders error:', error);
    
    return Response.json({
      success: false,
      error: 'Failed to load folders'
    }, { status: 500 });
  }
}

// POST /api/client/folders - Create new folder
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
    const { name, description, parentId, color, icon } = createFolderSchema.parse(body);

    // Verify parent folder exists and user has access
    let parentFolder: any = null;
    let depth = 0;
    let ancestors: any[] = [];
    let path = '';

    if (parentId) {
      parentFolder = await Folder.findOne({
        _id: parentId,
        $or: [
          { owner: user._id },
          { team: user.currentTeam, visibility: { $in: ['team', 'public'] } }
        ],
        isTrashed: false
      });

      if (!parentFolder) {
        return Response.json({
          success: false,
          error: 'Parent folder not found or access denied'
        }, { status: 404 });
      }

      // Check folder depth limit
      if (parentFolder.depth >= 19) { // Max depth 20
        return Response.json({
          success: false,
          error: 'Maximum folder depth exceeded'
        }, { status: 400 });
      }

      depth = parentFolder.depth + 1;
      ancestors = [...(parentFolder.ancestors || []), parentFolder._id];
      path = `${parentFolder.path}/${generateSafeFilename(name)}`;
    } else {
      path = `/${generateSafeFilename(name)}`;
    }

    // Check for duplicate folder names in the same parent
    const existingFolder = await Folder.findOne({
      name,
      parent: parentId || null,
      owner: user._id,
      isTrashed: false
    });

    if (existingFolder) {
      return Response.json({
        success: false,
        error: 'A folder with this name already exists in the selected location'
      }, { status: 409 });
    }

    // Create folder
    const folderData = {
      name,
      description,
      path,
      depth,
      parent: parentId || undefined,
      ancestors,
      owner: user._id,
      team: user.currentTeam,
      visibility: parentFolder?.visibility || 'private',
      color,
      icon,
      fileCount: 0,
      folderCount: 0,
      totalSize: 0,
      children: [],
      isStarred: false,
      isTrashed: false,
      shareCount: 0,
      syncStatus: 'synced'
    };

    const folder = new Folder(folderData);
    await folder.save();

    // Update parent folder's child count
    if (parentFolder) {
      await Folder.findByIdAndUpdate(parentId, {
        $inc: { folderCount: 1 },
        $push: { children: folder._id }
      });
    }

    // Populate references for response
    await folder.populate('owner', 'name email avatar');
    if (folder.parent) {
      await folder.populate('parent', 'name path');
    }

    const response: BaseResponse<any> = {
      success: true,
      data: folder.toObject()
    };

    return Response.json(response, {
      headers: formatHeaders(rateLimitResult.headers)
    });

  } catch (error: any) {
    console.error('Create folder error:', error);
    
    if (error.name === 'ZodError') {
      return Response.json({
        success: false,
        error: 'Invalid request data',
        details: error.errors
      }, { status: 400 });
    }

    return Response.json({
      success: false,
      error: 'Failed to create folder'
    }, { status: 500 });
  }
}