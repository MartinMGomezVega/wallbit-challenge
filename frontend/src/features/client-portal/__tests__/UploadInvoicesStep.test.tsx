import { describe, it, expect } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import { server } from '../../../mocks/server';
import { UploadInvoicesStep } from '../steps/UploadInvoicesStep';
import { Providers, createQueryClient } from '../../../app/providers';

describe('UploadInvoicesStep & Forms', () => {
  it('renders upload area, sample link, and ComingSoonButton', () => {
    const qc = createQueryClient();
    render(
      <Providers client={qc}>
        <UploadInvoicesStep periodMonthName="octubre 2026" regime="MONOTRIBUTO" invoicesCount={2} />
      </Providers>
    );

    expect(screen.getByText(/1\. Subí tus facturas de octubre 2026/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /emitir factura en arca/i })).toBeInTheDocument();
    expect(screen.getByText('Descargar CSV de ejemplo')).toBeInTheDocument();
    expect(screen.getByText('2 facturas en el período')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /cargar una factura a mano/i })).toBeInTheDocument();
    // In Monotributo, purchase invoice button is not rendered
    expect(screen.queryByRole('button', { name: /cargar factura de compra/i })).not.toBeInTheDocument();
  });

  it('renders purchase invoice button for RI', () => {
    const qc = createQueryClient();
    render(
      <Providers client={qc}>
        <UploadInvoicesStep periodMonthName="octubre 2026" regime="RI" invoicesCount={0} />
      </Providers>
    );

    expect(screen.getByRole('button', { name: /cargar factura de compra/i })).toBeInTheDocument();
  });

  it('handles CSV upload and displays result with imported count', async () => {
    const qc = createQueryClient();
    render(
      <Providers client={qc}>
        <UploadInvoicesStep periodMonthName="octubre 2026" regime="MONOTRIBUTO" invoicesCount={0} />
      </Providers>
    );

    const file = new File(['Fecha,Tipo...'], 'mis_comprobantes.csv', { type: 'text/csv' });
    const input = screen.getByTestId('csv-file-input');

    fireEvent.change(input, { target: { files: [file] } });

    expect(await screen.findByText(/3 comprobantes importados/i)).toBeInTheDocument();
  });

  it('handles CSV upload with warnings and line errors', async () => {
    const qc = createQueryClient();
    server.use(
      http.post('/api/v1/me/invoices/import', () => {
        return HttpResponse.json({
          imported: 1,
          skippedDuplicates: 2,
          warnings: [{ line: 5, message: 'CAE vencido' }],
          errors: [{ line: 8, message: 'Tipo no permitido para Monotributo' }],
        });
      })
    );

    render(
      <Providers client={qc}>
        <UploadInvoicesStep periodMonthName="octubre 2026" regime="MONOTRIBUTO" invoicesCount={0} />
      </Providers>
    );

    const file = new File(['test'], 'test.csv', { type: 'text/csv' });
    const input = screen.getByTestId('csv-file-input');
    fireEvent.change(input, { target: { files: [file] } });

    expect(await screen.findByText(/1 comprobante importado \(2 duplicados omitidos\)/i)).toBeInTheDocument();
    expect(screen.getByText(/Línea 5: CAE vencido/i)).toBeInTheDocument();
    expect(screen.getByText(/Línea 8: Tipo no permitido para Monotributo/i)).toBeInTheDocument();
  });

  it('rejects non-csv files with an error message', async () => {
    const qc = createQueryClient();
    render(
      <Providers client={qc}>
        <UploadInvoicesStep periodMonthName="octubre 2026" regime="MONOTRIBUTO" invoicesCount={0} />
      </Providers>
    );

    const file = new File(['test'], 'factura.pdf', { type: 'application/pdf' });
    const input = screen.getByTestId('csv-file-input');
    fireEvent.change(input, { target: { files: [file] } });

    expect(await screen.findByText(/Por favor seleccioná un archivo CSV válido/i)).toBeInTheDocument();
  });

  it('handles drag and drop events', async () => {
    const qc = createQueryClient();
    render(
      <Providers client={qc}>
        <UploadInvoicesStep periodMonthName="octubre 2026" regime="MONOTRIBUTO" invoicesCount={0} />
      </Providers>
    );

    const dropzone = screen.getByText(/Arrastrá el CSV de Mis Comprobantes acá/i).closest('div')!;

    fireEvent.dragOver(dropzone);
    fireEvent.dragLeave(dropzone);

    const file = new File(['data'], 'test.csv', { type: 'text/csv' });
    fireEvent.drop(dropzone, {
      dataTransfer: { files: [file] },
    });

    expect(await screen.findByText(/3 comprobantes importados/i)).toBeInTheDocument();
  });

  it('allows manual invoice submission through modal', async () => {
    const user = userEvent.setup();
    const qc = createQueryClient();
    render(
      <Providers client={qc}>
        <UploadInvoicesStep periodMonthName="octubre 2026" regime="MONOTRIBUTO" invoicesCount={0} />
      </Providers>
    );

    await user.click(screen.getByRole('button', { name: /cargar una factura a mano/i }));

    expect(await screen.findByRole('heading', { name: 'Cargar factura emitida' })).toBeInTheDocument();

    const amountInput = screen.getByPlaceholderText('1000.00');
    await user.type(amountInput, '1500.50');

    await user.click(screen.getByRole('button', { name: 'Guardar factura' }));

    await waitFor(() => {
      expect(screen.queryByRole('heading', { name: 'Cargar factura emitida' })).not.toBeInTheDocument();
    });
  });

  it('allows purchase invoice submission for RI', async () => {
    const user = userEvent.setup();
    const qc = createQueryClient();
    render(
      <Providers client={qc}>
        <UploadInvoicesStep periodMonthName="octubre 2026" regime="RI" invoicesCount={0} />
      </Providers>
    );

    await user.click(screen.getByRole('button', { name: /cargar factura de compra/i }));

    expect(await screen.findByRole('heading', { name: 'Cargar factura de compra (RI)' })).toBeInTheDocument();

    await user.type(screen.getByPlaceholderText('1000.00'), '2000');
    await user.type(screen.getByPlaceholderText('210.00'), '420');
    await user.type(screen.getByPlaceholderText('30-71122334-5'), '30-99999999-9');
    await user.type(screen.getByPlaceholderText('Distribuidora S.A.'), 'Insumos Tech');

    await user.click(screen.getByRole('button', { name: 'Guardar compra' }));

    await waitFor(() => {
      expect(screen.queryByRole('heading', { name: 'Cargar factura de compra (RI)' })).not.toBeInTheDocument();
    });
  });
});
