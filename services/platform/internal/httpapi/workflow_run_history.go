package httpapi

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	workflowRunListDefaultLimit = 25
	workflowRunListMaxLimit     = 100
	workflowRunListMaxRange     = 90 * 24 * time.Hour
	workflowRunListFutureSkew   = 5 * time.Minute
)

type WorkflowRunListRequest struct {
	Limit           int
	Cursor          string
	Status          WorkflowRunStatus
	DraftID         string
	StartedFrom     *time.Time
	StartedTo       *time.Time
	FailureCode     WorkflowRunFailureCode
	FailureBoundary WorkflowRunFailureBoundary
	Provider        string
	Model           string
	StaleRunning    *bool
}

type WorkflowRunSummary struct {
	SchemaVersion            string                            `json:"schema_version"`
	RecordVersion            int                               `json:"record_version"`
	RunID                    string                            `json:"run_id"`
	PlanID                   string                            `json:"plan_id"`
	ConfirmationID           string                            `json:"confirmation_id"`
	ToolAttemptStatus        WorkflowHTTPToolAttemptStatus     `json:"tool_attempt_status"`
	DraftID                  string                            `json:"draft_id"`
	DraftVersion             int                               `json:"draft_version"`
	DraftDigest              string                            `json:"draft_digest,omitempty"`
	ExecutionKind            string                            `json:"execution_kind,omitempty"`
	ExecutionSourceKind      string                            `json:"execution_source_kind,omitempty"`
	ExecutionSourceID        string                            `json:"execution_source_id,omitempty"`
	ExecutionSourceVersion   int                               `json:"execution_source_version,omitempty"`
	RuntimeAssignmentID      string                            `json:"runtime_assignment_id,omitempty"`
	RuntimeAssignmentVersion int                               `json:"runtime_assignment_version,omitempty"`
	PublishCandidateID       string                            `json:"publish_candidate_id,omitempty"`
	PublishReviewVersion     int                               `json:"publish_review_version,omitempty"`
	EffectiveSnapshotRole    string                            `json:"effective_snapshot_role,omitempty"`
	WorkspaceID              string                            `json:"workspace_id"`
	ApplicationID            string                            `json:"application_id"`
	Status                   WorkflowRunStatus                 `json:"status"`
	FailureCode              WorkflowRunFailureCode            `json:"failure_code"`
	StartedAt                string                            `json:"started_at"`
	CompletedAt              string                            `json:"completed_at"`
	DurationMS               int64                             `json:"duration_ms"`
	SelectedProvider         string                            `json:"selected_provider"`
	SelectedProfile          string                            `json:"selected_profile"`
	SelectedModel            string                            `json:"selected_model"`
	RequestID                string                            `json:"request_id"`
	AuditRef                 string                            `json:"audit_ref"`
	SideEffects              WorkflowRunSideEffects            `json:"side_effects"`
	StaleRunning             bool                              `json:"stale_running"`
	FailureBoundary          WorkflowRunFailureBoundary        `json:"failure_boundary"`
	FailedNodeID             string                            `json:"failed_node_id"`
	LastCompletedNodeID      string                            `json:"last_completed_node_id"`
	GatewayFailureCategory   WorkflowRunGatewayFailureCategory `json:"gateway_failure_category"`
	ToolFailureCategory      WorkflowHTTPToolFailureCategory   `json:"tool_failure_category"`
	SnapshotID               string                            `json:"snapshot_id,omitempty"`
	SnapshotVersion          int                               `json:"snapshot_version,omitempty"`
	SnapshotDigest           string                            `json:"snapshot_digest,omitempty"`
	RAGRef                   string                            `json:"rag_ref,omitempty"`
	RetrievalNodeID          string                            `json:"retrieval_node_id,omitempty"`
	RetrievalAttemptStatus   string                            `json:"retrieval_attempt_status,omitempty"`
	RetrievalProfileID       string                            `json:"retrieval_profile_id,omitempty"`
	RetrievalProfileVersion  int                               `json:"retrieval_profile_version,omitempty"`
	RetrievalProfileDigest   string                            `json:"retrieval_profile_digest,omitempty"`
	QueryDigest              string                            `json:"query_digest,omitempty"`
	QueryBytes               int                               `json:"query_bytes,omitempty"`
	CandidateCount           int                               `json:"candidate_count,omitempty"`
	SelectedFragments        []workflowRAGRunSelectedFragment  `json:"selected_fragments,omitempty"`
	CitationRefs             []string                          `json:"citation_refs,omitempty"`
	RetrievalLatencyMS       int                               `json:"retrieval_latency_ms,omitempty"`
	RetrievalContextBytes    int                               `json:"retrieval_context_bytes,omitempty"`
	RetrievalFailureCategory string                            `json:"retrieval_failure_category,omitempty"`
	RecommendedReviewAction  WorkflowRunReviewAction           `json:"recommended_review_action"`
	retrievalAttemptPresent  bool
}

