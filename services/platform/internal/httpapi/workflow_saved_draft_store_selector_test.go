package httpapi

import "testing"

func TestSelectWorkflowSavedDraftStore(t *testing.T) {
	t.Run("empty and memory dev mode use memory store", func(t *testing.T) {
		defaultSelection := SelectWorkflowSavedDraftStore("", WorkflowSavedDraftStoreSelector{})
		if defaultSelection.Mode != WorkflowSavedDraftStoreModeMemoryDev ||
			defaultSelection.Store == nil ||
			defaultSelection.FailureCode != "" {
			t.Fatalf("empty mode should select memory_dev: %#v", defaultSelection)
		}

		memoryStore := newMemorySavedWorkflowDraftStore()
		selection := SelectWorkflowSavedDraftStore(" memory_dev ", WorkflowSavedDraftStoreSelector{
			MemoryDevStore: memoryStore,
		})
		if selection.Mode != WorkflowSavedDraftStoreModeMemoryDev ||
			selection.Store != memoryStore ||
			selection.FailureCode != "" {
			t.Fatalf("memory_dev mode should reuse provided memory store: %#v", selection)
		}

		service := newSavedWorkflowDraftService(selection.Store)
		saveResult := service.SaveDraft(savedWorkflowDraftTestContext(), SaveWorkflowDraftRequest{
			ExpectedDraftVersion: 0,
			Payload:              validSavedWorkflowDraftPayload(),
		})
		if saveResult.FailureCode != "" || saveResult.Draft == nil {
			t.Fatalf("memory_dev selection should preserve save behavior: %#v", saveResult)
		}
		if got := memoryStore.SideEffects(); got.DraftWriteCount != 1 || hasSavedWorkflowDraftRuntimeSideEffect(got) {
			t.Fatalf("memory_dev selection should only record draft write side effect: %#v", got)
		}
	})

	t.Run("reserved repository modes fail closed without memory fallback", func(t *testing.T) {
		for _, mode := range []string{"repository_disabled", "repository"} {
			t.Run(mode, func(t *testing.T) {
				selection := SelectWorkflowSavedDraftStore(mode, WorkflowSavedDraftStoreSelector{})
				if selection.Store == nil ||
					selection.FailureCode != SavedWorkflowDraftFailureRepositoryStoreDisabled {
					t.Fatalf("reserved mode should select disabled store: %#v", selection)
				}
				assertSavedWorkflowDraftStoreSelectionFailure(
					t,
					selection.Store,
					SavedWorkflowDraftFailureRepositoryStoreDisabled,
				)
			})
		}
	})

	t.Run("postgres dev test mode requires an explicitly constructed store", func(t *testing.T) {
		disabled := SelectWorkflowSavedDraftStore("postgres_dev_test", WorkflowSavedDraftStoreSelector{})
		if disabled.Mode != WorkflowSavedDraftStoreModePostgresDevTest ||
			disabled.FailureCode != SavedWorkflowDraftFailureStoreUnavailable {
			t.Fatalf("missing postgres dev/test store must fail closed: %#v", disabled)
		}
		assertSavedWorkflowDraftStoreSelectionFailure(
			t,
			disabled.Store,
			SavedWorkflowDraftFailureStoreUnavailable,
		)

		injectedStore := newMemorySavedWorkflowDraftStore()
		selected := SelectWorkflowSavedDraftStore(" postgres_dev_test ", WorkflowSavedDraftStoreSelector{
			PostgresDevTestStore: injectedStore,
		})
		if selected.Mode != WorkflowSavedDraftStoreModePostgresDevTest ||
			selected.Store != injectedStore ||
			selected.FailureCode != "" {
			t.Fatalf("explicit postgres dev/test store should be selected: %#v", selected)
		}
	})

	t.Run("unknown mode fails closed without memory fallback", func(t *testing.T) {
		selection := SelectWorkflowSavedDraftStore("future_backend", WorkflowSavedDraftStoreSelector{})
		if selection.Store == nil ||
			selection.FailureCode != SavedWorkflowDraftFailureInvalidStoreMode {
			t.Fatalf("unknown mode should select disabled store: %#v", selection)
		}
		assertSavedWorkflowDraftStoreSelectionFailure(
			t,
			selection.Store,
			SavedWorkflowDraftFailureInvalidStoreMode,
		)
	})
}

func assertSavedWorkflowDraftStoreSelectionFailure(
	t *testing.T,
	store savedWorkflowDraftStore,
	expectedFailureCode SavedWorkflowDraftFailureCode,
) {
	t.Helper()
	service := newSavedWorkflowDraftService(store)
	context := savedWorkflowDraftTestContext()
	payload := validSavedWorkflowDraftPayload()

	saveResult := service.SaveDraft(context, SaveWorkflowDraftRequest{
		ExpectedDraftVersion: 0,
		Payload:              payload,
	})
	if saveResult.FailureCode != expectedFailureCode || saveResult.Draft != nil {
		t.Fatalf("save should fail closed with %s: %#v", expectedFailureCode, saveResult)
	}

	readResult := service.ReadDraft(context, ReadWorkflowDraftRequest{DraftID: payload.DraftID})
	if readResult.FailureCode != expectedFailureCode || readResult.Draft != nil {
		t.Fatalf("read should fail closed with %s: %#v", expectedFailureCode, readResult)
	}

	listResult := service.ListDrafts(context, ListWorkflowDraftsRequest{})
	if listResult.FailureCode != expectedFailureCode || len(listResult.Summaries) != 0 {
		t.Fatalf("list should fail closed with %s: %#v", expectedFailureCode, listResult)
	}

	if got := store.SideEffects(); got.DraftWriteCount != 0 || hasSavedWorkflowDraftRuntimeSideEffect(got) {
		t.Fatalf("disabled selection must not emit side effects: %#v", got)
	}
}
