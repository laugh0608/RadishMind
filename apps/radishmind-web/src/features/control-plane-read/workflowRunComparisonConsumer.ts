import type { WorkflowExecutorConsumerConfig } from "./workflowExecutorConsumer.ts";

export type WorkflowRunComparisonRun = {
  runId: string;
  schemaVersion: string;
  draftId: string;
  draftVersion: number;
  executionKind: string;
  executionSourceKind: string;
  executionSourceId: string;
  executionSourceVersion: number;
  status: "running" | "succeeded" | "failed" | "canceled";
  failureCode: string;
  failureBoundary: string;
  gatewayFailureCategory: string;
  selectedProvider: string;
  selectedProfile: string;
  selectedModel: string;
  durationMs: number;
  staleRunning: boolean;
  requestId: string;
  auditRef: string;
  sideEffects: {
    retrievalCalls: number;
    providerCalls: number;
    toolCalls: 0;
    confirmationCalls: 0;
    businessWrites: 0;
    replayWrites: 0;
  };
};

export type WorkflowRunApplicationRAGAuthorityComparison = {
  assignmentId: string;
  assignmentVersion: number;
  assignmentDigest: string;
  publishCandidateId: string;
  publishReviewVersion: number;
  draftId: string;
  draftVersion: number;
  draftDigest: string;
  bindingId: string;
  bindingVersion: number;
  bindingDigest: string;
  snapshotId: string;
  snapshotVersion: number;
  snapshotDigest: string;
  ragRef: string;
  profileId: string;
  profileVersion: number;
  profileDigest: string;
  configuredProtocol: string;
  configuredModel: string;
};

export type WorkflowRunNodeComparison = {
  nodeId: string;
  nodeType: string;
  change: "added" | "removed" | "changed" | "unchanged";
  baselineStatus: string;
  candidateStatus: string;
  baselineDurationMs: number;
  candidateDurationMs: number;
  durationDeltaMs: number;
};

export type WorkflowRunRetrievalFragmentComparison = {
  fragmentRef: string;
  contentDigest: string;
  sourceType: string;
  isOfficial: boolean;
  baselineRank: number;
  candidateRank: number;
  change: "added" | "removed" | "changed" | "moved" | "unchanged";
};

export type WorkflowRunRetrievalComparison = {
  runProfile: "workflow_rag_retrieval.v1" | "workflow_rag_application_invocation.v1";
  snapshotId: string;
  snapshotVersion: number;
  snapshotDigest: string;
  ragRef: string;
  profileId: string;
  profileVersion: number;
  profileDigest: string;
  baselineAuthority: WorkflowRunApplicationRAGAuthorityComparison | null;
  candidateAuthority: WorkflowRunApplicationRAGAuthorityComparison | null;
  authorityChanged: boolean;
  queryDigest: string;
  queryBytes: number;
  retrievalNodeId: string;
  baselineAttemptStatus: string;
  candidateAttemptStatus: string;
  baselineCandidateCount: number;
  candidateCandidateCount: number;
  candidateCountDelta: number;
  baselineSelectedCount: number;
  candidateSelectedCount: number;
  baselineCitationCount: number;
  candidateCitationCount: number;
  contextBytesDelta: number;
  latencyDeltaMs: number;
  evidenceChanged: boolean;
  rankingChanged: boolean;
  citationChanged: boolean;
  citationAddedRefs: string[];
  citationRemovedRefs: string[];
  fragments: WorkflowRunRetrievalFragmentComparison[];
};

export type WorkflowRunComparison = {
  schemaVersion: "workflow_run_comparison.v1" | "workflow_run_comparison.v2" | "workflow_run_comparison.v3";
  classification: "regression" | "improvement" | "changed" | "unchanged" | "inconclusive";
  comparisonState: "comparable" | "legacy_partial" | "running_inconclusive";
  baseline: WorkflowRunComparisonRun;
  candidate: WorkflowRunComparisonRun;
  draftChanged: boolean;
  executionSourceChanged: boolean;
  providerChanged: boolean;
  modelChanged: boolean;
  statusChanged: boolean;
  failureChanged: boolean;
  durationDeltaMs: number;
  providerCallDelta: number;
  nodes: WorkflowRunNodeComparison[];
  retrieval: WorkflowRunRetrievalComparison | null;
  findings: Array<{ code: string; severity: "info" | "review_required" }>;
  recommendedReviewAction: string;
};

