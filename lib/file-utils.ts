import { FILE_CATEGORIES, PROCESSING_STATUS, DEFAULTS } from '@/lib/constants';
import { formatBytes } from '@/lib/format-utils';
import { getMimeType, getCategoryFromMimeType } from '@/lib/mime-types';
import type { File, FileMetadata } from '@/types';

/**
 * Get file extension from filename
 */
export function getFileExtension(filename: string): string {
  const lastDotIndex = filename.lastIndexOf('.');
  if (lastDotIndex === -1 || lastDotIndex === 0) return '';
  return filename.slice(lastDotIndex + 1).toLowerCase();
}

/**
 * Get filename without extension
 */
export function getFileNameWithoutExtension(filename: string): string {
  const lastDotIndex = filename.lastIndexOf('.');
  if (lastDotIndex === -1 || lastDotIndex === 0) return filename;
  return filename.slice(0, lastDotIndex);
}

/**
 * Get file category based on MIME type
 */
export function getFileCategory(mimeType: string): string {
  return getCategoryFromMimeType(mimeType);
}

/**
 * Check if file is an image
 */
export function isImage(mimeType: string): boolean {
  return mimeType.startsWith('image/');
}

/**
 * Check if file is a video
 */
export function isVideo(mimeType: string): boolean {
  return mimeType.startsWith('video/');
}

/**
 * Check if file is audio
 */
export function isAudio(mimeType: string): boolean {
  return mimeType.startsWith('audio/');
}

/**
 * Check if file is a document
 */
export function isDocument(mimeType: string): boolean {
  return mimeType.startsWith('text/') || 
         mimeType === 'application/pdf' ||
         mimeType.includes('document') ||
         mimeType.includes('word') ||
         mimeType.includes('excel') ||
         mimeType.includes('powerpoint') ||
         mimeType === 'application/rtf';
}

/**
 * Check if file is an archive
 */
export function isArchive(mimeType: string): boolean {
  return mimeType.includes('zip') || 
         mimeType.includes('archive') || 
         mimeType.includes('compressed') ||
         mimeType.includes('tar') ||
         mimeType.includes('rar');
}

/**
 * Check if file is code
 */
export function isCode(mimeType: string): boolean {
  return mimeType.includes('javascript') || 
         mimeType.includes('typescript') ||
         mimeType.includes('json') || 
         mimeType.includes('xml') ||
         mimeType === 'text/css' ||
         mimeType === 'text/html';
}

/**
 * Check if file has thumbnail support
 */
export function hasThumbnailSupport(mimeType: string): boolean {
  return isImage(mimeType) || isVideo(mimeType) || mimeType === 'application/pdf';
}

/**
 * Check if file has preview support
 */
export function hasPreviewSupport(mimeType: string): boolean {
  return (
    isImage(mimeType) ||
    isVideo(mimeType) ||
    isAudio(mimeType) ||
    mimeType === 'application/pdf' ||
    mimeType.startsWith('text/') ||
    isCode(mimeType)
  );
}

/**
 * Check if file supports OCR
 */
export function supportsOCR(mimeType: string): boolean {
  return isImage(mimeType) || mimeType === 'application/pdf';
}

/**
 * Generate safe filename
 */
