package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	workflowRAGPromotionCandidateSchemaVersion = "workflow_rag_knowledge_promotion_candidate.v1"
	workflowRAGPromotionDecisionSchemaVersion  = "workflow_rag_knowledge_promotion_decision.v1"
	workflowRAGApplicationBindingSchemaVersion = "workflow_rag_application_binding.v1"
	workflowRAGPromotionAuditSchemaVersion     = "workflow_rag_knowledge_promotion_audit.v1"

	workflowRAGPromotionStatePending  = "pending"
	workflowRAGPromotionStateDeferred = "deferred"
	workflowRAGPromotionStateApproved = "approved"
	workflowRAGPromotionStateRejected = "rejected"
	workflowRAGPromotionStateCanceled = "canceled"

	workflowRAGPromotionDecisionApprove = "approve"
	workflowRAGPromotionDecisionReject  = "reject"
	workflowRAGPromotionDecisionDefer   = "defer"
	workflowRAGPromotionDecisionCancel  = "cancel"

	workflowRAGPromotionDefaultListLimit  = 50
	workflowRAGPromotionMaximumListLimit  = 200
	workflowRAGPromotionMaximumCandidates = 1024
)

const (
	WorkflowRAGPromotionFailureScopeDenied           = "workflow_rag_promotion_scope_denied"
	WorkflowRAGPromotionFailurePayloadInvalid        = "workflow_rag_promotion_payload_invalid"
	WorkflowRAGPromotionFailureSecretForbidden       = "workflow_rag_promotion_secret_material_forbidden"
	WorkflowRAGPromotionFailureNotFound              = "workflow_rag_promotion_not_found"
	WorkflowRAGPromotionFailureDatasetChanged        = "workflow_rag_promotion_dataset_changed"
	WorkflowRAGPromotionFailureDatasetArchived       = "workflow_rag_promotion_dataset_archived"
	WorkflowRAGPromotionFailureReviewInvalid         = "workflow_rag_promotion_review_invalid"
	WorkflowRAGPromotionFailureReviewNotEligible     = "workflow_rag_promotion_review_not_eligible"
	WorkflowRAGPromotionFailureSnapshotChanged       = "workflow_rag_promotion_snapshot_changed"
	WorkflowRAGPromotionFailureSnapshotArchived      = "workflow_rag_promotion_snapshot_archived"
	WorkflowRAGPromotionFailureProfileChanged        = "workflow_rag_promotion_profile_changed"
	WorkflowRAGPromotionFailureDraftChanged          = "workflow_rag_promotion_draft_changed"
	WorkflowRAGPromotionFailureDraftInvalid          = "workflow_rag_promotion_draft_invalid"
	WorkflowRAGPromotionFailureApplicationArchived   = "workflow_rag_promotion_application_archived"
	WorkflowRAGPromotionFailureRecordConflict        = "workflow_rag_promotion_record_version_conflict"
	WorkflowRAGPromotionFailureTransitionInvalid     = "workflow_rag_promotion_transition_invalid"
	WorkflowRAGPromotionFailureBindingNotEligible    = "workflow_rag_binding_not_eligible"
	WorkflowRAGPromotionFailureStoreUnavailable      = "workflow_rag_promotion_store_unavailable"
	WorkflowRAGPromotionFailureStoreContractMismatch = "workflow_rag_promotion_store_contract_mismatch"
	WorkflowRAGPromotionFailureWriteDisabled         = "workflow_rag_promotion_write_disabled"
	workflowRAGPromotionBlockerNotApproved           = "workflow_rag_promotion_not_approved"
)

type WorkflowRAGPromotionContext struct {
	RequestContext  context.Context
	RequestID       string
	TenantRef       string
	WorkspaceID     string
	ApplicationID   string
	ActorRef        string
	OwnerSubjectRef string
	AuditRef        string
	WriteEnabled    bool
}

type WorkflowRAGPromotionSourceDraftBinding struct {
	DraftID                  string `json:"draft_id"`
	DraftVersion             int    `json:"draft_version"`
	DraftDigest              string `json:"draft_digest"`
	BaseApplicationUpdatedAt string `json:"base_application_updated_at"`
}

type WorkflowRAGPromotionEvidenceBinding struct {
	Dataset           WorkflowRAGQualityDatasetBinding       `json:"dataset"`
	CandidateReviewID string                                 `json:"candidate_review_id"`
	BaselineSnapshot  WorkflowRAGEvaluationSnapshotBinding   `json:"baseline_snapshot"`
	CandidateSnapshot WorkflowRAGEvaluationSnapshotBinding   `json:"candidate_snapshot"`
	Profile           WorkflowRAGEvaluationProfileBinding    `json:"profile"`
	SourceDraft       WorkflowRAGPromotionSourceDraftBinding `json:"source_draft"`
}

type WorkflowRAGApplicationBindingRef struct {
	BindingID      string `json:"binding_id"`
	BindingVersion int    `json:"binding_version"`
	BindingDigest  string `json:"binding_digest"`
}

type WorkflowRAGKnowledgePromotionCandidate struct {
	SchemaVersion     string                              `json:"schema_version"`
	CandidateID       string                              `json:"candidate_id"`
	CandidateDigest   string                              `json:"candidate_digest"`
	TenantRef         string                              `json:"tenant_ref"`
	WorkspaceID       string                              `json:"workspace_id"`
	ApplicationID     string                              `json:"application_id"`
	OwnerSubjectRef   string                              `json:"owner_subject_ref"`
	Evidence          WorkflowRAGPromotionEvidenceBinding `json:"evidence"`
	CandidateState    string                              `json:"candidate_state"`
	RecordVersion     int                                 `json:"record_version"`
	BindingRef        *WorkflowRAGApplicationBindingRef   `json:"binding_ref"`
	CreatedAt         string                              `json:"created_at"`
	UpdatedAt         string                              `json:"updated_at"`
	CreatedByActorRef string                              `json:"created_by_actor_ref"`
	UpdatedByActorRef string                              `json:"updated_by_actor_ref"`
	RequestID         string                              `json:"request_id"`
	AuditRef          string                              `json:"audit_ref"`
}

