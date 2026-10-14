import type { ErrorCode } from "../api/client";

/** The API error codes and the client-side validation code. */
export type MessageCode = ErrorCode | "EMPTY_EXPRESSION";

const somethingWentWrong = "Something went wrong. Please try again.";

/** Returns the user-friendly message for a code. The UI never shows raw server text. */
export function errorMessage(code: MessageCode, position?: number): string {
  switch (code) {
    case "EMPTY_EXPRESSION":
      return "Enter an expression.";
    case "INVALID_EXPRESSION":
      return position === undefined
        ? "The expression is not complete or not valid."
        : `The expression is not complete or not valid. Check character ${String(position + 1)}.`;
    case "INVALID_REQUEST":
      return "The request was not valid. Please try again.";
    case "REQUEST_TOO_LARGE":
      return "The expression is too long.";
    case "DIVISION_BY_ZERO":
      return "You cannot divide by zero.";
    case "INVALID_DOMAIN":
      return "The input is outside the domain of the operation, for example the square root of a negative number.";
    case "RESULT_OUT_OF_RANGE":
      return "The result is too large to show.";
    case "NETWORK_ERROR":
      return "The server is not available. Check your connection and try again.";
    case "METHOD_NOT_ALLOWED":
    case "INTERNAL_ERROR":
      return somethingWentWrong;
  }
}
