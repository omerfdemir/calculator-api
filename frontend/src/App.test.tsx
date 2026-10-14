import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent, { type UserEvent } from "@testing-library/user-event";
import { delay, http, HttpResponse } from "msw";
import { describe, expect, it } from "vitest";
import { App } from "./App";
import { allOperations, calculateUrl, operationsUrl, server } from "./test/server";

const keyNames: Record<string, string> = {
  "+": "plus",
  "−": "minus",
  "×": "times",
  "÷": "divide",
  "^": "power",
  "√": "square root",
  "%": "percent",
  ".": "decimal point",
  e: "exponent",
  "(": "open bracket",
  ")": "close bracket",
};

async function renderApp() {
  const user = userEvent.setup();
  render(<App />);
  await screen.findByRole("button", { name: "equals" });
  return user;
}

/** Clicks the keypad key for each character of keys. */
async function press(user: UserEvent, keys: string) {
  for (const key of keys) {
    await user.click(screen.getByRole("button", { name: keyNames[key] ?? key }));
  }
}

function expressionInput() {
  return screen.getByRole("textbox", { name: "Expression" });
}

async function expectResult(text: string) {
  await waitFor(() => {
    expect(screen.getByRole("status")).toHaveTextContent(text);
  });
}

describe("operations", () => {
  it("shows a key for each operation from the API", async () => {
    await renderApp();
    for (const name of ["plus", "minus", "times", "divide", "power", "square root", "percent"]) {
      expect(screen.getByRole("button", { name })).toBeInTheDocument();
    }
  });

  it("shows only the operations that the API lists", async () => {
    server.use(
      http.get(operationsUrl, () =>
        HttpResponse.json({ operations: allOperations.filter((op) => op.name !== "power") }),
      ),
    );
    await renderApp();
    expect(screen.queryByRole("button", { name: "power" })).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "plus" })).toBeInTheDocument();
  });

  it("shows an operation that is not in the keypad grid in an extra row", async () => {
    server.use(
      http.get(operationsUrl, () =>
        HttpResponse.json({
          operations: [
            ...allOperations,
            { name: "modulo", symbol: "mod", notation: "infix", arity: 2 },
          ],
        }),
      ),
    );
    await renderApp();
    expect(screen.getByRole("button", { name: "modulo" })).toHaveTextContent("mod");
  });

  it("shows an error and a Retry button when the operations do not load", async () => {
    server.use(http.get(operationsUrl, () => HttpResponse.error(), { once: true }));
    const user = userEvent.setup();
    render(<App />);

    expect(await screen.findByRole("alert")).toHaveTextContent("The server is not available");
    expect(screen.queryByRole("button", { name: "equals" })).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Retry" }));
    expect(await screen.findByRole("button", { name: "equals" })).toBeInTheDocument();
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });
});

describe("calculation with the keypad", () => {
  it.each([
    ["add", "2+3", "2 + 3 = 5"],
    ["subtract", "10−4", "10 − 4 = 6"],
    ["multiply", "6×7", "6 × 7 = 42"],
    ["divide", "1÷4", "1 ÷ 4 = 0.25"],
    ["power", "2^10", "2 ^ 10 = 1024"],
    ["sqrt", "√9", "√9 = 3"],
    ["percentage of a number", "15%×80", "15% × 80 = 12"],
    ["phone percentage", "80+15%", "80 + 15% = 92"],
    ["multi-operator expression", "2+3×4", "2 + 3 × 4 = 14"],
    ["brackets", "(2+3)×4", "(2 + 3) × 4 = 20"],
    ["negative decimal number", "−1.5×2", "−1.5 × 2 = −3"],
    ["exponent notation", "1e6+1", "1e6 + 1 = 1000001"],
  ])("%s: %s", async (_name, keys, result) => {
    const user = await renderApp();
    await press(user, keys);
    expect(expressionInput()).toHaveValue(keys);
    await user.click(screen.getByRole("button", { name: "equals" }));
    await expectResult(result);
  });

  it("rounds the display to 12 significant digits", async () => {
    const user = await renderApp();
    await press(user, "0.1+0.2");
    await user.click(screen.getByRole("button", { name: "equals" }));
    await expectResult("0.1 + 0.2 = 0.3");
  });

  it("deletes the last character with the delete key", async () => {
    const user = await renderApp();
    await press(user, "2+34");
    await user.click(screen.getByRole("button", { name: "delete" }));
    expect(expressionInput()).toHaveValue("2+3");
  });

  it("clears the expression, the result and the error with AC", async () => {
    const user = await renderApp();
    await press(user, "2+3");
    await user.click(screen.getByRole("button", { name: "equals" }));
    await expectResult("= 5");
    await user.click(screen.getByRole("button", { name: "all clear" }));
    expect(expressionInput()).toHaveValue("");
    expect(screen.getByRole("status")).toHaveTextContent("");
  });

  it("continues from the result when an operator follows a result", async () => {
    const user = await renderApp();
    await press(user, "2+3×4");
    await user.click(screen.getByRole("button", { name: "equals" }));
    await expectResult("= 14");
    await press(user, "×2");
    expect(expressionInput()).toHaveValue("14×2");
    await user.click(screen.getByRole("button", { name: "equals" }));
    await expectResult("14 × 2 = 28");
  });

  it("continues from a negative result with the minus sign of the display", async () => {
    const user = await renderApp();
    await press(user, "−1.5×2");
    await user.click(screen.getByRole("button", { name: "equals" }));
    await expectResult("= −3");
    await press(user, "×2");
    expect(expressionInput()).toHaveValue("−3×2");
  });

  it("starts a new expression when a digit follows a result", async () => {
    const user = await renderApp();
    await press(user, "2+3");
    await user.click(screen.getByRole("button", { name: "equals" }));
    await expectResult("= 5");
    await press(user, "7");
    expect(expressionInput()).toHaveValue("7");
  });

  it("does not open the phone keyboard", async () => {
    await renderApp();
    expect(expressionInput()).toHaveAttribute("inputmode", "none");
  });
});

