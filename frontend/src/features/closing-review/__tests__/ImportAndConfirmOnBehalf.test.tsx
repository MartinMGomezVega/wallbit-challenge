import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor, cleanup, fireEvent } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { axe } from 'vitest-axe';
import { ImportResultView } from '../import/ImportResultView';
import { ImportInvoicesDialog } from '../import/ImportInvoicesDialog';
import { ConfirmOnBehalfForm } from '../confirm/ConfirmOnBehalfForm';
import { InvoicesSection } from '../sections/InvoicesSection';
import { HistorySection } from '../sections/HistorySection';
import { ClosingDetailDTO } from '../schemas';
import * as api from '../api';

function renderWithClient(ui: React.ReactElement) {
  const qc = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });
  return {
    qc,
    ...render(<QueryClientProvider client={qc}>{ui}</QueryClientProvider>),
  };
}

const baseDetail: ClosingDetailDTO = {
  closing: {
    id: 'clo_test_1',
    clientId: 'cli_test_1',
    period: '2026-10',
    status: 'READY_FOR_REVIEW',
    version: 1,
    resultHash: 'hash-12345',
    severity: 'GREEN',
    description: 'Todo ok',
    signals: [],
    mono: {
      billedInPeriodUSD: { amount: 50000, currency: 'USD' },
      billedInPeriodARS: { amount: 75000000, currency: 'ARS' },
      currentCategory: 'B',
      billed12mARS: { amount: 900000000, currency: 'ARS' },
      pctOfKCapBp: 700,
      fee: { amount: 500000, currency: 'ARS' },
      feeDueDate: '2026-11-20T00:00:00Z',
    },
    ri: null,
    computedAt: '2026-11-01T03:00:00Z',
    paidAt: null,
    approvedAt: null,
    approvedBy: null,
  },
  client: {
    id: 'cli_test_1',
    name: 'Paula Rojas',
    cuit: '27-33444444-4',
    regime: 'MONOTRIBUTO',
    monotributoCategory: 'B',
  },
  periodMovements: [],
  periodInvoices: [
    {
      invoice: {
        id: 'inv_1',
        clientId: 'cli_test_1',
        pointOfSale: 1,
        number: 42,
        type: 'C',
        concept: 'SERVICIOS',
        cae: '12345678901234',
        caeDueDate: '2026-10-20',
        issueDate: '2026-10-10T00:00:00Z',
        totalAmount: { amount: 50000, currency: 'ARS' },
        currency: 'ARS',
        uploadedBy: 'ACCOUNTANT',
      },
      amountArs: { amount: 50000, currency: 'ARS' },
      fxRateUsed: 150000,
    },
    {
      invoice: {
        id: 'inv_2',
        clientId: 'cli_test_1',
        pointOfSale: 1,
        number: 43,
        type: 'C',
        concept: 'SERVICIOS',
        cae: '12345678901235',
        caeDueDate: '2026-10-25',
        issueDate: '2026-10-15T00:00:00Z',
        totalAmount: { amount: 30000, currency: 'ARS' },
        currency: 'ARS',
        uploadedBy: 'CLIENT',
      },
      amountArs: { amount: 30000, currency: 'ARS' },
      fxRateUsed: 150000,
    },
  ],
  reconciliation: {
    matchedPairs: [],
    unmatchedIncome: [],
    unmatchedInvoices: [],
  },
  clientInput: {
    clientId: 'cli_test_1',
    period: '2026-10',
    declaredNoExternalIncome: true,
    confirmedAt: '2026-11-05T10:00:00Z',
    confirmedBy: 'ACCOUNTANT',
    confirmationNote: 'Me lo confirmó por mail el 05/11',
  },
};

