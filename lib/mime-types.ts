/**
 * MIME type mappings for file extensions
 */
export const MIME_TYPES: Record<string, string> = {
  // Images
  'jpg': 'image/jpeg',
  'jpeg': 'image/jpeg',
  'png': 'image/png',
  'gif': 'image/gif',
  'webp': 'image/webp',
  'svg': 'image/svg+xml',
  'bmp': 'image/bmp',
  'ico': 'image/x-icon',
  'tiff': 'image/tiff',
  'tif': 'image/tiff',
  
  // Videos
  'mp4': 'video/mp4',
  'avi': 'video/x-msvideo',
  'mov': 'video/quicktime',
  'wmv': 'video/x-ms-wmv',
  'flv': 'video/x-flv',
  'webm': 'video/webm',
  'mkv': 'video/x-matroska',
  'm4v': 'video/x-m4v',
  '3gp': 'video/3gpp',
  'ogv': 'video/ogg',
  
  // Audio
  'mp3': 'audio/mpeg',
  'wav': 'audio/wav',
  'ogg': 'audio/ogg',
  'aac': 'audio/aac',
  'flac': 'audio/flac',
  'm4a': 'audio/mp4',
  'wma': 'audio/x-ms-wma',
  'aiff': 'audio/aiff',
  'au': 'audio/basic',
  
  // Documents
  'pdf': 'application/pdf',
  'doc': 'application/msword',
  'docx': 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
  'xls': 'application/vnd.ms-excel',
  'xlsx': 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
  'ppt': 'application/vnd.ms-powerpoint',
  'pptx': 'application/vnd.openxmlformats-officedocument.presentationml.presentation',
  'txt': 'text/plain',
  'rtf': 'application/rtf',
  'odt': 'application/vnd.oasis.opendocument.text',
  'ods': 'application/vnd.oasis.opendocument.spreadsheet',
  'odp': 'application/vnd.oasis.opendocument.presentation',
  
  // Archives
  'zip': 'application/zip',
  'rar': 'application/vnd.rar',
  '7z': 'application/x-7z-compressed',
  'tar': 'application/x-tar',
  'gz': 'application/gzip',
  'bz2': 'application/x-bzip2',
  'xz': 'application/x-xz',
  
  // Code
  'js': 'text/javascript',
  'ts': 'text/typescript',
  'jsx': 'text/jsx',
  'tsx': 'text/tsx',
  'html': 'text/html',
  'htm': 'text/html',
  'css': 'text/css',
  'scss': 'text/scss',
  'less': 'text/less',
  'json': 'application/json',
  'xml': 'application/xml',
  'yaml': 'text/yaml',
  'yml': 'text/yaml',
  'php': 'text/x-php',
  'py': 'text/x-python',
  'java': 'text/x-java-source',
  'c': 'text/x-c',
  'cpp': 'text/x-c++',
  'cs': 'text/x-csharp',
  'go': 'text/x-go',
  'rs': 'text/x-rust',
  'rb': 'text/x-ruby',
  'swift': 'text/x-swift',
  'kt': 'text/x-kotlin',
  
  // Fonts
  'ttf': 'font/ttf',
  'otf': 'font/otf',
  'woff': 'font/woff',
  'woff2': 'font/woff2',
  'eot': 'application/vnd.ms-fontobject',
  
  // Data
  'csv': 'text/csv',
  'sql': 'application/sql',
  'db': 'application/x-sqlite3',
  'sqlite': 'application/x-sqlite3',
};

/**
 * Get MIME type from file extension
 */
export function getMimeType(filename: string): string {
  const extension = filename.split('.').pop()?.toLowerCase();
  return extension ? MIME_TYPES[extension] || 'application/octet-stream' : 'application/octet-stream';
}

/**
 * Get file extension from MIME type
 */
export function getExtensionFromMimeType(mimeType: string): string {
  const entry = Object.entries(MIME_TYPES).find(([_, mime]) => mime === mimeType);
  return entry ? entry[0] : '';
}

/**
 * Check if MIME type is supported
 */
export function isMimeTypeSupported(mimeType: string): boolean {
  return Object.values(MIME_TYPES).includes(mimeType);
}

/**
 * Get category from MIME type
 */
export function getCategoryFromMimeType(mimeType: string): string {
  if (mimeType.startsWith('image/')) return 'IMAGES';
  if (mimeType.startsWith('video/')) return 'VIDEOS';
  if (mimeType.startsWith('audio/')) return 'AUDIO';
  if (mimeType.startsWith('text/') || mimeType.includes('document') || mimeType === 'application/pdf') return 'DOCUMENTS';
  if (mimeType.includes('zip') || mimeType.includes('archive') || mimeType.includes('compressed')) return 'ARCHIVES';
  if (mimeType.includes('javascript') || mimeType.includes('json') || mimeType.includes('xml')) return 'CODE';
  return 'OTHER';
}

// lib/permissions.ts
import type { User, TeamMember } from '@/types';
import { USER_ROLES, TEAM_ROLES, TEAM_PERMISSIONS } from '@/lib/constants';

/**
 * Check if user has required role
 */
export function hasRole(user: User | null, requiredRole: string): boolean {
  if (!user) return false;
  
  const roleHierarchy = {
    'user': 0,
    'moderator': 1,
    'admin': 2,
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
  if (user.role === 'admin') return true;

  // Check email verification for sensitive actions
  const sensitiveActions = ['upload', 'share', 'delete', 'admin'];
  if (sensitiveActions.includes(action) && !user.emailVerified) {
    return false;
  }

  // Resource ownership check
  if (resource && resource.owner && resource.owner.toString() === user._id.toString()) {
    return true;
  }

  // Role-based permissions
  const permissions = {
    user: {
      read: true,
      upload: true,
      create: true,
      update: true,
      delete: false,
      share: true,
      admin: false,
    },
    moderator: {
      read: true,
      upload: true,
      create: true,
      update: true,
      delete: true,
      share: true,
      admin: false,
    },
    admin: {
      read: true,
      upload: true,
      create: true,
      update: true,
      delete: true,
      share: true,
      admin: true,
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
 * Get effective permissions for user
 */
export function getEffectivePermissions(user: User | null): Record<string, boolean> {
  if (!user) {
    return {
      read: false,
      upload: false,
      create: false,
      update: false,
      delete: false,
      share: false,
      admin: false,
    };
  }

  const basePermissions = {
    read: hasPermission(user, 'read'),
    upload: hasPermission(user, 'upload'),
    create: hasPermission(user, 'create'),
    update: hasPermission(user, 'update'),
    delete: hasPermission(user, 'delete'),
    share: hasPermission(user, 'share'),
    admin: hasPermission(user, 'admin'),
  };

  return basePermissions;
}

/**
 * Check if user can access resource
 */
export function canAccessResource(
  user: User | null,
  resource: any,
  action: string = 'read'
): boolean {
  if (!user || !resource) return false;

  // Check basic permission
  if (!hasPermission(user, action, resource)) return false;

  // Public resources
  if (resource.visibility === 'public') return true;

  // Owner access
  if (resource.owner && resource.owner.toString() === user._id.toString()) {
    return true;
  }

  // Team access
  if (resource.visibility === 'team' && resource.team) {
    return user.currentTeam?.toString() === resource.team.toString();
  }

  // Private resources - only owner can access
  if (resource.visibility === 'private') {
    return resource.owner && resource.owner.toString() === user._id.toString();
  }

  return false;
}
