package httpapi

import (
	"context"
	"net/http"
	"strings"
)

var workflowRAGExecutionRequiredScopes = []string{
	"workflow_rag:execute",
	"workflow_runs:execute",
	"workflow_drafts:read",
	"workflow_rag_snapshots:read",
}

type workflowRAGExecutionHTTPBody struct {
	WorkspaceID   string   `json:"workspace_id"`
	ApplicationID string   `json:"application_id"`
	DraftVersion  int      `json:"draft_version"`
	InputText     string   `json:"input_text"`
	Model         string   `json:"model"`
	Temperature   *float64 `json:"temperature"`
}

func (server *Server) handleWorkflowRAGExecution(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowRAGExecutionRoute)
	if !server.allowWorkflowRAGExecutionDev(writer, trace) {
		return
	}
	var body workflowRAGExecutionHTTPBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{
		maxBytes: maxControlJSONRequestBodyBytes, rejectUnknownFields: true,
	}) {
		return
	}
	runContext, failureCode := workflowRunContextFromRequest(
		request, trace, body.WorkspaceID, body.ApplicationID, "rag-execute", workflowRAGExecutionRequiredScopes...,
	)
	if failureCode != "" || !workflowRAGHasVerifiedActor(request) {
		writeWorkflowRunResult(writer, trace, runContext, workflowRAGExecutionFailure(WorkflowRAGFailureScopeDenied))
		return
	}
	result := server.workflowRAGExecutionService().Execute(runContext, WorkflowRAGExecutionRequest{
		DraftID: strings.TrimSpace(request.PathValue("draft_id")), DraftVersion: body.DraftVersion,
		InputText: body.InputText, Model: body.Model, Temperature: body.Temperature,
	})
	writeWorkflowRunResult(writer, trace, runContext, result)
}

func workflowRAGHasVerifiedActor(request *http.Request) bool {
	auth, ok := request.Context().Value(controlPlaneReadAuthContextKey{}).(controlPlaneReadAuthContext)
	if !ok || auth.VerifiedIdentity == nil || !auth.ResourceBinding.TenantVerified {
		return false
	}
	return strings.TrimSpace(auth.VerifiedIdentity.SubjectRef) == strings.TrimSpace(auth.SubjectBinding) &&
		strings.TrimSpace(auth.VerifiedIdentity.TenantRef) == strings.TrimSpace(auth.TenantBinding) &&
		strings.TrimSpace(auth.ResourceBinding.TenantRef) == strings.TrimSpace(auth.TenantBinding)
}

func (server *Server) allowWorkflowRAGExecutionDev(writer http.ResponseWriter, trace requestTrace) bool {
	if server.config.WorkflowRAGExecutionDevEnabled {
		return true
	}
	server.writePlatformError(writer, trace, "WORKFLOW_RAG_EXECUTION_DEV_DISABLED", "workflow RAG execution route requires explicit development opt-in")
	return false
}

func (server *Server) workflowRAGExecutionService() workflowRAGExecutionService {
	service := newWorkflowRAGExecutionService(
		server.savedWorkflowDraftService().ReadDraft,
		server.workflowRAGSnapshotRepository,
		server.bridge,
		server.workflowRunStore,
	)
	service.defaultTemperature = server.config.Temperature
	service.resolveSelection = func(ctx context.Context, requestedModel string) northboundSelection {
		return server.resolveNorthboundSelection(ctx, requestedModel, nil)
	}
	service.envelopeOptions = server.buildBridgeEnvelopeOptions
	return service
}
