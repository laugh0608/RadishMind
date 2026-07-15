import type { ModelGatewayOverviewStatus, ModelGatewayOverviewViewModel } from "./modelGatewayOverview";
import type { ControlPlaneReadShellViewModel } from "./readShell";

export type ModelGatewayRouteEvidenceStatus = ModelGatewayOverviewStatus;

export type ModelGatewayRouteEvidenceSource = {
  overview: ModelGatewayOverviewViewModel;
  readShell: ControlPlaneReadShellViewModel;
};

export type ModelGatewayProviderProfileEvidence = {
  profileId: string;
  providerId: string;
  profileName: string;
  resolvedModel: string;
  credentialState: string;
  deploymentMode: string;
  authMode: string;
  streaming: "enabled" | "disabled";
  status: ModelGatewayRouteEvidenceStatus;
  routes: string[];
  protocols: string[];
  capabilities: string[];
  evidenceRef: string;
  summary: string;
};

export type ModelGatewayRouteBindingEvidence = {
  routeId: string;
  path: string;
  protocol: string;
  selectedProfileId: string;
  selectionSource: string;
  metadataFields: string[];
  status: ModelGatewayRouteEvidenceStatus;
  evidenceRef: string;
  summary: string;
};

export type ModelGatewaySelectionCaseEvidence = {
  caseId: string;
  label: string;
  requestedInput: string;
  result: string;
  failureCode: string;
  fallbackAllowed: boolean;
  defaultFastBaseline: boolean;
  status: ModelGatewayRouteEvidenceStatus;
  evidenceRef: string;
  summary: string;
};

export type ModelGatewayRuntimeGuardEvidence = {
  guardId: string;
  label: string;
  status: ModelGatewayRouteEvidenceStatus;
  sourceRef: string;
  enforcedBy: string;
  summary: string;
};

export type ModelGatewayRouteEvidenceViewModel = {
  pageId: "model-gateway-route-evidence-offline";
  sourcePageIds: string[];
  routeEvidenceMode: "offline_provider_profile_route_evidence";
  gatewayOverviewPageId: ModelGatewayOverviewViewModel["pageId"];
  tenantRef: string;
  requestId: string;
  auditRef: string;
  routeNarrative: string;
  profiles: ModelGatewayProviderProfileEvidence[];
  routeBindings: ModelGatewayRouteBindingEvidence[];
  selectionCases: ModelGatewaySelectionCaseEvidence[];
  runtimeGuards: ModelGatewayRuntimeGuardEvidence[];
  canRenderRouteEvidenceDetail: boolean;
  canRequestLiveProviderHealth: false;
  canExecuteProviderCall: false;
  canExecuteFallback: false;
  canResolveProductionSecret: false;
  canChangeProfileChain: false;
  canWriteCostRecord: false;
  canUseProductionGateway: false;
};

export function buildModelGatewayRouteEvidenceViewModel(
  source: ModelGatewayRouteEvidenceSource,
): ModelGatewayRouteEvidenceViewModel {
  const profiles = buildProfiles();
  const routeBindings = buildRouteBindings(source.overview);
  const selectionCases = buildSelectionCases();
  const runtimeGuards = buildRuntimeGuards(source);

  return {
    pageId: "model-gateway-route-evidence-offline",
    sourcePageIds: [
      source.overview.pageId,
      "provider-capability-matrix-v1",
      "provider-health-smoke-v1",
      "provider-selection-policy-v1",
      "provider-retry-fallback-policy-v1",
      "provider-runtime-docs-refresh",
      "platform-ops-smoke",
      "control-plane-read-shared-shell",
    ],
    routeEvidenceMode: "offline_provider_profile_route_evidence",
    gatewayOverviewPageId: source.overview.pageId,
    tenantRef: source.overview.tenantRef,
    requestId: source.overview.requestId,
    auditRef: source.overview.auditRef,
    routeNarrative: buildRouteNarrative(source, profiles, routeBindings),
    profiles,
    routeBindings,
    selectionCases,
    runtimeGuards,
    canRenderRouteEvidenceDetail:
      source.overview.canRenderModelGatewayOverview &&
      source.readShell.usesCanonicalRoutes &&
      profiles.length >= 5 &&
      routeBindings.length >= 5 &&
      selectionCases.length >= 7 &&
      runtimeGuards.length >= 6,
    canRequestLiveProviderHealth: false,
    canExecuteProviderCall: false,
    canExecuteFallback: false,
    canResolveProductionSecret: false,
    canChangeProfileChain: false,
    canWriteCostRecord: false,
    canUseProductionGateway: false,
  };
}

