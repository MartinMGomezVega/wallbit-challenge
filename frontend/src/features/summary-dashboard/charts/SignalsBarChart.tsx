import { useNavigate } from 'react-router-dom';
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Cell,
} from 'recharts';
import { SignalStat } from '../schemas';
import {
  SEVERITY_FILL_COLORS,
  TOOLTIP_CONTENT_STYLE,
  usePrefersReducedMotion,
} from './chartUtils';

export interface SignalsBarChartProps {
  period: string;
  signals: SignalStat[];
  className?: string;
}

export function SignalsBarChart({
  period,
  signals,
  className = '',
}: SignalsBarChartProps) {
  const navigate = useNavigate();
  const reducedMotion = usePrefersReducedMotion();
  const topSignals = signals.slice(0, 8);

  const handleSelectSignal = (code: string) => {
    navigate(`/contador?period=${encodeURIComponent(period)}&signal=${encodeURIComponent(code)}`);
  };

  return (
    <section
      className={`rounded-lg border border-app-border bg-app-surface p-5 ${className}`}
      aria-label="Distribución de situaciones detectadas"
    >
      <div className="mb-4">
        <h2 className="text-lg font-semibold text-app-text">
          Situaciones detectadas
        </h2>
        <p className="text-sm text-app-text-muted">
          Click en una barra para ver esos clientes
        </p>
      </div>

      {topSignals.length === 0 ? (
        <p className="py-12 text-center text-sm text-app-text-muted">
          No hay situaciones detectadas en este período.
        </p>
      ) : (
        <div
          className="h-64 w-full"
          role="img"
          aria-label={`Gráfico de barras horizontales con las ${topSignals.length} situaciones principales`}
        >
          <ResponsiveContainer width="100%" height="100%">
            <BarChart
              data={topSignals}
              layout="vertical"
              margin={{ top: 4, right: 16, left: 8, bottom: 4 }}
            >
              <CartesianGrid
                stroke="#E5E5E5"
                strokeDasharray="3 3"
                vertical={false}
              />
              <XAxis
                type="number"
                axisLine={false}
                tickLine={false}
                tick={{ fill: '#6B6B6B', fontSize: 12 }}
                allowDecimals={false}
              />
              <YAxis
                type="category"
                dataKey="label"
                width={170}
                axisLine={false}
                tickLine={false}
                tick={{ fill: '#6B6B6B', fontSize: 12 }}
              />
              <Tooltip
                cursor={{ fill: '#F7F7F7' }}
                contentStyle={TOOLTIP_CONTENT_STYLE}
                itemStyle={{ color: '#FFFFFF' }}
                labelStyle={{ color: '#FFFFFF', fontWeight: 600 }}
                formatter={(value) => [`${value ?? 0} clientes`, 'Cantidad']}
              />
              <Bar
                dataKey="count"
                maxBarSize={28}
                radius={[0, 4, 4, 0]}
                isAnimationActive={!reducedMotion}
                className="cursor-pointer"
                onClick={(entry) => {
                  const payload = entry?.payload as SignalStat | undefined;
                  if (payload?.code) {
                    handleSelectSignal(payload.code);
                  }
                }}
              >
                {topSignals.map((sig) => (
                  <Cell
                    key={sig.code}
                    fill={SEVERITY_FILL_COLORS[sig.severity]}
                  />
                ))}
              </Bar>
            </BarChart>
          </ResponsiveContainer>
        </div>
      )}

      <table className="sr-only">
        <caption>Situaciones detectadas en el período</caption>
        <thead>
          <tr>
            <th>Situación</th>
            <th>Severidad</th>
            <th>Clientes</th>
            <th>Acción</th>
          </tr>
        </thead>
        <tbody>
          {topSignals.map((sig) => (
            <tr key={sig.code}>
              <td>{sig.label}</td>
              <td>{sig.severity}</td>
              <td>{sig.count}</td>
              <td>
                <button
                  type="button"
                  onClick={() => handleSelectSignal(sig.code)}
                >
                  Ver clientes con {sig.label}
                </button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </section>
  );
}

