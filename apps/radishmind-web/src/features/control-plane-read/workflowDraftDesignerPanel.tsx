import { formatDisplayDate } from "../../i18n/formatters.ts";
import { workflowDraftStatusLabel, workflowDraftConsumerMessage, workflowDraftNodeTypeLabel, workflowDraftNodeTypeSummary } from "./workflowDraftMessages.ts";
import { useTranslation } from "react-i18next";
import "../../i18n/workflowDraftResources.ts";
import { lazy, Suspense } from "react";
import { applyWorkflowDraftEdit, canRemoveWorkflowDraftNode, WORKFLOW_DRAFT_NODE_TYPE_OPTIONS, type WorkflowDraftEdit } from "./workflowDraftEditing.ts";

import {
  workflowSavedDraftConflictRequiresResolution,
  type WorkflowSavedDraftConflictReviewSummary,
  type WorkflowSavedDraftConsumerState,
  type WorkflowSavedDraftSummary,
} from "./savedWorkflowDraftConsumer";
import {
  type WorkflowDraftDesignerBlockedCapability,
  type WorkflowDraftDesignerDraft,
  type WorkflowDraftDesignerEdge,
  type WorkflowDraftDesignerNode,
  type WorkflowDraftDesignerReadiness,
  type WorkflowDraftDesignerRisk,
  type WorkflowDraftDesignerTemplate,
  type WorkflowDraftDesignerViewModel,
} from "./workflowDraftDesigner";
import type { WorkflowDraftValidationInspectorViewModel } from "./workflowDraftValidationInspector";
import { canDeriveSavedWorkflowDraft } from "./workflowSavedDraftDerivation";

const WorkflowNodeDesigner = lazy(() =>
  import("./workflowNodeDesigner").then((module) => ({ default: module.WorkflowNodeDesigner })),
);

type WorkflowDraftNodeMoveDirection = "up" | "down";

export type WorkflowDraftNodeTypeOption = {
  nodeType: WorkflowDraftDesignerNode["nodeType"];
  lane: WorkflowDraftDesignerNode["lane"];
  label: string;
  summary: string;
};

type WorkflowDraftDesignerPanelProps = {
  designer: WorkflowDraftDesignerViewModel;
  selectedDraft: WorkflowDraftDesignerDraft;
  validationInspector: WorkflowDraftValidationInspectorViewModel;
  selectedDraftId: string;
  savedDraftConsumerState: WorkflowSavedDraftConsumerState;
  savedDraftConflictReviewSummary: WorkflowSavedDraftConflictReviewSummary | null;
  savedDraftConflictOpenSummary: WorkflowSavedDraftSummary | null;
  draftEditDirty: boolean;
  executorOperationPending: boolean;
  onSelectDraft: (draftId: string) => void;
  onEditDraft: (update: (draft: WorkflowDraftDesignerDraft) => WorkflowDraftDesignerDraft) => boolean;
  onResetDraftEdits: () => void;
  onContinueLocalDraftAfterConflict: () => void;
  onOpenConflictSavedDraft: () => void;
  onDeriveSavedDraft: () => void;
  onValidateDraft: () => void;
  onSaveDraft: () => void;
  onReadDraft: () => void;
};

