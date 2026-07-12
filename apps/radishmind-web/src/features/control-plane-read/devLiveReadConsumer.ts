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

export type ControlPlaneReadDataSourceMode = "offline_fixture" | "dev_live_http";

export type ControlPlaneReadDevLiveConfig = {
  mode: ControlPlaneReadDataSourceMode;
  baseUrl: string;
  tenantRef: string;
  subjectRef: string;
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
    message: "Loading fake-store-backed read routes over dev HTTP.",
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
      const envelope = await fetchDevLiveEnvelope(routeId, config);
      if (controlPlaneReadResponseHasForbiddenOutput(envelope)) {
        throw new Error(`${routeId} returned forbidden read-side output`);
      }
      return [
        routeId,
        toControlPlaneReadCollectionViewModel(routeId, envelope, { source: "dev_live_http" }),
      ] as const;
    }),
  );
  return Object.fromEntries(entries) as Partial<
    Record<ControlPlaneReadRouteId, ControlPlaneReadCollectionViewModel>
  >;
}

async function fetchDevLiveEnvelope(routeId: ControlPlaneReadRouteId, config: ControlPlaneReadDevLiveConfig) {
  const response = await fetch(devLiveRouteUrl(routeId, config), {
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

function devLiveRouteUrl(routeId: ControlPlaneReadRouteId, config: ControlPlaneReadDevLiveConfig): string {
  const route = CONTROL_PLANE_READ_ROUTE_DEFINITIONS[routeId];
  const path = route.path.replace("{tenant_ref}", encodeURIComponent(config.tenantRef));
  return `${config.baseUrl}${path}`;
}

function devLiveHeaders(routeId: ControlPlaneReadRouteId, config: ControlPlaneReadDevLiveConfig): HeadersInit {
  return {
    Accept: "application/json",
    "X-Request-Id": `dev-live-${routeId}`,
    "X-RadishMind-Dev-Read-Identity": "dev-live-read-consumer",
    "X-RadishMind-Dev-Read-Tenant": config.tenantRef,
    "X-RadishMind-Dev-Read-Subject": config.subjectRef,
    "X-RadishMind-Dev-Read-Scopes": DEFAULT_SCOPES,
    "X-RadishMind-Dev-Read-Audit": "audit_dev_live_read_consumer",
  };
}

function normalizeBaseUrl(baseUrl: string): string {
  const normalized = baseUrl.trim() || DEFAULT_BASE_URL;
  return normalized.endsWith("/") ? normalized.slice(0, -1) : normalized;
}
