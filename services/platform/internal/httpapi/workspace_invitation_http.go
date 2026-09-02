package httpapi

import (
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	workspaceInvitationAdminListRoute    = "/v1/admin/local-identity/workspaces/{workspace_id}/invitations"
	workspaceInvitationAdminCreateRoute  = "/v1/admin/local-identity/workspaces/{workspace_id}/invitations"
	workspaceInvitationAdminRevokeRoute  = "/v1/admin/local-identity/workspaces/{workspace_id}/invitations/{invitation_id}/revoke"
	workspaceInvitationClaimPreviewRoute = "/v1/auth/workspace-invitations/preview"
	workspaceInvitationClaimRoute        = "/v1/auth/workspace-invitations/claim"
)

type workspaceInvitationAdminCreateRequest struct {
	RoleKey                      string `json:"role_key"`
	ExpectedCatalogVersion       string `json:"expected_catalog_version"`
	ExpectedRoleDefinitionDigest string `json:"expected_role_definition_digest"`
	TTLPolicy                    string `json:"ttl_policy"`
	Confirmed                    bool   `json:"confirmed"`
}

type workspaceInvitationAdminRevokeRequest struct {
	ExpectedRecordVersion int  `json:"expected_record_version"`
	Confirmed             bool `json:"confirmed"`
}

type workspaceInvitationPreviewRequest struct {
	InvitationCode string `json:"invitation_code"`
}

type workspaceInvitationClaimRequest struct {
	InvitationCode        string `json:"invitation_code"`
	ExpectedRecordVersion int    `json:"expected_record_version"`
	Confirmed             bool   `json:"confirmed"`
}

type workspaceInvitationListHTTPResponse struct {
	RequestID   string `json:"request_id"`
	TenantRef   string `json:"tenant_ref"`
	WorkspaceID string `json:"workspace_id"`
	WorkspaceInvitationPage
}

type workspaceInvitationCreationHTTPResponse struct {
	RequestID string `json:"request_id"`
	WorkspaceInvitationCreation
}

type workspaceInvitationPreviewHTTPResponse struct {
	RequestID string `json:"request_id"`
	WorkspaceInvitationPreview
}

type workspaceInvitationMutationHTTPResponse struct {
	RequestID string `json:"request_id"`
	WorkspaceInvitationMutation
}

func registerWorkspaceInvitationHTTPRoutes(mux *http.ServeMux, server *Server) {
	mux.HandleFunc("GET "+workspaceInvitationAdminListRoute, server.handleWorkspaceInvitationAdminList)
	mux.HandleFunc("POST "+workspaceInvitationAdminCreateRoute, server.handleWorkspaceInvitationAdminCreate)
	mux.HandleFunc("POST "+workspaceInvitationAdminRevokeRoute, server.handleWorkspaceInvitationAdminRevoke)
	mux.HandleFunc("POST "+workspaceInvitationClaimPreviewRoute, server.handleWorkspaceInvitationPreview)
	mux.HandleFunc("POST "+workspaceInvitationClaimRoute, server.handleWorkspaceInvitationClaim)
}

func (server *Server) handleWorkspaceInvitationAdminList(writer http.ResponseWriter, request *http.Request) {
	trace := newWorkspaceInvitationRequestTrace(request, workspaceInvitationAdminListRoute)
	service, actor, ok := server.requireWorkspaceInvitationAdministrator(
		writer, request, trace, false, localIdentityPermissionMembersRead, localIdentityPermissionRolesRead,
	)
	if !ok {
		return
	}
	query, err := workspaceInvitationAdminListQuery(request, actor.TenantRef, actor.WorkspaceID)
	if err != nil {
		server.writeWorkspaceInvitationError(writer, trace, err, true)
		return
	}
	page, err := service.List(request.Context(), actor, query)
	if err != nil {
		server.writeWorkspaceInvitationError(writer, trace, err, true)
		return
	}
	writeObservedJSON(writer, http.StatusOK, trace, workspaceInvitationListHTTPResponse{
		RequestID: trace.requestID, TenantRef: actor.TenantRef, WorkspaceID: actor.WorkspaceID,
		WorkspaceInvitationPage: page,
	})
}

