package httpapi

import (
	"net/http"
	"strconv"
	"strings"
)

const (
	workflowRAGSnapshotCreateRoute  = "POST /v1/user-workspace/workflow-retrieval-snapshots"
	workflowRAGSnapshotListRoute    = "GET /v1/user-workspace/workflow-retrieval-snapshots"
	workflowRAGSnapshotReadRoute    = "GET /v1/user-workspace/workflow-retrieval-snapshots/{snapshot_id}"
	workflowRAGSnapshotVersionRoute = "POST /v1/user-workspace/workflow-retrieval-snapshots/{snapshot_id}/versions"
	workflowRAGSnapshotArchiveRoute = "POST /v1/user-workspace/workflow-retrieval-snapshots/{snapshot_id}/archive"
	workflowRAGSnapshotMaxBodyBytes = int64(2 << 20)
)

type workflowRAGSnapshotCreateBody struct {
	WorkspaceID           string                     `json:"workspace_id"`
	ApplicationID         string                     `json:"application_id"`
	SnapshotKey           string                     `json:"snapshot_key"`
	DisplayName           string                     `json:"display_name"`
	ContentClassification string                     `json:"content_classification"`
	Fragments             []WorkflowRAGFragmentInput `json:"fragments"`
}

type workflowRAGSnapshotVersionBody struct {
	WorkspaceID           string                     `json:"workspace_id"`
	ApplicationID         string                     `json:"application_id"`
	ExpectedLatestVersion int                        `json:"expected_latest_version"`
	DisplayName           string                     `json:"display_name"`
	ContentClassification string                     `json:"content_classification"`
	Fragments             []WorkflowRAGFragmentInput `json:"fragments"`
}

type workflowRAGSnapshotArchiveBody struct {
	WorkspaceID           string `json:"workspace_id"`
	ApplicationID         string `json:"application_id"`
	ExpectedLatestVersion int    `json:"expected_latest_version"`
}

type workflowRAGSnapshotEnvelope struct {
	RequestID            string                     `json:"request_id"`
	TenantRef            string                     `json:"tenant_ref"`
	WorkspaceID          string                     `json:"workspace_id"`
	ApplicationID        string                     `json:"application_id"`
	Record               *WorkflowRAGSnapshotRecord `json:"record"`
	FailureCode          *string                    `json:"failure_code"`
	CurrentLatestVersion int                        `json:"current_latest_version"`
	CurrentLifecycle     string                     `json:"current_lifecycle_state"`
	AuditRef             string                     `json:"audit_ref"`
}

type workflowRAGSnapshotListEnvelope struct {
	RequestID     string                        `json:"request_id"`
	TenantRef     string                        `json:"tenant_ref"`
	WorkspaceID   string                        `json:"workspace_id"`
	ApplicationID string                        `json:"application_id"`
	Items         []WorkflowRAGSnapshotResource `json:"items"`
	NextCursor    *string                       `json:"next_cursor"`
	FailureCode   *string                       `json:"failure_code"`
	AuditRef      string                        `json:"audit_ref"`
}

func (server *Server) handleCreateWorkflowRAGSnapshot(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowRAGSnapshotCreateRoute)
	if !server.allowWorkflowRAGSnapshotDev(writer, trace) {
		return
	}
	var body workflowRAGSnapshotCreateBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: workflowRAGSnapshotMaxBodyBytes, rejectUnknownFields: true}) {
		return
	}
	ctx, failure := workflowRAGSnapshotContextFromRequest(request, trace, body.WorkspaceID, body.ApplicationID, "create", "workflow_rag_snapshots:write")
	if failure != "" {
		writeWorkflowRAGSnapshotResult(writer, trace, ctx, WorkflowRAGSnapshotResult{FailureCode: failure})
		return
	}
	result := server.workflowRAGSnapshotService().Create(ctx, WorkflowRAGSnapshotCreateInput{
		SnapshotKey: body.SnapshotKey, DisplayName: body.DisplayName,
		ContentClassification: body.ContentClassification, Fragments: body.Fragments,
	})
	writeWorkflowRAGSnapshotResult(writer, trace, ctx, result)
}

