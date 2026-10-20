import "../../i18n/workflowDraftResources.ts";
import { workflowProjectionText, workflowProjectionStatusLabel } from "./workflowProjectionCopy.ts";
import "../../i18n/workflowInspectionProjectionResources.ts";
import "../../i18n/workflowInspectionResources.ts";
import { useTranslation } from "react-i18next";
import {
  type WorkflowDraftBlockedCapabilityCheck,
  type WorkflowDraftContractCheck,
  type WorkflowDraftStructuralCheck,
  type WorkflowDraftValidationInspectorViewModel,
  type WorkflowDraftValidationSummary,
} from "./workflowDraftValidationInspector";
import {
  type WorkflowExecutionPlanBlockedReason,
  type WorkflowExecutionPlanGate,
  type WorkflowExecutionPlanNodeMapping,
  type WorkflowExecutionPlanPreviewViewModel,
  type WorkflowExecutionPlanProviderRequirement,
  type WorkflowExecutionPlanStage,
  type WorkflowExecutionPlanSummary,
} from "./workflowExecutionPlanPreview";
import {
  type WorkflowRuntimeReadinessBlocker,
  type WorkflowRuntimeReadinessGate,
  type WorkflowRuntimeReadinessInspectorViewModel,
  type WorkflowRuntimeReadinessPrerequisite,
  type WorkflowRuntimeReadinessStatus,
  type WorkflowRuntimeReadinessSummary,
} from "./workflowRuntimeReadinessInspector";

export function WorkflowDraftValidationInspectorPanel({
  inspector,
}: {
  inspector: WorkflowDraftValidationInspectorViewModel;
}) {
  const { t } = useTranslation("workflow");
  return (
    <div
      className="workflow-draft-validation-inspector"
      id="workflow-draft-validation-inspector"
      aria-label={t($ => $.inspection.workflowDraftValidationInspectorOfflineSurface)}
    >
      <p className="boundary-note">{t($ => $.inspection.sourceEvidence)}</p>
      <div className="section-heading compact-heading">
        <div>
          <p className="eyebrow">{t($ => $.inspection.draftValidationInspector)}</p>
          <h4>{inspector.inspectedDraftId}</h4>
        </div>
        <StatusBadge tone={inspector.validationStatus === "blocked" ? "bad" : "neutral"}>
          {inspector.validationStatus}
        </StatusBadge>
      </div>

      <div className="workflow-draft-validation-summary-grid" aria-label={t($ => $.inspection.workflowDraftValidationSummary)}>
        {inspector.summary.map((summary) => (
          <WorkflowDraftValidationSummaryCard key={summary.label} summary={summary} />
        ))}
      </div>

      <div className="workflow-draft-structural-check-grid" aria-label={t($ => $.inspection.workflowDraftStructuralChecks)}>
        {inspector.structuralChecks.map((check) => (
          <WorkflowDraftStructuralCheckCard key={check.checkId} check={check} />
        ))}
      </div>

      <div className="workflow-draft-contract-check-grid" aria-label={t($ => $.inspection.workflowDraftContractChecks)}>
        {inspector.contractChecks.map((check) => (
          <WorkflowDraftContractCheckCard key={check.checkId} check={check} />
        ))}
      </div>

      <div
        className="workflow-draft-validation-blocked-grid"
        aria-label={t($ => $.inspection.workflowDraftValidationBlockedCapabilityChecks)}
      >
        {inspector.blockedCapabilityChecks.map((check) => (
          <WorkflowDraftBlockedCapabilityCheckCard key={check.checkId} check={check} />
        ))}
      </div>

      <article className="workflow-draft-validation-card">
        <div className="workflow-draft-validation-row-main">
          <div>
            <p className="eyebrow">{inspector.auditMetadata.sourceRouteId}</p>
            <h5>{inspector.auditMetadata.draftRouteId}</h5>
          </div>
          <StatusBadge tone="neutral">{t($ => $.inspection.offline)}</StatusBadge>
        </div>
        <dl className="workflow-run-guard-meta">
          <div>
            <dt>{t($ => $.inspection.request)}</dt>
            <dd>{inspector.auditMetadata.requestId}</dd>
          </div>
          <div>
            <dt>{t($ => $.inspection.audit)}</dt>
            <dd>{inspector.auditMetadata.auditRef}</dd>
          </div>
          <div>
            <dt>{t($ => $.inspection.draft)}</dt>
            <dd>{inspector.auditMetadata.inspectedDraftId}</dd>
          </div>
        </dl>
      </article>
    </div>
  );
}

