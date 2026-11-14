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
} from '../features/closing-review/schemas';
import { PeriodStatsSchema } from '../features/summary-dashboard/schemas';
import { ParametersDTOSchema } from '../features/parameters/schemas';

import closingsFixture from './fixtures/api/closings.json';
import closingsSummaryFixture from './fixtures/api/closings_summary.json';
import closingsIdsFixture from './fixtures/api/closings_ids.json';
import closingDetailFixture from './fixtures/api/closing_detail.json';
import statsFixture from './fixtures/api/stats.json';
import signalsCatalogFixture from './fixtures/api/signals_catalog.json';
import parametersFixture from './fixtures/api/parameters.json';

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

  it('valida closing_detail.json con ClosingDetailDTOSchema', () => {
    const parsed = ClosingDetailDTOSchema.parse(closingDetailFixture);
    expect(parsed.closing.id).toBeTruthy();
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
