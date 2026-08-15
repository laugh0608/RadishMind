package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/bridge"
	"radishmind.local/services/platform/internal/config"
)

type gatewayProviderAttemptBridgeResult struct {
	envelope bridge.GatewayEnvelope
	err      error
}

type gatewayProviderAttemptScriptedBridge struct {
	mu        sync.Mutex
	inventory bridge.ProviderInventory
	results   []gatewayProviderAttemptBridgeResult
	options   []bridge.EnvelopeOptions
	onCall    func(int)
}

func (client *gatewayProviderAttemptScriptedBridge) DescribeProviders(context.Context) ([]bridge.ProviderDescription, error) {
	return nil, nil
}

func (client *gatewayProviderAttemptScriptedBridge) DescribeInventory(context.Context) (bridge.ProviderInventory, error) {
	client.mu.Lock()
	defer client.mu.Unlock()
	return client.inventory, nil
}

func (client *gatewayProviderAttemptScriptedBridge) HandleEnvelope(
	_ context.Context,
	_ []byte,
	options bridge.EnvelopeOptions,
) (bridge.GatewayEnvelope, error) {
	client.mu.Lock()
	index := len(client.options)
	client.options = append(client.options, options)
	if index >= len(client.results) {
		client.mu.Unlock()
		return bridge.GatewayEnvelope{}, errors.New("unexpected Provider attempt")
	}
	result := client.results[index]
	hook := client.onCall
	client.mu.Unlock()
	if hook != nil {
		hook(index)
	}
	return result.envelope, result.err
}

func (client *gatewayProviderAttemptScriptedBridge) StreamEnvelope(
	context.Context,
	[]byte,
	bridge.EnvelopeOptions,
	func(bridge.StreamEvent) error,
) error {
	return errors.New("streaming must remain outside Provider fallback execution")
}

func (client *gatewayProviderAttemptScriptedBridge) reset(results ...gatewayProviderAttemptBridgeResult) {
	client.mu.Lock()
	defer client.mu.Unlock()
	client.results = append([]gatewayProviderAttemptBridgeResult(nil), results...)
	client.options = nil
	client.onCall = nil
}

func (client *gatewayProviderAttemptScriptedBridge) callOptions() []bridge.EnvelopeOptions {
	client.mu.Lock()
	defer client.mu.Unlock()
	return append([]bridge.EnvelopeOptions(nil), client.options...)
}

type gatewayProviderAttemptExecutorFixture struct {
	server           *Server
	token            string
	history          gatewayRequestStore
	quota            *memoryGatewayRequestQuotaRepository
	pricing          *memoryGatewayModelPricingRepository
	bridge           *gatewayProviderAttemptScriptedBridge
	snapshotProvider *mutableGatewayProviderRouteSnapshotProvider
	historyContext   GatewayRequestContext
	quotaContext     GatewayRequestQuotaContext
}

func TestGatewayProviderAttemptUnaryFallbackSuccessAcrossThreeProtocols(t *testing.T) {
	fixture := newGatewayProviderAttemptExecutorFixture(t, 20)
	for index, test := range gatewayProviderAttemptUnaryRoutes(fixture.server) {
		requestID := "request-fallback-success-" + test.name
		fixture.bridge.reset(
			gatewayProviderAttemptFailedResult(t, bridge.ProviderFailureTemporarilyUnavailable, bridge.ProviderFallbackEligible),
			gatewayProviderAttemptSucceededResult("fallback succeeded"),
		)
		recorder := fixture.invoke(t, test, requestID, true, false)
		if recorder.Code != http.StatusOK || recorder.Header().Get("X-RadishMind-Provider-Attempts") != "2" ||
			recorder.Header().Get("X-RadishMind-Fallback-Used") != "true" {
			t.Fatalf("%s fallback response: status=%d headers=%v body=%s", test.name, recorder.Code, recorder.Header(), recorder.Body.String())
		}
		options := fixture.bridge.callOptions()
		if len(options) != 2 || options[0].ProviderProfile != "mock-primary" ||
			options[1].ProviderProfile != "mock-secondary" || options[0].Model != "mock-model-primary" ||
			options[1].Model != "mock-model-secondary" {
			t.Fatalf("%s did not preserve reviewed target order: %#v", test.name, options)
		}
		record := fixture.readHistory(t, requestID)
		if record.SchemaVersion != gatewayRequestRecordSchemaVersionV3 || record.Status != GatewayRequestStatusSucceeded ||
			record.ProviderAttemptCount != 2 || !record.FallbackAllowed || !record.FallbackUsed ||
			record.ProviderAttempts[0].Failure == nil ||
			record.ProviderAttempts[1].Status != GatewayProviderAttemptStatusSucceeded ||
			record.ProviderAttemptPlan.Protocol != test.protocol {
			t.Fatalf("%s history lost attempt lineage: %#v", test.name, record)
		}
		if index == 0 && record.ProviderAttemptPlan.MaxAttempts != 2 {
			t.Fatalf("fallback plan did not allow two attempts: %#v", record.ProviderAttemptPlan)
		}
	}
}

