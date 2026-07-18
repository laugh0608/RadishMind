package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

var (
	workflowRAGPromotionCandidateIDPattern = regexp.MustCompile(`^wragp_[a-z2-7]{16}$`)
	workflowRAGPromotionDecisionIDPattern  = regexp.MustCompile(`^wragpd_[a-z2-7]{16}$`)
	workflowRAGPromotionBindingIDPattern   = regexp.MustCompile(`^wragb_[a-z2-7]{16}$`)
	workflowRAGPromotionAuditIDPattern     = regexp.MustCompile(`^wragpa_[a-z2-7]{16}$`)

	errWorkflowRAGPromotionNotFound      = errors.New(WorkflowRAGPromotionFailureNotFound)
	errWorkflowRAGPromotionScopeDenied   = errors.New(WorkflowRAGPromotionFailureScopeDenied)
	errWorkflowRAGPromotionConflict      = errors.New(WorkflowRAGPromotionFailureRecordConflict)
	errWorkflowRAGPromotionTransition    = errors.New(WorkflowRAGPromotionFailureTransitionInvalid)
	errWorkflowRAGPromotionStore         = errors.New(WorkflowRAGPromotionFailureStoreUnavailable)
	errWorkflowRAGPromotionStoreContract = errors.New(WorkflowRAGPromotionFailureStoreContractMismatch)
)

type workflowRAGPromotionConflictError struct {
	CurrentVersion int
	CurrentState   string
}

func (failure workflowRAGPromotionConflictError) Error() string {
	return WorkflowRAGPromotionFailureRecordConflict
}

func (failure workflowRAGPromotionConflictError) Is(target error) bool {
	return target == errWorkflowRAGPromotionConflict
}

type workflowRAGPromotionMemoryEntry struct {
	candidate WorkflowRAGKnowledgePromotionCandidate
	decisions []WorkflowRAGKnowledgePromotionDecision
	binding   *WorkflowRAGApplicationBinding
	audits    []WorkflowRAGPromotionAudit
}

type memoryWorkflowRAGPromotionRepository struct {
	ownerLock *sync.RWMutex
	entries   map[string]workflowRAGPromotionMemoryEntry
	capacity  int
	available bool
}

func newMemoryWorkflowRAGPromotionRepository(ownerLock *sync.RWMutex) *memoryWorkflowRAGPromotionRepository {
	if ownerLock == nil {
		ownerLock = &sync.RWMutex{}
	}
	return &memoryWorkflowRAGPromotionRepository{
		ownerLock: ownerLock, entries: make(map[string]workflowRAGPromotionMemoryEntry),
		capacity: workflowRAGPromotionMaximumCandidates, available: true,
	}
}

func newWorkflowRAGPromotionRepositoryForRunStore(store workflowRunStore) (workflowRAGPromotionRepository, error) {
	switch typed := store.(type) {
	case *memoryWorkflowRunStore:
		return newMemoryWorkflowRAGPromotionRepository(&typed.mu), nil
	case *sqliteWorkflowRunStore, *postgresWorkflowRunStore:
		return nil, fmt.Errorf("workflow RAG promotion durable repository requires batch B migrations")
	default:
		return nil, fmt.Errorf("workflow RAG promotion repository requires a supported workflow run store")
	}
}

func (repository *memoryWorkflowRAGPromotionRepository) Create(ctx WorkflowRAGPromotionContext, candidate WorkflowRAGKnowledgePromotionCandidate, audit WorkflowRAGPromotionAudit) error {
	if !repository.available {
		return errWorkflowRAGPromotionStore
	}
	if validateStoredWorkflowRAGPromotionCandidate(candidate, ctx) != nil || candidate.RecordVersion != 1 || candidate.CandidateState != workflowRAGPromotionStatePending || candidate.BindingRef != nil ||
		validateStoredWorkflowRAGPromotionAudit(audit, ctx, true) != nil || audit.EventKind != "promotion_candidate_created" || audit.CandidateID != candidate.CandidateID || audit.CandidateDigest != candidate.CandidateDigest ||
		audit.RecordVersion != 1 || audit.CandidateState != workflowRAGPromotionStatePending || audit.BindingRef != nil || audit.ActorRef != candidate.CreatedByActorRef ||
		audit.OccurredAt != candidate.CreatedAt || audit.RequestID != candidate.RequestID || audit.AuditRef != candidate.AuditRef ||
		candidate.UpdatedAt != candidate.CreatedAt || candidate.UpdatedByActorRef != candidate.CreatedByActorRef {
		return errWorkflowRAGPromotionStoreContract
	}
	repository.ownerLock.Lock()
	defer repository.ownerLock.Unlock()
	if len(repository.entries) >= repository.capacity {
		return errWorkflowRAGPromotionStore
	}
	key := workflowRAGPromotionStoreKey(ctx, candidate.CandidateID)
	if _, exists := repository.entries[key]; exists {
		return errWorkflowRAGPromotionStoreContract
	}
	for _, entry := range repository.entries {
		if entry.candidate.CandidateID == candidate.CandidateID {
			return errWorkflowRAGPromotionStoreContract
		}
	}
	entry := workflowRAGPromotionMemoryEntry{candidate: cloneWorkflowRAGPromotionCandidate(candidate), decisions: []WorkflowRAGKnowledgePromotionDecision{}, audits: []WorkflowRAGPromotionAudit{cloneWorkflowRAGPromotionAudit(audit)}}
	if validateWorkflowRAGPromotionMemoryEntry(entry) != nil {
		return errWorkflowRAGPromotionStoreContract
	}
	repository.entries[key] = entry
	return nil
}

