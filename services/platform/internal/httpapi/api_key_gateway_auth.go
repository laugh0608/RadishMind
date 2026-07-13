package httpapi

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"radishmind.local/services/platform/internal/config"
)

const gatewayAPIKeyAuthenticationSource = "api_key_dev_test"

type gatewayAPIKeyAuthenticationResult struct {
	RequestContext GatewayRequestContext
	FailureCode    string
}

func (server *Server) prepareGatewayRequest(
	writer http.ResponseWriter,
	request *http.Request,
	trace *requestTrace,
	protocol string,
	requiredScope string,
) bool {
	if trace == nil {
		return false
	}
	if config.EffectiveGatewayAuthMode(server.config) == gatewayAPIKeyAuthenticationSource {
		result := server.authenticateGatewayAPIKey(request, *trace, requiredScope)
		if result.FailureCode != "" {
			server.writePlatformError(writer, *trace, result.FailureCode, "")
			return false
		}
		trace.gatewayRequestContext = result.RequestContext
		if !server.config.GatewayRequestHistoryDevEnabled {
			server.writePlatformError(writer, *trace, string(GatewayRequestHistoryFailureStore), "")
			return false
		}
	}
	if err := server.startGatewayRequestTrace(request, trace, protocol); err != nil {
		server.writePlatformError(writer, *trace, string(GatewayRequestHistoryFailureStore), "")
		return false
	}
	return true
}

func (server *Server) authenticateGatewayAPIKey(
	request *http.Request,
	trace requestTrace,
	requiredScope string,
) gatewayAPIKeyAuthenticationResult {
	if request == nil || server == nil || server.apiKeyRepository == nil || server.applicationCatalogRepository == nil {
		return gatewayAPIKeyAuthenticationResult{FailureCode: APIKeyFailureStoreUnavailable}
	}
	if gatewayRequestHasAnyDevIdentityHeader(request) {
		return gatewayAPIKeyAuthenticationResult{FailureCode: APIKeyFailureCredentialConflict}
	}
	values := request.Header.Values("Authorization")
	if len(values) == 0 || strings.TrimSpace(values[0]) == "" {
		return gatewayAPIKeyAuthenticationResult{FailureCode: APIKeyFailureMissing}
	}
	parts := strings.Fields(values[0])
	if len(values) != 1 || len(parts) != 2 || parts[0] != "Bearer" {
		return gatewayAPIKeyAuthenticationResult{FailureCode: APIKeyFailureInvalid}
	}
	token := parts[1]
	apiKeyID, ok := parseAPIKeyCredential(token)
	if !ok {
		return gatewayAPIKeyAuthenticationResult{FailureCode: APIKeyFailureInvalid}
	}
	record, err := server.apiKeyRepository.FindCredential(request.Context(), apiKeyID)
	if errors.Is(err, errAPIKeyNotFound) {
		return gatewayAPIKeyAuthenticationResult{FailureCode: APIKeyFailureInvalid}
	}
	if err != nil {
		return gatewayAPIKeyAuthenticationResult{FailureCode: APIKeyFailureStoreUnavailable}
	}
	if !apiKeyCredentialMatches(token, record.credentialDigest) {
		return gatewayAPIKeyAuthenticationResult{FailureCode: APIKeyFailureInvalid}
	}
	if record.LifecycleState == apiKeyLifecycleRevoked {
		return gatewayAPIKeyAuthenticationResult{FailureCode: APIKeyFailureRevoked}
	}
	now := time.Now().UTC()
	if effectiveAPIKeyState(record, now) == apiKeyEffectiveExpired {
		return gatewayAPIKeyAuthenticationResult{FailureCode: APIKeyFailureExpired}
	}
	if !apiKeyScopeGranted(record.Scopes, requiredScope) {
		return gatewayAPIKeyAuthenticationResult{FailureCode: APIKeyFailureScopeDenied}
	}
	applicationContext := ApplicationCatalogContext{
		RequestContext: request.Context(), RequestID: trace.requestID,
		TenantRef: record.TenantRef, WorkspaceID: record.WorkspaceID,
		ActorRef: record.OwnerSubjectRef, OwnerSubjectRef: record.OwnerSubjectRef,
		AuditRef: "audit_" + trace.requestID + "_api-key-application",
	}
	active := newApplicationCatalogService(server.applicationCatalogRepository).RequireActive(applicationContext, record.ApplicationID)
	if active.FailureCode == ApplicationCatalogFailureStoreUnavailable {
		return gatewayAPIKeyAuthenticationResult{FailureCode: APIKeyFailureStoreUnavailable}
	}
	if active.FailureCode != "" || active.Record == nil {
		return gatewayAPIKeyAuthenticationResult{FailureCode: APIKeyFailureApplicationUnavailable}
	}
	updated, err := server.apiKeyRepository.RecordSuccessfulAuthentication(request.Context(), record.APIKeyID, record.RecordVersion, now)
	if err != nil {
		switch {
		case errors.Is(err, errAPIKeyRevoked):
			return gatewayAPIKeyAuthenticationResult{FailureCode: APIKeyFailureRevoked}
		case errors.Is(err, errAPIKeyExpired):
			return gatewayAPIKeyAuthenticationResult{FailureCode: APIKeyFailureExpired}
		default:
			return gatewayAPIKeyAuthenticationResult{FailureCode: APIKeyFailureStoreUnavailable}
		}
	}
	return gatewayAPIKeyAuthenticationResult{RequestContext: GatewayRequestContext{
		RequestContext: request.Context(), TenantRef: updated.TenantRef, WorkspaceID: updated.WorkspaceID,
		ConsumerRef: "api_key:" + updated.APIKeyID, ApplicationID: updated.ApplicationID,
		SubjectRef: updated.OwnerSubjectRef, ScopeGrants: append([]string{}, updated.Scopes...),
		AuditContext: "api-key-dev-test", Source: gatewayAPIKeyAuthenticationSource,
		RequestID: trace.requestID, AuditRef: "audit_" + trace.requestID + "_gateway-request",
	}}
}

func apiKeyScopeGranted(scopes []string, required string) bool {
	required = strings.TrimSpace(required)
	for _, scope := range scopes {
		if strings.TrimSpace(scope) == required {
			return true
		}
	}
	return false
}

func gatewayRequestHasAnyDevIdentityHeader(request *http.Request) bool {
	if request == nil {
		return false
	}
	for _, header := range []string{
		gatewayRequestDevTenantHeader, gatewayRequestDevWorkspaceHeader, gatewayRequestDevConsumerHeader,
		gatewayRequestDevApplicationHeader, gatewayRequestDevSubjectHeader, gatewayRequestDevScopesHeader,
		gatewayRequestDevAuditHeader,
	} {
		if len(request.Header.Values(header)) > 0 {
			return true
		}
	}
	return false
}

func isGatewayNorthboundRequest(request *http.Request) bool {
	if request == nil {
		return false
	}
	path := request.URL.Path
	if request.Method == http.MethodGet && (path == "/v1/models" || strings.HasPrefix(path, "/v1/models/")) {
		return true
	}
	return request.Method == http.MethodPost && (path == "/v1/chat/completions" || path == "/v1/responses" || path == "/v1/messages")
}
