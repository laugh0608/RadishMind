import type {
  WorkflowWorkspaceReviewBlockedCapabilityGroup,
  WorkflowWorkspaceReviewContextCard,
  WorkflowWorkspaceReviewRelation,
  WorkflowWorkspaceReviewStage,
  WorkflowWorkspaceReviewStatus,
  WorkflowWorkspaceReviewStopLine,
  WorkflowWorkspaceReviewViewModel,
} from "./workflowWorkspaceReview";

type StatusBadgeTone = "good" | "bad" | "neutral";

export function WorkflowWorkspaceReviewPanel({
  review,
}: {
  review: WorkflowWorkspaceReviewViewModel;
}) {
  return (
    <section
      className="surface-band workflow-workspace-review"
      id="workflow-workspace-review"
      aria-labelledby="workflow-workspace-review-title"
    >
      <div className="section-heading">
        <div>
          <p className="eyebrow">Workflow Review Workspace</p>
          <h3 id="workflow-workspace-review-title">Context, scenario, draft, blockers</h3>
        </div>
        <StatusBadge tone={review.canRenderWorkspaceReview ? "neutral" : "bad"}>
          {review.canRenderWorkspaceReview ? "offline advisory" : "blocked"}
        </StatusBadge>
      </div>

      <article className="workflow-workspace-review-hero">
        <div className="workflow-workspace-review-row-main">
          <div>
            <p className="eyebrow">{review.reviewMode}</p>
            <h5>{review.scenarioId}</h5>
          </div>
          <StatusBadge tone="neutral">inspect only</StatusBadge>
        </div>
        <dl className="workflow-run-guard-meta">
          <div>
            <dt>Application</dt>
            <dd>{review.applicationId}</dd>
          </div>
          <div>
            <dt>Workflow definition</dt>
            <dd>{review.workflowDefinitionId}</dd>
          </div>
          <div>
            <dt>Run</dt>
            <dd>{review.runId}</dd>
          </div>
          <div>
            <dt>Draft</dt>
            <dd>{review.draftId}</dd>
          </div>
          <div>
            <dt>Request</dt>
            <dd>{review.requestId}</dd>
          </div>
          <div>
            <dt>Audit</dt>
            <dd>{review.auditRef}</dd>
          </div>
        </dl>
        <p>{review.reviewNarrative}</p>
      </article>

      <div className="workflow-workspace-review-context-grid" aria-label="Workflow review selected context">
        {review.contextCards.map((context) => (
          <WorkflowWorkspaceReviewContextCard key={context.contextId} context={context} />
        ))}
      </div>

      <div className="workflow-workspace-review-stage-grid" aria-label="Workflow review stage order">
        {review.reviewStages.map((stage) => (
          <WorkflowWorkspaceReviewStageCard key={stage.stageId} stage={stage} />
        ))}
      </div>

      <div className="workflow-workspace-review-relation-grid" aria-label="Workflow review relationship map">
        {review.relationMap.map((relation) => (
          <WorkflowWorkspaceReviewRelationCard key={relation.relationId} relation={relation} />
        ))}
      </div>

      <div className="workflow-workspace-review-blocked-grid" aria-label="Workflow review blocked capability rollup">
        {review.blockedCapabilityGroups.map((group) => (
          <WorkflowWorkspaceReviewBlockedCapabilityGroupCard key={group.groupId} group={group} />
        ))}
      </div>

      <div className="workflow-workspace-review-stopline-grid" aria-label="Workflow review stop lines">
        {review.stopLines.map((stopLine) => (
          <WorkflowWorkspaceReviewStopLineCard key={stopLine.stopLineId} stopLine={stopLine} />
        ))}
      </div>
    </section>
  );
}

function WorkflowWorkspaceReviewContextCard({ context }: { context: WorkflowWorkspaceReviewContextCard }) {
  return (
    <article className="workflow-workspace-review-context">
      <div className="workflow-workspace-review-row-main">
        <div>
          <p className="eyebrow">{context.contextId}</p>
          <h5>{context.label}</h5>
        </div>
        <StatusBadge tone={workflowWorkspaceReviewTone(context.status)}>{context.status}</StatusBadge>
      </div>
      <dl className="workflow-run-guard-meta">
        <div>
          <dt>Primary</dt>
          <dd>{context.primaryRef}</dd>
        </div>
        <div>
          <dt>Secondary</dt>
          <dd>{context.secondaryRef}</dd>
        </div>
        <div>
          <dt>Audit</dt>
          <dd>{context.auditRef}</dd>
        </div>
      </dl>
      <p>{context.summary}</p>
    </article>
  );
}

