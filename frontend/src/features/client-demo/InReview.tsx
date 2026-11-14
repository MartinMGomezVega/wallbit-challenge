import { ClientClosing } from './schemas';
import { ValidatedSummaryCard } from './ValidatedSummaryCard';

interface InReviewProps {
  octoberClosing: ClientClosing;
  previousClosing?: ClientClosing;
  clientId: string;
  onRefresh?: () => void;
}

export function InReview({
  octoberClosing,
  previousClosing,
  clientId,
  onRefresh,
}: InReviewProps) {
  const invCount = octoberClosing.input.invoicesCount;

  return (
    <div className="space-y-4">
      {/* Banner En revisión */}
      <div className="rounded-2xl border border-blue-200 bg-blue-50/70 p-4 text-blue-950 shadow-sm space-y-2.5">
        <div className="flex items-center gap-2">
          <span className="flex h-7 w-7 items-center justify-center rounded-full bg-blue-600 text-xs text-white" aria-hidden="true">
            ⏳
          </span>
          <h2 className="text-sm font-bold text-blue-950">
            Tu contadora está revisando tu cierre de octubre
          </h2>
        </div>

        <p className="text-xs text-blue-900 leading-relaxed">
          Te avisaremos por correo y notificación en la app cuando el resumen y la liquidación estén listos.
        </p>

        {/* Resumen de lo cargado */}
        <div className="rounded-xl bg-white/80 p-3 text-xs space-y-1.5 border border-blue-100">
          <span className="text-[11px] font-semibold text-blue-800 uppercase tracking-wider block">
            Datos enviados para revisión
          </span>
          <div className="flex items-center justify-between text-neutral-800">
            <span>Comprobantes cargados:</span>
            <span className="font-semibold">{invCount} {invCount === 1 ? 'comprobante' : 'comprobantes'}</span>
          </div>
          <div className="flex items-center justify-between text-neutral-800">
            <span>Cobros fuera de Wallbit:</span>
            <span className="font-semibold">Declarados</span>
          </div>
        </div>
      </div>

      {/* Tarjeta validada del mes anterior si existe */}
      {previousClosing && previousClosing.published && (
        <div className="space-y-2 pt-2">
          <div className="flex items-center justify-between px-1">
            <h3 className="text-xs font-bold text-neutral-500 uppercase tracking-wider">
              Mes anterior ({previousClosing.period === '2026-09' ? 'Septiembre' : previousClosing.period})
            </h3>
            <span className="text-[11px] text-neutral-400">Validado</span>
          </div>
          <ValidatedSummaryCard
            closing={previousClosing}
            clientId={clientId}
            onPaidSuccess={onRefresh}
          />
        </div>
      )}
    </div>
  );
}

