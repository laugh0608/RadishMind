import assert from "node:assert/strict";
import test from "node:test";
import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { createUiI18n, initializeUiI18n } from "../src/i18n/instance.ts";
import { workflowInspectionProjection as enInspection } from "../src/i18n/locales/en-US/workflowInspectionProjection.ts";
import { workflowHandoffProjection as enHandoff } from "../src/i18n/locales/en-US/workflowHandoffProjection.ts";
import { workflowInspectionProjection as zhInspection } from "../src/i18n/locales/zh-CN/workflowInspectionProjection.ts";
import { workflowHandoffProjection as zhHandoff } from "../src/i18n/locales/zh-CN/workflowHandoffProjection.ts";
import { workflowDraft as enDraft } from "../src/i18n/locales/en-US/workflowDraft.ts";
import { workflowDraft as zhDraft } from "../src/i18n/locales/zh-CN/workflowDraft.ts";
import { workflowProjectionMessage, workflowProjectionText, workflowProjectionStatusLabel, type WorkflowProjectionCopy } from "../src/features/control-plane-read/workflowProjectionCopy.ts";
import { buildWorkflowSavedDraftConflictReviewSummary, initialWorkflowSavedDraftConsumerState, saveWorkflowDraftDevRecord, type WorkflowSavedDraftSummary } from "../src/features/control-plane-read/savedWorkflowDraftConsumer.ts";
import { buildWorkflowExecutorV0Draft } from "../src/features/control-plane-read/workflowExecutorConsumer.ts";
import type { WorkflowDraftDesignerDraft } from "../src/features/control-plane-read/workflowDraftDesigner.ts";

const en = { ...enInspection, ...enHandoff };
const zh = { ...zhInspection, ...zhHandoff };
const config = { mode: "dev_saved_draft_http" as const, baseUrl: "http://platform.test", workspaceId: "workspace_demo", tenantRef: "tenant_demo", subjectRef: "actor_demo" };
async function languageInstance() {
  const instance = createUiI18n(); await initializeUiI18n(instance, "en-US");
  instance.addResourceBundle("en-US", "workflow", { projection: en, draft: enDraft });
  instance.addResourceBundle("zh-CN", "workflow", { projection: zh, draft: zhDraft });
  return instance;
}
function draft(): WorkflowDraftDesignerDraft {
  return buildWorkflowExecutorV0Draft({
    draftId: "draft_original", templateRef: "wf_original", label: "用户 draft <script>", applicationRef: "app_demo",
    workflowDefinitionId: "wf_original", providerProfileRef: "provider:mock", summary: "Original description 原文",
    nodes: [], edges: [], readiness: [], risks: [], blockedCapabilities: [],
    designerLayout: { source: "workflow_node_designer", persistence: "ui_only", nodePositions: [] },
    routeMetadata: { sourceRouteId: "workflow-definition-summary-list-route", draftRouteId: "workflow-draft-designer-offline-draft", routePath: "/v1/user-workspace/workflow-definitions", requestId: "req_original", auditRef: "audit_original" },
    localOnlyInteraction: "inspect_only", executionProfile: "review_only",
  }, 1);
}

test("projection copy switches nested text and states while retaining unannotated source prose", async () => {
  const instance = await languageInstance();
  const source: WorkflowProjectionCopy & { summary: string } = { summary: "Original diagnostic", summaryMessage: { parts: [
    { key: "graphContractSummary", values: { value1: { key: "inputExecutorContract" }, value2: "tenant_ref <script>" } },
    { status: "review_required" }, { key: "targetEdges", values: { edges: 2, nodes: 3 } },
  ], separator: " / " } };
  const before = structuredClone(source);
  assert.match(workflowProjectionText(instance.getFixedT("en-US", "workflow"), source, "summary"), /Missing fields: tenant_ref <script>/);
  await instance.changeLanguage("zh-CN"); const t = instance.getFixedT("zh-CN", "workflow");
  const chinese = workflowProjectionText(t, source, "summary");
  assert.match(chinese, /缺少字段：tenant_ref <script>/);
  assert.match(chinese, /2 条相关连线 \/ 3 个相关节点/);
  assert.doesNotMatch(chinese, /Missing fields|review_required/);
  assert.deepEqual(source, before);
  assert.equal(workflowProjectionText(t, { summary: en.input_contract_fields_label }, "summary"), en.input_contract_fields_label);
  assert.equal(workflowProjectionText(t, { summary: null }, "summary"), "");
  assert.equal(workflowProjectionStatusLabel(t, "future_protocol_state"), "future_protocol_state");
  const markup = renderToStaticMarkup(createElement("p", null, chinese));
  assert.ok(markup.includes("&lt;script&gt;")); assert.ok(!markup.includes("<script>"));
  for (const key of Object.keys(en).filter(key => key.startsWith("status_"))) assert.equal(workflowProjectionStatusLabel(t, key.slice(7)), zh[key as keyof typeof zh]);
  assert.equal(workflowProjectionMessage(t, { key: "targetNodes", values: { nodes: 0 } }), "0 个相关节点");
});

