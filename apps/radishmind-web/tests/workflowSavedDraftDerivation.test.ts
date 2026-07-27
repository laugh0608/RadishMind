import assert from "node:assert/strict";
import test from "node:test";

import type { WorkflowDraftDesignerDraft } from "../src/features/control-plane-read/workflowDraftDesigner.ts";
import {
  buildDerivedWorkflowDraft,
  canDeriveSavedWorkflowDraft,
} from "../src/features/control-plane-read/workflowSavedDraftDerivation.ts";
import { saveWorkflowDraftDevRecord } from "../src/features/control-plane-read/savedWorkflowDraftConsumer.ts";

test("saved draft derivation requires a clean exact saved version", () => {
  const saved = { status: "saved_dev_record", currentDraftVersion: 3 };

  assert.equal(canDeriveSavedWorkflowDraft(saved, false, false), true);
  assert.equal(canDeriveSavedWorkflowDraft(saved, true, false), false);
  assert.equal(canDeriveSavedWorkflowDraft(saved, false, true), false);
  assert.equal(
    canDeriveSavedWorkflowDraft({ status: "validation_ready", currentDraftVersion: 3 }, false, false),
    false,
  );
  assert.equal(
    canDeriveSavedWorkflowDraft({ status: "saved_dev_record", currentDraftVersion: 0 }, false, false),
    false,
  );
});

test("derived draft gets an independent id and deep-copied editable state", () => {
  const source = sourceDraft();
  source.designerLayout = {
    source: "workflow_node_designer",
    persistence: "saved_draft_metadata",
    nodePositions: [{ nodeId: source.nodes[0]!.nodeId, x: 120, y: 80 }],
  };
  const derived = buildDerivedWorkflowDraft(source, 4, [source.draftId]);

  assert.match(derived.draftId, /^draft_derived_[0-9a-f]{8}_01$/);
  assert.notEqual(derived.draftId, source.draftId);
  assert.equal(derived.localOnlyInteraction, "local_edit");
  assert.equal(derived.designerLayout.persistence, "ui_only");
  assert.deepEqual(derived.derivation, {
    version: 1,
    sourceKind: "saved_workflow_draft",
    sourceDraftId: source.draftId,
    sourceDraftVersion: 4,
  });
  assert.equal(derived.readiness[0]?.checkId, "saved_draft_derivation_source");
  assert.notEqual(derived.nodes, source.nodes);
  assert.notEqual(derived.nodes[0], source.nodes[0]);
  assert.notEqual(derived.nodes[0]?.inputContractFields, source.nodes[0]?.inputContractFields);
  assert.notEqual(derived.edges, source.edges);
  assert.notEqual(derived.designerLayout.nodePositions, source.designerLayout.nodePositions);

  derived.nodes[0]!.label = "Derived node";
  derived.nodes[0]!.inputContractFields.push("derived_only");
  derived.edges[0]!.conditionSummary = "derived condition";
  derived.designerLayout.nodePositions[0]!.x = 999;

  assert.notEqual(source.nodes[0]!.label, "Derived node");
  assert.equal(source.nodes[0]!.inputContractFields.includes("derived_only"), false);
  assert.notEqual(source.edges[0]!.conditionSummary, "derived condition");
  assert.equal(source.designerLayout.nodePositions[0]!.x, 120);
});

test("derived draft ids increment per source without exceeding label limits", () => {
  const source = {
    ...sourceDraft(),
    label: "草案".repeat(100),
  };
  const first = buildDerivedWorkflowDraft(source, 2, [source.draftId]);
  const second = buildDerivedWorkflowDraft(source, 2, [source.draftId, first.draftId]);

  assert.match(first.draftId, /_01$/);
  assert.match(second.draftId, /_02$/);
  assert.equal(second.label.endsWith("派生 02"), true);
  assert.equal(second.label.length <= 160, true);
  assert.throws(
    () => buildDerivedWorkflowDraft(source, 0, []),
    /requires an exact source version/,
  );
});

