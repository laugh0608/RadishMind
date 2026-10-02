import "../../i18n/workflowDraftResources.ts";
import { workflowProjectionText, workflowProjectionStatusLabel } from "./workflowProjectionCopy.ts";
import "../../i18n/workflowHandoffProjectionResources.ts";
import "../../i18n/workflowHandoffResources.ts";
import { useTranslation } from "react-i18next";
import type {
  WorkflowReviewHandoffActiveDraftReviewSection,
  WorkflowReviewHandoffBoundaryLock,
  WorkflowReviewHandoffDecisionBlocker,
  WorkflowReviewHandoffEvidence,
  WorkflowReviewHandoffFinding,
  WorkflowReviewHandoffNodeDesignerGraphFinding,
  WorkflowReviewHandoffNodeDesignerReviewSection,
  WorkflowReviewHandoffRecipient,
  WorkflowReviewHandoffStatus,
  WorkflowReviewHandoffViewModel,
} from "./workflowReviewHandoff";
import type { WorkflowSavedDraftConflictReviewSummary } from "./savedWorkflowDraftConsumer";

type StatusBadgeTone = "good" | "bad" | "neutral";

type WorkflowReviewHandoffGraphFindingGroup = {
  targetKind: WorkflowReviewHandoffNodeDesignerGraphFinding["targetKind"];
  label: string;
  count: number;
  findings: WorkflowReviewHandoffNodeDesignerGraphFinding[];
};

