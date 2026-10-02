import { test, expect, holdDraftResponse, isEndpoint } from "./workflow-fixtures";
import { setTestLanguage } from "./ui-language";
import { preparePromotionEvidence, promotionPath, promotionRequest } from "./workflow-rag-promotion-fixtures";

const labels = {
  load: ["Load sources", "加载来源"], dataset: ["Active dataset", "可用数据集"], review: ["Eligible candidate review", "合格候选审查"],
  draft: ["Valid source draft", "有效来源草案"], create: ["Create promotion candidate", "创建晋级候选"],
  decision: ["Decision", "人工决定"], reason: ["Sanitized reason", "脱敏理由"], submit: ["Record decision", "记录决定"],
  refresh: ["Refresh candidates", "刷新候选"], refreshRecord: ["Refresh current record v2", "刷新当前记录 v2"],
  conflict: ["The record changed. Your reason is retained; refresh the record before deciding again.", "记录已变更，当前理由已保留；请刷新记录后重新决定。"],
  pending: ["Recording the human decision…", "正在记录人工决定…"], attach: ["Open configuration draft attach step", "打开配置草案附加步骤"],
  invalidReason: ["Enter a reason of 4–500 characters.", "请输入 4–500 个字符的理由。"], canceled: ["Canceled", "已取消"],
} as const;

