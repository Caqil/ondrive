import mongoose, { Document, Schema } from 'mongoose';

export interface IStorageProvider extends Document {
  _id: string;
  
  // Provider details
  name: string;
  type: 'local' | 's3' | 'r2' | 'wasabi' | 'gcs' | 'azure';
  
  // Configuration
  config: {
    // S3-compatible
    accessKeyId?: string;
    secretAccessKey?: string;
    region?: string;
    bucket?: string;
    endpoint?: string;
    forcePathStyle?: boolean;
    
    // Google Cloud Storage
    projectId?: string;
    keyFilename?: string;
    
    // Azure
    accountName?: string;
    accountKey?: string;
    containerName?: string;
    
    // Local
    path?: string;
    
    // Common
    maxFileSize?: number;
    allowedMimeTypes?: string[];
  };
  
  // Status
  isActive: boolean;
  isDefault: boolean;
  
  // Usage stats
  totalFiles: number;
  totalSize: number; // bytes
  
  // Health check
  lastHealthCheck?: Date;
  healthStatus: 'healthy' | 'warning' | 'error' | 'unknown';
  healthMessage?: string;
  
  // Performance metrics
  averageUploadTime?: number; // milliseconds
  averageDownloadTime?: number; // milliseconds
  
  createdAt: Date;
  updatedAt: Date;
}

const storageProviderSchema = new Schema<IStorageProvider>({
  name: {
    type: String,
    required: true,
    trim: true,
    maxlength: 100
  },
  type: {
    type: String,
    enum: ['local', 's3', 'r2', 'wasabi', 'gcs', 'azure'],
    required: true,
    index: true
  },
  
  config: {
    accessKeyId: String,
    secretAccessKey: String,
    region: String,
    bucket: String,
    endpoint: String,
    forcePathStyle: Boolean,
    projectId: String,
    keyFilename: String,
    accountName: String,
    accountKey: String,
    containerName: String,
    path: String,
    maxFileSize: Number,
    allowedMimeTypes: [String]
  },
  
  isActive: {
    type: Boolean,
    default: true,
    index: true
  },
  isDefault: {
    type: Boolean,
    default: false,
    index: true
  },
  
  totalFiles: {
    type: Number,
    default: 0,
    min: 0
  },
  totalSize: {
    type: Number,
    default: 0,
    min: 0
  },
  
  lastHealthCheck: Date,
  healthStatus: {
    type: String,
    enum: ['healthy', 'warning', 'error', 'unknown'],
    default: 'unknown',
    index: true
  },
  healthMessage: String,
  
  averageUploadTime: Number,
  averageDownloadTime: Number
}, {
  timestamps: true,
  collection: 'storage_providers'
});

// Ensure only one default provider
storageProviderSchema.index({ isDefault: 1 }, { 
  unique: true, 
  partialFilterExpression: { isDefault: true } 
});

export const StorageProvider = mongoose.models.StorageProvider || mongoose.model<IStorageProvider>('StorageProvider', storageProviderSchema);