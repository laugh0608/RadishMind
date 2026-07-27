package httpapi

import (
	"net/http"
	"strings"
)

const (
	agentCopilotRuntimeReadRoute     = "GET /v1/user-workspace/applications/{application_id}/agent-copilot-runtime-assignment"
	agentCopilotRuntimeEventsRoute   = "GET /v1/user-workspace/applications/{application_id}/agent-copilot-runtime-assignment/events"
	agentCopilotRuntimeDecisionRoute = "POST /v1/user-workspace/applications/{application_id}/agent-copilot-runtime-assignment/decisions"

	agentCopilotRuntimeWorkspaceHeader   = "X-RadishMind-Dev-Agent-Copilot-Runtime-Workspace"
	agentCopilotRuntimeApplicationHeader = "X-RadishMind-Dev-Agent-Copilot-Runtime-Application"
)

type agentCopilotRuntimeDecisionBody struct {
	WorkspaceID               string `json:"workspace_id"`
	ExpectedAssignmentVersion int    `json:"expected_assignment_version"`
	Action                    string `json:"action"`
	CandidateID               string `json:"candidate_id"`
}

type agentCopilotRuntimeEnvelope struct {
	RequestID                string                                 `json:"request_id"`
	TenantRef                string                                 `json:"tenant_ref"`
	WorkspaceID              string                                 `json:"workspace_id"`
	ApplicationID            string                                 `json:"application_id"`
	Assignment               *AgentCopilotRuntimeAssignmentV1       `json:"assignment"`
	Events                   []AgentCopilotRuntimeAssignmentEventV1 `json:"events"`
	FailureCode              *string                                `json:"failure_code"`
	CurrentAssignmentVersion int                                    `json:"current_assignment_version"`
	CurrentState             string                                 `json:"current_state"`
	AuditRef                 string                                 `json:"audit_ref"`
}

func (server *Server) handleReadAgentCopilotRuntimeAssignment(writer http.ResponseWriter, request *http.Request) {
	server.handleReadAgentCopilotRuntime(writer, request, false)
}

func (server *Server) handleReadAgentCopilotRuntimeEvents(writer http.ResponseWriter, request *http.Request) {
	server.handleReadAgentCopilotRuntime(writer, request, true)
}

func (server *Server) handleReadAgentCopilotRuntime(writer http.ResponseWriter, request *http.Request, includeEvents bool) {
	route := agentCopilotRuntimeReadRoute
	if includeEvents {
		route = agentCopilotRuntimeEventsRoute
	}
	trace := newRequestTrace(request, route)
	if !server.allowAgentCopilotRuntimeDevHTTP(writer, request, trace) {
		return
	}
	ctx, failure := agentCopilotRuntimeContextFromRequest(
		request, trace, request.URL.Query().Get("workspace_id"), request.PathValue("application_id"),
		"agent_copilot_runtime:read", false, "read",
	)
	if failure != "" {
		writeAgentCopilotRuntimeResult(writer, trace, ctx, agentCopilotRuntimeFailure(failure), includeEvents)
		return
	}
	writeAgentCopilotRuntimeResult(writer, trace, ctx, server.agentCopilotRuntimeService().Read(ctx), includeEvents)
}

func (server *Server) handleDecideAgentCopilotRuntimeAssignment(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, agentCopilotRuntimeDecisionRoute)
	if !server.allowAgentCopilotRuntimeDevHTTP(writer, request, trace) {
		return
	}
	auth, failure, status := server.authorizeWorkspaceScopedPermissions(request, "agent_copilot_runtime:write")
	ctx := agentCopilotRuntimeMutationContext(
		request, trace, auth, request.PathValue("application_id"),
		server.config.AgentCopilotRuntimeDevWriteEnabled, "decision",
	)
	if failure != "" {
		writeAgentCopilotRuntimeResultWithStatus(writer, status, trace, ctx, agentCopilotRuntimeFailure(failure), true)
		return
	}
	var body agentCopilotRuntimeDecisionBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: maxControlJSONRequestBodyBytes, rejectUnknownFields: true}) {
		return
	}
	if body.WorkspaceID != auth.ResourceBinding.WorkspaceID ||
		strings.TrimSpace(request.Header.Get(agentCopilotRuntimeWorkspaceHeader)) != auth.ResourceBinding.WorkspaceID ||
		strings.TrimSpace(request.Header.Get(agentCopilotRuntimeApplicationHeader)) != ctx.ApplicationID ||
		!validControlPlaneReadAuthReference(ctx.ApplicationID, false) {
		writeAgentCopilotRuntimeResultWithStatus(writer, http.StatusForbidden, trace, ctx, agentCopilotRuntimeFailure("workspace_binding_mismatch"), true)
		return
	}
	result := server.agentCopilotRuntimeService().Decide(ctx, AgentCopilotRuntimeDecisionInput{
		ExpectedAssignmentVersion: body.ExpectedAssignmentVersion,
		Action:                    body.Action,
		CandidateID:               body.CandidateID,
	})
	writeAgentCopilotRuntimeResult(writer, trace, ctx, result, true)
}

