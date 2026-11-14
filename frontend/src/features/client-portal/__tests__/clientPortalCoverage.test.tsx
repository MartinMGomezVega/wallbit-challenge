import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import { server } from '../../../mocks/server';
import { ConfirmStep } from '../steps/ConfirmStep';
import { ExternalIncomeStep } from '../steps/ExternalIncomeStep';
import { UploadInvoicesStep } from '../steps/UploadInvoicesStep';
import { ManualInvoiceForm } from '../forms/ManualInvoiceForm';
import { PurchaseInvoiceForm } from '../forms/PurchaseInvoiceForm';
import { History } from '../History';
import { ClientPage } from '../ClientPage';
import { ComingSoonButton } from '../../../shared/ui/ComingSoonButton';
import { Providers, createQueryClient } from '../../../app/providers';
import { setRole, resetRoleForTests } from '../../../app/roleStore';
import { ClientClosing } from '../schemas';

describe('Client Portal Coverage Boost', () => {
  it('ConfirmStep handles server error', async () => {
    const user = userEvent.setup();
    const qc = createQueryClient();

    server.use(
      http.post('/api/v1/me/periods/:period/confirm', () => {
        return HttpResponse.json({ error: { code: 'FAIL', message: 'Error de servidor' } }, { status: 500 });
      })
    );

    render(
      <Providers client={qc}>
        <ConfirmStep period="2026-10" answered={true} />
      </Providers>
    );

    await user.click(screen.getByRole('button', { name: /Listo, ya cargué todo/i }));

    expect(await screen.findByText('Error de servidor')).toBeInTheDocument();
  });

  it('ExternalIncomeStep handles server error, empty description, and item removal', async () => {
    const user = userEvent.setup();
    const qc = createQueryClient();

    render(
      <Providers client={qc}>
        <ExternalIncomeStep period="2026-10" periodMonthName="octubre 2026" answered={false} />
      </Providers>
    );

    await user.click(screen.getByRole('button', { name: /Sí, tuve otros cobros/i }));

    // Agregar segundo ítem
    await user.click(screen.getByRole('button', { name: /Agregar otro cobro/i }));

    const amountInputs = screen.getAllByPlaceholderText(/Monto US\$/i);
    const descInputs = screen.getAllByPlaceholderText(/Descripción/i);

    // Llenar primer monto pero dejar descripción vacía
    await user.type(amountInputs[0]!, '100');
    // Segundo monto inválido o vacío
    await user.click(screen.getByRole('button', { name: /Guardar cobros externos/i }));

    expect(await screen.findByText('Por favor completá la descripción de cada cobro')).toBeInTheDocument();

    // Eliminar el segundo ítem
    const removeButtons = screen.getAllByRole('button', { name: /Eliminar cobro/i });
    await user.click(removeButtons[1]!);

    // Simular error de servidor al guardar
    server.use(
      http.put('/api/v1/me/periods/:period/external-income', () => {
        return HttpResponse.json({ error: { code: 'FAIL', message: 'Fallo al guardar externos' } }, { status: 500 });
      })
    );

    await user.type(descInputs[0]!, 'Trabajo exterior');
    await user.click(screen.getByRole('button', { name: /Guardar cobros externos/i }));

    expect(await screen.findByText('Fallo al guardar externos')).toBeInTheDocument();
  });

  it('UploadInvoicesStep dropzone keyboard trigger', async () => {
    const qc = createQueryClient();
    render(
      <Providers client={qc}>
        <UploadInvoicesStep periodMonthName="octubre 2026" regime="MONOTRIBUTO" invoicesCount={0} />
      </Providers>
    );

    const dropzone = screen.getByRole('button', { name: /Arrastrá el CSV/i });
    fireEvent.keyDown(dropzone, { key: 'Enter' });
    fireEvent.keyDown(dropzone, { key: ' ' });
  });

  it('ManualInvoiceForm validates amount, handles server error, and cancels', async () => {
    const user = userEvent.setup();
    const qc = createQueryClient();

    render(
      <Providers client={qc}>
        <ManualInvoiceForm regime="MONOTRIBUTO" />
      </Providers>
    );

    await user.click(screen.getByRole('button', { name: /Cargar una factura a mano/i }));

    // Monto vacío / 0
    await user.click(screen.getByRole('button', { name: 'Guardar factura' }));
    expect(await screen.findByText('El importe debe ser mayor a 0')).toBeInTheDocument();

    // Simular error de servidor
    server.use(
      http.post('/api/v1/me/invoices', () => {
        return HttpResponse.json({ error: { code: 'FAIL', message: 'No se pudo crear factura' } }, { status: 500 });
      })
    );

    const amountInput = screen.getByPlaceholderText('1000.00');
    await user.type(amountInput, '500');
    await user.click(screen.getByRole('button', { name: 'Guardar factura' }));

    expect(await screen.findByText('No se pudo crear factura')).toBeInTheDocument();

    // Cancelar
    await user.click(screen.getByRole('button', { name: 'Cancelar' }));
    await waitFor(() => {
      expect(screen.queryByRole('heading', { name: 'Cargar factura emitida' })).not.toBeInTheDocument();
    });
  });

  it('PurchaseInvoiceForm validates amount, handles server error, and cancels', async () => {
    const user = userEvent.setup();
    const qc = createQueryClient();

    render(
      <Providers client={qc}>
        <PurchaseInvoiceForm />
      </Providers>
    );

    await user.click(screen.getByRole('button', { name: /Cargar factura de compra/i }));

    // Monto vacío / 0
    await user.click(screen.getByRole('button', { name: 'Guardar compra' }));
    expect(await screen.findByText('El importe neto debe ser mayor a 0')).toBeInTheDocument();

    // Simular error de servidor
    server.use(
      http.post('/api/v1/me/purchase-invoices', () => {
        return HttpResponse.json({ error: { code: 'FAIL', message: 'Error compra servidor' } }, { status: 500 });
      })
    );

    await user.type(screen.getByPlaceholderText('1000.00'), '500');
    await user.click(screen.getByRole('button', { name: 'Guardar compra' }));

    expect(await screen.findByText('Error compra servidor')).toBeInTheDocument();

    // Cancelar
    await user.click(screen.getByRole('button', { name: 'Cancelar' }));
    await waitFor(() => {
      expect(screen.queryByRole('heading', { name: 'Cargar factura de compra (RI)' })).not.toBeInTheDocument();
    });
  });

  it('History supports keyboard navigation on items', () => {
    const mockClosing: ClientClosing = {
      closingId: 'c1',
      period: '2026-08',
      state: 'READY',
      published: {
        version: 1,
        title: 'Cierre Agosto',
        body: 'Texto agosto',
        figures: {
          billedUSD: { amount: 100000, currency: 'USD' },
          billedARS: { amount: 120000000, currency: 'ARS' },
          items: [],
          total: { amount: 2660000, currency: 'ARS' },
        },
        publishedAt: '2026-09-01T10:00:00Z',
      },
      input: { answered: true, confirmed: true, invoicesCount: 1 },
    };

    render(<History closings={[mockClosing]} />);

    const itemRow = screen.getByRole('button', { name: /Agosto 2026/i });
    fireEvent.keyDown(itemRow, { key: 'Enter' });
    expect(screen.getByText('Texto agosto')).toBeInTheDocument();

    fireEvent.keyDown(itemRow, { key: ' ' });
    expect(screen.queryByText('Texto agosto')).not.toBeInTheDocument();
  });

  it('ComingSoonButton forwards onKeyDown for non-inert keys', () => {
    const handleKeyDown = vi.fn();
    render(<ComingSoonButton onKeyDown={handleKeyDown}>Botón</ComingSoonButton>);

    const button = screen.getByRole('button', { name: 'Botón' });
    fireEvent.keyDown(button, { key: 'ArrowDown' });
    expect(handleKeyDown).toHaveBeenCalled();
  });

  it('ClientPage renders empty state when no closings exist', async () => {
    resetRoleForTests();
    setRole({ kind: 'client', clientId: 'cli_demo_15' });
    const qc = createQueryClient();
    server.use(
      http.get('/api/v1/me/closings', () => {
        return HttpResponse.json([]);
      })
    );
    render(
      <Providers client={qc}>
        <ClientPage />
      </Providers>
    );

    expect(await screen.findByText('No hay cierres disponibles para tu cuenta.')).toBeInTheDocument();
  });

  it('ConfirmStep calls onSuccess on successful confirmation', async () => {
    const user = userEvent.setup();
    const qc = createQueryClient();
    const onSuccess = vi.fn();

    server.use(
      http.post('/api/v1/me/periods/:period/confirm', () => {
        return HttpResponse.json({ confirmedAt: '2026-10-31T20:00:00Z' });
      })
    );

    render(
      <Providers client={qc}>
        <ConfirmStep period="2026-10" answered={true} onSuccess={onSuccess} />
      </Providers>
    );

    await user.click(screen.getByRole('button', { name: /Listo, ya cargué todo/i }));
    await waitFor(() => {
      expect(onSuccess).toHaveBeenCalled();
    });
  });

  it('ManualInvoiceForm fills all fields and submits successfully', async () => {
    const user = userEvent.setup();
    const qc = createQueryClient();
    const onSuccess = vi.fn();

    server.use(
      http.post('/api/v1/me/invoices', () => {
        return HttpResponse.json({ id: 'inv_manual_1' });
      })
    );

    render(
      <Providers client={qc}>
        <ManualInvoiceForm regime="RI" onSuccess={onSuccess} />
      </Providers>
    );

    await user.click(screen.getByRole('button', { name: /Cargar una factura a mano/i }));

    const typeSelect = screen.getByLabelText('Tipo');
    fireEvent.change(typeSelect, { target: { value: 'B' } });

    const posInput = screen.getByLabelText('Punto de venta');
    await user.clear(posInput);
    await user.type(posInput, '2');

    const numInput = screen.getByLabelText('Número');
    await user.clear(numInput);
    await user.type(numInput, '15');

    const currencySelect = screen.getByLabelText('Moneda');
    fireEvent.change(currencySelect, { target: { value: 'USD' } });

    const amountInput = screen.getByLabelText(/Importe Total/i);
    await user.type(amountInput, '1200.50');

    const docInput = screen.getByLabelText(/Doc\. Receptor/i);
    await user.type(docInput, '30112233445');

    const nameInput = screen.getByLabelText(/Cliente \/ Razón Social/i);
    await user.type(nameInput, 'Cliente Corp');

    await user.click(screen.getByRole('button', { name: 'Guardar factura' }));

    await waitFor(() => {
      expect(onSuccess).toHaveBeenCalled();
    });
  });

  it('PurchaseInvoiceForm fills all fields and submits successfully', async () => {
    const user = userEvent.setup();
    const qc = createQueryClient();
    const onSuccess = vi.fn();

    server.use(
      http.post('/api/v1/me/purchase-invoices', () => {
        return HttpResponse.json({ id: 'pinv_1' });
      })
    );

    render(
      <Providers client={qc}>
        <PurchaseInvoiceForm onSuccess={onSuccess} />
      </Providers>
    );

    await user.click(screen.getByRole('button', { name: /Cargar factura de compra/i }));

    const typeSelect = screen.getByLabelText('Tipo');
    fireEvent.change(typeSelect, { target: { value: 'B' } });

    const posInput = screen.getByLabelText('Punto de venta');
    await user.clear(posInput);
    await user.type(posInput, '3');

    const numInput = screen.getByLabelText('Número');
    await user.clear(numInput);
    await user.type(numInput, '45');

    const netInput = screen.getByLabelText(/Neto Gravado/i);
    await user.type(netInput, '1000');

    const vatInput = screen.getByLabelText(/^IVA/i);
    await user.type(vatInput, '210');

    const docInput = screen.getByLabelText(/CUIT Proveedor/i);
    await user.type(docInput, '20998877665');

    const nameInput = screen.getByLabelText(/Proveedor \/ Razón Social/i);
    await user.type(nameInput, 'Proveedor SA');

    await user.click(screen.getByRole('button', { name: 'Guardar compra' }));

    await waitFor(() => {
      expect(onSuccess).toHaveBeenCalled();
    });
  });

  it('ExternalIncomeStep selects No and saves', async () => {
    const user = userEvent.setup();
    const qc = createQueryClient();
    let savedPayload: unknown = null;

    server.use(
      http.put('/api/v1/me/periods/:period/external-income', async ({ request }) => {
        savedPayload = await request.json();
        return HttpResponse.json({ success: true, declaredNoExternalIncome: true, externalIncome: [] });
      })
    );

    render(
      <Providers client={qc}>
        <ExternalIncomeStep period="2026-10" periodMonthName="octubre 2026" answered={false} />
      </Providers>
    );

    await user.click(screen.getByRole('button', { name: /No, solo en Wallbit/i }));

    await waitFor(() => {
      expect(savedPayload).toEqual({
        declaredNoExternalIncome: true,
        externalIncome: [],
      });
    });
  });
});
