package httpapi

import (
	"strings"
	"time"
)

const (
	workflowRunComparisonSchemaVersion       = "workflow_run_comparison.v1"
	workflowRAGRunComparisonSchemaVersion    = "workflow_run_comparison.v2"
	workflowRAGAppRunComparisonSchemaVersion = "workflow_run_comparison.v3"
)

type WorkflowRunComparisonClassification string

const (
	WorkflowRunComparisonRegression   WorkflowRunComparisonClassification = "regression"
	WorkflowRunComparisonImprovement  WorkflowRunComparisonClassification = "improvement"
	WorkflowRunComparisonChanged      WorkflowRunComparisonClassification = "changed"
	WorkflowRunComparisonUnchanged    WorkflowRunComparisonClassification = "unchanged"
	WorkflowRunComparisonInconclusive WorkflowRunComparisonClassification = "inconclusive"
)

type WorkflowRunComparisonState string

const (
	WorkflowRunComparisonComparable          WorkflowRunComparisonState = "comparable"
	WorkflowRunComparisonLegacyPartial       WorkflowRunComparisonState = "legacy_partial"
	WorkflowRunComparisonRunningInconclusive WorkflowRunComparisonState = "running_inconclusive"
)

type WorkflowRunComparisonFinding struct {
	Code     string `json:"code"`
	Severity string `json:"severity"`
}

type WorkflowRunComparisonRun struct {
	RunID                  string                            `json:"run_id"`
	SchemaVersion          string                            `json:"schema_version"`
	DraftID                string                            `json:"draft_id"`
	DraftVersion           int                               `json:"draft_version"`
	ExecutionKind          string                            `json:"execution_kind,omitempty"`
	ExecutionSourceKind    string                            `json:"execution_source_kind,omitempty"`
	ExecutionSourceID      string                            `json:"execution_source_id,omitempty"`
	ExecutionSourceVersion int                               `json:"execution_source_version,omitempty"`
	Status                 WorkflowRunStatus                 `json:"status"`
	FailureCode            WorkflowRunFailureCode            `json:"failure_code"`
	FailureBoundary        WorkflowRunFailureBoundary        `json:"failure_boundary"`
	GatewayFailureCategory WorkflowRunGatewayFailureCategory `json:"gateway_failure_category"`
	SelectedProvider       string                            `json:"selected_provider"`
	SelectedProfile        string                            `json:"selected_profile"`
	SelectedModel          string                            `json:"selected_model"`
	DurationMS             int64                             `json:"duration_ms"`
	StaleRunning           bool                              `json:"stale_running"`
	RequestID              string                            `json:"request_id"`
	AuditRef               string                            `json:"audit_ref"`
	SideEffects            WorkflowRunSideEffects            `json:"side_effects"`
}

type WorkflowRunNodeComparison struct {
	NodeID              string                `json:"node_id"`
	NodeType            string                `json:"node_type"`
	Change              string                `json:"change"`
	BaselineStatus      WorkflowRunNodeStatus `json:"baseline_status"`
	CandidateStatus     WorkflowRunNodeStatus `json:"candidate_status"`
	BaselineDurationMS  int64                 `json:"baseline_duration_ms"`
	CandidateDurationMS int64                 `json:"candidate_duration_ms"`
	DurationDeltaMS     int64                 `json:"duration_delta_ms"`
}

