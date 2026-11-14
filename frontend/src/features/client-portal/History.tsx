import { useState } from 'react';
import { ChevronDown, ChevronUp, CheckCircle2 } from 'lucide-react';
import { ClientClosing } from './schemas';

interface Props {
  closings: ClientClosing[];
}

function formatCurrency(amountCents: number, currency: string): string {
  const num = amountCents / 100;
  const formatted = num.toLocaleString('es-AR', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  });
  return currency === 'USD' ? `US$ ${formatted}` : `$ ${formatted}`;
}

function formatDateDM(isoStr: string): string {
  try {
    const d = new Date(isoStr);
    return d.toLocaleDateString('es-AR', {
      day: '2-digit',
      month: '2-digit',
    });
  } catch {
    return isoStr;
  }
}

const MONTH_NAMES: Record<string, string> = {
  '01': 'Enero',
  '02': 'Febrero',
  '03': 'Marzo',
  '04': 'Abril',
  '05': 'Mayo',
  '06': 'Junio',
  '07': 'Julio',
  '08': 'Agosto',
  '09': 'Septiembre',
  '10': 'Octubre',
  '11': 'Noviembre',
  '12': 'Diciembre',
};

function formatPeriodLabel(periodStr: string): string {
  const parts = periodStr.split('-');
  const month = parts[1];
  const year = parts[0];
  if (parts.length === 2 && month && MONTH_NAMES[month]) {
    return `${MONTH_NAMES[month]} ${year}`;
  }
  return periodStr;
}

export function History({ closings }: Props) {
  const [expandedId, setExpandedId] = useState<string | null>(null);

  // Períodos anteriores que cuentan con snapshot publicado
  const historyClosings = closings.filter((c) => c.published != null);

  if (historyClosings.length === 0) {
    return (
      <div className="rounded-card border border-app-border bg-app-surface p-6 text-center text-sm text-app-text-muted">
        No hay períodos anteriores cerrados todavía.
      </div>
    );
  }

  const toggleExpand = (id: string) => {
    setExpandedId((prev) => (prev === id ? null : id));
  };

  return (
    <div className="rounded-card border border-app-border bg-app-surface shadow-sm overflow-hidden">
      <div className="px-5 py-4 border-b border-app-border">
        <h3 className="text-sm font-semibold text-app-text">
          Historial de cierres anteriores
        </h3>
      </div>

      <div className="divide-y divide-app-border">
        {historyClosings.map((c) => {
          const pub = c.published!;
          const isExpanded = expandedId === c.closingId;
          const totalFormatted = formatCurrency(pub.figures.total.amount, pub.figures.total.currency);

          return (
            <div key={c.closingId} className="p-4 transition-colors hover:bg-app-surface-muted/30">
              <div
                className="flex items-center justify-between cursor-pointer select-none"
                onClick={() => toggleExpand(c.closingId)}
                role="button"
                tabIndex={0}
                onKeyDown={(e) => {
                  if (e.key === 'Enter' || e.key === ' ') {
                    e.preventDefault();
                    toggleExpand(c.closingId);
                  }
                }}
              >
                <div>
                  <h4 className="text-sm font-semibold text-app-text">
                    {formatPeriodLabel(c.period)}
                  </h4>
                  <div className="flex items-center gap-2 mt-1">
                    <span className="text-xs text-app-text-muted">
                      Total:{' '}
                      <strong className="text-app-text font-semibold tabular-nums">
                        {totalFormatted}
                      </strong>
                    </span>
                    {c.paidAt && (
                      <span className="inline-flex items-center gap-1 text-[11px] font-medium text-status-paid-text">
                        <CheckCircle2 className="w-3 h-3" />
                        Pagado el {formatDateDM(c.paidAt)}
                      </span>
                    )}
                  </div>
                </div>

                <div className="flex items-center gap-2">
                  <span className="text-xs text-app-text-muted">
                    {isExpanded ? 'Ocultar' : 'Ver detalle'}
                  </span>
                  {isExpanded ? (
                    <ChevronUp className="w-4 h-4 text-app-text-muted" />
                  ) : (
                    <ChevronDown className="w-4 h-4 text-app-text-muted" />
                  )}
                </div>
              </div>

              {/* Detalle expandido */}
              {isExpanded && (
                <div className="mt-4 pt-3 border-t border-app-border space-y-3 text-xs">
                  <p className="text-app-text-muted whitespace-pre-line leading-relaxed">
                    {pub.body}
                  </p>

                  <div className="grid grid-cols-2 gap-3 pt-1 text-app-text-muted">
                    <div>
                      <span>Facturación USD:</span>{' '}
                      <strong className="text-app-text font-medium tabular-nums">
                        {formatCurrency(pub.figures.billedUSD.amount, pub.figures.billedUSD.currency)}
                      </strong>
                    </div>
                    <div>
                      <span>Facturación ARS:</span>{' '}
                      <strong className="text-app-text font-medium tabular-nums">
                        {formatCurrency(pub.figures.billedARS.amount, pub.figures.billedARS.currency)}
                      </strong>
                    </div>
                  </div>

                  <table className="w-full text-xs mt-2 border border-app-border rounded-control overflow-hidden">
                    <thead className="bg-app-surface-muted text-app-text-muted">
                      <tr>
                        <th className="py-1.5 px-3 text-left font-medium">Concepto</th>
                        <th className="py-1.5 px-3 text-left font-medium">Vencimiento</th>
                        <th className="py-1.5 px-3 text-right font-medium">Importe</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-app-border">
                      {pub.figures.items.map((it, idx) => (
                        <tr key={idx}>
                          <td className="py-1.5 px-3 text-app-text">{it.concept}</td>
                          <td className="py-1.5 px-3 text-app-text-muted tabular-nums">
                            {formatDateDM(it.dueDate)}
                          </td>
                          <td className="py-1.5 px-3 text-right text-app-text tabular-nums font-medium">
                            {formatCurrency(it.amount.amount, it.amount.currency)}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}
