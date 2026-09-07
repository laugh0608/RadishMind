import { lazy, Suspense, useEffect, useMemo, useState, type ReactNode } from "react";

import type { ApplicationConfigurationBaseline } from "./applicationConfigurationDraftConsumer.ts";
import type { ApplicationPublishCandidate } from "./applicationPublishCandidateConsumer.ts";
import type {
  ApplicationDevelopmentStageId,
  ApplicationDevelopmentWorkspaceContext,
} from "./applicationDevelopmentWorkspace.ts";
import type { ApplicationDevelopmentOwnerEvidence } from "./applicationDevelopmentReadiness.ts";
import {
  promptAgentTypeWorkspaceSurfaceForHash,
  promptAgentTypeWorkspaceTasks,
  type PromptAgentTypeWorkspaceSurface,
} from "./promptAgentTypeWorkspaceModel.ts";

const AgentCopilotProfilePanel = lazy(() => import("./agentCopilotProfilePanel.tsx"));
const AgentCopilotRuntimePanel = lazy(() => import("./agentCopilotRuntimePanel.tsx"));
const AgentCopilotSessionPanel = lazy(() => import("./agentCopilotSessionPanel.tsx"));
const ApplicationConfigurationDraftPanel = lazy(() => import("./applicationConfigurationDraftPanel.tsx"));
const ApplicationEvaluationCampaignPanel = lazy(() => import("./applicationEvaluationCampaignPanel.tsx"));
const ApplicationPublishCandidatePanel = lazy(() => import("./applicationPublishCandidatePanel.tsx"));
const PromptApplicationInvocationPanel = lazy(() => import("./promptApplicationInvocationPanel.tsx"));
const PromptApplicationRuntimePanel = lazy(() => import("./promptApplicationRuntimePanel.tsx"));
const PromptApplicationSessionPanel = lazy(() => import("./promptApplicationSessionPanel.tsx"));
const PromptApplicationTemplatePanel = lazy(() => import("./promptApplicationTemplatePanel.tsx"));

