import { RATE_LIMITS } from '@/lib/constants';

interface RateLimitStore {
  [key: string]: {
    count: number;
    resetTime: number;
  };
}
interface RateLimitOptions {
  interval: number; // Time window in milliseconds
  uniqueTokenPerInterval: number; // Maximum unique tokens per interval
}

interface RateLimitResult {
  remaining: number;
  reset: number;
  total: number;
}
// In-memory store (in production, use Redis or similar)
const store: RateLimitStore = {};

/**
 * Rate limiter implementation
 */
export class RateLimiter {
  private windowMs: number;
  public maxRequests: number;

  constructor(windowMs: number, maxRequests: number) {
    this.windowMs = windowMs;
    this.maxRequests = maxRequests;
  }

  /**
   * Check if request is allowed
   */
  isAllowed(identifier: string): {
    allowed: boolean;
    remaining: number;
    resetTime: number;
  } {
    const now = Date.now();
    const key = `${identifier}`;

    if (!store[key] || now > store[key].resetTime) {
      store[key] = {
        count: 1,
        resetTime: now + this.windowMs,
      };

      return {
        allowed: true,
        remaining: this.maxRequests - 1,
        resetTime: store[key].resetTime,
      };
    }

    if (store[key].count >= this.maxRequests) {
      return {
        allowed: false,
        remaining: 0,
        resetTime: store[key].resetTime,
      };
    }

    store[key].count++;

    return {
      allowed: true,
      remaining: this.maxRequests - store[key].count,
      resetTime: store[key].resetTime,
    };
  }

  /**
   * Reset rate limit for identifier
   */
  reset(identifier: string): void {
    delete store[identifier];
  }
}

// Pre-configured rate limiters
export const authRateLimiter = new RateLimiter(
  RATE_LIMITS.AUTH.WINDOW_MS,
  RATE_LIMITS.AUTH.MAX_REQUESTS
);

export const apiRateLimiter = new RateLimiter(
  RATE_LIMITS.API.WINDOW_MS,
  RATE_LIMITS.API.MAX_REQUESTS
);

export const uploadRateLimiter = new RateLimiter(
  RATE_LIMITS.UPLOAD.WINDOW_MS,
  RATE_LIMITS.UPLOAD.MAX_REQUESTS
);

export const downloadRateLimiter = new RateLimiter(
  RATE_LIMITS.DOWNLOAD.WINDOW_MS,
  RATE_LIMITS.DOWNLOAD.MAX_REQUESTS
);

export const shareRateLimiter = new RateLimiter(
  RATE_LIMITS.SHARE.WINDOW_MS,
  RATE_LIMITS.SHARE.MAX_REQUESTS
);
export default function rateLimit(options: RateLimitOptions) {
  const tokenCache = new Map<string, number[]>();

  return {
    check: async (
      res: any,
      limit: number,
      token: string
    ): Promise<RateLimitResult> => {
      const now = Date.now();
      const windowStart = now - options.interval;

      // Get existing timestamps for this token
      const timestamps = tokenCache.get(token) || [];
      
      // Remove timestamps outside the window
      const validTimestamps = timestamps.filter(time => time > windowStart);
      
      // Check if limit exceeded
      if (validTimestamps.length >= limit) {
        const oldestTimestamp = Math.min(...validTimestamps);
        const resetTime = oldestTimestamp + options.interval;
        
        res.setHeader('X-RateLimit-Limit', limit.toString());
        res.setHeader('X-RateLimit-Remaining', '0');
        res.setHeader('X-RateLimit-Reset', Math.ceil(resetTime / 1000).toString());
        
        const error = new Error('Rate limit exceeded');
        (error as any).statusCode = 429;
        throw error;
      }

      // Add current timestamp
      validTimestamps.push(now);
      tokenCache.set(token, validTimestamps);

      // Clean up old entries to prevent memory leaks
      if (tokenCache.size > options.uniqueTokenPerInterval) {
        const sortedEntries = Array.from(tokenCache.entries())
          .sort(([, a], [, b]) => Math.max(...b) - Math.max(...a));
        
        // Keep only the most recent entries
        tokenCache.clear();
        for (const [key, value] of sortedEntries.slice(0, options.uniqueTokenPerInterval)) {
          tokenCache.set(key, value);
        }
      }

      const remaining = limit - validTimestamps.length;
      const reset = Math.ceil((now + options.interval) / 1000);

      res.setHeader('X-RateLimit-Limit', limit.toString());
      res.setHeader('X-RateLimit-Remaining', remaining.toString());
      res.setHeader('X-RateLimit-Reset', reset.toString());

      return {
        remaining,
        reset,
        total: limit
      };
    }
  };
}
/**
 * Rate limit middleware helper
 */
export function createRateLimitMiddleware(rateLimiter: RateLimiter) {
  return (identifier: string) => {
    const result = rateLimiter.isAllowed(identifier);

    if (!result.allowed) {
      const error = new Error('Too many requests');
      (error as any).status = 429;
      (error as any).headers = {
        'X-RateLimit-Limit': rateLimiter.maxRequests,
        'X-RateLimit-Remaining': result.remaining,
        'X-RateLimit-Reset': new Date(result.resetTime).toISOString(),
      };
      throw error;
    }

    return {
      headers: {
        'X-RateLimit-Limit': rateLimiter.maxRequests,
        'X-RateLimit-Remaining': result.remaining,
        'X-RateLimit-Reset': new Date(result.resetTime).toISOString(),
      },
    };
  };
}
