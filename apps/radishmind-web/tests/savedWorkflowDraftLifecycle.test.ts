import assert from "node:assert/strict";
import test from "node:test";

import {
  nextWorkflowSavedDraftExpectedVersion,
  resolveWorkflowSavedDraftVersion,
  workflowSavedDraftConflictRequiresResolution,
} from "../src/features/control-plane-read/savedWorkflowDraftLifecycle.ts";

test("saved draft edits keep their persisted base version", () => {
  assert.equal(nextWorkflowSavedDraftExpectedVersion({ status: "saved_dev_record", currentDraftVersion: 3 }), 3);
  assert.equal(nextWorkflowSavedDraftExpectedVersion({ status: "unsaved_local", currentDraftVersion: 3 }), 3);
  assert.equal(nextWorkflowSavedDraftExpectedVersion({ status: "validation_ready", currentDraftVersion: 3 }), 3);
  assert.equal(nextWorkflowSavedDraftExpectedVersion({ status: "unsaved_local", currentDraftVersion: 0 }), 0);
});

test("unresolved conflicts block save until the user explicitly continues", () => {
  const conflict = { status: "version_conflict", currentDraftVersion: 4 };
  assert.equal(workflowSavedDraftConflictRequiresResolution(conflict), true);
  assert.equal(nextWorkflowSavedDraftExpectedVersion(conflict), null);
  assert.equal(
    nextWorkflowSavedDraftExpectedVersion({ status: "conflict_local_continued", currentDraftVersion: 4 }),
    4,
  );
});

test("validation and non-conflict failures preserve the saved base version", () => {
  assert.equal(
    resolveWorkflowSavedDraftVersion({
      operation: "validate",
      failureCode: null,
      envelopeDraftVersion: 0,
      previousDraftVersion: 5,
    }),
    5,
  );
  assert.equal(
    resolveWorkflowSavedDraftVersion({
      operation: "save",
      failureCode: "draft_store_unavailable",
      envelopeDraftVersion: 0,
      previousDraftVersion: 5,
    }),
    5,
  );
  assert.equal(
    resolveWorkflowSavedDraftVersion({
      operation: "save",
      failureCode: "draft_version_conflict",
      envelopeDraftVersion: 6,
      previousDraftVersion: 5,
    }),
    6,
  );
  assert.equal(
    resolveWorkflowSavedDraftVersion({
      operation: "read",
      failureCode: null,
      envelopeDraftVersion: 7,
      previousDraftVersion: 5,
    }),
    7,
  );
});
