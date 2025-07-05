import { CacheInterface, CacheStats } from './index';

// Redis client interface (compatible with ioredis, node-redis, etc.)
interface RedisClient {
  get(key: string): Promise<string | null>;
  set(key: string, value: string, mode?: string, duration?: number): Promise<string | null>;
  setex(key: string, seconds: number, value: string): Promise<string>;
  del(...keys: string[]): Promise<number>;
  exists(...keys: string[]): Promise<number>;
  keys(pattern: string): Promise<string[]>;
  flushdb(): Promise<string>;
  ttl(key: string): Promise<number>;
  expire(key: string, seconds: number): Promise<number>;
  incr(key: string): Promise<number>;
  incrby(key: string, increment: number): Promise<number>;
  decr(key: string): Promise<number>;
  decrby(key: string, decrement: number): Promise<number>;
  mget(...keys: string[]): Promise<(string | null)[]>;
  mset(...keyValues: string[]): Promise<string>;
  info(section?: string): Promise<string>;
  ping(): Promise<string>;
  quit(): Promise<string>;
  on(event: string, listener: (...args: any[]) => void): void;
}

interface RedisConfig {
  host?: string;
  port?: number;
  password?: string;
  db?: number;
  maxRetries?: number;
  retryDelayOnFailover?: number;
  keyPrefix?: string;
}

export class RedisCache implements CacheInterface {
  private client: RedisClient | null = null;
  private config: RedisConfig;
  private stats = {
    hits: 0,
    misses: 0,
    errors: 0,
  };
  private connected = false;
  private keyPrefix: string;

  constructor(config: RedisConfig = {}) {
    this.config = {
      host: config.host || 'localhost',
      port: config.port || 6379,
      password: config.password,
      db: config.db || 0,
      maxRetries: config.maxRetries || 3,
      retryDelayOnFailover: config.retryDelayOnFailover || 100,
      keyPrefix: config.keyPrefix || 'cache:',
    };
    this.keyPrefix = this.config.keyPrefix!;
  }

  private async ensureConnection(): Promise<void> {
    if (this.client && this.connected) return;

    try {
      // Try to import Redis client
      let Redis;
      try {
        // Try ioredis first (recommended)
        Redis = require('ioredis');
      } catch {
        try {
          // Fallback to node-redis
          Redis = require('redis');
        } catch {
          throw new Error('No Redis client found. Install ioredis or redis package.');
        }
      }

      // Create client based on the package
      if (Redis.createClient) {
        // node-redis v4+
        this.client = Redis.createClient({
          socket: {
            host: this.config.host,
            port: this.config.port,
          },
          password: this.config.password,
          database: this.config.db,
        });
        await (this.client as any).connect();
      } else {
        // ioredis
        this.client = new Redis({
          host: this.config.host,
          port: this.config.port,
          password: this.config.password,
          db: this.config.db,
          maxRetriesPerRequest: this.config.maxRetries,
          retryDelayOnFailover: this.config.retryDelayOnFailover,
        });
      }

      this.setupEventHandlers();
      this.connected = true;
    } catch (error) {
      this.stats.errors++;
      throw new Error(`Failed to connect to Redis: ${error}`);
    }
  }

  private setupEventHandlers(): void {
    if (!this.client) return;

    this.client.on('connect', () => {
      this.connected = true;
    });

    this.client.on('error', (error) => {
      this.stats.errors++;
      console.error('Redis error:', error);
      this.connected = false;
    });

    this.client.on('close', () => {
      this.connected = false;
    });
  }

  private prefixKey(key: string): string {
    return `${this.keyPrefix}${key}`;
  }

  private unprefixKey(key: string): string {
    return key.startsWith(this.keyPrefix) ? key.slice(this.keyPrefix.length) : key;
  }