type ComparisonEnvelope = {
  request_id: string;
  workspace_id: string;
  application_id: string;
  comparison: ComparisonDocument | null;
  failure_code: string | null;
  failure_summary: string;
  audit_ref: string;
};

type ComparisonRunDocument = {
  run_id: string;
  schema_version: string;
  draft_id: string;
  draft_version: number;
  execution_kind?: string;
  execution_source_kind?: string;
  execution_source_id?: string;
  execution_source_version?: number;
  status: WorkflowRunComparisonRun["status"];
  failure_code: string;
  failure_boundary: string;
  gateway_failure_category: string;
  selected_provider: string;
  selected_profile: string;
  selected_model: string;
  duration_ms: number;
  stale_running: boolean;
  request_id: string;
  audit_ref: string;
  side_effects: {
    retrieval_calls?: number;
    provider_calls: number;
    tool_calls: number;
    confirmation_calls: number;
    business_writes: number;
    replay_writes: number;
  };
};

type ApplicationRAGAuthorityDocument = {
  assignment_id: string;
  assignment_version: number;
  assignment_digest: string;
  publish_candidate_id: string;
  publish_review_version: number;
  draft_id: string;
  draft_version: number;
  draft_digest: string;
  binding_ref: { binding_id: string; binding_version: number; binding_digest: string };
  snapshot_id: string;
  snapshot_version: number;
  snapshot_digest: string;
  rag_ref: string;
  profile_id: string;
  profile_version: number;
  profile_digest: string;
  configured_protocol: string;
  configured_model: string;
};

type RetrievalFragmentDocument = {
  fragment_ref: string;
  content_digest: string;
  source_type: string;
  is_official: boolean;
  baseline_rank: number;
  candidate_rank: number;
  change: WorkflowRunRetrievalFragmentComparison["change"];
};

type RetrievalDocument = {
  run_profile: "workflow_rag_retrieval.v1" | "workflow_rag_application_invocation.v1";
  snapshot_id?: string;
  snapshot_version?: number;
  snapshot_digest?: string;
  rag_ref?: string;
  profile_id?: string;
  profile_version?: number;
  profile_digest?: string;
  baseline_authority?: ApplicationRAGAuthorityDocument;
  candidate_authority?: ApplicationRAGAuthorityDocument;
  authority_changed?: boolean;
  query_digest: string;
  query_bytes: number;
  retrieval_node_id: string;
  baseline_attempt_status: string;
  candidate_attempt_status: string;
  baseline_candidate_count: number;
  candidate_candidate_count: number;
  candidate_count_delta: number;
  baseline_selected_count: number;
  candidate_selected_count: number;
  baseline_citation_count: number;
  candidate_citation_count: number;
  context_bytes_delta: number;
  latency_delta_ms: number;
  evidence_changed: boolean;
  ranking_changed: boolean;
  citation_changed: boolean;
  citation_added_refs: string[];
  citation_removed_refs: string[];
  fragments: RetrievalFragmentDocument[];
};

type ComparisonDocument = {
  schema_version: WorkflowRunComparison["schemaVersion"];
  classification: WorkflowRunComparison["classification"];
  comparison_state: WorkflowRunComparison["comparisonState"];
  baseline: ComparisonRunDocument;
  candidate: ComparisonRunDocument;
  draft_changed: boolean;
  execution_source_changed: boolean;
  provider_changed: boolean;
  model_changed: boolean;
  status_changed: boolean;
  failure_changed: boolean;
  duration_delta_ms: number;
  provider_call_delta: number;
  nodes: Array<{
    node_id: string;
    node_type: string;
    change: WorkflowRunNodeComparison["change"];
    baseline_status: string;
    candidate_status: string;
    baseline_duration_ms: number;
    candidate_duration_ms: number;
    duration_delta_ms: number;
  }>;
  retrieval?: RetrievalDocument;
  findings: Array<{ code: string; severity: "info" | "review_required" }>;
  recommended_review_action: string;
};

