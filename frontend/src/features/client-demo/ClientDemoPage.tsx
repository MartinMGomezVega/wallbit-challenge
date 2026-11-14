import { useEffect, useState, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { DEMO_CLIENTS } from './demoClients';
import { PhoneFrame } from './PhoneFrame';
import {
  getClientProfile,
  getClientClosings,
  getClientInvoices,
  setClientExternalIncome,
  confirmClientPeriod,
} from './api';
import { ClientProfile, ClientClosing, ClientInvoice } from './schemas';
import { InvoicesStep } from './CheckIn/InvoicesStep';
import { ExternalIncomeStep, ExternalIncomeItemForm } from './CheckIn/ExternalIncomeStep';
import { ConfirmStep } from './CheckIn/ConfirmStep';
import { InReview } from './InReview';
import { ValidatedSummaryCard } from './ValidatedSummaryCard';
import { ApiError } from '../../shared/api/http';

export function ClientDemoPage() {
  const { clientId = 'cli_demo_15' } = useParams<{ clientId: string }>();
  const navigate = useNavigate();

  const [profile, setProfile] = useState<ClientProfile | null>(null);
  const [closings, setClosings] = useState<ClientClosing[]>([]);
  const [invoices, setInvoices] = useState<ClientInvoice[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // CheckIn form state
  const [extAnswered, setExtAnswered] = useState(false);
  const [extNone, setExtNone] = useState(true);
  const [extItems, setExtItems] = useState<ExternalIncomeItemForm[]>([]);
  const [confirming, setConfirming] = useState(false);
  const [confirmError, setConfirmError] = useState<string | null>(null);

  const loadData = useCallback(async (cid: string) => {
    setLoading(true);
    setError(null);
    try {
      const [profRes, closRes, invsRes] = await Promise.all([
        getClientProfile(cid),
        getClientClosings(cid),
        getClientInvoices(cid, '2026-10'),
      ]);
      setProfile(profRes);
      setClosings(closRes);
      setInvoices(invsRes);
    } catch (err: unknown) {
      if (err instanceof ApiError) {
        setError(err.message);
      } else {
        setError('Error al cargar la información del cliente.');
      }
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadData(clientId);
  }, [clientId, loadData]);

  const handleExternalIncomeChange = useCallback(
    (answered: boolean, none: boolean, items: ExternalIncomeItemForm[]) => {
      setExtAnswered(answered);
      setExtNone(none);
      setExtItems(items);
    },
    [],
  );

  const handleConfirmCheckIn = async () => {
    setConfirming(true);
    setConfirmError(null);
    try {
      const apiItems = extItems.map(it => ({
        amount: { amount: it.amount.trim(), currency: 'USD' },
        description: it.description.trim(),
      }));
      await setClientExternalIncome(clientId, '2026-10', {
        items: apiItems,
        none: extNone,
      });
      await confirmClientPeriod(clientId, '2026-10');
      await loadData(clientId);
    } catch (err: unknown) {
      if (err instanceof ApiError) {
        setConfirmError(err.message || 'Error al confirmar el período');
      } else {
        setConfirmError('Error al confirmar el período');
      }
    } finally {
      setConfirming(false);
    }
  };

  const octClosing = closings.find(c => c.period === '2026-10');
  const sepClosing = closings.find(c => c.period === '2026-09');

  return (
    <div className="space-y-6">
      {/* Top Demo Selector */}
      <div className="mx-auto max-w-2xl rounded-2xl border border-neutral-800 bg-[#171717] p-4 text-white shadow-md">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 border-b border-neutral-800 pb-3">
          <div>
            <h1 className="text-sm font-bold text-white">Vista Demo del Cliente</h1>
            <p className="text-xs text-neutral-400">
              Elegí un cliente para ver cómo vive el cierre dentro de la app móvil de Wallbit:
            </p>
          </div>
          <span className="self-start rounded-full bg-emerald-500/10 px-2.5 py-0.5 text-[11px] font-semibold text-emerald-400 border border-emerald-500/20">
            API Real /api/v1/me/**
          </span>
        </div>

        <nav aria-label="Clientes de demo" className="mt-3 flex flex-wrap gap-2">
          {DEMO_CLIENTS.map(cl => {
            const isSelected = cl.id === clientId;
            return (
              <button
                key={cl.id}
                type="button"
                onClick={() => navigate(`/cliente/${cl.id}`)}
                aria-pressed={isSelected}
                className={`rounded-xl px-3 py-1.5 text-xs font-medium transition-all text-left ${
                  isSelected
                    ? 'bg-white text-neutral-900 shadow font-semibold ring-2 ring-white/50'
                    : 'bg-[#262626] text-neutral-300 hover:bg-[#333333] hover:text-white'
                }`}
              >
                <div className="flex items-center gap-1.5">
                  <span>{cl.name}</span>
                  <span className="text-[10px] opacity-70">({cl.regime === 'Monotributo' ? 'Mono' : 'RI'})</span>
                </div>
                <div className="text-[10px] opacity-60 truncate max-w-[170px]">
                  {cl.shortScenario}
                </div>
              </button>
            );
          })}
        </nav>
      </div>

      {/* Phone Frame */}
      <PhoneFrame clientName={profile?.name}>
        {loading ? (
          <div className="flex h-64 flex-col items-center justify-center gap-2 text-neutral-500">
            <span className="inline-block h-6 w-6 animate-spin rounded-full border-2 border-neutral-300 border-t-neutral-800" />
            <span className="text-xs font-medium">Cargando datos de la app...</span>
          </div>
        ) : error ? (
          <div className="rounded-2xl border border-red-200 bg-red-50 p-4 text-center text-xs text-red-700">
            <p className="font-semibold">{error}</p>
            <button
              type="button"
              onClick={() => loadData(clientId)}
              className="mt-3 rounded-lg bg-red-700 px-3 py-1.5 text-xs font-semibold text-white hover:bg-red-800"
            >
              Reintentar
            </button>
          </div>
        ) : !octClosing ? (
          <div className="rounded-2xl border border-neutral-200 bg-white p-6 text-center text-xs text-neutral-500">
            Todavía no hay un cierre para este mes.
          </div>
        ) : octClosing.state === 'READY' ? (
          <div className="space-y-3">
            <h2 className="text-base font-bold text-neutral-900">
              Cierre de octubre 2026
            </h2>
            <ValidatedSummaryCard
              closing={octClosing}
              clientId={clientId}
              onPaidSuccess={() => loadData(clientId)}
            />
          </div>
        ) : octClosing.state === 'IN_REVIEW' ? (
          <div className="space-y-3">
            <h2 className="text-base font-bold text-neutral-900">
              Cierre de octubre 2026
            </h2>
            <InReview
              octoberClosing={octClosing}
              previousClosing={sepClosing}
              clientId={clientId}
              onRefresh={() => loadData(clientId)}
            />
          </div>
        ) : (
          /* ACTION_REQUIRED: CheckIn */
          <div className="space-y-4">
            <div>
              <h2 className="text-base font-bold text-neutral-900">
                Cierre de octubre 2026
              </h2>
              <p className="text-xs text-neutral-500 mt-0.5">
                Completá los pasos para que tu contadora pueda liquidar tus impuestos.
              </p>
            </div>

            <InvoicesStep
              clientId={clientId}
              clientRegime={profile?.regime ?? 'MONOTRIBUTO'}
              invoices={invoices}
              onRefresh={() => loadData(clientId)}
            />

            <ExternalIncomeStep
              initialDeclaredNone={octClosing.input.answered && extNone}
              onChange={handleExternalIncomeChange}
            />

            <ConfirmStep
              canConfirm={extAnswered}
              loading={confirming}
              error={confirmError}
              onConfirm={handleConfirmCheckIn}
            />
          </div>
        )}
      </PhoneFrame>
    </div>
  );
}

