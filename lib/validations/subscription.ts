// lib/validations/subscription.ts
import { z } from 'zod';
import { objectIdSchema, currencySchema, metadataSchema } from './base';

// Subscription Features - matches models/Subscription.ts exactly
const subscriptionFeaturesSchema = z.object({
  storageLimit: z.number().min(0, 'Storage limit cannot be negative'),
  memberLimit: z.number().min(1, 'Member limit must be at least 1'),
  fileUploadLimit: z.number().min(0, 'File upload limit cannot be negative'),
  apiRequestLimit: z.number().min(0, 'API request limit cannot be negative'),
  enableAdvancedFeatures: z.boolean().default(false)
});

// Subscription Usage - matches models/Subscription.ts
const subscriptionUsageSchema = z.object({
  storageUsed: z.number().min(0, 'Storage used cannot be negative').default(0),
  apiRequestsUsed: z.number().min(0, 'API requests used cannot be negative').default(0),
  lastResetAt: z.string().datetime().default(new Date().toISOString())
});

// Create Subscription - comprehensive
export const createSubscriptionSchema = z.object({
  subscriberId: objectIdSchema,
  subscriberType: z.enum(['user', 'team']), // From models/Subscription.ts
  planId: objectIdSchema,
  interval: z.enum(['month', 'year']),
  intervalCount: z.number().min(1).default(1),
  currency: currencySchema.default('USD'),
  provider: z.enum(['stripe', 'paypal', 'paddle', 'lemonsqueezy']),
  paymentMethodId: z.string().min(1, 'Payment method is required').optional(),
  couponCode: z.string().max(50, 'Coupon code cannot exceed 50 characters').toUpperCase().optional(),
  trialDays: z.number().min(0).max(365).default(0),
  startDate: z.string().datetime().optional(),
  metadata: metadataSchema
});

// Update Subscription - for plan changes and modifications
export const updateSubscriptionSchema = z.object({
  planId: objectIdSchema.optional(),
  interval: z.enum(['month', 'year']).optional(),
  intervalCount: z.number().min(1).optional(),
  paymentMethodId: z.string().min(1, 'Payment method is required').optional(),
  prorationBehavior: z.enum(['create_prorations', 'none', 'always_invoice']).default('create_prorations'),
  billingCycleAnchor: z.string().datetime().optional(),
  metadata: metadataSchema
});

// Cancel Subscription - for subscription cancellation
export const cancelSubscriptionSchema = z.object({
  cancelAt: z.enum(['now', 'period_end']).default('period_end'),
  reason: z.enum([
    'customer_request',
    'payment_failed',
    'downgrade_to_free',
    'business_closure',
    'dissatisfaction',
    'too_expensive',
    'missing_features',
    'competitor',
    'other'
  ]).optional(),
  feedback: z.string().max(1000, 'Feedback cannot exceed 1000 characters').optional(),
  cancelImmediately: z.boolean().default(false),
  refundUnusedTime: z.boolean().default(false)
});

// Pause Subscription - for temporary suspension
export const pauseSubscriptionSchema = z.object({
  pauseUntil: z.string().datetime().optional(), // If not provided, pause indefinitely
  reason: z.string().max(500, 'Reason cannot exceed 500 characters').optional(),
  pauseBilling: z.boolean().default(true),
  notifyCustomer: z.boolean().default(true)
});

// Resume Subscription - for resuming paused subscriptions
export const resumeSubscriptionSchema = z.object({
  resumeAt: z.string().datetime().optional(), // If not provided, resume immediately
  prorateBilling: z.boolean().default(true),
  notifyCustomer: z.boolean().default(true)
});

