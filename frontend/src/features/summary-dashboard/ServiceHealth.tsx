import { PeriodStats } from './schemas';
import { formatPercentBp } from '../../shared/format/percent';
import { KpiCard } from './KpiCard';

export interface ServiceHealthProps {
  success: PeriodStats['success'];
}

export function ServiceHealth({ success }: ServiceHealthProps) {
  return (
    <section
      aria-labelledby="service-health-heading"
      className="col-span-12 mt-2 border-t border-app-border pt-6"
      data-testid="service-health"
    >
      <div className="mb-4">
        <h2 id="service-health-heading" className="text-lg font-semibold text-app-text">
          ¿Está funcionando el servicio?
        </h2>
        <p className="mt-0.5 text-sm text-app-text-muted">
          Las métricas con las que mido si el servicio cumple su objetivo.
        </p>
      </div>
      <div className="grid grid-cols-12 gap-4">
        <KpiCard
          label="Clientes en verde"
          value={formatPercentBp(success.greenPct)}
          secondary={`${success.greenCount} de ${success.total} — Cierres sin alertas ni errores`}
          className="col-span-12 md:col-span-4"
        />
        <KpiCard
          label="Avisados antes del día 10"
          value={formatPercentBp(success.notifiedBeforeDay10Pct)}
          secondary={`${success.notifiedBeforeDay10Count} de ${success.total} — Notificados a tiempo al cliente`}
          className="col-span-12 md:col-span-4"
        />
        <KpiCard
          label="Pagos en término"
          value={formatPercentBp(success.paidOnTimePct)}
          secondary={`${success.paidOnTimeCount} de ${success.notifiedCount} — Sobre cierres notificados`}
          className="col-span-12 md:col-span-4"
        />
      </div>
    </section>
  );
}