export async function compareWorkflowRuns(
  applicationId: string,
  baselineRunId: string,
  candidateRunId: string,
  config: WorkflowExecutorConsumerConfig,
): Promise<WorkflowRunComparison> {
  if (config.mode !== "dev_workflow_executor_http") throw new Error("Workflow run comparison is unavailable in offline mode.");
  if (!baselineRunId || !candidateRunId || baselineRunId === candidateRunId) throw new Error("Choose two different workflow runs.");
  const query = new URLSearchParams({ baseline_run_id: baselineRunId, workspace_id: config.workspaceId, application_id: applicationId });
  const response = await fetch(`${config.baseUrl}/v1/user-workspace/workflow-runs/${encodeURIComponent(candidateRunId)}/comparison?${query}`, {
    headers: comparisonHeaders(config, applicationId),
  });
  const body: unknown = await response.json();
  if (!response.ok || !isComparisonEnvelope(body)) throw new Error(`workflow run comparison route failed with HTTP ${response.status}`);
  if (body.workspace_id !== config.workspaceId || body.application_id !== applicationId) throw new Error("workflow run comparison response scope drifted");
  if (body.failure_code || !body.comparison) throw new Error(`${body.failure_code ?? "workflow_run_comparison_unavailable"}: ${body.failure_summary}`);
  return mapComparison(body.comparison);
}

function comparisonHeaders(config: WorkflowExecutorConsumerConfig, applicationId: string): HeadersInit {
  return {
    Accept: "application/json",
    "X-Request-Id": `dev-workflow-run-comparison-${applicationId}`,
    "X-RadishMind-Dev-Read-Identity": "dev-workflow-run-comparison-consumer",
    "X-RadishMind-Dev-Read-Tenant": config.tenantRef,
    "X-RadishMind-Dev-Read-Subject": config.subjectRef,
    "X-RadishMind-Dev-Read-Scopes": "workflow_runs:read",
    "X-RadishMind-Dev-Read-Audit": "audit_dev_workflow_run_comparison_consumer",
    "X-RadishMind-Dev-Workflow-Workspace": config.workspaceId,
    "X-RadishMind-Dev-Workflow-Application": applicationId,
  };
}

function mapRun(value: ComparisonRunDocument): WorkflowRunComparisonRun {
  const sideEffects = value.side_effects;
  const retrievalCalls = sideEffects.retrieval_calls ?? 0;
  if (sideEffects.tool_calls || sideEffects.confirmation_calls || sideEffects.business_writes || sideEffects.replay_writes ||
    !Number.isInteger(retrievalCalls) || retrievalCalls < 0 || retrievalCalls > 1 ||
    (value.schema_version === "workflow_run_record.v3" || value.schema_version === "workflow_run_record.v4" ? false : retrievalCalls !== 0)) {
    throw new Error("workflow run comparison contains a forbidden side effect count");
  }
  return {
    runId: value.run_id,
    schemaVersion: value.schema_version,
    draftId: value.draft_id,
    draftVersion: value.draft_version,
    executionKind: value.execution_kind ?? "",
    executionSourceKind: value.execution_source_kind ?? "",
    executionSourceId: value.execution_source_id ?? "",
    executionSourceVersion: value.execution_source_version ?? 0,
    status: value.status,
    failureCode: value.failure_code,
    failureBoundary: value.failure_boundary,
    gatewayFailureCategory: value.gateway_failure_category,
    selectedProvider: value.selected_provider,
    selectedProfile: value.selected_profile,
    selectedModel: value.selected_model,
    durationMs: value.duration_ms,
    staleRunning: value.stale_running,
    requestId: value.request_id,
    auditRef: value.audit_ref,
    sideEffects: { retrievalCalls, providerCalls: sideEffects.provider_calls, toolCalls: 0, confirmationCalls: 0, businessWrites: 0, replayWrites: 0 },
  };
}

