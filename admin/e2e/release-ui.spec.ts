import { test, expect } from "@playwright/test";

test("MVP shell works with keyboard and reduced motion at mobile, tablet and desktop widths", async ({ page }) => {
  const problems: string[] = [];
  page.on("pageerror", error => problems.push(error.message));
  await page.emulateMedia({ reducedMotion: "reduce" });
  for (const [width, height] of [[390, 844], [768, 1024], [1440, 900]]) {
    await page.setViewportSize({ width, height });
    await page.goto("http://localhost:3000/");
    await expect(page.getByRole("main").getByRole("heading", { level: 1 })).toBeVisible();
    await page.keyboard.press("Tab");
    const skip = page.getByRole("link", { name: "Bỏ qua đến nội dung" });
    await expect(skip).toBeFocused();
    await page.keyboard.press("Enter");
    await expect(page).toHaveURL(/#main$/);
    expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(width);
    if (width < 850) {
      const menu = page.getByRole("button", { name: "Menu" });
      await menu.focus();
      await page.keyboard.press("Enter");
      await expect(page.getByRole("dialog")).toBeVisible();
      await page.keyboard.press("Escape");
      await expect(menu).toBeFocused();
    }
    await page.goto("http://localhost:3000/");
    await page.screenshot({ path: test.info().outputPath("homepage-" + width + ".png"), animations: "disabled" });
  }
  await page.goto("http://localhost:3000/tours/ninh-chu-2-ngay-1-dem");
  await expect(page.getByRole("link", { name: "Tải PDF lịch trình" })).toBeVisible();
  await page.getByRole("button", { name: /Xem poster đầy đủ/ }).click();
  await expect(page.getByRole("dialog")).toBeVisible();
  await page.keyboard.press("Escape");
  await expect(page.getByRole("dialog")).not.toBeVisible();
  expect(problems).toEqual([]);
});