func (server *Server) handleListWorkflowRAGSnapshots(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowRAGSnapshotListRoute)
	if !server.allowWorkflowRAGSnapshotDev(writer, trace) {
		return
	}
	for key := range request.URL.Query() {
		switch key {
		case "workspace_id", "application_id", "lifecycle_state", "limit", "cursor":
		default:
			server.writePlatformError(writer, trace, WorkflowRAGFailurePayloadInvalid, "workflow RAG snapshot list query is invalid")
			return
		}
	}
	ctx, failure := workflowRAGSnapshotContextFromRequest(request, trace, request.URL.Query().Get("workspace_id"), request.URL.Query().Get("application_id"), "list", "workflow_rag_snapshots:read")
	if failure != "" {
		writeWorkflowRAGSnapshotListResult(writer, trace, ctx, WorkflowRAGSnapshotListResult{Records: []WorkflowRAGSnapshotResource{}, FailureCode: failure})
		return
	}
	limit := 0
	if value := strings.TrimSpace(request.URL.Query().Get("limit")); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			writeWorkflowRAGSnapshotListResult(writer, trace, ctx, WorkflowRAGSnapshotListResult{Records: []WorkflowRAGSnapshotResource{}, FailureCode: WorkflowRAGFailurePayloadInvalid})
			return
		}
		limit = parsed
	}
	result := server.workflowRAGSnapshotService().List(ctx, WorkflowRAGSnapshotListInput{
		LifecycleState: request.URL.Query().Get("lifecycle_state"), Limit: limit, Cursor: request.URL.Query().Get("cursor"),
	})
	writeWorkflowRAGSnapshotListResult(writer, trace, ctx, result)
}

func (server *Server) handleReadWorkflowRAGSnapshot(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowRAGSnapshotReadRoute)
	if !server.allowWorkflowRAGSnapshotDev(writer, trace) {
		return
	}
	for key := range request.URL.Query() {
		switch key {
		case "workspace_id", "application_id", "snapshot_version":
		default:
			server.writePlatformError(writer, trace, WorkflowRAGFailurePayloadInvalid, "workflow RAG snapshot detail query is invalid")
			return
		}
	}
	ctx, failure := workflowRAGSnapshotContextFromRequest(request, trace, request.URL.Query().Get("workspace_id"), request.URL.Query().Get("application_id"), "read", "workflow_rag_snapshots:read")
	if failure != "" {
		writeWorkflowRAGSnapshotResult(writer, trace, ctx, WorkflowRAGSnapshotResult{FailureCode: failure})
		return
	}
	version, err := strconv.Atoi(strings.TrimSpace(request.URL.Query().Get("snapshot_version")))
	if err != nil {
		writeWorkflowRAGSnapshotResult(writer, trace, ctx, WorkflowRAGSnapshotResult{FailureCode: WorkflowRAGFailurePayloadInvalid})
		return
	}
	writeWorkflowRAGSnapshotResult(writer, trace, ctx, server.workflowRAGSnapshotService().Read(ctx, request.PathValue("snapshot_id"), version))
}

func (server *Server) handleVersionWorkflowRAGSnapshot(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowRAGSnapshotVersionRoute)
	if !server.allowWorkflowRAGSnapshotDev(writer, trace) {
		return
	}
	var body workflowRAGSnapshotVersionBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: workflowRAGSnapshotMaxBodyBytes, rejectUnknownFields: true}) {
		return
	}
	ctx, failure := workflowRAGSnapshotContextFromRequest(request, trace, body.WorkspaceID, body.ApplicationID, "version", "workflow_rag_snapshots:write")
	if failure != "" {
		writeWorkflowRAGSnapshotResult(writer, trace, ctx, WorkflowRAGSnapshotResult{FailureCode: failure})
		return
	}
	result := server.workflowRAGSnapshotService().Version(ctx, request.PathValue("snapshot_id"), WorkflowRAGSnapshotVersionInput{
		ExpectedLatestVersion: body.ExpectedLatestVersion, DisplayName: body.DisplayName,
		ContentClassification: body.ContentClassification, Fragments: body.Fragments,
	})
	writeWorkflowRAGSnapshotResult(writer, trace, ctx, result)
}

