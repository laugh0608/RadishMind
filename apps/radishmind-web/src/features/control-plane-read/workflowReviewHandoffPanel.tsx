import type {
  WorkflowReviewHandoffBoundaryLock,
  WorkflowReviewHandoffDecisionBlocker,
  WorkflowReviewHandoffEvidence,
  WorkflowReviewHandoffFinding,
  WorkflowReviewHandoffRecipient,
  WorkflowReviewHandoffStatus,
  WorkflowReviewHandoffViewModel,
} from "./workflowReviewHandoff";

type StatusBadgeTone = "good" | "bad" | "neutral";

export function WorkflowReviewHandoffPanel({
  handoff,
}: {
  handoff: WorkflowReviewHandoffViewModel;
}) {
  const primaryEvidence = handoff.evidenceChecklist.slice(0, 8);
  const primaryBoundaries = handoff.boundaryLocks.slice(0, 8);

  return (
    <section
      className="surface-band workflow-review-handoff workflow-user-workspace-home"
      id="workflow-review-handoff"
      aria-labelledby="workflow-review-handoff-title"
    >
      <div className="section-heading">
        <div>
          <p className="eyebrow">Workflow Review Handoff</p>
          <h3 id="workflow-review-handoff-title">Advisory package for human review</h3>
        </div>
        <StatusBadge tone={handoff.canRenderReviewHandoff ? "neutral" : "bad"}>
          {handoff.canRenderReviewHandoff ? "offline advisory" : "blocked"}
        </StatusBadge>
      </div>

      <article className="workflow-user-workspace-home-hero">
        <div>
          <p className="eyebrow">{handoff.handoffMode}</p>
          <h4>{handoff.handoffPackageId}</h4>
          <p>{handoff.handoffNarrative}</p>
        </div>
        <dl className="workflow-user-workspace-home-meta">
          <div>
            <dt>Application</dt>
            <dd>{handoff.applicationId}</dd>
          </div>
          <div>
            <dt>Workflow</dt>
            <dd>{handoff.workflowDefinitionId}</dd>
          </div>
          <div>
            <dt>Run</dt>
            <dd>{handoff.runId}</dd>
          </div>
          <div>
            <dt>Draft</dt>
            <dd>{handoff.draftId}</dd>
          </div>
          <div>
            <dt>Scenario</dt>
            <dd>{handoff.scenarioId}</dd>
          </div>
          <div>
            <dt>Audit</dt>
            <dd>{handoff.auditRef}</dd>
          </div>
        </dl>
      </article>

      <div className="workflow-user-workspace-home-section">
        <div className="workflow-user-workspace-home-subheading">
          <p className="eyebrow">Review Recipients</p>
          <h4>Who needs the handoff</h4>
        </div>
        <div className="workflow-user-workspace-home-application-grid" aria-label="Workflow review handoff recipients">
          {handoff.recipients.map((recipient) => (
            <WorkflowReviewHandoffRecipientCard key={recipient.recipientId} recipient={recipient} />
          ))}
        </div>
      </div>

      <div className="workflow-user-workspace-home-section">
        <div className="workflow-user-workspace-home-subheading">
          <p className="eyebrow">Key Findings</p>
          <h4>What the reviewer should read first</h4>
        </div>
        <div className="workflow-user-workspace-home-readiness-grid" aria-label="Workflow review handoff findings">
          {handoff.keyFindings.map((finding) => (
            <WorkflowReviewHandoffFindingCard key={finding.findingId} finding={finding} />
          ))}
        </div>
      </div>

      <div className="workflow-user-workspace-home-section">
        <div className="workflow-user-workspace-home-subheading">
          <p className="eyebrow">Evidence Checklist</p>
          <h4>Read-side route and page evidence</h4>
        </div>
        <div className="workflow-user-workspace-home-route-grid" aria-label="Workflow review handoff evidence">
          {primaryEvidence.map((evidence) => (
            <WorkflowReviewHandoffEvidenceCard key={evidence.evidenceId} evidence={evidence} />
          ))}
        </div>
      </div>

      <div className="workflow-user-workspace-home-section">
        <div className="workflow-user-workspace-home-subheading">
          <p className="eyebrow">Decision Blockers</p>
          <h4>Why this remains advisory-only</h4>
        </div>
        <div className="workflow-user-workspace-home-readiness-grid" aria-label="Workflow review handoff blockers">
          {handoff.decisionBlockers.map((blocker) => (
            <WorkflowReviewHandoffDecisionBlockerCard key={blocker.blockerId} blocker={blocker} />
          ))}
        </div>
      </div>

      <div className="workflow-user-workspace-home-section">
        <div className="workflow-user-workspace-home-subheading">
          <p className="eyebrow">Boundary Locks</p>
          <h4>Capabilities that stay closed</h4>
        </div>
        <div className="workflow-user-workspace-home-stopline-grid" aria-label="Workflow review handoff boundaries">
          {primaryBoundaries.map((boundary) => (
            <WorkflowReviewHandoffBoundaryLockCard key={boundary.boundaryId} boundary={boundary} />
          ))}
        </div>
      </div>
    </section>
  );
}

