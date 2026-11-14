import { describe, it, expect } from 'vitest';
import {
  formatMoney,
  formatMoneyParts,
  formatDate,
  formatDueDate,
  formatPeriodName,
  formatPercentBp,
} from '../index';

describe('shared/format', () => {
  describe('formatMoney', () => {
    it('formatea montos en ARS según es-AR ($ 1.234.567,89)', () => {
      expect(formatMoney({ amount: 123456789, currency: 'ARS' })).toBe('$ 1.234.567,89');
      expect(formatMoney({ amount: 0, currency: 'ARS' })).toBe('$ 0,00');
      expect(formatMoney({ amount: 50, currency: 'ARS' })).toBe('$ 0,50');
      expect(formatMoney({ amount: -150000, currency: 'ARS' })).toBe('-$ 1.500,00');
    });

    it('formatea montos en USD según es-AR (US$ 1.234,56)', () => {
      expect(formatMoney({ amount: 123456, currency: 'USD' })).toBe('US$ 1.234,56');
      expect(formatMoney({ amount: 500000, currency: 'USD' })).toBe('US$ 5.000,00');
      expect(formatMoney({ amount: 0, currency: 'USD' })).toBe('US$ 0,00');
    });

    it('devuelve "-" si money es nulo o indefinido', () => {
      expect(formatMoney(null)).toBe('-');
      expect(formatMoney(undefined)).toBe('-');
    });

    it('formatMoneyParts maneja centavos y moneda directamente', () => {
      expect(formatMoneyParts(75000, 'USD')).toBe('US$ 750,00');
      expect(formatMoneyParts(25000000, 'ARS')).toBe('$ 250.000,00');
    });
  });

  describe('formatDate', () => {
    it('formatea fechas en modo corto (dd/MM)', () => {
      expect(formatDate('2026-10-15', 'short')).toBe('15/10');
      expect(formatDate('2026-05-03', 'short')).toBe('03/05');
      expect(formatDate('2026-10-15T14:30:00Z', 'short')).toBe('15/10');
      expect(formatDate(new Date(2026, 9, 15), 'short')).toBe('15/10');
    });

    it('formatea fechas en modo full (dd/MM/yyyy)', () => {
      expect(formatDate('2026-10-15', 'full')).toBe('15/10/2026');
      expect(formatDate('2026-01-05T00:00:00Z', 'full')).toBe('05/01/2026');
      expect(formatDate(new Date(2026, 9, 15), 'full')).toBe('15/10/2026');
    });

    it('devuelve "-" para valores nulos o inválidos', () => {
      expect(formatDate(null)).toBe('-');
      expect(formatDate(undefined)).toBe('-');
      expect(formatDate('invalid-date')).toBe('-');
    });

    it('formatDueDate formatea "vence dd/MM"', () => {
      expect(formatDueDate('2026-11-20')).toBe('vence 20/11');
      expect(formatDueDate(null)).toBe('-');
    });

    it('formatPeriodName convierte YYYY-MM a "mes año"', () => {
      expect(formatPeriodName('2026-10')).toBe('octubre 2026');
      expect(formatPeriodName('2025-01')).toBe('enero 2025');
      expect(formatPeriodName('invalid')).toBe('invalid');
    });
  });

  describe('formatPercentBp', () => {
    it('formatea basis points a porcentaje', () => {
      expect(formatPercentBp(8700)).toBe('87%');
      expect(formatPercentBp(10000)).toBe('100%');
      expect(formatPercentBp(7540)).toBe('75%');
      expect(formatPercentBp(0)).toBe('0%');
      expect(formatPercentBp(null)).toBe('0%');
      expect(formatPercentBp(undefined)).toBe('0%');
    });
  });
});
