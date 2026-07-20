package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/bridge"
	"radishmind.local/services/platform/internal/config"
)

type apiKeyGatewayTestBridge struct {
	fakeBridge
	inventoryCalls atomic.Int32
}

func (testBridge *apiKeyGatewayTestBridge) DescribeInventory(context.Context) (bridge.ProviderInventory, error) {
	testBridge.inventoryCalls.Add(1)
	return testBridge.inventory, nil
}

func TestGatewayAPIKeyAuthenticationProtectsNorthboundAndWritesHistory(t *testing.T) {
	runGatewayAPIKeyAuthenticationProtectsNorthboundAndWritesHistory(t, newMemoryAPIKeyRepository(), newMemoryApplicationCatalogRepository())
}

func runGatewayAPIKeyAuthenticationProtectsNorthboundAndWritesHistory(t *testing.T, keyRepository apiKeyRepository, applicationRepository applicationCatalogRepository) {
	runGatewayAPIKeyAuthenticationProtectsNorthboundWithHistoryStore(
		t,
		keyRepository,
		applicationRepository,
		newMemoryGatewayRequestStore(20),
		gatewayRequestStoreModeMemoryDev,
	)
}

func runGatewayAPIKeyAuthenticationProtectsNorthboundWithHistoryStore(
	t *testing.T,
	keyRepository apiKeyRepository,
	applicationRepository applicationCatalogRepository,
	historyStore gatewayRequestStore,
	historyStoreMode string,
) {
	t.Helper()
	managementContext := apiKeyTestContext("subject_owner")
	seedAPIKeyTestApplication(t, applicationRepository, managementContext, "app_aaaaaaaaaaaaaaaa", applicationCatalogLifecycleActive)
	service := newAPIKeyService(keyRepository, applicationRepository)
	service.newID = func() (string, error) { return "key_aaaaaaaaaaaaaaaa", nil }
	issued := service.Create(managementContext, APIKeyCreateInput{
		ApplicationID: "app_aaaaaaaaaaaaaaaa", DisplayName: "Gateway SDK key",
		Scopes: []string{"models:read", "chat:invoke", "responses:invoke", "messages:invoke"}, ExpiresInDays: 30,
	})
	if issued.FailureCode != "" || issued.CredentialToken == "" {
		t.Fatalf("issue Gateway API key: failure=%s record_present=%t credential_present=%t", issued.FailureCode, issued.Record != nil, issued.CredentialToken != "")
	}

	testBridge := &apiKeyGatewayTestBridge{}
	server := &Server{
		config: config.Config{
			GatewayAuthMode: gatewayAPIKeyAuthenticationSource, GatewayRequestHistoryDevEnabled: true,
			GatewayRequestDatabaseTimeout: time.Second, BridgeTimeout: time.Second,
		},
		bridge: testBridge, applicationCatalogRepository: applicationRepository,
		apiKeyRepository: keyRepository, gatewayRequestHistoryStore: historyStore,
		gatewayRequestHistoryStoreMode: historyStoreMode,
	}
	for _, scope := range []string{"models:read", "chat:invoke", "responses:invoke", "messages:invoke"} {
		request := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
		request.Header.Set("Authorization", "Bearer "+issued.CredentialToken)
		result := server.authenticateGatewayAPIKey(request, newRequestTrace(request, "/v1/models"), scope)
		if result.FailureCode != "" || result.RequestContext.Source != gatewayAPIKeyAuthenticationSource ||
			result.RequestContext.ConsumerRef != "api_key:"+issued.Record.APIKeyID || result.RequestContext.TenantRef != managementContext.TenantRef ||
			result.RequestContext.WorkspaceID != managementContext.WorkspaceID || result.RequestContext.ApplicationID != issued.Record.ApplicationID ||
			result.RequestContext.SubjectRef != managementContext.OwnerSubjectRef {
			t.Fatalf("scope %s did not restore trusted API key context: %#v", scope, result)
		}
	}

	request := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	request.Header.Set("Authorization", "Bearer "+issued.CredentialToken)
	request.Header.Set("X-Request-Id", "request_gateway_success")
	recorder := httptest.NewRecorder()
	handleModels(recorder, request, server)
	if recorder.Code != http.StatusOK || testBridge.inventoryCalls.Load() != 1 {
		t.Fatalf("valid API key did not reach model inventory: status=%d calls=%d body=%s", recorder.Code, testBridge.inventoryCalls.Load(), recorder.Body.String())
	}
	historyContext := GatewayRequestContext{
		TenantRef: managementContext.TenantRef, WorkspaceID: managementContext.WorkspaceID,
		ConsumerRef: "api_key:" + issued.Record.APIKeyID, ApplicationID: issued.Record.ApplicationID,
		SubjectRef: managementContext.OwnerSubjectRef, AuditContext: "api-key-dev-test", Source: gatewayAPIKeyAuthenticationSource,
	}
	record, found, err := historyStore.ReadRequest(historyContext, "request_gateway_success")
	if err != nil || !found || record.Status != GatewayRequestStatusSucceeded || record.Protocol != northboundProtocolModels ||
		record.ConsumerRef != historyContext.ConsumerRef || record.ApplicationID != issued.Record.ApplicationID {
		t.Fatalf("successful API key call did not write sanitized history: found=%v record=%#v err=%v", found, record, err)
	}
	stored, err := keyRepository.Read(managementContext, issued.Record.APIKeyID)
	if err != nil || stored.LastUsedAt == nil {
		t.Fatalf("successful authentication did not update last_used_at: record=%#v err=%v", stored, err)
	}
	if strings.Contains(recorder.Body.String(), issued.CredentialToken) || strings.Contains(recorder.Body.String(), "credential_digest") {
		t.Fatalf("northbound response leaked credential material: %s", recorder.Body.String())
	}
}

