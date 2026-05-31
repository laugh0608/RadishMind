import {
  CONTROL_PLANE_READ_FORBIDDEN_OUTPUT_KEYS,
  CONTROL_PLANE_READ_ROUTE_DEFINITIONS,
  CONTROL_PLANE_READ_ROUTES,
  controlPlaneReadResponseHasForbiddenOutput,
  toControlPlaneReadCollectionViewModel,
  type APIKeySummary,
  type ControlPlaneReadCollectionViewModel,
  type ControlPlaneReadResponseByRoute,
} from "../../../../../contracts/typescript/control-plane-read-api";

export type WorkspaceApiKeysStateId =
  | "ready"
  | "empty"
  | "denied"
  | "stale"
  | "partial_failure"
  | "forbidden_projection";

export type WorkspaceApiKeysMetric = {
  label: string;
  value: string;
  detail: string;
};

export type WorkspaceApiKeyRow = {
  apiKeyId: string;
  ownerSubjectRef: string;
  scopes: string[];
  state: string;
  createdAt: string;
  expiresAt: string | null;
  lastUsedAt: string | null;
};

export type WorkspaceApiKeysStatePreview = {
  id: WorkspaceApiKeysStateId;
  label: string;
  status: string;
  summary: string;
  itemCount: number;
  failureCode: string;
};

export type WorkspaceApiKeysViewModel = {
  pageId: "workspace-api-keys";
  routeId: "api-key-summary-list-route";
  routePath: typeof CONTROL_PLANE_READ_ROUTES.apiKeys;
  readModel: "api-key-summary";
  requiredScope: "api_keys:read";
  collection: ControlPlaneReadCollectionViewModel;
  apiKeys: WorkspaceApiKeyRow[];
  metrics: WorkspaceApiKeysMetric[];
  statePreviews: WorkspaceApiKeysStatePreview[];
  auditRef: string;
  requestId: string;
  nextCursor: string | null;
  forbiddenProjectionBlocked: boolean;
  canRenderApiKeys: boolean;
  canRequestLiveBackend: false;
  canMutate: false;
};

export function buildWorkspaceApiKeysViewModel(): WorkspaceApiKeysViewModel {
  const route = CONTROL_PLANE_READ_ROUTE_DEFINITIONS["api-key-summary-list-route"];
  const collection = toControlPlaneReadCollectionViewModel("api-key-summary-list-route", buildApiKeysEnvelope());
  const apiKeys = collection.items.map((item) => toApiKeyRow(item as APIKeySummary));
  const forbiddenProjectionBlocked = controlPlaneReadResponseHasForbiddenOutput({
    items: [
      {
        [CONTROL_PLANE_READ_FORBIDDEN_OUTPUT_KEYS[1]]: "blocked",
        [CONTROL_PLANE_READ_FORBIDDEN_OUTPUT_KEYS[2]]: "blocked",
      },
    ],
  });

  return {
    pageId: "workspace-api-keys",
    routeId: "api-key-summary-list-route",
    routePath: CONTROL_PLANE_READ_ROUTES.apiKeys,
    readModel: "api-key-summary",
    requiredScope: "api_keys:read",
    collection,
    apiKeys,
    metrics: buildMetrics(collection, apiKeys),
    statePreviews: buildStatePreviews(collection, forbiddenProjectionBlocked),
    auditRef: collection.auditRef,
    requestId: collection.requestId,
    nextCursor: collection.nextCursor,
    forbiddenProjectionBlocked,
    canRenderApiKeys:
      route.path === CONTROL_PLANE_READ_ROUTES.apiKeys &&
      route.canMutate === false &&
      collection.canRenderItems &&
      !controlPlaneReadResponseHasForbiddenOutput(collection),
    canRequestLiveBackend: false,
    canMutate: false,
  };
}

