// lib/rate-limit.ts
interface RateLimitOptions {
  interval: number; // Time window in milliseconds
  uniqueTokenPerInterval: number; // Maximum unique tokens per interval
}

interface RateLimitResult {
  remaining: number;
  reset: number;
  total: number;
}

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

// lib/middleware/auth.ts
import type { NextApiRequest, NextApiResponse } from 'next';
import jwt from 'jsonwebtoken';
import { connectDB } from '@/lib/db';
import { User } from '@/models';
import { API_RESPONSE_CODES } from '@/lib/constants';

export interface AuthenticatedRequest extends NextApiRequest {
  user: {
    _id: string;
    email: string;
    role: string;
    emailVerified: boolean;
  };
}

/**
 * JWT Authentication middleware
 */
export function requireAuth(handler: any) {
  return async (req: NextApiRequest, res: NextApiResponse) => {
    try {
      // Get token from Authorization header
      const authHeader = req.headers.authorization;
      if (!authHeader || !authHeader.startsWith('Bearer ')) {
        return res.status(401).json({
          error: 'Authentication required',
          code: API_RESPONSE_CODES.AUTHENTICATION_ERROR
        });
      }

      const token = authHeader.substring(7); // Remove 'Bearer ' prefix

      // Verify JWT token
      let decoded;
      try {
        decoded = jwt.verify(token, process.env.NEXTAUTH_SECRET!) as any;
      } catch (jwtError) {
        return res.status(401).json({
          error: 'Invalid or expired token',
          code: API_RESPONSE_CODES.AUTHENTICATION_ERROR
        });
      }

      await connectDB();

      // Get user from database to ensure they still exist and are active
      const user = await User.findById(decoded.userId);
      if (!user) {
        return res.status(401).json({
          error: 'User not found',
          code: API_RESPONSE_CODES.AUTHENTICATION_ERROR
        });
      }

      // Check if user is banned
      if (user.isBanned) {
        return res.status(403).json({
          error: 'Account suspended',
          code: API_RESPONSE_CODES.AUTHORIZATION_ERROR
        });
      }

      // Check if user is active
      if (!user.isActive) {
        return res.status(403).json({
          error: 'Account deactivated',
          code: API_RESPONSE_CODES.AUTHORIZATION_ERROR
        });
      }

      // Add user to request object
      (req as AuthenticatedRequest).user = {
        _id: user._id.toString(),
        email: user.email,
        role: user.role,
        emailVerified: user.emailVerified
      };

      return handler(req, res);
    } catch (error) {
      console.error('Auth middleware error:', error);
      return res.status(500).json({
        error: 'Internal server error',
        code: API_RESPONSE_CODES.SERVER_ERROR
      });
    }
  };
}

/**
 * Role-based authorization middleware
 */
export function requireRole(requiredRole: string) {
  return (handler: any) => {
    return requireAuth(async (req: AuthenticatedRequest, res: NextApiResponse) => {
      const roleHierarchy: { [key: string]: number } = {
        'user': 0,
        'moderator': 1,
        'admin': 2
      };

      const userRoleLevel = roleHierarchy[req.user.role] ?? 0;
      const requiredRoleLevel = roleHierarchy[requiredRole] ?? 0;

      if (userRoleLevel < requiredRoleLevel) {
        return res.status(403).json({
          error: 'Insufficient permissions',
          code: API_RESPONSE_CODES.AUTHORIZATION_ERROR
        });
      }

      return handler(req, res);
    });
  };
}

/**
 * Email verification middleware
 */
export function requireEmailVerified(handler: any) {
  return requireAuth(async (req: AuthenticatedRequest, res: NextApiResponse) => {
    if (!req.user.emailVerified) {
      return res.status(403).json({
        error: 'Email verification required',
        code: API_RESPONSE_CODES.AUTHORIZATION_ERROR
      });
    }

    return handler(req, res);
  });
}

/**
 * Admin only middleware
 */
export function requireAdmin(handler: any) {
  return requireRole('admin')(handler);
}

/**
 * Optional authentication middleware (doesn't fail if no token)
 */
export function optionalAuth(handler: any) {
  return async (req: NextApiRequest, res: NextApiResponse) => {
    try {
      // Get token from Authorization header
      const authHeader = req.headers.authorization;
      if (!authHeader || !authHeader.startsWith('Bearer ')) {
        // No auth provided, continue without user
        return handler(req, res);
      }

      const token = authHeader.substring(7); // Remove 'Bearer ' prefix

      // Verify JWT token
      let decoded;
      try {
        decoded = jwt.verify(token, process.env.NEXTAUTH_SECRET!) as any;
      } catch (jwtError) {
        // Invalid token, continue without user
        return handler(req, res);
      }

      await connectDB();

      // Get user from database
      const user = await User.findById(decoded.userId);
      if (user && !user.isBanned && user.isActive) {
        // Add user to request object
        (req as AuthenticatedRequest).user = {
          _id: user._id.toString(),
          email: user.email,
          role: user.role,
          emailVerified: user.emailVerified
        };
      }

      return handler(req, res);
    } catch (error) {
      console.error('Optional auth middleware error:', error);
      // Continue without user on error
      return handler(req, res);
    }
  };
}