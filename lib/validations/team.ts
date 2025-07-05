import { z } from 'zod';

// Create Team - matches team store and models/Team.ts
export const createTeamSchema = z.object({
  name: z.string().min(1, 'Team name is required').max(100, 'Team name cannot exceed 100 characters'),
  description: z.string().max(500, 'Description cannot exceed 500 characters').optional(),
  slug: z.string()
    .min(3, 'Slug must be at least 3 characters')
    .max(50, 'Slug cannot exceed 50 characters')
    .regex(/^[a-z0-9-]+$/, 'Slug can only contain lowercase letters, numbers, and hyphens')
    .refine(slug => !slug.startsWith('-') && !slug.endsWith('-'), 'Slug cannot start or end with a hyphen'),
  billingEmail: z.string().email('Invalid billing email').optional()
});

// Update Team - matches models/Team.ts structure
export const updateTeamSchema = z.object({
  name: z.string().min(1).max(100).optional(),
  description: z.string().max(500).optional(),
  logo: z.string().url('Invalid logo URL').optional(),
  color: z.string().regex(/^#[0-9A-F]{6}$/i, 'Invalid color format').optional(),
  billingEmail: z.string().email('Invalid billing email').optional(),
  
  // Settings from models/Team.ts
  settings: z.object({
    allowMemberInvites: z.boolean().optional(),
    requireApprovalForJoining: z.boolean().optional(),
    defaultMemberRole: z.enum(['member', 'editor', 'admin']).optional(), // From models/Team.ts
    enableGuestAccess: z.boolean().optional(),
    enablePublicSharing: z.boolean().optional(),
    enforceSSO: z.boolean().optional(),
    allowedDomains: z.array(z.string().email().transform(email => email.split('@')[1])).optional()
  }).optional(),
  
  // Features from models/Team.ts
  features: z.object({
    enableAdvancedSharing: z.boolean().optional(),
    enableVersionHistory: z.boolean().optional(),
    enableAuditLogs: z.boolean().optional(),
    enableAPIAccess: z.boolean().optional(),
    enableIntegrations: z.boolean().optional(),
    maxFileSize: z.number().min(1024, 'Max file size must be at least 1KB').optional(),
    enableOCR: z.boolean().optional()
  }).optional()
});

// Invite Team Member - matches team store and models/TeamMember.ts
export const inviteTeamMemberSchema = z.object({
  email: z.string().email('Invalid email address'),
  role: z.enum(['admin', 'editor', 'member', 'viewer']), // From models/TeamMember.ts (excluding 'owner')
  message: z.string().max(1000, 'Message cannot exceed 1000 characters').optional()
});

// Update Member Role - matches team store
export const updateMemberRoleSchema = z.object({
  role: z.enum(['admin', 'editor', 'member', 'viewer']) // From models/TeamMember.ts (excluding 'owner')
});

// Update Member Permissions - matches models/TeamMember.ts permissions
export const updateMemberPermissionsSchema = z.object({
  permissions: z.object({
    canUpload: z.boolean().optional(),
    canDownload: z.boolean().optional(),
    canShare: z.boolean().optional(),
    canDelete: z.boolean().optional(),
    canInvite: z.boolean().optional(),
    canManageTeam: z.boolean().optional(),
    canViewBilling: z.boolean().optional(),
    canManageBilling: z.boolean().optional()
  })
});

// Team Filters
export const teamFiltersSchema = z.object({
  plan: z.enum(['free', 'basic', 'pro', 'enterprise']).optional(), // From models/Team.ts
  isActive: z.boolean().optional(),
  isSuspended: z.boolean().optional(),
  dateRange: z.object({
    start: z.string().datetime().optional(),
    end: z.string().datetime().optional()
  }).optional(),
  page: z.number().min(1).default(1),
  limit: z.number().min(1).max(100).default(20),
  sort: z.enum(['name', 'createdAt', 'memberCount', 'storageUsed']).default('createdAt'),
  order: z.enum(['asc', 'desc']).default('desc')
});

export type CreateTeamRequest = z.infer<typeof createTeamSchema>;
export type UpdateTeamRequest = z.infer<typeof updateTeamSchema>;
export type InviteTeamMemberRequest = z.infer<typeof inviteTeamMemberSchema>;
export type UpdateMemberRoleRequest = z.infer<typeof updateMemberRoleSchema>;
export type UpdateMemberPermissionsRequest = z.infer<typeof updateMemberPermissionsSchema>;
export type TeamFiltersRequest = z.infer<typeof teamFiltersSchema>;