func TestGatewayProviderAttemptDoesNotFallbackImplicitlyOrForIneligibleFailure(t *testing.T) {
	fixture := newGatewayProviderAttemptExecutorFixture(t, 10)
	test := gatewayProviderAttemptUnaryRoutes(fixture.server)[0]
	for _, scenario := range []struct {
		name          string
		allowFallback bool
		failureClass  bridge.ProviderAttemptFailureClass
		disposition   bridge.ProviderAttemptFallbackDisposition
	}{
		{name: "omitted opt-in", failureClass: bridge.ProviderFailureTemporarilyUnavailable, disposition: bridge.ProviderFallbackEligible},
		{name: "ineligible failure", allowFallback: true, failureClass: bridge.ProviderFailureInvalidRequest, disposition: bridge.ProviderFallbackIneligible},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			fixture.bridge.reset(gatewayProviderAttemptFailedResult(t, scenario.failureClass, scenario.disposition))
			requestID := "request-no-fallback-" + strings.ReplaceAll(scenario.name, " ", "-")
			recorder := fixture.invoke(t, test, requestID, scenario.allowFallback, false)
			if recorder.Code != http.StatusBadGateway || len(fixture.bridge.callOptions()) != 1 ||
				recorder.Header().Get("X-RadishMind-Provider-Attempts") != "1" ||
				recorder.Header().Get("X-RadishMind-Fallback-Used") != "false" {
				t.Fatalf("unsafe fallback occurred: status=%d headers=%v calls=%d body=%s",
					recorder.Code, recorder.Header(), len(fixture.bridge.callOptions()), recorder.Body.String())
			}
			record := fixture.readHistory(t, requestID)
			if record.ProviderAttemptCount != 1 || record.FallbackUsed || record.Status != GatewayRequestStatusFailed {
				t.Fatalf("unsafe fallback history: %#v", record)
			}
		})
	}
}

func TestGatewayProviderAttemptPrimarySuccessAndTwoFailuresHaveExactTerminalEvidence(t *testing.T) {
	t.Run("primary success", func(t *testing.T) {
		fixture := newGatewayProviderAttemptExecutorFixture(t, 10)
		fixture.bridge.reset(gatewayProviderAttemptSucceededResult("primary succeeded"))
		recorder := fixture.invoke(t, gatewayProviderAttemptUnaryRoutes(fixture.server)[0], "request-primary-success", true, false)
		if recorder.Code != http.StatusOK || len(fixture.bridge.callOptions()) != 1 ||
			recorder.Header().Get("X-RadishMind-Fallback-Used") != "false" {
			t.Fatalf("primary success reached backup: status=%d calls=%d headers=%v", recorder.Code, len(fixture.bridge.callOptions()), recorder.Header())
		}
		record := fixture.readHistory(t, "request-primary-success")
		if record.ProviderAttemptCount != 1 || record.FallbackUsed || record.TerminalAttemptID != record.ProviderAttempts[0].AttemptID {
			t.Fatalf("primary success terminal evidence drifted: %#v", record)
		}
	})

	t.Run("two typed failures", func(t *testing.T) {
		fixture := newGatewayProviderAttemptExecutorFixture(t, 10)
		fixture.bridge.reset(
			gatewayProviderAttemptFailedResult(t, bridge.ProviderFailureRateLimited, bridge.ProviderFallbackEligible),
			gatewayProviderAttemptFailedResult(t, bridge.ProviderFailureTemporarilyUnavailable, bridge.ProviderFallbackEligible),
		)
		recorder := fixture.invoke(t, gatewayProviderAttemptUnaryRoutes(fixture.server)[1], "request-two-failures", true, false)
		if recorder.Code != http.StatusBadGateway || len(fixture.bridge.callOptions()) != 2 ||
			recorder.Header().Get("X-RadishMind-Fallback-Used") != "true" {
			t.Fatalf("two failures lost fallback boundary: status=%d calls=%d headers=%v body=%s",
				recorder.Code, len(fixture.bridge.callOptions()), recorder.Header(), recorder.Body.String())
		}
		record := fixture.readHistory(t, "request-two-failures")
		if record.Status != GatewayRequestStatusFailed || record.ProviderAttemptCount != 2 || !record.FallbackUsed ||
			record.TerminalAttemptID != record.ProviderAttempts[1].AttemptID || record.ProviderAttempts[1].Failure == nil {
			t.Fatalf("two failures terminal evidence drifted: %#v", record)
		}
	})

	t.Run("missing typed failure", func(t *testing.T) {
		fixture := newGatewayProviderAttemptExecutorFixture(t, 10)
		failed := gatewayProviderAttemptFailedResult(t, bridge.ProviderFailureRateLimited, bridge.ProviderFallbackEligible)
		failed.envelope.ProviderAttemptFailure = nil
		fixture.bridge.reset(failed)
		recorder := fixture.invoke(t, gatewayProviderAttemptUnaryRoutes(fixture.server)[2], "request-untyped-failure", true, false)
		if recorder.Code != http.StatusBadGateway || len(fixture.bridge.callOptions()) != 1 ||
			recorder.Header().Get("X-RadishMind-Fallback-Used") != "false" {
			t.Fatalf("untyped failure triggered fallback: status=%d calls=%d body=%s", recorder.Code, len(fixture.bridge.callOptions()), recorder.Body.String())
		}
		record := fixture.readHistory(t, "request-untyped-failure")
		if record.ProviderAttempts[0].Status != GatewayProviderAttemptStatusOutcomeUnknown ||
			record.ProviderAttempts[0].Failure != nil || record.ProviderAttempts[0].FailureBoundary != errorBoundarySouthboundProvider {
			t.Fatalf("untyped failure was not failed closed: %#v", record)
		}
	})
}

