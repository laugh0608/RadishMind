package httpapi

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestSavedWorkflowDraftRepositoryAdapterSaveReadList(t *testing.T) {
	executor := newFakeSavedWorkflowDraftRepositoryQueryExecutor()
	adapter := NewSavedWorkflowDraftRepositoryAdapter(SavedWorkflowDraftRepositoryAdapterConfig{
		QueryExecutor: executor,
	})
	actor := validSavedWorkflowDraftRepositoryActorContext()
	draft := validSavedWorkflowDraftRepositoryDraft()

	saveResult := adapter.SaveWorkflowDraftRecord(actor, SaveWorkflowDraftRecordRequest{
		ExpectedDraftVersion: 0,
		Draft:                draft,
	})
	if saveResult.FailureCode != "" || saveResult.Draft == nil {
		t.Fatalf("save should succeed: %#v", saveResult)
	}
	if saveResult.Draft.DraftVersion != 1 || saveResult.CurrentDraftVersion != 1 {
		t.Fatalf("save should return current version metadata: %#v", saveResult)
	}
	if saveResult.Draft.SampleOrUnsavedDraftStatus != "saved_draft_record" {
		t.Fatalf("save must return saved draft record status: %#v", saveResult.Draft)
	}

	readResult := adapter.ReadWorkflowDraftRecord(actor, ReadWorkflowDraftRecordRequest{
		DraftID: draft.DraftID,
	})
	if readResult.FailureCode != "" || readResult.Draft == nil {
		t.Fatalf("read should succeed: %#v", readResult)
	}
	if readResult.Draft.DraftID != draft.DraftID ||
		readResult.Draft.WorkspaceID != actor.WorkspaceID ||
		readResult.Draft.ApplicationID != actor.ApplicationID {
		t.Fatalf("read returned wrong sanitized draft: %#v", readResult.Draft)
	}

	listResult := adapter.ListWorkflowDraftRecords(actor, ListWorkflowDraftRecordsRequest{})
	if listResult.FailureCode != "" {
		t.Fatalf("list should succeed: %#v", listResult)
	}
	if len(listResult.Summaries) != 1 ||
		listResult.Summaries[0].DraftID != draft.DraftID ||
		listResult.Summaries[0].SampleOrUnsavedDraftStatus != "saved_draft_record" {
		t.Fatalf("list should return sanitized saved draft summary: %#v", listResult.Summaries)
	}
	if executor.saveCalls != 1 || executor.readCalls != 1 || executor.listCalls != 1 {
		t.Fatalf("adapter did not use query executor boundary: %#v", executor)
	}
}

