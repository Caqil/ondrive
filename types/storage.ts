import { BaseDocument } from ".";

export interface StorageProviderConfig {
  accessKeyId?: string;
  secretAccessKey?: string;
  region?: string;
  bucket?: string;
  endpoint?: string;
  forcePathStyle?: boolean;
  projectId?: string;
  keyFilename?: string;
  accountName?: string;
  accountKey?: string;
  containerName?: string;
  path?: string;
  maxFileSize?: number;
  allowedMimeTypes?: string[];
}

export interface StorageProvider extends BaseDocument {
  name: string;
  type: 'local' | 's3' | 'r2' | 'wasabi' | 'gcs' | 'azure';
  config: StorageProviderConfig;
  isActive: boolean;
  isDefault: boolean;
  totalFiles: number;
  totalSize: number;
  lastHealthCheck?: Date;
  healthStatus: 'healthy' | 'warning' | 'error' | 'unknown';
  healthMessage?: string;
  averageUploadTime?: number;
  averageDownloadTime?: number;
}

export interface StorageUsage {
  totalSize: number;
  totalFiles: number;
  usageByProvider: Record<string, { size: number; files: number }>;
  usageByType: Record<string, { size: number; files: number }>;
  growthRate: number;
}

export interface UploadConfig {
  maxFileSize: number;
  allowedMimeTypes: string[];
  chunkSize: number;
  maxChunks: number;
  storageProvider: string;
}