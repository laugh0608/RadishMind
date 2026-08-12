package httpapi

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGatewayRequestHistoryCapturesImmutablePricingSnapshotAfterProviderAttempt(t *testing.T) {
	server := newGatewayRequestHistoryHTTPTestServer()
	t.Cleanup(server.Close)
	server.config.GatewayModelPricingCaptureDevEnabled = true
	server.config.GatewayModelPricingEnvironment = "test"
	server.config.ProviderProfile = "profile_demo"

	repository := newMemoryGatewayModelPricingRepository()
	server.gatewayModelPricingRepository = repository
	inner := server.bridge.(*fakeBridge)
	server.bridge = newGatewayProviderAttemptBridgeClient(inner)

	pricingContext := gatewayModelPricingTestContext("platform-model")
	pricingContext.Environment = "test"
	pricingContext.ProviderID = "mock"
	pricingContext.ProfileID = "profile_demo"
	initial := putGatewayModelPricingTestPolicy(
		t, repository, pricingContext, 0, 1_000_000, 2_000_000, "initial runtime rate",
	)
	var updated GatewayModelPricingResult
	inner.handleHook = func() {
		updated = newGatewayModelPricingService(repository).PutRevision(pricingContext, GatewayModelPricingPolicyInput{
			ExpectedVersion: 1, Currency: GatewayModelPricingCurrency,
			InputPriceMicrosPerTokenUnit: 10_000_000, OutputPriceMicrosPerTokenUnit: 20_000_000,
			Reason: "rate changed after provider selection",
		})
	}

	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/chat/completions",
		strings.NewReader(`{"model":"platform-model","messages":[{"role":"user","content":"price snapshot"}]}`),
	)
	request.Header.Set("X-Request-Id", "request_gateway_priced")
	setGatewayRequestDevHeaders(request, "gateway_requests:read")
	response := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK || updated.FailureCode != "" || updated.Policy == nil || updated.Policy.RecordVersion != 2 {
		t.Fatalf("provider request or post-selection policy update failed: status=%d update=%+v", response.Code, updated)
	}
	record, found, err := server.gatewayRequestStore().ReadRequest(gatewayRequestTestContext(), "request_gateway_priced")
	if err != nil || !found {
		t.Fatalf("priced request history missing: found=%v err=%v", found, err)
	}
	estimate := record.CostEstimate
	if record.SchemaVersion != gatewayRequestRecordSchemaVersionV2 ||
		estimate.Availability != GatewayRequestCostEstimated || estimate.EstimatedCostMicros == nil ||
		*estimate.EstimatedCostMicros != 17 || estimate.PricingPolicyVersion == nil ||
		*estimate.PricingPolicyVersion != 1 || estimate.PricingPolicyID != initial.PolicyID ||
		estimate.PricingPolicyDigest != initial.PolicyDigest || estimate.InputPriceMicrosPerTokenUnit == nil ||
		*estimate.InputPriceMicrosPerTokenUnit != 1_000_000 {
		t.Fatalf("request did not preserve the selection-time pricing snapshot: %+v", record)
	}

	listRequest := httptest.NewRequest(
		http.MethodGet,
		"/v1/model-gateway/requests?workspace_id=workspace_demo&consumer_ref=consumer_demo",
		nil,
	)
	setGatewayRequestDevHeaders(listRequest, "gateway_requests:read")
	listResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(listResponse, listRequest)
	list := decodeGatewayRequestListEnvelope(t, listResponse)
	if len(list.Requests) != 1 || list.Requests[0].CostEstimate.Availability != GatewayRequestCostEstimated ||
		list.Requests[0].CostEstimate.EstimatedCostMicros == nil || *list.Requests[0].CostEstimate.EstimatedCostMicros != 17 {
		t.Fatalf("history list omitted immutable cost evidence: %+v", list)
	}
}

func TestGatewayRequestHistoryDistinguishesMissingPricingAndPreProviderFailure(t *testing.T) {
	server := newGatewayRequestHistoryHTTPTestServer()
	t.Cleanup(server.Close)
	server.config.GatewayModelPricingCaptureDevEnabled = true
	server.config.GatewayModelPricingEnvironment = "test"
	server.config.ProviderProfile = "profile_demo"
	server.gatewayModelPricingRepository = newMemoryGatewayModelPricingRepository()
	server.bridge = newGatewayProviderAttemptBridgeClient(server.bridge)

	missingPrice := httptest.NewRequest(
		http.MethodPost,
		"/v1/chat/completions",
		strings.NewReader(`{"model":"platform-model","messages":[{"role":"user","content":"missing price"}]}`),
	)
	missingPrice.Header.Set("X-Request-Id", "request_gateway_missing_price")
	setGatewayRequestDevHeaders(missingPrice, "gateway_requests:read")
	missingPriceResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(missingPriceResponse, missingPrice)

	invalid := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":`))
	invalid.Header.Set("X-Request-Id", "request_gateway_pre_provider_failure")
	setGatewayRequestDevHeaders(invalid, "gateway_requests:read")
	invalidResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(invalidResponse, invalid)

	missingRecord, missingFound, missingErr := server.gatewayRequestStore().ReadRequest(
		gatewayRequestTestContext(), "request_gateway_missing_price",
	)
	invalidRecord, invalidFound, invalidErr := server.gatewayRequestStore().ReadRequest(
		gatewayRequestTestContext(), "request_gateway_pre_provider_failure",
	)
	if missingPriceResponse.Code != http.StatusOK || missingErr != nil || !missingFound ||
		missingRecord.CostEstimate.Availability != GatewayRequestCostPriceNotConfigured {
		t.Fatalf("actual provider attempt without pricing was not explicit: status=%d record=%+v err=%v", missingPriceResponse.Code, missingRecord, missingErr)
	}
	if invalidResponse.Code != http.StatusBadRequest || invalidErr != nil || !invalidFound ||
		invalidRecord.Usage.Availability != GatewayRequestUsageNotApplicable ||
		invalidRecord.CostEstimate.Availability != GatewayRequestCostNotApplicable {
		t.Fatalf("pre-provider failure was misclassified as a pricing failure: status=%d record=%+v err=%v", invalidResponse.Code, invalidRecord, invalidErr)
	}
}