test("promotion retains evidence and input through locale changes, CAS conflict, approval, cancellation and reload", async ({ page, application }, testInfo) => {
  const evidence = await preparePromotionEvidence(page, application);
  let locale: "en-US" | "zh-CN" = testInfo.project.use.locale === "zh-CN" ? "zh-CN" : "en-US";
  const text = (key: keyof typeof labels) => labels[key][locale === "zh-CN" ? 1 : 0];
  const panel = page.locator("#workflow-rag-promotion-review");
  const button = (key: keyof typeof labels) => panel.getByRole("button", { name: text(key), exact: true });
  const select = (key: keyof typeof labels) => panel.getByRole("combobox", { name: text(key), exact: true });
  const reasonBox = () => panel.getByRole("textbox", { name: text("reason"), exact: true });
  const requests: string[] = [];
  page.on("request", request => { if (new URL(request.url()).port === "17000") requests.push(`${request.method()} ${new URL(request.url()).pathname}`); });
  async function switchLanguage() {
    const count = requests.length; locale = locale === "zh-CN" ? "en-US" : "zh-CN";
    await page.evaluate(value => { localStorage.setItem("radishmind.uiLocale.v1", value); window.dispatchEvent(new StorageEvent("storage", { key: "radishmind.uiLocale.v1", newValue: value, storageArea: localStorage })); }, locale);
    setTestLanguage(page, locale); await expect(page.locator("html")).toHaveAttribute("lang", locale);
    await page.evaluate(() => new Promise<void>(resolve => requestAnimationFrame(() => requestAnimationFrame(() => resolve()))));
    expect(requests, "Language changes must not fetch, decide, attach or execute").toHaveLength(count);
  }
  async function openStage() {
    await page.locator('a[href="#application-publish-review"]:visible').first().click();
    const review = page.locator('button[aria-controls="application-development-owner-surface"]');
    if (await review.getAttribute("aria-expanded") === "false") await review.click();
    await expect(panel).toBeVisible();
  }
  await openStage(); await button("load").click(); await expect(button("create")).toBeEnabled();
  const heldCreate = await holdDraftResponse(page, promotionPath, "POST"); let candidateId = "";
  try {
    await button("create").click(); await heldCreate.arrived; await switchLanguage();
    await expect(button("create")).toBeDisabled();
    await expect(select("dataset")).toHaveValue(evidence.datasetId); await expect(select("review")).toHaveValue(evidence.reviewId); await expect(select("draft")).toHaveValue(evidence.draftId);
    const completion = page.waitForResponse(response => isEndpoint(response, promotionPath, "POST")); await heldCreate.deliver();
    const response = await completion, body = await response.json(); expect(body.failure_code, JSON.stringify(body)).toBeNull(); candidateId = body.candidate.candidate_id;
    expect(response.request().postDataJSON()).toEqual({ ...evidence.scope, dataset_id: evidence.datasetId, dataset_version: 1, dataset_digest: evidence.digest, candidate_review_id: evidence.reviewId, draft_id: evidence.draftId, expected_draft_version: 1 }); expect(body.binding).toBeNull();
  } finally { await heldCreate.dispose(); }
  await expect(panel.locator(".workflow-rag-promotion-list")).toContainText(candidateId);
  await reasonBox().fill("短"); await expect(panel).toContainText(text("invalidReason")); await switchLanguage();
  await expect(panel).toContainText(text("invalidReason")); await expect(button("submit")).toBeDisabled();
  const reason = "Reviewed exact evidence 原始人工理由 <literal>"; await reasonBox().fill(reason);
  const decisionPath = `${promotionPath}/${candidateId}/decisions`;
  await promotionRequest(page, application.id, decisionPath, ["workflow_rag_promotions:review"], { ...evidence.scope, expected_record_version: 1, decision: "defer", reason: "Another reviewer requests a later decision." });
  const conflict = page.waitForResponse(response => isEndpoint(response, decisionPath, "POST")); await button("submit").click();
  expect((await (await conflict).json()).failure_code).toBe("workflow_rag_promotion_record_version_conflict");
  await expect(panel).toContainText(text("conflict")); await switchLanguage(); await expect(panel).toContainText(text("conflict"));
  await expect(reasonBox()).toHaveValue(reason); await expect(select("decision")).toHaveValue("approve");
  await button("refreshRecord").click(); await expect(panel.locator(".workflow-rag-decision-record")).toHaveCount(1); await expect(reasonBox()).toHaveValue(reason);
  const heldDecision = await holdDraftResponse(page, decisionPath, "POST"); let bindingId = "", bindingDigest = "";
  try {
    await button("submit").click(); await heldDecision.arrived; await switchLanguage();
    await expect(panel).toContainText(text("pending")); await expect(button("submit")).toBeDisabled(); await expect(reasonBox()).toHaveValue(reason);
    const completion = page.waitForResponse(response => isEndpoint(response, decisionPath, "POST")); await heldDecision.deliver();
    const response = await completion, body = await response.json(); expect(body.failure_code, JSON.stringify(body)).toBeNull();
    expect(response.request().postDataJSON()).toEqual({ ...evidence.scope, expected_record_version: 2, decision: "approve", reason });
    expect(body.candidate.record_version).toBe(3); expect(body.eligibility.eligible).toBe(true); bindingId = body.binding.binding_id; bindingDigest = body.binding.binding_digest;
  } finally { await heldDecision.dispose(); }
  await expect(panel.locator(".workflow-rag-promotion-list")).toContainText(candidateId);
  await expect(panel.locator(".workflow-rag-decision-record").last()).toContainText(reason); await switchLanguage();
  await expect(panel.locator(".workflow-rag-binding-card")).toContainText(bindingId); await expect(panel.locator(".workflow-rag-binding-card")).toContainText(bindingDigest);
  await expect(panel.locator(".workflow-rag-promotion-detail")).toContainText(evidence.digest); await expect(panel.locator(".workflow-rag-decision-record").last()).toContainText(reason); await expect(button("attach")).toBeEnabled();
  for (const width of [1440, 720, 390]) {
    await page.setViewportSize({ width, height: 900 }); await panel.scrollIntoViewIfNeeded();
    expect(await panel.locator("article, .workflow-rag-evidence-row, .workflow-rag-binding-card, .workflow-rag-promotion-list").evaluateAll(elements => elements.filter(element => element.scrollWidth > element.clientWidth + 1).map(element => element.className))).toEqual([]);
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= document.documentElement.clientWidth + 1)).toBeTruthy();
    await testInfo.attach(`rag-promotion-${locale}-${width}`, { body: await panel.screenshot(), contentType: "image/png" });
  }
  await page.setViewportSize({ width: 1440, height: 900 });
  expect(requests.filter(value => value === `POST ${promotionPath}`)).toHaveLength(1); expect(requests.filter(value => value === `POST ${decisionPath}`)).toHaveLength(2);
  expect(requests.filter(value => /POST .*\/application-drafts|POST .*executions|POST .*publish|POST .*releases/.test(value))).toEqual([]);
  const writes = requests.filter(value => value.startsWith("POST ")).length; await button("attach").click();
  await expect(page.locator("#application-configuration-draft")).toContainText(bindingId); expect(requests.filter(value => value.startsWith("POST "))).toHaveLength(writes);
  await openStage(); await button("refresh").click(); await panel.locator(".workflow-rag-promotion-list button").filter({ hasText: candidateId }).click();
  await expect(select("decision")).toHaveValue("cancel"); await reasonBox().fill("Cancel binding eligibility after review.");
  const canceled = page.waitForResponse(response => isEndpoint(response, decisionPath, "POST")); await button("submit").click(); const canceledBody = await (await canceled).json();
  expect(canceledBody.candidate.candidate_state).toBe("canceled"); expect(canceledBody.eligibility.eligible).toBe(false); expect(canceledBody.binding.binding_id).toBe(bindingId);
  await expect(button("attach")).toBeDisabled(); await expect(panel.locator(".workflow-rag-promotion-list")).toContainText(text("canceled")); await switchLanguage();
  await expect(panel.locator(".section-heading > .status-badge")).toHaveText(text("canceled"));
  await page.reload(); await openStage(); await button("refresh").click(); await panel.locator(".workflow-rag-promotion-list button").filter({ hasText: candidateId }).click();
  await expect(panel.locator(".workflow-rag-decision-record")).toHaveCount(3); await expect(panel.locator(".workflow-rag-binding-card")).toContainText(bindingId); await expect(button("attach")).toBeDisabled();
});
