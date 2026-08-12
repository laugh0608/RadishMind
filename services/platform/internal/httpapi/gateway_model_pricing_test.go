package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestMemoryGatewayModelPricingRepositoryRetainsImmutableRevisions(t *testing.T) {
	repository := newMemoryGatewayModelPricingRepository()
	pricingContext := gatewayModelPricingTestContext("model_alpha")
	first := putGatewayModelPricingTestPolicy(t, repository, pricingContext, 0, 1_000_000, 3_000_000, "initial development evidence")
	second := putGatewayModelPricingTestPolicy(t, repository, pricingContext, 1, 2_000_000, 4_000_000, "updated test evidence")

	if first.RecordVersion != 1 || second.RecordVersion != 2 || first.PolicyID != second.PolicyID || first.PolicyDigest == second.PolicyDigest {
		t.Fatalf("unexpected immutable revisions: first=%+v second=%+v", first, second)
	}
	current, found, err := repository.ReadCurrent(pricingContext)
	if err != nil || !found || current.RecordVersion != 2 || current.InputPriceMicrosPerTokenUnit != 2_000_000 {
		t.Fatalf("unexpected current policy: policy=%+v found=%v err=%v", current, found, err)
	}
	retained, found, err := repository.ReadRevision(pricingContext, 1)
	if err != nil || !found || retained != first {
		t.Fatalf("first revision was not retained: policy=%+v found=%v err=%v", retained, found, err)
	}
}

func TestMemoryGatewayModelPricingRepositoryIsolatesExactScope(t *testing.T) {
	repository := newMemoryGatewayModelPricingRepository()
	alphaContext := gatewayModelPricingTestContext("model_alpha")
	betaContext := gatewayModelPricingTestContext("model_beta")
	alpha := putGatewayModelPricingTestPolicy(t, repository, alphaContext, 0, 10, 20, "alpha model evidence")
	beta := putGatewayModelPricingTestPolicy(t, repository, betaContext, 0, 30, 40, "beta model evidence")

	if alpha.PolicyID == beta.PolicyID {
		t.Fatal("exact model scopes must not share a policy identity")
	}
	missingContext := betaContext
	missingContext.WorkspaceID = "workspace_other"
	if _, found, err := repository.ReadCurrent(missingContext); err != nil || found {
		t.Fatalf("cross-workspace policy leaked: found=%v err=%v", found, err)
	}
}

func TestMemoryGatewayModelPricingRepositoryCASHasSingleWinner(t *testing.T) {
	repository := newMemoryGatewayModelPricingRepository()
	pricingContext := gatewayModelPricingTestContext("model_alpha")
	putGatewayModelPricingTestPolicy(t, repository, pricingContext, 0, 10, 20, "initial evidence")

	var winners atomic.Int64
	var conflicts atomic.Int64
	var waitGroup sync.WaitGroup
	for index := 0; index < 24; index++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			_, err := repository.PutRevision(pricingContext, GatewayModelPricingPolicyInput{
				ExpectedVersion: 1, Currency: GatewayModelPricingCurrency,
				InputPriceMicrosPerTokenUnit: 100, OutputPriceMicrosPerTokenUnit: 200,
				Reason: "concurrent update evidence",
			}, time.Now().UTC())
			switch {
			case err == nil:
				winners.Add(1)
			case errors.Is(err, errGatewayModelPricingVersionConflict):
				conflicts.Add(1)
			default:
				t.Errorf("unexpected concurrent error: %v", err)
			}
		}()
	}
	waitGroup.Wait()
	if winners.Load() != 1 || conflicts.Load() != 23 {
		t.Fatalf("CAS did not produce one winner: winners=%d conflicts=%d", winners.Load(), conflicts.Load())
	}
	current, found, err := repository.ReadCurrent(pricingContext)
	if err != nil || !found || current.RecordVersion != 2 {
		t.Fatalf("unexpected current revision after concurrency: policy=%+v found=%v err=%v", current, found, err)
	}
}

func TestGatewayModelPricingServiceReturnsStableFailuresAndConflictVersion(t *testing.T) {
	repository := newMemoryGatewayModelPricingRepository()
	service := newGatewayModelPricingService(repository)
	service.now = func() time.Time { return time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC) }
	pricingContext := gatewayModelPricingTestContext("model_alpha")

	if result := service.ReadCurrent(pricingContext); result.FailureCode != GatewayModelPricingFailurePolicyNotFound {
		t.Fatalf("unexpected missing policy result: %+v", result)
	}
	invalidEnvironment := pricingContext
	invalidEnvironment.Environment = "production"
	if result := service.ReadCurrent(invalidEnvironment); result.FailureCode != GatewayModelPricingFailureEnvironmentForbidden {
		t.Fatalf("unexpected environment result: %+v", result)
	}
	created := service.PutRevision(pricingContext, GatewayModelPricingPolicyInput{
		ExpectedVersion: 0, Currency: GatewayModelPricingCurrency,
		InputPriceMicrosPerTokenUnit: 1, OutputPriceMicrosPerTokenUnit: 2,
		Reason: "service create evidence",
	})
	if created.FailureCode != "" || created.Policy == nil || created.CurrentVersion != 1 {
		t.Fatalf("unexpected create result: %+v", created)
	}
	conflict := service.PutRevision(pricingContext, GatewayModelPricingPolicyInput{
		ExpectedVersion: 0, Currency: GatewayModelPricingCurrency,
		InputPriceMicrosPerTokenUnit: 3, OutputPriceMicrosPerTokenUnit: 4,
		Reason: "stale update evidence",
	})
	if conflict.FailureCode != GatewayModelPricingFailureVersionConflict || conflict.CurrentVersion != 1 || conflict.Policy != nil {
		t.Fatalf("unexpected conflict result: %+v", conflict)
	}
	if result := newGatewayModelPricingService(nil).ReadCurrent(pricingContext); result.FailureCode != GatewayModelPricingFailureStoreUnavailable {
		t.Fatalf("nil repository did not fail closed: %+v", result)
	}
}