func (server *Server) handleWorkspaceInvitationAdminCreate(writer http.ResponseWriter, request *http.Request) {
	trace := newWorkspaceInvitationRequestTrace(request, workspaceInvitationAdminCreateRoute)
	service, actor, ok := server.requireWorkspaceInvitationAdministrator(
		writer, request, trace, true, localIdentityPermissionMembershipsWrite, localIdentityPermissionRolesAssign,
	)
	if !ok {
		return
	}
	var body workspaceInvitationAdminCreateRequest
	if !server.decodeWorkspaceInvitationBody(writer, request, trace, &body) {
		return
	}
	if !body.Confirmed {
		server.writeLocalIdentityError(writer, trace, localIdentityPayloadInvalid)
		return
	}
	requestRef, auditRef := workspaceInvitationHTTPRefs(trace, "admin-create")
	creation, err := service.Create(request.Context(), actor, WorkspaceInvitationCreateInput{
		TenantRef: actor.TenantRef, WorkspaceID: actor.WorkspaceID, RoleKey: body.RoleKey,
		ExpectedCatalogVersion:       body.ExpectedCatalogVersion,
		ExpectedRoleDefinitionDigest: body.ExpectedRoleDefinitionDigest,
		TTLPolicy:                    body.TTLPolicy, Confirmed: body.Confirmed, RequestRef: requestRef, AuditRef: auditRef,
	})
	if err != nil {
		server.writeWorkspaceInvitationError(writer, trace, err, true)
		return
	}
	writer.Header().Set("Cache-Control", "no-store")
	writeObservedJSON(writer, http.StatusCreated, trace, workspaceInvitationCreationHTTPResponse{
		RequestID: trace.requestID, WorkspaceInvitationCreation: creation,
	})
}

func (server *Server) handleWorkspaceInvitationAdminRevoke(writer http.ResponseWriter, request *http.Request) {
	trace := newWorkspaceInvitationRequestTrace(request, workspaceInvitationAdminRevokeRoute)
	service, actor, ok := server.requireWorkspaceInvitationAdministrator(
		writer, request, trace, true, localIdentityPermissionMembershipsWrite, localIdentityPermissionRolesAssign,
	)
	if !ok {
		return
	}
	var body workspaceInvitationAdminRevokeRequest
	if !server.decodeWorkspaceInvitationBody(writer, request, trace, &body) {
		return
	}
	if !body.Confirmed {
		server.writeLocalIdentityError(writer, trace, localIdentityPayloadInvalid)
		return
	}
	requestRef, auditRef := workspaceInvitationHTTPRefs(trace, "admin-revoke")
	mutation, err := service.Revoke(request.Context(), actor, WorkspaceInvitationRevokeInput{
		TenantRef: actor.TenantRef, WorkspaceID: actor.WorkspaceID,
		InvitationID: request.PathValue("invitation_id"), ExpectedVersion: body.ExpectedRecordVersion,
		Confirmed: body.Confirmed, RequestRef: requestRef, AuditRef: auditRef,
	})
	if err != nil {
		server.writeWorkspaceInvitationError(writer, trace, err, true)
		return
	}
	writeObservedJSON(writer, http.StatusOK, trace, workspaceInvitationMutationHTTPResponse{
		RequestID: trace.requestID, WorkspaceInvitationMutation: mutation,
	})
}

func (server *Server) handleWorkspaceInvitationPreview(writer http.ResponseWriter, request *http.Request) {
	trace := newWorkspaceInvitationRequestTrace(request, workspaceInvitationClaimPreviewRoute)
	service, actor, _, ok := server.requireWorkspaceInvitationClaimant(writer, request, trace, false)
	if !ok {
		return
	}
	var body workspaceInvitationPreviewRequest
	if !server.decodeWorkspaceInvitationBody(writer, request, trace, &body) {
		return
	}
	preview, err := service.Preview(request.Context(), actor, body.InvitationCode)
	if err != nil {
		server.writeWorkspaceInvitationError(writer, trace, err, false)
		return
	}
	writer.Header().Set("Cache-Control", "no-store")
	writeObservedJSON(writer, http.StatusOK, trace, workspaceInvitationPreviewHTTPResponse{
		RequestID: trace.requestID, WorkspaceInvitationPreview: preview,
	})
}

