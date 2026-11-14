import { describe, it, expect } from 'vitest';
import { render, screen, waitFor, fireEvent, act } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { RouterProvider, createMemoryRouter } from 'react-router-dom';
import { server } from '../../../mocks/server';
import { ParametersPage } from '../ParametersPage';

function renderParametersPage(initialEntries = ['/contador/parametros']) {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  });

  const routes = [
    {
      path: '/contador/parametros',
      element: (
        <QueryClientProvider client={queryClient}>
          <ParametersPage />
        </QueryClientProvider>
      ),
    },
    {
      path: '/contador',
      element: <div>Página de Cierres</div>,
    },
  ];

  const router = createMemoryRouter(routes, { initialEntries });

  return {
    ...render(<RouterProvider router={router} />),
    router,
  };
}

describe('ParametersPage', () => {
  it('renderiza el título, las secciones y los valores iniciales con botón Guardar deshabilitado', async () => {
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

    // Botón Guardar deshabilitado por defecto (sin cambios)
    const saveButton = screen.getByRole('button', { name: /^guardar$/i });
    expect(saveButton).toBeDisabled();
  });

  it('el botón Guardar se habilita al editar y se vuelve a deshabilitar si se restaura el valor original', async () => {
    const user = userEvent.setup();
    renderParametersPage();

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'Parámetros' })).toBeInTheDocument();
    });

    const categoryCapInput = screen.getByLabelText('Alerta de tope de categoría (%)');
    const saveButton = screen.getByRole('button', { name: /^guardar$/i });
    expect(saveButton).toBeDisabled();

    // 1. Modificar valor -> botón habilitado
    await user.clear(categoryCapInput);
    await user.type(categoryCapInput, '85');
    expect(saveButton).not.toBeDisabled();

    // 2. Restaurar valor original -> botón deshabilitado nuevamente
    await user.clear(categoryCapInput);
    await user.type(categoryCapInput, '80');
    expect(saveButton).toBeDisabled();
  });

  it('permite modificar parámetros, guardarlos con éxito y muestra toast flotante', async () => {
    const user = userEvent.setup();
    renderParametersPage();

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'Parámetros' })).toBeInTheDocument();
    });

    const categoryCapInput = screen.getByLabelText('Alerta de tope de categoría (%)');
    await user.clear(categoryCapInput);
    await user.type(categoryCapInput, '85');

    const saveButton = screen.getByRole('button', { name: /^guardar$/i });
    expect(saveButton).not.toBeDisabled();
    await user.click(saveButton);

    const toast = await screen.findByRole('status');
    expect(toast).toBeInTheDocument();
    expect(toast).toHaveTextContent('Parámetros guardados. Se recalcularon 42 cierres.');

    // Tras guardar con éxito, no hay cambios pendientes por lo que Guardar vuelve a deshabilitarse
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /^guardar$/i })).toBeDisabled();
    });

    // Cerrar toast manualmente
    const closeToastBtn = screen.getByRole('button', { name: 'Cerrar notificación' });
    await user.click(closeToastBtn);
    expect(screen.queryByRole('status')).not.toBeInTheDocument();
  });

  it('sin cambios permite navegar libremente sin mostrar modal', async () => {
    const { router } = renderParametersPage();

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'Parámetros' })).toBeInTheDocument();
    });

    act(() => {
      router.navigate('/contador');
    });

    expect(screen.queryByText('¿Salir sin guardar?')).not.toBeInTheDocument();
    expect(screen.getByText('Página de Cierres')).toBeInTheDocument();
  });

  it('con cambios bloquea la navegación y el botón Seguir editando cierra el modal y se queda', async () => {
    const user = userEvent.setup();
    const { router } = renderParametersPage();

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'Parámetros' })).toBeInTheDocument();
    });

    const categoryCapInput = screen.getByLabelText('Alerta de tope de categoría (%)');
    await user.clear(categoryCapInput);
    await user.type(categoryCapInput, '88');

    act(() => {
      router.navigate('/contador');
    });

    expect(await screen.findByText('¿Salir sin guardar?')).toBeInTheDocument();
    expect(screen.getByText('Tenés cambios sin guardar en los parámetros.')).toBeInTheDocument();

    const keepEditingBtn = screen.getByRole('button', { name: 'Seguir editando' });
    await user.click(keepEditingBtn);

    await waitFor(() => {
      expect(screen.queryByText('¿Salir sin guardar?')).not.toBeInTheDocument();
    });
    expect(router.state.location.pathname).toBe('/contador/parametros');
  });

  it('con cambios bloquea la navegación y el botón Salir sin guardar descarta y navega', async () => {
    const user = userEvent.setup();
    const { router } = renderParametersPage();

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'Parámetros' })).toBeInTheDocument();
    });

    const categoryCapInput = screen.getByLabelText('Alerta de tope de categoría (%)');
    await user.clear(categoryCapInput);
    await user.type(categoryCapInput, '88');

    act(() => {
      router.navigate('/contador');
    });

    expect(await screen.findByText('¿Salir sin guardar?')).toBeInTheDocument();

    const discardBtn = screen.getByRole('button', { name: 'Salir sin guardar' });
    await user.click(discardBtn);

    await waitFor(() => {
      expect(screen.getByText('Página de Cierres')).toBeInTheDocument();
    });
    expect(router.state.location.pathname).toBe('/contador');
  });

  it('con cambios bloquea la navegación y el botón Guardar y salir guarda y navega', async () => {
    const user = userEvent.setup();
    const { router } = renderParametersPage();

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'Parámetros' })).toBeInTheDocument();
    });

    const categoryCapInput = screen.getByLabelText('Alerta de tope de categoría (%)');
    await user.clear(categoryCapInput);
    await user.type(categoryCapInput, '92');

    act(() => {
      router.navigate('/contador');
    });

    expect(await screen.findByText('¿Salir sin guardar?')).toBeInTheDocument();

    const saveAndLeaveBtn = screen.getByRole('button', { name: 'Guardar y salir' });
    await user.click(saveAndLeaveBtn);

    await waitFor(() => {
      expect(screen.getByText('Página de Cierres')).toBeInTheDocument();
    });
    expect(router.state.location.pathname).toBe('/contador');
  });

  it('con cambios y error al guardar en Guardar y salir se queda en la pantalla y muestra error', async () => {
    server.use(
      http.put('/api/v1/accountant/parameters', () => {
        return new HttpResponse(null, { status: 500 });
      })
    );

    const user = userEvent.setup();
    const { router } = renderParametersPage();

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'Parámetros' })).toBeInTheDocument();
    });

    const categoryCapInput = screen.getByLabelText('Alerta de tope de categoría (%)');
    await user.clear(categoryCapInput);
    await user.type(categoryCapInput, '99');

    act(() => {
      router.navigate('/contador');
    });

    expect(await screen.findByText('¿Salir sin guardar?')).toBeInTheDocument();
    const saveAndLeaveBtn = screen.getByRole('button', { name: 'Guardar y salir' });
    await user.click(saveAndLeaveBtn);

    expect(await screen.findByRole('alert')).toBeInTheDocument();
    expect(router.state.location.pathname).toBe('/contador/parametros');
  });

  it('permite modificar todos los campos del formulario y activa el aviso de beforeunload', async () => {
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
    fireEvent.change(reminderDaysInput, { target: { value: '5' } });

    const atypicalWindowInput = screen.getByLabelText('Ventana de cobro atípico (meses)');
    fireEvent.change(atypicalWindowInput, { target: { value: '8' } });

    const toleranceInput = screen.getByLabelText('Tolerancia de importe (%)');
    fireEvent.change(toleranceInput, { target: { value: '4.5' } });

    const windowBeforeInput = screen.getByLabelText('Días previos (emisión antes de cobro)');
    fireEvent.change(windowBeforeInput, { target: { value: '30' } });

    const windowAfterInput = screen.getByLabelText('Días posteriores (cobro posterior)');
    fireEvent.change(windowAfterInput, { target: { value: '15' } });

    // Disparar beforeunload cuando está dirty
    const beforeUnloadEvent = new Event('beforeunload', { cancelable: true });
    window.dispatchEvent(beforeUnloadEvent);
    expect(beforeUnloadEvent.defaultPrevented).toBe(true);
  });

  it('muestra errores de validación de campo cuando el backend devuelve 422', async () => {
    const user = userEvent.setup();
    renderParametersPage();

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'Parámetros' })).toBeInTheDocument();
    });

    const windowInput = screen.getByLabelText('Ventana de cobro atípico (meses)');
    fireEvent.change(windowInput, { target: { value: '0' } });

    const saveButton = screen.getByRole('button', { name: /^guardar$/i });
    expect(saveButton).not.toBeDisabled();
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
});