test("derived draft save sends only the direct sanitized source reference", async (t) => {
  const source = sourceDraft();
  const derived = buildDerivedWorkflowDraft(source, 5, [source.draftId]);
  const originalFetch = globalThis.fetch;
  let requestBody: any = null;
  t.after(() => {
    globalThis.fetch = originalFetch;
  });
  globalThis.fetch = async (_input, init) => {
    requestBody = JSON.parse(String(init?.body));
    return new Response(JSON.stringify({
      request_id: "req_derivation_save",
      workspace_id: "workspace_demo",
      application_id: source.applicationRef,
      draft: null,
      failure_code: "draft_store_unavailable",
      current_draft_version: 0,
      validation_summary: { validation_state: "unavailable", valid_for_review: false },
      blocked_capabilities: [],
      audit_ref: "audit_derivation_save",
    }), { status: 200, headers: { "Content-Type": "application/json" } });
  };

  await saveWorkflowDraftDevRecord(derived, {
    mode: "dev_saved_draft_http",
    baseUrl: "http://platform.test",
    workspaceId: "workspace_demo",
    tenantRef: "tenant_demo",
    subjectRef: "subject_demo_user",
  }, 0);

  assert.deepEqual(requestBody.draft.additional_fields.derivation_v1, {
    version: 1,
    source_kind: "saved_workflow_draft",
    source_draft_id: source.draftId,
    source_draft_version: 5,
  });
  assert.equal(JSON.stringify(requestBody).includes(source.routeMetadata.requestId), false);
  assert.equal(JSON.stringify(requestBody).includes(source.routeMetadata.auditRef), false);
});

function sourceDraft(): WorkflowDraftDesignerDraft {
  return {
    draftId: "draft_source",
    templateRef: "wf_source",
    label: "RadishFlow advisory",
    applicationRef: "app_flow_copilot",
    workflowDefinitionId: "wf_radishflow_copilot_latest",
    baseDefinitionVersion: 3,
    providerProfileRef: "provider:mock",
    summary: "source",
    nodes: [
      {
        nodeId: "node_prompt",
        label: "Prompt",
        nodeType: "prompt",
        lane: "context",
        readiness: "ready",
        inputSummary: "input",
        outputSummary: "output",
        providerRef: "",
        toolRef: "",
        ragRef: "",
        inputContractFields: ["input"],
        outputContractFields: ["context"],
        outputMappingSummary: "map prompt",
        riskLevel: "low",
        requiresConfirmation: false,
        previewOnlyReason: "review",
      },
      {
        nodeId: "node_output",
        label: "Output",
        nodeType: "output",
        lane: "output",
        readiness: "ready",
        inputSummary: "context",
        outputSummary: "answer",
        providerRef: "",
        toolRef: "",
        ragRef: "",
        inputContractFields: ["context"],
        outputContractFields: ["answer"],
        outputMappingSummary: "map output",
        riskLevel: "low",
        requiresConfirmation: false,
        previewOnlyReason: "review",
      },
    ],
    edges: [
      {
        edgeId: "edge_prompt_output",
        fromNodeId: "node_prompt",
        toNodeId: "node_output",
        edgeKind: "context",
        conditionSummary: "always",
      },
    ],
    designerLayout: {
      source: "workflow_node_designer",
      persistence: "saved_draft_metadata",
      nodePositions: [{ nodeId: "node_prompt", x: 120, y: 80 }],
    },
    readiness: [],
    risks: [],
    blockedCapabilities: [],
    routeMetadata: {
      sourceRouteId: "workflow-definition-summary-list-route",
      draftRouteId: "workflow-draft-designer-offline-draft",
      routePath: "/v1/user-workspace/workflow-definitions",
      requestId: "req_source",
      auditRef: "audit_source",
    },
    localOnlyInteraction: "inspect_only",
    executionProfile: "review_only",
  };
}
