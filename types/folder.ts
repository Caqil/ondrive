import { BaseDocument, ObjectId } from ".";

export interface Folder extends BaseDocument {
  name: string;
  description?: string;
  path: string;
  depth: number;
  parent?: ObjectId;
  children: ObjectId[];
  ancestors: ObjectId[];
  owner: ObjectId;
  team?: ObjectId;
  visibility: 'private' | 'team' | 'public';
  isStarred: boolean;
  isTrashed: boolean;
  trashedAt?: Date;
  trashedBy?: ObjectId;
  fileCount: number;
  folderCount: number;
  totalSize: number;
  color?: string;
  icon?: string;
  shareCount: number;
  syncStatus: 'synced' | 'pending' | 'conflict' | 'error';
  lastSyncAt?: Date;
}

export interface CreateFolderRequest {
  name: string;
  description?: string;
  parentId?: ObjectId;
  color?: string;
  icon?: string;
}

export interface FolderTreeNode {
  id: ObjectId;
  name: string;
  path: string;
  depth: number;
  children: FolderTreeNode[];
  fileCount: number;
  folderCount: number;
  totalSize: number;
  isExpanded?: boolean;
}

export interface FolderContents {
  folder: Folder;
  subfolders: Folder[];
  files: File[];
  breadcrumb: { id: ObjectId; name: string; path: string }[];
}

// types/share.ts
export interface ShareAccessLog {
  ip: string;
  userAgent: string;
  userId?: ObjectId;
  email?: string;
  accessedAt: Date;
  action: 'view' | 'download' | 'edit';
}

export interface ShareAllowedUser {
  email: string;
  permission: 'view' | 'comment' | 'edit';
  userId?: ObjectId;
}

export interface Share extends BaseDocument {
  token: string;
  resource: ObjectId;
  resourceType: 'file' | 'folder';
  owner: ObjectId;
  sharedBy: ObjectId;
  type: 'public' | 'restricted' | 'domain';
  permission: 'view' | 'comment' | 'edit';
  allowDownload: boolean;
  allowPrint: boolean;
  allowCopy: boolean;
  requireAuth: boolean;
  password?: string;
  expiresAt?: Date;
  isExpired: boolean;
  allowedDomains: string[];
  allowedUsers: ShareAllowedUser[];
  accessCount: number;
  lastAccessedAt?: Date;
  accessLog: ShareAccessLog[];
  isActive: boolean;
  isRevoked: boolean;
  revokedAt?: Date;
  revokedBy?: ObjectId;
}

export interface CreateShareRequest {
  resourceId: ObjectId;
  resourceType: 'file' | 'folder';
  type: 'public' | 'restricted' | 'domain';
  permission: 'view' | 'comment' | 'edit';
  allowDownload?: boolean;
  allowPrint?: boolean;
  allowCopy?: boolean;
  requireAuth?: boolean;
  password?: string;
  expiresAt?: Date;
  allowedDomains?: string[];
  allowedUsers?: ShareAllowedUser[];
}

export interface ShareLinkInfo {
  share: Share;
  resource: File | Folder;
  canAccess: boolean;
  requiresPassword: boolean;
  isExpired: boolean;
}
