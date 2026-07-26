const DEV_SOURCE = "dev-admin-provider-route-http";
const DEFAULT_BASE_URL = "http://127.0.0.1:7000";
const IDENTIFIER = /^[A-Za-z0-9][A-Za-z0-9._:-]{0,159}$/u;
const MODEL_IDENTIFIER = /^[A-Za-z0-9][A-Za-z0-9._:/-]{0,159}$/u;
const DIGEST = /^sha256:[a-f0-9]{64}$/u;
const FORBIDDEN_RESPONSE_FIELDS = new Set([
  "authorization", "api_key", "credential", "secret", "endpoint", "base_url", "dsn", "headers", "cookie",
  "raw_request", "raw_response", "provider_raw_response", "prompt", "messages", "input", "output",
]);

export type AdminProviderRouteEnvironment = "development" | "test";
export type AdminProviderRouteProtocol = "chat_completions" | "responses" | "messages";
export type AdminProviderRouteCandidateState = "pending_review" | "approved" | "rejected";
export type AdminProviderRouteDecision = "approve" | "reject";
export type AdminProviderRouteActivationAction = "activate" | "rollback";

export type AdminProviderRouteConfig = {
  mode: "offline" | "dev_admin_provider_route_http";
  baseUrl: string;
  tenantRef: string;
  workspaceId: string;
  environment: AdminProviderRouteEnvironment;
  configurationId: string;
  applicationId: string;
  subjectRef: string;
  defaultProviderId?: string;
  defaultRuntimeProfile?: string;
  defaultModelId?: string;
};

export type AdminProviderProfileAssignment = {
  profileId: string;
  displayName: string;
  providerId: string;
  runtimeProfileRef: string;
  capabilities: AdminProviderRouteProtocol[];
};

export type AdminModelRouteDefinition = {
  routeId: string;
  protocol: AdminProviderRouteProtocol;
  modelId: string;
  providerProfileId: string;
};

export type AdminProviderRouteConfiguration = {
  displayName: string;
  providerProfiles: AdminProviderProfileAssignment[];
  modelRoutes: AdminModelRouteDefinition[];
};

export type AdminProviderRouteDraftInput = AdminProviderRouteConfiguration & {
  expectedRevision: number;
};

export type AdminProviderRouteDraft = AdminProviderRouteConfiguration & {
  schemaVersion: "admin_provider_route_configuration_draft.v1";
  tenantRef: string;
  workspaceId: string;
  environment: AdminProviderRouteEnvironment;
  configurationId: string;
  draftRevision: number;
  draftDigest: string;
  createdAt: string;
  updatedAt: string;
  createdByActorRef: string;
  updatedByActorRef: string;
  requestId: string;
  auditRef: string;
};

export type AdminProviderInventoryBinding = {
  profileId: string;
  providerId: string;
  runtimeProfileRef: string;
  environment: AdminProviderRouteEnvironment;
  capabilities: AdminProviderRouteProtocol[];
  inventoryDigest: string;
  enabled: boolean;
};

export type AdminProviderRouteReview = {
  schemaVersion: "admin_provider_route_review.v1";
  reviewVersion: number;
  decision: AdminProviderRouteDecision;
  reason: string;
  resultingState: AdminProviderRouteCandidateState;
  reviewedAt: string;
  reviewerRef: string;
  requestId: string;
  auditRef: string;
};

export type AdminProviderRouteCandidate = {
  schemaVersion: "admin_provider_route_candidate.v1";
  tenantRef: string;
  workspaceId: string;
  environment: AdminProviderRouteEnvironment;
  configurationId: string;
  candidateId: string;
  sourceDraftRevision: number;
  sourceDraftDigest: string;
  configuration: AdminProviderRouteConfiguration;
  inventoryBindings: AdminProviderInventoryBinding[];
  candidateDigest: string;
  candidateState: AdminProviderRouteCandidateState;
  reviewVersion: number;
  review: AdminProviderRouteReview | null;
  createdAt: string;
  createdByActorRef: string;
  requestId: string;
  auditRef: string;
};

export type AdminProviderRouteSnapshot = {
  schemaVersion: "admin_provider_route_snapshot.v1";
  tenantRef: string;
  workspaceId: string;
  environment: AdminProviderRouteEnvironment;
  configurationId: string;
  generation: number;
  candidateId: string;
  candidateDigest: string;
  configuration: AdminProviderRouteConfiguration;
  inventoryBindings: AdminProviderInventoryBinding[];
  snapshotDigest: string;
  activatedAt: string;
  activatedByActorRef: string;
  requestId: string;
  auditRef: string;
};

