import React, {
  createContext,
  useContext,
  useState,
  useEffect,
  useCallback,
  useMemo,
} from 'react';
import { useSearchParams } from 'react-router-dom';
import { useQueryClient } from '@tanstack/react-query';
import {
  ReviewQueue,
  createQueue,
  next,
  prev,
  canNext,
  canPrev,
  current,
  markReviewed,
  pageOf,
  progress,
} from './reviewQueue';
import { getClosingDetail, getClosingIds } from './api';

export interface OpenQueueOptions {
  mode: 'row' | 'chip' | 'selection';
  ids?: string[];
  startId?: string;
  label?: string;
  period?: string;
  filterParams?: URLSearchParams | Record<string, string | number | undefined>;
}

export interface ReviewQueueContextValue {
  queue: ReviewQueue | null;
  isOpen: boolean;
  currentClosingId: string | undefined;
  canGoNext: boolean;
  canGoPrev: boolean;
  progress: { reviewedCount: number; total: number };
  isOutsideFilter: boolean;
  openQueue: (options: OpenQueueOptions) => Promise<void>;
  closeQueue: () => { lastViewedId?: string; targetPage?: number };
  goNext: () => void;
  goPrev: () => void;
  markCurrentReviewed: () => void;
}

const ReviewQueueContext = createContext<ReviewQueueContextValue | null>(null);

export interface ReviewQueueProviderProps {
  children: React.ReactNode;
  pageSize?: number;
  onCloseQueue?: (info: { lastViewedId: string; targetPage: number }) => void;
}

