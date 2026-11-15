import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import {
  getClientProfile,
  getClientClosings,
  getClientInvoices,
  createClientInvoice,
  importClientInvoicesCSV,
  setClientExternalIncome,
  confirmClientPeriod,
  markClientClosingPaid,
} from '../api';

function getFetchCall(index: number): [string, RequestInit] {
  const calls = (globalThis.fetch as unknown as { mock: { calls: [string, RequestInit][] } }).mock.calls;
  const call = calls[index];
  if (!call) throw new Error(`Expected fetch call at index ${index}`);
  return call;
}

describe('client-demo api', () => {
  const originalFetch = globalThis.fetch;

  beforeEach(() => {
    vi.restoreAllMocks();
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
  });

  it('getClientProfile sends request with X-Demo-User and returns profile', async () => {
    const mockProfile = {
      id: 'cli_demo_15',
      name: 'Paula Rossi',
      cuitMasked: '27-***-4',
      regime: 'MONOTRIBUTO',
      category: 'H',
    };
    globalThis.fetch = vi.fn().mockImplementation(() =>
      Promise.resolve(
        new Response(JSON.stringify(mockProfile), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        }),
      ),
    );

    const res = await getClientProfile('cli_demo_15');
    expect(res).toEqual(mockProfile);

    const [url, init] = getFetchCall(0);
    expect(url).toBe('/v1/clients/cli_demo_15');
    expect((init.headers as Headers).get('X-Demo-User')).toBe('client:cli_demo_15');
  });

  it('getClientClosings returns client closings array', async () => {
    const mockClosings = [
      {
        closingId: 'cls_1',
        period: '2026-10',
        state: 'READY' as const,
        input: {
          answered: true,
          confirmed: true,
          invoicesCount: 4,
        },
      },
    ];
    globalThis.fetch = vi.fn().mockImplementation(() =>
      Promise.resolve(
        new Response(JSON.stringify(mockClosings), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        }),
      ),
    );

    const res = await getClientClosings('cli_demo_15');
    expect(res).toEqual(mockClosings);
  });

  it('getClientInvoices supports optional period query param', async () => {
    const mockInvoices = [
      {
        id: 'inv_1',
        clientId: 'cli_demo_15',
        type: 'E',
        pointOfSale: 1,
        number: 12,
        issueDate: '2026-10-15',
        currency: 'USD',
        netAmount: { amount: '2500.00', currency: 'USD' },
        vatAmount: { amount: '0.00', currency: 'USD' },
        totalAmount: { amount: '2500.00', currency: 'USD' },
        fxRate: '1450.00',
        receiverName: 'Acme Corp',
      },
    ];
    globalThis.fetch = vi.fn().mockImplementation(() =>
      Promise.resolve(
        new Response(JSON.stringify(mockInvoices), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        }),
      ),
    );

    const resWithPeriod = await getClientInvoices('cli_demo_15', '2026-10');
    expect(resWithPeriod).toEqual(mockInvoices);
    expect(getFetchCall(0)[0]).toBe('/v1/clients/cli_demo_15/invoices?period=2026-10');

    const resWithoutPeriod = await getClientInvoices('cli_demo_15');
    expect(resWithoutPeriod).toEqual(mockInvoices);
    expect(getFetchCall(1)[0]).toBe('/v1/clients/cli_demo_15/invoices');
  });

  it('createClientInvoice sends JSON payload and returns created invoice', async () => {
    const payload = {
      type: 'E',
      pointOfSale: 1,
      number: 99,
      issueDate: '2026-10-20',
      currency: 'USD',
      netAmount: { amount: '1500.00', currency: 'USD' },
      vatAmount: { amount: '0.00', currency: 'USD' },
      totalAmount: { amount: '1500.00', currency: 'USD' },
      receiverName: 'Test Client',
    };
    const mockCreated = {
      id: 'inv_99',
      clientId: 'cli_demo_15',
      ...payload,
    };
    globalThis.fetch = vi.fn().mockImplementation(() =>
      Promise.resolve(
        new Response(JSON.stringify(mockCreated), {
          status: 201,
          headers: { 'Content-Type': 'application/json' },
        }),
      ),
    );

    const res = await createClientInvoice('cli_demo_15', payload);
    expect(res.id).toBe('inv_99');

    const [url, init] = getFetchCall(0);
    expect(url).toBe('/v1/clients/cli_demo_15/invoices');
    expect(init.method).toBe('POST');
    expect((init.headers as Headers).get('Content-Type')).toBe('application/json');
    expect((init.headers as Headers).get('X-Demo-User')).toBe('client:cli_demo_15');
  });

  it('importClientInvoicesCSV sends FormData and returns summary', async () => {
    const file = new File(['mock,csv'], 'comprobantes.csv', { type: 'text/csv' });
    const mockResult = {
      imported: 3,
      skippedDuplicates: 0,
      warnings: [],
      errors: [],
    };
    globalThis.fetch = vi.fn().mockImplementation(() =>
      Promise.resolve(
        new Response(JSON.stringify(mockResult), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        }),
      ),
    );

    const res = await importClientInvoicesCSV('cli_demo_15', file);
    expect(res).toEqual(mockResult);

    const [url, init] = getFetchCall(0);
    expect(url).toBe('/v1/clients/cli_demo_15/invoice-imports');
    expect(init.method).toBe('POST');
    expect(init.body).toBeInstanceOf(FormData);
    expect((init.headers as Headers).get('X-Demo-User')).toBe('client:cli_demo_15');
  });

  it('setClientExternalIncome sends external income payload', async () => {
    const payload = {
      none: false,
      items: [{ amount: { amount: '500.00', currency: 'USD' }, description: 'Consulting' }],
    };
    const mockRes = {
      period: '2026-10',
      confirmedAt: null,
      declaredNoneExternalIncome: false,
      externalIncome: [{ amount: { amount: '500.00', currency: 'USD' }, description: 'Consulting' }],
    };
    globalThis.fetch = vi.fn().mockImplementation(() =>
      Promise.resolve(
        new Response(JSON.stringify(mockRes), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        }),
      ),
    );

    const res = await setClientExternalIncome('cli_demo_15', '2026-10', payload);
    expect(res.period).toBe('2026-10');
    expect(getFetchCall(0)[0]).toBe('/v1/clients/cli_demo_15/periods/2026-10/external-income');
  });

  it('confirmClientPeriod sends POST to /v1/clients/:id/periods/:period/confirmation', async () => {
    const mockRes = {
      period: '2026-10',
      confirmedAt: '2026-11-14T20:00:00Z',
      declaredNoneExternalIncome: true,
      externalIncome: [],
    };
    globalThis.fetch = vi.fn().mockImplementation(() =>
      Promise.resolve(
        new Response(JSON.stringify(mockRes), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        }),
      ),
    );

    const res = await confirmClientPeriod('cli_demo_15', '2026-10');
    expect(res.confirmedAt).toBe('2026-11-14T20:00:00Z');
    expect(getFetchCall(0)[0]).toBe('/v1/clients/cli_demo_15/periods/2026-10/confirmation');
  });

  it('markClientClosingPaid sends PUT to /v1/closings/:id/payment', async () => {
    const mockRes = {
      closingId: 'cls_1',
      paidAt: '2026-11-14T20:00:00Z',
    };
    globalThis.fetch = vi.fn().mockImplementation(() =>
      Promise.resolve(
        new Response(JSON.stringify(mockRes), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        }),
      ),
    );

    const res = await markClientClosingPaid('cli_demo_15', 'cls_1');
    expect(res.paidAt).toBe('2026-11-14T20:00:00Z');
    const [url, init] = getFetchCall(0);
    expect(url).toBe('/v1/closings/cls_1/payment');
    expect(init.method).toBe('PUT');
  });
});
