import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { ReviewHeader } from '../ReviewHeader';

describe('ReviewHeader', () => {
  it('muestra posición, total, label del filtro y contador de revisados', () => {
    render(
      <ReviewHeader
        currentIndex={2}
        total={47}
        filterLabel="Rojos · RI · 'Faltan facturas'"
        reviewedCount={12}
        canGoPrev={true}
        canGoNext={true}
        onPrev={vi.fn()}
        onNext={vi.fn()}
        onClose={vi.fn()}
      />
    );

    expect(screen.getByText('Cliente 3 de 47')).toBeInTheDocument();
    expect(screen.getByText("Rojos · RI · 'Faltan facturas'")).toBeInTheDocument();
    expect(screen.getByText('12 de 47 revisados')).toBeInTheDocument();
  });

  it('deshabilita Anterior con aria-disabled en el primer elemento', async () => {
    const onPrev = vi.fn();
    render(
      <ReviewHeader
        currentIndex={0}
        total={10}
        filterLabel="Todos"
        reviewedCount={0}
        canGoPrev={false}
        canGoNext={true}
        onPrev={onPrev}
        onNext={vi.fn()}
        onClose={vi.fn()}
      />
    );

    const prevBtn = screen.getByRole('button', { name: /anterior/i });
    expect(prevBtn).toBeDisabled();
    expect(prevBtn).toHaveAttribute('aria-disabled', 'true');
  });

  it('deshabilita Siguiente con aria-disabled en el último elemento', async () => {
    const onNext = vi.fn();
    render(
      <ReviewHeader
        currentIndex={9}
        total={10}
        filterLabel="Todos"
        reviewedCount={5}
        canGoPrev={true}
        canGoNext={false}
        onPrev={vi.fn()}
        onNext={onNext}
        onClose={vi.fn()}
      />
    );

    const nextBtn = screen.getByRole('button', { name: /siguiente/i });
    expect(nextBtn).toBeDisabled();
    expect(nextBtn).toHaveAttribute('aria-disabled', 'true');
  });

  it('llama a callbacks al hacer click en Anterior, Siguiente y Cerrar', async () => {
    const user = userEvent.setup();
    const onPrev = vi.fn();
    const onNext = vi.fn();
    const onClose = vi.fn();

    render(
      <ReviewHeader
        currentIndex={3}
        total={10}
        filterLabel="Filtro"
        reviewedCount={2}
        canGoPrev={true}
        canGoNext={true}
        onPrev={onPrev}
        onNext={onNext}
        onClose={onClose}
      />
    );

    await user.click(screen.getByRole('button', { name: /anterior/i }));
    expect(onPrev).toHaveBeenCalledTimes(1);

    await user.click(screen.getByRole('button', { name: /siguiente/i }));
    expect(onNext).toHaveBeenCalledTimes(1);

    await user.click(screen.getByRole('button', { name: /cerrar modal de revisión/i }));
    expect(onClose).toHaveBeenCalledTimes(1);
  });
});

