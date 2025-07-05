import mongoose, { Document, Schema } from 'mongoose';

export interface IInvitation extends Document {
  _id: string;
  
  // What's being invited to
  invitationType: 'team' | 'share';
  team?: mongoose.Types.ObjectId;
  share?: mongoose.Types.ObjectId;
  
  // Who's involved
  inviter: mongoose.Types.ObjectId;
  inviteeEmail: string;
  invitee?: mongoose.Types.ObjectId; // Set when user exists
  
  // Invitation details
  role?: 'member' | 'editor' | 'admin'; // For team invitations
  permission?: 'view' | 'comment' | 'edit'; // For share invitations
  message?: string;
  
  // Status
  status: 'pending' | 'accepted' | 'declined' | 'expired' | 'revoked';
  token: string; // Unique invitation token
  
  // Timing
  expiresAt: Date;
  acceptedAt?: Date;
  declinedAt?: Date;
  revokedAt?: Date;
  revokedBy?: mongoose.Types.ObjectId;
  
  // Metadata
  metadata: {
    userAgent?: string;
    ip?: string;
    acceptedFrom?: string;
  };
  
  createdAt: Date;
  updatedAt: Date;
}

const invitationSchema = new Schema<IInvitation>({
  invitationType: {
    type: String,
    enum: ['team', 'share'],
    required: true,
    index: true
  },
  team: {
    type: Schema.Types.ObjectId,
    ref: 'Team',
    index: true
  },
  share: {
    type: Schema.Types.ObjectId,
    ref: 'Share',
    index: true
  },
  
  inviter: {
    type: Schema.Types.ObjectId,
    ref: 'User',
    required: true,
    index: true
  },
  inviteeEmail: {
    type: String,
    required: true,
    lowercase: true,
    trim: true,
    index: true
  },
  invitee: {
    type: Schema.Types.ObjectId,
    ref: 'User',
    index: true
  },
  
  role: {
    type: String,
    enum: ['member', 'editor', 'admin']
  },
  permission: {
    type: String,
    enum: ['view', 'comment', 'edit']
  },
  message: {
    type: String,
    trim: true,
    maxlength: 1000
  },
  
  status: {
    type: String,
    enum: ['pending', 'accepted', 'declined', 'expired', 'revoked'],
    default: 'pending',
    index: true
  },
  token: {
    type: String,
    required: true,
    unique: true,
    index: true
  },
  
  expiresAt: {
    type: Date,
    required: true,
    index: true
  },
  acceptedAt: Date,
  declinedAt: Date,
  revokedAt: Date,
  revokedBy: {
    type: Schema.Types.ObjectId,
    ref: 'User'
  },
  
  metadata: {
    userAgent: String,
    ip: String,
    acceptedFrom: String
  }
}, {
  timestamps: true,
  collection: 'invitations'
});

// Compound indexes
invitationSchema.index({ inviteeEmail: 1, status: 1 });
invitationSchema.index({ team: 1, status: 1 });
invitationSchema.index({ inviter: 1, status: 1 });
invitationSchema.index({ token: 1, status: 1 });

// TTL index for expired invitations
invitationSchema.index({ expiresAt: 1 }, { expireAfterSeconds: 0 });

// Validation
invitationSchema.pre('save', function(next) {
  if (this.invitationType === 'team' && !this.team) {
    return next(new Error('Team is required for team invitations'));
  }
  if (this.invitationType === 'share' && !this.share) {
    return next(new Error('Share is required for share invitations'));
  }
  if (this.invitationType === 'team' && !this.role) {
    return next(new Error('Role is required for team invitations'));
  }
  if (this.invitationType === 'share' && !this.permission) {
    return next(new Error('Permission is required for share invitations'));
  }
  next();
});

export const Invitation = mongoose.models.Invitation || mongoose.model<IInvitation>('Invitation', invitationSchema);
