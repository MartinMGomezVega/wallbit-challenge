import { describe, it, expect } from 'vitest';
import { render } from '@testing-library/react';
import { WallbitLogo } from '../WallbitLogo';

describe('WallbitLogo', () => {
  it('renderiza variant white por defecto', () => {
    const { container } = render(<WallbitLogo />);
    const text = container.querySelector('text');
    expect(text).toHaveAttribute('fill', '#FFFFFF');
  });

  it('renderiza variant dark', () => {
    const { container } = render(<WallbitLogo variant="dark" />);
    const text = container.querySelector('text');
    expect(text).toHaveAttribute('fill', '#0A0A0A');
  });
});
