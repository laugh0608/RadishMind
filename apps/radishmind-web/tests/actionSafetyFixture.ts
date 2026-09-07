type OwnerKind = "agent_copilot_response" | "agent_copilot_runtime_assignment" |
  "workflow_http_tool_action_plan" | "workflow_run";
type Level = "answer_only" | "handoff_ready" | "tool_callable";

export function recordedActionSafetyFixture({
  ownerKind,
  ownerId,
  ownerVersion,
  applicationId,
  level,
  observed = false,
}: {
  ownerKind: OwnerKind;
  ownerId: string;
  ownerVersion: number;
  applicationId: string;
  level: Level;
  observed?: boolean;
}): Record<string, any> {
  const tool = level === "tool_callable";
  const handoff = level === "handoff_ready";
  return {
    schema_version: "action_safety_read_projection.v1",
    status: "recorded",
    owner: { kind: ownerKind, id: ownerId, version: ownerVersion, digest: digest("a") },
    projection_version: ownerKind === "agent_copilot_response"
      ? "action_safety_response_projection.v1"
      : ownerKind === "agent_copilot_runtime_assignment"
        ? "action_safety_assignment_projection.v1"
        : ownerKind === "workflow_http_tool_action_plan"
          ? "action_safety_plan_projection.v1"
          : "action_safety_run_projection.v1",
    projection_digest: digest("b"),
    decisions: [{
      decision_id: "asd_aaaaaaaaaaaaaaaa",
      decision_digest: digest("c"),
      scope: {
        tenant_ref: "tenant_demo",
        workspace_id: "workspace_demo",
        environment: "development",
        application_id: applicationId,
      },
      source: {
        kind: tool ? "workflow_http_tool_action" : handoff ? "agent_copilot_proposed_action" : "copilot_response",
        schema_version: tool ? "workflow_http_tool_action.v1" : handoff ? "agent_copilot_proposed_action.v1" : "copilot_response.v1",
        id: ownerId,
        version: ownerVersion,
        digest: digest("d"),
      },
      project: "radishflow",
      task: tool ? "workflow_http_tool" : "suggest_flowsheet_edits",
      action_kind: tool ? "workflow_http_tool" : handoff ? "candidate_edit" : "none",
      risk_level: tool ? "medium" : handoff ? "medium" : "low",
      target_kind: tool ? "workflow_http_tool" : handoff ? "human_review_owner" : "none",
      method: tool ? "GET" : "none",
      requested_level: level,
      maximum_allowed_level: level,
      effective_level: level,
      requires_confirmation: tool,
      confirmation_state: tool ? "approved" : "not_required",
      writes_business_truth: false,
      side_effect_budget: {
        provider_calls: level === "answer_only" ? 1 : 0,
        handoff_refs: handoff ? 1 : 0,
        tool_network_calls: tool ? 1 : 0,
        confirmation_consumptions: tool ? 1 : 0,
        business_writes: 0,
        replay_writes: 0,
      },
      blockers: [],
      policy: { version: "action_safety_runtime_policy.v1", digest: digest("e") },
      created_at: "2026-08-30T02:00:00Z",
    }],
    ...(observed ? {
      observed_side_effects: {
        provider_calls: tool ? 0 : 1,
        tool_calls: tool ? 1 : 0,
        confirmation_calls: tool ? 1 : 0,
        business_writes: 0,
        replay_writes: 0,
      },
    } : {}),
  };
}

export function legacyActionSafetyFixture(
  ownerKind: OwnerKind,
  ownerId: string,
  ownerVersion: number,
): Record<string, any> {
  return {
    schema_version: "action_safety_read_projection.v1",
    status: "not_recorded_legacy",
    owner: { kind: ownerKind, id: ownerId, version: ownerVersion, digest: digest("f") },
    decisions: [],
  };
}

function digest(character: string): string {
  return `sha256:${character.repeat(64)}`;
}