func (summary WorkflowRunSummary) MarshalJSON() ([]byte, error) {
	type Alias WorkflowRunSummary
	if !summary.retrievalAttemptPresent {
		return json.Marshal(Alias(summary))
	}
	return json.Marshal(struct {
		Alias
		CandidateCount        int                              `json:"candidate_count"`
		SelectedFragments     []workflowRAGRunSelectedFragment `json:"selected_fragments"`
		CitationRefs          []string                         `json:"citation_refs"`
		RetrievalLatencyMS    int                              `json:"retrieval_latency_ms"`
		RetrievalContextBytes int                              `json:"retrieval_context_bytes"`
	}{
		Alias:                 Alias(summary),
		CandidateCount:        summary.CandidateCount,
		SelectedFragments:     append([]workflowRAGRunSelectedFragment{}, summary.SelectedFragments...),
		CitationRefs:          append([]string{}, summary.CitationRefs...),
		RetrievalLatencyMS:    summary.RetrievalLatencyMS,
		RetrievalContextBytes: summary.RetrievalContextBytes,
	})
}

type WorkflowRunListResult struct {
	Runs           []WorkflowRunSummary
	NextCursor     string
	HasMore        bool
	FailureCode    WorkflowRunFailureCode
	FailureSummary string
}

type workflowRunCursor struct {
	Version      int    `json:"v"`
	StartedAt    string `json:"started_at"`
	RunID        string `json:"run_id"`
	FilterDigest string `json:"filter_digest"`
}

func (service workflowExecutorService) ListRuns(
	runContext WorkflowRunContext,
	request WorkflowRunListRequest,
) WorkflowRunListResult {
	filter, failureCode := normalizeWorkflowRunListRequest(request)
	if failureCode != "" {
		return workflowRunListFailure(failureCode)
	}
	if request.Cursor != "" {
		cursor, err := decodeWorkflowRunCursor(request.Cursor, filter)
		if err != nil {
			return workflowRunListFailure(WorkflowRunFailureCursorInvalid)
		}
		filter.BeforeTime = &cursor.StartedAt
		filter.BeforeRunID = cursor.RunID
	}
	page, err := service.store.ListRuns(runContext, filter)
	if err != nil {
		return workflowRunListFailure(workflowRunStoreFailureCode(err))
	}
	summaries := make([]WorkflowRunSummary, 0, len(page.Records))
	for _, record := range page.Records {
		summaries = append(summaries, summarizeWorkflowRun(record, time.Now().UTC()))
	}
	nextCursor := ""
	if page.HasMore && len(page.Records) > 0 {
		last := page.Records[len(page.Records)-1]
		nextCursor, err = encodeWorkflowRunCursor(last, filter)
		if err != nil {
			return workflowRunListFailure(WorkflowRunFailureStoreContractMismatch)
		}
	}
	return WorkflowRunListResult{Runs: summaries, NextCursor: nextCursor, HasMore: page.HasMore}
}

