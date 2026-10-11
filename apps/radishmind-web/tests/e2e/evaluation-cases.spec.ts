import type { Page, Locator, TestInfo, Route } from "@playwright/test";
import { test, expect, isEndpoint, holdDraftResponse, selectApplication, finishResponseRender } from "./workflow-fixtures";
import { prepareRuns, switchLanguage } from "./run-review-fixtures";

type Locale = "en-US" | "zh-CN";
const text = (locale: Locale, en: string, zh: string) => locale === "en-US" ? en : zh;
const cases = (page: Page) => page.locator(".workflow-review-cases");
const suites = (page: Page) => page.locator(".workflow-review-release");
const casePath = "/v1/user-workspace/workflow-evaluation-cases";
const suitePath = "/v1/user-workspace/workflow-evaluation-suites";
const button = (panel: Locator, locale: Locale, en: string, zh: string) => panel.getByRole("button", { name: text(locale, en, zh), exact: true });

async function layout(page: Page, panel: Locator, testInfo: TestInfo, name: string) {
  for (const width of [1440, 720, 390]) {
    await page.setViewportSize({ width, height: 900 });
    expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBe(width);
    expect(await panel.evaluate(element => [...element.querySelectorAll("strong, small, dd, label, .status-badge, .workflow-evaluation-review-item")].filter(node => node.scrollWidth > node.clientWidth + 1).map(node => node.className))).toEqual([]);
    if (width !== 720) await panel.screenshot({ style: ".product-nav { visibility: hidden !important; }", path: testInfo.outputPath(`${name}-${width}.png`) });
  }
  await page.setViewportSize({ width: 1440, height: 900 });
}

async function createCase(page: Page, locale: Locale, ids: string[]) {
  await page.locator('a.workflow-review-task[href="#workflow-evaluation-cases"]').click();
  const panel = cases(page);
  await panel.getByRole("textbox", { name: text(locale, "Case name", "用例名称"), exact: true }).fill("Synthetic case 名称");
  await panel.getByRole("combobox", { name: text(locale, "Baseline", "基线"), exact: true }).selectOption(ids[0]);
  await panel.getByRole("combobox", { name: text(locale, "Candidate", "候选"), exact: true }).selectOption(ids[1]);
  await button(panel, locale, "Add candidate", "添加候选").click();
  const saved = page.waitForResponse(r => isEndpoint(r, casePath, "POST"));
  await button(panel, locale, "Create case", "创建用例").click();
  const response = await saved;
  const body = await response.json();
  expect(body.failure_code).toBeNull();
  await expect(panel.locator(".workflow-evaluation-case-list > article")).toHaveCount(1);
  return { record: body.case, headers: await response.request().allHeaders() };
}

async function inspectCase(page: Page, locale: Locale, caseId: string) {
  const read = page.waitForResponse(r => new URL(r.url()).pathname === `${casePath}/${caseId}/review`);
  await button(cases(page), locale, "Review case", "审查用例").click();
  const body = await (await read).json();
  expect(body.failure_code).toBeNull();
  await expect(cases(page).locator(".workflow-evaluation-review")).toBeVisible();
  return body.review;
}

