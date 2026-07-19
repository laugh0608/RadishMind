package httpapi

import (
	"encoding/json"
	"strings"
	"time"
)

const (
	workflowRAGApplicationExecutionKind         = "application_rag_invocation"
	workflowRAGApplicationExecutionSourceKind   = "application_configuration_draft"
	workflowRAGApplicationEffectiveSnapshotRole = "candidate"
)

type workflowRAGApplicationRunAuthority struct {
	AssignmentID          string                               `json:"assignment_id"`
	AssignmentVersion     int                                  `json:"assignment_version"`
	AssignmentDigest      string                               `json:"assignment_digest"`
	PublishCandidateID    string                               `json:"publish_candidate_id"`
	PublishReviewVersion  int                                  `json:"publish_review_version"`
	PublishCandidateState string                               `json:"publish_candidate_state"`
	DraftID               string                               `json:"draft_id"`
	DraftVersion          int                                  `json:"draft_version"`
	DraftDigest           string                               `json:"draft_digest"`
	BindingRef            WorkflowRAGApplicationBindingRef     `json:"binding_ref"`
	Dataset               WorkflowRAGQualityDatasetBinding     `json:"dataset"`
	CandidateReviewID     string                               `json:"candidate_review_id"`
	BaselineSnapshot      WorkflowRAGEvaluationSnapshotBinding `json:"baseline_snapshot"`
	CandidateSnapshot     WorkflowRAGEvaluationSnapshotBinding `json:"candidate_snapshot"`
	EffectiveSnapshotRole string                               `json:"effective_snapshot_role"`
	Profile               WorkflowRAGEvaluationProfileBinding  `json:"profile"`
	ConfiguredProtocol    string                               `json:"configured_protocol"`
	ConfiguredModel       string                               `json:"configured_model"`
}

type workflowRAGApplicationRunRecordV4 struct {
	SchemaVersion          string                             `json:"schema_version"`
	RecordVersion          int                                `json:"record_version"`
	RunID                  string                             `json:"run_id"`
	TenantRef              string                             `json:"tenant_ref"`
	WorkspaceID            string                             `json:"workspace_id"`
	ApplicationID          string                             `json:"application_id"`
	ExecutionKind          string                             `json:"execution_kind"`
	ExecutionSourceKind    string                             `json:"execution_source_kind"`
	ExecutionSourceID      string                             `json:"execution_source_id"`
	ExecutionSourceVersion int                                `json:"execution_source_version"`
	Status                 string                             `json:"status"`
	FailureCode            *string                            `json:"failure_code"`
	FailureSummary         string                             `json:"failure_summary"`
	StartedAt              string                             `json:"started_at"`
	CompletedAt            *string                            `json:"completed_at"`
	InputDigest            string                             `json:"input_digest"`
	InputBytes             int                                `json:"input_bytes"`
	Authority              workflowRAGApplicationRunAuthority `json:"authority"`
	Snapshot               workflowRAGRunSnapshotBinding      `json:"snapshot"`
	RetrievalAttempt       workflowRAGRunRetrievalAttempt     `json:"retrieval_attempt"`
	Answer                 *WorkflowRAGApplicationAnswer      `json:"answer"`
	SelectedProvider       string                             `json:"selected_provider"`
	SelectedModel          string                             `json:"selected_model"`
	RequestID              string                             `json:"request_id"`
	AuditRef               string                             `json:"audit_ref"`
	ActorRef               string                             `json:"actor_ref"`
	SideEffects            workflowRAGRunSideEffects          `json:"side_effects"`
	Diagnostic             workflowRAGRunDiagnostic           `json:"diagnostic"`
}

