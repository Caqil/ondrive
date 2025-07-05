// lib/validations/base.ts
import { z } from 'zod';

// Base ObjectId validation - shared across all validation files
export const objectIdSchema = z.string().regex(/^[0-9a-fA-F]{24}$/, 'Invalid ObjectId format');

// Common currency validation
export const currencySchema = z.enum(['USD', 'EUR', 'GBP', 'CAD', 'AUD', 'JPY', 'CHF', 'CNY', 'INR', 'BRL', 'MXN', 'SGD', 'HKD', 'NOK', 'SEK', 'DKK', 'PLN', 'CZK', 'HUF', 'RON', 'BGN', 'HRK', 'RSD', 'BAM', 'MKD', 'ALL', 'TRY', 'RUB', 'UAH', 'KZT', 'UZS', 'AZN', 'GEL', 'AMD', 'BYN', 'MDL', 'KGS', 'TJS', 'TMT', 'MNT', 'LAK', 'KHR', 'MMK', 'BDT', 'BTN', 'NPR', 'LKR', 'MVR', 'PKR', 'AFN', 'IRR', 'IQD', 'SYP', 'LBP', 'JOD', 'SAR', 'QAR', 'AED', 'OMR', 'YER', 'KWD', 'BHD', 'ILS', 'EGP', 'LYD', 'TND', 'DZD', 'MAD', 'AUD', 'ETB', 'KES', 'UGX', 'TZS', 'RWF', 'BIF', 'DJF', 'SOS', 'MGA', 'KMF', 'SCR', 'MUR', 'BWP', 'SZL', 'LSL', 'ZAR', 'NAD', 'AOA', 'XAF', 'XOF', 'CDF', 'STN', 'GHS', 'GMD', 'GNF', 'LRD', 'SLE', 'CVE', 'MRU', 'XPF']);

// Base email validation
export const emailSchema = z.string()
  .email('Invalid email address')
  .min(1, 'Email is required')
  .max(254, 'Email cannot exceed 254 characters')
  .toLowerCase();

// Base name validation
export const nameSchema = z.string()
  .min(1, 'Name is required')
  .max(100, 'Name cannot exceed 100 characters')
  .transform((name) => name.trim());

// Base phone validation
export const phoneSchema = z.string()
  .regex(/^\+?[1-9]\d{1,14}$/, 'Invalid phone number format')
  .optional();

// Base URL validation
export const urlSchema = z.string()
  .url('Invalid URL format')
  .optional();

// Pagination base schemas
export const paginationSchema = z.object({
  page: z.number().min(1).default(1),
  limit: z.number().min(1).max(100).default(20),
  sort: z.string().optional(),
  order: z.enum(['asc', 'desc']).default('desc')
});

// Date range validation
export const dateRangeSchema = z.object({
  start: z.string().datetime().optional(),
  end: z.string().datetime().optional()
}).refine((data) => {
  if (data.start && data.end) {
    return new Date(data.start) <= new Date(data.end);
  }
  return true;
}, {
  message: 'Start date must be before end date',
  path: ['end']
});

// Base address schema
export const addressSchema = z.object({
  line1: z.string().min(1, 'Address line 1 is required').max(100),
  line2: z.string().max(100).optional(),
  city: z.string().min(1, 'City is required').max(50),
  state: z.string().max(50).optional(),
  postalCode: z.string().min(1, 'Postal code is required').max(20),
  country: z.string().length(2, 'Country code must be exactly 2 characters').toUpperCase()
});

// Base metadata schema
export const metadataSchema = z.record(z.any()).optional();

// Status enums commonly used
export const statusSchema = z.enum(['active', 'inactive', 'pending', 'cancelled', 'expired']);

// Base file size validation (in bytes)
export const fileSizeSchema = z.number()
  .min(1, 'File size must be greater than 0')
  .max(750 * 1024 * 1024 * 1024, 'File size cannot exceed 750GB'); // 750GB max