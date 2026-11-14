import { useState } from 'react';
import { Check } from 'lucide-react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { confirmPeriodInput } from '../api';

interface Props {
  period: string;
  answered: boolean;
  onSuccess?: () => void;
}

export function ConfirmStep({ period, answered, onSuccess }: Props) {
  const [error, setError] = useState<string | null>(null);
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: () => confirmPeriodInput(period),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['my-closings'] });
      setError(null);
      onSuccess?.();
    },
    onError: (err: Error) => {
      setError(err.message || 'Error al confirmar período');
    },
  });

  return (
    <div className="pt-2 border-t border-app-border space-y-2">
      <div className="flex items-center justify-between flex-wrap gap-3">
        <div>
          <h4 className="text-sm font-semibold text-app-text">3. Confirmación</h4>
          <p className="text-xs text-app-text-muted mt-0.5">
            {answered
              ? 'Revisá que hayas cargado todas tus facturas e ingresos antes de confirmar.'
              : 'Respondé el paso 2 sobre ingresos externos para habilitar la confirmación.'}
          </p>
        </div>

        <button
          type="button"
          disabled={!answered || mutation.isPending}
          onClick={() => mutation.mutate()}
          className="h-10 px-5 rounded-control bg-app-primary text-white text-sm font-semibold hover:bg-app-primary-hover disabled:opacity-50 disabled:cursor-not-allowed transition-colors inline-flex items-center gap-2 shadow-sm"
        >
          <Check className="w-4 h-4" />
          {mutation.isPending ? 'Confirmando...' : 'Listo, ya cargué todo'}
        </button>
      </div>

      {error && (
        <div className="rounded-control bg-severity-red-bg border border-severity-red-border p-2.5 text-xs text-severity-red-text">
          {error}
        </div>
      )}
    </div>
  );
}

