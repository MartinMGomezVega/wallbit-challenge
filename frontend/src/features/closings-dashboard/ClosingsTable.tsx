import { useMemo, useState } from 'react';
import {
  useReactTable,
  getCoreRowModel,
  flexRender,
  RowSelectionState,
} from '@tanstack/react-table';
import { SearchX } from 'lucide-react';
import { ClosingRow } from './schemas';
import { createColumns } from './columns';
import { Pagination } from '../../shared/ui/Pagination';

export interface ClosingsTableProps {
  data: ClosingRow[];
  total: number;
  page: number;
  pageSize: number;
  sort: 'severity' | 'name' | 'billed' | 'toPay' | 'pctKCap';
  dir: 'asc' | 'desc';
  isLoading?: boolean;
  onSortChange: (field: 'severity' | 'name' | 'billed' | 'toPay' | 'pctKCap') => void;
  onPageChange: (page: number) => void;
  onPageSizeChange: (pageSize: number) => void;
  onReviewRow?: (row: ClosingRow) => void;
  rowSelection?: RowSelectionState;
  onRowSelectionChange?: (updaterOrValue: RowSelectionState | ((old: RowSelectionState) => RowSelectionState)) => void;
  onClearFilters?: () => void;
  highlightedClosingId?: string | null;
  className?: string;
}

