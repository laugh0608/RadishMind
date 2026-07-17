package httpapi

import (
	"net/http"
	"strings"
)

const (
	workflowHTTPToolPlanCreateRoute = "POST /v1/user-workspace/workflow-drafts/{draft_id}/tool-action-plans"
	workflowHTTPToolPlanReadRoute   = "GET /v1/user-workspace/workflow-tool-action-plans/{plan_id}"
	workflowHTTPToolDecisionRoute   = "POST /v1/user-workspace/workflow-tool-action-plans/{plan_id}/decisions"
	workflowHTTPToolExecutionRoute  = "POST /v1/user-workspace/workflow-tool-action-plans/{plan_id}/executions"
)

type workflowHTTPToolCreatePlanBody struct {
	WorkspaceID     string         `json:"workspace_id"`
	ApplicationID   string         `json:"application_id"`
	DraftVersion    int            `json:"draft_version"`
	NodeID          string         `json:"node_id"`
	PublicArguments map[string]any `json:"public_arguments"`
}

type workflowHTTPToolDecisionBody struct {
	WorkspaceID           string                              `json:"workspace_id"`
	ApplicationID         string                              `json:"application_id"`
	ExpectedRecordVersion int                                 `json:"expected_record_version"`
	Decision              WorkflowHTTPToolConfirmationOutcome `json:"decision"`
}

type workflowHTTPToolExecutionBody struct {
	WorkspaceID           string   `json:"workspace_id"`
	ApplicationID         string   `json:"application_id"`
	ExpectedRecordVersion int      `json:"expected_record_version"`
	InputText             string   `json:"input_text"`
	Model                 string   `json:"model"`
	Temperature           *float64 `json:"temperature"`
}

type workflowHTTPToolActionEnvelope struct {
	RequestID            string                                `json:"request_id"`
	WorkspaceID          string                                `json:"workspace_id"`
	ApplicationID        string                                `json:"application_id"`
	ActionPlan           *WorkflowHTTPToolActionPlan           `json:"action_plan"`
	ConfirmationDecision *WorkflowHTTPToolConfirmationDecision `json:"confirmation_decision"`
	FailureCode          *string                               `json:"failure_code"`
	FailureSummary       string                                `json:"failure_summary"`
	AuditRef             string                                `json:"audit_ref"`
}

type workflowHTTPToolExecutionEnvelope struct {
	RequestID      string                      `json:"request_id"`
	WorkspaceID    string                      `json:"workspace_id"`
	ApplicationID  string                      `json:"application_id"`
	ActionPlan     *WorkflowHTTPToolActionPlan `json:"action_plan"`
	Run            *WorkflowRunRecord          `json:"run"`
	FailureCode    *string                     `json:"failure_code"`
	FailureSummary string                      `json:"failure_summary"`
	AuditRef       string                      `json:"audit_ref"`
}

func (s *Server) handleCreateWorkflowHTTPToolActionPlan(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowHTTPToolPlanCreateRoute)
	if !s.allowWorkflowHTTPToolActionDev(writer, trace) {
		return
	}
	var body workflowHTTPToolCreatePlanBody
	if !s.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: maxControlJSONRequestBodyBytes, rejectUnknownFields: true}) {
		return
	}
	ctx, failure := workflowHTTPToolActionContextFromRequest(request, trace, body.WorkspaceID, body.ApplicationID, "plan", "workflow_drafts:read", "workflow_tool_actions:plan")
	if failure != "" {
		writeWorkflowHTTPToolActionResult(writer, trace, ctx, workflowHTTPToolActionFailure(failure, "Workflow HTTP tool plan scope is denied."))
		return
	}
	service, err := s.workflowHTTPToolActionService()
	if err != nil {
		writeWorkflowHTTPToolActionResult(writer, trace, ctx, workflowHTTPToolActionFailure(WorkflowHTTPToolActionFailureStoreUnavailable, "Workflow HTTP tool action service is unavailable."))
		return
	}
	result := service.CreatePlan(ctx, WorkflowHTTPToolCreatePlanRequest{
		DraftID: strings.TrimSpace(request.PathValue("draft_id")), DraftVersion: body.DraftVersion,
		NodeID: body.NodeID, PublicArguments: body.PublicArguments,
	})
	writeWorkflowHTTPToolActionResult(writer, trace, ctx, result)
}

