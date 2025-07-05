import type { NextApiRequest, NextApiResponse } from 'next';
import { connectDB } from '@/lib/db';
import { User, Settings } from '@/models';
import { emailVerificationSchema } from '@/lib/validations/auth';
import { sendTemplateEmail } from '@/lib/email/utils';
import { API_RESPONSE_CODES } from '@/lib/constants';

export default async function handler(req: NextApiRequest, res: NextApiResponse) {
  if (req.method !== 'POST') {
    return res.status(405).json({ 
      error: 'Method not allowed',
      code: API_RESPONSE_CODES.VALIDATION_ERROR 
    });
  }

  try {
    // Validate request body
    const validation = emailVerificationSchema.safeParse(req.body);
    if (!validation.success) {
      return res.status(400).json({
        error: 'Validation failed',
        code: API_RESPONSE_CODES.VALIDATION_ERROR,
        details: validation.error.format()
      });
    }

    const { token } = validation.data;

    await connectDB();

    // Find user with verification token
    const user = await User.findOne({ emailVerificationToken: token });
    if (!user) {
      return res.status(400).json({
        error: 'Invalid verification token',
        code: API_RESPONSE_CODES.VALIDATION_ERROR
      });
    }

    // Update user email verification status
    await User.findByIdAndUpdate(user._id, {
      emailVerified: true,
      emailVerificationToken: undefined
    });

    // Get settings for email configuration
    const settings = await Settings.findOne();
    
    // Send welcome email if enabled
    if (settings?.email.enabled && settings.email.templates.welcome) {
      try {
        await sendTemplateEmail(
          settings.email,
          'welcome',
          user.email,
          {
            name: user.name,
            appName: settings.app.name || 'Drive Clone',
            dashboardUrl: `${process.env.NEXTAUTH_URL}/dashboard`
          }
        );
      } catch (emailError) {
        console.error('Failed to send welcome email:', emailError);
        // Don't fail verification if email fails
      }
    }

    res.status(200).json({
      success: true,
      message: 'Email verified successfully',
      code: API_RESPONSE_CODES.SUCCESS
    });

  } catch (error) {
    console.error('Email verification error:', error);
    res.status(500).json({
      error: 'Internal server error',
      code: API_RESPONSE_CODES.SERVER_ERROR
    });
  }
}
