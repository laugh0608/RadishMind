const DEV_SOURCE = "dev-workflow-rag-http";
const DEFAULT_BASE_URL = "http://127.0.0.1:7000";
const COLLECTION_PATH = "/v1/user-workspace/workflow-retrieval-snapshots";
const SNAPSHOT_SCHEMA = "workflow_rag_snapshot.v1";
const FRAGMENT_SCHEMA = "workflow_rag_fragment.v1";
const SNAPSHOT_ID_PATTERN = /^rags_[a-z2-7]{16}$/u;
const SNAPSHOT_KEY_PATTERN = /^[a-z][a-z0-9_]{2,47}$/u;
const FRAGMENT_REF_PATTERN = /^[a-z][a-z0-9_]{2,63}$/u;
const SCOPE_ID_PATTERN = /^[A-Za-z0-9][A-Za-z0-9._:-]{2,119}$/u;
const DIGEST_PATTERN = /^sha256:[a-f0-9]{64}$/u;
const SOURCE_TYPES = ["document", "wiki", "faq", "forum", "manual"] as const;
const CLASSIFICATIONS = ["public", "workspace_internal"] as const;
const RESOURCE_KEYS = [
  "snapshot_id", "tenant_ref", "workspace_id", "application_id", "snapshot_key", "display_name",
  "lifecycle_state", "latest_version", "latest_rag_ref", "latest_digest", "fragment_count",
  "total_content_bytes", "created_at", "updated_at", "archived_at",
] as const;
const RECORD_KEYS = [
  "schema_version", "snapshot_id", "tenant_ref", "workspace_id", "application_id", "snapshot_key", "rag_ref",
  "snapshot_version", "display_name", "lifecycle_state", "content_classification", "profile_ref", "fragment_count",
  "total_content_bytes", "snapshot_digest", "created_at", "created_by_actor_ref", "request_id", "audit_ref", "fragments",
] as const;
const FRAGMENT_KEYS = [
  "schema_version", "fragment_ref", "source_type", "source_ref", "page_slug", "title", "is_official",
  "content", "content_classification", "content_bytes", "content_digest",
] as const;
const OPERATION_ENVELOPE_KEYS = [
  "request_id", "tenant_ref", "workspace_id", "application_id", "record", "failure_code",
  "current_latest_version", "current_lifecycle_state", "audit_ref",
] as const;
const LIST_ENVELOPE_KEYS = [
  "request_id", "tenant_ref", "workspace_id", "application_id", "items", "next_cursor", "failure_code", "audit_ref",
] as const;
const FORBIDDEN_MATERIAL_KEYS = new Set([
  "authorization", "api_key", "secret", "credential", "endpoint", "headers", "cookie", "dsn", "raw_request",
  "raw_response", "query", "prompt", "messages", "selected_fragments",
]);

export type WorkflowRAGSnapshotMode = "offline" | "dev_workflow_rag_http";
export type WorkflowRAGSnapshotLifecycle = "active" | "archived";
export type WorkflowRAGContentClassification = typeof CLASSIFICATIONS[number];
export type WorkflowRAGSourceType = typeof SOURCE_TYPES[number];
export type WorkflowRAGSnapshotScope = "workflow_rag_snapshots:read" | "workflow_rag_snapshots:write" | "workflow_rag_snapshots:archive";
export type WorkflowRAGExecutionScope = "workflow_rag:execute" | "workflow_runs:execute" | "workflow_drafts:read";
export type WorkflowRAGScope = WorkflowRAGSnapshotScope | WorkflowRAGExecutionScope;

export type WorkflowRAGSnapshotConfig = {
  mode: WorkflowRAGSnapshotMode;
  baseUrl: string;
  tenantRef: string;
  workspaceId: string;
  subjectRef: string;
  authMode: "dev_headers" | "signed_test_token" | "radish_oidc_integration_test";
  scopes: ReadonlySet<WorkflowRAGScope>;
};

export type WorkflowRAGFragmentInput = {
  fragmentRef: string;
  sourceType: WorkflowRAGSourceType;
  sourceRef: string;
  pageSlug: string;
  title: string;
  isOfficial: boolean;
  content: string;
};

