import { randomUUID } from "node:crypto";
import { test, expect, evaluation, planOwner, button, evaluationLanguage, openEvaluation, createPlan, createStructuredPlan, evaluationSubmit, evaluationLayout } from "./evaluation-fixtures";
import { holdDraftResponse, selectApplication } from "./workflow-fixtures";
import { prepareApplication } from "./prompt-fixtures";

test.use({ applicationKind: "prompt_application" });
test.afterEach(async ({ page }) => { if (!page.isClosed()) await page.close(); });

test("Prompt plan preserves invalid JSON and negative fixtures and requires explicit draft discard", async ({ page, application }, info) => {
  const ui = evaluationLanguage(page, info), copy = ui.copy;
  await openEvaluation(page, "plan");
  const panel = planOwner(page);
  await button(panel, copy("Read current template", "读取当前模板")).click();
  await expect(panel.locator(".prompt-plan-reference")).toHaveAttribute("data-reference-state", "unavailable");
  await panel.getByRole("textbox", { name: copy("Plan name", "计划名称"), exact: true }).fill("Negative fixture plan");
  await button(panel, copy("Advanced JSON", "高级 JSON")).click();
  const input = () => panel.getByRole("textbox", { name: copy("Ordered samples · advanced JSON", "有序样本 · 高级 JSON"), exact: true });
  const original = JSON.stringify([{
    item_key: "negative_fixture", name: "Null fixture", expected_classification: "inconclusive",
    workflow_definition: null, application_rag: null, agent_copilot: null,
    prompt_application: { variables: { question: null, old_context: "Retained 原文", count: 2 } },
  }], null, 2).replace('"count": 2', '"count": 9007199254740993');
  await input().fill(original);
  const saves = ui.requests.filter(request => request.startsWith("POST ")).length;
  await button(panel, copy("Variable form", "变量表单")).click();
  await expect(panel).toHaveAttribute("data-edit-mode", "json");
  await expect(input()).toHaveValue(original);
  await button(panel, copy("Review save", "审查保存")).click();
  await expect(panel.locator(".application-evaluation-confirmation")).toHaveCount(0);
  expect(ui.requests.filter(request => request.startsWith("POST ")).length).toBe(saves);
  await ui.switch();
  await expect(input()).toHaveValue(original);
  await expect(panel.locator(".prompt-plan-errors")).toContainText(copy("without loss", "无损"));
  await evaluationLayout(page, evaluation(page), info, "prompt-json-number-error");
  await input().fill(original.replace("9007199254740993", "2"));
  await button(panel, copy("Variable form", "变量表单")).click();
  await panel.getByRole("textbox", { name: copy("Sample name", "样本名称"), exact: true }).fill("Retained null fixture");
  await button(panel, copy("Advanced JSON", "高级 JSON")).click();
  expect(JSON.parse(await input().inputValue())[0].prompt_application.variables).toEqual({ question: null, old_context: "Retained 原文", count: 2 });
  await button(panel, copy("Review save", "审查保存")).click();
  await expect(panel.locator(".application-evaluation-confirmation")).toContainText(copy("Template matching incomplete or mismatched", "未完成模板匹配或存在不匹配"));
  const saved = await evaluationSubmit(page, button(panel, copy("Confirm Create", "确认创建")), `/v1/user-workspace/applications/${application.id}/evaluation-plans`);
  expect(saved.body.failure_code).toBeNull();
  expect(saved.request.postDataJSON().items[0].prompt_application.variables.question).toBeNull();
  await panel.getByRole("textbox", { name: copy("Plan name", "计划名称"), exact: true }).fill("Unsaved new title");
  const profile = panel.getByRole("combobox", { name: copy("Execution profile", "执行档案"), exact: true });
  await profile.selectOption("workflow_definition_executor_v1");
  const discard = evaluation(page).getByRole("alertdialog");
  await expect(discard).toBeVisible();
  await button(discard, copy("Keep editing", "继续编辑")).click();
  await expect(profile).toHaveValue("prompt_application_invocation_v1");
  await expect(input()).toHaveValue(/Retained/);
  await button(panel, copy("New plan", "新建计划")).click();
  await button(panel, copy("Discard and continue", "放弃并继续")).click();
  await expect(panel.getByRole("textbox", { name: copy("Plan name", "计划名称"), exact: true })).toHaveValue("");
  await expect(panel.locator(".prompt-plan-sample-list [role=option]")).toHaveCount(0);
  await button(panel, copy("Load exact version", "加载精确版本")).click();
  await expect(panel.getByRole("textbox", { name: copy("Plan name", "计划名称"), exact: true })).toHaveValue("Negative fixture plan");
});

