import "../../i18n/runReviewResources.ts";
import { useTranslation } from "react-i18next";
import { lazy, Suspense, useCallback, useEffect, useState } from "react";

import {
  workflowReviewSurfaceForHash,
  type ApplicationDevelopmentWorkspaceContext,
  type WorkflowReviewSurface,
} from "./applicationDevelopmentWorkspace.ts";
import type { ApplicationDevelopmentWorkspaceControls } from "./applicationDevelopmentWorkspaceControls.ts";

const WorkflowReviewOwner = lazy(() => import("./workflowReviewOwner.tsx"));

const WORKFLOW_REVIEW_TASKS: ReadonlyArray<{
  surface: WorkflowReviewSurface;
  anchor: string;
  label: string;
  summary: string;
  number: string;
}> = [
  {
    surface: "runs",
    anchor: "workspace-run-history",
    label: "Runs",
    summary: "Locate exact run evidence",
    number: "01",
  },
  {
    surface: "comparison",
    anchor: "workflow-run-comparison",
    label: "Compare",
    summary: "Review compatible deltas",
    number: "02",
  },
  {
    surface: "cases",
    anchor: "workflow-evaluation-cases",
    label: "Cases",
    summary: "Bind immutable run refs",
    number: "03",
  },
  {
    surface: "release",
    anchor: "workflow-evaluation-release-review",
    label: "Release",
    summary: "Record human evidence",
    number: "04",
  },
];

export default function WorkflowReviewWorkspace({
  context,
  controls,
  refreshKey,
}: {
  context: ApplicationDevelopmentWorkspaceContext;
  surfaceKey: string;
  controls: ApplicationDevelopmentWorkspaceControls;
  refreshKey: number;
}) {
  const { t } = useTranslation("evaluation");
  const [activeSurface, setActiveSurface] = useState<WorkflowReviewSurface | null>(() => (
    workflowReviewSurfaceForHash(window.location.hash)
  ));

  useEffect(() => {
    function synchronizeSurface() {
      const nextSurface = workflowReviewSurfaceForHash(window.location.hash);
      setActiveSurface(nextSurface);
      if (nextSurface) {
        window.requestAnimationFrame(() => {
          document.querySelector<HTMLElement>(".workflow-review-workspace")?.scrollIntoView({ block: "start" });
        });
      }
    }
    synchronizeSurface();
    window.addEventListener("hashchange", synchronizeSurface);
    return () => window.removeEventListener("hashchange", synchronizeSurface);
  }, []);

  useEffect(() => {
    if (!activeSurface) return;
    let secondFrame = 0;
    const firstFrame = window.requestAnimationFrame(() => {
      secondFrame = window.requestAnimationFrame(() => {
        document.querySelector<HTMLElement>(".workflow-review-workspace")?.scrollIntoView({ block: "start" });
      });
    });
    const settledScroll = window.setTimeout(() => {
      document.querySelector<HTMLElement>(".workflow-review-workspace")?.scrollIntoView({ block: "start" });
    }, 250);
    return () => {
      window.cancelAnimationFrame(firstFrame);
      if (secondFrame) window.cancelAnimationFrame(secondFrame);
      window.clearTimeout(settledScroll);
    };
  }, [activeSurface]);

  const pendingRunHandoff = controls.pendingHandoff?.targetStage === "evidence_review" &&
    controls.pendingHandoff.refKind === "run"
    ? controls.pendingHandoff
    : null;
  const consumeRunHandoff = useCallback((handoffId: string) => {
    controls.consumeHandoff("evidence_review", handoffId);
  }, [controls]);
  const activeTask = WORKFLOW_REVIEW_TASKS.find((task) => task.surface === activeSurface);

  return (
    <section
      className="workflow-review-workspace"
      id={activeTask?.anchor}
      hidden={activeSurface === null}
      aria-labelledby="workflow-review-workspace-title"
      data-active-surface={activeSurface ?? "inactive"}
      data-application-active={String(context.applicationActive)}
    >
      <header className="workflow-review-heading">
        <div>
          <p className="eyebrow">{t($ => $.runReview.s6WorkflowRunAndEvaluationReview)}</p>
          <h3 id="workflow-review-workspace-title">{t($ => $.runReview.runAndEvaluationReview)}</h3>
          <p>{t($ => $.runReview.followExactRunEvidenceIntoCompatibleComparisonVersioned)}</p>
        </div>
        <dl>
          <div><dt>{t($ => $.runReview.application)}</dt><dd>{context.displayName}</dd></div>
          <div><dt>{t($ => $.runReview.workspace)}</dt><dd>{context.workspaceId || t($ => $.runReview.unavailable)}</dd></div>
          <div><dt>{t($ => $.runReview.lifecycle)}</dt><dd className={context.applicationActive ? "is-active" : "is-archived"}>{context.lifecycleState}</dd></div>
        </dl>
      </header>

      <div className="workflow-review-layout">
        <nav className="workflow-review-path" aria-label={t($ => $.runReview.workflowReviewTasks)}>
          <header><span>{t($ => $.runReview.reviewPath)}</span><strong>{t($ => $.runReview.oneOwnerAtATime)}</strong></header>
          {WORKFLOW_REVIEW_TASKS.map((task) => {
            const selected = task.surface === activeSurface;
            return (
              <a
                key={task.surface}
                className={`workflow-review-task ${selected ? "is-selected" : ""}`}
                href={`#${task.anchor}`}
                aria-current={selected ? "step" : undefined}
              >
                <i aria-hidden="true" />
                <b>{task.number}</b>
                <span><strong>{t($ => $.runReview.surfaces[task.surface].label)}</strong><small>{t($ => $.runReview.surfaces[task.surface].summary)}</small></span>
                {selected ? <em aria-hidden="true">›</em> : null}
              </a>
            );
          })}
          <p className="workflow-review-boundary">
            <span aria-hidden="true">!</span>
            {t($ => $.runReview.approvedIsAppendOnlyReviewEvidenceItNever)}</p>
        </nav>

        <main className="workflow-review-owner" data-owner={activeSurface ?? "inactive"}>
          <Suspense fallback={<WorkflowReviewFallback />}>
            {activeSurface ? (
              <WorkflowReviewOwner
                key={`${context.generationKey}:workflow-review`}
                applicationId={context.applicationId}
                workspaceId={context.workspaceId}
                applicationActive={context.applicationActive}
                activeSurface={activeSurface}
                refreshKey={refreshKey}
                handoffRunId={pendingRunHandoff?.refId}
                handoffId={pendingRunHandoff?.handoffId}
                onHandoffConsumed={consumeRunHandoff}
              />
            ) : null}
          </Suspense>
        </main>
      </div>
    </section>
  );
}

function WorkflowReviewFallback() {
  const { t } = useTranslation("evaluation");
  return <div className="workflow-review-loading" role="status">{t($ => $.runReview.loadingTheCurrentWorkflowReviewOwner)}</div>;
}
