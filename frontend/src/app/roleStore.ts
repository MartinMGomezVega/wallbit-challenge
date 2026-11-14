import { useSyncExternalStore } from 'react';

export type Role = { kind: 'accountant' } | { kind: 'client'; clientId: string };

const STORAGE_KEY = 'wallbit.demoRole';
const DEFAULT_ROLE: Role = { kind: 'accountant' };

let memoryRole: Role | null = null;
const listeners = new Set<() => void>();

function parseRole(raw: string | null): Role | null {
  if (!raw) return null;
  if (raw === 'accountant') return { kind: 'accountant' };
  if (raw.startsWith('client:') && raw.length > 'client:'.length) {
    return { kind: 'client', clientId: raw.slice('client:'.length) };
  }
  return null;
}

/** Valor del header X-Demo-User para un rol. */
export function roleHeader(role: Role): string {
  return role.kind === 'accountant' ? 'accountant' : `client:${role.clientId}`;
}

function readStorage(): Role | null {
  try {
    return parseRole(window.localStorage.getItem(STORAGE_KEY));
  } catch {
    return null;
  }
}

export function getRole(): Role {
  if (memoryRole) return memoryRole;
  memoryRole = readStorage() ?? DEFAULT_ROLE;
  return memoryRole;
}

export function setRole(role: Role): void {
  memoryRole = role;
  try {
    window.localStorage.setItem(STORAGE_KEY, roleHeader(role));
  } catch {
    // localStorage no disponible (modo privado, cuota): queda en memoria.
  }
  listeners.forEach((l) => l());
}

export function subscribeRole(listener: () => void): () => void {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
}

/** Solo para tests: olvida el rol en memoria. */
export function resetRoleForTests(): void {
  memoryRole = null;
}

export function useRole(): Role {
  return useSyncExternalStore(subscribeRole, getRole, getRole);
}

