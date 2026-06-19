import type {
  ModelGatewayBoundaryLock,
  ModelGatewayOverviewStatus,
  ModelGatewayOverviewViewModel,
} from "./modelGatewayOverview";
import type {
  ModelGatewayRouteEvidenceViewModel,
  ModelGatewayRuntimeGuardEvidence,
} from "./modelGatewayRouteEvidence";
import type {
  ModelGatewayUsageAuditEvidenceViewModel,
  ModelGatewayUsageAuditGuard,
} from "./modelGatewayUsageAuditEvidence";

export type ModelGatewayEvidenceReviewStatus = ModelGatewayOverviewStatus;

export type ModelGatewayEvidenceReviewSource = {
  overview: ModelGatewayOverviewViewModel;
  routeEvidence: ModelGatewayRouteEvidenceViewModel;
  usageAuditEvidence: ModelGatewayUsageAuditEvidenceViewModel;
};

export type ModelGatewayEvidenceReadinessRollup = {
  rollupId: string;
  label: string;
  value: string;
  status: ModelGatewayEvidenceReviewStatus;
  sourceRef: string;
  summary: string;
};

export type ModelGatewayEvidenceChecklistItem = {
  checklistId: string;
  label: string;
  sourceSurface: "overview" | "route" | "usage" | "audit" | "lock";
  routeOrPageId: string;
  requestId: string;
  auditRef: string;
  status: ModelGatewayEvidenceReviewStatus;
  summary: string;
};

export type ModelGatewayEvidenceRisk = {
  riskId: string;
  label: string;
  sourceSurface: "route" | "usage" | "audit";
  status: "review_required" | "blocked" | "locked";
  evidenceRef: string;
  summary: string;
  reviewQuestion: string;
};

export type ModelGatewayEvidenceLockedCapability = {
  capabilityId: string;
  label: string;
  sourceSurface: "overview" | "route" | "usage";
  status: "locked";
  sourceRef: string;
  missingPrerequisite: string;
  summary: string;
};

export type ModelGatewayEvidenceReviewViewModel = {
  pageId: "model-gateway-evidence-review-offline";
  sourcePageIds: string[];
  reviewMode: "offline_gateway_evidence_review_readiness";
  tenantRef: string;
  requestId: string;
  auditRef: string;
  reviewNarrative: string;
  readinessRollup: ModelGatewayEvidenceReadinessRollup[];
  evidenceChecklist: ModelGatewayEvidenceChecklistItem[];
  routeRisks: ModelGatewayEvidenceRisk[];
  usageRisks: ModelGatewayEvidenceRisk[];
  auditRisks: ModelGatewayEvidenceRisk[];
  lockedCapabilities: ModelGatewayEvidenceLockedCapability[];
  canRenderEvidenceReview: boolean;
  canInspectGatewayEvidenceLocally: true;
  canRequestProductionGateway: false;
  canIssueApiKey: false;
  canValidateApiKey: false;
  canEnforceQuota: false;
  canWriteCostRecord: false;
  canResolveProductionSecret: false;
  canAttachDatabase: false;
  canEnableRadishAuth: false;
  canReplayRun: false;
};

