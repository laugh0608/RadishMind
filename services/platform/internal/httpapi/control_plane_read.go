package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

const (
	controlPlaneTenantSummaryRoute                 = "/v1/control-plane/tenants/{tenant_ref}/summary"
	controlPlaneApplicationSummaryListRoute        = "/v1/user-workspace/applications"
	controlPlaneAPIKeySummaryListRoute             = "/v1/user-workspace/api-keys"
	controlPlaneQuotaSummaryRoute                  = "/v1/user-workspace/usage/quota-summary"
	controlPlaneWorkflowDefinitionSummaryListRoute = "/v1/user-workspace/workflow-definitions"
	controlPlaneRunRecordSummaryListRoute          = "/v1/user-workspace/runs"
	controlPlaneAuditSummaryListRoute              = "/v1/control-plane/audit"
)

const (
	controlPlaneReadDevIdentityHeader = "X-RadishMind-Dev-Read-Identity"
	controlPlaneReadDevTenantHeader   = "X-RadishMind-Dev-Read-Tenant"
	controlPlaneReadDevSubjectHeader  = "X-RadishMind-Dev-Read-Subject"
	controlPlaneReadDevScopesHeader   = "X-RadishMind-Dev-Read-Scopes"
	controlPlaneReadDevAuditHeader    = "X-RadishMind-Dev-Read-Audit"
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

type controlPlaneReadCursorListSpec struct {
	RoutePattern  string
	RequiredScope string
	AuditSuffix   string
	AllowedFilter []string
	ReadItems     func(ControlPlaneReadRepository, ReadRepositoryContext, ReadRepositoryRequest) controlPlaneReadRepositoryListResult
}

type controlPlaneReadRepositoryListResult struct {
	TenantRef   string
	Items       []map[string]any
	NextCursor  *string
	FailureCode ReadRepositoryFailureCode
	AuditRef    string
}

func withControlPlaneReadFakeAuthContext(ctx context.Context, auth controlPlaneReadAuthContext) context.Context {
	return context.WithValue(ctx, controlPlaneReadAuthContextKey{}, auth)
}

func withControlPlaneReadDevAuth(next http.Handler, enabled bool) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if enabled {
			if auth, ok := controlPlaneReadDevAuthFromHeaders(request); ok {
				request = request.WithContext(withControlPlaneReadFakeAuthContext(request.Context(), auth))
			}
		}
		next.ServeHTTP(writer, request)
	})
}

func controlPlaneReadDevAuthFromHeaders(request *http.Request) (controlPlaneReadAuthContext, bool) {
	identity := strings.TrimSpace(request.Header.Get(controlPlaneReadDevIdentityHeader))
	tenantRef := strings.TrimSpace(request.Header.Get(controlPlaneReadDevTenantHeader))
	subjectRef := strings.TrimSpace(request.Header.Get(controlPlaneReadDevSubjectHeader))
	scopeHeader := strings.TrimSpace(request.Header.Get(controlPlaneReadDevScopesHeader))
	if identity == "" || tenantRef == "" || subjectRef == "" || scopeHeader == "" {
		return controlPlaneReadAuthContext{}, false
	}
	return controlPlaneReadAuthContext{
		IdentityContext: identity,
		TenantBinding:   tenantRef,
		SubjectBinding:  subjectRef,
		ScopeGrants:     splitControlPlaneReadDevScopes(scopeHeader),
		AuditContext:    strings.TrimSpace(request.Header.Get(controlPlaneReadDevAuditHeader)),
	}, true
}

func splitControlPlaneReadDevScopes(rawScopes string) []string {
	scopes := make([]string, 0)
	for _, scope := range strings.Split(rawScopes, ",") {
		normalized := strings.TrimSpace(scope)
		if normalized != "" {
			scopes = append(scopes, normalized)
		}
	}
	return scopes
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

	auditRef := auditRefForControlPlaneRead(trace, "tenant-summary")
	result := s.controlPlaneReadRepository().ReadTenantSummary(
		controlPlaneReadRepositoryContext(trace, auth, auditRef),
		ReadTenantSummaryRequest{},
	)
	if result.FailureCode != "" {
		writeControlPlaneReadFailure(writer, trace, auth.TenantBinding, string(result.FailureCode), auditRef)
		return
	}
	if len(result.Items) == 0 {
		writeControlPlaneReadFailure(writer, trace, tenantRef, "tenant_not_found", auditRefForControlPlaneRead(trace, "tenant-summary-not-found"))
		return
	}

	writeControlPlaneReadSuccess(writer, trace, result.TenantRef, []map[string]any{tenantSummaryToControlPlaneReadMap(result.Items[0])}, nil, auditRef)
}