type WorkflowRAGKnowledgePromotionDecision struct {
	SchemaVersion       string `json:"schema_version"`
	DecisionID          string `json:"decision_id"`
	CandidateID         string `json:"candidate_id"`
	CandidateDigest     string `json:"candidate_digest"`
	Decision            string `json:"decision"`
	Reason              string `json:"reason"`
	FromState           string `json:"from_state"`
	ToState             string `json:"to_state"`
	BeforeRecordVersion int    `json:"before_record_version"`
	AfterRecordVersion  int    `json:"after_record_version"`
	ActorRef            string `json:"actor_ref"`
	OccurredAt          string `json:"occurred_at"`
	RequestID           string `json:"request_id"`
	AuditRef            string `json:"audit_ref"`
}

type WorkflowRAGApplicationBinding struct {
	SchemaVersion string `json:"schema_version"`
	WorkflowRAGApplicationBindingRef
	CandidateID           string                              `json:"candidate_id"`
	CandidateDigest       string                              `json:"candidate_digest"`
	ApprovedDecisionID    string                              `json:"approved_decision_id"`
	ApprovedRecordVersion int                                 `json:"approved_record_version"`
	TenantRef             string                              `json:"tenant_ref"`
	WorkspaceID           string                              `json:"workspace_id"`
	ApplicationID         string                              `json:"application_id"`
	OwnerSubjectRef       string                              `json:"owner_subject_ref"`
	Evidence              WorkflowRAGPromotionEvidenceBinding `json:"evidence"`
	IssuedAt              string                              `json:"issued_at"`
	IssuedByActorRef      string                              `json:"issued_by_actor_ref"`
	RequestID             string                              `json:"request_id"`
	AuditRef              string                              `json:"audit_ref"`
}

type WorkflowRAGPromotionAudit struct {
	SchemaVersion   string                            `json:"schema_version"`
	EventID         string                            `json:"event_id"`
	EventKind       string                            `json:"event_kind"`
	CandidateID     string                            `json:"candidate_id"`
	CandidateDigest string                            `json:"candidate_digest"`
	CandidateState  string                            `json:"candidate_state"`
	RecordVersion   int                               `json:"record_version"`
	BindingRef      *WorkflowRAGApplicationBindingRef `json:"binding_ref"`
	TenantRef       string                            `json:"tenant_ref"`
	WorkspaceID     string                            `json:"workspace_id"`
	ApplicationID   string                            `json:"application_id"`
	OwnerSubjectRef string                            `json:"owner_subject_ref"`
	ActorRef        string                            `json:"actor_ref"`
	OccurredAt      string                            `json:"occurred_at"`
	RequestID       string                            `json:"request_id"`
	AuditRef        string                            `json:"audit_ref"`
}

type WorkflowRAGPromotionBlocker struct {
	Code string `json:"code"`
}

type WorkflowRAGPromotionEligibility struct {
	Eligible bool                          `json:"eligible"`
	Status   string                        `json:"status"`
	Blockers []WorkflowRAGPromotionBlocker `json:"blockers"`
}

type WorkflowRAGPromotionCandidateSummary struct {
	CandidateID       string                                 `json:"candidate_id"`
	Dataset           WorkflowRAGQualityDatasetBinding       `json:"dataset"`
	CandidateReviewID string                                 `json:"candidate_review_id"`
	SourceDraft       WorkflowRAGPromotionSourceDraftBinding `json:"source_draft"`
	CandidateState    string                                 `json:"candidate_state"`
	RecordVersion     int                                    `json:"record_version"`
	BindingRef        *WorkflowRAGApplicationBindingRef      `json:"binding_ref"`
	EligibilityStatus string                                 `json:"eligibility_status"`
	BlockerCount      int                                    `json:"blocker_count"`
	CreatedAt         string                                 `json:"created_at"`
	UpdatedAt         string                                 `json:"updated_at"`
}

type WorkflowRAGPromotionCreateInput struct {
	DatasetID            string
	DatasetVersion       int
	DatasetDigest        string
	CandidateReviewID    string
	DraftID              string
	ExpectedDraftVersion int
}

type WorkflowRAGPromotionDecisionInput struct {
	ExpectedRecordVersion int
	Decision              string
	Reason                string
}

type WorkflowRAGPromotionListInput struct {
	Limit  int
	Cursor string
}

type WorkflowRAGPromotionResult struct {
	Candidate            *WorkflowRAGKnowledgePromotionCandidate
	Decisions            []WorkflowRAGKnowledgePromotionDecision
	Binding              *WorkflowRAGApplicationBinding
	Eligibility          WorkflowRAGPromotionEligibility
	FailureCode          string
	CurrentRecordVersion int
	CurrentState         string
}

type WorkflowRAGPromotionListResult struct {
	Items       []WorkflowRAGPromotionCandidateSummary
	NextCursor  *string
	FailureCode string
}

type workflowRAGPromotionListQuery struct {
	Limit             int
	BeforeCreatedAt   string
	BeforeCandidateID string
}

type workflowRAGPromotionRepository interface {
	Create(WorkflowRAGPromotionContext, WorkflowRAGKnowledgePromotionCandidate, WorkflowRAGPromotionAudit) error
	Read(WorkflowRAGPromotionContext, string) (WorkflowRAGKnowledgePromotionCandidate, []WorkflowRAGKnowledgePromotionDecision, *WorkflowRAGApplicationBinding, []WorkflowRAGPromotionAudit, error)
	List(WorkflowRAGPromotionContext, workflowRAGPromotionListQuery) ([]WorkflowRAGKnowledgePromotionCandidate, error)
	AppendDecision(WorkflowRAGPromotionContext, string, int, WorkflowRAGKnowledgePromotionCandidate, WorkflowRAGKnowledgePromotionDecision, *WorkflowRAGApplicationBinding, []WorkflowRAGPromotionAudit) error
}

type workflowRAGPromotionApplicationReader func(WorkflowRAGPromotionContext) (ApplicationSummary, error)

