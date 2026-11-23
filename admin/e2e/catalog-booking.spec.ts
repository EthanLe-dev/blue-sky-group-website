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
test("catalog, mock payment, private lookup and verified customer history", async ({
  page,
}) => {
  const analytics = () => page.evaluate(() => (window as typeof window & { dataLayer?: Array<ArrayLike<unknown>> }).dataLayer?.map(e => Array.from(e)) || []);
  await page.route("**/measurement", route => route.fulfill({ json: { id: "G-DEMO123", debug: true } }));
  await page.route("https://www.googletagmanager.com/gtag/js?*", route => route.fulfill({ contentType: "application/javascript", body: "" }));
  await page.goto("http://localhost:3000/tours?budget_max=1700000");
  await expect(page.locator(".tour-card")).not.toHaveCount(0);
  await page.goto("http://localhost:3000/tours/ninh-chu-2-ngay-1-dem");
  await expect(page.locator("main h1")).toContainText("Ninh Chữ");
  const pdf = await page.request.get("http://localhost:3000/api/v1/tours/ninh-chu-2-ngay-1-dem/itinerary.pdf");
  expect(pdf.headers()["content-disposition"]).toContain("attachment");
  expect((await pdf.body()).subarray(0, 5).toString()).toBe("%PDF-");
  await expect(page.getByRole("button", { name: "Cho phép phân tích" })).toBeVisible();
  expect(await analytics()).toEqual([]);
  await page.getByRole("button", { name: "Cho phép phân tích" }).click();
  await expect.poll(async () => (await analytics()).map(e => e[1])).toContain("view_item");
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
  await page.getByRole("button", { name: "Thanh toán toàn bộ giả lập", exact: true }).click();
  await expect(page.getByRole("img", { name: "Mã QR chỉ chứa thông báo giả lập" })).toBeVisible();
  await page.getByRole("button", { name: "Xác nhận thanh toán giả lập", exact: true }).evaluate((node: HTMLButtonElement) => {
    node.click();
    node.click();
  });
  await expect(page.getByRole("heading", { name: "Đã xác nhận đặt tour" })).toBeVisible();
  await expect(page.locator(".summary")).toContainText("Đã thanh toán đủ");
  await expect(page.locator("summary").filter({ hasText: "Email giả lập · Xác nhận đặt tour" })).toBeVisible({ timeout: 15_000 });
  await page.getByRole("link", { name: "Mở lại đơn trên trình duyệt này" }).click();
  await expect(page.getByRole("heading", { name: "Đã xác nhận đặt tour" })).toBeVisible();
  await expect.poll(async () => (await analytics()).filter(e => e[1] === "purchase").length).toBe(1);
  const events = await analytics();
  expect(events.some(e => e[1] === "begin_checkout")).toBe(true);
  expect(events.find(e => e[1] === "purchase")?.[2]).toMatchObject({ currency: "VND", is_demo: true });
  expect(JSON.stringify(events)).not.toMatch(/browser-demo@example\.com|0901234567|Khách Kiểm Thử|\?booking=/u);
  expect(JSON.stringify(events)).not.toContain(code);
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
  await page.getByLabel("Số tiền hoàn (₫)", { exact: true }).fill("1000");
  await page.getByLabel("Lý do hoàn tiền", { exact: true }).fill("Hoàn thử 1.000 đồng");
  await page.getByRole("button", { name: "Ghi nhận hoàn tiền giả lập", exact: true }).click();
  await expect(page.getByText("Đã ghi nhận hoàn tiền giả lập.", { exact: true })).toBeVisible();
  await page.getByRole("button", { name: "Khách hàng", exact: true }).click();
  await page.getByRole("button", { name: "Thêm khách hàng", exact: true }).click();
  await page.getByLabel("Họ và tên", { exact: true }).fill("Khách CRM Kiểm Thử");
  await page.getByLabel("Email", { exact: true }).fill("browser-demo@example.com");
  await page.getByLabel("Ghi chú nội bộ", { exact: true }).fill("Hồ sơ kiểm thử CI");
  await page.getByRole("button", { name: "Lưu khách hàng", exact: true }).click();
  await expect(page.getByText("Chưa có đơn đã xác minh.")).toBeVisible();
  await page.getByLabel("Mã đơn cần xác minh").fill(code);
  await page.getByRole("button", { name: "Kiểm tra mã đơn" }).click();
  await expect(page.getByText("Khách Kiểm Thử · browser-demo@example.com")).toBeVisible();
  await page.getByRole("checkbox", { name: /Tôi đã xác minh chủ đơn/ }).check();
  await page.getByRole("button", { name: "Gắn đơn đã xác minh" }).click();
  await expect(page.getByRole("listitem").filter({ hasText: code })).toContainText("Đã hủy");
});
test("admin tour editor persists edits and verifies poster", async ({
  page,
}) => {
  await login(page);
  await expect(page.getByRole("option", { name: /Bài viết|Danh mục bài viết|Thẻ bài viết/ })).toHaveCount(0);
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
    "/destinations",
    "/faq",
    "/gallery",
    "/policies/privacy",
  ]) {
    const r = await page.goto("http://localhost:3000" + path);
    expect(r?.status(), path).toBe(200);
    await expect(page.locator("main h1")).toBeVisible();
  }
  for (const path of ["/blog", "/blog/chuan-bi-di-ninh-chu", "/blog?category=cam-nang"]) {
    const response = await page.goto("http://localhost:3000" + path);
    expect(response?.status(), path).toBe(404);
  }
  const sitemap = await page.request.get("http://localhost:3000/sitemap.xml");
  expect(sitemap.status()).toBe(200);
  expect(await sitemap.text()).not.toContain("/blog");
  const r = await page.goto("http://localhost:3000/tours/not-existing");
  expect(r?.status()).toBe(404);
});