export type AdminProviderRouteActivation = {
  schemaVersion: "admin_provider_route_activation_record.v1";
  activationId: string;
  configurationId: string;
  action: AdminProviderRouteActivationAction;
  reason: string;
  beforeGeneration: number;
  afterGeneration: number;
  beforeCandidateId: string;
  beforeSnapshotDigest: string;
  afterCandidateId: string;
  afterSnapshotDigest: string;
  previousRecordDigest: string;
  recordDigest: string;
  createdAt: string;
  actorRef: string;
  requestId: string;
  auditRef: string;
};

export type AdminProviderRouteEnvelope = {
  requestId: string;
  workspaceId: string;
  environment: AdminProviderRouteEnvironment;
  configurationId: string;
  candidateId: string;
  draft: AdminProviderRouteDraft | null;
  candidate: AdminProviderRouteCandidate | null;
  snapshot: AdminProviderRouteSnapshot | null;
  activation: AdminProviderRouteActivation | null;
  activationHistory: AdminProviderRouteActivation[];
  failureCode: string;
  currentDraftRevision: number;
  currentReviewVersion: number;
  currentCandidateState: string;
  currentGeneration: number;
  auditRef: string;
};

export type AdminProviderRouteValidationFinding = {
  field: string;
  summary: string;
};

export type AdminProviderRouteDiffItem = {
  kind: "display_name" | "provider_profile" | "model_route";
  change: "added" | "removed" | "changed";
  resourceId: string;
  before: string;
  after: string;
};

export type AdminProviderRouteCandidateDiff = {
  baselineGeneration: number;
  changed: boolean;
  items: AdminProviderRouteDiffItem[];
  summary: string;
};

type Document = Record<string, unknown>;

export function readAdminProviderRouteConfig(): AdminProviderRouteConfig {
  const env = import.meta.env as Record<string, string | undefined>;
  const configuredEnvironment = env.VITE_RADISHMIND_ADMIN_PROVIDER_ROUTE_ENVIRONMENT?.trim();
  return {
    mode: env.VITE_RADISHMIND_ADMIN_PROVIDER_ROUTE_SOURCE?.trim() === DEV_SOURCE
      ? "dev_admin_provider_route_http"
      : "offline",
    baseUrl: normalizeBaseUrl(
      env.VITE_RADISHMIND_ADMIN_PROVIDER_ROUTE_BASE_URL ??
      env.VITE_RADISHMIND_CONTROL_PLANE_READ_BASE_URL ??
      DEFAULT_BASE_URL,
    ),
    tenantRef: env.VITE_RADISHMIND_DEV_READ_TENANT_REF?.trim() || "tenant_demo",
    workspaceId: env.VITE_RADISHMIND_ADMIN_PROVIDER_ROUTE_WORKSPACE_ID?.trim() || "workspace_demo",
    environment: configuredEnvironment === "development" ? "development" : "test",
    configurationId: env.VITE_RADISHMIND_ADMIN_PROVIDER_ROUTE_CONFIGURATION_ID?.trim() || "gateway-default",
    applicationId: env.VITE_RADISHMIND_ADMIN_PROVIDER_ROUTE_APPLICATION_ID?.trim() || "app_flow_copilot",
    subjectRef: env.VITE_RADISHMIND_DEV_READ_SUBJECT_REF?.trim() || "subject_demo_user",
    defaultProviderId: env.VITE_RADISHMIND_ADMIN_PROVIDER_ROUTE_DEFAULT_PROVIDER_ID?.trim() || "",
    defaultRuntimeProfile: env.VITE_RADISHMIND_ADMIN_PROVIDER_ROUTE_DEFAULT_RUNTIME_PROFILE?.trim() || "",
    defaultModelId: env.VITE_RADISHMIND_ADMIN_PROVIDER_ROUTE_DEFAULT_MODEL_ID?.trim() || "",
  };
}

export function createAdminProviderRouteDraftInput(
  config: AdminProviderRouteConfig,
  providerId = "",
  runtimeProfile = "",
  modelId = "",
): AdminProviderRouteDraftInput {
  providerId = providerId.trim() || config.defaultProviderId?.trim() || "";
  runtimeProfile = runtimeProfile.trim() || config.defaultRuntimeProfile?.trim() || "";
  modelId = modelId.trim() || config.defaultModelId?.trim() || "";
  const profileId = "primary";
  return {
    expectedRevision: 0,
    displayName: `${config.environment === "test" ? "Test" : "Development"} Gateway routing`,
    providerProfiles: [{
      profileId,
      displayName: "Primary runtime profile",
      providerId,
      runtimeProfileRef: runtimeProfile
        ? `ref:radishmind/${config.environment}/provider-profiles/${runtimeProfile}`
        : "",
      capabilities: ["chat_completions"],
    }],
    modelRoutes: [{
      routeId: "route-chat-primary",
      protocol: "chat_completions",
      modelId,
      providerProfileId: profileId,
    }],
  };
}

export function draftInputFromAdminProviderRouteDraft(
  draft: AdminProviderRouteDraft,
): AdminProviderRouteDraftInput {
  return {
    expectedRevision: draft.draftRevision,
    displayName: draft.displayName,
    providerProfiles: draft.providerProfiles.map(copyProviderProfile),
    modelRoutes: draft.modelRoutes.map(copyModelRoute),
  };
}

