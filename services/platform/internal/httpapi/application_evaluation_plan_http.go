package httpapi

import (
	"net/http"
	"strconv"
	"strings"
)

const (
	applicationEvaluationPlanCreateRoute   = "POST /v1/user-workspace/applications/{application_id}/evaluation-plans"
	applicationEvaluationPlanListRoute     = "GET /v1/user-workspace/applications/{application_id}/evaluation-plans"
	applicationEvaluationPlanReadRoute     = "GET /v1/user-workspace/applications/{application_id}/evaluation-plans/{plan_id}"
	applicationEvaluationPlanReviseRoute   = "POST /v1/user-workspace/applications/{application_id}/evaluation-plans/{plan_id}/revisions"
	applicationEvaluationPlanArchiveRoute  = "POST /v1/user-workspace/applications/{application_id}/evaluation-plans/{plan_id}/archive"
	applicationEvaluationVersionListRoute  = "GET /v1/user-workspace/applications/{application_id}/evaluation-plans/{plan_id}/versions"
	applicationEvaluationVersionReadRoute  = "GET /v1/user-workspace/applications/{application_id}/evaluation-plans/{plan_id}/versions/{version}"
	applicationEvaluationEnvironmentHeader = "X-RadishMind-Dev-Application-Evaluation-Environment"
)

type applicationEvaluationPlanCreateBody struct {
	WorkspaceID      string                          `json:"workspace_id"`
	Environment      string                          `json:"environment"`
	Name             string                          `json:"name"`
	ExecutionProfile string                          `json:"execution_profile"`
	Target           ApplicationEvaluationPlanTarget `json:"target"`
	Items            []ApplicationEvaluationPlanItem `json:"items"`
}

type applicationEvaluationPlanReviseBody struct {
	WorkspaceID      string                          `json:"workspace_id"`
	Environment      string                          `json:"environment"`
	ExpectedVersion  int                             `json:"expected_version"`
	Name             string                          `json:"name"`
	ExecutionProfile string                          `json:"execution_profile"`
	Target           ApplicationEvaluationPlanTarget `json:"target"`
	Items            []ApplicationEvaluationPlanItem `json:"items"`
}

type applicationEvaluationPlanArchiveBody struct {
	WorkspaceID               string `json:"workspace_id"`
	Environment               string `json:"environment"`
	ExpectedVersion           int    `json:"expected_version"`
	AcknowledgeNoNewCampaigns bool   `json:"acknowledge_no_new_campaigns"`
}

type applicationEvaluationPlanEnvelope struct {
	RequestID            string                            `json:"request_id"`
	TenantRef            string                            `json:"tenant_ref"`
	WorkspaceID          string                            `json:"workspace_id"`
	Environment          string                            `json:"environment"`
	ApplicationID        string                            `json:"application_id"`
	Plan                 *ApplicationEvaluationPlan        `json:"plan"`
	Version              *ApplicationEvaluationPlanVersion `json:"version"`
	FailureCode          *string                           `json:"failure_code"`
	FailureSummary       string                            `json:"failure_summary"`
	CurrentRecordVersion int                               `json:"current_record_version"`
	CurrentState         string                            `json:"current_state"`
	AuditRef             string                            `json:"audit_ref"`
}

type applicationEvaluationPlanListEnvelope struct {
	RequestID      string                      `json:"request_id"`
	TenantRef      string                      `json:"tenant_ref"`
	WorkspaceID    string                      `json:"workspace_id"`
	Environment    string                      `json:"environment"`
	ApplicationID  string                      `json:"application_id"`
	Plans          []ApplicationEvaluationPlan `json:"plans"`
	NextCursor     string                      `json:"next_cursor"`
	HasMore        bool                        `json:"has_more"`
	FailureCode    *string                     `json:"failure_code"`
	FailureSummary string                      `json:"failure_summary"`
	AuditRef       string                      `json:"audit_ref"`
}

type applicationEvaluationVersionListEnvelope struct {
	RequestID      string                             `json:"request_id"`
	TenantRef      string                             `json:"tenant_ref"`
	WorkspaceID    string                             `json:"workspace_id"`
	Environment    string                             `json:"environment"`
	ApplicationID  string                             `json:"application_id"`
	Versions       []ApplicationEvaluationPlanVersion `json:"versions"`
	NextCursor     string                             `json:"next_cursor"`
	HasMore        bool                               `json:"has_more"`
	FailureCode    *string                            `json:"failure_code"`
	FailureSummary string                             `json:"failure_summary"`
	AuditRef       string                             `json:"audit_ref"`
}

