import { lazy, Suspense } from "react";

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
  nodeTypeOptions: readonly WorkflowDraftNodeTypeOption[];
  canRemoveNode: (nodeId: string) => boolean;
  onSelectDraft: (draftId: string) => void;
  onUpdateDraftLabel: (label: string) => void;
  onUpdateDraftSummary: (summary: string) => void;
  onUpdateNodeLabel: (nodeId: string, label: string) => void;
  onUpdateNodeInputSummary: (nodeId: string, inputSummary: string) => void;
  onUpdateNodeOutputSummary: (nodeId: string, outputSummary: string) => void;
  onUpdateNodeProviderRef: (nodeId: string, providerRef: string) => void;
  onUpdateNodeToolRef: (nodeId: string, toolRef: string) => void;
  onUpdateNodeRagRef: (nodeId: string, ragRef: string) => void;
  onUpdateNodeInputFields: (nodeId: string, inputFieldsText: string) => void;
  onUpdateNodeOutputFields: (nodeId: string, outputFieldsText: string) => void;
  onUpdateNodeOutputMapping: (nodeId: string, outputMappingSummary: string) => void;
  onUpdateNodeDesignerPosition: (nodeId: string, x: number, y: number) => void;
  onUpdateEdgeCondition: (edgeId: string, conditionSummary: string) => void;
  onAddEdge: (fromNodeId: string, toNodeId: string) => boolean;
  onRemoveEdge: (edgeId: string) => boolean;
  onAddNode: (nodeType: WorkflowDraftDesignerNode["nodeType"]) => void;
  onMoveNode: (nodeId: string, direction: WorkflowDraftNodeMoveDirection) => void;
  onRemoveNode: (nodeId: string) => void;
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
  nodeTypeOptions,
  canRemoveNode,
  onSelectDraft,
  onUpdateDraftLabel,
  onUpdateDraftSummary,
  onUpdateNodeLabel,
  onUpdateNodeInputSummary,
  onUpdateNodeOutputSummary,
  onUpdateNodeProviderRef,
  onUpdateNodeToolRef,
  onUpdateNodeRagRef,
  onUpdateNodeInputFields,
  onUpdateNodeOutputFields,
  onUpdateNodeOutputMapping,
  onUpdateNodeDesignerPosition,
  onUpdateEdgeCondition,
  onAddEdge,
  onRemoveEdge,
  onAddNode,
  onMoveNode,
  onRemoveNode,
  onResetDraftEdits,
  onContinueLocalDraftAfterConflict,
  onOpenConflictSavedDraft,
  onDeriveSavedDraft,
  onValidateDraft,
  onSaveDraft,
  onReadDraft,
}: WorkflowDraftDesignerPanelProps) {
  const canCallDevConsumer = savedDraftConsumerState.mode === "dev_saved_draft_http";
  const operationPending = ["saving", "validating", "reading"].includes(savedDraftConsumerState.status);
  const conflictRequiresResolution = workflowSavedDraftConflictRequiresResolution(savedDraftConsumerState);
  const lifecycleReadOnly =
    savedDraftConsumerState.currentDraftVersion > 0 &&
    savedDraftConsumerState.currentLifecycleState !== "active";
  const lifecycleReadOnlyLabel = savedDraftConsumerState.currentLifecycleState === "unknown"
    ? "reopen required"
    : `${savedDraftConsumerState.currentLifecycleState} read-only review`;
  const requestInteractionDisabled = operationPending || conflictRequiresResolution || executorOperationPending;
  const interactionDisabled = requestInteractionDisabled || lifecycleReadOnly;
  const editStateLabel = draftEditDirty ? "unsaved local" : selectedDraft.localOnlyInteraction;
  const conflictOpenUnavailableMessage =
    savedDraftConflictReviewSummary?.openUnavailableReason ??
    "Saved version metadata is refreshing from the dev-only saved draft list before open is enabled.";

  return (
    <section
      className="workflow-draft-designer workflow-designer-workbench"
      id="workflow-draft-designer"
      aria-label="Workflow draft designer development workbench"
    >
      <header className="workflow-designer-header">
        <div className="workflow-designer-breadcrumb-row">
          <span>Workflows</span>
          <span aria-hidden="true">/</span>
          <strong>{selectedDraft.workflowDefinitionId}</strong>
          <code>{selectedDraft.routeMetadata.routePath}</code>
        </div>
        <div className="workflow-designer-title-row">
          <div>
            <h4>Workflow Designer</h4>
            <p>{selectedDraft.label}</p>
          </div>
          <StatusBadge tone={designer.canRenderDraftDesigner ? "good" : "bad"}>
            {lifecycleReadOnly
              ? lifecycleReadOnlyLabel
              : designer.canRenderDraftDesigner
                ? "development designer"
                : "blocked"}
          </StatusBadge>
        </div>
        <dl className="workflow-designer-context" aria-label="Active workflow draft context">
          <div>
            <dt>Application</dt>
            <dd>{selectedDraft.applicationRef}</dd>
          </div>
          <div>
            <dt>Draft</dt>
            <dd>{selectedDraft.draftId}</dd>
          </div>
          <div>
            <dt>Version</dt>
            <dd>content {savedDraftConsumerState.currentDraftVersion} / lifecycle {savedDraftConsumerState.currentLifecycleVersion}</dd>
          </div>
          <div>
            <dt>Lifecycle</dt>
            <dd>{savedDraftConsumerState.currentLifecycleState}</dd>
          </div>
          <div className={draftEditDirty ? "attention" : "neutral"}>
            <dt>Edit state</dt>
            <dd>{editStateLabel}</dd>
          </div>
        </dl>
        <div className="workflow-designer-actions" aria-label="Workflow draft actions">
          <button
            type="button"
            disabled={!canCallDevConsumer || interactionDisabled}
            onClick={onValidateDraft}
          >
            Validate
          </button>
          <button
            type="button"
            className="primary"
            disabled={!canCallDevConsumer || interactionDisabled}
            onClick={onSaveDraft}
          >
            Save draft
          </button>
          <button
            type="button"
            disabled={!canCallDevConsumer || requestInteractionDisabled}
            onClick={onReadDraft}
          >
            Read saved
          </button>
        </div>
        {lifecycleReadOnly ? (
          <p className="workflow-draft-revision-stopline" role="status">
            {savedDraftConsumerState.currentLifecycleState === "unknown"
              ? "草案已解除归档，但当前浏览器仍保留解除归档前的只读快照；请从活动草案库重新打开，以读取最新 lifecycle 和内容版本。"
              : `当前草案 lifecycle 为 ${savedDraftConsumerState.currentLifecycleState}：内容、修订历史和比较保持可读；本地编辑、保存、派生、恢复、晋级和直接执行全部禁用。`}
          </p>
        ) : null}
      </header>

      <div className="workflow-designer-primary-layout">
        <aside className="workflow-designer-rail" aria-label="Draft references and node types">
          <div className="workflow-designer-rail-heading">
            <div>
              <span>Draft references</span>
              <strong>{designer.templates.length} available</strong>
            </div>
            <a href="#workflow-user-workspace-home">Open library</a>
          </div>
          <div className="workflow-draft-template-grid" aria-label="Workflow draft templates">
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
                <span>Add node</span>
                <strong>{selectedDraft.nodes.length} nodes</strong>
              </div>
              <small>Expand</small>
            </summary>
            <div className="workflow-draft-add-node-grid" aria-label="Add workflow draft node">
              {nodeTypeOptions.map((option) => (
                <button
                  key={option.nodeType}
                  type="button"
                  className="workflow-draft-node-type-button"
                  disabled={interactionDisabled}
                  onClick={() => onAddNode(option.nodeType)}
                >
                  <span>{option.lane}</span>
                  <strong>{option.label}</strong>
                  <small>{option.summary}</small>
                </button>
              ))}
            </div>
          </details>
        </aside>

        <div className="workflow-designer-canvas-column">
          <Suspense
            fallback={(
              <section className="workflow-node-designer-shell" aria-label="Loading node designer">
                <p>Loading node designer…</p>
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
            <span>Review surfaces</span>
            <strong>
              {validationInspector.validationStatus} · {validationInspector.structuralChecks.length + validationInspector.contractChecks.length} validation checks
            </strong>
            <small>Plan, readiness, and handoff remain derived or read only</small>
          </summary>
          <nav className="workflow-designer-review-links" aria-label="Workflow draft review surfaces">
            <a href="#workflow-draft-validation-inspector">
              <span>Validation</span>
              <strong>{validationInspector.validationStatus}</strong>
              <small>{validationInspector.structuralChecks.length + validationInspector.contractChecks.length} checks</small>
            </a>
            <a href="#workflow-execution-plan-preview">
              <span>Preview plan</span>
              <strong>Derived only</strong>
              <small>No execution owner</small>
            </a>
            <a href="#workflow-runtime-readiness-inspector">
              <span>Readiness</span>
              <strong>Read only</strong>
              <small>{selectedDraft.readiness.length} checks</small>
            </a>
            <a href="#workflow-review-handoff">
              <span>Review handoff</span>
              <strong>Browser only</strong>
              <small>No save, export, or send</small>
            </a>
          </nav>
        </details>
      </div>

      {savedDraftConflictReviewSummary ? (
        <details className="workflow-designer-disclosure workflow-draft-conflict-review" open>
          <summary>
            <span>Version conflict review</span>
            <StatusBadge
              tone={savedDraftConflictReviewSummary.status === "local_draft_continued" ? "neutral" : "bad"}
            >
              {savedDraftConflictReviewSummary.status}
            </StatusBadge>
          </summary>
          <article className="workflow-draft-card workflow-draft-conflict-review-card">
            <dl className="workflow-run-guard-meta">
              <div><dt>Local draft</dt><dd>{savedDraftConflictReviewSummary.draftId}</dd></div>
              <div><dt>Saved version</dt><dd>{savedDraftConflictReviewSummary.savedDraftVersion}</dd></div>
              <div><dt>Updated</dt><dd>{savedDraftConflictReviewSummary.savedUpdatedAt}</dd></div>
              <div><dt>Actor</dt><dd>{savedDraftConflictReviewSummary.savedUpdatedByActorRef}</dd></div>
              <div><dt>Validation</dt><dd>{savedDraftConflictReviewSummary.savedValidationState}</dd></div>
              <div><dt>Blocked</dt><dd>{savedDraftConflictReviewSummary.savedBlockedCapabilityCount ?? "not loaded"}</dd></div>
              <div><dt>Metadata</dt><dd>{savedDraftConflictReviewSummary.savedMetadataState}</dd></div>
              <div><dt>Open</dt><dd>{savedDraftConflictReviewSummary.openActionState}</dd></div>
            </dl>
            <p>{savedDraftConflictReviewSummary.summary}</p>
            <p>{savedDraftConflictReviewSummary.localDraftPreservationSummary}</p>
            <div className="workflow-workspace-review-token-list" aria-label="Saved draft conflict review locks">
              <code>auto_overwrite_locked</code>
              <code>auto_merge_locked</code>
              <code>{savedDraftConflictReviewSummary.openActionState}</code>
            </div>
            <div className="workflow-draft-conflict-action-row" aria-label="Saved draft conflict review actions">
              <button
                type="button"
                disabled={operationPending || savedDraftConflictReviewSummary.status === "local_draft_continued"}
                onClick={onContinueLocalDraftAfterConflict}
              >
                Continue local draft
              </button>
              <button
                type="button"
                disabled={operationPending || !savedDraftConflictOpenSummary || !savedDraftConflictReviewSummary.canOpenSavedDraft}
                onClick={onOpenConflictSavedDraft}
              >
                Open saved draft
              </button>
            </div>
            <p>
              {savedDraftConflictReviewSummary.canOpenSavedDraft
                ? "Open saved draft is available from sanitized saved draft metadata; it replaces the active draft only after explicit selection."
                : conflictOpenUnavailableMessage}
            </p>
            <p>{savedDraftConflictReviewSummary.nextReviewerStep}</p>
            <p>{savedDraftConflictReviewSummary.reviewerQuestion}</p>
          </article>
        </details>
      ) : null}

      <details className="workflow-designer-disclosure">
        <summary>
          <span>Draft identity, saved state, and persistence boundary</span>
          <StatusBadge tone={workflowSavedDraftConsumerTone(savedDraftConsumerState.status)}>
            {savedDraftConsumerState.status}
          </StatusBadge>
        </summary>
        <div className="workflow-draft-summary-grid" aria-label="Selected workflow draft summary">
          <WorkflowDraftFact label="Draft" value={selectedDraft.draftId} detail={selectedDraft.summary} />
          <WorkflowDraftFact label="Route" value={selectedDraft.routeMetadata.draftRouteId} detail={selectedDraft.routeMetadata.routePath} />
          <WorkflowDraftFact label="Source" value={selectedDraft.routeMetadata.sourceRouteId} detail={selectedDraft.workflowDefinitionId} />
          <WorkflowDraftFact label="Request" value={selectedDraft.routeMetadata.requestId} detail={selectedDraft.routeMetadata.auditRef} />
          <WorkflowDraftFact label="Saved state" value={savedDraftConsumerState.sourceLabel} detail={savedDraftConsumerState.summary} />
          <WorkflowDraftFact label="Failure" value={savedDraftConsumerState.failureCode ?? "none"} detail={savedDraftConsumerState.requestId} />
          {selectedDraft.derivation ? (
            <WorkflowDraftFact
              label="派生来源"
              value={selectedDraft.derivation.sourceDraftId}
              detail={`saved version ${selectedDraft.derivation.sourceDraftVersion} · direct parent only`}
            />
          ) : null}
        </div>
        <div className="workflow-draft-action-row" aria-label="Saved draft dev consumer secondary actions">
          <button
            type="button"
            disabled={!canDeriveSavedWorkflowDraft(savedDraftConsumerState, draftEditDirty, interactionDisabled)}
            onClick={onDeriveSavedDraft}
          >
            派生新草案
          </button>
          <button type="button" disabled={!draftEditDirty || interactionDisabled} onClick={onResetDraftEdits}>
            Reset local edits
          </button>
        </div>
      </details>

      <details className="workflow-designer-disclosure">
        <summary>
          <span>Fine-grained draft and graph fields</span>
          <strong>{selectedDraft.nodes.length} nodes / {selectedDraft.edges.length} edges</strong>
        </summary>
        <div className="workflow-draft-edit-grid" aria-label="Workflow draft local editing">
          <label className="workflow-draft-edit-field">
            <span>Draft name</span>
            <input
              type="text"
              value={selectedDraft.label}
              maxLength={160}
              disabled={interactionDisabled}
              onChange={(event) => onUpdateDraftLabel(event.currentTarget.value)}
            />
          </label>
          <label className="workflow-draft-edit-field wide">
            <span>Draft summary</span>
            <textarea
              value={selectedDraft.summary}
              maxLength={4000}
              rows={3}
              disabled={interactionDisabled}
              onChange={(event) => onUpdateDraftSummary(event.currentTarget.value)}
            />
          </label>
        </div>
        <div className="workflow-draft-node-grid" aria-label="Workflow draft nodes">
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
        <div className="workflow-draft-edge-grid" aria-label="Workflow draft edges">
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
          <span>Readiness, risks, and blocked capabilities</span>
          <strong>Read-only evidence</strong>
        </summary>
        <div className="workflow-draft-readiness-grid" aria-label="Workflow draft readiness">
          {selectedDraft.readiness.map((readiness) => (
            <WorkflowDraftReadinessCard key={readiness.checkId} readiness={readiness} />
          ))}
        </div>
        <div className="workflow-draft-risk-grid" aria-label="Workflow draft risk summary">
          {selectedDraft.risks.map((risk) => (
            <WorkflowDraftRiskCard key={risk.riskId} risk={risk} />
          ))}
        </div>
        <div className="workflow-draft-blocked-grid" aria-label="Workflow draft blocked capabilities">
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
      <small>{template.status} / risk {template.riskLevel} / nodes {template.nodeCount}</small>
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
  return (
    <article className="workflow-draft-node">
      <div className="workflow-draft-row-main">
        <div>
          <p className="eyebrow">{node.lane} / {node.nodeType}</p>
          <input
            className="workflow-draft-node-label-input"
            type="text"
            value={node.label}
            maxLength={160}
            disabled={editingDisabled}
            aria-label={`Node label ${node.nodeId}`}
            onChange={(event) => onUpdateLabel(node.nodeId, event.currentTarget.value)}
          />
        </div>
        <StatusBadge tone={node.readiness === "blocked" ? "bad" : node.readiness === "ready" ? "good" : "neutral"}>
          {node.readiness}
        </StatusBadge>
      </div>
      <div className="workflow-draft-node-actions" aria-label={`Structure controls ${node.nodeId}`}>
        <button type="button" disabled={editingDisabled || nodeIndex === 0} onClick={() => onMoveNode(node.nodeId, "up")}>Up</button>
        <button type="button" disabled={editingDisabled || nodeIndex === nodeCount - 1} onClick={() => onMoveNode(node.nodeId, "down")}>Down</button>
        <button type="button" disabled={editingDisabled || !canDelete} onClick={() => onRemoveNode(node.nodeId)}>Remove</button>
      </div>
      <dl className="workflow-detail-node-meta">
        <div><dt>Input</dt><dd>{node.inputSummary}</dd></div>
        <div><dt>Output</dt><dd>{node.outputSummary}</dd></div>
        <div><dt>Risk</dt><dd>{node.riskLevel}</dd></div>
        <div><dt>Preview</dt><dd>{node.previewOnlyReason}</dd></div>
      </dl>
      <div className="workflow-draft-node-attribute-grid" aria-label={`Node attributes ${node.nodeId}`}>
        <DraftNodeTextField label="Provider ref" value={node.providerRef} disabled={editingDisabled} onChange={(value) => onUpdateProviderRef(node.nodeId, value)} />
        <DraftNodeTextField label="Tool ref" value={node.toolRef} disabled={editingDisabled} onChange={(value) => onUpdateToolRef(node.nodeId, value)} />
        <DraftNodeTextField label="RAG ref" value={node.ragRef} disabled={editingDisabled} onChange={(value) => onUpdateRagRef(node.nodeId, value)} />
        <DraftNodeTextArea label="Input summary" value={node.inputSummary} disabled={editingDisabled} wide onChange={(value) => onUpdateInputSummary(node.nodeId, value)} />
        <DraftNodeTextArea label="Output summary" value={node.outputSummary} disabled={editingDisabled} wide onChange={(value) => onUpdateOutputSummary(node.nodeId, value)} />
        <DraftNodeTextArea label="Input fields" value={node.inputContractFields.join(", ")} disabled={editingDisabled} maxLength={1000} onChange={(value) => onUpdateInputFields(node.nodeId, value)} />
        <DraftNodeTextArea label="Output fields" value={node.outputContractFields.join(", ")} disabled={editingDisabled} maxLength={1000} onChange={(value) => onUpdateOutputFields(node.nodeId, value)} />
        <DraftNodeTextArea label="Output mapping" value={node.outputMappingSummary} disabled={editingDisabled} wide onChange={(value) => onUpdateOutputMapping(node.nodeId, value)} />
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
  return (
    <article className="workflow-draft-edge">
      <div className="workflow-draft-edge-heading">
        <div className="workflow-draft-edge-heading-main">
          <span>{edge.edgeKind}</span>
          <strong>{edge.fromNodeId} to {edge.toNodeId}</strong>
          <small>{edge.edgeId}</small>
        </div>
        <button type="button" disabled={editingDisabled} onClick={() => onRemoveEdge(edge.edgeId)}>Remove</button>
      </div>
      <textarea
        className="workflow-draft-edge-condition-input"
        value={edge.conditionSummary}
        maxLength={4000}
        rows={3}
        disabled={editingDisabled}
        aria-label={`Edge condition ${edge.edgeId}`}
        onChange={(event) => onUpdateCondition(edge.edgeId, event.currentTarget.value)}
      />
    </article>
  );
}

function WorkflowDraftReadinessCard({ readiness }: { readiness: WorkflowDraftDesignerReadiness }) {
  return (
    <article className="workflow-draft-readiness">
      <div className="workflow-draft-row-main">
        <div><p className="eyebrow">{readiness.checkId}</p><h5>{readiness.label}</h5></div>
        <StatusBadge tone={readiness.status === "blocked" ? "bad" : readiness.status === "ready" ? "good" : "neutral"}>{readiness.status}</StatusBadge>
      </div>
      <p>{readiness.summary}</p>
    </article>
  );
}

function WorkflowDraftRiskCard({ risk }: { risk: WorkflowDraftDesignerRisk }) {
  return (
    <article className="workflow-draft-risk">
      <div className="workflow-draft-row-main">
        <div><p className="eyebrow">{risk.riskId}</p><h5>{risk.label}</h5></div>
        <StatusBadge tone={risk.riskLevel === "high" ? "bad" : risk.riskLevel === "low" ? "good" : "neutral"}>{risk.riskLevel}</StatusBadge>
      </div>
      <p>{risk.summary}</p>
      <small>{risk.requiresConfirmation ? "future human review required" : "advisory only"}</small>
    </article>
  );
}

function WorkflowDraftBlockedCapabilityCard({ capability }: { capability: WorkflowDraftDesignerBlockedCapability }) {
  return (
    <article className="workflow-draft-blocked-capability">
      <div className="workflow-draft-row-main">
        <div><p className="eyebrow">{capability.capabilityId}</p><h5>{capability.label}</h5></div>
        <StatusBadge tone="bad">{capability.status}</StatusBadge>
      </div>
      <dl className="workflow-run-guard-meta">
        <div><dt>Missing prerequisite</dt><dd>{capability.missingPrerequisite}</dd></div>
        <div><dt>Audit</dt><dd>{capability.auditRef}</dd></div>
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
