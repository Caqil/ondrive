// store/file-store.ts
import { create } from 'zustand';
import { devtools, subscribeWithSelector } from 'zustand/middleware';
import { immer } from 'zustand/middleware/immer';
import type { 
  File, 
  Folder, 
  FolderContents,
  FolderTreeNode,
  FileSearchParams,
  FileStats,
  CreateFolderRequest,
  ObjectId 
} from '@/types';

interface FileState {
  // Current State
  currentFolder: Folder | null;
  currentFolderId: ObjectId | null;
  files: File[];
  folders: Folder[];
  folderTree: FolderTreeNode[];
  breadcrumb: { id: ObjectId; name: string; path: string }[];
  
  // Selection State
  selectedFiles: ObjectId[];
  selectedFolders: ObjectId[];
  
  // View State
  viewMode: 'grid' | 'list';
  sortBy: 'name' | 'size' | 'modified' | 'type';
  sortOrder: 'asc' | 'desc';
  showHidden: boolean;
  
  // Loading States
  isLoading: boolean;
  isLoadingTree: boolean;
  isCreatingFolder: boolean;
  error: string | null;
  
  // Stats
  stats: FileStats | null;
  
  // Starred Files
  starredFiles: File[];
  isLoadingStarred: boolean;
  
  // Recent Files
  recentFiles: File[];
  isLoadingRecent: boolean;
  
  // Trash
  trashedFiles: File[];
  trashedFolders: Folder[];
  isLoadingTrash: boolean;
  
  // Search
  searchResults: File[];
  searchQuery: string;
  isSearching: boolean;
  searchFilters: Partial<FileSearchParams>;
  
  // Actions
  loadFolderContents: (folderId?: ObjectId) => Promise<void>;
  loadFolderTree: () => Promise<void>;
  createFolder: (data: CreateFolderRequest) => Promise<Folder>;
  renameFile: (fileId: ObjectId, name: string) => Promise<void>;
  renameFolder: (folderId: ObjectId, name: string) => Promise<void>;
  moveFiles: (fileIds: ObjectId[], targetFolderId: ObjectId) => Promise<void>;
  moveFolders: (folderIds: ObjectId[], targetFolderId: ObjectId) => Promise<void>;
  copyFiles: (fileIds: ObjectId[], targetFolderId: ObjectId) => Promise<void>;
  deleteFiles: (fileIds: ObjectId[]) => Promise<void>;
  deleteFolders: (folderIds: ObjectId[]) => Promise<void>;
  
  // Star Actions
  starFile: (fileId: ObjectId) => Promise<void>;
  unstarFile: (fileId: ObjectId) => Promise<void>;
  loadStarredFiles: () => Promise<void>;
  
  // Trash Actions
  trashFiles: (fileIds: ObjectId[]) => Promise<void>;
  trashFolders: (folderIds: ObjectId[]) => Promise<void>;
  restoreFiles: (fileIds: ObjectId[]) => Promise<void>;
  restoreFolders: (folderIds: ObjectId[]) => Promise<void>;
  emptyTrash: () => Promise<void>;
  loadTrash: () => Promise<void>;
  
  // Recent Files
  loadRecentFiles: () => Promise<void>;
  
  // Search Actions
  searchFiles: (query: string, filters?: Partial<FileSearchParams>) => Promise<void>;
  clearSearch: () => void;
  setSearchFilters: (filters: Partial<FileSearchParams>) => void;
  
  // Selection Actions
  selectFile: (fileId: ObjectId) => void;
  selectFolder: (folderId: ObjectId) => void;
  selectMultipleFiles: (fileIds: ObjectId[]) => void;
  selectMultipleFolders: (folderIds: ObjectId[]) => void;
  unselectFile: (fileId: ObjectId) => void;
  unselectFolder: (folderId: ObjectId) => void;
  selectAllFiles: () => void;
  selectAllFolders: () => void;
  clearSelection: () => void;
  
  // View Actions
  setViewMode: (mode: 'grid' | 'list') => void;
  setSortBy: (sortBy: string, order?: 'asc' | 'desc') => void;
  toggleShowHidden: () => void;
  