// Subscription Filters - for listing subscriptions
export const subscriptionFiltersSchema = z.object({
  subscriberId: objectIdSchema.optional(),
  subscriberType: z.enum(['user', 'team']).optional(),
  planId: objectIdSchema.optional(),
  status: z.enum(['trial', 'active', 'past_due', 'cancelled', 'expired', 'paused']).optional(),
  provider: z.enum(['stripe', 'paypal', 'paddle', 'lemonsqueezy']).optional(),
  interval: z.enum(['month', 'year']).optional(),
  currency: currencySchema.optional(),
  amountRange: z.object({
    min: z.number().min(0).optional(),
    max: z.number().min(0).optional()
  }).optional(),
  trialEndRange: z.object({
    start: z.string().datetime().optional(),
    end: z.string().datetime().optional()
  }).optional(),
  billingDateRange: z.object({
    start: z.string().datetime().optional(),
    end: z.string().datetime().optional()
  }).optional(),
  createdDateRange: z.object({
    start: z.string().datetime().optional(),
    end: z.string().datetime().optional()
  }).optional(),
  query: z.string().max(200, 'Search query cannot exceed 200 characters').optional(),
  page: z.number().min(1).default(1),
  limit: z.number().min(1).max(100).default(20),
  sort: z.enum(['createdAt', 'currentPeriodEnd', 'nextBillingDate', 'amount', 'status']).default('createdAt'),
  order: z.enum(['asc', 'desc']).default('desc')
});

// Usage Update - for tracking subscription usage
export const updateUsageSchema = z.object({
  storageUsed: z.number().min(0, 'Storage used cannot be negative').optional(),
  apiRequestsUsed: z.number().min(0, 'API requests used cannot be negative').optional(),
  resetUsage: z.boolean().default(false) // Reset usage counters to 0
});

// Change Payment Method - for updating payment method
export const changePaymentMethodSchema = z.object({
  paymentMethodId: z.string().min(1, 'Payment method ID is required'),
  setAsDefault: z.boolean().default(true),
  processImmediately: z.boolean().default(false) // Whether to charge immediately for verification
});

// Apply Coupon - for applying coupons to existing subscriptions
export const applyCouponSchema = z.object({
  couponCode: z.string().min(1, 'Coupon code is required').max(50).toUpperCase(),
  validationOnly: z.boolean().default(false) // Just validate without applying
});

// Subscription Preview - for previewing changes before applying
export const subscriptionPreviewSchema = z.object({
  planId: objectIdSchema.optional(),
  interval: z.enum(['month', 'year']).optional(),
  couponCode: z.string().max(50).toUpperCase().optional(),
  prorationBehavior: z.enum(['create_prorations', 'none', 'always_invoice']).default('create_prorations'),
  billingCycleAnchor: z.string().datetime().optional()
});

// Upgrade/Downgrade Request - matches types/subscription.ts
export const upgradeSubscriptionSchema = z.object({
  newPlanId: objectIdSchema,
  interval: z.enum(['month', 'year']).optional(),
  prorationBehavior: z.enum(['create_prorations', 'none']).default('create_prorations'),
  effectiveDate: z.enum(['now', 'next_billing_cycle']).default('now'),
  paymentMethodId: z.string().optional() // Required for upgrades that increase cost
});

// Export feature and usage schemas for reuse
export {
  subscriptionFeaturesSchema,
  subscriptionUsageSchema
};

// Type exports
export type CreateSubscriptionRequest = z.infer<typeof createSubscriptionSchema>;
export type UpdateSubscriptionRequest = z.infer<typeof updateSubscriptionSchema>;
export type CancelSubscriptionRequest = z.infer<typeof cancelSubscriptionSchema>;
export type PauseSubscriptionRequest = z.infer<typeof pauseSubscriptionSchema>;
export type ResumeSubscriptionRequest = z.infer<typeof resumeSubscriptionSchema>;
export type SubscriptionFiltersRequest = z.infer<typeof subscriptionFiltersSchema>;
export type UpdateUsageRequest = z.infer<typeof updateUsageSchema>;
export type ChangePaymentMethodRequest = z.infer<typeof changePaymentMethodSchema>;
export type ApplyCouponRequest = z.infer<typeof applyCouponSchema>;
export type SubscriptionPreviewRequest = z.infer<typeof subscriptionPreviewSchema>;
export type UpgradeSubscriptionRequest = z.infer<typeof upgradeSubscriptionSchema>;