import { useState, useRef, DragEvent, ChangeEvent } from 'react';
import * as Dialog from '@radix-ui/react-dialog';
import { UploadCloud, FileText, AlertTriangle, X, Download } from 'lucide-react';
import { useImportInvoices } from './useImportInvoices';
import { ImportResultView } from './ImportResultView';
import { ImportResultDTO } from '../schemas';

export interface ImportInvoicesDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  clientId: string;
  closingId: string;
  closingStatus: string;
  onImportFinished?: (importedCount: number) => void;
}

const MAX_FILE_SIZE_BYTES = 2 * 1024 * 1024; // 2 MB

export function ImportInvoicesDialog({
  open,
  onOpenChange,
  clientId,
  closingId,
  closingStatus,
  onImportFinished,
}: ImportInvoicesDialogProps) {
  const [file, setFile] = useState<File | null>(null);
  const [fileError, setFileError] = useState<string | null>(null);
  const [isDragging, setIsDragging] = useState(false);
  const [result, setResult] = useState<ImportResultDTO | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const isApprovedOrNotified = closingStatus === 'APPROVED' || closingStatus === 'NOTIFIED';

  const importMutation = useImportInvoices(clientId, closingId, {
    onSuccess: (data) => {
      setResult(data);
    },
    onError: (err) => {
      const msg = err instanceof Error ? err.message : 'Error al importar facturas';
      setFileError(msg);
    },
  });

  const resetState = () => {
    setFile(null);
    setFileError(null);
    setResult(null);
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  };

  const handleOpenChange = (nextOpen: boolean) => {
    if (!nextOpen) {
      if (result && result.imported > 0) {
        onImportFinished?.(result.imported);
      }
      resetState();
    }
    onOpenChange(nextOpen);
  };

  const validateAndSelectFile = (candidate: File | null) => {
    setFileError(null);
    setResult(null);
    if (!candidate) {
      setFile(null);
      return;
    }
    if (!candidate.name.toLowerCase().endsWith('.csv')) {
      setFileError('El archivo debe tener formato CSV (.csv)');
      setFile(null);
      return;
    }
    if (candidate.size > MAX_FILE_SIZE_BYTES) {
      setFileError('El archivo supera el tamaño máximo permitido de 2 MB');
      setFile(null);
      return;
    }
    setFile(candidate);
  };

  const handleFileChange = (e: ChangeEvent<HTMLInputElement>) => {
    const selected = e.target.files?.[0] || null;
    validateAndSelectFile(selected);
  };

  const handleDragOver = (e: DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    setIsDragging(true);
  };

  const handleDragLeave = () => {
    setIsDragging(false);
  };

  const handleDrop = (e: DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    setIsDragging(false);
    const dropped = e.dataTransfer.files?.[0] || null;
    validateAndSelectFile(dropped);
  };

  const handleImport = async () => {
    if (!file) return;
    setFileError(null);
    importMutation.mutate(file);
  };

  return (
    <Dialog.Root open={open} onOpenChange={handleOpenChange}>
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-50 bg-[#0A0A0A]/50 backdrop-blur-xs transition-opacity" />
        <Dialog.Content
          aria-describedby="import-invoices-desc"
          className="fixed left-1/2 top-1/2 z-50 w-full max-w-[520px] -translate-x-1/2 -translate-y-1/2 rounded-[10px] border border-app-border bg-app-surface p-6 shadow-xl focus:outline-none"
        >
          <div className="flex items-start justify-between gap-4 border-b border-app-border pb-4">
            <div className="flex items-center gap-2.5">
              <div className="flex h-9 w-9 items-center justify-center rounded-full bg-app-primary-soft text-app-primary">
                <FileText className="h-5 w-5" aria-hidden="true" />
              </div>
              <div>
                <Dialog.Title className="text-base font-semibold text-app-text">
                  Importar facturas de venta (CSV de ARCA)
                </Dialog.Title>
                <Dialog.Description id="import-invoices-desc" className="text-xs text-app-text-muted">
                  Carga el reporte de «Mis Comprobantes Emitidos» descargado de ARCA.
                </Dialog.Description>
              </div>
            </div>
            <Dialog.Close asChild>
              <button
                type="button"
                className="rounded-control p-1 text-app-text-muted hover:bg-app-surface-muted hover:text-app-text focus:outline-none focus:ring-2 focus:ring-app-primary"
                aria-label="Cerrar ventana"
              >
                <X className="h-5 w-5" aria-hidden="true" />
              </button>
            </Dialog.Close>
          </div>

          <div className="space-y-4 py-4">
            {/* Aviso si ya está aprobado o notificado */}
            {isApprovedOrNotified && (
              <div
                role="alert"
                className="flex items-start gap-2.5 rounded-control border border-severity-yellow-border bg-severity-yellow-bg p-3 text-xs text-severity-yellow-text"
              >
                <AlertTriangle className="h-4 w-4 flex-shrink-0 mt-0.5 text-severity-yellow-icon" aria-hidden="true" />
                <div>
                  <span className="font-semibold block mb-0.5">Atención: Cierre previamente aprobado</span>
                  Este cierre ya fue aprobado. Si importás facturas, vuelve a «Para revisar» con una versión nueva.
                </div>
              </div>
            )}

            {/* Zona de soltar archivo / File input */}
            <div
              onDragOver={handleDragOver}
              onDragLeave={handleDragLeave}
              onDrop={handleDrop}
              onClick={() => fileInputRef.current?.click()}
              className={`flex flex-col items-center justify-center rounded-control border-2 border-dashed p-6 text-center cursor-pointer transition-colors ${
                isDragging
                  ? 'border-app-primary bg-app-primary-soft/30'
                  : 'border-app-border bg-app-surface-muted hover:border-app-primary/50'
              }`}
            >
              <input
                ref={fileInputRef}
                type="file"
                accept=".csv"
                className="hidden"
                onChange={handleFileChange}
              />
              <UploadCloud className="h-8 w-8 text-app-text-muted mb-2" aria-hidden="true" />
              <p className="text-xs font-medium text-app-text">
                {file ? file.name : 'Arrastrá tu archivo CSV acá o hacé clic para buscarlo'}
              </p>
              <p className="text-[11px] text-app-text-muted mt-1">Formato CSV de ARCA, hasta 2 MB</p>
            </div>

            {/* Link al CSV de ejemplo */}
            <div className="flex items-center justify-between text-xs">
              <a
                href="/samples/mis_comprobantes_ok.csv"
                download
                className="inline-flex items-center gap-1.5 text-app-primary hover:underline font-medium focus:outline-none focus:ring-2 focus:ring-app-primary rounded px-1"
              >
                <Download className="h-3.5 w-3.5" aria-hidden="true" />
                <span>Descargar CSV de ejemplo</span>
              </a>
              {file && (
                <span className="text-app-text-muted text-[11px] font-mono">
                  {(file.size / 1024).toFixed(1)} KB
                </span>
              )}
            </div>

            {/* Error de validación local o API */}
            {fileError && (
              <div
                role="alert"
                className="rounded-control border border-severity-red-border bg-severity-red-bg p-3 text-xs text-severity-red-text"
              >
                {fileError}
              </div>
            )}

            {/* Resultado de la importación */}
            {result && <ImportResultView result={result} />}
          </div>

          <div className="flex items-center justify-end gap-3 border-t border-app-border pt-4">
            <button
              type="button"
              onClick={() => handleOpenChange(false)}
              className="rounded-control border border-app-border px-4 py-2 text-xs font-semibold text-app-text hover:bg-app-surface-muted focus:outline-none focus:ring-2 focus:ring-app-primary"
            >
              {result ? 'Listo' : 'Cancelar'}
            </button>
            <button
              type="button"
              disabled={!file || importMutation.isPending}
              onClick={handleImport}
              className="rounded-control bg-app-primary px-4 py-2 text-xs font-semibold text-white hover:bg-app-primary-hover focus:outline-none focus:ring-2 focus:ring-app-primary disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {importMutation.isPending ? 'Importando...' : 'Importar'}
            </button>
          </div>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}

