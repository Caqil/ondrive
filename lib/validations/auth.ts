// lib/validations/auth.ts
import { z } from 'zod';

// Base validation schemas
const emailSchema = z.string()
  .email('Invalid email address')
  .min(1, 'Email is required')
  .max(254, 'Email cannot exceed 254 characters')
  .toLowerCase();

// Password validation matching your lib/auth.ts patterns
const passwordSchema = z.string()
  .min(8, 'Password must be at least 8 characters long')
  .max(128, 'Password cannot exceed 128 characters')
  .regex(/(?=.*[a-z])/, 'Password must contain at least one lowercase letter')
  .regex(/(?=.*[A-Z])/, 'Password must contain at least one uppercase letter')
  .regex(/(?=.*\d)/, 'Password must contain at least one number')
  .regex(/(?=.*[!@#$%^&*(),.?":{}|<>])/, 'Password must contain at least one special character');

const nameSchema = z.string()
  .min(1, 'Name is required')
  .max(100, 'Name cannot exceed 100 characters')
  .regex(/^[a-zA-Z\s\-'\.]+$/, 'Name can only contain letters, spaces, hyphens, apostrophes, and periods')
  .transform((name) => name.trim());

const twoFactorCodeSchema = z.string()
  .length(6, '2FA code must be exactly 6 digits')
  .regex(/^\d{6}$/, '2FA code must contain only numbers');

// Login Request - matches auth store login method
export const loginSchema = z.object({
  email: emailSchema,
  password: z.string().min(1, 'Password is required'),
  rememberMe: z.boolean().optional().default(false)
});

// Register Request - matches auth store register method and types/auth.ts
export const registerSchema = z.object({
  email: emailSchema,
  password: passwordSchema,
  name: nameSchema,
  acceptTerms: z.boolean().refine((val) => val === true, {
    message: 'You must accept the terms and conditions'
  })
});

// Forgot Password - matches auth store forgotPassword method
export const forgotPasswordSchema = z.object({
  email: emailSchema
});

// Reset Password - matches auth store resetPassword method and types/auth.ts
export const resetPasswordSchema = z.object({
  token: z.string().min(1, 'Reset token is required'),
  password: passwordSchema
});

// Change Password - matches auth store changePassword method and types/auth.ts
export const changePasswordSchema = z.object({
  currentPassword: z.string().min(1, 'Current password is required'),
  newPassword: passwordSchema
}).refine((data) => data.currentPassword !== data.newPassword, {
  message: 'New password must be different from current password',
  path: ['newPassword']
});

// Two Factor Setup - matches auth store setup2FA method and types/auth.ts
export const twoFactorSetupSchema = z.object({
  secret: z.string().min(1, 'Secret is required'),
  code: twoFactorCodeSchema
});

// Two Factor Verify - matches auth store verify2FA method and types/auth.ts
export const twoFactorVerifySchema = z.object({
  code: twoFactorCodeSchema,
  sessionToken: z.string().optional()
});

// Two Factor Disable - matches auth store disable2FA method
export const twoFactorDisableSchema = z.object({
  code: twoFactorCodeSchema
});

// Update Profile - matches auth store updateProfile method and types/user.ts UserPreferences
export const updateProfileSchema = z.object({
  name: nameSchema.optional(),
  avatar: z.string().url('Invalid avatar URL').optional(),
  preferences: z.object({
    theme: z.enum(['light', 'dark', 'system']).optional(),
    language: z.string().min(2).max(5).optional(), // Language codes like 'en', 'en-US'
    timezone: z.string().optional(),
    emailNotifications: z.boolean().optional(),
    pushNotifications: z.boolean().optional(),
    defaultView: z.enum(['grid', 'list']).optional(),
    uploadQuality: z.enum(['original', 'high', 'medium']).optional()
  }).optional()
});

// Email Verification
export const emailVerificationSchema = z.object({
  token: z.string().min(1, 'Verification token is required')
});

// Resend Email Verification
export const resendEmailVerificationSchema = z.object({
  email: emailSchema
});

// OAuth Provider (for Google, GitHub integration)
export const oauthProviderSchema = z.object({
  provider: z.enum(['google', 'github']),
  code: z.string().min(1, 'Authorization code is required'),
  state: z.string().optional(),
  redirectUri: z.string().url('Invalid redirect URI').optional()
});

// Session Management
export const sessionSchema = z.object({
  deviceName: z.string().max(100, 'Device name cannot exceed 100 characters').optional(),
  userAgent: z.string().optional(),
  ipAddress: z.string().ip('Invalid IP address').optional()
});

// Account Deletion
export const deleteAccountSchema = z.object({
  password: z.string().min(1, 'Password is required to delete account'),
  confirmation: z.string().refine((val) => val === 'DELETE', {
    message: 'Please type "DELETE" to confirm account deletion'
  }),
  reason: z.string().max(500, 'Reason cannot exceed 500 characters').optional()
});

// Avatar Upload
export const avatarUploadSchema = z.object({
  file: z.instanceof(File, { message: 'File is required' })
    .refine((file) => file.size <= 5 * 1024 * 1024, 'Avatar must be less than 5MB')
    .refine((file) => ['image/jpeg', 'image/jpg', 'image/png', 'image/webp'].includes(file.type), 
      'Avatar must be a JPEG, PNG, or WebP image')
});

// Rate Limiting - for login attempts tracking
export const loginAttemptSchema = z.object({
  email: emailSchema,
  ip: z.string().ip('Invalid IP address'),
  userAgent: z.string().optional(),
  success: z.boolean()
});

// Password Strength Check
export const passwordStrengthSchema = z.object({
  password: z.string().min(1, 'Password is required')
});

// Check Email Availability
export const checkEmailSchema = z.object({
  email: emailSchema
});

// Backup Codes (for 2FA recovery)
export const generateBackupCodesSchema = z.object({
  password: z.string().min(1, 'Password is required to generate backup codes')
});

export const useBackupCodeSchema = z.object({
  email: emailSchema,
  password: z.string().min(1, 'Password is required'),
  backupCode: z.string().length(8, 'Backup code must be exactly 8 characters'),
  rememberMe: z.boolean().optional().default(false)
});

// Export types matching your types/auth.ts
export type LoginRequest = z.infer<typeof loginSchema>;
export type RegisterRequest = z.infer<typeof registerSchema>;
export type ForgotPasswordRequest = z.infer<typeof forgotPasswordSchema>;
export type ResetPasswordRequest = z.infer<typeof resetPasswordSchema>;
export type ChangePasswordRequest = z.infer<typeof changePasswordSchema>;
export type TwoFactorSetupRequest = z.infer<typeof twoFactorSetupSchema>;
export type TwoFactorVerifyRequest = z.infer<typeof twoFactorVerifySchema>;
export type TwoFactorDisableRequest = z.infer<typeof twoFactorDisableSchema>;
export type UpdateUserRequest = z.infer<typeof updateProfileSchema>; // Matches your types/auth.ts
export type EmailVerificationRequest = z.infer<typeof emailVerificationSchema>;
export type ResendEmailVerificationRequest = z.infer<typeof resendEmailVerificationSchema>;
export type OAuthProviderRequest = z.infer<typeof oauthProviderSchema>;
export type SessionRequest = z.infer<typeof sessionSchema>;
export type DeleteAccountRequest = z.infer<typeof deleteAccountSchema>;
export type AvatarUploadRequest = z.infer<typeof avatarUploadSchema>;
export type LoginAttemptRequest = z.infer<typeof loginAttemptSchema>;
export type PasswordStrengthRequest = z.infer<typeof passwordStrengthSchema>;
export type CheckEmailRequest = z.infer<typeof checkEmailSchema>;
export type GenerateBackupCodesRequest = z.infer<typeof generateBackupCodesSchema>;
export type UseBackupCodeRequest = z.infer<typeof useBackupCodeSchema>;