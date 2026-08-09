package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"radishmind.local/services/platform/internal/config"
)

func TestGatewayRequestQuotaAdminHTTPManagesScopedPolicy(t *testing.T) {
	repository := newMemoryGatewayRequestQuotaRepository()
	server := &Server{
		config: configForGatewayRequestQuotaHTTPTest(), gatewayRequestQuotaRepository: repository,
		workspaceMembershipProvider: newDeterministicDevTestWorkspaceMembershipProvider(),
	}
	mux := http.NewServeMux()
	mux.HandleFunc(gatewayRequestQuotaAdminReadRoute, server.handleReadGatewayRequestQuota)
	mux.HandleFunc(gatewayRequestQuotaAdminPutRoute, server.handlePutGatewayRequestQuota)

	putRequest := newGatewayRequestQuotaAdminTestRequest(http.MethodPut, []byte(`{"expected_version":0,"request_limit":2}`), "admin_gateway_quotas:write")
	putResponse := httptest.NewRecorder()
	mux.ServeHTTP(putResponse, putRequest)
	putEnvelope := decodeGatewayRequestQuotaEnvelope(t, putResponse, http.StatusOK)
	if putEnvelope.Policy == nil || putEnvelope.Policy.RecordVersion != 1 ||
		putEnvelope.Policy.LastRequestID != putEnvelope.RequestID || putEnvelope.Policy.LastAuditRef == putEnvelope.Policy.LastRequestID || putEnvelope.Usage == nil ||
		putEnvelope.Usage.RemainingRequestCount != 2 || putEnvelope.FailureCode != nil {
		t.Fatalf("unexpected quota PUT response: %+v", putEnvelope)
	}

	getRequest := newGatewayRequestQuotaAdminTestRequest(http.MethodGet, nil, "admin_gateway_quotas:read")
	getResponse := httptest.NewRecorder()
	mux.ServeHTTP(getResponse, getRequest)
	getEnvelope := decodeGatewayRequestQuotaEnvelope(t, getResponse, http.StatusOK)
	if getEnvelope.Policy == nil || getEnvelope.Policy.PolicyID != putEnvelope.Policy.PolicyID || getEnvelope.Usage == nil ||
		getResponse.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("unexpected quota GET response: envelope=%+v cache=%q", getEnvelope, getResponse.Header().Get("Cache-Control"))
	}
}

func TestGatewayRequestQuotaAdminHTTPRejectsEnvironmentAndVersion(t *testing.T) {
	repository := newMemoryGatewayRequestQuotaRepository()
	server := &Server{
		config: configForGatewayRequestQuotaHTTPTest(), gatewayRequestQuotaRepository: repository,
		workspaceMembershipProvider: newDeterministicDevTestWorkspaceMembershipProvider(),
	}
	mux := http.NewServeMux()
	mux.HandleFunc(gatewayRequestQuotaAdminPutRoute, server.handlePutGatewayRequestQuota)

	production := newGatewayRequestQuotaAdminTestRequest(http.MethodPut, []byte(`{"expected_version":0,"request_limit":2}`), "admin_gateway_quotas:write")
	production.Header.Set(gatewayRequestQuotaEnvironmentHeader, "production")
	productionResponse := httptest.NewRecorder()
	mux.ServeHTTP(productionResponse, production)
	productionEnvelope := decodeGatewayRequestQuotaEnvelope(t, productionResponse, http.StatusForbidden)
	if productionEnvelope.FailureCode == nil || *productionEnvelope.FailureCode != GatewayRequestQuotaFailureEnvironmentForbidden {
		t.Fatalf("production environment was accepted: %+v", productionEnvelope)
	}

	create := newGatewayRequestQuotaAdminTestRequest(http.MethodPut, []byte(`{"expected_version":0,"request_limit":2}`), "admin_gateway_quotas:write")
	mux.ServeHTTP(httptest.NewRecorder(), create)
	conflict := newGatewayRequestQuotaAdminTestRequest(http.MethodPut, []byte(`{"expected_version":0,"request_limit":3}`), "admin_gateway_quotas:write")
	conflictResponse := httptest.NewRecorder()
	mux.ServeHTTP(conflictResponse, conflict)
	conflictEnvelope := decodeGatewayRequestQuotaEnvelope(t, conflictResponse, http.StatusConflict)
	if conflictEnvelope.FailureCode == nil || *conflictEnvelope.FailureCode != GatewayRequestQuotaFailurePolicyVersionConflict {
		t.Fatalf("quota version conflict drifted: %+v", conflictEnvelope)
	}
}

