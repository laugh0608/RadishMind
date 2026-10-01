import "../../i18n/workspacePanelResources.ts";
import { useTranslation } from "react-i18next";
import { useCallback, useEffect, useMemo, useRef, useState, type ReactNode } from "react";

import {
  applicationDevelopmentHashTargetsOwnerSurface,
  type ApplicationDevelopmentStage,
  type ApplicationDevelopmentStageId,
  type ApplicationDevelopmentWorkspaceContext,
} from "./applicationDevelopmentWorkspace.ts";
import {
  applicationDevelopmentRouteAcceptsResponse,
  initialApplicationDevelopmentRouteState,
  transitionApplicationDevelopmentRoute,
} from "./applicationDevelopmentWorkspaceRoute.ts";
import {
  clearApplicationDevelopmentHandoff,
  consumeApplicationDevelopmentHandoff,
  initialApplicationDevelopmentHandoffState,
  issueApplicationDevelopmentHandoff,
  type ApplicationDevelopmentHandoffInput,
} from "./applicationDevelopmentHandoff.ts";
import {
  APPLICATION_DEVELOPMENT_CONTRIBUTION_IDS,
  applicationDevelopmentEvidenceMatchesScope,
  applyApplicationDevelopmentEvidence,
  buildApplicationDevelopmentReadinessViewModel,
  initialApplicationDevelopmentEvidenceState,
  type ApplicationDevelopmentEvidenceContribution,
  type ApplicationDevelopmentEvidenceStatus,
  type ApplicationDevelopmentReadinessSource,
  type ApplicationDevelopmentSourceGroupId,
} from "./applicationDevelopmentReadiness.ts";
import type {
  ApplicationDevelopmentEvidenceReport,
  ApplicationDevelopmentWorkspaceControls,
} from "./applicationDevelopmentWorkspaceControls.ts";
import { promptAgentTypeWorkspaceOwnsHash } from "./promptAgentTypeWorkspaceModel.ts";

const REPRESENTATIVE_CONTRIBUTIONS = [
  {
    kind: "contribution",
    contributionId: "publish_candidate",
    label: "Application candidate",
    owner: "Application owner",
    mark: "A",
  },
  {
    kind: "contribution",
    contributionId: "workflow_definition",
    label: "Workflow definition",
    owner: "Workflow owner",
    mark: "W",
  },
  {
    kind: "source",
    sourceGroupId: "rag_authority",
    label: "RAG authority",
    owner: "RAG authority",
    mark: "R",
  },
] as const;

const SOURCE_PRIORITY: readonly ApplicationDevelopmentSourceGroupId[] = [
  "application",
  "rag_authority",
  "workflow_authority",
  "configuration_candidate",
  "operations",
  "prompt_authority",
  "agent_authority",
  "controlled_test",
  "evaluation",
];

type RepresentativeContribution = {
  id: string;
  label: string;
  owner: string;
  mark: string;
  status: ApplicationDevelopmentEvidenceStatus;
  coverage: ApplicationDevelopmentReadinessSource["coverage"];
  description: string;
  descriptionKind: "raw" | "reference" | "no_owner_required" | "not_loaded";
  reference: string;
  nextAnchor: string;
};

