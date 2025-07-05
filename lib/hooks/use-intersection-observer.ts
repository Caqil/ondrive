// lib/hooks/use-intersection-observer.ts
import { useEffect, useRef, useState, useCallback } from 'react';

interface IntersectionObserverOptions {
  root?: Element | null;
  rootMargin?: string;
  threshold?: number | number[];
  triggerOnce?: boolean;
  skip?: boolean;
}

interface IntersectionObserverResult {
  ref: (node?: Element | null) => void;
  inView: boolean;
  entry?: IntersectionObserverEntry;
}

export const useIntersectionObserver = (
  options: IntersectionObserverOptions = {}
): IntersectionObserverResult => {
  const {
    root = null,
    rootMargin = '0px',
    threshold = 0,
    triggerOnce = false,
    skip = false,
  } = options;

  const [inView, setInView] = useState(false);
  const [entry, setEntry] = useState<IntersectionObserverEntry>();
  const elementRef = useRef<Element | null>(null);
  const observerRef = useRef<IntersectionObserver | null>(null);

  const ref = useCallback((node: Element | null) => {
    if (elementRef.current) {
      observerRef.current?.unobserve(elementRef.current);
    }

    elementRef.current = node;

    if (skip || !node) return;

    if (observerRef.current) {
      observerRef.current.observe(node);
    }
  }, [skip]);

  useEffect(() => {
    if (skip || !elementRef.current) return;

    const observer = new IntersectionObserver(
      ([entry]) => {
        const isIntersecting = entry.isIntersecting;
        
        setInView(isIntersecting);
        setEntry(entry);

        if (triggerOnce && isIntersecting && observerRef.current && elementRef.current) {
          observerRef.current.unobserve(elementRef.current);
        }
      },
      {
        root,
        rootMargin,
        threshold,
      }
    );

    observerRef.current = observer;

    if (elementRef.current) {
      observer.observe(elementRef.current);
    }

    return () => {
      observer.disconnect();
      observerRef.current = null;
    };
  }, [root, rootMargin, threshold, triggerOnce, skip]);

  return { ref, inView, entry };
};

// Hook for infinite scrolling
export const useInfiniteScroll = (
  callback: () => void | Promise<void>,
  options: {
    hasNextPage?: boolean;
    isLoading?: boolean;
    rootMargin?: string;
    threshold?: number;
  } = {}
) => {
  const {
    hasNextPage = true,
    isLoading = false,
    rootMargin = '100px',
    threshold = 0.1,
  } = options;

  const { ref, inView } = useIntersectionObserver({
    rootMargin,
    threshold,
    skip: !hasNextPage || isLoading,
  });

  useEffect(() => {
    if (inView && hasNextPage && !isLoading) {
      callback();
    }
  }, [inView, hasNextPage, isLoading, callback]);

  return { ref, inView };
};

// Hook for lazy loading
export const useLazyLoad = (
  options: {
    rootMargin?: string;
    threshold?: number;
    triggerOnce?: boolean;
  } = {}
) => {
  const {
    rootMargin = '50px',
    threshold = 0.1,
    triggerOnce = true,
  } = options;

  return useIntersectionObserver({
    rootMargin,
    threshold,
    triggerOnce,
  });
};

// Hook for viewport detection
export const useViewportEntry = (
  options: {
    rootMargin?: string;
    threshold?: number;
  } = {}
) => {
  const { rootMargin = '0px', threshold = 0 } = options;
  const [entries, setEntries] = useState<IntersectionObserverEntry[]>([]);
  const refs = useRef<Map<Element, boolean>>(new Map());
  const observerRef = useRef<IntersectionObserver | null>(null);

  const observe = useCallback((element: Element) => {
    if (!observerRef.current) {
      observerRef.current = new IntersectionObserver(
        (observerEntries) => {
          setEntries(observerEntries);
        },
        { rootMargin, threshold }
      );
    }

    refs.current.set(element, true);
    observerRef.current.observe(element);
  }, [rootMargin, threshold]);

  const unobserve = useCallback((element: Element) => {
    if (observerRef.current) {
      observerRef.current.unobserve(element);
    }
    refs.current.delete(element);
  }, []);

  useEffect(() => {
    return () => {
      if (observerRef.current) {
        observerRef.current.disconnect();
      }
    };
  }, []);

  return { entries, observe, unobserve };
};