import assert from "node:assert/strict";
import test from "node:test";
import { isPromptClientKey, promptExecutionNeedsReview } from "../src/features/control-plane-read/promptExecutionRecovery.ts";

test("Prompt execution keys reject invalid input before dispatch", () => {
  for (const key of ["12345678", "prompt-turn.1_2:3", "a".repeat(160)]) assert.equal(isPromptClientKey(key), true);
  for (const key of ["", "1234567", "a".repeat(161), " bad-key", "valid-key ", "-bad-key", "bad/key!"]) assert.equal(isPromptClientKey(key), false);
});

test("Prompt interrupted or unresolved execution stays under review", () => {
  for (const code of ["prompt_invocation_transport_failed", "prompt_invocation_response_invalid", "prompt_invocation_canceled", "prompt_invocation_duplicate_running", "prompt_invocation_outcome_unknown", "application_session_store_unavailable", "application_session_response_invalid", "application_session_request_canceled", "application_session_run_outcome_unknown", "application_session_turn_interrupted", "application_session_turn_canceled"]) {
    assert.equal(promptExecutionNeedsReview(false, code), true, code);
  }
  for (const status of ["running", "canceled", "outcome_unknown", "unexpected_status"]) assert.equal(promptExecutionNeedsReview(false, "", status), true);
  assert.equal(promptExecutionNeedsReview(false, "prompt_invocation_outcome_unknown", "failed"), true);
});

test("Prompt metadata reads and later pre-dispatch rejections cannot resolve an earlier attempt", () => {
  for (const code of ["", "api_key_invalid", "api_key_expired", "api_key_revoked", "api_key_scope_denied", "application_session_authority_changed", "prompt_invocation_input_invalid"]) {
    assert.equal(promptExecutionNeedsReview(true, code), true);
    assert.equal(promptExecutionNeedsReview(false, code), false);
  }
});

test("Prompt known execution terminal metadata resolves the earlier attempt", () => {
  assert.equal(promptExecutionNeedsReview(true, "", "succeeded"), false);
  assert.equal(promptExecutionNeedsReview(true, "prompt_invocation_output_contract_failed", "failed"), false);
});