export function validateAdminProviderRouteDraft(
  config: AdminProviderRouteConfig,
  draft: AdminProviderRouteDraftInput,
): AdminProviderRouteValidationFinding[] {
  const findings: AdminProviderRouteValidationFinding[] = [];
  if (draft.expectedRevision < 0 || !Number.isInteger(draft.expectedRevision)) {
    findings.push({ field: "expected_revision", summary: "Draft revision must be a non-negative integer." });
  }
  if (draft.displayName.trim().length < 2 || draft.displayName.trim().length > 120) {
    findings.push({ field: "display_name", summary: "Display name must contain 2 to 120 characters." });
  }
  if (containsSensitiveMaterial(JSON.stringify(draft))) {
    findings.push({ field: "draft", summary: "Credentials, endpoints, authorization material, cookies, and DSNs are forbidden." });
  }
  if (draft.providerProfiles.length < 1 || draft.providerProfiles.length > 32) {
    findings.push({ field: "provider_profiles", summary: "The draft must contain 1 to 32 provider profile assignments." });
  }
  if (draft.modelRoutes.length < 1 || draft.modelRoutes.length > 128) {
    findings.push({ field: "model_routes", summary: "The draft must contain 1 to 128 model routes." });
  }
  validateProviderProfiles(config, draft.providerProfiles, findings);
  validateModelRoutes(draft.providerProfiles, draft.modelRoutes, findings);
  return findings;
}

export function buildAdminProviderRouteCandidateDiff(
  candidate: AdminProviderRouteCandidate,
  snapshot: AdminProviderRouteSnapshot | null,
): AdminProviderRouteCandidateDiff {
  const items: AdminProviderRouteDiffItem[] = [];
  const baseline = snapshot?.configuration ?? { displayName: "", providerProfiles: [], modelRoutes: [] };
  if (baseline.displayName !== candidate.configuration.displayName) {
    items.push({
      kind: "display_name",
      change: baseline.displayName ? "changed" : "added",
      resourceId: "display_name",
      before: baseline.displayName,
      after: candidate.configuration.displayName,
    });
  }
  compareResources(
    baseline.providerProfiles,
    candidate.configuration.providerProfiles,
    (item) => item.profileId,
    providerProfileSummary,
    "provider_profile",
    items,
  );
  compareResources(
    baseline.modelRoutes,
    candidate.configuration.modelRoutes,
    (item) => item.routeId,
    modelRouteSummary,
    "model_route",
    items,
  );
  return {
    baselineGeneration: snapshot?.generation ?? 0,
    changed: items.length > 0,
    items,
    summary: items.length
      ? `${items.length} configuration changes compared with generation ${snapshot?.generation ?? 0}.`
      : `Candidate configuration matches generation ${snapshot?.generation ?? 0}.`,
  };
}

export async function readAdminProviderRouteDraft(
  config: AdminProviderRouteConfig,
): Promise<AdminProviderRouteEnvelope> {
  return requestAdminProviderRoute(config, configurationPath(config), "GET", null, "read");
}

export async function saveAdminProviderRouteDraft(
  config: AdminProviderRouteConfig,
  draft: AdminProviderRouteDraftInput,
): Promise<AdminProviderRouteEnvelope> {
  return requestAdminProviderRoute(config, configurationPath(config), "PUT", {
    expected_revision: draft.expectedRevision,
    display_name: draft.displayName.trim(),
    provider_profiles: draft.providerProfiles.map(providerProfilePayload),
    model_routes: draft.modelRoutes.map(modelRoutePayload),
  }, "draft");
}

export async function createAdminProviderRouteCandidate(
  config: AdminProviderRouteConfig,
  candidateId: string,
  expectedDraftRevision: number,
): Promise<AdminProviderRouteEnvelope> {
  return requestAdminProviderRoute(config, `${configurationPath(config)}/candidates`, "POST", {
    candidate_id: candidateId.trim(),
    expected_draft_revision: expectedDraftRevision,
  }, "draft");
}

export async function readAdminProviderRouteCandidate(
  config: AdminProviderRouteConfig,
  candidateId: string,
): Promise<AdminProviderRouteEnvelope> {
  return requestAdminProviderRoute(
    config,
    `${configurationPath(config)}/candidates/${encodeURIComponent(candidateId.trim())}`,
    "GET",
    null,
    "read",
  );
}

export async function reviewAdminProviderRouteCandidate(
  config: AdminProviderRouteConfig,
  candidateId: string,
  expectedReviewVersion: number,
  decision: AdminProviderRouteDecision,
  reason: string,
): Promise<AdminProviderRouteEnvelope> {
  return requestAdminProviderRoute(
    config,
    `${configurationPath(config)}/candidates/${encodeURIComponent(candidateId.trim())}/reviews`,
    "POST",
    { expected_review_version: expectedReviewVersion, decision, reason: reason.trim() },
    "review",
  );
}

