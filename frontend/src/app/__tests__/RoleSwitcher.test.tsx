import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { RoleSwitcher } from '../RoleSwitcher';
import { Providers, createQueryClient } from '../providers';
import { resetRoleForTests, getRole } from '../roleStore';

describe('RoleSwitcher', () => {
  beforeEach(() => {
    resetRoleForTests();
    window.localStorage.clear();
  });

  it('fetches demo clients from API and lists them in dropdown', async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    const qc = createQueryClient();

    render(
      <Providers client={qc}>
        <RoleSwitcher onChange={onChange} />
      </Providers>
    );

    const trigger = screen.getByRole('button', { name: /cambiar rol/i });
    expect(trigger).toHaveTextContent('Contador ▾');

    await user.click(trigger);

    // Debe mostrar los clientes mockeados de la API: Lucía Fernández, Ignacio Sosa, Paula Rojas
    const paulaItem = await screen.findByText('Paula Rojas');
    expect(paulaItem).toBeInTheDocument();
    expect(screen.getByText('Lucía Fernández')).toBeInTheDocument();
    expect(screen.getByText('Ignacio Sosa')).toBeInTheDocument();

    await user.click(paulaItem);

    expect(onChange).toHaveBeenCalledWith({ kind: 'client', clientId: 'cli_demo_15' });
    expect(getRole()).toEqual({ kind: 'client', clientId: 'cli_demo_15' });
    expect(screen.getByRole('button', { name: /cambiar rol/i })).toHaveTextContent('Cliente: Paula Rojas ▾');
  });

  it('allows switching back to accountant', async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    const qc = createQueryClient();

    render(
      <Providers client={qc}>
        <RoleSwitcher onChange={onChange} />
      </Providers>
    );

    // Switch to client
    await user.click(screen.getByRole('button', { name: /cambiar rol/i }));
    await user.click(await screen.findByText('Paula Rojas'));

    // Switch back to accountant
    await user.click(screen.getByRole('button', { name: /cambiar rol/i }));
    await user.click(screen.getByText('Contador'));

    expect(onChange).toHaveBeenCalledWith({ kind: 'accountant' });
    expect(getRole()).toEqual({ kind: 'accountant' });
    expect(screen.getByRole('button', { name: /cambiar rol/i })).toHaveTextContent('Contador ▾');
  });

  it('uses provided clients prop when passed', async () => {
    const user = userEvent.setup();
    const qc = createQueryClient();
    const customClients = [
      { id: 'custom_1', name: 'Custom Client One' },
    ];

    render(
      <Providers client={qc}>
        <RoleSwitcher clients={customClients} />
      </Providers>
    );

    await user.click(screen.getByRole('button', { name: /cambiar rol/i }));
    expect(await screen.findByText('Custom Client One')).toBeInTheDocument();
    expect(screen.queryByText('Paula Rojas')).not.toBeInTheDocument();
  });

  it('renders manual input form when clients list is empty', async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    const qc = createQueryClient();

    render(
      <Providers client={qc}>
        <RoleSwitcher clients={[]} onChange={onChange} />
      </Providers>
    );

    const input = screen.getByPlaceholderText('cli_demo_15');
    await user.type(input, 'cli_manual_99');
    await user.click(screen.getByRole('button', { name: 'Ir' }));

    expect(onChange).toHaveBeenCalledWith({ kind: 'client', clientId: 'cli_manual_99' });
    expect(getRole()).toEqual({ kind: 'client', clientId: 'cli_manual_99' });
  });
});
