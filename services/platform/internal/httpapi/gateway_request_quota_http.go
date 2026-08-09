package httpapi

import (
	"net/http"
	"strings"
	"time"
)

const (
	gatewayRequestQuotaAdminRoute        = "/v1/admin/gateway-request-quotas/{application_id}"
	gatewayRequestQuotaAdminReadRoute    = "GET " + gatewayRequestQuotaAdminRoute
	gatewayRequestQuotaAdminPutRoute     = "PUT " + gatewayRequestQuotaAdminRoute
	gatewayRequestQuotaEnvironmentHeader = "X-RadishMind-Dev-Gateway-Quota-Environment"
)

type gatewayRequestQuotaPutBody struct {
	ExpectedVersion int64 `json:"expected_version"`
	RequestLimit    int64 `json:"request_limit"`
}

type gatewayRequestQuotaEnvelope struct {
	RequestID     string                     `json:"request_id"`
	TenantRef     string                     `json:"tenant_ref"`
	WorkspaceID   string                     `json:"workspace_id"`
	Environment   string                     `json:"environment"`
	ApplicationID string                     `json:"application_id"`
	Policy        *GatewayRequestQuotaPolicy `json:"policy"`
	Usage         *GatewayRequestQuotaUsage  `json:"usage"`
	FailureCode   *string                    `json:"failure_code"`
	AuditRef      string                     `json:"audit_ref"`
}

func (server *Server) handleReadGatewayRequestQuota(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, gatewayRequestQuotaAdminReadRoute)
	quotaContext, failureCode, status := server.gatewayRequestQuotaAdminContext(request, trace, "admin_gateway_quotas:read", false)
	if failureCode != "" {
		writeGatewayRequestQuotaResult(writer, trace, quotaContext, nil, nil, failureCode, status)
		return
	}
	policy, found, err := server.gatewayRequestQuotaRepository.ReadPolicy(quotaContext)
	if err != nil || !found {
		failureCode = GatewayRequestQuotaFailureStoreUnavailable
		if err == nil {
			failureCode = GatewayRequestQuotaFailurePolicyNotFound
		}
		writeGatewayRequestQuotaResult(writer, trace, quotaContext, nil, nil, failureCode, gatewayRequestQuotaHTTPStatus(failureCode))
		return
	}
	usage := currentGatewayRequestQuotaUsage(server.gatewayRequestQuotaRepository, quotaContext, policy, time.Now().UTC())
	if usage == nil {
		writeGatewayRequestQuotaResult(writer, trace, quotaContext, nil, nil, GatewayRequestQuotaFailureStoreUnavailable, http.StatusServiceUnavailable)
		return
	}
	writeGatewayRequestQuotaResult(writer, trace, quotaContext, &policy, usage, "", http.StatusOK)
}

func (server *Server) handlePutGatewayRequestQuota(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, gatewayRequestQuotaAdminPutRoute)
	quotaContext, failureCode, status := server.gatewayRequestQuotaAdminContext(request, trace, "admin_gateway_quotas:write", true)
	if failureCode != "" {
		writeGatewayRequestQuotaResult(writer, trace, quotaContext, nil, nil, failureCode, status)
		return
	}
	var body gatewayRequestQuotaPutBody
	if err := decodeJSONRequestBodyValue(writer, request, &body, jsonRequestBodyOptions{
		maxBytes: maxControlJSONRequestBodyBytes, rejectUnknownFields: true, rejectDuplicateFields: true,
	}); err != nil {
		writeGatewayRequestQuotaResult(
			writer, trace, quotaContext, nil, nil,
			GatewayRequestQuotaFailurePayloadInvalid, http.StatusBadRequest,
		)
		return
	}
	policy, err := server.gatewayRequestQuotaRepository.PutPolicy(
		quotaContext, body.ExpectedVersion, body.RequestLimit, time.Now().UTC(),
	)
	if err != nil {
		failureCode = gatewayRequestQuotaFailureFromRepositoryError(err)
		writeGatewayRequestQuotaResult(writer, trace, quotaContext, nil, nil, failureCode, gatewayRequestQuotaHTTPStatus(failureCode))
		return
	}
	usage := currentGatewayRequestQuotaUsage(server.gatewayRequestQuotaRepository, quotaContext, policy, time.Now().UTC())
	if usage == nil {
		writeGatewayRequestQuotaResult(writer, trace, quotaContext, nil, nil, GatewayRequestQuotaFailureStoreUnavailable, http.StatusServiceUnavailable)
		return
	}
	writeGatewayRequestQuotaResult(writer, trace, quotaContext, &policy, usage, "", http.StatusOK)
}

