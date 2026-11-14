import React, { useState, useEffect, useMemo } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useBlocker } from 'react-router-dom';
import * as Dialog from '@radix-ui/react-dialog';
import { AlertCircle, Loader2 } from 'lucide-react';
import { getParameters, updateParameters } from './api';
import { ParametersDTO } from './schemas';
import { ApiError } from '../../shared/api/http';
import { Toast } from '../../shared/ui/Toast';
import { useDocumentTitle } from '../../shared/hooks/useDocumentTitle';

export function ParametersPage() {
  useDocumentTitle('Parámetros · Wallbit Contabilidad');
  const queryClient = useQueryClient();
  const { data: initialParams, isLoading, isError, error } = useQuery({
    queryKey: ['parameters'],
    queryFn: getParameters,
  });

  const [form, setForm] = useState<ParametersDTO | null>(null);
  const [savedParams, setSavedParams] = useState<ParametersDTO | null>(null);
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});
  const [toast, setToast] = useState<{ message: string; type: 'success' | 'error' } | null>(null);

  useEffect(() => {
    if (initialParams) {
      setForm(initialParams);
      setSavedParams(initialParams);
    }
  }, [initialParams]);

  const isDirty = useMemo(() => {
    if (!form || !savedParams) return false;
    return JSON.stringify(form) !== JSON.stringify(savedParams);
  }, [form, savedParams]);

  useEffect(() => {
    const handleBeforeUnload = (e: BeforeUnloadEvent) => {
      if (isDirty) {
        e.preventDefault();
        e.returnValue = '';
      }
    };
    window.addEventListener('beforeunload', handleBeforeUnload);
    return () => window.removeEventListener('beforeunload', handleBeforeUnload);
  }, [isDirty]);

  const blocker = useBlocker(
    ({ currentLocation, nextLocation }) =>
      isDirty && currentLocation.pathname !== nextLocation.pathname
  );

  const mutation = useMutation({
    mutationFn: updateParameters,
    onSuccess: (res) => {
      setFieldErrors({});
      setSavedParams(res.parameters);
      setForm(res.parameters);
      setToast({
        message: `Parámetros guardados. Se recalcularon ${res.recalculated} cierres.`,
        type: 'success',
      });
      queryClient.setQueryData(['parameters'], res.parameters);
      queryClient.invalidateQueries({ queryKey: ['closings'] });
      queryClient.invalidateQueries({ queryKey: ['closings-summary'] });
    },
    onError: (err) => {
      if (err instanceof ApiError && err.status === 422) {
        setFieldErrors(err.fields || {});
      }
      setToast({
        message: (err instanceof ApiError && err.message) || 'Error al guardar los parámetros',
        type: 'error',
      });
    },
  });

  if (isLoading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-app-primary" aria-label="Cargando parámetros" />
      </div>
    );
  }

  if (isError || !form) {
    return (
      <div className="rounded-card border border-severity-red-border bg-severity-red-bg p-6 text-center">
        <AlertCircle className="mx-auto h-8 w-8 text-severity-red-icon mb-2" />
        <h3 className="text-sm font-semibold text-severity-red-text">
          Error al cargar los parámetros
        </h3>
        <p className="text-xs text-severity-red-text/80 mt-1">
          {(error as Error)?.message || 'No se pudieron recuperar los parámetros del contador.'}
        </p>
      </div>
    );
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!isDirty || mutation.isPending) return;
    setFieldErrors({});
    mutation.mutate(form);
  };

  return (
    <div className="max-w-4xl space-y-6">
      <div>
        <h1 className="text-xl font-bold text-app-text">
          Parámetros
        </h1>
        <p className="text-xs text-app-text-muted mt-0.5">
          Parámetros globales de contabilidad. Configurá las reglas de negocio, umbrales de alerta y tolerancias aplicadas a todos los cierres.
        </p>
      </div>

      <form onSubmit={handleSubmit} noValidate className="space-y-6">
        {/* Sección 1: Monotributo y categorías */}
        <div className="rounded-card border border-app-border bg-app-surface p-5 shadow-sm space-y-4">
          <h2 className="text-sm font-bold text-app-text border-b border-app-border pb-2">
            Monotributo y alertas de categoría
          </h2>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* invoiceECountsForCap */}
            <div className="md:col-span-2 flex items-start gap-3 pt-1">
              <input
                id="invoiceECountsForCap"
                type="checkbox"
                checked={form.invoiceECountsForCap}
                onChange={(e) =>
                  setForm({ ...form, invoiceECountsForCap: e.target.checked })
                }
                className="mt-0.5 h-4 w-4 rounded-[4px] border-app-border text-app-primary focus:ring-app-focus"
              />
              <label htmlFor="invoiceECountsForCap" className="text-xs cursor-pointer">
                <span className="font-semibold text-app-text block">
                  Computar facturas E en los límites de Monotributo
                </span>
                <span className="text-app-text-muted text-[11px]">
                  Si está activo, las facturas de exportación de servicios se acumulan para el cálculo de categoría y exclusión.
                </span>
              </label>
            </div>

            {/* categoryCapWarningBp */}
            <div>
              <label htmlFor="categoryCapWarningBp" className="block text-xs font-semibold text-app-text mb-1">
                Alerta de tope de categoría (%)
              </label>
              <input
                id="categoryCapWarningBp"
                type="number"
                min="1"
                max="100"
                value={Math.round(form.categoryCapWarningBp / 100)}
                onChange={(e) =>
                  setForm({
                    ...form,
                    categoryCapWarningBp: Number(e.target.value) * 100,
                  })
                }
                className="h-8 w-full rounded-control border border-app-border bg-white px-3 text-xs text-app-text focus:outline-none focus:ring-2 focus:ring-app-focus"
              />
              <span className="text-[11px] text-app-text-muted mt-0.5 block">
                Porcentaje de facturación para alertar cercanía al tope (default: 80%).
              </span>
              {fieldErrors.categoryCapWarningBp && (
                <span className="text-xs text-severity-red-text font-medium mt-1 block">
                  {fieldErrors.categoryCapWarningBp}
                </span>
              )}
            </div>

            {/* kCapWarningBp */}
            <div>
              <label htmlFor="kCapWarningBp" className="block text-xs font-semibold text-app-text mb-1">
                Alerta de tope K de exclusión (%)
              </label>
              <input
                id="kCapWarningBp"
                type="number"
                min="1"
                max="100"
                value={Math.round(form.kCapWarningBp / 100)}
                onChange={(e) =>
                  setForm({
                    ...form,
                    kCapWarningBp: Number(e.target.value) * 100,
                  })
                }
                className="h-8 w-full rounded-control border border-app-border bg-white px-3 text-xs text-app-text focus:outline-none focus:ring-2 focus:ring-app-focus"
              />
              <span className="text-[11px] text-app-text-muted mt-0.5 block">
                Porcentaje de facturación respecto al tope K máximo (default: 80%).
              </span>
              {fieldErrors.kCapWarningBp && (
                <span className="text-xs text-severity-red-text font-medium mt-1 block">
                  {fieldErrors.kCapWarningBp}
                </span>
              )}
            </div>

            {/* cardSpendKCapBp */}
            <div>
              <label htmlFor="cardSpendKCapBp" className="block text-xs font-semibold text-app-text mb-1">
                Tope de gastos con tarjeta (%)
              </label>
              <input
                id="cardSpendKCapBp"
                type="number"
                min="1"
                max="100"
                value={Math.round(form.cardSpendKCapBp / 100)}
                onChange={(e) =>
                  setForm({
                    ...form,
                    cardSpendKCapBp: Number(e.target.value) * 100,
                  })
                }
                className="h-8 w-full rounded-control border border-app-border bg-white px-3 text-xs text-app-text focus:outline-none focus:ring-2 focus:ring-app-focus"
              />
              <span className="text-[11px] text-app-text-muted mt-0.5 block">
                Porcentaje del tope K para alertar gastos con tarjeta en 12 meses (default: 40%).
              </span>
              {fieldErrors.cardSpendKCapBp && (
                <span className="text-xs text-severity-red-text font-medium mt-1 block">
                  {fieldErrors.cardSpendKCapBp}
                </span>
              )}
            </div>
          </div>
        </div>

        {/* Sección 2: Tipo de cambio (FX) */}
        <div className="rounded-card border border-app-border bg-app-surface p-5 shadow-sm space-y-4">
          <h2 className="text-sm font-bold text-app-text border-b border-app-border pb-2">
            Tipo de cambio (FX)
          </h2>

          <div>
            <label htmlFor="fxRule" className="block text-xs font-semibold text-app-text mb-1">
              Regla de cotización de divisas
            </label>
            <select
              id="fxRule"
              value={form.fxRule}
              onChange={(e) =>
                setForm({
                  ...form,
                  fxRule: e.target.value as ParametersDTO['fxRule'],
                })
              }
              className="h-8 w-full md:w-96 rounded-control border border-app-border bg-white px-2.5 text-xs text-app-text focus:outline-none focus:ring-2 focus:ring-app-focus"
            >
              <option value="BNA_BUY_PREV_BUSINESS_DAY">
                BNA Comprador del día hábil anterior (Recomendado ARCA)
              </option>
              <option value="BNA_SELL_PREV_BUSINESS_DAY">
                BNA Vendedor del día hábil anterior
              </option>
              <option value="BNA_BUY_SAME_DAY">BNA Comprador del mismo día</option>
              <option value="BNA_SELL_SAME_DAY">BNA Vendedor del mismo día</option>
            </select>
            <span className="text-[11px] text-app-text-muted mt-0.5 block">
              Tipo de cambio oficial BNA para pesificar facturas emitidas en dólares.
            </span>
            {fieldErrors.fxRule && (
              <span className="text-xs text-severity-red-text font-medium mt-1 block">
                {fieldErrors.fxRule}
              </span>
            )}
          </div>
        </div>

        {/* Sección 3: Cobros y facturación */}
        <div className="rounded-card border border-app-border bg-app-surface p-5 shadow-sm space-y-4">
          <h2 className="text-sm font-bold text-app-text border-b border-app-border pb-2">
            Cobros y facturación
          </h2>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* missingInvoiceMinUsd */}
            <div>
              <label htmlFor="missingInvoiceMinUsd" className="block text-xs font-semibold text-app-text mb-1">
                Monto mínimo para alerta de facturas faltantes (US$)
              </label>
              <input
                id="missingInvoiceMinUsd"
                type="number"
                step="0.01"
                min="0"
                value={form.missingInvoiceMinUsd.amount}
                onChange={(e) =>
                  setForm({
                    ...form,
                    missingInvoiceMinUsd: {
                      ...form.missingInvoiceMinUsd,
                      amount: e.target.value,
                    },
                  })
                }
                className="h-8 w-full rounded-control border border-app-border bg-white px-3 text-xs text-app-text focus:outline-none focus:ring-2 focus:ring-app-focus"
              />
              <span className="text-[11px] text-app-text-muted mt-0.5 block">
                Monto mínimo en USD sin factura para alertar (default: US$ 1.000,00).
              </span>
              {(fieldErrors.missingInvoiceMinUsd || fieldErrors.missingInvoiceMinUSD) && (
                <span className="text-xs text-severity-red-text font-medium mt-1 block">
                  {fieldErrors.missingInvoiceMinUsd || fieldErrors.missingInvoiceMinUSD}
                </span>
              )}
            </div>

            {/* atypicalMultiplierBp */}
            <div>
              <label htmlFor="atypicalMultiplierBp" className="block text-xs font-semibold text-app-text mb-1">
                Multiplicador de cobro atípico (ratio)
              </label>
              <input
                id="atypicalMultiplierBp"
                type="number"
                step="0.1"
                min="1"
                value={(form.atypicalMultiplierBp / 10000).toFixed(1)}
                onChange={(e) =>
                  setForm({
                    ...form,
                    atypicalMultiplierBp: Math.round(Number(e.target.value) * 10000),
                  })
                }
                className="h-8 w-full rounded-control border border-app-border bg-white px-3 text-xs text-app-text focus:outline-none focus:ring-2 focus:ring-app-focus"
              />
              <span className="text-[11px] text-app-text-muted mt-0.5 block">
                Factor sobre la mediana histórica para considerar ingreso atípico (default: 2,5×).
              </span>
              {fieldErrors.atypicalMultiplierBp && (
                <span className="text-xs text-severity-red-text font-medium mt-1 block">
                  {fieldErrors.atypicalMultiplierBp}
                </span>
              )}
            </div>

            {/* atypicalWindowMonths */}
            <div>
              <label htmlFor="atypicalWindowMonths" className="block text-xs font-semibold text-app-text mb-1">
                Ventana de cobro atípico (meses)
              </label>
              <input
                id="atypicalWindowMonths"
                type="number"
                min="1"
                max="24"
                value={form.atypicalWindowMonths}
                onChange={(e) =>
                  setForm({
                    ...form,
                    atypicalWindowMonths: Number(e.target.value),
                  })
                }
                className="h-8 w-full rounded-control border border-app-border bg-white px-3 text-xs text-app-text focus:outline-none focus:ring-2 focus:ring-app-focus"
              />
              <span className="text-[11px] text-app-text-muted mt-0.5 block">
                Cantidad de meses previos para el cálculo de la mediana (default: 6 meses).
              </span>
              {fieldErrors.atypicalWindowMonths && (
                <span className="text-xs text-severity-red-text font-medium mt-1 block">
                  {fieldErrors.atypicalWindowMonths}
                </span>
              )}
            </div>

            {/* reminderDaysBeforeDue */}
            <div>
              <label htmlFor="reminderDaysBeforeDue" className="block text-xs font-semibold text-app-text mb-1">
                Días de aviso previo a vencimiento (días)
              </label>
              <input
                id="reminderDaysBeforeDue"
                type="number"
                min="1"
                max="30"
                value={form.reminderDaysBeforeDue}
                onChange={(e) =>
                  setForm({
                    ...form,
                    reminderDaysBeforeDue: Number(e.target.value),
                  })
                }
                className="h-8 w-full rounded-control border border-app-border bg-white px-3 text-xs text-app-text focus:outline-none focus:ring-2 focus:ring-app-focus"
              />
              <span className="text-[11px] text-app-text-muted mt-0.5 block">
                Anticipación para recordatorios al cliente (default: 5 días).
              </span>
              {fieldErrors.reminderDaysBeforeDue && (
                <span className="text-xs text-severity-red-text font-medium mt-1 block">
                  {fieldErrors.reminderDaysBeforeDue}
                </span>
              )}
            </div>
          </div>
        </div>

        {/* Sección 4: Conciliación */}
        <div className="rounded-card border border-app-border bg-app-surface p-5 shadow-sm space-y-4">
          <h2 className="text-sm font-bold text-app-text border-b border-app-border pb-2">
            Reglas de conciliación automática
          </h2>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            {/* reconciliationToleranceBp */}
            <div>
              <label htmlFor="reconciliationToleranceBp" className="block text-xs font-semibold text-app-text mb-1">
                Tolerancia de importe (%)
              </label>
              <input
                id="reconciliationToleranceBp"
                type="number"
                step="0.1"
                min="0"
                max="10"
                value={(form.reconciliationToleranceBp / 100).toFixed(1)}
                onChange={(e) =>
                  setForm({
                    ...form,
                    reconciliationToleranceBp: Math.round(Number(e.target.value) * 100),
                  })
                }
                className="h-8 w-full rounded-control border border-app-border bg-white px-3 text-xs text-app-text focus:outline-none focus:ring-2 focus:ring-app-focus"
              />
              <span className="text-[11px] text-app-text-muted mt-0.5 block">
                Diferencia porcentual tolerada entre cobro y factura (default: 1%).
              </span>
              {fieldErrors.reconciliationToleranceBp && (
                <span className="text-xs text-severity-red-text font-medium mt-1 block">
                  {fieldErrors.reconciliationToleranceBp}
                </span>
              )}
            </div>

            {/* reconciliationWindowBeforeDays */}
            <div>
              <label htmlFor="reconciliationWindowBeforeDays" className="block text-xs font-semibold text-app-text mb-1">
                Días previos (emisión antes de cobro)
              </label>
              <input
                id="reconciliationWindowBeforeDays"
                type="number"
                min="0"
                max="60"
                value={form.reconciliationWindowBeforeDays}
                onChange={(e) =>
                  setForm({
                    ...form,
                    reconciliationWindowBeforeDays: Number(e.target.value),
                  })
                }
                className="h-8 w-full rounded-control border border-app-border bg-white px-3 text-xs text-app-text focus:outline-none focus:ring-2 focus:ring-app-focus"
              />
              <span className="text-[11px] text-app-text-muted mt-0.5 block">
                Días de emisión anteriores al cobro (default: 5 días).
              </span>
              {fieldErrors.reconciliationWindowBeforeDays && (
                <span className="text-xs text-severity-red-text font-medium mt-1 block">
                  {fieldErrors.reconciliationWindowBeforeDays}
                </span>
              )}
            </div>

            {/* reconciliationWindowAfterDays */}
            <div>
              <label htmlFor="reconciliationWindowAfterDays" className="block text-xs font-semibold text-app-text mb-1">
                Días posteriores (cobro posterior)
              </label>
              <input
                id="reconciliationWindowAfterDays"
                type="number"
                min="0"
                max="90"
                value={form.reconciliationWindowAfterDays}
                onChange={(e) =>
                  setForm({
                    ...form,
                    reconciliationWindowAfterDays: Number(e.target.value),
                  })
                }
                className="h-8 w-full rounded-control border border-app-border bg-white px-3 text-xs text-app-text focus:outline-none focus:ring-2 focus:ring-app-focus"
              />
              <span className="text-[11px] text-app-text-muted mt-0.5 block">
                Días de cobro posteriores a la emisión (default: 30 días).
              </span>
              {fieldErrors.reconciliationWindowAfterDays && (
                <span className="text-xs text-severity-red-text font-medium mt-1 block">
                  {fieldErrors.reconciliationWindowAfterDays}
                </span>
              )}
            </div>
          </div>
        </div>

        {/* Botón de Guardar */}
        <div className="flex items-center justify-end gap-3 pt-2">
          <button
            type="submit"
            disabled={!isDirty || mutation.isPending}
            className="inline-flex h-9 items-center rounded-control bg-app-primary px-4 text-xs font-semibold text-white shadow-sm hover:bg-app-primary-hover disabled:opacity-50 disabled:cursor-not-allowed focus:outline-none focus:ring-2 focus:ring-app-focus"
          >
            {mutation.isPending ? (
              <span className="inline-flex items-center gap-2">
                <Loader2 className="h-4 w-4 animate-spin" />
                Guardando...
              </span>
            ) : (
              <span>Guardar</span>
            )}
          </button>
        </div>
      </form>

      {/* Toast de notificación */}
      {toast && (
        <Toast
          message={toast.message}
          type={toast.type}
          onClose={() => setToast(null)}
        />
      )}

      {/* Modal de cambios sin guardar al navegar */}
      <Dialog.Root
        open={blocker.state === 'blocked'}
        onOpenChange={(open) => {
          if (!open && blocker.state === 'blocked') {
            blocker.reset?.();
          }
        }}
      >
        <Dialog.Portal>
          <Dialog.Overlay className="fixed inset-0 z-50 bg-[#0A0A0A]/50 backdrop-blur-xs transition-opacity" />
          <Dialog.Content
            aria-describedby="unsaved-modal-desc"
            className="fixed left-1/2 top-1/2 z-50 w-full max-w-md -translate-x-1/2 -translate-y-1/2 rounded-card border border-app-border bg-app-surface p-6 shadow-xl focus:outline-none"
          >
            <Dialog.Title className="text-base font-semibold text-app-text">
              ¿Salir sin guardar?
            </Dialog.Title>
            <p id="unsaved-modal-desc" className="mt-2 text-xs text-app-text-muted">
              Tenés cambios sin guardar en los parámetros.
            </p>
            <div className="mt-6 flex flex-wrap items-center justify-end gap-2">
              <button
                type="button"
                onClick={() => blocker.reset?.()}
                className="h-8 rounded-control border border-app-border bg-white px-3 text-xs font-medium text-app-text hover:bg-app-surface-muted focus:outline-none focus:ring-1 focus:ring-app-focus"
              >
                Seguir editando
              </button>
              <button
                type="button"
                onClick={() => blocker.proceed?.()}
                className="h-8 rounded-control border border-severity-red-border bg-white px-3 text-xs font-medium text-severity-red-text hover:bg-severity-red-bg focus:outline-none focus:ring-1 focus:ring-severity-red-border"
              >
                Salir sin guardar
              </button>
              <button
                type="button"
                disabled={mutation.isPending}
                onClick={async () => {
                  if (!form) return;
                  try {
                    await mutation.mutateAsync(form);
                    blocker.proceed?.();
                  } catch {
                    blocker.reset?.();
                  }
                }}
                className="h-8 rounded-control bg-app-primary px-3 text-xs font-semibold text-white shadow-sm hover:bg-app-primary-hover disabled:opacity-50 focus:outline-none focus:ring-2 focus:ring-app-focus"
              >
                {mutation.isPending ? 'Guardando...' : 'Guardar y salir'}
              </button>
            </div>
          </Dialog.Content>
        </Dialog.Portal>
      </Dialog.Root>
    </div>
  );
}