function buildRouteNarrative(
  source: ModelGatewayRouteEvidenceSource,
  profiles: ModelGatewayProviderProfileEvidence[],
  routeBindings: ModelGatewayRouteBindingEvidence[],
): string {
  const configuredProfiles = profiles.filter((profile) => profile.credentialState === "configured").length;
  const blockedProfiles = profiles.filter((profile) => profile.status === "blocked").length;
  return `${profiles.length} provider/profile records, ${configuredProfiles} configured credential states, ${blockedProfiles} blocked credential cases, and ${routeBindings.length} northbound route bindings are grouped under ${source.overview.pageId} for offline audit.`;
}

function buildProfiles(): ModelGatewayProviderProfileEvidence[] {
  const chatRoutes = ["/v1/models", "/v1/chat/completions"];
  return [
    {
      profileId: "profile:anyrouter",
      providerId: "openai-compatible",
      profileName: "anyrouter",
      resolvedModel: "deepseek-chat",
      credentialState: "configured",
      deploymentMode: "remote_api",
      authMode: "profile",
      streaming: "enabled",
      status: "ready",
      routes: chatRoutes,
      protocols: ["chat.completions"],
      capabilities: ["chat", "streaming", "per-request timeout", "caller-managed retry"],
      evidenceRef: "provider-selection-policy-v1",
      summary: "OpenAI-compatible profile inventory can be selected by profile id and keeps provider/profile metadata attached to the request.",
    },
    {
      profileId: "provider:huggingface:profile:hf-chat",
      providerId: "huggingface",
      profileName: "hf-chat",
      resolvedModel: "meta-llama/Meta-Llama-3.1-8B-Instruct",
      credentialState: "configured",
      deploymentMode: "remote_api",
      authMode: "profile",
      streaming: "enabled",
      status: "ready",
      routes: chatRoutes,
      protocols: ["chat.completions"],
      capabilities: ["chat", "streaming", "per-request timeout", "caller-managed retry"],
      evidenceRef: "provider-selection-policy-v1",
      summary: "Provider-qualified profile selection resolves to the configured Hugging Face profile without provider fallback.",
    },
    {
      profileId: "provider:ollama:profile:local",
      providerId: "ollama",
      profileName: "local",
      resolvedModel: "local daemon profile",
      credentialState: "optional_missing",
      deploymentMode: "local_daemon",
      authMode: "optional",
      streaming: "enabled",
      status: "offline_only",
      routes: chatRoutes,
      protocols: ["chat.completions"],
      capabilities: ["chat", "streaming", "local cost profile", "caller-managed retry"],
      evidenceRef: "provider-health-smoke-v1",
      summary: "Ollama local profile can be represented as config-level inventory; live daemon reachability remains a separate manual health task.",
    },
    {
      profileId: "profile:missing",
      providerId: "openai-compatible",
      profileName: "missing",
      resolvedModel: "missing credential example",
      credentialState: "missing",
      deploymentMode: "remote_api",
      authMode: "profile",
      streaming: "enabled",
      status: "blocked",
      routes: chatRoutes,
      protocols: ["chat.completions"],
      capabilities: ["chat", "streaming", "credential required"],
      evidenceRef: "provider-health-smoke-v1",
      summary: "Config-level inventory can expose missing credentials, but the route must remain blocked instead of silently swapping profiles.",
    },
    {
      profileId: "provider:huggingface:profile:hf-missing",
      providerId: "huggingface",
      profileName: "hf-missing",
      resolvedModel: "missing credential example",
      credentialState: "missing",
      deploymentMode: "remote_api",
      authMode: "profile",
      streaming: "enabled",
      status: "blocked",
      routes: chatRoutes,
      protocols: ["chat.completions"],
      capabilities: ["chat", "streaming", "credential required"],
      evidenceRef: "provider-health-smoke-v1",
      summary: "The Hugging Face missing-credential example is a provider health signal only, not a production outage or fallback trigger.",
    },
  ];
}

