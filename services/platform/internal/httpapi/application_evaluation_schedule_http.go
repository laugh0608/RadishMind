package httpapi

import (
	"net/http"
	"strconv"
	"strings"
)

const (
	applicationEvaluationScheduleCreateRoute         = "POST /v1/user-workspace/applications/{application_id}/evaluation-schedules"
	applicationEvaluationScheduleListRoute           = "GET /v1/user-workspace/applications/{application_id}/evaluation-schedules"
	applicationEvaluationScheduleReadRoute           = "GET /v1/user-workspace/applications/{application_id}/evaluation-schedules/{schedule_id}"
	applicationEvaluationScheduleReviseRoute         = "POST /v1/user-workspace/applications/{application_id}/evaluation-schedules/{schedule_id}/revisions"
	applicationEvaluationScheduleActivateRoute       = "POST /v1/user-workspace/applications/{application_id}/evaluation-schedules/{schedule_id}/activate"
	applicationEvaluationSchedulePauseRoute          = "POST /v1/user-workspace/applications/{application_id}/evaluation-schedules/{schedule_id}/pause"
	applicationEvaluationScheduleResumeRoute         = "POST /v1/user-workspace/applications/{application_id}/evaluation-schedules/{schedule_id}/resume"
	applicationEvaluationScheduleArchiveRoute        = "POST /v1/user-workspace/applications/{application_id}/evaluation-schedules/{schedule_id}/archive"
	applicationEvaluationScheduleVersionReadRoute    = "GET /v1/user-workspace/applications/{application_id}/evaluation-schedules/{schedule_id}/versions/{version}"
	applicationEvaluationScheduleOccurrenceReadRoute = "GET /v1/user-workspace/applications/{application_id}/evaluation-schedules/{schedule_id}/occurrences/{schedule_version}/{scheduled_for_utc}"
)

type applicationEvaluationScheduleCreateBody struct {
	WorkspaceID                    string                                `json:"workspace_id"`
	Environment                    string                                `json:"environment"`
	PlanID                         string                                `json:"plan_id"`
	PlanVersion                    int                                   `json:"plan_version"`
	PlanDigest                     string                                `json:"plan_digest"`
	ExpectedPlanRecordVersion      int                                   `json:"expected_plan_record_version"`
	QuotaAPIKeyID                  string                                `json:"quota_api_key_id"`
	Schedule                       ApplicationEvaluationScheduleDailyUTC `json:"schedule"`
	AcknowledgeProviderConsumption bool                                  `json:"acknowledge_provider_consumption"`
}

type applicationEvaluationScheduleReviseBody struct {
	WorkspaceID                    string                                `json:"workspace_id"`
	Environment                    string                                `json:"environment"`
	ExpectedVersion                int                                   `json:"expected_version"`
	PlanID                         string                                `json:"plan_id"`
	PlanVersion                    int                                   `json:"plan_version"`
	PlanDigest                     string                                `json:"plan_digest"`
	ExpectedPlanRecordVersion      int                                   `json:"expected_plan_record_version"`
	QuotaAPIKeyID                  string                                `json:"quota_api_key_id"`
	Schedule                       ApplicationEvaluationScheduleDailyUTC `json:"schedule"`
	AcknowledgeProviderConsumption bool                                  `json:"acknowledge_provider_consumption"`
}

type applicationEvaluationScheduleLifecycleBody struct {
	WorkspaceID                    string `json:"workspace_id"`
	Environment                    string `json:"environment"`
	ExpectedVersion                int    `json:"expected_version"`
	AcknowledgeProviderConsumption bool   `json:"acknowledge_provider_consumption"`
	AcknowledgeNoFutureOccurrences bool   `json:"acknowledge_no_future_occurrences"`
}

type applicationEvaluationScheduleEnvelope struct {
	RequestID            string                                   `json:"request_id"`
	TenantRef            string                                   `json:"tenant_ref"`
	WorkspaceID          string                                   `json:"workspace_id"`
	Environment          string                                   `json:"environment"`
	ApplicationID        string                                   `json:"application_id"`
	Schedule             *ApplicationEvaluationSchedule           `json:"schedule"`
	Version              *ApplicationEvaluationScheduleVersion    `json:"version"`
	Occurrence           *ApplicationEvaluationScheduleOccurrence `json:"occurrence"`
	FailureCode          *string                                  `json:"failure_code"`
	FailureSummary       string                                   `json:"failure_summary"`
	CurrentRecordVersion int                                      `json:"current_record_version"`
	CurrentState         string                                   `json:"current_state"`
	AuditRef             string                                   `json:"audit_ref"`
}

