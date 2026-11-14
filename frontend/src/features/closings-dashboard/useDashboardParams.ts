import { useMemo, useCallback } from 'react';
import { useSearchParams } from 'react-router-dom';

export interface DashboardFilters {
  period: string;
  severity: string[];
  signal: string[];
  regime: string;
  status: string[];
  q: string;
  sort: 'severity' | 'name' | 'billed' | 'toPay' | 'pctKCap';
  dir: 'asc' | 'desc';
  page: number;
  pageSize: number;
}

export const DEFAULT_PERIOD = '2026-10';

export const DEFAULT_FILTERS: DashboardFilters = {
  period: DEFAULT_PERIOD,
  severity: [],
  signal: [],
  regime: '',
  status: [],
  q: '',
  sort: 'severity',
  dir: 'desc',
  page: 1,
  pageSize: 50,
};

export function parseFiltersFromSearchParams(
  searchParams: URLSearchParams,
  defaultPeriod: string = DEFAULT_PERIOD
): DashboardFilters {
  const period = searchParams.get('period') || defaultPeriod;
  const severityRaw = searchParams.get('severity');
  const severity = severityRaw ? severityRaw.split(',').filter(Boolean) : [];

  const signalRaw = searchParams.get('signal');
  const signal = signalRaw ? signalRaw.split(',').filter(Boolean) : [];

  const regime = searchParams.get('regime') || '';

  const statusRaw = searchParams.get('status');
  const status = statusRaw ? statusRaw.split(',').filter(Boolean) : [];

  const q = searchParams.get('q') || '';

  const sortRaw = searchParams.get('sort');
  const sort =
    sortRaw === 'name' ||
    sortRaw === 'billed' ||
    sortRaw === 'toPay' ||
    sortRaw === 'pctKCap'
      ? sortRaw
      : 'severity';

  const defaultDir = sort === 'name' ? 'asc' : 'desc';
  const dirRaw = searchParams.get('dir');
  const dir = dirRaw === 'asc' || dirRaw === 'desc' ? dirRaw : defaultDir;

  const pageRaw = searchParams.get('page');
  const page = pageRaw ? Math.max(1, parseInt(pageRaw, 10) || 1) : 1;

  const pageSizeRaw = searchParams.get('pageSize');
  const pageSize = pageSizeRaw ? Math.max(10, parseInt(pageSizeRaw, 10) || 50) : 50;

  return {
    period,
    severity,
    signal,
    regime,
    status,
    q,
    sort,
    dir,
    page,
    pageSize,
  };
}

export function filtersToSearchParams(filters: Partial<DashboardFilters>): URLSearchParams {
  const sp = new URLSearchParams();

  if (filters.period && filters.period !== DEFAULT_PERIOD) {
    sp.set('period', filters.period);
  }
  if (filters.severity && filters.severity.length > 0) {
    sp.set('severity', filters.severity.join(','));
  }
  if (filters.signal && filters.signal.length > 0) {
    sp.set('signal', filters.signal.join(','));
  }
  if (filters.regime) {
    sp.set('regime', filters.regime);
  }
  if (filters.status && filters.status.length > 0) {
    sp.set('status', filters.status.join(','));
  }
  if (filters.q) {
    sp.set('q', filters.q);
  }
  if (filters.sort && filters.sort !== 'severity') {
    sp.set('sort', filters.sort);
  }
  const defaultDir = (filters.sort || 'severity') === 'name' ? 'asc' : 'desc';
  if (filters.dir && filters.dir !== defaultDir) {
    sp.set('dir', filters.dir);
  }
  if (filters.page && filters.page > 1) {
    sp.set('page', filters.page.toString());
  }
  if (filters.pageSize && filters.pageSize !== 50) {
    sp.set('pageSize', filters.pageSize.toString());
  }

  return sp;
}

export function filtersToQueryString(filters: Partial<DashboardFilters>): string {
  const sp = filtersToSearchParams(filters);
  const str = sp.toString();
  return str ? `?${str}` : '';
}

