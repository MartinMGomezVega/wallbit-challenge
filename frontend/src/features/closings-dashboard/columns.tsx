import { ColumnDef } from '@tanstack/react-table';
import { ArrowUpDown, ArrowUp, ArrowDown } from 'lucide-react';
import { ClosingRow } from './schemas';
import { SeverityIndicator } from '../../shared/ui/SeverityIndicator';
import { SignalChip, OtherSignalsChip } from '../../shared/ui/SignalChip';
import { ProgressBar } from '../../shared/ui/ProgressBar';
import { StatusBadge } from '../../shared/ui/StatusBadge';
import { Tooltip } from '../../shared/ui/Tooltip';
import { formatMoney } from '../../shared/format/money';
import { formatDate } from '../../shared/format/date';
import { SortField } from './useDashboardParams';

export interface ColumnOptions {
  sort?: SortField;
  dir?: 'asc' | 'desc';
  onSort?: (sortField: SortField) => void;
  onReview?: (row: ClosingRow) => void;
}

export function createColumns(options: ColumnOptions = {}): ColumnDef<ClosingRow>[] {
  const { sort, dir, onSort, onReview } = options;

  const renderSortHeader = (
    label: string,
    field: SortField,
    align: 'left' | 'right' = 'left'
  ) => {
    const isSorted = sort === field;
    const ariaSort = isSorted ? (dir === 'asc' ? 'ascending' : 'descending') : 'none';

    return (
      <button
        type="button"
        onClick={() => onSort?.(field)}
        aria-sort={ariaSort}
        className={`group inline-flex items-center gap-1 text-xs font-semibold uppercase tracking-wider text-app-text-muted hover:text-app-text focus:outline-none focus:ring-1 focus:ring-app-focus rounded-[3px] ${
          align === 'right' ? 'flex-row-reverse' : ''
        }`}
      >
        <span>{label}</span>
        {isSorted ? (
          dir === 'asc' ? (
            <ArrowUp className="h-3.5 w-3.5 text-app-primary" aria-hidden="true" />
          ) : (
            <ArrowDown className="h-3.5 w-3.5 text-app-primary" aria-hidden="true" />
          )
        ) : (
          <ArrowUpDown
            className="h-3.5 w-3.5 opacity-0 group-hover:opacity-60 transition-opacity"
            aria-hidden="true"
          />
        )}
      </button>
    );
  };

  return [
    // 1. Checkbox de selección (sticky col 1)
    {
      id: 'select',
      header: ({ table }) => (
        <div className="flex items-center justify-center pl-2">
          <input
            type="checkbox"
            checked={table.getIsAllPageRowsSelected()}
            ref={(input) => {
              if (input) {
                input.indeterminate = table.getIsSomePageRowsSelected();
              }
            }}
            onChange={(e) => table.toggleAllPageRowsSelected(e.target.checked)}
            aria-label="Seleccionar todas las filas de la página"
            className="h-4 w-4 rounded-[4px] border-app-border text-app-primary focus:ring-app-focus"
          />
        </div>
      ),
      cell: ({ row }) => (
        <div
          className="flex items-center justify-center pl-2"
          onClick={(e) => e.stopPropagation()}
        >
          <input
            type="checkbox"
            checked={row.getIsSelected()}
            disabled={!row.getCanSelect()}
            onChange={row.getToggleSelectedHandler()}
            aria-label={`Seleccionar fila de ${row.original.clientName}`}
            className="h-4 w-4 rounded-[4px] border-app-border text-app-primary focus:ring-app-focus"
          />
        </div>
      ),
      enableSorting: false,
      size: 44,
    },

    // 2. Semáforo (sticky col 2)
    {
      id: 'severity',
      header: () => (
        <div className="flex items-center justify-center">
          {renderSortHeader('●', 'severity')}
        </div>
      ),
      cell: ({ row }) => (
        <div className="flex items-center justify-center">
          <SeverityIndicator severity={row.original.severity} mode="table" />
        </div>
      ),
      size: 44,
    },

    // 3. Cliente (sticky col 3, abre modal desde el nombre)
    {
      id: 'client',
      header: () => renderSortHeader('Cliente', 'name'),
      cell: ({ row }) => {
        const item = row.original;
        return (
          <div className="flex flex-col py-1">
            <button
              type="button"
              onClick={() => onReview?.(item)}
              className="text-left text-xs font-semibold text-app-text hover:text-app-primary hover:underline focus:outline-none focus:ring-1 focus:ring-app-focus rounded-[2px]"
            >
              {item.clientName}
            </button>
            <span className="text-[11px] tabular-nums text-app-text-muted">
              {item.cuit}
            </span>
          </div>
        );
      },
      size: 180,
    },

    // 4. Situación (columna flexible con tooltip)
    {
      id: 'situation',
      header: () => renderSortHeader('Situación', 'severity'),
      cell: ({ row }) => {
        const item = row.original;
        return (
          <div className="flex flex-col gap-1 py-1 min-w-[220px] max-w-sm">
            <Tooltip content={item.description}>
              <span className="line-clamp-2 text-xs text-app-text font-normal">
                {item.description}
              </span>
            </Tooltip>
            <div className="flex flex-wrap items-center gap-1.5">
              {item.primarySignal && <SignalChip signal={item.primarySignal} />}
              {item.otherSignals.length > 0 && (
                <OtherSignalsChip signals={item.otherSignals} />
              )}
            </div>
          </div>
        );
      },
    },

    // 5. Régimen (orden por regime + categoría, nombre completo y tooltip para RI)
    {
      id: 'regime',
      header: () => renderSortHeader('Régimen', 'regime'),
      cell: ({ row }) => {
        const item = row.original;
        if (item.regime === 'MONOTRIBUTO') {
          const cat = item.category ? ` ${item.category}` : '';
          return (
            <span className="inline-flex items-center rounded-control border border-app-border bg-app-surface-muted px-2 py-0.5 text-xs font-medium text-app-text whitespace-nowrap">
              Monotributo{cat}
            </span>
          );
        }
        return (
          <Tooltip content="Responsable Inscripto">
            <span className="inline-flex items-center rounded-control border border-app-border bg-app-surface-muted px-2 py-0.5 text-xs font-medium text-app-text whitespace-nowrap cursor-help">
              Resp. Inscripto
            </span>
          </Tooltip>
        );
      },
      size: 140,
    },

    // 6. Facturado (monto USD grande y ARS chico, whitespace-nowrap, tabular-nums)
    {
      id: 'billed',
      header: () => renderSortHeader('Facturado', 'billed', 'right'),
      cell: ({ row }) => {
        const item = row.original;
        return (
          <div className="flex flex-col items-end py-1 whitespace-nowrap">
            <span className="text-xs font-medium tabular-nums text-app-text whitespace-nowrap">
              {formatMoney(item.billedInPeriod.usd)}
            </span>
            <span className="text-[11px] tabular-nums text-app-text-muted whitespace-nowrap">
              {formatMoney(item.billedInPeriod.ars)}
            </span>
          </div>
        );
      },
      size: 140,
    },

    // 7. Acumulado 12 meses (header "Acum. 12 meses", tramo naranja, whitespace-nowrap)
    {
      id: 'accumulated12m',
      header: () => renderSortHeader('Acum. 12 meses', 'pctKCap', 'right'),
      cell: ({ row }) => {
        const item = row.original;
        return (
          <div className="flex justify-end py-1 whitespace-nowrap">
            <ProgressBar
              pctBp={item.accumulated12m.pctOfKCapBp}
              arsAmount={item.accumulated12m.ars}
            />
          </div>
        );
      },
      size: 170,
    },

    // 8. A pagar (monto + vencimiento, whitespace-nowrap, tabular-nums)
    {
      id: 'toPay',
      header: () => renderSortHeader('A pagar', 'toPay', 'right'),
      cell: ({ row }) => {
        const item = row.original;
        return (
          <div className="flex flex-col items-end py-1 whitespace-nowrap">
            <span className="text-xs font-medium tabular-nums text-app-text whitespace-nowrap">
              {formatMoney(item.toPay.amount)}
            </span>
            <span className="text-[11px] tabular-nums text-app-text-muted whitespace-nowrap">
              vence {formatDate(item.toPay.dueDate, 'short')}
            </span>
          </div>
        );
      },
      size: 140,
    },

    // 9. Estado (orden por status flujo del cierre, whitespace-nowrap)
    {
      id: 'status',
      header: () => renderSortHeader('Estado', 'status'),
      cell: ({ row }) => {
        const item = row.original;
        return (
          <div className="whitespace-nowrap">
            <StatusBadge status={item.status} paidAt={item.paidAt} />
          </div>
        );
      },
      size: 150,
    },

    // 10. Acción (botón Revisar abre modal)
    {
      id: 'action',
      header: () => null,
      cell: ({ row }) => {
        const item = row.original;
        return (
          <button
            type="button"
            onClick={() => onReview?.(item)}
            className="inline-flex h-8 items-center rounded-control border border-app-border bg-white px-2.5 text-xs font-medium text-app-text hover:bg-app-surface-muted focus:outline-none focus:ring-2 focus:ring-app-focus"
          >
            Revisar
          </button>
        );
      },
      size: 88,
    },
  ];
}
