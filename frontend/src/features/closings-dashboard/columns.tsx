import { ColumnDef } from '@tanstack/react-table';
import { ArrowUpDown, ArrowUp, ArrowDown } from 'lucide-react';
import { ClosingRow } from './schemas';
import { SeverityIndicator } from '../../shared/ui/SeverityIndicator';
import { SignalChip, OtherSignalsChip } from '../../shared/ui/SignalChip';
import { ProgressBar } from '../../shared/ui/ProgressBar';
import { StatusBadge } from '../../shared/ui/StatusBadge';
import { formatMoney } from '../../shared/format/money';
import { formatDate } from '../../shared/format/date';

export interface ColumnOptions {
  sort?: string;
  dir?: 'asc' | 'desc';
  onSort?: (sortField: 'severity' | 'name' | 'billed' | 'toPay' | 'pctKCap') => void;
  onReview?: (row: ClosingRow) => void;
}

export function createColumns(options: ColumnOptions = {}): ColumnDef<ClosingRow>[] {
  const { sort, dir, onSort, onReview } = options;

  const renderSortHeader = (
    label: string,
    field: 'severity' | 'name' | 'billed' | 'toPay' | 'pctKCap',
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
    // 1. Checkbox de selección
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
      size: 40,
    },

    // 2. Semáforo
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
      size: 50,
    },

    // 3. Situación
    {
      id: 'situation',
      header: () => renderSortHeader('Situación', 'severity'),
      cell: ({ row }) => {
        const item = row.original;
        return (
          <div className="flex flex-col gap-1 py-1 max-w-sm">
            <span
              className="line-clamp-2 text-xs text-app-text font-normal"
              title={item.description}
            >
              {item.description}
            </span>
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

    // 4. Cliente
    {
      id: 'client',
      header: () => renderSortHeader('Cliente', 'name'),
      cell: ({ row }) => {
        const item = row.original;
        return (
          <div className="flex flex-col py-1">
            <span className="text-xs font-semibold text-app-text">
              {item.clientName}
            </span>
            <span className="text-[11px] tabular-nums text-app-text-muted">
              {item.cuit}
            </span>
          </div>
        );
      },
    },

    // 5. Régimen
    {
      id: 'regime',
      header: () => (
        <span className="text-xs font-semibold uppercase tracking-wider text-app-text-muted">
          Rég.
        </span>
      ),
      cell: ({ row }) => {
        const item = row.original;
        const text =
          item.regime === 'MONOTRIBUTO'
            ? `Mono ${item.category || ''}`.trim()
            : 'RI';
        return (
          <span className="inline-flex items-center rounded-control border border-app-border bg-app-surface-muted px-2 py-0.5 text-xs font-medium text-app-text">
            {text}
          </span>
        );
      },
    },

    // 6. Facturado
    {
      id: 'billed',
      header: () => renderSortHeader('Facturado', 'billed', 'right'),
      cell: ({ row }) => {
        const item = row.original;
        return (
          <div className="flex flex-col items-end py-1">
            <span className="text-xs font-medium tabular-nums text-app-text">
              {formatMoney(item.billedInPeriod.usd)}
            </span>
            <span className="text-[11px] tabular-nums text-app-text-muted">
              {formatMoney(item.billedInPeriod.ars)}
            </span>
          </div>
        );
      },
    },

    // 7. Acumulado 12 meses
    {
      id: 'accumulated12m',
      header: () => renderSortHeader('Acum. 12m', 'pctKCap', 'right'),
      cell: ({ row }) => {
        const item = row.original;
        return (
          <div className="flex justify-end py-1">
            <ProgressBar
              pctBp={item.accumulated12m.pctOfKCapBp}
              arsAmount={item.accumulated12m.ars}
            />
          </div>
        );
      },
    },

    // 8. A pagar
    {
      id: 'toPay',
      header: () => renderSortHeader('A pagar', 'toPay', 'right'),
      cell: ({ row }) => {
        const item = row.original;
        return (
          <div className="flex flex-col items-end py-1">
            <span className="text-xs font-medium tabular-nums text-app-text">
              {formatMoney(item.toPay.amount)}
            </span>
            <span className="text-[11px] tabular-nums text-app-text-muted">
              vence {formatDate(item.toPay.dueDate, 'short')}
            </span>
          </div>
        );
      },
    },

    // 9. Estado
    {
      id: 'status',
      header: () => (
        <span className="text-xs font-semibold uppercase tracking-wider text-app-text-muted">
          Estado
        </span>
      ),
      cell: ({ row }) => {
        const item = row.original;
        return <StatusBadge status={item.status} paidAt={item.paidAt} />;
      },
    },

    // 10. Acción
    {
      id: 'action',
      header: () => null,
      cell: ({ row }) => {
        const item = row.original;
        return (
          <button
            type="button"
            onClick={(e) => {
              e.stopPropagation();
              onReview?.(item);
            }}
            className="inline-flex h-8 items-center rounded-control border border-app-border bg-white px-2.5 text-xs font-medium text-app-text hover:bg-app-surface-muted focus:outline-none focus:ring-2 focus:ring-app-focus"
          >
            Revisar
          </button>
        );
      },
      size: 80,
    },
  ];
}

