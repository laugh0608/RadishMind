import { useTranslation } from "react-i18next";
import "../../i18n/workflowHTTPToolResources.ts";
import { workflowHTTPToolStatus, workflowHTTPToolExecutionFeedback } from "./workflowHTTPToolMessages.ts";
import type { WorkflowHTTPToolActionPermissions, WorkflowHTTPToolActionPlan } from "./workflowHTTPToolActionConsumer.ts";
import type { WorkflowHTTPToolExecutionState } from "./workflowHTTPToolExecutionConsumer.ts";
import ActionSafetyReadPanel from "./ActionSafetyReadPanel.tsx";

export function WorkflowHTTPToolExecutionPanel({
  plan,
  state,
  permissions,
  inputText,
  model,
  onInputTextChange,
  onModelChange,
  onExecute,
  sectionId = "workflow-http-tool-execution",
}: {
  plan: WorkflowHTTPToolActionPlan | null;
  state: WorkflowHTTPToolExecutionState;
  permissions: WorkflowHTTPToolActionPermissions;
  inputText: string;
  model: string;
  onInputTextChange: (value: string) => void;
  onModelChange: (value: string) => void;
  onExecute: () => void;
  sectionId?: string;
}) {
  const { t } = useTranslation("workflow");
  const executing = state.status === "executing";
  const executionPermission = plan?.sourceKind === "workflow_definition" ? permissions.definitionExecute : permissions.execute;
  const canExecute = plan?.status === "approved" && executionPermission.available &&
    inputText.trim().length > 0 && new TextEncoder().encode(inputText).byteLength <= 8_192 && !executing;
  const run = state.run;
  const attempt = run?.toolAttempt ?? null;

  return (
    <section className="workflow-executor-v0 workflow-http-tool-execution" id={sectionId} aria-labelledby={`${sectionId}-title`}>
      <div className="section-heading compact-heading">
        <div>
          <p className="eyebrow">{t($ => $.httpTool.executionEyebrow)}</p>
          <h4 id={`${sectionId}-title`}>{t($ => $.httpTool.executionTitle)}</h4>
        </div>
        <span className={`status-badge ${stateTone(state.status)}`}>{workflowHTTPToolStatus(t, state.status)}</span>
      </div>

      <div className="workflow-executor-summary-grid">
        <article><span>{t($ => $.httpTool.durablePlan)}</span><strong>{plan?.planId ?? t($ => $.httpTool.notSelected)}</strong><p>{plan ? t($ => $.httpTool.planVersion, { status: workflowHTTPToolStatus(t, plan.status), version: plan.recordVersion }) : t($ => $.httpTool.createApproveFirst)}</p></article>
        <article><span>{t($ => $.httpTool.executionGrant)}</span><strong>{t($ => executionPermission.available ? $.httpTool.available : $.httpTool.blocked)}</strong><p><code>{executionPermission.requiredGrants.join(" + ")}</code></p></article>
        <article><span>{t($ => $.httpTool.networkRule)}</span><strong>{t($ => $.httpTool.explicitOnce)}</strong><p>{t($ => $.httpTool.noRetry)}</p></article>
        <article><span>{t($ => $.httpTool.recovery)}</span><strong>{t($ => $.httpTool.planHistory)}</strong><p>{t($ => $.httpTool.recoveryHint, { version: plan?.sourceKind === "workflow_definition" ? "v9" : "v2" })}</p></article>
      </div>

      <div className="workflow-executor-input-grid">
        <label className="workflow-executor-input-field">
          <span id={`${sectionId}-input-label`}>{t($ => $.httpTool.reviewPrompt)}</span>
          <textarea aria-labelledby={`${sectionId}-input-label`} aria-describedby={`${sectionId}-input-hint`} value={inputText} maxLength={8192} disabled={executing} onChange={(event) => onInputTextChange(event.currentTarget.value)} />
          <small id={`${sectionId}-input-hint`}>{t($ => $.httpTool.inputHint)}</small>
        </label>
        <label className="workflow-executor-input-field">
          <span id={`${sectionId}-model-label`}>{t($ => $.httpTool.model)}</span>
          <input aria-labelledby={`${sectionId}-model-label`} aria-describedby={`${sectionId}-model-hint`} value={model} maxLength={256} disabled={executing} placeholder={t($ => $.httpTool.modelDefault)} onChange={(event) => onModelChange(event.currentTarget.value)} />
          <small id={`${sectionId}-model-hint`}>{t($ => $.httpTool.modelHint)}</small>
        </label>
      </div>
      <div className="workflow-executor-action-row">
        <button type="button" disabled={!canExecute} onClick={onExecute}>
          {executing ? t($ => $.httpTool.executing) : t($ => $.httpTool.execute)}
        </button>
      </div>
      {plan?.status === "approved" ? <p className="boundary-note">{t($ => $.httpTool.approvedExecution)}</p> : null}
      {plan?.status === "consumed" ? <p className="boundary-note">{t($ => $.httpTool.consumedNote, { version: plan.sourceKind === "workflow_definition" ? "v9" : "v2" })}</p> : null}

      <article className="workflow-executor-state-card" aria-live="polite">
        <div><span>{t($ => $.httpTool.executionState)}</span><strong>{workflowHTTPToolStatus(t, state.status)}</strong></div>
        <p>{workflowHTTPToolExecutionFeedback(t, state)}</p>
        <dl className="workflow-user-workspace-home-meta">
          <div><dt>{t($ => $.httpTool.failure)}</dt><dd>{state.failureCode || t($ => $.httpTool.none)}</dd></div>
          <div><dt>{t($ => $.httpTool.request)}</dt><dd>{state.requestId || t($ => $.httpTool.none)}</dd></div>
          <div><dt>{t($ => $.httpTool.audit)}</dt><dd>{state.auditRef || t($ => $.httpTool.none)}</dd></div>
          <div><dt>{t($ => $.httpTool.run)}</dt><dd>{run?.runId ?? t($ => $.httpTool.notCreated)}</dd></div>
        </dl>
      </article>

      {state.failureCode && state.summary ? <details><summary>{t($ => $.httpTool.originalDiagnostic)}</summary><p>{state.summary}</p></details> : null}

      {run ? (
        <article className="workflow-executor-record" aria-label={t($ => $.httpTool.runDetail, { schema: run.schemaVersion })}>
          <div className="workflow-executor-record-heading">
            <div><span>{run.schemaVersion}</span><strong>{run.runId}</strong></div>
            <span className={`status-badge ${run.status === "succeeded" ? "status-good" : run.status === "outcome_unknown" ? "status-neutral" : "status-bad"}`}>{workflowHTTPToolStatus(t, run.status)}</span>
          </div>
          <dl className="workflow-executor-record-meta">
            <div><dt>{t($ => $.httpTool.confirmation)}</dt><dd>{run.confirmationId}</dd></div>
            <div><dt>{t($ => $.httpTool.attempt)}</dt><dd>{attempt?.attemptId ?? t($ => $.httpTool.unavailable)} · {attempt ? workflowHTTPToolStatus(t, attempt.status) : t($ => $.httpTool.unavailable)}</dd></div>
            <div><dt>{t($ => $.httpTool.httpClass)}</dt><dd>{attempt?.httpStatusClass || t($ => $.httpTool.notRecorded)}</dd></div>
            <div><dt>{t($ => $.httpTool.projection)}</dt><dd>{attempt ? `${attempt.responseBytes} bytes · ${attempt.durationMs} ms` : t($ => $.httpTool.unavailable)}</dd></div>
            <div><dt>{t($ => $.httpTool.sideEffects)}</dt><dd>{t($ => $.httpTool.observedCalls, { tool: run.sideEffects.toolCalls, confirmation: run.sideEffects.confirmationCalls })}</dd></div>
            <div><dt>{t($ => $.httpTool.forbiddenWrites)}</dt><dd>{t($ => $.httpTool.writeCounts, { business: run.sideEffects.businessWrites, replay: run.sideEffects.replayWrites })}</dd></div>
            <div><dt>{t($ => $.httpTool.failureBoundary)}</dt><dd>{run.diagnostic?.failureBoundary || t($ => $.httpTool.none)}</dd></div>
            <div><dt>{t($ => $.httpTool.toolCategory)}</dt><dd>{run.diagnostic?.toolFailureCategory || t($ => $.httpTool.none)}</dd></div>
          </dl>
          {attempt && Object.keys(attempt.outputProjection).length > 0 ? (
            <dl className="workflow-executor-record-meta" aria-label={t($ => $.httpTool.safeProjection)}>
              {Object.entries(attempt.outputProjection).map(([key, value]) => <div key={key}><dt>{key}</dt><dd>{String(value ?? "")}</dd></div>)}
            </dl>
          ) : null}
          {run.status === "outcome_unknown" ? <p className="failure-summary">{t($ => $.httpTool.unknownOutcome)}</p> : null}
        </article>
      ) : null}
      <ActionSafetyReadPanel projection={state.actionSafety} title={t($ => $.httpTool.runSafety)} />
    </section>
  );
}

function stateTone(status: WorkflowHTTPToolExecutionState["status"]): "status-good" | "status-bad" | "status-neutral" {
  if (status === "succeeded") return "status-good";
  if (status === "failed") return "status-bad";
  return "status-neutral";
}
