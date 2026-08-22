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
  label: string;
  summary: string;
  number: string;
}> = [
  {
    surface: "run",
    anchor: "model-gateway-playground",
    label: "Run request",
    summary: "Temporary input and output",
    number: "01",
  },
  {
    surface: "request",
    anchor: "model-gateway-request-history",
    label: "Review request",
    summary: "Exact sanitized record",
    number: "02",
  },
  {
    surface: "evidence",
    anchor: "application-operations",
    label: "Application evidence",
    summary: "Current loaded windows",
    number: "03",
  },
  {
    surface: "results",
    anchor: "application-result-artifact-library",
    label: "Saved results",
    summary: "Cross-Session exact artifacts",
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
          <p className="eyebrow">S5 · Application runtime review</p>
          <h3 id="application-runtime-review-title">Run, review, and observe</h3>
          <p>Move from one controlled request to its sanitized record, then inspect the current application windows without implying correlation.</p>
        </div>
        <dl>
          <div><dt>Application</dt><dd>{context.displayName}</dd></div>
          <div><dt>Workspace</dt><dd>{context.workspaceId || "unavailable"}</dd></div>
          <div><dt>Lifecycle</dt><dd className={context.applicationActive ? "is-active" : "is-archived"}>{context.lifecycleState}</dd></div>
        </dl>
      </header>

      <div className="application-runtime-review-layout">
        <nav className="application-runtime-review-path" aria-label="Application runtime review tasks">
          <header><span>Runtime path</span><strong>One task at a time</strong></header>
          {RUNTIME_REVIEW_TASKS.map((task) => {
            const selected = task.surface === activeSurface;
            const blocked = task.surface === "run" && !context.applicationActive;
            const content = (
              <>
                <i aria-hidden="true" />
                <b>{task.number}</b>
                <span><strong>{task.label}</strong><small>{blocked ? "archived · read only" : task.summary}</small></span>
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
            Request input and output are temporary. History is sanitized. Operations combines independent current windows only. Saved results are explicit owner-scoped artifacts.
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
  return (
    <div className="application-runtime-review-loading" role="status">
      Loading the current runtime owner…
    </div>
  );
}
