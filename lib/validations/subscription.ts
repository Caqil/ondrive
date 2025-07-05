import { z } from 'zod';

const currencySchema = z.enum(['USD', 'EUR', 'GBP', 'CAD', 'AUD', 'JPY', 'CHF', 'CNY']);

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
  trialDays: z.number().min(0).max(365).optional(),
  features: subscriptionFeaturesSchema.optional()
});

// Update Subscription
export const updateSubscriptionSchema = z.object({
  status: z.enum(['trial', 'active', 'past_due', 'cancelled', 'expired']).optional(),
  currentPeriodStart: z.string().datetime().optional(),
  currentPeriodEnd: z.string().datetime().optional(),
  nextBillingDate: z.string().datetime().optional(),
  features: subscriptionFeaturesSchema.optional(),
  usage: subscriptionUsageSchema.optional()
});

// Upgrade/Downgrade Subscription
export const upgradeSubscriptionSchema = z.object({
  newPlanId: objectIdSchema,
  interval: z.enum(['month', 'year']).optional(),
  prorationBehavior: z.enum(['create_prorations', 'none']).default('create_prorations'),
  effectiveDate: z.enum(['now', 'end_of_period']).default('now')
});

// Usage Tracking and Limits
export const updateUsageSchema = z.object({
  subscriptionId: objectIdSchema,
  storageUsed: z.number().min(0).optional(),
  apiRequestsUsed: z.number().min(0).optional(),
  period: z.enum(['add', 'set']).default('add') // Add to current or set absolute value
});

// Usage History Request
export const usageHistorySchema = z.object({
  period: z.enum(['daily', 'monthly', 'yearly']).default('monthly'),
  start: z.string().datetime().optional(),
  end: z.string().datetime().optional(),
  metrics: z.array(z.enum(['storage', 'apiRequests', 'members', 'costs'])).optional()
});

// Plan Management (Admin)
export const createPlanSchema = z.object({
  name: z.string().min(1, 'Plan name is required').max(100, 'Plan name cannot exceed 100 characters'),
  description: z.string().max(500, 'Description cannot exceed 500 characters'),
  isActive: z.boolean().default(true),
  isPublic: z.boolean().default(true),
  sortOrder: z.number().min(0).default(0),
  trialDays: z.number().min(0).max(365).default(14),
  
  // Pricing
  prices: z.object({
    monthly: z.object({
      amount: z.number().min(0, 'Amount cannot be negative'),
      currency: currencySchema,
      providerId: z.object({
        stripe: z.string().optional(),
        paypal: z.string().optional(),
        paddle: z.string().optional(),
        lemonsqueezy: z.string().optional()
      }).optional()
    }),
    yearly: z.object({
      amount: z.number().min(0, 'Amount cannot be negative'),
      currency: currencySchema,
      providerId: z.object({
        stripe: z.string().optional(),
        paypal: z.string().optional(),
        paddle: z.string().optional(),
        lemonsqueezy: z.string().optional()
      }).optional()
    })
  }),
  
  // Features - matches models/Plan.ts exactly
  features: z.object({
    storageLimit: z.number().min(0),
    memberLimit: z.number().min(1),
    fileUploadLimit: z.number().min(0),
    apiRequestLimit: z.number().min(0),
    enableAdvancedSharing: z.boolean().default(false),
    enableVersionHistory: z.boolean().default(true),
    enableOCR: z.boolean().default(false),
    enablePrioritySupport: z.boolean().default(false),
    enableAPIAccess: z.boolean().default(false),
    enableIntegrations: z.boolean().default(false),
    enableCustomBranding: z.boolean().default(false),
    enableAuditLogs: z.boolean().default(false),
    enableSSO: z.boolean().default(false)
  })
});

// Subscription Filters
export const subscriptionFiltersSchema = z.object({
  subscriberType: z.enum(['user', 'team']).optional(),
  status: z.enum(['trial', 'active', 'past_due', 'cancelled', 'expired']).optional(),
  provider: z.enum(['stripe', 'paypal', 'paddle', 'lemonsqueezy']).optional(),
  interval: z.enum(['month', 'year']).optional(),
  planId: objectIdSchema.optional(),
  dateRange: z.object({
    start: z.string().datetime().optional(),
    end: z.string().datetime().optional()
  }).optional(),
  amountRange: z.object({
    min: z.number().min(0).optional(),
    max: z.number().min(0).optional()
  }).optional(),
  page: z.number().min(1).default(1),
  limit: z.number().min(1).max(100).default(20),
  sort: z.enum(['createdAt', 'currentPeriodEnd', 'amount', 'status']).default('createdAt'),
  order: z.enum(['asc', 'desc']).default('desc')
});

export type CreateSubscriptionRequest = z.infer<typeof createSubscriptionSchema>;
export type UpdateSubscriptionRequest = z.infer<typeof updateSubscriptionSchema>;
export type UpgradeSubscriptionRequest = z.infer<typeof upgradeSubscriptionSchema>;
export type UpdateUsageRequest = z.infer<typeof updateUsageSchema>;
export type UsageHistoryRequest = z.infer<typeof usageHistorySchema>;
export type CreatePlanRequest = z.infer<typeof createPlanSchema>;
export type SubscriptionFiltersRequest = z.infer<typeof subscriptionFiltersSchema>;
