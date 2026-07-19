package httpapi

import (
	"net/http"
	"strconv"
	"strings"
)

const (
	applicationSessionCreateRoute   = "POST /v1/user-workspace/application-sessions"
	applicationSessionListRoute     = "GET /v1/user-workspace/application-sessions"
	applicationSessionReadRoute     = "GET /v1/user-workspace/application-sessions/{session_id}"
	applicationSessionCloseRoute    = "POST /v1/user-workspace/application-sessions/{session_id}/close"
	applicationSessionTurnListRoute = "GET /v1/user-workspace/application-sessions/{session_id}/turns"
)

type applicationInteractionSessionCreateBody struct {
	WorkspaceID      string `json:"workspace_id"`
	ApplicationID    string `json:"application_id"`
	ExecutionProfile string `json:"execution_profile"`
	DefinitionID     string `json:"definition_id,omitempty"`
}

type applicationInteractionSessionCloseBody struct {
	WorkspaceID     string `json:"workspace_id"`
	ApplicationID   string `json:"application_id"`
	ExpectedVersion int    `json:"expected_version"`
}

type applicationInteractionSessionEnvelope struct {
	RequestID            string                         `json:"request_id"`
	TenantRef            string                         `json:"tenant_ref"`
	WorkspaceID          string                         `json:"workspace_id"`
	ApplicationID        string                         `json:"application_id"`
	Session              *ApplicationInteractionSession `json:"session"`
	FailureCode          *string                        `json:"failure_code"`
	CurrentRecordVersion int                            `json:"current_record_version"`
	CurrentState         string                         `json:"current_state"`
	IdempotentReplay     bool                           `json:"idempotent_replay"`
	AuditRef             string                         `json:"audit_ref"`
}

type applicationInteractionSessionListEnvelope struct {
	RequestID     string                          `json:"request_id"`
	TenantRef     string                          `json:"tenant_ref"`
	WorkspaceID   string                          `json:"workspace_id"`
	ApplicationID string                          `json:"application_id"`
	Items         []ApplicationInteractionSession `json:"items"`
	NextCursor    *string                         `json:"next_cursor"`
	FailureCode   *string                         `json:"failure_code"`
	AuditRef      string                          `json:"audit_ref"`
}

type applicationInteractionTurnListEnvelope struct {
	RequestID     string                       `json:"request_id"`
	TenantRef     string                       `json:"tenant_ref"`
	WorkspaceID   string                       `json:"workspace_id"`
	ApplicationID string                       `json:"application_id"`
	SessionID     string                       `json:"session_id"`
	Items         []ApplicationInteractionTurn `json:"items"`
	FailureCode   *string                      `json:"failure_code"`
	AuditRef      string                       `json:"audit_ref"`
}

func (server *Server) handleCreateApplicationInteractionSession(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, applicationSessionCreateRoute)
	if !server.allowApplicationInteractionSessionDev(writer, trace) {
		return
	}
	var body applicationInteractionSessionCreateBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: maxControlJSONRequestBodyBytes, rejectUnknownFields: true}) {
		return
	}
	ctx, failure, status := applicationInteractionContextFromRequest(request, trace, body.WorkspaceID, body.ApplicationID, "create", "application_sessions:write")
	if failure != "" {
		writeApplicationInteractionSessionResult(writer, status, trace, ctx, applicationInteractionSessionFailure(failure))
		return
	}
	ctx.WriteEnabled = true
	result := server.applicationInteractionSessionService().Create(ctx, ApplicationInteractionSessionCreateInput{ProfileBinding: ApplicationInteractionProfileBinding{ExecutionProfile: strings.TrimSpace(body.ExecutionProfile), DefinitionID: strings.TrimSpace(body.DefinitionID)}})
	writeApplicationInteractionSessionResult(writer, http.StatusOK, trace, ctx, result)
}