export default function PromptAgentTypeWorkspace({
  context,
  activeStage,
  baseline,
  accessSurface,
  handoffDraftId,
  handoffId,
  onHandoffConsumed,
  onEvidenceChange,
  onOpenPublishReview,
  onRunRecorded,
  onOpenRun,
}: {
  context: ApplicationDevelopmentWorkspaceContext;
  activeStage: ApplicationDevelopmentStageId;
  baseline: ApplicationConfigurationBaseline;
  accessSurface: ReactNode;
  handoffDraftId?: string;
  handoffId?: string;
  onHandoffConsumed?: (handoffId: string) => void;
  onEvidenceChange?: (evidence: ApplicationDevelopmentOwnerEvidence) => void;
  onOpenPublishReview?: (draftId: string) => void;
  onRunRecorded?: (runId: string) => void;
  onOpenRun?: (runId: string) => void;
}) {
  const [activeSurface, setActiveSurface] = useState<PromptAgentTypeWorkspaceSurface | null>(() => (
    promptAgentTypeWorkspaceSurfaceForHash(window.location.hash, context.surfaceKind, activeStage)
  ));
  const [selectedCandidate, setSelectedCandidate] = useState<ApplicationPublishCandidate | null>(null);
  const tasks = useMemo(
    () => promptAgentTypeWorkspaceTasks(context.surfaceKind, context.status),
    [context.status, context.surfaceKind],
  );

  useEffect(() => {
    function synchronizeSurface() {
      setActiveSurface(promptAgentTypeWorkspaceSurfaceForHash(window.location.hash, context.surfaceKind, activeStage));
    }
    synchronizeSurface();
    window.addEventListener("hashchange", synchronizeSurface);
    return () => window.removeEventListener("hashchange", synchronizeSurface);
  }, [activeStage, context.surfaceKind]);

  useEffect(() => {
    setSelectedCandidate(null);
  }, [context.generationKey]);

  if (!activeSurface) return null;
  const currentTask = tasks.find((task) => task.surface === activeSurface) ?? null;
  const typeLabel = context.surfaceKind === "prompt_application" ? "Prompt Application" : "Agent / Copilot";
  const sourceLabel = context.surfaceKind === "prompt_application" ? "Template" : "Profile";
  const runProfile = context.surfaceKind === "prompt_application"
    ? "Run v6 · Comparison v5"
    : "Run v7 · Comparison v6";
  const ownerBlocked = currentTask?.availability === "blocked";

  return (
    <section
      className="prompt-agent-type-workspace"
      aria-labelledby="prompt-agent-type-workspace-title"
      data-active-surface={activeSurface}
      data-application-kind={context.surfaceKind}
      data-application-active={String(context.applicationActive)}
    >
      <header className="prompt-agent-type-heading">
        <div>
          <p className="eyebrow">S8 · Type workspace · dev / test</p>
          <h3 id="prompt-agent-type-workspace-title">{typeLabel} workspace</h3>
          <p>Move through exact {sourceLabel} authority, human governance, controlled use, and evaluation without duplicating an owner.</p>
        </div>
        <dl>
          <div><dt>Application</dt><dd>{context.displayName}</dd></div>
          <div><dt>Source</dt><dd>{sourceLabel} owner</dd></div>
          <div><dt>Evidence</dt><dd>{runProfile}</dd></div>
          <div><dt>Lifecycle</dt><dd className={context.applicationActive ? "is-active" : "is-archived"}>{context.lifecycleState}</dd></div>
        </dl>
      </header>

      <div className="prompt-agent-type-layout">
        <nav className="prompt-agent-type-path" aria-label={`${typeLabel} tasks`}>
          <header><span>Type path</span><strong>One owner at a time</strong></header>
          {tasks.map((task) => {
            const selected = task.surface === activeSurface;
            const content = (
              <>
                <i aria-hidden="true" />
                <b>{task.number}</b>
                <span>
                  <strong>{task.label}</strong>
                  <small>{task.availability === "blocked" ? "archived · blocked" : task.availability === "read_only" ? "archived · read only" : task.summary}</small>
                </span>
                {selected ? <em aria-hidden="true">›</em> : null}
              </>
            );
            return task.availability === "blocked" ? (
              <span key={task.surface} className="prompt-agent-type-task is-blocked" aria-disabled="true">{content}</span>
            ) : (
              <a
                key={task.surface}
                className={`prompt-agent-type-task ${selected ? "is-selected" : ""}`}
                href={`#${task.anchor}`}
                aria-current={selected ? "step" : undefined}
              >
                {content}
              </a>
            );
          })}
          <p className="prompt-agent-type-boundary">
            <span aria-hidden="true">!</span>
            Approval never activates the next owner. Input and output remain volatile; production authorization stays closed.
          </p>
        </nav>

        <main className="prompt-agent-type-owner" data-owner={activeSurface}>
          {ownerBlocked ? (
            <BlockedTypeOwner typeLabel={typeLabel} />
          ) : (
            <Suspense key={context.generationKey} fallback={<TypeOwnerFallback />}>
              {activeSurface === "source" && context.surfaceKind === "prompt_application" ? (
                <PromptApplicationTemplatePanel
                  applicationId={context.applicationId}
                  applicationName={context.displayName}
                  applicationKind={context.applicationKind}
                  applicationActive={context.applicationActive}
                  onOpenPublishReview={onOpenPublishReview}
                  onEvidenceChange={onEvidenceChange}
                />
              ) : null}
              {activeSurface === "source" && context.surfaceKind === "agent_copilot" ? (
                <AgentCopilotProfilePanel
                  applicationId={context.applicationId}
                  applicationName={context.displayName}
                  applicationKind={context.applicationKind}
                  applicationActive={context.applicationActive}
                  onOpenPublishReview={onOpenPublishReview}
                  onEvidenceChange={onEvidenceChange}
                />
              ) : null}
              {activeSurface === "configuration" ? (
                <ApplicationConfigurationDraftPanel
                  readOnly={!context.applicationActive}
                  baseline={baseline}
                  onEvidenceChange={onEvidenceChange}
                  onOpenPublishReview={onOpenPublishReview}
                />
              ) : null}
              {activeSurface === "candidate" ? (
                <ApplicationPublishCandidatePanel
                  readOnly={!context.applicationActive}
                  baseline={baseline}
                  onEvidenceChange={onEvidenceChange}
                  handoffDraftId={handoffDraftId}
                  handoffId={handoffId}
                  onHandoffConsumed={onHandoffConsumed}
                  includeRuntimeAssignment={false}
                  onSelectedCandidateChange={setSelectedCandidate}
                />
              ) : null}
              {activeSurface === "assignment" && context.surfaceKind === "prompt_application" ? (
                <PromptApplicationRuntimePanel
                  applicationId={context.applicationId}
                  publishCandidateId={selectedCandidate?.schemaVersion === "application_publish_candidate.v3" ? selectedCandidate.candidateId : ""}
                  candidateApproved={selectedCandidate?.schemaVersion === "application_publish_candidate.v3" && selectedCandidate.candidateState === "approved"}
                  readOnly={!context.applicationActive}
                  onEvidenceChange={onEvidenceChange}
                />
              ) : null}
              {activeSurface === "assignment" && context.surfaceKind === "agent_copilot" ? (
                <AgentCopilotRuntimePanel
                  applicationId={context.applicationId}
                  applicationName={context.displayName}
                  applicationActive={context.applicationActive}
                  onEvidenceChange={onEvidenceChange}
                />
              ) : null}
              {activeSurface === "access" ? accessSurface : null}
              {activeSurface === "controlled_use" && context.surfaceKind === "prompt_application" ? (
                <PromptApplicationInvocationPanel
                  applicationId={context.applicationId}
                  applicationName={context.displayName}
                  onRunRecorded={onRunRecorded}
                  onOpenRun={onOpenRun}
                  onEvidenceChange={onEvidenceChange}
                />
              ) : null}
              {activeSurface === "controlled_use" && context.surfaceKind === "agent_copilot" ? (
                <AgentCopilotSessionPanel
                  applicationId={context.applicationId}
                  applicationName={context.displayName}
                  onRunRecorded={onRunRecorded}
                  onOpenRun={onOpenRun}
                  onEvidenceChange={onEvidenceChange}
                />
              ) : null}
              {activeSurface === "session" && context.surfaceKind === "prompt_application" ? (
                <PromptApplicationSessionPanel
                  applicationId={context.applicationId}
                  onRunRecorded={onRunRecorded}
                  onOpenRun={onOpenRun}
                />
              ) : null}
              {activeSurface === "evaluation" ? (
                <ApplicationEvaluationCampaignPanel
                  applicationId={context.applicationId}
                  applicationName={context.displayName}
                  applicationKind={context.applicationKind}
                  workspaceId={context.workspaceId}
                  applicationActive={context.applicationActive}
                />
              ) : null}
            </Suspense>
          )}
        </main>
      </div>
    </section>
  );
}

function BlockedTypeOwner({ typeLabel }: { typeLabel: string }) {
  return (
    <article className="prompt-agent-type-blocked" role="status">
      <p className="eyebrow">Lifecycle enforcement</p>
      <h4>{typeLabel} controlled use is blocked</h4>
      <p>Archived applications retain sanitized source and run evidence, but cannot start an invocation, Session, or provider side effect.</p>
    </article>
  );
}

function TypeOwnerFallback() {
  return <div className="prompt-agent-type-loading" role="status">Loading the current type owner…</div>;
}
