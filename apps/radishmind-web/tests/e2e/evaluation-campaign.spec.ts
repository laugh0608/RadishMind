import { randomUUID } from "node:crypto";
import { test, expect, evaluation, planOwner, button, evaluationLanguage, openEvaluation, createPlan, evaluationSubmit, evaluationLayout } from "./evaluation-fixtures";
import { holdDraftResponse, selectApplication } from "./workflow-fixtures";
import { prepareApplication } from "./prompt-fixtures";

test.use({ applicationKind: "prompt_application" });
test.afterEach(async ({ page }) => { if (!page.isClosed()) await page.close(); });

test("evaluation plan keeps JSON, confirmation, pending revision and real CAS conflicts across languages", async ({ page, application }, info) => {
  const ui = evaluationLanguage(page, info), copy = ui.copy;
  await openEvaluation(page, "plan");
  const panel = planOwner(page);
  const items = () => panel.getByRole("textbox", { name: copy("Ordered fixtures · 1–20 · strict JSON", "有序夹具 · 1–20 项 · 严格 JSON"), exact: true });
  await items().fill("{");
  await button(panel, copy("Review create", "审查创建")).click();
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
  await button(panel, copy("Review revision", "审查修订")).click();
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
  await button(panel, copy("Review revision", "审查修订")).click();
  const conflict = await evaluationSubmit(page, button(panel, copy("Confirm Revise", "确认修订")), path);
  expect(conflict.body.failure_code).toBe("application_evaluation_version_conflict");
  await ui.switch();
  await expect(evaluation(page)).toContainText(copy("local input is retained", "本地输入仍保留"));
  await expect(name()).toHaveValue("Keep conflict 编辑");
  await expect(items()).toHaveValue(originalJSON);
  await expect(button(panel, copy("Review revision", "审查修订"))).toBeDisabled();
  await evaluationLayout(page, evaluation(page), info, "plan-conflict");
  await button(evaluation(page), copy("Reload exact owner", "重新加载精确记录")).click();
  await expect(panel.locator(".status-badge")).toContainText("v3");
  await expect(name()).toHaveValue("Keep conflict 编辑");
  await button(panel, copy("Load exact version", "加载精确版本")).click();
  await expect(name()).toHaveValue("Concurrent v3");
  await button(panel, copy("Review archive", "审查归档")).click();
  await ui.switch();
  await expect(panel.locator(".application-evaluation-confirmation")).toContainText(copy("existing campaigns remain readable", "既有批次仍可读取"));
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
  const created = await createPlan(page, application.id, ui, caseId);
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
    await expect(panel.locator(".application-evaluation-item")).toHaveCount(1);
  }
  await expect(panel.locator(".application-evaluation-campaign-context small")).toHaveCSS("text-transform", "none");
  await evaluationLayout(page, evaluation(page), info, "campaign-terminal");
  await button(panel, copy("Review handoff", "审查交接")).click();
  const pair = page.locator(".application-evaluation-pair-owner");
  await button(pair, copy("Preview exact pair", "预览精确配对")).click();
  await expect(pair.locator(".application-evaluation-pair-items article")).toHaveCount(1);
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
  expect(observed).toHaveLength(2);
  expect(observed.every((item: { accepted: boolean }) => item.accepted)).toBe(true);
});
