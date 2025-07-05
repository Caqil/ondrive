import { NextRequest } from 'next/server';
import { Folder } from '@/models/Folder';
import { connectDB } from '@/lib/db';
import { apiRateLimiter, createRateLimitMiddleware } from '@/lib/rate-limit';
import { getServerSession } from 'next-auth';
import { authOptions } from '@/lib/auth';
import { getClientIP, formatHeaders } from '@/lib/api-utils';
import type { BaseResponse, FolderTreeNode, User } from '@/types';

// GET /api/client/folders/tree - Get complete folder tree structure
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
    
    // Parse query parameters
    const maxDepth = parseInt(url.searchParams.get('maxDepth') || '10');
    const includeTrash = url.searchParams.get('includeTrash') === 'true';
    const parentId = url.searchParams.get('parentId'); // For partial tree loading

    // Build query based on user permissions
    const baseQuery: any = {
      isTrashed: includeTrash ? { $in: [true, false] } : false,
      $or: [
        { owner: user._id },
        { visibility: 'public' },
        { team: user.currentTeam, visibility: 'team' }
      ]
    };

    // Limit depth to prevent performance issues
    if (maxDepth && maxDepth < 20) {
      baseQuery.depth = { $lte: maxDepth };
    }

    // If parentId is specified, load only that subtree
    if (parentId) {
      baseQuery.$or = [
        { _id: parentId },
        { ancestors: parentId }
      ];
    }

    // Get all folders that match criteria
    const folders = await Folder.find(baseQuery)
      .sort({ path: 1 })
      .lean();

    // Build tree structure
    const tree = buildFolderTree(folders, parentId);

    const response: BaseResponse<FolderTreeNode[]> = {
      success: true,
      data: tree
    };

    return Response.json(response, {
      headers: formatHeaders(rateLimitResult.headers)
    });

  } catch (error: any) {
    console.error('Get folder tree error:', error);
    
    return Response.json({
      success: false,
      error: 'Failed to load folder tree'
    }, { status: 500 });
  }
}

// Helper function to build folder tree structure
function buildFolderTree(folders: any[], rootId: string | null = null): FolderTreeNode[] {
  const folderMap = new Map();
  const tree: FolderTreeNode[] = [];

  // Create nodes
  folders.forEach(folder => {
    const node: FolderTreeNode = {
      id: folder._id.toString(),
      name: folder.name,
      path: folder.path,
      depth: folder.depth,
      children: [],
      fileCount: folder.fileCount,
      folderCount: folder.folderCount,
      totalSize: folder.totalSize,
      isExpanded: false
    };
    folderMap.set(folder._id.toString(), node);
  });

  // Build tree relationships
  folders.forEach(folder => {
    const node = folderMap.get(folder._id.toString());
    if (!node) return;

    if (!folder.parent || folder.parent.toString() === rootId) {
      tree.push(node);
    } else {
      const parent = folderMap.get(folder.parent.toString());
      if (parent) {
        parent.children.push(node);
      }
    }
  });

  return tree;
}