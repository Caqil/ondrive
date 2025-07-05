import mongoose, { Document, Schema } from 'mongoose';

export interface ITeam extends Document {
  _id: string;
  name: string;
  description?: string;
  slug: string; // URL-friendly identifier
  
  // Branding
  logo?: string;
  color?: string;
  
  // Owner and billing
  owner: mongoose.Types.ObjectId;
  billingEmail: string;
  
  // Subscription and limits
  subscription?: mongoose.Types.ObjectId;
  plan: 'free' | 'basic' | 'pro' | 'enterprise';
  
  // Storage and usage
  storageUsed: number; // bytes
  storageQuota: number; // bytes
  memberCount: number;
  memberLimit: number;
  
  // Settings
  settings: {
    allowMemberInvites: boolean;
    requireApprovalForJoining: boolean;
    defaultMemberRole: 'member' | 'editor' | 'admin';
    enableGuestAccess: boolean;
    enablePublicSharing: boolean;
    enforceSSO: boolean;
    allowedDomains: string[];
  };
  
  // Features enabled
  features: {
    enableAdvancedSharing: boolean;
    enableVersionHistory: boolean;
    enableAuditLogs: boolean;
    enableAPIAccess: boolean;
    enableIntegrations: boolean;
    maxFileSize: number;
    enableOCR: boolean;
  };
  
  // Status
  isActive: boolean;
  isSuspended: boolean;
  suspendedAt?: Date;
  suspendedReason?: string;
  
  // Trial
  trialEndsAt?: Date;
  isTrialActive: boolean;
  
  createdAt: Date;
  updatedAt: Date;
}

const teamSchema = new Schema<ITeam>({
  name: {
    type: String,
    required: true,
    trim: true,
    maxlength: 100,
    index: true
  },
  description: {
    type: String,
    trim: true,
    maxlength: 500
  },
  slug: {
    type: String,
    required: true,
    unique: true,
    lowercase: true,
    trim: true,
    match: /^[a-z0-9-]+$/,
    index: true
  },
  
  logo: String,
  color: {
    type: String,
    match: /^#[0-9A-F]{6}$/i
  },
  
  owner: {
    type: Schema.Types.ObjectId,
    ref: 'User',
    required: true,
    index: true
  },
  billingEmail: {
    type: String,
    required: true,
    lowercase: true,
    trim: true
  },
  
  subscription: {
    type: Schema.Types.ObjectId,
    ref: 'Subscription'
  },
  plan: {
    type: String,
    enum: ['free', 'basic', 'pro', 'enterprise'],
    default: 'free',
    index: true
  },
  
  storageUsed: {
    type: Number,
    default: 0,
    min: 0
  },
  storageQuota: {
    type: Number,
    default: 100 * 1024 * 1024 * 1024, // 100GB default
    min: 0
  },
  memberCount: {
    type: Number,
    default: 1,
    min: 1
  },
  memberLimit: {
    type: Number,
    default: 5
  },
  
  settings: {
    allowMemberInvites: {
      type: Boolean,
      default: true
    },
    requireApprovalForJoining: {
      type: Boolean,
      default: false
    },
    defaultMemberRole: {
      type: String,
      enum: ['member', 'editor', 'admin'],
      default: 'member'
    },
    enableGuestAccess: {
      type: Boolean,
      default: true
    },
    enablePublicSharing: {
      type: Boolean,
      default: true
    },
    enforceSSO: {
      type: Boolean,
      default: false
    },
    allowedDomains: [{
      type: String,
      lowercase: true,
      trim: true
    }]
  },
  
  features: {
    enableAdvancedSharing: {
      type: Boolean,
      default: false
    },
    enableVersionHistory: {
      type: Boolean,
      default: true
    },
    enableAuditLogs: {
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
    maxFileSize: {
      type: Number,
      default: 100 * 1024 * 1024 // 100MB
    },
    enableOCR: {
      type: Boolean,
      default: false
    }
  },
  
  isActive: {
    type: Boolean,
    default: true,
    index: true
  },
  isSuspended: {
    type: Boolean,
    default: false,
    index: true
  },
  suspendedAt: Date,
  suspendedReason: String,
  
  trialEndsAt: Date,
  isTrialActive: {
    type: Boolean,
    default: false,
    index: true
  }
}, {
  timestamps: true,
  collection: 'teams'
});

// Indexes
teamSchema.index({ owner: 1, isActive: 1 });
teamSchema.index({ plan: 1, isActive: 1 });
teamSchema.index({ trialEndsAt: 1, isTrialActive: 1 });

export const Team = mongoose.models.Team || mongoose.model<ITeam>('Team', teamSchema);
