import { expect, test } from "@playwright/test";

// Takes the README screenshot. Run it with `npm run screenshot` while the stack runs.
test.skip(!process.env.SCREENSHOT, "Set SCREENSHOT=1 to take the README screenshot.");

test.use({ viewport: { width: 420, height: 900 }, colorScheme: "dark" });

test("README screenshot", async ({ page }) => {
  await page.goto("/");
  const expression = page.getByRole("textbox", { name: "Expression" });
  const status = page.getByRole("status");

  for (const [input, result] of [
    ["0.1+0.2", "0.1 + 0.2 = 0.3"],
    ["(2+3)*4-6/3", "(2 + 3) × 4 − 6 ÷ 3 = 18"],
    ["80+-15%", "80 + −15% = 68"],
    ["1/3", "1 ÷ 3 = 0.333333333333"],
  ] as const) {
    await expression.fill(input);
    await expression.press("Enter");
    await expect(status).toHaveText(result);
  }

  // Continue from the full result: the display shows 0.333333333333, but 1 is the answer.
  await expression.pressSequentially("*3");
  await expression.press("Enter");
  await expect(status).toHaveText("0.333333333333 × 3 = 1");

  await page.screenshot({ path: "../docs/screenshot.png", fullPage: true });
});
