import type { NextApiRequest, NextApiResponse } from 'next';
import { connectDB } from '@/lib/db';
import { User } from '@/models';
import { twoFactorDisableSchema } from '@/lib/validations/auth';
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
    // Authentication middleware
    const authHeader = req.headers.authorization;
    if (!authHeader || !authHeader.startsWith('Bearer ')) {
      return res.status(401).json({
        error: 'Authentication required',
        code: API_RESPONSE_CODES.AUTHENTICATION_ERROR
      });
    }

    const token = authHeader.substring(7);
    let decoded;
    try {
      decoded = jwt.verify(token, process.env.NEXTAUTH_SECRET!) as any;
    } catch (jwtError) {
      return res.status(401).json({
        error: 'Invalid or expired token',
        code: API_RESPONSE_CODES.AUTHENTICATION_ERROR
      });
    }

    // Validate request body
    const validation = twoFactorDisableSchema.safeParse(req.body);
    if (!validation.success) {
      return res.status(400).json({
        error: 'Validation failed',
        code: API_RESPONSE_CODES.VALIDATION_ERROR,
        details: validation.error.format()
      });
    }

    const { code } = validation.data;
    const userId = decoded.userId;

    await connectDB();

    // Get user with 2FA secret
    const user = await User.findById(userId).select('+twoFactorSecret');
    if (!user) {
      return res.status(404).json({
        error: 'User not found',
        code: API_RESPONSE_CODES.NOT_FOUND
      });
    }

    if (!user.twoFactorEnabled) {
      return res.status(400).json({
        error: '2FA is not enabled',
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

    // Disable 2FA
    await User.findByIdAndUpdate(userId, {
      twoFactorEnabled: false,
      twoFactorSecret: undefined
    });

    res.status(200).json({
      success: true,
      message: '2FA disabled successfully',
      code: API_RESPONSE_CODES.SUCCESS
    });

  } catch (error) {
    console.error('2FA disable error:', error);
    res.status(500).json({
      error: 'Internal server error',
      code: API_RESPONSE_CODES.SERVER_ERROR
    });
  }
}