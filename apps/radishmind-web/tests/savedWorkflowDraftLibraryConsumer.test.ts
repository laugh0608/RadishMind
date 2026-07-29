import assert from "node:assert/strict";
import test from "node:test";

import {
  archiveWorkflowDraftDevRecord,
  emptyWorkflowSavedDraftLibraryFilters,
  initialWorkflowSavedDraftListState,
  listWorkflowDraftDevRecords,
  mergeWorkflowSavedDraftListPage,
  workflowSavedDraftRequestIsCurrent,
  type WorkflowSavedDraftConsumerConfig,
  type WorkflowSavedDraftListState,
  type WorkflowSavedDraftSummary,
} from "../src/features/control-plane-read/savedWorkflowDraftConsumer.ts";

const config: WorkflowSavedDraftConsumerConfig = {
  mode: "dev_saved_draft_http",
  baseUrl: "http://platform.test",
  workspaceId: "workspace_demo",
  tenantRef: "tenant_demo",
  subjectRef: "subject_demo_user",
};

test("saved draft library list sends the exact query and maps lifecycle pagination", async (t) => {
  const originalFetch = globalThis.fetch;
  t.after(() => {
    globalThis.fetch = originalFetch;
  });
  globalThis.fetch = async (input, init) => {
    const url = new URL(String(input));
    assert.equal(url.pathname, "/v1/user-workspace/workflow-drafts");
    assert.deepEqual(Object.fromEntries(url.searchParams), {
      workspace_id: "workspace_demo",
      application_id: "app_demo",
      lifecycle_state: "archived",
      limit: "25",
      cursor: "opaque-cursor",
      name_prefix: "Review",
      validation_state: "valid_for_review",
      provenance_kind: "workflow_definition",
    });
    const headers = new Headers(init?.headers);
    assert.equal(headers.get("X-RadishMind-Dev-Read-Scopes"), "workflow_drafts:read");
    assert.equal(headers.get("X-RadishMind-Active-Workspace"), null);
    return jsonResponse(listEnvelope([summaryDocument("draft_review", "archived")], {
      nextCursor: "opaque-next",
      hasMore: true,
    }));
  };

  const state = await listWorkflowDraftDevRecords("app_demo", config, {
    lifecycleState: "archived",
    filters: {
      namePrefix: " Review ",
      validationState: "valid_for_review",
      provenanceKind: "workflow_definition",
    },
    cursor: "opaque-cursor",
  });

  assert.equal(state.status, "ready");
  assert.equal(state.lifecycleState, "archived");
  assert.equal(state.nextCursor, "opaque-next");
  assert.equal(state.hasMore, true);
  assert.equal(state.summaries[0]?.lifecycleVersion, 2);
  assert.equal(state.summaries[0]?.archivedAt, "2026-07-29T08:00:00Z");
  assert.equal(state.filters.namePrefix, "Review");
});

test("saved draft library parser rejects unknown keys and inconsistent cursor state", async (t) => {
  const originalFetch = globalThis.fetch;
  t.after(() => {
    globalThis.fetch = originalFetch;
  });
  const invalidBodies = [
    { ...listEnvelope([]), unexpected: "field" },
    listEnvelope([], { nextCursor: "", hasMore: true }),
    listEnvelope([{ ...summaryDocument("draft_bad", "active"), archived_at: "2026-07-29T08:00:00Z" }]),
  ];
  for (const body of invalidBodies) {
    globalThis.fetch = async () => jsonResponse(body);
    await assert.rejects(
      () => listWorkflowDraftDevRecords("app_demo", config),
      /unexpected envelope/u,
    );
  }
});

test("load-more merge de-duplicates draft ids and refuses cross-query mixing", () => {
  const filters = emptyWorkflowSavedDraftLibraryFilters();
  const current = listState("active", [
    summary("draft_a", "active"),
    summary("draft_b", "active"),
  ], filters);
  const page = {
    ...listState("active", [
      summary("draft_b", "active"),
      summary("draft_c", "active"),
    ], filters),
    nextCursor: "opaque-next",
    hasMore: true,
  };

  const merged = mergeWorkflowSavedDraftListPage(current, page);
  assert.deepEqual(merged.summaries.map((item) => item.draftId), ["draft_a", "draft_b", "draft_c"]);
  assert.equal(merged.nextCursor, "opaque-next");

  const archived = listState("archived", [summary("draft_archived", "archived")], filters);
  assert.deepEqual(
    mergeWorkflowSavedDraftListPage(current, archived).summaries.map((item) => item.draftId),
    ["draft_archived"],
  );
});

test("late saved draft responses are rejected after generation or scope changes", () => {
  assert.equal(workflowSavedDraftRequestIsCurrent(3, 3, "workspace_a:app_a", "workspace_a:app_a"), true);
  assert.equal(workflowSavedDraftRequestIsCurrent(3, 4, "workspace_a:app_a", "workspace_a:app_a"), false);
  assert.equal(workflowSavedDraftRequestIsCurrent(3, 3, "workspace_a:app_a", "workspace_b:app_a"), false);
});