export async function activateAdminProviderRouteCandidate(
  config: AdminProviderRouteConfig,
  candidateId: string,
  expectedGeneration: number,
  action: AdminProviderRouteActivationAction,
  reason: string,
): Promise<AdminProviderRouteEnvelope> {
  return requestAdminProviderRoute(
    config,
    `${configurationPath(config)}/candidates/${encodeURIComponent(candidateId.trim())}/activations`,
    "POST",
    { expected_generation: expectedGeneration, action, reason: reason.trim() },
    "activate",
  );
}

export async function readAdminProviderRouteSnapshot(
  config: AdminProviderRouteConfig,
): Promise<AdminProviderRouteEnvelope> {
  return requestAdminProviderRoute(config, `${configurationPath(config)}/active-snapshot`, "GET", null, "read");
}

export async function listAdminProviderRouteActivations(
  config: AdminProviderRouteConfig,
): Promise<AdminProviderRouteEnvelope> {
  return requestAdminProviderRoute(config, `${configurationPath(config)}/activation-history`, "GET", null, "read");
}

function validateProviderProfiles(
  config: AdminProviderRouteConfig,
  profiles: AdminProviderProfileAssignment[],
  findings: AdminProviderRouteValidationFinding[],
) {
  const profileIds = new Set<string>();
  const expectedPrefix = `ref:radishmind/${config.environment}/provider-profiles/`;
  for (const [index, profile] of profiles.entries()) {
    const field = `provider_profiles[${index}]`;
    if (!IDENTIFIER.test(profile.profileId) || !IDENTIFIER.test(profile.providerId)) {
      findings.push({ field, summary: "Profile and provider identifiers do not match the Admin route contract." });
    }
    if (profileIds.has(profile.profileId)) {
      findings.push({ field, summary: `Profile ${profile.profileId} is duplicated.` });
    }
    profileIds.add(profile.profileId);
    if (profile.displayName.trim().length < 2 || profile.displayName.trim().length > 120) {
      findings.push({ field, summary: "Profile display name must contain 2 to 120 characters." });
    }
    if (!profile.runtimeProfileRef.startsWith(expectedPrefix) ||
      !IDENTIFIER.test(profile.runtimeProfileRef.slice(expectedPrefix.length))) {
      findings.push({ field, summary: `Runtime profile ref must use ${expectedPrefix}<profile>.` });
    }
    if (profile.capabilities.length < 1 || profile.capabilities.length > 8 ||
      new Set(profile.capabilities).size !== profile.capabilities.length ||
      profile.capabilities.some((capability) => !isProtocol(capability))) {
      findings.push({ field, summary: "Capabilities must be unique supported northbound protocols." });
    }
  }
}

function validateModelRoutes(
  profiles: AdminProviderProfileAssignment[],
  routes: AdminModelRouteDefinition[],
  findings: AdminProviderRouteValidationFinding[],
) {
  const profileIds = new Set(profiles.map((profile) => profile.profileId));
  const routeIds = new Set<string>();
  const bindings = new Set<string>();
  for (const [index, route] of routes.entries()) {
    const field = `model_routes[${index}]`;
    if (!IDENTIFIER.test(route.routeId) || !MODEL_IDENTIFIER.test(route.modelId) ||
      !IDENTIFIER.test(route.providerProfileId) || !isProtocol(route.protocol)) {
      findings.push({ field, summary: "Route identifier, protocol, model, or profile reference is invalid." });
    }
    if (routeIds.has(route.routeId)) {
      findings.push({ field, summary: `Route ${route.routeId} is duplicated.` });
    }
    routeIds.add(route.routeId);
    const binding = `${route.protocol}\u0000${route.modelId}`;
    if (bindings.has(binding)) {
      findings.push({ field, summary: `Protocol and model binding ${route.protocol} / ${route.modelId} is duplicated.` });
    }
    bindings.add(binding);
    const profile = profiles.find((item) => item.profileId === route.providerProfileId);
    if (!profileIds.has(route.providerProfileId) || !profile?.capabilities.includes(route.protocol)) {
      findings.push({ field, summary: "Route must reference a profile assignment with the same capability." });
    }
  }
}

async function requestAdminProviderRoute(
  config: AdminProviderRouteConfig,
  path: string,
  method: "GET" | "PUT" | "POST",
  body: unknown,
  operation: "read" | "draft" | "review" | "activate",
): Promise<AdminProviderRouteEnvelope> {
  if (config.mode === "offline") {
    return offlineEnvelope(config);
  }
  const requestId = createRequestId(operation);
  const response = await fetch(`${config.baseUrl}${path}`, {
    method,
    headers: adminProviderRouteHeaders(config, requestId, operation, body !== null),
    body: body === null ? undefined : JSON.stringify(body),
  });
  const value: unknown = await response.json();
  assertNoForbiddenFields(value);
  if (!isEnvelopeDocument(value, config)) {
    throw new Error(`Admin Provider route returned an invalid HTTP ${response.status} envelope.`);
  }
  return mapEnvelope(value);
}

