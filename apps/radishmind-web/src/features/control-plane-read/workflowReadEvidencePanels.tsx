import "../../i18n/readEvidenceResources.ts";
import { useTranslation } from "react-i18next";
import { workflowProjectionText } from "./workflowProjectionCopy.ts";
import { readEvidenceStatus } from "./readEvidenceMessages.ts";
import { formatDisplayDate } from "../../i18n/formatters.ts";
import { resolveUiLocale } from "../../i18n/localePreference.ts";
import type { KeyboardEvent } from "react";
import type { ControlPlaneReadRouteCard, ControlPlaneReadStatePreview } from "./readShell";
import type { WorkspaceApplicationRow, WorkspaceApplicationsStatePreview } from "./workspaceApplications";
import type { WorkflowApplicationBlockedCapabilityPreview, WorkflowApplicationDetailViewModel, WorkflowApplicationRiskSummary, WorkflowApplicationRouteMetadata } from "./workflowApplicationDetail";
import type { WorkspaceWorkflowDefinitionRow, WorkspaceWorkflowDefinitionsMetric, WorkspaceWorkflowDefinitionsStatePreview } from "./workspaceWorkflowDefinitions";
import type { WorkflowDefinitionBlockedActionPreview, WorkflowDefinitionDetailEdge, WorkflowDefinitionDetailNode, WorkflowDefinitionDetailSchemaSummary, WorkflowDefinitionDetailViewModel } from "./workflowDefinitionDetail";
import type { WorkflowSurfaceOverviewBlockedCapability, WorkflowSurfaceOverviewMetric, WorkflowSurfaceOverviewRelation, WorkflowSurfaceOverviewStatus, WorkflowSurfaceOverviewStopLine, WorkflowSurfaceOverviewViewModel } from "./workflowSurfaceOverview";

export function WorkflowSurfaceOverviewPanel({ overview }: { overview: WorkflowSurfaceOverviewViewModel }) {
  const { t } = useTranslation("workflow");
  return (
    <section
      className="surface-band workflow-surface-overview"
      id="workflow-surface-overview"
      aria-labelledby="workflow-surface-overview-title"
    >
      <div className="section-heading">
        <div>
          <p className="eyebrow">{t($ => $.readEvidence.overview)}</p>
          <h3 id="workflow-surface-overview-title">{t($ => $.readEvidence.overviewTitle)}</h3>
        </div>
        <StatusBadge tone={overview.canRenderSurfaceOverview ? "neutral" : "bad"}>
          {overview.canRenderSurfaceOverview ? t($ => $.readEvidence.states.offline_advisory) : t($ => $.readEvidence.states.blocked)}
        </StatusBadge>
      </div>

      <div className="workflow-surface-overview-summary-grid" aria-label={t($ => $.readEvidence.overviewSummary)}>
        {overview.summary.map((metric) => (
          <WorkflowSurfaceOverviewMetricCard key={metric.metricId} metric={metric} />
        ))}
      </div>

      <article className="workflow-surface-overview-card">
        <div className="workflow-surface-overview-row-main">
          <div>
            <p className="eyebrow">{readEvidenceStatus(t, overview.overviewMode)}</p>
            <h5>{overview.applicationId}</h5>
          </div>
          <StatusBadge tone="neutral">{t($ => $.readEvidence.inspectOnly)}</StatusBadge>
        </div>
        <dl className="workflow-run-guard-meta">
          <div>
            <dt>{t($ => $.readEvidence.definition)}</dt>
            <dd>{overview.workflowDefinitionId}</dd>
          </div>
          <div>
            <dt>{t($ => $.readEvidence.draft)}</dt>
            <dd>{overview.selectedDraftId}</dd>
          </div>
          <div>
            <dt>{t($ => $.readEvidence.latestRun)}</dt>
            <dd>{overview.latestRunId}</dd>
          </div>
          <div>
            <dt>{t($ => $.readEvidence.request)}</dt>
            <dd>{overview.requestId}</dd>
          </div>
          <div>
            <dt>{t($ => $.readEvidence.audit)}</dt>
            <dd>{overview.auditRef}</dd>
          </div>
        </dl>
      </article>

      <div className="workflow-surface-overview-relation-grid" aria-label={t($ => $.readEvidence.overviewRelations)}>
        {overview.relationMap.map((relation) => (
          <WorkflowSurfaceOverviewRelationCard key={relation.relationId} relation={relation} />
        ))}
      </div>

      <div
        className="workflow-surface-overview-blocked-grid"
        aria-label={t($ => $.readEvidence.overviewBlocked)}
      >
        {overview.blockedCapabilities.map((capability) => (
          <WorkflowSurfaceOverviewBlockedCapabilityCard
            key={capability.capabilityId}
            capability={capability}
          />
        ))}
      </div>

      <div className="workflow-surface-overview-stopline-grid" aria-label={t($ => $.readEvidence.overviewStops)}>
        {overview.stopLines.map((stopLine) => (
          <WorkflowSurfaceOverviewStopLineCard key={stopLine.stopLineId} stopLine={stopLine} />
        ))}
      </div>
    </section>
  );
}

