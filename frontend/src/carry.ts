import type { Calculation } from "./api/client";
import { formatExact, formatNumber } from "./format";

/**
 * A result that the expression continues from. The input shows the rounded
 * display text, and the API gets the full value.
 */
export interface Carry {
  display: string;
  full: string;
}

/** Puts a negative number in brackets, so that `^` raises the whole value: (−3)^2. */
function bracketNegative(text: string, value: number): string {
  return value < 0 ? `(${text})` : text;
}

/** Applies the phone rule for text after a result: continue from it, or start again. */
export function continueFromResult(
  result: Calculation,
  text: string,
  startsNew: (text: string) => boolean,
): { expression: string; carry: Carry | null } {
  if (startsNew(text)) {
    return { expression: text, carry: null };
  }
  const display = bracketNegative(formatNumber(result.result), result.result);
  return {
    expression: display + text,
    carry: { display, full: bracketNegative(formatExact(result.result), result.result) },
  };
}

/**
 * Returns the expression for the API. While the expression starts with the
 * carried display text, the API gets the full value. A digit right after the
 * display text changes the number, so then the typed text is sent.
 */
export function toRequest(
  expression: string,
  carry: Carry | null,
): { sent: string; used: Carry | null } {
  if (carry === null || !expression.startsWith(carry.display)) {
    return { sent: expression, used: null };
  }
  const rest = expression.slice(carry.display.length);
  if (/^[0-9.eE]/.test(rest)) {
    return { sent: expression, used: null };
  }
  return { sent: carry.full + rest, used: carry };
}

/** Shows the display text again in place of the full value that the API echoes. */
export function toDisplay(calculation: Calculation, used: Carry | null): Calculation {
  if (used === null || !calculation.expression.startsWith(used.full)) {
    return calculation;
  }
  return {
    ...calculation,
    expression: used.display + calculation.expression.slice(used.full.length),
  };
}
