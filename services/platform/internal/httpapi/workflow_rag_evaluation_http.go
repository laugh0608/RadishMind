package httpapi

import (
	"net/http"
	"strconv"
	"strings"
)

const (
	workflowRAGEvaluationDatasetCreateRoute  = "POST /v1/user-workspace/workflow-rag-evaluation-datasets"
	workflowRAGEvaluationDatasetListRoute    = "GET /v1/user-workspace/workflow-rag-evaluation-datasets"
	workflowRAGEvaluationDatasetReadRoute    = "GET /v1/user-workspace/workflow-rag-evaluation-datasets/{dataset_id}"
	workflowRAGEvaluationDatasetVersionRoute = "POST /v1/user-workspace/workflow-rag-evaluation-datasets/{dataset_id}/versions"
	workflowRAGEvaluationDatasetArchiveRoute = "POST /v1/user-workspace/workflow-rag-evaluation-datasets/{dataset_id}/archive"
	workflowRAGCandidateReviewCreateRoute    = "POST /v1/user-workspace/workflow-rag-evaluation-datasets/{dataset_id}/candidate-reviews"
	workflowRAGCandidateReviewListRoute      = "GET /v1/user-workspace/workflow-rag-evaluation-datasets/{dataset_id}/candidate-reviews"
	workflowRAGCandidateReviewReadRoute      = "GET /v1/user-workspace/workflow-rag-evaluation-datasets/{dataset_id}/candidate-reviews/{review_id}"
	workflowRAGEvaluationMaxRequestBodyBytes = int64(1 << 20)
)

type workflowRAGEvaluationDatasetCreateBody struct {
	WorkspaceID           string                               `json:"workspace_id"`
	ApplicationID         string                               `json:"application_id"`
	DatasetKey            string                               `json:"dataset_key"`
	DisplayName           string                               `json:"display_name"`
	ContentClassification string                               `json:"content_classification"`
	BaselineSnapshot      WorkflowRAGEvaluationSnapshotBinding `json:"baseline_snapshot"`
	Thresholds            WorkflowRAGEvaluationThresholds      `json:"thresholds"`
	ReviewSummary         string                               `json:"review_summary"`
	Samples               []WorkflowRAGEvaluationSample        `json:"samples"`
}

type workflowRAGEvaluationDatasetVersionBody struct {
	WorkspaceID           string                               `json:"workspace_id"`
	ApplicationID         string                               `json:"application_id"`
	ExpectedLatestVersion int                                  `json:"expected_latest_version"`
	DisplayName           string                               `json:"display_name"`
	ContentClassification string                               `json:"content_classification"`
	BaselineSnapshot      WorkflowRAGEvaluationSnapshotBinding `json:"baseline_snapshot"`
	Thresholds            WorkflowRAGEvaluationThresholds      `json:"thresholds"`
	ReviewSummary         string                               `json:"review_summary"`
	Samples               []WorkflowRAGEvaluationSample        `json:"samples"`
}

type workflowRAGEvaluationDatasetArchiveBody struct {
	WorkspaceID           string `json:"workspace_id"`
	ApplicationID         string `json:"application_id"`
	ExpectedLatestVersion int    `json:"expected_latest_version"`
}

type workflowRAGCandidateReviewCreateBody struct {
	WorkspaceID       string                               `json:"workspace_id"`
	ApplicationID     string                               `json:"application_id"`
	DatasetVersion    int                                  `json:"dataset_version"`
	DatasetDigest     string                               `json:"dataset_digest"`
	CandidateSnapshot WorkflowRAGEvaluationSnapshotBinding `json:"candidate_snapshot"`
}

type workflowRAGEvaluationEnvelope struct {
	RequestID            string                                `json:"request_id"`
	TenantRef            string                                `json:"tenant_ref"`
	WorkspaceID          string                                `json:"workspace_id"`
	ApplicationID        string                                `json:"application_id"`
	Resource             *WorkflowRAGEvaluationDatasetResource `json:"resource"`
	Version              *WorkflowRAGEvaluationDatasetVersion  `json:"version"`
	Review               *WorkflowRAGCandidateSnapshotReview   `json:"review"`
	FailureCode          *string                               `json:"failure_code"`
	CurrentLatestVersion int                                   `json:"current_latest_version"`
	CurrentLifecycle     string                                `json:"current_lifecycle_state"`
	AuditRef             string                                `json:"audit_ref"`
}

