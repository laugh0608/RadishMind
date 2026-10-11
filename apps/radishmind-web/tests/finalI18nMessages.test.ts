import assert from "node:assert/strict";
import test from "node:test";
import { createUiI18n, initializeUiI18n } from "../src/i18n/instance.ts";
import { interaction as enInteraction } from "../src/i18n/locales/en-US/interaction.ts";
import { interaction as zhInteraction } from "../src/i18n/locales/zh-CN/interaction.ts";
import { readEvidence as enRead } from "../src/i18n/locales/en-US/readEvidence.ts";
import { readEvidence as zhRead } from "../src/i18n/locales/zh-CN/readEvidence.ts";
import { workflowReadProjection as enProjection } from "../src/i18n/locales/en-US/workflowReadProjection.ts";
import { workflowReadProjection as zhProjection } from "../src/i18n/locales/zh-CN/workflowReadProjection.ts";
import { interactionFailureCode, interactionFailureMessage, interactionStatus } from "../src/features/control-plane-read/applicationInteractionMessages.ts";
import { workflowProjectionText } from "../src/features/control-plane-read/workflowProjectionCopy.ts";
import { readEvidenceStatus } from "../src/features/control-plane-read/readEvidenceMessages.ts";
import { buildControlPlaneReadShellViewModel } from "../src/features/control-plane-read/readShell.ts";

async function setup() {
  const i18n = createUiI18n(); await initializeUiI18n(i18n, "en-US");
  for (const [locale, interaction, readEvidence, projection] of [
    ["en-US", enInteraction, enRead, enProjection], ["zh-CN", zhInteraction, zhRead, zhProjection],
  ] as const) {
    i18n.addResourceBundle(locale, "applications", { interaction });
    i18n.addResourceBundle(locale, "workflow", { readEvidence, projection });
  }
  return i18n;
}

test("interaction errors retain bounded diagnostic codes and retranslate recovery without server prose", async () => {
  const i18n = await setup();
  for (const locale of ["en-US", "zh-CN"] as const) {
    const t = i18n.getFixedT(locale, "applications");
    assert.match(interactionFailureMessage(t, "application_session_version_conflict"), locale === "en-US" ? /Reload current metadata/ : /重新读取当前元数据/);
    assert.match(interactionFailureMessage(t, "application_session_request_canceled"), locale === "en-US" ? /server state/ : /服务端状态/);
    assert.doesNotMatch(interactionFailureMessage(t, "application_session_new_failure"), /No provider call|未调用提供方/);
    assert.equal(interactionFailureMessage(t, "secret server prose"), interactionFailureMessage(t, "unknown"));
    assert.equal(interactionFailureCode("secret server prose"), "");
    assert.equal(interactionFailureCode("application_session_new_failure"), "application_session_new_failure");
    assert.equal(interactionStatus(t, "future_state"), "future_state");
    assert.equal(interactionStatus(t, "closed"), locale === "en-US" ? "Closed" : "已关闭");
    assert.equal(t($ => $.interaction.turns, { count: 1 }), locale === "en-US" ? "1 turn" : "1 轮");
    assert.equal(t($ => $.interaction.turns, { count: 2 }), locale === "en-US" ? "2 turns" : "2 轮");
    assert.match(t($ => $.interaction.structuredSubmitted, { fields: "original_field" }), /original_field/);
  }
});

test("read shell labels and fixed state explanations change without changing canonical routes or evidence", async () => {
  const i18n = await setup(), shell = buildControlPlaneReadShellViewModel();
  const original = structuredClone(shell);
  for (const locale of ["en-US", "zh-CN", "en-US"] as const) {
    const t = i18n.getFixedT(locale, "workflow");
    for (const state of shell.statePreviews) {
      const label = workflowProjectionText(t, state, "label");
      const summary = workflowProjectionText(t, state, "summary");
      assert.match(summary, locale === "en-US" ? /[A-Za-z]/ : /[\u4e00-\u9fff]/);
      if (state.id === "empty") assert.equal(label, locale === "en-US" ? "Empty" : "空");
    }
    assert.equal(t($ => $.readEvidence.routes["application-summary-list-route"]), locale === "en-US" ? "Applications" : "应用列表");
    assert.equal(readEvidenceStatus(t, "cursor_required"), locale === "en-US" ? "Cursor required" : "需要游标");
    assert.equal(readEvidenceStatus(t, "future_protocol_value"), "future_protocol_value");
    assert.equal(workflowProjectionText(t, { summary: "Original 原文 <literal>" }, "summary"), "Original 原文 <literal>");
  }
  assert.deepEqual(shell, original);
  assert.equal(shell.routeCards.find(route => route.routeId === "application-summary-list-route")?.path, "/v1/user-workspace/applications");
});
