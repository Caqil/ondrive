import { BaseDocument, ObjectId, User } from ".";

export interface TeamSettings {
  allowMemberInvites: boolean;
  requireApprovalForJoining: boolean;
  defaultMemberRole: 'member' | 'editor' | 'admin';
  enableGuestAccess: boolean;
  enablePublicSharing: boolean;
  enforceSSO: boolean;
  allowedDomains: string[];
}

export interface TeamFeatures {
  enableAdvancedSharing: boolean;
  enableVersionHistory: boolean;
  enableAuditLogs: boolean;
  enableAPIAccess: boolean;
  enableIntegrations: boolean;
  maxFileSize: number;
  enableOCR: boolean;
}

export interface Team extends BaseDocument {
  name: string;
  description?: string;
  slug: string;
  logo?: string;
  color?: string;
  owner: ObjectId;
  billingEmail: string;
  subscription?: ObjectId;
  plan: 'free' | 'basic' | 'pro' | 'enterprise';
  storageUsed: number;
  storageQuota: number;
  memberCount: number;
  memberLimit: number;
  settings: TeamSettings;
  features: TeamFeatures;
  isActive: boolean;
  isSuspended: boolean;
  suspendedAt?: Date;
  suspendedReason?: string;
  trialEndsAt?: Date;
  isTrialActive: boolean;
}

export interface TeamMemberPermissions {
  canUpload: boolean;
  canDownload: boolean;
  canShare: boolean;
  canDelete: boolean;
  canInvite: boolean;
  canManageTeam: boolean;
  canViewBilling: boolean;
  canManageBilling: boolean;
}

export interface TeamMember extends BaseDocument {
  team: ObjectId;
  user: ObjectId;
  role: 'owner' | 'admin' | 'editor' | 'member' | 'viewer';
  permissions: TeamMemberPermissions;
  status: 'active' | 'pending' | 'suspended' | 'removed';
  invitedBy?: ObjectId;
  invitedAt?: Date;
  joinedAt?: Date;
  lastActiveAt?: Date;
}

export interface CreateTeamRequest {
  name: string;
  description?: string;
  slug: string;
  billingEmail?: string;
}

export interface InviteTeamMemberRequest {
  email: string;
  role: 'admin' | 'editor' | 'member' | 'viewer';
  message?: string;
}

export interface TeamWithMembers extends Team {
  members: (TeamMember & { user: User })[];
  pendingInvitations: number;
}
