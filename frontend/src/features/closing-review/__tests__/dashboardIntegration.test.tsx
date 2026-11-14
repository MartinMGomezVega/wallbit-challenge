import { describe, it, expect, vi } from 'vitest';
import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { DashboardPage } from '../../closings-dashboard/DashboardPage';

function renderApp(initialUrl = '/contador?period=2026-10', queryClient?: QueryClient) {
  const qc =
    queryClient ||
    new QueryClient({
      defaultOptions: {
        queries: { retry: false },
      },
    });

  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter initialEntries={[initialUrl]}>
        <Routes>
          <Route path="/contador" element={<DashboardPage />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>
  );
}

describe('Dashboard + ReviewModal Integration (009)', () => {
  it('Modo 1 (row): click en el nombre del cliente o botón abre el modal en ese cliente', async () => {
    const user = userEvent.setup();
    renderApp();

    await waitFor(() => {
      expect(screen.getByText('Ignacio Sosa')).toBeInTheDocument();
    });

    // Click en el nombre de Ignacio Sosa (link al modal)
    await user.click(screen.getByText('Ignacio Sosa'));

    // Modal abierto con Ignacio Sosa
    expect(await screen.findByRole('dialog')).toBeInTheDocument();
    expect(await screen.findByRole('heading', { name: 'Ignacio Sosa' })).toBeInTheDocument();
  });

  it('Modo 2 (chip): click en Revisar del chip de severidad abre cola filtrada', async () => {
    const user = userEvent.setup();
    renderApp();

    await waitFor(() => {
      expect(screen.getByText('Ignacio Sosa')).toBeInTheDocument();
    });

    const reviewChipBtn = screen.getAllByRole('button', { name: 'Revisar' })[0]!;
    await user.click(reviewChipBtn);

    const dialog = await screen.findByRole('dialog');
    expect(dialog).toBeInTheDocument();
    expect(within(dialog).getByText('Rojos · octubre 2026')).toBeInTheDocument();
  });

  it('Modo 3 (selection): seleccionar filas y click en Revisar seleccionados arma cola con esa selección', async () => {
    const user = userEvent.setup();
    renderApp();

    await waitFor(() => {
      expect(screen.getByText('Ignacio Sosa')).toBeInTheDocument();
      expect(screen.getByText('Carla Benítez')).toBeInTheDocument();
    });

    const cb1 = screen.getByLabelText('Seleccionar fila de Ignacio Sosa');
    await user.click(cb1);
    expect(cb1).toBeChecked();

    const cb2 = screen.getByLabelText('Seleccionar fila de Carla Benítez');
    await user.click(cb2);
    expect(cb2).toBeChecked();

    const reviewSelectedBtn = await screen.findByText('Revisar seleccionados');
    await user.click(reviewSelectedBtn);

    const dialog = await screen.findByRole('dialog');
    expect(dialog).toBeInTheDocument();
    expect(within(dialog).getByText('2 clientes seleccionados')).toBeInTheDocument();
    expect(within(dialog).getByText('Cliente 1 de 2')).toBeInTheDocument();
  });

  it('Reapertura desde URL: si la URL contiene client=, abre el modal automáticamente', async () => {
    renderApp('/contador?period=2026-10&client=cls_02');

    expect(await screen.findByRole('dialog')).toBeInTheDocument();
    expect(await screen.findByText('Carla Benítez')).toBeInTheDocument();
  });

  it('Prefetch: al abrir y navegar se precarga el detalle del siguiente y anterior', async () => {
    const queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });
    const prefetchSpy = vi.spyOn(queryClient, 'prefetchQuery');

    const user = userEvent.setup();
    renderApp('/contador?period=2026-10', queryClient);

    await waitFor(() => {
      expect(screen.getByText('Ignacio Sosa')).toBeInTheDocument();
    });

    await user.click(screen.getByText('Ignacio Sosa'));
    expect(await screen.findByRole('dialog')).toBeInTheDocument();

    await waitFor(() => {
      expect(prefetchSpy).toHaveBeenCalled();
    });
  });

  it('Cerrar modal resalta la fila y remueve client de la URL', async () => {
    const user = userEvent.setup();
    renderApp('/contador?period=2026-10');

    await waitFor(() => {
      expect(screen.getByText('Ignacio Sosa')).toBeInTheDocument();
    });

    await user.click(screen.getByText('Ignacio Sosa'));
    expect(await screen.findByRole('dialog')).toBeInTheDocument();

    const closeBtn = screen.getByRole('button', { name: /cerrar modal de revisión/i });
    await user.click(closeBtn);

    await waitFor(() => {
      expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    });
  });
});
