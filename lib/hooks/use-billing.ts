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