export function WorkflowReviewHandoffPanel({
  handoff,
}: {
  handoff: WorkflowReviewHandoffViewModel;
}) {
  const { t } = useTranslation("workflow");
  const primaryEvidence = handoff.evidenceChecklist.slice(0, 11);
  const primaryBoundaries = handoff.boundaryLocks.slice(0, 8);
  const graphFindingGroups = workflowReviewHandoffGraphFindingGroups(handoff);

  return (
    <section
      className="surface-band workflow-review-handoff workflow-user-workspace-home"
      id="workflow-review-handoff"
      aria-labelledby="workflow-review-handoff-title"
    >
      <div className="section-heading">
        <div>
          <p className="eyebrow">{t($ => $.handoff.workflowReviewHandoff)}</p>
          <h3 id="workflow-review-handoff-title">{t($ => $.handoff.advisoryPackageForHumanReview)}</h3>
        </div>
        <StatusBadge tone={handoff.canRenderReviewHandoff ? "neutral" : "bad"}>
          {handoff.canRenderReviewHandoff ? "offline_advisory" : "blocked"}
        </StatusBadge>
      </div>

      <p className="boundary-note">{t($ => $.handoff.sourceEvidence)}</p>
      <article className="workflow-user-workspace-home-hero">
        <div>
          <p className="eyebrow">{workflowProjectionStatusLabel(t, handoff.handoffMode)}</p>
          <h4>{handoff.handoffPackageId}</h4>
          <p>{workflowProjectionText(t, handoff, "handoffNarrative")}</p>
        </div>
        <dl className="workflow-user-workspace-home-meta">
          <div>
            <dt>{t($ => $.handoff.application)}</dt>
            <dd>{handoff.applicationId}</dd>
          </div>
          <div>
            <dt>{t($ => $.handoff.workflow)}</dt>
            <dd>{handoff.workflowDefinitionId}</dd>
          </div>
          <div>
            <dt>{t($ => $.handoff.run)}</dt>
            <dd>{handoff.runId}</dd>
          </div>
          <div>
            <dt>{t($ => $.handoff.draft)}</dt>
            <dd>{handoff.draftId}</dd>
          </div>
          <div>
            <dt>{t($ => $.handoff.scenario)}</dt>
            <dd>{handoff.scenarioId}</dd>
          </div>
          <div>
            <dt>{t($ => $.handoff.audit)}</dt>
            <dd>{handoff.auditRef}</dd>
          </div>
        </dl>
      </article>

      <div className="workflow-user-workspace-home-section">
        <div className="workflow-user-workspace-home-subheading">
          <p className="eyebrow">{t($ => $.handoff.activeDraftReviewRecord)}</p>
          <h4>{t($ => $.handoff.validationPlanReadiness)}</h4>
        </div>
        <div className="workflow-user-workspace-home-route-grid" aria-label={t($ => $.handoff.workflowReviewHandoffActiveDraftRecord)}>
          {handoff.activeDraftReviewRecord.sections.map((section) => (
            <WorkflowReviewHandoffActiveDraftSectionCard key={section.sectionId} section={section} />
          ))}
        </div>
      </div>

      {handoff.savedDraftConflictReviewSummary ? (
        <div className="workflow-user-workspace-home-section">
          <div className="workflow-user-workspace-home-subheading">
            <p className="eyebrow">{t($ => $.handoff.savedDraftConflictReview)}</p>
            <h4>{t($ => $.handoff.localDraftAndSavedVersionMetadata)}</h4>
          </div>
          <WorkflowReviewHandoffSavedDraftConflictCard
            summary={handoff.savedDraftConflictReviewSummary}
          />
        </div>
      ) : null}

      <div className="workflow-user-workspace-home-section">
        <div className="workflow-user-workspace-home-subheading">
          <p className="eyebrow">{t($ => $.handoff.nodeDesignerReviewHandoff)}</p>
          <h4>{t($ => $.handoff.canvasOverlayInspectorMapping)}</h4>
        </div>
        <div className="workflow-user-workspace-home-route-grid" aria-label={t($ => $.handoff.workflowNodeDesignerReviewHandoff)}>
          {handoff.nodeDesignerReviewRecord.sections.map((section) => (
            <WorkflowReviewHandoffNodeDesignerSectionCard key={section.sectionId} section={section} />
          ))}
        </div>
        <div className="workflow-review-handoff-graph-summary" aria-label={t($ => $.handoff.workflowNodeDesignerGraphReviewSummary)}>
          {graphFindingGroups.map((group) => (
            <article key={workflowProjectionStatusLabel(t, group.targetKind)}>
              <span>{group.targetKind === "node" ? t($ => $.handoff.nodeFindings) : group.targetKind === "edge" ? t($ => $.handoff.edgeFindings) : t($ => $.handoff.graphFindings)}</span>
              <strong>{group.count}</strong>
            </article>
          ))}
        </div>
        <div className="workflow-review-handoff-graph-groups" aria-label={t($ => $.handoff.workflowNodeDesignerGraphReviewFindings)}>
          {graphFindingGroups.map((group) => (
            <section key={workflowProjectionStatusLabel(t, group.targetKind)} className="workflow-review-handoff-graph-group">
              <div className="workflow-review-handoff-graph-group-heading">
                <div>
                  <p className="eyebrow">{workflowProjectionStatusLabel(t, group.targetKind)}</p>
                  <h5>{group.targetKind === "node" ? t($ => $.handoff.nodeFindings) : group.targetKind === "edge" ? t($ => $.handoff.edgeFindings) : t($ => $.handoff.graphFindings)}</h5>
                </div>
                <StatusBadge tone={group.count > 0 ? "neutral" : "good"}>{t($ => $.handoff.findingCount, { count: group.count })}</StatusBadge>
              </div>
              <div className="workflow-user-workspace-home-readiness-grid">
                {group.findings.map((finding) => (
                  <WorkflowReviewHandoffNodeDesignerGraphFindingCard key={finding.findingId} finding={finding} />
                ))}
              </div>
            </section>
          ))}
        </div>
      </div>

      <div className="workflow-user-workspace-home-section">
        <div className="workflow-user-workspace-home-subheading">
          <p className="eyebrow">{t($ => $.handoff.reviewRecipients)}</p>
          <h4>{t($ => $.handoff.whoNeedsTheHandoff)}</h4>
        </div>
        <div className="workflow-user-workspace-home-application-grid" aria-label={t($ => $.handoff.workflowReviewHandoffRecipients)}>
          {handoff.recipients.map((recipient) => (
            <WorkflowReviewHandoffRecipientCard key={recipient.recipientId} recipient={recipient} />
          ))}
        </div>
      </div>

      <div className="workflow-user-workspace-home-section">
        <div className="workflow-user-workspace-home-subheading">
          <p className="eyebrow">{t($ => $.handoff.keyFindings)}</p>
          <h4>{t($ => $.handoff.whatTheReviewerShouldReadFirst)}</h4>
        </div>
        <div className="workflow-user-workspace-home-readiness-grid" aria-label={t($ => $.handoff.workflowReviewHandoffFindings)}>
          {handoff.keyFindings.map((finding) => (
            <WorkflowReviewHandoffFindingCard key={finding.findingId} finding={finding} />
          ))}
        </div>
      </div>

      <div className="workflow-user-workspace-home-section">
        <div className="workflow-user-workspace-home-subheading">
          <p className="eyebrow">{t($ => $.handoff.evidenceChecklist)}</p>
          <h4>{t($ => $.handoff.readSideRouteAndPageEvidence)}</h4>
        </div>
        <div className="workflow-user-workspace-home-route-grid" aria-label={t($ => $.handoff.workflowReviewHandoffEvidence)}>
          {primaryEvidence.map((evidence) => (
            <WorkflowReviewHandoffEvidenceCard key={evidence.evidenceId} evidence={evidence} />
          ))}
        </div>
      </div>

      <div className="workflow-user-workspace-home-section">
        <div className="workflow-user-workspace-home-subheading">
          <p className="eyebrow">{t($ => $.handoff.decisionBlockers)}</p>
          <h4>{t($ => $.handoff.whyThisRemainsAdvisoryOnly)}</h4>
        </div>
        <div className="workflow-user-workspace-home-readiness-grid" aria-label={t($ => $.handoff.workflowReviewHandoffBlockers)}>
          {handoff.decisionBlockers.map((blocker) => (
            <WorkflowReviewHandoffDecisionBlockerCard key={blocker.blockerId} blocker={blocker} />
          ))}
        </div>
      </div>

      <div className="workflow-user-workspace-home-section">
        <div className="workflow-user-workspace-home-subheading">
          <p className="eyebrow">{t($ => $.handoff.boundaryLocks)}</p>
          <h4>{t($ => $.handoff.capabilitiesThatStayClosed)}</h4>
        </div>
        <div className="workflow-user-workspace-home-stopline-grid" aria-label={t($ => $.handoff.workflowReviewHandoffBoundaries)}>
          {primaryBoundaries.map((boundary) => (
            <WorkflowReviewHandoffBoundaryLockCard key={boundary.boundaryId} boundary={boundary} />
          ))}
        </div>
      </div>
    </section>
  );
}

