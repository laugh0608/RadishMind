import { defineConfig } from "@playwright/test";

const baseURL = process.env.RADISHMIND_E2E_WEB_URL;
const output = process.env.RADISHMIND_E2E_OUTPUT_DIR;
const suite = process.env.RADISHMIND_E2E_SUITE;
if (baseURL !== "http://127.0.0.1:4100" || !output || !["templates", "workflow", "rag", "rag-promotion", "rag-application", "http-tool", "offline-projection", "identity"].includes(suite ?? "")) {
  throw new Error("Run browser regressions with npm run test:e2e so the isolated services are owned and cleaned up.");
}

export default defineConfig({
  testDir: ".",
  testMatch: suite === "identity" ? ["identity.spec.ts", "identity-members.spec.ts"] : suite === "offline-projection" ? "workflow-projection.spec.ts" : suite === "http-tool" ? "workflow-http-tool.spec.ts" : suite === "rag-application" ? "workflow-rag-application.spec.ts" : suite === "rag-promotion" ? "workflow-rag-promotion.spec.ts" : suite === "rag" ? ["workflow-rag.spec.ts", "workflow-rag-execution.spec.ts"] : suite === "templates" ? "workflow-template.spec.ts" : ["prompt.spec.ts", "workflow.spec.ts"],
  projects: [
    { name: "chromium-en-US", use: { locale: "en-US" } },
    { name: "chromium-zh-CN", use: { locale: "zh-CN" } },
  ],
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
    screenshot: suite === "identity" ? "off" : "only-on-failure",
    // Identity responses can contain one-time invitation codes. Use only explicitly
    // masked screenshots; never retain automatic screenshots or request traces.
    trace: suite === "identity" ? "off" : "retain-on-failure",
    serviceWorkers: "block",
  },
});