func TestSavedWorkflowDraftRepositoryAdapterFailureSemantics(t *testing.T) {
	t.Run("version conflict preserves current version and no sample fallback", func(t *testing.T) {
		executor := newFakeSavedWorkflowDraftRepositoryQueryExecutor()
		adapter := NewSavedWorkflowDraftRepositoryAdapter(SavedWorkflowDraftRepositoryAdapterConfig{
			QueryExecutor: executor,
		})
		actor := validSavedWorkflowDraftRepositoryActorContext()
		draft := validSavedWorkflowDraftRepositoryDraft()
		first := adapter.SaveWorkflowDraftRecord(actor, SaveWorkflowDraftRecordRequest{Draft: draft})
		if first.FailureCode != "" || first.Draft == nil {
			t.Fatalf("first save should succeed: %#v", first)
		}

		conflict := adapter.SaveWorkflowDraftRecord(actor, SaveWorkflowDraftRecordRequest{
			ExpectedDraftVersion: 0,
			Draft:                draft,
		})
		if conflict.FailureCode != SavedWorkflowDraftFailureVersionConflict ||
			conflict.CurrentDraftVersion != first.Draft.DraftVersion ||
			conflict.Draft != nil {
			t.Fatalf("stale overwrite should fail closed with current version: %#v", conflict)
		}

		missing := adapter.ReadWorkflowDraftRecord(actor, ReadWorkflowDraftRecordRequest{DraftID: "missing_draft"})
		if missing.FailureCode != SavedWorkflowDraftFailureNotFound || missing.Draft != nil {
			t.Fatalf("missing read should not fallback to sample: %#v", missing)
		}
	})

	t.Run("actor context failures map to explicit auth codes", func(t *testing.T) {
		for _, tc := range []struct {
			name        string
			mutate      func(SavedWorkflowDraftRepositoryActorContext) SavedWorkflowDraftRepositoryActorContext
			failureCode SavedWorkflowDraftFailureCode
		}{
			{
				name: "missing identity",
				mutate: func(actor SavedWorkflowDraftRepositoryActorContext) SavedWorkflowDraftRepositoryActorContext {
					actor.ActorSubjectRef = ""
					return actor
				},
				failureCode: SavedWorkflowDraftFailureIdentityContextMissing,
			},
			{
				name: "missing tenant",
				mutate: func(actor SavedWorkflowDraftRepositoryActorContext) SavedWorkflowDraftRepositoryActorContext {
					actor.TenantRef = ""
					return actor
				},
				failureCode: SavedWorkflowDraftFailureTenantBindingMissing,
			},
			{
				name: "missing membership",
				mutate: func(actor SavedWorkflowDraftRepositoryActorContext) SavedWorkflowDraftRepositoryActorContext {
					actor.OwnerSubjectRef = ""
					return actor
				},
				failureCode: SavedWorkflowDraftFailureWorkspaceMembershipDenied,
			},
			{
				name: "missing scope",
				mutate: func(actor SavedWorkflowDraftRepositoryActorContext) SavedWorkflowDraftRepositoryActorContext {
					actor.ScopeGrants = []string{"workflow_drafts:read"}
					return actor
				},
				failureCode: SavedWorkflowDraftFailureScopeGrantMissing,
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				adapter := NewSavedWorkflowDraftRepositoryAdapter(SavedWorkflowDraftRepositoryAdapterConfig{
					QueryExecutor: newFakeSavedWorkflowDraftRepositoryQueryExecutor(),
				})
				result := adapter.SaveWorkflowDraftRecord(
					tc.mutate(validSavedWorkflowDraftRepositoryActorContext()),
					SaveWorkflowDraftRecordRequest{Draft: validSavedWorkflowDraftRepositoryDraft()},
				)
				if result.FailureCode != tc.failureCode || result.Draft != nil {
					t.Fatalf("unexpected auth failure result: %#v", result)
				}
			})
		}
	})

	t.Run("schema preflight fails before query execution", func(t *testing.T) {
		executor := newFakeSavedWorkflowDraftRepositoryQueryExecutor()
		adapter := NewSavedWorkflowDraftRepositoryAdapter(SavedWorkflowDraftRepositoryAdapterConfig{
			QueryExecutor: executor,
			SchemaPreflight: SavedWorkflowDraftRepositorySchemaPreflight{
				StoreSchemaVersion: "wrong_store_schema",
			},
		})
		result := adapter.SaveWorkflowDraftRecord(
			validSavedWorkflowDraftRepositoryActorContext(),
			SaveWorkflowDraftRecordRequest{Draft: validSavedWorkflowDraftRepositoryDraft()},
		)
		if result.FailureCode != SavedWorkflowDraftFailureStoreSchemaVersionMismatch || executor.saveCalls != 0 {
			t.Fatalf("schema mismatch should fail before query execution: result=%#v executor=%#v", result, executor)
		}

		notApplied := NewSavedWorkflowDraftRepositoryAdapter(SavedWorkflowDraftRepositoryAdapterConfig{
			QueryExecutor: executor,
			SchemaPreflight: SavedWorkflowDraftRepositorySchemaPreflight{
				MigrationState: savedWorkflowDraftRepositoryMigrationNotApplied,
			},
		})
		result = notApplied.SaveWorkflowDraftRecord(
			validSavedWorkflowDraftRepositoryActorContext(),
			SaveWorkflowDraftRecordRequest{Draft: validSavedWorkflowDraftRepositoryDraft()},
		)
		if result.FailureCode != SavedWorkflowDraftFailureSchemaMigrationNotApplied {
			t.Fatalf("migration gate should fail closed: %#v", result)
		}
	})

	t.Run("store unavailable and contract mismatch do not expose raw details", func(t *testing.T) {
		nilAdapter := NewSavedWorkflowDraftRepositoryAdapter(SavedWorkflowDraftRepositoryAdapterConfig{})
		nilResult := nilAdapter.ReadWorkflowDraftRecord(
			validSavedWorkflowDraftRepositoryActorContext(),
			ReadWorkflowDraftRecordRequest{DraftID: "draft_radishflow_copilot_saved_v1"},
		)
		if nilResult.FailureCode != SavedWorkflowDraftFailureStoreUnavailable {
			t.Fatalf("nil query executor should fail closed as store unavailable: %#v", nilResult)
		}

		executor := newFakeSavedWorkflowDraftRepositoryQueryExecutor()
		actor := validSavedWorkflowDraftRepositoryActorContext()
		record := savedWorkflowDraftRepositoryRecordFromDraft(
			actor,
			validSavedWorkflowDraftRepositoryDraft(),
			normalizedSavedWorkflowDraftRepositorySchemaPreflight(SavedWorkflowDraftRepositorySchemaPreflight{}),
		)
		record.Draft.ApplicationID = "app_other"
		executor.records[executor.key(actor, record.DraftID)] = record

		adapter := NewSavedWorkflowDraftRepositoryAdapter(SavedWorkflowDraftRepositoryAdapterConfig{
			QueryExecutor: executor,
		})
		result := adapter.ReadWorkflowDraftRecord(actor, ReadWorkflowDraftRecordRequest{DraftID: record.DraftID})
		if result.FailureCode != SavedWorkflowDraftFailureStoreContractMismatch || result.Draft != nil {
			t.Fatalf("stored record mismatch should fail closed without raw detail: %#v", result)
		}
	})
}

