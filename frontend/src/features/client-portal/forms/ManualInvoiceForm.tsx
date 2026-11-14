import { useState } from 'react';
import * as Dialog from '@radix-ui/react-dialog';
import { X, Plus } from 'lucide-react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { createManualInvoice, ManualInvoicePayload } from '../api';

interface Props {
  regime: string;
  defaultDate?: string;
  onSuccess?: () => void;
}

export function ManualInvoiceForm({ regime, defaultDate, onSuccess }: Props) {
  const [open, setOpen] = useState(false);
  const queryClient = useQueryClient();

  const isRI = regime === 'RI';
  const defaultType = isRI ? 'A' : 'C';

  const [type, setType] = useState(defaultType);
  const [pointOfSale, setPointOfSale] = useState('1');
  const [number, setNumber] = useState('1');
  const [issueDate, setIssueDate] = useState(defaultDate || new Date().toISOString().slice(0, 10));
  const [currency, setCurrency] = useState<'ARS' | 'USD'>('ARS');
  const [amount, setAmount] = useState('');
  const [receiverDoc, setReceiverDoc] = useState('');
  const [receiverName, setReceiverName] = useState('');
  const [error, setError] = useState<string | null>(null);

  const mutation = useMutation({
    mutationFn: createManualInvoice,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['my-closings'] });
      setOpen(false);
      setError(null);
      onSuccess?.();
    },
    onError: (err: Error) => {
      setError(err.message || 'Error al guardar la factura');
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const numAmount = parseFloat(amount.replace(',', '.'));
    if (isNaN(numAmount) || numAmount <= 0) {
      setError('El importe debe ser mayor a 0');
      return;
    }

    const payload: ManualInvoicePayload = {
      type,
      pointOfSale: parseInt(pointOfSale, 10) || 1,
      number: parseInt(number, 10) || 1,
      issueDate: new Date(issueDate).toISOString(),
      currency,
      netAmount: { amount: Math.round(numAmount * 100), currency },
      totalAmount: { amount: Math.round(numAmount * 100), currency },
      receiverDoc: receiverDoc.trim() || undefined,
      receiverName: receiverName.trim() || undefined,
    };

    mutation.mutate(payload);
  };

  return (
    <Dialog.Root open={open} onOpenChange={setOpen}>
      <Dialog.Trigger asChild>
        <button
          type="button"
          className="inline-flex items-center gap-1.5 text-xs font-medium text-app-primary hover:underline focus:outline-none focus-visible:ring-2 focus-visible:ring-app-focus rounded px-1 py-0.5"
        >
          <Plus className="w-3.5 h-3.5" aria-hidden="true" />
          Cargar una factura a mano
        </button>
      </Dialog.Trigger>
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-50 bg-black/40 backdrop-blur-sm" />
        <Dialog.Content
          className="fixed left-1/2 top-1/2 z-50 w-full max-w-md -translate-x-1/2 -translate-y-1/2 rounded-card border border-app-border bg-app-surface p-6 shadow-xl"
          aria-describedby="manual-invoice-desc"
        >
          <div className="flex items-center justify-between pb-3 border-b border-app-border">
            <Dialog.Title className="text-base font-semibold text-app-text">
              Cargar factura emitida
            </Dialog.Title>
            <Dialog.Close asChild>
              <button
                type="button"
                className="rounded-control p-1 text-app-text-muted hover:bg-app-surface-muted hover:text-app-text focus:outline-none focus-visible:ring-2 focus-visible:ring-app-focus"
                aria-label="Cerrar"
              >
                <X className="h-4 w-4" />
              </button>
            </Dialog.Close>
          </div>

          <p id="manual-invoice-desc" className="sr-only">
            Formulario para cargar manualmente un comprobante emitido
          </p>

          <form onSubmit={handleSubmit} className="mt-4 space-y-3.5 text-sm">
            {error && (
              <div className="rounded-control bg-severity-red-bg border border-severity-red-border p-2 text-xs text-severity-red-text">
                {error}
              </div>
            )}

            <div className="grid grid-cols-3 gap-2.5">
              <div>
                <label htmlFor="inv-type" className="block text-xs font-medium text-app-text-muted mb-1">
                  Tipo
                </label>
                <select
                  id="inv-type"
                  value={type}
                  onChange={(e) => setType(e.target.value)}
                  className="w-full h-9 rounded-control border border-app-border bg-app-surface px-2 text-sm text-app-text focus:outline-none focus:ring-2 focus:ring-app-focus"
                >
                  {isRI ? (
                    <>
                      <option value="A">Factura A</option>
                      <option value="B">Factura B</option>
                      <option value="E">Factura E</option>
                    </>
                  ) : (
                    <>
                      <option value="C">Factura C</option>
                      <option value="E">Factura E</option>
                    </>
                  )}
                </select>
              </div>

              <div>
                <label htmlFor="inv-pos" className="block text-xs font-medium text-app-text-muted mb-1">
                  Punto de venta
                </label>
                <input
                  id="inv-pos"
                  type="number"
                  min="1"
                  value={pointOfSale}
                  onChange={(e) => setPointOfSale(e.target.value)}
                  className="w-full h-9 rounded-control border border-app-border bg-app-surface px-2 text-sm text-app-text focus:outline-none focus:ring-2 focus:ring-app-focus"
                />
              </div>

              <div>
                <label htmlFor="inv-num" className="block text-xs font-medium text-app-text-muted mb-1">
                  Número
                </label>
                <input
                  id="inv-num"
                  type="number"
                  min="1"
                  value={number}
                  onChange={(e) => setNumber(e.target.value)}
                  className="w-full h-9 rounded-control border border-app-border bg-app-surface px-2 text-sm text-app-text focus:outline-none focus:ring-2 focus:ring-app-focus"
                />
              </div>
            </div>

            <div className="grid grid-cols-2 gap-2.5">
              <div>
                <label htmlFor="inv-date" className="block text-xs font-medium text-app-text-muted mb-1">
                  Fecha de emisión
                </label>
                <input
                  id="inv-date"
                  type="date"
                  value={issueDate}
                  onChange={(e) => setIssueDate(e.target.value)}
                  className="w-full h-9 rounded-control border border-app-border bg-app-surface px-2 text-sm text-app-text focus:outline-none focus:ring-2 focus:ring-app-focus"
                />
              </div>

              <div>
                <label htmlFor="inv-currency" className="block text-xs font-medium text-app-text-muted mb-1">
                  Moneda
                </label>
                <select
                  id="inv-currency"
                  value={currency}
                  onChange={(e) => setCurrency(e.target.value as 'ARS' | 'USD')}
                  className="w-full h-9 rounded-control border border-app-border bg-app-surface px-2 text-sm text-app-text focus:outline-none focus:ring-2 focus:ring-app-focus"
                >
                  <option value="ARS">ARS ($)</option>
                  <option value="USD">USD (US$)</option>
                </select>
              </div>
            </div>

            <div>
              <label htmlFor="inv-amount" className="block text-xs font-medium text-app-text-muted mb-1">
                Importe Total
              </label>
              <input
                id="inv-amount"
                type="text"
                placeholder="1000.00"
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
                className="w-full h-9 rounded-control border border-app-border bg-app-surface px-2.5 text-sm text-app-text tabular-nums focus:outline-none focus:ring-2 focus:ring-app-focus"
              />
            </div>

            <div className="grid grid-cols-2 gap-2.5">
              <div>
                <label htmlFor="inv-doc" className="block text-xs font-medium text-app-text-muted mb-1">
                  Doc. Receptor (opcional)
                </label>
                <input
                  id="inv-doc"
                  type="text"
                  placeholder="30-71122334-5"
                  value={receiverDoc}
                  onChange={(e) => setReceiverDoc(e.target.value)}
                  className="w-full h-9 rounded-control border border-app-border bg-app-surface px-2 text-sm text-app-text focus:outline-none focus:ring-2 focus:ring-app-focus"
                />
              </div>

              <div>
                <label htmlFor="inv-name" className="block text-xs font-medium text-app-text-muted mb-1">
                  Cliente / Razón Social
                </label>
                <input
                  id="inv-name"
                  type="text"
                  placeholder="Empresa S.A."
                  value={receiverName}
                  onChange={(e) => setReceiverName(e.target.value)}
                  className="w-full h-9 rounded-control border border-app-border bg-app-surface px-2 text-sm text-app-text focus:outline-none focus:ring-2 focus:ring-app-focus"
                />
              </div>
            </div>

            <div className="flex justify-end gap-2 pt-3 border-t border-app-border">
              <button
                type="button"
                onClick={() => setOpen(false)}
                className="h-9 px-3 rounded-control border border-app-border bg-app-surface text-app-text text-sm font-medium hover:bg-app-surface-muted"
              >
                Cancelar
              </button>
              <button
                type="submit"
                disabled={mutation.isPending}
                className="h-9 px-4 rounded-control bg-app-primary text-white text-sm font-medium hover:bg-app-primary-hover disabled:opacity-50"
              >
                {mutation.isPending ? 'Guardando...' : 'Guardar factura'}
              </button>
            </div>
          </form>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
