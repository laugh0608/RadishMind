package httpapi

import (
	"net/http"
	"strings"
)

const (
	adminProviderRouteConfigurationRoute   = "/v1/admin/provider-route-configurations/{configuration_id}"
	adminProviderRouteDraftReadRoute       = "GET " + adminProviderRouteConfigurationRoute
	adminProviderRouteDraftPutRoute        = "PUT " + adminProviderRouteConfigurationRoute
	adminProviderRouteCandidateCreateRoute = "POST " +
		"/v1/admin/provider-route-configurations/{configuration_id}/candidates"
	adminProviderRouteCandidateReadRoute = "GET " +
		"/v1/admin/provider-route-configurations/{configuration_id}/candidates/{candidate_id}"
	adminProviderRouteReviewRoute = "POST " +
		"/v1/admin/provider-route-configurations/{configuration_id}/candidates/{candidate_id}/reviews"
	adminProviderRouteActivationRoute = "POST " +
		"/v1/admin/provider-route-configurations/{configuration_id}/candidates/{candidate_id}/activations"
	adminProviderRouteActiveSnapshotRoute = "GET " +
		"/v1/admin/provider-route-configurations/{configuration_id}/active-snapshot"
	adminProviderRouteActivationHistoryRoute = "GET " +
		"/v1/admin/provider-route-configurations/{configuration_id}/activation-history"

	adminProviderRouteDevWorkspaceHeader   = "X-RadishMind-Dev-Admin-Provider-Route-Workspace"
	adminProviderRouteDevEnvironmentHeader = "X-RadishMind-Dev-Admin-Provider-Route-Environment"
)

type adminProviderRouteDraftPutBody struct {
	ExpectedRevision int                              `json:"expected_revision"`
	DisplayName      string                           `json:"display_name"`
	ProviderProfiles []AdminProviderProfileAssignment `json:"provider_profiles"`
	ModelRoutes      []AdminModelRouteDefinition      `json:"model_routes"`
}

type adminProviderRouteCandidateCreateBody struct {
	CandidateID           string `json:"candidate_id"`
	ExpectedDraftRevision int    `json:"expected_draft_revision"`
}

type adminProviderRouteReviewBody struct {
	ExpectedReviewVersion int    `json:"expected_review_version"`
	Decision              string `json:"decision"`
	Reason                string `json:"reason"`
}

type adminProviderRouteActivationBody struct {
	ExpectedGeneration int    `json:"expected_generation"`
	Action             string `json:"action"`
	Reason             string `json:"reason"`
}

type adminProviderRouteEnvelope struct {
	RequestID             string                                `json:"request_id"`
	WorkspaceID           string                                `json:"workspace_id"`
	Environment           string                                `json:"environment"`
	ConfigurationID       string                                `json:"configuration_id"`
	CandidateID           string                                `json:"candidate_id,omitempty"`
	Draft                 *AdminProviderRouteConfigurationDraft `json:"draft,omitempty"`
	Candidate             *AdminProviderRouteCandidate          `json:"candidate,omitempty"`
	Snapshot              *AdminProviderRouteSnapshot           `json:"snapshot,omitempty"`
	Activation            *AdminProviderRouteActivationRecord   `json:"activation,omitempty"`
	ActivationHistory     []AdminProviderRouteActivationRecord  `json:"activation_history"`
	FailureCode           *string                               `json:"failure_code"`
	CurrentDraftRevision  int                                   `json:"current_draft_revision,omitempty"`
	CurrentReviewVersion  int                                   `json:"current_review_version,omitempty"`
	CurrentCandidateState string                                `json:"current_candidate_state,omitempty"`
	CurrentGeneration     int                                   `json:"current_generation,omitempty"`
	AuditRef              string                                `json:"audit_ref"`
}

