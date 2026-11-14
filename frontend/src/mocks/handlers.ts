import { http, HttpResponse } from 'msw';
import { SignalCatalogItem } from '../features/closings-dashboard/schemas';

export interface MockClosingRow {
  closingId: string;
  clientId: string;
  clientName: string;
  cuit: string;
  regime: 'MONOTRIBUTO' | 'RI';
  category?: string | null;
  severity: 'RED' | 'YELLOW' | 'GREEN';
  primarySignal?: { code: string; label: string; severity?: 'RED' | 'YELLOW' | 'GREEN' } | null;
  description: string;
  otherSignals: Array<{ code: string; label: string; severity?: 'RED' | 'YELLOW' | 'GREEN' }>;
  billedInPeriod: {
    usd: { amount: number; currency: 'USD' };
    ars: { amount: number; currency: 'ARS' };
  };
  accumulated12m: {
    ars: { amount: number; currency: 'ARS' };
    pctOfKCapBp: number;
  };
  toPay: {
    amount: { amount: number; currency: 'ARS' };
    dueDate: string;
  };
  status: 'PENDING_CLIENT_INPUT' | 'READY_FOR_REVIEW' | 'APPROVED' | 'NOTIFIED';
  version: number;
  paidAt?: string | null;
}

export const mockCatalog: SignalCatalogItem[] = [
  { code: 'MONO_EXCLUSION_K_CAP_EXCEEDED', severity: 'RED', priority: 1, label: 'Excede tope K (pasa a RI)' },
  { code: 'POST_EXCLUSION_C_INVOICES', severity: 'RED', priority: 2, label: 'Facturas C post-exclusión' },
  { code: 'MISSING_INVOICES', severity: 'RED', priority: 3, label: 'Faltan facturas' },
  { code: 'DEPOSITS_EXCEED_K_CAP', severity: 'RED', priority: 4, label: 'Depósitos sobre tope K' },
  { code: 'MONO_NEAR_K_CAP', severity: 'YELLOW', priority: 5, label: 'Cerca del tope K' },
  { code: 'MONO_CATEGORY_CHANGE_REQUIRED', severity: 'YELLOW', priority: 6, label: 'Debe recategorizar' },
  { code: 'MONO_NEAR_CATEGORY_CAP', severity: 'YELLOW', priority: 7, label: 'Cerca del tope de su categoría' },
  { code: 'ATYPICAL_INCOME', severity: 'YELLOW', priority: 8, label: 'Cobro fuera de lo habitual' },
  { code: 'UNCLASSIFIED_MOVEMENTS', severity: 'YELLOW', priority: 9, label: 'Movimientos sin clasificar' },
  { code: 'EXTERNAL_INCOME_DECLARED', severity: 'YELLOW', priority: 10, label: 'Cobros fuera de Wallbit' },
  { code: 'INVOICE_WITHOUT_PAYMENT', severity: 'YELLOW', priority: 11, label: 'Facturas sin cobro' },
  { code: 'HIGH_CARD_SPENDING', severity: 'YELLOW', priority: 12, label: 'Gastos con tarjeta altos' },
  { code: 'MONO_CATEGORY_DOWNGRADE_POSSIBLE', severity: 'YELLOW', priority: 13, label: 'Puede bajar de categoría' },
];

export const mockParameters = {
  invoiceECountsForCap: true,
  fxRule: 'BNA_BUY_PREV_BUSINESS_DAY' as const,
  categoryCapWarningBp: 8000,
  kCapWarningBp: 8000,
  atypicalMultiplierBp: 25000,
  atypicalWindowMonths: 6,
  missingInvoiceMinUsd: { amount: '50.00', currency: 'USD' },
  reconciliationToleranceBp: 300,
  reconciliationWindowBeforeDays: 45,
  reconciliationWindowAfterDays: 10,
  cardSpendKCapBp: 4000,
  reminderDaysBeforeDue: 3,
  updatedAt: '2026-10-01T12:00:00Z',
  updatedBy: 'system',
};

