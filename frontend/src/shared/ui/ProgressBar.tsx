import { formatMoney, MoneyDTO } from '../format/money';
import { formatPercentBp } from '../format/percent';
import { Tooltip } from './Tooltip';

export interface ProgressBarProps {
  pctBp: number;
  arsAmount?: MoneyDTO;
  className?: string;
}

export function ProgressBar({ pctBp, arsAmount, className = '' }: ProgressBarProps) {
  const pct = Math.max(0, Math.round(pctBp / 100));
  const clampedWidth = Math.min(100, pct);

  let fillColor = 'bg-severity-green-icon';
  if (pct >= 100) {
    fillColor = 'bg-severity-red-icon';
  } else if (pct >= 80) {
    fillColor = 'bg-severity-yellow-fill';
  }

  const tooltipText = arsAmount
    ? `${formatMoney(arsAmount)} (${formatPercentBp(pctBp)} del tope K)`
    : `${formatPercentBp(pctBp)} del tope K`;

  const bar = (
    <div
      className={`inline-flex items-center gap-2 ${className}`}
      role="progressbar"
      aria-valuenow={pct}
      aria-valuemin={0}
      aria-valuemax={100}
      aria-label={`Acumulado al ${pct}% del tope de Monotributo`}
    >
      <div className="h-1.5 w-24 overflow-hidden rounded-full bg-app-border">
        <div
          className={`h-full rounded-full transition-all duration-300 ${fillColor}`}
          style={{ width: `${clampedWidth}%` }}
        />
      </div>
      <span className="w-10 text-right text-xs font-medium tabular-nums text-app-text-muted">
        {formatPercentBp(pctBp)}
      </span>
    </div>
  );

  return <Tooltip content={tooltipText}>{bar}</Tooltip>;
}

