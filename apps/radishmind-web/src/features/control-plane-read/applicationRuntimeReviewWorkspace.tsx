import "../../i18n/runtimeReviewResources.ts";
import { gatewayReviewState } from "./gatewayReviewMessages.ts";
import "../../i18n/gatewayReviewResources.ts";
import { useTranslation } from "react-i18next";
import { lazy, Suspense, useCallback, useEffect, useState } from "react";

import {
  applicationRuntimeReviewSurfaceForHash,
  type ApplicationDevelopmentWorkspaceContext,
  type ApplicationRuntimeReviewSurface,
} from "./applicationDevelopmentWorkspace.ts";
import type { ApplicationDevelopmentWorkspaceControls } from "./applicationDevelopmentWorkspaceControls.ts";
import type { ApplicationDevelopmentOwnerEvidence } from "./applicationDevelopmentReadiness.ts";
import { requestGatewayRequestHistoryReview } from "./modelGatewayPlaygroundEvents.ts";

const ModelGatewayPlaygroundPanel = lazy(() => import("./modelGatewayPlaygroundPanel.tsx"));
const ModelGatewayRequestHistoryPanel = lazy(() => import("./modelGatewayRequestHistoryPanel.tsx"));
const ApplicationOperationsPanel = lazy(() => import("./applicationOperationsPanel.tsx"));
const ApplicationResultArtifactLibraryPanel = lazy(() => import("./applicationResultArtifactLibraryPanel.tsx"));

const RUNTIME_REVIEW_TASKS: ReadonlyArray<{
  surface: ApplicationRuntimeReviewSurface;
  anchor: string;
  number: string;
}> = [
  {
    surface: "run",
    anchor: "model-gateway-playground",
    number: "01",
  },
  {
    surface: "request",
    anchor: "model-gateway-request-history",
    number: "02",
  },
  {
    surface: "evidence",
    anchor: "application-operations",
    number: "03",
  },
  {
    surface: "results",
    anchor: "application-result-artifact-library",
    number: "04",
  },
];

