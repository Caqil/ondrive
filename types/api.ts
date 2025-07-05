import { BaseDocument, BaseResponse, ObjectId } from ".";

export interface ApiKeyPermissions {
  read: boolean;
  write: boolean;
  delete: boolean;
  share: boolean;
  admin: boolean;
}

export interface ApiKey extends BaseDocument {
  user: ObjectId;
  team?: ObjectId;
  name: string;
  description?: string;
  key: string;
  keyPreview: string;
  permissions: ApiKeyPermissions;
  allowedIPs: string[];
  allowedDomains: string[];
  rateLimit: number;
  lastUsedAt?: Date;
  usageCount: number;
  isActive: boolean;
  expiresAt?: Date;
}

export interface CreateApiKeyRequest {
  name: string;
  description?: string;
  permissions: Partial<ApiKeyPermissions>;
  allowedIPs?: string[];
  allowedDomains?: string[];
  rateLimit?: number;
  expiresAt?: Date;
}

export interface ApiResponse<T = any> extends BaseResponse<T> {
  meta?: {
    timestamp: Date;
    requestId: string;
    version: string;
    rateLimit?: {
      limit: number;
      remaining: number;
      reset: Date;
    };
  };
}

export interface ApiError {
  code: string;
  message: string;
  details?: any;
  field?: string;
}

export interface ApiValidationError extends ApiError {
  code: 'VALIDATION_ERROR';
  errors: {
    field: string;
    message: string;
    code: string;
  }[];
}