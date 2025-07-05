import mongoose, { Document, Schema } from 'mongoose';

export interface INotification extends Document {
  _id: string;
  
  // Recipient
  user: mongoose.Types.ObjectId;
  
  // Notification details
  type: 'share_received' | 'share_accepted' | 'file_uploaded' | 'file_deleted' | 'team_invite' | 'payment_success' | 'payment_failed' | 'storage_limit' | 'trial_ending' | 'subscription_cancelled' | 'security_alert';
  title: string;
  message: string;
  
  // Related resources
  relatedResource?: mongoose.Types.ObjectId;
  relatedResourceType?: 'file' | 'folder' | 'share' | 'team' | 'subscription' | 'payment';
  
  // Action details
  actionUrl?: string;
  actionText?: string;
  
  // Status
  isRead: boolean;
  readAt?: Date;
  
  // Delivery
  channels: {
    inApp: boolean;
    email: boolean;
    push: boolean;
  };
  deliveryStatus: {
    inApp: 'pending' | 'delivered' | 'failed';
    email: 'pending' | 'sent' | 'delivered' | 'failed';
    push: 'pending' | 'sent' | 'delivered' | 'failed';
  };
  
  // Priority
  priority: 'low' | 'normal' | 'high' | 'urgent';
  
  // Metadata
  metadata: {
    [key: string]: any;
  };
  
  // Expiration
  expiresAt?: Date;
  
  createdAt: Date;
  updatedAt: Date;
}

const notificationSchema = new Schema<INotification>({
  user: {
    type: Schema.Types.ObjectId,
    ref: 'User',
    required: true,
    index: true
  },
  
  type: {
    type: String,
    enum: [
      'share_received', 'share_accepted', 'file_uploaded', 'file_deleted',
      'team_invite', 'payment_success', 'payment_failed', 'storage_limit',
      'trial_ending', 'subscription_cancelled', 'security_alert'
    ],
    required: true,
    index: true
  },
  title: {
    type: String,
    required: true,
    trim: true,
    maxlength: 200
  },
  message: {
    type: String,
    required: true,
    trim: true,
    maxlength: 1000
  },
  
  relatedResource: {
    type: Schema.Types.ObjectId,
    refPath: 'relatedResourceType',
    index: true
  },
  relatedResourceType: {
    type: String,
    enum: ['file', 'folder', 'share', 'team', 'subscription', 'payment']
  },
  
  actionUrl: String,
  actionText: {
    type: String,
    maxlength: 50
  },
  
  isRead: {
    type: Boolean,
    default: false,
    index: true
  },
  readAt: Date,
  
  channels: {
    inApp: {
      type: Boolean,
      default: true
    },
    email: {
      type: Boolean,
      default: false
    },
    push: {
      type: Boolean,
      default: false
    }
  },
  deliveryStatus: {
    inApp: {
      type: String,
      enum: ['pending', 'delivered', 'failed'],
      default: 'pending'
    },
    email: {
      type: String,
      enum: ['pending', 'sent', 'delivered', 'failed'],
      default: 'pending'
    },
    push: {
      type: String,
      enum: ['pending', 'sent', 'delivered', 'failed'],
      default: 'pending'
    }
  },
  
  priority: {
    type: String,
    enum: ['low', 'normal', 'high', 'urgent'],
    default: 'normal',
    index: true
  },
  
  metadata: {
    type: Schema.Types.Mixed,
    default: {}
  },
  
  expiresAt: {
    type: Date,
    index: true
  }
}, {
  timestamps: true,
  collection: 'notifications'
});

// Compound indexes
notificationSchema.index({ user: 1, isRead: 1, createdAt: -1 });
notificationSchema.index({ user: 1, type: 1, createdAt: -1 });
notificationSchema.index({ user: 1, priority: 1, isRead: 1 });

// TTL index for expired notifications
notificationSchema.index({ expiresAt: 1 }, { expireAfterSeconds: 0 });

export const Notification = mongoose.models.Notification || mongoose.model<INotification>('Notification', notificationSchema);