function mapApplicationRAGAuthority(value: ApplicationRAGAuthorityDocument): WorkflowRunApplicationRAGAuthorityComparison {
  return {
    assignmentId: value.assignment_id,
    assignmentVersion: value.assignment_version,
    assignmentDigest: value.assignment_digest,
    publishCandidateId: value.publish_candidate_id,
    publishReviewVersion: value.publish_review_version,
    draftId: value.draft_id,
    draftVersion: value.draft_version,
    draftDigest: value.draft_digest,
    bindingId: value.binding_ref.binding_id,
    bindingVersion: value.binding_ref.binding_version,
    bindingDigest: value.binding_ref.binding_digest,
    snapshotId: value.snapshot_id,
    snapshotVersion: value.snapshot_version,
    snapshotDigest: value.snapshot_digest,
    ragRef: value.rag_ref,
    profileId: value.profile_id,
    profileVersion: value.profile_version,
    profileDigest: value.profile_digest,
    configuredProtocol: value.configured_protocol,
    configuredModel: value.configured_model,
  };
}

function mapRetrieval(value: RetrievalDocument): WorkflowRunRetrievalComparison {
  return {
    runProfile: value.run_profile,
    snapshotId: value.snapshot_id ?? "",
    snapshotVersion: value.snapshot_version ?? 0,
    snapshotDigest: value.snapshot_digest ?? "",
    ragRef: value.rag_ref ?? "",
    profileId: value.profile_id ?? "",
    profileVersion: value.profile_version ?? 0,
    profileDigest: value.profile_digest ?? "",
    baselineAuthority: value.baseline_authority ? mapApplicationRAGAuthority(value.baseline_authority) : null,
    candidateAuthority: value.candidate_authority ? mapApplicationRAGAuthority(value.candidate_authority) : null,
    authorityChanged: value.authority_changed ?? false,
    queryDigest: value.query_digest,
    queryBytes: value.query_bytes,
    retrievalNodeId: value.retrieval_node_id,
    baselineAttemptStatus: value.baseline_attempt_status,
    candidateAttemptStatus: value.candidate_attempt_status,
    baselineCandidateCount: value.baseline_candidate_count,
    candidateCandidateCount: value.candidate_candidate_count,
    candidateCountDelta: value.candidate_count_delta,
    baselineSelectedCount: value.baseline_selected_count,
    candidateSelectedCount: value.candidate_selected_count,
    baselineCitationCount: value.baseline_citation_count,
    candidateCitationCount: value.candidate_citation_count,
    contextBytesDelta: value.context_bytes_delta,
    latencyDeltaMs: value.latency_delta_ms,
    evidenceChanged: value.evidence_changed,
    rankingChanged: value.ranking_changed,
    citationChanged: value.citation_changed,
    citationAddedRefs: [...value.citation_added_refs],
    citationRemovedRefs: [...value.citation_removed_refs],
    fragments: value.fragments.map((fragment) => ({
      fragmentRef: fragment.fragment_ref,
      contentDigest: fragment.content_digest,
      sourceType: fragment.source_type,
      isOfficial: fragment.is_official,
      baselineRank: fragment.baseline_rank,
      candidateRank: fragment.candidate_rank,
      change: fragment.change,
    })),
  };
}