func marshalWorkflowRAGApplicationRunRecord(record WorkflowRunRecord) ([]byte, error) {
	if record.ExecutionSource == nil || record.RAGApplication == nil || record.RAGSnapshot == nil || record.RetrievalAttempt == nil || record.Diagnostic == nil {
		return nil, errWorkflowRunStoreContract
	}
	document := workflowRAGApplicationRunRecordV4{
		SchemaVersion: record.SchemaVersion, RecordVersion: record.RecordVersion, RunID: record.RunID,
		TenantRef: record.TenantRef, WorkspaceID: record.WorkspaceID, ApplicationID: record.ApplicationID,
		ExecutionKind: record.ExecutionSource.Kind, ExecutionSourceKind: record.ExecutionSource.SourceKind,
		ExecutionSourceID: record.ExecutionSource.ID, ExecutionSourceVersion: record.ExecutionSource.Version,
		Status: string(record.Status), FailureCode: workflowRAGFailureCodePointer(record.FailureCode), FailureSummary: record.FailureSummary,
		StartedAt: record.StartedAt, CompletedAt: workflowRAGTimestampPointer(record.CompletedAt),
		InputDigest: record.RetrievalAttempt.QueryDigest, InputBytes: record.InputBytes,
		Authority: *record.RAGApplication, Snapshot: *record.RAGSnapshot, RetrievalAttempt: *record.RetrievalAttempt,
		Answer: nil, SelectedProvider: record.SelectedProvider, SelectedModel: record.SelectedModel,
		RequestID: record.RequestID, AuditRef: record.AuditRef, ActorRef: record.ActorRef,
		SideEffects: workflowRAGRunSideEffects{RetrievalCalls: record.SideEffects.RetrievalCalls, ProviderCalls: record.SideEffects.ProviderCalls, ToolCalls: record.SideEffects.ToolCalls, ConfirmationCalls: record.SideEffects.ConfirmationCalls, BusinessWrites: record.SideEffects.BusinessWrites, ReplayWrites: record.SideEffects.ReplayWrites},
		Diagnostic:  workflowRAGRunDiagnostic{FailureBoundary: workflowRAGDiagnosticValue(string(record.Diagnostic.FailureBoundary)), RetrievalFailureCategory: workflowRAGDiagnosticValue(record.Diagnostic.RetrievalFailureCategory)},
	}
	return json.Marshal(document)
}

func validateWorkflowRAGApplicationRunStoreRecord(runContext WorkflowRunContext, record *WorkflowRunRecord) error {
	if record == nil || record.SchemaVersion != workflowRunRecordAppRAGSchemaVersion || record.TenantRef != strings.TrimSpace(runContext.TenantRef) || record.WorkspaceID != strings.TrimSpace(runContext.WorkspaceID) || record.ApplicationID != strings.TrimSpace(runContext.ApplicationID) || record.ExecutionSource == nil || record.ExecutionSource.Kind != workflowRAGApplicationExecutionKind || record.ExecutionSource.SourceKind != workflowRAGApplicationExecutionSourceKind || !applicationDraftIdentifierPattern.MatchString(record.ExecutionSource.ID) || record.ExecutionSource.Version < 1 || record.DraftID != "" || record.DraftVersion != 0 || record.DraftDigest != "" || record.RAGApplication == nil || record.RAGSnapshot == nil || record.RetrievalAttempt == nil || record.Diagnostic == nil || record.RAGAnswer != nil || record.Output != "" || record.ToolAttempt != nil || record.PlanID != "" || record.ConfirmationID != "" || len(record.ConditionNodeIDs) != 0 || len(record.Nodes) != 0 || record.InputBytes < 1 || record.InputBytes > workflowRAGApplicationInvocationMaxBytes {
		return errWorkflowRunStoreContract
	}
	authority := *record.RAGApplication
	if !workflowRAGApplicationRunAuthorityValid(authority, *record.RAGSnapshot, *record.RetrievalAttempt, record.ExecutionSource.ID, record.ExecutionSource.Version) {
		return errWorkflowRunStoreContract
	}
	version := record.RecordVersion
	if version == 0 {
		version = 1
	}
	document := workflowRAGApplicationRunRecordV4{SchemaVersion: record.SchemaVersion, RecordVersion: version, RunID: record.RunID, TenantRef: record.TenantRef, WorkspaceID: record.WorkspaceID, ApplicationID: record.ApplicationID, ExecutionKind: record.ExecutionSource.Kind, ExecutionSourceKind: record.ExecutionSource.SourceKind, ExecutionSourceID: record.ExecutionSource.ID, ExecutionSourceVersion: record.ExecutionSource.Version, Status: string(record.Status), FailureCode: workflowRAGFailureCodePointer(record.FailureCode), FailureSummary: record.FailureSummary, StartedAt: record.StartedAt, CompletedAt: workflowRAGTimestampPointer(record.CompletedAt), InputDigest: record.RetrievalAttempt.QueryDigest, InputBytes: record.InputBytes, Authority: authority, Snapshot: *record.RAGSnapshot, RetrievalAttempt: *record.RetrievalAttempt, Answer: nil, SelectedProvider: record.SelectedProvider, SelectedModel: record.SelectedModel, RequestID: record.RequestID, AuditRef: record.AuditRef, ActorRef: record.ActorRef, SideEffects: workflowRAGRunSideEffects{RetrievalCalls: record.SideEffects.RetrievalCalls, ProviderCalls: record.SideEffects.ProviderCalls, ToolCalls: record.SideEffects.ToolCalls, ConfirmationCalls: record.SideEffects.ConfirmationCalls, BusinessWrites: record.SideEffects.BusinessWrites, ReplayWrites: record.SideEffects.ReplayWrites}, Diagnostic: workflowRAGRunDiagnostic{FailureBoundary: workflowRAGDiagnosticValue(string(record.Diagnostic.FailureBoundary)), RetrievalFailureCategory: workflowRAGDiagnosticValue(record.Diagnostic.RetrievalFailureCategory)}}
	return validateWorkflowRAGApplicationRunRecordV4(document)
}