function WorkflowDraftValidationSummaryCard({ summary }: { summary: WorkflowDraftValidationSummary }) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-draft-validation-card">
      <span>{workflowProjectionText(t, summary, "label")}</span>
      <strong>{summary.value}</strong>
      <p>{workflowProjectionText(t, summary, "summary")}</p>
    </article>
  );
}

function WorkflowDraftStructuralCheckCard({ check }: { check: WorkflowDraftStructuralCheck }) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-draft-structural-check">
      <div className="workflow-draft-validation-row-main">
        <div>
          <p className="eyebrow">{check.checkId}</p>
          <h5>{workflowProjectionText(t, check, "label")}</h5>
        </div>
        <StatusBadge tone={check.status === "blocked" ? "bad" : check.status === "passed" ? "good" : "neutral"}>
          {check.status}
        </StatusBadge>
      </div>
      <p>{workflowProjectionText(t, check, "summary")}</p>
      <div className="workflow-draft-validation-evidence" aria-label={t($ => $.inspection.workflowDraftStructuralCheckEvidence)}>
        {check.evidenceRefs.map((evidenceRef, index) => (
          <code key={`${evidenceRef}:${index}`}>{evidenceRef}</code>
        ))}
      </div>
    </article>
  );
}

function WorkflowDraftContractCheckCard({ check }: { check: WorkflowDraftContractCheck }) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-draft-contract-check">
      <div className="workflow-draft-validation-row-main">
        <div>
          <p className="eyebrow">{check.checkId}</p>
          <h5>{workflowProjectionText(t, check, "label")}</h5>
        </div>
        <StatusBadge tone={check.status === "passed" ? "good" : "neutral"}>{check.status}</StatusBadge>
      </div>
      <p>{workflowProjectionText(t, check, "summary")}</p>
      <dl className="workflow-run-guard-meta">
        <div>
          <dt>{t($ => $.inspection.required)}</dt>
          <dd>{check.requiredFields.join(", ")}</dd>
        </div>
        <div>
          <dt>{t($ => $.inspection.present)}</dt>
          <dd>{check.presentFields.join(", ") || workflowProjectionStatusLabel(t, "none")}</dd>
        </div>
        <div>
          <dt>{t($ => $.inspection.missing)}</dt>
          <dd>{check.missingFields.join(", ") || workflowProjectionStatusLabel(t, "none")}</dd>
        </div>
      </dl>
    </article>
  );
}

function WorkflowDraftBlockedCapabilityCheckCard({
  check,
}: {
  check: WorkflowDraftBlockedCapabilityCheck;
}) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-draft-validation-blocked-check">
      <div className="workflow-draft-validation-row-main">
        <div>
          <p className="eyebrow">{check.capabilityId}</p>
          <h5>{workflowProjectionText(t, check, "label")}</h5>
        </div>
        <StatusBadge tone="bad">{check.status}</StatusBadge>
      </div>
      <dl className="workflow-run-guard-meta">
        <div>
          <dt>{t($ => $.inspection.missingPrerequisite)}</dt>
          <dd>{workflowProjectionText(t, check, "missingPrerequisite")}</dd>
        </div>
        <div>
          <dt>{t($ => $.inspection.audit)}</dt>
          <dd>{check.auditRef}</dd>
        </div>
      </dl>
      <p>{workflowProjectionText(t, check, "summary")}</p>
    </article>
  );
}

