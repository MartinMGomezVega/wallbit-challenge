import { z } from 'zod';

export const MoneySchema = z.object({
  amount: z.number(),
  currency: z.string(),
});
export type Money = z.infer<typeof MoneySchema>;

export const DemoClientSchema = z.object({
  id: z.string(),
  name: z.string(),
  regime: z.string(),
});
export type DemoClient = z.infer<typeof DemoClientSchema>;

export const ClientProfileSchema = z.object({
  id: z.string(),
  name: z.string(),
  cuitMasked: z.string(),
  regime: z.string(),
  category: z.string().optional(),
});
export type ClientProfile = z.infer<typeof ClientProfileSchema>;

export const PublishedFigureItemSchema = z.object({
  concept: z.string(),
  amount: MoneySchema,
  dueDate: z.string(),
});
export type PublishedFigureItem = z.infer<typeof PublishedFigureItemSchema>;

export const PublishedFiguresSchema = z.object({
  billedUSD: MoneySchema,
  billedARS: MoneySchema,
  items: z.array(PublishedFigureItemSchema),
  total: MoneySchema,
});
export type PublishedFigures = z.infer<typeof PublishedFiguresSchema>;

export const PublishedSchema = z.object({
  version: z.number(),
  title: z.string(),
  body: z.string(),
  figures: PublishedFiguresSchema,
  publishedAt: z.string(),
});
export type Published = z.infer<typeof PublishedSchema>;

export const ClientClosingInputSchema = z.object({
  answered: z.boolean(),
  confirmed: z.boolean(),
  invoicesCount: z.number(),
});
export type ClientClosingInput = z.infer<typeof ClientClosingInputSchema>;

export const ClientClosingStateSchema = z.enum(['ACTION_REQUIRED', 'IN_REVIEW', 'READY']);
export type ClientClosingState = z.infer<typeof ClientClosingStateSchema>;

export const ClientClosingSchema = z.object({
  closingId: z.string(),
  period: z.string(),
  state: ClientClosingStateSchema,
  updating: z.boolean().optional(),
  published: PublishedSchema.nullable().optional(),
  paidAt: z.string().nullable().optional(),
  input: ClientClosingInputSchema,
});
export type ClientClosing = z.infer<typeof ClientClosingSchema>;

export const NotificationSchema = z.object({
  id: z.string(),
  clientId: z.string(),
  closingId: z.string().optional(),
  kind: z.string(),
  title: z.string(),
  body: z.string(),
  readAt: z.string().nullable().optional(),
  createdAt: z.string(),
});
export type NotificationItem = z.infer<typeof NotificationSchema>;

export const LineIssueSchema = z.object({
  line: z.number(),
  message: z.string(),
});
export type LineIssue = z.infer<typeof LineIssueSchema>;

export const ImportResultSchema = z.object({
  imported: z.number(),
  skippedDuplicates: z.number(),
  warnings: z.array(LineIssueSchema),
  errors: z.array(LineIssueSchema),
});
export type ImportResult = z.infer<typeof ImportResultSchema>;

export const ExternalIncomeItemSchema = z.object({
  amount: MoneySchema,
  description: z.string(),
});
export type ExternalIncomeItem = z.infer<typeof ExternalIncomeItemSchema>;