type workflowRAGEvaluationListEnvelope struct {
	RequestID     string                                 `json:"request_id"`
	TenantRef     string                                 `json:"tenant_ref"`
	WorkspaceID   string                                 `json:"workspace_id"`
	ApplicationID string                                 `json:"application_id"`
	Items         []WorkflowRAGEvaluationDatasetResource `json:"items"`
	NextCursor    *string                                `json:"next_cursor"`
	FailureCode   *string                                `json:"failure_code"`
	AuditRef      string                                 `json:"audit_ref"`
}

type workflowRAGCandidateReviewListEnvelope struct {
	RequestID     string                               `json:"request_id"`
	TenantRef     string                               `json:"tenant_ref"`
	WorkspaceID   string                               `json:"workspace_id"`
	ApplicationID string                               `json:"application_id"`
	Items         []WorkflowRAGCandidateSnapshotReview `json:"items"`
	NextCursor    *string                              `json:"next_cursor"`
	FailureCode   *string                              `json:"failure_code"`
	AuditRef      string                               `json:"audit_ref"`
}

func (server *Server) handleCreateWorkflowRAGEvaluationDataset(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowRAGEvaluationDatasetCreateRoute)
	if !server.allowWorkflowRAGEvaluationDev(writer, trace) {
		return
	}
	var body workflowRAGEvaluationDatasetCreateBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: workflowRAGEvaluationMaxRequestBodyBytes, rejectUnknownFields: true}) {
		return
	}
	ctx, failure := workflowRAGEvaluationContextFromRequest(request, trace, body.WorkspaceID, body.ApplicationID, "create", "workflow_rag_evaluation_datasets:write", "workflow_rag_snapshots:read")
	if failure != "" {
		writeWorkflowRAGEvaluationResult(writer, trace, ctx, workflowRAGEvaluationFailure(failure))
		return
	}
	writeWorkflowRAGEvaluationResult(writer, trace, ctx, server.workflowRAGEvaluationDatasetService().Create(ctx, WorkflowRAGEvaluationDatasetCreateInput{DatasetKey: body.DatasetKey, DisplayName: body.DisplayName, ContentClassification: body.ContentClassification, BaselineSnapshot: body.BaselineSnapshot, Thresholds: body.Thresholds, ReviewSummary: body.ReviewSummary, Samples: body.Samples}))
}

func (server *Server) handleListWorkflowRAGEvaluationDatasets(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowRAGEvaluationDatasetListRoute)
	if !server.allowWorkflowRAGEvaluationDev(writer, trace) {
		return
	}
	values := request.URL.Query()
	if !workflowRAGEvaluationQueryAllowed(values, "workspace_id", "application_id", "lifecycle_state", "limit", "cursor") {
		writeWorkflowRAGEvaluationListResult(writer, trace, WorkflowRAGSnapshotContext{}, WorkflowRAGEvaluationListResult{Resources: []WorkflowRAGEvaluationDatasetResource{}, FailureCode: WorkflowRAGEvaluationFailurePayloadInvalid})
		return
	}
	ctx, failure := workflowRAGEvaluationContextFromRequest(request, trace, values.Get("workspace_id"), values.Get("application_id"), "list", "workflow_rag_evaluation_datasets:read")
	if failure != "" {
		writeWorkflowRAGEvaluationListResult(writer, trace, ctx, WorkflowRAGEvaluationListResult{Resources: []WorkflowRAGEvaluationDatasetResource{}, FailureCode: failure})
		return
	}
	limit, ok := parseWorkflowRAGEvaluationLimit(values.Get("limit"))
	if !ok {
		writeWorkflowRAGEvaluationListResult(writer, trace, ctx, WorkflowRAGEvaluationListResult{Resources: []WorkflowRAGEvaluationDatasetResource{}, FailureCode: WorkflowRAGEvaluationFailurePayloadInvalid})
		return
	}
	writeWorkflowRAGEvaluationListResult(writer, trace, ctx, server.workflowRAGEvaluationDatasetService().List(ctx, WorkflowRAGEvaluationListInput{LifecycleState: values.Get("lifecycle_state"), Limit: limit, Cursor: values.Get("cursor")}))
}

