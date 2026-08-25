package httpapi

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	localIdentitySelfServiceSessionListRoute         = "/v1/auth/sessions"
	localIdentitySelfServiceSessionRevokeRoute       = "/v1/auth/sessions/{session_id}/revoke"
	localIdentitySelfServiceSessionRevokeOthersRoute = "/v1/auth/sessions/revoke-others"
	localIdentitySelfServiceCredentialRotateRoute    = "/v1/auth/local/credential/rotate"
)

type localIdentitySelfServiceSessionRevokeRequest struct {
	ExpectedRecordVersion int  `json:"expected_record_version"`
	Confirmed             bool `json:"confirmed"`
}

type localIdentitySelfServiceSessionRevokeOthersRequest struct {
	Confirmed bool `json:"confirmed"`
}

type localIdentitySelfServiceCredentialRotateRequest struct {
	CurrentPassword        string `json:"current_password"`
	NewPassword            string `json:"new_password"`
	SessionImpactConfirmed bool   `json:"session_impact_confirmed"`
}

func registerLocalIdentitySelfServiceSecurityHTTPRoutes(mux *http.ServeMux, server *Server) {
	mux.HandleFunc("GET "+localIdentitySelfServiceSessionListRoute, server.handleLocalIdentitySelfServiceSessionList)
	mux.HandleFunc("POST "+localIdentitySelfServiceSessionRevokeRoute, server.handleLocalIdentitySelfServiceSessionRevoke)
	mux.HandleFunc("POST "+localIdentitySelfServiceSessionRevokeOthersRoute, server.handleLocalIdentitySelfServiceSessionRevokeOthers)
	mux.HandleFunc("POST "+localIdentitySelfServiceCredentialRotateRoute, server.handleLocalIdentitySelfServiceCredentialRotate)
}

func (server *Server) handleLocalIdentitySelfServiceSessionList(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, localIdentitySelfServiceSessionListRoute)
	identity, security, actor, requestSession, ok := server.requireLocalIdentitySelfServiceSecurityRequest(writer, request, trace)
	if !ok {
		return
	}
	query, err := localIdentitySelfServiceSessionListQuery(request)
	if err != nil {
		server.writeLocalIdentitySelfServiceSecurityError(writer, trace, err)
		return
	}
	page, err := security.ListSessions(request.Context(), actor, query)
	if err != nil {
		server.writeLocalIdentitySelfServiceSecurityError(writer, trace, err)
		return
	}
	identity.setCSRFCookie(writer, requestSession.csrfToken, requestSession.expiresAt)
	writeObservedJSON(writer, http.StatusOK, trace, page)
}

func (server *Server) handleLocalIdentitySelfServiceSessionRevoke(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, localIdentitySelfServiceSessionRevokeRoute)
	identity, security, actor, requestSession, ok := server.requireLocalIdentitySelfServiceSecurityMutation(writer, request, trace)
	if !ok {
		return
	}
	var body localIdentitySelfServiceSessionRevokeRequest
	if !server.decodeLocalIdentitySelfServiceSecurityBody(writer, request, trace, &body) {
		return
	}
	result, err := security.RevokeSession(request.Context(), actor, LocalIdentityRevokeOwnedSessionInput{
		SessionID:       request.PathValue("session_id"),
		ExpectedVersion: body.ExpectedRecordVersion,
		Confirmed:       body.Confirmed,
		AuditRef:        localIdentitySelfServiceSecurityAuditRef(trace, "session-revoke"),
	})
	if err != nil {
		server.writeLocalIdentitySelfServiceSecurityError(writer, trace, err)
		return
	}
	if result.CurrentSessionRevoked {
		identity.clearAuthenticationCookies(writer)
	} else {
		identity.setCSRFCookie(writer, requestSession.csrfToken, requestSession.expiresAt)
	}
	writeObservedJSON(writer, http.StatusOK, trace, result)
}

func (server *Server) handleLocalIdentitySelfServiceSessionRevokeOthers(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, localIdentitySelfServiceSessionRevokeOthersRoute)
	identity, security, actor, requestSession, ok := server.requireLocalIdentitySelfServiceSecurityMutation(writer, request, trace)
	if !ok {
		return
	}
	var body localIdentitySelfServiceSessionRevokeOthersRequest
	if !server.decodeLocalIdentitySelfServiceSecurityBody(writer, request, trace, &body) {
		return
	}
	result, err := security.RevokeOtherSessions(request.Context(), actor, LocalIdentityRevokeOtherSessionsInput{
		Confirmed: body.Confirmed,
		AuditRef:  localIdentitySelfServiceSecurityAuditRef(trace, "session-revoke-others"),
	})
	if err != nil {
		server.writeLocalIdentitySelfServiceSecurityError(writer, trace, err)
		return
	}
	identity.setCSRFCookie(writer, requestSession.csrfToken, requestSession.expiresAt)
	writeObservedJSON(writer, http.StatusOK, trace, result)
}