func (server *Server) gatewayRequestQuotaAdminContext(
	request *http.Request,
	trace requestTrace,
	requiredPermission string,
	writeRequired bool,
) (GatewayRequestQuotaContext, string, int) {
	quotaContext := GatewayRequestQuotaContext{
		RequestContext: request.Context(),
		Environment:    strings.TrimSpace(request.Header.Get(gatewayRequestQuotaEnvironmentHeader)),
		ApplicationID:  strings.TrimSpace(request.PathValue("application_id")),
		RequestID:      trace.requestID,
		AuditRef:       "audit_" + trace.requestID + "_gateway-quota-admin",
	}
	writerGateFailure := ""
	if !server.config.GatewayRequestQuotaDevHTTPEnabled {
		writerGateFailure = GatewayRequestQuotaFailureDisabled
	} else if writeRequired && !server.config.GatewayRequestQuotaDevWriteEnabled {
		writerGateFailure = GatewayRequestQuotaFailureDisabled
	}
	if writerGateFailure != "" {
		return quotaContext, writerGateFailure, http.StatusForbidden
	}
	if len(request.URL.Query()) != 0 || !gatewayRequestQuotaIdentifierPattern.MatchString(quotaContext.ApplicationID) {
		return quotaContext, GatewayRequestQuotaFailurePayloadInvalid, http.StatusBadRequest
	}
	auth, failureCode, status := server.authorizeWorkspaceScopedPermissions(request, requiredPermission)
	if failureCode != "" {
		if failureCode == "scope_denied" || failureCode == "workspace_permission_denied" {
			return quotaContext, GatewayRequestQuotaFailureScopeDenied, http.StatusForbidden
		}
		return quotaContext, failureCode, status
	}
	if !workspacePermissionEnabled(auth, requiredPermission) {
		return quotaContext, GatewayRequestQuotaFailureScopeDenied, http.StatusForbidden
	}
	quotaContext.TenantRef = strings.TrimSpace(auth.TenantBinding)
	quotaContext.WorkspaceID = strings.TrimSpace(auth.ResourceBinding.WorkspaceID)
	quotaContext.ActorRef = strings.TrimSpace(auth.SubjectBinding)
	if quotaContext.Environment != "development" && quotaContext.Environment != "test" {
		return quotaContext, GatewayRequestQuotaFailureEnvironmentForbidden, http.StatusForbidden
	}
	if server.gatewayRequestQuotaRepository == nil {
		return quotaContext, GatewayRequestQuotaFailureStoreUnavailable, http.StatusServiceUnavailable
	}
	return quotaContext, "", http.StatusOK
}

func currentGatewayRequestQuotaUsage(
	repository GatewayRequestQuotaRepository,
	quotaContext GatewayRequestQuotaContext,
	policy GatewayRequestQuotaPolicy,
	now time.Time,
) *GatewayRequestQuotaUsage {
	periodStart := gatewayRequestQuotaPeriodStart(now)
	usage, found, err := repository.ReadUsage(quotaContext, periodStart)
	if err != nil {
		return nil
	}
	if !found {
		usage = GatewayRequestQuotaUsage{
			SchemaVersion: GatewayRequestQuotaSchemaVersion, TenantRef: quotaContext.TenantRef,
			WorkspaceID: quotaContext.WorkspaceID, Environment: quotaContext.Environment,
			ApplicationID: quotaContext.ApplicationID, Period: GatewayRequestQuotaPeriod, PeriodStart: periodStart,
			PolicyID: policy.PolicyID, PolicyVersion: policy.RecordVersion, RequestLimit: policy.RequestLimit,
			RemainingRequestCount: policy.RequestLimit, UpdatedAt: policy.UpdatedAt,
		}
	}
	return &usage
}

func writeGatewayRequestQuotaResult(
	writer http.ResponseWriter,
	trace requestTrace,
	quotaContext GatewayRequestQuotaContext,
	policy *GatewayRequestQuotaPolicy,
	usage *GatewayRequestQuotaUsage,
	failureCode string,
	status int,
) {
	writer.Header().Set("Cache-Control", "no-store")
	writeObservedJSON(writer, status, trace, gatewayRequestQuotaEnvelope{
		RequestID: trace.requestID, TenantRef: quotaContext.TenantRef, WorkspaceID: quotaContext.WorkspaceID,
		Environment: quotaContext.Environment, ApplicationID: quotaContext.ApplicationID,
		Policy: policy, Usage: usage, FailureCode: optionalGatewayRequestQuotaFailure(failureCode), AuditRef: quotaContext.AuditRef,
	})
}

func optionalGatewayRequestQuotaFailure(failureCode string) *string {
	failureCode = strings.TrimSpace(failureCode)
	if failureCode == "" {
		return nil
	}
	return &failureCode
}
