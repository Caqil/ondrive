import type { User, TeamMember, File, Folder, Share, Team } from '@/types';
import { 
  USER_ROLES, 
  TEAM_ROLES, 
  TEAM_PERMISSIONS, 
  SUBSCRIPTION_STATUS,
  SHARE_PERMISSIONS,
  VISIBILITY_OPTIONS 
} from '@/lib/constants';

/**
 * Check if user has required role
 */
export function hasRole(user: User | null, requiredRole: string): boolean {
  if (!user) return false;
  
  const roleHierarchy = {
    [USER_ROLES.VIEWER]: 0,
    [USER_ROLES.USER]: 1,
    [USER_ROLES.MODERATOR]: 2,
    [USER_ROLES.ADMIN]: 3,
  };

  const userRoleLevel = roleHierarchy[user.role as keyof typeof roleHierarchy] ?? 0;
  const requiredRoleLevel = roleHierarchy[requiredRole as keyof typeof roleHierarchy] ?? 0;

  return userRoleLevel >= requiredRoleLevel;
}

/**
 * Check if user has specific permission
 */
export function hasPermission(
  user: User | null,
  action: string,
  resource?: any
): boolean {
  if (!user || user.isBanned) return false;

  // Admin can do everything
  if (user.role === USER_ROLES.ADMIN) return true;

  // Check email verification for sensitive actions
  const sensitiveActions = ['upload', 'share', 'delete', 'admin', 'create_team', 'invite_member'];
  if (sensitiveActions.includes(action) && !user.emailVerified) {
    return false;
  }

  // Resource ownership check
  if (resource && resource.owner && resource.owner.toString() === user._id.toString()) {
    return true;
  }

  // Role-based permissions
  const permissions = {
    [USER_ROLES.VIEWER]: {
      read: true,
      download: true,
      upload: false,
      create: false,
      update: false,
      rename: false,
      delete: false,
      share: false,
      admin: false,
      create_team: false,
      invite_member: false,
    },
    [USER_ROLES.USER]: {
      read: true,
      download: true,
      upload: true,
      create: true,
      update: true,
      rename: true,
      delete: false,
      share: true,
      admin: false,
      create_team: true,
      invite_member: false,
    },
    [USER_ROLES.MODERATOR]: {
      read: true,
      download: true,
      upload: true,
      create: true,
      update: true,
      rename: true,
      delete: true,
      share: true,
      admin: false,
      create_team: true,
      invite_member: true,
    },
    [USER_ROLES.ADMIN]: {
      read: true,
      download: true,
      upload: true,
      create: true,
      update: true,
      rename: true,
      delete: true,
      share: true,
      admin: true,
      create_team: true,
      invite_member: true,
    },
  };

  const userPermissions = permissions[user.role as keyof typeof permissions];
  return userPermissions?.[action as keyof typeof userPermissions] || false;
}

/**
 * Check team member permissions
 */
export function hasTeamPermission(
  teamMember: TeamMember | null,
  action: keyof typeof TEAM_PERMISSIONS.owner
): boolean {
  if (!teamMember || teamMember.status !== 'active') return false;

  const permissions = TEAM_PERMISSIONS[teamMember.role];
  return permissions?.[action] || false;
}

/**
 * Check if user can access resource
 */
export function canAccessResource(
  user: User | null,
  resource: File | Folder,
  action: string = 'read'
): boolean {
  if (!user || !resource) return false;

  // Check basic permission
  if (!hasPermission(user, action, resource)) return false;

  // Public resources
  if (resource.visibility === VISIBILITY_OPTIONS.PUBLIC) return true;

  // Owner access
  if (resource.owner && resource.owner.toString() === user._id.toString()) {
    return true;
  }

  // Team access
  if (resource.visibility === VISIBILITY_OPTIONS.TEAM && resource.team) {
    return user.currentTeam?.toString() === resource.team.toString();
  }

  // Private resources - only owner can access
  if (resource.visibility === VISIBILITY_OPTIONS.PRIVATE) {
    return !!resource.owner && resource.owner.toString() === user._id.toString();
  }

  return false;
}

/**
 * Check if user can access shared resource
 */
export function canAccessSharedResource(
  user: User | null,
  share: Share,
  action: string = 'view'
): boolean {
  if (!share.isActive || share.isRevoked) return false;

  // Check if share has expired
  if (share.expiresAt && new Date() > new Date(share.expiresAt)) return false;

  // Public shares
  if (share.type === 'public') {
    return canPerformShareAction(share, action);
  }

  // Restricted shares
  if (share.type === 'restricted') {
    if (!user) return false;
    
    const allowedUser = share.allowedUsers.find(
      u => u.email === user.email || u.userId?.toString() === user._id.toString()
    );
    
    if (!allowedUser) return false;
    
    return canPerformShareAction(share, action, allowedUser.permission as 'view' | 'comment' | 'edit');
  }

  // Domain restricted shares
  if (share.type === 'domain') {
    if (!user) return false;
    
    const userDomain = user.email.split('@')[1];
    if (!share.allowedDomains.includes(userDomain)) return false;
    
    return canPerformShareAction(share, action);
  }

  return false;
}

