import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { immer } from 'zustand/middleware/immer';
import type { 
  Invoice,
  Payment,
  Coupon,
  CreateInvoiceRequest,
  CreatePaymentRequest,
  ObjectId 
} from '@/types';

interface BillingState {
  // Invoices
  invoices: Invoice[];
  currentInvoice: Invoice | null;
  invoicesLoading: boolean;
  
  // Payments
  payments: Payment[];
  paymentsLoading: boolean;
  
  // Coupons
  coupons: Coupon[];
  couponsLoading: boolean;
  
  // General
  isLoading: boolean;
  error: string | null;
  
  // Actions
  loadInvoices: (filters?: any) => Promise<void>;
  loadInvoice: (invoiceId: ObjectId) => Promise<void>;
  createInvoice: (data: CreateInvoiceRequest) => Promise<Invoice>;
  downloadInvoice: (invoiceId: ObjectId) => Promise<void>;
  loadPayments: (filters?: any) => Promise<void>;
  createPayment: (data: CreatePaymentRequest) => Promise<Payment>;
  loadCoupons: () => Promise<void>;
  applyCoupon: (code: string) => Promise<Coupon>;
  clearError: () => void;
}

export const useBillingStore = create<BillingState>()(
  devtools(
    immer((set, get) => ({
      // Initial State
      invoices: [],
      currentInvoice: null,
      invoicesLoading: false,
      payments: [],
      paymentsLoading: false,
      coupons: [],
      couponsLoading: false,
      isLoading: false,
      error: null,

      // Load Invoices
      loadInvoices: async (filters = {}) => {
        set((state) => {
          state.invoicesLoading = true;
          state.error = null;
        });

        try {
          const searchParams = new URLSearchParams({
            ...Object.fromEntries(
              Object.entries(filters)
                .filter(([_, value]) => value !== undefined && value !== '')
                .map(([key, value]) => [key, String(value)])
            ),
          });

          const response = await fetch(`/api/client/billing/invoices?${searchParams}`);
          const result = await response.json();

          if (!response.ok) {
            throw new Error(result.error || 'Failed to load invoices');
          }

          set((state) => {
            state.invoices = result.data;
            state.invoicesLoading = false;
          });

        } catch (error) {
          set((state) => {
            state.error = error instanceof Error ? error.message : 'Failed to load invoices';
            state.invoicesLoading = false;
          });
        }
      },

      // Load Single Invoice
      loadInvoice: async (invoiceId: ObjectId) => {
        set((state) => {
          state.isLoading = true;
          state.error = null;
        });

        try {
          const response = await fetch(`/api/client/billing/invoices/${invoiceId}`);
          const result = await response.json();

          if (!response.ok) {
            throw new Error(result.error || 'Failed to load invoice');
          }

          set((state) => {
            state.currentInvoice = result.data;
            state.isLoading = false;
          });

        } catch (error) {
          set((state) => {
            state.error = error instanceof Error ? error.message : 'Failed to load invoice';
            state.isLoading = false;
          });
        }
      },

      // Create Invoice
      createInvoice: async (data: CreateInvoiceRequest) => {
        set((state) => {
          state.isLoading = true;
          state.error = null;
        });

        try {
          const response = await fetch('/api/client/billing/invoices', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data),
          });

          const result = await response.json();

          if (!response.ok) {
            throw new Error(result.error || 'Failed to create invoice');
          }

          const newInvoice: Invoice = result.data;

          set((state) => {
            state.invoices.unshift(newInvoice);
            state.isLoading = false;
          });

          return newInvoice;

        } catch (error) {
          set((state) => {
            state.error = error instanceof Error ? error.message : 'Failed to create invoice';
            state.isLoading = false;
          });
          throw error;
        }
      },

      // Download Invoice
      downloadInvoice: async (invoiceId: ObjectId) => {
        try {
          const response = await fetch(`/api/client/billing/download-invoice/${invoiceId}`);

          if (!response.ok) {
            const result = await response.json();
            throw new Error(result.error || 'Failed to download invoice');
          }

          // Trigger download
          const blob = await response.blob();
          const url = window.URL.createObjectURL(blob);
          const a = document.createElement('a');
          a.href = url;
          a.download = `invoice-${invoiceId}.pdf`;
          document.body.appendChild(a);
          a.click();
          window.URL.revokeObjectURL(url);
          document.body.removeChild(a);

        } catch (error) {
          set((state) => {
            state.error = error instanceof Error ? error.message : 'Failed to download invoice';
          });
          throw error;
        }
      },

      // Load Payments
      loadPayments: async (filters = {}) => {
        set((state) => {
          state.paymentsLoading = true;
          state.error = null;
        });

        try {
          const searchParams = new URLSearchParams(
            Object.fromEntries(
              Object.entries(filters)
                .filter(([_, value]) => value !== undefined && value !== '')
                .map(([key, value]) => [key, String(value)])
            )
          );

          const response = await fetch(`/api/client/billing/payments?${searchParams}`);
          const result = await response.json();

          if (!response.ok) {
            throw new Error(result.error || 'Failed to load payments');
          }

          set((state) => {
            state.payments = result.data;
            state.paymentsLoading = false;
          });

        } catch (error) {
          set((state) => {
            state.error = error instanceof Error ? error.message : 'Failed to load payments';
            state.paymentsLoading = false;
          });
        }
      },

      // Create Payment
      createPayment: async (data: CreatePaymentRequest) => {
        set((state) => {
          state.isLoading = true;
          state.error = null;
        });

        try {
          const response = await fetch('/api/client/billing/payments', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data),
          });

          const result = await response.json();

          if (!response.ok) {
            throw new Error(result.error || 'Failed to create payment');
          }

          const newPayment: Payment = result.data;

          set((state) => {
            state.payments.unshift(newPayment);
            state.isLoading = false;
          });

          return newPayment;

        } catch (error) {
          set((state) => {
            state.error = error instanceof Error ? error.message : 'Failed to create payment';
            state.isLoading = false;
          });
          throw error;
        }
      },

      // Load Coupons
      loadCoupons: async () => {
        set((state) => {
          state.couponsLoading = true;
          state.error = null;
        });

        try {
          const response = await fetch('/api/client/billing/coupons');
          const result = await response.json();

          if (!response.ok) {
            throw new Error(result.error || 'Failed to load coupons');
          }

          set((state) => {
            state.coupons = result.data;
            state.couponsLoading = false;
          });

        } catch (error) {
          set((state) => {
            state.error = error instanceof Error ? error.message : 'Failed to load coupons';
            state.couponsLoading = false;
          });
        }
      },

      // Apply Coupon
      applyCoupon: async (code: string) => {
        try {
          const response = await fetch('/api/client/billing/coupons/apply', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ code }),
          });

          const result = await response.json();

          if (!response.ok) {
            throw new Error(result.error || 'Failed to apply coupon');
          }

          return result.data;

        } catch (error) {
          set((state) => {
            state.error = error instanceof Error ? error.message : 'Failed to apply coupon';
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
    })),
    { name: 'billing-store' }
  )
);
