import { useState, useRef } from 'react';
import { ClientInvoice } from '../schemas';
import { InvoiceFormDialog } from './InvoiceFormDialog';
import { importClientInvoicesCSV } from '../api';
import { ImportResult } from '../schemas';
import { ApiError } from '../../../shared/api/http';

interface InvoicesStepProps {
  clientId: string;
  clientRegime: string;
  invoices: ClientInvoice[];
  onRefresh: () => void;
}

export function InvoicesStep({
  clientId,
  clientRegime,
  invoices,
  onRefresh,
}: InvoicesStepProps) {
  const [dialogOpen, setDialogOpen] = useState(false);
  const [importing, setImporting] = useState(false);
  const [importResult, setImportResult] = useState<ImportResult | null>(null);
  const [importError, setImportError] = useState<string | null>(null);

  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    setImporting(true);
    setImportResult(null);
    setImportError(null);

    try {
      const res = await importClientInvoicesCSV(clientId, file);
      setImportResult(res);
      onRefresh();
    } catch (err: unknown) {
      if (err instanceof ApiError) {
        setImportError(err.message || 'Error al importar archivo CSV');
      } else {
        setImportError('Error al importar archivo CSV');
      }
    } finally {
      setImporting(false);
      if (fileInputRef.current) fileInputRef.current.value = '';
    }
  };

  const formatAmount = (inv: ClientInvoice) => {
    const symbol = inv.currency === 'USD' ? 'US$' : '$';
    const num = parseFloat(inv.totalAmount.amount);
    const formatted = isNaN(num)
      ? inv.totalAmount.amount
      : num.toLocaleString('es-AR', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
    return `${symbol} ${formatted}`;
  };

  const formatDate = (iso: string) => {
    if (!iso) return '-';
    const firstPart = iso.split('T')[0];
    if (!firstPart) return '-';
    const parts = firstPart.split('-');
    if (parts.length === 3) {
      return `${parts[2]}/${parts[1]}`;
    }
    return iso;
  };

  return (
    <div className="rounded-2xl border border-neutral-200/80 bg-white p-4 shadow-sm space-y-3">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-bold text-neutral-900">
          Comprobantes que emitiste en octubre
        </h3>
        <span className="text-xs font-semibold text-neutral-500">
          {invoices.length} {invoices.length === 1 ? 'comprobante' : 'comprobantes'}
        </span>
      </div>

      {/* Table of invoices */}
      {invoices.length === 0 ? (
        <div className="rounded-xl bg-[#F9F9F9] p-3 text-center text-xs text-neutral-500">
          No hay comprobantes cargados para este período.
        </div>
      ) : (
        <div className="overflow-x-auto rounded-xl border border-neutral-100">
          <table className="w-full text-left text-xs" aria-label="Comprobantes emitidos">
            <thead className="bg-[#F9F9F9] text-[11px] font-semibold text-neutral-500 uppercase tracking-wider">
              <tr>
                <th scope="col" className="px-2.5 py-2">Fecha</th>
                <th scope="col" className="px-2 py-2">Tipo</th>
                <th scope="col" className="px-2 py-2">Nro.</th>
                <th scope="col" className="px-2 py-2">Cliente</th>
                <th scope="col" className="px-2.5 py-2 text-right">Monto</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-neutral-100 text-neutral-700">
              {invoices.map(inv => {
                const posStr = String(inv.pointOfSale).padStart(4, '0');
                const numStr = String(inv.number).padStart(8, '0');
                return (
                  <tr key={inv.id} className="hover:bg-neutral-50/50">
                    <td className="px-2.5 py-2 whitespace-nowrap text-neutral-500 font-mono text-[11px]">
                      {formatDate(inv.issueDate)}
                    </td>
                    <td className="px-2 py-2 whitespace-nowrap font-medium text-neutral-800">
                      Fact. {inv.type}
                    </td>
                    <td className="px-2 py-2 whitespace-nowrap font-mono text-[11px] text-neutral-600">
                      {posStr}-{numStr}
                    </td>
                    <td className="px-2 py-2 truncate max-w-[90px]" title={inv.receiverName}>
                      {inv.receiverName}
                    </td>
                    <td className="px-2.5 py-2 text-right font-medium text-neutral-900 whitespace-nowrap tabular-nums">
                      {formatAmount(inv)}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}

      {/* Import feedback */}
      {importResult && (
        <div
          role="status"
          className="rounded-xl bg-green-50 p-2.5 text-xs text-green-800 border border-green-200"
        >
          ✓ Se importaron {importResult.imported} comprobantes
          {importResult.skippedDuplicates > 0 && ` (${importResult.skippedDuplicates} duplicados omitidos)`}.
        </div>
      )}
      {importError && (
        <div
          role="alert"
          className="rounded-xl bg-red-50 p-2.5 text-xs text-red-700 border border-red-200"
        >
          {importError}
        </div>
      )}

      {/* Actions */}
      <div className="flex flex-col gap-2 pt-1">
        <button
          type="button"
          onClick={() => setDialogOpen(true)}
          className="flex h-10 w-full items-center justify-center rounded-xl bg-neutral-900 px-4 text-xs font-semibold text-white transition-colors hover:bg-neutral-800 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-black"
        >
          Alta de comprobante
        </button>

        <div className="text-center">
          <input
            ref={fileInputRef}
            type="file"
            accept=".csv"
            onChange={handleFileChange}
            className="hidden"
            id="csv-file-input"
            aria-label="Seleccionar archivo CSV de Mis Comprobantes"
          />
          <button
            type="button"
            disabled={importing}
            onClick={() => fileInputRef.current?.click()}
            className="text-xs font-medium text-neutral-600 underline underline-offset-2 hover:text-neutral-900 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-black"
          >
            {importing ? 'Importando archivo CSV...' : 'Importar CSV de Mis Comprobantes'}
          </button>
        </div>
      </div>

      <InvoiceFormDialog
        isOpen={dialogOpen}
        onClose={() => setDialogOpen(false)}
        clientId={clientId}
        clientRegime={clientRegime}
        onSuccess={onRefresh}
      />
    </div>
  );
}