export default function ApplicationDevelopmentWorkspacePanel({
  context,
  renderStageSurfaces,
  renderPersistentSurfaces,
}: {
  context: ApplicationDevelopmentWorkspaceContext;
  renderStageSurfaces?: (
    activeStage: ApplicationDevelopmentStageId | null,
    surfaceKey: string,
    controls: ApplicationDevelopmentWorkspaceControls,
  ) => ReactNode;
  renderPersistentSurfaces?: (
    surfaceKey: string,
    controls: ApplicationDevelopmentWorkspaceControls,
  ) => ReactNode;
}) {
  const { t } = useTranslation("applications");
  const [routeState, setRouteState] = useState(() => initialApplicationDevelopmentRouteState(context, ""));
  const routeStateRef = useRef(routeState);
  routeStateRef.current = routeState;
  const [handoffState, setHandoffState] = useState(() => initialApplicationDevelopmentHandoffState(context));
  const handoffStateRef = useRef(handoffState);
  const [evidenceState, setEvidenceState] = useState(() => initialApplicationDevelopmentEvidenceState(context));
  const [stageMenuOpen, setStageMenuOpen] = useState(false);
  const [ownerSurfaceOpen, setOwnerSurfaceOpen] = useState(false);
  const previousActiveStage = useRef<ApplicationDevelopmentStageId | null>(routeState.activeStage);

  useEffect(() => {
    function synchronizeStage() {
      setRouteState((current) => transitionApplicationDevelopmentRoute(current, context, window.location.hash));
      if (ownerSurfaceTargeted(context, window.location.hash)) setOwnerSurfaceOpen(true);
    }
    synchronizeStage();
    window.addEventListener("hashchange", synchronizeStage);
    return () => window.removeEventListener("hashchange", synchronizeStage);
  }, [context]);

  const activeStage = routeState.activeStage;
  const currentStageIndex = context.stages.findIndex((stage) => stage.stageId === activeStage);
  const currentStage = currentStageIndex >= 0 ? context.stages[currentStageIndex] : null;
  const stageLabels = {
    configure_build: t($ => $.workspacePanel.stageConfigureBuild),
    human_promotion: t($ => $.workspacePanel.stageHumanPromotion),
    controlled_test: t($ => $.workspacePanel.stageControlledTest),
    evidence_review: t($ => $.workspacePanel.stageEvidenceReview),
    release_readiness: t($ => $.workspacePanel.stageReleaseReadiness),
  } satisfies Record<ApplicationDevelopmentStageId, string>;
  const readiness = useMemo(
    () => buildApplicationDevelopmentReadinessViewModel(evidenceState),
    [evidenceState],
  );
  const orderedSources = useMemo(
    () => SOURCE_PRIORITY.map((sourceId) => readiness.sources.find((source) => source.sourceGroupId === sourceId))
      .filter((source): source is ApplicationDevelopmentReadinessSource => Boolean(source))
      .map(applicationDevelopmentPresentationSource),
    [readiness.sources],
  );
  const representativeContributions = useMemo(
    () => buildRepresentativeContributions(evidenceState.contributions, orderedSources),
    [evidenceState.contributions, orderedSources],
  );
  const referencedSourceCount = readiness.sources.filter((source) => source.evidenceRefs.length > 0).length;
  const blockedSourceCount = orderedSources.filter(
    (source) => source.status === "blocked" || source.status === "partial_failure",
  ).length;
  const applicationSource = orderedSources.find((source) => source.sourceGroupId === "application");
  const ragSource = orderedSources.find((source) => source.sourceGroupId === "rag_authority");
  const ownerRagSource = readiness.sources.find((source) => source.sourceGroupId === "rag_authority");

  useEffect(() => {
    if (previousActiveStage.current !== null && activeStage === null) {
      const clearedHandoff = clearApplicationDevelopmentHandoff(handoffStateRef.current, context);
      handoffStateRef.current = clearedHandoff;
      setHandoffState(clearedHandoff);
      setEvidenceState(initialApplicationDevelopmentEvidenceState(context));
    }
    if (previousActiveStage.current !== activeStage) {
      setStageMenuOpen(false);
      const targetsPendingOwner = handoffStateRef.current.pending?.targetStage === activeStage;
      setOwnerSurfaceOpen(Boolean(targetsPendingOwner) || ownerSurfaceTargeted(context, window.location.hash));
    }
    previousActiveStage.current = activeStage;
  }, [activeStage, context]);

  const reportEvidence = useCallback((input: ApplicationDevelopmentEvidenceReport) => {
    if (input.applicationId !== context.applicationId || input.workspaceGenerationKey !== context.generationKey) return;
    if (!applicationDevelopmentRouteAcceptsResponse(input.surfaceKey, routeStateRef.current)) return;
    setEvidenceState((current) => {
      if (!applicationDevelopmentEvidenceMatchesScope(current, context, input)) return current;
      return applyApplicationDevelopmentEvidence(current, context, input);
    });
  }, [context]);

  const issueHandoff = useCallback((input: ApplicationDevelopmentHandoffInput) => {
    const next = issueApplicationDevelopmentHandoff(handoffStateRef.current, context, input);
    handoffStateRef.current = next;
    setHandoffState(next);
    if (next.pending) {
      setOwnerSurfaceOpen(true);
      window.location.hash = next.pending.targetAnchor;
    }
  }, [context]);

  const consumeHandoff = useCallback((targetStage: ApplicationDevelopmentStageId, handoffId: string) => {
    const consumed = consumeApplicationDevelopmentHandoff(
      handoffStateRef.current,
      context,
      targetStage,
      handoffId,
    );
    if (!consumed.handoff) return;
    handoffStateRef.current = consumed.state;
    setHandoffState(consumed.state);
  }, [context]);

  const controls = useMemo<ApplicationDevelopmentWorkspaceControls>(() => ({
    readiness,
    pendingHandoff: handoffState.pending,
    reportEvidence,
    issueHandoff,
    consumeHandoff,
  }), [consumeHandoff, handoffState.pending, issueHandoff, readiness, reportEvidence]);

  const navigateToStage = useCallback((stage: ApplicationDevelopmentStage) => {
    setStageMenuOpen(false);
    setRouteState((current) => transitionApplicationDevelopmentRoute(current, context, stage.anchor));
  }, [context]);

  const ownerSurfaceAvailable = activeStage !== null && activeStage !== "release_readiness";

  return (
    <section
      className={`surface-band application-development-workspace ${context.status}`}
      id="application-development-workspace"
      aria-labelledby="application-development-workspace-title"
      data-active-stage={activeStage ?? "inactive"}
      data-route-generation={routeState.routeGeneration}
      data-source-group-count={readiness.sources.length}
      data-contribution-count={APPLICATION_DEVELOPMENT_CONTRIBUTION_IDS.length}
      data-application-coverage={applicationSource?.coverage ?? "none"}
      data-rag-status={ragSource?.status ?? "not_started"}
      data-rag-owner-status={ownerRagSource?.status ?? "not_started"}
      data-publishable={String(readiness.canPublish)}
    >
      <header className="application-development-heading">
        <h3 id="application-development-workspace-title">{t($ => $.workspacePanel.applicationWorkspace)}</h3>
        <span>{t($ => $.workspacePanel.countSourceContributions, { sourceCount: readiness.sources.length, contributionCount: APPLICATION_DEVELOPMENT_CONTRIBUTION_IDS.length })}</span>
      </header>

      <section className="application-development-context" aria-label={t($ => $.workspacePanel.applicationDevelopmentContext)}>
        <div className="application-development-entity">
          <span className="application-development-entity-mark" aria-hidden="true">
            {applicationInitials(context.displayName)}
          </span>
          <div>
            <div className="application-development-entity-name">
              <strong>{context.displayName}</strong>
              <span className={`application-development-state ${context.status}`}>
                <i aria-hidden="true" />{context.status === "active" ? t($ => $.workspacePanel.statusActive) : context.status === "archived" ? t($ => $.workspacePanel.statusArchived) : t($ => $.workspacePanel.statusUnavailable)}
              </span>
            </div>
            <small>{context.applicationKind} · {context.applicationId || t($ => $.workspacePanel.applicationScopeUnavailable)}</small>
          </div>
        </div>
        <dl className="application-development-context-facts">
          <div>
            <dt>{t($ => $.workspacePanel.revision)}</dt>
            <dd>{context.recordVersion > 0 ? `v${context.recordVersion}` : t($ => $.workspacePanel.unavailable)}</dd>
          </div>
          <div>
            <dt>{t($ => $.workspacePanel.readiness)}</dt>
            <dd className={applicationSource?.coverage === "complete" ? "available" : "partial"}>
              {applicationSource?.coverage === "complete" ? t($ => $.workspacePanel.coverageComplete) : applicationSource?.coverage === "partial" ? t($ => $.workspacePanel.coveragePartial) : t($ => $.workspacePanel.coverageNone)}
            </dd>
          </div>
          <div>
            <dt>{t($ => $.workspacePanel.currentStage)}</dt>
            <dd>{currentStage ? stageLabels[currentStage.stageId] : t($ => $.workspacePanel.chooseStage)}</dd>
          </div>
        </dl>
      </section>

      <div className="application-development-mobile-stage">
        <button
          type="button"
          aria-expanded={stageMenuOpen}
          aria-controls="application-development-mobile-stage-menu"
          onClick={() => setStageMenuOpen((open) => !open)}
        >
          <span className="application-development-mobile-stage-current">
            <b>{currentStageIndex >= 0 ? String(currentStageIndex + 1).padStart(2, "0") : "—"}</b>
            <strong>{currentStage ? stageLabels[currentStage.stageId] : t($ => $.workspacePanel.chooseStage)}</strong>
          </span>
          <span className="application-development-stage-segments" aria-label={`${currentStageIndex + 1} of ${context.stages.length}`}>
            {context.stages.map((stage, index) => (
              <i key={stage.stageId} className={index === currentStageIndex ? "current" : ""} />
            ))}
          </span>
          <span aria-hidden="true">⌄</span>
        </button>
        {stageMenuOpen ? (
          <nav id="application-development-mobile-stage-menu" aria-label={t($ => $.workspacePanel.developmentStages)}>
            {context.stages.map((stage, index) => (
              <ApplicationDevelopmentStageLink
                key={stage.stageId}
                stage={stage}
                label={stageLabels[stage.stageId]}
                index={index}
                active={stage.stageId === activeStage}
                onNavigate={() => navigateToStage(stage)}
              />
            ))}
          </nav>
        ) : null}
      </div>

      <div className="application-development-workbench">
        <aside className="application-development-stage-rail" aria-label={t($ => $.workspacePanel.developmentPathAria)}>
          <header>
            <span>{t($ => $.workspacePanel.developmentPath)}</span>
            <div><strong>{t($ => $.workspacePanel.reviewPath)}</strong><b>{currentStageIndex >= 0 ? String(currentStageIndex + 1).padStart(2, "0") : "—"} / {String(context.stages.length).padStart(2, "0")}</b></div>
          </header>
          <nav className="application-development-stages" aria-label={t($ => $.workspacePanel.developmentStages)}>
            {context.stages.map((stage, index) => (
              <ApplicationDevelopmentStageLink
                key={stage.stageId}
                stage={stage}
                label={stageLabels[stage.stageId]}
                index={index}
                active={stage.stageId === activeStage}
                onNavigate={() => navigateToStage(stage)}
              />
            ))}
          </nav>
          <footer>
            <span>{t($ => $.workspacePanel.currentStep)}</span>
            <strong>{currentStage ? stageLabels[currentStage.stageId] : t($ => $.workspacePanel.noActiveStage)}</strong>
            <div className="application-development-stage-segments" aria-hidden="true">
              {context.stages.map((stage, index) => (
                <i key={stage.stageId} className={index === currentStageIndex ? "current" : ""} />
              ))}
            </div>
          </footer>
        </aside>

        <section className="application-development-contribution-pane" aria-label={t($ => $.workspacePanel.currentStageOwnerContributions)}>
          <header>
            <div>
              <div>
                <h4>{currentStage ? stageLabels[currentStage.stageId] : t($ => $.workspacePanel.chooseDevelopmentStage)}</h4>
                {currentStage ? <span>{t($ => $.workspacePanel.current)}</span> : null}
              </div>
              <p>{t($ => $.workspacePanel.countRepresentativeContributions, { representativeCount: representativeContributions.length, contributionCount: APPLICATION_DEVELOPMENT_CONTRIBUTION_IDS.length })}</p>
            </div>
            <button
              type="button"
              disabled={!ownerSurfaceAvailable}
              aria-expanded={ownerSurfaceOpen}
              aria-controls="application-development-owner-surface"
              onClick={() => setOwnerSurfaceOpen((open) => !open)}
            >
              {ownerSurfaceOpen ? t($ => $.workspacePanel.closeReview) : t($ => $.workspacePanel.openReview)}<span aria-hidden="true">↗</span>
            </button>
          </header>

          <div className="application-development-contributions">
            {representativeContributions.map((contribution) => (
              <RepresentativeContributionRow key={contribution.id} contribution={contribution} />
            ))}
          </div>

          <footer className="application-development-contribution-window">
            <div>
              <span><strong>{t($ => $.workspacePanel.contributionWindow)}</strong><small>{t($ => $.workspacePanel.countRepresentativeTotal, { representativeCount: representativeContributions.length, contributionCount: APPLICATION_DEVELOPMENT_CONTRIBUTION_IDS.length })}</small></span>
              <a href="#application-development-readiness-all-sources">{t($ => $.workspacePanel.viewAll)}<span aria-hidden="true">→</span></a>
            </div>
            <div className="application-development-contribution-labels">
              <span>{t($ => $.workspacePanel.countShown, { value: String(representativeContributions.length).padStart(2, "0") })}</span>
              <span>{t($ => $.workspacePanel.countAdditional, { value: String(APPLICATION_DEVELOPMENT_CONTRIBUTION_IDS.length - representativeContributions.length).padStart(2, "0") })}</span>
            </div>
            <div className="application-development-contribution-segments" aria-label={t($ => $.workspacePanel.countContributionsShown, { shownCount: representativeContributions.length, totalCount: APPLICATION_DEVELOPMENT_CONTRIBUTION_IDS.length })}>
              {APPLICATION_DEVELOPMENT_CONTRIBUTION_IDS.map((contributionId, index) => (
                <i key={contributionId} className={index < representativeContributions.length ? "shown" : ""} />
              ))}
            </div>
            <small>{t($ => $.workspacePanel.countAcrossSources, { count: readiness.sources.length })}</small>
          </footer>
        </section>

        <aside className="application-development-readiness-pane" aria-label={t($ => $.workspacePanel.evidenceReadinessAria)}>
          <header>
            <div><h4>{t($ => $.workspacePanel.evidenceReadiness)}</h4><span>{t($ => $.workspacePanel.readonlyProjection)}</span></div>
            <b>{t($ => $.workspacePanel.liveView)}</b>
          </header>

          <div className="application-development-readiness-signal">
            <div className="application-development-readiness-coverage">
              <span>{t($ => $.workspacePanel.ownerReferences)}</span>
              <div>
                <strong>{referencedSourceCount}<small> / {readiness.sources.length}</small></strong>
                <em>{t($ => $.workspacePanel.countSourcesReferenced, { count: referencedSourceCount })}</em>
                <div className="application-development-readiness-matrix" aria-label={t($ => $.workspacePanel.countSourcesReferenced, { count: referencedSourceCount })}>
                  {orderedSources.map((source) => <i key={source.sourceGroupId} className={source.status} />)}
                </div>
              </div>
            </div>
            <dl className="application-development-readiness-risks">
              <div><dt>{t($ => $.workspacePanel.blocked)}<small>{t($ => $.workspacePanel.authority)}</small></dt><dd>{blockedSourceCount}</dd></div>
              <div><dt>{t($ => $.workspacePanel.missing)}<small>{t($ => $.workspacePanel.references)}</small></dt><dd>{readiness.missingCount}</dd></div>
            </dl>
          </div>

          <span className="application-development-source-label">{t($ => $.workspacePanel.sourceGroupsCurrentWindow)}</span>
          <div className="application-development-source-preview">
            {orderedSources.slice(0, 5).map((source) => <ReadinessSourceRow key={source.sourceGroupId} source={source} />)}
          </div>

          <details className="application-development-all-sources" id="application-development-readiness-all-sources">
            <summary>{t($ => $.workspacePanel.viewAllSources, { count: readiness.sources.length })}<span aria-hidden="true">→</span></summary>
            <div>
              {orderedSources.map((source) => <ReadinessSourceRow key={source.sourceGroupId} source={source} />)}
            </div>
          </details>

          <section className="application-development-authorization-path" aria-label={t($ => $.workspacePanel.authorizationPath)}>
            <header><strong>{t($ => $.workspacePanel.authorizationPath)}</strong><span>{t($ => $.workspacePanel.readOnly)}</span></header>
            <dl>
              <div><dt>{t($ => $.workspacePanel.evidenceStep)}</dt><dd>{t($ => $.workspacePanel.countReferenced, { count: referencedSourceCount })}</dd></div>
              <div><dt>{t($ => $.workspacePanel.reviewStep)}</dt><dd>{t($ => $.workspacePanel.human)}</dd></div>
              <div><dt>{t($ => $.workspacePanel.productionStep)}</dt><dd>{t($ => $.workspacePanel.closed)}</dd></div>
            </dl>
          </section>

          <p className="application-development-stop-line">
            <span aria-hidden="true">!</span>
            {t($ => $.workspacePanel.projectionBoundary)}</p>
        </aside>

        <section
          className="application-development-owner-surface"
          id="application-development-owner-surface"
          hidden={!ownerSurfaceOpen}
          aria-label={t($ => $.workspacePanel.currentStageOwnerSurface)}
        >
          {renderStageSurfaces?.(activeStage, routeState.surfaceKey, controls)}
        </section>
      </div>
      {renderPersistentSurfaces?.(routeState.surfaceKey, controls)}
    </section>
  );
}