type workflowRAGPromotionService struct {
	repository           workflowRAGPromotionRepository
	evaluationRepository workflowRAGEvaluationDatasetRepository
	snapshotRepository   workflowRAGSnapshotRepository
	draftRepository      applicationConfigurationDraftRepository
	readApplication      workflowRAGPromotionApplicationReader
	currentProfile       func() WorkflowRAGExecutionProfile
	now                  func() time.Time
	newID                func(string) (string, error)
}

func newWorkflowRAGPromotionService(
	repository workflowRAGPromotionRepository,
	evaluations workflowRAGEvaluationDatasetRepository,
	snapshots workflowRAGSnapshotRepository,
	drafts applicationConfigurationDraftRepository,
	readApplication workflowRAGPromotionApplicationReader,
) workflowRAGPromotionService {
	return workflowRAGPromotionService{
		repository: repository, evaluationRepository: evaluations, snapshotRepository: snapshots,
		draftRepository: drafts, readApplication: readApplication, currentProfile: workflowRAGLexicalProfile,
		now: func() time.Time { return time.Now().UTC() }, newID: newWorkflowRAGStableID,
	}
}

func (service workflowRAGPromotionService) Create(ctx WorkflowRAGPromotionContext, input WorkflowRAGPromotionCreateInput) WorkflowRAGPromotionResult {
	if validateWorkflowRAGPromotionContext(ctx) != nil {
		return workflowRAGPromotionFailure(WorkflowRAGPromotionFailureScopeDenied)
	}
	if !ctx.WriteEnabled {
		return workflowRAGPromotionFailure(WorkflowRAGPromotionFailureWriteDisabled)
	}
	input = normalizeWorkflowRAGPromotionCreateInput(input)
	if !workflowRAGDatasetIDPattern.MatchString(input.DatasetID) || input.DatasetVersion < 1 || !workflowRAGDigestPattern.MatchString(input.DatasetDigest) ||
		!workflowRAGEvaluationReviewIDPattern.MatchString(input.CandidateReviewID) || !applicationDraftIdentifierPattern.MatchString(input.DraftID) || input.ExpectedDraftVersion < 1 {
		return workflowRAGPromotionFailure(WorkflowRAGPromotionFailurePayloadInvalid)
	}
	if workflowRAGContainsForbiddenMaterial(strings.Join([]string{input.DatasetID, input.DatasetDigest, input.CandidateReviewID, input.DraftID}, "\n")) {
		return workflowRAGPromotionFailure(WorkflowRAGPromotionFailureSecretForbidden)
	}
	evidence, failure := service.loadPromotionEvidence(ctx, input)
	if failure != "" {
		return workflowRAGPromotionFailure(failure)
	}
	candidateID, err := service.newID("wragp_")
	if err != nil || !workflowRAGPromotionCandidateIDPattern.MatchString(candidateID) {
		return workflowRAGPromotionFailure(WorkflowRAGPromotionFailureStoreUnavailable)
	}
	eventID, err := service.newID("wragpa_")
	if err != nil || !workflowRAGPromotionAuditIDPattern.MatchString(eventID) {
		return workflowRAGPromotionFailure(WorkflowRAGPromotionFailureStoreUnavailable)
	}
	at := service.now().UTC().Format(time.RFC3339Nano)
	candidate := WorkflowRAGKnowledgePromotionCandidate{
		SchemaVersion: workflowRAGPromotionCandidateSchemaVersion, CandidateID: candidateID,
		TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, OwnerSubjectRef: ctx.OwnerSubjectRef,
		Evidence: evidence, CandidateState: workflowRAGPromotionStatePending, RecordVersion: 1,
		CreatedAt: at, UpdatedAt: at, CreatedByActorRef: ctx.ActorRef, UpdatedByActorRef: ctx.ActorRef,
		RequestID: ctx.RequestID, AuditRef: ctx.AuditRef,
	}
	candidate.CandidateDigest, err = workflowRAGPromotionCandidateDigest(candidate)
	if err != nil {
		return workflowRAGPromotionFailure(WorkflowRAGPromotionFailureStoreContractMismatch)
	}
	audit := workflowRAGPromotionAuditFromCandidate(ctx, eventID, "promotion_candidate_created", candidate, at)
	if service.repository == nil {
		return workflowRAGPromotionFailure(WorkflowRAGPromotionFailureStoreUnavailable)
	}
	if err = service.repository.Create(ctx, candidate, audit); err != nil {
		return workflowRAGPromotionRepositoryFailure(err)
	}
	return service.resultWithEligibility(ctx, candidate, []WorkflowRAGKnowledgePromotionDecision{}, nil)
}

func (service workflowRAGPromotionService) Read(ctx WorkflowRAGPromotionContext, candidateID string) WorkflowRAGPromotionResult {
	if validateWorkflowRAGPromotionContext(ctx) != nil {
		return workflowRAGPromotionFailure(WorkflowRAGPromotionFailureScopeDenied)
	}
	candidateID = strings.TrimSpace(candidateID)
	if !workflowRAGPromotionCandidateIDPattern.MatchString(candidateID) {
		return workflowRAGPromotionFailure(WorkflowRAGPromotionFailurePayloadInvalid)
	}
	if service.repository == nil {
		return workflowRAGPromotionFailure(WorkflowRAGPromotionFailureStoreUnavailable)
	}
	candidate, decisions, binding, _, err := service.repository.Read(ctx, candidateID)
	if err != nil {
		return workflowRAGPromotionRepositoryFailure(err)
	}
	return service.resultWithEligibility(ctx, candidate, decisions, binding)
}

