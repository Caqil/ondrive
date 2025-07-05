// lib/hooks/use-notifications.ts
import { useNotificationStore } from '@/store/notification-store';

/**
 * Simple hook that re-exports notification store functionality
 * Maintains consistency with existing store patterns
 */
export const useNotifications = () => {
  const store = useNotificationStore();

  return {
    // State
    notifications: store.notifications,
    unreadCount: store.unreadCount,
    isLoading: store.isLoading,
    error: store.error,
    pagination: store.pagination,

    // Actions
    loadNotifications: store.loadNotifications,
    markAsRead: store.markAsRead,
    markAllAsRead: store.markAllAsRead,
    deleteNotification: store.deleteNotification,
    clearError: store.clearError,
  };
};