type applicationEvaluationScheduleListEnvelope struct {
	RequestID      string                          `json:"request_id"`
	TenantRef      string                          `json:"tenant_ref"`
	WorkspaceID    string                          `json:"workspace_id"`
	Environment    string                          `json:"environment"`
	ApplicationID  string                          `json:"application_id"`
	Schedules      []ApplicationEvaluationSchedule `json:"schedules"`
	NextCursor     string                          `json:"next_cursor"`
	HasMore        bool                            `json:"has_more"`
	FailureCode    *string                         `json:"failure_code"`
	FailureSummary string                          `json:"failure_summary"`
	AuditRef       string                          `json:"audit_ref"`
}

func (server *Server) handleCreateApplicationEvaluationSchedule(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, applicationEvaluationScheduleCreateRoute)
	ctx, body, ok := server.applicationEvaluationScheduleMutationRequest(writer, request, trace, "schedule-create")
	if !ok {
		return
	}
	result := server.applicationEvaluationScheduleService().Create(ctx, ApplicationEvaluationScheduleCreateInput{
		ApplicationEvaluationScheduleDefinitionInput: applicationEvaluationScheduleDefinitionFromCreateBody(body),
	})
	writeApplicationEvaluationScheduleResult(writer, applicationEvaluationScheduleHTTPStatus(result.FailureCode), trace, ctx, result)
}

func (server *Server) handleReviseApplicationEvaluationSchedule(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, applicationEvaluationScheduleReviseRoute)
	if !server.allowApplicationEvaluationCampaignDev(writer, trace) {
		return
	}
	auth, failure, status := server.authorizeWorkspaceScopedPermissions(request, applicationEvaluationScheduleRequiredPermissions...)
	ctx := applicationEvaluationMutationContext(request, trace, auth, request.PathValue("application_id"), "schedule-revise")
	if failure != "" {
		writeApplicationEvaluationScheduleResult(writer, status, trace, ctx, applicationEvaluationScheduleFailure(failure))
		return
	}
	var body applicationEvaluationScheduleReviseBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: maxControlJSONRequestBodyBytes, rejectUnknownFields: true, rejectDuplicateFields: true}) {
		return
	}
	ctx.Environment = strings.TrimSpace(body.Environment)
	if !server.applicationEvaluationBindingMatches(request, auth, body.WorkspaceID, ctx.Environment, ctx.ApplicationID) {
		writeApplicationEvaluationScheduleResult(writer, http.StatusForbidden, trace, ctx, applicationEvaluationScheduleFailure(ApplicationEvaluationFailureScopeDenied))
		return
	}
	result := server.applicationEvaluationScheduleService().Revise(ctx, request.PathValue("schedule_id"), ApplicationEvaluationScheduleReviseInput{
		ExpectedVersion: body.ExpectedVersion,
		ApplicationEvaluationScheduleDefinitionInput: ApplicationEvaluationScheduleDefinitionInput{
			PlanID: body.PlanID, PlanVersion: body.PlanVersion, PlanDigest: body.PlanDigest,
			ExpectedPlanRecordVersion: body.ExpectedPlanRecordVersion, QuotaAPIKeyID: body.QuotaAPIKeyID,
			Schedule: body.Schedule, AcknowledgeProviderConsumption: body.AcknowledgeProviderConsumption,
		},
	})
	writeApplicationEvaluationScheduleResult(writer, applicationEvaluationScheduleHTTPStatus(result.FailureCode), trace, ctx, result)
}

func (server *Server) handleActivateApplicationEvaluationSchedule(writer http.ResponseWriter, request *http.Request) {
	server.handleApplicationEvaluationScheduleLifecycle(writer, request, applicationEvaluationScheduleActivateRoute, "schedule-activate", "activate")
}

func (server *Server) handlePauseApplicationEvaluationSchedule(writer http.ResponseWriter, request *http.Request) {
	server.handleApplicationEvaluationScheduleLifecycle(writer, request, applicationEvaluationSchedulePauseRoute, "schedule-pause", "pause")
}

