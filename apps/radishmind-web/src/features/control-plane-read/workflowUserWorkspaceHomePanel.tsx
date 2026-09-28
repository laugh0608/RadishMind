import { formatDisplayDate } from "../../i18n/formatters.ts";
import "../../i18n/workflowLibraryResources.ts";
import { workflowDraftStatusLabel, workflowDraftLibraryMessage, workflowDraftLifecycleMessage } from "./workflowDraftMessages.ts";
import { useTranslation } from "react-i18next";
import "../../i18n/workflowDraftResources.ts";
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
  WorkflowSavedDraftListState,
  WorkflowSavedDraftLibraryFilters,
  WorkflowSavedDraftLifecycleOperationState,
  WorkflowSavedDraftLifecycleState,
  WorkflowSavedDraftSummary,
} from "./savedWorkflowDraftConsumer";
import {
  workflowSavedDraftLibraryActions,
} from "./savedWorkflowDraftLibraryPresentation.ts";
import type {
  WorkflowWorkspaceReviewStage,
  WorkflowWorkspaceReviewStopLine,
} from "./workflowWorkspaceReview";

type StatusBadgeTone = "good" | "bad" | "neutral";

export function WorkflowUserWorkspaceHomePanel({
  home,
  createdDraftCountsByWorkflowDefinition,
  savedDraftListState,
  onCreateDraftForWorkflowDefinition,
  libraryLifecycle,
  libraryFilters,
  lifecycleOperation,
  onLibraryLifecycleChange,
  onLibraryFiltersChange,
  onRefreshSavedDrafts,
  onLoadMoreSavedDrafts,
  onOpenSavedDraft,
  onArchiveSavedDraft,
  onUnarchiveSavedDraft,
}: {
  home: WorkflowUserWorkspaceHomeViewModel;
  createdDraftCountsByWorkflowDefinition: Record<string, number>;
  savedDraftListState: WorkflowSavedDraftListState;
  libraryLifecycle: WorkflowSavedDraftLifecycleState;
  libraryFilters: WorkflowSavedDraftLibraryFilters;
  lifecycleOperation: WorkflowSavedDraftLifecycleOperationState;
  onCreateDraftForWorkflowDefinition: (workflowDefinitionId: string) => void;
  onLibraryLifecycleChange: (lifecycleState: WorkflowSavedDraftLifecycleState) => void;
  onLibraryFiltersChange: (filters: WorkflowSavedDraftLibraryFilters) => void;
  onRefreshSavedDrafts: () => void;
  onLoadMoreSavedDrafts: () => void;
  onOpenSavedDraft: (summary: WorkflowSavedDraftSummary) => void;
  onArchiveSavedDraft: (summary: WorkflowSavedDraftSummary) => void;
  onUnarchiveSavedDraft: (summary: WorkflowSavedDraftSummary) => void;
}) {
  const { t } = useTranslation("workflow");
  const [draftFilters, setDraftFilters] = useState(libraryFilters);
  const [archiveConfirmDraftId, setArchiveConfirmDraftId] = useState("");
  const primaryReadiness = home.readinessRollup.slice(0, 6);
  const primaryRouteEvidence = home.routeEvidence.slice(0, 4);
  const primaryStopLines = home.stopLines.slice(0, 4);

  useEffect(() => {
    setDraftFilters(libraryFilters);
    setArchiveConfirmDraftId("");
  }, [libraryFilters, savedDraftListState.applicationRef, libraryLifecycle]);

  return (
    <section
      className="surface-band workflow-user-workspace-home"
      id="workflow-user-workspace-home"
      aria-labelledby="workflow-user-workspace-home-title"
    >
      <div className="section-heading">
        <div>
          <p className="eyebrow">{t($ => $.library.home)}</p>
          <h3 id="workflow-user-workspace-home-title">{t($ => $.library.homeTitle)}</h3>
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
            <dt>{t($ => $.draft.application)}</dt>
            <dd>{home.applicationId}</dd>
          </div>
          <div>
            <dt>{t($ => $.library.workflow)}</dt>
            <dd>{home.workflowDefinitionId}</dd>
          </div>
          <div>
            <dt>{t($ => $.library.run)}</dt>
            <dd>{home.runId}</dd>
          </div>
          <div>
            <dt>{t($ => $.draft.draft)}</dt>
            <dd>{home.draftId}</dd>
          </div>
          <div>
            <dt>{t($ => $.library.scenario)}</dt>
            <dd>{home.scenarioId}</dd>
          </div>
          <div>
            <dt>{t($ => $.draft.audit)}</dt>
            <dd>{home.auditRef}</dd>
          </div>
        </dl>
      </article>

      <div className="workflow-user-workspace-home-metric-grid" aria-label={t($ => $.library.homeMetrics)}>
        {home.metrics.map((metric) => (
          <WorkflowUserWorkspaceHomeMetricCard key={metric.metricId} metric={metric} />
        ))}
      </div>

      <div className="workflow-user-workspace-home-layout">
        <div className="workflow-user-workspace-home-column">
          <div className="workflow-user-workspace-home-subheading">
            <p className="eyebrow">{t($ => $.library.portfolio)}</p>
            <h4>{t($ => $.library.workspaceApps)}</h4>
          </div>
          <div className="workflow-user-workspace-home-application-grid" aria-label={t($ => $.library.portfolioRegion)}>
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
            <p className="eyebrow">{t($ => $.library.reviewPath)}</p>
            <h4>{t($ => $.library.inspectionOrder)}</h4>
          </div>
          <div className="workflow-user-workspace-home-stage-grid" aria-label={t($ => $.library.reviewStages)}>
            {home.currentReviewStages.map((stage) => (
              <WorkflowUserWorkspaceHomeStageCard key={stage.stageId} stage={stage} />
            ))}
          </div>
        </div>
      </div>

      <div className="workflow-user-workspace-home-section saved-draft-list" aria-label={t($ => $.library.savedList)}>
        <div className="workflow-user-workspace-home-subheading saved-draft-list-heading">
          <div>
            <p className="eyebrow">{t($ => $.library.savedDrafts)}</p>
            <h4>{t($ => $.library.library)}</h4>
          </div>
          <button
            type="button"
            disabled={savedDraftListState.mode !== "dev_saved_draft_http" || savedDraftListState.status === "loading"}
            onClick={onRefreshSavedDrafts}
          >{t($ => $.library.refresh)}</button>
        </div>
        <div
          className="saved-draft-library-tabs"
          role="tablist"
          aria-label={t($ => $.library.lifecycleViews)}
        >
          {(["active", "archived"] as const).map((lifecycleState) => (
            <button
              key={lifecycleState}
              type="button"
              role="tab"
              aria-selected={libraryLifecycle === lifecycleState}
              className={libraryLifecycle === lifecycleState ? "selected" : ""}
              disabled={savedDraftListState.status === "loading"}
              onClick={() => onLibraryLifecycleChange(lifecycleState)}
            >
              {lifecycleState === "active" ? t($ => $.library.activeDrafts) : t($ => $.library.archivedDrafts)}
            </button>
          ))}
        </div>
        <div className="saved-draft-library-filters" aria-label={t($ => $.library.libraryFilters)}>
          <label>
            <span>{t($ => $.library.namePrefix)}</span>
            <input
              value={draftFilters.namePrefix}
              maxLength={80}
              onChange={(event) => setDraftFilters({
                ...draftFilters,
                namePrefix: event.currentTarget.value,
              })}
            />
          </label>
          <label>
            <span>{t($ => $.draft.validation)}</span>
            <select
              value={draftFilters.validationState}
              onChange={(event) => setDraftFilters({
                ...draftFilters,
                validationState: event.currentTarget.value as WorkflowSavedDraftLibraryFilters["validationState"],
              })}
            >
              <option value="">{t($ => $.library.all)}</option>
              <option value="valid_for_review">{t($ => $.library.validForReview)}</option>
              <option value="invalid_draft">{t($ => $.library.invalidDraft)}</option>
              <option value="blocked_capability">{t($ => $.library.blockedCapability)}</option>
              <option value="schema_unsupported">{t($ => $.library.schemaUnsupported)}</option>
            </select>
          </label>
          <label>
            <span>{t($ => $.library.provenance)}</span>
            <select
              value={draftFilters.provenanceKind}
              onChange={(event) => setDraftFilters({
                ...draftFilters,
                provenanceKind: event.currentTarget.value as WorkflowSavedDraftLibraryFilters["provenanceKind"],
              })}
            >
              <option value="">{t($ => $.library.all)}</option>
              <option value="unversioned">{t($ => $.library.unversioned)}</option>
              <option value="workflow_definition">{t($ => $.library.workflowDefinition)}</option>
              <option value="saved_draft_derivation">{t($ => $.library.savedDerivation)}</option>
              <option value="workspace_template_derivation">{t($ => $.library.templateDerivation)}</option>
            </select>
          </label>
          <div className="saved-draft-library-filter-actions">
            <button
              type="button"
              disabled={savedDraftListState.status === "loading"}
              onClick={() => onLibraryFiltersChange(draftFilters)}
            >{t($ => $.library.applyFilters)}</button>
            <button
              type="button"
              disabled={savedDraftListState.status === "loading"}
              onClick={() => {
                const cleared: WorkflowSavedDraftLibraryFilters = {
                  namePrefix: "",
                  validationState: "",
                  provenanceKind: "",
                };
                setDraftFilters(cleared);
                onLibraryFiltersChange(cleared);
              }}
            >{t($ => $.library.clear)}</button>
          </div>
        </div>
        <article className="workflow-user-workspace-home-card saved-draft-list-status">
          <div className="workflow-user-workspace-home-row-main">
            <div>
              <span>{t($ => $.library.savedDrafts)}</span>
              <strong>{workflowDraftStatusLabel(t, savedDraftListState.status)}</strong>
            </div>
            <StatusBadge tone={workflowSavedDraftListTone(savedDraftListState.status)}>
              {savedDraftListState.mode}
            </StatusBadge>
          </div>
          <dl className="workflow-user-workspace-home-meta">
            <div>
              <dt>{t($ => $.draft.application)}</dt>
              <dd>{savedDraftListState.applicationRef || home.applicationId}</dd>
            </div>
            <div>
              <dt>{t($ => $.draft.lifecycle)}</dt>
              <dd>{workflowDraftStatusLabel(t, savedDraftListState.lifecycleState)}</dd>
            </div>
            <div>
              <dt>{t($ => $.library.loaded)}</dt>
              <dd>{savedDraftListState.summaries.length}</dd>
            </div>
            <div>
              <dt>{t($ => $.library.more)}</dt>
              <dd>{savedDraftListState.hasMore ? t($ => $.library.available) : t($ => $.draft.none)}</dd>
            </div>
            <div>
              <dt>{t($ => $.draft.failure)}</dt>
              <dd>{savedDraftListState.failureCode ?? t($ => $.draft.none)}</dd>
            </div>
            <div>
              <dt>{t($ => $.draft.audit)}</dt>
              <dd>{savedDraftListState.auditRef}</dd>
            </div>
          </dl>
          <p>{workflowDraftLibraryMessage(t, savedDraftListState)}</p>
          {lifecycleOperation.status !== "idle" ? (
            <p>
              {workflowDraftLifecycleMessage(t, lifecycleOperation)}
            </p>
          ) : null}
        </article>
        <div className="saved-draft-summary-grid" aria-label={t($ => $.library.savedSummaries)}>
          {savedDraftListState.summaries.map((summary) => (
            <WorkflowSavedDraftSummaryCard
              key={summary.draftId}
              summary={summary}
              operationPending={
                savedDraftListState.status === "loading" ||
                lifecycleOperation.status === "transitioning"
              }
              archiveConfirmation={archiveConfirmDraftId === summary.draftId}
              onPrepareArchive={() => setArchiveConfirmDraftId(summary.draftId)}
              onCancelArchive={() => setArchiveConfirmDraftId("")}
              onOpenSavedDraft={onOpenSavedDraft}
              onArchiveSavedDraft={(draft) => {
                setArchiveConfirmDraftId("");
                onArchiveSavedDraft(draft);
              }}
              onUnarchiveSavedDraft={onUnarchiveSavedDraft}
            />
          ))}
        </div>
        {savedDraftListState.hasMore ? (
          <button
            className="saved-draft-library-load-more"
            type="button"
            disabled={savedDraftListState.status === "loading"}
            onClick={onLoadMoreSavedDrafts}
          >
            {savedDraftListState.status === "loading" ? t($ => $.library.loading) : t($ => $.library.loadMore)}
          </button>
        ) : null}
      </div>

      <div className="workflow-user-workspace-home-section">
        <div className="workflow-user-workspace-home-subheading">
          <p className="eyebrow">{t($ => $.library.readinessPriorities)}</p>
          <h4>{t($ => $.library.blockedSignals)}</h4>
        </div>
        <div className="workflow-user-workspace-home-readiness-grid" aria-label={t($ => $.library.readinessRollup)}>
          {primaryReadiness.map((readiness) => (
            <WorkflowUserWorkspaceHomeReadinessCard key={readiness.readinessId} readiness={readiness} />
          ))}
        </div>
      </div>

      <div className="workflow-user-workspace-home-section">
        <div className="workflow-user-workspace-home-subheading">
          <p className="eyebrow">{t($ => $.library.runEvidence)}</p>
          <h4>{t($ => $.library.recentRuns)}</h4>
        </div>
        <div className="workflow-user-workspace-home-run-grid" aria-label={t($ => $.library.recentRuns)}>
          {home.recentRuns.map((run) => (
            <WorkflowUserWorkspaceHomeRunCard key={run.runId} run={run} />
          ))}
        </div>
      </div>

      <div className="workflow-user-workspace-home-section">
        <div className="workflow-user-workspace-home-subheading">
          <p className="eyebrow">{t($ => $.library.routeEvidence)}</p>
          <h4>{t($ => $.library.sourceRoutes)}</h4>
        </div>
        <div className="workflow-user-workspace-home-route-grid" aria-label={t($ => $.library.homeRouteEvidence)}>
          {primaryRouteEvidence.map((route) => (
            <WorkflowUserWorkspaceHomeRouteCard key={route.evidenceId} route={route} />
          ))}
        </div>
      </div>

      <div className="workflow-user-workspace-home-section">
        <div className="workflow-user-workspace-home-subheading">
          <p className="eyebrow">{t($ => $.library.stopLines)}</p>
          <h4>{t($ => $.library.lockedCapabilities)}</h4>
        </div>
        <div className="workflow-user-workspace-home-stopline-grid" aria-label={t($ => $.library.homeStopLines)}>
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
  const { t } = useTranslation("workflow");
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
          <dt>{t($ => $.draft.application)}</dt>
          <dd>{application.applicationRef}</dd>
        </div>
        <div>
          <dt>{t($ => $.library.workflow)}</dt>
          <dd>{application.workflowDefinitionId}</dd>
        </div>
        <div>
          <dt>{t($ => $.library.runStatus)}</dt>
          <dd>{application.latestRunStatus}</dd>
        </div>
        <div>
          <dt>{t($ => $.draft.audit)}</dt>
          <dd>{application.auditRef}</dd>
        </div>
      </dl>
      <p>{application.summary}</p>
      <div className="workflow-user-workspace-home-actions">
        <span>{t($ => $.library.localDraftCount, { count: createdDraftCount })}</span>
        <button
          type="button"
          onClick={() => onCreateDraftForWorkflowDefinition(application.workflowDefinitionId)}
        >{t($ => $.library.createDraft)}</button>
      </div>
    </article>
  );
}