function adminProviderRouteHeaders(
  config: AdminProviderRouteConfig,
  requestId: string,
  operation: "read" | "draft" | "review" | "activate",
  hasBody: boolean,
): Record<string, string> {
  const scopes = operation === "read"
    ? ["admin_provider_routes:read"]
    : [`admin_provider_routes:${operation}`, "admin_provider_routes:read"];
  return {
    Accept: "application/json",
    ...(hasBody ? { "Content-Type": "application/json" } : {}),
    "X-Request-Id": requestId,
    "X-RadishMind-Dev-Read-Identity": "admin-provider-route-web",
    "X-RadishMind-Dev-Read-Tenant": config.tenantRef,
    "X-RadishMind-Dev-Read-Subject": config.subjectRef,
    "X-RadishMind-Dev-Read-Scopes": scopes.join(","),
    "X-RadishMind-Dev-Read-Audit": `audit_${requestId}`,
    "X-RadishMind-Dev-Admin-Provider-Route-Workspace": config.workspaceId,
    "X-RadishMind-Dev-Admin-Provider-Route-Environment": config.environment,
  };
}

function configurationPath(config: AdminProviderRouteConfig): string {
  return `/v1/admin/provider-route-configurations/${encodeURIComponent(config.configurationId)}`;
}

function providerProfilePayload(profile: AdminProviderProfileAssignment) {
  return {
    profile_id: profile.profileId.trim(),
    display_name: profile.displayName.trim(),
    provider_id: profile.providerId.trim(),
    runtime_profile_ref: profile.runtimeProfileRef.trim(),
    capabilities: profile.capabilities,
  };
}

function modelRoutePayload(route: AdminModelRouteDefinition) {
  return {
    route_id: route.routeId.trim(),
    protocol: route.protocol,
    model_id: route.modelId.trim(),
    provider_profile_id: route.providerProfileId.trim(),
  };
}

function mapEnvelope(value: Document): AdminProviderRouteEnvelope {
  return {
    requestId: String(value.request_id),
    workspaceId: String(value.workspace_id),
    environment: value.environment as AdminProviderRouteEnvironment,
    configurationId: String(value.configuration_id),
    candidateId: optionalString(value.candidate_id),
    draft: isDocument(value.draft) ? mapDraft(value.draft) : null,
    candidate: isDocument(value.candidate) ? mapCandidate(value.candidate) : null,
    snapshot: isDocument(value.snapshot) ? mapSnapshot(value.snapshot) : null,
    activation: isDocument(value.activation) ? mapActivation(value.activation) : null,
    activationHistory: (value.activation_history as Document[]).map(mapActivation),
    failureCode: optionalString(value.failure_code),
    currentDraftRevision: optionalInteger(value.current_draft_revision),
    currentReviewVersion: optionalInteger(value.current_review_version),
    currentCandidateState: optionalString(value.current_candidate_state),
    currentGeneration: optionalInteger(value.current_generation),
    auditRef: String(value.audit_ref),
  };
}

function mapDraft(value: Document): AdminProviderRouteDraft {
  return {
    schemaVersion: value.schema_version as AdminProviderRouteDraft["schemaVersion"],
    tenantRef: String(value.tenant_ref),
    workspaceId: String(value.workspace_id),
    environment: value.environment as AdminProviderRouteEnvironment,
    configurationId: String(value.configuration_id),
    draftRevision: Number(value.draft_revision),
    ...mapConfiguration(value),
    draftDigest: String(value.draft_digest),
    createdAt: String(value.created_at),
    updatedAt: String(value.updated_at),
    createdByActorRef: String(value.created_by_actor_ref),
    updatedByActorRef: String(value.updated_by_actor_ref),
    requestId: String(value.request_id),
    auditRef: String(value.audit_ref),
  };
}

function mapCandidate(value: Document): AdminProviderRouteCandidate {
  return {
    schemaVersion: value.schema_version as AdminProviderRouteCandidate["schemaVersion"],
    tenantRef: String(value.tenant_ref),
    workspaceId: String(value.workspace_id),
    environment: value.environment as AdminProviderRouteEnvironment,
    configurationId: String(value.configuration_id),
    candidateId: String(value.candidate_id),
    sourceDraftRevision: Number(value.source_draft_revision),
    sourceDraftDigest: String(value.source_draft_digest),
    configuration: mapConfiguration(value.configuration as Document),
    inventoryBindings: (value.inventory_bindings as Document[]).map(mapInventoryBinding),
    candidateDigest: String(value.candidate_digest),
    candidateState: value.candidate_state as AdminProviderRouteCandidateState,
    reviewVersion: Number(value.review_version),
    review: isDocument(value.review) ? mapReview(value.review) : null,
    createdAt: String(value.created_at),
    createdByActorRef: String(value.created_by_actor_ref),
    requestId: String(value.request_id),
    auditRef: String(value.audit_ref),
  };
}

