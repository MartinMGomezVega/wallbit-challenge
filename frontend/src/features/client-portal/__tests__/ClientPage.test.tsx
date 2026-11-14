import { describe, it, expect, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { QueryClient } from '@tanstack/react-query';
import { http, HttpResponse } from 'msw';
import { server } from '../../../mocks/server';
import { ClientPage } from '../ClientPage';
import { Providers } from '../../../app/providers';
import { setRole, resetRoleForTests } from '../../../app/roleStore';

function createTestQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
        staleTime: 0,
      },
    },
  });
}

describe('ClientPage', () => {
  beforeEach(() => {
    resetRoleForTests();
    setRole({ kind: 'client', clientId: 'cli_demo_15' });
  });

  it('renders client greeting with CUIT masked and period cards', async () => {
    const qc = createTestQueryClient();
    render(
      <Providers client={qc}>
        <ClientPage />
      </Providers>
    );

    expect(await screen.findByRole('heading', { name: /Hola, Paula Rojas/i })).toBeInTheDocument();
    expect(screen.getByText(/CUIT 20-\*\*\*\*5678-6 · MONOTRIBUTO \(Categoría A\)/i)).toBeInTheDocument();

    // Active period (Octubre - ACTION_REQUIRED)
    expect(screen.getByText(/Cierre de octubre 2026/i)).toBeInTheDocument();
    expect(screen.getByText('Acción requerida')).toBeInTheDocument();

    // History section (Septiembre - READY)
    expect(screen.getByText('Historial de cierres anteriores')).toBeInTheDocument();
    expect(screen.getByText('Septiembre 2026')).toBeInTheDocument();
  });

  it('renders error state when profile query fails', async () => {
    const qc = createTestQueryClient();
    server.use(
      http.get('/api/v1/me', () => {
        return HttpResponse.json({ error: { code: 'INTERNAL', message: 'Fail' } }, { status: 500 });
      })
    );

    render(
      <Providers client={qc}>
        <ClientPage />
      </Providers>
    );

    expect(
      await screen.findByText('Ocurrió un error al cargar la información. Por favor intentá nuevamente.')
    ).toBeInTheDocument();
  });
});
