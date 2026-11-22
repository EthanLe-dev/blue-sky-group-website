import { test, expect, type Page } from "@playwright/test";
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
async function login(page: Page) {
  await page.goto("/");
  await page
    .getByLabel("Email", { exact: true })
    .fill(env.BOOTSTRAP_ADMIN_EMAIL);
  await page
    .getByLabel("Mật khẩu", { exact: true })
    .fill(env.BOOTSTRAP_ADMIN_PASSWORD);
  await page.getByRole("button", { name: "Đăng nhập", exact: true }).click();
  await expect(
    page.getByRole("heading", { name: "Danh mục và nội dung" }),
  ).toBeVisible();
}
test("catalog, poster, favourites, checkout, private lookup and admin cancellation", async ({
  page,
}) => {
  await page.goto("http://localhost:3000/tours?budget_max=1700000");
  await expect(page.locator(".tour-card")).not.toHaveCount(0);
  await page.goto("http://localhost:3000/tours/ninh-chu-2-ngay-1-dem");
  await expect(page.locator("main h1")).toContainText("Ninh Chữ");
  await page.route("https://tile.openstreetmap.org/**", (route) =>
    route.fulfill({
      contentType: "image/png",
      body: Buffer.from(
        "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+jNioAAAAASUVORK5CYII=",
        "base64",
      ),
    }),
  );
  await page
    .getByRole("button", { name: "Mở bản đồ điểm đến", exact: true })
    .click();
  await expect(page.locator(".leaflet-marker-icon").first()).toBeVisible();
  await page
    .getByRole("combobox", { name: "Chọn điểm trên bản đồ", exact: true })
    .selectOption("1");
  await page.getByRole("button", { name: "Ẩn bản đồ", exact: true }).click();
  await page.getByRole("button", { name: "♡ Lưu tour" }).click();
  await expect(
    page.getByRole("button", { name: "♥ Đã lưu tour" }),
  ).toBeVisible();
  await page.getByRole("button", { name: /Xem poster đầy đủ/ }).click();
  await expect(page.getByRole("dialog")).toBeVisible();
  await page.keyboard.press("Escape");
  await page.getByRole("link", { name: "Đặt tour ngày này" }).first().click();
  await page.getByRole("button", { name: "Xem báo giá", exact: true }).click();
  await page
    .getByLabel("Họ tên người liên hệ", { exact: true })
    .fill("Khách Kiểm Thử");
  await page
    .getByLabel("Email", { exact: true })
    .fill("browser-demo@example.com");
  await page.getByLabel("Số điện thoại", { exact: true }).fill("0901234567");
  await page
    .getByLabel("Họ tên hành khách 1", { exact: true })
    .fill("Khách Kiểm Thử");
  await page
    .getByLabel("Ngày sinh hành khách 1", { exact: true })
    .fill("1990-01-01");
  await page.getByRole("button", { name: "Kiểm tra đơn", exact: true }).click();
  await page.getByRole("checkbox").check();
  await page
    .getByRole("button", { name: "Tạo đơn giữ chỗ", exact: true })
    .click();
  await expect(
    page.getByRole("heading", { name: "Đã tạo đơn giữ chỗ" }),
  ).toBeVisible();
  const code = await page.locator(".summary strong").first().innerText();
  await page.goto("http://localhost:3000/lookup");
  await page.getByLabel("Mã đơn", { exact: true }).fill(code);
  await page
    .getByLabel("Email hoặc số điện thoại", { exact: true })
    .fill("browser-demo@example.com");
  await page.getByRole("button", { name: "Tra cứu đơn", exact: true }).click();
  await expect(page.locator(".summary")).toContainText("b***@example.com");
  await expect(page.locator(".summary")).not.toContainText("Khách Kiểm Thử");
  await login(page);
  await page.getByRole("button", { name: "Đơn tour", exact: true }).click();
  await page.getByLabel("Mã đơn / tên liên hệ", { exact: true }).fill(code);
  await page.getByRole("button", { name: "Tìm đơn", exact: true }).click();
  await page.locator(".list-row").filter({ hasText: code }).click();
  page.once("dialog", (d) => d.accept("Dọn đơn E2E demo"));
  await page
    .getByRole("button", { name: "Huỷ đơn và trả chỗ", exact: true })
    .click();
  await expect(page.getByRole("status").first()).toContainText("Đã huỷ đơn");
});
test("admin tour editor persists edits and verifies poster", async ({
  page,
}) => {
  await login(page);
  await page
    .locator(".list-row")
    .filter({ hasText: "Ninh Chữ · Biển và văn hoá Chăm" })
    .first()
    .click();
  const summary = page.getByRole("textbox", { name: "Tóm tắt", exact: true });
  const old = await summary.inputValue();
  try {
    await summary.fill(old + " Kiểm thử lưu.");
    await page
      .getByRole("button", { name: "Lưu nội dung", exact: true })
      .click();
    await expect(
      page.getByRole("button", { name: "Lưu nội dung", exact: true }),
    ).toBeEnabled();
    await expect(page.getByRole("status").first()).toContainText(
      "Đã lưu nội dung",
    );
    await page.getByRole("button", { name: "Đóng trình sửa" }).click();
    await page
      .locator(".list-row")
      .filter({ hasText: "Ninh Chữ · Biển và văn hoá Chăm" })
      .first()
      .click();
    await expect(summary).toHaveValue(old + " Kiểm thử lưu.");
  } finally {
    await summary.fill(old);
    await page
      .getByRole("button", { name: "Lưu nội dung", exact: true })
      .click();
    await expect(
      page.getByRole("button", { name: "Lưu nội dung", exact: true }),
    ).toBeEnabled();
    await expect(page.getByRole("status").first()).toContainText(
      "Đã lưu nội dung",
    );
    await page
      .getByRole("button", { name: "Xác minh poster khớp dữ liệu" })
      .click();
    await expect(
      page.getByText("Đã xác minh poster.", { exact: true }),
    ).toBeVisible();
  }
  await page.getByRole("button", { name: "Media", exact: true }).click();
  await expect(
    page.getByRole("heading", { name: "Thư viện media" }),
  ).toBeVisible();
});
test("responsive menu and public content pages", async ({ page }) => {
  for (const width of [320, 390, 768, 820, 1024, 1440]) {
    await page.setViewportSize({ width, height: 900 });
    await page.emulateMedia({ reducedMotion: "reduce" });
    await page.goto("http://localhost:3000/");
    await expect(page.locator(".hero h1")).toBeVisible();
    expect(
      await page.evaluate(
        () => document.documentElement.scrollWidth <= innerWidth,
      ),
    ).toBe(true);
    if (width < 850) {
      await page.getByRole("button", { name: "Menu" }).click();
      await expect(page.getByRole("dialog")).toBeVisible();
      await page.keyboard.press("Escape");
      await expect(page.getByRole("button", { name: "Menu" })).toBeFocused();
    }
  }
  for (const path of [
    "/about",
    "/contact",
    "/schools",
    "/companies",
    "/blog",
    "/destinations",
    "/faq",
    "/gallery",
    "/policies/privacy",
  ]) {
    const r = await page.goto("http://localhost:3000" + path);
    expect(r?.status(), path).toBe(200);
    await expect(page.locator("main h1")).toBeVisible();
  }
  const r = await page.goto("http://localhost:3000/tours/not-existing");
  expect(r?.status()).toBe(404);
});

test("home search, blog facets and media filters", async ({ page }) => {
  await page.goto("http://localhost:3000/");
  await page.getByLabel("Bạn muốn đi đâu?", { exact: true }).fill("Ninh Chữ");
  await page.locator(".hero").getByRole("button", { name: "Tìm tour" }).click();
  await expect(page).toHaveURL(/q=/);
  await expect(page.locator(".tour-card").first()).toBeVisible();
  await page.getByText("Bộ lọc nâng cao", { exact: true }).click();
  await expect(
    page.getByRole("combobox", { name: "Danh mục tour", exact: true }),
  ).toBeVisible();
  await page.goto("http://localhost:3000/blog");
  await expect(
    page.getByRole("combobox", { name: "Danh mục", exact: true }),
  ).toBeVisible();
  await page.getByRole("button", { name: "Lọc bài viết" }).click();
  await expect(page.locator("main .content-list a").first()).toBeVisible();
  await page.goto("http://localhost:3000/gallery?kind=video");
  await expect(
    page.getByText("Chưa có media phù hợp bộ lọc.", { exact: true }),
  ).toBeVisible();
  await page.goto("http://localhost:3000/gallery?kind=image");
  await expect(page.locator(".gallery-grid img").first()).toBeVisible();
});
