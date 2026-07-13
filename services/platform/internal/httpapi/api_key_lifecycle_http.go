package httpapi

import (
	"net/http"
	"strconv"
	"strings"
)

const (
	apiKeyCreateRoute = "POST /v1/user-workspace/api-keys"
	apiKeyReadRoute   = "GET /v1/user-workspace/api-keys/{api_key_id}"
	apiKeyRevokeRoute = "POST /v1/user-workspace/api-keys/{api_key_id}/revoke"
)

type apiKeyCreateBody struct {
	WorkspaceID   string   `json:"workspace_id"`
	ApplicationID string   `json:"application_id"`
	DisplayName   string   `json:"display_name"`
	Scopes        []string `json:"scopes"`
	ExpiresInDays int      `json:"expires_in_days"`
}

type apiKeyRevokeBody struct {
	WorkspaceID     string `json:"workspace_id"`
	ExpectedVersion int    `json:"expected_version"`
}

type apiKeyIssueCredential struct {
	Token string `json:"token"`
}

type apiKeyEnvelope struct {
	RequestID             string                 `json:"request_id"`
	TenantRef             string                 `json:"tenant_ref"`
	WorkspaceID           string                 `json:"workspace_id"`
	Record                *APIKeyRecord          `json:"record"`
	Credential            *apiKeyIssueCredential `json:"credential,omitempty"`
	FailureCode           *string                `json:"failure_code"`
	CurrentRecordVersion  int                    `json:"current_record_version"`
	CurrentEffectiveState string                 `json:"current_effective_state"`
	AuditRef              string                 `json:"audit_ref"`
}

func (server *Server) handleCreateAPIKey(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, apiKeyCreateRoute)
	if !server.allowAPIKeyLifecycleDevHTTP(writer, trace) {
		return
	}
	var body apiKeyCreateBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: maxControlJSONRequestBodyBytes, rejectUnknownFields: true}) {
		return
	}
	requestContext, failureCode, status := server.apiKeyContextFromRequest(request, trace, body.WorkspaceID, "api_keys:write", "create")
	if failureCode != "" {
		writeAPIKeyResult(writer, status, trace, requestContext, APIKeyResult{FailureCode: failureCode}, true)
		return
	}
	requestContext.WriteEnabled = server.config.APIKeyLifecycleDevWriteEnabled
	result := server.apiKeyService().Create(requestContext, APIKeyCreateInput{
		ApplicationID: body.ApplicationID, DisplayName: body.DisplayName,
		Scopes: body.Scopes, ExpiresInDays: body.ExpiresInDays,
	})
	status = apiKeyResultHTTPStatus(result.FailureCode, http.StatusCreated)
	writeAPIKeyResult(writer, status, trace, requestContext, result, true)
}

func (server *Server) handleReadAPIKey(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, apiKeyReadRoute)
	if !server.allowAPIKeyLifecycleDevHTTP(writer, trace) {
		return
	}
	requestContext, failureCode, status := server.apiKeyContextFromRequest(request, trace, request.URL.Query().Get("workspace_id"), "api_keys:read", "read")
	if failureCode != "" {
		writeAPIKeyResult(writer, status, trace, requestContext, APIKeyResult{FailureCode: failureCode}, false)
		return
	}
	if !allowAPIKeyQuery(request, "workspace_id") {
		writeAPIKeyResult(writer, http.StatusBadRequest, trace, requestContext, APIKeyResult{FailureCode: APIKeyFailurePayloadInvalid}, false)
		return
	}
	result := server.apiKeyService().Read(requestContext, request.PathValue("api_key_id"))
	writeAPIKeyResult(writer, apiKeyResultHTTPStatus(result.FailureCode, http.StatusOK), trace, requestContext, result, false)
}

