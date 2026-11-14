import { z } from 'zod';
import { apiFetch, jsonBody } from '../../shared/api/http';
import {
  DemoClient,
  DemoClientSchema,
  ClientProfile,
  ClientProfileSchema,
  ClientClosing,
  ClientClosingSchema,
  NotificationItem,
  NotificationSchema,
  ImportResult,
  ImportResultSchema,
  ExternalIncomeItem,
} from './schemas';

export async function fetchDemoClients(): Promise<DemoClient[]> {
  return apiFetch('/api/v1/demo/clients', z.array(DemoClientSchema));
}

export async function fetchProfile(): Promise<ClientProfile> {
  return apiFetch('/api/v1/me', ClientProfileSchema);
}

export async function fetchMyClosings(): Promise<ClientClosing[]> {
  return apiFetch('/api/v1/me/closings', z.array(ClientClosingSchema));
}

export async function markClosingPaid(closingId: string): Promise<{ closingId: string; paidAt: string }> {
  const schema = z.object({
    closingId: z.string(),
    paidAt: z.string(),
  });
  return apiFetch(`/api/v1/me/closings/${closingId}/paid`, schema, { method: 'POST' });
}

export async function fetchMyNotifications(): Promise<NotificationItem[]> {
  return apiFetch('/api/v1/me/notifications', z.array(NotificationSchema));
}

export async function markNotificationRead(id: string): Promise<{ id: string; status: string }> {
  const schema = z.object({
    id: z.string(),
    status: z.string(),
  });
  return apiFetch(`/api/v1/me/notifications/${id}/read`, schema, { method: 'POST' });
}

export async function importInvoicesCSV(file: File): Promise<ImportResult> {
  const formData = new FormData();
  formData.append('file', file);
  return apiFetch('/api/v1/me/invoices/import', ImportResultSchema, {
    method: 'POST',
    body: formData,
  });
}

export interface ManualInvoicePayload {
  type: string;
  pointOfSale: number;
  number: number;
  issueDate: string;
  currency: string;
  netAmount: { amount: number; currency: string };
  vatAmount?: { amount: number; currency: string };
  totalAmount: { amount: number; currency: string };
  receiverDoc?: string;
  receiverName?: string;
}

export async function createManualInvoice(payload: ManualInvoicePayload): Promise<{ id: string }> {
  return apiFetch('/api/v1/me/invoices', z.object({ id: z.string() }), jsonBody('POST', payload));
}

export interface PurchaseInvoicePayload {
  type: string;
  pointOfSale: number;
  number: number;
  issueDate: string;
  currency: string;
  netAmount: { amount: number; currency: string };
  vatAmount: { amount: number; currency: string };
  totalAmount: { amount: number; currency: string };
  supplierDoc: string;
  supplierName: string;
}

export async function createPurchaseInvoice(payload: PurchaseInvoicePayload): Promise<{ id: string }> {
  return apiFetch('/api/v1/me/purchase-invoices', z.object({ id: z.string() }), jsonBody('POST', payload));
}

export async function saveExternalIncome(
  period: string,
  data: { declaredNoExternalIncome: boolean; externalIncome: ExternalIncomeItem[] }
): Promise<void> {
  await apiFetch(
    `/api/v1/me/periods/${period}/external-income`,
    z.any(),
    jsonBody('PUT', data)
  );
}

export async function confirmPeriodInput(period: string): Promise<{ confirmedAt: string }> {
  return apiFetch(
    `/api/v1/me/periods/${period}/confirm`,
    z.object({ confirmedAt: z.string() }),
    { method: 'POST' }
  );
}

