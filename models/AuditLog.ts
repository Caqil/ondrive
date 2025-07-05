import mongoose, { Document, Schema } from 'mongoose';

export interface IAuditLog extends Document {
  _id: string;
  
  // Who performed the action
  user: mongoose.Types.ObjectId;
  impersonatedBy?: mongoose.Types.ObjectId;
  
  // What action was performed
  action: string; // e.g., 'user.create', 'file.delete', 'settings.update'
  category: 'auth' | 'file' | 'admin' | 'billing' | 'team' | 'security';
  
  // What was affected
  resource?: mongoose.Types.ObjectId;
  resourceType?: string;
  resourceName?: string;
  
  // Additional context
  details: {
    oldValue?: any;
    newValue?: any;
    metadata?: any;
  };
  
  // Request context
  ip: string;
  userAgent: string;
  
  // Result
  success: boolean;
  errorMessage?: string;
  
  // Team context
  team?: mongoose.Types.ObjectId;
  
  createdAt: Date;
}

const auditLogSchema = new Schema<IAuditLog>({
  user: {
    type: Schema.Types.ObjectId,
    ref: 'User',
    required: true,
    index: true
  },
  impersonatedBy: {
    type: Schema.Types.ObjectId,
    ref: 'User',
    index: true
  },
  
  action: {
    type: String,
    required: true,
    index: true
  },
  category: {
    type: String,
    enum: ['auth', 'file', 'admin', 'billing', 'team', 'security'],
    required: true,
    index: true
  },
  
  resource: {
    type: Schema.Types.ObjectId,
    refPath: 'resourceType',
    index: true
  },
  resourceType: String,
  resourceName: String,
  
  details: {
    oldValue: Schema.Types.Mixed,
    newValue: Schema.Types.Mixed,
    metadata: Schema.Types.Mixed
  },
  
  ip: {
    type: String,
    required: true,
    index: true
  },
  userAgent: {
    type: String,
    required: true
  },
  
  success: {
    type: Boolean,
    required: true,
    index: true
  },
  errorMessage: String,
  
  team: {
    type: Schema.Types.ObjectId,
    ref: 'Team',
    index: true
  }
}, {
  timestamps: { createdAt: true, updatedAt: false },
  collection: 'audit_logs'
});

// Compound indexes
auditLogSchema.index({ user: 1, createdAt: -1 });
auditLogSchema.index({ category: 1, action: 1, createdAt: -1 });
auditLogSchema.index({ team: 1, createdAt: -1 });
auditLogSchema.index({ resource: 1, resourceType: 1, createdAt: -1 });

// TTL index to automatically delete old audit logs (1 year)
auditLogSchema.index({ createdAt: 1 }, { expireAfterSeconds: 365 * 24 * 60 * 60 });

export const AuditLog = mongoose.models.AuditLog || mongoose.model<IAuditLog>('AuditLog', auditLogSchema);