func (server *Server) handleReadAdminProviderRouteDraft(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, adminProviderRouteDraftReadRoute)
	if !server.prepareAdminProviderRouteHTTP(writer, request, trace) {
		return
	}
	ctx, failureCode, status := server.adminProviderRouteContextFromRequest(request, trace, "admin_provider_routes:read", false, "draft-read")
	if failureCode != "" {
		writeAdminProviderRouteResult(writer, trace, ctx, request.PathValue("configuration_id"), "", AdminProviderRouteResult{FailureCode: failureCode}, status)
		return
	}
	result := server.adminProviderRouteService().ReadDraft(ctx, request.PathValue("configuration_id"))
	writeAdminProviderRouteResult(writer, trace, ctx, request.PathValue("configuration_id"), "", result, http.StatusOK)
}

func (server *Server) handlePutAdminProviderRouteDraft(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, adminProviderRouteDraftPutRoute)
	if !server.prepareAdminProviderRouteHTTP(writer, request, trace) {
		return
	}
	ctx, failureCode, status := server.adminProviderRouteContextFromRequest(request, trace, "admin_provider_routes:draft", true, "draft-put")
	if failureCode != "" {
		writeAdminProviderRouteResult(writer, trace, ctx, request.PathValue("configuration_id"), "", AdminProviderRouteResult{FailureCode: failureCode}, status)
		return
	}
	var body adminProviderRouteDraftPutBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, adminProviderRouteJSONRequestBodyOptions()) {
		return
	}
	if !adminProviderRouteHTTPV1RoutesOnly(body.ModelRoutes) {
		writeAdminProviderRouteResult(
			writer, trace, ctx, request.PathValue("configuration_id"), "",
			AdminProviderRouteResult{FailureCode: AdminProviderRouteFailurePayloadInvalid}, http.StatusBadRequest,
		)
		return
	}
	result := server.adminProviderRouteService().PutDraft(ctx, AdminProviderRouteDraftInput{
		ConfigurationID: request.PathValue("configuration_id"), ExpectedRevision: body.ExpectedRevision,
		DisplayName: body.DisplayName, ProviderProfiles: body.ProviderProfiles, ModelRoutes: body.ModelRoutes,
	})
	writeAdminProviderRouteResult(writer, trace, ctx, request.PathValue("configuration_id"), "", result, http.StatusOK)
}

func (server *Server) handleCreateAdminProviderRouteCandidate(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, adminProviderRouteCandidateCreateRoute)
	if !server.prepareAdminProviderRouteHTTP(writer, request, trace) {
		return
	}
	ctx, failureCode, status := server.adminProviderRouteContextFromRequest(request, trace, "admin_provider_routes:draft", true, "candidate-create")
	if failureCode != "" {
		writeAdminProviderRouteResult(writer, trace, ctx, request.PathValue("configuration_id"), "", AdminProviderRouteResult{FailureCode: failureCode}, status)
		return
	}
	var body adminProviderRouteCandidateCreateBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, adminProviderRouteJSONRequestBodyOptions()) {
		return
	}
	result := server.adminProviderRouteService().CreateCandidate(ctx, AdminProviderRouteCandidateInput{
		ConfigurationID: request.PathValue("configuration_id"), CandidateID: body.CandidateID,
		ExpectedDraftRevision: body.ExpectedDraftRevision,
	})
	writeAdminProviderRouteResult(writer, trace, ctx, request.PathValue("configuration_id"), body.CandidateID, result, http.StatusCreated)
}

func (server *Server) handleReadAdminProviderRouteCandidate(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, adminProviderRouteCandidateReadRoute)
	if !server.prepareAdminProviderRouteHTTP(writer, request, trace) {
		return
	}
	ctx, failureCode, status := server.adminProviderRouteContextFromRequest(request, trace, "admin_provider_routes:read", false, "candidate-read")
	if failureCode != "" {
		writeAdminProviderRouteResult(writer, trace, ctx, request.PathValue("configuration_id"), request.PathValue("candidate_id"), AdminProviderRouteResult{FailureCode: failureCode}, status)
		return
	}
	result := server.adminProviderRouteService().ReadCandidate(ctx, request.PathValue("configuration_id"), request.PathValue("candidate_id"))
	writeAdminProviderRouteResult(writer, trace, ctx, request.PathValue("configuration_id"), request.PathValue("candidate_id"), result, http.StatusOK)
}