func (server *Server) handleListApplicationInteractionSessions(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, applicationSessionListRoute)
	if !server.allowApplicationInteractionSessionDev(writer, trace) {
		return
	}
	values := request.URL.Query()
	for key := range values {
		switch key {
		case "workspace_id", "application_id", "state", "execution_profile", "limit", "cursor":
		default:
			ctx, _, _ := applicationInteractionContextFromRequest(request, trace, values.Get("workspace_id"), values.Get("application_id"), "list", "application_sessions:read")
			writeApplicationInteractionSessionListResult(writer, http.StatusBadRequest, trace, ctx, ApplicationInteractionSessionListResult{Sessions: []ApplicationInteractionSession{}, FailureCode: ApplicationInteractionFailurePayloadInvalid})
			return
		}
	}
	ctx, failure, status := applicationInteractionContextFromRequest(request, trace, values.Get("workspace_id"), values.Get("application_id"), "list", "application_sessions:read")
	if failure != "" {
		writeApplicationInteractionSessionListResult(writer, status, trace, ctx, ApplicationInteractionSessionListResult{Sessions: []ApplicationInteractionSession{}, FailureCode: failure})
		return
	}
	limit := 0
	if raw := strings.TrimSpace(values.Get("limit")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			writeApplicationInteractionSessionListResult(writer, http.StatusBadRequest, trace, ctx, ApplicationInteractionSessionListResult{Sessions: []ApplicationInteractionSession{}, FailureCode: ApplicationInteractionFailurePayloadInvalid})
			return
		}
		limit = parsed
	}
	result := server.applicationInteractionSessionService().List(ctx, ApplicationInteractionSessionListInput{State: values.Get("state"), ExecutionProfile: values.Get("execution_profile"), Limit: limit, Cursor: values.Get("cursor")})
	writeApplicationInteractionSessionListResult(writer, http.StatusOK, trace, ctx, result)
}

func (server *Server) handleReadApplicationInteractionSession(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, applicationSessionReadRoute)
	if !server.allowApplicationInteractionSessionDev(writer, trace) {
		return
	}
	values := request.URL.Query()
	if !applicationInteractionSessionQueryAllowed(values, "workspace_id", "application_id") {
		ctx, _, _ := applicationInteractionContextFromRequest(request, trace, values.Get("workspace_id"), values.Get("application_id"), "read", "application_sessions:read")
		writeApplicationInteractionSessionResult(writer, http.StatusBadRequest, trace, ctx, applicationInteractionSessionFailure(ApplicationInteractionFailurePayloadInvalid))
		return
	}
	ctx, failure, status := applicationInteractionContextFromRequest(request, trace, values.Get("workspace_id"), values.Get("application_id"), "read", "application_sessions:read")
	if failure != "" {
		writeApplicationInteractionSessionResult(writer, status, trace, ctx, applicationInteractionSessionFailure(failure))
		return
	}
	writeApplicationInteractionSessionResult(writer, http.StatusOK, trace, ctx, server.applicationInteractionSessionService().Read(ctx, request.PathValue("session_id")))
}

func (server *Server) handleCloseApplicationInteractionSession(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, applicationSessionCloseRoute)
	if !server.allowApplicationInteractionSessionDev(writer, trace) {
		return
	}
	var body applicationInteractionSessionCloseBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: maxControlJSONRequestBodyBytes, rejectUnknownFields: true}) {
		return
	}
	ctx, failure, status := applicationInteractionContextFromRequest(request, trace, body.WorkspaceID, body.ApplicationID, "close", "application_sessions:write")
	if failure != "" {
		writeApplicationInteractionSessionResult(writer, status, trace, ctx, applicationInteractionSessionFailure(failure))
		return
	}
	ctx.WriteEnabled = true
	writeApplicationInteractionSessionResult(writer, http.StatusOK, trace, ctx, server.applicationInteractionSessionService().Close(ctx, request.PathValue("session_id"), body.ExpectedVersion))
}

func (server *Server) handleListApplicationInteractionTurns(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, applicationSessionTurnListRoute)
	if !server.allowApplicationInteractionSessionDev(writer, trace) {
		return
	}
	values := request.URL.Query()
	if !applicationInteractionSessionQueryAllowed(values, "workspace_id", "application_id") {
		ctx, _, _ := applicationInteractionContextFromRequest(request, trace, values.Get("workspace_id"), values.Get("application_id"), "turn-list", "application_sessions:read")
		writeApplicationInteractionTurnListResult(writer, http.StatusBadRequest, trace, ctx, request.PathValue("session_id"), []ApplicationInteractionTurn{}, ApplicationInteractionFailurePayloadInvalid)
		return
	}
	ctx, failure, status := applicationInteractionContextFromRequest(request, trace, values.Get("workspace_id"), values.Get("application_id"), "turn-list", "application_sessions:read")
	if failure != "" {
		writeApplicationInteractionTurnListResult(writer, status, trace, ctx, request.PathValue("session_id"), []ApplicationInteractionTurn{}, failure)
		return
	}
	turns, failure := server.applicationInteractionSessionService().ListTurns(ctx, request.PathValue("session_id"))
	writeApplicationInteractionTurnListResult(writer, http.StatusOK, trace, ctx, request.PathValue("session_id"), turns, failure)
}

