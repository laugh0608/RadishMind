package httpapi

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"regexp"
	"sort"
	"strings"
)

const (
	promptApplicationConfigurationDraftV3Schema   = "application_configuration_draft.v3"
	promptApplicationPublishCandidateV3Schema     = "application_publish_candidate.v3"
	promptApplicationRuntimeAssignmentSchema      = "prompt_application_runtime_assignment.v1"
	promptApplicationRuntimeAssignmentEventSchema = "prompt_application_runtime_assignment_event.v1"
	promptApplicationRuntimeAuthorityV2Schema     = "application_runtime_authority.v2"
	promptApplicationSessionV2Schema              = "application_session.v2"
	promptApplicationSessionTurnV2Schema          = "application_session_turn.v2"
	promptApplicationRunV6Schema                  = "workflow_run_record.v6"
	promptApplicationInvocationProfile            = "prompt_application_invocation_v1"
)

var (
	promptApplicationAssignmentIDPattern = regexp.MustCompile(`^ptra_[a-z2-7]{16}$`)
	promptApplicationEventIDPattern      = regexp.MustCompile(`^ptrae_[a-z2-7]{16}$`)
)

var errPromptApplicationVNextContract = errors.New("prompt application vNext contract mismatch")

type PromptApplicationTemplateRef struct {
	TemplateID      string `json:"template_id"`
	TemplateVersion int    `json:"template_version"`
	TemplateDigest  string `json:"template_digest"`
}

type PromptApplicationConfigurationDraftV3 struct {
	SchemaVersion            string                                  `json:"schema_version"`
	DraftID                  string                                  `json:"draft_id"`
	WorkspaceID              string                                  `json:"workspace_id"`
	ApplicationID            string                                  `json:"application_id"`
	BaseApplicationUpdatedAt string                                  `json:"base_application_updated_at"`
	DisplayName              string                                  `json:"display_name"`
	Description              string                                  `json:"description"`
	ApplicationKind          string                                  `json:"application_kind"`
	DefaultProtocol          string                                  `json:"default_protocol"`
	DefaultModel             string                                  `json:"default_model"`
	AllowedProtocols         []string                                `json:"allowed_protocols"`
	PromptTemplateRef        PromptApplicationTemplateRef            `json:"prompt_template_ref"`
	DraftVersion             int                                     `json:"draft_version"`
	DraftDigest              string                                  `json:"draft_digest"`
	ValidationSummary        ApplicationConfigurationDraftValidation `json:"validation_summary"`
	CreatedAt                string                                  `json:"created_at"`
	UpdatedAt                string                                  `json:"updated_at"`
	CreatedByActorRef        string                                  `json:"created_by_actor_ref"`
	UpdatedByActorRef        string                                  `json:"updated_by_actor_ref"`
	RequestID                string                                  `json:"request_id"`
	AuditRef                 string                                  `json:"audit_ref"`
}

type PromptApplicationPublishConfigurationV3 struct {
	DisplayName       string                       `json:"display_name"`
	Description       string                       `json:"description"`
	ApplicationKind   string                       `json:"application_kind"`
	DefaultProtocol   string                       `json:"default_protocol"`
	DefaultModel      string                       `json:"default_model"`
	AllowedProtocols  []string                     `json:"allowed_protocols"`
	PromptTemplateRef PromptApplicationTemplateRef `json:"prompt_template_ref"`
}

type PromptApplicationPublishCandidateV3 struct {
	SchemaVersion            string                                  `json:"schema_version"`
	CandidateID              string                                  `json:"candidate_id"`
	WorkspaceID              string                                  `json:"workspace_id"`
	ApplicationID            string                                  `json:"application_id"`
	DraftID                  string                                  `json:"draft_id"`
	DraftVersion             int                                     `json:"draft_version"`
	DraftDigest              string                                  `json:"draft_digest"`
	BaseApplicationUpdatedAt string                                  `json:"base_application_updated_at"`
	Configuration            PromptApplicationPublishConfigurationV3 `json:"configuration"`
	EvidenceRequestIDs       []string                                `json:"evidence_request_ids"`
	CandidateState           string                                  `json:"candidate_state"`
	ReviewVersion            int                                     `json:"review_version"`
	Reviews                  []ApplicationPublishReviewRecord        `json:"reviews"`
	PromotionEligibility     ApplicationPromotionEligibility         `json:"promotion_eligibility"`
	CreatedAt                string                                  `json:"created_at"`
	UpdatedAt                string                                  `json:"updated_at"`
	CreatedByActorRef        string                                  `json:"created_by_actor_ref"`
	UpdatedByActorRef        string                                  `json:"updated_by_actor_ref"`
	RequestID                string                                  `json:"request_id"`
	AuditRef                 string                                  `json:"audit_ref"`
}

