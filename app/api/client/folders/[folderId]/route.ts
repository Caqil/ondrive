import { NextRequest } from 'next/server';
import { Folder } from '@/models/Folder';
import { File } from '@/models/File';
import { connectDB } from '@/lib/db';
import { updateFolderSchema } from '@/lib/validations/file';
import { apiRateLimiter, createRateLimitMiddleware } from '@/lib/rate-limit';
import { getServerSession } from 'next-auth';
import { authOptions } from '@/lib/auth';
import { getClientIP, formatHeaders } from '@/lib/api-utils';
import { generateSafeFilename } from '@/lib/file-utils';
import type { BaseResponse, User } from '@/types';

// GET /api/client/folders/[folderId] - Get folder details
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

    // Find folder with access control
    const folderQuery = await Folder.findOne({
      _id: folderId,
      $or: [
        { owner: user._id },
        { visibility: 'public' },
        { team: user.currentTeam, visibility: 'team' }
      ],
      isTrashed: false
    })
    .populate('owner', 'name email avatar')
    .populate('parent', 'name path')
    .populate('children', 'name path fileCount folderCount totalSize')
    .lean();

    const folder = folderQuery as any;

    if (!folder) {
      return Response.json({
        success: false,
        error: 'Folder not found or access denied'
      }, { status: 404 });
    }

    const response: BaseResponse<any> = {
      success: true,
      data: folder
    };

    return Response.json(response, {
      headers: formatHeaders(rateLimitResult.headers)
    });

  } catch (error: any) {
    console.error('Get folder error:', error);
    
    return Response.json({
      success: false,
      error: 'Failed to load folder'
    }, { status: 500 });
  }
}

// PUT /api/client/folders/[folderId] - Update folder
export async function PUT(
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
    const body = await request.json();
    
    // Validate request data
    const updates = updateFolderSchema.parse(body);

    // Find folder and verify ownership
    const folderQuery = await Folder.findOne({
      _id: folderId,
      owner: user._id,
      isTrashed: false
    });

    const folder = folderQuery as any;

    if (!folder) {
      return Response.json({
        success: false,
        error: 'Folder not found or access denied'
      }, { status: 404 });
    }

    // Check for duplicate names if name is being updated
    if (updates.name && updates.name !== folder.name) {
      const existingFolder = await Folder.findOne({
        name: updates.name,
        parent: folder.parent,
        owner: user._id,
        isTrashed: false,
        _id: { $ne: folderId }
      });

      if (existingFolder) {
        return Response.json({
          success: false,
          error: 'A folder with this name already exists in the same location'
        }, { status: 409 });
      }

      // Update path if name changed
      const safeName = generateSafeFilename(updates.name);
      const newPath = folder.parent 
        ? `${folder.path.substring(0, folder.path.lastIndexOf('/'))}/${safeName}`
        : `/${safeName}`;
      
      (updates as any).path = newPath;

      // Update all descendant paths
      await updateDescendantPaths(folderId, folder.path, newPath);
    }

    // Update folder
    const updatedFolder = await Folder.findByIdAndUpdate(
      folderId,
      { ...updates, updatedAt: new Date() },
      { new: true }
    )
    .populate('owner', 'name email avatar')
    .populate('parent', 'name path');

    const response: BaseResponse<any> = {
      success: true,
      data: updatedFolder?.toObject()
    };

    return Response.json(response, {
      headers: formatHeaders(rateLimitResult.headers)
    });

  } catch (error: any) {
    console.error('Update folder error:', error);
    
    if (error.name === 'ZodError') {
      return Response.json({
        success: false,
        error: 'Invalid request data',
        details: error.errors
      }, { status: 400 });
    }

    return Response.json({
      success: false,
      error: 'Failed to update folder'
    }, { status: 500 });
  }
}

// DELETE /api/client/folders/[folderId] - Delete folder
export async function DELETE(
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
    const permanent = url.searchParams.get('permanent') === 'true';

    // Find folder and verify ownership
    const folderQuery = await Folder.findOne({
      _id: folderId,
      owner: user._id
    });

    const folder = folderQuery as any;

    if (!folder) {
      return Response.json({
        success: false,
        error: 'Folder not found or access denied'
      }, { status: 404 });
    }

    if (permanent) {
      // Permanent deletion - recursively delete all contents
      await deleteFolderPermanently(folderId, user._id.toString());
    } else {
      // Move to trash
      await moveToTrash(folderId, user._id.toString());
    }

    // Update parent folder counts
    if (folder.parent) {
      await Folder.findByIdAndUpdate(folder.parent, {
        $inc: { folderCount: -1 },
        $pull: { children: folderId }
      });
    }

    const response: BaseResponse<{ deleted: boolean }> = {
      success: true,
      data: { deleted: true }
    };

    return Response.json(response, {
      headers: formatHeaders(rateLimitResult.headers)
    });

  } catch (error: any) {
    console.error('Delete folder error:', error);
    
    return Response.json({
      success: false,
      error: 'Failed to delete folder'
    }, { status: 500 });
  }
}

// Helper function to update descendant folder paths
async function updateDescendantPaths(folderId: string, oldPath: string, newPath: string) {
  const descendants = await Folder.find({
    ancestors: folderId
  });

  for (const folder of descendants) {
    const updatedPath = folder.path.replace(oldPath, newPath);
    await Folder.findByIdAndUpdate(folder._id, { path: updatedPath });
  }
}

// Helper function to move folder to trash
async function moveToTrash(folderId: string, userId: string) {
  const updates = {
    isTrashed: true,
    trashedAt: new Date(),
    trashedBy: userId
  };

  await Folder.findByIdAndUpdate(folderId, updates);

  // Also move all files in the folder to trash
  await File.updateMany(
    { folder: folderId },
    updates
  );

  // Recursively move subfolders to trash
  const subfolders = await Folder.find({ parent: folderId });
  for (const subfolder of subfolders) {
    await moveToTrash(subfolder._id.toString(), userId);
  }
}

// Helper function to permanently delete folder
async function deleteFolderPermanently(folderId: string, userId: string) {
  // Delete all files in the folder
  await File.deleteMany({ folder: folderId });

  // Recursively delete subfolders
  const subfolders = await Folder.find({ parent: folderId });
  for (const subfolder of subfolders) {
    await deleteFolderPermanently(subfolder._id.toString(), userId);
  }

  // Delete the folder itself
  await Folder.findByIdAndDelete(folderId);
}