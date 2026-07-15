import {
  readModelGatewayPlaygroundConfig,
  type ModelGatewayPlaygroundConfig,
} from "./modelGatewayPlaygroundConsumer.ts";

export type ApplicationApiProtocol = "chat_completions" | "responses" | "messages";
export type ApplicationApiExampleLanguage = "curl" | "python" | "typescript";

export type ApplicationApiIntegrationConfig = {
  mode: "offline" | "dev_application_api_http";
  authMode: "dev_headers" | "api_key_dev_test";
  baseUrl: string;
  tenantRef: string;
  workspaceId: string;
  consumerRef: string;
  subjectRef: string;
  apiKeyToken: string;
};

export type ApplicationModelCatalogItem = {
  id: string;
  ownedBy: string;
  protocols: ApplicationApiProtocol[];
};

export type ApplicationModelCatalogState = {
  status: "offline" | "idle" | "loading" | "ready" | "empty" | "failed";
  applicationId: string;
  models: ApplicationModelCatalogItem[];
  selectedModel: string;
  failureCode: string;
  summary: string;
};

type ModelListDocument = {
  object: "list";
  data: Array<{
    id: string;
    object: "model";
    created: number;
    owned_by: string;
    metadata?: Record<string, unknown>;
  }>;
};

const FORBIDDEN_MODEL_FIELDS = new Set([
  "authorization", "api_key", "secret", "credential", "endpoint", "base_url", "headers", "cookie",
  "raw_error", "stderr", "stack_trace", "provider_raw_envelope",
]);
const SUPPORTED_PROTOCOLS = new Set<ApplicationApiProtocol>(["chat_completions", "responses", "messages"]);
const MAX_MODEL_ID_CHARS = 160;

class ModelCatalogResponseInvalidError extends Error {}

export function readApplicationApiIntegrationConfig(): ApplicationApiIntegrationConfig {
  const gateway = readModelGatewayPlaygroundConfig();
  return applicationApiIntegrationConfigFromGateway(gateway);
}

export function applicationApiIntegrationConfigFromGateway(
  gateway: ModelGatewayPlaygroundConfig,
): ApplicationApiIntegrationConfig {
  return {
    mode: gateway.mode === "dev_gateway_playground_http" ? "dev_application_api_http" : "offline",
    authMode: gateway.authMode,
    baseUrl: gateway.baseUrl,
    tenantRef: gateway.tenantRef,
    workspaceId: gateway.workspaceId,
    consumerRef: gateway.consumerRef,
    subjectRef: gateway.subjectRef,
    apiKeyToken: gateway.apiKeyToken,
  };
}

export function initialApplicationModelCatalogState(
  config: ApplicationApiIntegrationConfig,
  applicationId: string,
): ApplicationModelCatalogState {
  const offline = config.mode === "offline";
  return {
    status: offline ? "offline" : "idle",
    applicationId: applicationId.trim(),
    models: [],
    selectedModel: "",
    failureCode: offline ? "gateway_model_catalog_disabled" : "",
    summary: offline
      ? "Offline mode does not request the Gateway model catalog."
      : "Load the model catalog for the selected application.",
  };
}

export function resetApplicationModelCatalogState(
  config: ApplicationApiIntegrationConfig,
  applicationId: string,
): ApplicationModelCatalogState {
  return initialApplicationModelCatalogState(config, applicationId);
}

export async function loadApplicationModelCatalog(
  config: ApplicationApiIntegrationConfig,
  applicationId: string,
  signal?: AbortSignal,
  requestId = createModelCatalogRequestId(),
): Promise<ApplicationModelCatalogState> {
  const normalizedApplicationId = applicationId.trim();
  if (config.mode !== "dev_application_api_http") {
    return initialApplicationModelCatalogState(config, normalizedApplicationId);
  }
  if (!isSafeScopeValue(normalizedApplicationId)) {
    return failedCatalogState(normalizedApplicationId, "gateway_model_catalog_scope_invalid", "The selected application scope is invalid.");
  }
  if (config.authMode === "api_key_dev_test" && !config.apiKeyToken) {
    return failedCatalogState(normalizedApplicationId, "gateway_api_key_handoff_required", "Issue an API key and hand it to the Playground before loading models.");
  }

  let response: Response;
  try {
    response = await fetch(`${config.baseUrl}/v1/models`, {
      headers: modelCatalogHeaders(config, normalizedApplicationId, requestId),
      signal,
    });
  } catch (error) {
    if (isAbortError(error)) throw error;
    return failedCatalogState(normalizedApplicationId, "gateway_model_catalog_network_error", "The Gateway model catalog could not be loaded.");
  }
  if (!response.ok) {
    return failedCatalogState(normalizedApplicationId, "gateway_model_catalog_http_failed", `The Gateway model catalog returned HTTP ${response.status}.`);
  }
  const correlatedRequestId = response.headers.get("X-Request-Id")?.trim();
  if (correlatedRequestId && correlatedRequestId !== requestId) {
    return failedCatalogState(normalizedApplicationId, "gateway_model_catalog_response_invalid", "The Gateway model catalog response correlation failed.");
  }

  try {
    const document: unknown = await response.json();
    const models = mapModelCatalog(document);
    return {
      status: models.length > 0 ? "ready" : "empty",
      applicationId: normalizedApplicationId,
      models,
      selectedModel: models[0]?.id ?? "",
      failureCode: "",
      summary: models.length > 0
        ? `Loaded ${models.length} models for the selected application.`
        : "The Gateway model catalog is valid but empty.",
    };
  } catch {
    return failedCatalogState(normalizedApplicationId, "gateway_model_catalog_response_invalid", "The Gateway returned an unsupported model catalog document.");
  }
}

