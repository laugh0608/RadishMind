import type { WorkspaceApplicationsViewModel } from "./workspaceApplications";
import type { WorkspaceApiKeysViewModel } from "./workspaceApiKeys";
import type { WorkspaceRunHistoryViewModel } from "./workspaceRunHistory";
import type { WorkspaceWorkflowDefinitionsViewModel } from "./workspaceWorkflowDefinitions";

export type WorkspaceOperationsInboxSourceId =
  | "applications"
  | "api_keys"
  | "workflow_definitions"
  | "runs";

export type WorkspaceOperationsInboxCoverageStatus =
  | "complete_window"
  | "partial_window"
  | "unavailable";

export type WorkspaceOperationsInboxStatus = "ready" | "partial" | "blocked";
export type WorkspaceOperationsInboxSeverity = "critical" | "high" | "medium" | "info";

export type WorkspaceOperationsInboxReason =
  | "application_archived"
  | "application_run_attention"
  | "api_key_rotation_required"
  | "api_key_expired"
  | "api_key_expiring"
  | "workflow_definition_review"
  | "run_failure"
  | "run_outcome_unknown";

export type WorkspaceOperationsInboxCoverage = {
  sourceId: WorkspaceOperationsInboxSourceId;
  label: string;
  status: WorkspaceOperationsInboxCoverageStatus;
  itemCount: number;
  failureCode: string;
  requestId: string;
  auditRef: string;
  summary: string;
};

export type WorkspaceOperationsInboxItem = {
  itemId: string;
  sourceId: WorkspaceOperationsInboxSourceId;
  reason: WorkspaceOperationsInboxReason;
  severity: WorkspaceOperationsInboxSeverity;
  title: string;
  summary: string;
  resourceRef: string;
  applicationRef: string;
  workflowDefinitionId: string;
  runId: string;
  targetAnchor: "#workspace-applications" | "#workspace-api-keys" | "#workspace-workflow-definitions" | "#workspace-run-history";
  occurredAt: string;
  requestId: string;
  auditRef: string;
};

export type WorkspaceOperationsInboxMetric = {
  label: string;
  value: string;
  detail: string;
};

export type WorkspaceOperationsInboxViewModel = {
  pageId: "workspace-operations-inbox";
  activeWorkspaceId: string;
  referenceTime: string;
  status: WorkspaceOperationsInboxStatus;
  coverage: WorkspaceOperationsInboxCoverage[];
  items: WorkspaceOperationsInboxItem[];
  metrics: WorkspaceOperationsInboxMetric[];
  summary: string;
  canRenderInbox: boolean;
  currentWindowHasNoAttentionItems: boolean;
  canMutate: false;
  canAcknowledge: false;
  canRemediate: false;
  canEnforceQuota: false;
  canWriteBusinessTruth: false;
};

export type WorkspaceOperationsInboxSource = {
  activeWorkspaceId: string;
  referenceTime: string;
  sourceSnapshotReady: boolean;
  workspaceApplications: WorkspaceApplicationsViewModel;
  workspaceApiKeys: WorkspaceApiKeysViewModel;
  workspaceWorkflowDefinitions: WorkspaceWorkflowDefinitionsViewModel;
  workspaceRunHistory: WorkspaceRunHistoryViewModel;
};

const EXPIRING_WINDOW_MILLISECONDS = 14 * 24 * 60 * 60 * 1000;
const SEVERITY_ORDER: Record<WorkspaceOperationsInboxSeverity, number> = {
  critical: 0,
  high: 1,
  medium: 2,
  info: 3,
};

export function buildWorkspaceOperationsInboxViewModel(
  source: WorkspaceOperationsInboxSource,
): WorkspaceOperationsInboxViewModel {
  const referenceTime = parseReferenceTime(source.referenceTime);
  const coverage = buildCoverage(source);
  const items = [
    ...buildApplicationItems(source),
    ...buildApiKeyItems(source, referenceTime),
    ...buildWorkflowDefinitionItems(source),
    ...buildRunItems(source),
  ].sort(compareInboxItems);
  const unavailableCount = coverage.filter((entry) => entry.status === "unavailable").length;
  const partialCount = coverage.filter((entry) => entry.status === "partial_window").length;
  const status: WorkspaceOperationsInboxStatus =
    unavailableCount === coverage.length
      ? "blocked"
      : unavailableCount > 0 || partialCount > 0
        ? "partial"
        : "ready";

  return {
    pageId: "workspace-operations-inbox",
    activeWorkspaceId: source.activeWorkspaceId,
    referenceTime: source.referenceTime,
    status,
    coverage,
    items,
    metrics: buildMetrics(items, coverage),
    summary: buildSummary(source.activeWorkspaceId, items.length, unavailableCount, partialCount),
    canRenderInbox: status !== "blocked",
    currentWindowHasNoAttentionItems: status === "ready" && items.length === 0,
    canMutate: false,
    canAcknowledge: false,
    canRemediate: false,
    canEnforceQuota: false,
    canWriteBusinessTruth: false,
  };
}