func (service workflowRAGPromotionService) List(ctx WorkflowRAGPromotionContext, input WorkflowRAGPromotionListInput) WorkflowRAGPromotionListResult {
	if validateWorkflowRAGPromotionContext(ctx) != nil {
		return WorkflowRAGPromotionListResult{Items: []WorkflowRAGPromotionCandidateSummary{}, FailureCode: WorkflowRAGPromotionFailureScopeDenied}
	}
	query, failure := parseWorkflowRAGPromotionListQuery(input)
	if failure != "" {
		return WorkflowRAGPromotionListResult{Items: []WorkflowRAGPromotionCandidateSummary{}, FailureCode: failure}
	}
	if service.repository == nil {
		return WorkflowRAGPromotionListResult{Items: []WorkflowRAGPromotionCandidateSummary{}, FailureCode: WorkflowRAGPromotionFailureStoreUnavailable}
	}
	query.Limit++
	candidates, err := service.repository.List(ctx, query)
	if err != nil {
		return WorkflowRAGPromotionListResult{Items: []WorkflowRAGPromotionCandidateSummary{}, FailureCode: workflowRAGPromotionRepositoryFailure(err).FailureCode}
	}
	result := WorkflowRAGPromotionListResult{Items: make([]WorkflowRAGPromotionCandidateSummary, 0, len(candidates))}
	if len(candidates) >= query.Limit {
		candidates = candidates[:query.Limit-1]
		last := candidates[len(candidates)-1]
		next := last.CreatedAt + "|" + last.CandidateID
		result.NextCursor = &next
	}
	for _, candidate := range candidates {
		_, _, binding, _, readErr := service.repository.Read(ctx, candidate.CandidateID)
		eligibility := WorkflowRAGPromotionEligibility{Status: "blocked", Blockers: []WorkflowRAGPromotionBlocker{{Code: WorkflowRAGPromotionFailureStoreUnavailable}}}
		if readErr == nil {
			eligibility = service.evaluateEligibility(ctx, candidate, binding)
		}
		result.Items = append(result.Items, workflowRAGPromotionSummary(candidate, eligibility))
	}
	return result
}

func (service workflowRAGPromotionService) Decide(ctx WorkflowRAGPromotionContext, candidateID string, input WorkflowRAGPromotionDecisionInput) WorkflowRAGPromotionResult {
	if validateWorkflowRAGPromotionContext(ctx) != nil {
		return workflowRAGPromotionFailure(WorkflowRAGPromotionFailureScopeDenied)
	}
	if !ctx.WriteEnabled {
		return workflowRAGPromotionFailure(WorkflowRAGPromotionFailureWriteDisabled)
	}
	candidateID = strings.TrimSpace(candidateID)
	input.Decision, input.Reason = strings.TrimSpace(input.Decision), strings.TrimSpace(input.Reason)
	if !workflowRAGPromotionCandidateIDPattern.MatchString(candidateID) || input.ExpectedRecordVersion < 1 ||
		!workflowRAGPromotionDecisionAllowed(input.Decision) || !utf8.ValidString(input.Reason) || utf8.RuneCountInString(input.Reason) < 4 || utf8.RuneCountInString(input.Reason) > 500 {
		return workflowRAGPromotionFailure(WorkflowRAGPromotionFailurePayloadInvalid)
	}
	if workflowRAGContainsForbiddenMaterial(input.Reason) || applicationDraftStringContainsSecret(input.Reason) {
		return workflowRAGPromotionFailure(WorkflowRAGPromotionFailureSecretForbidden)
	}
	if service.repository == nil {
		return workflowRAGPromotionFailure(WorkflowRAGPromotionFailureStoreUnavailable)
	}
	candidate, decisions, existingBinding, _, err := service.repository.Read(ctx, candidateID)
	if err != nil {
		return workflowRAGPromotionRepositoryFailure(err)
	}
	if candidate.RecordVersion != input.ExpectedRecordVersion {
		return WorkflowRAGPromotionResult{FailureCode: WorkflowRAGPromotionFailureRecordConflict, CurrentRecordVersion: candidate.RecordVersion, CurrentState: candidate.CandidateState}
	}
	nextState, allowed := workflowRAGPromotionNextState(candidate.CandidateState, input.Decision)
	if !allowed {
		return WorkflowRAGPromotionResult{FailureCode: WorkflowRAGPromotionFailureTransitionInvalid, CurrentRecordVersion: candidate.RecordVersion, CurrentState: candidate.CandidateState}
	}
	if input.Decision == workflowRAGPromotionDecisionApprove {
		if failure := service.validatePromotionEvidence(ctx, candidate.Evidence); failure != "" {
			return WorkflowRAGPromotionResult{FailureCode: failure, CurrentRecordVersion: candidate.RecordVersion, CurrentState: candidate.CandidateState}
		}
	}
	decisionID, err := service.newID("wragpd_")
	if err != nil || !workflowRAGPromotionDecisionIDPattern.MatchString(decisionID) {
		return workflowRAGPromotionFailure(WorkflowRAGPromotionFailureStoreUnavailable)
	}
	at := service.now().UTC().Format(time.RFC3339Nano)
	decision := WorkflowRAGKnowledgePromotionDecision{
		SchemaVersion: workflowRAGPromotionDecisionSchemaVersion, DecisionID: decisionID,
		CandidateID: candidate.CandidateID, CandidateDigest: candidate.CandidateDigest, Decision: input.Decision, Reason: input.Reason,
		FromState: candidate.CandidateState, ToState: nextState, BeforeRecordVersion: candidate.RecordVersion, AfterRecordVersion: candidate.RecordVersion + 1,
		ActorRef: ctx.ActorRef, OccurredAt: at, RequestID: ctx.RequestID, AuditRef: ctx.AuditRef,
	}
	updated := candidate
	updated.CandidateState, updated.RecordVersion = nextState, decision.AfterRecordVersion
	updated.UpdatedAt, updated.UpdatedByActorRef = at, ctx.ActorRef
	audits := make([]WorkflowRAGPromotionAudit, 0, 2)
	decisionAuditID, idErr := service.newID("wragpa_")
	if idErr != nil || !workflowRAGPromotionAuditIDPattern.MatchString(decisionAuditID) {
		return workflowRAGPromotionFailure(WorkflowRAGPromotionFailureStoreUnavailable)
	}
	var binding *WorkflowRAGApplicationBinding
	if input.Decision == workflowRAGPromotionDecisionApprove {
		bindingID, bindingErr := service.newID("wragb_")
		if bindingErr != nil || !workflowRAGPromotionBindingIDPattern.MatchString(bindingID) {
			return workflowRAGPromotionFailure(WorkflowRAGPromotionFailureStoreUnavailable)
		}
		created := workflowRAGApplicationBindingFromApproval(ctx, bindingID, candidate, decision, at)
		created.BindingDigest, bindingErr = workflowRAGApplicationBindingDigest(created)
		if bindingErr != nil {
			return workflowRAGPromotionFailure(WorkflowRAGPromotionFailureStoreContractMismatch)
		}
		binding = &created
		ref := created.WorkflowRAGApplicationBindingRef
		updated.BindingRef = &ref
	}
	audits = append(audits, workflowRAGPromotionAuditFromCandidate(ctx, decisionAuditID, "promotion_decision_"+input.Decision, updated, at))
	if binding != nil {
		bindingAuditID, auditErr := service.newID("wragpa_")
		if auditErr != nil || !workflowRAGPromotionAuditIDPattern.MatchString(bindingAuditID) {
			return workflowRAGPromotionFailure(WorkflowRAGPromotionFailureStoreUnavailable)
		}
		audits = append(audits, workflowRAGPromotionAuditFromCandidate(ctx, bindingAuditID, "promotion_binding_issued", updated, at))
	}
	if err = service.repository.AppendDecision(ctx, candidateID, input.ExpectedRecordVersion, updated, decision, binding, audits); err != nil {
		return workflowRAGPromotionRepositoryFailure(err)
	}
	decisions = append(decisions, decision)
	if binding == nil {
		binding = existingBinding
	}
	return service.resultWithEligibility(ctx, updated, decisions, binding)
}

