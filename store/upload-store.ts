// store/upload-store.ts
import { create } from 'zustand';
import { devtools, subscribeWithSelector } from 'zustand/middleware';
import { immer } from 'zustand/middleware/immer';
import type { 
  ObjectId, 
  FileUploadRequest, 
  FileUploadResponse,
  UploadConfig 
} from '@/types';

export interface UploadItem {
  id: string;
  file: File;
  name: string;
  size: number;
  mimeType: string;
  folderId: ObjectId;
  status: 'pending' | 'uploading' | 'processing' | 'completed' | 'failed' | 'cancelled';
  progress: number;
  uploadedBytes: number;
  speed: number; // bytes per second
  eta: number; // estimated time remaining in seconds
  error?: string;
  uploadId?: string;
  fileId?: ObjectId;
  url?: string;
  thumbnailUrl?: string;
  startTime?: number;
  endTime?: number;
  retryCount: number;
  chunkIndex?: number;
  totalChunks?: number;
}

interface UploadState {
  // Upload Queue
  queue: UploadItem[];
  activeUploads: string[];
  completedUploads: string[];
  failedUploads: string[];
  
  // Global State
  isUploading: boolean;
  isPaused: boolean;
  maxConcurrentUploads: number;
  totalProgress: number;
  totalSpeed: number;
  
  // Configuration
  config: UploadConfig | null;
  
  // Drag & Drop State
  isDragOver: boolean;
  dropZoneActive: boolean;
  
  // Statistics
  totalUploaded: number;
  totalFiles: number;
  successCount: number;
  errorCount: number;
  
  // Actions
  addFiles: (files: File[], folderId: ObjectId) => void;
  removeUpload: (id: string) => void;
  clearCompleted: () => void;
  clearFailed: () => void;
  clearAll: () => void;
  
  // Upload Control
  startUpload: (id: string) => Promise<void>;
  pauseUpload: (id: string) => void;
  resumeUpload: (id: string) => void;
  cancelUpload: (id: string) => void;
  retryUpload: (id: string) => Promise<void>;
  
  // Queue Management
  startQueue: () => void;
  pauseQueue: () => void;
  resumeQueue: () => void;
  processQueue: () => void;
  
  // Upload Methods - MISSING METHODS ADDED HERE
  performUpload: (id: string, uploadResponse: FileUploadResponse) => Promise<void>;
  performChunkedUpload: (id: string, uploadResponse: FileUploadResponse) => Promise<void>;
  performDirectUpload: (id: string, uploadResponse: FileUploadResponse) => Promise<void>;
  uploadChunk: (id: string, chunk: Blob, chunkIndex: number, uploadId: string) => Promise<void>;
  completeChunkedUpload: (id: string, uploadId: string) => Promise<void>;
  
  // Configuration
  loadConfig: () => Promise<void>;
  setMaxConcurrentUploads: (count: number) => void;
  
  // Drag & Drop
  setDragOver: (isDragOver: boolean) => void;
  setDropZoneActive: (active: boolean) => void;
  handleDrop: (files: FileList, folderId: ObjectId) => void;
  
  // Utility
  getUploadById: (id: string) => UploadItem | undefined;
  getUploadsByStatus: (status: UploadItem['status']) => UploadItem[];
  calculateTotalProgress: () => void;
  updateUploadProgress: (id: string, progress: number, uploadedBytes: number) => void;
  setUploadStatus: (id: string, status: UploadItem['status'], error?: string) => void;
}

// Utility functions
const generateUploadId = () => Math.random().toString(36).substring(2) + Date.now().toString(36);

const formatSpeed = (bytesPerSecond: number): string => {
  const units = ['B/s', 'KB/s', 'MB/s', 'GB/s'];
  let size = bytesPerSecond;
  let unitIndex = 0;
  
  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024;
    unitIndex++;
  }
  
  return `${size.toFixed(1)} ${units[unitIndex]}`;
};

const calculateETA = (remainingBytes: number, speed: number): number => {
  if (speed === 0) return Infinity;
  return Math.round(remainingBytes / speed);
};

