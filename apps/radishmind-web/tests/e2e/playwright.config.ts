import { defineConfig } from "@playwright/test";

const baseURL = process.env.RADISHMIND_E2E_WEB_URL;
const output = process.env.RADISHMIND_E2E_OUTPUT_DIR;
if (baseURL !== "http://127.0.0.1:4100" || !output) {
  throw new Error("Run browser regressions with npm run test:e2e so the isolated services are owned and cleaned up.");
}

export default defineConfig({
  testDir: ".",
  testMatch: "*.spec.ts",
  fullyParallel: false,
  workers: 1,
  forbidOnly: Boolean(process.env.CI),
  retries: 0,
  timeout: 90_000,
  expect: { timeout: 10_000 },
  reporter: [["list"], ["html", { outputFolder: `${output}/report`, open: "never" }]],
  outputDir: `${output}/results`,
  use: {
    baseURL,
    browserName: "chromium",
    viewport: { width: 1440, height: 900 },
    actionTimeout: 10_000,
    navigationTimeout: 15_000,
    screenshot: "only-on-failure",
    trace: "retain-on-failure",
    serviceWorkers: "block",
  },
});
