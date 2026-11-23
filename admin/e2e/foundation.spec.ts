import { test, expect } from "@playwright/test";
import { readFileSync } from "node:fs";
const env = Object.fromEntries(
  readFileSync("../.env", "utf8")
    .split(/\r?\n/)
    .filter((l) => l && !l.startsWith("#"))
    .map((l) => {
      const i = l.indexOf("=");
      return [l.slice(0, i), l.slice(i + 1)];
    }),
);
test("admin login, settings save, reload and logout", async ({ page }) => {
  await page.goto("/");
  await page
    .getByLabel("Email", { exact: true })
    .fill(env.BOOTSTRAP_ADMIN_EMAIL);
  await page
    .getByLabel("Mật khẩu", { exact: true })
    .fill(env.BOOTSTRAP_ADMIN_PASSWORD);
  await page.getByRole("button", { name: "Đăng nhập", exact: true }).click();
  await page.getByRole("button", { name: "Cài đặt", exact: true }).click();
  await expect(
    page.getByRole("heading", { name: "Cài đặt demo" }),
  ).toBeVisible();
  const navigation = page.getByRole("navigation", { name: "Module quản trị" });
  const reports = navigation.getByRole("button", { name: "Báo cáo" });
  const settings = navigation.getByRole("button", { name: "Cài đặt" });
  const reportPosition = (await reports.boundingBox())!;
  const settingsPosition = (await settings.boundingBox())!;
  const workspace = (await page.getByRole("main").boundingBox())!;
  expect(reportPosition.x).toBe(settingsPosition.x);
  expect(reportPosition.x + reportPosition.width).toBeLessThan(workspace.x);
  expect(reportPosition.y).toBeLessThan(settingsPosition.y);
  await expect(settings).toHaveAttribute("aria-pressed", "true");
  await page.mouse.move(900, 0);
  await expect(page.locator(".admin-sidebar")).toHaveCSS("background-color", "rgb(18, 59, 86)");
  await expect(settings).toHaveCSS("background-color", "rgb(244, 200, 76)");
  await settings.hover();
  await expect(settings).toHaveCSS("color", "rgb(18, 59, 86)");

  await expect(page.getByRole("main")).toHaveCSS("font-family", /Nunito Sans Variable/);
  expect(await page.evaluate(async () => {
    const faces = await document.fonts.load('700 24px "Nunito Sans Variable"', "Nguyễn Thị Ánh · Ắ Ằ Ẳ Ẵ Ặ · 1.250.000 ₫");
    return faces.some(face => face.status === "loaded");
  })).toBe(true);
  await page.screenshot({ path: test.info().outputPath("admin-desktop.png"), animations: "disabled" });
  const hold = page.getByLabel("Thời gian giữ chỗ (phút)", { exact: true });
  await expect(hold).toBeVisible();
  await hold.fill("21");
  await page.getByRole("button", { name: "Lưu cấu hình" }).click();
  await expect(page.getByRole("status")).toContainText("Đã lưu");
  await page.reload();
  await page.getByRole("button", { name: "Cài đặt", exact: true }).click();
  await expect(hold).toHaveValue("21");
  await hold.fill("20");
  await page.getByRole("button", { name: "Lưu cấu hình" }).click();
  await expect(page.getByRole("status")).toContainText("Đã lưu");
  await page.setViewportSize({ width: 390, height: 844 });
  await expect(
    page.getByRole("button", { name: "Lưu cấu hình" }),
  ).toBeVisible();
  const menu = page.getByRole("button", { name: "Menu", exact: true });
  await expect(navigation).toBeHidden();
  await menu.click();
  await expect(settings).toBeVisible();
  await expect(page.getByRole("dialog", { name: "Điều hướng quản trị" })).toBeVisible();
  await page.screenshot({ path: test.info().outputPath("admin-mobile.png"), animations: "disabled" });
  await settings.focus();
  await page.keyboard.press("Escape");
  await expect(menu).toBeFocused();
  await expect(navigation).toBeHidden();
  await menu.click();
  await settings.click();
  await expect(navigation).toBeHidden();
  await expect(page.getByRole("main")).toBeFocused();
  expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(390);
  const bookingRequests = "**/api/v1/admin/bookings?*";
  let respond!: () => void;
  const pending = new Promise<void>(resolve => { respond = resolve; });
  await page.route(bookingRequests, async route => {
    await pending;
    await route.fulfill({ status: 503, json: { success: false } });
  });
  try {
    await menu.click();
    await navigation.getByRole("button", { name: "Đơn tour", exact: true }).click();
    await expect(page.getByText("Đang tải danh sách đơn…")).toBeVisible();
  } finally {
    respond();
  }
  await expect(page.getByRole("alert")).toHaveText("Không tải được danh sách đơn.");
  await page.unroute(bookingRequests);
  await page.route(bookingRequests, route => route.fulfill({ json: { success: true, data: [] } }));
  await page.getByRole("button", { name: "Thử lại" }).click();
  await expect(page.getByText("Chưa có đơn phù hợp.")).toBeVisible();
  await page.unroute(bookingRequests);
  await page.getByRole("button", { name: "Đăng xuất" }).click();
  await expect(
    page.getByRole("heading", { name: "Đăng nhập quản trị" }),
  ).toBeVisible();
});
test("public proxy forwards guest cookie and CSRF", async ({ request }) => {
  const r = await request.post("http://localhost:3000/api/v1/guest-sessions", {
    headers: { Origin: "http://localhost:3000" },
    data: {},
  });
  expect(r.status()).toBe(201);
  expect(r.headers()["set-cookie"]).toContain("HttpOnly");
  expect((await r.json()).data.csrf_token).toBeTruthy();
  const blocked = await request.get("http://localhost:3000/_next/image", {
    params: {
      url: "/api/v1/media/00000000-0000-0000-0000-000000000001?expires=9999999999&signature=private",
      w: "640",
      q: "75",
    },
  });
  expect(blocked.status()).toBe(400);
});