test("evaluation cases preserve drafts, pending writes, conflict and exact historical reviews across languages", async ({ page, application }, testInfo) => {
  expect(application.kind).toBe("workflow_copilot");
  let locale: Locale = testInfo.project.use.locale === "zh-CN" ? "zh-CN" : "en-US";
  const other: Locale = locale === "en-US" ? "zh-CN" : "en-US";
  const ids = await prepareRuns(page);
  const { record, headers } = await createCase(page, locale, ids);
  const review = await inspectCase(page, locale, record.case_id);
  const panel = cases(page);
  const revision = panel.locator("details").nth(1);
  const revisionName = revision.getByRole("textbox");
  await revisionName.fill("Unsaved 修订");
  const requests: string[] = [];
  page.on("request", r => { if (new URL(r.url()).port === "17000") requests.push(r.method() + " " + new URL(r.url()).pathname); });
  const before = [...requests];
  await switchLanguage(page, other); locale = other;
  await expect(revisionName).toHaveValue("Unsaved 修订");
  await expect(revision.getByRole("combobox", { name: text(locale, "Baseline", "基线"), exact: true })).toHaveValue(ids[0]);
  expect(requests).toEqual(before);

  const revisionPath = `${casePath}/${record.case_id}/revisions`;
  const held = await holdDraftResponse(page, revisionPath, "POST");
  try {
    await button(panel, locale, "Create revision", "创建修订").click();
    await held.arrived;
    await expect(button(panel, locale, "Create revision", "创建修订")).toBeDisabled();
    const beforeSwitch = [...requests];
    locale = locale === "en-US" ? "zh-CN" : "en-US";
    await switchLanguage(page, locale);
    await expect(panel.getByRole("status").first()).toContainText(text(locale, "Saving evidence", "正在保存证据"));
    expect(requests).toEqual(beforeSwitch);
    await held.deliver();
  } finally { await held.dispose(); }
  await expect(revision.locator("summary")).toContainText(text(locale, "current v2", "当前 v2"));
  // Create a real concurrent version using the same authorized development identity.
  const external = await page.request.post(`http://127.0.0.1:17000${revisionPath}`, { headers, data: {
    workspace_id: record.workspace_id, application_id: application.id, expected_version: 2,
    revision_kind: "case_revision", name: "Concurrent v3", baseline_run_id: ids[0], expectations: record.expectations,
  } });
  expect((await external.json()).case.version).toBe(3);
  await revisionName.fill("Keep local conflict 编辑");
  await button(panel, locale, "Create revision", "创建修订").click();
  await expect(panel.getByRole("alert")).toContainText(text(locale, "Local edits are retained", "本地编辑已保留"));
  await expect(revisionName).toHaveValue("Keep local conflict 编辑");
  await expect(button(panel, locale, "Create revision", "创建修订")).toBeDisabled();
  const conflictRequests = [...requests];
  locale = locale === "en-US" ? "zh-CN" : "en-US"; await switchLanguage(page, locale);
  await expect(revisionName).toHaveValue("Keep local conflict 编辑");
  expect(requests).toEqual(conflictRequests);
  await button(panel, locale, "Load current version and replace local edits", "载入当前版本并替换本地编辑").click();
  await expect(revisionName).toHaveValue("Concurrent v3");
  await button(panel, locale, "Review v1", "审查 v1").click();
  await expect(panel.locator(".workflow-evaluation-review .eyebrow")).toContainText(text(locale, "version 1", "版本 1"));
  await expect(panel.locator(".workflow-evaluation-review")).toContainText(ids[1]);
  expect(review.case_id).toBe(record.case_id);
  // A late historical response must not replace the version selected afterwards.
  let release!: () => void; let arrived!: () => void;
  const gate = new Promise<void>(resolve => { release = resolve; });
  const waiting = new Promise<void>(resolve => { arrived = resolve; });
  const pattern = new RegExp(`/workflow-evaluation-cases/${record.case_id}/review\\?`);
  const hold = async (route: Route) => {
    if (new URL(route.request().url()).searchParams.get("version") !== "1") return route.fallback();
    const response = await route.fetch(); arrived(); await gate; await route.fulfill({ response });
  };
  await page.route(pattern, hold);
  try {
    await button(panel, locale, "Review v1", "审查 v1").click();
    await waiting;
    await button(panel, locale, "Review v3", "审查 v3").click();
    await expect(panel.locator(".workflow-evaluation-review .eyebrow")).toContainText(text(locale, "version 3", "版本 3"));
    const delivered = page.waitForResponse(r => new URL(r.url()).pathname === `${casePath}/${record.case_id}/review` && new URL(r.url()).searchParams.get("version") === "1");
    release(); await finishResponseRender(page, await delivered);
    await expect(panel.locator(".workflow-evaluation-review .eyebrow")).toContainText(text(locale, "version 3", "版本 3"));
  } finally { release(); await page.unroute(pattern, hold); }
  await layout(page, panel, testInfo, `case-${locale}`);
});

