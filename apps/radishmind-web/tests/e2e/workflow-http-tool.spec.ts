import { test, expect, createDraft, saveDraft, openDraft, selectApplication, holdDraftResponse, isEndpoint, draftRoute } from "./workflow-fixtures";
import { setTestLanguage, uiText } from "./ui-language";

// Independent visible expectations; business locale deliberately differs from both UI languages.
const labels = {
  resource: ["Resource key", "资源标识"], locale: ["Tool locale", "工具语言参数"],
  create: ["Create immutable plan", "创建不可变计划"], creating: ["Creating…", "正在创建…"],
  approve: ["Approve qualification", "批准执行资格"], reject: ["Reject", "拒绝"],
  rejected: ["Rejected", "已拒绝"], approved: ["Approved", "已批准"], consumed: ["Consumed", "已消费"],
  conflict: ["The plan changed. Its current durable version was reloaded", "计划已变化，已重新读取当前持久化版本"],
  deciding: ["Recording the decision against the displayed record version.", "正在针对当前显示的记录版本记录决定。"],
  localeInvalid: ["locale must be a short language tag", "locale 必须是"],
  input: ["Review prompt", "审查输入"], model: ["Model", "模型"],
  execute: ["Execute approved plan", "执行已批准计划"], executing: ["Executing single attempt…", "正在执行单次尝试…"],
  transport: ["The HTTP Tool transport failed.", "HTTP Tool 传输失败。"],
  planSafety: ["Plan execution eligibility", "计划执行资格"], runSafety: ["Run safety evidence", "运行安全证据"],
} as const;

