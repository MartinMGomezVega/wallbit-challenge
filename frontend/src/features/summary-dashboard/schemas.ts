import { z } from 'zod';

export const MoneySchema = z.object({
  amount: z.number(),
  currency: z.enum(['ARS', 'USD']),
});

export type MoneyDTO = z.infer<typeof MoneySchema>;

export const SeveritySchema = z.enum(['RED', 'YELLOW', 'GREEN']);
export type Severity = z.infer<typeof SeveritySchema>;

export const ClosingStatusSchema = z.enum([
  'PENDING_CLIENT_INPUT',
  'READY_FOR_REVIEW',
  'APPROVED',
  'NOTIFIED',
]);
export type ClosingStatus = z.infer<typeof ClosingStatusSchema>;

export const DueDateTotalSchema = z.object({
  date: z.string(),
  amount: MoneySchema,
  clients: z.number(),
});
export type DueDateTotal = z.infer<typeof DueDateTotalSchema>;

export const SignalStatSchema = z.object({
  code: z.string(),
  label: z.string(),
  severity: SeveritySchema,
  count: z.number(),
});
export type SignalStat = z.infer<typeof SignalStatSchema>;

export const CategoryStatSchema = z.object({
  category: z.string(),
  clients: z.number(),
});
export type CategoryStat = z.infer<typeof CategoryStatSchema>;

export const PeriodStatsSchema = z.object({
  period: z.string(),
  progress: z.object({
    total: z.number(),
    byStatus: z.object({
      PENDING_CLIENT_INPUT: z.number(),
      READY_FOR_REVIEW: z.number(),
      APPROVED: z.number(),
      NOTIFIED: z.number(),
    }),
    approvedPct: z.number(),
    pendingReview: z.object({
      RED: z.number(),
      YELLOW: z.number(),
      GREEN: z.number(),
    }),
    estimatedReviewMinutes: z.number(),
  }),
  totals: z.object({
    billedUSD: MoneySchema,
    billedARS: MoneySchema,
    toPayARS: MoneySchema,
    toPayByDueDate: z.array(DueDateTotalSchema),
  }),
  risk: z.object({
    bySeverity: z.object({
      RED: z.number(),
      YELLOW: z.number(),
      GREEN: z.number(),
    }),
    bySignal: z.array(SignalStatSchema),
    byCategory: z.array(CategoryStatSchema),
    nearKCap: z.number(),
  }),
  success: z.object({
    greenPct: z.number(),
    greenCount: z.number().default(0),
    total: z.number().default(0),
    notifiedBeforeDay10Pct: z.number(),
    notifiedBeforeDay10Count: z.number().default(0),
    notifiedCount: z.number().default(0),
    paidOnTimePct: z.number(),
    paidOnTimeCount: z.number().default(0),
  }),
});

export type PeriodStats = z.infer<typeof PeriodStatsSchema>;

