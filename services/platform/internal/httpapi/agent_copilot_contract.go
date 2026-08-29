package httpapi

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"regexp"
	"strings"
)

const (
	agentCopilotProfileDraftSchema             = "agent_copilot_profile_draft.v1"
	agentCopilotProfileVersionSchema           = "agent_copilot_profile_version.v1"
	agentCopilotConfigurationDraftV4Schema     = "application_configuration_draft.v4"
	agentCopilotPublishCandidateV4Schema       = "application_publish_candidate.v4"
	agentCopilotRuntimeAssignmentSchema        = "agent_copilot_runtime_assignment.v1"
	agentCopilotRuntimeAssignmentEventSchema   = "agent_copilot_runtime_assignment_event.v1"
	agentCopilotRuntimeAuthorityV3Schema       = "application_runtime_authority.v3"
	agentCopilotSessionV3Schema                = "application_session.v3"
	agentCopilotSessionTurnV3Schema            = "application_session_turn.v3"
	agentCopilotRunV7Schema                    = "workflow_run_record.v7"
	agentCopilotSuggestionProfile              = "agent_copilot_suggestion_v1"
	agentCopilotExecutionSourceKind            = "agent_copilot_profile"
	agentCopilotMaximumProfileSourceBytes      = 64 * 1024
	agentCopilotMaximumInvocationBytes         = 512 * 1024
	agentCopilotMaximumContextBytes            = 128 * 1024
	agentCopilotMaximumArtifactItemBytes       = 128 * 1024
	agentCopilotMaximumArtifactTotalBytes      = 256 * 1024
	agentCopilotMaximumVisibleResponseTextByte = 16 * 1024
)

var (
	errAgentCopilotContract       = errors.New("agent copilot contract mismatch")
	agentCopilotProfileIDPattern  = regexp.MustCompile(`^acpf_[a-z2-7]{16}$`)
	agentCopilotAssignmentPattern = regexp.MustCompile(`^acra_[a-z2-7]{16}$`)
	agentCopilotEventPattern      = regexp.MustCompile(`^acrae_[a-z2-7]{16}$`)
	agentCopilotLocalePattern     = regexp.MustCompile(`^[A-Za-z]{2,8}(-[A-Za-z0-9]{1,8})*$`)
)

type AgentCopilotContextPolicy struct {
	AllowedFields      []string `json:"allowed_fields"`
	MaxBytes           int      `json:"max_bytes"`
	RequireTaskContext bool     `json:"require_task_context"`
}

type AgentCopilotArtifactPolicy struct {
	AllowedKinds  []string `json:"allowed_kinds"`
	AllowedRoles  []string `json:"allowed_roles"`
	MaxCount      int      `json:"max_count"`
	MaxItemBytes  int      `json:"max_item_bytes"`
	MaxTotalBytes int      `json:"max_total_bytes"`
}

type AgentCopilotResponsePolicy struct {
	AllowedActionKinds  []string `json:"allowed_action_kinds"`
	MaxAnswers          int      `json:"max_answers"`
	MaxIssues           int      `json:"max_issues"`
	MaxActions          int      `json:"max_actions"`
	MaxCitations        int      `json:"max_citations"`
	MaxVisibleTextBytes int      `json:"max_visible_text_bytes"`
}

type AgentCopilotRiskPolicy struct {
	Mode                           string   `json:"mode"`
	RequiresConfirmationForActions bool     `json:"requires_confirmation_for_actions"`
	ConfirmationActionKinds        []string `json:"confirmation_action_kinds"`
}

type AgentCopilotToolHintsPolicy struct {
	AllowRetrieval      bool `json:"allow_retrieval"`
	AllowToolCalls      bool `json:"allow_tool_calls"`
	AllowImageReasoning bool `json:"allow_image_reasoning"`
}

type AgentCopilotProfileRef struct {
	ProfileID      string `json:"profile_id"`
	ProfileVersion int    `json:"profile_version"`
	ProfileDigest  string `json:"profile_digest"`
	PolicyDigest   string `json:"policy_digest"`
}

type AgentCopilotProfileSource struct {
	ProfileName     string                      `json:"profile_name"`
	Description     string                      `json:"description"`
	Project         string                      `json:"project"`
	AllowedTasks    []string                    `json:"allowed_tasks"`
	DefaultLocale   string                      `json:"default_locale"`
	AllowedLocales  []string                    `json:"allowed_locales"`
	ContextPolicy   AgentCopilotContextPolicy   `json:"context_policy"`
	ArtifactPolicy  AgentCopilotArtifactPolicy  `json:"artifact_policy"`
	ResponsePolicy  AgentCopilotResponsePolicy  `json:"response_policy"`
	RiskPolicy      AgentCopilotRiskPolicy      `json:"risk_policy"`
	ToolHintsPolicy AgentCopilotToolHintsPolicy `json:"tool_hints_policy"`
}