func (service workflowRAGPromotionService) loadPromotionEvidence(ctx WorkflowRAGPromotionContext, input WorkflowRAGPromotionCreateInput) (WorkflowRAGPromotionEvidenceBinding, string) {
	application, failure := service.readPromotionApplication(ctx)
	if failure != "" {
		return WorkflowRAGPromotionEvidenceBinding{}, failure
	}
	draftContext := workflowRAGPromotionDraftContext(ctx)
	if service.draftRepository == nil {
		return WorkflowRAGPromotionEvidenceBinding{}, WorkflowRAGPromotionFailureStoreUnavailable
	}
	draft, err := service.draftRepository.Read(draftContext, input.DraftID)
	if errors.Is(err, errApplicationDraftNotFound) {
		return WorkflowRAGPromotionEvidenceBinding{}, WorkflowRAGPromotionFailureDraftChanged
	}
	if err != nil {
		return WorkflowRAGPromotionEvidenceBinding{}, WorkflowRAGPromotionFailureStoreUnavailable
	}
	if draft.DraftVersion != input.ExpectedDraftVersion {
		return WorkflowRAGPromotionEvidenceBinding{}, WorkflowRAGPromotionFailureDraftChanged
	}
	validation := validateApplicationConfigurationDraftPayload(draftContext, draft.ApplicationConfigurationDraftPayload)
	if draft.SchemaVersion != applicationConfigurationDraftSchemaVersion || !draft.ValidationSummary.IsValid || draft.ValidationSummary.State != applicationDraftValidationValid || !validation.IsValid {
		return WorkflowRAGPromotionEvidenceBinding{}, WorkflowRAGPromotionFailureDraftInvalid
	}
	if strings.TrimSpace(draft.BaseApplicationUpdatedAt) != strings.TrimSpace(application.UpdatedAt) {
		return WorkflowRAGPromotionEvidenceBinding{}, WorkflowRAGPromotionFailureDraftChanged
	}
	draftDigest, err := applicationPublishSnapshotDigest(applicationPublishSnapshotFromDraft(draft))
	if err != nil || !workflowRAGDigestPattern.MatchString(draftDigest) {
		return WorkflowRAGPromotionEvidenceBinding{}, WorkflowRAGPromotionFailureStoreContractMismatch
	}
	if service.evaluationRepository == nil {
		return WorkflowRAGPromotionEvidenceBinding{}, WorkflowRAGPromotionFailureStoreUnavailable
	}
	snapshotCtx := workflowRAGPromotionSnapshotContext(ctx)
	resource, version, err := service.evaluationRepository.ReadVersion(snapshotCtx, input.DatasetID, input.DatasetVersion)
	if failure = workflowRAGPromotionDatasetReadFailure(err); failure != "" {
		return WorkflowRAGPromotionEvidenceBinding{}, failure
	}
	if resource.LifecycleState != workflowRAGEvaluationActive {
		return WorkflowRAGPromotionEvidenceBinding{}, WorkflowRAGPromotionFailureDatasetArchived
	}
	if resource.LatestVersion != input.DatasetVersion || resource.LatestDigest != input.DatasetDigest || version.Dataset.DatasetID != input.DatasetID ||
		version.Dataset.DatasetVersion != input.DatasetVersion || version.Dataset.DatasetDigest != input.DatasetDigest {
		return WorkflowRAGPromotionEvidenceBinding{}, WorkflowRAGPromotionFailureDatasetChanged
	}
	review, err := service.evaluationRepository.ReadReview(snapshotCtx, input.DatasetID, input.CandidateReviewID)
	if err != nil {
		if errors.Is(err, errWorkflowRAGEvaluationScopeDenied) {
			return WorkflowRAGPromotionEvidenceBinding{}, WorkflowRAGPromotionFailureScopeDenied
		}
		if errors.Is(err, errWorkflowRAGEvaluationStore) {
			return WorkflowRAGPromotionEvidenceBinding{}, WorkflowRAGPromotionFailureStoreUnavailable
		}
		return WorkflowRAGPromotionEvidenceBinding{}, WorkflowRAGPromotionFailureReviewInvalid
	}
	reviewContext := snapshotCtx
	reviewContext.ActorRef, reviewContext.RequestID, reviewContext.AuditRef = review.CreatedByActorRef, review.RequestID, review.AuditRef
	if validateStoredWorkflowRAGCandidateReview(review, reviewContext) != nil || review.Dataset != version.Dataset.DatasetBinding() || review.BaselineSnapshot != version.Dataset.Snapshot || review.Profile != version.Dataset.Profile {
		return WorkflowRAGPromotionEvidenceBinding{}, WorkflowRAGPromotionFailureReviewInvalid
	}
	if review.Candidate.Status != "passed" || (review.Conclusion != "improved" && review.Conclusion != "unchanged") {
		return WorkflowRAGPromotionEvidenceBinding{}, WorkflowRAGPromotionFailureReviewNotEligible
	}
	if failure = service.validatePromotionSnapshot(snapshotCtx, review.BaselineSnapshot); failure != "" {
		return WorkflowRAGPromotionEvidenceBinding{}, failure
	}
	if failure = service.validatePromotionSnapshot(snapshotCtx, review.CandidateSnapshot); failure != "" {
		return WorkflowRAGPromotionEvidenceBinding{}, failure
	}
	if service.currentProfile == nil {
		return WorkflowRAGPromotionEvidenceBinding{}, WorkflowRAGPromotionFailureStoreUnavailable
	}
	profile := service.currentProfile()
	currentProfile := WorkflowRAGEvaluationProfileBinding{ProfileID: profile.ProfileID, ProfileVersion: profile.ProfileVersion, ProfileDigest: profile.ProfileDigest}
	if currentProfile != review.Profile {
		return WorkflowRAGPromotionEvidenceBinding{}, WorkflowRAGPromotionFailureProfileChanged
	}
	evidence := WorkflowRAGPromotionEvidenceBinding{
		Dataset: review.Dataset, CandidateReviewID: review.ReviewID, BaselineSnapshot: review.BaselineSnapshot,
		CandidateSnapshot: review.CandidateSnapshot, Profile: review.Profile,
		SourceDraft: WorkflowRAGPromotionSourceDraftBinding{DraftID: draft.DraftID, DraftVersion: draft.DraftVersion, DraftDigest: draftDigest, BaseApplicationUpdatedAt: draft.BaseApplicationUpdatedAt},
	}
	if workflowRAGPromotionEvidenceContainsForbiddenMaterial(evidence) {
		return WorkflowRAGPromotionEvidenceBinding{}, WorkflowRAGPromotionFailureSecretForbidden
	}
	return evidence, ""
}

