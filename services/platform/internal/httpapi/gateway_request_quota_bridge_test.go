package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/bridge"
	"radishmind.local/services/platform/internal/config"
)

func TestGatewayRequestQuotaBridgeAdmitsBeforeProviderAndFailsClosed(t *testing.T) {
	repository := newMemoryGatewayRequestQuotaRepository()
	quotaContext := testGatewayRequestQuotaContext("app-bridge")
	now := time.Date(2026, 8, 9, 15, 0, 0, 0, time.UTC)
	if _, err := repository.PutPolicy(quotaContext, 0, 1, now); err != nil {
		t.Fatalf("put bridge quota policy: %v", err)
	}
	inner := &fakeBridge{envelope: bridge.GatewayEnvelope{Status: "ok", Response: map[string]any{"summary": "ok"}}}
	client := newGatewayRequestQuotaBridgeClient(inner, repository)
	client.now = func() time.Time { return now }

	firstContext := gatewayRequestQuotaBridgeTestContext(quotaContext, "request-bridge-one")
	if _, err := client.HandleEnvelope(firstContext, []byte(`{"task":"test"}`), bridge.EnvelopeOptions{}); err != nil {
		t.Fatalf("first quota bridge call: %v", err)
	}
	if inner.handleCalls != 1 {
		t.Fatalf("expected one provider call, got %d", inner.handleCalls)
	}
	secondContext := gatewayRequestQuotaBridgeTestContext(quotaContext, "request-bridge-two")
	if _, err := client.HandleEnvelope(secondContext, []byte(`{"task":"test"}`), bridge.EnvelopeOptions{}); gatewayRequestQuotaFailureCode(err) != GatewayRequestQuotaFailureExceeded {
		t.Fatalf("expected stable quota exceeded failure, got %v", err)
	}
	if inner.handleCalls != 1 {
		t.Fatalf("quota rejection reached provider: %d calls", inner.handleCalls)
	}
	usage, found, err := repository.ReadUsage(quotaContext, "2026-08-09")
	if err != nil || !found || usage.AdmittedRequestCount != 1 {
		t.Fatalf("unexpected bridge quota usage: found=%t usage=%+v err=%v", found, usage, err)
	}
}

func TestGatewayRequestQuotaBridgeProtectsStreamingProviderAttempt(t *testing.T) {
	repository := newMemoryGatewayRequestQuotaRepository()
	quotaContext := testGatewayRequestQuotaContext("app-stream")
	now := time.Date(2026, 8, 9, 15, 30, 0, 0, time.UTC)
	if _, err := repository.PutPolicy(quotaContext, 0, 1, now); err != nil {
		t.Fatalf("put streaming quota policy: %v", err)
	}
	inner := &fakeBridge{envelope: bridge.GatewayEnvelope{Status: "ok", Response: map[string]any{"summary": "streamed"}}}
	client := newGatewayRequestQuotaBridgeClient(inner, repository)
	client.now = func() time.Time { return now }
	if err := client.StreamEnvelope(
		gatewayRequestQuotaBridgeTestContext(quotaContext, "request-stream-one"), nil, bridge.EnvelopeOptions{}, nil,
	); err != nil || !inner.streamCalled {
		t.Fatalf("first streaming attempt was not admitted: called=%t err=%v", inner.streamCalled, err)
	}
	inner.streamCalled = false
	err := client.StreamEnvelope(
		gatewayRequestQuotaBridgeTestContext(quotaContext, "request-stream-two"), nil, bridge.EnvelopeOptions{}, nil,
	)
	if gatewayRequestQuotaFailureCode(err) != GatewayRequestQuotaFailureExceeded || inner.streamCalled {
		t.Fatalf("over-limit streaming attempt crossed provider: called=%t err=%v", inner.streamCalled, err)
	}
}

