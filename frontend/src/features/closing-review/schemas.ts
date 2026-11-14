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
  source: z.string().optional(),
  uploadedBy: z.enum(['CLIENT', 'ACCOUNTANT', 'SEED']).optional(),
  receiverName: z.string().optional(),
  receiverDoc: z.string().optional(),
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
  pctOfCurrentCapBp: z.number().optional(),
  pctOfKCapBp: z.number(),
  currentCapARS: MoneySchema.optional(),
  excluded: z.boolean().optional(),
  excessARS: MoneySchema.nullable().optional(),
  canReturnFrom: z.string().nullable().optional(),
  fee: MoneySchema,
  feeDueDate: z.string(),
  exclusionReason: z.string().optional(),
  exclusionEffectiveDate: z.string().optional(),
  exclusionDate: z.string().optional(),
});
export type MonotributoAssessment = z.infer<typeof MonotributoAssessmentSchema>;

// RI Assessment
export const RIAssessmentSchema = z.object({
  billedInPeriodUSD: MoneySchema,
  billedInPeriodARS: MoneySchema,
  exportSalesARS: MoneySchema,
  localSalesNetARS: MoneySchema,
  outputVat: MoneySchema,
  inputVat: MoneySchema,
  previousVatCredit: MoneySchema,
  vatToPay: MoneySchema,
  newVatCredit: MoneySchema,
  vatDueDate: z.string(),
  selfEmployedContribution: MoneySchema,
  selfEmployedDueDate: z.string(),
  incomeTaxAdvance: MoneySchema.nullable().optional(),
  incomeTaxAdvanceDueDate: z.string().nullable().optional(),
  totalToPay: MoneySchema,
});
export type RIAssessment = z.infer<typeof RIAssessmentSchema>;

// Client DTO
export const RegimeHistoryItemSchema = z.object({
  regime: z.enum(['MONOTRIBUTO', 'RI']).optional(),
  from: z.string(),
  to: z.string().nullable().optional(),
  reason: z.string().optional().default(''),
  note: z.string().optional(),
});
export type RegimeHistoryItem = z.infer<typeof RegimeHistoryItemSchema>;

export const ClientDTOSchema = z.object({
  id: z.string(),
  name: z.string(),
  cuit: z.string(),
  regime: z.enum(['MONOTRIBUTO', 'RI']),
  monotributoCategory: z.string().optional().default(''),
  regimeHistory: z.array(RegimeHistoryItemSchema).optional(),
});
export type ClientDTO = z.infer<typeof ClientDTOSchema>;

// Approval DTO
export const ApprovalSchema = z.object({
  id: z.string(),
  version: z.number(),
  resultHash: z.string(),
  approvedBy: z.string(),
  approvedAt: z.string(),
  note: z.string().optional(),
  batch: z.boolean(),
});
export type ApprovalDTO = z.infer<typeof ApprovalSchema>;

// Published snapshot
export const PublishedFigureItemSchema = z.object({
  concept: z.string(),
  amount: MoneySchema,
  dueDate: z.string(),
});
export type PublishedFigureItem = z.infer<typeof PublishedFigureItemSchema>;

export const PublishedFiguresSchema = z.object({
  billedUSD: MoneySchema,
  billedARS: MoneySchema,
  items: z.array(PublishedFigureItemSchema).default([]),
  total: MoneySchema,
});
export type PublishedFigures = z.infer<typeof PublishedFiguresSchema>;

export const ValidatedBySchema = z.object({
  name: z.string(),
  license: z.string(),
});
export type ValidatedByDTO = z.infer<typeof ValidatedBySchema>;

export const PublishedSchema = z.object({
  version: z.number(),
  title: z.string(),
  body: z.string(),
  figures: PublishedFiguresSchema.optional(),
  publishedAt: z.string(),
  validatedBy: ValidatedBySchema.optional(),
  regime: z.string().optional(),
  category: z.string().nullable().optional(),
  pctOfCategoryCapBp: z.number().nullable().optional(),
  categoryCapARS: MoneySchema.nullable().optional(),
  billed12mARS: MoneySchema.nullable().optional(),
});
export type PublishedDTO = z.infer<typeof PublishedSchema>;

// Client Period Input
export const ExternalIncomeItemSchema = z.object({
  amount: z.union([
    MoneySchema,
    z.object({
      amount: z.string(),
      currency: z.string(),
    }),
  ]),
  description: z.string(),
});
export type ExternalIncomeItem = z.infer<typeof ExternalIncomeItemSchema>;

export const ClientPeriodInputDTOSchema = z.object({
  clientId: z.string().optional(),
  period: z.string().optional(),
  externalIncome: z.array(ExternalIncomeItemSchema).optional(),
  declaredNoExternalIncome: z.boolean().optional(),
  confirmedAt: z.string().nullable().optional(),
  confirmedBy: z.enum(['CLIENT', 'ACCOUNTANT', 'SEED']).nullable().optional(),
  confirmationNote: z.string().nullable().optional(),
  updatedAt: z.string().optional(),
}).passthrough();
export type ClientPeriodInputDTO = z.infer<typeof ClientPeriodInputDTOSchema>;

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
    firstPublishedAt: z.string().nullable().optional(),
    published: PublishedSchema.nullable().optional(),
    approvals: z.array(ApprovalSchema).optional(),
    approvedAt: z.string().nullable().optional(),
    approvedBy: z.string().nullable().optional(),
  }),
  client: ClientDTOSchema,
  periodMovements: z.array(PeriodMovementDTOSchema).default([]),
  periodInvoices: z.array(PeriodInvoiceDTOSchema).default([]),
  reconciliation: ReconciliationDetailSchema,
  clientInput: ClientPeriodInputDTOSchema.optional().default({}),
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

export const ApproveClosingResponseSchema = ClosingDetailDTOSchema.extend({
  warning: z.string().optional(),
});
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

export const RegimeChangeResponseSchema = ClientDTOSchema.extend({
  effectiveDate: z.string().optional(),
  updatedAt: z.string().optional(),
}).passthrough();
export type RegimeChangeResponse = z.infer<typeof RegimeChangeResponseSchema>;

export const LineIssueSchema = z.object({
  line: z.number(),
  message: z.string(),
});
export type LineIssue = z.infer<typeof LineIssueSchema>;

export const WarningIssueSchema = z.union([
  LineIssueSchema,
  z.string().transform((message) => ({ line: 0, message })),
]);

export const ImportResultSchema = z.object({
  imported: z.number(),
  skippedDuplicates: z.number(),
  warnings: z.array(WarningIssueSchema).default([]),
  errors: z.array(LineIssueSchema).default([]),
});
export type ImportResultDTO = z.infer<typeof ImportResultSchema>;

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