func TestGatewayAPIKeyAuthenticationFailuresStopBeforeInventoryAndHistory(t *testing.T) {
	applicationRepository := newMemoryApplicationCatalogRepository()
	managementContext := apiKeyTestContext("subject_owner")
	seedAPIKeyTestApplication(t, applicationRepository, managementContext, "app_aaaaaaaaaaaaaaaa", applicationCatalogLifecycleActive)
	apiKeyRepository := newMemoryAPIKeyRepository()
	service := newAPIKeyService(apiKeyRepository, applicationRepository)
	service.newID = func() (string, error) { return "key_bbbbbbbbbbbbbbbb", nil }
	issued := service.Create(managementContext, validAPIKeyCreateInput("app_aaaaaaaaaaaaaaaa", 30))
	if issued.FailureCode != "" {
		t.Fatalf("issue restricted API key: failure=%s credential_present=%t", issued.FailureCode, issued.CredentialToken != "")
	}
	testBridge := &apiKeyGatewayTestBridge{}
	historyStore := newMemoryGatewayRequestStore(20)
	server := &Server{
		config: config.Config{GatewayAuthMode: gatewayAPIKeyAuthenticationSource, GatewayRequestHistoryDevEnabled: true, BridgeTimeout: time.Second},
		bridge: testBridge, applicationCatalogRepository: applicationRepository, apiKeyRepository: apiKeyRepository,
		gatewayRequestHistoryStore: historyStore, gatewayRequestHistoryStoreMode: gatewayRequestStoreModeMemoryDev,
	}

	tests := []struct {
		name       string
		token      string
		devHeader  bool
		expected   string
		statusCode int
	}{
		{name: "missing", expected: APIKeyFailureMissing, statusCode: http.StatusUnauthorized},
		{name: "malformed", token: "invalid", expected: APIKeyFailureInvalid, statusCode: http.StatusUnauthorized},
		{name: "digest mismatch", token: strings.TrimSuffix(issued.CredentialToken, "A") + "B", expected: APIKeyFailureInvalid, statusCode: http.StatusUnauthorized},
		{name: "credential conflict", token: issued.CredentialToken, devHeader: true, expected: APIKeyFailureCredentialConflict, statusCode: http.StatusBadRequest},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
			if test.token != "" {
				request.Header.Set("Authorization", "Bearer "+test.token)
			}
			if test.devHeader {
				request.Header.Set(gatewayRequestDevTenantHeader, managementContext.TenantRef)
			}
			recorder := httptest.NewRecorder()
			handleModels(recorder, request, server)
			if recorder.Code != test.statusCode || !strings.Contains(recorder.Body.String(), test.expected) {
				t.Fatalf("unexpected failure: status=%d body=%s", recorder.Code, recorder.Body.String())
			}
		})
	}
	if testBridge.inventoryCalls.Load() != 0 || len(historyStore.records) != 0 {
		t.Fatalf("authentication failures crossed side-effect boundary: inventory=%d history=%d", testBridge.inventoryCalls.Load(), len(historyStore.records))
	}

	for _, route := range []struct {
		path   string
		body   string
		handle func(http.ResponseWriter, *http.Request)
	}{
		{path: "/v1/chat/completions", body: `{"model":"mock","messages":[{"role":"user","content":"hello"}]}`, handle: server.handleChatCompletions},
		{path: "/v1/responses", body: `{"model":"mock","input":"hello"}`, handle: server.handleResponses},
		{path: "/v1/messages", body: `{"model":"mock","max_tokens":16,"messages":[{"role":"user","content":"hello"}]}`, handle: server.handleMessages},
	} {
		scopeRequest := httptest.NewRequest(http.MethodPost, route.path, strings.NewReader(route.body))
		scopeRequest.Header.Set("Authorization", "Bearer "+issued.CredentialToken)
		scopeRecorder := httptest.NewRecorder()
		route.handle(scopeRecorder, scopeRequest)
		if scopeRecorder.Code != http.StatusForbidden || !strings.Contains(scopeRecorder.Body.String(), APIKeyFailureScopeDenied) || len(historyStore.records) != 0 {
			t.Fatalf("%s scope denial crossed history boundary: status=%d body=%s history=%d", route.path, scopeRecorder.Code, scopeRecorder.Body.String(), len(historyStore.records))
		}
	}

	revoked := service.Revoke(managementContext, issued.Record.APIKeyID, 1)
	if revoked.FailureCode != "" {
		t.Fatalf("revoke restricted API key: %#v", revoked)
	}
	revokedRequest := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	revokedRequest.Header.Set("Authorization", "Bearer "+issued.CredentialToken)
	revokedRecorder := httptest.NewRecorder()
	handleModels(revokedRecorder, revokedRequest, server)
	if revokedRecorder.Code != http.StatusForbidden || !strings.Contains(revokedRecorder.Body.String(), APIKeyFailureRevoked) || testBridge.inventoryCalls.Load() != 0 {
		t.Fatalf("revoked key crossed inventory boundary: status=%d body=%s", revokedRecorder.Code, revokedRecorder.Body.String())
	}
}

