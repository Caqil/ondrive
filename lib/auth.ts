// lib/auth.ts
import { NextAuthOptions } from 'next-auth';
import { MongoDBAdapter } from '@next-auth/mongodb-adapter';
import GoogleProvider from 'next-auth/providers/google';
import GitHubProvider from 'next-auth/providers/github';
import CredentialsProvider from 'next-auth/providers/credentials';
import bcrypt from 'bcryptjs';
import { User } from '@/models';
import { connectDB } from '@/lib/db';
import type { User as UserType } from '@/types';

// Ensure UserType includes 'admin' in the role type
import { generateSecretKey, verifyTOTP } from '@/lib/crypto';
import { MongoClient } from 'mongodb';

// Password validation regex
export const PASSWORD_REGEX = {
  minLength: /.{8,}/,
  hasNumber: /\d/,
  hasSymbol: /[!@#$%^&*(),.?":{}|<>]/,
  hasUppercase: /[A-Z]/,
  hasLowercase: /[a-z]/,
};

// Email validation regex
export const EMAIL_REGEX = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

/**
 * Validate password strength based on requirements
 */
export function validatePassword(password: string): {
  isValid: boolean;
  errors: string[];
} {
  const errors: string[] = [];

  if (!PASSWORD_REGEX.minLength.test(password)) {
    errors.push('Password must be at least 8 characters long');
  }
  if (!PASSWORD_REGEX.hasNumber.test(password)) {
    errors.push('Password must contain at least one number');
  }
  if (!PASSWORD_REGEX.hasSymbol.test(password)) {
    errors.push('Password must contain at least one special character');
  }
  if (!PASSWORD_REGEX.hasUppercase.test(password)) {
    errors.push('Password must contain at least one uppercase letter');
  }
  if (!PASSWORD_REGEX.hasLowercase.test(password)) {
    errors.push('Password must contain at least one lowercase letter');
  }

  return {
    isValid: errors.length === 0,
    errors,
  };
}

/**
 * Hash password using bcrypt
 */
export async function hashPassword(password: string): Promise<string> {
  const saltRounds = 12;
  return bcrypt.hash(password, saltRounds);
}

/**
 * Verify password against hash
 */
export async function verifyPassword(password: string, hash: string): Promise<boolean> {
  return bcrypt.compare(password, hash);
}

/**
 * Generate random token for email verification, password reset, etc.
 */
export function generateToken(): string {
  return generateSecretKey(32);
}

/**
 * Check if user has required role
 */
export function hasRole(user: UserType | null, requiredRole: string): boolean {
  if (!user) return false;
  
  const roleHierarchy = {
    'user': 0,
    'moderator': 1,
    'viewer': 2,
    'admin': 3,
  };

  const userRoleLevel = roleHierarchy[user.role as keyof typeof roleHierarchy] ?? 0;
  const requiredRoleLevel = roleHierarchy[requiredRole as keyof typeof roleHierarchy] ?? 0;

  return userRoleLevel >= requiredRoleLevel;
}

/**
 * Check if user has permission for specific action
 */
export function hasPermission(
  user: UserType | null,
  action: string,
  resource?: any
): boolean {
  if (!user) return false;

  // Admin can do everything
  if (user.role === 'admin') return true;

  // Check if user is banned
  if (user.isBanned) return false;

  // Check email verification for sensitive actions
  const sensitiveActions = ['upload', 'share', 'delete', 'admin-action'];
  if (sensitiveActions.includes(action) && !user.emailVerified) {
    return false;
  }

  // Resource-specific permissions
  if (resource) {
    // Check if user owns the resource
    if (resource.owner && resource.owner.toString() === user._id.toString()) {
      return true;
    }

    // Check team permissions
    if (resource.team && user.currentTeam?.toString() === resource.team.toString()) {
      // Team-based permission logic here
      return true;
    }
  }

  // Role-based permissions
  switch (action) {
    case 'read':
      return true; // All authenticated users can read
    case 'upload':
    case 'create':
      return user.role !== 'viewer';
    case 'update':
    case 'rename':
      return user.role !== 'viewer';
    case 'delete':
      return ['admin', 'moderator'].includes(user.role);
    case 'share':
      return user.role !== 'viewer';
    default:
      return false;
  }
}

/**
 * Get user storage quota based on subscription
 */
export function getUserStorageQuota(user: UserType): number {
  // Default quota for trial/free users
  let quota = 15 * 1024 * 1024 * 1024; // 15GB

  // Admin users get unlimited storage
  if (user.role === 'admin') {
    return Number.MAX_SAFE_INTEGER;
  }

  // Subscription-based quotas
  switch (user.subscriptionStatus) {
    case 'active':
      // This would be determined by the actual subscription plan
      quota = 100 * 1024 * 1024 * 1024; // 100GB for active subscription
      break;
    case 'trial':
      quota = 50 * 1024 * 1024 * 1024; // 50GB for trial
      break;
    case 'expired':
    case 'cancelled':
      quota = 5 * 1024 * 1024 * 1024; // 5GB for expired/cancelled
      break;
    default:
      quota = 15 * 1024 * 1024 * 1024; // 15GB default
  }

  return Math.max(quota, user.storageQuota || quota);
}

/**
 * Check if user has exceeded storage quota
 */
export function hasExceededStorageQuota(user: UserType): boolean {
  const quota = getUserStorageQuota(user);
  return user.storageUsed >= quota;
}

/**
 * Generate 2FA secret for user
 */
export function generate2FASecret(): {
  secret: string;
  qrCode: string;
  backupCodes: string[];
} {
  const secret = generateSecretKey(20);
  
  // Generate QR code URL for authenticator apps
  const serviceName = process.env.NEXT_PUBLIC_APP_NAME || 'Drive Clone';
  const qrCode = `otpauth://totp/${serviceName}?secret=${secret}&issuer=${serviceName}`;
  
  // Generate backup codes
  const backupCodes = Array.from({ length: 8 }, () => generateSecretKey(8));

  return {
    secret,
    qrCode,
    backupCodes,
  };
}

/**
 * Verify 2FA token
 */
export function verify2FAToken(secret: string, token: string): boolean {
  return verifyTOTP(secret, token);
}
/**
 * NextAuth configuration
 */
export const authOptions: NextAuthOptions = {
  adapter: MongoDBAdapter(MongoClient.connect(process.env.MONGODB_URI!)),
  providers: [
    // OAuth Providers
    GoogleProvider({
      clientId: process.env.GOOGLE_CLIENT_ID!,
      clientSecret: process.env.GOOGLE_CLIENT_SECRET!,
    }),
    GitHubProvider({
      clientId: process.env.GITHUB_CLIENT_ID!,
      clientSecret: process.env.GITHUB_CLIENT_SECRET!,
    }),
    
    // Credentials Provider
    CredentialsProvider({
      name: 'credentials',
      credentials: {
        email: { label: 'Email', type: 'email' },
        password: { label: 'Password', type: 'password' },
        code: { label: '2FA Code', type: 'text' },
      },
      async authorize(credentials) {
        if (!credentials?.email || !credentials?.password) {
          throw new Error('Email and password are required');
        }

        await connectDB();

        // Find user with password
        const user = await User.findOne({ 
          email: credentials.email.toLowerCase() 
        }).select('+password +twoFactorSecret');

        if (!user) {
          throw new Error('Invalid email or password');
        }

        // Check if account is banned
        if (user.isBanned) {
          throw new Error('Account has been suspended');
        }

        // Verify password
        const isValidPassword = await verifyPassword(credentials.password, user.password);
        if (!isValidPassword) {
          throw new Error('Invalid email or password');
        }

        // Check 2FA if enabled
        if (user.twoFactorEnabled) {
          if (!credentials.code) {
            throw new Error('2FA code is required');
          }

          const isValid2FA = verify2FAToken(user.twoFactorSecret, credentials.code);
          if (!isValid2FA) {
            throw new Error('Invalid 2FA code');
          }
        }

        // Update last login
        await User.findByIdAndUpdate(user._id, {
          lastLogin: new Date(),
        });

        return {
          id: user._id.toString(),
          email: user.email,
          name: user.name,
          image: user.avatar,
          role: user.role,
        };
      },
    }),
  ],
  
  session: {
    strategy: 'jwt',
    maxAge: 30 * 24 * 60 * 60, // 30 days
  },
  
  jwt: {
    maxAge: 30 * 24 * 60 * 60, // 30 days
  },
  
  pages: {
    signIn: '/login',
    newUser: '/register',
    error: '/auth/error',
    verifyRequest: '/auth/verify-request',
  },
  
  callbacks: {
    async signIn({ user, account, profile }) {
      if (account?.provider === 'google' || account?.provider === 'github') {
        await connectDB();
        
        // Check if user exists
        const existingUser = await User.findOne({ email: user.email });
        
        if (existingUser) {
          // Update OAuth provider info
          const providerKey = account.provider as 'google' | 'github';
          const providerData = account.provider === 'google' 
            ? { id: user.id, email: user.email }
            : { id: user.id, username: (profile as any)?.login };

          await User.findByIdAndUpdate(existingUser._id, {
            [`providers.${providerKey}`]: providerData,
            emailVerified: true, // OAuth emails are considered verified
            lastLogin: new Date(),
          });
        } else {
          // Create new user
          const newUser = new User({
            email: user.email,
            name: user.name,
            avatar: user.image,
            emailVerified: true,
            providers: {
              [account.provider]: account.provider === 'google'
                ? { id: user.id, email: user.email }
                : { id: user.id, username: (profile as any)?.login },
            },
          });
          
          await newUser.save();
        }
      }
      
      return true;
    },
    
    async jwt({ token, user, account }) {
      if (user) {
        token.role = (user as any).role;
      }
      
      // Refresh user data from database on each request
      if (token.email) {
        await connectDB();
        const dbUser = await User.findOne({ email: token.email });
        
        if (dbUser) {
          token.role = dbUser.role;
          token.name = dbUser.name;
          token.picture = dbUser.avatar;
          token.emailVerified = dbUser.emailVerified;
          token.isBanned = dbUser.isBanned;
        }
      }
      
      return token;
    },
    
    async session({ session, token }) {
      if (session.user) {
        (session.user as any).role = token.role;
        (session.user as any).emailVerified = token.emailVerified;
        (session.user as any).isBanned = token.isBanned;
      }
      
      return session;
    },
  },
  
  events: {
    async signIn({ user, account, isNewUser }) {
      // Log successful sign-in
      console.log(`User ${user.email} signed in with ${account?.provider}`);
    },
    
    async signOut({ session, token }) {
      // Log sign-out
      console.log(`User signed out: ${session?.user?.email || token?.email}`);
    },
  },
  
  debug: process.env.NODE_ENV === 'development',
};

/**
 * Server-side auth utilities
 */
export async function getServerUser(req: any): Promise<UserType | null> {
  try {
    // This would integrate with your session management
    // For now, return null
    return null;
  } catch (error) {
    return null;
  }
}

/**
 * Middleware helper for protecting routes
 */
export function requireAuth(handler: any) {
  return async (req: any, res: any) => {
    const user = await getServerUser(req);
    
    if (!user) {
      return res.status(401).json({ error: 'Authentication required' });
    }
    
    if (user.isBanned) {
      return res.status(403).json({ error: 'Account suspended' });
    }
    
    req.user = user;
    return handler(req, res);
  };
}

/**
 * Middleware helper for requiring specific role
 */
export function requireRole(role: string) {
  return (handler: any) => {
    return async (req: any, res: any) => {
      const user = await getServerUser(req);
      
      if (!user) {
        return res.status(401).json({ error: 'Authentication required' });
      }
      
      if (!hasRole(user, role)) {
        return res.status(403).json({ error: 'Insufficient permissions' });
      }
      
      req.user = user;
      return handler(req, res);
    };
  };
}

/**
 * Middleware helper for requiring specific permission
 */
export function requirePermission(action: string) {
  return (handler: any) => {
    return async (req: any, res: any) => {
      const user = await getServerUser(req);
      
      if (!user) {
        return res.status(401).json({ error: 'Authentication required' });
      }
      
      if (!hasPermission(user, action)) {
        return res.status(403).json({ error: 'Permission denied' });
      }
      
      req.user = user;
      return handler(req, res);
    };
  };
}