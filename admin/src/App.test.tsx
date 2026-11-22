// @vitest-environment jsdom
import { afterEach, expect, test, vi } from "vitest";
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
vi.mock("./lib/session", () => ({
  restore: async () => null,
  login: async () => {
    throw new Error("Đăng nhập thất bại");
  },
  logout: async () => {},
}));
import App from "./App";
afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});
test("login route exposes labelled accessible fields", () => {
  render(<App />);
  expect(
    screen.getByRole("heading", { name: "Đăng nhập quản trị" }),
  ).toBeTruthy();
  expect(screen.getByLabelText("Email")).toBeTruthy();
  expect(screen.getByLabelText("Mật khẩu")).toBeTruthy();
});
test("failed login keeps form and presents error", async () => {
  vi.stubGlobal(
    "fetch",
    vi
      .fn()
      .mockResolvedValue({ ok: false, json: async () => ({ success: false }) }),
  );
  render(<App />);
  fireEvent.change(screen.getByLabelText("Email"), {
    target: { value: "staff@example.com" },
  });
  fireEvent.change(screen.getByLabelText("Mật khẩu"), {
    target: { value: "wrong-password" },
  });
  fireEvent.submit(
    screen.getByRole("button", { name: "Đăng nhập" }).closest("form")!,
  );
  expect(await screen.findByRole("alert")).toBeTruthy();
  expect((screen.getByLabelText("Email") as HTMLInputElement).value).toBe(
    "staff@example.com",
  );
});
