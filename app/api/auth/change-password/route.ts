import type { NextApiRequest, NextApiResponse } from 'next';
import { connectDB } from '@/lib/db';
import { User, Settings } from '@/models';
import { changePasswordSchema } from '@/lib/validations/auth';
import { hashPassword, verifyPassword } from '@/lib/crypto';
import { sendTemplateEmail } from '@/lib/email/utils';
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

    // Validate request body
    const validation = changePasswordSchema.safeParse(req.body);
    if (!validation.success) {
      return res.status(400).json({
        error: 'Validation failed',
        code: API_RESPONSE_CODES.VALIDATION_ERROR,
        details: validation.error.format()
      });
    }

    const { currentPassword, newPassword } = validation.data;
    const userId = decoded.userId;

    await connectDB();

    // Get user with password
    const user = await User.findById(userId).select('+password');
    if (!user) {
      return res.status(404).json({
        error: 'User not found',
        code: API_RESPONSE_CODES.NOT_FOUND
      });
    }

    // Verify current password
    const isValidPassword = await verifyPassword(currentPassword, user.password);
    if (!isValidPassword) {
      return res.status(400).json({
        error: 'Current password is incorrect',
        code: API_RESPONSE_CODES.VALIDATION_ERROR
      });
    }

    // Hash new password
    const hashedPassword = await hashPassword(newPassword);

    // Update password
    await User.findByIdAndUpdate(userId, {
      password: hashedPassword
    });

    // Get settings for email configuration
    const settings = await Settings.findOne();
    
    // Send password changed confirmation email
    if (settings?.email.enabled && settings.email.templates.passwordChanged) {
      try {
        await sendTemplateEmail(
          settings.email,
          'passwordChanged',
          user.email,
          {
            name: user.name,
            appName: settings.app.name || 'Drive Clone',
            changeTime: new Date().toLocaleString()
          }
        );
      } catch (emailError) {
        console.error('Failed to send password changed email:', emailError);
        // Don't fail the operation if email fails
      }
    }

    res.status(200).json({
      success: true,
      message: 'Password changed successfully',
      code: API_RESPONSE_CODES.SUCCESS
    });

  } catch (error) {
    console.error('Change password error:', error);
    res.status(500).json({
      error: 'Internal server error',
      code: API_RESPONSE_CODES.SERVER_ERROR
    });
  }
}