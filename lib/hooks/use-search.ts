// lib/hooks/use-search.ts
import { useSearchStore } from '@/store/search-store';

/**
 * Simple hook that re-exports search store functionality
 * Maintains consistency with existing store patterns
 */
export const useSearch = () => {
  const store = useSearchStore();

  return {
    // State
    results: store.results,
    query: store.query,
    filters: store.filters,
    isSearching: store.isSearching,
    hasSearched: store.hasSearched,
    suggestions: store.suggestions,
    showSuggestions: store.showSuggestions,
    history: store.history,
    savedSearches: store.savedSearches,
    error: store.error,

    // Actions
    search: store.search,
    setQuery: store.setQuery,
    setFilters: store.setFilters,
    clearSearch: store.clearSearch,
    saveSearch: store.saveSearch,
    loadSavedSearches: store.loadSavedSearches,
    deleteSavedSearch: store.deleteSavedSearch,
    loadSuggestions: store.loadSuggestions,
    clearError: store.clearError,
  };
};