export function WorkflowExecutionPlanPreviewPanel({
  preview,
}: {
  preview: WorkflowExecutionPlanPreviewViewModel;
}) {
  const { t } = useTranslation("workflow");
  return (
    <div
      className="workflow-execution-plan-preview"
      id="workflow-execution-plan-preview"
      aria-label={t($ => $.inspection.workflowExecutionPlanPreviewOfflineSurface)}
    >
      <p className="boundary-note">{t($ => $.inspection.sourceEvidence)}</p>
      <div className="section-heading compact-heading">
        <div>
          <p className="eyebrow">{t($ => $.inspection.fullRuntimeExecutionPlanPreview)}</p>
          <h4>{preview.selectedDraftId}</h4>
        </div>
        <StatusBadge tone={preview.canRenderExecutionPlanPreview ? "neutral" : "bad"}>
          {preview.canRenderExecutionPlanPreview ? "offline_preview" : "blocked"}
        </StatusBadge>
      </div>

      <div className="workflow-execution-plan-summary-grid" aria-label={t($ => $.inspection.workflowExecutionPlanSummary)}>
        {preview.summary.map((summary) => (
          <WorkflowExecutionPlanSummaryCard key={summary.label} summary={summary} />
        ))}
      </div>

      <div className="workflow-execution-plan-stage-grid" aria-label={t($ => $.inspection.workflowExecutionPlanStageOrder)}>
        {preview.stageOrder.map((stage) => (
          <WorkflowExecutionPlanStageCard key={stage.stageId} stage={stage} />
        ))}
      </div>

      <div className="workflow-execution-plan-node-grid" aria-label={t($ => $.inspection.workflowExecutionPlanNodeToStageMapping)}>
        {preview.nodeStageMappings.map((mapping) => (
          <WorkflowExecutionPlanNodeMappingCard key={mapping.nodeId} mapping={mapping} />
        ))}
      </div>

      <div className="workflow-execution-plan-provider-grid" aria-label={t($ => $.inspection.workflowExecutionPlanProviderRequirements)}>
        {preview.providerProfileRequirements.map((requirement) => (
          <WorkflowExecutionPlanProviderRequirementCard
            key={requirement.requirementId}
            requirement={requirement}
          />
        ))}
      </div>

      <div className="workflow-execution-plan-gate-grid" aria-label={t($ => $.inspection.workflowExecutionPlanConfirmationAndAuditGates)}>
        {preview.confirmationAuditGates.map((gate) => (
          <WorkflowExecutionPlanGateCard key={gate.gateId} gate={gate} />
        ))}
      </div>

      <div className="workflow-execution-plan-blocked-grid" aria-label={t($ => $.inspection.workflowExecutionPlanBlockedReasons)}>
        {preview.blockedPlanReasons.map((reason) => (
          <WorkflowExecutionPlanBlockedReasonCard key={reason.reasonId} reason={reason} />
        ))}
      </div>

      <article className="workflow-execution-plan-card">
        <div className="workflow-execution-plan-row-main">
          <div>
            <p className="eyebrow">{preview.auditMetadata.sourceRouteId}</p>
            <h5>{preview.auditMetadata.draftRouteId}</h5>
          </div>
          <StatusBadge tone="neutral">{preview.validationStatus}</StatusBadge>
        </div>
        <dl className="workflow-run-guard-meta">
          <div>
            <dt>{t($ => $.inspection.validationRoute)}</dt>
            <dd>{preview.auditMetadata.validationRouteId}</dd>
          </div>
          <div>
            <dt>{t($ => $.inspection.request)}</dt>
            <dd>{preview.auditMetadata.requestId}</dd>
          </div>
          <div>
            <dt>{t($ => $.inspection.audit)}</dt>
            <dd>{preview.auditMetadata.auditRef}</dd>
          </div>
          <div>
            <dt>{t($ => $.inspection.draft)}</dt>
            <dd>{preview.auditMetadata.selectedDraftId}</dd>
          </div>
        </dl>
      </article>
    </div>
  );
}

function WorkflowExecutionPlanSummaryCard({ summary }: { summary: WorkflowExecutionPlanSummary }) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-execution-plan-card">
      <span>{workflowProjectionText(t, summary, "label")}</span>
      <strong>{summary.value}</strong>
      <p>{workflowProjectionText(t, summary, "summary")}</p>
    </article>
  );
}

function WorkflowExecutionPlanStageCard({ stage }: { stage: WorkflowExecutionPlanStage }) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-execution-plan-stage">
      <div className="workflow-execution-plan-row-main">
        <div>
          <p className="eyebrow">
            {stage.order} / {workflowProjectionStatusLabel(t, stage.stageKind)}
          </p>
          <h5>{workflowProjectionText(t, stage, "label")}</h5>
        </div>
        <StatusBadge tone={stage.status === "blocked" ? "bad" : stage.status === "ready" ? "good" : "neutral"}>
          {stage.status}
        </StatusBadge>
      </div>
      <p>{workflowProjectionText(t, stage, "summary")}</p>
      <dl className="workflow-run-guard-meta">
        <div>
          <dt>{t($ => $.inspection.nodes)}</dt>
          <dd>{stage.nodeIds.join(", ") || workflowProjectionStatusLabel(t, "none")}</dd>
        </div>
        <div>
          <dt>{t($ => $.inspection.blockedReason)}</dt>
          <dd>{workflowProjectionText(t, stage, "blockedReason")}</dd>
        </div>
      </dl>
    </article>
  );
}

