/**
 * Revenue Analytics - Financial metrics and subscription analytics
 */

import { ObjectId, AnalyticsDateRange } from '../../types';

export interface RevenueMetrics {
  totalRevenue: number;
  monthlyRecurringRevenue: number;
  annualRecurringRevenue: number;
  averageRevenuePerUser: number;
  revenueGrowthRate: number;
  churnRate: number;
  customerLifetimeValue: number;
  revenueByPlan: { plan: string; revenue: number; subscribers: number }[];
  revenueTrend: { date: string; revenue: number; subscribers: number }[];
}

export interface SubscriptionMetrics {
  totalSubscriptions: number;
  activeSubscriptions: number;
  trialSubscriptions: number;
  cancelledSubscriptions: number;
  conversionRate: number;
  subscriptionsByPlan: { plan: string; count: number; percentage: number }[];
  subscriptionGrowth: { date: string; new: number; churned: number; net: number }[];
}

export interface PaymentMetrics {
  totalPayments: number;
  successfulPayments: number;
  failedPayments: number;
  refunds: number;
  chargebacks: number;
  paymentMethodDistribution: { method: string; count: number; percentage: number }[];
  averageTransactionValue: number;
  paymentTrends: { date: string; amount: number; count: number }[];
}

export interface BillingAnalytics {
  outstandingInvoices: number;
  overdueInvoices: number;
  paidInvoices: number;
  totalOutstanding: number;
  averagePaymentDelay: number;
  billingIssues: {
    type: string;
    count: number;
    description: string;
  }[];
}

export class RevenueAnalytics {
  /**
   * Get revenue overview metrics
   */
  static async getRevenueMetrics(dateRange: AnalyticsDateRange): Promise<RevenueMetrics> {
    try {
      const response = await fetch('/api/admin/analytics/revenue', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ dateRange })
      });

      const result = await response.json();

      if (!response.ok) {
        throw new Error(result.error || 'Failed to fetch revenue metrics');
      }

      return result.data;
    } catch (error) {
      console.error('Error fetching revenue metrics:', error);
      throw error;
    }
  }

  /**
   * Get subscription analytics
   */
  static async getSubscriptionMetrics(dateRange: AnalyticsDateRange): Promise<SubscriptionMetrics> {
    try {
      const response = await fetch('/api/admin/analytics/subscriptions', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ dateRange })
      });

      const result = await response.json();

      if (!response.ok) {
        throw new Error(result.error || 'Failed to fetch subscription metrics');
      }

      return result.data;
    } catch (error) {
      console.error('Error fetching subscription metrics:', error);
      throw error;
    }
  }

  /**
   * Get payment analytics
   */
  static async getPaymentMetrics(dateRange: AnalyticsDateRange): Promise<PaymentMetrics> {
    try {
      const response = await fetch('/api/admin/analytics/payments', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ dateRange })
      });

      const result = await response.json();

      if (!response.ok) {
        throw new Error(result.error || 'Failed to fetch payment metrics');
      }

      return result.data;
    } catch (error) {
      console.error('Error fetching payment metrics:', error);
      throw error;
    }
  }

  /**
   * Get billing analytics
   */
  static async getBillingAnalytics(dateRange: AnalyticsDateRange): Promise<BillingAnalytics> {
    try {
      const response = await fetch('/api/admin/analytics/billing', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ dateRange })
      });

      const result = await response.json();

      if (!response.ok) {
        throw new Error(result.error || 'Failed to fetch billing analytics');
      }

      return result.data;
    } catch (error) {
      console.error('Error fetching billing analytics:', error);
      throw error;
    }
  }

  /**
   * Get revenue forecasting
   */
  static async getRevenueForecast(months: number = 12): Promise<{
    forecast: { date: string; predicted: number; confidence: number }[];
    assumptions: string[];
    factors: { name: string; impact: number; description: string }[];
  }> {
    try {
      const response = await fetch(`/api/admin/analytics/revenue-forecast?months=${months}`);
      const result = await response.json();
      
      if (!response.ok) {
        throw new Error(result.error || 'Failed to fetch revenue forecast');
      }

      return result.data;
    } catch (error) {
      console.error('Error fetching revenue forecast:', error);
      throw error;
    }
  }

  /**
   * Get cohort revenue analysis
   */
  static async getCohortRevenue(dateRange: AnalyticsDateRange): Promise<{
    cohorts: {
      cohort: string;
      size: number;
      revenueByMonth: number[];
      lifetimeValue: number;
    }[];
    summary: {
      averageLTV: number;
      bestCohort: string;
      trendDirection: 'up' | 'down' | 'stable';
    };
  }> {
    try {
      const response = await fetch('/api/admin/analytics/cohort-revenue', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ dateRange })
      });

      const result = await response.json();

      if (!response.ok) {
        throw new Error(result.error || 'Failed to fetch cohort revenue');
      }

      return result.data;
    } catch (error) {
      console.error('Error fetching cohort revenue:', error);
      throw error;
    }
  }

  /**
   * Get revenue by geography
   */
  static async getRevenueByGeography(dateRange: AnalyticsDateRange): Promise<{
    countries: { country: string; revenue: number; subscribers: number }[];
    regions: { region: string; revenue: number; percentage: number }[];
    topMarkets: { market: string; growth: number; potential: string }[];
  }> {
    try {
      const response = await fetch('/api/admin/analytics/revenue-geography', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ dateRange })
      });

      const result = await response.json();

      if (!response.ok) {
        throw new Error(result.error || 'Failed to fetch revenue geography');
      }

      return result.data;
    } catch (error) {
      console.error('Error fetching revenue geography:', error);
      throw error;
    }
  }

  /**
   * Get price optimization analysis
   */
  static async getPriceOptimization(): Promise<{
    currentPlans: { plan: string; price: number; subscribers: number; churnRate: number }[];
    suggestions: {
      plan: string;
      currentPrice: number;
      suggestedPrice: number;
      expectedImpact: string;
      reasoning: string;
    }[];
    elasticity: { plan: string; elasticity: number; confidence: number }[];
  }> {
    try {
      const response = await fetch('/api/admin/analytics/price-optimization');
      const result = await response.json();
      
      if (!response.ok) {
        throw new Error(result.error || 'Failed to fetch price optimization');
      }

      return result.data;
    } catch (error) {
      console.error('Error fetching price optimization:', error);
      throw error;
    }
  }
}
