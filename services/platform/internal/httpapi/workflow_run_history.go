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
	SchemaVersion           string                            `json:"schema_version"`
	RecordVersion           int                               `json:"record_version"`
	RunID                   string                            `json:"run_id"`
	DraftID                 string                            `json:"draft_id"`
	DraftVersion            int                               `json:"draft_version"`
	WorkspaceID             string                            `json:"workspace_id"`
	ApplicationID           string                            `json:"application_id"`
	Status                  WorkflowRunStatus                 `json:"status"`
	FailureCode             WorkflowRunFailureCode            `json:"failure_code"`
	StartedAt               string                            `json:"started_at"`
	CompletedAt             string                            `json:"completed_at"`
	DurationMS              int64                             `json:"duration_ms"`
	SelectedProvider        string                            `json:"selected_provider"`
	SelectedProfile         string                            `json:"selected_profile"`
	SelectedModel           string                            `json:"selected_model"`
	RequestID               string                            `json:"request_id"`
	AuditRef                string                            `json:"audit_ref"`
	SideEffects             WorkflowRunSideEffects            `json:"side_effects"`
	StaleRunning            bool                              `json:"stale_running"`
	FailureBoundary         WorkflowRunFailureBoundary        `json:"failure_boundary"`
	FailedNodeID            string                            `json:"failed_node_id"`
	LastCompletedNodeID     string                            `json:"last_completed_node_id"`
	GatewayFailureCategory  WorkflowRunGatewayFailureCategory `json:"gateway_failure_category"`
	RecommendedReviewAction WorkflowRunReviewAction           `json:"recommended_review_action"`
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
		DraftID: record.DraftID, DraftVersion: record.DraftVersion, WorkspaceID: record.WorkspaceID,
		ApplicationID: record.ApplicationID, Status: record.Status, FailureCode: record.FailureCode,
		StartedAt: record.StartedAt, CompletedAt: record.CompletedAt, DurationMS: duration,
		SelectedProvider: record.SelectedProvider, SelectedProfile: record.SelectedProfile,
		SelectedModel: record.SelectedModel, RequestID: record.RequestID, AuditRef: record.AuditRef,
		SideEffects:  record.SideEffects,
		StaleRunning: record.Status == WorkflowRunStatusRunning && now.Sub(startedAt) > workflowExecutorDefaultMaxRuntime,
	}
	if record.Diagnostic != nil {
		summary.FailureBoundary = record.Diagnostic.FailureBoundary
		summary.FailedNodeID = record.Diagnostic.FailedNodeID
		summary.LastCompletedNodeID = record.Diagnostic.LastCompletedNodeID
		summary.GatewayFailureCategory = record.Diagnostic.GatewayFailureCategory
		summary.RecommendedReviewAction = record.Diagnostic.RecommendedReviewAction
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