export type WorkflowRAGFragment = WorkflowRAGFragmentInput & {
  schemaVersion: typeof FRAGMENT_SCHEMA;
  contentClassification: WorkflowRAGContentClassification;
  contentBytes: number;
  contentDigest: string;
};

export type WorkflowRAGSnapshotResource = {
  snapshotId: string;
  tenantRef: string;
  workspaceId: string;
  applicationId: string;
  snapshotKey: string;
  displayName: string;
  lifecycleState: WorkflowRAGSnapshotLifecycle;
  latestVersion: number;
  latestRAGRef: string;
  latestDigest: string;
  fragmentCount: number;
  totalContentBytes: number;
  createdAt: string;
  updatedAt: string;
  archivedAt: string | null;
};

export type WorkflowRAGSnapshotRecord = {
  schemaVersion: typeof SNAPSHOT_SCHEMA;
  snapshotId: string;
  tenantRef: string;
  workspaceId: string;
  applicationId: string;
  snapshotKey: string;
  ragRef: string;
  snapshotVersion: number;
  displayName: string;
  lifecycleState: WorkflowRAGSnapshotLifecycle;
  contentClassification: WorkflowRAGContentClassification;
  profileRef: "workflow.rag.lexical-ngram-dev.v1";
  fragmentCount: number;
  totalContentBytes: number;
  snapshotDigest: string;
  createdAt: string;
  createdByActorRef: string;
  requestId: string;
  auditRef: string;
  fragments: WorkflowRAGFragment[];
};

export type WorkflowRAGSnapshotWriteInput = {
  snapshotKey: string;
  displayName: string;
  contentClassification: WorkflowRAGContentClassification;
  fragments: WorkflowRAGFragmentInput[];
};

export type WorkflowRAGSnapshotListResult = {
  status: "offline" | "scope_denied" | "ready" | "empty" | "failed";
  records: WorkflowRAGSnapshotResource[];
  nextCursor: string;
  failureCode: string;
  summary: string;
};

export type WorkflowRAGSnapshotOperationResult = {
  status: "offline" | "scope_denied" | "created" | "loaded" | "versioned" | "archived" | "version_conflict" | "failed";
  record: WorkflowRAGSnapshotRecord | null;
  failureCode: string;
  currentLatestVersion: number;
  currentLifecycleState: string;
  summary: string;
};

type FragmentDocument = {
  schema_version: string; fragment_ref: string; source_type: string; source_ref: string; page_slug: string;
  title: string; is_official: boolean; content: string; content_classification: string; content_bytes: number; content_digest: string;
};
type SnapshotRecordDocument = {
  schema_version: string; snapshot_id: string; tenant_ref: string; workspace_id: string; application_id: string;
  snapshot_key: string; rag_ref: string; snapshot_version: number; display_name: string; lifecycle_state: string;
  content_classification: string; profile_ref: string; fragment_count: number; total_content_bytes: number;
  snapshot_digest: string; created_at: string; created_by_actor_ref: string; request_id: string; audit_ref: string;
  fragments: FragmentDocument[];
};
type ResourceDocument = {
  snapshot_id: string; tenant_ref: string; workspace_id: string; application_id: string; snapshot_key: string;
  display_name: string; lifecycle_state: string; latest_version: number; latest_rag_ref: string; latest_digest: string;
  fragment_count: number; total_content_bytes: number; created_at: string; updated_at: string; archived_at: string | null;
};
type OperationEnvelope = {
  request_id: string; tenant_ref: string; workspace_id: string; application_id: string; record: SnapshotRecordDocument | null;
  failure_code: string | null; current_latest_version: number; current_lifecycle_state: string; audit_ref: string;
};
type ListEnvelope = {
  request_id: string; tenant_ref: string; workspace_id: string; application_id: string; items: ResourceDocument[];
  next_cursor: string | null; failure_code: string | null; audit_ref: string;
};

