import { z } from 'zod';

const objectIdSchema = z.string().regex(/^[0-9a-fA-F]{24}$/, 'Invalid ObjectId format');
const emailSchema = z.string().email('Invalid email address').toLowerCase();
const nameSchema = z.string().min(1, 'Name is required').max(100, 'Name cannot exceed 100 characters');

// OAuth Provider Data - matches models/User.ts providers
const oauthProvidersSchema = z.object({
  google: z.object({
    id: z.string().min(1, 'Google ID is required'),
    email: z.string().email('Invalid Google email')
  }).optional(),
  github: z.object({
    id: z.string().min(1, 'GitHub ID is required'),
    username: z.string().min(1, 'GitHub username is required')
  }).optional()
});

// User Preferences - matches models/User.ts exactly
const userPreferencesSchema = z.object({
  theme: z.enum(['light', 'dark', 'system']).default('system'),
  language: z.string().min(2).max(5).default('en'),
  timezone: z.string().default('UTC'),
  emailNotifications: z.boolean().default(true),
  pushNotifications: z.boolean().default(true),
  defaultView: z.enum(['grid', 'list']).default('grid'),
  uploadQuality: z.enum(['original', 'high', 'medium']).default('original')
});

// Create User - comprehensive user creation with all fields
export const createUserSchema = z.object({
  email: emailSchema,
  password: z.string().min(8, 'Password must be at least 8 characters').optional(), // Optional for OAuth users
  name: nameSchema,
  avatar: z.string().url('Invalid avatar URL').optional(),
  role: z.enum(['viewer', 'user', 'moderator', 'admin']).default('user'),
  emailVerified: z.boolean().default(false),
  emailVerificationToken: z.string().optional(),
  twoFactorEnabled: z.boolean().default(false),
  isActive: z.boolean().default(true),
  storageQuota: z.number().min(1024 * 1024, 'Storage quota must be at least 1MB').default(15 * 1024 * 1024 * 1024), // 15GB
  providers: oauthProvidersSchema.optional(),
  preferences: userPreferencesSchema.optional(),
  subscriptionStatus: z.enum(['active', 'inactive', 'trial', 'expired', 'cancelled']).default('trial'),
  trialEndsAt: z.string().datetime().optional()
});

// Admin User Update - comprehensive admin operations
export const adminUpdateUserSchema = z.object({
  name: nameSchema.optional(),
  email: emailSchema.optional(),
  avatar: z.string().url('Invalid avatar URL').optional(),
  role: z.enum(['viewer', 'user', 'moderator', 'admin']).optional(),
  emailVerified: z.boolean().optional(),
  isActive: z.boolean().optional(),
  isBanned: z.boolean().optional(),
  banReason: z.string().max(500, 'Ban reason cannot exceed 500 characters').optional(),
  bannedAt: z.string().datetime().optional(),
  bannedBy: objectIdSchema.optional(),
  storageQuota: z.number().min(1024 * 1024, 'Storage quota must be at least 1MB').optional(),
  subscriptionStatus: z.enum(['active', 'inactive', 'trial', 'expired', 'cancelled']).optional(),
  trialEndsAt: z.string().datetime().optional(),
  preferences: userPreferencesSchema.optional(),
  twoFactorEnabled: z.boolean().optional()
});

// Ban User - specific ban operation
export const banUserSchema = z.object({
  userId: objectIdSchema,
  reason: z.string().min(10, 'Ban reason must be at least 10 characters').max(500, 'Ban reason cannot exceed 500 characters'),
  duration: z.enum(['temporary', 'permanent']).default('permanent'),
  expiresAt: z.string().datetime().optional()
}).refine((data) => {
  if (data.duration === 'temporary' && !data.expiresAt) {
    return false;
  }
  return true;
}, {
  message: 'Expiration date is required for temporary bans',
  path: ['expiresAt']
});

// User Verification
export const verifyUserEmailSchema = z.object({
  userId: objectIdSchema,
  sendNotification: z.boolean().default(true)
});

// Reset User Password (Admin)
export const adminResetPasswordSchema = z.object({
  userId: objectIdSchema,
  newPassword: z.string().min(8, 'Password must be at least 8 characters').optional(),
  sendEmail: z.boolean().default(true),
  forcePasswordChange: z.boolean().default(true)
});

// Impersonate User
export const impersonateUserSchema = z.object({
  userId: objectIdSchema,
  reason: z.string().min(5, 'Reason must be at least 5 characters').max(200, 'Reason cannot exceed 200 characters'),
  duration: z.number().min(5).max(120).default(30) // minutes
});

// User Profile Update - for regular users
export const updateUserProfileSchema = z.object({
  name: nameSchema.optional(),
  avatar: z.string().url('Invalid avatar URL').optional(),
  preferences: userPreferencesSchema.optional()
});

// Advanced User Filters
export const userFiltersSchema = z.object({
  query: z.string().max(100, 'Search query cannot exceed 100 characters').optional(),
  role: z.enum(['viewer', 'user', 'moderator', 'admin']).optional(),
  isActive: z.boolean().optional(),
  isBanned: z.boolean().optional(),
  emailVerified: z.boolean().optional(),
  twoFactorEnabled: z.boolean().optional(),
  subscriptionStatus: z.enum(['active', 'inactive', 'trial', 'expired', 'cancelled']).optional(),
  hasOAuthProvider: z.enum(['google', 'github']).optional(),
  dateRange: z.object({
    start: z.string().datetime().optional(),
    end: z.string().datetime().optional()
  }).optional(),
  storageUsage: z.object({
    min: z.number().min(0).optional(),
    max: z.number().min(0).optional()
  }).optional(),
  lastLoginRange: z.object({
    start: z.string().datetime().optional(),
    end: z.string().datetime().optional()
  }).optional(),
  page: z.number().min(1).default(1),
  limit: z.number().min(1).max(100).default(20),
  sort: z.enum(['name', 'email', 'createdAt', 'lastLogin', 'storageUsed', 'storageQuota']).default('createdAt'),
  order: z.enum(['asc', 'desc']).default('desc')
});

// User Export
export const exportUsersSchema = z.object({
  format: z.enum(['csv', 'json', 'xlsx']).default('csv'),
  fields: z.array(z.enum([
    '_id', 'email', 'name', 'role', 'emailVerified', 'isActive', 'isBanned', 'twoFactorEnabled',
    'storageUsed', 'storageQuota', 'subscriptionStatus', 'createdAt', 'lastLogin', 'providers'
  ])).min(1, 'At least one field must be selected'),
  filters: userFiltersSchema.omit({ page: true, limit: true }).optional()
});

export type CreateUserRequest = z.infer<typeof createUserSchema>;
export type AdminUpdateUserRequest = z.infer<typeof adminUpdateUserSchema>;
export type BanUserRequest = z.infer<typeof banUserSchema>;
export type VerifyUserEmailRequest = z.infer<typeof verifyUserEmailSchema>;
export type AdminResetPasswordRequest = z.infer<typeof adminResetPasswordSchema>;
export type ImpersonateUserRequest = z.infer<typeof impersonateUserSchema>;
export type UpdateUserRequest = z.infer<typeof updateUserProfileSchema>;
export type UserFiltersRequest = z.infer<typeof userFiltersSchema>;
export type ExportUsersRequest = z.infer<typeof exportUsersSchema>;
