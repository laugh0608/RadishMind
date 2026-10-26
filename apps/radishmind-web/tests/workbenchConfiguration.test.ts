import assert from "node:assert/strict";
import test from "node:test";
import { WORKBENCH_FEATURES, parseWorkbenchPublicConfig } from "../../../contracts/typescript/workbench-public-config.ts";
import { initializeWorkbenchConfiguration, workbenchConfiguration, workbenchConsumerConfiguration } from "../src/runtime/workbenchConfiguration.ts";

const fixture = { schema_version: 1, public_origin: "http://127.0.0.1:14100", tenant_ref: "tenant_local", workspace_id: "workspace_local",
  environment: "test", provider_mode: "mock", features: [...WORKBENCH_FEATURES] };

test("workbench configuration rejects secret fields, wrong origins, incomplete scope and feature drift", () => {
  for (const invalid of [null, [], {...fixture, api_key: "secret"}, {...fixture, public_origin: "http://localhost:14100"},
    {...fixture, schema_version: 2}, {...fixture, workspace_id: ""}, {...fixture, features: ["applications"]},
    {...fixture, features: [...WORKBENCH_FEATURES.slice(1), "prompt"]}, {...fixture, provider_mode: "implicit-profile"}]) {
    assert.throws(() => parseWorkbenchPublicConfig(invalid, fixture.public_origin));
  }
  assert.throws(() => parseWorkbenchPublicConfig({...fixture, public_origin:"https://remote.example"}, "https://remote.example"));
});

test("workbench consumers share Session, origin and scope without enabling unrelated sources", () => {
  const config = parseWorkbenchPublicConfig(fixture, fixture.public_origin);
  const environment = workbenchConsumerConfiguration(config);
  assert.equal(environment.VITE_RADISHMIND_READ_AUTH_MODE, "local_session_dev_test");
  assert.equal(environment.VITE_RADISHMIND_CONTROL_PLANE_READ_BASE_URL, fixture.public_origin);
  assert.equal(environment.VITE_RADISHMIND_PROMPT_APPLICATION_WORKSPACE_ID, fixture.workspace_id);
  assert.equal(environment.VITE_RADISHMIND_WORKFLOW_EXECUTOR_WORKSPACE_ID, fixture.workspace_id);
  assert.equal(environment.VITE_RADISHMIND_AGENT_COPILOT_SOURCE, undefined);
  assert.equal(environment.VITE_RADISHMIND_APPLICATION_EVALUATION_SOURCE, undefined);
  assert.equal(environment.VITE_RADISHMIND_WORKFLOW_RAG_SOURCE, undefined);
});

test("bootstrap configuration failure does not install a fixture fallback", async () => {
  await assert.rejects(initializeWorkbenchConfiguration("local-workbench", fixture.public_origin, async () => new Response("unavailable", {status:503})));
  assert.equal(workbenchConfiguration(), null);
  await assert.rejects(initializeWorkbenchConfiguration("local-workbench", fixture.public_origin, async () => new Response("<html>wrong route</html>")));
  assert.equal(workbenchConfiguration(), null);
  await initializeWorkbenchConfiguration("local-workbench", fixture.public_origin, async (url, options) => {
    assert.equal(url, `${fixture.public_origin}/workbench-config.json`);
    assert.equal(options?.cache, "no-store");
    return new Response(JSON.stringify(fixture), {headers:{"Content-Type":"application/json"}});
  });
  assert.equal(workbenchConfiguration()?.workspace_id, fixture.workspace_id);
});
