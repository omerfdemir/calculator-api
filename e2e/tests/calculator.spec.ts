import AxeBuilder from "@axe-core/playwright";
import { expect, test, type Page } from "@playwright/test";

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

/** Clicks (or taps) the keypad key for each character of keys. */
async function press(page: Page, keys: string, { tap = false } = {}) {
  for (const key of keys) {
    const button = page.getByRole("button", { name: keyNames[key] ?? key, exact: true });
    await (tap ? button.tap() : button.click());
  }
}

async function equals(page: Page, { tap = false } = {}) {
  const button = page.getByRole("button", { name: "equals" });
  await (tap ? button.tap() : button.click());
}

test.beforeEach(async ({ page }) => {
  await page.goto("/");
  await expect(page.getByRole("button", { name: "equals" })).toBeVisible();
});

test("shows a key for each of the 7 operations from the real API", async ({ page }) => {
  for (const name of ["plus", "minus", "times", "divide", "power", "square root", "percent"]) {
    await expect(page.getByRole("button", { name, exact: true })).toBeVisible();
  }
});

test.describe("happy path for each operation", () => {
  const cases = [
    ["add", "2+3", "2 + 3 = 5"],
    ["subtract", "10−4", "10 − 4 = 6"],
    ["multiply", "6×7", "6 × 7 = 42"],
    ["divide", "1÷4", "1 ÷ 4 = 0.25"],
    ["power", "2^10", "2 ^ 10 = 1024"],
    ["sqrt", "√9", "√9 = 3"],
    ["percentage", "15%×80", "15% × 80 = 12"],
  ] as const;

  for (const [name, keys, result] of cases) {
    test(name, async ({ page }) => {
      await press(page, keys);
      await equals(page);
      await expect(page.getByRole("status")).toHaveText(result);
    });
  }
});

test("calculates a multi-operator expression with brackets", async ({ page }) => {
  await press(page, "(2+3)×4−6÷3");
  await equals(page);
  await expect(page.getByRole("status")).toHaveText("(2 + 3) × 4 − 6 ÷ 3 = 18");
});

test("adds a percentage like a phone calculator", async ({ page }) => {
  await press(page, "80+15%");
  await equals(page);
  await expect(page.getByRole("status")).toHaveText("80 + 15% = 92");
});

test("shows 0.1 + 0.2 as 0.3 and continues from the result", async ({ page }) => {
  await press(page, "0.1+0.2");
  await equals(page);
  await expect(page.getByRole("status")).toHaveText("0.1 + 0.2 = 0.3");
  await press(page, "×10");
  await expect(page.getByRole("textbox", { name: "Expression" })).toHaveValue("0.3×10");
  await equals(page);
  await expect(page.getByRole("status")).toHaveText("0.3 × 10 = 3");
  await expect(page.getByRole("list", { name: "History" }).getByRole("listitem")).toHaveText([
    "0.3 × 10 = 3",
    "0.1 + 0.2 = 0.3",
  ]);
});

test.describe("error messages", () => {
  const cases = [
    ["division by zero", "1÷0", "You cannot divide by zero."],
    ["square root of a negative number", "√−1", "outside the domain of the operation"],
    ["result too large", "1e308×10", "The result is too large to show."],
    ["syntax error", "2×(1+2", "Check character 3."],
  ] as const;

  for (const [name, keys, message] of cases) {
    test(name, async ({ page }) => {
      await press(page, keys);
      await equals(page);
      await expect(page.getByRole("alert")).toContainText(message);
    });
  }
});

