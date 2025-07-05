import { NextRequest } from 'next/server';
import { Subscription } from '@/models/Subscription';
import { Plan } from '@/models/Plan';
import { User } from '@/models/User';
import { connectDB } from '@/lib/db';
import { upgradeSubscriptionSchema, subscriptionPreviewSchema } from '@/lib/validations/subscription';
import { apiRateLimiter, createRateLimitMiddleware } from '@/lib/rate-limit';
import { getServerSession } from 'next-auth';
import { authOptions } from '@/lib/auth';
import { getClientIP, formatHeaders } from '@/lib/api-utils';
import type { BaseResponse, User as UserType } from '@/types';

// POST /api/client/subscription/upgrade - Upgrade subscription plan
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
    const upgradeData = upgradeSubscriptionSchema.parse(body);

    // Find current subscription
    const currentSubscriptionQuery = await Subscription.findOne({
      subscriber: user._id,
      subscriberType: 'user',
      status: { $in: ['trial', 'active', 'past_due'] }
    })
    .populate('plan', 'name features prices');

    const currentSubscription = currentSubscriptionQuery as any;

    if (!currentSubscription) {
      return Response.json({
        success: false,
        error: 'No active subscription found'
      }, { status: 404 });
    }

    // Verify new plan exists and is active
    const newPlanQuery = await Plan.findOne({
      _id: upgradeData.newPlanId,
      isActive: true,
      isPublic: true
    });

    const newPlan = newPlanQuery as any;

    if (!newPlan) {
      return Response.json({
        success: false,
        error: 'Target plan not found or not available'
      }, { status: 404 });
    }

    // Determine interval (use current if not specified)
    const interval = upgradeData.interval || currentSubscription.interval;
    const newPlanPrice = interval === 'year' ? newPlan.prices.yearly : newPlan.prices.monthly;

    if (!newPlanPrice) {
      return Response.json({
        success: false,
        error: 'Pricing not available for selected interval'
      }, { status: 400 });
    }

    // Check if this is actually an upgrade (higher price)
    const currentPrice = currentSubscription.amount;
    const newPrice = newPlanPrice.amount;
    const isUpgrade = newPrice > currentPrice;
    const isDowngrade = newPrice < currentPrice;

    if (isDowngrade) {
      return Response.json({
        success: false,
        error: 'Use the downgrade endpoint for plan downgrades'
      }, { status: 400 });
    }

    if (newPrice === currentPrice && upgradeData.newPlanId === currentSubscription.plan._id.toString()) {
      return Response.json({
        success: false,
        error: 'You are already on this plan'
      }, { status: 400 });
    }

    // Calculate proration amount
    const now = new Date();
    const currentPeriodEnd = new Date(currentSubscription.currentPeriodEnd);
    const remainingDays = Math.max(0, Math.ceil((currentPeriodEnd.getTime() - now.getTime()) / (1000 * 60 * 60 * 24)));
    const totalDays = interval === 'year' ? 365 : 30;
    
    let prorationAmount = 0;
    let immediateCharge = 0;

    if (upgradeData.effectiveDate === 'now' && upgradeData.prorationBehavior === 'create_prorations') {
      // Calculate prorated refund for current plan
      const currentDailyRate = currentPrice / totalDays;
      const currentRefund = Math.round(currentDailyRate * remainingDays);
      
      // Calculate prorated charge for new plan
      const newDailyRate = newPrice / totalDays;
      const newCharge = Math.round(newDailyRate * remainingDays);
      
      prorationAmount = newCharge - currentRefund;
      immediateCharge = isUpgrade ? prorationAmount : 0;
    } else if (upgradeData.effectiveDate === 'next_billing_cycle') {
      // No immediate charge, change takes effect at next billing cycle
      immediateCharge = 0;
    }

    // In a real implementation, you would:
    // 1. Update subscription in payment provider (Stripe, etc.)
    // 2. Handle proration and immediate charges
    // 3. Update payment method if required for upgrades
    // For this demo, we'll simulate the upgrade

    const updateData: any = {
      plan: newPlan._id,
      amount: newPrice,
      currency: newPlanPrice.currency,
      interval,
      features: {
        storageLimit: newPlan.features.storageLimit,
        memberLimit: newPlan.features.memberLimit,
        fileUploadLimit: newPlan.features.fileUploadLimit,
        apiRequestLimit: newPlan.features.apiRequestLimit,
        enableAdvancedFeatures: newPlan.features.enableAdvancedSharing
      },
      updatedAt: now
    };

    // Update billing dates if effective immediately
    if (upgradeData.effectiveDate === 'now') {
      updateData.currentPeriodStart = now;
      
      const newPeriodEnd = new Date(now);
      if (interval === 'year') {
        newPeriodEnd.setFullYear(newPeriodEnd.getFullYear() + 1);
      } else {
        newPeriodEnd.setMonth(newPeriodEnd.getMonth() + 1);
      }
      updateData.currentPeriodEnd = newPeriodEnd;
      updateData.nextBillingDate = newPeriodEnd;
    }

    // Update subscription
    const updatedSubscription = await Subscription.findByIdAndUpdate(
      currentSubscription._id,
      { $set: updateData },
      { new: true }
    )
    .populate('plan', 'name description features prices');

    // Update user storage quota
    await User.findByIdAndUpdate(user._id, {
      storageQuota: newPlan.features.storageLimit
    });

    // Create upgrade record/activity log (in real implementation)
    const upgradeRecord = {
      subscriptionId: currentSubscription._id,
      fromPlan: currentSubscription.plan.name,
      toPlan: newPlan.name,
      fromAmount: currentPrice,
      toAmount: newPrice,
      prorationAmount,
      immediateCharge,
      effectiveDate: upgradeData.effectiveDate,
      upgradedAt: now
    };

    const response: BaseResponse<{
      subscription: any;
      upgrade: any;
      billing: {
        prorationAmount: number;
        immediateCharge: number;
        nextBillingDate: string;
        newMonthlyPrice: number;
      };
      features: {
        newLimits: any;
        upgradedFeatures: string[];
      };
    }> = {
      success: true,
      data: {
        subscription: updatedSubscription?.toObject(),
        upgrade: upgradeRecord,
        billing: {
          prorationAmount,
          immediateCharge,
          nextBillingDate: updatedSubscription?.nextBillingDate?.toISOString() || '',
          newMonthlyPrice: newPrice
        },
        features: {
          newLimits: newPlan.features,
          upgradedFeatures: getUpgradedFeatures(currentSubscription.plan.features, newPlan.features)
        }
      }
    };

    return Response.json(response, {
      headers: formatHeaders(rateLimitResult.headers)
    });

  } catch (error: any) {
    console.error('Upgrade subscription error:', error);
    
    if (error.name === 'ZodError') {
      return Response.json({
        success: false,
        error: 'Invalid request data',
        details: error.errors
      }, { status: 400 });
    }

    return Response.json({
      success: false,
      error: 'Failed to upgrade subscription'
    }, { status: 500 });
  }
}

