import type {
  WorkflowUserWorkspaceHomeApplication,
  WorkflowUserWorkspaceHomeMetric,
  WorkflowUserWorkspaceHomeReadiness,
  WorkflowUserWorkspaceHomeRouteEvidence,
  WorkflowUserWorkspaceHomeRun,
  WorkflowUserWorkspaceHomeStatus,
  WorkflowUserWorkspaceHomeViewModel,
} from "./workflowUserWorkspaceHome";
import type {
  WorkflowWorkspaceReviewStage,
  WorkflowWorkspaceReviewStopLine,
} from "./workflowWorkspaceReview";

type StatusBadgeTone = "good" | "bad" | "neutral";

export function WorkflowUserWorkspaceHomePanel({
  home,
  createdDraftCountsByWorkflowDefinition,
  onCreateDraftForWorkflowDefinition,
}: {
  home: WorkflowUserWorkspaceHomeViewModel;
  createdDraftCountsByWorkflowDefinition: Record<string, number>;
  onCreateDraftForWorkflowDefinition: (workflowDefinitionId: string) => void;
}) {
  const primaryReadiness = home.readinessRollup.slice(0, 6);
  const primaryRouteEvidence = home.routeEvidence.slice(0, 4);
  const primaryStopLines = home.stopLines.slice(0, 4);

  return (
    <section
      className="surface-band workflow-user-workspace-home"
      id="workflow-user-workspace-home"
      aria-labelledby="workflow-user-workspace-home-title"
    >
      <div className="section-heading">
        <div>
          <p className="eyebrow">User Workspace Home</p>
          <h3 id="workflow-user-workspace-home-title">Workspace status and review handoff</h3>
        </div>
        <StatusBadge tone={home.canRenderUserWorkspaceHome ? "neutral" : "bad"}>
          {home.canRenderUserWorkspaceHome ? "offline advisory" : "blocked"}
        </StatusBadge>
      </div>

      <article className="workflow-user-workspace-home-hero">
        <div>
          <p className="eyebrow">{home.homeMode}</p>
          <h4>{home.tenantRef}</h4>
          <p>{home.homeNarrative}</p>
        </div>
        <dl className="workflow-user-workspace-home-meta">
          <div>
            <dt>Application</dt>
            <dd>{home.applicationId}</dd>
          </div>
          <div>
            <dt>Workflow</dt>
            <dd>{home.workflowDefinitionId}</dd>
          </div>
          <div>
            <dt>Run</dt>
            <dd>{home.runId}</dd>
          </div>
          <div>
            <dt>Draft</dt>
            <dd>{home.draftId}</dd>
          </div>
          <div>
            <dt>Scenario</dt>
            <dd>{home.scenarioId}</dd>
          </div>
          <div>
            <dt>Audit</dt>
            <dd>{home.auditRef}</dd>
          </div>
        </dl>
      </article>

      <div className="workflow-user-workspace-home-metric-grid" aria-label="User workspace home metrics">
        {home.metrics.map((metric) => (
          <WorkflowUserWorkspaceHomeMetricCard key={metric.metricId} metric={metric} />
        ))}
      </div>

      <div className="workflow-user-workspace-home-layout">
        <div className="workflow-user-workspace-home-column">
          <div className="workflow-user-workspace-home-subheading">
            <p className="eyebrow">Application Portfolio</p>
            <h4>Workspace apps</h4>
          </div>
          <div className="workflow-user-workspace-home-application-grid" aria-label="Workspace application portfolio">
            {home.applicationPortfolio.map((application) => (
              <WorkflowUserWorkspaceHomeApplicationCard
                key={application.applicationRef}
                application={application}
                createdDraftCount={
                  createdDraftCountsByWorkflowDefinition[application.workflowDefinitionId] ?? 0
                }
                onCreateDraftForWorkflowDefinition={onCreateDraftForWorkflowDefinition}
              />
            ))}
          </div>
        </div>

        <div className="workflow-user-workspace-home-column">
          <div className="workflow-user-workspace-home-subheading">
            <p className="eyebrow">Review Path</p>
            <h4>Current inspection order</h4>
          </div>
          <div className="workflow-user-workspace-home-stage-grid" aria-label="Current workflow review stages">
            {home.currentReviewStages.map((stage) => (
              <WorkflowUserWorkspaceHomeStageCard key={stage.stageId} stage={stage} />
            ))}
          </div>
        </div>
      </div>

      <div className="workflow-user-workspace-home-section">
        <div className="workflow-user-workspace-home-subheading">
          <p className="eyebrow">Readiness Priorities</p>
          <h4>Primary blocked and offline signals</h4>
        </div>
        <div className="workflow-user-workspace-home-readiness-grid" aria-label="Workspace readiness rollup">
          {primaryReadiness.map((readiness) => (
            <WorkflowUserWorkspaceHomeReadinessCard key={readiness.readinessId} readiness={readiness} />
          ))}
        </div>
      </div>

      <div className="workflow-user-workspace-home-section">
        <div className="workflow-user-workspace-home-subheading">
          <p className="eyebrow">Run Evidence</p>
          <h4>Recent workspace runs</h4>
        </div>
        <div className="workflow-user-workspace-home-run-grid" aria-label="Recent workspace runs">
          {home.recentRuns.map((run) => (
            <WorkflowUserWorkspaceHomeRunCard key={run.runId} run={run} />
          ))}
        </div>
      </div>

      <div className="workflow-user-workspace-home-section">
        <div className="workflow-user-workspace-home-subheading">
          <p className="eyebrow">Route Evidence</p>
          <h4>Read-side source routes</h4>
        </div>
        <div className="workflow-user-workspace-home-route-grid" aria-label="Workspace home route evidence">
          {primaryRouteEvidence.map((route) => (
            <WorkflowUserWorkspaceHomeRouteCard key={route.evidenceId} route={route} />
          ))}
        </div>
      </div>

      <div className="workflow-user-workspace-home-section">
        <div className="workflow-user-workspace-home-subheading">
          <p className="eyebrow">Stop Lines</p>
          <h4>Locked capabilities</h4>
        </div>
        <div className="workflow-user-workspace-home-stopline-grid" aria-label="Workspace home stop lines">
          {primaryStopLines.map((stopLine) => (
            <WorkflowUserWorkspaceHomeStopLineCard key={stopLine.stopLineId} stopLine={stopLine} />
          ))}
        </div>
      </div>
    </section>
  );
}

