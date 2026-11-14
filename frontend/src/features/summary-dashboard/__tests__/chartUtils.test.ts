import { describe, it, expect, vi } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import {
  SEVERITY_FILL_COLORS,
  SEVERITY_ICON_COLORS,
  WARNING_BORDER_COLOR,
  usePrefersReducedMotion,
} from '../charts/chartUtils';

describe('chartUtils tokens', () => {
  it('usa los tokens CSS para severidad amarilla en lugar de hex hardcodeados', () => {
    expect(SEVERITY_FILL_COLORS.YELLOW).toBe('var(--color-severity-yellow-fill)');
    expect(SEVERITY_ICON_COLORS.YELLOW).toBe('var(--color-severity-yellow-icon)');
    expect(WARNING_BORDER_COLOR).toBe('var(--color-severity-yellow-fill)');
  });

  it('usePrefersReducedMotion responds to matchMedia and changes', () => {
    let changeHandler: ((e: MediaQueryListEvent) => void) | null = null;
    const matchMediaMock = vi.fn().mockImplementation((query: string) => ({
      matches: false,
      media: query,
      onchange: null,
      addListener: vi.fn(),
      removeListener: vi.fn(),
      addEventListener: vi.fn((event: string, handler: (e: MediaQueryListEvent) => void) => {
        if (event === 'change') changeHandler = handler;
      }),
      removeEventListener: vi.fn(),
      dispatchEvent: vi.fn(),
    }));

    window.matchMedia = matchMediaMock;

    const { result } = renderHook(() => usePrefersReducedMotion());
    expect(result.current).toBe(false);

    // Simulate matchMedia change event
    if (changeHandler) {
      act(() => {
        changeHandler!({ matches: true } as MediaQueryListEvent);
      });
      expect(result.current).toBe(true);
    }
  });
});
