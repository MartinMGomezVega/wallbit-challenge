import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { RouterProvider, createMemoryRouter } from 'react-router-dom';
import { RoleSwitcher } from '../RoleSwitcher';
import { routes } from '../router';
import { Providers, createQueryClient } from '../providers';
import { setRole, resetRoleForTests, getRole } from '../roleStore';

describe('App & Navigation', () => {
  beforeEach(() => {
    resetRoleForTests();
    window.localStorage.clear();
  });

  it('renderiza RoleSwitcher con dropdown y permite cambiar a cliente', async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    const clients = [
      { id: 'cli_01', name: 'Lucía Fernández' },
      { id: 'cli_02', name: 'Ignacio Sosa' },
    ];

    render(<RoleSwitcher clients={clients} onChange={onChange} />);

    expect(screen.getByRole('button', { name: /cambiar rol/i })).toHaveTextContent('Contador ▾');

    await user.click(screen.getByRole('button', { name: /cambiar rol/i }));
    const clientItem = await screen.findByText('Lucía Fernández');
    await user.click(clientItem);

    expect(onChange).toHaveBeenCalledWith({ kind: 'client', clientId: 'cli_01' });
    expect(getRole()).toEqual({ kind: 'client', clientId: 'cli_01' });
  });

  it('permite ingresar un clientId manual cuando no hay lista', async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();

    render(<RoleSwitcher clients={[]} onChange={onChange} />);

    const input = screen.getByPlaceholderText('cli_demo_15');
    await user.type(input, 'cli_custom');
    await user.click(screen.getByRole('button', { name: 'Ir' }));

    expect(onChange).toHaveBeenCalledWith({ kind: 'client', clientId: 'cli_custom' });
    expect(getRole()).toEqual({ kind: 'client', clientId: 'cli_custom' });
  });

  it('redirige la ruta raíz / según el rol activo', async () => {
    setRole({ kind: 'accountant' });
    const qc = createQueryClient();
    const router = createMemoryRouter(routes, { initialEntries: ['/'] });

    render(
      <Providers client={qc}>
        <RouterProvider router={router} />
      </Providers>,
    );

    expect(await screen.findByRole('heading', { name: /Cierre de/i })).toBeInTheDocument();
    expect(screen.getByText('Contabilidad')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Cierres' })).toBeInTheDocument();
  });

  it('permite navegar a Parámetros en modo contador', async () => {
    setRole({ kind: 'accountant' });
    const qc = createQueryClient();
    const router = createMemoryRouter(routes, { initialEntries: ['/contador/parametros'] });

    render(
      <Providers client={qc}>
        <RouterProvider router={router} />
      </Providers>,
    );

    expect(await screen.findByRole('heading', { name: 'Parámetros' })).toBeInTheDocument();
  });

  it('muestra título de Impuestos en vista cliente y oculta nav de contador', async () => {
    setRole({ kind: 'client', clientId: 'cli_01' });
    const qc = createQueryClient();
    const router = createMemoryRouter(routes, { initialEntries: ['/cliente'] });

    render(
      <Providers client={qc}>
        <RouterProvider router={router} />
      </Providers>,
    );

    expect(await screen.findByRole('heading', { name: /Hola,/i })).toBeInTheDocument();
    expect(screen.getByText('Impuestos')).toBeInTheDocument();
    expect(screen.queryByRole('link', { name: 'Cierres' })).not.toBeInTheDocument();
  });
});
