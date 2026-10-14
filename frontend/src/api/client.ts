/** The error codes of the API. They are part of the public contract. */
const apiErrorCodes = [
  "INVALID_REQUEST",
  "REQUEST_TOO_LARGE",
  "INVALID_EXPRESSION",
  "DIVISION_BY_ZERO",
  "INVALID_DOMAIN",
  "RESULT_OUT_OF_RANGE",
  "NOT_FOUND",
  "METHOD_NOT_ALLOWED",
  "INTERNAL_ERROR",
] as const;

export type ErrorCode = (typeof apiErrorCodes)[number] | "NETWORK_ERROR";

export type Notation = "infix" | "prefix" | "postfix";

export interface Operation {
  name: string;
  symbol: string;
  notation: Notation;
  arity: number;
}

export interface Calculation {
  expression: string;
  result: number;
}

export interface Failure {
  ok: false;
  code: ErrorCode;
  position?: number;
}

export type ApiResult<T> = { ok: true; value: T } | Failure;

export interface RequestOptions {
  /** The API origin. The default is VITE_API_URL, or the same origin. */
  baseUrl?: string;
  timeoutMs?: number;
}

const defaultTimeoutMs = 5000;

export function calculate(
  expression: string,
  options?: RequestOptions,
): Promise<ApiResult<Calculation>> {
  return request(
    "/api/v1/calculate",
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ expression }),
    },
    parseCalculation,
    options,
  );
}

export function fetchOperations(options?: RequestOptions): Promise<ApiResult<Operation[]>> {
  return request("/api/v1/operations", { method: "GET" }, parseOperations, options);
}

async function request<T>(
  path: string,
  init: RequestInit,
  parse: (body: unknown) => T | null,
  options: RequestOptions = {},
): Promise<ApiResult<T>> {
  const baseUrl = options.baseUrl ?? import.meta.env.VITE_API_URL ?? "";
  const controller = new AbortController();
  const timer = setTimeout(() => {
    controller.abort();
  }, options.timeoutMs ?? defaultTimeoutMs);

  try {
    const response = await fetch(new URL(path, baseUrl || window.location.origin), {
      ...init,
      signal: controller.signal,
    });
    const body = await readJson(response);
    if (response.ok) {
      const value = parse(body);
      return value === null ? failure("INTERNAL_ERROR") : { ok: true, value };
    }
    return parseError(response.status, body);
  } catch {
    // A network failure or the timeout.
    return failure("NETWORK_ERROR");
  } finally {
    clearTimeout(timer);
  }
}

async function readJson(response: Response): Promise<unknown> {
  try {
    return (await response.json()) as unknown;
  } catch (error) {
    if (error instanceof DOMException && error.name === "AbortError") throw error;
    return undefined;
  }
}

function failure(code: ErrorCode, position?: number): Failure {
  return position === undefined ? { ok: false, code } : { ok: false, code, position };
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

function isApiErrorCode(value: unknown): value is (typeof apiErrorCodes)[number] {
  return apiErrorCodes.some((code) => code === value);
}

function parseError(status: number, body: unknown): Failure {
  const error = isRecord(body) ? body.error : undefined;
  if (isRecord(error) && isApiErrorCode(error.code)) {
    return failure(error.code, typeof error.position === "number" ? error.position : undefined);
  }
  // A proxy error page means that the API is not available.
  return failure(status >= 502 && status <= 504 ? "NETWORK_ERROR" : "INTERNAL_ERROR");
}

function parseCalculation(body: unknown): Calculation | null {
  if (isRecord(body) && typeof body.expression === "string" && typeof body.result === "number") {
    return { expression: body.expression, result: body.result };
  }
  return null;
}

function isOperation(value: unknown): value is Operation {
  return (
    isRecord(value) &&
    typeof value.name === "string" &&
    typeof value.symbol === "string" &&
    (value.notation === "infix" || value.notation === "prefix" || value.notation === "postfix") &&
    typeof value.arity === "number"
  );
}

function parseOperations(body: unknown): Operation[] | null {
  if (!isRecord(body) || !Array.isArray(body.operations)) return null;
  const operations: unknown[] = body.operations;
  return operations.every(isOperation) ? operations : null;
}
