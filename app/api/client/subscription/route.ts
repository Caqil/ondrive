import { NextRequest } from 'next/server';
import { Subscription } from '@/models/Subscription';
import { Plan } from '@/models/Plan';
import { User } from '@/models/User';
import { connectDB } from '@/lib/db';
import { createSubscriptionSchema, subscriptionFiltersSchema } from '@/lib/validations/subscription';
import { apiRateLimiter, createRateLimitMiddleware } from '@/lib/rate-limit';
import { getServerSession } from 'next-auth';
import { authOptions } from '@/lib/auth';
import { getClientIP, formatHeaders } from '@/lib/api-utils';
import type { BaseResponse, User as UserType, Subscription as SubscriptionType } from '@/types';

// GET /api/client/subscription - Get user's current subscription
export async function GET(request: NextRequest) {
  try {
    const session = await getServerSession(authOptions);
    if (!session?.user) {
      return Response.json({
        success: false,
        error: 'Authentication required'
      }, { status: 401 });
    }

    const user = session.user as UserType;
    const clientIP = getClientIP(request);
    const rateLimitResult = createRateLimitMiddleware(apiRateLimiter)(clientIP);
    
    await connectDB();
    const url = new URL(request.url);
    
    // Parse query parameters for filtering
    const queryParams: any = {};
    url.searchParams.forEach((value, key) => {
      queryParams[key] = value;
    });

    const {
      subscriberType = 'user',
      includeHistory = false,
      includeUsage = true
    } = queryParams;

    // Find current subscription
    const subscriptionQuery = await Subscription.findOne({
      subscriber: user._id,
      subscriberType,
      status: { $in: ['trial', 'active', 'past_due'] }
    })
    .populate('plan', 'name description features prices')
    .sort({ createdAt: -1 })
    .lean();

    const subscription = subscriptionQuery as any;

    let subscriptionHistory: SubscriptionType[] = [];
    if (includeHistory === 'true') {
      const historyQuery = await Subscription.find({
        subscriber: user._id,
        subscriberType,
        status: { $in: ['cancelled', 'expired'] }
      })
      .populate('plan', 'name description')
      .sort({ createdAt: -1 })
      .limit(10)
      .lean();

      subscriptionHistory = historyQuery as any[];
    }

    // Calculate usage percentages if subscription exists
    let usageData: {
      storage: {
        used: number;
        limit: number;
        percentage: number;
        remaining: number;
      };
      apiRequests: {
        used: number;
        limit: number;
        percentage: number;
        remaining: number;
      };
      lastResetAt: Date;
    } | null = null;
    if (subscription && includeUsage === 'true') {
      const storagePercentage = subscription.features.storageLimit > 0 
        ? (subscription.usage.storageUsed / subscription.features.storageLimit) * 100 
        : 0;

      const apiRequestsPercentage = subscription.features.apiRequestLimit > 0
        ? (subscription.usage.apiRequestsUsed / subscription.features.apiRequestLimit) * 100
        : 0;

      usageData = {
        storage: {
          used: subscription.usage.storageUsed,
          limit: subscription.features.storageLimit,
          percentage: Math.min(100, storagePercentage),
          remaining: Math.max(0, subscription.features.storageLimit - subscription.usage.storageUsed)
        },
        apiRequests: {
          used: subscription.usage.apiRequestsUsed,
          limit: subscription.features.apiRequestLimit,
          percentage: Math.min(100, apiRequestsPercentage),
          remaining: Math.max(0, subscription.features.apiRequestLimit - subscription.usage.apiRequestsUsed)
        },
        lastResetAt: subscription.usage.lastResetAt
      };
    }

    // Determine subscription status
    let subscriptionStatus = 'none';
    let daysUntilRenewal: number | null = null;
    let isTrialExpiring = false;

    if (subscription) {
      subscriptionStatus = subscription.status;
      
      if (subscription.status === 'trial' && subscription.trialEnd) {
        const trialEndDate = new Date(subscription.trialEnd);
        const now = new Date();
        daysUntilRenewal = Math.ceil((trialEndDate.getTime() - now.getTime()) / (1000 * 60 * 60 * 24));
        isTrialExpiring = daysUntilRenewal <= 3 && daysUntilRenewal > 0;
      } else if (subscription.nextBillingDate) {
        const billingDate = new Date(subscription.nextBillingDate);
        const now = new Date();
        daysUntilRenewal = Math.ceil((billingDate.getTime() - now.getTime()) / (1000 * 60 * 60 * 24));
      }
    }

    const response: BaseResponse<{
      subscription: any;
      usage?: any;
      status: string;
      daysUntilRenewal?: number | null;
      isTrialExpiring: boolean;
      history?: any[];
    }> = {
      success: true,
      data: {
        subscription,
        usage: usageData,
        status: subscriptionStatus,
        daysUntilRenewal,
        isTrialExpiring,
        ...(includeHistory === 'true' && { history: subscriptionHistory })
      }
    };

    return Response.json(response, {
      headers: formatHeaders(rateLimitResult.headers)
    });

  } catch (error: any) {
    console.error('Get subscription error:', error);
    
    return Response.json({
      success: false,
      error: 'Failed to load subscription'
    }, { status: 500 });
  }
}

