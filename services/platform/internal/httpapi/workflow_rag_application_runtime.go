package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

const (
	workflowRAGApplicationRuntimeAssignmentSchemaVersion = "workflow_rag_application_runtime_assignment.v1"
	workflowRAGApplicationRuntimeEventSchemaVersion      = "workflow_rag_application_runtime_assignment_event.v1"
	workflowRAGApplicationRuntimeAuditSchemaVersion      = "workflow_rag_application_runtime_audit.v1"
	workflowRAGApplicationAnswerSchemaVersion            = "workflow_rag_application_answer.v1"

	workflowRAGApplicationRuntimeStateActive  = "active"
	workflowRAGApplicationRuntimeStateRevoked = "revoked"

	workflowRAGApplicationRuntimeDecisionActivate = "activate"
	workflowRAGApplicationRuntimeDecisionReplace  = "replace"
	workflowRAGApplicationRuntimeDecisionRevoke   = "revoke"
)

const (
	WorkflowRAGApplicationFailureAssignmentNotFound    = "workflow_rag_runtime_assignment_not_found"
	WorkflowRAGApplicationFailureAssignmentRevoked     = "workflow_rag_runtime_assignment_revoked"
	WorkflowRAGApplicationFailureVersionConflict       = "workflow_rag_runtime_assignment_version_conflict"
	WorkflowRAGApplicationFailureCandidateNotApproved  = "workflow_rag_runtime_candidate_not_approved"
	WorkflowRAGApplicationFailureCandidateSuperseded   = "workflow_rag_runtime_candidate_superseded"
	WorkflowRAGApplicationFailureConfigurationChanged  = "workflow_rag_runtime_configuration_changed"
	WorkflowRAGApplicationFailureConfigurationInvalid  = "workflow_rag_runtime_configuration_invalid"
	WorkflowRAGApplicationFailureBindingNotEligible    = "workflow_rag_runtime_binding_not_eligible"
	WorkflowRAGApplicationFailureApplicationArchived   = "workflow_rag_runtime_application_archived"
	WorkflowRAGApplicationFailureScopeDenied           = "workflow_rag_runtime_scope_denied"
	WorkflowRAGApplicationFailurePayloadInvalid        = "workflow_rag_runtime_payload_invalid"
	WorkflowRAGApplicationFailureSecretForbidden       = "workflow_rag_runtime_secret_material_forbidden"
	WorkflowRAGApplicationFailureNoEvidence            = "workflow_rag_runtime_no_evidence"
	WorkflowRAGApplicationFailureBudgetExceeded        = "workflow_rag_runtime_budget_exceeded"
	WorkflowRAGApplicationFailureGatewayFailed         = "workflow_rag_runtime_gateway_failed"
	WorkflowRAGApplicationFailureAnswerInvalid         = "workflow_rag_runtime_answer_invalid"
	WorkflowRAGApplicationFailureCitationInvalid       = "workflow_rag_runtime_citation_invalid"
	WorkflowRAGApplicationFailureStoreUnavailable      = "workflow_rag_runtime_store_unavailable"
	WorkflowRAGApplicationFailureStoreContractMismatch = "workflow_rag_runtime_store_contract_mismatch"
	WorkflowRAGApplicationFailureWriteDisabled         = "workflow_rag_runtime_write_disabled"
	WorkflowRAGApplicationFailureTransitionInvalid     = "workflow_rag_runtime_transition_invalid"
)

var (
	workflowRAGApplicationAssignmentIDPattern = regexp.MustCompile(`^wragra_[a-z2-7]{16}$`)
	workflowRAGApplicationEventIDPattern      = regexp.MustCompile(`^wragre_[a-z2-7]{16}$`)
	workflowRAGApplicationAuditIDPattern      = regexp.MustCompile(`^wragru_[a-z2-7]{16}$`)

	errWorkflowRAGApplicationAssignmentNotFound = errors.New(WorkflowRAGApplicationFailureAssignmentNotFound)
	errWorkflowRAGApplicationVersionConflict    = errors.New(WorkflowRAGApplicationFailureVersionConflict)
	errWorkflowRAGApplicationTransition         = errors.New(WorkflowRAGApplicationFailureTransitionInvalid)
	errWorkflowRAGApplicationStore              = errors.New(WorkflowRAGApplicationFailureStoreUnavailable)
	errWorkflowRAGApplicationStoreContract      = errors.New(WorkflowRAGApplicationFailureStoreContractMismatch)
)

