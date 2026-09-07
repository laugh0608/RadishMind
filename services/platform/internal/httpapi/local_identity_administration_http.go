package httpapi

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	localIdentityAdminMemberListPath       = "/v1/admin/local-identity/workspaces/{workspace_id}/members"
	localIdentityAdminMemberReadPath       = "/v1/admin/local-identity/workspaces/{workspace_id}/members/{user_id}"
	localIdentityAdminRoleCatalogPath      = "/v1/admin/local-identity/role-catalog"
	localIdentityAdminMembershipCreatePath = "/v1/admin/local-identity/workspaces/{workspace_id}/memberships"
	localIdentityAdminMembershipRevokePath = "/v1/admin/local-identity/workspaces/{workspace_id}/memberships/{membership_id}/revoke"
	localIdentityAdminRoleAssignPath       = "/v1/admin/local-identity/workspaces/{workspace_id}/role-assignments"
	localIdentityAdminRoleRevokePath       = "/v1/admin/local-identity/workspaces/{workspace_id}/role-assignments/{assignment_id}/revoke"
)

type localIdentityAdminMembershipCreateRequest struct {
	UserID    string     `json:"user_id"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	Confirmed bool       `json:"confirmed"`
}

type localIdentityAdminMembershipRevokeRequest struct {
	ExpectedRecordVersion int  `json:"expected_record_version"`
	Confirmed             bool `json:"confirmed"`
}

type localIdentityAdminRoleAssignRequest struct {
	UserID                       string     `json:"user_id"`
	RoleKey                      string     `json:"role_key"`
	ExpectedCatalogVersion       string     `json:"expected_catalog_version"`
	ExpectedRoleDefinitionDigest string     `json:"expected_role_definition_digest"`
	ExpiresAt                    *time.Time `json:"expires_at,omitempty"`
	Confirmed                    bool       `json:"confirmed"`
}

type localIdentityAdminRoleRevokeRequest struct {
	ExpectedRecordVersion int  `json:"expected_record_version"`
	Confirmed             bool `json:"confirmed"`
}

type localIdentityAdminMemberListResponse struct {
	RequestID   string                                `json:"request_id"`
	TenantRef   string                                `json:"tenant_ref"`
	WorkspaceID string                                `json:"workspace_id"`
	Members     []LocalIdentityWorkspaceMemberSummary `json:"members"`
	NextCursor  string                                `json:"next_cursor,omitempty"`
}

type localIdentityAdminMemberReadResponse struct {
	RequestID   string                             `json:"request_id"`
	TenantRef   string                             `json:"tenant_ref"`
	WorkspaceID string                             `json:"workspace_id"`
	Member      LocalIdentityWorkspaceMemberDetail `json:"member"`
}

type localIdentityAdminRoleCatalogResponse struct {
	RequestID   string                   `json:"request_id"`
	TenantRef   string                   `json:"tenant_ref"`
	WorkspaceID string                   `json:"workspace_id"`
	Catalog     LocalIdentityRoleCatalog `json:"catalog"`
}

type localIdentityAdminMembershipResponse struct {
	RequestID              string                                     `json:"request_id"`
	TenantRef              string                                     `json:"tenant_ref"`
	WorkspaceID            string                                     `json:"workspace_id"`
	Membership             LocalIdentityWorkspaceMembershipView       `json:"membership"`
	RevokedRoleAssignments []LocalIdentityWorkspaceRoleAssignmentView `json:"revoked_role_assignments,omitempty"`
}

type localIdentityAdminRoleAssignmentResponse struct {
	RequestID      string                                   `json:"request_id"`
	TenantRef      string                                   `json:"tenant_ref"`
	WorkspaceID    string                                   `json:"workspace_id"`
	RoleAssignment LocalIdentityWorkspaceRoleAssignmentView `json:"role_assignment"`
}

func registerLocalIdentityAdministrationHTTPRoutes(mux *http.ServeMux, server *Server) {
	mux.HandleFunc("GET "+localIdentityAdminMemberListPath, server.handleLocalIdentityAdminMemberList)
	mux.HandleFunc("GET "+localIdentityAdminMemberReadPath, server.handleLocalIdentityAdminMemberRead)
	mux.HandleFunc("GET "+localIdentityAdminRoleCatalogPath, server.handleLocalIdentityAdminRoleCatalog)
	mux.HandleFunc("POST "+localIdentityAdminMembershipCreatePath, server.handleLocalIdentityAdminMembershipCreate)
	mux.HandleFunc("POST "+localIdentityAdminMembershipRevokePath, server.handleLocalIdentityAdminMembershipRevoke)
	mux.HandleFunc("POST "+localIdentityAdminRoleAssignPath, server.handleLocalIdentityAdminRoleAssign)
	mux.HandleFunc("POST "+localIdentityAdminRoleRevokePath, server.handleLocalIdentityAdminRoleRevoke)
}

func (server *Server) handleLocalIdentityAdminMemberList(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, localIdentityAdminMemberListPath)
	_, administration, actor, ok := server.requireLocalIdentityAdministrationRequest(writer, request, trace)
	if !ok || !server.preAuthorizeLocalIdentityAdministration(writer, request, trace, administration, actor, localIdentityPermissionMembersRead) {
		return
	}
	query, err := localIdentityAdminMemberListQuery(request, actor.TenantRef, actor.WorkspaceID)
	if err != nil {
		server.writeLocalIdentityAdministrationError(writer, trace, err)
		return
	}
	page, err := administration.ListWorkspaceMembers(request.Context(), actor, query)
	if err != nil {
		server.writeLocalIdentityAdministrationError(writer, trace, err)
		return
	}
	writeObservedJSON(writer, http.StatusOK, trace, localIdentityAdminMemberListResponse{
		RequestID: trace.requestID, TenantRef: actor.TenantRef, WorkspaceID: actor.WorkspaceID,
		Members: page.Members, NextCursor: page.NextCursor,
	})
}

func (server *Server) handleLocalIdentityAdminMemberRead(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, localIdentityAdminMemberReadPath)
	_, administration, actor, ok := server.requireLocalIdentityAdministrationRequest(writer, request, trace)
	if !ok || !server.preAuthorizeLocalIdentityAdministration(writer, request, trace, administration, actor, localIdentityPermissionMembersRead) {
		return
	}
	if request.URL.RawQuery != "" {
		server.writeLocalIdentityError(writer, trace, localIdentityPayloadInvalid)
		return
	}
	detail, err := administration.ReadWorkspaceMember(
		request.Context(), actor, actor.TenantRef, actor.WorkspaceID, request.PathValue("user_id"),
	)
	if err != nil {
		server.writeLocalIdentityAdministrationError(writer, trace, err)
		return
	}
	writeObservedJSON(writer, http.StatusOK, trace, localIdentityAdminMemberReadResponse{
		RequestID: trace.requestID, TenantRef: actor.TenantRef, WorkspaceID: actor.WorkspaceID, Member: detail,
	})
}

func (server *Server) handleLocalIdentityAdminRoleCatalog(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, localIdentityAdminRoleCatalogPath)
	_, administration, actor, ok := server.requireLocalIdentityAdministrationRequest(writer, request, trace)
	if !ok || !server.preAuthorizeLocalIdentityAdministration(writer, request, trace, administration, actor, localIdentityPermissionRolesRead) {
		return
	}
	if request.URL.RawQuery != "" {
		server.writeLocalIdentityError(writer, trace, localIdentityPayloadInvalid)
		return
	}
	catalog, err := administration.ReadRoleCatalog(request.Context(), actor)
	if err != nil {
		server.writeLocalIdentityAdministrationError(writer, trace, err)
		return
	}
	writeObservedJSON(writer, http.StatusOK, trace, localIdentityAdminRoleCatalogResponse{
		RequestID: trace.requestID, TenantRef: actor.TenantRef, WorkspaceID: actor.WorkspaceID, Catalog: catalog,
	})
}

func (server *Server) handleLocalIdentityAdminMembershipCreate(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, localIdentityAdminMembershipCreatePath)
	_, administration, actor, ok := server.requireLocalIdentityAdministrationMutation(
		writer, request, trace, localIdentityPermissionMembershipsWrite,
	)
	if !ok {
		return
	}
	var body localIdentityAdminMembershipCreateRequest
	if !server.decodeLocalIdentityAdministrationBody(writer, request, trace, &body) {
		return
	}
	if !body.Confirmed {
		server.writeLocalIdentityError(writer, trace, localIdentityPayloadInvalid)
		return
	}
	membership, err := administration.CreateWorkspaceMembership(request.Context(), actor, LocalIdentityCreateWorkspaceMembershipInput{
		TenantRef: actor.TenantRef, WorkspaceID: actor.WorkspaceID, UserID: body.UserID,
		ExpiresAt: body.ExpiresAt, AuditRef: localIdentityAdministrationAuditRef(trace, "membership-create"),
	})
	if err != nil {
		server.writeLocalIdentityAdministrationError(writer, trace, err)
		return
	}
	writeObservedJSON(writer, http.StatusCreated, trace, localIdentityAdminMembershipResponse{
		RequestID: trace.requestID, TenantRef: actor.TenantRef, WorkspaceID: actor.WorkspaceID,
		Membership: localIdentityMembershipView(membership, true),
	})
}

func (server *Server) handleLocalIdentityAdminMembershipRevoke(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, localIdentityAdminMembershipRevokePath)
	_, administration, actor, ok := server.requireLocalIdentityAdministrationMutation(
		writer, request, trace, localIdentityPermissionMembershipsWrite,
	)
	if !ok {
		return
	}
	var body localIdentityAdminMembershipRevokeRequest
	if !server.decodeLocalIdentityAdministrationBody(writer, request, trace, &body) {
		return
	}
	revocation, err := administration.RevokeWorkspaceMembership(request.Context(), actor, LocalIdentityRevokeWorkspaceMembershipInput{
		TenantRef: actor.TenantRef, WorkspaceID: actor.WorkspaceID,
		MembershipID: request.PathValue("membership_id"), ExpectedVersion: body.ExpectedRecordVersion,
		Confirmed: body.Confirmed, AuditRef: localIdentityAdministrationAuditRef(trace, "membership-revoke"),
	})
	if err != nil {
		server.writeLocalIdentityAdministrationError(writer, trace, err)
		return
	}
	now := administration.currentTime()
	revokedAssignments := make([]LocalIdentityWorkspaceRoleAssignmentView, 0, len(revocation.RevokedRoleAssignments))
	for _, assignment := range revocation.RevokedRoleAssignments {
		revokedAssignments = append(revokedAssignments, localIdentityRoleAssignmentView(assignment, false, now))
	}
	writeObservedJSON(writer, http.StatusOK, trace, localIdentityAdminMembershipResponse{
		RequestID: trace.requestID, TenantRef: actor.TenantRef, WorkspaceID: actor.WorkspaceID,
		Membership: localIdentityMembershipView(revocation.Membership, false), RevokedRoleAssignments: revokedAssignments,
	})
}

func (server *Server) handleLocalIdentityAdminRoleAssign(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, localIdentityAdminRoleAssignPath)
	_, administration, actor, ok := server.requireLocalIdentityAdministrationMutation(
		writer, request, trace, localIdentityPermissionRolesAssign,
	)
	if !ok {
		return
	}
	var body localIdentityAdminRoleAssignRequest
	if !server.decodeLocalIdentityAdministrationBody(writer, request, trace, &body) {
		return
	}
	if !body.Confirmed {
		server.writeLocalIdentityError(writer, trace, localIdentityPayloadInvalid)
		return
	}
	assignment, err := administration.AssignWorkspaceRole(request.Context(), actor, LocalIdentityAssignWorkspaceRoleInput{
		TenantRef: actor.TenantRef, WorkspaceID: actor.WorkspaceID, UserID: body.UserID, RoleKey: body.RoleKey,
		ExpectedCatalogVersion: body.ExpectedCatalogVersion, ExpectedRoleDefinitionDigest: body.ExpectedRoleDefinitionDigest,
		ExpiresAt: body.ExpiresAt, AuditRef: localIdentityAdministrationAuditRef(trace, "role-assign"),
	})
	if err != nil {
		server.writeLocalIdentityAdministrationError(writer, trace, err)
		return
	}
	now := administration.currentTime()
	writeObservedJSON(writer, http.StatusCreated, trace, localIdentityAdminRoleAssignmentResponse{
		RequestID: trace.requestID, TenantRef: actor.TenantRef, WorkspaceID: actor.WorkspaceID,
		RoleAssignment: localIdentityRoleAssignmentView(assignment, true, now),
	})
}

func (server *Server) handleLocalIdentityAdminRoleRevoke(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, localIdentityAdminRoleRevokePath)
	_, administration, actor, ok := server.requireLocalIdentityAdministrationMutation(
		writer, request, trace, localIdentityPermissionRolesAssign,
	)
	if !ok {
		return
	}
	var body localIdentityAdminRoleRevokeRequest
	if !server.decodeLocalIdentityAdministrationBody(writer, request, trace, &body) {
		return
	}
	assignment, err := administration.RevokeWorkspaceRole(request.Context(), actor, LocalIdentityRevokeWorkspaceRoleInput{
		TenantRef: actor.TenantRef, WorkspaceID: actor.WorkspaceID,
		AssignmentID: request.PathValue("assignment_id"), ExpectedVersion: body.ExpectedRecordVersion,
		Confirmed: body.Confirmed, AuditRef: localIdentityAdministrationAuditRef(trace, "role-revoke"),
	})
	if err != nil {
		server.writeLocalIdentityAdministrationError(writer, trace, err)
		return
	}
	writeObservedJSON(writer, http.StatusOK, trace, localIdentityAdminRoleAssignmentResponse{
		RequestID: trace.requestID, TenantRef: actor.TenantRef, WorkspaceID: actor.WorkspaceID,
		RoleAssignment: localIdentityRoleAssignmentView(assignment, false, administration.currentTime()),
	})
}

func (server *Server) requireLocalIdentityAdministrationMutation(
	writer http.ResponseWriter,
	request *http.Request,
	trace requestTrace,
	permission string,
) (*localIdentityHTTPService, *localIdentityAdministrationService, LocalIdentityAdministrationActor, bool) {
	identity, administration, actor, ok := server.requireLocalIdentityAdministrationRequest(writer, request, trace)
	if !ok || !server.preAuthorizeLocalIdentityAdministration(writer, request, trace, administration, actor, permission) {
		return nil, nil, LocalIdentityAdministrationActor{}, false
	}
	if request.URL.RawQuery != "" {
		server.writeLocalIdentityError(writer, trace, localIdentityPayloadInvalid)
		return nil, nil, LocalIdentityAdministrationActor{}, false
	}
	requestSession, _ := requireLocalIdentityRequestSession(request)
	if !identity.requireAuthenticatedWriteRequest(
		writer, request, trace, requestSession.csrfToken, requestSession.csrfCookieValid,
	) {
		return nil, nil, LocalIdentityAdministrationActor{}, false
	}
	return identity, administration, actor, true
}

func (server *Server) requireLocalIdentityAdministrationRequest(
	writer http.ResponseWriter,
	request *http.Request,
	trace requestTrace,
) (*localIdentityHTTPService, *localIdentityAdministrationService, LocalIdentityAdministrationActor, bool) {
	identity, ok := server.requireLocalIdentityHTTP(writer, trace)
	if !ok {
		return nil, nil, LocalIdentityAdministrationActor{}, false
	}
	requestSession, ok := requireLocalIdentityRequestSession(request)
	if !ok {
		server.writeLocalIdentityError(writer, trace, localIdentityAuthenticationRequired)
		return nil, nil, LocalIdentityAdministrationActor{}, false
	}
	auth, ok := request.Context().Value(controlPlaneReadAuthContextKey{}).(controlPlaneReadAuthContext)
	if !ok || auth.AuthMode != localIdentityAuthMode || auth.FailureCode != "" || auth.VerifiedIdentity == nil ||
		auth.SubjectBinding != "user:"+requestSession.userID || !auth.ResourceBinding.TenantVerified ||
		strings.TrimSpace(auth.TenantBinding) == "" || auth.TenantBinding != auth.ResourceBinding.TenantRef {
		server.writeLocalIdentityAdministrationError(writer, trace, errLocalIdentityAdminScopeMismatch)
		return nil, nil, LocalIdentityAdministrationActor{}, false
	}
	workspaceValues := request.Header.Values(activeWorkspaceHeader)
	workspaceID := ""
	if len(workspaceValues) == 1 {
		workspaceID = strings.TrimSpace(workspaceValues[0])
	}
	pathWorkspaceID := strings.TrimSpace(request.PathValue("workspace_id"))
	if len(workspaceValues) != 1 || !validControlPlaneReadAuthReference(workspaceID, false) ||
		(pathWorkspaceID != "" && pathWorkspaceID != workspaceID) {
		server.writeLocalIdentityAdministrationError(writer, trace, errLocalIdentityAdminScopeMismatch)
		return nil, nil, LocalIdentityAdministrationActor{}, false
	}
	administration := server.localIdentityAdministrationService
	if administration == nil || administration.repository == nil {
		server.writeLocalIdentityAdministrationError(writer, trace, errLocalIdentityAdminUnavailable)
		return nil, nil, LocalIdentityAdministrationActor{}, false
	}
	return identity, administration, LocalIdentityAdministrationActor{
		UserID: requestSession.userID, TenantRef: auth.TenantBinding, WorkspaceID: workspaceID,
		AuthenticatedAt: requestSession.lastVerifiedAt,
	}, true
}

func (server *Server) preAuthorizeLocalIdentityAdministration(
	writer http.ResponseWriter,
	request *http.Request,
	trace requestTrace,
	administration *localIdentityAdministrationService,
	actor LocalIdentityAdministrationActor,
	permission string,
) bool {
	if err := administration.authorize(request.Context(), actor, actor.TenantRef, actor.WorkspaceID, false, permission); err != nil {
		server.writeLocalIdentityAdministrationError(writer, trace, err)
		return false
	}
	return true
}

func (server *Server) decodeLocalIdentityAdministrationBody(
	writer http.ResponseWriter,
	request *http.Request,
	trace requestTrace,
	target any,
) bool {
	return server.decodeJSONRequestBody(writer, request, trace, target, jsonRequestBodyOptions{
		maxBytes: 16 << 10, rejectUnknownFields: true, rejectDuplicateFields: true,
	})
}

func localIdentityAdminMemberListQuery(
	request *http.Request,
	tenantRef string,
	workspaceID string,
) (LocalIdentityWorkspaceMemberListQuery, error) {
	values, err := url.ParseQuery(request.URL.RawQuery)
	if err != nil {
		return LocalIdentityWorkspaceMemberListQuery{}, errLocalIdentityMemberCursorInvalid
	}
	for key, entries := range values {
		if (key != "membership_state" && key != "limit" && key != "cursor") || len(entries) != 1 {
			return LocalIdentityWorkspaceMemberListQuery{}, errLocalIdentityMemberCursorInvalid
		}
		if strings.TrimSpace(entries[0]) == "" {
			return LocalIdentityWorkspaceMemberListQuery{}, errLocalIdentityMemberCursorInvalid
		}
	}
	query := LocalIdentityWorkspaceMemberListQuery{
		TenantRef: tenantRef, WorkspaceID: workspaceID,
		MembershipState: strings.TrimSpace(values.Get("membership_state")), Cursor: strings.TrimSpace(values.Get("cursor")),
	}
	if rawLimit := strings.TrimSpace(values.Get("limit")); rawLimit != "" {
		query.Limit, err = strconv.Atoi(rawLimit)
		if err != nil || query.Limit < 1 || query.Limit > localIdentityWorkspaceMemberMaximumLimit {
			return LocalIdentityWorkspaceMemberListQuery{}, errLocalIdentityMemberCursorInvalid
		}
	}
	return query, nil
}

func localIdentityAdministrationAuditRef(trace requestTrace, suffix string) string {
	suffix = strings.TrimSpace(suffix)
	return "audit:local-identity:" + suffix + ":" + localIdentityDigest(
		"local_identity_administration_http_v1", trace.requestID, suffix,
	)
}

func (server *Server) writeLocalIdentityAdministrationError(writer http.ResponseWriter, trace requestTrace, err error) {
	code := localIdentityRepositoryError(err)
	if code == LocalIdentityFailureContractMismatch {
		code = localIdentityPayloadInvalid
	}
	status, errorType, message := localIdentityAdministrationHTTPErrorDefinition(code)
	writeTraceHeaders(writer, trace)
	writeJSON(writer, status, errorDocument{Error: errorBody{
		Message: message, Type: errorType, Code: code, RequestID: trace.requestID,
		Route: trace.route, FailureBoundary: "local_identity_administration",
		Metadata: map[string]any{"recovery": localIdentityAdministrationRecovery(code)},
	}})
	logRequestTrace(trace, status, code, "local_identity_administration")
}

func localIdentityAdministrationRecovery(code string) string {
	switch code {
	case LocalIdentityFailureAdminScopeMismatch, LocalIdentityFailureMembershipDenied, LocalIdentityFailurePermissionDenied:
		return "refresh_active_workspace_authorization"
	case LocalIdentityFailureMemberUnavailable:
		return "refresh_member_directory"
	case LocalIdentityFailureMemberCursorInvalid:
		return "restart_member_list"
	case LocalIdentityFailureRoleCatalogMismatch:
		return "refresh_role_catalog"
	case LocalIdentityFailureMembershipConflict, LocalIdentityFailureRoleAssignmentConflict:
		return "refresh_member_detail"
	case LocalIdentityFailureSelfMembershipRevoke, LocalIdentityFailureLastAdminRemoval:
		return "retain_current_administrator"
	case LocalIdentityFailureRecentAuthentication:
		return "reauthenticate"
	case localIdentityPayloadInvalid:
		return "correct_request"
	default:
		return "retry_after_service_recovery"
	}
}

func localIdentityAdministrationHTTPErrorDefinition(code string) (int, string, string) {
	switch code {
	case LocalIdentityFailureAdminScopeMismatch:
		return http.StatusForbidden, "permission_error", "the selected workspace is outside the active administration scope"
	case LocalIdentityFailureMemberUnavailable:
		return http.StatusNotFound, "invalid_request_error", "the workspace member is unavailable"
	case LocalIdentityFailureMemberCursorInvalid:
		return http.StatusBadRequest, "invalid_request_error", "the workspace member cursor or filter is invalid"
	case LocalIdentityFailureRoleCatalogMismatch:
		return http.StatusConflict, "invalid_request_error", "the role catalog changed before this request"
	case LocalIdentityFailureMembershipConflict:
		return http.StatusConflict, "invalid_request_error", "the workspace membership changed before this request"
	case LocalIdentityFailureRoleAssignmentConflict:
		return http.StatusConflict, "invalid_request_error", "the role assignment changed before this request"
	case LocalIdentityFailureSelfMembershipRevoke:
		return http.StatusConflict, "permission_error", "the current administrator cannot revoke their own workspace membership"
	case LocalIdentityFailureLastAdminRemoval:
		return http.StatusConflict, "permission_error", "the workspace must retain an effective local identity administrator"
	case LocalIdentityFailureRecentAuthentication:
		return http.StatusUnauthorized, "authentication_error", "local identity administration requires recent authentication"
	case LocalIdentityFailureMembershipDenied:
		return http.StatusForbidden, "permission_error", "workspace membership is required"
	case LocalIdentityFailurePermissionDenied:
		return http.StatusForbidden, "permission_error", "workspace permission is required"
	case localIdentityPayloadInvalid:
		return localIdentityHTTPErrorDefinition(localIdentityPayloadInvalid)
	case LocalIdentityFailureAdminUnavailable:
		fallthrough
	default:
		return http.StatusServiceUnavailable, "service_unavailable_error", "local identity administration is unavailable"
	}
}
