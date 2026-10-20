import assert from "node:assert/strict";
import test from "node:test";
import { createUiI18n, initializeUiI18n } from "../src/i18n/instance.ts";
import { evaluationCases as en } from "../src/i18n/locales/en-US/evaluationCases.ts";
import { evaluationCases as zh } from "../src/i18n/locales/zh-CN/evaluationCases.ts";
import { evaluationState, evaluationFailure, evaluationFailureCode } from "../src/features/control-plane-read/evaluationMessages.ts";
import { EvaluationRequestFailure, evaluationEnvelopeFailure } from "../src/features/control-plane-read/evaluationFailure.ts";

test("evaluation messages retranslate stable state, plural counts and failures without changing exact values", async () => {
  const instance = createUiI18n(); await initializeUiI18n(instance, "en-US");
  instance.addResourceBundle("en-US", "evaluation", { cases: en });
  instance.addResourceBundle("zh-CN", "evaluation", { cases: zh });
  assert.deepEqual(Object.keys(en), Object.keys(zh));
  assert.deepEqual(Object.keys(en.states), Object.keys(zh.states));
  const evidence = Object.freeze({ expected: "regression", decision: "approved", version: 2, name: "人工 name", profile: "workflow_definition_executor.v1" });
  const before = JSON.stringify(evidence);
  const code = evaluationFailureCode(evaluationEnvelopeFailure({ failure_code: "workflow_evaluation_suite_approval_blocked", failure_summary: "RAW PRIVATE PROSE" }));
  for (const locale of ["en-US", "zh-CN", "en-US"] as const) {
    await instance.changeLanguage(locale);
    const t = instance.getFixedT(locale, "evaluation");
    assert.equal(evaluationState(t, evidence.expected), locale === "en-US" ? "Regression" : "回归");
    assert.equal(evaluationState(t, evidence.decision), locale === "en-US" ? "Approved" : "已批准");
    assert.equal(evaluationState(t, evidence.profile), evidence.profile);
    assert.equal(t($ => $.cases.candidateCount, { count: 1 }), locale === "en-US" ? "1 candidate" : "1 个候选");
    assert.equal(t($ => $.cases.candidateCount, { count: 2 }), locale === "en-US" ? "2 candidates" : "2 个候选");
    assert.match(evaluationFailure(t, code), locale === "en-US" ? /Approval is blocked/ : /不允许批准/);
    assert.doesNotMatch(evaluationFailure(t, code), /RAW PRIVATE PROSE/);
    for (const key of Object.keys(en.states)) assert.ok(evaluationState(t, key));
    for (const code of ["workflow_evaluation_invalid", "workflow_evaluation_suite_invalid", "workflow_evaluation_not_found", "workflow_evaluation_suite_not_found", "workflow_evaluation_run_not_eligible", "workflow_evaluation_suite_case_not_eligible", "workflow_evaluation_store_unavailable", "workflow_evaluation_suite_store_unavailable", "workflow_evaluation_store_contract_mismatch", "workflow_evaluation_suite_store_contract_mismatch", "workflow_evaluation_suite_review_changed", "offline", "contract", "future"]) assert.ok(evaluationFailure(t, code));
  }
  assert.equal(JSON.stringify(evidence), before);
});

test("evaluation failure state contains machine codes only and does not parse exception prose", () => {
  for (const value of [null, 42, {}, { failure_code: "https://private.invalid/token" }, { failure_code: "workflow_evaluation_" + "a".repeat(100) }]) {
    assert.equal(evaluationEnvelopeFailure(value).code, "contract");
  }
  const error = evaluationEnvelopeFailure({ failure_code: "workflow_evaluation_invalid", failure_summary: "SECRET-PROSE" });
  assert.ok(error instanceof EvaluationRequestFailure);
  assert.equal(evaluationFailureCode(error), "workflow_evaluation_invalid");
  assert.doesNotMatch(error.message, /SECRET-PROSE/);
  assert.equal(evaluationFailureCode(new Error("workflow_evaluation_invalid: SECRET-PROSE")), "unknownFailure");
});
