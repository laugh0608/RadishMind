import type { Page, Route } from "@playwright/test";
import { test, expect, createDraft, saveDraft, isEndpoint } from "./workflow-fixtures";
import { uiText, setTestLanguage } from "./ui-language";

type Locale = "en-US" | "zh-CN";
const owner = (page: Page) => page.locator(".workflow-review-owner-surface");
const comparison = (page: Page) => page.locator(".workflow-run-comparison");
async function switchLanguage(page: Page, locale: Locale) {
  await page.evaluate(value => {
    localStorage.setItem("radishmind.uiLocale.v1", value);
    window.dispatchEvent(new StorageEvent("storage", { key: "radishmind.uiLocale.v1", newValue: value, storageArea: localStorage }));
  }, locale);
  setTestLanguage(page, locale);
  await expect(page.locator("html")).toHaveAttribute("lang", locale);
}
async function prepareRuns(page: Page) {
  await createDraft(page); await saveDraft(page, 1);
  await page.getByRole("link", { name: uiText(page, "Open Workflow definition owner"), exact: true }).click();
  await page.getByRole("button", { name: uiText(page, "Create promotion candidate"), exact: true }).click();
  await page.getByRole("button", { name: uiText(page, "Append review v1"), exact: true }).click();
  await page.getByRole("button", { name: uiText(page, "Activate · expected pointer v0"), exact: true }).click();
  const ids: string[] = [];
  for (let i = 0; i < 2; i++) {
    await page.getByRole("textbox", { name: uiText(page, "One-time input"), exact: true }).fill(`Synthetic run review ${i}`);
    const response = page.waitForResponse(r => isEndpoint(r, "/v1/user-workspace/workflow-definition-runs", "POST"));
    await page.getByRole("button", { name: uiText(page, "Start exact version run"), exact: true }).click();
    const received = await response;
    expect(received.ok()).toBeTruthy();
    const body = await received.json();
    await expect(page.locator(".workflow-definition-run-result strong")).toHaveText(body.run.run_id);
    await expect(page.getByRole("textbox", { name: uiText(page, "One-time input"), exact: true })).toHaveValue("");
    ids.push(body.run.run_id);
  }
  expect(ids[0]).not.toBe(ids[1]);
  await page.getByRole("button", { name: uiText(page, "Open Run History"), exact: true }).click();
  await expect(owner(page).locator(".workflow-review-run-row")).toHaveCount(2);
  await expect(owner(page).locator(".workflow-review-run-detail")).toContainText(ids[1]);
  return ids;
}
async function choosePair(page: Page, ids: string[]) {
  await page.locator('a.workflow-review-task[href="#workflow-run-comparison"]').click();
  const selects = owner(page).locator(".workflow-review-pair-rail select");
  await selects.nth(0).selectOption(ids[0]); await selects.nth(1).selectOption(ids[1]);
}

test("run review preserves exact references, filters and pending comparison across languages", async ({ page, application }, testInfo) => {
  expect(application.kind).toBe("workflow_copilot");
  const locale: Locale = testInfo.project.use.locale === "zh-CN" ? "zh-CN" : "en-US";
  const next: Locale = locale === "zh-CN" ? "en-US" : "zh-CN";
  const ids = await prepareRuns(page);
  await owner(page).locator(".workflow-review-disclosure").first().locator("summary").click();
  const filter = owner(page).locator('input[placeholder="workflow_run_…"]');
  await filter.fill("unsaved_exact_code");
  const requests: string[] = [];
  page.on("request", request => { if (new URL(request.url()).port === "17000") requests.push(request.method() + " " + new URL(request.url()).pathname); });
  await switchLanguage(page, next);
  await expect(filter).toHaveValue("unsaved_exact_code");
  await expect(owner(page).locator(".workflow-review-run-detail")).toContainText(ids[1]);
  await expect(owner(page).locator(".workflow-review-run-status").first()).toHaveClass(/status-succeeded/);
  expect(requests).toEqual([]);
  for (const width of [1440, 720, 390]) {
    await page.setViewportSize({ width, height: 900 });
    expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBe(width);
    expect(await owner(page).locator(".workflow-review-run-detail").evaluate(element => [...element.querySelectorAll("dd, p, h5")].every(node => node.scrollWidth <= node.clientWidth + 1))).toBe(true);
    if (width !== 720) await owner(page).locator(".workflow-review-run-detail").screenshot({ style: ".product-nav { visibility: hidden !important; }", path: testInfo.outputPath(`detail-${next}-${width}.png`) });
  }
  await page.setViewportSize({ width: 1440, height: 900 });
  await filter.fill("");
  await choosePair(page, ids);

  let release!: () => void; let arrived!: () => void;
  const gate = new Promise<void>(resolve => { release = resolve; });
  const waiting = new Promise<void>(resolve => { arrived = resolve; });
  const pattern = /\/workflow-runs\/[^/]+\/comparison\?/;
  const hold = async (route: Route) => { const response = await route.fetch(); arrived(); await gate; await route.fulfill({ response }); };
  await page.route(pattern, hold);
  try {
    await owner(page).locator(".workflow-review-pair-rail button").click();
    await waiting;
    const before = [...requests];
    await switchLanguage(page, locale);
    await expect(comparison(page)).toContainText(locale === "en-US" ? "Comparing durable run records" : "正在对比持久化运行记录");
    expect(requests).toEqual(before);
    release();
    await expect(comparison(page).locator("h4")).toBeVisible();
  } finally { release(); await page.unroute(pattern, hold); }
  const before = [...requests];
  for (const language of [next, locale]) {
    await switchLanguage(page, language);
    await expect(comparison(page)).toHaveAttribute("aria-label", language === "en-US" ? "Workflow run regression comparison" : "工作流运行回归对比");
    await expect(owner(page).locator(".workflow-review-pair-rail select").nth(0)).toHaveValue(ids[0]);
    await expect(owner(page).locator(".workflow-review-pair-rail select").nth(1)).toHaveValue(ids[1]);
    await expect(comparison(page)).toContainText(ids[0]);
    await expect(comparison(page)).toContainText(ids[1]);
    await expect(comparison(page).locator(".workflow-run-comparison-findings code").first()).toHaveCSS("text-transform", "none");
    for (const width of [1440, 720, 390]) {
      await page.setViewportSize({ width, height: 900 });
      expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBe(width);
      const overflow = await comparison(page).evaluate(element => [...element.querySelectorAll("dd, .status-badge, .boundary-note, .workflow-run-history-node-row, .workflow-run-history-node-row span, .workflow-run-history-node-row strong, .workflow-run-history-node-row small, .workflow-run-history-node-row code")].filter(node => node.scrollWidth > node.clientWidth + 1).map(node => node.className));
      expect(overflow).toEqual([]);
      if (width !== 720) await comparison(page).screenshot({ style: ".product-nav { visibility: hidden !important; }", path: testInfo.outputPath(`comparison-${language}-${width}.png`) });
    }
  }
  expect(requests).toEqual(before);
  expect(requests.filter(value => value.startsWith("POST "))).toEqual([]);
});