function mapComparison(value: ComparisonDocument): WorkflowRunComparison {
  return {
    schemaVersion: value.schema_version,
    classification: value.classification,
    comparisonState: value.comparison_state,
    baseline: mapRun(value.baseline),
    candidate: mapRun(value.candidate),
    draftChanged: value.draft_changed,
    executionSourceChanged: value.execution_source_changed,
    providerChanged: value.provider_changed,
    modelChanged: value.model_changed,
    statusChanged: value.status_changed,
    failureChanged: value.failure_changed,
    durationDeltaMs: value.duration_delta_ms,
    providerCallDelta: value.provider_call_delta,
    nodes: value.nodes.map((node) => ({
      nodeId: node.node_id,
      nodeType: node.node_type,
      change: node.change,
      baselineStatus: node.baseline_status,
      candidateStatus: node.candidate_status,
      baselineDurationMs: node.baseline_duration_ms,
      candidateDurationMs: node.candidate_duration_ms,
      durationDeltaMs: node.duration_delta_ms,
    })),
    retrieval: value.retrieval ? mapRetrieval(value.retrieval) : null,
    findings: value.findings,
    recommendedReviewAction: value.recommended_review_action,
  };
}

function isComparisonEnvelope(value: unknown): value is ComparisonEnvelope {
  if (!value || typeof value !== "object" || hasForbiddenComparisonKey(value)) return false;
  const item = value as Partial<ComparisonEnvelope>;
  return hasOnlyKeys(value as Record<string, unknown>, ["request_id", "workspace_id", "application_id", "comparison", "failure_code", "failure_summary", "audit_ref"]) &&
    typeof item.request_id === "string" && typeof item.workspace_id === "string" && typeof item.application_id === "string" &&
    (item.failure_code === null || typeof item.failure_code === "string") && typeof item.failure_summary === "string" &&
    typeof item.audit_ref === "string" && (item.comparison === null || isComparisonDocument(item.comparison));
}

function hasForbiddenComparisonKey(value: unknown): boolean {
  if (Array.isArray(value)) return value.some(hasForbiddenComparisonKey);
  if (!value || typeof value !== "object") return false;
  const forbidden = new Set([
    "input_text", "input_bytes", "condition_values", "condition_node_ids", "query", "raw_query",
    "fragment_content", "content", "prompt", "prompt_packet", "messages", "answer", "limitations",
    "confidence", "credential", "authorization", "cookie", "endpoint", "raw_response", "provider_raw_envelope",
    "output", "output_preview", "actor_ref",
  ]);
  return Object.entries(value as Record<string, unknown>).some(([key, child]) => forbidden.has(key.toLowerCase()) || hasForbiddenComparisonKey(child));
}

function isComparisonDocument(value: unknown): value is ComparisonDocument {
  if (!value || typeof value !== "object") return false;
  const item = value as Partial<ComparisonDocument>;
  const schemaValid = ["workflow_run_comparison.v1", "workflow_run_comparison.v2", "workflow_run_comparison.v3"].includes(item.schema_version ?? "");
  const baseKeys = ["schema_version", "classification", "comparison_state", "baseline", "candidate", "draft_changed", "execution_source_changed", "provider_changed", "model_changed", "status_changed", "failure_changed", "duration_delta_ms", "provider_call_delta", "nodes", "findings", "recommended_review_action"];
  if (!schemaValid || !hasOnlyKeys(value as Record<string, unknown>, item.schema_version === "workflow_run_comparison.v1" ? baseKeys : [...baseKeys, "retrieval"])) return false;
  const retrievalValid = item.schema_version === "workflow_run_comparison.v2" || item.schema_version === "workflow_run_comparison.v3"
    ? isRetrievalDocument(item.retrieval, item.schema_version) &&
      ((item.schema_version === "workflow_run_comparison.v2" && item.retrieval.run_profile === "workflow_rag_retrieval.v1" && item.baseline?.schema_version === "workflow_run_record.v3" && item.candidate?.schema_version === "workflow_run_record.v3") ||
        (item.schema_version === "workflow_run_comparison.v3" && item.retrieval.run_profile === "workflow_rag_application_invocation.v1" && item.baseline?.schema_version === "workflow_run_record.v4" && item.candidate?.schema_version === "workflow_run_record.v4"))
    : item.retrieval === undefined;
  return schemaValid && retrievalValid &&
    ["regression", "improvement", "changed", "unchanged", "inconclusive"].includes(item.classification ?? "") &&
    ["comparable", "legacy_partial", "running_inconclusive"].includes(item.comparison_state ?? "") &&
    isComparisonRun(item.baseline) && isComparisonRun(item.candidate) && Array.isArray(item.nodes) && item.nodes.every(isComparisonNode) &&
    Array.isArray(item.findings) && item.findings.every(isFinding) && typeof item.recommended_review_action === "string" &&
    [item.draft_changed, item.execution_source_changed, item.provider_changed, item.model_changed, item.status_changed, item.failure_changed].every((field) => typeof field === "boolean") &&
    [item.duration_delta_ms, item.provider_call_delta].every(Number.isInteger);
}

