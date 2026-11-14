import { CheckCircle2, AlertTriangle, XCircle, Info } from 'lucide-react';
import { ImportResultDTO } from '../schemas';

export interface ImportResultViewProps {
  result: ImportResultDTO;
}

export function ImportResultView({ result }: ImportResultViewProps) {
  const { imported, skippedDuplicates, warnings, errors } = result;
  const isSuccess = errors.length === 0 && imported > 0;
  const hasOnlyDuplicates = errors.length === 0 && imported === 0 && skippedDuplicates > 0;

  return (
    <div className="space-y-3 rounded-control border border-app-border bg-app-surface-muted p-4 text-xs" aria-live="polite">
      <div className="flex items-center gap-2">
        {isSuccess ? (
          <CheckCircle2 className="h-5 w-5 text-severity-green-icon flex-shrink-0" aria-hidden="true" />
        ) : hasOnlyDuplicates ? (
          <Info className="h-5 w-5 text-app-primary flex-shrink-0" aria-hidden="true" />
        ) : errors.length > 0 ? (
          <XCircle className="h-5 w-5 text-severity-red-icon flex-shrink-0" aria-hidden="true" />
        ) : (
          <AlertTriangle className="h-5 w-5 text-severity-yellow-icon flex-shrink-0" aria-hidden="true" />
        )}
        <div className="font-semibold text-app-text text-sm">
          {isSuccess
            ? `Se importaron ${imported} ${imported === 1 ? 'factura' : 'facturas'} correctamente`
            : hasOnlyDuplicates
            ? 'No se importaron facturas nuevas (ya existían en el sistema)'
            : errors.length > 0
            ? 'Ocurrieron errores al procesar el archivo'
            : 'No se procesaron facturas'}
        </div>
      </div>

      <div className="grid grid-cols-2 gap-2 text-xs">
        <div className="rounded border border-app-border bg-app-surface p-2">
          <span className="text-app-text-muted block text-[11px]">Importadas</span>
          <span className="font-semibold text-app-text text-base tabular-nums">{imported}</span>
        </div>
        <div className="rounded border border-app-border bg-app-surface p-2">
          <span className="text-app-text-muted block text-[11px]">Duplicadas omitidas</span>
          <span className="font-semibold text-app-text text-base tabular-nums">{skippedDuplicates}</span>
        </div>
      </div>

      {warnings.length > 0 && (
        <div className="space-y-1.5 pt-1">
          <div className="flex items-center gap-1.5 font-medium text-severity-yellow-icon">
            <AlertTriangle className="h-3.5 w-3.5" aria-hidden="true" />
            <span>Advertencias ({warnings.length})</span>
          </div>
          <ul className="list-disc list-inside space-y-1 text-app-text pl-1 text-[11px]">
            {warnings.map((w, idx) => (
              <li key={`warn-${idx}`}>
                {w.line > 0 ? <span className="font-mono font-medium">Línea {w.line}: </span> : null}
                {w.message}
              </li>
            ))}
          </ul>
        </div>
      )}

      {errors.length > 0 && (
        <div className="space-y-1.5 pt-1">
          <div className="flex items-center gap-1.5 font-medium text-severity-red-icon">
            <XCircle className="h-3.5 w-3.5" aria-hidden="true" />
            <span>Errores ({errors.length})</span>
          </div>
          <ul className="list-disc list-inside space-y-1 text-severity-red-text pl-1 text-[11px]">
            {errors.map((e, idx) => (
              <li key={`err-${idx}`}>
                <span className="font-mono font-medium">Línea {e.line}: </span>
                {e.message}
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
}

