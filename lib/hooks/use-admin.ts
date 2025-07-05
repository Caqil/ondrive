import { useAdminStore } from '@/store/admin-store';

// Selectors - following your pattern like useUser, useIsAuthenticated
export const useAdminStats = () => useAdminStore((state) => state.stats);
export const useAdminUsers = () => useAdminStore((state) => state.users);
export const useAdminSystemHealth = () => useAdminStore((state) => state.systemHealth);
export const useAdminSettings = () => useAdminStore((state) => state.settings);
export const useAdminLoading = () => useAdminStore((state) => state.isLoading);
export const useAdminError = () => useAdminStore((state) => state.error);

// Actions - following your pattern like useAuthActions, useFileActions
export const useAdminActions = () => {
  const store = useAdminStore();
  return {
    loadDashboardStats: store.loadDashboardStats,
    loadSystemHealth: store.loadSystemHealth,
    loadUsers: store.loadUsers,
    performUserAction: store.performUserAction,
    loadSettings: store.loadSettings,
    updateSettings: store.updateSettings,
    testEmailSettings: store.testEmailSettings,
    testStorageSettings: store.testStorageSettings,
    clearError: store.clearError,
  };
};

// Main convenience hook - following your pattern like useAuth
export const useAdmin = () => {
  const store = useAdminStore();
  const actions = useAdminActions();

  return {
    // State
    stats: store.stats,
    users: store.users,
    systemHealth: store.systemHealth,
    settings: store.settings,
    isLoading: store.isLoading,
    error: store.error,
    usersPagination: store.usersPagination,

    // Actions
    ...actions,
  };
};

// lib/hooks/use-billing.ts
import { useBillingStore } from '@/store/billing-store';

// Selectors - following your pattern
export const useBillingInvoices = () => useBillingStore((state) => state.invoices);
export const useBillingPayments = () => useBillingStore((state) => state.payments);
export const useBillingCoupons = () => useBillingStore((state) => state.coupons);
export const useBillingLoading = () => useBillingStore((state) => state.isLoading);
export const useBillingError = () => useBillingStore((state) => state.error);

// Actions - following your pattern
export const useBillingActions = () => {
  const store = useBillingStore();
  return {
    loadInvoices: store.loadInvoices,
    loadInvoice: store.loadInvoice,
    createInvoice: store.createInvoice,
    downloadInvoice: store.downloadInvoice,
    loadPayments: store.loadPayments,
    createPayment: store.createPayment,
    loadCoupons: store.loadCoupons,
    applyCoupon: store.applyCoupon,
    clearError: store.clearError,
  };
};

// Main convenience hook
export const useBilling = () => {
  const store = useBillingStore();
  const actions = useBillingActions();

  return {
    // State
    invoices: store.invoices,
    currentInvoice: store.currentInvoice,
    payments: store.payments,
    coupons: store.coupons,
    isLoading: store.isLoading,
    error: store.error,

    // Actions
    ...actions,
  };
};

// lib/hooks/use-intersection-observer.ts
import { useEffect, useRef, useState, useCallback } from 'react';

interface IntersectionObserverOptions {
  root?: Element | null;
  rootMargin?: string;
  threshold?: number | number[];
  triggerOnce?: boolean;
  skip?: boolean;
}

interface IntersectionObserverResult {
  ref: (node?: Element | null) => void;
  inView: boolean;
  entry?: IntersectionObserverEntry;
}

export const useIntersectionObserver = (
  options: IntersectionObserverOptions = {}
): IntersectionObserverResult => {
  const {
    root = null,
    rootMargin = '0px',
    threshold = 0,
    triggerOnce = false,
    skip = false,
  } = options;

  const [inView, setInView] = useState(false);
  const [entry, setEntry] = useState<IntersectionObserverEntry>();
  const elementRef = useRef<Element | null>(null);
  const observerRef = useRef<IntersectionObserver | null>(null);

  const ref = useCallback((node?: Element | null) => {
    if (elementRef.current && observerRef.current) {
      observerRef.current.unobserve(elementRef.current);
    }

    elementRef.current = node ?? null;

    if (skip || !node) return;

    if (observerRef.current && node) {
      observerRef.current.observe(node);
    }
  }, [skip]);

  useEffect(() => {
    if (skip) return;

    const observer = new IntersectionObserver(
      ([entry]) => {
        const isIntersecting = entry.isIntersecting;
        
        setInView(isIntersecting);
        setEntry(entry);

        if (triggerOnce && isIntersecting && observerRef.current && elementRef.current) {
          observerRef.current.unobserve(elementRef.current);
        }
      },
      { root, rootMargin, threshold }
    );

    observerRef.current = observer;

    if (elementRef.current) {
      observer.observe(elementRef.current);
    }

    return () => {
      observer.disconnect();
      observerRef.current = null;
    };
  }, [root, rootMargin, threshold, triggerOnce, skip]);

  return { ref, inView, entry };
};

export const useInfiniteScroll = (
  callback: () => void | Promise<void>,
  options: {
    hasNextPage?: boolean;
    isLoading?: boolean;
    rootMargin?: string;
    threshold?: number;
  } = {}
) => {
  const {
    hasNextPage = true,
    isLoading = false,
    rootMargin = '100px',
    threshold = 0.1,
  } = options;

  const { ref, inView } = useIntersectionObserver({
    rootMargin,
    threshold,
    skip: !hasNextPage || isLoading,
  });

  useEffect(() => {
    if (inView && hasNextPage && !isLoading) {
      callback();
    }
  }, [inView, hasNextPage, isLoading, callback]);

  return { ref, inView };
};