type AgentCopilotProfileDraftV1 struct {
	SchemaVersion   string `json:"schema_version"`
	ProfileID       string `json:"profile_id"`
	TenantRef       string `json:"tenant_ref"`
	WorkspaceID     string `json:"workspace_id"`
	ApplicationID   string `json:"application_id"`
	OwnerSubjectRef string `json:"owner_subject_ref"`
	AgentCopilotProfileSource
	DraftVersion      int                                     `json:"draft_version"`
	ProfileDigest     string                                  `json:"profile_digest"`
	PolicyDigest      string                                  `json:"policy_digest"`
	ValidationSummary ApplicationConfigurationDraftValidation `json:"validation_summary"`
	CreatedAt         string                                  `json:"created_at"`
	UpdatedAt         string                                  `json:"updated_at"`
	CreatedByActorRef string                                  `json:"created_by_actor_ref"`
	UpdatedByActorRef string                                  `json:"updated_by_actor_ref"`
	RequestID         string                                  `json:"request_id"`
	AuditRef          string                                  `json:"audit_ref"`
}

type AgentCopilotProfileVersionV1 struct {
	SchemaVersion      string `json:"schema_version"`
	ProfileID          string `json:"profile_id"`
	ProfileVersion     int    `json:"profile_version"`
	SourceDraftVersion int    `json:"source_draft_version"`
	TenantRef          string `json:"tenant_ref"`
	WorkspaceID        string `json:"workspace_id"`
	ApplicationID      string `json:"application_id"`
	OwnerSubjectRef    string `json:"owner_subject_ref"`
	AgentCopilotProfileSource
	ProfileDigest       string `json:"profile_digest"`
	PolicyDigest        string `json:"policy_digest"`
	PublishedAt         string `json:"published_at"`
	PublishedByActorRef string `json:"published_by_actor_ref"`
	RequestID           string `json:"request_id"`
	AuditRef            string `json:"audit_ref"`
}

