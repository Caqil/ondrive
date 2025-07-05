import mongoose, { Document, Schema } from 'mongoose';

export interface IApiKey extends Document {
  _id: string;
  
  // Owner
  user: mongoose.Types.ObjectId;
  team?: mongoose.Types.ObjectId;
  
  // Key details
  name: string;
  description?: string;
  key: string; // The actual API key (hashed)
  keyPreview: string; // First few characters for display
  
  // Permissions
  permissions: {
    read: boolean;
    write: boolean;
    delete: boolean;
    share: boolean;
    admin: boolean;
  };
  
  // Restrictions
  allowedIPs: string[]; // Empty array means no IP restrictions
  allowedDomains: string[]; // Empty array means no domain restrictions
  rateLimit: number; // Requests per minute, 0 means no limit
  
  // Usage tracking
  lastUsedAt?: Date;
  usageCount: number;
  
  // Status
  isActive: boolean;
  
  // Expiration
  expiresAt?: Date;
  
  createdAt: Date;
  updatedAt: Date;
}

const apiKeySchema = new Schema<IApiKey>({
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
  key: {
    type: String,
    required: true,
    unique: true,
    index: true,
    select: false // Don't include in queries by default
  },
  keyPreview: {
    type: String,
    required: true
  },
  
  permissions: {
    read: {
      type: Boolean,
      default: true
    },
    write: {
      type: Boolean,
      default: false
    },
    delete: {
      type: Boolean,
      default: false
    },
    share: {
      type: Boolean,
      default: false
    },
    admin: {
      type: Boolean,
      default: false
    }
  },
  
  allowedIPs: [{
    type: String,
    validate: {
      validator: (v: string) => /^(\d{1,3}\.){3}\d{1,3}(\/\d{1,2})?$/.test(v),
      message: 'Invalid IP address or CIDR notation'
    }
  }],
  allowedDomains: [{
    type: String,
    lowercase: true,
    trim: true
  }],
  rateLimit: {
    type: Number,
    default: 1000,
    min: 0
  },
  
  lastUsedAt: Date,
  usageCount: {
    type: Number,
    default: 0,
    min: 0
  },
  
  isActive: {
    type: Boolean,
    default: true,
    index: true
  },
  
  expiresAt: {
    type: Date,
    index: true
  }
}, {
  timestamps: true,
  collection: 'api_keys'
});

// Compound indexes
apiKeySchema.index({ user: 1, isActive: 1 });
apiKeySchema.index({ team: 1, isActive: 1 });
apiKeySchema.index({ key: 1, isActive: 1 });

// TTL index for expired keys
apiKeySchema.index({ expiresAt: 1 }, { expireAfterSeconds: 0 });

export const ApiKey = mongoose.models.ApiKey || mongoose.model<IApiKey>('ApiKey', apiKeySchema);
