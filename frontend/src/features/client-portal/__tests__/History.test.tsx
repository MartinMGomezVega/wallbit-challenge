import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { History } from '../History';
import { ClientClosing } from '../schemas';

const mockPastClosings: ClientClosing[] = [
  {
    closingId: 'clo_past_sep',
    period: '2026-09',
    state: 'READY',
    updating: false,
    published: {
      version: 1,
      title: 'Cierre de Septiembre 2026',
      body: 'Detalle del mes de septiembre.',
      figures: {
        billedUSD: { amount: 150000, currency: 'USD' },
        billedARS: { amount: 180000000, currency: 'ARS' },
        items: [
          {
            concept: 'Cuota Monotributo',
            amount: { amount: 2660000, currency: 'ARS' },
            dueDate: '2026-10-20T00:00:00Z',
          },
        ],
        total: { amount: 2660000, currency: 'ARS' },
      },
      publishedAt: '2026-10-02T10:00:00Z',
    },
    paidAt: '2026-10-15T10:00:00Z',
    input: {
      answered: true,
      confirmed: true,
      invoicesCount: 3,
    },
  },
];

describe('History', () => {
  it('renders list of past periods with total and paid badge', () => {
    render(<History closings={mockPastClosings} />);

    expect(screen.getByText('Historial de cierres anteriores')).toBeInTheDocument();
    expect(screen.getByText('Septiembre 2026')).toBeInTheDocument();
    expect(screen.getByText('$ 26.600,00')).toBeInTheDocument();
    expect(screen.getByText(/Pagado el 15\/10/i)).toBeInTheDocument();
  });

  it('toggles detail accordion on click and shows billing and concepts', async () => {
    const user = userEvent.setup();
    render(<History closings={mockPastClosings} />);

    expect(screen.queryByText('Detalle del mes de septiembre.')).not.toBeInTheDocument();

    const toggle = screen.getByText('Ver detalle');
    await user.click(toggle);

    expect(screen.getByText('Detalle del mes de septiembre.')).toBeInTheDocument();
    expect(screen.getByText('US$ 1.500,00')).toBeInTheDocument();
    expect(screen.getByText('$ 1.800.000,00')).toBeInTheDocument();
    expect(screen.getByText('Cuota Monotributo')).toBeInTheDocument();

    await user.click(screen.getByText('Ocultar'));
    expect(screen.queryByText('Detalle del mes de septiembre.')).not.toBeInTheDocument();
  });

  it('renders empty state when no past periods have published data', () => {
    render(<History closings={[]} />);
    expect(screen.getByText('No hay períodos anteriores cerrados todavía.')).toBeInTheDocument();
  });
});

