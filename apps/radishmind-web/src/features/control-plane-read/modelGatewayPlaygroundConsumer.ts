export type ModelGatewayPlaygroundProtocol = "chat_completions" | "responses" | "messages";

export type ModelGatewayPlaygroundConfig = {
  mode: "offline" | "dev_gateway_playground_http";
  baseUrl: string;
  tenantRef: string;
  workspaceId: string;
  consumerRef: string;
  applicationId: string;
  subjectRef: string;
  defaultModel: string;
};

export type ModelGatewayPlaygroundInput = {
  protocol: ModelGatewayPlaygroundProtocol;
  model: string;
  inputText: string;
  stream: boolean;
  requestId: string;
};

export type ModelGatewayPlaygroundResult = {
  status: "offline" | "idle" | "submitting" | "succeeded" | "failed" | "canceled";
  requestId: string;
  route: string;
  protocol: ModelGatewayPlaygroundProtocol;
  stream: boolean;
  outputText: string;
  httpStatus: number;
  failureCode: string;
  failureBoundary: string;
  summary: string;
  historyReviewAvailable: boolean;
};

type GatewayErrorDocument = {
  error: {
    message: string;
    code: string;
    failure_boundary: string;
  };
};

const DEV_SOURCE = "dev-gateway-playground-http";
const DEFAULT_BASE_URL = "http://127.0.0.1:7000";
const MAX_INPUT_CHARS = 8_000;
const MAX_OUTPUT_CHARS = 20_000;
const MAX_MODEL_CHARS = 160;

class PlaygroundOutputTooLargeError extends Error {}
class PlaygroundResponseInvalidError extends Error {}

export function readModelGatewayPlaygroundConfig(): ModelGatewayPlaygroundConfig {
  const env = import.meta.env as Record<string, string | undefined>;
  return {
    mode: env.VITE_RADISHMIND_GATEWAY_PLAYGROUND_SOURCE?.trim() === DEV_SOURCE
      ? "dev_gateway_playground_http"
      : "offline",
    baseUrl: normalizeBaseUrl(
      env.VITE_RADISHMIND_GATEWAY_PLAYGROUND_BASE_URL ??
        env.VITE_RADISHMIND_GATEWAY_REQUEST_HISTORY_BASE_URL ??
        env.VITE_RADISHMIND_CONTROL_PLANE_READ_BASE_URL ??
        DEFAULT_BASE_URL,
    ),
    tenantRef: env.VITE_RADISHMIND_GATEWAY_REQUEST_HISTORY_TENANT_REF?.trim() || "tenant_demo",
    workspaceId: env.VITE_RADISHMIND_GATEWAY_REQUEST_HISTORY_WORKSPACE_ID?.trim() || "workspace_demo",
    consumerRef: env.VITE_RADISHMIND_GATEWAY_REQUEST_HISTORY_CONSUMER_REF?.trim() || "consumer_web_dev",
    applicationId: env.VITE_RADISHMIND_GATEWAY_REQUEST_HISTORY_APPLICATION_ID?.trim() || "",
    subjectRef: env.VITE_RADISHMIND_GATEWAY_REQUEST_HISTORY_SUBJECT_REF?.trim() || "subject_web_dev",
    defaultModel: env.VITE_RADISHMIND_GATEWAY_PLAYGROUND_MODEL?.trim() || "radishmind-local-dev",
  };
}

export function initialModelGatewayPlaygroundResult(config: ModelGatewayPlaygroundConfig): ModelGatewayPlaygroundResult {
  return {
    status: config.mode === "dev_gateway_playground_http" ? "idle" : "offline",
    requestId: "",
    route: "",
    protocol: "chat_completions",
    stream: false,
    outputText: "",
    httpStatus: 0,
    failureCode: config.mode === "dev_gateway_playground_http" ? "" : "gateway_playground_disabled",
    failureBoundary: "",
    summary: config.mode === "dev_gateway_playground_http"
      ? "Ready for an explicit dev/test Gateway request."
      : "Offline mode does not send northbound requests.",
    historyReviewAvailable: false,
  };
}

export function modelGatewayPlaygroundConfigForApplication(
  config: ModelGatewayPlaygroundConfig,
  applicationId: string,
): ModelGatewayPlaygroundConfig {
  return { ...config, applicationId: applicationId.trim() };
}

export function createGatewayPlaygroundRequestId(): string {
  const randomPart = globalThis.crypto?.randomUUID?.().replaceAll("-", "") ?? Math.random().toString(16).slice(2);
  return `playground-${Date.now()}-${randomPart.slice(0, 16)}`;
}

