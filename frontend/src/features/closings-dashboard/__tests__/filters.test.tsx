import { describe, it, expect, vi, afterEach } from 'vitest';
import { render, screen, act, fireEvent } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { SummaryChips } from '../SummaryChips';
import { SignalMultiSelect } from '../SignalMultiSelect';
import { FiltersBar } from '../FiltersBar';
import { mockCatalog } from '../../../mocks/handlers';
import { DEFAULT_FILTERS } from '../useDashboardParams';

describe('Filters components', () => {
  afterEach(() => {
    vi.useRealTimers();
  });

  const mockSummary = {
    total: 100,
    bySeverity: { RED: 12, YELLOW: 25, GREEN: 63 },
    bySignal: { MISSING_INVOICES: 5, ATYPICAL_INCOME: 8 },
    byStatus: { READY_FOR_REVIEW: 80, PENDING_CLIENT_INPUT: 20 },
  };

  describe('SummaryChips', () => {
    it('muestra los chips de severidad con sus contadores', () => {
      render(
        <SummaryChips
          summary={mockSummary}
          selectedSeverities={[]}
          onToggleSeverity={vi.fn()}
        />
      );

      expect(screen.getByText('Rojos')).toBeInTheDocument();
      expect(screen.getByText('12')).toBeInTheDocument();
      expect(screen.getByText('Amarillos')).toBeInTheDocument();
      expect(screen.getByText('25')).toBeInTheDocument();
      expect(screen.getByText('Verdes')).toBeInTheDocument();
      expect(screen.getByText('63')).toBeInTheDocument();
    });

    it('llama a onToggleSeverity al hacer click en un chip', async () => {
      const onToggle = vi.fn();
      const user = userEvent.setup();

      render(
        <SummaryChips
          summary={mockSummary}
          selectedSeverities={['RED']}
          onToggleSeverity={onToggle}
        />
      );

      const redBtn = screen.getByRole('button', { name: /Rojos/ });
      expect(redBtn).toHaveAttribute('aria-pressed', 'true');

      await user.click(redBtn);
      expect(onToggle).toHaveBeenCalledWith('RED');
    });

    it('llama a onReviewSeverity si se proporciona la acción secundaria', async () => {
      const onReview = vi.fn();
      const user = userEvent.setup();

      render(
        <SummaryChips
          summary={mockSummary}
          selectedSeverities={[]}
          onToggleSeverity={vi.fn()}
          onReviewSeverity={onReview}
        />
      );

      const reviewBtns = screen.getAllByRole('button', { name: 'Revisar' });
      await user.click(reviewBtns[0]!);
      expect(onReview).toHaveBeenCalledWith('RED');
    });
  });

  describe('SignalMultiSelect', () => {
    it('abre el menú desplegable y permite seleccionar situaciones', async () => {
      const onChange = vi.fn();
      const user = userEvent.setup();

      render(
        <SignalMultiSelect
          catalog={mockCatalog}
          bySignalCounts={mockSummary.bySignal}
          selectedSignals={[]}
          onChange={onChange}
        />
      );

      const trigger = screen.getByRole('button', { name: /Situación/ });
      await user.click(trigger);

      expect(screen.getByText('Requiere acción (Rojo)')).toBeInTheDocument();
      const label = screen.getByText('Faltan facturas');
      expect(label).toBeInTheDocument();

      await user.click(label);
      expect(onChange).toHaveBeenCalledWith(['MISSING_INVOICES']);
    });

    it('muestra botón de limpiar cuando hay señales seleccionadas y permite deseleccionar', async () => {
      const onChange = vi.fn();
      const user = userEvent.setup();

      render(
        <SignalMultiSelect
          catalog={mockCatalog}
          bySignalCounts={mockSummary.bySignal}
          selectedSignals={['MISSING_INVOICES']}
          onChange={onChange}
        />
      );

      const trigger = screen.getByRole('button', { name: /Situación/ });
      await user.click(trigger);

      // Botón limpiar
      const clearBtn = screen.getByRole('button', { name: 'Limpiar' });
      await user.click(clearBtn);
      expect(onChange).toHaveBeenCalledWith([]);

      // Deseleccionar al hacer click en el item ya seleccionado
      const label = screen.getByText('Faltan facturas');
      await user.click(label);
      expect(onChange).toHaveBeenCalledWith([]);
    });

    it('cierra el desplegable al hacer click afuera o al presionar Escape', async () => {
      const user = userEvent.setup();

      render(
        <div>
          <button data-testid="outside">Afuera</button>
          <SignalMultiSelect
            catalog={mockCatalog}
            bySignalCounts={mockSummary.bySignal}
            selectedSignals={[]}
            onChange={vi.fn()}
          />
        </div>
      );

      const trigger = screen.getByRole('button', { name: /Situación/ });
      await user.click(trigger);
      expect(screen.getByRole('listbox')).toBeInTheDocument();

      const outsideBtn = screen.getByTestId('outside');
      fireEvent.mouseDown(outsideBtn);
      expect(screen.queryByRole('listbox')).not.toBeInTheDocument();

      // Probar Escape
      await user.click(trigger);
      expect(screen.getByRole('listbox')).toBeInTheDocument();
      fireEvent.keyDown(document, { key: 'Escape' });
      expect(screen.queryByRole('listbox')).not.toBeInTheDocument();
    });
  });

  describe('FiltersBar', () => {
    it('renderiza selector de régimen y llama a onSetRegime', async () => {
      const onSetRegime = vi.fn();
      const user = userEvent.setup();

      render(
        <FiltersBar
          filters={DEFAULT_FILTERS}
          catalog={mockCatalog}
          onSetSignal={vi.fn()}
          onSetRegime={onSetRegime}
          onSetStatus={vi.fn()}
          onSetQ={vi.fn()}
          onClearFilters={vi.fn()}
        />
      );

      const select = screen.getByLabelText('Filtrar por régimen');
      await user.selectOptions(select, 'MONOTRIBUTO');
      expect(onSetRegime).toHaveBeenCalledWith('MONOTRIBUTO');
    });

    it('permite abrir el menú de Estado, alternar y limpiar estados', async () => {
      const onSetStatus = vi.fn();
      const user = userEvent.setup();

      const { rerender } = render(
        <div>
          <button data-testid="outside-status">Afuera</button>
          <FiltersBar
            filters={DEFAULT_FILTERS}
            catalog={mockCatalog}
            summary={mockSummary}
            onSetSignal={vi.fn()}
            onSetRegime={vi.fn()}
            onSetStatus={onSetStatus}
            onSetQ={vi.fn()}
            onClearFilters={vi.fn()}
          />
        </div>
      );

      const statusTrigger = screen.getByRole('button', { name: /Estado/ });
      await user.click(statusTrigger);

      expect(screen.getByRole('listbox')).toBeInTheDocument();
      const option = screen.getByText('Para revisar');
      await user.click(option);
      expect(onSetStatus).toHaveBeenCalledWith(['READY_FOR_REVIEW']);

      // Rerender con estado activo para ver botón Limpiar y badge
      rerender(
        <div>
          <button data-testid="outside-status">Afuera</button>
          <FiltersBar
            filters={{ ...DEFAULT_FILTERS, status: ['READY_FOR_REVIEW'] }}
            catalog={mockCatalog}
            summary={mockSummary}
            onSetSignal={vi.fn()}
            onSetRegime={vi.fn()}
            onSetStatus={onSetStatus}
            onSetQ={vi.fn()}
            onClearFilters={vi.fn()}
          />
        </div>
      );

      const clearBtn = screen.getByRole('button', { name: 'Limpiar' });
      await user.click(clearBtn);
      expect(onSetStatus).toHaveBeenCalledWith([]);

      // Click afuera (mousedown) para cerrar
      const outsideBtn = screen.getByTestId('outside-status');
      fireEvent.mouseDown(outsideBtn);
      expect(screen.queryByRole('listbox')).not.toBeInTheDocument();
    });

    it('maneja búsqueda con debounce y botón de limpiar búsqueda', async () => {
      vi.useFakeTimers();
      const onSetQ = vi.fn();

      render(
        <FiltersBar
          filters={DEFAULT_FILTERS}
          catalog={mockCatalog}
          onSetSignal={vi.fn()}
          onSetRegime={vi.fn()}
          onSetStatus={vi.fn()}
          onSetQ={onSetQ}
          onClearFilters={vi.fn()}
        />
      );

      const input = screen.getByPlaceholderText('Buscar nombre o CUIT');
      fireEvent.change(input, { target: { value: 'Perez' } });

      expect(onSetQ).not.toHaveBeenCalled();

      // Debounce de 300 ms
      act(() => {
        vi.advanceTimersByTime(350);
      });

      expect(onSetQ).toHaveBeenCalledWith('Perez');

      // Botón limpiar búsqueda
      const clearSearchBtn = screen.getByRole('button', { name: 'Limpiar búsqueda' });
      fireEvent.click(clearSearchBtn);
      expect(onSetQ).toHaveBeenCalledWith('');
    });

    it('muestra "Limpiar filtros" cuando hay filtros activos y llama a onClearFilters', async () => {
      const onClear = vi.fn();
      const user = userEvent.setup();

      render(
        <FiltersBar
          filters={{ ...DEFAULT_FILTERS, q: 'Lopez' }}
          catalog={mockCatalog}
          onSetSignal={vi.fn()}
          onSetRegime={vi.fn()}
          onSetStatus={vi.fn()}
          onSetQ={vi.fn()}
          onClearFilters={onClear}
        />
      );

      const clearBtn = screen.getByRole('button', { name: 'Limpiar filtros' });
      await user.click(clearBtn);
      expect(onClear).toHaveBeenCalled();
    });
  });
});
