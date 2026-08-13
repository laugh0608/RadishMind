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

type mutableGatewayProviderRouteSnapshotProvider struct {
	snapshot AdminProviderRouteSnapshot
	found    bool
	err      error
	reads    int
}

func (provider *mutableGatewayProviderRouteSnapshotProvider) ReadActiveSnapshot(
	gatewayProviderRouteScope,
) (AdminProviderRouteSnapshot, bool, error) {
	provider.reads++
	return provider.snapshot, provider.found, provider.err
}

func TestGatewayProviderRouteSelectsExactActivatedSnapshotRoutes(t *testing.T) {
	profile := adminProviderRouteBridgeProfile()
	snapshot := gatewayProviderRouteTestSnapshot(t, profile, 3, "candidate-v3")
	provider := &mutableGatewayProviderRouteSnapshotProvider{snapshot: snapshot, found: true}
	testBridge := &fakeBridge{inventory: bridge.ProviderInventory{Profiles: []bridge.ProviderProfileDescription{profile}}}
	server := gatewayProviderRouteTestServer(testBridge, provider)

	for _, test := range []struct {
		protocol string
		model    string
	}{
		{northboundProtocolChatCompletions, "radish-chat"},
		{northboundProtocolResponses, "radish-responses"},
		{northboundProtocolMessages, "radish-messages"},
	} {
		selection, failure := server.resolveGatewayNorthboundSelection(
			t.Context(), gatewayProviderRouteTestRequestContext(t.Context()),
			test.protocol, test.model, nil,
		)
		if failure != "" {
			t.Fatalf("%s selection failed: %s", test.protocol, failure)
		}
		if selection.provider != "mock" || selection.providerProfile != "mock-primary" ||
			selection.model != test.model || selection.upstreamModel != "mock-model" ||
			selection.source != gatewayProviderRouteSelectionSource ||
			selection.routeConfigurationID != "gateway-default" ||
			selection.routeGeneration != 3 ||
			selection.routeSnapshotDigest != snapshot.SnapshotDigest {
			t.Fatalf("%s selected an unexpected route: %#v", test.protocol, selection)
		}
	}
	if provider.reads != 3 || testBridge.inventoryCalls.Load() != 3 {
		t.Fatalf("each request must pin one snapshot and validate one inventory: reads=%d inventory=%d", provider.reads, testBridge.inventoryCalls.Load())
	}
}

func TestGatewayProviderRouteSnapshotProviderReadsOnlyActiveScopedSnapshot(t *testing.T) {
	profile := adminProviderRouteBridgeProfile()
	snapshot := gatewayProviderRouteTestSnapshot(t, profile, 2, "candidate-v2")
	repository := newMemoryAdminProviderRouteRepository()
	adminContext := AdminProviderRouteContext{
		RequestContext: t.Context(),
		TenantRef:      snapshot.TenantRef,
		WorkspaceID:    snapshot.WorkspaceID,
		Environment:    snapshot.Environment,
	}
	repository.snapshots[adminProviderRouteConfigurationKey(adminContext, snapshot.ConfigurationID)] = snapshot
	provider := adminProviderRouteSnapshotProvider{repository: repository}
	scope := gatewayProviderRouteScope{
		RequestContext: t.Context(), RequestID: "req-consumer",
		TenantRef: snapshot.TenantRef, WorkspaceID: snapshot.WorkspaceID,
		Environment: snapshot.Environment, ConfigurationID: snapshot.ConfigurationID,
		ActorRef: "subject-a",
	}
	read, found, err := provider.ReadActiveSnapshot(scope)
	if err != nil || !found || read.Generation != 2 || read.SnapshotDigest != snapshot.SnapshotDigest {
		t.Fatalf("read active scoped snapshot: found=%v snapshot=%#v err=%v", found, read, err)
	}
	scope.WorkspaceID = "workspace-other"
	if _, found, err = provider.ReadActiveSnapshot(scope); err != nil || found {
		t.Fatalf("snapshot provider leaked a different workspace: found=%v err=%v", found, err)
	}
	repository.unavailable = true
	scope.WorkspaceID = snapshot.WorkspaceID
	if _, _, err = provider.ReadActiveSnapshot(scope); !errors.Is(err, errGatewayProviderRouteSnapshotUnavailable) {
		t.Fatalf("snapshot store failure did not fail closed: %v", err)
	}
}

