import { executorStateMessage, executorBlockerMessage } from "./workflowExecutorMessages.ts";
import "../../i18n/workflowDraftResources.ts";
import { workflowDraftStatusLabel } from "./workflowDraftMessages.ts";
import "../../i18n/workflowExecutorResources.ts";
import { useTranslation } from "react-i18next";
import type { WorkflowDraftDesignerDraft } from "./workflowDraftDesigner";
import type {
  WorkflowExecutorConsumerState,
  WorkflowExecutorEligibility,
} from "./workflowExecutorConsumer";
import type { WorkflowRunNodeRecord } from "./workflowRunRecordConsumer.ts";

export function WorkflowExecutorPanel({
  draft,
  consumerState,
  eligibility,
  inputText,
  model,
  conditionValues,
  onCreateExecutorDraft,
  onInputTextChange,
  onModelChange,
  onConditionValueChange,
  onStartRun,
  onReloadRun,
}: {
  draft: WorkflowDraftDesignerDraft;
  consumerState: WorkflowExecutorConsumerState;
  eligibility: WorkflowExecutorEligibility;
  inputText: string;
  model: string;
  conditionValues: Record<string, boolean>;
  onCreateExecutorDraft: () => void;
  onInputTextChange: (value: string) => void;
  onModelChange: (value: string) => void;
  onConditionValueChange: (nodeId: string, value: boolean) => void;
  onStartRun: () => void;
  onReloadRun: () => void;
}) {
  const { t } = useTranslation("workflow");
  const { t: shellT } = useTranslation("shell");
  const pending = consumerState.status === "starting" || consumerState.status === "reading";
  const canStart = consumerState.mode === "dev_workflow_executor_http" &&
    eligibility.eligible && inputText.trim().length > 0 && !pending;
  const record = consumerState.record;
  return (
    <section
      className="workflow-executor-v0"
      id="workflow-executor-v0"
      aria-labelledby="workflow-executor-v0-title"
    >
      <div className="section-heading compact-heading">
        <div>
          <p className="eyebrow">{t($ => $.executor.workflowExecutorV0)}</p>
          <h4 id="workflow-executor-v0-title">{t($ => $.executor.boundedDevelopmentRun)}</h4>
        </div>
        <ExecutorStatusBadge state={consumerState} eligible={eligibility.eligible} />
      </div>

      <div className="workflow-executor-summary-grid">
        <article>
          <span>{t($ => $.executor.draft)}</span>
          <strong>{draft.draftId}</strong>
          <p>{t($ => $.executor.savedProfile, { profile: draft.executionProfile ?? "review_only", version: eligibility.savedDraftVersion })}</p>
        </article>
        <article>
          <span>{t($ => $.executor.graph)}</span>
          <strong>{t($ => $.executor.graphSize, { nodes: draft.nodes.length, edges: draft.edges.length })}</strong>
          <p>{eligibility.eligible ? t($ => $.executor.eligible) : t($ => $.executor.blockers, { count: eligibility.reasons.length })}</p>
        </article>
        <article>
          <span>{t($ => $.executor.runStore)}</span>
          <strong>{t($ => $.executor.memoryDev100Records)}</strong>
          <p>{t($ => $.executor.scopedReadIsAvailableRestartRecoveryAndReplayRemainDisabled)}</p>
        </article>
        <article>
          <span>{t($ => $.executor.boundary)}</span>
          <strong>{t($ => $.executor.gatewayAdvisoryOnly)}</strong>
          <p>{t($ => $.executor.noToolRAGConfirmationCommitBusinessWriteOrReplayAccess)}</p>
        </article>
      </div>

      <div className="workflow-executor-action-row">
        <button type="button" disabled={pending} onClick={onCreateExecutorDraft}>
          {shellT($ => $.appShell.workflowCreateExecutor)}
        </button>
        <button type="button" disabled={!canStart} onClick={onStartRun}>
          {consumerState.status === "starting" ? t($ => $.executor.running) : t($ => $.executor.start)}
        </button>
        <button type="button" disabled={pending || !record} onClick={onReloadRun}>
          {consumerState.status === "reading" ? t($ => $.executor.reading) : t($ => $.executor.reload)}
        </button>
      </div>

      <div className="workflow-executor-input-grid">
        <label className="workflow-executor-input-field wide">
          <span>{t($ => $.executor.runInput)}</span>
          <textarea
            rows={4}
            maxLength={8192}
            value={inputText}
            disabled={pending}
            onChange={(event) => onInputTextChange(event.currentTarget.value)}
          />
          <small>{t($ => $.executor.theServerRecordsOnlyTheByteCountRawInputIs)}</small>
        </label>
        <label className="workflow-executor-input-field">
          <span>{t($ => $.executor.gatewayModelOverride)}</span>
          <input
            type="text"
            maxLength={256}
            value={model}
            placeholder={t($ => $.executor.useConfiguredDefault)}
            disabled={pending}
            onChange={(event) => onModelChange(event.currentTarget.value)}
          />
          <small>{t($ => $.executor.optionalProviderEndpointAndCredentialsCannotBeSuppliedHere)}</small>
        </label>
      </div>

      {eligibility.conditionNodeIds.length > 0 ? (
        <div className="workflow-executor-condition-grid" aria-label={t($ => $.executor.workflowExecutorExplicitConditions)}>
          {eligibility.conditionNodeIds.map((nodeId) => (
            <label key={nodeId}>
              <span>{nodeId}</span>
              <select
                value={conditionValues[nodeId] ? "true" : "false"}
                disabled={pending}
                onChange={(event) => onConditionValueChange(nodeId, event.currentTarget.value === "true")}
              >
                <option value="false">false</option>
                <option value="true">true</option>
              </select>
            </label>
          ))}
        </div>
      ) : null}

      {!eligibility.eligible ? (
        <div className="workflow-executor-blocker-list" aria-label={t($ => $.executor.workflowExecutorEligibilityBlockers)}>
          {eligibility.reasons.map((reason) => (
            <article key={`${reason.code}-${reason.summary}`}>
              <code>{reason.code}</code>
              <p>{executorBlockerMessage(t, reason.code)}</p>
            </article>
          ))}
        </div>
      ) : null}

      <article className="workflow-executor-state-card">
        <div>
          <span>{t($ => $.executor.consumerState)}</span>
          <strong>{workflowDraftStatusLabel(t, consumerState.status)}</strong>
        </div>
        <p>{executorStateMessage(t, consumerState)}</p>
        <dl>
          <div>
            <dt>{t($ => $.executor.failure)}</dt>
            <dd>{consumerState.failureCode ?? workflowDraftStatusLabel(t, "none")}</dd>
          </div>
          <div>
            <dt>{t($ => $.executor.request)}</dt>
            <dd>{consumerState.requestId}</dd>
          </div>
          <div>
            <dt>{t($ => $.executor.audit)}</dt>
            <dd>{consumerState.auditRef}</dd>
          </div>
        </dl>
      </article>

      {record ? (
        <div className="workflow-executor-record" aria-label={t($ => $.executor.workflowExecutorRunRecord)}>
          <div className="workflow-executor-record-heading">
            <div>
              <span>{t($ => $.executor.runRecord)}</span>
              <strong>{record.runId}</strong>
            </div>
            <span className={`status-badge ${record.status === "succeeded" ? "good" : "bad"}`}>
              {workflowDraftStatusLabel(t, record.status)}
            </span>
          </div>
          <dl className="workflow-executor-record-meta">
            <div>
              <dt>{t($ => $.executor.draftVersion)}</dt>
              <dd>{record.draftVersion}</dd>
            </div>
            <div>
              <dt>{t($ => $.executor.provider)}</dt>
              <dd>{record.selectedProvider || workflowDraftStatusLabel(t, "not_selected")}</dd>
            </div>
            <div>
              <dt>{t($ => $.executor.profile)}</dt>
              <dd>{record.selectedProfile || workflowDraftStatusLabel(t, "none")}</dd>
            </div>
            <div>
              <dt>{t($ => $.executor.model)}</dt>
              <dd>{record.selectedModel || workflowDraftStatusLabel(t, "not_selected")}</dd>
            </div>
            <div>
              <dt>{t($ => $.executor.inputBytes)}</dt>
              <dd>{record.inputBytes}</dd>
            </div>
            <div>
              <dt>{t($ => $.executor.providerCalls)}</dt>
              <dd>{record.sideEffects.providerCalls}</dd>
            </div>
            <div>
              <dt>{t($ => $.executor.toolConfirmation)}</dt>
              <dd>{record.sideEffects.toolCalls} / {record.sideEffects.confirmationCalls}</dd>
            </div>
            <div>
              <dt>{t($ => $.executor.businessReplayWrites)}</dt>
              <dd>{record.sideEffects.businessWrites} / {record.sideEffects.replayWrites}</dd>
            </div>
          </dl>
          {record.output ? (
            <div className="workflow-executor-output">
              <span>{t($ => $.executor.advisoryOutput)}</span>
              <pre>{record.output}</pre>
            </div>
          ) : null}
          <div className="workflow-executor-node-list">
            {record.nodes.map((node) => (
              <WorkflowRunNodeRecordCard key={node.nodeId} node={node} />
            ))}
          </div>
        </div>
      ) : null}
    </section>
  );
}

