import { useEffect, useState, type ReactNode } from "react";

import {
  type ApplicationDevelopmentStage,
  type ApplicationDevelopmentStageId,
  type ApplicationDevelopmentWorkspaceContext,
} from "./applicationDevelopmentWorkspace.ts";
import {
  initialApplicationDevelopmentRouteState,
  transitionApplicationDevelopmentRoute,
} from "./applicationDevelopmentWorkspaceRoute.ts";

export default function ApplicationDevelopmentWorkspacePanel({
  context,
  renderStageSurfaces,
}: {
  context: ApplicationDevelopmentWorkspaceContext;
  renderStageSurfaces?: (
    activeStage: ApplicationDevelopmentStageId | null,
    surfaceKey: string,
  ) => ReactNode;
}) {
  const [routeState, setRouteState] = useState(() => initialApplicationDevelopmentRouteState(context, ""));

  useEffect(() => {
    function synchronizeStage() {
      setRouteState((current) => transitionApplicationDevelopmentRoute(current, context, window.location.hash));
    }
    synchronizeStage();
    window.addEventListener("hashchange", synchronizeStage);
    return () => window.removeEventListener("hashchange", synchronizeStage);
  }, [context]);

  const activeStage = routeState.activeStage;

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

      {renderStageSurfaces?.(activeStage, routeState.surfaceKey)}

      <article className="application-development-readiness-boundary" id="application-development-workspace-readiness">
        <div>
          <p className="eyebrow">Release readiness boundary</p>
          <h4>Evidence review starts without a publish claim</h4>
        </div>
        <span className="status-badge neutral">review_not_started</span>
        <p>
          Batch A establishes the Application scope and navigation boundary. Source coverage, drift, and blockers remain with their existing owners; the read-only readiness projection follows in Batch B.
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
