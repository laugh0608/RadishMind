import assert from "node:assert/strict";
import test from "node:test";

import { buildWorkspaceApplicationsViewModel } from "../src/features/control-plane-read/workspaceApplications.ts";
import { buildWorkspaceApiKeysViewModel } from "../src/features/control-plane-read/workspaceApiKeys.ts";
import {
  buildWorkspaceOperationsInboxViewModel,
  type WorkspaceOperationsInboxSource,
} from "../src/features/control-plane-read/workspaceOperationsInbox.ts";
import { buildWorkspaceRunHistoryViewModel } from "../src/features/control-plane-read/workspaceRunHistory.ts";
import { buildWorkspaceWorkflowDefinitionsViewModel } from "../src/features/control-plane-read/workspaceWorkflowDefinitions.ts";

const REFERENCE_TIME = "2026-07-26T00:00:00Z";

test("projects stable cross-source attention ordering and explicit partial coverage", () => {
  const inbox = buildWorkspaceOperationsInboxViewModel(defaultSource());

  assert.equal(inbox.status, "partial");
  assert.deepEqual(
    inbox.items.map((item) => [item.sourceId, item.reason, item.severity]),
    [
      ["runs", "run_failure", "critical"],
      ["applications", "application_run_attention", "high"],
      ["api_keys", "api_key_rotation_required", "high"],
      ["workflow_definitions", "workflow_definition_review", "medium"],
    ],
  );
  assert.deepEqual(
    inbox.coverage.filter((entry) => entry.status === "partial_window").map((entry) => entry.sourceId),
    ["applications", "workflow_definitions", "runs"],
  );
  assert.equal(inbox.currentWindowHasNoAttentionItems, false);
  assert.equal(inbox.canMutate, false);
  assert.equal(inbox.canRemediate, false);
  assert.equal(inbox.canEnforceQuota, false);
  assert.equal(inbox.canWriteBusinessTruth, false);
});

test("uses the injected reference time for the inclusive fourteen-day expiry boundary", () => {
  const source = completeSource();
  source.workspaceApiKeys = {
    ...source.workspaceApiKeys,
    apiKeys: [
      {
        apiKeyId: "key_boundary",
        ownerSubjectRef: "subject_demo_user",
        scopes: ["runs:read"],
        state: "active",
        createdAt: "2026-07-01T00:00:00Z",
        expiresAt: "2026-08-09T00:00:00Z",
        lastUsedAt: null,
      },
      {
        apiKeyId: "key_later",
        ownerSubjectRef: "subject_demo_user",
        scopes: ["runs:read"],
        state: "active",
        createdAt: "2026-07-01T00:00:00Z",
        expiresAt: "2026-08-09T00:00:01Z",
        lastUsedAt: null,
      },
      {
        apiKeyId: "key_revoked",
        ownerSubjectRef: "subject_demo_user",
        scopes: ["runs:read"],
        state: "revoked",
        createdAt: "2026-07-01T00:00:00Z",
        expiresAt: "2026-07-27T00:00:00Z",
        lastUsedAt: null,
      },
    ],
  };

  const inbox = buildWorkspaceOperationsInboxViewModel(source);

  const apiKeyItems = inbox.items.filter((item) => item.sourceId === "api_keys");
  assert.deepEqual(apiKeyItems.map((item) => item.resourceRef), ["key_boundary"]);
  assert.equal(apiKeyItems[0]?.reason, "api_key_expiring");
  assert.throws(
    () => buildWorkspaceOperationsInboxViewModel({ ...source, referenceTime: "not-a-time" }),
    /reference time/,
  );
});

test("excludes rows from unavailable sources while preserving successful source items", () => {
  const source = completeSource();
  source.workspaceApplications = {
    ...source.workspaceApplications,
    canRenderApplications: false,
    applications: [{
      ...source.workspaceApplications.applications[1]!,
      displayName: "must-not-render",
    }],
    collection: {
      ...source.workspaceApplications.collection,
      failureCode: "workspace_membership_denied",
      canRenderItems: false,
      statusLabel: "denied",
    },
  };

  const inbox = buildWorkspaceOperationsInboxViewModel(source);

  assert.equal(inbox.status, "partial");
  assert.equal(inbox.coverage[0]?.status, "unavailable");
  assert.equal(inbox.coverage[0]?.failureCode, "workspace_membership_denied");
  assert.equal(inbox.items.some((item) => item.title.includes("must-not-render")), false);
  assert.equal(inbox.items.some((item) => item.sourceId === "runs"), true);
});

