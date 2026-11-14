import { describe, it, expect } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { http, HttpResponse } from 'msw';
import { server } from '../../../mocks/server';
import { SimulatedDateBadge, SIMULATED_DATE_TOOLTIP } from '../SimulatedDateBadge';
import { formatDate } from '../../format/date';

function renderWithClient(ui: React.ReactElement) {
  const qc = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  });
  return render(<QueryClientProvider client={qc}>{ui}</QueryClientProvider>);
}

describe('SimulatedDateBadge', () => {
  it('muestra la fecha fija de /health con "fecha simulada"', async () => {
    server.use(
      http.get('/api/v1/health', () =>
        HttpResponse.json({ status: 'ok', now: '2026-11-05T09:00:00-03:00' })
      )
    );

    renderWithClient(<SimulatedDateBadge />);

    const badge = await screen.findByTestId('simulated-date-badge');
    expect(badge).toBeInTheDocument();
    expect(badge).toHaveTextContent('Hoy: 05/11/2026 · fecha simulada');
  });

  it('si coincide con la fecha real del sistema (APP_NOW vacía), no incluye "fecha simulada"', async () => {
    const today = new Date();
    const pad = (n: number) => String(n).padStart(2, '0');
    const localDateStr = `${today.getFullYear()}-${pad(today.getMonth() + 1)}-${pad(today.getDate())}T12:00:00-03:00`;
    const todayFormatted = formatDate(localDateStr, 'full');

    server.use(
      http.get('/api/v1/health', () =>
        HttpResponse.json({ status: 'ok', now: localDateStr })
      )
    );

    renderWithClient(<SimulatedDateBadge />);

    const badge = await screen.findByTestId('simulated-date-badge');
    await waitFor(() => {
      expect(badge).toHaveTextContent(`Hoy: ${todayFormatted}`);
      expect(badge).not.toHaveTextContent('fecha simulada');
    });
  });

  it('el tooltip se activa y es accesible por foco de teclado', async () => {
    const user = userEvent.setup();
    server.use(
      http.get('/api/v1/health', () =>
        HttpResponse.json({ status: 'ok', now: '2026-11-05T09:00:00-03:00' })
      )
    );

    renderWithClient(<SimulatedDateBadge />);

    const badge = await screen.findByTestId('simulated-date-badge');

    // Navegar y enfocar con teclado
    await user.tab();
    expect(badge).toHaveFocus();

    // El tooltip debe ser visible con el texto explicativo
    await waitFor(() => {
      expect(screen.getByText(SIMULATED_DATE_TOOLTIP)).toBeInTheDocument();
    });
  });
});