func TestGatewayProviderRouteFailsClosedBeforeProviderInvocation(t *testing.T) {
	profile := adminProviderRouteBridgeProfile()
	snapshot := gatewayProviderRouteTestSnapshot(t, profile, 1, "candidate-v1")
	tests := []struct {
		name        string
		model       string
		extension   *chatCompletionExtension
		found       bool
		providerErr error
		mutate      func(*fakeBridge)
		wantFailure string
	}{
		{
			name: "snapshot missing", model: "radish-chat", found: false,
			wantFailure: gatewayProviderRouteFailureSnapshotUnavailable,
		},
		{
			name: "snapshot store unavailable", model: "radish-chat", found: true,
			providerErr: errors.New("store unavailable"),
			wantFailure: gatewayProviderRouteFailureSnapshotUnavailable,
		},
		{
			name: "route missing", model: "not-routed", found: true,
			wantFailure: gatewayProviderRouteFailureNotFound,
		},
		{
			name: "override forbidden", model: "radish-chat", found: true,
			extension:   &chatCompletionExtension{Provider: "other"},
			wantFailure: gatewayProviderRouteFailureOverrideForbidden,
		},
		{
			name: "inventory missing", model: "radish-chat", found: true,
			mutate: func(testBridge *fakeBridge) {
				testBridge.inventory.Profiles = nil
			},
			wantFailure: gatewayProviderRouteFailureInventoryMismatch,
		},
		{
			name: "inventory drift", model: "radish-chat", found: true,
			mutate: func(testBridge *fakeBridge) {
				testBridge.inventory.Profiles[0].ResolvedModel = "drifted-model"
			},
			wantFailure: gatewayProviderRouteFailureInventoryMismatch,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testBridge := &fakeBridge{inventory: bridge.ProviderInventory{Profiles: []bridge.ProviderProfileDescription{profile}}}
			if test.mutate != nil {
				test.mutate(testBridge)
			}
			provider := &mutableGatewayProviderRouteSnapshotProvider{
				snapshot: snapshot, found: test.found, err: test.providerErr,
			}
			server := gatewayProviderRouteTestServer(testBridge, provider)
			selection, failure := server.resolveGatewayNorthboundSelection(
				t.Context(), gatewayProviderRouteTestRequestContext(t.Context()),
				northboundProtocolChatCompletions, test.model, test.extension,
			)
			if failure != test.wantFailure {
				t.Fatalf("unexpected failure: got=%q want=%q selection=%#v", failure, test.wantFailure, selection)
			}
			if testBridge.handleCalls != 0 || testBridge.streamCalled {
				t.Fatalf("failed route reached Provider bridge: handle=%d stream=%v", testBridge.handleCalls, testBridge.streamCalled)
			}
		})
	}
}

