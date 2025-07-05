import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { immer } from 'zustand/middleware/immer';
import type { 
  AdminDashboardStats,
  AdminUser,
  AdminActionRequest,
  SystemHealth,
  Settings,
  UpdateSettingsRequest,
  ObjectId 
} from '@/types';

interface AdminState {
  // Dashboard
  stats: AdminDashboardStats | null;
  systemHealth: SystemHealth | null;
  
  // Users Management
  users: AdminUser[];
  usersLoading: boolean;
  usersError: string | null;
  usersPagination: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
  };
  
  // Settings
  settings: Settings | null;
  settingsLoading: boolean;
  settingsError: string | null;
  
  // General Loading
  isLoading: boolean;
  error: string | null;
  
  // Actions
  loadDashboardStats: () => Promise<void>;
  loadSystemHealth: () => Promise<void>;
  loadUsers: (page?: number, filters?: any) => Promise<void>;
  performUserAction: (action: AdminActionRequest) => Promise<void>;
  loadSettings: () => Promise<void>;
  updateSettings: (updates: UpdateSettingsRequest) => Promise<void>;
  testEmailSettings: () => Promise<void>;
  testStorageSettings: (provider: string) => Promise<void>;
  clearError: () => void;
}

export const useAdminStore = create<AdminState>()(
  devtools(
    immer((set, get) => ({
      // Initial State
      stats: null,
      systemHealth: null,
      users: [],
      usersLoading: false,
      usersError: null,
      usersPagination: {
        page: 1,
        limit: 20,
        total: 0,
        totalPages: 0,
      },
      settings: null,
      settingsLoading: false,
      settingsError: null,
      isLoading: false,
      error: null,

      // Load Dashboard Stats
      loadDashboardStats: async () => {
        set((state) => {
          state.isLoading = true;
          state.error = null;
        });

        try {
          const response = await fetch('/api/admin/dashboard');
          const result = await response.json();

          if (!response.ok) {
            throw new Error(result.error || 'Failed to load dashboard stats');
          }

          set((state) => {
            state.stats = result.data;
            state.isLoading = false;
          });

        } catch (error) {
          set((state) => {
            state.error = error instanceof Error ? error.message : 'Failed to load stats';
            state.isLoading = false;
          });
        }
      },

      // Load System Health
      loadSystemHealth: async () => {
        try {
          const response = await fetch('/api/admin/monitoring/health');
          const result = await response.json();

          if (!response.ok) {
            throw new Error(result.error || 'Failed to load system health');
          }

          set((state) => {
            state.systemHealth = result.data;
          });

        } catch (error) {
          console.warn('Failed to load system health:', error);
        }
      },

      // Load Users
      loadUsers: async (page = 1, filters = {}) => {
        set((state) => {
          state.usersLoading = true;
          state.usersError = null;
        });

        try {
          const searchParams = new URLSearchParams({
            page: page.toString(),
            limit: get().usersPagination.limit.toString(),
            ...Object.fromEntries(
              Object.entries(filters).filter(([_, value]) => value !== undefined && value !== '')
            ),
          });

          const response = await fetch(`/api/admin/users?${searchParams}`);
          const result = await response.json();

          if (!response.ok) {
            throw new Error(result.error || 'Failed to load users');
          }

          set((state) => {
            state.users = result.data;
            state.usersPagination = result.pagination;
            state.usersLoading = false;
          });

        } catch (error) {
          set((state) => {
            state.usersError = error instanceof Error ? error.message : 'Failed to load users';
            state.usersLoading = false;
          });
        }
      },

      // Perform User Action
      performUserAction: async (action: AdminActionRequest) => {
        try {
          const response = await fetch(`/api/admin/users/${action.userId}/${action.action}`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ reason: action.reason, data: action.data }),
          });

          const result = await response.json();

          if (!response.ok) {
            throw new Error(result.error || 'Action failed');
          }

          // Refresh users list
          await get().loadUsers(get().usersPagination.page);

        } catch (error) {
          set((state) => {
            state.error = error instanceof Error ? error.message : 'Action failed';
          });
          throw error;
        }
      },

      // Load Settings
      loadSettings: async () => {
        set((state) => {
          state.settingsLoading = true;
          state.settingsError = null;
        });

        try {
          const response = await fetch('/api/admin/settings');
          const result = await response.json();

          if (!response.ok) {
            throw new Error(result.error || 'Failed to load settings');
          }

          set((state) => {
            state.settings = result.data;
            state.settingsLoading = false;
          });

        } catch (error) {
          set((state) => {
            state.settingsError = error instanceof Error ? error.message : 'Failed to load settings';
            state.settingsLoading = false;
          });
        }
      },

      // Update Settings
      updateSettings: async (updates: UpdateSettingsRequest) => {
        set((state) => {
          state.settingsLoading = true;
          state.settingsError = null;
        });

        try {
          const response = await fetch('/api/admin/settings', {
            method: 'PATCH',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(updates),
          });

          const result = await response.json();

          if (!response.ok) {
            throw new Error(result.error || 'Failed to update settings');
          }

          set((state) => {
            state.settings = result.data;
            state.settingsLoading = false;
          });

        } catch (error) {
          set((state) => {
            state.settingsError = error instanceof Error ? error.message : 'Failed to update settings';
            state.settingsLoading = false;
          });
          throw error;
        }
      },

      // Test Email Settings
      testEmailSettings: async () => {
        try {
          const response = await fetch('/api/admin/settings/smtp/test', {
            method: 'POST',
          });

          const result = await response.json();

          if (!response.ok) {
            throw new Error(result.error || 'Email test failed');
          }

        } catch (error) {
          throw error;
        }
      },

      // Test Storage Settings
      testStorageSettings: async (provider: string) => {
        try {
          const response = await fetch('/api/admin/settings/storage/test', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ provider }),
          });

          const result = await response.json();

          if (!response.ok) {
            throw new Error(result.error || 'Storage test failed');
          }

        } catch (error) {
          throw error;
        }
      },

      // Clear Error
      clearError: () => {
        set((state) => {
          state.error = null;
          state.usersError = null;
          state.settingsError = null;
        });
      },
    })),
    { name: 'admin-store' }
  )
);