func normalizeWorkflowRunListRequest(request WorkflowRunListRequest) (WorkflowRunListFilter, WorkflowRunFailureCode) {
	limit := request.Limit
	if limit == 0 {
		limit = workflowRunListDefaultLimit
	}
	if limit < 1 || limit > workflowRunListMaxLimit || (request.Status != "" && !validWorkflowRunStatus(request.Status)) {
		return WorkflowRunListFilter{}, WorkflowRunFailureFilterInvalid
	}
	draftID := strings.TrimSpace(request.DraftID)
	if len([]rune(draftID)) > 160 {
		return WorkflowRunListFilter{}, WorkflowRunFailureFilterInvalid
	}
	provider, model := strings.TrimSpace(request.Provider), strings.TrimSpace(request.Model)
	if (request.FailureCode != "" && !validWorkflowRunFailureCode(request.FailureCode)) ||
		(request.FailureBoundary != "" && !validWorkflowRunFailureBoundary(request.FailureBoundary)) ||
		len([]rune(provider)) > workflowExecutorMaxModelChars || len([]rune(model)) > workflowExecutorMaxModelChars ||
		strings.Contains(provider, "://") || strings.Contains(model, "://") {
		return WorkflowRunListFilter{}, WorkflowRunFailureFilterInvalid
	}
	if request.StartedFrom != nil && request.StartedTo != nil {
		if request.StartedFrom.After(*request.StartedTo) || request.StartedTo.Sub(*request.StartedFrom) > workflowRunListMaxRange {
			return WorkflowRunListFilter{}, WorkflowRunFailureFilterInvalid
		}
	}
	nowLimit := time.Now().UTC().Add(workflowRunListFutureSkew)
	if (request.StartedFrom != nil && request.StartedFrom.After(nowLimit)) ||
		(request.StartedTo != nil && request.StartedTo.After(nowLimit)) {
		return WorkflowRunListFilter{}, WorkflowRunFailureFilterInvalid
	}
	return WorkflowRunListFilter{
		Status: request.Status, DraftID: draftID, StartedFrom: request.StartedFrom,
		StartedTo: request.StartedTo, Limit: limit, FailureCode: request.FailureCode,
		FailureBoundary: request.FailureBoundary, Provider: provider,
		Model: model, StaleRunning: request.StaleRunning,
	}, ""
}

func parseWorkflowRunListRequest(values map[string][]string) (WorkflowRunListRequest, WorkflowRunFailureCode) {
	allowed := map[string]bool{"workspace_id": true, "application_id": true, "limit": true, "cursor": true, "status": true, "draft_id": true, "started_from": true, "started_to": true, "failure_code": true, "failure_boundary": true, "provider": true, "model": true, "stale_running": true}
	for key, entries := range values {
		if !allowed[key] || len(entries) != 1 {
			return WorkflowRunListRequest{}, WorkflowRunFailureFilterInvalid
		}
	}
	request := WorkflowRunListRequest{Cursor: strings.TrimSpace(firstQueryValue(values, "cursor")), Status: WorkflowRunStatus(strings.TrimSpace(firstQueryValue(values, "status"))), DraftID: strings.TrimSpace(firstQueryValue(values, "draft_id")), FailureCode: WorkflowRunFailureCode(strings.TrimSpace(firstQueryValue(values, "failure_code"))), FailureBoundary: WorkflowRunFailureBoundary(strings.TrimSpace(firstQueryValue(values, "failure_boundary"))), Provider: strings.TrimSpace(firstQueryValue(values, "provider")), Model: strings.TrimSpace(firstQueryValue(values, "model"))}
	if raw := strings.TrimSpace(firstQueryValue(values, "stale_running")); raw != "" {
		if raw != "true" && raw != "false" {
			return WorkflowRunListRequest{}, WorkflowRunFailureFilterInvalid
		}
		value := raw == "true"
		request.StaleRunning = &value
	}
	if raw := strings.TrimSpace(firstQueryValue(values, "limit")); raw != "" {
		limit, err := strconv.Atoi(raw)
		if err != nil {
			return WorkflowRunListRequest{}, WorkflowRunFailureFilterInvalid
		}
		request.Limit = limit
	}
	for key, target := range map[string]**time.Time{"started_from": &request.StartedFrom, "started_to": &request.StartedTo} {
		raw := strings.TrimSpace(firstQueryValue(values, key))
		if raw == "" {
			continue
		}
		parsed, err := time.Parse(time.RFC3339Nano, raw)
		if err != nil {
			return WorkflowRunListRequest{}, WorkflowRunFailureFilterInvalid
		}
		parsed = parsed.UTC()
		*target = &parsed
	}
	return request, ""
}

