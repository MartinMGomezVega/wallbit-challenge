import { parseSeverity, SEVERITY_CONFIG, SeverityType } from './SeverityIndicator';
import { Tooltip } from './Tooltip';

export interface SignalItem {
  code: string;
  label: string;
  severity?: SeverityType;
}

export interface SignalChipProps {
  signal: SignalItem;
  className?: string;
}

export function SignalChip({ signal, className = '' }: SignalChipProps) {
  const norm = parseSeverity(signal.severity ?? 'YELLOW');
  const cfg = SEVERITY_CONFIG[norm];

  return (
    <span
      className={`inline-flex max-w-full truncate items-center rounded-control border px-2 py-0.5 text-xs font-medium ${cfg.bgClass} ${cfg.borderClass} ${cfg.textClass} ${className}`}
      title={signal.label}
    >
      <span className="truncate">{signal.label}</span>
    </span>
  );
}

export interface OtherSignalsChipProps {
  signals: SignalItem[];
  className?: string;
}

export function OtherSignalsChip({ signals, className = '' }: OtherSignalsChipProps) {
  if (!signals || signals.length === 0) return null;

  const tooltipContent = (
    <div className="space-y-1 py-0.5">
      <div className="font-semibold text-white/90">Otras señales:</div>
      <ul className="space-y-1">
        {signals.map((s, idx) => {
          const norm = parseSeverity(s.severity ?? 'YELLOW');
          const dotColor =
            norm === 'RED'
              ? 'bg-severity-red-icon'
              : norm === 'YELLOW'
              ? 'bg-severity-yellow-icon'
              : 'bg-severity-green-icon';
          return (
            <li key={idx} className="flex items-center gap-1.5 text-xs">
              <span className={`h-2 w-2 rounded-full ${dotColor}`} aria-hidden="true" />
              <span>{s.label || s.code}</span>
            </li>
          );
        })}
      </ul>
    </div>
  );

  return (
    <Tooltip content={tooltipContent}>
      <span
        tabIndex={0}
        role="button"
        aria-label={`${signals.length} otras señales`}
        className={`inline-flex items-center rounded-control border border-app-border bg-app-surface-muted px-1.5 py-0.5 text-xs font-medium text-app-text-muted hover:bg-app-border/40 focus:outline-none focus:ring-2 focus:ring-app-focus ${className}`}
      >
        +{signals.length}
      </span>
    </Tooltip>
  );
}