func TestGatewayProviderAttemptSecondQuotaDenialStopsBeforeBackupProvider(t *testing.T) {
	fixture := newGatewayProviderAttemptExecutorFixture(t, 1)
	test := gatewayProviderAttemptUnaryRoutes(fixture.server)[1]
	fixture.bridge.reset(
		gatewayProviderAttemptFailedResult(t, bridge.ProviderFailureRateLimited, bridge.ProviderFallbackEligible),
	)
	recorder := fixture.invoke(t, test, "request-second-quota-denied", true, false)
	if recorder.Code != http.StatusTooManyRequests || len(fixture.bridge.callOptions()) != 1 ||
		recorder.Header().Get("X-RadishMind-Provider-Attempts") != "1" ||
		recorder.Header().Get("X-RadishMind-Fallback-Used") != "false" {
		t.Fatalf("second quota denial crossed Provider boundary: status=%d headers=%v calls=%d body=%s",
			recorder.Code, recorder.Header(), len(fixture.bridge.callOptions()), recorder.Body.String())
	}
	record := fixture.readHistory(t, "request-second-quota-denied")
	if record.ProviderAttemptCount != 2 || record.ProviderAttempts[1].Status != GatewayProviderAttemptStatusQuotaRejected ||
		record.ProviderAttempts[1].QuotaRejectionCode != GatewayRequestQuotaFailureExceeded || record.FallbackUsed {
		t.Fatalf("second quota rejection evidence drifted: %#v", record)
	}
}

func TestGatewayProviderAttemptPrimaryQuotaDenialHasNoProviderAttemptHeader(t *testing.T) {
	fixture := newGatewayProviderAttemptExecutorFixture(t, 1)
	if _, err := fixture.quota.AdmitProviderAttempt(
		fixture.quotaContext,
		testQuotaAdmission("request-preexisting-attempt", time.Now().UTC()),
	); err != nil {
		t.Fatalf("seed exhausted Provider attempt quota: %v", err)
	}
	fixture.bridge.reset()
	recorder := fixture.invoke(
		t, gatewayProviderAttemptUnaryRoutes(fixture.server)[0], "request-primary-quota-denied", true, false,
	)
	if recorder.Code != http.StatusTooManyRequests || len(fixture.bridge.callOptions()) != 0 ||
		recorder.Header().Get("X-RadishMind-Provider-Attempts") != "" ||
		recorder.Header().Get("X-RadishMind-Fallback-Used") != "" {
		t.Fatalf("primary quota denial reported an unstarted Provider attempt: status=%d headers=%v calls=%d body=%s",
			recorder.Code, recorder.Header(), len(fixture.bridge.callOptions()), recorder.Body.String())
	}
	record := fixture.readHistory(t, "request-primary-quota-denied")
	if record.ProviderAttemptCount != 1 || record.ProviderAttempts[0].Status != GatewayProviderAttemptStatusQuotaRejected ||
		record.ProviderAttempts[0].QuotaRejectionCode != GatewayRequestQuotaFailureExceeded || record.FallbackUsed ||
		record.Usage.Availability != GatewayRequestUsageNotApplicable ||
		record.CostEstimate.Availability != GatewayRequestCostNotApplicable {
		t.Fatalf("primary quota rejection evidence drifted: %#v", record)
	}
}

