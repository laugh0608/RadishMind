package httpapi

import (
	"net/http"
	"strconv"
	"strings"
)

const (
	applicationResultArtifactListRoute      = "GET /v1/user-workspace/application-sessions/{session_id}/result-artifacts"
	applicationResultArtifactReadRoute      = "GET /v1/user-workspace/application-sessions/{session_id}/result-artifacts/{artifact_id}"
	applicationResultArtifactArchiveRoute   = "POST /v1/user-workspace/application-sessions/{session_id}/result-artifacts/{artifact_id}/archive"
	applicationResultArtifactUnarchiveRoute = "POST /v1/user-workspace/application-sessions/{session_id}/result-artifacts/{artifact_id}/unarchive"
)

type applicationResultArtifactLifecycleBody struct {
	WorkspaceID              string `json:"workspace_id"`
	ApplicationID            string `json:"application_id"`
	ExpectedLifecycleVersion int    `json:"expected_lifecycle_version"`
}

type applicationResultArtifactEnvelope struct {
	RequestID     string                              `json:"request_id"`
	TenantRef     string                              `json:"tenant_ref"`
	WorkspaceID   string                              `json:"workspace_id"`
	ApplicationID string                              `json:"application_id"`
	SessionID     string                              `json:"session_id"`
	Artifact      *ApplicationResultArtifact          `json:"artifact"`
	Lifecycle     *ApplicationResultArtifactLifecycle `json:"lifecycle"`
	FailureCode   *string                             `json:"failure_code"`
	AuditRef      string                              `json:"audit_ref"`
}

type applicationResultArtifactLifecycleEnvelope struct {
	RequestID               string                                   `json:"request_id"`
	TenantRef               string                                   `json:"tenant_ref"`
	WorkspaceID             string                                   `json:"workspace_id"`
	ApplicationID           string                                   `json:"application_id"`
	SessionID               string                                   `json:"session_id"`
	Lifecycle               *ApplicationResultArtifactLifecycle      `json:"lifecycle"`
	Event                   *ApplicationResultArtifactLifecycleEvent `json:"event"`
	FailureCode             *string                                  `json:"failure_code"`
	CurrentLifecycleVersion int                                      `json:"current_lifecycle_version"`
	CurrentLifecycleState   ApplicationResultArtifactLifecycleState  `json:"current_lifecycle_state"`
	AuditRef                string                                   `json:"audit_ref"`
}

type applicationResultArtifactListEnvelope struct {
	RequestID     string                             `json:"request_id"`
	TenantRef     string                             `json:"tenant_ref"`
	WorkspaceID   string                             `json:"workspace_id"`
	ApplicationID string                             `json:"application_id"`
	SessionID     string                             `json:"session_id"`
	Items         []ApplicationResultArtifactSummary `json:"items"`
	NextCursor    *string                            `json:"next_cursor"`
	FailureCode   *string                            `json:"failure_code"`
	AuditRef      string                             `json:"audit_ref"`
}

func (server *Server) handleListApplicationResultArtifacts(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, applicationResultArtifactListRoute)
	if !server.allowApplicationInteractionSessionDev(writer, trace) {
		return
	}
	writer.Header().Set("Cache-Control", "no-store")
	values := request.URL.Query()
	ctx, failure, status := server.applicationResultArtifactReadContext(request, trace, values.Get("workspace_id"), values.Get("application_id"), "result-artifact-list")
	if failure != "" {
		writeApplicationResultArtifactList(writer, status, trace, ctx, request.PathValue("session_id"), ApplicationResultArtifactListResult{Items: []ApplicationResultArtifactSummary{}, FailureCode: failure})
		return
	}
	if !applicationInteractionSessionQueryAllowed(values, "workspace_id", "application_id", "lifecycle_state", "limit", "cursor") {
		writeApplicationResultArtifactList(writer, http.StatusBadRequest, trace, ctx, request.PathValue("session_id"), ApplicationResultArtifactListResult{Items: []ApplicationResultArtifactSummary{}, FailureCode: ApplicationResultArtifactFailurePayloadInvalid})
		return
	}
	limit := 0
	if raw := strings.TrimSpace(values.Get("limit")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			writeApplicationResultArtifactList(writer, http.StatusBadRequest, trace, ctx, request.PathValue("session_id"), ApplicationResultArtifactListResult{Items: []ApplicationResultArtifactSummary{}, FailureCode: ApplicationResultArtifactFailurePayloadInvalid})
			return
		}
		limit = parsed
	}
	result := server.applicationResultArtifactService().List(ctx, ApplicationResultArtifactListInput{
		SessionID: request.PathValue("session_id"), LifecycleState: ApplicationResultArtifactLifecycleState(strings.TrimSpace(values.Get("lifecycle_state"))), Limit: limit, Cursor: values.Get("cursor"),
	})
	writeApplicationResultArtifactList(writer, http.StatusOK, trace, ctx, request.PathValue("session_id"), result)
}

