// lib/hooks/use-storage.ts
import { useState, useEffect, useCallback } from 'react';
import { useAuth } from './use-auth';
import type { ObjectId, BaseResponse } from '@/types';

interface StorageStats {
  used: number;
  quota: number;
  percentage: number;
  breakdown: {
    images: number;
    videos: number;
    documents: number;
    archives: number;
    others: number;
  };
}

interface StorageState {
  stats: StorageStats | null;
  isLoading: boolean;
  error: string | null;
}

export const useStorage = () => {
  const { user } = useAuth();
  const [state, setState] = useState<StorageState>({
    stats: null,
    isLoading: false,
    error: null,
  });

  // Load storage statistics
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
        stats: result.data,
        isLoading: false,
      }));
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to load storage stats',
        isLoading: false,
      }));
    }
  }, [user]);

  // Check if storage quota is exceeded
  const isQuotaExceeded = state.stats ? state.stats.percentage >= 100 : false;

  // Check if storage is near quota (>90%)
  const isNearQuota = state.stats ? state.stats.percentage >= 90 : false;

  // Get available storage
  const availableStorage = state.stats ? state.stats.quota - state.stats.used : 0;

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

  return {
    // State
    ...state,
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