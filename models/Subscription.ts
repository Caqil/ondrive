import mongoose, { Document, Schema } from 'mongoose';

export interface ISubscription extends Document {
  _id: string;
  
  // Subscriber (User or Team)
  subscriber: mongoose.Types.ObjectId;
  subscriberType: 'user' | 'team';
  
  // Plan details
  plan: mongoose.Types.ObjectId;
  
  // Payment provider
  provider: 'stripe' | 'paypal' | 'paddle' | 'lemonsqueezy';
  providerId: string; // External subscription ID
  customerId: string; // External customer ID
  
  // Status and timing
  status: 'trial' | 'active' | 'past_due' | 'cancelled' | 'expired';
  currentPeriodStart: Date;
  currentPeriodEnd: Date;
  trialStart?: Date;
  trialEnd?: Date;
  cancelledAt?: Date;
  endedAt?: Date;
  
  // Pricing
  currency: string;
  amount: number; // in cents
  interval: 'month' | 'year';
  intervalCount: number;
  
  // Features and limits
  features: {
    storageLimit: number; // bytes
    memberLimit: number;
    fileUploadLimit: number; // bytes per file
    apiRequestLimit: number; // per month
    enableAdvancedFeatures: boolean;
  };
  
  // Usage tracking
  usage: {
    storageUsed: number;
    apiRequestsUsed: number;
    lastResetAt: Date;
  };
  
  // Billing
  nextBillingDate?: Date;
  lastPaymentDate?: Date;
  lastPaymentAmount?: number;
  
  createdAt: Date;
  updatedAt: Date;
}

const subscriptionSchema = new Schema<ISubscription>({
  subscriber: {
    type: Schema.Types.ObjectId,
    required: true,
    refPath: 'subscriberType',
    index: true
  },
  subscriberType: {
    type: String,
    required: true,
    enum: ['user', 'team']
  },
  
  plan: {
    type: Schema.Types.ObjectId,
    ref: 'Plan',
    required: true,
    index: true
  },
  
  provider: {
    type: String,
    enum: ['stripe', 'paypal', 'paddle', 'lemonsqueezy'],
    required: true,
    index: true
  },
  providerId: {
    type: String,
    required: true,
    index: true
  },
  customerId: {
    type: String,
    required: true,
    index: true
  },
  
  status: {
    type: String,
    enum: ['trial', 'active', 'past_due', 'cancelled', 'expired'],
    required: true,
    index: true
  },
  currentPeriodStart: {
    type: Date,
    required: true
  },
  currentPeriodEnd: {
    type: Date,
    required: true,
    index: true
  },
  trialStart: Date,
  trialEnd: Date,
  cancelledAt: Date,
  endedAt: Date,
  
  currency: {
    type: String,
    required: true,
    uppercase: true,
    length: 3
  },
  amount: {
    type: Number,
    required: true,
    min: 0
  },
  interval: {
    type: String,
    enum: ['month', 'year'],
    required: true
  },
  intervalCount: {
    type: Number,
    required: true,
    min: 1
  },
  
  features: {
    storageLimit: {
      type: Number,
      required: true,
      min: 0
    },
    memberLimit: {
      type: Number,
      required: true,
      min: 1
    },
    fileUploadLimit: {
      type: Number,
      required: true,
      min: 0
    },
    apiRequestLimit: {
      type: Number,
      required: true,
      min: 0
    },
    enableAdvancedFeatures: {
      type: Boolean,
      default: false
    }
  },
  
  usage: {
    storageUsed: {
      type: Number,
      default: 0,
      min: 0
    },
    apiRequestsUsed: {
      type: Number,
      default: 0,
      min: 0
    },
    lastResetAt: {
      type: Date,
      default: Date.now
    }
  },
  
  nextBillingDate: Date,
  lastPaymentDate: Date,
  lastPaymentAmount: Number
}, {
  timestamps: true,
  collection: 'subscriptions'
});

// Compound indexes
subscriptionSchema.index({ subscriber: 1, subscriberType: 1 });
subscriptionSchema.index({ provider: 1, providerId: 1 });
subscriptionSchema.index({ status: 1, currentPeriodEnd: 1 });

export const Subscription = mongoose.models.Subscription || mongoose.model<ISubscription>('Subscription', subscriptionSchema);