function mapSnapshot(value: Document): AdminProviderRouteSnapshot {
  return {
    schemaVersion: value.schema_version as AdminProviderRouteSnapshot["schemaVersion"],
    tenantRef: String(value.tenant_ref),
    workspaceId: String(value.workspace_id),
    environment: value.environment as AdminProviderRouteEnvironment,
    configurationId: String(value.configuration_id),
    generation: Number(value.generation),
    candidateId: String(value.candidate_id),
    candidateDigest: String(value.candidate_digest),
    configuration: mapConfiguration(value.configuration as Document),
    inventoryBindings: (value.inventory_bindings as Document[]).map(mapInventoryBinding),
    snapshotDigest: String(value.snapshot_digest),
    activatedAt: String(value.activated_at),
    activatedByActorRef: String(value.activated_by_actor_ref),
    requestId: String(value.request_id),
    auditRef: String(value.audit_ref),
  };
}

function mapActivation(value: Document): AdminProviderRouteActivation {
  return {
    schemaVersion: value.schema_version as AdminProviderRouteActivation["schemaVersion"],
    activationId: String(value.activation_id),
    configurationId: String(value.configuration_id),
    action: value.action as AdminProviderRouteActivationAction,
    reason: String(value.reason),
    beforeGeneration: Number(value.before_generation),
    afterGeneration: Number(value.after_generation),
    beforeCandidateId: optionalString(value.before_candidate_id),
    beforeSnapshotDigest: optionalString(value.before_snapshot_digest),
    afterCandidateId: String(value.after_candidate_id),
    afterSnapshotDigest: String(value.after_snapshot_digest),
    previousRecordDigest: optionalString(value.previous_record_digest),
    recordDigest: String(value.record_digest),
    createdAt: String(value.created_at),
    actorRef: String(value.actor_ref),
    requestId: String(value.request_id),
    auditRef: String(value.audit_ref),
  };
}

function mapReview(value: Document): AdminProviderRouteReview {
  return {
    schemaVersion: value.schema_version as AdminProviderRouteReview["schemaVersion"],
    reviewVersion: Number(value.review_version),
    decision: value.decision as AdminProviderRouteDecision,
    reason: String(value.reason),
    resultingState: value.resulting_state as AdminProviderRouteCandidateState,
    reviewedAt: String(value.reviewed_at),
    reviewerRef: String(value.reviewer_ref),
    requestId: String(value.request_id),
    auditRef: String(value.audit_ref),
  };
}

function mapConfiguration(value: Document): AdminProviderRouteConfiguration {
  return {
    displayName: String(value.display_name),
    providerProfiles: (value.provider_profiles as Document[]).map((profile) => ({
      profileId: String(profile.profile_id),
      displayName: String(profile.display_name),
      providerId: String(profile.provider_id),
      runtimeProfileRef: String(profile.runtime_profile_ref),
      capabilities: profile.capabilities as AdminProviderRouteProtocol[],
    })),
    modelRoutes: (value.model_routes as Document[]).map((route) => ({
      routeId: String(route.route_id),
      protocol: route.protocol as AdminProviderRouteProtocol,
      modelId: String(route.model_id),
      providerProfileId: String(route.provider_profile_id),
    })),
  };
}

function mapInventoryBinding(value: Document): AdminProviderInventoryBinding {
  return {
    profileId: String(value.profile_id),
    providerId: String(value.provider_id),
    runtimeProfileRef: String(value.runtime_profile_ref),
    environment: value.environment as AdminProviderRouteEnvironment,
    capabilities: value.capabilities as AdminProviderRouteProtocol[],
    inventoryDigest: String(value.inventory_digest),
    enabled: Boolean(value.enabled),
  };
}

function isEnvelopeDocument(value: unknown, config: AdminProviderRouteConfig): value is Document {
  if (!isDocument(value) ||
    !stringFields(value, ["request_id", "workspace_id", "environment", "configuration_id", "audit_ref"]) ||
    value.workspace_id !== config.workspaceId ||
    value.environment !== config.environment ||
    value.configuration_id !== config.configurationId ||
    !Array.isArray(value.activation_history) ||
    !value.activation_history.every(isActivationDocument) ||
    !isNullableString(value.failure_code)) {
    return false;
  }
  return isOptionalDocument(value.draft, isDraftDocument) &&
    isOptionalDocument(value.candidate, isCandidateDocument) &&
    isOptionalDocument(value.snapshot, isSnapshotDocument) &&
    isOptionalDocument(value.activation, isActivationDocument) &&
    optionalNumberFields(value, [
      "current_draft_revision", "current_review_version", "current_generation",
    ]) &&
    optionalStringFields(value, ["candidate_id", "current_candidate_state"]);
}