function WorkflowSurfaceOverviewMetricCard({ metric }: { metric: WorkflowSurfaceOverviewMetric }) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-surface-overview-card">
      <span>{workflowProjectionText(t, metric, "label")}</span>
      <strong>{workflowProjectionText(t, metric, "value")}</strong>
      <StatusBadge tone={workflowSurfaceOverviewTone(metric.status)}>{readEvidenceStatus(t, metric.status)}</StatusBadge>
      <p>{workflowProjectionText(t, metric, "summary")}</p>
    </article>
  );
}

function WorkflowSurfaceOverviewRelationCard({ relation }: { relation: WorkflowSurfaceOverviewRelation }) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-surface-overview-relation">
      <div className="workflow-surface-overview-row-main">
        <div>
          <p className="eyebrow">{relation.relationId}</p>
          <h5>{workflowProjectionText(t, relation, "label")}</h5>
        </div>
        <StatusBadge tone={workflowSurfaceOverviewTone(relation.status)}>{readEvidenceStatus(t, relation.status)}</StatusBadge>
      </div>
      <dl className="workflow-run-guard-meta">
        <div>
          <dt>{t($ => $.readEvidence.source)}</dt>
          <dd>{relation.sourceRef}</dd>
        </div>
        <div>
          <dt>{t($ => $.readEvidence.target)}</dt>
          <dd>{relation.targetRef}</dd>
        </div>
        <div>
          <dt>{t($ => $.readEvidence.audit)}</dt>
          <dd>{relation.auditRef}</dd>
        </div>
      </dl>
      <p>{workflowProjectionText(t, relation, "summary")}</p>
    </article>
  );
}

function WorkflowSurfaceOverviewBlockedCapabilityCard({
  capability,
}: {
  capability: WorkflowSurfaceOverviewBlockedCapability;
}) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-surface-overview-blocked-capability">
      <div className="workflow-surface-overview-row-main">
        <div>
          <p className="eyebrow">{readEvidenceStatus(t, capability.sourceSurface)}</p>
          <h5>{workflowProjectionText(t, capability, "label")}</h5>
        </div>
        <StatusBadge tone="bad">{readEvidenceStatus(t, capability.status)}</StatusBadge>
      </div>
      <dl className="workflow-run-guard-meta">
        <div>
          <dt>{t($ => $.readEvidence.capability)}</dt>
          <dd>{capability.capabilityId}</dd>
        </div>
        <div>
          <dt>{t($ => $.readEvidence.missing)}</dt>
          <dd>{workflowProjectionText(t, capability, "missingPrerequisite")}</dd>
        </div>
        <div>
          <dt>{t($ => $.readEvidence.audit)}</dt>
          <dd>{capability.auditRef}</dd>
        </div>
      </dl>
      <p>{workflowProjectionText(t, capability, "summary")}</p>
    </article>
  );
}

function WorkflowSurfaceOverviewStopLineCard({ stopLine }: { stopLine: WorkflowSurfaceOverviewStopLine }) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-surface-overview-stopline">
      <div className="workflow-surface-overview-row-main">
        <div>
          <p className="eyebrow">{stopLine.stopLineId}</p>
          <h5>{workflowProjectionText(t, stopLine, "label")}</h5>
        </div>
        <StatusBadge tone="bad">{readEvidenceStatus(t, stopLine.status)}</StatusBadge>
      </div>
      <p>{workflowProjectionText(t, stopLine, "summary")}</p>
    </article>
  );
}

