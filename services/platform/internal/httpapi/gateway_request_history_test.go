package httpapi

import (
	"sync"
	"testing"
	"time"
)

func TestMemoryGatewayRequestHistoryScopeFilterAndCursor(t *testing.T) {
	store := newMemoryGatewayRequestStore(20)
	requestContext := gatewayRequestTestContext()
	base := time.Date(2026, 7, 12, 8, 0, 0, 0, time.UTC)
	for index, status := range []GatewayRequestStatus{
		GatewayRequestStatusSucceeded, GatewayRequestStatusFailed, GatewayRequestStatusSucceeded,
	} {
		record := gatewayRequestTestRecord(requestContext, "request_"+string(rune('a'+index)), base.Add(time.Duration(index)*time.Minute))
		if err := store.CreateRequest(requestContext, &record); err != nil {
			t.Fatal(err)
		}
		record.Status = status
		record.CompletedAt = base.Add(time.Duration(index)*time.Minute + time.Second).Format(time.RFC3339Nano)
		record.DurationMS = 1000
		record.HTTPStatusCode = 200
		if status == GatewayRequestStatusFailed {
			record.HTTPStatusCode = 502
			record.FailureCode = "PLATFORM_GATEWAY_FAILED"
			record.FailureBoundary = errorBoundaryPythonBridge
		}
		if err := store.UpdateRequest(requestContext, &record); err != nil {
			t.Fatal(err)
		}
	}
	service := newGatewayRequestHistoryService(store)
	first := service.List(requestContext, GatewayRequestListRequest{Limit: 2})
	if first.FailureCode != "" || len(first.Records) != 2 || !first.HasMore || first.NextCursor == "" || first.Records[0].RequestID != "request_c" {
		t.Fatalf("unexpected first page: %#v", first)
	}
	second := service.List(requestContext, GatewayRequestListRequest{Limit: 2, Cursor: first.NextCursor})
	if second.FailureCode != "" || len(second.Records) != 1 || second.Records[0].RequestID != "request_a" {
		t.Fatalf("unexpected second page: %#v", second)
	}
	filtered := service.List(requestContext, GatewayRequestListRequest{Status: GatewayRequestStatusFailed})
	if len(filtered.Records) != 1 || filtered.Records[0].RequestID != "request_b" {
		t.Fatalf("unexpected filter: %#v", filtered)
	}
	if changed := service.List(requestContext, GatewayRequestListRequest{Limit: 3, Cursor: first.NextCursor}); changed.FailureCode != GatewayRequestHistoryFailureCursorInvalid {
		t.Fatalf("cursor must bind filter: %#v", changed)
	}
	other := requestContext
	other.TenantRef = "tenant_other"
	if scoped := service.List(other, GatewayRequestListRequest{}); len(scoped.Records) != 0 {
		t.Fatalf("cross-tenant request history leaked: %#v", scoped)
	}
}

func TestMemoryGatewayRequestStoreRejectsConcurrentTerminalRewrite(t *testing.T) {
	store := newMemoryGatewayRequestStore(10)
	requestContext := gatewayRequestTestContext()
	record := gatewayRequestTestRecord(requestContext, "request_concurrent", time.Now().UTC())
	if err := store.CreateRequest(requestContext, &record); err != nil {
		t.Fatal(err)
	}
	left, right := record, record
	var wait sync.WaitGroup
	results := make(chan error, 2)
	for _, candidate := range []*GatewayRequestRecord{&left, &right} {
		wait.Add(1)
		go func(value *GatewayRequestRecord) {
			defer wait.Done()
			value.Status = GatewayRequestStatusSucceeded
			value.CompletedAt = time.Now().UTC().Format(time.RFC3339Nano)
			value.HTTPStatusCode = 200
			results <- store.UpdateRequest(requestContext, value)
		}(candidate)
	}
	wait.Wait()
	close(results)
	successes := 0
	for err := range results {
		if err == nil {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("exactly one terminal update must succeed: %d", successes)
	}
	stored, _, _ := store.ReadRequest(requestContext, record.RequestID)
	stored.Status = GatewayRequestStatusStarted
	stored.CompletedAt = ""
	stored.HTTPStatusCode = 0
	if err := store.UpdateRequest(requestContext, &stored); err != errGatewayRequestStoreConflict {
		t.Fatalf("terminal rewrite must fail: %v", err)
	}
}

func TestGatewayRequestStoreRejectsSensitiveAndInvalidUsage(t *testing.T) {
	store := newMemoryGatewayRequestStore(10)
	requestContext := gatewayRequestTestContext()
	record := gatewayRequestTestRecord(requestContext, "request_endpoint", time.Now().UTC())
	record.SelectedProvider = "https://provider.invalid/raw"
	if err := store.CreateRequest(requestContext, &record); err != errGatewayRequestStoreContract {
		t.Fatalf("provider endpoint accepted: %v", err)
	}
	record = gatewayRequestTestRecord(requestContext, "request_usage", time.Now().UTC())
	record.Usage = GatewayRequestUsage{Availability: GatewayRequestUsageReported, Source: "provider_response", InputTokens: 2, OutputTokens: 3, TotalTokens: 4}
	if err := store.CreateRequest(requestContext, &record); err != errGatewayRequestStoreContract {
		t.Fatalf("inconsistent usage accepted: %v", err)
	}
}

func gatewayRequestTestContext() GatewayRequestContext {
	return GatewayRequestContext{
		TenantRef: "tenant_demo", WorkspaceID: "workspace_demo", ConsumerRef: "consumer_demo",
		ApplicationID: "application_demo", SubjectRef: "subject_demo", ScopeGrants: []string{"gateway_requests:read"},
		AuditContext: "audit_context_demo", Source: "dev_headers", RequestID: "query_request", AuditRef: "query_audit",
	}
}

func gatewayRequestTestRecord(
	requestContext GatewayRequestContext,
	requestID string,
	startedAt time.Time,
) GatewayRequestRecord {
	return GatewayRequestRecord{
		SchemaVersion: gatewayRequestRecordSchemaVersion, RequestID: requestID, AuditRef: "audit_" + requestID,
		TenantRef: requestContext.TenantRef, WorkspaceID: requestContext.WorkspaceID,
		ConsumerRef: requestContext.ConsumerRef, ApplicationID: requestContext.ApplicationID,
		SubjectRef: requestContext.SubjectRef, Route: "/v1/chat/completions",
		Protocol: northboundProtocolChatCompletions, Status: GatewayRequestStatusStarted,
		StartedAt: startedAt.UTC().Format(time.RFC3339Nano), Usage: GatewayRequestUsage{Availability: GatewayRequestUsageNotReported},
	}
}