export function readWorkflowRAGSnapshotConfig(): WorkflowRAGSnapshotConfig {
  const env = import.meta.env as Record<string, string | undefined>;
  const requestedScopes = new Set((env.VITE_RADISHMIND_WORKFLOW_RAG_SCOPES ?? "").split(",").map((value) => value.trim()).filter(isWorkflowRAGScope));
  return {
    mode: env.VITE_RADISHMIND_WORKFLOW_RAG_SOURCE?.trim() === DEV_SOURCE ? "dev_workflow_rag_http" : "offline",
    baseUrl: normalizeBaseUrl(env.VITE_RADISHMIND_WORKFLOW_RAG_BASE_URL ?? env.VITE_RADISHMIND_CONTROL_PLANE_READ_BASE_URL ?? DEFAULT_BASE_URL),
    tenantRef: env.VITE_RADISHMIND_DEV_READ_TENANT_REF?.trim() || "tenant_demo",
    workspaceId: env.VITE_RADISHMIND_WORKFLOW_RAG_WORKSPACE_ID?.trim() || "workspace_demo",
    subjectRef: env.VITE_RADISHMIND_DEV_READ_SUBJECT_REF?.trim() || "subject_demo_user",
    authMode: normalizeAuthMode(env.VITE_RADISHMIND_READ_AUTH_MODE),
    scopes: requestedScopes,
  };
}

export function validateWorkflowRAGSnapshotWriteInput(input: WorkflowRAGSnapshotWriteInput): string {
  if (!SNAPSHOT_KEY_PATTERN.test(input.snapshotKey.trim()) || input.displayName.trim().length < 2 || input.displayName.trim().length > 120 ||
    !CLASSIFICATIONS.includes(input.contentClassification) || input.fragments.length < 1 || input.fragments.length > 256 ||
    containsSecretMaterial(input.displayName)) return "workflow_rag_snapshot_payload_invalid";
  const refs = new Set<string>();
  let totalBytes = 0;
  for (const fragment of input.fragments) {
    const contentBytes = new TextEncoder().encode(fragment.content.trim()).length;
    if (!FRAGMENT_REF_PATTERN.test(fragment.fragmentRef.trim()) || refs.has(fragment.fragmentRef.trim()) ||
      !SOURCE_TYPES.includes(fragment.sourceType) || !isReference(fragment.sourceRef) || fragment.sourceRef.includes("://") ||
      !/^[a-z0-9][a-z0-9._/-]{0,119}$/u.test(fragment.pageSlug.trim()) || fragment.title.trim().length > 160 ||
      !fragment.content.trim() || contentBytes > 8192) return "workflow_rag_fragment_invalid";
    if (containsSecretMaterial([fragment.sourceRef, fragment.pageSlug, fragment.title, fragment.content].join("\n"))) {
      return "workflow_rag_secret_material_forbidden";
    }
    refs.add(fragment.fragmentRef.trim());
    totalBytes += contentBytes;
  }
  return totalBytes > 1048576 ? "workflow_rag_budget_exceeded" : "";
}

export async function listWorkflowRAGSnapshots(config: WorkflowRAGSnapshotConfig, applicationId: string, lifecycle: WorkflowRAGSnapshotLifecycle, cursor = ""): Promise<WorkflowRAGSnapshotListResult> {
  const boundary = readBoundary(config, applicationId);
  if (boundary) return listBoundaryResult(boundary);
  const query = new URLSearchParams({ workspace_id: config.workspaceId, application_id: applicationId, lifecycle_state: lifecycle, limit: "100" });
  if (cursor) {
    if (!SNAPSHOT_KEY_PATTERN.test(cursor)) return failedListResult("workflow_rag_snapshot_payload_invalid");
    query.set("cursor", cursor);
  }
  try {
    const response = await fetch(`${config.baseUrl}${COLLECTION_PATH}?${query}`, { headers: buildWorkflowRAGRequestHeaders(config, applicationId, ["workflow_rag_snapshots:read"], "list") });
    const value: unknown = await response.json();
    if (!isListEnvelope(value, config, applicationId, lifecycle)) return failedListResult();
    if (value.failure_code) return { ...failedListResult(value.failure_code), status: value.failure_code === "workflow_rag_snapshot_scope_denied" ? "scope_denied" : "failed" };
    const records = value.items.map(mapResource);
    return { status: records.length ? "ready" : "empty", records, nextCursor: value.next_cursor ?? "", failureCode: "", summary: `Loaded ${records.length} ${lifecycle} knowledge snapshots.` };
  } catch {
    return failedListResult();
  }
}

