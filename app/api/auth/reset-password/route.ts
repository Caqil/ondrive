import type { NextApiRequest, NextApiResponse } from 'next';
import { connectDB } from '@/lib/db';
import { User, Settings } from '@/models';
import { resetPasswordSchema } from '@/lib/validations/auth';
import { hashPassword } from '@/lib/crypto';
import { sendTemplateEmail } from '@/lib/email/utils';
import { API_RESPONSE_CODES } from '@/lib/constants';
import { authRateLimiter, createRateLimitMiddleware } from '@/lib/rate-limit';

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
    
    Object.entries(rateLimitResult.headers).forEach(([key, value]) => {
      res.setHeader(key, value);
    });

    // Validate request body
    const validation = resetPasswordSchema.safeParse(req.body);
    if (!validation.success) {
      return res.status(400).json({
        error: 'Validation failed',
        code: API_RESPONSE_CODES.VALIDATION_ERROR,
        details: validation.error.format()
      });
    }

    const { token, password } = validation.data;

    await connectDB();

    // Find user with valid reset token
    const user = await User.findOne({
      passwordResetToken: token,
      passwordResetExpires: { $gt: new Date() }
    });

    if (!user) {
      return res.status(400).json({
        error: 'Invalid or expired reset token',
        code: API_RESPONSE_CODES.VALIDATION_ERROR
      });
    }

    // Hash new password
    const hashedPassword = await hashPassword(password);

    // Update user password and clear reset token
    await User.findByIdAndUpdate(user._id, {
      password: hashedPassword,
      passwordResetToken: undefined,
      passwordResetExpires: undefined
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
      message: 'Password has been reset successfully',
      code: API_RESPONSE_CODES.SUCCESS
    });

  } catch (error: any) {
    console.error('Reset password error:', error);
    
    if (error.status === 429) {
      return res.status(429).json({
        error: 'Too many password reset attempts. Please try again later.',
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
