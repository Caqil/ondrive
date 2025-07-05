import mongoose, { Document, Schema } from 'mongoose';

export interface IPayment extends Document {
  _id: string;
  
  // Payment identification
  paymentNumber: string; // Internal payment number
  
  // Related entities
  subscription?: mongoose.Types.ObjectId;
  invoice?: mongoose.Types.ObjectId;
  user: mongoose.Types.ObjectId;
  team?: mongoose.Types.ObjectId;
  
  // Payment details
  type: 'subscription' | 'one_time' | 'refund' | 'adjustment' | 'credit';
  description: string;
  
  // Amount details
  amount: number; // Total amount in cents
  currency: string;
  subtotal: number; // Amount before taxes/fees
  taxAmount: number;
  feeAmount: number; // Processing fees
  discountAmount: number; // Coupon/discount applied
  netAmount: number; // Amount after fees
  
  // Tax information
  taxDetails: {
    taxRate: number; // Percentage
    taxType: string; // VAT, GST, Sales Tax, etc.
    taxId?: string; // Tax registration number
    taxRegion?: string; // Country/state where tax applies
  };
  
  // Payment method
  paymentMethod: {
    type: 'card' | 'bank_transfer' | 'paypal' | 'crypto' | 'wallet' | 'other';
    brand?: string; // visa, mastercard, etc.
    last4?: string; // Last 4 digits
    expiryMonth?: number;
    expiryYear?: number;
    holderName?: string;
    fingerprint?: string; // Unique identifier for the payment method
  };
  
  // Provider details
  provider: 'stripe' | 'paypal' | 'paddle' | 'lemonsqueezy' | 'razorpay' | 'manual';
  providerId: string; // External payment ID
  providerCustomerId?: string; // External customer ID
  providerFee?: number; // Fee charged by payment provider
  
  // Status and timing
  status: 'pending' | 'processing' | 'succeeded' | 'failed' | 'cancelled' | 'refunded' | 'partially_refunded' | 'disputed';
  failureReason?: string;
  failureCode?: string;
  
  // Important dates
  processedAt?: Date;
  capturedAt?: Date;
  settledAt?: Date; // When funds were settled to account
  failedAt?: Date;
  refundedAt?: Date;
  disputedAt?: Date;
  
  // Refund information
  refunds: {
    refundId: string;
    amount: number;
    reason: string;
    status: 'pending' | 'succeeded' | 'failed';
    processedAt?: Date;
    refundedBy?: mongoose.Types.ObjectId;
  }[];
  totalRefunded: number;
  
  // Dispute information
  dispute?: {
    disputeId: string;
    reason: string;
    status: 'warning_needs_response' | 'warning_under_review' | 'warning_closed' | 'needs_response' | 'under_review' | 'charge_refunded' | 'won' | 'lost';
    amount: number;
    evidence?: any;
    dueBy?: Date;
  };
  
  // Risk assessment
  riskScore?: number; // 0-100, higher is riskier
  riskLevel?: 'low' | 'medium' | 'high';
  fraudCheck?: {
    passed: boolean;
    score: number;
    checks: {
      cvv: boolean;
      address: boolean;
      zip: boolean;
    };
  };
  
  // Billing address
  billingAddress: {
    line1: string;
    line2?: string;
    city: string;
    state?: string;
    postalCode: string;
    country: string;
  };
  
  // Customer information
  customerInfo: {
    email: string;
    name: string;
    phone?: string;
    ipAddress?: string;
    userAgent?: string;
  };
  
  // Applied discounts/coupons
  appliedCoupons: {
    couponId: mongoose.Types.ObjectId;
    code: string;
    discountAmount: number;
    discountType: 'percentage' | 'fixed_amount';
  }[];
  
  // Webhook information
  webhookEvents: {
    eventType: string;
    eventId: string;
    receivedAt: Date;
    processed: boolean;
  }[];
  
  // Metadata for additional information
  metadata: {
    planName?: string;
    billingPeriod?: string;
    prorationAmount?: number;
    upgradeDowngrade?: 'upgrade' | 'downgrade';
    isRenewal?: boolean;
    campaignId?: string;
    affiliateId?: string;
    [key: string]: any;
  };
  
