package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/bridge"
	"radishmind.local/services/platform/internal/config"
)

func TestGatewayRequestHistoryRecordsNorthboundAndReadsScopedHistory(t *testing.T) {
	server := newGatewayRequestHistoryHTTPTestServer()
	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/chat/completions",
		strings.NewReader(`{"model":"platform-model","messages":[{"role":"user","content":"private prompt must not persist"}]}`),
	)
	request.Header.Set("X-Request-Id", "request_gateway_success")
	setGatewayRequestDevHeaders(request, "gateway_requests:read")
	response := httptest.NewRecorder()

	server.httpServer.Handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("northbound request failed: %d %s", response.Code, response.Body.String())
	}
	listRequest := httptest.NewRequest(
		http.MethodGet,
		"/v1/model-gateway/requests?workspace_id=workspace_demo&consumer_ref=consumer_demo&status=succeeded",
		nil,
	)
	setGatewayRequestDevHeaders(listRequest, "gateway_requests:read")
	listResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(listResponse, listRequest)
	listEnvelope := decodeGatewayRequestListEnvelope(t, listResponse)
	if listEnvelope.FailureCode != nil || len(listEnvelope.Requests) != 1 {
		t.Fatalf("unexpected history list: %#v", listEnvelope)
	}
	summary := listEnvelope.Requests[0]
	if summary.RequestID != "request_gateway_success" || summary.Route != "/v1/chat/completions" ||
		summary.Protocol != northboundProtocolChatCompletions || summary.SelectedProvider != "mock" ||
		summary.SelectedModel != "platform-model" || !summary.ProviderDurationAvailable || summary.ProviderDurationMS != 3 ||
		summary.UsageAvailability != GatewayRequestUsageNotReported {
		t.Fatalf("unexpected request summary: %#v", summary)
	}
	if strings.Contains(listResponse.Body.String(), "private prompt") || strings.Contains(listResponse.Body.String(), "bridge summary") {
		t.Fatalf("request history leaked payload: %s", listResponse.Body.String())
	}

	readRequest := httptest.NewRequest(
		http.MethodGet,
		"/v1/model-gateway/requests/request_gateway_success?workspace_id=workspace_demo&consumer_ref=consumer_demo",
		nil,
	)
	setGatewayRequestDevHeaders(readRequest, "gateway_requests:read")
	readResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(readResponse, readRequest)
	readEnvelope := decodeGatewayRequestReadEnvelope(t, readResponse)
	if readEnvelope.FailureCode != nil || readEnvelope.Request == nil ||
		readEnvelope.Request.Status != GatewayRequestStatusSucceeded || !readEnvelope.Request.GatewayDurationAvailable ||
		readEnvelope.Request.GatewayDurationMS != 5 {
		t.Fatalf("unexpected request detail: %#v", readEnvelope)
	}
}