func TestGatewayRequestQuotaProtectsThreeStandardAPIKeyInferenceRoutes(t *testing.T) {
	applicationRepository := newMemoryApplicationCatalogRepository()
	apiKeyRepository := newMemoryAPIKeyRepository()
	managementContext := apiKeyTestContext("subject-quota-owner")
	seedAPIKeyTestApplication(t, applicationRepository, managementContext, "app_aaaaaaaaaaaaaaaa", applicationCatalogLifecycleActive)
	apiKeyService := newAPIKeyService(apiKeyRepository, applicationRepository)
	apiKeyService.newID = func() (string, error) { return "key_qqqqqqqqqqqqqqqq", nil }
	issued := apiKeyService.Create(managementContext, APIKeyCreateInput{
		ApplicationID: "app_aaaaaaaaaaaaaaaa", DisplayName: "Quota route key",
		Scopes: []string{"chat:invoke", "responses:invoke", "messages:invoke"}, ExpiresInDays: 30,
	})
	if issued.FailureCode != "" {
		t.Fatalf("issue quota route API key: %+v", issued)
	}
	quotaRepository := newMemoryGatewayRequestQuotaRepository()
	quotaContext := GatewayRequestQuotaContext{
		RequestContext: context.Background(), TenantRef: managementContext.TenantRef,
		WorkspaceID: managementContext.WorkspaceID, Environment: "test", ApplicationID: issued.Record.ApplicationID,
		ActorRef: managementContext.OwnerSubjectRef, RequestID: "request-quota-route-policy", AuditRef: "audit-quota-route-policy",
	}
	if _, err := quotaRepository.PutPolicy(quotaContext, 0, 3, time.Now().UTC()); err != nil {
		t.Fatalf("put route quota policy: %v", err)
	}
	inner := &fakeBridge{envelope: bridge.GatewayEnvelope{Status: "ok", Response: map[string]any{"summary": "quota route response"}}}
	server := &Server{
		config: config.Config{
			GatewayAuthMode: gatewayAPIKeyAuthenticationSource, GatewayRequestHistoryDevEnabled: true,
			GatewayRequestQuotaEnforcementDevEnabled: true, GatewayRequestQuotaEnvironment: "test",
			GatewayRequestDatabaseTimeout: time.Second, BridgeTimeout: time.Second,
		},
		bridge:                       newGatewayRequestQuotaBridgeClient(inner, quotaRepository),
		applicationCatalogRepository: applicationRepository, apiKeyRepository: apiKeyRepository,
		gatewayRequestHistoryStore: newMemoryGatewayRequestStore(20), gatewayRequestHistoryStoreMode: gatewayRequestStoreModeMemoryDev,
	}
	routes := []struct {
		path   string
		body   string
		handle func(http.ResponseWriter, *http.Request)
	}{
		{path: "/v1/chat/completions", body: `{"model":"mock","messages":[{"role":"user","content":"hello"}]}`, handle: server.handleChatCompletions},
		{path: "/v1/responses", body: `{"model":"mock","input":"hello"}`, handle: server.handleResponses},
		{path: "/v1/messages", body: `{"model":"mock","max_tokens":16,"messages":[{"role":"user","content":"hello"}]}`, handle: server.handleMessages},
	}
	for index, route := range routes {
		request := httptest.NewRequest(http.MethodPost, route.path, strings.NewReader(route.body))
		request.Header.Set("Authorization", "Bearer "+issued.CredentialToken)
		request.Header.Set("X-Request-Id", "request-quota-route-"+string(rune('a'+index)))
		response := httptest.NewRecorder()
		route.handle(response, request)
		if response.Code != http.StatusOK {
			t.Fatalf("%s quota admission failed: status=%d body=%s", route.path, response.Code, response.Body.String())
		}
	}
	if inner.handleCalls != 3 {
		t.Fatalf("three standard routes produced %d provider calls", inner.handleCalls)
	}
	overLimit := httptest.NewRequest(http.MethodPost, "/v1/responses", strings.NewReader(`{"model":"mock","input":"over"}`))
	overLimit.Header.Set("Authorization", "Bearer "+issued.CredentialToken)
	overLimit.Header.Set("X-Request-Id", "request-quota-route-over")
	overLimitResponse := httptest.NewRecorder()
	server.handleResponses(overLimitResponse, overLimit)
	if overLimitResponse.Code != http.StatusTooManyRequests ||
		!strings.Contains(overLimitResponse.Body.String(), GatewayRequestQuotaFailureExceeded) || inner.handleCalls != 3 {
		t.Fatalf("over-limit standard route crossed provider: status=%d calls=%d body=%s", overLimitResponse.Code, inner.handleCalls, overLimitResponse.Body.String())
	}
}

func TestGatewayRequestQuotaBridgeConsumesAdmittedProviderFailure(t *testing.T) {
	repository := newMemoryGatewayRequestQuotaRepository()
	quotaContext := testGatewayRequestQuotaContext("app-provider-failure")
	now := time.Date(2026, 8, 9, 15, 0, 0, 0, time.UTC)
	if _, err := repository.PutPolicy(quotaContext, 0, 2, now); err != nil {
		t.Fatalf("put bridge quota policy: %v", err)
	}
	providerFailure := errors.New("provider failed")
	inner := &fakeBridge{handleErr: providerFailure}
	client := newGatewayRequestQuotaBridgeClient(inner, repository)
	client.now = func() time.Time { return now }
	_, err := client.HandleEnvelope(gatewayRequestQuotaBridgeTestContext(quotaContext, "request-provider-failure"), nil, bridge.EnvelopeOptions{})
	if !errors.Is(err, providerFailure) || inner.handleCalls != 1 {
		t.Fatalf("provider failure was not delegated after admission: calls=%d err=%v", inner.handleCalls, err)
	}
	usage, found, err := repository.ReadUsage(quotaContext, "2026-08-09")
	if err != nil || !found || usage.AdmittedRequestCount != 1 || usage.RemainingRequestCount != 1 {
		t.Fatalf("provider failure refunded admission: found=%t usage=%+v err=%v", found, usage, err)
	}
	if _, err := client.DescribeInventory(context.Background()); err != nil {
		t.Fatalf("inventory delegation failed: %v", err)
	}
	if usageAfter, _, _ := repository.ReadUsage(quotaContext, "2026-08-09"); usageAfter.AdmittedRequestCount != 1 {
		t.Fatalf("inventory read consumed quota: %+v", usageAfter)
	}
}

func gatewayRequestQuotaBridgeTestContext(quotaContext GatewayRequestQuotaContext, requestID string) context.Context {
	return gatewayRequestQuotaBridgeTestContextForRoute(quotaContext, requestID, "POST /v1/responses")
}

func gatewayRequestQuotaBridgeTestContextForRoute(
	quotaContext GatewayRequestQuotaContext,
	requestID string,
	route string,
) context.Context {
	base := context.Background()
	quotaContext.RequestContext = base
	return withGatewayRequestQuotaBinding(base, gatewayRequestQuotaBinding{
		QuotaContext: quotaContext, APIKeyID: "key-bridge", RequestID: requestID, Route: route,
	})
}
