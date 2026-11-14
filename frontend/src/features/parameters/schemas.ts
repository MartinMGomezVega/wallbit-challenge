import { z } from 'zod';

export const MoneyDTOSchema = z.object({
  amount: z.string(),
  currency: z.string(),
});

export const ParametersDTOSchema = z.object({
  invoiceECountsForCap: z.boolean(),
  fxRule: z.enum([
    'BNA_BUY_PREV_BUSINESS_DAY',
    'BNA_SELL_PREV_BUSINESS_DAY',
    'BNA_BUY_SAME_DAY',
    'BNA_SELL_SAME_DAY',
  ]),
  categoryCapWarningBp: z.number(),
  kCapWarningBp: z.number(),
  atypicalMultiplierBp: z.number(),
  atypicalWindowMonths: z.number(),
  missingInvoiceMinUsd: MoneyDTOSchema,
  reconciliationToleranceBp: z.number(),
  reconciliationWindowBeforeDays: z.number(),
  reconciliationWindowAfterDays: z.number(),
  cardSpendKCapBp: z.number(),
  reminderDaysBeforeDue: z.number(),
  updatedAt: z.string().optional(),
  updatedBy: z.string().optional(),
});

export type ParametersDTO = z.infer<typeof ParametersDTOSchema>;

export const PutParametersResponseSchema = z.object({
  parameters: ParametersDTOSchema,
  recalculated: z.number(),
});

export type PutParametersResponse = z.infer<typeof PutParametersResponseSchema>;

