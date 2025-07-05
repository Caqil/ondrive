import { useState, useEffect, useCallback } from 'react';
import { useAuth } from './use-auth';
import type { ObjectId, BaseResponse } from '@/types';

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

interface UsageRecord {
  id: ObjectId;
  userId: ObjectId;
  type: 'upload' | 'download' | 'api_request' | 'share_created';
  amount: number;
  timestamp: Date;
  metadata?: Record<string, any>;
}

interface UsageTrackingEvent {
  type: 'upload' | 'download' | 'api_request' | 'share_created';
  amount?: number;
  metadata?: Record<string, any>;
}

interface UsageState {
  stats: UsageStats | null;
  history: UsageRecord[];
  isLoading: boolean;
  error: string | null;
  lastUpdated: Date | null;
}

// Following your pattern - standalone hook for usage tracking
export const useUsageTracking = () => {
  const { user } = useAuth();
  const [state, setState] = useState<UsageState>({
    stats: null,
    history: [],
    isLoading: false,
    error: null,
    lastUpdated: null,
  });

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
        stats: result.data ?? null,
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
        history: result.data ?? [],
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

  const trackUsage = useCallback(async (event: UsageTrackingEvent) => {
    if (!user) return;

    try {
      const response = await fetch('/api/client/usage/track', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(event),
      });

      const result: BaseResponse<UsageRecord> = await response.json();

      if (!response.ok) {
        throw new Error(result.error || 'Failed to track usage');
      }

      await loadStats();
    } catch (error) {
      console.error('Failed to track usage:', error);
    }
  }, [user, loadStats]);

  const getUsagePercentage = useCallback((type: keyof UsageStats['current']) => {
    if (!state.stats) return 0;
    
    const current = state.stats.current[type];
    const limit = state.stats.limits[`${type.replace('Used', 'Limit')}` as keyof UsageStats['limits']];
    
    return limit > 0 ? Math.min(100, (current / limit) * 100) : 0;
  }, [state.stats]);

  const clearError = useCallback(() => {
    setState(prev => ({ ...prev, error: null }));
  }, []);

  // Computed values
  const hasExceededLimits = state.stats ? Object.keys(state.stats.current).some(key => {
    const currentKey = key as keyof UsageStats['current'];
    const limitKey = key.replace('Used', 'Limit') as keyof UsageStats['limits'];
    return state.stats!.current[currentKey] >= state.stats!.limits[limitKey];
  }) : false;

  useEffect(() => {
    if (user) {
      loadStats();
    }
  }, [user, loadStats]);

  return {
    // State
    stats: state.stats,
    history: state.history,
    isLoading: state.isLoading,
    error: state.error,
    lastUpdated: state.lastUpdated,
    hasExceededLimits,

    // Actions
    loadStats,
    loadHistory,
    trackUsage,
    clearError,

    // Helpers
    getUsagePercentage,
  };
};