export function WorkflowDraftDesignerPanel({
  designer,
  selectedDraft,
  validationInspector,
  selectedDraftId,
  savedDraftConsumerState,
  savedDraftConflictReviewSummary,
  savedDraftConflictOpenSummary,
  draftEditDirty,
  executorOperationPending,
  onSelectDraft,
  onEditDraft,
  onResetDraftEdits,
  onContinueLocalDraftAfterConflict,
  onOpenConflictSavedDraft,
  onDeriveSavedDraft,
  onValidateDraft,
  onSaveDraft,
  onReadDraft,
}: WorkflowDraftDesignerPanelProps) {
  const { t, i18n } = useTranslation("workflow");
  const nodeTypeOptions = WORKFLOW_DRAFT_NODE_TYPE_OPTIONS;
  const canRemoveNode = (nodeId: string) => canRemoveWorkflowDraftNode(selectedDraft, nodeId);
  const editDraft = (edit: WorkflowDraftEdit) => onEditDraft((draft) => applyWorkflowDraftEdit(draft, edit));
  const onUpdateDraftLabel: (label: string) => void = (label) => editDraft({ type: "label", label });
  const onUpdateDraftSummary: (summary: string) => void = (summary) => editDraft({ type: "summary", summary });
  const onUpdateNodeLabel: (nodeId: string, label: string) => void = (nodeId, label) => editDraft({ type: "node", nodeId, patch: { label } });
  const onUpdateNodeInputSummary: (nodeId: string, inputSummary: string) => void = (nodeId, inputSummary) => editDraft({ type: "node", nodeId, patch: { inputSummary } });
  const onUpdateNodeOutputSummary: (nodeId: string, outputSummary: string) => void = (nodeId, outputSummary) => editDraft({ type: "node", nodeId, patch: { outputSummary } });
  const onUpdateNodeProviderRef: (nodeId: string, providerRef: string) => void = (nodeId, providerRef) => editDraft({ type: "node", nodeId, patch: { providerRef } });
  const onUpdateNodeToolRef: (nodeId: string, toolRef: string) => void = (nodeId, toolRef) => editDraft({ type: "node", nodeId, patch: { toolRef } });
  const onUpdateNodeRagRef: (nodeId: string, ragRef: string) => void = (nodeId, ragRef) => editDraft({ type: "node", nodeId, patch: { ragRef } });
  const onUpdateNodeInputFields: (nodeId: string, inputFieldsText: string) => void = (nodeId, text) => editDraft({ type: "input_fields", nodeId, text });
  const onUpdateNodeOutputFields: (nodeId: string, outputFieldsText: string) => void = (nodeId, text) => editDraft({ type: "output_fields", nodeId, text });
  const onUpdateNodeOutputMapping: (nodeId: string, outputMappingSummary: string) => void = (nodeId, outputMappingSummary) => editDraft({ type: "node", nodeId, patch: { outputMappingSummary } });
  const onUpdateNodeDesignerPosition: (nodeId: string, x: number, y: number) => void = (nodeId, x, y) => editDraft({ type: "position", nodeId, x, y });
  const onUpdateEdgeCondition: (edgeId: string, conditionSummary: string) => void = (edgeId, conditionSummary) => editDraft({ type: "edge_condition", edgeId, conditionSummary });
  const onAddEdge: (fromNodeId: string, toNodeId: string) => boolean = (fromNodeId, toNodeId) => editDraft({ type: "add_edge", fromNodeId, toNodeId });
  const onRemoveEdge: (edgeId: string) => boolean = (edgeId) => editDraft({ type: "remove_edge", edgeId });
  const onAddNode: (nodeType: WorkflowDraftDesignerNode["nodeType"]) => void = (nodeType) => editDraft({ type: "add_node", nodeType });
  const onMoveNode: (nodeId: string, direction: WorkflowDraftNodeMoveDirection) => void = (nodeId, direction) => editDraft({ type: "move_node", nodeId, direction });
  const onRemoveNode: (nodeId: string) => void = (nodeId) => editDraft({ type: "remove_node", nodeId });
  const canCallDevConsumer = savedDraftConsumerState.mode === "dev_saved_draft_http";
  const operationPending = ["saving", "validating", "reading"].includes(savedDraftConsumerState.status);
  const conflictRequiresResolution = workflowSavedDraftConflictRequiresResolution(savedDraftConsumerState);
  const lifecycleReadOnly =
    savedDraftConsumerState.currentDraftVersion > 0 &&
    savedDraftConsumerState.currentLifecycleState !== "active";
  const lifecycleReadOnlyLabel = savedDraftConsumerState.currentLifecycleState === "unknown"
    ? t($ => $.draft.reopenRequired)
    : t($ => $.draft.archivedReview);
  const requestInteractionDisabled = operationPending || conflictRequiresResolution || executorOperationPending;
  const interactionDisabled = requestInteractionDisabled || lifecycleReadOnly;
  const editStateLabel = draftEditDirty ? t($ => $.draft.unsavedLocal) : workflowDraftStatusLabel(t, selectedDraft.localOnlyInteraction);
  const conflictOpenUnavailableMessage =
    t($ => $.draft.openUnavailable, { status: workflowDraftStatusLabel(t, savedDraftConflictReviewSummary?.savedMetadataState ?? "refreshing") });

  return (
    <section
      className="workflow-draft-designer workflow-designer-workbench"
      id="workflow-draft-designer"
      aria-label={t($ => $.draft.workbench)}
    >
      <header className="workflow-designer-header">
        <div className="workflow-designer-breadcrumb-row">
          <span>{t($ => $.draft.workflows)}</span>
          <span aria-hidden="true">/</span>
          <strong>{selectedDraft.workflowDefinitionId}</strong>
          <code>{selectedDraft.routeMetadata.routePath}</code>
        </div>
        <div className="workflow-designer-title-row">
          <div>
            <h4>{t($ => $.draft.designer)}</h4>
            <p>{selectedDraft.label}</p>
          </div>
          <StatusBadge tone={designer.canRenderDraftDesigner ? "good" : "bad"}>
            {lifecycleReadOnly
              ? lifecycleReadOnlyLabel
              : designer.canRenderDraftDesigner
                ? t($ => $.draft.developmentDesigner)
                : t($ => $.draft.blocked)}
          </StatusBadge>
        </div>
        <dl className="workflow-designer-context" aria-label={t($ => $.draft.activeContext)}>
          <div>
            <dt>{t($ => $.draft.application)}</dt>
            <dd>{selectedDraft.applicationRef}</dd>
          </div>
          <div>
            <dt>{t($ => $.draft.draft)}</dt>
            <dd>{selectedDraft.draftId}</dd>
          </div>
          <div>
            <dt>{t($ => $.draft.version)}</dt>
            <dd>{t($ => $.draft.versions, { content: savedDraftConsumerState.currentDraftVersion, lifecycle: savedDraftConsumerState.currentLifecycleVersion })}</dd>
          </div>
          <div>
            <dt>{t($ => $.draft.lifecycle)}</dt>
            <dd>{workflowDraftStatusLabel(t, savedDraftConsumerState.currentLifecycleState)}</dd>
          </div>
          <div className={draftEditDirty ? "attention" : "neutral"}>
            <dt>{t($ => $.draft.editState)}</dt>
            <dd>{editStateLabel}</dd>
          </div>
        </dl>
        <div className="workflow-designer-actions" aria-label={t($ => $.draft.draftActions)}>
          <button
            type="button"
            disabled={!canCallDevConsumer || interactionDisabled}
            onClick={onValidateDraft}
          >{t($ => $.draft.validate)}</button>
          <button
            type="button"
            className="primary"
            disabled={!canCallDevConsumer || interactionDisabled}
            onClick={onSaveDraft}
          >{t($ => $.draft.saveDraft)}</button>
          <button
            type="button"
            disabled={!canCallDevConsumer || requestInteractionDisabled}
            onClick={onReadDraft}
          >{t($ => $.draft.readSaved)}</button>
        </div>
        {lifecycleReadOnly ? (
          <p className="workflow-draft-revision-stopline" role="status">
            {savedDraftConsumerState.currentLifecycleState === "unknown"
              ? t($ => $.draft.reopenExplanation)
              : t($ => $.draft.archivedExplanation)}
          </p>
        ) : null}
      </header>

      <div className="workflow-designer-primary-layout">
        <aside className="workflow-designer-rail" aria-label={t($ => $.draft.referenceRail)}>
          <div className="workflow-designer-rail-heading">
            <div>
              <span>{t($ => $.draft.references)}</span>
              <strong>{t($ => $.draft.availableCount, { count: designer.templates.length })}</strong>
            </div>
            <a href="#workflow-user-workspace-home">{t($ => $.draft.openLibrary)}</a>
          </div>
          <div className="workflow-draft-template-grid" aria-label={t($ => $.draft.templates)}>
            {designer.templates.map((template) => (
              <WorkflowDraftTemplateButton
                key={template.draftId}
                template={template}
                selected={template.draftId === selectedDraftId}
                disabled={interactionDisabled}
                onSelectDraft={onSelectDraft}
              />
            ))}
          </div>

          <details className="workflow-designer-node-palette">
            <summary className="workflow-designer-rail-heading">
              <div>
                <span>{t($ => $.draft.addNode)}</span>
                <strong>{t($ => $.draft.nodeCount, { count: selectedDraft.nodes.length })}</strong>
              </div>
              <small>{t($ => $.draft.expand)}</small>
            </summary>
            <div className="workflow-draft-add-node-grid" aria-label={t($ => $.draft.addDraftNode)}>
              {nodeTypeOptions.map((option) => (
                <button
                  key={option.nodeType}
                  type="button"
                  className="workflow-draft-node-type-button"
                  disabled={interactionDisabled}
                  onClick={() => onAddNode(option.nodeType)}
                >
                  <span>{workflowDraftStatusLabel(t, option.lane)}</span>
                  <strong>{workflowDraftNodeTypeLabel(t, option.nodeType)}</strong>
                  <small>{workflowDraftNodeTypeSummary(t, option.nodeType)}</small>
                </button>
              ))}
            </div>
          </details>
        </aside>

        <div className="workflow-designer-canvas-column">
          <Suspense
            fallback={(
              <section className="workflow-node-designer-shell" aria-label={t($ => $.draft.loadingDesigner)}>
                <p>{t($ => $.draft.loadingDesignerText)}</p>
              </section>
            )}
          >
            <WorkflowNodeDesigner
              key={selectedDraft.draftId}
              draft={selectedDraft}
              validationInspector={validationInspector}
              editingDisabled={interactionDisabled}
              canRemoveNode={canRemoveNode}
              onUpdateNodeLabel={onUpdateNodeLabel}
              onUpdateNodeInputSummary={onUpdateNodeInputSummary}
              onUpdateNodeOutputSummary={onUpdateNodeOutputSummary}
              onUpdateNodeProviderRef={onUpdateNodeProviderRef}
              onUpdateNodeToolRef={onUpdateNodeToolRef}
              onUpdateNodeRagRef={onUpdateNodeRagRef}
              onUpdateNodeOutputMapping={onUpdateNodeOutputMapping}
              onUpdateNodeDesignerPosition={onUpdateNodeDesignerPosition}
              onAddEdge={onAddEdge}
              onRemoveEdge={onRemoveEdge}
              onRemoveNode={onRemoveNode}
            />
          </Suspense>
        </div>

        <details className="workflow-designer-review-dock">
          <summary>
            <span>{t($ => $.draft.reviewSurfaces)}</span>
            <strong>
              {t($ => $.draft.validationCount, { status: workflowDraftStatusLabel(t, validationInspector.validationStatus), count: validationInspector.structuralChecks.length + validationInspector.contractChecks.length })}
            </strong>
            <small>{t($ => $.draft.reviewBoundary)}</small>
          </summary>
          <nav className="workflow-designer-review-links" aria-label={t($ => $.draft.reviewNavigation)}>
            <a href="#workflow-draft-validation-inspector">
              <span>{t($ => $.draft.validation)}</span>
              <strong>{workflowDraftStatusLabel(t, validationInspector.validationStatus)}</strong>
              <small>{t($ => $.draft.checkCount, { count: validationInspector.structuralChecks.length + validationInspector.contractChecks.length })}</small>
            </a>
            <a href="#workflow-execution-plan-preview">
              <span>{t($ => $.draft.previewPlan)}</span>
              <strong>{t($ => $.draft.derivedOnly)}</strong>
              <small>{t($ => $.draft.noExecutor)}</small>
            </a>
            <a href="#workflow-runtime-readiness-inspector">
              <span>{t($ => $.draft.readiness)}</span>
              <strong>{t($ => $.draft.readOnly)}</strong>
              <small>{t($ => $.draft.checkCount, { count: selectedDraft.readiness.length })}</small>
            </a>
            <a href="#workflow-review-handoff">
              <span>{t($ => $.draft.reviewHandoff)}</span>
              <strong>{t($ => $.draft.browserOnly)}</strong>
              <small>{t($ => $.draft.handoffBoundary)}</small>
            </a>
          </nav>
        </details>
      </div>

      {savedDraftConflictReviewSummary ? (
        <details className="workflow-designer-disclosure workflow-draft-conflict-review" open>
          <summary>
            <span>{t($ => $.draft.conflictReview)}</span>
            <StatusBadge
              tone={savedDraftConflictReviewSummary.status === "local_draft_continued" ? "neutral" : "bad"}
            >
              {workflowDraftStatusLabel(t, savedDraftConflictReviewSummary.status)}
            </StatusBadge>
          </summary>
          <article className="workflow-draft-card workflow-draft-conflict-review-card">
            <dl className="workflow-run-guard-meta">
              <div><dt>{t($ => $.draft.localDraft)}</dt><dd>{savedDraftConflictReviewSummary.draftId}</dd></div>
              <div><dt>{t($ => $.draft.savedVersion)}</dt><dd>{savedDraftConflictReviewSummary.savedDraftVersion}</dd></div>
              <div><dt>{t($ => $.draft.updated)}</dt><dd title={savedDraftConflictReviewSummary.savedUpdatedAt}>{formatDisplayDate(savedDraftConflictReviewSummary.savedUpdatedAt, i18n.language === "en-US" ? "en-US" : "zh-CN") ?? t($ => $.draft.notLoaded)}</dd></div>
              <div><dt>{t($ => $.draft.actor)}</dt><dd>{savedDraftConflictReviewSummary.savedUpdatedByActorRef}</dd></div>
              <div><dt>{t($ => $.draft.validation)}</dt><dd>{workflowDraftStatusLabel(t, savedDraftConflictReviewSummary.savedValidationState)}</dd></div>
              <div><dt>{t($ => $.draft.blocked)}</dt><dd>{savedDraftConflictReviewSummary.savedBlockedCapabilityCount ?? t($ => $.draft.notLoaded)}</dd></div>
              <div><dt>{t($ => $.draft.metadata)}</dt><dd>{workflowDraftStatusLabel(t, savedDraftConflictReviewSummary.savedMetadataState)}</dd></div>
              <div><dt>{t($ => $.draft.open)}</dt><dd>{workflowDraftStatusLabel(t, savedDraftConflictReviewSummary.openActionState)}</dd></div>
            </dl>
            <p>{t($ => $.draft.conflictSummary, { draft: savedDraftConflictReviewSummary.draftId, version: savedDraftConflictReviewSummary.savedDraftVersion })}</p>
            <p>{savedDraftConflictReviewSummary.status === "local_draft_continued" ? t($ => $.draft.continuedExplanation) : t($ => $.draft.conflictPreserved)}</p>
            <div className="workflow-workspace-review-token-list" aria-label={t($ => $.draft.conflictLocks)}>
              <code>auto_overwrite_locked</code>
              <code>auto_merge_locked</code>
              <code>{savedDraftConflictReviewSummary.openActionState}</code>
            </div>
            <div className="workflow-draft-conflict-action-row" aria-label={t($ => $.draft.conflictActions)}>
              <button
                type="button"
                disabled={operationPending || savedDraftConflictReviewSummary.status === "local_draft_continued"}
                onClick={onContinueLocalDraftAfterConflict}
              >{t($ => $.draft.continueLocal)}</button>
              <button
                type="button"
                disabled={operationPending || !savedDraftConflictOpenSummary || !savedDraftConflictReviewSummary.canOpenSavedDraft}
                onClick={onOpenConflictSavedDraft}
              >{t($ => $.draft.openSaved)}</button>
            </div>
            <p>
              {savedDraftConflictReviewSummary.canOpenSavedDraft
                ? t($ => $.draft.openExplanation)
                : conflictOpenUnavailableMessage}
            </p>
            <p>{t($ => $.draft.conflictNext)}</p>

          </article>
        </details>
      ) : null}

      <details className="workflow-designer-disclosure">
        <summary>
          <span>{t($ => $.draft.persistenceBoundary)}</span>
          <StatusBadge tone={workflowSavedDraftConsumerTone(savedDraftConsumerState.status)}>
            {workflowDraftStatusLabel(t, savedDraftConsumerState.status)}
          </StatusBadge>
        </summary>
        <div className="workflow-draft-summary-grid" aria-label={t($ => $.draft.selectedSummary)}>
          <WorkflowDraftFact label={t($ => $.draft.draft)} value={selectedDraft.draftId} detail={selectedDraft.summary} />
          <WorkflowDraftFact label={t($ => $.draft.route)} value={selectedDraft.routeMetadata.draftRouteId} detail={selectedDraft.routeMetadata.routePath} />
          <WorkflowDraftFact label={t($ => $.draft.source)} value={selectedDraft.routeMetadata.sourceRouteId} detail={selectedDraft.workflowDefinitionId} />
          <WorkflowDraftFact label={t($ => $.draft.request)} value={selectedDraft.routeMetadata.requestId} detail={selectedDraft.routeMetadata.auditRef} />
          <WorkflowDraftFact label={t($ => $.draft.savedState)} value={workflowDraftStatusLabel(t, savedDraftConsumerState.status)} detail={workflowDraftConsumerMessage(t, savedDraftConsumerState)} />
          <WorkflowDraftFact label={t($ => $.draft.failure)} value={savedDraftConsumerState.failureCode ?? t($ => $.draft.none)} detail={savedDraftConsumerState.requestId} />
          {selectedDraft.derivation ? (
            <WorkflowDraftFact
              label={t($ => $.draft.derivationSource)}
              value={selectedDraft.derivation.version === 1
                ? selectedDraft.derivation.sourceDraftId
                : selectedDraft.derivation.templateId}
              detail={selectedDraft.derivation.version === 1
                ? t($ => $.draft.savedParent, { version: selectedDraft.derivation.sourceDraftVersion })
                : t($ => $.draft.templateParent, { version: selectedDraft.derivation.templateVersion, digest: selectedDraft.derivation.templateDigest })}
            />
          ) : null}
        </div>
        <div className="workflow-draft-action-row" aria-label={t($ => $.draft.secondaryActions)}>
          <button
            type="button"
            disabled={!canDeriveSavedWorkflowDraft(savedDraftConsumerState, draftEditDirty, interactionDisabled)}
            onClick={onDeriveSavedDraft}
          >{t($ => $.draft.derive)}</button>
          <button type="button" disabled={!draftEditDirty || interactionDisabled} onClick={onResetDraftEdits}>{t($ => $.draft.resetEdits)}</button>
        </div>
      </details>

      <details className="workflow-designer-disclosure">
        <summary>
          <span>{t($ => $.draft.graphFields)}</span>
          <strong>{t($ => $.draft.graphSize, { nodes: selectedDraft.nodes.length, edges: selectedDraft.edges.length })}</strong>
        </summary>
        <div className="workflow-draft-edit-grid" aria-label={t($ => $.draft.localEditing)}>
          <label className="workflow-draft-edit-field">
            <span>{t($ => $.draft.draftName)}</span>
            <input
              type="text"
              value={selectedDraft.label}
              maxLength={160}
              disabled={interactionDisabled}
              onChange={(event) => onUpdateDraftLabel(event.currentTarget.value)}
            />
          </label>
          <label className="workflow-draft-edit-field wide">
            <span>{t($ => $.draft.draftSummary)}</span>
            <textarea
              value={selectedDraft.summary}
              maxLength={4000}
              rows={3}
              disabled={interactionDisabled}
              onChange={(event) => onUpdateDraftSummary(event.currentTarget.value)}
            />
          </label>
        </div>
        <div className="workflow-draft-node-grid" aria-label={t($ => $.draft.draftNodes)}>
          {selectedDraft.nodes.map((node, nodeIndex) => (
            <WorkflowDraftNodeCard
              key={node.nodeId}
              node={node}
              nodeIndex={nodeIndex}
              nodeCount={selectedDraft.nodes.length}
              canDelete={canRemoveNode(node.nodeId)}
              editingDisabled={interactionDisabled}
              onUpdateLabel={onUpdateNodeLabel}
              onUpdateInputSummary={onUpdateNodeInputSummary}
              onUpdateOutputSummary={onUpdateNodeOutputSummary}
              onUpdateProviderRef={onUpdateNodeProviderRef}
              onUpdateToolRef={onUpdateNodeToolRef}
              onUpdateRagRef={onUpdateNodeRagRef}
              onUpdateInputFields={onUpdateNodeInputFields}
              onUpdateOutputFields={onUpdateNodeOutputFields}
              onUpdateOutputMapping={onUpdateNodeOutputMapping}
              onMoveNode={onMoveNode}
              onRemoveNode={onRemoveNode}
            />
          ))}
        </div>
        <div className="workflow-draft-edge-grid" aria-label={t($ => $.draft.draftEdges)}>
          {selectedDraft.edges.map((edge) => (
            <WorkflowDraftEdgeCard
              key={edge.edgeId}
              edge={edge}
              editingDisabled={interactionDisabled}
              onUpdateCondition={onUpdateEdgeCondition}
              onRemoveEdge={onRemoveEdge}
            />
          ))}
        </div>
      </details>

      <details className="workflow-designer-disclosure">
        <summary>
          <span>{t($ => $.draft.riskSection)}</span>
          <strong>{t($ => $.draft.readOnlyEvidence)}</strong>
        </summary>
        <div className="workflow-draft-readiness-grid" aria-label={t($ => $.draft.draftReadiness)}>
          {selectedDraft.readiness.map((readiness) => (
            <WorkflowDraftReadinessCard key={readiness.checkId} readiness={readiness} />
          ))}
        </div>
        <div className="workflow-draft-risk-grid" aria-label={t($ => $.draft.draftRisks)}>
          {selectedDraft.risks.map((risk) => (
            <WorkflowDraftRiskCard key={risk.riskId} risk={risk} />
          ))}
        </div>
        <div className="workflow-draft-blocked-grid" aria-label={t($ => $.draft.draftBlocked)}>
          {selectedDraft.blockedCapabilities.map((capability) => (
            <WorkflowDraftBlockedCapabilityCard key={capability.capabilityId} capability={capability} />
          ))}
        </div>
      </details>
    </section>
  );
}