function WorkflowUserWorkspaceHomeMetricCard({ metric }: { metric: WorkflowUserWorkspaceHomeMetric }) {
  return (
    <article className="workflow-user-workspace-home-card">
      <span>{metric.label}</span>
      <strong>{metric.value}</strong>
      <StatusBadge tone={workflowUserWorkspaceHomeTone(metric.status)}>{metric.status}</StatusBadge>
      <p>{metric.summary}</p>
    </article>
  );
}

function WorkflowUserWorkspaceHomeApplicationCard({
  application,
  createdDraftCount,
  onCreateDraftForWorkflowDefinition,
}: {
  application: WorkflowUserWorkspaceHomeApplication;
  createdDraftCount: number;
  onCreateDraftForWorkflowDefinition: (workflowDefinitionId: string) => void;
}) {
  return (
    <article className="workflow-user-workspace-home-card">
      <div className="workflow-user-workspace-home-row-main">
        <div>
          <p className="eyebrow">{application.applicationKind}</p>
          <h5>{application.displayName}</h5>
        </div>
        <StatusBadge tone={workflowUserWorkspaceHomeTone(application.status)}>
          {application.selected ? "selected" : application.status}
        </StatusBadge>
      </div>
      <dl className="workflow-user-workspace-home-meta">
        <div>
          <dt>Application</dt>
          <dd>{application.applicationRef}</dd>
        </div>
        <div>
          <dt>Workflow</dt>
          <dd>{application.workflowDefinitionId}</dd>
        </div>
        <div>
          <dt>Run status</dt>
          <dd>{application.latestRunStatus}</dd>
        </div>
        <div>
          <dt>Audit</dt>
          <dd>{application.auditRef}</dd>
        </div>
      </dl>
      <p>{application.summary}</p>
      <div className="workflow-user-workspace-home-actions">
        <span>{createdDraftCount} local drafts</span>
        <button
          type="button"
          onClick={() => onCreateDraftForWorkflowDefinition(application.workflowDefinitionId)}
        >
          Create draft
        </button>
      </div>
    </article>
  );
}