describe('ImportResultView', () => {
  it('renderiza resumen de importación exitosa', () => {
    render(
      <ImportResultView
        result={{
          imported: 3,
          skippedDuplicates: 1,
          warnings: [{ line: 4, message: 'Receptor sin CUIT' }],
          errors: [],
        }}
      />
    );

    expect(screen.getByText(/se importaron 3 facturas correctamente/i)).toBeInTheDocument();
    expect(screen.getByText('3')).toBeInTheDocument();
    expect(screen.getByText('1')).toBeInTheDocument();
    expect(screen.getByText(/línea 4:/i)).toBeInTheDocument();
    expect(screen.getByText(/receptor sin cuit/i)).toBeInTheDocument();
  });

  it('renderiza caso de 1 sola factura importada y advertencia general sin línea', () => {
    render(
      <ImportResultView
        result={{
          imported: 1,
          skippedDuplicates: 0,
          warnings: [{ line: 0, message: 'Advertencia global' }],
          errors: [],
        }}
      />
    );

    expect(screen.getByText(/se importaron 1 factura correctamente/i)).toBeInTheDocument();
    expect(screen.getByText('Advertencia global')).toBeInTheDocument();
  });

  it('renderiza mensaje cuando solo hubo duplicadas omitidas', () => {
    render(
      <ImportResultView
        result={{
          imported: 0,
          skippedDuplicates: 2,
          warnings: [],
          errors: [],
        }}
      />
    );

    expect(screen.getByText(/no se importaron facturas nuevas \(ya existían en el sistema\)/i)).toBeInTheDocument();
  });

  it('renderiza mensaje cuando no se procesaron facturas', () => {
    render(
      <ImportResultView
        result={{
          imported: 0,
          skippedDuplicates: 0,
          warnings: [],
          errors: [],
        }}
      />
    );

    expect(screen.getByText(/no se procesaron facturas/i)).toBeInTheDocument();
  });

  it('renderiza errores cuando fallan líneas', () => {
    render(
      <ImportResultView
        result={{
          imported: 0,
          skippedDuplicates: 0,
          warnings: [],
          errors: [{ line: 5, message: 'Fecha inválida' }],
        }}
      />
    );

    expect(screen.getByText(/ocurrieron errores al procesar el archivo/i)).toBeInTheDocument();
    expect(screen.getByText(/fecha inválida/i)).toBeInTheDocument();
  });
});

