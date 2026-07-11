package httpapi

import (
	"context"
	"net/http"
	"strings"
)

const (
	workflowExecutorStartRoute = "POST /v1/user-workspace/workflow-drafts/{draft_id}/runs"
	workflowRunListRoute       = "GET /v1/user-workspace/workflow-runs"
	workflowRunReadRoute       = "GET /v1/user-workspace/workflow-runs/{run_id}"
	workflowRunComparisonRoute = "GET /v1/user-workspace/workflow-runs/{candidate_run_id}/comparison"
)

type workflowRunStartHTTPBody struct {
	WorkspaceID        string                        `json:"workspace_id"`
	ApplicationID      string                        `json:"application_id"`
	InputText          string                        `json:"input_text"`
	ConditionValues    map[string]bool               `json:"condition_values"`
	Model              string                        `json:"model"`
	Temperature        *float64                      `json:"temperature"`
	DevFailureScenario WorkflowRunDevFailureScenario `json:"dev_failure_scenario"`
}

type workflowRunEnvelope struct {
	RequestID      string             `json:"request_id"`
	WorkspaceID    string             `json:"workspace_id"`
	ApplicationID  string             `json:"application_id"`
	Run            *WorkflowRunRecord `json:"run"`
	FailureCode    *string            `json:"failure_code"`
	FailureSummary string             `json:"failure_summary"`
	AuditRef       string             `json:"audit_ref"`
}

type workflowRunListEnvelope struct {
	RequestID      string               `json:"request_id"`
	WorkspaceID    string               `json:"workspace_id"`
	ApplicationID  string               `json:"application_id"`
	Runs           []WorkflowRunSummary `json:"runs"`
	NextCursor     string               `json:"next_cursor"`
	HasMore        bool                 `json:"has_more"`
	FailureCode    *string              `json:"failure_code"`
	FailureSummary string               `json:"failure_summary"`
	AuditRef       string               `json:"audit_ref"`
}

type workflowRunComparisonEnvelope struct {
	RequestID      string                 `json:"request_id"`
	WorkspaceID    string                 `json:"workspace_id"`
	ApplicationID  string                 `json:"application_id"`
	Comparison     *WorkflowRunComparison `json:"comparison"`
	FailureCode    *string                `json:"failure_code"`
	FailureSummary string                 `json:"failure_summary"`
	AuditRef       string                 `json:"audit_ref"`
}

func (s *Server) handleStartWorkflowRun(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowExecutorStartRoute)
	if !s.allowWorkflowExecutorDev(writer, trace) {
		return
	}
	var body workflowRunStartHTTPBody
	if !s.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{
		maxBytes:            maxControlJSONRequestBodyBytes,
		rejectUnknownFields: true,
	}) {
		return
	}
	runContext, failureCode := workflowRunContextFromRequest(
		request,
		trace,
		body.WorkspaceID,
		body.ApplicationID,
		"start",
		"workflow_runs:execute",
		"workflow_drafts:read",
	)
	if failureCode != "" {
		writeWorkflowRunResult(writer, trace, runContext, workflowRunFailure(failureCode, "Workflow run scope is denied."))
		return
	}
	result := s.workflowExecutorService().StartRun(runContext, WorkflowRunRequest{
		DraftID:            strings.TrimSpace(request.PathValue("draft_id")),
		InputText:          body.InputText,
		ConditionValues:    body.ConditionValues,
		Model:              body.Model,
		Temperature:        body.Temperature,
		DevFailureScenario: body.DevFailureScenario,
	})
	writeWorkflowRunResult(writer, trace, runContext, result)
}

func (s *Server) handleReadWorkflowRun(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowRunReadRoute)
	if !s.allowWorkflowExecutorDev(writer, trace) {
		return
	}
	workspaceID := strings.TrimSpace(request.URL.Query().Get("workspace_id"))
	applicationID := strings.TrimSpace(request.URL.Query().Get("application_id"))
	runContext, failureCode := workflowRunContextFromRequest(
		request,
		trace,
		workspaceID,
		applicationID,
		"read",
		"workflow_runs:read",
	)
	if failureCode != "" {
		writeWorkflowRunResult(writer, trace, runContext, workflowRunFailure(failureCode, "Workflow run scope is denied."))
		return
	}
	result := s.workflowExecutorService().ReadRun(runContext, strings.TrimSpace(request.PathValue("run_id")))
	writeWorkflowRunResult(writer, trace, runContext, result)
}

