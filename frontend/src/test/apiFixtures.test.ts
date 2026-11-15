import { describe, it, expect } from 'vitest';
import { z } from 'zod';
import {
  ClosingsListResponseSchema,
  ClosingsSummaryResponseSchema,
  SignalCatalogItemSchema,
} from '../features/closings-dashboard/schemas';
import {
  ClosingIDsResponseSchema,
  ClosingDetailDTOSchema,
  ApproveClosingResponseSchema,
  ApproveBatchResponseSchema,
  RegimeChangeResponseSchema,
  ImportResultSchema,
  ClientPeriodInputDTOSchema,
} from '../features/closing-review/schemas';
import { PeriodStatsSchema } from '../features/summary-dashboard/schemas';
import { ParametersDTOSchema } from '../features/parameters/schemas';
import {
  ClientProfileSchema,
  ClientClosingsResponseSchema,
  ClientInvoicesResponseSchema,
  ClientInvoiceSchema,
  MarkPaidResponseSchema,
} from '../features/client-demo/schemas';

import closingsFixture from './fixtures/v1/closings.json';
import closingsSummaryFixture from './fixtures/v1/closings_summary.json';
import closingsIdsFixture from './fixtures/v1/closings_ids.json';
import closingDetailFixture from './fixtures/v1/closing_detail.json';
import closingDetailMonoFixture from './fixtures/v1/closing_detail_mono.json';
import closingDetailMonoExcludedFixture from './fixtures/v1/closing_detail_mono_excluded.json';
import closingDetailRIFixture from './fixtures/v1/closing_detail_ri.json';
import closingDetailNotifiedFixture from './fixtures/v1/closing_detail_notified.json';
import closingDetailPendingInputFixture from './fixtures/v1/closing_detail_pending_input.json';
import closingDetailReopenedFixture from './fixtures/v1/closing_detail_reopened.json';
import statsFixture from './fixtures/v1/stats.json';
import signalsCatalogFixture from './fixtures/v1/signals_catalog.json';
import parametersFixture from './fixtures/v1/parameters.json';
import approveResponseFixture from './fixtures/v1/approve_response.json';
import approveBatchResponseFixture from './fixtures/v1/approve_batch_response.json';
import regimeChangeResponseFixture from './fixtures/v1/regime_change_response.json';
import invoicesImportClientResponseFixture from './fixtures/v1/invoices_import_client_response.json';
import invoicesImportAccountantResponseFixture from './fixtures/v1/invoices_import_accountant_response.json';
import confirmInputResponseFixture from './fixtures/v1/confirm_input_response.json';
import meProfileFixture from './fixtures/v1/me_profile.json';
import meClosingsActionRequiredFixture from './fixtures/v1/me_closings_action_required.json';
import meClosingsInReviewFixture from './fixtures/v1/me_closings_in_review.json';
import meClosingsReadyMonoFixture from './fixtures/v1/me_closings_ready_mono.json';
import meClosingsReadyRIFixture from './fixtures/v1/me_closings_ready_ri.json';
import meInvoicesFixture from './fixtures/v1/me_invoices.json';
import meInvoicesPostResponseFixture from './fixtures/v1/me_invoices_post_response.json';
import meExternalIncomeResponseFixture from './fixtures/v1/me_external_income_response.json';
import meConfirmResponseFixture from './fixtures/v1/me_confirm_response.json';
import mePaidResponseFixture from './fixtures/v1/me_paid_response.json';