type PromptApplicationRuntimeAssignment struct {
	SchemaVersion          string                       `json:"schema_version"`
	AssignmentID           string                       `json:"assignment_id"`
	TenantRef              string                       `json:"tenant_ref"`
	WorkspaceID            string                       `json:"workspace_id"`
	ApplicationID          string                       `json:"application_id"`
	OwnerSubjectRef        string                       `json:"owner_subject_ref"`
	AssignmentVersion      int                          `json:"assignment_version"`
	State                  string                       `json:"state"`
	CandidateID            string                       `json:"candidate_id"`
	CandidateReviewVersion int                          `json:"candidate_review_version"`
	DraftID                string                       `json:"draft_id"`
	DraftVersion           int                          `json:"draft_version"`
	DraftDigest            string                       `json:"draft_digest"`
	PromptTemplateRef      PromptApplicationTemplateRef `json:"prompt_template_ref"`
	AssignmentDigest       string                       `json:"assignment_digest"`
	ActivatedAt            string                       `json:"activated_at"`
	UpdatedAt              string                       `json:"updated_at"`
	RevokedAt              *string                      `json:"revoked_at"`
	ActivatedByActorRef    string                       `json:"activated_by_actor_ref"`
	UpdatedByActorRef      string                       `json:"updated_by_actor_ref"`
	RequestID              string                       `json:"request_id"`
	AuditRef               string                       `json:"audit_ref"`
}

type PromptApplicationRuntimeAssignmentEvent struct {
	SchemaVersion              string                       `json:"schema_version"`
	EventID                    string                       `json:"event_id"`
	AssignmentID               string                       `json:"assignment_id"`
	TenantRef                  string                       `json:"tenant_ref"`
	WorkspaceID                string                       `json:"workspace_id"`
	ApplicationID              string                       `json:"application_id"`
	OwnerSubjectRef            string                       `json:"owner_subject_ref"`
	EventSequence              int                          `json:"event_sequence"`
	Action                     string                       `json:"action"`
	ExpectedAssignmentVersion  int                          `json:"expected_assignment_version"`
	ResultingAssignmentVersion int                          `json:"resulting_assignment_version"`
	CandidateID                string                       `json:"candidate_id"`
	CandidateReviewVersion     int                          `json:"candidate_review_version"`
	DraftID                    string                       `json:"draft_id"`
	DraftVersion               int                          `json:"draft_version"`
	DraftDigest                string                       `json:"draft_digest"`
	PromptTemplateRef          PromptApplicationTemplateRef `json:"prompt_template_ref"`
	AssignmentDigest           string                       `json:"assignment_digest"`
	OccurredAt                 string                       `json:"occurred_at"`
	ActorRef                   string                       `json:"actor_ref"`
	RequestID                  string                       `json:"request_id"`
	AuditRef                   string                       `json:"audit_ref"`
}

type PromptApplicationAuthorityV2 struct {
	AssignmentID           string                       `json:"assignment_id"`
	AssignmentVersion      int                          `json:"assignment_version"`
	AssignmentDigest       string                       `json:"assignment_digest"`
	PublishCandidateID     string                       `json:"publish_candidate_id"`
	PublishReviewVersion   int                          `json:"publish_review_version"`
	DraftID                string                       `json:"draft_id"`
	DraftVersion           int                          `json:"draft_version"`
	DraftDigest            string                       `json:"draft_digest"`
	PromptTemplateRef      PromptApplicationTemplateRef `json:"prompt_template_ref"`
	DefaultProtocol        string                       `json:"default_protocol"`
	DefaultModel           string                       `json:"default_model"`
	ProtocolPolicyDigest   string                       `json:"protocol_policy_digest"`
	ModelEligibilityDigest string                       `json:"model_eligibility_digest"`
}

type PromptApplicationRuntimeAuthorityV2 struct {
	SchemaVersion            string                       `json:"schema_version"`
	ExecutionProfile         string                       `json:"execution_profile"`
	ApplicationID            string                       `json:"application_id"`
	ApplicationRecordVersion int                          `json:"application_record_version"`
	ApplicationLifecycle     string                       `json:"application_lifecycle"`
	PromptApplication        PromptApplicationAuthorityV2 `json:"prompt_application"`
	AuthorityDigest          string                       `json:"authority_digest"`
}

type PromptApplicationSessionProfileBindingV2 struct {
	ExecutionProfile string `json:"execution_profile"`
}

type PromptApplicationSessionV2 struct {
	SchemaVersion     string                                   `json:"schema_version"`
	SessionID         string                                   `json:"session_id"`
	TenantRef         string                                   `json:"tenant_ref"`
	WorkspaceID       string                                   `json:"workspace_id"`
	ApplicationID     string                                   `json:"application_id"`
	OwnerSubjectRef   string                                   `json:"owner_subject_ref"`
	State             string                                   `json:"state"`
	RecordVersion     int                                      `json:"record_version"`
	ProfileBinding    PromptApplicationSessionProfileBindingV2 `json:"profile_binding"`
	Authority         PromptApplicationRuntimeAuthorityV2      `json:"authority"`
	ContentRetention  string                                   `json:"content_retention"`
	TurnCount         int                                      `json:"turn_count"`
	LastTurnID        *string                                  `json:"last_turn_id"`
	CreatedAt         string                                   `json:"created_at"`
	UpdatedAt         string                                   `json:"updated_at"`
	ClosedAt          *string                                  `json:"closed_at"`
	CreatedByActorRef string                                   `json:"created_by_actor_ref"`
	UpdatedByActorRef string                                   `json:"updated_by_actor_ref"`
	RequestID         string                                   `json:"request_id"`
	AuditRef          string                                   `json:"audit_ref"`
}

