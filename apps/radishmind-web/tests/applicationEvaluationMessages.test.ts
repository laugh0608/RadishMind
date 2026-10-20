import assert from "node:assert/strict";
import test from "node:test";
import { createUiI18n, initializeUiI18n } from "../src/i18n/instance.ts";
import { evaluationCampaign as en } from "../src/i18n/locales/en-US/evaluationCampaign.ts";
import { evaluationCampaign as zh } from "../src/i18n/locales/zh-CN/evaluationCampaign.ts";
import { campaignState, campaignFailure, campaignMessage, campaignOperation, type CampaignMessage } from "../src/features/control-plane-read/applicationEvaluationMessages.ts";
import { APPLICATION_EVALUATION_SCHEDULE_FAILURE_CODES } from "../src/features/control-plane-read/applicationEvaluationScheduleConsumer.ts";

test("campaign and schedule feedback retranslate immutable messages, raw references and plural counts", async () => {
  const instance = createUiI18n(); await initializeUiI18n(instance, "en-US");
  instance.addResourceBundle("en-US", "evaluation", { campaign: en });
  instance.addResourceBundle("zh-CN", "evaluation", { campaign: zh });
  const messages: CampaignMessage[] = [
    { kind: "planLoadedMessage", id: "aeplan_raw引用", version: 3 },
    { kind: "planArchivedMessage", id: "aeplan_raw引用" },
    { kind: "planSavedMessage", id: "aeplan_raw引用", version: 3, operation: "revise" },
    { kind: "campaignFinishedMessage", id: "aecamp_raw引用", state: "succeeded" },
    { kind: "campaignReconciledMessage", id: "aecamp_raw引用", state: "interrupted" },
    { kind: "handoffMessage", count: 1, state: "partial", suite: "" },
    { kind: "handoffMessage", count: 2, state: "complete", suite: "suite_raw引用" },
    { kind: "scheduleSavedMessage", operation: "pause", version: 7 },
    { kind: "occurrenceLoadedMessage", time: "2026-10-11T02:30:00Z" },
    { kind: "strictFailure" }, { kind: "invalidPlan" }, { kind: "invalidCadence" },
  ];
  const before = JSON.stringify(messages);
  for (const locale of ["en-US", "zh-CN", "en-US"] as const) {
    await instance.changeLanguage(locale);
    const t = instance.getFixedT(locale, "evaluation");
    assert.equal(campaignState(t, "paused"), locale === "en-US" ? "paused" : "已暂停");
    assert.equal(campaignState(t, "scheduled_future_code"), "scheduled_future_code");
    assert.equal(campaignOperation(t, "activate"), locale === "en-US" ? "Activate" : "激活");
    const rendered = messages.map(message => campaignMessage(t, message));
    assert.match(rendered[0], /aeplan_raw引用.*3/);
    assert.match(rendered[5], locale === "en-US" ? /1 exact Case ref; Suite pending/ : /1 个精确用例引用/);
    assert.match(rendered[6], locale === "en-US" ? /2 exact Case refs/ : /2 个精确用例引用/);
    assert.match(rendered[8], /2026-10-11T02:30:00Z/);
    assert.match(rendered[10], locale === "en-US" ? /valid JSON/ : /有效 JSON/);
    assert.equal(campaignMessage(t, null), "");
    assert.equal(JSON.stringify(messages), before);
    assert.ok(rendered.every(value => !value.includes("campaign.") && !value.includes("{{")));
  }
});

test("all bounded evaluation and schedule errors have local guidance; unknown codes cannot expose server prose", async () => {
  const instance = createUiI18n(); await initializeUiI18n(instance, "en-US");
  instance.addResourceBundle("en-US", "evaluation", { campaign: en });
  instance.addResourceBundle("zh-CN", "evaluation", { campaign: zh });
  assert.deepEqual(Object.keys(en.failures).sort(), [...APPLICATION_EVALUATION_SCHEDULE_FAILURE_CODES].sort());
  for (const locale of ["en-US", "zh-CN"] as const) {
    const t = instance.getFixedT(locale, "evaluation");
    for (const code of APPLICATION_EVALUATION_SCHEDULE_FAILURE_CODES) {
      const result = campaignFailure(t, code);
      assert.notEqual(result, campaignFailure(t, null));
      assert.ok(locale === "en-US" ? /[a-z]/.test(result) : /[\u4e00-\u9fff]/.test(result));
    }
    assert.match(campaignFailure(t, "application_evaluation_schedule_version_conflict"), locale === "en-US" ? /Local inputs are retained/ : /本地输入仍保留/);
    assert.equal(campaignFailure(t, "SERVER-PRIVATE-PROSE"), campaignFailure(t, null));
    assert.doesNotMatch(campaignFailure(t, "SERVER-PRIVATE-PROSE"), /SERVER-PRIVATE/);
  }
});
