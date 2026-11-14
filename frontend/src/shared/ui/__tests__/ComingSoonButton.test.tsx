import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { ComingSoonButton } from '../ComingSoonButton';

describe('ComingSoonButton', () => {
  it('renders children with aria-disabled and focusable button', () => {
    render(<ComingSoonButton>Emitir factura en ARCA</ComingSoonButton>);

    const button = screen.getByRole('button', { name: /emitir factura en arca/i });
    expect(button).toBeInTheDocument();
    expect(button).toHaveAttribute('aria-disabled', 'true');
    // Button must NOT be disabled attribute so it can be focused and trigger tooltips
    expect(button).not.toBeDisabled();
  });

  it('renders tooltip on hover', async () => {
    const user = userEvent.setup();
    render(<ComingSoonButton>Pagar en ARCA</ComingSoonButton>);

    const button = screen.getByRole('button', { name: /pagar en arca/i });
    await user.hover(button);

    await waitFor(() => {
      expect(screen.getByText('Próximamente')).toBeInTheDocument();
    });

    await user.unhover(button);
  });

  it('renders tooltip on keyboard focus', async () => {
    const user = userEvent.setup();
    render(<ComingSoonButton tooltipContent="Disponible en v2">Pagar en ARCA</ComingSoonButton>);

    const button = screen.getByRole('button', { name: /pagar en arca/i });
    await user.tab();
    expect(button).toHaveFocus();

    await waitFor(() => {
      expect(screen.getByText('Disponible en v2')).toBeInTheDocument();
    });
  });

  it('is inert on click and keyboard activation', async () => {
    const user = userEvent.setup();
    const handleClick = vi.fn();
    render(
      <ComingSoonButton onClick={handleClick}>
        Emitir factura
      </ComingSoonButton>
    );

    const button = screen.getByRole('button', { name: /emitir factura/i });

    // Click
    await user.click(button);
    expect(handleClick).not.toHaveBeenCalled();

    // Enter
    button.focus();
    await user.keyboard('{Enter}');
    expect(handleClick).not.toHaveBeenCalled();

    // Space
    await user.keyboard(' ');
    expect(handleClick).not.toHaveBeenCalled();
  });

  it('renders sm and md sizes properly', () => {
    const { rerender } = render(
      <ComingSoonButton size="sm">Acción</ComingSoonButton>
    );
    expect(screen.getByRole('button')).toHaveClass('h-8');

    rerender(<ComingSoonButton size="md">Acción</ComingSoonButton>);
    expect(screen.getByRole('button')).toHaveClass('h-10');
  });
});