export async function submitModelGatewayPlaygroundRequest(
  config: ModelGatewayPlaygroundConfig,
  input: ModelGatewayPlaygroundInput,
  signal?: AbortSignal,
  onStreamOutput?: (outputText: string) => void,
): Promise<ModelGatewayPlaygroundResult> {
  if (config.mode !== "dev_gateway_playground_http") return initialModelGatewayPlaygroundResult(config);
  const validationFailure = validatePlaygroundInput(input);
  const route = protocolRoute(input.protocol);
  if (validationFailure) {
    return failureResult(input, route, 0, "gateway_playground_input_invalid", "client_input", validationFailure, false);
  }

  try {
    const response = await fetch(`${config.baseUrl}${route}`, {
      method: "POST",
      headers: playgroundHeaders(config, input.requestId),
      body: JSON.stringify(buildRequestDocument(input)),
      signal,
    });
    const correlatedRequestId = response.headers.get("X-Request-Id")?.trim();
    if (correlatedRequestId && correlatedRequestId !== input.requestId) {
      return failureResult(input, route, response.status, "gateway_playground_response_invalid", "northbound_response", "Gateway response correlation failed.", true);
    }
    if (!response.ok) return await mapGatewayFailure(response, input, route);

    const outputText = input.stream
      ? await readGatewayPlaygroundStream(response, input.protocol, onStreamOutput)
      : extractUnaryOutput(input.protocol, await response.json());
    if (!outputText.trim()) {
      return failureResult(input, route, response.status, "gateway_playground_response_invalid", "northbound_response", "Gateway response did not contain supported text output.", true);
    }
    return {
      status: "succeeded",
      requestId: input.requestId,
      route,
      protocol: input.protocol,
      stream: input.stream,
      outputText,
      httpStatus: response.status,
      failureCode: "",
      failureBoundary: "",
      summary: input.stream ? "Gateway stream completed." : "Gateway request completed.",
      historyReviewAvailable: true,
    };
  } catch (error) {
    if (isAbortError(error)) {
      return {
        status: "canceled", requestId: input.requestId, route, protocol: input.protocol, stream: input.stream,
        outputText: "", httpStatus: 0, failureCode: "gateway_playground_request_canceled", failureBoundary: "client",
        summary: "Gateway request was canceled by the user.", historyReviewAvailable: true,
      };
    }
    if (error instanceof PlaygroundOutputTooLargeError) {
      return failureResult(input, route, 0, "gateway_playground_output_too_large", "client_output", "Gateway output exceeded the Playground display budget.", true);
    }
    if (error instanceof PlaygroundResponseInvalidError || error instanceof SyntaxError) {
      return failureResult(input, route, 0, "gateway_playground_response_invalid", "northbound_response", "Gateway returned an unsupported response document.", true);
    }
    return failureResult(input, route, 0, "gateway_playground_network_error", "network", "Gateway request could not be completed.", true);
  }
}

function validatePlaygroundInput(input: ModelGatewayPlaygroundInput): string {
  if (!(["chat_completions", "responses", "messages"] as string[]).includes(input.protocol)) return "Choose a supported Gateway protocol.";
  const model = input.model.trim();
  if (!model || model.length > MAX_MODEL_CHARS || /[\r\n\0]/u.test(model)) return `Model must contain 1-${MAX_MODEL_CHARS} single-line characters.`;
  const inputText = input.inputText.trim();
  if (!inputText || inputText.length > MAX_INPUT_CHARS || inputText.includes("\0")) return `Input must contain 1-${MAX_INPUT_CHARS} characters.`;
  if (!/^[A-Za-z0-9._:-]{8,160}$/u.test(input.requestId)) return "Request id is invalid.";
  return "";
}

function buildRequestDocument(input: ModelGatewayPlaygroundInput): Record<string, unknown> {
  const model = input.model.trim();
  const text = input.inputText.trim();
  if (input.protocol === "responses") return { model, input: text, stream: input.stream };
  const document: Record<string, unknown> = { model, messages: [{ role: "user", content: text }], stream: input.stream };
  if (input.protocol === "messages") document.max_tokens = 1_024;
  return document;
}

function playgroundHeaders(config: ModelGatewayPlaygroundConfig, requestId: string): HeadersInit {
  const headers: Record<string, string> = {
    Accept: "application/json, text/event-stream",
    "Content-Type": "application/json",
    "X-Request-Id": requestId,
    "X-RadishMind-Dev-Gateway-Tenant": config.tenantRef,
    "X-RadishMind-Dev-Gateway-Workspace": config.workspaceId,
    "X-RadishMind-Dev-Gateway-Consumer": config.consumerRef,
    "X-RadishMind-Dev-Gateway-Subject": config.subjectRef,
    "X-RadishMind-Dev-Gateway-Scopes": "gateway_requests:invoke,gateway_requests:read",
    "X-RadishMind-Dev-Gateway-Audit": `audit_${requestId}_playground`,
  };
  if (config.applicationId) headers["X-RadishMind-Dev-Gateway-Application"] = config.applicationId;
  return headers;
}

async function mapGatewayFailure(
  response: Response,
  input: ModelGatewayPlaygroundInput,
  route: string,
): Promise<ModelGatewayPlaygroundResult> {
  let document: unknown;
  try {
    document = await response.json();
  } catch {
    document = null;
  }
  if (!isGatewayErrorDocument(document)) {
    return failureResult(input, route, response.status, "gateway_playground_response_invalid", "northbound_response", "Gateway returned an invalid failure document.", true);
  }
  return failureResult(
    input,
    route,
    response.status,
    document.error.code,
    document.error.failure_boundary,
    `Gateway request failed with ${document.error.code}.`,
    true,
  );
}

