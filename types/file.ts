import { BaseDocument, ObjectId, SearchParams } from "./global";

export interface FileMetadata {
  width?: number;
  height?: number;
  duration?: number;
  pages?: number;
  [key: string]: any;
}

export interface FileLabel {
  name: string;
  color: string;
}

export interface File extends BaseDocument {
  name: string;
  originalName: string;
  mimeType: string;
  size: number;
  path: string;
  key: string;
  url?: string;
  thumbnailUrl?: string;
  previewUrl?: string;
  extension: string;
  encoding?: string;
  checksum: string;
  ocrText?: string;
  metadata: FileMetadata;
  folder: ObjectId;
  owner: ObjectId;
  team?: ObjectId;
  visibility: 'private' | 'team' | 'public';
  isStarred: boolean;
  isTrashed: boolean;
  trashedAt?: Date;
  trashedBy?: ObjectId;
  shareCount: number;
  downloadCount: number;
  viewCount: number;
  lastAccessedAt?: Date;
  version: number;
  parentVersion?: ObjectId;
  isLatestVersion: boolean;
  versionHistory: ObjectId[];
  processingStatus: 'pending' | 'processing' | 'completed' | 'failed';
  processingError?: string;
  storageProvider: 'local' | 's3' | 'r2' | 'wasabi' | 'gcs' | 'azure';
  tags: string[];
  labels: FileLabel[];
  syncStatus: 'synced' | 'pending' | 'conflict' | 'error';
  lastSyncAt?: Date;
}

export interface FileUploadRequest {
  name: string;
  size: number;
  mimeType: string;
  folderId: ObjectId;
  tags?: string[];
  labels?: FileLabel[];
}

export interface FileUploadResponse {
  uploadUrl: string;
  uploadId: string;
  fileId: ObjectId;
  chunkSize?: number;
  maxChunks?: number;
}

export interface FileSearchParams extends SearchParams {
  folderId?: ObjectId;
  mimeType?: string;
  size?: { min?: number; max?: number };
  dateRange?: { start?: Date; end?: Date };
  tags?: string[];
  owner?: ObjectId;
  isStarred?: boolean;
  isTrashed?: boolean;
}

export interface FileStats {
  totalFiles: number;
  totalSize: number;
  filesByType: Record<string, { count: number; size: number }>;
  recentUploads: File[];
  largestFiles: File[];
}