func firstQueryValue(values map[string][]string, key string) string {
	if entries := values[key]; len(entries) == 1 {
		return entries[0]
	}
	return ""
}

func encodeWorkflowRunCursor(record WorkflowRunRecord, filter WorkflowRunListFilter) (string, error) {
	if _, err := time.Parse(time.RFC3339Nano, record.StartedAt); err != nil {
		return "", err
	}
	document, err := json.Marshal(workflowRunCursor{Version: 2, StartedAt: record.StartedAt, RunID: record.RunID, FilterDigest: workflowRunFilterDigest(filter)})
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(document), nil
}

type decodedWorkflowRunCursor struct {
	StartedAt time.Time
	RunID     string
}

func decodeWorkflowRunCursor(value string, filter WorkflowRunListFilter) (decodedWorkflowRunCursor, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(value))
	if err != nil || len(decoded) > 1024 {
		return decodedWorkflowRunCursor{}, errorsForWorkflowRunCursor()
	}
	var cursor workflowRunCursor
	decoder := json.NewDecoder(strings.NewReader(string(decoded)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&cursor); err != nil || cursor.Version != 2 || strings.TrimSpace(cursor.RunID) == "" || cursor.FilterDigest != workflowRunFilterDigest(filter) {
		return decodedWorkflowRunCursor{}, errorsForWorkflowRunCursor()
	}
	startedAt, err := time.Parse(time.RFC3339Nano, cursor.StartedAt)
	if err != nil {
		return decodedWorkflowRunCursor{}, errorsForWorkflowRunCursor()
	}
	return decodedWorkflowRunCursor{StartedAt: startedAt, RunID: cursor.RunID}, nil
}

func workflowRunFilterDigest(filter WorkflowRunListFilter) string {
	value := fmt.Sprintf("v2\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%d", filter.Status, filter.DraftID, filter.FailureCode, filter.FailureBoundary, filter.Provider, filter.Model, workflowRunFilterBool(filter.StaleRunning), workflowRunFilterTime(filter.StartedFrom)+"\n"+workflowRunFilterTime(filter.StartedTo), filter.Limit)
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:16])
}

func workflowRunFilterBool(value *bool) string {
	if value == nil {
		return ""
	}
	return strconv.FormatBool(*value)
}

func workflowRunFilterTime(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.UTC().Format(time.RFC3339Nano)
}

func errorsForWorkflowRunCursor() error { return fmt.Errorf("invalid workflow run cursor") }

