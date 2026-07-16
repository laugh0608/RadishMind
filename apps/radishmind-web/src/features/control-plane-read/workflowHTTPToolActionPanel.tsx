import type { FormEvent } from "react";

import type { WorkflowDraftDesignerDraft } from "./workflowDraftDesigner";
import {
  validateWorkflowHTTPToolPublicArguments,
  type WorkflowHTTPToolActionConsumerState,
  type WorkflowHTTPToolActionEligibility,
  type WorkflowHTTPToolActionPermissions,
  type WorkflowHTTPToolHumanDecision,
  type WorkflowHTTPToolPublicArguments,
} from "./workflowHTTPToolActionConsumer";

export function WorkflowHTTPToolActionPanel({
  draft,
  consumerState,
  eligibility,
  permissions,
  resourceKey,
  locale,
  onResourceKeyChange,
  onLocaleChange,
  onCreatePlan,
  onReloadPlan,
  onDecision,
}: {
  draft: WorkflowDraftDesignerDraft;
  consumerState: WorkflowHTTPToolActionConsumerState;
  eligibility: WorkflowHTTPToolActionEligibility;
  permissions: WorkflowHTTPToolActionPermissions;
  resourceKey: string;
  locale: string;
  onResourceKeyChange: (value: string) => void;
  onLocaleChange: (value: string) => void;
  onCreatePlan: (publicArguments: WorkflowHTTPToolPublicArguments) => void;
  onReloadPlan: () => void;
  onDecision: (decision: WorkflowHTTPToolHumanDecision) => void;
}) {
  const pending = consumerState.status === "creating" || consumerState.status === "reading" || consumerState.status === "deciding";
  const argumentValidation = validateWorkflowHTTPToolPublicArguments({ resourceKey, locale });
  const plan = consumerState.actionPlan;
  const canCreate = consumerState.mode === "dev_workflow_http_tool_http" && eligibility.eligible &&
    permissions.plan.available && argumentValidation.valid && Boolean(argumentValidation.value) && !pending;
  const decisions = availableDecisions(plan?.status ?? null);

  const submitCreate = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (canCreate && argumentValidation.value) onCreatePlan(argumentValidation.value);
  };

  return (
    <section
      className="workflow-executor-v0 workflow-http-tool-action-review"
      id="workflow-http-tool-action-review"
      aria-labelledby="workflow-http-tool-action-review-title"
    >
      <div className="section-heading compact-heading">
        <div>
          <p className="eyebrow">Workflow HTTP Tool · Batch A</p>
          <h4 id="workflow-http-tool-action-review-title">Durable action plan review</h4>
        </div>
        <ActionStatusBadge state={consumerState} eligible={eligibility.eligible} />
      </div>

      <div className="workflow-executor-summary-grid">
        <article>
          <span>Exact saved draft</span>
          <strong>{draft.draftId}</strong>
          <p>Version {eligibility.draftVersion || "not saved"}; local edits must be clean.</p>
        </article>
        <article>
          <span>HTTP Tool node</span>
          <strong>{eligibility.nodeId || "not eligible"}</strong>
          <p>{eligibility.toolId || "Select one exact versioned HTTP Tool reference."}</p>
        </article>
        <article>
          <span>Review lifecycle</span>
          <strong>Plan → human decision</strong>
          <p>Every decision uses the displayed durable record version as its CAS boundary.</p>
        </article>
        <article>
          <span>Batch A boundary</span>
          <strong>Governance metadata only</strong>
          <p>Approval records qualification only; network, provider, run, business, and replay counts remain zero.</p>
        </article>
      </div>

      <div className="workflow-executor-summary-grid" aria-label="Workflow HTTP Tool operation grants">
        <PermissionCard permission={permissions.plan} label="Create plan" />
        <PermissionCard permission={permissions.read} label="Read detail" />
        <PermissionCard permission={permissions.confirm} label="Human decision" />
        <PermissionCard permission={permissions.execute} label="Execute tool" />
      </div>

      <form onSubmit={submitCreate}>
        <div className="workflow-executor-input-grid">
          <label className="workflow-executor-input-field">
            <span>Resource key</span>
            <input
              type="text"
              maxLength={160}
              autoComplete="off"
              value={resourceKey}
              placeholder="reviewed-resource"
              disabled={pending || !permissions.plan.available}
              onChange={(event) => onResourceKeyChange(event.currentTarget.value)}
            />
            <small>Required public identifier; only the registered schema field is accepted.</small>
          </label>
          <label className="workflow-executor-input-field">
            <span>Locale</span>
            <input
              type="text"
              maxLength={35}
              autoComplete="off"
              value={locale}
              placeholder="zh-CN"
              disabled={pending || !permissions.plan.available}
              onChange={(event) => onLocaleChange(event.currentTarget.value)}
            />
            <small>Optional short language tag; it cannot change the server-owned target policy.</small>
          </label>
        </div>
        <div className="workflow-executor-action-row">
          <button type="submit" disabled={!canCreate}>
            {consumerState.status === "creating" ? "Creating…" : "Create immutable plan"}
          </button>
          <button type="button" disabled={pending || !plan || !permissions.read.available} onClick={onReloadPlan}>
            {consumerState.status === "reading" ? "Reloading…" : "Reload durable detail"}
          </button>
        </div>
        {!argumentValidation.valid && (resourceKey || locale) ? (
          <p className="boundary-note" role="alert">{argumentValidation.summary}</p>
        ) : null}
      </form>

      {!eligibility.eligible ? (
        <div className="workflow-executor-blocker-list" aria-label="HTTP Tool action plan eligibility blockers">
          {eligibility.reasons.map((reason) => (
            <article key={`${reason.code}-${reason.summary}`}>
              <code>{reason.code}</code>
              <p>{reason.summary}</p>
            </article>
          ))}
        </div>
      ) : null}

      <article className="workflow-executor-state-card" aria-live="polite">
        <div>
          <span>Consumer state</span>
          <strong>{consumerState.status}</strong>
        </div>
        <p>{consumerState.summary}</p>
        <dl className="workflow-user-workspace-home-meta">
          <div><dt>Failure</dt><dd>{consumerState.failureCode || "none"}</dd></div>
          <div><dt>Request</dt><dd>{consumerState.requestId || "none"}</dd></div>
          <div><dt>Audit</dt><dd>{consumerState.auditRef || "none"}</dd></div>
          <div><dt>Source</dt><dd>{consumerState.mode}</dd></div>
        </dl>
      </article>

      {plan ? (
        <article className="workflow-executor-record" aria-label="Durable HTTP Tool action plan detail">
          <div className="workflow-executor-record-heading">
            <div>
              <span>Action plan</span>
              <strong>{plan.planId}</strong>
            </div>
            <span className={`status-badge ${planStatusTone(plan.status)}`}>{plan.status}</span>
          </div>
          <dl className="workflow-executor-record-meta">
            <div><dt>Record version</dt><dd>{plan.recordVersion}</dd></div>
            <div><dt>Draft</dt><dd>{plan.draftId} · v{plan.draftVersion}</dd></div>
            <div><dt>Node</dt><dd>{plan.nodeId}</dd></div>
            <div><dt>Tool</dt><dd>{plan.toolId} · version {plan.toolVersion}</dd></div>
            <div><dt>Method</dt><dd>{plan.method}</dd></div>
            <div><dt>Target policy</dt><dd>{plan.targetPolicyKey}</dd></div>
            <div><dt>Resource key</dt><dd>{plan.publicArguments.resourceKey}</dd></div>
            <div><dt>Locale</dt><dd>{plan.publicArguments.locale ?? "default"}</dd></div>
            <div><dt>Output fields</dt><dd>{plan.outputFields.join(", ")}</dd></div>
            <div><dt>Output schema digest</dt><dd>{plan.outputSchemaDigest}</dd></div>
            <div><dt>Timeout</dt><dd>{plan.timeoutMs} ms</dd></div>
            <div><dt>Response / projection</dt><dd>{plan.maxResponseBytes} / {plan.maxOutputBytes} bytes</dd></div>
            <div><dt>Created</dt><dd>{plan.createdAt}</dd></div>
            <div><dt>Expires</dt><dd>{plan.expiresAt}</dd></div>
            <div><dt>Planned by</dt><dd>{plan.plannedByActorRef}</dd></div>
            <div><dt>Last decision</dt><dd>{plan.lastDecisionByActorRef ?? "none"}</dd></div>
          </dl>

          <div className="workflow-executor-action-row" aria-label="HTTP Tool action plan decisions">
            {decisions.map((decision) => (
              <button key={decision} type="button" disabled={pending || !permissions.confirm.available} onClick={() => onDecision(decision)}>
                {decisionLabel(decision)}
              </button>
            ))}
            <button type="button" disabled title={permissions.execute.summary}>
              Execute · Batch C unavailable
            </button>
          </div>
          {plan.status === "approved" ? (
            <p className="boundary-note">
              Approved qualification is durable. Batch A deliberately provides no network-start action and creates no workflow run.
            </p>
          ) : null}
        </article>
      ) : null}

      {consumerState.confirmationDecision ? (
        <article className="workflow-executor-state-card" aria-label="Latest confirmation decision">
          <div><span>Latest decision</span><strong>{consumerState.confirmationDecision.outcome}</strong></div>
          <p>
            {consumerState.confirmationDecision.decidedByActorRef} recorded a {consumerState.confirmationDecision.actorSource} decision at {consumerState.confirmationDecision.decidedAt}.
          </p>
          <dl className="workflow-user-workspace-home-meta">
            <div><dt>Confirmation</dt><dd>{consumerState.confirmationDecision.confirmationId}</dd></div>
            <div><dt>Audit</dt><dd>{consumerState.confirmationDecision.auditRef}</dd></div>
          </dl>
        </article>
      ) : null}
    </section>
  );
}

