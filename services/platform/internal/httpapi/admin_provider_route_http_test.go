package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"radishmind.local/services/platform/internal/bridge"
	"radishmind.local/services/platform/internal/config"
)

type adminProviderRouteHTTPFixture struct {
	server     *Server
	handler    http.Handler
	repository *memoryAdminProviderRouteRepository
	bridge     *fakeBridge
	auth       controlPlaneReadAuthContext
}

func TestAdminProviderRoutePermissionsProjectIndependently(t *testing.T) {
	grants := projectControlPlaneReadPermissions([]string{
		"radishmind.admin-provider-routes.read",
		"radishmind.admin-provider-routes.draft",
		"radishmind.admin-provider-routes.review",
		"radishmind.admin-provider-routes.activate",
	})
	for _, expected := range []string{
		"admin_provider_routes:read",
		"admin_provider_routes:draft",
		"admin_provider_routes:review",
		"admin_provider_routes:activate",
	} {
		if !controlPlaneReadHasScope(grants, expected) {
			t.Fatalf("admin provider route permission did not project %q: %#v", expected, grants)
		}
	}
	if controlPlaneReadHasScope(grants, "applications:write") ||
		controlPlaneReadHasScope(grants, "agent_copilot_profiles:write") {
		t.Fatalf("admin provider route permissions implied unrelated authority: %#v", grants)
	}
}

func TestBridgeAdminProviderInventoryResolverUsesSanitizedDeterministicSnapshot(t *testing.T) {
	testBridge := &fakeBridge{inventory: bridge.ProviderInventory{Profiles: []bridge.ProviderProfileDescription{
		adminProviderRouteBridgeProfile(),
	}}}
	resolver := bridgeAdminProviderInventoryResolver{bridge: testBridge}
	first, err := resolver.ResolveProviderProfile(
		t.Context(), "test", "mock", "ref:radishmind/test/provider-profiles/mock-primary",
	)
	if err != nil {
		t.Fatalf("resolve inventory binding: %v", err)
	}
	if !first.Enabled || first.ProfileID != "mock-primary" ||
		strings.Join(first.Capabilities, ",") != "chat_completions,messages,responses" ||
		!strings.HasPrefix(first.InventoryDigest, "sha256:") {
		t.Fatalf("unexpected inventory binding: %#v", first)
	}
	fallbackProfile := adminProviderRouteBridgeProfile()
	fallbackProfile.Profile = "mock-backup"
	fallbackProfile.NormalizedProfile = "mock-backup"
	fallbackProfile.Active = false
	fallbackProfile.Fallback = true
	fallbackProfile.ChainIndex = 1
	testBridge.inventory.Profiles = []bridge.ProviderProfileDescription{fallbackProfile}
	fallback, err := resolver.ResolveProviderProfile(
		t.Context(), "test", "mock", "ref:radishmind/test/provider-profiles/mock-backup",
	)
	if err != nil || !fallback.Enabled {
		t.Fatalf("enabled fallback profile was not routable: %#v err=%v", fallback, err)
	}

	reordered := adminProviderRouteBridgeProfile()
	reordered.NorthboundProtocols = []string{"responses", "messages", "chat.completions"}
	testBridge.inventory.Profiles = []bridge.ProviderProfileDescription{reordered}
	second, err := resolver.ResolveProviderProfile(
		t.Context(), "test", "mock", "ref:radishmind/test/provider-profiles/mock-primary",
	)
	if err != nil || second.InventoryDigest != first.InventoryDigest {
		t.Fatalf("equivalent inventory ordering changed digest: first=%#v second=%#v err=%v", first, second, err)
	}

	reordered.ResolvedModel = "mock-model-v2"
	testBridge.inventory.Profiles = []bridge.ProviderProfileDescription{reordered}
	drifted, err := resolver.ResolveProviderProfile(
		t.Context(), "test", "mock", "ref:radishmind/test/provider-profiles/mock-primary",
	)
	if err != nil || drifted.InventoryDigest == first.InventoryDigest {
		t.Fatalf("runtime model drift was not reflected in inventory digest: %#v err=%v", drifted, err)
	}

	if _, err := resolver.ResolveProviderProfile(
		t.Context(), "test", "missing", "ref:radishmind/test/provider-profiles/mock-primary",
	); !errors.Is(err, errAdminProviderRouteInventoryNotFound) {
		t.Fatalf("missing provider profile did not fail closed: %v", err)
	}
	testBridge.inventory.Profiles = []bridge.ProviderProfileDescription{
		adminProviderRouteBridgeProfile(), adminProviderRouteBridgeProfile(),
	}
	if _, err := resolver.ResolveProviderProfile(
		t.Context(), "test", "mock", "ref:radishmind/test/provider-profiles/mock-primary",
	); !errors.Is(err, errAdminProviderRouteInventoryUnavailable) {
		t.Fatalf("duplicate provider profile was not treated as ambiguous inventory: %v", err)
	}
	testBridge.inventoryErr = errors.New("bridge unavailable")
	if _, err := resolver.ResolveProviderProfile(
		t.Context(), "test", "mock", "ref:radishmind/test/provider-profiles/mock-primary",
	); !errors.Is(err, errAdminProviderRouteInventoryUnavailable) {
		t.Fatalf("bridge failure did not fail closed: %v", err)
	}
}

