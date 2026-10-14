import type { Calculation } from "../api/client";
import { formatNumber } from "../format";

interface HistoryProps {
  items: Calculation[];
  onClear: () => void;
}

export function History({ items, onClear }: HistoryProps) {
  if (items.length === 0) return null;
  return (
    <section className="history" aria-labelledby="history-title">
      <div className="history-header">
        <h2 id="history-title">History</h2>
        <button type="button" className="link-button" onClick={onClear}>
          Clear history
        </button>
      </div>
      <ol className="history-list" aria-label="History">
        {items.map((item, index) => (
          <li key={`${String(items.length - index)}-${item.expression}`}>
            {item.expression} = {formatNumber(item.result)}
          </li>
        ))}
      </ol>
    </section>
  );
}
