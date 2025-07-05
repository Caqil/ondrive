// lib/validations/payment.ts
import { z } from 'zod';
import { objectIdSchema, currencySchema, addressSchema, emailSchema, nameSchema, metadataSchema } from './base';

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
  email: emailSchema,
  name: nameSchema,
  phone: z.string().optional(),
  ipAddress: z.string().ip('Invalid IP address').optional(),
  userAgent: z.string().optional()
});

// Applied Coupon - matches models/Payment.ts
const appliedCouponSchema = z.object({
  couponId: objectIdSchema,
  code: z.string().min(1, 'Coupon code is required').max(50).toUpperCase(),
  discountAmount: z.number().min(0),
  discountType: z.enum(['percentage', 'fixed_amount'])
});

// Payment Event - matches models/Payment.ts
export const paymentEventSchema = z.object({
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
  metadata: metadataSchema
});

// Payment Filters - comprehensive
export const paymentFiltersSchema = z.object({
  status: z.enum(['pending', 'processing', 'succeeded', 'failed', 'cancelled', 'refunded']).optional(),
  type: z.enum(['subscription', 'one_time', 'refund', 'adjustment', 'credit']).optional(),
  provider: z.enum(['stripe', 'paypal', 'paddle', 'lemonsqueezy', 'razorpay', 'manual']).optional(),
  userId: objectIdSchema.optional(),
  teamId: objectIdSchema.optional(),
  subscriptionId: objectIdSchema.optional(),
  invoiceId: objectIdSchema.optional(),
  currency: currencySchema.optional(),
  paymentMethod: z.enum(['card', 'bank_transfer', 'paypal', 'crypto', 'wallet', 'other']).optional(),
  amountRange: z.object({
    min: z.number().min(0).optional(),
    max: z.number().min(0).optional()
  }).optional(),
  dateRange: z.object({
    start: z.string().datetime().optional(),
    end: z.string().datetime().optional()
  }).optional(),
  reconciled: z.boolean().optional(),
  query: z.string().max(200, 'Search query cannot exceed 200 characters').optional(),
  page: z.number().min(1).default(1),
  limit: z.number().min(1).max(100).default(20),
  sort: z.enum(['paymentNumber', 'amount', 'createdAt', 'processedAt', 'status']).default('createdAt'),
  order: z.enum(['asc', 'desc']).default('desc')
});

// Update Payment - for manual adjustments
export const updatePaymentSchema = z.object({
  status: z.enum(['pending', 'processing', 'succeeded', 'failed', 'cancelled', 'refunded']).optional(),
  description: z.string().min(1, 'Description is required').max(500).optional(),
  metadata: metadataSchema,
  notes: z.string().max(1000, 'Notes cannot exceed 1000 characters').optional()
});

// Refund Payment - for processing refunds
export const refundPaymentSchema = z.object({
  amount: z.number().min(1, 'Refund amount must be greater than 0').optional(), // If not provided, full refund
  reason: z.enum(['requested_by_customer', 'duplicate', 'fraudulent', 'subscription_canceled', 'other']),
  description: z.string().max(500, 'Refund description cannot exceed 500 characters').optional(),
  refundToOriginalMethod: z.boolean().default(true),
  notifyCustomer: z.boolean().default(true)
});

// Manual Payment - for recording manual payments
export const manualPaymentSchema = z.object({
  userId: objectIdSchema,
  invoiceId: objectIdSchema.optional(),
  subscriptionId: objectIdSchema.optional(),
  amount: z.number().min(1, 'Payment amount must be greater than 0'),
  currency: currencySchema,
  paymentMethod: z.enum(['cash', 'check', 'bank_transfer', 'other']),
  reference: z.string().max(100, 'Payment reference cannot exceed 100 characters').optional(),
  description: z.string().min(1, 'Description is required').max(500),
  paymentDate: z.string().datetime().default(new Date().toISOString()),
  notes: z.string().max(1000, 'Notes cannot exceed 1000 characters').optional()
});

// Payment Verification - for verifying external payments
export const verifyPaymentSchema = z.object({
  providerId: z.string().min(1, 'Provider payment ID is required'),
  provider: z.enum(['stripe', 'paypal', 'paddle', 'lemonsqueezy', 'razorpay']),
  webhookSignature: z.string().optional(),
  rawWebhookData: z.string().optional()
});

// Export payment method and other schemas for reuse
export {
  paymentMethodSchema,
  taxDetailsSchema,
  billingAddressSchema,
  customerInfoSchema,
  appliedCouponSchema
};

// Type exports
export type CreatePaymentRequest = z.infer<typeof createPaymentSchema>;
export type PaymentFiltersRequest = z.infer<typeof paymentFiltersSchema>;
export type UpdatePaymentRequest = z.infer<typeof updatePaymentSchema>;
export type RefundPaymentRequest = z.infer<typeof refundPaymentSchema>;
export type ManualPaymentRequest = z.infer<typeof manualPaymentSchema>;
export type VerifyPaymentRequest = z.infer<typeof verifyPaymentSchema>;
export type PaymentEventRequest = z.infer<typeof paymentEventSchema>;
export type ReconcilePaymentRequest = z.infer<typeof reconcilePaymentSchema>;