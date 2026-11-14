import { apiFetch } from '../../shared/api/http';
import { PeriodStats, PeriodStatsSchema } from './schemas';

export async function getPeriodStats(period: string): Promise<PeriodStats> {
  return apiFetch(`/api/v1/accountant/periods/${encodeURIComponent(period)}/stats`, PeriodStatsSchema);
}

