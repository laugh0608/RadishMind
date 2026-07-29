import assert from "node:assert/strict";
import test from "node:test";

import {
  workflowSavedDraftArchiveConfirmationSummary,
  workflowSavedDraftLibraryActions,
} from "../src/features/control-plane-read/savedWorkflowDraftLibraryPresentation.ts";

test("active saved drafts open the editor and require archive confirmation", () => {
  assert.deepEqual(workflowSavedDraftLibraryActions({ lifecycleState: "active" }), {
    primaryLabel: "打开草案",
    readOnly: false,
    lifecycleLabel: "归档",
    lifecycleTarget: "archived",
    lifecycleConfirmationRequired: true,
  });
  assert.match(
    workflowSavedDraftArchiveConfirmationSummary("draft_active"),
    /当前内容、不可变历史和比较仍可只读审查/u,
  );
});

test("archived saved drafts expose read-only review and explicit unarchive", () => {
  assert.deepEqual(workflowSavedDraftLibraryActions({ lifecycleState: "archived" }), {
    primaryLabel: "只读审查",
    readOnly: true,
    lifecycleLabel: "解除归档",
    lifecycleTarget: "active",
    lifecycleConfirmationRequired: false,
  });
});