func TestAdminProviderRouteHTTPStatusMapping(t *testing.T) {
	for _, test := range []struct {
		failure string
		success int
		want    int
	}{
		{success: http.StatusCreated, want: http.StatusCreated},
		{failure: "identity_context_missing", success: http.StatusOK, want: http.StatusUnauthorized},
		{failure: AdminProviderRouteFailureScopeDenied, success: http.StatusOK, want: http.StatusForbidden},
		{failure: AdminProviderRouteFailurePayloadInvalid, success: http.StatusOK, want: http.StatusBadRequest},
		{failure: AdminProviderRouteFailureDraftNotFound, success: http.StatusOK, want: http.StatusNotFound},
		{failure: AdminProviderRouteFailureDraftRevisionConflict, success: http.StatusOK, want: http.StatusConflict},
		{failure: AdminProviderRouteFailureCandidateNotApproved, success: http.StatusOK, want: http.StatusUnprocessableEntity},
		{failure: AdminProviderRouteFailureInventoryUnavailable, success: http.StatusOK, want: http.StatusServiceUnavailable},
		{failure: "unknown_failure", success: http.StatusOK, want: http.StatusInternalServerError},
	} {
		if got := adminProviderRouteHTTPStatus(test.failure, test.success); got != test.want {
			t.Fatalf("status mapping for %q: got=%d want=%d", test.failure, got, test.want)
		}
	}
}

func TestAdminProviderRouteHTTPFullLifecycle(t *testing.T) {
	fixture := newAdminProviderRouteHTTPFixture()
	draftInput := adminProviderRouteTestDraftInput(0, "mock-primary")
	draft := fixture.serve(t, http.MethodPut, "/v1/admin/provider-route-configurations/gateway-default",
		adminProviderRouteDraftPutBody{
			ExpectedRevision: draftInput.ExpectedRevision,
			DisplayName:      draftInput.DisplayName, ProviderProfiles: draftInput.ProviderProfiles,
			ModelRoutes: draftInput.ModelRoutes,
		}, fixture.auth, http.StatusOK)
	if draft.Draft == nil || draft.Draft.DraftRevision != 1 || draft.FailureCode != nil {
		t.Fatalf("unexpected draft response: %#v", draft)
	}

	candidate := fixture.serve(t, http.MethodPost,
		"/v1/admin/provider-route-configurations/gateway-default/candidates",
		adminProviderRouteCandidateCreateBody{CandidateID: "candidate-one", ExpectedDraftRevision: 1},
		fixture.auth, http.StatusCreated)
	if candidate.Candidate == nil || candidate.Candidate.CandidateState != adminProviderRouteCandidatePending {
		t.Fatalf("unexpected candidate response: %#v", candidate)
	}
	review := fixture.serve(t, http.MethodPost,
		"/v1/admin/provider-route-configurations/gateway-default/candidates/candidate-one/reviews",
		adminProviderRouteReviewBody{
			ExpectedReviewVersion: 0, Decision: "approve",
			Reason: "Reviewed runtime inventory and model route assignment.",
		}, fixture.auth, http.StatusOK)
	if review.Candidate == nil || review.Candidate.CandidateState != adminProviderRouteCandidateApproved {
		t.Fatalf("unexpected review response: %#v", review)
	}
	activation := fixture.serve(t, http.MethodPost,
		"/v1/admin/provider-route-configurations/gateway-default/candidates/candidate-one/activations",
		adminProviderRouteActivationBody{
			ExpectedGeneration: 0, Action: "activate",
			Reason: "Enable the reviewed development route configuration.",
		}, fixture.auth, http.StatusOK)
	if activation.Snapshot == nil || activation.Snapshot.Generation != 1 ||
		activation.Activation == nil || activation.FailureCode != nil {
		t.Fatalf("unexpected activation response: %#v", activation)
	}
	snapshot := fixture.serve(t, http.MethodGet,
		"/v1/admin/provider-route-configurations/gateway-default/active-snapshot",
		nil, fixture.auth, http.StatusOK)
	if snapshot.Snapshot == nil || snapshot.Snapshot.CandidateID != "candidate-one" {
		t.Fatalf("unexpected active snapshot response: %#v", snapshot)
	}
	history := fixture.serve(t, http.MethodGet,
		"/v1/admin/provider-route-configurations/gateway-default/activation-history",
		nil, fixture.auth, http.StatusOK)
	if len(history.ActivationHistory) != 1 ||
		history.ActivationHistory[0].AfterCandidateID != "candidate-one" {
		t.Fatalf("unexpected activation history: %#v", history)
	}
}