function WorkflowDraftFact({ label, value, detail }: { label: string; value: string; detail: string }) {
  return (
    <article className="workflow-draft-card">
      <span>{label}</span>
      <strong>{value}</strong>
      <p>{detail}</p>
    </article>
  );
}

function WorkflowDraftTemplateButton({
  template,
  selected,
  disabled,
  onSelectDraft,
}: {
  template: WorkflowDraftDesignerTemplate;
  selected: boolean;
  disabled: boolean;
  onSelectDraft: (draftId: string) => void;
}) {
  const { t } = useTranslation("workflow");
  return (
    <button
      type="button"
      className={`workflow-draft-template-button${selected ? " selected" : ""}`}
      data-draft-status={template.status}
      aria-pressed={selected}
      disabled={disabled}
      onClick={() => onSelectDraft(template.draftId)}
    >
      <span>{template.workflowKind}</span>
      <strong>{template.label}</strong>
      <p>{template.summary}</p>
      <small>{t($ => $.draft.templateFacts, { status: workflowDraftStatusLabel(t, template.status), risk: workflowDraftStatusLabel(t, template.riskLevel), count: template.nodeCount })}</small>
    </button>
  );
}

function WorkflowDraftNodeCard({
  node,
  nodeIndex,
  nodeCount,
  canDelete,
  editingDisabled,
  onUpdateLabel,
  onUpdateInputSummary,
  onUpdateOutputSummary,
  onUpdateProviderRef,
  onUpdateToolRef,
  onUpdateRagRef,
  onUpdateInputFields,
  onUpdateOutputFields,
  onUpdateOutputMapping,
  onMoveNode,
  onRemoveNode,
}: {
  node: WorkflowDraftDesignerNode;
  nodeIndex: number;
  nodeCount: number;
  canDelete: boolean;
  editingDisabled: boolean;
  onUpdateLabel: (nodeId: string, label: string) => void;
  onUpdateInputSummary: (nodeId: string, inputSummary: string) => void;
  onUpdateOutputSummary: (nodeId: string, outputSummary: string) => void;
  onUpdateProviderRef: (nodeId: string, providerRef: string) => void;
  onUpdateToolRef: (nodeId: string, toolRef: string) => void;
  onUpdateRagRef: (nodeId: string, ragRef: string) => void;
  onUpdateInputFields: (nodeId: string, inputFieldsText: string) => void;
  onUpdateOutputFields: (nodeId: string, outputFieldsText: string) => void;
  onUpdateOutputMapping: (nodeId: string, outputMappingSummary: string) => void;
  onMoveNode: (nodeId: string, direction: WorkflowDraftNodeMoveDirection) => void;
  onRemoveNode: (nodeId: string) => void;
}) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-draft-node">
      <div className="workflow-draft-row-main">
        <div>
          <p className="eyebrow">{workflowDraftStatusLabel(t, node.lane)} / {workflowDraftStatusLabel(t, node.nodeType)}</p>
          <input
            className="workflow-draft-node-label-input"
            type="text"
            value={node.label}
            maxLength={160}
            disabled={editingDisabled}
            aria-label={t($ => $.draft.nodeLabel, { id: node.nodeId })}
            onChange={(event) => onUpdateLabel(node.nodeId, event.currentTarget.value)}
          />
        </div>
        <StatusBadge tone={node.readiness === "blocked" ? "bad" : node.readiness === "ready" ? "good" : "neutral"}>
          {workflowDraftStatusLabel(t, node.readiness)}
        </StatusBadge>
      </div>
      <div className="workflow-draft-node-actions" aria-label={t($ => $.draft.structureControls, { id: node.nodeId })}>
        <button type="button" disabled={editingDisabled || nodeIndex === 0} onClick={() => onMoveNode(node.nodeId, "up")}>{t($ => $.draft.up)}</button>
        <button type="button" disabled={editingDisabled || nodeIndex === nodeCount - 1} onClick={() => onMoveNode(node.nodeId, "down")}>{t($ => $.draft.down)}</button>
        <button type="button" disabled={editingDisabled || !canDelete} onClick={() => onRemoveNode(node.nodeId)}>{t($ => $.draft.remove)}</button>
      </div>
      <dl className="workflow-detail-node-meta">
        <div><dt>{t($ => $.draft.input)}</dt><dd>{node.inputSummary}</dd></div>
        <div><dt>{t($ => $.draft.output)}</dt><dd>{node.outputSummary}</dd></div>
        <div><dt>{t($ => $.draft.risk)}</dt><dd>{workflowDraftStatusLabel(t, node.riskLevel)}</dd></div>
        <div><dt>{t($ => $.draft.preview)}</dt><dd>{node.previewOnlyReason}</dd></div>
      </dl>
      <div className="workflow-draft-node-attribute-grid" aria-label={t($ => $.draft.nodeAttributes, { id: node.nodeId })}>
        <DraftNodeTextField label={t($ => $.draft.providerRef)} value={node.providerRef} disabled={editingDisabled} onChange={(value) => onUpdateProviderRef(node.nodeId, value)} />
        <DraftNodeTextField label={t($ => $.draft.toolRef)} value={node.toolRef} disabled={editingDisabled} onChange={(value) => onUpdateToolRef(node.nodeId, value)} />
        <DraftNodeTextField label={t($ => $.draft.ragRef)} value={node.ragRef} disabled={editingDisabled} onChange={(value) => onUpdateRagRef(node.nodeId, value)} />
        <DraftNodeTextArea label={t($ => $.draft.inputSummary)} value={node.inputSummary} disabled={editingDisabled} wide onChange={(value) => onUpdateInputSummary(node.nodeId, value)} />
        <DraftNodeTextArea label={t($ => $.draft.outputSummary)} value={node.outputSummary} disabled={editingDisabled} wide onChange={(value) => onUpdateOutputSummary(node.nodeId, value)} />
        <DraftNodeTextArea label={t($ => $.draft.inputFields)} value={node.inputContractFields.join(", ")} disabled={editingDisabled} maxLength={1000} onChange={(value) => onUpdateInputFields(node.nodeId, value)} />
        <DraftNodeTextArea label={t($ => $.draft.outputFields)} value={node.outputContractFields.join(", ")} disabled={editingDisabled} maxLength={1000} onChange={(value) => onUpdateOutputFields(node.nodeId, value)} />
        <DraftNodeTextArea label={t($ => $.draft.outputMapping)} value={node.outputMappingSummary} disabled={editingDisabled} wide onChange={(value) => onUpdateOutputMapping(node.nodeId, value)} />
      </div>
    </article>
  );
}