describe('ImportInvoicesDialog', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('muestra aviso si el cierre ya fue aprobado o notificado', () => {
    renderWithClient(
      <ImportInvoicesDialog
        open={true}
        onOpenChange={vi.fn()}
        clientId="cli_1"
        closingId="clo_1"
        closingStatus="APPROVED"
      />
    );

    expect(
      screen.getByText(/este cierre ya fue aprobado\. si importás facturas, vuelve a «para revisar»/i)
    ).toBeInTheDocument();
  });

  it('no muestra aviso si el cierre está en estado para revisar', () => {
    renderWithClient(
      <ImportInvoicesDialog
        open={true}
        onOpenChange={vi.fn()}
        clientId="cli_1"
        closingId="clo_1"
        closingStatus="READY_FOR_REVIEW"
      />
    );

    expect(
      screen.queryByText(/este cierre ya fue aprobado/i)
    ).not.toBeInTheDocument();
  });

  it('incluye enlace para descargar CSV de ejemplo', () => {
    renderWithClient(
      <ImportInvoicesDialog
        open={true}
        onOpenChange={vi.fn()}
        clientId="cli_1"
        closingId="clo_1"
        closingStatus="READY_FOR_REVIEW"
      />
    );

    const link = screen.getByRole('link', { name: /descargar csv de ejemplo/i });
    expect(link).toHaveAttribute('href', '/samples/mis_comprobantes_ok.csv');
    expect(link).toHaveAttribute('download');
  });

  afterEach(() => {
    cleanup();
  });

  it('valida que el archivo sea CSV y no supere los 2 MB', async () => {
    const { unmount } = renderWithClient(
      <ImportInvoicesDialog
        open={true}
        onOpenChange={vi.fn()}
        clientId="cli_1"
        closingId="clo_1"
        closingStatus="READY_FOR_REVIEW"
      />
    );

    const fileInput = document.body.querySelector('input[type="file"]') as HTMLInputElement;

    // Archivo no CSV
    const txtFile = new File(['dummy'], 'facturas.txt', { type: 'text/plain' });
    fireEvent.change(fileInput, { target: { files: [txtFile] } });
    expect(screen.getByText(/el archivo debe tener formato csv/i)).toBeInTheDocument();

    // Archivo mayor a 2 MB
    const largeContent = new Uint8Array(2 * 1024 * 1024 + 10);
    const bigFile = new File([largeContent], 'grande.csv', { type: 'text/csv' });
    fireEvent.change(fileInput, { target: { files: [bigFile] } });
    expect(screen.getByText(/supera el tamaño máximo permitido de 2 mb/i)).toBeInTheDocument();

    unmount();
  });

  it('importa archivo exitosamente e invoca callback al cerrar', async () => {
    const user = userEvent.setup();
    const onImportFinished = vi.fn();
    const onOpenChange = vi.fn();

    vi.spyOn(api, 'importInvoicesAsAccountant').mockResolvedValueOnce({
      imported: 2,
      skippedDuplicates: 0,
      warnings: [],
      errors: [],
    });

    renderWithClient(
      <ImportInvoicesDialog
        open={true}
        onOpenChange={onOpenChange}
        clientId="cli_1"
        closingId="clo_1"
        closingStatus="READY_FOR_REVIEW"
        onImportFinished={onImportFinished}
      />
    );

    const fileInput = document.body.querySelector('input[type="file"]') as HTMLInputElement;
    const validFile = new File(['col1,col2\nval1,val2'], 'mis_comprobantes.csv', { type: 'text/csv' });
    await user.upload(fileInput, validFile);

    const importBtn = screen.getByRole('button', { name: /^importar$/i });
    expect(importBtn).not.toBeDisabled();
    await user.click(importBtn);

    await waitFor(() => {
      expect(screen.getByText(/se importaron 2 facturas correctamente/i)).toBeInTheDocument();
    });

    const readyBtn = screen.getByRole('button', { name: /listo/i });
    await user.click(readyBtn);

    expect(onImportFinished).toHaveBeenCalledWith(2);
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });

  it('soporta drag and drop y click en zona de carga', async () => {
    renderWithClient(
      <ImportInvoicesDialog
        open={true}
        onOpenChange={vi.fn()}
        clientId="cli_1"
        closingId="clo_1"
        closingStatus="READY_FOR_REVIEW"
      />
    );

    const dropZone = screen.getByText(/arrastrá tu archivo csv acá/i).closest('div')!;
    fireEvent.dragOver(dropZone);
    fireEvent.dragLeave(dropZone);

    const file = new File(['csv data'], 'drop.csv', { type: 'text/csv' });
    fireEvent.drop(dropZone, { dataTransfer: { files: [file] } });

    expect(screen.getByText('drop.csv')).toBeInTheDocument();
  });

  it('muestra mensaje de error cuando falla la importación', async () => {
    const user = userEvent.setup();
    vi.spyOn(api, 'importInvoicesAsAccountant').mockRejectedValueOnce(new Error('CSV mal formado en línea 2'));

    renderWithClient(
      <ImportInvoicesDialog
        open={true}
        onOpenChange={vi.fn()}
        clientId="cli_1"
        closingId="clo_1"
        closingStatus="READY_FOR_REVIEW"
      />
    );

    const fileInput = document.body.querySelector('input[type="file"]') as HTMLInputElement;
    const validFile = new File(['bad,csv'], 'bad.csv', { type: 'text/csv' });
    fireEvent.change(fileInput, { target: { files: [validFile] } });

    const importBtn = screen.getByRole('button', { name: /^importar$/i });
    await user.click(importBtn);

    await waitFor(() => {
      expect(screen.getByText(/csv mal formado en línea 2/i)).toBeInTheDocument();
    });
  });

  it('cumple accesibilidad (sin violaciones axe)', async () => {
    const { container } = renderWithClient(
      <ImportInvoicesDialog
        open={true}
        onOpenChange={vi.fn()}
        clientId="cli_1"
        closingId="clo_1"
        closingStatus="APPROVED"
      />
    );

    const results = await axe(container);
    expect(results.violations).toHaveLength(0);
  });
});

