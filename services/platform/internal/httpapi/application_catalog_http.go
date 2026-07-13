package httpapi

import (
	"net/http"
	"strconv"
	"strings"
)

const (
	applicationCatalogCreateRoute  = "POST /v1/user-workspace/applications"
	applicationCatalogReadRoute    = "GET /v1/user-workspace/applications/{application_id}"
	applicationCatalogUpdateRoute  = "PUT /v1/user-workspace/applications/{application_id}"
	applicationCatalogArchiveRoute = "POST /v1/user-workspace/applications/{application_id}/archive"
)

type applicationCatalogCreateBody struct {
	WorkspaceID     string `json:"workspace_id"`
	DisplayName     string `json:"display_name"`
	Description     string `json:"description"`
	ApplicationKind string `json:"application_kind"`
}

type applicationCatalogUpdateBody struct {
	WorkspaceID     string `json:"workspace_id"`
	ExpectedVersion int    `json:"expected_version"`
	DisplayName     string `json:"display_name"`
	Description     string `json:"description"`
	ApplicationKind string `json:"application_kind"`
}

type applicationCatalogArchiveBody struct {
	WorkspaceID     string `json:"workspace_id"`
	ExpectedVersion int    `json:"expected_version"`
}

type applicationCatalogEnvelope struct {
	RequestID             string                    `json:"request_id"`
	TenantRef             string                    `json:"tenant_ref"`
	WorkspaceID           string                    `json:"workspace_id"`
	Record                *ApplicationCatalogRecord `json:"record"`
	FailureCode           *string                   `json:"failure_code"`
	CurrentRecordVersion  int                       `json:"current_record_version"`
	CurrentLifecycleState string                    `json:"current_lifecycle_state"`
	AuditRef              string                    `json:"audit_ref"`
}

func (server *Server) handleCreateApplicationCatalogRecord(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, applicationCatalogCreateRoute)
	if !server.allowApplicationCatalogDevHTTP(writer, trace) {
		return
	}
	var body applicationCatalogCreateBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: maxControlJSONRequestBodyBytes, rejectUnknownFields: true}) {
		return
	}
	requestContext, failureCode, status := server.applicationCatalogContextFromRequest(request, trace, body.WorkspaceID, "applications:write", "create")
	if failureCode != "" {
		writeApplicationCatalogResult(writer, status, trace, requestContext, ApplicationCatalogResult{FailureCode: failureCode})
		return
	}
	requestContext.WriteEnabled = server.config.ApplicationCatalogDevWriteEnabled
	result := server.applicationCatalogService().Create(requestContext, ApplicationCatalogCreateInput{
		DisplayName: body.DisplayName, Description: body.Description, ApplicationKind: body.ApplicationKind,
	})
	writeApplicationCatalogResult(writer, http.StatusOK, trace, requestContext, result)
}

func (server *Server) handleReadApplicationCatalogRecord(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, applicationCatalogReadRoute)
	if !server.allowApplicationCatalogDevHTTP(writer, trace) {
		return
	}
	requestContext, failureCode, status := server.applicationCatalogContextFromRequest(request, trace, request.URL.Query().Get("workspace_id"), "applications:read", "read")
	if failureCode != "" {
		writeApplicationCatalogResult(writer, status, trace, requestContext, ApplicationCatalogResult{FailureCode: failureCode})
		return
	}
	writeApplicationCatalogResult(writer, http.StatusOK, trace, requestContext, server.applicationCatalogService().Read(requestContext, request.PathValue("application_id")))
}

func (server *Server) handleUpdateApplicationCatalogRecord(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, applicationCatalogUpdateRoute)
	if !server.allowApplicationCatalogDevHTTP(writer, trace) {
		return
	}
	var body applicationCatalogUpdateBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: maxControlJSONRequestBodyBytes, rejectUnknownFields: true}) {
		return
	}
	requestContext, failureCode, status := server.applicationCatalogContextFromRequest(request, trace, body.WorkspaceID, "applications:write", "update")
	if failureCode != "" {
		writeApplicationCatalogResult(writer, status, trace, requestContext, ApplicationCatalogResult{FailureCode: failureCode})
		return
	}
	requestContext.WriteEnabled = server.config.ApplicationCatalogDevWriteEnabled
	result := server.applicationCatalogService().Update(requestContext, request.PathValue("application_id"), ApplicationCatalogUpdateInput{
		ExpectedVersion: body.ExpectedVersion, DisplayName: body.DisplayName, Description: body.Description, ApplicationKind: body.ApplicationKind,
	})
	writeApplicationCatalogResult(writer, http.StatusOK, trace, requestContext, result)
}

func (server *Server) handleArchiveApplicationCatalogRecord(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, applicationCatalogArchiveRoute)
	if !server.allowApplicationCatalogDevHTTP(writer, trace) {
		return
	}
	var body applicationCatalogArchiveBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: maxControlJSONRequestBodyBytes, rejectUnknownFields: true}) {
		return
	}
	requestContext, failureCode, status := server.applicationCatalogContextFromRequest(request, trace, body.WorkspaceID, "applications:archive", "archive")
	if failureCode != "" {
		writeApplicationCatalogResult(writer, status, trace, requestContext, ApplicationCatalogResult{FailureCode: failureCode})
		return
	}
	requestContext.WriteEnabled = server.config.ApplicationCatalogDevWriteEnabled
	writeApplicationCatalogResult(writer, http.StatusOK, trace, requestContext, server.applicationCatalogService().Archive(requestContext, request.PathValue("application_id"), body.ExpectedVersion))
}

