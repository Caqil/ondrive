// lib/hooks/use-local-storage.ts
import { useState, useEffect, useCallback } from 'react';

type SetValue<T> = T | ((val: T) => T);

interface UseLocalStorageOptions<T> {
  serialize?: (value: T) => string;
  deserialize?: (value: string) => T;
}

export const useLocalStorage = <T>(
  key: string,
  initialValue: T,
  options: UseLocalStorageOptions<T> = {}
): [T, (value: SetValue<T>) => void, () => void] => {
  const {
    serialize = JSON.stringify,
    deserialize = JSON.parse,
  } = options;

  // Get value from localStorage or use initial value
  const [storedValue, setStoredValue] = useState<T>(() => {
    if (typeof window === 'undefined') {
      return initialValue;
    }

    try {
      const item = window.localStorage.getItem(key);
      return item ? deserialize(item) : initialValue;
    } catch (error) {
      console.warn(`Error reading localStorage key "${key}":`, error);
      return initialValue;
    }
  });

  // Set value in localStorage and state
  const setValue = useCallback((value: SetValue<T>) => {
    try {
      const valueToStore = value instanceof Function ? value(storedValue) : value;
      setStoredValue(valueToStore);

      if (typeof window !== 'undefined') {
        window.localStorage.setItem(key, serialize(valueToStore));
      }
    } catch (error) {
      console.warn(`Error setting localStorage key "${key}":`, error);
    }
  }, [key, serialize, storedValue]);

  // Remove value from localStorage and reset to initial value
  const removeValue = useCallback(() => {
    try {
      setStoredValue(initialValue);
      if (typeof window !== 'undefined') {
        window.localStorage.removeItem(key);
      }
    } catch (error) {
      console.warn(`Error removing localStorage key "${key}":`, error);
    }
  }, [key, initialValue]);

  // Listen for changes to localStorage from other tabs/windows
  useEffect(() => {
    if (typeof window === 'undefined') return;

    const handleStorageChange = (e: StorageEvent) => {
      if (e.key === key && e.newValue !== null) {
        try {
          setStoredValue(deserialize(e.newValue));
        } catch (error) {
          console.warn(`Error parsing localStorage value for key "${key}":`, error);
        }
      } else if (e.key === key && e.newValue === null) {
        setStoredValue(initialValue);
      }
    };

    window.addEventListener('storage', handleStorageChange);
    return () => window.removeEventListener('storage', handleStorageChange);
  }, [key, initialValue, deserialize]);

  return [storedValue, setValue, removeValue];
};

// Hook for session storage
export const useSessionStorage = <T>(
  key: string,
  initialValue: T,
  options: UseLocalStorageOptions<T> = {}
): [T, (value: SetValue<T>) => void, () => void] => {
  const {
    serialize = JSON.stringify,
    deserialize = JSON.parse,
  } = options;

  const [storedValue, setStoredValue] = useState<T>(() => {
    if (typeof window === 'undefined') {
      return initialValue;
    }

    try {
      const item = window.sessionStorage.getItem(key);
      return item ? deserialize(item) : initialValue;
    } catch (error) {
      console.warn(`Error reading sessionStorage key "${key}":`, error);
      return initialValue;
    }
  });

  const setValue = useCallback((value: SetValue<T>) => {
    try {
      const valueToStore = value instanceof Function ? value(storedValue) : value;
      setStoredValue(valueToStore);

      if (typeof window !== 'undefined') {
        window.sessionStorage.setItem(key, serialize(valueToStore));
      }
    } catch (error) {
      console.warn(`Error setting sessionStorage key "${key}":`, error);
    }
  }, [key, serialize, storedValue]);

  const removeValue = useCallback(() => {
    try {
      setStoredValue(initialValue);
      if (typeof window !== 'undefined') {
        window.sessionStorage.removeItem(key);
      }
    } catch (error) {
      console.warn(`Error removing sessionStorage key "${key}":`, error);
    }
  }, [key, initialValue]);

  return [storedValue, setValue, removeValue];
};

// Hook for checking if localStorage is available
export const useLocalStorageAvailable = (): boolean => {
  const [isAvailable, setIsAvailable] = useState(false);

  useEffect(() => {
    try {
      if (typeof window !== 'undefined' && window.localStorage) {
        const testKey = '__localStorage_test__';
        window.localStorage.setItem(testKey, 'test');
        window.localStorage.removeItem(testKey);
        setIsAvailable(true);
      }
    } catch (error) {
      setIsAvailable(false);
    }
  }, []);

  return isAvailable;
};

// Hook for managing multiple localStorage keys with a prefix
export const usePrefixedStorage = (prefix: string) => {
  const getItem = useCallback(<T>(key: string, defaultValue: T): T => {
    if (typeof window === 'undefined') return defaultValue;

    try {
      const item = window.localStorage.getItem(`${prefix}_${key}`);
      return item ? JSON.parse(item) : defaultValue;
    } catch (error) {
      console.warn(`Error reading localStorage key "${prefix}_${key}":`, error);
      return defaultValue;
    }
  }, [prefix]);

  const setItem = useCallback(<T>(key: string, value: T): void => {
    if (typeof window === 'undefined') return;

    try {
      window.localStorage.setItem(`${prefix}_${key}`, JSON.stringify(value));
    } catch (error) {
      console.warn(`Error setting localStorage key "${prefix}_${key}":`, error);
    }
  }, [prefix]);

  const removeItem = useCallback((key: string): void => {
    if (typeof window === 'undefined') return;

    try {
      window.localStorage.removeItem(`${prefix}_${key}`);
    } catch (error) {
      console.warn(`Error removing localStorage key "${prefix}_${key}":`, error);
    }
  }, [prefix]);

  const clear = useCallback((): void => {
    if (typeof window === 'undefined') return;

    try {
      const keys = Object.keys(window.localStorage);
      keys.forEach(key => {
        if (key.startsWith(`${prefix}_`)) {
          window.localStorage.removeItem(key);
        }
      });
    } catch (error) {
      console.warn(`Error clearing localStorage keys with prefix "${prefix}":`, error);
    }
  }, [prefix]);

  return { getItem, setItem, removeItem, clear };
};