// store/ui-store.ts
import { create } from 'zustand';
import { devtools, persist } from 'zustand/middleware';
import { immer } from 'zustand/middleware/immer';

interface ToastMessage {
  id: string;
  type: 'success' | 'error' | 'warning' | 'info';
  title: string;
  description?: string;
  duration?: number;
  action?: {
    label: string;
    onClick: () => void;
  };
}

interface Modal {
  id: string;
  type: string;
  title: string;
  content?: any;
  size?: 'sm' | 'md' | 'lg' | 'xl' | 'full';
  closable?: boolean;
  onClose?: () => void;
  onConfirm?: () => void;
  confirmText?: string;
  cancelText?: string;
  isLoading?: boolean;
}

interface ContextMenu {
  id: string;
  x: number;
  y: number;
  items: {
    id: string;
    label: string;
    icon?: string;
    onClick: () => void;
    disabled?: boolean;
    divider?: boolean;
    destructive?: boolean;
  }[];
}

interface Sidebar {
  isCollapsed: boolean;
  width: number;
  isResizing: boolean;
}

interface UIState {
  // Theme
  theme: 'light' | 'dark' | 'system';
  actualTheme: 'light' | 'dark';
  
  // Layout
  sidebar: Sidebar;
  showMobileNav: boolean;
  
  // Loading States
  globalLoading: boolean;
  loadingStates: Record<string, boolean>;
  
  // Toasts
  toasts: ToastMessage[];
  
  // Modals
  modals: Modal[];
  
  // Context Menus
  contextMenus: ContextMenu[];
  
  // Command Palette
  commandPaletteOpen: boolean;
  
  // Keyboard Shortcuts
  keyboardShortcutsEnabled: boolean;
  shortcutHelpOpen: boolean;
  
  // Drag and Drop
  isDragging: boolean;
  dragData: any;
  
  // Selection
  multiSelectMode: boolean;
  
  // View Options
  showHiddenFiles: boolean;
  compactMode: boolean;
  
  // Panels
  rightPanelOpen: boolean;
  rightPanelContent: 'details' | 'activity' | 'sharing' | null;
  
  // Actions
  setTheme: (theme: 'light' | 'dark' | 'system') => void;
  toggleTheme: () => void;
  
  // Sidebar
  toggleSidebar: () => void;
  setSidebarCollapsed: (collapsed: boolean) => void;
  setSidebarWidth: (width: number) => void;
  setSidebarResizing: (resizing: boolean) => void;
  
  // Mobile Navigation
  toggleMobileNav: () => void;
  setMobileNav: (show: boolean) => void;
  
  // Loading
  setGlobalLoading: (loading: boolean) => void;
  setLoading: (key: string, loading: boolean) => void;
  clearLoading: (key: string) => void;
  
  // Toasts
  addToast: (toast: Omit<ToastMessage, 'id'>) => string;
  removeToast: (id: string) => void;
  clearToasts: () => void;
  
  // Modals
  openModal: (modal: Omit<Modal, 'id'>) => string;
  closeModal: (id: string) => void;
  closeAllModals: () => void;
  updateModal: (id: string, updates: Partial<Modal>) => void;
  
  // Context Menus
  openContextMenu: (menu: Omit<ContextMenu, 'id'>) => string;
  closeContextMenu: (id: string) => void;
  closeAllContextMenus: () => void;
  
  // Command Palette
  toggleCommandPalette: () => void;
  setCommandPaletteOpen: (open: boolean) => void;
  
  // Keyboard Shortcuts
  toggleKeyboardShortcuts: () => void;
  setKeyboardShortcutsEnabled: (enabled: boolean) => void;
  toggleShortcutHelp: () => void;
  
  // Drag and Drop
  setDragging: (dragging: boolean, data?: any) => void;
  
  // Selection
  setMultiSelectMode: (enabled: boolean) => void;
  
  // View Options
  toggleHiddenFiles: () => void;
  setShowHiddenFiles: (show: boolean) => void;
  toggleCompactMode: () => void;
  setCompactMode: (compact: boolean) => void;
  
