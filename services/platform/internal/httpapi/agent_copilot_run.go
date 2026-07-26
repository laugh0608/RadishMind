package httpapi

import (
	"encoding/json"
	"strings"
)

func agentCopilotRunDocument(record WorkflowRunRecord) (AgentCopilotRunRecordV7, error) {
	if record.SchemaVersion != agentCopilotRunV7Schema || record.AgentCopilotAuthority == nil ||
		record.PromptDiagnostic == nil || record.ExecutionSource == nil {
		return AgentCopilotRunRecordV7{}, errWorkflowRunStoreContract
	}
	return AgentCopilotRunRecordV7{
		SchemaVersion: record.SchemaVersion, RecordVersion: record.RecordVersion, RunID: record.RunID,
		TenantRef: record.TenantRef, WorkspaceID: record.WorkspaceID, ApplicationID: record.ApplicationID,
		ExecutionKind: record.ExecutionSource.Kind, ExecutionSourceKind: record.ExecutionSource.SourceKind,
		ExecutionSourceID: record.ExecutionSource.ID, ExecutionSourceVersion: record.ExecutionSource.Version,
		ExecutionProfile: record.ExecutionProfile, Authority: *record.AgentCopilotAuthority,
		Project: record.AgentProject, Task: record.AgentTask, Locale: record.AgentLocale,
		InputDigest: record.InputDigest, InputBytes: record.InputBytes, ContextBytes: record.AgentContextBytes,
		ArtifactCount: record.AgentArtifactCount, ArtifactBytes: record.AgentArtifactBytes,
		RequestedProtocol: record.RequestedProtocol, SelectedProtocol: record.SelectedProtocol,
		RequestedModel: record.RequestedModel, SelectedProvider: record.SelectedProvider,
		SelectedProfile: record.SelectedProfile, SelectedModel: record.SelectedModel,
		UpstreamModel: record.UpstreamModel, SelectionSource: record.SelectionSource,
		ResponseStatus: record.AgentResponseStatus, ResponseDigest: record.AgentResponseDigest,
		AnswerCount: record.AgentAnswerCount, IssueCount: record.AgentIssueCount,
		ActionCount: record.AgentActionCount, CitationCount: record.AgentCitationCount,
		RiskLevel: record.AgentRiskLevel, RequiresConfirmation: record.AgentRequiresConfirmation,
		Status: string(record.Status), FailureCode: string(record.FailureCode), FailureSummary: record.FailureSummary,
		StartedAt: record.StartedAt, CompletedAt: record.CompletedAt, Output: record.Output,
		Usage: record.PromptUsage,
		SideEffects: AgentCopilotRunSideEffectsV7{
			RetrievalCalls: record.SideEffects.RetrievalCalls, ProviderCalls: record.SideEffects.ProviderCalls,
			ToolCalls: record.SideEffects.ToolCalls, ConfirmationCalls: record.SideEffects.ConfirmationCalls,
			BusinessWrites: record.SideEffects.BusinessWrites, ReplayWrites: record.SideEffects.ReplayWrites,
		},
		Diagnostic: *record.PromptDiagnostic, RequestID: record.RequestID, AuditRef: record.AuditRef, ActorRef: record.ActorRef,
	}, nil
}

