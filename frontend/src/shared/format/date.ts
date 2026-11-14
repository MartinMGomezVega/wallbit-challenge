/**
 * Formatea fechas según las UI guidelines:
 * - short: "dd/MM" (para tablas)
 * - full: "dd/MM/yyyy"
 */
export function formatDate(
  dateInput: string | Date | null | undefined,
  mode: 'short' | 'full' = 'short'
): string {
  if (!dateInput) return '-';

  // Si es string YYYY-MM-DD o ISO
  if (typeof dateInput === 'string') {
    const clean = dateInput.trim();
    if (clean.length >= 10 && clean[4] === '-' && clean[7] === '-') {
      const year = clean.substring(0, 4);
      const month = clean.substring(5, 7);
      const day = clean.substring(8, 10);
      if (mode === 'short') {
        return `${day}/${month}`;
      }
      return `${day}/${month}/${year}`;
    }
  }

  const d = dateInput instanceof Date ? dateInput : new Date(dateInput);
  if (isNaN(d.getTime())) return '-';

  const day = d.getDate().toString().padStart(2, '0');
  const month = (d.getMonth() + 1).toString().padStart(2, '0');
  if (mode === 'short') {
    return `${day}/${month}`;
  }
  const year = d.getFullYear().toString();
  return `${day}/${month}/${year}`;
}

export function formatDueDate(dateInput: string | Date | null | undefined): string {
  if (!dateInput) return '-';
  const formatted = formatDate(dateInput, 'short');
  if (formatted === '-') return '-';
  return `vence ${formatted}`;
}

const MONTH_NAMES = [
  'enero',
  'febrero',
  'marzo',
  'abril',
  'mayo',
  'junio',
  'julio',
  'agosto',
  'septiembre',
  'octubre',
  'noviembre',
  'diciembre',
];

/**
 * Convierte un período "YYYY-MM" a "mes año" (ej: "2026-10" -> "octubre 2026").
 */
export function formatPeriodName(periodStr: string): string {
  const parts = periodStr.split('-');
  const month = parts[1];
  const year = parts[0];
  if (parts.length !== 2 || !month || !year) return periodStr;
  const monthNum = parseInt(month, 10);
  if (isNaN(monthNum) || monthNum < 1 || monthNum > 12) return periodStr;
  return `${MONTH_NAMES[monthNum - 1]} ${year}`;
}
