import { useState } from 'react';
import { ClientClosing } from './schemas';
import { markClientClosingPaid } from './api';
import { ApiError } from '../../shared/http/http';

interface ValidatedSummaryCardProps {
  closing: ClientClosing;
  clientId: string;
  onPaidSuccess?: () => void;
}

function formatCents(cents: number, currency: string): string {
  const symbol = currency === 'USD' ? 'US$' : '$';
  const val = (cents / 100).toLocaleString('es-AR', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  });
  return `${symbol} ${val}`;
}

function formatDateDM(iso: string): string {
  if (!iso) return '';
  const d = new Date(iso);
  if (isNaN(d.getTime())) return iso;
  const day = String(d.getDate()).padStart(2, '0');
  const month = String(d.getMonth() + 1).padStart(2, '0');
  return `${day}/${month}`;
}

function formatDateDMY(iso: string): string {
  if (!iso) return '';
  const d = new Date(iso);
  if (isNaN(d.getTime())) return iso;
  const day = String(d.getDate()).padStart(2, '0');
  const month = String(d.getMonth() + 1).padStart(2, '0');
  const year = d.getFullYear();
  return `${day}/${month}/${year}`;
}

export function ValidatedSummaryCard({
  closing,
  clientId,
  onPaidSuccess,
}: ValidatedSummaryCardProps) {
  const pub = closing.published;
  const [paying, setPaying] = useState(false);
  const [payError, setPayError] = useState<string | null>(null);
  const [localPaidAt, setLocalPaidAt] = useState<string | null>(closing.paidAt ?? null);

  if (!pub) return null;

  const handleMarkPaid = async () => {
    setPaying(true);
    setPayError(null);
    try {
      const res = await markClientClosingPaid(clientId, closing.closingId);
      setLocalPaidAt(res.paidAt);
      onPaidSuccess?.();
    } catch (err: unknown) {
      if (err instanceof ApiError) {
        setPayError(err.message || 'Error al registrar el pago');
      } else {
        setPayError('Error al registrar el pago');
      }
    } finally {
      setPaying(false);
    }
  };

  const isMono = pub.regime === 'MONOTRIBUTO';
  const pct = pub.pctOfCategoryCapBp != null ? pub.pctOfCategoryCapBp / 100 : null;
  const pctRounded = pct != null ? Math.round(pct) : null;

  let progressColor = '#10B981'; // verde < 80%
  if (pct != null && pct >= 100) {
    progressColor = '#EF4444'; // rojo >= 100%
  } else if (pct != null && pct >= 80) {
    progressColor = '#EA580C'; // naranja 80-99%
  }

  const items = pub.figures?.items ?? [];
  const totalCents = pub.figures?.total?.amount ?? 0;
  const totalCurr = pub.figures?.total?.currency ?? 'ARS';
  const effectivePaidAt = localPaidAt || closing.paidAt;

  return (
    <article
      aria-label={`Resumen validado de ${closing.period}`}
      className="rounded-2xl border border-neutral-200/90 bg-white p-4 shadow-sm space-y-4 text-neutral-900"
    >
      {/* Facturación del período */}
      <div className="space-y-1">
        <span className="text-[11px] font-semibold tracking-wider text-neutral-500 uppercase">
          Facturaste en {closing.period === '2026-10' ? 'octubre' : 'septiembre'}
        </span>
        <div className="flex items-baseline gap-2">
          <span className="text-2xl font-extrabold tracking-tight text-neutral-950">
            {formatCents(pub.figures?.billedUSD?.amount ?? 0, 'USD')}
          </span>
          <span className="text-xs font-medium text-neutral-500">
            ({formatCents(pub.figures?.billedARS?.amount ?? 0, 'ARS')})
          </span>
        </div>
        <p className="text-[11px] text-neutral-400">
          Pesificado con el tipo de cambio de cada factura
        </p>
      </div>

      {/* Categoría / Régimen */}
      <div className="rounded-xl bg-[#F9F9F9] p-3 space-y-2 border border-neutral-100">
        <span className="text-[11px] font-semibold text-neutral-500 uppercase">
          Tu situación impositiva
        </span>
        {isMono ? (
          <div className="space-y-1.5">
            <div className="flex items-center justify-between text-xs font-semibold text-neutral-800">
              <span>Categoría {pub.category ?? '-'}</span>
              {pctRounded != null && <span>{pctRounded}% del tope anual</span>}
            </div>
            {pct != null && (
              <div
                role="progressbar"
                aria-valuenow={pctRounded ?? 0}
                aria-valuemin={0}
                aria-valuemax={100}
                aria-label={`Porcentaje de tope consumido: ${pctRounded}%`}
                className="h-2 w-full overflow-hidden rounded-full bg-neutral-200"
              >
                <div
                  className="h-full rounded-full transition-all duration-300"
                  style={{
                    width: `${Math.min(pct, 100)}%`,
                    backgroundColor: progressColor,
                  }}
                />
              </div>
            )}
          </div>
        ) : (
          <div className="text-xs font-semibold text-neutral-800">
            Responsable Inscripto
          </div>
        )}
      </div>

      {/* Qué tenés que pagar */}
      <div className="space-y-2">
        <span className="text-[11px] font-semibold text-neutral-500 uppercase">
          Qué tenés que pagar este mes
        </span>

        {items.length === 0 ? (
          <p className="text-xs text-neutral-500">No hay pagos pendientes para este período.</p>
        ) : (
          <div className="space-y-2 divide-y divide-neutral-100">
            {items.map((it, idx) => (
              <div key={idx} className="flex items-center justify-between pt-2 first:pt-0">
                <div>
                  <p className="text-xs font-semibold text-neutral-800">{it.concept}</p>
                  <p className="text-[11px] text-neutral-400">
                    Vence el {formatDateDM(it.dueDate)}
                  </p>
                </div>
                <span className="text-xs font-bold text-neutral-900 tabular-nums">
                  {formatCents(it.amount.amount, it.amount.currency)}
                </span>
              </div>
            ))}
          </div>
        )}

        <div className="flex items-center justify-between border-t border-neutral-200 pt-2 font-bold text-neutral-900">
          <span className="text-xs uppercase">Total</span>
          <span className="text-sm tabular-nums">
            {formatCents(totalCents, totalCurr)}
          </span>
        </div>
      </div>

      {/* Mensaje de la contadora */}
      {pub.body && (
        <div className="rounded-xl border border-neutral-100 bg-[#FAFAFA] p-3 text-xs leading-relaxed text-neutral-700">
          <span className="block mb-1 text-[10px] font-bold uppercase tracking-wider text-neutral-400">
            Nota de la contadora
          </span>
          {pub.body}
        </div>
      )}

      {/* Badge de validación */}
      <div
        data-testid="validated-badge"
        className="flex items-center gap-2 rounded-xl border border-emerald-200 bg-emerald-50/80 px-3 py-2 text-[11px] font-medium text-emerald-800"
      >
        <span className="text-emerald-600 font-bold" aria-hidden="true">✓</span>
        <span>
          Validado por {pub.validatedBy?.name ?? 'C.P. Mariana Quiroga'} · {pub.validatedBy?.license ?? 'Matrícula de ejemplo'} · {formatDateDMY(pub.publishedAt)}
        </span>
      </div>

      {/* Botón Ya pagué o Badge de pagado */}
      <div className="pt-1">
        {effectivePaidAt ? (
          <div
            data-testid="paid-badge"
            className="flex h-10 w-full items-center justify-center rounded-xl bg-neutral-100 text-xs font-bold text-neutral-700"
          >
            ✓ Pagado el {formatDateDM(effectivePaidAt)}
          </div>
        ) : (
          <div className="space-y-1">
            {payError && (
              <p className="text-center text-[11px] text-red-600">{payError}</p>
            )}
            <button
              type="button"
              disabled={paying}
              onClick={handleMarkPaid}
              className="flex h-10 w-full items-center justify-center rounded-xl bg-neutral-900 px-4 text-xs font-semibold text-white shadow transition-colors hover:bg-neutral-800 disabled:opacity-50"
            >
              {paying ? 'Registrando pago...' : 'Ya pagué'}
            </button>
          </div>
        )}
      </div>
    </article>
  );
}
