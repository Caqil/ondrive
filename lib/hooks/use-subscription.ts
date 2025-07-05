import { useSubscriptionStore } from '@/store/subscription-store';

// Selectors - following your pattern
export const useCurrentSubscription = () => useSubscriptionStore((state) => state.subscription);
export const useSubscriptionPlans = () => useSubscriptionStore((state) => state.plans);
export const useSubscriptionUsage = () => useSubscriptionStore((state) => state.usage);
export const useSubscriptionLoading = () => useSubscriptionStore((state) => state.isLoading);
export const useSubscriptionError = () => useSubscriptionStore((state) => state.error);

// Actions - following your pattern
export const useSubscriptionActions = () => {
  const store = useSubscriptionStore();
  return {
    loadSubscription: store.loadSubscription,
    loadPlans: store.loadPlans,
    createSubscription: store.createSubscription,
    upgradeSubscription: store.upgradeSubscription,
    cancelSubscription: store.cancelSubscription,
    renewSubscription: store.renewSubscription,
    loadUsage: store.loadUsage,
    loadUsageHistory: store.loadUsageHistory,
    clearError: store.clearError,
  };
};

// Main convenience hook
export const useSubscription = () => {
  const store = useSubscriptionStore();
  const actions = useSubscriptionActions();

  return {
    // State
    subscription: store.subscription,
    plans: store.plans,
    usage: store.usage,
    usageHistory: store.usageHistory,
    isLoading: store.isLoading,
    plansLoading: store.plansLoading,
    usageLoading: store.usageLoading,
    error: store.error,
    nextBilling: store.nextBilling,

    // Actions
    ...actions,
  };
};