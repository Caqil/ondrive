import type { NextApiRequest, NextApiResponse } from 'next';
import { connectDB } from '@/lib/db';
import { User } from '@/models';
import { twoFactorSetupSchema } from '@/lib/validations/auth';
import { verifyTOTP, generateSecretKey } from '@/lib/crypto';
import { API_RESPONSE_CODES } from '@/lib/constants';
import jwt from 'jsonwebtoken';

export default async function handler(req: NextApiRequest, res: NextApiResponse) {
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

  if (req.method === 'GET') {
    // Generate 2FA setup data
    try {
      const userId = decoded.userId;
      
      await connectDB();
      
      const user = await User.findById(userId);
      if (!user) {
        return res.status(404).json({
          error: 'User not found',
          code: API_RESPONSE_CODES.NOT_FOUND
        });
      }

      if (user.twoFactorEnabled) {
        return res.status(400).json({
          error: '2FA is already enabled',
          code: API_RESPONSE_CODES.VALIDATION_ERROR
        });
      }

      // Generate secret and QR code data
      const secret = generateSecretKey(20);
      const appName = process.env.NEXT_PUBLIC_APP_NAME || 'Drive Clone';
      const qrCodeUrl = `otpauth://totp/${encodeURIComponent(`${appName}:${user.email}`)}?secret=${secret}&issuer=${encodeURIComponent(appName)}`;

      res.status(200).json({
        success: true,
        data: {
          secret,
          qrCodeUrl,
          backupCodes: [], // Generate backup codes in production
        },
        code: API_RESPONSE_CODES.SUCCESS
      });

    } catch (error) {
      console.error('2FA setup generation error:', error);
      res.status(500).json({
        error: 'Internal server error',
        code: API_RESPONSE_CODES.SERVER_ERROR
      });
    }
  } else if (req.method === 'POST') {
    // Complete 2FA setup
    try {
      const validation = twoFactorSetupSchema.safeParse(req.body);
      if (!validation.success) {
        return res.status(400).json({
          error: 'Validation failed',
          code: API_RESPONSE_CODES.VALIDATION_ERROR,
          details: validation.error.format()
        });
      }

      const { secret, code } = validation.data;
      const userId = decoded.userId;

      // Verify the TOTP code
      const isValidCode = verifyTOTP(secret, code);
      if (!isValidCode) {
        return res.status(400).json({
          error: 'Invalid 2FA code',
          code: API_RESPONSE_CODES.VALIDATION_ERROR
        });
      }

      await connectDB();

      // Enable 2FA for user
      await User.findByIdAndUpdate(userId, {
        twoFactorEnabled: true,
        twoFactorSecret: secret
      });

      res.status(200).json({
        success: true,
        message: '2FA enabled successfully',
        code: API_RESPONSE_CODES.SUCCESS
      });

    } catch (error) {
      console.error('2FA setup error:', error);
      res.status(500).json({
        error: 'Internal server error',
        code: API_RESPONSE_CODES.SERVER_ERROR
      });
    }
  } else {
    res.status(405).json({ 
      error: 'Method not allowed',
      code: API_RESPONSE_CODES.VALIDATION_ERROR 
    });
  }
}