type PromptApplicationRunRefV6 struct {
	RunID         string `json:"run_id"`
	SchemaVersion string `json:"schema_version"`
}

type PromptApplicationSessionTurnV2 struct {
	SchemaVersion    string                              `json:"schema_version"`
	TurnID           string                              `json:"turn_id"`
	SessionID        string                              `json:"session_id"`
	Sequence         int                                 `json:"sequence"`
	ClientTurnKey    string                              `json:"client_turn_key"`
	TenantRef        string                              `json:"tenant_ref"`
	WorkspaceID      string                              `json:"workspace_id"`
	ApplicationID    string                              `json:"application_id"`
	OwnerSubjectRef  string                              `json:"owner_subject_ref"`
	ExecutionProfile string                              `json:"execution_profile"`
	Authority        PromptApplicationRuntimeAuthorityV2 `json:"authority"`
	Status           string                              `json:"status"`
	InputDigest      string                              `json:"input_digest"`
	InputBytes       int                                 `json:"input_bytes"`
	RunRef           *PromptApplicationRunRefV6          `json:"run_ref"`
	FailureCode      string                              `json:"failure_code"`
	FailureSummary   string                              `json:"failure_summary"`
	StartedAt        string                              `json:"started_at"`
	CompletedAt      *string                             `json:"completed_at"`
	ActorRef         string                              `json:"actor_ref"`
	RequestID        string                              `json:"request_id"`
	AuditRef         string                              `json:"audit_ref"`
}

type PromptApplicationRunUsageV6 struct {
	State        string `json:"state"`
	InputTokens  int    `json:"input_tokens"`
	OutputTokens int    `json:"output_tokens"`
	TotalTokens  int    `json:"total_tokens"`
}

type PromptApplicationRunDiagnosticV6 struct {
	FailureBoundary         string `json:"failure_boundary"`
	FailureStage            string `json:"failure_stage"`
	TerminalWriteState      string `json:"terminal_write_state"`
	GatewayFailureCategory  string `json:"gateway_failure_category"`
	Summary                 string `json:"summary"`
	RecommendedReviewAction string `json:"recommended_review_action"`
	ObservedAt              string `json:"observed_at"`
}

type PromptApplicationRunRecordV6 struct {
	SchemaVersion          string                              `json:"schema_version"`
	RecordVersion          int                                 `json:"record_version"`
	RunID                  string                              `json:"run_id"`
	TenantRef              string                              `json:"tenant_ref"`
	WorkspaceID            string                              `json:"workspace_id"`
	ApplicationID          string                              `json:"application_id"`
	ExecutionKind          string                              `json:"execution_kind"`
	ExecutionSourceKind    string                              `json:"execution_source_kind"`
	ExecutionSourceID      string                              `json:"execution_source_id"`
	ExecutionSourceVersion int                                 `json:"execution_source_version"`
	ExecutionProfile       string                              `json:"execution_profile"`
	Authority              PromptApplicationRuntimeAuthorityV2 `json:"authority"`
	InputDigest            string                              `json:"input_digest"`
	InputBytes             int                                 `json:"input_bytes"`
	VariableNames          []string                            `json:"variable_names"`
	VariableNamesDigest    string                              `json:"variable_names_digest"`
	RequestedProtocol      string                              `json:"requested_protocol"`
	SelectedProtocol       string                              `json:"selected_protocol"`
	RequestedModel         string                              `json:"requested_model"`
	SelectedProvider       string                              `json:"selected_provider"`
	SelectedProfile        string                              `json:"selected_profile"`
	SelectedModel          string                              `json:"selected_model"`
	UpstreamModel          string                              `json:"upstream_model"`
	SelectionSource        string                              `json:"selection_source"`
	Status                 string                              `json:"status"`
	FailureCode            string                              `json:"failure_code"`
	FailureSummary         string                              `json:"failure_summary"`
	StartedAt              string                              `json:"started_at"`
	CompletedAt            string                              `json:"completed_at"`
	Output                 string                              `json:"output"`
	Usage                  PromptApplicationRunUsageV6         `json:"usage"`
	SideEffects            WorkflowRunSideEffects              `json:"side_effects"`
	Diagnostic             PromptApplicationRunDiagnosticV6    `json:"diagnostic"`
	RequestID              string                              `json:"request_id"`
	AuditRef               string                              `json:"audit_ref"`
	ActorRef               string                              `json:"actor_ref"`
}

