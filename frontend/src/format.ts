/**
 * Formats a result for the display: 12 significant digits, no trailing zeros,
 * and the − sign of the keypad. The API returns the full float64 value.
 */
export function formatNumber(value: number): string {
  return Number(value.toPrecision(12)).toString().replace("e+", "e").replaceAll("-", "−");
}
