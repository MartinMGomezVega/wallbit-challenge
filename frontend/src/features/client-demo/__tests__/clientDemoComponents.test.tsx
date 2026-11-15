import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { ConfirmStep } from '../CheckIn/ConfirmStep';
import { ExternalIncomeStep } from '../CheckIn/ExternalIncomeStep';
import { InvoiceFormDialog } from '../CheckIn/InvoiceFormDialog';
import { InvoicesStep } from '../CheckIn/InvoicesStep';
import { ValidatedSummaryCard } from '../ValidatedSummaryCard';
import { PhoneFrame } from '../PhoneFrame';
import { InReview } from '../InReview';
import { DEMO_CLIENTS } from '../demoClients';
import * as clientApi from '../api';
import { ApiError } from '../../../shared/http/http';
import type { ClientClosing, ClientInvoice } from '../schemas';

describe('client-demo isolated components', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  describe('ConfirmStep', () => {
    it('renders error message when present', () => {
      render(
        <ConfirmStep
          canConfirm={true}
          loading={false}
          onConfirm={vi.fn()}
          error="No se pudo confirmar el período"
        />,
      );
      expect(screen.getByRole('alert')).toHaveTextContent('No se pudo confirmar el período');
    });

    it('renders helper text when cannot confirm', () => {
      render(
        <ConfirmStep
          canConfirm={false}
          loading={false}
          onConfirm={vi.fn()}
        />,
      );
      expect(screen.getByText(/Respondé la pregunta sobre cobros externos/i)).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /Listo, ya cargué todo/i })).toBeDisabled();
    });

    it('renders loading state when confirming', () => {
      render(
        <ConfirmStep
          canConfirm={true}
          loading={true}
          onConfirm={vi.fn()}
        />,
      );
      expect(screen.getByRole('button', { name: /Confirmando\.\.\./i })).toBeDisabled();
    });

    it('triggers onConfirm when button is clicked', async () => {
      const onConfirm = vi.fn();
      render(
        <ConfirmStep
          canConfirm={true}
          loading={false}
          onConfirm={onConfirm}
        />,
      );
      const btn = screen.getByRole('button', { name: /Listo, ya cargué todo/i });
      expect(btn).toBeEnabled();
      fireEvent.click(btn);
      expect(onConfirm).toHaveBeenCalledTimes(1);
    });
  });

  describe('ExternalIncomeStep', () => {
    it('initializes with declaredNone=true when initialDeclaredNone is passed', () => {
      const onChange = vi.fn();
      render(
        <ExternalIncomeStep
          initialDeclaredNone={true}
          onChange={onChange}
        />,
      );
      const radioNo = screen.getByLabelText(/No, todos mis cobros fueron a través de Wallbit/i);
      expect(radioNo).toBeChecked();
      expect(onChange).toHaveBeenCalledWith(true, true, []);
    });

    it('allows toggling between No and Yes, and adding/removing item rows', async () => {
      const onChange = vi.fn();
      render(
        <ExternalIncomeStep
          onChange={onChange}
        />,
      );

      // Select 'No'
      const radioNo = screen.getByLabelText(/No, todos mis cobros fueron a través de Wallbit/i);
      fireEvent.click(radioNo);
      expect(onChange).toHaveBeenLastCalledWith(true, true, []);

      // Select 'Sí'
      const radioYes = screen.getByLabelText(/Sí, cobré ingresos por fuera/i);
      fireEvent.click(radioYes);

      // Initially empty fields mean invalid
      expect(onChange).toHaveBeenLastCalledWith(false, false, expect.any(Array));

      // Fill in amount and description
      const amountInput = screen.getByLabelText(/Monto en dólares cobro 1/i);
      const descInput = screen.getByLabelText(/Descripción cobro 1/i);

      fireEvent.change(amountInput, { target: { value: '250.50' } });
      fireEvent.change(descInput, { target: { value: 'Asesoría a cliente' } });

      expect(onChange).toHaveBeenLastCalledWith(true, false, [
        { amount: '250.50', description: 'Asesoría a cliente' },
      ]);

      // Add another item
      const addBtn = screen.getByRole('button', { name: /\+ Agregar otro cobro/i });
      fireEvent.click(addBtn);

      // Second row appears
      expect(screen.getByLabelText(/Monto en dólares cobro 2/i)).toBeInTheDocument();
      expect(onChange).toHaveBeenLastCalledWith(false, false, expect.any(Array));

      // Remove row 2
      const removeBtn = screen.getByLabelText(/Eliminar cobro 2/i);
      fireEvent.click(removeBtn);
      expect(screen.queryByLabelText(/Monto en dólares cobro 2/i)).not.toBeInTheDocument();
      expect(onChange).toHaveBeenLastCalledWith(true, false, [
        { amount: '250.50', description: 'Asesoría a cliente' },
      ]);
    });

    it('initializes with initialItems when passed', () => {
      const onChange = vi.fn();
      render(
        <ExternalIncomeStep
          initialItems={[{ amount: '500', description: 'Consulting' }]}
          onChange={onChange}
        />,
      );
      const radioYes = screen.getByLabelText(/Sí, cobré ingresos por fuera/i);
      expect(radioYes).toBeChecked();
      expect(screen.getByDisplayValue('500')).toBeInTheDocument();
      expect(screen.getByDisplayValue('Consulting')).toBeInTheDocument();
    });
  });

  describe('InvoiceFormDialog', () => {
    it('renders with RI options and handles currency toggle and validation', async () => {
      const onSuccess = vi.fn();
      const onClose = vi.fn();
      const user = userEvent.setup();

      vi.spyOn(clientApi, 'createClientInvoice').mockResolvedValue({
        id: 'inv_10',
        clientId: 'cli_demo_09',
        type: 'A',
        pointOfSale: 3,
        number: 45,
        issueDate: '2026-10-18',
        currency: 'ARS',
        netAmount: { amount: '150000.00', currency: 'ARS' },
        vatAmount: { amount: '0.00', currency: 'ARS' },
        totalAmount: { amount: '150000.00', currency: 'ARS' },
        receiverName: 'Empresa SA',
      });

      render(
        <InvoiceFormDialog
          isOpen={true}
          clientId="cli_demo_09"
          clientRegime="RI"
          onClose={onClose}
          onSuccess={onSuccess}
        />,
      );

      expect(screen.getByRole('heading', { name: /Alta de comprobante/i })).toBeInTheDocument();

      // Check RI options
      const typeSelect = screen.getByLabelText('Tipo');
      fireEvent.change(typeSelect, { target: { value: 'A' } });

      // Change currency to ARS
      const currencySelect = screen.getByLabelText('Moneda');
      fireEvent.change(currencySelect, { target: { value: 'ARS' } });
      expect(currencySelect).toHaveValue('ARS');

      // Fill in valid inputs
      fireEvent.change(screen.getByLabelText('Punto de venta'), { target: { value: '00003' } });
      fireEvent.change(screen.getByLabelText('Número'), { target: { value: '00000045' } });
      fireEvent.change(screen.getByLabelText('Fecha'), { target: { value: '2026-10-18' } });
      fireEvent.change(screen.getByLabelText('Cliente (denominación)'), { target: { value: 'Empresa SA' } });
      fireEvent.change(screen.getByLabelText('Monto total'), { target: { value: '150000.00' } });

      // Click Guardar / Cargar
      const submitBtn = screen.getByRole('button', { name: /Cargar comprobante/i });
      await user.click(submitBtn);

      expect(clientApi.createClientInvoice).toHaveBeenCalledWith(
        'cli_demo_09',
        expect.objectContaining({
          type: 'A',
          pointOfSale: 3,
          number: 45,
          currency: 'ARS',
          receiverName: 'Empresa SA',
        }),
      );
      expect(onSuccess).toHaveBeenCalled();
      expect(onClose).toHaveBeenCalled();
    });

    it('muestra errores locales de validación si faltan campos obligatorios', async () => {
      render(
        <InvoiceFormDialog
          isOpen={true}
          clientId="cli_demo_15"
          clientRegime="MONOTRIBUTO"
          onClose={vi.fn()}
          onSuccess={vi.fn()}
        />,
      );

      // Clear default values
      fireEvent.change(screen.getByLabelText('Fecha'), { target: { value: '' } });
      fireEvent.change(screen.getByLabelText('Punto de venta'), { target: { value: '0' } });
      fireEvent.change(screen.getByLabelText('Número'), { target: { value: '' } });
      fireEvent.change(screen.getByLabelText('Cliente (denominación)'), { target: { value: '' } });
      fireEvent.change(screen.getByLabelText('Monto total'), { target: { value: '' } });

      const form = screen.getByRole('dialog').querySelector('form')!;
      fireEvent.submit(form);

      expect(screen.getByText('La fecha de emisión es requerida')).toBeInTheDocument();
      expect(screen.getByText('El punto de venta debe ser mayor a 0')).toBeInTheDocument();
      expect(screen.getByText('El número debe ser mayor a 0')).toBeInTheDocument();
      expect(screen.getByText('La denominación del receptor es requerida')).toBeInTheDocument();
      expect(screen.getByText('El importe total debe ser mayor a 0')).toBeInTheDocument();
    });

    it('muestra errores de validación devueltos por el backend', async () => {
      const apiErr = new ApiError(422, 'VALIDATION_ERROR', 'Error de validación', {
        number: 'Número de factura ya utilizado',
      });
      vi.spyOn(clientApi, 'createClientInvoice').mockRejectedValueOnce(apiErr);

      render(
        <InvoiceFormDialog
          isOpen={true}
          clientId="cli_demo_15"
          clientRegime="MONOTRIBUTO"
          onClose={vi.fn()}
          onSuccess={vi.fn()}
        />,
      );

      fireEvent.change(screen.getByLabelText('Punto de venta'), { target: { value: '1' } });
      fireEvent.change(screen.getByLabelText('Número'), { target: { value: '10' } });
      fireEvent.change(screen.getByLabelText('Cliente (denominación)'), { target: { value: 'Receptor' } });
      fireEvent.change(screen.getByLabelText('Monto total'), { target: { value: '100' } });

      const form = screen.getByRole('dialog').querySelector('form')!;
      fireEvent.submit(form);

      expect(await screen.findByText('Número de factura ya utilizado')).toBeInTheDocument();
    });

    it('does not render when isOpen is false', () => {
      const { container } = render(
        <InvoiceFormDialog
          isOpen={false}
          clientId="cli_demo_15"
          clientRegime="MONOTRIBUTO"
          onClose={vi.fn()}
          onSuccess={vi.fn()}
        />,
      );
      expect(container.firstChild).toBeNull();
    });
  });

  describe('InvoicesStep', () => {
    it('renders empty invoice list state and handles CSV upload success and error', async () => {
      const onRefresh = vi.fn();
      vi.spyOn(clientApi, 'importClientInvoicesCSV').mockResolvedValueOnce({
        imported: 2,
        skippedDuplicates: 1,
        warnings: [],
        errors: [],
      });

      render(
        <InvoicesStep
          clientId="cli_demo_15"
          clientRegime="MONOTRIBUTO"
          invoices={[]}
          onRefresh={onRefresh}
        />,
      );

      expect(screen.getByText('No hay comprobantes cargados para este período.')).toBeInTheDocument();

      // Click "Alta de comprobante" opens dialog
      const addBtn = screen.getByRole('button', { name: /Alta de comprobante/i });
      fireEvent.click(addBtn);
      expect(screen.getByRole('heading', { name: /Alta de comprobante/i })).toBeInTheDocument();

      // Close dialog
      fireEvent.click(screen.getByLabelText('Cerrar formulario'));

      // Simulate file upload success
      const file = new File(['test'], 'comprobantes.csv', { type: 'text/csv' });
      const input = screen.getByLabelText(/Seleccionar archivo CSV de Mis Comprobantes/i);
      fireEvent.change(input, { target: { files: [file] } });

      await waitFor(() => {
        expect(screen.getByRole('status')).toHaveTextContent(/Se importaron 2 comprobantes/i);
      });
      expect(onRefresh).toHaveBeenCalled();

      // Simulate file upload error
      vi.spyOn(clientApi, 'importClientInvoicesCSV').mockRejectedValueOnce(
        new Error('Error al importar archivo CSV'),
      );
      fireEvent.change(input, { target: { files: [file] } });

      await waitFor(() => {
        expect(screen.getByRole('alert')).toHaveTextContent('Error al importar archivo CSV');
      });
    });

    it('formats invoice row with fallback if issueDate is missing or malformed', () => {
      const invoices: ClientInvoice[] = [
        {
          id: 'inv_1',
          clientId: 'cli_demo_15',
          type: 'C',
          pointOfSale: 1,
          number: 1,
          issueDate: '',
          currency: 'ARS',
          netAmount: { amount: '500.00', currency: 'ARS' },
          vatAmount: { amount: '0.00', currency: 'ARS' },
          totalAmount: { amount: '500.00', currency: 'ARS' },
          receiverName: 'Consumidor Final',
        },
      ];

      render(
        <InvoicesStep
          clientId="cli_demo_15"
          clientRegime="MONOTRIBUTO"
          invoices={invoices}
          onRefresh={vi.fn()}
        />,
      );

      expect(screen.getByText('Fact. C')).toBeInTheDocument();
      expect(screen.getByText('0001-00000001')).toBeInTheDocument();
      expect(screen.getByText('-')).toBeInTheDocument();
    });
  });

  describe('ValidatedSummaryCard', () => {
    const baseClosing: ClientClosing = {
      closingId: 'cls_test',
      period: '2026-10',
      state: 'READY',
      input: {
        answered: true,
        confirmed: true,
        invoicesCount: 4,
      },
      published: {
        version: 1,
        title: 'Cierre de octubre 2026',
        publishedAt: '2026-11-03T10:00:00Z',
        validatedBy: {
          name: 'C.P. Mariana Quiroga',
          license: 'Matrícula de ejemplo',
        },
        regime: 'MONOTRIBUTO',
        category: 'H',
        categoryCapARS: { amount: 400000000, currency: 'ARS' },
        pctOfCategoryCapBp: 9062, // 90.62% -> should trigger orange progress bar
        billed12mARS: { amount: 362500000, currency: 'ARS' },
        figures: {
          billedUSD: { amount: 250000, currency: 'USD' },
          billedARS: { amount: 362500000, currency: 'ARS' },
          total: { amount: 6611100, currency: 'ARS' },
          items: [
            {
              concept: 'Cuota Monotributo categoría H',
              amount: { amount: 6611100, currency: 'ARS' },
              dueDate: '2026-11-20',
            },
          ],
        },
        body: 'Revisión finalizada sin inconsistencias.',
      },
    };

    it('renders Monotributo progress bar in warning state (80-99%)', () => {
      render(
        <ValidatedSummaryCard
          closing={baseClosing}
          clientId="cli_demo_01"
        />,
      );

      expect(screen.getAllByText(/Categoría H/i)[0]).toBeInTheDocument();
      const progressBar = screen.getByRole('progressbar');
      expect(progressBar.firstElementChild).toHaveStyle({ backgroundColor: '#EA580C' });
    });

    it('renders Monotributo progress bar in danger state (>= 100%)', () => {
      const redClosing: ClientClosing = {
        ...baseClosing,
        published: {
          ...baseClosing.published!,
          pctOfCategoryCapBp: 10500, // 105%
        },
      };

      render(
        <ValidatedSummaryCard
          closing={redClosing}
          clientId="cli_demo_01"
        />,
      );

      const progressBar = screen.getByRole('progressbar');
      expect(progressBar.firstElementChild).toHaveStyle({ backgroundColor: '#EF4444' });
    });

    it('renders RI figures and tax breakdown when regime is RI', () => {
      const riClosing: ClientClosing = {
        closingId: 'cls_ri',
        period: '2026-10',
        state: 'READY',
        input: {
          answered: true,
          confirmed: true,
          invoicesCount: 2,
        },
        paidAt: '2026-11-10T12:00:00Z',
        published: {
          version: 1,
          title: 'Cierre de octubre 2026',
          publishedAt: '2026-11-03T10:00:00Z',
          validatedBy: {
            name: 'C.P. Mariana Quiroga',
            license: 'CPCECABA T° 120 F° 45',
          },
          regime: 'RI',
          figures: {
            billedUSD: { amount: 500000, currency: 'USD' },
            billedARS: { amount: 725000000, currency: 'ARS' },
            total: { amount: 15225000, currency: 'ARS' },
            items: [
              {
                concept: 'IVA DJ Mensual F.2002',
                amount: { amount: 15225000, currency: 'ARS' },
                dueDate: '2026-11-18',
              },
            ],
          },
          body: 'Liquidación de IVA finalizada.',
        },
      };

      render(
        <ValidatedSummaryCard
          closing={riClosing}
          clientId="cli_demo_09"
        />,
      );

      expect(screen.getByText('Responsable Inscripto')).toBeInTheDocument();
      expect(screen.getByTestId('paid-badge')).toBeInTheDocument();
    });

    it('handles onMarkPaid click and failure state with ApiError', async () => {
      vi.spyOn(clientApi, 'markClientClosingPaid').mockRejectedValue(
        new ApiError(400, 'BAD_REQUEST', 'El pago no se pudo procesar'),
      );
      render(
        <ValidatedSummaryCard
          closing={baseClosing}
          clientId="cli_demo_01"
        />,
      );

      const payBtn = screen.getByRole('button', { name: /Ya pagué/i });
      fireEvent.click(payBtn);

      await waitFor(() => {
        expect(screen.getByText('El pago no se pudo procesar')).toBeInTheDocument();
      });
    });

    it('handles onMarkPaid click and failure state', async () => {
      vi.spyOn(clientApi, 'markClientClosingPaid').mockRejectedValue(
        new Error('Error al registrar el pago'),
      );
      render(
        <ValidatedSummaryCard
          closing={baseClosing}
          clientId="cli_demo_01"
        />,
      );

      const payBtn = screen.getByRole('button', { name: /Ya pagué/i });
      fireEvent.click(payBtn);

      await waitFor(() => {
        expect(screen.getByText('Error al registrar el pago')).toBeInTheDocument();
      });
    });
  });

  describe('PhoneFrame and InReview', () => {
    it('PhoneFrame renders client name and demo notice', () => {
      render(
        <PhoneFrame clientName="Paula Rossi">
          <div>Contenido móvil</div>
        </PhoneFrame>,
      );
      expect(screen.getByText('Paula Rossi')).toBeInTheDocument();
      expect(screen.getByText('Contenido móvil')).toBeInTheDocument();
      expect(screen.getByText('Así se vería dentro de la app de Wallbit')).toBeInTheDocument();
    });

    it('InReview renders waiting notice and previous month card if available', () => {
      const octClosing: ClientClosing = {
        closingId: 'cls_oct',
        period: '2026-10',
        state: 'IN_REVIEW',
        input: {
          answered: true,
          confirmed: true,
          invoicesCount: 4,
        },
      };

      const prevClosing: ClientClosing = {
        closingId: 'cls_prev',
        period: '2026-09',
        state: 'READY',
        input: {
          answered: true,
          confirmed: true,
          invoicesCount: 3,
        },
        paidAt: '2026-10-05T10:00:00Z',
        published: {
          version: 1,
          title: 'Cierre de septiembre 2026',
          publishedAt: '2026-10-03T10:00:00Z',
          validatedBy: {
            name: 'C.P. Mariana Quiroga',
            license: 'Matrícula de ejemplo',
          },
          figures: {
            billedUSD: { amount: 100000, currency: 'USD' },
            billedARS: { amount: 145000000, currency: 'ARS' },
            total: { amount: 0, currency: 'ARS' },
            items: [],
          },
          body: 'Septiembre validado',
        },
      };

      render(
        <InReview
          octoberClosing={octClosing}
          previousClosing={prevClosing}
          clientId="cli_demo_02"
        />,
      );

      expect(screen.getByText(/Tu contadora está revisando tu cierre de octubre/i)).toBeInTheDocument();
      expect(screen.getByText(/Mes anterior \(Septiembre\)/i)).toBeInTheDocument();
      expect(screen.getByText('Septiembre validado')).toBeInTheDocument();
    });

    it('InReview renders previous closing label with raw period when not 2026-09', () => {
      const octClosing: ClientClosing = {
        closingId: 'cls_oct',
        period: '2026-10',
        state: 'IN_REVIEW',
        input: { answered: true, confirmed: true, invoicesCount: 1 },
      };
      const prevClosing: ClientClosing = {
        closingId: 'cls_prev',
        period: '2026-08',
        state: 'READY',
        input: { answered: true, confirmed: true, invoicesCount: 1 },
        published: {
          version: 1,
          title: 'Cierre de agosto 2026',
          publishedAt: '2026-09-01T10:00:00Z',
          figures: {
            billedUSD: { amount: 100000, currency: 'USD' },
            billedARS: { amount: 145000000, currency: 'ARS' },
            total: { amount: 0, currency: 'ARS' },
            items: [],
          },
          body: 'Agosto validado',
        },
      };

      render(
        <InReview
          octoberClosing={octClosing}
          previousClosing={prevClosing}
          clientId="cli_demo_02"
        />,
      );

      expect(screen.getByText(/Mes anterior \(2026-08\)/i)).toBeInTheDocument();
    });
  });

  describe('demoClients list', () => {
    it('defines 5 demo clients with required properties', () => {
      expect(DEMO_CLIENTS).toHaveLength(5);
      for (const client of DEMO_CLIENTS) {
        expect(client.id).toBeTruthy();
        expect(client.name).toBeTruthy();
        expect(client.regime).toMatch(/Monotributo|Responsable Inscripto/);
        expect(client.scenario).toBeTruthy();
        expect(client.shortScenario).toBeTruthy();
      }
    });
  });
});
