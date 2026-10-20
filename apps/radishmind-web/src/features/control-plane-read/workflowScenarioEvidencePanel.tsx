import "../../i18n/readEvidenceResources.ts";
import { useTranslation } from "react-i18next";
import { workflowProjectionText } from "./workflowProjectionCopy.ts";
import { readEvidenceStatus } from "./readEvidenceMessages.ts";
import type { KeyboardEvent } from "react";
import type { WorkflowScenario, WorkflowScenarioBlockedReason, WorkflowScenarioExpectedOutput, WorkflowScenarioInputField, WorkflowScenarioInspectorViewModel, WorkflowScenarioRelation, WorkflowScenarioStatus, WorkflowScenarioStopLine, WorkflowScenarioSummary } from "./workflowScenarioInspector";

export function WorkflowScenarioInspectorPanel({
  inspector,
  selectedScenarioId,
  onSelectScenario,
}: {
  inspector: WorkflowScenarioInspectorViewModel;
  selectedScenarioId: string;
  onSelectScenario: (scenarioId: string) => void;
}) {
  const { t } = useTranslation("workflow");
  return (
    <section
      className="surface-band workflow-scenario-inspector"
      id="workflow-scenario-inspector"
      aria-labelledby="workflow-scenario-inspector-title"
    >
      <div className="section-heading">
        <div>
          <p className="eyebrow">{t($ => $.readEvidence.scenario)}</p>
          <h3 id="workflow-scenario-inspector-title">{t($ => $.readEvidence.scenarioTitle)}</h3>
        </div>
        <StatusBadge tone={inspector.canRenderScenarioInspector ? "neutral" : "bad"}>
          {inspector.canRenderScenarioInspector ? t($ => $.readEvidence.states.offline_advisory) : t($ => $.readEvidence.states.blocked)}
        </StatusBadge>
      </div>

      <div className="workflow-scenario-selector-grid" aria-label={t($ => $.readEvidence.scenarioSelector)}>
        {inspector.scenarios.map((scenario) => (
          <WorkflowScenarioSelectorCard
            key={scenario.scenarioId}
            scenario={scenario}
            selected={scenario.scenarioId === selectedScenarioId}
            onSelectScenario={onSelectScenario}
          />
        ))}
      </div>

      <div className="workflow-scenario-summary-grid" aria-label={t($ => $.readEvidence.scenarioSummary)}>
        {inspector.summary.map((summary, index) => (
          <WorkflowScenarioSummaryCard key={index} summary={summary} />
        ))}
      </div>

      <WorkflowScenarioDetailCard scenario={inspector.selectedScenario} inspector={inspector} />

      <div className="workflow-scenario-input-grid" aria-label={t($ => $.readEvidence.scenarioInput)}>
        {inspector.selectedScenario.inputContract.map((input) => (
          <WorkflowScenarioInputCard key={input.fieldId} input={input} />
        ))}
      </div>

      <div className="workflow-scenario-output-grid" aria-label={t($ => $.readEvidence.scenarioOutput)}>
        {inspector.selectedScenario.expectedOutputs.map((output) => (
          <WorkflowScenarioOutputCard key={output.outputId} output={output} />
        ))}
      </div>

      <div className="workflow-scenario-relation-grid" aria-label={t($ => $.readEvidence.scenarioRelations)}>
        {inspector.relationMap.map((relation) => (
          <WorkflowScenarioRelationCard key={relation.relationId} relation={relation} />
        ))}
      </div>

      <div className="workflow-scenario-blocked-grid" aria-label={t($ => $.readEvidence.scenarioBlocked)}>
        {inspector.blockedReasons.map((reason) => (
          <WorkflowScenarioBlockedReasonCard key={reason.reasonId} reason={reason} />
        ))}
      </div>

      <div className="workflow-scenario-stopline-grid" aria-label={t($ => $.readEvidence.scenarioStops)}>
        {inspector.stopLines.map((stopLine) => (
          <WorkflowScenarioStopLineCard key={stopLine.stopLineId} stopLine={stopLine} />
        ))}
      </div>
    </section>
  );
}

