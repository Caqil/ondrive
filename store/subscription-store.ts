import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { immer } from 'zustand/middleware/immer';
import type { 
  Subscription,
  Plan,
  CreateSubscriptionRequest,
  UpgradeSubscriptionRequest,
  UsageRecord,
  ObjectId 
} from '@/types';

interface SubscriptionState {
  // Current Subscription
  subscription: Subscription | null;
  
  // Plans
  plans: Plan[];
  plansLoading: boolean;
  
  // Usage
  usage: UsageRecord | null;
  usageHistory: UsageRecord[];
  usageLoading: boolean;
  
  // Billing
  nextBilling: {
    date: Date | null;
    amount: number;
    currency: string;
  };
  
  // Loading
  isLoading: boolean;
  error: string | null;
  
  // Actions
  loadSubscription: () => Promise<void>;
  loadPlans: () => Promise<void>;
  createSubscription: (data: CreateSubscriptionRequest) => Promise<void>;
  upgradeSubscription: (data: UpgradeSubscriptionRequest) => Promise<void>;
  cancelSubscription: () => Promise<void>;
  renewSubscription: () => Promise<void>;
  loadUsage: () => Promise<void>;
  loadUsageHistory: (period?: 'daily' | 'monthly' | 'yearly') => Promise<void>;
  clearError: () => void;
}

export const useSubscriptionStore = create<SubscriptionState>()(
  devtools(
    immer((set, get) => ({
      // Initial State
      subscription: null,
      plans: [],
      plansLoading: false,
      usage: null,
      usageHistory: [],
      usageLoading: false,
      nextBilling: {
        date: null,
        amount: 0,
        currency: 'USD',
      },
      isLoading: false,
      error: null,

      // Load Subscription
      loadSubscription: async () => {
        set((state) => {
          state.isLoading = true;
          state.error = null;
        });

        try {
          const response = await fetch('/api/client/subscription');
          const result = await response.json();

          if (!response.ok) {
            throw new Error(result.error || 'Failed to load subscription');
          }

          set((state) => {
            state.subscription = result.data;
            if (result.data) {
              state.nextBilling = {
                date: result.data.nextBillingDate ? new Date(result.data.nextBillingDate) : null,
                amount: result.data.amount,
                currency: result.data.currency,
              };
            }
            state.isLoading = false;
          });

        } catch (error) {
          set((state) => {
            state.error = error instanceof Error ? error.message : 'Failed to load subscription';
            state.isLoading = false;
          });
        }
      },

      // Load Plans
      loadPlans: async () => {
        set((state) => {
          state.plansLoading = true;
          state.error = null;
        });

        try {
          const response = await fetch('/api/client/plans');
          const result = await response.json();

          if (!response.ok) {
            throw new Error(result.error || 'Failed to load plans');
          }

          set((state) => {
            state.plans = result.data;
            state.plansLoading = false;
          });

        } catch (error) {
          set((state) => {
            state.error = error instanceof Error ? error.message : 'Failed to load plans';
            state.plansLoading = false;
          });
        }
      },

      // Create Subscription
      createSubscription: async (data: CreateSubscriptionRequest) => {
        set((state) => {
          state.isLoading = true;
          state.error = null;
        });

        try {
          const response = await fetch('/api/client/subscription', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data),
          });

          const result = await response.json();

          if (!response.ok) {
            throw new Error(result.error || 'Failed to create subscription');
          }

          set((state) => {
            state.subscription = result.data;
            state.isLoading = false;
          });

        } catch (error) {
          set((state) => {
            state.error = error instanceof Error ? error.message : 'Failed to create subscription';
            state.isLoading = false;
          });
          throw error;
        }
      },

      // Upgrade Subscription
      upgradeSubscription: async (data: UpgradeSubscriptionRequest) => {
        set((state) => {
          state.isLoading = true;
          state.error = null;
        });

        try {
          const response = await fetch('/api/client/subscription/upgrade', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data),
          });

          const result = await response.json();

          if (!response.ok) {
            throw new Error(result.error || 'Failed to upgrade subscription');
          }

          set((state) => {
            state.subscription = result.data;
            state.isLoading = false;
          });

        } catch (error) {
          set((state) => {
            state.error = error instanceof Error ? error.message : 'Failed to upgrade subscription';
            state.isLoading = false;
          });
          throw error;
        }
      },

      // Cancel Subscription
      cancelSubscription: async () => {
        set((state) => {
          state.isLoading = true;
          state.error = null;
        });

        try {
          const response = await fetch('/api/client/subscription/cancel', {
            method: 'POST',
          });

          const result = await response.json();

          if (!response.ok) {
            throw new Error(result.error || 'Failed to cancel subscription');
          }

          set((state) => {
            state.subscription = result.data;
            state.isLoading = false;
          });

        } catch (error) {
          set((state) => {
            state.error = error instanceof Error ? error.message : 'Failed to cancel subscription';
            state.isLoading = false;
          });
          throw error;
        }
      },

      // Renew Subscription
      renewSubscription: async () => {
        set((state) => {
          state.isLoading = true;
          state.error = null;
        });

        try {
          const response = await fetch('/api/client/subscription/renew', {
            method: 'POST',
          });

          const result = await response.json();

          if (!response.ok) {
            throw new Error(result.error || 'Failed to renew subscription');
          }

          set((state) => {
            state.subscription = result.data;
            state.isLoading = false;
          });

        } catch (error) {
          set((state) => {
            state.error = error instanceof Error ? error.message : 'Failed to renew subscription';
            state.isLoading = false;
          });
          throw error;
        }
      },

      // Load Usage
      loadUsage: async () => {
        set((state) => {
          state.usageLoading = true;
          state.error = null;
        });

        try {
          const response = await fetch('/api/client/subscription/usage');
          const result = await response.json();

          if (!response.ok) {
            throw new Error(result.error || 'Failed to load usage');
          }

          set((state) => {
            state.usage = result.data;
            state.usageLoading = false;
          });

        } catch (error) {
          set((state) => {
            state.error = error instanceof Error ? error.message : 'Failed to load usage';
            state.usageLoading = false;
          });
        }
      },

      // Load Usage History
      loadUsageHistory: async (period = 'monthly') => {
        try {
          const response = await fetch(`/api/client/subscription/usage/history?period=${period}`);
          const result = await response.json();

          if (!response.ok) {
            throw new Error(result.error || 'Failed to load usage history');
          }

          set((state) => {
            state.usageHistory = result.data;
          });

        } catch (error) {
          set((state) => {
            state.error = error instanceof Error ? error.message : 'Failed to load usage history';
          });
        }
      },

      // Clear Error
      clearError: () => {
        set((state) => {
          state.error = null;
        });
      },
    })),
    { name: 'subscription-store' }
  )
);