func (repository *memoryWorkflowRAGPromotionRepository) Read(ctx WorkflowRAGPromotionContext, candidateID string) (WorkflowRAGKnowledgePromotionCandidate, []WorkflowRAGKnowledgePromotionDecision, *WorkflowRAGApplicationBinding, []WorkflowRAGPromotionAudit, error) {
	if !repository.available {
		return WorkflowRAGKnowledgePromotionCandidate{}, nil, nil, nil, errWorkflowRAGPromotionStore
	}
	repository.ownerLock.RLock()
	defer repository.ownerLock.RUnlock()
	entry, err := repository.readEntryLocked(ctx, candidateID)
	if err != nil {
		return WorkflowRAGKnowledgePromotionCandidate{}, nil, nil, nil, err
	}
	if validateWorkflowRAGPromotionMemoryEntry(entry) != nil {
		return WorkflowRAGKnowledgePromotionCandidate{}, nil, nil, nil, errWorkflowRAGPromotionStoreContract
	}
	cloned := cloneWorkflowRAGPromotionMemoryEntry(entry)
	return cloned.candidate, cloned.decisions, cloned.binding, cloned.audits, nil
}

func (repository *memoryWorkflowRAGPromotionRepository) List(ctx WorkflowRAGPromotionContext, query workflowRAGPromotionListQuery) ([]WorkflowRAGKnowledgePromotionCandidate, error) {
	if !repository.available {
		return nil, errWorkflowRAGPromotionStore
	}
	repository.ownerLock.RLock()
	defer repository.ownerLock.RUnlock()
	candidates := make([]WorkflowRAGKnowledgePromotionCandidate, 0)
	for _, entry := range repository.entries {
		candidate := entry.candidate
		if candidate.TenantRef != ctx.TenantRef || candidate.WorkspaceID != ctx.WorkspaceID || candidate.ApplicationID != ctx.ApplicationID || candidate.OwnerSubjectRef != ctx.OwnerSubjectRef {
			continue
		}
		if validateWorkflowRAGPromotionMemoryEntry(entry) != nil {
			return nil, errWorkflowRAGPromotionStoreContract
		}
		if query.BeforeCreatedAt != "" && (candidate.CreatedAt > query.BeforeCreatedAt || (candidate.CreatedAt == query.BeforeCreatedAt && candidate.CandidateID >= query.BeforeCandidateID)) {
			continue
		}
		candidates = append(candidates, cloneWorkflowRAGPromotionCandidate(candidate))
	}
	sort.Slice(candidates, func(left, right int) bool {
		if candidates[left].CreatedAt == candidates[right].CreatedAt {
			return candidates[left].CandidateID > candidates[right].CandidateID
		}
		return candidates[left].CreatedAt > candidates[right].CreatedAt
	})
	if len(candidates) > query.Limit {
		candidates = candidates[:query.Limit]
	}
	return candidates, nil
}

