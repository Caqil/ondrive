import { CacheInterface, CacheStats } from './index';

interface CacheEntry {
  value: any;
  expires: number;
  accessCount: number;
  lastAccess: number;
}

interface MemoryCacheConfig {
  maxSize?: number;
  maxAge?: number;
  checkPeriod?: number;
}

export class MemoryCache implements CacheInterface {
  private cache = new Map<string, CacheEntry>();
  private stats = {
    hits: 0,
    misses: 0,
    sets: 0,
    deletes: 0,
    evictions: 0,
  };
  private maxSize: number;
  private maxAge: number;
  private cleanupInterval: NodeJS.Timeout | null = null;

  constructor(config: MemoryCacheConfig = {}) {
    this.maxSize = config.maxSize || 1000;
    this.maxAge = config.maxAge || 1000 * 60 * 60; // 1 hour default
    
    // Start cleanup interval
    const checkPeriod = config.checkPeriod || 1000 * 60 * 10; // 10 minutes
    this.cleanupInterval = setInterval(() => {
      this.cleanup();
    }, checkPeriod);
  }

  async get<T = any>(key: string): Promise<T | null> {
    const entry = this.cache.get(key);
    
    if (!entry) {
      this.stats.misses++;
      return null;
    }

    // Check if expired
    if (Date.now() > entry.expires) {
      this.cache.delete(key);
      this.stats.misses++;
      return null;
    }

    // Update access stats
    entry.accessCount++;
    entry.lastAccess = Date.now();
    this.stats.hits++;

    return entry.value as T;
  }

  async set(key: string, value: any, ttl?: number): Promise<void> {
    const expires = Date.now() + (ttl ? ttl * 1000 : this.maxAge);
    
    // Check if we need to evict entries
    if (this.cache.size >= this.maxSize && !this.cache.has(key)) {
      this.evictLRU();
    }

    this.cache.set(key, {
      value,
      expires,
      accessCount: 0,
      lastAccess: Date.now(),
    });

    this.stats.sets++;
  }

  async del(key: string): Promise<void> {
    if (this.cache.delete(key)) {
      this.stats.deletes++;
    }
  }

  async exists(key: string): Promise<boolean> {
    const entry = this.cache.get(key);
    if (!entry) return false;
    
    // Check if expired
    if (Date.now() > entry.expires) {
      this.cache.delete(key);
      return false;
    }
    
    return true;
  }

  async keys(pattern?: string): Promise<string[]> {
    const allKeys = Array.from(this.cache.keys());
    
    if (!pattern) {
      return allKeys.filter(key => {
        const entry = this.cache.get(key);
        return entry && Date.now() <= entry.expires;
      });
    }

    // Simple glob pattern matching
    const regex = new RegExp(
      pattern.replace(/\*/g, '.*').replace(/\?/g, '.')
    );

    return allKeys.filter(key => {
      const entry = this.cache.get(key);
      if (!entry || Date.now() > entry.expires) return false;
      return regex.test(key);
    });
  }

  async clear(): Promise<void> {
    this.cache.clear();
    this.resetStats();
  }

  async ttl(key: string): Promise<number> {
    const entry = this.cache.get(key);
    if (!entry) return -1;
    
    const remaining = Math.max(0, entry.expires - Date.now());
    return Math.ceil(remaining / 1000);
  }

  async expire(key: string, ttl: number): Promise<void> {
    const entry = this.cache.get(key);
    if (entry) {
      entry.expires = Date.now() + (ttl * 1000);
    }
  }

  async increment(key: string, value: number = 1): Promise<number> {
    const current = await this.get<number>(key) || 0;
    const newValue = current + value;
    await this.set(key, newValue);
    return newValue;
  }

  async decrement(key: string, value: number = 1): Promise<number> {
    return this.increment(key, -value);
  }

  async mget<T = any>(keys: string[]): Promise<(T | null)[]> {
    return Promise.all(keys.map(key => this.get<T>(key)));
  }

  async mset(keyValuePairs: Record<string, any>, ttl?: number): Promise<void> {
    const promises = Object.entries(keyValuePairs).map(([key, value]) =>
      this.set(key, value, ttl)
    );
    await Promise.all(promises);
  }

  async flush(): Promise<void> {
    this.cache.clear();
    this.resetStats();
  }

  async size(): Promise<number> {
    // Remove expired entries first
    this.cleanup();
    return this.cache.size;
  }

  async getStats(): Promise<CacheStats> {
    const hitRate = this.stats.hits + this.stats.misses > 0
      ? this.stats.hits / (this.stats.hits + this.stats.misses)
      : 0;

    return {
      hits: this.stats.hits,
      misses: this.stats.misses,
      keys: this.cache.size,
      memory: this.getMemoryUsage(),
      hitRate: Math.round(hitRate * 100) / 100,
    };
  }

  // Internal methods
  private evictLRU(): void {
    let oldestKey: string | null = null;
    let oldestAccess = Date.now();

    for (const [key, entry] of this.cache.entries()) {
      if (entry.lastAccess < oldestAccess) {
        oldestAccess = entry.lastAccess;
        oldestKey = key;
      }
    }

    if (oldestKey) {
      this.cache.delete(oldestKey);
      this.stats.evictions++;
    }
  }

  private cleanup(): void {
    const now = Date.now();
    const expiredKeys: string[] = [];

    for (const [key, entry] of this.cache.entries()) {
      if (now > entry.expires) {
        expiredKeys.push(key);
      }
    }

    expiredKeys.forEach(key => this.cache.delete(key));
  }

  private resetStats(): void {
    this.stats = {
      hits: 0,
      misses: 0,
      sets: 0,
      deletes: 0,
      evictions: 0,
    };
  }

  private getMemoryUsage(): number {
    // Rough estimation of memory usage
    let size = 0;
    for (const [key, entry] of this.cache.entries()) {
      size += key.length * 2; // String overhead
      size += JSON.stringify(entry.value).length * 2;
      size += 32; // Entry overhead
    }
    return size;
  }

  // Cleanup on destruction
  destroy(): void {
    if (this.cleanupInterval) {
      clearInterval(this.cleanupInterval);
      this.cleanupInterval = null;
    }
    this.cache.clear();
  }
}

// Export singleton instance for convenience
export const memoryCache = new MemoryCache();