func (server *Server) handleArchiveApplicationResultArtifact(writer http.ResponseWriter, request *http.Request) {
	server.handleApplicationResultArtifactLifecycleTransition(writer, request, applicationResultArtifactArchiveRoute, ApplicationResultArtifactLifecycleArchived, "result-artifact-archive")
}

func (server *Server) handleUnarchiveApplicationResultArtifact(writer http.ResponseWriter, request *http.Request) {
	server.handleApplicationResultArtifactLifecycleTransition(writer, request, applicationResultArtifactUnarchiveRoute, ApplicationResultArtifactLifecycleActive, "result-artifact-unarchive")
}

func (server *Server) handleApplicationResultArtifactLifecycleTransition(
	writer http.ResponseWriter,
	request *http.Request,
	route string,
	target ApplicationResultArtifactLifecycleState,
	auditSuffix string,
) {
	trace := newRequestTrace(request, route)
	if !server.allowApplicationInteractionSessionDev(writer, trace) {
		return
	}
	auth, failure, status := server.authorizeWorkspaceScopedPermissions(
		request,
		"application_sessions:read",
		"application_result_artifacts:archive",
	)
	ctx := applicationInteractionMutationContext(request, trace, auth, "", auditSuffix)
	if failure != "" {
		writeApplicationResultArtifactLifecycle(writer, status, trace, ctx, request.PathValue("session_id"), applicationResultArtifactLifecycleFailure(failure))
		return
	}
	var body applicationResultArtifactLifecycleBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{
		maxBytes: maxControlJSONRequestBodyBytes, rejectUnknownFields: true,
	}) {
		return
	}
	ctx = applicationInteractionMutationContext(request, trace, auth, strings.TrimSpace(body.ApplicationID), auditSuffix)
	if len(request.URL.Query()) != 0 || strings.TrimSpace(body.WorkspaceID) != auth.ResourceBinding.WorkspaceID ||
		validateApplicationInteractionContext(ctx) != nil {
		writeApplicationResultArtifactLifecycle(writer, http.StatusForbidden, trace, ctx, request.PathValue("session_id"), applicationResultArtifactLifecycleFailure("workspace_binding_mismatch"))
		return
	}
	ctx.WriteEnabled = true
	input := ApplicationResultArtifactLifecycleTransitionInput{
		SessionID: request.PathValue("session_id"), ArtifactID: request.PathValue("artifact_id"),
		ExpectedLifecycleVersion: body.ExpectedLifecycleVersion,
	}
	var result ApplicationResultArtifactLifecycleTransitionResult
	if target == ApplicationResultArtifactLifecycleActive {
		result = server.applicationResultArtifactService().Unarchive(ctx, input)
	} else {
		result = server.applicationResultArtifactService().Archive(ctx, input)
	}
	writeApplicationResultArtifactLifecycle(writer, applicationResultArtifactHTTPStatus(result.FailureCode), trace, ctx, request.PathValue("session_id"), result)
}

func (server *Server) handleReadApplicationResultArtifact(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, applicationResultArtifactReadRoute)
	if !server.allowApplicationInteractionSessionDev(writer, trace) {
		return
	}
	writer.Header().Set("Cache-Control", "no-store")
	values := request.URL.Query()
	ctx, failure, status := server.applicationResultArtifactReadContext(request, trace, values.Get("workspace_id"), values.Get("application_id"), "result-artifact-read")
	if failure != "" {
		writeApplicationResultArtifact(writer, status, trace, ctx, request.PathValue("session_id"), applicationResultArtifactFailure(failure))
		return
	}
	if !applicationInteractionSessionQueryAllowed(values, "workspace_id", "application_id") {
		writeApplicationResultArtifact(writer, http.StatusBadRequest, trace, ctx, request.PathValue("session_id"), applicationResultArtifactFailure(ApplicationResultArtifactFailurePayloadInvalid))
		return
	}
	result := server.applicationResultArtifactService().Read(ctx, request.PathValue("artifact_id"))
	if result.Artifact != nil && result.Artifact.SessionID != strings.TrimSpace(request.PathValue("session_id")) {
		result = applicationResultArtifactFailure(ApplicationResultArtifactFailureNotFound)
	}
	writeApplicationResultArtifact(writer, http.StatusOK, trace, ctx, request.PathValue("session_id"), result)
}

