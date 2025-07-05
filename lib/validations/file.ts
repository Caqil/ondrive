// lib/validations/file.ts
import { z } from 'zod';

// Base ObjectId validation
const objectIdSchema = z.string().regex(/^[0-9a-fA-F]{24}$/, 'Invalid ObjectId format');

// File Label validation - matches types/file.ts
const fileLabelSchema = z.object({
  name: z.string().min(1, 'Label name is required').max(50, 'Label name cannot exceed 50 characters'),
  color: z.string().regex(/^#[0-9A-F]{6}$/i, 'Invalid color format')
});

// File Upload Request - matches file store and types/file.ts
export const fileUploadSchema = z.object({
  name: z.string()
    .min(1, 'File name is required')
    .max(255, 'File name cannot exceed 255 characters')
    .regex(/^[^<>:"/\\|?*]+$/, 'File name contains invalid characters'),
  size: z.number()
    .min(1, 'File size must be greater than 0')
    .max(750 * 1024 * 1024 * 1024, 'File size cannot exceed 750GB'), // Based on your app settings
  mimeType: z.string().min(1, 'MIME type is required'),
  folderId: objectIdSchema,
  tags: z.array(z.string().min(1).max(50)).max(20, 'Cannot have more than 20 tags').optional(),
  labels: z.array(fileLabelSchema).max(10, 'Cannot have more than 10 labels').optional()
});

// Create Folder Request - matches file store createFolder and types/folder.ts
export const createFolderSchema = z.object({
  name: z.string()
    .min(1, 'Folder name is required')
    .max(255, 'Folder name cannot exceed 255 characters')
    .regex(/^[^<>:"/\\|?*]+$/, 'Folder name contains invalid characters'),
  description: z.string().max(1000, 'Description cannot exceed 1000 characters').optional(),
  parentId: objectIdSchema.optional(),
  color: z.string().regex(/^#[0-9A-F]{6}$/i, 'Invalid color format').optional(),
  icon: z.string().max(50, 'Icon cannot exceed 50 characters').optional()
});

// Update File/Folder Request - for rename operations
export const updateFileSchema = z.object({
  name: z.string()
    .min(1, 'Name is required')
    .max(255, 'Name cannot exceed 255 characters')
    .regex(/^[^<>:"/\\|?*]+$/, 'Name contains invalid characters')
    .optional(),
  tags: z.array(z.string().min(1).max(50)).max(20, 'Cannot have more than 20 tags').optional(),
  labels: z.array(fileLabelSchema).max(10, 'Cannot have more than 10 labels').optional(),
  visibility: z.enum(['private', 'team', 'public']).optional() // Matches models/File.ts
});

export const updateFolderSchema = z.object({
  name: z.string()
    .min(1, 'Name is required')
    .max(255, 'Name cannot exceed 255 characters')
    .regex(/^[^<>:"/\\|?*]+$/, 'Name contains invalid characters')
    .optional(),
  description: z.string().max(1000, 'Description cannot exceed 1000 characters').optional(),
  color: z.string().regex(/^#[0-9A-F]{6}$/i, 'Invalid color format').optional(),
  icon: z.string().max(50, 'Icon cannot exceed 50 characters').optional(),
  visibility: z.enum(['private', 'team', 'public']).optional()
});

// Bulk Operations - matches file store patterns
export const bulkFileOperationSchema = z.object({
  fileIds: z.array(objectIdSchema).min(1, 'At least one file must be selected').max(100, 'Cannot perform bulk operations on more than 100 files'),
  action: z.enum(['move', 'copy', 'delete', 'star', 'unstar', 'trash', 'restore']),
  targetFolderId: objectIdSchema.optional() // Required for move/copy operations
}).refine((data) => {
  if (['move', 'copy'].includes(data.action) && !data.targetFolderId) {
    return false;
  }
  return true;
}, {
  message: 'Target folder is required for move and copy operations',
  path: ['targetFolderId']
});

export const bulkFolderOperationSchema = z.object({
  folderIds: z.array(objectIdSchema).min(1, 'At least one folder must be selected').max(100, 'Cannot perform bulk operations on more than 100 folders'),
  action: z.enum(['move', 'delete', 'trash', 'restore']),
  targetFolderId: objectIdSchema.optional() // Required for move operations
}).refine((data) => {
  if (data.action === 'move' && !data.targetFolderId) {
    return false;
  }
  return true;
}, {
  message: 'Target folder is required for move operations',
  path: ['targetFolderId']
});

// File Search - matches types/file.ts FileSearchParams and search store
export const fileSearchSchema = z.object({
  query: z.string().min(1, 'Search query is required').max(100, 'Search query cannot exceed 100 characters'),
  folderId: objectIdSchema.optional(),
  mimeType: z.string().optional(),
  size: z.object({
    min: z.number().min(0).optional(),
    max: z.number().min(0).optional()
  }).optional().refine((data) => {
    if (data?.min && data?.max && data.min > data.max) {
      return false;
    }
    return true;
  }, {
    message: 'Minimum size cannot be greater than maximum size'
  }),
  dateRange: z.object({
    start: z.string().datetime().optional(),
    end: z.string().datetime().optional()
  }).optional().refine((data) => {
    if (data?.start && data?.end && new Date(data.start) > new Date(data.end)) {
      return false;
    }
    return true;
  }, {
    message: 'Start date cannot be after end date'
  }),
  tags: z.array(z.string()).optional(),
  owner: objectIdSchema.optional(),
  isStarred: z.boolean().optional(),
  isTrashed: z.boolean().optional(),
  page: z.number().min(1).default(1),
  limit: z.number().min(1).max(100).default(20),
  sort: z.enum(['name', 'size', 'modified', 'type', 'createdAt']).default('name'),
  order: z.enum(['asc', 'desc']).default('asc')
});

// Save Search - for search store saveSearch
export const saveSearchSchema = z.object({
  name: z.string().min(1, 'Search name is required').max(100, 'Search name cannot exceed 100 characters'),
  query: z.string().min(1, 'Search query is required'),
  filters: fileSearchSchema.omit({ query: true, page: true, limit: true }).optional()
});

// File Stats Request
export const fileStatsSchema = z.object({
  period: z.enum(['today', 'week', 'month', 'year']).default('month'),
  groupBy: z.enum(['day', 'week', 'month']).default('day')
});

// Version Management
export const createVersionSchema = z.object({
  fileId: objectIdSchema,
  file: z.instanceof(File),
  comment: z.string().max(500, 'Comment cannot exceed 500 characters').optional()
});

export const restoreVersionSchema = z.object({
  fileId: objectIdSchema,
  versionId: objectIdSchema
});

// Download Request
export const downloadRequestSchema = z.object({
  fileIds: z.array(objectIdSchema).min(1, 'At least one file must be selected').max(100, 'Cannot download more than 100 files'),
  format: z.enum(['zip', 'tar']).default('zip').optional()
});

// Trash Operations
export const trashRequestSchema = z.object({
  fileIds: z.array(objectIdSchema).optional(),
  folderIds: z.array(objectIdSchema).optional()
}).refine((data) => {
  return (data.fileIds && data.fileIds.length > 0) || (data.folderIds && data.folderIds.length > 0);
}, {
  message: 'At least one file or folder must be specified'
});

export const restoreRequestSchema = z.object({
  fileIds: z.array(objectIdSchema).optional(),
  folderIds: z.array(objectIdSchema).optional()
}).refine((data) => {
  return (data.fileIds && data.fileIds.length > 0) || (data.folderIds && data.folderIds.length > 0);
}, {
  message: 'At least one file or folder must be specified'
});

// Upload Configuration
export const uploadConfigSchema = z.object({
  maxFileSize: z.number().min(1024, 'Max file size must be at least 1KB'),
  allowedMimeTypes: z.array(z.string()),
  chunkSize: z.number().min(1024 * 1024, 'Chunk size must be at least 1MB'), // 1MB minimum
  maxChunks: z.number().min(1, 'Max chunks must be at least 1'),
  storageProvider: z.enum(['local', 's3', 'r2', 'wasabi', 'gcs', 'azure'])
});

// Chunk Upload
export const chunkUploadSchema = z.object({
  uploadId: z.string().min(1, 'Upload ID is required'),
  chunkIndex: z.number().min(0, 'Chunk index cannot be negative'),
  chunk: z.instanceof(File)
});

// Complete Upload
export const completeUploadSchema = z.object({
  uploadId: z.string().min(1, 'Upload ID is required')
});

// Abort Upload
export const abortUploadSchema = z.object({
  uploadId: z.string().min(1, 'Upload ID is required')
});

// Export types matching your types/file.ts
export type FileUploadRequest = z.infer<typeof fileUploadSchema>;
export type CreateFolderRequest = z.infer<typeof createFolderSchema>;
export type UpdateFileRequest = z.infer<typeof updateFileSchema>;
export type UpdateFolderRequest = z.infer<typeof updateFolderSchema>;
export type BulkFileOperationRequest = z.infer<typeof bulkFileOperationSchema>;
export type BulkFolderOperationRequest = z.infer<typeof bulkFolderOperationSchema>;
export type FileSearchParams = z.infer<typeof fileSearchSchema>;
export type SaveSearchRequest = z.infer<typeof saveSearchSchema>;
export type FileStatsRequest = z.infer<typeof fileStatsSchema>;
export type CreateVersionRequest = z.infer<typeof createVersionSchema>;
export type RestoreVersionRequest = z.infer<typeof restoreVersionSchema>;
export type DownloadRequest = z.infer<typeof downloadRequestSchema>;
export type TrashRequest = z.infer<typeof trashRequestSchema>;
export type RestoreRequest = z.infer<typeof restoreRequestSchema>;
export type UploadConfigRequest = z.infer<typeof uploadConfigSchema>;
export type ChunkUploadRequest = z.infer<typeof chunkUploadSchema>;
export type CompleteUploadRequest = z.infer<typeof completeUploadSchema>;
export type AbortUploadRequest = z.infer<typeof abortUploadSchema>;