func (server *Server) handleReviewAdminProviderRouteCandidate(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, adminProviderRouteReviewRoute)
	if !server.prepareAdminProviderRouteHTTP(writer, request, trace) {
		return
	}
	ctx, failureCode, status := server.adminProviderRouteContextFromRequest(request, trace, "admin_provider_routes:review", true, "candidate-review")
	if failureCode != "" {
		writeAdminProviderRouteResult(writer, trace, ctx, request.PathValue("configuration_id"), request.PathValue("candidate_id"), AdminProviderRouteResult{FailureCode: failureCode}, status)
		return
	}
	var body adminProviderRouteReviewBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, adminProviderRouteJSONRequestBodyOptions()) {
		return
	}
	result := server.adminProviderRouteService().ReviewCandidate(
		ctx, request.PathValue("configuration_id"), request.PathValue("candidate_id"),
		AdminProviderRouteReviewInput{
			ExpectedReviewVersion: body.ExpectedReviewVersion, Decision: body.Decision, Reason: body.Reason,
		},
	)
	writeAdminProviderRouteResult(writer, trace, ctx, request.PathValue("configuration_id"), request.PathValue("candidate_id"), result, http.StatusOK)
}

func (server *Server) handleActivateAdminProviderRouteCandidate(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, adminProviderRouteActivationRoute)
	if !server.prepareAdminProviderRouteHTTP(writer, request, trace) {
		return
	}
	ctx, failureCode, status := server.adminProviderRouteContextFromRequest(request, trace, "admin_provider_routes:activate", true, "candidate-activation")
	if failureCode != "" {
		writeAdminProviderRouteResult(writer, trace, ctx, request.PathValue("configuration_id"), request.PathValue("candidate_id"), AdminProviderRouteResult{FailureCode: failureCode}, status)
		return
	}
	var body adminProviderRouteActivationBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, adminProviderRouteJSONRequestBodyOptions()) {
		return
	}
	result := server.adminProviderRouteService().Activate(ctx, AdminProviderRouteActivationInput{
		ConfigurationID: request.PathValue("configuration_id"), CandidateID: request.PathValue("candidate_id"),
		ExpectedGeneration: body.ExpectedGeneration, Action: body.Action, Reason: body.Reason,
	})
	writeAdminProviderRouteResult(writer, trace, ctx, request.PathValue("configuration_id"), request.PathValue("candidate_id"), result, http.StatusOK)
}

func (server *Server) handleReadAdminProviderRouteActiveSnapshot(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, adminProviderRouteActiveSnapshotRoute)
	if !server.prepareAdminProviderRouteHTTP(writer, request, trace) {
		return
	}
	ctx, failureCode, status := server.adminProviderRouteContextFromRequest(request, trace, "admin_provider_routes:read", false, "snapshot-read")
	if failureCode != "" {
		writeAdminProviderRouteResult(writer, trace, ctx, request.PathValue("configuration_id"), "", AdminProviderRouteResult{FailureCode: failureCode}, status)
		return
	}
	result := server.adminProviderRouteService().ReadActiveSnapshot(ctx, request.PathValue("configuration_id"))
	writeAdminProviderRouteResult(writer, trace, ctx, request.PathValue("configuration_id"), "", result, http.StatusOK)
}

func (server *Server) handleListAdminProviderRouteActivations(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, adminProviderRouteActivationHistoryRoute)
	if !server.prepareAdminProviderRouteHTTP(writer, request, trace) {
		return
	}
	ctx, failureCode, status := server.adminProviderRouteContextFromRequest(request, trace, "admin_provider_routes:read", false, "activation-history")
	if failureCode != "" {
		writeAdminProviderRouteResult(writer, trace, ctx, request.PathValue("configuration_id"), "", AdminProviderRouteResult{FailureCode: failureCode}, status)
		return
	}
	result := server.adminProviderRouteService().ListActivations(ctx, request.PathValue("configuration_id"))
	writeAdminProviderRouteResult(writer, trace, ctx, request.PathValue("configuration_id"), "", result, http.StatusOK)
}