func TestGatewayAPIKeyHistoryAndCredentialStoreFailuresDoNotFallback(t *testing.T) {
	applicationRepository := newMemoryApplicationCatalogRepository()
	managementContext := apiKeyTestContext("subject_owner")
	seedAPIKeyTestApplication(t, applicationRepository, managementContext, "app_aaaaaaaaaaaaaaaa", applicationCatalogLifecycleActive)
	apiKeyRepository := newMemoryAPIKeyRepository()
	service := newAPIKeyService(apiKeyRepository, applicationRepository)
	service.newID = func() (string, error) { return "key_cccccccccccccccc", nil }
	issued := service.Create(managementContext, validAPIKeyCreateInput("app_aaaaaaaaaaaaaaaa", 30))
	if issued.FailureCode != "" {
		t.Fatalf("issue API key: failure=%s credential_present=%t", issued.FailureCode, issued.CredentialToken != "")
	}
	testBridge := &apiKeyGatewayTestBridge{}
	server := &Server{
		config: config.Config{GatewayAuthMode: gatewayAPIKeyAuthenticationSource, GatewayRequestHistoryDevEnabled: true, BridgeTimeout: time.Second},
		bridge: testBridge, applicationCatalogRepository: applicationRepository, apiKeyRepository: apiKeyRepository,
		gatewayRequestHistoryStore: failingGatewayRequestStore{}, gatewayRequestHistoryStoreMode: gatewayRequestStoreModeMemoryDev,
	}
	request := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	request.Header.Set("Authorization", "Bearer "+issued.CredentialToken)
	recorder := httptest.NewRecorder()
	handleModels(recorder, request, server)
	if recorder.Code != http.StatusServiceUnavailable || !strings.Contains(recorder.Body.String(), string(GatewayRequestHistoryFailureStore)) || testBridge.inventoryCalls.Load() != 0 {
		t.Fatalf("history failure did not stop before inventory: status=%d calls=%d body=%s", recorder.Code, testBridge.inventoryCalls.Load(), recorder.Body.String())
	}

	apiKeyRepository.unavailable = true
	server.gatewayRequestHistoryStore = newMemoryGatewayRequestStore(10)
	credentialRequest := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	credentialRequest.Header.Set("Authorization", "Bearer "+issued.CredentialToken)
	credentialRecorder := httptest.NewRecorder()
	handleModels(credentialRecorder, credentialRequest, server)
	if credentialRecorder.Code != http.StatusServiceUnavailable || !strings.Contains(credentialRecorder.Body.String(), APIKeyFailureStoreUnavailable) || testBridge.inventoryCalls.Load() != 0 {
		t.Fatalf("credential store failure fell through: status=%d calls=%d body=%s", credentialRecorder.Code, testBridge.inventoryCalls.Load(), credentialRecorder.Body.String())
	}

	fallbackStore := newMemoryGatewayRequestStore(10)
	server.apiKeyRepository = apiKeyRepository
	apiKeyRepository.unavailable = false
	server.gatewayRequestHistoryStore = fallbackStore
	fallbackRequest := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	fallbackRequest.Header.Set(gatewayRequestDevTenantHeader, managementContext.TenantRef)
	fallbackRequest.Header.Set(gatewayRequestDevWorkspaceHeader, managementContext.WorkspaceID)
	fallbackRequest.Header.Set(gatewayRequestDevConsumerHeader, "consumer_dev_fallback")
	fallbackRequest.Header.Set(gatewayRequestDevApplicationHeader, issued.Record.ApplicationID)
	fallbackRequest.Header.Set(gatewayRequestDevSubjectHeader, managementContext.OwnerSubjectRef)
	fallbackRequest.Header.Set(gatewayRequestDevScopesHeader, "models:read")
	fallbackRequest.Header.Set(gatewayRequestDevAuditHeader, "audit-dev-fallback")
	fallbackTrace := newRequestTrace(fallbackRequest, "/v1/models")
	fallbackTrace.gatewayRequestContext = GatewayRequestContext{Source: gatewayAPIKeyAuthenticationSource}
	if err := server.startGatewayRequestTrace(fallbackRequest, &fallbackTrace, northboundProtocolModels); !errors.Is(err, errGatewayRequestStoreContract) {
		t.Fatalf("invalid trusted API key context did not fail closed: %v", err)
	}
	if len(fallbackStore.records) != 0 {
		t.Fatalf("invalid trusted API key context fell back to development headers: history=%d", len(fallbackStore.records))
	}
}

