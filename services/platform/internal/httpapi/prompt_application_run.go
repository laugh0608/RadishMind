package httpapi

import (
	"encoding/json"
	"sort"
	"strings"
)

const (
	promptApplicationExecutionKind       = "prompt_application_invocation"
	promptApplicationExecutionSourceKind = "prompt_application_template"
)

func promptApplicationRunDocument(record WorkflowRunRecord) (PromptApplicationRunRecordV6, error) {
	if record.SchemaVersion != workflowRunRecordPromptSchemaVersion || record.PromptApplication == nil || record.PromptDiagnostic == nil ||
		record.ExecutionSource == nil {
		return PromptApplicationRunRecordV6{}, errWorkflowRunStoreContract
	}
	return PromptApplicationRunRecordV6{
		SchemaVersion: record.SchemaVersion, RecordVersion: record.RecordVersion, RunID: record.RunID,
		TenantRef: record.TenantRef, WorkspaceID: record.WorkspaceID, ApplicationID: record.ApplicationID,
		ExecutionKind: record.ExecutionSource.Kind, ExecutionSourceKind: record.ExecutionSource.SourceKind,
		ExecutionSourceID: record.ExecutionSource.ID, ExecutionSourceVersion: record.ExecutionSource.Version,
		ExecutionProfile: record.ExecutionProfile, Authority: *record.PromptApplication,
		InputDigest: record.InputDigest, InputBytes: record.InputBytes,
		VariableNames: append([]string(nil), record.VariableNames...), VariableNamesDigest: record.VariableNamesDigest,
		RequestedProtocol: record.RequestedProtocol, SelectedProtocol: record.SelectedProtocol,
		RequestedModel: record.RequestedModel, SelectedProvider: record.SelectedProvider,
		SelectedProfile: record.SelectedProfile, SelectedModel: record.SelectedModel,
		UpstreamModel: record.UpstreamModel, SelectionSource: record.SelectionSource,
		Status: string(record.Status), FailureCode: string(record.FailureCode), FailureSummary: record.FailureSummary,
		StartedAt: record.StartedAt, CompletedAt: record.CompletedAt, Output: record.Output,
		Usage: record.PromptUsage, SideEffects: record.SideEffects, Diagnostic: *record.PromptDiagnostic,
		ScheduleExecution: cloneApplicationEvaluationScheduleExecutionRef(record.ScheduleExecution),
		RequestID:         record.RequestID, AuditRef: record.AuditRef, ActorRef: record.ActorRef,
	}, nil
}

func workflowRunRecordFromPromptApplicationDocument(document PromptApplicationRunRecordV6) WorkflowRunRecord {
	authority := document.Authority
	diagnostic := document.Diagnostic
	return WorkflowRunRecord{
		SchemaVersion: document.SchemaVersion, RecordVersion: document.RecordVersion, RunID: document.RunID,
		TenantRef: document.TenantRef, WorkspaceID: document.WorkspaceID, ApplicationID: document.ApplicationID,
		ExecutionSource: &workflowRunExecutionSource{
			Kind: document.ExecutionKind, SourceKind: document.ExecutionSourceKind,
			ID: document.ExecutionSourceID, Version: document.ExecutionSourceVersion,
		},
		ExecutionKind: document.ExecutionKind, ExecutionSourceKind: document.ExecutionSourceKind,
		ExecutionSourceID: document.ExecutionSourceID, ExecutionSourceVersion: document.ExecutionSourceVersion,
		ExecutionProfile: document.ExecutionProfile, PromptApplication: &authority,
		InputDigest: document.InputDigest, InputBytes: document.InputBytes,
		VariableNames: append([]string(nil), document.VariableNames...), VariableNamesDigest: document.VariableNamesDigest,
		RequestedProtocol: document.RequestedProtocol, SelectedProtocol: document.SelectedProtocol,
		RequestedModel: document.RequestedModel, SelectedProvider: document.SelectedProvider,
		SelectedProfile: document.SelectedProfile, SelectedModel: document.SelectedModel,
		UpstreamModel: document.UpstreamModel, SelectionSource: document.SelectionSource,
		Status: WorkflowRunStatus(document.Status), FailureCode: WorkflowRunFailureCode(document.FailureCode),
		FailureSummary: document.FailureSummary, StartedAt: document.StartedAt, CompletedAt: document.CompletedAt,
		Output: document.Output, PromptUsage: document.Usage, SideEffects: document.SideEffects,
		ScheduleExecution: cloneApplicationEvaluationScheduleExecutionRef(document.ScheduleExecution),
		PromptDiagnostic:  &diagnostic, RequestID: document.RequestID, AuditRef: document.AuditRef, ActorRef: document.ActorRef,
	}
}

func validatePromptApplicationWorkflowRunRecord(runContext WorkflowRunContext, record *WorkflowRunRecord) error {
	document, err := promptApplicationRunDocument(*record)
	if err != nil || record.RecordVersion < 0 || record.TenantRef != strings.TrimSpace(runContext.TenantRef) ||
		record.WorkspaceID != strings.TrimSpace(runContext.WorkspaceID) || record.ApplicationID != strings.TrimSpace(runContext.ApplicationID) {
		return errWorkflowRunStoreContract
	}
	if document.RecordVersion == 0 {
		document.RecordVersion = 1
	}
	if validatePromptApplicationRunRecordV6(document) != nil {
		return errWorkflowRunStoreContract
	}
	return nil
}

func encodePromptApplicationRunStorageRecord(record WorkflowRunRecord) ([]byte, error) {
	document, err := promptApplicationRunDocument(record)
	if err != nil || validatePromptApplicationRunRecordV6(document) != nil {
		return nil, errWorkflowRunStoreContract
	}
	payload, err := json.Marshal(document)
	if err != nil {
		return nil, errWorkflowRunStoreContract
	}
	return payload, nil
}

func decodePromptApplicationRunStorageRecord(runContext WorkflowRunContext, payload []byte) (WorkflowRunRecord, error) {
	value, err := decodePromptApplicationVNextContract(promptApplicationRunV6Schema, payload)
	if err != nil {
		return WorkflowRunRecord{}, errWorkflowRunStoreContract
	}
	document, ok := value.(*PromptApplicationRunRecordV6)
	if !ok {
		return WorkflowRunRecord{}, errWorkflowRunStoreContract
	}
	record := workflowRunRecordFromPromptApplicationDocument(*document)
	if validateWorkflowRunStoreRecord(runContext, &record) != nil {
		return WorkflowRunRecord{}, errWorkflowRunStoreContract
	}
	return record, nil
}

func calculatePromptApplicationVariableNamesDigest(names []string) (string, error) {
	normalized := append([]string(nil), names...)
	sort.Strings(normalized)
	payload, err := json.Marshal(normalized)
	if err != nil {
		return "", err
	}
	return workflowRAGSHA256(string(payload)), nil
}
