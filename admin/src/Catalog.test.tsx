// @vitest-environment jsdom
import { afterEach, expect, test, vi } from "vitest";
import { cleanup, fireEvent, render, screen, within } from "@testing-library/react";
vi.mock("./lib/session", () => ({ authorized: vi.fn() }));
import { authorized } from "./lib/session";
import Catalog from "./Catalog";

afterEach(() => { cleanup(); vi.resetAllMocks(); });
test("grouped tour editor keeps edited fields after a failed save", async () => {
  const id = "00000000-0000-0000-0000-000000000001";
  vi.mocked(authorized).mockImplementation(async (path, options) => {
    if (options?.method === "PATCH") throw new Error("Không lưu được, thử lại.");
    if (path === "/admin/tours/" + id) return { id, title: "Tour mẫu", summary: "Tóm tắt cũ", slug: "tour-mau" };
    if (path.startsWith("/admin/tours?")) return [{ id, title: "Tour mẫu" }];
    return [];
  });
  render(<Catalog permissions={["catalog.write"]} />);
  fireEvent.click(await screen.findByRole("button", { name: /Tour mẫu/ }));
  const overview = await screen.findByRole("group", { name: "Thông tin hành trình" });
  expect(within(overview).getByLabelText("Tên tour")).toBeTruthy();
  expect(within(screen.getByRole("group", { name: "Hình ảnh và phân loại" })).getByLabelText("ID ảnh bìa")).toBeTruthy();
  expect(within(screen.getByRole("group", { name: "Hiển thị tìm kiếm" })).getByLabelText("Tiêu đề SEO")).toBeTruthy();
  const summary = within(overview).getByLabelText("Tóm tắt") as HTMLTextAreaElement;
  fireEvent.change(summary, { target: { value: "Nội dung chưa lưu" } });
  fireEvent.click(screen.getByRole("button", { name: "Lưu nội dung" }));
  expect(await screen.findByRole("alert")).toBeTruthy();
  expect(summary.value).toBe("Nội dung chưa lưu");
  expect((screen.getByLabelText("Tên tour") as HTMLInputElement).value).toBe("Tour mẫu");
});
