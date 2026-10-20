import "../../i18n/workflowReadProjectionResources.ts";
import "../../i18n/workflowWorkspaceProjectionResources.ts";
import { useTranslation } from "react-i18next";
import { workflowProjectionText, workflowProjectionMessage, workflowProjectionStatusLabel } from "./workflowProjectionCopy.ts";
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
  const { t } = useTranslation("workflow");
  return (
    <section
      className="surface-band workflow-workspace-review"
      id="workflow-workspace-review"
      aria-labelledby="workflow-workspace-review-title"
    >
      <div className="section-heading">
        <div>
          <p className="eyebrow">{t($ => $.projection.workflowWorkspaceProjection_ui_Workflow_Review_Workspace)}</p>
          <h3 id="workflow-workspace-review-title">{t($ => $.projection.workflowWorkspaceProjection_ui_Context_scenario_draft_blockers)}</h3>
        </div>
        <StatusBadge tone={review.canRenderWorkspaceReview ? "neutral" : "bad"}>
          {review.canRenderWorkspaceReview ? t($ => $.projection.workflowWorkspaceProjection_ui_offline_advisory) : t($ => $.projection.workflowWorkspaceProjection_ui_blocked)}
        </StatusBadge>
      </div>

      <article className="workflow-workspace-review-hero">
        <div className="workflow-workspace-review-row-main">
          <div>
            <p className="eyebrow">{workflowProjectionStatusLabel(t, review.reviewMode)}</p>
            <h5>{review.scenarioId}</h5>
          </div>
          <StatusBadge tone="neutral">{t($ => $.projection.workflowWorkspaceProjection_ui_inspect_only)}</StatusBadge>
        </div>
        <dl className="workflow-run-guard-meta">
          <div>
            <dt>{t($ => $.projection.workflowWorkspaceProjection_ui_Application)}</dt>
            <dd>{review.applicationId}</dd>
          </div>
          <div>
            <dt>{t($ => $.projection.workspaceReview_selected_definition_label)}</dt>
            <dd>{review.workflowDefinitionId}</dd>
          </div>
          <div>
            <dt>{t($ => $.projection.workflowWorkspaceProjection_ui_Run)}</dt>
            <dd>{review.runId}</dd>
          </div>
          <div>
            <dt>{t($ => $.projection.workflowWorkspaceProjection_ui_Draft)}</dt>
            <dd>{review.draftId}</dd>
          </div>
          <div>
            <dt>{t($ => $.projection.workflowWorkspaceProjection_ui_Request)}</dt>
            <dd>{review.requestId}</dd>
          </div>
          <div>
            <dt>{t($ => $.projection.workflowWorkspaceProjection_ui_Audit)}</dt>
            <dd>{review.auditRef}</dd>
          </div>
        </dl>
        <p>{workflowProjectionText(t, review, "reviewNarrative")}</p>
      </article>

      <div className="workflow-workspace-review-context-grid" aria-label={t($ => $.projection.workflowWorkspaceProjection_ui_Workflow_review_selected_context)}>
        {review.contextCards.map((context) => (
          <WorkflowWorkspaceReviewContextCard key={context.contextId} context={context} />
        ))}
      </div>

      <div className="workflow-workspace-review-stage-grid" aria-label={t($ => $.projection.workflowWorkspaceProjection_ui_Workflow_review_stage_order)}>
        {review.reviewStages.map((stage) => (
          <WorkflowWorkspaceReviewStageCard key={stage.stageId} stage={stage} />
        ))}
      </div>

      <div className="workflow-workspace-review-relation-grid" aria-label={t($ => $.projection.workflowWorkspaceProjection_ui_Workflow_review_relationship_map)}>
        {review.relationMap.map((relation) => (
          <WorkflowWorkspaceReviewRelationCard key={relation.relationId} relation={relation} />
        ))}
      </div>

      <div className="workflow-workspace-review-blocked-grid" aria-label={t($ => $.projection.workflowWorkspaceProjection_ui_Workflow_review_blocked_capability_rollup)}>
        {review.blockedCapabilityGroups.map((group) => (
          <WorkflowWorkspaceReviewBlockedCapabilityGroupCard key={group.groupId} group={group} />
        ))}
      </div>

      <div className="workflow-workspace-review-stopline-grid" aria-label={t($ => $.projection.workflowWorkspaceProjection_ui_Workflow_review_stop_lines)}>
        {review.stopLines.map((stopLine) => (
          <WorkflowWorkspaceReviewStopLineCard key={stopLine.stopLineId} stopLine={stopLine} />
        ))}
      </div>
    </section>
  );
}

function WorkflowWorkspaceReviewContextCard({ context }: { context: WorkflowWorkspaceReviewContextCard }) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-workspace-review-context">
      <div className="workflow-workspace-review-row-main">
        <div>
          <p className="eyebrow">{context.contextId}</p>
          <h5>{workflowProjectionText(t, context, "label")}</h5>
        </div>
        <StatusBadge tone={workflowWorkspaceReviewTone(context.status)}>{workflowProjectionStatusLabel(t, context.status)}</StatusBadge>
      </div>
      <dl className="workflow-run-guard-meta">
        <div>
          <dt>{t($ => $.projection.workflowWorkspaceProjection_ui_Primary)}</dt>
          <dd>{workflowProjectionText(t, context, "primaryRef")}</dd>
        </div>
        <div>
          <dt>{t($ => $.projection.workflowWorkspaceProjection_ui_Secondary)}</dt>
          <dd>{workflowProjectionText(t, context, "secondaryRef")}</dd>
        </div>
        <div>
          <dt>{t($ => $.projection.workflowWorkspaceProjection_ui_Audit)}</dt>
          <dd>{context.auditRef}</dd>
        </div>
      </dl>
      <p>{workflowProjectionText(t, context, "summary")}</p>
    </article>
  );
}