func TestGatewayProviderAttemptPinsRouteAndPricingAcrossFallback(t *testing.T) {
	fixture := newGatewayProviderAttemptExecutorFixture(t, 10)
	seedGatewayProviderAttemptPricing(t, fixture, "mock-primary", 1)
	seedGatewayProviderAttemptPricing(t, fixture, "mock-secondary", 1)
	originalSnapshot := fixture.snapshotProvider.snapshot
	fixture.bridge.reset(
		gatewayProviderAttemptFailedResult(t, bridge.ProviderFailureUpstreamGatewayUnavailable, bridge.ProviderFallbackEligible),
		gatewayProviderAttemptSucceededResult("pinned fallback"),
	)
	fixture.bridge.onCall = func(index int) {
		if index != 0 {
			return
		}
		seedGatewayProviderAttemptPricing(t, fixture, "mock-secondary", 2)
		fixture.snapshotProvider.snapshot.Generation = originalSnapshot.Generation + 1
		fixture.snapshotProvider.snapshot.SnapshotDigest = "sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
	}
	recorder := fixture.invoke(t, gatewayProviderAttemptUnaryRoutes(fixture.server)[2], "request-pinned-plan", true, false)
	if recorder.Code != http.StatusOK || fixture.snapshotProvider.reads != 1 {
		t.Fatalf("request-local plan was not pinned: status=%d reads=%d body=%s", recorder.Code, fixture.snapshotProvider.reads, recorder.Body.String())
	}
	record := fixture.readHistory(t, "request-pinned-plan")
	if record.ProviderRouteGeneration != originalSnapshot.Generation ||
		record.ProviderRouteSnapshotDigest != originalSnapshot.SnapshotDigest ||
		record.ProviderAttemptPlan.Targets[1].PricingSnapshot.PricingPolicyVersion != 1 {
		t.Fatalf("route or pricing changed inside request: %#v", record)
	}
}

func TestGatewayProviderAttemptCancellationAndHistoryCheckpointFailClosed(t *testing.T) {
	t.Run("cancellation", func(t *testing.T) {
		fixture := newGatewayProviderAttemptExecutorFixture(t, 10)
		fixture.bridge.reset(gatewayProviderAttemptBridgeResult{err: context.Canceled})
		recorder := fixture.invoke(t, gatewayProviderAttemptUnaryRoutes(fixture.server)[0], "request-provider-canceled", true, false)
		if recorder.Code != http.StatusRequestTimeout || len(fixture.bridge.callOptions()) != 1 {
			t.Fatalf("canceled attempt crossed fallback boundary: status=%d calls=%d body=%s", recorder.Code, len(fixture.bridge.callOptions()), recorder.Body.String())
		}
		record := fixture.readHistory(t, "request-provider-canceled")
		if record.Status != GatewayRequestStatusCanceled || record.ProviderAttemptCount != 1 || record.FallbackUsed ||
			record.ProviderAttempts[0].Status != GatewayProviderAttemptStatusOutcomeUnknown ||
			record.ProviderAttempts[0].Failure != nil || record.ProviderAttempts[0].FailureBoundary != errorBoundaryPythonBridge {
			t.Fatalf("cancellation evidence drifted: %#v", record)
		}
	})

	t.Run("fallback checkpoint", func(t *testing.T) {
		fixture := newGatewayProviderAttemptExecutorFixture(t, 10)
		fixture.history = &gatewayProviderAttemptFallbackPendingFailStore{gatewayRequestStore: fixture.history}
		fixture.server.gatewayRequestHistoryStore = fixture.history
		fixture.bridge.reset(gatewayProviderAttemptFailedResult(
			t, bridge.ProviderFailureTemporarilyUnavailable, bridge.ProviderFallbackEligible,
		))
		recorder := fixture.invoke(t, gatewayProviderAttemptUnaryRoutes(fixture.server)[0], "request-checkpoint-failed", true, false)
		if recorder.Code != http.StatusServiceUnavailable || len(fixture.bridge.callOptions()) != 1 ||
			!strings.Contains(recorder.Body.String(), gatewayProviderAttemptFailureHistoryUnavailable) {
			t.Fatalf("history checkpoint failure reached backup: status=%d calls=%d body=%s", recorder.Code, len(fixture.bridge.callOptions()), recorder.Body.String())
		}
	})

	t.Run("success terminal checkpoint", func(t *testing.T) {
		fixture := newGatewayProviderAttemptExecutorFixture(t, 10)
		fixture.history = &gatewayProviderAttemptSuccessTerminalFailStore{gatewayRequestStore: fixture.history}
		fixture.server.gatewayRequestHistoryStore = fixture.history
		fixture.bridge.reset(gatewayProviderAttemptSucceededResult("success survives terminal history failure"))
		recorder := fixture.invoke(t, gatewayProviderAttemptUnaryRoutes(fixture.server)[1], "request-success-terminal-failed", true, false)
		if recorder.Code != http.StatusOK || len(fixture.bridge.callOptions()) != 1 ||
			recorder.Header().Get("X-RadishMind-Provider-Attempts") != "1" {
			t.Fatalf("terminal history failure rewrote Provider success: status=%d calls=%d body=%s",
				recorder.Code, len(fixture.bridge.callOptions()), recorder.Body.String())
		}
		record := fixture.readHistory(t, "request-success-terminal-failed")
		if record.Status != GatewayRequestStatusStarted || record.ProviderAttemptPhase != GatewayProviderAttemptPhasePrimaryRunning ||
			record.ProviderAttempts[0].Status != GatewayProviderAttemptStatusRunning {
			t.Fatalf("failed terminal checkpoint fabricated durable success: %#v", record)
		}
	})
}