for (const source of ["saved", "definition"] as const) {
  test(`${source} HTTP Tool keeps confirmation, one attempt and durable recovery across language changes`, async ({ page, application }, testInfo) => {
    let locale: "en-US" | "zh-CN" = testInfo.project.use.locale === "zh-CN" ? "zh-CN" : "en-US";
    const text = (key: keyof typeof labels) => labels[key][locale === "zh-CN" ? 1 : 0];
    const draftId = await createDraft(page);
    const saved = await saveDraft(page, 1);
    const payload = saved.request().postDataJSON();
    const headers = await saved.request().allHeaders(); delete headers["content-length"];
    const prompt = payload.draft.nodes.find((node: { node_type: string }) => node.node_type === "prompt");
    const modelNode = payload.draft.nodes.find((node: { node_type: string }) => node.node_type === "llm");
    const output = payload.draft.nodes.find((node: { node_type: string }) => node.node_type === "output");
    const tool = { ...prompt, node_id: "node_http_tool", node_type: "http_tool", label: "HTTP Tool 原始节点", tool_ref: "workflow.http.reviewed-json-read.v1", risk_level: "medium", requires_confirmation: true };
    const nodes = [prompt, tool, modelNode, output];
    const seeded = await page.request.post(`http://127.0.0.1:17000${draftRoute}`, { headers, data: {
      ...payload, expected_draft_version: 1, expected_lifecycle_version: 1,
      draft: { ...payload.draft, nodes, edges: nodes.slice(1).map((node, index) => ({ edge_id: `http_edge_${index}`, from_node_id: nodes[index].node_id, to_node_id: node.node_id, condition_summary: "Reviewed linear sequence" })), tool_refs: [tool.tool_ref], additional_fields: {} },
    } });
    expect(seeded.ok(), await seeded.text()).toBeTruthy();
    expect((await seeded.json()).current_draft_version).toBe(2);
    await page.reload();
    await selectApplication(page, application);
    await openDraft(page, draftId);
    let expectedSourceId = draftId;
    if (source === "definition") {
      await page.getByRole("link", { name: uiText(page, "Open Workflow definition owner"), exact: true }).click();
      const candidateResponse = page.waitForResponse(value => isEndpoint(value, "/v1/user-workspace/workflow-definition-candidates", "POST"));
      await page.getByRole("button", { name: uiText(page, "Create promotion candidate"), exact: true }).click();
      const candidate = (await (await candidateResponse).json()).candidate;
      expect(candidate.snapshot.execution_profile).toBe("workflow_definition_http_tool_v1");
      expectedSourceId = candidate.definition_id;
      await page.getByRole("button", { name: uiText(page, "Append review v1"), exact: true }).click();
      await page.getByRole("button", { name: uiText(page, "Activate · expected pointer v0"), exact: true }).click();
    }
    const prefix = source === "definition" ? "workflow-definition-http-tool" : "workflow-http-tool";
    const action = page.locator(`#${prefix}-action-review`);
    const execution = page.locator(`#${prefix}-execution`);
    await expect(action).toBeVisible();
    const sourceId = await action.locator(".workflow-executor-summary-grid > article").first().locator("strong").innerText();
    expect(sourceId).toBe(expectedSourceId);
    const requests: string[] = [];
    page.on("request", request => { if (new URL(request.url()).port === "17000") requests.push(`${request.method()} ${new URL(request.url()).pathname}`); });
    const executions = () => requests.filter(value => value.endsWith("/executions") && value.startsWith("POST "));
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
      expect(requests, "Language changes must not send business requests").toHaveLength(count);
    }
    await action.getByRole("textbox", { name: text("resource"), exact: true }).fill("catalog/review:literal-1");
    await action.getByRole("textbox", { name: text("locale"), exact: true }).fill("not a locale");
    await expect(action.getByRole("alert")).toContainText(text("localeInvalid"));
    await switchLanguage();
    await expect(action.getByRole("alert")).toContainText(text("localeInvalid"));
    const localeInput = () => action.getByRole("textbox", { name: text("locale"), exact: true });
    await localeInput().fill("ja-JP"); await localeInput().focus();
    await switchLanguage();
    await expect(localeInput()).toHaveValue("ja-JP"); await expect(localeInput()).toBeFocused();
    // Hold real server responses without replacing their payloads.
    const createPath = source === "saved" ? `${draftRoute}/${draftId}/tool-action-plans`
      : `/v1/user-workspace/workflow-definitions/${sourceId}/tool-action-plans`;
    const heldCreate = await holdDraftResponse(page, createPath, "POST");
    let planId = "";
    try {
      await action.getByRole("button", { name: text("create"), exact: true }).click();
      await heldCreate.arrived; await switchLanguage();
      await expect(action.getByRole("button", { name: text("creating"), exact: true })).toBeDisabled();
      await expect(localeInput()).toHaveValue("ja-JP");
      const completion = page.waitForResponse(value => isEndpoint(value, createPath, "POST"));
      await heldCreate.deliver(); const response = await completion;
      expect(response.ok(), await response.text()).toBeTruthy();
      const result = await response.json(); planId = result.action_plan.plan_id;
      expect(result.action_plan.schema_version).toBe(`workflow_http_tool_action_plan.v${source === "saved" ? 1 : 2}`);
      expect(response.request().postDataJSON().public_arguments).toEqual({ resource_key: "catalog/review:literal-1", locale: "ja-JP" });
    } finally { await heldCreate.dispose(); }
    let decisionPath = `/v1/user-workspace/workflow-tool-action-plans/${planId}/decisions`;
    const rejection = page.waitForResponse(value => isEndpoint(value, decisionPath, "POST"));
    await action.getByRole("button", { name: text("reject"), exact: true }).click();
    expect((await rejection).ok()).toBeTruthy();
    await switchLanguage(); await expect(action.locator(".workflow-executor-record .status-badge")).toHaveText(text("rejected"));
    await expect(execution.getByRole("button", { name: text("execute"), exact: true })).toBeDisabled();
    expect(executions()).toEqual([]);
    const secondPlan = page.waitForResponse(value => isEndpoint(value, createPath, "POST"));
    await action.getByRole("button", { name: text("create"), exact: true }).click();
    const second = await (await secondPlan).json(); planId = second.action_plan.plan_id;
    decisionPath = `/v1/user-workspace/workflow-tool-action-plans/${planId}/decisions`;
    // A concurrent request changes the actual durable plan, leaving this tab at v1.
    const decisionHeaders = { ...headers, "x-radishmind-dev-read-scopes": "workflow_tool_actions:confirm", "x-radishmind-dev-read-membership-permissions": "workflow_tool_actions:confirm" };
    const deferred = await page.request.post(`http://127.0.0.1:17000${decisionPath}`, { headers: decisionHeaders, data: { workspace_id: "workspace_demo", application_id: application.id, expected_record_version: 1, decision: "defer" } });
    expect(deferred.ok(), await deferred.text()).toBeTruthy();
    await action.getByRole("button", { name: text("approve"), exact: true }).click();
    await expect(action).toContainText(text("conflict"));
    await switchLanguage(); await expect(action).toContainText(text("conflict"));
    expect(executions()).toEqual([]);
    const heldDecision = await holdDraftResponse(page, decisionPath, "POST");
    try {
      await action.getByRole("button", { name: text("approve"), exact: true }).click();
      await heldDecision.arrived; await switchLanguage();
      await expect(action).toContainText(text("deciding"));
      await expect(action.getByRole("button", { name: text("approve"), exact: true })).toBeDisabled();
      const completion = page.waitForResponse(value => isEndpoint(value, decisionPath, "POST"));
      await heldDecision.deliver(); const response = await completion;
      expect(response.request().postDataJSON().expected_record_version).toBe(2);
      expect((await response.json()).action_plan.status).toBe("approved");
    } finally { await heldDecision.dispose(); }
    await expect(action.locator(".workflow-executor-record .status-badge")).toHaveText(text("approved"));
    await expect(action.getByLabel(text("planSafety"), { exact: true })).toBeVisible();
    expect(executions()).toEqual([]);
    const input = "Synthetic HTTP review 输入 <literal>; no business action.";
    await execution.getByRole("textbox", { name: text("input"), exact: true }).fill(input);
    await execution.getByRole("textbox", { name: text("model"), exact: true }).fill("mock");
    const executionPath = `/v1/user-workspace/workflow-tool-action-plans/${planId}/executions`;
    const heldExecution = await holdDraftResponse(page, executionPath, "POST");
    let runId = "";
    try {
      await execution.getByRole("button", { name: text("execute"), exact: true }).click();
      await heldExecution.arrived; await switchLanguage();
      await expect(execution.getByRole("button", { name: text("executing"), exact: true })).toBeDisabled();
      await expect(execution.getByRole("textbox", { name: text("input"), exact: true })).toHaveValue(source === "definition" ? "" : input);
      const completion = page.waitForResponse(value => isEndpoint(value, executionPath, "POST"));
      await heldExecution.deliver(); const response = await completion; const result = await response.json();
      expect(response.request().postDataJSON()).toEqual({ workspace_id: "workspace_demo", application_id: application.id, expected_record_version: 3, input_text: input, model: "mock" });
      // Product transport keeps the registered .invalid target and all SSRF restrictions.
      expect(result.failure_code, JSON.stringify(result)).toBe("workflow_tool_transport_failed");
      expect(result.action_plan.status).toBe("consumed");
      expect(result.action_plan.public_arguments.locale).toBe("ja-JP");
      expect(result.run.schema_version).toBe(`workflow_run_record.v${source === "saved" ? 2 : 9}`);
      expect(result.run.side_effects).toEqual({ provider_calls: 0, tool_calls: 1, confirmation_calls: 1, business_writes: 0, replay_writes: 0 });
      expect(JSON.stringify(result.run)).not.toContain(input);
      runId = result.run.run_id;
    } finally { await heldExecution.dispose(); }
    await expect(execution).toContainText(text("transport")); await switchLanguage();
    await expect(execution).toContainText(text("transport"));
    await expect(execution.getByLabel(text("runSafety"), { exact: true })).toBeVisible();
    await expect(action.locator(".workflow-executor-record .status-badge")).toHaveText(text("consumed"));
    expect(executions()).toHaveLength(1);
    for (const width of [1440, 720, 390]) {
      await page.setViewportSize({ width, height: 900 });
      for (const [name, panel] of [["plan", action], ["execution", execution]] as const) {
        await panel.scrollIntoViewIfNeeded();
        const clipped = await panel.locator("article, input, textarea, button, .workflow-executor-input-grid, dl > div, .prompt-template-summary").evaluateAll(elements => elements.filter(element => element.scrollWidth > element.clientWidth + 1).map(element => element.className || element.tagName));
        expect(clipped, `${source} ${name} ${width} must fit its container`).toEqual([]);
        await testInfo.attach(`http-tool-${source}-${name}-${locale}-${width}`, { body: await panel.screenshot(), contentType: "image/png" });
      }
      expect(await page.evaluate(() => document.documentElement.scrollWidth <= document.documentElement.clientWidth + 1)).toBeTruthy();
    }
    await page.setViewportSize({ width: 1440, height: 900 });
    await page.reload(); await selectApplication(page, application);
    if (source === "saved") await openDraft(page, draftId);
    else await page.getByRole("link", { name: uiText(page, "Open Workflow definition owner"), exact: true }).click();
    await expect(action.locator(".workflow-executor-record .status-badge")).toHaveText(text("consumed"));
    await expect(execution).toContainText(runId); await expect(execution).toContainText(text("transport"));
    await expect(execution.getByRole("button", { name: text("execute"), exact: true })).toBeDisabled();
    await switchLanguage(); await expect(execution).toContainText(text("transport"));
    expect(executions()).toHaveLength(1);
    const storage = await page.evaluate(() => JSON.stringify({ local: { ...localStorage }, session: { ...sessionStorage }, url: location.href }));
    expect(storage).not.toContain(input); expect(storage).not.toContain("ja-JP");
  });
}
