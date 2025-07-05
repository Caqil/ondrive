import { ObjectId, User } from ".";

export interface AdminDashboardStats {
  users: {
    total: number;
    active: number;
    new: number;
    banned: number;
  };
  files: {
    total: number;
    totalSize: number;
    uploads: number;
    downloads: number;
  };
  storage: {
    used: number;
    available: number;
    providers: Record<string, { used: number; files: number }>;
  };
  revenue: {
    total: number;
    monthly: number;
    growth: number;
  };
  subscriptions: {
    active: number;
    trial: number;
    cancelled: number;
    churn: number;
  };
}

export interface AdminUser extends User {
  fileCount: number;
  storagePercent: number;
  lastActivity?: Date;
  subscriptionPlan?: string;
  totalRevenue: number;
}

export interface AdminActionRequest {
  userId: ObjectId;
  action: 'ban' | 'unban' | 'verify_email' | 'reset_password' | 'update_quota' | 'impersonate';
  reason?: string;
  data?: Record<string, any>;
}

export interface SystemHealth {
  status: 'healthy' | 'warning' | 'critical';
  services: {
    database: 'healthy' | 'warning' | 'critical';
    storage: 'healthy' | 'warning' | 'critical';
    email: 'healthy' | 'warning' | 'critical';
    payment: 'healthy' | 'warning' | 'critical';
  };
  metrics: {
    uptime: number;
    responseTime: number;
    errorRate: number;
    memoryUsage: number;
    cpuUsage: number;
  };
}
