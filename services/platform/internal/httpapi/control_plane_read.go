package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

const (
	controlPlaneTenantSummaryRoute = "/v1/control-plane/tenants/{tenant_ref}/summary"
	controlPlaneQuotaSummaryRoute  = "/v1/user-workspace/usage/quota-summary"
)

type controlPlaneReadAuthContext struct {
	IdentityContext string
	TenantBinding   string
	SubjectBinding  string
	ScopeGrants     []string
	AuditContext    string
}

type controlPlaneReadAuthContextKey struct{}

type controlPlaneReadEnvelope struct {
	RequestID   string           `json:"request_id"`
	TenantRef   string           `json:"tenant_ref"`
	Items       []map[string]any `json:"items"`
	NextCursor  *string          `json:"next_cursor"`
	FailureCode *string          `json:"failure_code"`
	AuditRef    string           `json:"audit_ref"`
}

func withControlPlaneReadFakeAuthContext(ctx context.Context, auth controlPlaneReadAuthContext) context.Context {
	return context.WithValue(ctx, controlPlaneReadAuthContextKey{}, auth)
}

func (s *Server) handleControlPlaneTenantSummary(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, controlPlaneTenantSummaryRoute)
	if !s.allowControlPlaneReadMethod(writer, request, trace) {
		return
	}
	if !s.allowControlPlaneReadQuery(writer, request, trace) {
		return
	}

	tenantRef := strings.TrimSpace(request.PathValue("tenant_ref"))
	auth, failureCode := authorizeControlPlaneReadRequest(request, tenantRef, "tenant:read")
	if failureCode != "" {
		writeControlPlaneReadFailure(writer, trace, tenantRef, failureCode, auditRefForControlPlaneRead(trace, "tenant-summary-denied"))
		return
	}

	item, found := s.controlPlaneReadDataStore().TenantSummary(tenantRef)
	if !found {
		writeControlPlaneReadFailure(writer, trace, tenantRef, "tenant_not_found", auditRefForControlPlaneRead(trace, "tenant-summary-not-found"))
		return
	}

	writeControlPlaneReadSuccess(writer, trace, auth.TenantBinding, []map[string]any{item}, nil, auditRefForControlPlaneRead(trace, "tenant-summary"))
}

func (s *Server) handleUserWorkspaceQuotaSummary(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, controlPlaneQuotaSummaryRoute)
	if !s.allowControlPlaneReadMethod(writer, request, trace) {
		return
	}
	if !s.allowControlPlaneReadQuery(writer, request, trace) {
		return
	}

	auth, failureCode := authorizeControlPlaneReadRequest(request, "", "usage:read")
	if failureCode != "" {
		writeControlPlaneReadFailure(writer, trace, controlPlaneReadTenantRefFromRequest(request), failureCode, auditRefForControlPlaneRead(trace, "quota-summary-denied"))
		return
	}

	item, found := s.controlPlaneReadDataStore().QuotaSummary(auth.TenantBinding)
	if !found {
		writeControlPlaneReadFailure(writer, trace, auth.TenantBinding, "quota_policy_missing", auditRefForControlPlaneRead(trace, "quota-summary-missing"))
		return
	}

	writeControlPlaneReadSuccess(writer, trace, auth.TenantBinding, []map[string]any{item}, nil, auditRefForControlPlaneRead(trace, "quota-summary"))
}

func (s *Server) allowControlPlaneReadMethod(writer http.ResponseWriter, request *http.Request, trace requestTrace) bool {
	if request.Method == http.MethodGet {
		return true
	}
	s.writePlatformError(
		writer,
		trace,
		"CONTROL_PLANE_READ_METHOD_NOT_ALLOWED",
		fmt.Sprintf("method %s is not allowed for read-only control plane route", request.Method),
	)
	return false
}

func (s *Server) allowControlPlaneReadQuery(writer http.ResponseWriter, request *http.Request, trace requestTrace) bool {
	if parameter, found := forbiddenControlPlaneReadQueryParameter(request); found {
		s.writePlatformError(
			writer,
			trace,
			"CONTROL_PLANE_READ_QUERY_FORBIDDEN",
			fmt.Sprintf("query parameter %q is forbidden for read-only control plane route", parameter),
		)
		return false
	}
	return true
}

func authorizeControlPlaneReadRequest(request *http.Request, tenantRef string, requiredScope string) (controlPlaneReadAuthContext, string) {
	auth, ok := request.Context().Value(controlPlaneReadAuthContextKey{}).(controlPlaneReadAuthContext)
	if !ok || strings.TrimSpace(auth.IdentityContext) == "" || strings.TrimSpace(auth.SubjectBinding) == "" {
		return controlPlaneReadAuthContext{}, "identity_context_missing"
	}

	auth.TenantBinding = strings.TrimSpace(auth.TenantBinding)
	if auth.TenantBinding == "" {
		return auth, "tenant_binding_missing"
	}
	if tenantRef != "" && auth.TenantBinding != tenantRef {
		return auth, "tenant_binding_missing"
	}
	if !controlPlaneReadHasScope(auth.ScopeGrants, requiredScope) {
		return auth, "scope_denied"
	}
	return auth, ""
}

func controlPlaneReadHasScope(scopes []string, requiredScope string) bool {
	requiredScope = strings.TrimSpace(requiredScope)
	for _, scope := range scopes {
		if strings.TrimSpace(scope) == requiredScope {
			return true
		}
	}
	return false
}

func forbiddenControlPlaneReadQueryParameter(request *http.Request) (string, bool) {
	query := request.URL.Query()
	for _, parameter := range []string{
		"execute",
		"replay",
		"confirmation_decision_ref",
		"writeback_payload",
		"raw_tool_payload",
		"include_secret",
	} {
		if _, found := query[parameter]; found {
			return parameter, true
		}
	}
	return "", false
}

func controlPlaneReadTenantRefFromRequest(request *http.Request) string {
	if auth, ok := request.Context().Value(controlPlaneReadAuthContextKey{}).(controlPlaneReadAuthContext); ok {
		return strings.TrimSpace(auth.TenantBinding)
	}
	return ""
}

func writeControlPlaneReadSuccess(writer http.ResponseWriter, trace requestTrace, tenantRef string, items []map[string]any, nextCursor *string, auditRef string) {
	writeObservedJSON(writer, http.StatusOK, trace, controlPlaneReadEnvelope{
		RequestID:   trace.requestID,
		TenantRef:   strings.TrimSpace(tenantRef),
		Items:       items,
		NextCursor:  nextCursor,
		FailureCode: nil,
		AuditRef:    auditRef,
	})
}

func writeControlPlaneReadFailure(writer http.ResponseWriter, trace requestTrace, tenantRef string, failureCode string, auditRef string) {
	normalizedFailureCode := strings.TrimSpace(failureCode)
	writeObservedJSON(writer, http.StatusOK, trace, controlPlaneReadEnvelope{
		RequestID:   trace.requestID,
		TenantRef:   strings.TrimSpace(tenantRef),
		Items:       []map[string]any{},
		NextCursor:  nil,
		FailureCode: &normalizedFailureCode,
		AuditRef:    auditRef,
	})
}

func auditRefForControlPlaneRead(trace requestTrace, suffix string) string {
	return strings.TrimSpace("audit_" + trace.requestID + "_" + suffix)
}
