import { NextRequest } from 'next/server';
import { Share } from '@/models/Share';
import { File } from '@/models/File';
import { Folder } from '@/models/Folder';
import { connectDB } from '@/lib/db';
import { updateShareSchema } from '@/lib/validations/share';
import { apiRateLimiter, createRateLimitMiddleware } from '@/lib/rate-limit';
import { getServerSession } from 'next-auth';
import { authOptions } from '@/lib/auth';
import { getClientIP, formatHeaders } from '@/lib/api-utils';
import type { BaseResponse, User } from '@/types';

// GET /api/client/share/[shareId] - Get share details
export async function GET(
  request: NextRequest,
  { params }: { params: { shareId: string } }
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
    const { shareId } = params;

    // Find share with access control
    const share = await Share.findOne({
      _id: shareId,
      $or: [
        { owner: user._id },
        { sharedBy: user._id }
      ]
    })
    .populate('resource')
    .populate('owner', 'name email avatar')
    .populate('sharedBy', 'name email avatar')
    .lean();

    if (!share) {
      return Response.json({
        success: false,
        error: 'Share not found or access denied'
      }, { status: 404 });
    }

    const response: BaseResponse<any> = {
      success: true,
      data: share
    };

    return Response.json(response, {
      headers: formatHeaders(rateLimitResult.headers)
    });

  } catch (error: any) {
    console.error('Get share error:', error);
    
    return Response.json({
      success: false,
      error: 'Failed to load share'
    }, { status: 500 });
  }
}

// PUT /api/client/share/[shareId] - Update share
export async function PUT(
  request: NextRequest,
  { params }: { params: { shareId: string } }
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
    const { shareId } = params;
    const body = await request.json();
    
    // Validate request data
    const updates = updateShareSchema.parse(body);

    // Find share and verify ownership
    const share = await Share.findOne({
      _id: shareId,
      $or: [
        { owner: user._id },
        { sharedBy: user._id }
      ]
    });

    if (!share) {
      return Response.json({
        success: false,
        error: 'Share not found or access denied'
      }, { status: 404 });
    }

    // Hash new password if provided
    if (updates.password) {
      const { hashPassword } = await import('@/lib/crypto');
      (updates as any).password = await hashPassword(updates.password);
    }

    // Update expiration status if expiresAt is changed
    if (updates.expiresAt) {
      (updates as any).isExpired = new Date(updates.expiresAt) < new Date();
    }

    // Update share
    const updatedShare = await Share.findByIdAndUpdate(
      shareId,
      { ...updates, updatedAt: new Date() },
      { new: true }
    )
    .populate('resource')
    .populate('owner', 'name email avatar')
    .populate('sharedBy', 'name email avatar');

    const response: BaseResponse<any> = {
      success: true,
      data: updatedShare?.toObject()
    };

    return Response.json(response, {
      headers: formatHeaders(rateLimitResult.headers)
    });

  } catch (error: any) {
    console.error('Update share error:', error);
    
    if (error.name === 'ZodError') {
      return Response.json({
        success: false,
        error: 'Invalid request data',
        details: error.errors
      }, { status: 400 });
    }

    return Response.json({
      success: false,
      error: 'Failed to update share'
    }, { status: 500 });
  }
}

// DELETE /api/client/share/[shareId] - Delete share
export async function DELETE(
  request: NextRequest,
  { params }: { params: { shareId: string } }
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
    const { shareId } = params;

    // Find share and verify ownership
    const share = await Share.findOne({
      _id: shareId,
      $or: [
        { owner: user._id },
        { sharedBy: user._id }
      ]
    });

    if (!share) {
      return Response.json({
        success: false,
        error: 'Share not found or access denied'
      }, { status: 404 });
    }

    // Delete share
    await Share.findByIdAndDelete(shareId);

    // Update resource share count
    if (share.resourceType === 'file') {
      await File.findByIdAndUpdate(share.resource, {
        $inc: { shareCount: -1 }
      });
    } else {
      await Folder.findByIdAndUpdate(share.resource, {
        $inc: { shareCount: -1 }
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
    console.error('Delete share error:', error);
    
    return Response.json({
      success: false,
      error: 'Failed to delete share'
    }, { status: 500 });
  }
}