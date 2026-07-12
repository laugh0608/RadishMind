package httpapi

import (
	"net/http"
	"strconv"
	"strings"
)

const (
	workflowEvaluationSuiteCreateRoute    = "POST /v1/user-workspace/workflow-evaluation-suites"
	workflowEvaluationSuiteListRoute      = "GET /v1/user-workspace/workflow-evaluation-suites"
	workflowEvaluationSuiteReadRoute      = "GET /v1/user-workspace/workflow-evaluation-suites/{suite_id}"
	workflowEvaluationSuiteReviewRoute    = "GET /v1/user-workspace/workflow-evaluation-suites/{suite_id}/review"
	workflowEvaluationDecisionCreateRoute = "POST /v1/user-workspace/workflow-evaluation-suites/{suite_id}/decisions"
	workflowEvaluationDecisionListRoute   = "GET /v1/user-workspace/workflow-evaluation-suites/{suite_id}/decisions"
)

type workflowEvaluationSuiteCreateHTTPBody struct {
	WorkspaceID   string                           `json:"workspace_id"`
	ApplicationID string                           `json:"application_id"`
	Name          string                           `json:"name"`
	CaseRefs      []WorkflowEvaluationSuiteCaseRef `json:"case_refs"`
}

type workflowEvaluationDecisionHTTPBody struct {
	WorkspaceID             string `json:"workspace_id"`
	ApplicationID           string `json:"application_id"`
	ExpectedDecisionVersion int    `json:"expected_decision_version"`
	Decision                string `json:"decision"`
	ReviewDigest            string `json:"review_digest"`
}

type workflowEvaluationSuiteEnvelope struct {
	RequestID      string                             `json:"request_id"`
	WorkspaceID    string                             `json:"workspace_id"`
	ApplicationID  string                             `json:"application_id"`
	Suite          *WorkflowEvaluationSuite           `json:"suite"`
	Decision       *WorkflowEvaluationReleaseDecision `json:"decision"`
	Review         *WorkflowEvaluationSuiteReview     `json:"review"`
	FailureCode    *string                            `json:"failure_code"`
	FailureSummary string                             `json:"failure_summary"`
	AuditRef       string                             `json:"audit_ref"`
}

type workflowEvaluationSuiteListEnvelope struct {
	RequestID      string                    `json:"request_id"`
	WorkspaceID    string                    `json:"workspace_id"`
	ApplicationID  string                    `json:"application_id"`
	Suites         []WorkflowEvaluationSuite `json:"suites"`
	NextCursor     string                    `json:"next_cursor"`
	HasMore        bool                      `json:"has_more"`
	FailureCode    *string                   `json:"failure_code"`
	FailureSummary string                    `json:"failure_summary"`
	AuditRef       string                    `json:"audit_ref"`
}

type workflowEvaluationDecisionListEnvelope struct {
	RequestID      string                              `json:"request_id"`
	WorkspaceID    string                              `json:"workspace_id"`
	ApplicationID  string                              `json:"application_id"`
	Decisions      []WorkflowEvaluationReleaseDecision `json:"decisions"`
	NextCursor     string                              `json:"next_cursor"`
	HasMore        bool                                `json:"has_more"`
	FailureCode    *string                             `json:"failure_code"`
	FailureSummary string                              `json:"failure_summary"`
	AuditRef       string                              `json:"audit_ref"`
}

func (s *Server) handleCreateWorkflowEvaluationSuite(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowEvaluationSuiteCreateRoute)
	if !s.allowWorkflowExecutorDev(writer, trace) {
		return
	}
	var body workflowEvaluationSuiteCreateHTTPBody
	if !s.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: maxControlJSONRequestBodyBytes, rejectUnknownFields: true}) {
		return
	}
	ctx, code := workflowRunContextFromRequest(request, trace, body.WorkspaceID, body.ApplicationID, "evaluation-suite-create", "workflow_evaluations:write", "workflow_runs:read")
	if code != "" {
		writeWorkflowEvaluationSuiteResult(writer, trace, ctx, suiteFailure(WorkflowEvaluationSuiteFailureCode(code)))
		return
	}
	writeWorkflowEvaluationSuiteResult(writer, trace, ctx, s.workflowEvaluationSuiteService().Create(ctx, WorkflowEvaluationSuiteCreateRequest{Name: body.Name, CaseRefs: body.CaseRefs}))
}