func TestGatewayModelPricingRejectsInvalidContractAndDetectsStoredDigestDrift(t *testing.T) {
	repository := newMemoryGatewayModelPricingRepository()
	pricingContext := gatewayModelPricingTestContext("model_alpha")
	invalidInputs := []GatewayModelPricingPolicyInput{
		{ExpectedVersion: -1, Currency: "USD", Reason: "negative version"},
		{ExpectedVersion: 0, Currency: "EUR", Reason: "unsupported currency"},
		{ExpectedVersion: 0, Currency: "USD", InputPriceMicrosPerTokenUnit: -1, Reason: "negative rate"},
		{ExpectedVersion: 0, Currency: "USD", Reason: "contains https://provider.example contract"},
		{ExpectedVersion: 0, Currency: "USD", Reason: "contains secret material"},
	}
	for _, input := range invalidInputs {
		if _, err := repository.PutRevision(pricingContext, input, time.Now().UTC()); !errors.Is(err, errGatewayModelPricingContract) {
			t.Fatalf("invalid input was accepted: input=%+v err=%v", input, err)
		}
	}
	policy := putGatewayModelPricingTestPolicy(t, repository, pricingContext, 0, 10, 20, "safe pricing evidence")
	key := gatewayModelPricingScopeKey(pricingContext)
	tampered := repository.revisions[key][policy.RecordVersion]
	tampered.InputPriceMicrosPerTokenUnit++
	repository.revisions[key][policy.RecordVersion] = tampered
	if _, _, err := repository.ReadCurrent(pricingContext); !errors.Is(err, errGatewayModelPricingStoreUnavailable) {
		t.Fatalf("stored digest drift did not fail closed: %v", err)
	}
}

func TestGatewayRequestCostEstimateAvailabilityPrecedence(t *testing.T) {
	reported := gatewayModelPricingTestUsage(10, 5)
	missingPrice := gatewayModelPricingUnavailableSnapshot(GatewayPricingSnapshotNotConfigured, "not found")
	unavailablePrice := gatewayModelPricingUnavailableSnapshot(GatewayPricingSnapshotUnavailable, "store unavailable")

	cases := []struct {
		name              string
		providerAttempted bool
		usage             GatewayRequestUsage
		snapshot          GatewayModelPricingSnapshot
		availability      string
	}{
		{"provider not attempted", false, reported, missingPrice, GatewayRequestCostNotApplicable},
		{"usage missing before price", true, GatewayRequestUsage{Availability: GatewayRequestUsageNotReported}, missingPrice, GatewayRequestCostUsageNotReported},
		{"price not configured", true, reported, missingPrice, GatewayRequestCostPriceNotConfigured},
		{"price unavailable", true, reported, unavailablePrice, GatewayRequestCostPriceUnavailable},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			estimate := buildGatewayRequestCostEstimate(testCase.providerAttempted, testCase.usage, testCase.snapshot)
			if estimate.SchemaVersion != GatewayRequestCostEstimateSchemaVersion || estimate.Availability != testCase.availability ||
				estimate.Reason == "" || estimate.EstimatedCostMicros != nil {
				t.Fatalf("unexpected unavailable estimate: %+v", estimate)
			}
		})
	}
	legacy := gatewayRequestLegacyCostEstimate()
	if legacy.Availability != GatewayRequestCostLegacyNotCaptured {
		t.Fatalf("unexpected legacy estimate: %+v", legacy)
	}
}

