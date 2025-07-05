import mongoose, { Document, Schema } from 'mongoose';

export interface IUser extends Document {
  _id: string;
  email: string;
  password?: string; // Optional for OAuth users
  name: string;
  avatar?: string;
  role: 'user' | 'admin' | 'moderator';
  emailVerified: boolean;
  emailVerificationToken?: string;
  passwordResetToken?: string;
  passwordResetExpires?: Date;
  twoFactorEnabled: boolean;
  twoFactorSecret?: string;
  lastLogin?: Date;
  isActive: boolean;
  isBanned: boolean;
  banReason?: string;
  bannedAt?: Date;
  bannedBy?: mongoose.Types.ObjectId;
  
  // Storage & Quota
  storageUsed: number; // bytes
  storageQuota: number; // bytes
  
  // Team
  currentTeam?: mongoose.Types.ObjectId;
  
  // OAuth
  providers: {
    google?: {
      id: string;
      email: string;
    };
    github?: {
      id: string;
      username: string;
    };
  };
  
  // Preferences
  preferences: {
    theme: 'light' | 'dark' | 'system';
    language: string;
    timezone: string;
    emailNotifications: boolean;
    pushNotifications: boolean;
    defaultView: 'grid' | 'list';
    uploadQuality: 'original' | 'high' | 'medium';
  };
  
  // Subscription
  subscription?: mongoose.Types.ObjectId;
  subscriptionStatus: 'active' | 'inactive' | 'trial' | 'expired' | 'cancelled';
  trialEndsAt?: Date;
  
  createdAt: Date;
  updatedAt: Date;
}

const userSchema = new Schema<IUser>({
  email: {
    type: String,
    required: true,
    unique: true,
    lowercase: true,
    trim: true,
    index: true
  },
  password: {
    type: String,
    select: false // Don't include password in queries by default
  },
  name: {
    type: String,
    required: true,
    trim: true,
    maxlength: 100
  },
  avatar: {
    type: String,
    validate: {
      validator: (v: string) => !v || /^https?:\/\/.+/.test(v),
      message: 'Avatar must be a valid URL'
    }
  },
  role: {
    type: String,
    enum: ['user', 'admin', 'moderator'],
    default: 'user',
    index: true
  },
  emailVerified: {
    type: Boolean,
    default: false,
    index: true
  },
  emailVerificationToken: String,
  passwordResetToken: String,
  passwordResetExpires: Date,
  twoFactorEnabled: {
    type: Boolean,
    default: false
  },
  twoFactorSecret: {
    type: String,
    select: false
  },
  lastLogin: Date,
  isActive: {
    type: Boolean,
    default: true,
    index: true
  },
  isBanned: {
    type: Boolean,
    default: false,
    index: true
  },
  banReason: String,
  bannedAt: Date,
  bannedBy: {
    type: Schema.Types.ObjectId,
    ref: 'User'
  },
  
  storageUsed: {
    type: Number,
    default: 0,
    min: 0
  },
  storageQuota: {
    type: Number,
    default: 15 * 1024 * 1024 * 1024, // 15GB default
    min: 0
  },
  
  currentTeam: {
    type: Schema.Types.ObjectId,
    ref: 'Team'
  },
  
  providers: {
    google: {
      id: String,
      email: String
    },
    github: {
      id: String,
      username: String
    }
  },
  
  preferences: {
    theme: {
      type: String,
      enum: ['light', 'dark', 'system'],
      default: 'system'
    },
    language: {
      type: String,
      default: 'en'
    },
    timezone: {
      type: String,
      default: 'UTC'
    },
    emailNotifications: {
      type: Boolean,
      default: true
    },
    pushNotifications: {
      type: Boolean,
      default: true
    },
    defaultView: {
      type: String,
      enum: ['grid', 'list'],
      default: 'grid'
    },
    uploadQuality: {
      type: String,
      enum: ['original', 'high', 'medium'],
      default: 'original'
    }
  },
  
  subscription: {
    type: Schema.Types.ObjectId,
    ref: 'Subscription'
  },
  subscriptionStatus: {
    type: String,
    enum: ['active', 'inactive', 'trial', 'expired', 'cancelled'],
    default: 'trial',
    index: true
  },
  trialEndsAt: Date
}, {
  timestamps: true,
  collection: 'users'
});

// Indexes
userSchema.index({ email: 1, isActive: 1 });
userSchema.index({ role: 1, isActive: 1 });
userSchema.index({ subscriptionStatus: 1 });
userSchema.index({ createdAt: -1 });

export const User = mongoose.models.User || mongoose.model<IUser>('User', userSchema);