function WorkflowScenarioSelectorCard({
  scenario,
  selected,
  onSelectScenario,
}: {
  scenario: WorkflowScenario;
  selected: boolean;
  onSelectScenario: (scenarioId: string) => void;
}) {
  const { t } = useTranslation("workflow");
  return (
    <article
      className={`workflow-scenario-selector-card selection-row${selected ? " selected" : ""}`}
      role="button"
      tabIndex={0}
      aria-pressed={selected}
      onClick={() => onSelectScenario(scenario.scenarioId)}
      onKeyDown={(event) => handleSelectionRowKeyDown(event, scenario.scenarioId, onSelectScenario)}
    >
      <div className="workflow-scenario-row-main">
        <div>
          <p className="eyebrow">{readEvidenceStatus(t, scenario.scenarioKind)}</p>
          <h5>{workflowProjectionText(t, scenario, "label")}</h5>
        </div>
        <StatusBadge tone={selected ? "neutral" : workflowScenarioTone(scenario.requiresConfirmation ? "blocked" : "review_required")}>
          {selected ? t($ => $.readEvidence.states.selected) : scenario.requiresConfirmation ? t($ => $.readEvidence.states.confirmation) : t($ => $.readEvidence.states.advisory)}
        </StatusBadge>
      </div>
      <p>{workflowProjectionText(t, scenario, "triggerSummary")}</p>
      <small>{scenario.scenarioId}</small>
    </article>
  );
}

function WorkflowScenarioSummaryCard({ summary }: { summary: WorkflowScenarioSummary }) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-scenario-card">
      <span>{workflowProjectionText(t, summary, "label")}</span>
      <strong>{workflowProjectionText(t, summary, "value")}</strong>
      <StatusBadge tone={workflowScenarioTone(summary.status)}>{readEvidenceStatus(t, summary.status)}</StatusBadge>
      <p>{workflowProjectionText(t, summary, "summary")}</p>
    </article>
  );
}

function WorkflowScenarioDetailCard({
  scenario,
  inspector,
}: {
  scenario: WorkflowScenario;
  inspector: WorkflowScenarioInspectorViewModel;
}) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-scenario-card">
      <div className="workflow-scenario-row-main">
        <div>
          <p className="eyebrow">{readEvidenceStatus(t, inspector.scenarioMode)}</p>
          <h5>{workflowProjectionText(t, scenario, "intent")}</h5>
        </div>
        <StatusBadge tone={workflowScenarioTone(scenario.requiresConfirmation ? "blocked" : "review_required")}>
          {scenario.requiresConfirmation ? t($ => $.readEvidence.states.confirmation_required) : t($ => $.readEvidence.states.advisory_only)}
        </StatusBadge>
      </div>
      <dl className="workflow-run-guard-meta">
        <div>
          <dt>{t($ => $.readEvidence.application)}</dt>
          <dd>{scenario.applicationRef}</dd>
        </div>
        <div>
          <dt>{t($ => $.readEvidence.definition)}</dt>
          <dd>{scenario.workflowDefinitionId}</dd>
        </div>
        <div>
          <dt>{t($ => $.readEvidence.draft)}</dt>
          <dd>{scenario.draftId}</dd>
        </div>
        <div>
          <dt>{t($ => $.readEvidence.latestRun)}</dt>
          <dd>{scenario.runId}</dd>
        </div>
        <div>
          <dt>{t($ => $.readEvidence.risk)}</dt>
          <dd>{readEvidenceStatus(t, scenario.riskLevel)}</dd>
        </div>
        <div>
          <dt>{t($ => $.readEvidence.scenarioLabel)}</dt>
          <dd>{inspector.selectedScenarioId}</dd>
        </div>
      </dl>
      <p>{workflowProjectionText(t, scenario, "validationFocus")}</p>
    </article>
  );
}

function WorkflowScenarioInputCard({ input }: { input: WorkflowScenarioInputField }) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-scenario-input">
      <div className="workflow-scenario-row-main">
        <div>
          <p className="eyebrow">{input.fieldId}</p>
          <h5>{workflowProjectionText(t, input, "label")}</h5>
        </div>
        <StatusBadge tone="neutral">{input.required ? t($ => $.readEvidence.states.required) : t($ => $.readEvidence.states.optional)}</StatusBadge>
      </div>
      <dl className="workflow-run-guard-meta">
        <div>
          <dt>{t($ => $.readEvidence.source)}</dt>
          <dd>{input.sourceRef}</dd>
        </div>
      </dl>
      <p>{workflowProjectionText(t, input, "summary")}</p>
    </article>
  );
}