func TestGatewayProviderAttemptRejectsStreamAndDisabledGateOptIn(t *testing.T) {
	fixture := newGatewayProviderAttemptExecutorFixture(t, 10)
	test := gatewayProviderAttemptUnaryRoutes(fixture.server)[0]
	recorder := fixture.invoke(t, test, "request-stream-fallback", true, true)
	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), gatewayProviderFallbackFailureStreamUnsupported) ||
		len(fixture.bridge.callOptions()) != 0 || fixture.snapshotProvider.reads != 0 {
		t.Fatalf("stream fallback was not rejected before Provider: status=%d reads=%d body=%s", recorder.Code, fixture.snapshotProvider.reads, recorder.Body.String())
	}

	fixture.server.config.GatewayProviderFallbackDevEnabled = false
	recorder = fixture.invoke(t, test, "request-gate-disabled", true, false)
	if recorder.Code != http.StatusForbidden || !strings.Contains(recorder.Body.String(), gatewayProviderFallbackFailureDisabled) ||
		len(fixture.bridge.callOptions()) != 0 {
		t.Fatalf("disabled fallback gate reached Provider: status=%d body=%s", recorder.Code, recorder.Body.String())
	}

	fixture.server.config.GatewayProviderFallbackDevEnabled = true
	invalidRequest := httptest.NewRequest(http.MethodPost, test.path, strings.NewReader(
		`{"model":"mock-model","messages":[{"role":"user","content":"hello"}],"radishmind":{"fallback_mode":"allow_configured","provider_targets":["secondary"]}}`,
	))
	invalidRequest.Header.Set("Authorization", "Bearer "+fixture.token)
	invalidRequest.Header.Set("X-Request-Id", "request-client-target-forbidden")
	invalidRecorder := httptest.NewRecorder()
	test.handle(invalidRecorder, invalidRequest)
	if invalidRecorder.Code != http.StatusBadRequest || !strings.Contains(invalidRecorder.Body.String(), "INVALID_JSON") ||
		len(fixture.bridge.callOptions()) != 0 {
		t.Fatalf("client target override was not rejected: status=%d body=%s", invalidRecorder.Code, invalidRecorder.Body.String())
	}
}

func TestGatewayProviderAttemptRejectsInvalidModeAndBackupInventoryDriftBeforeProvider(t *testing.T) {
	fixture := newGatewayProviderAttemptExecutorFixture(t, 10)
	test := gatewayProviderAttemptUnaryRoutes(fixture.server)[0]
	invalidMode := httptest.NewRequest(http.MethodPost, test.path, strings.NewReader(
		`{"model":"mock-model","messages":[{"role":"user","content":"hello"}],"radishmind":{"fallback_mode":"automatic"}}`,
	))
	invalidMode.Header.Set("Authorization", "Bearer "+fixture.token)
	invalidMode.Header.Set("X-Request-Id", "request-invalid-fallback-mode")
	invalidRecorder := httptest.NewRecorder()
	test.handle(invalidRecorder, invalidMode)
	if invalidRecorder.Code != http.StatusBadRequest ||
		!strings.Contains(invalidRecorder.Body.String(), gatewayProviderFallbackFailureModeInvalid) ||
		len(fixture.bridge.callOptions()) != 0 || fixture.snapshotProvider.reads != 0 {
		t.Fatalf("invalid fallback mode crossed plan boundary: status=%d reads=%d body=%s",
			invalidRecorder.Code, fixture.snapshotProvider.reads, invalidRecorder.Body.String())
	}

	fixture.bridge.inventory.Profiles[1].ResolvedModel = "drifted-backup-model"
	driftRecorder := fixture.invoke(t, test, "request-backup-inventory-drift", true, false)
	if driftRecorder.Code != http.StatusServiceUnavailable ||
		!strings.Contains(driftRecorder.Body.String(), gatewayProviderRouteFailureInventoryMismatch) ||
		len(fixture.bridge.callOptions()) != 0 {
		t.Fatalf("backup inventory drift reached primary Provider: status=%d body=%s",
			driftRecorder.Code, driftRecorder.Body.String())
	}
}

