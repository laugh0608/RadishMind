import { test, expect, createDraft, saveDraft, openDraft, selectApplication, promptLabel, designer, draftField, holdDraftResponse, isEndpoint, draftRoute } from "./workflow-fixtures";
import { setTestLanguage, uiText, draftVersionText } from "./ui-language";
import type { Page } from "@playwright/test";

type Locale = "en-US" | "zh-CN";
const panelIds = ["workflow-draft-validation-inspector", "workflow-execution-plan-preview", "workflow-runtime-readiness-inspector", "workflow-review-handoff"] as const;
function languageSwitch(page: Page, initial: Locale) {
  let locale = initial;
  const requests: string[] = [];
  page.on("request", request => { if (new URL(request.url()).port === "17000") requests.push(`${request.method()} ${new URL(request.url()).pathname}`); });
  return {
    text: (en: string, zh: string) => locale === "zh-CN" ? zh : en,
    locale: () => locale,
    requests,
    async switch() {
      const count = requests.length;
      locale = locale === "zh-CN" ? "en-US" : "zh-CN";
      await page.evaluate(value => {
        localStorage.setItem("radishmind.uiLocale.v1", value);
        window.dispatchEvent(new StorageEvent("storage", { key: "radishmind.uiLocale.v1", newValue: value, storageArea: localStorage }));
      }, locale);
      setTestLanguage(page, locale);
      await expect(page.locator("html")).toHaveAttribute("lang", locale);
      await page.evaluate(() => new Promise<void>(resolve => requestAnimationFrame(() => requestAnimationFrame(() => resolve()))));
      expect(requests, "Changing UI language must not issue a business request").toHaveLength(count);
    },
  };
}

test("offline inspection and handoff explain passed, review and blocked graphs in both languages", async ({ page, application: _application }, testInfo) => {
  const lang = languageSwitch(page, testInfo.project.use.locale === "zh-CN" ? "zh-CN" : "en-US");
  const validation = page.locator(`#${panelIds[0]}`), plan = page.locator(`#${panelIds[1]}`), readiness = page.locator(`#${panelIds[2]}`), handoff = page.locator(`#${panelIds[3]}`);
  for (const id of panelIds) await expect(page.locator(`#${id}`)).toBeVisible();
  const refs = () => handoff.locator("code").allTextContents();
  const originalRefs = await refs();
  for (let round = 0; round < 2; round++) {
    await expect(validation).toContainText(lang.text("A draft must start from a context collection lane before model reasoning.", "草案须先收集上下文，再进入模型推理。"));
    await expect(plan).toContainText(lang.text("This preview has no workflow executor or tool adapter binding.", "此预览未绑定工作流执行器或工具适配器。"));
    await expect(readiness).toContainText(lang.text("Prerequisites are grouped by executor", "前置条件按执行器"));
    await expect(handoff).toContainText(lang.text("is prepared for human review", "已整理为人工审查材料"));
    await expect(handoff).toContainText(lang.text("Which structural, contract, or blocked capability findings need review", "哪些结构、契约或受阻能力问题需要审查"));
    await expect(handoff).toContainText(lang.text("is reviewed against", "审查草案"));
    expect(await refs()).toEqual(originalRefs);
    await lang.switch();
  }
  const id = await createDraft(page);
  await expect(validation.locator(".section-heading > .status-badge")).toHaveText(lang.text("Passed", "通过"));
  const original = "Literal 输入 <script> preserves focus";
  await (await promptLabel(page)).fill(original); await (await promptLabel(page)).focus();
  await lang.switch();
  await expect(await promptLabel(page)).toHaveValue(original); await expect(await promptLabel(page)).toBeFocused();
  await expect(plan).toContainText(original);
  await expect(validation).toContainText(lang.text("Checks cover the bounded node allowlist", "检查覆盖受限节点类型"));
  const edgeActions = designer(page).locator(".workflow-node-designer-inspector-details").filter({ has: page.getByRole("button", { name: lang.text("Remove edge", "移除连线"), exact: true, includeHidden: true }) });
  await edgeActions.locator("summary").click();
  await edgeActions.getByRole("button", { name: lang.text("Remove edge", "移除连线"), exact: true }).first().click();
  await expect(validation.locator(".section-heading > .status-badge")).toHaveText(lang.text("Blocked", "已阻断"));
  const topology = validation.locator(".workflow-draft-structural-check").filter({ hasText: "executor_v0_topology" });
  await expect(topology).toContainText(lang.text("Every node stays on an acyclic path", "每个节点都位于"));
  await expect(handoff).toContainText(lang.text("Does this graph finding identify the node or edge context", "该图问题是否标明"));
  await expect(handoff).toContainText(id);
  for (const width of [1440, 720, 390]) {
    await page.setViewportSize({ width, height: 900 });
    for (const panelId of panelIds) {
      const panel = page.locator(`#${panelId}`);
      const clipped = await panel.locator("article, dl > div, h5, p, .workflow-workspace-review-token-list").evaluateAll(elements => elements.filter(element => element.scrollWidth > element.clientWidth + 1).map(element => element.className || element.tagName));
      expect(clipped, `${panelId} ${width} must fit its container`).toEqual([]);
      await panel.evaluate(element => element.scrollIntoView({ block: "start" }));
      await testInfo.attach(`${panelId}-${lang.locale()}-${width}`, { body: await page.screenshot(), contentType: "image/png" });
    }
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= document.documentElement.clientWidth + 1)).toBeTruthy();
  }
  expect(lang.requests.filter(request => request.startsWith("POST "))).toEqual([]);
  const persisted = await page.evaluate(() => JSON.stringify({ local: { ...localStorage }, session: { ...sessionStorage }, url: location.href }));
  expect(persisted).not.toContain(original); expect(persisted).not.toContain("Message");
});