func (s *Server) handleUserWorkspaceApplicationSummaryList(writer http.ResponseWriter, request *http.Request) {
	s.handleControlPlaneReadCursorList(writer, request, controlPlaneReadCursorListSpec{
		RoutePattern:  controlPlaneApplicationSummaryListRoute,
		RequiredScope: "applications:read",
		AuditSuffix:   "applications",
		AllowedFilter: []string{"application_kind", "owner_subject_ref", "last_run_status"},
		ReadItems: func(repository ControlPlaneReadRepository, context ReadRepositoryContext, request ReadRepositoryRequest) controlPlaneReadRepositoryListResult {
			result := repository.ListApplicationSummaries(context, ListApplicationSummariesRequest{ReadRepositoryRequest: request})
			return controlPlaneReadRepositoryListResult{
				TenantRef:   result.TenantRef,
				Items:       applicationSummariesToControlPlaneReadMaps(result.Items),
				NextCursor:  result.NextCursor,
				FailureCode: result.FailureCode,
				AuditRef:    result.AuditRef,
			}
		},
	})
}

func (s *Server) handleUserWorkspaceAPIKeySummaryList(writer http.ResponseWriter, request *http.Request) {
	s.handleControlPlaneReadCursorList(writer, request, controlPlaneReadCursorListSpec{
		RoutePattern:  controlPlaneAPIKeySummaryListRoute,
		RequiredScope: "api_keys:read",
		AuditSuffix:   "api-keys",
		AllowedFilter: []string{"state", "owner_subject_ref", "scope"},
		ReadItems: func(repository ControlPlaneReadRepository, context ReadRepositoryContext, request ReadRepositoryRequest) controlPlaneReadRepositoryListResult {
			result := repository.ListAPIKeySummaries(context, ListAPIKeySummariesRequest{ReadRepositoryRequest: request})
			return controlPlaneReadRepositoryListResult{
				TenantRef:   result.TenantRef,
				Items:       apiKeySummariesToControlPlaneReadMaps(result.Items),
				NextCursor:  result.NextCursor,
				FailureCode: result.FailureCode,
				AuditRef:    result.AuditRef,
			}
		},
	})
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

	auditRef := auditRefForControlPlaneRead(trace, "quota-summary")
	result := s.controlPlaneReadRepository().ReadQuotaSummary(
		controlPlaneReadRepositoryContext(trace, auth, auditRef),
		ReadQuotaSummaryRequest{},
	)
	if result.FailureCode != "" {
		writeControlPlaneReadFailure(writer, trace, auth.TenantBinding, string(result.FailureCode), auditRef)
		return
	}
	if len(result.Items) == 0 {
		writeControlPlaneReadFailure(writer, trace, auth.TenantBinding, "quota_policy_missing", auditRefForControlPlaneRead(trace, "quota-summary-missing"))
		return
	}

	writeControlPlaneReadSuccess(writer, trace, result.TenantRef, []map[string]any{quotaSummaryToControlPlaneReadMap(result.Items[0])}, nil, auditRef)
}

func (s *Server) handleUserWorkspaceWorkflowDefinitionSummaryList(writer http.ResponseWriter, request *http.Request) {
	s.handleControlPlaneReadCursorList(writer, request, controlPlaneReadCursorListSpec{
		RoutePattern:  controlPlaneWorkflowDefinitionSummaryListRoute,
		RequiredScope: "applications:read",
		AuditSuffix:   "workflow-definitions",
		AllowedFilter: []string{"application_ref", "definition_status", "risk_level"},
		ReadItems: func(repository ControlPlaneReadRepository, context ReadRepositoryContext, request ReadRepositoryRequest) controlPlaneReadRepositoryListResult {
			result := repository.ListWorkflowDefinitionSummaries(context, ListWorkflowDefinitionSummariesRequest{ReadRepositoryRequest: request})
			return controlPlaneReadRepositoryListResult{
				TenantRef:   result.TenantRef,
				Items:       workflowDefinitionSummariesToControlPlaneReadMaps(result.Items),
				NextCursor:  result.NextCursor,
				FailureCode: result.FailureCode,
				AuditRef:    result.AuditRef,
			}
		},
	})
}

func (s *Server) handleUserWorkspaceRunRecordSummaryList(writer http.ResponseWriter, request *http.Request) {
	s.handleControlPlaneReadCursorList(writer, request, controlPlaneReadCursorListSpec{
		RoutePattern:  controlPlaneRunRecordSummaryListRoute,
		RequiredScope: "runs:read",
		AuditSuffix:   "runs",
		AllowedFilter: []string{"application_ref", "workflow_definition_id", "status", "failure_code"},
		ReadItems: func(repository ControlPlaneReadRepository, context ReadRepositoryContext, request ReadRepositoryRequest) controlPlaneReadRepositoryListResult {
			result := repository.ListRunRecordSummaries(context, ListRunRecordSummariesRequest{ReadRepositoryRequest: request})
			return controlPlaneReadRepositoryListResult{
				TenantRef:   result.TenantRef,
				Items:       runRecordSummariesToControlPlaneReadMaps(result.Items),
				NextCursor:  result.NextCursor,
				FailureCode: result.FailureCode,
				AuditRef:    result.AuditRef,
			}
		},
	})
}

