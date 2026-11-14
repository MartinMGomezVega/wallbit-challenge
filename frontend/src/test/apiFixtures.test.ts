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
} from '../features/closing-review/schemas';
import { PeriodStatsSchema } from '../features/summary-dashboard/schemas';
import { ParametersDTOSchema } from '../features/parameters/schemas';

import closingsFixture from './fixtures/api/closings.json';
import closingsSummaryFixture from './fixtures/api/closings_summary.json';
import closingsIdsFixture from './fixtures/api/closings_ids.json';
import closingDetailFixture from './fixtures/api/closing_detail.json';
import closingDetailMonoFixture from './fixtures/api/closing_detail_mono.json';
import closingDetailMonoExcludedFixture from './fixtures/api/closing_detail_mono_excluded.json';
import closingDetailRIFixture from './fixtures/api/closing_detail_ri.json';
import closingDetailNotifiedFixture from './fixtures/api/closing_detail_notified.json';
import closingDetailPendingInputFixture from './fixtures/api/closing_detail_pending_input.json';
import closingDetailReopenedFixture from './fixtures/api/closing_detail_reopened.json';
import statsFixture from './fixtures/api/stats.json';
import signalsCatalogFixture from './fixtures/api/signals_catalog.json';
import parametersFixture from './fixtures/api/parameters.json';
import approveResponseFixture from './fixtures/api/approve_response.json';
import approveBatchResponseFixture from './fixtures/api/approve_batch_response.json';
import regimeChangeResponseFixture from './fixtures/api/regime_change_response.json';
import invoicesImportClientResponseFixture from './fixtures/api/invoices_import_client_response.json';

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