test("conflict copy covers loaded, missing, failed, pending and continued states without changing recovery", async () => {
  const instance = await languageInstance(), local = draft();
  const state = { ...initialWorkflowSavedDraftConsumerState(config), status: "version_conflict" as const, conflictDraftVersion: 7, currentDraftVersion: 6 };
  const before = structuredClone({ local, state });
  const saved: WorkflowSavedDraftSummary = {
    draftId: local.draftId, workspaceId: config.workspaceId, applicationRef: local.applicationRef, workflowDefinitionId: local.workflowDefinitionId,
    draftVersion: 7, lifecycleState: "active", lifecycleVersion: 1, archivedAt: null, libraryUpdatedAt: "2026-10-02T12:00:00Z", lifecycleUpdatedByActorRef: "actor_remote",
    provenanceKind: "workflow_definition", draftStatus: "draft", name: "Remote 原文", description: "Remote description", updatedAt: "2026-10-02T12:00:00Z", updatedByActorRef: "actor_remote", nodeCount: 3, edgeCount: 2, blockedCapabilityCount: 0, validationState: "valid_for_review", validForReview: true, sampleOrUnsavedDraftStatus: "saved_dev_record",
  };
  for (const status of ["sample", "empty", "list_failed", "open_failed", "loading", "ready", undefined] as const) {
    const review = buildWorkflowSavedDraftConflictReviewSummary(state, local, [], status, "original_failure_99")!;
    assert.equal(review.canOpenSavedDraft, false); assert.equal(review.canAutoOverwriteLocalDraft, false); assert.equal(review.canAutoMergeDraft, false);
    const snapshot = structuredClone(review);
    for (const locale of ["en-US", "zh-CN"] as const) {
      await instance.changeLanguage(locale); const t = instance.getFixedT(locale, "workflow");
      assert.ok(workflowProjectionText(t, review, "summary").includes(local.draftId));
      assert.ok(workflowProjectionText(t, review, "summary").includes("7"));
      const reason = workflowProjectionText(t, review, "openUnavailableReason"); assert.ok(reason.length > 0);
      if (status === "list_failed" || status === "open_failed") assert.ok(reason.includes("original_failure_99"));
      for (const field of ["localDraftPreservationSummary", "nextReviewerStep", "reviewerQuestion"] as const) {
        const text = workflowProjectionText(t, review, field); assert.ok(text.length > 0 && !text.includes("{{"));
        if (locale === "zh-CN") assert.match(text, /[\u3400-\u9fff]/);
      }
    }
    assert.deepEqual(review, snapshot);
  }
  for (const status of ["version_conflict", "conflict_local_continued"] as const) {
    const review = buildWorkflowSavedDraftConflictReviewSummary({ ...state, status }, local, [saved], "ready")!;
    assert.equal(review.canOpenSavedDraft, true); const t = instance.getFixedT("zh-CN", "workflow");
    assert.match(workflowProjectionText(t, review, "summary"), /actor_remote.*2026-10-02T12:00:00Z/);
    assert.equal(workflowProjectionText(t, review, "openUnavailableReason"), "");
    assert.match(workflowProjectionText(t, review, "nextReviewerStep"), /本地|保存/);
    assert.match(workflowProjectionText(t, review, "localDraftPreservationSummary"), /本地草案/);
  }
  assert.equal(buildWorkflowSavedDraftConflictReviewSummary({ ...state, status: "saved_dev_record" }, local, [saved]), null);
  assert.deepEqual({ local, state }, before);
});

test("display metadata and language changes do not enter saved draft requests", async context => {
  const instance = await languageInstance(), local = draft(), requests: unknown[] = [];
  const originalFetch = globalThis.fetch; context.after(() => { globalThis.fetch = originalFetch; });
  globalThis.fetch = async (_input, init) => { requests.push(JSON.parse(String(init?.body))); return new Response("{}", { status: 503, headers: { "Content-Type": "application/json" } }); };
  local.labelMessage = { key: "validationDraft" }; local.summaryMessage = { key: "validationGraphSummary" };
  local.nodes[0]!.labelMessage = { key: "stageLabel_context" }; const before = structuredClone(local);
  for (const locale of ["en-US", "zh-CN"] as const) {
    await instance.changeLanguage(locale); workflowProjectionText(instance.getFixedT(locale, "workflow"), local, "summary");
    await assert.rejects(saveWorkflowDraftDevRecord(local, config, 6, 2), /HTTP 503 without a stable failure envelope/);
  }
  assert.equal(requests.length, 2); assert.deepEqual(requests[0], requests[1]);
  const body = requests[0] as { expected_draft_version: number; expected_lifecycle_version: number; draft: { name: string; description: string } };
  assert.equal(body.expected_draft_version, 6); assert.equal(body.expected_lifecycle_version, 2);
  assert.equal(body.draft.name, local.label); assert.equal(body.draft.description, local.summary);
  assert.doesNotMatch(JSON.stringify(body), /Message|validationGraphSummary|stageLabel_context|uiLocale/);
  assert.deepEqual(local, before);
});
