import type {
  WorkflowDraftDesignerDraft,
  WorkflowDraftDesignerLayout,
} from "./workflowDraftDesigner.ts";

const MAX_DRAFT_LABEL_LENGTH = 160;

export type WorkflowSavedDraftDerivationState = {
  status: string;
  currentDraftVersion: number;
};

export function canDeriveSavedWorkflowDraft(
  state: WorkflowSavedDraftDerivationState,
  draftEditDirty: boolean,
  operationPending: boolean,
): boolean {
  return state.status === "saved_dev_record" &&
    state.currentDraftVersion >= 1 &&
    !draftEditDirty &&
    !operationPending;
}

export function buildDerivedWorkflowDraft(
  source: WorkflowDraftDesignerDraft,
  sourceDraftVersion: number,
  existingDraftIds: Iterable<string>,
): WorkflowDraftDesignerDraft {
  if (!Number.isInteger(sourceDraftVersion) || sourceDraftVersion < 1) {
    throw new Error("saved workflow draft derivation requires an exact source version");
  }

  const occupiedDraftIds = new Set(existingDraftIds);
  const idPrefix = `draft_derived_${stableDraftReferenceHash(source.draftId)}`;
  let sequence = 1;
  let draftId = `${idPrefix}_${String(sequence).padStart(2, "0")}`;
  while (occupiedDraftIds.has(draftId)) {
    sequence += 1;
    draftId = `${idPrefix}_${String(sequence).padStart(2, "0")}`;
  }
  const sequenceLabel = String(sequence).padStart(2, "0");
  const draft = cloneWorkflowDraftForEditing(source);

  return {
    ...draft,
    draftId,
    templateRef: source.draftId,
    label: derivedDraftLabel(source.label, sequenceLabel),
    designerLayout: {
      ...draft.designerLayout,
      persistence: "ui_only",
    },
    readiness: [
      {
        checkId: "saved_draft_derivation_source",
        label: "Saved draft derivation source",
        status: "ready",
        summary: `Derived locally from saved draft ${source.draftId} version ${sourceDraftVersion}.`,
      },
      ...draft.readiness.filter((item) => item.checkId !== "saved_draft_derivation_source"),
    ],
    routeMetadata: {
      ...draft.routeMetadata,
      requestId: `local_derive_${draftId}`,
      auditRef: `audit_local_derive_${draftId}`,
    },
    localOnlyInteraction: "local_edit",
    derivation: {
      version: 1,
      sourceKind: "saved_workflow_draft",
      sourceDraftId: source.draftId,
      sourceDraftVersion,
    },
  };
}

export function cloneWorkflowDraftForEditing(
  draft: WorkflowDraftDesignerDraft,
): WorkflowDraftDesignerDraft {
  return {
    ...draft,
    nodes: draft.nodes.map((node) => ({
      ...node,
      inputContractFields: [...node.inputContractFields],
      outputContractFields: [...node.outputContractFields],
    })),
    edges: draft.edges.map((edge) => ({ ...edge })),
    designerLayout: cloneWorkflowDraftDesignerLayout(draft.designerLayout),
    readiness: draft.readiness.map((readiness) => ({ ...readiness })),
    risks: draft.risks.map((risk) => ({ ...risk })),
    blockedCapabilities: draft.blockedCapabilities.map((capability) => ({ ...capability })),
    routeMetadata: { ...draft.routeMetadata },
    ...(draft.derivation ? { derivation: { ...draft.derivation } } : {}),
  };
}

function cloneWorkflowDraftDesignerLayout(
  layout: WorkflowDraftDesignerLayout,
): WorkflowDraftDesignerLayout {
  return {
    source: "workflow_node_designer",
    persistence: layout.persistence,
    nodePositions: layout.nodePositions.map((position) => ({ ...position })),
  };
}

function derivedDraftLabel(sourceLabel: string, sequenceLabel: string): string {
  const suffix = ` 派生 ${sequenceLabel}`;
  const maximumSourceLength = Math.max(0, MAX_DRAFT_LABEL_LENGTH - suffix.length);
  return `${sourceLabel.slice(0, maximumSourceLength).trimEnd()}${suffix}`;
}

function stableDraftReferenceHash(value: string): string {
  let hash = 0x811c9dc5;
  for (let index = 0; index < value.length; index += 1) {
    hash ^= value.charCodeAt(index);
    hash = Math.imul(hash, 0x01000193);
  }
  return (hash >>> 0).toString(16).padStart(8, "0");
}