export function buildModelGatewayEvidenceReviewViewModel(
  source: ModelGatewayEvidenceReviewSource,
): ModelGatewayEvidenceReviewViewModel {
  const readinessRollup = buildReadinessRollup(source);
  const evidenceChecklist = buildEvidenceChecklist(source);
  const routeRisks = buildRouteRisks(source.routeEvidence);
  const usageRisks = buildUsageRisks(source.usageAuditEvidence);
  const auditRisks = buildAuditRisks(source);
  const lockedCapabilities = buildLockedCapabilities(source);

  return {
    pageId: "model-gateway-evidence-review-offline",
    sourcePageIds: [
      source.overview.pageId,
      source.routeEvidence.pageId,
      source.usageAuditEvidence.pageId,
      "gateway-api-key-quota-readiness",
      "provider-selection-policy-v1",
      "provider-retry-fallback-policy-v1",
      "control-plane-read-formal-ui-readiness-close-v1",
    ],
    reviewMode: "offline_gateway_evidence_review_readiness",
    tenantRef: source.overview.tenantRef,
    requestId: source.overview.requestId,
    auditRef: source.overview.auditRef,
    reviewNarrative: buildReviewNarrative(source, routeRisks, usageRisks, auditRisks, lockedCapabilities),
    readinessRollup,
    evidenceChecklist,
    routeRisks,
    usageRisks,
    auditRisks,
    lockedCapabilities,
    canRenderEvidenceReview:
      source.overview.canRenderModelGatewayOverview &&
      source.routeEvidence.canRenderRouteEvidenceDetail &&
      source.usageAuditEvidence.canRenderUsageAuditEvidence &&
      readinessRollup.length >= 8 &&
      evidenceChecklist.length >= 12 &&
      routeRisks.length >= 4 &&
      usageRisks.length >= 3 &&
      auditRisks.length >= 2 &&
      lockedCapabilities.length >= 12,
    canInspectGatewayEvidenceLocally: true,
    canRequestProductionGateway: false,
    canIssueApiKey: false,
    canValidateApiKey: false,
    canEnforceQuota: false,
    canWriteCostRecord: false,
    canResolveProductionSecret: false,
    canAttachDatabase: false,
    canEnableRadishAuth: false,
    canReplayRun: false,
  };
}

function buildReviewNarrative(
  source: ModelGatewayEvidenceReviewSource,
  routeRisks: ModelGatewayEvidenceRisk[],
  usageRisks: ModelGatewayEvidenceRisk[],
  auditRisks: ModelGatewayEvidenceRisk[],
  lockedCapabilities: ModelGatewayEvidenceLockedCapability[],
): string {
  return `${source.overview.pageId}, ${source.routeEvidence.pageId}, and ${source.usageAuditEvidence.pageId} are summarized as one offline evidence review with ${routeRisks.length} route risks, ${usageRisks.length} usage risks, ${auditRisks.length} audit risks, and ${lockedCapabilities.length} locked capabilities.`;
}

function buildReadinessRollup(source: ModelGatewayEvidenceReviewSource): ModelGatewayEvidenceReadinessRollup[] {
  const blockedProfiles = source.routeEvidence.profiles.filter((profile) => profile.status === "blocked").length;
  const selectionRisks = source.routeEvidence.selectionCases.filter(
    (selectionCase) => selectionCase.status === "blocked" || selectionCase.status === "review_required",
  ).length;
  const failedTraces = source.usageAuditEvidence.traceAudit.filter((trace) => trace.failureCode !== "none").length;
  const deniedAuditEvents = source.usageAuditEvidence.auditCorrelations.filter(
    (audit) => audit.decision === "denied",
  ).length;

  return [
    {
      rollupId: "overview_ready",
      label: "Gateway overview",
      value: source.overview.canRenderModelGatewayOverview ? "rendered" : "blocked",
      status: source.overview.canRenderModelGatewayOverview ? "offline_only" : "blocked",
      sourceRef: source.overview.pageId,
      summary: "Overview supplies tenant, northbound API, policy evidence, trace evidence, and base stop lines.",
    },
    {
      rollupId: "route_evidence_ready",
      label: "Route evidence",
      value: source.routeEvidence.canRenderRouteEvidenceDetail ? "rendered" : "blocked",
      status: source.routeEvidence.canRenderRouteEvidenceDetail ? "offline_only" : "blocked",
      sourceRef: source.routeEvidence.pageId,
      summary: "Route evidence supplies provider/profile inventory, route binding, selection cases, and runtime guards.",
    },
    {
      rollupId: "usage_audit_ready",
      label: "Usage audit evidence",
      value: source.usageAuditEvidence.canRenderUsageAuditEvidence ? "rendered" : "blocked",
      status: source.usageAuditEvidence.canRenderUsageAuditEvidence ? "offline_only" : "blocked",
      sourceRef: source.usageAuditEvidence.pageId,
      summary: "Usage evidence supplies key scope, quota, cost snapshot, trace, audit, and usage locks.",
    },
    {
      rollupId: "api_surfaces",
      label: "API surfaces",
      value: String(source.overview.apiSurfaces.length),
      status: "offline_only",
      sourceRef: source.overview.pageId,
      summary: "Northbound route inventory remains compatibility evidence and does not create production gateway access.",
    },
    {
      rollupId: "provider_profiles",
      label: "Provider profiles",
      value: `${blockedProfiles}/${source.routeEvidence.profiles.length} blocked`,
      status: blockedProfiles > 0 ? "review_required" : "ready",
      sourceRef: source.routeEvidence.pageId,
      summary: "Blocked profiles are configuration evidence; they must not trigger implicit profile replacement.",
    },
    {
      rollupId: "selection_risks",
      label: "Selection cases",
      value: `${selectionRisks}/${source.routeEvidence.selectionCases.length} risk`,
      status: selectionRisks > 0 ? "review_required" : "ready",
      sourceRef: "provider-selection-policy-v1",
      summary: "Negative selection cases stay visible so route review can distinguish unsupported input from fallback.",
    },
    {
      rollupId: "trace_failures",
      label: "Trace failures",
      value: `${failedTraces}/${source.usageAuditEvidence.traceAudit.length}`,
      status: failedTraces > 0 ? "review_required" : "ready",
      sourceRef: source.usageAuditEvidence.pageId,
      summary: "Failed traces are review evidence only; retry, replay, resume, and result materialization remain unavailable.",
    },
    {
      rollupId: "audit_denials",
      label: "Audit denials",
      value: String(deniedAuditEvents),
      status: deniedAuditEvents > 0 ? "review_required" : "ready",
      sourceRef: source.usageAuditEvidence.auditRef,
      summary: "Denied audit events remain governance evidence without raw audit payload export or mutation.",
    },
    {
      rollupId: "locked_capabilities",
      label: "Locked capabilities",
      value: String(buildLockedCapabilities(source).length),
      status: "locked",
      sourceRef: "gateway-api-key-quota-readiness",
      summary: "Production gateway, key lifecycle, quota enforcement, cost writes, secret resolver, database, auth, and replay stay closed.",
    },
  ];
}

