import { test, expect } from "@playwright/test";

test("MVP shell works with keyboard and reduced motion at mobile, tablet and desktop widths", async ({ page }) => {
  const problems: string[] = [];
  page.on("pageerror", error => problems.push(error.message));
  await page.emulateMedia({ reducedMotion: "reduce" });
  for (const [width, height] of [[320, 720], [390, 844], [768, 1024], [1024, 768], [1440, 900]]) {
    await page.setViewportSize({ width, height });
    await page.goto("http://localhost:3000/");
    const hero = page.locator(".hero");
    await expect(hero.getByRole("heading", { name: "Khám phá những hành trình cùng Blue Sky Group" })).toBeVisible();
    await expect(hero).not.toContainText(/Ninh Chữ|Vĩnh Hy|ngày.*đêm/i);
    await expect(hero.getByRole("link", { name: /Ninh Chữ|Vĩnh Hy/i })).toHaveCount(0);
    await expect(page.getByRole("heading", { name: "Khám phá các tour" })).toBeVisible();
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
    const photo = page.locator("img.hero-photo");
    await expect.poll(() => photo.evaluate(img => (img as HTMLImageElement).naturalWidth)).toBeGreaterThan(0);
    await expect(page.locator(".hero h1")).toHaveCSS("opacity", "1");
    await expect(page.locator(".hero h1")).toHaveCSS("font-family", /Nunito Sans Variable/);
    expect(await page.evaluate(async () => {
      const faces = await document.fonts.load('800 32px "Nunito Sans Variable"', "Ắ Ằ Ẳ Ẵ Ặ");
      return faces.some(face => face.status === "loaded");
    })).toBe(true);
    const contrast = await page.locator(".hero input").evaluate(input => {
      const luminance = (color: string) => {
        const values = (color.match(/[0-9]+(?:\.[0-9]+)?/g) || []).slice(0, 3).map(Number);
        return values.reduce((sum, value, i) => {
          const channel = value / 255;
          const linear = channel <= .04045 ? channel / 12.92 : ((channel + .055) / 1.055) ** 2.4;
          return sum + linear * [.2126, .7152, .0722][i];
        }, 0);
      };
      const text = luminance(getComputedStyle(input, "::placeholder").color);
      const background = luminance(getComputedStyle(input).backgroundColor);
      return (Math.max(text, background) + .05) / (Math.min(text, background) + .05);
    });
    expect(contrast).toBeGreaterThanOrEqual(4.5);
    await page.evaluate(() => scrollTo(0, 0));
    await page.screenshot({ path: test.info().outputPath("homepage-" + width + ".png"), animations: "disabled" });
  }
  await page.locator('#journeys a[href="/tours/ninh-chu-2-ngay-1-dem"]').click();
  await expect(page).toHaveURL(/ninh-chu-2-ngay-1-dem$/);
  await expect(page.getByRole("link", { name: "Tải PDF lịch trình" })).toBeVisible();
  await page.getByRole("button", { name: /Xem poster đầy đủ/ }).click();
  await expect(page.getByRole("dialog")).toBeVisible();
  await page.keyboard.press("Escape");
  await expect(page.getByRole("dialog")).not.toBeVisible();
  await page.route("https://tile.openstreetmap.org/**", route => route.fulfill({
    contentType: "image/png",
    body: Buffer.from("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+jNioAAAAASUVORK5CYII=", "base64"),
  }));
  await page.getByRole("link", { name: "Liên hệ ↗", exact: true }).click();
  await expect(page).toHaveURL(/\/contact$/);
  await expect(page.getByText("Vị trí minh họa. Vui lòng xác nhận địa chỉ trước khi đến.")).toBeVisible();
  await page.getByRole("button", { name: "Mở bản đồ điểm đến" }).click();
  await expect(page.locator(".leaflet-marker-icon")).toBeVisible();
  expect(problems).toEqual([]);
});
