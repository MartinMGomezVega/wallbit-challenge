import { Clock, RefreshCw, AlertCircle } from 'lucide-react';
import { ClientClosing } from './schemas';
import { UploadInvoicesStep } from './steps/UploadInvoicesStep';
import { ExternalIncomeStep } from './steps/ExternalIncomeStep';
import { ConfirmStep } from './steps/ConfirmStep';
import { ReadySummary } from './ReadySummary';

interface Props {
  closing: ClientClosing;
  regime: string;
}

const MONTH_NAMES: Record<string, string> = {
  '01': 'enero',
  '02': 'febrero',
  '03': 'marzo',
  '04': 'abril',
  '05': 'mayo',
  '06': 'junio',
  '07': 'julio',
  '08': 'agosto',
  '09': 'septiembre',
  '10': 'octubre',
  '11': 'noviembre',
  '12': 'diciembre',
};

function formatPeriodMonth(periodStr: string): string {
  const parts = periodStr.split('-');
  const month = parts[1];
  const year = parts[0];
  if (parts.length === 2 && month && MONTH_NAMES[month]) {
    return `${MONTH_NAMES[month]} ${year}`;
  }
  return periodStr;
}

export function PeriodCard({ closing, regime }: Props) {
  const monthName = formatPeriodMonth(closing.period);

  if (closing.state === 'READY') {
    return <ReadySummary closing={closing} />;
  }

  if (closing.state === 'IN_REVIEW') {
    return (
      <div className="space-y-4">
        <div className="rounded-card border border-app-border bg-app-surface p-6 shadow-sm space-y-4">
          <div className="flex items-center gap-2">
            <span className="inline-flex items-center gap-1.5 rounded-control bg-status-ready-bg px-2.5 py-1 text-xs font-semibold text-status-ready-text">
              <Clock className="w-3.5 h-3.5" aria-hidden="true" />
              En revisión
            </span>
            {closing.updating && (
              <span
                data-testid="updating-badge"
                className="inline-flex items-center gap-1 rounded-control bg-severity-yellow-bg border border-severity-yellow-border px-2 py-0.5 text-xs font-medium text-severity-yellow-text"
              >
                <RefreshCw className="w-3 h-3 animate-spin" aria-hidden="true" />
                Actualizando
              </span>
            )}
          </div>

          <div>
            <h3 className="text-lg font-semibold text-app-text">
              Cierre de {monthName}
            </h3>
            <p className="mt-1 text-sm text-app-text-muted leading-relaxed">
              Tu contador está revisando tu cierre de {monthName}. Te avisamos cuando esté listo.
            </p>
          </div>

          {/* Resumen de lo cargado */}
          <div className="rounded-control bg-app-surface-muted/50 p-4 border border-app-border space-y-2 text-xs">
            <h4 className="font-semibold text-app-text">Resumen de lo cargado</h4>
            <div className="grid grid-cols-2 gap-3 text-app-text-muted">
              <div>
                <span>Facturas cargadas:</span>
                <p className="font-medium text-app-text text-sm tabular-nums">
                  {closing.input.invoicesCount}
                </p>
              </div>
              <div>
                <span>Ingresos externos:</span>
                <p className="font-medium text-app-text text-sm">
                  {closing.input.answered ? 'Declarados' : 'No declarados'}
                </p>
              </div>
            </div>
          </div>
        </div>

        {/* Si tiene un snapshot publicado previo y se reabrió, mostramos el anterior con advertencia de actualización */}
        {closing.published && (
          <div className="space-y-2">
            <div className="rounded-control bg-severity-yellow-bg border border-severity-yellow-border p-3 text-xs text-severity-yellow-text flex items-center gap-2">
              <AlertCircle className="w-4 h-4 flex-shrink-0" />
              <span>
                Mostrando liquidación anterior. Los nuevos valores se actualizarán una vez aprobados por tu contador.
              </span>
            </div>
            <ReadySummary closing={closing} />
          </div>
        )}
      </div>
    );
  }

  // ACTION_REQUIRED
  return (
    <div className="rounded-card border border-app-border bg-app-surface p-6 shadow-sm space-y-6">
      <div className="flex items-center justify-between border-b border-app-border pb-4">
        <div>
          <span className="inline-flex items-center gap-1.5 rounded-control bg-severity-yellow-bg border border-severity-yellow-border px-2.5 py-1 text-xs font-semibold text-severity-yellow-text">
            Acción requerida
          </span>
          <h2 className="text-lg font-bold text-app-text mt-2">
            Cierre de {monthName}
          </h2>
          <p className="text-xs text-app-text-muted mt-0.5">
            Completá estos 3 pasos para que tu contador pueda liquidar tus impuestos.
          </p>
        </div>
      </div>

      <div className="space-y-6">
        <UploadInvoicesStep
          periodMonthName={monthName}
          regime={regime}
          invoicesCount={closing.input.invoicesCount}
        />

        <div className="border-t border-app-border pt-5">
          <ExternalIncomeStep
            period={closing.period}
            periodMonthName={monthName}
            answered={closing.input.answered}
          />
        </div>

        <ConfirmStep
          period={closing.period}
          answered={closing.input.answered}
        />
      </div>
    </div>
  );
}
