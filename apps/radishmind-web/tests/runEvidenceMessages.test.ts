import assert from "node:assert/strict";
import test from "node:test";
import { createUiI18n, initializeUiI18n } from "../src/i18n/instance.ts";
import { runEvidence as en } from "../src/i18n/locales/en-US/runEvidence.ts";
import { runEvidence as zh } from "../src/i18n/locales/zh-CN/runEvidence.ts";
import { runEvidenceState, runEvidenceBoundary, runEvidenceAction, runEvidenceFinding, runEvidenceNodeType, runEvidenceSourceType, runComparisonFailure } from "../src/features/control-plane-read/runEvidenceMessages.ts";

test("run evidence translates stable states without rewriting record values or unknown codes", async () => {
  const instance = createUiI18n(); await initializeUiI18n(instance, "en-US");
  instance.addResourceBundle("en-US", "evaluation", { evidence: en });
  instance.addResourceBundle("zh-CN", "evaluation", { evidence: zh });
  const evidence = Object.freeze({ state: "failed", boundary: "gateway", action: "check_gateway_capacity", finding: "status_regressed", code: "future_code", node: "llm", source: "manual" });
  for (const locale of ["en-US", "zh-CN", "en-US"] as const) {
    await instance.changeLanguage(locale);
    const t = instance.getFixedT(locale, "evaluation");
    assert.equal(runEvidenceState(t, evidence.state), locale === "en-US" ? "Failed" : "失败");
    assert.equal(runEvidenceBoundary(t, evidence.boundary), locale === "en-US" ? "Gateway" : "网关");
    assert.equal(runEvidenceAction(t, evidence.action), locale === "en-US" ? "Check gateway capacity" : "检查网关容量");
    assert.equal(runEvidenceFinding(t, evidence.finding), locale === "en-US" ? "Status regressed" : "运行状态回归");
    assert.equal(runEvidenceNodeType(t, evidence.node), locale === "en-US" ? "Language model" : "语言模型");
    assert.equal(runEvidenceSourceType(t, evidence.source), locale === "en-US" ? "Manual" : "手册");
    const resources = locale === "en-US" ? en : zh;
    for (const value of Object.keys(resources.states)) assert.ok(runEvidenceState(t, value));
    for (const value of Object.keys(resources.actions)) assert.ok(runEvidenceAction(t, value));
    for (const value of Object.keys(resources.findings)) assert.ok(runEvidenceFinding(t, value));
    assert.equal(runEvidenceState(t, ""), resources.absent);
    assert.equal(runEvidenceBoundary(t, ""), resources.none);
    assert.equal(runEvidenceAction(t, ""), resources.none);
    assert.equal(runEvidenceState(t, evidence.code), evidence.code);
    assert.equal(runEvidenceBoundary(t, evidence.code), evidence.code);
    assert.equal(runEvidenceNodeType(t, evidence.code), evidence.code);
    assert.equal(runEvidenceSourceType(t, evidence.code), evidence.code);
    assert.equal(runEvidenceFinding(t, evidence.code), resources.unknownFinding);
    assert.equal(runEvidenceAction(t, evidence.code), resources.unknownAction);
    assert.equal(runComparisonFailure(t, evidence.code), resources.failures.unknown);
    assert.equal(runComparisonFailure(t, "scope"), resources.failures.scope);
    assert.equal(evidence.state, "failed");
    assert.equal(t($ => $.evidence.bytes, { value: "0" }), locale === "en-US" ? "0 bytes" : "0 字节");
  }
});