  // Panels
  toggleRightPanel: () => void;
  setRightPanelOpen: (open: boolean) => void;
  setRightPanelContent: (content: 'details' | 'activity' | 'sharing' | null) => void;
  
  // Utility
  getModal: (id: string) => Modal | undefined;
  getToast: (id: string) => ToastMessage | undefined;
  isLoading: (key: string) => boolean;
}

// Utility functions
const generateId = () => Math.random().toString(36).substring(2) + Date.now().toString(36);

const getSystemTheme = (): 'light' | 'dark' => {
  if (typeof window !== 'undefined') {
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
  }
  return 'light';
};

export const useUIStore = create<UIState>()(
  devtools(
    persist(
      immer((set, get) => ({
        // Initial State
        theme: 'system',
        actualTheme: getSystemTheme(),
        sidebar: {
          isCollapsed: false,
          width: 280,
          isResizing: false,
        },
        showMobileNav: false,
        globalLoading: false,
        loadingStates: {},
        toasts: [],
        modals: [],
        contextMenus: [],
        commandPaletteOpen: false,
        keyboardShortcutsEnabled: true,
        shortcutHelpOpen: false,
        isDragging: false,
        dragData: null,
        multiSelectMode: false,
        showHiddenFiles: false,
        compactMode: false,
        rightPanelOpen: false,
        rightPanelContent: null,

        // Theme Actions
        setTheme: (theme: 'light' | 'dark' | 'system') => {
          set((state) => {
            state.theme = theme;
            state.actualTheme = theme === 'system' ? getSystemTheme() : theme;
          });

          // Update DOM
          if (typeof window !== 'undefined') {
            const root = window.document.documentElement;
            const actualTheme = theme === 'system' ? getSystemTheme() : theme;
            root.classList.remove('light', 'dark');
            root.classList.add(actualTheme);
          }
        },

        toggleTheme: () => {
          const { theme } = get();
          const newTheme = theme === 'light' ? 'dark' : theme === 'dark' ? 'system' : 'light';
          get().setTheme(newTheme);
        },

        // Sidebar Actions
        toggleSidebar: () => {
          set((state) => {
            state.sidebar.isCollapsed = !state.sidebar.isCollapsed;
          });
        },

        setSidebarCollapsed: (collapsed: boolean) => {
          set((state) => {
            state.sidebar.isCollapsed = collapsed;
          });
        },

        setSidebarWidth: (width: number) => {
          set((state) => {
            state.sidebar.width = Math.max(200, Math.min(400, width));
          });
        },

        setSidebarResizing: (resizing: boolean) => {
          set((state) => {
            state.sidebar.isResizing = resizing;
          });
        },

        // Mobile Navigation
        toggleMobileNav: () => {
          set((state) => {
            state.showMobileNav = !state.showMobileNav;
          });
        },

        setMobileNav: (show: boolean) => {
          set((state) => {
            state.showMobileNav = show;
          });
        },

        // Loading Actions
        setGlobalLoading: (loading: boolean) => {
          set((state) => {
            state.globalLoading = loading;
          });
        },

        setLoading: (key: string, loading: boolean) => {
          set((state) => {
            if (loading) {
              state.loadingStates[key] = true;
            } else {
              delete state.loadingStates[key];
            }
          });
        },

        clearLoading: (key: string) => {
          set((state) => {
            delete state.loadingStates[key];
          });
        },

        // Toast Actions
        addToast: (toast: Omit<ToastMessage, 'id'>) => {
          const id = generateId();
          const duration = toast.duration || 5000;

          set((state) => {
            state.toasts.push({
              ...toast,
              id,
              duration,
            });
          });

          // Auto-remove toast after duration
          if (duration > 0) {
            setTimeout(() => {
              get().removeToast(id);
            }, duration);
          }

          return id;
        },

        removeToast: (id: string) => {
          set((state) => {
            state.toasts = state.toasts.filter(toast => toast.id !== id);
          });
        },

        clearToasts: () => {
          set((state) => {
            state.toasts = [];
          });
        },

        // Modal Actions
        openModal: (modal: Omit<Modal, 'id'>) => {
          const id = generateId();

          set((state) => {
            state.modals.push({
              ...modal,
              id,
              size: modal.size || 'md',
              closable: modal.closable !== false,
            });
          });

          return id;
        },

        closeModal: (id: string) => {
          const modal = get().getModal(id);
          if (modal?.onClose) {
            modal.onClose();
          }

          set((state) => {
            state.modals = state.modals.filter(m => m.id !== id);
          });
        },

        closeAllModals: () => {
          const { modals } = get();
          modals.forEach(modal => {
            if (modal.onClose) {
              modal.onClose();
            }
          });

          set((state) => {
            state.modals = [];
          });
        },

        updateModal: (id: string, updates: Partial<Modal>) => {
          set((state) => {
            const modalIndex = state.modals.findIndex(m => m.id === id);
            if (modalIndex !== -1) {
              Object.assign(state.modals[modalIndex], updates);
            }
          });
        },

        // Context Menu Actions
        openContextMenu: (menu: Omit<ContextMenu, 'id'>) => {
          const id = generateId();

          // Close all existing context menus first
          get().closeAllContextMenus();

          set((state) => {
            state.contextMenus.push({
              ...menu,
              id,
            });
          });

          // Auto-close on click outside
          setTimeout(() => {
            const handleClickOutside = (event: MouseEvent) => {
              const target = event.target as Element;
              if (!target.closest('[data-context-menu]')) {
                get().closeContextMenu(id);
                document.removeEventListener('click', handleClickOutside);
              }
            };
            document.addEventListener('click', handleClickOutside);
          }, 0);

          return id;
        },

        closeContextMenu: (id: string) => {
          set((state) => {
            state.contextMenus = state.contextMenus.filter(m => m.id !== id);
          });
        },

        closeAllContextMenus: () => {
          set((state) => {
            state.contextMenus = [];
          });
        },

        // Command Palette Actions
        toggleCommandPalette: () => {
          set((state) => {
            state.commandPaletteOpen = !state.commandPaletteOpen;
          });
        },

        setCommandPaletteOpen: (open: boolean) => {
          set((state) => {
            state.commandPaletteOpen = open;
          });
        },

        // Keyboard Shortcuts Actions
        toggleKeyboardShortcuts: () => {
          set((state) => {
            state.keyboardShortcutsEnabled = !state.keyboardShortcutsEnabled;
          });
        },

        setKeyboardShortcutsEnabled: (enabled: boolean) => {
          set((state) => {
            state.keyboardShortcutsEnabled = enabled;
          });
        },

        toggleShortcutHelp: () => {
          set((state) => {
            state.shortcutHelpOpen = !state.shortcutHelpOpen;
          });
        },

        // Drag and Drop Actions
        setDragging: (dragging: boolean, data?: any) => {
          set((state) => {
            state.isDragging = dragging;
            state.dragData = dragging ? data : null;
          });
        },

        // Selection Actions
        setMultiSelectMode: (enabled: boolean) => {
          set((state) => {
            state.multiSelectMode = enabled;
          });
        },

        // View Options Actions
        toggleHiddenFiles: () => {
          set((state) => {
            state.showHiddenFiles = !state.showHiddenFiles;
          });
        },

        setShowHiddenFiles: (show: boolean) => {
          set((state) => {
            state.showHiddenFiles = show;
          });
        },

        toggleCompactMode: () => {
          set((state) => {
            state.compactMode = !state.compactMode;
          });
        },

        setCompactMode: (compact: boolean) => {
          set((state) => {
            state.compactMode = compact;
          });
        },

        // Panel Actions
        toggleRightPanel: () => {
          set((state) => {
            state.rightPanelOpen = !state.rightPanelOpen;
          });
        },

        setRightPanelOpen: (open: boolean) => {
          set((state) => {
            state.rightPanelOpen = open;
            if (!open) {
              state.rightPanelContent = null;
            }
          });
        },

        setRightPanelContent: (content: 'details' | 'activity' | 'sharing' | null) => {
          set((state) => {
            state.rightPanelContent = content;
            state.rightPanelOpen = content !== null;
          });
        },

        // Utility Functions
        getModal: (id: string) => {
          return get().modals.find(m => m.id === id);
        },

        getToast: (id: string) => {
          return get().toasts.find(t => t.id === id);
        },

        isLoading: (key: string) => {
          return get().loadingStates[key] || false;
        },
      })),
      {
        name: 'ui-store',
        partialize: (state) => ({
          theme: state.theme,
          sidebar: {
            isCollapsed: state.sidebar.isCollapsed,
            width: state.sidebar.width,
          },
          keyboardShortcutsEnabled: state.keyboardShortcutsEnabled,
          showHiddenFiles: state.showHiddenFiles,
          compactMode: state.compactMode,
        }),
        version: 1,
        migrate: (persistedState: any, version: number) => {
          // Handle migrations if needed
          return persistedState;
        },
      }
    ),
    { name: 'ui-store' }
  )
);