func (s *Server) handleListWorkflowEvaluationSuites(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowEvaluationSuiteListRoute)
	if !s.allowWorkflowExecutorDev(writer, trace) {
		return
	}
	ctx, limit, cursor, code := workflowEvaluationSuiteListContext(request, trace, "evaluation-suite-list")
	if code != "" {
		writeWorkflowEvaluationSuiteListResult(writer, trace, ctx, suiteListFailure(code))
		return
	}
	writeWorkflowEvaluationSuiteListResult(writer, trace, ctx, s.workflowEvaluationSuiteService().List(ctx, WorkflowEvaluationSuiteListRequest{Limit: limit, Cursor: cursor}))
}

func (s *Server) handleReadWorkflowEvaluationSuite(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowEvaluationSuiteReadRoute)
	if !s.allowWorkflowExecutorDev(writer, trace) {
		return
	}
	ctx, code := workflowEvaluationSuiteReadContext(request, trace, "evaluation-suite-read")
	if code != "" {
		writeWorkflowEvaluationSuiteResult(writer, trace, ctx, suiteFailure(code))
		return
	}
	writeWorkflowEvaluationSuiteResult(writer, trace, ctx, s.workflowEvaluationSuiteService().Read(ctx, request.PathValue("suite_id")))
}

func (s *Server) handleReviewWorkflowEvaluationSuite(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowEvaluationSuiteReviewRoute)
	if !s.allowWorkflowExecutorDev(writer, trace) {
		return
	}
	ctx, code := workflowEvaluationSuiteReadContext(request, trace, "evaluation-suite-review")
	if code != "" {
		writeWorkflowEvaluationSuiteReviewResult(writer, trace, ctx, suiteReviewFailure(code))
		return
	}
	writeWorkflowEvaluationSuiteReviewResult(writer, trace, ctx, s.workflowEvaluationSuiteService().Review(ctx, request.PathValue("suite_id")))
}

func (s *Server) handleCreateWorkflowEvaluationDecision(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowEvaluationDecisionCreateRoute)
	if !s.allowWorkflowExecutorDev(writer, trace) {
		return
	}
	var body workflowEvaluationDecisionHTTPBody
	if !s.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: maxControlJSONRequestBodyBytes, rejectUnknownFields: true}) {
		return
	}
	ctx, code := workflowRunContextFromRequest(request, trace, body.WorkspaceID, body.ApplicationID, "evaluation-suite-decision", "workflow_evaluations:write", "workflow_runs:read")
	if code != "" {
		writeWorkflowEvaluationSuiteResult(writer, trace, ctx, suiteFailure(WorkflowEvaluationSuiteFailureCode(code)))
		return
	}
	writeWorkflowEvaluationSuiteResult(writer, trace, ctx, s.workflowEvaluationSuiteService().Decide(ctx, request.PathValue("suite_id"), WorkflowEvaluationDecisionRequest{ExpectedDecisionVersion: body.ExpectedDecisionVersion, Decision: strings.TrimSpace(body.Decision), ReviewDigest: strings.TrimSpace(body.ReviewDigest)}))
}

func (s *Server) handleListWorkflowEvaluationDecisions(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowEvaluationDecisionListRoute)
	if !s.allowWorkflowExecutorDev(writer, trace) {
		return
	}
	ctx, limit, cursor, code := workflowEvaluationSuiteListContext(request, trace, "evaluation-suite-decision-list")
	if code != "" {
		writeWorkflowEvaluationDecisionListResult(writer, trace, ctx, decisionListFailure(code))
		return
	}
	writeWorkflowEvaluationDecisionListResult(writer, trace, ctx, s.workflowEvaluationSuiteService().ListDecisions(ctx, request.PathValue("suite_id"), WorkflowEvaluationDecisionListRequest{Limit: limit, Cursor: cursor}))
}

