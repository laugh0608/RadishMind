import {
  isPlatformOverviewResponse,
  PLATFORM_OVERVIEW_ROUTE,
  toPlatformOverviewConsoleViewModel,
  type PlatformOverviewConsoleViewModel,
  type PlatformOverviewResponse,
} from "../../../contracts/typescript/platform-overview-api.ts";

export const DEFAULT_PLATFORM_BASE_URL = "http://127.0.0.1:8080";

export type PlatformOverviewReadyState = {
  status: "ready";
  endpoint: string;
  overview: PlatformOverviewResponse;
  viewModel: PlatformOverviewConsoleViewModel;
  loadedAt: string;
};

export type PlatformOverviewLoadState =
  | {
      status: "idle";
      endpoint: string;
    }
  | {
      status: "loading";
      endpoint: string;
      previous?: PlatformOverviewReadyState;
    }
  | PlatformOverviewReadyState
  | {
      status: "error";
      endpoint: string;
      message: string;
      diagnostics: string[];
      previous?: PlatformOverviewReadyState;
    };

export function resolvePlatformBaseUrl(): string {
  const configuredBaseUrl = import.meta.env.VITE_RADISHMIND_PLATFORM_BASE_URL;
  if (typeof configuredBaseUrl === "string" && configuredBaseUrl.trim() !== "") {
    return configuredBaseUrl.trim().replace(/\/+$/, "");
  }
  return DEFAULT_PLATFORM_BASE_URL;
}

export function buildPlatformOverviewEndpoint(baseUrl: string): string {
  return `${baseUrl.replace(/\/+$/, "")}${PLATFORM_OVERVIEW_ROUTE}`;
}

export async function loadPlatformOverview(
  baseUrl = resolvePlatformBaseUrl(),
): Promise<PlatformOverviewReadyState> {
  const endpoint = buildPlatformOverviewEndpoint(baseUrl);
  let response: Response;
  try {
    response = await fetch(endpoint, {
      method: "GET",
      headers: {
        Accept: "application/json",
      },
    });
  } catch (error) {
    throw new PlatformOverviewRequestError(
      `could not reach platform overview at ${endpoint}`,
      buildConnectionDiagnostics(endpoint, error),
    );
  }

  if (!response.ok) {
    throw new PlatformOverviewRequestError(`overview request failed with HTTP ${response.status}`, [
      "Confirm the platform service is listening at the configured Platform URL.",
      "Run `pwsh ../../scripts/run-platform-service.ps1 serve` from apps/radishmind-console.",
      "If the service is running, inspect its response body or platform logs for the HTTP failure.",
    ]);
  }

  const document: unknown = await response.json();
  if (!isPlatformOverviewResponse(document)) {
    throw new PlatformOverviewRequestError("overview response does not match platform overview contract", [
      "Confirm the service exposes the current `/v1/platform/overview` schema.",
      "Run `python ../../scripts/run-platform-overview-consumer-smoke.py --base-url <platform-url> --check`.",
      "Rebuild the console after changing `contracts/typescript/platform-overview-api.ts`.",
    ]);
  }

  return {
    status: "ready",
    endpoint,
    overview: document,
    viewModel: toPlatformOverviewConsoleViewModel(document),
    loadedAt: new Date().toISOString(),
  };
}

export class PlatformOverviewRequestError extends Error {
  diagnostics: string[];

  constructor(message: string, diagnostics: string[]) {
    super(message);
    this.name = "PlatformOverviewRequestError";
    this.diagnostics = diagnostics;
  }
}

export function getPlatformOverviewDiagnostics(error: unknown): string[] {
  if (error instanceof PlatformOverviewRequestError) {
    return error.diagnostics;
  }
  return [
    "Confirm the platform service is running and reachable from this browser.",
    "Confirm the configured Platform URL uses the same host and port as the service.",
    "Confirm local CORS allows `http://127.0.0.1:5173` or `http://localhost:5173`.",
  ];
}

function buildConnectionDiagnostics(endpoint: string, error: unknown): string[] {
  const details = error instanceof Error ? error.message : String(error);
  return [
    "Start the platform service with `pwsh ../../scripts/run-platform-service.ps1 serve`.",
    `Confirm ${endpoint} opens and returns JSON.`,
    "If the service is already running, check the local console origin and CORS preflight.",
    `Browser fetch detail: ${details}`,
  ];
}
