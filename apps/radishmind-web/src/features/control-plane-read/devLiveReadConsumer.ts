import {
  CONTROL_PLANE_READ_ROUTE_DEFINITIONS,
  CONTROL_PLANE_READ_ROUTE_IDS,
  controlPlaneReadResponseHasForbiddenOutput,
  isControlPlaneReadEnvelope,
  toControlPlaneReadCollectionViewModel,
  type ControlPlaneReadCollectionViewModel,
  type ControlPlaneReadRouteId,
  type ControlPlaneReadSummaryItem,
} from "../../../../../contracts/typescript/control-plane-read-api.ts";

const DEV_LIVE_SOURCE = "dev-live-http";
const DEFAULT_BASE_URL = "http://127.0.0.1:7000";
const DEFAULT_TENANT_REF = "tenant_demo";
const DEFAULT_SUBJECT_REF = "subject_demo_user";
const DEFAULT_SCOPES = "tenant:read,applications:read,api_keys:read,usage:read,runs:read,audit:read";
const ACTIVE_WORKSPACE_PATTERN = /^[A-Za-z0-9][A-Za-z0-9_.:/-]{0,159}$/u;
const WORKSPACE_ROUTE_IDS = new Set<ControlPlaneReadRouteId>([
  "application-summary-list-route",
  "api-key-summary-list-route",
  "quota-summary-route",
  "workflow-definition-summary-list-route",
  "run-record-summary-list-route",
]);

export type ControlPlaneReadDataSourceMode = "offline_fixture" | "dev_live_http";
export type ControlPlaneReadAuthMode = "dev_headers" | "signed_test_token" | "radish_oidc_integration_test";
export type ControlPlaneReadStoreMode = "fake_store_dev" | "postgres_dev_test";

export type ControlPlaneReadDevLiveConfig = {
  mode: ControlPlaneReadDataSourceMode;
  baseUrl: string;
  tenantRef: string;
  subjectRef: string;
  authMode?: ControlPlaneReadAuthMode;
  storeMode?: ControlPlaneReadStoreMode;
  workspaceId?: string;
};

export type ControlPlaneReadPageRequest = {
  cursor?: string;
  limit?: number;
  sort?: "recorded_at_desc";
};

export type ControlPlaneReadDevLiveLoadState =
  | {
      status: "idle";
      mode: "offline_fixture";
      message: string;
    }
  | {
      status: "loading";
      mode: "dev_live_http";
      message: string;
    }
  | {
      status: "ready";
      mode: "dev_live_http";
      message: string;
      collections: Partial<Record<ControlPlaneReadRouteId, ControlPlaneReadCollectionViewModel>>;
    }
  | {
      status: "failed";
      mode: "dev_live_http";
      message: string;
    };

export function readControlPlaneReadDevLiveConfig(): ControlPlaneReadDevLiveConfig {
  const env = import.meta.env as Record<string, string | undefined>;
  const source = env.VITE_RADISHMIND_READ_SOURCE?.trim();
  return {
    mode: source === DEV_LIVE_SOURCE ? "dev_live_http" : "offline_fixture",
    baseUrl: normalizeBaseUrl(env.VITE_RADISHMIND_CONTROL_PLANE_READ_BASE_URL ?? DEFAULT_BASE_URL),
    tenantRef: env.VITE_RADISHMIND_DEV_READ_TENANT_REF?.trim() || DEFAULT_TENANT_REF,
    subjectRef: env.VITE_RADISHMIND_DEV_READ_SUBJECT_REF?.trim() || DEFAULT_SUBJECT_REF,
    authMode: normalizeAuthMode(env.VITE_RADISHMIND_READ_AUTH_MODE),
    storeMode: env.VITE_RADISHMIND_READ_STORE_MODE?.trim() === "postgres_dev_test" ? "postgres_dev_test" : "fake_store_dev",
    workspaceId: normalizeActiveWorkspaceId(
      env.VITE_RADISHMIND_ACTIVE_WORKSPACE_ID ??
      env.VITE_RADISHMIND_APPLICATION_CATALOG_WORKSPACE_ID ??
      "workspace_demo",
    ) ?? "workspace_demo",
  };
}

