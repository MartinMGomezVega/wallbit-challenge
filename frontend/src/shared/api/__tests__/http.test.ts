import { describe, it, expect } from 'vitest';
import { z } from 'zod';
import { http, HttpResponse } from 'msw';
import { server } from '../../../mocks/server';
import { apiFetch, ApiError, jsonBody } from '../http';

describe('apiFetch', () => {
  const schema = z.object({
    status: z.string(),
    count: z.number().optional(),
  });

  it('realiza una petición exitosa y parsea con zod', async () => {
    server.use(
      http.get('/api/v1/test', ({ request }) => {
        expect(request.headers.get('X-Demo-User')).toBe('accountant');
        return HttpResponse.json({ status: 'ok', count: 42 });
      }),
    );

    const res = await apiFetch('/api/v1/test', schema);
    expect(res).toEqual({ status: 'ok', count: 42 });
  });

  it('agrega Content-Type application/json automáticamente si hay body string sin header', async () => {
    server.use(
      http.post('/api/v1/test-body', async ({ request }) => {
        expect(request.headers.get('Content-Type')).toBe('application/json');
        return HttpResponse.json({ status: 'ok' });
      }),
    );

    const res = await apiFetch('/api/v1/test-body', schema, {
      method: 'POST',
      body: '{"foo":"bar"}',
    });
    expect(res.status).toBe('ok');
  });

  it('respeta Content-Type explícito si se provee', async () => {
    server.use(
      http.post('/api/v1/test-custom-ct', async ({ request }) => {
        expect(request.headers.get('Content-Type')).toBe('text/plain');
        return HttpResponse.json({ status: 'ok' });
      }),
    );

    const res = await apiFetch('/api/v1/test-custom-ct', schema, {
      method: 'POST',
      body: 'plain text',
      headers: { 'Content-Type': 'text/plain' },
    });
    expect(res.status).toBe('ok');
  });

  it('maneja errores estándar de la API con código y campos', async () => {
    server.use(
      http.post('/api/v1/test-err', () => {
        return HttpResponse.json(
          {
            error: {
              code: 'VALIDATION_ERROR',
              message: 'Datos inválidos',
              fields: { email: 'inválido' },
            },
          },
          { status: 422 },
        );
      }),
    );

    await expect(apiFetch('/api/v1/test-err', schema, jsonBody('POST', {}))).rejects.toSatisfy(
      (err: unknown) => {
        return (
          err instanceof ApiError &&
          err.status === 422 &&
          err.code === 'VALIDATION_ERROR' &&
          err.message === 'Datos inválidos' &&
          err.fields.email === 'inválido'
        );
      },
    );
  });

  it('maneja error HTTP sin payload estructurado', async () => {
    server.use(
      http.get('/api/v1/test-500', () => {
        return new HttpResponse('Error no JSON', { status: 500 });
      }),
    );

    await expect(apiFetch('/api/v1/test-500', schema)).rejects.toSatisfy(
      (err: unknown) => err instanceof ApiError && err.status === 500 && err.code === 'HTTP_ERROR',
    );
  });

  it('maneja esquema de respuesta inválido (INVALID_RESPONSE)', async () => {
    server.use(
      http.get('/api/v1/test-bad-schema', () => {
        return HttpResponse.json({ status: 123 }); // esperaba string
      }),
    );

    await expect(apiFetch('/api/v1/test-bad-schema', schema)).rejects.toSatisfy(
      (err: unknown) => err instanceof ApiError && err.code === 'INVALID_RESPONSE',
    );
  });

  it('maneja error de red (NETWORK_ERROR)', async () => {
    server.use(
      http.get('/api/v1/test-network', () => {
        return HttpResponse.error();
      }),
    );

    await expect(apiFetch('/api/v1/test-network', schema)).rejects.toSatisfy(
      (err: unknown) => err instanceof ApiError && err.code === 'NETWORK_ERROR',
    );
  });
});