test("evaluation cases and suites retain exact refs, translate failures and record explicit digest-bound decisions", async ({ page, application }, testInfo) => {
  expect(application.kind).toBe("workflow_copilot");
  let locale: Locale = testInfo.project.use.locale === "zh-CN" ? "zh-CN" : "en-US";
  const ids = await prepareRuns(page);
  const { record } = await createCase(page, locale, ids);
  const review = await inspectCase(page, locale, record.case_id);
  // Set the expected classification from the actual durable review, without changing its run references.
  const actual = review.items[0].actual_classification;
  expect(["unchanged", "changed", "regression", "improvement"]).toContain(actual);
  const revision = cases(page).locator("details").nth(1);
  await revision.getByRole("textbox").fill("Suite exact case v2");
  await revision.locator(".workflow-evaluation-expectations select").selectOption(actual);
  await button(cases(page), locale, "Create revision", "创建修订").click();
  await expect(revision.locator("summary")).toContainText(text(locale, "current v2", "当前 v2"));
  await page.locator('a.workflow-review-task[href="#workflow-evaluation-release-review"]').click();
  const panel = suites(page);
  await panel.getByRole("textbox", { name: text(locale, "Suite name", "套件名称"), exact: true }).fill("Synthetic suite 名称");
  await panel.getByRole("combobox", { name: text(locale, "Evaluation case", "评测用例"), exact: true }).selectOption(record.case_id);
  await expect(panel.getByRole("spinbutton")).toHaveValue("2");
  await button(panel, locale, "Add exact version", "添加精确版本").click();
  const saved = page.waitForResponse(r => isEndpoint(r, suitePath, "POST"));
  await button(panel, locale, "Create suite", "创建套件").click();
  const savedResponse = await saved; const body = await savedResponse.json();
  expect(body.failure_code).toBeNull();
  expect(body.suite.case_refs).toEqual([{ case_id: record.case_id, version: 2 }]);
  const suiteId = body.suite.suite_id;
  const row = panel.locator(".workflow-run-history-live-row");
  await row.click();
  await expect(panel.locator(".workflow-evaluation-review h5")).toHaveText(text(locale, "Passed", "通过"));
  await panel.getByRole("combobox", { name: text(locale, "Human decision", "人工决定"), exact: true }).selectOption("approved");
  const decisionPath = `${suitePath}/${suiteId}/decisions`;
  const held = await holdDraftResponse(page, decisionPath, "POST");
  const posts: string[] = [];
  page.on("request", r => { if (r.method() === "POST") posts.push(new URL(r.url()).pathname); });
  try {
    await button(panel, locale, "Record decision v1", "记录决定 v1").click();
    await held.arrived;
    locale = locale === "en-US" ? "zh-CN" : "en-US"; await switchLanguage(page, locale);
    await expect(panel.getByRole("combobox", { name: text(locale, "Human decision", "人工决定"), exact: true })).toHaveValue("approved");
    await expect(button(panel, locale, "Record decision v1", "记录决定 v1")).toBeDisabled();
    expect(posts).toEqual([decisionPath]);
    await held.deliver();
  } finally { await held.dispose(); }
  await expect(panel.locator(".workflow-review-decision-history")).toContainText(text(locale, "Approved", "已批准"));
  // Concurrent decision must become evidence to review; no automatic resubmission.
  const headers = await savedResponse.request().allHeaders();
  const latest = await page.request.get(`http://127.0.0.1:17000${suitePath}/${suiteId}/review?workspace_id=${record.workspace_id}&application_id=${application.id}`, { headers });
  const latestReview = (await latest.json()).review;
  const external = await page.request.post(`http://127.0.0.1:17000${decisionPath}`, { headers, data: { workspace_id: record.workspace_id, application_id: application.id, expected_decision_version: 1, decision: "needs_review", review_digest: latestReview.review_digest } });
  expect((await external.json()).suite.current_decision_version).toBe(2);
  await button(panel, locale, "Record decision v2", "记录决定 v2").click();
  await expect(panel.getByRole("alert")).toContainText(text(locale, "The decision changed", "决定已变更"));
  await expect(button(panel, locale, "Record decision v3", "记录决定 v3")).toBeEnabled();
  expect(posts).toEqual([decisionPath, decisionPath]);
  const pattern = new RegExp(`/workflow-evaluation-suites/${suiteId}/review\\?`);
  await page.route(pattern, route => route.fulfill({ status: 503, contentType: "application/json", body: JSON.stringify({ failure_code: "workflow_evaluation_suite_store_unavailable", failure_summary: "RAW PRIVATE SERVER PROSE" }) }));
  await row.click();
  await expect(panel.getByRole("alert")).toContainText(text(locale, "Evidence storage is unavailable", "证据存储不可用"));
  locale = locale === "en-US" ? "zh-CN" : "en-US"; await switchLanguage(page, locale);
  await expect(panel.getByRole("alert")).toContainText(text(locale, "Evidence storage is unavailable", "证据存储不可用"));
  await expect(panel).not.toContainText("RAW PRIVATE SERVER PROSE");
  await page.unroute(pattern);
  await button(panel, locale, "Read review again", "重新读取审查").click();
  await expect(panel.locator(".workflow-review-decision-history > div")).toHaveCount(2);
  await layout(page, panel, testInfo, `suite-${locale}`);
  await page.reload();
  await selectApplication(page, application);
  await page.locator(".product-nav").getByRole("link", { name: text(locale, "Run history", "运行记录"), exact: true }).click();
  await page.locator('a.workflow-review-task[href="#workflow-evaluation-release-review"]').click();
  await expect(suites(page).locator(".workflow-run-history-live-row")).toHaveCount(1);
  await suites(page).locator(".workflow-run-history-live-row").click();
  await expect(suites(page).locator(".workflow-review-decision-history > div")).toHaveCount(2);
});