/**
 * Check if action can be performed on shared resource
 */
function canPerformShareAction(
  share: Share,
  action: string,
  userPermission?: 'view' | 'comment' | 'edit'
): boolean {
  const permission = userPermission || share.permission;
  
  switch (action) {
    case 'view':
    case 'read':
      return [SHARE_PERMISSIONS.VIEW, SHARE_PERMISSIONS.COMMENT, SHARE_PERMISSIONS.EDIT].includes(permission);
    
    case 'comment':
      return [SHARE_PERMISSIONS.COMMENT, SHARE_PERMISSIONS.EDIT].includes(permission as any);
    
    case 'edit':
    case 'update':
      return permission === SHARE_PERMISSIONS.EDIT;
    
    case 'download':
      return share.allowDownload;
    
    case 'print':
      return share.allowPrint;
    
    case 'copy':
      return share.allowCopy;
    
    default:
      return false;
  }
}

/**
 * Check if user can manage team
 */
export function canManageTeam(user: User | null, team: Team, teamMember?: TeamMember): boolean {
  if (!user || !team) return false;

  // Admin can manage any team
  if (user.role === USER_ROLES.ADMIN) return true;

  // Team owner can always manage
  if (team.owner.toString() === user._id.toString()) return true;

  // Check team member permissions
  if (teamMember) {
    return hasTeamPermission(teamMember, 'canManageTeam');
  }

  return false;
}

/**
 * Check if user can invite team members
 */
export function canInviteTeamMembers(user: User | null, team: Team, teamMember?: TeamMember): boolean {
  if (!user || !team) return false;

  // Check if team allows member invites
  if (!team.settings.allowMemberInvites && teamMember?.role !== TEAM_ROLES.OWNER && teamMember?.role !== TEAM_ROLES.ADMIN) {
    return false;
  }

  // Check team member limit
  if (team.memberCount >= team.memberLimit) return false;

  // Admin can invite to any team
  if (user.role === USER_ROLES.ADMIN) return true;

  // Team owner can always invite
  if (team.owner.toString() === user._id.toString()) return true;

  // Check team member permissions
  if (teamMember) {
    return hasTeamPermission(teamMember, 'canInvite');
  }

  return false;
}

/**
 * Check if user can view billing information
 */
export function canViewBilling(user: User | null, team?: Team, teamMember?: TeamMember): boolean {
  if (!user) return false;

  // Admin can view any billing
  if (user.role === USER_ROLES.ADMIN) return true;

  // For personal billing
  if (!team) return true;

  // Team owner can always view billing
  if (team.owner.toString() === user._id.toString()) return true;

  // Check team member permissions
  if (teamMember) {
    return hasTeamPermission(teamMember, 'canViewBilling');
  }

  return false;
}

/**
 * Check if user can manage billing
 */
export function canManageBilling(user: User | null, team?: Team, teamMember?: TeamMember): boolean {
  if (!user) return false;

  // Admin can manage any billing
  if (user.role === USER_ROLES.ADMIN) return true;

  // For personal billing
  if (!team) return true;

  // Team owner can always manage billing
  if (team.owner.toString() === user._id.toString()) return true;

  // Check team member permissions
  if (teamMember) {
    return hasTeamPermission(teamMember, 'canManageBilling');
  }

  return false;
}

/**
 * Get user storage quota based on subscription
 */
export function getUserStorageQuota(user: User): number {
  // Admin users get unlimited storage
  if (user.role === USER_ROLES.ADMIN) {
    return Number.MAX_SAFE_INTEGER;
  }

  // Use explicit quota if set
  if (user.storageQuota) {
    return user.storageQuota;
  }

  // Subscription-based quotas
  switch (user.subscriptionStatus) {
    case SUBSCRIPTION_STATUS.ACTIVE:
      return 100 * 1024 * 1024 * 1024; // 100GB for active subscription
    case SUBSCRIPTION_STATUS.TRIAL:
      return 50 * 1024 * 1024 * 1024; // 50GB for trial
    case SUBSCRIPTION_STATUS.EXPIRED:
    case SUBSCRIPTION_STATUS.CANCELLED:
      return 5 * 1024 * 1024 * 1024; // 5GB for expired/cancelled
    default:
      return 15 * 1024 * 1024 * 1024; // 15GB default
  }
}

/**
 * Check if user has exceeded storage quota
 */
export function hasExceededStorageQuota(user: User): boolean {
  const quota = getUserStorageQuota(user);
  return user.storageUsed >= quota;
}

/**
 * Get storage usage percentage
 */
export function getStorageUsagePercentage(user: User): number {
  const quota = getUserStorageQuota(user);
  if (quota === Number.MAX_SAFE_INTEGER) return 0; // Unlimited
  return Math.min(100, (user.storageUsed / quota) * 100);
}

/**
 * Check if user can upload files
 */
