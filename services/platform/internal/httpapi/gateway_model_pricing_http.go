package httpapi

import (
	"net/http"
	"strings"
)

const (
	gatewayModelPricingAdminRoute        = "/v1/admin/gateway-model-pricing-policy"
	gatewayModelPricingAdminReadRoute    = "GET " + gatewayModelPricingAdminRoute
	gatewayModelPricingAdminPutRoute     = "PUT " + gatewayModelPricingAdminRoute
	gatewayModelPricingEnvironmentHeader = "X-RadishMind-Dev-Gateway-Pricing-Environment"
)

type gatewayModelPricingPutBody struct {
	ExpectedVersion               int64  `json:"expected_version"`
	ProviderID                    string `json:"provider_id"`
	ProfileID                     string `json:"profile_id"`
	ModelID                       string `json:"model_id"`
	Currency                      string `json:"currency"`
	InputPriceMicrosPerTokenUnit  int64  `json:"input_price_micros_per_token_unit"`
	OutputPriceMicrosPerTokenUnit int64  `json:"output_price_micros_per_token_unit"`
	Reason                        string `json:"reason"`
}

type gatewayModelPricingEnvelope struct {
	RequestID      string                     `json:"request_id"`
	TenantRef      string                     `json:"tenant_ref"`
	WorkspaceID    string                     `json:"workspace_id"`
	Environment    string                     `json:"environment"`
	ProviderID     string                     `json:"provider_id"`
	ProfileID      string                     `json:"profile_id"`
	ModelID        string                     `json:"model_id"`
	Policy         *GatewayModelPricingPolicy `json:"policy"`
	CurrentVersion int64                      `json:"current_version"`
	FailureCode    *string                    `json:"failure_code"`
	AuditRef       string                     `json:"audit_ref"`
}

func (server *Server) handleReadGatewayModelPricing(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, gatewayModelPricingAdminReadRoute)
	providerID, profileID, modelID, failureCode := parseGatewayModelPricingReadQuery(request)
	pricingContext, contextFailure, status := server.gatewayModelPricingAdminContext(
		request, trace, "admin_gateway_pricing:read", false, providerID, profileID, modelID,
	)
	if failureCode == "" {
		failureCode = contextFailure
	}
	if failureCode != "" {
		writeGatewayModelPricingResult(writer, trace, pricingContext, GatewayModelPricingResult{FailureCode: failureCode}, gatewayModelPricingHTTPStatus(failureCode, status))
		return
	}
	result := newGatewayModelPricingService(server.gatewayModelPricingRepository).ReadCurrent(pricingContext)
	writeGatewayModelPricingResult(writer, trace, pricingContext, result, gatewayModelPricingHTTPStatus(result.FailureCode, http.StatusOK))
}

func (server *Server) handlePutGatewayModelPricing(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, gatewayModelPricingAdminPutRoute)
	if len(request.URL.Query()) != 0 {
		writeGatewayModelPricingResult(writer, trace, GatewayModelPricingContext{}, GatewayModelPricingResult{
			FailureCode: GatewayModelPricingFailurePayloadInvalid,
		}, http.StatusBadRequest)
		return
	}
	var body gatewayModelPricingPutBody
	if err := decodeJSONRequestBodyValue(writer, request, &body, jsonRequestBodyOptions{
		maxBytes: maxControlJSONRequestBodyBytes, rejectUnknownFields: true, rejectDuplicateFields: true,
	}); err != nil {
		writeGatewayModelPricingResult(writer, trace, GatewayModelPricingContext{}, GatewayModelPricingResult{
			FailureCode: GatewayModelPricingFailurePayloadInvalid,
		}, http.StatusBadRequest)
		return
	}
	pricingContext, failureCode, status := server.gatewayModelPricingAdminContext(
		request, trace, "admin_gateway_pricing:write", true,
		body.ProviderID, body.ProfileID, body.ModelID,
	)
	if failureCode != "" {
		writeGatewayModelPricingResult(writer, trace, pricingContext, GatewayModelPricingResult{FailureCode: failureCode}, status)
		return
	}
	result := newGatewayModelPricingService(server.gatewayModelPricingRepository).PutRevision(
		pricingContext,
		GatewayModelPricingPolicyInput{
			ExpectedVersion: body.ExpectedVersion, Currency: body.Currency,
			InputPriceMicrosPerTokenUnit:  body.InputPriceMicrosPerTokenUnit,
			OutputPriceMicrosPerTokenUnit: body.OutputPriceMicrosPerTokenUnit,
			Reason:                        body.Reason,
		},
	)
	writeGatewayModelPricingResult(writer, trace, pricingContext, result, gatewayModelPricingHTTPStatus(result.FailureCode, http.StatusOK))
}