function buildEvidenceChecklist(source: ModelGatewayEvidenceReviewSource): ModelGatewayEvidenceChecklistItem[] {
  const base = {
    requestId: source.overview.requestId,
    auditRef: source.overview.auditRef,
  };

  return [
    checklistItem("overview_render", "Overview render", "overview", source.overview.pageId, base, "offline_only", "Gateway overview rendered from read-side evidence."),
    checklistItem("api_surfaces", "Northbound API surfaces", "overview", source.overview.pageId, base, "offline_only", "Models, model lookup, chat completions, responses, and messages are visible as compatibility evidence."),
    checklistItem("policy_evidence", "Policy evidence", "overview", source.overview.pageId, base, "offline_only", "Key, scope, quota, cost, trace, audit, and provider policy evidence stay read-only."),
    checklistItem("overview_trace", "Overview trace evidence", "overview", source.overview.pageId, base, "offline_only", "Run trace and cost summaries are inspectable without replay or cost writes."),
    checklistItem("overview_locks", "Overview boundary locks", "lock", source.overview.pageId, base, "locked", "Overview stop lines keep production gateway capabilities closed."),
    checklistItem("provider_profiles", "Provider/profile inventory", "route", source.routeEvidence.pageId, base, "offline_only", "Credential state, deployment mode, auth mode, and streaming evidence remain inventory only."),
    checklistItem("route_bindings", "Route bindings", "route", source.routeEvidence.pageId, base, "offline_only", "Route metadata is grouped under the overview and does not become a second product contract."),
    checklistItem("selection_cases", "Selection cases", "route", source.routeEvidence.pageId, base, "review_required", "Positive and negative selection cases explain routing risk without enabling fallback."),
    checklistItem("runtime_guards", "Runtime guards", "route", source.routeEvidence.pageId, base, "locked", "Live health, fallback, retry, secret value, and execution boundaries remain guarded."),
    checklistItem("usage_rollups", "Usage rollups", "usage", source.usageAuditEvidence.pageId, base, "offline_only", "Key scope, request usage, token usage, failure, denial, and estimated cost rollups are read-only."),
    checklistItem("key_scope", "Key scope evidence", "usage", source.usageAuditEvidence.pageId, base, "offline_only", "Key ids, owners, scopes, states, and timestamps render without raw key material."),
    checklistItem("quota_cost", "Quota and cost snapshots", "usage", source.usageAuditEvidence.pageId, base, "offline_only", "Quota and cost snapshots do not enforce quota, rate limit, billing, or durable cost writes."),
    checklistItem("trace_audit", "Trace audit evidence", "usage", source.usageAuditEvidence.pageId, base, "review_required", "Trace records expose failure classification without retry, replay, or writeback."),
    checklistItem("audit_correlation", "Audit correlation", "audit", source.usageAuditEvidence.pageId, base, "review_required", "Allowed and denied decisions are correlated without exporting raw audit payloads."),
    checklistItem("usage_guards", "Usage guard locks", "lock", source.usageAuditEvidence.pageId, base, "locked", "Raw key, quota execution, cost write, trace replay, and audit payload locks remain active."),
  ];
}

