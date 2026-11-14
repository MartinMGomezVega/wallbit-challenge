import { ClosingDetailDTO } from '../schemas';
import { formatMoney, formatDate, formatPercentBp } from '../../../shared/format';
import { ProgressBar } from '../../../shared/ui/ProgressBar';

export interface SummarySectionProps {
  detail: ClosingDetailDTO;
}

export function SummarySection({ detail }: SummarySectionProps) {
  const { closing, client } = detail;
  const isMono = client.regime === 'MONOTRIBUTO';
  const mono = closing.mono;
  const ri = closing.ri;

  return (
    <div className="space-y-6">
      {/* Facturado del período */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        <div className="rounded-card border border-app-border bg-app-surface p-4">
          <span className="text-[11px] font-semibold uppercase tracking-wider text-app-text-muted">
            Facturado en el mes
          </span>
          <div className="mt-1 flex items-baseline gap-2">
            <span className="text-lg font-bold text-app-text">
              {formatMoney(
                isMono
                  ? mono?.billedInPeriodUSD
                  : ri?.billedInPeriodUSD
              )}
            </span>
            <span className="text-xs text-app-text-muted">
              {formatMoney(
                isMono
                  ? mono?.billedInPeriodARS
                  : ri?.billedInPeriodARS
              )}
            </span>
          </div>
        </div>

        {/* Monotributo: Acumulado 12m y topes */}
        {isMono && mono && (
          <div className="rounded-card border border-app-border bg-app-surface p-4 sm:col-span-2">
            <div className="flex items-center justify-between text-xs">
              <span className="text-[11px] font-semibold uppercase tracking-wider text-app-text-muted">
                Acumulado 12 meses
              </span>
              <span className="font-semibold text-app-text">
                {formatMoney(mono.billed12mARS)} ({formatPercentBp(mono.pctOfKCapBp)} de tope K)
              </span>
            </div>
            <div className="mt-2.5">
              <ProgressBar pctBp={mono.pctOfKCapBp} />
            </div>
            <div className="mt-2 flex items-center justify-between text-[11px] text-app-text-muted">
              <span>Categoría actual: {mono.currentCategory}</span>
              {mono.requiredCategory && mono.requiredCategory !== mono.currentCategory && (
                <span className="font-medium text-severity-yellow-text">
                  Requiere pase a: {mono.requiredCategory}
                </span>
              )}
            </div>
          </div>
        )}
      </div>

      {/* A pagar y vencimientos */}
      <div className="rounded-card border border-app-border bg-app-surface p-4 space-y-3">
        <h3 className="text-xs font-bold uppercase tracking-wider text-app-text-muted border-b border-app-border pb-2">
          Liquidación impositiva del período
        </h3>

        {isMono && mono && (
          <div className="flex flex-wrap items-center justify-between gap-4 pt-1">
            <div>
              <span className="text-xs font-medium text-app-text-muted block">
                Cuota mensual Monotributo
              </span>
              <span className="text-base font-bold text-app-text">
                {formatMoney(mono.fee)}
              </span>
            </div>
            <div className="text-right">
              <span className="text-xs text-app-text-muted block">Fecha de vencimiento</span>
              <span className="text-xs font-semibold text-app-text">
                {formatDate(mono.feeDueDate, 'full')}
              </span>
            </div>
          </div>
        )}

        {!isMono && ri && (
          <div className="space-y-4 pt-1">
            {/* Total RI */}
            <div className="flex items-baseline justify-between border-b border-app-border pb-2">
              <span className="text-xs font-bold text-app-text">Total a ingresar</span>
              <span className="text-lg font-bold text-app-text">
                {formatMoney(ri.totalToPay)}
              </span>
            </div>

            {/* Desglose RI */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-3 text-xs">
              {/* IVA */}
              <div className="rounded-control border border-app-border bg-app-surface-muted p-3">
                <span className="font-semibold text-app-text block mb-1">IVA</span>
                <div className="space-y-1 text-[11px] text-app-text-muted">
                  <div className="flex justify-between">
                    <span>Débito fiscal:</span>
                    <span className="font-medium text-app-text">{formatMoney(ri.debitoFiscalARS)}</span>
                  </div>
                  <div className="flex justify-between">
                    <span>Crédito fiscal:</span>
                    <span className="font-medium text-app-text">{formatMoney(ri.creditoFiscalARS)}</span>
                  </div>
                  <div className="flex justify-between pt-1 border-t border-app-border font-semibold text-app-text">
                    <span>A pagar:</span>
                    <span>{formatMoney(ri.ivaToPay)}</span>
                  </div>
                  <div className="text-[10px] text-right text-app-text-muted">
                    Vence: {formatDate(ri.vatDueDate, 'short')}
                  </div>
                </div>
              </div>

              {/* Autónomos */}
              <div className="rounded-control border border-app-border bg-app-surface-muted p-3">
                <span className="font-semibold text-app-text block mb-1">Autónomos</span>
                <div className="space-y-1 text-[11px] text-app-text-muted">
                  <div className="flex justify-between">
                    <span>Categoría:</span>
                    <span className="font-medium text-app-text">{ri.autonomousCategory}</span>
                  </div>
                  <div className="flex justify-between pt-1 border-t border-app-border font-semibold text-app-text">
                    <span>A pagar:</span>
                    <span>{formatMoney(ri.autonomousFee)}</span>
                  </div>
                  <div className="text-[10px] text-right text-app-text-muted">
                    Vence: {formatDate(ri.selfEmployedDueDate, 'short')}
                  </div>
                </div>
              </div>

              {/* Anticipo Ganancias */}
              <div className="rounded-control border border-app-border bg-app-surface-muted p-3">
                <span className="font-semibold text-app-text block mb-1">Anticipo Ganancias</span>
                <div className="space-y-1 text-[11px] text-app-text-muted">
                  <div className="flex justify-between pt-1 font-semibold text-app-text">
                    <span>A pagar:</span>
                    <span>{formatMoney(ri.incomeTaxAdvance)}</span>
                  </div>
                  <div className="text-[10px] text-right text-app-text-muted">
                    Vence: {formatDate(ri.incomeTaxAdvanceDueDate, 'short')}
                  </div>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
