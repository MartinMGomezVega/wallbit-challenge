import { apiFetch } from '../../shared/api/http';
import {
  ClosingIDsResponse,
  ClosingIDsResponseSchema,
  ClosingDetailDTO,
  ClosingDetailDTOSchema,
  UpdateClassificationResponse,
  UpdateClassificationResponseSchema,
  ApproveClosingResponse,
  ApproveClosingResponseSchema,
  ApproveBatchResponse,
  ApproveBatchResponseSchema,
} from './schemas';
import { z } from 'zod';

export async function getClosingIds(
  period: string,
  params: URLSearchParams | Record<string, string | number | undefined> = {}
): Promise<ClosingIDsResponse> {
  const search =
    params instanceof URLSearchParams
      ? params.toString()
      : new URLSearchParams(
          Object.entries(params)
            .filter(([k, v]) => k !== 'page' && k !== 'pageSize' && v !== undefined && v !== '')
            .map(([k, v]) => [k, String(v)])
        ).toString();

  const qs = search ? `?${search}` : '';
  return apiFetch(
    `/api/v1/accountant/periods/${period}/closings/ids${qs}`,
    ClosingIDsResponseSchema
  );
}

export async function getClosingDetail(closingId: string): Promise<ClosingDetailDTO> {
  return apiFetch(
    `/api/v1/accountant/closings/${closingId}`,
    ClosingDetailDTOSchema
  );
}

export async function updateMovementClassification(
  movementId: string,
  nature: 'INCOME' | 'OTHER',
  comment: string = 'Modificado por el contador'
): Promise<UpdateClassificationResponse> {
  return apiFetch(
    `/api/v1/accountant/movements/${movementId}/classification`,
    UpdateClassificationResponseSchema,
    {
      method: 'PATCH',
      body: JSON.stringify({ nature, comment }),
    }
  );
}

export async function confirmRegimeTransition(
  clientId: string,
  effectiveDate: string,
  force: boolean = false,
  note?: string
): Promise<unknown> {
  return apiFetch(
    `/api/v1/accountant/clients/${clientId}/regime-change`,
    z.unknown(),
    {
      method: 'POST',
      body: JSON.stringify({
        to: 'RI',
        effectiveDate,
        force,
        note,
      }),
    }
  );
}

export async function approveClosing(
  closingId: string,
  version: number,
  clientNote?: string
): Promise<ApproveClosingResponse> {
  return apiFetch(
    `/api/v1/accountant/closings/${closingId}/approve`,
    ApproveClosingResponseSchema,
    {
      method: 'POST',
      body: JSON.stringify({
        version,
        clientNote: clientNote || undefined,
      }),
    }
  );
}

export async function approveBatch(
  period: string,
  items: Array<{ id: string; version: number }>
): Promise<ApproveBatchResponse> {
  return apiFetch(
    `/api/v1/accountant/periods/${period}/closings/approve-batch`,
    ApproveBatchResponseSchema,
    {
      method: 'POST',
      body: JSON.stringify({ items }),
    }
  );
}

