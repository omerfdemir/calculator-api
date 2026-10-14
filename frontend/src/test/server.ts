import { http, HttpResponse, type JsonBodyType } from "msw";
import { setupServer } from "msw/node";

export const operationsUrl = "*/api/v1/operations";
export const calculateUrl = "*/api/v1/calculate";

export const allOperations = [
  { name: "add", symbol: "+", notation: "infix", arity: 2 },
  { name: "subtract", symbol: "−", notation: "infix", arity: 2 },
  { name: "multiply", symbol: "×", notation: "infix", arity: 2 },
  { name: "divide", symbol: "÷", notation: "infix", arity: 2 },
  { name: "power", symbol: "^", notation: "infix", arity: 2 },
  { name: "sqrt", symbol: "√", notation: "prefix", arity: 1 },
  { name: "percentage", symbol: "%", notation: "postfix", arity: 1 },
];

/** Fixed API responses, keyed by the expression that the app sends. */
const fixtures: Record<string, { status: number; body: JsonBodyType }> = {
  "2+3": ok("2 + 3", 5),
  "10−4": ok("10 − 4", 6),
  "6×7": ok("6 × 7", 42),
  "1÷4": ok("1 ÷ 4", 0.25),
  "2^10": ok("2 ^ 10", 1024),
  "√9": ok("√9", 3),
  "15%×80": ok("15% × 80", 12),
  "80+15%": ok("80 + 15%", 92),
  "2+3×4": ok("2 + 3 × 4", 14),
  "(2+3)×4": ok("(2 + 3) × 4", 20),
  "−1.5×2": ok("−1.5 × 2", -3),
  "1e6+1": ok("1e6 + 1", 1000001),
  "0.1+0.2": ok("0.1 + 0.2", 0.30000000000000004),
  "14×2": ok("14 × 2", 28),
  "−3×2": ok("−3 × 2", -6),
  "1÷3": ok("1 ÷ 3", 0.3333333333333333),
  "0.3333333333333333×3": ok("0.3333333333333333 × 3", 1),
  "−1÷3": ok("−1 ÷ 3", -0.3333333333333333),
  "−0.3333333333333333×3": ok("−0.3333333333333333 × 3", -1),
  "0.33333333333×3": ok("0.33333333333 × 3", 0.99999999999),
  "0.3333333333335×3": ok("0.3333333333335 × 3", 1.0000000000005),
  "1÷0": error(422, "DIVISION_BY_ZERO"),
  "√−1": error(422, "INVALID_DOMAIN"),
  "1e308×10": error(422, "RESULT_OUT_OF_RANGE"),
  "2+": error(400, "INVALID_EXPRESSION", 2),
};

function ok(expression: string, result: number) {
  return { status: 200, body: { expression, result } };
}

function error(status: number, code: string, position?: number) {
  return { status, body: { error: { code, message: "server text", position } } };
}

export const handlers = [
  http.get(operationsUrl, () => HttpResponse.json({ operations: allOperations })),
  http.post(calculateUrl, async ({ request }) => {
    const body = (await request.json()) as { expression: string };
    const fixture = fixtures[body.expression];
    if (fixture === undefined) {
      throw new Error(`No fixture for the expression ${JSON.stringify(body.expression)}`);
    }
    return HttpResponse.json(fixture.body, { status: fixture.status });
  }),
];

export const server = setupServer(...handlers);
