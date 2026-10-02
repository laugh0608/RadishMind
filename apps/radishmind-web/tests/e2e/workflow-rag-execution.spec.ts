import { randomUUID } from "node:crypto";
import { test, expect, holdDraftResponse, isEndpoint, saveDraft, draftField, openDraft } from "./workflow-fixtures";
import { setTestLanguage } from "./ui-language";

const snapshotsRoute = "/v1/user-workspace/workflow-retrieval-snapshots";
const labels = {
  newSnapshot: ["New knowledge snapshot", "新建知识快照"], key: ["Snapshot key", "快照键"],
  name: ["Display name", "显示名称"], content: ["Content", "正文"], createV1: ["Create v1", "创建 v1"],
  createDraft: ["Create RAG retrieval v1 draft", "创建 RAG 检索 v1 草案"],
  snapshot: ["Active snapshot", "可用快照"], version: ["Exact version", "精确版本"],
  question: ["Question", "问题"], model: ["Model override", "指定模型"], temperature: ["Temperature", "采样温度"],
  execute: ["Start one retrieval execution", "启动一次检索执行"], executing: ["Executing…", "执行中…"],
  notSaved: ["Save the current exact draft before execution.", "执行前请保存当前精确草案。"],
  success: ["One lexical retrieval and one Gateway call completed with validated citations.", "已完成一次词法检索与一次 Gateway 调用，引用已通过校验。"],
  noEvidence: ["No matching evidence met the retrieval threshold.", "没有匹配证据达到检索阈值"],
  ineligible: ["The draft version, graph or execution parameters do not meet the execution requirements.", "草案版本、拓扑或执行参数不符合要求"],
  archived: ["This snapshot has been archived.", "此快照已归档"],
} as const;

