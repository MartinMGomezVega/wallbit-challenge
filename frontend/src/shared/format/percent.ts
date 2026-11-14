/**
 * Formatea basis points (ej: 8700 -> "87%", 10000 -> "100%").
 */
export function formatPercentBp(bp: number | null | undefined): string {
  if (bp == null) return '0%';
  const pct = Math.round(bp / 100);
  return `${pct}%`;
}

