package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"
)

func encodeWorkflowRunStorageRecord(record WorkflowRunRecord) ([]byte, time.Time, *time.Time, error) {
	payload, err := json.Marshal(record)
	if err != nil {
		return nil, time.Time{}, nil, errWorkflowRunStoreContract
	}
	startedAt, err := time.Parse(time.RFC3339Nano, record.StartedAt)
	if err != nil {
		return nil, time.Time{}, nil, errWorkflowRunStoreContract
	}
	startedAt = startedAt.UTC()
	var completedAt *time.Time
	if record.CompletedAt != "" {
		parsed, parseErr := time.Parse(time.RFC3339Nano, record.CompletedAt)
		if parseErr != nil {
			return nil, time.Time{}, nil, errWorkflowRunStoreContract
		}
		parsed = parsed.UTC()
		if parsed.Before(startedAt) {
			return nil, time.Time{}, nil, errWorkflowRunStoreContract
		}
		completedAt = &parsed
	}
	return payload, startedAt, completedAt, nil
}

func decodeWorkflowRunStorageRecord(
	runContext WorkflowRunContext,
	payload []byte,
) (WorkflowRunRecord, error) {
	var envelope struct {
		SchemaVersion string `json:"schema_version"`
	}
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return WorkflowRunRecord{}, errWorkflowRunStoreContract
	}
	if envelope.SchemaVersion == workflowRunRecordAppRAGSchemaVersion {
		return decodeWorkflowRAGApplicationRunStorageRecord(runContext, payload)
	}
	var record WorkflowRunRecord
	decoder := json.NewDecoder(strings.NewReader(string(payload)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&record); err != nil {
		return WorkflowRunRecord{}, errWorkflowRunStoreContract
	}
	if err := rejectTrailingWorkflowRunJSON(decoder); err != nil {
		return WorkflowRunRecord{}, errWorkflowRunStoreContract
	}
	if record.SchemaVersion == workflowRunRecordDefinitionSchemaVersion {
		record.ExecutionSource = &workflowRunExecutionSource{Kind: record.ExecutionKind, SourceKind: record.ExecutionSourceKind, ID: record.ExecutionSourceID, Version: record.ExecutionSourceVersion}
	}
	if err := validateWorkflowRunStoreRecord(runContext, &record); err != nil || record.RecordVersion <= 0 {
		return WorkflowRunRecord{}, errWorkflowRunStoreContract
	}
	return record, nil
}

func decodeWorkflowRAGApplicationRunStorageRecord(runContext WorkflowRunContext, payload []byte) (WorkflowRunRecord, error) {
	var document workflowRAGApplicationRunRecordV4
	decoder := json.NewDecoder(strings.NewReader(string(payload)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&document); err != nil || rejectTrailingWorkflowRunJSON(decoder) != nil || validateWorkflowRAGApplicationRunRecordV4(document) != nil {
		return WorkflowRunRecord{}, errWorkflowRunStoreContract
	}
	record := WorkflowRunRecord{
		SchemaVersion: document.SchemaVersion, RecordVersion: document.RecordVersion, RunID: document.RunID,
		TenantRef: document.TenantRef, WorkspaceID: document.WorkspaceID, ApplicationID: document.ApplicationID,
		ExecutionSource: &workflowRunExecutionSource{Kind: document.ExecutionKind, SourceKind: document.ExecutionSourceKind, ID: document.ExecutionSourceID, Version: document.ExecutionSourceVersion},
		Status:          WorkflowRunStatus(document.Status), FailureCode: workflowRunFailureCodeFromPointer(document.FailureCode), FailureSummary: document.FailureSummary,
		StartedAt: document.StartedAt, CompletedAt: workflowRunStringFromPointer(document.CompletedAt), InputBytes: document.InputBytes,
		SelectedProvider: document.SelectedProvider, SelectedModel: document.SelectedModel,
		RequestID: document.RequestID, AuditRef: document.AuditRef, ActorRef: document.ActorRef,
		RAGSnapshot: &document.Snapshot, RetrievalAttempt: &document.RetrievalAttempt, RAGApplication: &document.Authority,
		SideEffects: WorkflowRunSideEffects{RetrievalCalls: document.SideEffects.RetrievalCalls, ProviderCalls: document.SideEffects.ProviderCalls, ToolCalls: document.SideEffects.ToolCalls, ConfirmationCalls: document.SideEffects.ConfirmationCalls, BusinessWrites: document.SideEffects.BusinessWrites, ReplayWrites: document.SideEffects.ReplayWrites},
		Diagnostic:  &WorkflowRunDiagnostic{FailureBoundary: WorkflowRunFailureBoundary(document.Diagnostic.FailureBoundary), RetrievalFailureCategory: document.Diagnostic.RetrievalFailureCategory},
	}
	if validateWorkflowRunStoreRecord(runContext, &record) != nil {
		return WorkflowRunRecord{}, errWorkflowRunStoreContract
	}
	return record, nil
}

func workflowRunFailureCodeFromPointer(value *string) WorkflowRunFailureCode {
	if value == nil {
		return ""
	}
	return WorkflowRunFailureCode(*value)
}

func workflowRunStringFromPointer(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func workflowRunStorageExecutionSource(record WorkflowRunRecord) (string, string, int, error) {
	if record.SchemaVersion == workflowRunRecordAppRAGSchemaVersion {
		if record.ExecutionSource == nil || record.ExecutionSource.SourceKind != workflowRAGApplicationExecutionSourceKind || strings.TrimSpace(record.ExecutionSource.ID) == "" || record.ExecutionSource.Version < 1 {
			return "", "", 0, errWorkflowRunStoreContract
		}
		return record.ExecutionSource.SourceKind, record.ExecutionSource.ID, record.ExecutionSource.Version, nil
	}
	if record.SchemaVersion == workflowRunRecordDefinitionSchemaVersion {
		if record.ExecutionSource == nil || record.ExecutionSource.SourceKind != workflowDefinitionExecutionSourceKind || strings.TrimSpace(record.ExecutionSource.ID) == "" || record.ExecutionSource.Version < 1 {
			return "", "", 0, errWorkflowRunStoreContract
		}
		return record.ExecutionSource.SourceKind, record.ExecutionSource.ID, record.ExecutionSource.Version, nil
	}
	if !validWorkflowRunRecordSchema(record.SchemaVersion) || strings.TrimSpace(record.DraftID) == "" || record.DraftVersion < 1 {
		return "", "", 0, errWorkflowRunStoreContract
	}
	return "workflow_draft", record.DraftID, record.DraftVersion, nil
}

func rejectTrailingWorkflowRunJSON(decoder *json.Decoder) error {
	var trailing any
	err := decoder.Decode(&trailing)
	if errors.Is(err, io.EOF) {
		return nil
	}
	return errWorkflowRunStoreContract
}

func workflowRunUnixNano(value time.Time) (int64, error) {
	value = value.UTC()
	unixNano := value.UnixNano()
	if !time.Unix(0, unixNano).UTC().Equal(value) {
		return 0, errWorkflowRunStoreContract
	}
	return unixNano, nil
}

func optionalWorkflowRunUnixNano(value *time.Time) (any, error) {
	if value == nil {
		return nil, nil
	}
	return workflowRunUnixNano(value.UTC())
}