func TestGatewayRequestHistoryRecordsInvalidBodyAndSkipsUnscopedRequest(t *testing.T) {
	server := newGatewayRequestHistoryHTTPTestServer()
	invalid := httptest.NewRequest(http.MethodPost, "/v1/responses", strings.NewReader(`{"model":`))
	invalid.Header.Set("X-Request-Id", "request_gateway_invalid")
	setGatewayRequestDevHeaders(invalid, "gateway_requests:read")
	invalidResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(invalidResponse, invalid)
	if invalidResponse.Code != http.StatusBadRequest {
		t.Fatalf("invalid request should fail: %d %s", invalidResponse.Code, invalidResponse.Body.String())
	}

	unscoped := httptest.NewRequest(
		http.MethodPost,
		"/v1/chat/completions",
		strings.NewReader(`{"model":"platform-model","messages":[{"role":"user","content":"unscoped"}]}`),
	)
	unscoped.Header.Set("X-Request-Id", "request_gateway_unscoped")
	unscopedResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(unscopedResponse, unscoped)
	if unscopedResponse.Code != http.StatusOK {
		t.Fatalf("unscoped compatibility request changed behavior: %d %s", unscopedResponse.Code, unscopedResponse.Body.String())
	}

	listRequest := httptest.NewRequest(
		http.MethodGet,
		"/v1/model-gateway/requests?workspace_id=workspace_demo&consumer_ref=consumer_demo&status=failed",
		nil,
	)
	setGatewayRequestDevHeaders(listRequest, "gateway_requests:read")
	listResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(listResponse, listRequest)
	listEnvelope := decodeGatewayRequestListEnvelope(t, listResponse)
	if len(listEnvelope.Requests) != 1 || listEnvelope.Requests[0].RequestID != "request_gateway_invalid" ||
		listEnvelope.Requests[0].FailureCode != "INVALID_JSON" ||
		listEnvelope.Requests[0].FailureBoundary != errorBoundaryNorthboundRequest {
		t.Fatalf("invalid request record missing: %#v", listEnvelope)
	}
	allRequest := httptest.NewRequest(
		http.MethodGet,
		"/v1/model-gateway/requests?workspace_id=workspace_demo&consumer_ref=consumer_demo",
		nil,
	)
	setGatewayRequestDevHeaders(allRequest, "gateway_requests:read")
	allResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(allResponse, allRequest)
	allEnvelope := decodeGatewayRequestListEnvelope(t, allResponse)
	if len(allEnvelope.Requests) != 1 {
		t.Fatalf("unscoped request must not create history: %#v", allEnvelope)
	}
}

func TestGatewayRequestHistoryRecordsResponsesAndMessagesStream(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		body      string
		requestID string
		protocol  string
		stream    bool
	}{
		{
			name: "responses unary", path: "/v1/responses",
			body:      `{"model":"platform-model","input":"private response input"}`,
			requestID: "request_gateway_responses", protocol: northboundProtocolResponses,
		},
		{
			name: "messages stream", path: "/v1/messages",
			body:      `{"model":"platform-model","messages":[{"role":"user","content":"private message"}],"stream":true}`,
			requestID: "request_gateway_messages_stream", protocol: northboundProtocolMessages, stream: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := newGatewayRequestHistoryHTTPTestServer()
			request := httptest.NewRequest(http.MethodPost, test.path, strings.NewReader(test.body))
			request.Header.Set("X-Request-Id", test.requestID)
			setGatewayRequestDevHeaders(request, "gateway_requests:read")
			response := httptest.NewRecorder()
			server.httpServer.Handler.ServeHTTP(response, request)
			if response.Code != http.StatusOK {
				t.Fatalf("northbound request failed: %d %s", response.Code, response.Body.String())
			}
			record, found, err := server.gatewayRequestStore().ReadRequest(gatewayRequestTestContext(), test.requestID)
			if err != nil || !found {
				t.Fatalf("request history missing: found=%v err=%v", found, err)
			}
			if record.Status != GatewayRequestStatusSucceeded || record.Protocol != test.protocol || record.Stream != test.stream {
				t.Fatalf("unexpected request history: %#v", record)
			}
			encoded, err := json.Marshal(record)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(encoded), "private response input") || strings.Contains(string(encoded), "private message") {
				t.Fatalf("request history leaked input: %s", encoded)
			}
		})
	}
}

