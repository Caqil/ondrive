// lib/validations/admin.ts
import { z } from 'zod';

// Base ObjectId validation
const objectIdSchema = z.string().regex(/^[0-9a-fA-F]{24}$/, 'Invalid ObjectId format');

// Admin Action Validation - matches AdminActionRequest from types/admin.ts
export const adminActionSchema = z.object({
  userId: objectIdSchema,
  action: z.enum(['ban', 'unban', 'verify_email', 'reset_password', 'update_quota', 'impersonate']),
  reason: z.string().min(3, 'Reason must be at least 3 characters').max(500, 'Reason cannot exceed 500 characters').optional(),
  data: z.record(z.any()).optional()
});

// Update User Quota - specific admin action
export const updateUserQuotaSchema = z.object({
  storageQuota: z.number()
    .min(1024 * 1024, 'Storage quota must be at least 1MB') // 1MB minimum
    .max(1024 * 1024 * 1024 * 1024 * 10, 'Storage quota cannot exceed 10TB') // 10TB maximum
});

// User Filters for admin user management - matches admin store loadUsers
export const adminUserFiltersSchema = z.object({
  query: z.string().optional(),
  role: z.enum(['viewer', 'user', 'moderator', 'admin']).optional(), // Matches types/user.ts
  status: z.enum(['active', 'inactive', 'banned']).optional(),
  subscriptionStatus: z.enum(['active', 'inactive', 'trial', 'expired', 'cancelled']).optional(),
  emailVerified: z.boolean().optional(),
  dateRange: z.object({
    start: z.string().datetime().optional(),
    end: z.string().datetime().optional()
  }).optional(),
  storageUsage: z.object({
    min: z.number().min(0).optional(),
    max: z.number().min(0).optional()
  }).optional(),
  page: z.number().min(1).default(1),
  limit: z.number().min(1).max(100).default(20),
  sort: z.enum(['name', 'email', 'createdAt', 'lastLogin', 'storageUsed']).default('createdAt'),
  order: z.enum(['asc', 'desc']).default('desc')
});

