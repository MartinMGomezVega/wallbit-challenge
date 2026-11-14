import { useQuery, keepPreviousData } from '@tanstack/react-query';
import { listClosings, getClosingsSummary, getSignalsCatalog } from './api';
import { DashboardFilters, filtersToSearchParams } from './useDashboardParams';

export function useClosingsList(filters: DashboardFilters) {
  return useQuery({
    queryKey: ['closings', filters],
    queryFn: () => listClosings(filters.period, filtersToSearchParams(filters)),
    placeholderData: keepPreviousData,
  });
}

export function useClosingsSummary(period: string, filters: Partial<DashboardFilters>) {
  return useQuery({
    queryKey: ['closings-summary', period, filters.regime, filters.q],
    queryFn: () => getClosingsSummary(period, filtersToSearchParams(filters)),
  });
}

export function useSignalsCatalog() {
  return useQuery({
    queryKey: ['signals-catalog'],
    queryFn: getSignalsCatalog,
    staleTime: 1000 * 60 * 60, // 1 hora de caché
  });
}