func (server *Server) handleReadWorkflowRAGEvaluationDataset(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowRAGEvaluationDatasetReadRoute)
	if !server.allowWorkflowRAGEvaluationDev(writer, trace) {
		return
	}
	values := request.URL.Query()
	if !workflowRAGEvaluationQueryAllowed(values, "workspace_id", "application_id", "dataset_version") {
		writeWorkflowRAGEvaluationResult(writer, trace, WorkflowRAGSnapshotContext{}, workflowRAGEvaluationFailure(WorkflowRAGEvaluationFailurePayloadInvalid))
		return
	}
	ctx, failure := workflowRAGEvaluationContextFromRequest(request, trace, values.Get("workspace_id"), values.Get("application_id"), "read", "workflow_rag_evaluation_datasets:read", "workflow_rag_evaluation_datasets:read_content")
	version, err := strconv.Atoi(strings.TrimSpace(values.Get("dataset_version")))
	if err != nil || version < 1 {
		failure = WorkflowRAGEvaluationFailurePayloadInvalid
	}
	if failure != "" {
		writeWorkflowRAGEvaluationResult(writer, trace, ctx, workflowRAGEvaluationFailure(failure))
		return
	}
	writeWorkflowRAGEvaluationResult(writer, trace, ctx, server.workflowRAGEvaluationDatasetService().Read(ctx, request.PathValue("dataset_id"), version))
}

func (server *Server) handleVersionWorkflowRAGEvaluationDataset(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowRAGEvaluationDatasetVersionRoute)
	if !server.allowWorkflowRAGEvaluationDev(writer, trace) {
		return
	}
	var body workflowRAGEvaluationDatasetVersionBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: workflowRAGEvaluationMaxRequestBodyBytes, rejectUnknownFields: true}) {
		return
	}
	ctx, failure := workflowRAGEvaluationContextFromRequest(request, trace, body.WorkspaceID, body.ApplicationID, "version", "workflow_rag_evaluation_datasets:write", "workflow_rag_snapshots:read")
	if failure != "" {
		writeWorkflowRAGEvaluationResult(writer, trace, ctx, workflowRAGEvaluationFailure(failure))
		return
	}
	writeWorkflowRAGEvaluationResult(writer, trace, ctx, server.workflowRAGEvaluationDatasetService().Version(ctx, request.PathValue("dataset_id"), WorkflowRAGEvaluationDatasetVersionInput{ExpectedLatestVersion: body.ExpectedLatestVersion, DisplayName: body.DisplayName, ContentClassification: body.ContentClassification, BaselineSnapshot: body.BaselineSnapshot, Thresholds: body.Thresholds, ReviewSummary: body.ReviewSummary, Samples: body.Samples}))
}

func (server *Server) handleArchiveWorkflowRAGEvaluationDataset(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowRAGEvaluationDatasetArchiveRoute)
	if !server.allowWorkflowRAGEvaluationDev(writer, trace) {
		return
	}
	var body workflowRAGEvaluationDatasetArchiveBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: maxControlJSONRequestBodyBytes, rejectUnknownFields: true}) {
		return
	}
	ctx, failure := workflowRAGEvaluationContextFromRequest(request, trace, body.WorkspaceID, body.ApplicationID, "archive", "workflow_rag_evaluation_datasets:archive")
	if failure != "" {
		writeWorkflowRAGEvaluationResult(writer, trace, ctx, workflowRAGEvaluationFailure(failure))
		return
	}
	writeWorkflowRAGEvaluationResult(writer, trace, ctx, server.workflowRAGEvaluationDatasetService().Archive(ctx, request.PathValue("dataset_id"), body.ExpectedLatestVersion))
}

