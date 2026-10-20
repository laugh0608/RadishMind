import type { Locator, Page, TestInfo } from "@playwright/test";
import { expect } from "./workflow-fixtures";
import { evaluationLanguage } from "./evaluation-fixtures";
export { evaluationLanguage as evidenceLanguage };
export type EvidenceLanguage = ReturnType<typeof evaluationLanguage>;
export const library = (page: Page) => page.locator("#application-result-artifact-library");
export const fixtureApplication = { id: "app_rrrrrrrrrrrrrrrr", name: "Result Library Fixture", kind: "workflow_copilot" };
export const resultId = "appres_rrrrrrrrrrrrrrrr";
export const resultSession = "appsess_rrrrrrrrrrrrrrrr";
export const resultContent = "# Stable workflow result\n\nSQLite exact-read and export evidence.";
export const resultReadPath = `/v1/user-workspace/application-sessions/${resultSession}/result-artifacts/${resultId}`;
export const resultExportPath = `/v1/user-workspace/applications/${fixtureApplication.id}/result-artifacts/${resultId}/export`;

export async function openLibrary(page: Page, applicationId = fixtureApplication.id) {
  await page.locator('a[href="#workspace-applications"]:visible').first().click();
  await page.locator(".application-catalog-panel").getByRole("button").filter({ hasText: applicationId }).click();
  await page.evaluate(() => { location.hash = "application-result-artifact-library"; });
  await expect(library(page)).toBeVisible();
  await expect(library(page).locator(".application-result-artifact-library-filters button").first()).toBeEnabled();
  await page.waitForLoadState("networkidle");
}
export async function readFixtureResult(page: Page) {
  await library(page).locator(".application-result-artifact-list").getByRole("button").filter({ hasText: resultId }).click();
  await expect(library(page).locator("pre")).toHaveText(resultContent);
}
export async function evidenceLayout(page: Page, panel: Locator, info: TestInfo, name: string) {
  for (const width of [1440, 720, 390]) {
    await page.setViewportSize({ width, height: 900 });
    await panel.scrollIntoViewIfNeeded();
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= document.documentElement.clientWidth + 1)).toBeTruthy();
    const clipped = await panel.locator("article, h3, h4, h5, p, dd, label, button, .status-badge, .workflow-rag-evaluation-workspace, .workflow-rag-evaluation-fields, .workflow-rag-thresholds")
      .evaluateAll(elements => elements.filter(element => element.clientWidth > 0 && element.scrollWidth > element.clientWidth + 1)
        .map(element => ({ tag: element.tagName, class: element.className, text: element.textContent?.slice(0, 65) })));
    expect(clipped, `${name} ${width} must fit its container`).toEqual([]);
    if (width !== 720) await panel.screenshot({ style: ".product-nav { visibility: hidden !important; }", path: info.outputPath(`${name}-${width}.png`) });
  }
  await page.setViewportSize({ width: 1440, height: 900 });
}

export async function holdEvidenceResponse(page: Page, path: string, method = "GET") {
  let release!: () => void, arrived!: () => void;
  let canceled = false;
  let complete!: () => void;
  const finished = new Promise<void>(resolve => { complete = resolve; });
  const gate = new Promise<void>(resolve => { release = resolve; });
  const ready = new Promise<void>(resolve => { arrived = resolve; });
  const pattern = new RegExp(`^http://127\\.0\\.0\\.1:17000${path.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")}(?:\\?.*)?$`);
  const handler = async (route: import("@playwright/test").Route) => {
    if (route.request().method() !== method) return route.fallback();
    const response = await route.fetch(); arrived(); await gate;
    if (!canceled) await route.fulfill({ response });
    complete();
  };
  await page.route(pattern, handler);
  return {
    arrived: ready,
    async deliver() { release(); await finished; await page.unroute(pattern, handler); },
    async dispose() { canceled = true; release(); if (!page.isClosed()) await page.unroute(pattern, handler); },
  };
}