// POST /api/client/subscription - Create new subscription
export async function POST(request: NextRequest) {
  try {
    const session = await getServerSession(authOptions);
    if (!session?.user) {
      return Response.json({
        success: false,
        error: 'Authentication required'
      }, { status: 401 });
    }

    const user = session.user as UserType;
    const clientIP = getClientIP(request);
    const rateLimitResult = createRateLimitMiddleware(apiRateLimiter)(clientIP);
    
    await connectDB();
    const body = await request.json();
    
    // Validate request data
    const subscriptionData = createSubscriptionSchema.parse({
      ...body,
      subscriberId: user._id,
      subscriberType: 'user'
    });

    // Check if user already has an active subscription
    const existingSubscription = await Subscription.findOne({
      subscriber: user._id,
      subscriberType: 'user',
      status: { $in: ['trial', 'active', 'past_due'] }
    });

    if (existingSubscription) {
      return Response.json({
        success: false,
        error: 'User already has an active subscription'
      }, { status: 409 });
    }

    // Verify plan exists and is active
    const planQuery = await Plan.findOne({
      _id: subscriptionData.planId,
      isActive: true,
      isPublic: true
    });

    const plan = planQuery as any;

    if (!plan) {
      return Response.json({
        success: false,
        error: 'Plan not found or not available'
      }, { status: 404 });
    }

    // Get plan pricing based on interval
    const planPrice = subscriptionData.interval === 'year' ? plan.prices.yearly : plan.prices.monthly;
    
    if (!planPrice) {
      return Response.json({
        success: false,
        error: 'Pricing not available for selected interval'
      }, { status: 400 });
    }

    // Calculate subscription period
    const now = new Date();
    const trialDays = subscriptionData.trialDays || plan.trialDays || 0;
    
    let trialStart: Date | undefined;
    let trialEnd: Date | undefined;
    let currentPeriodStart: Date;
    let currentPeriodEnd: Date;
    let status: 'trial' | 'active' = 'active';

    if (trialDays > 0) {
      trialStart = now;
      trialEnd = new Date(now.getTime() + trialDays * 24 * 60 * 60 * 1000);
      currentPeriodStart = trialEnd;
      status = 'trial';
    } else {
      currentPeriodStart = subscriptionData.startDate ? new Date(subscriptionData.startDate) : now;
    }

    // Calculate period end based on interval
    currentPeriodEnd = new Date(currentPeriodStart);
    if (subscriptionData.interval === 'year') {
      currentPeriodEnd.setFullYear(currentPeriodEnd.getFullYear() + subscriptionData.intervalCount);
    } else {
      currentPeriodEnd.setMonth(currentPeriodEnd.getMonth() + subscriptionData.intervalCount);
    }

    // In a real implementation, you would:
    // 1. Create customer in payment provider (Stripe, etc.)
    // 2. Create subscription in payment provider
    // 3. Handle payment method and initial payment
    // For this demo, we'll create a mock subscription

    const subscriptionRecord = {
      subscriber: user._id,
      subscriberType: 'user',
      plan: plan._id,
      provider: subscriptionData.provider,
      providerId: `mock_${Date.now()}`, // Would be real provider ID
      customerId: `cus_${Date.now()}`, // Would be real customer ID
      status,
      currentPeriodStart,
      currentPeriodEnd,
      trialStart,
      trialEnd,
      currency: planPrice.currency,
      amount: planPrice.amount,
      interval: subscriptionData.interval,
      intervalCount: subscriptionData.intervalCount,
      features: {
        storageLimit: plan.features.storageLimit,
        memberLimit: plan.features.memberLimit,
        fileUploadLimit: plan.features.fileUploadLimit,
        apiRequestLimit: plan.features.apiRequestLimit,
        enableAdvancedFeatures: plan.features.enableAdvancedSharing
      },
      usage: {
        storageUsed: 0,
        apiRequestsUsed: 0,
        lastResetAt: now
      },
      nextBillingDate: status === 'trial' ? currentPeriodEnd : new Date(currentPeriodEnd.getTime() + 24 * 60 * 60 * 1000)
    };

    const subscription = new Subscription(subscriptionRecord);
    await subscription.save();

    // Update user subscription status
    await User.findByIdAndUpdate(user._id, {
      subscriptionStatus: status,
      ...(trialEnd && { trialEndsAt: trialEnd }),
      storageQuota: plan.features.storageLimit
    });

    // Populate plan details for response
    await subscription.populate('plan', 'name description features prices');

    const response: BaseResponse<any> = {
      success: true,
      data: subscription.toObject()
    };

    return Response.json(response, {
      headers: formatHeaders(rateLimitResult.headers)
    });

  } catch (error: any) {
    console.error('Create subscription error:', error);
    
    if (error.name === 'ZodError') {
      return Response.json({
        success: false,
        error: 'Invalid request data',
        details: error.errors
      }, { status: 400 });
    }

    return Response.json({
      success: false,
      error: 'Failed to create subscription'
    }, { status: 500 });
  }
}