import { ChevronLeft, ChevronRight } from 'lucide-react';

export interface PaginationProps {
  page: number;
  pageSize: number;
  total: number;
  onPageChange: (page: number) => void;
  onPageSizeChange?: (pageSize: number) => void;
  pageSizeOptions?: number[];
  className?: string;
}

export function Pagination({
  page,
  pageSize,
  total,
  onPageChange,
  onPageSizeChange,
  pageSizeOptions = [50, 100, 200],
  className = '',
}: PaginationProps) {
  const totalPages = Math.max(1, Math.ceil(total / pageSize));
  const currentPage = Math.min(Math.max(1, page), totalPages);

  const from = total === 0 ? 0 : (currentPage - 1) * pageSize + 1;
  const to = Math.min(currentPage * pageSize, total);

  // Generar números de página a mostrar
  const getPageNumbers = () => {
    const delta = 1;
    const range: (number | string)[] = [];
    const rangeWithDots: (number | string)[] = [];

    for (
      let i = Math.max(2, currentPage - delta);
      i <= Math.min(totalPages - 1, currentPage + delta);
      i++
    ) {
      range.push(i);
    }

    if (currentPage - delta > 2) {
      rangeWithDots.push(1, '…');
    } else {
      rangeWithDots.push(1);
    }

    rangeWithDots.push(...range);

    if (currentPage + delta < totalPages - 1) {
      rangeWithDots.push('…', totalPages);
    } else if (totalPages > 1) {
      rangeWithDots.push(totalPages);
    }

    return totalPages <= 1 ? [1] : rangeWithDots;
  };

  return (
    <nav
      role="navigation"
      aria-label="Paginación de resultados"
      className={`flex flex-wrap items-center justify-between gap-4 py-3 text-xs text-app-text-muted ${className}`}
    >
      <div className="font-medium tabular-nums">
        Mostrando <span className="text-app-text">{from}</span>–
        <span className="text-app-text">{to}</span> de{' '}
        <span className="text-app-text">{total}</span>
      </div>

      <div className="flex items-center gap-2">
        <div className="flex items-center space-x-1">
          <button
            type="button"
            onClick={() => onPageChange(currentPage - 1)}
            disabled={currentPage <= 1}
            aria-label="Página anterior"
            className="inline-flex h-8 w-8 items-center justify-center rounded-control border border-app-border bg-app-surface text-app-text hover:bg-app-surface-muted disabled:cursor-not-allowed disabled:opacity-40 focus:outline-none focus:ring-2 focus:ring-app-focus"
          >
            <ChevronLeft className="h-4 w-4" aria-hidden="true" />
          </button>

          {getPageNumbers().map((p, idx) => {
            if (p === '…') {
              return (
                <span
                  key={`dots-${idx}`}
                  className="inline-flex h-8 w-6 items-center justify-center text-app-text-muted"
                >
                  …
                </span>
              );
            }

            const pageNum = p as number;
            const isActive = pageNum === currentPage;

            return (
              <button
                key={pageNum}
                type="button"
                onClick={() => onPageChange(pageNum)}
                aria-label={`Página ${pageNum}`}
                aria-current={isActive ? 'page' : undefined}
                className={`inline-flex h-8 min-w-[2rem] items-center justify-center rounded-control px-2 font-medium tabular-nums focus:outline-none focus:ring-2 focus:ring-app-focus ${
                  isActive
                    ? 'bg-app-primary text-white'
                    : 'border border-app-border bg-app-surface text-app-text hover:bg-app-surface-muted'
                }`}
              >
                {pageNum}
              </button>
            );
          })}

          <button
            type="button"
            onClick={() => onPageChange(currentPage + 1)}
            disabled={currentPage >= totalPages}
            aria-label="Página siguiente"
            className="inline-flex h-8 w-8 items-center justify-center rounded-control border border-app-border bg-app-surface text-app-text hover:bg-app-surface-muted disabled:cursor-not-allowed disabled:opacity-40 focus:outline-none focus:ring-2 focus:ring-app-focus"
          >
            <ChevronRight className="h-4 w-4" aria-hidden="true" />
          </button>
        </div>

        {onPageSizeChange && (
          <div className="flex items-center gap-1.5 pl-2">
            <select
              value={pageSize}
              onChange={(e) => onPageSizeChange(Number(e.target.value))}
              aria-label="Resultados por página"
              className="h-8 rounded-control border border-app-border bg-app-surface px-2 py-0.5 text-xs text-app-text focus:outline-none focus:ring-2 focus:ring-app-focus"
            >
              {pageSizeOptions.map((opt) => (
                <option key={opt} value={opt}>
                  {opt} / pág
                </option>
              ))}
            </select>
          </div>
        )}
      </div>
    </nav>
  );
}

