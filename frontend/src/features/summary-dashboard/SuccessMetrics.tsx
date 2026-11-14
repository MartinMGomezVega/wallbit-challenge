import { PeriodStats } from './schemas';
import { formatPercentBp } from '../../shared/format/percent';
import { KpiCard } from './KpiCard';

export interface SuccessMetricsProps {
  success: PeriodStats['success'];
}

export function SuccessMetrics({ success }: SuccessMetricsProps) {
  return (
    <div className="col-span-12 grid grid-cols-12 gap-4">
      <KpiCard
        label="Clientes en verde"
        value={formatPercentBp(success.greenPct)}
        secondary="Cierres sin alertas ni errores"
        className="col-span-12 md:col-span-4"
      />
      <KpiCard
        label="Avisados antes del día 10"
        value={formatPercentBp(success.notifiedBeforeDay10Pct)}
        secondary="Notificados a tiempo al cliente"
        className="col-span-12 md:col-span-4"
      />
      <KpiCard
        label="Pagos en término"
        value={formatPercentBp(success.paidOnTimePct)}
        secondary="Sobre cierres notificados"
        className="col-span-12 md:col-span-4"
      />
    </div>
  );
}

