import { ClosingDetailDTO } from '../schemas';
import { formatMoney, formatDate } from '../../../shared/format';
import { CheckCircle2, AlertCircle, FileText, ArrowRightLeft } from 'lucide-react';

export interface InvoicesSectionProps {
  detail: ClosingDetailDTO;
}

export function InvoicesSection({ detail }: InvoicesSectionProps) {
  const { periodInvoices, reconciliation } = detail;
  const { matchedPairs, unmatchedIncome, unmatchedInvoices } = reconciliation;

  return (
    <div className="space-y-6">
      {/* Facturas emitidas en el período */}
      <div className="rounded-card border border-app-border bg-app-surface overflow-hidden shadow-sm">
        <div className="px-4 py-3 border-b border-app-border flex items-center justify-between">
          <div className="flex items-center gap-2">
            <FileText className="h-4 w-4 text-app-primary" />
            <h3 className="text-xs font-bold uppercase tracking-wider text-app-text">
              Facturas del período ({periodInvoices.length})
            </h3>
          </div>
        </div>

        {periodInvoices.length === 0 ? (
          <div className="p-6 text-center text-xs text-app-text-muted">
            No se emitieron facturas en este período.
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left border-collapse text-xs">
              <thead>
                <tr className="border-b border-app-border bg-app-surface-muted text-[11px] font-semibold uppercase tracking-wider text-app-text-muted">
                  <th className="px-3 py-2">Fecha</th>
                  <th className="px-3 py-2">Comprobante</th>
                  <th className="px-3 py-2">Tipo</th>
                  <th className="px-3 py-2 text-right">Importe original</th>
                  <th className="px-3 py-2 text-right">TC BNA</th>
                  <th className="px-3 py-2 text-right">Pesificado (ARS)</th>
                  <th className="px-3 py-2">CAE</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-app-border">
                {periodInvoices.map((pi) => {
                  const inv = pi.invoice;
                  const formattedNum = `${String(inv.pointOfSale).padStart(4, '0')}-${String(
                    inv.number
                  ).padStart(8, '0')}`;

                  return (
                    <tr key={inv.id} className="hover:bg-app-surface-muted/50">
                      <td className="px-3 py-2 tabular-nums text-app-text">
                        {formatDate(inv.issueDate, 'short')}
                      </td>
                      <td className="px-3 py-2 font-mono font-medium text-app-text">
                        {formattedNum}
                      </td>
                      <td className="px-3 py-2">
                        <span className="inline-block px-1.5 py-0.5 rounded bg-app-surface-muted border border-app-border text-[10px] font-semibold text-app-text">
                          Factura {inv.type}
                        </span>
                      </td>
                      <td className="px-3 py-2 text-right tabular-nums font-semibold text-app-text">
                        {formatMoney(inv.totalAmount)}
                      </td>
                      <td className="px-3 py-2 text-right tabular-nums text-app-text-muted">
                        {pi.fxRateUsed ? `$ ${(pi.fxRateUsed / 100).toFixed(2)}` : '-'}
                      </td>
                      <td className="px-3 py-2 text-right tabular-nums font-medium text-app-text">
                        {formatMoney(pi.amountArs)}
                      </td>
                      <td className="px-3 py-2 text-[11px] font-mono text-app-text-muted">
                        {inv.cae || '-'}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Conciliación automática: Pares conciliados */}
      <div className="rounded-card border border-app-border bg-app-surface p-4 space-y-3">
        <div className="flex items-center gap-2 border-b border-app-border pb-2">
          <ArrowRightLeft className="h-4 w-4 text-severity-green-icon" />
          <h3 className="text-xs font-bold uppercase tracking-wider text-app-text">
            Pares conciliados automáticamente ({matchedPairs.length})
          </h3>
        </div>

        {matchedPairs.length === 0 ? (
          <div className="text-xs text-app-text-muted py-2">
            No se identificaron coincidencias exactas entre cobros y facturas.
          </div>
        ) : (
          <div className="space-y-2">
            {matchedPairs.map((pair, idx) => (
              <div
                key={`pair-${idx}`}
                className="flex flex-wrap items-center justify-between gap-3 rounded-control border border-severity-green-border/50 bg-severity-green-bg/20 p-2.5 text-xs"
              >
                <div className="flex items-center gap-2">
                  <CheckCircle2 className="h-4 w-4 text-severity-green-icon flex-shrink-0" />
                  <div>
                    <span className="font-semibold text-app-text">
                      Cobro {formatMoney(pair.movement.amount)}
                    </span>
                    <span className="text-[11px] text-app-text-muted ml-2">
                      ({formatDate(pair.movement.occurredAt, 'short')} • {pair.movement.counterparty || 'Wallbit'})
                    </span>
                  </div>
                </div>

                <div className="text-right">
                  <span className="font-semibold text-app-text">
                    Factura {pair.invoice.type} {String(pair.invoice.number).padStart(8, '0')}
                  </span>
                  <span className="text-[11px] text-app-text-muted ml-2">
                    ({formatMoney(pair.invoice.totalAmount)})
                  </span>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Discrepancias / Faltantes */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {/* Cobros sin factura */}
        <div className="rounded-card border border-app-border bg-app-surface p-4 space-y-2.5">
          <div className="flex items-center gap-2 border-b border-app-border pb-2">
            <AlertCircle className="h-4 w-4 text-severity-yellow-icon" />
            <h4 className="text-xs font-bold text-app-text">
              Cobros sin factura ({unmatchedIncome.length})
            </h4>
          </div>

          {unmatchedIncome.length === 0 ? (
            <p className="text-xs text-app-text-muted">Todos los cobros de ingreso tienen factura asociada.</p>
          ) : (
            <div className="space-y-2">
              {unmatchedIncome.map((m) => (
                <div
                  key={m.id}
                  className="rounded-control border border-severity-yellow-border/60 bg-severity-yellow-bg/20 p-2 text-xs flex justify-between"
                >
                  <div>
                    <span className="font-semibold text-app-text">{formatMoney(m.amount)}</span>
                    <span className="text-[11px] text-app-text-muted block">
                      {formatDate(m.occurredAt, 'short')} • {m.counterparty || m.description}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Facturas sin cobro */}
        <div className="rounded-card border border-app-border bg-app-surface p-4 space-y-2.5">
          <div className="flex items-center gap-2 border-b border-app-border pb-2">
            <AlertCircle className="h-4 w-4 text-severity-yellow-icon" />
            <h4 className="text-xs font-bold text-app-text">
              Facturas sin cobro ({unmatchedInvoices.length})
            </h4>
          </div>

          {unmatchedInvoices.length === 0 ? (
            <p className="text-xs text-app-text-muted">Todas las facturas tienen cobro registrado.</p>
          ) : (
            <div className="space-y-2">
              {unmatchedInvoices.map((inv) => (
                <div
                  key={inv.id}
                  className="rounded-control border border-severity-yellow-border/60 bg-severity-yellow-bg/20 p-2 text-xs flex justify-between"
                >
                  <div>
                    <span className="font-semibold text-app-text">
                      Factura {inv.type} {String(inv.number).padStart(8, '0')}
                    </span>
                    <span className="text-[11px] text-app-text-muted block">
                      {formatDate(inv.issueDate, 'short')} • {formatMoney(inv.totalAmount)}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
