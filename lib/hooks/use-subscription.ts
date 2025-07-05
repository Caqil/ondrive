// lib/hooks/use-subscription.ts
import { useSubscriptionStore } from '@/store/subscription-store';

/**
 * Simple hook that re-exports subscription store functionality
 * Maintains consistency with existing store patterns
 */
export const useSubscription = () => {
  const store = useSubscriptionStore();

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

    // Actions
    loadSubscription: store.loadSubscription,
    loadPlans: store.loadPlans,
    createSubscription: store.createSubscription,
    upgradeSubscription: store.upgradeSubscription,
    cancelSubscription: store.cancelSubscription,
    updatePaymentMethod: store.updatePaymentMethod,
    loadUsage: store.loadUsage,
    loadUsageHistory: store.loadUsageHistory,
    clearError: store.clearError,
  };
};