import { useEffect } from "react";
import type { Operation } from "../api/client";
import { useCalculator } from "../hooks/useCalculator";
import { useOperations } from "../hooks/useOperations";
import { Display } from "./Display";
import { ErrorMessage } from "./ErrorMessage";
import { History } from "./History";
import { Keypad } from "./Keypad";

/** The container: it loads the operations, then shows the calculator. */
export function Calculator() {
  const operations = useOperations();

  switch (operations.status) {
    case "loading":
      return <p className="loading">Loading…</p>;
    case "error":
      return (
        <div className="load-error">
          <ErrorMessage code={operations.code} />
          <button type="button" className="retry" onClick={operations.retry}>
            Retry
          </button>
        </div>
      );
    case "ready":
      return <CalculatorForm operations={operations.operations} />;
  }
}

function CalculatorForm({ operations }: { operations: Operation[] }) {
  const calc = useCalculator(operations);
  const { clear } = calc;

  // Escape clears everything, wherever the focus is.
  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") clear();
    };
    document.addEventListener("keydown", onKeyDown);
    return () => {
      document.removeEventListener("keydown", onKeyDown);
    };
  }, [clear]);

  const loading = calc.status === "loading";
  return (
    <>
      <form
        className="calculator"
        aria-label="Calculator"
        onSubmit={(event) => {
          event.preventDefault();
          void calc.evaluate();
        }}
      >
        <Display
          expression={calc.expression}
          loading={loading}
          result={calc.result}
          onType={calc.type}
        />
        {calc.error && <ErrorMessage code={calc.error.code} position={calc.error.position} />}
        <Keypad
          operations={operations}
          loading={loading}
          onInsert={calc.insert}
          onClear={calc.clear}
          onDelete={calc.deleteLast}
        />
      </form>
      <History items={calc.history} onClear={calc.clearHistory} />
    </>
  );
}
