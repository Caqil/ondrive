/**
 * Usage Analytics - Platform usage and user behavior tracking
 */

import { ObjectId, AnalyticsDateRange, UsageRecord } from '../../types';

export interface UsageMetrics {
  totalUsers: number;
  activeUsers: {
    daily: number;
    weekly: number;
    monthly: number;
  };
  sessionMetrics: {
    averageDuration: number;
    averagePagesPerSession: number;
    bounceRate: number;
  };
  featureUsage: {
    feature: string;
    users: number;
    usage: number;
    adoption: number;
  }[];
  deviceStats: {
    desktop: number;
    mobile: number;
    tablet: number;
  };
  browserStats: {
    browser: string;
    users: number;
    percentage: number;
  }[];
}

export interface FeatureAnalytics {
  featureName: string;
  totalUsage: number;
  uniqueUsers: number;
  adoptionRate: number;
  usageTrend: { date: string; usage: number }[];
  userSegments: {
    segment: string;
    users: number;
    usage: number;
  }[];
}

export interface StorageUsageAnalytics {
  totalStorage: number;
  usedStorage: number;
  storageByUser: {
    userId: ObjectId;
    username: string;
    storageUsed: number;
    percentage: number;
  }[];
  storageByFileType: {
    type: string;
    size: number;
    count: number;
    percentage: number;
  }[];
  storageGrowth: {
    date: string;
    storage: number;
    files: number;
  }[];
}

export class UsageAnalytics {
  /**
   * Get platform usage metrics
   */
  static async getUsageMetrics(dateRange: AnalyticsDateRange): Promise<UsageMetrics> {
    try {
      const response = await fetch('/api/admin/analytics/usage', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ dateRange })
      });

      const result = await response.json();

      if (!response.ok) {
        throw new Error(result.error || 'Failed to fetch usage metrics');
      }

      return result.data;
    } catch (error) {
      console.error('Error fetching usage metrics:', error);
      throw error;
    }
  }

  /**
   * Get feature usage analytics
   */
  static async getFeatureAnalytics(
    feature: string,
    dateRange: AnalyticsDateRange
  ): Promise<FeatureAnalytics> {
    try {
      const response = await fetch('/api/admin/analytics/feature-usage', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ feature, dateRange })
      });

      const result = await response.json();

      if (!response.ok) {
        throw new Error(result.error || 'Failed to fetch feature analytics');
      }

      return result.data;
    } catch (error) {
      console.error('Error fetching feature analytics:', error);
      throw error;
    }
  }

  /**
   * Get storage usage analytics
   */
  static async getStorageUsageAnalytics(dateRange: AnalyticsDateRange): Promise<StorageUsageAnalytics> {
    try {
      const response = await fetch('/api/admin/analytics/storage-usage', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ dateRange })
      });

      const result = await response.json();

      if (!response.ok) {
        throw new Error(result.error || 'Failed to fetch storage usage analytics');
      }

      return result.data;
    } catch (error) {
      console.error('Error fetching storage usage analytics:', error);
      throw error;
    }
  }

  /**
   * Get user usage records (client endpoint)
   */
  static async getUserUsageRecords(period: 'daily' | 'monthly' | 'yearly' = 'monthly'): Promise<UsageRecord[]> {
    try {
      const response = await fetch(`/api/client/usage/history?period=${period}`);
      const result = await response.json();

      if (!response.ok) {
        throw new Error(result.error || 'Failed to fetch usage records');
      }

      return result.data;
    } catch (error) {
      console.error('Error fetching usage records:', error);
      throw error;
    }
  }

  /**
   * Get current user usage stats (client endpoint)
   */
  static async getCurrentUsageStats(): Promise<{
    current: {
      storageUsed: number;
      bandwidthUsed: number;
      apiRequestsUsed: number;
      fileUploadsCount: number;
      fileDownloadsCount: number;
      shareLinksCreated: number;
    };
    limits: {
      storageLimit: number;
      bandwidthLimit: number;
      apiRequestLimit: number;
      fileUploadLimit: number;
      shareLinksLimit: number;
    };
    period: {
      start: Date;
      end: Date;
    };
  }> {
    try {
      const response = await fetch('/api/client/usage/stats');
      const result = await response.json();

      if (!response.ok) {
        throw new Error(result.error || 'Failed to fetch current usage stats');
      }

      return result.data;
    } catch (error) {
      console.error('Error fetching current usage stats:', error);
      throw error;
    }
  }

  /**
   * Track usage event (client endpoint)
   */
  static async trackUsage(event: {
    type: 'upload' | 'download' | 'api_request' | 'share_created';
    metadata?: Record<string, any>;
  }): Promise<void> {
    try {
      const response = await fetch('/api/client/usage/track', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(event)
      });

      const result = await response.json();

      if (!response.ok) {
        throw new Error(result.error || 'Failed to track usage');
      }
    } catch (error) {
      console.error('Error tracking usage:', error);
      // Don't throw here as tracking is non-critical
    }
  }

  /**
   * Get API usage analytics
   */
  static async getAPIUsageAnalytics(dateRange: AnalyticsDateRange): Promise<{
    totalRequests: number;
    requestsByEndpoint: { endpoint: string; count: number; percentage: number }[];
    requestsByUser: { userId: ObjectId; username: string; requests: number }[];
    requestTrends: { date: string; requests: number }[];
    averageResponseTime: number;
    errorRate: number;
  }> {
    try {
      const response = await fetch('/api/admin/analytics/api-usage', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ dateRange })
      });

      const result = await response.json();

      if (!response.ok) {
        throw new Error(result.error || 'Failed to fetch API usage analytics');
      }

      return result.data;
    } catch (error) {
      console.error('Error fetching API usage analytics:', error);
      throw error;
    }
  }

  /**
   * Get user engagement analytics
   */
  static async getUserEngagement(dateRange: AnalyticsDateRange): Promise<{
    dailyActiveUsers: { date: string; users: number }[];
    weeklyActiveUsers: { date: string; users: number }[];
    monthlyActiveUsers: { date: string; users: number }[];
    sessionDuration: { average: number; median: number; p95: number }[];
    retentionCohorts: {
      cohort: string;
      size: number;
      retention: number[];
    }[];
    engagementScore: number;
  }> {
    try {
      const response = await fetch('/api/admin/analytics/user-engagement', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ dateRange })
      });

      const result = await response.json();

      if (!response.ok) {
        throw new Error(result.error || 'Failed to fetch user engagement');
      }

      return result.data;
    } catch (error) {
      console.error('Error fetching user engagement:', error);
      throw error;
    }
  }
}