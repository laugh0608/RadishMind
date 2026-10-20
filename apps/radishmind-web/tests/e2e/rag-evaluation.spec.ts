import { randomUUID } from "node:crypto";
import { test, expect, isEndpoint, selectApplication } from "./workflow-fixtures";
import { promotionRequest } from "./workflow-rag-promotion-fixtures";
import { evidenceLanguage, evidenceLayout, holdEvidenceResponse } from "./results-evidence-fixtures";

const collection = "/v1/user-workspace/workflow-rag-evaluation-datasets";
test.afterEach(async ({ page }) => { if (!page.isClosed()) await page.close(); });

test("RAG datasets preserve raw JSON through bilingual pending, real CAS conflict, candidate review and archive recovery", async ({ page, application }, info) => {
  test.setTimeout(150_000);
  const ui = evidenceLanguage(page, info), copy = ui.copy;
  const suffix = randomUUID().replaceAll("-", "").slice(0, 12);
  const snapshot = await promotionRequest(page, application.id, "/v1/user-workspace/workflow-retrieval-snapshots", ["workflow_rag_snapshots:write"], {
    workspace_id: "workspace_demo", application_id: application.id, snapshot_key: `eval_${suffix}`, display_name: "Evaluation 原始快照", content_classification: "public",
    fragments: [{ fragment_ref: "evidence_original", source_type: "manual", source_ref: "manual.evaluation", page_slug: "evaluation/fixture", title: "Original title 原文", is_official: true, content: "Evaluation fixture immutable original evidence." }],
  });
  const record = snapshot.record;
  const binding = JSON.stringify({ tenantRef: record.tenant_ref, workspaceId: "workspace_demo", applicationId: application.id, snapshotId: record.snapshot_id, snapshotVersion: record.snapshot_version, snapshotDigest: record.snapshot_digest, ragRef: record.rag_ref }, null, 2);
  await page.evaluate(() => { location.hash = "workflow-rag-evaluation-panel"; });
  const panel = page.locator("#workflow-rag-evaluation-panel"); await expect(panel).toBeVisible();
  await panel.getByRole("button", { name: copy("New evaluation dataset", "新建评测数据集"), exact: true }).click();
  await panel.getByRole("textbox", { name: copy("Dataset key", "数据集键"), exact: true }).fill(`evaluation_${suffix}`);
  await panel.getByRole("textbox", { name: copy("Display name", "显示名称"), exact: true }).fill("Dataset 原始名称 <literal>");
  await panel.getByRole("combobox", { name: copy("Classification", "内容分类"), exact: true }).selectOption("synthetic_public");
  const baseline = () => panel.getByRole("textbox", { name: copy("Baseline snapshot binding", "基线快照绑定"), exact: true });
  const summary = () => panel.getByRole("textbox", { name: copy("Human review summary", "人工审查摘要"), exact: true });
  const samples = () => panel.getByRole("textbox", { name: copy("Complete samples (camelCase JSON array)", "完整样本（camelCase JSON 数组）"), exact: true });
  await baseline().fill(binding); await summary().fill("Human original review 人工原文"); await samples().fill("{");
  const beforeInvalid = ui.requests.length;
  await panel.getByRole("button", { name: copy("Create dataset v1", "创建数据集 v1"), exact: true }).click();
  await expect(panel.locator(".workflow-rag-operation")).toContainText("workflow_rag_evaluation_payload_invalid"); expect(ui.requests).toHaveLength(beforeInvalid);
  await ui.switch(); await expect(panel.locator(".workflow-rag-operation")).toContainText(copy("Check the dataset key", "请检查数据集键"));
  const sampleJSON = JSON.stringify([{ sampleId: "evaluation_original", queryText: "Evaluation fixture immutable original evidence", expectation: "evidence_required", expectedCitationRefs: ["evidence_original"], requiredOfficialRefs: ["evidence_original"], topK: 1, reviewNote: "Original review note 人工原文" }], null, 2);
  await samples().fill(sampleJSON);
  const held = await holdEvidenceResponse(page, collection, "POST");
  let datasetId = "", createPayload: Record<string, unknown> = {};
  try {
    await panel.getByRole("button", { name: copy("Create dataset v1", "创建数据集 v1"), exact: true }).click(); await held.arrived;
    await ui.switch(); await expect(samples()).toHaveValue(sampleJSON); await expect(baseline()).toHaveValue(binding);
    const pending = page.waitForResponse(response => isEndpoint(response, collection, "POST")); await held.deliver(); const response = await pending;
    expect(response.ok()).toBeTruthy(); datasetId = (await response.json()).resource.dataset_id; createPayload = response.request().postDataJSON();
    expect(JSON.stringify(createPayload)).not.toMatch(/uiLocale|Message/);
  } finally { await held.dispose(); }
  await expect(panel.locator(".workflow-rag-operation strong")).toHaveText(copy("Created", "已创建"));
  const local = "Local unsaved summary 本地未提交"; await summary().fill(local); await summary().focus(); await ui.switch();
  await expect(summary()).toHaveValue(local); await expect(summary()).toBeFocused();
  const { dataset_key: _datasetKey, ...revision } = createPayload;
  await promotionRequest(page, application.id, `${collection}/${datasetId}/versions`, ["workflow_rag_evaluation_datasets:write", "workflow_rag_snapshots:read"], { ...revision, expected_latest_version: 1 });
  await panel.getByRole("button", { name: copy("Replace all content and create v2", "完整替换并创建 v2"), exact: true }).click();
  await expect(panel.locator(".workflow-rag-operation strong")).toHaveText(copy("Version conflict", "版本冲突"));
  await expect(summary()).toHaveValue(local); await ui.switch();
  await expect(panel.locator(".workflow-rag-operation")).toContainText(copy("local draft is preserved", "本地草稿已保留"));
  await expect(panel.getByRole("button", { name: copy("Replace all content and create v2", "完整替换并创建 v2"), exact: true })).toBeDisabled();
  await panel.getByRole("button", { name: copy("Reload current evidence", "重新加载当前证据"), exact: true }).click();
  await expect(panel.getByRole("button", { name: copy("Replace all content and create v3", "完整替换并创建 v3"), exact: true })).toBeEnabled();
  await summary().fill(local);
  await panel.getByRole("button", { name: copy("Replace all content and create v3", "完整替换并创建 v3"), exact: true }).click();
  await expect(panel.locator(".workflow-rag-operation strong")).toHaveText(copy("Version saved", "版本已写入"));
  await panel.getByRole("textbox", { name: copy("Candidate snapshot binding (camelCase JSON)", "候选快照绑定（camelCase JSON）"), exact: true }).fill(binding);
  await panel.getByRole("button", { name: copy("Review candidate against dataset v3", "对数据集 v3 执行候选审查"), exact: true }).click();
  const evidence = panel.locator(".workflow-rag-review-evidence"); await expect(evidence).toContainText(copy("Unchanged", "无变化"));
  await expect(evidence).toContainText(copy("Passed", "通过")); await ui.switch();
  await expect(evidence).toContainText(copy("Mean reciprocal rank", "平均倒数排名"));
  // A failed history read is a failure, not an empty history or an unrecorded review.
  await page.route(`http://127.0.0.1:17000${collection}/${datasetId}/candidate-reviews?*`, async route => {
    const response = await route.fetch(); const body = await response.json();
    body.items = []; body.next_cursor = null; body.failure_code = "workflow_rag_evaluation_store_unavailable";
    await route.fulfill({ response, status: 503, json: body });
  }, { times: 1 });
  await panel.locator(".workflow-rag-list").getByRole("button").filter({ hasText: `evaluation_${suffix}` }).click();
  await expect(panel.locator(".workflow-rag-review-history [role=alert]")).toContainText("workflow_rag_evaluation_store_unavailable");
  await ui.switch();
  await expect(panel.locator(".workflow-rag-review-history [role=alert]")).toContainText(copy("evaluation store is unavailable", "评测存储不可用"));
  await expect(panel.locator(".workflow-rag-review-history")).not.toContainText(copy("No candidate snapshot reviews yet", "尚无候选快照审查记录"));
  await panel.getByRole("button", { name: copy("Reload current evidence", "重新加载当前证据"), exact: true }).click();
  await expect(panel.locator(".workflow-rag-review-history [role=alert]")).toHaveCount(0);
  await panel.locator(".workflow-rag-review-history").getByRole("button").first().click();
  await expect(evidence).toContainText(copy("Unchanged", "无变化"));
  await evidenceLayout(page, panel, info, "rag-evaluation");
  await panel.getByRole("button", { name: copy("Archive dataset", "归档数据集"), exact: true }).click(); await ui.switch();
  await expect(panel.locator(".workflow-rag-archive-confirm")).toContainText(copy("Historical versions and reviews remain readable", "历史版本和审查仍可读"));
  await panel.getByRole("button", { name: copy("Confirm archive", "确认归档"), exact: true }).click();
  await expect(panel.locator(".workflow-rag-operation strong")).toHaveText(copy("Archived", "已归档"));
  await page.reload(); await selectApplication(page, application); await page.evaluate(() => { location.hash = "workflow-rag-evaluation-panel"; });
  await panel.locator(".workflow-rag-filter").getByRole("button", { name: copy("Archived", "已归档"), exact: true }).click();
  await panel.locator(".workflow-rag-list").getByRole("button").filter({ hasText: `evaluation_${suffix}` }).click();
  await expect(summary()).toHaveValue(local); await expect(samples()).toHaveValue(sampleJSON);
  await panel.locator(".workflow-rag-review-history").getByRole("button").first().click(); await expect(evidence).toContainText(copy("Unchanged", "无变化"));
  expect(ui.requests.filter(request => request.startsWith("POST ") && /executions|\/runs|\/v1\/chat|\/v1\/responses/.test(request))).toEqual([]);
});
