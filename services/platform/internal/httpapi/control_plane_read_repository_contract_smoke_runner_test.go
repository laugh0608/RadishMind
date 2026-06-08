package httpapi

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"testing"
)

func TestControlPlaneReadRepositoryContractSmokeRunnerConsumesTypeContracts(t *testing.T) {
	report := RunControlPlaneReadRepositoryContractSmoke(validReadRepositorySmokeContext())
	if !report.Summary.Passed {
		t.Fatalf("expected static contract smoke report to pass: %#v", report)
	}
	if report.Summary.TotalRoutes != 7 || report.Summary.PassedRoutes != 7 {
		t.Fatalf("expected all seven routes to pass, got %#v", report.Summary)
	}
	if len(report.ContractMismatchReport) != 0 {
		t.Fatalf("unexpected contract mismatches: %#v", report.ContractMismatchReport)
	}
	assertNoReadRepositorySmokeSideEffects(t, report.SideEffectReport)

	resultsByRoute := map[string]ReadRepositoryContractSmokeRouteResult{}
	for _, result := range report.RouteResults {
		resultsByRoute[result.RouteID] = result
	}
	for _, contract := range controlPlaneReadRepositoryRouteTypeContracts() {
		result, found := resultsByRoute[contract.RouteID]
		if !found {
			t.Fatalf("missing route result for %s", contract.RouteID)
		}
		if result.Operation != contract.Operation || result.RequestType != contract.RequestType || result.ResultType != contract.ResultType {
			t.Fatalf("route result did not consume type contract for %s: %#v", contract.RouteID, result)
		}
	}
}

func TestControlPlaneReadRepositoryContractSmokeCasesMatchFixture(t *testing.T) {
	fixture := loadRepositoryContractSmokeFixture(t)
	casesByRoute := map[string]ReadRepositoryContractSmokeCase{}
	for _, smokeCase := range controlPlaneReadRepositoryContractSmokeCases() {
		casesByRoute[smokeCase.RouteID] = smokeCase
	}
	if len(casesByRoute) != len(fixture.RouteSmokeMatrix) {
		t.Fatalf("route smoke case count drifted: got %d fixture %d", len(casesByRoute), len(fixture.RouteSmokeMatrix))
	}

	for _, row := range fixture.RouteSmokeMatrix {
		smokeCase, found := casesByRoute[row.RouteID]
		if !found {
			t.Fatalf("missing static smoke case for %s", row.RouteID)
		}
		if smokeCase.Method != row.Method ||
			smokeCase.Path != row.Path ||
			smokeCase.ReadModel != row.ReadModel ||
			smokeCase.RequiredScope != row.RequiredScope ||
			smokeCase.Operation != row.Operation ||
			smokeCase.ReadMode != row.ReadMode {
			t.Fatalf("static smoke case metadata drifted for %s: %#v", row.RouteID, smokeCase)
		}
		if !sameOptionalInt(smokeCase.Request.Limit, row.Request.Limit) {
			t.Fatalf("limit drifted for %s", row.RouteID)
		}
		if !sameOptionalString(smokeCase.Request.Cursor, row.Request.Cursor) {
			t.Fatalf("cursor drifted for %s", row.RouteID)
		}
		if !sameOptionalSort(smokeCase.Request.Sort, row.Request.Sort) {
			t.Fatalf("sort drifted for %s", row.RouteID)
		}
		if string(smokeCase.Request.Projection) != row.Request.Projection {
			t.Fatalf("projection drifted for %s", row.RouteID)
		}
		if !sameNullableStringMap(smokeCase.Request.Filters, row.Request.Filters) {
			t.Fatalf("filters drifted for %s: got %#v fixture %#v", row.RouteID, smokeCase.Request.Filters, row.Request.Filters)
		}
		if smokeCase.DatabaseModeFailureCode != ReadRepositoryFailureCode(row.DatabaseModeFailureCode) {
			t.Fatalf("database mode failure drifted for %s", row.RouteID)
		}
		if smokeCase.FakeFallbackAllowed != row.FakeFallbackAllowed || smokeCase.SideEffectAllowed != row.SideEffectAllowed {
			t.Fatalf("fallback/side-effect policy drifted for %s", row.RouteID)
		}
	}
}