function ApplicationDevelopmentStageLink({
  stage,
  label,
  index,
  active,
  onNavigate,
}: {
  stage: ApplicationDevelopmentStage;
  label: string;
  index: number;
  active: boolean;
  onNavigate: () => void;
}) {
  const { t } = useTranslation("applications");
  const blocked = stage.availability === "blocked";
  const content = (
    <>
      <i className="application-development-stage-accent" aria-hidden="true" />
      <b>{String(index + 1).padStart(2, "0")}</b>
      <span><strong>{label}</strong><small>{active ? t($ => $.workspacePanel.current) : stage.availability === "available" ? t($ => $.workspacePanel.statusAvailable) : stage.availability === "read_only" ? t($ => $.workspacePanel.statusReadOnly) : t($ => $.workspacePanel.statusBlocked)}</small></span>
      {active ? <em aria-hidden="true">›</em> : null}
    </>
  );
  if (blocked) {
    return <span className="application-development-stage blocked" aria-disabled="true">{content}</span>;
  }
  return (
    <a
      className={`application-development-stage ${active ? "active" : ""}`}
      href={`#${stage.anchor}`}
      aria-current={active ? "step" : undefined}
      onClick={onNavigate}
    >
      {content}
    </a>
  );
}

function RepresentativeContributionRow({ contribution }: { contribution: RepresentativeContribution }) {
  const { t } = useTranslation("applications");
  const status = contribution.status === "not_started" ? t($ => $.workspacePanel.missing)
    : contribution.status === "incomplete" ? contribution.coverage === "partial" ? t($ => $.workspacePanel.coveragePartial) : t($ => $.workspacePanel.missing)
      : contribution.status === "available" ? t($ => $.workspacePanel.statusAvailable)
        : contribution.status === "blocked" ? t($ => $.workspacePanel.statusBlocked) : t($ => $.workspacePanel.statusPartialFailure);
  const label = contribution.id === "publish_candidate" ? t($ => $.workspacePanel.applicationCandidate)
    : contribution.id === "workflow_definition" ? t($ => $.workspacePanel.workflowDefinition)
      : t($ => $.workspacePanel.ragAuthority);
  const owner = contribution.id === "publish_candidate" ? t($ => $.workspacePanel.applicationOwner)
    : contribution.id === "workflow_definition" ? t($ => $.workspacePanel.workflowOwner)
      : t($ => $.workspacePanel.ragAuthority);
  return (
    <article className="application-development-contribution" data-status={contribution.status}>
      <span className={`application-development-contribution-mark ${contribution.status}`} aria-hidden="true">
        {contribution.mark}
      </span>
      <div>
        <header><strong>{label}</strong><span className={contribution.status}>{status}</span></header>
        <p>{contribution.descriptionKind === "reference"
          ? t($ => $.workspacePanel.evidenceAvailableFromOwner, { reference: contribution.reference })
          : contribution.descriptionKind === "no_owner_required"
            ? t($ => $.workspacePanel.noOwnerEvidenceRequired)
            : contribution.descriptionKind === "not_loaded"
              ? t($ => $.workspacePanel.ownerEvidenceNotLoaded)
              : contribution.description}</p>
        <footer><small>{owner}</small><a href={`#${contribution.nextAnchor}`} aria-label={t($ => $.workspacePanel.openContributionOwner, { label })}>↗</a></footer>
      </div>
    </article>
  );
}

