import { ArrowLeft, ArrowRight, Check, X } from 'lucide-react';
import { Tooltip } from '../../shared/ui/Tooltip';

export interface ReviewHeaderProps {
  currentIndex: number;
  total: number;
  filterLabel: string;
  reviewedCount: number;
  canGoPrev: boolean;
  canGoNext: boolean;
  onPrev: () => void;
  onNext: () => void;
  onClose: () => void;
  className?: string;
}

export function ReviewHeader({
  currentIndex,
  total,
  filterLabel,
  reviewedCount,
  canGoPrev,
  canGoNext,
  onPrev,
  onNext,
  onClose,
  className = '',
}: ReviewHeaderProps) {
  const currentNum = total > 0 ? currentIndex + 1 : 0;

  return (
    <div
      className={`flex items-center justify-between border-b border-app-border bg-app-surface px-6 py-3.5 select-none ${className}`}
    >
      {/* Botón Anterior */}
      <div className="flex items-center gap-2">
        {!canGoPrev ? (
          <Tooltip content="Es el primer cliente del filtro">
            <span className="inline-block">
              <button
                type="button"
                aria-disabled="true"
                disabled
                className="inline-flex h-8 items-center gap-1.5 rounded-control border border-app-border bg-app-surface-muted px-2.5 text-xs font-medium text-app-text-muted opacity-50 cursor-not-allowed"
              >
                <ArrowLeft className="h-3.5 w-3.5" aria-hidden="true" />
                <span>Anterior</span>
              </button>
            </span>
          </Tooltip>
        ) : (
          <button
            type="button"
            onClick={onPrev}
            className="inline-flex h-8 items-center gap-1.5 rounded-control border border-app-border bg-white px-2.5 text-xs font-medium text-app-text hover:bg-app-surface-muted focus:outline-none focus:ring-2 focus:ring-app-focus"
          >
            <ArrowLeft className="h-3.5 w-3.5" aria-hidden="true" />
            <span>Anterior</span>
          </button>
        )}
      </div>

      {/* Centro: Posición, Filtro y Progreso */}
      <div className="flex flex-col items-center gap-1 text-center">
        <div className="flex items-center gap-2">
          <span className="text-xs font-bold text-app-text">
            Cliente {currentNum} de {total}
          </span>
          <span className="text-app-text-muted/60">•</span>
          <span className="text-xs text-app-text-muted truncate max-w-xs md:max-w-md">
            {filterLabel}
          </span>
        </div>

        {/* Mini-indicador de progreso (Criterio 9) */}
        <div className="inline-flex items-center gap-1 rounded-full bg-app-surface-muted px-2 py-0.5 text-[11px] font-medium text-app-text-muted">
          <Check className="h-3 w-3 text-severity-green-icon stroke-[3]" aria-hidden="true" />
          <span>
            {reviewedCount} de {total} revisados
          </span>
        </div>
      </div>

      {/* Botón Siguiente y Cerrar */}
      <div className="flex items-center gap-3">
        {!canGoNext ? (
          <Tooltip content="Es el último cliente del filtro">
            <span className="inline-block">
              <button
                type="button"
                aria-disabled="true"
                disabled
                className="inline-flex h-8 items-center gap-1.5 rounded-control border border-app-border bg-app-surface-muted px-2.5 text-xs font-medium text-app-text-muted opacity-50 cursor-not-allowed"
              >
                <span>Siguiente</span>
                <ArrowRight className="h-3.5 w-3.5" aria-hidden="true" />
              </button>
            </span>
          </Tooltip>
        ) : (
          <button
            type="button"
            onClick={onNext}
            className="inline-flex h-8 items-center gap-1.5 rounded-control border border-app-border bg-white px-2.5 text-xs font-medium text-app-text hover:bg-app-surface-muted focus:outline-none focus:ring-2 focus:ring-app-focus"
          >
            <span>Siguiente</span>
            <ArrowRight className="h-3.5 w-3.5" aria-hidden="true" />
          </button>
        )}

        <div className="h-4 w-px bg-app-border" aria-hidden="true" />

        <button
          type="button"
          onClick={onClose}
          aria-label="Cerrar modal de revisión"
          className="rounded-control p-1.5 text-app-text-muted hover:bg-app-surface-muted hover:text-app-text focus:outline-none focus:ring-2 focus:ring-app-focus"
        >
          <X className="h-4 w-4" aria-hidden="true" />
        </button>
      </div>
    </div>
  );
}