describe('ConfirmOnBehalfForm', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  afterEach(() => {
    cleanup();
  });

  it('confirma con declaredNone = true y nota obligatoria', async () => {
    const user = userEvent.setup();
    const onSuccess = vi.fn();

    const spy = vi.spyOn(api, 'confirmInputOnBehalf').mockResolvedValueOnce({
      clientId: 'cli_1',
      period: '2026-10',
      declaredNoExternalIncome: true,
      confirmedBy: 'ACCOUNTANT',
      confirmationNote: 'Confirmado por WhatsApp',
    });

    renderWithClient(
      <ConfirmOnBehalfForm
        clientId="cli_1"
        period="2026-10"
        closingId="clo_1"
        onSuccess={onSuccess}
      />
    );

    expect(screen.getByText(/confirmar en nombre del cliente/i)).toBeInTheDocument();
    const noteInput = screen.getByLabelText(/¿cómo te lo confirmó el cliente\?/i);

    await user.type(noteInput, 'Confirmado por WhatsApp');
    const submitBtn = screen.getByRole('button', { name: /confirmar período/i });
    await user.click(submitBtn);

    await waitFor(() => {
      expect(spy).toHaveBeenCalledWith('cli_1', '2026-10', {
        externalIncome: { none: true },
        note: 'Confirmado por WhatsApp',
      });
      expect(onSuccess).toHaveBeenCalled();
    });
  });

  it('permite ingresar cobros externos si se selecciona "Tuvo cobros"', async () => {
    const user = userEvent.setup();

    const spy = vi.spyOn(api, 'confirmInputOnBehalf').mockResolvedValueOnce({
      clientId: 'cli_1',
      period: '2026-10',
      declaredNoExternalIncome: false,
      confirmedBy: 'ACCOUNTANT',
      confirmationNote: 'Mandó resumen bancario',
    });

    renderWithClient(
      <ConfirmOnBehalfForm
        clientId="cli_1"
        period="2026-10"
        closingId="clo_1"
      />
    );

    const radioHadIncome = screen.getByRole('radio', { name: /^tuvo cobros fuera de wallbit$/i });
    await user.click(radioHadIncome);

    expect(screen.getByPlaceholderText(/monto \(usd\)/i)).toBeInTheDocument();
    const amountInput = screen.getByPlaceholderText(/monto \(usd\)/i);
    const descInput = screen.getByPlaceholderText(/descripción/i);
    const noteInput = screen.getByLabelText(/¿cómo te lo confirmó el cliente\?/i);

    await user.type(amountInput, '150.50');
    await user.type(descInput, 'Asesoría exterior');
    await user.type(noteInput, 'Mandó resumen bancario');

    const submitBtn = screen.getByRole('button', { name: /confirmar período/i });
    await user.click(submitBtn);

    await waitFor(() => {
      expect(spy).toHaveBeenCalledWith('cli_1', '2026-10', {
        externalIncome: {
          items: [{ amount: { amount: 15050, currency: 'USD' }, description: 'Asesoría exterior' }],
        },
        note: 'Mandó resumen bancario',
      });
    });
  });

  it('permite agregar y eliminar filas de cobros externos', async () => {
    const user = userEvent.setup();
    renderWithClient(
      <ConfirmOnBehalfForm
        clientId="cli_1"
        period="2026-10"
        closingId="clo_1"
      />
    );

    const radioHadIncome = screen.getByRole('radio', { name: /^tuvo cobros fuera de wallbit$/i });
    await user.click(radioHadIncome);

    const addBtn = screen.getByRole('button', { name: /agregar otro cobro/i });
    await user.click(addBtn);

    const deleteButtons = screen.getAllByLabelText(/eliminar cobro/i);
    expect(deleteButtons.length).toBe(2);

    await user.click(deleteButtons[0]!);
    expect(screen.queryByLabelText(/eliminar cobro/i)).not.toBeInTheDocument();
  });

  it('muestra error de validación cuando el monto de un cobro es inválido', async () => {
    const user = userEvent.setup();
    renderWithClient(
      <ConfirmOnBehalfForm
        clientId="cli_1"
        period="2026-10"
        closingId="clo_1"
      />
    );

    await user.click(screen.getByRole('radio', { name: /^tuvo cobros fuera de wallbit$/i }));
    await user.type(screen.getByLabelText(/¿cómo te lo confirmó el cliente\?/i), 'Por teléfono');
    await user.type(screen.getByPlaceholderText(/monto \(usd\)/i), '-50');
    await user.type(screen.getByPlaceholderText(/descripción/i), 'Venta');

    await user.click(screen.getByRole('button', { name: /confirmar período/i }));
    expect(screen.getByText(/todos los cobros deben tener un monto mayor a cero/i)).toBeInTheDocument();
  });

  it('muestra error de validación cuando la descripción de un cobro está vacía', async () => {
    const user = userEvent.setup();
    renderWithClient(
      <ConfirmOnBehalfForm
        clientId="cli_1"
        period="2026-10"
        closingId="clo_1"
      />
    );

    await user.click(screen.getByRole('radio', { name: /^tuvo cobros fuera de wallbit$/i }));
    await user.type(screen.getByLabelText(/¿cómo te lo confirmó el cliente\?/i), 'Por teléfono');
    await user.type(screen.getByPlaceholderText(/monto \(usd\)/i), '100');

    await user.click(screen.getByRole('button', { name: /confirmar período/i }));
    expect(screen.getByText(/todos los cobros deben incluir una descripción/i)).toBeInTheDocument();
  });

  it('muestra error si la nota supera 300 caracteres', async () => {
    const user = userEvent.setup();
    renderWithClient(
      <ConfirmOnBehalfForm
        clientId="cli_1"
        period="2026-10"
        closingId="clo_1"
      />
    );

    const noteInput = screen.getByLabelText(/¿cómo te lo confirmó el cliente\?/i);
    noteInput.removeAttribute('maxlength');
    await user.type(noteInput, 'a'.repeat(305));
    await user.click(screen.getByRole('button', { name: /confirmar período/i }));
    expect(screen.getByText(/la nota no puede superar los 300 caracteres/i)).toBeInTheDocument();
  });

  it('muestra error cuando el backend falla al confirmar período', async () => {
    const user = userEvent.setup();
    vi.spyOn(api, 'confirmInputOnBehalf').mockRejectedValueOnce(new Error('Error de servidor al confirmar'));

    renderWithClient(
      <ConfirmOnBehalfForm
        clientId="cli_1"
        period="2026-10"
        closingId="clo_1"
      />
    );

    await user.type(screen.getByLabelText(/¿cómo te lo confirmó el cliente\?/i), 'Confirmado por WhatsApp');
    await user.click(screen.getByRole('button', { name: /confirmar período/i }));

    await waitFor(() => {
      expect(screen.getByText(/error de servidor al confirmar/i)).toBeInTheDocument();
    });
  });

  it('cumple accesibilidad (sin violaciones axe)', async () => {
    const { container } = renderWithClient(
      <ConfirmOnBehalfForm
        clientId="cli_1"
        period="2026-10"
        closingId="clo_1"
      />
    );

    const results = await axe(container);
    expect(results.violations).toHaveLength(0);
  });
});

