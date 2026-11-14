import { useState, useMemo, useCallback } from 'react';
import { RowSelectionState } from '@tanstack/react-table';
import { AlertCircle, RotateCcw } from 'lucide-react';
import { useDashboardParams, DashboardFilters, filtersToSearchParams } from './useDashboardParams';
import {
  useClosingsList,
  useClosingsSummary,
  useSignalsCatalog,
} from './useClosingsQuery';
import { SummaryChips } from './SummaryChips';
import { FiltersBar } from './FiltersBar';
import { ClosingsTable } from './ClosingsTable';
import { SelectionBar } from './SelectionBar';
import { BatchApproveDialog } from './BatchApproveDialog';
import { formatPeriodName } from '../../shared/format/date';
import { useDocumentTitle } from '../../shared/hooks/useDocumentTitle';
import { ClosingRow } from './schemas';
import {
  ReviewQueueProvider,
  ReviewModal,
  useReviewQueue,
} from '../closing-review';

export interface DashboardPageProps {
  onReviewClient?: (clientId: string) => void;
}

function formatFilterLabel(filters: DashboardFilters): string {
  const parts: string[] = [];
  if (filters.severity.length > 0) {
    const sevMap: Record<string, string> = {
      RED: 'Rojos',
      YELLOW: 'Amarillos',
      GREEN: 'Verdes',
    };
    parts.push(filters.severity.map((s: string) => sevMap[s] || s).join(', '));
  }
  if (filters.regime) {
    parts.push(filters.regime === 'MONOTRIBUTO' ? 'Monotributo' : 'RI');
  }
  if (filters.signal.length > 0) {
    parts.push(filters.signal.join(', '));
  }
  if (filters.status.length > 0) {
    parts.push(filters.status.join(', '));
  }
  if (filters.q) {
    parts.push(`"${filters.q}"`);
  }
  if (parts.length === 0) {
    return `Todos los clientes (${formatPeriodName(filters.period)})`;
  }
  return parts.join(' · ');
}

