package httpapi

import (
	"net/http"
	"strings"
)

const (
	applicationEvaluationPairPreviewRoute = "POST /v1/user-workspace/applications/{application_id}/evaluation-campaign-pairs/preview"
	applicationEvaluationHandoffRoute     = "POST /v1/user-workspace/applications/{application_id}/evaluation-campaign-pairs/handoff"
)

type applicationEvaluationPairPreviewBody struct {
	WorkspaceID         string `json:"workspace_id"`
	Environment         string `json:"environment"`
	BaselineCampaignID  string `json:"baseline_campaign_id"`
	CandidateCampaignID string `json:"candidate_campaign_id"`
}

type applicationEvaluationHandoffBody struct {
	WorkspaceID                            string `json:"workspace_id"`
	Environment                            string `json:"environment"`
	BaselineCampaignID                     string `json:"baseline_campaign_id"`
	CandidateCampaignID                    string `json:"candidate_campaign_id"`
	ExpectedBaselineCampaignRecordVersion  int    `json:"expected_baseline_campaign_record_version"`
	ExpectedCandidateCampaignRecordVersion int    `json:"expected_candidate_campaign_record_version"`
	AcknowledgeEvidenceMaterializing       bool   `json:"acknowledge_evidence_materializing"`
}

type applicationEvaluationPairEnvelope struct {
	RequestID               string                           `json:"request_id"`
	TenantRef               string                           `json:"tenant_ref"`
	WorkspaceID             string                           `json:"workspace_id"`
	Environment             string                           `json:"environment"`
	ApplicationID           string                           `json:"application_id"`
	Review                  *ApplicationEvaluationPairReview `json:"review"`
	CandidateCampaign       *ApplicationEvaluationCampaign   `json:"candidate_campaign"`
	Handoff                 *ApplicationEvaluationHandoffRef `json:"handoff"`
	IdempotentReplay        bool                             `json:"idempotent_replay"`
	FailureCode             *string                          `json:"failure_code"`
	FailureSummary          string                           `json:"failure_summary"`
	CurrentBaselineVersion  int                              `json:"current_baseline_record_version"`
	CurrentCandidateVersion int                              `json:"current_candidate_record_version"`
	AuditRef                string                           `json:"audit_ref"`
}

func (server *Server) handlePreviewApplicationEvaluationCampaignPair(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, applicationEvaluationPairPreviewRoute)
	if !server.allowApplicationEvaluationCampaignDev(writer, trace) {
		return
	}
	auth, failure, status := server.authorizeWorkspaceScopedPermissions(request, "application_evaluations:read", "workflow_runs:read")
	ctx := applicationEvaluationMutationContext(request, trace, auth, request.PathValue("application_id"), "pair-preview")
	if failure != "" {
		writeApplicationEvaluationPairResult(writer, status, trace, ctx, applicationEvaluationPairFailure(ApplicationEvaluationFailureScopeDenied))
		return
	}
	var body applicationEvaluationPairPreviewBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: maxControlJSONRequestBodyBytes, rejectUnknownFields: true}) {
		return
	}
	ctx.Environment = strings.TrimSpace(body.Environment)
	if !server.applicationEvaluationBindingMatches(request, auth, body.WorkspaceID, ctx.Environment, ctx.ApplicationID) {
		writeApplicationEvaluationPairResult(writer, http.StatusForbidden, trace, ctx, applicationEvaluationPairFailure(ApplicationEvaluationFailureScopeDenied))
		return
	}
	result := server.applicationEvaluationHandoffService().Preview(ctx, ApplicationEvaluationPairInput{
		BaselineCampaignID: body.BaselineCampaignID, CandidateCampaignID: body.CandidateCampaignID,
	})
	writeApplicationEvaluationPairResult(writer, applicationEvaluationHTTPStatus(result.FailureCode), trace, ctx, result)
}

func (server *Server) handleMaterializeApplicationEvaluationHandoff(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, applicationEvaluationHandoffRoute)
	if !server.allowApplicationEvaluationCampaignDev(writer, trace) {
		return
	}
	auth, failure, status := server.authorizeWorkspaceScopedPermissions(request, "application_evaluations:read", "workflow_runs:read", "workflow_evaluations:write")
	ctx := applicationEvaluationMutationContext(request, trace, auth, request.PathValue("application_id"), "pair-handoff")
	if failure != "" {
		writeApplicationEvaluationPairResult(writer, status, trace, ctx, applicationEvaluationPairFailure(ApplicationEvaluationFailureScopeDenied))
		return
	}
	var body applicationEvaluationHandoffBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: maxControlJSONRequestBodyBytes, rejectUnknownFields: true}) {
		return
	}
	ctx.Environment = strings.TrimSpace(body.Environment)
	if !server.applicationEvaluationBindingMatches(request, auth, body.WorkspaceID, ctx.Environment, ctx.ApplicationID) {
		writeApplicationEvaluationPairResult(writer, http.StatusForbidden, trace, ctx, applicationEvaluationPairFailure(ApplicationEvaluationFailureScopeDenied))
		return
	}
	result := server.applicationEvaluationHandoffService().Materialize(ctx, ApplicationEvaluationHandoffInput{
		BaselineCampaignID: body.BaselineCampaignID, CandidateCampaignID: body.CandidateCampaignID,
		ExpectedBaselineRecordVersion:    body.ExpectedBaselineCampaignRecordVersion,
		ExpectedCandidateRecordVersion:   body.ExpectedCandidateCampaignRecordVersion,
		AcknowledgeEvidenceMaterializing: body.AcknowledgeEvidenceMaterializing,
	})
	writeApplicationEvaluationPairResult(writer, applicationEvaluationHTTPStatus(result.FailureCode), trace, ctx, result)
}

func writeApplicationEvaluationPairResult(writer http.ResponseWriter, status int, trace requestTrace, ctx ApplicationEvaluationContext, result ApplicationEvaluationPairResult) {
	writeObservedJSON(writer, status, trace, applicationEvaluationPairEnvelope{
		RequestID: trace.requestID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, Environment: ctx.Environment, ApplicationID: ctx.ApplicationID,
		Review: result.Review, CandidateCampaign: result.CandidateCampaign, Handoff: result.Handoff, IdempotentReplay: result.IdempotentReplay,
		FailureCode: applicationEvaluationFailurePointer(result.FailureCode), FailureSummary: result.FailureSummary,
		CurrentBaselineVersion: result.CurrentBaselineVersion, CurrentCandidateVersion: result.CurrentCandidateVersion, AuditRef: ctx.AuditRef,
	})
}
