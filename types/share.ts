import { Folder } from ".";

export type ShareRole = 'viewer' | 'commenter' | 'editor';
export type ShareType = 'link' | 'email' | 'internal';
export type ShareResourceType = 'file' | 'folder';

export interface SharePermission {
  user?: string;
  email?: string;
  role: ShareRole;
  canDownload: boolean;
  canShare: boolean;
  expiresAt?: string;
}

export interface ShareSettings {
  allowDownload: boolean;
  allowPrint: boolean;
  allowCopy: boolean;
  allowComments: boolean;
  allowAnonymous: boolean;
  notifyOnAccess: boolean;
}

export interface Share {
  id: string;
  resource: string;
  resourceType: ShareResourceType;
  owner: string;
  shareType: ShareType;
  shareToken: string;
  isPublic: boolean;
  requiresPassword: boolean;
  password?: string;
  permissions: SharePermission[];
  settings: ShareSettings;
  accessCount: number;
  lastAccessedAt?: string;
  expiresAt?: string;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
  // Computed properties
  isExpired: boolean;
  accessUrl: string;
}

export interface ShareCreate {
  resourceId: string;
  resourceType: ShareResourceType;
  shareType: ShareType;
  permissions: Omit<SharePermission, 'user' | 'email'>[];
  users?: string[];
  emails?: string[];
  settings: ShareSettings;
  expiresAt?: string;
  password?: string;
}

export interface ShareUpdate {
  permissions?: SharePermission[];
  settings?: Partial<ShareSettings>;
  expiresAt?: string;
  password?: string;
  isActive?: boolean;
}

export interface ShareAccess {
  shareToken: string;
  password?: string;
  email?: string;
}

export interface ShareAccessResponse {
  success: boolean;
  share?: Share;
  resource?: File | Folder;
  canDownload: boolean;
  canComment: boolean;
  canEdit: boolean;
  error?: string;
}