describe('API contract fixtures (backend -> Zod)', () => {
  it('valida closings.json con ClosingsListResponseSchema', () => {
    const parsed = ClosingsListResponseSchema.parse(closingsFixture);
    expect(parsed.total).toBe(300);
    expect(parsed.items.length).toBe(50);
  });

  it('valida closings_summary.json con ClosingsSummaryResponseSchema', () => {
    const parsed = ClosingsSummaryResponseSchema.parse(closingsSummaryFixture);
    expect(parsed.total).toBe(300);
  });

  it('valida closings_ids.json con ClosingIDsResponseSchema', () => {
    const parsed = ClosingIDsResponseSchema.parse(closingsIdsFixture);
    expect(parsed.ids.length).toBe(300);
  });

  it('valida closing_detail.json y closing_detail_mono.json con ClosingDetailDTOSchema', () => {
    const parsed = ClosingDetailDTOSchema.parse(closingDetailFixture);
    expect(parsed.closing.id).toBeTruthy();
    const monoParsed = ClosingDetailDTOSchema.parse(closingDetailMonoFixture);
    expect(monoParsed.closing.id).toBeTruthy();
    expect(monoParsed.closing.mono).not.toBeNull();
  });

  it('valida closing_detail_mono_excluded.json (Monotributo con exclusión: exclusionDate presente y excluded true)', () => {
    const parsed = ClosingDetailDTOSchema.parse(closingDetailMonoExcludedFixture);
    expect(parsed.closing.mono).not.toBeNull();
    expect(parsed.closing.mono?.exclusionDate).toBeTruthy();
    expect(parsed.closing.mono?.excluded).toBe(true);
  });

  it('valida closing_detail_ri.json (RI: mono nulo/omitido, ri no null con 15 campos)', () => {
    const parsed = ClosingDetailDTOSchema.parse(closingDetailRIFixture);
    expect(parsed.closing.mono == null).toBe(true);
    expect(parsed.closing.ri).not.toBeNull();
    const ri = parsed.closing.ri!;
    expect(ri.billedInPeriodUSD).toBeDefined();
    expect(ri.billedInPeriodARS).toBeDefined();
    expect(ri.exportSalesARS).toBeDefined();
    expect(ri.localSalesNetARS).toBeDefined();
    expect(ri.outputVat).toBeDefined();
    expect(ri.inputVat).toBeDefined();
    expect(ri.previousVatCredit).toBeDefined();
    expect(ri.vatToPay).toBeDefined();
    expect(ri.newVatCredit).toBeDefined();
    expect(ri.vatDueDate).toBeDefined();
    expect(ri.selfEmployedContribution).toBeDefined();
    expect(ri.selfEmployedDueDate).toBeDefined();
    expect(ri.totalToPay).toBeDefined();
  });

  it('valida closing_detail_notified.json (cierre notificado septiembre: status NOTIFIED, published presente, approvals >= 1)', () => {
    const parsed = ClosingDetailDTOSchema.parse(closingDetailNotifiedFixture);
    expect(parsed.closing.status).toBe('NOTIFIED');
    expect(parsed.closing.published).toBeTruthy();
    expect(parsed.closing.approvals?.length).toBeGreaterThanOrEqual(1);
  });

  it('valida closing_detail_pending_input.json (Paula Rojas: PENDING_CLIENT_INPUT)', () => {
    const parsed = ClosingDetailDTOSchema.parse(closingDetailPendingInputFixture);
    expect(parsed.closing.status).toBe('PENDING_CLIENT_INPUT');
  });

  it('valida closing_detail_reopened.json (cierre reabierto: version 2, READY_FOR_REVIEW, approvals >= 1)', () => {
    const parsed = ClosingDetailDTOSchema.parse(closingDetailReopenedFixture);
    expect(parsed.closing.version).toBe(2);
    expect(parsed.closing.status).toBe('READY_FOR_REVIEW');
    expect(parsed.closing.approvals?.length).toBeGreaterThanOrEqual(1);
  });

  it('valida approve_response.json con ApproveClosingResponseSchema', () => {
    const parsed = ApproveClosingResponseSchema.parse(approveResponseFixture);
    expect(parsed.closing.id).toBeTruthy();
    expect(parsed.closing.status).toBe('NOTIFIED');
  });

  it('valida approve_batch_response.json con ApproveBatchResponseSchema', () => {
    const parsed = ApproveBatchResponseSchema.parse(approveBatchResponseFixture);
    expect(parsed.approved).toBeGreaterThan(0);
    expect(parsed.results.length).toBeGreaterThan(0);
  });

  it('valida regime_change_response.json con RegimeChangeResponseSchema', () => {
    const parsed = RegimeChangeResponseSchema.parse(regimeChangeResponseFixture);
    expect(parsed.id).toBe('cli_demo_02');
    expect(parsed.regime).toBe('RI');
  });

  it('valida invoices_import_client_response.json con ImportResultSchema', () => {
    const parsed = ImportResultSchema.parse(invoicesImportClientResponseFixture);
    expect(parsed.imported).toBe(1);
    expect(parsed.skippedDuplicates).toBe(0);
  });

  it('valida invoices_import_accountant_response.json con ImportResultSchema', () => {
    const parsed = ImportResultSchema.parse(invoicesImportAccountantResponseFixture);
    expect(parsed.imported).toBe(1);
    expect(parsed.skippedDuplicates).toBe(0);
  });

  it('valida confirm_input_response.json con ClientPeriodInputDTOSchema', () => {
    const parsed = ClientPeriodInputDTOSchema.parse(confirmInputResponseFixture);
    expect(parsed.confirmedBy).toBe('ACCOUNTANT');
    expect(parsed.confirmationNote).toBe('Me lo confirmó por mail el 06/11');
    expect(parsed.declaredNoExternalIncome).toBe(true);
  });

  it('valida stats.json con PeriodStatsSchema', () => {
    const parsed = PeriodStatsSchema.parse(statsFixture);
    expect(parsed.progress.total).toBe(300);
  });

  it('valida signals_catalog.json con SignalCatalogItemSchema[]', () => {
    const parsed = z.array(SignalCatalogItemSchema).parse(signalsCatalogFixture);
    expect(parsed.length).toBeGreaterThan(0);
  });

  it('valida parameters.json con ParametersDTOSchema', () => {
    const parsed = ParametersDTOSchema.parse(parametersFixture);
    expect(parsed.fxRule).toBeTruthy();
  });

  it('valida me_profile.json con ClientProfileSchema', () => {
    const parsed = ClientProfileSchema.parse(meProfileFixture);
    expect(parsed.id).toBe('cli_demo_15');
    expect(parsed.name).toBe('Paula Rojas');
    expect(parsed.regime).toBe('MONOTRIBUTO');
  });

  it('valida me_closings_action_required.json con ClientClosingsResponseSchema', () => {
    const parsed = ClientClosingsResponseSchema.parse(meClosingsActionRequiredFixture);
    expect(parsed.length).toBeGreaterThan(0);
    const oct = parsed.find(c => c.period === '2026-10');
    expect(oct).toBeDefined();
    expect(oct?.state).toBe('ACTION_REQUIRED');
  });

  it('valida me_closings_in_review.json con ClientClosingsResponseSchema', () => {
    const parsed = ClientClosingsResponseSchema.parse(meClosingsInReviewFixture);
    expect(parsed.length).toBeGreaterThan(0);
    const oct = parsed.find(c => c.period === '2026-10');
    expect(oct).toBeDefined();
    expect(oct?.state).toBe('IN_REVIEW');
  });

  it('valida me_closings_ready_mono.json (Monotributo: READY, validatedBy, régimen, categoría, tope)', () => {
    const parsed = ClientClosingsResponseSchema.parse(meClosingsReadyMonoFixture);
    const oct = parsed.find(c => c.period === '2026-10');
    expect(oct).toBeDefined();
    expect(oct?.state).toBe('READY');
    expect(oct?.published).toBeDefined();
    expect(oct?.published?.validatedBy?.name).toBe('C.P. Mariana Quiroga');
    expect(oct?.published?.regime).toBe('MONOTRIBUTO');
    expect(oct?.published?.category).toBe('A');
    expect(oct?.published?.categoryCapARS).toBeDefined();
    expect(oct?.published?.pctOfCategoryCapBp).toBeDefined();
  });

  it('valida me_closings_ready_ri.json (RI: READY, validatedBy, régimen RI)', () => {
    const parsed = ClientClosingsResponseSchema.parse(meClosingsReadyRIFixture);
    const oct = parsed.find(c => c.period === '2026-10');
    expect(oct).toBeDefined();
    expect(oct?.state).toBe('READY');
    expect(oct?.published).toBeDefined();
    expect(oct?.published?.validatedBy?.name).toBe('C.P. Mariana Quiroga');
    expect(oct?.published?.regime).toBe('RI');
  });

  it('valida me_invoices.json con ClientInvoicesResponseSchema', () => {
    const parsed = ClientInvoicesResponseSchema.parse(meInvoicesFixture);
    expect(parsed.length).toBeGreaterThan(0);
    expect(parsed[0]?.type).toBe('E');
  });

  it('valida me_invoices_post_response.json con ClientInvoiceSchema', () => {
    const parsed = ClientInvoiceSchema.parse(meInvoicesPostResponseFixture);
    expect(parsed.clientId).toBe('cli_demo_15');
    expect(parsed.number).toBe(99);
    expect(parsed.receiverName).toBe('Global Corp LLC');
  });

  it('valida me_external_income_response.json con ClientPeriodInputDTOSchema', () => {
    const parsed = ClientPeriodInputDTOSchema.parse(meExternalIncomeResponseFixture);
    expect(parsed.clientId).toBe('cli_demo_15');
    expect(parsed.externalIncome?.length).toBe(1);
    expect(parsed.declaredNoExternalIncome).toBe(false);
  });

  it('valida me_confirm_response.json con ClientPeriodInputDTOSchema (confirmedBy CLIENT)', () => {
    const parsed = ClientPeriodInputDTOSchema.parse(meConfirmResponseFixture);
    expect(parsed.clientId).toBe('cli_demo_15');
    expect(parsed.confirmedBy).toBe('CLIENT');
  });

  it('valida me_paid_response.json con MarkPaidResponseSchema', () => {
    const parsed = MarkPaidResponseSchema.parse(mePaidResponseFixture);
    expect(parsed.closingId).toBeTruthy();
    expect(parsed.paidAt).toBeTruthy();
  });

  it('ningún fixture devuelve null en campos que deben ser array u objeto', () => {
    expect(Array.isArray(closingDetailFixture.client.regimeHistory)).toBe(true);
    expect(Array.isArray(closingDetailFixture.clientInput.externalIncome)).toBe(true);
    expect(Array.isArray(closingDetailFixture.periodMovements)).toBe(true);
    expect(Array.isArray(closingDetailFixture.periodInvoices)).toBe(true);
    expect(Array.isArray(closingsFixture.items)).toBe(true);
    expect(closingsSummaryFixture.byStatus).not.toBeNull();
    expect(closingsSummaryFixture.bySignal).not.toBeNull();
    expect(Array.isArray(closingsIdsFixture.ids)).toBe(true);
  });
});
