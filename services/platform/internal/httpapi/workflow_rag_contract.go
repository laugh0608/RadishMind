package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"
)

type WorkflowRAGCitation struct {
	FragmentRef  string `json:"fragment_ref"`
	ClaimSummary string `json:"claim_summary"`
}

type WorkflowRAGAnswer struct {
	SchemaVersion string                `json:"schema_version"`
	Answer        string                `json:"answer"`
	Citations     []WorkflowRAGCitation `json:"citations"`
	Limitations   []string              `json:"limitations"`
	Confidence    string                `json:"confidence"`
}

type workflowRAGRunSnapshotBinding struct {
	SnapshotID      string `json:"snapshot_id"`
	SnapshotVersion int    `json:"snapshot_version"`
	SnapshotDigest  string `json:"snapshot_digest"`
	RAGRef          string `json:"rag_ref"`
}

type workflowRAGRunSelectedFragment struct {
	FragmentRef      string `json:"fragment_ref"`
	ContentDigest    string `json:"content_digest"`
	Rank             int    `json:"rank"`
	SourceType       string `json:"source_type"`
	IsOfficial       bool   `json:"is_official"`
	ExcerptTruncated bool   `json:"excerpt_truncated"`
}

type workflowRAGRunRetrievalAttempt struct {
	NodeID             string                           `json:"node_id"`
	Status             string                           `json:"status"`
	ProfileID          string                           `json:"profile_id"`
	ProfileVersion     int                              `json:"profile_version"`
	ProfileDigest      string                           `json:"profile_digest"`
	QueryDigest        string                           `json:"query_digest"`
	QueryBytes         int                              `json:"query_bytes"`
	CandidateCount     int                              `json:"candidate_count"`
	SelectedFragments  []workflowRAGRunSelectedFragment `json:"selected_fragments"`
	RetrievalLatencyMS int                              `json:"retrieval_latency_ms"`
	ContextBytes       int                              `json:"context_bytes"`
	CitationRefs       []string                         `json:"citation_refs"`
}

type workflowRAGRunSideEffects struct {
	RetrievalCalls    int `json:"retrieval_calls"`
	ProviderCalls     int `json:"provider_calls"`
	ToolCalls         int `json:"tool_calls"`
	ConfirmationCalls int `json:"confirmation_calls"`
	BusinessWrites    int `json:"business_writes"`
	ReplayWrites      int `json:"replay_writes"`
}

type workflowRAGRunDiagnostic struct {
	FailureBoundary          string `json:"failure_boundary"`
	RetrievalFailureCategory string `json:"retrieval_failure_category"`
}

type workflowRAGRunRecordV3 struct {
	SchemaVersion    string                         `json:"schema_version"`
	RecordVersion    int                            `json:"record_version"`
	RunID            string                         `json:"run_id"`
	TenantRef        string                         `json:"tenant_ref"`
	WorkspaceID      string                         `json:"workspace_id"`
	ApplicationID    string                         `json:"application_id"`
	DraftID          string                         `json:"draft_id"`
	DraftVersion     int                            `json:"draft_version"`
	DraftDigest      string                         `json:"draft_digest"`
	Status           string                         `json:"status"`
	FailureCode      *string                        `json:"failure_code"`
	FailureSummary   string                         `json:"failure_summary"`
	StartedAt        string                         `json:"started_at"`
	CompletedAt      *string                        `json:"completed_at"`
	Snapshot         workflowRAGRunSnapshotBinding  `json:"snapshot"`
	RetrievalAttempt workflowRAGRunRetrievalAttempt `json:"retrieval_attempt"`
	Answer           *WorkflowRAGAnswer             `json:"answer"`
	SelectedProvider string                         `json:"selected_provider"`
	SelectedModel    string                         `json:"selected_model"`
	RequestID        string                         `json:"request_id"`
	AuditRef         string                         `json:"audit_ref"`
	ActorRef         string                         `json:"actor_ref"`
	SideEffects      workflowRAGRunSideEffects      `json:"side_effects"`
	Diagnostic       workflowRAGRunDiagnostic       `json:"diagnostic"`
}

