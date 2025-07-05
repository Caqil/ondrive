import { formatDistanceToNow, format, parseISO, isValid, differenceInDays, differenceInHours, differenceInMinutes } from 'date-fns';

/**
 * Format date in a human-readable relative format
 */
export function formatRelativeTime(date: Date | string): string {
  const parsedDate = typeof date === 'string' ? parseISO(date) : date;
  
  if (!isValid(parsedDate)) {
    return 'Invalid date';
  }
  
  return formatDistanceToNow(parsedDate, { addSuffix: true });
}

/**
 * Format date in various formats
 */
export function formatDate(date: Date | string, formatString: string = 'MMM d, yyyy'): string {
  const parsedDate = typeof date === 'string' ? parseISO(date) : date;
  
  if (!isValid(parsedDate)) {
    return 'Invalid date';
  }
  
  return format(parsedDate, formatString);
}

/**
 * Format date with time
 */
export function formatDateTime(date: Date | string): string {
  return formatDate(date, 'MMM d, yyyy h:mm a');
}

/**
 * Format date for API (ISO format)
 */
export function formatDateISO(date: Date | string): string {
  const parsedDate = typeof date === 'string' ? parseISO(date) : date;
  
  if (!isValid(parsedDate)) {
    throw new Error('Invalid date');
  }
  
  return parsedDate.toISOString();
}

/**
 * Get time ago in different granularities
 */
export function getTimeAgo(date: Date | string): string {
  const parsedDate = typeof date === 'string' ? parseISO(date) : date;
  const now = new Date();
  
  if (!isValid(parsedDate)) {
    return 'Invalid date';
  }
  
  const days = differenceInDays(now, parsedDate);
  const hours = differenceInHours(now, parsedDate);
  const minutes = differenceInMinutes(now, parsedDate);
  
  if (days > 0) {
    return days === 1 ? '1 day ago' : `${days} days ago`;
  } else if (hours > 0) {
    return hours === 1 ? '1 hour ago' : `${hours} hours ago`;
  } else if (minutes > 0) {
    return minutes === 1 ? '1 minute ago' : `${minutes} minutes ago`;
  } else {
    return 'Just now';
  }
}

/**
 * Check if date is today
 */
export function isToday(date: Date | string): boolean {
  const parsedDate = typeof date === 'string' ? parseISO(date) : date;
  const today = new Date();
  
  return (
    parsedDate.getDate() === today.getDate() &&
    parsedDate.getMonth() === today.getMonth() &&
    parsedDate.getFullYear() === today.getFullYear()
  );
}

/**
 * Check if date is this week
 */
export function isThisWeek(date: Date | string): boolean {
  const parsedDate = typeof date === 'string' ? parseISO(date) : date;
  const now = new Date();
  
  return differenceInDays(now, parsedDate) <= 7;
}

/**
 * Get start and end of day
 */
export function getStartOfDay(date: Date | string): Date {
  const parsedDate = typeof date === 'string' ? parseISO(date) : date;
  const startOfDay = new Date(parsedDate);
  startOfDay.setHours(0, 0, 0, 0);
  return startOfDay;
}

export function getEndOfDay(date: Date | string): Date {
  const parsedDate = typeof date === 'string' ? parseISO(date) : date;
  const endOfDay = new Date(parsedDate);
  endOfDay.setHours(23, 59, 59, 999);
  return endOfDay;
}

/**
 * Get date range for analytics
 */
export function getDateRange(period: 'today' | 'week' | 'month' | 'year' | 'custom', customStart?: Date, customEnd?: Date): {
  start: Date;
  end: Date;
} {
  const now = new Date();
  const start = new Date();
  const end = new Date();
  
  switch (period) {
    case 'today':
      return {
        start: getStartOfDay(now),
        end: getEndOfDay(now),
      };
    
    case 'week':
      start.setDate(now.getDate() - 7);
      return {
        start: getStartOfDay(start),
        end: getEndOfDay(now),
      };
    
    case 'month':
      start.setMonth(now.getMonth() - 1);
      return {
        start: getStartOfDay(start),
        end: getEndOfDay(now),
      };
    
    case 'year':
      start.setFullYear(now.getFullYear() - 1);
      return {
        start: getStartOfDay(start),
        end: getEndOfDay(now),
      };
    
    case 'custom':
      if (!customStart || !customEnd) {
        throw new Error('Custom start and end dates are required for custom period');
      }
      return {
        start: getStartOfDay(customStart),
        end: getEndOfDay(customEnd),
      };
    
    default:
      throw new Error('Invalid period');
  }
}