func workflowRunRecordFromAgentCopilotDocument(document AgentCopilotRunRecordV7) WorkflowRunRecord {
	authority := document.Authority
	diagnostic := document.Diagnostic
	return WorkflowRunRecord{
		SchemaVersion: document.SchemaVersion, RecordVersion: document.RecordVersion, RunID: document.RunID,
		TenantRef: document.TenantRef, WorkspaceID: document.WorkspaceID, ApplicationID: document.ApplicationID,
		ExecutionSource: &workflowRunExecutionSource{
			Kind: document.ExecutionKind, SourceKind: document.ExecutionSourceKind,
			ID: document.ExecutionSourceID, Version: document.ExecutionSourceVersion,
		},
		ExecutionProfile: document.ExecutionProfile, AgentCopilotAuthority: &authority,
		AgentProject: document.Project, AgentTask: document.Task, AgentLocale: document.Locale,
		InputDigest: document.InputDigest, InputBytes: document.InputBytes, AgentContextBytes: document.ContextBytes,
		AgentArtifactCount: document.ArtifactCount, AgentArtifactBytes: document.ArtifactBytes,
		RequestedProtocol: document.RequestedProtocol, SelectedProtocol: document.SelectedProtocol,
		RequestedModel: document.RequestedModel, SelectedProvider: document.SelectedProvider,
		SelectedProfile: document.SelectedProfile, SelectedModel: document.SelectedModel,
		UpstreamModel: document.UpstreamModel, SelectionSource: document.SelectionSource,
		AgentResponseStatus: document.ResponseStatus, AgentResponseDigest: document.ResponseDigest,
		AgentAnswerCount: document.AnswerCount, AgentIssueCount: document.IssueCount,
		AgentActionCount: document.ActionCount, AgentCitationCount: document.CitationCount,
		AgentRiskLevel: document.RiskLevel, AgentRequiresConfirmation: document.RequiresConfirmation,
		Status: WorkflowRunStatus(document.Status), FailureCode: WorkflowRunFailureCode(document.FailureCode),
		FailureSummary: document.FailureSummary, StartedAt: document.StartedAt, CompletedAt: document.CompletedAt,
		Output: document.Output, PromptUsage: document.Usage,
		SideEffects: WorkflowRunSideEffects{
			RetrievalCalls: document.SideEffects.RetrievalCalls, ProviderCalls: document.SideEffects.ProviderCalls,
			ToolCalls: document.SideEffects.ToolCalls, ConfirmationCalls: document.SideEffects.ConfirmationCalls,
			BusinessWrites: document.SideEffects.BusinessWrites, ReplayWrites: document.SideEffects.ReplayWrites,
		},
		PromptDiagnostic: &diagnostic, RequestID: document.RequestID, AuditRef: document.AuditRef, ActorRef: document.ActorRef,
	}
}

func validateAgentCopilotWorkflowRunRecord(runContext WorkflowRunContext, record *WorkflowRunRecord) error {
	document, err := agentCopilotRunDocument(*record)
	if err != nil || record.RecordVersion < 0 || record.TenantRef != strings.TrimSpace(runContext.TenantRef) ||
		record.WorkspaceID != strings.TrimSpace(runContext.WorkspaceID) || record.ApplicationID != strings.TrimSpace(runContext.ApplicationID) {
		return errWorkflowRunStoreContract
	}
	if document.RecordVersion == 0 {
		document.RecordVersion = 1
	}
	if validateAgentCopilotRun(document) != nil {
		return errWorkflowRunStoreContract
	}
	return nil
}

func encodeAgentCopilotRunStorageRecord(record WorkflowRunRecord) ([]byte, error) {
	document, err := agentCopilotRunDocument(record)
	if err != nil || validateAgentCopilotRun(document) != nil {
		return nil, errWorkflowRunStoreContract
	}
	payload, err := json.Marshal(document)
	if err != nil {
		return nil, errWorkflowRunStoreContract
	}
	return payload, nil
}

func decodeAgentCopilotRunStorageRecord(runContext WorkflowRunContext, payload []byte) (WorkflowRunRecord, error) {
	value, err := decodeAgentCopilotContract(agentCopilotRunV7Schema, payload)
	if err != nil {
		return WorkflowRunRecord{}, errWorkflowRunStoreContract
	}
	document, ok := value.(*AgentCopilotRunRecordV7)
	if !ok {
		return WorkflowRunRecord{}, errWorkflowRunStoreContract
	}
	record := workflowRunRecordFromAgentCopilotDocument(*document)
	if validateWorkflowRunStoreRecord(runContext, &record) != nil {
		return WorkflowRunRecord{}, errWorkflowRunStoreContract
	}
	return record, nil
}
