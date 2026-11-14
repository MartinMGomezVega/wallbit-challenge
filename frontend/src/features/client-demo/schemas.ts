import { z } from 'zod';
import {
  PublishedSchema,
  PublishedDTO,
  ValidatedBySchema,
  ValidatedByDTO,
  PublishedFiguresSchema,
  PublishedFigureItemSchema,
  ClientPeriodInputDTOSchema,
  ClientPeriodInputDTO,
  ImportResultSchema,
  ImportResultDTO as ImportResult,
} from '../closing-review/schemas';

export {
  PublishedSchema,
  ValidatedBySchema,
  PublishedFiguresSchema,
  PublishedFigureItemSchema,
  ClientPeriodInputDTOSchema,
  ImportResultSchema,
};
export type { PublishedDTO, ValidatedByDTO, ClientPeriodInputDTO, ImportResult };

export const ClientProfileSchema = z.object({
  id: z.string(),
  name: z.string(),
  cuitMasked: z.string(),
  regime: z.string(),
  category: z.string().optional(),
});
export type ClientProfile = z.infer<typeof ClientProfileSchema>;

export const ClientClosingInputSchema = z.object({
  answered: z.boolean(),
  confirmed: z.boolean(),
  invoicesCount: z.number(),
});
export type ClientClosingInput = z.infer<typeof ClientClosingInputSchema>;

export const ClientClosingSchema = z.object({
  closingId: z.string(),
  period: z.string(),
  state: z.enum(['ACTION_REQUIRED', 'IN_REVIEW', 'READY']),
  updating: z.boolean().optional(),
  published: PublishedSchema.optional(),
  paidAt: z.string().nullable().optional(),
  input: ClientClosingInputSchema,
});
export type ClientClosing = z.infer<typeof ClientClosingSchema>;

export const ClientClosingsResponseSchema = z.array(ClientClosingSchema);
export type ClientClosingsResponse = z.infer<typeof ClientClosingsResponseSchema>;

export const InvoiceMoneySchema = z.object({
  amount: z.string(),
  currency: z.string(),
});

export const ClientInvoiceSchema = z.object({
  id: z.string(),
  clientId: z.string(),
  type: z.string(),
  pointOfSale: z.number(),
  number: z.number(),
  issueDate: z.string(),
  currency: z.string(),
  netAmount: InvoiceMoneySchema,
  vatAmount: InvoiceMoneySchema,
  totalAmount: InvoiceMoneySchema,
  fxRate: z.string().nullable().optional(),
  receiverName: z.string(),
  receiverDoc: z.string().nullable().optional(),
  source: z.string().optional(),
  uploadedBy: z.string().optional(),
  createdAt: z.string().optional(),
});
export type ClientInvoice = z.infer<typeof ClientInvoiceSchema>;

export const ClientInvoicesResponseSchema = z.array(ClientInvoiceSchema);
export type ClientInvoicesResponse = z.infer<typeof ClientInvoicesResponseSchema>;

export const MarkPaidResponseSchema = z.object({
  closingId: z.string(),
  paidAt: z.string(),
});
export type MarkPaidResponse = z.infer<typeof MarkPaidResponseSchema>;