func validateWorkflowRAGApplicationRunRecordV4(record workflowRAGApplicationRunRecordV4) error {
	startedAt, err := time.Parse(time.RFC3339Nano, record.StartedAt)
	if err != nil || record.SchemaVersion != workflowRunRecordAppRAGSchemaVersion || record.RecordVersion < 1 || !workflowRAGRunIDPattern.MatchString(record.RunID) ||
		!workflowRAGReferencePattern.MatchString(record.TenantRef) || !workflowRAGScopedIDPattern.MatchString(record.WorkspaceID) || !workflowRAGScopedIDPattern.MatchString(record.ApplicationID) ||
		record.ExecutionKind != workflowRAGApplicationExecutionKind || record.ExecutionSourceKind != workflowRAGApplicationExecutionSourceKind || !applicationDraftIdentifierPattern.MatchString(record.ExecutionSourceID) || record.ExecutionSourceVersion < 1 ||
		!workflowRAGDigestPattern.MatchString(record.InputDigest) || record.InputBytes < 1 || record.InputBytes > workflowRAGApplicationInvocationMaxBytes || record.Answer != nil || !workflowRAGRunStatusAllowed(record.Status) ||
		!workflowRAGApplicationRunAuthorityValid(record.Authority, record.Snapshot, record.RetrievalAttempt, record.ExecutionSourceID, record.ExecutionSourceVersion) ||
		!workflowRAGRetrievalAttemptStatusAllowed(record.RetrievalAttempt.Status) || !workflowRAGScopedIDPattern.MatchString(record.RetrievalAttempt.NodeID) ||
		record.RetrievalAttempt.QueryDigest != record.InputDigest || record.RetrievalAttempt.QueryBytes != record.InputBytes || record.RetrievalAttempt.QueryBytes > workflowRAGMaxQueryBytes ||
		record.RetrievalAttempt.CandidateCount < 0 || record.RetrievalAttempt.CandidateCount > workflowRAGMaxFragments || len(record.RetrievalAttempt.SelectedFragments) > workflowRAGMaxTopK || len(record.RetrievalAttempt.SelectedFragments) > record.RetrievalAttempt.CandidateCount ||
		record.RetrievalAttempt.RetrievalLatencyMS < 0 || record.RetrievalAttempt.RetrievalLatencyMS > 2000 || record.RetrievalAttempt.ContextBytes < 0 || record.RetrievalAttempt.ContextBytes > workflowRAGMaxContextBytes ||
		record.SideEffects.RetrievalCalls < 0 || record.SideEffects.RetrievalCalls > 1 || record.SideEffects.ProviderCalls < 0 || record.SideEffects.ProviderCalls > 1 || record.SideEffects.ToolCalls != 0 || record.SideEffects.ConfirmationCalls != 0 || record.SideEffects.BusinessWrites != 0 || record.SideEffects.ReplayWrites != 0 ||
		!workflowRAGSafeContractText(record.SelectedProvider, 256) || !workflowRAGSafeContractText(record.SelectedModel, 256) || !workflowRAGReferencePattern.MatchString(record.RequestID) || !workflowRAGReferencePattern.MatchString(record.AuditRef) || !workflowRAGReferencePattern.MatchString(record.ActorRef) || len([]rune(record.FailureSummary)) > 256 ||
		!workflowRAGFailureBoundaryAllowed(record.Diagnostic.FailureBoundary) || !workflowRAGFailureCategoryAllowed(record.Diagnostic.RetrievalFailureCategory) {
		return errWorkflowRunStoreContract
	}
	if record.Status == string(WorkflowRunStatusRunning) {
		if record.CompletedAt != nil || record.FailureCode != nil {
			return errWorkflowRunStoreContract
		}
	} else {
		if record.CompletedAt == nil || record.FailureCode == nil && record.Status != string(WorkflowRunStatusSucceeded) {
			return errWorkflowRunStoreContract
		}
		completedAt, parseErr := time.Parse(time.RFC3339Nano, *record.CompletedAt)
		if parseErr != nil || completedAt.Before(startedAt) {
			return errWorkflowRunStoreContract
		}
	}
	if record.FailureCode != nil && !workflowRAGRunFailurePattern.MatchString(*record.FailureCode) {
		return errWorkflowRunStoreContract
	}
	if record.Status == string(WorkflowRunStatusSucceeded) && (record.FailureCode != nil || record.RetrievalAttempt.Status != "succeeded" || record.SideEffects.RetrievalCalls != 1 || record.SideEffects.ProviderCalls != 1 || len(record.RetrievalAttempt.CitationRefs) == 0) {
		return errWorkflowRunStoreContract
	}
	if record.RetrievalAttempt.Status == "not_started" && record.SideEffects.RetrievalCalls != 0 || record.RetrievalAttempt.Status != "not_started" && record.SideEffects.RetrievalCalls != 1 {
		return errWorkflowRunStoreContract
	}
	if record.SideEffects.ProviderCalls == 1 && record.RetrievalAttempt.Status != "succeeded" {
		return errWorkflowRunStoreContract
	}
	selected := make(map[string]bool, len(record.RetrievalAttempt.SelectedFragments))
	for index, fragment := range record.RetrievalAttempt.SelectedFragments {
		if !workflowRAGFragmentRefPattern.MatchString(fragment.FragmentRef) || !workflowRAGDigestPattern.MatchString(fragment.ContentDigest) || fragment.Rank != index+1 || !workflowRAGSourceTypeAllowed(fragment.SourceType) || selected[fragment.FragmentRef] {
			return errWorkflowRunStoreContract
		}
		selected[fragment.FragmentRef] = true
	}
	citations := make(map[string]bool, len(record.RetrievalAttempt.CitationRefs))
	for _, citation := range record.RetrievalAttempt.CitationRefs {
		if !selected[citation] || citations[citation] {
			return errWorkflowRunStoreContract
		}
		citations[citation] = true
	}
	return nil
}

