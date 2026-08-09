package httpapi

import (
	"context"
	"net/http"
	"strings"
)

const (
	workflowRAGApplicationRuntimeAssignmentReadRoute     = "GET /v1/user-workspace/applications/{application_id}/workflow-rag-runtime-assignment"
	workflowRAGApplicationRuntimeAssignmentDecisionRoute = "POST /v1/user-workspace/applications/{application_id}/workflow-rag-runtime-assignment/decisions"
)

type workflowRAGApplicationRuntimeDecisionBody struct {
	WorkspaceID           string `json:"workspace_id"`
	ExpectedRecordVersion int    `json:"expected_record_version"`
	Decision              string `json:"decision"`
	PublishCandidateID    string `json:"publish_candidate_id"`
	Reason                string `json:"reason"`
}

type workflowRAGApplicationInvocationBody struct {
	Input string `json:"input"`
}

type workflowRAGApplicationRuntimeEnvelope struct {
	RequestID            string                                   `json:"request_id"`
	TenantRef            string                                   `json:"tenant_ref"`
	WorkspaceID          string                                   `json:"workspace_id"`
	ApplicationID        string                                   `json:"application_id"`
	Assignment           *WorkflowRAGApplicationRuntimeAssignment `json:"assignment"`
	Events               []WorkflowRAGApplicationRuntimeEvent     `json:"events"`
	Audits               []WorkflowRAGApplicationRuntimeAudit     `json:"audits"`
	FailureCode          *string                                  `json:"failure_code"`
	CurrentRecordVersion int                                      `json:"current_record_version"`
	CurrentState         string                                   `json:"current_state"`
	AuditRef             string                                   `json:"audit_ref"`
}

type workflowRAGApplicationInvocationEnvelope struct {
	RequestID      string                        `json:"request_id"`
	TenantRef      string                        `json:"tenant_ref"`
	WorkspaceID    string                        `json:"workspace_id"`
	ApplicationID  string                        `json:"application_id"`
	Run            *WorkflowRunRecord            `json:"run"`
	Answer         *WorkflowRAGApplicationAnswer `json:"answer"`
	FailureCode    *string                       `json:"failure_code"`
	FailureSummary string                        `json:"failure_summary"`
	AuditRef       string                        `json:"audit_ref"`
}

func (server *Server) handleReadWorkflowRAGApplicationRuntimeAssignment(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowRAGApplicationRuntimeAssignmentReadRoute)
	if !server.allowWorkflowRAGApplicationInvocationDev(writer, trace) {
		return
	}
	values := request.URL.Query()
	if !workflowRAGEvaluationQueryAllowed(values, "workspace_id") {
		writeWorkflowRAGApplicationRuntimeResult(writer, trace, WorkflowRAGApplicationRuntimeContext{}, workflowRAGApplicationRuntimeFailure(WorkflowRAGApplicationFailurePayloadInvalid))
		return
	}
	ctx, failure := workflowRAGApplicationRuntimeContextFromRequest(request, trace, values.Get("workspace_id"), request.PathValue("application_id"), "read", "workflow_rag_runtime:read")
	if failure != "" {
		writeWorkflowRAGApplicationRuntimeResult(writer, trace, ctx, workflowRAGApplicationRuntimeFailure(failure))
		return
	}
	writeWorkflowRAGApplicationRuntimeResult(writer, trace, ctx, server.workflowRAGApplicationRuntimeService().Read(ctx))
}

