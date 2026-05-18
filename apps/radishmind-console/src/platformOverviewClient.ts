import {
  isPlatformOverviewResponse,
  PLATFORM_OVERVIEW_ROUTE,
  toPlatformOverviewConsoleViewModel,
  type PlatformOverviewConsoleViewModel,
  type PlatformOverviewResponse,
} from "../../../contracts/typescript/platform-overview-api.ts";

export const DEFAULT_PLATFORM_BASE_URL = "http://127.0.0.1:8080";

export type PlatformOverviewLoadState =
  | {
      status: "idle" | "loading";
      endpoint: string;
    }
  | {
      status: "ready";
      endpoint: string;
      overview: PlatformOverviewResponse;
      viewModel: PlatformOverviewConsoleViewModel;
      loadedAt: string;
    }
  | {
      status: "error";
      endpoint: string;
      message: string;
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
): Promise<Extract<PlatformOverviewLoadState, { status: "ready" }>> {
  const endpoint = buildPlatformOverviewEndpoint(baseUrl);
  const response = await fetch(endpoint, {
    method: "GET",
    headers: {
      Accept: "application/json",
    },
  });

  if (!response.ok) {
    throw new Error(`overview request failed with HTTP ${response.status}`);
  }

  const document: unknown = await response.json();
  if (!isPlatformOverviewResponse(document)) {
    throw new Error("overview response does not match platform overview contract");
  }

  return {
    status: "ready",
    endpoint,
    overview: document,
    viewModel: toPlatformOverviewConsoleViewModel(document),
    loadedAt: new Date().toISOString(),
  };
}