func (server *Server) handleResumeApplicationEvaluationSchedule(writer http.ResponseWriter, request *http.Request) {
	server.handleApplicationEvaluationScheduleLifecycle(writer, request, applicationEvaluationScheduleResumeRoute, "schedule-resume", "resume")
}

func (server *Server) handleArchiveApplicationEvaluationSchedule(writer http.ResponseWriter, request *http.Request) {
	server.handleApplicationEvaluationScheduleLifecycle(writer, request, applicationEvaluationScheduleArchiveRoute, "schedule-archive", "archive")
}

func (server *Server) handleApplicationEvaluationScheduleLifecycle(
	writer http.ResponseWriter,
	request *http.Request,
	route string,
	auditSuffix string,
	action string,
) {
	trace := newRequestTrace(request, route)
	if !server.allowApplicationEvaluationCampaignDev(writer, trace) {
		return
	}
	auth, failure, status := server.authorizeWorkspaceScopedPermissions(request, applicationEvaluationScheduleRequiredPermissions...)
	ctx := applicationEvaluationMutationContext(request, trace, auth, request.PathValue("application_id"), auditSuffix)
	if failure != "" {
		writeApplicationEvaluationScheduleResult(writer, status, trace, ctx, applicationEvaluationScheduleFailure(failure))
		return
	}
	var body applicationEvaluationScheduleLifecycleBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: maxControlJSONRequestBodyBytes, rejectUnknownFields: true, rejectDuplicateFields: true}) {
		return
	}
	ctx.Environment = strings.TrimSpace(body.Environment)
	if !server.applicationEvaluationBindingMatches(request, auth, body.WorkspaceID, ctx.Environment, ctx.ApplicationID) {
		writeApplicationEvaluationScheduleResult(writer, http.StatusForbidden, trace, ctx, applicationEvaluationScheduleFailure(ApplicationEvaluationFailureScopeDenied))
		return
	}
	input := ApplicationEvaluationScheduleLifecycleInput{
		ExpectedVersion: body.ExpectedVersion, AcknowledgeProviderConsumption: body.AcknowledgeProviderConsumption,
		AcknowledgeNoFutureOccurrences: body.AcknowledgeNoFutureOccurrences,
	}
	service := server.applicationEvaluationScheduleService()
	var result ApplicationEvaluationScheduleResult
	switch action {
	case "activate":
		result = service.Activate(ctx, request.PathValue("schedule_id"), input)
	case "pause":
		result = service.Pause(ctx, request.PathValue("schedule_id"), input)
	case "resume":
		result = service.Resume(ctx, request.PathValue("schedule_id"), input)
	case "archive":
		result = service.Archive(ctx, request.PathValue("schedule_id"), input)
	default:
		result = applicationEvaluationScheduleFailure(ApplicationEvaluationFailurePayloadInvalid)
	}
	writeApplicationEvaluationScheduleResult(writer, applicationEvaluationScheduleHTTPStatus(result.FailureCode), trace, ctx, result)
}

func (server *Server) handleReadApplicationEvaluationSchedule(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, applicationEvaluationScheduleReadRoute)
	ctx, status, failure := server.applicationEvaluationReadContext(request, trace, "schedule-read")
	if failure != "" {
		writeApplicationEvaluationScheduleResult(writer, status, trace, ctx, applicationEvaluationScheduleFailure(failure))
		return
	}
	result := server.applicationEvaluationScheduleService().Read(ctx, request.PathValue("schedule_id"))
	writeApplicationEvaluationScheduleResult(writer, applicationEvaluationScheduleHTTPStatus(result.FailureCode), trace, ctx, result)
}

