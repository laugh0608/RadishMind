import assert from "node:assert/strict";
import test from "node:test";

import {
  initialModelGatewayPlaygroundResult,
  modelGatewayPlaygroundConfigForApplication,
  submitModelGatewayPlaygroundRequest,
  type ModelGatewayPlaygroundConfig,
  type ModelGatewayPlaygroundInput,
  type ModelGatewayPlaygroundProtocol,
} from "../src/features/control-plane-read/modelGatewayPlaygroundConsumer.ts";

const offline: ModelGatewayPlaygroundConfig = {
  mode: "offline",
  baseUrl: "http://127.0.0.1:7000",
  tenantRef: "tenant_demo",
  workspaceId: "workspace_demo",
  consumerRef: "consumer_web_dev",
  applicationId: "application_demo",
  subjectRef: "subject_demo",
  defaultModel: "radishmind-local-dev",
};
const live: ModelGatewayPlaygroundConfig = { ...offline, mode: "dev_gateway_playground_http" };

test("Gateway Playground stays offline without fetching", async () => {
  const originalFetch = globalThis.fetch;
  let called = false;
  globalThis.fetch = async () => { called = true; throw new Error("unexpected fetch"); };
  try {
    const result = await submitModelGatewayPlaygroundRequest(offline, input("chat_completions", false));
    assert.equal(initialModelGatewayPlaygroundResult(offline).status, "offline");
    assert.equal(result.failureCode, "gateway_playground_disabled");
    assert.equal(called, false);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Gateway Playground maps all unary protocols and scoped caller headers", async () => {
  const cases: Array<{ protocol: ModelGatewayPlaygroundProtocol; path: string; response: unknown; expected: string }> = [
    { protocol: "chat_completions", path: "/v1/chat/completions", response: { choices: [{ message: { content: "chat output" } }] }, expected: "chat output" },
    { protocol: "responses", path: "/v1/responses", response: { output_text: "responses output" }, expected: "responses output" },
    { protocol: "messages", path: "/v1/messages", response: { content: [{ type: "text", text: "messages output" }] }, expected: "messages output" },
  ];
  const originalFetch = globalThis.fetch;
  try {
    for (const testCase of cases) {
      globalThis.fetch = async (url, init) => {
        assert.equal(new URL(String(url)).pathname, testCase.path);
        const headers = new Headers(init?.headers);
        assert.equal(headers.get("X-RadishMind-Dev-Gateway-Tenant"), "tenant_demo");
        assert.equal(headers.get("X-RadishMind-Dev-Gateway-Scopes"), "gateway_requests:invoke,gateway_requests:read");
        assert.equal(headers.get("X-RadishMind-Dev-Gateway-Application"), "application_demo");
        const body = JSON.parse(String(init?.body)) as Record<string, unknown>;
        assert.equal(body.model, "radishmind-local-dev");
        assert.equal(body.stream, false);
        return jsonResponse(testCase.response, "playground-test-request");
      };
      const result = await submitModelGatewayPlaygroundRequest(live, input(testCase.protocol, false));
      assert.equal(result.status, "succeeded");
      assert.equal(result.outputText, testCase.expected);
      assert.equal(result.requestId, "playground-test-request");
      assert.equal(JSON.stringify(result).includes("private playground input"), false);
    }
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Gateway Playground replaces fixed application config with the current handoff scope", async () => {
  const originalFetch = globalThis.fetch;
  try {
    globalThis.fetch = async (_url, init) => {
      const headers = new Headers(init?.headers);
      assert.equal(headers.get("X-RadishMind-Dev-Gateway-Application"), "app_docs_assistant");
      return jsonResponse({ output_text: "scoped output" }, "playground-test-request");
    };
    const scoped = modelGatewayPlaygroundConfigForApplication(live, "app_docs_assistant");
    const result = await submitModelGatewayPlaygroundRequest(scoped, input("responses", false));
    assert.equal(result.status, "succeeded");
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Gateway Playground parses terminal SSE output for all protocols", async () => {
  const cases: Array<{ protocol: ModelGatewayPlaygroundProtocol; body: string; expected: string }> = [
    {
      protocol: "chat_completions",
      body: 'data: {"choices":[{"delta":{"content":"chat "}}]}\n\ndata: {"choices":[{"delta":{"content":"stream"}}]}\n\ndata: [DONE]\n\n',
      expected: "chat stream",
    },
    {
      protocol: "responses",
      body: 'event: response.output_text.delta\ndata: {"type":"response.output_text.delta","delta":"responses stream"}\n\nevent: response.completed\ndata: {"type":"response.completed"}\n\n',
      expected: "responses stream",
    },
    {
      protocol: "messages",
      body: 'event: content_block_delta\ndata: {"type":"content_block_delta","delta":{"type":"text_delta","text":"messages stream"}}\n\nevent: message_stop\ndata: {"type":"message_stop"}\n\n',
      expected: "messages stream",
    },
  ];
  const originalFetch = globalThis.fetch;
  try {
    for (const testCase of cases) {
      globalThis.fetch = async () => new Response(testCase.body, {
        status: 200,
        headers: { "Content-Type": "text/event-stream", "X-Request-Id": "playground-test-request" },
      });
      const outputs: string[] = [];
      const result = await submitModelGatewayPlaygroundRequest(live, input(testCase.protocol, true), undefined, (output) => outputs.push(output));
      assert.equal(result.status, "succeeded");
      assert.equal(result.outputText, testCase.expected);
      assert.equal(outputs.at(-1), testCase.expected);
    }
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Gateway Playground preserves stable HTTP failures and rejects correlation mismatch", async () => {
  const originalFetch = globalThis.fetch;
  try {
    globalThis.fetch = async () => jsonResponse({
      error: { message: "sanitized", code: "BRIDGE_WORKER_TIMEOUT", failure_boundary: "python_bridge" },
    }, "playground-test-request", 504);
    const failed = await submitModelGatewayPlaygroundRequest(live, input("responses", false));
    assert.equal(failed.status, "failed");
    assert.equal(failed.httpStatus, 504);
    assert.equal(failed.failureCode, "BRIDGE_WORKER_TIMEOUT");
    assert.equal(failed.failureBoundary, "python_bridge");

    globalThis.fetch = async () => jsonResponse({ output_text: "must be rejected" }, "different-request");
    const mismatch = await submitModelGatewayPlaygroundRequest(live, input("responses", false));
    assert.equal(mismatch.failureCode, "gateway_playground_response_invalid");
    assert.equal(mismatch.outputText, "");

    globalThis.fetch = async () => jsonResponse({ unsupported: true }, "playground-test-request");
    const malformed = await submitModelGatewayPlaygroundRequest(live, input("responses", false));
    assert.equal(malformed.failureCode, "gateway_playground_response_invalid");
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Gateway Playground rejects invalid input and maps user abort without retry", async () => {
  const originalFetch = globalThis.fetch;
  let calls = 0;
  try {
    globalThis.fetch = async (_url, init) => {
      calls++;
      return await new Promise<Response>((_resolve, reject) => {
        init?.signal?.addEventListener("abort", () => reject(new DOMException("aborted", "AbortError")), { once: true });
      });
    };
    const invalid = await submitModelGatewayPlaygroundRequest(live, { ...input("messages", false), inputText: "" });
    assert.equal(invalid.failureCode, "gateway_playground_input_invalid");
    assert.equal(calls, 0);

    const controller = new AbortController();
    const pending = submitModelGatewayPlaygroundRequest(live, input("messages", true), controller.signal);
    controller.abort();
    const canceled = await pending;
    assert.equal(canceled.status, "canceled");
    assert.equal(canceled.httpStatus, 0);
    assert.equal(canceled.failureCode, "gateway_playground_request_canceled");
    assert.equal(canceled.failureBoundary, "client");
    assert.equal(calls, 1);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Gateway Playground separates output budget and network failures", async () => {
  const originalFetch = globalThis.fetch;
  try {
    globalThis.fetch = async () => jsonResponse({ output_text: "x".repeat(20_001) }, "playground-test-request");
    const oversized = await submitModelGatewayPlaygroundRequest(live, input("responses", false));
    assert.equal(oversized.failureCode, "gateway_playground_output_too_large");
    assert.equal(oversized.outputText, "");

    globalThis.fetch = async () => { throw new TypeError("connection failed"); };
    const network = await submitModelGatewayPlaygroundRequest(live, input("responses", false));
    assert.equal(network.failureCode, "gateway_playground_network_error");
    assert.equal(network.summary.includes("connection failed"), false);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

function input(protocol: ModelGatewayPlaygroundProtocol, stream: boolean): ModelGatewayPlaygroundInput {
  return {
    protocol,
    model: "radishmind-local-dev",
    inputText: "private playground input",
    stream,
    requestId: "playground-test-request",
  };
}

function jsonResponse(body: unknown, requestId: string, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json", "X-Request-Id": requestId },
  });
}