// GET /api/client/subscription/upgrade - Preview upgrade options
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
    
    const targetPlanId = url.searchParams.get('planId');
    const interval = url.searchParams.get('interval') || 'month';

    // Find current subscription
    const currentSubscriptionQuery = await Subscription.findOne({
      subscriber: user._id,
      subscriberType: 'user',
      status: { $in: ['trial', 'active', 'past_due'] }
    })
    .populate('plan', 'name features prices');

    const currentSubscription = currentSubscriptionQuery as any;

    if (!currentSubscription) {
      return Response.json({
        success: false,
        error: 'No active subscription found'
      }, { status: 404 });
    }

    // Get available upgrade plans
    let availablePlans;
    if (targetPlanId) {
      // Get specific plan for preview
      const planQuery = await Plan.findOne({
        _id: targetPlanId,
        isActive: true,
        isPublic: true
      });
      availablePlans = planQuery ? [planQuery] : [];
    } else {
      // Get all available plans
      const plansQuery = await Plan.find({
        isActive: true,
        isPublic: true
      }).sort({ sortOrder: 1 });
      availablePlans = plansQuery as any[];
    }

    // Calculate upgrade previews
    const upgradePreviews = availablePlans
      .map((plan: any) => {
        const planPrice = interval === 'year' ? plan.prices.yearly : plan.prices.monthly;
        if (!planPrice) return null;

        const currentPrice = currentSubscription.amount;
        const newPrice = planPrice.amount;
        const isUpgrade = newPrice > currentPrice;
        const savings = interval === 'year' && plan.prices.monthly 
          ? (plan.prices.monthly.amount * 12) - newPrice 
          : 0;

        // Calculate proration for immediate upgrade
        const now = new Date();
        const currentPeriodEnd = new Date(currentSubscription.currentPeriodEnd);
        const remainingDays = Math.max(0, Math.ceil((currentPeriodEnd.getTime() - now.getTime()) / (1000 * 60 * 60 * 24)));
        const totalDays = interval === 'year' ? 365 : 30;
        
        const currentDailyRate = currentPrice / totalDays;
        const currentRefund = Math.round(currentDailyRate * remainingDays);
        const newDailyRate = newPrice / totalDays;
        const newCharge = Math.round(newDailyRate * remainingDays);
        const prorationAmount = newCharge - currentRefund;

        return {
          plan: {
            id: plan._id.toString(),
            name: plan.name,
            description: plan.description,
            features: plan.features
          },
          pricing: {
            currentAmount: currentPrice,
            newAmount: newPrice,
            difference: newPrice - currentPrice,
            isUpgrade,
            savings: savings > 0 ? savings : 0,
            currency: planPrice.currency
          },
          proration: {
            amount: prorationAmount,
            immediateCharge: isUpgrade ? Math.max(0, prorationAmount) : 0,
            refund: !isUpgrade ? Math.abs(Math.min(0, prorationAmount)) : 0
          },
          features: {
            newLimits: plan.features,
            upgradedFeatures: getUpgradedFeatures(currentSubscription.plan.features, plan.features),
            downgradedFeatures: getDowngradedFeatures(currentSubscription.plan.features, plan.features)
          }
        };
      })
      .filter(Boolean);

    const response: BaseResponse<{
      currentSubscription: any;
      upgradePreviews: any[];
      recommendations: any[];
    }> = {
      success: true,
      data: {
        currentSubscription: {
          id: currentSubscription._id.toString(),
          plan: currentSubscription.plan,
          amount: currentSubscription.amount,
          interval: currentSubscription.interval,
          status: currentSubscription.status
        },
        upgradePreviews,
        recommendations: generateUpgradeRecommendations(currentSubscription, upgradePreviews)
      }
    };

    return Response.json(response, {
      headers: formatHeaders(rateLimitResult.headers)
    });

  } catch (error: any) {
    console.error('Get upgrade preview error:', error);
    
    return Response.json({
      success: false,
      error: 'Failed to load upgrade preview'
    }, { status: 500 });
  }
}

