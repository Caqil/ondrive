/**
 * Customer Analytics - User behavior and engagement metrics
 */

import { ObjectId, AnalyticsDateRange } from '../../types';

export interface CustomerMetrics {
  totalUsers: number;
  activeUsers: number;
  newUsers: number;
  churnedUsers: number;
  retentionRate: number;
  averageSessionDuration: number;
  topCountries: { country: string; count: number }[];
  userGrowthTrend: { date: string; count: number }[];
  engagementScore: number;
}

export interface CustomerSegment {
  id: string;
  name: string;
  criteria: Record<string, any>;
  userCount: number;
  conversionRate: number;
  averageValue: number;
}

export interface UserActivity {
  userId: ObjectId;
  lastActive: Date;
  sessionCount: number;
  totalFileUploads: number;
  totalFileDownloads: number;
  totalShares: number;
  storageUsed: number;
  signupDate: Date;
  subscriptionStatus: string;
  lifetimeValue: number;
}

export class CustomerAnalytics {
  /**
   * Get customer overview metrics
   */
  static async getCustomerMetrics(dateRange: AnalyticsDateRange): Promise<CustomerMetrics> {
    try {
      const response = await fetch('/api/admin/analytics/customers', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ dateRange })
      });

      const result = await response.json();

      if (!response.ok) {
        throw new Error(result.error || 'Failed to fetch customer metrics');
      }

      return result.data;
    } catch (error) {
      console.error('Error fetching customer metrics:', error);
      throw error;
    }
  }

  /**
   * Get user activity patterns
   */
  static async getUserActivity(
    userId?: ObjectId,
    dateRange?: AnalyticsDateRange
  ): Promise<UserActivity[]> {
    try {
      const params = new URLSearchParams();
      if (userId) params.append('userId', userId);
      if (dateRange) {
        params.append('startDate', dateRange.start.toISOString());
        params.append('endDate', dateRange.end.toISOString());
      }

      const response = await fetch(`/api/admin/analytics/user-activity?${params}`);
      const result = await response.json();
      
      if (!response.ok) {
        throw new Error(result.error || 'Failed to fetch user activity');
      }

      return result.data;
    } catch (error) {
      console.error('Error fetching user activity:', error);
      throw error;
    }
  }

  /**
   * Get customer segments
   */
  static async getCustomerSegments(): Promise<CustomerSegment[]> {
    try {
      const response = await fetch('/api/admin/analytics/customer-segments');
      const result = await response.json();
      
      if (!response.ok) {
        throw new Error(result.error || 'Failed to fetch customer segments');
      }

      return result.data;
    } catch (error) {
      console.error('Error fetching customer segments:', error);
      throw error;
    }
  }

  /**
   * Calculate customer lifetime value
   */
  static async getCustomerLifetimeValue(userId: ObjectId): Promise<number> {
    try {
      const response = await fetch(`/api/admin/analytics/customer-ltv/${userId}`);
      const result = await response.json();
      
      if (!response.ok) {
        throw new Error(result.error || 'Failed to fetch customer LTV');
      }

      return result.data.lifetimeValue;
    } catch (error) {
      console.error('Error fetching customer LTV:', error);
      throw error;
    }
  }

  /**
   * Get user retention analysis
   */
  static async getRetentionAnalysis(dateRange: AnalyticsDateRange): Promise<{
    cohorts: { cohort: string; retention: number[] }[];
    overallRetention: number;
  }> {
    try {
      const response = await fetch('/api/admin/analytics/retention', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ dateRange })
      });

      const result = await response.json();

      if (!response.ok) {
        throw new Error(result.error || 'Failed to fetch retention analysis');
      }

      return result.data;
    } catch (error) {
      console.error('Error fetching retention analysis:', error);
      throw error;
    }
  }

  /**
   * Get churn analysis
   */
  static async getChurnAnalysis(dateRange: AnalyticsDateRange): Promise<{
    churnRate: number;
    churnReasons: { reason: string; count: number }[];
    riskSegments: CustomerSegment[];
  }> {
    try {
      const response = await fetch('/api/admin/analytics/churn', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ dateRange })
      });

      const result = await response.json();

      if (!response.ok) {
        throw new Error(result.error || 'Failed to fetch churn analysis');
      }

      return result.data;
    } catch (error) {
      console.error('Error fetching churn analysis:', error);
      throw error;
    }
  }
}