function ExecutorStatusBadge({
  state,
  eligible,
}: {
  state: WorkflowExecutorConsumerState;
  eligible: boolean;
}) {
  const tone = state.status === "failed"
    ? "bad"
    : state.status === "succeeded" || (state.status === "idle" && eligible)
      ? "good"
      : "neutral";
  const label = state.mode === "disabled"
    ? "dev_gate_disabled"
    : state.status === "idle" && eligible
      ? "ready_to_run"
      : state.status;
  const { t } = useTranslation("workflow");
  return <span className={`status-badge ${tone}`}>{workflowDraftStatusLabel(t, label)}</span>;
}

function WorkflowRunNodeRecordCard({ node }: { node: WorkflowRunNodeRecord }) {
  const { t } = useTranslation("workflow");
  const tone = node.status === "succeeded"
    ? "good"
    : node.status === "failed"
      ? "bad"
      : "neutral";
  return (
    <article>
      <div className="workflow-executor-record-heading">
        <div>
          <span>{workflowDraftStatusLabel(t, node.nodeType)}</span>
          <strong>{node.label || node.nodeId}</strong>
        </div>
        <span className={`status-badge ${tone}`}>{workflowDraftStatusLabel(t, node.status)}</span>
      </div>
      <dl>
        <div>
          <dt>{t($ => $.executor.node)}</dt>
          <dd>{node.nodeId}</dd>
        </div>
        <div>
          <dt>{t($ => $.executor.duration)}</dt>
          <dd>{node.durationMs} ms</dd>
        </div>
        <div>
          <dt>{t($ => $.executor.failure)}</dt>
          <dd>{node.failureCode || workflowDraftStatusLabel(t, "none")}</dd>
        </div>
      </dl>
      <p>{node.outputPreview || t($ => $.executor.noPreview)}</p>
    </article>
  );
}