function isComparisonRun(value: unknown): value is ComparisonRunDocument {
  if (!value || typeof value !== "object") return false;
  const item = value as Partial<ComparisonRunDocument>;
  const baseKeys = ["run_id", "schema_version", "draft_id", "draft_version", "status", "failure_code", "failure_boundary", "gateway_failure_category", "selected_provider", "selected_profile", "selected_model", "duration_ms", "stale_running", "request_id", "audit_ref", "side_effects"];
  const isV4 = item.schema_version === "workflow_run_record.v4";
  if (!hasOnlyKeys(value as Record<string, unknown>, isV4 ? [...baseKeys, "execution_kind", "execution_source_kind", "execution_source_id", "execution_source_version"] : baseKeys)) return false;
  const sourceValid = isV4
    ? item.draft_id === "" && item.draft_version === 0 && item.execution_kind === "application_rag_invocation" &&
      item.execution_source_kind === "application_configuration_draft" && typeof item.execution_source_id === "string" &&
      item.execution_source_id.length > 0 && positiveInteger(item.execution_source_version)
    : typeof item.draft_id === "string" && item.draft_id.length > 0 && positiveInteger(item.draft_version);
  return typeof item.run_id === "string" && typeof item.schema_version === "string" && sourceValid &&
    ["running", "succeeded", "failed", "canceled"].includes(item.status ?? "") &&
    typeof item.failure_code === "string" && typeof item.failure_boundary === "string" && typeof item.gateway_failure_category === "string" &&
    typeof item.selected_provider === "string" && typeof item.selected_profile === "string" && typeof item.selected_model === "string" &&
    Number.isInteger(item.duration_ms) && typeof item.stale_running === "boolean" && typeof item.request_id === "string" &&
    typeof item.audit_ref === "string" && isSideEffects(item.side_effects, item.schema_version === "workflow_run_record.v3" || item.schema_version === "workflow_run_record.v4");
}

function isSideEffects(value: unknown, retrievalProfile: boolean): value is ComparisonRunDocument["side_effects"] {
  if (!value || typeof value !== "object") return false;
  const item = value as Record<string, unknown>;
  const baseKeys = ["provider_calls", "tool_calls", "confirmation_calls", "business_writes", "replay_writes"];
  return (retrievalProfile ? hasOnlyKnownKeys(item, [...baseKeys, "retrieval_calls"]) : hasOnlyKeys(item, baseKeys)) &&
    baseKeys.every((key) => Number.isInteger(item[key])) &&
    (item.retrieval_calls === undefined || Number.isInteger(item.retrieval_calls));
}

function isComparisonNode(value: unknown): value is ComparisonDocument["nodes"][number] {
  if (!value || typeof value !== "object") return false;
  const item = value as Record<string, unknown>;
  return hasOnlyKeys(item, ["node_id", "node_type", "change", "baseline_status", "candidate_status", "baseline_duration_ms", "candidate_duration_ms", "duration_delta_ms"]) &&
    typeof item.node_id === "string" && typeof item.node_type === "string" &&
    ["added", "removed", "changed", "unchanged"].includes(String(item.change)) &&
    typeof item.baseline_status === "string" && typeof item.candidate_status === "string" &&
    Number.isInteger(item.baseline_duration_ms) && Number.isInteger(item.candidate_duration_ms) && Number.isInteger(item.duration_delta_ms);
}