function isDraftDocument(value: Document): boolean {
  return value.schema_version === "admin_provider_route_configuration_draft.v1" &&
    isScopedResource(value) && isConfigurationDocument(value) &&
    isPositiveInteger(value.draft_revision) && isDigest(value.draft_digest) &&
    stringFields(value, [
      "created_at", "updated_at", "created_by_actor_ref", "updated_by_actor_ref", "request_id", "audit_ref",
    ]);
}

function isCandidateDocument(value: Document): boolean {
  return value.schema_version === "admin_provider_route_candidate.v1" &&
    isScopedResource(value) && stringFields(value, ["candidate_id", "created_at", "created_by_actor_ref", "request_id", "audit_ref"]) &&
    isPositiveInteger(value.source_draft_revision) && isDigest(value.source_draft_digest) &&
    isDocument(value.configuration) && isConfigurationDocument(value.configuration) &&
    Array.isArray(value.inventory_bindings) && value.inventory_bindings.every(isInventoryBindingDocument) &&
    isDigest(value.candidate_digest) &&
    ["pending_review", "approved", "rejected"].includes(String(value.candidate_state)) &&
    isNonNegativeInteger(value.review_version) &&
    isOptionalDocument(value.review, isReviewDocument);
}

function isSnapshotDocument(value: Document): boolean {
  return value.schema_version === "admin_provider_route_snapshot.v1" &&
    isScopedResource(value) && isPositiveInteger(value.generation) &&
    stringFields(value, ["candidate_id", "activated_at", "activated_by_actor_ref", "request_id", "audit_ref"]) &&
    isDigest(value.candidate_digest) && isDigest(value.snapshot_digest) &&
    isDocument(value.configuration) && isConfigurationDocument(value.configuration) &&
    Array.isArray(value.inventory_bindings) && value.inventory_bindings.every(isInventoryBindingDocument);
}

function isActivationDocument(value: unknown): value is Document {
  if (!isDocument(value)) return false;
  return value.schema_version === "admin_provider_route_activation_record.v1" &&
    stringFields(value, [
      "activation_id", "configuration_id", "reason", "after_candidate_id", "created_at", "actor_ref", "request_id", "audit_ref",
    ]) &&
    ["activate", "rollback"].includes(String(value.action)) &&
    isNonNegativeInteger(value.before_generation) && isPositiveInteger(value.after_generation) &&
    isDigest(value.after_snapshot_digest) && isDigest(value.record_digest) &&
    optionalDigest(value.before_snapshot_digest) && optionalDigest(value.previous_record_digest) &&
    optionalStringFields(value, ["before_candidate_id"]);
}

function isConfigurationDocument(value: Document): boolean {
  return typeof value.display_name === "string" &&
    Array.isArray(value.provider_profiles) && value.provider_profiles.length > 0 &&
    value.provider_profiles.every(isProviderProfileDocument) &&
    Array.isArray(value.model_routes) && value.model_routes.length > 0 &&
    value.model_routes.every(isModelRouteDocument);
}

function isProviderProfileDocument(value: unknown): boolean {
  return isDocument(value) &&
    stringFields(value, ["profile_id", "display_name", "provider_id", "runtime_profile_ref"]) &&
    Array.isArray(value.capabilities) && value.capabilities.length > 0 &&
    value.capabilities.every((capability) => typeof capability === "string" && isProtocol(capability));
}

function isModelRouteDocument(value: unknown): boolean {
  return isDocument(value) &&
    stringFields(value, ["route_id", "model_id", "provider_profile_id"]) &&
    typeof value.protocol === "string" && isProtocol(value.protocol);
}

function isInventoryBindingDocument(value: unknown): boolean {
  return isDocument(value) &&
    stringFields(value, ["profile_id", "provider_id", "runtime_profile_ref", "environment"]) &&
    (value.environment === "development" || value.environment === "test") &&
    Array.isArray(value.capabilities) && value.capabilities.every((capability) => typeof capability === "string" && isProtocol(capability)) &&
    isDigest(value.inventory_digest) && typeof value.enabled === "boolean";
}

function isReviewDocument(value: Document): boolean {
  return value.schema_version === "admin_provider_route_review.v1" &&
    isPositiveInteger(value.review_version) &&
    ["approve", "reject"].includes(String(value.decision)) &&
    ["approved", "rejected"].includes(String(value.resulting_state)) &&
    stringFields(value, ["reason", "reviewed_at", "reviewer_ref", "request_id", "audit_ref"]);
}

function isScopedResource(value: Document): boolean {
  return stringFields(value, ["tenant_ref", "workspace_id", "environment", "configuration_id"]) &&
    (value.environment === "development" || value.environment === "test");
}

