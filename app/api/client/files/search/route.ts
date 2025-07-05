import { NextRequest } from 'next/server';
import { File } from '@/models/File';
import { connectDB } from '@/lib/db';
import { apiRateLimiter, createRateLimitMiddleware } from '@/lib/rate-limit';
import { getServerSession } from 'next-auth';
import { authOptions } from '@/lib/auth';
import { getClientIP, formatHeaders } from '@/lib/api-utils';
import { z } from 'zod';
import type { PaginatedResponse, File as FileType, User } from '@/types';

const searchFilesSchema = z.object({
  query: z.string().min(1, 'Search query is required'),
  page: z.coerce.number().min(1).default(1),
  limit: z.coerce.number().min(1).max(100).default(20),
  mimeType: z.string().optional(),
  folderId: z.string().optional(),
  tags: z.string().optional().transform(val => val ? val.split(',') : undefined),
  sort: z.enum(['relevance', 'name', 'size', 'created', 'modified']).default('relevance'),
  order: z.enum(['asc', 'desc']).default('desc')
});

// GET /api/client/files/search - Search files
export async function GET(request: NextRequest) {
  try {
    const session = await getServerSession(authOptions);
    if (!session?.user) {
      return Response.json({
        success: false,
        error: 'Authentication required'
      }, { status: 401 });
    }

    const user = session.user as User;
    const clientIP = getClientIP(request);
    const rateLimitResult = createRateLimitMiddleware(apiRateLimiter)(clientIP);
    
    await connectDB();
    const { searchParams } = new URL(request.url);
    
    const queryParams = Object.fromEntries(searchParams.entries());
    const { query, page, limit, mimeType, folderId, tags, sort, order } = searchFilesSchema.parse(queryParams);

    // Build search criteria
    const searchCriteria: any = {
      isTrashed: false,
      $or: [
        { owner: user._id },
        { visibility: 'public' },
        { team: user.currentTeam, visibility: 'team' }
      ],
      $text: { $search: query }
    };

    if (mimeType) searchCriteria.mimeType = new RegExp(mimeType, 'i');
    if (folderId) searchCriteria.folder = folderId;
    if (tags) searchCriteria.tags = { $in: tags };

    // Execute search with pagination
    const skip = (page - 1) * limit;
    
    let sortCriteria: any = {};
    if (sort === 'relevance') {
      sortCriteria = { score: { $meta: 'textScore' } };
    } else {
      const sortField = sort === 'created' ? 'createdAt' : 
                      sort === 'modified' ? 'updatedAt' : sort;
      sortCriteria[sortField] = order === 'asc' ? 1 : -1;
    }

    const [files, total] = await Promise.all([
      File.find(searchCriteria, sort === 'relevance' ? { score: { $meta: 'textScore' } } : {})
        .sort(sortCriteria)
        .skip(skip)
        .limit(limit)
        .populate('owner', 'name email avatar')
        .populate('folder', 'name path')
        .lean(),
      File.countDocuments(searchCriteria)
    ]);

    const response: PaginatedResponse<FileType> = {
      success: true,
      data: files as unknown as FileType[],
      pagination: {
        page,
        limit,
        total,
        totalPages: Math.ceil(total / limit),
        hasNext: page < Math.ceil(total / limit),
        hasPrev: page > 1
      }
    };

    return Response.json(response, {
      headers: formatHeaders(rateLimitResult.headers)
    });
  } catch (error: any) {
    console.error('File search error:', error);
    
    if (error.name === 'ZodError') {
      return Response.json({
        success: false,
        error: 'Invalid search parameters',
        details: error.errors
      }, { status: 400 });
    }

    return Response.json({
      success: false,
      error: 'Search failed'
    }, { status: 500 });
  }
}