describe('InvoicesSection and HistorySection integration', () => {
  it('muestra badge "Cargada por el contador" solo en las facturas correspondientes', () => {
    renderWithClient(<InvoicesSection detail={baseDetail} />);

    expect(screen.getByText('Cargada por el contador')).toBeInTheDocument();
    expect(screen.getByText('0001-00000042')).toBeInTheDocument();
    expect(screen.getByText('0001-00000043')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /importar facturas \(csv de arca\)/i })).toBeInTheDocument();
  });

  it('abre el diálogo de importación al hacer click en el botón', async () => {
    const user = userEvent.setup();
    renderWithClient(<InvoicesSection detail={baseDetail} />);

    const openBtn = screen.getByRole('button', { name: /importar facturas \(csv de arca\)/i });
    await user.click(openBtn);

    expect(screen.getByText(/importar facturas de venta \(csv de arca\)/i)).toBeInTheDocument();
  });

  it('muestra ConfirmOnBehalfForm cuando el cierre está PENDING_CLIENT_INPUT', () => {
    const pendingDetail: ClosingDetailDTO = {
      ...baseDetail,
      closing: {
        ...baseDetail.closing,
        status: 'PENDING_CLIENT_INPUT',
      },
    };

    renderWithClient(<InvoicesSection detail={pendingDetail} />);
    expect(screen.getByText(/confirmar en nombre del cliente/i)).toBeInTheDocument();
  });

  it('no muestra ConfirmOnBehalfForm cuando el cierre está READY_FOR_REVIEW', () => {
    renderWithClient(<InvoicesSection detail={baseDetail} />);
    expect(screen.queryByText(/confirmar en nombre del cliente/i)).not.toBeInTheDocument();
  });

  it('HistorySection muestra la evidencia si fue confirmado por el contador', () => {
    renderWithClient(<HistorySection detail={baseDetail} />);
    expect(
      screen.getByText(/período confirmado por el contador:/i)
    ).toBeInTheDocument();
    expect(screen.getByText('Me lo confirmó por mail el 05/11')).toBeInTheDocument();
  });
});