function WorkflowWorkspaceReviewStageCard({ stage }: { stage: WorkflowWorkspaceReviewStage }) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-workspace-review-stage">
      <div className="workflow-workspace-review-row-main">
        <div>
          <p className="eyebrow">
            {stage.order}. {workflowProjectionStatusLabel(t, stage.sourceSurface)}
          </p>
          <h5>{workflowProjectionText(t, stage, "label")}</h5>
        </div>
        <StatusBadge tone={workflowWorkspaceReviewTone(stage.status)}>{workflowProjectionStatusLabel(t, stage.status)}</StatusBadge>
      </div>
      <dl className="workflow-run-guard-meta">
        <div>
          <dt>{t($ => $.projection.workflowWorkspaceProjection_ui_Primary_ref)}</dt>
          <dd>{workflowProjectionText(t, stage, "primaryRef")}</dd>
        </div>
        <div>
          <dt>{t($ => $.projection.workflowWorkspaceProjection_ui_Blocked_count)}</dt>
          <dd>{stage.blockedCount}</dd>
        </div>
        <div>
          <dt>{t($ => $.projection.workflowWorkspaceProjection_ui_Audit)}</dt>
          <dd>{stage.auditRef}</dd>
        </div>
      </dl>
      <p>{workflowProjectionText(t, stage, "summary")}</p>
      <small>{workflowProjectionText(t, stage, "reviewQuestion")}</small>
    </article>
  );
}

function WorkflowWorkspaceReviewRelationCard({ relation }: { relation: WorkflowWorkspaceReviewRelation }) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-workspace-review-relation">
      <div className="workflow-workspace-review-row-main">
        <div>
          <p className="eyebrow">{relation.relationId}</p>
          <h5>{workflowProjectionText(t, relation, "label")}</h5>
        </div>
        <StatusBadge tone={workflowWorkspaceReviewTone(relation.status)}>{workflowProjectionStatusLabel(t, relation.status)}</StatusBadge>
      </div>
      <dl className="workflow-run-guard-meta">
        <div>
          <dt>{t($ => $.projection.workflowWorkspaceProjection_ui_Source)}</dt>
          <dd>{relation.sourceRef}</dd>
        </div>
        <div>
          <dt>{t($ => $.projection.workflowWorkspaceProjection_ui_Target)}</dt>
          <dd>{workflowProjectionText(t, relation, "targetRef")}</dd>
        </div>
        <div>
          <dt>{t($ => $.projection.workflowWorkspaceProjection_ui_Audit)}</dt>
          <dd>{relation.auditRef}</dd>
        </div>
      </dl>
      <p>{workflowProjectionText(t, relation, "summary")}</p>
    </article>
  );
}

function WorkflowWorkspaceReviewBlockedCapabilityGroupCard({
  group,
}: {
  group: WorkflowWorkspaceReviewBlockedCapabilityGroup;
}) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-workspace-review-blocked-group">
      <div className="workflow-workspace-review-row-main">
        <div>
          <p className="eyebrow">{workflowProjectionStatusLabel(t, group.sourceSurface)}</p>
          <h5>{workflowProjectionText(t, group, "label")}</h5>
        </div>
        <StatusBadge tone="bad">{t($ => $.projection.workflowWorkspaceProjection_ui_count_blocked, { count: group.count })}</StatusBadge>
      </div>
      <div className="workflow-workspace-review-token-list" aria-label={t($ => $.projection.workflowWorkspaceProjection_ui_label_missing_prerequisites, { label: workflowProjectionText(t, group, "label") })}>
        {group.missingPrerequisiteMessage ? <p>{workflowProjectionMessage(t, group.missingPrerequisiteMessage)}</p> : group.missingPrerequisites.map(prerequisite => <code key={prerequisite}>{prerequisite}</code>)}
      </div>
      <div className="workflow-workspace-review-token-list" aria-label={t($ => $.projection.workflowWorkspaceProjection_ui_label_audit_refs, { label: workflowProjectionText(t, group, "label") })}>
        {group.auditRefs.map((auditRef) => (
          <code key={auditRef}>{auditRef}</code>
        ))}
      </div>
      <p>{workflowProjectionText(t, group, "exampleSummary")}</p>
    </article>
  );
}

function WorkflowWorkspaceReviewStopLineCard({ stopLine }: { stopLine: WorkflowWorkspaceReviewStopLine }) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-workspace-review-stopline">
      <div className="workflow-workspace-review-row-main">
        <div>
          <p className="eyebrow">{workflowProjectionStatusLabel(t, stopLine.sourceSurface)}</p>
          <h5>{workflowProjectionText(t, stopLine, "label")}</h5>
        </div>
        <StatusBadge tone="bad">{workflowProjectionStatusLabel(t, stopLine.status)}</StatusBadge>
      </div>
      <p>{workflowProjectionText(t, stopLine, "summary")}</p>
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