func marshalWorkflowRAGRunRecord(record WorkflowRunRecord) ([]byte, error) {
	if record.RAGSnapshot == nil || record.RetrievalAttempt == nil || record.Diagnostic == nil {
		return nil, errWorkflowRunStoreContract
	}
	document := workflowRAGRunRecordV3{
		SchemaVersion: record.SchemaVersion, RecordVersion: record.RecordVersion, RunID: record.RunID,
		TenantRef: record.TenantRef, WorkspaceID: record.WorkspaceID, ApplicationID: record.ApplicationID,
		DraftID: record.DraftID, DraftVersion: record.DraftVersion, DraftDigest: record.DraftDigest, Status: string(record.Status),
		FailureCode: workflowRAGFailureCodePointer(record.FailureCode), FailureSummary: record.FailureSummary,
		StartedAt: record.StartedAt, CompletedAt: workflowRAGTimestampPointer(record.CompletedAt),
		Snapshot: *record.RAGSnapshot, RetrievalAttempt: *record.RetrievalAttempt, Answer: record.RAGAnswer,
		SelectedProvider: record.SelectedProvider, SelectedModel: record.SelectedModel,
		RequestID: record.RequestID, AuditRef: record.AuditRef, ActorRef: record.ActorRef,
		SideEffects: workflowRAGRunSideEffects{
			RetrievalCalls: record.SideEffects.RetrievalCalls, ProviderCalls: record.SideEffects.ProviderCalls,
			ToolCalls: record.SideEffects.ToolCalls, ConfirmationCalls: record.SideEffects.ConfirmationCalls,
			BusinessWrites: record.SideEffects.BusinessWrites, ReplayWrites: record.SideEffects.ReplayWrites,
		},
		Diagnostic: workflowRAGRunDiagnostic{
			FailureBoundary:          workflowRAGDiagnosticValue(string(record.Diagnostic.FailureBoundary)),
			RetrievalFailureCategory: workflowRAGDiagnosticValue(record.Diagnostic.RetrievalFailureCategory),
		},
	}
	return json.Marshal(document)
}

func workflowRAGFailureCodePointer(value WorkflowRunFailureCode) *string {
	if value == "" {
		return nil
	}
	normalized := string(value)
	return &normalized
}

func workflowRAGTimestampPointer(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return &value
}

func workflowRAGDiagnosticValue(value string) string {
	if strings.TrimSpace(value) == "" {
		return "none"
	}
	return value
}