test("evaluation plan keeps JSON, confirmation, pending revision and real CAS conflicts across languages", async ({ page, application }, info) => {
  const ui = evaluationLanguage(page, info), copy = ui.copy;
  await openEvaluation(page, "plan");
  const panel = planOwner(page);
  await button(panel, copy("Advanced JSON", "高级 JSON")).click();
  const items = () => panel.getByRole("textbox", { name: copy("Ordered samples · advanced JSON", "有序样本 · 高级 JSON"), exact: true });
  await items().fill("{");
  await button(panel, copy("Review save", "审查保存")).click();
  await expect(evaluation(page)).toContainText(copy("valid JSON", "有效 JSON"));
  await ui.switch();
  await expect(items()).toHaveValue("{");
  await expect(evaluation(page)).toContainText(copy("valid JSON", "有效 JSON"));
  const created = await createPlan(page, application.id, ui);
  const id = created.body.plan.plan_id;
  const path = `/v1/user-workspace/applications/${application.id}/evaluation-plans/${id}/revisions`;
  const originalJSON = await items().inputValue();
  const name = () => panel.getByRole("textbox", { name: copy("Plan name", "计划名称"), exact: true });
  await name().fill("Revision 原文");
  await button(panel, copy("Review save", "审查保存")).click();
  await ui.switch();
  await expect(panel.locator(".application-evaluation-confirmation")).toContainText("Revision 原文");
  await expect(items()).toHaveValue(originalJSON);
  const held = await holdDraftResponse(page, path, "POST");
  try {
    const pending = evaluationSubmit(page, button(panel, copy("Confirm Revise", "确认修订")), path);
    await held.arrived;
    await ui.switch();
    await expect(evaluation(page).getByRole("status")).toContainText(copy("Request in progress", "请求处理中"));
    await expect(button(panel, copy("Confirm Revise", "确认修订"))).toBeDisabled();
    await held.deliver();
    expect((await pending).body.version.plan_version).toBe(2);
  } finally { await held.dispose(); }
  await expect(evaluation(page).locator(".application-evaluation-operation")).toContainText("v2");
  const headers = await created.request.allHeaders();
  const external = await page.request.post(`http://127.0.0.1:17000${path}`, { headers, data: { ...created.request.postDataJSON(), expected_version: 2, name: "Concurrent v3" } });
  expect((await external.json()).version.plan_version).toBe(3);
  await name().fill("Keep conflict 编辑");
  await button(panel, copy("Review save", "审查保存")).click();
  const conflict = await evaluationSubmit(page, button(panel, copy("Confirm Revise", "确认修订")), path);
  expect(conflict.body.failure_code).toBe("application_evaluation_version_conflict");
  await ui.switch();
  await expect(evaluation(page)).toContainText(copy("Local input is retained", "本地输入仍保留"));
  await expect(name()).toHaveValue("Keep conflict 编辑");
  await expect(items()).toHaveValue(originalJSON);
  await expect(button(panel, copy("Review save", "审查保存"))).toBeDisabled();
  await evaluationLayout(page, evaluation(page), info, "plan-conflict");
  await button(evaluation(page), copy("Read remote version", "重读远端版本")).click();
  await expect(panel.locator(".prompt-plan-selected-context")).toContainText("v3");
  await expect(name()).toHaveValue("Keep conflict 编辑");
  await button(panel, copy("Load exact version", "加载精确版本")).click();
  await expect(name()).toHaveValue("Keep conflict 编辑");
  await button(panel, copy("Discard and continue", "放弃并继续")).click();
  await expect(name()).toHaveValue("Concurrent v3");
  await button(panel, copy("Review archive", "审查归档")).click();
  await ui.switch();
  await expect(panel.locator(".application-evaluation-confirmation")).toContainText(copy("existing versions and campaigns remain readable", "既有批次仍可读取"));
  const archived = await evaluationSubmit(page, button(panel, copy("Confirm Archive", "确认归档")), path.replace("/revisions", "/archive"));
  expect(archived.body.plan.lifecycle_state).toBe("archived");
  await page.reload(); await selectApplication(page, application); await openEvaluation(page, "plan");
  await expect(evaluation(page)).toContainText(copy("No active evaluation plans", "没有活动评测计划"));
});

