import { BaseDocument, Currency, ObjectId } from ".";

export interface SubscriptionFeatures {
  storageLimit: number;
  memberLimit: number;
  fileUploadLimit: number;
  apiRequestLimit: number;
  enableAdvancedFeatures: boolean;
}

export interface SubscriptionUsage {
  storageUsed: number;
  apiRequestsUsed: number;
  lastResetAt: Date;
}

export interface Subscription extends BaseDocument {
  subscriber: ObjectId;
  subscriberType: 'user' | 'team';
  plan: ObjectId;
  provider: 'stripe' | 'paypal' | 'paddle' | 'lemonsqueezy';
  providerId: string;
  customerId: string;
  status: 'trial' | 'active' | 'past_due' | 'cancelled' | 'expired';
  currentPeriodStart: Date;
  currentPeriodEnd: Date;
  trialStart?: Date;
  trialEnd?: Date;
  cancelledAt?: Date;
  endedAt?: Date;
  currency: Currency;
  amount: number;
  interval: 'month' | 'year';
  intervalCount: number;
  features: SubscriptionFeatures;
  usage: SubscriptionUsage;
  nextBillingDate?: Date;
  lastPaymentDate?: Date;
  lastPaymentAmount?: number;
}

export interface PlanFeatures {
  storageLimit: number;
  memberLimit: number;
  fileUploadLimit: number;
  apiRequestLimit: number;
  enableAdvancedSharing: boolean;
  enableVersionHistory: boolean;
  enableOCR: boolean;
  enablePrioritySupport: boolean;
  enableAPIAccess: boolean;
  enableIntegrations: boolean;
  enableCustomBranding: boolean;
  enableAuditLogs: boolean;
  enableSSO: boolean;
}

export interface PlanPrice {
  amount: number;
  currency: Currency;
  providerId?: {
    stripe?: string;
    paypal?: string;
    paddle?: string;
    lemonsqueezy?: string;
  };
}

export interface Plan extends BaseDocument {
  name: string;
  description: string;
  isActive: boolean;
  isPublic: boolean;
  sortOrder: number;
  prices: {
    monthly: PlanPrice;
    yearly: PlanPrice;
  };
  features: PlanFeatures;
  trialDays: number;
}

export interface CreateSubscriptionRequest {
  planId: ObjectId;
  interval: 'month' | 'year';
  paymentMethodId?: string;
  couponCode?: string;
}

export interface UpgradeSubscriptionRequest {
  newPlanId: ObjectId;
  interval?: 'month' | 'year';
  prorationBehavior?: 'create_prorations' | 'none';
}