function workflowReviewHandoffGraphFindingGroups(
  handoff: WorkflowReviewHandoffViewModel,
): WorkflowReviewHandoffGraphFindingGroup[] {
  const findings = handoff.nodeDesignerReviewRecord.graphReviewFindings;
  return [
    {
      targetKind: "node",
      label: "Node graph findings",
      count: handoff.nodeDesignerReviewRecord.nodeTargetedFindingCount,
      findings: findings.filter((finding) => finding.targetKind === "node"),
    },
    {
      targetKind: "edge",
      label: "Edge graph findings",
      count: handoff.nodeDesignerReviewRecord.edgeTargetedFindingCount,
      findings: findings.filter((finding) => finding.targetKind === "edge"),
    },
    {
      targetKind: "graph",
      label: "Graph-level findings",
      count: handoff.nodeDesignerReviewRecord.graphLevelFindingCount,
      findings: findings.filter((finding) => finding.targetKind === "graph"),
    },
  ];
}

function WorkflowReviewHandoffSavedDraftConflictCard({
  summary,
}: {
  summary: WorkflowSavedDraftConflictReviewSummary;
}) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-user-workspace-home-card">
      <div className="workflow-user-workspace-home-row-main">
        <div>
          <p className="eyebrow">{summary.failureCode}</p>
          <h5>{summary.draftId}</h5>
        </div>
        <StatusBadge tone={summary.status === "local_draft_continued" ? "neutral" : "bad"}>
          {summary.status}
        </StatusBadge>
      </div>
      <dl className="workflow-user-workspace-home-meta">
        <div>
          <dt>{t($ => $.handoff.savedVersion)}</dt>
          <dd>{summary.savedDraftVersion}</dd>
        </div>
        <div>
          <dt>{t($ => $.handoff.updated)}</dt>
          <dd>{summary.savedUpdatedAt}</dd>
        </div>
        <div>
          <dt>{t($ => $.handoff.actor)}</dt>
          <dd>{summary.savedUpdatedByActorRef}</dd>
        </div>
        <div>
          <dt>{t($ => $.handoff.validation)}</dt>
          <dd>{workflowProjectionStatusLabel(t, summary.savedValidationState)}</dd>
        </div>
        <div>
          <dt>{t($ => $.handoff.blocked)}</dt>
          <dd>{summary.savedBlockedCapabilityCount ?? workflowProjectionStatusLabel(t, "not_loaded")}</dd>
        </div>
        <div>
          <dt>{t($ => $.handoff.metadata)}</dt>
          <dd>{workflowProjectionStatusLabel(t, summary.savedMetadataState)}</dd>
        </div>
        <div>
          <dt>{t($ => $.handoff.open)}</dt>
          <dd>{workflowProjectionStatusLabel(t, summary.openActionState)}</dd>
        </div>
        <div>
          <dt>{t($ => $.handoff.localGraph)}</dt>
          <dd>{t($ => $.handoff.graphSize, { nodes: summary.localNodeCount, edges: summary.localEdgeCount })}</dd>
        </div>
      </dl>
      <div className="workflow-workspace-review-token-list" aria-label={t($ => $.handoff.savedDraftConflictReviewLocks)}>
        <code>auto_overwrite_locked</code>
        <code>auto_merge_locked</code>
        <code>{workflowProjectionStatusLabel(t, summary.openActionState)}</code>
      </div>
      <p>{workflowProjectionText(t, summary, "summary")}</p>
      <p>{workflowProjectionText(t, summary, "localDraftPreservationSummary")}</p>
      {summary.openUnavailableReason ? (
        <p>{workflowProjectionText(t, summary, "openUnavailableReason")}</p>
      ) : (
        <p>{t($ => $.handoff.openSavedDraftIsAvailableFromSanitizedSavedDraftMetadata)}</p>
      )}
      <p>{workflowProjectionText(t, summary, "nextReviewerStep")}</p>
      <p>{workflowProjectionText(t, summary, "reviewerQuestion")}</p>
    </article>
  );
}

