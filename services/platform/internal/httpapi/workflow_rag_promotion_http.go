package httpapi

import (
	"net/http"
	"strings"
)

const (
	workflowRAGPromotionCandidateCreateRoute = "POST /v1/user-workspace/workflow-rag-knowledge-promotion-candidates"
	workflowRAGPromotionCandidateListRoute   = "GET /v1/user-workspace/workflow-rag-knowledge-promotion-candidates"
	workflowRAGPromotionCandidateReadRoute   = "GET /v1/user-workspace/workflow-rag-knowledge-promotion-candidates/{candidate_id}"
	workflowRAGPromotionDecisionRoute        = "POST /v1/user-workspace/workflow-rag-knowledge-promotion-candidates/{candidate_id}/decisions"
)

type workflowRAGPromotionCandidateCreateBody struct {
	WorkspaceID          string `json:"workspace_id"`
	ApplicationID        string `json:"application_id"`
	DatasetID            string `json:"dataset_id"`
	DatasetVersion       int    `json:"dataset_version"`
	DatasetDigest        string `json:"dataset_digest"`
	CandidateReviewID    string `json:"candidate_review_id"`
	DraftID              string `json:"draft_id"`
	ExpectedDraftVersion int    `json:"expected_draft_version"`
}

type workflowRAGPromotionDecisionBody struct {
	WorkspaceID           string `json:"workspace_id"`
	ApplicationID         string `json:"application_id"`
	ExpectedRecordVersion int    `json:"expected_record_version"`
	Decision              string `json:"decision"`
	Reason                string `json:"reason"`
}

type workflowRAGPromotionEnvelope struct {
	RequestID            string                                  `json:"request_id"`
	TenantRef            string                                  `json:"tenant_ref"`
	WorkspaceID          string                                  `json:"workspace_id"`
	ApplicationID        string                                  `json:"application_id"`
	Candidate            *WorkflowRAGKnowledgePromotionCandidate `json:"candidate"`
	Decisions            []WorkflowRAGKnowledgePromotionDecision `json:"decisions"`
	Binding              *WorkflowRAGApplicationBinding          `json:"binding"`
	Eligibility          WorkflowRAGPromotionEligibility         `json:"eligibility"`
	FailureCode          *string                                 `json:"failure_code"`
	CurrentRecordVersion int                                     `json:"current_record_version"`
	CurrentState         string                                  `json:"current_state"`
	AuditRef             string                                  `json:"audit_ref"`
}

type workflowRAGPromotionListEnvelope struct {
	RequestID     string                                 `json:"request_id"`
	TenantRef     string                                 `json:"tenant_ref"`
	WorkspaceID   string                                 `json:"workspace_id"`
	ApplicationID string                                 `json:"application_id"`
	Items         []WorkflowRAGPromotionCandidateSummary `json:"items"`
	NextCursor    *string                                `json:"next_cursor"`
	FailureCode   *string                                `json:"failure_code"`
	AuditRef      string                                 `json:"audit_ref"`
}

func (server *Server) handleCreateWorkflowRAGPromotionCandidate(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowRAGPromotionCandidateCreateRoute)
	if !server.allowWorkflowRAGPromotionDev(writer, trace) {
		return
	}
	var body workflowRAGPromotionCandidateCreateBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: maxControlJSONRequestBodyBytes, rejectUnknownFields: true}) {
		return
	}
	ctx, failure := workflowRAGPromotionContextFromRequest(request, trace, body.WorkspaceID, body.ApplicationID, "create",
		"workflow_rag_promotions:write", "workflow_rag_evaluation_datasets:read", "workflow_rag_snapshots:read", "application_drafts:read")
	if failure != "" {
		writeWorkflowRAGPromotionResult(writer, trace, ctx, workflowRAGPromotionFailure(failure))
		return
	}
	result := server.workflowRAGPromotionService().Create(ctx, WorkflowRAGPromotionCreateInput{
		DatasetID: body.DatasetID, DatasetVersion: body.DatasetVersion, DatasetDigest: body.DatasetDigest,
		CandidateReviewID: body.CandidateReviewID, DraftID: body.DraftID, ExpectedDraftVersion: body.ExpectedDraftVersion,
	})
	writeWorkflowRAGPromotionResult(writer, trace, ctx, result)
}

