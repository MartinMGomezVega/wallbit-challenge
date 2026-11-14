export interface ReviewQueue {
  ids: string[];
  index: number;
  reviewed: Set<string>;
  label: string;
  outsideId?: string;
  isOutsideFilter?: boolean;
}

/**
 * Crea una cola inmutable de revisión de clientes.
 * Criterio 6: si startId no está en ids, se antepone y se marca `isOutsideFilter: true`.
 */
export function createQueue(
  ids: string[],
  startId?: string,
  label: string = 'Todos los clientes'
): ReviewQueue {
  const safeIds = [...ids];
  let outsideId: string | undefined;
  let index = 0;

  if (startId) {
    const existingIndex = safeIds.indexOf(startId);
    if (existingIndex >= 0) {
      index = existingIndex;
    } else {
      safeIds.unshift(startId);
      index = 0;
      outsideId = startId;
    }
  }

  return {
    ids: safeIds,
    index,
    reviewed: new Set<string>(),
    label,
    outsideId,
    isOutsideFilter: Boolean(outsideId && index === 0),
  };
}

export function canNext(q: ReviewQueue): boolean {
  return q.index < q.ids.length - 1;
}

export function canPrev(q: ReviewQueue): boolean {
  return q.index > 0;
}

export function current(q: ReviewQueue): string | undefined {
  return q.ids[q.index];
}

export function next(q: ReviewQueue): ReviewQueue {
  if (!canNext(q)) return q;
  const newIndex = q.index + 1;
  return {
    ...q,
    index: newIndex,
    isOutsideFilter: Boolean(q.outsideId && newIndex === 0),
  };
}

export function prev(q: ReviewQueue): ReviewQueue {
  if (!canPrev(q)) return q;
  const newIndex = q.index - 1;
  return {
    ...q,
    index: newIndex,
    isOutsideFilter: Boolean(q.outsideId && newIndex === 0),
  };
}

export function markReviewed(q: ReviewQueue, id?: string): ReviewQueue {
  const targetId = id ?? current(q);
  if (!targetId) return q;
  const nextReviewed = new Set(q.reviewed);
  nextReviewed.add(targetId);
  return {
    ...q,
    reviewed: nextReviewed,
  };
}

/**
 * Calcula la página en la tabla (1-based) en la que se encuentra el cliente actual según el pageSize.
 * Criterio 8: si está en otra página, navega a esa página.
 */
export function pageOf(q: ReviewQueue, pageSize: number): number {
  if (pageSize <= 0 || q.ids.length === 0) return 1;
  return Math.floor(q.index / pageSize) + 1;
}

export function progress(q: ReviewQueue): { reviewedCount: number; total: number } {
  return {
    reviewedCount: q.reviewed.size,
    total: q.ids.length,
  };
}
