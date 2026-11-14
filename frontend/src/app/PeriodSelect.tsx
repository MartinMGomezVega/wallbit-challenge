import { useSearchParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { formatPeriodName } from '../shared/format/date';
import { getPeriods } from '../features/closings-dashboard/api';

export const DEFAULT_PERIOD = '2026-10';

export function PeriodSelect() {
  const [searchParams, setSearchParams] = useSearchParams();
  const { data: periods = [DEFAULT_PERIOD] } = useQuery({
    queryKey: ['accountant-periods'],
    queryFn: getPeriods,
    staleTime: 5 * 60 * 1000,
  });

  const periodOptions = periods && periods.length > 0 ? periods : [DEFAULT_PERIOD];
  const urlPeriod = searchParams.get('period');
  const period = urlPeriod || periodOptions[0] || DEFAULT_PERIOD;

  const displayOptions = periodOptions.includes(period)
    ? periodOptions
    : [period, ...periodOptions];

  const handlePeriodChange = (nextPeriod: string) => {
    setSearchParams((prev) => {
      const next = new URLSearchParams(prev);
      next.set('period', nextPeriod);
      if (next.has('page')) {
        next.set('page', '1');
      }
      return next;
    });
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
        {displayOptions.map((p) => (
          <option key={p} value={p}>
            {formatPeriodName(p)}
          </option>
        ))}
      </select>
    </div>
  );
}
