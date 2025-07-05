import { BaseDocument, CountryCode, Currency, ObjectId } from ".";

export interface PaymentTaxDetails {
  taxRate: number;
  taxType: string;
  taxId?: string;
  taxRegion?: string;
}

export interface PaymentMethod {
  type: 'card' | 'bank_transfer' | 'paypal' | 'crypto' | 'wallet' | 'other';
  brand?: string;
  last4?: string;
  expiryMonth?: number;
  expiryYear?: number;
  holderName?: string;
  fingerprint?: string;
}

export interface PaymentRefund {
  refundId: string;
  amount: number;
  reason: string;
  status: 'pending' | 'succeeded' | 'failed';
  processedAt?: Date;
  refundedBy?: ObjectId;
}

export interface PaymentDispute {
  disputeId: string;
  reason: string;
  status: 'warning_needs_response' | 'warning_under_review' | 'warning_closed' | 'needs_response' | 'under_review' | 'charge_refunded' | 'won' | 'lost';
  amount: number;
  evidence?: any;
  dueBy?: Date;
}

export interface PaymentBillingAddress {
  line1: string;
  line2?: string;
  city: string;
  state?: string;
  postalCode: string;
  country: CountryCode;
}

export interface PaymentCustomerInfo {
  email: string;
  name: string;
  phone?: string;
  ipAddress?: string;
  userAgent?: string;
}

export interface AppliedCoupon {
  couponId: ObjectId;
  code: string;
  discountAmount: number;
  discountType: 'percentage' | 'fixed_amount';
}

export interface Payment extends BaseDocument {
  paymentNumber: string;
  subscription?: ObjectId;
  invoice?: ObjectId;
  user: ObjectId;
  team?: ObjectId;
  type: 'subscription' | 'one_time' | 'refund' | 'adjustment' | 'credit';
  description: string;
  amount: number;
  currency: Currency;
  subtotal: number;
  taxAmount: number;
  feeAmount: number;
  discountAmount: number;
  netAmount: number;
  taxDetails: PaymentTaxDetails;
  paymentMethod: PaymentMethod;
  provider: 'stripe' | 'paypal' | 'paddle' | 'lemonsqueezy' | 'razorpay' | 'manual';
  providerId: string;
  providerCustomerId?: string;
  providerFee?: number;
  status: 'pending' | 'processing' | 'succeeded' | 'failed' | 'cancelled' | 'refunded' | 'partially_refunded' | 'disputed';
  failureReason?: string;
  failureCode?: string;
  processedAt?: Date;
  capturedAt?: Date;
  settledAt?: Date;
  failedAt?: Date;
  refundedAt?: Date;
  disputedAt?: Date;
  refunds: PaymentRefund[];
  totalRefunded: number;
  dispute?: PaymentDispute;
  riskScore?: number;
  riskLevel?: 'low' | 'medium' | 'high';
  billingAddress: PaymentBillingAddress;
  customerInfo: PaymentCustomerInfo;
  appliedCoupons: AppliedCoupon[];
  metadata: Record<string, any>;
  reconciled: boolean;
  reconciledAt?: Date;
  reconciledBy?: ObjectId;
}

export interface CreatePaymentRequest {
  amount: number;
  currency: Currency;
  paymentMethodId: string;
  description: string;
  subscriptionId?: ObjectId;
  metadata?: Record<string, any>;
}