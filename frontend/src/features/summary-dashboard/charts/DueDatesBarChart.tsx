import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from 'recharts';
import { DueDateTotal } from '../schemas';
import { formatDate } from '../../../shared/format/date';
import { formatMoneyParts } from '../../../shared/format/money';
import {
  PRIMARY_COLOR,
  TOOLTIP_CONTENT_STYLE,
  usePrefersReducedMotion,
} from './chartUtils';

export interface DueDatesBarChartProps {
  dueDates: DueDateTotal[];
  className?: string;
}

interface DueDateChartPoint {
  date: string;
  label: string;
  amountCents: number;
  clients: number;
}

export function DueDatesBarChart({
  dueDates,
  className = '',
}: DueDatesBarChartProps) {
  const reducedMotion = usePrefersReducedMotion();

  const data: DueDateChartPoint[] = dueDates.map((d) => ({
    date: d.date,
    label: formatDate(d.date, 'short'),
    amountCents: d.amount.amount,
    clients: d.clients,
  }));

  return (
    <section
      className={`rounded-lg border border-app-border bg-app-surface p-5 ${className}`}
      aria-label="Vencimientos de pago por fecha"
    >
      <div className="mb-4">
        <h2 className="text-lg font-semibold text-app-text">
          Vencimientos próximos
        </h2>
        <p className="text-sm text-app-text-muted">
          Total estimado a pagar en ARS por fecha de vencimiento
        </p>
      </div>

      {data.length === 0 ? (
        <p className="py-12 text-center text-sm text-app-text-muted">
          No hay vencimientos con saldo a pagar en este período.
        </p>
      ) : (
        <div
          className="h-64 w-full"
          role="img"
          aria-label="Gráfico de barras con los vencimientos próximos en pesos"
        >
          <ResponsiveContainer width="100%" height="100%">
            <BarChart
              data={data}
              margin={{ top: 8, right: 16, left: 16, bottom: 4 }}
            >
              <CartesianGrid
                stroke="#E5E5E5"
                strokeDasharray="3 3"
                vertical={false}
              />
              <XAxis
                dataKey="label"
                axisLine={false}
                tickLine={false}
                tick={{ fill: '#6B6B6B', fontSize: 12 }}
              />
              <YAxis
                axisLine={false}
                tickLine={false}
                width={90}
                tick={{ fill: '#6B6B6B', fontSize: 12 }}
                tickFormatter={(v: number) =>
                  `$ ${Math.round(v / 100).toLocaleString('es-AR')}`
                }
              />
              <Tooltip
                cursor={{ fill: '#F7F7F7' }}
                contentStyle={TOOLTIP_CONTENT_STYLE}
                itemStyle={{ color: '#FFFFFF' }}
                labelStyle={{ color: '#FFFFFF', fontWeight: 600 }}
                formatter={(value, _name, item) => {
                  const point = item?.payload as DueDateChartPoint | undefined;
                  const cents = typeof value === 'number' ? value : 0;
                  const clientsSuffix = point
                    ? ` (${point.clients} clientes)`
                    : '';
                  return [
                    `${formatMoneyParts(cents, 'ARS')}${clientsSuffix}`,
                    'Total a pagar',
                  ];
                }}
              />
              <Bar
                dataKey="amountCents"
                fill={PRIMARY_COLOR}
                maxBarSize={28}
                radius={[4, 4, 0, 0]}
                isAnimationActive={!reducedMotion}
              />
            </BarChart>
          </ResponsiveContainer>
        </div>
      )}

      <table className="sr-only">
        <caption>Vencimientos próximos</caption>
        <thead>
          <tr>
            <th>Fecha</th>
            <th>Total a pagar</th>
            <th>Clientes</th>
          </tr>
        </thead>
        <tbody>
          {data.map((d) => (
            <tr key={d.date}>
              <td>{d.label}</td>
              <td>{formatMoneyParts(d.amountCents, 'ARS')}</td>
              <td>{d.clients}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </section>
  );
}