function isFinding(value: unknown): value is ComparisonDocument["findings"][number] {
  if (!value || typeof value !== "object") return false;
  const item = value as Record<string, unknown>;
  return hasOnlyKeys(item, ["code", "severity"]) && typeof item.code === "string" && ["info", "review_required"].includes(String(item.severity));
}

function isRetrievalDocument(value: unknown, schemaVersion: "workflow_run_comparison.v2" | "workflow_run_comparison.v3"): value is RetrievalDocument {
  if (!value || typeof value !== "object") return false;
  const item = value as Partial<RetrievalDocument>;
  const commonKeys = [
    "run_profile",
    "query_digest", "query_bytes", "retrieval_node_id", "baseline_attempt_status", "candidate_attempt_status", "baseline_candidate_count",
    "candidate_candidate_count", "candidate_count_delta", "baseline_selected_count", "candidate_selected_count", "baseline_citation_count",
    "candidate_citation_count", "context_bytes_delta", "latency_delta_ms", "evidence_changed", "ranking_changed", "citation_changed",
    "citation_added_refs", "citation_removed_refs", "fragments",
  ];
  const bindingKeys = schemaVersion === "workflow_run_comparison.v2"
    ? ["snapshot_id", "snapshot_version", "snapshot_digest", "rag_ref", "profile_id", "profile_version", "profile_digest"]
    : ["baseline_authority", "candidate_authority", "authority_changed"];
  if (!hasOnlyKeys(value as Record<string, unknown>, [...commonKeys, ...bindingKeys])) return false;
  const digest = /^sha256:[a-f0-9]{64}$/u;
  const statuses = ["not_started", "succeeded", "failed"];
  const bindingValid = schemaVersion === "workflow_run_comparison.v2"
    ? item.run_profile === "workflow_rag_retrieval.v1" && /^rags_[a-z2-7]{16}$/u.test(item.snapshot_id ?? "") && positiveInteger(item.snapshot_version) && digest.test(item.snapshot_digest ?? "") && /^workflow\.rag\.[a-z][a-z0-9_]{2,47}\.v[1-9][0-9]*$/u.test(item.rag_ref ?? "") && typeof item.profile_id === "string" && positiveInteger(item.profile_version) && digest.test(item.profile_digest ?? "")
    : item.run_profile === "workflow_rag_application_invocation.v1" && isApplicationRAGAuthority(item.baseline_authority) && isApplicationRAGAuthority(item.candidate_authority) && typeof item.authority_changed === "boolean";
  if (!bindingValid ||
    !digest.test(item.query_digest ?? "") || !nonNegativeInteger(item.query_bytes) || typeof item.retrieval_node_id !== "string" ||
    !statuses.includes(item.baseline_attempt_status ?? "") || !statuses.includes(item.candidate_attempt_status ?? "") ||
    !nonNegativeInteger(item.baseline_candidate_count) || !nonNegativeInteger(item.candidate_candidate_count) ||
    !Number.isInteger(item.candidate_count_delta) || !nonNegativeInteger(item.baseline_selected_count) ||
    !nonNegativeInteger(item.candidate_selected_count) || !nonNegativeInteger(item.baseline_citation_count) ||
    !nonNegativeInteger(item.candidate_citation_count) || !Number.isInteger(item.context_bytes_delta) ||
    !Number.isInteger(item.latency_delta_ms) || typeof item.evidence_changed !== "boolean" ||
    typeof item.ranking_changed !== "boolean" || typeof item.citation_changed !== "boolean" ||
    !Array.isArray(item.citation_added_refs) || !item.citation_added_refs.every(isFragmentRef) ||
    !Array.isArray(item.citation_removed_refs) || !item.citation_removed_refs.every(isFragmentRef) ||
    !Array.isArray(item.fragments) || !item.fragments.every(isRetrievalFragment)) return false;
  const refs = new Set(item.fragments.map((fragment) => fragment.fragment_ref));
  return new Set(item.citation_added_refs).size === item.citation_added_refs.length &&
    new Set(item.citation_removed_refs).size === item.citation_removed_refs.length &&
    item.citation_added_refs.every((ref) => refs.has(ref)) && item.citation_removed_refs.every((ref) => refs.has(ref));
}

