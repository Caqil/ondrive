import mongoose, { Document, Schema } from 'mongoose';

export interface IFolder extends Document {
  _id: string;
  name: string;
  description?: string;
  path: string; // Full path for quick access
  depth: number; // Folder depth level
  
  // Hierarchy
  parent?: mongoose.Types.ObjectId;
  children: mongoose.Types.ObjectId[];
  ancestors: mongoose.Types.ObjectId[]; // For quick ancestor queries
  
  // Ownership
  owner: mongoose.Types.ObjectId;
  team?: mongoose.Types.ObjectId;
  
  // Access control
  visibility: 'private' | 'team' | 'public';
  isStarred: boolean;
  isTrashed: boolean;
  trashedAt?: Date;
  trashedBy?: mongoose.Types.ObjectId;
  
  // Content counts
  fileCount: number;
  folderCount: number;
  totalSize: number; // Total size of all files in this folder and subfolders
  
  // Visual customization
  color?: string;
  icon?: string;
  
  // Sharing
  shareCount: number;
  
  // Sync status
  syncStatus: 'synced' | 'pending' | 'conflict' | 'error';
  lastSyncAt?: Date;
  
  createdAt: Date;
  updatedAt: Date;
}

const folderSchema = new Schema<IFolder>({
  name: {
    type: String,
    required: true,
    trim: true,
    maxlength: 255,
    index: true
  },
  description: {
    type: String,
    trim: true,
    maxlength: 1000
  },
  path: {
    type: String,
    required: true,
    index: true
  },
  depth: {
    type: Number,
    required: true,
    min: 0,
    max: 20, // Prevent infinite nesting
    index: true
  },
  
  parent: {
    type: Schema.Types.ObjectId,
    ref: 'Folder',
    index: true
  },
  children: [{
    type: Schema.Types.ObjectId,
    ref: 'Folder'
  }],
  ancestors: [{
    type: Schema.Types.ObjectId,
    ref: 'Folder'
  }],
  
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
  
  fileCount: {
    type: Number,
    default: 0,
    min: 0
  },
  folderCount: {
    type: Number,
    default: 0,
    min: 0
  },
  totalSize: {
    type: Number,
    default: 0,
    min: 0
  },
  
  color: {
    type: String,
    match: /^#[0-9A-F]{6}$/i
  },
  icon: {
    type: String,
    maxlength: 50
  },
  
  shareCount: {
    type: Number,
    default: 0
  },
  
  syncStatus: {
    type: String,
    enum: ['synced', 'pending', 'conflict', 'error'],
    default: 'synced',
    index: true
  },
  lastSyncAt: Date
}, {
  timestamps: true,
  collection: 'folders'
});

// Compound indexes
folderSchema.index({ owner: 1, parent: 1 });
folderSchema.index({ owner: 1, isTrashed: 1 });
folderSchema.index({ parent: 1, isTrashed: 1 });
folderSchema.index({ team: 1, visibility: 1 });
folderSchema.index({ ancestors: 1 });

// Prevent circular references
folderSchema.pre('save', function(next) {
  if (this.parent && this.parent.equals(this._id)) {
    return next(new Error('Folder cannot be its own parent'));
  }
  if (this.ancestors.some(ancestor => ancestor.equals(this._id))) {
    return next(new Error('Circular reference detected in folder hierarchy'));
  }
  next();
});

export const Folder = mongoose.models.Folder || mongoose.model<IFolder>('Folder', folderSchema);
