package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"radishmind.local/services/platform/internal/config"
)

type controlPlaneReadEnvelopeForTest struct {
	RequestID   string           `json:"request_id"`
	TenantRef   string           `json:"tenant_ref"`
	Items       []map[string]any `json:"items"`
	NextCursor  *string          `json:"next_cursor"`
	FailureCode *string          `json:"failure_code"`
	AuditRef    string           `json:"audit_ref"`
}

func TestControlPlaneReadFakeStoreRoutes(t *testing.T) {
	server := NewServer(config.Config{}, Options{BuildVersion: "test"})
	server.controlPlaneReadStore = newControlPlaneReadFakeStore()

	t.Run("tenant summary succeeds with test-only auth context", func(t *testing.T) {
		req := newControlPlaneReadRequest(
			http.MethodGet,
			"/v1/control-plane/tenants/tenant_demo/summary",
			controlPlaneReadTestAuth("tenant_demo", "tenant:read"),
		)
		req.Header.Set("X-Request-Id", "req-tenant-summary")
		rec := httptest.NewRecorder()

		server.httpServer.Handler.ServeHTTP(rec, req)

		envelope := decodeControlPlaneReadEnvelope(t, rec, http.StatusOK)
		if envelope.RequestID != "req-tenant-summary" || envelope.TenantRef != "tenant_demo" {
			t.Fatalf("unexpected envelope identity: %#v", envelope)
		}
		if envelope.FailureCode != nil || len(envelope.Items) != 1 {
			t.Fatalf("unexpected tenant summary response: %#v", envelope)
		}
		if got := envelope.Items[0]["tenant_display_name"]; got != "Demo Tenant" {
			t.Fatalf("unexpected tenant display name: %#v", got)
		}
		assertControlPlaneReadNoForbiddenPayload(t, rec.Body.String())
	})

	t.Run("quota summary succeeds with test-only auth context", func(t *testing.T) {
		req := newControlPlaneReadRequest(
			http.MethodGet,
			"/v1/user-workspace/usage/quota-summary",
			controlPlaneReadTestAuth("tenant_demo", "usage:read"),
		)
		req.Header.Set("X-Request-Id", "req-quota-summary")
		rec := httptest.NewRecorder()

		server.httpServer.Handler.ServeHTTP(rec, req)

		envelope := decodeControlPlaneReadEnvelope(t, rec, http.StatusOK)
		if envelope.RequestID != "req-quota-summary" || envelope.TenantRef != "tenant_demo" {
			t.Fatalf("unexpected envelope identity: %#v", envelope)
		}
		if envelope.FailureCode != nil || len(envelope.Items) != 1 {
			t.Fatalf("unexpected quota summary response: %#v", envelope)
		}
		if got := envelope.Items[0]["quota_id"]; got != "quota_demo_current" {
			t.Fatalf("unexpected quota id: %#v", got)
		}
		assertControlPlaneReadNoForbiddenPayload(t, rec.Body.String())
	})

	t.Run("cursor list routes succeed with test-only auth context", func(t *testing.T) {
		for _, tc := range []struct {
			name           string
			path           string
			scope          string
			requestID      string
			itemField      string
			itemValue      string
			expectedCursor *string
		}{
			{
				name:           "application summaries",
				path:           "/v1/user-workspace/applications?application_kind=workflow",
				scope:          "applications:read",
				requestID:      "req-applications",
				itemField:      "application_ref",
				itemValue:      "app_demo_workflow",
				expectedCursor: stringPointerForTest("cursor_apps_next"),
			},
			{
				name:      "api key summaries",
				path:      "/v1/user-workspace/api-keys?scope=chat:invoke",
				scope:     "api_keys:read",
				requestID: "req-api-keys",
				itemField: "api_key_id",
				itemValue: "ak_demo_001",
			},
			{
				name:           "workflow definition summaries",
				path:           "/v1/user-workspace/workflow-definitions?risk_level=medium",
				scope:          "applications:read",
				requestID:      "req-workflow-definitions",
				itemField:      "workflow_definition_id",
				itemValue:      "wf_demo_v1",
				expectedCursor: stringPointerForTest("cursor_workflows_next"),
			},
			{
				name:           "run record summaries",
				path:           "/v1/user-workspace/runs?status=succeeded",
				scope:          "runs:read",
				requestID:      "req-runs",
				itemField:      "run_id",
				itemValue:      "run_demo_001",
				expectedCursor: stringPointerForTest("cursor_runs_next"),
			},
			{
				name:           "audit summaries",
				path:           "/v1/control-plane/audit?event_kind=read_model_accessed",
				scope:          "audit:read",
				requestID:      "req-audit",
				itemField:      "audit_ref",
				itemValue:      "audit_read_runs_001",
				expectedCursor: stringPointerForTest("cursor_audit_next"),
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				req := newControlPlaneReadRequest(
					http.MethodGet,
					tc.path,
					controlPlaneReadTestAuth("tenant_demo", tc.scope),
				)
				req.Header.Set("X-Request-Id", tc.requestID)
				rec := httptest.NewRecorder()

				server.httpServer.Handler.ServeHTTP(rec, req)

				envelope := decodeControlPlaneReadEnvelope(t, rec, http.StatusOK)
				if envelope.RequestID != tc.requestID || envelope.TenantRef != "tenant_demo" {
					t.Fatalf("unexpected envelope identity: %#v", envelope)
				}
				if envelope.FailureCode != nil || len(envelope.Items) != 1 {
					t.Fatalf("unexpected cursor list response: %#v", envelope)
				}
				if got := envelope.Items[0][tc.itemField]; got != tc.itemValue {
					t.Fatalf("unexpected %s: %#v", tc.itemField, got)
				}
				assertControlPlaneReadNextCursor(t, envelope.NextCursor, tc.expectedCursor)
				assertControlPlaneReadNoForbiddenPayload(t, rec.Body.String())
			})
		}
	})

	t.Run("missing identity fails closed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/control-plane/tenants/tenant_demo/summary", nil)
		req.Header.Set("X-Request-Id", "req-missing-identity")
		rec := httptest.NewRecorder()

		server.httpServer.Handler.ServeHTTP(rec, req)

		envelope := decodeControlPlaneReadEnvelope(t, rec, http.StatusOK)
		assertControlPlaneReadFailure(t, envelope, "identity_context_missing")
	})

	t.Run("tenant binding mismatch fails closed", func(t *testing.T) {
		req := newControlPlaneReadRequest(
			http.MethodGet,
			"/v1/control-plane/tenants/tenant_demo/summary",
			controlPlaneReadTestAuth("tenant_other", "tenant:read"),
		)
		req.Header.Set("X-Request-Id", "req-tenant-binding-missing")
		rec := httptest.NewRecorder()

		server.httpServer.Handler.ServeHTTP(rec, req)

		envelope := decodeControlPlaneReadEnvelope(t, rec, http.StatusOK)
		assertControlPlaneReadFailure(t, envelope, "tenant_binding_missing")
	})

	t.Run("scope denied fails closed", func(t *testing.T) {
		req := newControlPlaneReadRequest(
			http.MethodGet,
			"/v1/user-workspace/usage/quota-summary",
			controlPlaneReadTestAuth("tenant_demo", "tenant:read"),
		)
		req.Header.Set("X-Request-Id", "req-scope-denied")
		rec := httptest.NewRecorder()

		server.httpServer.Handler.ServeHTTP(rec, req)

		envelope := decodeControlPlaneReadEnvelope(t, rec, http.StatusOK)
		assertControlPlaneReadFailure(t, envelope, "scope_denied")
	})

	t.Run("cursor list tenant query mismatch fails closed", func(t *testing.T) {
		req := newControlPlaneReadRequest(
			http.MethodGet,
			"/v1/user-workspace/applications?tenant_ref=tenant_other",
			controlPlaneReadTestAuth("tenant_demo", "applications:read"),
		)
		req.Header.Set("X-Request-Id", "req-cursor-cross-tenant")
		rec := httptest.NewRecorder()

		server.httpServer.Handler.ServeHTTP(rec, req)

		envelope := decodeControlPlaneReadEnvelope(t, rec, http.StatusOK)
		assertControlPlaneReadFailure(t, envelope, "tenant_binding_missing")
	})

	t.Run("cursor list invalid filter fails closed", func(t *testing.T) {
		req := newControlPlaneReadRequest(
			http.MethodGet,
			"/v1/user-workspace/workflow-definitions?executor_state=running",
			controlPlaneReadTestAuth("tenant_demo", "applications:read"),
		)
		req.Header.Set("X-Request-Id", "req-cursor-invalid-filter")
		rec := httptest.NewRecorder()

		server.httpServer.Handler.ServeHTTP(rec, req)

		envelope := decodeControlPlaneReadEnvelope(t, rec, http.StatusOK)
		assertControlPlaneReadFailure(t, envelope, "invalid_filter")
	})

	t.Run("forbidden sensitive projection fails closed", func(t *testing.T) {
		req := newControlPlaneReadRequest(
			http.MethodGet,
			"/v1/user-workspace/api-keys?fields=api_key_value",
			controlPlaneReadTestAuth("tenant_demo", "api_keys:read"),
		)
		req.Header.Set("X-Request-Id", "req-sensitive-projection")
		rec := httptest.NewRecorder()

		server.httpServer.Handler.ServeHTTP(rec, req)

		envelope := decodeControlPlaneReadEnvelope(t, rec, http.StatusOK)
		assertControlPlaneReadFailure(t, envelope, "invalid_filter")
		assertControlPlaneReadNoForbiddenPayload(t, rec.Body.String())
	})

	t.Run("forbidden method returns platform error without side effects", func(t *testing.T) {
		req := newControlPlaneReadRequest(
			http.MethodPost,
			"/v1/user-workspace/runs",
			controlPlaneReadTestAuth("tenant_demo", "runs:read"),
		)
		rec := httptest.NewRecorder()

		server.httpServer.Handler.ServeHTTP(rec, req)

		response := decodePlatformError(t, rec, http.StatusMethodNotAllowed)
		if response.Error.Code != "CONTROL_PLANE_READ_METHOD_NOT_ALLOWED" {
			t.Fatalf("unexpected error code: %#v", response.Error.Code)
		}
		assertControlPlaneReadNoForbiddenPayload(t, rec.Body.String())
	})

	t.Run("forbidden query returns platform error without side effects", func(t *testing.T) {
		req := newControlPlaneReadRequest(
			http.MethodGet,
			"/v1/user-workspace/runs?execute=true",
			controlPlaneReadTestAuth("tenant_demo", "runs:read"),
		)
		rec := httptest.NewRecorder()

		server.httpServer.Handler.ServeHTTP(rec, req)

		response := decodePlatformError(t, rec, http.StatusBadRequest)
		if response.Error.Code != "CONTROL_PLANE_READ_QUERY_FORBIDDEN" {
			t.Fatalf("unexpected error code: %#v", response.Error.Code)
		}
		assertControlPlaneReadNoForbiddenPayload(t, rec.Body.String())
	})

	if sideEffects := server.controlPlaneReadDataStore().SideEffects(); sideEffects != (controlPlaneReadSideEffects{}) {
		t.Fatalf("fake store recorded side effects: %#v", sideEffects)
	}
}