test("evaluation campaign preserves execution inputs, waiting, pair handoff and durable refresh in both languages", async ({ page, application }, info) => {
  test.setTimeout(150_000);
  const keyIdPromise = page.waitForResponse(response => /\/api-keys$/.test(new URL(response.url()).pathname) && response.request().method() === "POST")
    .then(async response => (await response.json()).record.api_key_id as string);
  await prepareApplication(page, application);
  const keyId = await keyIdPromise;
  // Only the public API Key ID is consumed by Campaign. Credentials never enter artifacts.
  const ui = evaluationLanguage(page, info), copy = ui.copy;
  const caseId = randomUUID();
  const created = await createStructuredPlan(page, application.id, ui, caseId);
  await page.reload(); await selectApplication(page, application); await openEvaluation(page, "plan");
  const restoredPlan = planOwner(page);
  await button(restoredPlan, copy("Load exact version", "加载精确版本")).click();
  await button(restoredPlan, copy("Read current template", "读取当前模板")).click();
  await expect(restoredPlan.locator(".prompt-plan-reference")).toHaveAttribute("data-reference-state", "ready");
  await expect(restoredPlan.getByRole("textbox", { name: copy("Sample key", "样本标识"), exact: true })).toHaveValue("typed_sample");
  await expect(restoredPlan.getByRole("textbox", { name: "question", exact: true })).toHaveValue(`E2E_CASE:${caseId}:valid`);
  await evaluationLayout(page, evaluation(page), info, "structured-plan-restored");
  const headers = { ...await created.request.allHeaders(),
    "x-radishmind-dev-read-scopes": "admin_gateway_quotas:write",
    "x-radishmind-dev-read-membership-permissions": "admin_gateway_quotas:write",
    "x-radishmind-dev-gateway-quota-environment": "test",
  };
  const quota = await page.request.put(`http://127.0.0.1:17000/v1/admin/gateway-request-quotas/${application.id}`, { headers, data: { expected_version: 0, request_limit: 20 } });
  expect((await quota.json()).failure_code).toBeNull();
  await openEvaluation(page, "campaign");
  const panel = page.locator(".application-evaluation-campaign-owner");
  const ids: string[] = [];
  for (let round = 0; round < 2; round++) {
    if (round) await button(panel, copy("New campaign", "新建批次")).click();
    await panel.getByRole("textbox", { name: copy("Active API Key ID · quota consumer only", "活动 API Key ID · 仅作配额主体"), exact: true }).fill(keyId);
    const input = () => panel.getByRole("textbox", { name: copy("Client campaign key · idempotent in this scope", "客户端批次键 · 当前作用域内幂等"), exact: true });
    const campaignKey = `evaluation-test-${randomUUID()}`;
    await input().fill(campaignKey);
    await button(panel, copy("Review execution", "审查执行")).click();
    await ui.switch();
    await expect(input()).toHaveValue(campaignKey);
    await expect(panel.locator(".application-evaluation-confirmation")).toContainText(keyId);
    const path = `/v1/user-workspace/applications/${application.id}/evaluation-campaigns`;
    const held = await holdDraftResponse(page, path, "POST");
    try {
      const running = evaluationSubmit(page, button(panel, copy("Confirm campaign", "确认批次")), path);
      await held.arrived; await ui.switch();
      await expect(button(panel, copy("Confirm campaign", "确认批次"))).toBeDisabled();
      await held.deliver();
      const result = await running;
      expect(result.request.postDataJSON().client_campaign_key).toBe(campaignKey);
      expect(result.request.postDataJSON().quota_api_key_id).toBe(keyId);
      expect(result.body.failure_code).toBeNull();
      expect(result.body.campaign.state).toBe("succeeded");
      ids.push(result.body.campaign.campaign_id);
    } finally { await held.dispose(); }
    await expect(panel.locator(".application-evaluation-item")).toHaveCount(2);
  }
  await expect(panel.locator(".application-evaluation-campaign-context small")).toHaveCSS("text-transform", "none");
  await evaluationLayout(page, evaluation(page), info, "campaign-terminal");
  await button(panel, copy("Review handoff", "审查交接")).click();
  const pair = page.locator(".application-evaluation-pair-owner");
  await button(pair, copy("Preview exact pair", "预览精确配对")).click();
  await expect(pair.locator(".application-evaluation-pair-items article")).toHaveCount(2);
  await ui.switch();
  await expect(pair.getByRole("combobox", { name: copy("Baseline campaign", "基线批次"), exact: true })).toHaveValue(ids[0]);
  await expect(pair.getByRole("combobox", { name: copy("Candidate campaign", "候选批次"), exact: true })).toHaveValue(ids[1]);
  await button(pair, copy("Review handoff", "审查交接")).click();
  await ui.switch();
  await expect(pair.locator(".application-evaluation-confirmation")).toContainText(copy("not deleted or automatically retried", "不删除或自动重试"));
  await evaluationLayout(page, evaluation(page), info, "handoff-confirmation");
  const response = page.waitForResponse(r => /\/evaluation-campaign-pairs\/handoff$/.test(new URL(r.url()).pathname) && r.request().method() === "POST");
  await button(pair, copy("Confirm handoff", "确认交接")).click();
  const handoff = await (await response).json();
  expect(handoff.failure_code).toBeNull();
  expect(handoff.handoff.state).toBe("complete");
  await expect(pair.locator(".application-evaluation-handoff-evidence")).toContainText(handoff.handoff.suite_id);
  await page.reload(); await selectApplication(page, application); await openEvaluation(page, "handoff");
  await button(pair, copy("Preview exact pair", "预览精确配对")).click();
  await expect(pair.locator(".application-evaluation-handoff-evidence")).toContainText(handoff.handoff.suite_id);
  const provider = process.env.RADISHMIND_E2E_PROVIDER_URL!;
  const observed = await (await page.request.get(`${provider}/observations/${caseId}`)).json();
  expect(observed).toHaveLength(4);
  expect(observed.every((item: { accepted: boolean }) => item.accepted)).toBe(true);
});