func validateWorkflowRAGContractJSON(contract string, payload []byte) error {
	switch contract {
	case workflowRAGFragmentSchemaVersion:
		var fragment WorkflowRAGFragment
		if err := decodeWorkflowRAGStrictJSON(payload, &fragment); err != nil {
			return err
		}
		return validateStoredWorkflowRAGFragment(fragment, fragment.ContentClassification)
	case workflowRAGSnapshotSchemaVersion:
		var record WorkflowRAGSnapshotRecord
		if err := decodeWorkflowRAGStrictJSON(payload, &record); err != nil {
			return err
		}
		ctx := WorkflowRAGSnapshotContext{TenantRef: record.TenantRef, WorkspaceID: record.WorkspaceID, ApplicationID: record.ApplicationID}
		return validateStoredWorkflowRAGRecord(record, ctx)
	case "workflow_rag_execution_profile.v1":
		var profile WorkflowRAGExecutionProfile
		if err := decodeWorkflowRAGStrictJSON(payload, &profile); err != nil {
			return err
		}
		expected := workflowRAGLexicalProfile()
		if profile != expected {
			return errWorkflowRAGStoreContract
		}
		return nil
	case "workflow_rag_answer.v1":
		var answer WorkflowRAGAnswer
		if err := decodeWorkflowRAGStrictJSON(payload, &answer); err != nil {
			return err
		}
		return validateWorkflowRAGAnswer(answer, nil)
	case workflowRAGAuditSchemaVersion:
		var audit WorkflowRAGExecutionAudit
		if err := decodeWorkflowRAGStrictJSON(payload, &audit); err != nil {
			return err
		}
		return validateWorkflowRAGContractAudit(audit)
	case "workflow_run_record.v3":
		var record workflowRAGRunRecordV3
		if err := decodeWorkflowRAGStrictJSON(payload, &record); err != nil {
			return err
		}
		return validateWorkflowRAGRunRecordV3(record)
	case workflowRunRecordAppRAGSchemaVersion:
		var record workflowRAGApplicationRunRecordV4
		if err := decodeWorkflowRAGStrictJSON(payload, &record); err != nil {
			return err
		}
		return validateWorkflowRAGApplicationRunRecordV4(record)
	case workflowRAGApplicationRuntimeAssignmentSchemaVersion:
		var assignment WorkflowRAGApplicationRuntimeAssignment
		if err := decodeWorkflowRAGStrictJSON(payload, &assignment); err != nil {
			return err
		}
		ctx := WorkflowRAGApplicationRuntimeContext{RequestContext: context.Background(), RequestID: assignment.RequestID, TenantRef: assignment.TenantRef, WorkspaceID: assignment.WorkspaceID, ApplicationID: assignment.ApplicationID, ActorRef: assignment.UpdatedByActorRef, OwnerSubjectRef: assignment.OwnerSubjectRef, AuditRef: assignment.AuditRef}
		return validateStoredWorkflowRAGApplicationAssignment(assignment, ctx)
	case workflowRAGApplicationRuntimeEventSchemaVersion:
		var event WorkflowRAGApplicationRuntimeEvent
		if err := decodeWorkflowRAGStrictJSON(payload, &event); err != nil {
			return err
		}
		return validateStoredWorkflowRAGApplicationEvent(event)
	case workflowRAGApplicationRuntimeAuditSchemaVersion:
		var audit WorkflowRAGApplicationRuntimeAudit
		if err := decodeWorkflowRAGStrictJSON(payload, &audit); err != nil {
			return err
		}
		ctx := WorkflowRAGApplicationRuntimeContext{RequestContext: context.Background(), RequestID: audit.RequestID, TenantRef: audit.TenantRef, WorkspaceID: audit.WorkspaceID, ApplicationID: audit.ApplicationID, ActorRef: audit.ActorRef, OwnerSubjectRef: audit.OwnerSubjectRef, AuditRef: audit.AuditRef}
		return validateStoredWorkflowRAGApplicationAudit(audit, ctx)
	case workflowRAGApplicationAnswerSchemaVersion:
		var answer WorkflowRAGApplicationAnswer
		if err := decodeWorkflowRAGStrictJSON(payload, &answer); err != nil {
			return err
		}
		return validateWorkflowRAGApplicationAnswer(answer, nil)
	case workflowRAGEvaluationDatasetSchemaVersion:
		_, err := DecodeWorkflowRAGEvaluationDataset(payload)
		return err
	case workflowRAGEvaluationResourceSchemaVersion:
		var version WorkflowRAGEvaluationDatasetVersion
		if err := decodeWorkflowRAGStrictJSON(payload, &version); err != nil {
			return err
		}
		ctx := WorkflowRAGSnapshotContext{TenantRef: version.Dataset.Snapshot.TenantRef, WorkspaceID: version.Dataset.Snapshot.WorkspaceID, ApplicationID: version.Dataset.Snapshot.ApplicationID}
		return validateStoredWorkflowRAGEvaluationVersion(version, ctx)
	case workflowRAGCandidateReviewSchemaVersion:
		var review WorkflowRAGCandidateSnapshotReview
		if err := decodeWorkflowRAGStrictJSON(payload, &review); err != nil {
			return err
		}
		ctx := WorkflowRAGSnapshotContext{TenantRef: review.TenantRef, WorkspaceID: review.WorkspaceID, ApplicationID: review.ApplicationID, ActorRef: review.CreatedByActorRef, RequestID: review.RequestID, AuditRef: review.AuditRef}
		return validateStoredWorkflowRAGCandidateReview(review, ctx)
	case workflowRAGPromotionCandidateSchemaVersion:
		_, err := decodeWorkflowRAGPromotionCandidate(payload)
		return err
	case workflowRAGPromotionDecisionSchemaVersion:
		_, err := decodeWorkflowRAGPromotionDecision(payload)
		return err
	case workflowRAGApplicationBindingSchemaVersion:
		_, err := decodeWorkflowRAGApplicationBinding(payload)
		return err
	case workflowRAGPromotionAuditSchemaVersion:
		_, err := decodeWorkflowRAGPromotionAudit(payload)
		return err
	case workflowRAGQualityReviewSchemaVersion:
		_, err := DecodeWorkflowRAGQualityReview(payload)
		return err
	default:
		return errWorkflowRAGStoreContract
	}
}

func decodeWorkflowRAGStrictJSON(payload []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return errWorkflowRAGStoreContract
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errWorkflowRAGStoreContract
	}
	return nil
}

