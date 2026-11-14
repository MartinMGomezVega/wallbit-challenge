import { z } from 'zod';

export const MoneySchema = z.object({
  amount: z.number(),
  currency: z.enum(['ARS', 'USD']),
});

export type Money = z.infer<typeof MoneySchema>;

export const SignalSummarySchema = z.object({
  code: z.string(),
  label: z.string(),
  severity: z.enum(['RED', 'YELLOW', 'GREEN']).optional(),
});

export type SignalSummary = z.infer<typeof SignalSummarySchema>;

export const BilledInPeriodSchema = z.object({
  usd: MoneySchema,
  ars: MoneySchema,
});

export const Accumulated12mSchema = z.object({
  ars: MoneySchema,
  pctOfKCapBp: z.number(),
});

export const ToPaySchema = z.object({
  amount: MoneySchema,
  dueDate: z.string(),
});

export const ClosingRowSchema = z.object({
  closingId: z.string(),
  clientId: z.string(),
  clientName: z.string(),
  cuit: z.string(),
  regime: z.enum(['MONOTRIBUTO', 'RI']),
  category: z.string().nullable().optional(),
  severity: z
    .union([z.enum(['RED', 'YELLOW', 'GREEN']), z.number()])
    .transform((s) => {
      if (s === 2 || s === 'RED') return 'RED' as const;
      if (s === 1 || s === 'YELLOW') return 'YELLOW' as const;
      return 'GREEN' as const;
    }),
  primarySignal: SignalSummarySchema.nullable().optional(),
  description: z.string(),
  otherSignals: z.array(SignalSummarySchema).default([]),
  billedInPeriod: BilledInPeriodSchema,
  accumulated12m: Accumulated12mSchema,
  toPay: ToPaySchema,
  status: z.enum(['PENDING_CLIENT_INPUT', 'READY_FOR_REVIEW', 'APPROVED', 'NOTIFIED']),
  version: z.number(),
  paidAt: z.string().nullable().optional(),
});

export type ClosingRow = z.infer<typeof ClosingRowSchema>;

export const ClosingsListResponseSchema = z.object({
  items: z.array(ClosingRowSchema),
  total: z.number(),
  page: z.number(),
  pageSize: z.number(),
  totalPages: z.number(),
});

export type ClosingsListResponse = z.infer<typeof ClosingsListResponseSchema>;

export const ClosingsSummaryResponseSchema = z.object({
  total: z.number(),
  bySeverity: z.record(z.string(), z.number()).default({}),
  bySignal: z.record(z.string(), z.number()).default({}),
  byStatus: z.record(z.string(), z.number()).default({}),
});

export type ClosingsSummaryResponse = z.infer<typeof ClosingsSummaryResponseSchema>;

export const SignalCatalogItemSchema = z.object({
  code: z.string(),
  label: z.string(),
  severity: z.enum(['RED', 'YELLOW', 'GREEN']),
  priority: z.number(),
});

export type SignalCatalogItem = z.infer<typeof SignalCatalogItemSchema>;

export const PeriodsResponseSchema = z.object({
  periods: z.array(z.string()),
});

export type PeriodsResponse = z.infer<typeof PeriodsResponseSchema>;