func (repository *memoryWorkflowRAGPromotionRepository) AppendDecision(
	ctx WorkflowRAGPromotionContext,
	candidateID string,
	expectedVersion int,
	updated WorkflowRAGKnowledgePromotionCandidate,
	decision WorkflowRAGKnowledgePromotionDecision,
	binding *WorkflowRAGApplicationBinding,
	audits []WorkflowRAGPromotionAudit,
) error {
	if !repository.available {
		return errWorkflowRAGPromotionStore
	}
	if validateStoredWorkflowRAGPromotionCandidate(updated, ctx) != nil || validateStoredWorkflowRAGPromotionDecision(decision) != nil || len(audits) < 1 || len(audits) > 2 {
		return errWorkflowRAGPromotionStoreContract
	}
	for _, audit := range audits {
		if validateStoredWorkflowRAGPromotionAudit(audit, ctx, true) != nil {
			return errWorkflowRAGPromotionStoreContract
		}
	}
	if binding != nil && validateStoredWorkflowRAGApplicationBinding(*binding) != nil {
		return errWorkflowRAGPromotionStoreContract
	}
	repository.ownerLock.Lock()
	defer repository.ownerLock.Unlock()
	key := workflowRAGPromotionStoreKey(ctx, candidateID)
	entry, err := repository.readEntryLocked(ctx, candidateID)
	if err != nil {
		return err
	}
	current := entry.candidate
	if current.RecordVersion != expectedVersion {
		return workflowRAGPromotionConflictError{CurrentVersion: current.RecordVersion, CurrentState: current.CandidateState}
	}
	nextState, allowed := workflowRAGPromotionNextState(current.CandidateState, decision.Decision)
	if !allowed {
		return errWorkflowRAGPromotionTransition
	}
	if updated.CandidateID != current.CandidateID || updated.CandidateDigest != current.CandidateDigest || updated.TenantRef != current.TenantRef ||
		updated.WorkspaceID != current.WorkspaceID || updated.ApplicationID != current.ApplicationID || updated.OwnerSubjectRef != current.OwnerSubjectRef ||
		updated.Evidence != current.Evidence || updated.CreatedAt != current.CreatedAt || updated.CreatedByActorRef != current.CreatedByActorRef ||
		updated.RequestID != current.RequestID || updated.AuditRef != current.AuditRef ||
		updated.UpdatedAt != decision.OccurredAt || updated.UpdatedByActorRef != decision.ActorRef ||
		updated.RecordVersion != expectedVersion+1 || updated.CandidateState != nextState || decision.CandidateID != current.CandidateID ||
		decision.CandidateDigest != current.CandidateDigest || decision.FromState != current.CandidateState || decision.ToState != nextState ||
		decision.BeforeRecordVersion != expectedVersion || decision.AfterRecordVersion != expectedVersion+1 {
		return errWorkflowRAGPromotionStoreContract
	}
	if decision.Decision == workflowRAGPromotionDecisionApprove {
		if entry.binding != nil || binding == nil || updated.BindingRef == nil || *updated.BindingRef != binding.WorkflowRAGApplicationBindingRef ||
			binding.CandidateID != current.CandidateID || binding.CandidateDigest != current.CandidateDigest || binding.Evidence != current.Evidence ||
			binding.TenantRef != current.TenantRef || binding.WorkspaceID != current.WorkspaceID || binding.ApplicationID != current.ApplicationID ||
			binding.OwnerSubjectRef != current.OwnerSubjectRef || binding.ApprovedDecisionID != decision.DecisionID ||
			binding.ApprovedRecordVersion != decision.AfterRecordVersion || binding.IssuedAt != decision.OccurredAt ||
			binding.IssuedByActorRef != decision.ActorRef || binding.RequestID != decision.RequestID || binding.AuditRef != decision.AuditRef || len(audits) != 2 {
			return errWorkflowRAGPromotionStoreContract
		}
	} else if binding != nil || !workflowRAGPromotionBindingRefsEqual(updated.BindingRef, current.BindingRef) || len(audits) != 1 {
		return errWorkflowRAGPromotionStoreContract
	}
	if audits[0].EventKind != "promotion_decision_"+decision.Decision || audits[0].CandidateID != updated.CandidateID || audits[0].RecordVersion != updated.RecordVersion || audits[0].CandidateState != updated.CandidateState ||
		!workflowRAGPromotionBindingRefsEqual(audits[0].BindingRef, updated.BindingRef) {
		return errWorkflowRAGPromotionStoreContract
	}
	if binding != nil && (audits[1].EventKind != "promotion_binding_issued" || !workflowRAGPromotionBindingRefsEqual(audits[1].BindingRef, updated.BindingRef)) {
		return errWorkflowRAGPromotionStoreContract
	}
	entry.candidate = cloneWorkflowRAGPromotionCandidate(updated)
	entry.decisions = append(entry.decisions, decision)
	if binding != nil {
		entry.binding = cloneWorkflowRAGApplicationBinding(binding)
	}
	entry.audits = append(entry.audits, audits...)
	if validateWorkflowRAGPromotionMemoryEntry(entry) != nil {
		return errWorkflowRAGPromotionStoreContract
	}
	repository.entries[key] = cloneWorkflowRAGPromotionMemoryEntry(entry)
	return nil
}

func (repository *memoryWorkflowRAGPromotionRepository) readEntryLocked(ctx WorkflowRAGPromotionContext, candidateID string) (workflowRAGPromotionMemoryEntry, error) {
	entry, found := repository.entries[workflowRAGPromotionStoreKey(ctx, candidateID)]
	if found {
		return entry, nil
	}
	for _, candidate := range repository.entries {
		if candidate.candidate.CandidateID == candidateID {
			return workflowRAGPromotionMemoryEntry{}, errWorkflowRAGPromotionScopeDenied
		}
	}
	return workflowRAGPromotionMemoryEntry{}, errWorkflowRAGPromotionNotFound
}

