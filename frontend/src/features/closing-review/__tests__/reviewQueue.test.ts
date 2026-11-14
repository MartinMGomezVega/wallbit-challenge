import { describe, it, expect } from 'vitest';
import {
  createQueue,
  canNext,
  canPrev,
  current,
  next,
  prev,
  markReviewed,
  pageOf,
  progress,
} from '../reviewQueue';

describe('reviewQueue', () => {
  it('crea una cola inicial con startId en la posición correcta', () => {
    const ids = ['clo_1', 'clo_2', 'clo_3'];
    const q = createQueue(ids, 'clo_2', 'Filtro Rojos');

    expect(q.ids).toEqual(['clo_1', 'clo_2', 'clo_3']);
    expect(q.index).toBe(1);
    expect(current(q)).toBe('clo_2');
    expect(q.label).toBe('Filtro Rojos');
    expect(q.isOutsideFilter).toBe(false);
    expect(q.reviewed.size).toBe(0);
  });

  it('si startId no está presente, lo antepone y marca isOutsideFilter = true (Criterio 6)', () => {
    const ids = ['clo_1', 'clo_2', 'clo_3'];
    const q = createQueue(ids, 'clo_999', 'Filtro Rojos');

    expect(q.ids).toEqual(['clo_999', 'clo_1', 'clo_2', 'clo_3']);
    expect(q.index).toBe(0);
    expect(current(q)).toBe('clo_999');
    expect(q.isOutsideFilter).toBe(true);
  });

  it('si startId no se pasa, arranca en el primer elemento (índice 0)', () => {
    const ids = ['clo_1', 'clo_2'];
    const q = createQueue(ids);

    expect(q.index).toBe(0);
    expect(current(q)).toBe('clo_1');
    expect(q.isOutsideFilter).toBe(false);
  });

  it('canPrev y canNext detectan correctamente los bordes', () => {
    const ids = ['clo_1', 'clo_2', 'clo_3'];
    const qStart = createQueue(ids, 'clo_1');

    expect(canPrev(qStart)).toBe(false);
    expect(canNext(qStart)).toBe(true);

    const qMid = next(qStart);
    expect(canPrev(qMid)).toBe(true);
    expect(canNext(qMid)).toBe(true);

    const qEnd = next(qMid);
    expect(canPrev(qEnd)).toBe(true);
    expect(canNext(qEnd)).toBe(false);
  });

  it('next no se pasa del borde final', () => {
    const ids = ['clo_1', 'clo_2'];
    let q = createQueue(ids, 'clo_1');
    q = next(q);
    expect(q.index).toBe(1);
    expect(current(q)).toBe('clo_2');

    const same = next(q);
    expect(same.index).toBe(1);
    expect(same).toBe(q);
  });

  it('prev no se pasa del borde inicial', () => {
    const ids = ['clo_1', 'clo_2'];
    const q = createQueue(ids, 'clo_1');
    const same = prev(q);
    expect(same.index).toBe(0);
    expect(same).toBe(q);
  });

  it('navegar hacia adelante apaga isOutsideFilter y regresar a 0 lo mantiene si era outside', () => {
    const ids = ['clo_1', 'clo_2'];
    let q = createQueue(ids, 'clo_outside');
    expect(q.isOutsideFilter).toBe(true);

    q = next(q);
    expect(q.index).toBe(1);
    expect(q.isOutsideFilter).toBe(false);

    q = prev(q);
    expect(q.index).toBe(0);
    expect(q.isOutsideFilter).toBe(true);
  });

  it('markReviewed agrega el ID actual o explícito al conjunto de revisados', () => {
    const ids = ['clo_1', 'clo_2', 'clo_3'];
    let q = createQueue(ids, 'clo_1');

    q = markReviewed(q);
    expect(q.reviewed.has('clo_1')).toBe(true);
    expect(progress(q)).toEqual({ reviewedCount: 1, total: 3 });

    // Con ID explícito
    q = markReviewed(q, 'clo_3');
    expect(q.reviewed.has('clo_3')).toBe(true);
    expect(progress(q)).toEqual({ reviewedCount: 2, total: 3 });

    // En cola vacía
    const emptyQ = createQueue([]);
    expect(markReviewed(emptyQ)).toBe(emptyQ);
  });

  it('pageOf calcula la página correspondiente al índice y pageSize', () => {
    const ids = Array.from({ length: 150 }, (_, i) => `clo_${i}`);
    const q1 = createQueue(ids, 'clo_0');
    expect(pageOf(q1, 50)).toBe(1);

    const q49 = createQueue(ids, 'clo_49');
    expect(pageOf(q49, 50)).toBe(1);

    const q50 = createQueue(ids, 'clo_50');
    expect(pageOf(q50, 50)).toBe(2);

    const q100 = createQueue(ids, 'clo_100');
    expect(pageOf(q100, 50)).toBe(3);

    // Casos borde
    expect(pageOf(q1, 0)).toBe(1);
    expect(pageOf(createQueue([]), 50)).toBe(1);
  });
});

