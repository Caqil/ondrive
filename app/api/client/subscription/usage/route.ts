import { NextRequest } from 'next/server';
import { Subscription } from '@/models/Subscription';
import { File } from '@/models/File';
import { connectDB } from '@/lib/db';
import { updateUsageSchema } from '@/lib/validations/subscription';
import { apiRateLimiter, createRateLimitMiddleware } from '@/lib/rate-limit';
import { getServerSession } from 'next-auth';
import { authOptions } from '@/lib/auth';
import { getClientIP, formatHeaders } from '@/lib/api-utils';
import type { BaseResponse, User } from '@/types';

// GET /api/client/subscription/usage - Get detailed usage information
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
    
    const subscriberType = url.searchParams.get('type') || 'user';
    const period = url.searchParams.get('period') || 'current'; // current, last30days, all

    // Find current subscription
    const subscriptionQuery = await Subscription.findOne({
      subscriber: user._id,
      subscriberType,
      status: { $in: ['trial', 'active', 'past_due'] }
    })
    .populate('plan', 'name features')
    .lean();

    const subscription = subscriptionQuery as any;

    if (!subscription) {
      return Response.json({
        success: false,
        error: 'No active subscription found'
      }, { status: 404 });
    }

    // Calculate actual storage usage from files
    const storageAggregation = await File.aggregate([
      {
        $match: {
          owner: user._id,
          isTrashed: false
        }
      },
      {
        $group: {
          _id: null,
          totalSize: { $sum: '$size' },
          totalFiles: { $sum: 1 },
          filesByType: {
            $push: {
              mimeType: '$mimeType',
              size: '$size'
            }
          }
        }
      }
    ]);

    const actualStorageUsed = storageAggregation[0]?.totalSize || 0;
    const totalFiles = storageAggregation[0]?.totalFiles || 0;

    // Group files by type for breakdown
    const filesByType: Record<string, { count: number; size: number }> = {};
    if (storageAggregation[0]?.filesByType) {
      storageAggregation[0].filesByType.forEach((file: any) => {
        const category = getFileCategory(file.mimeType);
        if (!filesByType[category]) {
          filesByType[category] = { count: 0, size: 0 };
        }
        filesByType[category].count++;
        filesByType[category].size += file.size;
      });
    }

    // Calculate usage percentages
    const storagePercentage = subscription.features.storageLimit > 0 
      ? (actualStorageUsed / subscription.features.storageLimit) * 100 
      : 0;

    const apiRequestsPercentage = subscription.features.apiRequestLimit > 0
      ? (subscription.usage.apiRequestsUsed / subscription.features.apiRequestLimit) * 100
      : 0;

    // Calculate usage trends (simplified for demo)
    const now = new Date();
    const thirtyDaysAgo = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000);
    
    // Get storage usage over time (last 30 days)
    const storageHistory = await File.aggregate([
      {
        $match: {
          owner: user._id,
          createdAt: { $gte: thirtyDaysAgo },
          isTrashed: false
        }
      },
      {
        $group: {
          _id: {
            $dateToString: {
              format: "%Y-%m-%d",
              date: "$createdAt"
            }
          },
          dailySize: { $sum: '$size' },
          dailyFiles: { $sum: 1 }
        }
      },
      {
        $sort: { _id: 1 }
      }
    ]);

    // Calculate warnings and recommendations
    const warnings: { type: string; level: string; message: string }[] = [];
    const recommendations: { type: string; message: string }[] = [];

    if (storagePercentage > 90) {
      warnings.push({
        type: 'storage',
        level: 'critical',
        message: 'Storage usage is critical. Consider upgrading your plan or removing unused files.'
      });
    } else if (storagePercentage > 80) {
      warnings.push({
        type: 'storage',
        level: 'warning',
        message: 'Storage usage is high. You may want to clean up old files or upgrade your plan.'
      });
    }

    if (apiRequestsPercentage > 90) {
      warnings.push({
        type: 'api',
        level: 'critical',
        message: 'API usage is critical. Consider upgrading your plan for higher limits.'
      });
    }

    if (storagePercentage > 50) {
      recommendations.push({
        type: 'cleanup',
        message: 'Consider organizing files and removing duplicates to free up space.'
      });
    }

    if (totalFiles > 1000) {
      recommendations.push({
        type: 'organization',
        message: 'Use folders to better organize your growing file collection.'
      });
    }

    // Determine if user is approaching limits
    const approachingLimits = {
      storage: storagePercentage > 75,
      apiRequests: apiRequestsPercentage > 75
    };

    const response: BaseResponse<{
      subscription: {
        id: string;
        plan: any;
        status: string;
        features: any;
      };
      usage: {
        storage: {
          used: number;
          limit: number;
          percentage: number;
          remaining: number;
          breakdown: Record<string, { count: number; size: number }>;
        };
        apiRequests: {
          used: number;
          limit: number;
          percentage: number;
          remaining: number;
        };
        files: {
          total: number;
          breakdown: Record<string, { count: number; size: number }>;
        };
      };
      trends: {
        storageHistory: any[];
      };
      warnings: any[];
      recommendations: any[];
      approachingLimits: any;
      lastUpdated: string;
    }> = {
      success: true,
      data: {
        subscription: {
          id: subscription._id.toString(),
          plan: subscription.plan,
          status: subscription.status,
          features: subscription.features
        },
        usage: {
          storage: {
            used: actualStorageUsed,
            limit: subscription.features.storageLimit,
            percentage: Math.min(100, storagePercentage),
            remaining: Math.max(0, subscription.features.storageLimit - actualStorageUsed),
            breakdown: filesByType
          },
          apiRequests: {
            used: subscription.usage.apiRequestsUsed,
            limit: subscription.features.apiRequestLimit,
            percentage: Math.min(100, apiRequestsPercentage),
            remaining: Math.max(0, subscription.features.apiRequestLimit - subscription.usage.apiRequestsUsed)
          },
          files: {
            total: totalFiles,
            breakdown: filesByType
          }
        },
        trends: {
          storageHistory
        },
        warnings,
        recommendations,
        approachingLimits,
        lastUpdated: new Date().toISOString()
      }
    };

    return Response.json(response, {
      headers: formatHeaders(rateLimitResult.headers)
    });

  } catch (error: any) {
    console.error('Get usage error:', error);
    
    return Response.json({
      success: false,
      error: 'Failed to load usage information'
    }, { status: 500 });
  }
}

