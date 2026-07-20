import assert from "node:assert/strict";
import test from "node:test";

import {
  generateApplicationApiIntegrationExample,
  initialApplicationModelCatalogState,
  loadApplicationModelCatalog,
  resetApplicationModelCatalogState,
  type ApplicationApiExampleLanguage,
  type ApplicationApiIntegrationConfig,
  type ApplicationApiProtocol,
} from "../src/features/control-plane-read/applicationApiIntegrationConsumer.ts";
import {
  APPLICATION_API_INTEGRATION_DRAFT_HANDOFF_EVENT,
  APPLICATION_MODEL_CATALOG_READY_EVENT,
  clearPendingApplicationApiIntegrationDraftHandoff,
  consumePendingApplicationApiIntegrationDraftHandoff,
  createApplicationApiIntegrationDraftHandoffDetail,
  createApplicationModelCatalogReadyDetail,
  requestApplicationApiIntegrationDraftHandoff,
  requestApplicationModelCatalogReady,
} from "../src/features/control-plane-read/applicationApiIntegrationEvents.ts";
import {
  MODEL_GATEWAY_PLAYGROUND_HANDOFF_EVENT,
  MODEL_GATEWAY_REQUEST_REVIEW_EVENT,
  createAPIKeyModelGatewayPlaygroundHandoffDetail,
  createModelGatewayPlaygroundHandoffDetail,
  createModelGatewayRequestReviewDetail,
  requestAPIKeyModelGatewayPlaygroundHandoff,
  requestGatewayRequestHistoryReview,
  requestModelGatewayPlaygroundHandoff,
} from "../src/features/control-plane-read/modelGatewayPlaygroundEvents.ts";

const offline: ApplicationApiIntegrationConfig = {
  mode: "offline",
  authMode: "dev_headers",
  baseUrl: "http://127.0.0.1:7000",
  tenantRef: "tenant_demo",
  workspaceId: "workspace_demo",
  consumerRef: "consumer_web_dev",
  subjectRef: "subject_web_dev",
  apiKeyToken: "",
};
const live: ApplicationApiIntegrationConfig = { ...offline, mode: "dev_application_api_http" };

