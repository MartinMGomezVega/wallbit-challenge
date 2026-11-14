import { describe, it, expect, beforeAll, afterAll, afterEach } from 'vitest';
import { setupServer } from 'msw/node';
import { handlers, mockCatalog } from '../../../mocks/handlers';
import { listClosings, getClosingsSummary, getSignalsCatalog } from '../api';
import { ClosingRowSchema } from '../schemas';

const server = setupServer(...handlers);

beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('closings-dashboard API & schemas', () => {
  it('getSignalsCatalog obtiene el catálogo de señales validado por Zod', async () => {
    const catalog = await getSignalsCatalog();
    expect(catalog).toHaveLength(mockCatalog.length);
    expect(catalog[0]).toMatchObject({
      code: 'MONO_EXCLUSION_K_CAP_EXCEEDED',
      severity: 'RED',
      priority: 1,
      label: 'Excede tope K (pasa a RI)',
    });
  });

  it('getClosingsSummary obtiene el resumen agrupado por severidad y estado', async () => {
    const summary = await getClosingsSummary('2026-10', {});
    expect(summary.total).toBe(10);
    expect(summary.bySeverity['RED']).toBe(3);
    expect(summary.bySeverity['YELLOW']).toBe(3);
    expect(summary.bySeverity['GREEN']).toBe(4);
    expect(summary.byStatus['READY_FOR_REVIEW']).toBeGreaterThan(0);
  });

  it('listClosings obtiene los 10 registros del mock ordenados y paginados', async () => {
    const res = await listClosings('2026-10', { page: 1, pageSize: 5 });
    expect(res.total).toBe(10);
    expect(res.page).toBe(1);
    expect(res.pageSize).toBe(5);
    expect(res.totalPages).toBe(2);
    expect(res.items).toHaveLength(5);

    // Valida que el primer elemento cumple con ClosingRowSchema
    const parsed = ClosingRowSchema.safeParse(res.items[0]);
    expect(parsed.success).toBe(true);
  });

  it('listClosings aplica filtros correctamente', async () => {
    const res = await listClosings('2026-10', { severity: 'RED' });
    expect(res.total).toBe(3);
    expect(res.items.every((i) => i.severity === 'RED')).toBe(true);
  });

  it('ClosingRowSchema transforma severidad numérica (0, 1, 2) a string', () => {
    const raw = {
      closingId: 'cls_test',
      clientId: 'cli_test',
      clientName: 'Test Client',
      cuit: '20-12345678-9',
      regime: 'MONOTRIBUTO',
      severity: 2,
      description: 'Test description',
      otherSignals: [],
      billedInPeriod: {
        usd: { amount: 1000, currency: 'USD' },
        ars: { amount: 1500000, currency: 'ARS' },
      },
      accumulated12m: {
        ars: { amount: 18000000, currency: 'ARS' },
        pctOfKCapBp: 1500,
      },
      toPay: {
        amount: { amount: 50000, currency: 'ARS' },
        dueDate: '2026-11-20T00:00:00Z',
      },
      status: 'READY_FOR_REVIEW',
      version: 1,
    };

    const parsed = ClosingRowSchema.parse(raw);
    expect(parsed.severity).toBe('RED');
  });
});
