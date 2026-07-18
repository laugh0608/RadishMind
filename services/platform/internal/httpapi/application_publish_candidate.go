package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	applicationPublishCandidateSchemaVersionV1 = "application_publish_candidate.v1"
	applicationPublishCandidateSchemaVersionV2 = "application_publish_candidate.v2"
	applicationPublishCandidateSchemaVersion   = applicationPublishCandidateSchemaVersionV1
)

const (
	ApplicationPublishFailureScopeDenied             = "publish_candidate_scope_denied"
	ApplicationPublishFailureNotFound                = "publish_candidate_not_found"
	ApplicationPublishFailurePayloadInvalid          = "publish_candidate_payload_invalid"
	ApplicationPublishFailureSecretForbidden         = "publish_candidate_secret_material_forbidden"
	ApplicationPublishFailureDraftNotFound           = "publish_candidate_draft_not_found"
	ApplicationPublishFailureDraftVersionConflict    = "publish_candidate_draft_version_conflict"
	ApplicationPublishFailureDraftInvalid            = "publish_candidate_draft_invalid"
	ApplicationPublishFailureDraftChanged            = "publish_candidate_draft_changed"
	ApplicationPublishFailureBindingNotEligible      = WorkflowRAGPromotionFailureBindingNotEligible
	ApplicationPublishFailureBaseRevisionChanged     = "application_base_revision_changed"
	ApplicationPublishFailureImmutableConflict       = "publish_candidate_immutable_conflict"
	ApplicationPublishFailureReviewVersionConflict   = "publish_candidate_review_version_conflict"
	ApplicationPublishFailureReviewTransitionInvalid = "publish_candidate_review_transition_invalid"
	ApplicationPublishFailureStoreUnavailable        = "publish_candidate_store_unavailable"
	ApplicationPublishFailureWriteDisabled           = "publish_candidate_write_disabled"
	ApplicationPublishFailureApplicationArchived     = ApplicationCatalogFailureArchived
)

const (
	applicationPublishStatePending        = "pending_review"
	applicationPublishStateApproved       = "approved"
	applicationPublishStateRejected       = "rejected"
	applicationPublishStateChangesNeeded  = "changes_requested"
	applicationPublishStateWithdrawn      = "withdrawn"
	applicationPublishStatusBlocked       = "promotion_blocked"
	applicationPublishDecisionApprove     = "approve"
	applicationPublishDecisionReject      = "reject"
	applicationPublishDecisionChanges     = "request_changes"
	applicationPublishDecisionWithdraw    = "withdraw"
	applicationPublishMaxEvidenceRequests = 20
)

var (
	errApplicationPublishNotFound            = errors.New(ApplicationPublishFailureNotFound)
	errApplicationPublishImmutableConflict   = errors.New(ApplicationPublishFailureImmutableConflict)
	errApplicationPublishReviewConflict      = errors.New(ApplicationPublishFailureReviewVersionConflict)
	errApplicationPublishReviewTransition    = errors.New(ApplicationPublishFailureReviewTransitionInvalid)
	errApplicationPublishStoreUnavailable    = errors.New(ApplicationPublishFailureStoreUnavailable)
	errApplicationPublishBaselineUnavailable = errors.New("application publish baseline unavailable")
	errApplicationPublishBaselineNotFound    = errors.New("application publish baseline not found")
)

type applicationPublishReviewConflictError struct {
	CurrentVersion int
	CurrentState   string
}

func (err applicationPublishReviewConflictError) Error() string {
	return ApplicationPublishFailureReviewVersionConflict
}
func (err applicationPublishReviewConflictError) Is(target error) bool {
	return target == errApplicationPublishReviewConflict
}

type ApplicationPublishContext struct {
	RequestContext          context.Context
	RequestID               string
	TenantRef               string
	WorkspaceID             string
	ApplicationID           string
	ActorRef                string
	OwnerSubjectRef         string
	AuditRef                string
	WriteEnabled            bool
	RAGPromotionReadEnabled bool
}

type ApplicationPublishConfigurationSnapshot struct {
	DisplayName           string                            `json:"display_name"`
	Description           string                            `json:"description"`
	ApplicationKind       string                            `json:"application_kind"`
	DefaultProtocol       string                            `json:"default_protocol"`
	DefaultModel          string                            `json:"default_model"`
	AllowedProtocols      []string                          `json:"allowed_protocols"`
	WorkflowRAGBindingRef *WorkflowRAGApplicationBindingRef `json:"workflow_rag_binding_ref,omitempty"`
}

