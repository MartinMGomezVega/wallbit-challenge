import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { WallbitLogo } from '../WallbitLogo';

describe('WallbitLogo', () => {
  it('renderiza variant white por defecto como img oficial', () => {
    render(<WallbitLogo />);
    const img = screen.getByRole('img', { name: 'Wallbit' });
    expect(img).toHaveAttribute('src', expect.stringContaining('wallbit-logo-white.svg'));
  });

  it('renderiza variant dark e isotype', () => {
    const { rerender } = render(<WallbitLogo variant="dark" />);
    expect(screen.getByRole('img', { name: 'Wallbit' })).toHaveAttribute(
      'src',
      expect.stringContaining('wallbit-logo-dark.svg'),
    );

    rerender(<WallbitLogo variant="isotype" />);
    expect(screen.getByRole('img', { name: 'Wallbit' })).toHaveAttribute(
      'src',
      expect.stringContaining('wallbit-isotype.svg'),
    );
  });
});
