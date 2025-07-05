import { useAuthStore, useAuthActions } from '@/store/auth-store';

/**
 * Simple hook that re-exports auth store functionality
 * Maintains consistency with existing store patterns
 */
export const useAuth = () => {
  const store = useAuthStore();
  const actions = useAuthActions();

  return {
    // State
    user: store.user,
    session: store.session,
    isAuthenticated: store.isAuthenticated,
    isLoading: store.isLoading,
    isInitialized: store.isInitialized,
    error: store.error,
    is2FARequired: store.is2FARequired,
    pendingSessionToken: store.pendingSessionToken,
    isResettingPassword: store.isResettingPassword,
    resetToken: store.resetToken,

    // Actions
    ...actions,
  };
};