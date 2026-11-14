import type { ZodType, ZodTypeDef } from 'zod';
import { getRole, roleHeader } from '../../app/roleStore';

export type ApiErrorFields = Record<string, string>;

/** Error estándar de la API: `{ error: { code, message, fields } }`. */
export class ApiError extends Error {
  readonly status: number;
  readonly code: string;
  readonly fields: ApiErrorFields;

  constructor(status: number, code: string, message: string, fields: ApiErrorFields = {}) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.code = code;
    this.fields = fields;
  }
}

type ErrorEnvelope = { error?: { code?: unknown; message?: unknown; fields?: unknown } };

function toApiError(status: number, body: unknown): ApiError {
  const err = (body as ErrorEnvelope | null)?.error;
  if (err && typeof err.code === 'string') {
    const fields = err.fields && typeof err.fields === 'object' ? (err.fields as ApiErrorFields) : {};
    return new ApiError(status, err.code, typeof err.message === 'string' ? err.message : err.code, fields);
  }
  return new ApiError(status, 'HTTP_ERROR', `Error ${status}`);
}

async function readJson(res: Response): Promise<unknown> {
  const text = await res.text();
  if (!text) return null;
  try {
    return JSON.parse(text);
  } catch {
    return undefined;
  }
}

/**
 * fetch a la API: agrega X-Demo-User según el rol, valida la respuesta con zod
 * y convierte errores al formato estándar.
 */
export async function apiFetch<T>(path: string, schema: ZodType<T, ZodTypeDef, unknown>, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers);
  headers.set('X-Demo-User', roleHeader(getRole()));
  if (init.body && typeof init.body === 'string' && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json');
  }
  let res: Response;
  try {
    res = await fetch(path, { ...init, headers });
  } catch {
    throw new ApiError(0, 'NETWORK_ERROR', 'No pudimos conectarnos con el servidor.');
  }
  const body = await readJson(res);
  if (!res.ok) throw toApiError(res.status, body);
  const parsed = schema.safeParse(body);
  if (!parsed.success) {
    throw new ApiError(res.status, 'INVALID_RESPONSE', 'Respuesta inesperada del servidor.');
  }
  return parsed.data;
}

/** Helper para JSON bodies. */
export function jsonBody(method: string, data: unknown): RequestInit {
  return { method, body: JSON.stringify(data), headers: { 'Content-Type': 'application/json' } };
}

