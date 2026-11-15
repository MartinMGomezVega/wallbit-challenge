import { apiFetch } from '../../shared/http/http';
import { PeriodStats, PeriodStatsSchema } from './schemas';

export async function getPeriodStats(period: string): Promise<PeriodStats> {
  return apiFetch(`/v1/periods/${encodeURIComponent(period)}/stats`, PeriodStatsSchema);
}

