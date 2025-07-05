import type { NextApiRequest, NextApiResponse } from 'next';
import { connectDB } from '@/lib/db';
import { User } from '@/models';
import { twoFactorVerifySchema } from '@/lib/validations/auth';
import { verifyTOTP } from '@/lib/crypto';
import { API_RESPONSE_CODES } from '@/lib/constants';
import jwt from 'jsonwebtoken';

export default async function handler(req: NextApiRequest, res: NextApiResponse) {
  if (req.method !== 'POST') {
    return res.status(405).json({ 
      error: 'Method not allowed',
      code: API_RESPONSE_CODES.VALIDATION_ERROR 
    });
  }

  try {
    // Validate request body
    const validation = twoFactorVerifySchema.safeParse(req.body);
    if (!validation.success) {
      return res.status(400).json({
        error: 'Validation failed',
        code: API_RESPONSE_CODES.VALIDATION_ERROR,
        details: validation.error.format()
      });
    }

    const { code, sessionToken } = validation.data;

    // Verify session token
    let tempSession;
    try {
      tempSession = jwt.verify(sessionToken!, process.env.NEXTAUTH_SECRET!) as any;
    } catch (error) {
      return res.status(401).json({
        error: 'Invalid or expired session token',
        code: API_RESPONSE_CODES.AUTHENTICATION_ERROR
      });
    }

    await connectDB();

    // Get user with 2FA secret
    const user = await User.findById(tempSession.userId).select('+twoFactorSecret');
    if (!user || !user.twoFactorEnabled) {
      return res.status(400).json({
        error: 'Invalid 2FA configuration',
        code: API_RESPONSE_CODES.VALIDATION_ERROR
      });
    }

    // Verify TOTP code
    const isValidCode = verifyTOTP(user.twoFactorSecret, code);
    if (!isValidCode) {
      return res.status(400).json({
        error: 'Invalid 2FA code',
        code: API_RESPONSE_CODES.VALIDATION_ERROR
      });
    }

    // Update last login
    await User.findByIdAndUpdate(user._id, { lastLogin: new Date() });

    // Generate full session JWT
    const tokenPayload = {
      userId: user._id,
      email: user.email,
      role: user.role,
      emailVerified: user.emailVerified
    };

    const accessToken = jwt.sign(tokenPayload, process.env.NEXTAUTH_SECRET!, { expiresIn: '1d' });

    // Create session object
    const session = {
      accessToken,
      expiresAt: new Date(Date.now() + 24 * 60 * 60 * 1000) // 24 hours
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
      message: '2FA verification successful',
      user: userResponse,
      session,
      code: API_RESPONSE_CODES.SUCCESS
    });

  } catch (error) {
    console.error('2FA verification error:', error);
    res.status(500).json({
      error: 'Internal server error',
      code: API_RESPONSE_CODES.SERVER_ERROR
    });
  }
}