function WorkflowScenarioOutputCard({ output }: { output: WorkflowScenarioExpectedOutput }) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-scenario-output">
      <div className="workflow-scenario-row-main">
        <div>
          <p className="eyebrow">{output.outputId}</p>
          <h5>{workflowProjectionText(t, output, "label")}</h5>
        </div>
        <StatusBadge tone={workflowScenarioTone(output.status)}>{readEvidenceStatus(t, output.status)}</StatusBadge>
      </div>
      <p>{workflowProjectionText(t, output, "summary")}</p>
    </article>
  );
}

function WorkflowScenarioRelationCard({ relation }: { relation: WorkflowScenarioRelation }) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-scenario-relation">
      <div className="workflow-scenario-row-main">
        <div>
          <p className="eyebrow">{relation.relationId}</p>
          <h5>{workflowProjectionText(t, relation, "label")}</h5>
        </div>
        <StatusBadge tone={workflowScenarioTone(relation.status)}>{readEvidenceStatus(t, relation.status)}</StatusBadge>
      </div>
      <dl className="workflow-run-guard-meta">
        <div>
          <dt>{t($ => $.readEvidence.source)}</dt>
          <dd>{relation.sourceRef}</dd>
        </div>
        <div>
          <dt>{t($ => $.readEvidence.target)}</dt>
          <dd>{relation.targetRef}</dd>
        </div>
        <div>
          <dt>{t($ => $.readEvidence.audit)}</dt>
          <dd>{relation.auditRef}</dd>
        </div>
      </dl>
      <p>{workflowProjectionText(t, relation, "summary")}</p>
    </article>
  );
}

function WorkflowScenarioBlockedReasonCard({ reason }: { reason: WorkflowScenarioBlockedReason }) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-scenario-blocked-reason">
      <div className="workflow-scenario-row-main">
        <div>
          <p className="eyebrow">{readEvidenceStatus(t, reason.sourceSurface)}</p>
          <h5>{workflowProjectionText(t, reason, "label")}</h5>
        </div>
        <StatusBadge tone="bad">{readEvidenceStatus(t, reason.status)}</StatusBadge>
      </div>
      <dl className="workflow-run-guard-meta">
        <div>
          <dt>{t($ => $.readEvidence.missing)}</dt>
          <dd>{workflowProjectionText(t, reason, "missingPrerequisite")}</dd>
        </div>
        <div>
          <dt>{t($ => $.readEvidence.audit)}</dt>
          <dd>{reason.auditRef}</dd>
        </div>
      </dl>
      <p>{workflowProjectionText(t, reason, "summary")}</p>
    </article>
  );
}

function WorkflowScenarioStopLineCard({ stopLine }: { stopLine: WorkflowScenarioStopLine }) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-scenario-stopline">
      <div className="workflow-scenario-row-main">
        <div>
          <p className="eyebrow">{stopLine.stopLineId}</p>
          <h5>{workflowProjectionText(t, stopLine, "label")}</h5>
        </div>
        <StatusBadge tone="bad">{readEvidenceStatus(t, stopLine.status)}</StatusBadge>
      </div>
      <p>{workflowProjectionText(t, stopLine, "summary")}</p>
    </article>
  );
}

function workflowScenarioTone(status: WorkflowScenarioStatus): "good" | "bad" | "neutral" {
  if (status === "blocked") {
    return "bad";
  }
  if (status === "ready") {
    return "good";
  }
  return "neutral";
}

function handleSelectionRowKeyDown(
  event: KeyboardEvent<HTMLElement>,
  selectionId: string,
  onSelect: (selectionId: string) => void,
) {
  if (event.key !== "Enter" && event.key !== " ") {
    return;
  }
  event.preventDefault();
  onSelect(selectionId);
}

function StatusBadge({ children, tone }: { children: string; tone: "good" | "bad" | "neutral" }) {
  return <span className={`status-badge ${tone}`}>{children}</span>;
}