func (server *Server) handleDecideWorkflowRAGApplicationRuntimeAssignment(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowRAGApplicationRuntimeAssignmentDecisionRoute)
	if !server.allowWorkflowRAGApplicationInvocationDev(writer, trace) {
		return
	}
	auth, failure, status := server.authorizeWorkspaceScopedPermissions(request, "workflow_rag_runtime:write")
	ctx := workflowRAGApplicationRuntimeMutationContext(request, trace, auth, request.PathValue("application_id"), "decision")
	if failure != "" {
		writeWorkflowRAGApplicationRuntimeResultWithStatus(writer, status, trace, ctx, workflowRAGApplicationRuntimeFailure(failure))
		return
	}
	var body workflowRAGApplicationRuntimeDecisionBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: maxControlJSONRequestBodyBytes, rejectUnknownFields: true}) {
		return
	}
	if body.WorkspaceID != auth.ResourceBinding.WorkspaceID ||
		strings.TrimSpace(request.Header.Get(savedWorkflowDraftDevWorkspaceHeader)) != auth.ResourceBinding.WorkspaceID ||
		strings.TrimSpace(request.Header.Get(savedWorkflowDraftDevApplicationHeader)) != ctx.ApplicationID ||
		!validControlPlaneReadAuthReference(ctx.ApplicationID, false) {
		writeWorkflowRAGApplicationRuntimeResultWithStatus(writer, http.StatusForbidden, trace, ctx, workflowRAGApplicationRuntimeFailure("workspace_binding_mismatch"))
		return
	}
	result := server.workflowRAGApplicationRuntimeService().Decide(ctx, WorkflowRAGApplicationRuntimeDecisionInput{ExpectedRecordVersion: body.ExpectedRecordVersion, Decision: body.Decision, PublishCandidateID: body.PublishCandidateID, Reason: body.Reason})
	writeWorkflowRAGApplicationRuntimeResult(writer, trace, ctx, result)
}

func (server *Server) handleWorkflowRAGApplicationInvocation(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowRAGApplicationInvocationRoute)
	if !server.allowWorkflowRAGApplicationInvocationDev(writer, trace) {
		return
	}
	authentication := server.authenticateGatewayAPIKey(request, trace, "application_rag:invoke")
	if authentication.FailureCode != "" {
		server.writePlatformError(writer, trace, authentication.FailureCode, "")
		return
	}
	var body workflowRAGApplicationInvocationBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: workflowRAGApplicationInvocationMaxBytes + 256, rejectUnknownFields: true}) {
		return
	}
	gatewayContext := authentication.RequestContext
	ctx := WorkflowRAGApplicationRuntimeContext{RequestContext: gatewayContext.RequestContext, RequestID: trace.requestID, TenantRef: gatewayContext.TenantRef, WorkspaceID: gatewayContext.WorkspaceID, ApplicationID: gatewayContext.ApplicationID, ActorRef: gatewayContext.SubjectRef, OwnerSubjectRef: gatewayContext.SubjectRef, AuditRef: "audit_" + trace.requestID + "_application-rag-invocation"}
	result := server.workflowRAGApplicationInvocationService().Invoke(ctx, WorkflowRAGApplicationInvocationInput{Input: body.Input})
	status := http.StatusOK
	if gatewayRequestQuotaFailureCodeFromValue(result.FailureCode) != "" {
		status = gatewayRequestQuotaHTTPStatus(result.FailureCode)
	}
	writeObservedJSON(writer, status, trace, workflowRAGApplicationInvocationEnvelope{RequestID: trace.requestID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, Run: result.Run, Answer: result.Answer, FailureCode: workflowRAGApplicationFailurePointer(result.FailureCode), FailureSummary: result.FailureSummary, AuditRef: ctx.AuditRef})
}

func (server *Server) allowWorkflowRAGApplicationInvocationDev(writer http.ResponseWriter, trace requestTrace) bool {
	if server.config.WorkflowRAGAppInvocationDevEnabled {
		return true
	}
	server.writePlatformError(writer, trace, "WORKFLOW_RAG_APPLICATION_INVOCATION_DEV_DISABLED", "workflow RAG application invocation requires explicit development opt-in")
	return false
}

func (server *Server) workflowRAGApplicationAuthorityResolver() workflowRAGApplicationAuthorityResolver {
	return workflowRAGApplicationAuthorityResolver{publishRepository: server.applicationPublishCandidateRepository, draftRepository: server.applicationDraftRepository, promotionService: server.workflowRAGPromotionService(), readApplication: server.readWorkflowRAGPromotionApplication}
}

