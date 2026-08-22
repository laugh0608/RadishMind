import assert from "node:assert/strict";
import test from "node:test";

import {
  activateAdminProviderRouteCandidate,
  buildAdminProviderRouteCandidateDiff,
  createAdminProviderRouteCandidate,
  createAdminProviderRouteDraftInput,
  listAdminProviderRouteActivations,
  readAdminProviderRouteCandidate,
  readAdminProviderRouteDraft,
  readAdminProviderRouteSnapshot,
  reviewAdminProviderRouteCandidate,
  saveAdminProviderRouteDraft,
  validateAdminProviderRouteDraft,
  type AdminProviderRouteConfig,
} from "../src/features/control-plane-read/adminProviderRouteConsumer.ts";

const config: AdminProviderRouteConfig = {
  mode: "dev_admin_provider_route_http",
  baseUrl: "http://platform.test",
  tenantRef: "tenant_demo",
  workspaceId: "workspace_demo",
  environment: "test",
  configurationId: "gateway-default",
  applicationId: "app_flow_copilot",
  subjectRef: "subject_demo_user",
};

test("Admin Provider route consumer stays offline without a request", async () => {
  let fetchCount = 0;
  const originalFetch = globalThis.fetch;
  globalThis.fetch = async () => {
    fetchCount += 1;
    throw new Error("unexpected fetch");
  };
  try {
    const offline = { ...config, mode: "offline" as const };
    assert.equal((await readAdminProviderRouteDraft(offline)).failureCode, "admin_provider_route_http_disabled");
    assert.equal((await readAdminProviderRouteSnapshot(offline)).snapshot, null);
    assert.equal((await listAdminProviderRouteActivations(offline)).activationHistory.length, 0);
    assert.equal(fetchCount, 0);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Admin Provider route draft sends exact scope, CAS, and sanitized configuration", async () => {
  const originalFetch = globalThis.fetch;
  let captured: { url: string; method: string; headers: Headers; body: any } | null = null;
  globalThis.fetch = async (input, init) => {
    captured = {
      url: String(input),
      method: String(init?.method),
      headers: new Headers(init?.headers),
      body: JSON.parse(String(init?.body)),
    };
    return jsonResponse(draftEnvelope());
  };
  try {
    const input = validDraftInput();
    const result = await saveAdminProviderRouteDraft(config, input);
    assert.equal(result.draft?.draftRevision, 1);
    assert.equal(result.draft?.draftDigest, digest("a"));
    assert.equal(captured?.url, "http://platform.test/v1/admin/provider-route-configurations/gateway-default");
    assert.equal(captured?.method, "PUT");
    assert.equal(captured?.headers.get("X-RadishMind-Dev-Read-Tenant"), "tenant_demo");
    assert.equal(captured?.headers.get("X-RadishMind-Dev-Read-Scopes"), "admin_provider_routes:draft,admin_provider_routes:read");
    assert.equal(captured?.headers.get("X-RadishMind-Dev-Admin-Provider-Route-Workspace"), "workspace_demo");
    assert.equal(captured?.headers.get("X-RadishMind-Dev-Admin-Provider-Route-Environment"), "test");
    assert.deepEqual(captured?.body, {
      expected_revision: 0,
      display_name: "Test Gateway routing",
      provider_profiles: [{
        profile_id: "primary",
        display_name: "Primary runtime profile",
        provider_id: "openai-compatible",
        runtime_profile_ref: "ref:radishmind/test/provider-profiles/anyrouter",
        capabilities: ["chat_completions"],
      }],
      model_routes: [{
        route_id: "route-chat-primary",
        protocol: "chat_completions",
        model_id: "gemini-2.5-pro",
        provider_profile_id: "primary",
      }],
    });
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Admin Provider route v2 sends and reads one ordered distinct attempt plan", async () => {
  const originalFetch = globalThis.fetch;
  let body: any = null;
  globalThis.fetch = async (_input, init) => {
    body = init?.body ? JSON.parse(String(init.body)) : null;
    return jsonResponse(v2DraftEnvelope());
  };
  try {
    const result = await saveAdminProviderRouteDraft(config, validV2DraftInput());
    assert.equal(result.draft?.schemaVersion, "admin_provider_route_configuration_draft.v2");
    assert.deepEqual(body.model_routes, [{
      route_id: "route-chat-primary",
      protocol: "chat_completions",
      model_id: "gemini-2.5-pro",
      execution_mode: "sequential_fallback",
      attempt_targets: [
        { ordinal: 1, provider_profile_id: "primary" },
        { ordinal: 2, provider_profile_id: "backup" },
      ],
    }]);
    const route = result.draft?.modelRoutes[0];
    assert.equal(route?.contractVersion, "v2");
    if (route?.contractVersion !== "v2") assert.fail("expected Route v2");
    assert.equal(route.executionMode, "sequential_fallback");
    assert.deepEqual(route.attemptTargets.map((target) => target.providerProfileId), ["primary", "backup"]);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Admin Provider route lifecycle preserves candidate, review, activation, snapshot, and history", async () => {
  const originalFetch = globalThis.fetch;
  const calls: Array<{ path: string; method: string; body: any; scopes: string }> = [];
  globalThis.fetch = async (input, init) => {
    const url = new URL(String(input));
    const body = init?.body ? JSON.parse(String(init.body)) : null;
    calls.push({
      path: url.pathname,
      method: String(init?.method),
      body,
      scopes: new Headers(init?.headers).get("X-RadishMind-Dev-Read-Scopes") ?? "",
    });
    if (url.pathname.endsWith("/candidates")) return jsonResponse(candidateEnvelope());
    if (url.pathname.endsWith("/reviews")) return jsonResponse(approvedCandidateEnvelope());
    if (url.pathname.endsWith("/activations")) return jsonResponse(activationEnvelope());
    if (url.pathname.endsWith("/active-snapshot")) return jsonResponse(snapshotEnvelope());
    if (url.pathname.endsWith("/activation-history")) return jsonResponse(historyEnvelope());
    return jsonResponse(candidateEnvelope());
  };
  try {
    assert.equal((await createAdminProviderRouteCandidate(config, "candidate-one", 1)).candidate?.candidateState, "pending_review");
    assert.equal((await readAdminProviderRouteCandidate(config, "candidate-one")).candidate?.candidateId, "candidate-one");
    assert.equal(
      (await reviewAdminProviderRouteCandidate(config, "candidate-one", 0, "approve", "Inventory and route reviewed.")).candidate?.reviewVersion,
      1,
    );
    assert.equal(
      (await activateAdminProviderRouteCandidate(config, "candidate-one", 0, "activate", "Enable reviewed route.")).snapshot?.generation,
      1,
    );
    assert.equal((await readAdminProviderRouteSnapshot(config)).snapshot?.snapshotDigest, digest("d"));
    assert.equal((await listAdminProviderRouteActivations(config)).activationHistory[0]?.afterCandidateId, "candidate-one");
    assert.deepEqual(calls[0]?.body, { candidate_id: "candidate-one", expected_draft_revision: 1 });
    assert.equal(calls[2]?.scopes, "admin_provider_routes:review,admin_provider_routes:read");
    assert.equal(calls[3]?.scopes, "admin_provider_routes:activate,admin_provider_routes:read");
    assert.deepEqual(calls[3]?.body, {
      expected_generation: 0,
      action: "activate",
      reason: "Enable reviewed route.",
    });
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Admin Provider route consumer keeps CAS conflict metadata from non-2xx envelopes", async () => {
  const originalFetch = globalThis.fetch;
  globalThis.fetch = async () => jsonResponse({
    ...candidateEnvelope(),
    candidate: null,
    failure_code: "admin_provider_route_review_version_conflict",
    current_review_version: 3,
    current_candidate_state: "approved",
  }, 409);
  try {
    const result = await reviewAdminProviderRouteCandidate(
      config,
      "candidate-one",
      1,
      "reject",
      "Stale review cannot replace approval.",
    );
    assert.equal(result.failureCode, "admin_provider_route_review_version_conflict");
    assert.equal(result.currentReviewVersion, 3);
    assert.equal(result.currentCandidateState, "approved");
    assert.equal(result.candidate, null);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Admin Provider route validation rejects scope drift, duplicates, and sensitive material", () => {
  const valid = validDraftInput();
  assert.deepEqual(validateAdminProviderRouteDraft(config, valid), []);

  const invalid = structuredClone(valid);
  invalid.providerProfiles[0]!.runtimeProfileRef = "ref:radishmind/development/provider-profiles/anyrouter";
  invalid.providerProfiles.push({ ...invalid.providerProfiles[0]!, capabilities: [...invalid.providerProfiles[0]!.capabilities] });
  invalid.modelRoutes.push({ ...invalid.modelRoutes[0]!, routeId: "route-duplicate" });
  invalid.displayName = "Authorization: Bearer hidden";
  const fields = validateAdminProviderRouteDraft(config, invalid).map((finding) => finding.field);
  assert.equal(fields.includes("draft"), true);
  assert.equal(fields.includes("provider_profiles[0]"), true);
  assert.equal(fields.includes("provider_profiles[1]"), true);
  assert.equal(fields.includes("model_routes[1]"), true);

  const duplicateTargets = validV2DraftInput();
  if (duplicateTargets.modelRoutes[0]?.contractVersion !== "v2") assert.fail("expected Route v2");
  duplicateTargets.modelRoutes[0].attemptTargets[1]!.providerProfileId = "primary";
  assert.equal(
    validateAdminProviderRouteDraft(config, duplicateTargets).some((finding) => finding.field === "model_routes[0]"),
    true,
  );

  const mixed = validV2DraftInput();
  mixed.modelRoutes.push({ ...valid.modelRoutes[0]!, routeId: "route-chat-legacy" });
  assert.equal(
    validateAdminProviderRouteDraft(config, mixed).some((finding) => finding.field === "model_routes"),
    true,
  );
});

test("Admin Provider route candidate diff reports changed model and baseline generation", async () => {
  const originalFetch = globalThis.fetch;
  globalThis.fetch = async (input) => String(input).endsWith("/active-snapshot")
    ? jsonResponse(previousSnapshotEnvelope())
    : jsonResponse(approvedCandidateEnvelope());
  try {
    const candidate = (await readAdminProviderRouteCandidate(config, "candidate-one")).candidate!;
    const snapshot = (await readAdminProviderRouteSnapshot(config)).snapshot!;
    const diff = buildAdminProviderRouteCandidateDiff(candidate, snapshot);
    assert.equal(diff.baselineGeneration, 1);
    assert.equal(diff.changed, true);
    assert.equal(diff.items.some((item) => item.kind === "display_name" && item.change === "changed"), true);
    assert.equal(diff.items.some((item) => item.kind === "model_route" && item.resourceId === "route-chat-primary"), true);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("Admin Provider route consumer rejects forbidden or scope-drifted responses", async () => {
  const originalFetch = globalThis.fetch;
  try {
    globalThis.fetch = async () => jsonResponse({ ...draftEnvelope(), debug: { endpoint: "https://forbidden.test" } });
    await assert.rejects(() => readAdminProviderRouteDraft(config), /forbidden field/);

    globalThis.fetch = async () => jsonResponse({ ...draftEnvelope(), workspace_id: "workspace-other" });
    await assert.rejects(() => readAdminProviderRouteDraft(config), /invalid HTTP 200 envelope/);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

function validDraftInput() {
  return createAdminProviderRouteDraftInput(config, "openai-compatible", "anyrouter", "gemini-2.5-pro");
}

function validV2DraftInput() {
  const input = validDraftInput();
  input.providerProfiles.push({
    profileId: "backup",
    displayName: "Backup runtime profile",
    providerId: "openai-compatible",
    runtimeProfileRef: "ref:radishmind/test/provider-profiles/backup-router",
    capabilities: ["chat_completions"],
  });
  input.modelRoutes = [{
    contractVersion: "v2",
    routeId: "route-chat-primary",
    protocol: "chat_completions",
    modelId: "gemini-2.5-pro",
    executionMode: "sequential_fallback",
    attemptTargets: [
      { ordinal: 1, providerProfileId: "primary" },
      { ordinal: 2, providerProfileId: "backup" },
    ],
  }];
  return input;
}

function draftEnvelope() {
  return {
    ...baseEnvelope(),
    draft: draftDocument(),
    current_draft_revision: 1,
  };
}

function v2DraftEnvelope() {
  return {
    ...baseEnvelope(),
    draft: {
      ...draftDocument(),
      schema_version: "admin_provider_route_configuration_draft.v2",
      ...v2ConfigurationDocument(),
    },
    current_draft_revision: 1,
  };
}

function candidateEnvelope() {
  return {
    ...baseEnvelope(),
    candidate_id: "candidate-one",
    candidate: candidateDocument(false),
    current_review_version: 0,
    current_candidate_state: "pending_review",
  };
}

function approvedCandidateEnvelope() {
  return {
    ...baseEnvelope(),
    candidate_id: "candidate-one",
    candidate: candidateDocument(true),
    current_review_version: 1,
    current_candidate_state: "approved",
  };
}

function activationEnvelope() {
  return {
    ...baseEnvelope(),
    candidate_id: "candidate-one",
    snapshot: snapshotDocument(),
    activation: activationDocument(),
    current_generation: 1,
  };
}

function snapshotEnvelope() {
  return {
    ...baseEnvelope(),
    snapshot: snapshotDocument(),
    current_generation: 1,
  };
}

function previousSnapshotEnvelope() {
  const snapshot = snapshotDocument();
  snapshot.configuration.display_name = "Previous Gateway routing";
  snapshot.configuration.model_routes[0].model_id = "previous-model";
  return { ...baseEnvelope(), snapshot, current_generation: 1 };
}

function historyEnvelope() {
  return {
    ...baseEnvelope(),
    activation_history: [activationDocument()],
    current_generation: 1,
  };
}

function baseEnvelope() {
  return {
    request_id: "admin-provider-route-web-request",
    workspace_id: "workspace_demo",
    environment: "test",
    configuration_id: "gateway-default",
    activation_history: [],
    failure_code: null,
    audit_ref: "audit_admin_provider_route_web_request",
  };
}

function configurationDocument() {
  return {
    display_name: "Test Gateway routing",
    provider_profiles: [{
      profile_id: "primary",
      display_name: "Primary runtime profile",
      provider_id: "openai-compatible",
      runtime_profile_ref: "ref:radishmind/test/provider-profiles/anyrouter",
      capabilities: ["chat_completions"],
    }],
    model_routes: [{
      route_id: "route-chat-primary",
      protocol: "chat_completions",
      model_id: "gemini-2.5-pro",
      provider_profile_id: "primary",
    }],
  };
}

function v2ConfigurationDocument() {
  const configuration = configurationDocument();
  return {
    ...configuration,
    provider_profiles: [
      ...configuration.provider_profiles,
      {
        profile_id: "backup",
        display_name: "Backup runtime profile",
        provider_id: "openai-compatible",
        runtime_profile_ref: "ref:radishmind/test/provider-profiles/backup-router",
        capabilities: ["chat_completions"],
      },
    ],
    model_routes: [{
      route_id: "route-chat-primary",
      protocol: "chat_completions",
      model_id: "gemini-2.5-pro",
      execution_mode: "sequential_fallback",
      attempt_targets: [
        { ordinal: 1, provider_profile_id: "primary" },
        { ordinal: 2, provider_profile_id: "backup" },
      ],
    }],
  };
}

function draftDocument() {
  return {
    schema_version: "admin_provider_route_configuration_draft.v1",
    tenant_ref: "tenant_demo",
    workspace_id: "workspace_demo",
    environment: "test",
    configuration_id: "gateway-default",
    draft_revision: 1,
    ...configurationDocument(),
    draft_digest: digest("a"),
    created_at: "2026-07-26T01:00:00Z",
    updated_at: "2026-07-26T01:00:00Z",
    created_by_actor_ref: "subject_demo_user",
    updated_by_actor_ref: "subject_demo_user",
    request_id: "draft-request",
    audit_ref: "audit_draft_request",
  };
}

function candidateDocument(approved: boolean) {
  return {
    schema_version: "admin_provider_route_candidate.v1",
    tenant_ref: "tenant_demo",
    workspace_id: "workspace_demo",
    environment: "test",
    configuration_id: "gateway-default",
    candidate_id: "candidate-one",
    source_draft_revision: 1,
    source_draft_digest: digest("a"),
    configuration: configurationDocument(),
    inventory_bindings: [inventoryBinding()],
    candidate_digest: digest("b"),
    candidate_state: approved ? "approved" : "pending_review",
    review_version: approved ? 1 : 0,
    ...(approved ? {
      review: {
        schema_version: "admin_provider_route_review.v1",
        review_version: 1,
        decision: "approve",
        reason: "Inventory and route reviewed.",
        resulting_state: "approved",
        reviewed_at: "2026-07-26T01:05:00Z",
        reviewer_ref: "subject_demo_user",
        request_id: "review-request",
        audit_ref: "audit_review_request",
      },
    } : {}),
    created_at: "2026-07-26T01:02:00Z",
    created_by_actor_ref: "subject_demo_user",
    request_id: "candidate-request",
    audit_ref: "audit_candidate_request",
  };
}

function snapshotDocument() {
  return {
    schema_version: "admin_provider_route_snapshot.v1",
    tenant_ref: "tenant_demo",
    workspace_id: "workspace_demo",
    environment: "test",
    configuration_id: "gateway-default",
    generation: 1,
    candidate_id: "candidate-one",
    candidate_digest: digest("b"),
    configuration: configurationDocument(),
    inventory_bindings: [inventoryBinding()],
    snapshot_digest: digest("d"),
    activated_at: "2026-07-26T01:10:00Z",
    activated_by_actor_ref: "subject_demo_user",
    request_id: "activation-request",
    audit_ref: "audit_activation_request",
  };
}

function inventoryBinding() {
  return {
    profile_id: "primary",
    provider_id: "openai-compatible",
    runtime_profile_ref: "ref:radishmind/test/provider-profiles/anyrouter",
    environment: "test",
    capabilities: ["chat_completions"],
    inventory_digest: digest("c"),
    enabled: true,
  };
}

function activationDocument() {
  return {
    schema_version: "admin_provider_route_activation_record.v1",
    activation_id: "provider-route-activation-1",
    configuration_id: "gateway-default",
    action: "activate",
    reason: "Enable reviewed route.",
    before_generation: 0,
    after_generation: 1,
    after_candidate_id: "candidate-one",
    after_snapshot_digest: digest("d"),
    record_digest: digest("e"),
    created_at: "2026-07-26T01:10:00Z",
    actor_ref: "subject_demo_user",
    request_id: "activation-request",
    audit_ref: "audit_activation_request",
  };
}

function digest(character: string): string {
  return `sha256:${character.repeat(64)}`;
}

function jsonResponse(document: unknown, status = 200): Response {
  return new Response(JSON.stringify(document), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}
