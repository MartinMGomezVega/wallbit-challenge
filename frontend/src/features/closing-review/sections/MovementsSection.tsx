import { useState } from 'react';
import { ClosingDetailDTO, PeriodMovementDTO } from '../schemas';
import { formatMoney, formatDate } from '../../../shared/format';
import { updateMovementClassification } from '../api';
import { ArrowDownLeft, ArrowUpRight, Loader2, RefreshCw } from 'lucide-react';

export interface MovementsSectionProps {
  detail: ClosingDetailDTO;
  onRefreshDetail: () => void;
}

export function MovementsSection({ detail, onRefreshDetail }: MovementsSectionProps) {
  const { periodMovements, closing } = detail;
  const [updatingId, setUpdatingId] = useState<string | null>(null);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  const handleToggleClassification = async (item: PeriodMovementDTO) => {
    const isIncome = item.classification.nature === 'INCOME';
    const newNature = isIncome ? 'OTHER' : 'INCOME';

    setUpdatingId(item.movement.id);
    setErrorMsg(null);

    try {
      await updateMovementClassification(
        item.movement.id,
        newNature,
        isIncome ? 'Reclasificado a No ingreso por el contador' : 'Reclasificado a Ingreso por el contador'
      );
      onRefreshDetail();
    } catch (err) {
      setErrorMsg((err as Error)?.message || 'Error al actualizar la clasificación');
    } finally {
      setUpdatingId(null);
    }
  };

  if (!periodMovements || periodMovements.length === 0) {
    return (
      <div className="rounded-card border border-app-border bg-app-surface p-8 text-center text-xs text-app-text-muted">
        No se registraron movimientos en el período {closing.period}.
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {errorMsg && (
        <div
          role="alert"
          className="rounded-card border border-severity-red-border bg-severity-red-bg px-4 py-2.5 text-xs text-severity-red-text"
        >
          {errorMsg}
        </div>
      )}

      <div className="rounded-card border border-app-border bg-app-surface overflow-hidden shadow-sm">
        <div className="overflow-x-auto">
          <table className="w-full text-left border-collapse text-xs">
            <thead>
              <tr className="border-b border-app-border bg-app-surface-muted text-[11px] font-semibold uppercase tracking-wider text-app-text-muted">
                <th className="px-3 py-2.5">Fecha</th>
                <th className="px-3 py-2.5">Tipo</th>
                <th className="px-3 py-2.5">Contraparte / Detalle</th>
                <th className="px-3 py-2.5 text-right">Importe</th>
                <th className="px-3 py-2.5">Clasificación</th>
                <th className="px-3 py-2.5 text-right">Acción</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-app-border">
              {periodMovements.map((pm) => {
                const isIncome = pm.classification.nature === 'INCOME';
                const isUpdating = updatingId === pm.movement.id;

                return (
                  <tr key={pm.movement.id} className="hover:bg-app-surface-muted/50">
                    <td className="px-3 py-2.5 tabular-nums text-app-text">
                      {formatDate(pm.movement.occurredAt, 'short')}
                    </td>
                    <td className="px-3 py-2.5">
                      <span className="inline-flex items-center gap-1 font-medium text-app-text">
                        {isIncome ? (
                          <ArrowDownLeft className="h-3 w-3 text-severity-green-icon" />
                        ) : (
                          <ArrowUpRight className="h-3 w-3 text-app-text-muted" />
                        )}
                        <span>{pm.movement.type}</span>
                      </span>
                    </td>
                    <td className="px-3 py-2.5">
                      <div className="font-medium text-app-text truncate max-w-xs">
                        {pm.movement.counterparty || pm.movement.description || '-'}
                      </div>
                    </td>
                    <td className="px-3 py-2.5 text-right tabular-nums font-semibold text-app-text">
                      {formatMoney(pm.movement.amount)}
                    </td>
                    <td className="px-3 py-2.5">
                      <div className="flex items-center gap-1.5">
                        <span
                          className={`inline-block px-2 py-0.5 rounded-full text-[10px] font-semibold ${
                            isIncome
                              ? 'bg-severity-green-bg text-severity-green-text border border-severity-green-border'
                              : 'bg-app-surface-muted text-app-text-muted border border-app-border'
                          }`}
                        >
                          {isIncome ? 'Ingreso' : 'No ingreso'}
                        </span>
                        {pm.classification.overridden && (
                          <span className="text-[10px] text-app-primary font-medium" title="Modificado manualmente">
                            (editado)
                          </span>
                        )}
                      </div>
                      <span className="text-[10px] text-app-text-muted block mt-0.5">
                        {pm.classification.rule}
                      </span>
                    </td>
                    <td className="px-3 py-2.5 text-right">
                      <button
                        type="button"
                        onClick={() => handleToggleClassification(pm)}
                        disabled={isUpdating}
                        className="inline-flex items-center gap-1 rounded-control border border-app-border bg-white px-2 py-1 text-[11px] font-medium text-app-text hover:bg-app-surface-muted disabled:opacity-50 focus:outline-none focus:ring-2 focus:ring-app-focus"
                      >
                        {isUpdating ? (
                          <Loader2 className="h-3 w-3 animate-spin text-app-primary" />
                        ) : (
                          <RefreshCw className="h-3 w-3 text-app-text-muted" />
                        )}
                        <span>{isIncome ? 'Cambiar a No ingreso' : 'Cambiar a Ingreso'}</span>
                      </button>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