func validateWorkflowRAGPromotionMemoryEntry(entry workflowRAGPromotionMemoryEntry) error {
	candidate := entry.candidate
	ctx := WorkflowRAGPromotionContext{TenantRef: candidate.TenantRef, WorkspaceID: candidate.WorkspaceID, ApplicationID: candidate.ApplicationID, OwnerSubjectRef: candidate.OwnerSubjectRef}
	if validateStoredWorkflowRAGPromotionCandidate(candidate, ctx) != nil || len(entry.audits) < 1 || entry.audits[0].EventKind != "promotion_candidate_created" ||
		entry.audits[0].CandidateID != candidate.CandidateID || entry.audits[0].CandidateDigest != candidate.CandidateDigest || entry.audits[0].RecordVersion != 1 ||
		entry.audits[0].CandidateState != workflowRAGPromotionStatePending || entry.audits[0].BindingRef != nil ||
		entry.audits[0].ActorRef != candidate.CreatedByActorRef || entry.audits[0].OccurredAt != candidate.CreatedAt ||
		entry.audits[0].RequestID != candidate.RequestID || entry.audits[0].AuditRef != candidate.AuditRef {
		return errWorkflowRAGPromotionStoreContract
	}
	state, version, auditIndex := workflowRAGPromotionStatePending, 1, 1
	approved := false
	decisionIDs := make(map[string]bool, len(entry.decisions))
	auditIDs := map[string]bool{entry.audits[0].EventID: true}
	var approvedDecision *WorkflowRAGKnowledgePromotionDecision
	var lastDecision *WorkflowRAGKnowledgePromotionDecision
	for _, decision := range entry.decisions {
		if validateStoredWorkflowRAGPromotionDecision(decision) != nil || decision.CandidateID != candidate.CandidateID || decision.CandidateDigest != candidate.CandidateDigest ||
			decision.FromState != state || decision.BeforeRecordVersion != version || decision.AfterRecordVersion != version+1 || decisionIDs[decision.DecisionID] {
			return errWorkflowRAGPromotionStoreContract
		}
		decisionIDs[decision.DecisionID] = true
		next, allowed := workflowRAGPromotionNextState(state, decision.Decision)
		if !allowed || decision.ToState != next || auditIndex >= len(entry.audits) {
			return errWorkflowRAGPromotionStoreContract
		}
		decisionAudit := entry.audits[auditIndex]
		var expectedBindingRef *WorkflowRAGApplicationBindingRef
		if approved || decision.Decision == workflowRAGPromotionDecisionApprove {
			if entry.binding == nil {
				return errWorkflowRAGPromotionStoreContract
			}
			expectedBindingRef = &entry.binding.WorkflowRAGApplicationBindingRef
		}
		if validateStoredWorkflowRAGPromotionAudit(decisionAudit, ctx, false) != nil || decisionAudit.EventKind != "promotion_decision_"+decision.Decision ||
			decisionAudit.CandidateID != decision.CandidateID || decisionAudit.CandidateDigest != decision.CandidateDigest ||
			decisionAudit.RecordVersion != decision.AfterRecordVersion || decisionAudit.CandidateState != decision.ToState ||
			!workflowRAGPromotionBindingRefsEqual(decisionAudit.BindingRef, expectedBindingRef) ||
			decisionAudit.ActorRef != decision.ActorRef || decisionAudit.OccurredAt != decision.OccurredAt || decisionAudit.RequestID != decision.RequestID ||
			decisionAudit.AuditRef != decision.AuditRef || auditIDs[decisionAudit.EventID] {
			return errWorkflowRAGPromotionStoreContract
		}
		auditIDs[decisionAudit.EventID] = true
		auditIndex++
		if decision.Decision == workflowRAGPromotionDecisionApprove {
			if approved || entry.binding == nil || auditIndex >= len(entry.audits) {
				return errWorkflowRAGPromotionStoreContract
			}
			bindingAudit := entry.audits[auditIndex]
			if validateStoredWorkflowRAGPromotionAudit(bindingAudit, ctx, false) != nil || bindingAudit.EventKind != "promotion_binding_issued" ||
				bindingAudit.CandidateID != decision.CandidateID || bindingAudit.CandidateDigest != decision.CandidateDigest ||
				bindingAudit.RecordVersion != decision.AfterRecordVersion || bindingAudit.CandidateState != decision.ToState ||
				!workflowRAGPromotionBindingRefsEqual(bindingAudit.BindingRef, &entry.binding.WorkflowRAGApplicationBindingRef) ||
				bindingAudit.ActorRef != decision.ActorRef || bindingAudit.OccurredAt != decision.OccurredAt || bindingAudit.RequestID != decision.RequestID ||
				bindingAudit.AuditRef != decision.AuditRef || auditIDs[bindingAudit.EventID] {
				return errWorkflowRAGPromotionStoreContract
			}
			auditIDs[bindingAudit.EventID] = true
			approved = true
			approval := decision
			approvedDecision = &approval
			auditIndex++
		}
		currentDecision := decision
		lastDecision = &currentDecision
		state, version = next, version+1
	}
	if auditIndex != len(entry.audits) || state != candidate.CandidateState || version != candidate.RecordVersion {
		return errWorkflowRAGPromotionStoreContract
	}
	if lastDecision == nil {
		if candidate.UpdatedAt != candidate.CreatedAt || candidate.UpdatedByActorRef != candidate.CreatedByActorRef {
			return errWorkflowRAGPromotionStoreContract
		}
	} else if candidate.UpdatedAt != lastDecision.OccurredAt || candidate.UpdatedByActorRef != lastDecision.ActorRef {
		return errWorkflowRAGPromotionStoreContract
	}
	if approved {
		if entry.binding == nil || validateStoredWorkflowRAGApplicationBinding(*entry.binding) != nil || candidate.BindingRef == nil ||
			*candidate.BindingRef != entry.binding.WorkflowRAGApplicationBindingRef || entry.binding.CandidateID != candidate.CandidateID || entry.binding.CandidateDigest != candidate.CandidateDigest ||
			entry.binding.TenantRef != candidate.TenantRef || entry.binding.WorkspaceID != candidate.WorkspaceID || entry.binding.ApplicationID != candidate.ApplicationID ||
			entry.binding.OwnerSubjectRef != candidate.OwnerSubjectRef || entry.binding.Evidence != candidate.Evidence || approvedDecision == nil ||
			entry.binding.ApprovedDecisionID != approvedDecision.DecisionID || entry.binding.ApprovedRecordVersion != approvedDecision.AfterRecordVersion ||
			entry.binding.IssuedAt != approvedDecision.OccurredAt || entry.binding.IssuedByActorRef != approvedDecision.ActorRef ||
			entry.binding.RequestID != approvedDecision.RequestID || entry.binding.AuditRef != approvedDecision.AuditRef {
			return errWorkflowRAGPromotionStoreContract
		}
	} else if entry.binding != nil || candidate.BindingRef != nil {
		return errWorkflowRAGPromotionStoreContract
	}
	return nil
}

