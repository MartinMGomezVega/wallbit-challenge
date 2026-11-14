import { useQuery } from '@tanstack/react-query';
import { fetchProfile, fetchMyClosings } from './api';
import { PeriodCard } from './PeriodCard';
import { History } from './History';

export function ClientPage() {
  const { data: profile, isLoading: loadingProfile, error: profileError } = useQuery({
    queryKey: ['client-profile'],
    queryFn: fetchProfile,
  });

  const { data: closings = [], isLoading: loadingClosings, error: closingsError } = useQuery({
    queryKey: ['my-closings'],
    queryFn: fetchMyClosings,
  });

  if (loadingProfile || loadingClosings) {
    return (
      <div className="mx-auto max-w-[640px] space-y-4">
        <div className="h-28 rounded-card bg-app-surface border border-app-border animate-pulse" />
        <div className="h-64 rounded-card bg-app-surface border border-app-border animate-pulse" />
      </div>
    );
  }

  if (profileError || closingsError) {
    return (
      <div className="mx-auto max-w-[640px] rounded-card border border-severity-red-border bg-severity-red-bg p-6 text-center text-sm text-severity-red-text">
        Ocurrió un error al cargar la información. Por favor intentá nuevamente.
      </div>
    );
  }

  if (!profile) return null;

  const currentClosing = closings[0];
  const historyClosings = closings.slice(1);

  return (
    <div className="mx-auto max-w-[640px] space-y-6">
      {/* Encabezado del cliente */}
      <div className="flex items-center justify-between flex-wrap gap-2 pb-1">
        <div>
          <h1 className="text-xl font-bold text-app-text">
            Hola, {profile.name}
          </h1>
          <p className="text-xs text-app-text-muted mt-0.5">
            CUIT {profile.cuitMasked} · {profile.regime}
            {profile.category && ` (Categoría ${profile.category})`}
          </p>
        </div>
      </div>

      {/* Tarjeta del período activo */}
      {currentClosing ? (
        <PeriodCard closing={currentClosing} regime={profile.regime} />
      ) : (
        <div className="rounded-card border border-app-border bg-app-surface p-6 text-center text-sm text-app-text-muted">
          No hay cierres disponibles para tu cuenta.
        </div>
      )}

      {/* Historial de períodos anteriores */}
      {historyClosings.length > 0 && <History closings={historyClosings} />}
    </div>
  );
}

