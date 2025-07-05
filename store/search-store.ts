import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { immer } from 'zustand/middleware/immer';
import type { 
  File,
  FileSearchParams,
  ObjectId 
} from '@/types';

interface SearchHistory {
  id: string;
  query: string;
  filters: Partial<FileSearchParams>;
  timestamp: Date;
  resultCount: number;
}

interface SavedSearch {
  id: string;
  name: string;
  query: string;
  filters: Partial<FileSearchParams>;
  createdAt: Date;
  updatedAt: Date;
}

interface SearchState {
  // Search Results
  results: File[];
  isSearching: boolean;
  hasSearched: boolean;
  
  // Current Search
  query: string;
  filters: Partial<FileSearchParams>;
  
  // Search History
  history: SearchHistory[];
  savedSearches: SavedSearch[];
  
  // Suggestions
  suggestions: string[];
  showSuggestions: boolean;
  
  // Error
  error: string | null;
  
  // Actions
  search: (query: string, filters?: Partial<FileSearchParams>) => Promise<void>;
  clearSearch: () => void;
  setQuery: (query: string) => void;
  setFilters: (filters: Partial<FileSearchParams>) => void;
  loadSuggestions: (query: string) => Promise<void>;
  clearSuggestions: () => void;
  saveSearch: (name: string) => Promise<void>;
  deleteSavedSearch: (id: string) => Promise<void>;
  loadSavedSearches: () => Promise<void>;
  clearHistory: () => void;
  clearError: () => void;
}

export const useSearchStore = create<SearchState>()(
  devtools(
    immer((set, get) => ({
      // Initial State
      results: [],
      isSearching: false,
      hasSearched: false,
      query: '',
      filters: {},
      history: [],
      savedSearches: [],
      suggestions: [],
      showSuggestions: false,
      error: null,

      // Search
      search: async (query: string, filters = {}) => {
        set((state) => {
          state.isSearching = true;
          state.error = null;
          state.query = query;
          state.filters = { ...state.filters, ...filters };
        });

        try {
          const searchParams = new URLSearchParams({
            query,
            ...Object.fromEntries(
              Object.entries({ ...get().filters, ...filters })
                .filter(([_, value]) => value !== undefined && value !== '')
                .map(([key, value]) => [key, String(value)])
            ),
          });

          const response = await fetch(`/api/client/search?${searchParams}`);
          const result = await response.json();

          if (!response.ok) {
            throw new Error(result.error || 'Search failed');
          }

          const searchResults: File[] = result.data;

          set((state) => {
            state.results = searchResults;
            state.isSearching = false;
            state.hasSearched = true;
            state.showSuggestions = false;
          });

          // Add to search history
          const historyItem: SearchHistory = {
            id: Date.now().toString(),
            query,
            filters: { ...get().filters, ...filters },
            timestamp: new Date(),
            resultCount: searchResults.length,
          };

          set((state) => {
            state.history.unshift(historyItem);
            // Keep only last 20 searches
            if (state.history.length > 20) {
              state.history = state.history.slice(0, 20);
            }
          });

        } catch (error) {
          set((state) => {
            state.error = error instanceof Error ? error.message : 'Search failed';
            state.isSearching = false;
          });
        }
      },

      // Clear Search
      clearSearch: () => {
        set((state) => {
          state.results = [];
          state.query = '';
          state.filters = {};
          state.hasSearched = false;
          state.showSuggestions = false;
          state.error = null;
        });
      },

      // Set Query
      setQuery: (query: string) => {
        set((state) => {
          state.query = query;
        });

        // Load suggestions if query is not empty
        if (query.trim() && query.length >= 2) {
          get().loadSuggestions(query);
        } else {
          get().clearSuggestions();
        }
      },

      // Set Filters
      setFilters: (filters: Partial<FileSearchParams>) => {
        set((state) => {
          state.filters = { ...state.filters, ...filters };
        });
      },

      // Load Suggestions
      loadSuggestions: async (query: string) => {
        try {
          const response = await fetch(`/api/client/search/suggestions?query=${encodeURIComponent(query)}`);
          const result = await response.json();

          if (response.ok) {
            set((state) => {
              state.suggestions = result.data;
              state.showSuggestions = true;
            });
          }
        } catch (error) {
          // Ignore suggestion errors
        }
      },

      // Clear Suggestions
      clearSuggestions: () => {
        set((state) => {
          state.suggestions = [];
          state.showSuggestions = false;
        });
      },

      // Save Search
      saveSearch: async (name: string) => {
        const { query, filters } = get();

        try {
          const response = await fetch('/api/client/search/saved', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name, query, filters }),
          });

          const result = await response.json();

          if (!response.ok) {
            throw new Error(result.error || 'Failed to save search');
          }

          const savedSearch: SavedSearch = result.data;

          set((state) => {
            state.savedSearches.unshift(savedSearch);
          });

        } catch (error) {
          set((state) => {
            state.error = error instanceof Error ? error.message : 'Failed to save search';
          });
          throw error;
        }
      },

      // Delete Saved Search
      deleteSavedSearch: async (id: string) => {
        try {
          const response = await fetch(`/api/client/search/saved/${id}`, {
            method: 'DELETE',
          });

          if (!response.ok) {
            const result = await response.json();
            throw new Error(result.error || 'Failed to delete saved search');
          }

          set((state) => {
            state.savedSearches = state.savedSearches.filter(s => s.id !== id);
          });

        } catch (error) {
          set((state) => {
            state.error = error instanceof Error ? error.message : 'Failed to delete saved search';
          });
          throw error;
        }
      },

      // Load Saved Searches
      loadSavedSearches: async () => {
        try {
          const response = await fetch('/api/client/search/saved');
          const result = await response.json();

          if (!response.ok) {
            throw new Error(result.error || 'Failed to load saved searches');
          }

          set((state) => {
            state.savedSearches = result.data;
          });

        } catch (error) {
          console.warn('Failed to load saved searches:', error);
        }
      },

      // Clear History
      clearHistory: () => {
        set((state) => {
          state.history = [];
        });
      },

      // Clear Error
      clearError: () => {
        set((state) => {
          state.error = null;
        });
      },
    })),
    { name: 'search-store' }
  )
);