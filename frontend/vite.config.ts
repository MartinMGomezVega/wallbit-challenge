/// <reference types="vitest" />
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/api': 'http://localhost:8080',
    },
  },
  test: {
    environment: 'jsdom',
    setupFiles: ['src/test/setup.ts'],
    globals: false,
    css: false,
    coverage: {
      provider: 'v8',
      include: ['src/**/*.{ts,tsx}'],
      exclude: ['src/main.tsx', 'src/**/*.d.ts', 'src/test/**', 'src/mocks/**', 'src/**/__tests__/**', 'src/**/*.test.{ts,tsx}'],
      thresholds: { lines: 90, branches: 90, functions: 90, statements: 90 },
      reporter: ['text-summary', 'text'],
    },
  },
});