function WorkflowExecutionPlanNodeMappingCard({ mapping }: { mapping: WorkflowExecutionPlanNodeMapping }) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-execution-plan-node">
      <div className="workflow-execution-plan-row-main">
        <div>
          <p className="eyebrow">{mapping.stageId}</p>
          <h5>{workflowProjectionText(t, mapping, "label")}</h5>
        </div>
        <StatusBadge tone={mapping.requiresConfirmation ? "bad" : "neutral"}>{mapping.executionMode}</StatusBadge>
      </div>
      <dl className="workflow-run-guard-meta">
        <div>
          <dt>{t($ => $.inspection.node)}</dt>
          <dd>{mapping.nodeId}</dd>
        </div>
        <div>
          <dt>{t($ => $.inspection.type)}</dt>
          <dd>{workflowProjectionStatusLabel(t, mapping.nodeType)}</dd>
        </div>
        <div>
          <dt>{t($ => $.inspection.provider)}</dt>
          <dd>{mapping.providerProfileRef}</dd>
        </div>
        <div>
          <dt>{t($ => $.inspection.input)}</dt>
          <dd>{mapping.inputSummary}</dd>
        </div>
        <div>
          <dt>{t($ => $.inspection.output)}</dt>
          <dd>{mapping.outputSummary}</dd>
        </div>
      </dl>
    </article>
  );
}

function WorkflowExecutionPlanProviderRequirementCard({
  requirement,
}: {
  requirement: WorkflowExecutionPlanProviderRequirement;
}) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-execution-plan-provider">
      <div className="workflow-execution-plan-row-main">
        <div>
          <p className="eyebrow">{requirement.requirementId}</p>
          <h5>{workflowProjectionText(t, requirement, "label")}</h5>
        </div>
        <StatusBadge tone={requirement.status === "blocked" ? "bad" : "neutral"}>{requirement.status}</StatusBadge>
      </div>
      <dl className="workflow-run-guard-meta">
        <div>
          <dt>{t($ => $.inspection.profile)}</dt>
          <dd>{requirement.providerProfileRef}</dd>
        </div>
        <div>
          <dt>{t($ => $.inspection.nodes)}</dt>
          <dd>{requirement.nodeIds.join(", ") || workflowProjectionStatusLabel(t, "none")}</dd>
        </div>
        <div>
          <dt>{t($ => $.inspection.missingPrerequisite)}</dt>
          <dd>{workflowProjectionText(t, requirement, "missingPrerequisite")}</dd>
        </div>
      </dl>
      <p>{workflowProjectionText(t, requirement, "summary")}</p>
    </article>
  );
}

function WorkflowExecutionPlanGateCard({ gate }: { gate: WorkflowExecutionPlanGate }) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-execution-plan-gate">
      <div className="workflow-execution-plan-row-main">
        <div>
          <p className="eyebrow">{workflowProjectionStatusLabel(t, gate.gateKind)}</p>
          <h5>{workflowProjectionText(t, gate, "label")}</h5>
        </div>
        <StatusBadge tone={gate.status === "blocked" ? "bad" : "neutral"}>{gate.status}</StatusBadge>
      </div>
      <dl className="workflow-run-guard-meta">
        <div>
          <dt>{t($ => $.inspection.beforeStage)}</dt>
          <dd>{gate.requiredBeforeStageId}</dd>
        </div>
        <div>
          <dt>{t($ => $.inspection.audit)}</dt>
          <dd>{gate.auditRef}</dd>
        </div>
      </dl>
      <p>{workflowProjectionText(t, gate, "summary")}</p>
    </article>
  );
}

