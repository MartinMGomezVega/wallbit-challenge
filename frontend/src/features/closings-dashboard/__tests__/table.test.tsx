import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { ClosingsTable } from '../ClosingsTable';
import { SelectionBar } from '../SelectionBar';
import { mockClosings } from '../../../mocks/handlers';

describe('ClosingsTable & SelectionBar', () => {
  const sampleData = mockClosings.slice(0, 3); // 1 RED, 1 RED, 1 YELLOW

  describe('ClosingsTable', () => {
    it('renderiza encabezados y filas con datos formateados', () => {
      render(
        <ClosingsTable
          data={sampleData}
          total={3}
          page={1}
          pageSize={50}
          sort="severity"
          dir="asc"
          onSortChange={vi.fn()}
          onPageChange={vi.fn()}
          onPageSizeChange={vi.fn()}
        />
      );

      // Encabezados
      expect(screen.getByText('Situación')).toBeInTheDocument();
      expect(screen.getByText('Cliente')).toBeInTheDocument();
      expect(screen.getByText('Régimen')).toBeInTheDocument();
      expect(screen.getByText('Facturado')).toBeInTheDocument();
      expect(screen.getByText('Acum. 12 meses')).toBeInTheDocument();
      expect(screen.getByText('A pagar')).toBeInTheDocument();
      expect(screen.getByText('Estado')).toBeInTheDocument();

      // Contenido de filas
      expect(screen.getByText('Ignacio Sosa')).toBeInTheDocument();
      expect(screen.getByText('Carla Benítez')).toBeInTheDocument();
      expect(screen.getByText('Valentina Ruiz')).toBeInTheDocument();
      expect(screen.getByText('20-28222221-1')).toBeInTheDocument();
      expect(screen.getByText('Monotributo K')).toBeInTheDocument();
    });

    it('llama a onSortChange al hacer click en un encabezado ordenable', async () => {
      const onSortChange = vi.fn();
      const user = userEvent.setup();

      render(
        <ClosingsTable
          data={sampleData}
          total={3}
          page={1}
          pageSize={50}
          sort="severity"
          dir="asc"
          onSortChange={onSortChange}
          onPageChange={vi.fn()}
          onPageSizeChange={vi.fn()}
        />
      );

      const clientHeader = screen.getByRole('button', { name: /Cliente/ });
      await user.click(clientHeader);
      expect(onSortChange).toHaveBeenCalledWith('name');

      const regimeHeader = screen.getByRole('button', { name: /Régimen/ });
      await user.click(regimeHeader);
      expect(onSortChange).toHaveBeenCalledWith('regime');

      const statusHeader = screen.getByRole('button', { name: /Estado/ });
      await user.click(statusHeader);
      expect(onSortChange).toHaveBeenCalledWith('status');
    });

    it('abre el modal desde el botón Revisar y desde el nombre del cliente, pero NO al hacer click en la fila', async () => {
      const onReviewRow = vi.fn();
      const user = userEvent.setup();

      render(
        <ClosingsTable
          data={sampleData}
          total={3}
          page={1}
          pageSize={50}
          sort="severity"
          dir="asc"
          onSortChange={vi.fn()}
          onPageChange={vi.fn()}
          onPageSizeChange={vi.fn()}
          onReviewRow={onReviewRow}
        />
      );

      // 1. Click en el texto de CUIT (dentro de la celda de cliente pero no el link del nombre)
      const cuitText = screen.getByText('20-28222221-1');
      await user.click(cuitText);
      expect(onReviewRow).not.toHaveBeenCalled();

      // 2. Click en el botón "Revisar" de la primera fila
      const reviewButtons = screen.getAllByRole('button', { name: 'Revisar' });
      await user.click(reviewButtons[0]!);
      expect(onReviewRow).toHaveBeenCalledWith(sampleData[0]);

      // 3. Click en el nombre del cliente (que actúa como link hacia el modal)
      onReviewRow.mockClear();
      const clientLink = screen.getByRole('button', { name: 'Ignacio Sosa' });
      await user.click(clientLink);
      expect(onReviewRow).toHaveBeenCalledWith(sampleData[0]);
    });

    it('renderiza números grandes completos sin truncamiento ni ellipsis con whitespace-nowrap', () => {
      const largeNumberData = [
        {
          ...sampleData[0]!,
          closingId: 'large-num-1',
          billedInPeriod: {
            usd: { amount: 1250000, currency: 'USD' as const },
            ars: { amount: 161444604, currency: 'ARS' as const },
          },
          toPay: {
            amount: { amount: 161444604, currency: 'ARS' as const },
            dueDate: '2026-11-20T00:00:00Z',
          },
        },
      ];

      render(
        <ClosingsTable
          data={largeNumberData}
          total={1}
          page={1}
          pageSize={50}
          sort="severity"
          dir="asc"
          onSortChange={vi.fn()}
          onPageChange={vi.fn()}
          onPageSizeChange={vi.fn()}
        />
      );

      // El número grande aparece completo en el DOM
      const billedArs = screen.getAllByText(/\$ 1\.614\.446,04/);
      expect(billedArs.length).toBeGreaterThan(0);
      billedArs.forEach((el) => {
        expect(el).toHaveClass('whitespace-nowrap');
        expect(el).toHaveClass('tabular-nums');
      });
    });

    it('muestra skeleton de carga cuando isLoading es true', () => {
      const { container } = render(
        <ClosingsTable
          data={[]}
          total={0}
          page={1}
          pageSize={50}
          sort="severity"
          dir="asc"
          isLoading={true}
          onSortChange={vi.fn()}
          onPageChange={vi.fn()}
          onPageSizeChange={vi.fn()}
        />
      );

      const pulsingRows = container.querySelectorAll('.animate-pulse');
      expect(pulsingRows.length).toBe(8);
    });

    it('muestra estado vacío cuando no hay datos', async () => {
      const onClearFilters = vi.fn();
      const user = userEvent.setup();

      render(
        <ClosingsTable
          data={[]}
          total={0}
          page={1}
          pageSize={50}
          sort="severity"
          dir="asc"
          isLoading={false}
          onSortChange={vi.fn()}
          onPageChange={vi.fn()}
          onPageSizeChange={vi.fn()}
          onClearFilters={onClearFilters}
        />
      );

      expect(screen.getByText('No hay clientes con estos filtros')).toBeInTheDocument();
      const clearBtn = screen.getByRole('button', { name: 'Limpiar filtros' });
      await user.click(clearBtn);
      expect(onClearFilters).toHaveBeenCalled();
    });

    it('permite seleccionar filas con checkboxes', async () => {
      const onSelectionChange = vi.fn();
      const user = userEvent.setup();

      render(
        <ClosingsTable
          data={sampleData}
          total={3}
          page={1}
          pageSize={50}
          sort="severity"
          dir="asc"
          onSortChange={vi.fn()}
          onPageChange={vi.fn()}
          onPageSizeChange={vi.fn()}
          onRowSelectionChange={onSelectionChange}
        />
      );

      const selectAll = screen.getByLabelText('Seleccionar todas las filas de la página');
      await user.click(selectAll);
      expect(onSelectionChange).toHaveBeenCalled();
    });
  });

  describe('SelectionBar', () => {
    it('no renderiza nada cuando selectedCount es 0', () => {
      const { container } = render(
        <SelectionBar
          selectedCount={0}
          greenCount={0}
          onClearSelection={vi.fn()}
        />
      );
      expect(container.firstChild).toBeNull();
    });

    it('renderiza barra flotante cuando hay clientes seleccionados', async () => {
      const onClear = vi.fn();
      const onApproveGreens = vi.fn();
      const onReviewSelected = vi.fn();
      const user = userEvent.setup();

      render(
        <SelectionBar
          selectedCount={3}
          greenCount={2}
          onApproveGreens={onApproveGreens}
          onReviewSelected={onReviewSelected}
          onClearSelection={onClear}
        />
      );

      expect(screen.getByText('3 seleccionados')).toBeInTheDocument();

      const approveBtn = screen.getByRole('button', { name: /Aprobar verdes \(2\)/ });
      expect(approveBtn).not.toBeDisabled();
      await user.click(approveBtn);
      expect(onApproveGreens).toHaveBeenCalled();

      const reviewBtn = screen.getByRole('button', { name: 'Revisar seleccionados' });
      await user.click(reviewBtn);
      expect(onReviewSelected).toHaveBeenCalled();

      const clearBtn = screen.getByLabelText('Deseleccionar todos');
      await user.click(clearBtn);
      expect(onClear).toHaveBeenCalled();
    });

    it('deshabilita botón de aprobar cuando greenCount es 0', () => {
      render(
        <SelectionBar
          selectedCount={2}
          greenCount={0}
          onClearSelection={vi.fn()}
        />
      );

      const approveBtn = screen.getByRole('button', { name: /Aprobar verdes \(0\)/ });
      expect(approveBtn).toBeDisabled();
    });
  });
});
