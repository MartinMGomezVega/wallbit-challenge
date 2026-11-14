import { useSearchParams } from 'react-router-dom';
import { formatPeriodName } from '../shared/format/date';

export const DEFAULT_PERIOD = '2026-10';
export const AVAILABLE_PERIODS = ['2026-10', '2026-09', '2026-08'];

export function PeriodSelect() {
  const [searchParams, setSearchParams] = useSearchParams();
  const period = searchParams.get('period') || DEFAULT_PERIOD;

  const periodOptions = AVAILABLE_PERIODS.includes(period)
    ? AVAILABLE_PERIODS
    : [period, ...AVAILABLE_PERIODS];

  const handlePeriodChange = (nextPeriod: string) => {
    const next = new URLSearchParams(searchParams);
    next.set('period', nextPeriod);
    if (next.has('page')) {
      next.set('page', '1');
    }
    setSearchParams(next);
  };

  return (
    <div className="flex items-center gap-2">
      <label
        htmlFor="global-period-select"
        className="text-xs font-medium text-neutral-400"
      >
        Período
      </label>
      <select
        id="global-period-select"
        aria-label="Período"
        value={period}
        onChange={(e) => handlePeriodChange(e.target.value)}
        className="h-8 rounded-md border border-[#262626] bg-[#171717] px-2.5 text-sm text-white focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-app-brand"
      >
        {periodOptions.map((p) => (
          <option key={p} value={p}>
            {formatPeriodName(p)}
          </option>
        ))}
      </select>
    </div>
  );
}