function buildCoverage(source: WorkspaceOperationsInboxSource): WorkspaceOperationsInboxCoverage[] {
  return [
    toCoverage(
      "applications",
      "Applications",
      source.sourceSnapshotReady && source.workspaceApplications.canRenderApplications,
      source.workspaceApplications.applications.length,
      source.workspaceApplications.nextCursor,
      source.workspaceApplications.collection.failureCode,
      source.workspaceApplications.requestId,
      source.workspaceApplications.auditRef,
    ),
    toCoverage(
      "api_keys",
      "API Keys",
      source.sourceSnapshotReady && source.workspaceApiKeys.canRenderApiKeys,
      source.workspaceApiKeys.apiKeys.length,
      source.workspaceApiKeys.nextCursor,
      source.workspaceApiKeys.collection.failureCode,
      source.workspaceApiKeys.requestId,
      source.workspaceApiKeys.auditRef,
    ),
    toCoverage(
      "workflow_definitions",
      "Workflow Definitions",
      source.sourceSnapshotReady && source.workspaceWorkflowDefinitions.canRenderWorkflowDefinitions,
      source.workspaceWorkflowDefinitions.workflowDefinitions.length,
      source.workspaceWorkflowDefinitions.nextCursor,
      source.workspaceWorkflowDefinitions.collection.failureCode,
      source.workspaceWorkflowDefinitions.requestId,
      source.workspaceWorkflowDefinitions.auditRef,
    ),
    toCoverage(
      "runs",
      "Runs",
      source.sourceSnapshotReady && source.workspaceRunHistory.canRenderRuns,
      source.workspaceRunHistory.runs.length,
      source.workspaceRunHistory.nextCursor,
      source.workspaceRunHistory.collection.failureCode,
      source.workspaceRunHistory.requestId,
      source.workspaceRunHistory.auditRef,
    ),
  ];
}

function toCoverage(
  sourceId: WorkspaceOperationsInboxSourceId,
  label: string,
  canRender: boolean,
  itemCount: number,
  nextCursor: string | null,
  failureCode: string | null,
  requestId: string,
  auditRef: string,
): WorkspaceOperationsInboxCoverage {
  const status: WorkspaceOperationsInboxCoverageStatus = !canRender
    ? "unavailable"
    : nextCursor
      ? "partial_window"
      : "complete_window";
  const summary =
    status === "unavailable"
      ? `${label} source is unavailable; no partial rows are consumed.`
      : status === "partial_window"
        ? `${label} attention items cover only the current page.`
        : `${label} current window is complete.`;
  return {
    sourceId,
    label,
    status,
    itemCount: canRender ? itemCount : 0,
    failureCode: failureCode ?? "none",
    requestId,
    auditRef,
    summary,
  };
}

function buildApplicationItems(
  source: WorkspaceOperationsInboxSource,
): WorkspaceOperationsInboxItem[] {
  if (!source.sourceSnapshotReady || !source.workspaceApplications.canRenderApplications) {
    return [];
  }
  const items: WorkspaceOperationsInboxItem[] = [];
  for (const application of source.workspaceApplications.applications) {
    if (application.lifecycleState === "archived") {
      items.push({
        itemId: `applications:${application.applicationRef}:application_archived`,
        sourceId: "applications",
        reason: "application_archived",
        severity: "info",
        title: `${application.displayName} is archived`,
        summary: "Archived applications remain available for read-only history review.",
        resourceRef: application.applicationRef,
        applicationRef: application.applicationRef,
        workflowDefinitionId: "",
        runId: "",
        targetAnchor: "#workspace-applications",
        occurredAt: application.archivedAt ?? application.updatedAt,
        requestId: source.workspaceApplications.requestId,
        auditRef: source.workspaceApplications.auditRef,
      });
    }
    if (["failed", "blocked"].includes(application.lastRunStatus)) {
      items.push({
        itemId: `applications:${application.applicationRef}:application_run_attention`,
        sourceId: "applications",
        reason: "application_run_attention",
        severity: "high",
        title: `${application.displayName} reports ${application.lastRunStatus}`,
        summary: "Open the application and then inspect its authoritative run history.",
        resourceRef: application.applicationRef,
        applicationRef: application.applicationRef,
        workflowDefinitionId: application.latestWorkflowDefinitionRef,
        runId: "",
        targetAnchor: "#workspace-applications",
        occurredAt: application.updatedAt,
        requestId: source.workspaceApplications.requestId,
        auditRef: source.workspaceApplications.auditRef,
      });
    }
  }
  return items;
}

