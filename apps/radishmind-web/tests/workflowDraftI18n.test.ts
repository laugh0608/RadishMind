import { workflowLibrary as enLibrary } from "../src/i18n/locales/en-US/workflowLibrary.ts";
import { workflowLibrary as zhLibrary } from "../src/i18n/locales/zh-CN/workflowLibrary.ts";
import { workflowRevision as enRevision } from "../src/i18n/locales/en-US/workflowRevision.ts";
import { workflowRevision as zhRevision } from "../src/i18n/locales/zh-CN/workflowRevision.ts";
import { workflowCanvas as enCanvas } from "../src/i18n/locales/en-US/workflowCanvas.ts";
import { workflowCanvas as zhCanvas } from "../src/i18n/locales/zh-CN/workflowCanvas.ts";
import assert from "node:assert/strict";
import test from "node:test";
import { createUiI18n, initializeUiI18n } from "../src/i18n/instance.ts";
import { workflowDraft as en } from "../src/i18n/locales/en-US/workflowDraft.ts";
import { workflowDraft as zh } from "../src/i18n/locales/zh-CN/workflowDraft.ts";
import { workflowDraftConsumerMessage, workflowDraftLifecycleMessage, workflowDraftNodeTypeLabel, workflowDraftStatusLabel } from "../src/features/control-plane-read/workflowDraftMessages.ts";
import { workflowCanvasMessage, workflowValidationLabel, type WorkflowCanvasMessage } from "../src/features/control-plane-read/workflowCanvasMessages.ts";
import { workflowRevisionOperationMessage } from "../src/features/control-plane-read/workflowRevisionMessages.ts";
import { initialWorkflowSavedDraftConsumerState, initialWorkflowSavedDraftLifecycleOperationState } from "../src/features/control-plane-read/savedWorkflowDraftConsumer.ts";

async function languageInstance() {
  const instance = createUiI18n();
  await initializeUiI18n(instance, "en-US");
  instance.addResourceBundle("en-US", "workflow", { canvas: enCanvas, revision: enRevision, library: enLibrary, draft: en });
  instance.addResourceBundle("zh-CN", "workflow", { canvas: zhCanvas, revision: zhRevision, library: zhLibrary, draft: zh });
  return instance;
}

test("canvas feedback already in state switches language while preserving user text and identifiers", async () => {
  const instance = await languageInstance();
  const message: WorkflowCanvasMessage = { code: "nodeSelected", label: "用户 Prompt <script>", id: "node_unchanged" };
  const before = structuredClone(message);
  assert.equal(workflowCanvasMessage(instance.getFixedT("en-US", "workflow"), message), "Selected node: 用户 Prompt <script> (node_unchanged).");
  await instance.changeLanguage("zh-CN");
  assert.equal(workflowCanvasMessage(instance.getFixedT("zh-CN", "workflow"), message), "已选择节点：用户 Prompt <script>（node_unchanged）。");
  assert.deepEqual(message, before);
});

test("unknown failure codes retain the reference without echoing a server summary or claiming success", async () => {
  const instance = await languageInstance();
  const state = { ...initialWorkflowSavedDraftConsumerState({ mode: "sample_only", baseUrl: "", workspaceId: "workspace_test", tenantRef: "tenant_test", subjectRef: "actor_test" }), status: "save_failed" as const, failureCode: "unknown_failure_v99", summary: "UNTRUSTED provider body", sourceLabel: "UNTRUSTED" };
  for (const locale of ["en-US", "zh-CN"] as const) {
    const text = workflowDraftConsumerMessage(instance.getFixedT(locale, "workflow"), state);
    assert.match(text, /unknown_failure_v99/);
    assert.doesNotMatch(text, /UNTRUSTED|success|成功/);
  }
});

test("restore and unarchive explain new revisions and explicit reread without changing authoritative versions", async () => {
  const instance = await languageInstance();
  const t = instance.getFixedT("zh-CN", "workflow");
  const transition = { ...initialWorkflowSavedDraftLifecycleOperationState(), status: "unarchived" as const, draftId: "draft_a", currentDraftVersion: 7, currentLifecycleVersion: 3 };
  const before = structuredClone(transition);
  assert.match(workflowDraftLifecycleMessage(t, transition), /重新打开/);
  assert.equal(workflowRevisionOperationMessage(t, { code: "restored", version: 8 }), "已创建修订 8，历史版本保持不变。");
  assert.match(workflowRevisionOperationMessage(t, { code: "restore_failed", failureCode: "draft_version_conflict" }), /draft_version_conflict/);
  assert.deepEqual(transition, before);
});

test("node roles and validation identifiers have localized labels without changing technical unknown identifiers", async () => {
  const instance = await languageInstance();
  const t = instance.getFixedT("zh-CN", "workflow");
  assert.equal(workflowDraftNodeTypeLabel(t, "condition"), "策略");
  assert.equal(workflowValidationLabel(t, "executor_v0_topology"), "无环提示词至输出路径");
  assert.equal(workflowDraftStatusLabel(t, "draft_future_code"), "draft_future_code");
});
