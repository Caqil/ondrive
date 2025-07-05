import type { NextApiRequest, NextApiResponse } from 'next';
import { connectDB } from '@/lib/db';
import { User } from '@/models';
import { API_RESPONSE_CODES } from '@/lib/constants';
import jwt from 'jsonwebtoken';

export default async function handler(req: NextApiRequest, res: NextApiResponse) {
  if (req.method !== 'GET') {
    return res.status(405).json({ 
      error: 'Method not allowed',
      code: API_RESPONSE_CODES.VALIDATION_ERROR 
    });
  }

  try {
    // Get token from Authorization header
    const authHeader = req.headers.authorization;
    if (!authHeader || !authHeader.startsWith('Bearer ')) {
      return res.status(401).json({
        error: 'Authentication required',
        code: API_RESPONSE_CODES.AUTHENTICATION_ERROR
      });
    }

    const token = authHeader.substring(7);

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

    const userId = decoded.userId;

    await connectDB();

    // Get current user data
    const user = await User.findById(userId);
    if (!user) {
      return res.status(404).json({
        error: 'User not found',
        code: API_RESPONSE_CODES.NOT_FOUND
      });
    }

    // Check if user is banned or inactive
    if (user.isBanned || !user.isActive) {
      return res.status(403).json({
        error: 'Account suspended or deactivated',
        code: API_RESPONSE_CODES.AUTHORIZATION_ERROR
      });
    }

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
      currentTeam: user.currentTeam,
      subscriptionStatus: user.subscriptionStatus,
      lastLogin: user.lastLogin,
      createdAt: user.createdAt
    };

    res.status(200).json({
      success: true,
      user: userResponse,
      code: API_RESPONSE_CODES.SUCCESS
    });

  } catch (error) {
    console.error('Session error:', error);
    res.status(500).json({
      error: 'Internal server error',
      code: API_RESPONSE_CODES.SERVER_ERROR
    });
  }
}