import { useState } from 'react';
import * as Dialog from '@radix-ui/react-dialog';
import { useQuery } from '@tanstack/react-query';
import { useReviewQueue } from './useReviewQueue';
import { useKeyboardNav } from './useKeyboardNav';
import { useApprove } from './useApprove';
import { getClosingDetail } from './api';
import { ReviewHeader } from './ReviewHeader';
import { SummarySection } from './sections/SummarySection';
import { SignalsSection } from './sections/SignalsSection';
import { MovementsSection } from './sections/MovementsSection';
import { InvoicesSection } from './sections/InvoicesSection';
import { ClientMessageSection } from './sections/ClientMessageSection';
import { HistorySection } from './sections/HistorySection';
import { ConfirmRegimeModal } from './ConfirmRegimeModal';
import { SeverityIndicator } from '../../shared/ui/SeverityIndicator';
import { StatusBadge } from '../../shared/ui/StatusBadge';
import { AlertCircle, AlertTriangle, Loader2 } from 'lucide-react';

export type ReviewTabKey =
  | 'resumen'
  | 'senales'
  | 'movimientos'
  | 'facturas'
  | 'aviso'
  | 'historial';

export function ReviewModal() {
  const {
    queue,
    isOpen,
    currentClosingId,
    canGoNext,
    canGoPrev,
    progress,
    isOutsideFilter,
    closeQueue,
    goNext,
    goPrev,
    markCurrentReviewed,
  } = useReviewQueue();

  const [activeTab, setActiveTab] = useState<ReviewTabKey>('resumen');
  const [clientNote, setClientNote] = useState('');
  const [confirmRegimeOpen, setConfirmRegimeOpen] = useState(false);
  const [conflictError, setConflictError] = useState(false);
  const [generalError, setGeneralError] = useState<string | null>(null);

  // Consulta de detalle del cliente activo
  const {
    data: detail,
    isLoading,
    isError,
    error,
    refetch,
  } = useQuery({
    queryKey: ['closing-detail', currentClosingId],
    queryFn: () => getClosingDetail(currentClosingId!),
    enabled: Boolean(isOpen && currentClosingId),
  });

  const approveMutation = useApprove({
    onError: (err) => {
      const apiErr = err as { code?: string; status?: number; message?: string };
      if (apiErr.code === 'VERSION_CONFLICT' || apiErr.status === 409) {
        setConflictError(true);
      } else {
        setGeneralError(apiErr.message || 'Error al aprobar');
      }
    },
  });

  const handleApprove = (andNext: boolean) => {
    if (!currentClosingId || !detail) return;
    setConflictError(false);
    setGeneralError(null);

    approveMutation.mutate(
      {
        closingId: currentClosingId,
        version: detail.closing.version,
        clientNote: clientNote.trim() || undefined,
      },
      {
        onSuccess: () => {
          markCurrentReviewed();
          setClientNote('');
          if (andNext) {
            if (canGoNext) {
              goNext();
            } else {
              closeQueue();
            }
          }
        },
      }
    );
  };

  // Atajos de teclado (Criterio 5 y Criterio 12 de spec 010: atajo A)
  useKeyboardNav({
    enabled: isOpen && !confirmRegimeOpen,
    onNext: canGoNext ? goNext : undefined,
    onPrev: canGoPrev ? goPrev : undefined,
    onClose: closeQueue,
    onApprove: () => {
      if (!approveMutation.isPending) {
        handleApprove(true);
      }
    },
  });

  if (!isOpen) return null;

  const total = queue?.ids.length ?? 0;
  const currentIndex = queue?.index ?? 0;
  const filterLabel = queue?.label ?? 'Filtro actual';

  const hasExclusionK = Boolean(
    detail?.closing.signals?.some((s) => s.code === 'MONO_EXCLUSION_K_CAP_EXCEEDED')
  );

  const effectiveExclusionDate =
    detail?.closing.mono?.exclusionEffectiveDate ||
    (detail?.closing.signals?.find((s) => s.code === 'MONO_EXCLUSION_K_CAP_EXCEEDED')?.data
      ?.exclusionDate as string) ||
    '2026-10-01';

  const tabs: { key: ReviewTabKey; label: string; count?: number }[] = [
    { key: 'resumen', label: 'Resumen' },
    {
      key: 'senales',
      label: 'Señales',
      count: detail?.closing.signals?.length,
    },
    {
      key: 'movimientos',
      label: 'Movimientos',
      count: detail?.periodMovements?.length,
    },
    {
      key: 'facturas',
      label: 'Facturas y conciliación',
      count: detail?.periodInvoices?.length,
    },
    { key: 'aviso', label: 'Aviso al cliente' },
    { key: 'historial', label: 'Historial' },
  ];

  return (
    <>
      <Dialog.Root open={isOpen} onOpenChange={(open) => !open && closeQueue()}>
        <Dialog.Portal>
          <Dialog.Overlay className="fixed inset-0 z-40 bg-[#0A0A0A]/50 backdrop-blur-xs transition-opacity" />
          <Dialog.Content
            aria-labelledby="review-modal-title"
            className="fixed left-1/2 top-1/2 z-50 flex flex-col w-[min(1120px,94vw)] h-[88vh] -translate-x-1/2 -translate-y-1/2 rounded-[10px] border border-app-border bg-app-surface shadow-xl focus:outline-none overflow-hidden"
          >
            {/* Header de navegación con Anterior / Siguiente / Contador */}
            <ReviewHeader
              currentIndex={currentIndex}
              total={total}
              filterLabel={filterLabel}
              reviewedCount={progress.reviewedCount}
              canGoPrev={canGoPrev}
              canGoNext={canGoNext}
              onPrev={goPrev}
              onNext={goNext}
              onClose={closeQueue}
            />

            {/* Barra de progreso de revisión (2 px) */}
            <div className="h-0.5 w-full bg-app-border shrink-0">
              <div
                className="h-full bg-app-primary transition-all duration-300"
                style={{
                  width: total > 0 ? `${(progress.reviewedCount / total) * 100}%` : '0%',
                }}
              />
            </div>

            {/* Banner de aviso si quedó fuera del filtro o cambió de color (Criterio 2 y 6) */}
            {isOutsideFilter && (
              <div
                role="status"
                className="flex items-center gap-2 border-b border-severity-yellow-border bg-severity-yellow-bg px-6 py-2 text-xs font-medium text-severity-yellow-text shrink-0"
              >
                <AlertTriangle className="h-4 w-4 shrink-0 text-severity-yellow-icon" aria-hidden="true" />
                <span>Este cliente ya no cumple el filtro</span>
              </div>
            )}

            {/* Sub-header con datos del cliente y estado del cierre (Criterio 10) */}
            <div className="flex flex-wrap items-center justify-between gap-4 border-b border-app-border bg-app-surface px-6 py-3.5 shrink-0">
              <div className="flex items-center gap-3">
                {detail ? (
                  <SeverityIndicator severity={detail.closing.severity} />
                ) : (
                  <div className="h-4 w-4 rounded-full bg-app-surface-muted animate-pulse" />
                )}

                <div>
                  <div className="flex items-center gap-2">
                    <Dialog.Title
                      id="review-modal-title"
                      className="text-lg font-semibold text-app-text"
                    >
                      {detail?.client.name || 'Cargando...'}
                    </Dialog.Title>
                    {detail && (
                      <span className="rounded-control bg-app-surface-muted px-2 py-0.5 text-xs font-medium text-app-text-muted">
                        CUIT {detail.client.cuit}
                      </span>
                    )}
                  </div>
                </div>
              </div>

              {detail && (
                <div className="flex items-center gap-2 text-xs">
                  <span className="rounded-control border border-app-border bg-white px-2 py-0.5 font-medium text-app-text">
                    {detail.client.regime === 'MONOTRIBUTO'
                      ? `Monotributo Cat. ${detail.closing.mono?.currentCategory || detail.client.monotributoCategory || 'A'}`
                      : 'Responsable Inscripto'}
                  </span>
                  <StatusBadge status={detail.closing.status} paidAt={detail.closing.paidAt} />
                  <span className="text-app-text-muted">v{detail.closing.version}</span>
                </div>
              )}
            </div>

            {/* Barra de pestañas */}
            <div className="flex items-center border-b border-app-border bg-app-surface px-6 shrink-0 overflow-x-auto">
              {tabs.map((tab) => {
                const isActive = activeTab === tab.key;
                return (
                  <button
                    key={tab.key}
                    type="button"
                    onClick={() => setActiveTab(tab.key)}
                    className={`relative inline-flex items-center gap-1.5 py-3 px-3 text-xs font-medium transition-colors focus:outline-none focus:ring-2 focus:ring-app-focus ${
                      isActive
                        ? 'text-app-primary font-semibold'
                        : 'text-app-text-muted hover:text-app-text'
                    }`}
                  >
                    <span>{tab.label}</span>
                    {typeof tab.count === 'number' && tab.count > 0 && (
                      <span
                        className={`rounded-full px-1.5 py-0.2 text-[10px] font-bold ${
                          isActive
                            ? 'bg-app-primary-soft text-app-primary'
                            : 'bg-app-surface-muted text-app-text-muted'
                        }`}
                      >
                        {tab.count}
                      </span>
                    )}
                    {isActive && (
                      <span className="absolute bottom-0 left-0 right-0 h-0.5 bg-app-primary" />
                    )}
                  </button>
                );
              })}
            </div>

            {/* Cuerpo del modal con scroll interno */}
            <div className="flex-1 overflow-y-auto p-6 bg-app-bg/40">
              {conflictError && (
                <div
                  role="alert"
                  className="mb-4 flex items-center justify-between rounded-control border border-severity-yellow-border bg-severity-yellow-bg px-4 py-2.5 text-xs text-severity-yellow-text"
                >
                  <div className="flex items-center gap-2">
                    <AlertTriangle className="h-4 w-4 shrink-0 text-severity-yellow-icon" />
                    <span className="font-semibold">El cierre cambió mientras lo revisabas</span>
                  </div>
                  <button
                    type="button"
                    onClick={() => {
                      setConflictError(false);
                      refetch();
                    }}
                    className="rounded-control border border-severity-yellow-border bg-white px-2.5 py-1 text-xs font-semibold text-severity-yellow-text hover:bg-severity-yellow-bg"
                  >
                    Ver cambios
                  </button>
                </div>
              )}

              {generalError && (
                <div
                  role="alert"
                  className="mb-4 flex items-center justify-between rounded-control border border-severity-red-border bg-severity-red-bg px-4 py-2.5 text-xs text-severity-red-text"
                >
                  <div className="flex items-center gap-2">
                    <AlertCircle className="h-4 w-4 shrink-0 text-severity-red-icon" />
                    <span>{generalError}</span>
                  </div>
                  <button
                    type="button"
                    onClick={() => setGeneralError(null)}
                    className="text-xs font-semibold hover:underline"
                  >
                    Descartar
                  </button>
                </div>
              )}

              {isLoading && (
                <div className="flex flex-col items-center justify-center py-20 text-app-text-muted">
                  <Loader2 className="h-8 w-8 animate-spin text-app-primary mb-3" />
                  <p className="text-sm">Cargando detalle del cierre...</p>
                </div>
              )}

              {isError && (
                <div className="rounded-card border border-severity-red-border bg-severity-red-bg p-6 text-center text-severity-red-text">
                  <AlertCircle className="mx-auto h-8 w-8 mb-2" />
                  <h4 className="text-sm font-semibold">Error al cargar el detalle</h4>
                  <p className="text-xs mt-1">{(error as Error)?.message || 'Intente nuevamente'}</p>
                  <button
                    type="button"
                    onClick={() => refetch()}
                    className="mt-4 rounded-control border border-severity-red-border bg-white px-3 py-1.5 text-xs font-semibold text-severity-red-text hover:bg-severity-red-bg"
                  >
                    Reintentar
                  </button>
                </div>
              )}

              {!isLoading && !isError && detail && (
                <>
                  {activeTab === 'resumen' && <SummarySection detail={detail} />}
                  {activeTab === 'senales' && <SignalsSection detail={detail} />}
                  {activeTab === 'movimientos' && (
                    <MovementsSection detail={detail} onRefreshDetail={refetch} />
                  )}
                  {activeTab === 'facturas' && <InvoicesSection detail={detail} />}
                  {activeTab === 'aviso' && (
                    <ClientMessageSection detail={detail} clientNote={clientNote} />
                  )}
                  {activeTab === 'historial' && <HistorySection detail={detail} />}
                </>
              )}
            </div>

            {/* Footer sticky (64 px) */}
            <div className="flex items-center justify-between border-t border-app-border bg-app-surface px-6 py-3.5 shrink-0">
              {/* Izquierda: Nota para el cliente */}
              <div className="flex items-center gap-2 w-full max-w-sm">
                <input
                  type="text"
                  placeholder="Nota para el cliente (opcional)..."
                  maxLength={500}
                  value={clientNote}
                  onChange={(e) => setClientNote(e.target.value)}
                  className="h-8 w-full rounded-control border border-app-border bg-white px-3 text-xs text-app-text placeholder:text-app-text-muted/60 focus:border-app-primary focus:outline-none focus:ring-1 focus:ring-app-primary"
                />
              </div>

              {/* Derecha: Botones de acción y atajos */}
              <div className="flex items-center gap-3">
                <div className="hidden lg:flex items-center gap-1.5 text-[11px] text-app-text-muted mr-2">
                  <kbd className="rounded border border-app-border bg-app-surface-muted px-1.5 py-0.5 font-mono">
                    ←
                  </kbd>
                  <kbd className="rounded border border-app-border bg-app-surface-muted px-1.5 py-0.5 font-mono">
                    →
                  </kbd>
                  <kbd className="rounded border border-app-border bg-app-surface-muted px-1.5 py-0.5 font-mono">
                    A
                  </kbd>
                  <kbd className="rounded border border-app-border bg-app-surface-muted px-1.5 py-0.5 font-mono">
                    Esc
                  </kbd>
                </div>

                {hasExclusionK && (
                  <button
                    type="button"
                    onClick={() => setConfirmRegimeOpen(true)}
                    className="inline-flex h-8 items-center rounded-control border border-severity-red-border bg-severity-red-bg px-3 text-xs font-semibold text-severity-red-text hover:bg-severity-red-border/40 focus:outline-none focus:ring-2 focus:ring-severity-red-icon"
                  >
                    Confirmar pase a RI
                  </button>
                )}

                <button
                  type="button"
                  disabled={approveMutation.isPending}
                  onClick={() => handleApprove(false)}
                  className="inline-flex h-8 items-center rounded-control border border-app-border bg-white px-3 text-xs font-medium text-app-text hover:bg-app-surface-muted disabled:opacity-50 focus:outline-none focus:ring-2 focus:ring-app-focus"
                >
                  Aprobar
                </button>

                <button
                  type="button"
                  disabled={approveMutation.isPending}
                  onClick={() => handleApprove(true)}
                  className="inline-flex h-8 items-center rounded-control bg-app-primary px-3 text-xs font-medium text-white hover:bg-app-primary-hover disabled:opacity-50 focus:outline-none focus:ring-2 focus:ring-app-focus"
                >
                  {approveMutation.isPending ? 'Aprobando...' : 'Aprobar y siguiente'}
                </button>
              </div>
            </div>
          </Dialog.Content>
        </Dialog.Portal>
      </Dialog.Root>

      {/* Diálogo de confirmación de pase a RI (Criterio 12 y T009.10) */}
      {detail && (
        <ConfirmRegimeModal
          open={confirmRegimeOpen}
          onOpenChange={setConfirmRegimeOpen}
          clientId={detail.client.id}
          clientName={detail.client.name}
          effectiveDate={effectiveExclusionDate}
          onSuccess={() => {
            refetch();
          }}
        />
      )}
    </>
  );
}
