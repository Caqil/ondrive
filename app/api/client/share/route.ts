import { NextRequest } from 'next/server';
import { Share } from '@/models/Share';
import { File } from '@/models/File';
import { Folder } from '@/models/Folder';
import { connectDB } from '@/lib/db';
import { createShareSchema, shareFiltersSchema } from '@/lib/validations/share';
import { apiRateLimiter, createRateLimitMiddleware } from '@/lib/rate-limit';
import { getServerSession } from 'next-auth';
import { authOptions } from '@/lib/auth';
import { getClientIP, formatHeaders } from '@/lib/api-utils';
import { generateToken } from '@/lib/crypto';
import type { BaseResponse, User } from '@/types';

// GET /api/client/share - List user's shares
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
    
    // Parse and validate query parameters
    const queryParams: any = {};
    url.searchParams.forEach((value, key) => {
      queryParams[key] = value;
    });
    
    const {
      page = 1,
      limit = 20,
      resourceType,
      type,
      permission,
      isActive,
      isExpired,
      sort = 'createdAt',
      order = 'desc'
    } = shareFiltersSchema.parse(queryParams);

    // Build query
    const query: any = {
      owner: user._id
    };

    if (resourceType) query.resourceType = resourceType;
    if (type) query.type = type;
    if (permission) query.permission = permission;
    if (isActive !== undefined) query.isActive = isActive;
    if (isExpired !== undefined) query.isExpired = isExpired;

    // Execute query with pagination
    const skip = (page - 1) * limit;
    const sortOrder = order === 'desc' ? -1 : 1;

    const [shares, total] = await Promise.all([
      Share.find(query)
        .populate('resource')
        .populate('owner', 'name email avatar')
        .populate('sharedBy', 'name email avatar')
        .sort({ [sort]: sortOrder })
        .skip(skip)
        .limit(limit)
        .lean(),
      Share.countDocuments(query)
    ]);

    const response: BaseResponse<{
      shares: any[];
      pagination: {
        page: number;
        limit: number;
        total: number;
        pages: number;
      };
    }> = {
      success: true,
      data: {
        shares,
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
    console.error('List shares error:', error);
    
    return Response.json({
      success: false,
      error: 'Failed to load shares'
    }, { status: 500 });
  }
}

// POST /api/client/share - Create new share
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
    const shareData = createShareSchema.parse(body);

    // Verify resource exists and user has access
    let resource;
    if (shareData.resourceType === 'file') {
      resource = await File.findOne({
        _id: shareData.resourceId,
        $or: [
          { owner: user._id },
          { team: user.currentTeam, visibility: { $in: ['team', 'public'] } }
        ],
        isTrashed: false
      });
    } else {
      resource = await Folder.findOne({
        _id: shareData.resourceId,
        $or: [
          { owner: user._id },
          { team: user.currentTeam, visibility: { $in: ['team', 'public'] } }
        ],
        isTrashed: false
      });
    }

    if (!resource) {
      return Response.json({
        success: false,
        error: 'Resource not found or access denied'
      }, { status: 404 });
    }

    // Check if user can share this resource
    if (resource.owner.toString() !== user._id.toString() && resource.visibility === 'private') {
      return Response.json({
        success: false,
        error: 'You do not have permission to share this resource'
      }, { status: 403 });
    }

    // Generate unique share token
    const token = generateToken(32);

    // Hash password if provided
    let hashedPassword;
    if (shareData.password) {
      const { hashPassword } = await import('@/lib/crypto');
      hashedPassword = await hashPassword(shareData.password);
    }

    // Create share
    const share = new Share({
      token,
      resource: shareData.resourceId,
      resourceType: shareData.resourceType,
      owner: resource.owner,
      sharedBy: user._id,
      type: shareData.type,
      permission: shareData.permission,
      allowDownload: shareData.allowDownload ?? true,
      allowPrint: shareData.allowPrint ?? true,
      allowCopy: shareData.allowCopy ?? true,
      requireAuth: shareData.requireAuth ?? false,
      password: hashedPassword,
      expiresAt: shareData.expiresAt ? new Date(shareData.expiresAt) : undefined,
      allowedDomains: shareData.allowedDomains || [],
      allowedUsers: shareData.allowedUsers || [],
      accessCount: 0,
      accessLog: [],
      isActive: true,
      isRevoked: false,
      isExpired: false
    });

    await share.save();

    // Update resource share count
    if (shareData.resourceType === 'file') {
      await File.findByIdAndUpdate(shareData.resourceId, {
        $inc: { shareCount: 1 }
      });
    } else {
      await Folder.findByIdAndUpdate(shareData.resourceId, {
        $inc: { shareCount: 1 }
      });
    }

    // Populate for response
    await share.populate('resource');
    await share.populate('owner', 'name email avatar');
    await share.populate('sharedBy', 'name email avatar');

    const response: BaseResponse<any> = {
      success: true,
      data: share.toObject()
    };

    return Response.json(response, {
      headers: formatHeaders(rateLimitResult.headers)
    });

  } catch (error: any) {
    console.error('Create share error:', error);
    
    if (error.name === 'ZodError') {
      return Response.json({
        success: false,
        error: 'Invalid request data',
        details: error.errors
      }, { status: 400 });
    }

    return Response.json({
      success: false,
      error: 'Failed to create share'
    }, { status: 500 });
  }
}