import mongoose, { Document, Schema } from 'mongoose';

export interface IActivity extends Document {
  _id: string;
  
  // Who performed the action
  user: mongoose.Types.ObjectId;
  impersonatedBy?: mongoose.Types.ObjectId; // If admin is impersonating
  
  // What action was performed
  action: 'create' | 'read' | 'update' | 'delete' | 'share' | 'download' | 'upload' | 'move' | 'copy' | 'rename' | 'star' | 'unstar' | 'trash' | 'restore';
  
  // What resource was affected
  resource: mongoose.Types.ObjectId;
  resourceType: 'file' | 'folder' | 'share' | 'user' | 'team' | 'settings';
  resourceName: string; // For display purposes
  
  // Additional context
  metadata: {
    oldValue?: any;
    newValue?: any;
    fileSize?: number;
    mimeType?: string;
    shareType?: string;
    permission?: string;
    [key: string]: any;
  };
  
  // Request context
  ip: string;
  userAgent: string;
  location?: {
    country?: string;
    city?: string;
    coordinates?: [number, number]; // [longitude, latitude]
  };
  
  // Team context
  team?: mongoose.Types.ObjectId;
  
  // Success/Error status
  status: 'success' | 'error' | 'warning';
  errorMessage?: string;
  
  // Processing time
  duration?: number; // milliseconds
  
  createdAt: Date;
}

const activitySchema = new Schema<IActivity>({
  user: {
    type: Schema.Types.ObjectId,
    ref: 'User',
    required: true,
    index: true
  },
  impersonatedBy: {
    type: Schema.Types.ObjectId,
    ref: 'User',
    index: true
  },
  
  action: {
    type: String,
    enum: ['create', 'read', 'update', 'delete', 'share', 'download', 'upload', 'move', 'copy', 'rename', 'star', 'unstar', 'trash', 'restore'],
    required: true,
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
    enum: ['file', 'folder', 'share', 'user', 'team', 'settings'],
    index: true
  },
  resourceName: {
    type: String,
    required: true,
    maxlength: 255
  },
  
  metadata: {
    type: Schema.Types.Mixed,
    default: {}
  },
  
  ip: {
    type: String,
    required: true,
    index: true
  },
  userAgent: {
    type: String,
    required: true
  },
  location: {
    country: String,
    city: String,
    coordinates: {
      type: [Number],
      index: '2dsphere'
    }
  },
  
  team: {
    type: Schema.Types.ObjectId,
    ref: 'Team',
    index: true
  },
  
  status: {
    type: String,
    enum: ['success', 'error', 'warning'],
    default: 'success',
    index: true
  },
  errorMessage: String,
  
  duration: {
    type: Number,
    min: 0
  }
}, {
  timestamps: { createdAt: true, updatedAt: false },
  collection: 'activities'
});

// Compound indexes
activitySchema.index({ user: 1, createdAt: -1 });
activitySchema.index({ resource: 1, resourceType: 1, createdAt: -1 });
activitySchema.index({ team: 1, createdAt: -1 });
activitySchema.index({ action: 1, status: 1, createdAt: -1 });

// TTL index to automatically delete old activities (90 days)
activitySchema.index({ createdAt: 1 }, { expireAfterSeconds: 90 * 24 * 60 * 60 });

export const Activity = mongoose.models.Activity || mongoose.model<IActivity>('Activity', activitySchema);
