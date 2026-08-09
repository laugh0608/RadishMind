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
      if (current.applicationId !== context.applicationId || current.workspaceGenerationKey !== context.generationKey) {
        return current;
      }
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
        <h3 id="application-development-workspace-title">Application Workspace</h3>
        <span>{readiness.sources.length} source groups · {APPLICATION_DEVELOPMENT_CONTRIBUTION_IDS.length} contributions</span>
      </header>

      <section className="application-development-context" aria-label="Application development context">
        <div className="application-development-entity">
          <span className="application-development-entity-mark" aria-hidden="true">
            {applicationInitials(context.displayName)}
          </span>
          <div>
            <div className="application-development-entity-name">
              <strong>{context.displayName}</strong>
              <span className={`application-development-state ${context.status}`}>
                <i aria-hidden="true" />{context.status}
              </span>
            </div>
            <small>{context.applicationKind} · {context.applicationId || "application scope unavailable"}</small>
          </div>
        </div>
        <dl className="application-development-context-facts">
          <div>
            <dt>Revision</dt>
            <dd>{context.recordVersion > 0 ? `v${context.recordVersion}` : "Unavailable"}</dd>
          </div>
          <div>
            <dt>Readiness</dt>
            <dd className={applicationSource?.coverage === "complete" ? "available" : "partial"}>
              {applicationSource?.coverage ?? "none"}
            </dd>
          </div>
          <div>
            <dt>Current stage</dt>
            <dd>{currentStage?.label ?? "Choose a stage"}</dd>
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
            <strong>{currentStage?.label ?? "Choose a stage"}</strong>
          </span>
          <span className="application-development-stage-segments" aria-label={`${currentStageIndex + 1} of ${context.stages.length}`}>
            {context.stages.map((stage, index) => (
              <i key={stage.stageId} className={index === currentStageIndex ? "current" : ""} />
            ))}
          </span>
          <span aria-hidden="true">⌄</span>
        </button>
        {stageMenuOpen ? (
          <nav id="application-development-mobile-stage-menu" aria-label="Application development stages">
            {context.stages.map((stage, index) => (
              <ApplicationDevelopmentStageLink
                key={stage.stageId}
                stage={stage}
                index={index}
                active={stage.stageId === activeStage}
                onNavigate={() => navigateToStage(stage)}
              />
            ))}
          </nav>
        ) : null}
      </div>

      <div className="application-development-workbench">
        <aside className="application-development-stage-rail" aria-label="Application development path">
          <header>
            <span>Development path</span>
            <div><strong>Review path</strong><b>{currentStageIndex >= 0 ? String(currentStageIndex + 1).padStart(2, "0") : "—"} / {String(context.stages.length).padStart(2, "0")}</b></div>
          </header>
          <nav className="application-development-stages" aria-label="Application development stages">
            {context.stages.map((stage, index) => (
              <ApplicationDevelopmentStageLink
                key={stage.stageId}
                stage={stage}
                index={index}
                active={stage.stageId === activeStage}
                onNavigate={() => navigateToStage(stage)}
              />
            ))}
          </nav>
          <footer>
            <span>Current step</span>
            <strong>{currentStage?.label ?? "No active stage"}</strong>
            <div className="application-development-stage-segments" aria-hidden="true">
              {context.stages.map((stage, index) => (
                <i key={stage.stageId} className={index === currentStageIndex ? "current" : ""} />
              ))}
            </div>
          </footer>
        </aside>

        <section className="application-development-contribution-pane" aria-label="Current stage owner contributions">
          <header>
            <div>
              <div>
                <h4>{currentStage?.label ?? "Choose an application development stage"}</h4>
                {currentStage ? <span>current</span> : null}
              </div>
              <p>{representativeContributions.length} representative items · {APPLICATION_DEVELOPMENT_CONTRIBUTION_IDS.length} total contributions</p>
            </div>
            <button
              type="button"
              disabled={!ownerSurfaceAvailable}
              aria-expanded={ownerSurfaceOpen}
              aria-controls="application-development-owner-surface"
              onClick={() => setOwnerSurfaceOpen((open) => !open)}
            >
              {ownerSurfaceOpen ? "Close review" : "Open review"}<span aria-hidden="true">↗</span>
            </button>
          </header>

          <div className="application-development-contributions">
            {representativeContributions.map((contribution) => (
              <RepresentativeContributionRow key={contribution.id} contribution={contribution} />
            ))}
          </div>

          <footer className="application-development-contribution-window">
            <div>
              <span><strong>Contribution window</strong><small>{representativeContributions.length} representative items · {APPLICATION_DEVELOPMENT_CONTRIBUTION_IDS.length} total</small></span>
              <a href="#application-development-readiness-all-sources">View all <span aria-hidden="true">→</span></a>
            </div>
            <div className="application-development-contribution-labels">
              <span>{String(representativeContributions.length).padStart(2, "0")} shown</span>
              <span>{String(APPLICATION_DEVELOPMENT_CONTRIBUTION_IDS.length - representativeContributions.length).padStart(2, "0")} additional</span>
            </div>
            <div className="application-development-contribution-segments" aria-label={`${representativeContributions.length} of ${APPLICATION_DEVELOPMENT_CONTRIBUTION_IDS.length} contributions shown`}>
              {APPLICATION_DEVELOPMENT_CONTRIBUTION_IDS.map((contributionId, index) => (
                <i key={contributionId} className={index < representativeContributions.length ? "shown" : ""} />
              ))}
            </div>
            <small>Across the current {readiness.sources.length} source groups</small>
          </footer>
        </section>

        <aside className="application-development-readiness-pane" aria-label="Evidence and release readiness">
          <header>
            <div><h4>Evidence / readiness</h4><span>Read-only projection · current owners</span></div>
            <b>live view</b>
          </header>

          <div className="application-development-readiness-signal">
            <div className="application-development-readiness-coverage">
              <span>Owner references</span>
              <div>
                <strong>{referencedSourceCount}<small> / {readiness.sources.length}</small></strong>
                <em>groups referenced</em>
                <div className="application-development-readiness-matrix" aria-label={`${referencedSourceCount} source groups referenced`}>
                  {orderedSources.map((source) => <i key={source.sourceGroupId} className={source.status} />)}
                </div>
              </div>
            </div>
            <dl className="application-development-readiness-risks">
              <div><dt>Blocked<small>authority</small></dt><dd>{blockedSourceCount}</dd></div>
              <div><dt>Missing<small>references</small></dt><dd>{readiness.missingCount}</dd></div>
            </dl>
          </div>

          <span className="application-development-source-label">Source groups · current window</span>
          <div className="application-development-source-preview">
            {orderedSources.slice(0, 5).map((source) => <ReadinessSourceRow key={source.sourceGroupId} source={source} />)}
          </div>

          <details className="application-development-all-sources" id="application-development-readiness-all-sources">
            <summary>View all {readiness.sources.length} source groups <span aria-hidden="true">→</span></summary>
            <div>
              {orderedSources.map((source) => <ReadinessSourceRow key={source.sourceGroupId} source={source} />)}
            </div>
          </details>

          <section className="application-development-authorization-path" aria-label="Authorization path">
            <header><strong>Authorization path</strong><span>read only</span></header>
            <dl>
              <div><dt>01 Evidence</dt><dd>{referencedSourceCount} referenced</dd></div>
              <div><dt>02 Review</dt><dd>Human</dd></div>
              <div><dt>03 Production</dt><dd>Closed</dd></div>
            </dl>
          </section>

          <p className="application-development-stop-line">
            <span aria-hidden="true">!</span>
            This projection is volatile, read-only and not publishable. Production authorization remains closed.
          </p>
        </aside>

        <section
          className="application-development-owner-surface"
          id="application-development-owner-surface"
          hidden={!ownerSurfaceOpen}
          aria-label="Current stage owner surface"
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
  index,
  active,
  onNavigate,
}: {
  stage: ApplicationDevelopmentStage;
  index: number;
  active: boolean;
  onNavigate: () => void;
}) {
  const blocked = stage.availability === "blocked";
  const content = (
    <>
      <i className="application-development-stage-accent" aria-hidden="true" />
      <b>{String(index + 1).padStart(2, "0")}</b>
      <span><strong>{stage.label}</strong><small>{active ? "current" : stage.availability.replace("_", " ")}</small></span>
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
  const status = contributionStatusLabel(contribution.status, contribution.coverage);
  return (
    <article className="application-development-contribution" data-status={contribution.status}>
      <span className={`application-development-contribution-mark ${contribution.status}`} aria-hidden="true">
        {contribution.mark}
      </span>
      <div>
        <header><strong>{contribution.label}</strong><span className={contribution.status}>{status}</span></header>
        <p>{contribution.description}</p>
        <footer><small>{contribution.owner}</small><a href={`#${contribution.nextAnchor}`} aria-label={`Open ${contribution.label} owner`}>↗</a></footer>
      </div>
    </article>
  );
}

function ReadinessSourceRow({ source }: { source: ApplicationDevelopmentReadinessSource }) {
  return (
    <div className="application-development-source-row">
      <span><i className={source.status} aria-hidden="true" />{source.label}</span>
      {source.status === "available" ? (
        <strong className={source.status}>{source.status}</strong>
      ) : (
        <a className={source.status} href={`#${source.nextAnchor}`}>{source.status.replace("_", " ")}</a>
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

function contributionStatusLabel(
  status: ApplicationDevelopmentEvidenceStatus,
  coverage: ApplicationDevelopmentReadinessSource["coverage"],
): string {
  if (status === "not_started") return "missing";
  if (status === "incomplete") return coverage === "partial" ? "partial" : "missing";
  return status.replace("_", " ");
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
