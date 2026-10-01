import { useEffect, useMemo, useRef, useState } from "react";
import {
  continueLocalWorkflowDraftAfterVersionConflict, initialWorkflowSavedDraftConsumerState,
  nextWorkflowSavedDraftExpectedVersion, openWorkflowDraftDevRecord, readWorkflowDraftDevRecord,
  saveWorkflowDraftDevRecord, validateWorkflowDraftDevRecord, workflowSavedDraftConflictRequiresResolution,
  type WorkflowSavedDraftConsumerConfig, type WorkflowSavedDraftConsumerState,
  type WorkflowSavedDraftSummary, type WorkflowSavedDraftLifecycleState,
} from "./savedWorkflowDraftConsumer.ts";
import { createWorkflowDraftEditorRequests } from "./workflowDraftEditorRequests.ts";
import type { WorkflowSavedDraftRevisionRestoreResult } from "./workflowSavedDraftRevisionConsumer.ts";
import type { WorkflowDraftDesignerDraft } from "./workflowDraftDesigner.ts";
import { buildDerivedWorkflowDraft, canDeriveSavedWorkflowDraft, cloneWorkflowDraftForEditing } from "./workflowSavedDraftDerivation.ts";
import { buildWorkflowExecutorV0Draft } from "./workflowExecutorConsumer.ts";
import { buildWorkflowWorkspaceContextViewModel, type WorkflowWorkspaceContextSource, type WorkflowWorkspaceSelectionPatch } from "./workflowWorkspaceContext";
import { buildWorkspaceCreatedDraft, workspaceDraftCreatedConsumerState, workflowTemplateDerivedConsumerState } from "./workflowDraftCreation.ts";
import { useWorkflowSavedDraftLibrary } from "./useWorkflowSavedDraftLibrary.ts";

type WorkflowDraftWorkspaceSource = Pick<WorkflowWorkspaceContextSource,
  "workspaceApplications" | "workspaceApiKeys" | "workspaceUsageQuota" | "workspaceWorkflowDefinitions" | "workspaceRunHistory" | "selection">;