func (server *Server) handleListWorkflowRAGPromotionCandidates(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowRAGPromotionCandidateListRoute)
	if !server.allowWorkflowRAGPromotionDev(writer, trace) {
		return
	}
	values := request.URL.Query()
	if !workflowRAGEvaluationQueryAllowed(values, "workspace_id", "application_id", "limit", "cursor") {
		writeWorkflowRAGPromotionListResult(writer, trace, WorkflowRAGPromotionContext{}, WorkflowRAGPromotionListResult{Items: []WorkflowRAGPromotionCandidateSummary{}, FailureCode: WorkflowRAGPromotionFailurePayloadInvalid})
		return
	}
	ctx, failure := workflowRAGPromotionContextFromRequest(request, trace, values.Get("workspace_id"), values.Get("application_id"), "list", "workflow_rag_promotions:read")
	limit, ok := parseWorkflowRAGEvaluationLimit(values.Get("limit"))
	if !ok {
		failure = WorkflowRAGPromotionFailurePayloadInvalid
	}
	if failure != "" {
		writeWorkflowRAGPromotionListResult(writer, trace, ctx, WorkflowRAGPromotionListResult{Items: []WorkflowRAGPromotionCandidateSummary{}, FailureCode: failure})
		return
	}
	writeWorkflowRAGPromotionListResult(writer, trace, ctx, server.workflowRAGPromotionService().List(ctx, WorkflowRAGPromotionListInput{Limit: limit, Cursor: values.Get("cursor")}))
}

func (server *Server) handleReadWorkflowRAGPromotionCandidate(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowRAGPromotionCandidateReadRoute)
	if !server.allowWorkflowRAGPromotionDev(writer, trace) {
		return
	}
	values := request.URL.Query()
	if !workflowRAGEvaluationQueryAllowed(values, "workspace_id", "application_id") {
		writeWorkflowRAGPromotionResult(writer, trace, WorkflowRAGPromotionContext{}, workflowRAGPromotionFailure(WorkflowRAGPromotionFailurePayloadInvalid))
		return
	}
	ctx, failure := workflowRAGPromotionContextFromRequest(request, trace, values.Get("workspace_id"), values.Get("application_id"), "read", "workflow_rag_promotions:read")
	if failure != "" {
		writeWorkflowRAGPromotionResult(writer, trace, ctx, workflowRAGPromotionFailure(failure))
		return
	}
	writeWorkflowRAGPromotionResult(writer, trace, ctx, server.workflowRAGPromotionService().Read(ctx, request.PathValue("candidate_id")))
}

func (server *Server) handleDecideWorkflowRAGPromotionCandidate(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowRAGPromotionDecisionRoute)
	if !server.allowWorkflowRAGPromotionDev(writer, trace) {
		return
	}
	var body workflowRAGPromotionDecisionBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: maxControlJSONRequestBodyBytes, rejectUnknownFields: true}) {
		return
	}
	ctx, failure := workflowRAGPromotionContextFromRequest(request, trace, body.WorkspaceID, body.ApplicationID, "decision", "workflow_rag_promotions:review")
	if failure != "" {
		writeWorkflowRAGPromotionResult(writer, trace, ctx, workflowRAGPromotionFailure(failure))
		return
	}
	writeWorkflowRAGPromotionResult(writer, trace, ctx, server.workflowRAGPromotionService().Decide(ctx, request.PathValue("candidate_id"), WorkflowRAGPromotionDecisionInput{
		ExpectedRecordVersion: body.ExpectedRecordVersion, Decision: body.Decision, Reason: body.Reason,
	}))
}

func (server *Server) allowWorkflowRAGPromotionDev(writer http.ResponseWriter, trace requestTrace) bool {
	if server.config.WorkflowRAGPromotionDevEnabled {
		return true
	}
	server.writePlatformError(writer, trace, "WORKFLOW_RAG_PROMOTION_DEV_DISABLED", "workflow RAG promotion route requires explicit development opt-in")
	return false
}