type AgentCopilotConfigurationDraftV4 struct {
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
	AgentCopilotProfileRef   AgentCopilotProfileRef                  `json:"agent_copilot_profile_ref"`
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

type AgentCopilotPublishConfigurationV4 struct {
	DisplayName            string                 `json:"display_name"`
	Description            string                 `json:"description"`
	ApplicationKind        string                 `json:"application_kind"`
	DefaultProtocol        string                 `json:"default_protocol"`
	DefaultModel           string                 `json:"default_model"`
	AllowedProtocols       []string               `json:"allowed_protocols"`
	AgentCopilotProfileRef AgentCopilotProfileRef `json:"agent_copilot_profile_ref"`
}

type AgentCopilotPublishCandidateV4 struct {
	SchemaVersion            string                             `json:"schema_version"`
	CandidateID              string                             `json:"candidate_id"`
	WorkspaceID              string                             `json:"workspace_id"`
	ApplicationID            string                             `json:"application_id"`
	DraftID                  string                             `json:"draft_id"`
	DraftVersion             int                                `json:"draft_version"`
	DraftDigest              string                             `json:"draft_digest"`
	BaseApplicationUpdatedAt string                             `json:"base_application_updated_at"`
	Configuration            AgentCopilotPublishConfigurationV4 `json:"configuration"`
	EvidenceRequestIDs       []string                           `json:"evidence_request_ids"`
	CandidateState           string                             `json:"candidate_state"`
	ReviewVersion            int                                `json:"review_version"`
	Reviews                  []ApplicationPublishReviewRecord   `json:"reviews"`
	PromotionEligibility     ApplicationPromotionEligibility    `json:"promotion_eligibility"`
	CreatedAt                string                             `json:"created_at"`
	UpdatedAt                string                             `json:"updated_at"`
	CreatedByActorRef        string                             `json:"created_by_actor_ref"`
	UpdatedByActorRef        string                             `json:"updated_by_actor_ref"`
	RequestID                string                             `json:"request_id"`
	AuditRef                 string                             `json:"audit_ref"`
}

type AgentCopilotRuntimeAssignmentV1 struct {
	SchemaVersion          string                              `json:"schema_version"`
	AssignmentID           string                              `json:"assignment_id"`
	TenantRef              string                              `json:"tenant_ref"`
	WorkspaceID            string                              `json:"workspace_id"`
	ApplicationID          string                              `json:"application_id"`
	OwnerSubjectRef        string                              `json:"owner_subject_ref"`
	AssignmentVersion      int                                 `json:"assignment_version"`
	State                  string                              `json:"state"`
	CandidateID            string                              `json:"candidate_id"`
	CandidateReviewVersion int                                 `json:"candidate_review_version"`
	DraftID                string                              `json:"draft_id"`
	DraftVersion           int                                 `json:"draft_version"`
	DraftDigest            string                              `json:"draft_digest"`
	AgentCopilotProfileRef AgentCopilotProfileRef              `json:"agent_copilot_profile_ref"`
	AssignmentDigest       string                              `json:"assignment_digest"`
	ActivatedAt            string                              `json:"activated_at"`
	UpdatedAt              string                              `json:"updated_at"`
	RevokedAt              *string                             `json:"revoked_at"`
	ActivatedByActorRef    string                              `json:"activated_by_actor_ref"`
	UpdatedByActorRef      string                              `json:"updated_by_actor_ref"`
	RequestID              string                              `json:"request_id"`
	AuditRef               string                              `json:"audit_ref"`
	ActionSafety           *ActionSafetyAssignmentProjectionV1 `json:"-"`
}

type AgentCopilotRuntimeAssignmentEventV1 struct {
	SchemaVersion              string                              `json:"schema_version"`
	EventID                    string                              `json:"event_id"`
	AssignmentID               string                              `json:"assignment_id"`
	TenantRef                  string                              `json:"tenant_ref"`
	WorkspaceID                string                              `json:"workspace_id"`
	ApplicationID              string                              `json:"application_id"`
	OwnerSubjectRef            string                              `json:"owner_subject_ref"`
	EventSequence              int                                 `json:"event_sequence"`
	Action                     string                              `json:"action"`
	ExpectedAssignmentVersion  int                                 `json:"expected_assignment_version"`
	ResultingAssignmentVersion int                                 `json:"resulting_assignment_version"`
	CandidateID                string                              `json:"candidate_id"`
	CandidateReviewVersion     int                                 `json:"candidate_review_version"`
	DraftID                    string                              `json:"draft_id"`
	DraftVersion               int                                 `json:"draft_version"`
	DraftDigest                string                              `json:"draft_digest"`
	AgentCopilotProfileRef     AgentCopilotProfileRef              `json:"agent_copilot_profile_ref"`
	AssignmentDigest           string                              `json:"assignment_digest"`
	OccurredAt                 string                              `json:"occurred_at"`
	ActorRef                   string                              `json:"actor_ref"`
	RequestID                  string                              `json:"request_id"`
	AuditRef                   string                              `json:"audit_ref"`
	ActionSafety               *ActionSafetyAssignmentProjectionV1 `json:"-"`
}

type AgentCopilotAuthorityV3 struct {
	AssignmentID           string                 `json:"assignment_id"`
	AssignmentVersion      int                    `json:"assignment_version"`
	AssignmentDigest       string                 `json:"assignment_digest"`
	PublishCandidateID     string                 `json:"publish_candidate_id"`
	PublishReviewVersion   int                    `json:"publish_review_version"`
	DraftID                string                 `json:"draft_id"`
	DraftVersion           int                    `json:"draft_version"`
	DraftDigest            string                 `json:"draft_digest"`
	AgentCopilotProfileRef AgentCopilotProfileRef `json:"agent_copilot_profile_ref"`
	Project                string                 `json:"project"`
	AllowedTasksDigest     string                 `json:"allowed_tasks_digest"`
	DefaultProtocol        string                 `json:"default_protocol"`
	DefaultModel           string                 `json:"default_model"`
	ProtocolPolicyDigest   string                 `json:"protocol_policy_digest"`
	ModelEligibilityDigest string                 `json:"model_eligibility_digest"`
}

type AgentCopilotRuntimeAuthorityV3 struct {
	SchemaVersion            string                  `json:"schema_version"`
	ExecutionProfile         string                  `json:"execution_profile"`
	ApplicationID            string                  `json:"application_id"`
	ApplicationRecordVersion int                     `json:"application_record_version"`
	ApplicationLifecycle     string                  `json:"application_lifecycle"`
	AgentCopilot             AgentCopilotAuthorityV3 `json:"agent_copilot"`
	AuthorityDigest          string                  `json:"authority_digest"`
}

type AgentCopilotSessionV3 struct {
	SchemaVersion     string                                   `json:"schema_version"`
	SessionID         string                                   `json:"session_id"`
	TenantRef         string                                   `json:"tenant_ref"`
	WorkspaceID       string                                   `json:"workspace_id"`
	ApplicationID     string                                   `json:"application_id"`
	OwnerSubjectRef   string                                   `json:"owner_subject_ref"`
	State             string                                   `json:"state"`
	RecordVersion     int                                      `json:"record_version"`
	ProfileBinding    PromptApplicationSessionProfileBindingV2 `json:"profile_binding"`
	Authority         AgentCopilotRuntimeAuthorityV3           `json:"authority"`
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

type AgentCopilotRunRefV7 struct {
	RunID         string `json:"run_id"`
	SchemaVersion string `json:"schema_version"`
}

type AgentCopilotRunSideEffectsV7 struct {
	RetrievalCalls    int `json:"retrieval_calls"`
	ProviderCalls     int `json:"provider_calls"`
	ToolCalls         int `json:"tool_calls"`
	ConfirmationCalls int `json:"confirmation_calls"`
	BusinessWrites    int `json:"business_writes"`
	ReplayWrites      int `json:"replay_writes"`
}

type AgentCopilotSessionTurnV3 struct {
	SchemaVersion    string                         `json:"schema_version"`
	TurnID           string                         `json:"turn_id"`
	SessionID        string                         `json:"session_id"`
	Sequence         int                            `json:"sequence"`
	ClientTurnKey    string                         `json:"client_turn_key"`
	TenantRef        string                         `json:"tenant_ref"`
	WorkspaceID      string                         `json:"workspace_id"`
	ApplicationID    string                         `json:"application_id"`
	OwnerSubjectRef  string                         `json:"owner_subject_ref"`
	ExecutionProfile string                         `json:"execution_profile"`
	Authority        AgentCopilotRuntimeAuthorityV3 `json:"authority"`
	Status           string                         `json:"status"`
	InputDigest      string                         `json:"input_digest"`
	InputBytes       int                            `json:"input_bytes"`
	RunRef           *AgentCopilotRunRefV7          `json:"run_ref"`
	FailureCode      string                         `json:"failure_code"`
	FailureSummary   string                         `json:"failure_summary"`
	StartedAt        string                         `json:"started_at"`
	CompletedAt      *string                        `json:"completed_at"`
	ActorRef         string                         `json:"actor_ref"`
	RequestID        string                         `json:"request_id"`
	AuditRef         string                         `json:"audit_ref"`
}

type AgentCopilotRunRecordV7 struct {
	SchemaVersion          string                           `json:"schema_version"`
	RecordVersion          int                              `json:"record_version"`
	RunID                  string                           `json:"run_id"`
	TenantRef              string                           `json:"tenant_ref"`
	WorkspaceID            string                           `json:"workspace_id"`
	ApplicationID          string                           `json:"application_id"`
	ExecutionKind          string                           `json:"execution_kind"`
	ExecutionSourceKind    string                           `json:"execution_source_kind"`
	ExecutionSourceID      string                           `json:"execution_source_id"`
	ExecutionSourceVersion int                              `json:"execution_source_version"`
	ExecutionProfile       string                           `json:"execution_profile"`
	Authority              AgentCopilotRuntimeAuthorityV3   `json:"authority"`
	Project                string                           `json:"project"`
	Task                   string                           `json:"task"`
	Locale                 string                           `json:"locale"`
	InputDigest            string                           `json:"input_digest"`
	InputBytes             int                              `json:"input_bytes"`
	ContextBytes           int                              `json:"context_bytes"`
	ArtifactCount          int                              `json:"artifact_count"`
	ArtifactBytes          int                              `json:"artifact_bytes"`
	RequestedProtocol      string                           `json:"requested_protocol"`
	SelectedProtocol       string                           `json:"selected_protocol"`
	RequestedModel         string                           `json:"requested_model"`
	SelectedProvider       string                           `json:"selected_provider"`
	SelectedProfile        string                           `json:"selected_profile"`
	SelectedModel          string                           `json:"selected_model"`
	UpstreamModel          string                           `json:"upstream_model"`
	SelectionSource        string                           `json:"selection_source"`
	ResponseStatus         string                           `json:"response_status"`
	ResponseDigest         string                           `json:"response_digest"`
	AnswerCount            int                              `json:"answer_count"`
	IssueCount             int                              `json:"issue_count"`
	ActionCount            int                              `json:"action_count"`
	CitationCount          int                              `json:"citation_count"`
	RiskLevel              string                           `json:"risk_level"`
	RequiresConfirmation   bool                             `json:"requires_confirmation"`
	Status                 string                           `json:"status"`
	FailureCode            string                           `json:"failure_code"`
	FailureSummary         string                           `json:"failure_summary"`
	StartedAt              string                           `json:"started_at"`
	CompletedAt            string                           `json:"completed_at"`
	Output                 string                           `json:"output"`
	Usage                  PromptApplicationRunUsageV6      `json:"usage"`
	SideEffects            AgentCopilotRunSideEffectsV7     `json:"side_effects"`
	Diagnostic             PromptApplicationRunDiagnosticV6 `json:"diagnostic"`
	RequestID              string                           `json:"request_id"`
	AuditRef               string                           `json:"audit_ref"`
	ActorRef               string                           `json:"actor_ref"`
}

func decodeAgentCopilotContract(schemaVersion string, payload []byte) (any, error) {
	var target any
	switch strings.TrimSpace(schemaVersion) {
	case agentCopilotProfileDraftSchema:
		target = &AgentCopilotProfileDraftV1{}
	case agentCopilotProfileVersionSchema:
		target = &AgentCopilotProfileVersionV1{}
	case agentCopilotConfigurationDraftV4Schema:
		target = &AgentCopilotConfigurationDraftV4{}
	case agentCopilotPublishCandidateV4Schema:
		target = &AgentCopilotPublishCandidateV4{}
	case agentCopilotRuntimeAssignmentSchema:
		target = &AgentCopilotRuntimeAssignmentV1{}
	case agentCopilotRuntimeAssignmentEventSchema:
		target = &AgentCopilotRuntimeAssignmentEventV1{}
	case agentCopilotRuntimeAuthorityV3Schema:
		target = &AgentCopilotRuntimeAuthorityV3{}
	case agentCopilotSessionV3Schema:
		target = &AgentCopilotSessionV3{}
	case agentCopilotSessionTurnV3Schema:
		target = &AgentCopilotSessionTurnV3{}
	case agentCopilotRunV7Schema:
		target = &AgentCopilotRunRecordV7{}
	default:
		return nil, errAgentCopilotContract
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil || decoder.Decode(&struct{}{}) != io.EOF || validateAgentCopilotContract(target) != nil {
		return nil, errAgentCopilotContract
	}
	return target, nil
}

func validateAgentCopilotContract(value any) error {
	switch contract := value.(type) {
	case *AgentCopilotProfileDraftV1:
		return validateAgentCopilotProfileDraft(*contract)
	case *AgentCopilotProfileVersionV1:
		return validateAgentCopilotProfileVersion(*contract)
	case *AgentCopilotConfigurationDraftV4:
		return validateAgentCopilotConfigurationDraft(*contract)
	case *AgentCopilotPublishCandidateV4:
		return validateAgentCopilotPublishCandidate(*contract)
	case *AgentCopilotRuntimeAssignmentV1:
		return validateAgentCopilotRuntimeAssignment(*contract)
	case *AgentCopilotRuntimeAssignmentEventV1:
		return validateAgentCopilotRuntimeAssignmentEvent(*contract)
	case *AgentCopilotRuntimeAuthorityV3:
		return validateAgentCopilotRuntimeAuthority(*contract)
	case *AgentCopilotSessionV3:
		return validateAgentCopilotSession(*contract)
	case *AgentCopilotSessionTurnV3:
		return validateAgentCopilotSessionTurn(*contract)
	case *AgentCopilotRunRecordV7:
		return validateAgentCopilotRun(*contract)
	default:
		return errAgentCopilotContract
	}
}

func validateAgentCopilotProfileDraft(value AgentCopilotProfileDraftV1) error {
	compiled, findings := CompileAgentCopilotProfileSource(value.AgentCopilotProfileSource)
	if value.SchemaVersion != agentCopilotProfileDraftSchema || !validAgentCopilotScope(value.TenantRef, value.WorkspaceID, value.ApplicationID, value.OwnerSubjectRef) ||
		!agentCopilotProfileIDPattern.MatchString(value.ProfileID) || len(findings) != 0 || !agentCopilotProfileSourceIsCanonical(value.AgentCopilotProfileSource, compiled.Source) ||
		value.DraftVersion < 1 || value.ProfileDigest != compiled.ProfileDigest || value.PolicyDigest != compiled.PolicyDigest ||
		value.ValidationSummary.State != applicationDraftValidationValid || !value.ValidationSummary.IsValid || len(value.ValidationSummary.Findings) != 0 ||
		!validPromptApplicationTimestampOrder(value.CreatedAt, value.UpdatedAt) || !validPromptApplicationAuditRefs(value.CreatedByActorRef, value.UpdatedByActorRef, value.RequestID, value.AuditRef) {
		return errAgentCopilotContract
	}
	return nil
}

func validateAgentCopilotProfileVersion(value AgentCopilotProfileVersionV1) error {
	compiled, findings := CompileAgentCopilotProfileSource(value.AgentCopilotProfileSource)
	if value.SchemaVersion != agentCopilotProfileVersionSchema || !validAgentCopilotScope(value.TenantRef, value.WorkspaceID, value.ApplicationID, value.OwnerSubjectRef) ||
		!agentCopilotProfileIDPattern.MatchString(value.ProfileID) || value.ProfileVersion < 1 || value.SourceDraftVersion < 1 ||
		len(findings) != 0 || !agentCopilotProfileSourceIsCanonical(value.AgentCopilotProfileSource, compiled.Source) ||
		value.ProfileDigest != compiled.ProfileDigest || value.PolicyDigest != compiled.PolicyDigest ||
		parsePromptApplicationTemplateTimestamp(value.PublishedAt) == nil || !validPromptApplicationAuditRefs(value.PublishedByActorRef, value.PublishedByActorRef, value.RequestID, value.AuditRef) {
		return errAgentCopilotContract
	}
	return nil
}

func validateAgentCopilotConfigurationDraft(value AgentCopilotConfigurationDraftV4) error {
	if value.SchemaVersion != agentCopilotConfigurationDraftV4Schema || !validPromptApplicationRef(value.DraftID) || !validPromptApplicationRef(value.WorkspaceID) ||
		!applicationCatalogIDPattern.MatchString(value.ApplicationID) || parsePromptApplicationTemplateTimestamp(value.BaseApplicationUpdatedAt) == nil ||
		value.ApplicationKind != "agent" || !validAgentCopilotPublicConfiguration(value.DisplayName, value.Description, value.DefaultProtocol, value.DefaultModel, value.AllowedProtocols, value.AgentCopilotProfileRef) ||
		value.DraftVersion < 1 || !workflowRAGDigestPattern.MatchString(value.DraftDigest) || value.ValidationSummary.State != applicationDraftValidationValid || !value.ValidationSummary.IsValid || len(value.ValidationSummary.Findings) != 0 ||
		!validPromptApplicationTimestampOrder(value.CreatedAt, value.UpdatedAt) || !validPromptApplicationAuditRefs(value.CreatedByActorRef, value.UpdatedByActorRef, value.RequestID, value.AuditRef) {
		return errAgentCopilotContract
	}
	return nil
}

func validateAgentCopilotPublishCandidate(value AgentCopilotPublishCandidateV4) error {
	config := value.Configuration
	if value.SchemaVersion != agentCopilotPublishCandidateV4Schema || !validPromptApplicationRef(value.CandidateID) || !validPromptApplicationRef(value.WorkspaceID) ||
		!applicationCatalogIDPattern.MatchString(value.ApplicationID) || !validPromptApplicationRef(value.DraftID) || value.DraftVersion < 1 || !workflowRAGDigestPattern.MatchString(value.DraftDigest) ||
		parsePromptApplicationTemplateTimestamp(value.BaseApplicationUpdatedAt) == nil || config.ApplicationKind != "agent" ||
		!validAgentCopilotPublicConfiguration(config.DisplayName, config.Description, config.DefaultProtocol, config.DefaultModel, config.AllowedProtocols, config.AgentCopilotProfileRef) ||
		len(value.EvidenceRequestIDs) > applicationPublishMaxEvidenceRequests || !validPromptApplicationPublishState(value.CandidateState) ||
		value.ReviewVersion < 0 || len(value.Reviews) != value.ReviewVersion || !validPromptApplicationReviews(value.Reviews, value.CandidateState) ||
		!validPromptApplicationPromotionEligibility(value.PromotionEligibility, value.CandidateState) ||
		!validPromptApplicationTimestampOrder(value.CreatedAt, value.UpdatedAt) || !validPromptApplicationAuditRefs(value.CreatedByActorRef, value.UpdatedByActorRef, value.RequestID, value.AuditRef) {
		return errAgentCopilotContract
	}
	seen := map[string]bool{}
	for _, ref := range value.EvidenceRequestIDs {
		if !validPromptApplicationRef(ref) || seen[ref] {
			return errAgentCopilotContract
		}
		seen[ref] = true
	}
	return nil
}

func validateAgentCopilotRuntimeAssignment(value AgentCopilotRuntimeAssignmentV1) error {
	if value.SchemaVersion != agentCopilotRuntimeAssignmentSchema || !agentCopilotAssignmentPattern.MatchString(value.AssignmentID) || !validAgentCopilotScope(value.TenantRef, value.WorkspaceID, value.ApplicationID, value.OwnerSubjectRef) ||
		value.AssignmentVersion < 1 || value.State != "active" && value.State != "revoked" || !validAgentCopilotLineage(value.CandidateID, value.CandidateReviewVersion, value.DraftID, value.DraftVersion, value.DraftDigest, value.AgentCopilotProfileRef) ||
		!workflowRAGDigestPattern.MatchString(value.AssignmentDigest) || !validPromptApplicationTimestampOrder(value.ActivatedAt, value.UpdatedAt) ||
		!validPromptApplicationAuditRefs(value.ActivatedByActorRef, value.UpdatedByActorRef, value.RequestID, value.AuditRef) {
		return errAgentCopilotContract
	}
	if value.State == "active" && value.RevokedAt != nil || value.State == "revoked" && (value.RevokedAt == nil || !validPromptApplicationTimestampOrder(value.ActivatedAt, *value.RevokedAt) || !validPromptApplicationTimestampOrder(*value.RevokedAt, value.UpdatedAt)) {
		return errAgentCopilotContract
	}
	if value.ActionSafety != nil && (validateActionSafetyAssignmentProjection(*value.ActionSafety) != nil ||
		value.ActionSafety.AssignmentVersion != value.AssignmentVersion) {
		return errAgentCopilotContract
	}
	return nil
}

func validateAgentCopilotRuntimeAssignmentEvent(value AgentCopilotRuntimeAssignmentEventV1) error {
	if value.SchemaVersion != agentCopilotRuntimeAssignmentEventSchema || !agentCopilotEventPattern.MatchString(value.EventID) || !agentCopilotAssignmentPattern.MatchString(value.AssignmentID) ||
		!validAgentCopilotScope(value.TenantRef, value.WorkspaceID, value.ApplicationID, value.OwnerSubjectRef) || value.EventSequence < 1 ||
		value.Action != "activate" && value.Action != "replace" && value.Action != "revoke" || value.ExpectedAssignmentVersion < 0 || value.ResultingAssignmentVersion != value.ExpectedAssignmentVersion+1 ||
		!validAgentCopilotLineage(value.CandidateID, value.CandidateReviewVersion, value.DraftID, value.DraftVersion, value.DraftDigest, value.AgentCopilotProfileRef) ||
		!workflowRAGDigestPattern.MatchString(value.AssignmentDigest) || parsePromptApplicationTemplateTimestamp(value.OccurredAt) == nil ||
		!validPromptApplicationAuditRefs(value.ActorRef, value.ActorRef, value.RequestID, value.AuditRef) {
		return errAgentCopilotContract
	}
	if value.Action == "activate" && value.ExpectedAssignmentVersion != 0 || value.Action != "activate" && value.ExpectedAssignmentVersion < 1 {
		return errAgentCopilotContract
	}
	if value.ActionSafety != nil && (validateActionSafetyAssignmentProjection(*value.ActionSafety) != nil ||
		value.ActionSafety.AssignmentVersion != value.ResultingAssignmentVersion) {
		return errAgentCopilotContract
	}
	return nil
}

func validateAgentCopilotRuntimeAuthority(value AgentCopilotRuntimeAuthorityV3) error {
	if value.SchemaVersion != agentCopilotRuntimeAuthorityV3Schema || value.ExecutionProfile != agentCopilotSuggestionProfile ||
		!applicationCatalogIDPattern.MatchString(value.ApplicationID) || value.ApplicationRecordVersion < 1 || value.ApplicationLifecycle != applicationCatalogLifecycleActive ||
		!validAgentCopilotAuthorityLineage(value.AgentCopilot) || !workflowRAGDigestPattern.MatchString(value.AuthorityDigest) {
		return errAgentCopilotContract
	}
	want, err := agentCopilotAuthorityDigest(value)
	if err != nil || want != value.AuthorityDigest {
		return errAgentCopilotContract
	}
	return nil
}

func validateAgentCopilotSession(value AgentCopilotSessionV3) error {
	if value.SchemaVersion != agentCopilotSessionV3Schema || !applicationSessionIDPattern.MatchString(value.SessionID) || !validAgentCopilotScope(value.TenantRef, value.WorkspaceID, value.ApplicationID, value.OwnerSubjectRef) ||
		value.State != applicationSessionStateActive && value.State != applicationSessionStateClosed || value.RecordVersion < 1 || value.ProfileBinding.ExecutionProfile != agentCopilotSuggestionProfile ||
		validateAgentCopilotRuntimeAuthority(value.Authority) != nil || value.Authority.ApplicationID != value.ApplicationID || value.ContentRetention != applicationSessionRetentionPolicy || value.TurnCount < 0 ||
		!validPromptApplicationTimestampOrder(value.CreatedAt, value.UpdatedAt) || !validPromptApplicationAuditRefs(value.CreatedByActorRef, value.UpdatedByActorRef, value.RequestID, value.AuditRef) {
		return errAgentCopilotContract
	}
	if value.TurnCount == 0 && value.LastTurnID != nil || value.TurnCount > 0 && (value.LastTurnID == nil || !applicationTurnIDPattern.MatchString(*value.LastTurnID)) ||
		value.State == applicationSessionStateActive && value.ClosedAt != nil || value.State == applicationSessionStateClosed && (value.ClosedAt == nil || !validPromptApplicationTimestampOrder(value.UpdatedAt, *value.ClosedAt)) {
		return errAgentCopilotContract
	}
	return nil
}

func validateAgentCopilotSessionTurn(value AgentCopilotSessionTurnV3) error {
	terminal := value.Status != string(WorkflowRunStatusRunning)
	if value.SchemaVersion != agentCopilotSessionTurnV3Schema || !applicationTurnIDPattern.MatchString(value.TurnID) || !applicationSessionIDPattern.MatchString(value.SessionID) || value.Sequence < 1 ||
		!validPromptApplicationRef(value.ClientTurnKey) || !validAgentCopilotScope(value.TenantRef, value.WorkspaceID, value.ApplicationID, value.OwnerSubjectRef) || value.ExecutionProfile != agentCopilotSuggestionProfile ||
		validateAgentCopilotRuntimeAuthority(value.Authority) != nil || value.Authority.ApplicationID != value.ApplicationID || !validPromptApplicationRunStatus(value.Status) ||
		!workflowRAGDigestPattern.MatchString(value.InputDigest) || value.InputBytes < 1 || value.InputBytes > agentCopilotMaximumInvocationBytes ||
		parsePromptApplicationTemplateTimestamp(value.StartedAt) == nil || !validPromptApplicationAuditRefs(value.ActorRef, value.ActorRef, value.RequestID, value.AuditRef) {
		return errAgentCopilotContract
	}
	if !terminal && (value.CompletedAt != nil || value.RunRef != nil || value.FailureCode != "" || value.FailureSummary != "") || terminal && (value.CompletedAt == nil || !validPromptApplicationTimestampOrder(value.StartedAt, *value.CompletedAt)) {
		return errAgentCopilotContract
	}
	if value.RunRef != nil && (!workflowHTTPToolRunIDPattern.MatchString(value.RunRef.RunID) || value.RunRef.SchemaVersion != agentCopilotRunV7Schema) ||
		value.Status == string(WorkflowRunStatusSucceeded) && (value.RunRef == nil || value.FailureCode != "" || value.FailureSummary != "") ||
		terminal && value.Status != string(WorkflowRunStatusSucceeded) && value.FailureCode == "" {
		return errAgentCopilotContract
	}
	return nil
}

func validateAgentCopilotRun(value AgentCopilotRunRecordV7) error {
	terminal := value.Status != string(WorkflowRunStatusRunning)
	ref := value.Authority.AgentCopilot.AgentCopilotProfileRef
	if value.SchemaVersion != agentCopilotRunV7Schema || value.RecordVersion < 1 || !workflowHTTPToolRunIDPattern.MatchString(value.RunID) ||
		!validAgentCopilotScope(value.TenantRef, value.WorkspaceID, value.ApplicationID, value.ActorRef) || value.ExecutionKind != "agent_copilot_suggestion" ||
		value.ExecutionSourceKind != agentCopilotExecutionSourceKind || !agentCopilotProfileIDPattern.MatchString(value.ExecutionSourceID) || value.ExecutionSourceVersion < 1 || value.ExecutionProfile != agentCopilotSuggestionProfile ||
		validateAgentCopilotRuntimeAuthority(value.Authority) != nil || value.Authority.ApplicationID != value.ApplicationID || value.ExecutionSourceID != ref.ProfileID || value.ExecutionSourceVersion != ref.ProfileVersion ||
		value.Project != value.Authority.AgentCopilot.Project || !validAgentCopilotProjectTask(value.Project, value.Task) || !agentCopilotLocalePattern.MatchString(value.Locale) ||
		!workflowRAGDigestPattern.MatchString(value.InputDigest) || value.InputBytes < 1 || value.InputBytes > agentCopilotMaximumInvocationBytes ||
		value.ContextBytes < 0 || value.ContextBytes > agentCopilotMaximumContextBytes || value.ArtifactCount < 0 || value.ArtifactCount > 16 || value.ArtifactBytes < 0 || value.ArtifactBytes > agentCopilotMaximumArtifactTotalBytes ||
		!isApplicationDraftProtocol(value.RequestedProtocol) || !isApplicationDraftProtocol(value.SelectedProtocol) ||
		!validPromptApplicationSafeSelection(value.RequestedModel, value.SelectedProvider, value.SelectedProfile, value.SelectedModel, value.UpstreamModel, value.SelectionSource) ||
		!validAgentCopilotResponseSummary(value) || !validPromptApplicationRunStatus(value.Status) || len(value.FailureSummary) > 256 ||
		parsePromptApplicationTemplateTimestamp(value.StartedAt) == nil || value.Output != "" || !validPromptApplicationRunUsage(value.Usage) ||
		!validAgentCopilotSideEffects(value.SideEffects) || !validPromptApplicationRunDiagnostic(value.Diagnostic, terminal) ||
		!validPromptApplicationAuditRefs(value.ActorRef, value.ActorRef, value.RequestID, value.AuditRef) {
		return errAgentCopilotContract
	}
	if terminal && parsePromptApplicationTemplateTimestamp(value.CompletedAt) == nil || !terminal && value.CompletedAt != "" ||
		terminal && !validPromptApplicationTimestampOrder(value.StartedAt, value.CompletedAt) ||
		value.Status == string(WorkflowRunStatusSucceeded) && (value.FailureCode != "" || value.FailureSummary != "" || value.ResponseStatus != "ok" && value.ResponseStatus != "partial" || !workflowRAGDigestPattern.MatchString(value.ResponseDigest)) ||
		terminal && value.Status != string(WorkflowRunStatusSucceeded) && value.FailureCode == "" ||
		!terminal && (value.FailureCode != "" || value.FailureSummary != "" || value.ResponseStatus != "unavailable" || value.ResponseDigest != "") {
		return errAgentCopilotContract
	}
	return nil
}

func validAgentCopilotProjectTask(project, task string) bool {
	return agentCopilotContainsString(agentCopilotCanonicalTasks(project), task)
}

func agentCopilotContainsString(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func validAgentCopilotProfileRef(value AgentCopilotProfileRef) bool {
	return agentCopilotProfileIDPattern.MatchString(value.ProfileID) && value.ProfileVersion > 0 && validAgentCopilotDigestPair(value.ProfileDigest, value.PolicyDigest)
}

func validAgentCopilotDigestPair(first, second string) bool {
	return workflowRAGDigestPattern.MatchString(first) && workflowRAGDigestPattern.MatchString(second)
}

func validAgentCopilotPublicConfiguration(displayName, description, defaultProtocol, defaultModel string, protocols []string, ref AgentCopilotProfileRef) bool {
	return validPromptApplicationPublicConfiguration(displayName, description, defaultProtocol, defaultModel, protocols) && validAgentCopilotProfileRef(ref)
}

func validAgentCopilotScope(tenantRef, workspaceID, applicationID, ownerRef string) bool {
	return validPromptApplicationScope(tenantRef, workspaceID, applicationID, ownerRef)
}

func validAgentCopilotLineage(candidateID string, reviewVersion int, draftID string, draftVersion int, draftDigest string, ref AgentCopilotProfileRef) bool {
	return validPromptApplicationRef(candidateID) && reviewVersion > 0 && validPromptApplicationRef(draftID) && draftVersion > 0 &&
		workflowRAGDigestPattern.MatchString(draftDigest) && validAgentCopilotProfileRef(ref)
}

func validAgentCopilotAuthorityLineage(value AgentCopilotAuthorityV3) bool {
	return agentCopilotAssignmentPattern.MatchString(value.AssignmentID) && value.AssignmentVersion > 0 && workflowRAGDigestPattern.MatchString(value.AssignmentDigest) &&
		validAgentCopilotLineage(value.PublishCandidateID, value.PublishReviewVersion, value.DraftID, value.DraftVersion, value.DraftDigest, value.AgentCopilotProfileRef) &&
		(value.Project == "radishflow" || value.Project == "radish") && workflowRAGDigestPattern.MatchString(value.AllowedTasksDigest) &&
		isApplicationDraftProtocol(value.DefaultProtocol) && validPromptApplicationRef(value.DefaultModel) &&
		validAgentCopilotDigestPair(value.ProtocolPolicyDigest, value.ModelEligibilityDigest)
}

func agentCopilotAuthorityDigest(value AgentCopilotRuntimeAuthorityV3) (string, error) {
	value.AuthorityDigest = ""
	payload, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(digest[:]), nil
}

func validAgentCopilotResponseSummary(value AgentCopilotRunRecordV7) bool {
	if !agentCopilotContainsString([]string{"unavailable", "ok", "partial", "failed"}, value.ResponseStatus) ||
		value.AnswerCount < 0 || value.AnswerCount > 64 || value.IssueCount < 0 || value.IssueCount > 64 ||
		value.ActionCount < 0 || value.ActionCount > 64 || value.CitationCount < 0 || value.CitationCount > 64 ||
		!agentCopilotContainsString([]string{"low", "medium", "high"}, value.RiskLevel) {
		return false
	}
	return value.ActionCount == 0 || value.RequiresConfirmation
}

func validAgentCopilotSideEffects(value AgentCopilotRunSideEffectsV7) bool {
	return value.RetrievalCalls == 0 && value.ProviderCalls >= 0 && value.ProviderCalls <= 1 && value.ToolCalls == 0 &&
		value.ConfirmationCalls == 0 && value.BusinessWrites == 0 && value.ReplayWrites == 0
}