func TestAdminProviderRouteHTTPRejectsBeforeRepositoryMutation(t *testing.T) {
	tests := []struct {
		name        string
		mutateAuth  func(*controlPlaneReadAuthContext)
		workspace   string
		environment string
		wantStatus  int
		wantFailure string
	}{
		{
			name: "identity missing",
			mutateAuth: func(auth *controlPlaneReadAuthContext) {
				*auth = controlPlaneReadAuthContext{}
			},
			wantStatus: http.StatusUnauthorized, wantFailure: "identity_context_missing",
		},
		{
			name:       "scope denied",
			mutateAuth: func(auth *controlPlaneReadAuthContext) { auth.ScopeGrants = nil },
			wantStatus: http.StatusForbidden, wantFailure: "scope_denied",
		},
		{
			name: "OIDC membership unavailable",
			mutateAuth: func(auth *controlPlaneReadAuthContext) {
				auth.AuthMode = controlPlaneReadAuthModeRadishOIDCIntegrationTest
			},
			wantStatus: http.StatusServiceUnavailable, wantFailure: "workspace_membership_unavailable",
		},
		{
			name: "verified identity mismatch",
			mutateAuth: func(auth *controlPlaneReadAuthContext) {
				auth.VerifiedIdentity.TenantRef = "tenant-other"
			},
			wantStatus: http.StatusUnauthorized, wantFailure: "auth_context_contract_mismatch",
		},
		{
			name:      "workspace invalid",
			workspace: "not allowed", wantStatus: http.StatusForbidden,
			wantFailure: AdminProviderRouteFailureScopeDenied,
		},
		{
			name:        "production environment forbidden",
			environment: "production", wantStatus: http.StatusForbidden,
			wantFailure: AdminProviderRouteFailureEnvironmentForbidden,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newAdminProviderRouteHTTPFixture()
			auth := fixture.auth
			verifiedIdentity := *auth.VerifiedIdentity
			auth.VerifiedIdentity = &verifiedIdentity
			if test.mutateAuth != nil {
				test.mutateAuth(&auth)
			}
			workspace := test.workspace
			if workspace == "" {
				workspace = "workspace-alpha"
			}
			environment := test.environment
			if environment == "" {
				environment = "test"
			}
			body := adminProviderRouteTestDraftInput(0, "mock-primary")
			requestBody, _ := json.Marshal(adminProviderRouteDraftPutBody{
				ExpectedRevision: body.ExpectedRevision, DisplayName: body.DisplayName,
				ProviderProfiles: body.ProviderProfiles, ModelRoutes: body.ModelRoutes,
			})
			request := httptest.NewRequest(http.MethodPut,
				"/v1/admin/provider-route-configurations/gateway-default", strings.NewReader(string(requestBody)))
			request.Header.Set("Content-Type", "application/json")
			request.Header.Set(adminProviderRouteDevWorkspaceHeader, workspace)
			request.Header.Set(adminProviderRouteDevEnvironmentHeader, environment)
			request = request.WithContext(withControlPlaneReadFakeAuthContext(request.Context(), auth))
			recorder := httptest.NewRecorder()

			fixture.handler.ServeHTTP(recorder, request)

			envelope := decodeAdminProviderRouteEnvelope(t, recorder, test.wantStatus)
			if envelope.FailureCode == nil || *envelope.FailureCode != test.wantFailure {
				t.Fatalf("unexpected failure envelope: %#v", envelope)
			}
			if len(fixture.repository.drafts) != 0 || fixture.bridge.inventoryCalls.Load() != 0 {
				t.Fatalf("denied request reached repository or bridge: drafts=%d inventory_calls=%d",
					len(fixture.repository.drafts), fixture.bridge.inventoryCalls.Load())
			}
		})
	}
}