func (server *Server) workflowRAGPromotionService() workflowRAGPromotionService {
	if server.workflowRAGPromotionRepository == nil {
		unavailable := newMemoryWorkflowRAGPromotionRepository(nil)
		unavailable.available = false
		server.workflowRAGPromotionRepository = unavailable
	}
	return newWorkflowRAGPromotionService(
		server.workflowRAGPromotionRepository, server.workflowRAGEvaluationDatasetRepository,
		server.workflowRAGSnapshotRepository, server.applicationDraftRepository, server.readWorkflowRAGPromotionApplication,
	)
}

func (server *Server) readWorkflowRAGPromotionApplication(ctx WorkflowRAGPromotionContext) (ApplicationSummary, error) {
	return server.readApplicationPublishBaseline(ApplicationPublishContext{
		RequestContext: ctx.RequestContext, RequestID: ctx.RequestID, TenantRef: ctx.TenantRef,
		WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, ActorRef: ctx.ActorRef,
		OwnerSubjectRef: ctx.OwnerSubjectRef, AuditRef: ctx.AuditRef,
	})
}

func workflowRAGPromotionContextFromRequest(request *http.Request, trace requestTrace, workspaceID, applicationID, suffix string, scopes ...string) (WorkflowRAGPromotionContext, string) {
	runContext, failure := workflowRunContextFromRequest(request, trace, workspaceID, applicationID, "rag-promotion-"+suffix, scopes...)
	ctx := WorkflowRAGPromotionContext{
		RequestContext: request.Context(), RequestID: trace.requestID, TenantRef: runContext.TenantRef,
		WorkspaceID: runContext.WorkspaceID, ApplicationID: runContext.ApplicationID, ActorRef: runContext.ActorRef,
		OwnerSubjectRef: runContext.ActorRef, AuditRef: runContext.AuditRef, WriteEnabled: true,
	}
	if failure != "" {
		return ctx, WorkflowRAGPromotionFailureScopeDenied
	}
	return ctx, ""
}

func writeWorkflowRAGPromotionResult(writer http.ResponseWriter, trace requestTrace, ctx WorkflowRAGPromotionContext, result WorkflowRAGPromotionResult) {
	if result.Decisions == nil {
		result.Decisions = []WorkflowRAGKnowledgePromotionDecision{}
	}
	if result.Eligibility.Blockers == nil {
		result.Eligibility.Blockers = []WorkflowRAGPromotionBlocker{}
	}
	if strings.TrimSpace(result.Eligibility.Status) == "" {
		result.Eligibility.Status = "blocked"
	}
	writeObservedJSON(writer, http.StatusOK, trace, workflowRAGPromotionEnvelope{
		RequestID: trace.requestID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID,
		Candidate: result.Candidate, Decisions: result.Decisions, Binding: result.Binding, Eligibility: result.Eligibility,
		FailureCode: workflowRAGPromotionFailurePointer(result.FailureCode), CurrentRecordVersion: result.CurrentRecordVersion,
		CurrentState: result.CurrentState, AuditRef: ctx.AuditRef,
	})
}

func writeWorkflowRAGPromotionListResult(writer http.ResponseWriter, trace requestTrace, ctx WorkflowRAGPromotionContext, result WorkflowRAGPromotionListResult) {
	if result.Items == nil {
		result.Items = []WorkflowRAGPromotionCandidateSummary{}
	}
	writeObservedJSON(writer, http.StatusOK, trace, workflowRAGPromotionListEnvelope{
		RequestID: trace.requestID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID,
		Items: result.Items, NextCursor: result.NextCursor, FailureCode: workflowRAGPromotionFailurePointer(result.FailureCode), AuditRef: ctx.AuditRef,
	})
}

func workflowRAGPromotionFailurePointer(code string) *string {
	if strings.TrimSpace(code) == "" {
		return nil
	}
	value := strings.TrimSpace(code)
	return &value
}
