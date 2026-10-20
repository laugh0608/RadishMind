import type { Locator } from "@playwright/test";
import { test, expect, holdDraftResponse, isEndpoint, finishResponseRender } from "./workflow-fixtures";
import { setTestLanguage, uiText } from "./ui-language";
import { prepareRAGApplication } from "./workflow-rag-application-fixtures";
import { promotionRequest } from "./workflow-rag-promotion-fixtures";

// Independent accessible names and recovery messages, not implementation resource imports.
const en = {
  assignment: { load: "Load current assignment", activate: "Activate runtime assignment", replace: "Replace runtime assignment", revoke: "Revoke runtime assignment", refresh: "Refresh current version and keep reason", recording: "Recording decision…", reason: "Sanitized decision reason", revoked: "This assignment is revoked. It cannot authorize an invocation.", conflict: "The assignment changed. Your reason is retained; refresh before deciding again." },
  invocation: { input: "Bounded input", invoke: "Invoke active RAG assignment", clear: "Clear sensitive memory" },
  failure: { workflow_rag_runtime_assignment_not_found: "No runtime assignment exists for this application.", workflow_rag_runtime_no_evidence: "The active candidate snapshot contains no evidence for this input.", workflow_rag_runtime_assignment_revoked: "The current runtime assignment is revoked." },
  status: { idle: "Not started" },
};
const zh = {
  assignment: { load: "加载当前运行分配", activate: "激活运行分配", replace: "替换运行分配", revoke: "撤销运行分配", refresh: "刷新当前版本并保留理由", recording: "正在记录决定…", reason: "脱敏决定理由", revoked: "此运行分配已撤销，不能用于授权调用。", conflict: "运行分配已变更，当前理由已保留；请刷新后重新决定。" },
  invocation: { input: "有界输入", invoke: "调用已激活的 RAG 运行分配", clear: "清除敏感内存" },
  failure: { workflow_rag_runtime_assignment_not_found: "此应用尚无运行分配。", workflow_rag_runtime_no_evidence: "当前候选快照没有与此输入匹配的证据。", workflow_rag_runtime_assignment_revoked: "当前运行分配已撤销。" },
  status: { idle: "尚未开始" },
};

