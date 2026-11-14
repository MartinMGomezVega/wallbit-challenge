import { describe, it, expect } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import { server } from '../../../mocks/server';
import { ExternalIncomeStep } from '../steps/ExternalIncomeStep';
import { ConfirmStep } from '../steps/ConfirmStep';
import { Providers, createQueryClient } from '../../../app/providers';

describe('ExternalIncomeStep & ConfirmStep', () => {
  it('handles selecting "No" external income', async () => {
    const user = userEvent.setup();
    const qc = createQueryClient();
    let sentPayload: unknown = null;

    server.use(
      http.put('/api/v1/me/periods/:period/external-income', async ({ request }) => {
        sentPayload = await request.json();
        return HttpResponse.json({ ok: true });
      })
    );

    render(
      <Providers client={qc}>
        <ExternalIncomeStep period="2026-10" periodMonthName="octubre 2026" answered={false} />
      </Providers>
    );

    const noButton = screen.getByRole('button', { name: /No, solo en Wallbit/i });
    await user.click(noButton);

    await waitFor(() => {
      expect(sentPayload).toEqual({
        declaredNoExternalIncome: true,
        externalIncome: [],
      });
      expect(screen.getByText('Guardado')).toBeInTheDocument();
    });
  });

  it('handles selecting "Sí" external income and adding items', async () => {
    const user = userEvent.setup();
    const qc = createQueryClient();
    let sentPayload: unknown = null;

    server.use(
      http.put('/api/v1/me/periods/:period/external-income', async ({ request }) => {
        sentPayload = await request.json();
        return HttpResponse.json({ ok: true });
      })
    );

    render(
      <Providers client={qc}>
        <ExternalIncomeStep period="2026-10" periodMonthName="octubre 2026" answered={false} />
      </Providers>
    );

    await user.click(screen.getByRole('button', { name: /Sí, tuve otros cobros/i }));

    const amountInput = screen.getByPlaceholderText(/Monto US\$/i);
    const descInput = screen.getByPlaceholderText(/Descripción/i);

    await user.type(amountInput, '500');
    await user.type(descInput, 'Trabajo freelance exterior');

    await user.click(screen.getByRole('button', { name: /Guardar cobros externos/i }));

    await waitFor(() => {
      expect(sentPayload).toEqual({
        declaredNoExternalIncome: false,
        externalIncome: [
          {
            amount: { amount: 50000, currency: 'USD' },
            description: 'Trabajo freelance exterior',
          },
        ],
      });
      expect(screen.getByText('Guardado')).toBeInTheDocument();
    });
  });

  it('validates invalid external income amounts', async () => {
    const user = userEvent.setup();
    const qc = createQueryClient();

    render(
      <Providers client={qc}>
        <ExternalIncomeStep period="2026-10" periodMonthName="octubre 2026" answered={false} />
      </Providers>
    );

    await user.click(screen.getByRole('button', { name: /Sí, tuve otros cobros/i }));

    const amountInput = screen.getByPlaceholderText(/Monto US\$/i);
    const descInput = screen.getByPlaceholderText(/Descripción/i);

    await user.type(amountInput, '-10');
    await user.type(descInput, 'Desc');

    await user.click(screen.getByRole('button', { name: /Guardar cobros externos/i }));

    expect(await screen.findByText('Todos los montos deben ser mayores a 0')).toBeInTheDocument();
  });

  it('ConfirmStep disables button when answered is false', () => {
    const qc = createQueryClient();
    render(
      <Providers client={qc}>
        <ConfirmStep period="2026-10" answered={false} />
      </Providers>
    );

    const button = screen.getByRole('button', { name: /Listo, ya cargué todo/i });
    expect(button).toBeDisabled();
    expect(
      screen.getByText(/Respondé el paso 2 sobre ingresos externos para habilitar la confirmación/i)
    ).toBeInTheDocument();
  });

  it('ConfirmStep enables button and calls confirm API when clicked', async () => {
    const user = userEvent.setup();
    const qc = createQueryClient();
    let confirmCalled = false;

    server.use(
      http.post('/api/v1/me/periods/:period/confirm', () => {
        confirmCalled = true;
        return HttpResponse.json({ confirmedAt: new Date().toISOString() });
      })
    );

    render(
      <Providers client={qc}>
        <ConfirmStep period="2026-10" answered={true} />
      </Providers>
    );

    const button = screen.getByRole('button', { name: /Listo, ya cargué todo/i });
    expect(button).not.toBeDisabled();

    await user.click(button);

    await waitFor(() => {
      expect(confirmCalled).toBe(true);
    });
  });
});
