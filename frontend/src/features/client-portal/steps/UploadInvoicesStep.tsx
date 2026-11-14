import React, { useState, useRef } from 'react';
import { UploadCloud, FileText, CheckCircle2, AlertTriangle, AlertCircle } from 'lucide-react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { importInvoicesCSV } from '../api';
import { ImportResult } from '../schemas';
import { ComingSoonButton } from '../../../shared/ui/ComingSoonButton';
import { ManualInvoiceForm } from '../forms/ManualInvoiceForm';
import { PurchaseInvoiceForm } from '../forms/PurchaseInvoiceForm';

interface Props {
  periodMonthName: string;
  regime: string;
  invoicesCount: number;
}

export function UploadInvoicesStep({ periodMonthName, regime, invoicesCount }: Props) {
  const [isDragging, setIsDragging] = useState(false);
  const [importResult, setImportResult] = useState<ImportResult | null>(null);
  const [uploadError, setUploadError] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const queryClient = useQueryClient();

  const isRI = regime === 'RI';

  const mutation = useMutation({
    mutationFn: importInvoicesCSV,
    onSuccess: (data) => {
      setImportResult(data);
      setUploadError(null);
      queryClient.invalidateQueries({ queryKey: ['my-closings'] });
    },
    onError: (err: Error) => {
      setUploadError(err.message || 'Error al procesar el archivo CSV');
      setImportResult(null);
    },
  });

  const handleFiles = (files: FileList | null) => {
    if (!files || files.length === 0) return;
    const file = files[0];
    if (!file) return;
    if (!file.name.toLowerCase().endsWith('.csv')) {
      setUploadError('Por favor seleccioná un archivo CSV válido.');
      return;
    }
    mutation.mutate(file);
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
    handleFiles(e.dataTransfer.files);
  };

  return (
    <div className="space-y-3.5">
      <div className="flex items-center justify-between flex-wrap gap-2">
        <h3 className="text-sm font-semibold text-app-text">
          1. Subí tus facturas de {periodMonthName}
        </h3>
        <ComingSoonButton size="sm">Emitir factura en ARCA</ComingSoonButton>
      </div>

      {/* Zona de arrastre */}
      <div
        onDragOver={(e) => {
          e.preventDefault();
          setIsDragging(true);
        }}
        onDragLeave={() => setIsDragging(false)}
        onDrop={handleDrop}
        onClick={() => fileInputRef.current?.click()}
        role="button"
        tabIndex={0}
        onKeyDown={(e) => {
          if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault();
            fileInputRef.current?.click();
          }
        }}
        className={`group relative flex flex-col items-center justify-center p-6 rounded-card border-2 border-dashed transition-all cursor-pointer text-center ${
          isDragging
            ? 'border-app-primary bg-app-primary-soft/40'
            : 'border-app-border hover:border-app-primary/60 bg-app-surface-muted/30 hover:bg-app-surface-muted/60'
        }`}
      >
        <input
          ref={fileInputRef}
          type="file"
          accept=".csv"
          className="hidden"
          onChange={(e) => handleFiles(e.target.files)}
          data-testid="csv-file-input"
        />

        <div className="flex h-10 w-10 items-center justify-center rounded-full bg-app-primary-soft text-app-primary mb-2.5 transition-transform group-hover:scale-105">
          <UploadCloud className="h-5 w-5" aria-hidden="true" />
        </div>

        <p className="text-sm font-medium text-app-text">
          {mutation.isPending ? 'Procesando archivo CSV...' : 'Arrastrá el CSV de Mis Comprobantes acá'}
        </p>
        <p className="text-xs text-app-text-muted mt-1">
          o hacé clic para seleccionarlo desde tu dispositivo
        </p>
      </div>

      {/* Enlaces y acciones secundarias */}
      <div className="flex items-center justify-between flex-wrap gap-3 text-xs pt-1">
        <div className="flex items-center gap-3">
          <a
            href="/samples/mis_comprobantes_ok.csv"
            download="mis_comprobantes_ejemplo.csv"
            className="text-app-primary hover:underline font-medium inline-flex items-center gap-1"
          >
            <FileText className="w-3.5 h-3.5" aria-hidden="true" />
            Descargar CSV de ejemplo
          </a>
          <span className="text-app-border">•</span>
          <ManualInvoiceForm regime={regime} />
          {isRI && (
            <>
              <span className="text-app-border">•</span>
              <PurchaseInvoiceForm />
            </>
          )}
        </div>

        {invoicesCount > 0 && (
          <span className="text-app-text-muted font-medium">
            {invoicesCount} factura{invoicesCount > 1 ? 's' : ''} en el período
          </span>
        )}
      </div>

      {/* Errores de carga */}
      {uploadError && (
        <div className="rounded-control bg-severity-red-bg border border-severity-red-border p-3 text-xs text-severity-red-text flex items-start gap-2">
          <AlertCircle className="w-4 h-4 flex-shrink-0 mt-0.5" aria-hidden="true" />
          <span>{uploadError}</span>
        </div>
      )}

      {/* Resultado de importación */}
      {importResult && (
        <div className="rounded-control border border-app-border bg-app-surface p-3.5 space-y-2 text-xs">
          <div className="flex items-center gap-2 font-medium">
            <CheckCircle2 className="w-4 h-4 text-severity-green-icon flex-shrink-0" />
            <span className="text-app-text">
              {importResult.imported} comprobante{importResult.imported === 1 ? '' : 's'} importado{importResult.imported === 1 ? '' : 's'}
              {importResult.skippedDuplicates > 0 && ` (${importResult.skippedDuplicates} duplicados omitidos)`}
            </span>
          </div>

          {importResult.warnings.length > 0 && (
            <div className="space-y-1 text-severity-yellow-text pt-1">
              <div className="flex items-center gap-1.5 font-medium">
                <AlertTriangle className="w-3.5 h-3.5 flex-shrink-0" />
                <span>Advertencias:</span>
              </div>
              <ul className="list-disc pl-5 space-y-0.5">
                {importResult.warnings.map((w, idx) => (
                  <li key={idx}>Línea {w.line}: {w.message}</li>
                ))}
              </ul>
            </div>
          )}

          {importResult.errors.length > 0 && (
            <div className="space-y-1 text-severity-red-text pt-1">
              <div className="flex items-center gap-1.5 font-medium">
                <AlertCircle className="w-3.5 h-3.5 flex-shrink-0" />
                <span>Errores en líneas omitidas:</span>
              </div>
              <ul className="list-disc pl-5 space-y-0.5">
                {importResult.errors.map((e, idx) => (
                  <li key={idx}>Línea {e.line}: {e.message}</li>
                ))}
              </ul>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