export async function createWorkflowRAGSnapshot(config: WorkflowRAGSnapshotConfig, applicationId: string, input: WorkflowRAGSnapshotWriteInput): Promise<WorkflowRAGSnapshotOperationResult> {
  const boundary = writeBoundary(config, applicationId, "workflow_rag_snapshots:write");
  if (boundary) return operationBoundaryResult(boundary);
  const failure = validateWorkflowRAGSnapshotWriteInput(input);
  if (failure) return localFailure(failure);
  return writeSnapshot(config, applicationId, COLLECTION_PATH, "create", {
    workspace_id: config.workspaceId, application_id: applicationId, ...mapCreateInput(input),
  }, "created");
}

export async function readWorkflowRAGSnapshot(config: WorkflowRAGSnapshotConfig, applicationId: string, snapshotId: string, version: number): Promise<WorkflowRAGSnapshotOperationResult> {
  const boundary = readBoundary(config, applicationId);
  if (boundary) return operationBoundaryResult(boundary);
  if (!SNAPSHOT_ID_PATTERN.test(snapshotId) || !Number.isInteger(version) || version < 1) return localFailure("workflow_rag_snapshot_payload_invalid");
  const query = new URLSearchParams({ workspace_id: config.workspaceId, application_id: applicationId, snapshot_version: String(version) });
  try {
    const response = await fetch(`${config.baseUrl}${COLLECTION_PATH}/${encodeURIComponent(snapshotId)}?${query}`, {
      headers: buildWorkflowRAGRequestHeaders(config, applicationId, ["workflow_rag_snapshots:read"], "read"),
    });
    const value: unknown = await response.json();
    return mapOperationEnvelope(value, config, applicationId, "loaded");
  } catch {
    return failedOperationResult();
  }
}

export async function versionWorkflowRAGSnapshot(config: WorkflowRAGSnapshotConfig, applicationId: string, snapshotId: string, expectedLatestVersion: number, input: WorkflowRAGSnapshotWriteInput): Promise<WorkflowRAGSnapshotOperationResult> {
  const boundary = writeBoundary(config, applicationId, "workflow_rag_snapshots:write");
  if (boundary) return operationBoundaryResult(boundary);
  const failure = validateWorkflowRAGSnapshotWriteInput(input);
  if (!SNAPSHOT_ID_PATTERN.test(snapshotId) || !Number.isInteger(expectedLatestVersion) || expectedLatestVersion < 1 || failure) {
    return localFailure(failure || "workflow_rag_snapshot_payload_invalid");
  }
  return writeSnapshot(config, applicationId, `${COLLECTION_PATH}/${encodeURIComponent(snapshotId)}/versions`, "version", {
    workspace_id: config.workspaceId, application_id: applicationId, expected_latest_version: expectedLatestVersion, ...mapVersionInput(input),
  }, "versioned");
}

export async function archiveWorkflowRAGSnapshot(config: WorkflowRAGSnapshotConfig, applicationId: string, snapshotId: string, expectedLatestVersion: number): Promise<WorkflowRAGSnapshotOperationResult> {
  const boundary = writeBoundary(config, applicationId, "workflow_rag_snapshots:archive");
  if (boundary) return operationBoundaryResult(boundary);
  if (!SNAPSHOT_ID_PATTERN.test(snapshotId) || !Number.isInteger(expectedLatestVersion) || expectedLatestVersion < 1) return localFailure("workflow_rag_snapshot_payload_invalid");
  return writeSnapshot(config, applicationId, `${COLLECTION_PATH}/${encodeURIComponent(snapshotId)}/archive`, "archive", {
    workspace_id: config.workspaceId, application_id: applicationId, expected_latest_version: expectedLatestVersion,
  }, "archived");
}