func validateWorkflowRAGPromotionContext(ctx WorkflowRAGPromotionContext) error {
	if !workflowRAGReferencePattern.MatchString(strings.TrimSpace(ctx.TenantRef)) || !workflowRAGScopedIDPattern.MatchString(strings.TrimSpace(ctx.WorkspaceID)) ||
		!workflowRAGScopedIDPattern.MatchString(strings.TrimSpace(ctx.ApplicationID)) || !workflowRAGReferencePattern.MatchString(strings.TrimSpace(ctx.ActorRef)) ||
		!workflowRAGReferencePattern.MatchString(strings.TrimSpace(ctx.OwnerSubjectRef)) || !workflowRAGReferencePattern.MatchString(strings.TrimSpace(ctx.RequestID)) ||
		!workflowRAGReferencePattern.MatchString(strings.TrimSpace(ctx.AuditRef)) || workflowRAGContainsForbiddenMaterial(strings.Join([]string{
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.ActorRef, ctx.OwnerSubjectRef, ctx.RequestID, ctx.AuditRef,
	}, "\n")) {
		return errWorkflowRAGPromotionScopeDenied
	}
	return nil
}

func validateStoredWorkflowRAGPromotionCandidate(candidate WorkflowRAGKnowledgePromotionCandidate, ctx WorkflowRAGPromotionContext) error {
	if candidate.SchemaVersion != workflowRAGPromotionCandidateSchemaVersion || !workflowRAGPromotionCandidateIDPattern.MatchString(candidate.CandidateID) ||
		candidate.TenantRef != ctx.TenantRef || candidate.WorkspaceID != ctx.WorkspaceID || candidate.ApplicationID != ctx.ApplicationID || candidate.OwnerSubjectRef != ctx.OwnerSubjectRef ||
		!workflowRAGReferencePattern.MatchString(candidate.CreatedByActorRef) || !workflowRAGReferencePattern.MatchString(candidate.UpdatedByActorRef) ||
		!workflowRAGReferencePattern.MatchString(candidate.RequestID) || !workflowRAGReferencePattern.MatchString(candidate.AuditRef) ||
		candidate.RecordVersion < 1 || !workflowRAGPromotionStateAllowed(candidate.CandidateState) || validateWorkflowRAGPromotionEvidence(candidate.Evidence, candidate.TenantRef, candidate.WorkspaceID, candidate.ApplicationID) != nil ||
		workflowRAGContainsForbiddenMaterial(strings.Join([]string{
			candidate.CandidateID, candidate.CandidateDigest, candidate.TenantRef, candidate.WorkspaceID, candidate.ApplicationID, candidate.OwnerSubjectRef,
			candidate.CreatedByActorRef, candidate.UpdatedByActorRef, candidate.RequestID, candidate.AuditRef,
		}, "\n")) {
		return errWorkflowRAGPromotionStoreContract
	}
	created, err := time.Parse(time.RFC3339Nano, candidate.CreatedAt)
	if err != nil {
		return errWorkflowRAGPromotionStoreContract
	}
	updated, err := time.Parse(time.RFC3339Nano, candidate.UpdatedAt)
	if err != nil || updated.Before(created) {
		return errWorkflowRAGPromotionStoreContract
	}
	digest, err := workflowRAGPromotionCandidateDigest(candidate)
	if err != nil || candidate.CandidateDigest != digest {
		return errWorkflowRAGPromotionStoreContract
	}
	if candidate.BindingRef != nil && validateWorkflowRAGApplicationBindingRef(*candidate.BindingRef) != nil {
		return errWorkflowRAGPromotionStoreContract
	}
	return nil
}

func validateWorkflowRAGPromotionEvidence(evidence WorkflowRAGPromotionEvidenceBinding, tenantRef, workspaceID, applicationID string) error {
	if !workflowRAGDatasetIDPattern.MatchString(evidence.Dataset.DatasetID) || evidence.Dataset.DatasetVersion < 1 || !workflowRAGDigestPattern.MatchString(evidence.Dataset.DatasetDigest) ||
		!workflowRAGEvaluationReviewIDPattern.MatchString(evidence.CandidateReviewID) || !validateWorkflowRAGEvaluationSnapshotBinding(evidence.BaselineSnapshot) ||
		!validateWorkflowRAGEvaluationSnapshotBinding(evidence.CandidateSnapshot) || evidence.BaselineSnapshot.TenantRef != tenantRef || evidence.CandidateSnapshot.TenantRef != tenantRef ||
		evidence.BaselineSnapshot.WorkspaceID != workspaceID || evidence.CandidateSnapshot.WorkspaceID != workspaceID ||
		evidence.BaselineSnapshot.ApplicationID != applicationID || evidence.CandidateSnapshot.ApplicationID != applicationID ||
		!workflowRAGReferencePattern.MatchString(evidence.Profile.ProfileID) || evidence.Profile.ProfileVersion < 1 || !workflowRAGDigestPattern.MatchString(evidence.Profile.ProfileDigest) ||
		!applicationDraftIdentifierPattern.MatchString(evidence.SourceDraft.DraftID) || evidence.SourceDraft.DraftVersion < 1 || !workflowRAGDigestPattern.MatchString(evidence.SourceDraft.DraftDigest) ||
		workflowRAGPromotionEvidenceContainsForbiddenMaterial(evidence) {
		return errWorkflowRAGPromotionStoreContract
	}
	if _, err := time.Parse(time.RFC3339Nano, evidence.SourceDraft.BaseApplicationUpdatedAt); err != nil {
		return errWorkflowRAGPromotionStoreContract
	}
	return nil
}