function WorkflowUserWorkspaceHomeStageCard({ stage }: { stage: WorkflowWorkspaceReviewStage }) {
  return (
    <article className="workflow-user-workspace-home-card">
      <div className="workflow-user-workspace-home-row-main">
        <div>
          <p className="eyebrow">
            {stage.order}. {stage.sourceSurface}
          </p>
          <h5>{stage.label}</h5>
        </div>
        <StatusBadge tone={workflowUserWorkspaceHomeTone(stage.status)}>{stage.status}</StatusBadge>
      </div>
      <dl className="workflow-user-workspace-home-meta">
        <div>
          <dt>Primary</dt>
          <dd>{stage.primaryRef}</dd>
        </div>
        <div>
          <dt>Blocked</dt>
          <dd>{stage.blockedCount}</dd>
        </div>
        <div>
          <dt>Audit</dt>
          <dd>{stage.auditRef}</dd>
        </div>
      </dl>
      <p>{stage.summary}</p>
    </article>
  );
}

function WorkflowUserWorkspaceHomeReadinessCard({
  readiness,
}: {
  readiness: WorkflowUserWorkspaceHomeReadiness;
}) {
  return (
    <article className="workflow-user-workspace-home-card">
      <div className="workflow-user-workspace-home-row-main">
        <div>
          <p className="eyebrow">{readiness.sourceSurface}</p>
          <h5>{readiness.label}</h5>
        </div>
        <StatusBadge tone={workflowUserWorkspaceHomeTone(readiness.status)}>{readiness.status}</StatusBadge>
      </div>
      <dl className="workflow-user-workspace-home-meta">
        <div>
          <dt>Value</dt>
          <dd>{readiness.value}</dd>
        </div>
        <div>
          <dt>Audit</dt>
          <dd>{readiness.auditRef}</dd>
        </div>
      </dl>
      <p>{readiness.summary}</p>
    </article>
  );
}

function WorkflowUserWorkspaceHomeRunCard({ run }: { run: WorkflowUserWorkspaceHomeRun }) {
  return (
    <article className="workflow-user-workspace-home-card">
      <div className="workflow-user-workspace-home-row-main">
        <div>
          <p className="eyebrow">{run.applicationRef}</p>
          <h5>{run.runId}</h5>
        </div>
        <StatusBadge tone={workflowUserWorkspaceHomeTone(run.status)}>
          {run.selected ? "selected" : run.status}
        </StatusBadge>
      </div>
      <dl className="workflow-user-workspace-home-meta">
        <div>
          <dt>Workflow</dt>
          <dd>{run.workflowDefinitionId}</dd>
        </div>
        <div>
          <dt>Failure</dt>
          <dd>{run.failureCode}</dd>
        </div>
        <div>
          <dt>Cost</dt>
          <dd>{run.cost}</dd>
        </div>
        <div>
          <dt>Trace</dt>
          <dd>{run.traceId}</dd>
        </div>
      </dl>
      <p>{run.summary}</p>
    </article>
  );
}

function WorkflowUserWorkspaceHomeRouteCard({ route }: { route: WorkflowUserWorkspaceHomeRouteEvidence }) {
  return (
    <article className="workflow-user-workspace-home-card">
      <div className="workflow-user-workspace-home-row-main">
        <div>
          <p className="eyebrow">{route.evidenceId}</p>
          <h5>{route.label}</h5>
        </div>
        <StatusBadge tone={workflowUserWorkspaceHomeTone(route.status)}>{route.status}</StatusBadge>
      </div>
      <dl className="workflow-user-workspace-home-meta">
        <div>
          <dt>Route</dt>
          <dd>{route.routeId}</dd>
        </div>
        <div>
          <dt>Request</dt>
          <dd>{route.requestId}</dd>
        </div>
        <div>
          <dt>Audit</dt>
          <dd>{route.auditRef}</dd>
        </div>
      </dl>
      <p>{route.summary}</p>
    </article>
  );
}

function WorkflowUserWorkspaceHomeStopLineCard({ stopLine }: { stopLine: WorkflowWorkspaceReviewStopLine }) {
  return (
    <article className="workflow-user-workspace-home-card">
      <div className="workflow-user-workspace-home-row-main">
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

function workflowUserWorkspaceHomeTone(status: WorkflowUserWorkspaceHomeStatus): StatusBadgeTone {
  if (status === "blocked" || status === "locked") {
    return "bad";
  }
  if (status === "ready") {
    return "good";
  }
  return "neutral";
}
