package httpapi

import (
	"fmt"
	"net/http"
	"strconv"
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

type controlPlaneReadEnvelope struct {
	RequestID   string           `json:"request_id"`
	TenantRef   string           `json:"tenant_ref"`
	Items       []map[string]any `json:"items"`
	NextCursor  *string          `json:"next_cursor"`
	FailureCode *string          `json:"failure_code"`
	AuditRef    string           `json:"audit_ref"`
}

type controlPlaneReadCursorListSpec struct {
	RoutePattern          string
	RequiredScope         string
	AuditSuffix           string
	AllowedFilter         []string
	StrictAuditPagination bool
	ReadItems             func(ControlPlaneReadRepository, ReadRepositoryContext, ReadRepositoryRequest) controlPlaneReadRepositoryListResult
}

type controlPlaneReadRepositoryListResult struct {
	TenantRef   string
	Items       []map[string]any
	NextCursor  *string
	FailureCode ReadRepositoryFailureCode
	AuditRef    string
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
	auth, failureCode, failureStatus := authorizeControlPlaneReadRequest(request, tenantRef, "tenant:read")
	if failureCode != "" {
		writeControlPlaneReadFailureWithStatus(writer, trace, tenantRef, failureCode, auditRefForControlPlaneRead(trace, "tenant-summary-denied"), failureStatus)
		return
	}

	auditRef := auditRefForControlPlaneRead(trace, "tenant-summary")
	result := s.controlPlaneReadRepository().ReadTenantSummary(
		controlPlaneReadRepositoryContext(request, trace, auth, auditRef),
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

	auth, failureCode, failureStatus := authorizeControlPlaneReadRequest(request, "", "usage:read")
	if failureCode != "" {
		writeControlPlaneReadFailureWithStatus(writer, trace, controlPlaneReadTenantRefFromRequest(request), failureCode, auditRefForControlPlaneRead(trace, "quota-summary-denied"), failureStatus)
		return
	}

	auditRef := auditRefForControlPlaneRead(trace, "quota-summary")
	result := s.controlPlaneReadRepository().ReadQuotaSummary(
		controlPlaneReadRepositoryContext(request, trace, auth, auditRef),
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
		RoutePattern:          controlPlaneAuditSummaryListRoute,
		RequiredScope:         "audit:read",
		AuditSuffix:           "audit",
		AllowedFilter:         []string{"event_kind", "resource_ref", "actor_subject_ref", "failure_code"},
		StrictAuditPagination: true,
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

	auth, failureCode, failureStatus := authorizeControlPlaneReadRequest(request, "", spec.RequiredScope)
	if failureCode != "" {
		writeControlPlaneReadFailureWithStatus(writer, trace, controlPlaneReadTenantRefFromRequest(request), failureCode, auditRefForControlPlaneRead(trace, spec.AuditSuffix+"-denied"), failureStatus)
		return
	}
	if requestedTenantRef := strings.TrimSpace(request.URL.Query().Get("tenant_ref")); requestedTenantRef != "" && requestedTenantRef != auth.TenantBinding {
		writeControlPlaneReadFailureWithStatus(writer, trace, auth.TenantBinding, "tenant_binding_missing", auditRefForControlPlaneRead(trace, spec.AuditSuffix+"-tenant-mismatch"), http.StatusForbidden)
		return
	}

	filters, filterFailureCode := controlPlaneReadFiltersFromQuery(request, spec.AllowedFilter)
	if filterFailureCode != "" {
		writeControlPlaneReadFailure(writer, trace, auth.TenantBinding, filterFailureCode, auditRefForControlPlaneRead(trace, spec.AuditSuffix+"-invalid-filter"))
		return
	}

	auditRef := auditRefForControlPlaneRead(trace, spec.AuditSuffix)
	repositoryRequest, requestFailureCode := controlPlaneReadRepositoryRequestFromQuery(request, filters, spec.StrictAuditPagination)
	if requestFailureCode != "" {
		writeControlPlaneReadFailure(writer, trace, auth.TenantBinding, requestFailureCode, auditRefForControlPlaneRead(trace, spec.AuditSuffix+"-invalid-pagination"))
		return
	}
	result := spec.ReadItems(
		s.controlPlaneReadRepository(),
		controlPlaneReadRepositoryContext(request, trace, auth, auditRef),
		repositoryRequest,
	)
	if result.FailureCode != "" {
		writeControlPlaneReadFailure(writer, trace, auth.TenantBinding, string(result.FailureCode), auditRef)
		return
	}
	writeControlPlaneReadSuccess(writer, trace, result.TenantRef, result.Items, result.NextCursor, auditRef)
}

func controlPlaneReadRepositoryContext(request *http.Request, trace requestTrace, auth controlPlaneReadAuthContext, auditRef string) ReadRepositoryContext {
	return ReadRepositoryContext{
		RequestContext: request.Context(),
		RequestID:      trace.requestID,
		TenantRef:      auth.TenantBinding,
		SubjectRef:     auth.SubjectBinding,
		ScopeGrants:    append([]string{}, auth.ScopeGrants...),
		AuditRef:       auditRef,
		IssuerRef:      auth.IssuerRef,
		SessionRef:     auth.SessionRef,
	}
}

func controlPlaneReadRepositoryRequestFromQuery(request *http.Request, filters map[string]string, strictAuditPagination bool) (ReadRepositoryRequest, string) {
	query := request.URL.Query()
	result := ReadRepositoryRequest{
		Cursor:  strings.TrimSpace(query.Get("cursor")),
		Filters: ReadRepositoryFilters(filters),
		Sort:    ReadRepositorySort(strings.TrimSpace(query.Get("sort"))),
	}
	if !strictAuditPagination {
		return result, ""
	}
	for _, key := range []string{"limit", "cursor", "sort"} {
		if values, found := query[key]; found && len(values) != 1 {
			return ReadRepositoryRequest{}, "invalid_filter"
		}
	}
	for key, values := range query {
		if key != "tenant_ref" && key != "limit" && key != "cursor" && key != "sort" && len(values) != 1 {
			return ReadRepositoryRequest{}, "invalid_filter"
		}
	}
	result.Limit = 50
	if rawLimit := strings.TrimSpace(query.Get("limit")); rawLimit != "" {
		limit, err := strconv.Atoi(rawLimit)
		if err != nil || limit < 1 || limit > 100 {
			return ReadRepositoryRequest{}, "invalid_filter"
		}
		result.Limit = limit
	} else if _, found := query["limit"]; found {
		return ReadRepositoryRequest{}, "invalid_filter"
	}
	if result.Sort == "" {
		result.Sort = "recorded_at_desc"
	} else if result.Sort != "recorded_at_desc" {
		return ReadRepositoryRequest{}, "invalid_filter"
	}
	if len(result.Cursor) > 1024 {
		return ReadRepositoryRequest{}, "invalid_filter"
	}
	if result.Cursor != "" {
		if _, err := decodeControlPlaneAuditCursor(result.Cursor); err != nil {
			return ReadRepositoryRequest{}, "invalid_filter"
		}
	}
	return result, ""
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

func authorizeControlPlaneReadRequest(request *http.Request, tenantRef string, requiredScope string) (controlPlaneReadAuthContext, string, int) {
	auth, ok := request.Context().Value(controlPlaneReadAuthContextKey{}).(controlPlaneReadAuthContext)
	if ok && strings.TrimSpace(auth.FailureCode) != "" {
		failure := strings.TrimSpace(auth.FailureCode)
		if auth.FailureStatus != 0 {
			return auth, failure, auth.FailureStatus
		}
		return auth, failure, controlPlaneReadFailureHTTPStatus(failure)
	}
	if !ok || strings.TrimSpace(auth.IdentityContext) == "" || strings.TrimSpace(auth.SubjectBinding) == "" {
		return controlPlaneReadAuthContext{}, "identity_context_missing", http.StatusUnauthorized
	}

	auth.TenantBinding = strings.TrimSpace(auth.TenantBinding)
	if auth.TenantBinding == "" {
		return auth, "tenant_binding_missing", http.StatusUnauthorized
	}
	if tenantRef != "" && auth.TenantBinding != tenantRef {
		return auth, "tenant_binding_missing", http.StatusForbidden
	}
	if auth.AuthMode == controlPlaneReadAuthModeRadishOIDCIntegrationTest && requiredScope != "tenant:read" && requiredScope != "audit:read" {
		return auth, "workspace_membership_unavailable", http.StatusServiceUnavailable
	}
	if !controlPlaneReadHasScope(auth.ScopeGrants, requiredScope) {
		return auth, "scope_denied", http.StatusForbidden
	}
	return auth, "", http.StatusOK
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
	writeControlPlaneReadFailureWithStatus(writer, trace, tenantRef, failureCode, auditRef, controlPlaneReadFailureHTTPStatus(failureCode))
}

func writeControlPlaneReadFailureWithStatus(writer http.ResponseWriter, trace requestTrace, tenantRef string, failureCode string, auditRef string, status int) {
	normalizedFailureCode := strings.TrimSpace(failureCode)
	writeObservedJSON(writer, status, trace, controlPlaneReadEnvelope{
		RequestID:   trace.requestID,
		TenantRef:   strings.TrimSpace(tenantRef),
		Items:       []map[string]any{},
		NextCursor:  nil,
		FailureCode: &normalizedFailureCode,
		AuditRef:    auditRef,
	})
}

func controlPlaneReadFailureHTTPStatus(failureCode string) int {
	switch strings.TrimSpace(failureCode) {
	case "identity_context_missing", "auth_context_contract_mismatch":
		return http.StatusUnauthorized
	case "tenant_binding_missing", "scope_denied":
		return http.StatusForbidden
	case "invalid_filter":
		return http.StatusBadRequest
	case "tenant_not_found", "quota_policy_missing":
		return http.StatusNotFound
	case "read_store_unavailable", "database_read_disabled", "schema_migration_not_applied", "schema_version_mismatch", "identity_provider_unavailable", "workspace_membership_unavailable":
		return http.StatusServiceUnavailable
	case "read_store_contract_mismatch":
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

func auditRefForControlPlaneRead(trace requestTrace, suffix string) string {
	return strings.TrimSpace("audit_" + trace.requestID + "_" + suffix)
}