test("saved draft conflict handoff translates pending and recovery copy without overwriting local edits", async ({ page, application }, testInfo) => {
  const id = await createDraft(page); const saved = await saveDraft(page, 1);
  const headers = await saved.request().allHeaders(); delete headers["content-length"];
  const payload = saved.request().postDataJSON();
  const concurrent = await page.request.post(`http://127.0.0.1:17000${draftRoute}`, { headers, data: { ...payload, expected_draft_version: 1, expected_lifecycle_version: 1 } });
  expect(concurrent.ok(), await concurrent.text()).toBeTruthy();
  expect((await concurrent.json()).current_draft_version).toBe(2);
  const original = "Conflict 本地原文 <literal>";
  await (await promptLabel(page)).fill(original);
  const handoff = page.locator("#workflow-review-handoff");
  const lang = languageSwitch(page, testInfo.project.use.locale === "zh-CN" ? "zh-CN" : "en-US");
  const held = await holdDraftResponse(page, draftRoute, "POST");
  try {
    await designer(page).getByRole("button", { name: uiText(page, "Save draft"), exact: true }).click();
    await held.arrived; await lang.switch();
    await expect(await promptLabel(page)).toHaveValue(original);
    await expect(handoff).toContainText(lang.text("Active draft validation", "活动草案校验"));
    const response = page.waitForResponse(value => isEndpoint(value, draftRoute, "POST"));
    await held.deliver(); const result = await response;
    expect((await result.json()).failure_code).toBe("draft_version_conflict");
    expect(JSON.stringify(result.request().postDataJSON())).not.toMatch(/Message|uiLocale/);
  } finally { await held.dispose(); }
  await expect(handoff).toContainText("draft_version_conflict");
  await expect(handoff).toContainText(lang.text("Version conflict keeps local draft", "版本冲突后，本地草案"));
  await expect(handoff).toContainText(lang.text("no saved version has been opened or merged", "尚未打开或合并任何已保存版本"));
  await lang.switch();
  await expect(handoff).toContainText(lang.text("Version conflict keeps local draft", "版本冲突后，本地草案"));
  await expect(await promptLabel(page)).toHaveValue(original);
  await designer(page).getByRole("button", { name: lang.text("Continue local draft", "继续本地草案"), exact: true }).click();
  await lang.switch();
  await expect(handoff).toContainText(lang.text("the next save retries against saved version 2", "下次保存将基于已保存版本 2 重试"));
  await expect(await promptLabel(page)).toHaveValue(original);
  expect(lang.requests.filter(request => request.startsWith("POST "))).toHaveLength(1);
  await saveDraft(page, 3);
  await expect(handoff).not.toContainText("draft_version_conflict");
  await page.reload(); await selectApplication(page, application); await openDraft(page, id);
  await expect(await promptLabel(page)).toHaveValue(original);
  await expect(draftField(page, "Version")).toHaveText(draftVersionText(page, 3, 1));
  await expect(handoff).not.toContainText("draft_version_conflict");
  await lang.switch();
  await expect(handoff).toContainText(lang.text("Active draft validation", "活动草案校验"));
  expect(lang.requests.filter(request => /executions|\/runs|decisions|activations/.test(request) && request.startsWith("POST "))).toEqual([]);
});
