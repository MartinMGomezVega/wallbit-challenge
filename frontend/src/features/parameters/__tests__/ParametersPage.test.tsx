import { describe, it, expect } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { server } from '../../../mocks/server';
import { ParametersPage } from '../ParametersPage';

function renderParametersPage() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  });

  return render(
    <QueryClientProvider client={queryClient}>
      <ParametersPage />
    </QueryClientProvider>
  );
}

describe('ParametersPage', () => {
  it('renderiza el título, las secciones y los valores iniciales', async () => {
    renderParametersPage();

    expect(screen.getByLabelText('Cargando parámetros')).toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'Parámetros' })).toBeInTheDocument();
    });

    // Secciones
    expect(screen.getByText('Monotributo y alertas de categoría')).toBeInTheDocument();
    expect(screen.getByText('Tipo de cambio (FX)')).toBeInTheDocument();
    expect(screen.getByText('Cobros y facturación')).toBeInTheDocument();
    expect(screen.getByText('Reglas de conciliación automática')).toBeInTheDocument();

    // Valores
    const invoiceECheckbox = screen.getByLabelText(/Computar facturas E en los límites de Monotributo/i) as HTMLInputElement;
    expect(invoiceECheckbox.checked).toBe(true);

    const categoryCapInput = screen.getByLabelText('Alerta de tope de categoría (%)') as HTMLInputElement;
    expect(categoryCapInput.value).toBe('80');

    const kCapInput = screen.getByLabelText('Alerta de tope K de exclusión (%)') as HTMLInputElement;
    expect(kCapInput.value).toBe('80');

    const fxSelect = screen.getByLabelText('Regla de cotización de divisas') as HTMLSelectElement;
    expect(fxSelect.value).toBe('BNA_BUY_PREV_BUSINESS_DAY');

    const minUsdInput = screen.getByLabelText(/Monto mínimo para alerta de facturas faltantes/i) as HTMLInputElement;
    expect(minUsdInput.value).toBe('50.00');

    const atypicalMultInput = screen.getByLabelText('Multiplicador de cobro atípico (ratio)') as HTMLInputElement;
    expect(atypicalMultInput.value).toBe('2.5');

    const atypicalWindowInput = screen.getByLabelText('Ventana de cobro atípico (meses)') as HTMLInputElement;
    expect(atypicalWindowInput.value).toBe('6');

    const cardSpendInput = screen.getByLabelText('Tope de gastos con tarjeta (%)') as HTMLInputElement;
    expect(cardSpendInput.value).toBe('40');

    const reminderDaysInput = screen.getByLabelText('Días de aviso previo a vencimiento (días)') as HTMLInputElement;
    expect(reminderDaysInput.value).toBe('3');

    const toleranceInput = screen.getByLabelText('Tolerancia de importe (%)') as HTMLInputElement;
    expect(toleranceInput.value).toBe('3.0');

    const windowBeforeInput = screen.getByLabelText('Días previos (emisión antes de cobro)') as HTMLInputElement;
    expect(windowBeforeInput.value).toBe('45');

    const windowAfterInput = screen.getByLabelText('Días posteriores (cobro posterior)') as HTMLInputElement;
    expect(windowAfterInput.value).toBe('10');
  });

  it('permite modificar parámetros y guardarlos con éxito', async () => {
    const user = userEvent.setup();
    renderParametersPage();

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'Parámetros' })).toBeInTheDocument();
    });

    const categoryCapInput = screen.getByLabelText('Alerta de tope de categoría (%)');
    await user.clear(categoryCapInput);
    await user.type(categoryCapInput, '85');

    const saveButton = screen.getByRole('button', { name: /Guardar parámetros/i });
    await user.click(saveButton);

    await waitFor(() => {
      expect(
        screen.getByText('Parámetros guardados. Se recalcularon 42 cierres.')
      ).toBeInTheDocument();
    });
  });

  it('muestra errores de validación de campo cuando el backend devuelve 422', async () => {
    const user = userEvent.setup();
    renderParametersPage();

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'Parámetros' })).toBeInTheDocument();
    });

    const windowInput = screen.getByLabelText('Ventana de cobro atípico (meses)');
    fireEvent.change(windowInput, { target: { value: '0' } });

    const saveButton = screen.getByRole('button', { name: /Guardar parámetros/i });
    await user.click(saveButton);

    await waitFor(() => {
      expect(screen.getByText('debe ser mayor a 0')).toBeInTheDocument();
    });
  });

  it('muestra mensaje de error cuando falla la carga inicial', async () => {
    server.use(
      http.get('/api/v1/accountant/parameters', () => {
        return new HttpResponse(null, { status: 500 });
      })
    );

    renderParametersPage();

    await waitFor(() => {
      expect(screen.getByText('Error al cargar los parámetros')).toBeInTheDocument();
    });
  });

  it('permite modificar todos los campos del formulario', async () => {
    const user = userEvent.setup();
    renderParametersPage();

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'Parámetros' })).toBeInTheDocument();
    });

    const invoiceECheckbox = screen.getByLabelText(/Computar facturas E en los límites de Monotributo/i);
    await user.click(invoiceECheckbox);

    const kCapInput = screen.getByLabelText('Alerta de tope K de exclusión (%)');
    fireEvent.change(kCapInput, { target: { value: '90' } });

    const cardSpendInput = screen.getByLabelText('Tope de gastos con tarjeta (%)');
    fireEvent.change(cardSpendInput, { target: { value: '50' } });

    const fxSelect = screen.getByLabelText('Regla de cotización de divisas');
    await user.selectOptions(fxSelect, 'BNA_SELL_PREV_BUSINESS_DAY');

    const minUsdInput = screen.getByLabelText(/Monto mínimo para alerta de facturas faltantes/i);
    fireEvent.change(minUsdInput, { target: { value: '100.00' } });

    const atypicalMultInput = screen.getByLabelText('Multiplicador de cobro atípico (ratio)');
    fireEvent.change(atypicalMultInput, { target: { value: '3.0' } });

    const reminderDaysInput = screen.getByLabelText('Días de aviso previo a vencimiento (días)');
    fireEvent.change(reminderDaysInput, { target: { value: '7' } });

    const toleranceInput = screen.getByLabelText('Tolerancia de importe (%)');
    fireEvent.change(toleranceInput, { target: { value: '2.5' } });

    const windowBeforeInput = screen.getByLabelText('Días previos (emisión antes de cobro)');
    fireEvent.change(windowBeforeInput, { target: { value: '30' } });

    const windowAfterInput = screen.getByLabelText('Días posteriores (cobro posterior)');
    fireEvent.change(windowAfterInput, { target: { value: '20' } });

    const saveButton = screen.getByRole('button', { name: /Guardar parámetros/i });
    await user.click(saveButton);

    await waitFor(() => {
      expect(screen.getByText('Parámetros guardados. Se recalcularon 42 cierres.')).toBeInTheDocument();
    });
  });

  it('muestra errores de validación devueltos por el backend para múltiples campos', async () => {
    server.use(
      http.put('/api/v1/accountant/parameters', () => {
        return HttpResponse.json(
          {
            error: {
              code: 'VALIDATION_ERROR',
              message: 'Parámetros inválidos',
              fields: {
                categoryCapWarningBp: 'tope categoría inválido',
                kCapWarningBp: 'tope K inválido',
                cardSpendKCapBp: 'tope tarjeta inválido',
                fxRule: 'regla inválida',
                missingInvoiceMinUsd: 'monto inválido',
                atypicalMultiplierBp: 'multiplicador inválido',
                atypicalWindowMonths: 'ventana inválida',
                reminderDaysBeforeDue: 'días recordatorio inválidos',
                reconciliationToleranceBp: 'tolerancia inválida',
                reconciliationWindowBeforeDays: 'días previos inválidos',
                reconciliationWindowAfterDays: 'días posteriores inválidos',
              },
            },
          },
          { status: 422 }
        );
      })
    );

    const user = userEvent.setup();
    renderParametersPage();

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'Parámetros' })).toBeInTheDocument();
    });

    const saveButton = screen.getByRole('button', { name: /Guardar parámetros/i });
    await user.click(saveButton);

    await waitFor(() => {
      expect(screen.getByText('tope categoría inválido')).toBeInTheDocument();
      expect(screen.getByText('tope K inválido')).toBeInTheDocument();
      expect(screen.getByText('tope tarjeta inválido')).toBeInTheDocument();
      expect(screen.getByText('regla inválida')).toBeInTheDocument();
      expect(screen.getByText('monto inválido')).toBeInTheDocument();
      expect(screen.getByText('multiplicador inválido')).toBeInTheDocument();
      expect(screen.getByText('ventana inválida')).toBeInTheDocument();
      expect(screen.getByText('días recordatorio inválidos')).toBeInTheDocument();
      expect(screen.getByText('tolerancia inválida')).toBeInTheDocument();
      expect(screen.getByText('días previos inválidos')).toBeInTheDocument();
      expect(screen.getByText('días posteriores inválidos')).toBeInTheDocument();
    });
  });

  it('muestra alerta de error general cuando falla el guardado con error 500', async () => {
    server.use(
      http.put('/api/v1/accountant/parameters', () => {
        return new HttpResponse(null, { status: 500 });
      })
    );

    const user = userEvent.setup();
    renderParametersPage();

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'Parámetros' })).toBeInTheDocument();
    });

    const saveButton = screen.getByRole('button', { name: /Guardar parámetros/i });
    await user.click(saveButton);

    await waitFor(() => {
      expect(screen.getByRole('alert')).toBeInTheDocument();
    });
  });
});