function checklistItem(
  checklistId: string,
  label: string,
  sourceSurface: ModelGatewayEvidenceChecklistItem["sourceSurface"],
  routeOrPageId: string,
  refs: Pick<ModelGatewayEvidenceChecklistItem, "requestId" | "auditRef">,
  status: ModelGatewayEvidenceReviewStatus,
  summary: string,
): ModelGatewayEvidenceChecklistItem {
  return {
    checklistId,
    label,
    sourceSurface,
    routeOrPageId,
    requestId: refs.requestId,
    auditRef: refs.auditRef,
    status,
    summary,
  };
}

function buildRouteRisks(routeEvidence: ModelGatewayRouteEvidenceViewModel): ModelGatewayEvidenceRisk[] {
  const profileRisks = routeEvidence.profiles
    .filter((profile) => profile.status === "blocked")
    .map((profile) => ({
      riskId: `profile_${profile.profileId}`,
      label: profile.profileName,
      sourceSurface: "route" as const,
      status: "blocked" as const,
      evidenceRef: profile.evidenceRef,
      summary: profile.summary,
      reviewQuestion: "Which credential or profile prerequisite blocks this provider profile?",
    }));
  const routeBindingRisks = routeEvidence.routeBindings
    .filter((route) => route.status === "review_required" || route.status === "blocked")
    .map((route) => ({
      riskId: `route_${route.routeId}`,
      label: route.path,
      sourceSurface: "route" as const,
      status: route.status === "blocked" ? ("blocked" as const) : ("review_required" as const),
      evidenceRef: route.evidenceRef,
      summary: route.summary,
      reviewQuestion: "Which provider/profile capability needs review before this route can execute?",
    }));
  const selectionRisks = routeEvidence.selectionCases
    .filter((selectionCase) => selectionCase.status === "blocked" || selectionCase.status === "review_required")
    .map((selectionCase) => ({
      riskId: `selection_${selectionCase.caseId}`,
      label: selectionCase.label,
      sourceSurface: "route" as const,
      status: selectionCase.status === "blocked" ? ("blocked" as const) : ("review_required" as const),
      evidenceRef: selectionCase.evidenceRef,
      summary: selectionCase.summary,
      reviewQuestion: "Does this case require explicit rejection rather than provider fallback?",
    }));

  return [...profileRisks, ...routeBindingRisks, ...selectionRisks].slice(0, 8);
}

function buildUsageRisks(usageAuditEvidence: ModelGatewayUsageAuditEvidenceViewModel): ModelGatewayEvidenceRisk[] {
  const rollupRisks = usageAuditEvidence.rollups
    .filter((rollup) => rollup.status === "review_required" || rollup.status === "blocked")
    .map((rollup) => ({
      riskId: `rollup_${rollup.rollupId}`,
      label: rollup.label,
      sourceSurface: "usage" as const,
      status: rollup.status === "blocked" ? ("blocked" as const) : ("review_required" as const),
      evidenceRef: rollup.sourceRef,
      summary: rollup.summary,
      reviewQuestion: "Which usage policy evidence needs operator review before production enablement?",
    }));
  const quotaLocks = usageAuditEvidence.quotaCosts
    .filter((quotaCost) => quotaCost.status === "locked" || quotaCost.status === "blocked")
    .map((quotaCost) => ({
      riskId: `quota_${quotaCost.evidenceId}`,
      label: quotaCost.label,
      sourceSurface: "usage" as const,
      status: quotaCost.status === "blocked" ? ("blocked" as const) : ("locked" as const),
      evidenceRef: quotaCost.sourceRef,
      summary: quotaCost.summary,
      reviewQuestion: "Which quota or billing prerequisite is still absent?",
    }));
  const traceRisks = usageAuditEvidence.traceAudit
    .filter((trace) => trace.status === "review_required")
    .map((trace) => ({
      riskId: `trace_${trace.traceId}`,
      label: trace.runId,
      sourceSurface: "usage" as const,
      status: "review_required" as const,
      evidenceRef: trace.auditRef,
      summary: trace.summary,
      reviewQuestion: "Which failure classification must be reviewed without retry or replay?",
    }));

  return [...rollupRisks, ...quotaLocks, ...traceRisks].slice(0, 8);
}

