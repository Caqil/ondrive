// store/auth-store.ts
import { create } from 'zustand';
import { devtools, persist } from 'zustand/middleware';
import { immer } from 'zustand/middleware/immer';
import type { 
  User, 
  AuthSession, 
  LoginRequest, 
  RegisterRequest, 
  ForgotPasswordRequest, 
  ResetPasswordRequest,
  ChangePasswordRequest,
  TwoFactorSetupRequest,
  TwoFactorVerifyRequest,
  UpdateUserRequest
} from '@/types';

interface AuthState {
  // State
  user: User | null;
  session: AuthSession | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  isInitialized: boolean;
  error: string | null;
  
  // 2FA State
  is2FARequired: boolean;
  pendingSessionToken: string | null;
  
  // Password Reset State
  isResettingPassword: boolean;
  resetToken: string | null;
  
  // Actions
  login: (credentials: LoginRequest) => Promise<void>;
  register: (data: RegisterRequest) => Promise<void>;
  logout: () => Promise<void>;
  refreshSession: () => Promise<void>;
  
  // Password Management
  forgotPassword: (data: ForgotPasswordRequest) => Promise<void>;
  resetPassword: (data: ResetPasswordRequest) => Promise<void>;
  changePassword: (data: ChangePasswordRequest) => Promise<void>;
  
  // 2FA Management
  setup2FA: () => Promise<{ secret: string; qrCode: string }>;
  verify2FA: (data: TwoFactorVerifyRequest) => Promise<void>;
  disable2FA: (code: string) => Promise<void>;
  
  // Profile Management
  updateProfile: (data: UpdateUserRequest) => Promise<void>;
  uploadAvatar: (file: File) => Promise<void>;
  
  // Utility Actions
  clearError: () => void;
  initialize: () => Promise<void>;
  setUser: (user: User | null) => void;
}

