import { z } from 'zod';
import { apiFetch } from '../../shared/api/http';
import {
  ClosingsListResponse,
  ClosingsListResponseSchema,
  ClosingsSummaryResponse,
  ClosingsSummaryResponseSchema,
  SignalCatalogItem,
  SignalCatalogItemSchema,
} from './schemas';

export async function listClosings(
  period: string,
  params: URLSearchParams | Record<string, string | number | undefined>
): Promise<ClosingsListResponse> {
  const search = params instanceof URLSearchParams ? params.toString() : new URLSearchParams(
    Object.entries(params)
      .filter(([, v]) => v !== undefined && v !== '')
      .map(([k, v]) => [k, String(v)])
  ).toString();

  const qs = search ? `?${search}` : '';
  return apiFetch(`/api/v1/accountant/periods/${period}/closings${qs}`, ClosingsListResponseSchema);
}

export async function getClosingsSummary(
  period: string,
  params: URLSearchParams | Record<string, string | number | undefined>
): Promise<ClosingsSummaryResponse> {
  const search = params instanceof URLSearchParams ? params.toString() : new URLSearchParams(
    Object.entries(params)
      .filter(([k, v]) => k !== 'severity' && v !== undefined && v !== '')
      .map(([k, v]) => [k, String(v)])
  ).toString();

  const qs = search ? `?${search}` : '';
  return apiFetch(
    `/api/v1/accountant/periods/${period}/closings/summary${qs}`,
    ClosingsSummaryResponseSchema
  );
}

export async function getSignalsCatalog(): Promise<SignalCatalogItem[]> {
  return apiFetch('/api/v1/accountant/signals/catalog', z.array(SignalCatalogItemSchema));
}