test("archive uses exact dual-version body and archive-only membership", async (t) => {
  const originalFetch = globalThis.fetch;
  t.after(() => {
    globalThis.fetch = originalFetch;
  });
  const source = summary("draft_archive", "active");
  globalThis.fetch = async (input, init) => {
    assert.match(String(input), /\/workflow-drafts\/draft_archive\/archive$/u);
    const headers = new Headers(init?.headers);
    assert.equal(headers.get("X-RadishMind-Dev-Read-Scopes"), "workflow_drafts:read,workflow_drafts:archive");
    assert.equal(
      headers.get("X-RadishMind-Dev-Read-Membership-Permissions"),
      "workflow_drafts:read,workflow_drafts:archive",
    );
    assert.deepEqual(JSON.parse(String(init?.body)), {
      workspace_id: "workspace_demo",
      application_id: "app_demo",
      expected_draft_version: 3,
      expected_lifecycle_version: 1,
    });
    return jsonResponse({
      request_id: "request_archive",
      workspace_id: "workspace_demo",
      application_id: "app_demo",
      lifecycle: {
        draft_id: "draft_archive",
        lifecycle_state: "archived",
        lifecycle_version: 2,
        archived_at: "2026-07-29T08:00:00Z",
        library_updated_at: "2026-07-29T08:00:00Z",
        lifecycle_updated_by_actor_ref: "actor_demo",
      },
      failure_code: null,
      current_draft_version: 3,
      current_lifecycle_version: 2,
      current_lifecycle_state: "archived",
      audit_ref: "audit_archive",
    });
  };

  const result = await archiveWorkflowDraftDevRecord(source, config);
  assert.equal(result.status, "archived");
  assert.equal(result.currentLifecycleVersion, 2);
  assert.equal(result.currentLifecycleState, "archived");
});

function listEnvelope(
  summaries: Record<string, unknown>[],
  options: { nextCursor?: string; hasMore?: boolean } = {},
) {
  return {
    request_id: "request_list",
    workspace_id: "workspace_demo",
    application_id: "app_demo",
    draft_summaries: summaries,
    next_cursor: options.nextCursor ?? "",
    has_more: options.hasMore ?? false,
    failure_code: null,
    audit_ref: "audit_list",
  };
}

function summaryDocument(draftId: string, lifecycleState: "active" | "archived") {
  return {
    draft_id: draftId,
    workspace_id: "workspace_demo",
    application_id: "app_demo",
    source_definition_id: "workflow_definition_demo",
    draft_version: 3,
    lifecycle_state: lifecycleState,
    lifecycle_version: lifecycleState === "active" ? 1 : 2,
    archived_at: lifecycleState === "active" ? null : "2026-07-29T08:00:00Z",
    library_updated_at: "2026-07-29T08:00:00Z",
    lifecycle_updated_by_actor_ref: lifecycleState === "active" ? "" : "actor_demo",
    provenance_kind: "workflow_definition",
    schema_version: "saved_workflow_draft.v1",
    draft_status: "valid_for_review",
    name: "Review draft",
    description: "Sanitized summary.",
    updated_at: "2026-07-29T07:00:00Z",
    updated_by_actor_ref: "actor_demo",
    node_count: 3,
    edge_count: 2,
    blocked_capability_count: 0,
    validation_state: "valid_for_review",
    valid_for_review: true,
    sample_or_unsaved_draft_status: "saved_dev_record",
  };
}

function summary(draftId: string, lifecycleState: "active" | "archived"): WorkflowSavedDraftSummary {
  const document = summaryDocument(draftId, lifecycleState);
  return {
    draftId: document.draft_id,
    workspaceId: document.workspace_id,
    applicationRef: document.application_id,
    workflowDefinitionId: document.source_definition_id,
    draftVersion: document.draft_version,
    lifecycleState,
    lifecycleVersion: document.lifecycle_version,
    archivedAt: document.archived_at,
    libraryUpdatedAt: document.library_updated_at,
    lifecycleUpdatedByActorRef: document.lifecycle_updated_by_actor_ref,
    provenanceKind: "workflow_definition",
    draftStatus: document.draft_status,
    name: document.name,
    description: document.description,
    updatedAt: document.updated_at,
    updatedByActorRef: document.updated_by_actor_ref,
    nodeCount: document.node_count,
    edgeCount: document.edge_count,
    blockedCapabilityCount: document.blocked_capability_count,
    validationState: document.validation_state,
    validForReview: document.valid_for_review,
    sampleOrUnsavedDraftStatus: document.sample_or_unsaved_draft_status,
  };
}

function listState(
  lifecycleState: "active" | "archived",
  summaries: WorkflowSavedDraftSummary[],
  filters: ReturnType<typeof emptyWorkflowSavedDraftLibraryFilters>,
): WorkflowSavedDraftListState {
  return {
    ...initialWorkflowSavedDraftListState(config, "app_demo", lifecycleState, filters),
    status: "ready",
    sourceLabel: lifecycleState,
    summary: "loaded",
    summaries,
  };
}

function jsonResponse(value: unknown): Response {
  return new Response(JSON.stringify(value), {
    status: 200,
    headers: { "Content-Type": "application/json" },
  });
}