func TestSavedWorkflowDraftRepositorySchemaPreflightMatchesManifest(t *testing.T) {
	manifest := loadSavedWorkflowDraftRepositorySchemaManifest(t)
	if manifest.StoreSchemaVersion != savedWorkflowDraftRepositoryStoreSchemaVersion {
		t.Fatalf(
			"repository store schema version constant drifted: manifest=%s const=%s",
			manifest.StoreSchemaVersion,
			savedWorkflowDraftRepositoryStoreSchemaVersion,
		)
	}
}

type fakeSavedWorkflowDraftRepositoryQueryExecutor struct {
	records   map[string]SavedWorkflowDraftRepositoryStoredRecord
	saveCalls int
	readCalls int
	listCalls int
	failure   SavedWorkflowDraftFailureCode
}

func newFakeSavedWorkflowDraftRepositoryQueryExecutor() *fakeSavedWorkflowDraftRepositoryQueryExecutor {
	return &fakeSavedWorkflowDraftRepositoryQueryExecutor{
		records: make(map[string]SavedWorkflowDraftRepositoryStoredRecord),
	}
}

func (executor *fakeSavedWorkflowDraftRepositoryQueryExecutor) SaveWorkflowDraftRecord(
	query savedWorkflowDraftRepositorySaveQuery,
) savedWorkflowDraftRepositoryQuerySaveResult {
	executor.saveCalls++
	if executor.failure != "" {
		return savedWorkflowDraftRepositoryQuerySaveResult{FailureCode: executor.failure}
	}
	key := executor.key(query.ActorContext, query.Record.DraftID)
	existing, found := executor.records[key]
	if found && query.ExpectedDraftVersion != existing.Draft.DraftVersion {
		return savedWorkflowDraftRepositoryQuerySaveResult{
			FailureCode:         SavedWorkflowDraftFailureVersionConflict,
			CurrentDraftVersion: existing.Draft.DraftVersion,
		}
	}
	if !found && query.ExpectedDraftVersion != 0 {
		return savedWorkflowDraftRepositoryQuerySaveResult{FailureCode: SavedWorkflowDraftFailureNotFound}
	}

	record := cloneSavedWorkflowDraftRepositoryStoredRecord(query.Record)
	record.Draft.DraftVersion = 1
	if found {
		record.Draft.DraftVersion = existing.Draft.DraftVersion + 1
	}
	executor.records[key] = cloneSavedWorkflowDraftRepositoryStoredRecord(record)
	return savedWorkflowDraftRepositoryQuerySaveResult{Record: record}
}

