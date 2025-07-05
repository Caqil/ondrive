/**
 * Format file size in human readable format
 */
export function formatBytes(bytes: number, decimals: number = 1): string {
  if (bytes === 0) return '0 B';
  
  const k = 1024;
  const dm = decimals < 0 ? 0 : decimals;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];
  
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  
  return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
}

/**
 * Format number with thousand separators
 */
export function formatNumber(num: number, locale: string = 'en-US'): string {
  return new Intl.NumberFormat(locale).format(num);
}

/**
 * Format currency
 */
export function formatCurrency(amount: number, currency: string = 'USD', locale: string = 'en-US'): string {
  return new Intl.NumberFormat(locale, {
    style: 'currency',
    currency: currency,
    minimumFractionDigits: 2,
  }).format(amount / 100); // Assuming amount is in cents
}

/**
 * Format percentage
 */
export function formatPercentage(value: number, decimals: number = 1): string {
  return `${(value * 100).toFixed(decimals)}%`;
}

/**
 * Format duration in seconds to human readable format
 */
export function formatDuration(seconds: number): string {
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  const remainingSeconds = Math.floor(seconds % 60);
  
  if (hours > 0) {
    return `${hours}:${minutes.toString().padStart(2, '0')}:${remainingSeconds.toString().padStart(2, '0')}`;
  } else {
    return `${minutes}:${remainingSeconds.toString().padStart(2, '0')}`;
  }
}

/**
 * Format speed (bytes per second)
 */
export function formatSpeed(bytesPerSecond: number): string {
  return `${formatBytes(bytesPerSecond)}/s`;
}

/**
 * Format transfer progress
 */
export function formatProgress(current: number, total: number): string {
  const percentage = total > 0 ? (current / total) * 100 : 0;
  return `${formatBytes(current)} / ${formatBytes(total)} (${percentage.toFixed(1)}%)`;
}

/**
 * Format phone number
 */
export function formatPhoneNumber(phone: string, countryCode: string = 'US'): string {
  try {
    // This would typically use a library like libphonenumber-js
    // For now, just return formatted US phone numbers
    const cleaned = phone.replace(/\D/g, '');
    
    if (countryCode === 'US' && cleaned.length === 10) {
      return `(${cleaned.slice(0, 3)}) ${cleaned.slice(3, 6)}-${cleaned.slice(6)}`;
    }
    
    return phone;
  } catch {
    return phone;
  }
}

/**
 * Truncate text with ellipsis
 */
export function truncateText(text: string, maxLength: number, suffix: string = '...'): string {
  if (text.length <= maxLength) {
    return text;
  }
  
  return text.slice(0, maxLength - suffix.length) + suffix;
}

/**
 * Format name initials
 */
export function formatInitials(name: string): string {
  return name
    .split(' ')
    .map(word => word.charAt(0))
    .join('')
    .toUpperCase()
    .slice(0, 2);
}

/**
 * Format filename for display
 */
export function formatFilename(filename: string, maxLength: number = 30): string {
  if (filename.length <= maxLength) {
    return filename;
  }
  
  const extension = filename.split('.').pop();
  const nameWithoutExt = filename.slice(0, filename.lastIndexOf('.'));
  const availableLength = maxLength - (extension ? extension.length + 4 : 3); // +4 for "..." and "."
  
  if (availableLength <= 0) {
    return '...' + (extension ? `.${extension}` : '');
  }
  
  return nameWithoutExt.slice(0, availableLength) + '...' + (extension ? `.${extension}` : '');
}