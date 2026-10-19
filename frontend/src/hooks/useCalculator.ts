import { useCallback, useMemo, useRef, useState } from "react";
import { calculate, type Calculation, type Operation } from "../api/client";
import type { MessageCode } from "../errors/messages";
import { continueFromResult, toDisplay, toRequest, type Carry } from "../carry";

export const maxExpressionLength = 200;
export const historyLimit = 10;

export type Status = "idle" | "loading" | "success" | "error";

export interface CalculatorError {
  code: MessageCode;
  position?: number;
}

interface State {
  expression: string;
  status: Status;
  result: Calculation | null;
  error: CalculatorError | null;
  history: Calculation[];
  /** True right after a result: the next key continues from it or starts again. */
  afterResult: boolean;
  carry: Carry | null;
}

const initialState: State = {
  expression: "",
  status: "idle",
  result: null,
  error: null,
  history: [],
  afterResult: false,
  carry: null,
};

/** The ASCII characters that the input shows as keypad symbols. */
const asciiSymbols: Record<string, string> = { "*": "×", "/": "÷", "-": "−" };

function clamp(expression: string): string {
  return Array.from(expression).slice(0, maxExpressionLength).join("");
}

/** Applies the phone rule for text that follows a result. */
function afterResultInput(
  s: State,
  result: Calculation,
  text: string,
  startsNew: (text: string) => boolean,
): State {
  const next = continueFromResult(result, text, startsNew);
  return { ...s, expression: clamp(next.expression), carry: next.carry, afterResult: false };
}

/**
 * Holds the calculator state, does the client-side validation and applies the
 * phone keypad rules. The API does all calculations.
 */
export function useCalculator(operations: Operation[]) {
  const [state, setState] = useState<State>(initialState);
  const requestId = useRef(0);

  const allowed = useMemo(() => {
    const symbols = operations.map((op) => op.symbol).filter((s) => Array.from(s).length === 1);
    return new Set([...Array.from("0123456789.eE() "), ...Object.keys(asciiSymbols), ...symbols]);
  }, [operations]);

  const startsNew = useMemo(() => {
    const prefixSymbols = operations
      .filter((op) => op.notation === "prefix")
      .map((op) => op.symbol);
    return (text: string) => /^[0-9.(]/.test(text) || prefixSymbols.some((s) => text.startsWith(s));
  }, [operations]);

  /** Adds text at the end. After a result, an operator continues from the result. */
  const insert = useCallback(
    (text: string) => {
      setState((s) => {
        if (s.afterResult && s.result !== null) {
          return afterResultInput(s, s.result, text, startsNew);
        }
        return { ...s, expression: clamp(s.expression + text), afterResult: false };
      });
    },
    [startsNew],
  );

  /** Sets the typed value. It keeps only calculator characters. */
  const type = useCallback(
    (value: string) => {
      const filtered = Array.from(value)
        .filter((c) => allowed.has(c))
        .map((c) => asciiSymbols[c] ?? c)
        .join("");
      setState((s) => {
        if (
          s.afterResult &&
          s.result !== null &&
          filtered.startsWith(s.expression) &&
          filtered.length > s.expression.length
        ) {
          const text = filtered.slice(s.expression.length);
          return afterResultInput(s, s.result, text, startsNew);
        }
        return { ...s, expression: clamp(filtered), afterResult: false };
      });
    },
    [allowed, startsNew],
  );

  const deleteLast = useCallback(() => {
    setState((s) => ({
      ...s,
      expression: Array.from(s.expression).slice(0, -1).join(""),
      afterResult: false,
    }));
  }, []);

  const clear = useCallback(() => {
    requestId.current++; // Ignore a response that is still running.
    setState((s) => ({ ...initialState, history: s.history }));
  }, []);

  const clearHistory = useCallback(() => {
    setState((s) => ({ ...s, history: [] }));
  }, []);

  const evaluate = useCallback(async () => {
    if (state.status === "loading") return;
    if (state.expression.trim() === "") {
      setState((s) => ({
        ...s,
        status: "error",
        error: { code: "EMPTY_EXPRESSION" },
        result: null,
      }));
      return;
    }

    const id = ++requestId.current;
    setState((s) => ({ ...s, status: "loading", error: null }));
    const { sent, used } = toRequest(state.expression, state.carry);
    const response = await calculate(sent);
    if (id !== requestId.current) return;

    if (response.ok) {
      const calculation = toDisplay(response.value, used);
      setState((s) => ({
        ...s,
        status: "success",
        result: calculation,
        error: null,
        history: [calculation, ...s.history].slice(0, historyLimit),
        afterResult: true,
        carry: null,
      }));
    } else {
      const error: CalculatorError =
        response.position === undefined
          ? { code: response.code }
          : { code: response.code, position: response.position };
      setState((s) => ({ ...s, status: "error", result: null, error, afterResult: false }));
    }
  }, [state.expression, state.status, state.carry]);

  return { ...state, insert, type, deleteLast, clear, clearHistory, evaluate };
}
