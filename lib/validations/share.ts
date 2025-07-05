// lib/validations/share.ts
import { z } from 'zod';

const objectIdSchema = z.string().regex(/^[0-9a-fA-F]{24}$/, 'Invalid ObjectId format');

// Create Share Request - matches models/Share.ts exactly
export const createShareSchema = z.object({
  resourceId: objectIdSchema,
  resourceType: z.enum(['file', 'folder']), // From models/Share.ts
  type: z.enum(['public', 'restricted', 'domain']), // From models/Share.ts
  permission: z.enum(['view', 'comment', 'edit']), // From models/Share.ts
  
  // Access control settings
  allowDownload: z.boolean().default(true),
  allowPrint: z.boolean().default(true),
  allowCopy: z.boolean().default(true),
  requireAuth: z.boolean().default(false),
  
  // Password protection
  password: z.string()
    .min(4, 'Password must be at least 4 characters')
    .max(50, 'Password cannot exceed 50 characters')
    .optional(),
  
  // Expiration
  expiresAt: z.string().datetime().optional(),
  
  // Domain restriction (for type: 'domain')
  allowedDomains: z.array(z.string().min(1, 'Domain cannot be empty')).optional(),
  
  // Specific users (for type: 'restricted')
  allowedUsers: z.array(z.object({
    email: z.string().email('Invalid email address'),
    permission: z.enum(['view', 'comment', 'edit']),
    userId: objectIdSchema.optional()
  })).optional()
}).refine((data) => {
  // Validate domain shares require domains
  if (data.type === 'domain' && (!data.allowedDomains || data.allowedDomains.length === 0)) {
    return false;
  }
  return true;
}, {
  message: 'Domain shares must specify at least one allowed domain',
  path: ['allowedDomains']
}).refine((data) => {
  // Validate restricted shares require users
  if (data.type === 'restricted' && (!data.allowedUsers || data.allowedUsers.length === 0)) {
    return false;
  }
  return true;
}, {
  message: 'Restricted shares must specify at least one allowed user',
  path: ['allowedUsers']
});

// Update Share Request - for modifying existing shares
export const updateShareSchema = z.object({
  permission: z.enum(['view', 'comment', 'edit']).optional(),
  allowDownload: z.boolean().optional(),
  allowPrint: z.boolean().optional(),
  allowCopy: z.boolean().optional(),
  requireAuth: z.boolean().optional(),
  password: z.string().min(4).max(50).optional(),
  expiresAt: z.string().datetime().optional(),
  isActive: z.boolean().optional(),
  allowedDomains: z.array(z.string().min(1)).optional(),
  allowedUsers: z.array(z.object({
    email: z.string().email(),
    permission: z.enum(['view', 'comment', 'edit']),
    userId: objectIdSchema.optional()
  })).optional()
});

// Share Access Request - for accessing shared content
export const shareAccessSchema = z.object({
  shareToken: z.string().min(1, 'Share token is required'),
  password: z.string().optional(),
  email: z.string().email('Invalid email address').optional()
});

// Share Filters - for listing shares
export const shareFiltersSchema = z.object({
  resourceType: z.enum(['file', 'folder']).optional(),
  type: z.enum(['public', 'restricted', 'domain']).optional(),
  permission: z.enum(['view', 'comment', 'edit']).optional(),
  isActive: z.boolean().optional(),
  isExpired: z.boolean().optional(),
  dateRange: z.object({
    start: z.string().datetime().optional(),
    end: z.string().datetime().optional()
  }).optional(),
  page: z.number().min(1).default(1),
  limit: z.number().min(1).max(100).default(20),
  sort: z.enum(['createdAt', 'accessCount', 'lastAccessedAt', 'expiresAt']).default('createdAt'),
  order: z.enum(['asc', 'desc']).default('desc')
});

// Bulk Share Operations
export const bulkShareOperationSchema = z.object({
  shareIds: z.array(objectIdSchema).min(1, 'At least one share must be selected').max(100, 'Cannot perform bulk operations on more than 100 shares'),
  action: z.enum(['activate', 'deactivate', 'revoke', 'delete', 'extend']),
  expiresAt: z.string().datetime().optional() // For extend action
}).refine((data) => {
  if (data.action === 'extend' && !data.expiresAt) {
    return false;
  }
  return true;
}, {
  message: 'Expiration date is required for extend action',
  path: ['expiresAt']
});

// Share Analytics Request
export const shareAnalyticsSchema = z.object({
  shareId: objectIdSchema.optional(),
  dateRange: z.object({
    start: z.string().datetime(),
    end: z.string().datetime()
  }).optional(),
  groupBy: z.enum(['hour', 'day', 'week', 'month']).default('day')
});

// Share Notification Settings
export const shareNotificationSchema = z.object({
  shareId: objectIdSchema,
  notifyOnAccess: z.boolean(),
  notifyOnDownload: z.boolean().optional(),
  notifyOnComment: z.boolean().optional()
});

// Type exports matching your project structure
export type CreateShareRequest = z.infer<typeof createShareSchema>;
export type UpdateShareRequest = z.infer<typeof updateShareSchema>;
export type ShareAccessRequest = z.infer<typeof shareAccessSchema>;
export type ShareFiltersRequest = z.infer<typeof shareFiltersSchema>;
export type BulkShareOperationRequest = z.infer<typeof bulkShareOperationSchema>;
export type ShareAnalyticsRequest = z.infer<typeof shareAnalyticsSchema>;
export type ShareNotificationRequest = z.infer<typeof shareNotificationSchema>;