func TestAdminProviderRouteHTTPGatesStrictPayloadAndConflictStatus(t *testing.T) {
	fixture := newAdminProviderRouteHTTPFixture()
	fixture.server.config.AdminProviderRouteDevHTTPEnabled = false
	recorder := fixture.rawServe(t, http.MethodPut,
		"/v1/admin/provider-route-configurations/gateway-default", `{}`, fixture.auth)
	if recorder.Code != http.StatusForbidden ||
		!strings.Contains(recorder.Body.String(), "ADMIN_PROVIDER_ROUTE_DEV_HTTP_DISABLED") ||
		recorder.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("HTTP gate did not fail closed: status=%d headers=%v body=%s",
			recorder.Code, recorder.Header(), recorder.Body.String())
	}
	fixture.server.config.AdminProviderRouteDevHTTPEnabled = true
	fixture.server.config.AdminProviderRouteDevWriteEnabled = false
	recorder = fixture.rawServe(t, http.MethodPut,
		"/v1/admin/provider-route-configurations/gateway-default", `{}`, fixture.auth)
	writeDisabled := decodeAdminProviderRouteEnvelope(t, recorder, http.StatusForbidden)
	if writeDisabled.FailureCode == nil ||
		*writeDisabled.FailureCode != AdminProviderRouteFailureDisabled ||
		len(fixture.repository.drafts) != 0 {
		t.Fatalf("write gate did not fail before mutation: %#v", writeDisabled)
	}
	fixture.server.config.AdminProviderRouteDevWriteEnabled = true

	recorder = fixture.rawServe(t, http.MethodPut,
		"/v1/admin/provider-route-configurations/gateway-default", `{"expected_revision":0,"unknown":true}`,
		fixture.auth)
	if recorder.Code != http.StatusBadRequest || len(fixture.repository.drafts) != 0 {
		t.Fatalf("unknown JSON field reached repository: status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	recorder = fixture.rawServe(t, http.MethodPut,
		"/v1/admin/provider-route-configurations/gateway-default",
		`{"expected_revision":0,"expected_revision":1,"display_name":"duplicate","provider_profiles":[],"model_routes":[]}`,
		fixture.auth)
	if recorder.Code != http.StatusBadRequest || len(fixture.repository.drafts) != 0 {
		t.Fatalf("duplicate JSON field reached repository: status=%d body=%s", recorder.Code, recorder.Body.String())
	}

	input := adminProviderRouteTestDraftInput(0, "mock-primary")
	fixture.serve(t, http.MethodPut, "/v1/admin/provider-route-configurations/gateway-default",
		adminProviderRouteDraftPutBody{
			ExpectedRevision: 0, DisplayName: input.DisplayName,
			ProviderProfiles: input.ProviderProfiles, ModelRoutes: input.ModelRoutes,
		}, fixture.auth, http.StatusOK)
	conflict := fixture.serve(t, http.MethodPut, "/v1/admin/provider-route-configurations/gateway-default",
		adminProviderRouteDraftPutBody{
			ExpectedRevision: 0, DisplayName: input.DisplayName,
			ProviderProfiles: input.ProviderProfiles, ModelRoutes: input.ModelRoutes,
		}, fixture.auth, http.StatusConflict)
	if conflict.FailureCode == nil ||
		*conflict.FailureCode != AdminProviderRouteFailureDraftRevisionConflict ||
		conflict.CurrentDraftRevision != 1 {
		t.Fatalf("revision conflict lost current state: %#v", conflict)
	}

	query := fixture.rawServe(t, http.MethodGet,
		"/v1/admin/provider-route-configurations/gateway-default?unexpected=1", "", fixture.auth)
	if query.Code != http.StatusBadRequest || query.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("query was not rejected deterministically: status=%d body=%s", query.Code, query.Body.String())
	}
}