// Helper functions
function getUpgradedFeatures(currentFeatures: any, newFeatures: any): string[] {
  const upgrades: string[] = [];
  
  if (newFeatures.storageLimit > currentFeatures.storageLimit) {
    upgrades.push(`Increased storage from ${formatBytes(currentFeatures.storageLimit)} to ${formatBytes(newFeatures.storageLimit)}`);
  }
  
  if (newFeatures.memberLimit > currentFeatures.memberLimit) {
    upgrades.push(`Increased team members from ${currentFeatures.memberLimit} to ${newFeatures.memberLimit}`);
  }
  
  if (newFeatures.fileUploadLimit > currentFeatures.fileUploadLimit) {
    upgrades.push(`Increased file upload limit from ${formatBytes(currentFeatures.fileUploadLimit)} to ${formatBytes(newFeatures.fileUploadLimit)}`);
  }
  
  if (newFeatures.apiRequestLimit > currentFeatures.apiRequestLimit) {
    upgrades.push(`Increased API requests from ${currentFeatures.apiRequestLimit} to ${newFeatures.apiRequestLimit} per month`);
  }
  
  if (newFeatures.enableAdvancedFeatures && !currentFeatures.enableAdvancedFeatures) {
    upgrades.push('Access to advanced features');
  }
  
  return upgrades;
}

function getDowngradedFeatures(currentFeatures: any, newFeatures: any): string[] {
  const downgrades: string[] = [];
  
  if (newFeatures.storageLimit < currentFeatures.storageLimit) {
    downgrades.push(`Reduced storage from ${formatBytes(currentFeatures.storageLimit)} to ${formatBytes(newFeatures.storageLimit)}`);
  }
  
  if (newFeatures.memberLimit < currentFeatures.memberLimit) {
    downgrades.push(`Reduced team members from ${currentFeatures.memberLimit} to ${newFeatures.memberLimit}`);
  }
  
  return downgrades;
}

interface Recommendation {
  type: string;
  priority: string;
  message: string;
  suggestedPlan: any;
}

function generateUpgradeRecommendations(currentSubscription: any, upgradePreviews: any[]): Recommendation[] {
  const recommendations: Recommendation[] = [];
  
  // Check if user is approaching storage limit
  const storageUsagePercent = currentSubscription.usage?.storageUsed / currentSubscription.features.storageLimit * 100;
  
  if (storageUsagePercent > 80) {
    const nextTierWithMoreStorage = upgradePreviews.find(preview => 
      preview.plan.features.storageLimit > currentSubscription.features.storageLimit
    );
    
    if (nextTierWithMoreStorage) {
      recommendations.push({
        type: 'storage',
        priority: 'high',
        message: `You're using ${storageUsagePercent.toFixed(1)}% of your storage. Consider upgrading to ${nextTierWithMoreStorage.plan.name} for more space.`,
        suggestedPlan: nextTierWithMoreStorage.plan.id
      });
    }
  }
  
  return recommendations;
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 Bytes';
  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}