func (server *Server) handleWorkspaceInvitationClaim(writer http.ResponseWriter, request *http.Request) {
	trace := newWorkspaceInvitationRequestTrace(request, workspaceInvitationClaimRoute)
	service, actor, requestSession, ok := server.requireWorkspaceInvitationClaimant(writer, request, trace, true)
	if !ok {
		return
	}
	var body workspaceInvitationClaimRequest
	if !server.decodeWorkspaceInvitationBody(writer, request, trace, &body) {
		return
	}
	if !body.Confirmed {
		server.writeLocalIdentityError(writer, trace, localIdentityPayloadInvalid)
		return
	}
	requestRef, auditRef := workspaceInvitationHTTPRefs(trace, "claim")
	mutation, err := service.Claim(request.Context(), actor, WorkspaceInvitationClaimInput{
		InvitationCode: body.InvitationCode, ExpectedVersion: body.ExpectedRecordVersion,
		Confirmed: body.Confirmed, RequestRef: requestRef, AuditRef: auditRef,
	})
	if err != nil {
		server.writeWorkspaceInvitationError(writer, trace, err, false)
		return
	}
	server.localIdentityHTTPService.setCSRFCookie(writer, requestSession.csrfToken, requestSession.expiresAt)
	writer.Header().Set("Cache-Control", "no-store")
	writeObservedJSON(writer, http.StatusOK, trace, workspaceInvitationMutationHTTPResponse{
		RequestID: trace.requestID, WorkspaceInvitationMutation: mutation,
	})
}

func (server *Server) requireWorkspaceInvitationAdministrator(
	writer http.ResponseWriter,
	request *http.Request,
	trace requestTrace,
	mutation bool,
	permissions ...string,
) (*workspaceInvitationService, LocalIdentityAdministrationActor, bool) {
	identity, administration, actor, ok := server.requireLocalIdentityAdministrationRequest(writer, request, trace)
	if !ok {
		return nil, LocalIdentityAdministrationActor{}, false
	}
	if err := administration.authorize(
		request.Context(), actor, actor.TenantRef, actor.WorkspaceID, false, permissions...,
	); err != nil {
		server.writeWorkspaceInvitationError(writer, trace, err, true)
		return nil, LocalIdentityAdministrationActor{}, false
	}
	service := server.workspaceInvitationService
	if service == nil || service.repository == nil {
		server.writeWorkspaceInvitationError(writer, trace, errWorkspaceInvitationAdminUnavailable, true)
		return nil, LocalIdentityAdministrationActor{}, false
	}
	if mutation {
		if request.URL.RawQuery != "" {
			server.writeLocalIdentityError(writer, trace, localIdentityPayloadInvalid)
			return nil, LocalIdentityAdministrationActor{}, false
		}
		requestSession, _ := requireLocalIdentityRequestSession(request)
		if !identity.requireAuthenticatedWriteRequest(
			writer, request, trace, requestSession.csrfToken, requestSession.csrfCookieValid,
		) {
			return nil, LocalIdentityAdministrationActor{}, false
		}
	}
	return service, actor, true
}

func (server *Server) requireWorkspaceInvitationClaimant(
	writer http.ResponseWriter,
	request *http.Request,
	trace requestTrace,
	mutation bool,
) (*workspaceInvitationService, WorkspaceInvitationClaimantActor, localIdentityRequestSession, bool) {
	identity, ok := server.requireLocalIdentityHTTP(writer, trace)
	if !ok {
		return nil, WorkspaceInvitationClaimantActor{}, localIdentityRequestSession{}, false
	}
	requestSession, ok := requireLocalIdentityRequestSession(request)
	if !ok {
		server.writeLocalIdentityError(writer, trace, localIdentityAuthenticationRequired)
		return nil, WorkspaceInvitationClaimantActor{}, localIdentityRequestSession{}, false
	}
	auth, ok := request.Context().Value(controlPlaneReadAuthContextKey{}).(controlPlaneReadAuthContext)
	if !ok || auth.AuthMode != localIdentityAuthMode || auth.FailureCode != "" || auth.VerifiedIdentity == nil ||
		auth.VerifiedIdentity.AuthSource != localIdentityAuthMode || auth.SubjectBinding != "user:"+requestSession.userID ||
		auth.SessionRef != "session:"+requestSession.sessionID || !auth.ResourceBinding.TenantVerified ||
		!validControlPlaneReadAuthReference(strings.TrimSpace(auth.TenantBinding), false) ||
		auth.TenantBinding != auth.ResourceBinding.TenantRef {
		server.writeWorkspaceInvitationError(writer, trace, errWorkspaceInvitationAccountIneligible, false)
		return nil, WorkspaceInvitationClaimantActor{}, localIdentityRequestSession{}, false
	}
	if request.URL.RawQuery != "" {
		server.writeLocalIdentityError(writer, trace, localIdentityPayloadInvalid)
		return nil, WorkspaceInvitationClaimantActor{}, localIdentityRequestSession{}, false
	}
	service := server.workspaceInvitationService
	if service == nil || service.repository == nil {
		server.writeWorkspaceInvitationError(writer, trace, errWorkspaceInvitationStoreUnavailable, false)
		return nil, WorkspaceInvitationClaimantActor{}, localIdentityRequestSession{}, false
	}
	if mutation {
		if !identity.requireAuthenticatedWriteRequest(
			writer, request, trace, requestSession.csrfToken, requestSession.csrfCookieValid,
		) {
			return nil, WorkspaceInvitationClaimantActor{}, localIdentityRequestSession{}, false
		}
	} else if !workspaceInvitationJSONContentType(request) {
		server.writeLocalIdentityError(writer, trace, localIdentityPayloadInvalid)
		return nil, WorkspaceInvitationClaimantActor{}, localIdentityRequestSession{}, false
	}
	return service, WorkspaceInvitationClaimantActor{
		UserID: requestSession.userID, TenantRef: auth.TenantBinding, AuthenticatedAt: requestSession.lastVerifiedAt,
	}, requestSession, true
}

