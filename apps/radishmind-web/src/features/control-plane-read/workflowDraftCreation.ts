import type { WorkflowDraftDesignerDraft, WorkflowDraftDesignerViewModel } from "./workflowDraftDesigner.ts";
import { initialWorkflowSavedDraftConsumerState, type WorkflowSavedDraftConsumerConfig, type WorkflowSavedDraftConsumerState } from "./savedWorkflowDraftConsumer.ts";
import { cloneWorkflowDraftForEditing } from "./workflowSavedDraftDerivation.ts";

export function workspaceDraftCreatedConsumerState(
  config: WorkflowSavedDraftConsumerConfig,
  draft: WorkflowDraftDesignerDraft,
  sourceLifecycleVersion = 0,
): WorkflowSavedDraftConsumerState {
  const initialState = initialWorkflowSavedDraftConsumerState(config);
  return {
    ...initialState,
    status: "unsaved_local",
    sourceLabel: "workspace draft",
    summary:
      config.mode === "dev_saved_draft_http"
        ? `Workspace draft ${draft.draftId} is ready for validation or save through the dev-only saved draft route.`
        : `Workspace draft ${draft.draftId} is local only until the dev-only saved draft route is enabled.`,
    failureCode: null,
    currentDraftVersion: 0,
    currentLifecycleVersion: sourceLifecycleVersion,
    currentLifecycleState: "active",
    conflictDraftVersion: null,
    auditRef: draft.routeMetadata.auditRef,
    requestId: draft.routeMetadata.requestId,
  };
}

export function workflowTemplateDerivedConsumerState(
  config: WorkflowSavedDraftConsumerConfig,
  draft: WorkflowDraftDesignerDraft,
  authority: {
    draftId: string;
    draftVersion: number;
    lifecycleVersion: number;
    lifecycleState: "active";
  },
): WorkflowSavedDraftConsumerState {
  return {
    ...initialWorkflowSavedDraftConsumerState(config),
    status: "saved_dev_record",
    sourceLabel: "template-derived saved draft",
    summary: `Template-derived Saved Draft ${authority.draftId} v${authority.draftVersion} is open from exact server authority.`,
    failureCode: null,
    currentDraftVersion: authority.draftVersion,
    currentLifecycleVersion: authority.lifecycleVersion,
    currentLifecycleState: authority.lifecycleState,
    conflictDraftVersion: null,
    auditRef: draft.routeMetadata.auditRef,
    requestId: draft.routeMetadata.requestId,
  };
}

export function buildWorkspaceCreatedDraft(
  workflowDefinitionId: string,
  designer: WorkflowDraftDesignerViewModel,
  existingDrafts: WorkflowDraftDesignerDraft[],
): WorkflowDraftDesignerDraft | null {
  const template = designer.templates.find(
    (draftTemplate) => draftTemplate.workflowDefinitionId === workflowDefinitionId,
  );
  const baseDraft = template
    ? designer.drafts.find((draft) => draft.draftId === template.draftId)
    : designer.drafts.find((draft) => draft.workflowDefinitionId === workflowDefinitionId);
  if (!baseDraft) {
    return null;
  }
  const nextDraftNumber =
    existingDrafts.filter((draft) => draft.workflowDefinitionId === workflowDefinitionId).length + 1;
  const draftNumberLabel = String(nextDraftNumber).padStart(2, "0");
  const createdDraftId = `draft_${workflowDefinitionId}_workspace_${draftNumberLabel}`;
  return {
    ...cloneWorkflowDraftForEditing(baseDraft),
    draftId: createdDraftId,
    templateRef: baseDraft.draftId,
    label: `${baseDraft.label} workspace ${draftNumberLabel}`,
    summary: `Workspace-created draft derived from ${workflowDefinitionId}; edit locally, validate, and save through the dev-only saved draft route before review.`,
    localOnlyInteraction: "local_edit",
    routeMetadata: {
      ...baseDraft.routeMetadata,
      requestId: `${baseDraft.routeMetadata.requestId}_workspace_${draftNumberLabel}`,
      auditRef: `${baseDraft.routeMetadata.auditRef}_workspace_${draftNumberLabel}`,
    },
  };
}

