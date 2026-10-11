import { test, expect, createDraft, saveDraft, isEndpoint } from "./workflow-fixtures";
import { uiText } from "./ui-language";
import { evidenceLanguage, evidenceLayout, holdEvidenceResponse } from "./results-evidence-fixtures";

const sessionsPath = "/v1/user-workspace/application-sessions";
test.afterEach(async ({ page }) => { if (!page.isClosed()) await page.close(); });
test("final interaction preserves input, pending and canceled state across languages and restores metadata only", async ({ page, application }, info) => {
  test.setTimeout(180_000);
  await createDraft(page); await saveDraft(page, 1);
  await page.getByRole("link", { name: uiText(page, "Open Workflow definition owner"), exact: true }).click();
  const promoted = page.waitForResponse(response => response.request().method() === "POST" && new URL(response.url()).pathname === "/v1/user-workspace/workflow-definition-candidates");
  await page.locator("#workflow-definition-promotion").getByRole("button", { name: uiText(page, "Create promotion candidate"), exact: true }).click();
  const promotion = await (await promoted).json();
  const definitionId = promotion.candidate.definition_id as string;
  await page.getByRole("button", { name: uiText(page, "Append review v1"), exact: true }).click();
  await page.getByRole("button", { name: uiText(page, "Activate · expected pointer v0"), exact: true }).click();
  await expect(page.getByRole("button", { name: uiText(page, "Start exact version run"), exact: true })).toBeEnabled();
  const ui = evidenceLanguage(page, info);
  const panel = page.locator("#application-interaction-session");
  async function open() {
    await page.evaluate(() => { location.hash = "application-interaction-session"; });
    const owner = page.locator('button[aria-controls="application-development-owner-surface"]');
    if (await owner.getAttribute("aria-expanded") === "false") await owner.click();
    const details = page.locator(".application-access-related-surfaces");
    if (!(await details.getAttribute("open"))) {
      if (!(await details.evaluate(element => (element as HTMLDetailsElement).open))) await details.locator("summary").click();
    }
    await expect(panel).toBeVisible(); await page.waitForLoadState("networkidle");
  }
  const button = (en: string, zh: string) => panel.getByRole("button", { name: ui.copy(en, zh), exact: true });
  const field = (en: string, zh: string) => panel.getByRole("textbox", { name: ui.copy(en, zh), exact: true });
  await open();
  await expect(panel).toContainText(ui.copy("No sessions match this state.", "当前状态下没有会话。"));
  await field("Active definition ID", "已激活定义 ID").fill(definitionId);
  const heldCreate = await holdEvidenceResponse(page, sessionsPath, "POST");
  let sessionId = "";
  try {
    const created = page.waitForResponse(response => isEndpoint(response, sessionsPath, "POST"));
    await button("Create bound session", "创建绑定会话").click(); await heldCreate.arrived;
    await ui.switch(); await expect(button("Creating…", "正在创建…")).toBeDisabled();
    await expect(field("Active definition ID", "已激活定义 ID")).toHaveValue(definitionId);
    await heldCreate.deliver(); const body = await (await created).json();
    expect(body.failure_code, JSON.stringify(body)).toBeNull(); sessionId = body.session.session_id;
  } finally { await heldCreate.dispose(); }
  await expect(panel).toContainText(sessionId);
  const rawInput = "Synthetic interaction 原始输入 <literal>; no business action.";
  await field("Transient input", "临时输入").fill(rawInput);
  await field("Condition values JSON", "条件值 JSON").fill("{");
  const writes = ui.requests.filter(request => request.startsWith("POST ")).length;
  await button("Run turn", "运行此轮").click(); await ui.switch();
  await expect(panel).toContainText("application_session_workflow_options_invalid");
  await expect(field("Transient input", "临时输入")).toHaveValue(rawInput);
  expect(ui.requests.filter(request => request.startsWith("POST "))).toHaveLength(writes);
  await field("Condition values JSON", "条件值 JSON").fill("{}");
  const turnsPath = `${sessionsPath}/${sessionId}/turns`;
  const heldTurn = await holdEvidenceResponse(page, turnsPath, "POST");
  try {
    await button("Run turn", "运行此轮").click(); await heldTurn.arrived; await ui.switch();
    await expect(button("Running…", "正在运行…")).toBeDisabled();
    await expect(panel.locator(".user-message p")).toHaveText(rawInput);
    await button("Cancel current turn", "取消当前轮次").click();
    await heldTurn.dispose(); await ui.switch();
    await expect(panel.locator(".application-interaction-transcript")).toContainText(ui.copy("Canceled", "已取消"));
    await expect(panel.locator(".assistant-message")).toHaveCount(0);
  } finally { await heldTurn.dispose(); }
  await button("Reload sessions", "重新加载会话").click();
  await panel.locator(".application-interaction-session-list button").filter({ hasText: sessionId }).click();
  await expect(panel.locator(".application-interaction-turn-history li")).toHaveCount(1);
  await expect(panel.locator(".application-interaction-transcript")).toHaveCount(0);
  await expect(field("Transient input", "临时输入")).toHaveValue("");
  await evidenceLayout(page, panel, info, "final-interaction");
  await button("Close session", "关闭会话").click();
  await expect(button("Close session", "关闭会话")).toBeDisabled(); await ui.switch();
  await expect(panel.locator(".section-heading > .status-badge")).toHaveText(ui.copy("Closed", "已关闭"));
  await expect(button("Run turn", "运行此轮")).toBeDisabled();
  expect(ui.requests.filter(request => request === `POST ${turnsPath}`)).toHaveLength(1);
  await expect(panel).toContainText(application.name);
});