func (server *Server) handleCreateApplicationEvaluationPlan(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, applicationEvaluationPlanCreateRoute)
	if !server.allowApplicationEvaluationCampaignDev(writer, trace) {
		return
	}
	auth, failure, status := server.authorizeWorkspaceScopedPermissions(request, "application_evaluations:write")
	ctx := applicationEvaluationMutationContext(request, trace, auth, request.PathValue("application_id"), "plan-create")
	if failure != "" {
		writeApplicationEvaluationPlanResult(writer, status, trace, ctx, applicationEvaluationPlanFailure(failure))
		return
	}
	var body applicationEvaluationPlanCreateBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: agentCopilotMaximumInvocationBytes*applicationEvaluationMaximumItems + maxControlJSONRequestBodyBytes, rejectUnknownFields: true}) {
		return
	}
	ctx.Environment = strings.TrimSpace(body.Environment)
	if !server.applicationEvaluationBindingMatches(request, auth, body.WorkspaceID, ctx.Environment, ctx.ApplicationID) {
		writeApplicationEvaluationPlanResult(writer, http.StatusForbidden, trace, ctx, applicationEvaluationPlanFailure(ApplicationEvaluationFailureScopeDenied))
		return
	}
	result := server.applicationEvaluationPlanService().Create(ctx, ApplicationEvaluationPlanCreateInput{
		Name: body.Name, ExecutionProfile: body.ExecutionProfile, Target: body.Target, Items: body.Items,
	})
	writeApplicationEvaluationPlanResult(writer, applicationEvaluationHTTPStatus(result.FailureCode), trace, ctx, result)
}

func (server *Server) handleReviseApplicationEvaluationPlan(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, applicationEvaluationPlanReviseRoute)
	if !server.allowApplicationEvaluationCampaignDev(writer, trace) {
		return
	}
	auth, failure, status := server.authorizeWorkspaceScopedPermissions(request, "application_evaluations:write")
	ctx := applicationEvaluationMutationContext(request, trace, auth, request.PathValue("application_id"), "plan-revise")
	if failure != "" {
		writeApplicationEvaluationPlanResult(writer, status, trace, ctx, applicationEvaluationPlanFailure(failure))
		return
	}
	var body applicationEvaluationPlanReviseBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: agentCopilotMaximumInvocationBytes*applicationEvaluationMaximumItems + maxControlJSONRequestBodyBytes, rejectUnknownFields: true}) {
		return
	}
	ctx.Environment = strings.TrimSpace(body.Environment)
	if !server.applicationEvaluationBindingMatches(request, auth, body.WorkspaceID, ctx.Environment, ctx.ApplicationID) {
		writeApplicationEvaluationPlanResult(writer, http.StatusForbidden, trace, ctx, applicationEvaluationPlanFailure(ApplicationEvaluationFailureScopeDenied))
		return
	}
	result := server.applicationEvaluationPlanService().Revise(ctx, request.PathValue("plan_id"), ApplicationEvaluationPlanReviseInput{
		ExpectedVersion: body.ExpectedVersion, Name: body.Name, ExecutionProfile: body.ExecutionProfile, Target: body.Target, Items: body.Items,
	})
	writeApplicationEvaluationPlanResult(writer, applicationEvaluationHTTPStatus(result.FailureCode), trace, ctx, result)
}