func (server *Server) handleListApplicationCatalogRecords(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, controlPlaneApplicationSummaryListRoute)
	requestContext, failureCode, status := server.applicationCatalogContextFromRequest(request, trace, request.URL.Query().Get("workspace_id"), "applications:read", "list")
	if failureCode != "" {
		writeApplicationCatalogListResult(writer, status, trace, requestContext, ApplicationCatalogListResult{Records: []ApplicationCatalogRecord{}, FailureCode: failureCode})
		return
	}
	for key := range request.URL.Query() {
		switch key {
		case "workspace_id", "lifecycle_state", "application_kind", "limit", "cursor":
		default:
			writeApplicationCatalogListResult(writer, http.StatusBadRequest, trace, requestContext, ApplicationCatalogListResult{Records: []ApplicationCatalogRecord{}, FailureCode: ApplicationCatalogFailurePayloadInvalid})
			return
		}
	}
	limit := 0
	if value := strings.TrimSpace(request.URL.Query().Get("limit")); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			writeApplicationCatalogListResult(writer, http.StatusBadRequest, trace, requestContext, ApplicationCatalogListResult{Records: []ApplicationCatalogRecord{}, FailureCode: ApplicationCatalogFailurePayloadInvalid})
			return
		}
		limit = parsed
	}
	result := server.applicationCatalogService().List(requestContext, ApplicationCatalogListInput{
		LifecycleState: request.URL.Query().Get("lifecycle_state"), ApplicationKind: request.URL.Query().Get("application_kind"),
		Limit: limit, Cursor: request.URL.Query().Get("cursor"),
	})
	writeApplicationCatalogListResult(writer, http.StatusOK, trace, requestContext, result)
}

func (server *Server) applicationCatalogContextFromRequest(request *http.Request, trace requestTrace, workspaceID, requiredScope, auditSuffix string) (ApplicationCatalogContext, string, int) {
	auth, failureCode, status := authorizeControlPlaneReadRequest(request, "", requiredScope)
	requestContext := ApplicationCatalogContext{
		RequestContext: request.Context(), RequestID: trace.requestID, TenantRef: strings.TrimSpace(auth.TenantBinding),
		WorkspaceID: strings.TrimSpace(workspaceID), ActorRef: strings.TrimSpace(auth.SubjectBinding),
		OwnerSubjectRef: strings.TrimSpace(auth.SubjectBinding), AuditRef: "audit_" + trace.requestID + "_application-catalog-" + auditSuffix,
	}
	if failureCode != "" {
		if failureCode == "workspace_membership_unavailable" {
			return requestContext, failureCode, status
		}
		return requestContext, ApplicationCatalogFailureScopeDenied, status
	}
	if requestContext.WorkspaceID == "" || !applicationDraftIdentifierPattern.MatchString(requestContext.WorkspaceID) {
		return requestContext, ApplicationCatalogFailureScopeDenied, http.StatusForbidden
	}
	return requestContext, "", http.StatusOK
}

func (server *Server) allowApplicationCatalogDevHTTP(writer http.ResponseWriter, trace requestTrace) bool {
	if server.config.ApplicationCatalogDevHTTPEnabled {
		return true
	}
	server.writePlatformError(writer, trace, "APPLICATION_CATALOG_DEV_HTTP_DISABLED", "application catalog dev route requires explicit opt-in")
	return false
}

func (server *Server) applicationCatalogService() applicationCatalogService {
	if server.applicationCatalogRepository == nil {
		server.applicationCatalogRepository = &memoryApplicationCatalogRepository{records: make(map[string]ApplicationCatalogRecord), unavailable: true}
	}
	return newApplicationCatalogService(server.applicationCatalogRepository)
}

func writeApplicationCatalogResult(writer http.ResponseWriter, status int, trace requestTrace, requestContext ApplicationCatalogContext, result ApplicationCatalogResult) {
	writeObservedJSON(writer, status, trace, applicationCatalogEnvelope{
		RequestID: trace.requestID, TenantRef: requestContext.TenantRef, WorkspaceID: requestContext.WorkspaceID,
		Record: result.Record, FailureCode: optionalApplicationDraftFailure(result.FailureCode),
		CurrentRecordVersion: result.CurrentRecordVersion, CurrentLifecycleState: result.CurrentLifecycleState,
		AuditRef: requestContext.AuditRef,
	})
}

func writeApplicationCatalogListResult(writer http.ResponseWriter, status int, trace requestTrace, requestContext ApplicationCatalogContext, result ApplicationCatalogListResult) {
	items := make([]map[string]any, 0, len(result.Records))
	for _, record := range result.Records {
		items = append(items, applicationCatalogRecordToSummaryMap(record))
	}
	writeObservedJSON(writer, status, trace, controlPlaneReadEnvelope{
		RequestID: trace.requestID, TenantRef: requestContext.TenantRef, Items: items,
		NextCursor: result.NextCursor, FailureCode: optionalApplicationDraftFailure(result.FailureCode), AuditRef: requestContext.AuditRef,
	})
}

func applicationCatalogRecordToSummaryMap(record ApplicationCatalogRecord) map[string]any {
	return map[string]any{
		"application_ref": record.ApplicationID, "tenant_ref": record.TenantRef, "workspace_id": record.WorkspaceID,
		"application_kind": record.ApplicationKind, "display_name": record.DisplayName, "description": record.Description,
		"owner_subject_ref": record.OwnerSubjectRef, "latest_workflow_definition_ref": "", "last_run_status": "not_available",
		"lifecycle_state": record.LifecycleState, "record_version": record.RecordVersion, "created_at": record.CreatedAt,
		"updated_at": record.UpdatedAt, "archived_at": record.ArchivedAt,
	}
}
