import { useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { confirmInputOnBehalf } from '../api';
import { CheckCircle2, Plus, Trash2, AlertCircle, HelpCircle } from 'lucide-react';

export interface ConfirmOnBehalfFormProps {
  clientId: string;
  period: string;
  closingId: string;
  onSuccess?: () => void;
}

interface IncomeRow {
  amountUSD: string;
  description: string;
}

export function ConfirmOnBehalfForm({
  clientId,
  period,
  closingId,
  onSuccess,
}: ConfirmOnBehalfFormProps) {
  const queryClient = useQueryClient();
  const [declaredNone, setDeclaredNone] = useState(true);
  const [items, setItems] = useState<IncomeRow[]>([
    { amountUSD: '', description: '' },
  ]);
  const [note, setNote] = useState('');
  const [formError, setFormError] = useState<string | null>(null);

  const confirmMutation = useMutation({
    mutationFn: async () => {
      const trimmedNote = note.trim();
      if (!trimmedNote) {
        throw new Error('La nota de cómo te lo confirmó el cliente es obligatoria.');
      }
      if (trimmedNote.length > 300) {
        throw new Error('La nota no puede superar los 300 caracteres.');
      }

      if (declaredNone) {
        return confirmInputOnBehalf(clientId, period, {
          externalIncome: { none: true },
          note: trimmedNote,
        });
      }

      // Validar ítems
      const parsedItems = items.map((it) => {
        const num = parseFloat(it.amountUSD.replace(',', '.'));
        if (isNaN(num) || num <= 0) {
          throw new Error('Todos los cobros deben tener un monto mayor a cero.');
        }
        if (!it.description.trim()) {
          throw new Error('Todos los cobros deben incluir una descripción.');
        }
        return {
          amount: {
            amount: Math.round(num * 100),
            currency: 'USD' as const,
          },
          description: it.description.trim(),
        };
      });

      if (parsedItems.length === 0) {
        throw new Error('Debes ingresar al menos un cobro o seleccionar que no tuvo.');
      }

      return confirmInputOnBehalf(clientId, period, {
        externalIncome: { items: parsedItems },
        note: trimmedNote,
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['closing-detail', closingId] });
      queryClient.invalidateQueries({ queryKey: ['closings'] });
      queryClient.invalidateQueries({ queryKey: ['closings-summary'] });
      queryClient.invalidateQueries({ queryKey: ['stats'] });
      onSuccess?.();
    },
    onError: (err) => {
      const msg = err instanceof Error ? err.message : 'Error al confirmar período';
      setFormError(msg);
    },
  });

  const handleAddItem = () => {
    setItems((prev) => [...prev, { amountUSD: '', description: '' }]);
  };

  const handleRemoveItem = (index: number) => {
    setItems((prev) => prev.filter((_, idx) => idx !== index));
  };

  const handleItemChange = (index: number, field: keyof IncomeRow, value: string) => {
    setItems((prev) =>
      prev.map((it, idx) => (idx === index ? { ...it, [field]: value } : it))
    );
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setFormError(null);
    confirmMutation.mutate();
  };

  return (
    <div className="rounded-card border-2 border-app-primary/30 bg-app-surface p-5 shadow-sm space-y-4">
      <div className="flex items-center gap-2 border-b border-app-border pb-3">
        <HelpCircle className="h-5 w-5 text-app-primary flex-shrink-0" aria-hidden="true" />
        <div>
          <h3 className="text-sm font-bold text-app-text">Confirmar en nombre del cliente</h3>
          <p className="text-xs text-app-text-muted">
            Este cierre está esperando respuesta sobre ingresos externos. Podés confirmarlo si el cliente te lo comunicó por otro canal.
          </p>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="space-y-4 text-xs">
        {/* Opciones de cobros */}
        <div className="space-y-2">
          <label className="text-xs font-semibold text-app-text block">
            Cobros percibidos fuera de Wallbit en el período:
          </label>
          <div className="space-y-2">
            <label htmlFor="radio-no-income" className="flex items-center gap-2.5 cursor-pointer">
              <input
                id="radio-no-income"
                type="radio"
                name="externalIncomeOption"
                checked={declaredNone}
                onChange={() => setDeclaredNone(true)}
                className="h-4 w-4 text-app-primary focus:ring-app-primary"
              />
              <span className="font-medium text-app-text">No tuvo cobros fuera de Wallbit</span>
            </label>

            <label htmlFor="radio-had-income" className="flex items-center gap-2.5 cursor-pointer">
              <input
                id="radio-had-income"
                type="radio"
                name="externalIncomeOption"
                checked={!declaredNone}
                onChange={() => setDeclaredNone(false)}
                className="h-4 w-4 text-app-primary focus:ring-app-primary"
              />
              <span className="font-medium text-app-text">Tuvo cobros fuera de Wallbit</span>
            </label>
          </div>
        </div>

        {/* Formulario de cobros cuando declaredNone es false */}
        {!declaredNone && (
          <div className="space-y-3 rounded-control border border-app-border bg-app-surface-muted p-3">
            <span className="font-semibold text-app-text block">Detalle de cobros externos:</span>
            {items.map((it, idx) => (
              <div key={`income-${idx}`} className="flex items-center gap-2">
                <div className="w-32 flex-shrink-0">
                  <input
                    type="text"
                    inputMode="decimal"
                    placeholder="Monto (USD)"
                    value={it.amountUSD}
                    onChange={(e) => handleItemChange(idx, 'amountUSD', e.target.value)}
                    className="w-full rounded border border-app-border bg-app-surface px-2.5 py-1.5 text-xs text-app-text tabular-nums focus:outline-none focus:ring-2 focus:ring-app-primary"
                  />
                </div>
                <div className="flex-1">
                  <input
                    type="text"
                    placeholder="Descripción (ej. Cobro directo exterior)"
                    value={it.description}
                    onChange={(e) => handleItemChange(idx, 'description', e.target.value)}
                    className="w-full rounded border border-app-border bg-app-surface px-2.5 py-1.5 text-xs text-app-text focus:outline-none focus:ring-2 focus:ring-app-primary"
                  />
                </div>
                {items.length > 1 && (
                  <button
                    type="button"
                    onClick={() => handleRemoveItem(idx)}
                    className="p-1.5 text-app-text-muted hover:text-severity-red-icon rounded focus:outline-none focus:ring-2 focus:ring-app-primary"
                    aria-label={`Eliminar cobro ${idx + 1}`}
                  >
                    <Trash2 className="h-4 w-4" aria-hidden="true" />
                  </button>
                )}
              </div>
            ))}
            <button
              type="button"
              onClick={handleAddItem}
              className="inline-flex items-center gap-1.5 text-app-primary hover:underline font-medium text-xs pt-1 focus:outline-none"
            >
              <Plus className="h-3.5 w-3.5" aria-hidden="true" />
              <span>Agregar otro cobro</span>
            </button>
          </div>
        )}

        {/* Campo de auditoría: ¿Cómo te lo confirmó? */}
        <div className="space-y-1.5">
          <label htmlFor="confirm-note" className="text-xs font-semibold text-app-text block">
            ¿Cómo te lo confirmó el cliente? <span className="text-severity-red-text">*</span>
          </label>
          <input
            id="confirm-note"
            type="text"
            required
            maxLength={300}
            placeholder="Ej. Me lo confirmó por mail el 06/11"
            value={note}
            onChange={(e) => setNote(e.target.value)}
            className="w-full rounded border border-app-border bg-app-surface px-3 py-2 text-xs text-app-text focus:outline-none focus:ring-2 focus:ring-app-primary"
          />
          <div className="flex justify-between text-[11px] text-app-text-muted">
            <span>Evidencia registrada con fines de auditoría</span>
            <span>{note.length}/300</span>
          </div>
        </div>

        {formError && (
          <div
            role="alert"
            className="flex items-center gap-2 rounded-control border border-severity-red-border bg-severity-red-bg p-3 text-xs text-severity-red-text"
          >
            <AlertCircle className="h-4 w-4 flex-shrink-0" aria-hidden="true" />
            <span>{formError}</span>
          </div>
        )}

        <div className="pt-2 flex justify-end">
          <button
            type="submit"
            disabled={confirmMutation.isPending || !note.trim()}
            className="inline-flex items-center gap-2 rounded-control bg-app-primary px-4 py-2 text-xs font-semibold text-white hover:bg-app-primary-hover focus:outline-none focus:ring-2 focus:ring-app-primary disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <CheckCircle2 className="h-4 w-4" aria-hidden="true" />
            <span>{confirmMutation.isPending ? 'Confirmando...' : 'Confirmar período'}</span>
          </button>
        </div>
      </form>
    </div>
  );
}