describe("keyboard", () => {
  it("calculates a typed expression with Enter", async () => {
    const user = await renderApp();
    await user.type(expressionInput(), "2+3*4{Enter}");
    expect(expressionInput()).toHaveValue("2+3×4");
    await expectResult("2 + 3 × 4 = 14");
  });

  it("shows / and - as ÷ and −", async () => {
    const user = await renderApp();
    await user.type(expressionInput(), "1/4-2");
    expect(expressionInput()).toHaveValue("1÷4−2");
  });

  it("ignores characters that are not calculator characters", async () => {
    const user = await renderApp();
    await user.type(expressionInput(), "2a+b3!");
    expect(expressionInput()).toHaveValue("2+3");
  });

  it("continues from the result when an operator is typed after a result", async () => {
    const user = await renderApp();
    await user.type(expressionInput(), "2+3*4{Enter}");
    await expectResult("= 14");
    await user.type(expressionInput(), "*2");
    expect(expressionInput()).toHaveValue("14×2");
  });

  it("starts a new expression when a digit is typed after a result", async () => {
    const user = await renderApp();
    await user.type(expressionInput(), "2+3{Enter}");
    await expectResult("= 5");
    await user.type(expressionInput(), "9");
    expect(expressionInput()).toHaveValue("9");
  });

  it("clears everything with Escape", async () => {
    const user = await renderApp();
    await user.type(expressionInput(), "1/0{Enter}");
    await screen.findByRole("alert");
    await user.keyboard("{Escape}");
    expect(expressionInput()).toHaveValue("");
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });

  it("deletes 1 character with Backspace", async () => {
    const user = await renderApp();
    await user.type(expressionInput(), "2+3{Backspace}");
    expect(expressionInput()).toHaveValue("2+");
  });
});

describe("validation and loading", () => {
  it("shows a message for an empty expression and sends no request", async () => {
    let requests = 0;
    server.events.on("request:start", ({ request }) => {
      if (request.url.endsWith("/calculate")) requests++;
    });
    const user = await renderApp();
    await user.click(screen.getByRole("button", { name: "equals" }));
    expect(await screen.findByRole("alert")).toHaveTextContent("Enter an expression");
    expect(requests).toBe(0);
    server.events.removeAllListeners();
  });

  it("disables the equals key and shows a loading state while a request runs", async () => {
    let requests = 0;
    server.use(
      http.post(calculateUrl, async () => {
        requests++;
        await delay(100);
        return HttpResponse.json({ expression: "2 + 3", result: 5 });
      }),
    );
    const user = await renderApp();
    await press(user, "2+3");
    await user.click(screen.getByRole("button", { name: "equals" }));

    expect(screen.getByRole("button", { name: "equals" })).toBeDisabled();
    expect(screen.getByText("Calculating…")).toBeInTheDocument();
    await user.type(expressionInput(), "{Enter}");

    await expectResult("2 + 3 = 5");
    expect(screen.getByRole("button", { name: "equals" })).toBeEnabled();
    expect(requests).toBe(1);
  });
});