export const useUploadStore = create<UploadState>()(
  devtools(
    subscribeWithSelector(
      immer((set, get) => ({
        // Initial State
        queue: [],
        activeUploads: [],
        completedUploads: [],
        failedUploads: [],
        isUploading: false,
        isPaused: false,
        maxConcurrentUploads: 3,
        totalProgress: 0,
        totalSpeed: 0,
        config: null,
        isDragOver: false,
        dropZoneActive: false,
        totalUploaded: 0,
        totalFiles: 0,
        successCount: 0,
        errorCount: 0,

        // Add Files to Queue
        addFiles: (files: File[], folderId: ObjectId) => {
          const newUploads: UploadItem[] = Array.from(files).map((file) => ({
            id: generateUploadId(),
            file,
            name: file.name,
            size: file.size,
            mimeType: file.type,
            folderId,
            status: 'pending',
            progress: 0,
            uploadedBytes: 0,
            speed: 0,
            eta: 0,
            retryCount: 0,
          }));

          set((state) => {
            state.queue.push(...newUploads);
            state.totalFiles += newUploads.length;
          });

          // Auto-start if not paused
          if (!get().isPaused) {
            get().processQueue();
          }
        },

        // Remove Upload from Queue
        removeUpload: (id: string) => {
          set((state) => {
            state.queue = state.queue.filter(u => u.id !== id);
            state.activeUploads = state.activeUploads.filter(uid => uid !== id);
            state.completedUploads = state.completedUploads.filter(uid => uid !== id);
            state.failedUploads = state.failedUploads.filter(uid => uid !== id);
          });
        },

        // Clear Completed Uploads
        clearCompleted: () => {
          set((state) => {
            state.queue = state.queue.filter(u => u.status !== 'completed');
            state.completedUploads = [];
          });
        },

        // Clear Failed Uploads
        clearFailed: () => {
          set((state) => {
            state.queue = state.queue.filter(u => u.status !== 'failed');
            state.failedUploads = [];
          });
        },

        // Clear All Uploads
        clearAll: () => {
          set((state) => {
            state.queue = [];
            state.activeUploads = [];
            state.completedUploads = [];
            state.failedUploads = [];
            state.totalFiles = 0;
            state.successCount = 0;
            state.errorCount = 0;
            state.isUploading = false;
          });
        },

        // Start Single Upload
        startUpload: async (id: string) => {
          const upload = get().getUploadById(id);
          if (!upload || upload.status !== 'pending') return;

          set((state) => {
            state.activeUploads.push(id);
            state.isUploading = true;
            const uploadIndex = state.queue.findIndex(u => u.id === id);
            if (uploadIndex !== -1) {
              state.queue[uploadIndex].status = 'uploading';
              state.queue[uploadIndex].startTime = Date.now();
            }
          });

          try {
            // Create file upload request
            const uploadRequest: FileUploadRequest = {
              name: upload.name,
              size: upload.size,
              mimeType: upload.mimeType,
              folderId: upload.folderId,
            };

            // Get upload URL from server
            const response = await fetch('/api/client/files/upload', {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify(uploadRequest),
            });

            if (!response.ok) {
              const error = await response.json();
              throw new Error(error.error || 'Failed to get upload URL');
            }

            const uploadResponse: FileUploadResponse = await response.json();

            // Update upload with server response
            set((state) => {
              const uploadIndex = state.queue.findIndex(u => u.id === id);
              if (uploadIndex !== -1) {
                state.queue[uploadIndex].uploadId = uploadResponse.uploadId;
                state.queue[uploadIndex].fileId = uploadResponse.fileId;
              }
            });

            // Perform the actual upload
            await get().performUpload(id, uploadResponse);

            // Mark as completed
            get().setUploadStatus(id, 'completed');
            
            set((state) => {
              state.completedUploads.push(id);
              state.successCount++;
              const uploadIndex = state.queue.findIndex(u => u.id === id);
              if (uploadIndex !== -1) {
                state.queue[uploadIndex].endTime = Date.now();
              }
            });

          } catch (error) {
            const upload = get().getUploadById(id);
            if (upload && upload.retryCount < 3) {
              // Retry logic
              set((state) => {
                const uploadIndex = state.queue.findIndex(u => u.id === id);
                if (uploadIndex !== -1) {
                  state.queue[uploadIndex].retryCount++;
                }
              });
              
              // Retry after delay
              setTimeout(() => {
                get().retryUpload(id);
              }, 2000 * (upload.retryCount + 1));
            } else {
              get().setUploadStatus(id, 'failed', error instanceof Error ? error.message : 'Upload failed');
              set((state) => {
                state.failedUploads.push(id);
                state.errorCount++;
              });
            }
          } finally {
            set((state) => {
              state.activeUploads = state.activeUploads.filter(uid => uid !== id);
              if (state.activeUploads.length === 0) {
                state.isUploading = false;
              }
            });

            get().calculateTotalProgress();
            
            // Start next upload in queue
            get().processQueue();
          }
        },

        // Perform Upload (chunk or direct)
        performUpload: async (id: string, uploadResponse: FileUploadResponse) => {
          const upload = get().getUploadById(id);
          if (!upload) throw new Error('Upload not found');

          const { config } = get();
          if (!config) throw new Error('Upload configuration not loaded');

          // Determine if we need chunked upload
          const useChunkedUpload = upload.size > config.chunkSize;

          if (useChunkedUpload) {
            await get().performChunkedUpload(id, uploadResponse);
          } else {
            await get().performDirectUpload(id, uploadResponse);
          }
        },

        // Direct Upload
        performDirectUpload: async (id: string, uploadResponse: FileUploadResponse) => {
          const upload = get().getUploadById(id);
          if (!upload) throw new Error('Upload not found');

          const xhr = new XMLHttpRequest();
          
          return new Promise<void>((resolve, reject) => {
            xhr.upload.addEventListener('progress', (event) => {
              if (event.lengthComputable) {
                const progress = (event.loaded / event.total) * 100;
                get().updateUploadProgress(id, progress, event.loaded);
              }
            });

            xhr.addEventListener('load', () => {
              if (xhr.status >= 200 && xhr.status < 300) {
                resolve();
              } else {
                reject(new Error(`Upload failed with status ${xhr.status}`));
              }
            });

            xhr.addEventListener('error', () => {
              reject(new Error('Upload failed'));
            });

            xhr.addEventListener('abort', () => {
              reject(new Error('Upload cancelled'));
            });

            const formData = new FormData();
            formData.append('file', upload.file);

            xhr.open('PUT', uploadResponse.uploadUrl);
            xhr.send(formData);

            // Store XMLHttpRequest for cancellation
            set((state) => {
              const uploadIndex = state.queue.findIndex(u => u.id === id);
              if (uploadIndex !== -1) {
                (state.queue[uploadIndex] as any).xhr = xhr;
              }
            });
          });
        },

        // Chunked Upload
        performChunkedUpload: async (id: string, uploadResponse: FileUploadResponse) => {
          const upload = get().getUploadById(id);
          const { config } = get();
          if (!upload || !config) throw new Error('Upload or config not found');

          const chunkSize = config.chunkSize;
          const totalChunks = Math.ceil(upload.size / chunkSize);

          set((state) => {
            const uploadIndex = state.queue.findIndex(u => u.id === id);
            if (uploadIndex !== -1) {
              state.queue[uploadIndex].totalChunks = totalChunks;
              state.queue[uploadIndex].chunkIndex = 0;
            }
          });

          for (let chunkIndex = 0; chunkIndex < totalChunks; chunkIndex++) {
            const currentUpload = get().getUploadById(id);
            if (!currentUpload || currentUpload.status === 'cancelled') {
              throw new Error('Upload cancelled');
            }

            const start = chunkIndex * chunkSize;
            const end = Math.min(start + chunkSize, upload.size);
            const chunk = upload.file.slice(start, end);

            set((state) => {
              const uploadIndex = state.queue.findIndex(u => u.id === id);
              if (uploadIndex !== -1) {
                state.queue[uploadIndex].chunkIndex = chunkIndex;
              }
            });

            await get().uploadChunk(id, chunk, chunkIndex, uploadResponse.uploadId!);

            const progress = ((chunkIndex + 1) / totalChunks) * 100;
            get().updateUploadProgress(id, progress, end);
          }

          // Complete chunked upload
          await get().completeChunkedUpload(id, uploadResponse.uploadId!);
        },

        // Upload Single Chunk
        uploadChunk: async (id: string, chunk: Blob, chunkIndex: number, uploadId: string) => {
          const formData = new FormData();
          formData.append('chunk', chunk);
          formData.append('chunkIndex', chunkIndex.toString());
          formData.append('uploadId', uploadId);

          const response = await fetch('/api/client/files/upload/chunk', {
            method: 'POST',
            body: formData,
          });

          if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || `Failed to upload chunk ${chunkIndex}`);
          }
        },

        // Complete Chunked Upload
        completeChunkedUpload: async (id: string, uploadId: string) => {
          const response = await fetch('/api/client/files/upload/complete', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ uploadId }),
          });

          if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || 'Failed to complete upload');
          }

          const result = await response.json();
          
          // Update upload with final file info
          set((state) => {
            const uploadIndex = state.queue.findIndex(u => u.id === id);
            if (uploadIndex !== -1) {
              state.queue[uploadIndex].url = result.data.url;
              state.queue[uploadIndex].thumbnailUrl = result.data.thumbnailUrl;
            }
          });
        },

        // Pause Upload
        pauseUpload: (id: string) => {
          const upload = get().getUploadById(id);
          if (upload && upload.status === 'uploading') {
            // Cancel XMLHttpRequest if exists
            const xhr = (upload as any).xhr;
            if (xhr) {
              xhr.abort();
            }

            get().setUploadStatus(id, 'pending');
            
            set((state) => {
              state.activeUploads = state.activeUploads.filter(uid => uid !== id);
              if (state.activeUploads.length === 0) {
                state.isUploading = false;
              }
            });
          }
        },

        // Resume Upload
        resumeUpload: (id: string) => {
          const upload = get().getUploadById(id);
          if (upload && upload.status === 'pending') {
            get().startUpload(id);
          }
        },

        // Cancel Upload
        cancelUpload: (id: string) => {
          const upload = get().getUploadById(id);
          if (upload) {
            // Cancel XMLHttpRequest if exists
            const xhr = (upload as any).xhr;
            if (xhr) {
              xhr.abort();
            }

            get().setUploadStatus(id, 'cancelled');
            
            set((state) => {
              state.activeUploads = state.activeUploads.filter(uid => uid !== id);
              if (state.activeUploads.length === 0) {
                state.isUploading = false;
              }
            });
          }
        },

        // Retry Upload
        retryUpload: async (id: string) => {
          const upload = get().getUploadById(id);
          if (upload && (upload.status === 'failed' || upload.status === 'cancelled')) {
            get().setUploadStatus(id, 'pending');
            set((state) => {
              const uploadIndex = state.queue.findIndex(u => u.id === id);
              if (uploadIndex !== -1) {
                state.queue[uploadIndex].progress = 0;
                state.queue[uploadIndex].uploadedBytes = 0;
                state.queue[uploadIndex].error = undefined;
              }
            });
            
            await get().startUpload(id);
          }
        },

        // Process Queue
        processQueue: () => {
          const { queue, activeUploads, maxConcurrentUploads, isPaused } = get();
          
          if (isPaused) return;

          const pendingUploads = queue.filter(u => u.status === 'pending');
          const availableSlots = maxConcurrentUploads - activeUploads.length;

          for (let i = 0; i < Math.min(availableSlots, pendingUploads.length); i++) {
            get().startUpload(pendingUploads[i].id);
          }
        },

        // Start Queue
        startQueue: () => {
          set((state) => {
            state.isPaused = false;
          });
          get().processQueue();
        },

        // Pause Queue
        pauseQueue: () => {
          set((state) => {
            state.isPaused = true;
            
            // Pause all active uploads
            state.activeUploads.forEach(id => {
              get().pauseUpload(id);
            });
          });
        },

        // Resume Queue
        resumeQueue: () => {
          set((state) => {
            state.isPaused = false;
          });
          get().processQueue();
        },

        // Load Configuration
        loadConfig: async () => {
          try {
            const response = await fetch('/api/client/files/upload/config');
            const result = await response.json();

            if (!response.ok) {
              throw new Error(result.error || 'Failed to load upload config');
            }

            set((state) => {
              state.config = result.data;
            });

          } catch (error) {
            console.warn('Failed to load upload config:', error);
            // Use default config
            set((state) => {
              state.config = {
                maxFileSize: 750 * 1024 * 1024 * 1024, // 750GB
                allowedMimeTypes: [],
                chunkSize: 10 * 1024 * 1024, // 10MB
                maxChunks: 1000,
                storageProvider: 'local',
              };
            });
          }
        },

        // Set Max Concurrent Uploads
        setMaxConcurrentUploads: (count: number) => {
          set((state) => {
            state.maxConcurrentUploads = Math.max(1, Math.min(10, count));
          });
        },

        // Drag & Drop
        setDragOver: (isDragOver: boolean) => {
          set((state) => {
            state.isDragOver = isDragOver;
          });
        },

        setDropZoneActive: (active: boolean) => {
          set((state) => {
            state.dropZoneActive = active;
          });
        },

        handleDrop: (files: FileList, folderId: ObjectId) => {
          const fileArray = Array.from(files);
          get().addFiles(fileArray, folderId);
          get().setDragOver(false);
          get().setDropZoneActive(false);
        },

        // Utility Functions
        getUploadById: (id: string) => {
          return get().queue.find(u => u.id === id);
        },

        getUploadsByStatus: (status: UploadItem['status']) => {
          return get().queue.filter(u => u.status === status);
        },

        calculateTotalProgress: () => {
          const { queue } = get();
          if (queue.length === 0) {
            set((state) => {
              state.totalProgress = 0;
            });
            return;
          }

          const totalBytes = queue.reduce((sum, upload) => sum + upload.size, 0);
          const uploadedBytes = queue.reduce((sum, upload) => sum + upload.uploadedBytes, 0);
          const progress = totalBytes > 0 ? (uploadedBytes / totalBytes) * 100 : 0;

          set((state) => {
            state.totalProgress = Math.min(100, Math.max(0, progress));
          });
        },

        updateUploadProgress: (id: string, progress: number, uploadedBytes: number) => {
          set((state) => {
            const uploadIndex = state.queue.findIndex(u => u.id === id);
            if (uploadIndex !== -1) {
              const upload = state.queue[uploadIndex];
              upload.progress = Math.min(100, Math.max(0, progress));
              upload.uploadedBytes = uploadedBytes;

              // Calculate speed and ETA
              if (upload.startTime) {
                const elapsed = (Date.now() - upload.startTime) / 1000; // seconds
                upload.speed = elapsed > 0 ? uploadedBytes / elapsed : 0;
                
                const remainingBytes = upload.size - uploadedBytes;
                upload.eta = calculateETA(remainingBytes, upload.speed);
              }
            }
          });

          // Update total speed
          const activeUploads = get().getUploadsByStatus('uploading');
          const totalSpeed = activeUploads.reduce((sum, upload) => sum + upload.speed, 0);
          
          set((state) => {
            state.totalSpeed = totalSpeed;
          });
        },

        setUploadStatus: (id: string, status: UploadItem['status'], error?: string) => {
          set((state) => {
            const uploadIndex = state.queue.findIndex(u => u.id === id);
            if (uploadIndex !== -1) {
              state.queue[uploadIndex].status = status;
              if (error) {
                state.queue[uploadIndex].error = error;
              }
              if (status === 'completed') {
                state.queue[uploadIndex].progress = 100;
                state.queue[uploadIndex].uploadedBytes = state.queue[uploadIndex].size;
              }
            }
          });
        },
      }))
    ),
    { name: 'upload-store' }
  )
);