func (server *Server) decodeWorkspaceInvitationBody(
	writer http.ResponseWriter,
	request *http.Request,
	trace requestTrace,
	target any,
) bool {
	return server.decodeJSONRequestBody(writer, request, trace, target, jsonRequestBodyOptions{
		maxBytes: 16 << 10, rejectUnknownFields: true, rejectDuplicateFields: true,
	})
}

func workspaceInvitationJSONContentType(request *http.Request) bool {
	mediaType, _, err := mime.ParseMediaType(request.Header.Get("Content-Type"))
	return err == nil && mediaType == "application/json"
}

func workspaceInvitationAdminListQuery(
	request *http.Request,
	tenantRef string,
	workspaceID string,
) (WorkspaceInvitationListQuery, error) {
	values, err := url.ParseQuery(request.URL.RawQuery)
	if err != nil {
		return WorkspaceInvitationListQuery{}, errWorkspaceInvitationCursorInvalid
	}
	for key, entries := range values {
		if (key != "effective_state" && key != "limit" && key != "cursor") || len(entries) != 1 ||
			strings.TrimSpace(entries[0]) == "" {
			return WorkspaceInvitationListQuery{}, errWorkspaceInvitationCursorInvalid
		}
	}
	query := WorkspaceInvitationListQuery{
		TenantRef: tenantRef, WorkspaceID: workspaceID,
		EffectiveState: strings.TrimSpace(values.Get("effective_state")), Cursor: strings.TrimSpace(values.Get("cursor")),
	}
	if rawLimit := strings.TrimSpace(values.Get("limit")); rawLimit != "" {
		query.Limit, err = strconv.Atoi(rawLimit)
		if err != nil || query.Limit < 1 || query.Limit > workspaceInvitationMaximumListLimit {
			return WorkspaceInvitationListQuery{}, errWorkspaceInvitationCursorInvalid
		}
	}
	return query, nil
}

func workspaceInvitationHTTPRefs(trace requestTrace, action string) (string, string) {
	action = strings.TrimSpace(action)
	digest := localIdentityDigest("workspace_invitation_http_v1", trace.requestID, action)
	return "request:workspace-invitation:" + digest, "audit:workspace-invitation:" + action + ":" + digest
}

func newWorkspaceInvitationRequestTrace(request *http.Request, route string) requestTrace {
	traceRequest := request.Clone(request.Context())
	traceRequest.Header = request.Header.Clone()
	traceRequest.Header.Del("X-Request-Id")
	traceRequest.Header.Del("X-Request-ID")
	traceRequest.Header.Del("OpenAI-Request-ID")
	return newRequestTrace(traceRequest, route)
}

