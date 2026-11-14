import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act, waitFor } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter } from 'react-router-dom';
import { server } from '../../../mocks/server';
import { ReviewQueueProvider, useReviewQueue } from '../useReviewQueue';

function createWrapper(initialUrl = '/contador?period=2026-10', onCloseQueue = vi.fn()) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });

  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={[initialUrl]}>
        <ReviewQueueProvider pageSize={2} onCloseQueue={onCloseQueue}>
          {children}
        </ReviewQueueProvider>
      </MemoryRouter>
    </QueryClientProvider>
  );
}

describe('useReviewQueue', () => {
  beforeEach(() => {
    server.use(
      http.get('/api/v1/accountant/periods/:period/closings/ids', () => {
        return HttpResponse.json({
          ids: ['clo_01', 'clo_02', 'clo_03'],
          total: 3,
        });
      }),
      http.get('/api/v1/accountant/closings/:id', () => {
        return HttpResponse.json({
          closing: {
            id: 'clo_01',
            clientId: 'cli_01',
            period: '2026-10',
            status: 'READY_FOR_REVIEW',
            version: 1,
            resultHash: 'hash',
            severity: 'RED',
            description: 'Test',
            signals: [],
            computedAt: '2026-11-01T00:00:00Z',
          },
          client: {
            id: 'cli_01',
            name: 'Test Client',
            cuit: '20-11111111-1',
            regime: 'MONOTRIBUTO',
          },
          periodMovements: [],
          periodInvoices: [],
          reconciliation: { matchedPairs: [], unmatchedIncome: [], unmatchedInvoices: [] },
          clientInput: {},
        });
      })
    );
  });

  it('lanza error si se usa fuera de ReviewQueueProvider', () => {
    expect(() => renderHook(() => useReviewQueue())).toThrow(
      'useReviewQueue debe ser usado dentro de ReviewQueueProvider'
    );
  });

  it('abre la cola en modo row, sincroniza URL y navega entre clientes', async () => {
    const { result } = renderHook(() => useReviewQueue(), {
      wrapper: createWrapper(),
    });

    expect(result.current.isOpen).toBe(false);

    await act(async () => {
      await result.current.openQueue({
        mode: 'row',
        startId: 'clo_02',
        label: 'Rojos',
      });
    });

    expect(result.current.isOpen).toBe(true);
    expect(result.current.currentClosingId).toBe('clo_02');
    expect(result.current.canGoPrev).toBe(true);
    expect(result.current.canGoNext).toBe(true);

    // Navegar al siguiente
    act(() => {
      result.current.goNext();
    });
    expect(result.current.currentClosingId).toBe('clo_03');
    expect(result.current.canGoNext).toBe(false);

    // Navegar al anterior
    act(() => {
      result.current.goPrev();
    });
    expect(result.current.currentClosingId).toBe('clo_02');
  });

  it('marca como revisado y calcula el progreso', async () => {
    const { result } = renderHook(() => useReviewQueue(), {
      wrapper: createWrapper(),
    });

    await act(async () => {
      await result.current.openQueue({
        mode: 'selection',
        ids: ['clo_01', 'clo_02'],
        startId: 'clo_01',
      });
    });

    expect(result.current.progress).toEqual({ reviewedCount: 0, total: 2 });

    act(() => {
      result.current.markCurrentReviewed();
    });

    expect(result.current.progress).toEqual({ reviewedCount: 1, total: 2 });
  });

  it('cierra la cola, limpia la URL y reporta targetPage y lastViewedId al callback', async () => {
    const onCloseQueue = vi.fn();
    const { result } = renderHook(() => useReviewQueue(), {
      wrapper: createWrapper('/contador?period=2026-10', onCloseQueue),
    });

    await act(async () => {
      await result.current.openQueue({
        mode: 'selection',
        ids: ['clo_01', 'clo_02', 'clo_03'], // con pageSize=2, clo_03 (index 2) está en página 2
        startId: 'clo_03',
      });
    });

    act(() => {
      const closed = result.current.closeQueue();
      expect(closed.lastViewedId).toBe('clo_03');
      expect(closed.targetPage).toBe(2);
    });

    expect(result.current.isOpen).toBe(false);
    expect(onCloseQueue).toHaveBeenCalledWith({
      lastViewedId: 'clo_03',
      targetPage: 2,
    });
  });

  it('reabre la cola automáticamente si la URL inicial contiene client= (Criterio 6)', async () => {
    const { result } = renderHook(() => useReviewQueue(), {
      wrapper: createWrapper('/contador?period=2026-10&client=clo_02'),
    });

    await waitFor(() => {
      expect(result.current.isOpen).toBe(true);
    });

    expect(result.current.currentClosingId).toBe('clo_02');
  });
});
