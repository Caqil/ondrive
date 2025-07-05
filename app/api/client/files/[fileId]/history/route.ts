import { NextRequest } from 'next/server';
import { File } from '@/models/File';
import { connectDB } from '@/lib/db';
import { apiRateLimiter, createRateLimitMiddleware } from '@/lib/rate-limit';
import { getServerSession } from 'next-auth';
import { authOptions } from '@/lib/auth';
import { getClientIP, formatHeaders } from '@/lib/api-utils';
import type { User } from '@/types';

// GET /api/client/files/[fileId]/history - Get file version history
export async function GET(
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
    
    await connectDB();

    const file = await File.findOne({
      _id: fileId,
      owner: user._id
    });

    if (!file) {
      return Response.json({
        success: false,
        error: 'File not found'
      }, { status: 404 });
    }

    // Get version history
    const versions = await File.find({
      _id: { $in: file.versionHistory }
    })
    .sort({ createdAt: -1 })
    .populate('owner', 'name email avatar')
    .lean();

    return Response.json({
      success: true,
      data: {
        currentVersion: file,
        versions
      }
    }, {
      headers: formatHeaders(rateLimitResult.headers)
    });
  } catch (error) {
    console.error('File history error:', error);
    return Response.json({
      success: false,
      error: 'Failed to fetch file history'
    }, { status: 500 });
  }
}