function buildRouteBindings(overview: ModelGatewayOverviewViewModel): ModelGatewayRouteBindingEvidence[] {
  return [
    {
      routeId: "models_inventory",
      path: "/v1/models",
      protocol: "model inventory",
      selectedProfileId: "provider profile inventory",
      selectionSource: "provider_profile_inventory",
      metadataFields: [
        "capabilities",
        "northbound_routes",
        "northbound_protocols",
        "credential_state",
        "deployment_mode",
        "auth_mode",
        "streaming",
      ],
      status: "offline_only",
      evidenceRef: "provider-runtime-docs-refresh",
      summary: "Inventory exposes discoverability metadata only; it is not profile provisioning, credential verification, or provider health execution.",
    },
    {
      routeId: "model_lookup",
      path: "/v1/models/{id}",
      protocol: "model lookup",
      selectedProfileId: "profile:anyrouter",
      selectionSource: "lookup_aliases",
      metadataFields: ["source", "inventory_kind", "selection", "lookup_aliases", "MODEL_NOT_FOUND"],
      status: "offline_only",
      evidenceRef: "provider-selection-policy-v1",
      summary: "Lookup accepts stable profile ids and provider-qualified aliases; unknown ids fail as MODEL_NOT_FOUND without profile replacement.",
    },
    {
      routeId: "chat_completions",
      path: "/v1/chat/completions",
      protocol: "chat.completions",
      selectedProfileId: "profile:anyrouter",
      selectionSource: "requested_profile_model",
      metadataFields: [
        "selected_provider",
        "selected_provider_profile",
        "selected_model",
        "upstream_model",
        "selection_source",
        "request_id",
      ],
      status: "ready",
      evidenceRef: overview.auditRef,
      summary: "Chat requests can carry provider/profile selection metadata through context.northbound for audit and diagnostics.",
    },
    {
      routeId: "responses",
      path: "/v1/responses",
      protocol: "responses compatibility",
      selectedProfileId: "configured default or requested model",
      selectionSource: "northbound route request",
      metadataFields: ["request_id", "route", "provider_profile", "selection_source", "failure_boundary"],
      status: "review_required",
      evidenceRef: overview.auditRef,
      summary: "Responses compatibility is visible at the gateway surface; profile capability metadata must still be checked before execution.",
    },
    {
      routeId: "messages",
      path: "/v1/messages",
      protocol: "messages compatibility",
      selectedProfileId: "configured default or requested model",
      selectionSource: "northbound route request",
      metadataFields: ["request_id", "route", "provider_profile", "selection_source", "failure_boundary"],
      status: "review_required",
      evidenceRef: overview.auditRef,
      summary: "Messages compatibility stays a northbound adapter surface and does not create a separate product protocol or provider fallback chain.",
    },
  ];
}