function WorkflowSavedDraftSummaryCard({
  summary,
  operationPending,
  archiveConfirmation,
  onPrepareArchive,
  onCancelArchive,
  onOpenSavedDraft,
  onArchiveSavedDraft,
  onUnarchiveSavedDraft,
}: {
  summary: WorkflowSavedDraftSummary;
  operationPending: boolean;
  archiveConfirmation: boolean;
  onPrepareArchive: () => void;
  onCancelArchive: () => void;
  onOpenSavedDraft: (summary: WorkflowSavedDraftSummary) => void;
  onArchiveSavedDraft: (summary: WorkflowSavedDraftSummary) => void;
  onUnarchiveSavedDraft: (summary: WorkflowSavedDraftSummary) => void;
}) {
  const { t, i18n } = useTranslation("workflow");
  const actions = workflowSavedDraftLibraryActions(summary);
  return (
    <article className="workflow-user-workspace-home-card saved-draft-summary-card">
      <div className="workflow-user-workspace-home-row-main">
        <div>
          <p className="eyebrow">{summary.workflowDefinitionId}</p>
          <h5 title={summary.name}>{summary.name}</h5>
        </div>
        <StatusBadge tone={summary.validForReview ? "good" : "neutral"}>{workflowDraftStatusLabel(t, summary.validationState)}</StatusBadge>
      </div>
      <dl className="workflow-user-workspace-home-meta">
        <div>
          <dt>{t($ => $.draft.draft)}</dt>
          <dd>{summary.draftId}</dd>
        </div>
        <div>
          <dt>{t($ => $.draft.version)}</dt>
          <dd>{t($ => $.draft.versions, { content: summary.draftVersion, lifecycle: summary.lifecycleVersion })}</dd>
        </div>
        <div>
          <dt>{t($ => $.library.provenance)}</dt>
          <dd>{workflowDraftStatusLabel(t, summary.provenanceKind)}</dd>
        </div>
        <div>
          <dt>{t($ => $.library.libraryUpdated)}</dt>
          <dd title={summary.libraryUpdatedAt}>{formatDisplayDate(summary.libraryUpdatedAt, i18n.language === "en-US" ? "en-US" : "zh-CN") ?? t($ => $.draft.statusUnknown)}</dd>
        </div>
        <div>
          <dt>{t($ => $.library.archived)}</dt>
          <dd title={summary.archivedAt ?? undefined}>{summary.archivedAt === null ? t($ => $.library.notArchived) : formatDisplayDate(summary.archivedAt, i18n.language === "en-US" ? "en-US" : "zh-CN") ?? t($ => $.draft.statusUnknown)}</dd>
        </div>
        <div>
          <dt>{t($ => $.library.graph)}</dt>
          <dd>
            {t($ => $.draft.graphSize, { nodes: summary.nodeCount, edges: summary.edgeCount })}
          </dd>
        </div>
        <div>
          <dt>{t($ => $.draft.blocked)}</dt>
          <dd>{summary.blockedCapabilityCount}</dd>
        </div>
      </dl>
      <p>{summary.description}</p>
      <div className="workflow-user-workspace-home-actions">
        <span>{summary.sampleOrUnsavedDraftStatus}</span>
        <button type="button" disabled={operationPending} onClick={() => onOpenSavedDraft(summary)}>
          {actions.readOnly ? t($ => $.library.readOnlyReview) : t($ => $.library.openDraft)}
        </button>
        {actions.lifecycleConfirmationRequired ? (
          <button type="button" disabled={operationPending} onClick={onPrepareArchive}>
            {actions.lifecycleTarget === "archived" ? t($ => $.library.archive) : t($ => $.library.unarchive)}
          </button>
        ) : (
          <button type="button" disabled={operationPending} onClick={() => onUnarchiveSavedDraft(summary)}>
            {actions.lifecycleTarget === "archived" ? t($ => $.library.archive) : t($ => $.library.unarchive)}
          </button>
        )}
      </div>
      {archiveConfirmation ? (
        <div className="saved-draft-archive-confirmation" role="group" aria-label={t($ => $.library.archiveLabel, { id: summary.draftId })}>
          <p>
            {t($ => $.library.archiveExplanation, { id: summary.draftId })}
          </p>
          <button type="button" disabled={operationPending} onClick={() => onArchiveSavedDraft(summary)}>{t($ => $.library.confirmArchive)}</button>
          <button type="button" disabled={operationPending} onClick={onCancelArchive}>{t($ => $.draft.cancel)}</button>
        </div>
      ) : null}
    </article>
  );
}