type WorkflowRunComparison struct {
	SchemaVersion           string                              `json:"schema_version"`
	Classification          WorkflowRunComparisonClassification `json:"classification"`
	ComparisonState         WorkflowRunComparisonState          `json:"comparison_state"`
	Baseline                WorkflowRunComparisonRun            `json:"baseline"`
	Candidate               WorkflowRunComparisonRun            `json:"candidate"`
	DraftChanged            bool                                `json:"draft_changed"`
	ExecutionSourceChanged  bool                                `json:"execution_source_changed"`
	ProviderChanged         bool                                `json:"provider_changed"`
	ModelChanged            bool                                `json:"model_changed"`
	StatusChanged           bool                                `json:"status_changed"`
	FailureChanged          bool                                `json:"failure_changed"`
	DurationDeltaMS         int64                               `json:"duration_delta_ms"`
	ProviderCallDelta       int                                 `json:"provider_call_delta"`
	Nodes                   []WorkflowRunNodeComparison         `json:"nodes"`
	Retrieval               *WorkflowRunRetrievalComparison     `json:"retrieval,omitempty"`
	Findings                []WorkflowRunComparisonFinding      `json:"findings"`
	RecommendedReviewAction WorkflowRunReviewAction             `json:"recommended_review_action"`
}

type WorkflowRunComparisonResult struct {
	Comparison     *WorkflowRunComparison
	FailureCode    WorkflowRunFailureCode
	FailureSummary string
}

func (service workflowExecutorService) CompareRuns(runContext WorkflowRunContext, baselineRunID, candidateRunID string) WorkflowRunComparisonResult {
	baselineRunID, candidateRunID = strings.TrimSpace(baselineRunID), strings.TrimSpace(candidateRunID)
	if baselineRunID == "" || candidateRunID == "" || baselineRunID == candidateRunID || len([]rune(baselineRunID)) > 160 || len([]rune(candidateRunID)) > 160 {
		return workflowRunComparisonFailure(WorkflowRunFailureComparisonInvalid)
	}
	baseline, result := service.readRunForComparison(runContext, baselineRunID)
	if result.FailureCode != "" {
		return result
	}
	candidate, result := service.readRunForComparison(runContext, candidateRunID)
	if result.FailureCode != "" {
		return result
	}
	if workflowRunRecordUsesRetrievalComparison(baseline) || workflowRunRecordUsesRetrievalComparison(candidate) {
		if baseline.SchemaVersion != candidate.SchemaVersion ||
			!workflowRAGRunsComparable(baseline, candidate) {
			return workflowRunComparisonFailure(WorkflowRunFailureRetrievalIncompatible)
		}
		comparison := buildWorkflowRAGRunComparison(baseline, candidate, time.Now().UTC())
		return WorkflowRunComparisonResult{Comparison: &comparison}
	}
	if workflowRunComparisonSideEffectProfileUnsupported(baseline) || workflowRunComparisonSideEffectProfileUnsupported(candidate) {
		return workflowRunComparisonFailure(WorkflowRunFailureSideEffectUnsupported)
	}
	if !workflowRunComparisonSideEffectsValid(baseline.SideEffects) || !workflowRunComparisonSideEffectsValid(candidate.SideEffects) {
		return workflowRunComparisonFailure(WorkflowRunFailureStoreContractMismatch)
	}
	comparison := buildWorkflowRunComparison(baseline, candidate, time.Now().UTC())
	return WorkflowRunComparisonResult{Comparison: &comparison}
}

func (service workflowExecutorService) readRunForComparison(ctx WorkflowRunContext, runID string) (WorkflowRunRecord, WorkflowRunComparisonResult) {
	record, found, err := service.store.ReadRun(ctx, runID)
	if err != nil {
		return WorkflowRunRecord{}, workflowRunComparisonFailure(workflowRunStoreFailureCode(err))
	}
	if !found {
		return WorkflowRunRecord{}, workflowRunComparisonFailure(WorkflowRunFailureRecordNotFound)
	}
	return record, WorkflowRunComparisonResult{}
}