  async get<T = any>(key: string): Promise<T | null> {
    try {
      await this.ensureConnection();
      if (!this.client) throw new Error('Redis client not available');

      const value = await this.client.get(this.prefixKey(key));
      
      if (value === null) {
        this.stats.misses++;
        return null;
      }

      this.stats.hits++;
      return JSON.parse(value) as T;
    } catch (error) {
      this.stats.errors++;
      console.error('Redis get error:', error);
      return null;
    }
  }

  async set(key: string, value: any, ttl?: number): Promise<void> {
    try {
      await this.ensureConnection();
      if (!this.client) throw new Error('Redis client not available');

      const serialized = JSON.stringify(value);
      const prefixedKey = this.prefixKey(key);

      if (ttl) {
        await this.client.setex(prefixedKey, ttl, serialized);
      } else {
        await this.client.set(prefixedKey, serialized);
      }
    } catch (error) {
      this.stats.errors++;
      console.error('Redis set error:', error);
      throw error;
    }
  }

  async del(key: string): Promise<void> {
    try {
      await this.ensureConnection();
      if (!this.client) throw new Error('Redis client not available');

      await this.client.del(this.prefixKey(key));
    } catch (error) {
      this.stats.errors++;
      console.error('Redis del error:', error);
      throw error;
    }
  }

  async exists(key: string): Promise<boolean> {
    try {
      await this.ensureConnection();
      if (!this.client) throw new Error('Redis client not available');

      const result = await this.client.exists(this.prefixKey(key));
      return result > 0;
    } catch (error) {
      this.stats.errors++;
      console.error('Redis exists error:', error);
      return false;
    }
  }

  async keys(pattern?: string): Promise<string[]> {
    try {
      await this.ensureConnection();
      if (!this.client) throw new Error('Redis client not available');

      const searchPattern = pattern 
        ? `${this.keyPrefix}${pattern}`
        : `${this.keyPrefix}*`;
      
      const keys = await this.client.keys(searchPattern);
      return keys.map(key => this.unprefixKey(key));
    } catch (error) {
      this.stats.errors++;
      console.error('Redis keys error:', error);
      return [];
    }
  }

  async clear(): Promise<void> {
    try {
      await this.ensureConnection();
      if (!this.client) throw new Error('Redis client not available');

      const keys = await this.client.keys(`${this.keyPrefix}*`);
      if (keys.length > 0) {
        await this.client.del(...keys);
      }
    } catch (error) {
      this.stats.errors++;
      console.error('Redis clear error:', error);
      throw error;
    }
  }

  async ttl(key: string): Promise<number> {
    try {
      await this.ensureConnection();
      if (!this.client) throw new Error('Redis client not available');

      return await this.client.ttl(this.prefixKey(key));
    } catch (error) {
      this.stats.errors++;
      console.error('Redis ttl error:', error);
      return -1;
    }
  }

  async expire(key: string, ttl: number): Promise<void> {
    try {
      await this.ensureConnection();
      if (!this.client) throw new Error('Redis client not available');

      await this.client.expire(this.prefixKey(key), ttl);
    } catch (error) {
      this.stats.errors++;
      console.error('Redis expire error:', error);
      throw error;
    }
  }

  async increment(key: string, value: number = 1): Promise<number> {
    try {
      await this.ensureConnection();
      if (!this.client) throw new Error('Redis client not available');

      const prefixedKey = this.prefixKey(key);
      if (value === 1) {
        return await this.client.incr(prefixedKey);
      } else {
        return await this.client.incrby(prefixedKey, value);
      }
    } catch (error) {
      this.stats.errors++;
      console.error('Redis increment error:', error);
      throw error;
    }
  }

  async decrement(key: string, value: number = 1): Promise<number> {
    try {
      await this.ensureConnection();
      if (!this.client) throw new Error('Redis client not available');

      const prefixedKey = this.prefixKey(key);
      if (value === 1) {
        return await this.client.decr(prefixedKey);
      } else {
        return await this.client.decrby(prefixedKey, value);
      }
    } catch (error) {
      this.stats.errors++;
      console.error('Redis decrement error:', error);
      throw error;
    }
  }