func (server *Server) handleRevokeAPIKey(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, apiKeyRevokeRoute)
	if !server.allowAPIKeyLifecycleDevHTTP(writer, trace) {
		return
	}
	var body apiKeyRevokeBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: maxControlJSONRequestBodyBytes, rejectUnknownFields: true}) {
		return
	}
	requestContext, failureCode, status := server.apiKeyContextFromRequest(request, trace, body.WorkspaceID, "api_keys:revoke", "revoke")
	if failureCode != "" {
		writeAPIKeyResult(writer, status, trace, requestContext, APIKeyResult{FailureCode: failureCode}, false)
		return
	}
	requestContext.WriteEnabled = server.config.APIKeyLifecycleDevWriteEnabled
	result := server.apiKeyService().Revoke(requestContext, request.PathValue("api_key_id"), body.ExpectedVersion)
	writeAPIKeyResult(writer, apiKeyResultHTTPStatus(result.FailureCode, http.StatusOK), trace, requestContext, result, false)
}

func (server *Server) handleListAPIKeys(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, controlPlaneAPIKeySummaryListRoute)
	if !server.allowAPIKeyLifecycleDevHTTP(writer, trace) {
		return
	}
	requestContext, failureCode, status := server.apiKeyContextFromRequest(request, trace, request.URL.Query().Get("workspace_id"), "api_keys:read", "list")
	if failureCode != "" {
		writeAPIKeyListResult(writer, status, trace, requestContext, APIKeyListResult{Records: []APIKeyRecord{}, FailureCode: failureCode})
		return
	}
	if !allowAPIKeyQuery(request, "workspace_id", "application_id", "effective_state", "limit", "cursor") {
		writeAPIKeyListResult(writer, http.StatusBadRequest, trace, requestContext, APIKeyListResult{Records: []APIKeyRecord{}, FailureCode: APIKeyFailurePayloadInvalid})
		return
	}
	limit := 0
	if value := strings.TrimSpace(request.URL.Query().Get("limit")); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			writeAPIKeyListResult(writer, http.StatusBadRequest, trace, requestContext, APIKeyListResult{Records: []APIKeyRecord{}, FailureCode: APIKeyFailurePayloadInvalid})
			return
		}
		limit = parsed
	}
	result := server.apiKeyService().List(requestContext, APIKeyListInput{
		ApplicationID: request.URL.Query().Get("application_id"), EffectiveState: request.URL.Query().Get("effective_state"),
		Limit: limit, Cursor: request.URL.Query().Get("cursor"),
	})
	writeAPIKeyListResult(writer, apiKeyResultHTTPStatus(result.FailureCode, http.StatusOK), trace, requestContext, result)
}

func (server *Server) apiKeyContextFromRequest(request *http.Request, trace requestTrace, workspaceID, requiredScope, auditSuffix string) (APIKeyContext, string, int) {
	auth, failureCode, status := authorizeControlPlaneReadRequest(request, "", requiredScope)
	requestContext := APIKeyContext{
		RequestContext: request.Context(), RequestID: trace.requestID, TenantRef: strings.TrimSpace(auth.TenantBinding),
		WorkspaceID: strings.TrimSpace(workspaceID), ActorRef: strings.TrimSpace(auth.SubjectBinding),
		OwnerSubjectRef: strings.TrimSpace(auth.SubjectBinding), AuditRef: "audit_" + trace.requestID + "_api-key-" + auditSuffix,
	}
	if failureCode != "" {
		if failureCode == "workspace_membership_unavailable" {
			return requestContext, failureCode, status
		}
		return requestContext, APIKeyFailureScopeDenied, status
	}
	if requestContext.WorkspaceID == "" || !applicationDraftIdentifierPattern.MatchString(requestContext.WorkspaceID) {
		return requestContext, APIKeyFailureScopeDenied, http.StatusForbidden
	}
	return requestContext, "", http.StatusOK
}

func (server *Server) allowAPIKeyLifecycleDevHTTP(writer http.ResponseWriter, trace requestTrace) bool {
	if server.config.APIKeyLifecycleDevHTTPEnabled && server.config.ApplicationCatalogDevHTTPEnabled {
		return true
	}
	server.writePlatformError(writer, trace, "API_KEY_LIFECYCLE_DEV_HTTP_DISABLED", "API key lifecycle dev route requires explicit opt-in and application catalog")
	return false
}