func validateWorkflowRAGAnswer(answer WorkflowRAGAnswer, selectedRefs map[string]bool) error {
	if answer.SchemaVersion != "workflow_rag_answer.v1" || strings.TrimSpace(answer.Answer) == "" ||
		len([]rune(answer.Answer)) > 16384 || len(answer.Citations) < 1 || len(answer.Citations) > workflowRAGMaxTopK ||
		len(answer.Limitations) > workflowRAGMaxTopK || (answer.Confidence != "low" && answer.Confidence != "medium" && answer.Confidence != "high") ||
		workflowRAGContainsForbiddenMaterial(answer.Answer) {
		return errWorkflowRAGStoreContract
	}
	seen := make(map[string]bool, len(answer.Citations))
	for _, citation := range answer.Citations {
		if !workflowRAGFragmentRefPattern.MatchString(citation.FragmentRef) || strings.TrimSpace(citation.ClaimSummary) == "" ||
			len([]rune(citation.ClaimSummary)) > 512 || seen[citation.FragmentRef] || (selectedRefs != nil && !selectedRefs[citation.FragmentRef]) ||
			workflowRAGContainsForbiddenMaterial(citation.ClaimSummary) {
			return errWorkflowRAGStoreContract
		}
		seen[citation.FragmentRef] = true
	}
	for _, limitation := range answer.Limitations {
		if strings.TrimSpace(limitation) == "" || len([]rune(limitation)) > 512 || workflowRAGContainsForbiddenMaterial(limitation) {
			return errWorkflowRAGStoreContract
		}
	}
	return nil
}

func validateWorkflowRAGContractAudit(audit WorkflowRAGExecutionAudit) error {
	if audit.SchemaVersion != workflowRAGAuditSchemaVersion || !workflowRAGAuditEventAllowed(audit.EventKind) ||
		!workflowRAGSnapshotIDPattern.MatchString(audit.SnapshotID) || !workflowRAGSnapshotKeyPattern.MatchString(audit.SnapshotKey) ||
		audit.SnapshotVersion < 1 || !workflowRAGDigestPattern.MatchString(audit.SnapshotDigest) || audit.FragmentCount < 0 ||
		audit.FragmentCount > workflowRAGMaxFragments || audit.TotalContentBytes < 0 || audit.TotalContentBytes > workflowRAGMaxSnapshotBytes ||
		!workflowRAGReferencePattern.MatchString(audit.TenantRef) || !workflowRAGScopedIDPattern.MatchString(audit.WorkspaceID) ||
		!workflowRAGScopedIDPattern.MatchString(audit.ApplicationID) || !workflowRAGReferencePattern.MatchString(audit.ActorRef) ||
		!workflowRAGReferencePattern.MatchString(audit.RequestID) || !workflowRAGReferencePattern.MatchString(audit.AuditRef) {
		return errWorkflowRAGStoreContract
	}
	if _, err := time.Parse(time.RFC3339Nano, audit.OccurredAt); err != nil {
		return errWorkflowRAGStoreContract
	}
	return nil
}

func workflowRAGAuditEventAllowed(value string) bool {
	switch value {
	case "snapshot_created", "snapshot_versioned", "snapshot_archived", "retrieval_started", "retrieval_succeeded", "retrieval_failed":
		return true
	default:
		return false
	}
}