export default function ApplicationRuntimeReviewWorkspace({
  context,
  surfaceKey,
  controls,
}: {
  context: ApplicationDevelopmentWorkspaceContext;
  surfaceKey: string;
  controls: ApplicationDevelopmentWorkspaceControls;
}) {
  const { t } = useTranslation("gateway");
  const [activeSurface, setActiveSurface] = useState<ApplicationRuntimeReviewSurface | null>(() => (
    applicationRuntimeReviewSurfaceForHash(window.location.hash)
  ));

  useEffect(() => {
    function synchronizeSurface() {
      const nextSurface = applicationRuntimeReviewSurfaceForHash(window.location.hash);
      setActiveSurface(nextSurface);
      if (nextSurface) {
        window.requestAnimationFrame(() => {
          document.querySelector<HTMLElement>(".application-runtime-review-workspace")
            ?.scrollIntoView({ block: "start" });
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
        document.querySelector<HTMLElement>(".application-runtime-review-workspace")
          ?.scrollIntoView({ block: "start" });
      });
    });
    return () => {
      window.cancelAnimationFrame(firstFrame);
      if (secondFrame) window.cancelAnimationFrame(secondFrame);
    };
  }, [activeSurface]);

  const reportOperationsEvidence = useCallback((evidence: ApplicationDevelopmentOwnerEvidence) => {
    controls.reportEvidence({
      ...evidence,
      applicationId: context.applicationId,
      workspaceGenerationKey: context.generationKey,
      surfaceKey,
    });
  }, [context.applicationId, context.generationKey, controls.reportEvidence, surfaceKey]);

  const openGatewayRequest = useCallback((requestId: string, consumerRef: string) => {
    requestGatewayRequestHistoryReview(requestId, context.applicationId, consumerRef);
    window.location.hash = "model-gateway-request-history";
  }, [context.applicationId]);

  const openWorkflowRun = useCallback((runId: string) => {
    controls.issueHandoff({
      applicationId: context.applicationId,
      sourceStage: "evidence_review",
      refKind: "run",
      refId: runId,
    });
  }, [context.applicationId, controls.issueHandoff]);

  return (
    <section
      className="application-runtime-review-workspace"
      hidden={activeSurface === null}
      aria-labelledby="application-runtime-review-title"
      data-active-surface={activeSurface ?? "inactive"}
      data-application-active={String(context.applicationActive)}
    >
      <header className="application-runtime-review-heading">
        <div>
          <p className="eyebrow">{t($ => $.runtimeReview.eyebrow)}</p>
          <h3 id="application-runtime-review-title">{t($ => $.runtimeReview.title)}</h3>
          <p>{t($ => $.runtimeReview.intro)}</p>
        </div>
        <dl>
          <div><dt>{t($ => $.runtimeReview.application)}</dt><dd>{context.displayName}</dd></div>
          <div><dt>{t($ => $.runtimeReview.workspace)}</dt><dd>{context.workspaceId || t($ => $.runtimeReview.unavailable)}</dd></div>
          <div><dt>{t($ => $.runtimeReview.lifecycle)}</dt><dd className={context.applicationActive ? "is-active" : "is-archived"}>{gatewayReviewState(t, context.lifecycleState)}</dd></div>
        </dl>
      </header>

      <div className="application-runtime-review-layout">
        <nav className="application-runtime-review-path" aria-label={t($ => $.runtimeReview.tasks)}>
          <header><span>{t($ => $.runtimeReview.path)}</span><strong>{t($ => $.runtimeReview.oneTask)}</strong></header>
          {RUNTIME_REVIEW_TASKS.map((task) => {
            const selected = task.surface === activeSurface;
            const blocked = task.surface === "run" && !context.applicationActive;
            const content = (
              <>
                <i aria-hidden="true" />
                <b>{task.number}</b>
                <span><strong>{t($ => $.runtimeReview.tasksBySurface[task.surface].label)}</strong><small>{blocked ? t($ => $.runtimeReview.archived) : t($ => $.runtimeReview.tasksBySurface[task.surface].summary)}</small></span>
                {selected ? <em aria-hidden="true">›</em> : null}
              </>
            );
            return blocked ? (
              <span key={task.surface} className="application-runtime-review-task is-blocked" aria-disabled="true">{content}</span>
            ) : (
              <a
                key={task.surface}
                className={`application-runtime-review-task ${selected ? "is-selected" : ""}`}
                href={`#${task.anchor}`}
                aria-current={selected ? "step" : undefined}
              >
                {content}
              </a>
            );
          })}
          <p className="application-runtime-review-boundary">
            <span aria-hidden="true">!</span>
            {t($ => $.runtimeReview.boundary)}
          </p>
        </nav>

        <div className="application-runtime-review-owner" data-owner={activeSurface ?? "inactive"}>
          <Suspense fallback={<RuntimeOwnerFallback />}>
            <div hidden={activeSurface !== "run"} className="application-runtime-review-owner-panel">
              <ModelGatewayPlaygroundPanel
                selectedApplicationId={context.applicationId}
                workspaceId={context.workspaceId}
                applicationActive={context.applicationActive}
                active={activeSurface === "run"}
              />
            </div>
            <div hidden={activeSurface !== "request"} className="application-runtime-review-owner-panel">
              <ModelGatewayRequestHistoryPanel
                selectedApplicationId={context.applicationId}
                workspaceId={context.workspaceId}
                active={activeSurface === "request"}
              />
            </div>
            <div hidden={activeSurface !== "evidence"} className="application-runtime-review-owner-panel">
              <ApplicationOperationsPanel
                applicationId={context.applicationId}
                applicationName={context.displayName}
                workspaceId={context.workspaceId}
                active={activeSurface === "evidence"}
                onEvidenceChange={reportOperationsEvidence}
                onOpenGatewayRequest={openGatewayRequest}
                onOpenWorkflowRun={openWorkflowRun}
              />
            </div>
            <div hidden={activeSurface !== "results"} className="application-runtime-review-owner-panel">
              <ApplicationResultArtifactLibraryPanel
                applicationId={context.applicationId}
                applicationName={context.displayName}
                active={activeSurface === "results"}
                onOpenRun={openWorkflowRun}
              />
            </div>
          </Suspense>
        </div>
      </div>
    </section>
  );
}

function RuntimeOwnerFallback() {
  const { t } = useTranslation("gateway");
  return (
    <div className="application-runtime-review-loading" role="status">
      {t($ => $.runtimeReview.loading)}
    </div>
  );
}