func (server *Server) handleArchiveWorkflowRAGSnapshot(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowRAGSnapshotArchiveRoute)
	if !server.allowWorkflowRAGSnapshotDev(writer, trace) {
		return
	}
	var body workflowRAGSnapshotArchiveBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: maxControlJSONRequestBodyBytes, rejectUnknownFields: true}) {
		return
	}
	ctx, failure := workflowRAGSnapshotContextFromRequest(request, trace, body.WorkspaceID, body.ApplicationID, "archive", "workflow_rag_snapshots:archive")
	if failure != "" {
		writeWorkflowRAGSnapshotResult(writer, trace, ctx, WorkflowRAGSnapshotResult{FailureCode: failure})
		return
	}
	writeWorkflowRAGSnapshotResult(writer, trace, ctx, server.workflowRAGSnapshotService().Archive(ctx, request.PathValue("snapshot_id"), body.ExpectedLatestVersion))
}

func workflowRAGSnapshotContextFromRequest(request *http.Request, trace requestTrace, workspaceID, applicationID, auditSuffix, scope string) (WorkflowRAGSnapshotContext, string) {
	runContext, failure := workflowRunContextFromRequest(request, trace, workspaceID, applicationID, "rag-snapshot-"+auditSuffix, scope)
	ctx := WorkflowRAGSnapshotContext{
		RequestContext: request.Context(), RequestID: trace.requestID, TenantRef: runContext.TenantRef,
		WorkspaceID: runContext.WorkspaceID, ApplicationID: runContext.ApplicationID,
		ActorRef: runContext.ActorRef, AuditRef: runContext.AuditRef,
	}
	if failure != "" {
		return ctx, WorkflowRAGFailureScopeDenied
	}
	return ctx, ""
}

func (server *Server) allowWorkflowRAGSnapshotDev(writer http.ResponseWriter, trace requestTrace) bool {
	if server.config.WorkflowRAGSnapshotDevEnabled {
		return true
	}
	server.writePlatformError(writer, trace, "WORKFLOW_RAG_SNAPSHOT_DEV_DISABLED", "workflow RAG snapshot route requires explicit development opt-in")
	return false
}

func (server *Server) workflowRAGSnapshotService() workflowRAGSnapshotService {
	if server.workflowRAGSnapshotRepository == nil {
		unavailable := newMemoryWorkflowRAGSnapshotRepository(nil)
		unavailable.available = false
		server.workflowRAGSnapshotRepository = unavailable
	}
	return newWorkflowRAGSnapshotService(server.workflowRAGSnapshotRepository)
}

func writeWorkflowRAGSnapshotResult(writer http.ResponseWriter, trace requestTrace, ctx WorkflowRAGSnapshotContext, result WorkflowRAGSnapshotResult) {
	writeObservedJSON(writer, http.StatusOK, trace, workflowRAGSnapshotEnvelope{
		RequestID: trace.requestID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID,
		Record: result.Record, FailureCode: optionalApplicationDraftFailure(result.FailureCode),
		CurrentLatestVersion: result.CurrentLatestVersion, CurrentLifecycle: result.CurrentLifecycle, AuditRef: ctx.AuditRef,
	})
}

func writeWorkflowRAGSnapshotListResult(writer http.ResponseWriter, trace requestTrace, ctx WorkflowRAGSnapshotContext, result WorkflowRAGSnapshotListResult) {
	writeObservedJSON(writer, http.StatusOK, trace, workflowRAGSnapshotListEnvelope{
		RequestID: trace.requestID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID,
		Items: result.Records, NextCursor: result.NextCursor, FailureCode: optionalApplicationDraftFailure(result.FailureCode), AuditRef: ctx.AuditRef,
	})
}