func (s *Server) handleControlPlaneAuditSummaryList(writer http.ResponseWriter, request *http.Request) {
	s.handleControlPlaneReadCursorList(writer, request, controlPlaneReadCursorListSpec{
		RoutePattern:  controlPlaneAuditSummaryListRoute,
		RequiredScope: "audit:read",
		AuditSuffix:   "audit",
		AllowedFilter: []string{"event_kind", "resource_ref", "actor_subject_ref", "failure_code"},
		ReadItems: func(repository ControlPlaneReadRepository, context ReadRepositoryContext, request ReadRepositoryRequest) controlPlaneReadRepositoryListResult {
			result := repository.ListAuditSummaries(context, ListAuditSummariesRequest{ReadRepositoryRequest: request})
			return controlPlaneReadRepositoryListResult{
				TenantRef:   result.TenantRef,
				Items:       auditSummariesToControlPlaneReadMaps(result.Items),
				NextCursor:  result.NextCursor,
				FailureCode: result.FailureCode,
				AuditRef:    result.AuditRef,
			}
		},
	})
}

func (s *Server) handleControlPlaneReadCursorList(writer http.ResponseWriter, request *http.Request, spec controlPlaneReadCursorListSpec) {
	trace := newRequestTrace(request, spec.RoutePattern)
	if !s.allowControlPlaneReadMethod(writer, request, trace) {
		return
	}
	if !s.allowControlPlaneReadQuery(writer, request, trace) {
		return
	}

	auth, failureCode := authorizeControlPlaneReadRequest(request, "", spec.RequiredScope)
	if failureCode != "" {
		writeControlPlaneReadFailure(writer, trace, controlPlaneReadTenantRefFromRequest(request), failureCode, auditRefForControlPlaneRead(trace, spec.AuditSuffix+"-denied"))
		return
	}
	if requestedTenantRef := strings.TrimSpace(request.URL.Query().Get("tenant_ref")); requestedTenantRef != "" && requestedTenantRef != auth.TenantBinding {
		writeControlPlaneReadFailure(writer, trace, auth.TenantBinding, "tenant_binding_missing", auditRefForControlPlaneRead(trace, spec.AuditSuffix+"-tenant-mismatch"))
		return
	}

	filters, filterFailureCode := controlPlaneReadFiltersFromQuery(request, spec.AllowedFilter)
	if filterFailureCode != "" {
		writeControlPlaneReadFailure(writer, trace, auth.TenantBinding, filterFailureCode, auditRefForControlPlaneRead(trace, spec.AuditSuffix+"-invalid-filter"))
		return
	}

	auditRef := auditRefForControlPlaneRead(trace, spec.AuditSuffix)
	result := spec.ReadItems(
		s.controlPlaneReadRepository(),
		controlPlaneReadRepositoryContext(trace, auth, auditRef),
		controlPlaneReadRepositoryRequestFromQuery(request, filters),
	)
	if result.FailureCode != "" {
		writeControlPlaneReadFailure(writer, trace, auth.TenantBinding, string(result.FailureCode), auditRef)
		return
	}
	writeControlPlaneReadSuccess(writer, trace, result.TenantRef, result.Items, result.NextCursor, auditRef)
}

func controlPlaneReadRepositoryContext(trace requestTrace, auth controlPlaneReadAuthContext, auditRef string) ReadRepositoryContext {
	return ReadRepositoryContext{
		RequestID:   trace.requestID,
		TenantRef:   auth.TenantBinding,
		SubjectRef:  auth.SubjectBinding,
		ScopeGrants: append([]string{}, auth.ScopeGrants...),
		AuditRef:    auditRef,
	}
}

func controlPlaneReadRepositoryRequestFromQuery(request *http.Request, filters map[string]string) ReadRepositoryRequest {
	query := request.URL.Query()
	return ReadRepositoryRequest{
		Cursor:  strings.TrimSpace(query.Get("cursor")),
		Filters: ReadRepositoryFilters(filters),
		Sort:    ReadRepositorySort(strings.TrimSpace(query.Get("sort"))),
	}
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

func controlPlaneReadFiltersFromQuery(request *http.Request, allowedFilters []string) (map[string]string, string) {
	allowed := map[string]bool{
		"limit":      true,
		"cursor":     true,
		"sort":       true,
		"tenant_ref": true,
	}
	for _, filter := range allowedFilters {
		allowed[strings.TrimSpace(filter)] = true
	}

	filters := map[string]string{}
	for key, values := range request.URL.Query() {
		normalizedKey := strings.TrimSpace(key)
		if !allowed[normalizedKey] {
			return nil, "invalid_filter"
		}
		if normalizedKey == "limit" || normalizedKey == "cursor" || normalizedKey == "sort" || normalizedKey == "tenant_ref" {
			continue
		}
		if len(values) == 0 {
			continue
		}
		filters[normalizedKey] = strings.TrimSpace(values[0])
	}
	return filters, ""
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
