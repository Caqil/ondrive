// lib/hooks/use-file-upload.ts
import { useUploadStore, useUploadActions } from '@/store/upload-store';

/**
 * Simple hook that re-exports upload store functionality
 * Maintains consistency with existing store patterns
 */
export const useFileUpload = () => {
  const store = useUploadStore();
  const actions = useUploadActions();

  return {
    // State
    queue: store.queue,
    activeUploads: store.activeUploads,
    completedUploads: store.completedUploads,
    failedUploads: store.failedUploads,
    isUploading: store.isUploading,
    isPaused: store.isPaused,
    maxConcurrentUploads: store.maxConcurrentUploads,
    totalProgress: store.totalProgress,
    totalSpeed: store.totalSpeed,
    config: store.config,
    isDragOver: store.isDragOver,
    dropZoneActive: store.dropZoneActive,
    totalUploaded: store.totalUploaded,
    totalFiles: store.totalFiles,
    successCount: store.successCount,
    errorCount: store.errorCount,

    // Actions
    ...actions,
  };
};