func decodePromptApplicationVNextContract(schemaVersion string, payload []byte) (any, error) {
	var target any
	switch strings.TrimSpace(schemaVersion) {
	case promptApplicationConfigurationDraftV3Schema:
		target = &PromptApplicationConfigurationDraftV3{}
	case promptApplicationPublishCandidateV3Schema:
		target = &PromptApplicationPublishCandidateV3{}
	case promptApplicationRuntimeAssignmentSchema:
		target = &PromptApplicationRuntimeAssignment{}
	case promptApplicationRuntimeAssignmentEventSchema:
		target = &PromptApplicationRuntimeAssignmentEvent{}
	case promptApplicationRuntimeAuthorityV2Schema:
		target = &PromptApplicationRuntimeAuthorityV2{}
	case promptApplicationSessionV2Schema:
		target = &PromptApplicationSessionV2{}
	case promptApplicationSessionTurnV2Schema:
		target = &PromptApplicationSessionTurnV2{}
	case promptApplicationRunV6Schema:
		target = &PromptApplicationRunRecordV6{}
	default:
		return nil, errPromptApplicationVNextContract
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil || decoder.Decode(&struct{}{}) != io.EOF {
		return nil, errPromptApplicationVNextContract
	}
	if validatePromptApplicationVNextValue(target) != nil {
		return nil, errPromptApplicationVNextContract
	}
	return target, nil
}

func validatePromptApplicationVNextValue(value any) error {
	switch contract := value.(type) {
	case *PromptApplicationConfigurationDraftV3:
		return validatePromptApplicationConfigurationDraftV3(*contract)
	case *PromptApplicationPublishCandidateV3:
		return validatePromptApplicationPublishCandidateV3(*contract)
	case *PromptApplicationRuntimeAssignment:
		return validatePromptApplicationRuntimeAssignment(*contract)
	case *PromptApplicationRuntimeAssignmentEvent:
		return validatePromptApplicationRuntimeAssignmentEvent(*contract)
	case *PromptApplicationRuntimeAuthorityV2:
		return validatePromptApplicationRuntimeAuthorityV2(*contract)
	case *PromptApplicationSessionV2:
		return validatePromptApplicationSessionV2(*contract)
	case *PromptApplicationSessionTurnV2:
		return validatePromptApplicationSessionTurnV2(*contract)
	case *PromptApplicationRunRecordV6:
		return validatePromptApplicationRunRecordV6(*contract)
	default:
		return errPromptApplicationVNextContract
	}
}

func validatePromptApplicationConfigurationDraftV3(contract PromptApplicationConfigurationDraftV3) error {
	if contract.SchemaVersion != promptApplicationConfigurationDraftV3Schema || !validPromptApplicationRef(contract.DraftID) ||
		!validPromptApplicationRef(contract.WorkspaceID) || !applicationCatalogIDPattern.MatchString(contract.ApplicationID) ||
		parsePromptApplicationTemplateTimestamp(contract.BaseApplicationUpdatedAt) == nil ||
		contract.ApplicationKind != "prompt_application" || !validPromptApplicationPublicConfiguration(contract.DisplayName, contract.Description, contract.DefaultProtocol, contract.DefaultModel, contract.AllowedProtocols) ||
		!validPromptApplicationTemplateRef(contract.PromptTemplateRef) || contract.DraftVersion < 1 || !workflowRAGDigestPattern.MatchString(contract.DraftDigest) ||
		contract.ValidationSummary.State != applicationDraftValidationValid || !contract.ValidationSummary.IsValid || len(contract.ValidationSummary.Findings) != 0 ||
		!validPromptApplicationTimestampOrder(contract.CreatedAt, contract.UpdatedAt) || !validPromptApplicationAuditRefs(contract.CreatedByActorRef, contract.UpdatedByActorRef, contract.RequestID, contract.AuditRef) {
		return errPromptApplicationVNextContract
	}
	return nil
}

func validatePromptApplicationPublishCandidateV3(contract PromptApplicationPublishCandidateV3) error {
	if contract.SchemaVersion != promptApplicationPublishCandidateV3Schema || !validPromptApplicationRef(contract.CandidateID) ||
		!validPromptApplicationRef(contract.WorkspaceID) || !applicationCatalogIDPattern.MatchString(contract.ApplicationID) || !validPromptApplicationRef(contract.DraftID) ||
		contract.DraftVersion < 1 || !workflowRAGDigestPattern.MatchString(contract.DraftDigest) || parsePromptApplicationTemplateTimestamp(contract.BaseApplicationUpdatedAt) == nil || contract.Configuration.ApplicationKind != "prompt_application" ||
		!validPromptApplicationPublicConfiguration(contract.Configuration.DisplayName, contract.Configuration.Description, contract.Configuration.DefaultProtocol, contract.Configuration.DefaultModel, contract.Configuration.AllowedProtocols) ||
		!validPromptApplicationTemplateRef(contract.Configuration.PromptTemplateRef) || len(contract.EvidenceRequestIDs) > applicationPublishMaxEvidenceRequests ||
		!validPromptApplicationPublishState(contract.CandidateState) || contract.ReviewVersion < 0 || len(contract.Reviews) != contract.ReviewVersion || !validPromptApplicationReviews(contract.Reviews, contract.CandidateState) ||
		!validPromptApplicationPromotionEligibility(contract.PromotionEligibility, contract.CandidateState) ||
		!validPromptApplicationTimestampOrder(contract.CreatedAt, contract.UpdatedAt) || !validPromptApplicationAuditRefs(contract.CreatedByActorRef, contract.UpdatedByActorRef, contract.RequestID, contract.AuditRef) {
		return errPromptApplicationVNextContract
	}
	seen := make(map[string]bool, len(contract.EvidenceRequestIDs))
	for _, ref := range contract.EvidenceRequestIDs {
		if !validPromptApplicationRef(ref) || seen[ref] {
			return errPromptApplicationVNextContract
		}
		seen[ref] = true
	}
	return nil
}

func validatePromptApplicationRuntimeAssignment(contract PromptApplicationRuntimeAssignment) error {
	if contract.SchemaVersion != promptApplicationRuntimeAssignmentSchema || !promptApplicationAssignmentIDPattern.MatchString(contract.AssignmentID) || !validPromptApplicationScope(contract.TenantRef, contract.WorkspaceID, contract.ApplicationID, contract.OwnerSubjectRef) ||
		contract.AssignmentVersion < 1 || (contract.State != "active" && contract.State != "revoked") || !validPromptApplicationLineage(contract.CandidateID, contract.CandidateReviewVersion, contract.DraftID, contract.DraftVersion, contract.DraftDigest, contract.PromptTemplateRef) ||
		!workflowRAGDigestPattern.MatchString(contract.AssignmentDigest) || !validPromptApplicationTimestampOrder(contract.ActivatedAt, contract.UpdatedAt) ||
		!validPromptApplicationAuditRefs(contract.ActivatedByActorRef, contract.UpdatedByActorRef, contract.RequestID, contract.AuditRef) {
		return errPromptApplicationVNextContract
	}
	if contract.State == "active" && contract.RevokedAt != nil || contract.State == "revoked" && (contract.RevokedAt == nil || !validPromptApplicationTimestampOrder(contract.ActivatedAt, *contract.RevokedAt) || !validPromptApplicationTimestampOrder(*contract.RevokedAt, contract.UpdatedAt)) {
		return errPromptApplicationVNextContract
	}
	return nil
}

func validatePromptApplicationRuntimeAssignmentEvent(contract PromptApplicationRuntimeAssignmentEvent) error {
	if contract.SchemaVersion != promptApplicationRuntimeAssignmentEventSchema || !promptApplicationEventIDPattern.MatchString(contract.EventID) || !promptApplicationAssignmentIDPattern.MatchString(contract.AssignmentID) ||
		!validPromptApplicationScope(contract.TenantRef, contract.WorkspaceID, contract.ApplicationID, contract.OwnerSubjectRef) || contract.EventSequence < 1 ||
		(contract.Action != "activate" && contract.Action != "replace" && contract.Action != "revoke") || contract.ExpectedAssignmentVersion < 0 || contract.ResultingAssignmentVersion != contract.ExpectedAssignmentVersion+1 ||
		!validPromptApplicationLineage(contract.CandidateID, contract.CandidateReviewVersion, contract.DraftID, contract.DraftVersion, contract.DraftDigest, contract.PromptTemplateRef) ||
		!workflowRAGDigestPattern.MatchString(contract.AssignmentDigest) || parsePromptApplicationTemplateTimestamp(contract.OccurredAt) == nil ||
		!validPromptApplicationAuditRefs(contract.ActorRef, contract.ActorRef, contract.RequestID, contract.AuditRef) {
		return errPromptApplicationVNextContract
	}
	if contract.Action == "activate" && contract.ExpectedAssignmentVersion != 0 || contract.Action != "activate" && contract.ExpectedAssignmentVersion < 1 {
		return errPromptApplicationVNextContract
	}
	return nil
}

func validatePromptApplicationRuntimeAuthorityV2(contract PromptApplicationRuntimeAuthorityV2) error {
	if contract.SchemaVersion != promptApplicationRuntimeAuthorityV2Schema || contract.ExecutionProfile != promptApplicationInvocationProfile ||
		!applicationCatalogIDPattern.MatchString(contract.ApplicationID) || contract.ApplicationRecordVersion < 1 || contract.ApplicationLifecycle != applicationCatalogLifecycleActive ||
		!validPromptApplicationAuthorityLineage(contract.PromptApplication) || !workflowRAGDigestPattern.MatchString(contract.AuthorityDigest) {
		return errPromptApplicationVNextContract
	}
	want, err := promptApplicationRuntimeAuthorityV2Digest(contract)
	if err != nil || want != contract.AuthorityDigest {
		return errPromptApplicationVNextContract
	}
	return nil
}

func validatePromptApplicationSessionV2(contract PromptApplicationSessionV2) error {
	if contract.SchemaVersion != promptApplicationSessionV2Schema || !applicationSessionIDPattern.MatchString(contract.SessionID) || !validPromptApplicationScope(contract.TenantRef, contract.WorkspaceID, contract.ApplicationID, contract.OwnerSubjectRef) ||
		(contract.State != applicationSessionStateActive && contract.State != applicationSessionStateClosed) || contract.RecordVersion < 1 || contract.ProfileBinding.ExecutionProfile != promptApplicationInvocationProfile ||
		validatePromptApplicationRuntimeAuthorityV2(contract.Authority) != nil || contract.Authority.ApplicationID != contract.ApplicationID || contract.ContentRetention != applicationSessionRetentionPolicy || contract.TurnCount < 0 ||
		!validPromptApplicationTimestampOrder(contract.CreatedAt, contract.UpdatedAt) || !validPromptApplicationAuditRefs(contract.CreatedByActorRef, contract.UpdatedByActorRef, contract.RequestID, contract.AuditRef) {
		return errPromptApplicationVNextContract
	}
	if contract.TurnCount == 0 && contract.LastTurnID != nil || contract.TurnCount > 0 && (contract.LastTurnID == nil || !applicationTurnIDPattern.MatchString(*contract.LastTurnID)) ||
		contract.State == applicationSessionStateActive && contract.ClosedAt != nil || contract.State == applicationSessionStateClosed && (contract.ClosedAt == nil || !validPromptApplicationTimestampOrder(contract.UpdatedAt, *contract.ClosedAt)) {
		return errPromptApplicationVNextContract
	}
	return nil
}

func validatePromptApplicationSessionTurnV2(contract PromptApplicationSessionTurnV2) error {
	terminal := contract.Status != string(WorkflowRunStatusRunning)
	if contract.SchemaVersion != promptApplicationSessionTurnV2Schema || !applicationTurnIDPattern.MatchString(contract.TurnID) || !applicationSessionIDPattern.MatchString(contract.SessionID) || contract.Sequence < 1 ||
		!validPromptApplicationRef(contract.ClientTurnKey) || !validPromptApplicationScope(contract.TenantRef, contract.WorkspaceID, contract.ApplicationID, contract.OwnerSubjectRef) || contract.ExecutionProfile != promptApplicationInvocationProfile ||
		validatePromptApplicationRuntimeAuthorityV2(contract.Authority) != nil || contract.Authority.ApplicationID != contract.ApplicationID || !validPromptApplicationRunStatus(contract.Status) || !workflowRAGDigestPattern.MatchString(contract.InputDigest) || contract.InputBytes < 1 || contract.InputBytes > promptApplicationTemplateMaximumSourceBytes ||
		parsePromptApplicationTemplateTimestamp(contract.StartedAt) == nil || !validPromptApplicationAuditRefs(contract.ActorRef, contract.ActorRef, contract.RequestID, contract.AuditRef) {
		return errPromptApplicationVNextContract
	}
	if !terminal && (contract.CompletedAt != nil || contract.RunRef != nil || contract.FailureCode != "" || contract.FailureSummary != "") || terminal && (contract.CompletedAt == nil || !validPromptApplicationTimestampOrder(contract.StartedAt, *contract.CompletedAt)) {
		return errPromptApplicationVNextContract
	}
	if contract.RunRef != nil && (!workflowHTTPToolRunIDPattern.MatchString(contract.RunRef.RunID) || contract.RunRef.SchemaVersion != promptApplicationRunV6Schema) || contract.Status == string(WorkflowRunStatusSucceeded) && contract.RunRef == nil ||
		contract.Status == string(WorkflowRunStatusSucceeded) && (contract.FailureCode != "" || contract.FailureSummary != "") || terminal && contract.Status != string(WorkflowRunStatusSucceeded) && contract.FailureCode == "" {
		return errPromptApplicationVNextContract
	}
	return nil
}

func validatePromptApplicationRunRecordV6(contract PromptApplicationRunRecordV6) error {
	terminal := contract.Status != string(WorkflowRunStatusRunning)
	if contract.SchemaVersion != promptApplicationRunV6Schema || contract.RecordVersion < 1 || !workflowHTTPToolRunIDPattern.MatchString(contract.RunID) ||
		!validPromptApplicationScope(contract.TenantRef, contract.WorkspaceID, contract.ApplicationID, contract.ActorRef) || contract.ExecutionKind != "prompt_application_invocation" ||
		contract.ExecutionSourceKind != "prompt_application_template" || !promptApplicationTemplateIDPattern.MatchString(contract.ExecutionSourceID) || contract.ExecutionSourceVersion < 1 || contract.ExecutionProfile != promptApplicationInvocationProfile ||
		validatePromptApplicationRuntimeAuthorityV2(contract.Authority) != nil || contract.Authority.ApplicationID != contract.ApplicationID || contract.ExecutionSourceID != contract.Authority.PromptApplication.PromptTemplateRef.TemplateID || contract.ExecutionSourceVersion != contract.Authority.PromptApplication.PromptTemplateRef.TemplateVersion ||
		!workflowRAGDigestPattern.MatchString(contract.InputDigest) || contract.InputBytes < 1 || contract.InputBytes > promptApplicationTemplateMaximumSourceBytes || !validPromptApplicationVariableNames(contract.VariableNames, contract.VariableNamesDigest) ||
		!isApplicationDraftProtocol(contract.RequestedProtocol) || !isApplicationDraftProtocol(contract.SelectedProtocol) || !validPromptApplicationSafeSelection(contract.RequestedModel, contract.SelectedProvider, contract.SelectedProfile, contract.SelectedModel, contract.UpstreamModel, contract.SelectionSource) ||
		!validPromptApplicationRunStatus(contract.Status) || len(contract.FailureSummary) > 256 || parsePromptApplicationTemplateTimestamp(contract.StartedAt) == nil || contract.Output != "" ||
		!validPromptApplicationRunUsage(contract.Usage) || !validPromptApplicationRunSideEffects(contract.SideEffects) || !validPromptApplicationRunDiagnostic(contract.Diagnostic, terminal) ||
		!validPromptApplicationAuditRefs(contract.ActorRef, contract.ActorRef, contract.RequestID, contract.AuditRef) {
		return errPromptApplicationVNextContract
	}
	if terminal && parsePromptApplicationTemplateTimestamp(contract.CompletedAt) == nil || !terminal && contract.CompletedAt != "" {
		return errPromptApplicationVNextContract
	}
	if terminal && !validPromptApplicationTimestampOrder(contract.StartedAt, contract.CompletedAt) || contract.Status == string(WorkflowRunStatusSucceeded) && (contract.FailureCode != "" || contract.FailureSummary != "") || terminal && contract.Status != string(WorkflowRunStatusSucceeded) && contract.FailureCode == "" || !terminal && (contract.FailureCode != "" || contract.FailureSummary != "") {
		return errPromptApplicationVNextContract
	}
	return nil
}

func promptApplicationRuntimeAuthorityV2Digest(contract PromptApplicationRuntimeAuthorityV2) (string, error) {
	contract.AuthorityDigest = ""
	payload, err := json.Marshal(contract)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(digest[:]), nil
}

func validPromptApplicationAuthorityLineage(value PromptApplicationAuthorityV2) bool {
	return promptApplicationAssignmentIDPattern.MatchString(value.AssignmentID) && value.AssignmentVersion > 0 && workflowRAGDigestPattern.MatchString(value.AssignmentDigest) &&
		validPromptApplicationLineage(value.PublishCandidateID, value.PublishReviewVersion, value.DraftID, value.DraftVersion, value.DraftDigest, value.PromptTemplateRef) &&
		isApplicationDraftProtocol(value.DefaultProtocol) && validPromptApplicationRef(value.DefaultModel) && workflowRAGDigestPattern.MatchString(value.ProtocolPolicyDigest) && workflowRAGDigestPattern.MatchString(value.ModelEligibilityDigest)
}

func validPromptApplicationReviews(reviews []ApplicationPublishReviewRecord, candidateState string) bool {
	if len(reviews) == 0 {
		return candidateState == applicationPublishStatePending
	}
	for index, review := range reviews {
		if review.ReviewVersion != index+1 || !isApplicationPublishDecision(review.Decision) || review.State != applicationPublishStateForDecision(review.Decision) || len(review.Reason) < 4 || len(review.Reason) > 500 ||
			parsePromptApplicationTemplateTimestamp(review.ReviewedAt) == nil || !validPromptApplicationAuditRefs(review.ReviewerRef, review.ReviewerRef, review.RequestID, review.AuditRef) {
			return false
		}
	}
	return reviews[len(reviews)-1].State == candidateState
}

func validPromptApplicationPromotionEligibility(eligibility ApplicationPromotionEligibility, candidateState string) bool {
	if len(eligibility.Blockers) > 32 {
		return false
	}
	for _, blocker := range eligibility.Blockers {
		if len(blocker.Code) == 0 || len(blocker.Code) > 160 || len(blocker.Summary) > 256 {
			return false
		}
	}
	if eligibility.Eligible {
		return candidateState == applicationPublishStateApproved && eligibility.Status == "eligible_for_promotion" && len(eligibility.Blockers) == 0
	}
	return eligibility.Status == applicationPublishStatusBlocked && len(eligibility.Blockers) > 0
}

func validPromptApplicationLineage(candidateID string, reviewVersion int, draftID string, draftVersion int, draftDigest string, templateRef PromptApplicationTemplateRef) bool {
	return validPromptApplicationRef(candidateID) && reviewVersion > 0 && validPromptApplicationRef(draftID) && draftVersion > 0 && workflowRAGDigestPattern.MatchString(draftDigest) && validPromptApplicationTemplateRef(templateRef)
}

func validPromptApplicationTemplateRef(ref PromptApplicationTemplateRef) bool {
	return promptApplicationTemplateIDPattern.MatchString(ref.TemplateID) && ref.TemplateVersion > 0 && workflowRAGDigestPattern.MatchString(ref.TemplateDigest)
}

func validPromptApplicationPublicConfiguration(displayName, description, defaultProtocol, defaultModel string, protocols []string) bool {
	if len(strings.TrimSpace(displayName)) < 2 || len(displayName) > 120 || len(description) > 1000 || !validPromptApplicationRef(defaultModel) || applicationDraftStringContainsSecret(displayName) || applicationDraftStringContainsSecret(description) || applicationDraftStringContainsSecret(defaultModel) {
		return false
	}
	normalized := normalizeApplicationDraftProtocols(protocols)
	return len(normalized) != 0 && len(normalized) == len(protocols) && containsApplicationDraftProtocol(normalized, defaultProtocol)
}

func validPromptApplicationScope(tenantRef, workspaceID, applicationID, ownerRef string) bool {
	return validPromptApplicationRef(tenantRef) && validPromptApplicationRef(workspaceID) && applicationCatalogIDPattern.MatchString(applicationID) && validPromptApplicationRef(ownerRef)
}

func validPromptApplicationRef(value string) bool {
	return controlPlaneReadAuthReferencePattern.MatchString(strings.TrimSpace(value))
}

func validPromptApplicationAuditRefs(actorA, actorB, requestID, auditRef string) bool {
	return validPromptApplicationRef(actorA) && validPromptApplicationRef(actorB) && validPromptApplicationRef(requestID) && validPromptApplicationRef(auditRef)
}

func validPromptApplicationTimestampOrder(createdAt, updatedAt string) bool {
	created, updated := parsePromptApplicationTemplateTimestamp(createdAt), parsePromptApplicationTemplateTimestamp(updatedAt)
	return created != nil && updated != nil && !updated.Before(*created)
}

func validPromptApplicationPublishState(state string) bool {
	switch state {
	case applicationPublishStatePending, applicationPublishStateApproved, applicationPublishStateRejected, applicationPublishStateChangesNeeded, applicationPublishStateWithdrawn:
		return true
	default:
		return false
	}
}

func validPromptApplicationRunStatus(status string) bool {
	switch WorkflowRunStatus(status) {
	case WorkflowRunStatusRunning, WorkflowRunStatusSucceeded, WorkflowRunStatusFailed, WorkflowRunStatusCanceled, WorkflowRunStatusOutcomeUnknown:
		return true
	default:
		return false
	}
}

func validPromptApplicationVariableNames(names []string, digest string) bool {
	if len(names) > promptApplicationTemplateMaximumVariables || !sort.StringsAreSorted(names) || !workflowRAGDigestPattern.MatchString(digest) {
		return false
	}
	seen := make(map[string]bool, len(names))
	for _, name := range names {
		if !promptApplicationVariableNamePattern.MatchString(name) || seen[name] {
			return false
		}
		seen[name] = true
	}
	payload, err := json.Marshal(names)
	if err != nil {
		return false
	}
	calculated := sha256.Sum256(payload)
	return digest == "sha256:"+hex.EncodeToString(calculated[:])
}

func validPromptApplicationSafeSelection(values ...string) bool {
	for _, value := range values {
		if !validPromptApplicationRef(value) || strings.Contains(value, "://") || applicationDraftStringContainsSecret(value) {
			return false
		}
	}
	return true
}

func validPromptApplicationRunUsage(usage PromptApplicationRunUsageV6) bool {
	if usage.InputTokens < 0 || usage.OutputTokens < 0 || usage.TotalTokens < 0 || usage.TotalTokens != usage.InputTokens+usage.OutputTokens {
		return false
	}
	return usage.State == "provider_reported" || usage.State == "unavailable" && usage.TotalTokens == 0
}

func validPromptApplicationRunSideEffects(effects WorkflowRunSideEffects) bool {
	return effects.ProviderCalls >= 0 && effects.ProviderCalls <= 1 && effects.RetrievalCalls == 0 && effects.ToolCalls == 0 && effects.ConfirmationCalls == 0 && effects.BusinessWrites == 0 && effects.ReplayWrites == 0
}

func validPromptApplicationRunDiagnostic(diagnostic PromptApplicationRunDiagnosticV6, terminal bool) bool {
	if len(diagnostic.FailureBoundary) > 64 || len(diagnostic.FailureStage) > 64 || len(diagnostic.Summary) > 256 || len(diagnostic.RecommendedReviewAction) > 64 || parsePromptApplicationTemplateTimestamp(diagnostic.ObservedAt) == nil ||
		(diagnostic.TerminalWriteState != "pending" && diagnostic.TerminalWriteState != "stored" && diagnostic.TerminalWriteState != "unknown") || !validPromptApplicationGatewayFailureCategory(diagnostic.GatewayFailureCategory) {
		return false
	}
	if !terminal {
		return diagnostic.FailureBoundary == "" && diagnostic.FailureStage == "" && diagnostic.GatewayFailureCategory == "none"
	}
	return true
}

func validPromptApplicationGatewayFailureCategory(value string) bool {
	switch value {
	case "none", "queue_full", "timeout", "canceled", "worker_crash", "protocol", "provider_failed", "output_unavailable", "unavailable":
		return true
	default:
		return false
	}
}