func (server *Server) prepareAdminProviderRouteHTTP(
	writer http.ResponseWriter,
	request *http.Request,
	trace requestTrace,
) bool {
	writer.Header().Set("Cache-Control", "no-store")
	if !server.config.AdminProviderRouteDevHTTPEnabled {
		server.writePlatformError(writer, trace, "ADMIN_PROVIDER_ROUTE_DEV_HTTP_DISABLED", "admin provider route management requires explicit development opt-in")
		return false
	}
	if len(request.URL.Query()) != 0 {
		writeAdminProviderRouteResult(
			writer, trace, AdminProviderRouteContext{}, request.PathValue("configuration_id"),
			request.PathValue("candidate_id"), AdminProviderRouteResult{FailureCode: AdminProviderRouteFailurePayloadInvalid},
			http.StatusBadRequest,
		)
		return false
	}
	return true
}

func adminProviderRouteJSONRequestBodyOptions() jsonRequestBodyOptions {
	return jsonRequestBodyOptions{
		maxBytes:              maxControlJSONRequestBodyBytes,
		rejectUnknownFields:   true,
		rejectDuplicateFields: true,
	}
}

func (server *Server) adminProviderRouteContextFromRequest(
	request *http.Request,
	trace requestTrace,
	requiredScope string,
	writeRequired bool,
	auditSuffix string,
) (AdminProviderRouteContext, string, int) {
	ctx := AdminProviderRouteContext{
		RequestContext: request.Context(), RequestID: trace.requestID,
		WorkspaceID: strings.TrimSpace(request.Header.Get(adminProviderRouteDevWorkspaceHeader)),
		Environment: strings.TrimSpace(request.Header.Get(adminProviderRouteDevEnvironmentHeader)),
		AuditRef:    "audit_" + trace.requestID + "_admin-provider-route-" + auditSuffix,
	}
	auth, failureCode, status := authorizeControlPlaneReadRequest(request, "", requiredScope)
	if failureCode != "" {
		return ctx, failureCode, status
	}
	if auth.VerifiedIdentity == nil || !auth.ResourceBinding.TenantVerified ||
		strings.TrimSpace(auth.VerifiedIdentity.SubjectRef) != strings.TrimSpace(auth.SubjectBinding) ||
		strings.TrimSpace(auth.VerifiedIdentity.TenantRef) != strings.TrimSpace(auth.TenantBinding) ||
		strings.TrimSpace(auth.ResourceBinding.TenantRef) != strings.TrimSpace(auth.TenantBinding) {
		return ctx, "auth_context_contract_mismatch", http.StatusUnauthorized
	}
	ctx.TenantRef = strings.TrimSpace(auth.TenantBinding)
	ctx.ActorRef = strings.TrimSpace(auth.SubjectBinding)
	if !adminProviderRouteIdentifierPattern.MatchString(ctx.WorkspaceID) {
		return ctx, AdminProviderRouteFailureScopeDenied, http.StatusForbidden
	}
	if ctx.Environment != adminProviderRouteEnvironmentDevelopment &&
		ctx.Environment != adminProviderRouteEnvironmentTest {
		return ctx, AdminProviderRouteFailureEnvironmentForbidden, http.StatusForbidden
	}
	if writeRequired && !server.config.AdminProviderRouteDevWriteEnabled {
		return ctx, AdminProviderRouteFailureDisabled, http.StatusForbidden
	}
	ctx.ReadEnabled = controlPlaneReadHasScope(auth.ScopeGrants, "admin_provider_routes:read")
	ctx.DraftEnabled = server.config.AdminProviderRouteDevWriteEnabled &&
		controlPlaneReadHasScope(auth.ScopeGrants, "admin_provider_routes:draft")
	ctx.ReviewEnabled = server.config.AdminProviderRouteDevWriteEnabled &&
		controlPlaneReadHasScope(auth.ScopeGrants, "admin_provider_routes:review")
	ctx.ActivateEnabled = server.config.AdminProviderRouteDevWriteEnabled &&
		controlPlaneReadHasScope(auth.ScopeGrants, "admin_provider_routes:activate")
	return ctx, "", http.StatusOK
}