export function generateSafeFilename(filename: string): string {
  // Remove or replace unsafe characters
  let safeName = filename
    .replace(/[<>:"/\\|?*]/g, '_') // Replace unsafe characters
    .replace(/\s+/g, '_') // Replace spaces with underscores
    .replace(/_{2,}/g, '_') // Replace multiple underscores with single
    .replace(/^_+|_+$/g, ''); // Remove leading/trailing underscores

  // Ensure filename is not empty
  if (!safeName) {
    safeName = 'untitled';
  }

  // Truncate if too long
  if (safeName.length > DEFAULTS.MAX_FILENAME_LENGTH) {
    const extension = getFileExtension(safeName);
    const nameWithoutExt = getFileNameWithoutExtension(safeName);
    const maxNameLength = DEFAULTS.MAX_FILENAME_LENGTH - extension.length - 1; // -1 for the dot
    safeName = nameWithoutExt.slice(0, maxNameLength) + '.' + extension;
  }

  return safeName;
}

/**
 * Generate unique filename to avoid conflicts
 */
export function generateUniqueFilename(filename: string, existingFiles: string[]): string {
  let safeName = generateSafeFilename(filename);
  
  if (!existingFiles.includes(safeName)) {
    return safeName;
  }

  const extension = getFileExtension(safeName);
  const nameWithoutExt = getFileNameWithoutExtension(safeName);
  
  let counter = 1;
  let uniqueName: string;
  
  do {
    uniqueName = extension 
      ? `${nameWithoutExt} (${counter}).${extension}`
      : `${nameWithoutExt} (${counter})`;
    counter++;
  } while (existingFiles.includes(uniqueName) && counter < 1000);
  
  return uniqueName;
}

/**
 * Validate file type against allowed types
 */
export function isFileTypeAllowed(mimeType: string, allowedTypes: string[]): boolean {
  if (allowedTypes.length === 0) return true; // No restrictions
  return allowedTypes.includes(mimeType);
}

/**
 * Validate file size against limit
 */
export function isFileSizeValid(size: number, maxSize: number): boolean {
  return size <= maxSize;
}

/**
 * Get file icon name based on MIME type or extension
 */
export function getFileIcon(mimeType: string, filename?: string): string {
  // Image files
  if (isImage(mimeType)) return 'image';
  
  // Video files
  if (isVideo(mimeType)) return 'video';
  
  // Audio files
  if (isAudio(mimeType)) return 'audio';
  
  // Archive files
  if (isArchive(mimeType)) return 'archive';
  
  // Specific document types
  switch (mimeType) {
    case 'application/pdf':
      return 'file-pdf';
    case 'application/msword':
    case 'application/vnd.openxmlformats-officedocument.wordprocessingml.document':
      return 'file-word';
    case 'application/vnd.ms-excel':
    case 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet':
      return 'file-excel';
    case 'application/vnd.ms-powerpoint':
    case 'application/vnd.openxmlformats-officedocument.presentationml.presentation':
      return 'file-powerpoint';
    case 'text/csv':
      return 'file-csv';
    case 'application/json':
      return 'file-json';
    case 'application/xml':
    case 'text/xml':
      return 'file-xml';
    case 'text/plain':
      return 'file-text';
    case 'text/html':
      return 'file-html';
    case 'text/css':
      return 'file-css';
    case 'text/javascript':
      return 'file-js';
    default:
      break;
  }
  
  // Check by file extension if filename is provided
  if (filename) {
    const extension = getFileExtension(filename);
    switch (extension) {
      case 'ts':
      case 'tsx':
        return 'file-typescript';
      case 'jsx':
        return 'file-react';
      case 'vue':
        return 'file-vue';
      case 'py':
        return 'file-python';
      case 'java':
        return 'file-java';
      case 'php':
        return 'file-php';
      case 'rb':
        return 'file-ruby';
      case 'go':
        return 'file-go';
      case 'rs':
        return 'file-rust';
      case 'cpp':
      case 'c':
        return 'file-c';
      case 'cs':
        return 'file-csharp';
      case 'sql':
        return 'file-sql';
      case 'md':
        return 'file-markdown';
      case 'yaml':
      case 'yml':
        return 'file-yaml';
      default:
        break;
    }
  }
  
  // Code files
  if (isCode(mimeType)) return 'file-code';
  
  // Documents
  if (isDocument(mimeType)) return 'file-text';
  
  // Default
  return 'file';
}

/**
 * Extract metadata from file
 */
export function extractFileMetadata(file: globalThis.File): Promise<FileMetadata> {
  return new Promise((resolve) => {
    const metadata: FileMetadata = {};
    
    // For images, extract dimensions
    if (isImage(file.type)) {
      const img = new Image();
      img.onload = () => {
        metadata.width = img.width;
        metadata.height = img.height;
        URL.revokeObjectURL(img.src);
        resolve(metadata);
      };
      img.onerror = () => {
        URL.revokeObjectURL(img.src);
        resolve(metadata);
      };
      img.src = URL.createObjectURL(file);
    }
    // For videos, extract dimensions and duration
    else if (isVideo(file.type)) {
      const video = document.createElement('video');
      video.onloadedmetadata = () => {
        metadata.width = video.videoWidth;
        metadata.height = video.videoHeight;
        metadata.duration = video.duration;
        URL.revokeObjectURL(video.src);
        resolve(metadata);
      };
      video.onerror = () => {
        URL.revokeObjectURL(video.src);
        resolve(metadata);
      };
      video.src = URL.createObjectURL(file);
    }
    // For audio, extract duration
    else if (isAudio(file.type)) {
      const audio = new Audio();
      audio.onloadedmetadata = () => {
        metadata.duration = audio.duration;
        URL.revokeObjectURL(audio.src);
        resolve(metadata);
      };
      audio.onerror = () => {
        URL.revokeObjectURL(audio.src);
        resolve(metadata);
      };
      audio.src = URL.createObjectURL(file);
    }
    else {
      resolve(metadata);
    }
  });
}

/**
 * Calculate file hash (simple client-side hash)
 */
export function calculateFileHash(file: globalThis.File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    
    reader.onload = async () => {
      try {
        const buffer = reader.result as ArrayBuffer;
        const hashBuffer = await crypto.subtle.digest('SHA-256', buffer);
        const hashArray = Array.from(new Uint8Array(hashBuffer));
        const hashHex = hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
        resolve(hashHex);
      } catch (error) {
        reject(error);
      }
    };
    
    reader.onerror = () => reject(reader.error);
    reader.readAsArrayBuffer(file);
  });
}

