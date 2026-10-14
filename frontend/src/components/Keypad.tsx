import type { Operation } from "../api/client";

type Cell =
  | { kind: "operation"; name: string }
  | { kind: "insert"; label: string; name: string; text: string }
  | { kind: "action"; label: string; name: string; action: "clear" | "delete" };

const digit = (d: string): Cell => ({ kind: "insert", label: d, name: d, text: d });
const operation = (name: string): Cell => ({ kind: "operation", name });

/** The fixed phone layout: 4 columns, 6 rows. */
const grid: Cell[] = [
  { kind: "action", label: "AC", name: "all clear", action: "clear" },
  { kind: "action", label: "⌫", name: "delete", action: "delete" },
  { kind: "insert", label: "(", name: "open bracket", text: "(" },
  { kind: "insert", label: ")", name: "close bracket", text: ")" },
  operation("sqrt"),
  operation("power"),
  operation("percentage"),
  operation("divide"),
  digit("7"),
  digit("8"),
  digit("9"),
  operation("multiply"),
  digit("4"),
  digit("5"),
  digit("6"),
  operation("subtract"),
  digit("1"),
  digit("2"),
  digit("3"),
  operation("add"),
  { kind: "insert", label: "EE", name: "exponent", text: "e" },
  digit("0"),
  { kind: "insert", label: ".", name: "decimal point", text: "." },
];

/** The accessible names of the operation keys. */
const operationNames: Record<string, string> = {
  add: "plus",
  subtract: "minus",
  multiply: "times",
  divide: "divide",
  power: "power",
  sqrt: "square root",
  percentage: "percent",
};

const gridOperations = new Set(
  grid.flatMap((cell) => (cell.kind === "operation" ? [cell.name] : [])),
);

interface KeypadProps {
  operations: Operation[];
  loading: boolean;
  onInsert: (text: string) => void;
  onClear: () => void;
  onDelete: () => void;
}

export function Keypad({ operations, loading, onInsert, onClear, onDelete }: KeypadProps) {
  const byName = new Map(operations.map((op) => [op.name, op]));
  const extra = operations.filter((op) => !gridOperations.has(op.name));

  const operationKey = (op: Operation) => (
    <button
      key={op.name}
      type="button"
      className="key key-operation"
      aria-label={operationNames[op.name] ?? op.name}
      onClick={() => {
        onInsert(op.symbol);
      }}
    >
      {op.symbol}
    </button>
  );

  return (
    <div className="keypad">
      {grid.map((cell, index) => {
        switch (cell.kind) {
          case "operation": {
            const op = byName.get(cell.name);
            return op ? (
              operationKey(op)
            ) : (
              <span key={index} className="key-empty" aria-hidden="true" />
            );
          }
          case "insert":
            return (
              <button
                key={index}
                type="button"
                className="key"
                aria-label={cell.name}
                onClick={() => {
                  onInsert(cell.text);
                }}
              >
                {cell.label}
              </button>
            );
          case "action":
            return (
              <button
                key={index}
                type="button"
                className="key key-function"
                aria-label={cell.name}
                onClick={cell.action === "clear" ? onClear : onDelete}
              >
                {cell.label}
              </button>
            );
        }
      })}
      <button type="submit" className="key key-equals" aria-label="equals" disabled={loading}>
        =
      </button>
      {extra.map(operationKey)}
    </div>
  );
}