func (server *Server) handleCreateWorkflowRAGCandidateReview(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowRAGCandidateReviewCreateRoute)
	if !server.allowWorkflowRAGEvaluationDev(writer, trace) {
		return
	}
	var body workflowRAGCandidateReviewCreateBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: maxControlJSONRequestBodyBytes, rejectUnknownFields: true}) {
		return
	}
	ctx, failure := workflowRAGEvaluationContextFromRequest(request, trace, body.WorkspaceID, body.ApplicationID, "candidate-review", "workflow_rag_evaluation_datasets:review", "workflow_rag_evaluation_datasets:read", "workflow_rag_snapshots:read")
	if failure != "" {
		writeWorkflowRAGEvaluationResult(writer, trace, ctx, workflowRAGEvaluationFailure(failure))
		return
	}
	writeWorkflowRAGEvaluationResult(writer, trace, ctx, server.workflowRAGEvaluationDatasetService().CreateCandidateReview(ctx, request.PathValue("dataset_id"), WorkflowRAGCandidateReviewInput{DatasetVersion: body.DatasetVersion, DatasetDigest: body.DatasetDigest, CandidateSnapshot: body.CandidateSnapshot}))
}

func (server *Server) handleListWorkflowRAGCandidateReviews(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowRAGCandidateReviewListRoute)
	if !server.allowWorkflowRAGEvaluationDev(writer, trace) {
		return
	}
	values := request.URL.Query()
	if !workflowRAGEvaluationQueryAllowed(values, "workspace_id", "application_id", "limit", "cursor") {
		writeWorkflowRAGCandidateReviewListResult(writer, trace, WorkflowRAGSnapshotContext{}, WorkflowRAGCandidateReviewListResult{Reviews: []WorkflowRAGCandidateSnapshotReview{}, FailureCode: WorkflowRAGEvaluationFailurePayloadInvalid})
		return
	}
	ctx, failure := workflowRAGEvaluationContextFromRequest(request, trace, values.Get("workspace_id"), values.Get("application_id"), "candidate-review-list", "workflow_rag_evaluation_datasets:read")
	limit, ok := parseWorkflowRAGEvaluationLimit(values.Get("limit"))
	if !ok {
		failure = WorkflowRAGEvaluationFailurePayloadInvalid
	}
	if failure != "" {
		writeWorkflowRAGCandidateReviewListResult(writer, trace, ctx, WorkflowRAGCandidateReviewListResult{Reviews: []WorkflowRAGCandidateSnapshotReview{}, FailureCode: failure})
		return
	}
	writeWorkflowRAGCandidateReviewListResult(writer, trace, ctx, server.workflowRAGEvaluationDatasetService().ListCandidateReviews(ctx, request.PathValue("dataset_id"), limit, values.Get("cursor")))
}

func (server *Server) handleReadWorkflowRAGCandidateReview(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowRAGCandidateReviewReadRoute)
	if !server.allowWorkflowRAGEvaluationDev(writer, trace) {
		return
	}
	values := request.URL.Query()
	if !workflowRAGEvaluationQueryAllowed(values, "workspace_id", "application_id") {
		writeWorkflowRAGEvaluationResult(writer, trace, WorkflowRAGSnapshotContext{}, workflowRAGEvaluationFailure(WorkflowRAGEvaluationFailurePayloadInvalid))
		return
	}
	ctx, failure := workflowRAGEvaluationContextFromRequest(request, trace, values.Get("workspace_id"), values.Get("application_id"), "candidate-review-read", "workflow_rag_evaluation_datasets:read")
	if failure != "" {
		writeWorkflowRAGEvaluationResult(writer, trace, ctx, workflowRAGEvaluationFailure(failure))
		return
	}
	writeWorkflowRAGEvaluationResult(writer, trace, ctx, server.workflowRAGEvaluationDatasetService().ReadCandidateReview(ctx, request.PathValue("dataset_id"), request.PathValue("review_id")))
}

