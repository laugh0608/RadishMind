package httpapi

import (
	"net/http"
	"strconv"
	"strings"
)

const (
	workflowEvaluationCreateRoute = "POST /v1/user-workspace/workflow-evaluation-cases"
	workflowEvaluationListRoute   = "GET /v1/user-workspace/workflow-evaluation-cases"
	workflowEvaluationReadRoute   = "GET /v1/user-workspace/workflow-evaluation-cases/{case_id}"
	workflowEvaluationReviewRoute = "GET /v1/user-workspace/workflow-evaluation-cases/{case_id}/review"
)

type workflowEvaluationCreateHTTPBody struct {
	WorkspaceID   string                          `json:"workspace_id"`
	ApplicationID string                          `json:"application_id"`
	Name          string                          `json:"name"`
	BaselineRunID string                          `json:"baseline_run_id"`
	Expectations  []WorkflowEvaluationExpectation `json:"expectations"`
}
type workflowEvaluationEnvelope struct {
	RequestID      string                    `json:"request_id"`
	WorkspaceID    string                    `json:"workspace_id"`
	ApplicationID  string                    `json:"application_id"`
	Case           *WorkflowEvaluationCase   `json:"case"`
	Review         *WorkflowEvaluationReview `json:"review"`
	FailureCode    *string                   `json:"failure_code"`
	FailureSummary string                    `json:"failure_summary"`
	AuditRef       string                    `json:"audit_ref"`
}
type workflowEvaluationListEnvelope struct {
	RequestID      string                   `json:"request_id"`
	WorkspaceID    string                   `json:"workspace_id"`
	ApplicationID  string                   `json:"application_id"`
	Cases          []WorkflowEvaluationCase `json:"cases"`
	NextCursor     string                   `json:"next_cursor"`
	HasMore        bool                     `json:"has_more"`
	FailureCode    *string                  `json:"failure_code"`
	FailureSummary string                   `json:"failure_summary"`
	AuditRef       string                   `json:"audit_ref"`
}

func (s *Server) handleCreateWorkflowEvaluation(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowEvaluationCreateRoute)
	if !s.allowWorkflowExecutorDev(writer, trace) {
		return
	}
	var body workflowEvaluationCreateHTTPBody
	if !s.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: maxControlJSONRequestBodyBytes, rejectUnknownFields: true}) {
		return
	}
	ctx, code := workflowRunContextFromRequest(request, trace, body.WorkspaceID, body.ApplicationID, "evaluation-create", "workflow_evaluations:write", "workflow_runs:read")
	if code != "" {
		writeWorkflowEvaluationResult(writer, trace, ctx, workflowEvaluationFailure(WorkflowEvaluationFailureCode(code)))
		return
	}
	result := s.workflowEvaluationService().Create(ctx, WorkflowEvaluationCreateRequest{Name: body.Name, BaselineRunID: body.BaselineRunID, Expectations: body.Expectations})
	writeWorkflowEvaluationResult(writer, trace, ctx, result)
}
func (s *Server) handleReadWorkflowEvaluation(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowEvaluationReadRoute)
	if !s.allowWorkflowExecutorDev(writer, trace) {
		return
	}
	ctx, failureCode := workflowEvaluationContextFromQuery(request, trace, "evaluation-read")
	if failureCode != "" {
		writeWorkflowEvaluationResult(writer, trace, ctx, workflowEvaluationFailure(failureCode))
		return
	}
	writeWorkflowEvaluationResult(writer, trace, ctx, s.workflowEvaluationService().Read(ctx, request.PathValue("case_id")))
}
func (s *Server) handleReviewWorkflowEvaluation(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowEvaluationReviewRoute)
	if !s.allowWorkflowExecutorDev(writer, trace) {
		return
	}
	ctx, failureCode := workflowEvaluationContextFromQuery(request, trace, "evaluation-review")
	if failureCode != "" {
		failure := workflowEvaluationFailure(failureCode)
		writeWorkflowEvaluationReviewResult(writer, trace, ctx, WorkflowEvaluationReviewResult{FailureCode: failureCode, FailureSummary: failure.FailureSummary})
		return
	}
	writeWorkflowEvaluationReviewResult(writer, trace, ctx, s.workflowEvaluationService().Review(ctx, request.PathValue("case_id")))
}
func (s *Server) handleListWorkflowEvaluations(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowEvaluationListRoute)
	if !s.allowWorkflowExecutorDev(writer, trace) {
		return
	}
	values := request.URL.Query()
	allowed := map[string]bool{"workspace_id": true, "application_id": true, "limit": true, "cursor": true, "baseline_run_id": true}
	for key, entries := range values {
		if !allowed[key] || len(entries) != 1 {
			writeWorkflowEvaluationListResult(writer, trace, WorkflowRunContext{}, workflowEvaluationListFailure(WorkflowEvaluationFailureInvalid))
			return
		}
	}
	ctx, code := workflowRunContextFromRequest(request, trace, values.Get("workspace_id"), values.Get("application_id"), "evaluation-list", "workflow_evaluations:read", "workflow_runs:read")
	if code != "" {
		writeWorkflowEvaluationListResult(writer, trace, ctx, workflowEvaluationListFailure(WorkflowEvaluationFailureCode(code)))
		return
	}
	limit := 0
	if raw := strings.TrimSpace(values.Get("limit")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			writeWorkflowEvaluationListResult(writer, trace, ctx, workflowEvaluationListFailure(WorkflowEvaluationFailureInvalid))
			return
		}
		limit = parsed
	}
	writeWorkflowEvaluationListResult(writer, trace, ctx, s.workflowEvaluationService().List(ctx, WorkflowEvaluationListRequest{Limit: limit, Cursor: values.Get("cursor"), BaselineRunID: values.Get("baseline_run_id")}))
}