export function initialControlPlaneReadDevLiveLoadState(
  config: ControlPlaneReadDevLiveConfig,
): ControlPlaneReadDevLiveLoadState {
  if (config.mode !== "dev_live_http") {
    return {
      status: "idle",
      mode: "offline_fixture",
      message: "Offline fixture view models are the default verification baseline.",
    };
  }
  return {
    status: "loading",
    mode: "dev_live_http",
    message: config.authMode === "radish_oidc_integration_test"
      ? "Loading Admin reads while workspace operations remain membership-blocked."
      : config.storeMode === "postgres_dev_test"
      ? "Loading routed PostgreSQL workspace projections over signed development/test HTTP."
      : "Loading development/test workspace owner projections over HTTP.",
  };
}

export async function loadControlPlaneReadDevLiveCollections(
  config: ControlPlaneReadDevLiveConfig,
): Promise<Partial<Record<ControlPlaneReadRouteId, ControlPlaneReadCollectionViewModel>>> {
  if (config.mode !== "dev_live_http") {
    return {};
  }
  const entries = await Promise.all(
    CONTROL_PLANE_READ_ROUTE_IDS.map(async (routeId) => {
      const collection = await loadControlPlaneReadDevLiveCollection(config, routeId);
      return [routeId, collection] as const;
    }),
  );
  return Object.fromEntries(entries) as Partial<
    Record<ControlPlaneReadRouteId, ControlPlaneReadCollectionViewModel>
  >;
}

export async function loadControlPlaneReadDevLiveCollection(
  config: ControlPlaneReadDevLiveConfig,
  routeId: ControlPlaneReadRouteId,
  pageRequest: ControlPlaneReadPageRequest = {},
): Promise<ControlPlaneReadCollectionViewModel> {
  if (config.mode !== "dev_live_http") {
    throw new Error("development/test live read source is disabled");
  }
  const envelope = await fetchDevLiveEnvelope(routeId, config, pageRequest);
  if (controlPlaneReadResponseHasForbiddenOutput(envelope)) {
    throw new Error(`${routeId} returned forbidden read-side output`);
  }
  return toControlPlaneReadCollectionViewModel(routeId, envelope, { source: "dev_live_http" });
}

async function fetchDevLiveEnvelope(
  routeId: ControlPlaneReadRouteId,
  config: ControlPlaneReadDevLiveConfig,
  pageRequest: ControlPlaneReadPageRequest,
) {
  const response = await fetch(devLiveRouteUrl(routeId, config, pageRequest), {
    method: "GET",
    headers: devLiveHeaders(routeId, config),
  });
  const body: unknown = await response.json();
  if (!isControlPlaneReadEnvelope(body)) {
    throw new Error(`${routeId} returned HTTP ${response.status} with a non read-side envelope`);
  }
  return body as Parameters<typeof toControlPlaneReadCollectionViewModel>[1] & {
    items: ControlPlaneReadSummaryItem[];
  };
}

function devLiveRouteUrl(
  routeId: ControlPlaneReadRouteId,
  config: ControlPlaneReadDevLiveConfig,
  pageRequest: ControlPlaneReadPageRequest,
): string {
  const route = CONTROL_PLANE_READ_ROUTE_DEFINITIONS[routeId];
  const path = route.path.replace("{tenant_ref}", encodeURIComponent(config.tenantRef));
  const query = auditPageQuery(routeId, pageRequest);
  return `${config.baseUrl}${path}${query ? `?${query}` : ""}`;
}

function auditPageQuery(
  routeId: ControlPlaneReadRouteId,
  pageRequest: ControlPlaneReadPageRequest,
): string {
  const hasPageRequest = pageRequest.cursor !== undefined ||
    pageRequest.limit !== undefined ||
    pageRequest.sort !== undefined;
  if (!hasPageRequest) {
    return "";
  }
  if (routeId !== "audit-summary-list-route") {
    throw new Error(`${routeId} does not accept cursor pagination`);
  }
  const query = new URLSearchParams();
  if (pageRequest.cursor !== undefined) {
    const cursor = pageRequest.cursor.trim();
    if (!cursor || cursor.length > 1024) {
      throw new Error("audit cursor is invalid");
    }
    query.set("cursor", cursor);
  }
  if (pageRequest.limit !== undefined) {
    if (!Number.isInteger(pageRequest.limit) || pageRequest.limit < 1 || pageRequest.limit > 100) {
      throw new Error("audit limit is invalid");
    }
    query.set("limit", String(pageRequest.limit));
  }
  if (pageRequest.sort !== undefined) {
    query.set("sort", pageRequest.sort);
  }
  return query.toString();
}