function ReadinessSourceRow({ source }: { source: ApplicationDevelopmentReadinessSource }) {
  const { t } = useTranslation("applications");
  const labels = {
    application: t($ => $.workspacePanel.sourceApplication),
    configuration_candidate: t($ => $.workspacePanel.sourceConfigurationCandidate),
    workflow_authority: t($ => $.workspacePanel.sourceWorkflowAuthority),
    rag_authority: t($ => $.workspacePanel.sourceRagAuthority),
    prompt_authority: t($ => $.workspacePanel.sourcePromptAuthority),
    agent_authority: t($ => $.workspacePanel.sourceAgentAuthority),
    controlled_test: t($ => $.workspacePanel.sourceControlledTest),
    evaluation: t($ => $.workspacePanel.sourceEvaluation),
    operations: t($ => $.workspacePanel.sourceOperations),
  } satisfies Record<ApplicationDevelopmentSourceGroupId, string>;
  const status = source.status === "available" ? t($ => $.workspacePanel.statusAvailable)
    : source.status === "blocked" ? t($ => $.workspacePanel.statusBlocked)
      : source.status === "not_started" ? t($ => $.workspacePanel.statusNotStarted)
        : source.status === "incomplete" ? t($ => $.workspacePanel.statusIncomplete) : t($ => $.workspacePanel.statusPartialFailure);
  return (
    <div className="application-development-source-row">
      <span><i className={source.status} aria-hidden="true" />{labels[source.sourceGroupId]}</span>
      {source.status === "available" ? (
        <strong className={source.status}>{status}</strong>
      ) : (
        <a className={source.status} href={`#${source.nextAnchor}`}>{status}</a>
      )}
    </div>
  );
}