// Initialize theme on store creation
if (typeof window !== 'undefined') {
  // Listen for system theme changes
  const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
  const handleSystemThemeChange = (e: MediaQueryListEvent) => {
    const { theme } = useUIStore.getState();
    if (theme === 'system') {
      useUIStore.getState().setTheme('system');
    }
  };
  
  mediaQuery.addEventListener('change', handleSystemThemeChange);
  
  // Set initial theme
  const { theme } = useUIStore.getState();
  useUIStore.getState().setTheme(theme);
}

// Selectors
export const useTheme = () => useUIStore((state) => ({
  theme: state.theme,
  actualTheme: state.actualTheme,
}));

export const useSidebar = () => useUIStore((state) => state.sidebar);
export const useMobileNav = () => useUIStore((state) => state.showMobileNav);
export const useGlobalLoading = () => useUIStore((state) => state.globalLoading);
export const useToasts = () => useUIStore((state) => state.toasts);
export const useModals = () => useUIStore((state) => state.modals);
export const useContextMenus = () => useUIStore((state) => state.contextMenus);
export const useCommandPalette = () => useUIStore((state) => state.commandPaletteOpen);
export const useKeyboardShortcuts = () => useUIStore((state) => state.keyboardShortcutsEnabled);
export const useDragState = () => useUIStore((state) => ({
  isDragging: state.isDragging,
  dragData: state.dragData,
}));
export const useViewOptions = () => useUIStore((state) => ({
  showHiddenFiles: state.showHiddenFiles,
  compactMode: state.compactMode,
  multiSelectMode: state.multiSelectMode,
}));
export const useRightPanel = () => useUIStore((state) => ({
  isOpen: state.rightPanelOpen,
  content: state.rightPanelContent,
}));

