/**
 * Application constants and configuration
 */

// File size limits
export const FILE_SIZE_LIMITS = {
  FREE_USER: 100 * 1024 * 1024, // 100MB
  BASIC_USER: 500 * 1024 * 1024, // 500MB
  PRO_USER: 5 * 1024 * 1024 * 1024, // 5GB
  ENTERPRISE_USER: 750 * 1024 * 1024 * 1024, // 750GB
  MAX_CHUNK_SIZE: 10 * 1024 * 1024, // 10MB
  MAX_UPLOAD_SIZE: 750 * 1024 * 1024 * 1024, // 750GB
} as const;

// Storage quotas
export const STORAGE_QUOTAS = {
  FREE: 15 * 1024 * 1024 * 1024, // 15GB
  TRIAL: 50 * 1024 * 1024 * 1024, // 50GB
  BASIC: 100 * 1024 * 1024 * 1024, // 100GB
  PRO: 1024 * 1024 * 1024 * 1024, // 1TB
  ENTERPRISE: Number.MAX_SAFE_INTEGER, // Unlimited
} as const;

// User roles
export const USER_ROLES = {
  VIEWER: 'viewer',
  USER: 'user',
  MODERATOR: 'moderator',
  ADMIN: 'admin',
} as const;

// Team roles
export const TEAM_ROLES = {
  OWNER: 'owner',
  ADMIN: 'admin',
  EDITOR: 'editor',
  MEMBER: 'member',
  VIEWER: 'viewer',
} as const;

// Team permissions by role
export const TEAM_PERMISSIONS = {
  owner: {
    canUpload: true,
    canDownload: true,
    canShare: true,
    canDelete: true,
    canInvite: true,
    canManageTeam: true,
    canViewBilling: true,
    canManageBilling: true,
  },
  admin: {
    canUpload: true,
    canDownload: true,
    canShare: true,
    canDelete: true,
    canInvite: true,
    canManageTeam: true,
    canViewBilling: true,
    canManageBilling: false,
  },
  editor: {
    canUpload: true,
    canDownload: true,
    canShare: true,
    canDelete: true,
    canInvite: false,
    canManageTeam: false,
    canViewBilling: false,
    canManageBilling: false,
  },
  member: {
    canUpload: true,
    canDownload: true,
    canShare: true,
    canDelete: false,
    canInvite: false,
    canManageTeam: false,
    canViewBilling: false,
    canManageBilling: false,
  },
  viewer: {
    canUpload: false,
    canDownload: true,
    canShare: false,
    canDelete: false,
    canInvite: false,
    canManageTeam: false,
    canViewBilling: false,
    canManageBilling: false,
  },
} as const;

// File categories
export const FILE_CATEGORIES = {
  IMAGES: ['image/jpeg', 'image/png', 'image/gif', 'image/webp', 'image/svg+xml', 'image/bmp', 'image/tiff'],
  VIDEOS: ['video/mp4', 'video/avi', 'video/mov', 'video/wmv', 'video/flv', 'video/webm', 'video/mkv'],
  AUDIO: ['audio/mpeg', 'audio/wav', 'audio/ogg', 'audio/aac', 'audio/flac', 'audio/mp4'],
  DOCUMENTS: [
    'application/pdf',
    'application/msword',
    'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
    'application/vnd.ms-excel',
    'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
    'application/vnd.ms-powerpoint',
    'application/vnd.openxmlformats-officedocument.presentationml.presentation',
    'text/plain',
    'text/csv',
    'application/rtf',
  ],
  ARCHIVES: [
    'application/zip',
    'application/x-rar-compressed',
    'application/x-7z-compressed',
    'application/x-tar',
    'application/gzip',
  ],
  CODE: [
    'text/javascript',
    'text/typescript',
    'text/html',
    'text/css',
    'application/json',
    'application/xml',
    'text/x-python',
    'text/x-java-source',
  ],
} as const;

// Rate limiting
export const RATE_LIMITS = {
  AUTH: {
    WINDOW_MS: 15 * 60 * 1000, // 15 minutes
    MAX_REQUESTS: 5, // 5 attempts per window
  },
  API: {
    WINDOW_MS: 60 * 1000, // 1 minute
    MAX_REQUESTS: 100, // 100 requests per minute
  },
  UPLOAD: {
    WINDOW_MS: 60 * 1000, // 1 minute
    MAX_REQUESTS: 10, // 10 uploads per minute
  },
  DOWNLOAD: {
    WINDOW_MS: 60 * 1000, // 1 minute
    MAX_REQUESTS: 60, // 60 downloads per minute
  },
  SHARE: {
    WINDOW_MS: 60 * 1000, // 1 minute
    MAX_REQUESTS: 20, // 20 shares per minute
  },
} as const;