func (s *Server) handleReadWorkflowHTTPToolActionPlan(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowHTTPToolPlanReadRoute)
	if !s.allowWorkflowHTTPToolActionDev(writer, trace) {
		return
	}
	values := request.URL.Query()
	for key, entries := range values {
		if (key != "workspace_id" && key != "application_id") || len(entries) != 1 {
			writeWorkflowHTTPToolActionResult(writer, trace, WorkflowHTTPToolActionContext{RequestContext: request.Context(), RequestID: trace.requestID}, workflowHTTPToolActionFailure(WorkflowHTTPToolActionFailureInputInvalid, "Workflow HTTP tool plan query is invalid."))
			return
		}
	}
	ctx, failure := workflowHTTPToolActionContextFromRequest(request, trace, values.Get("workspace_id"), values.Get("application_id"), "read", "workflow_tool_actions:read")
	if failure != "" {
		writeWorkflowHTTPToolActionResult(writer, trace, ctx, workflowHTTPToolActionFailure(failure, "Workflow HTTP tool plan scope is denied."))
		return
	}
	service, err := s.workflowHTTPToolActionService()
	if err != nil {
		writeWorkflowHTTPToolActionResult(writer, trace, ctx, workflowHTTPToolActionFailure(WorkflowHTTPToolActionFailureStoreUnavailable, "Workflow HTTP tool action service is unavailable."))
		return
	}
	writeWorkflowHTTPToolActionResult(writer, trace, ctx, service.ReadPlan(ctx, request.PathValue("plan_id")))
}

func (s *Server) handleDecideWorkflowHTTPToolActionPlan(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowHTTPToolDecisionRoute)
	if !s.allowWorkflowHTTPToolActionDev(writer, trace) {
		return
	}
	var body workflowHTTPToolDecisionBody
	if !s.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: maxControlJSONRequestBodyBytes, rejectUnknownFields: true}) {
		return
	}
	ctx, failure := workflowHTTPToolActionContextFromRequest(request, trace, body.WorkspaceID, body.ApplicationID, "decision", "workflow_tool_actions:confirm")
	if failure != "" {
		writeWorkflowHTTPToolActionResult(writer, trace, ctx, workflowHTTPToolActionFailure(failure, "Workflow HTTP tool confirmation scope is denied."))
		return
	}
	service, err := s.workflowHTTPToolActionService()
	if err != nil {
		writeWorkflowHTTPToolActionResult(writer, trace, ctx, workflowHTTPToolActionFailure(WorkflowHTTPToolActionFailureStoreUnavailable, "Workflow HTTP tool action service is unavailable."))
		return
	}
	result := service.DecidePlan(ctx, WorkflowHTTPToolDecisionRequest{
		PlanID: strings.TrimSpace(request.PathValue("plan_id")), ExpectedRecordVersion: body.ExpectedRecordVersion, Decision: body.Decision,
	})
	writeWorkflowHTTPToolActionResult(writer, trace, ctx, result)
}

func (s *Server) handleExecuteWorkflowHTTPToolActionPlan(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowHTTPToolExecutionRoute)
	if !s.allowWorkflowHTTPToolExecutionDev(writer, trace) {
		return
	}
	var body workflowHTTPToolExecutionBody
	if !s.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: maxControlJSONRequestBodyBytes, rejectUnknownFields: true}) {
		return
	}
	ctx, failure := workflowHTTPToolActionContextFromRequest(
		request,
		trace,
		body.WorkspaceID,
		body.ApplicationID,
		"execution",
		workflowHTTPToolExecutionRequiredScopes...,
	)
	if failure != "" {
		writeWorkflowHTTPToolExecutionResult(writer, trace, ctx, workflowHTTPToolExecutionFailure(WorkflowRunFailureScopeDenied, "Workflow HTTP tool execution scope is denied."))
		return
	}
	service, err := s.workflowHTTPToolExecutionService()
	if err != nil {
		writeWorkflowHTTPToolExecutionResult(writer, trace, ctx, workflowHTTPToolExecutionFailure(WorkflowRunFailureToolStore, "Workflow HTTP tool execution service is unavailable."))
		return
	}
	result := service.Execute(ctx, WorkflowHTTPToolExecutionRequest{
		PlanID: strings.TrimSpace(request.PathValue("plan_id")), ExpectedRecordVersion: body.ExpectedRecordVersion,
		InputText: body.InputText, Model: body.Model, Temperature: body.Temperature,
	})
	writeWorkflowHTTPToolExecutionResult(writer, trace, ctx, result)
}