export function ReviewQueueProvider({
  children,
  pageSize = 50,
  onCloseQueue,
}: ReviewQueueProviderProps) {
  const [searchParams, setSearchParams] = useSearchParams();
  const queryClient = useQueryClient();

  const [queue, setQueue] = useState<ReviewQueue | null>(null);
  const [isOpen, setIsOpen] = useState(false);
  const isClosingRef = React.useRef(false);

  const clientParam = searchParams.get('client');
  const periodParam = searchParams.get('period') || '2026-10';

  // Sincronizar navegación con prefetch de anterior y siguiente (Criterio 7)
  const prefetchNeighbors = useCallback(
    (q: ReviewQueue) => {
      const prevId = q.index > 0 ? q.ids[q.index - 1] : null;
      const nextId = q.index < q.ids.length - 1 ? q.ids[q.index + 1] : null;

      if (prevId) {
        queryClient.prefetchQuery({
          queryKey: ['closing-detail', prevId],
          queryFn: () => getClosingDetail(prevId),
          staleTime: 60_000,
        });
      }
      if (nextId) {
        queryClient.prefetchQuery({
          queryKey: ['closing-detail', nextId],
          queryFn: () => getClosingDetail(nextId),
          staleTime: 60_000,
        });
      }
    },
    [queryClient]
  );

  // Abrir la cola
  const openQueue = useCallback(
    async (options: OpenQueueOptions) => {
      let ids = options.ids || [];
      const label = options.label || 'Todos los clientes';

      // Si no vienen IDs (modo row o chip), consultamos al backend
      if (options.mode !== 'selection' && ids.length === 0) {
        try {
          const res = await getClosingIds(
            options.period || periodParam,
            options.filterParams || {}
          );
          ids = res.ids;
        } catch {
          // Si falla la consulta de IDs pero tenemos startId, creamos cola con startId
          if (options.startId) {
            ids = [options.startId];
          }
        }
      }

      const newQ = createQueue(ids, options.startId, label);
      setQueue(newQ);
      setIsOpen(true);

      const activeId = current(newQ);
      if (activeId) {
        // Actualizar URL con client= (Criterio 6) y filtro de chip si aplica
        setSearchParams(
          (prevParams) => {
            const nextParams = new URLSearchParams(prevParams);
            if (options.mode === 'chip') {
              const sev =
                options.filterParams instanceof URLSearchParams
                  ? options.filterParams.get('severity')
                  : (options.filterParams as Record<string, unknown> | undefined)?.severity;
              if (sev) {
                nextParams.set('severity', String(sev));
              }
            }
            nextParams.set('client', activeId);
            return nextParams;
          },
          { replace: true }
        );
        prefetchNeighbors(newQ);
      }
    },
    [periodParam, prefetchNeighbors, setSearchParams]
  );

  // Cerrar la cola (Criterio 8)
  const closeQueue = useCallback(() => {
    isClosingRef.current = true;
    const lastId = queue ? current(queue) : undefined;
    const targetPage = queue ? pageOf(queue, pageSize) : 1;

    setIsOpen(false);
    setQueue(null);

    // Remover client de URL inmediatamente
    setSearchParams(
      (prevParams) => {
        const nextParams = new URLSearchParams(prevParams);
        nextParams.delete('client');
        return nextParams;
      },
      { replace: true }
    );

    if (lastId && onCloseQueue) {
      onCloseQueue({ lastViewedId: lastId, targetPage });
    }

    // Resetear flag de cierre en siguiente tick
    setTimeout(() => {
      isClosingRef.current = false;
    }, 0);

    return { lastViewedId: lastId, targetPage };
  }, [queue, pageSize, onCloseQueue, setSearchParams]);

  // Avanzar al siguiente
  const goNext = useCallback(() => {
    setQueue((prevQ) => {
      if (!prevQ || !canNext(prevQ)) return prevQ;
      const nextQ = next(prevQ);
      const nextId = current(nextQ);
      if (nextId) {
        setSearchParams(
          (prevParams) => {
            const nextParams = new URLSearchParams(prevParams);
            nextParams.set('client', nextId);
            return nextParams;
          },
          { replace: true }
        );
        prefetchNeighbors(nextQ);
      }
      return nextQ;
    });
  }, [prefetchNeighbors, setSearchParams]);

  // Retroceder al anterior
  const goPrev = useCallback(() => {
    setQueue((prevQ) => {
      if (!prevQ || !canPrev(prevQ)) return prevQ;
      const prevQueue = prev(prevQ);
      const prevId = current(prevQueue);
      if (prevId) {
        setSearchParams(
          (prevParams) => {
            const nextParams = new URLSearchParams(prevParams);
            nextParams.set('client', prevId);
            return nextParams;
          },
          { replace: true }
        );
        prefetchNeighbors(prevQueue);
      }
      return prevQueue;
    });
  }, [prefetchNeighbors, setSearchParams]);

  // Marcar como revisado
  const markCurrentReviewed = useCallback(() => {
    setQueue((prevQ) => (prevQ ? markReviewed(prevQ) : null));
  }, []);

  // Criterio 6: Recargar la página con client= reabre el modal reconstruyendo la cola
  useEffect(() => {
    if (clientParam && !isOpen && !queue && !isClosingRef.current) {
      openQueue({
        mode: 'row',
        startId: clientParam,
        period: periodParam,
        label: 'Filtro actual',
      });
    }
  }, [clientParam, isOpen, queue, openQueue, periodParam]);

  const currentClosingId = queue ? current(queue) : undefined;
  const canGoNext = queue ? canNext(queue) : false;
  const canGoPrev = queue ? canPrev(queue) : false;
  const currentProgress = useMemo(
    () => (queue ? progress(queue) : { reviewedCount: 0, total: 0 }),
    [queue]
  );
  const isOutsideFilter = Boolean(queue?.isOutsideFilter);

  const value = useMemo(
    () => ({
      queue,
      isOpen,
      currentClosingId,
      canGoNext,
      canGoPrev,
      progress: currentProgress,
      isOutsideFilter,
      openQueue,
      closeQueue,
      goNext,
      goPrev,
      markCurrentReviewed,
    }),
    [
      queue,
      isOpen,
      currentClosingId,
      canGoNext,
      canGoPrev,
      currentProgress,
      isOutsideFilter,
      openQueue,
      closeQueue,
      goNext,
      goPrev,
      markCurrentReviewed,
    ]
  );

  return (
    <ReviewQueueContext.Provider value={value}>
      {children}
    </ReviewQueueContext.Provider>
  );
}

export function useReviewQueue(): ReviewQueueContextValue {
  const ctx = useContext(ReviewQueueContext);
  if (!ctx) {
    throw new Error('useReviewQueue debe ser usado dentro de ReviewQueueProvider');
  }
  return ctx;
}
