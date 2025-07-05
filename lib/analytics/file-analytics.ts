/**
 * File Analytics - File usage, types, and storage metrics
 */

import { ObjectId, AnalyticsDateRange } from '../../types';

export interface FileMetrics {
  totalFiles: number;
  totalSize: number;
  uploadsToday: number;
  downloadsToday: number;
  sharesTotal: number;
  averageFileSize: number;
  storageGrowthTrend: { date: string; size: number }[];
  fileTypeDistribution: { type: string; count: number; size: number }[];
  topFiles: {
    id: ObjectId;
    name: string;
    size: number;
    downloads: number;
    shares: number;
  }[];
}

export interface FileUsageStats {
  uploadStats: {
    total: number;
    daily: number;
    weekly: number;
    monthly: number;
  };
  downloadStats: {
    total: number;
    daily: number;
    weekly: number;
    monthly: number;
  };
  shareStats: {
    total: number;
    daily: number;
    weekly: number;
    monthly: number;
  };
  popularFiles: {
    mostDownloaded: { id: ObjectId; name: string; downloads: number }[];
    mostShared: { id: ObjectId; name: string; shares: number }[];
    largest: { id: ObjectId; name: string; size: number }[];
  };
}

export interface StorageAnalytics {
  totalStorage: number;
  usedStorage: number;
  availableStorage: number;
  storageByProvider: { provider: string; used: number; files: number }[];
  storageByFileType: { type: string; size: number; percentage: number }[];
  storageGrowth: { date: string; total: number; delta: number }[];
  projectedStorage: { date: string; projected: number }[];
}

export class FileAnalytics {
  /**
   * Get file overview metrics
   */
  static async getFileMetrics(dateRange: AnalyticsDateRange): Promise<FileMetrics> {
    try {
      const response = await fetch('/api/admin/analytics/files', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ dateRange })
      });

      const result = await response.json();

      if (!response.ok) {
        throw new Error(result.error || 'Failed to fetch file metrics');
      }

      return result.data;
    } catch (error) {
      console.error('Error fetching file metrics:', error);
      throw error;
    }
  }

  /**
   * Get detailed file usage statistics
   */
  static async getFileUsageStats(dateRange: AnalyticsDateRange): Promise<FileUsageStats> {
    try {
      const response = await fetch('/api/admin/analytics/file-usage', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ dateRange })
      });

      const result = await response.json();

      if (!response.ok) {
        throw new Error(result.error || 'Failed to fetch file usage stats');
      }

      return result.data;
    } catch (error) {
      console.error('Error fetching file usage stats:', error);
      throw error;
    }
  }

  /**
   * Get storage analytics
   */
  static async getStorageAnalytics(dateRange: AnalyticsDateRange): Promise<StorageAnalytics> {
    try {
      const response = await fetch('/api/admin/analytics/storage', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ dateRange })
      });

      const result = await response.json();

      if (!response.ok) {
        throw new Error(result.error || 'Failed to fetch storage analytics');
      }

      return result.data;
    } catch (error) {
      console.error('Error fetching storage analytics:', error);
      throw error;
    }
  }

  /**
   * Get file type analysis
   */
  static async getFileTypeAnalysis(dateRange: AnalyticsDateRange): Promise<{
    distribution: { type: string; count: number; size: number; percentage: number }[];
    trends: { type: string; data: { date: string; count: number }[] }[];
  }> {
    try {
      const response = await fetch('/api/admin/analytics/file-types', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ dateRange })
      });

      const result = await response.json();

      if (!response.ok) {
        throw new Error(result.error || 'Failed to fetch file type analysis');
      }

      return result.data;
    } catch (error) {
      console.error('Error fetching file type analysis:', error);
      throw error;
    }
  }

  /**
   * Get storage optimization suggestions
   */
  static async getStorageOptimization(): Promise<{
    duplicateFiles: { id: ObjectId; name: string; size: number; duplicateCount: number }[];
    largeFiles: { id: ObjectId; name: string; size: number; lastAccessed: Date }[];
    unusedFiles: { id: ObjectId; name: string; size: number; lastAccessed: Date }[];
    potentialSavings: number;
  }> {
    try {
      const response = await fetch('/api/admin/analytics/storage-optimization');
      const result = await response.json();
      
      if (!response.ok) {
        throw new Error(result.error || 'Failed to fetch storage optimization data');
      }

      return result.data;
    } catch (error) {
      console.error('Error fetching storage optimization:', error);
      throw error;
    }
  }

  /**
   * Get bandwidth usage analytics
   */
  static async getBandwidthAnalytics(dateRange: AnalyticsDateRange): Promise<{
    totalBandwidth: number;
    uploadBandwidth: number;
    downloadBandwidth: number;
    bandwidthTrend: { date: string; upload: number; download: number }[];
    topConsumers: { userId: ObjectId; username: string; bandwidth: number }[];
  }> {
    try {
      const response = await fetch('/api/admin/analytics/bandwidth', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ dateRange })
      });

      const result = await response.json();

      if (!response.ok) {
        throw new Error(result.error || 'Failed to fetch bandwidth analytics');
      }

      return result.data;
    } catch (error) {
      console.error('Error fetching bandwidth analytics:', error);
      throw error;
    }
  }
}