import type { NextApiRequest, NextApiResponse } from 'next';
import { connectDB } from '@/lib/db';
import { User, Settings } from '@/models';
import { resendEmailVerificationSchema } from '@/lib/validations/auth';
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
    const validation = resendEmailVerificationSchema.safeParse(req.body);
    if (!validation.success) {
      return res.status(400).json({
        error: 'Validation failed',
        code: API_RESPONSE_CODES.VALIDATION_ERROR,
        details: validation.error.format()
      });
    }

    const { email } = validation.data;

    await connectDB();

    // Find user
    const user = await User.findOne({ email });
    if (!user) {
      return res.status(404).json({
        error: 'User not found',
        code: API_RESPONSE_CODES.NOT_FOUND
      });
    }

    // Check if already verified
    if (user.emailVerified) {
      return res.status(400).json({
        error: 'Email is already verified',
        code: API_RESPONSE_CODES.VALIDATION_ERROR
      });
    }

    // Generate new verification token
    const emailVerificationToken = generateToken();

    // Update user with new token
    await User.findByIdAndUpdate(user._id, {
      emailVerificationToken
    });

    // Get settings for email configuration
    const settings = await Settings.findOne();
    
    // Send verification email if enabled
    if (settings?.email.enabled && settings.email.templates.emailVerification) {
      try {
        await sendTemplateEmail(
          settings.email,
          'emailVerification',
          user.email,
          {
            name: user.name,
            verificationUrl: `${process.env.NEXTAUTH_URL}/auth/verify-email?token=${emailVerificationToken}`,
            appName: settings.app.name || 'Drive Clone'
          }
        );
      } catch (emailError) {
        console.error('Failed to send verification email:', emailError);
        return res.status(500).json({
          error: 'Failed to send verification email',
          code: API_RESPONSE_CODES.SERVER_ERROR
        });
      }
    }

    res.status(200).json({
      success: true,
      message: 'Verification email sent successfully',
      code: API_RESPONSE_CODES.SUCCESS
    });

  } catch (error: any) {
    console.error('Resend verification error:', error);
    
    if (error.status === 429) {
      return res.status(429).json({
        error: 'Too many verification requests. Please try again later.',
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