function WorkflowWorkspaceReviewStageCard({ stage }: { stage: WorkflowWorkspaceReviewStage }) {
  return (
    <article className="workflow-workspace-review-stage">
      <div className="workflow-workspace-review-row-main">
        <div>
          <p className="eyebrow">
            {stage.order}. {stage.sourceSurface}
          </p>
          <h5>{stage.label}</h5>
        </div>
        <StatusBadge tone={workflowWorkspaceReviewTone(stage.status)}>{stage.status}</StatusBadge>
      </div>
      <dl className="workflow-run-guard-meta">
        <div>
          <dt>Primary ref</dt>
          <dd>{stage.primaryRef}</dd>
        </div>
        <div>
          <dt>Blocked count</dt>
          <dd>{stage.blockedCount}</dd>
        </div>
        <div>
          <dt>Audit</dt>
          <dd>{stage.auditRef}</dd>
        </div>
      </dl>
      <p>{stage.summary}</p>
      <small>{stage.reviewQuestion}</small>
    </article>
  );
}

function WorkflowWorkspaceReviewRelationCard({ relation }: { relation: WorkflowWorkspaceReviewRelation }) {
  return (
    <article className="workflow-workspace-review-relation">
      <div className="workflow-workspace-review-row-main">
        <div>
          <p className="eyebrow">{relation.relationId}</p>
          <h5>{relation.label}</h5>
        </div>
        <StatusBadge tone={workflowWorkspaceReviewTone(relation.status)}>{relation.status}</StatusBadge>
      </div>
      <dl className="workflow-run-guard-meta">
        <div>
          <dt>Source</dt>
          <dd>{relation.sourceRef}</dd>
        </div>
        <div>
          <dt>Target</dt>
          <dd>{relation.targetRef}</dd>
        </div>
        <div>
          <dt>Audit</dt>
          <dd>{relation.auditRef}</dd>
        </div>
      </dl>
      <p>{relation.summary}</p>
    </article>
  );
}

function WorkflowWorkspaceReviewBlockedCapabilityGroupCard({
  group,
}: {
  group: WorkflowWorkspaceReviewBlockedCapabilityGroup;
}) {
  return (
    <article className="workflow-workspace-review-blocked-group">
      <div className="workflow-workspace-review-row-main">
        <div>
          <p className="eyebrow">{group.sourceSurface}</p>
          <h5>{group.label}</h5>
        </div>
        <StatusBadge tone="bad">{`${group.count} blocked`}</StatusBadge>
      </div>
      <div className="workflow-workspace-review-token-list" aria-label={`${group.label} missing prerequisites`}>
        {group.missingPrerequisites.map((prerequisite) => (
          <code key={prerequisite}>{prerequisite}</code>
        ))}
      </div>
      <div className="workflow-workspace-review-token-list" aria-label={`${group.label} audit refs`}>
        {group.auditRefs.map((auditRef) => (
          <code key={auditRef}>{auditRef}</code>
        ))}
      </div>
      <p>{group.exampleSummary}</p>
    </article>
  );
}

function WorkflowWorkspaceReviewStopLineCard({ stopLine }: { stopLine: WorkflowWorkspaceReviewStopLine }) {
  return (
    <article className="workflow-workspace-review-stopline">
      <div className="workflow-workspace-review-row-main">
        <div>
          <p className="eyebrow">{stopLine.sourceSurface}</p>
          <h5>{stopLine.label}</h5>
        </div>
        <StatusBadge tone="bad">{stopLine.status}</StatusBadge>
      </div>
      <p>{stopLine.summary}</p>
    </article>
  );
}

function StatusBadge({ children, tone }: { children: string; tone: StatusBadgeTone }) {
  return <span className={`status-badge ${tone}`}>{children}</span>;
}

function workflowWorkspaceReviewTone(status: WorkflowWorkspaceReviewStatus): StatusBadgeTone {
  if (status === "blocked" || status === "locked") {
    return "bad";
  }
  if (status === "ready") {
    return "good";
  }
  return "neutral";
}