describe("error messages", () => {
  it.each([
    ["INVALID_REQUEST", 400, "The request was not valid"],
    ["REQUEST_TOO_LARGE", 413, "The expression is too long"],
    ["DIVISION_BY_ZERO", 422, "You cannot divide by zero"],
    ["INVALID_DOMAIN", 422, "The input is outside the domain of the operation"],
    ["RESULT_OUT_OF_RANGE", 422, "The result is too large"],
    ["METHOD_NOT_ALLOWED", 405, "Something went wrong"],
    ["INTERNAL_ERROR", 500, "Something went wrong"],
  ])("%s shows the correct message", async (code, status, message) => {
    server.use(
      http.post(calculateUrl, () =>
        HttpResponse.json({ error: { code, message: "raw server text" } }, { status }),
      ),
    );
    const user = await renderApp();
    await press(user, "2+3");
    await user.click(screen.getByRole("button", { name: "equals" }));

    const alert = await screen.findByRole("alert");
    expect(alert).toHaveTextContent(message);
    expect(alert).not.toHaveTextContent("raw server text");
  });

  it("shows the 1-based position for INVALID_EXPRESSION", async () => {
    const user = await renderApp();
    await press(user, "2+");
    await user.click(screen.getByRole("button", { name: "equals" }));
    expect(await screen.findByRole("alert")).toHaveTextContent("character 3");
  });

  it.each([
    ["division by zero", "1÷0", "You cannot divide by zero"],
    ["square root of a negative number", "√−1", "The input is outside the domain"],
    ["result too large", "1e308×10", "The result is too large"],
  ])("%s", async (_name, keys, message) => {
    const user = await renderApp();
    await press(user, keys);
    await user.click(screen.getByRole("button", { name: "equals" }));
    expect(await screen.findByRole("alert")).toHaveTextContent(message);
  });

  it("shows a message when the server is not available", async () => {
    server.use(http.post(calculateUrl, () => HttpResponse.error()));
    const user = await renderApp();
    await press(user, "2+3");
    await user.click(screen.getByRole("button", { name: "equals" }));
    expect(await screen.findByRole("alert")).toHaveTextContent("The server is not available");
  });

  it("shows a message when the proxy cannot reach the API", async () => {
    server.use(
      http.post(calculateUrl, () => new HttpResponse("<html>Bad Gateway</html>", { status: 502 })),
    );
    const user = await renderApp();
    await press(user, "2+3");
    await user.click(screen.getByRole("button", { name: "equals" }));
    expect(await screen.findByRole("alert")).toHaveTextContent("The server is not available");
  });

  it("clears the error when the next calculation succeeds", async () => {
    const user = await renderApp();
    await press(user, "1÷0");
    await user.click(screen.getByRole("button", { name: "equals" }));
    await screen.findByRole("alert");
    await user.click(screen.getByRole("button", { name: "all clear" }));
    await press(user, "2+3");
    await user.click(screen.getByRole("button", { name: "equals" }));
    await expectResult("= 5");
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });
});

describe("history", () => {
  it("keeps the last 10 calculations, the newest first, and clears them", async () => {
    let n = 0;
    server.use(
      http.post(calculateUrl, () => {
        n++;
        return HttpResponse.json({ expression: `${String(n)} + 0`, result: n });
      }),
    );
    const user = await renderApp();
    for (let i = 1; i <= 11; i++) {
      await press(user, "1");
      await user.click(screen.getByRole("button", { name: "equals" }));
      await expectResult(`${String(i)} + 0 = ${String(i)}`);
    }

    const history = screen.getByRole("list", { name: "History" });
    const items = within(history).getAllByRole("listitem");
    expect(items).toHaveLength(10);
    expect(items[0]).toHaveTextContent("11 + 0 = 11");
    expect(items[9]).toHaveTextContent("2 + 0 = 2");

    await user.click(screen.getByRole("button", { name: "Clear history" }));
    await waitFor(() => {
      expect(screen.queryByRole("list", { name: "History" })).not.toBeInTheDocument();
    });
  });

  it("does not add a failed calculation to the history", async () => {
    const user = await renderApp();
    await press(user, "1÷0");
    await user.click(screen.getByRole("button", { name: "equals" }));
    await screen.findByRole("alert");
    expect(screen.queryByRole("list", { name: "History" })).not.toBeInTheDocument();
  });
});