function workflowSurfaceOverviewTone(status: WorkflowSurfaceOverviewStatus): "good" | "bad" | "neutral" {
  if (status === "blocked") {
    return "bad";
  }
  if (status === "ready") {
    return "good";
  }
  return "neutral";
}

function handleSelectionRowKeyDown(
  event: KeyboardEvent<HTMLElement>,
  selectionId: string,
  onSelect: (selectionId: string) => void,
) {
  if (event.key !== "Enter" && event.key !== " ") {
    return;
  }
  event.preventDefault();
  onSelect(selectionId);
}

export function WorkflowDefinitionMetric({ metric }: { metric: WorkspaceWorkflowDefinitionsMetric }) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-definition-metric">
      <span>{workflowProjectionText(t, metric, "label")}</span>
      <strong>{workflowProjectionText(t, metric, "value")}</strong>
      <p>{workflowProjectionText(t, metric, "detail")}</p>
    </article>
  );
}

export function WorkflowDefinitionRow({
  workflowDefinition,
  selected,
  createdDraftCount,
  onSelectWorkflowDefinition,
  onCreateDraftForWorkflowDefinition,
}: {
  workflowDefinition: WorkspaceWorkflowDefinitionRow;
  selected: boolean;
  createdDraftCount: number;
  onSelectWorkflowDefinition: (workflowDefinitionId: string) => void;
  onCreateDraftForWorkflowDefinition: (workflowDefinitionId: string) => void;
}) {
  const { t, i18n } = useTranslation("workflow");
  return (
    <article
      className={`workflow-definition-row selection-row${selected ? " selected" : ""}`}
      role="button"
      tabIndex={0}
      aria-pressed={selected}
      onClick={() => onSelectWorkflowDefinition(workflowDefinition.workflowDefinitionId)}
      onKeyDown={(event) =>
        handleSelectionRowKeyDown(event, workflowDefinition.workflowDefinitionId, onSelectWorkflowDefinition)
      }
    >
      <div className="workflow-definition-row-main">
        <div>
          <p className="eyebrow">{workflowDefinition.applicationRef}</p>
          <h4>{workflowDefinition.workflowDefinitionId}</h4>
        </div>
        <StatusBadge tone={selected ? "neutral" : workflowDefinition.definitionStatus === "published" ? "good" : "neutral"}>
          {selected ? t($ => $.readEvidence.states.selected) : readEvidenceStatus(t, workflowDefinition.definitionStatus)}
        </StatusBadge>
      </div>
      <dl className="workflow-definition-row-meta">
        <div>
          <dt>{t($ => $.readEvidence.version)}</dt>
          <dd>{workflowDefinition.version}</dd>
        </div>
        <div>
          <dt>{t($ => $.readEvidence.nodes)}</dt>
          <dd>{workflowDefinition.nodeCount}</dd>
        </div>
        <div>
          <dt>{t($ => $.readEvidence.risk)}</dt>
          <dd>{readEvidenceStatus(t, workflowDefinition.riskLevel)}</dd>
        </div>
        <div>
          <dt>{t($ => $.readEvidence.confirmation)}</dt>
          <dd>{workflowDefinition.requiresConfirmationCapable ? t($ => $.readEvidence.states.capable) : t($ => $.readEvidence.states.not_required)}</dd>
        </div>
        <div>
          <dt>{t($ => $.readEvidence.updated)}</dt>
          <dd title={workflowDefinition.updatedAt}>{formatDisplayDate(workflowDefinition.updatedAt, resolveUiLocale(i18n.resolvedLanguage, [])) ?? t($ => $.readEvidence.unknown)}</dd>
        </div>
      </dl>
      <div className="workflow-definition-row-actions">
        <span>{t($ => $.readEvidence.localDrafts, { count: createdDraftCount })}</span>
        <button
          type="button"
          onClick={(event) => {
            event.stopPropagation();
            onCreateDraftForWorkflowDefinition(workflowDefinition.workflowDefinitionId);
          }}
        >
          {t($ => $.readEvidence.createDraft)}</button>
      </div>
    </article>
  );
}