func validateWorkflowRAGRunRecordV3(record workflowRAGRunRecordV3) error {
	if record.SchemaVersion != "workflow_run_record.v3" || record.RecordVersion < 1 || !workflowRAGRunIDPattern.MatchString(record.RunID) ||
		!workflowRAGReferencePattern.MatchString(record.TenantRef) || !workflowRAGScopedIDPattern.MatchString(record.WorkspaceID) ||
		!workflowRAGScopedIDPattern.MatchString(record.ApplicationID) || !workflowRAGScopedIDPattern.MatchString(record.DraftID) ||
		record.DraftVersion < 1 || !workflowRAGDigestPattern.MatchString(record.DraftDigest) || !workflowRAGRunStatusAllowed(record.Status) || !workflowRAGScopedIDPattern.MatchString(record.RetrievalAttempt.NodeID) ||
		!workflowRAGSnapshotIDPattern.MatchString(record.Snapshot.SnapshotID) || record.Snapshot.SnapshotVersion < 1 ||
		!workflowRAGDigestPattern.MatchString(record.Snapshot.SnapshotDigest) ||
		!workflowRAGRAGRefPattern.MatchString(record.Snapshot.RAGRef) ||
		record.RetrievalAttempt.ProfileID != workflowRAGProfileID || record.RetrievalAttempt.ProfileVersion != workflowRAGProfileVersion ||
		!workflowRAGRetrievalAttemptStatusAllowed(record.RetrievalAttempt.Status) ||
		!workflowRAGDigestPattern.MatchString(record.RetrievalAttempt.ProfileDigest) || !workflowRAGDigestPattern.MatchString(record.RetrievalAttempt.QueryDigest) ||
		record.RetrievalAttempt.QueryBytes < 0 || record.RetrievalAttempt.QueryBytes > workflowRAGMaxQueryBytes ||
		record.RetrievalAttempt.CandidateCount < 0 || record.RetrievalAttempt.CandidateCount > workflowRAGMaxFragments ||
		len(record.RetrievalAttempt.SelectedFragments) > workflowRAGMaxTopK || record.RetrievalAttempt.ContextBytes < 0 ||
		record.RetrievalAttempt.ContextBytes > workflowRAGMaxContextBytes || record.RetrievalAttempt.RetrievalLatencyMS < 0 ||
		record.RetrievalAttempt.RetrievalLatencyMS > 2000 || record.SideEffects.RetrievalCalls < 0 || record.SideEffects.RetrievalCalls > 1 ||
		record.SideEffects.ProviderCalls < 0 || record.SideEffects.ProviderCalls > 1 || record.SideEffects.ToolCalls != 0 ||
		record.SideEffects.ConfirmationCalls != 0 || record.SideEffects.BusinessWrites != 0 || record.SideEffects.ReplayWrites != 0 ||
		!workflowRAGSafeContractText(record.SelectedProvider, 256) || !workflowRAGSafeContractText(record.SelectedModel, 256) ||
		!workflowRAGReferencePattern.MatchString(record.RequestID) || !workflowRAGReferencePattern.MatchString(record.AuditRef) ||
		!workflowRAGReferencePattern.MatchString(record.ActorRef) || len([]rune(record.FailureSummary)) > 256 ||
		!workflowRAGFailureBoundaryAllowed(record.Diagnostic.FailureBoundary) ||
		!workflowRAGFailureCategoryAllowed(record.Diagnostic.RetrievalFailureCategory) ||
		len(record.RetrievalAttempt.SelectedFragments) > record.RetrievalAttempt.CandidateCount {
		return errWorkflowRAGStoreContract
	}
	startedAt, err := time.Parse(time.RFC3339Nano, record.StartedAt)
	if err != nil {
		return errWorkflowRAGStoreContract
	}
	if record.Status == "running" {
		if record.CompletedAt != nil || record.FailureCode != nil || record.Answer != nil {
			return errWorkflowRAGStoreContract
		}
	} else if record.CompletedAt == nil || record.FailureCode == nil && record.Status != "succeeded" {
		return errWorkflowRAGStoreContract
	}
	if record.FailureCode != nil && !workflowRAGRunFailurePattern.MatchString(*record.FailureCode) {
		return errWorkflowRAGStoreContract
	}
	if record.Status == "succeeded" && (record.FailureCode != nil || record.Answer != nil || len(record.RetrievalAttempt.CitationRefs) == 0 ||
		record.RetrievalAttempt.Status != "succeeded" || record.SideEffects.RetrievalCalls != 1 || record.SideEffects.ProviderCalls != 1) {
		return errWorkflowRAGStoreContract
	}
	if (record.Status == "failed" || record.Status == "canceled") && record.Answer != nil {
		return errWorkflowRAGStoreContract
	}
	if record.CompletedAt != nil {
		completedAt, parseErr := time.Parse(time.RFC3339Nano, *record.CompletedAt)
		if parseErr != nil || completedAt.Before(startedAt) {
			return errWorkflowRAGStoreContract
		}
	}
	if record.RetrievalAttempt.Status == "not_started" && record.SideEffects.RetrievalCalls != 0 ||
		record.RetrievalAttempt.Status != "not_started" && record.SideEffects.RetrievalCalls != 1 {
		return errWorkflowRAGStoreContract
	}
	selectedRefs := make(map[string]bool, len(record.RetrievalAttempt.SelectedFragments))
	for index, selected := range record.RetrievalAttempt.SelectedFragments {
		if !workflowRAGFragmentRefPattern.MatchString(selected.FragmentRef) || !workflowRAGDigestPattern.MatchString(selected.ContentDigest) ||
			selected.Rank != index+1 || !workflowRAGSourceTypeAllowed(selected.SourceType) || selectedRefs[selected.FragmentRef] {
			return errWorkflowRAGStoreContract
		}
		selectedRefs[selected.FragmentRef] = true
	}
	citationRefs := make(map[string]bool, len(record.RetrievalAttempt.CitationRefs))
	for _, citationRef := range record.RetrievalAttempt.CitationRefs {
		if !selectedRefs[citationRef] || citationRefs[citationRef] {
			return errWorkflowRAGStoreContract
		}
		citationRefs[citationRef] = true
	}
	if record.Answer != nil {
		return errWorkflowRAGStoreContract
	}
	return nil
}