func workflowRAGApplicationRunAuthorityValid(authority workflowRAGApplicationRunAuthority, snapshot workflowRAGRunSnapshotBinding, retrieval workflowRAGRunRetrievalAttempt, sourceID string, sourceVersion int) bool {
	return workflowRAGApplicationAssignmentIDPattern.MatchString(authority.AssignmentID) && authority.AssignmentVersion >= 1 && workflowRAGDigestPattern.MatchString(authority.AssignmentDigest) &&
		applicationDraftIdentifierPattern.MatchString(authority.PublishCandidateID) && authority.PublishReviewVersion >= 1 && authority.PublishCandidateState == applicationPublishStateApproved &&
		authority.DraftID == sourceID && authority.DraftVersion == sourceVersion && workflowRAGDigestPattern.MatchString(authority.DraftDigest) && validWorkflowRAGApplicationBindingRef(authority.BindingRef) &&
		workflowRAGDatasetIDPattern.MatchString(authority.Dataset.DatasetID) && authority.Dataset.DatasetVersion >= 1 && workflowRAGDigestPattern.MatchString(authority.Dataset.DatasetDigest) && workflowRAGEvaluationReviewIDPattern.MatchString(authority.CandidateReviewID) &&
		workflowRAGApplicationSnapshotBindingValid(authority.BaselineSnapshot) && workflowRAGApplicationSnapshotBindingValid(authority.CandidateSnapshot) && authority.EffectiveSnapshotRole == workflowRAGApplicationEffectiveSnapshotRole &&
		authority.CandidateSnapshot.SnapshotID == snapshot.SnapshotID && authority.CandidateSnapshot.SnapshotVersion == snapshot.SnapshotVersion && authority.CandidateSnapshot.SnapshotDigest == snapshot.SnapshotDigest && authority.CandidateSnapshot.RAGRef == snapshot.RAGRef &&
		authority.Profile.ProfileID == workflowRAGProfileID && authority.Profile.ProfileVersion == workflowRAGProfileVersion && workflowRAGDigestPattern.MatchString(authority.Profile.ProfileDigest) &&
		authority.Profile.ProfileID == retrieval.ProfileID && authority.Profile.ProfileVersion == retrieval.ProfileVersion && authority.Profile.ProfileDigest == retrieval.ProfileDigest &&
		workflowRAGSafeContractText(authority.ConfiguredProtocol, 256) && workflowRAGSafeContractText(authority.ConfiguredModel, 256)
}

func workflowRAGApplicationSnapshotBindingValid(binding WorkflowRAGEvaluationSnapshotBinding) bool {
	return workflowRAGSnapshotIDPattern.MatchString(binding.SnapshotID) && binding.SnapshotVersion >= 1 && workflowRAGDigestPattern.MatchString(binding.SnapshotDigest) && workflowRAGRAGRefPattern.MatchString(binding.RAGRef)
}
