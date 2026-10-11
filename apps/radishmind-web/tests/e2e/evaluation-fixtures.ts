import { randomUUID } from "node:crypto";
import type { Locator, Page, TestInfo } from "@playwright/test";
import { test, expect, isEndpoint } from "./workflow-fixtures";
import { switchLanguage } from "./run-review-fixtures";
export { test, expect };
export type Locale = "en-US" | "zh-CN";
export const evaluation = (page: Page) => page.locator(".application-evaluation-workspace");
export const planOwner = (page: Page) => page.locator(".application-evaluation-plan-owner");
export const button = (panel: Locator, name: string) => panel.getByRole("button", { name, exact: true });
export function evaluationLanguage(page: Page, info: TestInfo) {
  let locale: Locale = info.project.use.locale === "zh-CN" ? "zh-CN" : "en-US";
  const requests: string[] = [];
  page.on("request", request => {
    const url = new URL(request.url());
    if (url.port === "17000") requests.push(`${request.method()} ${url.pathname}`);
  });
  return {
    copy: (en: string, zh: string) => locale === "en-US" ? en : zh,
    requests,
    async switch() {
      const before = [...requests];
      locale = locale === "en-US" ? "zh-CN" : "en-US";
      await switchLanguage(page, locale);
      await page.evaluate(() => new Promise<void>(resolve => requestAnimationFrame(() => requestAnimationFrame(() => resolve()))));
      expect(requests).toEqual(before);
    },
  };
}
export async function openEvaluation(page: Page, task: "plan" | "schedule" | "occurrence" | "campaign" | "handoff") {
  // Product navigation chooses the owning application surface; hash only chooses its task.
  const link = page.locator('a[href="#workspace-run-history"]:visible').first();
  await link.click();
  await expect(evaluation(page)).toBeVisible();
  await page.evaluate(task => { window.location.hash = `application-evaluation-${task}`; }, task);
  await expect(evaluation(page)).toHaveAttribute("data-active-task", task);
}
export async function evaluationSubmit(page: Page, control: Locator, path: string) {
  const pending = page.waitForResponse(response => isEndpoint(response, path, "POST"));
  await control.click();
  const response = await pending;
  const body = await response.json();
  return { body, request: response.request() };
}
export async function createPlan(page: Page, applicationId: string, ui: ReturnType<typeof evaluationLanguage>, caseId = randomUUID()) {
  await openEvaluation(page, "plan");
  const panel = planOwner(page), copy = ui.copy;
  await panel.getByRole("textbox", { name: copy("Plan name", "计划名称"), exact: true }).fill("Synthetic Plan 原始名称");
  await panel.getByRole("textbox", { name: copy("Ordered fixtures · 1–20 · strict JSON", "有序夹具 · 1–20 项 · 严格 JSON"), exact: true }).fill(JSON.stringify([{
    item_key: "fixture_original", name: "Fixture 原文", expected_classification: "unchanged", workflow_definition: null, application_rag: null,
    prompt_application: { variables: { question: `E2E_CASE:${caseId}:valid`, tone: "E2E_PROMPT_SYSTEM_V1" } }, agent_copilot: null,
  }], null, 2));
  await button(panel, copy("Review create", "审查创建")).click();
  const result = await evaluationSubmit(page, button(panel, copy("Confirm Create", "确认创建")), `/v1/user-workspace/applications/${applicationId}/evaluation-plans`);
  expect(result.body.failure_code).toBeNull();
  await expect(evaluation(page).locator(".application-evaluation-operation")).toContainText(result.body.plan.plan_id);
  return result;
}
export async function evaluationLayout(page: Page, panel: Locator, info: TestInfo, name: string) {
  for (const width of [1440, 720, 390]) {
    await page.setViewportSize({ width, height: 900 });
    expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBe(width);
    const overflow = await panel.evaluate(element => [...element.querySelectorAll("strong, small, dd, label, .status-badge, .application-evaluation-confirmation, .application-evaluation-item, .application-evaluation-schedule-main, .application-evaluation-schedule-editor, .application-evaluation-schedule-metrics, .application-evaluation-campaign-surface, .application-evaluation-workbench")]
      .filter(node => node.clientWidth > 0 && node.scrollWidth > node.clientWidth + 1)
      .map(node => ({ tag: node.tagName, class: node.className, text: node.textContent?.slice(0, 90) })));
    expect(overflow).toEqual([]);
    if (width !== 720) await panel.screenshot({ style: ".product-nav { visibility: hidden !important; }", path: info.outputPath(`${name}-${width}.png`) });
  }
  await page.setViewportSize({ width: 1440, height: 900 });
}