export const mockClosings: MockClosingRow[] = [
  {
    closingId: 'cls_01',
    clientId: 'cli_01',
    clientName: 'Ignacio Sosa',
    cuit: '20-28222221-1',
    regime: 'MONOTRIBUTO',
    category: 'K',
    severity: 'RED',
    primarySignal: { code: 'MONO_EXCLUSION_K_CAP_EXCEEDED', label: 'Excede tope K (pasa a RI)', severity: 'RED' },
    description: 'Superó el tope de la categoría K el 14/10/2026. Corresponde pase a Responsable Inscripto.',
    otherSignals: [],
    billedInPeriod: { usd: { amount: 750000, currency: 'USD' }, ars: { amount: 1125000000, currency: 'ARS' } },
    accumulated12m: { ars: { amount: 12840000000, currency: 'ARS' }, pctOfKCapBp: 10140 },
    toPay: { amount: { amount: 161444604, currency: 'ARS' }, dueDate: '2026-11-20T00:00:00Z' },
    status: 'READY_FOR_REVIEW',
    version: 1,
    paidAt: null,
  },
  {
    closingId: 'cls_02',
    clientId: 'cli_02',
    clientName: 'Carla Benítez',
    cuit: '27-34333332-2',
    regime: 'MONOTRIBUTO',
    category: 'E',
    severity: 'RED',
    primarySignal: { code: 'MISSING_INVOICES', label: 'Faltan facturas', severity: 'RED' },
    description: '2 cobros por US$ 3.000,00 sin factura asociada.',
    otherSignals: [],
    billedInPeriod: { usd: { amount: 0, currency: 'USD' }, ars: { amount: 0, currency: 'ARS' } },
    accumulated12m: { ars: { amount: 3300000000, currency: 'ARS' }, pctOfKCapBp: 2606 },
    toPay: { amount: { amount: 11981100, currency: 'ARS' }, dueDate: '2026-11-20T00:00:00Z' },
    status: 'READY_FOR_REVIEW',
    version: 1,
    paidAt: null,
  },
  {
    closingId: 'cls_03',
    clientId: 'cli_03',
    clientName: 'Valentina Ruiz',
    cuit: '27-36555554-4',
    regime: 'MONOTRIBUTO',
    category: 'F',
    severity: 'YELLOW',
    primarySignal: { code: 'MONO_NEAR_CATEGORY_CAP', label: 'Cerca del tope de su categoría', severity: 'YELLOW' },
    description: 'Acumulado al 91% del tope de la categoría F.',
    otherSignals: [],
    billedInPeriod: { usd: { amount: 228000, currency: 'USD' }, ars: { amount: 342000000, currency: 'ARS' } },
    accumulated12m: { ars: { amount: 4104000000, currency: 'ARS' }, pctOfKCapBp: 3241 },
    toPay: { amount: { amount: 15078400, currency: 'ARS' }, dueDate: '2026-11-20T00:00:00Z' },
    status: 'READY_FOR_REVIEW',
    version: 1,
    paidAt: null,
  },
  {
    closingId: 'cls_04',
    clientId: 'cli_04',
    clientName: 'Tomás Álvarez',
    cuit: '20-32666665-5',
    regime: 'MONOTRIBUTO',
    category: 'D',
    severity: 'YELLOW',
    primarySignal: { code: 'ATYPICAL_INCOME', label: 'Cobro fuera de lo habitual', severity: 'YELLOW' },
    description: 'Cobro de US$ 3.500,00, 3,7× su mediana de 6 meses.',
    otherSignals: [
      { code: 'MONO_CATEGORY_DOWNGRADE_POSSIBLE', label: 'Puede bajar de categoría', severity: 'YELLOW' },
    ],
    billedInPeriod: { usd: { amount: 350000, currency: 'USD' }, ars: { amount: 525000000, currency: 'ARS' } },
    accumulated12m: { ars: { amount: 2092500000, currency: 'ARS' }, pctOfKCapBp: 1653 },
    toPay: { amount: { amount: 8461300, currency: 'ARS' }, dueDate: '2026-11-20T00:00:00Z' },
    status: 'READY_FOR_REVIEW',
    version: 1,
    paidAt: null,
  },
  {
    closingId: 'cls_05',
    clientId: 'cli_05',
    clientName: 'Federico Díaz',
    cuit: '20-33445566-7',
    regime: 'RI',
    category: null,
    severity: 'YELLOW',
    primarySignal: { code: 'INVOICE_WITHOUT_PAYMENT', label: 'Facturas sin cobro', severity: 'YELLOW' },
    description: '1 factura por US$ 4.000,00 sin cobro en Wallbit.',
    otherSignals: [],
    billedInPeriod: { usd: { amount: 400000, currency: 'USD' }, ars: { amount: 600000000, currency: 'ARS' } },
    accumulated12m: { ars: { amount: 5400000000, currency: 'ARS' }, pctOfKCapBp: 4265 },
    toPay: { amount: { amount: 105000000, currency: 'ARS' }, dueDate: '2026-11-18T00:00:00Z' },
    status: 'READY_FOR_REVIEW',
    version: 1,
    paidAt: null,
  },
  {
    closingId: 'cls_06',
    clientId: 'cli_06',
    clientName: 'Lucía Fernández',
    cuit: '27-32111110-0',
    regime: 'MONOTRIBUTO',
    category: 'A',
    severity: 'GREEN',
    primarySignal: null,
    description: 'Facturó US$ 500,00 · cuota $ 49.527,18 · todo conciliado.',
    otherSignals: [],
    billedInPeriod: { usd: { amount: 50000, currency: 'USD' }, ars: { amount: 75000000, currency: 'ARS' } },
    accumulated12m: { ars: { amount: 900000000, currency: 'ARS' }, pctOfKCapBp: 710 },
    toPay: { amount: { amount: 4952718, currency: 'ARS' }, dueDate: '2026-11-20T00:00:00Z' },
    status: 'APPROVED',
    version: 1,
    paidAt: null,
  },
  {
    closingId: 'cls_07',
    clientId: 'cli_07',
    clientName: 'Martín Gómez',
    cuit: '20-31222333-4',
    regime: 'MONOTRIBUTO',
    category: 'B',
    severity: 'GREEN',
    primarySignal: null,
    description: 'Facturó US$ 720,00 · cuota $ 56.379,08 · todo conciliado.',
    otherSignals: [],
    billedInPeriod: { usd: { amount: 72000, currency: 'USD' }, ars: { amount: 108000000, currency: 'ARS' } },
    accumulated12m: { ars: { amount: 1296000000, currency: 'ARS' }, pctOfKCapBp: 1023 },
    toPay: { amount: { amount: 5637908, currency: 'ARS' }, dueDate: '2026-11-20T00:00:00Z' },
    status: 'NOTIFIED',
    version: 1,
    paidAt: '2026-11-10T15:00:00Z',
  },
  {
    closingId: 'cls_08',
    clientId: 'cli_08',
    clientName: 'Javier Domínguez',
    cuit: '20-25666777-8',
    regime: 'MONOTRIBUTO',
    category: 'H',
    severity: 'GREEN',
    primarySignal: null,
    description: 'Facturó US$ 3.300,00 · cuota $ 522.706,68 · todo conciliado.',
    otherSignals: [],
    billedInPeriod: { usd: { amount: 330000, currency: 'USD' }, ars: { amount: 495000000, currency: 'ARS' } },
    accumulated12m: { ars: { amount: 5940000000, currency: 'ARS' }, pctOfKCapBp: 4691 },
    toPay: { amount: { amount: 52270668, currency: 'ARS' }, dueDate: '2026-11-20T00:00:00Z' },
    status: 'READY_FOR_REVIEW',
    version: 1,
    paidAt: null,
  },
  {
    closingId: 'cls_09',
    clientId: 'cli_09',
    clientName: 'Camila Torres',
    cuit: '27-39999998-8',
    regime: 'RI',
    category: null,
    severity: 'GREEN',
    primarySignal: null,
    description: 'Facturó US$ 3.500,00 · IVA $ 0,00 · todo conciliado.',
    otherSignals: [],
    billedInPeriod: { usd: { amount: 350000, currency: 'USD' }, ars: { amount: 525000000, currency: 'ARS' } },
    accumulated12m: { ars: { amount: 6300000000, currency: 'ARS' }, pctOfKCapBp: 4975 },
    toPay: { amount: { amount: 4500000, currency: 'ARS' }, dueDate: '2026-11-18T00:00:00Z' },
    status: 'READY_FOR_REVIEW',
    version: 1,
    paidAt: null,
  },
  {
    closingId: 'cls_10',
    clientId: 'cli_10',
    clientName: 'Paula Rojas',
    cuit: '27-33444444-4',
    regime: 'MONOTRIBUTO',
    category: 'B',
    severity: 'RED',
    primarySignal: { code: 'MISSING_INVOICES', label: 'Faltan facturas', severity: 'RED' },
    description: '1 cobro por US$ 850,00 sin factura asociada.',
    otherSignals: [],
    billedInPeriod: { usd: { amount: 0, currency: 'USD' }, ars: { amount: 0, currency: 'ARS' } },
    accumulated12m: { ars: { amount: 1530000000, currency: 'ARS' }, pctOfKCapBp: 1208 },
    toPay: { amount: { amount: 5637908, currency: 'ARS' }, dueDate: '2026-11-20T00:00:00Z' },
    status: 'PENDING_CLIENT_INPUT',
    version: 1,
    paidAt: null,
  },
];