func TestGatewayRequestCostEstimateUsesIntegerHalfUpAndPreservesZero(t *testing.T) {
	pricingContext := gatewayModelPricingTestContext("model_alpha")
	policy := GatewayModelPricingPolicy{
		SchemaVersion: GatewayModelPricingPolicySchemaVersion,
		PolicyID:      gatewayModelPricingPolicyID(pricingContext), RecordVersion: 1,
		TenantRef: pricingContext.TenantRef, WorkspaceID: pricingContext.WorkspaceID,
		Environment: pricingContext.Environment, ProviderID: pricingContext.ProviderID,
		ProfileID: pricingContext.ProfileID, ModelID: pricingContext.ModelID,
		Currency: GatewayModelPricingCurrency, TokenUnit: GatewayModelPricingTokenUnit,
		InputPriceMicrosPerTokenUnit: 500_000, OutputPriceMicrosPerTokenUnit: 0,
		Reason: "rounding evidence", UpdatedAt: time.Now().UTC(), UpdatedByActorRef: pricingContext.ActorRef,
		RequestID: pricingContext.RequestID, AuditRef: pricingContext.AuditRef,
	}
	policy.PolicyDigest = gatewayModelPricingPolicyDigest(policy)
	snapshot := gatewayModelPricingSnapshotFromPolicy(policy)
	estimate := buildGatewayRequestCostEstimate(true, gatewayModelPricingTestUsage(1, 0), snapshot)
	if estimate.Availability != GatewayRequestCostEstimated || estimate.EstimatedCostMicros == nil || *estimate.EstimatedCostMicros != 1 {
		t.Fatalf("half-up rounding failed: %+v", estimate)
	}

	policy.InputPriceMicrosPerTokenUnit = 0
	policy.PolicyDigest = gatewayModelPricingPolicyDigest(policy)
	snapshot = gatewayModelPricingSnapshotFromPolicy(policy)
	zero := buildGatewayRequestCostEstimate(true, gatewayModelPricingTestUsage(0, 0), snapshot)
	if zero.Availability != GatewayRequestCostEstimated || zero.EstimatedCostMicros == nil || *zero.EstimatedCostMicros != 0 ||
		zero.InputPriceMicrosPerTokenUnit == nil || *zero.InputPriceMicrosPerTokenUnit != 0 {
		t.Fatalf("zero-cost evidence was lost: %+v", zero)
	}
	payload, err := json.Marshal(zero)
	if err != nil || !strings.Contains(string(payload), `"estimated_cost_micros":0`) ||
		!strings.Contains(string(payload), `"input_price_micros_per_token_unit":0`) {
		t.Fatalf("zero-cost JSON did not preserve explicit zeros: payload=%s err=%v", payload, err)
	}
}

func TestGatewayRequestCostEstimateFailsClosedOnDriftOverflowAndInvalidUsage(t *testing.T) {
	repository := newMemoryGatewayModelPricingRepository()
	pricingContext := gatewayModelPricingTestContext("model_alpha")
	policy := putGatewayModelPricingTestPolicy(t, repository, pricingContext, 0, math.MaxInt64, math.MaxInt64, "overflow boundary evidence")
	snapshot := gatewayModelPricingSnapshotFromPolicy(policy)

	drifted := snapshot
	drifted.OutputPriceMicrosPerTokenUnit--
	if estimate := buildGatewayRequestCostEstimate(true, gatewayModelPricingTestUsage(1, 1), drifted); estimate.Availability != GatewayRequestCostPriceUnavailable {
		t.Fatalf("snapshot drift did not fail closed: %+v", estimate)
	}
	if estimate := buildGatewayRequestCostEstimate(true, gatewayModelPricingTestUsage(math.MaxInt, 0), snapshot); estimate.Availability != GatewayRequestCostPriceUnavailable {
		t.Fatalf("overflow did not fail closed: %+v", estimate)
	}
	invalidUsage := GatewayRequestUsage{
		Availability: GatewayRequestUsageReported, Source: "openai_compatible_usage",
		InputTokens: 2, OutputTokens: 3, TotalTokens: 4,
	}
	if estimate := buildGatewayRequestCostEstimate(true, invalidUsage, snapshot); estimate.Availability != GatewayRequestCostUsageNotReported {
		t.Fatalf("invalid usage did not fail closed: %+v", estimate)
	}
}

func gatewayModelPricingTestContext(modelID string) GatewayModelPricingContext {
	return GatewayModelPricingContext{
		RequestContext: context.Background(), TenantRef: "tenant_demo", WorkspaceID: "workspace_demo",
		Environment: "development", ProviderID: "provider_demo", ProfileID: "profile_demo", ModelID: modelID,
		ActorRef: "subject_admin", RequestID: "request_pricing", AuditRef: "audit_pricing",
	}
}

func putGatewayModelPricingTestPolicy(
	t *testing.T,
	repository GatewayModelPricingRepository,
	pricingContext GatewayModelPricingContext,
	expectedVersion int64,
	inputRate int64,
	outputRate int64,
	reason string,
) GatewayModelPricingPolicy {
	t.Helper()
	policy, err := repository.PutRevision(pricingContext, GatewayModelPricingPolicyInput{
		ExpectedVersion: expectedVersion, Currency: GatewayModelPricingCurrency,
		InputPriceMicrosPerTokenUnit: inputRate, OutputPriceMicrosPerTokenUnit: outputRate,
		Reason: reason,
	}, time.Date(2026, 8, 12, 0, int(expectedVersion), 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("put pricing policy: %v", err)
	}
	return policy
}

func gatewayModelPricingTestUsage(inputTokens int, outputTokens int) GatewayRequestUsage {
	return GatewayRequestUsage{
		Availability: GatewayRequestUsageReported, Source: "openai_compatible_usage",
		InputTokens: inputTokens, OutputTokens: outputTokens, TotalTokens: inputTokens + outputTokens,
	}
}