test("run review translates failures, rejects raw error prose and recovers explicitly", async ({ page, application }, testInfo) => {
  expect(application.kind).toBe("workflow_copilot");
  const locale: Locale = testInfo.project.use.locale === "zh-CN" ? "zh-CN" : "en-US";
  const next: Locale = locale === "zh-CN" ? "en-US" : "zh-CN";
  const ids = await prepareRuns(page);
  const detailPattern = new RegExp(`/workflow-runs/${ids[0]}\\?`);
  await page.route(detailPattern, route => route.fulfill({ status: 503, body: "untrusted raw detail prose" }));
  await owner(page).locator(".workflow-review-run-row").filter({ hasText: ids[0] }).click();
  await expect(owner(page).getByRole("alert")).toContainText(locale === "en-US" ? "Unable to read this run" : "无法读取该运行");
  await switchLanguage(page, next);
  await expect(owner(page).getByRole("alert")).toContainText(next === "en-US" ? "Unable to read this run" : "无法读取该运行");
  await expect(owner(page)).not.toContainText("untrusted raw detail prose");
  await page.unroute(detailPattern);
  await owner(page).locator(".workflow-review-run-row").filter({ hasText: ids[0] }).click();
  await expect(owner(page).locator(".workflow-review-run-detail")).toContainText(ids[0]);
  let release!: () => void; let arrived!: () => void;
  const gate = new Promise<void>(resolve => { release = resolve; });
  const waiting = new Promise<void>(resolve => { arrived = resolve; });
  const hold = async (route: Route) => { const response = await route.fetch(); arrived(); await gate; await route.fulfill({ response }); };
  await page.route(detailPattern, hold);
  try {
    await owner(page).locator(".workflow-review-run-row").filter({ hasText: ids[0] }).click();
    await waiting;
    await expect(owner(page)).toContainText(next === "en-US" ? "Reading the selected run" : "正在读取所选运行");
    await owner(page).locator(".workflow-review-run-row").filter({ hasText: ids[1] }).click();
    await expect(owner(page).locator(".workflow-review-run-detail")).toContainText(ids[1]);
    const lateResponse = page.waitForResponse(response => new URL(response.url()).pathname.endsWith(ids[0]));
    release(); await (await lateResponse).finished();
    await page.evaluate(() => new Promise<void>(resolve => requestAnimationFrame(() => requestAnimationFrame(() => resolve()))));
    await expect(owner(page).locator(".workflow-review-run-detail")).toContainText(ids[1]);
  } finally { release(); await page.unroute(detailPattern, hold); }
  await choosePair(page, ids);
  const pattern = /\/workflow-runs\/[^/]+\/comparison\?/;
  let count = 0;
  await page.route(pattern, async route => {
    count++;
    const response = await route.fetch(); const body = await response.json();
    await route.fulfill({ status: 409, json: { ...body, comparison: null, failure_code: "future_failure_code", failure_summary: "untrusted raw comparison prose" } });
  });
  await owner(page).locator(".workflow-review-pair-rail button").click();
  await expect(comparison(page).getByRole("alert")).toContainText("future_failure_code");
  const requestsBeforeSwitch = count;
  await switchLanguage(page, locale);
  await expect(comparison(page)).toContainText(locale === "en-US" ? "Unable to read the comparison" : "无法读取比较结果");
  await expect(comparison(page)).not.toContainText("untrusted raw comparison prose");
  expect(count).toBe(requestsBeforeSwitch);
  await page.unroute(pattern);
  await comparison(page).getByRole("button", { name: locale === "en-US" ? "Retry comparison" : "重试对比", exact: true }).click();
  await expect(comparison(page).locator("h4")).toBeVisible();
  await expect(comparison(page).getByRole("alert")).toHaveCount(0);
});
