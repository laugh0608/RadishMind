import assert from "node:assert/strict";
import test from "node:test";
import { createUiI18n, initializeUiI18n } from "../src/i18n/instance.ts";
import { artifactFeedback as artifactEn } from "../src/i18n/locales/en-US/artifactFeedback.ts";
import { artifactFeedback as artifactZh } from "../src/i18n/locales/zh-CN/artifactFeedback.ts";
import { workflowRAGEvaluation as ragEn } from "../src/i18n/locales/en-US/workflowRAGEvaluation.ts";
import { workflowRAGEvaluation as ragZh } from "../src/i18n/locales/zh-CN/workflowRAGEvaluation.ts";
import { auxiliaryEvidence as auxiliaryEn } from "../src/i18n/locales/en-US/auxiliaryEvidence.ts";
import { auxiliaryEvidence as auxiliaryZh } from "../src/i18n/locales/zh-CN/auxiliaryEvidence.ts";
import { gatewayOverviewEvidence as overviewEn } from "../src/i18n/locales/en-US/gatewayOverviewEvidence.ts";
import { gatewayOverviewEvidence as overviewZh } from "../src/i18n/locales/zh-CN/gatewayOverviewEvidence.ts";
import { artifactFailureMessage } from "../src/features/control-plane-read/artifactFeedback.ts";
import { ragEvaluationMessage, ragEvaluationState, ragEvaluationFailureCode, ragEvaluationMetric } from "../src/features/control-plane-read/workflowRAGEvaluationMessages.ts";
import { evidenceProjectionMessage, evidenceProjectionStatus, evidenceProjectionText, type EvidenceProjectionMessage } from "../src/features/control-plane-read/evidenceProjectionCopy.ts";

async function setup() {
  const instance = createUiI18n(); await initializeUiI18n(instance, "en-US");
  for (const [locale, artifact, rag, auxiliary, overview] of [
    ["en-US", artifactEn, ragEn, auxiliaryEn, overviewEn], ["zh-CN", artifactZh, ragZh, auxiliaryZh, overviewZh],
  ] as const) {
    instance.addResourceBundle(locale, "applications", { artifactFeedback: artifact });
    instance.addResourceBundle(locale, "workflow", { ragEvaluation: rag });
    instance.addResourceBundle(locale, "evaluation", { auxiliary: { ...auxiliary, ...overview } });
  }
  return instance;
}

test("result failures give bilingual recovery guidance without exposing unknown server prose", async () => {
  const instance = await setup();
  for (const locale of ["en-US", "zh-CN"] as const) {
    const t = instance.getFixedT(locale, "applications");
    for (const code of ["application_result_artifact_lifecycle_version_conflict", "application_result_artifact_store_contract_mismatch", "application_session_scope_denied"]) {
      assert.match(artifactFailureMessage(t, code), new RegExp(code));
      assert.match(artifactFailureMessage(t, code), locale === "en-US" ? /[a-z]/ : /[\u4e00-\u9fff]/);
    }
    assert.doesNotMatch(artifactFailureMessage(t, "private server prose"), /private server prose/);
    assert.equal(artifactFailureMessage(t, "private server prose"), artifactFailureMessage(t, ""));
  }
});

test("RAG review feedback, metrics and plural counts retranslate stable state while raw codes remain exact", async () => {
  const instance = await setup();
  const conflict = { status: "version_conflict", failureCode: "workflow_rag_evaluation_version_conflict" } as const;
  const before = JSON.stringify(conflict);
  for (const locale of ["en-US", "zh-CN", "en-US"] as const) {
    const t = instance.getFixedT(locale, "workflow");
    assert.match(ragEvaluationMessage(t, conflict), locale === "en-US" ? /local draft is preserved/ : /本地草稿已保留/);
    assert.equal(ragEvaluationFailureCode(conflict.failureCode), conflict.failureCode);
    assert.equal(ragEvaluationFailureCode("SERVER-PRIVATE-PROSE"), "");
    assert.equal(ragEvaluationFailureCode("future_evaluation_failure"), "future_evaluation_failure");
    assert.doesNotMatch(ragEvaluationMessage(t, { status: "failed", failureCode: "SERVER-PRIVATE-PROSE" }), /SERVER-PRIVATE/);
    assert.match(ragEvaluationMessage(t, { status: "failed", failureCode: "workflow_rag_evaluation_binding_invalid" }), locale === "en-US" ? /exact snapshot/ : /精确快照/);
    assert.equal(ragEvaluationState(t, "future_state"), "future_state");
    assert.equal(ragEvaluationState(t, "improved"), locale === "en-US" ? "Improved" : "改善");
    assert.equal(ragEvaluationMetric(t, "meanReciprocalRank"), locale === "en-US" ? "Mean reciprocal rank" : "平均倒数排名");
    assert.equal(t($ => $.ragEvaluation.sampleCount, { count: 1 }), locale === "en-US" ? "1 sample" : "1 个样本");
    assert.equal(t($ => $.ragEvaluation.sampleCount, { count: 2 }), locale === "en-US" ? "2 samples" : "2 个样本");
    assert.equal(JSON.stringify(conflict), before);
  }
});

test("auxiliary presentation metadata translates counts and known states without changing raw references", async () => {
  const instance = await setup();
  const message: EvidenceProjectionMessage = { key: "gatewayOverviewEvidence_quota_counts", values: { used: 1200, limit: 2500 } };
  const before = JSON.stringify(message);
  for (const locale of ["en-US", "zh-CN"] as const) {
    const t = instance.getFixedT(locale, "evaluation");
    assert.equal(evidenceProjectionMessage(t, message), "1,200 / 2,500");
    assert.equal(evidenceProjectionStatus(t, "credential required"), locale === "en-US" ? "credential required" : "需要凭据");
    assert.equal(evidenceProjectionStatus(t, "profile_raw-reference"), "profile_raw-reference");
    assert.equal(evidenceProjectionText(t, { value: "raw source 文本" }, "value"), "raw source 文本");
    assert.equal(evidenceProjectionText(t, { value: "raw counts", valueMessage: message }, "value"), "1,200 / 2,500");
    assert.equal(JSON.stringify(message), before);
  }
});
