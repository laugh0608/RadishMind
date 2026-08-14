package httpapi

import (
	"errors"
	"net/http"
	"strings"
	"time"
)

const (
	applicationEvaluationCampaignExecuteRoute   = "POST /v1/user-workspace/applications/{application_id}/evaluation-campaigns"
	applicationEvaluationCampaignListRoute      = "GET /v1/user-workspace/applications/{application_id}/evaluation-campaigns"
	applicationEvaluationCampaignReadRoute      = "GET /v1/user-workspace/applications/{application_id}/evaluation-campaigns/{campaign_id}"
	applicationEvaluationCampaignReconcileRoute = "POST /v1/user-workspace/applications/{application_id}/evaluation-campaigns/{campaign_id}/reconcile"
)

type applicationEvaluationCampaignExecuteBody struct {
	WorkspaceID                    string `json:"workspace_id"`
	Environment                    string `json:"environment"`
	PlanID                         string `json:"plan_id"`
	PlanVersion                    int    `json:"plan_version"`
	PlanDigest                     string `json:"plan_digest"`
	ExpectedPlanRecordVersion      int    `json:"expected_plan_record_version"`
	ClientCampaignKey              string `json:"client_campaign_key"`
	QuotaAPIKeyID                  string `json:"quota_api_key_id"`
	AcknowledgeSequentialExecution bool   `json:"acknowledge_sequential_execution"`
	AcknowledgeQuotaConsumption    bool   `json:"acknowledge_quota_consumption"`
}

type applicationEvaluationCampaignReconcileBody struct {
	WorkspaceID     string `json:"workspace_id"`
	Environment     string `json:"environment"`
	ExpectedVersion int    `json:"expected_version"`
}

type applicationEvaluationCampaignEnvelope struct {
	RequestID            string                         `json:"request_id"`
	TenantRef            string                         `json:"tenant_ref"`
	WorkspaceID          string                         `json:"workspace_id"`
	Environment          string                         `json:"environment"`
	ApplicationID        string                         `json:"application_id"`
	Campaign             *ApplicationEvaluationCampaign `json:"campaign"`
	IdempotentReplay     bool                           `json:"idempotent_replay"`
	FailureCode          *string                        `json:"failure_code"`
	FailureSummary       string                         `json:"failure_summary"`
	CurrentRecordVersion int                            `json:"current_record_version"`
	CurrentState         string                         `json:"current_state"`
	AuditRef             string                         `json:"audit_ref"`
}

type applicationEvaluationCampaignListEnvelope struct {
	RequestID      string                          `json:"request_id"`
	TenantRef      string                          `json:"tenant_ref"`
	WorkspaceID    string                          `json:"workspace_id"`
	Environment    string                          `json:"environment"`
	ApplicationID  string                          `json:"application_id"`
	Campaigns      []ApplicationEvaluationCampaign `json:"campaigns"`
	NextCursor     string                          `json:"next_cursor"`
	HasMore        bool                            `json:"has_more"`
	FailureCode    *string                         `json:"failure_code"`
	FailureSummary string                          `json:"failure_summary"`
	AuditRef       string                          `json:"audit_ref"`
}

