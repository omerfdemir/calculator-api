import { useCallback, useEffect, useState } from "react";
import { fetchOperations, type ErrorCode, type Operation } from "../api/client";

export type OperationsState =
  | { status: "loading" }
  | { status: "ready"; operations: Operation[] }
  | { status: "error"; code: ErrorCode };

/** Loads the operations from the API at start. retry loads them again. */
export function useOperations(): OperationsState & { retry: () => void } {
  const [state, setState] = useState<OperationsState>({ status: "loading" });
  const [attempt, setAttempt] = useState(0);

  useEffect(() => {
    let active = true;
    void fetchOperations().then((result) => {
      if (!active) return;
      setState(
        result.ok
          ? { status: "ready", operations: result.value }
          : { status: "error", code: result.code },
      );
    });
    return () => {
      active = false;
    };
  }, [attempt]);

  const retry = useCallback(() => {
    setState({ status: "loading" });
    setAttempt((n) => n + 1);
  }, []);

  return { ...state, retry };
}