function devLiveHeaders(routeId: ControlPlaneReadRouteId, config: ControlPlaneReadDevLiveConfig): HeadersInit {
  const workspaceHeaders = devLiveWorkspaceHeaders(routeId, config);
  if (config.authMode === "radish_oidc_integration_test") {
    const tokenProvider = (
      globalThis as typeof globalThis & {
        __RADISHMIND_CONTROL_PLANE_OIDC_INTEGRATION_TOKEN__?: () => string;
      }
    ).__RADISHMIND_CONTROL_PLANE_OIDC_INTEGRATION_TOKEN__;
    const token = tokenProvider?.().trim() ?? "";
    if (!token) {
      throw new Error("OIDC integration token is unavailable in browser memory");
    }
    return {
      Accept: "application/json",
      "X-Request-Id": `dev-live-${routeId}`,
      Authorization: `Bearer ${token}`,
      ...workspaceHeaders,
    };
  }
  if (config.authMode === "signed_test_token") {
    const tokenProvider = (
      globalThis as typeof globalThis & {
        __RADISHMIND_CONTROL_PLANE_SIGNED_TEST_TOKEN__?: () => string;
      }
    ).__RADISHMIND_CONTROL_PLANE_SIGNED_TEST_TOKEN__;
    const token = tokenProvider?.().trim() ?? "";
    if (!token) {
      throw new Error("signed test token is unavailable in browser memory");
    }
    return {
      Accept: "application/json",
      "X-Request-Id": `dev-live-${routeId}`,
      Authorization: `Bearer ${token}`,
      ...workspaceHeaders,
    };
  }
  return {
    Accept: "application/json",
    "X-Request-Id": `dev-live-${routeId}`,
    "X-RadishMind-Dev-Read-Identity": "dev-live-read-consumer",
    "X-RadishMind-Dev-Read-Tenant": config.tenantRef,
    "X-RadishMind-Dev-Read-Subject": config.subjectRef,
    "X-RadishMind-Dev-Read-Scopes": DEFAULT_SCOPES,
    "X-RadishMind-Dev-Read-Audit": "audit_dev_live_read_consumer",
    ...workspaceHeaders,
  };
}

function devLiveWorkspaceHeaders(
  routeId: ControlPlaneReadRouteId,
  config: ControlPlaneReadDevLiveConfig,
): Record<string, string> {
  if (!WORKSPACE_ROUTE_IDS.has(routeId)) {
    return {};
  }
  const workspaceId = normalizeActiveWorkspaceId(config.workspaceId ?? "workspace_demo");
  if (!workspaceId) {
    throw new Error("active workspace selection is invalid");
  }
  const headers: Record<string, string> = {
    "X-RadishMind-Active-Workspace": workspaceId,
  };
  if ((config.authMode ?? "dev_headers") === "dev_headers") {
    headers["X-RadishMind-Dev-Read-Membership-Workspace"] = workspaceId;
    headers["X-RadishMind-Dev-Read-Membership-Permissions"] =
      CONTROL_PLANE_READ_ROUTE_DEFINITIONS[routeId].requiredScope;
  }
  return headers;
}

export function normalizeActiveWorkspaceId(value: string): string | null {
  const normalized = value.trim();
  return ACTIVE_WORKSPACE_PATTERN.test(normalized) ? normalized : null;
}

function normalizeAuthMode(value: string | undefined): ControlPlaneReadAuthMode {
  const normalized = value?.trim();
  if (normalized === "signed_test_token" || normalized === "radish_oidc_integration_test") {
    return normalized;
  }
  return "dev_headers";
}

function normalizeBaseUrl(baseUrl: string): string {
  const normalized = baseUrl.trim() || DEFAULT_BASE_URL;
  return normalized.endsWith("/") ? normalized.slice(0, -1) : normalized;
}