// Subscription plans
export const SUBSCRIPTION_PLANS = {
  FREE: 'free',
  BASIC: 'basic',
  PRO: 'pro',
  ENTERPRISE: 'enterprise',
} as const;

// Plan features
export const PLAN_FEATURES = {
  [SUBSCRIPTION_PLANS.FREE]: {
    storageLimit: STORAGE_QUOTAS.FREE,
    memberLimit: 1,
    fileUploadLimit: FILE_SIZE_LIMITS.FREE_USER,
    apiRequestLimit: 1000,
    enableAdvancedSharing: false,
    enableVersionHistory: false,
    enableOCR: false,
    enablePrioritySupport: false,
    enableAPIAccess: false,
    enableIntegrations: false,
    enableCustomBranding: false,
    enableAuditLogs: false,
    enableSSO: false,
  },
  [SUBSCRIPTION_PLANS.BASIC]: {
    storageLimit: STORAGE_QUOTAS.BASIC,
    memberLimit: 5,
    fileUploadLimit: FILE_SIZE_LIMITS.BASIC_USER,
    apiRequestLimit: 10000,
    enableAdvancedSharing: true,
    enableVersionHistory: true,
    enableOCR: false,
    enablePrioritySupport: false,
    enableAPIAccess: false,
    enableIntegrations: false,
    enableCustomBranding: false,
    enableAuditLogs: false,
    enableSSO: false,
  },
  [SUBSCRIPTION_PLANS.PRO]: {
    storageLimit: STORAGE_QUOTAS.PRO,
    memberLimit: 25,
    fileUploadLimit: FILE_SIZE_LIMITS.PRO_USER,
    apiRequestLimit: 100000,
    enableAdvancedSharing: true,
    enableVersionHistory: true,
    enableOCR: true,
    enablePrioritySupport: true,
    enableAPIAccess: true,
    enableIntegrations: true,
    enableCustomBranding: false,
    enableAuditLogs: true,
    enableSSO: false,
  },
  [SUBSCRIPTION_PLANS.ENTERPRISE]: {
    storageLimit: STORAGE_QUOTAS.ENTERPRISE,
    memberLimit: Number.MAX_SAFE_INTEGER,
    fileUploadLimit: FILE_SIZE_LIMITS.ENTERPRISE_USER,
    apiRequestLimit: Number.MAX_SAFE_INTEGER,
    enableAdvancedSharing: true,
    enableVersionHistory: true,
    enableOCR: true,
    enablePrioritySupport: true,
    enableAPIAccess: true,
    enableIntegrations: true,
    enableCustomBranding: true,
    enableAuditLogs: true,
    enableSSO: true,
  },
} as const;

// Subscription status
export const SUBSCRIPTION_STATUS = {
  TRIAL: 'trial',
  ACTIVE: 'active',
  PAST_DUE: 'past_due',
  CANCELLED: 'cancelled',
  EXPIRED: 'expired',
} as const;

// Payment status
export const PAYMENT_STATUS = {
  PENDING: 'pending',
  PROCESSING: 'processing',
  SUCCEEDED: 'succeeded',
  FAILED: 'failed',
  CANCELLED: 'cancelled',
  REFUNDED: 'refunded',
  PARTIALLY_REFUNDED: 'partially_refunded',
  DISPUTED: 'disputed',
} as const;

// Invoice status
export const INVOICE_STATUS = {
  DRAFT: 'draft',
  OPEN: 'open',
  PAID: 'paid',
  VOID: 'void',
  UNCOLLECTIBLE: 'uncollectible',
  OVERDUE: 'overdue',
} as const;

// File processing status
export const PROCESSING_STATUS = {
  PENDING: 'pending',
  PROCESSING: 'processing',
  COMPLETED: 'completed',
  FAILED: 'failed',
} as const;

// Sync status
export const SYNC_STATUS = {
  SYNCED: 'synced',
  PENDING: 'pending',
  CONFLICT: 'conflict',
  ERROR: 'error',
} as const;

// Share types
export const SHARE_TYPES = {
  PUBLIC: 'public',
  RESTRICTED: 'restricted',
  DOMAIN: 'domain',
} as const;

// Share permissions
export const SHARE_PERMISSIONS = {
  VIEW: 'view',
  COMMENT: 'comment',
  EDIT: 'edit',
} as const;