func TestGatewayRequestQuotaAdminHTTPFailsClosedBeforeRepositoryMutation(t *testing.T) {
	repository := newMemoryGatewayRequestQuotaRepository()
	server := &Server{
		config: configForGatewayRequestQuotaHTTPTest(), gatewayRequestQuotaRepository: repository,
		workspaceMembershipProvider: newDeterministicDevTestWorkspaceMembershipProvider(),
	}
	mux := http.NewServeMux()
	mux.HandleFunc(gatewayRequestQuotaAdminReadRoute, server.handleReadGatewayRequestQuota)
	mux.HandleFunc(gatewayRequestQuotaAdminPutRoute, server.handlePutGatewayRequestQuota)

	missing := newGatewayRequestQuotaAdminTestRequest(http.MethodGet, nil, "admin_gateway_quotas:read")
	missingResponse := httptest.NewRecorder()
	mux.ServeHTTP(missingResponse, missing)
	missingEnvelope := decodeGatewayRequestQuotaEnvelope(t, missingResponse, http.StatusServiceUnavailable)
	if missingEnvelope.FailureCode == nil || *missingEnvelope.FailureCode != GatewayRequestQuotaFailurePolicyNotFound {
		t.Fatalf("missing quota policy did not fail closed: %+v", missingEnvelope)
	}

	for _, body := range [][]byte{
		[]byte(`{"expected_version":0,"request_limit":2,"unknown":true}`),
		[]byte(`{"expected_version":0,"expected_version":1,"request_limit":2}`),
	} {
		request := newGatewayRequestQuotaAdminTestRequest(http.MethodPut, body, "admin_gateway_quotas:write")
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		envelope := decodeGatewayRequestQuotaEnvelope(t, response, http.StatusBadRequest)
		if envelope.FailureCode == nil || *envelope.FailureCode != GatewayRequestQuotaFailurePayloadInvalid || len(repository.policies) != 0 {
			t.Fatalf("invalid quota payload reached repository: envelope=%+v policies=%d", envelope, len(repository.policies))
		}
	}

	denied := newGatewayRequestQuotaAdminTestRequest(http.MethodPut, []byte(`{"expected_version":0,"request_limit":2}`), "admin_gateway_quotas:read")
	deniedResponse := httptest.NewRecorder()
	mux.ServeHTTP(deniedResponse, denied)
	deniedEnvelope := decodeGatewayRequestQuotaEnvelope(t, deniedResponse, http.StatusForbidden)
	if deniedEnvelope.FailureCode == nil || *deniedEnvelope.FailureCode != GatewayRequestQuotaFailureScopeDenied || len(repository.policies) != 0 {
		t.Fatalf("quota write permission denial reached repository: envelope=%+v policies=%d", deniedEnvelope, len(repository.policies))
	}

	server.config.GatewayRequestQuotaDevWriteEnabled = false
	writeDisabled := newGatewayRequestQuotaAdminTestRequest(http.MethodPut, []byte(`{"expected_version":0,"request_limit":2}`), "admin_gateway_quotas:write")
	writeDisabledResponse := httptest.NewRecorder()
	mux.ServeHTTP(writeDisabledResponse, writeDisabled)
	writeDisabledEnvelope := decodeGatewayRequestQuotaEnvelope(t, writeDisabledResponse, http.StatusForbidden)
	if writeDisabledEnvelope.FailureCode == nil || *writeDisabledEnvelope.FailureCode != GatewayRequestQuotaFailureDisabled || len(repository.policies) != 0 {
		t.Fatalf("quota write gate reached repository: envelope=%+v policies=%d", writeDisabledEnvelope, len(repository.policies))
	}
}

func configForGatewayRequestQuotaHTTPTest() config.Config {
	return config.Config{GatewayRequestQuotaDevHTTPEnabled: true, GatewayRequestQuotaDevWriteEnabled: true}
}

func newGatewayRequestQuotaAdminTestRequest(method string, body []byte, permission string) *http.Request {
	request := httptest.NewRequest(method, "/v1/admin/gateway-request-quotas/app-demo", bytes.NewReader(body))
	auth := controlPlaneReadTestAuth("tenant_demo", permission)
	request = request.WithContext(withControlPlaneReadFakeAuthContext(request.Context(), auth))
	request.Header.Set(activeWorkspaceHeader, "workspace_demo")
	request.Header.Set(gatewayRequestQuotaEnvironmentHeader, "test")
	request.Header.Set("Content-Type", "application/json")
	return request
}

func decodeGatewayRequestQuotaEnvelope(t *testing.T, response *httptest.ResponseRecorder, expectedStatus int) gatewayRequestQuotaEnvelope {
	t.Helper()
	if response.Code != expectedStatus {
		t.Fatalf("unexpected quota status %d, body=%s", response.Code, response.Body.String())
	}
	var envelope gatewayRequestQuotaEnvelope
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode quota response: %v", err)
	}
	return envelope
}