test("home search and media filters", async ({ page }) => {
  await page.goto("http://localhost:3000/");
  await page.getByLabel("Từ khóa", { exact: true }).fill("Ninh Chữ");
  await page.locator(".hero").getByRole("button", { name: "Tìm tour" }).click();
  await expect(page).toHaveURL(/q=/);
  await expect(page.locator(".tour-card").first()).toBeVisible();
  await page.getByText("Bộ lọc nâng cao", { exact: true }).click();
  await expect(
    page.getByRole("combobox", { name: "Danh mục tour", exact: true }),
  ).toBeVisible();
  await page.goto("http://localhost:3000/gallery?kind=video");
  await expect(
    page.getByText("Chưa có media phù hợp bộ lọc.", { exact: true }),
  ).toBeVisible();
  await page.goto("http://localhost:3000/gallery?kind=image");
  await expect(page.locator(".gallery-grid img").first()).toBeVisible();
});

test("counter booking and report Excel/CSV exports", async ({ page }) => {
  await login(page);
  await page.getByRole("button", { name: "Đơn tour", exact: true }).click();
  await page.getByRole("button", { name: "Tạo đơn thủ công", exact: true }).click();
  await page.getByLabel("Tour", { exact: true }).selectOption("ninh-chu-2-ngay-1-dem");
  await expect(page.getByRole("button", { name: "Lấy báo giá", exact: true })).toBeEnabled();
  await page.getByRole("button", { name: "Lấy báo giá", exact: true }).click();
  await page.getByLabel("Họ tên liên hệ", { exact: true }).fill("Khách tại quầy demo");
  await page.getByLabel("Email liên hệ", { exact: true }).fill("quay-m7@blue-sky.example");
  await page.getByLabel("Số điện thoại liên hệ", { exact: true }).fill("0900000000");
  await page.getByLabel("Họ tên hành khách 1", { exact: true }).fill("Khách tại quầy demo");
  await page.getByLabel("Ngày sinh hành khách 1", { exact: true }).fill("1990-01-01");
  await page.getByLabel("Nguồn đơn").selectOption("counter");
  await page.getByRole("checkbox", { name: /Khách đã đồng ý/ }).check();
  await page.getByRole("button", { name: "Tạo đơn thủ công và giữ chỗ" }).click();
  await expect(page.getByRole("heading", { name: /^Đơn BS/ })).toBeVisible();
  await page.getByRole("button", { name: "Tạo mã thanh toán · Toàn bộ" }).click();
  await page.getByLabel("Lý do xác nhận").fill("Kiểm thử đặt tour tại quầy");
  await page.getByRole("button", { name: "Xác nhận thanh toán giả lập", exact: true }).click();
  await expect(page.getByText("Sổ thu và hoàn khớp với đơn.")).toBeVisible();
  await expect(page.getByText("Còn nợ 0 đ", { exact: false })).toBeVisible();
  await page.getByRole("button", { name: "Báo cáo", exact: true }).click();
  await page.getByLabel("Từ ngày", { exact: true }).fill("2000-01-01");
  await page.getByLabel("Đến ngày", { exact: true }).fill("2000-01-30");
  await page.getByRole("button", { name: "Xem báo cáo", exact: true }).click();
  await expect(page.getByText("Chưa có giao dịch trong kỳ.")).toBeVisible();
  const xlsxPromise = page.waitForEvent("download");
  await page.getByRole("button", { name: "Tải báo cáo doanh thu", exact: true }).click();
  const xlsx = await xlsxPromise;
  expect(xlsx.suggestedFilename()).toBe("bao-cao-doanh-thu.xlsx");
  expect(readFileSync((await xlsx.path())!).subarray(0, 2).toString()).toBe("PK");
  await page.getByLabel("Định dạng báo cáo doanh thu", { exact: true }).selectOption("csv");
  const csvPromise = page.waitForEvent("download");
  await page.getByRole("button", { name: "Tải báo cáo doanh thu", exact: true }).click();
  const csv = await csvPromise;
  expect(readFileSync((await csv.path())!, "utf8")).toContain("Thu ròng (VND)");
});
