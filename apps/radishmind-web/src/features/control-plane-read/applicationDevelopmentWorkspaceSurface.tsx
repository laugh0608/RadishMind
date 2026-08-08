import { lazy, Suspense, useCallback, useEffect, useState, type ReactNode } from "react";

import type { WorkspaceApiKeysViewModel } from "./workspaceApiKeys.ts";
import type { WorkflowDraftDesignerDraft } from "./workflowDraftDesigner.ts";
import type {
  ApplicationDevelopmentStageId,
  ApplicationDevelopmentWorkspaceContext,
} from "./applicationDevelopmentWorkspace.ts";
import type { ApplicationDevelopmentWorkspaceControls } from "./applicationDevelopmentWorkspaceControls.ts";
import type { ApplicationDevelopmentOwnerEvidence } from "./applicationDevelopmentReadiness.ts";

const APIKeyLifecyclePanel = lazy(() =>
  import("./apiKeyLifecyclePanel.tsx").then((module) => ({ default: module.APIKeyLifecyclePanel })),
);
const ApplicationApiIntegrationPanel = lazy(() => import("./applicationApiIntegrationPanel.tsx"));
const ApplicationConfigurationDraftPanel = lazy(() => import("./applicationConfigurationDraftPanel.tsx"));
const ApplicationInteractionSessionPanel = lazy(() => import("./applicationInteractionSessionPanel.tsx"));
const ApplicationPublishCandidatePanel = lazy(() => import("./applicationPublishCandidatePanel.tsx"));
const PromptAgentTypeWorkspace = lazy(() => import("./promptAgentTypeWorkspace.tsx"));
const ApplicationRAGInvocationPanel = lazy(() => import("./workflowRAGApplicationRuntimePanel.tsx"));
const WorkflowRAGEvaluationDatasetPanel = lazy(() => import("./workflowRAGEvaluationDatasetPanel.tsx"));
const WorkflowDefinitionPromotionPanel = lazy(() => import("./workflowDefinitionPromotionPanel.tsx"));
const WorkflowRAGPromotionPanel = lazy(() => import("./workflowRAGPromotionPanel.tsx"));
const WorkflowRAGSnapshotPanel = lazy(() => import("./workflowRAGSnapshotPanel.tsx"));

type Props = {
  context: ApplicationDevelopmentWorkspaceContext;
  activeStage: ApplicationDevelopmentStageId | null;
  surfaceKey: string;
  controls: ApplicationDevelopmentWorkspaceControls;
  offlineApiKeys: WorkspaceApiKeysViewModel;
  suggestedDefinitionId: string;
  activeWorkflowDraft: WorkflowDraftDesignerDraft;
  savedDraftVersion: number;
  savedDraftLifecycleVersion: number;
  savedDraftLifecycleState: "active" | "archived" | "unknown";
  nextDerivedDraftNumber: number;
  onDerivedDraft: (draft: WorkflowDraftDesignerDraft) => void;
  onRunRecorded: () => void;
};

