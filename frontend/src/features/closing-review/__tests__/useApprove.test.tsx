import { describe, it, expect, vi } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useApprove } from '../useApprove';
import { server } from '../../../mocks/server';
import { http, HttpResponse } from 'msw';

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { mutations: { retry: false }, queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
}

describe('useApprove', () => {
  it('ejecuta la mutación con éxito y llama a onSuccess', async () => {
    const onSuccess = vi.fn();
    const { result } = renderHook(() => useApprove({ onSuccess }), {
      wrapper: createWrapper(),
    });

    result.current.mutate({
      closingId: 'cls_01',
      version: 1,
      clientNote: 'Nota test',
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(onSuccess).toHaveBeenCalled();
  });

  it('llama a onError si la mutación falla', async () => {
    server.use(
      http.post('/api/v1/accountant/closings/:id/approve', () => {
        return HttpResponse.json(
          { error: { code: 'VERSION_CONFLICT', message: 'Conflicto' } },
          { status: 409 }
        );
      })
    );

    const onError = vi.fn();
    const { result } = renderHook(() => useApprove({ onError }), {
      wrapper: createWrapper(),
    });

    result.current.mutate({
      closingId: 'cls_01',
      version: 1,
    });

    await waitFor(() => {
      expect(result.current.isError).toBe(true);
    });

    expect(onError).toHaveBeenCalled();
  });

  it('funciona sin callbacks opcionales', async () => {
    const { result } = renderHook(() => useApprove(), {
      wrapper: createWrapper(),
    });

    result.current.mutate({
      closingId: 'cls_01',
      version: 1,
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });
  });
});

