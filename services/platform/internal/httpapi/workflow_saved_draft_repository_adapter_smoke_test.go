package httpapi

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestSavedWorkflowDraftRepositoryAdapterSmokeExecutionConsumesStaticCases(t *testing.T) {
	staticReport := RunSavedWorkflowDraftRepositoryContractSmoke(validSavedWorkflowDraftRepositoryActorContext())
	if !staticReport.Summary.Passed {
		t.Fatalf("static repository contract smoke must pass before adapter smoke: %#v", staticReport)
	}

	executor := newFakeSavedWorkflowDraftRepositoryQueryExecutor()
	adapter := NewSavedWorkflowDraftRepositoryAdapter(SavedWorkflowDraftRepositoryAdapterConfig{
		QueryExecutor: executor,
	})
	actor := validSavedWorkflowDraftRepositoryActorContext()
	draft := validSavedWorkflowDraftRepositoryDraft()

	resultsByOperation := make(map[string]SavedWorkflowDraftFailureCode)
	for _, smokeCase := range savedWorkflowDraftRepositoryContractSmokeCases() {
		if !smokeCase.UsesRepositoryOperationContract ||
			!smokeCase.UsesContractSmokeFixture ||
			!smokeCase.UsesAuthContextContract ||
			!smokeCase.UsesSchemaArtifactGate ||
			!smokeCase.UsesSelectorSmokeGate {
			t.Fatalf("smoke case does not consume required contracts: %#v", smokeCase)
		}
		switch smokeCase.Operation {
		case "SaveWorkflowDraftRecord":
			result := adapter.SaveWorkflowDraftRecord(context.Background(), actor, SaveWorkflowDraftRecordRequest{
				ExpectedDraftVersion: 0,
				Draft:                draft,
			})
			resultsByOperation[smokeCase.Operation] = result.FailureCode
			if result.FailureCode != "" || result.Draft == nil || result.CurrentDraftVersion != 1 {
				t.Fatalf("adapter smoke save failed: %#v", result)
			}
		case "ReadWorkflowDraftRecord":
			result := adapter.ReadWorkflowDraftRecord(context.Background(), actor, ReadWorkflowDraftRecordRequest{
				DraftID: draft.DraftID,
			})
			resultsByOperation[smokeCase.Operation] = result.FailureCode
			if result.FailureCode != "" || result.Draft == nil || result.Draft.DraftID != draft.DraftID {
				t.Fatalf("adapter smoke read failed: %#v", result)
			}
		case "ListWorkflowDraftRecords":
			result := adapter.ListWorkflowDraftRecords(context.Background(), actor, ListWorkflowDraftRecordsRequest{})
			resultsByOperation[smokeCase.Operation] = result.FailureCode
			if result.FailureCode != "" || len(result.Summaries) != 1 ||
				result.Summaries[0].DraftID != draft.DraftID {
				t.Fatalf("adapter smoke list failed: %#v", result)
			}
		default:
			t.Fatalf("unknown adapter smoke operation: %s", smokeCase.Operation)
		}
	}

	if len(resultsByOperation) != 3 ||
		resultsByOperation["SaveWorkflowDraftRecord"] != "" ||
		resultsByOperation["ReadWorkflowDraftRecord"] != "" ||
		resultsByOperation["ListWorkflowDraftRecords"] != "" {
		t.Fatalf("adapter smoke operation results drifted: %#v", resultsByOperation)
	}
	if executor.saveCalls != 1 || executor.readCalls != 1 || executor.listCalls != 1 {
		t.Fatalf("adapter smoke must use injected query executor exactly once per operation: %#v", executor)
	}
}

