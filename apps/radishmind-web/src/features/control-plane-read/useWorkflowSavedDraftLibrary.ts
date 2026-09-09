import { useEffect, useRef, useState } from "react";
import {
  archiveWorkflowDraftDevRecord, unarchiveWorkflowDraftDevRecord,
  emptyWorkflowSavedDraftLibraryFilters, initialWorkflowSavedDraftLifecycleOperationState,
  initialWorkflowSavedDraftListState, listWorkflowDraftDevRecords, mergeWorkflowSavedDraftListPage,
  workflowSavedDraftRequestIsCurrent,
  type WorkflowSavedDraftConsumerConfig, type WorkflowSavedDraftLibraryFilters,
  type WorkflowSavedDraftLifecycleOperationState, type WorkflowSavedDraftLifecycleState,
  type WorkflowSavedDraftListState, type WorkflowSavedDraftSummary,
} from "./savedWorkflowDraftConsumer.ts";

export function useWorkflowSavedDraftLibrary({ config: activeSavedDraftConsumerConfig, applicationId: workflowScopedApplicationId, generationKey }: {
  config: WorkflowSavedDraftConsumerConfig;
  applicationId: string;
  generationKey: string;
}) {
  const [savedDraftLibraryLifecycle, setSavedDraftLibraryLifecycle] =
    useState<WorkflowSavedDraftLifecycleState>("active");
  const [savedDraftLibraryFilters, setSavedDraftLibraryFilters] =
    useState<WorkflowSavedDraftLibraryFilters>(() => emptyWorkflowSavedDraftLibraryFilters());
  const [savedDraftListStates, setSavedDraftListStates] = useState<
    Record<WorkflowSavedDraftLifecycleState, WorkflowSavedDraftListState>
  >(() => ({
    active: initialWorkflowSavedDraftListState(activeSavedDraftConsumerConfig, "", "active"),
    archived: initialWorkflowSavedDraftListState(activeSavedDraftConsumerConfig, "", "archived"),
  }));
  const savedDraftListRequestGenerationRef = useRef<Record<WorkflowSavedDraftLifecycleState, number>>({
    active: 0,
    archived: 0,
  });
  const savedDraftLifecycleOperationGenerationRef = useRef(0);
  const savedDraftOpenRequestGenerationRef = useRef(0);
  const [savedDraftLifecycleOperation, setSavedDraftLifecycleOperation] =
    useState<WorkflowSavedDraftLifecycleOperationState>(() =>
      initialWorkflowSavedDraftLifecycleOperationState()
    );
  const savedDraftListState = savedDraftListStates[savedDraftLibraryLifecycle];
  const activeSavedDraftListState = savedDraftListStates.active;
  const savedDraftLibraryScopeKey = JSON.stringify([activeSavedDraftConsumerConfig, workflowScopedApplicationId, generationKey]);
  const [listScope, setListScope] = useState(savedDraftLibraryScopeKey);
  if (listScope !== savedDraftLibraryScopeKey) {
    setListScope(savedDraftLibraryScopeKey);
    setSavedDraftLibraryLifecycle("active");
    setSavedDraftLibraryFilters(emptyWorkflowSavedDraftLibraryFilters());
    setSavedDraftLifecycleOperation(initialWorkflowSavedDraftLifecycleOperationState());
    setSavedDraftListStates({
      active: initialWorkflowSavedDraftListState(activeSavedDraftConsumerConfig, workflowScopedApplicationId, "active"),
      archived: initialWorkflowSavedDraftListState(activeSavedDraftConsumerConfig, workflowScopedApplicationId, "archived"),
    });
  }
  const savedDraftLibraryScopeKeyRef = useRef(savedDraftLibraryScopeKey);
  savedDraftLibraryScopeKeyRef.current = savedDraftLibraryScopeKey;
  const invalidatePending = () => {
    savedDraftListRequestGenerationRef.current.active += 1;
    savedDraftListRequestGenerationRef.current.archived += 1;
    savedDraftLifecycleOperationGenerationRef.current += 1;
    savedDraftOpenRequestGenerationRef.current += 1;
  };
  useEffect(() => invalidatePending, []);
  const beginOpen = () => {
    const generation = ++savedDraftOpenRequestGenerationRef.current;
    const scope = savedDraftLibraryScopeKeyRef.current;
    return () => workflowSavedDraftRequestIsCurrent(generation, savedDraftOpenRequestGenerationRef.current, scope, savedDraftLibraryScopeKeyRef.current);
  };
  const recordOpenFailure = (summary: WorkflowSavedDraftSummary, message: string, failureCode: string) => {
    setSavedDraftListStates((states) => ({ ...states, [summary.lifecycleState]: {
      ...states[summary.lifecycleState], status: "open_failed", sourceLabel: "open_failed", summary: message, failureCode,
    } }));
  };
  const refreshSavedWorkflowDraftList = (
    applicationRef: string,
    lifecycleState: WorkflowSavedDraftLifecycleState = savedDraftLibraryLifecycle,
    filters: WorkflowSavedDraftLibraryFilters = savedDraftLibraryFilters,
    append = false,
  ) => {
    const current = savedDraftListStates[lifecycleState];
    if (append && (!current.hasMore || current.status === "loading")) return;
    const generation = ++savedDraftListRequestGenerationRef.current[lifecycleState];
    const requestScopeKey = savedDraftLibraryScopeKeyRef.current;
    if (activeSavedDraftConsumerConfig.mode !== "dev_saved_draft_http" || !applicationRef) {
      setSavedDraftListStates((states) => ({
        ...states,
        [lifecycleState]: initialWorkflowSavedDraftListState(
          activeSavedDraftConsumerConfig,
          applicationRef,
          lifecycleState,
          filters,
        ),
      }));
      return;
    }
    const cursor = append ? current.nextCursor : "";
    setSavedDraftListStates((states) => ({
      ...states,
      [lifecycleState]: {
        ...states[lifecycleState],
        status: "loading",
        mode: "dev_saved_draft_http",
        sourceLabel: cursor ? "loading more" : "loading",
        summary: cursor
          ? `Loading more ${lifecycleState} saved drafts.`
          : `Loading ${lifecycleState} saved drafts for the selected application.`,
        applicationRef,
        lifecycleState,
        filters,
        failureCode: null,
        ...(cursor ? {} : { summaries: [], nextCursor: "", hasMore: false }),
      },
    }));
    listWorkflowDraftDevRecords(applicationRef, activeSavedDraftConsumerConfig, {
      lifecycleState,
      filters,
      cursor,
      limit: 25,
    })
      .then((page) => {
        if (!workflowSavedDraftRequestIsCurrent(
          generation,
          savedDraftListRequestGenerationRef.current[lifecycleState],
          requestScopeKey,
          savedDraftLibraryScopeKeyRef.current,
        )) {
          return;
        }
        setSavedDraftListStates((states) => ({
          ...states,
          [lifecycleState]: cursor
            ? mergeWorkflowSavedDraftListPage(states[lifecycleState], page)
            : page,
        }));
      })
      .catch((error: unknown) => {
        if (!workflowSavedDraftRequestIsCurrent(
          generation,
          savedDraftListRequestGenerationRef.current[lifecycleState],
          requestScopeKey,
          savedDraftLibraryScopeKeyRef.current,
        )) {
          return;
        }
        setSavedDraftListStates((states) => ({
          ...states,
          [lifecycleState]: {
            ...states[lifecycleState],
            status: "list_failed",
            sourceLabel: "list_failed",
            summary: error instanceof Error ? error.message : "Saved draft list failed.",
            applicationRef,
            lifecycleState,
            filters,
            failureCode: "dev_saved_draft_list_failed",
            ...(cursor ? {} : { summaries: [], nextCursor: "", hasMore: false }),
          },
        }));
      });
  };
  useEffect(() => {
    savedDraftListRequestGenerationRef.current.active += 1;
    savedDraftListRequestGenerationRef.current.archived += 1;
    savedDraftLifecycleOperationGenerationRef.current += 1;
    savedDraftOpenRequestGenerationRef.current += 1;
    setSavedDraftLibraryLifecycle("active");
    setSavedDraftLibraryFilters(emptyWorkflowSavedDraftLibraryFilters());
    setSavedDraftLifecycleOperation(initialWorkflowSavedDraftLifecycleOperationState());
    setSavedDraftListStates({
      active: initialWorkflowSavedDraftListState(
        activeSavedDraftConsumerConfig,
        workflowScopedApplicationId,
        "active",
      ),
      archived: initialWorkflowSavedDraftListState(
        activeSavedDraftConsumerConfig,
        workflowScopedApplicationId,
        "archived",
      ),
    });
    refreshSavedWorkflowDraftList(
      workflowScopedApplicationId,
      "active",
      emptyWorkflowSavedDraftLibraryFilters(),
    );
  }, [
    activeSavedDraftConsumerConfig,
    generationKey,
    workflowScopedApplicationId,
  ]);
  const handleRefreshSavedWorkflowDraftList = () => {
    refreshSavedWorkflowDraftList(
      workflowScopedApplicationId,
      savedDraftLibraryLifecycle,
      savedDraftLibraryFilters,
    );
  };
  const handleSavedDraftLibraryLifecycleChange = (lifecycleState: WorkflowSavedDraftLifecycleState) => {
    savedDraftListRequestGenerationRef.current[lifecycleState] += 1;
    savedDraftLifecycleOperationGenerationRef.current += 1;
    savedDraftOpenRequestGenerationRef.current += 1;
    setSavedDraftLibraryLifecycle(lifecycleState);
    setSavedDraftLifecycleOperation(initialWorkflowSavedDraftLifecycleOperationState());
    setSavedDraftListStates((states) => ({
      ...states,
      [lifecycleState]: initialWorkflowSavedDraftListState(
        activeSavedDraftConsumerConfig,
        workflowScopedApplicationId,
        lifecycleState,
        savedDraftLibraryFilters,
      ),
    }));
    refreshSavedWorkflowDraftList(
      workflowScopedApplicationId,
      lifecycleState,
      savedDraftLibraryFilters,
    );
  };
  const handleSavedDraftLibraryFiltersChange = (filters: WorkflowSavedDraftLibraryFilters) => {
    savedDraftListRequestGenerationRef.current.active += 1;
    savedDraftListRequestGenerationRef.current.archived += 1;
    savedDraftLifecycleOperationGenerationRef.current += 1;
    savedDraftOpenRequestGenerationRef.current += 1;
    setSavedDraftLibraryFilters(filters);
    setSavedDraftLifecycleOperation(initialWorkflowSavedDraftLifecycleOperationState());
    setSavedDraftListStates({
      active: initialWorkflowSavedDraftListState(
        activeSavedDraftConsumerConfig,
        workflowScopedApplicationId,
        "active",
        filters,
      ),
      archived: initialWorkflowSavedDraftListState(
        activeSavedDraftConsumerConfig,
        workflowScopedApplicationId,
        "archived",
        filters,
      ),
    });
    refreshSavedWorkflowDraftList(workflowScopedApplicationId, savedDraftLibraryLifecycle, filters);
  };
  const handleLoadMoreSavedWorkflowDrafts = () => {
    refreshSavedWorkflowDraftList(
      workflowScopedApplicationId,
      savedDraftLibraryLifecycle,
      savedDraftLibraryFilters,
      true,
    );
  };
  const handleSavedWorkflowDraftLifecycleTransition = async (
    summary: WorkflowSavedDraftSummary,
    targetState: WorkflowSavedDraftLifecycleState,
    editor: { draftId: string; dirty: boolean; onApplied: (result: WorkflowSavedDraftLifecycleOperationState) => void },
  ) => {
    if (activeSavedDraftConsumerConfig.mode !== "dev_saved_draft_http") {
      return;
    }
    const operationGeneration = savedDraftLifecycleOperationGenerationRef.current + 1;
    savedDraftLifecycleOperationGenerationRef.current = operationGeneration;
    const operationScopeKey = savedDraftLibraryScopeKeyRef.current;
    if (
      targetState === "archived" &&
      editor.draftId === summary.draftId &&
      editor.dirty
    ) {
      setSavedDraftLifecycleOperation({
        status: "failed",
        draftId: summary.draftId,
        targetState,
        currentDraftVersion: summary.draftVersion,
        currentLifecycleVersion: summary.lifecycleVersion,
        currentLifecycleState: summary.lifecycleState,
        failureCode: "draft_local_edits_pending",
        requestId: `saved-draft-${targetState}-${summary.draftId}`,
        auditRef: "not_sent",
        summary: "Save or reset local edits before archiving this exact saved draft version.",
      });
      return;
    }
    setSavedDraftLifecycleOperation({
      status: "transitioning",
      draftId: summary.draftId,
      targetState,
      currentDraftVersion: summary.draftVersion,
      currentLifecycleVersion: summary.lifecycleVersion,
      currentLifecycleState: summary.lifecycleState,
      failureCode: null,
      requestId: `saved-draft-${targetState}-${summary.draftId}`,
      auditRef: "pending",
      summary: `${targetState === "archived" ? "Archiving" : "Unarchiving"} ${summary.draftId}.`,
    });
    try {
      const result = targetState === "archived"
        ? await archiveWorkflowDraftDevRecord(summary, activeSavedDraftConsumerConfig)
        : await unarchiveWorkflowDraftDevRecord(summary, activeSavedDraftConsumerConfig);
      if (!workflowSavedDraftRequestIsCurrent(
        operationGeneration,
        savedDraftLifecycleOperationGenerationRef.current,
        operationScopeKey,
        savedDraftLibraryScopeKeyRef.current,
      )) {
        return;
      }
      setSavedDraftLifecycleOperation(result);
      if (result.status === "failed") {
        return;
      }
      editor.onApplied(result);
      refreshSavedWorkflowDraftList(workflowScopedApplicationId, "active", savedDraftLibraryFilters);
      refreshSavedWorkflowDraftList(workflowScopedApplicationId, "archived", savedDraftLibraryFilters);
    } catch (error) {
      if (!workflowSavedDraftRequestIsCurrent(
        operationGeneration,
        savedDraftLifecycleOperationGenerationRef.current,
        operationScopeKey,
        savedDraftLibraryScopeKeyRef.current,
      )) {
        return;
      }
      setSavedDraftLifecycleOperation({
        status: "failed",
        draftId: summary.draftId,
        targetState,
        currentDraftVersion: summary.draftVersion,
        currentLifecycleVersion: summary.lifecycleVersion,
        currentLifecycleState: summary.lifecycleState,
        failureCode: "dev_saved_draft_lifecycle_request_failed",
        requestId: `saved-draft-${targetState}-${summary.draftId}`,
        auditRef: "unavailable",
        summary: error instanceof Error ? error.message : "Saved draft lifecycle request failed.",
      });
    }
  };

  return {
    savedDraftLibraryLifecycle, savedDraftLibraryFilters, savedDraftListState, activeSavedDraftListState,
    savedDraftLifecycleOperation, refreshSavedWorkflowDraftList, handleRefreshSavedWorkflowDraftList,
    handleSavedDraftLibraryLifecycleChange, handleSavedDraftLibraryFiltersChange, handleLoadMoreSavedWorkflowDrafts,
    transition: handleSavedWorkflowDraftLifecycleTransition, beginOpen, recordOpenFailure, invalidatePending,
  };
}