test("Application API Integration stays offline without fetching", async () => {
  const originalFetch = globalThis.fetch;
  let called = false;
  globalThis.fetch = async () => { called = true; throw new Error("unexpected fetch"); };
  try {
    const state = await loadApplicationModelCatalog(offline, "app_flow_copilot");
    assert.equal(initialApplicationModelCatalogState(offline, "app_flow_copilot").status, "offline");
    assert.equal(state.failureCode, "gateway_model_catalog_disabled");
    assert.equal(called, false);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Application model catalog validates success and carries current scope", async () => {
  const originalFetch = globalThis.fetch;
  try {
    globalThis.fetch = async (url, init) => {
      assert.equal(String(url), "http://127.0.0.1:7000/v1/models");
      const headers = new Headers(init?.headers);
      assert.equal(headers.get("X-RadishMind-Dev-Gateway-Workspace"), "workspace_demo");
      assert.equal(headers.get("X-RadishMind-Dev-Gateway-Application"), "app_docs_assistant");
      assert.equal(headers.get("X-RadishMind-Dev-Gateway-Scopes"), "models:read,gateway_requests:invoke,gateway_requests:read");
      return jsonResponse({
        object: "list",
        data: [{
          id: "profile:local-dev",
          object: "model",
          created: 1,
          owned_by: "radishmind",
          metadata: { northbound_protocols: ["openai-chat-completions", "openai-responses", "anthropic-messages"], credential_state: "configured" },
        }],
      }, "model-catalog-test-request");
    };
    const state = await loadApplicationModelCatalog(live, "app_docs_assistant", undefined, "model-catalog-test-request");
    assert.equal(state.status, "ready");
    assert.equal(state.applicationId, "app_docs_assistant");
    assert.equal(state.selectedModel, "profile:local-dev");
    assert.deepEqual(state.models[0]?.protocols, ["chat_completions", "responses", "messages"]);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Application model catalog uses only Bearer auth after an API key handoff", async () => {
  const originalFetch = globalThis.fetch;
  const apiKeyToken = `rmd_dev_key_aaaaaaaaaaaaaaaa.${"A".repeat(43)}`;
  const apiKeyConfig: ApplicationApiIntegrationConfig = {
    ...live,
    authMode: "api_key_dev_test",
    apiKeyToken,
  };
  try {
    globalThis.fetch = async (_url, init) => {
      const headers = new Headers(init?.headers);
      assert.equal(headers.get("Authorization"), `Bearer ${apiKeyToken}`);
      assert.equal(headers.has("X-RadishMind-Dev-Gateway-Tenant"), false);
      assert.equal(headers.has("X-RadishMind-Dev-Gateway-Application"), false);
      return jsonResponse({ object: "list", data: [] }, "models-api-key-request");
    };
    const loaded = await loadApplicationModelCatalog(apiKeyConfig, "app_aaaaaaaaaaaaaaaa", undefined, "models-api-key-request");
    assert.equal(loaded.status, "empty");
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Application model catalog does not fetch before API key handoff", async () => {
  const originalFetch = globalThis.fetch;
  let called = false;
  globalThis.fetch = async () => { called = true; throw new Error("unexpected fetch"); };
  try {
    const waiting: ApplicationApiIntegrationConfig = { ...live, authMode: "api_key_dev_test", apiKeyToken: "" };
    const state = await loadApplicationModelCatalog(waiting, "app_aaaaaaaaaaaaaaaa");
    assert.equal(state.failureCode, "gateway_api_key_handoff_required");
    assert.equal(called, false);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Application model catalog separates empty, HTTP failure, and invalid responses", async () => {
  const originalFetch = globalThis.fetch;
  try {
    globalThis.fetch = async () => jsonResponse({ object: "list", data: [] }, "empty-catalog-request");
    const empty = await loadApplicationModelCatalog(live, "app_flow_copilot", undefined, "empty-catalog-request");
    assert.equal(empty.status, "empty");
    assert.equal(empty.models.length, 0);

    globalThis.fetch = async () => new Response("upstream detail must stay unread", { status: 503 });
    const httpFailure = await loadApplicationModelCatalog(live, "app_flow_copilot", undefined, "http-catalog-request");
    assert.equal(httpFailure.failureCode, "gateway_model_catalog_http_failed");
    assert.equal(httpFailure.summary.includes("upstream detail"), false);

    const invalidDocuments: unknown[] = [
      { object: "models", data: [] },
      { object: "list", data: [{ id: "bad model", object: "model", created: 1, owned_by: "radishmind" }] },
      { object: "list", data: [{ id: "safe-model", object: "model", created: 1, owned_by: "radishmind", metadata: { api_key: "sk-private" } }] },
      { object: "list", data: [{ id: "duplicate", object: "model", created: 1, owned_by: "radishmind" }, { id: "duplicate", object: "model", created: 2, owned_by: "radishmind" }] },
    ];
    for (const [index, document] of invalidDocuments.entries()) {
      const requestId = `invalid-catalog-${index}`;
      globalThis.fetch = async () => jsonResponse(document, requestId);
      const invalid = await loadApplicationModelCatalog(live, "app_flow_copilot", undefined, requestId);
      assert.equal(invalid.failureCode, "gateway_model_catalog_response_invalid");
      assert.equal(JSON.stringify(invalid).includes("sk-private"), false);
    }
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Application API examples cover all protocol and language pairs without secret or dev headers", () => {
  const protocols: ApplicationApiProtocol[] = ["chat_completions", "responses", "messages"];
  const languages: ApplicationApiExampleLanguage[] = ["curl", "python", "typescript"];
  const routes: Record<ApplicationApiProtocol, string> = {
    chat_completions: "/v1/chat/completions",
    responses: "/v1/responses",
    messages: "/v1/messages",
  };
  for (const protocol of protocols) {
    for (const language of languages) {
      const example = generateApplicationApiIntegrationExample({ protocol, language, model: "profile:local-dev" });
      assert.equal(example.includes(routes[protocol]), true);
      assert.equal(example.includes("profile:local-dev"), true);
      assert.equal(example.includes("RADISHMIND_BASE_URL"), true);
      assert.equal(example.includes("RADISHMIND_API_KEY"), true);
      assert.equal(example.includes("RADISHMIND_INPUT"), true);
      assert.equal(example.includes("sk-private"), false);
      assert.equal(example.includes("X-RadishMind-Dev-"), false);
      assert.equal(example.toLowerCase().includes("key_hash"), false);
    }
  }
});

test("Application and request-id handoffs preserve exact scope", () => {
  assert.deepEqual(
    createApplicationApiIntegrationDraftHandoffDetail(" app_docs_assistant ", "responses", " profile:local-dev "),
    { applicationId: "app_docs_assistant", protocol: "responses", model: "profile:local-dev" },
  );
  assert.deepEqual(
    createApplicationModelCatalogReadyDetail(
      " app_aaaaaaaaaaaaaaaa ",
      [{ id: " profile:local-dev ", ownedBy: "radishmind", protocols: ["responses", "responses"] }],
      " profile:local-dev ",
    ),
    {
      applicationId: "app_aaaaaaaaaaaaaaaa",
      models: [{ id: "profile:local-dev", ownedBy: "radishmind", protocols: ["responses"] }],
      selectedModel: "profile:local-dev",
    },
  );
  assert.throws(
    () => createApplicationModelCatalogReadyDetail(
      "app_aaaaaaaaaaaaaaaa",
      [{ id: "profile:local-dev", ownedBy: "radishmind", protocols: ["responses"] }],
      "missing-model",
    ),
    /selection is invalid/,
  );
  assert.deepEqual(
    createModelGatewayPlaygroundHandoffDetail(" app_docs_assistant ", "responses", " profile:local-dev "),
    { applicationId: "app_docs_assistant", protocol: "responses", model: "profile:local-dev" },
  );
  const apiKeyToken = `rmd_dev_key_aaaaaaaaaaaaaaaa.${"A".repeat(43)}`;
  assert.deepEqual(
    createAPIKeyModelGatewayPlaygroundHandoffDetail(
      "app_aaaaaaaaaaaaaaaa",
      "key_aaaaaaaaaaaaaaaa",
      apiKeyToken,
      "profile:local-dev",
    ),
    {
      applicationId: "app_aaaaaaaaaaaaaaaa",
      protocol: "responses",
      model: "profile:local-dev",
      apiKeyCredential: { apiKeyId: "key_aaaaaaaaaaaaaaaa", token: apiKeyToken },
    },
  );
  assert.deepEqual(
    createModelGatewayRequestReviewDetail(" playground-request-001 ", " app_docs_assistant "),
    { requestId: "playground-request-001", applicationId: "app_docs_assistant" },
  );
  assert.deepEqual(
    createModelGatewayRequestReviewDetail(
      " playground-request-002 ",
      " app_aaaaaaaaaaaaaaaa ",
      " api_key:key_aaaaaaaaaaaaaaaa ",
    ),
    {
      requestId: "playground-request-002",
      applicationId: "app_aaaaaaaaaaaaaaaa",
      consumerRef: "api_key:key_aaaaaaaaaaaaaaaa",
    },
  );
});

test("Application and Gateway handoff events dispatch their validated details", () => {
  const originalWindow = Object.getOwnPropertyDescriptor(globalThis, "window");
  const eventTarget = new EventTarget();
  const received: Array<{ type: string; detail: unknown }> = [];
  for (const type of [
    APPLICATION_API_INTEGRATION_DRAFT_HANDOFF_EVENT,
    APPLICATION_MODEL_CATALOG_READY_EVENT,
    MODEL_GATEWAY_PLAYGROUND_HANDOFF_EVENT,
    MODEL_GATEWAY_REQUEST_REVIEW_EVENT,
  ]) {
    eventTarget.addEventListener(type, (event) => {
      received.push({ type: event.type, detail: (event as CustomEvent<unknown>).detail });
    });
  }
  Object.defineProperty(globalThis, "window", { configurable: true, value: eventTarget });
  const apiKeyToken = `rmd_dev_key_aaaaaaaaaaaaaaaa.${"A".repeat(43)}`;
  try {
    requestApplicationApiIntegrationDraftHandoff(" app_docs_assistant ", "responses", " profile:local-dev ");
    requestApplicationModelCatalogReady(
      " app_docs_assistant ",
      [{ id: " profile:local-dev ", ownedBy: "radishmind", protocols: ["responses"] }],
      " profile:local-dev ",
    );
    requestModelGatewayPlaygroundHandoff(" app_docs_assistant ", "responses", " profile:local-dev ");
    requestAPIKeyModelGatewayPlaygroundHandoff(
      "app_aaaaaaaaaaaaaaaa",
      "key_aaaaaaaaaaaaaaaa",
      apiKeyToken,
      "profile:local-dev",
    );
    requestGatewayRequestHistoryReview(" playground-request-001 ", " app_docs_assistant ");
  } finally {
    clearPendingApplicationApiIntegrationDraftHandoff();
    if (originalWindow) {
      Object.defineProperty(globalThis, "window", originalWindow);
    } else {
      Reflect.deleteProperty(globalThis, "window");
    }
  }

  assert.deepEqual(received.map((event) => event.type), [
    APPLICATION_API_INTEGRATION_DRAFT_HANDOFF_EVENT,
    APPLICATION_MODEL_CATALOG_READY_EVENT,
    MODEL_GATEWAY_PLAYGROUND_HANDOFF_EVENT,
    MODEL_GATEWAY_PLAYGROUND_HANDOFF_EVENT,
    MODEL_GATEWAY_REQUEST_REVIEW_EVENT,
  ]);
  assert.deepEqual(received[0]?.detail, {
    applicationId: "app_docs_assistant",
    protocol: "responses",
    model: "profile:local-dev",
  });
  assert.deepEqual(received[4]?.detail, {
    requestId: "playground-request-001",
    applicationId: "app_docs_assistant",
  });
});

test("Application API handoff survives one route mount and remains exact-scope one-time memory", () => {
  const originalWindow = Object.getOwnPropertyDescriptor(globalThis, "window");
  Object.defineProperty(globalThis, "window", { configurable: true, value: new EventTarget() });
  clearPendingApplicationApiIntegrationDraftHandoff();
  try {
    requestApplicationApiIntegrationDraftHandoff("app_docs_assistant", "responses", "profile:local-dev");
    assert.equal(consumePendingApplicationApiIntegrationDraftHandoff("app_flow_copilot"), null);
    assert.deepEqual(consumePendingApplicationApiIntegrationDraftHandoff("app_docs_assistant"), {
      applicationId: "app_docs_assistant",
      protocol: "responses",
      model: "profile:local-dev",
    });
    assert.equal(consumePendingApplicationApiIntegrationDraftHandoff("app_docs_assistant"), null);
  } finally {
    clearPendingApplicationApiIntegrationDraftHandoff();
    if (originalWindow) {
      Object.defineProperty(globalThis, "window", originalWindow);
    } else {
      Reflect.deleteProperty(globalThis, "window");
    }
  }
});

test("Application and Gateway handoffs reject ambiguous or unsafe scope", () => {
  const validModel = { id: "profile:local-dev", ownedBy: "radishmind", protocols: ["responses" as const] };
  assert.throws(
    () => createApplicationApiIntegrationDraftHandoffDetail("bad scope", "responses", "profile:local-dev"),
    /scope is invalid/,
  );
  assert.throws(
    () => createApplicationApiIntegrationDraftHandoffDetail("app_docs_assistant", "responses", "unsafe model"),
    /selection is invalid/,
  );
  assert.throws(() => createApplicationModelCatalogReadyDetail("bad scope", [validModel], validModel.id), /scope is invalid/);
  assert.throws(
    () => createApplicationModelCatalogReadyDetail("app_docs_assistant", [validModel, validModel], validModel.id),
    /handoff is invalid/,
  );
  assert.throws(
    () => createApplicationModelCatalogReadyDetail(
      "app_docs_assistant",
      [{ ...validModel, ownedBy: "unsafe\nowner" }],
      validModel.id,
    ),
    /handoff is invalid/,
  );
  assert.throws(
    () => createApplicationModelCatalogReadyDetail(
      "app_docs_assistant",
      [{ ...validModel, protocols: ["unsafe" as never] }],
      validModel.id,
    ),
    /handoff is invalid/,
  );
  assert.throws(() => createModelGatewayPlaygroundHandoffDetail("bad scope", "responses", validModel.id), /application/);
  assert.throws(() => createModelGatewayPlaygroundHandoffDetail("app_docs_assistant", "responses", "bad model"), /model/);
  assert.throws(
    () => createModelGatewayPlaygroundHandoffDetail("app_docs_assistant", "unsafe" as never, validModel.id),
    /protocol/,
  );
  assert.throws(
    () => createAPIKeyModelGatewayPlaygroundHandoffDetail(
      "app_aaaaaaaaaaaaaaaa",
      "key_aaaaaaaaaaaaaaab",
      `rmd_dev_key_aaaaaaaaaaaaaaaa.${"A".repeat(43)}`,
      validModel.id,
    ),
    /API key/,
  );
  assert.throws(() => createModelGatewayRequestReviewDetail("short", "app_docs_assistant"), /request/);
  assert.throws(() => createModelGatewayRequestReviewDetail("request-123", "bad scope"), /application/);
  assert.throws(
    () => createModelGatewayRequestReviewDetail("request-123", "app_docs_assistant", "unsafe/consumer"),
    /consumer/,
  );
});

test("Application switch resets catalog, model selection, and failure state", () => {
  const previous = {
    status: "failed" as const,
    applicationId: "app_flow_copilot",
    models: [{ id: "old-model", ownedBy: "radishmind", protocols: ["responses" as const] }],
    selectedModel: "old-model",
    failureCode: "old_failure",
    summary: "old summary",
  };
  const next = resetApplicationModelCatalogState(live, "app_docs_assistant");
  assert.notDeepEqual(next, previous);
  assert.equal(next.applicationId, "app_docs_assistant");
  assert.equal(next.status, "idle");
  assert.equal(next.models.length, 0);
  assert.equal(next.selectedModel, "");
  assert.equal(next.failureCode, "");
  assert.equal(next.summary.includes("old"), false);
});

function jsonResponse(body: unknown, requestId: string): Response {
  return new Response(JSON.stringify(body), {
    status: 200,
    headers: { "Content-Type": "application/json", "X-Request-Id": requestId },
  });
}
