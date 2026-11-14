# 008 — Plan

## Archivos
```
features/closings-dashboard/
├── index.ts
├── api.ts                 # listClosings, getSummary, getCatalog (zod schemas)
├── schemas.ts
├── useClosingsQuery.ts    # TanStack Query: keys ['closings', filters]
├── useDashboardParams.ts  # lee/escribe la URL (useSearchParams) → ClosingsFilter tipado
├── DashboardPage.tsx
├── SummaryChips.tsx
├── FiltersBar.tsx
├── SignalMultiSelect.tsx
├── ClosingsTable.tsx      # TanStack Table, manualSorting/manualPagination
├── columns.tsx
├── SelectionBar.tsx
└── __tests__/
features/parameters/
├── index.ts, api.ts, ParametersPage.tsx, __tests__/
shared/ui/
├── SeverityIndicator.tsx  # props: severity; renderiza color+ícono+aria-label
├── SignalChip.tsx
├── ProgressBar.tsx
├── StatusBadge.tsx
├── Tooltip.tsx            # Radix
├── Pagination.tsx
shared/format/
├── money.ts (formatMoney(MoneyDTO)), date.ts (dd/MM, dd/MM/yyyy), percent.ts (bp → "87%")
```

## Detalles
- `useDashboardParams` es la **única** fuente de verdad del filtro; exporta también `toQueryString(filter)` que reusa 009 para pedir los IDs con el mismo filtro.
- Query keys incluyen el filtro serializado; `placeholderData: keepPreviousData` para que la tabla no parpadee al paginar.
- Colores (Tailwind): rojo `red-600`, amarillo `amber-500`, verde `emerald-600`; fondos de fila sutiles opcionales solo en rojo (`red-50`).
- Tabla: encabezado sticky, filas de 56 px, números alineados a la derecha con `tabular-nums`.
- Ícono + color + texto en tooltip (accesibilidad, criterio 4.2).