function DashboardContent({
  onReviewClient,
  highlightedClosingId,
}: DashboardPageProps & { highlightedClosingId: string | null }) {
  useDocumentTitle('Cierres · Wallbit Contabilidad');
  const {
    filters,
    toggleSeverity,
    setSignal,
    setRegime,
    setStatus,
    setQ,
    setSort,
    setPage,
    setPageSize,
    clearFilters,
  } = useDashboardParams();

  const [rowSelection, setRowSelection] = useState<RowSelectionState>({});
  const [batchDialogOpen, setBatchDialogOpen] = useState(false);

  const { data: catalog = [] } = useSignalsCatalog();
  const { data: summary } = useClosingsSummary(filters.period, filters);
  const closingsQuery = useClosingsList(filters);
  const { openQueue } = useReviewQueue();

  const items = useMemo(() => closingsQuery.data?.items ?? [], [closingsQuery.data?.items]);
  const total = closingsQuery.data?.total ?? 0;

  // Clientes seleccionados y conteo de verdes
  const selectedCount = useMemo(() => {
    return Object.keys(rowSelection).filter((k) => rowSelection[k]).length;
  }, [rowSelection]);

  const selectedGreenItems = useMemo(() => {
    if (selectedCount === 0) return [];
    const selectedIds = new Set(
      Object.keys(rowSelection).filter((k) => rowSelection[k])
    );
    return items
      .filter((it) => selectedIds.has(it.closingId) && it.severity === 'GREEN')
      .map((it) => ({
        id: it.closingId,
        version: it.version,
        clientName: it.clientName,
      }));
  }, [items, rowSelection, selectedCount]);

  const greenSelectedCount = selectedGreenItems.length;

  // Modo 1: Click en fila o botón Revisar de fila
  const handleReviewRow = useCallback(
    (row: ClosingRow) => {
      onReviewClient?.(row.clientId);
      openQueue({
        mode: 'row',
        startId: row.closingId,
        period: filters.period,
        filterParams: filtersToSearchParams(filters),
        label: formatFilterLabel(filters),
      });
    },
    [filters, onReviewClient, openQueue]
  );

  // Modo 2: Revisar chip de color
  const handleReviewSeverityChip = useCallback(
    (sev: string) => {
      openQueue({
        mode: 'chip',
        period: filters.period,
        filterParams: filtersToSearchParams({
          ...filters,
          severity: [sev],
        }),
        label: `${
          sev === 'RED' ? 'Rojos' : sev === 'YELLOW' ? 'Amarillos' : 'Verdes'
        } · ${formatPeriodName(filters.period)}`,
      });
    },
    [filters, openQueue]
  );

  // Modo 3: Revisar seleccionados
  const handleReviewSelected = useCallback(() => {
    const selectedIds = items
      .filter((it) => rowSelection[it.closingId])
      .map((it) => it.closingId);

    if (selectedIds.length > 0) {
      openQueue({
        mode: 'selection',
        ids: selectedIds,
        startId: selectedIds[0],
        label: `${selectedIds.length} clientes seleccionados`,
      });
    }
  }, [items, openQueue, rowSelection]);

  const handleClearSelection = () => {
    setRowSelection({});
  };

  return (
    <div className="space-y-6">
      {/* Título de la página y subtítulo */}
      <div className="flex flex-wrap items-baseline justify-between gap-4">
        <div>
          <h1 className="text-xl font-bold text-app-text">
            Cierre de {formatPeriodName(filters.period)}
          </h1>
          <p className="text-xs text-app-text-muted mt-0.5">
            {total} {total === 1 ? 'cliente' : 'clientes'}
          </p>
        </div>

        {/* Chips de resumen */}
        <SummaryChips
          summary={summary}
          selectedSeverities={filters.severity}
          onToggleSeverity={toggleSeverity}
          onReviewSeverity={handleReviewSeverityChip}
        />
      </div>

      {/* Barra de filtros */}
      <FiltersBar
        filters={filters}
        summary={summary}
        catalog={catalog}
        onSetSignal={setSignal}
        onSetRegime={setRegime}
        onSetStatus={setStatus}
        onSetQ={setQ}
        onClearFilters={clearFilters}
      />

      {/* Estado de error o Tabla */}
      {closingsQuery.isError ? (
        <div className="rounded-card border border-severity-red-border bg-severity-red-bg p-6 text-center">
          <div className="flex flex-col items-center justify-center gap-2">
            <AlertCircle className="h-8 w-8 text-severity-red-icon" aria-hidden="true" />
            <h3 className="text-sm font-semibold text-severity-red-text">
              No se pudieron cargar los cierres
            </h3>
            <p className="text-xs text-severity-red-text/80 max-w-md">
              {(closingsQuery.error as Error)?.message || 'Ocurrió un error inesperado al consultar la API.'}
            </p>
            <button
              type="button"
              onClick={() => closingsQuery.refetch()}
              className="mt-2 inline-flex items-center gap-1.5 rounded-control bg-severity-red-icon px-3 py-1.5 text-xs font-semibold text-white shadow-sm hover:opacity-90 focus:outline-none focus:ring-2 focus:ring-app-focus"
            >
              <RotateCcw className="h-3.5 w-3.5" aria-hidden="true" />
              <span>Reintentar</span>
            </button>
          </div>
        </div>
      ) : (
        <ClosingsTable
          data={items}
          total={total}
          page={filters.page}
          pageSize={filters.pageSize}
          sort={filters.sort}
          dir={filters.dir}
          isLoading={closingsQuery.isLoading}
          onSortChange={(field) => setSort(field)}
          onPageChange={setPage}
          onPageSizeChange={setPageSize}
          onReviewRow={handleReviewRow}
          rowSelection={rowSelection}
          onRowSelectionChange={setRowSelection}
          onClearFilters={clearFilters}
          highlightedClosingId={highlightedClosingId}
        />
      )}

      {/* Barra de selección flotante */}
      <SelectionBar
        selectedCount={selectedCount}
        greenCount={greenSelectedCount}
        onClearSelection={handleClearSelection}
        onReviewSelected={handleReviewSelected}
        onApproveGreens={() => setBatchDialogOpen(true)}
      />

      {/* Diálogo de aprobación en lote (spec 010 criterio 13) */}
      <BatchApproveDialog
        open={batchDialogOpen}
        onOpenChange={setBatchDialogOpen}
        period={filters.period}
        items={selectedGreenItems}
        onSuccess={() => {
          setRowSelection({});
        }}
      />
    </div>
  );
}

export function DashboardPage(props: DashboardPageProps) {
  const { filters, setPage } = useDashboardParams();
  const [highlightedClosingId, setHighlightedClosingId] = useState<string | null>(null);

  const handleCloseQueue = useCallback(
    ({ lastViewedId, targetPage }: { lastViewedId: string; targetPage: number }) => {
      if (targetPage !== filters.page) {
        setPage(targetPage);
      }
      setHighlightedClosingId(lastViewedId);
      setTimeout(() => {
        setHighlightedClosingId(null);
      }, 2000);
    },
    [filters.page, setPage]
  );

  return (
    <ReviewQueueProvider pageSize={filters.pageSize} onCloseQueue={handleCloseQueue}>
      <DashboardContent
        {...props}
        highlightedClosingId={highlightedClosingId}
      />
      <ReviewModal />
    </ReviewQueueProvider>
  );
}
