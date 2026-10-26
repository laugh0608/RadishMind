export const WORKBENCH_FEATURES = ["applications", "prompt", "workflow", "sessions", "results", "api_keys", "local_identity"] as const;
export type WorkbenchFeature = typeof WORKBENCH_FEATURES[number];
export type WorkbenchPublicConfig = {
  schema_version: 1;
  public_origin: string;
  tenant_ref: string;
  workspace_id: string;
  environment: "development" | "test";
  provider_mode: "mock" | "openai-compatible";
  features: WorkbenchFeature[];
};

export function parseWorkbenchPublicConfig(value: unknown, expectedOrigin: string): WorkbenchPublicConfig {
  if (!value || typeof value !== "object" || Array.isArray(value)) throw new Error("Invalid workbench configuration");
  const record = value as Record<string, unknown>;
  const fields = ["schema_version", "public_origin", "tenant_ref", "workspace_id", "environment", "provider_mode", "features"];
  if (Object.keys(record).length !== fields.length || Object.keys(record).some(key => !fields.includes(key))) throw new Error("Unknown workbench configuration field");
  if (record.schema_version !== 1 || typeof record.public_origin !== "string" || record.public_origin !== expectedOrigin) throw new Error("Workbench origin or version mismatch");
  const origin = new URL(record.public_origin);
  if (origin.protocol !== "http:" || !["127.0.0.1", "[::1]"].includes(origin.hostname) || !origin.port || origin.origin !== record.public_origin) throw new Error("Workbench requires a loopback origin");
  if (typeof record.tenant_ref !== "string" || !/^[A-Za-z0-9][A-Za-z0-9_-]{0,127}$/u.test(record.tenant_ref) ||
    typeof record.workspace_id !== "string" || !/^[A-Za-z0-9][A-Za-z0-9_-]{0,127}$/u.test(record.workspace_id)) throw new Error("Invalid workbench scope");
  if (record.environment !== "development" && record.environment !== "test") throw new Error("Invalid workbench environment");
  if (record.provider_mode !== "mock" && record.provider_mode !== "openai-compatible") throw new Error("Invalid workbench provider mode");
  if (!Array.isArray(record.features) || record.features.length !== WORKBENCH_FEATURES.length || new Set(record.features).size !== WORKBENCH_FEATURES.length ||
    !WORKBENCH_FEATURES.every(feature => record.features instanceof Array && record.features.includes(feature))) throw new Error("Unsupported workbench feature set");
  return { schema_version: 1, public_origin: record.public_origin, tenant_ref: record.tenant_ref, workspace_id: record.workspace_id,
    environment: record.environment, provider_mode: record.provider_mode, features: [...WORKBENCH_FEATURES] };
}