func (s *Server) handleListWorkflowRuns(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowRunListRoute)
	if !s.allowWorkflowExecutorDev(writer, trace) {
		return
	}
	workspaceID := strings.TrimSpace(request.URL.Query().Get("workspace_id"))
	applicationID := strings.TrimSpace(request.URL.Query().Get("application_id"))
	runContext, failureCode := workflowRunContextFromRequest(
		request, trace, workspaceID, applicationID, "list", "workflow_runs:read",
	)
	if failureCode != "" {
		writeWorkflowRunListResult(writer, trace, runContext, workflowRunListFailure(failureCode))
		return
	}
	listRequest, failureCode := parseWorkflowRunListRequest(request.URL.Query())
	if failureCode != "" {
		writeWorkflowRunListResult(writer, trace, runContext, workflowRunListFailure(failureCode))
		return
	}
	writeWorkflowRunListResult(writer, trace, runContext, s.workflowExecutorService().ListRuns(runContext, listRequest))
}

func (s *Server) handleCompareWorkflowRuns(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowRunComparisonRoute)
	if !s.allowWorkflowExecutorDev(writer, trace) {
		return
	}
	values := request.URL.Query()
	allowed := map[string]bool{"workspace_id": true, "application_id": true, "baseline_run_id": true}
	for key, entries := range values {
		if !allowed[key] || len(entries) != 1 {
			writeWorkflowRunComparisonResult(writer, trace, WorkflowRunContext{}, workflowRunComparisonFailure(WorkflowRunFailureComparisonInvalid))
			return
		}
	}
	workspaceID, applicationID := strings.TrimSpace(values.Get("workspace_id")), strings.TrimSpace(values.Get("application_id"))
	runContext, failureCode := workflowRunContextFromRequest(request, trace, workspaceID, applicationID, "compare", "workflow_runs:read")
	if failureCode != "" {
		writeWorkflowRunComparisonResult(writer, trace, runContext, workflowRunComparisonFailure(failureCode))
		return
	}
	result := s.workflowExecutorService().CompareRuns(runContext, values.Get("baseline_run_id"), request.PathValue("candidate_run_id"))
	writeWorkflowRunComparisonResult(writer, trace, runContext, result)
}

func (s *Server) allowWorkflowExecutorDev(writer http.ResponseWriter, trace requestTrace) bool {
	if s.config.WorkflowExecutorDevEnabled {
		return true
	}
	s.writePlatformError(writer, trace, "WORKFLOW_EXECUTOR_DEV_DISABLED", "workflow executor dev route requires explicit opt-in")
	return false
}

func (s *Server) workflowExecutorService() workflowExecutorService {
	if s.workflowRunStore == nil {
		s.workflowRunStore = newMemoryWorkflowRunStore(defaultWorkflowRunStoreCapacity)
	}
	service := newWorkflowExecutorService(
		s.savedWorkflowDraftService().ReadDraft,
		s.bridge,
		s.workflowRunStore,
	)
	service.defaultTemperature = s.config.Temperature
	service.resolveSelection = func(ctx context.Context, requestedModel string) northboundSelection {
		return s.resolveNorthboundSelection(ctx, requestedModel, nil)
	}
	service.envelopeOptions = s.buildBridgeEnvelopeOptions
	service.diagnosticsDevEnabled = s.config.WorkflowDiagnosticsDevEnabled && strings.EqualFold(strings.TrimSpace(s.config.Provider), "mock")
	return service
}

func (s *Server) workflowEvaluationService() workflowEvaluationService {
	if s.workflowEvaluationStore == nil {
		s.workflowEvaluationStore = newWorkflowEvaluationStoreForRunStore(s.workflowRunStore)
	}
	return newWorkflowEvaluationService(s.workflowEvaluationStore, s.workflowRunStore)
}

func (s *Server) workflowEvaluationSuiteService() workflowEvaluationSuiteService {
	if s.workflowEvaluationSuiteStore == nil {
		s.workflowEvaluationSuiteStore = newWorkflowEvaluationSuiteStoreForRunStore(s.workflowRunStore)
	}
	return newWorkflowEvaluationSuiteService(s.workflowEvaluationSuiteStore, s.workflowEvaluationService())
}