function PermissionCard({
  permission,
  label,
}: {
  permission: WorkflowHTTPToolActionPermissions[keyof WorkflowHTTPToolActionPermissions];
  label: string;
}) {
  return (
    <article>
      <span>{label}</span>
      <strong>{permission.available ? "grant available" : `${permission.phase} unavailable`}</strong>
      <p><code>{permission.requiredGrants.join(" + ")}</code></p>
      <small>{permission.summary}</small>
    </article>
  );
}

function availableDecisions(status: string | null): WorkflowHTTPToolHumanDecision[] {
  if (status === "pending") return ["approve", "reject", "defer", "cancel"];
  if (status === "deferred") return ["approve", "reject", "cancel"];
  if (status === "approved") return ["cancel"];
  return [];
}

function decisionLabel(decision: WorkflowHTTPToolHumanDecision): string {
  if (decision === "approve") return "Approve qualification";
  if (decision === "reject") return "Reject";
  if (decision === "defer") return "Defer review";
  return "Cancel plan";
}

function planStatusTone(status: string): "good" | "bad" | "neutral" {
  if (status === "approved") return "good";
  if (["rejected", "canceled", "expired", "invalidated"].includes(status)) return "bad";
  return "neutral";
}

function ActionStatusBadge({
  state,
  eligible,
}: {
  state: WorkflowHTTPToolActionConsumerState;
  eligible: boolean;
}) {
  const tone = state.status === "failed" ? "bad" : state.status === "ready" ? "good" : "neutral";
  const label = state.mode === "disabled" ? "offline · zero requests" : eligible ? state.status : "blocked";
  return <span className={`status-badge ${tone}`}>{label}</span>;
}
