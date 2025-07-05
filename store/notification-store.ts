import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { immer } from 'zustand/middleware/immer';
import type { 
  Notification,
  CreateNotificationRequest,
  ObjectId 
} from '@/types';

interface NotificationState {
  // Notifications
  notifications: Notification[];
  unreadCount: number;
  
  // Loading
  isLoading: boolean;
  error: string | null;
  
  // Filters
  showOnlyUnread: boolean;
  notificationTypes: string[];
  
  // Actions
  loadNotifications: () => Promise<void>;
  markAsRead: (notificationId: ObjectId) => Promise<void>;
  markAllAsRead: () => Promise<void>;
  deleteNotification: (notificationId: ObjectId) => Promise<void>;
  clearAllNotifications: () => Promise<void>;
  setShowOnlyUnread: (show: boolean) => void;
  setNotificationTypes: (types: string[]) => void;
  clearError: () => void;
}

export const useNotificationStore = create<NotificationState>()(
  devtools(
    immer((set, get) => ({
      // Initial State
      notifications: [],
      unreadCount: 0,
      isLoading: false,
      error: null,
      showOnlyUnread: false,
      notificationTypes: [],

      // Load Notifications
      loadNotifications: async () => {
        set((state) => {
          state.isLoading = true;
          state.error = null;
        });

        try {
          const { showOnlyUnread, notificationTypes } = get();
          const searchParams = new URLSearchParams();
          
          if (showOnlyUnread) {
            searchParams.append('unread', 'true');
          }
          
          if (notificationTypes.length > 0) {
            searchParams.append('types', notificationTypes.join(','));
          }

          const response = await fetch(`/api/client/notifications?${searchParams}`);
          const result = await response.json();

          if (!response.ok) {
            throw new Error(result.error || 'Failed to load notifications');
          }

          set((state) => {
            state.notifications = result.data;
            state.unreadCount = result.data.filter((n: Notification) => !n.isRead).length;
            state.isLoading = false;
          });

        } catch (error) {
          set((state) => {
            state.error = error instanceof Error ? error.message : 'Failed to load notifications';
            state.isLoading = false;
          });
        }
      },

      // Mark as Read
      markAsRead: async (notificationId: ObjectId) => {
        try {
          const response = await fetch(`/api/client/notifications/${notificationId}/read`, {
            method: 'POST',
          });

          const result = await response.json();

          if (!response.ok) {
            throw new Error(result.error || 'Failed to mark notification as read');
          }

          set((state) => {
            const notificationIndex = state.notifications.findIndex(n => n._id === notificationId);
            if (notificationIndex !== -1 && !state.notifications[notificationIndex].isRead) {
              state.notifications[notificationIndex].isRead = true;
              state.notifications[notificationIndex].readAt = new Date();
              state.unreadCount = Math.max(0, state.unreadCount - 1);
            }
          });

        } catch (error) {
          set((state) => {
            state.error = error instanceof Error ? error.message : 'Failed to mark as read';
          });
        }
      },

      // Mark All as Read
      markAllAsRead: async () => {
        try {
          const response = await fetch('/api/client/notifications/mark-read', {
            method: 'POST',
          });

          const result = await response.json();

          if (!response.ok) {
            throw new Error(result.error || 'Failed to mark all notifications as read');
          }

          set((state) => {
            state.notifications.forEach(notification => {
              if (!notification.isRead) {
                notification.isRead = true;
                notification.readAt = new Date();
              }
            });
            state.unreadCount = 0;
          });

        } catch (error) {
          set((state) => {
            state.error = error instanceof Error ? error.message : 'Failed to mark all as read';
          });
        }
      },

      // Delete Notification
      deleteNotification: async (notificationId: ObjectId) => {
        try {
          const response = await fetch(`/api/client/notifications/${notificationId}`, {
            method: 'DELETE',
          });

          const result = await response.json();

          if (!response.ok) {
            throw new Error(result.error || 'Failed to delete notification');
          }

          set((state) => {
            const notificationIndex = state.notifications.findIndex(n => n._id === notificationId);
            if (notificationIndex !== -1) {
              const wasUnread = !state.notifications[notificationIndex].isRead;
              state.notifications.splice(notificationIndex, 1);
              if (wasUnread) {
                state.unreadCount = Math.max(0, state.unreadCount - 1);
              }
            }
          });

        } catch (error) {
          set((state) => {
            state.error = error instanceof Error ? error.message : 'Failed to delete notification';
          });
        }
      },

      // Clear All Notifications
      clearAllNotifications: async () => {
        try {
          const response = await fetch('/api/client/notifications', {
            method: 'DELETE',
          });

          const result = await response.json();

          if (!response.ok) {
            throw new Error(result.error || 'Failed to clear all notifications');
          }

          set((state) => {
            state.notifications = [];
            state.unreadCount = 0;
          });

        } catch (error) {
          set((state) => {
            state.error = error instanceof Error ? error.message : 'Failed to clear notifications';
          });
        }
      },

      // Set Show Only Unread
      setShowOnlyUnread: (show: boolean) => {
        set((state) => {
          state.showOnlyUnread = show;
        });
        get().loadNotifications();
      },

      // Set Notification Types
      setNotificationTypes: (types: string[]) => {
        set((state) => {
          state.notificationTypes = types;
        });
        get().loadNotifications();
      },

      // Clear Error
      clearError: () => {
        set((state) => {
          state.error = null;
        });
      },
    })),
    { name: 'notification-store' }
  )
);