/**
 * Check if file is processing
 */
export function isFileProcessing(file: File): boolean {
  return file.processingStatus === PROCESSING_STATUS.PENDING || 
         file.processingStatus === PROCESSING_STATUS.PROCESSING;
}

/**
 * Check if file processing failed
 */
export function hasFileProcessingFailed(file: File): boolean {
  return file.processingStatus === PROCESSING_STATUS.FAILED;
}

/**
 * Check if file processing is complete
 */
export function isFileProcessingComplete(file: File): boolean {
  return file.processingStatus === PROCESSING_STATUS.COMPLETED;
}

/**
 * Get file display name (truncated if too long)
 */
export function getFileDisplayName(filename: string, maxLength: number = 30): string {
  if (filename.length <= maxLength) {
    return filename;
  }
  
  const extension = getFileExtension(filename);
  const nameWithoutExt = getFileNameWithoutExtension(filename);
  const availableLength = maxLength - (extension ? extension.length + 4 : 3); // +4 for "..." and "."
  
  if (availableLength <= 0) {
    return '...' + (extension ? `.${extension}` : '');
  }
  
  return nameWithoutExt.slice(0, availableLength) + '...' + (extension ? `.${extension}` : '');
}

/**
 * Format file info for display
 */
export function formatFileInfo(file: File): string {
  const parts: string[] = [];
  
  // Add file size
  parts.push(formatBytes(file.size));
  
  // Add dimensions for images/videos
  if (file.metadata.width && file.metadata.height) {
    parts.push(`${file.metadata.width}×${file.metadata.height}`);
  }
  
  // Add duration for videos/audio
  if (file.metadata.duration) {
    const minutes = Math.floor(file.metadata.duration / 60);
    const seconds = Math.floor(file.metadata.duration % 60);
    parts.push(`${minutes}:${seconds.toString().padStart(2, '0')}`);
  }
  
  // Add page count for PDFs
  if (file.metadata.pages) {
    parts.push(`${file.metadata.pages} page${file.metadata.pages !== 1 ? 's' : ''}`);
  }
  
  return parts.join(' • ');
}

/**
 * Get file type description
 */
export function getFileTypeDescription(mimeType: string): string {
  switch (mimeType) {
    case 'image/jpeg':
      return 'JPEG Image';
    case 'image/png':
      return 'PNG Image';
    case 'image/gif':
      return 'GIF Image';
    case 'image/webp':
      return 'WebP Image';
    case 'image/svg+xml':
      return 'SVG Image';
    case 'video/mp4':
      return 'MP4 Video';
    case 'video/avi':
      return 'AVI Video';
    case 'video/mov':
      return 'QuickTime Video';
    case 'video/webm':
      return 'WebM Video';
    case 'audio/mpeg':
      return 'MP3 Audio';
    case 'audio/wav':
      return 'WAV Audio';
    case 'audio/ogg':
      return 'OGG Audio';
    case 'application/pdf':
      return 'PDF Document';
    case 'application/msword':
      return 'Word Document';
    case 'application/vnd.openxmlformats-officedocument.wordprocessingml.document':
      return 'Word Document';
    case 'application/vnd.ms-excel':
      return 'Excel Spreadsheet';
    case 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet':
      return 'Excel Spreadsheet';
    case 'application/vnd.ms-powerpoint':
      return 'PowerPoint Presentation';
    case 'application/vnd.openxmlformats-officedocument.presentationml.presentation':
      return 'PowerPoint Presentation';
    case 'text/plain':
      return 'Text File';
    case 'text/csv':
      return 'CSV File';
    case 'application/json':
      return 'JSON File';
    case 'application/xml':
      return 'XML File';
    case 'text/html':
      return 'HTML File';
    case 'text/css':
      return 'CSS File';
    case 'text/javascript':
      return 'JavaScript File';
    case 'application/zip':
      return 'ZIP Archive';
    case 'application/x-rar-compressed':
      return 'RAR Archive';
    case 'application/x-7z-compressed':
      return '7Z Archive';
    default:
      const category = getFileCategory(mimeType);
      if (category !== 'OTHER') {
        return `${category.charAt(0).toUpperCase() + category.slice(1).toLowerCase()} File`;
      }
      return 'File';
  }
}

