/**
 * Performance Analytics - System performance and health metrics
 */

import { AnalyticsDateRange } from '../../types';

export interface PerformanceMetrics {
  responseTime: {
    average: number;
    p50: number;
    p95: number;
    p99: number;
  };
  throughput: {
    requestsPerSecond: number;
    uploadsPerMinute: number;
    downloadsPerMinute: number;
  };
  errorRates: {
    total: number;
    api: number;
    upload: number;
    download: number;
  };
  systemHealth: {
    cpuUsage: number;
    memoryUsage: number;
    diskUsage: number;
    networkUsage: number;
  };
  availability: {
    uptime: number;
    downtimeEvents: { start: Date; end: Date; reason?: string }[];
  };
}

export interface APIPerformance {
  endpoint: string;
  method: string;
  averageResponseTime: number;
  requestCount: number;
  errorRate: number;
  slowestRequests: {
    timestamp: Date;
    responseTime: number;
    userId?: string;
  }[];
}

export interface DatabasePerformance {
  queryPerformance: {
    averageQueryTime: number;
    slowQueries: { query: string; time: number; count: number }[];
  };
  connectionPool: {
    activeConnections: number;
    maxConnections: number;
    utilizationPercent: number;
  };
  indexEfficiency: {
    indexName: string;
    collection: string;
    efficiency: number;
    suggestions: string[];
  }[];
}

export interface StorageProviderPerformance {
  provider: string;
  averageUploadTime: number;
  averageDownloadTime: number;
  errorRate: number;
  throughput: number;
  healthStatus: 'healthy' | 'warning' | 'error';
  latencyTrend: { timestamp: Date; latency: number }[];
}

export class PerformanceAnalytics {
  /**
   * Get overall system performance metrics
   */
  static async getPerformanceMetrics(dateRange: AnalyticsDateRange): Promise<PerformanceMetrics> {
    try {
      const response = await fetch('/api/admin/monitoring/performance', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ dateRange })
      });

      const result = await response.json();

      if (!response.ok) {
        throw new Error(result.error || 'Failed to fetch performance metrics');
      }

      return result.data;
    } catch (error) {
      console.error('Error fetching performance metrics:', error);
      throw error;
    }
  }

  /**
   * Get API endpoint performance analysis
   */
  static async getAPIPerformance(dateRange: AnalyticsDateRange): Promise<APIPerformance[]> {
    try {
      const response = await fetch('/api/admin/monitoring/api-performance', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ dateRange })
      });

      const result = await response.json();

      if (!response.ok) {
        throw new Error(result.error || 'Failed to fetch API performance');
      }

      return result.data;
    } catch (error) {
      console.error('Error fetching API performance:', error);
      throw error;
    }
  }

  /**
   * Get database performance metrics
   */
  static async getDatabasePerformance(dateRange: AnalyticsDateRange): Promise<DatabasePerformance> {
    try {
      const response = await fetch('/api/admin/monitoring/database-performance', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ dateRange })
      });

      const result = await response.json();

      if (!response.ok) {
        throw new Error(result.error || 'Failed to fetch database performance');
      }

      return result.data;
    } catch (error) {
      console.error('Error fetching database performance:', error);
      throw error;
    }
  }

  /**
   * Get storage provider performance
   */
  static async getStorageProviderPerformance(
    dateRange: AnalyticsDateRange
  ): Promise<StorageProviderPerformance[]> {
    try {
      const response = await fetch('/api/admin/monitoring/storage-performance', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ dateRange })
      });

      const result = await response.json();

      if (!response.ok) {
        throw new Error(result.error || 'Failed to fetch storage provider performance');
      }

      return result.data;
    } catch (error) {
      console.error('Error fetching storage provider performance:', error);
      throw error;
    }
  }

  /**
   * Get real-time system status
   */
  static async getSystemStatus(): Promise<{
    status: 'healthy' | 'warning' | 'critical';
    services: {
      database: 'healthy' | 'warning' | 'critical';
      storage: 'healthy' | 'warning' | 'critical';
      email: 'healthy' | 'warning' | 'critical';
      payment: 'healthy' | 'warning' | 'critical';
    };
    alerts: {
      level: 'info' | 'warning' | 'critical';
      message: string;
      timestamp: Date;
    }[];
  }> {
    try {
      const response = await fetch('/api/admin/monitoring/health');
      const result = await response.json();
      
      if (!response.ok) {
        throw new Error(result.error || 'Failed to fetch system status');
      }

      return result.data;
    } catch (error) {
      console.error('Error fetching system status:', error);
      throw error;
    }
  }

  /**
   * Get error analysis
   */
  static async getErrorAnalysis(dateRange: AnalyticsDateRange): Promise<{
    errorsByType: { type: string; count: number; percentage: number }[];
    errorsByEndpoint: { endpoint: string; count: number; errorRate: number }[];
    recentErrors: {
      timestamp: Date;
      type: string;
      message: string;
      endpoint: string;
      userId?: string;
    }[];
    errorTrends: { date: string; count: number }[];
  }> {
    try {
      const response = await fetch('/api/admin/monitoring/errors', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ dateRange })
      });

      const result = await response.json();

      if (!response.ok) {
        throw new Error(result.error || 'Failed to fetch error analysis');
      }

      return result.data;
    } catch (error) {
      console.error('Error fetching error analysis:', error);
      throw error;
    }
  }

  /**
   * Get performance optimization suggestions
   */
  static async getOptimizationSuggestions(): Promise<{
    category: 'database' | 'api' | 'storage' | 'caching';
    priority: 'high' | 'medium' | 'low';
    title: string;
    description: string;
    estimatedImpact: string;
    implementation: string;
  }[]> {
    try {
      const response = await fetch('/api/admin/monitoring/optimization-suggestions');
      const result = await response.json();
      
      if (!response.ok) {
        throw new Error(result.error || 'Failed to fetch optimization suggestions');
      }

      return result.data;
    } catch (error) {
      console.error('Error fetching optimization suggestions:', error);
      throw error;
    }
  }
}