test("application RAG preserves pending input, answers and CAS reasons across languages and clears sensitive state", async ({ page, application }, testInfo) => {
  test.setTimeout(120_000);
  const evidence = await prepareRAGApplication(page, application);
  let locale: "en-US" | "zh-CN" = testInfo.project.use.locale === "zh-CN" ? "zh-CN" : "en-US";
  const messages = () => locale === "zh-CN" ? zh : en;
  const assignment = page.locator(".workflow-rag-runtime-assignment");
  const invocation = page.locator("#application-rag-invocation");
  const runtimePath = `/v1/user-workspace/applications/${application.id}/workflow-rag-runtime-assignment`;
  const invocationPath = "/v1/application-rag/invocations";
  const requests: string[] = [];
  page.on("request", request => { if (new URL(request.url()).port === "17000") requests.push(`${request.method()} ${new URL(request.url()).pathname}`); });
  const assignmentButton = (key: "load" | "activate" | "replace" | "revoke" | "refresh" | "recording") => assignment.getByRole("button", { name: messages().assignment[key], exact: true });
  const reasonBox = () => assignment.getByRole("textbox", { name: messages().assignment.reason, exact: true });
  const inputBox = () => invocation.getByRole("textbox", { name: messages().invocation.input, exact: true });
  const invokeButton = () => invocation.getByRole("button", { name: messages().invocation.invoke, exact: true });
  async function switchLanguage() {
    const count = requests.length; locale = locale === "zh-CN" ? "en-US" : "zh-CN";
    await page.evaluate(value => { localStorage.setItem("radishmind.uiLocale.v1", value); window.dispatchEvent(new StorageEvent("storage", { key: "radishmind.uiLocale.v1", newValue: value, storageArea: localStorage })); }, locale);
    setTestLanguage(page, locale); await expect(page.locator("html")).toHaveAttribute("lang", locale);
    await page.evaluate(() => new Promise<void>(resolve => requestAnimationFrame(() => requestAnimationFrame(() => resolve()))));
    expect(requests, "Changing language must not load, decide or invoke").toHaveLength(count);
  }
  async function openCandidate(candidateId: string) {
    await page.locator('a[href="#application-publish-review"]:visible').first().click();
    const owner = page.locator('button[aria-controls="application-development-owner-surface"]');
    if (await owner.getAttribute("aria-expanded") === "false") await owner.click();
    const panel = page.locator("#application-publish-review");
    await panel.getByRole("button", { name: locale === "zh-CN" ? "刷新候选版本" : "Refresh candidates", exact: true }).click();
    await panel.locator(".application-publish-candidate-list button").filter({ hasText: candidateId }).click();
    await expect(assignment).toBeVisible();
  }
  async function handoff() {
    await page.locator('a[href="#workspace-api-keys"]:visible').first().click();
    const keys = page.locator("#workspace-api-keys");
    await keys.locator(".api-key-issue-disclosure > summary").click();
    await keys.getByRole("textbox", { name: uiText(page, "Display name"), exact: true }).fill("RAG synthetic invocation");
    for (const label of await keys.locator("fieldset label").all()) if (!(await label.innerText()).includes("application_rag:invoke")) await label.getByRole("checkbox").uncheck();
    await keys.getByRole("button", { name: uiText(page, "Issue API key"), exact: true }).click();
    const use = keys.getByRole("button", { name: locale === "zh-CN" ? "在应用 RAG 中使用" : "Use in Application RAG", exact: true });
    await expect(use).toBeVisible();
    const token = await keys.locator(".api-key-one-time-token").innerText();
    await use.click(); await expect(invocation).toBeVisible(); await expect(inputBox()).toBeEnabled();
    return token;
  }
  async function screenshot(panel: Locator, name: string) {
    expect(await panel.evaluate(element => getComputedStyle(element).backgroundColor), "Cards need an opaque readable surface").not.toBe("rgba(0, 0, 0, 0)");
    for (const width of [1440, 720, 390]) {
      await page.setViewportSize({ width, height: 900 }); await panel.scrollIntoViewIfNeeded();
      expect(await panel.locator("article, dl > div, button, code").evaluateAll(elements => elements.filter(element => element.scrollWidth > element.clientWidth + 1).map(element => element.className))).toEqual([]);
      expect(await page.evaluate(() => document.documentElement.scrollWidth <= document.documentElement.clientWidth + 1)).toBeTruthy();
      await testInfo.attach(`${name}-${locale}-${width}`, { body: await panel.screenshot(), contentType: "image/png" });
    }
    await page.setViewportSize({ width: 1440, height: 900 });
  }
  await openCandidate(evidence.candidateId); await assignmentButton("load").click();
  await expect(assignment).toContainText(messages().failure.workflow_rag_runtime_assignment_not_found);
  const reason = "Reviewed activation 原始人工理由 <literal>"; await reasonBox().fill(reason);
  const heldActivation = await holdDraftResponse(page, `${runtimePath}/decisions`, "POST");
  try {
    await assignmentButton("activate").click(); await heldActivation.arrived; await switchLanguage();
    await expect(reasonBox()).toHaveValue(reason); await expect(assignmentButton("recording")).toBeDisabled();
    const completion = page.waitForResponse(response => isEndpoint(response, `${runtimePath}/decisions`, "POST")); await heldActivation.deliver();
    const response = await completion, body = await response.json(); expect(body.failure_code).toBeNull();
    expect(response.request().postDataJSON()).toEqual({ workspace_id: "workspace_demo", expected_record_version: 0, decision: "activate", publish_candidate_id: evidence.candidateId, reason });
    expect(body.assignment.state).toBe("active"); expect(body.events.at(-1).reason).toBe(reason);
  } finally { await heldActivation.dispose(); }
  await expect(assignment).toContainText(evidence.binding.binding_id); await switchLanguage(); await screenshot(assignment, "rag-assignment");
  const token = await handoff();
  const input = "Promotion authority evidence 原始问题 <literal>"; await inputBox().fill(input); await switchLanguage(); await expect(inputBox()).toHaveValue(input);
  const heldInvocation = await holdDraftResponse(page, invocationPath, "POST"); let runId = "", answer = "";
  try {
    await invokeButton().click(); await heldInvocation.arrived; await switchLanguage(); await expect(inputBox()).toHaveValue(input); await expect(inputBox()).toBeDisabled();
    const completion = page.waitForResponse(response => isEndpoint(response, invocationPath, "POST")); await heldInvocation.deliver();
    const response = await completion, body = await response.json(); expect(body.failure_code, JSON.stringify(body)).toBeNull();
    expect(response.request().postDataJSON()).toEqual({ input }); expect(body.run.selected_provider).toBe("mock"); expect(body.run.side_effects.provider_calls).toBe(1); expect(body.run.side_effects.retrieval_calls).toBe(1);
    runId = body.run.run_id; answer = body.answer.answer;
    expect(body.answer.citations[0].fragment_ref).toBe("official_promotion");
    expect(JSON.stringify(body.run)).not.toContain(input); expect(JSON.stringify(body.run)).not.toContain(answer); expect(JSON.stringify(body.run)).not.toContain(token);
  } finally { await heldInvocation.dispose(); }
  await expect(invocation.locator(".workflow-rag-application-answer")).toContainText(answer); await switchLanguage();
  await expect(invocation.locator(".workflow-rag-application-answer")).toContainText(answer); await expect(invocation).toContainText(runId); await expect(inputBox()).toHaveValue("");
  await screenshot(invocation, "rag-invocation");
  const storage = await page.evaluate(() => JSON.stringify({ local: { ...localStorage }, session: { ...sessionStorage }, url: location.href }));
  for (const sensitive of [token, input, answer]) expect(storage).not.toContain(sensitive);
  await inputBox().fill("xylophonicquasarunmatched");
  const noEvidence = page.waitForResponse(response => isEndpoint(response, invocationPath, "POST")); await invokeButton().click();
  const failed = await (await noEvidence).json(); expect(failed.failure_code).toBe("workflow_rag_runtime_no_evidence"); expect(failed.run.side_effects.provider_calls).toBe(0);
  await expect(invocation.getByRole("alert")).toContainText(messages().failure.workflow_rag_runtime_no_evidence); await switchLanguage(); await expect(invocation.getByRole("alert")).toContainText(messages().failure.workflow_rag_runtime_no_evidence);
  await inputBox().fill(input);
  const heldClear = await holdDraftResponse(page, invocationPath, "POST");
  try {
    await invokeButton().click(); await heldClear.arrived;
    await invocation.getByRole("button", { name: messages().invocation.clear, exact: true }).click();
    const completion = page.waitForResponse(response => isEndpoint(response, invocationPath, "POST")); await heldClear.deliver(); await finishResponseRender(page, await completion);
    await expect(inputBox()).toHaveValue(""); await expect(inputBox()).toBeDisabled(); await expect(invocation.locator(".workflow-rag-application-answer")).toHaveCount(0);
    await expect(invocation.locator(".section-heading > .status-badge")).toHaveText(messages().status.idle);
  } finally { await heldClear.dispose(); }
  const secondCandidate = await evidence.publish(); await openCandidate(secondCandidate); await assignmentButton("load").click();
  await expect(assignment).toContainText(evidence.candidateId); const conflictReason = "Replace reviewed candidate 原始理由"; await reasonBox().fill(conflictReason);
  await promotionRequest(page, application.id, `${runtimePath}/decisions`, ["workflow_rag_runtime:write"], { workspace_id: "workspace_demo", expected_record_version: 1, decision: "replace", publish_candidate_id: secondCandidate, reason: "Another reviewer selected the approved candidate." });
  const conflict = page.waitForResponse(response => isEndpoint(response, `${runtimePath}/decisions`, "POST")); await assignmentButton("replace").click();
  expect((await (await conflict).json()).failure_code).toBe("workflow_rag_runtime_assignment_version_conflict");
  await switchLanguage(); await expect(assignment).toContainText(messages().assignment.conflict); await expect(reasonBox()).toHaveValue(conflictReason);
  await assignmentButton("refresh").click(); await expect(assignment).toContainText(secondCandidate); await expect(reasonBox()).toHaveValue(conflictReason);
  const revoke = page.waitForResponse(response => isEndpoint(response, `${runtimePath}/decisions`, "POST")); await assignmentButton("revoke").click();
  expect((await (await revoke).json()).assignment.state).toBe("revoked");
  await handoff(); await inputBox().fill(input);
  const revoked = page.waitForResponse(response => isEndpoint(response, invocationPath, "POST")); await invokeButton().click(); const revokedBody = await (await revoked).json();
  expect(revokedBody.failure_code).toBe("workflow_rag_runtime_assignment_revoked"); expect(revokedBody.run).toBeNull();
  await expect(invocation.getByRole("alert")).toContainText(messages().failure.workflow_rag_runtime_assignment_revoked);
  expect(requests.filter(value => value === `POST ${invocationPath}`)).toHaveLength(4);
  await page.reload(); await openCandidate(secondCandidate); await assignmentButton("load").click(); await expect(assignment).toContainText(messages().assignment.revoked);
});
