import { RATE_LIMITS } from '@/lib/constants';

interface RateLimitStore {
  [key: string]: {
    count: number;
    resetTime: number;
  };
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
