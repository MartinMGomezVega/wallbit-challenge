import { useQuery } from '@tanstack/react-query';
import { Clock } from 'lucide-react';
import { z } from 'zod';
import { apiFetch } from '../api/http';
import { formatDate } from '../format/date';
import { Tooltip } from './Tooltip';

const HealthSchema = z.object({
  status: z.string(),
  now: z.string(),
});

export const SIMULATED_DATE_TOOLTIP =
  'El prototipo usa una fecha fija (APP_NOW) para que la demo sea siempre igual: el cierre de octubre ya corrió y estamos en la ventana de revisión, antes de los vencimientos de noviembre.';

export function SimulatedDateBadge() {
  const { data } = useQuery({
    queryKey: ['health-now'],
    queryFn: () => apiFetch('/api/v1/health', HealthSchema),
    staleTime: 5 * 60 * 1000,
  });

  const rawDate = data?.now || '2026-11-05T09:00:00-03:00';
  const formattedDate = formatDate(rawDate, 'full');

  // Si now coincide con la fecha real del sistema (o sea, APP_NOW no está seteada),
  // el badge dice solo "Hoy: dd/MM/yyyy" sin el texto de "fecha simulada".
  const todayRealFormatted = formatDate(new Date(), 'full');
  const isSimulated = formattedDate !== todayRealFormatted;

  const label = isSimulated
    ? `Hoy: ${formattedDate} · fecha simulada`
    : `Hoy: ${formattedDate}`;

  return (
    <Tooltip content={SIMULATED_DATE_TOOLTIP}>
      <span
        tabIndex={0}
        role="status"
        aria-label={label}
        data-testid="simulated-date-badge"
        className="inline-flex h-8 cursor-default items-center gap-1.5 rounded-md border border-[#262626] bg-[#171717] px-2.5 text-xs text-[#D4D4D4] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-app-brand"
      >
        <Clock className="h-3.5 w-3.5 text-neutral-400" aria-hidden="true" />
        <span>{label}</span>
      </span>
    </Tooltip>
  );
}
