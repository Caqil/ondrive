import type { NextApiRequest, NextApiResponse } from 'next';
import { connectDB } from '@/lib/db';
import { User } from '@/models';
import { loginSchema } from '@/lib/validations/auth';
import { verifyPassword, generateToken, verifyTOTP } from '@/lib/crypto';
import { API_RESPONSE_CODES } from '@/lib/constants';
import { authRateLimiter, createRateLimitMiddleware } from '@/lib/rate-limit';
import jwt from 'jsonwebtoken';

const rateLimitCheck = createRateLimitMiddleware(authRateLimiter);

export default async function handler(req: NextApiRequest, res: NextApiResponse) {
  if (req.method !== 'POST') {
    return res.status(405).json({ 
      error: 'Method not allowed',
      code: API_RESPONSE_CODES.VALIDATION_ERROR 
    });
  }

  try {
    // Rate limiting
    const identifier = req.headers['x-forwarded-for'] as string || req.socket.remoteAddress || 'unknown';
    const rateLimitResult = rateLimitCheck(identifier);
    
    // Set rate limit headers
    Object.entries(rateLimitResult.headers).forEach(([key, value]) => {
      res.setHeader(key, value);
    });

    // Validate request body
    const validation = loginSchema.safeParse(req.body);
    if (!validation.success) {
      return res.status(400).json({
        error: 'Validation failed',
        code: API_RESPONSE_CODES.VALIDATION_ERROR,
        details: validation.error.format()
      });
    }

    const { email, password, rememberMe } = validation.data;

    await connectDB();

    // Find user with password and 2FA secret
    const user = await User.findOne({ email }).select('+password +twoFactorSecret');
    if (!user) {
      return res.status(401).json({
        error: 'Invalid email or password',
        code: API_RESPONSE_CODES.AUTHENTICATION_ERROR
      });
    }

    // Check if account is banned
    if (user.isBanned) {
      return res.status(403).json({
        error: 'Account has been suspended',
        code: API_RESPONSE_CODES.AUTHORIZATION_ERROR
      });
    }

    // Verify password
    const isValidPassword = await verifyPassword(password, user.password);
    if (!isValidPassword) {
      return res.status(401).json({
        error: 'Invalid email or password',
        code: API_RESPONSE_CODES.AUTHENTICATION_ERROR
      });
    }

    // Check if 2FA is required
    if (user.twoFactorEnabled) {
      // Generate temporary session token for 2FA verification
      const sessionToken = generateToken();
      const tempSession = {
        userId: user._id,
        email: user.email,
        requires2FA: true,
        expiresAt: new Date(Date.now() + 10 * 60 * 1000) // 10 minutes
      };

      // Use JWT for temporary session
      const tempJWT = jwt.sign(tempSession, process.env.NEXTAUTH_SECRET!, { expiresIn: '10m' });

      return res.status(200).json({
        success: true,
        requires2FA: true,
        sessionToken: tempJWT,
        message: 'Please provide your 2FA code',
        code: API_RESPONSE_CODES.SUCCESS
      });
    }

    // Update last login
    await User.findByIdAndUpdate(user._id, { lastLogin: new Date() });

    // Generate JWT token
    const tokenPayload = {
      userId: user._id,
      email: user.email,
      role: user.role,
      emailVerified: user.emailVerified
    };

    const expiresIn = rememberMe ? '30d' : '1d';
    const accessToken = jwt.sign(tokenPayload, process.env.NEXTAUTH_SECRET!, { expiresIn });

    // Create session object
    const session = {
      accessToken,
      expiresAt: new Date(Date.now() + (rememberMe ? 30 * 24 * 60 * 60 * 1000 : 24 * 60 * 60 * 1000))
    };

    // Return user without sensitive data
    const userResponse = {
      _id: user._id,
      email: user.email,
      name: user.name,
      avatar: user.avatar,
      role: user.role,
      emailVerified: user.emailVerified,
      twoFactorEnabled: user.twoFactorEnabled,
      storageUsed: user.storageUsed,
      storageQuota: user.storageQuota,
      preferences: user.preferences,
      lastLogin: new Date()
    };

    res.status(200).json({
      success: true,
      message: 'Login successful',
      user: userResponse,
      session,
      code: API_RESPONSE_CODES.SUCCESS
    });

  } catch (error: any) {
    console.error('Login error:', error);
    
    // Handle rate limit errors
    if (error.status === 429) {
      return res.status(429).json({
        error: 'Too many login attempts. Please try again later.',
        code: API_RESPONSE_CODES.RATE_LIMITED,
        headers: error.headers
      });
    }

    res.status(500).json({
      error: 'Internal server error',
      code: API_RESPONSE_CODES.SERVER_ERROR
    });
  }
}