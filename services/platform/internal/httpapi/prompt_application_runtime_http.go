package httpapi

import (
	"net/http"
	"strings"
)

const (
	promptApplicationRuntimeReadRoute     = "GET /v1/user-workspace/applications/{application_id}/prompt-runtime-assignment"
	promptApplicationRuntimeEventsRoute   = "GET /v1/user-workspace/applications/{application_id}/prompt-runtime-assignment/events"
	promptApplicationRuntimeDecisionRoute = "POST /v1/user-workspace/applications/{application_id}/prompt-runtime-assignment/decisions"

	promptApplicationRuntimeWorkspaceHeader   = "X-RadishMind-Dev-Prompt-Runtime-Workspace"
	promptApplicationRuntimeApplicationHeader = "X-RadishMind-Dev-Prompt-Runtime-Application"
)

type promptApplicationRuntimeDecisionBody struct {
	WorkspaceID               string `json:"workspace_id"`
	ExpectedAssignmentVersion int    `json:"expected_assignment_version"`
	Action                    string `json:"action"`
	CandidateID               string `json:"candidate_id"`
}

type promptApplicationRuntimeEnvelope struct {
	RequestID                string                                    `json:"request_id"`
	TenantRef                string                                    `json:"tenant_ref"`
	WorkspaceID              string                                    `json:"workspace_id"`
	ApplicationID            string                                    `json:"application_id"`
	Assignment               *PromptApplicationRuntimeAssignment       `json:"assignment"`
	Events                   []PromptApplicationRuntimeAssignmentEvent `json:"events"`
	FailureCode              *string                                   `json:"failure_code"`
	CurrentAssignmentVersion int                                       `json:"current_assignment_version"`
	CurrentState             string                                    `json:"current_state"`
	AuditRef                 string                                    `json:"audit_ref"`
}

func (server *Server) handleReadPromptApplicationRuntimeAssignment(writer http.ResponseWriter, request *http.Request) {
	server.handleReadPromptApplicationRuntime(writer, request, false)
}

func (server *Server) handleReadPromptApplicationRuntimeEvents(writer http.ResponseWriter, request *http.Request) {
	server.handleReadPromptApplicationRuntime(writer, request, true)
}

func (server *Server) handleReadPromptApplicationRuntime(writer http.ResponseWriter, request *http.Request, includeEvents bool) {
	route := promptApplicationRuntimeReadRoute
	if includeEvents {
		route = promptApplicationRuntimeEventsRoute
	}
	trace := newRequestTrace(request, route)
	if !server.allowPromptApplicationRuntimeDevHTTP(writer, request, trace) {
		return
	}
	ctx, failure := promptApplicationRuntimeContextFromRequest(request, trace, request.URL.Query().Get("workspace_id"), request.PathValue("application_id"), "prompt_application_runtime:read", false, "read")
	if failure != "" {
		writePromptApplicationRuntimeResult(writer, trace, ctx, promptApplicationRuntimeFailure(failure), includeEvents)
		return
	}
	writePromptApplicationRuntimeResult(writer, trace, ctx, server.promptApplicationRuntimeService().Read(ctx), includeEvents)
}

func (server *Server) handleDecidePromptApplicationRuntimeAssignment(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, promptApplicationRuntimeDecisionRoute)
	if !server.allowPromptApplicationRuntimeDevHTTP(writer, request, trace) {
		return
	}
	auth, failure, status := server.authorizeWorkspaceScopedPermissions(request, "prompt_application_runtime:write")
	ctx := promptApplicationRuntimeMutationContext(
		request, trace, auth, request.PathValue("application_id"),
		server.config.PromptApplicationRuntimeDevWriteEnabled, "decision",
	)
	if failure != "" {
		writePromptApplicationRuntimeResultWithStatus(writer, status, trace, ctx, promptApplicationRuntimeFailure(failure), true)
		return
	}
	var body promptApplicationRuntimeDecisionBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: maxControlJSONRequestBodyBytes, rejectUnknownFields: true}) {
		return
	}
	if body.WorkspaceID != auth.ResourceBinding.WorkspaceID ||
		strings.TrimSpace(request.Header.Get(promptApplicationRuntimeWorkspaceHeader)) != auth.ResourceBinding.WorkspaceID ||
		strings.TrimSpace(request.Header.Get(promptApplicationRuntimeApplicationHeader)) != ctx.ApplicationID ||
		!validControlPlaneReadAuthReference(ctx.ApplicationID, false) {
		writePromptApplicationRuntimeResultWithStatus(writer, http.StatusForbidden, trace, ctx, promptApplicationRuntimeFailure("workspace_binding_mismatch"), true)
		return
	}
	result := server.promptApplicationRuntimeService().Decide(ctx, PromptApplicationRuntimeDecisionInput{
		ExpectedAssignmentVersion: body.ExpectedAssignmentVersion, Action: body.Action, CandidateID: body.CandidateID,
	})
	writePromptApplicationRuntimeResult(writer, trace, ctx, result, true)
}