// Settings Update - matches types/settings.ts structure
export const updateSettingsSchema = z.object({
  app: z.object({
    name: z.string().min(1).max(100).optional(),
    description: z.string().max(500).optional(),
    logo: z.string().url().optional(),
    favicon: z.string().url().optional(),
    primaryColor: z.string().regex(/^#[0-9A-F]{6}$/i, 'Invalid color format').optional(),
    secondaryColor: z.string().regex(/^#[0-9A-F]{6}$/i, 'Invalid color format').optional(),
    maxFileSize: z.number().min(1024, 'Must be at least 1KB').optional(),
    maxFilesPerUpload: z.number().min(1).max(1000).optional(),
    allowedFileTypes: z.array(z.string()).optional(),
    defaultUserQuota: z.number().min(1024 * 1024, 'Must be at least 1MB').optional(),
    enableRegistration: z.boolean().optional(),
    enableGuestUploads: z.boolean().optional(),
    enablePublicSharing: z.boolean().optional(),
    maintenanceMode: z.boolean().optional(),
    maintenanceMessage: z.string().max(1000).optional()
  }).optional(),
  
  storage: z.object({
    default: z.enum(['local', 's3', 'r2', 'wasabi', 'gcs', 'azure']).optional(),
    providers: z.object({
      local: z.object({
        enabled: z.boolean().optional(),
        path: z.string().optional(),
        maxSize: z.number().min(0).optional()
      }).optional(),
      s3: z.object({
        enabled: z.boolean().optional(),
        accessKeyId: z.string().optional(),
        secretAccessKey: z.string().optional(),
        region: z.string().optional(),
        bucket: z.string().optional(),
        endpoint: z.string().url().optional(),
        forcePathStyle: z.boolean().optional()
      }).optional(),
      r2: z.object({
        enabled: z.boolean().optional(),
        accessKeyId: z.string().optional(),
        secretAccessKey: z.string().optional(),
        accountId: z.string().optional(),
        bucket: z.string().optional(),
        endpoint: z.string().url().optional()
      }).optional(),
      wasabi: z.object({
        enabled: z.boolean().optional(),
        accessKeyId: z.string().optional(),
        secretAccessKey: z.string().optional(),
        region: z.string().optional(),
        bucket: z.string().optional(),
        endpoint: z.string().url().optional()
      }).optional(),
      gcs: z.object({
        enabled: z.boolean().optional(),
        projectId: z.string().optional(),
        keyFilename: z.string().optional(),
        bucket: z.string().optional()
      }).optional(),
      azure: z.object({
        enabled: z.boolean().optional(),
        accountName: z.string().optional(),
        accountKey: z.string().optional(),
        containerName: z.string().optional()
      }).optional()
    }).optional()
  }).optional(),
  
  email: z.object({
    enabled: z.boolean().optional(),
    provider: z.enum(['smtp', 'sendgrid', 'ses', 'mailgun', 'resend']).optional(),
    from: z.string().email().optional(),
    replyTo: z.string().email().optional(),
    smtp: z.object({
      host: z.string().optional(),
      port: z.number().min(1).max(65535).optional(),
      secure: z.boolean().optional(),
      auth: z.object({
        user: z.string().optional(),
        pass: z.string().optional()
      }).optional()
    }).optional(),
    sendgrid: z.object({
      apiKey: z.string().optional()
    }).optional(),
    ses: z.object({
      accessKeyId: z.string().optional(),
      secretAccessKey: z.string().optional(),
      region: z.string().optional()
    }).optional(),
    mailgun: z.object({
      apiKey: z.string().optional(),
      domain: z.string().optional()
    }).optional(),
    resend: z.object({
      apiKey: z.string().optional()
    }).optional(),
    templates: z.object({
      welcome: z.boolean().optional(),
      shareNotification: z.boolean().optional(),
      passwordReset: z.boolean().optional(),
      emailVerification: z.boolean().optional()
    }).optional()
  }).optional(),
  
  security: z.object({
    enableTwoFactor: z.boolean().optional(),
    sessionTimeout: z.number().min(15).optional(), // minimum 15 minutes
    maxLoginAttempts: z.number().min(3).max(10).optional(),
    lockoutDuration: z.number().min(5).optional(),
    passwordMinLength: z.number().min(6).max(128).optional(),
    passwordRequireNumbers: z.boolean().optional(),
    passwordRequireSymbols: z.boolean().optional(),
    passwordRequireUppercase: z.boolean().optional(),
    passwordRequireLowercase: z.boolean().optional(),
    enableCaptcha: z.boolean().optional(),
    captchaSiteKey: z.string().optional(),
    captchaSecretKey: z.string().optional()
  }).optional(),
  
  features: z.object({
    enableSearch: z.boolean().optional(),
    enableOCR: z.boolean().optional(),
    enableThumbnails: z.boolean().optional(),
    enableVersioning: z.boolean().optional(),
    enableTeams: z.boolean().optional(),
    enableAPI: z.boolean().optional(),
    enableWebDAV: z.boolean().optional(),
    enableOfflineSync: z.boolean().optional(),
    maxVersionsPerFile: z.number().min(1).max(100).optional()
  }).optional(),
  
  rateLimiting: z.object({
    enabled: z.boolean().optional(),
    uploadLimit: z.number().min(1).optional(),
    downloadLimit: z.number().min(1).optional(),
    apiLimit: z.number().min(1).optional(),
    shareLimit: z.number().min(1).optional()
  }).optional()
});

// Analytics Date Range - for dashboard stats
export const analyticsDateRangeSchema = z.object({
  start: z.string().datetime(),
  end: z.string().datetime(),
  metrics: z.array(z.enum([
    'users', 'files', 'storage', 'revenue', 'subscriptions', 'activity'
  ])).min(1, 'At least one metric is required'),
  groupBy: z.enum(['hour', 'day', 'week', 'month', 'year']).default('day')
}).refine((data) => new Date(data.start) < new Date(data.end), {
  message: 'Start date must be before end date',
  path: ['end']
});

// Bulk Operations
export const bulkUserOperationSchema = z.object({
  userIds: z.array(objectIdSchema).min(1, 'At least one user must be selected').max(100, 'Cannot perform bulk operations on more than 100 users'),
  action: z.enum(['ban', 'unban', 'verify_email', 'delete', 'export']),
  reason: z.string().min(5).max(500).optional()
});

// Export Configuration
export const exportUsersSchema = z.object({
  format: z.enum(['csv', 'json', 'xlsx']).default('csv'),
  fields: z.array(z.enum([
    '_id', 'email', 'name', 'role', 'emailVerified', 'isActive', 'isBanned',
    'storageUsed', 'storageQuota', 'subscriptionStatus', 'createdAt', 'lastLogin'
  ])).min(1, 'At least one field must be selected'),
  filters: adminUserFiltersSchema.omit({ page: true, limit: true }).optional()
});

// System Health Check
export const healthCheckSchema = z.object({
  service: z.enum(['database', 'storage', 'email', 'payment', 'all']).default('all')
});

// Type exports matching your types/admin.ts
export type AdminActionRequest = z.infer<typeof adminActionSchema>;
export type UpdateUserQuotaRequest = z.infer<typeof updateUserQuotaSchema>;
export type AdminUserFilters = z.infer<typeof adminUserFiltersSchema>;
export type UpdateSettingsRequest = z.infer<typeof updateSettingsSchema>;
export type AnalyticsDateRangeRequest = z.infer<typeof analyticsDateRangeSchema>;
export type BulkUserOperationRequest = z.infer<typeof bulkUserOperationSchema>;
export type ExportUsersRequest = z.infer<typeof exportUsersSchema>;
export type HealthCheckRequest = z.infer<typeof healthCheckSchema>;