func validateWorkflowRAGRunStoreRecord(runContext WorkflowRunContext, record *WorkflowRunRecord) error {
	if record == nil || record.SchemaVersion != workflowRunRecordRAGSchemaVersion || record.RecordVersion < 0 ||
		record.TenantRef != strings.TrimSpace(runContext.TenantRef) || record.WorkspaceID != strings.TrimSpace(runContext.WorkspaceID) ||
		record.ApplicationID != strings.TrimSpace(runContext.ApplicationID) || record.RAGSnapshot == nil ||
		record.RetrievalAttempt == nil || record.Diagnostic == nil || record.Output != "" || record.ToolAttempt != nil ||
		record.PlanID != "" || record.ConfirmationID != "" || len(record.ConditionNodeIDs) != 0 || len(record.Nodes) != 0 {
		return errWorkflowRunStoreContract
	}
	version := record.RecordVersion
	if version == 0 {
		version = 1
	}
	document := workflowRAGRunRecordV3{
		SchemaVersion: record.SchemaVersion, RecordVersion: version, RunID: record.RunID,
		TenantRef: record.TenantRef, WorkspaceID: record.WorkspaceID, ApplicationID: record.ApplicationID,
		DraftID: record.DraftID, DraftVersion: record.DraftVersion, DraftDigest: record.DraftDigest, Status: string(record.Status),
		FailureCode: workflowRAGFailureCodePointer(record.FailureCode), FailureSummary: record.FailureSummary,
		StartedAt: record.StartedAt, CompletedAt: workflowRAGTimestampPointer(record.CompletedAt),
		Snapshot: *record.RAGSnapshot, RetrievalAttempt: *record.RetrievalAttempt, Answer: record.RAGAnswer,
		SelectedProvider: record.SelectedProvider, SelectedModel: record.SelectedModel,
		RequestID: record.RequestID, AuditRef: record.AuditRef, ActorRef: record.ActorRef,
		SideEffects: workflowRAGRunSideEffects{
			RetrievalCalls: record.SideEffects.RetrievalCalls, ProviderCalls: record.SideEffects.ProviderCalls,
			ToolCalls: record.SideEffects.ToolCalls, ConfirmationCalls: record.SideEffects.ConfirmationCalls,
			BusinessWrites: record.SideEffects.BusinessWrites, ReplayWrites: record.SideEffects.ReplayWrites,
		},
		Diagnostic: workflowRAGRunDiagnostic{
			FailureBoundary:          workflowRAGDiagnosticValue(string(record.Diagnostic.FailureBoundary)),
			RetrievalFailureCategory: workflowRAGDiagnosticValue(record.Diagnostic.RetrievalFailureCategory),
		},
	}
	return validateWorkflowRAGRunRecordV3(document)
}

func workflowRAGRunStatusAllowed(value string) bool {
	return value == "running" || value == "succeeded" || value == "failed" || value == "canceled"
}

func workflowRAGRetrievalAttemptStatusAllowed(value string) bool {
	return value == "not_started" || value == "succeeded" || value == "failed"
}

func workflowRAGFailureBoundaryAllowed(value string) bool {
	switch value {
	case "none", "retrieval_policy", "retrieval_store", "retrieval_rank", "retrieval_context", "retrieval_citation", "provider_selection", "provider_call", "run_store":
		return true
	default:
		return false
	}
}

func workflowRAGFailureCategoryAllowed(value string) bool {
	switch value {
	case "none", "scope", "snapshot", "profile", "query", "budget", "no_evidence", "provider", "answer", "citation", "store":
		return true
	default:
		return false
	}
}

func workflowRAGSafeContractText(value string, maximumRunes int) bool {
	return strings.TrimSpace(value) != "" && len([]rune(value)) <= maximumRunes && !strings.Contains(value, "://") && !workflowRAGContainsForbiddenMaterial(value)
}