func workflowRunContextFromRequest(
	request *http.Request,
	trace requestTrace,
	workspaceID string,
	applicationID string,
	auditSuffix string,
	requiredScopes ...string,
) (WorkflowRunContext, WorkflowRunFailureCode) {
	runContext := WorkflowRunContext{
		RequestContext: request.Context(),
		RequestID:      trace.requestID,
		WorkspaceID:    strings.TrimSpace(workspaceID),
		ApplicationID:  strings.TrimSpace(applicationID),
		AuditRef:       auditRefForWorkflowRun(trace, auditSuffix),
	}
	auth, ok := request.Context().Value(controlPlaneReadAuthContextKey{}).(controlPlaneReadAuthContext)
	if !ok || strings.TrimSpace(auth.IdentityContext) == "" || strings.TrimSpace(auth.SubjectBinding) == "" ||
		strings.TrimSpace(auth.TenantBinding) == "" {
		return runContext, WorkflowRunFailureScopeDenied
	}
	runContext.TenantRef = strings.TrimSpace(auth.TenantBinding)
	runContext.ActorRef = strings.TrimSpace(auth.SubjectBinding)
	runContext.ScopeGrants = cloneStringSlice(auth.ScopeGrants)
	for _, requiredScope := range requiredScopes {
		if !controlPlaneReadHasScope(auth.ScopeGrants, requiredScope) {
			return runContext, WorkflowRunFailureScopeDenied
		}
	}
	headerWorkspaceID := strings.TrimSpace(request.Header.Get(savedWorkflowDraftDevWorkspaceHeader))
	headerApplicationID := strings.TrimSpace(request.Header.Get(savedWorkflowDraftDevApplicationHeader))
	if runContext.WorkspaceID == "" || runContext.ApplicationID == "" ||
		headerWorkspaceID != runContext.WorkspaceID || headerApplicationID != runContext.ApplicationID {
		return runContext, WorkflowRunFailureScopeDenied
	}
	return runContext, ""
}

func writeWorkflowRunResult(
	writer http.ResponseWriter,
	trace requestTrace,
	runContext WorkflowRunContext,
	result WorkflowRunResult,
) {
	writeObservedJSON(writer, http.StatusOK, trace, workflowRunEnvelope{
		RequestID:      trace.requestID,
		WorkspaceID:    runContext.WorkspaceID,
		ApplicationID:  runContext.ApplicationID,
		Run:            result.Record,
		FailureCode:    workflowRunFailureCodePointer(result.FailureCode),
		FailureSummary: result.FailureSummary,
		AuditRef:       runContext.AuditRef,
	})
}

func workflowRunFailureCodePointer(failureCode WorkflowRunFailureCode) *string {
	if failureCode == "" {
		return nil
	}
	value := string(failureCode)
	return &value
}

func writeWorkflowRunListResult(
	writer http.ResponseWriter,
	trace requestTrace,
	runContext WorkflowRunContext,
	result WorkflowRunListResult,
) {
	if result.Runs == nil {
		result.Runs = []WorkflowRunSummary{}
	}
	writeObservedJSON(writer, http.StatusOK, trace, workflowRunListEnvelope{
		RequestID: trace.requestID, WorkspaceID: runContext.WorkspaceID, ApplicationID: runContext.ApplicationID,
		Runs: result.Runs, NextCursor: result.NextCursor, HasMore: result.HasMore,
		FailureCode: workflowRunFailureCodePointer(result.FailureCode), FailureSummary: result.FailureSummary,
		AuditRef: runContext.AuditRef,
	})
}

func writeWorkflowRunComparisonResult(writer http.ResponseWriter, trace requestTrace, runContext WorkflowRunContext, result WorkflowRunComparisonResult) {
	writeObservedJSON(writer, http.StatusOK, trace, workflowRunComparisonEnvelope{RequestID: trace.requestID, WorkspaceID: runContext.WorkspaceID, ApplicationID: runContext.ApplicationID, Comparison: result.Comparison, FailureCode: workflowRunFailureCodePointer(result.FailureCode), FailureSummary: result.FailureSummary, AuditRef: runContext.AuditRef})
}

func auditRefForWorkflowRun(trace requestTrace, suffix string) string {
	return strings.TrimSpace("audit_" + trace.requestID + "_workflow-run-" + suffix)
}