func workflowRAGEvaluationContextFromRequest(request *http.Request, trace requestTrace, workspaceID, applicationID, suffix string, scopes ...string) (WorkflowRAGSnapshotContext, string) {
	runContext, failure := workflowRunContextFromRequest(request, trace, workspaceID, applicationID, "rag-evaluation-"+suffix, scopes...)
	ctx := WorkflowRAGSnapshotContext{RequestContext: request.Context(), RequestID: trace.requestID, TenantRef: runContext.TenantRef, WorkspaceID: runContext.WorkspaceID, ApplicationID: runContext.ApplicationID, ActorRef: runContext.ActorRef, AuditRef: runContext.AuditRef}
	if failure != "" {
		return ctx, WorkflowRAGEvaluationFailureScopeDenied
	}
	return ctx, ""
}

func (server *Server) allowWorkflowRAGEvaluationDev(writer http.ResponseWriter, trace requestTrace) bool {
	if server.config.WorkflowRAGEvaluationDevEnabled {
		return true
	}
	server.writePlatformError(writer, trace, "WORKFLOW_RAG_EVALUATION_DEV_DISABLED", "workflow RAG evaluation route requires explicit development opt-in")
	return false
}

func (server *Server) workflowRAGEvaluationDatasetService() workflowRAGEvaluationDatasetService {
	if server.workflowRAGEvaluationDatasetRepository == nil {
		unavailable := newMemoryWorkflowRAGEvaluationDatasetRepository(nil)
		unavailable.available = false
		server.workflowRAGEvaluationDatasetRepository = unavailable
	}
	return newWorkflowRAGEvaluationDatasetService(server.workflowRAGEvaluationDatasetRepository, server.workflowRAGSnapshotRepository)
}

func workflowRAGEvaluationQueryAllowed(values map[string][]string, allowed ...string) bool {
	permitted := make(map[string]bool, len(allowed))
	for _, key := range allowed {
		permitted[key] = true
	}
	for key, entries := range values {
		if !permitted[key] || len(entries) != 1 {
			return false
		}
	}
	return true
}

func parseWorkflowRAGEvaluationLimit(raw string) (int, bool) {
	if strings.TrimSpace(raw) == "" {
		return 0, true
	}
	limit, err := strconv.Atoi(strings.TrimSpace(raw))
	return limit, err == nil
}

func workflowRAGEvaluationFailurePointer(code string) *string {
	if code == "" {
		return nil
	}
	return &code
}

func writeWorkflowRAGEvaluationResult(writer http.ResponseWriter, trace requestTrace, ctx WorkflowRAGSnapshotContext, result WorkflowRAGEvaluationResult) {
	writeObservedJSON(writer, http.StatusOK, trace, workflowRAGEvaluationEnvelope{RequestID: trace.requestID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, Resource: result.Resource, Version: result.Version, Review: result.Review, FailureCode: workflowRAGEvaluationFailurePointer(result.FailureCode), CurrentLatestVersion: result.CurrentLatestVersion, CurrentLifecycle: result.CurrentLifecycle, AuditRef: ctx.AuditRef})
}

func writeWorkflowRAGEvaluationListResult(writer http.ResponseWriter, trace requestTrace, ctx WorkflowRAGSnapshotContext, result WorkflowRAGEvaluationListResult) {
	if result.Resources == nil {
		result.Resources = []WorkflowRAGEvaluationDatasetResource{}
	}
	writeObservedJSON(writer, http.StatusOK, trace, workflowRAGEvaluationListEnvelope{RequestID: trace.requestID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, Items: result.Resources, NextCursor: result.NextCursor, FailureCode: workflowRAGEvaluationFailurePointer(result.FailureCode), AuditRef: ctx.AuditRef})
}

func writeWorkflowRAGCandidateReviewListResult(writer http.ResponseWriter, trace requestTrace, ctx WorkflowRAGSnapshotContext, result WorkflowRAGCandidateReviewListResult) {
	if result.Reviews == nil {
		result.Reviews = []WorkflowRAGCandidateSnapshotReview{}
	}
	writeObservedJSON(writer, http.StatusOK, trace, workflowRAGCandidateReviewListEnvelope{RequestID: trace.requestID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, Items: result.Reviews, NextCursor: result.NextCursor, FailureCode: workflowRAGEvaluationFailurePointer(result.FailureCode), AuditRef: ctx.AuditRef})
}