func (server *Server) handleListApplicationEvaluationSchedules(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, applicationEvaluationScheduleListRoute)
	ctx, status, failure := server.applicationEvaluationReadContext(request, trace, "schedule-list", "lifecycle_state", "limit", "cursor")
	if failure != "" {
		writeApplicationEvaluationScheduleListResult(writer, status, trace, ctx, applicationEvaluationScheduleListFailure(failure))
		return
	}
	limit, ok := applicationEvaluationQueryLimit(request)
	if !ok {
		writeApplicationEvaluationScheduleListResult(writer, http.StatusBadRequest, trace, ctx, applicationEvaluationScheduleListFailure(ApplicationEvaluationFailurePayloadInvalid))
		return
	}
	result := server.applicationEvaluationScheduleService().List(ctx, ApplicationEvaluationScheduleListInput{
		LifecycleState: request.URL.Query().Get("lifecycle_state"), Limit: limit, Cursor: request.URL.Query().Get("cursor"),
	})
	writeApplicationEvaluationScheduleListResult(writer, applicationEvaluationScheduleHTTPStatus(result.FailureCode), trace, ctx, result)
}

func (server *Server) handleReadApplicationEvaluationScheduleVersion(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, applicationEvaluationScheduleVersionReadRoute)
	ctx, status, failure := server.applicationEvaluationReadContext(request, trace, "schedule-version-read")
	version, err := strconv.Atoi(request.PathValue("version"))
	if err != nil || version < 1 {
		failure, status = ApplicationEvaluationFailurePayloadInvalid, http.StatusBadRequest
	}
	if failure != "" {
		writeApplicationEvaluationScheduleResult(writer, status, trace, ctx, applicationEvaluationScheduleFailure(failure))
		return
	}
	result := server.applicationEvaluationScheduleService().ReadVersion(ctx, request.PathValue("schedule_id"), version)
	writeApplicationEvaluationScheduleResult(writer, applicationEvaluationScheduleHTTPStatus(result.FailureCode), trace, ctx, result)
}

func (server *Server) handleReadApplicationEvaluationScheduleOccurrence(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, applicationEvaluationScheduleOccurrenceReadRoute)
	ctx, status, failure := server.applicationEvaluationReadContext(request, trace, "schedule-occurrence-read")
	version, err := strconv.Atoi(request.PathValue("schedule_version"))
	if err != nil || version < 1 {
		failure, status = ApplicationEvaluationFailurePayloadInvalid, http.StatusBadRequest
	}
	if failure != "" {
		writeApplicationEvaluationScheduleResult(writer, status, trace, ctx, applicationEvaluationScheduleFailure(failure))
		return
	}
	result := server.applicationEvaluationScheduleService().ReadOccurrence(ctx, request.PathValue("schedule_id"), version, request.PathValue("scheduled_for_utc"))
	writeApplicationEvaluationScheduleResult(writer, applicationEvaluationScheduleHTTPStatus(result.FailureCode), trace, ctx, result)
}

func (server *Server) applicationEvaluationScheduleMutationRequest(
	writer http.ResponseWriter,
	request *http.Request,
	trace requestTrace,
	suffix string,
) (ApplicationEvaluationContext, applicationEvaluationScheduleCreateBody, bool) {
	if !server.allowApplicationEvaluationCampaignDev(writer, trace) {
		return ApplicationEvaluationContext{}, applicationEvaluationScheduleCreateBody{}, false
	}
	auth, failure, status := server.authorizeWorkspaceScopedPermissions(request, applicationEvaluationScheduleRequiredPermissions...)
	ctx := applicationEvaluationMutationContext(request, trace, auth, request.PathValue("application_id"), suffix)
	if failure != "" {
		writeApplicationEvaluationScheduleResult(writer, status, trace, ctx, applicationEvaluationScheduleFailure(failure))
		return ctx, applicationEvaluationScheduleCreateBody{}, false
	}
	var body applicationEvaluationScheduleCreateBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: maxControlJSONRequestBodyBytes, rejectUnknownFields: true, rejectDuplicateFields: true}) {
		return ctx, body, false
	}
	ctx.Environment = strings.TrimSpace(body.Environment)
	if !server.applicationEvaluationBindingMatches(request, auth, body.WorkspaceID, ctx.Environment, ctx.ApplicationID) {
		writeApplicationEvaluationScheduleResult(writer, http.StatusForbidden, trace, ctx, applicationEvaluationScheduleFailure(ApplicationEvaluationFailureScopeDenied))
		return ctx, body, false
	}
	return ctx, body, true
}

