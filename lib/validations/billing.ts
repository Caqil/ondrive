import { z } from 'zod';

// Collection Attempt - matches models/Invoice.ts
const collectionAttemptSchema = z.object({
  attemptNumber: z.number().min(1),
  method: z.enum(['auto', 'manual']),
  result: z.enum(['succeeded', 'failed', 'pending']),
  failureReason: z.string().optional(),
  nextAttempt: z.string().datetime().optional()
});

// Communication Record - matches models/Invoice.ts
const communicationSchema = z.object({
  type: z.enum(['sent', 'viewed', 'downloaded', 'reminder', 'overdue_notice']),
  channel: z.enum(['email', 'sms', 'postal', 'portal']),
  deliveredAt: z.string().datetime().optional(),
  openedAt: z.string().datetime().optional(),
  metadata: z.record(z.any()).optional()
});

// Dunning Configuration - matches models/Invoice.ts
const dunningSchema = z.object({
  enabled: z.boolean().default(true),
  level: z.number().min(0).max(3).default(0),
  lastNoticeAt: z.string().datetime().optional(),
  nextNoticeAt: z.string().datetime().optional(),
  finalNoticeAt: z.string().datetime().optional(),
  collectionAgencyAt: z.string().datetime().optional()
});

// Workflow - matches models/Invoice.ts
const workflowSchema = z.object({
  requiresApproval: z.boolean().default(false),
  approvedBy: objectIdSchema.optional(),
  approvedAt: z.string().datetime().optional(),
  rejectedBy: objectIdSchema.optional(),
  rejectedAt: z.string().datetime().optional(),
  rejectionReason: z.string().max(500).optional()
});

// Invoice File - matches models/Invoice.ts
const invoiceFileSchema = z.object({
  type: z.enum(['pdf', 'xml', 'csv']),
  url: z.string().url('Invalid file URL'),
  size: z.number().min(0),
  generatedAt: z.string().datetime().default(new Date().toISOString())
});

// Tax Details - matches models/Invoice.ts
const invoiceTaxDetailsSchema = z.object({
  taxId: z.string().optional(),
  taxRegion: z.string().min(1, 'Tax region is required'),
  taxType: z.string().min(1, 'Tax type is required'),
  reverseCharge: z.boolean().default(false),
  taxExempt: z.boolean().default(false),
  taxExemptReason: z.string().optional()
});

// Legal Entity - matches models/Invoice.ts
const legalEntitySchema = z.object({
  name: z.string().min(1, 'Legal entity name is required'),
  address: z.record(z.any()),
  taxId: z.string().min(1, 'Tax ID is required'),
  registrationNumber: z.string().min(1, 'Registration number is required')
});

// Credit Note - matches models/Invoice.ts
const creditNoteSchema = z.object({
  originalInvoiceId: objectIdSchema,
  reason: z.string().min(1, 'Credit note reason is required'),
  type: z.enum(['full', 'partial'])
});

// Invoice Line Item - comprehensive
export const invoiceLineItemSchema = z.object({
  id: z.string().optional(), // Auto-generated if not provided
  description: z.string().min(1, 'Description is required').max(500),
  quantity: z.number().min(1).max(1000000),
  unitPrice: z.number().min(0).max(999999999),
  amount: z.number().min(0), // quantity * unitPrice
  taxRate: z.number().min(0).max(100).default(0),
  taxAmount: z.number().min(0).default(0),
  periodStart: z.string().datetime().optional(),
  periodEnd: z.string().datetime().optional(),
  planId: objectIdSchema.optional(),
  metadata: z.record(z.any()).optional()
});

// Create Invoice - comprehensive
export const createInvoiceSchema = z.object({
  subscriptionId: objectIdSchema.optional(),
  userId: objectIdSchema,
  teamId: objectIdSchema.optional(),
  type: z.enum(['subscription', 'one_time', 'credit_note', 'proforma', 'recurring']).default('one_time'),
  description: z.string().min(1, 'Description is required').max(1000),
  lineItems: z.array(invoiceLineItemSchema).min(1, 'At least one line item is required'),
  currency: currencySchema.default('USD'),
  exchangeRate: z.number().min(0).optional(),
  baseCurrency: currencySchema.optional(),
  taxDetails: invoiceTaxDetailsSchema,
  billingAddress: billingAddressSchema,
  invoiceDate: z.string().datetime().default(new Date().toISOString()),
  dueDate: z.string().datetime(),
  periodStart: z.string().datetime().optional(),
  periodEnd: z.string().datetime().optional(),
  appliedCoupons: z.array(appliedCouponSchema).optional(),
  creditNote: creditNoteSchema.optional(),
  legalEntity: legalEntitySchema.optional(),
  notes: z.string().max(2000).optional(),
  internalNotes: z.string().max(2000).optional(),
  tags: z.array(z.string().max(50)).optional(),
  metadata: z.record(z.any()).optional(),
  workflow: workflowSchema.optional(),
  dunning: dunningSchema.optional()
});

// Update Invoice
export const updateInvoiceSchema = z.object({
  description: z.string().max(1000).optional(),
  dueDate: z.string().datetime().optional(),
  status: z.enum(['draft', 'open', 'paid', 'void', 'uncollectible', 'overdue']).optional(),
  notes: z.string().max(2000).optional(),
  internalNotes: z.string().max(2000).optional(),
  tags: z.array(z.string().max(50)).optional(),
  workflow: workflowSchema.optional(),
  dunning: dunningSchema.optional()
});

// Invoice Collection Attempt
export const recordCollectionAttemptSchema = z.object({
  invoiceId: objectIdSchema,
  method: z.enum(['auto', 'manual']),
  result: z.enum(['succeeded', 'failed', 'pending']),
  failureReason: z.string().optional(),
  nextAttempt: z.string().datetime().optional()
});

