import { CheckCircle2, Eye, X } from 'lucide-react';
import { Tooltip } from '../../shared/ui/Tooltip';

export interface SelectionBarProps {
  selectedCount: number;
  greenCount: number;
  onApproveGreens?: () => void;
  onReviewSelected?: () => void;
  onClearSelection: () => void;
  className?: string;
}

export function SelectionBar({
  selectedCount,
  greenCount,
  onApproveGreens,
  onReviewSelected,
  onClearSelection,
  className = '',
}: SelectionBarProps) {
  if (selectedCount === 0) return null;

  const approveButton = (
    <button
      type="button"
      onClick={onApproveGreens}
      disabled={greenCount === 0}
      className="inline-flex h-8 items-center gap-1.5 rounded-control bg-app-primary px-3 text-xs font-semibold text-white shadow-sm hover:bg-app-primary-hover disabled:cursor-not-allowed disabled:opacity-50 focus:outline-none focus:ring-2 focus:ring-app-focus"
    >
      <CheckCircle2 className="h-3.5 w-3.5" aria-hidden="true" />
      <span>Aprobar verdes ({greenCount})</span>
    </button>
  );

  return (
    <div
      role="region"
      aria-label="Acciones para clientes seleccionados"
      className={`fixed bottom-6 left-1/2 z-40 -translate-x-1/2 flex items-center gap-4 rounded-full border border-app-border bg-white px-5 py-2.5 shadow-xl transition-all duration-200 animate-in fade-in-0 slide-in-from-bottom-4 ${className}`}
    >
      <div className="flex items-center gap-2 border-r border-app-border pr-3">
        <span className="text-xs font-semibold text-app-text">
          {selectedCount} {selectedCount === 1 ? 'seleccionado' : 'seleccionados'}
        </span>
        <button
          type="button"
          onClick={onClearSelection}
          aria-label="Deseleccionar todos"
          className="rounded-full p-1 text-app-text-muted hover:bg-app-surface-muted hover:text-app-text focus:outline-none focus:ring-1 focus:ring-app-focus"
        >
          <X className="h-3.5 w-3.5" />
        </button>
      </div>

      <div className="flex items-center gap-2.5">
        {greenCount === 0 ? (
          <Tooltip content="Ninguno de los clientes seleccionados es verde o ya están aprobados">
            <span>{approveButton}</span>
          </Tooltip>
        ) : (
          approveButton
        )}

        <button
          type="button"
          onClick={onReviewSelected}
          className="inline-flex h-8 items-center gap-1.5 rounded-control border border-app-border bg-white px-3 text-xs font-medium text-app-text hover:bg-app-surface-muted focus:outline-none focus:ring-2 focus:ring-app-focus"
        >
          <Eye className="h-3.5 w-3.5 text-app-text-muted" aria-hidden="true" />
          <span>Revisar seleccionados</span>
        </button>
      </div>
    </div>
  );
}