function WorkflowReviewHandoffRecipientCard({
  recipient,
}: {
  recipient: WorkflowReviewHandoffRecipient;
}) {
  return (
    <article className="workflow-user-workspace-home-card">
      <div className="workflow-user-workspace-home-row-main">
        <div>
          <p className="eyebrow">{recipient.role}</p>
          <h5>{recipient.label}</h5>
        </div>
        <StatusBadge tone={workflowReviewHandoffTone(recipient.status)}>{recipient.status}</StatusBadge>
      </div>
      <dl className="workflow-user-workspace-home-meta">
        <div>
          <dt>Recipient</dt>
          <dd>{recipient.recipientId}</dd>
        </div>
        <div>
          <dt>Evidence</dt>
          <dd>{recipient.evidenceRefs.join(", ")}</dd>
        </div>
      </dl>
      <p>{recipient.handoffNeed}</p>
    </article>
  );
}

function WorkflowReviewHandoffFindingCard({
  finding,
}: {
  finding: WorkflowReviewHandoffFinding;
}) {
  return (
    <article className="workflow-user-workspace-home-card">
      <div className="workflow-user-workspace-home-row-main">
        <div>
          <p className="eyebrow">{finding.sourceSurface}</p>
          <h5>{finding.label}</h5>
        </div>
        <StatusBadge tone={workflowReviewHandoffTone(finding.status)}>{finding.status}</StatusBadge>
      </div>
      <dl className="workflow-user-workspace-home-meta">
        <div>
          <dt>Evidence</dt>
          <dd>{finding.evidenceRef}</dd>
        </div>
      </dl>
      <p>{finding.summary}</p>
      <p>{finding.humanReviewQuestion}</p>
    </article>
  );
}

function WorkflowReviewHandoffEvidenceCard({
  evidence,
}: {
  evidence: WorkflowReviewHandoffEvidence;
}) {
  return (
    <article className="workflow-user-workspace-home-card">
      <div className="workflow-user-workspace-home-row-main">
        <div>
          <p className="eyebrow">{evidence.sourceSurface}</p>
          <h5>{evidence.label}</h5>
        </div>
        <StatusBadge tone={workflowReviewHandoffTone(evidence.status)}>{evidence.status}</StatusBadge>
      </div>
      <dl className="workflow-user-workspace-home-meta">
        <div>
          <dt>Route/page</dt>
          <dd>{evidence.routeOrPageId}</dd>
        </div>
        <div>
          <dt>Request</dt>
          <dd>{evidence.requestId}</dd>
        </div>
        <div>
          <dt>Audit</dt>
          <dd>{evidence.auditRef}</dd>
        </div>
      </dl>
      <p>{evidence.summary}</p>
    </article>
  );
}

function WorkflowReviewHandoffDecisionBlockerCard({
  blocker,
}: {
  blocker: WorkflowReviewHandoffDecisionBlocker;
}) {
  return (
    <article className="workflow-user-workspace-home-card">
      <div className="workflow-user-workspace-home-row-main">
        <div>
          <p className="eyebrow">{blocker.sourceSurface}</p>
          <h5>{blocker.label}</h5>
        </div>
        <StatusBadge tone="bad">{blocker.status}</StatusBadge>
      </div>
      <dl className="workflow-user-workspace-home-meta">
        <div>
          <dt>Missing</dt>
          <dd>{blocker.missingPrerequisite}</dd>
        </div>
      </dl>
      <div className="workflow-workspace-review-token-list" aria-label={`${blocker.label} audit refs`}>
        {blocker.auditRefs.map((auditRef) => (
          <code key={auditRef}>{auditRef}</code>
        ))}
      </div>
      <p>{blocker.summary}</p>
    </article>
  );
}

function WorkflowReviewHandoffBoundaryLockCard({
  boundary,
}: {
  boundary: WorkflowReviewHandoffBoundaryLock;
}) {
  return (
    <article className="workflow-user-workspace-home-card">
      <div className="workflow-user-workspace-home-row-main">
        <div>
          <p className="eyebrow">{boundary.boundaryId}</p>
          <h5>{boundary.label}</h5>
        </div>
        <StatusBadge tone="bad">{boundary.status}</StatusBadge>
      </div>
      <p>{boundary.summary}</p>
    </article>
  );
}

function StatusBadge({ children, tone }: { children: string; tone: StatusBadgeTone }) {
  return <span className={`status-badge ${tone}`}>{children}</span>;
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
