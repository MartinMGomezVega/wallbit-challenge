import { z } from 'zod';
import { MoneySchema } from '../closings-dashboard/schemas';

export const ClosingIDsResponseSchema = z.object({
  ids: z.array(z.string()),
  total: z.number(),
});
export type ClosingIDsResponse = z.infer<typeof ClosingIDsResponseSchema>;

// Movement DTO
export const MovementDTOSchema = z.object({
  id: z.string(),
  clientId: z.string(),
  type: z.string(),
  amount: MoneySchema,
  counterparty: z
    .union([z.string(), z.object({ name: z.string() }).transform((c) => c.name)])
    .optional()
    .default(''),
  description: z.string().optional().default(''),
  occurredAt: z.string(),
});
export type MovementDTO = z.infer<typeof MovementDTOSchema>;

// Classification DTO
export const ClassificationDTOSchema = z
  .object({
    nature: z.enum(['INCOME', 'EXPENSE', 'TRANSFER', 'OTHER', 'NOT_INCOME', 'UNCLASSIFIED']).optional(),
    kind: z.enum(['INCOME', 'NOT_INCOME', 'UNCLASSIFIED']).optional(),
    rule: z.string(),
    confidence: z.union([z.number(), z.string()]).optional().default(1),
    comment: z.string().optional().default(''),
    note: z.string().optional().default(''),
    overridden: z.boolean().optional(),
    overriddenBy: z.string().optional(),
  })
  .transform((c) => ({
    nature: c.nature ?? c.kind ?? 'OTHER',
    rule: c.rule,
    confidence: c.confidence,
    comment: c.comment || c.note || '',
    overridden: c.overridden ?? Boolean(c.overriddenBy),
  }));
export type ClassificationDTO = z.infer<typeof ClassificationDTOSchema>;

export const PeriodMovementDTOSchema = z.object({
  movement: MovementDTOSchema,
  classification: ClassificationDTOSchema,
});
export type PeriodMovementDTO = z.infer<typeof PeriodMovementDTOSchema>;

// Invoice DTO
export const InvoiceDTOSchema = z.object({
  id: z.string(),
  clientId: z.string(),
  pointOfSale: z.number(),
  number: z.number(),
  type: z.string(),
  concept: z.string().optional().default(''),
  issueDate: z.string(),
  totalAmount: MoneySchema,
  currency: z.enum(['ARS', 'USD']),
  cae: z.string().optional().default(''),
  caeDueDate: z.string().optional().default(''),
});
export type InvoiceDTO = z.infer<typeof InvoiceDTOSchema>;

export const PeriodInvoiceDTOSchema = z.object({
  invoice: InvoiceDTOSchema,
  amountArs: MoneySchema,
  fxRateUsed: z.number().optional().nullable(),
});
export type PeriodInvoiceDTO = z.infer<typeof PeriodInvoiceDTOSchema>;

// Reconciliation DTO
export const MatchedPairSchema = z.object({
  movement: MovementDTOSchema,
  invoice: InvoiceDTOSchema,
  diffAmount: z.number().optional().default(0),
});
export type MatchedPair = z.infer<typeof MatchedPairSchema>;

export const ReconciliationDetailSchema = z.object({
  matchedPairs: z.array(MatchedPairSchema).default([]),
  unmatchedIncome: z.array(MovementDTOSchema).default([]),
  unmatchedInvoices: z.array(InvoiceDTOSchema).default([]),
});
export type ReconciliationDetail = z.infer<typeof ReconciliationDetailSchema>;

// Monotributo Assessment
export const MonotributoAssessmentSchema = z.object({
  billedInPeriodUSD: MoneySchema,
  billedInPeriodARS: MoneySchema,
  currentCategory: z.string(),
  requiredCategory: z.string().nullable().optional(),
  billed12mARS: MoneySchema,
  kCapARS: MoneySchema.optional(),
  pctOfKCapBp: z.number(),
  currentCapARS: MoneySchema.optional(),
  fee: MoneySchema,
  feeDueDate: z.string(),
  exclusionReason: z.string().optional(),
  exclusionEffectiveDate: z.string().optional(),
});
export type MonotributoAssessment = z.infer<typeof MonotributoAssessmentSchema>;

