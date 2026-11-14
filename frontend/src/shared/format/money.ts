export interface MoneyDTO {
  amount: number; // Centavos
  currency: 'ARS' | 'USD';
}

/**
 * Formatea un MoneyDTO a string según las UI guidelines:
 * - ARS: "$ 1.234.567,89"
 * - USD: "US$ 1.234,56"
 */
export function formatMoney(money: MoneyDTO | null | undefined): string {
  if (!money) return '-';
  return formatMoneyParts(money.amount, money.currency);
}

export function formatMoneyParts(cents: number, currency: 'ARS' | 'USD'): string {
  const isNegative = cents < 0;
  const absCents = Math.abs(cents);
  const units = Math.floor(absCents / 100);
  const decimals = absCents % 100;

  // Formato es-AR con puntos para miles y coma para decimales
  const unitsStr = units.toLocaleString('es-AR');
  const decimalsStr = decimals.toString().padStart(2, '0');

  const prefix = currency === 'USD' ? 'US$ ' : '$ ';
  const sign = isNegative ? '-' : '';

  return `${sign}${prefix}${unitsStr},${decimalsStr}`;
}

