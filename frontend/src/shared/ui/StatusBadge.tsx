import { formatDate } from '../format/date';

export type ClosingStatus =
  | 'PENDING_CLIENT_INPUT'
  | 'READY_FOR_REVIEW'
  | 'APPROVED'
  | 'NOTIFIED';

export interface StatusBadgeProps {
  status: ClosingStatus;
  paidAt?: string | null;
  className?: string;
}

const STATUS_CONFIG: Record<
  ClosingStatus,
  { label: string; bgClass: string; textClass: string }
> = {
  PENDING_CLIENT_INPUT: {
    label: 'Esperando al cliente',
    bgClass: 'bg-status-pending-bg',
    textClass: 'text-status-pending-text',
  },
  READY_FOR_REVIEW: {
    label: 'Para revisar',
    bgClass: 'bg-status-ready-bg',
    textClass: 'text-status-ready-text',
  },
  APPROVED: {
    label: 'Aprobado',
    bgClass: 'bg-status-approved-bg',
    textClass: 'text-status-approved-text',
  },
  NOTIFIED: {
    label: 'Notificado',
    bgClass: 'bg-status-notified-bg',
    textClass: 'text-status-notified-text',
  },
};

export function StatusBadge({ status, paidAt, className = '' }: StatusBadgeProps) {
  const cfg = STATUS_CONFIG[status] ?? {
    label: status,
    bgClass: 'bg-status-pending-bg',
    textClass: 'text-status-pending-text',
  };

  return (
    <div className={`inline-flex flex-wrap items-center gap-1.5 ${className}`}>
      <span
        className={`inline-flex items-center rounded-control px-2 py-0.5 text-xs font-medium ${cfg.bgClass} ${cfg.textClass}`}
      >
        {cfg.label}
      </span>
      {paidAt && (
        <span className="inline-flex items-center rounded-control border border-status-paid-border bg-severity-green-bg px-2 py-0.5 text-xs font-medium text-status-paid-text">
          Pagado {formatDate(paidAt, 'short')}
        </span>
      )}
    </div>
  );
}