  // Internal tracking
  reconciled: boolean; // Whether payment has been reconciled with bank statement
  reconciledAt?: Date;
  reconciledBy?: mongoose.Types.ObjectId;
  
  // Notifications
  notificationsSent: {
    customer: boolean;
    admin: boolean;
    accounting: boolean;
  };
  
  createdAt: Date;
  updatedAt: Date;
}

const paymentSchema = new Schema<IPayment>({
  paymentNumber: {
    type: String,
    required: true,
    unique: true,
    index: true
  },
  
  subscription: {
    type: Schema.Types.ObjectId,
    ref: 'Subscription',
    index: true
  },
  invoice: {
    type: Schema.Types.ObjectId,
    ref: 'Invoice',
    index: true
  },
  user: {
    type: Schema.Types.ObjectId,
    ref: 'User',
    required: true,
    index: true
  },
  team: {
    type: Schema.Types.ObjectId,
    ref: 'Team',
    index: true
  },
  
  type: {
    type: String,
    enum: ['subscription', 'one_time', 'refund', 'adjustment', 'credit'],
    required: true,
    index: true
  },
  description: {
    type: String,
    required: true,
    trim: true,
    maxlength: 500
  },
  
  amount: {
    type: Number,
    required: true,
    min: 0
  },
  currency: {
    type: String,
    required: true,
    uppercase: true,
    length: 3
  },
  subtotal: {
    type: Number,
    required: true,
    min: 0
  },
  taxAmount: {
    type: Number,
    default: 0,
    min: 0
  },
  feeAmount: {
    type: Number,
    default: 0,
    min: 0
  },
  discountAmount: {
    type: Number,
    default: 0,
    min: 0
  },
  netAmount: {
    type: Number,
    required: true,
    min: 0
  },
  
  taxDetails: {
    taxRate: {
      type: Number,
      default: 0,
      min: 0,
      max: 100
    },
    taxType: {
      type: String,
      default: 'VAT'
    },
    taxId: String,
    taxRegion: String
  },
  
  paymentMethod: {
    type: {
      type: String,
      enum: ['card', 'bank_transfer', 'paypal', 'crypto', 'wallet', 'other'],
      required: true
    },
    brand: String,
    last4: {
      type: String,
      length: 4
    },
    expiryMonth: {
      type: Number,
      min: 1,
      max: 12
    },
    expiryYear: {
      type: Number,
      min: new Date().getFullYear()
    },
    holderName: String,
    fingerprint: String
  },
  
  provider: {
    type: String,
    enum: ['stripe', 'paypal', 'paddle', 'lemonsqueezy', 'razorpay', 'manual'],
    required: true,
    index: true
  },
  providerId: {
    type: String,
    required: true,
    index: true
  },
  providerCustomerId: {
    type: String,
    index: true
  },
  providerFee: {
    type: Number,
    min: 0
  },
  
  status: {
    type: String,
    enum: ['pending', 'processing', 'succeeded', 'failed', 'cancelled', 'refunded', 'partially_refunded', 'disputed'],
    required: true,
    index: true
  },
  failureReason: String,
  failureCode: String,
  
  processedAt: Date,
  capturedAt: Date,
  settledAt: Date,
  failedAt: Date,
  refundedAt: Date,
  disputedAt: Date,
  
  refunds: [{
    refundId: {
      type: String,
      required: true
    },
    amount: {
      type: Number,
      required: true,
      min: 0
    },
    reason: {
      type: String,
      required: true,
      maxlength: 500
    },
    status: {
      type: String,
      enum: ['pending', 'succeeded', 'failed'],
      required: true
    },
    processedAt: Date,
    refundedBy: {
      type: Schema.Types.ObjectId,
      ref: 'User'
    }
  }],
  totalRefunded: {
    type: Number,
    default: 0,
    min: 0
  },
  
  dispute: {
    disputeId: String,
    reason: String,
    status: {
      type: String,
      enum: ['warning_needs_response', 'warning_under_review', 'warning_closed', 'needs_response', 'under_review', 'charge_refunded', 'won', 'lost']
    },
    amount: Number,
    evidence: Schema.Types.Mixed,
    dueBy: Date
  },
  
  riskScore: {
    type: Number,
    min: 0,
    max: 100
  },
  riskLevel: {
    type: String,
    enum: ['low', 'medium', 'high']
  },
  fraudCheck: {
    passed: Boolean,
    score: Number,
    checks: {
      cvv: Boolean,
      address: Boolean,
      zip: Boolean
    }
  },
  
  billingAddress: {
    line1: {
      type: String,
      required: true,
      trim: true
    },
    line2: {
      type: String,
      trim: true
    },
    city: {
      type: String,
      required: true,
      trim: true
    },
    state: {
      type: String,
      trim: true
    },
    postalCode: {
      type: String,
      required: true,
      trim: true
    },
    country: {
      type: String,
      required: true,
      uppercase: true,
      length: 2 // ISO country code
    }
  },
  
  customerInfo: {
    email: {
      type: String,
      required: true,
      lowercase: true,
      trim: true
    },
    name: {
      type: String,
      required: true,
      trim: true
    },
    phone: String,
    ipAddress: String,
    userAgent: String
  },
  
  appliedCoupons: [{
    couponId: {
      type: Schema.Types.ObjectId,
      ref: 'Coupon',
      required: true
    },
    code: {
      type: String,
      required: true,
      uppercase: true
    },
    discountAmount: {
      type: Number,
      required: true,
      min: 0
    },
    discountType: {
      type: String,
      enum: ['percentage', 'fixed_amount'],
      required: true
    }
  }],
  
  webhookEvents: [{
    eventType: {
      type: String,
      required: true
    },
    eventId: {
      type: String,
      required: true
    },
    receivedAt: {
      type: Date,
      required: true,
      default: Date.now
    },
    processed: {
      type: Boolean,
      default: false
    }
  }],
  
  metadata: {
    type: Schema.Types.Mixed,
    default: {}
  },
  
  reconciled: {
    type: Boolean,
    default: false,
    index: true
  },
  reconciledAt: Date,
  reconciledBy: {
    type: Schema.Types.ObjectId,
    ref: 'User'
  },
  
  notificationsSent: {
    customer: {
      type: Boolean,
      default: false
    },
    admin: {
      type: Boolean,
      default: false
    },
    accounting: {
      type: Boolean,
      default: false
    }
  }
}, {
  timestamps: true,
  collection: 'payments'
});

