
export * from './memory';
export * from './redis';

// Cache interface that both implementations follow
export interface CacheInterface {
  get<T = any>(key: string): Promise<T | null>;
  set(key: string, value: any, ttl?: number): Promise<void>;
  del(key: string): Promise<void>;
  exists(key: string): Promise<boolean>;
  keys(pattern?: string): Promise<string[]>;
  clear(): Promise<void>;
  ttl(key: string): Promise<number>;
  expire(key: string, ttl: number): Promise<void>;
  increment(key: string, value?: number): Promise<number>;
  decrement(key: string, value?: number): Promise<number>;
  mget<T = any>(keys: string[]): Promise<(T | null)[]>;
  mset(keyValuePairs: Record<string, any>, ttl?: number): Promise<void>;
  flush(): Promise<void>;
  size(): Promise<number>;
  getStats(): Promise<CacheStats>;
}

export interface CacheStats {
  hits: number;
  misses: number;
  keys: number;
  memory?: number;
  hitRate: number;
}

export interface CacheConfig {
  type: 'memory' | 'redis';
  redis?: {
    host?: string;
    port?: number;
    password?: string;
    db?: number;
    maxRetries?: number;
    retryDelayOnFailover?: number;
  };
  memory?: {
    maxSize?: number;
    maxAge?: number;
    checkPeriod?: number;
  };
}

// Cache factory
export function createCache(config: CacheConfig): CacheInterface {
  switch (config.type) {
    case 'redis':
      const { RedisCache } = require('./redis');
      return new RedisCache(config.redis);
    case 'memory':
    default:
      const { MemoryCache } = require('./memory');
      return new MemoryCache(config.memory);
  }
}

// Default cache instance (memory-based for development)
let defaultCache: CacheInterface | null = null;

export function getDefaultCache(): CacheInterface {
  if (!defaultCache) {
    defaultCache = createCache({
      type: process.env.CACHE_TYPE as 'memory' | 'redis' || 'memory',
      redis: {
        host: process.env.REDIS_HOST || 'localhost',
        port: parseInt(process.env.REDIS_PORT || '6379'),
        password: process.env.REDIS_PASSWORD,
        db: parseInt(process.env.REDIS_DB || '0'),
      },
      memory: {
        maxSize: 1000,
        maxAge: 1000 * 60 * 60, // 1 hour
        checkPeriod: 1000 * 60 * 10, // 10 minutes
      },
    });
  }
  return defaultCache;
}

// Convenience functions using default cache
export const cache = {
  get: <T = any>(key: string) => getDefaultCache().get<T>(key),
  set: (key: string, value: any, ttl?: number) => getDefaultCache().set(key, value, ttl),
  del: (key: string) => getDefaultCache().del(key),
  exists: (key: string) => getDefaultCache().exists(key),
  keys: (pattern?: string) => getDefaultCache().keys(pattern),
  clear: () => getDefaultCache().clear(),
  ttl: (key: string) => getDefaultCache().ttl(key),
  expire: (key: string, ttl: number) => getDefaultCache().expire(key, ttl),
  increment: (key: string, value?: number) => getDefaultCache().increment(key, value),
  decrement: (key: string, value?: number) => getDefaultCache().decrement(key, value),
  mget: <T = any>(keys: string[]) => getDefaultCache().mget<T>(keys),
  mset: (keyValuePairs: Record<string, any>, ttl?: number) => getDefaultCache().mset(keyValuePairs, ttl),
  flush: () => getDefaultCache().flush(),
  size: () => getDefaultCache().size(),
  getStats: () => getDefaultCache().getStats(),
};

// Cache decorators for functions
export function cached(ttl: number = 3600) {
  return function (target: any, propertyKey: string, descriptor: PropertyDescriptor) {
    const originalMethod = descriptor.value;

    descriptor.value = async function (...args: any[]) {
      const cacheKey = `${target.constructor.name}:${propertyKey}:${JSON.stringify(args)}`;
      
      try {
        const cached = await cache.get(cacheKey);
        if (cached !== null) {
          return cached;
        }

        const result = await originalMethod.apply(this, args);
        await cache.set(cacheKey, result, ttl);
        return result;
      } catch (error) {
        // Fallback to original method if cache fails
        return originalMethod.apply(this, args);
      }
    };

    return descriptor;
  };
}

// Cache key utilities
export const CacheKeys = {
  user: (userId: string) => `user:${userId}`,
  session: (sessionId: string) => `session:${sessionId}`,
  file: (fileId: string) => `file:${fileId}`,
  folder: (folderId: string) => `folder:${folderId}`,
  stats: (type: string, period: string) => `stats:${type}:${period}`,
  rateLimit: (identifier: string) => `rate_limit:${identifier}`,
  otp: (userId: string) => `otp:${userId}`,
  passwordReset: (token: string) => `password_reset:${token}`,
  emailVerification: (token: string) => `email_verification:${token}`,
  subscription: (subscriptionId: string) => `subscription:${subscriptionId}`,
  usage: (userId: string, period: string) => `usage:${userId}:${period}`,
  analytics: (type: string, date: string) => `analytics:${type}:${date}`,
  thumbnail: (fileId: string, size: string) => `thumbnail:${fileId}:${size}`,
  share: (shareId: string) => `share:${shareId}`,
  apiKey: (keyId: string) => `api_key:${keyId}`,
  settings: (key: string) => `settings:${key}`,
};