func buildWorkflowRunComparison(baseline, candidate WorkflowRunRecord, now time.Time) WorkflowRunComparison {
	b := summarizeWorkflowRun(baseline, now)
	c := summarizeWorkflowRun(candidate, now)
	comparison := WorkflowRunComparison{
		SchemaVersion: workflowRunComparisonSchemaVersion,
		Baseline:      comparisonRun(b), Candidate: comparisonRun(c),
		DraftChanged:           baseline.DraftID != candidate.DraftID || baseline.DraftVersion != candidate.DraftVersion || baseline.DraftDigest != candidate.DraftDigest,
		ExecutionSourceChanged: workflowRunExecutionSourceChanged(baseline, candidate),
		ProviderChanged:        baseline.SelectedProvider != candidate.SelectedProvider || baseline.SelectedProfile != candidate.SelectedProfile,
		ModelChanged:           baseline.SelectedModel != candidate.SelectedModel,
		StatusChanged:          baseline.Status != candidate.Status,
		FailureChanged:         baseline.FailureCode != candidate.FailureCode || diagnosticBoundary(baseline) != diagnosticBoundary(candidate) || diagnosticCategory(baseline) != diagnosticCategory(candidate),
		DurationDeltaMS:        c.DurationMS - b.DurationMS,
		ProviderCallDelta:      candidate.SideEffects.ProviderCalls - baseline.SideEffects.ProviderCalls,
		Nodes:                  compareWorkflowRunNodes(baseline.Nodes, candidate.Nodes),
	}
	comparison.ComparisonState = WorkflowRunComparisonComparable
	if (baseline.Status == WorkflowRunStatusRunning && !b.StaleRunning) || (candidate.Status == WorkflowRunStatusRunning && !c.StaleRunning) {
		comparison.ComparisonState = WorkflowRunComparisonRunningInconclusive
	} else if baseline.Diagnostic == nil || candidate.Diagnostic == nil {
		comparison.ComparisonState = WorkflowRunComparisonLegacyPartial
	}
	comparison.Classification = classifyWorkflowRunComparison(comparison)
	comparison.Findings = workflowRunComparisonFindings(comparison)
	comparison.RecommendedReviewAction = comparisonReviewAction(comparison, candidate)
	return comparison
}

func comparisonRun(summary WorkflowRunSummary) WorkflowRunComparisonRun {
	return WorkflowRunComparisonRun{RunID: summary.RunID, SchemaVersion: summary.SchemaVersion, DraftID: summary.DraftID, DraftVersion: summary.DraftVersion, ExecutionKind: summary.ExecutionKind, ExecutionSourceKind: summary.ExecutionSourceKind, ExecutionSourceID: summary.ExecutionSourceID, ExecutionSourceVersion: summary.ExecutionSourceVersion, Status: summary.Status, FailureCode: summary.FailureCode, FailureBoundary: summary.FailureBoundary, GatewayFailureCategory: summary.GatewayFailureCategory, SelectedProvider: summary.SelectedProvider, SelectedProfile: summary.SelectedProfile, SelectedModel: summary.SelectedModel, DurationMS: summary.DurationMS, StaleRunning: summary.StaleRunning, RequestID: summary.RequestID, AuditRef: summary.AuditRef, SideEffects: summary.SideEffects}
}

func workflowRunRecordUsesRetrievalComparison(record WorkflowRunRecord) bool {
	return record.SchemaVersion == workflowRunRecordRAGSchemaVersion || record.SchemaVersion == workflowRunRecordAppRAGSchemaVersion
}

func workflowRunExecutionSourceChanged(baseline, candidate WorkflowRunRecord) bool {
	if baseline.ExecutionSource == nil || candidate.ExecutionSource == nil {
		return baseline.ExecutionSource != candidate.ExecutionSource
	}
	return *baseline.ExecutionSource != *candidate.ExecutionSource
}