func (server *Server) adminProviderRouteService() adminProviderRouteService {
	if server.adminProviderRouteRepository == nil {
		repository := newMemoryAdminProviderRouteRepository()
		repository.setUnavailableForTest(true)
		server.adminProviderRouteRepository = repository
	}
	return newAdminProviderRouteService(
		server.adminProviderRouteRepository,
		bridgeAdminProviderInventoryResolver{bridge: server.bridge},
	)
}

func writeAdminProviderRouteResult(
	writer http.ResponseWriter,
	trace requestTrace,
	ctx AdminProviderRouteContext,
	configurationID string,
	candidateID string,
	result AdminProviderRouteResult,
	successStatus int,
) {
	history := result.ActivationHistory
	if history == nil {
		history = []AdminProviderRouteActivationRecord{}
	}
	status := successStatus
	if status >= http.StatusOK && status < http.StatusMultipleChoices {
		status = adminProviderRouteHTTPStatus(result.FailureCode, successStatus)
	}
	writeObservedJSON(writer, status, trace, adminProviderRouteEnvelope{
		RequestID: trace.requestID, WorkspaceID: ctx.WorkspaceID, Environment: ctx.Environment,
		ConfigurationID: strings.TrimSpace(configurationID), CandidateID: strings.TrimSpace(candidateID),
		Draft: result.Draft, Candidate: result.Candidate, Snapshot: result.Snapshot,
		Activation: result.Activation, ActivationHistory: history,
		FailureCode:          optionalAdminProviderRouteFailure(result.FailureCode),
		CurrentDraftRevision: result.CurrentDraftRevision, CurrentReviewVersion: result.CurrentReviewVersion,
		CurrentCandidateState: result.CurrentCandidateState, CurrentGeneration: result.CurrentGeneration,
		AuditRef: ctx.AuditRef,
	})
}

func adminProviderRouteHTTPStatus(failureCode string, successStatus int) int {
	switch strings.TrimSpace(failureCode) {
	case "":
		return successStatus
	case "identity_context_missing", "auth_context_contract_mismatch", "tenant_binding_missing":
		return http.StatusUnauthorized
	case AdminProviderRouteFailureDisabled, AdminProviderRouteFailureScopeDenied,
		AdminProviderRouteFailureEnvironmentForbidden, "scope_denied":
		return http.StatusForbidden
	case AdminProviderRouteFailurePayloadInvalid, AdminProviderRouteFailureSensitiveForbidden:
		return http.StatusBadRequest
	case AdminProviderRouteFailureDraftNotFound, AdminProviderRouteFailureCandidateNotFound:
		return http.StatusNotFound
	case AdminProviderRouteFailureDraftRevisionConflict, AdminProviderRouteFailureCandidateConflict,
		AdminProviderRouteFailureReviewVersionConflict, AdminProviderRouteFailureReviewTransitionInvalid,
		AdminProviderRouteFailureGenerationConflict, AdminProviderRouteFailureAlreadyActive:
		return http.StatusConflict
	case AdminProviderRouteFailureCandidateNotApproved, AdminProviderRouteFailureInventoryNotFound,
		AdminProviderRouteFailureInventoryMismatch, AdminProviderRouteFailureRollbackTargetInvalid:
		return http.StatusUnprocessableEntity
	case AdminProviderRouteFailureInventoryUnavailable, AdminProviderRouteFailureStoreUnavailable,
		"identity_provider_unavailable", "workspace_membership_unavailable":
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

func optionalAdminProviderRouteFailure(failureCode string) *string {
	failureCode = strings.TrimSpace(failureCode)
	if failureCode == "" {
		return nil
	}
	return &failureCode
}
