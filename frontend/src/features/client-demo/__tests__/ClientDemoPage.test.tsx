import { describe, it, expect } from 'vitest';
import { render, screen, fireEvent, cleanup } from '@testing-library/react';
import { axe } from 'vitest-axe';
import { RouterProvider, createMemoryRouter } from 'react-router-dom';
import { routes } from '../../../app/router';
import { Providers, createQueryClient } from '../../../app/providers';
import { server } from '../../../mocks/server';
import { http, HttpResponse } from 'msw';

function renderAt(initialUrl: string) {
  const qc = createQueryClient();
  const router = createMemoryRouter(routes, { initialEntries: [initialUrl] });
  return render(
    <Providers client={qc}>
      <RouterProvider router={router} />
    </Providers>,
  );
}

describe('ClientDemoPage', () => {
  it('muestra el marco de celular, la leyenda y el selector con los 5 clientes', async () => {
    renderAt('/cliente/cli_demo_15');

    expect(await screen.findByRole('region', { name: 'Vista demo de la app móvil de Wallbit' })).toBeInTheDocument();
    expect(screen.getByText('Así se vería dentro de la app de Wallbit')).toBeInTheDocument();

    const selector = screen.getByRole('navigation', { name: 'Clientes de demo' });
    expect(selector).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Paula Rojas/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Lucía Fernández/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Camila Torres/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Ignacio Sosa/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Carla Benítez/i })).toBeInTheDocument();
  });

  it('Paula Rojas (cli_demo_15): muestra Check-in con pasos de comprobantes y cobros externos', async () => {
    renderAt('/cliente/cli_demo_15');

    expect(await screen.findByText('Comprobantes que emitiste en octubre')).toBeInTheDocument();
    expect(screen.getByText('¿Cobraste algo fuera de Wallbit en octubre?')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Alta de comprobante' })).toBeInTheDocument();

    const confirmBtn = screen.getByRole('button', { name: 'Listo, ya cargué todo' });
    expect(confirmBtn).toBeDisabled();
    expect(screen.getByText(/Respondé la pregunta sobre cobros externos/i)).toBeInTheDocument();
  });

  it('Alta de comprobante: abre diálogo, valida y maneja error 409 de comprobante duplicado', async () => {
    renderAt('/cliente/cli_demo_15');

    const altaBtn = await screen.findByRole('button', { name: 'Alta de comprobante' });
    fireEvent.click(altaBtn);

    expect(screen.getByRole('dialog', { name: 'Alta de comprobante' })).toBeInTheDocument();

    // Intentar enviar sin número ni cliente
    fireEvent.click(screen.getByRole('button', { name: 'Cargar comprobante' }));
    expect(await screen.findByText('El número debe ser mayor a 0')).toBeInTheDocument();
    expect(screen.getByText('La denominación del receptor es requerida')).toBeInTheDocument();

    // Simular error 409 en el backend
    server.use(
      http.post('/api/v1/me/invoices', () => {
        return HttpResponse.json(
          { error: { code: 'INVOICE_ALREADY_EXISTS', message: 'Ya existe una factura con este tipo, punto de venta y número' } },
          { status: 409 },
        );
      }),
    );

    fireEvent.change(screen.getByLabelText('Número'), { target: { value: '101' } });
    fireEvent.change(screen.getByLabelText('Cliente (denominación)'), { target: { value: 'Acme Inc' } });
    fireEvent.change(screen.getByLabelText('Monto total'), { target: { value: '1500.00' } });

    fireEvent.click(screen.getByRole('button', { name: 'Cargar comprobante' }));
    expect(await screen.findByText('Ese comprobante ya está cargado')).toBeInTheDocument();

    // Cerrar formulario
    fireEvent.click(screen.getByLabelText('Cerrar formulario'));
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  });

  it('Check-in: responder No habilita Listo ya cargué todo y confirma pasando a En revisión', async () => {
    renderAt('/cliente/cli_demo_15');

    expect(await screen.findByText('¿Cobraste algo fuera de Wallbit en octubre?')).toBeInTheDocument();

    const radioNo = screen.getByLabelText(/No, todos mis cobros fueron a través de Wallbit/i);
    fireEvent.click(radioNo);

    const confirmBtn = screen.getByRole('button', { name: 'Listo, ya cargué todo' });
    expect(confirmBtn).toBeEnabled();

    // Mock transición a IN_REVIEW al recargar
    server.use(
      http.get('/api/v1/me/closings', () => {
        return HttpResponse.json([
          {
            closingId: 'clo_000300',
            period: '2026-10',
            state: 'IN_REVIEW',
            input: { answered: true, confirmed: true, invoicesCount: 1 },
          },
        ]);
      }),
    );

    fireEvent.click(confirmBtn);

    expect(await screen.findByText(/Tu contadora está revisando tu cierre de octubre/i)).toBeInTheDocument();
  });

  it('Lucía Fernández (cli_demo_01): muestra tarjeta validada Monotributo Cat A y botón Ya pagué', async () => {
    renderAt('/cliente/cli_demo_01');

    const card = await screen.findByRole('article', { name: /Resumen validado de 2026-10/i });
    expect(card).toBeInTheDocument();
    expect(card).toHaveTextContent(/Categoría A/i);
    expect(card).toHaveTextContent(/71% del tope anual/i);

    const badge = screen.getByTestId('validated-badge');
    expect(badge).toHaveTextContent(/Validado por C.P. Mariana Quiroga · Matrícula de ejemplo/i);

    const payBtn = screen.getByRole('button', { name: 'Ya pagué' });
    expect(payBtn).toBeInTheDocument();

    // Registrar pago
    fireEvent.click(payBtn);
    expect(await screen.findByTestId('paid-badge')).toHaveTextContent(/Pagado el 05\/11/i);
  });

  it('Camila Torres (cli_demo_09): muestra tarjeta validada Responsable Inscripto con IVA y autónomos', async () => {
    renderAt('/cliente/cli_demo_09');

    expect(await screen.findByRole('article', { name: /Resumen validado de 2026-10/i })).toBeInTheDocument();
    expect(screen.getByText('Responsable Inscripto')).toBeInTheDocument();
    expect(screen.getByText('IVA')).toBeInTheDocument();
    expect(screen.getByText('Autónomos')).toBeInTheDocument();

    const badge = screen.getByTestId('validated-badge');
    expect(badge).toHaveTextContent(/Validado por C.P. Mariana Quiroga/i);
  });

  it('Ignacio Sosa (cli_demo_02): muestra En revisión de octubre y tarjeta validada de septiembre abajo', async () => {
    renderAt('/cliente/cli_demo_02');

    expect(await screen.findByText(/Tu contadora está revisando tu cierre de octubre/i)).toBeInTheDocument();
    expect(screen.getByText(/Mes anterior \(Septiembre\)/i)).toBeInTheDocument();
    expect(screen.getByRole('article', { name: /Resumen validado de 2026-09/i })).toBeInTheDocument();
  });

  it('muestra mensaje de error y permite reintentar cuando falla la carga', async () => {
    let failOnce = true;
    server.use(
      http.get('/api/v1/me/closings', () => {
        if (failOnce) {
          failOnce = false;
          return HttpResponse.json({ error: { code: 'SERVER_ERROR', message: 'Falla al cargar datos' } }, { status: 500 });
        }
        return HttpResponse.json([
          {
            closingId: 'cls_15',
            period: '2026-10',
            state: 'ACTION_REQUIRED',
            input: { answered: false, confirmed: false, invoicesCount: 0 },
          },
        ]);
      }),
    );

    renderAt('/cliente/cli_demo_15');
    expect(await screen.findByText('Falla al cargar datos')).toBeInTheDocument();

    const retryBtn = screen.getByRole('button', { name: /Reintentar/i });
    fireEvent.click(retryBtn);

    expect(await screen.findByText('Comprobantes que emitiste en octubre')).toBeInTheDocument();
  });

  it('muestra estado vacío cuando no hay cierre para este mes', async () => {
    server.use(
      http.get('/api/v1/me/closings', () => {
        return HttpResponse.json([]);
      }),
    );

    renderAt('/cliente/cli_demo_15');
    expect(await screen.findByText('Todavía no hay un cierre para este mes.')).toBeInTheDocument();
  });

  it('cumple accesibilidad (vitest-axe sin violaciones en Check-in y tarjeta validada)', async () => {
    const { container: containerCheckIn } = renderAt('/cliente/cli_demo_15');
    await screen.findByRole('region', { name: 'Vista demo de la app móvil de Wallbit' });
    const resultsCheckIn = await axe(containerCheckIn);
    expect(resultsCheckIn.violations).toHaveLength(0);

    cleanup();

    const { container: containerValidated } = renderAt('/cliente/cli_demo_01');
    await screen.findByRole('article', { name: /Resumen validado de 2026-10/i });
    const resultsValidated = await axe(containerValidated);
    expect(resultsValidated.violations).toHaveLength(0);
  });
});
