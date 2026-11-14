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
import { CategoryStat } from '../schemas';
import {
  PRIMARY_COLOR,
  TOOLTIP_CONTENT_STYLE,
  WARNING_BORDER_COLOR,
  usePrefersReducedMotion,
} from './chartUtils';

export interface CategoryBarChartProps {
  categories: CategoryStat[];
  nearKCap: number;
  className?: string;
}

const NEAR_CAP_CATEGORIES = new Set(['I', 'J', 'K']);

export function CategoryBarChart({
  categories,
  nearKCap,
  className = '',
}: CategoryBarChartProps) {
  const reducedMotion = usePrefersReducedMotion();

  return (
    <section
      className={`rounded-lg border border-app-border bg-app-surface p-5 ${className}`}
      aria-label="Distribución de clientes por categoría"
    >
      <div className="mb-4 flex items-start justify-between gap-2">
        <div>
          <h2 className="text-lg font-semibold text-app-text">
            Distribución por categoría
          </h2>
          <p className="text-sm text-app-text-muted">
            Categorías A–K de Monotributo y Responsable Inscripto
          </p>
        </div>
        {nearKCap > 0 && (
          <span className="inline-flex shrink-0 items-center rounded-full border border-app-severity-yellow-border bg-app-severity-yellow-bg px-2.5 py-0.5 text-xs font-medium text-app-severity-yellow-text">
            {nearKCap} cerca del tope K
          </span>
        )}
      </div>

      {categories.length === 0 ? (
        <p className="py-12 text-center text-sm text-app-text-muted">
          Sin datos de categorías para este período.
        </p>
      ) : (
        <div
          className="h-64 w-full"
          role="img"
          aria-label="Gráfico de barras verticales con la cantidad de clientes por categoría"
        >
          <ResponsiveContainer width="100%" height="100%">
            <BarChart
              data={categories}
              margin={{ top: 8, right: 8, left: 0, bottom: 4 }}
            >
              <CartesianGrid
                stroke="#E5E5E5"
                strokeDasharray="3 3"
                vertical={false}
              />
              <XAxis
                dataKey="category"
                axisLine={false}
                tickLine={false}
                tick={{ fill: '#6B6B6B', fontSize: 12 }}
              />
              <YAxis
                axisLine={false}
                tickLine={false}
                tick={{ fill: '#6B6B6B', fontSize: 12 }}
                allowDecimals={false}
              />
              <Tooltip
                cursor={{ fill: '#F7F7F7' }}
                contentStyle={TOOLTIP_CONTENT_STYLE}
                itemStyle={{ color: '#FFFFFF' }}
                labelStyle={{ color: '#FFFFFF', fontWeight: 600 }}
                formatter={(value) => [`${value ?? 0} clientes`, 'Cantidad']}
              />
              <Bar
                dataKey="clients"
                maxBarSize={28}
                radius={[4, 4, 0, 0]}
                isAnimationActive={!reducedMotion}
              >
                {categories.map((cat) => {
                  const highlight =
                    nearKCap > 0 && NEAR_CAP_CATEGORIES.has(cat.category);
                  return (
                    <Cell
                      key={cat.category}
                      fill={PRIMARY_COLOR}
                      stroke={highlight ? WARNING_BORDER_COLOR : undefined}
                      strokeWidth={highlight ? 2 : 0}
                    />
                  );
                })}
              </Bar>
            </BarChart>
          </ResponsiveContainer>
        </div>
      )}

      <table className="sr-only">
        <caption>Clientes por categoría</caption>
        <thead>
          <tr>
            <th>Categoría</th>
            <th>Clientes</th>
          </tr>
        </thead>
        <tbody>
          {categories.map((cat) => (
            <tr key={cat.category}>
              <td>{cat.category}</td>
              <td>{cat.clients}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </section>
  );
}