export function useDashboardParams(defaultPeriod: string = DEFAULT_PERIOD) {
  const [searchParams, setSearchParams] = useSearchParams();

  const filters = useMemo(() => {
    return parseFiltersFromSearchParams(searchParams, defaultPeriod);
  }, [searchParams, defaultPeriod]);

  const updateFilters = useCallback(
    (newPartial: Partial<DashboardFilters>, resetPage: boolean = true) => {
      setSearchParams(
        (prev) => {
          const current = parseFiltersFromSearchParams(prev, defaultPeriod);
          const next: DashboardFilters = {
            ...current,
            ...newPartial,
            page: resetPage ? 1 : newPartial.page ?? current.page,
          };
          return filtersToSearchParams(next);
        },
        { replace: true }
      );
    },
    [setSearchParams, defaultPeriod]
  );

  const setPeriod = useCallback(
    (period: string) => updateFilters({ period }, true),
    [updateFilters]
  );

  const setSeverity = useCallback(
    (severity: string[]) => updateFilters({ severity }, true),
    [updateFilters]
  );

  const toggleSeverity = useCallback(
    (color: string) => {
      setSearchParams(
        (prev) => {
          const current = parseFiltersFromSearchParams(prev, defaultPeriod);
          const hasColor = current.severity.includes(color);
          const nextSeverity = hasColor
            ? current.severity.filter((c) => c !== color)
            : [...current.severity, color];

          return filtersToSearchParams({
            ...current,
            severity: nextSeverity,
            page: 1,
          });
        },
        { replace: true }
      );
    },
    [setSearchParams, defaultPeriod]
  );

  const setSignal = useCallback(
    (signal: string[]) => updateFilters({ signal }, true),
    [updateFilters]
  );

  const setRegime = useCallback(
    (regime: string) => updateFilters({ regime }, true),
    [updateFilters]
  );

  const setStatus = useCallback(
    (status: string[]) => updateFilters({ status }, true),
    [updateFilters]
  );

  const setQ = useCallback(
    (q: string) => updateFilters({ q }, true),
    [updateFilters]
  );

  const setSort = useCallback(
    (sort: DashboardFilters['sort'], dir?: 'asc' | 'desc') => {
      setSearchParams(
        (prev) => {
          const current = parseFiltersFromSearchParams(prev, defaultPeriod);
          let newDir = dir;
          if (!newDir) {
            const defaultDir = sort === 'name' ? 'asc' : 'desc';
            newDir = current.sort === sort ? (current.dir === 'asc' ? 'desc' : 'asc') : defaultDir;
          }
          return filtersToSearchParams({
            ...current,
            sort,
            dir: newDir,
            page: 1,
          });
        },
        { replace: true }
      );
    },
    [setSearchParams, defaultPeriod]
  );

  const setPage = useCallback(
    (page: number) => updateFilters({ page }, false),
    [updateFilters]
  );

  const setPageSize = useCallback(
    (pageSize: number) => updateFilters({ pageSize }, true),
    [updateFilters]
  );

  const clearFilters = useCallback(() => {
    setSearchParams(
      (prev) => {
        const current = parseFiltersFromSearchParams(prev, defaultPeriod);
        return filtersToSearchParams({
          period: current.period,
          severity: [],
          signal: [],
          regime: '',
          status: [],
          q: '',
          sort: 'severity',
          dir: 'desc',
          page: 1,
          pageSize: current.pageSize,
        });
      },
      { replace: true }
    );
  }, [setSearchParams, defaultPeriod]);

  const toQueryString = useCallback(() => {
    return filtersToQueryString(filters);
  }, [filters]);

  return {
    filters,
    setPeriod,
    setSeverity,
    toggleSeverity,
    setSignal,
    setRegime,
    setStatus,
    setQ,
    setSort,
    setPage,
    setPageSize,
    clearFilters,
    toQueryString,
  };
}

