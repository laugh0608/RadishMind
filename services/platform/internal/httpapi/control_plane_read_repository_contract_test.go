package httpapi

import (
	"reflect"
	"testing"
)

func TestControlPlaneReadRepositoryContextTypeBoundary(t *testing.T) {
	contextType := reflect.TypeOf(ReadRepositoryContext{})
	expectedFields := []string{
		"RequestID",
		"TenantRef",
		"SubjectRef",
		"ScopeGrants",
		"AuditRef",
		"IssuerRef",
		"SessionRef",
	}
	for _, field := range expectedFields {
		if _, found := contextType.FieldByName(field); !found {
			t.Fatalf("ReadRepositoryContext missing field %s", field)
		}
	}
	for _, forbidden := range []string{
		"RawAuthorizationHeader",
		"CookieValue",
		"APIKeyValue",
		"APIKeyHash",
		"QueryStringTenantOverride",
		"GlobalCurrentTenant",
		"ProviderRuntime",
		"WorkflowExecutor",
		"ConfirmationDecision",
		"BusinessWritebackPayload",
	} {
		if _, found := contextType.FieldByName(forbidden); found {
			t.Fatalf("ReadRepositoryContext must not include %s", forbidden)
		}
	}
}

func TestControlPlaneReadRepositoryRouteTypeContracts(t *testing.T) {
	contracts := controlPlaneReadRepositoryRouteTypeContracts()
	if len(contracts) != 7 {
		t.Fatalf("expected seven route type contracts, got %d", len(contracts))
	}
	byRoute := map[string]ReadRepositoryRouteTypeContract{}
	for _, contract := range contracts {
		if contract.RouteID == "" || contract.Operation == "" || contract.RequestType == "" || contract.ResultType == "" {
			t.Fatalf("route type contract must define identifiers: %#v", contract)
		}
		if _, exists := byRoute[contract.RouteID]; exists {
			t.Fatalf("duplicate route type contract for %s", contract.RouteID)
		}
		byRoute[contract.RouteID] = contract
	}
	assertRouteTypeContract(t, byRoute, "tenant-summary-route", "ReadTenantSummary", "ReadTenantSummaryRequest", "ReadTenantSummaryResult", nil, nil)
	assertRouteTypeContract(t, byRoute, "application-summary-list-route", "ListApplicationSummaries", "ListApplicationSummariesRequest", "ListApplicationSummariesResult", []string{"status", "kind"}, []string{"updated_at_desc"})
	assertRouteTypeContract(t, byRoute, "api-key-summary-list-route", "ListAPIKeySummaries", "ListAPIKeySummariesRequest", "ListAPIKeySummariesResult", []string{"state", "owner_subject_ref"}, []string{"created_at_desc"})
	assertRouteTypeContract(t, byRoute, "quota-summary-route", "ReadQuotaSummary", "ReadQuotaSummaryRequest", "ReadQuotaSummaryResult", []string{"period"}, nil)
	assertRouteTypeContract(t, byRoute, "workflow-definition-summary-list-route", "ListWorkflowDefinitionSummaries", "ListWorkflowDefinitionSummariesRequest", "ListWorkflowDefinitionSummariesResult", []string{"application_ref", "definition_status", "risk_level"}, []string{"updated_at_desc"})
	assertRouteTypeContract(t, byRoute, "run-record-summary-list-route", "ListRunRecordSummaries", "ListRunRecordSummariesRequest", "ListRunRecordSummariesResult", []string{"application_ref", "workflow_definition_ref", "status", "failure_code"}, []string{"started_at_desc"})
	assertRouteTypeContract(t, byRoute, "audit-summary-list-route", "ListAuditSummaries", "ListAuditSummariesRequest", "ListAuditSummariesResult", []string{"actor_subject_ref", "event_kind", "resource_kind", "decision", "failure_code"}, []string{"recorded_at_desc"})
}

func TestControlPlaneReadRepositoryFailureCodes(t *testing.T) {
	expected := map[ReadRepositoryFailureCode]bool{
		ReadRepositoryFailureTenantBindingMissing:      false,
		ReadRepositoryFailureScopeDenied:               false,
		ReadRepositoryFailureInvalidFilter:             false,
		ReadRepositoryFailureStoreUnavailable:          false,
		ReadRepositoryFailureContractMismatch:          false,
		ReadRepositoryFailureDatabaseReadDisabled:      false,
		ReadRepositoryFailureAuthContextMismatch:       false,
		ReadRepositoryFailureSchemaMigrationNotApplied: false,
		ReadRepositoryFailureSchemaVersionMismatch:     false,
	}
	for code := range expected {
		if string(code) == "" {
			t.Fatalf("failure code must not be empty")
		}
		expected[code] = true
	}
	for code, seen := range expected {
		if !seen {
			t.Fatalf("failure code was not observed: %s", code)
		}
	}
}

func assertRouteTypeContract(
	t *testing.T,
	contracts map[string]ReadRepositoryRouteTypeContract,
	routeID string,
	operation string,
	requestType string,
	resultType string,
	filters []string,
	sort []string,
) {
	t.Helper()
	contract, found := contracts[routeID]
	if !found {
		t.Fatalf("missing route type contract %s", routeID)
	}
	if contract.Operation != operation || contract.RequestType != requestType || contract.ResultType != resultType {
		t.Fatalf("unexpected route contract for %s: %#v", routeID, contract)
	}
	if !reflect.DeepEqual(contract.AllowedFilters, filters) {
		t.Fatalf("unexpected filters for %s: %#v", routeID, contract.AllowedFilters)
	}
	if !reflect.DeepEqual(contract.AllowedSort, sort) {
		t.Fatalf("unexpected sort for %s: %#v", routeID, contract.AllowedSort)
	}
}