func workflowEvaluationContextFromQuery(request *http.Request, trace requestTrace, suffix string) (WorkflowRunContext, WorkflowEvaluationFailureCode) {
	values := request.URL.Query()
	allowed := map[string]bool{"workspace_id": true, "application_id": true}
	for key, entries := range values {
		if !allowed[key] || len(entries) != 1 {
			return WorkflowRunContext{}, WorkflowEvaluationFailureInvalid
		}
	}
	ctx, code := workflowRunContextFromRequest(request, trace, values.Get("workspace_id"), values.Get("application_id"), suffix, "workflow_evaluations:read", "workflow_runs:read")
	return ctx, WorkflowEvaluationFailureCode(code)
}
func workflowEvaluationFailurePointer(code WorkflowEvaluationFailureCode) *string {
	if code == "" {
		return nil
	}
	value := string(code)
	return &value
}
func writeWorkflowEvaluationResult(writer http.ResponseWriter, trace requestTrace, ctx WorkflowRunContext, result WorkflowEvaluationResult) {
	writeObservedJSON(writer, http.StatusOK, trace, workflowEvaluationEnvelope{RequestID: trace.requestID, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, Case: result.Case, FailureCode: workflowEvaluationFailurePointer(result.FailureCode), FailureSummary: result.FailureSummary, AuditRef: ctx.AuditRef})
}
func writeWorkflowEvaluationReviewResult(writer http.ResponseWriter, trace requestTrace, ctx WorkflowRunContext, result WorkflowEvaluationReviewResult) {
	writeObservedJSON(writer, http.StatusOK, trace, workflowEvaluationEnvelope{RequestID: trace.requestID, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, Review: result.Review, FailureCode: workflowEvaluationFailurePointer(result.FailureCode), FailureSummary: result.FailureSummary, AuditRef: ctx.AuditRef})
}
func writeWorkflowEvaluationListResult(writer http.ResponseWriter, trace requestTrace, ctx WorkflowRunContext, result WorkflowEvaluationListResult) {
	if result.Cases == nil {
		result.Cases = []WorkflowEvaluationCase{}
	}
	writeObservedJSON(writer, http.StatusOK, trace, workflowEvaluationListEnvelope{RequestID: trace.requestID, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, Cases: result.Cases, NextCursor: result.NextCursor, HasMore: result.HasMore, FailureCode: workflowEvaluationFailurePointer(result.FailureCode), FailureSummary: result.FailureSummary, AuditRef: ctx.AuditRef})
}