func (server *Server) writeWorkspaceInvitationError(
	writer http.ResponseWriter,
	trace requestTrace,
	err error,
	adminSurface bool,
) {
	code := workspaceInvitationFailureCode(err)
	status, errorType, message := workspaceInvitationHTTPErrorDefinition(code, adminSurface)
	writeTraceHeaders(writer, trace)
	writeJSON(writer, status, errorDocument{Error: errorBody{
		Message: message, Type: errorType, Code: code, RequestID: trace.requestID,
		Route: trace.route, FailureBoundary: workspaceInvitationFailureBoundary(adminSurface),
		Metadata: map[string]any{"recovery": workspaceInvitationRecovery(code)},
	}})
	logRequestTrace(trace, status, code, workspaceInvitationFailureBoundary(adminSurface))
}

func workspaceInvitationFailureBoundary(adminSurface bool) string {
	if adminSurface {
		return "workspace_invitation_administration"
	}
	return "workspace_invitation_claim"
}

func workspaceInvitationRecovery(code string) string {
	switch code {
	case WorkspaceInvitationFailureCursorInvalid:
		return "restart_invitation_list"
	case WorkspaceInvitationFailureRoleIneligible, WorkspaceInvitationFailureRoleCatalogMismatch:
		return "refresh_role_catalog"
	case WorkspaceInvitationFailureVersionConflict:
		return "refresh_invitation_state"
	case WorkspaceInvitationFailureTransitionInvalid:
		return "refresh_invitation_directory"
	case WorkspaceInvitationFailureInvalid:
		return "reenter_invitation_code"
	case WorkspaceInvitationFailureNotClaimable:
		return "discard_terminal_invitation"
	case WorkspaceInvitationFailureAccountIneligible:
		return "select_matching_tenant_or_account"
	case WorkspaceInvitationFailureMembershipConflict:
		return "refresh_workspace_access"
	case LocalIdentityFailureRecentAuthentication:
		return "reauthenticate"
	case LocalIdentityFailureMembershipDenied, LocalIdentityFailurePermissionDenied:
		return "refresh_active_workspace_authorization"
	default:
		return "retry_after_service_recovery"
	}
}

func workspaceInvitationHTTPErrorDefinition(code string, adminSurface bool) (int, string, string) {
	switch code {
	case WorkspaceInvitationFailureCursorInvalid:
		return http.StatusBadRequest, "invalid_request_error", "workspace invitation list query or cursor is invalid"
	case WorkspaceInvitationFailureRoleIneligible:
		return http.StatusBadRequest, "invalid_request_error", "the selected role is not eligible for invitation"
	case WorkspaceInvitationFailureRoleCatalogMismatch:
		return http.StatusConflict, "invalid_request_error", "the role catalog changed before this request"
	case WorkspaceInvitationFailureVersionConflict:
		return http.StatusConflict, "invalid_request_error", "the workspace invitation changed before this request"
	case WorkspaceInvitationFailureTransitionInvalid:
		return http.StatusConflict, "invalid_request_error", "the workspace invitation transition is unavailable"
	case WorkspaceInvitationFailureInvalid:
		return http.StatusBadRequest, "invalid_request_error", "the workspace invitation code is invalid"
	case WorkspaceInvitationFailureNotClaimable:
		return http.StatusConflict, "invalid_request_error", "the workspace invitation is no longer claimable"
	case WorkspaceInvitationFailureAccountIneligible:
		return http.StatusForbidden, "permission_error", "the current local account is not eligible for this invitation"
	case WorkspaceInvitationFailureMembershipConflict:
		return http.StatusConflict, "permission_error", "workspace access already exists or conflicts with this invitation"
	case LocalIdentityFailureRecentAuthentication:
		return http.StatusUnauthorized, "authentication_error", "workspace invitation access requires recent authentication"
	case LocalIdentityFailureMembershipDenied:
		return http.StatusForbidden, "permission_error", "workspace membership is required"
	case LocalIdentityFailurePermissionDenied:
		return http.StatusForbidden, "permission_error", "workspace permission is required"
	case WorkspaceInvitationFailureAdminUnavailable:
		return http.StatusServiceUnavailable, "service_unavailable_error", "workspace invitation administration is unavailable"
	case WorkspaceInvitationFailureStoreUnavailable:
		return http.StatusServiceUnavailable, "service_unavailable_error", "workspace invitation storage is unavailable"
	default:
		if adminSurface {
			return http.StatusServiceUnavailable, "service_unavailable_error", "workspace invitation administration is unavailable"
		}
		return http.StatusServiceUnavailable, "service_unavailable_error", "workspace invitation storage is unavailable"
	}
}