func newAdminProviderRouteHTTPFixture() adminProviderRouteHTTPFixture {
	repository := newMemoryAdminProviderRouteRepository()
	testBridge := &fakeBridge{inventory: bridge.ProviderInventory{
		Profiles: []bridge.ProviderProfileDescription{adminProviderRouteBridgeProfile()},
	}}
	server := &Server{
		config: config.Config{
			AdminProviderRouteDevHTTPEnabled: true, AdminProviderRouteDevWriteEnabled: true,
		},
		bridge: testBridge, adminProviderRouteRepository: repository,
	}
	mux := http.NewServeMux()
	mux.HandleFunc(adminProviderRouteDraftReadRoute, server.handleReadAdminProviderRouteDraft)
	mux.HandleFunc(adminProviderRouteDraftPutRoute, server.handlePutAdminProviderRouteDraft)
	mux.HandleFunc(adminProviderRouteCandidateCreateRoute, server.handleCreateAdminProviderRouteCandidate)
	mux.HandleFunc(adminProviderRouteCandidateReadRoute, server.handleReadAdminProviderRouteCandidate)
	mux.HandleFunc(adminProviderRouteReviewRoute, server.handleReviewAdminProviderRouteCandidate)
	mux.HandleFunc(adminProviderRouteActivationRoute, server.handleActivateAdminProviderRouteCandidate)
	mux.HandleFunc(adminProviderRouteActiveSnapshotRoute, server.handleReadAdminProviderRouteActiveSnapshot)
	mux.HandleFunc(adminProviderRouteActivationHistoryRoute, server.handleListAdminProviderRouteActivations)
	auth := controlPlaneReadAuthContext{
		AuthMode: controlPlaneReadAuthModeDevHeaders, IdentityContext: "verified:admin-provider-route",
		TenantBinding: "tenant-alpha", SubjectBinding: "subject-admin",
		ScopeGrants: []string{
			"admin_provider_routes:read", "admin_provider_routes:draft",
			"admin_provider_routes:review", "admin_provider_routes:activate",
		},
		VerifiedIdentity: &VerifiedControlPlaneIdentity{
			SubjectRef: "subject-admin", TenantRef: "tenant-alpha",
		},
		ResourceBinding: ControlPlaneResourceBinding{
			TenantRef: "tenant-alpha", TenantVerified: true,
		},
	}
	return adminProviderRouteHTTPFixture{
		server: server, handler: mux, repository: repository, bridge: testBridge, auth: auth,
	}
}

func (fixture adminProviderRouteHTTPFixture) serve(
	t *testing.T,
	method string,
	target string,
	body any,
	auth controlPlaneReadAuthContext,
	wantStatus int,
) adminProviderRouteEnvelope {
	t.Helper()
	var payload string
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		payload = string(raw)
	}
	return decodeAdminProviderRouteEnvelope(t, fixture.rawServe(t, method, target, payload, auth), wantStatus)
}

func (fixture adminProviderRouteHTTPFixture) rawServe(
	t *testing.T,
	method string,
	target string,
	payload string,
	auth controlPlaneReadAuthContext,
) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, target, strings.NewReader(payload))
	if payload != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	request.Header.Set(adminProviderRouteDevWorkspaceHeader, "workspace-alpha")
	request.Header.Set(adminProviderRouteDevEnvironmentHeader, "test")
	request = request.WithContext(withControlPlaneReadFakeAuthContext(request.Context(), auth))
	recorder := httptest.NewRecorder()
	fixture.handler.ServeHTTP(recorder, request)
	if recorder.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("response is cacheable: headers=%v", recorder.Header())
	}
	return recorder
}

func decodeAdminProviderRouteEnvelope(
	t *testing.T,
	recorder *httptest.ResponseRecorder,
	wantStatus int,
) adminProviderRouteEnvelope {
	t.Helper()
	if recorder.Code != wantStatus {
		t.Fatalf("unexpected status: got=%d want=%d body=%s", recorder.Code, wantStatus, recorder.Body.String())
	}
	var envelope adminProviderRouteEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode admin provider route envelope: %v body=%s", err, recorder.Body.String())
	}
	return envelope
}

func adminProviderRouteBridgeProfile() bridge.ProviderProfileDescription {
	return bridge.ProviderProfileDescription{
		Profile: "mock-primary", NormalizedProfile: "mock-primary", ProviderID: "mock",
		ResolvedModel: "mock-model", APIStyle: "openai_compatible",
		HasBaseURL: true, HasAPIKey: true, RequestTimeoutSeconds: 30, Active: true, Enabled: true,
		Capabilities:        map[string]any{"chat": true, "responses": true, "messages": true},
		NorthboundProtocols: []string{"chat.completions", "responses", "messages"},
		NorthboundRoutes:    []string{"/v1/chat/completions", "/v1/responses", "/v1/messages"},
		CredentialState:     "configured", DeploymentMode: "remote", AuthMode: "bearer",
		Streaming: true,
	}
}
