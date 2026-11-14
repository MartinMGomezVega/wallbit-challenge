import '@testing-library/jest-dom/vitest';
import { cleanup } from '@testing-library/react';
import { afterAll, afterEach, beforeAll } from 'vitest';
import { server } from '../mocks/server';
import { resetRoleForTests } from '../app/roleStore';

// Polyfill para compatibilidad entre jsdom y undici (Node 22/25 Web API):
// jsdom crea instancias de AbortSignal que no tienen los símbolos internos de undici,
// provocando "Expected signal to be an instance of AbortSignal".
const NativeRequest = globalThis.Request;
if (NativeRequest) {
  globalThis.Request = class extends NativeRequest {
    constructor(input: RequestInfo | URL, init?: RequestInit) {
      if (init && 'signal' in init && init.signal) {
        const safeInit = { ...init };
        delete safeInit.signal;
        super(input, safeInit);
        return;
      }
      super(input, init);
    }
  } as typeof NativeRequest;

  if (typeof window !== 'undefined') {
    window.Request = globalThis.Request;
  }
}

// Polyfill robusto de localStorage para jsdom
const storage = new Map<string, string>();
const localStorageMock: Storage = {
  getItem: (key: string) => storage.get(key) ?? null,
  setItem: (key: string, value: string) => {
    storage.set(key, String(value));
  },
  removeItem: (key: string) => {
    storage.delete(key);
  },
  clear: () => {
    storage.clear();
  },
  key: (index: number) => Array.from(storage.keys())[index] ?? null,
  get length() {
    return storage.size;
  },
};
Object.defineProperty(window, 'localStorage', {
  value: localStorageMock,
  configurable: true,
  writable: true,
});

// Polyfills que Radix necesita en jsdom.
class ResizeObserverStub {
  observe() {}
  unobserve() {}
  disconnect() {}
}
globalThis.ResizeObserver ??= ResizeObserverStub as unknown as typeof ResizeObserver;
if (!Element.prototype.hasPointerCapture) {
  Element.prototype.hasPointerCapture = () => false;
  Element.prototype.releasePointerCapture = () => {};
  Element.prototype.setPointerCapture = () => {};
}
Element.prototype.scrollIntoView ??= function scrollIntoView() {};
window.HTMLElement.prototype.scrollIntoView ??= function scrollIntoView() {};

beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));
afterEach(() => {
  cleanup();
  server.resetHandlers();
  window.localStorage.clear();
  resetRoleForTests();
});
afterAll(() => server.close());
