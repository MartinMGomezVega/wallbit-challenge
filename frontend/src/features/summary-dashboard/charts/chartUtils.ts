import { useEffect, useState } from 'react';

export const SEVERITY_FILL_COLORS = {
  RED: 'var(--color-severity-red-icon)',
  YELLOW: 'var(--color-severity-yellow-fill)',
  GREEN: 'var(--color-severity-green-icon)',
} as const;

export const SEVERITY_ICON_COLORS = {
  RED: 'var(--color-severity-red-icon)',
  YELLOW: 'var(--color-severity-yellow-icon)',
  GREEN: 'var(--color-severity-green-icon)',
} as const;

export const PRIMARY_COLOR = 'var(--color-primary)';
export const WARNING_BORDER_COLOR = 'var(--color-severity-yellow-fill)';

export function usePrefersReducedMotion(): boolean {
  const [reduced, setReduced] = useState(() => {
    if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') {
      return true;
    }
    return window.matchMedia('(prefers-reduced-motion: reduce)').matches;
  });

  useEffect(() => {
    if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') {
      return;
    }
    const mql = window.matchMedia('(prefers-reduced-motion: reduce)');
    const onChange = (e: MediaQueryListEvent) => setReduced(e.matches);
    mql.addEventListener?.('change', onChange);
    return () => mql.removeEventListener?.('change', onChange);
  }, []);

  return reduced;
}

export const TOOLTIP_CONTENT_STYLE = {
  backgroundColor: '#171717',
  border: 'none',
  borderRadius: '6px',
  color: '#FFFFFF',
  fontSize: '12px',
  padding: '8px 10px',
} as const;

