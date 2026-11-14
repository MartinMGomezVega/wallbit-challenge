import { CheckCircle2 } from 'lucide-react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { markClosingPaid } from './api';
import { ClientClosing } from './schemas';
import { ComingSoonButton } from '../../shared/ui/ComingSoonButton';

interface Props {
  closing: ClientClosing;
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

export function ReadySummary({ closing }: Props) {
  const queryClient = useQueryClient();
  const pub = closing.published;

  const mutation = useMutation({
    mutationFn: () => markClosingPaid(closing.closingId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['my-closings'] });
    },
  });

  if (!pub) return null;

  const totalFormatted = formatCurrency(pub.figures.total.amount, pub.figures.total.currency);
  const earliestDue = pub.figures.items.length > 0 && pub.figures.items[0] ? formatDateDM(pub.figures.items[0].dueDate) : '';

  return (
    <div className="space-y-5">
      {/* Cifra destacada en text-2xl (UI guidelines §3.2) */}
      <div className="rounded-card border border-app-border bg-app-surface p-5 shadow-sm space-y-2">
        <div className="flex items-center justify-between flex-wrap gap-2">
          <span className="text-xs font-medium uppercase tracking-wider text-app-text-muted">
            Total a pagar
          </span>
          {closing.paidAt ? (
            <span
              data-testid="paid-badge"
              className="inline-flex items-center gap-1 rounded-control border border-status-paid-border bg-status-notified-bg px-2.5 py-1 text-xs font-semibold text-status-paid-text"
            >
              <CheckCircle2 className="w-3.5 h-3.5" aria-hidden="true" />
              Pagado el {formatDateDM(closing.paidAt)}
            </span>
          ) : (
            <span className="text-xs font-medium text-app-text-muted">
              Vence el {earliestDue}
            </span>
          )}
        </div>

        <div className="text-2xl font-bold tracking-tight text-app-text tabular-nums">
          Tenés que pagar {totalFormatted}
          {earliestDue && !closing.paidAt && (
            <span className="text-base font-normal text-app-text-muted ml-2">
              · vence el {earliestDue}
            </span>
          )}
        </div>
      </div>

      {/* Mensaje del contador */}
      <div className="rounded-card border border-app-border bg-app-surface p-5 shadow-sm space-y-3">
        <h4 className="text-base font-semibold text-app-text">{pub.title}</h4>
        <p className="text-sm text-app-text-muted whitespace-pre-line leading-relaxed">
          {pub.body}
        </p>

        {/* Resumen de facturación del período */}
        <div className="pt-2 border-t border-app-border grid grid-cols-2 gap-4 text-xs">
          <div>
            <span className="text-app-text-muted">Facturación en dólares:</span>
            <p className="font-semibold text-app-text tabular-nums text-sm">
              {formatCurrency(pub.figures.billedUSD.amount, pub.figures.billedUSD.currency)}
            </p>
          </div>
          <div>
            <span className="text-app-text-muted">Facturación en pesos:</span>
            <p className="font-semibold text-app-text tabular-nums text-sm">
              {formatCurrency(pub.figures.billedARS.amount, pub.figures.billedARS.currency)}
            </p>
          </div>
        </div>
      </div>

      {/* Tabla de conceptos a pagar */}
      <div className="rounded-card border border-app-border bg-app-surface overflow-hidden shadow-sm">
        <div className="px-4 py-3 border-b border-app-border bg-app-surface-muted">
          <h4 className="text-xs font-semibold uppercase tracking-wider text-app-text-muted">
            Detalle de conceptos a pagar
          </h4>
        </div>
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-app-border text-xs text-app-text-muted text-left">
              <th className="py-2.5 px-4 font-medium">Concepto</th>
              <th className="py-2.5 px-4 font-medium">Vencimiento</th>
              <th className="py-2.5 px-4 font-medium text-right">Importe</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-app-border">
            {pub.figures.items.map((item, idx) => (
              <tr key={idx} className="hover:bg-app-surface-muted/50 transition-colors">
                <td className="py-2.5 px-4 font-medium text-app-text">{item.concept}</td>
                <td className="py-2.5 px-4 text-app-text-muted tabular-nums">
                  {formatDateDM(item.dueDate)}
                </td>
                <td className="py-2.5 px-4 text-right font-medium text-app-text tabular-nums">
                  {formatCurrency(item.amount.amount, item.amount.currency)}
                </td>
              </tr>
            ))}
          </tbody>
          <tfoot>
            <tr className="border-t border-app-border bg-app-surface-muted/30 font-semibold">
              <td className="py-3 px-4 text-app-text" colSpan={2}>
                Total
              </td>
              <td className="py-3 px-4 text-right text-app-text tabular-nums">
                {totalFormatted}
              </td>
            </tr>
          </tfoot>
        </table>
      </div>

      {/* Acciones: Pagar en ARCA (Próximamente) y Ya pagué */}
      <div className="flex items-center justify-end gap-3 pt-2">
        <ComingSoonButton size="md">Pagar en ARCA</ComingSoonButton>

        {!closing.paidAt ? (
          <button
            type="button"
            disabled={mutation.isPending}
            onClick={() => mutation.mutate()}
            className="h-10 px-5 rounded-control bg-app-primary text-white text-sm font-semibold hover:bg-app-primary-hover disabled:opacity-50 transition-colors shadow-sm"
          >
            {mutation.isPending ? 'Guardando...' : 'Ya pagué'}
          </button>
        ) : (
          <span className="text-xs text-app-text-muted font-medium">
            Pago registrado
          </span>
        )}
      </div>
    </div>
  );
}
