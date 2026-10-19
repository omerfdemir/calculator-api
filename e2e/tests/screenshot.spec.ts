import { expect, test } from "@playwright/test";

// Takes the README screenshot. Run it with `npm run screenshot` while the stack runs.
test.skip(!process.env.SCREENSHOT, "Set SCREENSHOT=1 to take the README screenshot.");

test.use({ viewport: { width: 420, height: 890 }, colorScheme: "dark" });

test("README screenshot", async ({ page }) => {
  await page.goto("/");
  const expression = page.getByRole("textbox", { name: "Expression" });
  const status = page.getByRole("status");

  for (const [input, result] of [
    ["0.1+0.2", "0.1 + 0.2 = 0.3"],
    ["2+3*4", "2 + 3 × 4 = 14"],
    ["(2+3)*4", "(2 + 3) × 4 = 20"],
    ["15%*80", "15% × 80 = 12"],
    ["80+15%", "80 + 15% = 92"],
    ["√9+2^3", "√9 + 2 ^ 3 = 11"],
    ["1e6/4", "1e6 ÷ 4 = 250000"],
    ["1-4", "1 − 4 = −3"],
  ] as const) {
    await expression.fill(input);
    await expression.press("Enter");
    await expect(status).toHaveText(result);
  }

  // Continue from the negative result: it goes in brackets, so ^ raises the whole value.
  await expression.pressSequentially("^2");
  await expect(expression).toHaveValue("(−3)^2");
  await expression.press("Enter");
  await expect(status).toHaveText("(−3) ^ 2 = 9");

  await page.screenshot({ path: "../docs/screenshot.png", fullPage: true });
});