func validateStoredWorkflowRAGPromotionDecision(decision WorkflowRAGKnowledgePromotionDecision) error {
	if decision.SchemaVersion != workflowRAGPromotionDecisionSchemaVersion || !workflowRAGPromotionDecisionIDPattern.MatchString(decision.DecisionID) ||
		!workflowRAGPromotionCandidateIDPattern.MatchString(decision.CandidateID) || !workflowRAGDigestPattern.MatchString(decision.CandidateDigest) ||
		!workflowRAGPromotionDecisionAllowed(decision.Decision) || !workflowRAGPromotionStateAllowed(decision.FromState) || !workflowRAGPromotionStateAllowed(decision.ToState) ||
		decision.BeforeRecordVersion < 1 || decision.AfterRecordVersion != decision.BeforeRecordVersion+1 || !utf8.ValidString(decision.Reason) || strings.TrimSpace(decision.Reason) != decision.Reason ||
		len([]rune(decision.Reason)) < 4 || len([]rune(decision.Reason)) > 500 || workflowRAGContainsForbiddenMaterial(decision.Reason) ||
		!workflowRAGReferencePattern.MatchString(decision.ActorRef) || !workflowRAGReferencePattern.MatchString(decision.RequestID) || !workflowRAGReferencePattern.MatchString(decision.AuditRef) ||
		workflowRAGContainsForbiddenMaterial(strings.Join([]string{
			decision.DecisionID, decision.CandidateID, decision.CandidateDigest, decision.ActorRef, decision.RequestID, decision.AuditRef,
		}, "\n")) {
		return errWorkflowRAGPromotionStoreContract
	}
	next, allowed := workflowRAGPromotionNextState(decision.FromState, decision.Decision)
	if !allowed || next != decision.ToState {
		return errWorkflowRAGPromotionStoreContract
	}
	if _, err := time.Parse(time.RFC3339Nano, decision.OccurredAt); err != nil {
		return errWorkflowRAGPromotionStoreContract
	}
	return nil
}

func validateStoredWorkflowRAGApplicationBinding(binding WorkflowRAGApplicationBinding) error {
	if binding.SchemaVersion != workflowRAGApplicationBindingSchemaVersion || validateWorkflowRAGApplicationBindingRef(binding.WorkflowRAGApplicationBindingRef) != nil ||
		!workflowRAGPromotionCandidateIDPattern.MatchString(binding.CandidateID) || !workflowRAGDigestPattern.MatchString(binding.CandidateDigest) ||
		!workflowRAGPromotionDecisionIDPattern.MatchString(binding.ApprovedDecisionID) || binding.ApprovedRecordVersion < 2 ||
		!workflowRAGReferencePattern.MatchString(binding.TenantRef) || !workflowRAGScopedIDPattern.MatchString(binding.WorkspaceID) || !workflowRAGScopedIDPattern.MatchString(binding.ApplicationID) ||
		!workflowRAGReferencePattern.MatchString(binding.OwnerSubjectRef) || validateWorkflowRAGPromotionEvidence(binding.Evidence, binding.TenantRef, binding.WorkspaceID, binding.ApplicationID) != nil ||
		!workflowRAGReferencePattern.MatchString(binding.IssuedByActorRef) || !workflowRAGReferencePattern.MatchString(binding.RequestID) || !workflowRAGReferencePattern.MatchString(binding.AuditRef) ||
		workflowRAGContainsForbiddenMaterial(strings.Join([]string{
			binding.BindingID, binding.BindingDigest, binding.CandidateID, binding.CandidateDigest, binding.ApprovedDecisionID,
			binding.TenantRef, binding.WorkspaceID, binding.ApplicationID, binding.OwnerSubjectRef, binding.IssuedByActorRef, binding.RequestID, binding.AuditRef,
		}, "\n")) {
		return errWorkflowRAGPromotionStoreContract
	}
	if _, err := time.Parse(time.RFC3339Nano, binding.IssuedAt); err != nil {
		return errWorkflowRAGPromotionStoreContract
	}
	digest, err := workflowRAGApplicationBindingDigest(binding)
	if err != nil || digest != binding.BindingDigest {
		return errWorkflowRAGPromotionStoreContract
	}
	return nil
}

