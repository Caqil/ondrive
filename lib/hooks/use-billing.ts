// lib/hooks/use-billing.ts
import { useBillingStore } from '@/store/billing-store';

/**
 * Simple hook that re-exports billing store functionality
 * Maintains consistency with existing store patterns
 */
export const useBilling = () => {
  const store = useBillingStore();

  return {
    // State
    invoices: store.invoices,
    payments: store.payments,
    paymentMethods: store.paymentMethods,
    isLoading: store.isLoading,
    error: store.error,
    pagination: store.pagination,

    // Actions
    loadInvoices: store.loadInvoices,
    loadPayments: store.loadPayments,
    loadPaymentMethods: store.loadPaymentMethods,
    createInvoice: store.createInvoice,
    sendInvoice: store.sendInvoice,
    processPayment: store.processPayment,
    addPaymentMethod: store.addPaymentMethod,
    removePaymentMethod: store.removePaymentMethod,
    clearError: store.clearError,
  };
};