export default function ApplicationDevelopmentWorkspaceSurface({
  context,
  activeStage,
  surfaceKey,
  controls,
  offlineApiKeys,
  suggestedDefinitionId,
  activeWorkflowDraft,
  savedDraftVersion,
  savedDraftLifecycleVersion,
  savedDraftLifecycleState,
  nextDerivedDraftNumber,
  onDerivedDraft,
  onRunRecorded,
}: Props) {
  const baseline = {
    applicationId: context.applicationId,
    displayName: context.displayName,
    applicationKind: context.applicationKind,
    updatedAt: context.updatedAt,
  };
  const reportOwnerEvidence = useCallback((evidence: ApplicationDevelopmentOwnerEvidence) => {
    controls.reportEvidence({
      ...evidence,
      applicationId: context.applicationId,
      workspaceGenerationKey: context.generationKey,
      surfaceKey,
    });
  }, [context.applicationId, context.generationKey, controls.reportEvidence, surfaceKey]);
  const handleRunRecorded = useCallback((_runId: string) => {
    onRunRecorded();
  }, [onRunRecorded]);
  const openRunEvidence = useCallback((runId: string) => {
    if (!activeStage || !runId) return;
    controls.issueHandoff({
      applicationId: context.applicationId,
      sourceStage: activeStage,
      refKind: "run",
      refId: runId,
    });
  }, [activeStage, context.applicationId, controls.issueHandoff]);
  const openPublishReview = useCallback((draftId: string) => {
    controls.issueHandoff({
      applicationId: context.applicationId,
      sourceStage: "configure_build",
      refKind: "draft",
      refId: draftId,
    });
  }, [context.applicationId, controls.issueHandoff]);
  const consumePromotionHandoff = useCallback((handoffId: string) => {
    controls.consumeHandoff("human_promotion", handoffId);
  }, [controls.consumeHandoff]);
  const pendingDraftHandoff = controls.pendingHandoff?.targetStage === "human_promotion" &&
    controls.pendingHandoff.refKind === "draft"
    ? controls.pendingHandoff
    : null;

  if (!activeStage) {
    return (
      <article className="application-development-stage-paused" role="status">
        <p className="eyebrow">Application Development Workspace</p>
        <h4>Stage surface paused</h4>
        <p>
          Return to an Application Development stage to reload its owner-scoped metadata. Pending component
          state is not retained while another product route is active.
        </p>
      </article>
    );
  }

  if (context.surfaceKind === "prompt_application" || context.surfaceKind === "agent_copilot") {
    if (activeStage === "release_readiness") {
      return (
        <article className="application-development-stage-paused" role="status">
          <p className="eyebrow">Release readiness boundary</p>
          <h4>{controls.readiness.status}</h4>
          <p>{controls.readiness.summary} The source cards below remain the only readiness projection.</p>
        </article>
      );
    }
    return (
      <Suspense fallback={<StageFallback label="Prompt / Agent type workspace" />}>
        <PromptAgentTypeWorkspace
          context={context}
          activeStage={activeStage}
          baseline={baseline}
          accessSurface={<ApplicationAccessWorkspace context={context} offlineApiKeys={offlineApiKeys} />}
          handoffDraftId={pendingDraftHandoff?.refId}
          handoffId={pendingDraftHandoff?.handoffId}
          onHandoffConsumed={consumePromotionHandoff}
          onEvidenceChange={reportOwnerEvidence}
          onOpenPublishReview={openPublishReview}
          onRunRecorded={handleRunRecorded}
          onOpenRun={openRunEvidence}
        />
      </Suspense>
    );
  }

  return (
    <div className="application-development-stage-surfaces">
      {activeStage === "configure_build" ? (
        <StageSurface stage="configure_build" title="Configure and build">
          {context.surfaceKind === "workflow_rag" ? (
            <Suspense fallback={<StageFallback label="application knowledge snapshots" />}>
              <WorkflowRAGSnapshotPanel
                key={`${context.generationKey}:rag-snapshot`}
                applicationId={context.applicationId}
                applicationName={context.displayName}
                applicationActive={context.applicationActive}
              />
            </Suspense>
          ) : null}
          {context.status === "unavailable" ? (
            <UnavailableApplication />
          ) : (
            <Suspense
              fallback={(
                <StageFallback
                  label={context.status === "archived"
                    ? "archived configuration history"
                    : "Application Configuration Draft"}
                />
              )}
            >
              <ApplicationConfigurationDraftPanel
                key={`${context.generationKey}:configuration:${context.status}`}
                readOnly={context.status === "archived"}
                baseline={baseline}
                onEvidenceChange={reportOwnerEvidence}
                onOpenPublishReview={openPublishReview}
              />
            </Suspense>
          )}
        </StageSurface>
      ) : null}

      {activeStage === "human_promotion" ? (
        <StageSurface stage="human_promotion" title="Human promotion">
          {context.surfaceKind === "workflow_rag" ? (
            <Suspense fallback={<StageFallback label="Workflow RAG promotion and binding review" />}>
              <WorkflowRAGPromotionPanel
                key={`${context.generationKey}:rag-promotion`}
                applicationId={context.applicationId}
                applicationName={context.displayName}
                applicationActive={context.applicationActive}
                onEvidenceChange={reportOwnerEvidence}
              />
            </Suspense>
          ) : null}
          {context.status === "unavailable" ? (
            <UnavailableApplication />
          ) : (
            <Suspense
              fallback={(
                <StageFallback
                  label={context.status === "archived" ? "archived publish history" : "Application Publish Review"}
                />
              )}
            >
              <ApplicationPublishCandidatePanel
                key={`${context.generationKey}:publish:${context.status}`}
                readOnly={context.status === "archived"}
                baseline={baseline}
                onEvidenceChange={reportOwnerEvidence}
                handoffDraftId={pendingDraftHandoff?.refId}
                handoffId={pendingDraftHandoff?.handoffId}
                onHandoffConsumed={consumePromotionHandoff}
              />
            </Suspense>
          )}
          {context.applicationActive && context.surfaceKind === "workflow_rag" ? (
            <Suspense fallback={<StageFallback label="Workflow Definition promotion" />}>
              <WorkflowDefinitionPromotionPanel
                key={`${context.generationKey}:workflow-definition-promotion`}
                applicationId={context.applicationId}
                activeDraft={activeWorkflowDraft}
                savedDraftVersion={savedDraftVersion}
                savedDraftLifecycleVersion={savedDraftLifecycleVersion}
                savedDraftLifecycleState={savedDraftLifecycleState}
                nextDerivedDraftNumber={nextDerivedDraftNumber}
                onDerivedDraft={onDerivedDraft}
                onRunRecorded={handleRunRecorded}
                onOpenRun={openRunEvidence}
                onEvidenceChange={reportOwnerEvidence}
              />
            </Suspense>
          ) : null}
        </StageSurface>
      ) : null}

      {activeStage === "controlled_test" ? (
        <StageSurface stage="controlled_test" title="Controlled test">
          {context.status === "unavailable" ? (
            <ControlledTestBlocked status={context.status} />
          ) : (
            <>
              <ApplicationAccessWorkspace
                context={context}
                offlineApiKeys={offlineApiKeys}
              />
              {context.applicationActive ? (
                <details className="application-access-related-surfaces">
                  <summary>
                    <span><strong>Application-specific controlled tests</strong><small>Existing owner surfaces</small></span>
                    <span aria-hidden="true">⌄</span>
                  </summary>
                  <div>
                    {context.surfaceKind === "workflow_rag" ? (
                      <Suspense fallback={<StageFallback label="Application Interaction" />}>
                        <ApplicationInteractionSessionPanel
                          key={`${context.generationKey}:interaction`}
                          applicationId={context.applicationId}
                          applicationName={context.displayName}
                          applicationActive={context.applicationActive}
                          suggestedDefinitionId={suggestedDefinitionId}
                          onRunRecorded={handleRunRecorded}
                          onOpenRun={openRunEvidence}
                          onEvidenceChange={reportOwnerEvidence}
                        />
                      </Suspense>
                    ) : null}
                    {context.surfaceKind === "workflow_rag" ? (
                      <Suspense fallback={<StageFallback label="Application RAG Invocation" />}>
                        <ApplicationRAGInvocationPanel
                          key={`${context.generationKey}:rag-invocation`}
                          applicationId={context.applicationId}
                          applicationName={context.displayName}
                          applicationActive={context.applicationActive}
                          onRunRecorded={handleRunRecorded}
                          onOpenRun={openRunEvidence}
                          onEvidenceChange={reportOwnerEvidence}
                        />
                      </Suspense>
                    ) : null}
                  </div>
                </details>
              ) : null}
            </>
          )}
        </StageSurface>
      ) : null}

      {activeStage === "evidence_review" ? (
        <StageSurface stage="evidence_review" title="Run and evaluation review">
          {context.surfaceKind === "workflow_rag" ? (
            <Suspense fallback={<StageFallback label="Workflow RAG evaluation datasets" />}>
              <WorkflowRAGEvaluationDatasetPanel
                key={`${context.generationKey}:rag-evaluation`}
                applicationId={context.applicationId}
                applicationName={context.displayName}
                applicationActive={context.applicationActive}
                onEvidenceChange={reportOwnerEvidence}
              />
            </Suspense>
          ) : null}
          <article className="application-development-stage-paused">
            <p className="eyebrow">Workflow review path</p>
            <h4>Run and evaluation owners are task-scoped</h4>
            <p>Use Runs, Compare, Cases, or Release below. Only the selected owner is mounted.</p>
          </article>
        </StageSurface>
      ) : null}

      {activeStage === "release_readiness" ? (
        <article className="application-development-stage-paused" role="status">
          <p className="eyebrow">Release readiness boundary</p>
          <h4>{controls.readiness.status}</h4>
          <p>{controls.readiness.summary} The source cards below remain the only readiness projection.</p>
        </article>
      ) : null}
    </div>
  );
}

