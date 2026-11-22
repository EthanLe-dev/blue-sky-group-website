import { test, expect } from "@playwright/test";
import { readFileSync } from "node:fs";

test("sales forms preserve input, verify captcha and show the configured receipt", async ({ page }) => {
  await page.goto("http://localhost:3000/schools");
  const form = page.locator("#bao-gia");
  await form.getByLabel("Tên trường").fill("Trường Kiểm Thử");
  await form.getByLabel("Họ và tên").fill("Khách Kiểm Thử");
  await form.getByLabel("Email").fill("lead-browser@example.com");
  await form.getByLabel("Số điện thoại").fill("0901234567");
  await form.getByLabel("Số người").fill("35");
  await form.getByLabel("Điểm đến").fill("Ninh Chữ");
  await form.getByRole("checkbox").check();

  let posts = 0;
  await page.route("**/api/v1/group-requests", async (route) => {
    posts++;
    if (posts === 1) {
      await route.continue({
        postData: JSON.stringify({ ...route.request().postDataJSON(), turnstile_token: "invalid" }),
      });
      return;
    }
    await route.continue();
  });
  await form.getByRole("button", { name: "Gửi yêu cầu" }).click();
  await expect(form.getByRole("alert")).toContainText("xác minh");
  await expect(form.getByLabel("Tên trường")).toHaveValue("Trường Kiểm Thử");
  await expect(form.getByRole("checkbox")).toBeChecked();
  await expect(form.getByRole("button", { name: "Gửi yêu cầu" })).toBeEnabled();
  await form.locator("form").evaluate((element: HTMLFormElement) => {
    element.requestSubmit();
    element.requestSubmit();
  });
  await expect(form.getByRole("status")).toContainText(/GR-[A-F0-9]{20}/);
  await expect(form.getByRole("status")).toContainText("một ngày làm việc");
  expect(posts).toBe(2);

  await page.goto("http://localhost:3000/contact#goi-lai");
  const callback = page.locator("#goi-lai");
  await callback.getByLabel("Số điện thoại").fill("0901234567");
  await callback.getByLabel("Khung giờ thuận tiện").selectOption("Buổi chiều");
  await callback.getByRole("checkbox").check();
  await callback.getByRole("button", { name: "Gửi yêu cầu" }).click();
  await expect(callback.getByRole("status")).toContainText(/CB-[A-F0-9]{20}/);
});

test("sales staff assigns a lead, issues a quote and serves a private attachment", async ({ page }) => {
  const name = "Đoàn kiểm thử " + Date.now();
  const publicBase = "http://localhost:3000";
  const guest = await page.request.post(publicBase + "/api/v1/guest-sessions", {
    headers: { Origin: publicBase }, data: {},
  });
  expect(guest.ok()).toBeTruthy();
  const { data: session } = await guest.json();
  const lead = await page.request.post(publicBase + "/api/v1/group-requests", {
    headers: { Origin: publicBase, "X-CSRF-Token": session.csrf_token, "Idempotency-Key": crypto.randomUUID() },
    data: {
      group_type: "company", organization_name: name, contact_name: "Khách Kiểm Thử",
      contact_email: "sales-flow@example.com", contact_phone: "0901234567",
      participant_count: 40, destination_text: "Ninh Chữ",
      consent: { accepted: true, policy_version: "sales-v1" }, turnstile_token: "local-test-token",
    },
  });
  expect(lead.status()).toBe(201);
  const { data: receipt } = await lead.json();

  const env = Object.fromEntries(readFileSync("../.env", "utf8").split(/\r?\n/)
    .filter((line) => line && !line.startsWith("#") && line.includes("="))
    .map((line) => [line.slice(0, line.indexOf("=")), line.slice(line.indexOf("=") + 1)]));
  await page.goto("/");
  await page.getByLabel("Email", { exact: true }).fill(env.BOOTSTRAP_ADMIN_EMAIL);
  await page.getByLabel("Mật khẩu", { exact: true }).fill(env.BOOTSTRAP_ADMIN_PASSWORD);
  await page.getByRole("button", { name: "Đăng nhập", exact: true }).click();
  await page.getByRole("button", { name: "Yêu cầu tư vấn" }).click();
  await page.getByRole("button", { name: new RegExp(receipt.code) }).click();

  const detail = page.getByRole("article", { name: "Yêu cầu " + receipt.code });
  await detail.getByLabel("Giao cho nhân viên").selectOption({ index: 1 });
  await detail.getByLabel("Chuyển trạng thái").selectOption("preparing");
  await detail.getByRole("button", { name: "Lưu phân công" }).click();
  await expect(detail.getByRole("heading", { name: /Đang chuẩn bị/ })).toBeVisible();
  await detail.getByLabel("Ghi chú mới").fill("Đã trao đổi với đại diện đoàn.");
  await detail.getByRole("button", { name: "Thêm ghi chú" }).click();
  await expect(detail.getByText("Đã trao đổi với đại diện đoàn.")).toBeVisible();

  await detail.getByRole("button", { name: "Tạo báo giá", exact: true }).click();
  await detail.getByLabel("Tiêu đề").fill("Báo giá mẫu");
  await detail.getByLabel("Lịch trình").fill("Ninh Chữ hai ngày");
  await detail.getByLabel("Tổng tiền (VND)").fill("12000000");
  await detail.getByLabel("Hiệu lực đến ngày").fill(new Date(Date.now() + 864000000).toISOString().slice(0, 10));
  await detail.getByRole("button", { name: "Lưu bản nháp" }).click();
  await expect(detail.getByRole("heading", { name: /Phiên bản 1.*Bản nháp/ })).toBeVisible();
  await detail.getByRole("button", { name: "Phát hành", exact: true }).click();
  await expect(detail.getByRole("heading", { name: /Phiên bản 1.*Đã phát hành/ })).toBeVisible();
  const downloadPromise = page.waitForEvent("download");
  await detail.getByRole("button", { name: "Tải PDF báo giá", exact: true }).click();
  expect((await downloadPromise).suggestedFilename()).toBe("bao-gia-v1.pdf");
  await expect(detail.getByRole("button", { name: "Sửa bản nháp" })).toHaveCount(0);
  await detail.getByLabel("Lý do hoặc cách xác minh").fill("Đại diện đoàn xác nhận qua điện thoại.");
  await detail.getByRole("button", { name: "Ghi nhận kết quả" }).click();
  await expect(detail.getByRole("heading", { name: /Đã chốt/ })).toBeVisible();

  await detail.getByLabel("Tệp báo giá").setInputFiles({
    name: "bao-gia-mau.pdf", mimeType: "application/pdf", buffer: Buffer.from("%PDF-1.4\n% Demo CI\n"),
  });
  await detail.getByRole("button", { name: "Đính kèm" }).click();
  await expect(detail.getByText("bao-gia-mau.pdf")).toBeVisible();
  await detail.getByRole("button", { name: "Tạo liên kết tải" }).click();
  const href = await detail.getByRole("link", { name: "Tải xuống" }).getAttribute("href");
  expect(href).toContain("signature=");
  const signed = new URL(href!, "http://localhost:5173");
  expect((await page.request.get(signed.origin + signed.pathname)).status()).toBe(404);
  expect((await page.request.get(signed.href)).status()).toBe(200);
});