func TestGatewayRequestHistoryPricesStreamingUsageAndKeepsProviderFailureExplicit(t *testing.T) {
	t.Run("stream reported usage", func(t *testing.T) {
		server := newGatewayRequestHistoryHTTPTestServer()
		t.Cleanup(server.Close)
		server.config.GatewayModelPricingCaptureDevEnabled = true
		server.config.GatewayModelPricingEnvironment = "test"
		server.config.ProviderProfile = "profile_demo"
		repository := newMemoryGatewayModelPricingRepository()
		server.gatewayModelPricingRepository = repository
		server.bridge = newGatewayProviderAttemptBridgeClient(server.bridge)

		pricingContext := gatewayModelPricingTestContext("platform-model")
		pricingContext.Environment = "test"
		pricingContext.ProviderID = "mock"
		pricingContext.ProfileID = "profile_demo"
		putGatewayModelPricingTestPolicy(
			t, repository, pricingContext, 0, 1_000_000, 2_000_000, "stream pricing evidence",
		)

		request := httptest.NewRequest(
			http.MethodPost,
			"/v1/messages",
			strings.NewReader(`{"model":"platform-model","messages":[{"role":"user","content":"stream price"}],"stream":true}`),
		)
		request.Header.Set("X-Request-Id", "request_gateway_priced_stream")
		setGatewayRequestDevHeaders(request, "gateway_requests:read")
		response := httptest.NewRecorder()
		server.httpServer.Handler.ServeHTTP(response, request)

		record, found, err := server.gatewayRequestStore().ReadRequest(
			gatewayRequestTestContext(), "request_gateway_priced_stream",
		)
		if response.Code != http.StatusOK || err != nil || !found || !record.Stream ||
			record.Usage.Availability != GatewayRequestUsageReported ||
			record.CostEstimate.Availability != GatewayRequestCostEstimated ||
			record.CostEstimate.EstimatedCostMicros == nil || *record.CostEstimate.EstimatedCostMicros != 17 {
			t.Fatalf("streaming price evidence drifted: status=%d found=%v record=%+v err=%v", response.Code, found, record, err)
		}
	})

	t.Run("provider failure without usage", func(t *testing.T) {
		server := newGatewayRequestHistoryHTTPTestServer()
		t.Cleanup(server.Close)
		server.config.GatewayModelPricingCaptureDevEnabled = true
		server.config.GatewayModelPricingEnvironment = "test"
		server.config.ProviderProfile = "profile_demo"
		repository := newMemoryGatewayModelPricingRepository()
		server.gatewayModelPricingRepository = repository
		inner := server.bridge.(*fakeBridge)
		inner.handleErr = errors.New("provider failed without usage")
		server.bridge = newGatewayProviderAttemptBridgeClient(inner)

		pricingContext := gatewayModelPricingTestContext("platform-model")
		pricingContext.Environment = "test"
		pricingContext.ProviderID = "mock"
		pricingContext.ProfileID = "profile_demo"
		putGatewayModelPricingTestPolicy(
			t, repository, pricingContext, 0, 1_000_000, 2_000_000, "provider failure pricing evidence",
		)

		request := httptest.NewRequest(
			http.MethodPost,
			"/v1/chat/completions",
			strings.NewReader(`{"model":"platform-model","messages":[{"role":"user","content":"provider failure"}]}`),
		)
		request.Header.Set("X-Request-Id", "request_gateway_provider_failure_without_usage")
		setGatewayRequestDevHeaders(request, "gateway_requests:read")
		response := httptest.NewRecorder()
		server.httpServer.Handler.ServeHTTP(response, request)

		record, found, err := server.gatewayRequestStore().ReadRequest(
			gatewayRequestTestContext(), "request_gateway_provider_failure_without_usage",
		)
		if response.Code < 400 || err != nil || !found || record.Status != GatewayRequestStatusFailed ||
			record.Usage.Availability != GatewayRequestUsageNotReported ||
			record.CostEstimate.Availability != GatewayRequestCostUsageNotReported ||
			record.CostEstimate.EstimatedCostMicros != nil {
			t.Fatalf("provider failure cost evidence drifted: status=%d found=%v record=%+v err=%v", response.Code, found, record, err)
		}
	})
}