export function useWorkflowDraftWorkspace({
  source, config: activeSavedDraftConsumerConfig, applicationId: workflowScopedApplicationId,
  generationKey, workflowExecutorOperationPending, workflowRAGOperationPending,
  onSelect: applyWorkflowSelectionPatch, onOpenDesigner,
}: {
  source: WorkflowDraftWorkspaceSource;
  config: WorkflowSavedDraftConsumerConfig;
  applicationId: string;
  generationKey: string;
  workflowExecutorOperationPending: boolean;
  workflowRAGOperationPending: boolean;
  onSelect: (selection: WorkflowWorkspaceSelectionPatch) => void;
  onOpenDesigner: () => void;
}) {
  const { workspaceApplications, workspaceApiKeys, workspaceUsageQuota, workspaceWorkflowDefinitions, workspaceRunHistory } = source;
  const { applicationRef: selectedApplicationRef, workflowDefinitionId: selectedWorkflowDefinitionId,
    runId: selectedRunId, draftId: selectedWorkflowDraftId, scenarioId: selectedWorkflowScenarioId } = source.selection;
  const [savedDraftConsumerState, setSavedDraftConsumerState] = useState<WorkflowSavedDraftConsumerState>(() =>
    initialWorkflowSavedDraftConsumerState(activeSavedDraftConsumerConfig),
  );
  const pendingSavedDraftConsumerStateRef = useRef<{
    draftId: string;
    state: WorkflowSavedDraftConsumerState;
  } | null>(null);
  const [workspaceCreatedDrafts, setWorkspaceCreatedDrafts] = useState<WorkflowDraftDesignerDraft[]>([]);
  const [editableWorkflowDraft, setEditableWorkflowDraft] = useState<WorkflowDraftDesignerDraft | null>(null);
  const [workflowDraftEditDirty, setWorkflowDraftEditDirty] = useState(false);
  const savedDraftLibraryScopeKey = JSON.stringify([activeSavedDraftConsumerConfig, workflowScopedApplicationId, generationKey]);
  const [editorScope, setEditorScope] = useState(savedDraftLibraryScopeKey);
  if (editorScope !== savedDraftLibraryScopeKey) {
    // Reset before children render: equal draft IDs in another owner scope are not the same content.
    setEditorScope(savedDraftLibraryScopeKey);
    setWorkspaceCreatedDrafts([]);
    setEditableWorkflowDraft(null);
    setWorkflowDraftEditDirty(false);
    setSavedDraftConsumerState(initialWorkflowSavedDraftConsumerState(activeSavedDraftConsumerConfig));
    pendingSavedDraftConsumerStateRef.current = null;
  }
  const library = useWorkflowSavedDraftLibrary({ config: activeSavedDraftConsumerConfig, applicationId: workflowScopedApplicationId, generationKey });
  const { activeSavedDraftListState, savedDraftLibraryFilters, refreshSavedWorkflowDraftList } = library;
  const workflowWorkspaceContext = useMemo(
    () =>
      buildWorkflowWorkspaceContextViewModel({
        workspaceApplications,
        workspaceApiKeys,
        workspaceUsageQuota,
        workspaceWorkflowDefinitions,
        workspaceRunHistory,
        localWorkflowDrafts: workspaceCreatedDrafts,
        activeWorkflowDraftOverride: editableWorkflowDraft,
        savedDraftConsumerState,
        savedDraftListStatus: activeSavedDraftListState.status,
        savedDraftListFailureCode: activeSavedDraftListState.failureCode,
        savedDraftSummaries: activeSavedDraftListState.summaries,
        selection: {
          applicationRef: selectedApplicationRef,
          workflowDefinitionId: selectedWorkflowDefinitionId,
          runId: selectedRunId,
          draftId: selectedWorkflowDraftId,
          scenarioId: selectedWorkflowScenarioId,
        },
      }),
    [
      workspaceApplications,
      workspaceApiKeys,
      workspaceUsageQuota,
      workspaceWorkflowDefinitions,
      workspaceRunHistory,
      workspaceCreatedDrafts,
      editableWorkflowDraft,
      savedDraftConsumerState,
      activeSavedDraftListState.failureCode,
      activeSavedDraftListState.status,
      activeSavedDraftListState.summaries,
      selectedApplicationRef,
      selectedWorkflowDefinitionId,
      selectedRunId,
      selectedWorkflowDraftId,
      selectedWorkflowScenarioId,
    ],
  );
  const { selectedWorkflowDraft, activeWorkflowDraft, workflowDraftDesigner } = workflowWorkspaceContext;
  const savedDraftEditorRequests = useRef(createWorkflowDraftEditorRequests()).current;
  savedDraftEditorRequests.setScope(JSON.stringify([
    savedDraftLibraryScopeKey,
    generationKey,
    selectedWorkflowDraft.draftId,
  ]));
  useEffect(() => () => savedDraftEditorRequests.invalidate(), [savedDraftEditorRequests]);
  const savedDraftConflictOpenSummary = useMemo(
    () =>
      activeSavedDraftListState.summaries.find(
        (summary) =>
          summary.draftId === activeWorkflowDraft.draftId &&
          summary.applicationRef === activeWorkflowDraft.applicationRef,
      ) ?? null,
    [activeSavedDraftListState.summaries, activeWorkflowDraft.applicationRef, activeWorkflowDraft.draftId],
  );
  const createdWorkspaceDraftCountsByDefinition = useMemo(
    () =>
      workspaceCreatedDrafts.reduce<Record<string, number>>((counts, draft) => {
        counts[draft.workflowDefinitionId] = (counts[draft.workflowDefinitionId] ?? 0) + 1;
        return counts;
      }, {}),
    [workspaceCreatedDrafts],
  );
  useEffect(() => {
    setEditableWorkflowDraft(cloneWorkflowDraftForEditing(selectedWorkflowDraft));
    const pendingConsumerState = pendingSavedDraftConsumerStateRef.current;
    if (pendingConsumerState) {
      pendingSavedDraftConsumerStateRef.current = null;
      if (pendingConsumerState.draftId === selectedWorkflowDraft.draftId) {
        setSavedDraftConsumerState(pendingConsumerState.state);
        setWorkflowDraftEditDirty(pendingConsumerState.state.status === "unsaved_local");
        return;
      }
    }
    if (selectedWorkflowDraft.localOnlyInteraction === "local_edit") {
      setWorkflowDraftEditDirty(true);
      setSavedDraftConsumerState(workspaceDraftCreatedConsumerState(activeSavedDraftConsumerConfig, selectedWorkflowDraft));
      return;
    }
    setSavedDraftConsumerState(initialWorkflowSavedDraftConsumerState(activeSavedDraftConsumerConfig));
    setWorkflowDraftEditDirty(false);
  }, [activeSavedDraftConsumerConfig, workflowScopedApplicationId, generationKey, selectedWorkflowDraft.draftId]);

  const clearWorkspace = () => {
    savedDraftEditorRequests.invalidate();
    library.invalidatePending();
    pendingSavedDraftConsumerStateRef.current = null;
    setWorkspaceCreatedDrafts([]);
    setEditableWorkflowDraft(null);
    setWorkflowDraftEditDirty(false);
    setSavedDraftConsumerState(initialWorkflowSavedDraftConsumerState(activeSavedDraftConsumerConfig));
  };
  const markWorkflowDraftLocallyEdited = () => {
    savedDraftEditorRequests.invalidate();
    setWorkflowDraftEditDirty(true);
    setSavedDraftConsumerState((state) => {
      if (state.status === "version_conflict") {
        return {
          ...state,
          summary:
            "Local edits remain active, but the version conflict still requires explicit Continue local draft or Open saved draft before another dev route action.",
        };
      }
      if (state.status === "conflict_local_continued") {
        return {
          ...state,
          summary: `Local draft has unsaved edits after explicit conflict review; the next save will use saved version ${state.currentDraftVersion}.`,
        };
      }
      return {
        ...state,
        status: "unsaved_local",
        sourceLabel: "unsaved local",
        summary:
          state.mode === "dev_saved_draft_http"
            ? "Local draft has unsaved edits; validate or save through the dev-only saved draft route."
            : "Local draft has unsaved edits and remains in sample-only mode.",
        failureCode: null,
        conflictDraftVersion: null,
      };
    });
  };

  const editWorkflowDraft = (update: (draft: WorkflowDraftDesignerDraft) => WorkflowDraftDesignerDraft): boolean => {
    const next = update(activeWorkflowDraft);
    if (next === activeWorkflowDraft) return false;
    setEditableWorkflowDraft((draft) => update(draft ?? selectedWorkflowDraft));
    markWorkflowDraftLocallyEdited();
    return true;
  };
  const handleWorkflowDraftEditReset = () => {
    savedDraftEditorRequests.invalidate();
    setEditableWorkflowDraft(cloneWorkflowDraftForEditing(selectedWorkflowDraft));
    if (selectedWorkflowDraft.localOnlyInteraction === "local_edit") {
      setWorkflowDraftEditDirty(true);
      setSavedDraftConsumerState(workspaceDraftCreatedConsumerState(activeSavedDraftConsumerConfig, selectedWorkflowDraft));
      return;
    }
    setWorkflowDraftEditDirty(false);
    setSavedDraftConsumerState(initialWorkflowSavedDraftConsumerState(activeSavedDraftConsumerConfig));
  };

  const handleCreateWorkspaceDraftFromDefinition = (workflowDefinitionId: string) => {
    if (workflowExecutorOperationPending) {
      return;
    }
    const createdDraft = buildWorkspaceCreatedDraft(
      workflowDefinitionId,
      workflowDraftDesigner,
      workspaceCreatedDrafts,
    );
    if (!createdDraft) {
      return;
    }
    const nextRun = workspaceRunHistory.runs.find(
      (run) =>
        run.applicationRef === createdDraft.applicationRef &&
        run.workflowDefinitionId === createdDraft.workflowDefinitionId,
    );
    setWorkspaceCreatedDrafts((drafts) => [...drafts, createdDraft]);
    applyWorkflowSelectionPatch({
      applicationRef: createdDraft.applicationRef,
      workflowDefinitionId: createdDraft.workflowDefinitionId,
      runId: nextRun?.runId ?? null,
      draftId: createdDraft.draftId,
      scenarioId: null,
    });
    setEditableWorkflowDraft(cloneWorkflowDraftForEditing(createdDraft));
    setWorkflowDraftEditDirty(true);
    setSavedDraftConsumerState(workspaceDraftCreatedConsumerState(activeSavedDraftConsumerConfig, createdDraft));
  };
  const handleCreateWorkflowExecutorDraft = () => {
    if (workflowExecutorOperationPending) {
      return;
    }
    const nextDraftNumber = workspaceCreatedDrafts.filter(
      (draft) =>
        draft.applicationRef === workflowScopedApplicationId &&
        draft.executionProfile === "executor_v0",
    ).length + 1;
    const createdDraft = buildWorkflowExecutorV0Draft(
      activeWorkflowDraft,
      nextDraftNumber,
      workflowScopedApplicationId,
    );
    setWorkspaceCreatedDrafts((drafts) => [...drafts, createdDraft]);
    applyWorkflowSelectionPatch({
      applicationRef: createdDraft.applicationRef,
      workflowDefinitionId: createdDraft.workflowDefinitionId,
      runId: null,
      draftId: createdDraft.draftId,
      scenarioId: null,
    });
    setEditableWorkflowDraft(cloneWorkflowDraftForEditing(createdDraft));
    setWorkflowDraftEditDirty(true);
    setSavedDraftConsumerState(workspaceDraftCreatedConsumerState(activeSavedDraftConsumerConfig, createdDraft));
    return true;
  };
  const handleCreateWorkflowRAGDraft = (createdDraft: WorkflowDraftDesignerDraft) => {
    if (workflowRAGOperationPending) return;
    setWorkspaceCreatedDrafts((drafts) => [...drafts, createdDraft]);
    applyWorkflowSelectionPatch({ applicationRef: createdDraft.applicationRef, workflowDefinitionId: createdDraft.workflowDefinitionId, runId: null, draftId: createdDraft.draftId, scenarioId: null });
    setEditableWorkflowDraft(cloneWorkflowDraftForEditing(createdDraft));
    setWorkflowDraftEditDirty(true);
    setSavedDraftConsumerState(workspaceDraftCreatedConsumerState(activeSavedDraftConsumerConfig, createdDraft));
  };
  const handleCreateDefinitionDerivedDraft = (createdDraft: WorkflowDraftDesignerDraft) => {
    if (workflowExecutorOperationPending || workflowRAGOperationPending) return;
    setWorkspaceCreatedDrafts((drafts) => [...drafts, createdDraft]);
    applyWorkflowSelectionPatch({ applicationRef: createdDraft.applicationRef, workflowDefinitionId: createdDraft.workflowDefinitionId, runId: null, draftId: createdDraft.draftId, scenarioId: null });
    setEditableWorkflowDraft(cloneWorkflowDraftForEditing(createdDraft));
    setWorkflowDraftEditDirty(true);
    setSavedDraftConsumerState(workspaceDraftCreatedConsumerState(activeSavedDraftConsumerConfig, createdDraft));
  };
  const handleOpenTemplateDerivedDraft = (
    createdDraft: WorkflowDraftDesignerDraft,
    authority: {
      draftId: string;
      draftVersion: number;
      lifecycleVersion: number;
      lifecycleState: "active";
      targetApplicationId: string;
    },
  ) => {
    if (workflowExecutorOperationPending || workflowRAGOperationPending) return;
    const consumerState = workflowTemplateDerivedConsumerState(
      activeSavedDraftConsumerConfig,
      createdDraft,
      authority,
    );
    pendingSavedDraftConsumerStateRef.current = { draftId: authority.draftId, state: consumerState };
    setWorkspaceCreatedDrafts((drafts) => [
      ...drafts.filter((draft) => draft.draftId !== authority.draftId),
      createdDraft,
    ]);
    applyWorkflowSelectionPatch({
      applicationRef: authority.targetApplicationId,
      workflowDefinitionId: createdDraft.workflowDefinitionId,
      runId: null,
      draftId: authority.draftId,
      scenarioId: null,
    });
    setEditableWorkflowDraft(cloneWorkflowDraftForEditing(createdDraft));
    setWorkflowDraftEditDirty(false);
    setSavedDraftConsumerState(consumerState);
    onOpenDesigner();
  };
  const handleDeriveSavedWorkflowDraft = () => {
    const operationPending = workflowExecutorOperationPending || workflowRAGOperationPending;
    if (!canDeriveSavedWorkflowDraft(savedDraftConsumerState, workflowDraftEditDirty, operationPending)) {
      return;
    }
    const createdDraft = buildDerivedWorkflowDraft(
      activeWorkflowDraft,
      savedDraftConsumerState.currentDraftVersion,
      workflowDraftDesigner.drafts.map((draft) => draft.draftId),
    );
    const derivedConsumerState = workspaceDraftCreatedConsumerState(
      activeSavedDraftConsumerConfig,
      createdDraft,
      savedDraftConsumerState.currentLifecycleVersion,
    );
    pendingSavedDraftConsumerStateRef.current = {
      draftId: createdDraft.draftId,
      state: derivedConsumerState,
    };
    setWorkspaceCreatedDrafts((drafts) => [...drafts, createdDraft]);
    applyWorkflowSelectionPatch({
      applicationRef: createdDraft.applicationRef,
      workflowDefinitionId: createdDraft.workflowDefinitionId,
      runId: null,
      draftId: createdDraft.draftId,
      scenarioId: null,
    });
    setEditableWorkflowDraft(cloneWorkflowDraftForEditing(createdDraft));
    setWorkflowDraftEditDirty(true);
    setSavedDraftConsumerState(derivedConsumerState);
  };
  const handleOpenSavedWorkflowDraft = (summary: WorkflowSavedDraftSummary) => {
    if (activeSavedDraftConsumerConfig.mode !== "dev_saved_draft_http") {
      return;
    }
    const isCurrentEditorRequest = savedDraftEditorRequests.begin();
    const isCurrentLibraryRequest = library.beginOpen();
    setSavedDraftConsumerState((state) => ({
      ...state,
      status: "reading",
      summary: summary.lifecycleState === "archived"
        ? `Opening archived saved draft ${summary.draftId} for read-only review.`
        : `Opening saved draft ${summary.draftId} through the dev-only read route.`,
      failureCode: null,
      currentDraftVersion: summary.draftVersion,
      currentLifecycleVersion: summary.lifecycleVersion,
      currentLifecycleState: summary.lifecycleState,
      conflictDraftVersion: null,
    }));
    openWorkflowDraftDevRecord(summary, activeSavedDraftConsumerConfig)
      .then((result) => {
        if (!isCurrentEditorRequest() || !isCurrentLibraryRequest()) {
          return;
        }
        setSavedDraftConsumerState(result.state);
        if (!result.draft) {
          library.recordOpenFailure(summary, result.state.summary, result.state.failureCode ?? "dev_saved_draft_open_failed");
          return;
        }
        const openedDraft = result.draft;
        pendingSavedDraftConsumerStateRef.current = {
          draftId: openedDraft.draftId,
          state: result.state,
        };
        const nextRun = workspaceRunHistory.runs.find(
          (run) =>
            run.applicationRef === openedDraft.applicationRef &&
            run.workflowDefinitionId === openedDraft.workflowDefinitionId,
        );
        setWorkspaceCreatedDrafts((drafts) => [
          ...drafts.filter((draft) => draft.draftId !== openedDraft.draftId),
          openedDraft,
        ]);
        applyWorkflowSelectionPatch({
          applicationRef: openedDraft.applicationRef,
          workflowDefinitionId: openedDraft.workflowDefinitionId,
          runId: nextRun?.runId ?? null,
          draftId: openedDraft.draftId,
          scenarioId: null,
        });
        setEditableWorkflowDraft(cloneWorkflowDraftForEditing(openedDraft));
        setWorkflowDraftEditDirty(false);
        onOpenDesigner();
      })
      .catch((error: unknown) => {
        if (!isCurrentEditorRequest() || !isCurrentLibraryRequest()) {
          return;
        }
        const message = error instanceof Error ? error.message : "Saved draft open failed.";
        setSavedDraftConsumerState((state) => ({
          ...state,
          status: "read_failed",
          sourceLabel: "open_failed",
          summary: message,
          failureCode: "dev_saved_draft_open_failed",
          conflictDraftVersion: null,
        }));
        library.recordOpenFailure(summary, message, "dev_saved_draft_open_failed");
      });
  };
  const handleSavedWorkflowDraftLifecycleTransition = (summary: WorkflowSavedDraftSummary, targetState: WorkflowSavedDraftLifecycleState) => {
    const isCurrentEditorRequest = savedDraftEditorRequests.current();
    return library.transition(summary, targetState, {
      draftId: activeWorkflowDraft.draftId, dirty: workflowDraftEditDirty,
      onApplied: (result) => {
        if (!isCurrentEditorRequest() || activeWorkflowDraft.draftId !== summary.draftId) return;
        savedDraftEditorRequests.invalidate();
        setSavedDraftConsumerState((state) => ({
          ...state, status: "saved_dev_record",
          sourceLabel: targetState === "active" ? "reopen required" : "archived read-only",
          currentDraftVersion: result.currentDraftVersion,
          currentLifecycleVersion: result.currentLifecycleVersion,
          currentLifecycleState: targetState === "active" ? "unknown" : result.currentLifecycleState,
          summary: targetState === "active"
            ? `${result.summary} The existing browser draft remains read-only until it is opened again.` : result.summary,
          failureCode: null, requestId: result.requestId, auditRef: result.auditRef,
        }));
        setWorkflowDraftEditDirty(false);
      },
    });
  };
  const handleContinueLocalWorkflowDraftAfterConflict = () => {
    setSavedDraftConsumerState((state) =>
      continueLocalWorkflowDraftAfterVersionConflict(state, activeWorkflowDraft),
    );
    setWorkflowDraftEditDirty(true);
  };
  const handleOpenConflictSavedWorkflowDraft = () => {
    if (!savedDraftConflictOpenSummary) {
      return;
    }
    handleOpenSavedWorkflowDraft(savedDraftConflictOpenSummary);
  };
  const handleValidateWorkflowDraft = () => {
    if (
      activeSavedDraftConsumerConfig.mode !== "dev_saved_draft_http" ||
      (savedDraftConsumerState.currentDraftVersion > 0 &&
        savedDraftConsumerState.currentLifecycleState !== "active") ||
      workflowSavedDraftConflictRequiresResolution(savedDraftConsumerState)
    ) {
      return;
    }
    const currentDraftVersion = savedDraftConsumerState.currentDraftVersion;
    const isCurrentEditorRequest = savedDraftEditorRequests.begin();
    setSavedDraftConsumerState((state) => ({
      ...state,
      status: "validating",
      summary: "Validating local draft through the dev-only saved draft route.",
      failureCode: null,
      conflictDraftVersion: null,
    }));
    validateWorkflowDraftDevRecord(
      activeWorkflowDraft,
      activeSavedDraftConsumerConfig,
      currentDraftVersion,
      savedDraftConsumerState.currentLifecycleVersion,
      savedDraftConsumerState.currentLifecycleState,
    )
      .then((nextState) => {
        if (isCurrentEditorRequest()) setSavedDraftConsumerState(nextState);
      })
      .catch((error: unknown) => {
        if (!isCurrentEditorRequest()) return;
        setSavedDraftConsumerState((state) => ({
          ...state,
          status: "validation_failed",
          sourceLabel: "validation_failed",
          summary: error instanceof Error ? error.message : "Saved draft validation failed.",
          failureCode: "dev_saved_draft_consumer_failed",
          conflictDraftVersion: null,
        }));
      });
  };
  const handleSaveWorkflowDraft = () => {
    if (
      activeSavedDraftConsumerConfig.mode !== "dev_saved_draft_http" ||
      (savedDraftConsumerState.currentDraftVersion > 0 &&
        savedDraftConsumerState.currentLifecycleState !== "active")
    ) {
      return;
    }
    const expectedDraftVersion = nextWorkflowSavedDraftExpectedVersion(savedDraftConsumerState);
    if (expectedDraftVersion === null) {
      return;
    }
    const isCurrentEditorRequest = savedDraftEditorRequests.begin();
    setSavedDraftConsumerState((state) => ({
      ...state,
      status: "saving",
      summary: "Saving local draft through the dev-only saved draft route.",
      failureCode: null,
      conflictDraftVersion: null,
    }));
    saveWorkflowDraftDevRecord(
      activeWorkflowDraft,
      activeSavedDraftConsumerConfig,
      expectedDraftVersion,
      savedDraftConsumerState.currentLifecycleVersion,
    )
      .then((nextState) => {
        if (!isCurrentEditorRequest()) return;
        setSavedDraftConsumerState(nextState);
        if (nextState.status === "version_conflict") {
          refreshSavedWorkflowDraftList(
            activeWorkflowDraft.applicationRef,
            "active",
            savedDraftLibraryFilters,
          );
          return;
        }
        if (nextState.status === "saved_dev_record") {
          setWorkspaceCreatedDrafts((drafts) =>
            drafts.map((draft) =>
              draft.draftId === activeWorkflowDraft.draftId
                ? { ...activeWorkflowDraft, localOnlyInteraction: "inspect_only" }
                : draft,
            ),
          );
          setEditableWorkflowDraft((draft) =>
            draft === null ? null : { ...draft, localOnlyInteraction: "inspect_only" },
          );
          setWorkflowDraftEditDirty(false);
          refreshSavedWorkflowDraftList(
            activeWorkflowDraft.applicationRef,
            "active",
            savedDraftLibraryFilters,
          );
        }
      })
      .catch((error: unknown) => {
        if (!isCurrentEditorRequest()) return;
        setSavedDraftConsumerState((state) => ({
          ...state,
          status: "save_failed",
          sourceLabel: "save_failed",
          summary: error instanceof Error ? error.message : "Saved draft save failed.",
          failureCode: "dev_saved_draft_consumer_failed",
          conflictDraftVersion: null,
        }));
      });
  };
  const handleReadWorkflowDraft = () => {
    if (
      activeSavedDraftConsumerConfig.mode !== "dev_saved_draft_http" ||
      workflowSavedDraftConflictRequiresResolution(savedDraftConsumerState)
    ) {
      return;
    }
    const currentDraftVersion = savedDraftConsumerState.currentDraftVersion;
    const isCurrentEditorRequest = savedDraftEditorRequests.begin();
    setSavedDraftConsumerState((state) => ({
      ...state,
      status: "reading",
      summary: "Reading local draft through the dev-only saved draft route.",
      failureCode: null,
      conflictDraftVersion: null,
    }));
    readWorkflowDraftDevRecord(activeWorkflowDraft, activeSavedDraftConsumerConfig, currentDraftVersion)
      .then((nextState) => {
        if (isCurrentEditorRequest()) setSavedDraftConsumerState(nextState);
      })
      .catch((error: unknown) => {
        if (!isCurrentEditorRequest()) return;
        setSavedDraftConsumerState((state) => ({
          ...state,
          status: "read_failed",
          sourceLabel: "read_failed",
          summary: error instanceof Error ? error.message : "Saved draft read failed.",
          failureCode: "dev_saved_draft_consumer_failed",
          conflictDraftVersion: null,
        }));
      });
  };
  const revisionRequestIsCurrent = savedDraftEditorRequests.current();
  const handleWorkflowDraftRevisionRestored = (
    restoredDraft: WorkflowDraftDesignerDraft,
    result: WorkflowSavedDraftRevisionRestoreResult,
  ) => {
    if (!revisionRequestIsCurrent() || restoredDraft.draftId !== activeWorkflowDraft.draftId || restoredDraft.applicationRef !== activeWorkflowDraft.applicationRef) return;
    savedDraftEditorRequests.invalidate();
    setWorkspaceCreatedDrafts((drafts) => [
      ...drafts.filter((draft) => draft.draftId !== restoredDraft.draftId),
      restoredDraft,
    ]);
    setEditableWorkflowDraft(cloneWorkflowDraftForEditing(restoredDraft));
    setWorkflowDraftEditDirty(false);
    setSavedDraftConsumerState({
      status: result.failureCode ? "save_failed" : "saved_dev_record",
      mode: "dev_saved_draft_http",
      sourceLabel: result.failureCode ?? "restored revision",
      summary: result.summary,
      failureCode: result.failureCode,
      currentDraftVersion: result.currentDraftVersion,
      currentLifecycleVersion: result.currentLifecycleVersion,
      currentLifecycleState: result.currentLifecycleState,
      conflictDraftVersion: null,
      auditRef: result.auditRef,
      requestId: result.requestId,
    });
    refreshSavedWorkflowDraftList(restoredDraft.applicationRef, "active", savedDraftLibraryFilters);
  };

  return {
    library, workflowWorkspaceContext, savedDraftConsumerState, workflowDraftEditDirty,
    editorScopeKey: JSON.stringify([savedDraftLibraryScopeKey, selectedWorkflowDraft.draftId]),
    savedDraftConflictOpenSummary, createdWorkspaceDraftCountsByDefinition,
    nextDefinitionDerivedDraftNumber: workspaceCreatedDrafts.filter((draft) => draft.applicationRef === workflowScopedApplicationId && (draft.baseDefinitionVersion ?? 0) > 0).length + 1,
    nextRAGDraftNumber: workspaceCreatedDrafts.filter((draft) => draft.applicationRef === workflowScopedApplicationId && draft.executionProfile === "rag_retrieval_v1").length + 1,
    editWorkflowDraft,
    bindRAGRef: (nodeId: string, ragRef: string) => editWorkflowDraft((draft) => ({
      ...draft, localOnlyInteraction: "local_edit", nodes: draft.nodes.map((node) => node.nodeId === nodeId ? { ...node, ragRef } : node),
    })),
    handleWorkflowDraftEditReset, handleCreateWorkspaceDraftFromDefinition,
    handleCreateWorkflowExecutorDraft, handleCreateWorkflowRAGDraft, handleCreateDefinitionDerivedDraft,
    handleOpenTemplateDerivedDraft, handleDeriveSavedWorkflowDraft, handleOpenSavedWorkflowDraft,
    handleSavedWorkflowDraftLifecycleTransition, handleContinueLocalWorkflowDraftAfterConflict,
    handleOpenConflictSavedWorkflowDraft, handleValidateWorkflowDraft, handleSaveWorkflowDraft,
    handleReadWorkflowDraft, handleWorkflowDraftRevisionRestored,
    invalidateSelection: savedDraftEditorRequests.invalidate, clearWorkspace,
  };
}