func (server *Server) applicationResultArtifactService() applicationResultArtifactService {
	return newApplicationResultArtifactService(server.applicationResultArtifactRepository, server.effectiveApplicationSessionRepository())
}

func (server *Server) applicationResultArtifactReadContext(
	request *http.Request,
	trace requestTrace,
	workspaceID string,
	applicationID string,
	suffix string,
) (ApplicationInteractionContext, string, int) {
	auth, failure, status := server.authorizeWorkspaceScopedPermissions(request, "application_sessions:read")
	ctx := applicationInteractionMutationContext(request, trace, auth, applicationID, suffix)
	if failure != "" {
		return ctx, failure, status
	}
	if strings.TrimSpace(workspaceID) != auth.ResourceBinding.WorkspaceID || validateApplicationInteractionContext(ctx) != nil {
		return ctx, "workspace_binding_mismatch", http.StatusForbidden
	}
	return ctx, "", http.StatusOK
}

func writeApplicationResultArtifact(
	writer http.ResponseWriter,
	status int,
	trace requestTrace,
	ctx ApplicationInteractionContext,
	sessionID string,
	result ApplicationResultArtifactResult,
) {
	writeObservedJSON(writer, status, trace, applicationResultArtifactEnvelope{
		RequestID: trace.requestID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID,
		ApplicationID: ctx.ApplicationID, SessionID: strings.TrimSpace(sessionID), Artifact: result.Artifact,
		Lifecycle:   result.Lifecycle,
		FailureCode: optionalApplicationDraftFailure(result.FailureCode), AuditRef: ctx.AuditRef,
	})
}

func writeApplicationResultArtifactLifecycle(
	writer http.ResponseWriter,
	status int,
	trace requestTrace,
	ctx ApplicationInteractionContext,
	sessionID string,
	result ApplicationResultArtifactLifecycleTransitionResult,
) {
	writeObservedJSON(writer, status, trace, applicationResultArtifactLifecycleEnvelope{
		RequestID: trace.requestID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID,
		ApplicationID: ctx.ApplicationID, SessionID: strings.TrimSpace(sessionID),
		Lifecycle: result.Lifecycle, Event: result.Event, FailureCode: optionalApplicationDraftFailure(result.FailureCode),
		CurrentLifecycleVersion: result.CurrentLifecycleVersion, CurrentLifecycleState: result.CurrentLifecycleState,
		AuditRef: ctx.AuditRef,
	})
}

func writeApplicationResultArtifactList(
	writer http.ResponseWriter,
	status int,
	trace requestTrace,
	ctx ApplicationInteractionContext,
	sessionID string,
	result ApplicationResultArtifactListResult,
) {
	if result.Items == nil {
		result.Items = []ApplicationResultArtifactSummary{}
	}
	writeObservedJSON(writer, status, trace, applicationResultArtifactListEnvelope{
		RequestID: trace.requestID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID,
		ApplicationID: ctx.ApplicationID, SessionID: strings.TrimSpace(sessionID), Items: result.Items,
		NextCursor: result.NextCursor, FailureCode: optionalApplicationDraftFailure(result.FailureCode), AuditRef: ctx.AuditRef,
	})
}

func applicationResultArtifactHTTPStatus(failureCode string) int {
	switch failureCode {
	case "":
		return http.StatusOK
	case ApplicationResultArtifactFailurePayloadInvalid:
		return http.StatusBadRequest
	case ApplicationResultArtifactFailureNotFound:
		return http.StatusNotFound
	case ApplicationResultArtifactFailureConflict,
		ApplicationResultArtifactFailureLifecycleVersion,
		ApplicationResultArtifactFailureLifecycleState:
		return http.StatusConflict
	case ApplicationResultArtifactFailureStoreUnavailable:
		return http.StatusServiceUnavailable
	case ApplicationResultArtifactFailureStoreContract:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