function WorkflowReviewHandoffActiveDraftSectionCard({
  section,
}: {
  section: WorkflowReviewHandoffActiveDraftReviewSection;
}) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-user-workspace-home-card">
      <div className="workflow-user-workspace-home-row-main">
        <div>
          <p className="eyebrow">{workflowProjectionStatusLabel(t, section.sourceSurface)}</p>
          <h5>{workflowProjectionText(t, section, "label")}</h5>
        </div>
        <StatusBadge tone={workflowReviewHandoffTone(section.status)}>{section.status}</StatusBadge>
      </div>
      <dl className="workflow-user-workspace-home-meta">
        <div>
          <dt>{t($ => $.handoff.primaryRef)}</dt>
          <dd>{section.primaryRef}</dd>
        </div>
        <div>
          <dt>{t($ => $.handoff.blockers)}</dt>
          <dd>{section.blockerCount}</dd>
        </div>
        <div>
          <dt>{t($ => $.handoff.request)}</dt>
          <dd>{section.requestId}</dd>
        </div>
        <div>
          <dt>{t($ => $.handoff.audit)}</dt>
          <dd>{section.auditRef}</dd>
        </div>
      </dl>
      <div className="workflow-workspace-review-token-list" aria-label={t($ => $.handoff.evidenceFor, { label: section.label })}>
        {section.evidenceRefs.map((evidenceRef) => (
          <code key={evidenceRef}>{evidenceRef}</code>
        ))}
      </div>
      <p>{workflowProjectionText(t, section, "summary")}</p>
      <p>{workflowProjectionText(t, section, "reviewerQuestion")}</p>
    </article>
  );
}

function WorkflowReviewHandoffNodeDesignerSectionCard({
  section,
}: {
  section: WorkflowReviewHandoffNodeDesignerReviewSection;
}) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-user-workspace-home-card">
      <div className="workflow-user-workspace-home-row-main">
        <div>
          <p className="eyebrow">{workflowProjectionStatusLabel(t, section.sourceSurface)}</p>
          <h5>{workflowProjectionText(t, section, "label")}</h5>
        </div>
        <StatusBadge tone={workflowReviewHandoffTone(section.status)}>{section.status}</StatusBadge>
      </div>
      <dl className="workflow-user-workspace-home-meta">
        <div>
          <dt>{t($ => $.handoff.primaryRef)}</dt>
          <dd>{section.primaryRef}</dd>
        </div>
        <div>
          <dt>{t($ => $.handoff.items)}</dt>
          <dd>{section.itemCount}</dd>
        </div>
        <div>
          <dt>{t($ => $.handoff.request)}</dt>
          <dd>{section.requestId}</dd>
        </div>
        <div>
          <dt>{t($ => $.handoff.audit)}</dt>
          <dd>{section.auditRef}</dd>
        </div>
      </dl>
      <div className="workflow-workspace-review-token-list" aria-label={t($ => $.handoff.evidenceFor, { label: section.label })}>
        {section.evidenceRefs.map((evidenceRef) => (
          <code key={evidenceRef}>{evidenceRef}</code>
        ))}
      </div>
      <p>{workflowProjectionText(t, section, "summary")}</p>
      <p>{workflowProjectionText(t, section, "reviewerQuestion")}</p>
    </article>
  );
}

