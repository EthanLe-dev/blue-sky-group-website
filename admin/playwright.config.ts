import { defineConfig } from "@playwright/test";

const engine = process.env.RELEASE_BROWSER;
if (engine && engine !== "firefox" && engine !== "webkit") {
  throw new Error("RELEASE_BROWSER must be firefox or webkit");
}
export default defineConfig({
  testDir: "e2e",
  ...(engine ? { testMatch: "release-ui.spec.ts" } : {}),
  workers: 1,
  use: {
    baseURL: "http://localhost:5173",
    ...(engine === "firefox" || engine === "webkit" ? { browserName: engine } : { channel: "chromium" }),
    trace: "off",
    screenshot: "only-on-failure",
  },
  reporter: "list",
});
