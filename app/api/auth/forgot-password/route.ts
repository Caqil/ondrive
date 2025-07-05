import type { NextApiRequest, NextApiResponse } from 'next';
import { connectDB } from '@/lib/db';
import { User, Settings } from '@/models';
import { forgotPasswordSchema } from '@/lib/validations/auth';
import { generateToken } from '@/lib/crypto';
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
    const validation = forgotPasswordSchema.safeParse(req.body);
    if (!validation.success) {
      return res.status(400).json({
        error: 'Validation failed',
        code: API_RESPONSE_CODES.VALIDATION_ERROR,
        details: validation.error.format()
      });
    }

    const { email } = validation.data;

    await connectDB();

    // Check if user exists
    const user = await User.findOne({ email });
    if (!user) {
      // Don't reveal if user exists or not for security
      return res.status(200).json({
        success: true,
        message: 'If an account with that email exists, a password reset link has been sent.',
        code: API_RESPONSE_CODES.SUCCESS
      });
    }

    // Generate password reset token
    const resetToken = generateToken();
    const resetExpires = new Date(Date.now() + 60 * 60 * 1000); // 1 hour

    // Update user with reset token
    await User.findByIdAndUpdate(user._id, {
      passwordResetToken: resetToken,
      passwordResetExpires: resetExpires
    });

    // Get settings for email configuration
    const settings = await Settings.findOne();
    
    // Send password reset email if email is enabled
    if (settings?.email.enabled && settings.email.templates.passwordReset) {
      try {
        await sendTemplateEmail(
          settings.email,
          'passwordReset',
          user.email,
          {
            name: user.name,
            resetUrl: `${process.env.NEXTAUTH_URL}/auth/reset-password?token=${resetToken}`,
            appName: settings.app.name || 'Drive Clone',
            expiresIn: '1 hour'
          }
        );
      } catch (emailError) {
        console.error('Failed to send password reset email:', emailError);
        return res.status(500).json({
          error: 'Failed to send password reset email',
          code: API_RESPONSE_CODES.SERVER_ERROR
        });
      }
    }

    res.status(200).json({
      success: true,
      message: 'If an account with that email exists, a password reset link has been sent.',
      code: API_RESPONSE_CODES.SUCCESS
    });

  } catch (error: any) {
    console.error('Forgot password error:', error);
    
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