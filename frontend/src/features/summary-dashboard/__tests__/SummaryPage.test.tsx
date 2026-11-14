import { describe, it, expect } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter, Route, Routes, useLocation } from 'react-router-dom';
import { axe } from 'vitest-axe';
import { server } from '../../../mocks/server';
import { SummaryPage } from '../SummaryPage';

function LocationDisplay() {
  const location = useLocation();
  return <div data-testid="location-display">{location.pathname + location.search}</div>;
}

function renderSummary(initialUrl = '/contador/resumen?period=2026-10') {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  });

  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={[initialUrl]}>
        <Routes>
          <Route path="/contador/resumen" element={<SummaryPage />} />
          <Route path="/contador" element={<LocationDisplay />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>
  );
}

describe('SummaryPage', () => {
  it('renderiza KPIs, gráficos, tablas sr-only y métricas de éxito', async () => {
    renderSummary();

    expect(screen.getByText(/Resumen de octubre 2026/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/Cargando resumen del período/i)).toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByText('220 de 300 sin aprobar')).toBeInTheDocument();
    });

    expect(screen.getByText('≈ 17 h 36 min de revisión')).toBeInTheDocument();
    expect(screen.getByText('US$ 100.000,00')).toBeInTheDocument();
    expect(screen.getByText('$ 150.000.000,00')).toBeInTheDocument();
    expect(screen.getAllByText('$ 5.000.000,00').length).toBeGreaterThanOrEqual(1);
    expect(screen.getByText(/Próximo vence 20\/11/i)).toBeInTheDocument();
    expect(screen.getByText('223 de 300 clientes')).toBeInTheDocument();

    // Títulos de gráficos y tablas equivalentes sr-only
    expect(screen.getByText('Situaciones detectadas')).toBeInTheDocument();
    expect(screen.getByText('Distribución por categoría')).toBeInTheDocument();
    expect(screen.getByText('9 cerca del tope K')).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'Vencimientos próximos' })).toBeInTheDocument();

    expect(screen.getByText('Faltan facturas')).toBeInTheDocument();
    expect(screen.getByText('Cerca del tope K')).toBeInTheDocument();

    // Métricas de éxito
    expect(screen.getByText('Clientes en verde')).toBeInTheDocument();
    expect(screen.getByText('Avisados antes del día 10')).toBeInTheDocument();
    expect(screen.getByText('Pagos en término')).toBeInTheDocument();
    expect(screen.getByText('27%')).toBeInTheDocument();
    expect(screen.getByText('63%')).toBeInTheDocument();
  });

  it('navega al tablero con filtros al hacer click en segmento de progreso, KPI Todo ok y señal', async () => {
    const user = userEvent.setup();

    // 1. Segmento de barra de progreso -> status=NOTIFIED
    const { unmount } = renderSummary();
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /Notificado: 60 clientes/i })).toBeInTheDocument();
    });
    await user.click(screen.getByRole('button', { name: /Notificado: 60 clientes/i }));
    expect(screen.getByTestId('location-display')).toHaveTextContent('/contador?period=2026-10&status=NOTIFIED');
    unmount();

    // 2. KPI Todo ok -> severity=GREEN
    const { unmount: unmount2 } = renderSummary();
    await waitFor(() => {
      expect(screen.getByRole('link', { name: /Todo ok/i })).toBeInTheDocument();
    });
    await user.click(screen.getByRole('link', { name: /Todo ok/i }));
    expect(screen.getByTestId('location-display')).toHaveTextContent('/contador?period=2026-10&severity=GREEN');
    unmount2();

    // 3. Señal -> signal=MISSING_INVOICES
    renderSummary();
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /Ver clientes con Faltan facturas/i })).toBeInTheDocument();
    });
    await user.click(screen.getByRole('button', { name: /Ver clientes con Faltan facturas/i }));
    expect(screen.getByTestId('location-display')).toHaveTextContent('/contador?period=2026-10&signal=MISSING_INVOICES');
  });

  it('cambia de período desde el selector', async () => {
    const user = userEvent.setup();
    renderSummary();

    await waitFor(() => {
      expect(screen.getByText('220 de 300 sin aprobar')).toBeInTheDocument();
    });

    await user.selectOptions(screen.getByLabelText(/Período/i), '2026-09');
    await waitFor(() => {
      expect(screen.getByText(/Resumen de septiembre 2026/i)).toBeInTheDocument();
    });
  });

  it('muestra estado vacío cuando no hay cierres en el período', async () => {
    server.use(
      http.get('/api/v1/accountant/periods/:period/stats', ({ params }) => {
        return HttpResponse.json({
          period: String(params.period),
          progress: {
            total: 0,
            byStatus: { PENDING_CLIENT_INPUT: 0, READY_FOR_REVIEW: 0, APPROVED: 0, NOTIFIED: 0 },
            approvedPct: 0,
            pendingReview: { RED: 0, YELLOW: 0, GREEN: 0 },
            estimatedReviewMinutes: 0,
          },
          totals: {
            billedUSD: { amount: 0, currency: 'USD' },
            billedARS: { amount: 0, currency: 'ARS' },
            toPayARS: { amount: 0, currency: 'ARS' },
            toPayByDueDate: [],
          },
          risk: {
            bySeverity: { RED: 0, YELLOW: 0, GREEN: 0 },
            bySignal: [],
            byCategory: [],
            nearKCap: 0,
          },
          success: { greenPct: 0, notifiedBeforeDay10Pct: 0, paidOnTimePct: 0 },
        });
      })
    );

    renderSummary('/contador/resumen?period=2026-08');
    await waitFor(() => {
      expect(screen.getByText(/No hay cierres para agosto 2026/i)).toBeInTheDocument();
    });
    expect(screen.getByRole('button', { name: /Ver octubre 2026/i })).toBeInTheDocument();
  });

  it('muestra error y permite reintentar cuando falla la API', async () => {
    const user = userEvent.setup();
    let shouldFail = true;
    server.use(
      http.get('/api/v1/accountant/periods/:period/stats', () => {
        if (shouldFail) {
          return HttpResponse.json(
            { error: { code: 'INTERNAL_ERROR', message: 'Fallo temporal' } },
            { status: 500 }
          );
        }
        return HttpResponse.json({
          period: '2026-10',
          progress: {
            total: 10,
            byStatus: { PENDING_CLIENT_INPUT: 0, READY_FOR_REVIEW: 5, APPROVED: 0, NOTIFIED: 5 },
            approvedPct: 5000,
            pendingReview: { RED: 0, YELLOW: 0, GREEN: 5 },
            estimatedReviewMinutes: 10,
          },
          totals: {
            billedUSD: { amount: 10000, currency: 'USD' },
            billedARS: { amount: 15000000, currency: 'ARS' },
            toPayARS: { amount: 100000, currency: 'ARS' },
            toPayByDueDate: [],
          },
          risk: {
            bySeverity: { RED: 0, YELLOW: 0, GREEN: 10 },
            bySignal: [],
            byCategory: [],
            nearKCap: 0,
          },
          success: { greenPct: 10000, notifiedBeforeDay10Pct: 5000, paidOnTimePct: 10000 },
        });
      })
    );

    renderSummary();
    await waitFor(() => {
      expect(screen.getByRole('alert')).toBeInTheDocument();
    });
    expect(screen.getByRole('alert')).toHaveAttribute('title', 'INTERNAL_ERROR');

    shouldFail = false;
    await user.click(screen.getByRole('button', { name: /Reintentar/i }));
    await waitFor(() => {
      expect(screen.getByText('5 de 10 sin aprobar')).toBeInTheDocument();
    });
  });

  it('cumple accesibilidad (vitest-axe sin violaciones)', async () => {
    const { container } = renderSummary();
    await waitFor(() => {
      expect(screen.getByText('220 de 300 sin aprobar')).toBeInTheDocument();
    });

    const results = await axe(container);
    expect(results.violations).toHaveLength(0);
  });
});