func classifyWorkflowRunComparison(value WorkflowRunComparison) WorkflowRunComparisonClassification {
	if value.ComparisonState == WorkflowRunComparisonRunningInconclusive {
		return WorkflowRunComparisonInconclusive
	}
	baselineGood := value.Baseline.Status == WorkflowRunStatusSucceeded && !value.Baseline.StaleRunning
	candidateGood := value.Candidate.Status == WorkflowRunStatusSucceeded && !value.Candidate.StaleRunning
	if baselineGood && !candidateGood {
		return WorkflowRunComparisonRegression
	}
	if !baselineGood && candidateGood {
		return WorkflowRunComparisonImprovement
	}
	if value.Retrieval != nil && value.Baseline.Status == WorkflowRunStatusSucceeded && value.Candidate.Status == WorkflowRunStatusSucceeded && len(value.Retrieval.CitationRemovedRefs) > 0 {
		return WorkflowRunComparisonRegression
	}
	durationChanged := value.DurationDeltaMS != 0 && value.Retrieval == nil
	if value.DraftChanged || value.ExecutionSourceChanged || value.ProviderChanged || value.ModelChanged || value.StatusChanged || value.FailureChanged || durationChanged || value.ProviderCallDelta != 0 || nodesChanged(value.Nodes) || workflowRAGComparisonMateriallyChanged(value.Retrieval) {
		return WorkflowRunComparisonChanged
	}
	return WorkflowRunComparisonUnchanged
}

func compareWorkflowRunNodes(baseline, candidate []WorkflowRunNodeRecord) []WorkflowRunNodeComparison {
	left, right := make(map[string]WorkflowRunNodeRecord, len(baseline)), make(map[string]WorkflowRunNodeRecord, len(candidate))
	order := make([]string, 0, len(baseline)+len(candidate))
	for _, node := range baseline {
		left[node.NodeID] = node
		order = append(order, node.NodeID)
	}
	for _, node := range candidate {
		right[node.NodeID] = node
		if _, ok := left[node.NodeID]; !ok {
			order = append(order, node.NodeID)
		}
	}
	result := make([]WorkflowRunNodeComparison, 0, len(order))
	for _, id := range order {
		b, hasB := left[id]
		c, hasC := right[id]
		change := "unchanged"
		if !hasB {
			change = "added"
		} else if !hasC {
			change = "removed"
		} else if b.NodeType != c.NodeType || b.Status != c.Status || b.DurationMS != c.DurationMS || b.FailureCode != c.FailureCode {
			change = "changed"
		}
		nodeType := c.NodeType
		if nodeType == "" {
			nodeType = b.NodeType
		}
		result = append(result, WorkflowRunNodeComparison{NodeID: id, NodeType: nodeType, Change: change, BaselineStatus: b.Status, CandidateStatus: c.Status, BaselineDurationMS: b.DurationMS, CandidateDurationMS: c.DurationMS, DurationDeltaMS: c.DurationMS - b.DurationMS})
	}
	return result
}

func workflowRunComparisonFindings(value WorkflowRunComparison) []WorkflowRunComparisonFinding {
	findings := make([]WorkflowRunComparisonFinding, 0, 8)
	add := func(code, severity string) {
		findings = append(findings, WorkflowRunComparisonFinding{Code: code, Severity: severity})
	}
	switch value.Classification {
	case WorkflowRunComparisonRegression:
		add("status_regressed", "review_required")
	case WorkflowRunComparisonImprovement:
		add("status_improved", "info")
	case WorkflowRunComparisonInconclusive:
		add("running_inconclusive", "review_required")
	}
	if value.ComparisonState == WorkflowRunComparisonLegacyPartial {
		add("legacy_diagnostic_unavailable", "review_required")
	}
	if value.DraftChanged {
		add("draft_changed", "info")
	}
	if value.ProviderChanged {
		add("provider_changed", "info")
	}
	if value.ModelChanged {
		add("model_changed", "info")
	}
	if value.FailureChanged {
		add("failure_changed", "review_required")
	}
	if nodesChanged(value.Nodes) {
		add("nodes_changed", "review_required")
	}
	if value.Retrieval != nil {
		if value.Retrieval.AuthorityChanged {
			add("application_rag_authority_changed", "review_required")
		}
		if len(value.Retrieval.CitationRemovedRefs) > 0 {
			add("retrieval_citations_removed", "review_required")
		}
		if len(value.Retrieval.CitationAddedRefs) > 0 {
			add("retrieval_citations_added", "info")
		}
		if value.Retrieval.EvidenceChanged {
			add("retrieval_evidence_changed", "review_required")
		}
		if value.Retrieval.RankingChanged {
			add("retrieval_ranking_changed", "review_required")
		}
		if value.Retrieval.CandidateCountDelta != 0 {
			add("retrieval_candidate_count_changed", "info")
		}
		if value.Retrieval.ContextBytesDelta != 0 {
			add("retrieval_context_changed", "info")
		}
		if value.Retrieval.LatencyDeltaMS > 0 {
			add("retrieval_latency_increased", "info")
		} else if value.Retrieval.LatencyDeltaMS < 0 {
			add("retrieval_latency_decreased", "info")
		}
	}
	if value.DurationDeltaMS > 0 {
		add("duration_increased", "info")
	} else if value.DurationDeltaMS < 0 {
		add("duration_decreased", "info")
	}
	if len(findings) == 0 {
		add("no_material_change", "info")
	}
	return findings
}

