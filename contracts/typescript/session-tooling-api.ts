export const PLATFORM_SESSION_TOOLING_ROUTES = {
  sessionMetadata: "/v1/session/metadata",
  toolsMetadata: "/v1/tools/metadata",
  toolActions: "/v1/tools/actions",
} as const;

export type PlatformMetadataBoundary = {
  surface: "platform_metadata";
  implemented: boolean;
  response_shape: "metadata_only";
};

export type SessionMetadataResponse = {
  schema_version: 1;
  kind: "session_metadata";
  stage: string;
  route: typeof PLATFORM_SESSION_TOOLING_ROUTES.sessionMetadata;
  api_boundary: PlatformMetadataBoundary;
  supported_extension_fields: Array<
    "conversation_id" | "turn_id" | "parent_turn_id" | "history_policy" | "history_window"
  >;
  history_policy: {
    supported_modes: Array<"windowed" | "stateless" | "summary_only" | "disabled">;
    default_mode: "windowed";
    default_max_turns: number;
    include_system: boolean;
    include_tool_results: false;
    compression: {
      enabled: false;
      strategy: "none";
    };
  };
  state_policy: {
    session_state_scope: "northbound_metadata";
    tool_result_cache_scope: "metadata_only";
    recovery_checkpoint_scope: "audit_refs_only";
    durable_memory_enabled: false;
  };
  capabilities: {
    session_metadata: boolean;
    durable_session_store: false;
    durable_checkpoint_store: false;
    long_term_memory: false;
    automatic_replay: false;
    business_truth_write: false;
  };
  recovery_boundary: {
    checkpoint_read_route: "/v1/session/recovery/checkpoints/{checkpoint_id}";
    metadata_only: true;
    materialized_results_included: false;
    result_ref_reader_enabled: false;
    replay_executor_enabled: false;
    requires_confirmation_for_actions: true;
  };
  audit: AdvisoryOnlyAudit;
};

export type ToolMetadata = {
  tool_id: string;
  display_name: string;
  tool_type: "retrieval" | "candidate_builder" | string;
  project_scope: "radish" | "radishflow" | string;
  risk_level: "low" | "medium" | "high" | string;
  requires_confirmation_for_actions: boolean;
  interface: {
    input_schema_ref: string;
    output_schema_ref: string;
  };
  execution: {
    mode: "contract_only";
    execution_enabled: false;
    status: "disabled";
  };
  state_policy: {
    result_cache_mode: "metadata_only";
    durable_memory_enabled: false;
    writes_business_truth: false;
    materialized_result_read: false;
  };
};

export type ToolingMetadataResponse = {
  schema_version: 1;
  kind: "tooling_metadata";
  registry_id: "radishmind-tool-registry-v1";
  route: typeof PLATFORM_SESSION_TOOLING_ROUTES.toolsMetadata;
  api_boundary: PlatformMetadataBoundary;
  registry_policy: {
    execution_enabled: false;
    durable_memory_enabled: false;
    network_default: "disabled";
    default_timeout_seconds: number;
    max_retry_attempts: number;
  };
  tools: ToolMetadata[];
  blocked_action_route: typeof PLATFORM_SESSION_TOOLING_ROUTES.toolActions;
  audit: AdvisoryOnlyAudit;
};

export type ToolActionRequest = {
  tool_id: string;
  action: string;
  request_id?: string;
  session_id?: string;
  turn_id?: string;
  payload?: Record<string, unknown>;
};

export type ToolActionBlockedResponse = {
  schema_version: 1;
  kind: "tool_action_blocked_response";
  status: "blocked";
  route: typeof PLATFORM_SESSION_TOOLING_ROUTES.toolActions;
  request_id: string;
  action: {
    tool_id: string;
    action: string;
    known_tool: boolean;
    session_id: string;
    turn_id: string;
  };
  policy_decision: {
    decision: "blocked";
    primary_code: "TOOL_EXECUTOR_DISABLED" | "TOOL_NOT_REGISTERED";
    denial_codes: Array<
      "TOOL_EXECUTOR_DISABLED" | "TOOL_NOT_REGISTERED" | "CONFIRMATION_REQUIRED"
    >;
    requires_confirmation: boolean;
    reason: string;
  };
  execution: {
    execution_enabled: false;
    executed: false;
    status: "not_executed";
    duration_ms: null;
  };
  result: {
    result_ref: null;
    materialized_result_included: false;
    materialized_result_read: false;
    result_cache_mode: "metadata_only";
  };
  side_effects: {
    network_request_sent: false;
    durable_memory_written: false;
    writes_business_truth: false;
    automatic_replay_started: false;
  };
  audit: AdvisoryOnlyAudit;
};

export type AdvisoryOnlyAudit = {
  advisory_only: true;
  writes_business_truth: false;
  notes: string[];
};

export type ToolActionBlockedView = {
  toolId: string;
  action: string;
  canExecute: false;
  statusLabel: "blocked";
  primaryCode: ToolActionBlockedResponse["policy_decision"]["primary_code"];
  requiresConfirmation: boolean;
  noSideEffects: boolean;
};

export function isToolActionBlockedResponse(value: unknown): value is ToolActionBlockedResponse {
  if (!isRecord(value)) {
    return false;
  }
  return value.kind === "tool_action_blocked_response" && value.status === "blocked";
}

export function toToolActionBlockedView(response: ToolActionBlockedResponse): ToolActionBlockedView {
  return {
    toolId: response.action.tool_id,
    action: response.action.action,
    canExecute: false,
    statusLabel: "blocked",
    primaryCode: response.policy_decision.primary_code,
    requiresConfirmation: response.policy_decision.requires_confirmation,
    noSideEffects:
      response.execution.executed === false &&
      response.result.result_ref === null &&
      response.side_effects.network_request_sent === false &&
      response.side_effects.durable_memory_written === false &&
      response.side_effects.writes_business_truth === false &&
      response.side_effects.automatic_replay_started === false,
  };
}

export function listToolActionOptions(metadata: ToolingMetadataResponse): ToolActionRequest[] {
  return metadata.tools.map((tool) => ({
    tool_id: tool.tool_id,
    action: "execute",
  }));
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}