/**
 * Sort files by various criteria
 */
export function sortFiles(files: File[], sortBy: string, sortOrder: 'asc' | 'desc' = 'asc'): File[] {
  const sorted = [...files].sort((a, b) => {
    let comparison = 0;
    
    switch (sortBy) {
      case 'name':
        comparison = a.name.localeCompare(b.name);
        break;
      case 'size':
        comparison = a.size - b.size;
        break;
      case 'modified':
        comparison = new Date(a.updatedAt).getTime() - new Date(b.updatedAt).getTime();
        break;
      case 'created':
        comparison = new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime();
        break;
      case 'type':
        comparison = a.mimeType.localeCompare(b.mimeType);
        break;
      default:
        comparison = a.name.localeCompare(b.name);
    }
    
    return sortOrder === 'desc' ? -comparison : comparison;
  });
  
  return sorted;
}

/**
 * Filter files by search query
 */
export function filterFiles(files: File[], query: string): File[] {
  if (!query.trim()) return files;
  
  const searchTerm = query.toLowerCase();
  
  return files.filter(file => {
    // Search in filename
    if (file.name.toLowerCase().includes(searchTerm)) return true;
    
    // Search in OCR text if available
    if (file.ocrText && file.ocrText.toLowerCase().includes(searchTerm)) return true;
    
    // Search in tags
    if (file.tags.some(tag => tag.toLowerCase().includes(searchTerm))) return true;
    
    // Search in file type
    if (getFileTypeDescription(file.mimeType).toLowerCase().includes(searchTerm)) return true;
    
    return false;
  });
}

/**
 * Group files by category
 */
export function groupFilesByCategory(files: File[]): Record<string, File[]> {
  const groups: Record<string, File[]> = {};
  
  files.forEach(file => {
    const category = getFileCategory(file.mimeType);
    if (!groups[category]) {
      groups[category] = [];
    }
    groups[category].push(file);
  });
  
  return groups;
}

/**
 * Calculate total size of files
 */
export function calculateTotalSize(files: File[]): number {
  return files.reduce((total, file) => total + file.size, 0);
}

/**
 * Get file statistics
 */
export function getFileStatistics(files: File[]): {
  totalFiles: number;
  totalSize: number;
  categories: Record<string, { count: number; size: number }>;
  averageSize: number;
  largestFile: File | null;
  smallestFile: File | null;
} {
  if (files.length === 0) {
    return {
      totalFiles: 0,
      totalSize: 0,
      categories: {},
      averageSize: 0,
      largestFile: null,
      smallestFile: null,
    };
  }
  
  const totalSize = calculateTotalSize(files);
  const categories: Record<string, { count: number; size: number }> = {};
  
  let largestFile = files[0];
  let smallestFile = files[0];
  
  files.forEach(file => {
    const category = getFileCategory(file.mimeType);
    
    if (!categories[category]) {
      categories[category] = { count: 0, size: 0 };
    }
    
    categories[category].count++;
    categories[category].size += file.size;
    
    if (file.size > largestFile.size) {
      largestFile = file;
    }
    
    if (file.size < smallestFile.size) {
      smallestFile = file;
    }
  });
  
  return {
    totalFiles: files.length,
    totalSize,
    categories,
    averageSize: totalSize / files.length,
    largestFile,
    smallestFile,
  };
}