func parseGatewayModelPricingReadQuery(request *http.Request) (string, string, string, string) {
	values := request.URL.Query()
	allowed := map[string]bool{"provider_id": true, "profile_id": true, "model_id": true}
	for key, entries := range values {
		if !allowed[key] || len(entries) != 1 {
			return "", "", "", GatewayModelPricingFailurePayloadInvalid
		}
	}
	providerID := strings.TrimSpace(values.Get("provider_id"))
	profileID := strings.TrimSpace(values.Get("profile_id"))
	modelID := strings.TrimSpace(values.Get("model_id"))
	if len(values) != len(allowed) || providerID == "" || profileID == "" || modelID == "" {
		return providerID, profileID, modelID, GatewayModelPricingFailurePayloadInvalid
	}
	return providerID, profileID, modelID, ""
}

func (server *Server) gatewayModelPricingAdminContext(
	request *http.Request,
	trace requestTrace,
	requiredPermission string,
	writeRequired bool,
	providerID string,
	profileID string,
	modelID string,
) (GatewayModelPricingContext, string, int) {
	pricingContext := GatewayModelPricingContext{
		RequestContext: request.Context(),
		Environment:    strings.TrimSpace(request.Header.Get(gatewayModelPricingEnvironmentHeader)),
		ProviderID:     strings.TrimSpace(providerID),
		ProfileID:      strings.TrimSpace(profileID),
		ModelID:        strings.TrimSpace(modelID),
		RequestID:      trace.requestID,
		AuditRef:       "audit_" + trace.requestID + "_gateway-pricing-admin",
	}
	if !server.config.GatewayModelPricingDevHTTPEnabled ||
		writeRequired && !server.config.GatewayModelPricingDevWriteEnabled {
		return pricingContext, GatewayModelPricingFailureDisabled, http.StatusForbidden
	}
	auth, failureCode, _ := server.authorizeWorkspaceScopedPermissions(request, requiredPermission)
	if failureCode != "" || !workspacePermissionEnabled(auth, requiredPermission) {
		return pricingContext, GatewayModelPricingFailureScopeDenied, http.StatusForbidden
	}
	pricingContext.TenantRef = strings.TrimSpace(auth.TenantBinding)
	pricingContext.WorkspaceID = strings.TrimSpace(auth.ResourceBinding.WorkspaceID)
	pricingContext.ActorRef = strings.TrimSpace(auth.SubjectBinding)
	configuredEnvironment := strings.TrimSpace(server.config.GatewayModelPricingEnvironment)
	if pricingContext.Environment != configuredEnvironment ||
		pricingContext.Environment != "development" && pricingContext.Environment != "test" {
		return pricingContext, GatewayModelPricingFailureEnvironmentForbidden, http.StatusForbidden
	}
	if !validGatewayModelPricingContext(pricingContext) {
		return pricingContext, GatewayModelPricingFailureScopeConflict, http.StatusConflict
	}
	if server.gatewayModelPricingRepository == nil {
		return pricingContext, GatewayModelPricingFailureStoreUnavailable, http.StatusServiceUnavailable
	}
	return pricingContext, "", http.StatusOK
}

func writeGatewayModelPricingResult(
	writer http.ResponseWriter,
	trace requestTrace,
	pricingContext GatewayModelPricingContext,
	result GatewayModelPricingResult,
	status int,
) {
	writer.Header().Set("Cache-Control", "no-store")
	writeObservedJSON(writer, status, trace, gatewayModelPricingEnvelope{
		RequestID: trace.requestID, TenantRef: pricingContext.TenantRef, WorkspaceID: pricingContext.WorkspaceID,
		Environment: pricingContext.Environment, ProviderID: pricingContext.ProviderID,
		ProfileID: pricingContext.ProfileID, ModelID: pricingContext.ModelID,
		Policy: result.Policy, CurrentVersion: result.CurrentVersion,
		FailureCode: optionalGatewayModelPricingFailure(result.FailureCode), AuditRef: pricingContext.AuditRef,
	})
}

func optionalGatewayModelPricingFailure(failureCode string) *string {
	failureCode = strings.TrimSpace(failureCode)
	if failureCode == "" {
		return nil
	}
	return &failureCode
}

func gatewayModelPricingHTTPStatus(failureCode string, fallback int) int {
	switch failureCode {
	case "":
		return fallback
	case GatewayModelPricingFailureDisabled, GatewayModelPricingFailureScopeDenied,
		GatewayModelPricingFailureEnvironmentForbidden:
		return http.StatusForbidden
	case GatewayModelPricingFailurePayloadInvalid:
		return http.StatusBadRequest
	case GatewayModelPricingFailurePolicyNotFound:
		return http.StatusNotFound
	case GatewayModelPricingFailureVersionConflict, GatewayModelPricingFailureScopeConflict:
		return http.StatusConflict
	default:
		return http.StatusServiceUnavailable
	}
}