func (service workflowRAGPromotionService) validatePromotionEvidence(ctx WorkflowRAGPromotionContext, expected WorkflowRAGPromotionEvidenceBinding) string {
	actual, failure := service.loadPromotionEvidence(ctx, WorkflowRAGPromotionCreateInput{
		DatasetID: expected.Dataset.DatasetID, DatasetVersion: expected.Dataset.DatasetVersion, DatasetDigest: expected.Dataset.DatasetDigest,
		CandidateReviewID: expected.CandidateReviewID, DraftID: expected.SourceDraft.DraftID, ExpectedDraftVersion: expected.SourceDraft.DraftVersion,
	})
	if failure != "" {
		return failure
	}
	if actual.Dataset != expected.Dataset {
		return WorkflowRAGPromotionFailureDatasetChanged
	}
	if actual.CandidateReviewID != expected.CandidateReviewID || actual.BaselineSnapshot != expected.BaselineSnapshot || actual.CandidateSnapshot != expected.CandidateSnapshot {
		return WorkflowRAGPromotionFailureReviewInvalid
	}
	if actual.Profile != expected.Profile {
		return WorkflowRAGPromotionFailureProfileChanged
	}
	if actual.SourceDraft != expected.SourceDraft {
		return WorkflowRAGPromotionFailureDraftChanged
	}
	return ""
}

func (service workflowRAGPromotionService) validatePromotionSnapshot(ctx WorkflowRAGSnapshotContext, binding WorkflowRAGEvaluationSnapshotBinding) string {
	if service.snapshotRepository == nil {
		return WorkflowRAGPromotionFailureStoreUnavailable
	}
	resource, record, err := service.snapshotRepository.ReadVersion(ctx, binding.SnapshotID, binding.SnapshotVersion)
	if errors.Is(err, errWorkflowRAGScopeDenied) {
		return WorkflowRAGPromotionFailureScopeDenied
	}
	if errors.Is(err, errWorkflowRAGNotFound) {
		return WorkflowRAGPromotionFailureSnapshotChanged
	}
	if err != nil {
		return WorkflowRAGPromotionFailureStoreUnavailable
	}
	if resource.LifecycleState != workflowRAGSnapshotActive {
		return WorkflowRAGPromotionFailureSnapshotArchived
	}
	if resource.LatestVersion != binding.SnapshotVersion || resource.LatestDigest != binding.SnapshotDigest || resource.LatestRAGRef != binding.RAGRef ||
		workflowRAGEvaluationSnapshotBinding(record) != binding {
		return WorkflowRAGPromotionFailureSnapshotChanged
	}
	return ""
}

func (service workflowRAGPromotionService) readPromotionApplication(ctx WorkflowRAGPromotionContext) (ApplicationSummary, string) {
	if service.readApplication == nil {
		return ApplicationSummary{}, WorkflowRAGPromotionFailureStoreUnavailable
	}
	application, err := service.readApplication(ctx)
	if errors.Is(err, errApplicationCatalogArchived) || errors.Is(err, errApplicationPublishBaselineNotFound) {
		return ApplicationSummary{}, WorkflowRAGPromotionFailureApplicationArchived
	}
	if err != nil {
		return ApplicationSummary{}, WorkflowRAGPromotionFailureStoreUnavailable
	}
	if application.ApplicationRef != ctx.ApplicationID || application.TenantRef != ctx.TenantRef || application.OwnerSubjectRef != ctx.OwnerSubjectRef {
		return ApplicationSummary{}, WorkflowRAGPromotionFailureScopeDenied
	}
	if strings.TrimSpace(application.UpdatedAt) == "" {
		return ApplicationSummary{}, WorkflowRAGPromotionFailureStoreContractMismatch
	}
	return application, ""
}

func (service workflowRAGPromotionService) resultWithEligibility(ctx WorkflowRAGPromotionContext, candidate WorkflowRAGKnowledgePromotionCandidate, decisions []WorkflowRAGKnowledgePromotionDecision, binding *WorkflowRAGApplicationBinding) WorkflowRAGPromotionResult {
	return WorkflowRAGPromotionResult{
		Candidate: &candidate, Decisions: cloneWorkflowRAGPromotionDecisions(decisions), Binding: cloneWorkflowRAGApplicationBinding(binding),
		Eligibility: service.evaluateEligibility(ctx, candidate, binding), CurrentRecordVersion: candidate.RecordVersion, CurrentState: candidate.CandidateState,
	}
}

