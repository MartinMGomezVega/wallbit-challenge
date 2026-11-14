import { useState } from 'react';
import { Plus, Trash2, CheckCircle2 } from 'lucide-react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { saveExternalIncome } from '../api';
import { ExternalIncomeItem } from '../schemas';

interface Props {
  period: string;
  periodMonthName: string;
  answered: boolean;
}

export function ExternalIncomeStep({ period, periodMonthName, answered }: Props) {
  const [hasExternal, setHasExternal] = useState<boolean | null>(answered ? false : null);
  const [items, setItems] = useState<Array<{ amountUsd: string; description: string }>>([
    { amountUsd: '', description: '' },
  ]);
  const [savedSuccess, setSavedSuccess] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (data: { declaredNoExternalIncome: boolean; externalIncome: ExternalIncomeItem[] }) =>
      saveExternalIncome(period, data),
    onSuccess: () => {
      setSavedSuccess(true);
      setError(null);
      queryClient.invalidateQueries({ queryKey: ['my-closings'] });
    },
    onError: (err: Error) => {
      setError(err.message || 'Error al guardar respuesta');
    },
  });

  const handleSelectNo = () => {
    setHasExternal(false);
    setSavedSuccess(false);
    mutation.mutate({
      declaredNoExternalIncome: true,
      externalIncome: [],
    });
  };

  const handleSelectYes = () => {
    setHasExternal(true);
    setSavedSuccess(false);
    if (items.length === 0) {
      setItems([{ amountUsd: '', description: '' }]);
    }
  };

  const handleAddItem = () => {
    setItems((prev) => [...prev, { amountUsd: '', description: '' }]);
  };

  const handleRemoveItem = (index: number) => {
    setItems((prev) => prev.filter((_, i) => i !== index));
  };

  const handleUpdateItem = (index: number, field: 'amountUsd' | 'description', value: string) => {
    setItems((prev) =>
      prev.map((it, i) => (i === index ? { ...it, [field]: value } : it))
    );
  };

  const handleSaveYes = (e: React.FormEvent) => {
    e.preventDefault();
    const parsedItems: ExternalIncomeItem[] = [];
    for (const it of items) {
      const num = parseFloat(it.amountUsd.replace(',', '.'));
      if (isNaN(num) || num <= 0) {
        setError('Todos los montos deben ser mayores a 0');
        return;
      }
      if (!it.description.trim()) {
        setError('Por favor completá la descripción de cada cobro');
        return;
      }
      parsedItems.push({
        amount: {
          amount: Math.round(num * 100),
          currency: 'USD',
        },
        description: it.description.trim(),
      });
    }

    if (parsedItems.length === 0) {
      setError('Cargá al menos un cobro o seleccioná "No"');
      return;
    }

    mutation.mutate({
      declaredNoExternalIncome: false,
      externalIncome: parsedItems,
    });
  };

  return (
    <div className="space-y-3">
      <h3 className="text-sm font-semibold text-app-text">
        2. ¿Cobraste algo fuera de Wallbit en {periodMonthName}?
      </h3>

      <div className="flex items-center gap-3">
        <button
          type="button"
          onClick={handleSelectNo}
          className={`h-9 px-4 rounded-control border text-sm font-medium transition-colors ${
            hasExternal === false
              ? 'border-app-primary bg-app-primary-soft text-app-primary'
              : 'border-app-border bg-app-surface text-app-text hover:bg-app-surface-muted'
          }`}
        >
          No, solo en Wallbit
        </button>

        <button
          type="button"
          onClick={handleSelectYes}
          className={`h-9 px-4 rounded-control border text-sm font-medium transition-colors ${
            hasExternal === true
              ? 'border-app-primary bg-app-primary-soft text-app-primary'
              : 'border-app-border bg-app-surface text-app-text hover:bg-app-surface-muted'
          }`}
        >
          Sí, tuve otros cobros
        </button>

        {savedSuccess && (
          <span className="text-xs text-severity-green-text flex items-center gap-1">
            <CheckCircle2 className="w-3.5 h-3.5" />
            Guardado
          </span>
        )}
      </div>

      {hasExternal === true && (
        <form onSubmit={handleSaveYes} className="mt-3 p-3.5 rounded-card border border-app-border bg-app-surface-muted/30 space-y-3">
          <p className="text-xs text-app-text-muted">
            Declarar tus cobros externos en USD para que tu contador los considere en el cálculo mensual.
          </p>

          {error && (
            <div className="rounded-control bg-severity-red-bg border border-severity-red-border p-2 text-xs text-severity-red-text">
              {error}
            </div>
          )}

          <div className="space-y-2">
            {items.map((it, idx) => (
              <div key={idx} className="flex items-center gap-2">
                <input
                  type="text"
                  placeholder="Monto US$ (ej. 1500)"
                  value={it.amountUsd}
                  onChange={(e) => handleUpdateItem(idx, 'amountUsd', e.target.value)}
                  className="w-36 h-9 rounded-control border border-app-border bg-app-surface px-2.5 text-xs text-app-text tabular-nums focus:outline-none focus:ring-2 focus:ring-app-focus"
                />
                <input
                  type="text"
                  placeholder="Descripción (ej. Consultoría cliente directo)"
                  value={it.description}
                  onChange={(e) => handleUpdateItem(idx, 'description', e.target.value)}
                  className="flex-1 h-9 rounded-control border border-app-border bg-app-surface px-2.5 text-xs text-app-text focus:outline-none focus:ring-2 focus:ring-app-focus"
                />
                {items.length > 1 && (
                  <button
                    type="button"
                    onClick={() => handleRemoveItem(idx)}
                    className="p-1.5 text-app-text-muted hover:text-severity-red-text"
                    aria-label="Eliminar cobro"
                  >
                    <Trash2 className="w-4 h-4" />
                  </button>
                )}
              </div>
            ))}
          </div>

          <div className="flex items-center justify-between pt-2">
            <button
              type="button"
              onClick={handleAddItem}
              className="inline-flex items-center gap-1 text-xs text-app-primary hover:underline font-medium"
            >
              <Plus className="w-3.5 h-3.5" />
              Agregar otro cobro
            </button>

            <button
              type="submit"
              disabled={mutation.isPending}
              className="h-8 px-3 rounded-control bg-app-primary text-white text-xs font-medium hover:bg-app-primary-hover disabled:opacity-50"
            >
              {mutation.isPending ? 'Guardando...' : 'Guardar cobros externos'}
            </button>
          </div>
        </form>
      )}
    </div>
  );
}
