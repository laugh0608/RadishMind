import type { WorkflowHTTPToolActionPermissions, WorkflowHTTPToolActionPlan } from "./workflowHTTPToolActionConsumer.ts";
import type { WorkflowHTTPToolExecutionState } from "./workflowHTTPToolExecutionConsumer.ts";

export function WorkflowHTTPToolExecutionPanel({
  plan,
  state,
  permissions,
  inputText,
  model,
  onInputTextChange,
  onModelChange,
  onExecute,
}: {
  plan: WorkflowHTTPToolActionPlan | null;
  state: WorkflowHTTPToolExecutionState;
  permissions: WorkflowHTTPToolActionPermissions;
  inputText: string;
  model: string;
  onInputTextChange: (value: string) => void;
  onModelChange: (value: string) => void;
  onExecute: () => void;
}) {
  const executing = state.status === "executing";
  const canExecute = plan?.status === "approved" && permissions.execute.available &&
    inputText.trim().length > 0 && new TextEncoder().encode(inputText).byteLength <= 8_192 && !executing;
  const run = state.run;
  const attempt = run?.toolAttempt ?? null;

  return (
    <section className="workflow-executor-v0 workflow-http-tool-execution" id="workflow-http-tool-execution" aria-labelledby="workflow-http-tool-execution-title">
      <div className="section-heading compact-heading">
        <div>
          <p className="eyebrow">Workflow HTTP Tool · Batch C</p>
          <h4 id="workflow-http-tool-execution-title">Explicit confirmed execution</h4>
        </div>
        <span className={`status-badge ${stateTone(state.status)}`}>{state.status}</span>
      </div>

      <div className="workflow-executor-summary-grid">
        <article><span>Durable plan</span><strong>{plan?.planId ?? "not selected"}</strong><p>{plan ? `${plan.status} · version ${plan.recordVersion}` : "Create and approve an exact plan first."}</p></article>
        <article><span>Execution grant</span><strong>{permissions.execute.available ? "available" : "blocked"}</strong><p><code>{permissions.execute.requiredGrants.join(" + ")}</code></p></article>
        <article><span>Network rule</span><strong>Explicit click · at most once</strong><p>Approval never starts the request; consumed plans cannot be retried.</p></article>
        <article><span>Durable recovery</span><strong>Plan + Run History</strong><p>Refresh reloads the consumed plan and its v2 run from server-owned records.</p></article>
      </div>

      <div className="workflow-executor-input-grid">
        <label className="workflow-executor-input-field">
          <span>Review prompt</span>
          <textarea value={inputText} maxLength={8192} disabled={executing} onChange={(event) => onInputTextChange(event.currentTarget.value)} />
          <small>Bounded request input is not retained in the run record.</small>
        </label>
        <label className="workflow-executor-input-field">
          <span>Model</span>
          <input value={model} maxLength={256} disabled={executing} placeholder="configured default" onChange={(event) => onModelChange(event.currentTarget.value)} />
          <small>The existing Gateway resolves the configured provider profile.</small>
        </label>
      </div>
      <div className="workflow-executor-action-row">
        <button type="button" disabled={!canExecute} onClick={onExecute}>
          {executing ? "Executing single attempt…" : "Execute approved plan"}
        </button>
      </div>
      {plan?.status === "approved" ? <p className="boundary-note">Approval is durable and no network attempt has started. Execution requires this separate action.</p> : null}
      {plan?.status === "consumed" ? <p className="boundary-note">This plan is consumed. Inspect its durable v2 record in Run History; retry and resume remain unavailable.</p> : null}

      <article className="workflow-executor-state-card" aria-live="polite">
        <div><span>Execution state</span><strong>{state.status}</strong></div>
        <p>{state.summary}</p>
        <dl className="workflow-user-workspace-home-meta">
          <div><dt>Failure</dt><dd>{state.failureCode || "none"}</dd></div>
          <div><dt>Request</dt><dd>{state.requestId || "none"}</dd></div>
          <div><dt>Audit</dt><dd>{state.auditRef || "none"}</dd></div>
          <div><dt>Run</dt><dd>{run?.runId ?? "not created"}</dd></div>
        </dl>
      </article>

      {run ? (
        <article className="workflow-executor-record" aria-label="Confirmed HTTP Tool run v2 detail">
          <div className="workflow-executor-record-heading">
            <div><span>Run v2</span><strong>{run.runId}</strong></div>
            <span className={`status-badge ${run.status === "succeeded" ? "status-good" : run.status === "outcome_unknown" ? "status-neutral" : "status-bad"}`}>{run.status}</span>
          </div>
          <dl className="workflow-executor-record-meta">
            <div><dt>Confirmation</dt><dd>{run.confirmationId}</dd></div>
            <div><dt>Tool attempt</dt><dd>{attempt?.attemptId ?? "unavailable"} · {attempt?.status ?? "unavailable"}</dd></div>
            <div><dt>HTTP class</dt><dd>{attempt?.httpStatusClass || "not recorded"}</dd></div>
            <div><dt>Response projection</dt><dd>{attempt ? `${attempt.responseBytes} bytes · ${attempt.durationMs} ms` : "unavailable"}</dd></div>
            <div><dt>Controlled side effects</dt><dd>tool {run.sideEffects.toolCalls} · confirmation {run.sideEffects.confirmationCalls}</dd></div>
            <div><dt>Forbidden writes</dt><dd>business {run.sideEffects.businessWrites} · replay {run.sideEffects.replayWrites}</dd></div>
            <div><dt>Failure boundary</dt><dd>{run.diagnostic?.failureBoundary || "none"}</dd></div>
            <div><dt>Tool category</dt><dd>{run.diagnostic?.toolFailureCategory || "none"}</dd></div>
          </dl>
          {attempt && Object.keys(attempt.outputProjection).length > 0 ? (
            <dl className="workflow-executor-record-meta" aria-label="Sanitized HTTP Tool projection">
              {Object.entries(attempt.outputProjection).map(([key, value]) => <div key={key}><dt>{key}</dt><dd>{String(value ?? "")}</dd></div>)}
            </dl>
          ) : null}
          {run.status === "outcome_unknown" ? <p className="failure-summary">The remote outcome is uncertain. Review metadata and create a new plan only after human assessment; do not retry this plan.</p> : null}
        </article>
      ) : null}
    </section>
  );
}

function stateTone(status: WorkflowHTTPToolExecutionState["status"]): "status-good" | "status-bad" | "status-neutral" {
  if (status === "succeeded") return "status-good";
  if (status === "failed") return "status-bad";
  return "status-neutral";
}
