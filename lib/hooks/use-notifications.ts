import { useNotificationStore } from '@/store/notification-store';

// Selectors - following your pattern
export const useNotificationsList = () => useNotificationStore((state) => state.notifications);
export const useUnreadCount = () => useNotificationStore((state) => state.unreadCount);
export const useNotificationsLoading = () => useNotificationStore((state) => state.isLoading);
export const useNotificationsError = () => useNotificationStore((state) => state.error);

// Actions - following your pattern
export const useNotificationActions = () => {
  const store = useNotificationStore();
  return {
    loadNotifications: store.loadNotifications,
    markAsRead: store.markAsRead,
    markAllAsRead: store.markAllAsRead,
    deleteNotification: store.deleteNotification,
    clearAllNotifications: store.clearAllNotifications,
    setShowOnlyUnread: store.setShowOnlyUnread,
    setNotificationTypes: store.setNotificationTypes,
    clearError: store.clearError,
  };
};

// Main convenience hook
export const useNotifications = () => {
  const store = useNotificationStore();
  const actions = useNotificationActions();

  return {
    // State
    notifications: store.notifications,
    unreadCount: store.unreadCount,
    isLoading: store.isLoading,
    error: store.error,
    showOnlyUnread: store.showOnlyUnread,
    notificationTypes: store.notificationTypes,

    // Actions
    ...actions,
  };
};