function DraftNodeTextField({
  label,
  value,
  disabled,
  onChange,
}: {
  label: string;
  value: string;
  disabled: boolean;
  onChange: (value: string) => void;
}) {
  return (
    <label className="workflow-draft-node-attribute-field">
      <span>{label}</span>
      <input type="text" value={value} maxLength={240} disabled={disabled} onChange={(event) => onChange(event.currentTarget.value)} />
    </label>
  );
}

function DraftNodeTextArea({
  label,
  value,
  disabled,
  maxLength = 4000,
  wide = false,
  onChange,
}: {
  label: string;
  value: string;
  disabled: boolean;
  maxLength?: number;
  wide?: boolean;
  onChange: (value: string) => void;
}) {
  return (
    <label className={`workflow-draft-node-attribute-field${wide ? " wide" : ""}`}>
      <span>{label}</span>
      <textarea value={value} maxLength={maxLength} rows={3} disabled={disabled} onChange={(event) => onChange(event.currentTarget.value)} />
    </label>
  );
}

function WorkflowDraftEdgeCard({
  edge,
  editingDisabled,
  onUpdateCondition,
  onRemoveEdge,
}: {
  edge: WorkflowDraftDesignerEdge;
  editingDisabled: boolean;
  onUpdateCondition: (edgeId: string, conditionSummary: string) => void;
  onRemoveEdge: (edgeId: string) => boolean;
}) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-draft-edge">
      <div className="workflow-draft-edge-heading">
        <div className="workflow-draft-edge-heading-main">
          <span>{edge.edgeKind}</span>
          <strong>{t($ => $.draft.edgeEndpoints, { from: edge.fromNodeId, to: edge.toNodeId })}</strong>
          <small>{edge.edgeId}</small>
        </div>
        <button type="button" disabled={editingDisabled} onClick={() => onRemoveEdge(edge.edgeId)}>{t($ => $.draft.remove)}</button>
      </div>
      <textarea
        className="workflow-draft-edge-condition-input"
        value={edge.conditionSummary}
        maxLength={4000}
        rows={3}
        disabled={editingDisabled}
        aria-label={t($ => $.draft.edgeCondition, { id: edge.edgeId })}
        onChange={(event) => onUpdateCondition(edge.edgeId, event.currentTarget.value)}
      />
    </article>
  );
}

