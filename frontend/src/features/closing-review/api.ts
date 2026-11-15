import { apiFetch } from '../../shared/http/http';
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
  ImportResultDTO,
  ImportResultSchema,
  ClientPeriodInputDTO,
  ClientPeriodInputDTOSchema,
} from './schemas';
import { z } from 'zod';

export async function getClosingIds(
  period: string,
  params: URLSearchParams | Record<string, string | number | undefined> = {}
): Promise<ClosingIDsResponse> {
  let sp: URLSearchParams;
  if (params instanceof URLSearchParams) {
    sp = new URLSearchParams(params);
    sp.delete('page');
    sp.delete('pageSize');
  } else {
    sp = new URLSearchParams(
      Object.entries(params)
        .filter(([k, v]) => k !== 'page' && k !== 'pageSize' && v !== undefined && v !== '')
        .map(([k, v]) => [k, String(v)])
    );
  }

  const search = sp.toString();
  const qs = search ? `?${search}` : '';
  return apiFetch(
    `/v1/periods/${period}/closings/ids${qs}`,
    ClosingIDsResponseSchema
  );
}

export async function getClosingDetail(closingId: string): Promise<ClosingDetailDTO> {
  return apiFetch(
    `/v1/closings/${closingId}`,
    ClosingDetailDTOSchema
  );
}

export async function updateMovementClassification(
  movementId: string,
  nature: 'INCOME' | 'OTHER',
  comment: string = 'Modificado por el contador'
): Promise<UpdateClassificationResponse> {
  return apiFetch(
    `/v1/movements/${movementId}/classification`,
    UpdateClassificationResponseSchema,
    {
      method: 'PUT',
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
    `/v1/clients/${clientId}/regime-changes`,
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
    `/v1/closings/${closingId}/approvals`,
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
    `/v1/periods/${period}/approvals`,
    ApproveBatchResponseSchema,
    {
      method: 'POST',
      body: JSON.stringify({ items }),
    }
  );
}

export async function importInvoicesAsAccountant(
  clientId: string,
  file: File
): Promise<ImportResultDTO> {
  const formData = new FormData();
  formData.append('file', file);
  return apiFetch(
    `/v1/clients/${clientId}/invoice-imports`,
    ImportResultSchema,
    {
      method: 'POST',
      body: formData,
    }
  );
}

export type ConfirmInputOnBehalfPayload = {
  externalIncome?: {
    none?: boolean;
    items?: Array<{ amount: { amount: number; currency: 'ARS' | 'USD' }; description: string }>;
  };
  note: string;
};

export async function confirmInputOnBehalf(
  clientId: string,
  period: string,
  payload: ConfirmInputOnBehalfPayload
): Promise<ClientPeriodInputDTO> {
  return apiFetch(
    `/v1/clients/${clientId}/periods/${period}/confirmation`,
    ClientPeriodInputDTOSchema,
    {
      method: 'POST',
      body: JSON.stringify(payload),
    }
  );
}

