package httpapi

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestMemoryGatewayRequestQuotaRepositoryAdmitsAtLimitAtomically(t *testing.T) {
	repository := newMemoryGatewayRequestQuotaRepository()
	quotaContext := testGatewayRequestQuotaContext("app-one")
	now := time.Date(2026, 8, 9, 10, 0, 0, 0, time.UTC)
	if _, err := repository.PutPolicy(quotaContext, 0, 1, now); err != nil {
		t.Fatalf("put policy: %v", err)
	}
	start := make(chan struct{})
	errorsByAttempt := make(chan error, 2)
	var waitGroup sync.WaitGroup
	for attempt := 0; attempt < 2; attempt++ {
		waitGroup.Add(1)
		go func(attempt int) {
			defer waitGroup.Done()
			<-start
			_, err := repository.AdmitProviderAttempt(quotaContext, GatewayRequestQuotaAdmissionInput{
				APIKeyID: "key-one", RequestID: fmt.Sprintf("request-%d", attempt),
				Route: "POST /v1/responses", AdmittedAt: now,
			})
			errorsByAttempt <- err
		}(attempt)
	}
	close(start)
	waitGroup.Wait()
	close(errorsByAttempt)
	admitted := 0
	exceeded := 0
	for err := range errorsByAttempt {
		switch {
		case err == nil:
			admitted++
		case errors.Is(err, errGatewayRequestQuotaExceeded):
			exceeded++
		default:
			t.Fatalf("unexpected admission error: %v", err)
		}
	}
	if admitted != 1 || exceeded != 1 {
		t.Fatalf("expected one admission and one rejection, got admitted=%d exceeded=%d", admitted, exceeded)
	}
}

func TestMemoryGatewayRequestQuotaRepositoryPreservesUsageAcrossPolicyUpdate(t *testing.T) {
	repository := newMemoryGatewayRequestQuotaRepository()
	quotaContext := testGatewayRequestQuotaContext("app-one")
	now := time.Date(2026, 8, 9, 23, 55, 0, 0, time.UTC)
	policy, err := repository.PutPolicy(quotaContext, 0, 2, now)
	if err != nil {
		t.Fatalf("put policy: %v", err)
	}
	if _, err := repository.AdmitProviderAttempt(quotaContext, testQuotaAdmission("request-one", now)); err != nil {
		t.Fatalf("admit attempt: %v", err)
	}
	policy, err = repository.PutPolicy(quotaContext, policy.RecordVersion, 3, now.Add(time.Minute))
	if err != nil {
		t.Fatalf("update policy: %v", err)
	}
	usage, found, err := repository.ReadUsage(quotaContext, "2026-08-09")
	if err != nil || !found {
		t.Fatalf("read usage: found=%t err=%v", found, err)
	}
	if usage.AdmittedRequestCount != 1 || usage.RemainingRequestCount != 2 || usage.PolicyVersion != policy.RecordVersion {
		t.Fatalf("unexpected usage after policy update: %+v", usage)
	}
	decision, err := repository.AdmitProviderAttempt(quotaContext, testQuotaAdmission("request-next-day", now.Add(10*time.Minute)))
	if err != nil {
		t.Fatalf("admit next period: %v", err)
	}
	if decision.Usage.PeriodStart != "2026-08-10" || decision.Usage.AdmittedRequestCount != 1 {
		t.Fatalf("unexpected next period usage: %+v", decision.Usage)
	}
}

func TestMemoryGatewayRequestQuotaRepositoryFailsClosedForConflictsAndScopes(t *testing.T) {
	repository := newMemoryGatewayRequestQuotaRepository()
	now := time.Date(2026, 8, 9, 10, 0, 0, 0, time.UTC)
	quotaContext := testGatewayRequestQuotaContext("app-one")
	if _, err := repository.PutPolicy(quotaContext, 0, 2, now); err != nil {
		t.Fatalf("put policy: %v", err)
	}
	if _, err := repository.PutPolicy(quotaContext, 0, 3, now); !errors.Is(err, errGatewayRequestQuotaPolicyVersionConflict) {
		t.Fatalf("expected version conflict, got %v", err)
	}
	input := testQuotaAdmission("request-one", now)
	if _, err := repository.AdmitProviderAttempt(quotaContext, input); err != nil {
		t.Fatalf("first admission: %v", err)
	}
	if _, err := repository.AdmitProviderAttempt(quotaContext, input); !errors.Is(err, errGatewayRequestQuotaAttemptConflict) {
		t.Fatalf("expected attempt conflict, got %v", err)
	}
	otherScope := testGatewayRequestQuotaContext("app-two")
	if _, err := repository.AdmitProviderAttempt(otherScope, testQuotaAdmission("request-two", now)); !errors.Is(err, errGatewayRequestQuotaPolicyNotFound) {
		t.Fatalf("expected policy not found across scope, got %v", err)
	}
	production := quotaContext
	production.Environment = "production"
	if _, err := repository.PutPolicy(production, 0, 1, now); !errors.Is(err, errGatewayRequestQuotaContract) {
		t.Fatalf("expected production contract rejection, got %v", err)
	}
}

func testGatewayRequestQuotaContext(applicationID string) GatewayRequestQuotaContext {
	return GatewayRequestQuotaContext{
		RequestContext: context.Background(),
		TenantRef:      "tenant-one", WorkspaceID: "workspace-one", Environment: "test",
		ApplicationID: applicationID, ActorRef: "subject-one", RequestID: "request-one", AuditRef: "audit-one",
	}
}

func testQuotaAdmission(requestID string, admittedAt time.Time) GatewayRequestQuotaAdmissionInput {
	return GatewayRequestQuotaAdmissionInput{
		APIKeyID: "key-one", RequestID: requestID, Route: "POST /v1/responses", AdmittedAt: admittedAt,
	}
}