type gatewayProviderAttemptUnaryRoute struct {
	name     string
	path     string
	protocol string
	body     string
	handle   func(http.ResponseWriter, *http.Request)
}

func gatewayProviderAttemptUnaryRoutes(server *Server) []gatewayProviderAttemptUnaryRoute {
	return []gatewayProviderAttemptUnaryRoute{
		{name: "chat", path: "/v1/chat/completions", protocol: northboundProtocolChatCompletions,
			body: `{"model":"mock-model","messages":[{"role":"user","content":"hello"}]}`, handle: server.handleChatCompletions},
		{name: "responses", path: "/v1/responses", protocol: northboundProtocolResponses,
			body: `{"model":"mock-model","input":"hello"}`, handle: server.handleResponses},
		{name: "messages", path: "/v1/messages", protocol: northboundProtocolMessages,
			body: `{"model":"mock-model","max_tokens":16,"messages":[{"role":"user","content":"hello"}]}`, handle: server.handleMessages},
	}
}

func newGatewayProviderAttemptExecutorFixture(t *testing.T, quotaLimit int64) *gatewayProviderAttemptExecutorFixture {
	t.Helper()
	applicationRepository := newMemoryApplicationCatalogRepository()
	keyRepository := newMemoryAPIKeyRepository()
	managementContext := apiKeyTestContext("subject-owner")
	const applicationID = "app_aaaaaaaaaaaaaaaa"
	seedAPIKeyTestApplication(t, applicationRepository, managementContext, applicationID, applicationCatalogLifecycleActive)
	keyService := newAPIKeyService(keyRepository, applicationRepository)
	keyService.newID = func() (string, error) { return "key_aaaaaaaaaaaaaaaa", nil }
	issued := keyService.Create(managementContext, APIKeyCreateInput{
		ApplicationID: applicationID, DisplayName: "Provider fallback key",
		Scopes: []string{"chat:invoke", "responses:invoke", "messages:invoke"}, ExpiresInDays: 30,
	})
	if issued.FailureCode != "" || issued.CredentialToken == "" {
		t.Fatalf("issue Provider fallback API key: %#v", issued)
	}

	primary := adminProviderRouteBridgeProfile()
	primary.ResolvedModel = "mock-model-primary"
	secondary := primary
	secondary.Profile = "mock-secondary"
	secondary.NormalizedProfile = "mock-secondary"
	secondary.ResolvedModel = "mock-model-secondary"
	snapshot := gatewayProviderAttemptExecutorSnapshot(t, primary, secondary)
	snapshotProvider := &mutableGatewayProviderRouteSnapshotProvider{snapshot: snapshot, found: true}
	testBridge := &gatewayProviderAttemptScriptedBridge{inventory: bridge.ProviderInventory{
		Profiles: []bridge.ProviderProfileDescription{primary, secondary},
	}}
	history := newMemoryGatewayRequestStore(30)
	quota := newMemoryGatewayRequestQuotaRepository()
	quotaContext := GatewayRequestQuotaContext{
		RequestContext: context.Background(), TenantRef: managementContext.TenantRef,
		WorkspaceID: managementContext.WorkspaceID, Environment: "test", ApplicationID: applicationID,
		ActorRef: managementContext.OwnerSubjectRef, RequestID: "request-quota-policy", AuditRef: "audit-quota-policy",
	}
	if _, err := quota.PutPolicy(quotaContext, 0, quotaLimit, time.Now().UTC()); err != nil {
		t.Fatalf("put Provider fallback quota: %v", err)
	}
	pricing := newMemoryGatewayModelPricingRepository()
	server := &Server{
		bridge: testBridge,
		config: config.Config{
			BridgeTimeout: time.Second, GatewayAuthMode: gatewayAPIKeyAuthenticationSource,
			GatewayRequestHistoryDevEnabled: true, GatewayRequestDatabaseTimeout: time.Second,
			GatewayProviderRouteSource: "admin_snapshot_dev_test", GatewayProviderRouteEnvironment: "test",
			GatewayProviderRouteConfigurationID: "gateway-default", GatewayProviderFallbackDevEnabled: true,
			GatewayRequestQuotaEnforcementDevEnabled: true, GatewayRequestQuotaEnvironment: "test",
			GatewayModelPricingCaptureDevEnabled: true, GatewayModelPricingEnvironment: "test",
		},
		applicationCatalogRepository: applicationRepository, apiKeyRepository: keyRepository,
		providerRouteSnapshotProvider: snapshotProvider,
		gatewayRequestHistoryStore:    history, gatewayRequestHistoryStoreMode: gatewayRequestStoreModeMemoryDev,
		gatewayRequestQuotaRepository: quota, gatewayModelPricingRepository: pricing,
	}
	historyContext := GatewayRequestContext{
		TenantRef: managementContext.TenantRef, WorkspaceID: managementContext.WorkspaceID,
		ConsumerRef: "api_key:" + issued.Record.APIKeyID, ApplicationID: applicationID,
		SubjectRef: managementContext.OwnerSubjectRef, AuditContext: "api-key-dev-test", Source: gatewayAPIKeyAuthenticationSource,
	}
	return &gatewayProviderAttemptExecutorFixture{
		server: server, token: issued.CredentialToken, history: history, quota: quota, pricing: pricing,
		bridge: testBridge, snapshotProvider: snapshotProvider, historyContext: historyContext, quotaContext: quotaContext,
	}
}

