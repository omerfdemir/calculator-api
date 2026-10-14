import { expect, test } from "@playwright/test";

// Takes the README screenshot. Run it with `npm run screenshot` while the stack runs.
test.skip(!process.env.SCREENSHOT, "Set SCREENSHOT=1 to take the README screenshot.");

test.use({ viewport: { width: 420, height: 860 }, colorScheme: "dark" });

test("README screenshot", async ({ page }) => {
  await page.goto("/");
  const expression = page.getByRole("textbox", { name: "Expression" });
  for (const input of ["0.1+0.2", "√(9)^2", "(2+3)*4-6/3", "80+15%"]) {
    await expression.fill(input);
    await expression.press("Enter");
    await expect(page.getByRole("status")).toContainText("=");
  }
  await page.screenshot({ path: "../docs/screenshot.png", fullPage: true });
});
