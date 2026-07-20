package httpapi

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestControlPlaneAuditCursorRoundTrip(t *testing.T) {
	encoded, err := encodeControlPlaneAuditCursor(AuditSummary{AuditRef: "audit_002", RecordedAt: "2026-07-12T10:20:30.123Z"})
	if err != nil {
		t.Fatalf("encode cursor: %v", err)
	}
	cursor, err := decodeControlPlaneAuditCursor(encoded)
	if err != nil {
		t.Fatalf("decode cursor: %v", err)
	}
	if cursor.Version != controlPlaneAuditCursorVersion || cursor.AuditRef != "audit_002" || cursor.RecordedAt != "2026-07-12T10:20:30.123Z" {
		t.Fatalf("unexpected cursor: %#v", cursor)
	}
	for _, invalid := range []string{"not-base64", "e30", strings.Repeat("a", 1025)} {
		if _, err := decodeControlPlaneAuditCursor(invalid); err == nil {
			t.Fatalf("invalid cursor was accepted: %s", invalid[:min(len(invalid), 32)])
		}
	}
}

func TestControlPlaneAuditStrictPaginationRequest(t *testing.T) {
	valid := httptest.NewRequest("GET", "/v1/control-plane/audit?limit=25&sort=recorded_at_desc&event_kind=read", nil)
	request, failure := controlPlaneReadRepositoryRequestFromQuery(valid, map[string]string{"event_kind": "read"}, true)
	if failure != "" || request.Limit != 25 || request.Sort != "recorded_at_desc" {
		t.Fatalf("unexpected valid request: %#v failure=%s", request, failure)
	}

	for _, target := range []string{
		"/v1/control-plane/audit?limit=0",
		"/v1/control-plane/audit?limit=101",
		"/v1/control-plane/audit?limit=abc",
		"/v1/control-plane/audit?limit=10&limit=20",
		"/v1/control-plane/audit?sort=recorded_at_asc",
		"/v1/control-plane/audit?cursor=not-base64",
		"/v1/control-plane/audit?cursor=" + strings.Repeat("a", 1025),
	} {
		req := httptest.NewRequest("GET", target, nil)
		if _, failure := controlPlaneReadRepositoryRequestFromQuery(req, nil, true); failure != "invalid_filter" {
			t.Fatalf("invalid pagination was accepted: %s", target)
		}
	}
}

func TestRoutedControlPlaneReadRepositoryKeepsWorkspaceOperationsOnFakeStore(t *testing.T) {
	workspace := newControlPlaneReadRepository(newControlPlaneReadFakeStore())
	routed := &routedControlPlaneReadRepository{workspace: workspace}
	context := ReadRepositoryContext{TenantRef: "tenant_demo", AuditRef: "audit_route_test"}
	if len(routed.ListApplicationSummaries(context, ListApplicationSummariesRequest{}).Items) == 0 {
		t.Fatal("applications were not routed to the workspace repository")
	}
	if len(routed.ListAPIKeySummaries(context, ListAPIKeySummariesRequest{}).Items) == 0 {
		t.Fatal("API keys were not routed to the workspace repository")
	}
	if len(routed.ReadQuotaSummary(context, ReadQuotaSummaryRequest{}).Items) == 0 {
		t.Fatal("quota was not routed to the workspace repository")
	}
	if len(routed.ListWorkflowDefinitionSummaries(context, ListWorkflowDefinitionSummariesRequest{}).Items) == 0 {
		t.Fatal("workflow definitions were not routed to the workspace repository")
	}
	if len(routed.ListRunRecordSummaries(context, ListRunRecordSummariesRequest{}).Items) == 0 {
		t.Fatal("run records were not routed to the workspace repository")
	}
}
