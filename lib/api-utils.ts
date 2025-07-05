import { NextRequest } from 'next/server';
import { StorageProvider } from '@/models/StorageProvider';

// Helper to get user IP address
export function getClientIP(request: NextRequest): string {
  const forwarded = request.headers.get('x-forwarded-for');
  const realIP = request.headers.get('x-real-ip');
  
  if (forwarded) {
    return forwarded.split(',')[0].trim();
  }
  
  if (realIP) {
    return realIP;
  }
  
  return '127.0.0.1';
}

// Helper to convert numbers to strings for headers
export function formatHeaders(headers: Record<string, number | string>): Record<string, string> {
  const result: Record<string, string> = {};
  for (const [key, value] of Object.entries(headers)) {
    result[key] = String(value);
  }
  return result;
}

// Storage provider helper
export async function getStorageProvider(providerType?: string) {
  const provider = await StorageProvider.findOne({
    $or: [
      { type: providerType, isActive: true },
      { isDefault: true, isActive: true }
    ]
  }).lean();

  return {
    provider: provider?.type || 'local',
    getSignedUrl: async (key: string, options: any) => {
      // For download/preview operations, return direct URL
      if (options.operation === 'download' || options.operation === 'preview') {
        return `https://storage.example.com/${key}?expires=${Date.now() + (options.expiresIn * 1000)}`;
      }
      
      // For upload operations, return upload configuration
      if (options.operation === 'upload') {
        return {
          uploadUrl: `https://storage.example.com/upload/${key}`,
          uploadId: `upload_${Date.now()}`,
          chunkSize: options.chunkSize || 5 * 1024 * 1024, // 5MB
          maxChunks: options.maxChunks || 1000
        };
      }
      
      // Default fallback
      return `https://storage.example.com/${key}`;
    }
  };
}