function isApplicationRAGAuthority(value: unknown): value is ApplicationRAGAuthorityDocument {
  if (!value || typeof value !== "object") return false;
  const item = value as Partial<ApplicationRAGAuthorityDocument>;
  const digest = /^sha256:[a-f0-9]{64}$/u;
  return hasOnlyKeys(value as Record<string, unknown>, ["assignment_id", "assignment_version", "assignment_digest", "publish_candidate_id", "publish_review_version", "draft_id", "draft_version", "draft_digest", "binding_ref", "snapshot_id", "snapshot_version", "snapshot_digest", "rag_ref", "profile_id", "profile_version", "profile_digest", "configured_protocol", "configured_model"]) &&
    /^wragra_[a-z2-7]{16}$/u.test(item.assignment_id ?? "") && positiveInteger(item.assignment_version) && digest.test(item.assignment_digest ?? "") &&
    typeof item.publish_candidate_id === "string" && positiveInteger(item.publish_review_version) && typeof item.draft_id === "string" && positiveInteger(item.draft_version) && digest.test(item.draft_digest ?? "") &&
    isApplicationRAGBindingRef(item.binding_ref) && /^rags_[a-z2-7]{16}$/u.test(item.snapshot_id ?? "") && positiveInteger(item.snapshot_version) && digest.test(item.snapshot_digest ?? "") &&
    /^workflow\.rag\.[a-z][a-z0-9_]{2,47}\.v[1-9][0-9]*$/u.test(item.rag_ref ?? "") && typeof item.profile_id === "string" && positiveInteger(item.profile_version) && digest.test(item.profile_digest ?? "") &&
    typeof item.configured_protocol === "string" && typeof item.configured_model === "string";
}

function isApplicationRAGBindingRef(value: unknown): value is ApplicationRAGAuthorityDocument["binding_ref"] {
  if (!value || typeof value !== "object") return false;
  const item = value as Record<string, unknown>;
  return hasOnlyKeys(item, ["binding_id", "binding_version", "binding_digest"]) && /^wragb_[a-z2-7]{16}$/u.test(String(item.binding_id)) && item.binding_version === 1 && /^sha256:[a-f0-9]{64}$/u.test(String(item.binding_digest));
}

function isRetrievalFragment(value: unknown): value is RetrievalFragmentDocument {
  if (!value || typeof value !== "object") return false;
  const item = value as Partial<RetrievalFragmentDocument>;
  return hasOnlyKeys(value as Record<string, unknown>, ["fragment_ref", "content_digest", "source_type", "is_official", "baseline_rank", "candidate_rank", "change"]) &&
    isFragmentRef(item.fragment_ref) && /^sha256:[a-f0-9]{64}$/u.test(item.content_digest ?? "") &&
    ["document", "wiki", "faq", "forum", "manual"].includes(item.source_type ?? "") && typeof item.is_official === "boolean" &&
    nonNegativeInteger(item.baseline_rank) && nonNegativeInteger(item.candidate_rank) &&
    ["added", "removed", "changed", "moved", "unchanged"].includes(item.change ?? "");
}

function isFragmentRef(value: unknown): value is string {
  return typeof value === "string" && /^[a-z][a-z0-9_]{2,63}$/u.test(value);
}

function positiveInteger(value: unknown): boolean {
  return Number.isInteger(value) && Number(value) > 0;
}

function nonNegativeInteger(value: unknown): boolean {
  return Number.isInteger(value) && Number(value) >= 0;
}

function hasOnlyKeys(value: Record<string, unknown>, allowed: readonly string[]): boolean {
  return Object.keys(value).length === allowed.length && hasOnlyKnownKeys(value, allowed);
}

function hasOnlyKnownKeys(value: Record<string, unknown>, allowed: readonly string[]): boolean {
  const expected = new Set(allowed);
  return Object.keys(value).every((key) => expected.has(key));
}