func (server *Server) allowAgentCopilotRuntimeDevHTTP(writer http.ResponseWriter, request *http.Request, trace requestTrace) bool {
	if server.config.AgentCopilotRuntimeDevHTTPEnabled {
		return true
	}
	server.writePlatformError(writer, trace, "AGENT_COPILOT_RUNTIME_DEV_HTTP_DISABLED", "Agent Copilot runtime dev route requires explicit opt-in")
	return false
}

func (server *Server) agentCopilotRuntimeService() agentCopilotRuntimeService {
	repository := server.agentCopilotRuntimeRepository
	if repository == nil {
		repository = &memoryAgentCopilotRuntimeRepository{entries: make(map[string]agentCopilotRuntimeMemoryEntry), unavailable: true}
	}
	resolver := agentCopilotRuntimeAuthorityResolver{
		publishRepository: server.applicationPublishCandidateRepository,
		draftRepository:   server.applicationDraftRepository,
		profileRepository: server.agentCopilotProfileRepository,
		readApplication:   server.readApplicationPublishBaseline,
	}
	return newAgentCopilotRuntimeService(repository, resolver)
}

func agentCopilotRuntimeContextFromRequest(
	request *http.Request,
	trace requestTrace,
	workspaceID string,
	applicationID string,
	requiredScope string,
	writeEnabled bool,
	auditSuffix string,
) (AgentCopilotRuntimeContext, string) {
	auth, ok := request.Context().Value(controlPlaneReadAuthContextKey{}).(controlPlaneReadAuthContext)
	ctx := AgentCopilotRuntimeContext{
		RequestContext: request.Context(), RequestID: trace.requestID,
		WorkspaceID: strings.TrimSpace(workspaceID), ApplicationID: strings.TrimSpace(applicationID),
		WriteEnabled: writeEnabled, AuditRef: "audit_" + trace.requestID + "_agent-copilot-runtime-" + auditSuffix,
	}
	if !ok || strings.TrimSpace(auth.IdentityContext) == "" || strings.TrimSpace(auth.SubjectBinding) == "" ||
		!controlPlaneReadHasScope(auth.ScopeGrants, requiredScope) {
		return ctx, AgentCopilotRuntimeFailureScopeDenied
	}
	ctx.TenantRef = strings.TrimSpace(auth.TenantBinding)
	ctx.ActorRef, ctx.OwnerSubjectRef = strings.TrimSpace(auth.SubjectBinding), strings.TrimSpace(auth.SubjectBinding)
	if ctx.TenantRef == "" || ctx.WorkspaceID == "" || !applicationCatalogIDPattern.MatchString(ctx.ApplicationID) ||
		strings.TrimSpace(request.Header.Get(agentCopilotRuntimeWorkspaceHeader)) != ctx.WorkspaceID ||
		strings.TrimSpace(request.Header.Get(agentCopilotRuntimeApplicationHeader)) != ctx.ApplicationID {
		return ctx, AgentCopilotRuntimeFailureScopeDenied
	}
	return ctx, ""
}

func agentCopilotRuntimeMutationContext(
	request *http.Request,
	trace requestTrace,
	auth controlPlaneReadAuthContext,
	applicationID string,
	writeEnabled bool,
	auditSuffix string,
) AgentCopilotRuntimeContext {
	return AgentCopilotRuntimeContext{
		RequestContext: request.Context(), RequestID: trace.requestID,
		TenantRef: strings.TrimSpace(auth.TenantBinding), WorkspaceID: strings.TrimSpace(auth.ResourceBinding.WorkspaceID),
		ApplicationID: strings.TrimSpace(applicationID), ActorRef: strings.TrimSpace(auth.SubjectBinding),
		OwnerSubjectRef: strings.TrimSpace(auth.SubjectBinding), WriteEnabled: writeEnabled,
		AuditRef: "audit_" + trace.requestID + "_agent-copilot-runtime-" + auditSuffix,
	}
}

func writeAgentCopilotRuntimeResult(
	writer http.ResponseWriter,
	trace requestTrace,
	ctx AgentCopilotRuntimeContext,
	result AgentCopilotRuntimeResult,
	includeEvents bool,
) {
	writeAgentCopilotRuntimeResultWithStatus(writer, http.StatusOK, trace, ctx, result, includeEvents)
}

func writeAgentCopilotRuntimeResultWithStatus(
	writer http.ResponseWriter,
	status int,
	trace requestTrace,
	ctx AgentCopilotRuntimeContext,
	result AgentCopilotRuntimeResult,
	includeEvents bool,
) {
	events := []AgentCopilotRuntimeAssignmentEventV1{}
	if includeEvents && result.Events != nil {
		events = result.Events
	}
	writeObservedJSON(writer, status, trace, agentCopilotRuntimeEnvelope{
		RequestID: trace.requestID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID,
		ApplicationID: ctx.ApplicationID, Assignment: result.Assignment, Events: events,
		FailureCode:              optionalApplicationDraftFailure(result.FailureCode),
		CurrentAssignmentVersion: result.CurrentAssignmentVersion,
		CurrentState:             result.CurrentState, AuditRef: ctx.AuditRef,
	})
}