function WorkflowReviewHandoffNodeDesignerGraphFindingCard({
  finding,
}: {
  finding: WorkflowReviewHandoffNodeDesignerGraphFinding;
}) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-user-workspace-home-card">
      <div className="workflow-user-workspace-home-row-main">
        <div>
          <p className="eyebrow">{workflowProjectionStatusLabel(t, finding.targetKind)}</p>
          <h5>{workflowProjectionText(t, finding, "label")}</h5>
        </div>
        <StatusBadge tone={workflowReviewHandoffTone(finding.status)}>{finding.status}</StatusBadge>
      </div>
      <dl className="workflow-user-workspace-home-meta">
        <div>
          <dt>{t($ => $.handoff.check)}</dt>
          <dd>{finding.sourceCheckId}</dd>
        </div>
        <div>
          <dt>{t($ => $.handoff.severity)}</dt>
          <dd>{workflowProjectionStatusLabel(t, finding.severity)}</dd>
        </div>
        <div>
          <dt>{t($ => $.handoff.target)}</dt>
          <dd>{workflowProjectionText(t, finding, "targetSummary")}</dd>
        </div>
        <div>
          <dt>{t($ => $.handoff.handoffPath)}</dt>
          <dd>{workflowProjectionText(t, finding, "handoffPath")}</dd>
        </div>
      </dl>
      <div className="workflow-review-handoff-graph-card-traces">
        <div>
          <span>{t($ => $.handoff.targetRefs)}</span>
          <div className="workflow-workspace-review-token-list" aria-label={t($ => $.handoff.targetsFor, { label: workflowProjectionText(t, finding, "label") })}>
            {finding.targetRefs.map((targetRef) => (
              <code key={targetRef}>{targetRef}</code>
            ))}
          </div>
        </div>
        <div>
          <span>{t($ => $.handoff.handoffPathRefs)}</span>
          <div className="workflow-workspace-review-token-list" aria-label={t($ => $.handoff.pathFor, { label: workflowProjectionText(t, finding, "label") })}>
            {finding.handoffPathRefs.map((handoffPathRef) => (
              <code key={handoffPathRef}>{handoffPathRef}</code>
            ))}
          </div>
        </div>
        <div>
          <span>{t($ => $.handoff.evidenceRefs)}</span>
          <div className="workflow-workspace-review-token-list" aria-label={t($ => $.handoff.evidenceFor, { label: workflowProjectionText(t, finding, "label") })}>
            {finding.evidenceRefs.map((evidenceRef) => (
              <code key={evidenceRef}>{evidenceRef}</code>
            ))}
          </div>
        </div>
      </div>
      <p>{workflowProjectionText(t, finding, "summary")}</p>
      <p>{workflowProjectionText(t, finding, "reviewerQuestion")}</p>
    </article>
  );
}

function WorkflowReviewHandoffRecipientCard({
  recipient,
}: {
  recipient: WorkflowReviewHandoffRecipient;
}) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-user-workspace-home-card">
      <div className="workflow-user-workspace-home-row-main">
        <div>
          <p className="eyebrow">{workflowProjectionText(t, recipient, "role")}</p>
          <h5>{workflowProjectionText(t, recipient, "label")}</h5>
        </div>
        <StatusBadge tone={workflowReviewHandoffTone(recipient.status)}>{recipient.status}</StatusBadge>
      </div>
      <dl className="workflow-user-workspace-home-meta">
        <div>
          <dt>{t($ => $.handoff.recipient)}</dt>
          <dd>{recipient.recipientId}</dd>
        </div>
        <div>
          <dt>{t($ => $.handoff.evidence)}</dt>
          <dd>{recipient.evidenceRefs.join(", ")}</dd>
        </div>
      </dl>
      <p>{workflowProjectionText(t, recipient, "handoffNeed")}</p>
    </article>
  );
}

