import { describe, it, expect } from 'vitest';
import {
  SEVERITY_FILL_COLORS,
  SEVERITY_ICON_COLORS,
  WARNING_BORDER_COLOR,
} from '../charts/chartUtils';

describe('chartUtils tokens', () => {
  it('usa los tokens CSS para severidad amarilla en lugar de hex hardcodeados', () => {
    expect(SEVERITY_FILL_COLORS.YELLOW).toBe('var(--color-severity-yellow-fill)');
    expect(SEVERITY_ICON_COLORS.YELLOW).toBe('var(--color-severity-yellow-icon)');
    expect(WARNING_BORDER_COLOR).toBe('var(--color-severity-yellow-fill)');
  });
});

