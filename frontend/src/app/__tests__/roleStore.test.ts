import { describe, it, expect, vi, beforeEach } from 'vitest';
import {
  getRole,
  setRole,
  roleHeader,
  resetRoleForTests,
  subscribeRole,
} from '../roleStore';

describe('roleStore', () => {
  beforeEach(() => {
    resetRoleForTests();
    window.localStorage.clear();
  });

  it('inicia por defecto como accountant', () => {
    expect(getRole()).toEqual({ kind: 'accountant' });
    expect(roleHeader(getRole())).toBe('accountant');
  });

  it('cambia el rol y persiste en localStorage', () => {
    setRole({ kind: 'client', clientId: 'cli_01' });
    expect(getRole()).toEqual({ kind: 'client', clientId: 'cli_01' });
    expect(roleHeader(getRole())).toBe('client:cli_01');
    expect(window.localStorage.getItem('wallbit.demoRole')).toBe('client:cli_01');
  });

  it('recupera el rol desde localStorage al inicializar', () => {
    window.localStorage.setItem('wallbit.demoRole', 'client:cli_persisted');
    expect(getRole()).toEqual({ kind: 'client', clientId: 'cli_persisted' });
  });

  it('ignora valores inválidos en localStorage y usa DEFAULT_ROLE', () => {
    window.localStorage.setItem('wallbit.demoRole', 'invalid_role');
    expect(getRole()).toEqual({ kind: 'accountant' });

    resetRoleForTests();
    window.localStorage.setItem('wallbit.demoRole', 'client:'); // clientId vacío
    expect(getRole()).toEqual({ kind: 'accountant' });
  });

  it('tolera cuando localStorage lanza excepciones (modo privado / cuota)', () => {
    const originalSetItem = window.localStorage.setItem;
    const originalGetItem = window.localStorage.getItem;

    vi.spyOn(window.localStorage, 'setItem').mockImplementation(() => {
      throw new Error('QuotaExceededError');
    });
    vi.spyOn(window.localStorage, 'getItem').mockImplementation(() => {
      throw new Error('SecurityError');
    });

    // No debe lanzar error
    expect(() => setRole({ kind: 'client', clientId: 'cli_fallback' })).not.toThrow();
    expect(getRole()).toEqual({ kind: 'client', clientId: 'cli_fallback' });

    window.localStorage.setItem = originalSetItem;
    window.localStorage.getItem = originalGetItem;
  });

  it('notifica a los suscriptores y permite desuscribirse', () => {
    const listener = vi.fn();
    const unsub = subscribeRole(listener);

    setRole({ kind: 'accountant' });
    expect(listener).toHaveBeenCalledTimes(1);

    unsub();
    setRole({ kind: 'client', clientId: 'cli_test' });
    expect(listener).toHaveBeenCalledTimes(1);
  });
});
