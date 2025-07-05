import { BaseResponse, ObjectId, PaginatedResponse, PaginationParams } from './global';

export * from './global';
export * from './user';
export * from './auth';
export * from './file';
export * from './folder';
export * from './team';
export * from './subscription';
export * from './payment';
export * from './billing';
export * from './storage';
export * from './settings';
export * from './notification';
export * from './analytics';
export * from './admin';
export * from './api';
export * from './database';

// Common utility types
export type WithId<T> = T & { _id: ObjectId };
export type WithoutId<T> = Omit<T, '_id' | 'createdAt' | 'updatedAt'>;
export type CreateRequest<T> = WithoutId<T>;
export type UpdateRequest<T> = Partial<WithoutId<T>>;

// API Request/Response types
export type GetRequest = {
  id?: ObjectId;
  params?: Record<string, any>;
  query?: Record<string, any>;
};

export type ListRequest = PaginationParams & {
  filters?: Record<string, any>;
  search?: string;
};

export type CreateResponse<T> = BaseResponse<T>;
export type UpdateResponse<T> = BaseResponse<T>;
export type DeleteResponse = BaseResponse<{ deleted: boolean }>;
export type ListResponse<T> = PaginatedResponse<T>;

// Form validation types
export type ValidationRule = {
  required?: boolean;
  minLength?: number;
  maxLength?: number;
  pattern?: RegExp;
  custom?: (value: any) => boolean | string;
};

export type FormValidation<T> = {
  [K in keyof T]?: ValidationRule;
};

// Event types for real-time updates
export type EventType = 
  | 'file.uploaded'
  | 'file.deleted'
  | 'file.shared'
  | 'folder.created'
  | 'share.accessed'
  | 'user.login'
  | 'team.member.added'
  | 'subscription.updated'
  | 'payment.succeeded'
  | 'storage.limit.reached';

export interface Event<T = any> {
  type: EventType;
  userId: ObjectId;
  teamId?: ObjectId;
  data: T;
  timestamp: Date;
  metadata?: Record<string, any>;
}