function WorkflowReviewHandoffFindingCard({
  finding,
}: {
  finding: WorkflowReviewHandoffFinding;
}) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-user-workspace-home-card">
      <div className="workflow-user-workspace-home-row-main">
        <div>
          <p className="eyebrow">{workflowProjectionStatusLabel(t, finding.sourceSurface)}</p>
          <h5>{workflowProjectionText(t, finding, "label")}</h5>
        </div>
        <StatusBadge tone={workflowReviewHandoffTone(finding.status)}>{finding.status}</StatusBadge>
      </div>
      <dl className="workflow-user-workspace-home-meta">
        <div>
          <dt>{t($ => $.handoff.evidence)}</dt>
          <dd>{workflowProjectionText(t, finding, "evidenceRef")}</dd>
        </div>
      </dl>
      <p>{workflowProjectionText(t, finding, "summary")}</p>
      <p>{workflowProjectionText(t, finding, "humanReviewQuestion")}</p>
    </article>
  );
}

function WorkflowReviewHandoffEvidenceCard({
  evidence,
}: {
  evidence: WorkflowReviewHandoffEvidence;
}) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-user-workspace-home-card">
      <div className="workflow-user-workspace-home-row-main">
        <div>
          <p className="eyebrow">{workflowProjectionStatusLabel(t, evidence.sourceSurface)}</p>
          <h5>{workflowProjectionText(t, evidence, "label")}</h5>
        </div>
        <StatusBadge tone={workflowReviewHandoffTone(evidence.status)}>{evidence.status}</StatusBadge>
      </div>
      <dl className="workflow-user-workspace-home-meta">
        <div>
          <dt>{t($ => $.handoff.routePage)}</dt>
          <dd>{evidence.routeOrPageId}</dd>
        </div>
        <div>
          <dt>{t($ => $.handoff.request)}</dt>
          <dd>{evidence.requestId}</dd>
        </div>
        <div>
          <dt>{t($ => $.handoff.audit)}</dt>
          <dd>{evidence.auditRef}</dd>
        </div>
      </dl>
      <p>{workflowProjectionText(t, evidence, "summary")}</p>
    </article>
  );
}

function WorkflowReviewHandoffDecisionBlockerCard({
  blocker,
}: {
  blocker: WorkflowReviewHandoffDecisionBlocker;
}) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-user-workspace-home-card">
      <div className="workflow-user-workspace-home-row-main">
        <div>
          <p className="eyebrow">{workflowProjectionStatusLabel(t, blocker.sourceSurface)}</p>
          <h5>{workflowProjectionText(t, blocker, "label")}</h5>
        </div>
        <StatusBadge tone="bad">{blocker.status}</StatusBadge>
      </div>
      <dl className="workflow-user-workspace-home-meta">
        <div>
          <dt>{t($ => $.handoff.missing)}</dt>
          <dd>{workflowProjectionText(t, blocker, "missingPrerequisite")}</dd>
        </div>
      </dl>
      <div className="workflow-workspace-review-token-list" aria-label={t($ => $.handoff.auditFor, { label: workflowProjectionText(t, blocker, "label") })}>
        {blocker.auditRefs.map((auditRef) => (
          <code key={auditRef}>{auditRef}</code>
        ))}
      </div>
      <p>{workflowProjectionText(t, blocker, "summary")}</p>
    </article>
  );
}

function WorkflowReviewHandoffBoundaryLockCard({
  boundary,
}: {
  boundary: WorkflowReviewHandoffBoundaryLock;
}) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-user-workspace-home-card">
      <div className="workflow-user-workspace-home-row-main">
        <div>
          <p className="eyebrow">{boundary.boundaryId}</p>
          <h5>{workflowProjectionText(t, boundary, "label")}</h5>
        </div>
        <StatusBadge tone="bad">{boundary.status}</StatusBadge>
      </div>
      <p>{workflowProjectionText(t, boundary, "summary")}</p>
    </article>
  );
}

function StatusBadge({ children, tone }: { children: string; tone: StatusBadgeTone }) {
  const { t } = useTranslation("workflow");
  return <span className={`status-badge ${tone}`}>{workflowProjectionStatusLabel(t, children)}</span>;
}

function workflowReviewHandoffTone(status: WorkflowReviewHandoffStatus): StatusBadgeTone {
  if (status === "blocked" || status === "locked") {
    return "bad";
  }
  if (status === "ready") {
    return "good";
  }
  return "neutral";
}