function WorkflowUserWorkspaceHomeStageCard({ stage }: { stage: WorkflowWorkspaceReviewStage }) {
  const { t } = useTranslation("workflow");
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
          <dt>{t($ => $.library.primary)}</dt>
          <dd>{stage.primaryRef}</dd>
        </div>
        <div>
          <dt>{t($ => $.draft.blocked)}</dt>
          <dd>{stage.blockedCount}</dd>
        </div>
        <div>
          <dt>{t($ => $.draft.audit)}</dt>
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
  const { t } = useTranslation("workflow");
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
          <dt>{t($ => $.library.value)}</dt>
          <dd>{readiness.value}</dd>
        </div>
        <div>
          <dt>{t($ => $.draft.audit)}</dt>
          <dd>{readiness.auditRef}</dd>
        </div>
      </dl>
      <p>{readiness.summary}</p>
    </article>
  );
}

function WorkflowUserWorkspaceHomeRunCard({ run }: { run: WorkflowUserWorkspaceHomeRun }) {
  const { t } = useTranslation("workflow");
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
          <dt>{t($ => $.library.workflow)}</dt>
          <dd>{run.workflowDefinitionId}</dd>
        </div>
        <div>
          <dt>{t($ => $.draft.failure)}</dt>
          <dd>{run.failureCode}</dd>
        </div>
        <div>
          <dt>{t($ => $.library.cost)}</dt>
          <dd>{run.cost}</dd>
        </div>
        <div>
          <dt>{t($ => $.library.trace)}</dt>
          <dd>{run.traceId}</dd>
        </div>
      </dl>
      <p>{run.summary}</p>
    </article>
  );
}

function WorkflowUserWorkspaceHomeRouteCard({ route }: { route: WorkflowUserWorkspaceHomeRouteEvidence }) {
  const { t } = useTranslation("workflow");
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
          <dt>{t($ => $.draft.route)}</dt>
          <dd>{route.routeId}</dd>
        </div>
        <div>
          <dt>{t($ => $.draft.request)}</dt>
          <dd>{route.requestId}</dd>
        </div>
        <div>
          <dt>{t($ => $.draft.audit)}</dt>
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

function workflowSavedDraftListTone(status: WorkflowSavedDraftListState["status"]): StatusBadgeTone {
  if (status === "ready") {
    return "good";
  }
  if (status === "list_failed" || status === "open_failed") {
    return "bad";
  }
  return "neutral";
}
import { useEffect, useState } from "react";
