import assert from "node:assert/strict";
import test from "node:test";

import { controlledUseFailureGuidance } from "../src/features/control-plane-read/controlledUseFailureGuidance.ts";

test("Prompt invocation authority failures hand off only to Prompt Assignment", () => {
  for (const failureCode of [
    "prompt_runtime_assignment_not_found",
    "prompt_runtime_candidate_ineligible",
    "prompt_runtime_authority_changed",
  ]) {
    const guidance = controlledUseFailureGuidance("prompt_invocation", failureCode);
    assert.equal(guidance?.assignmentAnchor, "prompt-application-runtime-assignment");
    assert.match(guidance?.sideEffectSummary ?? "", /No provider call was made/u);
  }
});

test("Prompt and Agent sessions resolve the same server failure to their exact Assignment owner", () => {
  for (const failureCode of [
    "application_session_authority_not_found",
    "application_session_authority_changed",
    "application_session_profile_ineligible",
  ]) {
    assert.equal(
      controlledUseFailureGuidance("prompt_session", failureCode)?.assignmentAnchor,
      "prompt-application-runtime-assignment",
    );
    assert.equal(
      controlledUseFailureGuidance("agent_session", failureCode)?.assignmentAnchor,
      "agent-copilot-runtime-assignment",
    );
  }
});

test("permission, input, transport, storage, cancellation, and outcome failures keep their original owner", () => {
  for (const failureCode of [
    "prompt_runtime_scope_denied",
    "prompt_invocation_input_invalid",
    "prompt_invocation_transport_failed",
    "application_session_store_unavailable",
    "application_session_request_canceled",
    "application_session_run_outcome_unknown",
  ]) {
    assert.equal(controlledUseFailureGuidance("prompt_invocation", failureCode), null);
    assert.equal(controlledUseFailureGuidance("prompt_session", failureCode), null);
    assert.equal(controlledUseFailureGuidance("agent_session", failureCode), null);
  }
});