func TestGatewayRequestHistoryRecordsBridgeTimeoutAndCancellationTerminalStates(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		body           string
		expectedStatus GatewayRequestStatus
		expectedCode   string
		expectedHTTP   int
		configure      func(*fakeBridge, context.CancelFunc)
	}{
		{
			name: "unary timeout", path: "/v1/chat/completions",
			body:           `{"model":"platform-model","messages":[{"role":"user","content":"timeout payload must not persist"}]}`,
			expectedStatus: GatewayRequestStatusFailed,
			expectedCode:   bridge.ErrorCodeWorkerTimeout,
			expectedHTTP:   http.StatusGatewayTimeout,
			configure: func(fake *fakeBridge, _ context.CancelFunc) {
				fake.handleErr = context.DeadlineExceeded
			},
		},
		{
			name: "unary canceled", path: "/v1/chat/completions",
			body:           `{"model":"platform-model","messages":[{"role":"user","content":"canceled payload must not persist"}]}`,
			expectedStatus: GatewayRequestStatusCanceled,
			expectedCode:   bridge.ErrorCodeWorkerCanceled,
			expectedHTTP:   http.StatusRequestTimeout,
			configure: func(fake *fakeBridge, cancel context.CancelFunc) {
				fake.handleHook = cancel
				fake.handleErr = context.Canceled
			},
		},
		{
			name: "stream canceled", path: "/v1/messages",
			body:           `{"model":"platform-model","messages":[{"role":"user","content":"stream payload must not persist"}],"stream":true}`,
			expectedStatus: GatewayRequestStatusCanceled,
			expectedCode:   bridge.ErrorCodeWorkerCanceled,
			expectedHTTP:   http.StatusRequestTimeout,
			configure: func(fake *fakeBridge, cancel context.CancelFunc) {
				fake.streamHook = cancel
				fake.streamErr = context.Canceled
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := newGatewayRequestHistoryHTTPTestServer()
			observedStore := &terminalContextObservingGatewayRequestStore{gatewayRequestStore: server.gatewayRequestStore()}
			server.gatewayRequestHistoryStore = observedStore
			requestContext, cancel := context.WithCancel(context.Background())
			defer cancel()
			test.configure(server.bridge.(*fakeBridge), cancel)
			request := httptest.NewRequest(http.MethodPost, test.path, strings.NewReader(test.body)).WithContext(requestContext)
			request.Header.Set("X-Request-Id", "request_gateway_"+strings.ReplaceAll(test.name, " ", "_"))
			setGatewayRequestDevHeaders(request, "gateway_requests:read")
			response := httptest.NewRecorder()

			server.httpServer.Handler.ServeHTTP(response, request)

			if response.Code != test.expectedHTTP {
				t.Fatalf("unexpected northbound status: %d %s", response.Code, response.Body.String())
			}
			record, found, err := observedStore.ReadRequest(gatewayRequestTestContext(), request.Header.Get("X-Request-Id"))
			if err != nil || !found {
				t.Fatalf("terminal request record missing: found=%v err=%v", found, err)
			}
			if record.Status != test.expectedStatus || record.FailureCode != test.expectedCode ||
				record.FailureBoundary != errorBoundaryPythonBridge || record.RecordVersion != 3 {
				t.Fatalf("unexpected terminal request record: %#v", record)
			}
			if observedStore.terminalUpdates != 1 {
				t.Fatalf("terminal store context was not observed exactly once: %d", observedStore.terminalUpdates)
			}
			encoded, marshalErr := json.Marshal(record)
			if marshalErr != nil {
				t.Fatal(marshalErr)
			}
			if strings.Contains(string(encoded), "payload must not persist") {
				t.Fatalf("terminal request history leaked payload: %s", encoded)
			}
		})
	}
}

func TestGatewayRequestHistoryRecordsQueueFullAsFailed(t *testing.T) {
	server := newGatewayRequestHistoryHTTPTestServer()
	request := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	request.Header.Set("X-Request-Id", "request_gateway_queue_full")
	setGatewayRequestDevHeaders(request, "gateway_requests:read")
	trace := newRequestTrace(request, "/v1/chat/completions")
	server.startGatewayRequestTrace(request, &trace, northboundProtocolChatCompletions)
	trace.applySelection(northboundSelection{provider: "mock", model: "platform-model", source: "config"})
	server.checkpointGatewayRequestTrace(&trace, false)
	response := httptest.NewRecorder()

	server.writePlatformError(response, trace, bridge.ErrorCodeWorkerQueueFull, "bridge worker queue is full")

	record, found, err := server.gatewayRequestStore().ReadRequest(gatewayRequestTestContext(), trace.requestID)
	if err != nil || !found {
		t.Fatalf("queue-full request record missing: found=%v err=%v", found, err)
	}
	if response.Code != http.StatusServiceUnavailable || record.Status != GatewayRequestStatusFailed ||
		record.FailureCode != bridge.ErrorCodeWorkerQueueFull || record.FailureBoundary != errorBoundaryPythonBridge {
		t.Fatalf("unexpected queue-full terminal evidence: response=%d record=%#v", response.Code, record)
	}
}