func TestGatewayAPIKeyExpiryApplicationArchiveAndRevokeRace(t *testing.T) {
	runGatewayAPIKeyExpiryApplicationArchiveAndRevokeRace(t, newMemoryAPIKeyRepository(), newMemoryApplicationCatalogRepository())
}

func runGatewayAPIKeyExpiryApplicationArchiveAndRevokeRace(t *testing.T, keyRepository apiKeyRepository, applicationRepository applicationCatalogRepository) {
	t.Helper()
	managementContext := apiKeyTestContext("subject_owner")
	seedAPIKeyTestApplication(t, applicationRepository, managementContext, "app_aaaaaaaaaaaaaaaa", applicationCatalogLifecycleActive)
	service := newAPIKeyService(keyRepository, applicationRepository)
	ids := []string{"key_dddddddddddddddd", "key_eeeeeeeeeeeeeeee"}
	service.newID = func() (string, error) {
		id := ids[0]
		ids = ids[1:]
		return id, nil
	}
	service.now = func() time.Time { return time.Now().UTC().Add(-48 * time.Hour) }
	expired := service.Create(managementContext, validAPIKeyCreateInput("app_aaaaaaaaaaaaaaaa", 1))
	service.now = func() time.Time { return time.Now().UTC() }
	active := service.Create(managementContext, validAPIKeyCreateInput("app_aaaaaaaaaaaaaaaa", 30))
	if expired.FailureCode != "" || active.FailureCode != "" {
		t.Fatalf("seed expiry and race API keys: expired=%s active=%s", expired.FailureCode, active.FailureCode)
	}
	server := &Server{config: config.Config{GatewayAuthMode: gatewayAPIKeyAuthenticationSource}, applicationCatalogRepository: applicationRepository, apiKeyRepository: keyRepository}
	authenticate := func(token string) gatewayAPIKeyAuthenticationResult {
		request := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
		request.Header.Set("Authorization", "Bearer "+token)
		return server.authenticateGatewayAPIKey(request, newRequestTrace(request, "/v1/models"), "models:read")
	}
	if result := authenticate(expired.CredentialToken); result.FailureCode != APIKeyFailureExpired {
		t.Fatalf("expired API key was accepted: %#v", result)
	}

	var wait sync.WaitGroup
	results := make(chan string, 16)
	start := make(chan struct{})
	for index := 0; index < 16; index++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-start
			results <- authenticate(active.CredentialToken).FailureCode
		}()
	}
	close(start)
	revoked := service.Revoke(managementContext, active.Record.APIKeyID, 1)
	wait.Wait()
	close(results)
	if revoked.FailureCode != "" {
		t.Fatalf("revoke API key during authentication: %#v", revoked)
	}
	for failure := range results {
		if failure != "" && failure != APIKeyFailureRevoked {
			t.Fatalf("authentication/revoke race produced unstable failure: %s", failure)
		}
	}
	if result := authenticate(active.CredentialToken); result.FailureCode != APIKeyFailureRevoked {
		t.Fatalf("post-revoke authentication was accepted: %#v", result)
	}

	service.newID = func() (string, error) { return "key_ffffffffffffffff", nil }
	archivedKey := service.Create(managementContext, validAPIKeyCreateInput("app_aaaaaaaaaaaaaaaa", 30))
	if archivedKey.FailureCode != "" {
		t.Fatalf("issue key before application archive: failure=%s credential_present=%t", archivedKey.FailureCode, archivedKey.CredentialToken != "")
	}
	catalogContext := ApplicationCatalogContext{
		RequestContext: context.Background(), RequestID: "request_archive_app", TenantRef: managementContext.TenantRef,
		WorkspaceID: managementContext.WorkspaceID, ActorRef: managementContext.ActorRef,
		OwnerSubjectRef: managementContext.OwnerSubjectRef, AuditRef: "audit_archive_app", WriteEnabled: true,
	}
	archived := newApplicationCatalogService(applicationRepository).Archive(catalogContext, "app_aaaaaaaaaaaaaaaa", 1)
	if archived.FailureCode != "" {
		t.Fatalf("archive API key application: %#v", archived)
	}
	if result := authenticate(archivedKey.CredentialToken); result.FailureCode != APIKeyFailureApplicationUnavailable {
		t.Fatalf("archived application API key was accepted: %#v", result)
	}
}

func TestControlPlaneAuthenticatorPreservesNorthboundAuthorization(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	request.Header.Set("Authorization", "Bearer rmd_dev_key_aaaaaaaaaaaaaaaa.secret")
	var observed string
	handler := withControlPlaneReadAuthenticator(http.HandlerFunc(func(_ http.ResponseWriter, request *http.Request) {
		observed = request.Header.Get("Authorization")
	}), &controlPlaneReadAuthenticator{mode: controlPlaneReadAuthModeSignedTestToken})
	handler.ServeHTTP(httptest.NewRecorder(), request)
	if observed != request.Header.Get("Authorization") {
		t.Fatalf("control-plane authenticator consumed northbound Authorization: %q", observed)
	}
}
