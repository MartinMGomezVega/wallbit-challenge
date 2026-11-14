import { useQuery } from '@tanstack/react-query';
import { getPeriodStats } from './api';

export function useStats(period: string) {
  return useQuery({
    queryKey: ['period-stats', period],
    queryFn: () => getPeriodStats(period),
    enabled: Boolean(period),
  });
}

