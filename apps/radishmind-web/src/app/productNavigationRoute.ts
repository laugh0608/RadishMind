const APPLICATION_API_ACCESS_PRIMARY_HREF = "#workspace-api-keys" as const;
const WORKFLOW_DESIGNER_PRIMARY_HREF = "#workspace-workflow-definitions" as const;

const APPLICATION_API_ACCESS_ANCHORS = new Set([
  "#application-api-integration",
  "#workspace-api-keys",
]);

const WORKFLOW_DESIGNER_ANCHORS = new Set([
  "#workflow-user-workspace-home",
  "#workflow-draft-designer",
  "#workflow-draft-validation-inspector",
  "#workflow-execution-plan-preview",
  "#workflow-runtime-readiness-inspector",
  "#workflow-review-handoff",
]);

export function applicationApiAccessPrimaryHref(
  activeHash: string,
): typeof APPLICATION_API_ACCESS_PRIMARY_HREF | null {
  return APPLICATION_API_ACCESS_ANCHORS.has(activeHash) ? APPLICATION_API_ACCESS_PRIMARY_HREF : null;
}

export function workflowDesignerPrimaryHref(activeHash: string): typeof WORKFLOW_DESIGNER_PRIMARY_HREF | null {
  return WORKFLOW_DESIGNER_ANCHORS.has(activeHash) ? WORKFLOW_DESIGNER_PRIMARY_HREF : null;
}