function WorkflowDraftReadinessCard({ readiness }: { readiness: WorkflowDraftDesignerReadiness }) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-draft-readiness">
      <div className="workflow-draft-row-main">
        <div><p className="eyebrow">{readiness.checkId}</p><h5>{readiness.label}</h5></div>
        <StatusBadge tone={readiness.status === "blocked" ? "bad" : readiness.status === "ready" ? "good" : "neutral"}>{workflowDraftStatusLabel(t, readiness.status)}</StatusBadge>
      </div>
      <p>{readiness.summary}</p>
    </article>
  );
}

function WorkflowDraftRiskCard({ risk }: { risk: WorkflowDraftDesignerRisk }) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-draft-risk">
      <div className="workflow-draft-row-main">
        <div><p className="eyebrow">{risk.riskId}</p><h5>{risk.label}</h5></div>
        <StatusBadge tone={risk.riskLevel === "high" ? "bad" : risk.riskLevel === "low" ? "good" : "neutral"}>{workflowDraftStatusLabel(t, risk.riskLevel)}</StatusBadge>
      </div>
      <p>{risk.summary}</p>
      <small>{risk.requiresConfirmation ? t($ => $.draft.humanReview) : t($ => $.draft.advisoryOnly)}</small>
    </article>
  );
}

function WorkflowDraftBlockedCapabilityCard({ capability }: { capability: WorkflowDraftDesignerBlockedCapability }) {
  const { t } = useTranslation("workflow");
  return (
    <article className="workflow-draft-blocked-capability">
      <div className="workflow-draft-row-main">
        <div><p className="eyebrow">{capability.capabilityId}</p><h5>{capability.label}</h5></div>
        <StatusBadge tone="bad">{workflowDraftStatusLabel(t, capability.status)}</StatusBadge>
      </div>
      <dl className="workflow-run-guard-meta">
        <div><dt>{t($ => $.draft.missingPrerequisite)}</dt><dd>{capability.missingPrerequisite}</dd></div>
        <div><dt>{t($ => $.draft.audit)}</dt><dd>{capability.auditRef}</dd></div>
      </dl>
      <p>{capability.summary}</p>
    </article>
  );
}

function workflowSavedDraftConsumerTone(status: WorkflowSavedDraftConsumerState["status"]): "good" | "bad" | "neutral" {
  if (status === "saved_dev_record" || status === "validation_ready") return "good";
  if (["version_conflict", "save_failed", "read_failed", "validation_failed"].includes(status)) return "bad";
  return "neutral";
}

function StatusBadge({ children, tone }: { children: string; tone: "good" | "bad" | "neutral" }) {
  return <span className={`status-badge ${tone}`}>{children}</span>;
}
