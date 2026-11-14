import { Link, useSearchParams } from 'react-router-dom';
import { AlertCircle, BarChart3, RotateCcw } from 'lucide-react';
import { useStats } from './useStats';
import { SeverityOverview } from './SeverityOverview';
import { KpiCard } from './KpiCard';
import { ProgressKpi } from './ProgressKpi';
import { SignalsBarChart } from './charts/SignalsBarChart';
import { CategoryBarChart } from './charts/CategoryBarChart';
import { DueDatesBarChart } from './charts/DueDatesBarChart';
import { ServiceHealth } from './ServiceHealth';
import { formatDate, formatPeriodName } from '../../shared/format/date';
import { formatMoney } from '../../shared/format/money';
import { ApiError } from '../../shared/api/http';
import { useDocumentTitle } from '../../shared/hooks/useDocumentTitle';

const DEFAULT_PERIOD = '2026-10';

function SummarySkeleton() {
  return (
    <div
      className="grid grid-cols-12 gap-4"
      aria-busy="true"
      aria-label="Cargando resumen del período"
    >
      {[0, 1, 2, 3].map((idx) => (
        <div
          key={idx}
          className="col-span-12 h-32 animate-pulse rounded-lg border border-app-border bg-app-surface p-5 md:col-span-6 xl:col-span-3"
        >
          <div className="h-3 w-24 rounded bg-app-border" />
          <div className="mt-4 h-7 w-36 rounded bg-app-border" />
          <div className="mt-4 h-3 w-28 rounded bg-app-border" />
        </div>
      ))}
      <div className="col-span-12 h-80 animate-pulse rounded-lg border border-app-border bg-app-surface p-5 lg:col-span-6" />
      <div className="col-span-12 h-80 animate-pulse rounded-lg border border-app-border bg-app-surface p-5 lg:col-span-6" />
      <div className="col-span-12 h-80 animate-pulse rounded-lg border border-app-border bg-app-surface p-5" />
      {[0, 1, 2].map((idx) => (
        <div
          key={idx}
          className="col-span-12 h-28 animate-pulse rounded-lg border border-app-border bg-app-surface p-5 md:col-span-4"
        />
      ))}
    </div>
  );
}

export function SummaryPage() {
  useDocumentTitle('Resumen · Wallbit Contabilidad');
  const [searchParams] = useSearchParams();
  const period = searchParams.get('period') || DEFAULT_PERIOD;
  const { data, isLoading, isError, error, refetch } = useStats(period);

  const errorCode = error instanceof ApiError ? error.code : undefined;

  return (
    <div className="mx-auto w-full max-w-[1440px] px-6 py-6">
      <div className="mb-6 flex flex-wrap items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight text-app-text">
            Resumen de {formatPeriodName(period)}
          </h1>
          {data && (
            <p className="mt-1 text-sm text-app-text-muted">
              {data.progress.total} clientes en el período
            </p>
          )}
        </div>
      </div>

      {isLoading && <SummarySkeleton />}

      {isError && (
        <div
          role="alert"
          title={errorCode}
          className="flex items-center justify-between gap-4 rounded-lg border border-app-severity-red-border bg-app-severity-red-bg p-5 text-app-severity-red-text"
        >
          <div className="flex items-center gap-3">
            <AlertCircle className="h-5 w-5 shrink-0" />
            <div>
              <p className="font-medium">
                No pudimos cargar el resumen del período.
              </p>
              <p className="text-sm">
                Revisá tu conexión e intentá nuevamente.
              </p>
            </div>
          </div>
          <button
            type="button"
            onClick={() => refetch()}
            className="inline-flex h-8 items-center gap-1.5 rounded-md border border-app-border bg-app-surface px-3 text-sm font-medium text-app-text transition-colors hover:bg-app-surface-muted focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-app-focus"
          >
            <RotateCcw className="h-4 w-4" />
            Reintentar
          </button>
        </div>
      )}

      {!isLoading && !isError && data && data.progress.total === 0 && (
        <div className="flex flex-col items-center justify-center rounded-lg border border-app-border bg-app-surface px-6 py-16 text-center">
          <BarChart3 className="h-8 w-8 text-app-text-muted" />
          <h2 className="mt-3 text-lg font-semibold text-app-text">
            No hay cierres para {formatPeriodName(period)}
          </h2>
          <p className="mt-1 text-sm text-app-text-muted">
            Elegí otro período para ver las métricas del estudio.
          </p>
          {period !== DEFAULT_PERIOD && (
            <Link
              to={`/contador/resumen?period=${DEFAULT_PERIOD}`}
              className="mt-4 inline-flex h-8 items-center rounded-md bg-app-primary px-3 text-sm font-medium text-white transition-colors hover:bg-app-primary-hover focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-app-focus"
            >
              Ver {formatPeriodName(DEFAULT_PERIOD)}
            </Link>
          )}
        </div>
      )}

      {!isLoading && !isError && data && data.progress.total > 0 && (
        <div className="grid grid-cols-12 gap-4">
          <SeverityOverview
            period={period}
            total={data.progress.total}
            bySeverity={data.risk.bySeverity}
          />

          <ProgressKpi
            period={period}
            progress={data.progress}
            className="col-span-12 md:col-span-4"
          />

          <KpiCard
            label="Facturado del mes"
            value={formatMoney(data.totals.billedUSD)}
            secondary={formatMoney(data.totals.billedARS)}
            className="col-span-12 md:col-span-4"
          />

          <KpiCard
            label="A pagar estimado"
            value={formatMoney(data.totals.toPayARS)}
            secondary={
              data.totals.toPayByDueDate[0]
                ? `Próximo vence ${formatDate(data.totals.toPayByDueDate[0].date, 'short')}`
                : 'Sin vencimientos'
            }
            className="col-span-12 md:col-span-4"
          />

          <SignalsBarChart
            period={period}
            signals={data.risk.bySignal}
            className="col-span-12 lg:col-span-6"
          />

          <CategoryBarChart
            categories={data.risk.byCategory}
            nearKCap={data.risk.nearKCap}
            className="col-span-12 lg:col-span-6"
          />

          <DueDatesBarChart
            dueDates={data.totals.toPayByDueDate}
            className="col-span-12"
          />

          <ServiceHealth success={data.success} />
        </div>
      )}
    </div>
  );
}