func summarizeWorkflowRun(record WorkflowRunRecord, now time.Time) WorkflowRunSummary {
	startedAt, _ := time.Parse(time.RFC3339Nano, record.StartedAt)
	completedAt, _ := time.Parse(time.RFC3339Nano, record.CompletedAt)
	duration := int64(0)
	if !completedAt.IsZero() {
		duration = completedAt.Sub(startedAt).Milliseconds()
	}
	summary := WorkflowRunSummary{
		SchemaVersion: record.SchemaVersion, RecordVersion: record.RecordVersion, RunID: record.RunID,
		PlanID: record.PlanID, ConfirmationID: record.ConfirmationID,
		DraftID: record.DraftID, DraftVersion: record.DraftVersion, DraftDigest: record.DraftDigest, WorkspaceID: record.WorkspaceID,
		ApplicationID: record.ApplicationID, Status: record.Status, FailureCode: record.FailureCode,
		StartedAt: record.StartedAt, CompletedAt: record.CompletedAt, DurationMS: duration,
		SelectedProvider: record.SelectedProvider, SelectedProfile: record.SelectedProfile,
		SelectedModel: record.SelectedModel, RequestID: record.RequestID, AuditRef: record.AuditRef,
		SideEffects:  record.SideEffects,
		StaleRunning: record.Status == WorkflowRunStatusRunning && now.Sub(startedAt) > workflowExecutorDefaultMaxRuntime,
	}
	if record.ToolAttempt != nil {
		summary.ToolAttemptStatus = record.ToolAttempt.Status
	}
	if record.ExecutionSource != nil {
		summary.ExecutionKind = record.ExecutionSource.Kind
		summary.ExecutionSourceKind = record.ExecutionSource.SourceKind
		summary.ExecutionSourceID = record.ExecutionSource.ID
		summary.ExecutionSourceVersion = record.ExecutionSource.Version
	}
	if record.RAGApplication != nil {
		summary.RuntimeAssignmentID = record.RAGApplication.AssignmentID
		summary.RuntimeAssignmentVersion = record.RAGApplication.AssignmentVersion
		summary.PublishCandidateID = record.RAGApplication.PublishCandidateID
		summary.PublishReviewVersion = record.RAGApplication.PublishReviewVersion
		summary.EffectiveSnapshotRole = record.RAGApplication.EffectiveSnapshotRole
	}
	if record.Diagnostic != nil {
		summary.FailureBoundary = record.Diagnostic.FailureBoundary
		summary.FailedNodeID = record.Diagnostic.FailedNodeID
		summary.LastCompletedNodeID = record.Diagnostic.LastCompletedNodeID
		summary.GatewayFailureCategory = record.Diagnostic.GatewayFailureCategory
		summary.ToolFailureCategory = record.Diagnostic.ToolFailureCategory
		summary.RetrievalFailureCategory = record.Diagnostic.RetrievalFailureCategory
		summary.RecommendedReviewAction = record.Diagnostic.RecommendedReviewAction
	}
	if record.RAGSnapshot != nil {
		summary.SnapshotID, summary.SnapshotVersion = record.RAGSnapshot.SnapshotID, record.RAGSnapshot.SnapshotVersion
		summary.SnapshotDigest, summary.RAGRef = record.RAGSnapshot.SnapshotDigest, record.RAGSnapshot.RAGRef
	}
	if record.RetrievalAttempt != nil {
		summary.RetrievalNodeID, summary.RetrievalAttemptStatus = record.RetrievalAttempt.NodeID, record.RetrievalAttempt.Status
		summary.RetrievalProfileID, summary.RetrievalProfileVersion = record.RetrievalAttempt.ProfileID, record.RetrievalAttempt.ProfileVersion
		summary.RetrievalProfileDigest = record.RetrievalAttempt.ProfileDigest
		summary.QueryDigest, summary.QueryBytes = record.RetrievalAttempt.QueryDigest, record.RetrievalAttempt.QueryBytes
		summary.CandidateCount = record.RetrievalAttempt.CandidateCount
		summary.SelectedFragments = append([]workflowRAGRunSelectedFragment(nil), record.RetrievalAttempt.SelectedFragments...)
		summary.CitationRefs = cloneStringSlice(record.RetrievalAttempt.CitationRefs)
		summary.RetrievalLatencyMS, summary.RetrievalContextBytes = record.RetrievalAttempt.RetrievalLatencyMS, record.RetrievalAttempt.ContextBytes
		summary.retrievalAttemptPresent = true
	}
	return summary
}

func workflowRunListFailure(code WorkflowRunFailureCode) WorkflowRunListResult {
	summary := "Workflow run history request is invalid."
	if code == WorkflowRunFailureStoreUnavailable || code == WorkflowRunFailureStoreContractMismatch {
		summary = "Workflow run history storage is unavailable."
	} else if code == WorkflowRunFailureCursorInvalid {
		summary = "Workflow run history cursor is invalid."
	}
	return WorkflowRunListResult{FailureCode: code, FailureSummary: summary}
}

func workflowRunStoreFailureCode(err error) WorkflowRunFailureCode {
	if errors.Is(err, errWorkflowRunStoreContract) || errors.Is(err, errWorkflowRunStoreConflict) {
		return WorkflowRunFailureStoreContractMismatch
	}
	return WorkflowRunFailureStoreUnavailable
}
