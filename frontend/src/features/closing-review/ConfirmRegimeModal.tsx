import { useState } from 'react';
import * as Dialog from '@radix-ui/react-dialog';
import { confirmRegimeTransition } from './api';
import { formatDate } from '../../shared/format';
import { AlertTriangle, X } from 'lucide-react';

export interface ConfirmRegimeModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  clientId: string;
  clientName: string;
  effectiveDate: string;
  onSuccess?: () => void;
}

export function ConfirmRegimeModal({
  open,
  onOpenChange,
  clientId,
  clientName,
  effectiveDate,
  onSuccess,
}: ConfirmRegimeModalProps) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const formattedDate = formatDate(effectiveDate, 'full');

  const handleConfirm = async () => {
    try {
      setLoading(true);
      setError(null);
      // Extraemos YYYY-MM-DD
      const dateOnly = effectiveDate.substring(0, 10);
      await confirmRegimeTransition(clientId, dateOnly, false);
      onOpenChange(false);
      onSuccess?.();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Error al confirmar pase a RI');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Dialog.Root open={open} onOpenChange={onOpenChange}>
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-50 bg-[#0A0A0A]/50 backdrop-blur-xs transition-opacity" />
        <Dialog.Content
          aria-describedby="confirm-regime-desc"
          className="fixed left-1/2 top-1/2 z-50 w-full max-w-[440px] -translate-x-1/2 -translate-y-1/2 rounded-[10px] border border-app-border bg-app-surface p-6 shadow-xl focus:outline-none"
        >
          <div className="flex items-start justify-between gap-4">
            <div className="flex items-center gap-2.5">
              <div className="flex h-9 w-9 items-center justify-center rounded-full bg-severity-red-bg text-severity-red-icon">
                <AlertTriangle className="h-5 w-5" aria-hidden="true" />
              </div>
              <Dialog.Title className="text-base font-semibold text-app-text">
                Confirmar pase a Responsable Inscripto
              </Dialog.Title>
            </div>
            <Dialog.Close asChild>
              <button
                type="button"
                aria-label="Cerrar confirmación"
                className="rounded-control p-1 text-app-text-muted hover:bg-app-surface-muted hover:text-app-text focus:outline-none focus:ring-2 focus:ring-app-focus"
              >
                <X className="h-4 w-4" aria-hidden="true" />
              </button>
            </Dialog.Close>
          </div>

          <div id="confirm-regime-desc" className="mt-4 space-y-3 text-sm text-app-text-muted">
            <p>
              Estás por confirmar el pase al régimen de <strong>Responsable Inscripto</strong> para{' '}
              <strong className="text-app-text">{clientName}</strong> con fecha de exclusión efectiva{' '}
              <span className="font-semibold text-app-text">{formattedDate}</span>.
            </p>
            <p className="text-xs">
              A partir de dicha fecha, el contribuyente deja de pertenecer al Régimen Simplificado (Monotributo)
              y sus liquidaciones deberán realizarse bajo IVA y Ganancias.
            </p>
          </div>

          {error && (
            <div className="mt-4 rounded-control border border-severity-red-border bg-severity-red-bg p-3 text-xs text-severity-red-text">
              {error}
            </div>
          )}

          <div className="mt-6 flex items-center justify-end gap-3">
            <Dialog.Close asChild>
              <button
                type="button"
                disabled={loading}
                className="inline-flex h-8 items-center rounded-control border border-app-border bg-white px-3 text-xs font-medium text-app-text hover:bg-app-surface-muted focus:outline-none focus:ring-2 focus:ring-app-focus"
              >
                Cancelar
              </button>
            </Dialog.Close>

            <button
              type="button"
              onClick={handleConfirm}
              disabled={loading}
              className="inline-flex h-8 items-center rounded-control border border-severity-red-border bg-severity-red-bg px-3 text-xs font-semibold text-severity-red-text hover:bg-severity-red-border/50 focus:outline-none focus:ring-2 focus:ring-severity-red-icon disabled:opacity-50"
            >
              {loading ? 'Confirmando...' : 'Confirmar pase a RI'}
            </button>
          </div>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
