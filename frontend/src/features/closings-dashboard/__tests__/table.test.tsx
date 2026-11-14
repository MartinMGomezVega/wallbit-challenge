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
      expect(screen.getByText('Rég.')).toBeInTheDocument();
      expect(screen.getByText('Facturado')).toBeInTheDocument();
      expect(screen.getByText('Acum. 12m')).toBeInTheDocument();
      expect(screen.getByText('A pagar')).toBeInTheDocument();
      expect(screen.getByText('Estado')).toBeInTheDocument();

      // Contenido de filas
      expect(screen.getByText('Ignacio Sosa')).toBeInTheDocument();
      expect(screen.getByText('Carla Benítez')).toBeInTheDocument();
      expect(screen.getByText('Valentina Ruiz')).toBeInTheDocument();
      expect(screen.getByText('20-28222221-1')).toBeInTheDocument();
      expect(screen.getByText('Mono K')).toBeInTheDocument();
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
    });

    it('llama a onReviewRow al hacer click en una fila o botón Revisar', async () => {
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

      const reviewButtons = screen.getAllByRole('button', { name: 'Revisar' });
      await user.click(reviewButtons[0]!);
      expect(onReviewRow).toHaveBeenCalledWith(sampleData[0]);
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