  async mget<T = any>(keys: string[]): Promise<(T | null)[]> {
    try {
      await this.ensureConnection();
      if (!this.client) throw new Error('Redis client not available');

      const prefixedKeys = keys.map(key => this.prefixKey(key));
      const values = await this.client.mget(...prefixedKeys);
      
      return values.map(value => {
        if (value === null) {
          this.stats.misses++;
          return null;
        }
        this.stats.hits++;
        return JSON.parse(value) as T;
      });
    } catch (error) {
      this.stats.errors++;
      console.error('Redis mget error:', error);
      return keys.map(() => null);
    }
  }

  async mset(keyValuePairs: Record<string, any>, ttl?: number): Promise<void> {
    try {
      await this.ensureConnection();
      if (!this.client) throw new Error('Redis client not available');

      const args: string[] = [];
      for (const [key, value] of Object.entries(keyValuePairs)) {
        args.push(this.prefixKey(key), JSON.stringify(value));
      }

      await this.client.mset(...args);

      // Set TTL for each key if specified
      if (ttl) {
        const expirePromises = Object.keys(keyValuePairs).map(key =>
          this.client!.expire(this.prefixKey(key), ttl)
        );
        await Promise.all(expirePromises);
      }
    } catch (error) {
      this.stats.errors++;
      console.error('Redis mset error:', error);
      throw error;
    }
  }

  async flush(): Promise<void> {
    try {
      await this.ensureConnection();
      if (!this.client) throw new Error('Redis client not available');

      await this.client.flushdb();
    } catch (error) {
      this.stats.errors++;
      console.error('Redis flush error:', error);
      throw error;
    }
  }

  async size(): Promise<number> {
    try {
      await this.ensureConnection();
      if (!this.client) throw new Error('Redis client not available');

      const keys = await this.client.keys(`${this.keyPrefix}*`);
      return keys.length;
    } catch (error) {
      this.stats.errors++;
      console.error('Redis size error:', error);
      return 0;
    }
  }

  async getStats(): Promise<CacheStats> {
    const hitRate = this.stats.hits + this.stats.misses > 0
      ? this.stats.hits / (this.stats.hits + this.stats.misses)
      : 0;

    try {
      const size = await this.size();
      return {
        hits: this.stats.hits,
        misses: this.stats.misses,
        keys: size,
        hitRate: Math.round(hitRate * 100) / 100,
      };
    } catch (error) {
      return {
        hits: this.stats.hits,
        misses: this.stats.misses,
        keys: 0,
        hitRate: Math.round(hitRate * 100) / 100,
      };
    }
  }

  // Redis-specific methods
  async ping(): Promise<boolean> {
    try {
      await this.ensureConnection();
      if (!this.client) return false;

      await this.client.ping();
      return true;
    } catch (error) {
      return false;
    }
  }

  async getInfo(): Promise<Record<string, string>> {
    try {
      await this.ensureConnection();
      if (!this.client) throw new Error('Redis client not available');

      const info = await this.client.info();
      const result: Record<string, string> = {};
      
      info.split('\r\n').forEach(line => {
        const [key, value] = line.split(':');
        if (key && value) {
          result[key] = value;
        }
      });

      return result;
    } catch (error) {
      this.stats.errors++;
      console.error('Redis info error:', error);
      return {};
    }
  }

  // Cleanup
  async disconnect(): Promise<void> {
    try {
      if (this.client) {
        await this.client.quit();
        this.client = null;
        this.connected = false;
      }
    } catch (error) {
      console.error('Redis disconnect error:', error);
    }
  }
}

// Export singleton instance for convenience
let redisCache: RedisCache | null = null;

export function getRedisCache(config?: RedisConfig): RedisCache {
  if (!redisCache) {
    redisCache = new RedisCache(config);
  }
  return redisCache;
}