export const useAuthStore = create<AuthState>()(
  devtools(
    persist(
      immer((set, get) => ({
        // Initial State
        user: null,
        session: null,
        isAuthenticated: false,
        isLoading: false,
        isInitialized: false,
        error: null,
        is2FARequired: false,
        pendingSessionToken: null,
        isResettingPassword: false,
        resetToken: null,

        // Login Action
        login: async (credentials: LoginRequest) => {
          set((state) => {
            state.isLoading = true;
            state.error = null;
          });

          try {
            const response = await fetch('/api/auth/login', {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify(credentials),
            });

            const data = await response.json();

            if (!response.ok) {
              throw new Error(data.error || 'Login failed');
            }

            // Handle 2FA requirement
            if (data.requires2FA) {
              set((state) => {
                state.is2FARequired = true;
                state.pendingSessionToken = data.sessionToken;
                state.isLoading = false;
              });
              return;
            }

            // Successful login
            set((state) => {
              state.user = data.user;
              state.session = data.session;
              state.isAuthenticated = true;
              state.isLoading = false;
              state.is2FARequired = false;
              state.pendingSessionToken = null;
            });

          } catch (error) {
            set((state) => {
              state.error = error instanceof Error ? error.message : 'Login failed';
              state.isLoading = false;
            });
            throw error;
          }
        },

        // Register Action
        register: async (data: RegisterRequest) => {
          set((state) => {
            state.isLoading = true;
            state.error = null;
          });

          try {
            const response = await fetch('/api/auth/register', {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify(data),
            });

            const result = await response.json();

            if (!response.ok) {
              throw new Error(result.error || 'Registration failed');
            }

            // Auto-login after registration if email is verified
            if (result.user && result.session) {
              set((state) => {
                state.user = result.user;
                state.session = result.session;
                state.isAuthenticated = true;
              });
            }

            set((state) => {
              state.isLoading = false;
            });

          } catch (error) {
            set((state) => {
              state.error = error instanceof Error ? error.message : 'Registration failed';
              state.isLoading = false;
            });
            throw error;
          }
        },

        // Logout Action
        logout: async () => {
          set((state) => {
            state.isLoading = true;
          });

          try {
            await fetch('/api/auth/logout', {
              method: 'POST',
            });
          } catch (error) {
            console.warn('Logout request failed:', error);
          } finally {
            set((state) => {
              state.user = null;
              state.session = null;
              state.isAuthenticated = false;
              state.isLoading = false;
              state.error = null;
              state.is2FARequired = false;
              state.pendingSessionToken = null;
            });
          }
        },

        // Refresh Session
        refreshSession: async () => {
          try {
            const response = await fetch('/api/auth/session');
            const data = await response.json();

            if (response.ok && data.user) {
              set((state) => {
                state.user = data.user;
                state.session = data.session;
                state.isAuthenticated = true;
              });
            } else {
              // Session expired or invalid
              set((state) => {
                state.user = null;
                state.session = null;
                state.isAuthenticated = false;
              });
            }
          } catch (error) {
            console.warn('Session refresh failed:', error);
            set((state) => {
              state.user = null;
              state.session = null;
              state.isAuthenticated = false;
            });
          }
        },

        // Forgot Password
        forgotPassword: async (data: ForgotPasswordRequest) => {
          set((state) => {
            state.isLoading = true;
            state.error = null;
          });

          try {
            const response = await fetch('/api/auth/forgot-password', {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify(data),
            });

            const result = await response.json();

            if (!response.ok) {
              throw new Error(result.error || 'Password reset request failed');
            }

            set((state) => {
              state.isLoading = false;
            });

          } catch (error) {
            set((state) => {
              state.error = error instanceof Error ? error.message : 'Password reset failed';
              state.isLoading = false;
            });
            throw error;
          }
        },

        // Reset Password
        resetPassword: async (data: ResetPasswordRequest) => {
          set((state) => {
            state.isResettingPassword = true;
            state.error = null;
          });

          try {
            const response = await fetch('/api/auth/reset-password', {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify(data),
            });

            const result = await response.json();

            if (!response.ok) {
              throw new Error(result.error || 'Password reset failed');
            }

            set((state) => {
              state.isResettingPassword = false;
              state.resetToken = null;
            });

          } catch (error) {
            set((state) => {
              state.error = error instanceof Error ? error.message : 'Password reset failed';
              state.isResettingPassword = false;
            });
            throw error;
          }
        },

        // Change Password
        changePassword: async (data: ChangePasswordRequest) => {
          set((state) => {
            state.isLoading = true;
            state.error = null;
          });

          try {
            const response = await fetch('/api/auth/change-password', {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify(data),
            });

            const result = await response.json();

            if (!response.ok) {
              throw new Error(result.error || 'Password change failed');
            }

            set((state) => {
              state.isLoading = false;
            });

          } catch (error) {
            set((state) => {
              state.error = error instanceof Error ? error.message : 'Password change failed';
              state.isLoading = false;
            });
            throw error;
          }
        },

        // Setup 2FA
        setup2FA: async () => {
          set((state) => {
            state.isLoading = true;
            state.error = null;
          });

          try {
            const response = await fetch('/api/auth/two-factor/setup', {
              method: 'POST',
            });

            const result = await response.json();

            if (!response.ok) {
              throw new Error(result.error || '2FA setup failed');
            }

            set((state) => {
              state.isLoading = false;
            });

            return result;

          } catch (error) {
            set((state) => {
              state.error = error instanceof Error ? error.message : '2FA setup failed';
              state.isLoading = false;
            });
            throw error;
          }
        },

        // Verify 2FA
        verify2FA: async (data: TwoFactorVerifyRequest) => {
          set((state) => {
            state.isLoading = true;
            state.error = null;
          });

          try {
            const { pendingSessionToken } = get();
            
            const response = await fetch('/api/auth/two-factor/verify', {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify({
                ...data,
                sessionToken: pendingSessionToken,
              }),
            });

            const result = await response.json();

            if (!response.ok) {
              throw new Error(result.error || '2FA verification failed');
            }

            // Complete login after 2FA verification
            set((state) => {
              state.user = result.user;
              state.session = result.session;
              state.isAuthenticated = true;
              state.isLoading = false;
              state.is2FARequired = false;
              state.pendingSessionToken = null;
            });

          } catch (error) {
            set((state) => {
              state.error = error instanceof Error ? error.message : '2FA verification failed';
              state.isLoading = false;
            });
            throw error;
          }
        },

        // Disable 2FA
        disable2FA: async (code: string) => {
          set((state) => {
            state.isLoading = true;
            state.error = null;
          });

          try {
            const response = await fetch('/api/auth/two-factor/disable', {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify({ code }),
            });

            const result = await response.json();

            if (!response.ok) {
              throw new Error(result.error || '2FA disable failed');
            }

            // Update user state
            set((state) => {
              if (state.user) {
                state.user.twoFactorEnabled = false;
              }
              state.isLoading = false;
            });

          } catch (error) {
            set((state) => {
              state.error = error instanceof Error ? error.message : '2FA disable failed';
              state.isLoading = false;
            });
            throw error;
          }
        },

        // Update Profile
        updateProfile: async (data: UpdateUserRequest) => {
          set((state) => {
            state.isLoading = true;
            state.error = null;
          });

          try {
            const response = await fetch('/api/client/profile', {
              method: 'PATCH',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify(data),
            });

            const result = await response.json();

            if (!response.ok) {
              throw new Error(result.error || 'Profile update failed');
            }

            set((state) => {
              state.user = result.data;
              state.isLoading = false;
            });

          } catch (error) {
            set((state) => {
              state.error = error instanceof Error ? error.message : 'Profile update failed';
              state.isLoading = false;
            });
            throw error;
          }
        },

        // Upload Avatar
        uploadAvatar: async (file: File) => {
          set((state) => {
            state.isLoading = true;
            state.error = null;
          });

          try {
            const formData = new FormData();
            formData.append('avatar', file);

            const response = await fetch('/api/client/profile/avatar', {
              method: 'POST',
              body: formData,
            });

            const result = await response.json();

            if (!response.ok) {
              throw new Error(result.error || 'Avatar upload failed');
            }

            set((state) => {
              if (state.user) {
                state.user.avatar = result.data.avatarUrl;
              }
              state.isLoading = false;
            });

          } catch (error) {
            set((state) => {
              state.error = error instanceof Error ? error.message : 'Avatar upload failed';
              state.isLoading = false;
            });
            throw error;
          }
        },

        // Clear Error
        clearError: () => {
          set((state) => {
            state.error = null;
          });
        },

        // Initialize Auth State
        initialize: async () => {
          if (get().isInitialized) return;

          set((state) => {
            state.isLoading = true;
          });

          try {
            await get().refreshSession();
          } catch (error) {
            console.warn('Auth initialization failed:', error);
          } finally {
            set((state) => {
              state.isInitialized = true;
              state.isLoading = false;
            });
          }
        },

        // Set User (for external updates)
        setUser: (user: User | null) => {
          set((state) => {
            state.user = user;
            state.isAuthenticated = !!user;
          });
        },
      })),
      {
        name: 'auth-store',
        partialize: (state) => ({
          user: state.user,
          session: state.session,
          isAuthenticated: state.isAuthenticated,
        }),
        version: 1,
      }
    ),
    { name: 'auth-store' }
  )
);

// Selectors for common use cases
export const useUser = () => useAuthStore((state) => state.user);
export const useIsAuthenticated = () => useAuthStore((state) => state.isAuthenticated);
export const useAuthLoading = () => useAuthStore((state) => state.isLoading);
export const useAuthError = () => useAuthStore((state) => state.error);
export const useIs2FARequired = () => useAuthStore((state) => state.is2FARequired);

// Auth Actions
export const useAuthActions = () => {
  const store = useAuthStore();
  return {
    login: store.login,
    register: store.register,
    logout: store.logout,
    forgotPassword: store.forgotPassword,
    resetPassword: store.resetPassword,
    changePassword: store.changePassword,
    setup2FA: store.setup2FA,
    verify2FA: store.verify2FA,
    disable2FA: store.disable2FA,
    updateProfile: store.updateProfile,
    uploadAvatar: store.uploadAvatar,
    clearError: store.clearError,
    initialize: store.initialize,
  };
};