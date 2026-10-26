import { parseWorkbenchPublicConfig, type WorkbenchPublicConfig } from "../../../../contracts/typescript/workbench-public-config.ts";

let configuration: WorkbenchPublicConfig | null = null;

export async function initializeWorkbenchConfiguration(
  profile: string | undefined,
  origin: string,
  fetcher: typeof fetch = fetch,
): Promise<void> {
  if (!profile) return;
  if (profile !== "local-workbench") throw new Error("Unsupported Web runtime profile");
  const response = await fetcher(`${origin}/workbench-config.json`, { cache: "no-store", credentials: "same-origin", headers: { Accept: "application/json" } });
  if (!response.ok || !response.headers.get("content-type")?.startsWith("application/json")) throw new Error("Workbench configuration unavailable");
  const text = await response.text();
  if (text.length > 16384) throw new Error("Workbench configuration too large");
  configuration = parseWorkbenchPublicConfig(JSON.parse(text), origin);
}

export function workbenchConfiguration(): WorkbenchPublicConfig | null { return configuration; }
export function isLocalWorkbench(): boolean { return configuration !== null; }

// Existing consumers retain their tested parsers. This explicit projection is
// their configuration input, not a mutation of Vite's build environment.
export function readWebConfigurationEnvironment(): Record<string, string | undefined> {
  if (configuration) return workbenchConsumerConfiguration(configuration);
  if (import.meta.env?.VITE_RADISHMIND_RUNTIME_PROFILE) throw new Error("Workbench configuration must load before consumers");
  return import.meta.env as Record<string, string | undefined>;
}

export function workbenchConsumerConfiguration(config: WorkbenchPublicConfig): Record<string, string> {
  return {
    VITE_RADISHMIND_LOCAL_IDENTITY_MODE: "local_identity_dev",
    VITE_RADISHMIND_LOCAL_IDENTITY_BASE_URL: config.public_origin,
    VITE_RADISHMIND_CONTROL_PLANE_READ_BASE_URL: config.public_origin,
    VITE_RADISHMIND_READ_SOURCE: "dev-live-http",
    VITE_RADISHMIND_READ_AUTH_MODE: "local_session_dev_test",
    VITE_RADISHMIND_READ_STORE_MODE: "sqlite_dev",
    VITE_RADISHMIND_DEV_READ_TENANT_REF: config.tenant_ref,
    VITE_RADISHMIND_ACTIVE_WORKSPACE_ID: config.workspace_id,
    VITE_RADISHMIND_DEV_READ_SUBJECT_REF: "local_session",
    VITE_RADISHMIND_APPLICATION_CATALOG_SOURCE: "dev-application-catalog-http",
    VITE_RADISHMIND_APPLICATION_CATALOG_WORKSPACE_ID: config.workspace_id,
    VITE_RADISHMIND_APPLICATION_DRAFT_SOURCE: "dev-application-draft-http",
    VITE_RADISHMIND_APPLICATION_DRAFT_WORKSPACE_ID: config.workspace_id,
    VITE_RADISHMIND_APPLICATION_PUBLISH_SOURCE: "dev-application-publish-http",
    VITE_RADISHMIND_APPLICATION_PUBLISH_WORKSPACE_ID: config.workspace_id,
    VITE_RADISHMIND_PROMPT_APPLICATION_SOURCE: "dev-prompt-application-http",
    VITE_RADISHMIND_PROMPT_APPLICATION_WORKSPACE_ID: config.workspace_id,
    VITE_RADISHMIND_WORKFLOW_SAVED_DRAFT_SOURCE: "dev-saved-draft-http",
    VITE_RADISHMIND_WORKFLOW_SAVED_DRAFT_WORKSPACE_ID: config.workspace_id,
    VITE_RADISHMIND_WORKFLOW_DEFINITION_PROMOTION_SOURCE: "dev-workflow-definition-promotion-http",
    VITE_RADISHMIND_WORKFLOW_DEFINITION_PROMOTION_WORKSPACE_ID: config.workspace_id,
    VITE_RADISHMIND_WORKFLOW_EXECUTOR_SOURCE: "dev-workflow-executor-http",
    VITE_RADISHMIND_WORKFLOW_EXECUTOR_WORKSPACE_ID: config.workspace_id,
    VITE_RADISHMIND_WORKFLOW_RUN_HISTORY_SOURCE: "dev-workflow-run-history-http",
    VITE_RADISHMIND_WORKFLOW_RUN_HISTORY_WORKSPACE_ID: config.workspace_id,
    VITE_RADISHMIND_APPLICATION_SESSION_SOURCE: "dev-application-session-http",
    VITE_RADISHMIND_APPLICATION_SESSION_WORKSPACE_ID: config.workspace_id,
    VITE_RADISHMIND_API_KEY_LIFECYCLE_SOURCE: "dev-api-key-lifecycle-http",
    VITE_RADISHMIND_API_KEY_LIFECYCLE_WORKSPACE_ID: config.workspace_id,
    VITE_RADISHMIND_GATEWAY_AUTH_MODE: "api_key_dev_test",
  };
}
