import { describe, it, expect } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { RouterProvider, createMemoryRouter } from 'react-router-dom';
import { routes } from '../router';
import { Providers, createQueryClient } from '../providers';

describe('App & Navigation', () => {
  it('redirige la ruta raíz / a /contador, muestra header con identidad y actualiza <title>', async () => {
    const qc = createQueryClient();
    const router = createMemoryRouter(routes, { initialEntries: ['/'] });

    render(
      <Providers client={qc}>
        <RouterProvider router={router} />
      </Providers>,
    );

    expect(await screen.findByRole('heading', { name: 'Cierre de octubre 2026' })).toBeInTheDocument();
    expect(document.title).toBe('Cierres · Wallbit Contabilidad');

    const cierresLink = screen.getByRole('link', { name: 'Cierres' });
    const resumenLink = screen.getByRole('link', { name: 'Resumen' });
    const paramsLink = screen.getByRole('link', { name: 'Parámetros' });
    expect(cierresLink.className).toContain('border-[#0D99FF]');
    expect(resumenLink.className).toContain('border-transparent');
    expect(paramsLink.className).toContain('border-transparent');

    const identity = screen.getByTestId('accountant-identity');
    expect(identity).toHaveAttribute('title', 'Sesión de demo');
    expect(identity).toHaveTextContent('Contador');

    const header = screen.getByRole('banner');
    expect(header).not.toHaveTextContent('Contabilidad');
  });

  it('permite navegar a Parámetros y actualiza <title> y pestaña activa', async () => {
    const qc = createQueryClient();
    const router = createMemoryRouter(routes, { initialEntries: ['/contador/parametros'] });

    render(
      <Providers client={qc}>
        <RouterProvider router={router} />
      </Providers>,
    );

    expect(await screen.findByRole('heading', { name: 'Parámetros' })).toBeInTheDocument();
    expect(document.title).toBe('Parámetros · Wallbit Contabilidad');
    expect(screen.getByRole('link', { name: 'Parámetros' }).className).toContain('border-[#0D99FF]');
  });

  it('cambia el período global en el header y lo comparte con Resumen y <title>', async () => {
    const qc = createQueryClient();
    const router = createMemoryRouter(routes, { initialEntries: ['/contador'] });

    render(
      <Providers client={qc}>
        <RouterProvider router={router} />
      </Providers>,
    );

    expect(await screen.findByRole('heading', { name: 'Cierre de octubre 2026' })).toBeInTheDocument();

    const periodSelect = screen.getByLabelText('Período');
    await waitFor(() => {
      expect(screen.getByRole('option', { name: 'septiembre 2026' })).toBeInTheDocument();
    });
    fireEvent.change(periodSelect, { target: { value: '2026-09' } });

    expect(await screen.findByRole('heading', { name: 'Cierre de septiembre 2026' })).toBeInTheDocument();
    expect(router.state.location.search).toContain('period=2026-09');

    fireEvent.click(screen.getByRole('link', { name: 'Resumen' }));
    expect(await screen.findByRole('heading', { name: 'Resumen de septiembre 2026' })).toBeInTheDocument();
    expect(document.title).toBe('Resumen · Wallbit Contabilidad');
    expect(screen.getByRole('link', { name: 'Resumen' }).className).toContain('border-[#0D99FF]');
  });

  it('muestra el switch de vistas en el header con Contador activo por defecto', async () => {
    const qc = createQueryClient();
    const router = createMemoryRouter(routes, { initialEntries: ['/contador'] });

    render(
      <Providers client={qc}>
        <RouterProvider router={router} />
      </Providers>,
    );

    const switchEl = await screen.findByTestId('view-switch');
    expect(switchEl).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Contador' })).toHaveAttribute('aria-current', 'page');
    expect(screen.getByRole('link', { name: 'Cliente (demo)' })).not.toHaveAttribute('aria-current');
    expect(screen.getByLabelText('Período')).toBeInTheDocument();
  });

  it('oculta el selector de período en /cliente y activa la opción Cliente (demo)', async () => {
    const qc = createQueryClient();
    const router = createMemoryRouter(routes, { initialEntries: ['/cliente/cli_demo_15'] });

    render(
      <Providers client={qc}>
        <RouterProvider router={router} />
      </Providers>,
    );

    const switchEl = await screen.findByTestId('view-switch');
    expect(switchEl).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Cliente (demo)' })).toHaveAttribute('aria-current', 'page');
    expect(screen.queryByLabelText('Período')).not.toBeInTheDocument();
    expect(screen.queryByTestId('accountant-identity')).not.toBeInTheDocument();
  });

  it('redirige /cliente a /cliente/cli_demo_15', async () => {
    const qc = createQueryClient();
    const router = createMemoryRouter(routes, { initialEntries: ['/cliente'] });

    render(
      <Providers client={qc}>
        <RouterProvider router={router} />
      </Providers>,
    );

    await waitFor(() => {
      expect(router.state.location.pathname).toBe('/cliente/cli_demo_15');
    });
  });

  it('redirige rutas desconocidas a /contador', async () => {
    const qc = createQueryClient();
    const router = createMemoryRouter(routes, { initialEntries: ['/ruta-inexistente'] });

    render(
      <Providers client={qc}>
        <RouterProvider router={router} />
      </Providers>,
    );

    expect(await screen.findByRole('heading', { name: 'Cierre de octubre 2026' })).toBeInTheDocument();
    expect(router.state.location.pathname).toBe('/contador');
  });
});
