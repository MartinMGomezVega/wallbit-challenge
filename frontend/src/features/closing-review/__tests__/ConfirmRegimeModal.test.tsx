import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { ConfirmRegimeModal } from '../ConfirmRegimeModal';
import * as api from '../api';

describe('ConfirmRegimeModal', () => {
  it('muestra datos del cliente y fecha efectiva de exclusión', () => {
    render(
      <ConfirmRegimeModal
        open={true}
        onOpenChange={vi.fn()}
        clientId="cli_01"
        clientName="Ignacio Sosa"
        effectiveDate="2026-10-14"
      />
    );

    expect(screen.getByText('Confirmar pase a Responsable Inscripto')).toBeInTheDocument();
    expect(screen.getByText('Ignacio Sosa')).toBeInTheDocument();
    expect(screen.getByText('14/10/2026')).toBeInTheDocument();
  });

  it('llama a confirmRegimeTransition y onSuccess al confirmar', async () => {
    const user = userEvent.setup();
    const onOpenChange = vi.fn();
    const onSuccess = vi.fn();
    const spy = vi.spyOn(api, 'confirmRegimeTransition').mockResolvedValueOnce({
      client: { id: 'cli_01', regime: 'RI' },
    });

    render(
      <ConfirmRegimeModal
        open={true}
        onOpenChange={onOpenChange}
        clientId="cli_01"
        clientName="Ignacio Sosa"
        effectiveDate="2026-10-14"
        onSuccess={onSuccess}
      />
    );

    const confirmBtn = screen.getByRole('button', { name: /confirmar pase a ri/i });
    await user.click(confirmBtn);

    expect(spy).toHaveBeenCalledWith('cli_01', '2026-10-14', false);
    expect(onOpenChange).toHaveBeenCalledWith(false);
    expect(onSuccess).toHaveBeenCalled();
  });

  it('muestra mensaje de error si falla la confirmación', async () => {
    const user = userEvent.setup();
    vi.spyOn(api, 'confirmRegimeTransition').mockRejectedValueOnce(
      new Error('Error de validación en ARCA')
    );

    render(
      <ConfirmRegimeModal
        open={true}
        onOpenChange={vi.fn()}
        clientId="cli_01"
        clientName="Ignacio Sosa"
        effectiveDate="2026-10-14"
      />
    );

    const confirmBtn = screen.getByRole('button', { name: /confirmar pase a ri/i });
    await user.click(confirmBtn);

    expect(await screen.findByText('Error de validación en ARCA')).toBeInTheDocument();
  });

  it('cierra el diálogo al hacer click en Cancelar', async () => {
    const user = userEvent.setup();
    const onOpenChange = vi.fn();

    render(
      <ConfirmRegimeModal
        open={true}
        onOpenChange={onOpenChange}
        clientId="cli_01"
        clientName="Ignacio Sosa"
        effectiveDate="2026-10-14"
      />
    );

    await user.click(screen.getByRole('button', { name: /cancelar/i }));
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });
});