// Notification types
export const NOTIFICATION_TYPES = {
  SHARE_RECEIVED: 'share_received',
  SHARE_ACCEPTED: 'share_accepted',
  FILE_UPLOADED: 'file_uploaded',
  FILE_DELETED: 'file_deleted',
  TEAM_INVITE: 'team_invite',
  PAYMENT_SUCCESS: 'payment_success',
  PAYMENT_FAILED: 'payment_failed',
  STORAGE_LIMIT: 'storage_limit',
  TRIAL_ENDING: 'trial_ending',
  SUBSCRIPTION_CANCELLED: 'subscription_cancelled',
  SECURITY_ALERT: 'security_alert',
} as const;

// Notification priorities
export const NOTIFICATION_PRIORITIES = {
  LOW: 'low',
  NORMAL: 'normal',
  HIGH: 'high',
  URGENT: 'urgent',
} as const;

// Activity actions
export const ACTIVITY_ACTIONS = {
  CREATE: 'create',
  READ: 'read',
  UPDATE: 'update',
  DELETE: 'delete',
  SHARE: 'share',
  DOWNLOAD: 'download',
  UPLOAD: 'upload',
  MOVE: 'move',
  COPY: 'copy',
  RENAME: 'rename',
  STAR: 'star',
  UNSTAR: 'unstar',
  TRASH: 'trash',
  RESTORE: 'restore',
} as const;

// Resource types
export const RESOURCE_TYPES = {
  FILE: 'file',
  FOLDER: 'folder',
  SHARE: 'share',
  USER: 'user',
  TEAM: 'team',
  SETTINGS: 'settings',
  SUBSCRIPTION: 'subscription',
  PAYMENT: 'payment',
} as const;

// Activity categories
export const ACTIVITY_CATEGORIES = {
  AUTH: 'auth',
  FILE: 'file',
  ADMIN: 'admin',
  BILLING: 'billing',
  TEAM: 'team',
  SECURITY: 'security',
} as const;

// Storage providers
export const STORAGE_PROVIDERS = {
  LOCAL: 'local',
  S3: 's3',
  R2: 'r2',
  WASABI: 'wasabi',
  GCS: 'gcs',
  AZURE: 'azure',
} as const;

// Payment providers
export const PAYMENT_PROVIDERS = {
  STRIPE: 'stripe',
  PAYPAL: 'paypal',
  PADDLE: 'paddle',
  LEMONSQUEEZY: 'lemonsqueezy',
  RAZORPAY: 'razorpay',
  MANUAL: 'manual',
} as const;

// Email providers
export const EMAIL_PROVIDERS = {
  SMTP: 'smtp',
  SENDGRID: 'sendgrid',
  SES: 'ses',
  MAILGUN: 'mailgun',
  RESEND: 'resend',
} as const;

// Currencies
export const CURRENCIES = {
  USD: 'USD',
  EUR: 'EUR',
  GBP: 'GBP',
  CAD: 'CAD',
  AUD: 'AUD',
  JPY: 'JPY',
  CHF: 'CHF',
  CNY: 'CNY',
} as const;

// Languages
export const LANGUAGES = {
  EN: 'en',
  ES: 'es',
  FR: 'fr',
  DE: 'de',
  IT: 'it',
  PT: 'pt',
  RU: 'ru',
  ZH: 'zh',
  JA: 'ja',
  KO: 'ko',
} as const;

// Themes
export const THEMES = {
  LIGHT: 'light',
  DARK: 'dark',
  SYSTEM: 'system',
} as const;

// View modes
export const VIEW_MODES = {
  GRID: 'grid',
  LIST: 'list',
} as const;

// Sort options
export const SORT_OPTIONS = {
  NAME: 'name',
  SIZE: 'size',
  MODIFIED: 'modified',
  TYPE: 'type',
  CREATED: 'created',
} as const;

// Sort orders
export const SORT_ORDERS = {
  ASC: 'asc',
  DESC: 'desc',
} as const;

// Upload qualities
export const UPLOAD_QUALITIES = {
  ORIGINAL: 'original',
  HIGH: 'high',
  MEDIUM: 'medium',
} as const;

// Visibility options
export const VISIBILITY_OPTIONS = {
  PRIVATE: 'private',
  TEAM: 'team',
  PUBLIC: 'public',
} as const;

// Member statuses
export const MEMBER_STATUSES = {
  ACTIVE: 'active',
  PENDING: 'pending',
  SUSPENDED: 'suspended',
  REMOVED: 'removed',
} as const;

// Invitation statuses
export const INVITATION_STATUSES = {
  PENDING: 'pending',
  ACCEPTED: 'accepted',
  DECLINED: 'declined',
  EXPIRED: 'expired',
  REVOKED: 'revoked',
} as const;

