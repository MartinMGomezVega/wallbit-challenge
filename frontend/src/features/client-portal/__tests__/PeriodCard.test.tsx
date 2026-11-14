import { describe, it, expect } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import { server } from '../../../mocks/server';
import { PeriodCard } from '../PeriodCard';
import { ReadySummary } from '../ReadySummary';
import { ClientClosing } from '../schemas';
import { Providers, createQueryClient } from '../../../app/providers';

const mockReadyClosing: ClientClosing = {
  closingId: 'clo_ready_1',
  period: '2026-10',
  state: 'READY',
  updating: false,
  published: {
    version: 1,
    title: 'Cierre de Octubre 2026 listo',
    body: 'Detalle de liquidación mensual.',
    figures: {
      billedUSD: { amount: 200000, currency: 'USD' },
      billedARS: { amount: 240000000, currency: 'ARS' },
      items: [
        {
          concept: 'Cuota Monotributo',
          amount: { amount: 2660000, currency: 'ARS' },
          dueDate: '2026-11-20T00:00:00Z',
        },
      ],
      total: { amount: 2660000, currency: 'ARS' },
    },
    publishedAt: '2026-11-02T10:00:00Z',
  },
  paidAt: null,
  input: {
    answered: true,
    confirmed: true,
    invoicesCount: 4,
  },
};

describe('ReadySummary', () => {
  it('renders headline, title, body, concepts table and buttons', () => {
    const qc = createQueryClient();
    render(
      <Providers client={qc}>
        <ReadySummary closing={mockReadyClosing} />
      </Providers>
    );

    expect(screen.getByText(/Tenés que pagar \$ 26\.600,00/i)).toBeInTheDocument();
    expect(screen.getByText('Cierre de Octubre 2026 listo')).toBeInTheDocument();
    expect(screen.getByText('Detalle de liquidación mensual.')).toBeInTheDocument();
    expect(screen.getByText('Cuota Monotributo')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Pagar en ARCA/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Ya pagué/i })).toBeInTheDocument();
  });

  it('handles clicking "Ya pagué" and shows paid badge when paidAt is present', async () => {
    const user = userEvent.setup();
    const qc = createQueryClient();
    let paidCalled = false;

    server.use(
      http.post('/api/v1/me/closings/:id/paid', ({ params }) => {
        paidCalled = true;
        return HttpResponse.json({
          closingId: params.id,
          paidAt: '2026-11-10T12:00:00Z',
        });
      })
    );

    const { rerender } = render(
      <Providers client={qc}>
        <ReadySummary closing={mockReadyClosing} />
      </Providers>
    );

    const payButton = screen.getByRole('button', { name: /Ya pagué/i });
    await user.click(payButton);

    await waitFor(() => {
      expect(paidCalled).toBe(true);
    });

    // Re-render with paidAt set
    const paidClosing: ClientClosing = {
      ...mockReadyClosing,
      paidAt: '2026-11-10T12:00:00Z',
    };

    rerender(
      <Providers client={qc}>
        <ReadySummary closing={paidClosing} />
      </Providers>
    );

    expect(screen.getByTestId('paid-badge')).toHaveTextContent(/Pagado el 10\/11/i);
    expect(screen.queryByRole('button', { name: /Ya pagué/i })).not.toBeInTheDocument();
    expect(screen.getByText('Pago registrado')).toBeInTheDocument();
  });
});

describe('PeriodCard', () => {
  it('renders ReadySummary when state is READY', () => {
    const qc = createQueryClient();
    render(
      <Providers client={qc}>
        <PeriodCard closing={mockReadyClosing} regime="MONOTRIBUTO" />
      </Providers>
    );

    expect(screen.getByText(/Tenés que pagar/i)).toBeInTheDocument();
  });

  it('renders review message and summary when state is IN_REVIEW', () => {
    const qc = createQueryClient();
    const inReviewClosing: ClientClosing = {
      ...mockReadyClosing,
      state: 'IN_REVIEW',
      published: null,
      input: {
        answered: true,
        confirmed: true,
        invoicesCount: 5,
      },
    };

    render(
      <Providers client={qc}>
        <PeriodCard closing={inReviewClosing} regime="MONOTRIBUTO" />
      </Providers>
    );

    expect(screen.getByText('En revisión')).toBeInTheDocument();
    expect(screen.getByText(/Tu contador está revisando tu cierre de octubre 2026/i)).toBeInTheDocument();
    expect(screen.getByText('5')).toBeInTheDocument();
    expect(screen.getByText('Declarados')).toBeInTheDocument();
  });

  it('renders updating badge and previous snapshot when IN_REVIEW has old published', () => {
    const qc = createQueryClient();
    const updatingClosing: ClientClosing = {
      ...mockReadyClosing,
      state: 'IN_REVIEW',
      updating: true,
      published: mockReadyClosing.published,
    };

    render(
      <Providers client={qc}>
        <PeriodCard closing={updatingClosing} regime="MONOTRIBUTO" />
      </Providers>
    );

    expect(screen.getByTestId('updating-badge')).toHaveTextContent('Actualizando');
    expect(
      screen.getByText(/Mostrando liquidación anterior. Los nuevos valores se actualizarán una vez aprobados/i)
    ).toBeInTheDocument();
    expect(screen.getByText(/Tenés que pagar/i)).toBeInTheDocument();
  });

  it('renders 3-step checklist when state is ACTION_REQUIRED', () => {
    const qc = createQueryClient();
    const actionRequiredClosing: ClientClosing = {
      ...mockReadyClosing,
      state: 'ACTION_REQUIRED',
      published: null,
      input: {
        answered: false,
        confirmed: false,
        invoicesCount: 0,
      },
    };

    render(
      <Providers client={qc}>
        <PeriodCard closing={actionRequiredClosing} regime="MONOTRIBUTO" />
      </Providers>
    );

    expect(screen.getByText('Acción requerida')).toBeInTheDocument();
    expect(screen.getByText(/1\. Subí tus facturas/i)).toBeInTheDocument();
    expect(screen.getByText(/2\. ¿Cobraste algo fuera de Wallbit/i)).toBeInTheDocument();
    expect(screen.getByText(/3\. Confirmación/i)).toBeInTheDocument();
  });
});