func TestSavedWorkflowDraftRepositoryAdapterSmokeExecutionFailureMapping(t *testing.T) {
	for _, tc := range []struct {
		name        string
		run         func() SavedWorkflowDraftFailureCode
		failureCode SavedWorkflowDraftFailureCode
	}{
		{
			name: "version conflict",
			run: func() SavedWorkflowDraftFailureCode {
				executor := newFakeSavedWorkflowDraftRepositoryQueryExecutor()
				adapter := NewSavedWorkflowDraftRepositoryAdapter(SavedWorkflowDraftRepositoryAdapterConfig{
					QueryExecutor: executor,
				})
				actor := validSavedWorkflowDraftRepositoryActorContext()
				draft := validSavedWorkflowDraftRepositoryDraft()
				first := adapter.SaveWorkflowDraftRecord(context.Background(), actor, SaveWorkflowDraftRecordRequest{Draft: draft})
				if first.FailureCode != "" {
					t.Fatalf("seed save failed: %#v", first)
				}
				return adapter.SaveWorkflowDraftRecord(context.Background(), actor, SaveWorkflowDraftRecordRequest{
					ExpectedDraftVersion: 0,
					Draft:                draft,
				}).FailureCode
			},
			failureCode: SavedWorkflowDraftFailureVersionConflict,
		},
		{
			name: "not found",
			run: func() SavedWorkflowDraftFailureCode {
				adapter := NewSavedWorkflowDraftRepositoryAdapter(SavedWorkflowDraftRepositoryAdapterConfig{
					QueryExecutor: newFakeSavedWorkflowDraftRepositoryQueryExecutor(),
				})
				return adapter.ReadWorkflowDraftRecord(
					context.Background(),
					validSavedWorkflowDraftRepositoryActorContext(),
					ReadWorkflowDraftRecordRequest{DraftID: "missing_draft"},
				).FailureCode
			},
			failureCode: SavedWorkflowDraftFailureNotFound,
		},
		{
			name: "store unavailable",
			run: func() SavedWorkflowDraftFailureCode {
				adapter := NewSavedWorkflowDraftRepositoryAdapter(SavedWorkflowDraftRepositoryAdapterConfig{})
				return adapter.ListWorkflowDraftRecords(
					context.Background(),
					validSavedWorkflowDraftRepositoryActorContext(),
					ListWorkflowDraftRecordsRequest{},
				).FailureCode
			},
			failureCode: SavedWorkflowDraftFailureStoreUnavailable,
		},
		{
			name: "contract mismatch",
			run: func() SavedWorkflowDraftFailureCode {
				executor := newFakeSavedWorkflowDraftRepositoryQueryExecutor()
				actor := validSavedWorkflowDraftRepositoryActorContext()
				record := savedWorkflowDraftRepositoryRecordFromDraft(
					actor,
					validSavedWorkflowDraftRepositoryDraft(),
					normalizedSavedWorkflowDraftRepositorySchemaPreflight(SavedWorkflowDraftRepositorySchemaPreflight{}),
				)
				record.Draft.WorkspaceID = "workspace_other"
				executor.records[executor.key(actor, record.DraftID)] = record
				adapter := NewSavedWorkflowDraftRepositoryAdapter(SavedWorkflowDraftRepositoryAdapterConfig{
					QueryExecutor: executor,
				})
				return adapter.ReadWorkflowDraftRecord(context.Background(), actor, ReadWorkflowDraftRecordRequest{
					DraftID: record.DraftID,
				}).FailureCode
			},
			failureCode: SavedWorkflowDraftFailureStoreContractMismatch,
		},
		{
			name: "schema migration not applied",
			run: func() SavedWorkflowDraftFailureCode {
				adapter := NewSavedWorkflowDraftRepositoryAdapter(SavedWorkflowDraftRepositoryAdapterConfig{
					QueryExecutor: newFakeSavedWorkflowDraftRepositoryQueryExecutor(),
					SchemaPreflight: SavedWorkflowDraftRepositorySchemaPreflight{
						MigrationState: savedWorkflowDraftRepositoryMigrationNotApplied,
					},
				})
				return adapter.SaveWorkflowDraftRecord(
					context.Background(),
					validSavedWorkflowDraftRepositoryActorContext(),
					SaveWorkflowDraftRecordRequest{Draft: validSavedWorkflowDraftRepositoryDraft()},
				).FailureCode
			},
			failureCode: SavedWorkflowDraftFailureSchemaMigrationNotApplied,
		},
		{
			name: "schema version mismatch",
			run: func() SavedWorkflowDraftFailureCode {
				adapter := NewSavedWorkflowDraftRepositoryAdapter(SavedWorkflowDraftRepositoryAdapterConfig{
					QueryExecutor: newFakeSavedWorkflowDraftRepositoryQueryExecutor(),
					SchemaPreflight: SavedWorkflowDraftRepositorySchemaPreflight{
						StoreSchemaVersion: "wrong_store_schema",
					},
				})
				return adapter.ReadWorkflowDraftRecord(
					context.Background(),
					validSavedWorkflowDraftRepositoryActorContext(),
					ReadWorkflowDraftRecordRequest{DraftID: validSavedWorkflowDraftRepositoryDraft().DraftID},
				).FailureCode
			},
			failureCode: SavedWorkflowDraftFailureStoreSchemaVersionMismatch,
		},
		{
			name: "store migration unavailable",
			run: func() SavedWorkflowDraftFailureCode {
				adapter := NewSavedWorkflowDraftRepositoryAdapter(SavedWorkflowDraftRepositoryAdapterConfig{
					QueryExecutor: newFakeSavedWorkflowDraftRepositoryQueryExecutor(),
					SchemaPreflight: SavedWorkflowDraftRepositorySchemaPreflight{
						MigrationState: savedWorkflowDraftRepositoryMigrationUnavailable,
					},
				})
				return adapter.ListWorkflowDraftRecords(
					context.Background(),
					validSavedWorkflowDraftRepositoryActorContext(),
					ListWorkflowDraftRecordsRequest{},
				).FailureCode
			},
			failureCode: SavedWorkflowDraftFailureStoreMigrationUnavailable,
		},
		{
			name: "auth context mismatch",
			run: func() SavedWorkflowDraftFailureCode {
				actor := validSavedWorkflowDraftRepositoryActorContext()
				actor.AuditRef = ""
				adapter := NewSavedWorkflowDraftRepositoryAdapter(SavedWorkflowDraftRepositoryAdapterConfig{
					QueryExecutor: newFakeSavedWorkflowDraftRepositoryQueryExecutor(),
				})
				return adapter.ReadWorkflowDraftRecord(context.Background(), actor, ReadWorkflowDraftRecordRequest{
					DraftID: validSavedWorkflowDraftRepositoryDraft().DraftID,
				}).FailureCode
			},
			failureCode: SavedWorkflowDraftFailureAuthContextMismatch,
		},
		{
			name: "scope grant missing",
			run: func() SavedWorkflowDraftFailureCode {
				actor := validSavedWorkflowDraftRepositoryActorContext()
				actor.ScopeGrants = []string{"workflow_drafts:read"}
				adapter := NewSavedWorkflowDraftRepositoryAdapter(SavedWorkflowDraftRepositoryAdapterConfig{
					QueryExecutor: newFakeSavedWorkflowDraftRepositoryQueryExecutor(),
				})
				return adapter.SaveWorkflowDraftRecord(context.Background(), actor, SaveWorkflowDraftRecordRequest{
					Draft: validSavedWorkflowDraftRepositoryDraft(),
				}).FailureCode
			},
			failureCode: SavedWorkflowDraftFailureScopeGrantMissing,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.run(); got != tc.failureCode {
				t.Fatalf("expected %s, got %s", tc.failureCode, got)
			}
		})
	}
}