async function writeSnapshot(config: WorkflowRAGSnapshotConfig, applicationId: string, path: string, operation: string, body: unknown, success: "created" | "versioned" | "archived") {
  try {
    const response = await fetch(`${config.baseUrl}${path}`, {
      method: "POST", headers: { ...buildWorkflowRAGRequestHeaders(config, applicationId, [success === "archived" ? "workflow_rag_snapshots:archive" : "workflow_rag_snapshots:write"], operation), "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });
    const value: unknown = await response.json();
    return mapOperationEnvelope(value, config, applicationId, success);
  } catch {
    return failedOperationResult();
  }
}

function mapOperationEnvelope(value: unknown, config: WorkflowRAGSnapshotConfig, applicationId: string, success: "created" | "loaded" | "versioned" | "archived"): WorkflowRAGSnapshotOperationResult {
  if (!isOperationEnvelope(value, config, applicationId)) return failedOperationResult();
  if (value.failure_code) {
    const status = value.failure_code === "workflow_rag_snapshot_version_conflict" ? "version_conflict" : value.failure_code === "workflow_rag_snapshot_scope_denied" ? "scope_denied" : "failed";
    return { status, record: null, failureCode: value.failure_code, currentLatestVersion: value.current_latest_version, currentLifecycleState: value.current_lifecycle_state, summary: "Knowledge snapshot operation failed without fallback." };
  }
  return { status: success, record: mapRecord(value.record!), failureCode: "", currentLatestVersion: value.current_latest_version, currentLifecycleState: value.current_lifecycle_state, summary: `Knowledge snapshot ${success}.` };
}

function isOperationEnvelope(value: unknown, config: WorkflowRAGSnapshotConfig, applicationId: string): value is OperationEnvelope {
  if (!isRecord(value) || !hasOnlyKeys(value, OPERATION_ENVELOPE_KEYS) || containsForbiddenMaterial(value, false) ||
    !isNonEmptyString(value.request_id) || value.tenant_ref !== config.tenantRef || value.workspace_id !== config.workspaceId || value.application_id !== applicationId ||
    !(value.failure_code === null || isNonEmptyString(value.failure_code)) || !isInteger(value.current_latest_version, 0) ||
    typeof value.current_lifecycle_state !== "string" || !isNonEmptyString(value.audit_ref)) return false;
  if (value.record === null) return value.failure_code !== null;
  return value.failure_code === null && isSnapshotRecord(value.record, config, applicationId) && value.current_latest_version >= value.record.snapshot_version;
}

function isListEnvelope(value: unknown, config: WorkflowRAGSnapshotConfig, applicationId: string, lifecycle: WorkflowRAGSnapshotLifecycle): value is ListEnvelope {
  return isRecord(value) && hasOnlyKeys(value, LIST_ENVELOPE_KEYS) && !containsForbiddenMaterial(value, true) &&
    isNonEmptyString(value.request_id) && value.tenant_ref === config.tenantRef && value.workspace_id === config.workspaceId && value.application_id === applicationId &&
    Array.isArray(value.items) && value.items.every((item) => isResource(item, config, applicationId, lifecycle)) &&
    (value.next_cursor === null || SNAPSHOT_KEY_PATTERN.test(String(value.next_cursor))) &&
    (value.failure_code === null || isNonEmptyString(value.failure_code)) && isNonEmptyString(value.audit_ref);
}

function isResource(value: unknown, config: WorkflowRAGSnapshotConfig, applicationId: string, lifecycle: WorkflowRAGSnapshotLifecycle): value is ResourceDocument {
  return isRecord(value) && hasOnlyKeys(value, RESOURCE_KEYS) && !containsForbiddenMaterial(value, true) &&
    SNAPSHOT_ID_PATTERN.test(String(value.snapshot_id)) && value.tenant_ref === config.tenantRef && value.workspace_id === config.workspaceId &&
    value.application_id === applicationId && SNAPSHOT_KEY_PATTERN.test(String(value.snapshot_key)) && isNonEmptyString(value.display_name) &&
    value.lifecycle_state === lifecycle && isInteger(value.latest_version, 1) && isRAGRef(value.latest_rag_ref, value.snapshot_key, value.latest_version) &&
    DIGEST_PATTERN.test(String(value.latest_digest)) && isInteger(value.fragment_count, 1, 256) && isInteger(value.total_content_bytes, 1, 1048576) &&
    isTimestamp(value.created_at) && isTimestamp(value.updated_at) && (lifecycle === "active" ? value.archived_at === null : isTimestamp(value.archived_at));
}

function isSnapshotRecord(value: unknown, config: WorkflowRAGSnapshotConfig, applicationId: string): value is SnapshotRecordDocument {
  return isRecord(value) && hasOnlyKeys(value, RECORD_KEYS) && !containsForbiddenMaterial(value, false) && value.schema_version === SNAPSHOT_SCHEMA &&
    SNAPSHOT_ID_PATTERN.test(String(value.snapshot_id)) && value.tenant_ref === config.tenantRef && value.workspace_id === config.workspaceId &&
    value.application_id === applicationId && SNAPSHOT_KEY_PATTERN.test(String(value.snapshot_key)) && isInteger(value.snapshot_version, 1) &&
    isRAGRef(value.rag_ref, value.snapshot_key, value.snapshot_version) && isNonEmptyString(value.display_name) &&
    (value.lifecycle_state === "active" || value.lifecycle_state === "archived") && CLASSIFICATIONS.includes(value.content_classification as WorkflowRAGContentClassification) &&
    value.profile_ref === "workflow.rag.lexical-ngram-dev.v1" && isInteger(value.fragment_count, 1, 256) && isInteger(value.total_content_bytes, 1, 1048576) &&
    DIGEST_PATTERN.test(String(value.snapshot_digest)) && isTimestamp(value.created_at) && isScopeIdentifier(value.created_by_actor_ref) &&
    isNonEmptyString(value.request_id) && isNonEmptyString(value.audit_ref) && Array.isArray(value.fragments) && value.fragments.length === value.fragment_count &&
    value.fragments.every((fragment) => isFragment(fragment, value.content_classification));
}

function isFragment(value: unknown, classification: unknown): value is FragmentDocument {
  return isRecord(value) && hasOnlyKeys(value, FRAGMENT_KEYS) && value.schema_version === FRAGMENT_SCHEMA &&
    FRAGMENT_REF_PATTERN.test(String(value.fragment_ref)) && SOURCE_TYPES.includes(value.source_type as WorkflowRAGSourceType) &&
    isReference(value.source_ref) && !String(value.source_ref).includes("://") && /^[a-z0-9][a-z0-9._/-]{0,119}$/u.test(String(value.page_slug)) &&
    typeof value.title === "string" && value.title.length <= 160 && typeof value.is_official === "boolean" &&
    isNonEmptyString(value.content) && !containsSecretMaterial(value.content) && value.content_classification === classification &&
    isInteger(value.content_bytes, 1, 8192) && new TextEncoder().encode(value.content).length === value.content_bytes && DIGEST_PATTERN.test(String(value.content_digest));
}

function mapResource(value: ResourceDocument): WorkflowRAGSnapshotResource {
  return { snapshotId: value.snapshot_id, tenantRef: value.tenant_ref, workspaceId: value.workspace_id, applicationId: value.application_id, snapshotKey: value.snapshot_key, displayName: value.display_name, lifecycleState: value.lifecycle_state as WorkflowRAGSnapshotLifecycle, latestVersion: value.latest_version, latestRAGRef: value.latest_rag_ref, latestDigest: value.latest_digest, fragmentCount: value.fragment_count, totalContentBytes: value.total_content_bytes, createdAt: value.created_at, updatedAt: value.updated_at, archivedAt: value.archived_at };
}

function mapRecord(value: SnapshotRecordDocument): WorkflowRAGSnapshotRecord {
  return { schemaVersion: SNAPSHOT_SCHEMA, snapshotId: value.snapshot_id, tenantRef: value.tenant_ref, workspaceId: value.workspace_id, applicationId: value.application_id, snapshotKey: value.snapshot_key, ragRef: value.rag_ref, snapshotVersion: value.snapshot_version, displayName: value.display_name, lifecycleState: value.lifecycle_state as WorkflowRAGSnapshotLifecycle, contentClassification: value.content_classification as WorkflowRAGContentClassification, profileRef: "workflow.rag.lexical-ngram-dev.v1", fragmentCount: value.fragment_count, totalContentBytes: value.total_content_bytes, snapshotDigest: value.snapshot_digest, createdAt: value.created_at, createdByActorRef: value.created_by_actor_ref, requestId: value.request_id, auditRef: value.audit_ref, fragments: value.fragments.map((fragment) => ({ schemaVersion: FRAGMENT_SCHEMA, fragmentRef: fragment.fragment_ref, sourceType: fragment.source_type as WorkflowRAGSourceType, sourceRef: fragment.source_ref, pageSlug: fragment.page_slug, title: fragment.title, isOfficial: fragment.is_official, content: fragment.content, contentClassification: fragment.content_classification as WorkflowRAGContentClassification, contentBytes: fragment.content_bytes, contentDigest: fragment.content_digest })) };
}

function mapCreateInput(input: WorkflowRAGSnapshotWriteInput) {
  return { snapshot_key: input.snapshotKey.trim(), display_name: input.displayName.trim(), content_classification: input.contentClassification, fragments: input.fragments.map((fragment) => ({ fragment_ref: fragment.fragmentRef.trim(), source_type: fragment.sourceType, source_ref: fragment.sourceRef.trim(), page_slug: fragment.pageSlug.trim(), title: fragment.title.trim(), is_official: fragment.isOfficial, content: fragment.content.trim() })) };
}

function mapVersionInput(input: WorkflowRAGSnapshotWriteInput) {
  const createInput = mapCreateInput(input);
  return { display_name: createInput.display_name, content_classification: createInput.content_classification, fragments: createInput.fragments };
}

export function buildWorkflowRAGRequestHeaders(
  config: WorkflowRAGSnapshotConfig,
  applicationId: string,
  scopes: readonly WorkflowRAGScope[],
  operation: string,
): Record<string, string> {
  const requestId = `workflow-rag-${operation}-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 10)}`;
  if (config.authMode !== "dev_headers") {
    const provider = config.authMode === "signed_test_token"
      ? (globalThis as typeof globalThis & { __RADISHMIND_CONTROL_PLANE_SIGNED_TEST_TOKEN__?: () => string }).__RADISHMIND_CONTROL_PLANE_SIGNED_TEST_TOKEN__
      : (globalThis as typeof globalThis & { __RADISHMIND_CONTROL_PLANE_OIDC_INTEGRATION_TOKEN__?: () => string }).__RADISHMIND_CONTROL_PLANE_OIDC_INTEGRATION_TOKEN__;
    const token = provider?.().trim() ?? "";
    if (!token) throw new Error("workflow RAG auth token is unavailable in browser memory");
    return { Accept: "application/json", "X-Request-Id": requestId, Authorization: `Bearer ${token}`, "X-RadishMind-Dev-Workflow-Workspace": config.workspaceId, "X-RadishMind-Dev-Workflow-Application": applicationId };
  }
  return { Accept: "application/json", "X-Request-Id": requestId, "X-RadishMind-Dev-Read-Identity": "radishmind-web-workflow-rag-dev", "X-RadishMind-Dev-Read-Tenant": config.tenantRef, "X-RadishMind-Dev-Read-Subject": config.subjectRef, "X-RadishMind-Dev-Read-Scopes": scopes.join(","), "X-RadishMind-Dev-Read-Audit": `audit-${requestId}`, "X-RadishMind-Dev-Workflow-Workspace": config.workspaceId, "X-RadishMind-Dev-Workflow-Application": applicationId };
}

function readBoundary(config: WorkflowRAGSnapshotConfig, applicationId: string): "offline" | "scope_denied" | "" {
  if (config.mode === "offline") return "offline";
  if (!SCOPE_ID_PATTERN.test(applicationId) || !config.scopes.has("workflow_rag_snapshots:read")) return "scope_denied";
  return "";
}
function writeBoundary(config: WorkflowRAGSnapshotConfig, applicationId: string, scope: WorkflowRAGSnapshotScope) {
  if (config.mode === "offline") return "offline" as const;
  if (!SCOPE_ID_PATTERN.test(applicationId) || !config.scopes.has(scope)) return "scope_denied" as const;
  return "" as const;
}
function listBoundaryResult(boundary: "offline" | "scope_denied"): WorkflowRAGSnapshotListResult { return { status: boundary, records: [], nextCursor: "", failureCode: boundary === "offline" ? "workflow_rag_snapshot_http_disabled" : WorkflowRAGScopeFailure, summary: boundary === "offline" ? "Offline mode sends zero knowledge snapshot requests." : "Knowledge snapshot read scope is unavailable; zero requests were sent." }; }
const WorkflowRAGScopeFailure = "workflow_rag_snapshot_scope_denied";
function operationBoundaryResult(boundary: "offline" | "scope_denied"): WorkflowRAGSnapshotOperationResult { return { status: boundary, record: null, failureCode: boundary === "offline" ? "workflow_rag_snapshot_http_disabled" : WorkflowRAGScopeFailure, currentLatestVersion: 0, currentLifecycleState: "", summary: boundary === "offline" ? "Offline mode sends zero knowledge snapshot requests." : "Required knowledge snapshot scope is unavailable; zero requests were sent." }; }
function localFailure(failureCode: string): WorkflowRAGSnapshotOperationResult { return { status: "failed", record: null, failureCode, currentLatestVersion: 0, currentLifecycleState: "", summary: "Knowledge snapshot input was rejected before any request." }; }
function failedOperationResult(failureCode = "workflow_rag_store_unavailable"): WorkflowRAGSnapshotOperationResult { return { status: "failed", record: null, failureCode, currentLatestVersion: 0, currentLifecycleState: "", summary: "Knowledge snapshot operation failed without trusted response or fallback." }; }
function failedListResult(failureCode = "workflow_rag_store_unavailable"): WorkflowRAGSnapshotListResult { return { status: "failed", records: [], nextCursor: "", failureCode, summary: "Knowledge snapshot list failed without fallback." }; }
function containsForbiddenMaterial(value: unknown, forbidContent: boolean): boolean { if (typeof value === "string") return containsSecretMaterial(value); if (Array.isArray(value)) return value.some((item) => containsForbiddenMaterial(item, forbidContent)); if (!isRecord(value)) return false; return Object.entries(value).some(([key, nested]) => (forbidContent && key === "content") || FORBIDDEN_MATERIAL_KEYS.has(key.toLowerCase()) || containsForbiddenMaterial(nested, forbidContent)); }
function containsSecretMaterial(value: string): boolean { return /authorization:|bearer\s|api[_-]?key\s*[:=]|x-radishmind-dev-|cookie:|password\s*=|secret\s*=|token\s*=|sk-[a-z0-9]|-----begin private key-----|(?:postgres(?:ql)?|mysql|mongodb):\/\//iu.test(value); }
function isWorkflowRAGScope(value: string): value is WorkflowRAGScope { return value === "workflow_rag_snapshots:read" || value === "workflow_rag_snapshots:write" || value === "workflow_rag_snapshots:archive" || value === "workflow_rag:execute" || value === "workflow_runs:execute" || value === "workflow_drafts:read"; }
function isRAGRef(value: unknown, key: unknown, version: unknown): boolean { return value === `workflow.rag.${String(key)}.v${String(version)}`; }
function isReference(value: unknown): value is string { return typeof value === "string" && /^[A-Za-z0-9][A-Za-z0-9_.:/-]{2,159}$/u.test(value); }
function isScopeIdentifier(value: unknown): value is string { return typeof value === "string" && SCOPE_ID_PATTERN.test(value); }
function isTimestamp(value: unknown): value is string { return typeof value === "string" && value.length >= 20 && Number.isFinite(Date.parse(value)); }
function isInteger(value: unknown, minimum: number, maximum = Number.MAX_SAFE_INTEGER): value is number { return typeof value === "number" && Number.isInteger(value) && value >= minimum && value <= maximum; }
function isNonEmptyString(value: unknown): value is string { return typeof value === "string" && value.trim().length > 0; }
function isRecord(value: unknown): value is Record<string, unknown> { return typeof value === "object" && value !== null && !Array.isArray(value); }
function hasOnlyKeys(value: Record<string, unknown>, allowed: readonly string[]): boolean { const expected = new Set(allowed); return Object.keys(value).length === allowed.length && Object.keys(value).every((key) => expected.has(key)); }
function normalizeBaseUrl(value: string): string { const normalized = value.trim() || DEFAULT_BASE_URL; return normalized.endsWith("/") ? normalized.slice(0, -1) : normalized; }
function normalizeAuthMode(value: string | undefined): WorkflowRAGSnapshotConfig["authMode"] { return value?.trim() === "signed_test_token" || value?.trim() === "radish_oidc_integration_test" ? value.trim() as WorkflowRAGSnapshotConfig["authMode"] : "dev_headers"; }