// Selectors
export const useUploadQueue = () => useUploadStore((state) => state.queue);
export const useIsUploading = () => useUploadStore((state) => state.isUploading);
export const useUploadProgress = () => useUploadStore((state) => state.totalProgress);
export const useUploadSpeed = () => useUploadStore((state) => state.totalSpeed);
export const useActiveUploads = () => useUploadStore((state) => state.activeUploads);
export const useCompletedUploads = () => useUploadStore((state) => state.completedUploads);
export const useFailedUploads = () => useUploadStore((state) => state.failedUploads);
export const useDragState = () => useUploadStore((state) => ({
  isDragOver: state.isDragOver,
  dropZoneActive: state.dropZoneActive,
}));
export const useUploadStats = () => useUploadStore((state) => ({
  totalFiles: state.totalFiles,
  successCount: state.successCount,
  errorCount: state.errorCount,
  totalUploaded: state.totalUploaded,
}));

// Actions
export const useUploadActions = () => {
  const store = useUploadStore();
  return {
    addFiles: store.addFiles,
    removeUpload: store.removeUpload,
    clearCompleted: store.clearCompleted,
    clearFailed: store.clearFailed,
    clearAll: store.clearAll,
    pauseUpload: store.pauseUpload,
    resumeUpload: store.resumeUpload,
    cancelUpload: store.cancelUpload,
    retryUpload: store.retryUpload,
    startQueue: store.startQueue,
    pauseQueue: store.pauseQueue,
    resumeQueue: store.resumeQueue,
    setDragOver: store.setDragOver,
    setDropZoneActive: store.setDropZoneActive,
    handleDrop: store.handleDrop,
    loadConfig: store.loadConfig,
    setMaxConcurrentUploads: store.setMaxConcurrentUploads,
  };
};