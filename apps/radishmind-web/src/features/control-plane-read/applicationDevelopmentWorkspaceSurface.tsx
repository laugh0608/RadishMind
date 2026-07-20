import { lazy, Suspense, useCallback, type ReactNode } from "react";

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
const ApplicationRAGInvocationPanel = lazy(() => import("./workflowRAGApplicationRuntimePanel.tsx"));
const ApplicationOperationsPanel = lazy(() => import("./applicationOperationsPanel.tsx"));
const WorkflowRAGEvaluationDatasetPanel = lazy(() => import("./workflowRAGEvaluationDatasetPanel.tsx"));
const WorkflowDefinitionPromotionPanel = lazy(() => import("./workflowDefinitionPromotionPanel.tsx"));
const WorkflowRAGPromotionPanel = lazy(() => import("./workflowRAGPromotionPanel.tsx"));
const WorkflowRAGSnapshotPanel = lazy(() => import("./workflowRAGSnapshotPanel.tsx"));
const WorkflowRunHistoryPanel = lazy(() => import("./workflowRunHistoryPanel.tsx"));

type Props = {
  context: ApplicationDevelopmentWorkspaceContext;
  activeStage: ApplicationDevelopmentStageId | null;
  surfaceKey: string;
  controls: ApplicationDevelopmentWorkspaceControls;
  offlineApiKeys: WorkspaceApiKeysViewModel;
  suggestedDefinitionId: string;
  runHistoryRefreshKey: number;
  activeWorkflowDraft: WorkflowDraftDesignerDraft;
  savedDraftVersion: number;
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
  runHistoryRefreshKey,
  activeWorkflowDraft,
  savedDraftVersion,
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
  const consumeEvidenceHandoff = useCallback((handoffId: string) => {
    controls.consumeHandoff("evidence_review", handoffId);
  }, [controls.consumeHandoff]);
  const pendingRunHandoff = controls.pendingHandoff?.targetStage === "evidence_review" &&
    controls.pendingHandoff.refKind === "run"
    ? controls.pendingHandoff
    : null;
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

  return (
    <div className="application-development-stage-surfaces">
      {activeStage === "configure_build" ? (
        <StageSurface stage="configure_build" title="Configure and build">
          <Suspense fallback={<StageFallback label="application knowledge snapshots" />}>
            <WorkflowRAGSnapshotPanel
              key={`${context.generationKey}:rag-snapshot`}
              applicationId={context.applicationId}
              applicationName={context.displayName}
              applicationActive={context.applicationActive}
            />
          </Suspense>
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
          <Suspense fallback={<StageFallback label="Workflow RAG promotion and binding review" />}>
            <WorkflowRAGPromotionPanel
              key={`${context.generationKey}:rag-promotion`}
              applicationId={context.applicationId}
              applicationName={context.displayName}
              applicationActive={context.applicationActive}
              onEvidenceChange={reportOwnerEvidence}
            />
          </Suspense>
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
          {context.applicationActive ? (
            <Suspense fallback={<StageFallback label="Workflow Definition promotion" />}>
              <WorkflowDefinitionPromotionPanel
                key={`${context.generationKey}:workflow-definition-promotion`}
                applicationId={context.applicationId}
                activeDraft={activeWorkflowDraft}
                savedDraftVersion={savedDraftVersion}
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
          {context.applicationActive ? (
            <>
              <Suspense fallback={<StageFallback label="Application API Integration" />}>
                <ApplicationApiIntegrationPanel
                  key={`${context.generationKey}:api-integration`}
                  applicationId={context.applicationId}
                  applicationName={context.displayName}
                />
              </Suspense>
              <Suspense fallback={<StageFallback label="API key lifecycle" />}>
                <APIKeyLifecyclePanel
                  key={`${context.generationKey}:api-key`}
                  applicationId={context.applicationId}
                  applicationName={context.displayName}
                  applicationActive={context.applicationActive}
                  offlineView={offlineApiKeys}
                />
              </Suspense>
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
            </>
          ) : (
            <ControlledTestBlocked status={context.status} />
          )}
        </StageSurface>
      ) : null}

      {activeStage === "evidence_review" ? (
        <StageSurface stage="evidence_review" title="Run and evaluation review">
          <Suspense fallback={<StageFallback label="Workflow RAG evaluation datasets" />}>
            <WorkflowRAGEvaluationDatasetPanel
              key={`${context.generationKey}:rag-evaluation`}
              applicationId={context.applicationId}
              applicationName={context.displayName}
              applicationActive={context.applicationActive}
              onEvidenceChange={reportOwnerEvidence}
            />
          </Suspense>
          <Suspense fallback={<StageFallback label="application operations" />}>
            <ApplicationOperationsPanel
              key={`${context.generationKey}:operations`}
              applicationId={context.applicationId}
              applicationName={context.displayName}
              onEvidenceChange={reportOwnerEvidence}
            />
          </Suspense>
          <Suspense fallback={<StageFallback label="run history" />}>
            <WorkflowRunHistoryPanel
              key={`${context.generationKey}:run-history`}
              applicationId={context.applicationId}
              refreshKey={runHistoryRefreshKey}
              handoffRunId={pendingRunHandoff?.refId}
              handoffId={pendingRunHandoff?.handoffId}
              onHandoffConsumed={consumeEvidenceHandoff}
            />
          </Suspense>
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
