import { test, expect } from "@playwright/test";

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