export function canUploadFiles(user: User | null, targetFolder?: Folder): boolean {
  if (!user) return false;

  // Check basic upload permission
  if (!hasPermission(user, 'upload')) return false;

  // Check storage quota
  if (hasExceededStorageQuota(user)) return false;

  // Check folder access if specified
  if (targetFolder && !canAccessResource(user, targetFolder, 'update')) {
    return false;
  }

  return true;
}

/**
 * Check if user can create shares
 */
export function canCreateShares(user: User | null, resource: File | Folder): boolean {
  if (!user) return false;

  // Check basic share permission
  if (!hasPermission(user, 'share')) return false;

  // Check resource access
  if (!canAccessResource(user, resource, 'read')) return false;

  // Check if resource owner allows sharing
  if (resource.visibility === VISIBILITY_OPTIONS.PRIVATE && 
      resource.owner.toString() !== user._id.toString()) {
    return false;
  }

  return true;
}

/**
 * Check if user can delete files/folders
 */
export function canDeleteResource(user: User | null, resource: File | Folder): boolean {
  if (!user) return false;

  // Owner can always delete (unless it's a system resource)
  if (resource.owner.toString() === user._id.toString()) {
    return hasPermission(user, 'delete');
  }

  // Admin and moderator can delete any resource
  if (user.role === USER_ROLES.ADMIN || user.role === USER_ROLES.MODERATOR) {
    return true;
  }

  return false;
}

/**
 * Check if user can restore from trash
 */
export function canRestoreFromTrash(user: User | null, resource: File | Folder): boolean {
  if (!user) return false;

  // Only trashed items can be restored
  if (!resource.isTrashed) return false;

  // Owner can restore
  if (resource.owner.toString() === user._id.toString()) {
    return true;
  }

  // Admin can restore any item
  if (user.role === USER_ROLES.ADMIN) {
    return true;
  }

  return false;
}

/**
 * Get effective permissions for user
 */
export function getEffectivePermissions(user: User | null): Record<string, boolean> {
  if (!user) {
    return {
      read: false,
      download: false,
      upload: false,
      create: false,
      update: false,
      delete: false,
      share: false,
      admin: false,
      create_team: false,
      invite_member: false,
    };
  }

  const basePermissions = {
    read: hasPermission(user, 'read'),
    download: hasPermission(user, 'download'),
    upload: hasPermission(user, 'upload'),
    create: hasPermission(user, 'create'),
    update: hasPermission(user, 'update'),
    delete: hasPermission(user, 'delete'),
    share: hasPermission(user, 'share'),
    admin: hasPermission(user, 'admin'),
    create_team: hasPermission(user, 'create_team'),
    invite_member: hasPermission(user, 'invite_member'),
  };

  return basePermissions;
}

/**
 * Get team member effective permissions
 */
export function getTeamMemberEffectivePermissions(teamMember: TeamMember | null): Record<string, boolean> {
  if (!teamMember || teamMember.status !== 'active') {
    return {
      canUpload: false,
      canDownload: false,
      canShare: false,
      canDelete: false,
      canInvite: false,
      canManageTeam: false,
      canViewBilling: false,
      canManageBilling: false,
    };
  }

  return {
    canUpload: hasTeamPermission(teamMember, 'canUpload'),
    canDownload: hasTeamPermission(teamMember, 'canDownload'),
    canShare: hasTeamPermission(teamMember, 'canShare'),
    canDelete: hasTeamPermission(teamMember, 'canDelete'),
    canInvite: hasTeamPermission(teamMember, 'canInvite'),
    canManageTeam: hasTeamPermission(teamMember, 'canManageTeam'),
    canViewBilling: hasTeamPermission(teamMember, 'canViewBilling'),
    canManageBilling: hasTeamPermission(teamMember, 'canManageBilling'),
  };
}

/**
 * Check if user requires subscription for feature
 */
export function requiresSubscription(user: User | null, feature: string): boolean {
  if (!user) return true;

  // Admin users have access to all features
  if (user.role === USER_ROLES.ADMIN) return false;

  // Features that require active subscription
  const premiumFeatures = [
    'advanced_sharing',
    'version_history',
    'ocr',
    'api_access',
    'integrations',
    'priority_support',
    'audit_logs',
    'sso',
  ];

  if (!premiumFeatures.includes(feature)) return false;

  // Check if user has active subscription
  return user.subscriptionStatus !== SUBSCRIPTION_STATUS.ACTIVE;
}

/**
 * Check if user can perform admin actions
 */
export function canPerformAdminActions(user: User | null): boolean {
  return user?.role === USER_ROLES.ADMIN;
}

/**
 * Check if user can impersonate other users
 */
export function canImpersonateUsers(user: User | null): boolean {
  return user?.role === USER_ROLES.ADMIN;
}

/**
 * Check if user can access analytics
 */
export function canAccessAnalytics(user: User | null): boolean {
  if (!user) return false;
  return user.role === USER_ROLES.ADMIN || user.role === USER_ROLES.MODERATOR;
}

/**
 * Check if user can manage system settings
 */
export function canManageSystemSettings(user: User | null): boolean {
  return user?.role === USER_ROLES.ADMIN;
}