func TestGatewayRequestHistoryStoreCreateFailureDoesNotChangeProviderOutcome(t *testing.T) {
	server := newGatewayRequestHistoryHTTPTestServer()
	server.gatewayRequestHistoryStore = failingGatewayRequestStore{}
	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/chat/completions",
		strings.NewReader(`{"model":"platform-model","messages":[{"role":"user","content":"provider outcome remains authoritative"}]}`),
	)
	request.Header.Set("X-Request-Id", "request_gateway_store_failure")
	setGatewayRequestDevHeaders(request, "gateway_requests:read")
	response := httptest.NewRecorder()

	server.httpServer.Handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "bridge summary") {
		t.Fatalf("history store failure changed provider outcome: %d %s", response.Code, response.Body.String())
	}
}

func TestGatewayRequestHistoryReadScopeAndDevGateFailClosed(t *testing.T) {
	server := newGatewayRequestHistoryHTTPTestServer()
	missingScope := httptest.NewRequest(
		http.MethodGet,
		"/v1/model-gateway/requests?workspace_id=workspace_demo&consumer_ref=consumer_demo",
		nil,
	)
	setGatewayRequestDevHeaders(missingScope, "models:read")
	missingScopeResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(missingScopeResponse, missingScope)
	missingScopeEnvelope := decodeGatewayRequestListEnvelope(t, missingScopeResponse)
	if missingScopeEnvelope.FailureCode == nil ||
		*missingScopeEnvelope.FailureCode != string(GatewayRequestHistoryFailureScopeDenied) {
		t.Fatalf("missing read scope was accepted: %#v", missingScopeEnvelope)
	}

	server.config.GatewayRequestHistoryDevEnabled = false
	disabled := httptest.NewRequest(
		http.MethodGet,
		"/v1/model-gateway/requests?workspace_id=workspace_demo&consumer_ref=consumer_demo",
		nil,
	)
	setGatewayRequestDevHeaders(disabled, "gateway_requests:read")
	disabledResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(disabledResponse, disabled)
	if disabledResponse.Code != http.StatusForbidden ||
		!strings.Contains(disabledResponse.Body.String(), "GATEWAY_REQUEST_HISTORY_DEV_DISABLED") {
		t.Fatalf("disabled history route did not fail closed: %d %s", disabledResponse.Code, disabledResponse.Body.String())
	}
}

func TestGatewayRequestHistoryCORSAllowsDedicatedHeaders(t *testing.T) {
	server := newGatewayRequestHistoryHTTPTestServer()
	request := httptest.NewRequest(http.MethodOptions, "/v1/model-gateway/requests", nil)
	request.Header.Set("Origin", "http://127.0.0.1:4100")
	request.Header.Set("Access-Control-Request-Method", http.MethodGet)
	request.Header.Set("Access-Control-Request-Headers", gatewayRequestDevTenantHeader)
	response := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("gateway request history preflight failed: %d %s", response.Code, response.Body.String())
	}
	allowed := response.Header().Get("Access-Control-Allow-Headers")
	for _, header := range []string{gatewayRequestDevTenantHeader, gatewayRequestDevWorkspaceHeader, gatewayRequestDevConsumerHeader, gatewayRequestDevScopesHeader} {
		if !strings.Contains(allowed, header) {
			t.Fatalf("missing CORS header %s: %s", header, allowed)
		}
	}
}