func (server *Server) handleLocalIdentitySelfServiceCredentialRotate(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, localIdentitySelfServiceCredentialRotateRoute)
	identity, security, actor, requestSession, ok := server.requireLocalIdentitySelfServiceSecurityMutation(writer, request, trace)
	if !ok {
		return
	}
	var body localIdentitySelfServiceCredentialRotateRequest
	if !server.decodeLocalIdentitySelfServiceSecurityBody(writer, request, trace, &body) {
		return
	}
	result, err := security.RotateCredential(request.Context(), actor, LocalIdentityRotateCredentialInput{
		CurrentPassword:        body.CurrentPassword,
		NewPassword:            body.NewPassword,
		SessionImpactConfirmed: body.SessionImpactConfirmed,
		AuditRef:               localIdentitySelfServiceSecurityAuditRef(trace, "credential-rotate"),
	})
	if err != nil {
		server.writeLocalIdentitySelfServiceSecurityError(writer, trace, err)
		return
	}
	if result.CurrentSessionRevoked {
		identity.clearAuthenticationCookies(writer)
	} else {
		identity.setCSRFCookie(writer, requestSession.csrfToken, requestSession.expiresAt)
	}
	writeObservedJSON(writer, http.StatusOK, trace, result)
}

func (server *Server) requireLocalIdentitySelfServiceSecurityMutation(
	writer http.ResponseWriter,
	request *http.Request,
	trace requestTrace,
) (*localIdentityHTTPService, *localIdentitySelfServiceSecurityService, LocalIdentitySelfServiceActor, localIdentityRequestSession, bool) {
	identity, security, actor, requestSession, ok := server.requireLocalIdentitySelfServiceSecurityRequest(writer, request, trace)
	if !ok {
		return nil, nil, LocalIdentitySelfServiceActor{}, localIdentityRequestSession{}, false
	}
	if request.URL.RawQuery != "" {
		server.writeLocalIdentityError(writer, trace, localIdentityPayloadInvalid)
		return nil, nil, LocalIdentitySelfServiceActor{}, localIdentityRequestSession{}, false
	}
	if !identity.requireAuthenticatedWriteRequest(
		writer, request, trace, requestSession.csrfToken, requestSession.csrfCookieValid,
	) {
		return nil, nil, LocalIdentitySelfServiceActor{}, localIdentityRequestSession{}, false
	}
	return identity, security, actor, requestSession, true
}

func (server *Server) requireLocalIdentitySelfServiceSecurityRequest(
	writer http.ResponseWriter,
	request *http.Request,
	trace requestTrace,
) (*localIdentityHTTPService, *localIdentitySelfServiceSecurityService, LocalIdentitySelfServiceActor, localIdentityRequestSession, bool) {
	identity, ok := server.requireLocalIdentityHTTP(writer, trace)
	if !ok {
		return nil, nil, LocalIdentitySelfServiceActor{}, localIdentityRequestSession{}, false
	}
	requestSession, ok := requireLocalIdentityRequestSession(request)
	if !ok {
		server.writeLocalIdentityError(writer, trace, localIdentityAuthenticationRequired)
		return nil, nil, LocalIdentitySelfServiceActor{}, localIdentityRequestSession{}, false
	}
	auth, ok := request.Context().Value(controlPlaneReadAuthContextKey{}).(controlPlaneReadAuthContext)
	if !ok || auth.AuthMode != localIdentityAuthMode || auth.FailureCode != "" || auth.VerifiedIdentity == nil ||
		auth.VerifiedIdentity.AuthSource != localIdentityAuthMode || auth.SubjectBinding != "user:"+requestSession.userID ||
		auth.SessionRef != "session:"+requestSession.sessionID {
		server.writeLocalIdentitySelfServiceSecurityError(writer, trace, errLocalIdentitySessionScopeDenied)
		return nil, nil, LocalIdentitySelfServiceActor{}, localIdentityRequestSession{}, false
	}
	security := server.localIdentitySelfServiceSecurityService
	if security == nil || security.repository == nil {
		server.writeLocalIdentitySelfServiceSecurityError(writer, trace, errLocalIdentitySelfServiceUnavailable)
		return nil, nil, LocalIdentitySelfServiceActor{}, localIdentityRequestSession{}, false
	}
	return identity, security, LocalIdentitySelfServiceActor{
		UserID:           requestSession.userID,
		CurrentSessionID: requestSession.sessionID,
		AuthenticatedAt:  requestSession.lastVerifiedAt,
	}, requestSession, true
}