func (service workflowRAGPromotionService) evaluateEligibility(ctx WorkflowRAGPromotionContext, candidate WorkflowRAGKnowledgePromotionCandidate, binding *WorkflowRAGApplicationBinding) WorkflowRAGPromotionEligibility {
	blockers := make([]WorkflowRAGPromotionBlocker, 0, 2)
	if candidate.CandidateState != workflowRAGPromotionStateApproved {
		blockers = append(blockers, WorkflowRAGPromotionBlocker{Code: workflowRAGPromotionBlockerNotApproved})
	}
	if candidate.CandidateState == workflowRAGPromotionStatePending || candidate.CandidateState == workflowRAGPromotionStateDeferred || candidate.CandidateState == workflowRAGPromotionStateApproved {
		if failure := service.validatePromotionEvidence(ctx, candidate.Evidence); failure != "" {
			blockers = append(blockers, WorkflowRAGPromotionBlocker{Code: failure})
		}
	}
	if candidate.CandidateState == workflowRAGPromotionStateApproved {
		if binding == nil || candidate.BindingRef == nil || *candidate.BindingRef != binding.WorkflowRAGApplicationBindingRef || validateStoredWorkflowRAGApplicationBinding(*binding) != nil {
			blockers = append(blockers, WorkflowRAGPromotionBlocker{Code: WorkflowRAGPromotionFailureBindingNotEligible})
		}
	}
	status := "blocked"
	if len(blockers) == 0 {
		status = "eligible"
	}
	return WorkflowRAGPromotionEligibility{Eligible: len(blockers) == 0, Status: status, Blockers: blockers}
}

func normalizeWorkflowRAGPromotionCreateInput(input WorkflowRAGPromotionCreateInput) WorkflowRAGPromotionCreateInput {
	input.DatasetID = strings.TrimSpace(input.DatasetID)
	input.DatasetDigest = strings.TrimSpace(input.DatasetDigest)
	input.CandidateReviewID = strings.TrimSpace(input.CandidateReviewID)
	input.DraftID = strings.TrimSpace(input.DraftID)
	return input
}

func workflowRAGPromotionCandidateDigest(candidate WorkflowRAGKnowledgePromotionCandidate) (string, error) {
	document := struct {
		SchemaVersion   string                              `json:"schema_version"`
		CandidateID     string                              `json:"candidate_id"`
		TenantRef       string                              `json:"tenant_ref"`
		WorkspaceID     string                              `json:"workspace_id"`
		ApplicationID   string                              `json:"application_id"`
		OwnerSubjectRef string                              `json:"owner_subject_ref"`
		Evidence        WorkflowRAGPromotionEvidenceBinding `json:"evidence"`
		CreatedAt       string                              `json:"created_at"`
		CreatedBy       string                              `json:"created_by_actor_ref"`
		RequestID       string                              `json:"request_id"`
		AuditRef        string                              `json:"audit_ref"`
	}{
		candidate.SchemaVersion, candidate.CandidateID, candidate.TenantRef, candidate.WorkspaceID, candidate.ApplicationID,
		candidate.OwnerSubjectRef, candidate.Evidence, candidate.CreatedAt, candidate.CreatedByActorRef, candidate.RequestID, candidate.AuditRef,
	}
	payload, err := json.Marshal(document)
	if err != nil {
		return "", err
	}
	return workflowRAGSHA256(string(payload)), nil
}

func workflowRAGApplicationBindingDigest(binding WorkflowRAGApplicationBinding) (string, error) {
	copy := binding
	copy.BindingDigest = ""
	payload, err := json.Marshal(copy)
	if err != nil {
		return "", err
	}
	return workflowRAGSHA256(string(payload)), nil
}

func workflowRAGApplicationBindingFromApproval(ctx WorkflowRAGPromotionContext, bindingID string, candidate WorkflowRAGKnowledgePromotionCandidate, decision WorkflowRAGKnowledgePromotionDecision, at string) WorkflowRAGApplicationBinding {
	return WorkflowRAGApplicationBinding{
		SchemaVersion:                    workflowRAGApplicationBindingSchemaVersion,
		WorkflowRAGApplicationBindingRef: WorkflowRAGApplicationBindingRef{BindingID: bindingID, BindingVersion: 1},
		CandidateID:                      candidate.CandidateID, CandidateDigest: candidate.CandidateDigest,
		ApprovedDecisionID: decision.DecisionID, ApprovedRecordVersion: decision.AfterRecordVersion,
		TenantRef: candidate.TenantRef, WorkspaceID: candidate.WorkspaceID, ApplicationID: candidate.ApplicationID, OwnerSubjectRef: candidate.OwnerSubjectRef,
		Evidence: candidate.Evidence, IssuedAt: at, IssuedByActorRef: ctx.ActorRef, RequestID: ctx.RequestID, AuditRef: ctx.AuditRef,
	}
}

func workflowRAGPromotionAuditFromCandidate(ctx WorkflowRAGPromotionContext, eventID, kind string, candidate WorkflowRAGKnowledgePromotionCandidate, at string) WorkflowRAGPromotionAudit {
	return WorkflowRAGPromotionAudit{
		SchemaVersion: workflowRAGPromotionAuditSchemaVersion, EventID: eventID, EventKind: kind,
		CandidateID: candidate.CandidateID, CandidateDigest: candidate.CandidateDigest, CandidateState: candidate.CandidateState,
		RecordVersion: candidate.RecordVersion, BindingRef: cloneWorkflowRAGApplicationBindingRef(candidate.BindingRef),
		TenantRef: candidate.TenantRef, WorkspaceID: candidate.WorkspaceID, ApplicationID: candidate.ApplicationID, OwnerSubjectRef: candidate.OwnerSubjectRef,
		ActorRef: ctx.ActorRef, OccurredAt: at, RequestID: ctx.RequestID, AuditRef: ctx.AuditRef,
	}
}