func newControlPlaneReadRequest(method string, target string, auth controlPlaneReadAuthContext) *http.Request {
	req := httptest.NewRequest(method, target, nil)
	return req.WithContext(withControlPlaneReadFakeAuthContext(req.Context(), auth))
}

func controlPlaneReadTestAuth(tenantRef string, scopes ...string) controlPlaneReadAuthContext {
	return controlPlaneReadAuthContext{
		IdentityContext: "subject_demo_user",
		TenantBinding:   tenantRef,
		SubjectBinding:  "subject_demo_user",
		ScopeGrants:     scopes,
		AuditContext:    "audit_test_context",
	}
}

func stringPointerForTest(value string) *string {
	return &value
}

func decodeControlPlaneReadEnvelope(t *testing.T, rec *httptest.ResponseRecorder, expectedStatus int) controlPlaneReadEnvelopeForTest {
	t.Helper()
	if rec.Code != expectedStatus {
		t.Fatalf("unexpected status: %d body=%s", rec.Code, rec.Body.String())
	}
	var envelope controlPlaneReadEnvelopeForTest
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode read envelope: %v", err)
	}
	return envelope
}

func decodePlatformError(t *testing.T, rec *httptest.ResponseRecorder, expectedStatus int) errorDocument {
	t.Helper()
	if rec.Code != expectedStatus {
		t.Fatalf("unexpected status: %d body=%s", rec.Code, rec.Body.String())
	}
	var response errorDocument
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode platform error: %v", err)
	}
	return response
}

