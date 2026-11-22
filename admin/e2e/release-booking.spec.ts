import { test, expect } from "@playwright/test";

const site = "http://localhost:3000";

test("deposit confirmation survives a lost response, then balance and lookup work", async ({ page }) => {
  await page.goto(site + "/tours/ninh-chu-2-ngay-1-dem");
  await page.getByRole("link", { name: "Đặt tour ngày này" }).first().click();
  await page.getByRole("button", { name: "Xem báo giá", exact: true }).click();
  const email = "m7-deposit@example.com";
  await page.getByLabel("Họ tên người liên hệ", { exact: true }).fill("Khách demo M7");
  await page.getByLabel("Email", { exact: true }).fill(email);
  await page.getByLabel("Số điện thoại", { exact: true }).fill("0900000000");
  await page.getByLabel("Họ tên hành khách 1", { exact: true }).fill("Khách demo M7");
  await page.getByLabel("Ngày sinh hành khách 1", { exact: true }).fill("1990-01-01");
  await page.getByRole("button", { name: "Kiểm tra đơn", exact: true }).click();
  await page.getByLabel("Phương án thanh toán").selectOption("deposit");
  await page.getByRole("checkbox").check();
  await page.getByRole("button", { name: "Tạo đơn giữ chỗ", exact: true }).click();
  await page.getByRole("button", { name: "Đặt cọc giả lập", exact: true }).click();

  const confirmation = page.getByRole("button", { name: "Xác nhận thanh toán giả lập", exact: true });
  const keys: string[] = [];
  await page.route("**/payments/*/mock-confirm", async route => {
    keys.push(route.request().headers()["idempotency-key"]);
    if (keys.length === 1) {
      const upstream = await route.fetch();
      expect(upstream.status()).toBe(200);
      await route.abort("failed");
    } else await route.continue();
  });
  await confirmation.click();
  await expect(page.getByRole("alert")).toBeVisible();
  await confirmation.click();
  await expect(page.getByRole("heading", { name: "Đã xác nhận đặt tour" })).toBeVisible();
  await expect(page.locator(".summary")).toContainText("Đã đặt cọc");
  expect(keys).toHaveLength(2);
  expect(keys[0]).toBeTruthy();
  expect(keys[0]).toBe(keys[1]);
  await page.getByRole("button", { name: "Thanh toán phần còn lại giả lập", exact: true }).click();
  await confirmation.click();
  await expect(page.locator(".summary")).toContainText("Đã thanh toán đủ");
  await expect(page.getByText("SMS giả lập · Thanh toán phần còn lại")).toBeVisible({ timeout: 15_000 });
  const code = await page.locator(".summary strong").first().innerText();
  await page.goto(site + "/lookup");
  await page.getByLabel("Mã đơn", { exact: true }).fill(code);
  await page.getByLabel("Email hoặc số điện thoại", { exact: true }).fill(email);
  await page.getByRole("button", { name: "Tra cứu đơn", exact: true }).click();
  await expect(page.locator(".summary")).toContainText("Đã thanh toán đủ");
});

test("private release routes reject anonymous requests", async ({ request }) => {
  for (const path of ["/admin/bookings/export", "/admin/reports/revenue", "/admin/group-quotes/00000000-0000-4000-8000-000000000000/document.pdf"]) {
    const response = await request.get("/api/v1" + path);
    expect(response.status(), path).toBe(401);
    expect(response.headers()["cache-control"]).toContain("no-store");
  }
});