type ApplicationAccessSurface = "integration" | "credentials";

function ApplicationAccessWorkspace({
  context,
  offlineApiKeys,
}: {
  context: ApplicationDevelopmentWorkspaceContext;
  offlineApiKeys: WorkspaceApiKeysViewModel;
}) {
  const [activeSurface, setActiveSurface] = useState<ApplicationAccessSurface>(() => (
    accessSurfaceForHash(window.location.hash, context.applicationActive)
  ));

  useEffect(() => {
    function synchronizeSurface() {
      const activeHash = window.location.hash.trim();
      setActiveSurface(accessSurfaceForHash(activeHash, context.applicationActive));
      if (activeHash === "#application-api-integration" || activeHash === "#workspace-api-keys") {
        window.requestAnimationFrame(() => {
          document.querySelector<HTMLElement>(".application-access-workspace")?.scrollIntoView({ block: "start" });
        });
      }
    }
    synchronizeSurface();
    window.addEventListener("hashchange", synchronizeSurface);
    return () => window.removeEventListener("hashchange", synchronizeSurface);
  }, [context.applicationActive]);

  return (
    <section
      className="application-access-workspace"
      aria-labelledby="application-access-workspace-title"
      data-active-surface={activeSurface}
      data-application-active={String(context.applicationActive)}
    >
      <header className="application-access-heading">
        <div>
          <p className="eyebrow">S4 · Application access</p>
          <h3 id="application-access-workspace-title">API Integration &amp; Keys</h3>
          <p>Choose one access task, keep credentials scoped to this application, then validate through an existing controlled surface.</p>
        </div>
        <div className="application-access-boundaries" aria-label="Application access boundaries">
          <span className="status-badge neutral">dev / test</span>
          <span className={`status-badge ${context.applicationActive ? "good" : "neutral"}`}>
            {context.applicationActive ? "active application" : "archived · read / revoke"}
          </span>
        </div>
      </header>

      <div className="application-access-workbench">
        <aside className="application-access-rail" aria-label="Application access tasks">
          <p>Access path</p>
          <nav>
            {context.applicationActive ? (
              <a href="#application-api-integration" aria-current={activeSurface === "integration" ? "step" : undefined}>
                <b>01</b><span><strong>Connect API</strong><small>Model, protocol, example</small></span>
              </a>
            ) : (
              <span className="is-disabled"><b>01</b><span><strong>Connect API</strong><small>Blocked for archived application</small></span></span>
            )}
            <a href="#workspace-api-keys" aria-current={activeSurface === "credentials" ? "step" : undefined}>
              <b>02</b><span><strong>Credentials</strong><small>Issue, inspect, revoke</small></span>
            </a>
            {context.applicationActive ? (
              <a href="#model-gateway-playground">
                <b>03</b><span><strong>Validate</strong><small>Existing Playground</small></span>
              </a>
            ) : (
              <span className="is-disabled"><b>03</b><span><strong>Validate</strong><small>Invocation blocked</small></span></span>
            )}
            <span><b>04</b><span><strong>Verify / retire</strong><small>Exact Gateway evidence</small></span></span>
          </nav>
          <p className="boundary-note">The path shows task position only. It does not assert production readiness or automatic completion.</p>
        </aside>

        <main className="application-access-main">
          {!context.applicationActive ? (
            <div className="application-access-archived-boundary" role="status">
              <strong>Archived application boundary</strong>
              <span>Sanitized Key metadata remains readable and revocable. Issue, rotation, model loading, and invocation stay blocked.</span>
            </div>
          ) : null}
          <div className="application-access-owner" hidden={activeSurface !== "integration"}>
            <Suspense fallback={<StageFallback label="Application API Integration" />}>
              <ApplicationApiIntegrationPanel
                key={`${context.generationKey}:api-integration`}
                applicationId={context.applicationId}
                applicationName={context.displayName}
                workspaceId={context.workspaceId}
              />
            </Suspense>
          </div>
          <div className="application-access-owner" hidden={activeSurface !== "credentials"}>
            <Suspense fallback={<StageFallback label="API key lifecycle" />}>
              <APIKeyLifecyclePanel
                key={`${context.generationKey}:api-key`}
                applicationId={context.applicationId}
                applicationName={context.displayName}
                applicationActive={context.applicationActive}
                workspaceId={context.workspaceId}
                offlineView={offlineApiKeys}
              />
            </Suspense>
          </div>
        </main>
      </div>
    </section>
  );
}

