import { NextRequest } from 'next/server';
import { Folder } from '@/models/Folder';
import { File } from '@/models/File';
import { connectDB } from '@/lib/db';
import { apiRateLimiter, createRateLimitMiddleware } from '@/lib/rate-limit';
import { getServerSession } from 'next-auth';
import { authOptions } from '@/lib/auth';
import { getClientIP, formatHeaders } from '@/lib/api-utils';
import type { BaseResponse, FolderContents, User } from '@/types';

// GET /api/client/folders/[folderId]/contents - Get folder contents (files and subfolders)
export async function GET(
  request: NextRequest,
  { params }: { params: { folderId: string } }
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
    
    await connectDB();
    const { folderId } = params;
    const url = new URL(request.url);
    
    // Parse query parameters
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

    // Verify folder access
    const folderQuery = await Folder.findOne({
      _id: folderId,
      $or: [
        { owner: user._id },
        { visibility: 'public' },
        { team: user.currentTeam, visibility: 'team' }
      ],
      isTrashed: false
    }).lean();

    const folder = folderQuery as any;

    if (!folder) {
      return Response.json({
        success: false,
        error: 'Folder not found or access denied'
      }, { status: 404 });
    }

    // Build base query for user permissions
    const baseQuery = {
      isTrashed: includeTrash ? { $in: [true, false] } : false,
      $or: [
        { owner: user._id },
        { visibility: 'public' },
        { team: user.currentTeam, visibility: 'team' }
      ]
    };

    const skip = (page - 1) * limit;
    const sortOrder = order === 'desc' ? -1 : 1;

    // Get files in folder
    const fileQuery = {
      ...baseQuery,
      folder: folderId
    };

    const files = await File.find(fileQuery)
      .populate('owner', 'name email avatar')
      .sort({ [sort]: sortOrder })
      .skip(skip)
      .limit(limit)
      .lean();

    // Get subfolders
    const subfolderQuery = {
      ...baseQuery,
      parent: folderId
    };

    const subfolders = subfolderQuery ? await Folder.find(subfolderQuery)
      .populate('owner', 'name email avatar')
      .sort({ [sort]: sortOrder })
      .lean() : [];

    // Generate breadcrumb
    const breadcrumb = await generateBreadcrumb(folderId, user);

    const folderContents: FolderContents = {
      folder: folder as any,
      subfolders: subfolders as any,
      files: files as any,
      breadcrumb
    };

    const response: BaseResponse<FolderContents> = {
      success: true,
      data: folderContents
    };

    return Response.json(response, {
      headers: formatHeaders(rateLimitResult.headers)
    });

  } catch (error: any) {
    console.error('Get folder contents error:', error);
    
    return Response.json({
      success: false,
      error: 'Failed to load folder contents'
    }, { status: 500 });
  }
}

// Helper function to generate breadcrumb navigation
async function generateBreadcrumb(folderId: string, user: User) {
  const breadcrumb: { id: string; name: string; path: string }[] = [];
  
  let currentFolderQuery = await Folder.findOne({
    _id: folderId,
    $or: [
      { owner: user._id },
      { visibility: 'public' },
      { team: user.currentTeam, visibility: 'team' }
    ]
  }).lean();

  let currentFolder = currentFolderQuery as any;

  while (currentFolder) {
    breadcrumb.unshift({
      id: currentFolder._id.toString(),
      name: currentFolder.name,
      path: currentFolder.path
    });

    if (currentFolder.parent) {
      const parentQuery = await Folder.findOne({
        _id: currentFolder.parent,
        $or: [
          { owner: user._id },
          { visibility: 'public' },
          { team: user.currentTeam, visibility: 'team' }
        ]
      }).lean();
      
      currentFolder = parentQuery as any;
    } else {
      break;
    }
  }

  return breadcrumb;
}