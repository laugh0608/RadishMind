export const PLATFORM_OVERVIEW_ROUTE = "/v1/platform/overview" as const;

export type PlatformOverviewResponse = {
  schema_version: 1;
  kind: "platform_overview";
  stage: "P3 Local Product Shell" | string;
  route: typeof PLATFORM_OVERVIEW_ROUTE;
  service: {
    name: "radishmind-platform" | string;
    version: string;
    status: "ok" | string;
  };
  product_surface: {
    mode: "local_read_only_product_shell";
    implemented: true;
    ui_consumable: true;
    routes: string[];
  };
  models: {
    route: "/v1/models";
    detail_route: "/v1/models/{id}";
    inventory_kind: "bridge_backed_provider_profile_inventory";
    default_provider: string;
    default_profile: string;
    default_model: string;
    status: "ok" | "unavailable" | string;
    model_count?: number;
    provider_count?: number;
    profile_count?: number;
    active_profile_chain?: string[];
    selectable_model_ids?: string[];
    failure_code?: string;
    failure_boundary?: string;
    message?: string;
  };
  session_tooling: {
    stage: "P2 close candidate shell" | string;
    session_metadata_route: "/v1/session/metadata";
    tools_metadata_route: "/v1/tools/metadata";
    blocked_action_route: "/v1/tools/actions";
    metadata_only: true;
    tool_count: number;
    execution_enabled: false;
    tool_action_status: "blocked";
    requires_confirmation_path: "future_upper_layer_confirmation_flow" | string;
  };
  stop_lines: PlatformOverviewStopLines;
  audit: {
    advisory_only: true;
    writes_business_truth: false;
    notes: string[];
  };
};

export type PlatformOverviewStopLines = {
  real_executor_enabled: false;
  durable_store_enabled: false;
  confirmation_flow_connected: false;
  materialized_result_reader: false;
  long_term_memory_enabled: false;
  business_truth_write_enabled: false;
  automatic_replay_enabled: false;
  production_secret_backend_ready: false;
};

export type PlatformConsoleServiceStatusViewModel = {
  serviceName: string;
  version: string;
  status: string;
  stage: string;
  mode: "local_read_only_product_shell";
  overviewRoute: typeof PLATFORM_OVERVIEW_ROUTE;
  healthyForLocalConsole: boolean;
};

export type PlatformConsoleModelInventoryViewModel = {
  status: string;
  inventoryKind: "bridge_backed_provider_profile_inventory";
  modelsRoute: "/v1/models";
  detailRoute: "/v1/models/{id}";
  defaultProvider: string;
  defaultProfile: string;
  defaultModel: string;
  modelCount: number;
  providerCount: number;
  profileCount: number;
  selectableModelIds: string[];
  activeProfileChain: string[];
  canShowProfileSelector: boolean;
};

export type PlatformConsoleSessionToolingViewModel = {
  sessionMetadataRoute: "/v1/session/metadata";
  toolsMetadataRoute: "/v1/tools/metadata";
  blockedActionRoute: "/v1/tools/actions";
  metadataOnly: true;
  executionEnabled: false;
  actionStatusLabel: "blocked";
  toolCount: number;
  requiresConfirmationPath: string;
};

export type PlatformConsoleStopLineViewModel = {
  allStopLinesEnforced: boolean;
  blockedCapabilityIds: Array<keyof PlatformOverviewStopLines>;
  canExecuteActions: false;
  canUseDurableStore: false;
  canWriteBusinessTruth: false;
  canReplayAutomatically: false;
};

export type PlatformOverviewConsoleViewModel = {
  serviceStatus: PlatformConsoleServiceStatusViewModel;
  modelInventory: PlatformConsoleModelInventoryViewModel;
  sessionTooling: PlatformConsoleSessionToolingViewModel;
  stopLines: PlatformConsoleStopLineViewModel;
};

export function isPlatformOverviewResponse(value: unknown): value is PlatformOverviewResponse {
  if (!isRecord(value)) {
    return false;
  }
  return value.kind === "platform_overview" && value.route === PLATFORM_OVERVIEW_ROUTE;
}

export function toPlatformOverviewConsoleViewModel(
  overview: PlatformOverviewResponse,
): PlatformOverviewConsoleViewModel {
  return {
    serviceStatus: toPlatformConsoleServiceStatusViewModel(overview),
    modelInventory: toPlatformConsoleModelInventoryViewModel(overview),
    sessionTooling: toPlatformConsoleSessionToolingViewModel(overview),
    stopLines: toPlatformConsoleStopLineViewModel(overview.stop_lines),
  };
}

export function toPlatformConsoleServiceStatusViewModel(
  overview: PlatformOverviewResponse,
): PlatformConsoleServiceStatusViewModel {
  return {
    serviceName: overview.service.name,
    version: overview.service.version,
    status: overview.service.status,
    stage: overview.stage,
    mode: overview.product_surface.mode,
    overviewRoute: overview.route,
    healthyForLocalConsole:
      overview.service.status === "ok" &&
      overview.product_surface.implemented === true &&
      overview.product_surface.ui_consumable === true,
  };
}

export function toPlatformConsoleModelInventoryViewModel(
  overview: PlatformOverviewResponse,
): PlatformConsoleModelInventoryViewModel {
  const models = overview.models;
  const selectableModelIds = models.selectable_model_ids ?? [];
  return {
    status: models.status,
    inventoryKind: models.inventory_kind,
    modelsRoute: models.route,
    detailRoute: models.detail_route,
    defaultProvider: models.default_provider,
    defaultProfile: models.default_profile,
    defaultModel: models.default_model,
    modelCount: models.model_count ?? 0,
    providerCount: models.provider_count ?? 0,
    profileCount: models.profile_count ?? 0,
    selectableModelIds,
    activeProfileChain: models.active_profile_chain ?? [],
    canShowProfileSelector: models.status === "ok" && selectableModelIds.length > 0,
  };
}

export function toPlatformConsoleSessionToolingViewModel(
  overview: PlatformOverviewResponse,
): PlatformConsoleSessionToolingViewModel {
  return {
    sessionMetadataRoute: overview.session_tooling.session_metadata_route,
    toolsMetadataRoute: overview.session_tooling.tools_metadata_route,
    blockedActionRoute: overview.session_tooling.blocked_action_route,
    metadataOnly: true,
    executionEnabled: false,
    actionStatusLabel: "blocked",
    toolCount: overview.session_tooling.tool_count,
    requiresConfirmationPath: overview.session_tooling.requires_confirmation_path,
  };
}

export function toPlatformConsoleStopLineViewModel(
  stopLines: PlatformOverviewStopLines,
): PlatformConsoleStopLineViewModel {
  return {
    allStopLinesEnforced: allPlatformOverviewStopLinesEnforced(stopLines),
    blockedCapabilityIds: Object.entries(stopLines)
      .filter(([, enabled]) => enabled === false)
      .map(([capability]) => capability) as Array<keyof PlatformOverviewStopLines>,
    canExecuteActions: false,
    canUseDurableStore: false,
    canWriteBusinessTruth: false,
    canReplayAutomatically: false,
  };
}

export function allPlatformOverviewStopLinesEnforced(stopLines: PlatformOverviewStopLines): boolean {
  return (
    stopLines.real_executor_enabled === false &&
    stopLines.durable_store_enabled === false &&
    stopLines.confirmation_flow_connected === false &&
    stopLines.materialized_result_reader === false &&
    stopLines.long_term_memory_enabled === false &&
    stopLines.business_truth_write_enabled === false &&
    stopLines.automatic_replay_enabled === false
  );
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}