export function WorkflowDefinitionDetailPanel({ detail }: { detail: WorkflowDefinitionDetailViewModel }) {
  const { t } = useTranslation("workflow");
  return (
    <div className="workflow-definition-detail" aria-label={t($ => $.readEvidence.definitionSurface)}>
      <div className="section-heading compact-heading">
        <div>
          <p className="eyebrow">{t($ => $.readEvidence.definitionDetail)}</p>
          <h4>{detail.workflowDefinitionId}</h4>
        </div>
        <StatusBadge tone={detail.canRenderDefinitionDetail ? "good" : "bad"}>
          {detail.canRenderDefinitionDetail ? t($ => $.readEvidence.states.detail_ready) : t($ => $.readEvidence.states.blocked)}
        </StatusBadge>
      </div>

      <div className="workflow-detail-summary-grid" aria-label={t($ => $.readEvidence.definitionSummary)}>
        <article className="workflow-detail-summary-card">
          <span>{t($ => $.readEvidence.route)}</span>
          <strong>{detail.draftRouteId}</strong>
          <p>{detail.routePath}</p>
        </article>
        <article className="workflow-detail-summary-card">
          <span>{t($ => $.readEvidence.request)}</span>
          <strong>{detail.requestId}</strong>
          <p>{detail.auditRef}</p>
        </article>
        <article className="workflow-detail-summary-card">
          <span>{t($ => $.readEvidence.risk)}</span>
          <strong>{readEvidenceStatus(t, detail.riskLevel)}</strong>
          <p>{detail.requiresConfirmationCapable ? t($ => $.readEvidence.states.confirmation_capable) : t($ => $.readEvidence.states.no_confirmation_marker)}</p>
        </article>
        <article className="workflow-detail-summary-card">
          <span>{t($ => $.readEvidence.source)}</span>
          <strong>{detail.sourceRouteId}</strong>
          <p>{detail.applicationRef}</p>
        </article>
      </div>

      <div className="workflow-detail-schema-grid" aria-label={t($ => $.readEvidence.definitionIO)}>
        <WorkflowDefinitionSchemaSummary summary={detail.inputSummary} />
        <WorkflowDefinitionSchemaSummary summary={detail.outputSummary} />
      </div>

      <div className="workflow-detail-node-list" aria-label={t($ => $.readEvidence.definitionNodes)}>
        {detail.nodes.map((node) => (
          <WorkflowDefinitionDetailNodeCard key={node.nodeId} node={node} />
        ))}
      </div>

      <div className="workflow-detail-edge-list" aria-label={t($ => $.readEvidence.definitionEdges)}>
        {detail.edges.map((edge) => (
          <WorkflowDefinitionDetailEdgeCard key={edge.edgeId} edge={edge} />
        ))}
      </div>

      <WorkflowDefinitionBlockedActionPreviewCard preview={detail.blockedActionPreview} />
    </div>
  );
}

function WorkflowDefinitionSchemaSummary({ summary }: { summary: WorkflowDefinitionDetailSchemaSummary }) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-detail-schema-card">
      <span>{workflowProjectionText(t, summary, "label")}</span>
      <strong>{summary.fields.join(", ")}</strong>
      <p>{workflowProjectionText(t, summary, "summary")}</p>
    </article>
  );
}

function WorkflowDefinitionDetailNodeCard({ node }: { node: WorkflowDefinitionDetailNode }) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-detail-node">
      <div className="workflow-detail-row-main">
        <div>
          <p className="eyebrow">{readEvidenceStatus(t, node.nodeType)}</p>
          <h5>{workflowProjectionText(t, node, "label")}</h5>
        </div>
        <StatusBadge tone={node.requiresConfirmation ? "neutral" : "good"}>
          {node.requiresConfirmation ? t($ => $.readEvidence.states.confirmation_marker) : t($ => $.readEvidence.states.read_only)}
        </StatusBadge>
      </div>
      <dl className="workflow-detail-node-meta">
        <div>
          <dt>{t($ => $.readEvidence.input)}</dt>
          <dd>{workflowProjectionText(t, node, "inputSummary")}</dd>
        </div>
        <div>
          <dt>{t($ => $.readEvidence.output)}</dt>
          <dd>{workflowProjectionText(t, node, "outputSummary")}</dd>
        </div>
        <div>
          <dt>{t($ => $.readEvidence.risk)}</dt>
          <dd>{readEvidenceStatus(t, node.riskLevel)}</dd>
        </div>
      </dl>
    </article>
  );
}