// PUT /api/client/subscription/usage - Update usage (admin/system use)
export async function PUT(request: NextRequest) {
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
    const { storageUsed, apiRequestsUsed, resetUsage } = updateUsageSchema.parse(body);

    // Find current subscription
    const subscriptionQuery = await Subscription.findOne({
      subscriber: user._id,
      subscriberType: 'user',
      status: { $in: ['trial', 'active', 'past_due'] }
    });

    const subscription = subscriptionQuery as any;

    if (!subscription) {
      return Response.json({
        success: false,
        error: 'No active subscription found'
      }, { status: 404 });
    }

    // Prepare update data
    const updateData: any = {};

    if (resetUsage) {
      updateData['usage.storageUsed'] = 0;
      updateData['usage.apiRequestsUsed'] = 0;
      updateData['usage.lastResetAt'] = new Date();
    } else {
      if (storageUsed !== undefined) {
        updateData['usage.storageUsed'] = storageUsed;
      }
      if (apiRequestsUsed !== undefined) {
        updateData['usage.apiRequestsUsed'] = apiRequestsUsed;
      }
    }

    // Update subscription
    const updatedSubscription = await Subscription.findByIdAndUpdate(
      subscription._id,
      { $set: updateData },
      { new: true }
    )
    .populate('plan', 'name features');

    const response: BaseResponse<any> = {
      success: true,
      data: {
        subscription: updatedSubscription?.toObject(),
        message: resetUsage ? 'Usage counters reset successfully' : 'Usage updated successfully'
      }
    };

    return Response.json(response, {
      headers: formatHeaders(rateLimitResult.headers)
    });

  } catch (error: any) {
    console.error('Update usage error:', error);
    
    if (error.name === 'ZodError') {
      return Response.json({
        success: false,
        error: 'Invalid request data',
        details: error.errors
      }, { status: 400 });
    }

    return Response.json({
      success: false,
      error: 'Failed to update usage'
    }, { status: 500 });
  }
}

// Helper function to categorize files by MIME type
function getFileCategory(mimeType: string): string {
  if (mimeType.startsWith('image/')) return 'images';
  if (mimeType.startsWith('video/')) return 'videos';
  if (mimeType.startsWith('audio/')) return 'audio';
  if (mimeType.includes('pdf')) return 'pdfs';
  if (mimeType.includes('document') || mimeType.includes('word')) return 'documents';
  if (mimeType.includes('spreadsheet') || mimeType.includes('excel')) return 'spreadsheets';
  if (mimeType.includes('presentation') || mimeType.includes('powerpoint')) return 'presentations';
  if (mimeType.includes('text/') || mimeType.includes('code')) return 'text';
  if (mimeType.includes('archive') || mimeType.includes('zip')) return 'archives';
  return 'other';
}