function buildRepresentativeContributions(
  contributions: Record<string, ApplicationDevelopmentEvidenceContribution>,
  sources: ApplicationDevelopmentReadinessSource[],
): RepresentativeContribution[] {
  return REPRESENTATIVE_CONTRIBUTIONS.map((definition) => {
    const evidence = definition.kind === "contribution"
      ? contributions[definition.contributionId]
      : sources.find((source) => source.sourceGroupId === definition.sourceGroupId);
    if (!evidence) throw new Error("Application development representative evidence is unavailable.");
    return {
      id: definition.kind === "contribution" ? definition.contributionId : definition.sourceGroupId,
      label: definition.label,
      owner: definition.owner,
      mark: definition.mark,
      status: evidence.status,
      coverage: evidence.coverage,
      description: evidenceDescription(evidence),
      descriptionKind: evidence.blockers[0] || evidence.missingEvidence[0] ? "raw"
        : evidence.evidenceRefs[0] ? "reference"
          : evidence.status === "available" ? "no_owner_required" : "not_loaded",
      reference: evidence.evidenceRefs[0]
        ? `${evidence.evidenceRefs[0].kind}:${evidence.evidenceRefs[0].id}${evidence.evidenceRefs[0].version ? ` · v${evidence.evidenceRefs[0].version}` : ""}`
        : "",
      nextAnchor: evidence.nextAnchor,
    };
  });
}