test("RAG retrieval keeps exact bindings, pending execution, answers and failures across language changes", async ({ page, application }, testInfo) => {
  let locale: "en-US" | "zh-CN" = testInfo.project.use.locale === "zh-CN" ? "zh-CN" : "en-US";
  const text = (key: keyof typeof labels) => labels[key][locale === "zh-CN" ? 1 : 0];
  const requests: string[] = [];
  page.on("request", request => { if (new URL(request.url()).port === "17000") requests.push(`${request.method()} ${new URL(request.url()).pathname}`); });
  async function switchLanguage() {
    const count = requests.length;
    locale = locale === "zh-CN" ? "en-US" : "zh-CN";
    await page.evaluate(value => {
      localStorage.setItem("radishmind.uiLocale.v1", value);
      window.dispatchEvent(new StorageEvent("storage", { key: "radishmind.uiLocale.v1", newValue: value, storageArea: localStorage }));
    }, locale);
    setTestLanguage(page, locale);
    await expect(page.locator("html")).toHaveAttribute("lang", locale);
    await page.evaluate(() => new Promise<void>(resolve => requestAnimationFrame(() => requestAnimationFrame(() => resolve()))));
    expect(requests, "Changing language must not fetch or execute business resources").toHaveLength(count);
  }
  await page.locator('a[href="#application-configuration-draft"]').first().click();
  await page.locator('button[aria-controls="application-development-owner-surface"]').click();
  const snapshots = page.locator("#workflow-rag-snapshot-panel");
  const panel = page.locator("#workflow-rag-execution");
  await snapshots.getByRole("button", { name: text("newSnapshot"), exact: true }).click();
  const key = `exec_${randomUUID().replaceAll("-", "").slice(0, 14)}`;
  const material = "Retrieval guidance: this workflow provides advisory answers from reviewed immutable knowledge and does not write business data. 原始材料保持。";
  await snapshots.getByRole("textbox", { name: text("key"), exact: true }).fill(key);
  await snapshots.getByRole("textbox", { name: text("name"), exact: true }).fill("Evidence 原文 <literal>");
  await snapshots.getByRole("textbox", { name: new RegExp(`^${text("content")} [0-9]`) }).fill(material);
  const creation = page.waitForResponse(value => isEndpoint(value, snapshotsRoute, "POST"));
  await snapshots.getByRole("button", { name: text("createV1"), exact: true }).click();
  const creationResponse = await creation;
  expect(creationResponse.ok()).toBeTruthy();
  const { record: snapshot } = await creationResponse.json();
  await expect(snapshots.locator(".workflow-rag-record")).toContainText(snapshot.rag_ref);
  await panel.getByRole("button", { name: text("createDraft"), exact: true }).click();
  await expect(panel.getByRole("combobox", { name: text("snapshot"), exact: true }).locator(`option[value="${snapshot.snapshot_id}"]`)).toHaveCount(1);
  await panel.getByRole("combobox", { name: text("snapshot"), exact: true }).selectOption(snapshot.snapshot_id);
  await expect(panel.locator(".workflow-rag-scope-grid")).toContainText(snapshot.snapshot_digest);
  const draftId = (await draftField(page, "Draft").innerText()).trim();
  const executionPath = `/v1/user-workspace/workflow-drafts/${draftId}/retrieval-executions`;
  const query = "Retrieval guidance advisory immutable knowledge";
  await panel.getByRole("textbox", { name: text("question"), exact: true }).fill(query);
  await panel.getByRole("textbox", { name: text("model"), exact: true }).fill("mock");
  await panel.getByRole("spinbutton", { name: text("temperature"), exact: true }).fill("0.2");
  await expect(panel.getByRole("button", { name: text("execute"), exact: true })).toBeDisabled();
  await expect(panel.locator(".workflow-rag-eligibility")).toContainText(text("notSaved"));
  await switchLanguage();
  await expect(panel.locator(".workflow-rag-eligibility")).toContainText(text("notSaved"));
  const saved = await saveDraft(page, 1);
  await expect(panel.getByRole("button", { name: text("execute"), exact: true })).toBeEnabled();
  const held = await holdDraftResponse(page, executionPath, "POST");
  let answer: { answer: string; citations: { fragment_ref: string; claim_summary: string }[]; limitations: string[] };
  try {
    await panel.getByRole("button", { name: text("execute"), exact: true }).click();
    await held.arrived;
    await switchLanguage();
    await expect(panel.getByRole("button", { name: text("executing"), exact: true })).toBeDisabled();
    await expect(panel.getByRole("textbox", { name: text("question"), exact: true })).toHaveValue(query);
    await expect(panel.getByRole("textbox", { name: text("model"), exact: true })).toHaveValue("mock");
    await expect(panel.getByRole("spinbutton", { name: text("temperature"), exact: true })).toHaveValue("0.2");
    await expect(panel.getByRole("combobox", { name: text("version"), exact: true })).toHaveValue("1");
    const completion = page.waitForResponse(value => isEndpoint(value, executionPath, "POST"));
    await held.deliver();
    const response = await completion;
    const body = await response.json();
    expect(response.ok(), JSON.stringify(body)).toBeTruthy();
    expect(body.failure_code, JSON.stringify(body)).toBeNull();
    expect(response.request().postDataJSON()).toEqual({ workspace_id: body.workspace_id, application_id: application.id, draft_version: 1, input_text: query, model: "mock", temperature: 0.2 });
    expect(body.run.snapshot.rag_ref).toBe(snapshot.rag_ref);
    expect(body.run.side_effects).toEqual({ retrieval_calls: 1, provider_calls: 1, tool_calls: 0, confirmation_calls: 0, business_writes: 0, replay_writes: 0 });
    answer = body.retrieval_answer;
    expect(answer.citations.length).toBeGreaterThan(0);
    expect(JSON.stringify(body.run)).not.toContain(query);
    expect(JSON.stringify(body.run)).not.toContain(material);
  } finally { await held.dispose(); }
  await expect(panel).toContainText(text("success"));
  await expect(panel.locator(".workflow-rag-answer")).toContainText(answer!.answer);
  await switchLanguage();
  await expect(panel).toContainText(text("success"));
  for (const citation of answer!.citations) {
    await expect(panel.locator(".workflow-rag-answer")).toContainText(citation.fragment_ref);
    await expect(panel.locator(".workflow-rag-answer")).toContainText(citation.claim_summary);
  }
  for (const limitation of answer!.limitations) await expect(panel.locator(".workflow-rag-answer")).toContainText(limitation);
  expect(requests.filter(value => value === `POST ${executionPath}`)).toHaveLength(1);
  for (const width of [1440, 720, 390]) {
    await page.setViewportSize({ width, height: 900 });
    await panel.scrollIntoViewIfNeeded();
    const overflow = await panel.locator(".workflow-rag-scope-grid article, .workflow-rag-binding-grid, .workflow-rag-execution-form, .workflow-rag-answer").evaluateAll(elements => elements.filter(element => element.scrollWidth > element.clientWidth + 1).map(element => element.className || element.tagName));
    expect(overflow).toEqual([]);
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= document.documentElement.clientWidth + 1)).toBeTruthy();
    await testInfo.attach(`rag-execution-${locale}-${width}`, { body: await panel.screenshot(), contentType: "image/png" });
  }
  await page.setViewportSize({ width: 1440, height: 900 });
  await panel.getByRole("textbox", { name: text("question"), exact: true }).fill("zzzzqqqqxxxx");
  const noEvidence = page.waitForResponse(value => isEndpoint(value, executionPath, "POST"));
  await panel.getByRole("button", { name: text("execute"), exact: true }).click();
  const failed = await (await noEvidence).json();
  expect(failed.failure_code).toBe("workflow_rag_no_evidence");
  expect(failed.run.side_effects.provider_calls).toBe(0);
  await expect(panel.locator(".workflow-rag-answer")).toHaveCount(0);
  await expect(panel.locator(".failure-summary")).toContainText(text("noEvidence"));
  await switchLanguage();
  await expect(panel.locator(".failure-summary")).toContainText(text("noEvidence"));
  await expect(panel.getByRole("textbox", { name: text("question"), exact: true })).toHaveValue("zzzzqqqqxxxx");

  // Advance the real saved draft through its API while this tab retains v1.
  const headers = await saved.request().allHeaders(); delete headers["content-length"];
  const savePayload = saved.request().postDataJSON();
  const concurrent = await page.request.post("http://127.0.0.1:17000/v1/user-workspace/workflow-drafts", { headers, data: { ...savePayload, expected_draft_version: 1, expected_lifecycle_version: 1 } });
  expect(concurrent.ok(), await concurrent.text()).toBeTruthy();
  await panel.getByRole("textbox", { name: text("question"), exact: true }).fill(query);
  const stale = page.waitForResponse(value => isEndpoint(value, executionPath, "POST"));
  await panel.getByRole("button", { name: text("execute"), exact: true }).click();
  const staleBody = await (await stale).json();
  expect(staleBody.failure_code).toBe("workflow_rag_draft_ineligible");
  expect(staleBody.run).toBeNull();
  await expect(panel.locator(".failure-summary")).toContainText(text("ineligible"));
  await switchLanguage();
  await expect(panel.locator(".failure-summary")).toContainText(text("ineligible"));
  expect(requests.filter(value => value === `POST ${executionPath}`)).toHaveLength(3);

  // Refresh restores saved metadata, never the transient answer or query.
  await page.reload();
  await page.locator('button[aria-controls="application-development-owner-surface"]').click();
  await openDraft(page, draftId);
  await expect(panel.locator(".workflow-rag-answer")).toHaveCount(0);
  await expect(panel.getByRole("textbox", { name: text("question"), exact: true })).not.toHaveValue(query);
  await expect(panel.locator(".workflow-rag-scope-grid")).toContainText(snapshot.snapshot_digest);
});