// Health statuses
export const HEALTH_STATUSES = {
  HEALTHY: 'healthy',
  WARNING: 'warning',
  CRITICAL: 'critical',
  UNKNOWN: 'unknown',
} as const;

// System status
export const SYSTEM_STATUS = {
  HEALTHY: 'healthy',
  WARNING: 'warning',
  CRITICAL: 'critical',
} as const;

// API response codes
export const API_RESPONSE_CODES = {
  SUCCESS: 'SUCCESS',
  VALIDATION_ERROR: 'VALIDATION_ERROR',
  AUTHENTICATION_ERROR: 'AUTHENTICATION_ERROR',
  AUTHORIZATION_ERROR: 'AUTHORIZATION_ERROR',
  NOT_FOUND: 'NOT_FOUND',
  CONFLICT: 'CONFLICT',
  RATE_LIMITED: 'RATE_LIMITED',
  SERVER_ERROR: 'SERVER_ERROR',
} as const;

// Error messages
export const ERROR_MESSAGES = {
  UNAUTHORIZED: 'You are not authorized to perform this action',
  FORBIDDEN: 'Access denied',
  NOT_FOUND: 'Resource not found',
  VALIDATION_FAILED: 'Validation failed',
  RATE_LIMITED: 'Too many requests. Please try again later',
  SERVER_ERROR: 'Internal server error',
  FILE_TOO_LARGE: 'File size exceeds the maximum allowed limit',
  INVALID_FILE_TYPE: 'File type is not supported',
  STORAGE_QUOTA_EXCEEDED: 'Storage quota exceeded',
  SUBSCRIPTION_REQUIRED: 'This feature requires an active subscription',
  TEAM_LIMIT_REACHED: 'Team member limit reached',
} as const;

// Success messages
export const SUCCESS_MESSAGES = {
  FILE_UPLOADED: 'File uploaded successfully',
  FOLDER_CREATED: 'Folder created successfully',
  ITEM_DELETED: 'Item deleted successfully',
  ITEM_MOVED: 'Item moved successfully',
  ITEM_COPIED: 'Item copied successfully',
  ITEM_RENAMED: 'Item renamed successfully',
  SETTINGS_UPDATED: 'Settings updated successfully',
  PROFILE_UPDATED: 'Profile updated successfully',
  PASSWORD_CHANGED: 'Password changed successfully',
  EMAIL_VERIFIED: 'Email verified successfully',
  SUBSCRIPTION_CREATED: 'Subscription created successfully',
  PAYMENT_SUCCESSFUL: 'Payment processed successfully',
} as const;

// Regex patterns
export const REGEX_PATTERNS = {
  EMAIL: /^[^\s@]+@[^\s@]+\.[^\s@]+$/,
  PASSWORD: /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]{8,}$/,
  PHONE: /^\+?[\d\s-()]+$/,
  URL: /^https?:\/\/.+/,
  HEX_COLOR: /^#[0-9A-F]{6}$/i,
  SLUG: /^[a-z0-9-]+$/,
  API_KEY: /^[a-zA-Z0-9_-]+$/,
  IP_ADDRESS: /^(\d{1,3}\.){3}\d{1,3}$/,
  FILENAME: /^[^<>:"/\\|?*]+$/,
} as const;

// Date formats
export const DATE_FORMATS = {
  SHORT: 'MMM d, yyyy',
  LONG: 'MMMM d, yyyy',
  WITH_TIME: 'MMM d, yyyy h:mm a',
  ISO: "yyyy-MM-dd'T'HH:mm:ss.SSS'Z'",
  DATE_ONLY: 'yyyy-MM-dd',
  TIME_ONLY: 'HH:mm:ss',
} as const;

// Time periods
export const TIME_PERIODS = {
  DAILY: 'daily',
  WEEKLY: 'weekly',
  MONTHLY: 'monthly',
  YEARLY: 'yearly',
} as const;

// Default values
export const DEFAULTS = {
  PAGE_SIZE: 20,
  MAX_PAGE_SIZE: 100,
  SESSION_TIMEOUT: 24 * 60 * 60 * 1000, // 24 hours
  TOKEN_EXPIRY: 15 * 60 * 1000, // 15 minutes
  TRIAL_DAYS: 14,
  MAX_FOLDER_DEPTH: 20,
  MAX_FILENAME_LENGTH: 255,
  MAX_DESCRIPTION_LENGTH: 1000,
  THUMBNAIL_SIZE: 200,
  PREVIEW_SIZE: 800,
} as const;