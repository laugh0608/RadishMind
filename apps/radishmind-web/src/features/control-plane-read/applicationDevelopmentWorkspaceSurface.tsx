import "../../i18n/workspaceSurfaceResources.ts";
import { useTranslation } from "react-i18next";
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
const ApplicationEvaluationCampaignPanel = lazy(() => import("./applicationEvaluationCampaignPanel.tsx"));
const ApplicationInteractionSessionPanel = lazy(() => import("./applicationInteractionSessionPanel.tsx"));
const ApplicationPublishCandidatePanel = lazy(() => import("./applicationPublishCandidatePanel.tsx"));
const PromptAgentTypeWorkspace = lazy(() => import("./promptAgentTypeWorkspace.tsx"));
const ApplicationRAGInvocationPanel = lazy(() => import("./workflowRAGApplicationRuntimePanel.tsx"));
const WorkflowRAGEvaluationDatasetPanel = lazy(() => import("./workflowRAGEvaluationDatasetPanel.tsx"));
const WorkflowDefinitionPromotionPanel = lazy(() => import("./workflowDefinitionPromotionPanel.tsx"));
const WorkflowTemplateCatalogPanel = lazy(() => import("./workflowTemplateCatalogPanel.tsx"));
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
  onTemplateDerivedDraft: (
    draft: WorkflowDraftDesignerDraft,
    authority: {
      draftId: string;
      draftVersion: number;
      lifecycleVersion: number;
      lifecycleState: "active";
      targetApplicationId: string;
    },
  ) => void;
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
  onTemplateDerivedDraft,
  onRunRecorded,
}: Props) {
  const { t } = useTranslation("applications");
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
  const openConfigurationAttach = useCallback((candidateId: string) => {
    controls.issueHandoff({
      applicationId: context.applicationId,
      sourceStage: "human_promotion",
      refKind: "binding",
      refId: candidateId,
    });
  }, [context.applicationId, controls.issueHandoff]);
  const consumeConfigurationHandoff = useCallback((handoffId: string) => {
    controls.consumeHandoff("configure_build", handoffId);
  }, [controls.consumeHandoff]);
  const pendingDraftHandoff = controls.pendingHandoff?.targetStage === "human_promotion" &&
    controls.pendingHandoff.refKind === "draft"
    ? controls.pendingHandoff
    : null;
  const pendingBindingHandoff = controls.pendingHandoff?.targetStage === "configure_build" &&
    controls.pendingHandoff.refKind === "binding"
    ? controls.pendingHandoff
    : null;
  const readinessStatus = controls.readiness.status === "review_not_started"
    ? t($ => $.workspaceSurface.readinessNotStarted)
    : controls.readiness.status === "review_blocked"
      ? t($ => $.workspaceSurface.readinessBlocked)
      : controls.readiness.status === "review_incomplete"
        ? t($ => $.workspaceSurface.readinessIncomplete)
        : t($ => $.workspaceSurface.readinessReviewable);
  const readinessSummary = controls.readiness.status === "review_not_started"
    ? t($ => $.workspaceSurface.readinessNotStartedSummary)
    : controls.readiness.status === "review_blocked"
      ? t($ => $.workspaceSurface.readinessBlockedSummary, { count: controls.readiness.blockerCount })
      : controls.readiness.status === "review_incomplete"
        ? t($ => $.workspaceSurface.readinessIncompleteSummary, { count: controls.readiness.missingCount })
        : t($ => $.workspaceSurface.readinessReviewableSummary);

  if (!activeStage) {
    return (
      <article className="application-development-stage-paused" role="status">
        <p className="eyebrow">{t($ => $.workspaceSurface.developmentWorkspace)}</p>
        <h4>{t($ => $.workspaceSurface.stagePaused)}</h4>
        <p>
          {t($ => $.workspaceSurface.stagePausedDescription)}</p>
      </article>
    );
  }

  if (context.surfaceKind === "prompt_application" || context.surfaceKind === "agent_copilot") {
    if (activeStage === "release_readiness") {
      return (
        <article className="application-development-stage-paused" role="status">
          <p className="eyebrow">{t($ => $.workspaceSurface.releaseReadinessBoundary)}</p>
          <h4>{readinessStatus}</h4>
          <p>{readinessSummary} {t($ => $.workspaceSurface.readinessProjectionOnly)}</p>
        </article>
      );
    }
    return (
      <Suspense fallback={<StageFallback label={t($ => $.workspaceSurface.promptAgentTypeWorkspace)} />}>
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
        <StageSurface stage="configure_build" title={t($ => $.workspaceSurface.configureBuild)}>
          {context.surfaceKind === "workflow_rag" ? (
            <Suspense fallback={<StageFallback label={t($ => $.workspaceSurface.knowledgeSnapshots)} />}>
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
                    ? t($ => $.workspaceSurface.archivedConfigurationHistory)
                    : t($ => $.workspaceSurface.applicationConfigurationDraft)}
                />
              )}
            >
              <ApplicationConfigurationDraftPanel
                key={`${context.generationKey}:configuration:${context.status}`}
                readOnly={context.status === "archived"}
                baseline={baseline}
                handoffBindingCandidateId={pendingBindingHandoff?.refId}
                handoffId={pendingBindingHandoff?.handoffId}
                onHandoffConsumed={consumeConfigurationHandoff}
                onEvidenceChange={reportOwnerEvidence}
                onOpenPublishReview={openPublishReview}
              />
            </Suspense>
          )}
        </StageSurface>
      ) : null}

      {activeStage === "human_promotion" ? (
        <StageSurface stage="human_promotion" title={t($ => $.workspaceSurface.humanPromotion)}>
          {context.surfaceKind === "workflow_rag" ? (
            <Suspense fallback={<StageFallback label={t($ => $.workspaceSurface.workflowRagPromotion)} />}>
              <WorkflowRAGPromotionPanel
                key={`${context.generationKey}:rag-promotion`}
                applicationId={context.applicationId}
                applicationName={context.displayName}
                applicationActive={context.applicationActive}
                onEvidenceChange={reportOwnerEvidence}
                onOpenConfigurationAttach={openConfigurationAttach}
              />
            </Suspense>
          ) : null}
          {context.status === "unavailable" ? (
            <UnavailableApplication />
          ) : (
            <Suspense
              fallback={(
                <StageFallback
                  label={context.status === "archived" ? t($ => $.workspaceSurface.archivedPublishHistory) : t($ => $.workspaceSurface.applicationPublishReview)}
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
            <Suspense fallback={<StageFallback label={t($ => $.workspaceSurface.workflowTemplateCatalog)} />}>
              <WorkflowTemplateCatalogPanel
                key={`${context.generationKey}:workflow-template-catalog`}
                workspaceId={context.workspaceId}
                applicationId={context.applicationId}
                applicationName={context.displayName}
                applicationActive={context.applicationActive}
                onDerivedDraft={onTemplateDerivedDraft}
              />
            </Suspense>
          ) : null}
          {context.applicationActive && context.surfaceKind === "workflow_rag" ? (
            <Suspense fallback={<StageFallback label={t($ => $.workspaceSurface.workflowDefinitionPromotion)} />}>
              <WorkflowDefinitionPromotionPanel
                key={`${context.generationKey}:workflow-definition-promotion`}
                workspaceId={context.workspaceId}
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
        <StageSurface stage="controlled_test" title={t($ => $.workspaceSurface.controlledTest)}>
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
                    <span><strong>{t($ => $.workspaceSurface.applicationSpecificControlledTests)}</strong><small>{t($ => $.workspaceSurface.existingOwnerSurfaces)}</small></span>
                    <span aria-hidden="true">⌄</span>
                  </summary>
                  <div>
                    {context.surfaceKind === "workflow_rag" ? (
                      <Suspense fallback={<StageFallback label={t($ => $.workspaceSurface.applicationInteraction)} />}>
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
                      <Suspense fallback={<StageFallback label={t($ => $.workspaceSurface.applicationRagInvocation)} />}>
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
        <StageSurface stage="evidence_review" title={t($ => $.workspaceSurface.runEvaluationReview)}>
          <Suspense fallback={<StageFallback label={t($ => $.workspaceSurface.applicationEvaluationCampaign)} />}>
            <ApplicationEvaluationCampaignPanel
              key={`${context.generationKey}:application-evaluation`}
              applicationId={context.applicationId}
              applicationName={context.displayName}
              applicationKind={context.applicationKind}
              workspaceId={context.workspaceId}
              applicationActive={context.applicationActive}
            />
          </Suspense>
          {context.surfaceKind === "workflow_rag" ? (
            <Suspense fallback={<StageFallback label={t($ => $.workspaceSurface.workflowRagDatasets)} />}>
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
            <p className="eyebrow">{t($ => $.workspaceSurface.workflowReviewPath)}</p>
            <h4>{t($ => $.workspaceSurface.taskScopedReviewOwners)}</h4>
            <p>{t($ => $.workspaceSurface.reviewOwnerInstruction)}</p>
          </article>
        </StageSurface>
      ) : null}

      {activeStage === "release_readiness" ? (
        <article className="application-development-stage-paused" role="status">
          <p className="eyebrow">{t($ => $.workspaceSurface.releaseReadinessBoundary)}</p>
          <h4>{readinessStatus}</h4>
          <p>{readinessSummary} {t($ => $.workspaceSurface.readinessProjectionOnly)}</p>
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
  const { t } = useTranslation("applications");
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
          <p className="eyebrow">{t($ => $.workspaceSurface.applicationAccess)}</p>
          <h3 id="application-access-workspace-title">{t($ => $.workspaceSurface.apiIntegrationKeys)}</h3>
          <p>{t($ => $.workspaceSurface.accessInstruction)}</p>
        </div>
        <div className="application-access-boundaries" aria-label={t($ => $.workspaceSurface.accessBoundaries)}>
          <span className="status-badge neutral">{t($ => $.workspaceSurface.devTest)}</span>
          <span className={`status-badge ${context.applicationActive ? "good" : "neutral"}`}>
            {context.applicationActive ? t($ => $.workspaceSurface.activeApplication) : t($ => $.workspaceSurface.archivedReadRevoke)}
          </span>
        </div>
      </header>

      <div className="application-access-workbench">
        <aside className="application-access-rail" aria-label={t($ => $.workspaceSurface.accessTasks)}>
          <p>{t($ => $.workspaceSurface.accessPath)}</p>
          <nav>
            {context.applicationActive ? (
              <a href="#application-api-integration" aria-current={activeSurface === "integration" ? "step" : undefined}>
                <b>01</b><span><strong>{t($ => $.workspaceSurface.connectApi)}</strong><small>{t($ => $.workspaceSurface.modelProtocolExample)}</small></span>
              </a>
            ) : (
              <span className="is-disabled"><b>01</b><span><strong>{t($ => $.workspaceSurface.connectApi)}</strong><small>{t($ => $.workspaceSurface.archivedAccessBlocked)}</small></span></span>
            )}
            <a href="#workspace-api-keys" aria-current={activeSurface === "credentials" ? "step" : undefined}>
              <b>02</b><span><strong>{t($ => $.workspaceSurface.credentials)}</strong><small>{t($ => $.workspaceSurface.credentialActions)}</small></span>
            </a>
            {context.applicationActive ? (
              <a href="#model-gateway-playground">
                <b>03</b><span><strong>{t($ => $.workspaceSurface.validate)}</strong><small>{t($ => $.workspaceSurface.existingPlayground)}</small></span>
              </a>
            ) : (
              <span className="is-disabled"><b>03</b><span><strong>{t($ => $.workspaceSurface.validate)}</strong><small>{t($ => $.workspaceSurface.invocationBlocked)}</small></span></span>
            )}
            <span><b>04</b><span><strong>{t($ => $.workspaceSurface.verifyRetire)}</strong><small>{t($ => $.workspaceSurface.exactGatewayEvidence)}</small></span></span>
          </nav>
          <p className="boundary-note">{t($ => $.workspaceSurface.accessPathBoundary)}</p>
        </aside>

        <main className="application-access-main">
          {!context.applicationActive ? (
            <div className="application-access-archived-boundary" role="status">
              <strong>{t($ => $.workspaceSurface.archivedApplicationBoundary)}</strong>
              <span>{t($ => $.workspaceSurface.archivedAccessBoundary)}</span>
            </div>
          ) : null}
          <div className="application-access-owner" hidden={activeSurface !== "integration"}>
            <Suspense fallback={<StageFallback label={t($ => $.workspaceSurface.applicationApiIntegration)} />}>
              <ApplicationApiIntegrationPanel
                key={`${context.generationKey}:api-integration`}
                applicationId={context.applicationId}
                applicationName={context.displayName}
                workspaceId={context.workspaceId}
              />
            </Suspense>
          </div>
          <div className="application-access-owner" hidden={activeSurface !== "credentials"}>
            <Suspense fallback={<StageFallback label={t($ => $.workspaceSurface.apiKeyLifecycle)} />}>
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
  const { t } = useTranslation("applications");
  return (
    <section
      className="application-development-stage-surface"
      data-application-development-stage={stage}
      aria-label={title}
    >
      <div className="application-development-stage-heading">
        <p className="eyebrow">{t($ => $.workspaceSurface.applicationDevelopmentStage)}</p>
        <h4>{title}</h4>
      </div>
      {children}
    </section>
  );
}

function StageFallback({ label }: { label: string }) {
  const { t } = useTranslation("applications");
  return (
    <div className="application-development-stage-fallback">
      <p>{t($ => $.workspaceSurface.loadingOwner, { label })}</p>
    </div>
  );
}

function UnavailableApplication() {
  const { t } = useTranslation("applications");
  return (
    <article className="application-catalog-downstream-blocked" role="status">
      <p className="eyebrow">{t($ => $.workspaceSurface.lifecycleEnforcement)}</p>
      <h4>{t($ => $.workspaceSurface.selectActiveApplication)}</h4>
      <p>{t($ => $.workspaceSurface.scopeUnavailable)}</p>
    </article>
  );
}

function ControlledTestBlocked({ status }: { status: ApplicationDevelopmentWorkspaceContext["status"] }) {
  const { t } = useTranslation("applications");
  return (
    <article className="application-catalog-downstream-blocked" role="status">
      <p className="eyebrow">{t($ => $.workspaceSurface.lifecycleEnforcement)}</p>
      <h4>{t($ => $.workspaceSurface.controlledTestsBlocked)}</h4>
      <p>
        {status === "archived"
          ? t($ => $.workspaceSurface.archivedTestsBlocked)
          : t($ => $.workspaceSurface.selectActiveForInvocation)}
      </p>
    </article>
  );
}
