import assert from "node:assert/strict";
import test from "node:test";
import { createUiI18n, initializeUiI18n } from "../src/i18n/instance.ts";
import { agent as en } from "../src/i18n/locales/en-US/agent.ts";
import { agent as zh } from "../src/i18n/locales/zh-CN/agent.ts";
import { agentSessionMessage, profileFindingMessage, profileOperationMessage } from "../src/features/control-plane-read/agentCopilotMessages.ts";
import type { AgentCopilotProfileOperation } from "../src/features/control-plane-read/agentCopilotProfileConsumer.ts";
import type { AgentCopilotSessionResult } from "../src/features/control-plane-read/agentCopilotSessionConsumer.ts";

async function messages() {
  const instance = createUiI18n(); await initializeUiI18n(instance, "en-US");
  instance.addResourceBundle("en-US", "agent", en); instance.addResourceBundle("zh-CN", "agent", zh);
  return { instance, t: instance.getFixedT(null, "agent") };
}

test("Agent diagnostics use code plus field, preserve raw source and reject unknown or inherited keys", async () => {
  const { instance, t } = await messages();
  const finding = { field: "profile_name", code: "agent_copilot_profile_payload_invalid", summary: "原始 backend diagnostic" };
  assert.equal(profileFindingMessage(t, finding), "Profile names must contain 2–80 characters.");
  for (const field of ["constructor", "toString", "unknown", "scope"]) {
    assert.match(profileFindingMessage(t, { ...finding, field, code: "unknown" }), /Review the diagnostic field/);
  }
  await instance.changeLanguage("zh-CN");
  assert.equal(profileFindingMessage(t, finding), "Profile 名称必须为 2 至 80 个字符。");
  assert.equal(finding.summary, "原始 backend diagnostic");
});

test("Agent Profile completion and conflict messages update with language without changing exact versions", async () => {
  const { instance, t } = await messages();
  const operation: AgentCopilotProfileOperation = { status: "versioned", draft: null, version: null, currentDraftVersion: 7, currentProfileVersion: 3, validation: { state: "valid", isValid: true, findings: [] }, failureCode: "", summary: "legacy fixed message" };
  const before = structuredClone(operation);
  assert.equal(profileOperationMessage(t, operation, false), "Immutable Profile v3 created.");
  await instance.changeLanguage("zh-CN");
  assert.equal(profileOperationMessage(t, operation, false), "不可变 Profile v3 已创建。");
  assert.deepEqual(operation, before);
  assert.match(profileOperationMessage(t, { ...operation, status: "version_conflict" }, true), /重读精确 Profile/);
  assert.equal(profileOperationMessage(t, { ...operation, status: "idle" }, true), "Profile 包含未保存的编辑。");
});

test("Agent replay and cancellation retain their meaning without implying restored answers", async () => {
  const { instance, t } = await messages();
  const result: AgentCopilotSessionResult = { status: "replayed", session: null, turn: null, response: null, actionSafety: null, resultArtifact: null, resultArtifactFailureCode: "", failureCode: "", failureSummary: "", idempotentReplay: true, summary: "old summary" };
  const before = structuredClone(result);
  assert.match(agentSessionMessage(t, result), /model was not called again and the answer was not restored/);
  await instance.changeLanguage("zh-CN");
  assert.match(agentSessionMessage(t, result), /未重复调用模型，也未恢复回答/);
  assert.match(agentSessionMessage(t, { ...result, status: "blocked", failureCode: "application_session_request_canceled" }), /迟到响应会被丢弃/);
  assert.deepEqual(result, before);
});