type ApplicationPublishReviewRecord struct {
	ReviewVersion int    `json:"review_version"`
	Decision      string `json:"decision"`
	Reason        string `json:"reason"`
	State         string `json:"state"`
	ReviewedAt    string `json:"reviewed_at"`
	ReviewerRef   string `json:"reviewer_ref"`
	RequestID     string `json:"request_id"`
	AuditRef      string `json:"audit_ref"`
}

type ApplicationPromotionBlocker struct {
	Code    string `json:"code"`
	Summary string `json:"summary"`
}

type ApplicationPromotionEligibility struct {
	Eligible bool                          `json:"eligible"`
	Status   string                        `json:"status"`
	Blockers []ApplicationPromotionBlocker `json:"blockers"`
}

type ApplicationPublishCandidate struct {
	SchemaVersion            string                                  `json:"schema_version"`
	CandidateID              string                                  `json:"candidate_id"`
	WorkspaceID              string                                  `json:"workspace_id"`
	ApplicationID            string                                  `json:"application_id"`
	DraftID                  string                                  `json:"draft_id"`
	DraftVersion             int                                     `json:"draft_version"`
	DraftDigest              string                                  `json:"draft_digest"`
	BaseApplicationUpdatedAt string                                  `json:"base_application_updated_at"`
	Configuration            ApplicationPublishConfigurationSnapshot `json:"configuration"`
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

type ApplicationPublishCandidateSummary struct {
	CandidateID           string                            `json:"candidate_id"`
	ApplicationID         string                            `json:"application_id"`
	DraftID               string                            `json:"draft_id"`
	DraftVersion          int                               `json:"draft_version"`
	DraftDigest           string                            `json:"draft_digest"`
	CandidateState        string                            `json:"candidate_state"`
	ReviewVersion         int                               `json:"review_version"`
	PromotionStatus       string                            `json:"promotion_status"`
	PromotionBlockers     int                               `json:"promotion_blockers"`
	WorkflowRAGBindingRef *WorkflowRAGApplicationBindingRef `json:"workflow_rag_binding_ref,omitempty"`
	CreatedAt             string                            `json:"created_at"`
	UpdatedAt             string                            `json:"updated_at"`
	UpdatedByActorRef     string                            `json:"updated_by_actor_ref"`
}

type ApplicationPublishCreateInput struct {
	CandidateID          string
	DraftID              string
	ExpectedDraftVersion int
	EvidenceRequestIDs   []string
}

type ApplicationPublishReviewInput struct {
	ExpectedReviewVersion int
	Decision              string
	Reason                string
}

type ApplicationPublishResult struct {
	Candidate             *ApplicationPublishCandidate
	FailureCode           string
	CurrentReviewVersion  int
	CurrentCandidateState string
	CurrentDraftVersion   int
}

type applicationPublishCandidateRepository interface {
	Create(ApplicationPublishContext, ApplicationPublishCandidate) (ApplicationPublishCandidate, error)
	Read(ApplicationPublishContext, string) (ApplicationPublishCandidate, error)
	List(ApplicationPublishContext) ([]ApplicationPublishCandidate, error)
	AppendReview(ApplicationPublishContext, string, int, ApplicationPublishReviewRecord) (ApplicationPublishCandidate, error)
}

type memoryApplicationPublishCandidateRepository struct {
	mu          sync.RWMutex
	candidates  map[string]ApplicationPublishCandidate
	unavailable bool
}

type applicationPublishBaselineReader func(ApplicationPublishContext) (ApplicationSummary, error)

type applicationPublishCandidateService struct {
	draftRepository     applicationConfigurationDraftRepository
	candidateRepository applicationPublishCandidateRepository
	readBaseline        applicationPublishBaselineReader
	validateBinding     func(ApplicationPublishContext, WorkflowRAGApplicationBindingRef) (WorkflowRAGApplicationBinding, string)
	now                 func() time.Time
}

func newMemoryApplicationPublishCandidateRepository() *memoryApplicationPublishCandidateRepository {
	return &memoryApplicationPublishCandidateRepository{candidates: make(map[string]ApplicationPublishCandidate)}
}

func newApplicationPublishCandidateService(
	draftRepository applicationConfigurationDraftRepository,
	candidateRepository applicationPublishCandidateRepository,
	readBaseline applicationPublishBaselineReader,
) applicationPublishCandidateService {
	return applicationPublishCandidateService{
		draftRepository: draftRepository, candidateRepository: candidateRepository, readBaseline: readBaseline,
		now: func() time.Time { return time.Now().UTC() },
	}
}

func (service applicationPublishCandidateService) Create(requestContext ApplicationPublishContext, input ApplicationPublishCreateInput) ApplicationPublishResult {
	if !requestContext.WriteEnabled {
		return ApplicationPublishResult{FailureCode: ApplicationPublishFailureWriteDisabled}
	}
	input.CandidateID = strings.TrimSpace(input.CandidateID)
	input.DraftID = strings.TrimSpace(input.DraftID)
	if !applicationDraftIdentifierPattern.MatchString(input.CandidateID) || !applicationDraftIdentifierPattern.MatchString(input.DraftID) || input.ExpectedDraftVersion < 1 {
		return ApplicationPublishResult{FailureCode: ApplicationPublishFailurePayloadInvalid}
	}
	evidence, failure := normalizeApplicationPublishEvidence(input.EvidenceRequestIDs)
	if failure != "" {
		return ApplicationPublishResult{FailureCode: failure}
	}
	draftContext := applicationPublishDraftContext(requestContext)
	draft, err := service.draftRepository.Read(draftContext, input.DraftID)
	if errors.Is(err, errApplicationDraftNotFound) {
		return ApplicationPublishResult{FailureCode: ApplicationPublishFailureDraftNotFound}
	}
	if err != nil {
		return ApplicationPublishResult{FailureCode: ApplicationPublishFailureStoreUnavailable}
	}
	if draft.DraftVersion != input.ExpectedDraftVersion {
		return ApplicationPublishResult{FailureCode: ApplicationPublishFailureDraftVersionConflict, CurrentDraftVersion: draft.DraftVersion}
	}
	validation := validateApplicationConfigurationDraftPayload(draftContext, draft.ApplicationConfigurationDraftPayload)
	if !draft.ValidationSummary.IsValid || !validation.IsValid {
		return ApplicationPublishResult{FailureCode: ApplicationPublishFailureDraftInvalid, CurrentDraftVersion: draft.DraftVersion}
	}
	baseline, err := service.readBaseline(requestContext)
	if errors.Is(err, errApplicationCatalogArchived) {
		return ApplicationPublishResult{FailureCode: ApplicationPublishFailureApplicationArchived, CurrentDraftVersion: draft.DraftVersion}
	}
	if err != nil || strings.TrimSpace(baseline.ApplicationRef) == "" {
		return ApplicationPublishResult{FailureCode: ApplicationPublishFailureStoreUnavailable}
	}
	if strings.TrimSpace(baseline.UpdatedAt) != strings.TrimSpace(draft.BaseApplicationUpdatedAt) {
		return ApplicationPublishResult{FailureCode: ApplicationPublishFailureBaseRevisionChanged, CurrentDraftVersion: draft.DraftVersion}
	}
	snapshot := applicationPublishSnapshotFromDraft(draft)
	digest, err := applicationConfigurationCanonicalDigest(snapshot)
	if err != nil || draft.DraftDigest != "" && draft.DraftDigest != digest {
		return ApplicationPublishResult{FailureCode: ApplicationPublishFailureStoreUnavailable}
	}
	if snapshot.WorkflowRAGBindingRef != nil {
		if !requestContext.RAGPromotionReadEnabled {
			return ApplicationPublishResult{FailureCode: ApplicationPublishFailureScopeDenied}
		}
		if _, failureCode := service.resolveBinding(requestContext, *snapshot.WorkflowRAGBindingRef); failureCode != "" {
			return ApplicationPublishResult{FailureCode: applicationPublishBindingMutationFailure(failureCode)}
		}
	}
	now := service.now().Format(time.RFC3339Nano)
	schemaVersion := applicationPublishCandidateSchemaVersionV1
	if snapshot.WorkflowRAGBindingRef != nil {
		schemaVersion = applicationPublishCandidateSchemaVersionV2
	}
	candidate := ApplicationPublishCandidate{
		SchemaVersion: schemaVersion, CandidateID: input.CandidateID,
		WorkspaceID: requestContext.WorkspaceID, ApplicationID: requestContext.ApplicationID,
		DraftID: draft.DraftID, DraftVersion: draft.DraftVersion, DraftDigest: digest,
		BaseApplicationUpdatedAt: draft.BaseApplicationUpdatedAt, Configuration: snapshot,
		EvidenceRequestIDs: evidence, CandidateState: applicationPublishStatePending, ReviewVersion: 0,
		Reviews: []ApplicationPublishReviewRecord{}, CreatedAt: now, UpdatedAt: now,
		CreatedByActorRef: requestContext.ActorRef, UpdatedByActorRef: requestContext.ActorRef,
		RequestID: requestContext.RequestID, AuditRef: requestContext.AuditRef,
	}
	created, err := service.candidateRepository.Create(requestContext, candidate)
	if err != nil {
		return applicationPublishRepositoryFailure(err)
	}
	created = service.decorate(requestContext, created)
	return ApplicationPublishResult{Candidate: &created, CurrentCandidateState: created.CandidateState}
}

func (service applicationPublishCandidateService) Read(requestContext ApplicationPublishContext, candidateID string) ApplicationPublishResult {
	candidateID = strings.TrimSpace(candidateID)
	if !applicationDraftIdentifierPattern.MatchString(candidateID) {
		return ApplicationPublishResult{FailureCode: ApplicationPublishFailurePayloadInvalid}
	}
	candidate, err := service.candidateRepository.Read(requestContext, candidateID)
	if err != nil {
		return applicationPublishRepositoryFailure(err)
	}
	candidate = service.decorate(requestContext, candidate)
	return ApplicationPublishResult{Candidate: &candidate, CurrentReviewVersion: candidate.ReviewVersion, CurrentCandidateState: candidate.CandidateState}
}

func (service applicationPublishCandidateService) List(requestContext ApplicationPublishContext) ([]ApplicationPublishCandidateSummary, string) {
	candidates, err := service.candidateRepository.List(requestContext)
	if err != nil {
		return []ApplicationPublishCandidateSummary{}, ApplicationPublishFailureStoreUnavailable
	}
	summaries := make([]ApplicationPublishCandidateSummary, 0, len(candidates))
	for _, candidate := range candidates {
		candidate = service.decorateFromCandidates(requestContext, candidate, candidates)
		summaries = append(summaries, applicationPublishCandidateSummary(candidate))
	}
	return summaries, ""
}

func (service applicationPublishCandidateService) Review(requestContext ApplicationPublishContext, candidateID string, input ApplicationPublishReviewInput) ApplicationPublishResult {
	if !requestContext.WriteEnabled {
		return ApplicationPublishResult{FailureCode: ApplicationPublishFailureWriteDisabled}
	}
	candidateID = strings.TrimSpace(candidateID)
	decision := strings.TrimSpace(input.Decision)
	reason := strings.TrimSpace(input.Reason)
	if !applicationDraftIdentifierPattern.MatchString(candidateID) || input.ExpectedReviewVersion < 0 || !isApplicationPublishDecision(decision) || len(reason) < 4 || len(reason) > 500 {
		return ApplicationPublishResult{FailureCode: ApplicationPublishFailurePayloadInvalid}
	}
	if applicationDraftStringContainsSecret(reason) {
		return ApplicationPublishResult{FailureCode: ApplicationPublishFailureSecretForbidden}
	}
	if _, err := service.readBaseline(requestContext); errors.Is(err, errApplicationCatalogArchived) {
		return ApplicationPublishResult{FailureCode: ApplicationPublishFailureApplicationArchived}
	} else if err != nil {
		return ApplicationPublishResult{FailureCode: ApplicationPublishFailureStoreUnavailable}
	}
	if decision == applicationPublishDecisionApprove {
		candidate, err := service.candidateRepository.Read(requestContext, candidateID)
		if err != nil {
			return applicationPublishRepositoryFailure(err)
		}
		if candidate.ReviewVersion != input.ExpectedReviewVersion {
			return ApplicationPublishResult{FailureCode: ApplicationPublishFailureReviewVersionConflict, CurrentReviewVersion: candidate.ReviewVersion, CurrentCandidateState: candidate.CandidateState}
		}
		if !applicationPublishTransitionAllowed(candidate.CandidateState, decision) {
			return ApplicationPublishResult{FailureCode: ApplicationPublishFailureReviewTransitionInvalid}
		}
		if failureCode := service.validateApproval(requestContext, candidate); failureCode != "" {
			return ApplicationPublishResult{FailureCode: failureCode, CurrentReviewVersion: candidate.ReviewVersion, CurrentCandidateState: candidate.CandidateState}
		}
	}
	state := applicationPublishStateForDecision(decision)
	now := service.now().Format(time.RFC3339Nano)
	review := ApplicationPublishReviewRecord{
		ReviewVersion: input.ExpectedReviewVersion + 1, Decision: decision, Reason: reason, State: state,
		ReviewedAt: now, ReviewerRef: requestContext.ActorRef, RequestID: requestContext.RequestID, AuditRef: requestContext.AuditRef,
	}
	updated, err := service.candidateRepository.AppendReview(requestContext, candidateID, input.ExpectedReviewVersion, review)
	if err != nil {
		return applicationPublishRepositoryFailure(err)
	}
	updated = service.decorate(requestContext, updated)
	return ApplicationPublishResult{Candidate: &updated, CurrentReviewVersion: updated.ReviewVersion, CurrentCandidateState: updated.CandidateState}
}

func (service applicationPublishCandidateService) decorate(requestContext ApplicationPublishContext, candidate ApplicationPublishCandidate) ApplicationPublishCandidate {
	candidates, err := service.candidateRepository.List(requestContext)
	if err != nil {
		candidate.PromotionEligibility = blockedApplicationPromotionEligibility([]ApplicationPromotionBlocker{{Code: ApplicationPublishFailureStoreUnavailable, Summary: "Candidate store is unavailable for eligibility review."}})
		return candidate
	}
	return service.decorateFromCandidates(requestContext, candidate, candidates)
}

func (service applicationPublishCandidateService) decorateFromCandidates(requestContext ApplicationPublishContext, candidate ApplicationPublishCandidate, candidates []ApplicationPublishCandidate) ApplicationPublishCandidate {
	blockers := applicationPublishReviewBlockers(candidate)
	baseline, err := service.readBaseline(requestContext)
	if errors.Is(err, errApplicationCatalogArchived) {
		blockers = append(blockers, ApplicationPromotionBlocker{Code: ApplicationPublishFailureApplicationArchived, Summary: "Application is archived and cannot continue publish review."})
	} else if err != nil {
		blockers = append(blockers, ApplicationPromotionBlocker{Code: "application_baseline_unavailable", Summary: "Current application baseline is unavailable."})
	} else if strings.TrimSpace(baseline.UpdatedAt) != strings.TrimSpace(candidate.BaseApplicationUpdatedAt) {
		blockers = append(blockers, ApplicationPromotionBlocker{Code: ApplicationPublishFailureBaseRevisionChanged, Summary: "Application baseline changed after candidate creation."})
	}
	draft, draftErr := service.draftRepository.Read(applicationPublishDraftContext(requestContext), candidate.DraftID)
	if draftErr != nil {
		blockers = append(blockers, ApplicationPromotionBlocker{Code: "publish_candidate_draft_unavailable", Summary: "Bound application draft is unavailable."})
	} else {
		digest, digestErr := applicationConfigurationCanonicalDigest(applicationPublishSnapshotFromDraft(draft))
		if digestErr != nil || draft.DraftVersion != candidate.DraftVersion || digest != candidate.DraftDigest {
			blockers = append(blockers, ApplicationPromotionBlocker{Code: ApplicationPublishFailureDraftChanged, Summary: "Saved draft version or sanitized digest changed after candidate creation."})
		}
	}
	if candidate.Configuration.WorkflowRAGBindingRef != nil {
		if !requestContext.RAGPromotionReadEnabled {
			blockers = append(blockers, ApplicationPromotionBlocker{Code: ApplicationPublishFailureScopeDenied, Summary: "Workflow RAG binding review permission is required."})
		} else if _, failureCode := service.resolveBinding(requestContext, *candidate.Configuration.WorkflowRAGBindingRef); failureCode != "" {
			blockers = append(blockers, applicationPublishBindingBlocker(failureCode))
		}
	}
	if applicationPublishCandidateIsSuperseded(candidate, candidates) {
		blockers = append(blockers, ApplicationPromotionBlocker{Code: "publish_candidate_superseded", Summary: "A newer candidate exists for the bound draft."})
	}
	blockers = append(blockers,
		ApplicationPromotionBlocker{Code: "formal_application_repository_unavailable", Summary: "Formal application repository is not configured."},
		ApplicationPromotionBlocker{Code: "publish_auth_not_configured", Summary: "Production auth and membership verification are not configured."},
		ApplicationPromotionBlocker{Code: "publish_owner_not_configured", Summary: "Production publish owner is not configured."},
		ApplicationPromotionBlocker{Code: "promotion_disabled", Summary: "Application promotion runtime is disabled."},
	)
	candidate.PromotionEligibility = blockedApplicationPromotionEligibility(blockers)
	return candidate
}

func applicationPublishReviewBlockers(candidate ApplicationPublishCandidate) []ApplicationPromotionBlocker {
	switch candidate.CandidateState {
	case applicationPublishStateApproved:
		return []ApplicationPromotionBlocker{}
	case applicationPublishStateRejected:
		return []ApplicationPromotionBlocker{{Code: "publish_review_rejected", Summary: "Publish candidate was rejected."}}
	case applicationPublishStateChangesNeeded:
		return []ApplicationPromotionBlocker{{Code: "publish_changes_requested", Summary: "Publish candidate requires a new draft and candidate."}}
	case applicationPublishStateWithdrawn:
		return []ApplicationPromotionBlocker{{Code: "publish_candidate_withdrawn", Summary: "Publish candidate was withdrawn."}}
	default:
		return []ApplicationPromotionBlocker{{Code: "publish_review_required", Summary: "Publish candidate requires an explicit review decision."}}
	}
}

func blockedApplicationPromotionEligibility(blockers []ApplicationPromotionBlocker) ApplicationPromotionEligibility {
	return ApplicationPromotionEligibility{Eligible: false, Status: applicationPublishStatusBlocked, Blockers: blockers}
}

func applicationPublishDraftContext(requestContext ApplicationPublishContext) ApplicationConfigurationDraftContext {
	return ApplicationConfigurationDraftContext{
		RequestContext: requestContext.RequestContext, RequestID: requestContext.RequestID,
		TenantRef: requestContext.TenantRef, WorkspaceID: requestContext.WorkspaceID, ApplicationID: requestContext.ApplicationID,
		ActorRef: requestContext.ActorRef, OwnerSubjectRef: requestContext.OwnerSubjectRef,
		AuditRef: requestContext.AuditRef, WriteEnabled: requestContext.WriteEnabled,
	}
}

func (service applicationPublishCandidateService) resolveBinding(requestContext ApplicationPublishContext, ref WorkflowRAGApplicationBindingRef) (WorkflowRAGApplicationBinding, string) {
	if service.validateBinding == nil {
		return WorkflowRAGApplicationBinding{}, WorkflowRAGPromotionFailureStoreUnavailable
	}
	return service.validateBinding(requestContext, ref)
}

func (service applicationPublishCandidateService) validateApproval(requestContext ApplicationPublishContext, candidate ApplicationPublishCandidate) string {
	baseline, err := service.readBaseline(requestContext)
	if errors.Is(err, errApplicationCatalogArchived) {
		return ApplicationPublishFailureApplicationArchived
	}
	if err != nil {
		return ApplicationPublishFailureStoreUnavailable
	}
	if strings.TrimSpace(baseline.UpdatedAt) != strings.TrimSpace(candidate.BaseApplicationUpdatedAt) {
		return ApplicationPublishFailureBaseRevisionChanged
	}
	draft, err := service.draftRepository.Read(applicationPublishDraftContext(requestContext), candidate.DraftID)
	if errors.Is(err, errApplicationDraftNotFound) {
		return ApplicationPublishFailureDraftChanged
	}
	if err != nil {
		return ApplicationPublishFailureStoreUnavailable
	}
	digest, err := applicationConfigurationCanonicalDigest(applicationPublishSnapshotFromDraft(draft))
	if err != nil || draft.DraftVersion != candidate.DraftVersion || digest != candidate.DraftDigest {
		return ApplicationPublishFailureDraftChanged
	}
	if !draft.ValidationSummary.IsValid || !validateApplicationConfigurationDraftPayload(applicationPublishDraftContext(requestContext), draft.ApplicationConfigurationDraftPayload).IsValid {
		return ApplicationPublishFailureDraftInvalid
	}
	if !workflowRAGPromotionBindingRefsEqual(draft.WorkflowRAGBindingRef, candidate.Configuration.WorkflowRAGBindingRef) {
		return ApplicationPublishFailureBindingNotEligible
	}
	if candidate.Configuration.WorkflowRAGBindingRef != nil {
		if !requestContext.RAGPromotionReadEnabled {
			return ApplicationPublishFailureScopeDenied
		}
		if _, failureCode := service.resolveBinding(requestContext, *candidate.Configuration.WorkflowRAGBindingRef); failureCode != "" {
			return applicationPublishBindingMutationFailure(failureCode)
		}
	}
	return ""
}

func applicationPublishBindingMutationFailure(failureCode string) string {
	switch failureCode {
	case WorkflowRAGPromotionFailureStoreUnavailable, WorkflowRAGPromotionFailureStoreContractMismatch:
		return failureCode
	case WorkflowRAGPromotionFailureScopeDenied:
		return ApplicationPublishFailureScopeDenied
	default:
		return ApplicationPublishFailureBindingNotEligible
	}
}

func applicationPublishBindingBlocker(failureCode string) ApplicationPromotionBlocker {
	summary := "Workflow RAG binding is no longer eligible."
	if failureCode == WorkflowRAGPromotionFailureStoreUnavailable || failureCode == WorkflowRAGPromotionFailureStoreContractMismatch {
		summary = "Workflow RAG binding authority is unavailable."
	}
	return ApplicationPromotionBlocker{Code: failureCode, Summary: summary}
}

func applicationPublishSnapshotFromDraft(draft ApplicationConfigurationDraft) ApplicationPublishConfigurationSnapshot {
	return ApplicationPublishConfigurationSnapshot{
		DisplayName: strings.TrimSpace(draft.DisplayName), Description: strings.TrimSpace(draft.Description),
		ApplicationKind: strings.TrimSpace(draft.ApplicationKind), DefaultProtocol: strings.TrimSpace(draft.DefaultProtocol),
		DefaultModel: strings.TrimSpace(draft.DefaultModel), AllowedProtocols: normalizeApplicationDraftProtocols(draft.AllowedProtocols),
		WorkflowRAGBindingRef: cloneWorkflowRAGApplicationBindingRef(draft.WorkflowRAGBindingRef),
	}
}

func applicationConfigurationCanonicalDigest(snapshot ApplicationPublishConfigurationSnapshot) (string, error) {
	payload, err := json.Marshal(snapshot)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(digest[:]), nil
}

func normalizeApplicationPublishEvidence(requestIDs []string) ([]string, string) {
	if len(requestIDs) > applicationPublishMaxEvidenceRequests {
		return nil, ApplicationPublishFailurePayloadInvalid
	}
	seen := make(map[string]bool)
	normalized := make([]string, 0, len(requestIDs))
	for _, requestID := range requestIDs {
		requestID = strings.TrimSpace(requestID)
		if !validGatewayRequestReference(requestID, 160) {
			return nil, ApplicationPublishFailurePayloadInvalid
		}
		if applicationDraftStringContainsSecret(requestID) {
			return nil, ApplicationPublishFailureSecretForbidden
		}
		if !seen[requestID] {
			seen[requestID] = true
			normalized = append(normalized, requestID)
		}
	}
	sort.Strings(normalized)
	return normalized, ""
}

func isApplicationPublishDecision(decision string) bool {
	return decision == applicationPublishDecisionApprove || decision == applicationPublishDecisionReject || decision == applicationPublishDecisionChanges || decision == applicationPublishDecisionWithdraw
}

func applicationPublishStateForDecision(decision string) string {
	switch decision {
	case applicationPublishDecisionApprove:
		return applicationPublishStateApproved
	case applicationPublishDecisionReject:
		return applicationPublishStateRejected
	case applicationPublishDecisionChanges:
		return applicationPublishStateChangesNeeded
	default:
		return applicationPublishStateWithdrawn
	}
}

func applicationPublishTransitionAllowed(currentState, decision string) bool {
	if currentState == applicationPublishStatePending {
		return isApplicationPublishDecision(decision)
	}
	return currentState == applicationPublishStateApproved && decision == applicationPublishDecisionWithdraw
}

func applicationPublishCandidateIsSuperseded(candidate ApplicationPublishCandidate, candidates []ApplicationPublishCandidate) bool {
	for _, other := range candidates {
		if other.CandidateID == candidate.CandidateID || other.DraftID != candidate.DraftID {
			continue
		}
		if other.DraftVersion > candidate.DraftVersion || other.DraftVersion == candidate.DraftVersion && (other.CreatedAt > candidate.CreatedAt || other.CreatedAt == candidate.CreatedAt && other.CandidateID > candidate.CandidateID) {
			return true
		}
	}
	return false
}

func applicationPublishRepositoryFailure(err error) ApplicationPublishResult {
	result := ApplicationPublishResult{FailureCode: ApplicationPublishFailureStoreUnavailable}
	switch {
	case errors.Is(err, errApplicationPublishNotFound):
		result.FailureCode = ApplicationPublishFailureNotFound
	case errors.Is(err, errApplicationPublishImmutableConflict):
		result.FailureCode = ApplicationPublishFailureImmutableConflict
	case errors.Is(err, errApplicationPublishReviewTransition):
		result.FailureCode = ApplicationPublishFailureReviewTransitionInvalid
	case errors.Is(err, errApplicationPublishReviewConflict):
		result.FailureCode = ApplicationPublishFailureReviewVersionConflict
		var conflict applicationPublishReviewConflictError
		if errors.As(err, &conflict) {
			result.CurrentReviewVersion = conflict.CurrentVersion
			result.CurrentCandidateState = conflict.CurrentState
		}
	}
	return result
}

func applicationPublishRepositoryKey(requestContext ApplicationPublishContext, candidateID string) string {
	return strings.Join([]string{requestContext.TenantRef, requestContext.WorkspaceID, requestContext.ApplicationID, requestContext.OwnerSubjectRef, candidateID}, "\x00")
}

func (repository *memoryApplicationPublishCandidateRepository) Create(requestContext ApplicationPublishContext, candidate ApplicationPublishCandidate) (ApplicationPublishCandidate, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if repository.unavailable {
		return ApplicationPublishCandidate{}, errApplicationPublishStoreUnavailable
	}
	key := applicationPublishRepositoryKey(requestContext, candidate.CandidateID)
	if _, exists := repository.candidates[key]; exists {
		return ApplicationPublishCandidate{}, errApplicationPublishImmutableConflict
	}
	repository.candidates[key] = cloneApplicationPublishCandidate(candidate)
	return cloneApplicationPublishCandidate(candidate), nil
}

func (repository *memoryApplicationPublishCandidateRepository) Read(requestContext ApplicationPublishContext, candidateID string) (ApplicationPublishCandidate, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	if repository.unavailable {
		return ApplicationPublishCandidate{}, errApplicationPublishStoreUnavailable
	}
	candidate, exists := repository.candidates[applicationPublishRepositoryKey(requestContext, candidateID)]
	if !exists {
		return ApplicationPublishCandidate{}, errApplicationPublishNotFound
	}
	return cloneApplicationPublishCandidate(candidate), nil
}

func (repository *memoryApplicationPublishCandidateRepository) List(requestContext ApplicationPublishContext) ([]ApplicationPublishCandidate, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	if repository.unavailable {
		return nil, errApplicationPublishStoreUnavailable
	}
	prefix := applicationPublishRepositoryKey(requestContext, "")
	candidates := make([]ApplicationPublishCandidate, 0)
	for key, candidate := range repository.candidates {
		if strings.HasPrefix(key, prefix) {
			candidates = append(candidates, cloneApplicationPublishCandidate(candidate))
		}
	}
	sortApplicationPublishCandidates(candidates)
	return candidates, nil
}

func (repository *memoryApplicationPublishCandidateRepository) AppendReview(requestContext ApplicationPublishContext, candidateID string, expectedVersion int, review ApplicationPublishReviewRecord) (ApplicationPublishCandidate, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if repository.unavailable {
		return ApplicationPublishCandidate{}, errApplicationPublishStoreUnavailable
	}
	key := applicationPublishRepositoryKey(requestContext, candidateID)
	candidate, exists := repository.candidates[key]
	if !exists {
		return ApplicationPublishCandidate{}, errApplicationPublishNotFound
	}
	if candidate.ReviewVersion != expectedVersion {
		return ApplicationPublishCandidate{}, applicationPublishReviewConflictError{CurrentVersion: candidate.ReviewVersion, CurrentState: candidate.CandidateState}
	}
	if !applicationPublishTransitionAllowed(candidate.CandidateState, review.Decision) {
		return ApplicationPublishCandidate{}, errApplicationPublishReviewTransition
	}
	candidate.ReviewVersion = review.ReviewVersion
	candidate.CandidateState = review.State
	candidate.Reviews = append(candidate.Reviews, review)
	candidate.UpdatedAt = review.ReviewedAt
	candidate.UpdatedByActorRef = review.ReviewerRef
	candidate.RequestID = review.RequestID
	candidate.AuditRef = review.AuditRef
	repository.candidates[key] = cloneApplicationPublishCandidate(candidate)
	return cloneApplicationPublishCandidate(candidate), nil
}

func applicationPublishCandidateSummary(candidate ApplicationPublishCandidate) ApplicationPublishCandidateSummary {
	return ApplicationPublishCandidateSummary{
		CandidateID: candidate.CandidateID, ApplicationID: candidate.ApplicationID, DraftID: candidate.DraftID,
		DraftVersion: candidate.DraftVersion, DraftDigest: candidate.DraftDigest, CandidateState: candidate.CandidateState,
		ReviewVersion: candidate.ReviewVersion, PromotionStatus: candidate.PromotionEligibility.Status,
		PromotionBlockers: len(candidate.PromotionEligibility.Blockers), CreatedAt: candidate.CreatedAt,
		UpdatedAt: candidate.UpdatedAt, UpdatedByActorRef: candidate.UpdatedByActorRef,
		WorkflowRAGBindingRef: cloneWorkflowRAGApplicationBindingRef(candidate.Configuration.WorkflowRAGBindingRef),
	}
}

func sortApplicationPublishCandidates(candidates []ApplicationPublishCandidate) {
	sort.Slice(candidates, func(left, right int) bool {
		if candidates[left].CreatedAt == candidates[right].CreatedAt {
			return candidates[left].CandidateID > candidates[right].CandidateID
		}
		return candidates[left].CreatedAt > candidates[right].CreatedAt
	})
}

func cloneApplicationPublishCandidate(candidate ApplicationPublishCandidate) ApplicationPublishCandidate {
	candidate.Configuration.AllowedProtocols = append([]string(nil), candidate.Configuration.AllowedProtocols...)
	candidate.Configuration.WorkflowRAGBindingRef = cloneWorkflowRAGApplicationBindingRef(candidate.Configuration.WorkflowRAGBindingRef)
	candidate.EvidenceRequestIDs = append([]string(nil), candidate.EvidenceRequestIDs...)
	candidate.Reviews = append([]ApplicationPublishReviewRecord(nil), candidate.Reviews...)
	candidate.PromotionEligibility.Blockers = append([]ApplicationPromotionBlocker(nil), candidate.PromotionEligibility.Blockers...)
	return candidate
}
