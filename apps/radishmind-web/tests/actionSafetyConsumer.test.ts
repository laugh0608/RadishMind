import assert from "node:assert/strict";
import test from "node:test";

import {
  parseActionSafetyReadProjection,
  primaryActionSafetyBlocker,
} from "../src/features/control-plane-read/actionSafetyConsumer.ts";

const expected = {
  tenantRef: "tenant_demo",
  workspaceId: "workspace_demo",
  applicationId: "app_flow_copilot",
  ownerKinds: ["agent_copilot_response" as const],
  ownerId: "run_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
  ownerVersion: 1,
};

test("Action Safety parser accepts only the canonical recorded read projection", () => {
  const parsed = parseActionSafetyReadProjection(recordedProjection(), expected);
  assert.equal(parsed?.status, "recorded");
  assert.equal(parsed?.owner.kind, "agent_copilot_response");
  assert.equal(parsed?.decisions[0]?.effectiveLevel, "answer_only");
  assert.equal(parsed?.decisions[0]?.sideEffectBudget.providerCalls, 1);
  assert.equal(primaryActionSafetyBlocker(parsed), "");
});

test("Action Safety parser preserves explicit legacy without current-policy backfill", () => {
  const legacy = {
    schema_version: "action_safety_read_projection.v1",
    status: "not_recorded_legacy",
    owner: {
      kind: "agent_copilot_runtime_assignment",
      id: "acra_aaaaaaaaaaaaaaaa",
      version: 1,
      digest: digest("b"),
    },
    decisions: [],
  };
  const parsed = parseActionSafetyReadProjection(legacy, {
    ...expected,
    ownerKinds: ["agent_copilot_runtime_assignment"],
    ownerId: "acra_aaaaaaaaaaaaaaaa",
  });
  assert.equal(parsed?.status, "not_recorded_legacy");
  assert.equal(parsed?.decisions.length, 0);
  assert.equal(parsed?.projectionDigest, "");
});

test("Action Safety parser rejects unknown, sensitive, drifted, duplicated, and unsafe fields", () => {
  const unknown = recordedProjection();
  unknown.unexpected = true;
  assert.equal(parseActionSafetyReadProjection(unknown, expected), null);

  const sensitive = recordedProjection();
  sensitive.decisions[0].headers = { authorization: "secret" };
  assert.equal(parseActionSafetyReadProjection(sensitive, expected), null);

  const scopeDrift = recordedProjection();
  scopeDrift.decisions[0].scope.workspace_id = "workspace_other";
  assert.equal(parseActionSafetyReadProjection(scopeDrift, expected), null);

  const duplicate = recordedProjection();
  duplicate.decisions.push(structuredClone(duplicate.decisions[0]));
  assert.equal(parseActionSafetyReadProjection(duplicate, expected), null);

  const forgedLevel = recordedProjection();
  forgedLevel.decisions[0].effective_level = "tool_callable";
  assert.equal(parseActionSafetyReadProjection(forgedLevel, expected), null);

  const unsafeObserved = recordedProjection();
  unsafeObserved.observed_side_effects = {
    provider_calls: 1,
    tool_calls: 0,
    confirmation_calls: 0,
    business_writes: 1,
    replay_writes: 0,
  };
  assert.equal(parseActionSafetyReadProjection(unsafeObserved, expected), null);
});

function recordedProjection(): Record<string, any> {
  return {
    schema_version: "action_safety_read_projection.v1",
    status: "recorded",
    owner: {
      kind: "agent_copilot_response",
      id: "run_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
      version: 1,
      digest: digest("a"),
    },
    projection_version: "action_safety_response_projection.v1",
    projection_digest: digest("b"),
    decisions: [{
      decision_id: "asd_aaaaaaaaaaaaaaaa",
      decision_digest: digest("c"),
      scope: {
        tenant_ref: "tenant_demo",
        workspace_id: "workspace_demo",
        environment: "development",
        application_id: "app_flow_copilot",
      },
      source: {
        kind: "copilot_response",
        schema_version: "copilot_response.v1",
        id: "run_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
        version: 1,
        digest: digest("d"),
      },
      project: "radishflow",
      task: "explain_diagnostics",
      action_kind: "none",
      risk_level: "low",
      target_kind: "none",
      method: "none",
      requested_level: "answer_only",
      maximum_allowed_level: "proposal_only",
      effective_level: "answer_only",
      requires_confirmation: false,
      confirmation_state: "not_required",
      writes_business_truth: false,
      side_effect_budget: {
        provider_calls: 1,
        handoff_refs: 0,
        tool_network_calls: 0,
        confirmation_consumptions: 0,
        business_writes: 0,
        replay_writes: 0,
      },
      blockers: [],
      policy: { version: "action_safety_runtime_policy.v1", digest: digest("e") },
      created_at: "2026-08-30T02:00:00Z",
    }],
  };
}

function digest(character: string): string {
  return `sha256:${character.repeat(64)}`;
}