func TestSavedWorkflowDraftRepositoryAdapterSmokeFixtureMatchesGoTest(t *testing.T) {
	fixture := loadSavedWorkflowDraftAdapterSmokeFixture(t)
	if fixture.Slice.Status != "draft_adapter_smoke_executed" {
		t.Fatalf("adapter smoke fixture status drifted: %s", fixture.Slice.Status)
	}
	if fixture.ExecutionBoundary.AdapterSmokeTestFile != "services/platform/internal/httpapi/workflow_saved_draft_repository_adapter_smoke_test.go" {
		t.Fatalf("adapter smoke test file drifted: %s", fixture.ExecutionBoundary.AdapterSmokeTestFile)
	}

	operations := map[string]bool{}
	for _, row := range fixture.OperationSmokeMatrix {
		if !row.StaticContractCaseConsumed || !row.AdapterMethodCalled || row.FallbackToSampleAllowed {
			t.Fatalf("operation smoke row has invalid execution boundary: %#v", row)
		}
		operations[row.Operation] = true
	}
	for _, operation := range []string{
		"SaveWorkflowDraftRecord",
		"ReadWorkflowDraftRecord",
		"ListWorkflowDraftRecords",
	} {
		if !operations[operation] {
			t.Fatalf("fixture missing operation smoke row: %s", operation)
		}
	}
}

type savedWorkflowDraftAdapterSmokeFixtureForTest struct {
	Slice struct {
		Status string `json:"status"`
	} `json:"slice"`
	ExecutionBoundary struct {
		AdapterSmokeTestFile string `json:"adapter_smoke_test_file"`
	} `json:"execution_boundary"`
	OperationSmokeMatrix []struct {
		Operation                  string `json:"operation"`
		StaticContractCaseConsumed bool   `json:"static_contract_case_consumed"`
		AdapterMethodCalled        bool   `json:"adapter_method_called"`
		FallbackToSampleAllowed    bool   `json:"fallback_to_sample_allowed"`
	} `json:"operation_smoke_matrix"`
}

func loadSavedWorkflowDraftAdapterSmokeFixture(
	t *testing.T,
) savedWorkflowDraftAdapterSmokeFixtureForTest {
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
		"scripts",
		"checks",
		"fixtures",
		"workflow-saved-draft-adapter-smoke-v1.json",
	)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read adapter smoke fixture: %v", err)
	}
	var fixture savedWorkflowDraftAdapterSmokeFixtureForTest
	if err := json.Unmarshal(content, &fixture); err != nil {
		t.Fatalf("failed to parse adapter smoke fixture: %v", err)
	}
	return fixture
}