func TestGatewayProviderRoutePinsSnapshotForInFlightRequest(t *testing.T) {
	oldProfile := adminProviderRouteBridgeProfile()
	oldSnapshot := gatewayProviderRouteTestSnapshot(t, oldProfile, 4, "candidate-v4")
	provider := &mutableGatewayProviderRouteSnapshotProvider{snapshot: oldSnapshot, found: true}
	testBridge := &fakeBridge{
		inventory: bridge.ProviderInventory{Profiles: []bridge.ProviderProfileDescription{oldProfile}},
		envelope:  bridge.GatewayEnvelope{Status: "ok"},
	}
	server := gatewayProviderRouteTestServer(testBridge, provider)

	selection, failure := server.resolveGatewayNorthboundSelection(
		t.Context(), gatewayProviderRouteTestRequestContext(t.Context()),
		northboundProtocolChatCompletions, "radish-chat", nil,
	)
	if failure != "" {
		t.Fatalf("pin old snapshot: %s", failure)
	}

	newProfile := oldProfile
	newProfile.ResolvedModel = "mock-model-v2"
	provider.snapshot = gatewayProviderRouteTestSnapshot(t, newProfile, 5, "candidate-v5")
	testBridge.inventory = bridge.ProviderInventory{Profiles: []bridge.ProviderProfileDescription{newProfile}}

	if _, err := testBridge.HandleEnvelope(
		context.Background(), []byte(`{}`), server.buildBridgeEnvelopeOptions(selection, 0),
	); err != nil {
		t.Fatalf("invoke pinned selection: %v", err)
	}
	if provider.reads != 1 || testBridge.lastOptions.Model != "mock-model" ||
		selection.routeGeneration != 4 || selection.routeSnapshotDigest != oldSnapshot.SnapshotDigest {
		t.Fatalf("in-flight request was not pinned: reads=%d selection=%#v options=%#v", provider.reads, selection, testBridge.lastOptions)
	}

	next, failure := server.resolveGatewayNorthboundSelection(
		t.Context(), gatewayProviderRouteTestRequestContext(t.Context()),
		northboundProtocolChatCompletions, "radish-chat", nil,
	)
	if failure != "" || next.routeGeneration != 5 || next.upstreamModel != "mock-model-v2" {
		t.Fatalf("next request did not observe activated generation: selection=%#v failure=%s", next, failure)
	}
}

func TestGatewayProviderRouteHandlersInvokePinnedRoutesAndRecordLineage(t *testing.T) {
	profile := adminProviderRouteBridgeProfile()
	snapshot := gatewayProviderRouteTestSnapshot(t, profile, 6, "candidate-v6")
	snapshot.TenantRef = "tenant_demo"
	snapshot.WorkspaceID = "workspace_demo"
	provider := &mutableGatewayProviderRouteSnapshotProvider{snapshot: snapshot, found: true}
	testBridge := &fakeBridge{
		inventory: bridge.ProviderInventory{Profiles: []bridge.ProviderProfileDescription{profile}},
		envelope: bridge.GatewayEnvelope{
			SchemaVersion: 1, Status: "ok", RequestID: "bridge-provider-route",
			Project: "radish", Task: "answer_docs_question",
			Response: map[string]any{"summary": "routed response"},
			Metadata: map[string]any{},
		},
	}
	store := newMemoryGatewayRequestStore(16)
	server := &Server{
		bridge: testBridge,
		config: config.Config{
			BridgeTimeout:                       time.Second,
			GatewayRequestHistoryDevEnabled:     true,
			GatewayProviderRouteSource:          "admin_snapshot_dev_test",
			GatewayProviderRouteEnvironment:     "test",
			GatewayProviderRouteConfigurationID: "gateway-default",
		},
		providerRouteSnapshotProvider:  provider,
		gatewayRequestHistoryStore:     store,
		gatewayRequestHistoryStoreMode: gatewayRequestStoreModeMemoryDev,
	}
	tests := []struct {
		path    string
		body    string
		handler func(http.ResponseWriter, *http.Request)
	}{
		{"/v1/chat/completions", `{"model":"radish-chat","messages":[{"role":"user","content":"hello"}]}`, server.handleChatCompletions},
		{"/v1/responses", `{"model":"radish-responses","input":"hello"}`, server.handleResponses},
		{"/v1/messages", `{"model":"radish-messages","messages":[{"role":"user","content":"hello"}]}`, server.handleMessages},
	}
	for _, test := range tests {
		request := httptest.NewRequest(http.MethodPost, test.path, strings.NewReader(test.body))
		setGatewayRequestDevHeaders(request, "gateway_requests:read")
		response := httptest.NewRecorder()
		test.handler(response, request)
		if response.Code != http.StatusOK {
			t.Fatalf("%s failed: status=%d body=%s", test.path, response.Code, response.Body.String())
		}
	}
	if testBridge.handleCalls != 3 || provider.reads != 3 {
		t.Fatalf("handlers did not invoke one pinned route per request: handle=%d reads=%d", testBridge.handleCalls, provider.reads)
	}
	requestContext := GatewayRequestContext{
		RequestContext: t.Context(), TenantRef: "tenant_demo", WorkspaceID: "workspace_demo",
		ConsumerRef: "consumer_demo", ApplicationID: "application_demo", SubjectRef: "subject_demo",
		AuditContext: "audit_context_demo", Source: "dev_headers",
	}
	page, err := store.ListRequests(requestContext, GatewayRequestListFilter{Limit: 10})
	if err != nil || len(page.Records) != 3 {
		t.Fatalf("list routed request history: records=%#v err=%v", page.Records, err)
	}
	for _, record := range page.Records {
		if record.ProviderRouteConfigurationID != "gateway-default" ||
			record.ProviderRouteGeneration != 6 ||
			record.ProviderRouteSnapshotDigest != snapshot.SnapshotDigest {
			t.Fatalf("handler request history lost snapshot lineage: %#v", record)
		}
	}
}