function accessSurfaceForHash(hash: string, applicationActive: boolean): ApplicationAccessSurface {
  if (!applicationActive) return "credentials";
  return hash.trim() === "#workspace-api-keys" ? "credentials" : "integration";
}

function StageSurface({ stage, title, children }: { stage: string; title: string; children: ReactNode }) {
  return (
    <section
      className="application-development-stage-surface"
      data-application-development-stage={stage}
      aria-label={title}
    >
      <div className="application-development-stage-heading">
        <p className="eyebrow">Application Development Stage</p>
        <h4>{title}</h4>
      </div>
      {children}
    </section>
  );
}

function StageFallback({ label }: { label: string }) {
  return (
    <div className="application-development-stage-fallback">
      <p>Loading {label}…</p>
    </div>
  );
}

function UnavailableApplication() {
  return (
    <article className="application-catalog-downstream-blocked" role="status">
      <p className="eyebrow">Lifecycle enforcement</p>
      <h4>Create or select an active application</h4>
      <p>The authoritative Application scope is unavailable. Configuration and promotion remain blocked.</p>
    </article>
  );
}

function ControlledTestBlocked({ status }: { status: ApplicationDevelopmentWorkspaceContext["status"] }) {
  return (
    <article className="application-catalog-downstream-blocked" role="status">
      <p className="eyebrow">Lifecycle enforcement</p>
      <h4>Controlled tests are blocked</h4>
      <p>
        {status === "archived"
          ? "Archived applications retain read-only evidence but cannot start new controlled tests."
          : "Select an active Application before opening invocation handoffs."}
      </p>
    </article>
  );
}
