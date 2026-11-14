import { useState } from 'react';
import * as Dialog from '@radix-ui/react-dialog';
import { X, ShoppingBag } from 'lucide-react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { createPurchaseInvoice, PurchaseInvoicePayload } from '../api';

interface Props {
  defaultDate?: string;
  onSuccess?: () => void;
}

export function PurchaseInvoiceForm({ defaultDate, onSuccess }: Props) {
  const [open, setOpen] = useState(false);
  const queryClient = useQueryClient();

  const [type, setType] = useState('A');
  const [pointOfSale, setPointOfSale] = useState('1');
  const [number, setNumber] = useState('1');
  const [issueDate, setIssueDate] = useState(defaultDate || new Date().toISOString().slice(0, 10));
  const [currency, setCurrency] = useState<'ARS' | 'USD'>('ARS');
  const [netAmount, setNetAmount] = useState('');
  const [vatAmount, setVatAmount] = useState('');
  const [supplierDoc, setSupplierDoc] = useState('');
  const [supplierName, setSupplierName] = useState('');
  const [error, setError] = useState<string | null>(null);

  const mutation = useMutation({
    mutationFn: createPurchaseInvoice,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['my-closings'] });
      setOpen(false);
      setError(null);
      onSuccess?.();
    },
    onError: (err: Error) => {
      setError(err.message || 'Error al guardar la factura de compra');
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const numNet = parseFloat(netAmount.replace(',', '.'));
    const numVat = parseFloat(vatAmount.replace(',', '.')) || 0;
    if (isNaN(numNet) || numNet <= 0) {
      setError('El importe neto debe ser mayor a 0');
      return;
    }
    const numTotal = numNet + numVat;

    const payload: PurchaseInvoicePayload = {
      type,
      pointOfSale: parseInt(pointOfSale, 10) || 1,
      number: parseInt(number, 10) || 1,
      issueDate: new Date(issueDate).toISOString(),
      currency,
      netAmount: { amount: Math.round(numNet * 100), currency },
      vatAmount: { amount: Math.round(numVat * 100), currency },
      totalAmount: { amount: Math.round(numTotal * 100), currency },
      supplierDoc: supplierDoc.trim(),
      supplierName: supplierName.trim(),
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
          <ShoppingBag className="w-3.5 h-3.5" aria-hidden="true" />
          Cargar factura de compra
        </button>
      </Dialog.Trigger>
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-50 bg-black/40 backdrop-blur-sm" />
        <Dialog.Content
          className="fixed left-1/2 top-1/2 z-50 w-full max-w-md -translate-x-1/2 -translate-y-1/2 rounded-card border border-app-border bg-app-surface p-6 shadow-xl"
          aria-describedby="purchase-invoice-desc"
        >
          <div className="flex items-center justify-between pb-3 border-b border-app-border">
            <Dialog.Title className="text-base font-semibold text-app-text">
              Cargar factura de compra (RI)
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

          <p id="purchase-invoice-desc" className="sr-only">
            Formulario para cargar comprobantes de compras o gastos para Responsables Inscriptos
          </p>

          <form onSubmit={handleSubmit} className="mt-4 space-y-3.5 text-sm">
            {error && (
              <div className="rounded-control bg-severity-red-bg border border-severity-red-border p-2 text-xs text-severity-red-text">
                {error}
              </div>
            )}

            <div className="grid grid-cols-3 gap-2.5">
              <div>
                <label htmlFor="pinv-type" className="block text-xs font-medium text-app-text-muted mb-1">
                  Tipo
                </label>
                <select
                  id="pinv-type"
                  value={type}
                  onChange={(e) => setType(e.target.value)}
                  className="w-full h-9 rounded-control border border-app-border bg-app-surface px-2 text-sm text-app-text focus:outline-none focus:ring-2 focus:ring-app-focus"
                >
                  <option value="A">Factura A</option>
                  <option value="B">Factura B</option>
                  <option value="C">Factura C</option>
                </select>
              </div>

              <div>
                <label htmlFor="pinv-pos" className="block text-xs font-medium text-app-text-muted mb-1">
                  Punto de venta
                </label>
                <input
                  id="pinv-pos"
                  type="number"
                  min="1"
                  value={pointOfSale}
                  onChange={(e) => setPointOfSale(e.target.value)}
                  className="w-full h-9 rounded-control border border-app-border bg-app-surface px-2 text-sm text-app-text focus:outline-none focus:ring-2 focus:ring-app-focus"
                />
              </div>

              <div>
                <label htmlFor="pinv-num" className="block text-xs font-medium text-app-text-muted mb-1">
                  Número
                </label>
                <input
                  id="pinv-num"
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
                <label htmlFor="pinv-date" className="block text-xs font-medium text-app-text-muted mb-1">
                  Fecha de emisión
                </label>
                <input
                  id="pinv-date"
                  type="date"
                  value={issueDate}
                  onChange={(e) => setIssueDate(e.target.value)}
                  className="w-full h-9 rounded-control border border-app-border bg-app-surface px-2 text-sm text-app-text focus:outline-none focus:ring-2 focus:ring-app-focus"
                />
              </div>

              <div>
                <label htmlFor="pinv-currency" className="block text-xs font-medium text-app-text-muted mb-1">
                  Moneda
                </label>
                <select
                  id="pinv-currency"
                  value={currency}
                  onChange={(e) => setCurrency(e.target.value as 'ARS' | 'USD')}
                  className="w-full h-9 rounded-control border border-app-border bg-app-surface px-2 text-sm text-app-text focus:outline-none focus:ring-2 focus:ring-app-focus"
                >
                  <option value="ARS">ARS ($)</option>
                  <option value="USD">USD (US$)</option>
                </select>
              </div>
            </div>

            <div className="grid grid-cols-2 gap-2.5">
              <div>
                <label htmlFor="pinv-net" className="block text-xs font-medium text-app-text-muted mb-1">
                  Neto Gravado
                </label>
                <input
                  id="pinv-net"
                  type="text"
                  placeholder="1000.00"
                  value={netAmount}
                  onChange={(e) => setNetAmount(e.target.value)}
                  className="w-full h-9 rounded-control border border-app-border bg-app-surface px-2.5 text-sm text-app-text tabular-nums focus:outline-none focus:ring-2 focus:ring-app-focus"
                />
              </div>

              <div>
                <label htmlFor="pinv-vat" className="block text-xs font-medium text-app-text-muted mb-1">
                  IVA
                </label>
                <input
                  id="pinv-vat"
                  type="text"
                  placeholder="210.00"
                  value={vatAmount}
                  onChange={(e) => setVatAmount(e.target.value)}
                  className="w-full h-9 rounded-control border border-app-border bg-app-surface px-2.5 text-sm text-app-text tabular-nums focus:outline-none focus:ring-2 focus:ring-app-focus"
                />
              </div>
            </div>

            <div className="grid grid-cols-2 gap-2.5">
              <div>
                <label htmlFor="pinv-doc" className="block text-xs font-medium text-app-text-muted mb-1">
                  CUIT Proveedor
                </label>
                <input
                  id="pinv-doc"
                  type="text"
                  placeholder="30-71122334-5"
                  value={supplierDoc}
                  onChange={(e) => setSupplierDoc(e.target.value)}
                  className="w-full h-9 rounded-control border border-app-border bg-app-surface px-2 text-sm text-app-text focus:outline-none focus:ring-2 focus:ring-app-focus"
                />
              </div>

              <div>
                <label htmlFor="pinv-name" className="block text-xs font-medium text-app-text-muted mb-1">
                  Proveedor / Razón Social
                </label>
                <input
                  id="pinv-name"
                  type="text"
                  placeholder="Distribuidora S.A."
                  value={supplierName}
                  onChange={(e) => setSupplierName(e.target.value)}
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
                {mutation.isPending ? 'Guardando...' : 'Guardar compra'}
              </button>
            </div>
          </form>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