func TestGatewayRequestHistoryStoresProviderRouteLineage(t *testing.T) {
	requestContext := gatewayProviderRouteTestRequestContext(t.Context())
	store := newMemoryGatewayRequestStore(4)
	record := GatewayRequestRecord{
		SchemaVersion: gatewayRequestRecordSchemaVersion,
		StoreMode:     gatewayRequestStoreModeMemoryDev,
		RequestID:     requestContext.RequestID,
		AuditRef:      requestContext.AuditRef,
		TenantRef:     requestContext.TenantRef,
		WorkspaceID:   requestContext.WorkspaceID,
		ConsumerRef:   requestContext.ConsumerRef,
		ApplicationID: requestContext.ApplicationID,
		SubjectRef:    requestContext.SubjectRef,
		Route:         "/v1/chat/completions",
		Protocol:      northboundProtocolChatCompletions,
		Status:        GatewayRequestStatusStarted,
		StartedAt:     time.Now().UTC().Format(time.RFC3339Nano),
		Usage:         GatewayRequestUsage{Availability: GatewayRequestUsageNotReported},
	}
	if err := store.CreateRequest(requestContext, &record); err != nil {
		t.Fatalf("create Gateway request: %v", err)
	}
	trace := requestTrace{hasSelection: true, selection: northboundSelection{
		provider: "mock", providerProfile: "mock-primary", model: "radish-chat",
		source: gatewayProviderRouteSelectionSource, routeConfigurationID: "gateway-default",
		routeGeneration: 7, routeSnapshotDigest: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}}
	applyGatewayRequestSelection(&record, trace)
	if err := store.UpdateRequest(requestContext, &record); err != nil {
		t.Fatalf("store Gateway request lineage: %v", err)
	}
	stored, found, err := store.ReadRequest(requestContext, record.RequestID)
	if err != nil || !found {
		t.Fatalf("read Gateway request lineage: found=%v err=%v", found, err)
	}
	if stored.ProviderRouteConfigurationID != "gateway-default" ||
		stored.ProviderRouteGeneration != 7 ||
		stored.ProviderRouteSnapshotDigest != trace.selection.routeSnapshotDigest {
		t.Fatalf("Gateway request lineage was not retained: %#v", stored)
	}
}

func gatewayProviderRouteTestServer(
	testBridge *fakeBridge,
	provider gatewayProviderRouteSnapshotProvider,
) *Server {
	return &Server{
		bridge: testBridge,
		config: config.Config{
			BridgeTimeout:                       time.Second,
			GatewayProviderRouteSource:          "admin_snapshot_dev_test",
			GatewayProviderRouteEnvironment:     "test",
			GatewayProviderRouteConfigurationID: "gateway-default",
		},
		providerRouteSnapshotProvider: provider,
	}
}