export const handlers = [
  http.get('/api/v1/health', () =>
    HttpResponse.json({ status: 'ok', now: '2026-11-05T09:00:00-03:00' })
  ),

  http.get('/api/v1/accountant/periods', () =>
    HttpResponse.json({ periods: ['2026-10', '2026-09'] })
  ),

  http.get('/api/v1/accountant/signals/catalog', () =>
    HttpResponse.json(mockCatalog)
  ),

  http.get('/api/v1/accountant/periods/:period/closings/summary', ({ request }) => {
    const url = new URL(request.url);
    const regime = url.searchParams.get('regime');
    const q = url.searchParams.get('q')?.toLowerCase();

    let filtered = [...mockClosings];
    if (regime) {
      filtered = filtered.filter((r) => r.regime === regime);
    }
    if (q) {
      filtered = filtered.filter(
        (r) =>
          r.clientName.toLowerCase().includes(q) ||
          r.cuit.replace(/[-_]/g, '').includes(q.replace(/[-_]/g, ''))
      );
    }

    const bySeverity: Record<string, number> = { RED: 0, YELLOW: 0, GREEN: 0 };
    const bySignal: Record<string, number> = {};
    const byStatus: Record<string, number> = {};

    for (const r of filtered) {
      bySeverity[r.severity] = (bySeverity[r.severity] || 0) + 1;
      byStatus[r.status] = (byStatus[r.status] || 0) + 1;
      if (r.primarySignal) {
        bySignal[r.primarySignal.code] = (bySignal[r.primarySignal.code] || 0) + 1;
      }
      for (const os of r.otherSignals) {
        bySignal[os.code] = (bySignal[os.code] || 0) + 1;
      }
    }

    return HttpResponse.json({
      total: filtered.length,
      bySeverity,
      bySignal,
      byStatus,
    });
  }),

  http.get('/api/v1/accountant/periods/:period/closings', ({ request }) => {
    const url = new URL(request.url);
    const severityParam = url.searchParams.get('severity');
    const signalParam = url.searchParams.get('signal');
    const regimeParam = url.searchParams.get('regime');
    const statusParam = url.searchParams.get('status');
    const q = url.searchParams.get('q')?.toLowerCase();
    const sort = url.searchParams.get('sort') || 'severity';
    const dir = url.searchParams.get('dir') || 'asc';
    const page = parseInt(url.searchParams.get('page') || '1', 10);
    const pageSize = parseInt(url.searchParams.get('pageSize') || '50', 10);

    let rows = [...mockClosings];

    if (severityParam) {
      const sevs = severityParam.split(',');
      rows = rows.filter((r) => sevs.includes(r.severity));
    }

    if (signalParam) {
      const sigs = signalParam.split(',');
      rows = rows.filter((r) => {
        if (r.primarySignal && sigs.includes(r.primarySignal.code)) return true;
        return r.otherSignals.some((os) => sigs.includes(os.code));
      });
    }

    if (regimeParam) {
      rows = rows.filter((r) => r.regime === regimeParam);
    }

    if (statusParam) {
      const statuses = statusParam.split(',');
      rows = rows.filter((r) => statuses.includes(r.status));
    }

    if (q) {
      const normQ = q.replace(/[-_]/g, '');
      rows = rows.filter(
        (r) =>
          r.clientName.toLowerCase().includes(q) ||
          r.cuit.replace(/[-_]/g, '').includes(normQ)
      );
    }

    // Sort
    const sevRank = { RED: 3, YELLOW: 2, GREEN: 1 };
    const statusRank = {
      PENDING_CLIENT_INPUT: 1,
      READY_FOR_REVIEW: 2,
      APPROVED: 3,
      NOTIFIED: 4,
    };
    rows.sort((a, b) => {
      let cmp = 0;
      if (sort === 'severity') {
        cmp = sevRank[b.severity] - sevRank[a.severity];
        if (cmp === 0) {
          cmp = a.clientName.localeCompare(b.clientName);
        }
      } else if (sort === 'name') {
        cmp = a.clientName.localeCompare(b.clientName);
      } else if (sort === 'regime') {
        cmp = a.regime.localeCompare(b.regime) || (a.category || '').localeCompare(b.category || '');
      } else if (sort === 'status') {
        cmp = (statusRank[a.status] || 0) - (statusRank[b.status] || 0);
      } else if (sort === 'billed') {
        cmp = a.billedInPeriod.usd.amount - b.billedInPeriod.usd.amount;
      } else if (sort === 'toPay') {
        cmp = a.toPay.amount.amount - b.toPay.amount.amount;
      } else if (sort === 'pctKCap') {
        cmp = a.accumulated12m.pctOfKCapBp - b.accumulated12m.pctOfKCapBp;
      }

      if (dir === 'desc' && sort !== 'severity') {
        cmp = -cmp;
      }
      return cmp || a.clientId.localeCompare(b.clientId);
    });

    const total = rows.length;
    const totalPages = Math.max(1, Math.ceil(total / pageSize));
    const start = (page - 1) * pageSize;
    const items = rows.slice(start, start + pageSize);

    return HttpResponse.json({
      items,
      total,
      page,
      pageSize,
      totalPages,
    });
  }),

  // GET /api/v1/accountant/parameters
  http.get('/api/v1/accountant/parameters', () => {
    return HttpResponse.json(mockParameters);
  }),

  // PUT /api/v1/accountant/parameters
  http.put('/api/v1/accountant/parameters', async ({ request }) => {
    const body = (await request.json()) as typeof mockParameters;
    if (body.atypicalWindowMonths <= 0) {
      return HttpResponse.json(
        {
          error: {
            code: 'VALIDATION_ERROR',
            message: 'Parámetros inválidos',
            fields: { atypicalWindowMonths: 'debe ser mayor a 0' },
          },
        },
        { status: 422 }
      );
    }
    const updated = {
      ...mockParameters,
      ...body,
      updatedAt: new Date().toISOString(),
      updatedBy: 'contador@estudio.com',
    };
    return HttpResponse.json({
      parameters: updated,
      recalculated: 42,
    });
  }),

  // GET /api/v1/accountant/periods/:period/closings/ids
  http.get('/api/v1/accountant/periods/:period/closings/ids', ({ request }) => {
    const url = new URL(request.url);
    const severityParam = url.searchParams.get('severity');
    const signalParam = url.searchParams.get('signal');
    const regimeParam = url.searchParams.get('regime');
    const statusParam = url.searchParams.get('status');
    const q = url.searchParams.get('q')?.toLowerCase();
    const sort = url.searchParams.get('sort') || 'severity';
    const dir = url.searchParams.get('dir') || 'asc';

    let rows = [...mockClosings];

    if (severityParam) {
      const sevs = severityParam.split(',');
      rows = rows.filter((r) => sevs.includes(r.severity));
    }

    if (signalParam) {
      const sigs = signalParam.split(',');
      rows = rows.filter((r) => {
        if (r.primarySignal && sigs.includes(r.primarySignal.code)) return true;
        return r.otherSignals.some((os) => sigs.includes(os.code));
      });
    }

    if (regimeParam) {
      rows = rows.filter((r) => r.regime === regimeParam);
    }

    if (statusParam) {
      const statuses = statusParam.split(',');
      rows = rows.filter((r) => statuses.includes(r.status));
    }

    if (q) {
      const normQ = q.replace(/[-_]/g, '');
      rows = rows.filter(
        (r) =>
          r.clientName.toLowerCase().includes(q) ||
          r.cuit.replace(/[-_]/g, '').includes(normQ)
      );
    }

    const sevRank = { RED: 3, YELLOW: 2, GREEN: 1 };
    const statusRank = {
      PENDING_CLIENT_INPUT: 1,
      READY_FOR_REVIEW: 2,
      APPROVED: 3,
      NOTIFIED: 4,
    };
    rows.sort((a, b) => {
      let cmp = 0;
      if (sort === 'severity') {
        cmp = sevRank[b.severity] - sevRank[a.severity];
        if (cmp === 0) {
          cmp = a.clientName.localeCompare(b.clientName);
        }
      } else if (sort === 'name') {
        cmp = a.clientName.localeCompare(b.clientName);
      } else if (sort === 'regime') {
        cmp = a.regime.localeCompare(b.regime) || (a.category || '').localeCompare(b.category || '');
      } else if (sort === 'status') {
        cmp = (statusRank[a.status] || 0) - (statusRank[b.status] || 0);
      } else if (sort === 'billed') {
        cmp = a.billedInPeriod.usd.amount - b.billedInPeriod.usd.amount;
      } else if (sort === 'toPay') {
        cmp = a.toPay.amount.amount - b.toPay.amount.amount;
      } else if (sort === 'pctKCap') {
        cmp = a.accumulated12m.pctOfKCapBp - b.accumulated12m.pctOfKCapBp;
      }

      if (dir === 'desc' && sort !== 'severity') {
        cmp = -cmp;
      }
      return cmp || a.clientId.localeCompare(b.clientId);
    });

    return HttpResponse.json({
      ids: rows.map((r) => r.closingId),
      total: rows.length,
    });
  }),

  // GET /api/v1/accountant/closings/:id
  http.get('/api/v1/accountant/closings/:id', ({ params }) => {
    const id = params.id as string;
    const defaultRow = mockClosings[0]!;
    const row = mockClosings.find((c) => c.closingId === id) || defaultRow;

    const isMono = row.regime === 'MONOTRIBUTO';
    const isExclusion = row.primarySignal?.code === 'MONO_EXCLUSION_K_CAP_EXCEEDED';

    return HttpResponse.json({
      closing: {
        id: row.closingId,
        clientId: row.clientId,
        period: '2026-10',
        status: row.status,
        version: row.version,
        resultHash: 'e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855',
        severity: row.severity,
        description: row.description,
        signals: [
          ...(row.primarySignal
            ? [
                {
                  code: row.primarySignal.code,
                  severity: row.primarySignal.severity || row.severity,
                  message: row.description,
                  data: isExclusion
                    ? {
                        exclusionDate: '2026-10-14',
                        excessARS: '25000000',
                      }
                    : {},
                },
              ]
            : []),
          ...row.otherSignals.map((s) => ({
            code: s.code,
            severity: s.severity || 'YELLOW',
            message: s.label,
          })),
        ],
        mono: isMono
          ? {
              billedInPeriodUSD: row.billedInPeriod.usd,
              billedInPeriodARS: row.billedInPeriod.ars,
              currentCategory: row.category || 'K',
              requiredCategory: isExclusion ? undefined : 'K',
              billed12mARS: row.accumulated12m.ars,
              kCapARS: { amount: 12660000000, currency: 'ARS' },
              pctOfKCapBp: row.accumulated12m.pctOfKCapBp,
              currentCapARS: { amount: 12660000000, currency: 'ARS' },
              fee: row.toPay.amount,
              feeDueDate: row.toPay.dueDate,
              exclusionReason: isExclusion ? 'K_CAP_EXCEEDED' : undefined,
              exclusionEffectiveDate: isExclusion ? '2026-10-14' : undefined,
              exclusionDate: isExclusion ? '2026-10-14T00:00:00Z' : undefined,
            }
          : null,
        ri: !isMono
          ? {
              billedInPeriodUSD: row.billedInPeriod.usd,
              billedInPeriodARS: row.billedInPeriod.ars,
              debitoFiscalARS: { amount: 4200000, currency: 'ARS' },
              creditoFiscalARS: { amount: 800000, currency: 'ARS' },
              saldoTecnicoARS: { amount: 3400000, currency: 'ARS' },
              retencionesPercepcionesARS: { amount: 200000, currency: 'ARS' },
              ivaToPay: { amount: 3200000, currency: 'ARS' },
              vatDueDate: '2026-11-18T00:00:00Z',
              autonomousCategory: 'Cat. III',
              autonomousFee: { amount: 450000, currency: 'ARS' },
              selfEmployedDueDate: '2026-11-05T00:00:00Z',
              incomeTaxAdvance: { amount: 1200000, currency: 'ARS' },
              incomeTaxAdvanceDueDate: '2026-11-15T00:00:00Z',
              totalToPay: row.toPay.amount,
            }
          : null,
        computedAt: '2026-11-01T03:00:00Z',
        paidAt: row.paidAt,
        approvedAt: row.status === 'APPROVED' ? '2026-11-02T10:00:00Z' : null,
        approvedBy: row.status === 'APPROVED' ? 'contador@estudio.com' : null,
      },
      client: {
        id: row.clientId,
        name: row.clientName,
        cuit: row.cuit,
        regime: row.regime,
        monotributoCategory: row.category || '',
      },
      periodMovements: [
        {
          movement: {
            id: 'mov_01',
            clientId: row.clientId,
            type: 'INCOMING_WIRE',
            amount: { amount: 150000, currency: 'USD' },
            counterparty: 'Acme Corp',
            description: 'Monthly consulting retainer',
            occurredAt: '2026-10-15T14:30:00Z',
          },
          classification: {
            nature: 'INCOME',
            rule: 'DEFAULT_INCOME',
            confidence: 0.98,
            comment: 'Ingreso habitual',
            overridden: false,
          },
        },
        {
          movement: {
            id: 'mov_02',
            clientId: row.clientId,
            type: 'CARD_PURCHASE',
            amount: { amount: 25000, currency: 'USD' },
            counterparty: 'AWS EMEA',
            description: 'Cloud hosting services',
            occurredAt: '2026-10-20T10:00:00Z',
          },
          classification: {
            nature: 'OTHER',
            rule: 'EXPENSE_RULE',
            confidence: 0.95,
            comment: 'Gasto operativo',
            overridden: false,
          },
        },
      ],
      periodInvoices: [
        {
          invoice: {
            id: 'inv_01',
            clientId: row.clientId,
            pointOfSale: 1,
            number: 104,
            type: 'Factura E',
            concept: 'Servicios de software',
            issueDate: '2026-10-15T15:00:00Z',
            totalAmount: { amount: 150000, currency: 'USD' },
            currency: 'USD',
            cae: '74125896325874',
            caeDueDate: '2026-10-25T00:00:00Z',
          },
          amountArs: { amount: 225000000, currency: 'ARS' },
          fxRateUsed: 1500.0,
        },
      ],
      reconciliation: {
        matchedPairs: [
          {
            movement: {
              id: 'mov_01',
              clientId: row.clientId,
              type: 'INCOMING_WIRE',
              amount: { amount: 150000, currency: 'USD' },
              counterparty: 'Acme Corp',
              description: 'Monthly consulting retainer',
              occurredAt: '2026-10-15T14:30:00Z',
            },
            invoice: {
              id: 'inv_01',
              clientId: row.clientId,
              pointOfSale: 1,
              number: 104,
              type: 'Factura E',
              concept: 'Servicios de software',
              issueDate: '2026-10-15T15:00:00Z',
              totalAmount: { amount: 150000, currency: 'USD' },
              currency: 'USD',
              cae: '74125896325874',
              caeDueDate: '2026-10-25T00:00:00Z',
            },
            diffAmount: 0,
          },
        ],
        unmatchedIncome: [],
        unmatchedInvoices: [],
      },
      clientInput: {},
      draftMessage: {
        title: 'Tu cierre de octubre 2026 está listo',
        body: isMono
          ? `Facturaste US$ 1.500,00 ($ 2.250.000,00). Cuota de Monotributo categoría ${row.category || 'Cat. C'}: $ 30.000,00, vence el 20/11/2026.`
          : `Facturaste US$ 1.500,00 ($ 2.250.000,00). IVA a pagar: $ 32.000,00 (vence 18/11/2026). Aportes de autónomos: $ 4.500,00 (vence 05/11/2026). Total: $ 36.500,00.`,
      },
    });
  }),

  // PATCH /api/v1/accountant/movements/:id/classification
  http.patch('/api/v1/accountant/movements/:id/classification', async ({ params, request }) => {
    const id = params.id as string;
    const body = (await request.json()) as { nature: 'INCOME' | 'OTHER'; comment?: string };

    return HttpResponse.json({
      id: `cls_${id}`,
      movementId: id,
      nature: body.nature,
      rule: 'MANUAL_OVERRIDE',
      comment: body.comment || 'Modificado por el contador',
      overridden: true,
      updatedAt: new Date().toISOString(),
      updatedBy: 'contador@estudio.com',
    });
  }),

  // POST /api/v1/accountant/clients/:id/regime-change
  http.post('/api/v1/accountant/clients/:id/regime-change', async ({ params, request }) => {
    const id = params.id as string;
    const body = (await request.json()) as { to: string; effectiveDate: string };

    return HttpResponse.json({
      id,
      name: 'Ignacio Sosa',
      cuit: '20-28222221-1',
      regime: body.to,
      effectiveDate: body.effectiveDate,
      updatedAt: new Date().toISOString(),
    });
  }),

  // POST /api/v1/accountant/closings/:id/approve
  http.post('/api/v1/accountant/closings/:id/approve', async ({ params, request }) => {
    const id = params.id as string;
    const body = (await request.json()) as { version: number; clientNote?: string };

    if (body.version === 99) {
      return HttpResponse.json(
        { error: { code: 'VERSION_CONFLICT', message: 'El cierre cambió mientras lo revisabas' } },
        { status: 409 }
      );
    }

    return HttpResponse.json({
      id,
      version: body.version,
      status: 'NOTIFIED',
      warning: id === 'clo_warning' ? 'CLIENT_INPUT_NOT_CONFIRMED' : undefined,
      approvals: [
        {
          version: body.version,
          approvedBy: 'contador@estudio.com',
          approvedAt: new Date().toISOString(),
          note: body.clientNote || '',
          batch: false,
        },
      ],
    });
  }),

  // POST /api/v1/accountant/periods/:period/closings/approve-batch
  http.post('/api/v1/accountant/periods/:period/closings/approve-batch', async ({ request }) => {
    const body = (await request.json()) as { items: Array<{ id: string; version: number }> };
    const items = body.items || [];

    const results = items.map((it) => {
      if (it.id === 'clo_fail_conflict') {
        return { id: it.id, ok: false, error: 'VERSION_CONFLICT' };
      }
      if (it.id === 'clo_fail_not_ready') {
        return { id: it.id, ok: false, error: 'NOT_READY' };
      }
      return { id: it.id, ok: true };
    });

    const approved = results.filter((r) => r.ok).length;
    const failed = results.filter((r) => !r.ok).length;

    return HttpResponse.json({
      approved,
      failed,
      results,
    });
  }),

  // GET /api/v1/accountant/periods/:period/stats
  http.get('/api/v1/accountant/periods/:period/stats', ({ params }) => {
    const period = String(params.period ?? '2026-10');
    return HttpResponse.json({
      period,
      progress: {
        total: 300,
        byStatus: {
          PENDING_CLIENT_INPUT: 30,
          READY_FOR_REVIEW: 190,
          APPROVED: 20,
          NOTIFIED: 60,
        },
        approvedPct: 2666,
        pendingReview: { RED: 21, YELLOW: 56, GREEN: 143 },
        estimatedReviewMinutes: 1056,
      },
      totals: {
        billedUSD: { amount: 10000000, currency: 'USD' },
        billedARS: { amount: 15000000000, currency: 'ARS' },
        toPayARS: { amount: 500000000, currency: 'ARS' },
        toPayByDueDate: [
          {
            date: '2026-11-20',
            amount: { amount: 500000000, currency: 'ARS' },
            clients: 255,
          },
        ],
      },
      risk: {
        bySeverity: { RED: 21, YELLOW: 56, GREEN: 223 },
        bySignal: [
          {
            code: 'MISSING_INVOICES',
            label: 'Faltan facturas',
            severity: 'RED',
            count: 14,
          },
          {
            code: 'NEAR_K_CAP',
            label: 'Cerca del tope K',
            severity: 'YELLOW',
            count: 9,
          },
        ],
        byCategory: [
          { category: 'A', clients: 12 },
          { category: 'K', clients: 9 },
          { category: 'RI', clients: 45 },
        ],
        nearKCap: 9,
      },
      success: {
        greenPct: 7433,
        greenCount: 223,
        total: 300,
        notifiedBeforeDay10Pct: 2666,
        notifiedBeforeDay10Count: 80,
        notifiedCount: 60,
        paidOnTimePct: 6250,
        paidOnTimeCount: 38,
      },
    });
  }),
];