func comparisonReviewAction(value WorkflowRunComparison, candidate WorkflowRunRecord) WorkflowRunReviewAction {
	if value.Classification == WorkflowRunComparisonInconclusive {
		return WorkflowRunReviewStartNewRun
	}
	if candidate.Status == WorkflowRunStatusSucceeded && workflowRAGComparisonMateriallyChanged(value.Retrieval) {
		return WorkflowRunReviewRetrievalEvidence
	}
	if candidate.Diagnostic != nil && candidate.Diagnostic.RecommendedReviewAction != "" {
		return candidate.Diagnostic.RecommendedReviewAction
	}
	if value.DraftChanged || nodesChanged(value.Nodes) {
		return WorkflowRunReviewDraft
	}
	return ""
}

func nodesChanged(nodes []WorkflowRunNodeComparison) bool {
	for _, node := range nodes {
		if node.Change != "unchanged" {
			return true
		}
	}
	return false
}
func diagnosticBoundary(record WorkflowRunRecord) WorkflowRunFailureBoundary {
	if record.Diagnostic != nil {
		return record.Diagnostic.FailureBoundary
	}
	return ""
}
func diagnosticCategory(record WorkflowRunRecord) WorkflowRunGatewayFailureCategory {
	if record.Diagnostic != nil {
		return record.Diagnostic.GatewayFailureCategory
	}
	return ""
}
func workflowRunComparisonSideEffectsValid(value WorkflowRunSideEffects) bool {
	return value.ToolCalls == 0 && value.ConfirmationCalls == 0 && value.BusinessWrites == 0 && value.ReplayWrites == 0
}

func workflowRunComparisonSideEffectProfileUnsupported(record WorkflowRunRecord) bool {
	return record.SchemaVersion == workflowRunRecordToolSchemaVersion ||
		record.SideEffects.ToolCalls > 0 || record.SideEffects.ConfirmationCalls > 0
}

func workflowRunComparisonFailure(code WorkflowRunFailureCode) WorkflowRunComparisonResult {
	summary := "Workflow run comparison request is invalid."
	if code == WorkflowRunFailureRecordNotFound {
		summary = "Workflow run record was not found in the current scope."
	}
	if code == WorkflowRunFailureStoreUnavailable || code == WorkflowRunFailureStoreContractMismatch {
		summary = "Workflow run comparison storage is unavailable."
	}
	if code == WorkflowRunFailureSideEffectUnsupported {
		summary = "Workflow run comparison does not support controlled tool side-effect profiles."
	}
	if code == WorkflowRunFailureRetrievalUnsupported {
		summary = "Workflow run comparison does not support workflow retrieval profiles."
	}
	if code == WorkflowRunFailureRetrievalIncompatible {
		summary = "Workflow run comparison requires matching retrieval snapshot, profile, query, and node bindings."
	}
	return WorkflowRunComparisonResult{FailureCode: code, FailureSummary: summary}
}
