import { useTranslation } from "react-i18next";
import "../../i18n/workflowHTTPToolResources.ts";
import { workflowHTTPToolStatus, workflowHTTPToolFailure, workflowHTTPToolActionFeedback, workflowHTTPToolArgumentsFeedback } from "./workflowHTTPToolMessages.ts";
import type { FormEvent } from "react";
import ActionSafetyReadPanel from "./ActionSafetyReadPanel.tsx";

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
  source,
  sectionId = "workflow-http-tool-action-review",
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
  draft?: WorkflowDraftDesignerDraft;
  source?: {
    kind: "saved_workflow_draft" | "workflow_definition";
    id: string;
    version: number;
    nodeId: string;
    toolId: string;
  };
  sectionId?: string;
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
  const { t } = useTranslation("workflow");
  const pending = consumerState.status === "creating" || consumerState.status === "reading" || consumerState.status === "deciding";
  const argumentValidation = validateWorkflowHTTPToolPublicArguments({ resourceKey, locale });
  const plan = consumerState.actionPlan;
  const sourceView = source ?? {
    kind: "saved_workflow_draft" as const,
    id: draft?.draftId ?? "",
    version: eligibility.draftVersion,
    nodeId: eligibility.nodeId,
    toolId: eligibility.toolId,
  };
  const planPermission = sourceView.kind === "workflow_definition" ? permissions.definitionPlan : permissions.plan;
  const executePermission = sourceView.kind === "workflow_definition" ? permissions.definitionExecute : permissions.execute;
  const canCreate = consumerState.mode === "dev_workflow_http_tool_http" && eligibility.eligible &&
    planPermission.available && argumentValidation.valid && Boolean(argumentValidation.value) && !pending;
  const decisions = availableDecisions(plan?.status ?? null);

  const submitCreate = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (canCreate && argumentValidation.value) onCreatePlan(argumentValidation.value);
  };

  return (
    <section
      className="workflow-executor-v0 workflow-http-tool-action-review"
      id={sectionId}
      aria-labelledby={`${sectionId}-title`}
    >
      <div className="section-heading compact-heading">
        <div>
          <p className="eyebrow">{t($ => $.httpTool.actionEyebrow)}</p>
          <h4 id={`${sectionId}-title`}>{t($ => $.httpTool.actionTitle)}</h4>
        </div>
        <ActionStatusBadge state={consumerState} eligible={eligibility.eligible} />
      </div>

      <div className="workflow-executor-summary-grid">
        <article>
          <span>{t($ => sourceView.kind === "workflow_definition" ? $.httpTool.definitionSource : $.httpTool.savedSource)}</span>
          <strong>{sourceView.id}</strong>
          <p>{t($ => $.httpTool.sourceVersion, { version: sourceView.version || t($ => $.httpTool.unavailable) })}</p>
        </article>
        <article>
          <span>{t($ => $.httpTool.toolNode)}</span>
          <strong>{sourceView.nodeId || t($ => $.httpTool.notEligible)}</strong>
          <p>{sourceView.toolId || t($ => $.httpTool.toolReference)}</p>
        </article>
        <article>
          <span>{t($ => $.httpTool.reviewLifecycle)}</span>
          <strong>{t($ => $.httpTool.planDecision)}</strong>
          <p>{t($ => $.httpTool.casBoundary)}</p>
        </article>
        <article>
          <span>{t($ => $.httpTool.confirmationBoundary)}</span>
          <strong>{t($ => $.httpTool.approvalNotExecution)}</strong>
          <p>{t($ => $.httpTool.approvalExplanation)}</p>
        </article>
      </div>

      <div className="workflow-executor-summary-grid" aria-label={t($ => $.httpTool.grants)}>
        <PermissionCard permission={planPermission} label={t($ => $.httpTool.createGrant)} />
        <PermissionCard permission={permissions.read} label={t($ => $.httpTool.readGrant)} />
        <PermissionCard permission={permissions.confirm} label={t($ => $.httpTool.confirmGrant)} />
        <PermissionCard permission={executePermission} label={t($ => $.httpTool.executeGrant)} />
      </div>

      <form onSubmit={submitCreate}>
        <div className="workflow-executor-input-grid">
          <label className="workflow-executor-input-field">
            <span id={`${sectionId}-resource-label`}>{t($ => $.httpTool.resourceKey)}</span>
            <input
              type="text"
              maxLength={160}
              autoComplete="off"
              aria-labelledby={`${sectionId}-resource-label`}
              aria-describedby={`${sectionId}-resource-hint`}
              value={resourceKey}
              placeholder="reviewed-resource"
              disabled={pending || !planPermission.available}
              onChange={(event) => onResourceKeyChange(event.currentTarget.value)}
            />
            <small id={`${sectionId}-resource-hint`}>{t($ => $.httpTool.resourceHint)}</small>
          </label>
          <label className="workflow-executor-input-field">
            <span id={`${sectionId}-locale-label`}>{t($ => $.httpTool.locale)}</span>
            <input
              type="text"
              maxLength={35}
              autoComplete="off"
              aria-labelledby={`${sectionId}-locale-label`}
              aria-describedby={`${sectionId}-locale-hint`}
              value={locale}
              placeholder="zh-CN"
              disabled={pending || !planPermission.available}
              onChange={(event) => onLocaleChange(event.currentTarget.value)}
            />
            <small id={`${sectionId}-locale-hint`}>{t($ => $.httpTool.localeHint)}</small>
          </label>
        </div>
        <div className="workflow-executor-action-row">
          <button type="submit" disabled={!canCreate}>
            {consumerState.status === "creating" ? t($ => $.httpTool.creating) : t($ => $.httpTool.create)}
          </button>
          <button type="button" disabled={pending || !plan || !permissions.read.available} onClick={onReloadPlan}>
            {consumerState.status === "reading" ? t($ => $.httpTool.reading) : t($ => $.httpTool.reload)}
          </button>
        </div>
        {!argumentValidation.valid && (resourceKey || locale) ? (
          <p className="boundary-note" role="alert">{workflowHTTPToolArgumentsFeedback(t, argumentValidation)}</p>
        ) : null}
      </form>

      {!eligibility.eligible ? (
        <div className="workflow-executor-blocker-list" aria-label={t($ => $.httpTool.blockers)}>
          {eligibility.reasons.map((reason) => (
            <article key={reason.code}>
              <code>{reason.code}</code>
              <p>{workflowHTTPToolFailure(t, reason.code)}</p>
            </article>
          ))}
        </div>
      ) : null}

      <article className="workflow-executor-state-card" aria-live="polite">
        <div>
          <span>{t($ => $.httpTool.consumerState)}</span>
          <strong>{workflowHTTPToolStatus(t, consumerState.status)}</strong>
        </div>
        <p>{workflowHTTPToolActionFeedback(t, consumerState)}</p>
        <dl className="workflow-user-workspace-home-meta">
          <div><dt>{t($ => $.httpTool.failure)}</dt><dd>{consumerState.failureCode || t($ => $.httpTool.none)}</dd></div>
          <div><dt>{t($ => $.httpTool.request)}</dt><dd>{consumerState.requestId || t($ => $.httpTool.none)}</dd></div>
          <div><dt>{t($ => $.httpTool.audit)}</dt><dd>{consumerState.auditRef || t($ => $.httpTool.none)}</dd></div>
          <div><dt>{t($ => $.httpTool.source)}</dt><dd>{t($ => consumerState.mode === "disabled" ? $.httpTool.offline : $.httpTool.liveSource)}</dd></div>
        </dl>
      </article>

      {consumerState.failureCode && consumerState.summary ? <details><summary>{t($ => $.httpTool.originalDiagnostic)}</summary><p>{consumerState.summary}</p></details> : null}

      {plan ? (
        <article className="workflow-executor-record" aria-label={t($ => $.httpTool.planDetail)}>
          <div className="workflow-executor-record-heading">
            <div>
              <span>{t($ => $.httpTool.plan)}</span>
              <strong>{plan.planId}</strong>
            </div>
            <span className={`status-badge ${planStatusTone(plan.status)}`}>{workflowHTTPToolStatus(t, plan.status)}</span>
          </div>
          <dl className="workflow-executor-record-meta">
            <div><dt>{t($ => $.httpTool.recordVersion)}</dt><dd>{plan.recordVersion}</dd></div>
            <div><dt>{t($ => $.httpTool.source)}</dt><dd>{plan.sourceKind === "workflow_definition" ? `${plan.workflowDefinitionId} · v${plan.workflowDefinitionVersion} · pointer v${plan.activationPointerVersion}` : `${plan.draftId} · v${plan.draftVersion}`}</dd></div>
            <div><dt>{t($ => $.httpTool.node)}</dt><dd>{plan.nodeId}</dd></div>
            <div><dt>{t($ => $.httpTool.tool)}</dt><dd>{plan.toolId} · v{plan.toolVersion}</dd></div>
            <div><dt>{t($ => $.httpTool.method)}</dt><dd>{plan.method}</dd></div>
            <div><dt>{t($ => $.httpTool.targetPolicy)}</dt><dd>{plan.targetPolicyKey}</dd></div>
            <div><dt>{t($ => $.httpTool.resourceKey)}</dt><dd>{plan.publicArguments.resourceKey}</dd></div>
            <div><dt>{t($ => $.httpTool.locale)}</dt><dd>{plan.publicArguments.locale ?? t($ => $.httpTool.default)}</dd></div>
            <div><dt>{t($ => $.httpTool.outputFields)}</dt><dd>{plan.outputFields.join(", ")}</dd></div>
            <div><dt>{t($ => $.httpTool.outputDigest)}</dt><dd>{plan.outputSchemaDigest}</dd></div>
            <div><dt>{t($ => $.httpTool.timeout)}</dt><dd>{plan.timeoutMs} ms</dd></div>
            <div><dt>{t($ => $.httpTool.responseProjection)}</dt><dd>{plan.maxResponseBytes} / {plan.maxOutputBytes} bytes</dd></div>
            <div><dt>{t($ => $.httpTool.created)}</dt><dd>{plan.createdAt}</dd></div>
            <div><dt>{t($ => $.httpTool.expires)}</dt><dd>{plan.expiresAt}</dd></div>
            <div><dt>{t($ => $.httpTool.plannedBy)}</dt><dd>{plan.plannedByActorRef}</dd></div>
            <div><dt>{t($ => $.httpTool.lastDecision)}</dt><dd>{plan.lastDecisionByActorRef ?? t($ => $.httpTool.none)}</dd></div>
          </dl>

          <div className="workflow-executor-action-row" aria-label={t($ => $.httpTool.decisions)}>
            {decisions.map((decision) => (
              <button key={decision} type="button" disabled={pending || !permissions.confirm.available} onClick={() => onDecision(decision)}>
                {t($ => $.httpTool[decision])}
              </button>
            ))}
          </div>
          {plan.status === "approved" ? (
            <p className="boundary-note">
              {t($ => $.httpTool.approvedNote)}
            </p>
          ) : null}
        </article>
      ) : null}

      <ActionSafetyReadPanel projection={consumerState.actionSafety} title={t($ => $.httpTool.planSafety)} />

      {consumerState.confirmationDecision ? (
        <article className="workflow-executor-state-card" aria-label={t($ => $.httpTool.latestDecision)}>
          <div><span>{t($ => $.httpTool.latestDecision)}</span><strong>{workflowHTTPToolStatus(t, consumerState.confirmationDecision.outcome)}</strong></div>
          <p>
            {t($ => $.httpTool.decisionRecorded, { actor: consumerState.confirmationDecision.decidedByActorRef, source: t($ => $.httpTool[consumerState.confirmationDecision!.actorSource]), time: consumerState.confirmationDecision.decidedAt })}
          </p>
          <dl className="workflow-user-workspace-home-meta">
            <div><dt>{t($ => $.httpTool.confirmation)}</dt><dd>{consumerState.confirmationDecision.confirmationId}</dd></div>
            <div><dt>{t($ => $.httpTool.audit)}</dt><dd>{consumerState.confirmationDecision.auditRef}</dd></div>
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
  const { t } = useTranslation("workflow");
  return (
    <article>
      <span>{label}</span>
      <strong>{t($ => permission.available ? $.httpTool.grantAvailable : $.httpTool.grantUnavailable)}</strong>
      <p><code>{permission.requiredGrants.join(" + ")}</code></p>
      <small>{t($ => $.httpTool.grantHint)}</small>
    </article>
  );
}

function availableDecisions(status: string | null): WorkflowHTTPToolHumanDecision[] {
  if (status === "pending") return ["approve", "reject", "defer", "cancel"];
  if (status === "deferred") return ["approve", "reject", "cancel"];
  if (status === "approved") return ["cancel"];
  return [];
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
  const { t } = useTranslation("workflow");
  const tone = state.status === "failed" ? "bad" : state.status === "ready" ? "good" : "neutral";
  const label = state.mode === "disabled" ? t($ => $.httpTool.offline) : eligible ? workflowHTTPToolStatus(t, state.status) : t($ => $.httpTool.blocked);
  return <span className={`status-badge ${tone}`}>{label}</span>;
}