func (server *Server) handleExecuteApplicationEvaluationCampaign(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, applicationEvaluationCampaignExecuteRoute)
	if !server.allowApplicationEvaluationCampaignDev(writer, trace) {
		return
	}
	auth, failure, status := server.authorizeWorkspaceScopedPermissions(request, "application_evaluations:execute", "workflow_runs:execute")
	ctx := applicationEvaluationMutationContext(request, trace, auth, request.PathValue("application_id"), "campaign-execute")
	if failure != "" {
		writeApplicationEvaluationCampaignResult(writer, status, trace, ctx, applicationEvaluationCampaignFailure(failure))
		return
	}
	var body applicationEvaluationCampaignExecuteBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: maxControlJSONRequestBodyBytes, rejectUnknownFields: true}) {
		return
	}
	ctx.Environment = strings.TrimSpace(body.Environment)
	if !server.applicationEvaluationBindingMatches(request, auth, body.WorkspaceID, ctx.Environment, ctx.ApplicationID) {
		writeApplicationEvaluationCampaignResult(writer, http.StatusForbidden, trace, ctx, applicationEvaluationCampaignFailure(ApplicationEvaluationFailureScopeDenied))
		return
	}
	versionResult := server.applicationEvaluationPlanService().ReadVersion(ctx, body.PlanID, body.PlanVersion)
	if versionResult.FailureCode != "" {
		writeApplicationEvaluationCampaignResult(writer, applicationEvaluationHTTPStatus(versionResult.FailureCode), trace, ctx, applicationEvaluationCampaignFailure(versionResult.FailureCode))
		return
	}
	if versionResult.Version != nil && versionResult.Version.ExecutionProfile == applicationInteractionProfileWorkflow {
		_, definitionFailure, definitionStatus := server.authorizeWorkspaceScopedPermissions(request, "workflow_definitions:read")
		if definitionFailure != "" {
			writeApplicationEvaluationCampaignResult(writer, definitionStatus, trace, ctx, applicationEvaluationCampaignFailure(ApplicationEvaluationFailureScopeDenied))
			return
		}
	}
	if quotaFailure := server.validateApplicationEvaluationQuotaConsumer(ctx, body.QuotaAPIKeyID); quotaFailure != "" {
		writeApplicationEvaluationCampaignResult(writer, applicationEvaluationHTTPStatus(quotaFailure), trace, ctx, applicationEvaluationCampaignFailure(quotaFailure))
		return
	}
	result := server.applicationEvaluationCampaignService().Execute(ctx, ApplicationEvaluationCampaignExecuteInput{
		PlanID: body.PlanID, PlanVersion: body.PlanVersion, PlanDigest: body.PlanDigest,
		ExpectedPlanRecordVersion: body.ExpectedPlanRecordVersion, ClientCampaignKey: body.ClientCampaignKey, QuotaAPIKeyID: body.QuotaAPIKeyID,
		AcknowledgeSequentialExecution: body.AcknowledgeSequentialExecution, AcknowledgeQuotaConsumption: body.AcknowledgeQuotaConsumption,
	})
	writeApplicationEvaluationCampaignResult(writer, applicationEvaluationHTTPStatus(result.FailureCode), trace, ctx, result)
}

func (server *Server) validateApplicationEvaluationQuotaConsumer(ctx ApplicationEvaluationContext, apiKeyID string) string {
	apiKeyID = strings.TrimSpace(apiKeyID)
	if !apiKeyIDPattern.MatchString(apiKeyID) {
		return ApplicationEvaluationFailureQuotaConsumerInvalid
	}
	if server.apiKeyRepository == nil {
		return ApplicationEvaluationFailureStoreUnavailable
	}
	record, err := server.apiKeyRepository.Read(APIKeyContext{
		RequestContext: ctx.RequestContext, RequestID: ctx.RequestID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID,
		ActorRef: ctx.ActorRef, OwnerSubjectRef: ctx.ActorRef, AuditRef: ctx.AuditRef,
	}, apiKeyID)
	if err != nil {
		if errors.Is(err, errAPIKeyStoreUnavailable) {
			return ApplicationEvaluationFailureStoreUnavailable
		}
		return ApplicationEvaluationFailureQuotaConsumerInvalid
	}
	if record.ApplicationID != ctx.ApplicationID || record.OwnerSubjectRef != ctx.ActorRef || effectiveAPIKeyState(record, time.Now().UTC()) != apiKeyLifecycleActive {
		return ApplicationEvaluationFailureQuotaConsumerInvalid
	}
	return ""
}