test("blocks when all owners are unavailable and never consumes their retained rows", () => {
  const source = completeSource();
  source.workspaceApplications = unavailableApplications(source);
  source.workspaceApiKeys = {
    ...source.workspaceApiKeys,
    canRenderApiKeys: false,
    collection: { ...source.workspaceApiKeys.collection, canRenderItems: false, statusLabel: "denied" },
  };
  source.workspaceWorkflowDefinitions = {
    ...source.workspaceWorkflowDefinitions,
    canRenderWorkflowDefinitions: false,
    collection: { ...source.workspaceWorkflowDefinitions.collection, canRenderItems: false, statusLabel: "denied" },
  };
  source.workspaceRunHistory = {
    ...source.workspaceRunHistory,
    canRenderRuns: false,
    collection: { ...source.workspaceRunHistory.collection, canRenderItems: false, statusLabel: "denied" },
  };

  const inbox = buildWorkspaceOperationsInboxViewModel(source);

  assert.equal(inbox.status, "blocked");
  assert.equal(inbox.canRenderInbox, false);
  assert.deepEqual(inbox.items, []);
  assert.equal(inbox.coverage.every((entry) => entry.status === "unavailable"), true);
});

test("keeps complete empty windows distinct from workspace health and switches workspace statelessly", () => {
  const source = completeSource();
  source.workspaceApplications = { ...source.workspaceApplications, applications: [] };
  source.workspaceApiKeys = { ...source.workspaceApiKeys, apiKeys: [] };
  source.workspaceWorkflowDefinitions = { ...source.workspaceWorkflowDefinitions, workflowDefinitions: [] };
  source.workspaceRunHistory = { ...source.workspaceRunHistory, runs: [] };

  const first = buildWorkspaceOperationsInboxViewModel(source);
  const second = buildWorkspaceOperationsInboxViewModel({
    ...source,
    activeWorkspaceId: "workspace_second",
  });

  assert.equal(first.status, "ready");
  assert.equal(first.currentWindowHasNoAttentionItems, true);
  assert.equal(second.activeWorkspaceId, "workspace_second");
  assert.equal(second.summary.includes("workspace_second"), true);
  assert.deepEqual(second.items, []);
  const serialized = JSON.stringify(second);
  for (const forbidden of ["credential_token", "authorization", "membership_assertion", "provider_endpoint", "raw_input"]) {
    assert.equal(serialized.toLowerCase().includes(forbidden), false);
  }
});

test("blocks stale fixture rows while a new live workspace snapshot is not ready", () => {
  const inbox = buildWorkspaceOperationsInboxViewModel({
    ...defaultSource(),
    activeWorkspaceId: "workspace_loading",
    sourceSnapshotReady: false,
  });

  assert.equal(inbox.status, "blocked");
  assert.deepEqual(inbox.items, []);
  assert.equal(inbox.coverage.every((entry) => entry.status === "unavailable"), true);
});

function defaultSource(): WorkspaceOperationsInboxSource {
  return {
    activeWorkspaceId: "workspace_demo",
    referenceTime: REFERENCE_TIME,
    sourceSnapshotReady: true,
    workspaceApplications: buildWorkspaceApplicationsViewModel(),
    workspaceApiKeys: buildWorkspaceApiKeysViewModel(),
    workspaceWorkflowDefinitions: buildWorkspaceWorkflowDefinitionsViewModel(),
    workspaceRunHistory: buildWorkspaceRunHistoryViewModel(),
  };
}

function completeSource(): WorkspaceOperationsInboxSource {
  const source = defaultSource();
  source.workspaceApplications = {
    ...source.workspaceApplications,
    nextCursor: null,
    collection: { ...source.workspaceApplications.collection, nextCursor: null, canFetchNextPage: false },
  };
  source.workspaceApiKeys = {
    ...source.workspaceApiKeys,
    nextCursor: null,
    collection: { ...source.workspaceApiKeys.collection, nextCursor: null, canFetchNextPage: false },
  };
  source.workspaceWorkflowDefinitions = {
    ...source.workspaceWorkflowDefinitions,
    nextCursor: null,
    collection: { ...source.workspaceWorkflowDefinitions.collection, nextCursor: null, canFetchNextPage: false },
  };
  source.workspaceRunHistory = {
    ...source.workspaceRunHistory,
    nextCursor: null,
    collection: { ...source.workspaceRunHistory.collection, nextCursor: null, canFetchNextPage: false },
  };
  return source;
}

function unavailableApplications(
  source: WorkspaceOperationsInboxSource,
): WorkspaceOperationsInboxSource["workspaceApplications"] {
  return {
    ...source.workspaceApplications,
    canRenderApplications: false,
    collection: { ...source.workspaceApplications.collection, canRenderItems: false, statusLabel: "denied" },
  };
}
