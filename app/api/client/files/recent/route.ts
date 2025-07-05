import { NextRequest } from 'next/server';
import { File } from '@/models/File';
import { connectDB } from '@/lib/db';
import { apiRateLimiter, createRateLimitMiddleware } from '@/lib/rate-limit';
import { getServerSession } from 'next-auth';
import { authOptions } from '@/lib/auth';
import { getClientIP, formatHeaders } from '@/lib/api-utils';
import type { BaseResponse, File as FileType, User } from '@/types';

// GET /api/client/files/recent - Get recently accessed/modified files
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
    const limit = parseInt(searchParams.get('limit') || '20');
    const type = searchParams.get('type') || 'accessed'; // 'accessed', 'modified', 'created'
    const days = parseInt(searchParams.get('days') || '30');

    // Validate parameters
    if (limit > 100) {
      return Response.json({
        success: false,
        error: 'Limit cannot exceed 100'
      }, { status: 400 });
    }

    if (!['accessed', 'modified', 'created'].includes(type)) {
      return Response.json({
        success: false,
        error: 'Invalid type parameter'
      }, { status: 400 });
    }

    // Calculate date threshold
    const dateThreshold = new Date();
    dateThreshold.setDate(dateThreshold.getDate() - days);

    // Build query based on user permissions
    const baseQuery = {
      isTrashed: false,
      processingStatus: 'completed',
      $or: [
        { owner: user._id },
        { visibility: 'public' },
        { team: user.currentTeam, visibility: 'team' }
      ]
    };

    let sortField: string;
    let dateFilter: any = {};

    switch (type) {
      case 'accessed':
        sortField = 'lastAccessedAt';
        dateFilter = {
          lastAccessedAt: { $gte: dateThreshold, $exists: true }
        };
        break;
      case 'modified':
        sortField = 'updatedAt';
        dateFilter = {
          updatedAt: { $gte: dateThreshold }
        };
        break;
      case 'created':
        sortField = 'createdAt';
        dateFilter = {
          createdAt: { $gte: dateThreshold }
        };
        break;
      default:
        sortField = 'lastAccessedAt';
        dateFilter = {
          lastAccessedAt: { $gte: dateThreshold, $exists: true }
        };
    }

    const query = {
      ...baseQuery,
      ...dateFilter
    };

    // For accessed files, prioritize files that have been accessed
    // For other types, show all files within the date range
    const files = await File.find(query)
      .populate('folder', 'name path')
      .populate('owner', 'name email avatar')
      .sort({ [sortField]: -1 })
      .limit(limit)
      .lean();

    // If we don't have enough recent accessed files, fallback to recently modified
    let finalFiles = files;
    if (type === 'accessed' && files.length < limit) {
      const fallbackQuery = {
        ...baseQuery,
        updatedAt: { $gte: dateThreshold },
        _id: { $nin: files.map(f => f._id) }
      };

      const fallbackFiles = await File.find(fallbackQuery)
        .populate('folder', 'name path')
        .populate('owner', 'name email avatar')
        .sort({ updatedAt: -1 })
        .limit(limit - files.length)
        .lean();

      finalFiles = [...files, ...fallbackFiles];
    }

    // Transform files to include computed properties
    const transformedFiles = finalFiles.map((file: any) => ({
      ...file,
      _id: file._id.toString(),
      folder: file.folder ? {
        ...file.folder,
        _id: file.folder._id.toString()
      } : null,
      owner: file.owner ? {
        ...file.owner,
        _id: file.owner._id.toString()
      } : null,
      // Add human-readable relative time
      relativeTime: getRelativeTime(file[sortField] || file.createdAt),
      // Add file type category for better grouping
      category: getFileCategory(file.mimeType),
      // Add file size in human readable format
      sizeFormatted: formatFileSize(file.size)
    }));

    // Group files by date for better UI organization
    const groupedFiles = groupFilesByDate(transformedFiles, sortField);

    const response: BaseResponse<{
      files: typeof transformedFiles;
      grouped: typeof groupedFiles;
      total: number;
      type: string;
      days: number;
    }> = {
      success: true,
      data: {
        files: transformedFiles,
        grouped: groupedFiles,
        total: transformedFiles.length,
        type,
        days
      }
    };

    return Response.json(response, {
      headers: formatHeaders(rateLimitResult.headers)
    });

  } catch (error: any) {
    console.error('Recent files error:', error);
    
    return Response.json({
      success: false,
      error: 'Failed to load recent files'
    }, { status: 500 });
  }
}

