import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {
  SeverityIndicator,
  SignalChip,
  OtherSignalsChip,
  ProgressBar,
  StatusBadge,
  Tooltip,
  Pagination,
} from '../index';

describe('shared/ui components', () => {
  describe('SeverityIndicator', () => {
    it('renderiza semáforo rojo con aria-label correcta', () => {
      render(<SeverityIndicator severity="RED" withTooltip={false} />);
      const el = screen.getByLabelText('Rojo: requiere acción');
      expect(el).toBeInTheDocument();
    });

    it('renderiza semáforo amarillo con aria-label correcta', () => {
      render(<SeverityIndicator severity="YELLOW" withTooltip={false} />);
      const el = screen.getByLabelText('Amarillo: revisar');
      expect(el).toBeInTheDocument();
    });

    it('renderiza semáforo verde con aria-label correcta', () => {
      render(<SeverityIndicator severity="GREEN" withTooltip={false} />);
      const el = screen.getByLabelText('Verde: todo ok');
      expect(el).toBeInTheDocument();
    });

    it('soporta números para severidad (0=GREEN, 1=YELLOW, 2=RED)', () => {
      const { rerender } = render(<SeverityIndicator severity={0} withTooltip={false} />);
      expect(screen.getByLabelText('Verde: todo ok')).toBeInTheDocument();

      rerender(<SeverityIndicator severity={1} withTooltip={false} />);
      expect(screen.getByLabelText('Amarillo: revisar')).toBeInTheDocument();

      rerender(<SeverityIndicator severity={2} withTooltip={false} />);
      expect(screen.getByLabelText('Rojo: requiere acción')).toBeInTheDocument();
    });

    it('renderiza modo badge con texto corto visible', () => {
      render(<SeverityIndicator severity="RED" mode="badge" />);
      expect(screen.getByText('Requiere acción')).toBeInTheDocument();
    });

    it('renderiza modo icon y modo dot', () => {
      const { rerender } = render(<SeverityIndicator severity="GREEN" mode="icon" withTooltip={false} />);
      expect(screen.getByLabelText('Verde: todo ok')).toBeInTheDocument();

      rerender(<SeverityIndicator severity="GREEN" mode="dot" withTooltip={false} />);
      expect(screen.getByLabelText('Verde: todo ok')).toBeInTheDocument();
    });
  });

  describe('SignalChip', () => {
    it('renderiza el chip de señal principal con su label', () => {
      render(
        <SignalChip
          signal={{ code: 'MONO_NEAR_K_CAP', label: 'Cerca del tope K', severity: 'YELLOW' }}
        />
      );
      expect(screen.getByText('Cerca del tope K')).toBeInTheDocument();
    });

    it('renderiza chip con severidad por defecto cuando es undefined', () => {
      render(
        <SignalChip
          signal={{ code: 'DEFAULT_SEV', label: 'Sin severidad' }}
        />
      );
      expect(screen.getByText('Sin severidad')).toBeInTheDocument();
    });

    it('OtherSignalsChip renderiza chip +N con aria-label y tooltip con colores y fallback a code', () => {
      const signals = [
        { code: 'EXCLUSION', label: 'Excede tope', severity: 'RED' as const },
        { code: 'ATYPICAL_INCOME', label: 'Cobro atípico', severity: 'YELLOW' as const },
        { code: 'ALL_GOOD', label: 'Todo bien', severity: 'GREEN' as const },
        { code: 'CODE_ONLY', label: '', severity: undefined },
      ];
      render(<OtherSignalsChip signals={signals} />);
      expect(screen.getByText('+4')).toBeInTheDocument();
      expect(screen.getByLabelText('4 otras señales')).toBeInTheDocument();
    });

    it('OtherSignalsChip devuelve null si no hay señales adicionales', () => {
      const { container } = render(<OtherSignalsChip signals={[]} />);
      expect(container.firstChild).toBeNull();
    });
  });

  describe('ProgressBar', () => {
    it('renderiza barra de progreso con role y atributos aria', () => {
      render(<ProgressBar pctBp={7500} />);
      const bar = screen.getByRole('progressbar');
      expect(bar).toBeInTheDocument();
      expect(bar).toHaveAttribute('aria-valuenow', '75');
      expect(screen.getByText('75%')).toBeInTheDocument();
    });

    it('renderiza porcentaje >= 80% (naranja) y negativos', () => {
      const { rerender } = render(<ProgressBar pctBp={8500} />);
      const bar = screen.getByRole('progressbar');
      expect(bar).toHaveAttribute('aria-valuenow', '85');
      const innerBar = bar.querySelector('.bg-progress-warning');
      expect(innerBar).toBeInTheDocument();

      rerender(<ProgressBar pctBp={-500} />);
      expect(bar).toHaveAttribute('aria-valuenow', '0');
    });

    it('renderiza porcentaje >= 100% correctamente', () => {
      render(<ProgressBar pctBp={10500} arsAmount={{ amount: 15000000000, currency: 'ARS' }} />);
      const bar = screen.getByRole('progressbar');
      expect(bar).toHaveAttribute('aria-valuenow', '105');
      expect(screen.getByText('105%')).toBeInTheDocument();
    });
  });

  describe('StatusBadge', () => {
    it('renderiza todos los estados de cierre según especificación', () => {
      const { rerender } = render(<StatusBadge status="PENDING_CLIENT_INPUT" />);
      expect(screen.getByText('Esperando al cliente')).toBeInTheDocument();

      rerender(<StatusBadge status="READY_FOR_REVIEW" />);
      expect(screen.getByText('Para revisar')).toBeInTheDocument();

      rerender(<StatusBadge status="APPROVED" />);
      expect(screen.getByText('Aprobado')).toBeInTheDocument();

      rerender(<StatusBadge status="NOTIFIED" />);
      expect(screen.getByText('Notificado')).toBeInTheDocument();
    });

    it('muestra fecha de pago cuando paidAt está presente', () => {
      render(<StatusBadge status="NOTIFIED" paidAt="2026-11-20" />);
      expect(screen.getByText('Notificado')).toBeInTheDocument();
      expect(screen.getByText('Pagado 20/11')).toBeInTheDocument();
    });

    it('maneja un estado desconocido con fallback', () => {
      // @ts-expect-error probando estado desconocido
      render(<StatusBadge status="UNKNOWN_STATUS" />);
      expect(screen.getByText('UNKNOWN_STATUS')).toBeInTheDocument();
    });
  });

  describe('Tooltip', () => {
    it('renderiza sus hijos correctamente', () => {
      render(
        <Tooltip content="Información extra">
          <button>Hover me</button>
        </Tooltip>
      );
      expect(screen.getByText('Hover me')).toBeInTheDocument();
    });

    it('renderiza solo hijos si no hay content', () => {
      render(<Tooltip content={null}><button>No tooltip</button></Tooltip>);
      expect(screen.getByText('No tooltip')).toBeInTheDocument();
    });
  });

  describe('Pagination', () => {
    it('muestra el rango de elementos y total', () => {
      render(
        <Pagination
          page={1}
          pageSize={50}
          total={145}
          onPageChange={vi.fn()}
        />
      );
      expect(screen.getByText(/Mostrando/)).toBeInTheDocument();
      expect(screen.getByRole('button', { name: 'Página 1' })).toBeInTheDocument();
      expect(screen.getByText('145')).toBeInTheDocument();
      expect(screen.getByLabelText('Paginación de resultados')).toBeInTheDocument();
    });

    it('llama a onPageChange al hacer click en página o siguiente/anterior', async () => {
      const onPageChange = vi.fn();
      const user = userEvent.setup();

      render(
        <Pagination
          page={2}
          pageSize={50}
          total={150}
          onPageChange={onPageChange}
        />
      );

      await user.click(screen.getByLabelText('Página anterior'));
      expect(onPageChange).toHaveBeenCalledWith(1);

      await user.click(screen.getByLabelText('Página siguiente'));
      expect(onPageChange).toHaveBeenCalledWith(3);

      await user.click(screen.getByLabelText('Página 3'));
      expect(onPageChange).toHaveBeenCalledWith(3);
    });

    it('llama a onPageSizeChange al cambiar selector', async () => {
      const onPageSizeChange = vi.fn();
      const user = userEvent.setup();

      render(
        <Pagination
          page={1}
          pageSize={50}
          total={300}
          onPageChange={vi.fn()}
          onPageSizeChange={onPageSizeChange}
        />
      );

      const select = screen.getByLabelText('Resultados por página');
      await user.selectOptions(select, '100');
      expect(onPageSizeChange).toHaveBeenCalledWith(100);
    });
  });
});
