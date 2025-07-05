// lib/validations/billing.ts
import { z } from 'zod';
import { objectIdSchema, currencySchema, addressSchema, emailSchema, nameSchema, metadataSchema } from './base';

// Billing Address - matches models/Invoice.ts
const billingAddressSchema = z.object({
  name: nameSchema,
  company: z.string().max(100, 'Company name cannot exceed 100 characters').optional(),
  line1: z.string().min(1, 'Address line 1 is required').max(100),
  line2: z.string().max(100).optional(),
  city: z.string().min(1, 'City is required').max(50),
  state: z.string().max(50).optional(),
  postalCode: z.string().min(1, 'Postal code is required').max(20),
  country: z.string().length(2, 'Country code must be exactly 2 characters').toUpperCase(),
  taxId: z.string().max(50, 'Tax ID cannot exceed 50 characters').optional()
});

// Invoice Tax Details - matches models/Invoice.ts
const invoiceTaxDetailsSchema = z.object({
  taxId: z.string().max(50, 'Tax ID cannot exceed 50 characters').optional(),
  taxRegion: z.string().min(1, 'Tax region is required').max(100),
  taxType: z.string().min(1, 'Tax type is required').max(50).default('VAT'),
  reverseCharge: z.boolean().default(false),
  taxExempt: z.boolean().default(false),
  taxExemptReason: z.string().max(200, 'Tax exempt reason cannot exceed 200 characters').optional()
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
  metadata: metadataSchema
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
  provider: z.enum(['stripe', 'paypal', 'paddle', 'lemonsqueezy', 'manual']).default('manual'),
  appliedCoupons: z.array(z.object({
    couponId: objectIdSchema,
    code: z.string().min(1, 'Coupon code is required').max(50).toUpperCase(),
    discountAmount: z.number().min(0),
    discountType: z.enum(['percentage', 'fixed_amount'])
  })).optional(),
  creditNote: creditNoteSchema.optional(),
  metadata: metadataSchema
});

// Update Invoice - for draft invoices
export const updateInvoiceSchema = z.object({
  description: z.string().min(1, 'Description is required').max(1000).optional(),
  lineItems: z.array(invoiceLineItemSchema).min(1, 'At least one line item is required').optional(),
  taxDetails: invoiceTaxDetailsSchema.optional(),
  billingAddress: billingAddressSchema.optional(),
  dueDate: z.string().datetime().optional(),
  periodStart: z.string().datetime().optional(),
  periodEnd: z.string().datetime().optional(),
  appliedCoupons: z.array(z.object({
    couponId: objectIdSchema,
    code: z.string().min(1, 'Coupon code is required').max(50).toUpperCase(),
    discountAmount: z.number().min(0),
    discountType: z.enum(['percentage', 'fixed_amount'])
  })).optional(),
  metadata: metadataSchema
});

// Invoice Filters - for listing invoices
export const invoiceFiltersSchema = z.object({
  userId: objectIdSchema.optional(),
  teamId: objectIdSchema.optional(),
  subscriptionId: objectIdSchema.optional(),
  status: z.enum(['draft', 'open', 'paid', 'void', 'uncollectible', 'overdue']).optional(),
  paymentStatus: z.enum(['not_paid', 'partially_paid', 'paid', 'refunded', 'failed']).optional(),
  type: z.enum(['subscription', 'one_time', 'credit_note', 'proforma', 'recurring']).optional(),
  currency: currencySchema.optional(),
  provider: z.enum(['stripe', 'paypal', 'paddle', 'lemonsqueezy', 'manual']).optional(),
  amountRange: z.object({
    min: z.number().min(0).optional(),
    max: z.number().min(0).optional()
  }).optional(),
  dateRange: z.object({
    start: z.string().datetime().optional(),
    end: z.string().datetime().optional()
  }).optional(),
  dueDateRange: z.object({
    start: z.string().datetime().optional(),
    end: z.string().datetime().optional()
  }).optional(),
  query: z.string().max(200, 'Search query cannot exceed 200 characters').optional(),
  page: z.number().min(1).default(1),
  limit: z.number().min(1).max(100).default(20),
  sort: z.enum(['invoiceNumber', 'invoiceDate', 'dueDate', 'totalAmount', 'status', 'createdAt']).default('createdAt'),
  order: z.enum(['asc', 'desc']).default('desc')
});

// Send Invoice - for manual invoice sending
export const sendInvoiceSchema = z.object({
  sendToCustomer: z.boolean().default(true),
  sendToAdmin: z.boolean().default(false),
  customMessage: z.string().max(1000, 'Custom message cannot exceed 1000 characters').optional(),
  recipientEmails: z.array(emailSchema).max(10, 'Cannot send to more than 10 recipients').optional()
});

// Void Invoice - for cancelling invoices
export const voidInvoiceSchema = z.object({
  reason: z.string().min(3, 'Void reason must be at least 3 characters').max(500, 'Void reason cannot exceed 500 characters')
});

// Record Payment - for manual payment recording
export const recordPaymentSchema = z.object({
  amount: z.number().min(1, 'Payment amount must be greater than 0'),
  paymentMethod: z.enum(['cash', 'check', 'bank_transfer', 'other']).default('other'),
  reference: z.string().max(100, 'Payment reference cannot exceed 100 characters').optional(),
  notes: z.string().max(500, 'Payment notes cannot exceed 500 characters').optional(),
  paymentDate: z.string().datetime().default(new Date().toISOString())
});

// Collection Attempt - for tracking collection efforts
export const collectionAttemptSchema = z.object({
  method: z.enum(['auto', 'manual']),
  result: z.enum(['succeeded', 'failed', 'pending']),
  failureReason: z.string().max(200, 'Failure reason cannot exceed 200 characters').optional(),
  nextAttempt: z.string().datetime().optional(),
  notes: z.string().max(500, 'Collection notes cannot exceed 500 characters').optional()
});

// Export all schemas for use in other files
export {
  billingAddressSchema,
  invoiceTaxDetailsSchema,
  legalEntitySchema,
  creditNoteSchema
};

// Type exports
export type CreateInvoiceRequest = z.infer<typeof createInvoiceSchema>;
export type UpdateInvoiceRequest = z.infer<typeof updateInvoiceSchema>;
export type InvoiceFiltersRequest = z.infer<typeof invoiceFiltersSchema>;
export type SendInvoiceRequest = z.infer<typeof sendInvoiceSchema>;
export type VoidInvoiceRequest = z.infer<typeof voidInvoiceSchema>;
export type RecordPaymentRequest = z.infer<typeof recordPaymentSchema>;
export type CollectionAttemptRequest = z.infer<typeof collectionAttemptSchema>;
export type InvoiceLineItemRequest = z.infer<typeof invoiceLineItemSchema>;