type WorkflowRAGApplicationRuntimeContext struct {
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

type WorkflowRAGApplicationRuntimeAssignment struct {
	SchemaVersion         string                           `json:"schema_version"`
	AssignmentID          string                           `json:"assignment_id"`
	RecordVersion         int                              `json:"record_version"`
	AssignmentDigest      string                           `json:"assignment_digest"`
	TenantRef             string                           `json:"tenant_ref"`
	WorkspaceID           string                           `json:"workspace_id"`
	ApplicationID         string                           `json:"application_id"`
	OwnerSubjectRef       string                           `json:"owner_subject_ref"`
	State                 string                           `json:"state"`
	PublishCandidateID    string                           `json:"publish_candidate_id"`
	PublishReviewVersion  int                              `json:"publish_review_version"`
	PublishCandidateState string                           `json:"publish_candidate_state"`
	DraftID               string                           `json:"draft_id"`
	DraftVersion          int                              `json:"draft_version"`
	DraftDigest           string                           `json:"draft_digest"`
	BindingRef            WorkflowRAGApplicationBindingRef `json:"binding_ref"`
	CreatedAt             string                           `json:"created_at"`
	UpdatedAt             string                           `json:"updated_at"`
	CreatedByActorRef     string                           `json:"created_by_actor_ref"`
	UpdatedByActorRef     string                           `json:"updated_by_actor_ref"`
	RequestID             string                           `json:"request_id"`
	AuditRef              string                           `json:"audit_ref"`
}

type WorkflowRAGApplicationRuntimeEvent struct {
	SchemaVersion       string `json:"schema_version"`
	EventID             string `json:"event_id"`
	AssignmentID        string `json:"assignment_id"`
	Decision            string `json:"decision"`
	Reason              string `json:"reason"`
	FromState           string `json:"from_state"`
	ToState             string `json:"to_state"`
	BeforeRecordVersion int    `json:"before_record_version"`
	AfterRecordVersion  int    `json:"after_record_version"`
	PublishCandidateID  string `json:"publish_candidate_id"`
	ActorRef            string `json:"actor_ref"`
	OccurredAt          string `json:"occurred_at"`
	RequestID           string `json:"request_id"`
	AuditRef            string `json:"audit_ref"`
}

type WorkflowRAGApplicationRuntimeAudit struct {
	SchemaVersion    string `json:"schema_version"`
	AuditEventID     string `json:"audit_event_id"`
	AssignmentID     string `json:"assignment_id"`
	AssignmentDigest string `json:"assignment_digest"`
	RecordVersion    int    `json:"record_version"`
	State            string `json:"state"`
	Decision         string `json:"decision"`
	TenantRef        string `json:"tenant_ref"`
	WorkspaceID      string `json:"workspace_id"`
	ApplicationID    string `json:"application_id"`
	OwnerSubjectRef  string `json:"owner_subject_ref"`
	ActorRef         string `json:"actor_ref"`
	OccurredAt       string `json:"occurred_at"`
	RequestID        string `json:"request_id"`
	AuditRef         string `json:"audit_ref"`
}

type WorkflowRAGApplicationAnswer struct {
	SchemaVersion string                `json:"schema_version"`
	Answer        string                `json:"answer"`
	Citations     []WorkflowRAGCitation `json:"citations"`
	Limitations   []string              `json:"limitations"`
	Confidence    string                `json:"confidence"`
}

type WorkflowRAGApplicationRuntimeDecisionInput struct {
	ExpectedRecordVersion int
	Decision              string
	PublishCandidateID    string
	Reason                string
}

type WorkflowRAGApplicationRuntimeResult struct {
	Assignment           *WorkflowRAGApplicationRuntimeAssignment
	Events               []WorkflowRAGApplicationRuntimeEvent
	Audits               []WorkflowRAGApplicationRuntimeAudit
	FailureCode          string
	CurrentRecordVersion int
	CurrentState         string
}

type workflowRAGApplicationRuntimeRepository interface {
	Read(WorkflowRAGApplicationRuntimeContext) (WorkflowRAGApplicationRuntimeAssignment, []WorkflowRAGApplicationRuntimeEvent, []WorkflowRAGApplicationRuntimeAudit, error)
	Apply(WorkflowRAGApplicationRuntimeContext, int, WorkflowRAGApplicationRuntimeAssignment, WorkflowRAGApplicationRuntimeEvent, WorkflowRAGApplicationRuntimeAudit) error
}

type workflowRAGApplicationRuntimeMemoryEntry struct {
	assignment WorkflowRAGApplicationRuntimeAssignment
	events     []WorkflowRAGApplicationRuntimeEvent
	audits     []WorkflowRAGApplicationRuntimeAudit
}

type memoryWorkflowRAGApplicationRuntimeRepository struct {
	ownerLock *sync.RWMutex
	entries   map[string]workflowRAGApplicationRuntimeMemoryEntry
	available bool
}

func newMemoryWorkflowRAGApplicationRuntimeRepository(ownerLock *sync.RWMutex) *memoryWorkflowRAGApplicationRuntimeRepository {
	if ownerLock == nil {
		ownerLock = &sync.RWMutex{}
	}
	return &memoryWorkflowRAGApplicationRuntimeRepository{ownerLock: ownerLock, entries: make(map[string]workflowRAGApplicationRuntimeMemoryEntry), available: true}
}

func newWorkflowRAGApplicationRuntimeRepositoryForRunStore(store workflowRunStore) (workflowRAGApplicationRuntimeRepository, error) {
	switch typed := store.(type) {
	case *memoryWorkflowRunStore:
		return newMemoryWorkflowRAGApplicationRuntimeRepository(&typed.mu), nil
	case *sqliteWorkflowRunStore:
		if typed.database == nil {
			return nil, errors.New("workflow RAG application runtime requires the shared SQLite database")
		}
		return newSQLiteWorkflowRAGApplicationRuntimeRepository(typed.database), nil
	case *postgresWorkflowRunStore:
		if typed.pool == nil {
			return nil, errors.New("workflow RAG application runtime requires the workflow PostgreSQL pool")
		}
		return newPostgresWorkflowRAGApplicationRuntimeRepository(typed.pool), nil
	default:
		return nil, errors.New("workflow RAG application runtime requires a supported workflow runtime backend")
	}
}

func (repository *memoryWorkflowRAGApplicationRuntimeRepository) Read(ctx WorkflowRAGApplicationRuntimeContext) (WorkflowRAGApplicationRuntimeAssignment, []WorkflowRAGApplicationRuntimeEvent, []WorkflowRAGApplicationRuntimeAudit, error) {
	if repository == nil || !repository.available {
		return WorkflowRAGApplicationRuntimeAssignment{}, nil, nil, errWorkflowRAGApplicationStore
	}
	repository.ownerLock.RLock()
	defer repository.ownerLock.RUnlock()
	entry, ok := repository.entries[workflowRAGApplicationRuntimeKey(ctx)]
	if !ok {
		return WorkflowRAGApplicationRuntimeAssignment{}, nil, nil, errWorkflowRAGApplicationAssignmentNotFound
	}
	if validateWorkflowRAGApplicationRuntimeEntry(ctx, entry) != nil {
		return WorkflowRAGApplicationRuntimeAssignment{}, nil, nil, errWorkflowRAGApplicationStoreContract
	}
	return cloneWorkflowRAGApplicationAssignment(entry.assignment), append([]WorkflowRAGApplicationRuntimeEvent(nil), entry.events...), append([]WorkflowRAGApplicationRuntimeAudit(nil), entry.audits...), nil
}

func (repository *memoryWorkflowRAGApplicationRuntimeRepository) Apply(ctx WorkflowRAGApplicationRuntimeContext, expectedVersion int, assignment WorkflowRAGApplicationRuntimeAssignment, event WorkflowRAGApplicationRuntimeEvent, audit WorkflowRAGApplicationRuntimeAudit) error {
	if repository == nil || !repository.available {
		return errWorkflowRAGApplicationStore
	}
	repository.ownerLock.Lock()
	defer repository.ownerLock.Unlock()
	key := workflowRAGApplicationRuntimeKey(ctx)
	entry, exists := repository.entries[key]
	currentVersion := 0
	if exists {
		if validateWorkflowRAGApplicationRuntimeEntry(ctx, entry) != nil {
			return errWorkflowRAGApplicationStoreContract
		}
		currentVersion = entry.assignment.RecordVersion
	}
	if currentVersion != expectedVersion {
		return errWorkflowRAGApplicationVersionConflict
	}
	if validateWorkflowRAGApplicationRuntimeMutation(ctx, entry, exists, assignment, event, audit) != nil {
		return errWorkflowRAGApplicationStoreContract
	}
	entry.assignment = cloneWorkflowRAGApplicationAssignment(assignment)
	entry.events = append(entry.events, event)
	entry.audits = append(entry.audits, audit)
	repository.entries[key] = entry
	return nil
}

type workflowRAGApplicationAuthority struct {
	Application ApplicationSummary
	Candidate   ApplicationPublishCandidate
	Draft       ApplicationConfigurationDraft
	Binding     WorkflowRAGApplicationBinding
	Snapshot    WorkflowRAGSnapshotRecord
	Profile     WorkflowRAGExecutionProfile
}

type workflowRAGApplicationAuthorityResolver struct {
	publishRepository applicationPublishCandidateRepository
	draftRepository   applicationConfigurationDraftRepository
	promotionService  workflowRAGPromotionService
	readApplication   workflowRAGPromotionApplicationReader
}

func (resolver workflowRAGApplicationAuthorityResolver) Resolve(ctx WorkflowRAGApplicationRuntimeContext, candidateID string, expected *WorkflowRAGApplicationRuntimeAssignment) (workflowRAGApplicationAuthority, string) {
	promotionContext := workflowRAGApplicationPromotionContext(ctx)
	application, err := resolver.readApplication(promotionContext)
	if errors.Is(err, errApplicationCatalogArchived) {
		return workflowRAGApplicationAuthority{}, WorkflowRAGApplicationFailureApplicationArchived
	}
	if err != nil || strings.TrimSpace(application.ApplicationRef) == "" {
		return workflowRAGApplicationAuthority{}, WorkflowRAGApplicationFailureStoreUnavailable
	}
	publishContext := workflowRAGApplicationPublishContext(ctx)
	candidate, err := resolver.publishRepository.Read(publishContext, strings.TrimSpace(candidateID))
	if errors.Is(err, errApplicationPublishNotFound) {
		return workflowRAGApplicationAuthority{}, WorkflowRAGApplicationFailureCandidateNotApproved
	}
	if err != nil {
		return workflowRAGApplicationAuthority{}, WorkflowRAGApplicationFailureStoreUnavailable
	}
	if candidate.SchemaVersion != applicationPublishCandidateSchemaVersionV2 || candidate.CandidateState != applicationPublishStateApproved || candidate.ReviewVersion < 1 || candidate.Configuration.WorkflowRAGBindingRef == nil {
		return workflowRAGApplicationAuthority{}, WorkflowRAGApplicationFailureCandidateNotApproved
	}
	candidates, err := resolver.publishRepository.List(publishContext)
	if err != nil {
		return workflowRAGApplicationAuthority{}, WorkflowRAGApplicationFailureStoreUnavailable
	}
	if applicationPublishCandidateIsSuperseded(candidate, candidates) {
		return workflowRAGApplicationAuthority{}, WorkflowRAGApplicationFailureCandidateSuperseded
	}
	draft, err := resolver.draftRepository.Read(workflowRAGApplicationDraftContext(ctx), candidate.DraftID)
	if err != nil {
		return workflowRAGApplicationAuthority{}, WorkflowRAGApplicationFailureConfigurationChanged
	}
	digest, err := applicationConfigurationCanonicalDigest(applicationPublishSnapshotFromDraft(draft))
	if err != nil || draft.DraftVersion != candidate.DraftVersion || digest != candidate.DraftDigest || strings.TrimSpace(draft.BaseApplicationUpdatedAt) != strings.TrimSpace(application.UpdatedAt) {
		return workflowRAGApplicationAuthority{}, WorkflowRAGApplicationFailureConfigurationChanged
	}
	if !draft.ValidationSummary.IsValid || !validateApplicationConfigurationDraftPayload(workflowRAGApplicationDraftContext(ctx), draft.ApplicationConfigurationDraftPayload).IsValid {
		return workflowRAGApplicationAuthority{}, WorkflowRAGApplicationFailureConfigurationInvalid
	}
	if !workflowRAGPromotionBindingRefsEqual(draft.WorkflowRAGBindingRef, candidate.Configuration.WorkflowRAGBindingRef) {
		return workflowRAGApplicationAuthority{}, WorkflowRAGApplicationFailureBindingNotEligible
	}
	binding, failure := resolver.promotionService.resolveEligibleBinding(promotionContext, *candidate.Configuration.WorkflowRAGBindingRef, false)
	if failure != "" {
		if failure == WorkflowRAGPromotionFailureStoreUnavailable || failure == WorkflowRAGPromotionFailureStoreContractMismatch {
			return workflowRAGApplicationAuthority{}, WorkflowRAGApplicationFailureStoreUnavailable
		}
		return workflowRAGApplicationAuthority{}, WorkflowRAGApplicationFailureBindingNotEligible
	}
	_, snapshot, err := resolver.promotionService.snapshotRepository.ReadVersion(workflowRAGPromotionSnapshotContext(promotionContext), binding.Evidence.CandidateSnapshot.SnapshotID, binding.Evidence.CandidateSnapshot.SnapshotVersion)
	if err != nil || workflowRAGEvaluationSnapshotBinding(snapshot) != binding.Evidence.CandidateSnapshot {
		return workflowRAGApplicationAuthority{}, WorkflowRAGApplicationFailureBindingNotEligible
	}
	profile := resolver.promotionService.currentProfile()
	if binding.Evidence.Profile != (WorkflowRAGEvaluationProfileBinding{ProfileID: profile.ProfileID, ProfileVersion: profile.ProfileVersion, ProfileDigest: profile.ProfileDigest}) {
		return workflowRAGApplicationAuthority{}, WorkflowRAGApplicationFailureBindingNotEligible
	}
	if !workflowRAGApplicationProtocolAllowed(candidate.Configuration.DefaultProtocol, candidate.Configuration.AllowedProtocols) || strings.TrimSpace(candidate.Configuration.DefaultModel) == "" {
		return workflowRAGApplicationAuthority{}, WorkflowRAGApplicationFailureConfigurationInvalid
	}
	if expected != nil && (candidate.CandidateID != expected.PublishCandidateID || candidate.ReviewVersion != expected.PublishReviewVersion || candidate.CandidateState != expected.PublishCandidateState ||
		candidate.DraftID != expected.DraftID || candidate.DraftVersion != expected.DraftVersion || candidate.DraftDigest != expected.DraftDigest || binding.WorkflowRAGApplicationBindingRef != expected.BindingRef) {
		return workflowRAGApplicationAuthority{}, WorkflowRAGApplicationFailureConfigurationChanged
	}
	return workflowRAGApplicationAuthority{Application: application, Candidate: candidate, Draft: draft, Binding: binding, Snapshot: snapshot, Profile: profile}, ""
}

type workflowRAGApplicationRuntimeService struct {
	repository workflowRAGApplicationRuntimeRepository
	resolver   workflowRAGApplicationAuthorityResolver
	now        func() time.Time
	newID      func(string) (string, error)
}

func newWorkflowRAGApplicationRuntimeService(repository workflowRAGApplicationRuntimeRepository, resolver workflowRAGApplicationAuthorityResolver) workflowRAGApplicationRuntimeService {
	return workflowRAGApplicationRuntimeService{repository: repository, resolver: resolver, now: func() time.Time { return time.Now().UTC() }, newID: newWorkflowRAGStableID}
}

func (service workflowRAGApplicationRuntimeService) Read(ctx WorkflowRAGApplicationRuntimeContext) WorkflowRAGApplicationRuntimeResult {
	if validateWorkflowRAGApplicationRuntimeContext(ctx) != nil {
		return workflowRAGApplicationRuntimeFailure(WorkflowRAGApplicationFailureScopeDenied)
	}
	assignment, events, audits, err := service.repository.Read(ctx)
	if err != nil {
		return workflowRAGApplicationRuntimeRepositoryFailure(err)
	}
	return WorkflowRAGApplicationRuntimeResult{Assignment: &assignment, Events: events, Audits: audits, CurrentRecordVersion: assignment.RecordVersion, CurrentState: assignment.State}
}

func (service workflowRAGApplicationRuntimeService) Decide(ctx WorkflowRAGApplicationRuntimeContext, input WorkflowRAGApplicationRuntimeDecisionInput) WorkflowRAGApplicationRuntimeResult {
	if validateWorkflowRAGApplicationRuntimeContext(ctx) != nil {
		return workflowRAGApplicationRuntimeFailure(WorkflowRAGApplicationFailureScopeDenied)
	}
	if !ctx.WriteEnabled {
		return workflowRAGApplicationRuntimeFailure(WorkflowRAGApplicationFailureWriteDisabled)
	}
	input.Decision, input.PublishCandidateID, input.Reason = strings.TrimSpace(input.Decision), strings.TrimSpace(input.PublishCandidateID), strings.TrimSpace(input.Reason)
	if input.ExpectedRecordVersion < 0 || !workflowRAGApplicationDecisionAllowed(input.Decision) || !utf8.ValidString(input.Reason) || utf8.RuneCountInString(input.Reason) < 4 || utf8.RuneCountInString(input.Reason) > 500 ||
		(input.Decision != workflowRAGApplicationRuntimeDecisionRevoke && !applicationDraftIdentifierPattern.MatchString(input.PublishCandidateID)) || (input.Decision == workflowRAGApplicationRuntimeDecisionRevoke && input.PublishCandidateID != "") {
		return workflowRAGApplicationRuntimeFailure(WorkflowRAGApplicationFailurePayloadInvalid)
	}
	if workflowRAGContainsForbiddenMaterial(input.Reason) || applicationDraftStringContainsSecret(input.Reason) {
		return workflowRAGApplicationRuntimeFailure(WorkflowRAGApplicationFailureSecretForbidden)
	}
	current, events, audits, readErr := service.repository.Read(ctx)
	exists := readErr == nil
	if readErr != nil && !errors.Is(readErr, errWorkflowRAGApplicationAssignmentNotFound) {
		return workflowRAGApplicationRuntimeRepositoryFailure(readErr)
	}
	if (!exists && (input.Decision != workflowRAGApplicationRuntimeDecisionActivate || input.ExpectedRecordVersion != 0)) || (exists && current.RecordVersion != input.ExpectedRecordVersion) {
		result := workflowRAGApplicationRuntimeFailure(WorkflowRAGApplicationFailureVersionConflict)
		if exists {
			result.CurrentRecordVersion, result.CurrentState = current.RecordVersion, current.State
		}
		return result
	}
	if exists && input.Decision == workflowRAGApplicationRuntimeDecisionActivate || !exists && input.Decision != workflowRAGApplicationRuntimeDecisionActivate ||
		exists && input.Decision == workflowRAGApplicationRuntimeDecisionRevoke && current.State != workflowRAGApplicationRuntimeStateActive ||
		exists && input.Decision == workflowRAGApplicationRuntimeDecisionReplace && input.PublishCandidateID == current.PublishCandidateID {
		return WorkflowRAGApplicationRuntimeResult{FailureCode: WorkflowRAGApplicationFailureTransitionInvalid, CurrentRecordVersion: current.RecordVersion, CurrentState: current.State}
	}
	var authority workflowRAGApplicationAuthority
	var failure string
	if input.Decision != workflowRAGApplicationRuntimeDecisionRevoke {
		authority, failure = service.resolver.Resolve(ctx, input.PublishCandidateID, nil)
		if failure != "" {
			return workflowRAGApplicationRuntimeFailure(failure)
		}
	}
	at := service.now().UTC().Format(time.RFC3339Nano)
	assignment := current
	if !exists {
		assignmentID, err := service.newID("wragra_")
		if err != nil {
			return workflowRAGApplicationRuntimeFailure(WorkflowRAGApplicationFailureStoreUnavailable)
		}
		assignment = WorkflowRAGApplicationRuntimeAssignment{SchemaVersion: workflowRAGApplicationRuntimeAssignmentSchemaVersion, AssignmentID: assignmentID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, OwnerSubjectRef: ctx.OwnerSubjectRef, CreatedAt: at, CreatedByActorRef: ctx.ActorRef}
	}
	assignment.RecordVersion = input.ExpectedRecordVersion + 1
	assignment.State, assignment.UpdatedAt, assignment.UpdatedByActorRef, assignment.RequestID, assignment.AuditRef = workflowRAGApplicationRuntimeStateActive, at, ctx.ActorRef, ctx.RequestID, ctx.AuditRef
	if input.Decision == workflowRAGApplicationRuntimeDecisionRevoke {
		assignment.State = workflowRAGApplicationRuntimeStateRevoked
	} else {
		assignment.PublishCandidateID, assignment.PublishReviewVersion, assignment.PublishCandidateState = authority.Candidate.CandidateID, authority.Candidate.ReviewVersion, authority.Candidate.CandidateState
		assignment.DraftID, assignment.DraftVersion, assignment.DraftDigest = authority.Draft.DraftID, authority.Draft.DraftVersion, authority.Candidate.DraftDigest
		assignment.BindingRef = authority.Binding.WorkflowRAGApplicationBindingRef
	}
	digest, err := workflowRAGApplicationAssignmentDigest(assignment)
	if err != nil {
		return workflowRAGApplicationRuntimeFailure(WorkflowRAGApplicationFailureStoreContractMismatch)
	}
	assignment.AssignmentDigest = digest
	eventID, err := service.newID("wragre_")
	if err != nil {
		return workflowRAGApplicationRuntimeFailure(WorkflowRAGApplicationFailureStoreUnavailable)
	}
	auditID, err := service.newID("wragru_")
	if err != nil {
		return workflowRAGApplicationRuntimeFailure(WorkflowRAGApplicationFailureStoreUnavailable)
	}
	fromState := "none"
	if exists {
		fromState = current.State
	}
	event := WorkflowRAGApplicationRuntimeEvent{SchemaVersion: workflowRAGApplicationRuntimeEventSchemaVersion, EventID: eventID, AssignmentID: assignment.AssignmentID, Decision: input.Decision, Reason: input.Reason, FromState: fromState, ToState: assignment.State, BeforeRecordVersion: input.ExpectedRecordVersion, AfterRecordVersion: assignment.RecordVersion, PublishCandidateID: assignment.PublishCandidateID, ActorRef: ctx.ActorRef, OccurredAt: at, RequestID: ctx.RequestID, AuditRef: ctx.AuditRef}
	audit := WorkflowRAGApplicationRuntimeAudit{SchemaVersion: workflowRAGApplicationRuntimeAuditSchemaVersion, AuditEventID: auditID, AssignmentID: assignment.AssignmentID, AssignmentDigest: assignment.AssignmentDigest, RecordVersion: assignment.RecordVersion, State: assignment.State, Decision: input.Decision, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, OwnerSubjectRef: ctx.OwnerSubjectRef, ActorRef: ctx.ActorRef, OccurredAt: at, RequestID: ctx.RequestID, AuditRef: ctx.AuditRef}
	if err = service.repository.Apply(ctx, input.ExpectedRecordVersion, assignment, event, audit); err != nil {
		return workflowRAGApplicationRuntimeRepositoryFailure(err)
	}
	events, audits = append(events, event), append(audits, audit)
	return WorkflowRAGApplicationRuntimeResult{Assignment: &assignment, Events: events, Audits: audits, CurrentRecordVersion: assignment.RecordVersion, CurrentState: assignment.State}
}

func validateWorkflowRAGApplicationRuntimeContext(ctx WorkflowRAGApplicationRuntimeContext) error {
	if ctx.RequestContext == nil || strings.TrimSpace(ctx.RequestID) == "" || strings.TrimSpace(ctx.TenantRef) == "" || strings.TrimSpace(ctx.WorkspaceID) == "" || strings.TrimSpace(ctx.ApplicationID) == "" || strings.TrimSpace(ctx.ActorRef) == "" || strings.TrimSpace(ctx.OwnerSubjectRef) == "" || strings.TrimSpace(ctx.AuditRef) == "" {
		return errWorkflowRAGApplicationStoreContract
	}
	return nil
}

func validateWorkflowRAGApplicationRuntimeMutation(ctx WorkflowRAGApplicationRuntimeContext, current workflowRAGApplicationRuntimeMemoryEntry, exists bool, assignment WorkflowRAGApplicationRuntimeAssignment, event WorkflowRAGApplicationRuntimeEvent, audit WorkflowRAGApplicationRuntimeAudit) error {
	if validateStoredWorkflowRAGApplicationAssignment(assignment, ctx) != nil || validateStoredWorkflowRAGApplicationEvent(event) != nil || validateStoredWorkflowRAGApplicationAudit(audit, ctx) != nil || event.AssignmentID != assignment.AssignmentID || audit.AssignmentID != assignment.AssignmentID || audit.AssignmentDigest != assignment.AssignmentDigest || event.AfterRecordVersion != assignment.RecordVersion || audit.RecordVersion != assignment.RecordVersion || event.ToState != assignment.State || audit.State != assignment.State || event.Decision != audit.Decision {
		return errWorkflowRAGApplicationStoreContract
	}
	if !exists {
		if event.Decision != workflowRAGApplicationRuntimeDecisionActivate || event.BeforeRecordVersion != 0 || event.FromState != "none" || assignment.RecordVersion != 1 {
			return errWorkflowRAGApplicationStoreContract
		}
		return nil
	}
	if assignment.AssignmentID != current.assignment.AssignmentID || assignment.CreatedAt != current.assignment.CreatedAt || assignment.CreatedByActorRef != current.assignment.CreatedByActorRef || event.BeforeRecordVersion != current.assignment.RecordVersion || assignment.RecordVersion != current.assignment.RecordVersion+1 || event.FromState != current.assignment.State {
		return errWorkflowRAGApplicationStoreContract
	}
	return nil
}

func validateWorkflowRAGApplicationRuntimeEntry(ctx WorkflowRAGApplicationRuntimeContext, entry workflowRAGApplicationRuntimeMemoryEntry) error {
	if validateStoredWorkflowRAGApplicationAssignment(entry.assignment, ctx) != nil || len(entry.events) == 0 || len(entry.events) != len(entry.audits) || len(entry.events) != entry.assignment.RecordVersion {
		return errWorkflowRAGApplicationStoreContract
	}
	for index := range entry.events {
		if validateStoredWorkflowRAGApplicationEvent(entry.events[index]) != nil || validateStoredWorkflowRAGApplicationAudit(entry.audits[index], ctx) != nil || entry.events[index].AfterRecordVersion != index+1 || entry.audits[index].RecordVersion != index+1 {
			return errWorkflowRAGApplicationStoreContract
		}
	}
	return nil
}

func validateStoredWorkflowRAGApplicationAssignment(value WorkflowRAGApplicationRuntimeAssignment, ctx WorkflowRAGApplicationRuntimeContext) error {
	created, createdErr := time.Parse(time.RFC3339Nano, value.CreatedAt)
	updated, updatedErr := time.Parse(time.RFC3339Nano, value.UpdatedAt)
	digest, digestErr := workflowRAGApplicationAssignmentDigest(value)
	if value.SchemaVersion != workflowRAGApplicationRuntimeAssignmentSchemaVersion || !workflowRAGApplicationAssignmentIDPattern.MatchString(value.AssignmentID) || value.RecordVersion < 1 || value.AssignmentDigest != digest || digestErr != nil || value.TenantRef != ctx.TenantRef || value.WorkspaceID != ctx.WorkspaceID || value.ApplicationID != ctx.ApplicationID || value.OwnerSubjectRef != ctx.OwnerSubjectRef || (value.State != workflowRAGApplicationRuntimeStateActive && value.State != workflowRAGApplicationRuntimeStateRevoked) || !applicationDraftIdentifierPattern.MatchString(value.PublishCandidateID) || value.PublishReviewVersion < 1 || value.PublishCandidateState != applicationPublishStateApproved || !applicationDraftIdentifierPattern.MatchString(value.DraftID) || value.DraftVersion < 1 || !workflowRAGDigestPattern.MatchString(value.DraftDigest) || !validWorkflowRAGApplicationBindingRef(value.BindingRef) || createdErr != nil || updatedErr != nil || updated.Before(created) || strings.TrimSpace(value.CreatedByActorRef) == "" || strings.TrimSpace(value.UpdatedByActorRef) == "" || strings.TrimSpace(value.RequestID) == "" || strings.TrimSpace(value.AuditRef) == "" {
		return errWorkflowRAGApplicationStoreContract
	}
	return nil
}

func validateStoredWorkflowRAGApplicationEvent(value WorkflowRAGApplicationRuntimeEvent) error {
	if value.SchemaVersion != workflowRAGApplicationRuntimeEventSchemaVersion || !workflowRAGApplicationEventIDPattern.MatchString(value.EventID) || !workflowRAGApplicationAssignmentIDPattern.MatchString(value.AssignmentID) || !workflowRAGApplicationDecisionAllowed(value.Decision) || !utf8.ValidString(value.Reason) || utf8.RuneCountInString(value.Reason) < 4 || utf8.RuneCountInString(value.Reason) > 500 || value.BeforeRecordVersion < 0 || value.AfterRecordVersion != value.BeforeRecordVersion+1 || !applicationDraftIdentifierPattern.MatchString(value.PublishCandidateID) || strings.TrimSpace(value.ActorRef) == "" || strings.TrimSpace(value.RequestID) == "" || strings.TrimSpace(value.AuditRef) == "" {
		return errWorkflowRAGApplicationStoreContract
	}
	if _, err := time.Parse(time.RFC3339Nano, value.OccurredAt); err != nil {
		return errWorkflowRAGApplicationStoreContract
	}
	return nil
}

func validateStoredWorkflowRAGApplicationAudit(value WorkflowRAGApplicationRuntimeAudit, ctx WorkflowRAGApplicationRuntimeContext) error {
	if value.SchemaVersion != workflowRAGApplicationRuntimeAuditSchemaVersion || !workflowRAGApplicationAuditIDPattern.MatchString(value.AuditEventID) || !workflowRAGApplicationAssignmentIDPattern.MatchString(value.AssignmentID) || !workflowRAGDigestPattern.MatchString(value.AssignmentDigest) || value.RecordVersion < 1 || (value.State != workflowRAGApplicationRuntimeStateActive && value.State != workflowRAGApplicationRuntimeStateRevoked) || !workflowRAGApplicationDecisionAllowed(value.Decision) || value.TenantRef != ctx.TenantRef || value.WorkspaceID != ctx.WorkspaceID || value.ApplicationID != ctx.ApplicationID || value.OwnerSubjectRef != ctx.OwnerSubjectRef || strings.TrimSpace(value.ActorRef) == "" || strings.TrimSpace(value.RequestID) == "" || strings.TrimSpace(value.AuditRef) == "" {
		return errWorkflowRAGApplicationStoreContract
	}
	if _, err := time.Parse(time.RFC3339Nano, value.OccurredAt); err != nil {
		return errWorkflowRAGApplicationStoreContract
	}
	return nil
}

func workflowRAGApplicationAssignmentDigest(value WorkflowRAGApplicationRuntimeAssignment) (string, error) {
	seal := struct {
		AssignmentID         string                           `json:"assignment_id"`
		RecordVersion        int                              `json:"record_version"`
		TenantRef            string                           `json:"tenant_ref"`
		WorkspaceID          string                           `json:"workspace_id"`
		ApplicationID        string                           `json:"application_id"`
		OwnerSubjectRef      string                           `json:"owner_subject_ref"`
		State                string                           `json:"state"`
		PublishCandidateID   string                           `json:"publish_candidate_id"`
		PublishReviewVersion int                              `json:"publish_review_version"`
		DraftID              string                           `json:"draft_id"`
		DraftVersion         int                              `json:"draft_version"`
		DraftDigest          string                           `json:"draft_digest"`
		BindingRef           WorkflowRAGApplicationBindingRef `json:"binding_ref"`
	}{value.AssignmentID, value.RecordVersion, value.TenantRef, value.WorkspaceID, value.ApplicationID, value.OwnerSubjectRef, value.State, value.PublishCandidateID, value.PublishReviewVersion, value.DraftID, value.DraftVersion, value.DraftDigest, value.BindingRef}
	payload, err := json.Marshal(seal)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(digest[:]), nil
}

func workflowRAGApplicationRuntimeKey(ctx WorkflowRAGApplicationRuntimeContext) string {
	return strings.Join([]string{ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef}, "\x00")
}

func workflowRAGApplicationDecisionAllowed(value string) bool {
	return value == workflowRAGApplicationRuntimeDecisionActivate || value == workflowRAGApplicationRuntimeDecisionReplace || value == workflowRAGApplicationRuntimeDecisionRevoke
}

func workflowRAGApplicationProtocolAllowed(protocol string, allowed []string) bool {
	protocol = strings.TrimSpace(protocol)
	for _, candidate := range allowed {
		if protocol == strings.TrimSpace(candidate) {
			return true
		}
	}
	return false
}

func workflowRAGApplicationPromotionContext(ctx WorkflowRAGApplicationRuntimeContext) WorkflowRAGPromotionContext {
	return WorkflowRAGPromotionContext{RequestContext: ctx.RequestContext, RequestID: ctx.RequestID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, ActorRef: ctx.ActorRef, OwnerSubjectRef: ctx.OwnerSubjectRef, AuditRef: ctx.AuditRef, WriteEnabled: ctx.WriteEnabled}
}

func workflowRAGApplicationPublishContext(ctx WorkflowRAGApplicationRuntimeContext) ApplicationPublishContext {
	return ApplicationPublishContext{RequestContext: ctx.RequestContext, RequestID: ctx.RequestID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, ActorRef: ctx.ActorRef, OwnerSubjectRef: ctx.OwnerSubjectRef, AuditRef: ctx.AuditRef, RAGPromotionReadEnabled: true}
}

func workflowRAGApplicationDraftContext(ctx WorkflowRAGApplicationRuntimeContext) ApplicationConfigurationDraftContext {
	return ApplicationConfigurationDraftContext{RequestContext: ctx.RequestContext, RequestID: ctx.RequestID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, ActorRef: ctx.ActorRef, OwnerSubjectRef: ctx.OwnerSubjectRef, AuditRef: ctx.AuditRef}
}

func cloneWorkflowRAGApplicationAssignment(value WorkflowRAGApplicationRuntimeAssignment) WorkflowRAGApplicationRuntimeAssignment {
	return value
}

func workflowRAGApplicationRuntimeFailure(code string) WorkflowRAGApplicationRuntimeResult {
	return WorkflowRAGApplicationRuntimeResult{Events: []WorkflowRAGApplicationRuntimeEvent{}, Audits: []WorkflowRAGApplicationRuntimeAudit{}, FailureCode: code}
}

func workflowRAGApplicationRuntimeRepositoryFailure(err error) WorkflowRAGApplicationRuntimeResult {
	switch {
	case errors.Is(err, errWorkflowRAGApplicationAssignmentNotFound):
		return workflowRAGApplicationRuntimeFailure(WorkflowRAGApplicationFailureAssignmentNotFound)
	case errors.Is(err, errWorkflowRAGApplicationVersionConflict):
		return workflowRAGApplicationRuntimeFailure(WorkflowRAGApplicationFailureVersionConflict)
	case errors.Is(err, errWorkflowRAGApplicationTransition):
		return workflowRAGApplicationRuntimeFailure(WorkflowRAGApplicationFailureTransitionInvalid)
	case errors.Is(err, errWorkflowRAGApplicationStoreContract):
		return workflowRAGApplicationRuntimeFailure(WorkflowRAGApplicationFailureStoreContractMismatch)
	default:
		return workflowRAGApplicationRuntimeFailure(WorkflowRAGApplicationFailureStoreUnavailable)
	}
}
