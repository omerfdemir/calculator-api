/**
 * Formats a result for the display: 12 significant digits, no trailing zeros,
 * and the − sign of the keypad. The API returns the full float64 value.
 */
export function formatNumber(value: number): string {
  return Number(value.toPrecision(12)).toString().replace("e+", "e").replaceAll("-", "−");
}

/**
 * Formats the full float64 value with the − sign of the keypad, so that the
 * API can read it back as the same value.
 */
export function formatExact(value: number): string {
  return String(value).replace("e+", "e").replaceAll("-", "−");
}
