// lib/hooks/use-usage-tracking.ts
import { useState, useEffect, useCallback } from 'react';
import { useAuth } from './use-auth';
import type { ObjectId, BaseResponse, UsageRecord } from '@/types';

interface UsageStats {
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
}

interface UsageState {
  stats: UsageStats | null;
  history: UsageRecord[];
  isLoading: boolean;
  error: string | null;
  lastUpdated: Date | null;
}

export const useUsageTracking = () => {
  const { user } = useAuth();
  const [state, setState] = useState<UsageState>({
    stats: null,
    history: [],
    isLoading: false,
    error: null,
    lastUpdated: null,
  });

  // Load current usage stats
  const loadStats = useCallback(async () => {
    if (!user) return;

    setState(prev => ({ ...prev, isLoading: true, error: null }));

    try {
      const response = await fetch('/api/client/usage/stats');
      const result: BaseResponse<UsageStats> = await response.json();

      if (!response.ok) {
        throw new Error(result.error || 'Failed to load usage stats');
      }

      setState(prev => ({
        ...prev,
        stats: result.data,
        isLoading: false,
        lastUpdated: new Date(),
      }));
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to load usage stats',
        isLoading: false,
      }));
    }
  }, [user]);

  // Load usage history
  const loadHistory = useCallback(async (period: 'daily' | 'monthly' | 'yearly' = 'monthly') => {
    if (!user) return;

    setState(prev => ({ ...prev, isLoading: true, error: null }));

    try {
      const response = await fetch(`/api/client/usage/history?period=${period}`);
      const result: BaseResponse<UsageRecord[]> = await response.json();

      if (!response.ok) {
        throw new Error(result.error || 'Failed to load usage history');
      }

      setState(prev => ({
        ...prev,
        history: result.data,
        isLoading: false,
      }));
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to load usage history',
        isLoading: false,
      }));
    }
  }, [user]);

  // Track specific usage event
  const trackUsage = useCallback(async (event: {
    type: 'upload' | 'download' | 'api_request' | 'share_created';
    metadata?: Record<string, any>;
  }) => {
    if (!user) return;

    try {
      await fetch('/api/client/usage/track', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(event),
      });

      // Optionally refresh stats after tracking
      // loadStats();
    } catch (error) {
      console.warn('Failed to track usage:', error);
      // Don't throw here as tracking is non-critical
    }
  }, [user]);

  // Check if usage limit is exceeded for a specific metric
  const isLimitExceeded = useCallback((metric: keyof UsageStats['current']) => {
    if (!state.stats) return false;
    
    const currentValue = state.stats.current[metric];
    const limitKey = `${metric.replace('Used', 'Limit')}` as keyof UsageStats['limits'];
    const limitValue = state.stats.limits[limitKey as any];
    
    return currentValue >= limitValue;
  }, [state.stats]);

  // Get usage percentage for a specific metric
  const getUsagePercentage = useCallback((metric: keyof UsageStats['current']) => {
    if (!state.stats) return 0;
    
    const currentValue = state.stats.current[metric];
    const limitKey = `${metric.replace('Used', 'Limit')}` as keyof UsageStats['limits'];
    const limitValue = state.stats.limits[limitKey as any];
    
    if (limitValue === 0) return 0;
    return Math.min(100, (currentValue / limitValue) * 100);
  }, [state.stats]);

  // Format bytes helper
  const formatBytes = useCallback((bytes: number) => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  }, []);

  // Clear error
  const clearError = useCallback(() => {
    setState(prev => ({ ...prev, error: null }));
  }, []);

  // Auto-load stats when user is available
  useEffect(() => {
    if (user) {
      loadStats();
    }
  }, [user, loadStats]);

  // Auto-refresh stats every 5 minutes
  useEffect(() => {
    if (!user) return;

    const interval = setInterval(() => {
      loadStats();
    }, 5 * 60 * 1000); // 5 minutes

    return () => clearInterval(interval);
  }, [user, loadStats]);

  return {
    // State
    ...state,

    // Actions
    loadStats,
    loadHistory,
    trackUsage,
    clearError,

    // Helpers
    isLimitExceeded,
    getUsagePercentage,
    formatBytes,
  };
};