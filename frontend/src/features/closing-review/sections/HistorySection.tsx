import { ClosingDetailDTO } from '../schemas';
import { formatDate } from '../../../shared/format';
import { StatusBadge } from '../../../shared/ui/StatusBadge';
import { History, ShieldCheck, CheckCircle2, Clock } from 'lucide-react';

export interface HistorySectionProps {
  detail: ClosingDetailDTO;
}

export function HistorySection({ detail }: HistorySectionProps) {
  const { closing } = detail;

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {/* Tarjeta de versión y cálculo */}
        <div className="rounded-card border border-app-border bg-app-surface p-5 space-y-4">
          <div className="flex items-center gap-2 border-b border-app-border pb-3">
            <History className="h-4 w-4 text-app-text-muted" aria-hidden="true" />
            <h4 className="text-xs font-semibold uppercase tracking-wider text-app-text-muted">
              Cálculo y Versión
            </h4>
          </div>

          <div className="space-y-3 text-xs">
            <div className="flex items-center justify-between">
              <span className="text-app-text-muted">Versión del cierre</span>
              <span className="font-semibold text-app-text">v{closing.version}</span>
            </div>

            <div className="flex items-center justify-between">
              <span className="text-app-text-muted">Fecha de cómputo</span>
              <span className="font-medium text-app-text">
                {formatDate(closing.computedAt, 'full')}
              </span>
            </div>

            <div className="pt-2">
              <span className="text-app-text-muted block mb-1">Hash de integridad (SHA-256)</span>
              <code className="block w-full overflow-x-auto rounded border border-app-border bg-app-surface-muted px-2.5 py-1.5 font-mono text-[11px] text-app-text select-all">
                {closing.resultHash}
              </code>
            </div>
          </div>
        </div>

        {/* Tarjeta de estado y aprobación */}
        <div className="rounded-card border border-app-border bg-app-surface p-5 space-y-4">
          <div className="flex items-center gap-2 border-b border-app-border pb-3">
            <ShieldCheck className="h-4 w-4 text-app-text-muted" aria-hidden="true" />
            <h4 className="text-xs font-semibold uppercase tracking-wider text-app-text-muted">
              Estado y Aprobación
            </h4>
          </div>

          <div className="space-y-3 text-xs">
            <div className="flex items-center justify-between">
              <span className="text-app-text-muted">Estado actual</span>
              <StatusBadge status={closing.status} paidAt={closing.paidAt} />
            </div>

            <div className="flex items-center justify-between">
              <span className="text-app-text-muted">Aprobación</span>
              {closing.approvedAt ? (
                <div className="flex items-center gap-1.5 text-status-approved-text">
                  <CheckCircle2 className="h-3.5 w-3.5" aria-hidden="true" />
                  <span className="font-medium">
                    {closing.approvedBy ?? 'Contador'} ({formatDate(closing.approvedAt, 'full')})
                  </span>
                </div>
              ) : (
                <div className="flex items-center gap-1.5 text-app-text-muted">
                  <Clock className="h-3.5 w-3.5" aria-hidden="true" />
                  <span>Pendiente de aprobación</span>
                </div>
              )}
            </div>

            <div className="flex items-center justify-between">
              <span className="text-app-text-muted">Registro de pago</span>
              {closing.paidAt ? (
                <span className="font-medium text-status-paid-text">
                  {formatDate(closing.paidAt, 'full')}
                </span>
              ) : (
                <span className="text-app-text-muted">Sin registrar</span>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