function WorkflowExecutionPlanBlockedReasonCard({
  reason,
}: {
  reason: WorkflowExecutionPlanBlockedReason;
}) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-execution-plan-blocked-reason">
      <div className="workflow-execution-plan-row-main">
        <div>
          <p className="eyebrow">{workflowProjectionStatusLabel(t, reason.blockedCapability)}</p>
          <h5>{workflowProjectionText(t, reason, "label")}</h5>
        </div>
        <StatusBadge tone="bad">{reason.status}</StatusBadge>
      </div>
      <dl className="workflow-run-guard-meta">
        <div>
          <dt>{t($ => $.inspection.missingPrerequisite)}</dt>
          <dd>{workflowProjectionText(t, reason, "missingPrerequisite")}</dd>
        </div>
        <div>
          <dt>{t($ => $.inspection.audit)}</dt>
          <dd>{reason.auditRef}</dd>
        </div>
      </dl>
      <p>{workflowProjectionText(t, reason, "summary")}</p>
    </article>
  );
}

export function WorkflowRuntimeReadinessInspectorPanel({
  readiness,
}: {
  readiness: WorkflowRuntimeReadinessInspectorViewModel;
}) {
  const { t } = useTranslation("workflow");
  return (
    <div
      className="workflow-runtime-readiness-inspector"
      id="workflow-runtime-readiness-inspector"
      aria-label={t($ => $.inspection.workflowRuntimeReadinessInspectorOfflineSurface)}
    >
      <p className="boundary-note">{t($ => $.inspection.sourceEvidence)}</p>
      <div className="section-heading compact-heading">
        <div>
          <p className="eyebrow">{t($ => $.inspection.fullRuntimeReadinessInspector)}</p>
          <h4>{readiness.selectedDraftId}</h4>
        </div>
        <StatusBadge tone={readiness.canRenderRuntimeReadinessInspector ? "bad" : "neutral"}>
          {readiness.canRenderRuntimeReadinessInspector ? "blocked_readiness" : "missing_evidence"}
        </StatusBadge>
      </div>

      <div className="workflow-runtime-readiness-summary-grid" aria-label={t($ => $.inspection.workflowRuntimeReadinessSummary)}>
        {readiness.summary.map((summary) => (
          <WorkflowRuntimeReadinessSummaryCard key={summary.label} summary={summary} />
        ))}
      </div>

      <div className="workflow-runtime-readiness-prerequisite-grid" aria-label={t($ => $.inspection.workflowRuntimePrerequisites)}>
        {readiness.runtimePrerequisites.map((prerequisite) => (
          <WorkflowRuntimeReadinessPrerequisiteCard
            key={prerequisite.prerequisiteId}
            prerequisite={prerequisite}
          />
        ))}
      </div>

      <div className="workflow-runtime-readiness-blocker-grid" aria-label={t($ => $.inspection.workflowRuntimeReadinessBlockers)}>
        {readiness.readinessBlockers.map((blocker) => (
          <WorkflowRuntimeReadinessBlockerCard key={blocker.blockerId} blocker={blocker} />
        ))}
      </div>

      <div className="workflow-runtime-readiness-gate-grid" aria-label={t($ => $.inspection.workflowRuntimeImplementationGates)}>
        {readiness.implementationGates.map((gate) => (
          <WorkflowRuntimeReadinessGateCard key={gate.gateId} gate={gate} />
        ))}
      </div>

      <article className="workflow-runtime-readiness-card">
        <div className="workflow-runtime-readiness-row-main">
          <div>
            <p className="eyebrow">{readiness.auditMetadata.sourcePageId}</p>
            <h5>{readiness.auditMetadata.readinessRouteId}</h5>
          </div>
          <StatusBadge tone={readiness.forbiddenProjectionBlocked ? "bad" : "neutral"}>
            {readiness.forbiddenProjectionBlocked ? "guard_active" : "metadata_only"}
          </StatusBadge>
        </div>
        <dl className="workflow-run-guard-meta">
          <div>
            <dt>{t($ => $.inspection.planRoute)}</dt>
            <dd>{readiness.auditMetadata.planRouteId}</dd>
          </div>
          <div>
            <dt>{t($ => $.inspection.request)}</dt>
            <dd>{readiness.auditMetadata.requestId}</dd>
          </div>
          <div>
            <dt>{t($ => $.inspection.audit)}</dt>
            <dd>{readiness.auditMetadata.auditRef}</dd>
          </div>
          <div>
            <dt>{t($ => $.inspection.draft)}</dt>
            <dd>{readiness.auditMetadata.selectedDraftId}</dd>
          </div>
        </dl>
      </article>
    </div>
  );
}

