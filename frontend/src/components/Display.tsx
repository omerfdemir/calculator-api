import type { Calculation } from "../api/client";
import { maxExpressionLength } from "../hooks/useCalculator";
import { formatNumber } from "../format";

interface DisplayProps {
  expression: string;
  loading: boolean;
  result: Calculation | null;
  onType: (value: string) => void;
}

export function Display({ expression, loading, result, onType }: DisplayProps) {
  return (
    <div className="display">
      <label htmlFor="expression" className="display-label">
        Expression
      </label>
      <input
        id="expression"
        className="display-input"
        type="text"
        inputMode="none"
        autoComplete="off"
        autoCorrect="off"
        spellCheck={false}
        maxLength={maxExpressionLength}
        value={expression}
        onChange={(event) => {
          onType(event.target.value);
        }}
      />
      <p role="status" className="display-result" aria-busy={loading}>
        {loading
          ? "Calculating…"
          : result && `${result.expression} = ${formatNumber(result.result)}`}
      </p>
    </div>
  );
}
