const WORKFLOW_DESIGNER_PRIMARY_HREF = "#workspace-workflow-definitions" as const;

const WORKFLOW_DESIGNER_ANCHORS = new Set([
  "#workflow-user-workspace-home",
  "#workflow-draft-designer",
  "#workflow-draft-validation-inspector",
  "#workflow-execution-plan-preview",
  "#workflow-runtime-readiness-inspector",
  "#workflow-review-handoff",
]);

export function workflowDesignerPrimaryHref(activeHash: string): typeof WORKFLOW_DESIGNER_PRIMARY_HREF | null {
  return WORKFLOW_DESIGNER_ANCHORS.has(activeHash) ? WORKFLOW_DESIGNER_PRIMARY_HREF : null;
}
