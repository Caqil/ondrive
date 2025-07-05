import { useState, useEffect, useCallback } from 'react';
import { useAuth } from './use-auth';
import type { BaseResponse } from '@/types';

interface StorageBreakdown {
  images: number;
  videos: number;
  documents: number;
  archives: number;
  others: number;
}

interface StorageStats {
  used: number;
  quota: number;
  percentage: number;
  breakdown: StorageBreakdown;
}

interface StorageState {
  stats: StorageStats | null;
  isLoading: boolean;
  error: string | null;
  lastUpdated: Date | null;
}

// Following your pattern - this is a standalone hook that doesn't need a store
export const useStorage = () => {
  const { user } = useAuth();
  const [state, setState] = useState<StorageState>({
    stats: null,
    isLoading: false,
    error: null,
    lastUpdated: null,
  });

  const loadStats = useCallback(async () => {
    if (!user) return;

    setState(prev => ({ ...prev, isLoading: true, error: null }));

    try {
      const response = await fetch('/api/client/storage/stats');
      const result: BaseResponse<StorageStats> = await response.json();

      if (!response.ok) {
        throw new Error(result.error || 'Failed to load storage stats');
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
        error: error instanceof Error ? error.message : 'Failed to load storage stats',
        isLoading: false,
      }));
    }
  }, [user]);

  const formatBytes = useCallback((bytes: number) => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  }, []);

  const clearError = useCallback(() => {
    setState(prev => ({ ...prev, error: null }));
  }, []);

  // Computed values
  const isQuotaExceeded = state.stats ? state.stats.percentage >= 100 : false;
  const isNearQuota = state.stats ? state.stats.percentage >= 90 : false;
  const availableStorage = state.stats ? state.stats.quota - state.stats.used : 0;

  useEffect(() => {
    if (user) {
      loadStats();
    }
  }, [user, loadStats]);

  return {
    // State
    stats: state.stats,
    isLoading: state.isLoading,
    error: state.error,
    lastUpdated: state.lastUpdated,
    isQuotaExceeded,
    isNearQuota,
    availableStorage,

    // Actions
    loadStats,
    clearError,

    // Helpers
    formatBytes,
  };
};