func newGatewayRequestHistoryHTTPTestServer() *Server {
	server := NewServer(config.Config{
		ControlPlaneReadDevAuthEnabled:  true,
		GatewayRequestHistoryDevEnabled: true,
		BridgeTimeout:                   time.Second,
		Provider:                        "mock",
		Model:                           "platform-model",
	}, Options{BuildVersion: "test"})
	server.bridge = &fakeBridge{
		envelope: bridge.GatewayEnvelope{
			SchemaVersion: 1,
			Status:        "ok",
			RequestID:     "bridge_request",
			Project:       "radish",
			Task:          "answer_docs_question",
			Response:      map[string]any{"summary": "bridge summary"},
			Metadata: map[string]any{
				"duration_ms":          5,
				"provider_duration_ms": 3,
			},
		},
	}
	return server
}

func setGatewayRequestDevHeaders(request *http.Request, scopes string) {
	request.Header.Set(gatewayRequestDevTenantHeader, "tenant_demo")
	request.Header.Set(gatewayRequestDevWorkspaceHeader, "workspace_demo")
	request.Header.Set(gatewayRequestDevConsumerHeader, "consumer_demo")
	request.Header.Set(gatewayRequestDevApplicationHeader, "application_demo")
	request.Header.Set(gatewayRequestDevSubjectHeader, "subject_demo")
	request.Header.Set(gatewayRequestDevScopesHeader, scopes)
	request.Header.Set(gatewayRequestDevAuditHeader, "audit_context_demo")
}

type terminalContextObservingGatewayRequestStore struct {
	gatewayRequestStore
	terminalUpdates int
}

func (store *terminalContextObservingGatewayRequestStore) UpdateRequest(
	requestContext GatewayRequestContext,
	record *GatewayRequestRecord,
) error {
	if record != nil && isTerminalGatewayRequestStatus(record.Status) {
		if requestContext.RequestContext == nil || requestContext.RequestContext.Err() != nil {
			return errGatewayRequestStoreContract
		}
		if _, ok := requestContext.RequestContext.Deadline(); !ok {
			return errGatewayRequestStoreContract
		}
		store.terminalUpdates++
	}
	return store.gatewayRequestStore.UpdateRequest(requestContext, record)
}

func decodeGatewayRequestListEnvelope(t *testing.T, response *httptest.ResponseRecorder) gatewayRequestListEnvelope {
	t.Helper()
	if response.Code != http.StatusOK {
		t.Fatalf("unexpected history list status: %d %s", response.Code, response.Body.String())
	}
	var envelope gatewayRequestListEnvelope
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode history list: %v", err)
	}
	return envelope
}

func decodeGatewayRequestReadEnvelope(t *testing.T, response *httptest.ResponseRecorder) gatewayRequestReadEnvelope {
	t.Helper()
	if response.Code != http.StatusOK {
		t.Fatalf("unexpected history read status: %d %s", response.Code, response.Body.String())
	}
	var envelope gatewayRequestReadEnvelope
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode history read: %v", err)
	}
	return envelope
}

type failingGatewayRequestStore struct{}

func (failingGatewayRequestStore) CreateRequest(GatewayRequestContext, *GatewayRequestRecord) error {
	return errGatewayRequestStoreContract
}

func (failingGatewayRequestStore) UpdateRequest(GatewayRequestContext, *GatewayRequestRecord) error {
	return errGatewayRequestStoreContract
}

func (failingGatewayRequestStore) ReadRequest(GatewayRequestContext, string) (GatewayRequestRecord, bool, error) {
	return GatewayRequestRecord{}, false, errGatewayRequestStoreContract
}

func (failingGatewayRequestStore) ListRequests(GatewayRequestContext, GatewayRequestListFilter) (GatewayRequestListPage, error) {
	return GatewayRequestListPage{}, errGatewayRequestStoreContract
}