// Actions
export const useUIActions = () => {
  const store = useUIStore();
  return {
    setTheme: store.setTheme,
    toggleTheme: store.toggleTheme,
    toggleSidebar: store.toggleSidebar,
    setSidebarCollapsed: store.setSidebarCollapsed,
    addToast: store.addToast,
    removeToast: store.removeToast,
    openModal: store.openModal,
    closeModal: store.closeModal,
    openContextMenu: store.openContextMenu,
    closeContextMenu: store.closeContextMenu,
    toggleCommandPalette: store.toggleCommandPalette,
    setDragging: store.setDragging,
    setMultiSelectMode: store.setMultiSelectMode,
    toggleHiddenFiles: store.toggleHiddenFiles,
    toggleCompactMode: store.toggleCompactMode,
    toggleRightPanel: store.toggleRightPanel,
    setRightPanelContent: store.setRightPanelContent,
    setLoading: store.setLoading,
    isLoading: store.isLoading,
  };
};

// Toast helpers
export const useToastActions = () => {
  const addToast = useUIStore((state) => state.addToast);
  
  return {
    toast: {
      success: (title: string, description?: string) => 
        addToast({ type: 'success', title, description }),
      error: (title: string, description?: string) => 
        addToast({ type: 'error', title, description }),
      warning: (title: string, description?: string) => 
        addToast({ type: 'warning', title, description }),
      info: (title: string, description?: string) => 
        addToast({ type: 'info', title, description }),
    },
  };
};