func (server *Server) handleReconcileApplicationEvaluationCampaign(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, applicationEvaluationCampaignReconcileRoute)
	if !server.allowApplicationEvaluationCampaignDev(writer, trace) {
		return
	}
	auth, failure, status := server.authorizeWorkspaceScopedPermissions(request, "application_evaluations:execute", "workflow_runs:execute")
	ctx := applicationEvaluationMutationContext(request, trace, auth, request.PathValue("application_id"), "campaign-reconcile")
	if failure != "" {
		writeApplicationEvaluationCampaignResult(writer, status, trace, ctx, applicationEvaluationCampaignFailure(failure))
		return
	}
	var body applicationEvaluationCampaignReconcileBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: maxControlJSONRequestBodyBytes, rejectUnknownFields: true}) {
		return
	}
	ctx.Environment = strings.TrimSpace(body.Environment)
	if !server.applicationEvaluationBindingMatches(request, auth, body.WorkspaceID, ctx.Environment, ctx.ApplicationID) {
		writeApplicationEvaluationCampaignResult(writer, http.StatusForbidden, trace, ctx, applicationEvaluationCampaignFailure(ApplicationEvaluationFailureScopeDenied))
		return
	}
	result := server.applicationEvaluationCampaignService().Reconcile(ctx, request.PathValue("campaign_id"), body.ExpectedVersion)
	writeApplicationEvaluationCampaignResult(writer, applicationEvaluationHTTPStatus(result.FailureCode), trace, ctx, result)
}

func (server *Server) handleReadApplicationEvaluationCampaign(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, applicationEvaluationCampaignReadRoute)
	ctx, status, failure := server.applicationEvaluationReadContext(request, trace, "campaign-read")
	if failure != "" {
		writeApplicationEvaluationCampaignResult(writer, status, trace, ctx, applicationEvaluationCampaignFailure(failure))
		return
	}
	result := server.applicationEvaluationCampaignService().Read(ctx, request.PathValue("campaign_id"))
	writeApplicationEvaluationCampaignResult(writer, applicationEvaluationHTTPStatus(result.FailureCode), trace, ctx, result)
}

func (server *Server) handleListApplicationEvaluationCampaigns(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, applicationEvaluationCampaignListRoute)
	ctx, status, failure := server.applicationEvaluationReadContext(request, trace, "campaign-list", "plan_id", "limit", "cursor")
	if failure != "" {
		writeApplicationEvaluationCampaignListResult(writer, status, trace, ctx, applicationEvaluationCampaignListFailure(failure))
		return
	}
	limit, ok := applicationEvaluationQueryLimit(request)
	if !ok {
		writeApplicationEvaluationCampaignListResult(writer, http.StatusBadRequest, trace, ctx, applicationEvaluationCampaignListFailure(ApplicationEvaluationFailurePayloadInvalid))
		return
	}
	result := server.applicationEvaluationCampaignService().List(ctx, ApplicationEvaluationCampaignListInput{
		PlanID: request.URL.Query().Get("plan_id"), Limit: limit, Cursor: request.URL.Query().Get("cursor"),
	})
	writeApplicationEvaluationCampaignListResult(writer, applicationEvaluationHTTPStatus(result.FailureCode), trace, ctx, result)
}

func writeApplicationEvaluationCampaignResult(writer http.ResponseWriter, status int, trace requestTrace, ctx ApplicationEvaluationContext, result ApplicationEvaluationCampaignResult) {
	writeObservedJSON(writer, status, trace, applicationEvaluationCampaignEnvelope{
		RequestID: trace.requestID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, Environment: ctx.Environment, ApplicationID: ctx.ApplicationID,
		Campaign: result.Campaign, IdempotentReplay: result.IdempotentReplay,
		FailureCode: applicationEvaluationFailurePointer(result.FailureCode), FailureSummary: result.FailureSummary,
		CurrentRecordVersion: result.CurrentRecordVersion, CurrentState: result.CurrentState, AuditRef: ctx.AuditRef,
	})
}

func writeApplicationEvaluationCampaignListResult(writer http.ResponseWriter, status int, trace requestTrace, ctx ApplicationEvaluationContext, result ApplicationEvaluationCampaignListResult) {
	if result.Campaigns == nil {
		result.Campaigns = []ApplicationEvaluationCampaign{}
	}
	writeObservedJSON(writer, status, trace, applicationEvaluationCampaignListEnvelope{
		RequestID: trace.requestID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, Environment: ctx.Environment, ApplicationID: ctx.ApplicationID,
		Campaigns: result.Campaigns, NextCursor: result.NextCursor, HasMore: result.HasMore,
		FailureCode: applicationEvaluationFailurePointer(result.FailureCode), FailureSummary: result.FailureSummary, AuditRef: ctx.AuditRef,
	})
}