// Compound indexes for efficient queries
paymentSchema.index({ user: 1, status: 1, createdAt: -1 });
paymentSchema.index({ subscription: 1, status: 1, createdAt: -1 });
paymentSchema.index({ provider: 1, providerId: 1 });
paymentSchema.index({ status: 1, processedAt: -1 });
paymentSchema.index({ currency: 1, amount: 1, createdAt: -1 });
paymentSchema.index({ 'customerInfo.email': 1, createdAt: -1 });
paymentSchema.index({ reconciled: 1, settledAt: -1 });
paymentSchema.index({ type: 1, status: 1, createdAt: -1 });

// Text search index for payment descriptions
paymentSchema.index({
  paymentNumber: 'text',
  description: 'text',
  'customerInfo.name': 'text',
  'customerInfo.email': 'text'
});

// Pre-save middleware for auto-generating payment number
paymentSchema.pre('save', async function(next) {
  if (this.isNew && !this.paymentNumber) {
    const year = new Date().getFullYear();
    const month = (new Date().getMonth() + 1).toString().padStart(2, '0');
    const count = await mongoose.models.Payment.countDocuments({
      createdAt: {
        $gte: new Date(year, new Date().getMonth(), 1),
        $lt: new Date(year, new Date().getMonth() + 1, 1)
      }
    }) + 1;
    
    this.paymentNumber = `PAY-${year}${month}-${count.toString().padStart(6, '0')}`;
  }
  
  // Calculate net amount
  this.netAmount = this.amount - this.feeAmount;
  
  next();
});

export const Payment = mongoose.models.Payment || mongoose.model<IPayment>('Payment', paymentSchema);