func validateWorkflowRAGApplicationBindingRef(ref WorkflowRAGApplicationBindingRef) error {
	if !workflowRAGPromotionBindingIDPattern.MatchString(ref.BindingID) || ref.BindingVersion != 1 || !workflowRAGDigestPattern.MatchString(ref.BindingDigest) {
		return errWorkflowRAGPromotionStoreContract
	}
	return nil
}

func validateStoredWorkflowRAGPromotionAudit(audit WorkflowRAGPromotionAudit, ctx WorkflowRAGPromotionContext, requireActor bool) error {
	if audit.SchemaVersion != workflowRAGPromotionAuditSchemaVersion || !workflowRAGPromotionAuditIDPattern.MatchString(audit.EventID) || !workflowRAGPromotionAuditKindAllowed(audit.EventKind) ||
		!workflowRAGPromotionCandidateIDPattern.MatchString(audit.CandidateID) || !workflowRAGDigestPattern.MatchString(audit.CandidateDigest) || !workflowRAGPromotionStateAllowed(audit.CandidateState) || audit.RecordVersion < 1 ||
		audit.TenantRef != ctx.TenantRef || audit.WorkspaceID != ctx.WorkspaceID || audit.ApplicationID != ctx.ApplicationID || audit.OwnerSubjectRef != ctx.OwnerSubjectRef ||
		!workflowRAGReferencePattern.MatchString(audit.ActorRef) || !workflowRAGReferencePattern.MatchString(audit.RequestID) || !workflowRAGReferencePattern.MatchString(audit.AuditRef) ||
		workflowRAGContainsForbiddenMaterial(strings.Join([]string{
			audit.EventID, audit.EventKind, audit.CandidateID, audit.CandidateDigest, audit.TenantRef, audit.WorkspaceID,
			audit.ApplicationID, audit.OwnerSubjectRef, audit.ActorRef, audit.RequestID, audit.AuditRef,
		}, "\n")) {
		return errWorkflowRAGPromotionStoreContract
	}
	if requireActor && (audit.ActorRef != ctx.ActorRef || audit.RequestID != ctx.RequestID || audit.AuditRef != ctx.AuditRef) {
		return errWorkflowRAGPromotionStoreContract
	}
	if audit.BindingRef != nil && validateWorkflowRAGApplicationBindingRef(*audit.BindingRef) != nil {
		return errWorkflowRAGPromotionStoreContract
	}
	if _, err := time.Parse(time.RFC3339Nano, audit.OccurredAt); err != nil {
		return errWorkflowRAGPromotionStoreContract
	}
	return nil
}

func workflowRAGPromotionRepositoryFailure(err error) WorkflowRAGPromotionResult {
	result := workflowRAGPromotionFailure(WorkflowRAGPromotionFailureStoreUnavailable)
	switch {
	case errors.Is(err, errWorkflowRAGPromotionScopeDenied):
		result.FailureCode = WorkflowRAGPromotionFailureScopeDenied
	case errors.Is(err, errWorkflowRAGPromotionNotFound):
		result.FailureCode = WorkflowRAGPromotionFailureNotFound
	case errors.Is(err, errWorkflowRAGPromotionConflict):
		result.FailureCode = WorkflowRAGPromotionFailureRecordConflict
	case errors.Is(err, errWorkflowRAGPromotionTransition):
		result.FailureCode = WorkflowRAGPromotionFailureTransitionInvalid
	case errors.Is(err, errWorkflowRAGPromotionStoreContract):
		result.FailureCode = WorkflowRAGPromotionFailureStoreContractMismatch
	}
	var conflict workflowRAGPromotionConflictError
	if errors.As(err, &conflict) {
		result.CurrentRecordVersion, result.CurrentState = conflict.CurrentVersion, conflict.CurrentState
	}
	return result
}

func workflowRAGPromotionStateAllowed(state string) bool {
	return state == workflowRAGPromotionStatePending || state == workflowRAGPromotionStateDeferred || state == workflowRAGPromotionStateApproved || state == workflowRAGPromotionStateRejected || state == workflowRAGPromotionStateCanceled
}

func workflowRAGPromotionAuditKindAllowed(kind string) bool {
	return kind == "promotion_candidate_created" || kind == "promotion_binding_issued" || strings.HasPrefix(kind, "promotion_decision_") && workflowRAGPromotionDecisionAllowed(strings.TrimPrefix(kind, "promotion_decision_"))
}

func workflowRAGPromotionStoreKey(ctx WorkflowRAGPromotionContext, candidateID string) string {
	return strings.Join([]string{ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, candidateID}, "\x00")
}

func workflowRAGPromotionBindingRefsEqual(left, right *WorkflowRAGApplicationBindingRef) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func cloneWorkflowRAGPromotionCandidate(candidate WorkflowRAGKnowledgePromotionCandidate) WorkflowRAGKnowledgePromotionCandidate {
	candidate.BindingRef = cloneWorkflowRAGApplicationBindingRef(candidate.BindingRef)
	return candidate
}

func cloneWorkflowRAGPromotionDecisions(decisions []WorkflowRAGKnowledgePromotionDecision) []WorkflowRAGKnowledgePromotionDecision {
	return append([]WorkflowRAGKnowledgePromotionDecision(nil), decisions...)
}

