import assert from "node:assert/strict";
import test from "node:test";

import type { WorkflowDraftDesignerDraft } from "../src/features/control-plane-read/workflowDraftDesigner.ts";
import { compareWorkflowSavedDraftRevision } from "../src/features/control-plane-read/workflowSavedDraftRevisionComparison.ts";
import {
  initialWorkflowSavedDraftRevisionHistoryState,
  listWorkflowSavedDraftRevisions,
  restoreWorkflowSavedDraftRevision,
} from "../src/features/control-plane-read/workflowSavedDraftRevisionConsumer.ts";
import type { WorkflowSavedDraftConsumerConfig } from "../src/features/control-plane-read/savedWorkflowDraftConsumer.ts";

const config: WorkflowSavedDraftConsumerConfig = {
  mode: "dev_saved_draft_http",
  baseUrl: "http://127.0.0.1:7000",
  workspaceId: "workspace_demo",
  tenantRef: "tenant_demo",
  subjectRef: "subject_demo_user",
};

test("revision comparison reports metadata, node, edge, layout, review, and provenance changes", () => {
  const historical = workflowDraftRevisionFixture();
  const active = structuredClone(historical);
  active.label = "Changed draft";
  active.nodes[0] = { ...active.nodes[0], providerRef: "profile:changed" };
  active.edges = active.edges.slice(1);
  active.designerLayout.nodePositions = [{ nodeId: "node_input", x: 24, y: 48 }];
  active.blockedCapabilities = [{
    capabilityId: "blocked_writeback",
    label: "Business writeback",
    status: "blocked",
    missingPrerequisite: "confirmation",
    summary: "Writeback remains blocked.",
    auditRef: "audit_blocked_writeback",
  }];
  active.derivation = {
    version: 1,
    sourceKind: "saved_workflow_draft",
    sourceDraftId: "draft_source",
    sourceDraftVersion: 2,
  };

  const comparison = compareWorkflowSavedDraftRevision(historical, active);

  assert.equal(comparison.changed, true);
  assert.equal(comparison.metadataChangeCount, 1);
  assert.equal(comparison.nodeChangeCount, 1);
  assert.equal(comparison.edgeChangeCount, 1);
  assert.equal(comparison.reviewContextChangeCount, 3);
  assert.deepEqual(
    comparison.changes.map((change) => change.kind),
    ["metadata", "node_changed", "edge_removed", "layout", "validation", "provenance"],
  );
});

test("revision history consumer preserves opaque pagination and summary metadata", async (t) => {
  const draft = workflowDraftRevisionFixture();
  const originalFetch = globalThis.fetch;
  t.after(() => { globalThis.fetch = originalFetch; });
  globalThis.fetch = async (input, init) => {
    assert.match(String(input), /revisions\?.*limit=1/);
    assert.equal(new Headers(init?.headers).get("X-RadishMind-Dev-Workflow-Application"), draft.applicationRef);
    return new Response(JSON.stringify({
      request_id: "request-revisions",
      workspace_id: config.workspaceId,
      application_id: draft.applicationRef,
      revisions: [{
        schema_version: "saved_workflow_draft_revision.v1",
        draft_id: draft.draftId,
        draft_version: 3,
        revision_kind: "restored",
        restored_from_version: 1,
        draft_status: "valid_for_review",
        name: draft.label,
        updated_at: "2026-07-27T13:00:00Z",
        updated_by_actor_ref: config.subjectRef,
        node_count: draft.nodes.length,
        edge_count: draft.edges.length,
        blocked_count: 0,
      }],
      next_cursor: "opaque-next",
      has_more: true,
      failure_code: null,
      audit_ref: "audit-revisions",
    }), { status: 200 });
  };

  const result = await listWorkflowSavedDraftRevisions(draft, config, "", 1);

  assert.equal(result.status, "ready");
  assert.equal(result.revisions[0].revisionKind, "restored");
  assert.equal(result.revisions[0].restoredFromVersion, 1);
  assert.equal(result.nextCursor, "opaque-next");
  assert.equal(result.hasMore, true);
});

