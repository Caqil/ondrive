import mongoose, { Document, Schema } from 'mongoose';

export interface IFile extends Document {
  _id: string;
  name: string;
  originalName: string;
  mimeType: string;
  size: number;
  path: string; // Storage path
  key: string; // Storage key/identifier
  url?: string; // Public URL if available
  thumbnailUrl?: string;
  previewUrl?: string;
  
  // File metadata
  extension: string;
  encoding?: string;
  checksum: string; // MD5 or SHA256
  
  // OCR and content
  ocrText?: string; // Extracted text for search
  metadata: {
    width?: number;
    height?: number;
    duration?: number; // For videos/audio
    pages?: number; // For PDFs
    [key: string]: any;
  };
  
  // Organization
  folder: mongoose.Types.ObjectId;
  owner: mongoose.Types.ObjectId;
  team?: mongoose.Types.ObjectId;
  
  // Access control
  visibility: 'private' | 'team' | 'public';
  isStarred: boolean;
  isTrashed: boolean;
  trashedAt?: Date;
  trashedBy?: mongoose.Types.ObjectId;
  
  // Sharing
  shareCount: number;
  downloadCount: number;
  viewCount: number;
  lastAccessedAt?: Date;
  
  // Versioning
  version: number;
  parentVersion?: mongoose.Types.ObjectId;
  isLatestVersion: boolean;
  versionHistory: mongoose.Types.ObjectId[];
  
  // Processing status
  processingStatus: 'pending' | 'processing' | 'completed' | 'failed';
  processingError?: string;
  
  // Storage provider
  storageProvider: 'local' | 's3' | 'r2' | 'wasabi' | 'gcs' | 'azure';
  
  // Tags and labels
  tags: string[];
  labels: {
    name: string;
    color: string;
  }[];
  
  // Sync status for offline
  syncStatus: 'synced' | 'pending' | 'conflict' | 'error';
  lastSyncAt?: Date;
  
  createdAt: Date;
  updatedAt: Date;
}

const fileSchema = new Schema<IFile>({
  name: {
    type: String,
    required: true,
    trim: true,
    maxlength: 255,
    index: true
  },
  originalName: {
    type: String,
    required: true,
    trim: true,
    maxlength: 255
  },
  mimeType: {
    type: String,
    required: true,
    index: true
  },
  size: {
    type: Number,
    required: true,
    min: 0,
    index: true
  },
  path: {
    type: String,
    required: true
  },
  key: {
    type: String,
    required: true,
    unique: true,
    index: true
  },
  url: String,
  thumbnailUrl: String,
  previewUrl: String,
  
  extension: {
    type: String,
    required: true,
    lowercase: true,
    index: true
  },
  encoding: String,
  checksum: {
    type: String,
    required: true,
    index: true
  },
  
  ocrText: {
    type: String,
    index: 'text' // Text search index
  },
  metadata: {
    type: Schema.Types.Mixed,
    default: {}
  },
  
  folder: {
    type: Schema.Types.ObjectId,
    ref: 'Folder',
    required: true,
    index: true
  },
  owner: {
    type: Schema.Types.ObjectId,
    ref: 'User',
    required: true,
    index: true
  },
  team: {
    type: Schema.Types.ObjectId,
    ref: 'Team',
    index: true
  },
  
  visibility: {
    type: String,
    enum: ['private', 'team', 'public'],
    default: 'private',
    index: true
  },
  isStarred: {
    type: Boolean,
    default: false,
    index: true
  },
  isTrashed: {
    type: Boolean,
    default: false,
    index: true
  },
  trashedAt: Date,
  trashedBy: {
    type: Schema.Types.ObjectId,
    ref: 'User'
  },
  
  shareCount: {
    type: Number,
    default: 0
  },
  downloadCount: {
    type: Number,
    default: 0
  },
  viewCount: {
    type: Number,
    default: 0
  },
  lastAccessedAt: Date,
  
  version: {
    type: Number,
    default: 1,
    min: 1
  },
  parentVersion: {
    type: Schema.Types.ObjectId,
    ref: 'File'
  },
  isLatestVersion: {
    type: Boolean,
    default: true,
    index: true
  },
  versionHistory: [{
    type: Schema.Types.ObjectId,
    ref: 'File'
  }],
  
  processingStatus: {
    type: String,
    enum: ['pending', 'processing', 'completed', 'failed'],
    default: 'pending',
    index: true
  },
  processingError: String,
  
  storageProvider: {
    type: String,
    enum: ['local', 's3', 'r2', 'wasabi', 'gcs', 'azure'],
    required: true,
    index: true
  },
  
  tags: [{
    type: String,
    trim: true,
    maxlength: 50
  }],
  labels: [{
    name: {
      type: String,
      trim: true,
      maxlength: 50
    },
    color: {
      type: String,
      match: /^#[0-9A-F]{6}$/i
    }
  }],
  
  syncStatus: {
    type: String,
    enum: ['synced', 'pending', 'conflict', 'error'],
    default: 'synced',
    index: true
  },
  lastSyncAt: Date
}, {
  timestamps: true,
  collection: 'files'
});

// Compound indexes
fileSchema.index({ owner: 1, folder: 1 });
fileSchema.index({ owner: 1, isTrashed: 1, createdAt: -1 });
fileSchema.index({ owner: 1, isStarred: 1 });
fileSchema.index({ folder: 1, isTrashed: 1 });
fileSchema.index({ team: 1, visibility: 1 });
fileSchema.index({ mimeType: 1, size: 1 });
fileSchema.index({ checksum: 1, owner: 1 });

// Text search index
fileSchema.index({
  name: 'text',
  originalName: 'text',
  ocrText: 'text',
  tags: 'text'
});

export const File = mongoose.models.File || mongoose.model<IFile>('File', fileSchema);