function offlineEnvelope(config: AdminProviderRouteConfig): AdminProviderRouteEnvelope {
  return {
    requestId: "admin-provider-route-offline",
    workspaceId: config.workspaceId,
    environment: config.environment,
    configurationId: config.configurationId,
    candidateId: "",
    draft: null,
    candidate: null,
    snapshot: null,
    activation: null,
    activationHistory: [],
    failureCode: "admin_provider_route_http_disabled",
    currentDraftRevision: 0,
    currentReviewVersion: 0,
    currentCandidateState: "",
    currentGeneration: 0,
    auditRef: "audit_admin_provider_route_offline",
  };
}

function compareResources<T>(
  baseline: T[],
  candidate: T[],
  identity: (item: T) => string,
  summarize: (item: T) => string,
  kind: "provider_profile" | "model_route",
  items: AdminProviderRouteDiffItem[],
) {
  const baselineById = new Map(baseline.map((item) => [identity(item), item]));
  const candidateById = new Map(candidate.map((item) => [identity(item), item]));
  const resourceIds = [...new Set([...baselineById.keys(), ...candidateById.keys()])].sort();
  for (const resourceId of resourceIds) {
    const beforeResource = baselineById.get(resourceId);
    const afterResource = candidateById.get(resourceId);
    const before = beforeResource ? summarize(beforeResource) : "";
    const after = afterResource ? summarize(afterResource) : "";
    if (before === after) continue;
    items.push({
      kind,
      change: beforeResource ? (afterResource ? "changed" : "removed") : "added",
      resourceId,
      before,
      after,
    });
  }
}

function providerProfileSummary(profile: AdminProviderProfileAssignment): string {
  return `${profile.providerId} · ${profile.runtimeProfileRef} · ${[...profile.capabilities].sort().join(",")}`;
}

function modelRouteSummary(route: AdminModelRouteDefinition): string {
  return `${route.protocol} · ${route.modelId} → ${route.providerProfileId}`;
}

function copyProviderProfile(profile: AdminProviderProfileAssignment): AdminProviderProfileAssignment {
  return { ...profile, capabilities: [...profile.capabilities] };
}

function copyModelRoute(route: AdminModelRouteDefinition): AdminModelRouteDefinition {
  return { ...route };
}

function containsSensitiveMaterial(value: string): boolean {
  return /(authorization|bearer\s+|api[_ -]?key|credential|secret|cookie|dsn|https?:\/\/|base[_ -]?url)/iu.test(value);
}

function isProtocol(value: unknown): value is AdminProviderRouteProtocol {
  return value === "chat_completions" || value === "responses" || value === "messages";
}

function createRequestId(operation: string): string {
  return `admin-provider-route-web-${operation}-${Date.now().toString(36)}`;
}

function normalizeBaseUrl(value: string): string {
  return value.trim().replace(/\/$/u, "");
}

function isDocument(value: unknown): value is Document {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function stringFields(value: Document, fields: string[]): boolean {
  return fields.every((field) => typeof value[field] === "string");
}

function optionalStringFields(value: Document, fields: string[]): boolean {
  return fields.every((field) => value[field] === undefined || typeof value[field] === "string");
}

function optionalNumberFields(value: Document, fields: string[]): boolean {
  return fields.every((field) => value[field] === undefined || isNonNegativeInteger(value[field]));
}

function isNullableString(value: unknown): boolean {
  return value === null || typeof value === "string";
}

function isOptionalDocument(value: unknown, validator: (document: Document) => boolean): boolean {
  return value === undefined || value === null || (isDocument(value) && validator(value));
}

function isNonNegativeInteger(value: unknown): value is number {
  return typeof value === "number" && Number.isInteger(value) && value >= 0;
}

function isPositiveInteger(value: unknown): value is number {
  return typeof value === "number" && Number.isInteger(value) && value > 0;
}

function optionalInteger(value: unknown): number {
  return isNonNegativeInteger(value) ? value : 0;
}

function optionalString(value: unknown): string {
  return typeof value === "string" ? value : "";
}

function isDigest(value: unknown): boolean {
  return typeof value === "string" && DIGEST.test(value);
}

function optionalDigest(value: unknown): boolean {
  return value === undefined || (typeof value === "string" && (value === "" || DIGEST.test(value)));
}

function assertNoForbiddenFields(value: unknown, path = "response") {
  if (Array.isArray(value)) {
    value.forEach((item, index) => assertNoForbiddenFields(item, `${path}[${index}]`));
    return;
  }
  if (!isDocument(value)) return;
  for (const [key, item] of Object.entries(value)) {
    if (FORBIDDEN_RESPONSE_FIELDS.has(key.toLowerCase())) {
      throw new Error(`Admin Provider route response contains forbidden field ${path}.${key}.`);
    }
    assertNoForbiddenFields(item, `${path}.${key}`);
  }
}
