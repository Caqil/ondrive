import mongoose, { Document, Schema } from 'mongoose';

export interface IShare extends Document {
  _id: string;
  token: string; // Unique share token
  
  // What's being shared
  resource: mongoose.Types.ObjectId;
  resourceType: 'file' | 'folder';
  
  // Who shared it
  owner: mongoose.Types.ObjectId;
  sharedBy: mongoose.Types.ObjectId; // Could be different from owner
  
  // Share settings
  type: 'public' | 'restricted' | 'domain';
  permission: 'view' | 'comment' | 'edit';
  
  // Access control
  allowDownload: boolean;
  allowPrint: boolean;
  allowCopy: boolean;
  requireAuth: boolean;
  password?: string; // Hashed password for protected shares
  
  // Expiration
  expiresAt?: Date;
  isExpired: boolean;
  
  // Domain restriction (for type: 'domain')
  allowedDomains: string[];
  
  // Specific users (for type: 'restricted')
  allowedUsers: {
    email: string;
    permission: 'view' | 'comment' | 'edit';
    userId?: mongoose.Types.ObjectId;
  }[];
  
  // Analytics
  accessCount: number;
  lastAccessedAt?: Date;
  accessLog: {
    ip: string;
    userAgent: string;
    userId?: mongoose.Types.ObjectId;
    email?: string;
    accessedAt: Date;
    action: 'view' | 'download' | 'edit';
  }[];
  
  // Status
  isActive: boolean;
  isRevoked: boolean;
  revokedAt?: Date;
  revokedBy?: mongoose.Types.ObjectId;
  
  createdAt: Date;
  updatedAt: Date;
}

const shareSchema = new Schema<IShare>({
  token: {
    type: String,
    required: true,
    unique: true,
    index: true
  },
  
  resource: {
    type: Schema.Types.ObjectId,
    required: true,
    refPath: 'resourceType',
    index: true
  },
  resourceType: {
    type: String,
    required: true,
    enum: ['file', 'folder']
  },
  
  owner: {
    type: Schema.Types.ObjectId,
    ref: 'User',
    required: true,
    index: true
  },
  sharedBy: {
    type: Schema.Types.ObjectId,
    ref: 'User',
    required: true,
    index: true
  },
  
  type: {
    type: String,
    enum: ['public', 'restricted', 'domain'],
    required: true,
    index: true
  },
  permission: {
    type: String,
    enum: ['view', 'comment', 'edit'],
    required: true,
    index: true
  },
  
  allowDownload: {
    type: Boolean,
    default: true
  },
  allowPrint: {
    type: Boolean,
    default: true
  },
  allowCopy: {
    type: Boolean,
    default: true
  },
  requireAuth: {
    type: Boolean,
    default: false
  },
  password: {
    type: String,
    select: false
  },
  
  expiresAt: {
    type: Date,
    index: true
  },
  isExpired: {
    type: Boolean,
    default: false,
    index: true
  },
  
  allowedDomains: [{
    type: String,
    lowercase: true,
    trim: true
  }],
  
  allowedUsers: [{
    email: {
      type: String,
      required: true,
      lowercase: true,
      trim: true
    },
    permission: {
      type: String,
      enum: ['view', 'comment', 'edit'],
      required: true
    },
    userId: {
      type: Schema.Types.ObjectId,
      ref: 'User'
    }
  }],
  
  accessCount: {
    type: Number,
    default: 0
  },
  lastAccessedAt: Date,
  accessLog: [{
    ip: {
      type: String,
      required: true
    },
    userAgent: {
      type: String,
      required: true
    },
    userId: {
      type: Schema.Types.ObjectId,
      ref: 'User'
    },
    email: String,
    accessedAt: {
      type: Date,
      required: true,
      default: Date.now
    },
    action: {
      type: String,
      enum: ['view', 'download', 'edit'],
      required: true
    }
  }],
  
  isActive: {
    type: Boolean,
    default: true,
    index: true
  },
  isRevoked: {
    type: Boolean,
    default: false,
    index: true
  },
  revokedAt: Date,
  revokedBy: {
    type: Schema.Types.ObjectId,
    ref: 'User'
  }
}, {
  timestamps: true,
  collection: 'shares'
});

// Compound indexes
shareSchema.index({ resource: 1, resourceType: 1 });
shareSchema.index({ owner: 1, isActive: 1 });
shareSchema.index({ expiresAt: 1, isExpired: 1 });
shareSchema.index({ token: 1, isActive: 1, isExpired: 1 });

// TTL index for expired shares
shareSchema.index({ expiresAt: 1 }, { expireAfterSeconds: 0 });

export const Share = mongoose.models.Share || mongoose.model<IShare>('Share', shareSchema);