func (server *Server) workflowRAGApplicationRuntimeService() workflowRAGApplicationRuntimeService {
	return newWorkflowRAGApplicationRuntimeService(server.workflowRAGAppRuntimeRepository, server.workflowRAGApplicationAuthorityResolver())
}

func (server *Server) workflowRAGApplicationInvocationService() workflowRAGApplicationInvocationService {
	service := newWorkflowRAGApplicationInvocationService(server.workflowRAGAppRuntimeRepository, server.workflowRAGApplicationAuthorityResolver(), server.workflowRunStore, server.bridge)
	service.defaultTemperature = server.config.Temperature
	service.resolveSelection = func(ctx context.Context, model string) northboundSelection {
		return server.resolveNorthboundSelection(ctx, model, nil)
	}
	service.envelopeOptions = server.buildBridgeEnvelopeOptions
	return service
}

func workflowRAGApplicationRuntimeContextFromRequest(request *http.Request, trace requestTrace, workspaceID, applicationID, suffix string, scopes ...string) (WorkflowRAGApplicationRuntimeContext, string) {
	runContext, failure := workflowRunContextFromRequest(request, trace, workspaceID, applicationID, "rag-runtime-"+suffix, scopes...)
	ctx := WorkflowRAGApplicationRuntimeContext{RequestContext: request.Context(), RequestID: trace.requestID, TenantRef: runContext.TenantRef, WorkspaceID: runContext.WorkspaceID, ApplicationID: runContext.ApplicationID, ActorRef: runContext.ActorRef, OwnerSubjectRef: runContext.ActorRef, AuditRef: runContext.AuditRef, WriteEnabled: true}
	if failure != "" || !workflowRAGHasVerifiedActor(request) {
		return ctx, WorkflowRAGApplicationFailureScopeDenied
	}
	return ctx, ""
}

func workflowRAGApplicationRuntimeMutationContext(
	request *http.Request,
	trace requestTrace,
	auth controlPlaneReadAuthContext,
	applicationID string,
	suffix string,
) WorkflowRAGApplicationRuntimeContext {
	return WorkflowRAGApplicationRuntimeContext{
		RequestContext: request.Context(), RequestID: trace.requestID,
		TenantRef: strings.TrimSpace(auth.TenantBinding), WorkspaceID: strings.TrimSpace(auth.ResourceBinding.WorkspaceID),
		ApplicationID: strings.TrimSpace(applicationID), ActorRef: strings.TrimSpace(auth.SubjectBinding),
		OwnerSubjectRef: strings.TrimSpace(auth.SubjectBinding), WriteEnabled: true,
		AuditRef: auditRefForWorkflowRun(trace, "rag-runtime-"+suffix),
	}
}

func writeWorkflowRAGApplicationRuntimeResult(writer http.ResponseWriter, trace requestTrace, ctx WorkflowRAGApplicationRuntimeContext, result WorkflowRAGApplicationRuntimeResult) {
	writeWorkflowRAGApplicationRuntimeResultWithStatus(writer, http.StatusOK, trace, ctx, result)
}

func writeWorkflowRAGApplicationRuntimeResultWithStatus(writer http.ResponseWriter, status int, trace requestTrace, ctx WorkflowRAGApplicationRuntimeContext, result WorkflowRAGApplicationRuntimeResult) {
	if result.Events == nil {
		result.Events = []WorkflowRAGApplicationRuntimeEvent{}
	}
	if result.Audits == nil {
		result.Audits = []WorkflowRAGApplicationRuntimeAudit{}
	}
	writeObservedJSON(writer, status, trace, workflowRAGApplicationRuntimeEnvelope{RequestID: trace.requestID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, Assignment: result.Assignment, Events: result.Events, Audits: result.Audits, FailureCode: workflowRAGApplicationFailurePointer(result.FailureCode), CurrentRecordVersion: result.CurrentRecordVersion, CurrentState: result.CurrentState, AuditRef: ctx.AuditRef})
}

func workflowRAGApplicationFailurePointer(code string) *string {
	if strings.TrimSpace(code) == "" {
		return nil
	}
	value := strings.TrimSpace(code)
	return &value
}
