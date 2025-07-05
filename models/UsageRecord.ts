import mongoose, { Schema } from "mongoose";

export interface IUsageRecord extends Document {
  _id: string;
  
  // Who/what this usage is for
  user?: mongoose.Types.ObjectId;
  team?: mongoose.Types.ObjectId;
  subscription?: mongoose.Types.ObjectId;
  
  // Usage metrics
  storageUsed: number; // bytes
  bandwidthUsed: number; // bytes
  apiRequestsUsed: number;
  fileUploadsCount: number;
  fileDownloadsCount: number;
  shareLinksCreated: number;
  
  // Time period
  period: 'daily' | 'monthly' | 'yearly';
  date: Date; // Start of the period
  
  // Breakdown by file types
  fileTypeBreakdown: {
    images: { count: number; size: number };
    videos: { count: number; size: number };
    documents: { count: number; size: number };
    archives: { count: number; size: number };
    others: { count: number; size: number };
  };
  
  // Costs (if applicable)
  costs: {
    storage: number; // in cents
    bandwidth: number; // in cents
    apiRequests: number; // in cents
    total: number; // in cents
  };
  
  createdAt: Date;
  updatedAt: Date;
}

const usageRecordSchema = new Schema<IUsageRecord>({
  user: {
    type: Schema.Types.ObjectId,
    ref: 'User',
    index: true
  },
  team: {
    type: Schema.Types.ObjectId,
    ref: 'Team',
    index: true
  },
  subscription: {
    type: Schema.Types.ObjectId,
    ref: 'Subscription',
    index: true
  },
  
  storageUsed: {
    type: Number,
    default: 0,
    min: 0
  },
  bandwidthUsed: {
    type: Number,
    default: 0,
    min: 0
  },
  apiRequestsUsed: {
    type: Number,
    default: 0,
    min: 0
  },
  fileUploadsCount: {
    type: Number,
    default: 0,
    min: 0
  },
  fileDownloadsCount: {
    type: Number,
    default: 0,
    min: 0
  },
  shareLinksCreated: {
    type: Number,
    default: 0,
    min: 0
  },
  
  period: {
    type: String,
    enum: ['daily', 'monthly', 'yearly'],
    required: true,
    index: true
  },
  date: {
    type: Date,
    required: true,
    index: true
  },
  
  fileTypeBreakdown: {
    images: {
      count: { type: Number, default: 0, min: 0 },
      size: { type: Number, default: 0, min: 0 }
    },
    videos: {
      count: { type: Number, default: 0, min: 0 },
      size: { type: Number, default: 0, min: 0 }
    },
    documents: {
      count: { type: Number, default: 0, min: 0 },
      size: { type: Number, default: 0, min: 0 }
    },
    archives: {
      count: { type: Number, default: 0, min: 0 },
      size: { type: Number, default: 0, min: 0 }
    },
    others: {
      count: { type: Number, default: 0, min: 0 },
      size: { type: Number, default: 0, min: 0 }
    }
  },
  
  costs: {
    storage: {
      type: Number,
      default: 0,
      min: 0
    },
    bandwidth: {
      type: Number,
      default: 0,
      min: 0
    },
    apiRequests: {
      type: Number,
      default: 0,
      min: 0
    },
    total: {
      type: Number,
      default: 0,
      min: 0
    }
  }
}, {
  timestamps: true,
  collection: 'usage_records'
});

// Compound indexes
usageRecordSchema.index({ user: 1, period: 1, date: -1 });
usageRecordSchema.index({ team: 1, period: 1, date: -1 });
usageRecordSchema.index({ subscription: 1, period: 1, date: -1 });
usageRecordSchema.index({ period: 1, date: -1 });

// Unique constraint to prevent duplicate records
usageRecordSchema.index({ user: 1, team: 1, period: 1, date: 1 }, { 
  unique: true,
  partialFilterExpression: { 
    $or: [
      { user: { $exists: true } },
      { team: { $exists: true } }
    ]
  }
});

export const UsageRecord = mongoose.models.UsageRecord || mongoose.model<IUsageRecord>('UsageRecord', usageRecordSchema);