func (server *Server) handleArchiveApplicationEvaluationPlan(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, applicationEvaluationPlanArchiveRoute)
	if !server.allowApplicationEvaluationCampaignDev(writer, trace) {
		return
	}
	auth, failure, status := server.authorizeWorkspaceScopedPermissions(request, "application_evaluations:write")
	ctx := applicationEvaluationMutationContext(request, trace, auth, request.PathValue("application_id"), "plan-archive")
	if failure != "" {
		writeApplicationEvaluationPlanResult(writer, status, trace, ctx, applicationEvaluationPlanFailure(failure))
		return
	}
	var body applicationEvaluationPlanArchiveBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: maxControlJSONRequestBodyBytes, rejectUnknownFields: true}) {
		return
	}
	ctx.Environment = strings.TrimSpace(body.Environment)
	if !server.applicationEvaluationBindingMatches(request, auth, body.WorkspaceID, ctx.Environment, ctx.ApplicationID) {
		writeApplicationEvaluationPlanResult(writer, http.StatusForbidden, trace, ctx, applicationEvaluationPlanFailure(ApplicationEvaluationFailureScopeDenied))
		return
	}
	result := server.applicationEvaluationPlanService().Archive(ctx, request.PathValue("plan_id"), ApplicationEvaluationPlanArchiveInput{
		ExpectedVersion: body.ExpectedVersion, AcknowledgeNoNewCampaigns: body.AcknowledgeNoNewCampaigns,
	})
	writeApplicationEvaluationPlanResult(writer, applicationEvaluationHTTPStatus(result.FailureCode), trace, ctx, result)
}

func (server *Server) handleReadApplicationEvaluationPlan(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, applicationEvaluationPlanReadRoute)
	ctx, status, failure := server.applicationEvaluationReadContext(request, trace, "plan-read")
	if failure != "" {
		writeApplicationEvaluationPlanResult(writer, status, trace, ctx, applicationEvaluationPlanFailure(failure))
		return
	}
	result := server.applicationEvaluationPlanService().Read(ctx, request.PathValue("plan_id"))
	writeApplicationEvaluationPlanResult(writer, applicationEvaluationHTTPStatus(result.FailureCode), trace, ctx, result)
}

func (server *Server) handleReadApplicationEvaluationPlanVersion(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, applicationEvaluationVersionReadRoute)
	ctx, status, failure := server.applicationEvaluationReadContext(request, trace, "version-read")
	version, err := strconv.Atoi(request.PathValue("version"))
	if err != nil || version < 1 {
		failure, status = ApplicationEvaluationFailurePayloadInvalid, http.StatusBadRequest
	}
	if failure != "" {
		writeApplicationEvaluationPlanResult(writer, status, trace, ctx, applicationEvaluationPlanFailure(failure))
		return
	}
	result := server.applicationEvaluationPlanService().ReadVersion(ctx, request.PathValue("plan_id"), version)
	writeApplicationEvaluationPlanResult(writer, applicationEvaluationHTTPStatus(result.FailureCode), trace, ctx, result)
}

func (server *Server) handleListApplicationEvaluationPlans(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, applicationEvaluationPlanListRoute)
	ctx, status, failure := server.applicationEvaluationReadContext(request, trace, "plan-list", "lifecycle_state", "limit", "cursor")
	if failure != "" {
		writeApplicationEvaluationPlanListResult(writer, status, trace, ctx, applicationEvaluationPlanListFailure(failure))
		return
	}
	limit, ok := applicationEvaluationQueryLimit(request)
	if !ok {
		writeApplicationEvaluationPlanListResult(writer, http.StatusBadRequest, trace, ctx, applicationEvaluationPlanListFailure(ApplicationEvaluationFailurePayloadInvalid))
		return
	}
	result := server.applicationEvaluationPlanService().List(ctx, ApplicationEvaluationPlanListInput{
		LifecycleState: request.URL.Query().Get("lifecycle_state"), Limit: limit, Cursor: request.URL.Query().Get("cursor"),
	})
	writeApplicationEvaluationPlanListResult(writer, applicationEvaluationHTTPStatus(result.FailureCode), trace, ctx, result)
}

func (server *Server) handleListApplicationEvaluationPlanVersions(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, applicationEvaluationVersionListRoute)
	ctx, status, failure := server.applicationEvaluationReadContext(request, trace, "version-list", "limit", "cursor")
	if failure != "" {
		writeApplicationEvaluationVersionListResult(writer, status, trace, ctx, applicationEvaluationVersionListFailure(failure))
		return
	}
	limit, ok := applicationEvaluationQueryLimit(request)
	if !ok {
		writeApplicationEvaluationVersionListResult(writer, http.StatusBadRequest, trace, ctx, applicationEvaluationVersionListFailure(ApplicationEvaluationFailurePayloadInvalid))
		return
	}
	result := server.applicationEvaluationPlanService().ListVersions(ctx, request.PathValue("plan_id"), ApplicationEvaluationVersionListInput{
		Limit: limit, Cursor: request.URL.Query().Get("cursor"),
	})
	writeApplicationEvaluationVersionListResult(writer, applicationEvaluationHTTPStatus(result.FailureCode), trace, ctx, result)
}

