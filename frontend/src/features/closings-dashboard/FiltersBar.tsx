import { useState, useEffect, useRef } from 'react';
import { Search, X, ChevronDown, Check } from 'lucide-react';
import { DashboardFilters } from './useDashboardParams';
import { ClosingsSummaryResponse, SignalCatalogItem } from './schemas';
import { SignalMultiSelect } from './SignalMultiSelect';

export interface FiltersBarProps {
  filters: DashboardFilters;
  summary?: ClosingsSummaryResponse;
  catalog: SignalCatalogItem[];
  onSetSignal: (signals: string[]) => void;
  onSetRegime: (regime: string) => void;
  onSetStatus: (statuses: string[]) => void;
  onSetQ: (q: string) => void;
  onClearFilters: () => void;
  className?: string;
}

const STATUS_OPTIONS = [
  { value: 'READY_FOR_REVIEW', label: 'Para revisar' },
  { value: 'PENDING_CLIENT_INPUT', label: 'Esperando al cliente' },
  { value: 'APPROVED', label: 'Aprobado' },
  { value: 'NOTIFIED', label: 'Notificado' },
];

export function FiltersBar({
  filters,
  summary,
  catalog,
  onSetSignal,
  onSetRegime,
  onSetStatus,
  onSetQ,
  onClearFilters,
  className = '',
}: FiltersBarProps) {
  // Estado local para input con debounce de 300 ms
  const [searchValue, setSearchValue] = useState(filters.q);
  const [isStatusOpen, setIsStatusOpen] = useState(false);
  const statusRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (statusRef.current && !statusRef.current.contains(event.target as Node)) {
        setIsStatusOpen(false);
      }
    }
    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === 'Escape') {
        setIsStatusOpen(false);
      }
    }
    if (isStatusOpen) {
      document.addEventListener('mousedown', handleClickOutside);
      document.addEventListener('keydown', handleKeyDown);
    }
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
      document.removeEventListener('keydown', handleKeyDown);
    };
  }, [isStatusOpen]);

  useEffect(() => {
    setSearchValue(filters.q);
  }, [filters.q]);

  useEffect(() => {
    const timer = setTimeout(() => {
      if (searchValue !== filters.q) {
        onSetQ(searchValue);
      }
    }, 300);
    return () => clearTimeout(timer);
  }, [searchValue, filters.q, onSetQ]);

  const hasActiveFilters =
    filters.severity.length > 0 ||
    filters.signal.length > 0 ||
    Boolean(filters.regime) ||
    filters.status.length > 0 ||
    Boolean(filters.q);

  const toggleStatus = (st: string) => {
    if (filters.status.includes(st)) {
      onSetStatus(filters.status.filter((s) => s !== st));
    } else {
      onSetStatus([...filters.status, st]);
    }
  };

  const selectedStatusCount = filters.status.length;

  return (
    <div className={`flex flex-wrap items-center justify-between gap-3 ${className}`}>
      <div className="flex flex-wrap items-center gap-2.5">
        {/* Selector de Situación */}
        <SignalMultiSelect
          catalog={catalog}
          bySignalCounts={summary?.bySignal}
          selectedSignals={filters.signal}
          onChange={onSetSignal}
        />

        {/* Selector de Régimen */}
        <div className="relative inline-block">
          <select
            value={filters.regime}
            onChange={(e) => onSetRegime(e.target.value)}
            aria-label="Filtrar por régimen"
            className={`h-8 rounded-control border px-2.5 pr-8 text-xs font-medium appearance-none focus:outline-none focus:ring-2 focus:ring-app-focus ${
              filters.regime
                ? 'border-app-primary bg-app-primary-soft text-app-primary'
                : 'border-app-border bg-white text-app-text hover:bg-app-surface-muted'
            }`}
          >
            <option value="">Régimen: Todos</option>
            <option value="MONOTRIBUTO">Monotributo</option>
            <option value="RI">Responsable Inscripto</option>
          </select>
          <ChevronDown
            className="pointer-events-none absolute right-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-app-text-muted"
            aria-hidden="true"
          />
        </div>

        {/* Selector de Estado */}
        <div ref={statusRef} className="relative inline-block">
          <button
            type="button"
            onClick={() => setIsStatusOpen((prev) => !prev)}
            aria-expanded={isStatusOpen}
            aria-haspopup="listbox"
            className={`inline-flex h-8 items-center gap-1.5 rounded-control border px-3 text-xs font-medium transition-colors focus:outline-none focus:ring-2 focus:ring-app-focus ${
              selectedStatusCount > 0
                ? 'border-app-primary bg-app-primary-soft text-app-primary'
                : 'border-app-border bg-white text-app-text hover:bg-app-surface-muted'
            }`}
          >
            <span>Estado {selectedStatusCount > 0 ? `(${selectedStatusCount})` : ''}</span>
            <ChevronDown className="h-3.5 w-3.5 opacity-70" aria-hidden="true" />
          </button>

          {isStatusOpen && (
            <div
              role="listbox"
              aria-multiselectable="true"
              className="absolute left-0 z-30 mt-1 w-52 rounded-control border border-app-border bg-white p-2 shadow-xl focus:outline-none"
            >
              <div className="flex items-center justify-between pb-1.5 border-b border-app-border px-1">
                <span className="text-[11px] font-semibold uppercase tracking-wider text-app-text-muted">
                  Estado del cierre
                </span>
                {selectedStatusCount > 0 && (
                  <button
                    type="button"
                    onClick={() => onSetStatus([])}
                    className="text-[11px] font-medium text-app-primary hover:underline"
                  >
                    Limpiar
                  </button>
                )}
              </div>
              <div className="mt-1 space-y-0.5">
                {STATUS_OPTIONS.map((opt) => {
                  const isChecked = filters.status.includes(opt.value);
                  const count = summary?.byStatus[opt.value] ?? 0;
                  return (
                    <label
                      key={opt.value}
                      className="flex cursor-pointer items-center justify-between rounded-control px-2 py-1.5 text-xs text-app-text hover:bg-app-surface-muted"
                    >
                      <div className="flex items-center gap-2">
                        <div
                          className={`flex h-4 w-4 items-center justify-center rounded-[3px] border ${
                            isChecked
                              ? 'border-app-primary bg-app-primary text-white'
                              : 'border-app-border bg-white'
                          }`}
                        >
                          {isChecked && <Check className="h-3 w-3 stroke-[3]" />}
                        </div>
                        <input
                          type="checkbox"
                          checked={isChecked}
                          onChange={() => toggleStatus(opt.value)}
                          className="sr-only"
                        />
                        <span>{opt.label}</span>
                      </div>
                      <span className="text-[11px] tabular-nums text-app-text-muted">
                        {count}
                      </span>
                    </label>
                  );
                })}
              </div>
            </div>
          )}
        </div>

        {/* Input de Búsqueda */}
        <div className="relative inline-flex items-center">
          <Search
            className="pointer-events-none absolute left-2.5 h-3.5 w-3.5 text-app-text-muted"
            aria-hidden="true"
          />
          <input
            type="text"
            value={searchValue}
            onChange={(e) => setSearchValue(e.target.value)}
            placeholder="Buscar nombre o CUIT"
            aria-label="Buscar nombre o CUIT"
            className="h-8 w-56 rounded-control border border-app-border bg-white pl-8 pr-7 text-xs text-app-text placeholder:text-app-text-muted focus:outline-none focus:ring-2 focus:ring-app-focus"
          />
          {searchValue && (
            <button
              type="button"
              onClick={() => {
                setSearchValue('');
                onSetQ('');
              }}
              aria-label="Limpiar búsqueda"
              className="absolute right-2 text-app-text-muted hover:text-app-text"
            >
              <X className="h-3.5 w-3.5" />
            </button>
          )}
        </div>

        {/* Limpiar Filtros */}
        {hasActiveFilters && (
          <button
            type="button"
            onClick={onClearFilters}
            className="inline-flex h-8 items-center px-2 text-xs font-medium text-app-text-muted hover:text-app-primary focus:outline-none focus:ring-2 focus:ring-app-focus rounded-control"
          >
            Limpiar filtros
          </button>
        )}
      </div>
    </div>
  );
}
