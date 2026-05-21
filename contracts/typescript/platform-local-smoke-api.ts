export const PLATFORM_LOCAL_SMOKE_ROUTE = "/v1/platform/local-smoke" as const;

export type PlatformLocalSmokeResponse = {
  schema_version: 1;
  kind: "platform_local_smoke";
  stage: "P3 Local Product Shell / Ops Surface" | string;
  route: typeof PLATFORM_LOCAL_SMOKE_ROUTE;
  summary: {
    status: "ok" | "degraded" | string;
    local_console_ready: boolean;
    read_only: true;
    default_ports: {
      platform: 7000;
      console: 4000;
    };
  };
  checks: {
    healthz: {
      route: "/healthz";
      status: "ok" | string;
      readable: boolean;
    };
    overview: {
      route: "/v1/platform/overview";
      readable: boolean;
      contract_kind: "platform_overview";
      ui_consumable: boolean;
      product_routes: string[];
    };
    model_inventory: {
      route: "/v1/models";
      detail_route: "/v1/models/{id}";
      status: "ok" | "unavailable" | string;
      readable: boolean;
      inventory_kind: "bridge_backed_provider_profile_inventory";
      model_count?: number;
      provider_count?: number;
      profile_count?: number;
      failure_code?: string | null;
    };
    session_tooling: {
      session_metadata_route: "/v1/session/metadata";
      tools_metadata_route: "/v1/tools/metadata";
      blocked_action_route: "/v1/tools/actions";
      session_metadata_readable: boolean;
      tools_metadata_readable: boolean;
      tool_count: number;
      metadata_only: true;
      execution_enabled: false;
      blocked_action_status: "blocked";
      blocked_action_primary_code: "TOOL_EXECUTOR_DISABLED" | string;
      requires_confirmation: boolean;
      blocked_action_no_side_effects: true;
    };
    local_console: {
      frontend_origin_default: "http://127.0.0.1:4000";
      backend_url_default: "http://127.0.0.1:7000";
      allowed_cors_origins: string[];
      cors_preflight_methods: Array<"GET" | "POST" | "OPTIONS" | string>;
      cors_scope: "local_dev_only";
    };
  };
  stop_lines: PlatformLocalSmokeStopLines;
  failure_hints: Array<{
    code: "PORT_IN_USE" | "CORS_ORIGIN_NOT_ALLOWED" | "ERR_UNSAFE_PORT" | string;
    message: string;
  }>;
  audit: {
    advisory_only: true;
    writes_business_truth: false;
    notes: string[];
  };
};

export type PlatformLocalSmokeStopLines = {
  real_executor_enabled: false;
  durable_store_enabled: false;
  confirmation_flow_connected: false;
  materialized_result_reader: false;
  long_term_memory_enabled: false;
  business_truth_write_enabled: false;
  automatic_replay_enabled: false;
  production_secret_backend_ready: false;
};

export type PlatformLocalSmokeReadinessViewModel = {
  status: string;
  localConsoleReady: boolean;
  readOnly: true;
  healthzOk: boolean;
  overviewContractReadable: boolean;
  modelInventoryReadable: boolean;
  sessionToolingMetadataReadable: boolean;
  blockedActionNoSideEffects: boolean;
  allowedCorsOrigins: string[];
  unsafePortHintPresent: boolean;
  allStopLinesEnforced: boolean;
  canExecuteActions: false;
  canUseDurableStore: false;
  canWriteBusinessTruth: false;
  canReplayAutomatically: false;
};

export function isPlatformLocalSmokeResponse(value: unknown): value is PlatformLocalSmokeResponse {
  if (!isRecord(value)) {
    return false;
  }
  return value.kind === "platform_local_smoke" && value.route === PLATFORM_LOCAL_SMOKE_ROUTE;
}

export function toPlatformLocalSmokeReadinessViewModel(
  response: PlatformLocalSmokeResponse,
): PlatformLocalSmokeReadinessViewModel {
  return {
    status: response.summary.status,
    localConsoleReady: response.summary.local_console_ready,
    readOnly: true,
    healthzOk: response.checks.healthz.status === "ok" && response.checks.healthz.readable === true,
    overviewContractReadable:
      response.checks.overview.contract_kind === "platform_overview" &&
      response.checks.overview.readable === true,
    modelInventoryReadable:
      response.checks.model_inventory.status === "ok" &&
      response.checks.model_inventory.readable === true,
    sessionToolingMetadataReadable:
      response.checks.session_tooling.session_metadata_readable === true &&
      response.checks.session_tooling.tools_metadata_readable === true,
    blockedActionNoSideEffects:
      response.checks.session_tooling.blocked_action_status === "blocked" &&
      response.checks.session_tooling.blocked_action_no_side_effects === true,
    allowedCorsOrigins: response.checks.local_console.allowed_cors_origins,
    unsafePortHintPresent: response.failure_hints.some((hint) => hint.code === "ERR_UNSAFE_PORT"),
    allStopLinesEnforced: allPlatformLocalSmokeStopLinesEnforced(response.stop_lines),
    canExecuteActions: false,
    canUseDurableStore: false,
    canWriteBusinessTruth: false,
    canReplayAutomatically: false,
  };
}

export function allPlatformLocalSmokeStopLinesEnforced(stopLines: PlatformLocalSmokeStopLines): boolean {
  return (
    stopLines.real_executor_enabled === false &&
    stopLines.durable_store_enabled === false &&
    stopLines.confirmation_flow_connected === false &&
    stopLines.materialized_result_reader === false &&
    stopLines.long_term_memory_enabled === false &&
    stopLines.business_truth_write_enabled === false &&
    stopLines.automatic_replay_enabled === false &&
    stopLines.production_secret_backend_ready === false
  );
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}