func assertControlPlaneReadNextCursor(t *testing.T, actual *string, expected *string) {
	t.Helper()
	if expected == nil {
		if actual != nil {
			t.Fatalf("unexpected next cursor: %#v", *actual)
		}
		return
	}
	if actual == nil || *actual != *expected {
		t.Fatalf("unexpected next cursor: actual=%#v expected=%#v", actual, expected)
	}
}

func assertControlPlaneReadFailure(t *testing.T, envelope controlPlaneReadEnvelopeForTest, expectedFailureCode string) {
	t.Helper()
	if envelope.FailureCode == nil || *envelope.FailureCode != expectedFailureCode {
		t.Fatalf("unexpected failure code: %#v", envelope.FailureCode)
	}
	if len(envelope.Items) != 0 {
		t.Fatalf("failure response returned items: %#v", envelope.Items)
	}
	if strings.TrimSpace(envelope.AuditRef) == "" {
		t.Fatalf("failure response must include audit_ref")
	}
}

func assertControlPlaneReadNoForbiddenPayload(t *testing.T, body string) {
	t.Helper()
	for _, forbidden := range []string{
		"raw_secret_value",
		"api_key_value",
		"api_key_hash",
		"authorization_header",
		"bearer_token",
		"cookie_value",
		"raw_request_body_dump",
		"raw_tool_payload",
		"business_writeback_payload",
		"full_prompt_dump_with_secret",
	} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("control plane read response leaked forbidden payload key %q: %s", forbidden, body)
		}
	}
}