func TestControlPlaneReadRepositoryContractSmokeRunnerRejectsInvalidContext(t *testing.T) {
	report := RunControlPlaneReadRepositoryContractSmoke(ReadRepositoryContext{})
	if report.Summary.Passed {
		t.Fatalf("invalid context must fail the static smoke report")
	}
	for _, result := range report.RouteResults {
		if result.FailureCode != ReadRepositoryFailureAuthContextMismatch {
			t.Fatalf("expected auth context mismatch, got %#v", result)
		}
	}
	assertNoReadRepositorySmokeSideEffects(t, report.SideEffectReport)
}

func TestControlPlaneReadRepositoryContractSmokeRunnerRejectsUnknownFilter(t *testing.T) {
	runner := NewReadRepositoryContractSmokeRunner()
	runner.smokeCases = append([]ReadRepositoryContractSmokeCase{}, runner.smokeCases...)
	runner.smokeCases[0].Request.Filters = map[string]*string{"unknown_filter": nil}

	report := runner.Run(validReadRepositorySmokeContext())
	if report.Summary.Passed {
		t.Fatalf("unknown filter must fail the static smoke report")
	}
	first := report.RouteResults[0]
	if first.RouteID != "tenant-summary-route" || first.FailureCode != ReadRepositoryFailureInvalidFilter {
		t.Fatalf("expected tenant-summary invalid_filter, got %#v", first)
	}
	assertNoReadRepositorySmokeSideEffects(t, report.SideEffectReport)
}

func validReadRepositorySmokeContext() ReadRepositoryContext {
	return ReadRepositoryContext{
		RequestID:   "req:contract-smoke",
		TenantRef:   "tenant:demo",
		SubjectRef:  "subject:demo",
		ScopeGrants: []string{"tenant:read", "applications:read", "api_keys:read", "usage:read", "runs:read", "audit:read"},
		AuditRef:    "audit:contract-smoke",
		IssuerRef:   "issuer:radishmind-dev",
		SessionRef:  "session:contract-smoke",
	}
}

func assertNoReadRepositorySmokeSideEffects(t *testing.T, report ReadRepositoryContractSmokeSideEffectReport) {
	t.Helper()
	if report.RepositoryWriteCount != 0 ||
		report.ExecutorCallCount != 0 ||
		report.ConfirmationCallCount != 0 ||
		report.BusinessWritebackCount != 0 ||
		report.ReplayCallCount != 0 {
		t.Fatalf("side-effect counters must remain zero: %#v", report)
	}
	expected := []string{
		"api_key_issue",
		"business_writeback",
		"confirmation_decision",
		"database_write",
		"quota_mutation",
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

type repositoryContractSmokeFixture struct {
	RouteSmokeMatrix []repositoryContractSmokeFixtureRoute `json:"route_smoke_matrix"`
}

type repositoryContractSmokeFixtureRoute struct {
	RouteID                 string                                `json:"route_id"`
	Method                  string                                `json:"method"`
	Path                    string                                `json:"path"`
	ReadModel               string                                `json:"read_model"`
	RequiredScope           string                                `json:"required_scope"`
	Operation               string                                `json:"operation"`
	ReadMode                string                                `json:"read_mode"`
	Request                 repositoryContractSmokeFixtureRequest `json:"request"`
	DatabaseModeFailureCode string                                `json:"database_mode_failure_code"`
	FakeFallbackAllowed     bool                                  `json:"fake_fallback_allowed"`
	SideEffectAllowed       bool                                  `json:"side_effect_allowed"`
}

type repositoryContractSmokeFixtureRequest struct {
	Limit      *int               `json:"limit"`
	Cursor     *string            `json:"cursor"`
	Filters    map[string]*string `json:"filters"`
	Sort       *string            `json:"sort"`
	Projection string             `json:"projection"`
}

func loadRepositoryContractSmokeFixture(t *testing.T) repositoryContractSmokeFixture {
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
		"control-plane-read-repository-contract-smoke-v1.json",
	)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read repository contract smoke fixture: %v", err)
	}
	var fixture repositoryContractSmokeFixture
	if err := json.Unmarshal(content, &fixture); err != nil {
		t.Fatalf("failed to parse repository contract smoke fixture: %v", err)
	}
	return fixture
}

func sameOptionalInt(left *int, right *int) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func sameOptionalString(left *string, right *string) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func sameOptionalSort(left *ReadRepositorySort, right *string) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return string(*left) == *right
}

func sameNullableStringMap(left map[string]*string, right map[string]*string) bool {
	if len(left) != len(right) {
		return false
	}
	for key, leftValue := range left {
		rightValue, found := right[key]
		if !found || !sameOptionalString(leftValue, rightValue) {
			return false
		}
	}
	return true
}