// Invoice Communication
export const recordCommunicationSchema = z.object({
  invoiceId: objectIdSchema,
  type: z.enum(['sent', 'viewed', 'downloaded', 'reminder', 'overdue_notice']),
  channel: z.enum(['email', 'sms', 'postal', 'portal']),
  deliveredAt: z.string().datetime().optional(),
  openedAt: z.string().datetime().optional(),
  metadata: z.record(z.any()).optional()
});

// Invoice Approval Workflow
export const approveInvoiceSchema = z.object({
  invoiceId: objectIdSchema,
  approvedBy: objectIdSchema,
  notes: z.string().max(1000).optional()
});

export const rejectInvoiceSchema = z.object({
  invoiceId: objectIdSchema,
  rejectedBy: objectIdSchema,
  rejectionReason: z.string().min(5, 'Rejection reason must be at least 5 characters').max(500)
});

// Generate Invoice File
export const generateInvoiceFileSchema = z.object({
  invoiceId: objectIdSchema,
  type: z.enum(['pdf', 'xml', 'csv']).default('pdf'),
  template: z.string().optional(),
  includePaymentInfo: z.boolean().default(true)
});

// Invoice Filters - comprehensive
export const invoiceFiltersSchema = z.object({
  status: z.enum(['draft', 'open', 'paid', 'void', 'uncollectible', 'overdue']).optional(),
  paymentStatus: z.enum(['not_paid', 'partially_paid', 'paid', 'refunded', 'failed']).optional(),
  type: z.enum(['subscription', 'one_time', 'credit_note', 'proforma', 'recurring']).optional(),
  provider: z.enum(['stripe', 'paypal', 'paddle', 'lemonsqueezy', 'manual']).optional(),
  currency: currencySchema.optional(),
  userId: objectIdSchema.optional(),
  teamId: objectIdSchema.optional(),
  subscriptionId: objectIdSchema.optional(),
  hasOverdueAmount: z.boolean().optional(),
  requiresApproval: z.boolean().optional(),
  isApproved: z.boolean().optional(),
  dateRange: z.object({
    start: z.string().datetime().optional(),
    end: z.string().datetime().optional()
  }).optional(),
  dueDateRange: z.object({
    start: z.string().datetime().optional(),
    end: z.string().datetime().optional()
  }).optional(),
  amountRange: z.object({
    min: z.number().min(0).optional(),
    max: z.number().min(0).optional()
  }).optional(),
  tags: z.array(z.string()).optional(),
  page: z.number().min(1).default(1),
  limit: z.number().min(1).max(100).default(20),
  sort: z.enum(['invoiceNumber', 'totalAmount', 'invoiceDate', 'dueDate', 'status', 'createdAt']).default('invoiceDate'),
  order: z.enum(['asc', 'desc']).default('desc')
});

// Create Coupon - comprehensive
export const createCouponSchema = z.object({
  code: z.string().min(1, 'Coupon code is required').max(50).toUpperCase().regex(/^[A-Z0-9_-]+$/, 'Coupon code can only contain letters, numbers, underscores, and hyphens'),
  name: z.string().min(1, 'Coupon name is required').max(100),
  description: z.string().max(500).optional(),
  type: z.enum(['percentage', 'fixed_amount']),
  value: z.number().min(0, 'Value cannot be negative'),
  currency: currencySchema.optional(),
  maxRedemptions: z.number().min(1).optional(),
  currentRedemptions: z.number().min(0).default(0),
  maxRedemptionsPerUser: z.number().min(1).optional(),
  isActive: z.boolean().default(true),
  validFrom: z.string().datetime(),
  validUntil: z.string().datetime().optional(),
  applicablePlans: z.array(objectIdSchema).optional(),
  minimumAmount: z.number().min(0).optional(),
  firstTimeCustomersOnly: z.boolean().default(false),
  createdBy: objectIdSchema
}).refine((data) => {
  if (data.type === 'fixed_amount' && !data.currency) {
    return false;
  }
  return true;
}, {
  message: 'Currency is required for fixed amount coupons',
  path: ['currency']
}).refine((data) => {
  if (data.type === 'percentage' && (data.value < 0 || data.value > 100)) {
    return false;
  }
  return true;
}, {
  message: 'Percentage value must be between 0 and 100',
  path: ['value']
});

// Apply Coupon
export const applyCouponSchema = z.object({
  code: z.string().min(1, 'Coupon code is required').toUpperCase(),
  subscriptionId: objectIdSchema.optional(),
  invoiceId: objectIdSchema.optional(),
  userId: objectIdSchema,
  planIds: z.array(objectIdSchema).optional()
});

export type CreateInvoiceRequest = z.infer<typeof createInvoiceSchema>;
export type UpdateInvoiceRequest = z.infer<typeof updateInvoiceSchema>;
export type RecordCollectionAttemptRequest = z.infer<typeof recordCollectionAttemptSchema>;
export type RecordCommunicationRequest = z.infer<typeof recordCommunicationSchema>;
export type ApproveInvoiceRequest = z.infer<typeof approveInvoiceSchema>;
export type RejectInvoiceRequest = z.infer<typeof rejectInvoiceSchema>;
export type GenerateInvoiceFileRequest = z.infer<typeof generateInvoiceFileSchema>;
export type InvoiceFiltersRequest = z.infer<typeof invoiceFiltersSchema>;
export type CreateCouponRequest = z.infer<typeof createCouponSchema>;
export type ApplyCouponRequest = z.infer<typeof applyCouponSchema>;