function evidenceDescription(
  evidence: Pick<
    ApplicationDevelopmentEvidenceContribution | ApplicationDevelopmentReadinessSource,
    "status" | "evidenceRefs" | "missingEvidence" | "blockers"
  >,
): string {
  const blocker = evidence.blockers[0]?.summary;
  if (blocker) return blocker;
  const missing = evidence.missingEvidence[0];
  if (missing) return missing;
  const ref = evidence.evidenceRefs[0];
  if (ref) return `${ref.kind}:${ref.id}${ref.version ? ` · v${ref.version}` : ""} is available from its current owner.`;
  if (evidence.status === "available") return "No owner evidence is required for this Application kind.";
  return "Owner evidence has not been loaded for the current Application generation.";
}

function applicationInitials(displayName: string): string {
  const initials = displayName.trim().split(/\s+/u).filter(Boolean).slice(0, 2).map((part) => part[0]?.toUpperCase()).join("");
  return initials || "—";
}

function ownerSurfaceTargeted(context: ApplicationDevelopmentWorkspaceContext, hash: string): boolean {
  return applicationDevelopmentHashTargetsOwnerSurface(hash) ||
    promptAgentTypeWorkspaceOwnsHash(hash, context.surfaceKind);
}

function applicationDevelopmentPresentationSource(
  source: ApplicationDevelopmentReadinessSource,
): ApplicationDevelopmentReadinessSource {
  if (source.sourceGroupId !== "rag_authority" || source.status === "available" || source.status === "blocked" || source.status === "partial_failure") {
    return source;
  }
  const assignmentMissing = source.missingEvidence.find((item) => /assignment/iu.test(item));
  if (!assignmentMissing) return source;
  return {
    ...source,
    status: "blocked",
    missingEvidence: [assignmentMissing, ...source.missingEvidence.filter((item) => item !== assignmentMissing)],
  };
}
