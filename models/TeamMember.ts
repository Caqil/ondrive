import mongoose, { Document, Schema } from 'mongoose';

export interface ITeamMember extends Document {
  _id: string;
  team: mongoose.Types.ObjectId;
  user: mongoose.Types.ObjectId;
  
  // Role and permissions
  role: 'owner' | 'admin' | 'editor' | 'member' | 'viewer';
  permissions: {
    canUpload: boolean;
    canDownload: boolean;
    canShare: boolean;
    canDelete: boolean;
    canInvite: boolean;
    canManageTeam: boolean;
    canViewBilling: boolean;
    canManageBilling: boolean;
  };
  
  // Status
  status: 'active' | 'pending' | 'suspended' | 'removed';
  
  // Invitation details
  invitedBy?: mongoose.Types.ObjectId;
  invitedAt?: Date;
  joinedAt?: Date;
  
  // Activity
  lastActiveAt?: Date;
  
  createdAt: Date;
  updatedAt: Date;
}

const teamMemberSchema = new Schema<ITeamMember>({
  team: {
    type: Schema.Types.ObjectId,
    ref: 'Team',
    required: true,
    index: true
  },
  user: {
    type: Schema.Types.ObjectId,
    ref: 'User',
    required: true,
    index: true
  },
  
  role: {
    type: String,
    enum: ['owner', 'admin', 'editor', 'member', 'viewer'],
    required: true,
    index: true
  },
  permissions: {
    canUpload: {
      type: Boolean,
      default: true
    },
    canDownload: {
      type: Boolean,
      default: true
    },
    canShare: {
      type: Boolean,
      default: true
    },
    canDelete: {
      type: Boolean,
      default: false
    },
    canInvite: {
      type: Boolean,
      default: false
    },
    canManageTeam: {
      type: Boolean,
      default: false
    },
    canViewBilling: {
      type: Boolean,
      default: false
    },
    canManageBilling: {
      type: Boolean,
      default: false
    }
  },
  
  status: {
    type: String,
    enum: ['active', 'pending', 'suspended', 'removed'],
    default: 'active',
    index: true
  },
  
  invitedBy: {
    type: Schema.Types.ObjectId,
    ref: 'User'
  },
  invitedAt: Date,
  joinedAt: Date,
  
  lastActiveAt: Date
}, {
  timestamps: true,
  collection: 'team_members'
});

// Compound indexes
teamMemberSchema.index({ team: 1, user: 1 }, { unique: true });
teamMemberSchema.index({ team: 1, status: 1 });
teamMemberSchema.index({ user: 1, status: 1 });

export const TeamMember = mongoose.models.TeamMember || mongoose.model<ITeamMember>('TeamMember', teamMemberSchema);
