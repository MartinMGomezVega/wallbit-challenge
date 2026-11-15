import { useState } from 'react';
import { createClientInvoice } from '../api';
import { ApiError } from '../../../shared/http/http';

interface InvoiceFormDialogProps {
  isOpen: boolean;
  onClose: () => void;
  clientId: string;
  clientRegime: string;
  onSuccess: () => void;
}

export function InvoiceFormDialog({
  isOpen,
  onClose,
  clientId,
  clientRegime,
  onSuccess,
}: InvoiceFormDialogProps) {
  const isRI = clientRegime === 'RI' || clientRegime === 'RESPONSABLE_INSCRIPTO';

  const [issueDate, setIssueDate] = useState('2026-10-20');
  const [type, setType] = useState('E');
  const [pointOfSale, setPointOfSale] = useState('1');
  const [number, setNumber] = useState('');
  const [receiverName, setReceiverName] = useState('');
  const [receiverDoc, setReceiverDoc] = useState('');
  const [currency, setCurrency] = useState<'USD' | 'ARS'>('USD');
  const [totalAmount, setTotalAmount] = useState('');
  const [fxRate, setFxRate] = useState('');

  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});
  const [generalError, setGeneralError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  if (!isOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setFieldErrors({});
    setGeneralError(null);

    const posNum = parseInt(pointOfSale, 10);
    const numNum = parseInt(number, 10);
    const amtNum = parseFloat(totalAmount);

    const localErrors: Record<string, string> = {};
    if (!issueDate) localErrors.issueDate = 'La fecha de emisión es requerida';
    if (isNaN(posNum) || posNum <= 0) localErrors.pointOfSale = 'El punto de venta debe ser mayor a 0';
    if (isNaN(numNum) || numNum <= 0) localErrors.number = 'El número debe ser mayor a 0';
    if (!receiverName.trim()) localErrors.receiverName = 'La denominación del receptor es requerida';
    if (isNaN(amtNum) || amtNum <= 0) localErrors.totalAmount = 'El importe total debe ser mayor a 0';

    if (Object.keys(localErrors).length > 0) {
      setFieldErrors(localErrors);
      return;
    }

    setLoading(true);
    try {
      await createClientInvoice(clientId, {
        type,
        pointOfSale: posNum,
        number: numNum,
        issueDate,
        currency,
        netAmount: { amount: totalAmount.trim(), currency },
        vatAmount: { amount: '0.00', currency },
        totalAmount: { amount: totalAmount.trim(), currency },
        fxRate: fxRate.trim() || undefined,
        receiverName: receiverName.trim(),
        receiverDoc: receiverDoc.trim() || undefined,
      });
      onSuccess();
      onClose();
    } catch (err: unknown) {
      if (err instanceof ApiError) {
        if (err.status === 409 || err.code === 'INVOICE_ALREADY_EXISTS') {
          setGeneralError('Ese comprobante ya está cargado');
        } else if (err.code === 'VALIDATION_ERROR' && err.fields) {
          setFieldErrors(err.fields);
        } else {
          setGeneralError(err.message || 'Error al guardar la factura');
        }
      } else {
        setGeneralError('Error al guardar la factura');
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div
      role="dialog"
      aria-modal="true"
      aria-labelledby="invoice-dialog-title"
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4"
    >
      <div className="w-full max-w-sm rounded-2xl bg-white p-5 shadow-2xl text-[14px] text-neutral-900 max-h-[90vh] overflow-y-auto">
        <div className="flex items-center justify-between border-b border-neutral-200 pb-3">
          <h2 id="invoice-dialog-title" className="text-base font-bold text-neutral-900">
            Alta de comprobante
          </h2>
          <button
            type="button"
            onClick={onClose}
            className="rounded p-1 text-neutral-500 hover:text-neutral-800"
            aria-label="Cerrar formulario"
          >
            ✕
          </button>
        </div>

        {generalError && (
          <div
            role="alert"
            className="mt-3 rounded-lg bg-red-50 p-2.5 text-xs font-medium text-red-700 border border-red-200"
          >
            {generalError}
          </div>
        )}

        <form onSubmit={handleSubmit} className="mt-4 space-y-3">
          {/* Fecha y Tipo */}
          <div className="grid grid-cols-2 gap-2">
            <div>
              <label htmlFor="inv-date" className="block text-xs font-medium text-neutral-700">
                Fecha
              </label>
              <input
                id="inv-date"
                type="date"
                value={issueDate}
                onChange={e => setIssueDate(e.target.value)}
                className="mt-1 w-full rounded-lg border border-neutral-300 px-2.5 py-1.5 text-xs focus:border-black focus:outline-none"
              />
              {fieldErrors.issueDate && (
                <p className="mt-0.5 text-[11px] text-red-600">{fieldErrors.issueDate}</p>
              )}
            </div>

            <div>
              <label htmlFor="inv-type" className="block text-xs font-medium text-neutral-700">
                Tipo
              </label>
              <select
                id="inv-type"
                value={type}
                onChange={e => setType(e.target.value)}
                className="mt-1 w-full rounded-lg border border-neutral-300 px-2 py-1.5 text-xs focus:border-black focus:outline-none bg-white"
              >
                <option value="E">Factura E</option>
                {isRI ? (
                  <>
                    <option value="A">Factura A</option>
                    <option value="B">Factura B</option>
                  </>
                ) : (
                  <option value="C">Factura C</option>
                )}
              </select>
              {fieldErrors.type && (
                <p className="mt-0.5 text-[11px] text-red-600">{fieldErrors.type}</p>
              )}
            </div>
          </div>

          {/* Punto de venta y Número */}
          <div className="grid grid-cols-2 gap-2">
            <div>
              <label htmlFor="inv-pos" className="block text-xs font-medium text-neutral-700">
                Punto de venta
              </label>
              <input
                id="inv-pos"
                type="number"
                min="1"
                value={pointOfSale}
                onChange={e => setPointOfSale(e.target.value)}
                className="mt-1 w-full rounded-lg border border-neutral-300 px-2.5 py-1.5 text-xs focus:border-black focus:outline-none"
              />
              {fieldErrors.pointOfSale && (
                <p className="mt-0.5 text-[11px] text-red-600">{fieldErrors.pointOfSale}</p>
              )}
            </div>

            <div>
              <label htmlFor="inv-number" className="block text-xs font-medium text-neutral-700">
                Número
              </label>
              <input
                id="inv-number"
                type="number"
                min="1"
                placeholder="Ej. 101"
                value={number}
                onChange={e => setNumber(e.target.value)}
                className="mt-1 w-full rounded-lg border border-neutral-300 px-2.5 py-1.5 text-xs focus:border-black focus:outline-none"
              />
              {fieldErrors.number && (
                <p className="mt-0.5 text-[11px] text-red-600">{fieldErrors.number}</p>
              )}
            </div>
          </div>

          {/* Cliente (receptor) */}
          <div>
            <label htmlFor="inv-receiver-name" className="block text-xs font-medium text-neutral-700">
              Cliente (denominación)
            </label>
            <input
              id="inv-receiver-name"
              type="text"
              placeholder="Ej. Acme Inc."
              value={receiverName}
              onChange={e => setReceiverName(e.target.value)}
              className="mt-1 w-full rounded-lg border border-neutral-300 px-2.5 py-1.5 text-xs focus:border-black focus:outline-none"
            />
            {fieldErrors.receiverName && (
              <p className="mt-0.5 text-[11px] text-red-600">{fieldErrors.receiverName}</p>
            )}
          </div>

          {/* Documento receptor */}
          <div>
            <label htmlFor="inv-receiver-doc" className="block text-xs font-medium text-neutral-700">
              Documento del cliente <span className="text-neutral-400 font-normal">(opcional)</span>
            </label>
            <input
              id="inv-receiver-doc"
              type="text"
              placeholder="CUIT o Tax ID exterior"
              value={receiverDoc}
              onChange={e => setReceiverDoc(e.target.value)}
              className="mt-1 w-full rounded-lg border border-neutral-300 px-2.5 py-1.5 text-xs focus:border-black focus:outline-none"
            />
          </div>

          {/* Moneda y Monto total */}
          <div className="grid grid-cols-3 gap-2">
            <div>
              <label htmlFor="inv-currency" className="block text-xs font-medium text-neutral-700">
                Moneda
              </label>
              <select
                id="inv-currency"
                value={currency}
                onChange={e => setCurrency(e.target.value as 'USD' | 'ARS')}
                className="mt-1 w-full rounded-lg border border-neutral-300 px-2 py-1.5 text-xs focus:border-black focus:outline-none bg-white"
              >
                <option value="USD">USD</option>
                <option value="ARS">ARS</option>
              </select>
            </div>

            <div className="col-span-2">
              <label htmlFor="inv-amount" className="block text-xs font-medium text-neutral-700">
                Monto total
              </label>
              <input
                id="inv-amount"
                type="text"
                placeholder="Ej. 1500.00"
                value={totalAmount}
                onChange={e => setTotalAmount(e.target.value)}
                className="mt-1 w-full rounded-lg border border-neutral-300 px-2.5 py-1.5 text-xs focus:border-black focus:outline-none"
              />
              {fieldErrors.totalAmount && (
                <p className="mt-0.5 text-[11px] text-red-600">{fieldErrors.totalAmount}</p>
              )}
            </div>
          </div>

          {/* Tipo de cambio */}
          <div>
            <label htmlFor="inv-fx" className="block text-xs font-medium text-neutral-700">
              Tipo de cambio <span className="text-neutral-400 font-normal">(opcional, si figura en la factura)</span>
            </label>
            <input
              id="inv-fx"
              type="text"
              placeholder="Ej. 1482.15"
              value={fxRate}
              onChange={e => setFxRate(e.target.value)}
              className="mt-1 w-full rounded-lg border border-neutral-300 px-2.5 py-1.5 text-xs focus:border-black focus:outline-none"
            />
          </div>

          <div className="pt-2 flex justify-end gap-2">
            <button
              type="button"
              onClick={onClose}
              disabled={loading}
              className="rounded-xl border border-neutral-300 px-3 py-2 text-xs font-medium text-neutral-700 hover:bg-neutral-50"
            >
              Cancelar
            </button>
            <button
              type="submit"
              disabled={loading}
              className="rounded-xl bg-neutral-900 px-4 py-2 text-xs font-medium text-white hover:bg-neutral-800 disabled:opacity-50"
            >
              {loading ? 'Guardando...' : 'Cargar comprobante'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

