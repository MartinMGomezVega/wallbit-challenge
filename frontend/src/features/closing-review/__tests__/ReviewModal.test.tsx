import { describe, it, expect } from 'vitest';
import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ReviewQueueProvider, useReviewQueue } from '../useReviewQueue';
import { ReviewModal } from '../ReviewModal';
import { server } from '../../../mocks/server';
import { http, HttpResponse } from 'msw';

function TestReviewWrapper({
  initialEntries = ['/accountant?period=2026-10'],
  children,
}: {
  initialEntries?: string[];
  children: React.ReactNode;
}) {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  });

  return (
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={initialEntries}>
        <ReviewQueueProvider pageSize={50}>{children}</ReviewQueueProvider>
      </MemoryRouter>
    </QueryClientProvider>
  );
}

function OpenModalTrigger({
  startId = 'cls_01',
  mode = 'row',
  ids = ['cls_01', 'cls_02'],
}: {
  startId?: string;
  mode?: 'row' | 'chip' | 'selection';
  ids?: string[];
}) {
  const { openQueue } = useReviewQueue();
  return (
    <button
      type="button"
      onClick={() =>
        openQueue({
          mode,
          startId,
          ids,
          label: 'Filtro Test',
        })
      }
    >
      Abrir Modal
    </button>
  );
}

