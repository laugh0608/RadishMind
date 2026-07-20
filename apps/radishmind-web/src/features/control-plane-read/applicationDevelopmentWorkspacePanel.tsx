import { useCallback, useEffect, useMemo, useRef, useState, type ReactNode } from "react";

import {
  type ApplicationDevelopmentStage,
  type ApplicationDevelopmentStageId,
  type ApplicationDevelopmentWorkspaceContext,
} from "./applicationDevelopmentWorkspace.ts";
import {
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
  applyApplicationDevelopmentEvidence,
  buildApplicationDevelopmentReadinessViewModel,
  initialApplicationDevelopmentEvidenceState,
  type ApplicationDevelopmentEvidenceInput,
} from "./applicationDevelopmentReadiness.ts";
import type { ApplicationDevelopmentWorkspaceControls } from "./applicationDevelopmentWorkspaceControls.ts";

export default function ApplicationDevelopmentWorkspacePanel({
  context,
  renderStageSurfaces,
}: {
  context: ApplicationDevelopmentWorkspaceContext;
  renderStageSurfaces?: (
    activeStage: ApplicationDevelopmentStageId | null,
    surfaceKey: string,
    controls: ApplicationDevelopmentWorkspaceControls,
  ) => ReactNode;
}) {
  const [routeState, setRouteState] = useState(() => initialApplicationDevelopmentRouteState(context, ""));
  const [handoffState, setHandoffState] = useState(() => initialApplicationDevelopmentHandoffState(context));
  const handoffStateRef = useRef(handoffState);
  const [evidenceState, setEvidenceState] = useState(() => initialApplicationDevelopmentEvidenceState(context));
  const previousActiveStage = useRef<ApplicationDevelopmentStageId | null>(routeState.activeStage);

  useEffect(() => {
    function synchronizeStage() {
      setRouteState((current) => transitionApplicationDevelopmentRoute(current, context, window.location.hash));
    }
    synchronizeStage();
    window.addEventListener("hashchange", synchronizeStage);
    return () => window.removeEventListener("hashchange", synchronizeStage);
  }, [context]);

  const activeStage = routeState.activeStage;
  const readiness = useMemo(
    () => buildApplicationDevelopmentReadinessViewModel(evidenceState),
    [evidenceState],
  );

  useEffect(() => {
    if (previousActiveStage.current !== null && activeStage === null) {
      const clearedHandoff = clearApplicationDevelopmentHandoff(handoffStateRef.current, context);
      handoffStateRef.current = clearedHandoff;
      setHandoffState(clearedHandoff);
      setEvidenceState(initialApplicationDevelopmentEvidenceState(context));
    }
    previousActiveStage.current = activeStage;
  }, [activeStage, context]);

  const reportEvidence = useCallback((input: ApplicationDevelopmentEvidenceInput) => {
    if (input.applicationId !== context.applicationId || input.workspaceGenerationKey !== context.generationKey) return;
    setEvidenceState((current) => applyApplicationDevelopmentEvidence(current, context, input));
  }, [context]);

  const issueHandoff = useCallback((input: ApplicationDevelopmentHandoffInput) => {
    const next = issueApplicationDevelopmentHandoff(handoffStateRef.current, context, input);
    handoffStateRef.current = next;
    setHandoffState(next);
    if (next.pending) window.location.hash = next.pending.targetAnchor;
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

  return (
    <section
      className={`surface-band application-development-workspace ${context.status}`}
      id="application-development-workspace"
      aria-labelledby="application-development-workspace-title"
      data-active-stage={activeStage ?? "inactive"}
      data-route-generation={routeState.routeGeneration}
    >
      <div className="section-heading">
        <div>
          <p className="eyebrow">User Workspace · Application Development</p>
          <h3 id="application-development-workspace-title">Build, test, and review one application scope</h3>
          <p className="application-development-summary">
            Existing domain owners remain authoritative. This shell only organizes navigation and resets transient state when the selected Application revision changes.
          </p>
        </div>
        <span className={`status-badge ${context.status === "active" ? "good" : context.status === "archived" ? "neutral" : "bad"}`}>
          {context.status}
        </span>
      </div>

      <dl className="application-development-context" aria-label="Application development context">
        <div><dt>Application</dt><dd>{context.displayName}</dd><code>{context.applicationId || "not selected"}</code></div>
        <div><dt>Kind</dt><dd>{context.applicationKind}</dd><code>{context.source}</code></div>
        <div><dt>Lifecycle</dt><dd>{context.lifecycleState}</dd><code>revision {context.recordVersion || "not available"}</code></div>
        <div><dt>Workspace</dt><dd>{context.workspaceId || "not available"}</dd><code>{context.ownerSubjectRef || "owner unavailable"}</code></div>
      </dl>

      <nav className="application-development-stages" aria-label="Application development stages">
        {context.stages.map((stage) => (
          <ApplicationDevelopmentStageLink
            key={stage.stageId}
            stage={stage}
            active={stage.stageId === activeStage}
            onNavigate={() => setRouteState((current) => transitionApplicationDevelopmentRoute(current, context, stage.anchor))}
          />
        ))}
      </nav>

      {renderStageSurfaces?.(activeStage, routeState.surfaceKey, controls)}

      <article
        className={`application-development-readiness-boundary ${readiness.status}`}
        id="application-development-workspace-readiness"
      >
        <div>
          <p className="eyebrow">Release readiness boundary</p>
          <h4>Development and test evidence review</h4>
        </div>
        <span className={`status-badge ${readiness.status === "dev_test_evidence_reviewable" ? "good" : readiness.status === "review_blocked" ? "bad" : "neutral"}`}>
          {readiness.status}
        </span>
        <p>{readiness.summary}</p>
        <dl className="application-development-readiness-metrics" aria-label="Readiness evidence metrics">
          <div><dt>Evidence refs</dt><dd>{readiness.evidenceCount}</dd></div>
          <div><dt>Blockers</dt><dd>{readiness.blockerCount}</dd></div>
          <div><dt>Missing</dt><dd>{readiness.missingCount}</dd></div>
        </dl>
        <div className="application-development-readiness-sources">
          {readiness.sources.map((source) => (
            <article key={source.sourceGroupId} className={`application-development-readiness-source ${source.status}`}>
              <div>
                <h5>{source.label}</h5>
                <span className={`status-badge ${source.status === "available" ? "good" : source.status === "blocked" || source.status === "partial_failure" ? "bad" : "neutral"}`}>
                  {source.status} · {source.coverage}
                </span>
              </div>
              {source.evidenceRefs.length ? (
                <ul className="application-development-evidence-refs">
                  {source.evidenceRefs.map((ref) => (
                    <li key={`${ref.kind}:${ref.id}:${ref.version ?? 0}`}>
                      <code>{ref.kind}:{ref.id}{ref.version ? ` · v${ref.version}` : ""}</code>
                    </li>
                  ))}
                </ul>
              ) : null}
              {source.blockers.map((blocker) => (
                <p className="failure-summary" key={blocker.code}><strong>{blocker.code}</strong> · {blocker.summary}</p>
              ))}
              {source.failureCodes.map((failureCode) => <code key={failureCode}>{failureCode}</code>)}
              {source.missingEvidence.map((missing) => <p className="boundary-note" key={missing}>{missing}</p>)}
              {source.status !== "available" ? <a href={`#${source.nextAnchor}`}>Open owning stage</a> : null}
            </article>
          ))}
        </div>
        <p className="boundary-note">
          This projection is ephemeral and read-only. It cannot persist a release decision, publish an Application,
          or satisfy production authorization.
        </p>
      </article>
    </section>
  );
}

function ApplicationDevelopmentStageLink({
  stage,
  active,
  onNavigate,
}: {
  stage: ApplicationDevelopmentStage;
  active: boolean;
  onNavigate: () => void;
}) {
  const blocked = stage.availability === "blocked";
  if (blocked) {
    return (
      <span className="application-development-stage blocked" aria-disabled="true">
        <strong>{stage.label}</strong>
        <small>{stage.summary}</small>
        <code>{stage.availability}</code>
      </span>
    );
  }
  return (
    <a
      className={`application-development-stage ${active ? "active" : ""}`}
      href={`#${stage.anchor}`}
      aria-current={active ? "step" : undefined}
      onClick={onNavigate}
    >
      <strong>{stage.label}</strong>
      <small>{stage.summary}</small>
      <code>{stage.availability}</code>
    </a>
  );
}
