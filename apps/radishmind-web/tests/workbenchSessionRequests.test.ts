import assert from "node:assert/strict";
import test from "node:test";
import { createPromptApplicationSession } from "../src/features/control-plane-read/promptApplicationSessionConsumer.ts";
import { createApplicationInteractionSession } from "../src/features/control-plane-read/applicationInteractionSessionConsumer.ts";
import { listApplicationResultArtifacts, exportApplicationResultArtifact, transitionApplicationResultArtifactLifecycle } from "../src/features/control-plane-read/applicationResultArtifactConsumer.ts";
import { createWorkflowDefinitionCandidate } from "../src/features/control-plane-read/workflowDefinitionPromotionConsumer.ts";
import { readWorkflowRunDevRecordByID } from "../src/features/control-plane-read/workflowExecutorConsumer.ts";

const identity = { authMode: "local_session_dev_test" as const, baseUrl: "http://platform.test", tenantRef: "tenant_local", workspaceId: "workspace_local", subjectRef: "user:usr_current" };
const applicationId = "app_abcdefghijklmnop";
const sessionId = "appsess_abcdefghijklmnop";
const artifactId = "appres_abcdefghijklmnop";

test("workbench Session, artifact and Workflow requests use cookies and resource scope without developer identity claims", async () => {
  const original = globalThis.fetch;
  const requests: RequestInit[] = [];
  globalThis.fetch = async (_url, init) => {
    requests.push(init ?? {});
    return new Response("{}", { status: 401, headers: { "Content-Type": "application/json" } });
  };
  try {
    const sessionConfig = { ...identity, mode: "dev_application_session_http" as const };
    await createPromptApplicationSession(sessionConfig, applicationId);
    await createApplicationInteractionSession(sessionConfig, { applicationId, executionProfile: "workflow_definition_executor_v1", definitionId: "definition_local" });
    await listApplicationResultArtifacts(sessionConfig, { applicationId, sessionId });
    await exportApplicationResultArtifact(sessionConfig, { applicationId, artifactId });
    await transitionApplicationResultArtifactLifecycle(sessionConfig, { applicationId, sessionId, artifactId, expectedLifecycleVersion: 1, targetState: "archived" });
    await assert.rejects(createWorkflowDefinitionCandidate({ ...identity, mode: "dev_workflow_definition_promotion_http" }, applicationId, { candidateId: "candidate_local", definitionId: "definition_local", draftId: "draft_local", expectedDraftVersion: 1, expectedLifecycleVersion: 1 }));
    await assert.rejects(readWorkflowRunDevRecordByID("run_abcdefghijklmnop", applicationId, { ...identity, mode: "dev_workflow_executor_http" }));
    assert.equal(requests.length, 7);
    for (const request of requests) {
      const headers = new Headers(request.headers);
      assert.equal(request.credentials, "include");
      assert.equal(request.cache, "no-store");
      assert.equal(headers.get("X-RadishMind-Active-Tenant"), identity.tenantRef);
      assert.equal(headers.get("X-RadishMind-Dev-Workflow-Workspace"), identity.workspaceId);
      assert.equal(headers.get("X-RadishMind-Dev-Workflow-Application"), applicationId);
      assert.equal([...headers.keys()].some(key => key.startsWith("x-radishmind-dev-read-")), false);
      assert.equal(headers.has("Authorization"), false);
    }
    for (const index of [0, 1, 2, 3, 4, 5]) assert.equal(new Headers(requests[index]!.headers).get("X-RadishMind-Active-Workspace"), identity.workspaceId);
  } finally {
    globalThis.fetch = original;
  }
});