function buildApiKeysEnvelope(): ControlPlaneReadResponseByRoute["api-key-summary-list-route"] {
  return {
    request_id: "req_workspace_api_keys_demo",
    tenant_ref: "tenant_demo",
    items: [
      {
        api_key_id: "key_ops_readonly",
        tenant_ref: "tenant_demo",
        owner_subject_ref: "subject_platform_ops",
        scopes: ["models:read", "runs:read"],
        state: "active",
        created_at: "2026-05-20T08:00:00Z",
        expires_at: "2026-08-20T08:00:00Z",
        last_used_at: "2026-05-31T08:45:00Z",
      },
      {
        api_key_id: "key_docs_pending_rotation",
        tenant_ref: "tenant_demo",
        owner_subject_ref: "subject_docs_team",
        scopes: ["applications:read"],
        state: "rotation_required",
        created_at: "2026-04-12T11:10:00Z",
        expires_at: null,
        last_used_at: null,
      },
    ],
    next_cursor: null,
    failure_code: null,
    audit_ref: "audit_workspace_api_keys_demo",
  };
}

function toApiKeyRow(apiKey: APIKeySummary): WorkspaceApiKeyRow {
  return {
    apiKeyId: apiKey.api_key_id,
    ownerSubjectRef: apiKey.owner_subject_ref,
    scopes: apiKey.scopes,
    state: apiKey.state,
    createdAt: apiKey.created_at,
    expiresAt: apiKey.expires_at,
    lastUsedAt: apiKey.last_used_at,
  };
}

function buildMetrics(
  collection: ControlPlaneReadCollectionViewModel,
  apiKeys: WorkspaceApiKeyRow[],
): WorkspaceApiKeysMetric[] {
  const activeCount = apiKeys.filter((apiKey) => apiKey.state === "active").length;
  const rotationRequiredCount = apiKeys.filter((apiKey) => apiKey.state === "rotation_required").length;
  const expiringCount = apiKeys.filter((apiKey) => apiKey.expiresAt !== null).length;

  return [
    {
      label: "API keys",
      value: String(collection.itemCount),
      detail: collection.tenantRef,
    },
    {
      label: "Active",
      value: String(activeCount),
      detail: "read-only access references",
    },
    {
      label: "Rotation",
      value: String(rotationRequiredCount),
      detail: "requires operator review",
    },
    {
      label: "Expiring",
      value: String(expiringCount),
      detail: "keys with explicit expiry",
    },
  ];
}

function buildStatePreviews(
  collection: ControlPlaneReadCollectionViewModel,
  forbiddenProjectionBlocked: boolean,
): WorkspaceApiKeysStatePreview[] {
  return [
    {
      id: "ready",
      label: "Ready",
      status: collection.statusLabel,
      summary: "API key summaries render from the offline consumer view model without secret material.",
      itemCount: collection.itemCount,
      failureCode: collection.failureCode ?? "none",
    },
    {
      id: "empty",
      label: "Empty",
      status: "empty",
      summary: "The page keeps route metadata visible when no API key summaries are returned.",
      itemCount: 0,
      failureCode: "none",
    },
    {
      id: "denied",
      label: "Denied",
      status: "scope_denied",
      summary: "Scope denial renders without leaking partial key rows or sensitive key fields.",
      itemCount: 0,
      failureCode: "scope_denied",
    },
    {
      id: "stale",
      label: "Stale",
      status: "stale",
      summary: "Cached API key summaries stay visibly separated from live gateway key lifecycle data.",
      itemCount: collection.itemCount,
      failureCode: "read_store_unavailable",
    },
    {
      id: "partial_failure",
      label: "Partial failure",
      status: "partial_failure",
      summary: "Failure metadata remains explicit when a later key summary page cannot be read.",
      itemCount: collection.itemCount,
      failureCode: "read_store_unavailable",
    },
    {
      id: "forbidden_projection",
      label: "Forbidden projection",
      status: forbiddenProjectionBlocked ? "blocked" : "clear",
      summary: "Sensitive key projections are blocked by the shared forbidden output guard.",
      itemCount: 0,
      failureCode: forbiddenProjectionBlocked ? "read_store_contract_mismatch" : "none",
    },
  ];
}
