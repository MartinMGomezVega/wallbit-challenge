import { apiFetch, jsonBody } from '../../shared/api/http';
import {
  ClientProfileSchema,
  ClientProfile,
  ClientClosingsResponseSchema,
  ClientClosing,
  ClientInvoicesResponseSchema,
  ClientInvoice,
  ClientInvoiceSchema,
  ClientPeriodInputDTOSchema,
  ClientPeriodInputDTO,
  ImportResultSchema,
  ImportResult,
  MarkPaidResponseSchema,
  MarkPaidResponse,
} from './schemas';

export interface CreateInvoiceInput {
  type: string;
  pointOfSale: number;
  number: number;
  issueDate: string;
  currency: string;
  netAmount: { amount: string; currency: string };
  vatAmount: { amount: string; currency: string };
  totalAmount: { amount: string; currency: string };
  fxRate?: string;
  receiverName: string;
  receiverDoc?: string;
}

export interface SetExternalIncomeInput {
  items: {
    amount: { amount: string; currency: string };
    description: string;
  }[];
  none: boolean;
}

export async function getClientProfile(clientId: string): Promise<ClientProfile> {
  return apiFetch('/api/v1/me', ClientProfileSchema, {}, { as: clientId });
}

export async function getClientClosings(clientId: string): Promise<ClientClosing[]> {
  return apiFetch('/api/v1/me/closings', ClientClosingsResponseSchema, {}, { as: clientId });
}

export async function getClientInvoices(clientId: string, period?: string): Promise<ClientInvoice[]> {
  const path = period ? `/api/v1/me/invoices?period=${encodeURIComponent(period)}` : '/api/v1/me/invoices';
  return apiFetch(path, ClientInvoicesResponseSchema, {}, { as: clientId });
}

export async function createClientInvoice(clientId: string, data: CreateInvoiceInput): Promise<ClientInvoice> {
  return apiFetch('/api/v1/me/invoices', ClientInvoiceSchema, jsonBody('POST', data), { as: clientId });
}

export async function importClientInvoicesCSV(clientId: string, file: File): Promise<ImportResult> {
  const formData = new FormData();
  formData.append('file', file);
  return apiFetch(
    '/api/v1/me/invoices/import',
    ImportResultSchema,
    {
      method: 'POST',
      body: formData,
    },
    { as: clientId },
  );
}

export async function setClientExternalIncome(
  clientId: string,
  period: string,
  data: SetExternalIncomeInput,
): Promise<ClientPeriodInputDTO> {
  return apiFetch(
    `/api/v1/me/periods/${encodeURIComponent(period)}/external-income`,
    ClientPeriodInputDTOSchema,
    jsonBody('PUT', data),
    { as: clientId },
  );
}

export async function confirmClientPeriod(clientId: string, period: string): Promise<ClientPeriodInputDTO> {
  return apiFetch(
    `/api/v1/me/periods/${encodeURIComponent(period)}/confirm`,
    ClientPeriodInputDTOSchema,
    { method: 'POST' },
    { as: clientId },
  );
}

export async function markClientClosingPaid(clientId: string, closingId: string): Promise<MarkPaidResponse> {
  return apiFetch(
    `/api/v1/me/closings/${encodeURIComponent(closingId)}/paid`,
    MarkPaidResponseSchema,
    { method: 'POST' },
    { as: clientId },
  );
}