export function generateApplicationApiIntegrationExample(input: {
  protocol: ApplicationApiProtocol;
  language: ApplicationApiExampleLanguage;
  model: string;
}): string {
  const model = input.model.trim();
  if (!isSafeModelId(model)) throw new Error("A valid model id is required to generate an integration example.");
  if (!SUPPORTED_PROTOCOLS.has(input.protocol)) throw new Error("A supported protocol is required.");

  const route = protocolRoute(input.protocol);
  const body = requestBodySource(input.protocol, model);
  if (input.language === "curl") {
    return [
      `curl \"\${RADISHMIND_BASE_URL}${route}\" \\`,
      "  -H \"Content-Type: application/json\" \\",
      "  -H \"Authorization: Bearer ${RADISHMIND_API_KEY}\" \\",
      `  -d '${JSON.stringify(body)}'`,
    ].join("\n");
  }
  if (input.language === "python") {
    return [
      "import os",
      "import requests",
      "",
      "base_url = os.environ[\"RADISHMIND_BASE_URL\"].rstrip(\"/\")",
      "api_key = os.environ[\"RADISHMIND_API_KEY\"]",
      "input_text = os.environ[\"RADISHMIND_INPUT\"]",
      `payload = ${pythonRequestBodySource(input.protocol, model)}`,
      `response = requests.post(f\"{base_url}${route}\", headers={\"Authorization\": f\"Bearer {api_key}\"}, json=payload, timeout=60)`,
      "response.raise_for_status()",
      "print(response.json())",
    ].join("\n");
  }
  if (input.language === "typescript") {
    return [
      "const baseUrl = process.env.RADISHMIND_BASE_URL?.replace(/\\/$/, \"\");",
      "const apiKey = process.env.RADISHMIND_API_KEY;",
      "const inputText = process.env.RADISHMIND_INPUT;",
      "if (!baseUrl || !apiKey || !inputText) throw new Error(\"Missing RadishMind environment variables\");",
      "",
      `const response = await fetch(\`\${baseUrl}${route}\`, {`,
      "  method: \"POST\",",
      "  headers: { \"Content-Type\": \"application/json\", Authorization: `Bearer ${apiKey}` },",
      `  body: JSON.stringify(${typescriptRequestBodySource(input.protocol, model)}),`,
      "});",
      "if (!response.ok) throw new Error(`RadishMind request failed: ${response.status}`);",
      "console.log(await response.json());",
    ].join("\n");
  }
  throw new Error("A supported example language is required.");
}

function mapModelCatalog(value: unknown): ApplicationModelCatalogItem[] {
  if (!isModelListDocument(value) || containsForbiddenModelField(value)) throw new ModelCatalogResponseInvalidError();
  const seen = new Set<string>();
  const models: ApplicationModelCatalogItem[] = [];
  for (const model of value.data) {
    if (seen.has(model.id)) throw new ModelCatalogResponseInvalidError();
    seen.add(model.id);
    models.push({
      id: model.id,
      ownedBy: model.owned_by,
      protocols: publicProtocols(model.metadata),
    });
  }
  return models;
}

function publicProtocols(metadata: Record<string, unknown> | undefined): ApplicationApiProtocol[] {
  const value = metadata?.northbound_protocols;
  if (value === undefined) return ["chat_completions", "responses", "messages"];
  if (!Array.isArray(value) || !value.every((item) => typeof item === "string")) throw new ModelCatalogResponseInvalidError();
  const protocols = value.map(normalizeProtocol).filter((item): item is ApplicationApiProtocol => item !== null);
  return [...new Set(protocols)];
}

function normalizeProtocol(value: string): ApplicationApiProtocol | null {
  if (value === "openai-chat-completions" || value === "chat_completions") return "chat_completions";
  if (value === "openai-responses" || value === "responses") return "responses";
  if (value === "anthropic-messages" || value === "messages") return "messages";
  return null;
}