export function ClosingsTable({
  data,
  total,
  page,
  pageSize,
  sort,
  dir,
  isLoading = false,
  onSortChange,
  onPageChange,
  onPageSizeChange,
  onReviewRow,
  rowSelection: controlledSelection,
  onRowSelectionChange: setControlledSelection,
  onClearFilters,
  highlightedClosingId,
  className = '',
}: ClosingsTableProps) {
  const [internalSelection, setInternalSelection] = useState<RowSelectionState>({});

  const rowSelection = controlledSelection ?? internalSelection;
  const setRowSelection = setControlledSelection ?? setInternalSelection;

  const columns = useMemo(
    () =>
      createColumns({
        sort,
        dir,
        onSort: onSortChange,
        onReview: onReviewRow,
      }),
    [sort, dir, onSortChange, onReviewRow]
  );

  const table = useReactTable({
    data,
    columns,
    state: {
      rowSelection,
    },
    onRowSelectionChange: setRowSelection,
    getRowId: (row) => row.closingId,
    getCoreRowModel: getCoreRowModel(),
    manualPagination: true,
    manualSorting: true,
    pageCount: Math.ceil(total / pageSize),
  });

  return (
    <div className={`flex flex-col rounded-card border border-app-border bg-app-surface shadow-sm ${className}`}>
      <div className="overflow-x-auto">
        <table className="w-full text-left border-collapse">
          <thead>
            {table.getHeaderGroups().map((headerGroup) => (
              <tr
                key={headerGroup.id}
                className="sticky top-0 z-10 h-10 border-b border-app-border bg-app-surface-muted"
              >
                {headerGroup.headers.map((header) => {
                  const isSorted = sort === header.id;
                  const ariaSort = isSorted
                    ? (dir === 'asc' ? 'ascending' : 'descending')
                    : undefined;
                  return (
                    <th
                      key={header.id}
                      aria-sort={ariaSort}
                      className="px-3 py-2 text-xs font-semibold uppercase tracking-wider text-app-text-muted"
                      style={{ width: header.getSize() !== 150 ? header.getSize() : undefined }}
                    >
                      {header.isPlaceholder
                        ? null
                        : flexRender(header.column.columnDef.header, header.getContext())}
                    </th>
                  );
                })}
              </tr>
            ))}
          </thead>

          <tbody className="divide-y divide-app-border">
            {isLoading ? (
              // Skeleton de carga: 8 filas con shimmer
              Array.from({ length: 8 }).map((_, rIdx) => (
                <tr key={`skeleton-${rIdx}`} className="h-[52px] animate-pulse">
                  <td className="px-3 py-3">
                    <div className="h-4 w-4 rounded bg-app-border/60" />
                  </td>
                  <td className="px-3 py-3">
                    <div className="h-3 w-6 rounded bg-app-border/60" />
                  </td>
                  <td className="px-3 py-3">
                    <div className="h-3 w-48 rounded bg-app-border/60 mb-1" />
                    <div className="h-2.5 w-24 rounded bg-app-border/40" />
                  </td>
                  <td className="px-3 py-3">
                    <div className="h-3 w-32 rounded bg-app-border/60 mb-1" />
                    <div className="h-2.5 w-20 rounded bg-app-border/40" />
                  </td>
                  <td className="px-3 py-3">
                    <div className="h-5 w-16 rounded bg-app-border/60" />
                  </td>
                  <td className="px-3 py-3">
                    <div className="h-3 w-20 rounded bg-app-border/60 ml-auto mb-1" />
                    <div className="h-2.5 w-16 rounded bg-app-border/40 ml-auto" />
                  </td>
                  <td className="px-3 py-3">
                    <div className="h-2 w-24 rounded bg-app-border/60 ml-auto" />
                  </td>
                  <td className="px-3 py-3">
                    <div className="h-3 w-20 rounded bg-app-border/60 ml-auto mb-1" />
                    <div className="h-2.5 w-16 rounded bg-app-border/40 ml-auto" />
                  </td>
                  <td className="px-3 py-3">
                    <div className="h-5 w-24 rounded bg-app-border/60" />
                  </td>
                  <td className="px-3 py-3">
                    <div className="h-7 w-16 rounded bg-app-border/60" />
                  </td>
                </tr>
              ))
            ) : data.length === 0 ? (
              // Estado vacío
              <tr>
                <td colSpan={columns.length} className="py-16 text-center">
                  <div className="flex flex-col items-center justify-center gap-2">
                    <SearchX className="h-8 w-8 text-app-text-muted/60" aria-hidden="true" />
                    <p className="text-sm font-medium text-app-text">
                      No hay clientes con estos filtros
                    </p>
                    {onClearFilters && (
                      <button
                        type="button"
                        onClick={onClearFilters}
                        className="mt-1 text-xs font-semibold text-app-primary hover:underline"
                      >
                        Limpiar filtros
                      </button>
                    )}
                  </div>
                </td>
              </tr>
            ) : (
              // Filas de datos
              table.getRowModel().rows.map((row) => {
                const item = row.original;
                const isRed = item.severity === 'RED';
                const isSelected = row.getIsSelected();
                const isHighlighted = highlightedClosingId === item.closingId;

                return (
                  <tr
                    key={row.id}
                    ref={(el) => {
                      if (isHighlighted && el) {
                        el.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
                      }
                    }}
                    onClick={() => onReviewRow?.(item)}
                    className={`h-[52px] cursor-pointer transition-colors duration-200 hover:bg-app-surface-muted ${
                      isHighlighted
                        ? 'bg-app-primary-soft ring-1 ring-app-primary'
                        : isSelected
                        ? 'bg-app-primary-soft'
                        : isRed
                        ? 'border-l-[3px] border-l-severity-red-icon'
                        : ''
                    }`}
                  >
                    {row.getVisibleCells().map((cell) => (
                      <td key={cell.id} className="px-3 py-2 text-xs">
                        {flexRender(cell.column.columnDef.cell, cell.getContext())}
                      </td>
                    ))}
                  </tr>
                );
              })
            )}
          </tbody>
        </table>
      </div>

      {/* Paginación */}
      {total > 0 && (
        <div className="border-t border-app-border px-4">
          <Pagination
            page={page}
            pageSize={pageSize}
            total={total}
            onPageChange={onPageChange}
            onPageSizeChange={onPageSizeChange}
          />
        </div>
      )}
    </div>
  );
}