test("restore request carries read and write membership permissions and exposes version conflicts", async (t) => {
  const draft = workflowDraftRevisionFixture();
  const originalFetch = globalThis.fetch;
  t.after(() => { globalThis.fetch = originalFetch; });
  globalThis.fetch = async (_input, init) => {
    const headers = new Headers(init?.headers);
    assert.equal(
      headers.get("X-RadishMind-Dev-Read-Membership-Permissions"),
      "workflow_drafts:read,workflow_drafts:write",
    );
    assert.deepEqual(JSON.parse(String(init?.body)), {
      expected_current_draft_version: 3,
      expected_lifecycle_version: 2,
    });
    return new Response(JSON.stringify({
      request_id: "request-restore-conflict",
      workspace_id: config.workspaceId,
      application_id: draft.applicationRef,
      revision: null,
      draft: null,
      failure_code: "draft_version_conflict",
      current_draft_version: 4,
      current_lifecycle_version: 2,
      current_lifecycle_state: "active",
      validation_summary: { validation_state: "", valid_for_review: false, findings: [] },
      audit_ref: "audit-restore-conflict",
    }), { status: 200 });
  };

  const result = await restoreWorkflowSavedDraftRevision(draft, 1, 3, 2, config);

  assert.equal(result.draft, null);
  assert.equal(result.failureCode, "draft_version_conflict");
  assert.equal(result.currentDraftVersion, 4);
  assert.equal(result.currentLifecycleVersion, 2);
});

test("sample-only mode exposes a disabled revision history state", () => {
  const state = initialWorkflowSavedDraftRevisionHistoryState({ ...config, mode: "sample_only" });
  assert.equal(state.status, "disabled");
  assert.deepEqual(state.revisions, []);
});

function workflowDraftRevisionFixture(): WorkflowDraftDesignerDraft {
  return {
    draftId: "draft_revision_test",
    templateRef: "draft_revision_test",
    label: "Revision draft",
    applicationRef: "app_flow_copilot",
    workflowDefinitionId: "wf_revision_test",
    baseDefinitionVersion: 1,
    providerProfileRef: "profile:default",
    summary: "Revision comparison fixture.",
    nodes: [
      {
        nodeId: "node_input",
        label: "Input",
        nodeType: "prompt",
        lane: "context",
        readiness: "ready",
        inputSummary: "Input",
        outputSummary: "Context",
        providerRef: "",
        toolRef: "",
        ragRef: "",
        inputContractFields: ["input"],
        outputContractFields: ["context"],
        outputMappingSummary: "Map input.",
        riskLevel: "low",
        requiresConfirmation: false,
        previewOnlyReason: "",
      },
      {
        nodeId: "node_output",
        label: "Output",
        nodeType: "output",
        lane: "output",
        readiness: "ready",
        inputSummary: "Context",
        outputSummary: "Answer",
        providerRef: "",
        toolRef: "",
        ragRef: "",
        inputContractFields: ["context"],
        outputContractFields: ["answer"],
        outputMappingSummary: "Map answer.",
        riskLevel: "low",
        requiresConfirmation: false,
        previewOnlyReason: "",
      },
    ],
    edges: [{
      edgeId: "edge_input_output",
      fromNodeId: "node_input",
      toNodeId: "node_output",
      edgeKind: "context",
      conditionSummary: "Always.",
    }],
    designerLayout: {
      source: "workflow_node_designer",
      persistence: "saved_draft_metadata",
      nodePositions: [],
    },
    readiness: [],
    risks: [],
    blockedCapabilities: [],
    routeMetadata: {
      sourceRouteId: "workflow-definition-summary-list-route",
      draftRouteId: "workflow-draft-designer-offline-draft",
      routePath: "/v1/user-workspace/workflow-definitions",
      requestId: "request-revision-fixture",
      auditRef: "audit-revision-fixture",
    },
    localOnlyInteraction: "inspect_only",
    executionProfile: "review_only",
  };
}
