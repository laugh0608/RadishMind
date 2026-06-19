package httpapi

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"testing"
)

func TestSavedWorkflowDraftRepositoryContractSmokeRunnerConsumesOperationContracts(t *testing.T) {
	report := RunSavedWorkflowDraftRepositoryContractSmoke(validSavedWorkflowDraftRepositorySmokeContext())
	if !report.Summary.Passed {
		t.Fatalf("expected saved draft repository contract smoke report to pass: %#v", report)
	}
	if report.Summary.TotalOperations != 3 || report.Summary.PassedOperations != 3 {
		t.Fatalf("expected all three operations to pass, got %#v", report.Summary)
	}
	if len(report.ContractMismatchReport) != 0 {
		t.Fatalf("unexpected contract mismatches: %#v", report.ContractMismatchReport)
	}
	if len(report.FallbackReport) != 0 {
		t.Fatalf("unexpected fallback report entries: %#v", report.FallbackReport)
	}
	assertNoSavedWorkflowDraftRepositorySmokeSideEffects(t, report.SideEffectReport)

	resultsByOperation := map[string]SavedWorkflowDraftRepositoryContractSmokeOperationResult{}
	for _, result := range report.OperationResults {
		resultsByOperation[result.Operation] = result
	}
	for _, contract := range savedWorkflowDraftRepositoryOperationContracts() {
		result, found := resultsByOperation[contract.Operation]
		if !found {
			t.Fatalf("missing operation result for %s", contract.Operation)
		}
		if result.RequestType != contract.RequestType || result.ResultType != contract.ResultType {
			t.Fatalf("operation result did not consume type contract for %s: %#v", contract.Operation, result)
		}
	}
}

func TestSavedWorkflowDraftRepositoryContractSmokeCasesMatchFixture(t *testing.T) {
	fixture := loadSavedWorkflowDraftRepositoryContractSmokeFixture(t)
	casesByOperation := map[string]SavedWorkflowDraftRepositoryContractSmokeCase{}
	for _, smokeCase := range savedWorkflowDraftRepositoryContractSmokeCases() {
		casesByOperation[smokeCase.Operation] = smokeCase
	}
	if len(casesByOperation) != len(fixture.OperationSmokeMatrix) {
		t.Fatalf("operation smoke case count drifted: got %d fixture %d", len(casesByOperation), len(fixture.OperationSmokeMatrix))
	}

	for _, row := range fixture.OperationSmokeMatrix {
		smokeCase, found := casesByOperation[row.Operation]
		if !found {
			t.Fatalf("missing static smoke case for %s", row.Operation)
		}
		if smokeCase.RequiredScope != row.RequiredScope ||
			smokeCase.RequestType != row.RequestType ||
			smokeCase.ResultType != row.ResultType ||
			smokeCase.ExpectedSuccessProjection != row.ExpectedSuccessProjection {
			t.Fatalf("static smoke case metadata drifted for %s: %#v", row.Operation, smokeCase)
		}
		if !sameSavedWorkflowDraftRepositorySmokeFailures(smokeCase.RequiredFailureCodes, row.RequiredFailureCodes) {
			t.Fatalf("failure codes drifted for %s: got %#v fixture %#v", row.Operation, smokeCase.RequiredFailureCodes, row.RequiredFailureCodes)
		}
		if smokeCase.RepositoryContractSmokeRunnerImplemented != row.RepositoryContractSmokeRunnerImplemented ||
			smokeCase.RepositoryAdapterDependencySatisfied != row.RepositoryAdapterDependencySatisfied ||
			smokeCase.FallbackToMemoryDevStoreAllowed != row.FallbackToMemoryDevStoreAllowed ||
			smokeCase.FallbackToSampleAllowed != row.FallbackToSampleAllowed ||
			smokeCase.FallbackToFixtureAllowed != row.FallbackToFixtureAllowed ||
			smokeCase.SideEffectAllowed != row.SideEffectAllowed {
			t.Fatalf("fallback/side-effect policy drifted for %s", row.Operation)
		}
	}
}

func TestSavedWorkflowDraftRepositoryContractSmokeRunnerRejectsInvalidContext(t *testing.T) {
	report := RunSavedWorkflowDraftRepositoryContractSmoke(SavedWorkflowDraftRepositoryActorContext{})
	if report.Summary.Passed {
		t.Fatalf("invalid context must fail the static smoke report")
	}
	for _, result := range report.OperationResults {
		if result.FailureCode != SavedWorkflowDraftRepositorySmokeFailureAuthContextMismatch {
			t.Fatalf("expected auth context mismatch, got %#v", result)
		}
	}
	assertNoSavedWorkflowDraftRepositorySmokeSideEffects(t, report.SideEffectReport)
}