function isModelListDocument(value: unknown): value is ModelListDocument {
  if (!isRecord(value) || value.object !== "list" || !Array.isArray(value.data)) return false;
  return value.data.every((item) => isRecord(item) && isSafeModelId(item.id) && item.object === "model" &&
    typeof item.created === "number" && Number.isInteger(item.created) && item.created >= 0 &&
    typeof item.owned_by === "string" && item.owned_by.length <= 160 && !/[\r\n\0]/u.test(item.owned_by) &&
    (item.metadata === undefined || isRecord(item.metadata)));
}

function containsForbiddenModelField(value: unknown): boolean {
  if (Array.isArray(value)) return value.some(containsForbiddenModelField);
  if (!isRecord(value)) return false;
  return Object.entries(value).some(([key, nested]) => FORBIDDEN_MODEL_FIELDS.has(key.toLowerCase()) || containsForbiddenModelField(nested));
}

function modelCatalogHeaders(
  config: ApplicationApiIntegrationConfig,
  applicationId: string,
  requestId: string,
): HeadersInit {
  const headers: Record<string, string> = {
    Accept: "application/json",
    "X-Request-Id": requestId,
  };
  if (config.authMode === "api_key_dev_test") {
    headers.Authorization = `Bearer ${config.apiKeyToken}`;
    return headers;
  }
  headers["X-RadishMind-Dev-Gateway-Tenant"] = config.tenantRef;
  headers["X-RadishMind-Dev-Gateway-Workspace"] = config.workspaceId;
  headers["X-RadishMind-Dev-Gateway-Consumer"] = config.consumerRef;
  headers["X-RadishMind-Dev-Gateway-Application"] = applicationId;
  headers["X-RadishMind-Dev-Gateway-Subject"] = config.subjectRef;
  headers["X-RadishMind-Dev-Gateway-Scopes"] = "models:read,gateway_requests:invoke,gateway_requests:read";
  headers["X-RadishMind-Dev-Gateway-Audit"] = `audit_${requestId}_model_catalog`;
  return headers;
}

function requestBodySource(protocol: ApplicationApiProtocol, model: string): Record<string, unknown> {
  if (protocol === "responses") return { model, input: "${RADISHMIND_INPUT}", stream: false };
  const body: Record<string, unknown> = { model, messages: [{ role: "user", content: "${RADISHMIND_INPUT}" }], stream: false };
  if (protocol === "messages") body.max_tokens = 1024;
  return body;
}

function pythonRequestBodySource(protocol: ApplicationApiProtocol, model: string): string {
  if (protocol === "responses") return `{\"model\": ${JSON.stringify(model)}, \"input\": input_text, \"stream\": False}`;
  const suffix = protocol === "messages" ? ", \"max_tokens\": 1024" : "";
  return `{\"model\": ${JSON.stringify(model)}, \"messages\": [{\"role\": \"user\", \"content\": input_text}], \"stream\": False${suffix}}`;
}

function typescriptRequestBodySource(protocol: ApplicationApiProtocol, model: string): string {
  if (protocol === "responses") return `{ model: ${JSON.stringify(model)}, input: inputText, stream: false }`;
  const suffix = protocol === "messages" ? ", max_tokens: 1024" : "";
  return `{ model: ${JSON.stringify(model)}, messages: [{ role: \"user\", content: inputText }], stream: false${suffix} }`;
}

function protocolRoute(protocol: ApplicationApiProtocol): string {
  if (protocol === "responses") return "/v1/responses";
  if (protocol === "messages") return "/v1/messages";
  return "/v1/chat/completions";
}

function failedCatalogState(applicationId: string, failureCode: string, summary: string): ApplicationModelCatalogState {
  return { status: "failed", applicationId, models: [], selectedModel: "", failureCode, summary };
}

function createModelCatalogRequestId(): string {
  const randomPart = globalThis.crypto?.randomUUID?.().replaceAll("-", "") ?? Math.random().toString(16).slice(2);
  return `models-${Date.now()}-${randomPart.slice(0, 16)}`;
}

function isSafeScopeValue(value: string): boolean {
  return /^[A-Za-z0-9._:-]{1,160}$/u.test(value);
}

function isSafeModelId(value: unknown): value is string {
  return typeof value === "string" && value.length > 0 && value.length <= MAX_MODEL_ID_CHARS &&
    /^[A-Za-z0-9._:/-]+$/u.test(value);
}

function isAbortError(error: unknown): boolean {
  return isRecord(error) && error.name === "AbortError";
}

function isRecord(value: unknown): value is Record<string, any> {
  return Boolean(value) && typeof value === "object" && !Array.isArray(value);
}
