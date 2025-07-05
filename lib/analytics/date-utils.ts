/**
 * Date Utilities for Analytics
 */

import { AnalyticsDateRange } from '../../types';

/**
 * Create date range object
 */
export function createDateRange(start: Date, end: Date): AnalyticsDateRange {
  return { start, end };
}

/**
 * Get predefined date range presets
 */
export function getDateRangePresets(): Record<string, AnalyticsDateRange> {
  const now = new Date();
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
  
  return {
    today: createDateRange(today, now),
    yesterday: createDateRange(
      new Date(today.getTime() - 24 * 60 * 60 * 1000),
      today
    ),
    last7Days: createDateRange(
      new Date(today.getTime() - 7 * 24 * 60 * 60 * 1000),
      now
    ),
    last30Days: createDateRange(
      new Date(today.getTime() - 30 * 24 * 60 * 60 * 1000),
      now
    ),
    last90Days: createDateRange(
      new Date(today.getTime() - 90 * 24 * 60 * 60 * 1000),
      now
    ),
    thisWeek: createDateRange(
      getStartOfWeek(today),
      now
    ),
    lastWeek: createDateRange(
      getStartOfWeek(new Date(today.getTime() - 7 * 24 * 60 * 60 * 1000)),
      getEndOfWeek(new Date(today.getTime() - 7 * 24 * 60 * 60 * 1000))
    ),
    thisMonth: createDateRange(
      new Date(today.getFullYear(), today.getMonth(), 1),
      now
    ),
    lastMonth: createDateRange(
      new Date(today.getFullYear(), today.getMonth() - 1, 1),
      new Date(today.getFullYear(), today.getMonth(), 0, 23, 59, 59)
    ),
    thisQuarter: createDateRange(
      getStartOfQuarter(today),
      now
    ),
    lastQuarter: createDateRange(
      getStartOfQuarter(new Date(today.getFullYear(), today.getMonth() - 3, 1)),
      getEndOfQuarter(new Date(today.getFullYear(), today.getMonth() - 3, 1))
    ),
    thisYear: createDateRange(
      new Date(today.getFullYear(), 0, 1),
      now
    ),
    lastYear: createDateRange(
      new Date(today.getFullYear() - 1, 0, 1),
      new Date(today.getFullYear() - 1, 11, 31, 23, 59, 59)
    ),
  };
}

/**
 * Get start of week (Monday)
 */
function getStartOfWeek(date: Date): Date {
  const d = new Date(date);
  const day = d.getDay();
  const diff = d.getDate() - day + (day === 0 ? -6 : 1); // Adjust when day is Sunday
  return new Date(d.setDate(diff));
}

/**
 * Get end of week (Sunday)
 */
function getEndOfWeek(date: Date): Date {
  const d = getStartOfWeek(date);
  return new Date(d.setDate(d.getDate() + 6));
}

/**
 * Get start of quarter
 */
function getStartOfQuarter(date: Date): Date {
  const quarter = Math.floor(date.getMonth() / 3);
  return new Date(date.getFullYear(), quarter * 3, 1);
}

/**
 * Get end of quarter
 */
function getEndOfQuarter(date: Date): Date {
  const quarter = Math.floor(date.getMonth() / 3);
  return new Date(date.getFullYear(), quarter * 3 + 3, 0, 23, 59, 59);
}

/**
 * Format date for API
 */
export function formatDateForAPI(date: Date): string {
  return date.toISOString();
}

/**
 * Parse date from API
 */
export function parseDateFromAPI(dateString: string): Date {
  return new Date(dateString);
}

/**
 * Get date range duration in days
 */
export function getDateRangeDuration(dateRange: AnalyticsDateRange): number {
  return Math.ceil((dateRange.end.getTime() - dateRange.start.getTime()) / (1000 * 60 * 60 * 24));
}

/**
 * Check if date range is valid
 */
export function isValidDateRange(dateRange: AnalyticsDateRange): boolean {
  return dateRange.start < dateRange.end;
}

/**
 * Adjust date range to fit within bounds
 */
export function adjustDateRange(
  dateRange: AnalyticsDateRange,
  maxDays?: number,
  minDate?: Date,
  maxDate?: Date
): AnalyticsDateRange {
  let { start, end } = dateRange;
  
  // Ensure start is not before minDate
  if (minDate && start < minDate) {
    start = minDate;
  }
  
  // Ensure end is not after maxDate
  if (maxDate && end > maxDate) {
    end = maxDate;
  }
  
  // Ensure range doesn't exceed maxDays
  if (maxDays) {
    const duration = getDateRangeDuration({ start, end });
    if (duration > maxDays) {
      start = new Date(end.getTime() - maxDays * 24 * 60 * 60 * 1000);
    }
  }
  
  return { start, end };
}