func (server *Server) apiKeyService() apiKeyService {
	if server.apiKeyRepository == nil {
		server.apiKeyRepository = &memoryAPIKeyRepository{records: make(map[string]APIKeyRecord), unavailable: true}
	}
	return newAPIKeyService(server.apiKeyRepository, server.applicationCatalogRepository)
}

func allowAPIKeyQuery(request *http.Request, allowed ...string) bool {
	allowlist := make(map[string]struct{}, len(allowed))
	for _, key := range allowed {
		allowlist[key] = struct{}{}
	}
	for key := range request.URL.Query() {
		if _, ok := allowlist[key]; !ok {
			return false
		}
	}
	return true
}

func writeAPIKeyResult(writer http.ResponseWriter, status int, trace requestTrace, requestContext APIKeyContext, result APIKeyResult, issueResponse bool) {
	var credential *apiKeyIssueCredential
	if issueResponse {
		writer.Header().Set("Cache-Control", "no-store")
		if result.FailureCode == "" && result.CredentialToken != "" {
			credential = &apiKeyIssueCredential{Token: result.CredentialToken}
		}
	}
	writeObservedJSON(writer, status, trace, apiKeyEnvelope{
		RequestID: trace.requestID, TenantRef: requestContext.TenantRef, WorkspaceID: requestContext.WorkspaceID,
		Record: result.Record, Credential: credential, FailureCode: optionalApplicationDraftFailure(result.FailureCode),
		CurrentRecordVersion: result.CurrentRecordVersion, CurrentEffectiveState: result.CurrentEffectiveState,
		AuditRef: requestContext.AuditRef,
	})
}

func writeAPIKeyListResult(writer http.ResponseWriter, status int, trace requestTrace, requestContext APIKeyContext, result APIKeyListResult) {
	items := make([]map[string]any, 0, len(result.Records))
	for _, record := range result.Records {
		items = append(items, apiKeyRecordToSummaryMap(record))
	}
	writeObservedJSON(writer, status, trace, controlPlaneReadEnvelope{
		RequestID: trace.requestID, TenantRef: requestContext.TenantRef, Items: items,
		NextCursor: result.NextCursor, FailureCode: optionalApplicationDraftFailure(result.FailureCode), AuditRef: requestContext.AuditRef,
	})
}

func apiKeyRecordToSummaryMap(record APIKeyRecord) map[string]any {
	return map[string]any{
		"api_key_id": record.APIKeyID, "tenant_ref": record.TenantRef, "workspace_id": record.WorkspaceID,
		"application_id": record.ApplicationID, "owner_subject_ref": record.OwnerSubjectRef,
		"display_name": record.DisplayName, "scopes": append([]string{}, record.Scopes...),
		"state": record.EffectiveState, "lifecycle_state": record.LifecycleState,
		"effective_state": record.EffectiveState, "record_version": record.RecordVersion,
		"created_at": record.CreatedAt, "expires_at": record.ExpiresAt,
		"last_used_at": record.LastUsedAt, "revoked_at": record.RevokedAt,
	}
}

func apiKeyResultHTTPStatus(failureCode string, successStatus int) int {
	switch failureCode {
	case "":
		return successStatus
	case APIKeyFailurePayloadInvalid, APIKeyFailureSecretForbidden, APIKeyFailureCursorInvalid:
		return http.StatusBadRequest
	case APIKeyFailureScopeDenied, APIKeyFailureApplicationUnavailable, APIKeyFailureWriteDisabled,
		APIKeyFailureRevoked, APIKeyFailureExpired, APIKeyFailureLifecycleDisabled:
		return http.StatusForbidden
	case APIKeyFailureNotFound:
		return http.StatusNotFound
	case APIKeyFailureVersionConflict, APIKeyFailureTransitionInvalid:
		return http.StatusConflict
	case APIKeyFailureCatalogRequired, APIKeyFailureStoreUnavailable, "workspace_membership_unavailable":
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}
