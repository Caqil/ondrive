import { BaseDocument, ObjectId } from ".";

export interface UsageRecord extends BaseDocument {
  user?: ObjectId;
  team?: ObjectId;
  subscription?: ObjectId;
  storageUsed: number;
  bandwidthUsed: number;
  apiRequestsUsed: number;
  fileUploadsCount: number;
  fileDownloadsCount: number;
  shareLinksCreated: number;
  period: 'daily' | 'monthly' | 'yearly';
  date: Date;
  fileTypeBreakdown: {
    images: { count: number; size: number };
    videos: { count: number; size: number };
    documents: { count: number; size: number };
    archives: { count: number; size: number };
    others: { count: number; size: number };
  };
  costs: {
    storage: number;
    bandwidth: number;
    apiRequests: number;
    total: number;
  };
}

export interface ActivityLog extends BaseDocument {
  user: ObjectId;
  impersonatedBy?: ObjectId;
  action: 'create' | 'read' | 'update' | 'delete' | 'share' | 'download' | 'upload' | 'move' | 'copy' | 'rename' | 'star' | 'unstar' | 'trash' | 'restore';
  resource: ObjectId;
  resourceType: 'file' | 'folder' | 'share' | 'user' | 'team' | 'settings';
  resourceName: string;
  metadata: {
    oldValue?: any;
    newValue?: any;
    fileSize?: number;
    mimeType?: string;
    shareType?: string;
    permission?: string;
    [key: string]: any;
  };
  ip: string;
  userAgent: string;
  location?: {
    country?: string;
    city?: string;
    coordinates?: [number, number];
  };
  team?: ObjectId;
  status: 'success' | 'error' | 'warning';
  errorMessage?: string;
  duration?: number;
}

export interface AnalyticsDateRange {
  start: Date;
  end: Date;
}

export interface AnalyticsMetrics {
  totalUsers: number;
  activeUsers: number;
  totalFiles: number;
  totalStorage: number;
  totalRevenue: number;
  newSignups: number;
  churnRate: number;
  averageSessionDuration: number;
}

export interface AnalyticsQuery {
  dateRange: AnalyticsDateRange;
  metrics: string[];
  groupBy?: 'day' | 'week' | 'month' | 'year';
  filters?: Record<string, any>;
}
