import { useNavigate } from 'react-router-dom';
import { PeriodStats } from './schemas';
import { KpiCard } from './KpiCard';

export interface ProgressKpiProps {
  period: string;
  progress: PeriodStats['progress'];
  className?: string;
}

interface Segment {
  status: 'NOTIFIED' | 'APPROVED' | 'READY_FOR_REVIEW' | 'PENDING_CLIENT_INPUT';
  label: string;
  count: number;
  color: string;
}

export function ProgressKpi({ period, progress, className = '' }: ProgressKpiProps) {
  const navigate = useNavigate();

  const { total, byStatus, estimatedReviewMinutes } = progress;
  const unapproved = byStatus.PENDING_CLIENT_INPUT + byStatus.READY_FOR_REVIEW;
  const hours = Math.floor(estimatedReviewMinutes / 60);
  const mins = estimatedReviewMinutes % 60;

  const segments: Segment[] = [
    {
      status: 'NOTIFIED',
      label: 'Notificado',
      count: byStatus.NOTIFIED,
      color: '#047857',
    },
    {
      status: 'APPROVED',
      label: 'Aprobado',
      count: byStatus.APPROVED,
      color: '#0369A1',
    },
    {
      status: 'READY_FOR_REVIEW',
      label: 'Para revisar',
      count: byStatus.READY_FOR_REVIEW,
      color: 'rgba(0, 111, 192, 0.4)',
    },
    {
      status: 'PENDING_CLIENT_INPUT',
      label: 'Esperando al cliente',
      count: byStatus.PENDING_CLIENT_INPUT,
      color: '#D4D4D4',
    },
  ];

  const handleSegmentClick = (status: Segment['status']) => {
    navigate(`/contador?period=${encodeURIComponent(period)}&status=${status}`);
  };

  return (
    <KpiCard
      label="Progreso del cierre"
      value={`${unapproved} de ${total} sin aprobar`}
      secondary={`≈ ${hours} h ${mins} min de revisión`}
      className={className}
    >
      <div
        className="mt-3 flex h-2 w-full overflow-hidden rounded-full bg-app-border"
        role="group"
        aria-label="Distribución de estados del cierre"
      >
        {total > 0 &&
          segments.map((seg) => {
            if (seg.count === 0) return null;
            const pct = (seg.count / total) * 100;
            return (
              <button
                key={seg.status}
                type="button"
                onClick={() => handleSegmentClick(seg.status)}
                title={`${seg.label}: ${seg.count} clientes`}
                aria-label={`${seg.label}: ${seg.count} clientes`}
                style={{ width: `${pct}%`, backgroundColor: seg.color }}
                className="h-full cursor-pointer transition-opacity hover:opacity-80 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-app-focus"
              />
            );
          })}
      </div>
    </KpiCard>
  );
}

