// lib/validations/settings.ts
import { z } from 'zod';

const objectIdSchema = z.string().regex(/^[0-9a-fA-F]{24}$/, 'Invalid ObjectId format');

// User Settings (Personal Preferences) - matches types/user.ts UserPreferences
export const userSettingsSchema = z.object({
  preferences: z.object({
    theme: z.enum(['light', 'dark', 'system']).optional(),
    language: z.string().min(2).max(5).optional(),
    timezone: z.string().optional(),
    emailNotifications: z.boolean().optional(),
    pushNotifications: z.boolean().optional(),
    defaultView: z.enum(['grid', 'list']).optional(),
    uploadQuality: z.enum(['original', 'high', 'medium']).optional()
  }).optional()
});

// App Settings - matches types/settings.ts AppSettings
export const appSettingsSchema = z.object({
  name: z.string().min(1, 'App name is required').max(100, 'App name cannot exceed 100 characters').optional(),
  description: z.string().max(500, 'Description cannot exceed 500 characters').optional(),
  logo: z.string().url('Invalid logo URL').optional(),
  favicon: z.string().url('Invalid favicon URL').optional(),
  primaryColor: z.string().regex(/^#[0-9A-F]{6}$/i, 'Invalid primary color format').optional(),
  secondaryColor: z.string().regex(/^#[0-9A-F]{6}$/i, 'Invalid secondary color format').optional(),
  maxFileSize: z.number().min(1024, 'Max file size must be at least 1KB').max(1024 * 1024 * 1024 * 1024 * 10, 'Max file size cannot exceed 10TB').optional(), // 10TB max
  maxFilesPerUpload: z.number().min(1, 'Must allow at least 1 file per upload').max(1000, 'Cannot exceed 1000 files per upload').optional(),
  allowedFileTypes: z.array(z.string().min(1, 'File type cannot be empty')).optional(),
  defaultUserQuota: z.number().min(1024 * 1024, 'Default quota must be at least 1MB').optional(),
  enableRegistration: z.boolean().optional(),
  enableGuestUploads: z.boolean().optional(),
  enablePublicSharing: z.boolean().optional(),
  maintenanceMode: z.boolean().optional(),
  maintenanceMessage: z.string().max(1000, 'Maintenance message cannot exceed 1000 characters').optional()
});

// Storage Settings - matches types/settings.ts StorageSettings
export const storageSettingsSchema = z.object({
  default: z.enum(['local', 's3', 'r2', 'wasabi', 'gcs', 'azure']).optional(),
  providers: z.object({
    local: z.object({
      enabled: z.boolean().optional(),
      path: z.string().min(1, 'Local path is required').optional(),
      maxSize: z.number().min(0, 'Max size cannot be negative').optional()
    }).optional(),
    s3: z.object({
      enabled: z.boolean().optional(),
      accessKeyId: z.string().min(1, 'Access Key ID is required').optional(),
      secretAccessKey: z.string().min(1, 'Secret Access Key is required').optional(),
      region: z.string().min(1, 'Region is required').optional(),
      bucket: z.string().min(1, 'Bucket name is required').optional(),
      endpoint: z.string().url('Invalid S3 endpoint URL').optional(),
      forcePathStyle: z.boolean().optional()
    }).optional(),
    r2: z.object({
      enabled: z.boolean().optional(),
      accessKeyId: z.string().min(1, 'Access Key ID is required').optional(),
      secretAccessKey: z.string().min(1, 'Secret Access Key is required').optional(),
      accountId: z.string().min(1, 'Account ID is required').optional(),
      bucket: z.string().min(1, 'Bucket name is required').optional(),
      endpoint: z.string().url('Invalid R2 endpoint URL').optional()
    }).optional(),
    wasabi: z.object({
      enabled: z.boolean().optional(),
      accessKeyId: z.string().min(1, 'Access Key ID is required').optional(),
      secretAccessKey: z.string().min(1, 'Secret Access Key is required').optional(),
      region: z.string().min(1, 'Region is required').optional(),
      bucket: z.string().min(1, 'Bucket name is required').optional(),
      endpoint: z.string().url('Invalid Wasabi endpoint URL').optional()
    }).optional(),
    gcs: z.object({
      enabled: z.boolean().optional(),
      projectId: z.string().min(1, 'Project ID is required').optional(),
      keyFilename: z.string().min(1, 'Key filename is required').optional(),
      bucket: z.string().min(1, 'Bucket name is required').optional()
    }).optional(),
    azure: z.object({
      enabled: z.boolean().optional(),
      accountName: z.string().min(1, 'Account name is required').optional(),
      accountKey: z.string().min(1, 'Account key is required').optional(),
      containerName: z.string().min(1, 'Container name is required').optional()
    }).optional()
  }).optional()
});

// Email Settings - matches types/settings.ts EmailSettings
export const emailSettingsSchema = z.object({
  enabled: z.boolean().optional(),
  provider: z.enum(['smtp', 'sendgrid', 'ses', 'mailgun', 'resend']).optional(),
  from: z.string().email('Invalid from email address').optional(),
  replyTo: z.string().email('Invalid reply-to email address').optional(),
  smtp: z.object({
    host: z.string().min(1, 'SMTP host is required').optional(),
    port: z.number().min(1, 'Port must be greater than 0').max(65535, 'Port cannot exceed 65535').optional(),
    secure: z.boolean().optional(),
    auth: z.object({
      user: z.string().min(1, 'SMTP username is required').optional(),
      pass: z.string().min(1, 'SMTP password is required').optional()
    }).optional()
  }).optional(),
  sendgrid: z.object({
    apiKey: z.string().min(1, 'SendGrid API key is required').optional()
  }).optional(),
  ses: z.object({
    accessKeyId: z.string().min(1, 'AWS Access Key ID is required').optional(),
    secretAccessKey: z.string().min(1, 'AWS Secret Access Key is required').optional(),
    region: z.string().min(1, 'AWS region is required').optional()
  }).optional(),
  mailgun: z.object({
    apiKey: z.string().min(1, 'Mailgun API key is required').optional(),
    domain: z.string().min(1, 'Mailgun domain is required').optional()
  }).optional(),
  resend: z.object({
    apiKey: z.string().min(1, 'Resend API key is required').optional()
  }).optional(),
  templates: z.object({
    welcome: z.boolean().optional(),
    shareNotification: z.boolean().optional(),
    passwordReset: z.boolean().optional(),
    emailVerification: z.boolean().optional()
  }).optional()
});

// Security Settings - matches types/settings.ts SecuritySettings
export const securitySettingsSchema = z.object({
  enableTwoFactor: z.boolean().optional(),
  sessionTimeout: z.number().min(15, 'Session timeout must be at least 15 minutes').max(43200, 'Session timeout cannot exceed 30 days (43200 minutes)').optional(),
  maxLoginAttempts: z.number().min(3, 'Must allow at least 3 login attempts').max(10, 'Cannot exceed 10 login attempts').optional(),
  lockoutDuration: z.number().min(5, 'Lockout duration must be at least 5 minutes').max(1440, 'Lockout duration cannot exceed 24 hours (1440 minutes)').optional(),
  passwordMinLength: z.number().min(6, 'Password minimum length must be at least 6').max(128, 'Password minimum length cannot exceed 128').optional(),
  passwordRequireNumbers: z.boolean().optional(),
  passwordRequireSymbols: z.boolean().optional(),
  passwordRequireUppercase: z.boolean().optional(),
  passwordRequireLowercase: z.boolean().optional(),
  enableCaptcha: z.boolean().optional(),
  captchaSiteKey: z.string().optional(),
  captchaSecretKey: z.string().optional()
});

// Feature Settings - matches types/settings.ts FeatureSettings
export const featureSettingsSchema = z.object({
  enableSearch: z.boolean().optional(),
  enableOCR: z.boolean().optional(),
  enableThumbnails: z.boolean().optional(),
  enableVersioning: z.boolean().optional(),
  enableTeams: z.boolean().optional(),
  enableAPI: z.boolean().optional(),
  enableWebDAV: z.boolean().optional(),
  enableOfflineSync: z.boolean().optional(),
  maxVersionsPerFile: z.number().min(1, 'Must keep at least 1 version per file').max(100, 'Cannot exceed 100 versions per file').optional()
});

// Analytics Settings - matches types/settings.ts
export const analyticsSettingsSchema = z.object({
  enabled: z.boolean().optional(),
  provider: z.enum(['google', 'mixpanel', 'amplitude']).optional(),
  trackingId: z.string().optional(),
  apiKey: z.string().optional(),
  collectUsageStats: z.boolean().optional(),
  collectErrorLogs: z.boolean().optional()
});

// Backup Settings - matches types/settings.ts
export const backupSettingsSchema = z.object({
  enabled: z.boolean().optional(),
  frequency: z.enum(['daily', 'weekly', 'monthly']).optional(),
  retention: z.number().min(1, 'Retention must be at least 1 day').max(365, 'Retention cannot exceed 365 days').optional(),
  destination: z.enum(['local', 's3', 'gcs']).optional(),
  encryption: z.boolean().optional(),
  encryptionKey: z.string().optional()
});

// Notification Settings - matches types/settings.ts
export const notificationSettingsSchema = z.object({
  enablePush: z.boolean().optional(),
  pushProvider: z.enum(['firebase', 'pusher', 'websocket']).optional(),
  firebaseServerKey: z.string().optional(),
  pusherAppId: z.string().optional(),
  pusherKey: z.string().optional(),
  pusherSecret: z.string().optional(),
  pusherCluster: z.string().optional()
});

// Rate Limiting Settings - matches types/settings.ts RateLimitSettings
export const rateLimitingSettingsSchema = z.object({
  enabled: z.boolean().optional(),
  uploadLimit: z.number().min(1, 'Upload limit must be at least 1 request per minute').max(1000, 'Upload limit cannot exceed 1000 requests per minute').optional(),
  downloadLimit: z.number().min(1, 'Download limit must be at least 1 request per minute').max(10000, 'Download limit cannot exceed 10000 requests per minute').optional(),
  apiLimit: z.number().min(1, 'API limit must be at least 1 request per minute').max(10000, 'API limit cannot exceed 10000 requests per minute').optional(),
  shareLimit: z.number().min(1, 'Share limit must be at least 1 request per minute').max(1000, 'Share limit cannot exceed 1000 requests per minute').optional()
});

// System Settings Update - matches admin store updateSettings and types/settings.ts UpdateSettingsRequest
export const updateSystemSettingsSchema = z.object({
  app: appSettingsSchema.optional(),
  storage: storageSettingsSchema.optional(),
  email: emailSettingsSchema.optional(),
  security: securitySettingsSchema.optional(),
  features: featureSettingsSchema.optional(),
  analytics: analyticsSettingsSchema.optional(),
  backup: backupSettingsSchema.optional(),
  notifications: notificationSettingsSchema.optional(),
  rateLimiting: rateLimitingSettingsSchema.optional()
});

// Test Settings - for admin testing configurations
export const testEmailSettingsSchema = z.object({
  recipientEmail: z.string().email('Invalid recipient email'),
  testType: z.enum(['connection', 'template']).default('connection')
});

export const testStorageSettingsSchema = z.object({
  provider: z.enum(['local', 's3', 'r2', 'wasabi', 'gcs', 'azure']),
  testType: z.enum(['connection', 'upload', 'download']).default('connection')
});

// Settings Import/Export
export const exportSettingsSchema = z.object({
  sections: z.array(z.enum(['app', 'storage', 'email', 'security', 'features', 'analytics', 'backup', 'notifications', 'rateLimiting'])).min(1, 'At least one section must be selected'),
  format: z.enum(['json', 'yaml']).default('json'),
  includeSecrets: z.boolean().default(false)
});

export const importSettingsSchema = z.object({
  settings: z.record(z.any()).refine((data) => {
    // Basic validation that it's a valid settings object
    return typeof data === 'object' && data !== null;
  }, 'Invalid settings data'),
  overwrite: z.boolean().default(false),
  backup: z.boolean().default(true)
});

// Type exports matching your types/settings.ts
export type UserSettingsRequest = z.infer<typeof userSettingsSchema>;
export type UpdateSettingsRequest = z.infer<typeof updateSystemSettingsSchema>;
export type AppSettingsRequest = z.infer<typeof appSettingsSchema>;
export type StorageSettingsRequest = z.infer<typeof storageSettingsSchema>;
export type EmailSettingsRequest = z.infer<typeof emailSettingsSchema>;
export type SecuritySettingsRequest = z.infer<typeof securitySettingsSchema>;
export type FeatureSettingsRequest = z.infer<typeof featureSettingsSchema>;
export type AnalyticsSettingsRequest = z.infer<typeof analyticsSettingsSchema>;
export type BackupSettingsRequest = z.infer<typeof backupSettingsSchema>;
export type NotificationSettingsRequest = z.infer<typeof notificationSettingsSchema>;
export type RateLimitingSettingsRequest = z.infer<typeof rateLimitingSettingsSchema>;
export type TestEmailSettingsRequest = z.infer<typeof testEmailSettingsSchema>;
export type TestStorageSettingsRequest = z.infer<typeof testStorageSettingsSchema>;
export type ExportSettingsRequest = z.infer<typeof exportSettingsSchema>;
export type ImportSettingsRequest = z.infer<typeof importSettingsSchema>;