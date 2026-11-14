import { ClosingDetailDTO } from '../schemas';
import { SeverityIndicator } from '../../../shared/ui/SeverityIndicator';
import { CheckCircle2 } from 'lucide-react';

export interface SignalsSectionProps {
  detail: ClosingDetailDTO;
}

export function SignalsSection({ detail }: SignalsSectionProps) {
  const { signals } = detail.closing;

  if (!signals || signals.length === 0) {
    return (
      <div className="rounded-card border border-severity-green-border bg-severity-green-bg/50 p-6 text-center">
        <CheckCircle2 className="mx-auto h-8 w-8 text-severity-green-icon mb-2" />
        <h4 className="text-sm font-semibold text-severity-green-text">
          Sin señales de alerta
        </h4>
        <p className="text-xs text-severity-green-text/80 mt-1">
          Todos los parámetros tributarios y de facturación se encuentran dentro de los límites esperados.
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-3">
      {signals.map((sig, idx) => {
        const isRed = sig.severity === 'RED';
        return (
          <div
            key={`${sig.code}-${idx}`}
            className={`rounded-card border p-4 transition-all ${
              isRed
                ? 'border-severity-red-border bg-severity-red-bg/30'
                : 'border-severity-yellow-border bg-severity-yellow-bg/30'
            }`}
          >
            <div className="flex items-start gap-3">
              <div className="mt-0.5">
                <SeverityIndicator severity={sig.severity} />
              </div>

              <div className="flex-1 min-w-0">
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <h4 className="text-xs font-bold text-app-text">
                    {sig.code}
                  </h4>
                  <span
                    className={`text-[10px] font-semibold px-2 py-0.5 rounded-full ${
                      isRed
                        ? 'bg-severity-red-border text-severity-red-text'
                        : 'bg-severity-yellow-border text-severity-yellow-text'
                    }`}
                  >
                    {isRed ? 'Requiere acción inmediata' : 'Revisión preventiva'}
                  </span>
                </div>

                <p className="text-xs text-app-text-muted mt-1 leading-relaxed">
                  {sig.message || detail.closing.description}
                </p>

                {sig.data && Object.keys(sig.data).length > 0 && (
                  <div className="mt-2.5 flex flex-wrap gap-2 pt-2 border-t border-app-border/40">
                    {Object.entries(sig.data).map(([key, val]) => (
                      <div
                        key={key}
                        className="rounded-control bg-white/80 border border-app-border px-2 py-1 text-[11px]"
                      >
                        <span className="text-app-text-muted font-medium">{key}: </span>
                        <span className="font-semibold text-app-text">
                          {typeof val === 'object' ? JSON.stringify(val) : String(val)}
                        </span>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </div>
          </div>
        );
      })}
    </div>
  );
}
