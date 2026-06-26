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
          <p className="eyebrow">Active Draft Review Record</p>
          <h4>Validation, plan, readiness</h4>
        </div>
        <div className="workflow-user-workspace-home-route-grid" aria-label="Workflow review handoff active draft record">
          {handoff.activeDraftReviewRecord.sections.map((section) => (
            <WorkflowReviewHandoffActiveDraftSectionCard key={section.sectionId} section={section} />
          ))}
        </div>
      </div>

      <div className="workflow-user-workspace-home-section">
        <div className="workflow-user-workspace-home-subheading">
          <p className="eyebrow">Node Designer Review Handoff</p>
          <h4>Canvas overlay, inspector, mapping</h4>
        </div>
        <div className="workflow-user-workspace-home-route-grid" aria-label="Workflow node designer review handoff">
          {handoff.nodeDesignerReviewRecord.sections.map((section) => (
            <WorkflowReviewHandoffNodeDesignerSectionCard key={section.sectionId} section={section} />
          ))}
        </div>
        <div className="workflow-review-handoff-graph-summary" aria-label="Workflow node designer graph review summary">
          {graphFindingGroups.map((group) => (
            <article key={group.targetKind}>
              <span>{group.label}</span>
              <strong>{group.count}</strong>
            </article>
          ))}
        </div>
        <div className="workflow-review-handoff-graph-groups" aria-label="Workflow node designer graph review findings">
          {graphFindingGroups.map((group) => (
            <section key={group.targetKind} className="workflow-review-handoff-graph-group">
              <div className="workflow-review-handoff-graph-group-heading">
                <div>
                  <p className="eyebrow">{group.targetKind}</p>
                  <h5>{group.label}</h5>
                </div>
                <StatusBadge tone={group.count > 0 ? "neutral" : "good"}>{`${group.count} findings`}</StatusBadge>
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

function WorkflowReviewHandoffActiveDraftSectionCard({
  section,
}: {
  section: WorkflowReviewHandoffActiveDraftReviewSection;
}) {
  return (
    <article className="workflow-user-workspace-home-card">
      <div className="workflow-user-workspace-home-row-main">
        <div>
          <p className="eyebrow">{section.sourceSurface}</p>
          <h5>{section.label}</h5>
        </div>
        <StatusBadge tone={workflowReviewHandoffTone(section.status)}>{section.status}</StatusBadge>
      </div>
      <dl className="workflow-user-workspace-home-meta">
        <div>
          <dt>Primary ref</dt>
          <dd>{section.primaryRef}</dd>
        </div>
        <div>
          <dt>Blockers</dt>
          <dd>{section.blockerCount}</dd>
        </div>
        <div>
          <dt>Request</dt>
          <dd>{section.requestId}</dd>
        </div>
        <div>
          <dt>Audit</dt>
          <dd>{section.auditRef}</dd>
        </div>
      </dl>
      <div className="workflow-workspace-review-token-list" aria-label={`${section.label} evidence refs`}>
        {section.evidenceRefs.map((evidenceRef) => (
          <code key={evidenceRef}>{evidenceRef}</code>
        ))}
      </div>
      <p>{section.summary}</p>
      <p>{section.reviewerQuestion}</p>
    </article>
  );
}

function WorkflowReviewHandoffNodeDesignerSectionCard({
  section,
}: {
  section: WorkflowReviewHandoffNodeDesignerReviewSection;
}) {
  return (
    <article className="workflow-user-workspace-home-card">
      <div className="workflow-user-workspace-home-row-main">
        <div>
          <p className="eyebrow">{section.sourceSurface}</p>
          <h5>{section.label}</h5>
        </div>
        <StatusBadge tone={workflowReviewHandoffTone(section.status)}>{section.status}</StatusBadge>
      </div>
      <dl className="workflow-user-workspace-home-meta">
        <div>
          <dt>Primary ref</dt>
          <dd>{section.primaryRef}</dd>
        </div>
        <div>
          <dt>Items</dt>
          <dd>{section.itemCount}</dd>
        </div>
        <div>
          <dt>Request</dt>
          <dd>{section.requestId}</dd>
        </div>
        <div>
          <dt>Audit</dt>
          <dd>{section.auditRef}</dd>
        </div>
      </dl>
      <div className="workflow-workspace-review-token-list" aria-label={`${section.label} evidence refs`}>
        {section.evidenceRefs.map((evidenceRef) => (
          <code key={evidenceRef}>{evidenceRef}</code>
        ))}
      </div>
      <p>{section.summary}</p>
      <p>{section.reviewerQuestion}</p>
    </article>
  );
}

function WorkflowReviewHandoffNodeDesignerGraphFindingCard({
  finding,
}: {
  finding: WorkflowReviewHandoffNodeDesignerGraphFinding;
}) {
  return (
    <article className="workflow-user-workspace-home-card">
      <div className="workflow-user-workspace-home-row-main">
        <div>
          <p className="eyebrow">{finding.targetKind}</p>
          <h5>{finding.label}</h5>
        </div>
        <StatusBadge tone={workflowReviewHandoffTone(finding.status)}>{finding.status}</StatusBadge>
      </div>
      <dl className="workflow-user-workspace-home-meta">
        <div>
          <dt>Check</dt>
          <dd>{finding.sourceCheckId}</dd>
        </div>
        <div>
          <dt>Severity</dt>
          <dd>{finding.severity}</dd>
        </div>
        <div>
          <dt>Target</dt>
          <dd>{finding.targetSummary}</dd>
        </div>
      </dl>
      <div className="workflow-workspace-review-token-list" aria-label={`${finding.label} target refs`}>
        {finding.targetRefs.map((targetRef) => (
          <code key={targetRef}>{targetRef}</code>
        ))}
      </div>
      <p>{finding.summary}</p>
      <p>{finding.reviewerQuestion}</p>
    </article>
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
