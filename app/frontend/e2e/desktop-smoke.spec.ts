import { expect, test } from "@playwright/test";

test("loads the local desktop shell", async ({ page }) => {
  await page.goto("/");

  await expect(page).toHaveTitle("Sweeters");
  await expect(page.getByRole("heading", { name: "Painel", exact: true })).toBeVisible();
});
