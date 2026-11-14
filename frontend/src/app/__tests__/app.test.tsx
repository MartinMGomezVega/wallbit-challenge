import { describe, it, expect } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
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
    fireEvent.change(periodSelect, { target: { value: '2026-09' } });

    expect(await screen.findByRole('heading', { name: 'Cierre de septiembre 2026' })).toBeInTheDocument();
    expect(router.state.location.search).toContain('period=2026-09');

    fireEvent.click(screen.getByRole('link', { name: 'Resumen' }));
    expect(await screen.findByRole('heading', { name: 'Resumen de septiembre 2026' })).toBeInTheDocument();
    expect(document.title).toBe('Resumen · Wallbit Contabilidad');
    expect(screen.getByRole('link', { name: 'Resumen' }).className).toContain('border-[#0D99FF]');
  });

  it('redirige rutas desconocidas (como /cliente) a /contador', async () => {
    const qc = createQueryClient();
    const router = createMemoryRouter(routes, { initialEntries: ['/cliente'] });

    render(
      <Providers client={qc}>
        <RouterProvider router={router} />
      </Providers>,
    );

    expect(await screen.findByRole('heading', { name: 'Cierre de octubre 2026' })).toBeInTheDocument();
  });
});
