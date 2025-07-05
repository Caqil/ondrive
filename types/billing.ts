import { AppliedCoupon, BaseDocument, CountryCode, Currency, ObjectId } from ".";

export interface InvoiceLineItem {
  id: string;
  description: string;
  quantity: number;
  unitPrice: number;
  amount: number;
  taxRate: number;
  taxAmount: number;
  periodStart?: Date;
  periodEnd?: Date;
  planId?: ObjectId;
  metadata?: any;
}

export interface InvoiceTaxDetails {
  taxId?: string;
  taxRegion: string;
  taxType: string;
  reverseCharge: boolean;
  taxExempt: boolean;
  taxExemptReason?: string;
}

export interface InvoiceBillingAddress {
  name: string;
  company?: string;
  line1: string;
  line2?: string;
  city: string;
  state?: string;
  postalCode: string;
  country: CountryCode;
  taxId?: string;
}

export interface Invoice extends BaseDocument {
  invoiceNumber: string;
  sequence: number;
  subscription?: ObjectId;
  user: ObjectId;
  team?: ObjectId;
  type: 'subscription' | 'one_time' | 'credit_note' | 'proforma' | 'recurring';
  description: string;
  lineItems: InvoiceLineItem[];
  subtotal: number;
  taxAmount: number;
  discountAmount: number;
  totalAmount: number;
  amountPaid: number;
  amountDue: number;
  currency: Currency;
  exchangeRate?: number;
  baseCurrency?: Currency;
  taxDetails: InvoiceTaxDetails;
  billingAddress: InvoiceBillingAddress;
  invoiceDate: Date;
  dueDate: Date;
  periodStart?: Date;
  periodEnd?: Date;
  paidAt?: Date;
  voidedAt?: Date;
  sentAt?: Date;
  status: 'draft' | 'open' | 'paid' | 'void' | 'uncollectible' | 'overdue';
  paymentStatus: 'not_paid' | 'partially_paid' | 'paid' | 'refunded' | 'failed';
  payments: ObjectId[];
  provider: 'stripe' | 'paypal' | 'paddle' | 'lemonsqueezy' | 'manual';
  providerId?: string;
  providerUrl?: string;
  appliedCoupons: AppliedCoupon[];
  metadata: Record<string, any>;
}

export interface CreateInvoiceRequest {
  userId: ObjectId;
  subscriptionId?: ObjectId;
  lineItems: Omit<InvoiceLineItem, 'id'>[];
  dueDate: Date;
  description: string;
}

export interface Coupon extends BaseDocument {
  code: string;
  name: string;
  description?: string;
  type: 'percentage' | 'fixed_amount';
  value: number;
  currency?: Currency;
  maxRedemptions?: number;
  currentRedemptions: number;
  maxRedemptionsPerUser?: number;
  isActive: boolean;
  validFrom: Date;
  validUntil?: Date;
  applicablePlans: ObjectId[];
  minimumAmount?: number;
  firstTimeCustomersOnly: boolean;
  createdBy: ObjectId;
}