function WorkflowDefinitionDetailEdgeCard({ edge }: { edge: WorkflowDefinitionDetailEdge }) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-detail-edge">
      <span>{edge.edgeId}</span>
      <strong>
        {t($ => $.readEvidence.edge, { source: edge.fromNodeId, target: edge.toNodeId })}
      </strong>
      <p>{workflowProjectionText(t, edge, "conditionSummary")}</p>
    </article>
  );
}

function WorkflowDefinitionBlockedActionPreviewCard({
  preview,
}: {
  preview: WorkflowDefinitionBlockedActionPreview;
}) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-detail-blocked-action">
      <div className="workflow-detail-row-main">
        <div>
          <p className="eyebrow">{preview.toolRef}</p>
          <h5>{preview.toolActionId}</h5>
        </div>
        <StatusBadge tone="bad">{readEvidenceStatus(t, preview.blockedState)}</StatusBadge>
      </div>
      <dl className="workflow-detail-node-meta">
        <div>
          <dt>{t($ => $.readEvidence.action)}</dt>
          <dd>{preview.actionKind}</dd>
        </div>
        <div>
          <dt>{t($ => $.readEvidence.risk)}</dt>
          <dd>{readEvidenceStatus(t, preview.riskLevel)}</dd>
        </div>
        <div>
          <dt>{t($ => $.readEvidence.confirmation)}</dt>
          <dd>{preview.requiresConfirmation ? t($ => $.readEvidence.states.legacy_requirement_only) : t($ => $.readEvidence.states.not_required)}</dd>
        </div>
        <div>
          <dt>{t($ => $.readEvidence.audit)}</dt>
          <dd>{preview.auditRef}</dd>
        </div>
      </dl>
      <p>{workflowProjectionText(t, preview, "policyReason")}</p>
    </article>
  );
}

export function WorkflowDefinitionStatePreview({ state }: { state: WorkspaceWorkflowDefinitionsStatePreview }) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-definition-state">
      <div>
        <strong>{workflowProjectionText(t, state, "label")}</strong>
        <span>{readEvidenceStatus(t, state.status)}</span>
      </div>
      <p>{workflowProjectionText(t, state, "summary")}</p>
      <small>
        {t($ => $.readEvidence.itemFailure, { count: state.itemCount, code: state.failureCode === "none" ? t($ => $.readEvidence.none) : state.failureCode })}
      </small>
    </article>
  );
}

export function WorkflowApplicationDetailPanel({ detail }: { detail: WorkflowApplicationDetailViewModel }) {
  const { t, i18n } = useTranslation("workflow");
  return (
    <div
      className="workflow-application-detail"
      id="workflow-application-detail"
      aria-label={t($ => $.readEvidence.applicationSurface)}
    >
      <div className="section-heading compact-heading">
        <div>
          <p className="eyebrow">{t($ => $.readEvidence.applicationDetail)}</p>
          <h4>{detail.displayName}</h4>
        </div>
        <StatusBadge tone={detail.canRenderApplicationDetail ? "good" : "bad"}>
          {detail.canRenderApplicationDetail ? t($ => $.readEvidence.states.detail_ready) : t($ => $.readEvidence.states.blocked)}
        </StatusBadge>
      </div>

      <div className="workflow-application-summary-grid" aria-label={t($ => $.readEvidence.applicationIdentity)}>
        <article className="workflow-application-detail-card">
          <span>{t($ => $.readEvidence.application)}</span>
          <strong>{detail.applicationId}</strong>
          <p>{detail.displayName}</p>
        </article>
        <article className="workflow-application-detail-card">
          <span>{t($ => $.readEvidence.tenantOwner)}</span>
          <strong>{detail.tenantRef}</strong>
          <p>{detail.ownerSubjectRef}</p>
        </article>
        <article className="workflow-application-detail-card">
          <span>{t($ => $.readEvidence.applicationType)}</span>
          <strong>{readEvidenceStatus(t, detail.applicationType)}</strong>
          <p>{readEvidenceStatus(t, detail.lifecycleStatus)}</p>
        </article>
        <article className="workflow-application-detail-card">
          <span>{t($ => $.readEvidence.provider)}</span>
          <strong>{detail.providerProfileRef}</strong>
          <p>{readEvidenceStatus(t, detail.lastRunStatus)}</p>
        </article>
      </div>

      <div className="workflow-application-risk-grid" aria-label={t($ => $.readEvidence.applicationRouteRisk)}>
        <WorkflowApplicationRiskCard riskSummary={detail.riskSummary} />
        <WorkflowApplicationRouteMetadataCard routeMetadata={detail.routeMetadata} />
        <article className="workflow-application-detail-card">
          <span>{t($ => $.readEvidence.latestWorkflow)}</span>
          <strong>{detail.latestWorkflowDefinitionRef}</strong>
          <p title={detail.updatedAt}>{t($ => $.readEvidence.updatedValue, { value: formatDisplayDate(detail.updatedAt, resolveUiLocale(i18n.resolvedLanguage, [])) ?? t($ => $.readEvidence.unknown) })}</p>
        </article>
        <article className="workflow-application-detail-card">
          <span>{t($ => $.readEvidence.latestRunRef)}</span>
          <strong>{detail.latestRunRef}</strong>
          <p>{t($ => $.readEvidence.readOnlyPointer)}</p>
        </article>
      </div>

      <div className="workflow-application-blocked-grid" aria-label={t($ => $.readEvidence.applicationBlocked)}>
        {detail.blockedCapabilities.map((capability) => (
          <WorkflowApplicationBlockedCapabilityCard key={capability.capabilityId} capability={capability} />
        ))}
      </div>
    </div>
  );
}