describe('ReviewModal', () => {
  it('abre el modal y muestra la información del cliente, semáforo y pestañas', async () => {
    const user = userEvent.setup();
    render(
      <TestReviewWrapper>
        <OpenModalTrigger />
        <ReviewModal />
      </TestReviewWrapper>
    );

    await user.click(screen.getByRole('button', { name: /abrir modal/i }));

    expect(await screen.findByRole('dialog')).toBeInTheDocument();
    expect(await screen.findByText('Ignacio Sosa')).toBeInTheDocument();
    expect(screen.getByText('CUIT 20-28222221-1')).toBeInTheDocument();
    expect(screen.getByText(/monotributo cat\. k/i)).toBeInTheDocument();

    // Pestañas
    expect(screen.getByRole('button', { name: /resumen/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /señales/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /movimientos/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /facturas y conciliación/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /historial/i })).toBeInTheDocument();
  });

  it('permite cambiar entre pestañas', async () => {
    const user = userEvent.setup();
    render(
      <TestReviewWrapper>
        <OpenModalTrigger />
        <ReviewModal />
      </TestReviewWrapper>
    );

    await user.click(screen.getByRole('button', { name: /abrir modal/i }));
    expect(await screen.findByText('Ignacio Sosa')).toBeInTheDocument();

    // Click en Historial
    await user.click(screen.getByRole('button', { name: /historial/i }));
    expect(await screen.findByText(/cálculo y versión/i)).toBeInTheDocument();

    // Click en Señales
    await user.click(screen.getByRole('button', { name: /señales/i }));
    expect(await screen.findByText('MONO_EXCLUSION_K_CAP_EXCEEDED')).toBeInTheDocument();
  });

  it('muestra aviso si el cliente quedó fuera del filtro', async () => {
    const user = userEvent.setup();
    render(
      <TestReviewWrapper>
        <OpenModalTrigger startId="cls_outside" ids={['cls_01', 'cls_02']} />
        <ReviewModal />
      </TestReviewWrapper>
    );

    await user.click(screen.getByRole('button', { name: /abrir modal/i }));
    expect(await screen.findByRole('dialog')).toBeInTheDocument();
    expect(await screen.findByText('Este cliente ya no cumple el filtro')).toBeInTheDocument();
  });

  it('navega con Anterior / Siguiente y actualiza el encabezado', async () => {
    const user = userEvent.setup();
    render(
      <TestReviewWrapper>
        <OpenModalTrigger startId="cls_01" ids={['cls_01', 'cls_02']} />
        <ReviewModal />
      </TestReviewWrapper>
    );

    await user.click(screen.getByRole('button', { name: /abrir modal/i }));
    expect(await screen.findByText('Cliente 1 de 2')).toBeInTheDocument();

    const nextBtn = screen.getByRole('button', { name: /^siguiente$/i });
    await user.click(nextBtn);

    expect(await screen.findByText('Cliente 2 de 2')).toBeInTheDocument();
  });

  it('permite abrir el diálogo de Confirmar pase a RI cuando hay exclusión de Monotributo y confirmarlo', async () => {
    const user = userEvent.setup();
    render(
      <TestReviewWrapper>
        <OpenModalTrigger startId="cls_01" ids={['cls_01']} />
        <ReviewModal />
      </TestReviewWrapper>
    );

    await user.click(screen.getByRole('button', { name: /abrir modal/i }));
    expect(await screen.findByText('Ignacio Sosa')).toBeInTheDocument();

    const confirmBtn = await screen.findByRole('button', { name: /confirmar pase a ri/i });
    expect(confirmBtn).toBeInTheDocument();

    await user.click(confirmBtn);
    const confirmDialog = await screen.findByRole('dialog', {
      name: /confirmar pase a responsable inscripto/i,
    });
    expect(confirmDialog).toBeInTheDocument();

    // Confirmar en el diálogo
    const submitBtn = within(confirmDialog).getByRole('button', { name: /^confirmar pase a ri$/i });
    await user.click(submitBtn);

    await waitFor(() => {
      expect(screen.queryByRole('dialog', { name: /confirmar pase a responsable inscripto/i })).not.toBeInTheDocument();
    });
  });

  it('permite hacer click en Aprobar y Aprobar y siguiente', async () => {
    const user = userEvent.setup();
    render(
      <TestReviewWrapper>
        <OpenModalTrigger startId="cls_01" ids={['cls_01', 'cls_02']} />
        <ReviewModal />
      </TestReviewWrapper>
    );

    await user.click(screen.getByRole('button', { name: /abrir modal/i }));
    expect(await screen.findByText('Ignacio Sosa')).toBeInTheDocument();

    const aprobarBtn = screen.getByRole('button', { name: /^aprobar$/i });
    await user.click(aprobarBtn);

    const aprobarYNextBtn = screen.getByRole('button', { name: /aprobar y siguiente/i });
    await user.click(aprobarYNextBtn);

    expect(await screen.findByText('Cliente 2 de 2')).toBeInTheDocument();
  });

  it('cierra el modal al presionar Escape o el botón de cerrar', async () => {
    const user = userEvent.setup();
    render(
      <TestReviewWrapper>
        <OpenModalTrigger />
        <ReviewModal />
      </TestReviewWrapper>
    );

    await user.click(screen.getByRole('button', { name: /abrir modal/i }));
    expect(await screen.findByRole('dialog')).toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: /cerrar modal de revisión/i }));
    await waitFor(() => {
      expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    });
  });

  it('muestra banner de conflicto 409 VERSION_CONFLICT y botón Ver cambios', async () => {
    server.use(
      http.post('/api/v1/accountant/closings/:id/approve', () => {
        return HttpResponse.json(
          { error: { code: 'VERSION_CONFLICT', message: 'El cierre cambió mientras lo revisabas' } },
          { status: 409 }
        );
      })
    );

    const user = userEvent.setup();
    render(
      <TestReviewWrapper>
        <OpenModalTrigger />
        <ReviewModal />
      </TestReviewWrapper>
    );

    await user.click(screen.getByRole('button', { name: /abrir modal/i }));
    expect(await screen.findByText('Ignacio Sosa')).toBeInTheDocument();

    const aprobarBtn = screen.getByRole('button', { name: /^aprobar$/i });
    await user.click(aprobarBtn);

    expect(await screen.findByText(/el cierre cambió mientras lo revisabas/i)).toBeInTheDocument();
    const verCambiosBtn = screen.getByRole('button', { name: /ver cambios/i });
    expect(verCambiosBtn).toBeInTheDocument();

    await user.click(verCambiosBtn);
    await waitFor(() => {
      expect(screen.queryByText(/el cierre cambió mientras lo revisabas/i)).not.toBeInTheDocument();
    });
  });

  it('soporta atajo de teclado A para aprobar y avanzar', async () => {
    const user = userEvent.setup();
    render(
      <TestReviewWrapper>
        <OpenModalTrigger startId="cls_01" ids={['cls_01', 'cls_02']} />
        <ReviewModal />
      </TestReviewWrapper>
    );

    await user.click(screen.getByRole('button', { name: /abrir modal/i }));
    expect(await screen.findByText('Ignacio Sosa')).toBeInTheDocument();

    await user.keyboard('a');

    expect(await screen.findByText('Cliente 2 de 2')).toBeInTheDocument();
  });

  it('permite ingresar nota para el cliente y verla reflejada en la pestaña de aviso', async () => {
    const user = userEvent.setup();
    render(
      <TestReviewWrapper>
        <OpenModalTrigger />
        <ReviewModal />
      </TestReviewWrapper>
    );

    await user.click(screen.getByRole('button', { name: /abrir modal/i }));
    expect(await screen.findByText('Ignacio Sosa')).toBeInTheDocument();

    const noteInput = screen.getByPlaceholderText(/nota para el cliente/i);
    await user.type(noteInput, 'Por favor revisar deducciones');

    // Cambiar a la pestaña de aviso
    await user.click(screen.getByRole('button', { name: /aviso/i }));

    expect(await screen.findByText(/por favor revisar deducciones/i)).toBeInTheDocument();
  });

  it('muestra banner de error general si falla con 500 y permite descartarlo', async () => {
    server.use(
      http.post('/api/v1/accountant/closings/:id/approve', () => {
        return HttpResponse.json(
          { error: { code: 'INTERNAL', message: 'Error interno del servidor' } },
          { status: 500 }
        );
      })
    );

    const user = userEvent.setup();
    render(
      <TestReviewWrapper>
        <OpenModalTrigger />
        <ReviewModal />
      </TestReviewWrapper>
    );

    await user.click(screen.getByRole('button', { name: /abrir modal/i }));
    expect(await screen.findByText('Ignacio Sosa')).toBeInTheDocument();

    const aprobarBtn = screen.getByRole('button', { name: /^aprobar$/i });
    await user.click(aprobarBtn);

    expect(await screen.findByText(/error interno del servidor/i)).toBeInTheDocument();
    const descartarBtn = screen.getByRole('button', { name: /descartar/i });
    await user.click(descartarBtn);

    expect(screen.queryByText(/error interno del servidor/i)).not.toBeInTheDocument();
  });

  it('cierra la cola al presionar Aprobar y siguiente en el último cliente', async () => {
    const user = userEvent.setup();
    render(
      <TestReviewWrapper>
        <OpenModalTrigger startId="cls_01" ids={['cls_01']} />
        <ReviewModal />
      </TestReviewWrapper>
    );

    await user.click(screen.getByRole('button', { name: /abrir modal/i }));
    expect(await screen.findByText('Ignacio Sosa')).toBeInTheDocument();

    const aprobarYNextBtn = screen.getByRole('button', { name: /aprobar y siguiente/i });
    await user.click(aprobarYNextBtn);

    await waitFor(() => {
      expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    });
  });
});