function buildAuditRisks(source: ModelGatewayEvidenceReviewSource): ModelGatewayEvidenceRisk[] {
  const auditDecisionRisks = source.usageAuditEvidence.auditCorrelations
    .filter((audit) => audit.status === "review_required")
    .map((audit) => ({
      riskId: `audit_${audit.auditRef}`,
      label: audit.eventKind,
      sourceSurface: "audit" as const,
      status: "review_required" as const,
      evidenceRef: audit.auditRef,
      summary: audit.summary,
      reviewQuestion: "Which denied decision or failure code should governance review first?",
    }));
  const traceRisks = source.overview.traceEvidence
    .filter((trace) => trace.status === "review_required")
    .map((trace) => ({
      riskId: `overview_trace_${trace.traceId}`,
      label: trace.runId,
      sourceSurface: "audit" as const,
      status: "review_required" as const,
      evidenceRef: trace.auditRef,
      summary: trace.summary,
      reviewQuestion: "Does this trace need audit follow-up before any future production route is enabled?",
    }));
  const auditPayloadLocks = source.usageAuditEvidence.guards
    .filter((guard) => guard.guardId === "audit_payload_locked")
    .map((guard) => ({
      riskId: `guard_${guard.guardId}`,
      label: guard.label,
      sourceSurface: "audit" as const,
      status: "locked" as const,
      evidenceRef: guard.sourceRef,
      summary: guard.summary,
      reviewQuestion: "Which export policy and sensitive projection review are still missing?",
    }));

  return [...auditDecisionRisks, ...traceRisks, ...auditPayloadLocks].slice(0, 6);
}

function buildLockedCapabilities(
  source: ModelGatewayEvidenceReviewSource,
): ModelGatewayEvidenceLockedCapability[] {
  const overviewLocks = source.overview.boundaryLocks.map((lock) => lockedFromOverview(lock));
  const routeLocks = source.routeEvidence.runtimeGuards
    .filter((guard) => guard.status === "locked")
    .map((guard) => lockedFromRouteGuard(guard));
  const usageLocks = source.usageAuditEvidence.guards
    .filter((guard) => guard.status === "locked")
    .map((guard) => lockedFromUsageGuard(guard));

  return [...overviewLocks, ...routeLocks, ...usageLocks];
}

function lockedFromOverview(lock: ModelGatewayBoundaryLock): ModelGatewayEvidenceLockedCapability {
  return {
    capabilityId: `overview_${lock.lockId}`,
    label: lock.label,
    sourceSurface: "overview",
    status: "locked",
    sourceRef: lock.lockId,
    missingPrerequisite: lock.missingPrerequisite,
    summary: lock.summary,
  };
}

function lockedFromRouteGuard(guard: ModelGatewayRuntimeGuardEvidence): ModelGatewayEvidenceLockedCapability {
  return {
    capabilityId: `route_${guard.guardId}`,
    label: guard.label,
    sourceSurface: "route",
    status: "locked",
    sourceRef: guard.sourceRef,
    missingPrerequisite: guard.enforcedBy,
    summary: guard.summary,
  };
}

function lockedFromUsageGuard(guard: ModelGatewayUsageAuditGuard): ModelGatewayEvidenceLockedCapability {
  return {
    capabilityId: `usage_${guard.guardId}`,
    label: guard.label,
    sourceSurface: "usage",
    status: "locked",
    sourceRef: guard.sourceRef,
    missingPrerequisite: guard.missingPrerequisite,
    summary: guard.summary,
  };
}