func (executor *fakeSavedWorkflowDraftRepositoryQueryExecutor) ReadWorkflowDraftRecord(
	query savedWorkflowDraftRepositoryReadQuery,
) savedWorkflowDraftRepositoryQueryReadResult {
	executor.readCalls++
	if executor.failure != "" {
		return savedWorkflowDraftRepositoryQueryReadResult{FailureCode: executor.failure}
	}
	record, found := executor.records[executor.key(query.ActorContext, query.DraftID)]
	if !found {
		return savedWorkflowDraftRepositoryQueryReadResult{FailureCode: SavedWorkflowDraftFailureNotFound}
	}
	return savedWorkflowDraftRepositoryQueryReadResult{
		Record:              cloneSavedWorkflowDraftRepositoryStoredRecord(record),
		CurrentDraftVersion: record.Draft.DraftVersion,
	}
}

func (executor *fakeSavedWorkflowDraftRepositoryQueryExecutor) ListWorkflowDraftRecords(
	query savedWorkflowDraftRepositoryListQuery,
) savedWorkflowDraftRepositoryQueryListResult {
	executor.listCalls++
	if executor.failure != "" {
		return savedWorkflowDraftRepositoryQueryListResult{FailureCode: executor.failure}
	}
	records := make([]SavedWorkflowDraftRepositoryStoredRecord, 0)
	for _, record := range executor.records {
		if record.TenantRef == query.ActorContext.TenantRef &&
			record.WorkspaceID == query.ActorContext.WorkspaceID &&
			record.ApplicationID == query.ActorContext.ApplicationID &&
			record.OwnerSubjectRef == query.ActorContext.OwnerSubjectRef {
			records = append(records, cloneSavedWorkflowDraftRepositoryStoredRecord(record))
		}
	}
	return savedWorkflowDraftRepositoryQueryListResult{Records: records}
}

func (executor *fakeSavedWorkflowDraftRepositoryQueryExecutor) key(
	actor SavedWorkflowDraftRepositoryActorContext,
	draftID string,
) string {
	return actor.TenantRef + "|" + actor.WorkspaceID + "|" + actor.ApplicationID + "|" + draftID
}

func validSavedWorkflowDraftRepositoryActorContext() SavedWorkflowDraftRepositoryActorContext {
	return SavedWorkflowDraftRepositoryActorContext{
		RequestID:       "req:saved-draft-repository-adapter",
		TenantRef:       "tenant:demo",
		WorkspaceID:     "workspace_demo",
		ApplicationID:   "app_flow_copilot",
		ActorSubjectRef: "subject_platform_ops",
		OwnerSubjectRef: "subject_platform_ops",
		ScopeGrants:     []string{"workflow_drafts:write", "workflow_drafts:read"},
		AuditRef:        "audit:saved-draft-repository-adapter",
	}
}

func validSavedWorkflowDraftRepositoryDraft() SavedWorkflowDraft {
	draft := savedWorkflowDraftFromPayload(validSavedWorkflowDraftPayload())
	draft.SampleOrUnsavedDraftStatus = "saved_draft_record"
	return draft
}

func cloneSavedWorkflowDraftRepositoryStoredRecord(
	record SavedWorkflowDraftRepositoryStoredRecord,
) SavedWorkflowDraftRepositoryStoredRecord {
	record.Draft = cloneSavedWorkflowDraft(record.Draft)
	return record
}

type savedWorkflowDraftRepositorySchemaManifestForTest struct {
	StoreSchemaVersion string `json:"store_schema_version"`
}

func loadSavedWorkflowDraftRepositorySchemaManifest(t *testing.T) savedWorkflowDraftRepositorySchemaManifestForTest {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to locate test file")
	}
	path := filepath.Join(
		filepath.Dir(file),
		"..",
		"..",
		"..",
		"..",
		"services",
		"platform",
		"migrations",
		"workflow_saved_drafts",
		"manifest.json",
	)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read saved draft schema manifest: %v", err)
	}
	var manifest savedWorkflowDraftRepositorySchemaManifestForTest
	if err := json.Unmarshal(content, &manifest); err != nil {
		t.Fatalf("failed to parse saved draft schema manifest: %v", err)
	}
	return manifest
}
