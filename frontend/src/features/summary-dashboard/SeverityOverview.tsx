import { Link } from 'react-router-dom';
import { AlertOctagon, AlertTriangle, CheckCircle2, ArrowRight } from 'lucide-react';
import { Severity } from './schemas';

export interface SeverityOverviewProps {
  period: string;
  total: number;
  bySeverity: {
    RED: number;
    YELLOW: number;
    GREEN: number;
  };
}

interface SeverityCardConfig {
  key: Severity;
  title: string;
  count: number;
  borderClass: string;
  iconColorClass: string;
  Icon: typeof AlertOctagon;
}

export function SeverityOverview({ period, total, bySeverity }: SeverityOverviewProps) {
  const cards: SeverityCardConfig[] = [
    {
      key: 'RED',
      title: 'Rojos',
      count: bySeverity.RED,
      borderClass: 'border-l-[#DC2626]',
      iconColorClass: 'text-[#DC2626]',
      Icon: AlertOctagon,
    },
    {
      key: 'YELLOW',
      title: 'Amarillos',
      count: bySeverity.YELLOW,
      borderClass: 'border-l-[#D97706]',
      iconColorClass: 'text-[#D97706]',
      Icon: AlertTriangle,
    },
    {
      key: 'GREEN',
      title: 'Verdes',
      count: bySeverity.GREEN,
      borderClass: 'border-l-[#16A34A]',
      iconColorClass: 'text-[#16A34A]',
      Icon: CheckCircle2,
    },
  ];

  return (
    <section
      aria-labelledby="severity-overview-heading"
      className="col-span-12 space-y-3"
      data-testid="severity-overview"
    >
      <h2 id="severity-overview-heading" className="text-sm font-semibold uppercase tracking-wide text-app-text-muted">
        Estado de los clientes
      </h2>
      <div className="grid grid-cols-12 gap-4">
        {cards.map(({ key, title, count, borderClass, iconColorClass, Icon }) => {
          const pct = total > 0 ? Math.round((count / total) * 100) : 0;
          const href = `/contador?period=${encodeURIComponent(period)}&severity=${key}`;
          return (
            <div
              key={key}
              className={`col-span-12 flex flex-col justify-between rounded-lg border border-app-border border-l-4 ${borderClass} bg-app-surface p-5 md:col-span-4`}
            >
              <div>
                <div className="flex items-center justify-between">
                  <span className="text-sm font-semibold text-app-text">{title}</span>
                  <Icon className={`h-5 w-5 ${iconColorClass}`} aria-hidden="true" />
                </div>
                <div className="mt-2 text-3xl font-semibold tabular-nums text-app-text">
                  {count}
                </div>
                <div className="mt-1 text-sm text-app-text-muted">
                  {pct}% del total
                </div>
              </div>
              <Link
                to={href}
                aria-label={`Ver ${title} en la tabla`}
                className="mt-4 inline-flex items-center gap-1 text-sm font-medium text-app-primary hover:underline focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-app-focus"
              >
                <span>Ver en la tabla</span>
                <ArrowRight className="h-4 w-4" aria-hidden="true" />
              </Link>
            </div>
          );
        })}
      </div>
    </section>
  );
}