function extractUnaryOutput(protocol: ModelGatewayPlaygroundProtocol, value: unknown): string {
  if (!isRecord(value)) throw new PlaygroundResponseInvalidError();
  let output = "";
  if (protocol === "chat_completions") {
    const choices = value.choices;
    if (Array.isArray(choices) && isRecord(choices[0]) && isRecord(choices[0].message) && typeof choices[0].message.content === "string") {
      output = choices[0].message.content;
    }
  } else if (protocol === "responses" && typeof value.output_text === "string") {
    output = value.output_text;
  } else if (protocol === "messages" && Array.isArray(value.content)) {
    output = value.content.filter(isRecord).map((item) => typeof item.text === "string" ? item.text : "").filter(Boolean).join("\n");
  }
  return enforceOutputBudget(output);
}

async function readGatewayPlaygroundStream(
  response: Response,
  protocol: ModelGatewayPlaygroundProtocol,
  onStreamOutput?: (outputText: string) => void,
): Promise<string> {
  if (!response.headers.get("Content-Type")?.toLowerCase().includes("text/event-stream") || !response.body) {
    throw new PlaygroundResponseInvalidError();
  }
  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";
  let output = "";
  let terminalSeen = false;
  for (;;) {
    const { value, done } = await reader.read();
    buffer += decoder.decode(value, { stream: !done });
    const events = buffer.split(/\r?\n\r?\n/u);
    buffer = events.pop() ?? "";
    for (const event of events) {
      const parsed = parseSSEData(event, protocol);
      terminalSeen ||= parsed.terminal;
      if (parsed.delta) {
        output = enforceOutputBudget(output + parsed.delta);
        onStreamOutput?.(output);
      }
    }
    if (done) break;
  }
  if (buffer.trim()) {
    const parsed = parseSSEData(buffer, protocol);
    terminalSeen ||= parsed.terminal;
    if (parsed.delta) output = enforceOutputBudget(output + parsed.delta);
  }
  if (!terminalSeen) throw new PlaygroundResponseInvalidError();
  return output;
}

function parseSSEData(event: string, protocol: ModelGatewayPlaygroundProtocol): { delta: string; terminal: boolean } {
  const data = event.split(/\r?\n/u).filter((line) => line.startsWith("data:"))
    .map((line) => line.slice(5).trimStart()).join("\n").trim();
  if (!data) return { delta: "", terminal: false };
  if (data === "[DONE]") return { delta: "", terminal: true };
  const document: unknown = JSON.parse(data);
  if (!isRecord(document)) throw new PlaygroundResponseInvalidError();
  if (protocol === "chat_completions") {
    const choices = document.choices;
    const delta = Array.isArray(choices) && isRecord(choices[0]) && isRecord(choices[0].delta) && typeof choices[0].delta.content === "string"
      ? choices[0].delta.content : "";
    return { delta, terminal: false };
  }
  if (protocol === "responses") {
    return { delta: document.type === "response.output_text.delta" && typeof document.delta === "string" ? document.delta : "", terminal: document.type === "response.completed" };
  }
  const delta = document.type === "content_block_delta" && isRecord(document.delta) && typeof document.delta.text === "string" ? document.delta.text : "";
  return { delta, terminal: document.type === "message_stop" };
}

function failureResult(
  input: ModelGatewayPlaygroundInput,
  route: string,
  httpStatus: number,
  failureCode: string,
  failureBoundary: string,
  summary: string,
  historyReviewAvailable: boolean,
): ModelGatewayPlaygroundResult {
  return {
    status: "failed", requestId: input.requestId, route, protocol: input.protocol, stream: input.stream,
    outputText: "", httpStatus, failureCode, failureBoundary, summary, historyReviewAvailable,
  };
}

function protocolRoute(protocol: ModelGatewayPlaygroundProtocol): string {
  if (protocol === "responses") return "/v1/responses";
  if (protocol === "messages") return "/v1/messages";
  return "/v1/chat/completions";
}

function enforceOutputBudget(output: string): string {
  if (output.length > MAX_OUTPUT_CHARS) throw new PlaygroundOutputTooLargeError();
  return output;
}

function isGatewayErrorDocument(value: unknown): value is GatewayErrorDocument {
  return isRecord(value) && isRecord(value.error) && typeof value.error.message === "string" &&
    typeof value.error.code === "string" && value.error.code.length > 0 &&
    typeof value.error.failure_boundary === "string" && value.error.failure_boundary.length > 0;
}

function isAbortError(error: unknown): boolean {
  return isRecord(error) && error.name === "AbortError";
}

function normalizeBaseUrl(value: string): string {
  return value.trim().replace(/\/$/, "");
}

function isRecord(value: unknown): value is Record<string, any> {
  return Boolean(value) && typeof value === "object" && !Array.isArray(value);
}