// RI Assessment
export const RIAssessmentSchema = z.object({
  billedInPeriodUSD: MoneySchema,
  billedInPeriodARS: MoneySchema,
  debitoFiscalARS: MoneySchema,
  creditoFiscalARS: MoneySchema,
  saldoTecnicoARS: MoneySchema,
  retencionesPercepcionesARS: MoneySchema,
  ivaToPay: MoneySchema,
  vatDueDate: z.string(),
  autonomousCategory: z.string(),
  autonomousFee: MoneySchema,
  selfEmployedDueDate: z.string(),
  incomeTaxAdvance: MoneySchema,
  incomeTaxAdvanceDueDate: z.string(),
  totalToPay: MoneySchema,
});
export type RIAssessment = z.infer<typeof RIAssessmentSchema>;

// Client DTO
export const ClientDTOSchema = z.object({
  id: z.string(),
  name: z.string(),
  cuit: z.string(),
  regime: z.enum(['MONOTRIBUTO', 'RI']),
  monotributoCategory: z.string().optional().default(''),
});
export type ClientDTO = z.infer<typeof ClientDTOSchema>;

// Published snapshot
export const PublishedSchema = z.object({
  version: z.number(),
  title: z.string(),
  body: z.string(),
  figures: z.object({
    billedUSD: MoneySchema,
    billedARS: MoneySchema,
    items: z.array(z.object({
      concept: z.string(),
      amount: MoneySchema,
      dueDate: z.string(),
    })).default([]),
    total: MoneySchema,
  }).optional(),
  publishedAt: z.string(),
});
export type PublishedDTO = z.infer<typeof PublishedSchema>;

// Closing Detail
export const ClosingDetailDTOSchema = z.object({
  closing: z.object({
    id: z.string(),
    clientId: z.string(),
    period: z.string(),
    status: z.enum(['PENDING_CLIENT_INPUT', 'READY_FOR_REVIEW', 'APPROVED', 'NOTIFIED']),
    version: z.number(),
    resultHash: z.string(),
    severity: z.union([z.enum(['RED', 'YELLOW', 'GREEN']), z.number()]).transform((s) => {
      if (s === 2 || s === 'RED') return 'RED' as const;
      if (s === 1 || s === 'YELLOW') return 'YELLOW' as const;
      return 'GREEN' as const;
    }),
    description: z.string(),
    signals: z.array(z.object({
      code: z.string(),
      severity: z.union([z.enum(['RED', 'YELLOW', 'GREEN']), z.number()]).transform((s) => {
        if (s === 2 || s === 'RED') return 'RED' as const;
        if (s === 1 || s === 'YELLOW') return 'YELLOW' as const;
        return 'GREEN' as const;
      }),
      message: z.string().optional(),
      data: z.record(z.unknown()).optional(),
    })).default([]),
    mono: MonotributoAssessmentSchema.nullable().optional(),
    ri: RIAssessmentSchema.nullable().optional(),
    computedAt: z.string(),
    paidAt: z.string().nullable().optional(),
    approvedAt: z.string().nullable().optional(),
    approvedBy: z.string().nullable().optional(),
    published: PublishedSchema.nullable().optional(),
  }),
  client: ClientDTOSchema,
  periodMovements: z.array(PeriodMovementDTOSchema).default([]),
  periodInvoices: z.array(PeriodInvoiceDTOSchema).default([]),
  reconciliation: ReconciliationDetailSchema,
  clientInput: z.record(z.unknown()).optional().default({}),
  draftMessage: z.object({
    title: z.string(),
    body: z.string(),
  }).optional().nullable(),
});
export type ClosingDetailDTO = z.infer<typeof ClosingDetailDTOSchema>;

export const MessageSchema = z.object({
  title: z.string(),
  body: z.string(),
});
export type MessageDTO = z.infer<typeof MessageSchema>;

export const ApproveClosingResponseSchema = z.object({
  id: z.string(),
  version: z.number(),
  status: z.string(),
  warning: z.string().optional(),
}).passthrough();
export type ApproveClosingResponse = z.infer<typeof ApproveClosingResponseSchema>;

export const ApproveBatchResponseSchema = z.object({
  approved: z.number(),
  failed: z.number(),
  results: z.array(
    z.object({
      id: z.string(),
      ok: z.boolean(),
      error: z.string().optional(),
    })
  ),
});
export type ApproveBatchResponse = z.infer<typeof ApproveBatchResponseSchema>;

export const UpdateClassificationResponseSchema = z.object({
  id: z.string(),
  movementId: z.string(),
  nature: z.enum(['INCOME', 'EXPENSE', 'TRANSFER', 'OTHER']),
  rule: z.string(),
  comment: z.string().optional().default(''),
  overridden: z.boolean(),
  updatedAt: z.string(),
  updatedBy: z.string(),
});
export type UpdateClassificationResponse = z.infer<typeof UpdateClassificationResponseSchema>;