func gatewayProviderAttemptExecutorSnapshot(
	t *testing.T,
	primary bridge.ProviderProfileDescription,
	secondary bridge.ProviderProfileDescription,
) AdminProviderRouteSnapshot {
	t.Helper()
	resolver := newFakeAdminProviderInventoryResolver()
	for profileID, profile := range map[string]bridge.ProviderProfileDescription{"primary": primary, "secondary": secondary} {
		binding, err := adminProviderRouteInventoryBindingFromProfile(
			"test", "mock", "ref:radishmind/test/provider-profiles/"+profile.Profile, profile,
		)
		if err != nil {
			t.Fatalf("build %s inventory binding: %v", profileID, err)
		}
		resolver.put(binding)
	}
	repository := newMemoryAdminProviderRouteRepository()
	service := newAdminProviderRouteService(repository, resolver)
	input := adminProviderRouteV2TestDraftInput(0)
	input.ConfigurationID = "gateway-default"
	input.ModelRoutes = []AdminModelRouteDefinition{
		gatewayProviderAttemptExecutorRoute("chat", "chat_completions"),
		gatewayProviderAttemptExecutorRoute("responses", "responses"),
		gatewayProviderAttemptExecutorRoute("messages", "messages"),
	}
	ctx := adminProviderRouteTestContext()
	ctx.TenantRef = "tenant_demo"
	ctx.WorkspaceID = "workspace_demo"
	draft := service.PutDraft(ctx, input)
	if draft.FailureCode != "" {
		t.Fatalf("create executor Route v2 draft: %#v", draft)
	}
	if result := service.CreateCandidate(ctx, AdminProviderRouteCandidateInput{
		ConfigurationID: input.ConfigurationID, CandidateID: "candidate-executor", ExpectedDraftRevision: 1,
	}); result.FailureCode != "" {
		t.Fatalf("create executor Route v2 candidate: %#v", result)
	}
	if result := service.ReviewCandidate(ctx, input.ConfigurationID, "candidate-executor", AdminProviderRouteReviewInput{
		ExpectedReviewVersion: 0, Decision: "approve", Reason: "Reviewed all unary fallback targets.",
	}); result.FailureCode != "" {
		t.Fatalf("review executor Route v2 candidate: %#v", result)
	}
	result := service.Activate(ctx, AdminProviderRouteActivationInput{
		ConfigurationID: input.ConfigurationID, CandidateID: "candidate-executor", ExpectedGeneration: 0,
		Action: "activate", Reason: "Activate all unary fallback routes.",
	})
	if result.FailureCode != "" || result.Snapshot == nil {
		t.Fatalf("activate executor Route v2 candidate: %#v", result)
	}
	return *result.Snapshot
}

func gatewayProviderAttemptExecutorRoute(routeID, protocol string) AdminModelRouteDefinition {
	return AdminModelRouteDefinition{
		RouteID: routeID, Protocol: protocol, ModelID: "mock-model",
		ExecutionMode: AdminProviderRouteExecutionSequentialFallback,
		AttemptTargets: []AdminProviderRouteAttemptTarget{
			{Ordinal: 1, ProviderProfileID: "primary"},
			{Ordinal: 2, ProviderProfileID: "secondary"},
		},
	}
}