func (server *Server) applicationInteractionSessionService() applicationInteractionSessionService {
	return newApplicationInteractionSessionService(server.applicationInteractionSessionRepository, server.applicationInteractionAuthorityResolver())
}

func (server *Server) applicationInteractionAuthorityResolver() exactApplicationInteractionAuthorityResolver {
	return newExactApplicationInteractionAuthorityResolver(server.applicationCatalogRepository, server.workflowDefinitionReleaseRepository, server.workflowRAGAppRuntimeRepository, server.workflowRAGApplicationAuthorityResolver())
}

func (server *Server) allowApplicationInteractionSessionDev(writer http.ResponseWriter, trace requestTrace) bool {
	if server.config.ApplicationSessionDevEnabled {
		return true
	}
	server.writePlatformError(writer, trace, "APPLICATION_SESSION_DEV_DISABLED", "application interaction sessions require explicit development opt-in")
	return false
}

func applicationInteractionContextFromRequest(request *http.Request, trace requestTrace, workspaceID, applicationID, suffix, requiredScope string) (ApplicationInteractionContext, string, int) {
	auth, failure, status := authorizeControlPlaneReadRequest(request, "", requiredScope)
	ctx := ApplicationInteractionContext{RequestContext: request.Context(), RequestID: trace.requestID, TenantRef: strings.TrimSpace(auth.TenantBinding), WorkspaceID: strings.TrimSpace(workspaceID), ApplicationID: strings.TrimSpace(applicationID), ActorRef: strings.TrimSpace(auth.SubjectBinding), OwnerSubjectRef: strings.TrimSpace(auth.SubjectBinding), AuditRef: "audit_" + trace.requestID + "_application-session-" + suffix}
	if failure != "" {
		return ctx, ApplicationInteractionFailureScopeDenied, status
	}
	if validateApplicationInteractionContext(ctx) != nil {
		return ctx, ApplicationInteractionFailureScopeDenied, http.StatusForbidden
	}
	return ctx, "", http.StatusOK
}

func applicationInteractionSessionQueryAllowed(values map[string][]string, allowed ...string) bool {
	accepted := make(map[string]bool, len(allowed))
	for _, key := range allowed {
		accepted[key] = true
	}
	for key := range values {
		if !accepted[key] {
			return false
		}
	}
	return true
}

func writeApplicationInteractionSessionResult(writer http.ResponseWriter, status int, trace requestTrace, ctx ApplicationInteractionContext, result ApplicationInteractionSessionResult) {
	writeObservedJSON(writer, status, trace, applicationInteractionSessionEnvelope{RequestID: trace.requestID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, Session: result.Session, FailureCode: optionalApplicationDraftFailure(result.FailureCode), CurrentRecordVersion: result.CurrentRecordVersion, CurrentState: result.CurrentState, IdempotentReplay: result.IdempotentReplay, AuditRef: ctx.AuditRef})
}

func writeApplicationInteractionSessionListResult(writer http.ResponseWriter, status int, trace requestTrace, ctx ApplicationInteractionContext, result ApplicationInteractionSessionListResult) {
	if result.Sessions == nil {
		result.Sessions = []ApplicationInteractionSession{}
	}
	writeObservedJSON(writer, status, trace, applicationInteractionSessionListEnvelope{RequestID: trace.requestID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, Items: result.Sessions, NextCursor: result.NextCursor, FailureCode: optionalApplicationDraftFailure(result.FailureCode), AuditRef: ctx.AuditRef})
}

func writeApplicationInteractionTurnListResult(writer http.ResponseWriter, status int, trace requestTrace, ctx ApplicationInteractionContext, sessionID string, turns []ApplicationInteractionTurn, failure string) {
	if turns == nil {
		turns = []ApplicationInteractionTurn{}
	}
	writeObservedJSON(writer, status, trace, applicationInteractionTurnListEnvelope{RequestID: trace.requestID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, SessionID: strings.TrimSpace(sessionID), Items: turns, FailureCode: optionalApplicationDraftFailure(failure), AuditRef: ctx.AuditRef})
}