function WorkflowApplicationRiskCard({ riskSummary }: { riskSummary: WorkflowApplicationRiskSummary }) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-application-detail-card">
      <span>{workflowProjectionText(t, riskSummary, "label")}</span>
      <strong>{readEvidenceStatus(t, riskSummary.riskLevel)}</strong>
      <p>{workflowProjectionText(t, riskSummary, "summary")}</p>
      <dl className="workflow-run-guard-meta">
        <div>
          <dt>{t($ => $.readEvidence.confirmation)}</dt>
          <dd>{riskSummary.requiresConfirmationCapable ? t($ => $.readEvidence.states.capability_marker_only) : t($ => $.readEvidence.states.not_required_by_fixture)}</dd>
        </div>
        <div>
          <dt>{t($ => $.readEvidence.policy)}</dt>
          <dd>{riskSummary.policyRef}</dd>
        </div>
      </dl>
    </article>
  );
}

function WorkflowApplicationRouteMetadataCard({
  routeMetadata,
}: {
  routeMetadata: WorkflowApplicationRouteMetadata;
}) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-application-detail-card">
      <span>{t($ => $.readEvidence.routeMetadata)}</span>
      <strong>{routeMetadata.draftRouteId}</strong>
      <p>{routeMetadata.routePath}</p>
      <dl className="workflow-run-guard-meta">
        <div>
          <dt>{t($ => $.readEvidence.sourceRoute)}</dt>
          <dd>{routeMetadata.sourceRouteId}</dd>
        </div>
        <div>
          <dt>{t($ => $.readEvidence.request)}</dt>
          <dd>{routeMetadata.requestId}</dd>
        </div>
        <div>
          <dt>{t($ => $.readEvidence.audit)}</dt>
          <dd>{routeMetadata.auditRef}</dd>
        </div>
      </dl>
    </article>
  );
}

function WorkflowApplicationBlockedCapabilityCard({
  capability,
}: {
  capability: WorkflowApplicationBlockedCapabilityPreview;
}) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-application-blocked-capability">
      <div className="workflow-application-row-main">
        <div>
          <p className="eyebrow">{capability.capabilityId}</p>
          <h5>{workflowProjectionText(t, capability, "label")}</h5>
        </div>
        <StatusBadge tone="bad">{readEvidenceStatus(t, capability.status)}</StatusBadge>
      </div>
      <dl className="workflow-run-guard-meta">
        <div>
          <dt>{t($ => $.readEvidence.missing)}</dt>
          <dd>{workflowProjectionText(t, capability, "missingPrerequisite")}</dd>
        </div>
        <div>
          <dt>{t($ => $.readEvidence.audit)}</dt>
          <dd>{capability.auditRef}</dd>
        </div>
      </dl>
      <p>{workflowProjectionText(t, capability, "reason")}</p>
    </article>
  );
}