func TestSavedWorkflowDraftRepositoryContractSmokeRunnerRejectsMissingScopeGrant(t *testing.T) {
	context := validSavedWorkflowDraftRepositorySmokeContext()
	context.ScopeGrants = []string{"workflow_drafts:read"}

	report := RunSavedWorkflowDraftRepositoryContractSmoke(context)
	if report.Summary.Passed {
		t.Fatalf("missing write scope must fail the static smoke report")
	}
	var saveResult SavedWorkflowDraftRepositoryContractSmokeOperationResult
	for _, result := range report.OperationResults {
		if result.Operation == "SaveWorkflowDraftRecord" {
			saveResult = result
			break
		}
	}
	if saveResult.Operation == "" || saveResult.FailureCode != SavedWorkflowDraftRepositorySmokeFailureScopeDenied {
		t.Fatalf("expected save operation scope denied, got %#v", saveResult)
	}
	assertNoSavedWorkflowDraftRepositorySmokeSideEffects(t, report.SideEffectReport)
}

func validSavedWorkflowDraftRepositorySmokeContext() SavedWorkflowDraftRepositoryActorContext {
	return SavedWorkflowDraftRepositoryActorContext{
		RequestID:       "req:saved-draft-contract-smoke",
		TenantRef:       "tenant:demo",
		WorkspaceID:     "workspace:demo",
		ApplicationID:   "app:demo",
		ActorSubjectRef: "subject:demo",
		OwnerSubjectRef: "subject:demo",
		ScopeGrants:     []string{"workflow_drafts:write", "workflow_drafts:read"},
		AuditRef:        "audit:saved-draft-contract-smoke",
	}
}

func assertNoSavedWorkflowDraftRepositorySmokeSideEffects(
	t *testing.T,
	report SavedWorkflowDraftRepositoryContractSmokeSideEffectReport,
) {
	t.Helper()
	if report.RepositoryWriteCount != 0 ||
		report.ExecutorCallCount != 0 ||
		report.ConfirmationCallCount != 0 ||
		report.BusinessWritebackCount != 0 ||
		report.ReplayCallCount != 0 {
		t.Fatalf("side-effect counters must remain zero: %#v", report)
	}
	expected := []string{
		"business_writeback",
		"confirmation_decision",
		"database_write",
		"replay_execution",
		"workflow_execution",
	}
	got := append([]string{}, report.ForbiddenSideEffects...)
	sort.Strings(got)
	if len(got) != len(expected) {
		t.Fatalf("forbidden side effects drifted: %#v", report.ForbiddenSideEffects)
	}
	for index := range expected {
		if got[index] != expected[index] {
			t.Fatalf("forbidden side effects drifted: %#v", report.ForbiddenSideEffects)
		}
	}
}

type savedWorkflowDraftRepositoryContractSmokeFixture struct {
	OperationSmokeMatrix []savedWorkflowDraftRepositoryContractSmokeFixtureOperation `json:"operation_smoke_matrix"`
}

type savedWorkflowDraftRepositoryContractSmokeFixtureOperation struct {
	Operation                                string   `json:"operation"`
	RequiredScope                            string   `json:"required_scope"`
	RequestType                              string   `json:"request_type"`
	ResultType                               string   `json:"result_type"`
	ExpectedSuccessProjection                string   `json:"expected_success_projection"`
	RequiredFailureCodes                     []string `json:"required_failure_codes"`
	RepositoryContractSmokeRunnerImplemented bool     `json:"repository_contract_smoke_runner_implemented"`
	RepositoryAdapterDependencySatisfied     bool     `json:"repository_adapter_dependency_satisfied"`
	FallbackToMemoryDevStoreAllowed          bool     `json:"fallback_to_memory_dev_store_allowed"`
	FallbackToSampleAllowed                  bool     `json:"fallback_to_sample_allowed"`
	FallbackToFixtureAllowed                 bool     `json:"fallback_to_fixture_allowed"`
	SideEffectAllowed                        bool     `json:"side_effect_allowed"`
}

func loadSavedWorkflowDraftRepositoryContractSmokeFixture(t *testing.T) savedWorkflowDraftRepositoryContractSmokeFixture {
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
		"workflow-saved-draft-repository-contract-smoke-v1.json",
	)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read saved draft repository contract smoke fixture: %v", err)
	}
	var fixture savedWorkflowDraftRepositoryContractSmokeFixture
	if err := json.Unmarshal(content, &fixture); err != nil {
		t.Fatalf("failed to parse saved draft repository contract smoke fixture: %v", err)
	}
	return fixture
}

func sameSavedWorkflowDraftRepositorySmokeFailures(
	left []SavedWorkflowDraftRepositoryContractSmokeFailureCode,
	right []string,
) bool {
	if len(left) != len(right) {
		return false
	}
	leftValues := make([]string, 0, len(left))
	for _, value := range left {
		leftValues = append(leftValues, string(value))
	}
	rightValues := append([]string{}, right...)
	sort.Strings(leftValues)
	sort.Strings(rightValues)
	for index := range leftValues {
		if leftValues[index] != rightValues[index] {
			return false
		}
	}
	return true
}
