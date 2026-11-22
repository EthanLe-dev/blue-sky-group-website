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
  await settings.focus();
  await page.keyboard.press("Escape");
  await expect(menu).toBeFocused();
  await expect(navigation).toBeHidden();
  await menu.click();
  await settings.click();
  await expect(navigation).toBeHidden();
  await expect(page.getByRole("main")).toBeFocused();
  expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(390);
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
