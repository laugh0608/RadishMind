import { randomUUID } from "node:crypto";
import { expect, type APIRequestContext } from "@playwright/test";

// Establish the existing catalog derivation prerequisite through its public dev/test API.
// The template launcher exposes metadata for this profile but disables provider calls.
export async function prepareTemplateRoute(request: APIRequestContext) {
  const base = "http://127.0.0.1:17000/v1/admin/provider-route-configurations/gateway-default";
  const profile = {
    profile_id: "radishmind-default-workflow", display_name: "Template test profile",
    provider_id: "openai-compatible", runtime_profile_ref: "ref:radishmind/test/provider-profiles/radishmind-default-workflow",
    capabilities: ["chat_completions"],
  };
  const route = { route_id: "template-route", protocol: "chat_completions", model_id: "mock-workflow-model", provider_profile_id: profile.profile_id };
  async function call(path: string, method: string, operation: string, data?: object) {
    const id = `e2e_${randomUUID()}`;
    const response = await request.fetch(base + path, {
      method, data,
      headers: {
        "X-Request-Id": id,
        "X-RadishMind-Dev-Read-Identity": "admin-provider-route-web",
        "X-RadishMind-Dev-Read-Tenant": "tenant_demo",
        "X-RadishMind-Dev-Read-Subject": "subject_demo_user",
        "X-RadishMind-Dev-Read-Scopes": `admin_provider_routes:read,admin_provider_routes:${operation}`,
        "X-RadishMind-Dev-Read-Audit": `audit_${id}`,
        "X-RadishMind-Dev-Admin-Provider-Route-Workspace": "workspace_demo",
        "X-RadishMind-Dev-Admin-Provider-Route-Environment": "test",
      },
    });
    const body = await response.json();
    expect(body.workspace_id).toBe("workspace_demo");
    expect(body.environment).toBe("test");
    expect(body.configuration_id).toBe("gateway-default");
    expect(response.ok(), JSON.stringify(body)).toBeTruthy();
    expect(body.failure_code).toBeNull();
    return body;
  }
  let current = await call("/active-snapshot", "GET", "read");
  if (!current.snapshot) {
    await call("", "PUT", "draft", { expected_revision: 0, display_name: "Isolated template derivation", provider_profiles: [profile], model_routes: [route] });
    const candidate = "candidate-template-e2e";
    await call("/candidates", "POST", "draft", { candidate_id: candidate, expected_draft_revision: 1 });
    await call(`/candidates/${candidate}/reviews`, "POST", "review", { expected_review_version: 0, decision: "approve", reason: "Approve fixed non-callable template test metadata." });
    current = await call(`/candidates/${candidate}/activations`, "POST", "activate", { expected_generation: 0, action: "activate", reason: "Activate exact metadata for isolated template derivation." });
  }
  expect(current.failure_code).toBeNull();
  expect(current.snapshot.configuration.provider_profiles).toEqual([profile]);
  expect(current.snapshot.configuration.model_routes).toEqual([route]);
  expect(current.snapshot.generation).toBe(1);
}
