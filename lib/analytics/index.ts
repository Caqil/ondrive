/**
 * Analytics Library - Main Export File
 * Centralized analytics functionality for file management platform
 */

export * from './customer-analytics';
export * from './file-analytics';
export * from './performance-analytics';
export * from './revenue-analytics';
export * from './usage-analytics';

// Re-export types for convenience
export type {
  AnalyticsDateRange,
  AnalyticsMetrics,
  AnalyticsQuery,
  UsageRecord,
  ActivityLog
} from '../../types/analytics';

export type {
  AdminDashboardStats
} from '../../types/admin';

// Common analytics utilities
export { formatBytes, formatCurrency, formatNumber } from './utils';
export { createDateRange, getDateRangePresets } from './date-utils';
