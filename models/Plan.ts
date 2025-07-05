import mongoose, { Document, Schema } from 'mongoose';

export interface IPlan extends Document {
  _id: string;
  name: string;
  description: string;
  
  // Plan details
  isActive: boolean;
  isPublic: boolean; // Whether shown on pricing page
  sortOrder: number;
  
  // Pricing
  prices: {
    monthly: {
      amount: number; // in cents
      currency: string;
      providerId?: {
        stripe?: string;
        paypal?: string;
        paddle?: string;
        lemonsqueezy?: string;
      };
    };
    yearly: {
      amount: number; // in cents
      currency: string;
      providerId?: {
        stripe?: string;
        paypal?: string;
        paddle?: string;
        lemonsqueezy?: string;
      };
    };
  };
  
  // Features and limits
  features: {
    storageLimit: number; // bytes
    memberLimit: number;
    fileUploadLimit: number; // bytes per file
    apiRequestLimit: number; // per month
    enableAdvancedSharing: boolean;
    enableVersionHistory: boolean;
    enableOCR: boolean;
    enablePrioritySupport: boolean;
    enableAPIAccess: boolean;
    enableIntegrations: boolean;
    enableCustomBranding: boolean;
    enableAuditLogs: boolean;
    enableSSO: boolean;
  };
  
  // Trial
  trialDays: number;
  
  createdAt: Date;
  updatedAt: Date;
}

const planSchema = new Schema<IPlan>({
  name: {
    type: String,
    required: true,
    trim: true,
    maxlength: 100,
    index: true
  },
  description: {
    type: String,
    required: true,
    trim: true,
    maxlength: 500
  },
  
  isActive: {
    type: Boolean,
    default: true,
    index: true
  },
  isPublic: {
    type: Boolean,
    default: true,
    index: true
  },
  sortOrder: {
    type: Number,
    default: 0,
    index: true
  },
  
  prices: {
    monthly: {
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
      providerId: {
        stripe: String,
        paypal: String,
        paddle: String,
        lemonsqueezy: String
      }
    },
    yearly: {
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
      providerId: {
        stripe: String,
        paypal: String,
        paddle: String,
        lemonsqueezy: String
      }
    }
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
    enableAdvancedSharing: {
      type: Boolean,
      default: false
    },
    enableVersionHistory: {
      type: Boolean,
      default: true
    },
    enableOCR: {
      type: Boolean,
      default: false
    },
    enablePrioritySupport: {
      type: Boolean,
      default: false
    },
    enableAPIAccess: {
      type: Boolean,
      default: false
    },
    enableIntegrations: {
      type: Boolean,
      default: false
    },
    enableCustomBranding: {
      type: Boolean,
      default: false
    },
    enableAuditLogs: {
      type: Boolean,
      default: false
    },
    enableSSO: {
      type: Boolean,
      default: false
    }
  },
  
  trialDays: {
    type: Number,
    default: 14,
    min: 0,
    max: 365
  }
}, {
  timestamps: true,
  collection: 'plans'
});

// Indexes
planSchema.index({ isActive: 1, isPublic: 1, sortOrder: 1 });

export const Plan = mongoose.models.Plan || mongoose.model<IPlan>('Plan', planSchema);
