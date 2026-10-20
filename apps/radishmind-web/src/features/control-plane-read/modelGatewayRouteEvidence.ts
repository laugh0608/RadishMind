import type { EvidenceProjectionCopy, EvidenceProjectionMessage } from "./evidenceProjectionCopy.ts";
import type { ModelGatewayOverviewStatus, ModelGatewayOverviewViewModel } from "./modelGatewayOverview";
import type { ControlPlaneReadShellViewModel } from "./readShell";

export type ModelGatewayRouteEvidenceStatus = ModelGatewayOverviewStatus;

export type ModelGatewayRouteEvidenceSource = {
  overview: ModelGatewayOverviewViewModel;
  readShell: ControlPlaneReadShellViewModel;
};

export type ModelGatewayProviderProfileEvidence = EvidenceProjectionCopy & {
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

export type ModelGatewayRouteBindingEvidence = EvidenceProjectionCopy & {
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

export type ModelGatewaySelectionCaseEvidence = EvidenceProjectionCopy & {
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

export type ModelGatewayRuntimeGuardEvidence = EvidenceProjectionCopy & {
  guardId: string;
  label: string;
  status: ModelGatewayRouteEvidenceStatus;
  sourceRef: string;
  enforcedBy: string;
  summary: string;
};

export type ModelGatewayRouteEvidenceViewModel = EvidenceProjectionCopy & {
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
    routeNarrativeMessage: buildRouteNarrativeMessage(source, profiles, routeBindings),
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
      summaryMessage: { key: "gatewayRouteEvidence_profile_anyrouter_summary" as const },
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
      summaryMessage: { key: "gatewayRouteEvidence_provider_huggingface_profile_hf_chat_summary" as const },
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
      summaryMessage: { key: "gatewayRouteEvidence_provider_ollama_profile_local_summary" as const },
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
      summaryMessage: { key: "gatewayRouteEvidence_profile_missing_summary" as const },
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
      summaryMessage: { key: "gatewayRouteEvidence_provider_huggingface_profile_hf_missing_summary" as const },
    },
  ];
}

function buildRouteBindings(overview: ModelGatewayOverviewViewModel): ModelGatewayRouteBindingEvidence[] {
  return [
    {
      routeId: "models_inventory",
      path: "/v1/models",
      protocol: "model inventory",
      protocolMessage: { key: "gatewayRouteEvidence_models_inventory_protocol" as const },
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
      summaryMessage: { key: "gatewayRouteEvidence_models_inventory_summary" as const },
    },
    {
      routeId: "model_lookup",
      path: "/v1/models/{id}",
      protocol: "model lookup",
      protocolMessage: { key: "gatewayRouteEvidence_model_lookup_protocol" as const },
      selectedProfileId: "profile:anyrouter",
      selectionSource: "lookup_aliases",
      metadataFields: ["source", "inventory_kind", "selection", "lookup_aliases", "MODEL_NOT_FOUND"],
      status: "offline_only",
      evidenceRef: "provider-selection-policy-v1",
      summary: "Lookup accepts stable profile ids and provider-qualified aliases; unknown ids fail as MODEL_NOT_FOUND without profile replacement.",
      summaryMessage: { key: "gatewayRouteEvidence_model_lookup_summary" as const },
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
      summaryMessage: { key: "gatewayRouteEvidence_chat_completions_summary" as const },
    },
    {
      routeId: "responses",
      path: "/v1/responses",
      protocol: "responses compatibility",
      protocolMessage: { key: "gatewayRouteEvidence_responses_protocol" as const },
      selectedProfileId: "configured default or requested model",
      selectionSource: "northbound route request",
      metadataFields: ["request_id", "route", "provider_profile", "selection_source", "failure_boundary"],
      status: "review_required",
      evidenceRef: overview.auditRef,
      summary: "Responses compatibility is visible at the gateway surface; profile capability metadata must still be checked before execution.",
      summaryMessage: { key: "gatewayRouteEvidence_responses_summary" as const },
    },
    {
      routeId: "messages",
      path: "/v1/messages",
      protocol: "messages compatibility",
      protocolMessage: { key: "gatewayRouteEvidence_messages_protocol" as const },
      selectedProfileId: "configured default or requested model",
      selectionSource: "northbound route request",
      metadataFields: ["request_id", "route", "provider_profile", "selection_source", "failure_boundary"],
      status: "review_required",
      evidenceRef: overview.auditRef,
      summary: "Messages compatibility stays a northbound adapter surface and does not create a separate product protocol or provider fallback chain.",
      summaryMessage: { key: "gatewayRouteEvidence_messages_summary" as const },
    },
  ];
}

function buildSelectionCases(): ModelGatewaySelectionCaseEvidence[] {
  return [
    {
      caseId: "profile-model-inventory",
      label: "Profile model inventory",
      labelMessage: { key: "gatewayRouteEvidence_profile_model_inventory_label" as const },
      requestedInput: "model=profile:anyrouter",
      result: "openai-compatible / anyrouter / deepseek-chat",
      failureCode: "none",
      fallbackAllowed: false,
      defaultFastBaseline: true,
      status: "ready",
      evidenceRef: "provider-selection-policy-v1",
      summary: "Requested profile model resolves through inventory and keeps credential_state=configured attached to selection metadata.",
      summaryMessage: { key: "gatewayRouteEvidence_profile_model_inventory_summary" as const },
    },
    {
      caseId: "provider-alias-active-profile",
      label: "Provider alias active profile",
      labelMessage: { key: "gatewayRouteEvidence_provider_alias_active_profile_label" as const },
      requestedInput: "model=provider:huggingface",
      result: "provider:huggingface:profile:hf-chat",
      failureCode: "none",
      fallbackAllowed: false,
      defaultFastBaseline: true,
      status: "ready",
      evidenceRef: "provider-selection-policy-v1",
      summary: "Provider alias resolves to the active configured profile; it is explicit selection, not a fallback chain.",
      summaryMessage: { key: "gatewayRouteEvidence_provider_alias_active_profile_summary" as const },
    },
    {
      caseId: "unknown-concrete-model-runtime-override",
      label: "Unknown concrete model",
      labelMessage: { key: "gatewayRouteEvidence_unknown_concrete_model_runtime_override_label" as const },
      requestedInput: "model=unlisted-model",
      result: "runtime_override_not_inventory_match",
      failureCode: "none",
      fallbackAllowed: false,
      defaultFastBaseline: true,
      status: "review_required",
      evidenceRef: "provider-selection-policy-v1",
      summary: "Unknown concrete model can remain a runtime override, but it is not represented as inventory evidence.",
      summaryMessage: { key: "gatewayRouteEvidence_unknown_concrete_model_runtime_override_summary" as const },
    },
    {
      caseId: "unknown-explicit-profile-no-fallback",
      label: "Unknown explicit profile",
      labelMessage: { key: "gatewayRouteEvidence_unknown_explicit_profile_no_fallback_label" as const },
      requestedInput: "provider=huggingface profile=missing-profile",
      result: "unknown_profile_no_fallback",
      failureCode: "none",
      fallbackAllowed: false,
      defaultFastBaseline: true,
      status: "blocked",
      evidenceRef: "provider-selection-policy-v1",
      summary: "Explicit unknown provider profile is preserved in metadata and must not be replaced by the active profile.",
      summaryMessage: { key: "gatewayRouteEvidence_unknown_explicit_profile_no_fallback_summary" as const },
    },
    {
      caseId: "model-detail-unknown",
      label: "Unknown model lookup",
      labelMessage: { key: "gatewayRouteEvidence_model_detail_unknown_label" as const },
      requestedInput: "GET /v1/models/does-not-exist",
      result: "not found",
      failureCode: "MODEL_NOT_FOUND",
      fallbackAllowed: false,
      defaultFastBaseline: true,
      status: "blocked",
      evidenceRef: "provider-selection-policy-v1",
      summary: "Model lookup has an explicit negative boundary, so callers can distinguish unknown inventory from provider failure.",
      summaryMessage: { key: "gatewayRouteEvidence_model_detail_unknown_summary" as const },
    },
    {
      caseId: "credential-missing",
      label: "Credential missing",
      labelMessage: { key: "gatewayRouteEvidence_credential_missing_label" as const },
      requestedInput: "credential_state=missing",
      result: "config_blocked_credential_missing",
      failureCode: "config_blocked_credential_missing",
      fallbackAllowed: false,
      defaultFastBaseline: true,
      status: "blocked",
      evidenceRef: "provider-health-smoke-v1",
      summary: "Missing credential is a configuration block and cannot trigger implicit provider switching.",
      summaryMessage: { key: "gatewayRouteEvidence_credential_missing_summary" as const },
    },
    {
      caseId: "unsupported-schema-tool-capability",
      label: "Unsupported capability",
      labelMessage: { key: "gatewayRouteEvidence_unsupported_schema_tool_capability_label" as const },
      requestedInput: "json_schema_output / tool_calling / image input-output",
      result: "unsupported_capability",
      failureCode: "unsupported_capability",
      fallbackAllowed: false,
      defaultFastBaseline: true,
      status: "blocked",
      evidenceRef: "provider-selection-policy-v1",
      summary: "Unsupported capability is surfaced as a capability mismatch; the model gateway detail does not route around it.",
      summaryMessage: { key: "gatewayRouteEvidence_unsupported_schema_tool_capability_summary" as const },
    },
    {
      caseId: "timeout",
      label: "Provider timeout",
      labelMessage: { key: "gatewayRouteEvidence_timeout_label" as const },
      requestedInput: "provider call timeout",
      result: "provider_timeout",
      failureCode: "provider_timeout",
      fallbackAllowed: false,
      defaultFastBaseline: false,
      status: "review_required",
      evidenceRef: "provider-selection-policy-v1",
      summary: "Timeout classification is available at the provider call boundary, but live timeout probe and retry execution are out of scope.",
      summaryMessage: { key: "gatewayRouteEvidence_timeout_summary" as const },
    },
  ];
}

function buildRuntimeGuards(source: ModelGatewayRouteEvidenceSource): ModelGatewayRuntimeGuardEvidence[] {
  return [
    {
      guardId: "live_health_manual_only",
      label: "Live health is manual",
      labelMessage: { key: "gatewayRouteEvidence_live_health_manual_only_label" as const },
      status: "locked",
      sourceRef: "provider-health-smoke-v1",
      enforcedBy: "optional_live_health=manual_only_future_slice",

      summary: "Default checks cover mock runtime and config inventory only; external provider reachability is not enabled by default.",
      summaryMessage: { key: "gatewayRouteEvidence_live_health_manual_only_summary" as const },
    },
    {
      guardId: "no_implicit_fallback",
      label: "No implicit fallback",
      labelMessage: { key: "gatewayRouteEvidence_no_implicit_fallback_label" as const },
      status: "locked",
      sourceRef: "provider-retry-fallback-policy-v1",
      enforcedBy: "fallback_policy=disabled",

      summary: "Unknown profile, missing credential, unsupported capability, schema failure, tool execution, writeback, and replay paths cannot trigger fallback.",
      summaryMessage: { key: "gatewayRouteEvidence_no_implicit_fallback_summary" as const },
    },
    {
      guardId: "caller_managed_retry",
      label: "Caller-managed retry",
      labelMessage: { key: "gatewayRouteEvidence_caller_managed_retry_label" as const },
      status: "locked",
      sourceRef: "provider-retry-fallback-policy-v1",
      enforcedBy: "max_attempts_by_platform=1",

      summary: "The platform records retry metadata but does not automatically retry, queue background retry, or retry after confirmation.",
      summaryMessage: { key: "gatewayRouteEvidence_caller_managed_retry_summary" as const },
    },
    {
      guardId: "secret_value_hidden",
      label: "Secret value hidden",
      labelMessage: { key: "gatewayRouteEvidence_secret_value_hidden_label" as const },
      status: "locked",
      sourceRef: "production-secret-backend-contract",
      enforcedBy: "credential_state and field_sources only",
      enforcedByMessage: { key: "gatewayRouteEvidence_secret_value_hidden_enforcedBy" as const },
      summary: "Profile evidence can show configured, missing, optional_missing, or not_required states without rendering raw credential values.",
      summaryMessage: { key: "gatewayRouteEvidence_secret_value_hidden_summary" as const },
    },
    {
      guardId: "request_observability",
      label: "Request observability",
      labelMessage: { key: "gatewayRouteEvidence_request_observability_label" as const },
      status: "offline_only",
      sourceRef: "docs/contracts/service-api.md",
      enforcedBy: "request_id plus route/provider/profile fields",
      enforcedByMessage: { key: "gatewayRouteEvidence_request_observability_enforcedBy" as const },
      summary: "Route, status, latency, provider/profile, selected model, selection source, error code, and failure boundary are the audit fields for future runs.",
      summaryMessage: { key: "gatewayRouteEvidence_request_observability_summary" as const },
    },
    {
      guardId: "overview_dependency",
      label: "Overview dependency",
      labelMessage: { key: "gatewayRouteEvidence_overview_dependency_label" as const },
      status: source.overview.canRenderModelGatewayOverview ? "ready" : "blocked",
      sourceRef: source.overview.pageId,
      enforcedBy: "overview evidence must render first",
      enforcedByMessage: { key: "gatewayRouteEvidence_overview_dependency_enforcedBy" as const },
      summary: "Route evidence detail remains subordinate to the Model Gateway Overview and does not create a second product source of truth.",
      summaryMessage: { key: "gatewayRouteEvidence_overview_dependency_summary" as const },
    },
  ];
}

function buildRouteNarrativeMessage(
  source: ModelGatewayRouteEvidenceSource,
  profiles: ModelGatewayProviderProfileEvidence[],
  routeBindings: ModelGatewayRouteBindingEvidence[],
): EvidenceProjectionMessage {
  const configuredProfiles = profiles.filter((profile) => profile.credentialState === "configured").length;
  const blockedProfiles = profiles.filter((profile) => profile.status === "blocked").length;
  return { key: "gatewayRouteEvidence_routeNarrative" as const, values: { profilesCount: profiles.length, configuredProfiles: configuredProfiles, blockedProfiles: blockedProfiles, routeBindingsCount: routeBindings.length, overviewPageId: source.overview.pageId } };
}
