package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"radishmind.local/services/platform/internal/config"
)

func TestGatewayModelPricingAdminHTTPManagesExactScopedRevisions(t *testing.T) {
	repository := newMemoryGatewayModelPricingRepository()
	server := newGatewayModelPricingHTTPTestServer(repository)
	mux := gatewayModelPricingHTTPTestMux(server)

	put := newGatewayModelPricingAdminTestRequest(http.MethodPut, gatewayModelPricingAdminRoute, []byte(`{
        "expected_version":0,"provider_id":"provider_demo","profile_id":"profile_demo","model_id":"model_alpha",
        "currency":"USD","input_price_micros_per_token_unit":1000000,
        "output_price_micros_per_token_unit":3000000,"reason":"initial development evidence"
    }`), "admin_gateway_pricing:write")
	putResponse := httptest.NewRecorder()
	mux.ServeHTTP(putResponse, put)
	putEnvelope := decodeGatewayModelPricingEnvelope(t, putResponse, http.StatusOK)
	if putEnvelope.Policy == nil || putEnvelope.Policy.RecordVersion != 1 ||
		putEnvelope.Policy.InputPriceMicrosPerTokenUnit != 1_000_000 || putEnvelope.FailureCode != nil ||
		putEnvelope.Policy.RequestID != putEnvelope.RequestID || putEnvelope.Policy.AuditRef == putEnvelope.RequestID {
		t.Fatalf("unexpected pricing PUT response: %+v", putEnvelope)
	}

	get := newGatewayModelPricingAdminTestRequest(http.MethodGet,
		gatewayModelPricingAdminRoute+"?provider_id=provider_demo&profile_id=profile_demo&model_id=model_alpha",
		nil, "admin_gateway_pricing:read")
	getResponse := httptest.NewRecorder()
	mux.ServeHTTP(getResponse, get)
	getEnvelope := decodeGatewayModelPricingEnvelope(t, getResponse, http.StatusOK)
	if getEnvelope.Policy == nil || getEnvelope.Policy.PolicyID != putEnvelope.Policy.PolicyID ||
		getResponse.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("unexpected pricing GET response: envelope=%+v cache=%q", getEnvelope, getResponse.Header().Get("Cache-Control"))
	}

	missing := newGatewayModelPricingAdminTestRequest(http.MethodGet,
		gatewayModelPricingAdminRoute+"?provider_id=provider_demo&profile_id=profile_demo&model_id=model_beta",
		nil, "admin_gateway_pricing:read")
	missingResponse := httptest.NewRecorder()
	mux.ServeHTTP(missingResponse, missing)
	missingEnvelope := decodeGatewayModelPricingEnvelope(t, missingResponse, http.StatusNotFound)
	if missingEnvelope.FailureCode == nil || *missingEnvelope.FailureCode != GatewayModelPricingFailurePolicyNotFound {
		t.Fatalf("exact model scope leaked: %+v", missingEnvelope)
	}
}

func TestGatewayModelPricingAdminHTTPReturnsConflictMetadata(t *testing.T) {
	repository := newMemoryGatewayModelPricingRepository()
	server := newGatewayModelPricingHTTPTestServer(repository)
	mux := gatewayModelPricingHTTPTestMux(server)
	body := []byte(`{
        "expected_version":0,"provider_id":"provider_demo","profile_id":"profile_demo","model_id":"model_alpha",
        "currency":"USD","input_price_micros_per_token_unit":1,
        "output_price_micros_per_token_unit":2,"reason":"pricing evidence"
    }`)
	mux.ServeHTTP(httptest.NewRecorder(), newGatewayModelPricingAdminTestRequest(
		http.MethodPut, gatewayModelPricingAdminRoute, body, "admin_gateway_pricing:write",
	))
	conflictResponse := httptest.NewRecorder()
	mux.ServeHTTP(conflictResponse, newGatewayModelPricingAdminTestRequest(
		http.MethodPut, gatewayModelPricingAdminRoute, body, "admin_gateway_pricing:write",
	))
	conflict := decodeGatewayModelPricingEnvelope(t, conflictResponse, http.StatusConflict)
	if conflict.FailureCode == nil || *conflict.FailureCode != GatewayModelPricingFailureVersionConflict ||
		conflict.CurrentVersion != 1 || conflict.Policy != nil {
		t.Fatalf("pricing conflict metadata drifted: %+v", conflict)
	}
}

