import { describe, it, expect } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { RouterProvider, createMemoryRouter } from 'react-router-dom';
import { routes } from '../../../app/router';
import { Providers, createQueryClient } from '../../../app/providers';
import { server } from '../../../mocks/server';
import { http, HttpResponse } from 'msw';
import meClosingsReadyMonoFixture from '../../../test/fixtures/api/me_closings_ready_mono.json';

describe('ClientIntegration (circuito completo de Paula - Criterio 13)', () => {
  it('circuito interactivo: Paula carga factura, responde No, confirma a En revisión, contador aprueba y Paula ve la tarjeta validada', async () => {
    // Estado mutable en el mock para simular el circuito en vivo
    let step = 1; // 1: ACTION_REQUIRED, 2: IN_REVIEW, 3: READY

    server.use(
      http.get('/api/v1/me/closings', ({ request }) => {
        const user = request.headers.get('X-Demo-User') ?? '';
        if (user.includes('cli_demo_15')) {
          if (step === 1) {
            return HttpResponse.json([
              {
                closingId: 'clo_000300',
                period: '2026-10',
                state: 'ACTION_REQUIRED',
                input: { answered: false, confirmed: false, invoicesCount: 0 },
              },
            ]);
          }
          if (step === 2) {
            return HttpResponse.json([
              {
                closingId: 'clo_000300',
                period: '2026-10',
                state: 'IN_REVIEW',
                input: { answered: true, confirmed: true, invoicesCount: 1 },
              },
            ]);
          }
          return HttpResponse.json([
            {
              closingId: 'clo_000300',
              period: '2026-10',
              state: 'READY',
              published: {
                version: 1,
                title: 'Tu cierre de octubre 2026 está listo',
                body: 'Facturaste US$ 850,00 ($ 1.259.827,50). Cuota de Monotributo categoría B: $ 56.379,08.',
                figures: {
                  billedUSD: { amount: 85000, currency: 'USD' },
                  billedARS: { amount: 125982750, currency: 'ARS' },
                  items: [
                    {
                      concept: 'Cuota Monotributo',
                      amount: { amount: 5637908, currency: 'ARS' },
                      dueDate: '2026-11-20T00:00:00-03:00',
                    },
                  ],
                  total: { amount: 5637908, currency: 'ARS' },
                },
                publishedAt: '2026-11-06T12:00:00-03:00',
                validatedBy: {
                  name: 'C.P. Mariana Quiroga',
                  license: 'Matrícula de ejemplo',
                },
                regime: 'MONOTRIBUTO',
                category: 'B',
                pctOfCategoryCapBp: 5400,
                categoryCapARS: { amount: 1700000000, currency: 'ARS' },
                billed12mARS: { amount: 918000000, currency: 'ARS' },
              },
              input: { answered: true, confirmed: true, invoicesCount: 1 },
            },
          ]);
        }
        if (user.includes('cli_demo_01')) {
          return HttpResponse.json(meClosingsReadyMonoFixture);
        }
        return HttpResponse.json([]);
      }),

      http.put('/api/v1/me/periods/2026-10/external-income', () => {
        return HttpResponse.json({ clientId: 'cli_demo_15', period: '2026-10', externalIncome: [], declaredNoExternalIncome: true });
      }),

      http.post('/api/v1/me/periods/2026-10/confirm', () => {
        step = 2; // Avanza a IN_REVIEW
        return HttpResponse.json({ clientId: 'cli_demo_15', period: '2026-10', confirmedBy: 'CLIENT' });
      }),
    );

    const qc = createQueryClient();
    const router = createMemoryRouter(routes, { initialEntries: ['/cliente/cli_demo_15'] });
    render(
      <Providers client={qc}>
        <RouterProvider router={router} />
      </Providers>,
    );

    // 1) Cliente: Paula en ACTION_REQUIRED carga un comprobante
    expect(await screen.findByText('Comprobantes que emitiste en octubre')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'Alta de comprobante' }));

    fireEvent.change(screen.getByLabelText('Número'), { target: { value: '101' } });
    fireEvent.change(screen.getByLabelText('Cliente (denominación)'), { target: { value: 'Brightloop GmbH' } });
    fireEvent.change(screen.getByLabelText('Monto total'), { target: { value: '850.00' } });
    fireEvent.click(screen.getByRole('button', { name: 'Cargar comprobante' }));

    await waitFor(() => {
      expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    });

    // Paula responde "No" a cobros externos
    const radioNo = screen.getByLabelText(/No, todos mis cobros fueron a través de Wallbit/i);
    fireEvent.click(radioNo);

    // Paula confirma con "Listo, ya cargué todo"
    const confirmBtn = screen.getByRole('button', { name: 'Listo, ya cargué todo' });
    expect(confirmBtn).toBeEnabled();
    fireEvent.click(confirmBtn);

    // Paula ve pantalla "En revisión"
    expect(await screen.findByText(/Tu contadora está revisando tu cierre de octubre/i)).toBeInTheDocument();

    // 2) Simulación de la aprobación por el contador:
    step = 3;

    // 3) Paula recarga / vuelve a consultar su cierre (navegando a otro cliente y volviendo):
    fireEvent.click(screen.getByRole('button', { name: /Lucía Fernández/i }));
    expect(await screen.findByRole('article', { name: /Resumen validado de 2026-10/i })).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: /Paula Rojas/i }));

    // Aparece la tarjeta validada con el badge de la contadora
    const paulaCard = await screen.findByRole('article', { name: /Resumen validado de 2026-10/i });
    expect(paulaCard).toBeInTheDocument();
    expect(paulaCard).toHaveTextContent(/Categoría B/i);
    expect(paulaCard).toHaveTextContent(/54% del tope anual/i);

    const badge = screen.getByTestId('validated-badge');
    expect(badge).toHaveTextContent(/Validado por C.P. Mariana Quiroga · Matrícula de ejemplo · 06\/11\/2026/i);
    expect(screen.getByRole('button', { name: 'Ya pagué' })).toBeInTheDocument();
  });
});