func applicationEvaluationScheduleDefinitionFromCreateBody(body applicationEvaluationScheduleCreateBody) ApplicationEvaluationScheduleDefinitionInput {
	return ApplicationEvaluationScheduleDefinitionInput{
		PlanID: body.PlanID, PlanVersion: body.PlanVersion, PlanDigest: body.PlanDigest,
		ExpectedPlanRecordVersion: body.ExpectedPlanRecordVersion, QuotaAPIKeyID: body.QuotaAPIKeyID,
		Schedule: body.Schedule, AcknowledgeProviderConsumption: body.AcknowledgeProviderConsumption,
	}
}

func (server *Server) applicationEvaluationScheduleService() applicationEvaluationScheduleService {
	return newApplicationEvaluationScheduleService(
		server.applicationEvaluationScheduleRepository,
		server.applicationEvaluationRepository,
		func(ctx ApplicationEvaluationContext, version ApplicationEvaluationPlanVersion) string {
			if _, failure := server.resolveApplicationEvaluationCampaignAuthority(ctx, version); failure != "" {
				if failure == ApplicationEvaluationFailureStoreUnavailable {
					return ApplicationEvaluationScheduleFailureAuthorizationUnavailable
				}
				return ApplicationEvaluationScheduleFailureAuthorityChanged
			}
			return ""
		},
		func(ctx ApplicationEvaluationContext, apiKeyID string) string {
			switch failure := server.validateApplicationEvaluationQuotaConsumer(ctx, apiKeyID); failure {
			case "":
				return ""
			case ApplicationEvaluationFailureStoreUnavailable:
				return ApplicationEvaluationScheduleFailureAuthorizationUnavailable
			default:
				return ApplicationEvaluationScheduleFailureQuotaConsumerInvalid
			}
		},
	)
}

func applicationEvaluationScheduleHTTPStatus(code string) int {
	switch code {
	case ApplicationEvaluationScheduleFailureNotFound:
		return http.StatusNotFound
	case ApplicationEvaluationScheduleFailureVersionConflict, ApplicationEvaluationScheduleFailurePlanChanged,
		ApplicationEvaluationScheduleFailureAuthorityChanged, ApplicationEvaluationScheduleFailureClaimConflict:
		return http.StatusConflict
	case ApplicationEvaluationScheduleFailureMembershipDenied, ApplicationEvaluationScheduleFailureQuotaConsumerInvalid,
		ApplicationEvaluationScheduleFailureQuotaDenied:
		return http.StatusForbidden
	case ApplicationEvaluationScheduleFailureAuthorizationUnavailable, ApplicationEvaluationScheduleFailureStoreUnavailable,
		ApplicationEvaluationScheduleFailureStoreContract:
		return http.StatusServiceUnavailable
	default:
		return applicationEvaluationHTTPStatus(code)
	}
}

func writeApplicationEvaluationScheduleResult(
	writer http.ResponseWriter,
	status int,
	trace requestTrace,
	ctx ApplicationEvaluationContext,
	result ApplicationEvaluationScheduleResult,
) {
	writeObservedJSON(writer, status, trace, applicationEvaluationScheduleEnvelope{
		RequestID: trace.requestID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, Environment: ctx.Environment,
		ApplicationID: ctx.ApplicationID, Schedule: result.Schedule, Version: result.Version, Occurrence: result.Occurrence,
		FailureCode: applicationEvaluationFailurePointer(result.FailureCode), FailureSummary: result.FailureSummary,
		CurrentRecordVersion: result.CurrentRecordVersion, CurrentState: result.CurrentState, AuditRef: ctx.AuditRef,
	})
}

func writeApplicationEvaluationScheduleListResult(
	writer http.ResponseWriter,
	status int,
	trace requestTrace,
	ctx ApplicationEvaluationContext,
	result ApplicationEvaluationScheduleListResult,
) {
	if result.Schedules == nil {
		result.Schedules = []ApplicationEvaluationSchedule{}
	}
	writeObservedJSON(writer, status, trace, applicationEvaluationScheduleListEnvelope{
		RequestID: trace.requestID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, Environment: ctx.Environment,
		ApplicationID: ctx.ApplicationID, Schedules: result.Schedules, NextCursor: result.NextCursor, HasMore: result.HasMore,
		FailureCode: applicationEvaluationFailurePointer(result.FailureCode), FailureSummary: result.FailureSummary, AuditRef: ctx.AuditRef,
	})
}