func gatewayProviderRouteTestRequestContext(ctx context.Context) GatewayRequestContext {
	return GatewayRequestContext{
		RequestContext: ctx,
		TenantRef:      "tenant-a",
		WorkspaceID:    "workspace-a",
		ConsumerRef:    "api_key:key_aaaaaaaaaaaaaaaa",
		ApplicationID:  "app-a",
		SubjectRef:     "subject-a",
		ScopeGrants:    []string{"chat:invoke", "responses:invoke", "messages:invoke"},
		AuditContext:   "api-key-dev-test",
		Source:         gatewayAPIKeyAuthenticationSource,
		RequestID:      "req-provider-route",
		AuditRef:       "audit_req-provider-route_gateway-request",
	}
}

func gatewayProviderRouteTestSnapshot(
	t *testing.T,
	profile bridge.ProviderProfileDescription,
	generation int,
	candidateID string,
) AdminProviderRouteSnapshot {
	t.Helper()
	assignment := AdminProviderProfileAssignment{
		ProfileID: "primary", DisplayName: "Mock Primary", ProviderID: "mock",
		RuntimeProfileRef: "ref:radishmind/test/provider-profiles/mock-primary",
		Capabilities:      []string{"chat_completions", "messages", "responses"},
	}
	binding, err := adminProviderRouteInventoryBindingFromProfile(
		"test", assignment.ProviderID, assignment.RuntimeProfileRef, profile,
	)
	if err != nil {
		t.Fatalf("build inventory binding: %v", err)
	}
	binding.ProfileID = assignment.ProfileID
	configuration, failureCode := normalizeAdminProviderRouteConfiguration(
		"Gateway Default",
		[]AdminProviderProfileAssignment{assignment},
		[]AdminModelRouteDefinition{
			{RouteID: "chat", Protocol: "chat_completions", ModelID: "radish-chat", ProviderProfileID: assignment.ProfileID},
			{RouteID: "responses", Protocol: "responses", ModelID: "radish-responses", ProviderProfileID: assignment.ProfileID},
			{RouteID: "messages", Protocol: "messages", ModelID: "radish-messages", ProviderProfileID: assignment.ProfileID},
		},
		"test",
	)
	if failureCode != "" {
		t.Fatalf("normalize snapshot configuration: %s", failureCode)
	}
	snapshot := AdminProviderRouteSnapshot{
		SchemaVersion:       adminProviderRouteSnapshotSchemaVersion,
		TenantRef:           "tenant-a",
		WorkspaceID:         "workspace-a",
		Environment:         "test",
		ConfigurationID:     "gateway-default",
		Generation:          generation,
		CandidateID:         candidateID,
		CandidateDigest:     "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		Configuration:       configuration,
		InventoryBindings:   []AdminProviderInventoryBinding{binding},
		ActivatedAt:         time.Now().UTC().Format(time.RFC3339Nano),
		ActivatedByActorRef: "admin-a",
		RequestID:           "req-activation",
		AuditRef:            "audit-activation",
	}
	digest, err := adminProviderRouteCanonicalDigest(struct {
		ConfigurationID string                                  `json:"configuration_id"`
		CandidateID     string                                  `json:"candidate_id"`
		CandidateDigest string                                  `json:"candidate_digest"`
		Configuration   AdminProviderRouteConfigurationSnapshot `json:"configuration"`
		Bindings        []AdminProviderInventoryBinding         `json:"inventory_bindings"`
	}{
		ConfigurationID: snapshot.ConfigurationID,
		CandidateID:     snapshot.CandidateID,
		CandidateDigest: snapshot.CandidateDigest,
		Configuration:   snapshot.Configuration,
		Bindings:        snapshot.InventoryBindings,
	})
	if err != nil {
		t.Fatalf("build snapshot digest: %v", err)
	}
	snapshot.SnapshotDigest = digest
	return snapshot
}
