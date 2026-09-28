import assert from "node:assert/strict";
import test from "node:test";
import { createUiI18n, initializeUiI18n } from "../src/i18n/instance.ts";
import { workflowPromotionMessage, workflowTemplateMessage, type WorkflowPromotionMessage, type WorkflowTemplateMessage } from "../src/features/control-plane-read/workflowOperationMessages.ts";
import { workflowHistoryMessage, type WorkflowHistoryMessage } from "../src/features/control-plane-read/workflowHistoryMessages.ts";
import { executorBlockerMessage, executorStateMessage } from "../src/features/control-plane-read/workflowExecutorMessages.ts";
import type { WorkflowExecutorConsumerState } from "../src/features/control-plane-read/workflowExecutorConsumer.ts";

test("workflow operation evidence renders in either locale without rewriting stable state", async () => {
  const instance = createUiI18n(); await initializeUiI18n(instance, "en-US");
  for (const locale of ["en-US", "zh-CN"]) {
    for (const group of ["Draft", "Promotion", "Template", "Executor", "History"]) {
      const resources = await import(`../src/i18n/locales/${locale}/workflow${group}.ts`);
      instance.addResourceBundle(locale, "workflow", { [group.toLowerCase()]: resources[`workflow${group}`] }, true);
    }
  }
  const promotion: WorkflowPromotionMessage[] = [
    { code: "created", id: "candidate_<literal>", version: 3 },
    { code: "reviewed", decision: "approve" },
    { code: "activated", decision: "activate", version: 4 },
    { code: "runFinished", id: "run_01", schema: "workflow_run_record.v5" },
    { code: "failed", failureCode: "authority_denied" },
    { code: "inputInvalid", failureCode: "input_required" },
    { code: "conflict", failureCode: "review_conflict", review: 7, pointer: 9 },
    { code: "exactDraftRequired" },
  ];
  const template: WorkflowTemplateMessage[] = [
    { code: "offline" }, { code: "loading" }, { code: "versionFailed" },
    { code: "ready", candidates: 0, templates: 5 },
    { code: "loadFailed", failureCode: "catalog_denied" },
    ...(["create", "review", "listing", "derive"] as const).map(action => ({ code: "working" as const, action })),
    { code: "failed", action: "review", failureCode: "review_conflict", review: 7, pointer: 9 },
    { code: "derived", id: "draft_01", version: 1 },
    { code: "completed", action: "listing", audit: "audit_01" },
  ];
  const history: WorkflowHistoryMessage[] = [
    ...(["handoffOffline", "handoffLoading", "handoffLoaded", "handoffUnavailable", "handoffFailed"] as const).map(code => ({ code, id: "run_01" })),
    { code: "diagnosticNeedsDraft" }, { code: "diagnosticWorking" },
    { code: "diagnosticRecorded", scenario: "blocked_tool" }, { code: "diagnosticFailed", failureCode: "diagnostic_denied" },
  ];
  const snapshot = JSON.stringify({ promotion, template, history });
  const render = () => {
    const t = instance.getFixedT(null, "workflow");
    return [...promotion.map(value => workflowPromotionMessage(t, value)), ...template.map(value => workflowTemplateMessage(t, value)), ...history.map(value => workflowHistoryMessage(t, value))];
  };
  const en = render(); await instance.changeLanguage("zh-CN"); const zh = render();
  en.forEach((value, index) => { assert.notEqual(value, zh[index]); assert.ok(!/\{\{|unavailable\. Try|暂时不可用/.test(value)); });
  for (const values of [en, zh]) {
    assert.match(values[0], /candidate_<literal>/);
    assert.match(values[3], /workflow_run_record\.v5/);
    assert.match(values[6], /review_conflict.*7.*9/);
  }
  assert.equal(JSON.stringify({ promotion, template, history }), snapshot);
  const t = instance.getFixedT(null, "workflow");
  for (const status of ["disabled", "idle", "starting", "reading", "succeeded", "failed"] as const) {
    assert.match(executorStateMessage(t, { status, failureCode: "run_failed" } as WorkflowExecutorConsumerState), /[\u4e00-\u9fff]/u);
  }
  assert.match(executorBlockerMessage(t, "unsaved_local_changes"), /[\u4e00-\u9fff]/u);
  assert.match(executorBlockerMessage(t, "future_blocker"), /future_blocker/);
});