export function ApplicationRow({
  application,
  selected,
  onSelectApplication,
}: {
  application: WorkspaceApplicationRow;
  selected: boolean;
  onSelectApplication: (applicationRef: string) => void;
}) {
  const { t, i18n } = useTranslation("workflow");
  return (
    <article
      className={`application-row selection-row${selected ? " selected" : ""}`}
      role="button"
      tabIndex={0}
      aria-pressed={selected}
      onClick={() => onSelectApplication(application.applicationRef)}
      onKeyDown={(event) => handleSelectionRowKeyDown(event, application.applicationRef, onSelectApplication)}
    >
      <div className="application-row-main">
        <div>
          <p className="eyebrow">{readEvidenceStatus(t, application.applicationKind)}</p>
          <h4>{application.displayName}</h4>
        </div>
        <StatusBadge tone={selected ? "neutral" : application.lastRunStatus === "blocked" ? "bad" : "good"}>
          {selected ? t($ => $.readEvidence.states.selected) : readEvidenceStatus(t, application.lastRunStatus)}
        </StatusBadge>
      </div>
      <dl className="application-row-meta">
        <div>
          <dt>{t($ => $.readEvidence.application)}</dt>
          <dd>{application.applicationRef}</dd>
        </div>
        <div>
          <dt>{t($ => $.readEvidence.owner)}</dt>
          <dd>{application.ownerSubjectRef}</dd>
        </div>
        <div>
          <dt>{t($ => $.readEvidence.workflow)}</dt>
          <dd>{application.latestWorkflowDefinitionRef}</dd>
        </div>
        <div>
          <dt>{t($ => $.readEvidence.updated)}</dt>
          <dd title={application.updatedAt}>{formatDisplayDate(application.updatedAt, resolveUiLocale(i18n.resolvedLanguage, [])) ?? t($ => $.readEvidence.unknown)}</dd>
        </div>
      </dl>
    </article>
  );
}

export function ApplicationStatePreview({ state }: { state: WorkspaceApplicationsStatePreview }) {
  const { t } = useTranslation("workflow");
  return (
    <article className="application-state">
      <div>
        <strong>{workflowProjectionText(t, state, "label")}</strong>
        <span>{readEvidenceStatus(t, state.status)}</span>
      </div>
      <p>{workflowProjectionText(t, state, "summary")}</p>
      <small>
        {t($ => $.readEvidence.itemFailure, { count: state.itemCount, code: state.failureCode === "none" ? t($ => $.readEvidence.none) : state.failureCode })}
      </small>
    </article>
  );
}

export function RouteCard({ route }: { route: ControlPlaneReadRouteCard }) {
  const { t } = useTranslation("workflow");
  return (
    <article className="route-card">
      <div className="card-title-row">
        <h4>{t($ => $.readEvidence.routes[route.routeId])}</h4>
        <StatusBadge tone="neutral">{readEvidenceStatus(t, route.surface)}</StatusBadge>
      </div>
      <p className="route-path">{route.path}</p>
      <dl className="route-meta">
        <div>
          <dt>{t($ => $.readEvidence.scope)}</dt>
          <dd>{route.requiredScope}</dd>
        </div>
        <div>
          <dt>{t($ => $.readEvidence.model)}</dt>
          <dd>{route.readModel}</dd>
        </div>
        <div>
          <dt>{t($ => $.readEvidence.pagination)}</dt>
          <dd>{readEvidenceStatus(t, route.pagination)}</dd>
        </div>
      </dl>
    </article>
  );
}

export function StatePreview({ state }: { state: ControlPlaneReadStatePreview }) {
  const { t } = useTranslation("workflow");
  return (
    <article className="state-card">
      <div className="card-title-row">
        <h4>{workflowProjectionText(t, state, "label")}</h4>
        <StatusBadge tone={state.tone}>{readEvidenceStatus(t, state.status)}</StatusBadge>
      </div>
      <p>{workflowProjectionText(t, state, "summary")}</p>
      <dl className="state-meta">
        <div>
          <dt>{t($ => $.readEvidence.items)}</dt>
          <dd>{state.itemCount}</dd>
        </div>
        <div>
          <dt>{t($ => $.readEvidence.failure)}</dt>
          <dd>{state.failureCode ?? t($ => $.readEvidence.states.none)}</dd>
        </div>
        <div>
          <dt>{t($ => $.readEvidence.audit)}</dt>
          <dd>{state.auditRef}</dd>
        </div>
      </dl>
    </article>
  );
}

function StatusBadge({ children, tone }: { children: string; tone: "good" | "bad" | "neutral" }) {
  return <span className={`status-badge ${tone}`}>{children}</span>;
}