func workflowRAGPromotionNextState(state, decision string) (string, bool) {
	switch state {
	case workflowRAGPromotionStatePending:
		switch decision {
		case workflowRAGPromotionDecisionApprove:
			return workflowRAGPromotionStateApproved, true
		case workflowRAGPromotionDecisionReject:
			return workflowRAGPromotionStateRejected, true
		case workflowRAGPromotionDecisionDefer:
			return workflowRAGPromotionStateDeferred, true
		case workflowRAGPromotionDecisionCancel:
			return workflowRAGPromotionStateCanceled, true
		}
	case workflowRAGPromotionStateDeferred:
		switch decision {
		case workflowRAGPromotionDecisionApprove:
			return workflowRAGPromotionStateApproved, true
		case workflowRAGPromotionDecisionReject:
			return workflowRAGPromotionStateRejected, true
		case workflowRAGPromotionDecisionCancel:
			return workflowRAGPromotionStateCanceled, true
		}
	case workflowRAGPromotionStateApproved:
		if decision == workflowRAGPromotionDecisionCancel {
			return workflowRAGPromotionStateCanceled, true
		}
	}
	return "", false
}

func workflowRAGPromotionDecisionAllowed(decision string) bool {
	return decision == workflowRAGPromotionDecisionApprove || decision == workflowRAGPromotionDecisionReject || decision == workflowRAGPromotionDecisionDefer || decision == workflowRAGPromotionDecisionCancel
}

func parseWorkflowRAGPromotionListQuery(input WorkflowRAGPromotionListInput) (workflowRAGPromotionListQuery, string) {
	limit := input.Limit
	if limit == 0 {
		limit = workflowRAGPromotionDefaultListLimit
	}
	if limit < 1 || limit > workflowRAGPromotionMaximumListLimit {
		return workflowRAGPromotionListQuery{}, WorkflowRAGPromotionFailurePayloadInvalid
	}
	query := workflowRAGPromotionListQuery{Limit: limit}
	cursor := strings.TrimSpace(input.Cursor)
	if cursor == "" {
		return query, ""
	}
	parts := strings.Split(cursor, "|")
	if len(parts) != 2 || !workflowRAGPromotionCandidateIDPattern.MatchString(parts[1]) {
		return workflowRAGPromotionListQuery{}, WorkflowRAGPromotionFailurePayloadInvalid
	}
	if _, err := time.Parse(time.RFC3339Nano, parts[0]); err != nil {
		return workflowRAGPromotionListQuery{}, WorkflowRAGPromotionFailurePayloadInvalid
	}
	query.BeforeCreatedAt, query.BeforeCandidateID = parts[0], parts[1]
	return query, ""
}

func workflowRAGPromotionSummary(candidate WorkflowRAGKnowledgePromotionCandidate, eligibility WorkflowRAGPromotionEligibility) WorkflowRAGPromotionCandidateSummary {
	return WorkflowRAGPromotionCandidateSummary{
		CandidateID: candidate.CandidateID, Dataset: candidate.Evidence.Dataset, CandidateReviewID: candidate.Evidence.CandidateReviewID,
		SourceDraft: candidate.Evidence.SourceDraft, CandidateState: candidate.CandidateState, RecordVersion: candidate.RecordVersion,
		BindingRef: cloneWorkflowRAGApplicationBindingRef(candidate.BindingRef), EligibilityStatus: eligibility.Status, BlockerCount: len(eligibility.Blockers),
		CreatedAt: candidate.CreatedAt, UpdatedAt: candidate.UpdatedAt,
	}
}

func workflowRAGPromotionEvidenceContainsForbiddenMaterial(evidence WorkflowRAGPromotionEvidenceBinding) bool {
	return workflowRAGContainsForbiddenMaterial(strings.Join([]string{
		evidence.Dataset.DatasetID, evidence.Dataset.DatasetDigest, evidence.CandidateReviewID,
		evidence.BaselineSnapshot.SnapshotID, evidence.BaselineSnapshot.SnapshotDigest, evidence.BaselineSnapshot.RAGRef,
		evidence.CandidateSnapshot.SnapshotID, evidence.CandidateSnapshot.SnapshotDigest, evidence.CandidateSnapshot.RAGRef,
		evidence.Profile.ProfileID, evidence.Profile.ProfileDigest, evidence.SourceDraft.DraftID, evidence.SourceDraft.DraftDigest,
	}, "\n"))
}

func workflowRAGPromotionDatasetReadFailure(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, errWorkflowRAGEvaluationScopeDenied) {
		return WorkflowRAGPromotionFailureScopeDenied
	}
	if errors.Is(err, errWorkflowRAGEvaluationNotFound) {
		return WorkflowRAGPromotionFailureDatasetChanged
	}
	if errors.Is(err, errWorkflowRAGEvaluationArchived) {
		return WorkflowRAGPromotionFailureDatasetArchived
	}
	return WorkflowRAGPromotionFailureStoreUnavailable
}

func workflowRAGPromotionFailure(code string) WorkflowRAGPromotionResult {
	return WorkflowRAGPromotionResult{Decisions: []WorkflowRAGKnowledgePromotionDecision{}, Eligibility: WorkflowRAGPromotionEligibility{Status: "blocked", Blockers: []WorkflowRAGPromotionBlocker{}}, FailureCode: code}
}

func workflowRAGPromotionSnapshotContext(ctx WorkflowRAGPromotionContext) WorkflowRAGSnapshotContext {
	return WorkflowRAGSnapshotContext{RequestContext: ctx.RequestContext, RequestID: ctx.RequestID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, ActorRef: ctx.ActorRef, AuditRef: ctx.AuditRef}
}

func workflowRAGPromotionDraftContext(ctx WorkflowRAGPromotionContext) ApplicationConfigurationDraftContext {
	return ApplicationConfigurationDraftContext{RequestContext: ctx.RequestContext, RequestID: ctx.RequestID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, ActorRef: ctx.ActorRef, OwnerSubjectRef: ctx.OwnerSubjectRef, AuditRef: ctx.AuditRef}
}

func (dataset WorkflowRAGEvaluationDataset) DatasetBinding() WorkflowRAGQualityDatasetBinding {
	return WorkflowRAGQualityDatasetBinding{DatasetID: dataset.DatasetID, DatasetVersion: dataset.DatasetVersion, DatasetDigest: dataset.DatasetDigest}
}
