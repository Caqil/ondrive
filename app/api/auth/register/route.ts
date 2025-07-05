import type { NextApiRequest, NextApiResponse } from 'next';
import { connectDB } from '@/lib/db';
import { User, Settings } from '@/models';
import { registerSchema } from '@/lib/validations/auth';
import { hashPassword, generateToken } from '@/lib/crypto';
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
    // Rate limiting - use IP address as identifier
    const identifier = req.headers['x-forwarded-for'] as string || req.socket.remoteAddress || 'unknown';
    const rateLimitResult = rateLimitCheck(identifier);
    
    // Set rate limit headers
    Object.entries(rateLimitResult.headers).forEach(([key, value]) => {
      res.setHeader(key, value);
    });

    // Validate request body
    const validation = registerSchema.safeParse(req.body);
    if (!validation.success) {
      return res.status(400).json({
        error: 'Validation failed',
        code: API_RESPONSE_CODES.VALIDATION_ERROR,
        details: validation.error.format()
      });
    }

    const { email, password, name, acceptTerms } = validation.data;

    await connectDB();

    // Check if registration is enabled
    const settings = await Settings.findOne();
    if (settings && !settings.app.enableRegistration) {
      return res.status(403).json({
        error: 'Registration is currently disabled',
        code: API_RESPONSE_CODES.AUTHORIZATION_ERROR
      });
    }

    // Check if user already exists
    const existingUser = await User.findOne({ email });
    if (existingUser) {
      return res.status(409).json({
        error: 'User already exists with this email',
        code: API_RESPONSE_CODES.CONFLICT
      });
    }

    // Hash password
    const hashedPassword = await hashPassword(password);

    // Generate email verification token
    const emailVerificationToken = generateToken();

    // Create new user
    const user = new User({
      email,
      password: hashedPassword,
      name,
      emailVerificationToken,
      storageQuota: settings?.app.defaultUserQuota || 15 * 1024 * 1024 * 1024, // 15GB default
    });

    await user.save();

    // Send email verification if email is enabled
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
        // Don't fail registration if email fails
      }
    }

    // Return user without sensitive data
    const userResponse = {
      _id: user._id,
      email: user.email,
      name: user.name,
      role: user.role,
      emailVerified: user.emailVerified,
      createdAt: user.createdAt
    };

    res.status(201).json({
      success: true,
      message: settings?.email.enabled 
        ? 'Account created successfully. Please check your email for verification.'
        : 'Account created successfully.',
      data: userResponse,
      code: API_RESPONSE_CODES.SUCCESS
    });

  } catch (error: any) {
    console.error('Registration error:', error);
    
    // Handle rate limit errors
    if (error.status === 429) {
      return res.status(429).json({
        error: 'Too many registration attempts. Please try again later.',
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