import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { BatchApproveDialog } from '../BatchApproveDialog';
import { server } from '../../../mocks/server';
import { http, HttpResponse } from 'msw';

function renderWithClient(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  return render(<QueryClientProvider client={queryClient}>{ui}</QueryClientProvider>);
}

describe('BatchApproveDialog', () => {
  const mockItems = [
    { id: 'clo_01', version: 1, clientName: 'Ignacio Sosa' },
    { id: 'clo_02', version: 2, clientName: 'Lucía Méndez' },
  ];

  it('muestra el conteo y mensaje de confirmación antes de aprobar', () => {
    renderWithClient(
      <BatchApproveDialog
        open={true}
        onOpenChange={vi.fn()}
        period="2026-10"
        items={mockItems}
      />
    );

    expect(screen.getByRole('dialog')).toBeInTheDocument();
    expect(screen.getByText('Aprobar clientes en lote')).toBeInTheDocument();
    expect(screen.getByText(/vas a aprobar y notificar/i)).toBeInTheDocument();
    expect(screen.getByText(/2/)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /aprobar y notificar/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /cancelar/i })).toBeInTheDocument();
  });

  it('procesa lote exitoso y muestra resultado', async () => {
    const user = userEvent.setup();
    const handleSuccess = vi.fn();
    const handleOpenChange = vi.fn();

    renderWithClient(
      <BatchApproveDialog
        open={true}
        onOpenChange={handleOpenChange}
        period="2026-10"
        items={mockItems}
        onSuccess={handleSuccess}
      />
    );

    const confirmBtn = screen.getByRole('button', { name: /aprobar y notificar/i });
    await user.click(confirmBtn);

    expect(await screen.findByText('Resultado de la aprobación')).toBeInTheDocument();
    expect(screen.getByText(/se aprobaron/i)).toBeInTheDocument();
    expect(handleSuccess).toHaveBeenCalled();

    const closeButtons = screen.getAllByRole('button', { name: /cerrar/i });
    await user.click(closeButtons[closeButtons.length - 1]!);
    expect(handleOpenChange).toHaveBeenCalledWith(false);
  });

  it('muestra ítems fallidos con descripciones claras', async () => {
    server.use(
      http.post('/api/v1/accountant/periods/:period/closings/approve-batch', () => {
        return HttpResponse.json({
          approved: 1,
          failed: 1,
          results: [
            { id: 'clo_01', ok: true },
            { id: 'clo_02', ok: false, error: 'VERSION_CONFLICT' },
          ],
        });
      })
    );

    const user = userEvent.setup();

    renderWithClient(
      <BatchApproveDialog
        open={true}
        onOpenChange={vi.fn()}
        period="2026-10"
        items={mockItems}
      />
    );

    const confirmBtn = screen.getByRole('button', { name: /aprobar y notificar/i });
    await user.click(confirmBtn);

    expect(await screen.findByText('Resultado de la aprobación')).toBeInTheDocument();
    expect(screen.getByText(/fallaron/i)).toBeInTheDocument();
    expect(screen.getByText('Lucía Méndez')).toBeInTheDocument();
    expect(screen.getByText('El cierre cambió mientras lo revisabas')).toBeInTheDocument();
  });

  it('muestra mensaje de error si falla la llamada de red', async () => {
    server.use(
      http.post('/api/v1/accountant/periods/:period/closings/approve-batch', () => {
        return HttpResponse.error();
      })
    );

    const user = userEvent.setup();

    renderWithClient(
      <BatchApproveDialog
        open={true}
        onOpenChange={vi.fn()}
        period="2026-10"
        items={mockItems}
      />
    );

    const confirmBtn = screen.getByRole('button', { name: /aprobar y notificar/i });
    await user.click(confirmBtn);

    expect(await screen.findByRole('alert')).toBeInTheDocument();
  });

  it('cierra el diálogo al hacer click en Cancelar', async () => {
    const user = userEvent.setup();
    const handleOpenChange = vi.fn();

    renderWithClient(
      <BatchApproveDialog
        open={true}
        onOpenChange={handleOpenChange}
        period="2026-10"
        items={mockItems}
      />
    );

    const cancelBtn = screen.getByRole('button', { name: /cancelar/i });
    await user.click(cancelBtn);

    expect(handleOpenChange).toHaveBeenCalledWith(false);
  });

  it('soporta un solo cliente (singular) y errores no mapeados o sin nombre', async () => {
    server.use(
      http.post('/api/v1/accountant/periods/:period/closings/approve-batch', () => {
        return HttpResponse.json({
          approved: 1,
          failed: 2,
          results: [
            { id: 'clo_sin_nombre', ok: false, error: 'UNKNOWN_CUSTOM_ERROR' },
            { id: 'clo_sin_error_prop', ok: false },
            { id: 'clo_single_ok', ok: true },
          ],
        });
      })
    );

    const user = userEvent.setup();

    renderWithClient(
      <BatchApproveDialog
        open={true}
        onOpenChange={vi.fn()}
        period="2026-10"
        items={[
          { id: 'clo_sin_nombre', version: 1 },
          { id: 'clo_sin_error_prop', version: 1, clientName: 'Sin Error Prop' },
          { id: 'clo_single_ok', version: 1, clientName: 'Single OK' },
        ]}
      />
    );

    const confirmBtn = screen.getByRole('button', { name: /aprobar y notificar/i });
    await user.click(confirmBtn);

    expect(await screen.findByText('Resultado de la aprobación')).toBeInTheDocument();
    expect(screen.getByText('clo_sin_nombre')).toBeInTheDocument();
    expect(screen.getByText('UNKNOWN_CUSTOM_ERROR')).toBeInTheDocument();
    expect(screen.getByText('Error')).toBeInTheDocument();
  });
});
