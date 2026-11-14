import { useEffect, useState } from 'react';

export const SEVERITY_ICON_COLORS = {
  RED: '#DC2626',
  YELLOW: '#D97706',
  GREEN: '#059669',
} as const;

export const PRIMARY_COLOR = '#006FC0';
export const WARNING_BORDER_COLOR = '#D97706';

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

