import { useState, useRef, useEffect } from 'react';
import { ChevronDown, Check } from 'lucide-react';
import { SignalCatalogItem } from './schemas';

export interface SignalMultiSelectProps {
  catalog: SignalCatalogItem[];
  bySignalCounts?: Record<string, number>;
  selectedSignals: string[];
  onChange: (signals: string[]) => void;
  className?: string;
}

export function SignalMultiSelect({
  catalog,
  bySignalCounts = {},
  selectedSignals,
  onChange,
  className = '',
}: SignalMultiSelectProps) {
  const [isOpen, setIsOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  // Cerrar al hacer click afuera o presionar Esc
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    }
    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === 'Escape') {
        setIsOpen(false);
      }
    }

    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside);
      document.addEventListener('keydown', handleKeyDown);
    }
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
      document.removeEventListener('keydown', handleKeyDown);
    };
  }, [isOpen]);

  const toggleSignal = (code: string) => {
    if (selectedSignals.includes(code)) {
      onChange(selectedSignals.filter((c) => c !== code));
    } else {
      onChange([...selectedSignals, code]);
    }
  };

  const redSignals = catalog.filter((s) => s.severity === 'RED');
  const yellowSignals = catalog.filter((s) => s.severity === 'YELLOW');

  const selectedCount = selectedSignals.length;

  return (
    <div ref={containerRef} className={`relative inline-block ${className}`}>
      <button
        type="button"
        onClick={() => setIsOpen((prev) => !prev)}
        aria-expanded={isOpen}
        aria-haspopup="listbox"
        className={`inline-flex h-8 items-center gap-1.5 rounded-control border px-3 text-xs font-medium transition-colors focus:outline-none focus:ring-2 focus:ring-app-focus ${
          selectedCount > 0
            ? 'border-app-primary bg-app-primary-soft text-app-primary'
            : 'border-app-border bg-white text-app-text hover:bg-app-surface-muted'
        }`}
      >
        <span>
          Situación {selectedCount > 0 ? `(${selectedCount})` : ''}
        </span>
        <ChevronDown className="h-3.5 w-3.5 opacity-70" aria-hidden="true" />
      </button>

      {isOpen && (
        <div
          role="listbox"
          aria-multiselectable="true"
          className="absolute left-0 z-30 mt-1 w-72 max-h-80 overflow-y-auto rounded-control border border-app-border bg-white p-2 shadow-xl focus:outline-none"
        >
          <div className="flex items-center justify-between pb-1.5 border-b border-app-border px-1">
            <span className="text-[11px] font-semibold uppercase tracking-wider text-app-text-muted">
              Filtrar por situación
            </span>
            {selectedCount > 0 && (
              <button
                type="button"
                onClick={() => onChange([])}
                className="text-[11px] font-medium text-app-primary hover:underline"
              >
                Limpiar
              </button>
            )}
          </div>

          {/* Grupo Rojo */}
          {redSignals.length > 0 && (
            <div className="mt-2">
              <div className="px-2 py-0.5 text-[11px] font-semibold text-severity-red-text">
                Requiere acción (Rojo)
              </div>
              <div className="space-y-0.5">
                {redSignals.map((item) => {
                  const isChecked = selectedSignals.includes(item.code);
                  const count = bySignalCounts[item.code] ?? 0;
                  return (
                    <label
                      key={item.code}
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
                          onChange={() => toggleSignal(item.code)}
                          className="sr-only"
                        />
                        <span>{item.label}</span>
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

          {/* Grupo Amarillo */}
          {yellowSignals.length > 0 && (
            <div className="mt-3">
              <div className="px-2 py-0.5 text-[11px] font-semibold text-severity-yellow-text">
                Revisar (Amarillo)
              </div>
              <div className="space-y-0.5">
                {yellowSignals.map((item) => {
                  const isChecked = selectedSignals.includes(item.code);
                  const count = bySignalCounts[item.code] ?? 0;
                  return (
                    <label
                      key={item.code}
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
                          onChange={() => toggleSignal(item.code)}
                          className="sr-only"
                        />
                        <span>{item.label}</span>
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
      )}
    </div>
  );
}