function buildApiKeyItems(
  source: WorkspaceOperationsInboxSource,
  referenceTime: number,
): WorkspaceOperationsInboxItem[] {
  if (!source.sourceSnapshotReady || !source.workspaceApiKeys.canRenderApiKeys) {
    return [];
  }
  const items: WorkspaceOperationsInboxItem[] = [];
  for (const apiKey of source.workspaceApiKeys.apiKeys) {
    const normalizedState = apiKey.state.trim().toLowerCase();
    if (normalizedState === "rotation_required") {
      items.push(apiKeyItem(source, apiKey.apiKeyId, "api_key_rotation_required", "high", apiKey.createdAt,
        "API key rotation is required", "Review the existing key lifecycle before issuing or revoking credentials."));
      continue;
    }
    const expiresAt = parseOptionalTime(apiKey.expiresAt);
    if (normalizedState === "expired") {
      items.push(apiKeyItem(source, apiKey.apiKeyId, "api_key_expired", "high", apiKey.expiresAt ?? apiKey.createdAt,
        "API key is expired", "Review the sanitized key record; no credential material is available here."));
      continue;
    }
    if (normalizedState !== "active") {
      continue;
    }
    if (expiresAt !== null && expiresAt <= referenceTime) {
      items.push(apiKeyItem(source, apiKey.apiKeyId, "api_key_expired", "high", apiKey.expiresAt ?? apiKey.createdAt,
        "API key is expired", "Review the sanitized key record; no credential material is available here."));
      continue;
    }
    if (expiresAt !== null && expiresAt - referenceTime <= EXPIRING_WINDOW_MILLISECONDS) {
      items.push(apiKeyItem(source, apiKey.apiKeyId, "api_key_expiring", "medium", apiKey.expiresAt ?? apiKey.createdAt,
        "API key expires within 14 days", "Review the existing key lifecycle before the explicit expiry time."));
    }
  }
  return items;
}

function apiKeyItem(
  source: WorkspaceOperationsInboxSource,
  apiKeyId: string,
  reason: WorkspaceOperationsInboxReason,
  severity: WorkspaceOperationsInboxSeverity,
  occurredAt: string,
  title: string,
  summary: string,
): WorkspaceOperationsInboxItem {
  return {
    itemId: `api_keys:${apiKeyId}:${reason}`,
    sourceId: "api_keys",
    reason,
    severity,
    title,
    summary,
    resourceRef: apiKeyId,
    applicationRef: "",
    workflowDefinitionId: "",
    runId: "",
    targetAnchor: "#workspace-api-keys",
    occurredAt,
    requestId: source.workspaceApiKeys.requestId,
    auditRef: source.workspaceApiKeys.auditRef,
  };
}

function buildWorkflowDefinitionItems(
  source: WorkspaceOperationsInboxSource,
): WorkspaceOperationsInboxItem[] {
  if (!source.sourceSnapshotReady || !source.workspaceWorkflowDefinitions.canRenderWorkflowDefinitions) {
    return [];
  }
  return source.workspaceWorkflowDefinitions.workflowDefinitions
    .filter((definition) => !["active", "published"].includes(definition.definitionStatus.trim().toLowerCase()))
    .map((definition) => ({
      itemId: `workflow_definitions:${definition.workflowDefinitionId}:workflow_definition_review`,
      sourceId: "workflow_definitions" as const,
      reason: "workflow_definition_review" as const,
      severity: "medium" as const,
      title: `${definition.workflowDefinitionId} is ${definition.definitionStatus}`,
      summary: "Review the immutable definition state; this inbox cannot approve or activate it.",
      resourceRef: definition.workflowDefinitionId,
      applicationRef: definition.applicationRef,
      workflowDefinitionId: definition.workflowDefinitionId,
      runId: "",
      targetAnchor: "#workspace-workflow-definitions" as const,
      occurredAt: definition.updatedAt,
      requestId: source.workspaceWorkflowDefinitions.requestId,
      auditRef: source.workspaceWorkflowDefinitions.auditRef,
    }));
}