func (server *Server) decodeLocalIdentitySelfServiceSecurityBody(
	writer http.ResponseWriter,
	request *http.Request,
	trace requestTrace,
	target any,
) bool {
	return server.decodeJSONRequestBody(writer, request, trace, target, jsonRequestBodyOptions{
		maxBytes: 4096, rejectUnknownFields: true, rejectDuplicateFields: true,
	})
}

func localIdentitySelfServiceSessionListQuery(request *http.Request) (LocalIdentitySelfServiceSessionListQuery, error) {
	values, err := url.ParseQuery(request.URL.RawQuery)
	if err != nil {
		return LocalIdentitySelfServiceSessionListQuery{}, errLocalIdentitySessionCursorInvalid
	}
	for key, entries := range values {
		if (key != "state" && key != "limit" && key != "cursor") || len(entries) != 1 || strings.TrimSpace(entries[0]) == "" {
			return LocalIdentitySelfServiceSessionListQuery{}, errLocalIdentitySessionCursorInvalid
		}
	}
	query := LocalIdentitySelfServiceSessionListQuery{
		State: strings.TrimSpace(values.Get("state")), Cursor: strings.TrimSpace(values.Get("cursor")),
	}
	if rawLimit := strings.TrimSpace(values.Get("limit")); rawLimit != "" {
		query.Limit, err = strconv.Atoi(rawLimit)
		if err != nil || query.Limit < 1 || query.Limit > localIdentitySelfServiceSessionMaximumLimit {
			return LocalIdentitySelfServiceSessionListQuery{}, errLocalIdentitySessionCursorInvalid
		}
	}
	return query, nil
}

func localIdentitySelfServiceSecurityAuditRef(trace requestTrace, action string) string {
	action = strings.TrimSpace(action)
	return "audit:local-identity:self-service:" + action + ":" + localIdentityDigest(
		"local_identity_self_service_security_http_v1", trace.requestID, action,
	)
}

func (server *Server) writeLocalIdentitySelfServiceSecurityError(writer http.ResponseWriter, trace requestTrace, err error) {
	code := localIdentityRepositoryError(err)
	if errors.Is(err, errLocalIdentityContractMismatch) || code == LocalIdentityFailureContractMismatch {
		server.writeLocalIdentityError(writer, trace, localIdentityPayloadInvalid)
		return
	}
	status, errorType, message := localIdentitySelfServiceSecurityHTTPErrorDefinition(code)
	writeTraceHeaders(writer, trace)
	writeJSON(writer, status, errorDocument{Error: errorBody{
		Message: message, Type: errorType, Code: code, RequestID: trace.requestID,
		Route: trace.route, FailureBoundary: "local_identity_self_service_security",
	}})
	logRequestTrace(trace, status, code, "local_identity_self_service_security")
}

func localIdentitySelfServiceSecurityHTTPErrorDefinition(code string) (int, string, string) {
	switch code {
	case LocalIdentityFailureSessionCursorInvalid:
		return http.StatusBadRequest, "invalid_request_error", "session list query or cursor is invalid"
	case LocalIdentityFailureSessionScopeDenied:
		return http.StatusForbidden, "permission_error", "the session operation is not available to this actor"
	case LocalIdentityFailureSessionVersionConflict:
		return http.StatusConflict, "invalid_request_error", "the session record changed before this request"
	case LocalIdentityFailureSessionRecentAuthentication:
		return http.StatusUnauthorized, "authentication_error", "the session operation requires recent authentication"
	case LocalIdentityFailureSessionBulkRevokeConflict:
		return http.StatusConflict, "invalid_request_error", "the session set changed before this request"
	case LocalIdentityFailureCredentialUnavailable:
		return http.StatusConflict, "authentication_error", "an active local credential is required"
	case LocalIdentityFailureCredentialCurrentInvalid:
		return http.StatusUnauthorized, "authentication_error", "the current credential could not be verified"
	case LocalIdentityFailureCredentialPolicyRejected:
		return http.StatusBadRequest, "invalid_request_error", "the replacement credential does not satisfy the accepted policy"
	case LocalIdentityFailureCredentialReuseDenied:
		return http.StatusConflict, "authentication_error", "the replacement credential must differ from the current credential"
	case LocalIdentityFailureCredentialRotationConflict:
		return http.StatusConflict, "invalid_request_error", "the local credential changed before this request"
	default:
		return http.StatusServiceUnavailable, "service_unavailable_error", "local identity self-service security is unavailable"
	}
}