func TestGatewayModelPricingAdminHTTPFailsBeforeRepositoryMutation(t *testing.T) {
	repository := newMemoryGatewayModelPricingRepository()
	server := newGatewayModelPricingHTTPTestServer(repository)
	mux := gatewayModelPricingHTTPTestMux(server)
	validBody := `{
        "expected_version":0,"provider_id":"provider_demo","profile_id":"profile_demo","model_id":"model_alpha",
        "currency":"USD","input_price_micros_per_token_unit":1,
        "output_price_micros_per_token_unit":2,"reason":"pricing evidence"
    }`
	unknownBody := `{
        "expected_version":0,"provider_id":"provider_demo","profile_id":"profile_demo","model_id":"model_alpha",
        "currency":"USD","input_price_micros_per_token_unit":1,
        "output_price_micros_per_token_unit":2,"reason":"pricing evidence","unknown":true
    }`
	sensitiveBody := `{
        "expected_version":0,"provider_id":"provider_demo","profile_id":"profile_demo","model_id":"model_alpha",
        "currency":"USD","input_price_micros_per_token_unit":1,
        "output_price_micros_per_token_unit":2,"reason":"contains secret material"
    }`
	tests := []struct {
		name    string
		request *http.Request
		status  int
		failure string
	}{
		{
			"unknown field",
			newGatewayModelPricingAdminTestRequest(http.MethodPut, gatewayModelPricingAdminRoute,
				[]byte(unknownBody), "admin_gateway_pricing:write"),
			http.StatusBadRequest, GatewayModelPricingFailurePayloadInvalid,
		},
		{
			"sensitive reason",
			newGatewayModelPricingAdminTestRequest(http.MethodPut, gatewayModelPricingAdminRoute,
				[]byte(sensitiveBody), "admin_gateway_pricing:write"),
			http.StatusBadRequest, GatewayModelPricingFailurePayloadInvalid,
		},
		{
			"permission denied",
			newGatewayModelPricingAdminTestRequest(http.MethodPut, gatewayModelPricingAdminRoute,
				[]byte(validBody), "admin_gateway_pricing:read"),
			http.StatusForbidden, GatewayModelPricingFailureScopeDenied,
		},
		{
			"forbidden query",
			newGatewayModelPricingAdminTestRequest(http.MethodGet,
				gatewayModelPricingAdminRoute+"?provider_id=provider_demo&profile_id=profile_demo&model_id=model_alpha&extra=1",
				nil, "admin_gateway_pricing:read"),
			http.StatusBadRequest, GatewayModelPricingFailurePayloadInvalid,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			mux.ServeHTTP(response, testCase.request)
			envelope := decodeGatewayModelPricingEnvelope(t, response, testCase.status)
			if envelope.FailureCode == nil || *envelope.FailureCode != testCase.failure || len(repository.current) != 0 {
				t.Fatalf("failure reached pricing repository: envelope=%+v current=%d", envelope, len(repository.current))
			}
		})
	}

	production := newGatewayModelPricingAdminTestRequest(http.MethodPut, gatewayModelPricingAdminRoute,
		[]byte(validBody), "admin_gateway_pricing:write")
	production.Header.Set(gatewayModelPricingEnvironmentHeader, "production")
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, production)
	envelope := decodeGatewayModelPricingEnvelope(t, response, http.StatusForbidden)
	if envelope.FailureCode == nil || *envelope.FailureCode != GatewayModelPricingFailureEnvironmentForbidden || len(repository.current) != 0 {
		t.Fatalf("production pricing reached repository: %+v", envelope)
	}
}

func newGatewayModelPricingHTTPTestServer(repository GatewayModelPricingRepository) *Server {
	return &Server{
		config: config.Config{
			GatewayModelPricingDevHTTPEnabled: true, GatewayModelPricingDevWriteEnabled: true,
			GatewayModelPricingEnvironment: "test",
		},
		gatewayModelPricingRepository: repository,
		workspaceMembershipProvider:   newDeterministicDevTestWorkspaceMembershipProvider(),
	}
}

func gatewayModelPricingHTTPTestMux(server *Server) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc(gatewayModelPricingAdminReadRoute, server.handleReadGatewayModelPricing)
	mux.HandleFunc(gatewayModelPricingAdminPutRoute, server.handlePutGatewayModelPricing)
	return mux
}

func newGatewayModelPricingAdminTestRequest(
	method string,
	path string,
	body []byte,
	permission string,
) *http.Request {
	request := httptest.NewRequest(method, path, bytes.NewReader(body))
	auth := controlPlaneReadTestAuth("tenant_demo", permission)
	request = request.WithContext(withControlPlaneReadFakeAuthContext(request.Context(), auth))
	request.Header.Set(activeWorkspaceHeader, "workspace_demo")
	request.Header.Set(gatewayModelPricingEnvironmentHeader, "test")
	request.Header.Set("Content-Type", "application/json")
	return request
}

func decodeGatewayModelPricingEnvelope(
	t *testing.T,
	response *httptest.ResponseRecorder,
	expectedStatus int,
) gatewayModelPricingEnvelope {
	t.Helper()
	if response.Code != expectedStatus {
		t.Fatalf("unexpected pricing status %d, body=%s", response.Code, response.Body.String())
	}
	var envelope gatewayModelPricingEnvelope
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode pricing response: %v", err)
	}
	return envelope
}