func workflowEvaluationSuiteReadContext(request *http.Request, trace requestTrace, suffix string) (WorkflowRunContext, WorkflowEvaluationSuiteFailureCode) {
	values := request.URL.Query()
	for key, entries := range values {
		if (key != "workspace_id" && key != "application_id") || len(entries) != 1 {
			return WorkflowRunContext{}, WorkflowEvaluationSuiteFailureInvalid
		}
	}
	ctx, code := workflowRunContextFromRequest(request, trace, values.Get("workspace_id"), values.Get("application_id"), suffix, "workflow_evaluations:read", "workflow_runs:read")
	return ctx, WorkflowEvaluationSuiteFailureCode(code)
}

func workflowEvaluationSuiteListContext(request *http.Request, trace requestTrace, suffix string) (WorkflowRunContext, int, string, WorkflowEvaluationSuiteFailureCode) {
	values := request.URL.Query()
	for key, entries := range values {
		if (key != "workspace_id" && key != "application_id" && key != "limit" && key != "cursor") || len(entries) != 1 {
			return WorkflowRunContext{}, 0, "", WorkflowEvaluationSuiteFailureInvalid
		}
	}
	ctx, code := workflowRunContextFromRequest(request, trace, values.Get("workspace_id"), values.Get("application_id"), suffix, "workflow_evaluations:read", "workflow_runs:read")
	limit := 0
	if raw := strings.TrimSpace(values.Get("limit")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return ctx, 0, "", WorkflowEvaluationSuiteFailureInvalid
		}
		limit = parsed
	}
	return ctx, limit, values.Get("cursor"), WorkflowEvaluationSuiteFailureCode(code)
}

func workflowEvaluationSuiteFailurePointer(code WorkflowEvaluationSuiteFailureCode) *string {
	if code == "" {
		return nil
	}
	value := string(code)
	return &value
}

func writeWorkflowEvaluationSuiteResult(writer http.ResponseWriter, trace requestTrace, ctx WorkflowRunContext, result WorkflowEvaluationSuiteResult) {
	writeObservedJSON(writer, http.StatusOK, trace, workflowEvaluationSuiteEnvelope{RequestID: trace.requestID, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, Suite: result.Suite, Decision: result.Decision, FailureCode: workflowEvaluationSuiteFailurePointer(result.FailureCode), FailureSummary: result.FailureSummary, AuditRef: ctx.AuditRef})
}

func writeWorkflowEvaluationSuiteReviewResult(writer http.ResponseWriter, trace requestTrace, ctx WorkflowRunContext, result WorkflowEvaluationSuiteReviewResult) {
	writeObservedJSON(writer, http.StatusOK, trace, workflowEvaluationSuiteEnvelope{RequestID: trace.requestID, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, Review: result.Review, FailureCode: workflowEvaluationSuiteFailurePointer(result.FailureCode), FailureSummary: result.FailureSummary, AuditRef: ctx.AuditRef})
}

func writeWorkflowEvaluationSuiteListResult(writer http.ResponseWriter, trace requestTrace, ctx WorkflowRunContext, result WorkflowEvaluationSuiteListResult) {
	if result.Suites == nil {
		result.Suites = []WorkflowEvaluationSuite{}
	}
	writeObservedJSON(writer, http.StatusOK, trace, workflowEvaluationSuiteListEnvelope{RequestID: trace.requestID, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, Suites: result.Suites, NextCursor: result.NextCursor, HasMore: result.HasMore, FailureCode: workflowEvaluationSuiteFailurePointer(result.FailureCode), FailureSummary: result.FailureSummary, AuditRef: ctx.AuditRef})
}

func writeWorkflowEvaluationDecisionListResult(writer http.ResponseWriter, trace requestTrace, ctx WorkflowRunContext, result WorkflowEvaluationDecisionListResult) {
	if result.Decisions == nil {
		result.Decisions = []WorkflowEvaluationReleaseDecision{}
	}
	writeObservedJSON(writer, http.StatusOK, trace, workflowEvaluationDecisionListEnvelope{RequestID: trace.requestID, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, Decisions: result.Decisions, NextCursor: result.NextCursor, HasMore: result.HasMore, FailureCode: workflowEvaluationSuiteFailurePointer(result.FailureCode), FailureSummary: result.FailureSummary, AuditRef: ctx.AuditRef})
}