func (s *Server) allowWorkflowHTTPToolActionDev(writer http.ResponseWriter, trace requestTrace) bool {
	if s.config.WorkflowToolActionDevEnabled {
		return true
	}
	s.writePlatformError(writer, trace, "WORKFLOW_TOOL_ACTION_DEV_DISABLED", "workflow tool action dev route requires explicit opt-in")
	return false
}

func (s *Server) allowWorkflowHTTPToolExecutionDev(writer http.ResponseWriter, trace requestTrace) bool {
	if s.config.WorkflowToolActionDevEnabled && s.config.WorkflowHTTPToolExecutionDevEnabled {
		return true
	}
	s.writePlatformError(writer, trace, "WORKFLOW_HTTP_TOOL_EXECUTION_DEV_DISABLED", "workflow HTTP tool execution route requires both explicit development gates")
	return false
}

func (s *Server) workflowHTTPToolActionService() (workflowHTTPToolActionService, error) {
	if s.workflowHTTPToolActionStore == nil {
		return workflowHTTPToolActionService{}, errWorkflowHTTPToolActionUnavailable
	}
	return newWorkflowHTTPToolActionService(s.savedWorkflowDraftService().ReadDraft, s.workflowHTTPToolActionStore)
}

func workflowHTTPToolActionContextFromRequest(
	request *http.Request,
	trace requestTrace,
	workspaceID string,
	applicationID string,
	auditSuffix string,
	requiredScopes ...string,
) (WorkflowHTTPToolActionContext, WorkflowHTTPToolActionFailureCode) {
	runContext, failure := workflowRunContextFromRequest(request, trace, workspaceID, applicationID, "tool-action-"+auditSuffix, requiredScopes...)
	ctx := WorkflowHTTPToolActionContext{
		RequestContext: runContext.RequestContext, RequestID: runContext.RequestID, TenantRef: runContext.TenantRef,
		WorkspaceID: runContext.WorkspaceID, ApplicationID: runContext.ApplicationID, ActorRef: runContext.ActorRef,
		ScopeGrants: cloneStringSlice(runContext.ScopeGrants), AuditRef: runContext.AuditRef,
	}
	if failure != "" {
		return ctx, WorkflowHTTPToolActionFailureScopeDenied
	}
	return ctx, ""
}

func writeWorkflowHTTPToolActionResult(writer http.ResponseWriter, trace requestTrace, ctx WorkflowHTTPToolActionContext, result WorkflowHTTPToolActionResult) {
	writeObservedJSON(writer, http.StatusOK, trace, workflowHTTPToolActionEnvelope{
		RequestID: trace.requestID, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID,
		ActionPlan: result.ActionPlan, ConfirmationDecision: result.ConfirmationDecision,
		FailureCode: workflowHTTPToolActionFailurePointer(result.FailureCode), FailureSummary: result.FailureSummary,
		AuditRef: ctx.AuditRef,
	})
}

func writeWorkflowHTTPToolExecutionResult(
	writer http.ResponseWriter,
	trace requestTrace,
	ctx WorkflowHTTPToolActionContext,
	result WorkflowHTTPToolExecutionResult,
) {
	writeObservedJSON(writer, http.StatusOK, trace, workflowHTTPToolExecutionEnvelope{
		RequestID: trace.requestID, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID,
		ActionPlan: result.ActionPlan, Run: result.Record,
		FailureCode: workflowRunFailureCodePointer(result.FailureCode), FailureSummary: result.FailureSummary,
		AuditRef: ctx.AuditRef,
	})
}

func workflowHTTPToolActionFailurePointer(code WorkflowHTTPToolActionFailureCode) *string {
	if code == "" {
		return nil
	}
	value := string(code)
	return &value
}
