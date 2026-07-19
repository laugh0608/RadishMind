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
	RequestID                 string                       `json:"request_id"`
	WorkspaceID               string                       `json:"workspace_id"`
	ApplicationID             string                       `json:"application_id"`
	Run                       *WorkflowRunRecord           `json:"run"`
	RetrievalAnswer           *WorkflowRAGAnswer           `json:"retrieval_answer,omitempty"`
	AdvisoryOutput            string                       `json:"advisory_output,omitempty"`
	FailureCode               *string                      `json:"failure_code"`
	FailureSummary            string                       `json:"failure_summary"`
	AuditRef                  string                       `json:"audit_ref"`
	RetrievalFragmentPreviews []WorkflowRAGFragmentPreview `json:"retrieval_fragment_previews,omitempty"`
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
	if !s.allowWorkflowRunHistoryDev(writer, trace) {
		return
	}
	values := request.URL.Query()
	for key, entries := range values {
		if (key != "workspace_id" && key != "application_id" && key != "include_retrieval_fragment_previews") || len(entries) != 1 {
			writeWorkflowRunResult(writer, trace, WorkflowRunContext{}, workflowRunFailure(WorkflowRunFailureInputInvalid, "Workflow run detail query is invalid."))
			return
		}
	}
	includePreviews := strings.TrimSpace(values.Get("include_retrieval_fragment_previews")) == "true"
	if raw := strings.TrimSpace(values.Get("include_retrieval_fragment_previews")); raw != "" && raw != "true" && raw != "false" {
		writeWorkflowRunResult(writer, trace, WorkflowRunContext{}, workflowRunFailure(WorkflowRunFailureInputInvalid, "Workflow run detail query is invalid."))
		return
	}
	workspaceID := strings.TrimSpace(values.Get("workspace_id"))
	applicationID := strings.TrimSpace(values.Get("application_id"))
	requiredScopes := []string{"workflow_runs:read"}
	if includePreviews {
		requiredScopes = append(requiredScopes, "workflow_rag_snapshots:read")
	}
	runContext, failureCode := workflowRunContextFromRequest(
		request,
		trace,
		workspaceID,
		applicationID,
		"read",
		requiredScopes...,
	)
	if failureCode != "" {
		writeWorkflowRunResult(writer, trace, runContext, workflowRunFailure(failureCode, "Workflow run scope is denied."))
		return
	}
	if s.config.WorkflowRAGExecutionDevEnabled {
		if reconciled := s.workflowRAGExecutionService().ReconcileStale(runContext); reconciled.FailureCode != "" {
			writeWorkflowRunResult(writer, trace, runContext, workflowRAGExecutionFailure(reconciled.FailureCode))
			return
		}
	}
	if s.config.WorkflowDefinitionReleaseDevEnabled && s.config.WorkflowExecutorDevEnabled {
		if reconciled := s.workflowDefinitionExecutionService().ReconcileStale(runContext); reconciled.FailureCode != "" {
			writeWorkflowRunResult(writer, trace, runContext, reconciled)
			return
		}
	}
	result := s.workflowExecutorService().ReadRun(runContext, strings.TrimSpace(request.PathValue("run_id")))
	previews := []WorkflowRAGFragmentPreview{}
	if includePreviews && result.Record != nil {
		var previewFailure string
		previews, previewFailure = s.workflowRAGHistoryPreviews(runContext, *result.Record)
		if previewFailure != "" {
			writeWorkflowRunResult(writer, trace, runContext, workflowRAGExecutionFailure(previewFailure))
			return
		}
	}
	writeWorkflowRunHistoryResult(writer, trace, runContext, result, previews)
}

func (s *Server) handleListWorkflowRuns(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowRunListRoute)
	if !s.allowWorkflowRunHistoryDev(writer, trace) {
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
	if s.config.WorkflowRAGExecutionDevEnabled {
		if reconciled := s.workflowRAGExecutionService().ReconcileStale(runContext); reconciled.FailureCode != "" {
			writeWorkflowRunListResult(writer, trace, runContext, workflowRunListFailure(WorkflowRunFailureStoreUnavailable))
			return
		}
	}
	if s.config.WorkflowDefinitionReleaseDevEnabled && s.config.WorkflowExecutorDevEnabled {
		if reconciled := s.workflowDefinitionExecutionService().ReconcileStale(runContext); reconciled.FailureCode != "" {
			writeWorkflowRunListResult(writer, trace, runContext, workflowRunListFailure(reconciled.FailureCode))
			return
		}
	}
	writeWorkflowRunListResult(writer, trace, runContext, s.workflowExecutorService().ListRuns(runContext, listRequest))
}

func (s *Server) handleCompareWorkflowRuns(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowRunComparisonRoute)
	if !s.allowWorkflowRunHistoryDev(writer, trace) {
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

func (s *Server) allowWorkflowRunHistoryDev(writer http.ResponseWriter, trace requestTrace) bool {
	if s.config.WorkflowExecutorDevEnabled || s.config.WorkflowRAGExecutionDevEnabled {
		return true
	}
	s.writePlatformError(writer, trace, "WORKFLOW_EXECUTOR_DEV_DISABLED", "workflow run history requires an explicit workflow execution development gate")
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
		RequestID:       trace.requestID,
		WorkspaceID:     runContext.WorkspaceID,
		ApplicationID:   runContext.ApplicationID,
		Run:             result.Record,
		RetrievalAnswer: result.RetrievalAnswer,
		AdvisoryOutput:  result.AdvisoryOutput,
		FailureCode:     workflowRunFailureCodePointer(result.FailureCode),
		FailureSummary:  result.FailureSummary,
		AuditRef:        runContext.AuditRef,
	})
}

func writeWorkflowRunHistoryResult(
	writer http.ResponseWriter,
	trace requestTrace,
	runContext WorkflowRunContext,
	result WorkflowRunResult,
	previews []WorkflowRAGFragmentPreview,
) {
	writeObservedJSON(writer, http.StatusOK, trace, workflowRunEnvelope{
		RequestID: trace.requestID, WorkspaceID: runContext.WorkspaceID, ApplicationID: runContext.ApplicationID,
		Run: result.Record, FailureCode: workflowRunFailureCodePointer(result.FailureCode),
		FailureSummary: result.FailureSummary, AuditRef: runContext.AuditRef, RetrievalFragmentPreviews: previews,
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