  // Stats
  loadStats: () => Promise<void>;
  
  // Utility Actions
  clearError: () => void;
  refreshCurrentFolder: () => Promise<void>;
  navigateToFolder: (folderId: ObjectId) => Promise<void>;
}

export const useFileStore = create<FileState>()(
  devtools(
    subscribeWithSelector(
      immer((set, get) => ({
        // Initial State
        currentFolder: null,
        currentFolderId: null,
        files: [],
        folders: [],
        folderTree: [],
        breadcrumb: [],
        selectedFiles: [],
        selectedFolders: [],
        viewMode: 'grid',
        sortBy: 'name',
        sortOrder: 'asc',
        showHidden: false,
        isLoading: false,
        isLoadingTree: false,
        isCreatingFolder: false,
        error: null,
        stats: null,
        starredFiles: [],
        isLoadingStarred: false,
        recentFiles: [],
        isLoadingRecent: false,
        trashedFiles: [],
        trashedFolders: [],
        isLoadingTrash: false,
        searchResults: [],
        searchQuery: '',
        isSearching: false,
        searchFilters: {},

        // Load Folder Contents
        loadFolderContents: async (folderId?: ObjectId) => {
          set((state) => {
            state.isLoading = true;
            state.error = null;
            state.currentFolderId = folderId || null;
          });

          try {
            const url = folderId 
              ? `/api/client/folders/${folderId}/contents`
              : '/api/client/folders/root/contents';
            
            const response = await fetch(url);
            const result = await response.json();

            if (!response.ok) {
              throw new Error(result.error || 'Failed to load folder contents');
            }

            const data: FolderContents = result.data;

            set((state) => {
              state.currentFolder = data.folder;
              state.files = data.files;
              state.folders = data.subfolders;
              state.breadcrumb = data.breadcrumb;
              state.isLoading = false;
              state.selectedFiles = [];
              state.selectedFolders = [];
            });

          } catch (error) {
            set((state) => {
              state.error = error instanceof Error ? error.message : 'Failed to load folder';
              state.isLoading = false;
            });
            throw error;
          }
        },

        // Load Folder Tree
        loadFolderTree: async () => {
          set((state) => {
            state.isLoadingTree = true;
            state.error = null;
          });

          try {
            const response = await fetch('/api/client/folders/tree');
            const result = await response.json();

            if (!response.ok) {
              throw new Error(result.error || 'Failed to load folder tree');
            }

            set((state) => {
              state.folderTree = result.data;
              state.isLoadingTree = false;
            });

          } catch (error) {
            set((state) => {
              state.error = error instanceof Error ? error.message : 'Failed to load folder tree';
              state.isLoadingTree = false;
            });
            throw error;
          }
        },

        // Create Folder
        createFolder: async (data: CreateFolderRequest) => {
          set((state) => {
            state.isCreatingFolder = true;
            state.error = null;
          });

          try {
            const response = await fetch('/api/client/folders', {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify({
                ...data,
                parentId: data.parentId || get().currentFolderId,
              }),
            });

            const result = await response.json();

            if (!response.ok) {
              throw new Error(result.error || 'Failed to create folder');
            }

            const newFolder: Folder = result.data;

            set((state) => {
              state.folders.push(newFolder);
              state.isCreatingFolder = false;
            });

            // Refresh folder tree
            get().loadFolderTree();

            return newFolder;

          } catch (error) {
            set((state) => {
              state.error = error instanceof Error ? error.message : 'Failed to create folder';
              state.isCreatingFolder = false;
            });
            throw error;
          }
        },

        // Rename File
        renameFile: async (fileId: ObjectId, name: string) => {
          try {
            const response = await fetch(`/api/client/files/${fileId}`, {
              method: 'PATCH',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify({ name }),
            });

            const result = await response.json();

            if (!response.ok) {
              throw new Error(result.error || 'Failed to rename file');
            }

            set((state) => {
              const fileIndex = state.files.findIndex(f => f._id === fileId);
              if (fileIndex !== -1) {
                state.files[fileIndex].name = name;
              }
            });

          } catch (error) {
            set((state) => {
              state.error = error instanceof Error ? error.message : 'Failed to rename file';
            });
            throw error;
          }
        },

        // Rename Folder
        renameFolder: async (folderId: ObjectId, name: string) => {
          try {
            const response = await fetch(`/api/client/folders/${folderId}`, {
              method: 'PATCH',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify({ name }),
            });

            const result = await response.json();

            if (!response.ok) {
              throw new Error(result.error || 'Failed to rename folder');
            }

            set((state) => {
              const folderIndex = state.folders.findIndex(f => f._id === folderId);
              if (folderIndex !== -1) {
                state.folders[folderIndex].name = name;
              }
            });

            // Refresh folder tree
            get().loadFolderTree();

          } catch (error) {
            set((state) => {
              state.error = error instanceof Error ? error.message : 'Failed to rename folder';
            });
            throw error;
          }
        },

        // Move Files
        moveFiles: async (fileIds: ObjectId[], targetFolderId: ObjectId) => {
          try {
            const response = await fetch('/api/client/files/bulk', {
              method: 'PATCH',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify({
                fileIds,
                action: 'move',
                targetFolderId,
              }),
            });

            const result = await response.json();

            if (!response.ok) {
              throw new Error(result.error || 'Failed to move files');
            }

            // Remove moved files from current view
            set((state) => {
              state.files = state.files.filter(f => !fileIds.includes(f._id));
              state.selectedFiles = [];
            });

          } catch (error) {
            set((state) => {
              state.error = error instanceof Error ? error.message : 'Failed to move files';
            });
            throw error;
          }
        },

        // Move Folders
        moveFolders: async (folderIds: ObjectId[], targetFolderId: ObjectId) => {
          try {
            const response = await fetch('/api/client/folders/bulk', {
              method: 'PATCH',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify({
                folderIds,
                action: 'move',
                targetFolderId,
              }),
            });

            const result = await response.json();

            if (!response.ok) {
              throw new Error(result.error || 'Failed to move folders');
            }

            // Remove moved folders from current view
            set((state) => {
              state.folders = state.folders.filter(f => !folderIds.includes(f._id));
              state.selectedFolders = [];
            });

            // Refresh folder tree
            get().loadFolderTree();

          } catch (error) {
            set((state) => {
              state.error = error instanceof Error ? error.message : 'Failed to move folders';
            });
            throw error;
          }
        },

        // Copy Files
        copyFiles: async (fileIds: ObjectId[], targetFolderId: ObjectId) => {
          try {
            const response = await fetch('/api/client/files/bulk', {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify({
                fileIds,
                action: 'copy',
                targetFolderId,
              }),
            });

            const result = await response.json();

            if (!response.ok) {
              throw new Error(result.error || 'Failed to copy files');
            }

          } catch (error) {
            set((state) => {
              state.error = error instanceof Error ? error.message : 'Failed to copy files';
            });
            throw error;
          }
        },

        // Delete Files
        deleteFiles: async (fileIds: ObjectId[]) => {
          try {
            const response = await fetch('/api/client/files/bulk', {
              method: 'DELETE',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify({ fileIds }),
            });

            const result = await response.json();

            if (!response.ok) {
              throw new Error(result.error || 'Failed to delete files');
            }

            set((state) => {
              state.files = state.files.filter(f => !fileIds.includes(f._id));
              state.selectedFiles = [];
            });

          } catch (error) {
            set((state) => {
              state.error = error instanceof Error ? error.message : 'Failed to delete files';
            });
            throw error;
          }
        },

        // Delete Folders
        deleteFolders: async (folderIds: ObjectId[]) => {
          try {
            const response = await fetch('/api/client/folders/bulk', {
              method: 'DELETE',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify({ folderIds }),
            });

            const result = await response.json();

            if (!response.ok) {
              throw new Error(result.error || 'Failed to delete folders');
            }

            set((state) => {
              state.folders = state.folders.filter(f => !folderIds.includes(f._id));
              state.selectedFolders = [];
            });

            // Refresh folder tree
            get().loadFolderTree();

          } catch (error) {
            set((state) => {
              state.error = error instanceof Error ? error.message : 'Failed to delete folders';
            });
            throw error;
          }
        },

        // Star File
        starFile: async (fileId: ObjectId) => {
          try {
            const response = await fetch(`/api/client/starred/${fileId}`, {
              method: 'POST',
            });

            const result = await response.json();

            if (!response.ok) {
              throw new Error(result.error || 'Failed to star file');
            }

            set((state) => {
              const fileIndex = state.files.findIndex(f => f._id === fileId);
              if (fileIndex !== -1) {
                state.files[fileIndex].isStarred = true;
              }
            });

          } catch (error) {
            set((state) => {
              state.error = error instanceof Error ? error.message : 'Failed to star file';
            });
            throw error;
          }
        },

        // Unstar File
        unstarFile: async (fileId: ObjectId) => {
          try {
            const response = await fetch(`/api/client/starred/${fileId}`, {
              method: 'DELETE',
            });

            const result = await response.json();

            if (!response.ok) {
              throw new Error(result.error || 'Failed to unstar file');
            }

            set((state) => {
              const fileIndex = state.files.findIndex(f => f._id === fileId);
              if (fileIndex !== -1) {
                state.files[fileIndex].isStarred = false;
              }
              state.starredFiles = state.starredFiles.filter(f => f._id !== fileId);
            });

          } catch (error) {
            set((state) => {
              state.error = error instanceof Error ? error.message : 'Failed to unstar file';
            });
            throw error;
          }
        },

        // Load Starred Files
        loadStarredFiles: async () => {
          set((state) => {
            state.isLoadingStarred = true;
            state.error = null;
          });

          try {
            const response = await fetch('/api/client/starred');
            const result = await response.json();

            if (!response.ok) {
              throw new Error(result.error || 'Failed to load starred files');
            }

            set((state) => {
              state.starredFiles = result.data;
              state.isLoadingStarred = false;
            });

          } catch (error) {
            set((state) => {
              state.error = error instanceof Error ? error.message : 'Failed to load starred files';
              state.isLoadingStarred = false;
            });
            throw error;
          }
        },

        // Trash Files
        trashFiles: async (fileIds: ObjectId[]) => {
          try {
            const response = await fetch('/api/client/trash', {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify({ fileIds }),
            });

            const result = await response.json();

            if (!response.ok) {
              throw new Error(result.error || 'Failed to trash files');
            }

            set((state) => {
              state.files = state.files.filter(f => !fileIds.includes(f._id));
              state.selectedFiles = [];
            });

          } catch (error) {
            set((state) => {
              state.error = error instanceof Error ? error.message : 'Failed to trash files';
            });
            throw error;
          }
        },

        // Trash Folders
        trashFolders: async (folderIds: ObjectId[]) => {
          try {
            const response = await fetch('/api/client/trash', {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify({ folderIds }),
            });

            const result = await response.json();

            if (!response.ok) {
              throw new Error(result.error || 'Failed to trash folders');
            }

            set((state) => {
              state.folders = state.folders.filter(f => !folderIds.includes(f._id));
              state.selectedFolders = [];
            });

            // Refresh folder tree
            get().loadFolderTree();

          } catch (error) {
            set((state) => {
              state.error = error instanceof Error ? error.message : 'Failed to trash folders';
            });
            throw error;
          }
        },

        // Restore Files
        restoreFiles: async (fileIds: ObjectId[]) => {
          try {
            const response = await fetch('/api/client/trash/restore', {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify({ fileIds }),
            });

            const result = await response.json();

            if (!response.ok) {
              throw new Error(result.error || 'Failed to restore files');
            }

            set((state) => {
              state.trashedFiles = state.trashedFiles.filter(f => !fileIds.includes(f._id));
            });

          } catch (error) {
            set((state) => {
              state.error = error instanceof Error ? error.message : 'Failed to restore files';
            });
            throw error;
          }
        },

        // Restore Folders
        restoreFolders: async (folderIds: ObjectId[]) => {
          try {
            const response = await fetch('/api/client/trash/restore', {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify({ folderIds }),
            });

            const result = await response.json();

            if (!response.ok) {
              throw new Error(result.error || 'Failed to restore folders');
            }

            set((state) => {
              state.trashedFolders = state.trashedFolders.filter(f => !folderIds.includes(f._id));
            });

            // Refresh folder tree
            get().loadFolderTree();

          } catch (error) {
            set((state) => {
              state.error = error instanceof Error ? error.message : 'Failed to restore folders';
            });
            throw error;
          }
        },

        // Empty Trash
        emptyTrash: async () => {
          try {
            const response = await fetch('/api/client/trash/empty', {
              method: 'DELETE',
            });

            const result = await response.json();

            if (!response.ok) {
              throw new Error(result.error || 'Failed to empty trash');
            }

            set((state) => {
              state.trashedFiles = [];
              state.trashedFolders = [];
            });

          } catch (error) {
            set((state) => {
              state.error = error instanceof Error ? error.message : 'Failed to empty trash';
            });
            throw error;
          }
        },

        // Load Trash
        loadTrash: async () => {
          set((state) => {
            state.isLoadingTrash = true;
            state.error = null;
          });

          try {
            const response = await fetch('/api/client/trash');
            const result = await response.json();

            if (!response.ok) {
              throw new Error(result.error || 'Failed to load trash');
            }

            set((state) => {
              state.trashedFiles = result.data.files;
              state.trashedFolders = result.data.folders;
              state.isLoadingTrash = false;
            });

          } catch (error) {
            set((state) => {
              state.error = error instanceof Error ? error.message : 'Failed to load trash';
              state.isLoadingTrash = false;
            });
            throw error;
          }
        },

        // Load Recent Files
        loadRecentFiles: async () => {
          set((state) => {
            state.isLoadingRecent = true;
            state.error = null;
          });

          try {
            const response = await fetch('/api/client/files/recent');
            const result = await response.json();

            if (!response.ok) {
              throw new Error(result.error || 'Failed to load recent files');
            }

            set((state) => {
              state.recentFiles = result.data;
              state.isLoadingRecent = false;
            });

          } catch (error) {
            set((state) => {
              state.error = error instanceof Error ? error.message : 'Failed to load recent files';
              state.isLoadingRecent = false;
            });
            throw error;
          }
        },

        // Search Files
        searchFiles: async (query: string, filters?: Partial<FileSearchParams>) => {
          set((state) => {
            state.isSearching = true;
            state.searchQuery = query;
            state.searchFilters = { ...state.searchFilters, ...filters };
            state.error = null;
          });

          try {
            const searchParams = new URLSearchParams({
              query,
              ...Object.fromEntries(
                Object.entries({ ...get().searchFilters, ...filters })
                  .filter(([_, value]) => value !== undefined && value !== '')
                  .map(([key, value]) => [key, String(value)])
              ),
            });

            const response = await fetch(`/api/client/search?${searchParams}`);
            const result = await response.json();

            if (!response.ok) {
              throw new Error(result.error || 'Search failed');
            }

            set((state) => {
              state.searchResults = result.data;
              state.isSearching = false;
            });

          } catch (error) {
            set((state) => {
              state.error = error instanceof Error ? error.message : 'Search failed';
              state.isSearching = false;
            });
            throw error;
          }
        },

        // Clear Search
        clearSearch: () => {
          set((state) => {
            state.searchResults = [];
            state.searchQuery = '';
            state.searchFilters = {};
          });
        },

        // Set Search Filters
        setSearchFilters: (filters: Partial<FileSearchParams>) => {
          set((state) => {
            state.searchFilters = { ...state.searchFilters, ...filters };
          });
        },

        // Selection Actions
        selectFile: (fileId: ObjectId) => {
          set((state) => {
            if (!state.selectedFiles.includes(fileId)) {
              state.selectedFiles.push(fileId);
            }
          });
        },

        selectFolder: (folderId: ObjectId) => {
          set((state) => {
            if (!state.selectedFolders.includes(folderId)) {
              state.selectedFolders.push(folderId);
            }
          });
        },

        selectMultipleFiles: (fileIds: ObjectId[]) => {
          set((state) => {
            state.selectedFiles = [...new Set([...state.selectedFiles, ...fileIds])];
          });
        },

        selectMultipleFolders: (folderIds: ObjectId[]) => {
          set((state) => {
            state.selectedFolders = [...new Set([...state.selectedFolders, ...folderIds])];
          });
        },

        unselectFile: (fileId: ObjectId) => {
          set((state) => {
            state.selectedFiles = state.selectedFiles.filter(id => id !== fileId);
          });
        },

        unselectFolder: (folderId: ObjectId) => {
          set((state) => {
            state.selectedFolders = state.selectedFolders.filter(id => id !== folderId);
          });
        },

        selectAllFiles: () => {
          set((state) => {
            state.selectedFiles = state.files.map(f => f._id);
          });
        },

        selectAllFolders: () => {
          set((state) => {
            state.selectedFolders = state.folders.map(f => f._id);
          });
        },

        clearSelection: () => {
          set((state) => {
            state.selectedFiles = [];
            state.selectedFolders = [];
          });
        },

        // View Actions
        setViewMode: (mode: 'grid' | 'list') => {
          set((state) => {
            state.viewMode = mode;
          });
        },

        setSortBy: (sortBy: string, order?: 'asc' | 'desc') => {
          set((state) => {
            state.sortBy = sortBy as any;
            state.sortOrder = order || (state.sortBy === sortBy && state.sortOrder === 'asc' ? 'desc' : 'asc');
          });
        },

        toggleShowHidden: () => {
          set((state) => {
            state.showHidden = !state.showHidden;
          });
        },

        // Load Stats
        loadStats: async () => {
          try {
            const response = await fetch('/api/client/files/stats');
            const result = await response.json();

            if (!response.ok) {
              throw new Error(result.error || 'Failed to load stats');
            }

            set((state) => {
              state.stats = result.data;
            });

          } catch (error) {
            set((state) => {
              state.error = error instanceof Error ? error.message : 'Failed to load stats';
            });
            throw error;
          }
        },

        // Utility Actions
        clearError: () => {
          set((state) => {
            state.error = null;
          });
        },

        refreshCurrentFolder: async () => {
          const { currentFolderId } = get();
          await get().loadFolderContents(currentFolderId || undefined);
        },

        navigateToFolder: async (folderId: ObjectId) => {
          await get().loadFolderContents(folderId);
        },
      }))
    ),
    { name: 'file-store' }
  )
);