function buildRunItems(source: WorkspaceOperationsInboxSource): WorkspaceOperationsInboxItem[] {
  if (!source.sourceSnapshotReady || !source.workspaceRunHistory.canRenderRuns) {
    return [];
  }
  return source.workspaceRunHistory.runs.flatMap((run) => {
    const normalizedStatus = run.status.trim().toLowerCase();
    const hasFailure = run.failureCode !== "none" || ["failed", "blocked"].includes(normalizedStatus);
    const outcomeUnknown = normalizedStatus === "outcome_unknown";
    if (!hasFailure && !outcomeUnknown) {
      return [];
    }
    return [{
      itemId: `runs:${run.runId}:${outcomeUnknown ? "run_outcome_unknown" : "run_failure"}`,
      sourceId: "runs" as const,
      reason: outcomeUnknown ? "run_outcome_unknown" as const : "run_failure" as const,
      severity: outcomeUnknown ? "high" as const : "critical" as const,
      title: outcomeUnknown ? `${run.runId} outcome is unknown` : `${run.runId} requires failure review`,
      summary: outcomeUnknown
        ? "Inspect the run metadata before deciding whether another explicit invocation is appropriate."
        : `Inspect failure ${run.failureCode}; replay, resume, and automatic remediation remain disabled.`,
      resourceRef: run.runId,
      applicationRef: run.applicationRef,
      workflowDefinitionId: run.workflowDefinitionId,
      runId: run.runId,
      targetAnchor: "#workspace-run-history" as const,
      occurredAt: run.completedAt === "still running" ? run.startedAt : run.completedAt,
      requestId: source.workspaceRunHistory.requestId,
      auditRef: source.workspaceRunHistory.auditRef,
    }];
  });
}

function buildMetrics(
  items: WorkspaceOperationsInboxItem[],
  coverage: WorkspaceOperationsInboxCoverage[],
): WorkspaceOperationsInboxMetric[] {
  return [
    {
      label: "Attention",
      value: String(items.length),
      detail: "current authorized windows only",
    },
    {
      label: "Critical",
      value: String(items.filter((item) => item.severity === "critical").length),
      detail: "failed or blocked runs",
    },
    {
      label: "Unavailable",
      value: String(coverage.filter((entry) => entry.status === "unavailable").length),
      detail: "sources excluded without partial rows",
    },
    {
      label: "Partial",
      value: String(coverage.filter((entry) => entry.status === "partial_window").length),
      detail: "sources with another cursor page",
    },
  ];
}

function buildSummary(
  activeWorkspaceId: string,
  itemCount: number,
  unavailableCount: number,
  partialCount: number,
): string {
  return `${itemCount} attention item(s) are projected for ${activeWorkspaceId}; ${unavailableCount} source(s) are unavailable and ${partialCount} source window(s) are partial.`;
}

function compareInboxItems(
  left: WorkspaceOperationsInboxItem,
  right: WorkspaceOperationsInboxItem,
): number {
  const severity = SEVERITY_ORDER[left.severity] - SEVERITY_ORDER[right.severity];
  if (severity !== 0) {
    return severity;
  }
  const occurredAt = sortableTime(right.occurredAt) - sortableTime(left.occurredAt);
  return occurredAt !== 0 ? occurredAt : left.itemId.localeCompare(right.itemId);
}

function parseReferenceTime(value: string): number {
  const parsed = Date.parse(value);
  if (!Number.isFinite(parsed)) {
    throw new RangeError("workspace operations inbox reference time must be a valid timestamp");
  }
  return parsed;
}

function parseOptionalTime(value: string | null): number | null {
  if (value === null) {
    return null;
  }
  const parsed = Date.parse(value);
  return Number.isFinite(parsed) ? parsed : null;
}

function sortableTime(value: string): number {
  const parsed = Date.parse(value);
  return Number.isFinite(parsed) ? parsed : 0;
}
