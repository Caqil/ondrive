import { z } from 'zod';

// Payment Method - matches models/Payment.ts exactly
const paymentMethodSchema = z.object({
  type: z.enum(['card', 'bank_transfer', 'paypal', 'crypto', 'wallet', 'other']),
  brand: z.string().optional(),
  last4: z.string().length(4, 'Last 4 digits must be exactly 4 characters').optional(),
  expiryMonth: z.number().min(1).max(12).optional(),
  expiryYear: z.number().min(new Date().getFullYear()).optional(),
  holderName: z.string().optional(),
  fingerprint: z.string().optional()
});

// Tax Details - matches models/Payment.ts
const taxDetailsSchema = z.object({
  taxRate: z.number().min(0).max(100).default(0),
  taxType: z.string().default('VAT'),
  taxId: z.string().optional(),
  taxRegion: z.string().optional()
});

// Billing Address - matches models/Payment.ts
const billingAddressSchema = z.object({
  line1: z.string().min(1, 'Address line 1 is required').max(100),
  line2: z.string().max(100).optional(),
  city: z.string().min(1, 'City is required').max(50),
  state: z.string().max(50).optional(),
  postalCode: z.string().min(1, 'Postal code is required').max(20),
  country: z.string().length(2, 'Country code must be exactly 2 characters').toUpperCase()
});

// Customer Info - matches models/Payment.ts
const customerInfoSchema = z.object({
  email: z.string().email('Invalid email address'),
  name: z.string().min(1, 'Name is required'),
  phone: z.string().optional(),
  ipAddress: z.string().ip('Invalid IP address').optional(),
  userAgent: z.string().optional()
});

// Applied Coupon - matches models/Payment.ts
const appliedCouponSchema = z.object({
  couponId: objectIdSchema,
  code: z.string().min(1).toUpperCase(),
  discountAmount: z.number().min(0),
  discountType: z.enum(['percentage', 'fixed_amount'])
});

// Refund Request - matches models/Payment.ts refunds structure
export const refundPaymentSchema = z.object({
  paymentId: objectIdSchema,
  amount: z.number().min(1, 'Refund amount must be greater than 0').optional(), // If not provided, full refund
  reason: z.string().min(5, 'Refund reason must be at least 5 characters').max(500, 'Reason cannot exceed 500 characters'),
  refundId: z.string().optional(), // External refund ID
  notifyCustomer: z.boolean().default(true)
});

// Dispute Information - matches models/Payment.ts
export const updateDisputeSchema = z.object({
  paymentId: objectIdSchema,
  disputeId: z.string().min(1, 'Dispute ID is required'),
  reason: z.string().min(1, 'Dispute reason is required'),
  status: z.enum(['warning_needs_response', 'warning_under_review', 'warning_closed', 'needs_response', 'under_review', 'charge_refunded', 'won', 'lost']),
  amount: z.number().min(0),
  evidence: z.record(z.any()).optional(),
  dueBy: z.string().datetime().optional()
});

// Risk Assessment - matches models/Payment.ts
export const updateRiskAssessmentSchema = z.object({
  paymentId: objectIdSchema,
  riskScore: z.number().min(0).max(100),
  riskLevel: z.enum(['low', 'medium', 'high']),
  fraudCheck: z.object({
    passed: z.boolean(),
    score: z.number().min(0).max(100),
    checks: z.object({
      cvv: z.boolean(),
      address: z.boolean(),
      zip: z.boolean()
    })
  }).optional()
});

// Webhook Event - matches models/Payment.ts
export const recordWebhookEventSchema = z.object({
  paymentId: objectIdSchema,
  eventType: z.string().min(1, 'Event type is required'),
  eventId: z.string().min(1, 'Event ID is required'),
  processed: z.boolean().default(false)
});

// Payment Reconciliation - matches models/Payment.ts
export const reconcilePaymentSchema = z.object({
  paymentId: objectIdSchema,
  reconciledBy: objectIdSchema,
  notes: z.string().max(1000, 'Notes cannot exceed 1000 characters').optional()
});

// Create Payment - comprehensive
export const createPaymentSchema = z.object({
  subscriptionId: objectIdSchema.optional(),
  invoiceId: objectIdSchema.optional(),
  userId: objectIdSchema,
  teamId: objectIdSchema.optional(),
  type: z.enum(['subscription', 'one_time', 'refund', 'adjustment', 'credit']),
  description: z.string().min(1, 'Description is required').max(500),
  amount: z.number().min(50, 'Amount must be at least $0.50').max(999999999), // in cents
  currency: currencySchema,
  subtotal: z.number().min(0),
  taxAmount: z.number().min(0).default(0),
  feeAmount: z.number().min(0).default(0),
  discountAmount: z.number().min(0).default(0),
  taxDetails: taxDetailsSchema,
  paymentMethod: paymentMethodSchema,
  provider: z.enum(['stripe', 'paypal', 'paddle', 'lemonsqueezy', 'razorpay', 'manual']),
  providerId: z.string().min(1, 'Provider payment ID is required'),
  providerCustomerId: z.string().optional(),
  billingAddress: billingAddressSchema,
  customerInfo: customerInfoSchema,
  appliedCoupons: z.array(appliedCouponSchema).optional(),
  metadata: z.record(z.any()).optional()
});

// Payment Filters - comprehensive
export const paymentFiltersSchema = z.object({
  status: z.enum(['pending', 'processing', 'succeeded', 'failed', 'cancelled', 'refunded', 'partially_refunded', 'disputed']).optional(),
  type: z.enum(['subscription', 'one_time', 'refund', 'adjustment', 'credit']).optional(),
  provider: z.enum(['stripe', 'paypal', 'paddle', 'lemonsqueezy', 'razorpay', 'manual']).optional(),
  currency: currencySchema.optional(),
  userId: objectIdSchema.optional(),
  teamId: objectIdSchema.optional(),
  subscriptionId: objectIdSchema.optional(),
  reconciled: z.boolean().optional(),
  dateRange: z.object({
    start: z.string().datetime().optional(),
    end: z.string().datetime().optional()
  }).optional(),
  amountRange: z.object({
    min: z.number().min(0).optional(),
    max: z.number().min(0).optional()
  }).optional(),
  riskLevel: z.enum(['low', 'medium', 'high']).optional(),
  hasDispute: z.boolean().optional(),
  hasRefund: z.boolean().optional(),
  page: z.number().min(1).default(1),
  limit: z.number().min(1).max(100).default(20),
  sort: z.enum(['paymentNumber', 'amount', 'processedAt', 'status', 'createdAt']).default('processedAt'),
  order: z.enum(['asc', 'desc']).default('desc')
});

export type CreatePaymentRequest = z.infer<typeof createPaymentSchema>;
export type RefundPaymentRequest = z.infer<typeof refundPaymentSchema>;
export type UpdateDisputeRequest = z.infer<typeof updateDisputeSchema>;
export type UpdateRiskAssessmentRequest = z.infer<typeof updateRiskAssessmentSchema>;
export type RecordWebhookEventRequest = z.infer<typeof recordWebhookEventSchema>;
export type ReconcilePaymentRequest = z.infer<typeof reconcilePaymentSchema>;
export type PaymentFiltersRequest = z.infer<typeof paymentFiltersSchema>;