function buildSelectionCases(): ModelGatewaySelectionCaseEvidence[] {
  return [
    {
      caseId: "profile-model-inventory",
      label: "Profile model inventory",
      requestedInput: "model=profile:anyrouter",
      result: "openai-compatible / anyrouter / deepseek-chat",
      failureCode: "none",
      fallbackAllowed: false,
      defaultFastBaseline: true,
      status: "ready",
      evidenceRef: "provider-selection-policy-v1",
      summary: "Requested profile model resolves through inventory and keeps credential_state=configured attached to selection metadata.",
    },
    {
      caseId: "provider-alias-active-profile",
      label: "Provider alias active profile",
      requestedInput: "model=provider:huggingface",
      result: "provider:huggingface:profile:hf-chat",
      failureCode: "none",
      fallbackAllowed: false,
      defaultFastBaseline: true,
      status: "ready",
      evidenceRef: "provider-selection-policy-v1",
      summary: "Provider alias resolves to the active configured profile; it is explicit selection, not a fallback chain.",
    },
    {
      caseId: "unknown-concrete-model-runtime-override",
      label: "Unknown concrete model",
      requestedInput: "model=unlisted-model",
      result: "runtime_override_not_inventory_match",
      failureCode: "none",
      fallbackAllowed: false,
      defaultFastBaseline: true,
      status: "review_required",
      evidenceRef: "provider-selection-policy-v1",
      summary: "Unknown concrete model can remain a runtime override, but it is not represented as inventory evidence.",
    },
    {
      caseId: "unknown-explicit-profile-no-fallback",
      label: "Unknown explicit profile",
      requestedInput: "provider=huggingface profile=missing-profile",
      result: "unknown_profile_no_fallback",
      failureCode: "none",
      fallbackAllowed: false,
      defaultFastBaseline: true,
      status: "blocked",
      evidenceRef: "provider-selection-policy-v1",
      summary: "Explicit unknown provider profile is preserved in metadata and must not be replaced by the active profile.",
    },
    {
      caseId: "model-detail-unknown",
      label: "Unknown model lookup",
      requestedInput: "GET /v1/models/does-not-exist",
      result: "not found",
      failureCode: "MODEL_NOT_FOUND",
      fallbackAllowed: false,
      defaultFastBaseline: true,
      status: "blocked",
      evidenceRef: "provider-selection-policy-v1",
      summary: "Model lookup has an explicit negative boundary, so callers can distinguish unknown inventory from provider failure.",
    },
    {
      caseId: "credential-missing",
      label: "Credential missing",
      requestedInput: "credential_state=missing",
      result: "config_blocked_credential_missing",
      failureCode: "config_blocked_credential_missing",
      fallbackAllowed: false,
      defaultFastBaseline: true,
      status: "blocked",
      evidenceRef: "provider-health-smoke-v1",
      summary: "Missing credential is a configuration block and cannot trigger implicit provider switching.",
    },
    {
      caseId: "unsupported-schema-tool-capability",
      label: "Unsupported capability",
      requestedInput: "json_schema_output / tool_calling / image input-output",
      result: "unsupported_capability",
      failureCode: "unsupported_capability",
      fallbackAllowed: false,
      defaultFastBaseline: true,
      status: "blocked",
      evidenceRef: "provider-selection-policy-v1",
      summary: "Unsupported capability is surfaced as a capability mismatch; the model gateway detail does not route around it.",
    },
    {
      caseId: "timeout",
      label: "Provider timeout",
      requestedInput: "provider call timeout",
      result: "provider_timeout",
      failureCode: "provider_timeout",
      fallbackAllowed: false,
      defaultFastBaseline: false,
      status: "review_required",
      evidenceRef: "provider-selection-policy-v1",
      summary: "Timeout classification is available at the provider call boundary, but live timeout probe and retry execution are out of scope.",
    },
  ];
}

function buildRuntimeGuards(source: ModelGatewayRouteEvidenceSource): ModelGatewayRuntimeGuardEvidence[] {
  return [
    {
      guardId: "live_health_manual_only",
      label: "Live health is manual",
      status: "locked",
      sourceRef: "provider-health-smoke-v1",
      enforcedBy: "optional_live_health=manual_only_future_slice",
      summary: "Default checks cover mock runtime and config inventory only; external provider reachability is not enabled by default.",
    },
    {
      guardId: "no_implicit_fallback",
      label: "No implicit fallback",
      status: "locked",
      sourceRef: "provider-retry-fallback-policy-v1",
      enforcedBy: "fallback_policy=disabled",
      summary: "Unknown profile, missing credential, unsupported capability, schema failure, tool execution, writeback, and replay paths cannot trigger fallback.",
    },
    {
      guardId: "caller_managed_retry",
      label: "Caller-managed retry",
      status: "locked",
      sourceRef: "provider-retry-fallback-policy-v1",
      enforcedBy: "max_attempts_by_platform=1",
      summary: "The platform records retry metadata but does not automatically retry, queue background retry, or retry after confirmation.",
    },
    {
      guardId: "secret_value_hidden",
      label: "Secret value hidden",
      status: "locked",
      sourceRef: "production-secret-backend-contract",
      enforcedBy: "credential_state and field_sources only",
      summary: "Profile evidence can show configured, missing, optional_missing, or not_required states without rendering raw credential values.",
    },
    {
      guardId: "request_observability",
      label: "Request observability",
      status: "offline_only",
      sourceRef: "docs/contracts/service-api.md",
      enforcedBy: "request_id plus route/provider/profile fields",
      summary: "Route, status, latency, provider/profile, selected model, selection source, error code, and failure boundary are the audit fields for future runs.",
    },
    {
      guardId: "overview_dependency",
      label: "Overview dependency",
      status: source.overview.canRenderModelGatewayOverview ? "ready" : "blocked",
      sourceRef: source.overview.pageId,
      enforcedBy: "overview evidence must render first",
      summary: "Route evidence detail remains subordinate to the Model Gateway Overview and does not create a second product source of truth.",
    },
  ];
}