func (server *Server) applicationEvaluationReadContext(request *http.Request, trace requestTrace, suffix string, optionalQueryKeys ...string) (ApplicationEvaluationContext, int, string) {
	ctx := ApplicationEvaluationContext{
		RequestContext: request.Context(), RequestID: trace.requestID, ApplicationID: strings.TrimSpace(request.PathValue("application_id")),
		AuditRef: "audit_" + trace.requestID + "_application-evaluation-" + suffix,
	}
	if !server.allowApplicationEvaluationCampaignDev(nil, trace) {
		return ctx, http.StatusServiceUnavailable, ApplicationEvaluationFailureWriteDisabled
	}
	auth, failure, status := server.authorizeWorkspaceScopedPermissions(request, "application_evaluations:read")
	ctx.TenantRef, ctx.WorkspaceID, ctx.ActorRef = strings.TrimSpace(auth.TenantBinding), strings.TrimSpace(auth.ResourceBinding.WorkspaceID), strings.TrimSpace(auth.SubjectBinding)
	if failure != "" {
		return ctx, status, failure
	}
	allowed := map[string]bool{"workspace_id": true, "environment": true}
	for _, key := range optionalQueryKeys {
		allowed[key] = true
	}
	for key, values := range request.URL.Query() {
		if !allowed[key] || len(values) != 1 {
			return ctx, http.StatusBadRequest, ApplicationEvaluationFailurePayloadInvalid
		}
	}
	ctx.Environment = strings.TrimSpace(request.URL.Query().Get("environment"))
	workspaceID := strings.TrimSpace(request.URL.Query().Get("workspace_id"))
	if !server.applicationEvaluationBindingMatches(request, auth, workspaceID, ctx.Environment, ctx.ApplicationID) {
		return ctx, http.StatusForbidden, ApplicationEvaluationFailureScopeDenied
	}
	return ctx, http.StatusOK, ""
}

func applicationEvaluationMutationContext(request *http.Request, trace requestTrace, auth controlPlaneReadAuthContext, applicationID, suffix string) ApplicationEvaluationContext {
	return ApplicationEvaluationContext{
		RequestContext: request.Context(), RequestID: trace.requestID,
		TenantRef: strings.TrimSpace(auth.TenantBinding), WorkspaceID: strings.TrimSpace(auth.ResourceBinding.WorkspaceID),
		ApplicationID: strings.TrimSpace(applicationID), ActorRef: strings.TrimSpace(auth.SubjectBinding),
		AuditRef: "audit_" + trace.requestID + "_application-evaluation-" + suffix, WriteEnabled: true,
	}
}

func (server *Server) applicationEvaluationBindingMatches(request *http.Request, auth controlPlaneReadAuthContext, workspaceID, environment, applicationID string) bool {
	configuredEnvironment := strings.TrimSpace(server.config.ApplicationEvaluationCampaignEnvironment)
	return strings.TrimSpace(workspaceID) == strings.TrimSpace(auth.ResourceBinding.WorkspaceID) &&
		strings.TrimSpace(request.Header.Get(savedWorkflowDraftDevWorkspaceHeader)) == strings.TrimSpace(auth.ResourceBinding.WorkspaceID) &&
		strings.TrimSpace(request.Header.Get(savedWorkflowDraftDevApplicationHeader)) == strings.TrimSpace(applicationID) &&
		strings.TrimSpace(request.Header.Get(applicationEvaluationEnvironmentHeader)) == configuredEnvironment &&
		strings.TrimSpace(environment) == configuredEnvironment && validApplicationEvaluationEnvironment(environment) &&
		validControlPlaneReadAuthReference(strings.TrimSpace(applicationID), false)
}

func (server *Server) allowApplicationEvaluationCampaignDev(writer http.ResponseWriter, trace requestTrace) bool {
	if server.config.ApplicationEvaluationCampaignDevEnabled {
		return true
	}
	if writer != nil {
		server.writePlatformError(writer, trace, "APPLICATION_EVALUATION_CAMPAIGN_DEV_DISABLED", "Application evaluation campaign requires explicit development opt-in")
	}
	return false
}

