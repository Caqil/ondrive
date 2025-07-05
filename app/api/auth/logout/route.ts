import { NextRequest, NextResponse } from 'next/server';
import { cache } from '@/lib/cache';
import jwt from 'jsonwebtoken';

export async function POST(request: NextRequest) {
  try {
    const authHeader = request.headers.get('authorization');
    const token = authHeader?.replace('Bearer ', '');

    if (token) {
      try {
        const decoded = jwt.verify(token, process.env.JWT_SECRET!) as any;
        
        // Remove session from cache
        await cache.del(`session:${decoded.userId}`);
        
        // Add token to blacklist
        await cache.set(
          `blacklist:${token}`, 
          true, 
          24 * 60 * 60 // 24 hours
        );
      } catch (jwtError) {
        // Token invalid, but still proceed with logout
      }
    }

    return NextResponse.json({
      success: true,
      message: 'Logged out successfully'
    });

  } catch (error) {
    console.error('Logout error:', error);
    return NextResponse.json(
      { error: 'Internal server error' },
      { status: 500 }
    );
  }
}
