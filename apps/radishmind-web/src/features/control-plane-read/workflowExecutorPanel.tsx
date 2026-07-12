import type { WorkflowDraftDesignerDraft } from "./workflowDraftDesigner";
import type {
  WorkflowExecutorConsumerState,
  WorkflowExecutorEligibility,
  WorkflowRunNodeRecord,
} from "./workflowExecutorConsumer";

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
          <p className="eyebrow">Workflow Executor v0</p>
          <h4 id="workflow-executor-v0-title">Bounded development run</h4>
        </div>
        <ExecutorStatusBadge state={consumerState} eligible={eligibility.eligible} />
      </div>

      <div className="workflow-executor-summary-grid">
        <article>
          <span>Draft</span>
          <strong>{draft.draftId}</strong>
          <p>{draft.executionProfile ?? "review_only"} / saved version {eligibility.savedDraftVersion}</p>
        </article>
        <article>
          <span>Graph</span>
          <strong>{draft.nodes.length} nodes / {draft.edges.length} edges</strong>
          <p>{eligibility.eligible ? "Eligible for server revalidation." : `${eligibility.reasons.length} blocker(s).`}</p>
        </article>
        <article>
          <span>Run store</span>
          <strong>memory dev / 100 records</strong>
          <p>Scoped read is available; restart recovery and replay remain disabled.</p>
        </article>
        <article>
          <span>Boundary</span>
          <strong>Gateway advisory only</strong>
          <p>No tool, RAG, confirmation commit, business write, or replay access.</p>
        </article>
      </div>

      <div className="workflow-executor-action-row">
        <button type="button" disabled={pending} onClick={onCreateExecutorDraft}>
          Create executor v0 draft
        </button>
        <button type="button" disabled={!canStart} onClick={onStartRun}>
          {consumerState.status === "starting" ? "Running…" : "Start bounded run"}
        </button>
        <button type="button" disabled={pending || !record} onClick={onReloadRun}>
          {consumerState.status === "reading" ? "Reloading…" : "Reload run record"}
        </button>
      </div>

      <div className="workflow-executor-input-grid">
        <label className="workflow-executor-input-field wide">
          <span>Run input</span>
          <textarea
            rows={4}
            maxLength={8192}
            value={inputText}
            disabled={pending}
            onChange={(event) => onInputTextChange(event.currentTarget.value)}
          />
          <small>The server records only the byte count; raw input is not written to the run record.</small>
        </label>
        <label className="workflow-executor-input-field">
          <span>Gateway model override</span>
          <input
            type="text"
            maxLength={256}
            value={model}
            placeholder="Use configured default"
            disabled={pending}
            onChange={(event) => onModelChange(event.currentTarget.value)}
          />
          <small>Optional; provider endpoint and credentials cannot be supplied here.</small>
        </label>
      </div>

      {eligibility.conditionNodeIds.length > 0 ? (
        <div className="workflow-executor-condition-grid" aria-label="Workflow executor explicit conditions">
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
        <div className="workflow-executor-blocker-list" aria-label="Workflow executor eligibility blockers">
          {eligibility.reasons.map((reason) => (
            <article key={`${reason.code}-${reason.summary}`}>
              <code>{reason.code}</code>
              <p>{reason.summary}</p>
            </article>
          ))}
        </div>
      ) : null}

      <article className="workflow-executor-state-card">
        <div>
          <span>Consumer state</span>
          <strong>{consumerState.status}</strong>
        </div>
        <p>{consumerState.summary}</p>
        <dl>
          <div>
            <dt>Failure</dt>
            <dd>{consumerState.failureCode ?? "none"}</dd>
          </div>
          <div>
            <dt>Request</dt>
            <dd>{consumerState.requestId}</dd>
          </div>
          <div>
            <dt>Audit</dt>
            <dd>{consumerState.auditRef}</dd>
          </div>
        </dl>
      </article>

      {record ? (
        <div className="workflow-executor-record" aria-label="Workflow executor run record">
          <div className="workflow-executor-record-heading">
            <div>
              <span>Run record</span>
              <strong>{record.runId}</strong>
            </div>
            <span className={`status-badge ${record.status === "succeeded" ? "good" : "bad"}`}>
              {record.status}
            </span>
          </div>
          <dl className="workflow-executor-record-meta">
            <div>
              <dt>Draft version</dt>
              <dd>{record.draftVersion}</dd>
            </div>
            <div>
              <dt>Provider</dt>
              <dd>{record.selectedProvider || "not selected"}</dd>
            </div>
            <div>
              <dt>Profile</dt>
              <dd>{record.selectedProfile || "none"}</dd>
            </div>
            <div>
              <dt>Model</dt>
              <dd>{record.selectedModel || "not selected"}</dd>
            </div>
            <div>
              <dt>Input bytes</dt>
              <dd>{record.inputBytes}</dd>
            </div>
            <div>
              <dt>Provider calls</dt>
              <dd>{record.sideEffects.providerCalls}</dd>
            </div>
            <div>
              <dt>Tool / confirmation</dt>
              <dd>{record.sideEffects.toolCalls} / {record.sideEffects.confirmationCalls}</dd>
            </div>
            <div>
              <dt>Business / replay writes</dt>
              <dd>{record.sideEffects.businessWrites} / {record.sideEffects.replayWrites}</dd>
            </div>
          </dl>
          {record.output ? (
            <div className="workflow-executor-output">
              <span>Advisory output</span>
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
    ? "dev gate disabled"
    : state.status === "idle" && eligible
      ? "ready to run"
      : state.status;
  return <span className={`status-badge ${tone}`}>{label}</span>;
}

function WorkflowRunNodeRecordCard({ node }: { node: WorkflowRunNodeRecord }) {
  const tone = node.status === "succeeded"
    ? "good"
    : node.status === "failed"
      ? "bad"
      : "neutral";
  return (
    <article>
      <div className="workflow-executor-record-heading">
        <div>
          <span>{node.nodeType}</span>
          <strong>{node.label || node.nodeId}</strong>
        </div>
        <span className={`status-badge ${tone}`}>{node.status}</span>
      </div>
      <dl>
        <div>
          <dt>Node</dt>
          <dd>{node.nodeId}</dd>
        </div>
        <div>
          <dt>Duration</dt>
          <dd>{node.durationMs} ms</dd>
        </div>
        <div>
          <dt>Failure</dt>
          <dd>{node.failureCode || "none"}</dd>
        </div>
      </dl>
      <p>{node.outputPreview || "No output preview was retained."}</p>
    </article>
  );
}
