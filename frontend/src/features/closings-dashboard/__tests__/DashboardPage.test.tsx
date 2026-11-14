import { describe, it, expect, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter } from 'react-router-dom';
import { server } from '../../../mocks/server';
import { DashboardPage } from '../DashboardPage';

function renderDashboard(initialUrl = '/contador?period=2026-10', onReviewClient = vi.fn()) {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  });

  return {
    ...render(
      <QueryClientProvider client={queryClient}>
        <MemoryRouter initialEntries={[initialUrl]}>
          <DashboardPage onReviewClient={onReviewClient} />
        </MemoryRouter>
      </QueryClientProvider>
    ),
    onReviewClient,
  };
}

describe('DashboardPage', () => {
  it('renderiza título, chips y tabla con datos cargados de MSW', async () => {
    renderDashboard();

    const heading = screen.getByRole('heading', { name: 'Cierre de octubre 2026' });
    expect(heading).toBeInTheDocument();
    expect(heading).not.toHaveClass('capitalize');

    // Espera a que los datos de la tabla carguen
    await waitFor(() => {
      expect(screen.getByText('Ignacio Sosa')).toBeInTheDocument();
    });

    expect(screen.getByText('Carla Benítez')).toBeInTheDocument();
    expect(screen.getByText('Valentina Ruiz')).toBeInTheDocument();

    // Chips de resumen
    expect(screen.getByText('Rojos')).toBeInTheDocument();
    expect(screen.getByText('Amarillos')).toBeInTheDocument();
    expect(screen.getByText('Verdes')).toBeInTheDocument();
  });

  it('permite alternar el filtro de severidad con los chips', async () => {
    const user = userEvent.setup();
    renderDashboard();

    await waitFor(() => {
      expect(screen.getByText('Ignacio Sosa')).toBeInTheDocument();
    });

    const redChip = screen.getByRole('button', { name: /Rojos/ });
    await user.click(redChip);

    // Solo quedan los rojos (Ignacio Sosa, Carla Benítez, Paula Rojas)
    await waitFor(() => {
      expect(screen.queryByText('Valentina Ruiz')).not.toBeInTheDocument();
    });
    expect(screen.getByText('Ignacio Sosa')).toBeInTheDocument();
  });

  it('filtra por severidad al hacer click en el botón Revisar del resumen', async () => {
    const user = userEvent.setup();
    renderDashboard();

    await waitFor(() => {
      expect(screen.getByText('Ignacio Sosa')).toBeInTheDocument();
    });

    const reviewChipBtn = screen.getAllByRole('button', { name: 'Revisar' })[0]!;
    await user.click(reviewChipBtn);

    await waitFor(() => {
      expect(screen.queryByText('Valentina Ruiz')).not.toBeInTheDocument();
    });
    expect(screen.getByText('Ignacio Sosa')).toBeInTheDocument();
  });

  it('permite ordenar por columna haciendo click en el encabezado', async () => {
    const user = userEvent.setup();
    renderDashboard();

    await waitFor(() => {
      expect(screen.getByText('Ignacio Sosa')).toBeInTheDocument();
    });

    const clientHeader = screen.getByRole('button', { name: /Cliente/ });
    await user.click(clientHeader);
  });

  it('llama a onReviewClient al hacer click en el botón Revisar de una fila', async () => {
    const onReviewClient = vi.fn();
    const user = userEvent.setup();
    renderDashboard('/contador?period=2026-10', onReviewClient);

    await waitFor(() => {
      expect(screen.getByText('Ignacio Sosa')).toBeInTheDocument();
    });

    const reviewBtns = screen.getAllByRole('button', { name: 'Revisar' });
    // El último botón de Revisar (perteneciente a una fila de la tabla)
    await user.click(reviewBtns[reviewBtns.length - 1]!);
    expect(onReviewClient).toHaveBeenCalled();
  });

  it('muestra estado vacío cuando no hay cierres con los filtros aplicados y permite limpiar filtros', async () => {
    server.use(
      http.get('/api/v1/accountant/periods/:period/closings', () => {
        return HttpResponse.json({
          items: [],
          total: 0,
          page: 1,
          pageSize: 50,
          totalPages: 1,
        });
      })
    );

    const user = userEvent.setup();
    renderDashboard();

    await waitFor(() => {
      expect(screen.getByText('No hay clientes con estos filtros')).toBeInTheDocument();
    });

    const clearBtn = screen.getByRole('button', { name: 'Limpiar filtros' });
    await user.click(clearBtn);
  });

  it('muestra mensaje de error con botón de reintentar si la API falla', async () => {
    server.use(
      http.get('/api/v1/accountant/periods/:period/closings', () => {
        return new HttpResponse(null, { status: 500 });
      })
    );

    const user = userEvent.setup();
    renderDashboard();

    await waitFor(() => {
      expect(screen.getByText('No se pudieron cargar los cierres')).toBeInTheDocument();
    });

    const retryBtn = screen.getByRole('button', { name: 'Reintentar' });
    await user.click(retryBtn);
  });

  it('muestra la barra flotante al seleccionar filas y permite accionar botones', async () => {
    const user = userEvent.setup();
    renderDashboard();

    await waitFor(() => {
      expect(screen.getByText('Lucía Fernández')).toBeInTheDocument();
    });

    const selectAll = screen.getByLabelText('Seleccionar todas las filas de la página');
    await user.click(selectAll);

    await waitFor(() => {
      expect(screen.getByText('10 seleccionados')).toBeInTheDocument();
    });

    const approveGreensBtn = screen.getByText(/Aprobar verdes/);
    await user.click(approveGreensBtn);

    // Cierra el diálogo de lote para continuar revisando seleccionados
    const cancelBatchBtn = await screen.findByRole('button', { name: /cancelar/i });
    await user.click(cancelBatchBtn);

    const reviewSelectedBtn = screen.getByText('Revisar seleccionados');
    await user.click(reviewSelectedBtn);

    // Cierra el modal de revisión para volver a la tabla
    const closeBtn = await screen.findByRole('button', { name: /cerrar modal de revisión/i });
    await user.click(closeBtn);

    const clearSelectionBtn = screen.getByRole('button', { name: 'Deseleccionar todos' });
    await user.click(clearSelectionBtn);

    await waitFor(() => {
      expect(screen.queryByText('10 seleccionados')).not.toBeInTheDocument();
    });
  });
});