func (server *Server) applicationEvaluationPlanService() applicationEvaluationPlanService {
	return newApplicationEvaluationPlanService(server.applicationEvaluationRepository, server.applicationCatalogRepository)
}

func applicationEvaluationQueryLimit(request *http.Request) (int, bool) {
	raw := strings.TrimSpace(request.URL.Query().Get("limit"))
	if raw == "" {
		return 0, true
	}
	value, err := strconv.Atoi(raw)
	return value, err == nil
}

func applicationEvaluationFailurePointer(code string) *string {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil
	}
	return &code
}

func applicationEvaluationHTTPStatus(code string) int {
	switch code {
	case "":
		return http.StatusOK
	case ApplicationEvaluationFailurePayloadInvalid, ApplicationEvaluationFailureCursorInvalid:
		return http.StatusBadRequest
	case ApplicationEvaluationFailureScopeDenied, ApplicationEvaluationFailureEnvironmentDenied,
		ApplicationEvaluationFailureSecretForbidden, ApplicationEvaluationFailureProfileIneligible,
		ApplicationEvaluationFailureQuotaConsumerInvalid:
		return http.StatusForbidden
	case ApplicationEvaluationFailureNotFound:
		return http.StatusNotFound
	case ApplicationEvaluationFailureVersionConflict, ApplicationEvaluationFailureArchived, ApplicationEvaluationFailureCampaignConflict:
		return http.StatusConflict
	case ApplicationEvaluationFailureAuthorityChanged, ApplicationEvaluationFailureHandoffPartial:
		return http.StatusConflict
	case GatewayRequestQuotaFailureExceeded, GatewayRequestQuotaFailureAttemptConflict,
		GatewayRequestQuotaFailurePolicyVersionConflict, GatewayRequestQuotaFailureDisabled,
		GatewayRequestQuotaFailureScopeDenied, GatewayRequestQuotaFailureEnvironmentForbidden,
		GatewayRequestQuotaFailurePolicyNotFound, GatewayRequestQuotaFailureStoreUnavailable:
		return gatewayRequestQuotaHTTPStatus(code)
	default:
		return http.StatusServiceUnavailable
	}
}

func writeApplicationEvaluationPlanResult(writer http.ResponseWriter, status int, trace requestTrace, ctx ApplicationEvaluationContext, result ApplicationEvaluationPlanResult) {
	writeObservedJSON(writer, status, trace, applicationEvaluationPlanEnvelope{
		RequestID: trace.requestID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, Environment: ctx.Environment,
		ApplicationID: ctx.ApplicationID, Plan: result.Plan, Version: result.Version,
		FailureCode: applicationEvaluationFailurePointer(result.FailureCode), FailureSummary: result.FailureSummary,
		CurrentRecordVersion: result.CurrentRecordVersion, CurrentState: result.CurrentState, AuditRef: ctx.AuditRef,
	})
}

func writeApplicationEvaluationPlanListResult(writer http.ResponseWriter, status int, trace requestTrace, ctx ApplicationEvaluationContext, result ApplicationEvaluationPlanListResult) {
	if result.Plans == nil {
		result.Plans = []ApplicationEvaluationPlan{}
	}
	writeObservedJSON(writer, status, trace, applicationEvaluationPlanListEnvelope{
		RequestID: trace.requestID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, Environment: ctx.Environment,
		ApplicationID: ctx.ApplicationID, Plans: result.Plans, NextCursor: result.NextCursor, HasMore: result.HasMore,
		FailureCode: applicationEvaluationFailurePointer(result.FailureCode), FailureSummary: result.FailureSummary, AuditRef: ctx.AuditRef,
	})
}

func writeApplicationEvaluationVersionListResult(writer http.ResponseWriter, status int, trace requestTrace, ctx ApplicationEvaluationContext, result ApplicationEvaluationVersionListResult) {
	if result.Versions == nil {
		result.Versions = []ApplicationEvaluationPlanVersion{}
	}
	writeObservedJSON(writer, status, trace, applicationEvaluationVersionListEnvelope{
		RequestID: trace.requestID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, Environment: ctx.Environment,
		ApplicationID: ctx.ApplicationID, Versions: result.Versions, NextCursor: result.NextCursor, HasMore: result.HasMore,
		FailureCode: applicationEvaluationFailurePointer(result.FailureCode), FailureSummary: result.FailureSummary, AuditRef: ctx.AuditRef,
	})
}