// Selectors
export const useCurrentFolder = () => useFileStore((state) => state.currentFolder);
export const useFiles = () => useFileStore((state) => state.files);
export const useFolders = () => useFileStore((state) => state.folders);
export const useFolderTree = () => useFileStore((state) => state.folderTree);
export const useSelectedFiles = () => useFileStore((state) => state.selectedFiles);
export const useSelectedFolders = () => useFileStore((state) => state.selectedFolders);
export const useViewMode = () => useFileStore((state) => state.viewMode);
export const useFileLoading = () => useFileStore((state) => state.isLoading);
export const useFileError = () => useFileStore((state) => state.error);
export const useSearchResults = () => useFileStore((state) => state.searchResults);
export const useIsSearching = () => useFileStore((state) => state.isSearching);

// Actions
export const useFileActions = () => {
  const store = useFileStore();
  return {
    loadFolderContents: store.loadFolderContents,
    createFolder: store.createFolder,
    renameFile: store.renameFile,
    renameFolder: store.renameFolder,
    moveFiles: store.moveFiles,
    moveFolders: store.moveFolders,
    deleteFiles: store.deleteFiles,
    deleteFolders: store.deleteFolders,
    starFile: store.starFile,
    unstarFile: store.unstarFile,
    searchFiles: store.searchFiles,
    clearSearch: store.clearSearch,
    selectFile: store.selectFile,
    clearSelection: store.clearSelection,
    setViewMode: store.setViewMode,
    clearError: store.clearError,
  };
};