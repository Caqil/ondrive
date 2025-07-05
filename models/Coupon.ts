import mongoose, { Document, Schema } from 'mongoose';

export interface ICoupon extends Document {
  _id: string;
  code: string;
  name: string;
  description?: string;
  
  // Discount details
  type: 'percentage' | 'fixed_amount';
  value: number; // percentage (0-100) or amount in cents
  currency?: string; // required for fixed_amount
  
  // Usage limits
  maxRedemptions?: number;
  currentRedemptions: number;
  maxRedemptionsPerUser?: number;
  
  // Validity
  isActive: boolean;
  validFrom: Date;
  validUntil?: Date;
  
  // Applicable plans
  applicablePlans: mongoose.Types.ObjectId[]; // empty array means all plans
  
  // Restrictions
  minimumAmount?: number; // minimum order amount in cents
  firstTimeCustomersOnly: boolean;
  
  // Metadata
  createdBy: mongoose.Types.ObjectId;
  
  createdAt: Date;
  updatedAt: Date;
}

const couponSchema = new Schema<ICoupon>({
  code: {
    type: String,
    required: true,
    unique: true,
    uppercase: true,
    trim: true,
    match: /^[A-Z0-9_-]+$/,
    index: true
  },
  name: {
    type: String,
    required: true,
    trim: true,
    maxlength: 100
  },
  description: {
    type: String,
    trim: true,
    maxlength: 500
  },
  
  type: {
    type: String,
    enum: ['percentage', 'fixed_amount'],
    required: true
  },
  value: {
    type: Number,
    required: true,
    min: 0
  },
  currency: {
    type: String,
    uppercase: true,
    length: 3
  },
  
  maxRedemptions: {
    type: Number,
    min: 1
  },
  currentRedemptions: {
    type: Number,
    default: 0,
    min: 0
  },
  maxRedemptionsPerUser: {
    type: Number,
    min: 1
  },
  
  isActive: {
    type: Boolean,
    default: true,
    index: true
  },
  validFrom: {
    type: Date,
    required: true,
    index: true
  },
  validUntil: {
    type: Date,
    index: true
  },
  
  applicablePlans: [{
    type: Schema.Types.ObjectId,
    ref: 'Plan'
  }],
  
  minimumAmount: {
    type: Number,
    min: 0
  },
  firstTimeCustomersOnly: {
    type: Boolean,
    default: false
  },
  
  createdBy: {
    type: Schema.Types.ObjectId,
    ref: 'User',
    required: true
  }
}, {
  timestamps: true,
  collection: 'coupons'
});

// Indexes
couponSchema.index({ isActive: 1, validFrom: 1, validUntil: 1 });
couponSchema.index({ code: 1, isActive: 1 });

// Validation
couponSchema.pre('save', function(next) {
  if (this.type === 'percentage' && (this.value < 0 || this.value > 100)) {
    return next(new Error('Percentage value must be between 0 and 100'));
  }
  if (this.type === 'fixed_amount' && !this.currency) {
    return next(new Error('Currency is required for fixed amount coupons'));
  }
  if (this.validUntil && this.validUntil <= this.validFrom) {
    return next(new Error('Valid until date must be after valid from date'));
  }
  next();
});

export const Coupon = mongoose.models.Coupon || mongoose.model<ICoupon>('Coupon', couponSchema);
