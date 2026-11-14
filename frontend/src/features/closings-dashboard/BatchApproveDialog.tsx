import { useState } from 'react';
import * as Dialog from '@radix-ui/react-dialog';
import { useQueryClient } from '@tanstack/react-query';
import { CheckCircle2, AlertTriangle, X, Loader2 } from 'lucide-react';
import { approveBatch } from '../closing-review/api';
import { ApproveBatchResponse } from '../closing-review/schemas';

export interface BatchItem {
  id: string;
  version: number;
  clientName?: string;
}

export interface BatchApproveDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  period: string;
  items: BatchItem[];
  onSuccess?: () => void;
}

const ERROR_LABELS: Record<string, string> = {
  NOT_GREEN: 'El cierre ya no está en estado verde',
  NOT_READY: 'El cierre no está listo para revisión',
  VERSION_CONFLICT: 'El cierre cambió mientras lo revisabas',
  NOT_FOUND: 'Cierre no encontrado',
};

export function BatchApproveDialog({
  open,
  onOpenChange,
  period,
  items,
  onSuccess,
}: BatchApproveDialogProps) {
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<ApproveBatchResponse | null>(null);
  const [error, setError] = useState<string | null>(null);

  const queryClient = useQueryClient();

  const handleClose = () => {
    setResult(null);
    setError(null);
    onOpenChange(false);
  };

  const handleConfirm = async () => {
    try {
      setLoading(true);
      setError(null);
      const batchItems = items.map((it) => ({ id: it.id, version: it.version }));
      const res = await approveBatch(period, batchItems);
      setResult(res);
      queryClient.invalidateQueries({ queryKey: ['closings'] });
      queryClient.invalidateQueries({ queryKey: ['closings-summary'] });
      onSuccess?.();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Error al procesar la aprobación en lote');
    } finally {
      setLoading(false);
    }
  };

  const itemNames = new Map(items.map((it) => [it.id, it.clientName || it.id]));

  return (
    <Dialog.Root open={open} onOpenChange={onOpenChange}>
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-50 bg-[#0A0A0A]/50 backdrop-blur-xs transition-opacity" />
        <Dialog.Content
          aria-describedby="batch-approve-desc"
          className="fixed left-1/2 top-1/2 z-50 w-full max-w-[460px] -translate-x-1/2 -translate-y-1/2 rounded-[10px] border border-app-border bg-app-surface p-6 shadow-xl focus:outline-none"
        >
          <div className="flex items-start justify-between gap-4">
            <div className="flex items-center gap-2.5">
              <div
                className={`flex h-9 w-9 items-center justify-center rounded-full ${
                  result && result.failed > 0
                    ? 'bg-severity-yellow-bg text-severity-yellow-icon'
                    : 'bg-severity-green-bg text-severity-green-icon'
                }`}
              >
                {result && result.failed > 0 ? (
                  <AlertTriangle className="h-5 w-5" aria-hidden="true" />
                ) : (
                  <CheckCircle2 className="h-5 w-5" aria-hidden="true" />
                )}
              </div>
              <Dialog.Title className="text-base font-semibold text-app-text">
                {result ? 'Resultado de la aprobación' : 'Aprobar clientes en lote'}
              </Dialog.Title>
            </div>
            <Dialog.Close asChild>
              <button
                type="button"
                onClick={handleClose}
                aria-label="Cerrar"
                className="rounded-control p-1 text-app-text-muted hover:bg-app-surface-muted hover:text-app-text focus:outline-none focus:ring-2 focus:ring-app-focus"
              >
                <X className="h-4 w-4" aria-hidden="true" />
              </button>
            </Dialog.Close>
          </div>

          <div id="batch-approve-desc" className="mt-4 space-y-3 text-sm text-app-text-muted">
            {!result && !error && (
              <p>
                Vas a aprobar y notificar <strong>{items.length}</strong>{' '}
                {items.length === 1 ? 'cliente verde' : 'clientes verdes'}. Cada cliente recibirá la
                notificación in-app con el detalle y los avisos salientes correspondientes.
              </p>
            )}

            {error && (
              <div
                role="alert"
                className="rounded-control border border-severity-red-border bg-severity-red-bg p-3 text-xs text-severity-red-text"
              >
                {error}
              </div>
            )}

            {result && (
              <div className="space-y-3">
                <p>
                  Se aprobaron <strong>{result.approved}</strong>{' '}
                  {result.approved === 1 ? 'cliente' : 'clientes'}{' '}
                  {result.failed > 0 && (
                    <>
                      y fallaron <strong className="text-severity-red-text">{result.failed}</strong>
                    </>
                  )}
                  .
                </p>

                {result.failed > 0 && (
                  <div className="max-h-48 overflow-y-auto space-y-1.5 rounded-control border border-app-border bg-app-surface-muted p-2.5">
                    {result.results
                      .filter((r) => !r.ok)
                      .map((r) => (
                        <div
                          key={r.id}
                          className="flex items-center justify-between text-xs py-1 border-b border-app-border last:border-b-0"
                        >
                          <span className="font-medium text-app-text truncate mr-2">
                            {itemNames.get(r.id) || r.id}
                          </span>
                          <span className="text-[11px] text-severity-red-text shrink-0">
                            {r.error ? ERROR_LABELS[r.error] || r.error : 'Error'}
                          </span>
                        </div>
                      ))}
                  </div>
                )}
              </div>
            )}
          </div>

          <div className="mt-6 flex items-center justify-end gap-3 border-t border-app-border pt-4">
            {result ? (
              <button
                type="button"
                onClick={handleClose}
                className="inline-flex h-8 items-center rounded-control bg-app-primary px-3 text-xs font-semibold text-white hover:bg-app-primary-hover focus:outline-none focus:ring-2 focus:ring-app-focus"
              >
                Cerrar
              </button>
            ) : (
              <>
                <button
                  type="button"
                  disabled={loading}
                  onClick={handleClose}
                  className="inline-flex h-8 items-center rounded-control border border-app-border bg-white px-3 text-xs font-medium text-app-text hover:bg-app-surface-muted disabled:opacity-50 focus:outline-none focus:ring-2 focus:ring-app-focus"
                >
                  Cancelar
                </button>
                <button
                  type="button"
                  disabled={loading}
                  onClick={handleConfirm}
                  className="inline-flex h-8 items-center gap-1.5 rounded-control bg-app-primary px-3 text-xs font-semibold text-white hover:bg-app-primary-hover disabled:opacity-50 focus:outline-none focus:ring-2 focus:ring-app-focus"
                >
                  {loading && <Loader2 className="h-3.5 w-3.5 animate-spin" />}
                  <span>{loading ? 'Aprobando...' : 'Aprobar y notificar'}</span>
                </button>
              </>
            )}
          </div>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