test("works with the keyboard only", async ({ page }) => {
  await page.keyboard.press("Tab");
  await expect(page.getByRole("textbox", { name: "Expression" })).toBeFocused();
  await page.keyboard.type("2+3*4");
  await page.keyboard.press("Enter");
  await expect(page.getByRole("status")).toHaveText("2 + 3 × 4 = 14");

  // Continue from the result. The keypad keys are in the tab order after the input.
  await page.keyboard.type("-4");
  await page.keyboard.press("Tab");
  await expect(page.getByRole("button", { name: "all clear" })).toBeFocused();
  await page.keyboard.press("Shift+Tab");
  await page.keyboard.press("Enter");
  await expect(page.getByRole("status")).toHaveText("14 − 4 = 10");

  await page.keyboard.press("Escape");
  await expect(page.getByRole("textbox", { name: "Expression" })).toHaveValue("");
  await expect(page.getByRole("status")).toHaveText("");
});

test.describe("mobile viewport 360 × 800", () => {
  test.use({ viewport: { width: 360, height: 800 }, hasTouch: true, isMobile: true });

  test("has no horizontal scroll, and the main flow works", async ({ page }) => {
    const overflow = await page.evaluate(
      () => document.documentElement.scrollWidth - document.documentElement.clientWidth,
    );
    expect(overflow).toBe(0);

    await press(page, "2+3×4", { tap: true });
    await equals(page, { tap: true });
    await expect(page.getByRole("status")).toHaveText("2 + 3 × 4 = 14");

    const overflowAfter = await page.evaluate(
      () => document.documentElement.scrollWidth - document.documentElement.clientWidth,
    );
    expect(overflowAfter).toBe(0);
  });
});

test.describe("accessibility", () => {
  for (const colorScheme of ["light", "dark"] as const) {
    test(`has no serious or critical axe violations (${colorScheme})`, async ({ page }) => {
      await page.emulateMedia({ colorScheme });
      await page.reload();
      await expect(page.getByRole("button", { name: "equals" })).toBeVisible();

      // Show a result, the history and an error, so that the scan covers them.
      await press(page, "2+3");
      await equals(page);
      await expect(page.getByRole("status")).toHaveText("2 + 3 = 5");
      await press(page, "1÷0");
      await equals(page);
      await expect(page.getByRole("alert")).toBeVisible();

      const results = await new AxeBuilder({ page })
        .withTags(["wcag2a", "wcag2aa", "wcag21a", "wcag21aa"])
        .analyze();
      const serious = results.violations.filter(
        (v) => v.impact === "serious" || v.impact === "critical",
      );
      expect(serious.map((v) => `${v.id}: ${v.help}`)).toEqual([]);
    });
  }
});

test.describe("nginx", () => {
  test("a page reload on a deep link shows the app", async ({ page }) => {
    await page.goto("/unknown/path");
    await expect(page.getByRole("button", { name: "equals" })).toBeVisible();
    await page.reload();
    await expect(page.getByRole("button", { name: "equals" })).toBeVisible();
    await press(page, "6×7");
    await equals(page);
    await expect(page.getByRole("status")).toHaveText("6 × 7 = 42");
  });

  test("a request to /api/v1/operations returns the operations list", async ({ request }) => {
    const response = await request.get("/api/v1/operations");
    expect(response.status()).toBe(200);
    expect(response.headers()["content-type"]).toBe("application/json");
    const body = (await response.json()) as { operations: { name: string }[] };
    expect(body.operations.map((op) => op.name)).toEqual([
      "add",
      "subtract",
      "multiply",
      "divide",
      "power",
      "sqrt",
      "percentage",
    ]);
  });

  test("an unknown API path returns the JSON NOT_FOUND error", async ({ request }) => {
    const response = await request.get("/api/v1/unknown");
    expect(response.status()).toBe(404);
    expect(response.headers()["content-type"]).toBe("application/json");
    const body = (await response.json()) as { error: { code: string } };
    expect(body.error.code).toBe("NOT_FOUND");
  });

  test("sets the security and cache headers", async ({ request }) => {
    const response = await request.get("/");
    const headers = response.headers();
    expect(headers["x-content-type-options"]).toBe("nosniff");
    expect(headers["referrer-policy"]).toBeTruthy();
    expect(headers["content-security-policy"]).toContain("default-src 'self'");
    expect(headers["cache-control"]).toBe("no-cache");
  });
});