function WorkflowRuntimeReadinessSummaryCard({ summary }: { summary: WorkflowRuntimeReadinessSummary }) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-runtime-readiness-card">
      <span>{workflowProjectionText(t, summary, "label")}</span>
      <strong>{summary.value}</strong>
      <p>{workflowProjectionText(t, summary, "summary")}</p>
    </article>
  );
}

function WorkflowRuntimeReadinessPrerequisiteCard({
  prerequisite,
}: {
  prerequisite: WorkflowRuntimeReadinessPrerequisite;
}) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-runtime-readiness-prerequisite">
      <div className="workflow-runtime-readiness-row-main">
        <div>
          <p className="eyebrow">{workflowProjectionStatusLabel(t, prerequisite.area)}</p>
          <h5>{workflowProjectionText(t, prerequisite, "label")}</h5>
        </div>
        <StatusBadge tone={workflowRuntimeReadinessTone(prerequisite.status)}>{prerequisite.status}</StatusBadge>
      </div>
      <dl className="workflow-run-guard-meta">
        <div>
          <dt>{t($ => $.inspection.evidence)}</dt>
          <dd>{workflowProjectionText(t, prerequisite, "currentEvidence")}</dd>
        </div>
        <div>
          <dt>{t($ => $.inspection.missingPrerequisite)}</dt>
          <dd>{workflowProjectionText(t, prerequisite, "missingPrerequisite")}</dd>
        </div>
        <div>
          <dt>{t($ => $.inspection.sourceRefs)}</dt>
          <dd>{prerequisite.sourceRefs.join(", ")}</dd>
        </div>
      </dl>
      <p>{workflowProjectionText(t, prerequisite, "summary")}</p>
    </article>
  );
}

function WorkflowRuntimeReadinessBlockerCard({ blocker }: { blocker: WorkflowRuntimeReadinessBlocker }) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-runtime-readiness-blocker">
      <div className="workflow-runtime-readiness-row-main">
        <div>
          <p className="eyebrow">{workflowProjectionStatusLabel(t, blocker.area)}</p>
          <h5>{workflowProjectionText(t, blocker, "label")}</h5>
        </div>
        <StatusBadge tone="bad">{workflowProjectionStatusLabel(t, blocker.severity)}</StatusBadge>
      </div>
      <dl className="workflow-run-guard-meta">
        <div>
          <dt>{t($ => $.inspection.source)}</dt>
          <dd>{blocker.sourceRef}</dd>
        </div>
        <div>
          <dt>{t($ => $.inspection.missingPrerequisite)}</dt>
          <dd>{workflowProjectionText(t, blocker, "missingPrerequisite")}</dd>
        </div>
        <div>
          <dt>{t($ => $.inspection.audit)}</dt>
          <dd>{blocker.auditRef}</dd>
        </div>
      </dl>
      <p>{workflowProjectionText(t, blocker, "summary")}</p>
    </article>
  );
}

function WorkflowRuntimeReadinessGateCard({ gate }: { gate: WorkflowRuntimeReadinessGate }) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-runtime-readiness-gate">
      <div className="workflow-runtime-readiness-row-main">
        <div>
          <p className="eyebrow">{workflowProjectionStatusLabel(t, gate.gateKind)}</p>
          <h5>{workflowProjectionText(t, gate, "label")}</h5>
        </div>
        <StatusBadge tone={workflowRuntimeReadinessTone(gate.status)}>{gate.status}</StatusBadge>
      </div>
      <dl className="workflow-run-guard-meta">
        <div>
          <dt>{t($ => $.inspection.requiredBefore)}</dt>
          <dd>{workflowProjectionStatusLabel(t, gate.requiredBefore)}</dd>
        </div>
        <div>
          <dt>{t($ => $.inspection.evidenceRefs)}</dt>
          <dd>{gate.evidenceRefs.join(", ")}</dd>
        </div>
      </dl>
      <p>{workflowProjectionText(t, gate, "summary")}</p>
    </article>
  );
}

function workflowRuntimeReadinessTone(status: WorkflowRuntimeReadinessStatus): "good" | "bad" | "neutral" {
  if (status === "blocked") {
    return "bad";
  }
  if (status === "satisfied") {
    return "good";
  }
  return "neutral";
}


function StatusBadge({ children, tone }: { children: string; tone: "good" | "bad" | "neutral" }) {
  const { t } = useTranslation("workflow");
  return <span className={`status-badge ${tone}`}>{workflowProjectionStatusLabel(t, children)}</span>;
}
