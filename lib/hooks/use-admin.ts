// lib/hooks/use-admin.ts
import { useAdminStore } from '@/store/admin-store';

/**
 * Simple hook that re-exports admin store functionality
 * Maintains consistency with existing store patterns
 */
export const useAdmin = () => {
  const store = useAdminStore();

  return {
    // State
    users: store.users,
    teams: store.teams,
    analytics: store.analytics,
    systemStats: store.systemStats,
    isLoading: store.isLoading,
    error: store.error,
    pagination: store.pagination,

    // Actions
    loadUsers: store.loadUsers,
    loadTeams: store.loadTeams,
    loadAnalytics: store.loadAnalytics,
    loadSystemStats: store.loadSystemStats,
    banUser: store.banUser,
    unbanUser: store.unbanUser,
    updateUserQuota: store.updateUserQuota,
    impersonateUser: store.impersonateUser,
    clearError: store.clearError,
  };
};