func (server *Server) allowPromptApplicationRuntimeDevHTTP(writer http.ResponseWriter, request *http.Request, trace requestTrace) bool {
	if server.config.PromptApplicationRuntimeDevHTTPEnabled {
		return true
	}
	server.writePlatformError(writer, trace, "PROMPT_APPLICATION_RUNTIME_DEV_HTTP_DISABLED", "Prompt application runtime dev route requires explicit opt-in")
	return false
}

func (server *Server) promptApplicationRuntimeService() promptApplicationRuntimeService {
	repository := server.promptApplicationRuntimeRepository
	if repository == nil {
		repository = &memoryPromptApplicationRuntimeRepository{entries: make(map[string]promptApplicationRuntimeMemoryEntry), unavailable: true}
	}
	resolver := promptApplicationRuntimeAuthorityResolver{
		publishRepository: server.applicationPublishCandidateRepository, draftRepository: server.applicationDraftRepository,
		templateRepository: server.promptApplicationTemplateRepository, readApplication: server.readApplicationPublishBaseline,
	}
	return newPromptApplicationRuntimeService(repository, resolver)
}

func promptApplicationRuntimeContextFromRequest(request *http.Request, trace requestTrace, workspaceID, applicationID, requiredScope string, writeEnabled bool, auditSuffix string) (PromptApplicationRuntimeContext, string) {
	auth, ok := request.Context().Value(controlPlaneReadAuthContextKey{}).(controlPlaneReadAuthContext)
	ctx := PromptApplicationRuntimeContext{
		RequestContext: request.Context(), RequestID: trace.requestID, WorkspaceID: strings.TrimSpace(workspaceID), ApplicationID: strings.TrimSpace(applicationID),
		WriteEnabled: writeEnabled, AuditRef: "audit_" + trace.requestID + "_prompt-runtime-" + auditSuffix,
	}
	if !ok || strings.TrimSpace(auth.IdentityContext) == "" || strings.TrimSpace(auth.SubjectBinding) == "" || !controlPlaneReadHasScope(auth.ScopeGrants, requiredScope) {
		return ctx, PromptApplicationRuntimeFailureScopeDenied
	}
	ctx.TenantRef, ctx.ActorRef, ctx.OwnerSubjectRef = strings.TrimSpace(auth.TenantBinding), strings.TrimSpace(auth.SubjectBinding), strings.TrimSpace(auth.SubjectBinding)
	if ctx.TenantRef == "" || ctx.WorkspaceID == "" || !applicationCatalogIDPattern.MatchString(ctx.ApplicationID) ||
		strings.TrimSpace(request.Header.Get(promptApplicationRuntimeWorkspaceHeader)) != ctx.WorkspaceID ||
		strings.TrimSpace(request.Header.Get(promptApplicationRuntimeApplicationHeader)) != ctx.ApplicationID {
		return ctx, PromptApplicationRuntimeFailureScopeDenied
	}
	return ctx, ""
}

func promptApplicationRuntimeMutationContext(
	request *http.Request,
	trace requestTrace,
	auth controlPlaneReadAuthContext,
	applicationID string,
	writeEnabled bool,
	auditSuffix string,
) PromptApplicationRuntimeContext {
	return PromptApplicationRuntimeContext{
		RequestContext: request.Context(), RequestID: trace.requestID,
		TenantRef: strings.TrimSpace(auth.TenantBinding), WorkspaceID: strings.TrimSpace(auth.ResourceBinding.WorkspaceID),
		ApplicationID: strings.TrimSpace(applicationID), ActorRef: strings.TrimSpace(auth.SubjectBinding),
		OwnerSubjectRef: strings.TrimSpace(auth.SubjectBinding), WriteEnabled: writeEnabled,
		AuditRef: "audit_" + trace.requestID + "_prompt-runtime-" + auditSuffix,
	}
}

func writePromptApplicationRuntimeResult(writer http.ResponseWriter, trace requestTrace, ctx PromptApplicationRuntimeContext, result PromptApplicationRuntimeResult, includeEvents bool) {
	writePromptApplicationRuntimeResultWithStatus(writer, http.StatusOK, trace, ctx, result, includeEvents)
}

func writePromptApplicationRuntimeResultWithStatus(writer http.ResponseWriter, status int, trace requestTrace, ctx PromptApplicationRuntimeContext, result PromptApplicationRuntimeResult, includeEvents bool) {
	events := []PromptApplicationRuntimeAssignmentEvent{}
	if includeEvents && result.Events != nil {
		events = result.Events
	}
	writeObservedJSON(writer, status, trace, promptApplicationRuntimeEnvelope{
		RequestID: trace.requestID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID,
		Assignment: result.Assignment, Events: events, FailureCode: optionalApplicationDraftFailure(result.FailureCode),
		CurrentAssignmentVersion: result.CurrentAssignmentVersion, CurrentState: result.CurrentState, AuditRef: ctx.AuditRef,
	})
}