func cloneWorkflowRAGApplicationBinding(binding *WorkflowRAGApplicationBinding) *WorkflowRAGApplicationBinding {
	if binding == nil {
		return nil
	}
	cloned := *binding
	return &cloned
}

func cloneWorkflowRAGApplicationBindingRef(ref *WorkflowRAGApplicationBindingRef) *WorkflowRAGApplicationBindingRef {
	if ref == nil {
		return nil
	}
	cloned := *ref
	return &cloned
}

func cloneWorkflowRAGPromotionAudit(audit WorkflowRAGPromotionAudit) WorkflowRAGPromotionAudit {
	audit.BindingRef = cloneWorkflowRAGApplicationBindingRef(audit.BindingRef)
	return audit
}

func cloneWorkflowRAGPromotionMemoryEntry(entry workflowRAGPromotionMemoryEntry) workflowRAGPromotionMemoryEntry {
	cloned := workflowRAGPromotionMemoryEntry{
		candidate: cloneWorkflowRAGPromotionCandidate(entry.candidate), decisions: cloneWorkflowRAGPromotionDecisions(entry.decisions),
		binding: cloneWorkflowRAGApplicationBinding(entry.binding), audits: make([]WorkflowRAGPromotionAudit, len(entry.audits)),
	}
	for index, audit := range entry.audits {
		cloned.audits[index] = cloneWorkflowRAGPromotionAudit(audit)
	}
	return cloned
}

func encodeWorkflowRAGPromotionCandidate(candidate WorkflowRAGKnowledgePromotionCandidate) ([]byte, error) {
	ctx := WorkflowRAGPromotionContext{TenantRef: candidate.TenantRef, WorkspaceID: candidate.WorkspaceID, ApplicationID: candidate.ApplicationID, OwnerSubjectRef: candidate.OwnerSubjectRef}
	if validateStoredWorkflowRAGPromotionCandidate(candidate, ctx) != nil {
		return nil, errWorkflowRAGPromotionStoreContract
	}
	return json.Marshal(candidate)
}

func decodeWorkflowRAGPromotionCandidate(payload []byte) (WorkflowRAGKnowledgePromotionCandidate, error) {
	var candidate WorkflowRAGKnowledgePromotionCandidate
	if decodeWorkflowRAGStrictJSON(payload, &candidate) != nil {
		return WorkflowRAGKnowledgePromotionCandidate{}, errWorkflowRAGPromotionStoreContract
	}
	if _, err := encodeWorkflowRAGPromotionCandidate(candidate); err != nil {
		return WorkflowRAGKnowledgePromotionCandidate{}, err
	}
	return candidate, nil
}

func encodeWorkflowRAGPromotionDecision(decision WorkflowRAGKnowledgePromotionDecision) ([]byte, error) {
	if validateStoredWorkflowRAGPromotionDecision(decision) != nil {
		return nil, errWorkflowRAGPromotionStoreContract
	}
	return json.Marshal(decision)
}

func decodeWorkflowRAGPromotionDecision(payload []byte) (WorkflowRAGKnowledgePromotionDecision, error) {
	var decision WorkflowRAGKnowledgePromotionDecision
	if decodeWorkflowRAGStrictJSON(payload, &decision) != nil || validateStoredWorkflowRAGPromotionDecision(decision) != nil {
		return WorkflowRAGKnowledgePromotionDecision{}, errWorkflowRAGPromotionStoreContract
	}
	return decision, nil
}

func encodeWorkflowRAGApplicationBinding(binding WorkflowRAGApplicationBinding) ([]byte, error) {
	if validateStoredWorkflowRAGApplicationBinding(binding) != nil {
		return nil, errWorkflowRAGPromotionStoreContract
	}
	return json.Marshal(binding)
}

func decodeWorkflowRAGApplicationBinding(payload []byte) (WorkflowRAGApplicationBinding, error) {
	var binding WorkflowRAGApplicationBinding
	if decodeWorkflowRAGStrictJSON(payload, &binding) != nil || validateStoredWorkflowRAGApplicationBinding(binding) != nil {
		return WorkflowRAGApplicationBinding{}, errWorkflowRAGPromotionStoreContract
	}
	return binding, nil
}

func encodeWorkflowRAGPromotionAudit(audit WorkflowRAGPromotionAudit) ([]byte, error) {
	ctx := WorkflowRAGPromotionContext{TenantRef: audit.TenantRef, WorkspaceID: audit.WorkspaceID, ApplicationID: audit.ApplicationID, OwnerSubjectRef: audit.OwnerSubjectRef}
	if validateStoredWorkflowRAGPromotionAudit(audit, ctx, false) != nil {
		return nil, errWorkflowRAGPromotionStoreContract
	}
	return json.Marshal(audit)
}

func decodeWorkflowRAGPromotionAudit(payload []byte) (WorkflowRAGPromotionAudit, error) {
	var audit WorkflowRAGPromotionAudit
	if decodeWorkflowRAGStrictJSON(payload, &audit) != nil {
		return WorkflowRAGPromotionAudit{}, errWorkflowRAGPromotionStoreContract
	}
	if _, err := encodeWorkflowRAGPromotionAudit(audit); err != nil {
		return WorkflowRAGPromotionAudit{}, err
	}
	return audit, nil
}