// PATCH /api/client/files/recent - Update file access timestamp
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
    const { fileId, action = 'view' } = body;

    if (!fileId) {
      return Response.json({
        success: false,
        error: 'File ID is required'
      }, { status: 400 });
    }

    // Verify file access
    const file = await File.findOne({
      _id: fileId,
      $or: [
        { owner: user._id },
        { visibility: 'public' },
        { team: user.currentTeam, visibility: 'team' }
      ],
      isTrashed: false
    });

    if (!file) {
      return Response.json({
        success: false,
        error: 'File not found or access denied'
      }, { status: 404 });
    }

    // Update access tracking
    const updateData: any = {
      lastAccessedAt: new Date()
    };

    // Increment appropriate counter
    switch (action) {
      case 'view':
        updateData.$inc = { viewCount: 1 };
        break;
      case 'download':
        updateData.$inc = { downloadCount: 1 };
        break;
      default:
        updateData.$inc = { viewCount: 1 };
    }

    await File.findByIdAndUpdate(fileId, updateData);

    const response: BaseResponse<{ success: boolean }> = {
      success: true,
      data: { success: true }
    };

    return Response.json(response, {
      headers: formatHeaders(rateLimitResult.headers)
    });

  } catch (error: any) {
    console.error('Update file access error:', error);
    
    return Response.json({
      success: false,
      error: 'Failed to update file access'
    }, { status: 500 });
  }
}

// Helper function to get relative time string
function getRelativeTime(date: Date): string {
  const now = new Date();
  const diffInSeconds = Math.floor((now.getTime() - new Date(date).getTime()) / 1000);

  if (diffInSeconds < 60) {
    return 'Just now';
  } else if (diffInSeconds < 3600) {
    const minutes = Math.floor(diffInSeconds / 60);
    return `${minutes} minute${minutes > 1 ? 's' : ''} ago`;
  } else if (diffInSeconds < 86400) {
    const hours = Math.floor(diffInSeconds / 3600);
    return `${hours} hour${hours > 1 ? 's' : ''} ago`;
  } else if (diffInSeconds < 604800) {
    const days = Math.floor(diffInSeconds / 86400);
    return `${days} day${days > 1 ? 's' : ''} ago`;
  } else {
    return new Date(date).toLocaleDateString();
  }
}

// Helper function to categorize files by MIME type
function getFileCategory(mimeType: string): string {
  if (mimeType.startsWith('image/')) return 'image';
  if (mimeType.startsWith('video/')) return 'video';
  if (mimeType.startsWith('audio/')) return 'audio';
  if (mimeType.includes('pdf')) return 'pdf';
  if (mimeType.includes('document') || mimeType.includes('word')) return 'document';
  if (mimeType.includes('spreadsheet') || mimeType.includes('excel')) return 'spreadsheet';
  if (mimeType.includes('presentation') || mimeType.includes('powerpoint')) return 'presentation';
  if (mimeType.includes('text/') || mimeType.includes('code')) return 'text';
  if (mimeType.includes('archive') || mimeType.includes('zip')) return 'archive';
  return 'other';
}

// Helper function to format file size
function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 Bytes';
  
  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

// Helper function to group files by date
function groupFilesByDate(files: any[], sortField: string) {
  const groups: Record<string, any[]> = {};
  const today = new Date();
  const yesterday = new Date(today.getTime() - 24 * 60 * 60 * 1000);
  const lastWeek = new Date(today.getTime() - 7 * 24 * 60 * 60 * 1000);

  files.forEach(file => {
    const fileDate = new Date(file[sortField] || file.createdAt);
    let groupKey: string;

    if (fileDate.toDateString() === today.toDateString()) {
      groupKey = 'Today';
    } else if (fileDate.toDateString() === yesterday.toDateString()) {
      groupKey = 'Yesterday';
    } else if (fileDate >= lastWeek) {
      groupKey = 'This Week';
    } else {
      groupKey = fileDate.toLocaleDateString('en-US', { 
        weekday: 'long', 
        year: 'numeric', 
        month: 'long', 
        day: 'numeric' 
      });
    }

    if (!groups[groupKey]) {
      groups[groupKey] = [];
    }
    groups[groupKey].push(file);
  });

  // Sort groups by date (most recent first)
  const sortedGroups = Object.entries(groups).sort(([a], [b]) => {
    const priority = { 'Today': 0, 'Yesterday': 1, 'This Week': 2 };
    return (priority[a] ?? 3) - (priority[b] ?? 3);
  });

  return Object.fromEntries(sortedGroups);
}