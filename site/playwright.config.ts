import { defineConfig, devices } from "@playwright/test";

const useSystemChrome =
  !process.env.CI && process.platform === "darwin";

export default defineConfig({
  testDir: "./e2e",
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  reporter: "list",
  use: {
    baseURL: "http://127.0.0.1:4173",
    trace: "on-first-retry",
  },
  projects: [
    {
      name: "chromium",
      use: {
        ...devices["Desktop Chrome"],
        ...(useSystemChrome ? { channel: "chrome" } : {}),
      },
    },
  ],
  webServer: {
    command: "npm run build && npm run serve:e2e",
    url: "http://127.0.0.1:4173/fgm/",
    reuseExistingServer: !process.env.CI,
    timeout: 90_000,
  },
});