func (fixture *gatewayProviderAttemptExecutorFixture) invoke(
	t *testing.T,
	route gatewayProviderAttemptUnaryRoute,
	requestID string,
	allowFallback bool,
	stream bool,
) *httptest.ResponseRecorder {
	t.Helper()
	body := strings.TrimSuffix(route.body, "}")
	if allowFallback {
		body += `,"radishmind":{"fallback_mode":"allow_configured"}`
	}
	if stream {
		body += `,"stream":true`
	}
	body += "}"
	request := httptest.NewRequest(http.MethodPost, route.path, strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer "+fixture.token)
	request.Header.Set("X-Request-Id", requestID)
	recorder := httptest.NewRecorder()
	route.handle(recorder, request)
	return recorder
}

func (fixture *gatewayProviderAttemptExecutorFixture) readHistory(t *testing.T, requestID string) GatewayRequestRecord {
	t.Helper()
	record, found, err := fixture.history.ReadRequest(fixture.historyContext, requestID)
	if err != nil || !found {
		t.Fatalf("read Provider attempt history %s: found=%t err=%v", requestID, found, err)
	}
	return record
}

func gatewayProviderAttemptFailedResult(
	t *testing.T,
	failureClass bridge.ProviderAttemptFailureClass,
	disposition bridge.ProviderAttemptFallbackDisposition,
) gatewayProviderAttemptBridgeResult {
	t.Helper()
	failure, ok := bridge.NewProviderAttemptFailure(
		failureClass, disposition, false, bridge.ProviderAttemptFailed,
		"PROVIDER_TYPED_FAILURE", "5xx",
	)
	if !ok {
		t.Fatal("build typed Provider failure")
	}
	return gatewayProviderAttemptBridgeResult{envelope: bridge.GatewayEnvelope{
		SchemaVersion: 1, Status: "failed", RequestID: "bridge-failed",
		Response:               map[string]any{"status": "failed"},
		Error:                  &bridge.GatewayError{Code: "GATEWAY_INFERENCE_FAILED", Message: "gateway inference failed"},
		ProviderAttemptFailure: &failure, Metadata: map[string]any{},
	}}
}

func gatewayProviderAttemptSucceededResult(summary string) gatewayProviderAttemptBridgeResult {
	return gatewayProviderAttemptBridgeResult{envelope: bridge.GatewayEnvelope{
		SchemaVersion: 1, Status: "ok", RequestID: "bridge-succeeded",
		Response: map[string]any{"summary": summary}, Metadata: map[string]any{},
	}}
}

func seedGatewayProviderAttemptPricing(
	t *testing.T,
	fixture *gatewayProviderAttemptExecutorFixture,
	profileID string,
	expectedVersion int64,
) {
	t.Helper()
	ctx := GatewayModelPricingContext{
		RequestContext: context.Background(), TenantRef: fixture.historyContext.TenantRef,
		WorkspaceID: fixture.historyContext.WorkspaceID, Environment: "test", ProviderID: "mock",
		ProfileID: profileID, ModelID: "mock-model", ActorRef: fixture.historyContext.SubjectRef,
		RequestID: "request-pricing-" + profileID, AuditRef: "audit-pricing-" + profileID,
	}
	result := newGatewayModelPricingService(fixture.pricing).PutRevision(ctx, GatewayModelPricingPolicyInput{
		ExpectedVersion: expectedVersion - 1,
		Currency:        GatewayModelPricingCurrency, InputPriceMicrosPerTokenUnit: 1000 * expectedVersion,
		OutputPriceMicrosPerTokenUnit: 2000 * expectedVersion, Reason: "Freeze deterministic attempt pricing.",
	})
	if result.FailureCode != "" || result.Policy == nil || result.Policy.RecordVersion != expectedVersion {
		t.Fatalf("seed pricing %s v%d: %#v", profileID, expectedVersion, result)
	}
}

type gatewayProviderAttemptFallbackPendingFailStore struct {
	gatewayRequestStore
}

type gatewayProviderAttemptSuccessTerminalFailStore struct {
	gatewayRequestStore
}

func (store *gatewayProviderAttemptSuccessTerminalFailStore) UpdateRequest(
	ctx GatewayRequestContext,
	record *GatewayRequestRecord,
) error {
	if record != nil && record.ProviderAttemptPhase == GatewayProviderAttemptPhaseTerminalPending &&
		len(record.ProviderAttempts) > 0 &&
		record.ProviderAttempts[len(record.ProviderAttempts)-1].Status == GatewayProviderAttemptStatusSucceeded {
		return errGatewayRequestStoreUnavailable
	}
	return store.gatewayRequestStore.UpdateRequest(ctx, record)
}

func (store *gatewayProviderAttemptFallbackPendingFailStore) UpdateRequest(
	ctx GatewayRequestContext,
	record *GatewayRequestRecord,
) error {
	if record != nil && record.ProviderAttemptPhase == GatewayProviderAttemptPhaseFallbackPending {
		return errGatewayRequestStoreUnavailable
	}
	return store.gatewayRequestStore.UpdateRequest(ctx, record)
}
