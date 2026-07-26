package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"
)

const (
	PromptApplicationRuntimeFailureScopeDenied      = "prompt_runtime_scope_denied"
	PromptApplicationRuntimeFailureNotFound         = "prompt_runtime_assignment_not_found"
	PromptApplicationRuntimeFailureVersionConflict  = "prompt_runtime_assignment_version_conflict"
	PromptApplicationRuntimeFailureCandidate        = "prompt_runtime_candidate_ineligible"
	PromptApplicationRuntimeFailureAuthorityChanged = "prompt_runtime_authority_changed"
	PromptApplicationRuntimeFailureTransition       = "prompt_runtime_transition_invalid"
	PromptApplicationRuntimeFailurePayload          = "prompt_runtime_payload_invalid"
	PromptApplicationRuntimeFailureWriteDisabled    = "prompt_runtime_write_disabled"
	PromptApplicationRuntimeFailureStoreUnavailable = "prompt_runtime_store_unavailable"
	PromptApplicationRuntimeFailureStoreContract    = "prompt_runtime_store_contract_mismatch"
)

var (
	errPromptApplicationRuntimeNotFound        = errors.New(PromptApplicationRuntimeFailureNotFound)
	errPromptApplicationRuntimeVersionConflict = errors.New(PromptApplicationRuntimeFailureVersionConflict)
	errPromptApplicationRuntimeStore           = errors.New(PromptApplicationRuntimeFailureStoreUnavailable)
	errPromptApplicationRuntimeContract        = errors.New(PromptApplicationRuntimeFailureStoreContract)
)

type PromptApplicationRuntimeContext struct {
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

type PromptApplicationRuntimeDecisionInput struct {
	ExpectedAssignmentVersion int    `json:"expected_assignment_version"`
	Action                    string `json:"action"`
	CandidateID               string `json:"candidate_id"`
}

type PromptApplicationRuntimeResult struct {
	Assignment               *PromptApplicationRuntimeAssignment
	Events                   []PromptApplicationRuntimeAssignmentEvent
	FailureCode              string
	CurrentAssignmentVersion int
	CurrentState             string
}

type promptApplicationRuntimeRepository interface {
	Read(PromptApplicationRuntimeContext) (PromptApplicationRuntimeAssignment, []PromptApplicationRuntimeAssignmentEvent, error)
	Apply(PromptApplicationRuntimeContext, int, PromptApplicationRuntimeAssignment, PromptApplicationRuntimeAssignmentEvent) error
}

type promptApplicationRuntimeMemoryEntry struct {
	assignment PromptApplicationRuntimeAssignment
	events     []PromptApplicationRuntimeAssignmentEvent
}

type memoryPromptApplicationRuntimeRepository struct {
	mu          sync.RWMutex
	entries     map[string]promptApplicationRuntimeMemoryEntry
	unavailable bool
}

func newMemoryPromptApplicationRuntimeRepository() *memoryPromptApplicationRuntimeRepository {
	return &memoryPromptApplicationRuntimeRepository{entries: make(map[string]promptApplicationRuntimeMemoryEntry)}
}

func newPromptApplicationRuntimeRepositoryForRunStore(store workflowRunStore) (promptApplicationRuntimeRepository, error) {
	switch typed := store.(type) {
	case *memoryWorkflowRunStore:
		return newMemoryPromptApplicationRuntimeRepository(), nil
	case *sqliteWorkflowRunStore:
		if typed.database == nil {
			return nil, errors.New("Prompt application runtime requires the shared SQLite database")
		}
		return newSQLitePromptApplicationRuntimeRepository(typed.database), nil
	case *postgresWorkflowRunStore:
		if typed.pool == nil {
			return nil, errors.New("Prompt application runtime requires the Workflow PostgreSQL pool")
		}
		return newPostgresPromptApplicationRuntimeRepository(typed.pool), nil
	default:
		return nil, errors.New("Prompt application runtime requires a supported Workflow runtime backend")
	}
}

func (repository *memoryPromptApplicationRuntimeRepository) Read(ctx PromptApplicationRuntimeContext) (PromptApplicationRuntimeAssignment, []PromptApplicationRuntimeAssignmentEvent, error) {
	if repository == nil || repository.unavailable {
		return PromptApplicationRuntimeAssignment{}, nil, errPromptApplicationRuntimeStore
	}
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	entry, ok := repository.entries[promptApplicationRuntimeKey(ctx)]
	if !ok {
		return PromptApplicationRuntimeAssignment{}, nil, errPromptApplicationRuntimeNotFound
	}
	if validatePromptApplicationRuntimeEntry(ctx, entry) != nil {
		return PromptApplicationRuntimeAssignment{}, nil, errPromptApplicationRuntimeContract
	}
	return entry.assignment, append([]PromptApplicationRuntimeAssignmentEvent(nil), entry.events...), nil
}

func (repository *memoryPromptApplicationRuntimeRepository) Apply(ctx PromptApplicationRuntimeContext, expectedVersion int, assignment PromptApplicationRuntimeAssignment, event PromptApplicationRuntimeAssignmentEvent) error {
	if repository == nil || repository.unavailable {
		return errPromptApplicationRuntimeStore
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	key := promptApplicationRuntimeKey(ctx)
	current, exists := repository.entries[key]
	currentVersion := 0
	if exists {
		if validatePromptApplicationRuntimeEntry(ctx, current) != nil {
			return errPromptApplicationRuntimeContract
		}
		currentVersion = current.assignment.AssignmentVersion
	}
	if currentVersion != expectedVersion {
		return errPromptApplicationRuntimeVersionConflict
	}
	if validatePromptApplicationRuntimeMutation(ctx, current, exists, assignment, event) != nil {
		return errPromptApplicationRuntimeContract
	}
	current.assignment = assignment
	current.events = append(current.events, event)
	repository.entries[key] = current
	return nil
}

type promptApplicationRuntimeAuthority struct {
	Application ApplicationSummary
	Candidate   ApplicationPublishCandidate
	Draft       ApplicationConfigurationDraft
	Template    PromptApplicationTemplateVersion
}

type promptApplicationRuntimeAuthorityResolver struct {
	publishRepository  applicationPublishCandidateRepository
	draftRepository    applicationConfigurationDraftRepository
	templateRepository promptApplicationTemplateRepository
	readApplication    applicationPublishBaselineReader
}

func (resolver promptApplicationRuntimeAuthorityResolver) Resolve(ctx PromptApplicationRuntimeContext, candidateID string, expected *PromptApplicationRuntimeAssignment) (promptApplicationRuntimeAuthority, string) {
	publishContext := promptApplicationRuntimePublishContext(ctx)
	application, err := resolver.readApplication(publishContext)
	if err != nil || application.ApplicationKind != "prompt_application" || strings.TrimSpace(application.ApplicationRef) == "" {
		return promptApplicationRuntimeAuthority{}, PromptApplicationRuntimeFailureCandidate
	}
	candidate, err := resolver.publishRepository.Read(publishContext, candidateID)
	if err != nil || candidate.SchemaVersion != applicationPublishCandidateSchemaVersionV3 || candidate.CandidateState != applicationPublishStateApproved || candidate.ReviewVersion < 1 || candidate.Configuration.PromptTemplateRef == nil {
		return promptApplicationRuntimeAuthority{}, PromptApplicationRuntimeFailureCandidate
	}
	candidates, err := resolver.publishRepository.List(publishContext)
	if err != nil {
		return promptApplicationRuntimeAuthority{}, PromptApplicationRuntimeFailureStoreUnavailable
	}
	if applicationPublishCandidateIsSuperseded(candidate, candidates) {
		return promptApplicationRuntimeAuthority{}, PromptApplicationRuntimeFailureCandidate
	}
	draftContext := promptApplicationRuntimeDraftContext(ctx)
	draft, err := resolver.draftRepository.Read(draftContext, candidate.DraftID)
	if err != nil {
		return promptApplicationRuntimeAuthority{}, PromptApplicationRuntimeFailureAuthorityChanged
	}
	digest, err := applicationConfigurationCanonicalDigest(applicationPublishSnapshotFromDraft(draft))
	if err != nil || draft.SchemaVersion != applicationConfigurationDraftSchemaVersionV3 || draft.DraftVersion != candidate.DraftVersion || digest != candidate.DraftDigest ||
		strings.TrimSpace(draft.BaseApplicationUpdatedAt) != strings.TrimSpace(application.UpdatedAt) || !draft.ValidationSummary.IsValid ||
		!validateApplicationConfigurationDraftPayload(draftContext, draft.ApplicationConfigurationDraftPayload).IsValid ||
		!promptApplicationTemplateRefsEqual(draft.PromptTemplateRef, candidate.Configuration.PromptTemplateRef) {
		return promptApplicationRuntimeAuthority{}, PromptApplicationRuntimeFailureAuthorityChanged
	}
	ref := *candidate.Configuration.PromptTemplateRef
	templateContext := promptApplicationRuntimeTemplateContext(ctx)
	template, err := resolver.templateRepository.ReadVersion(templateContext, ref.TemplateID, ref.TemplateVersion)
	if err != nil || validateStoredPromptApplicationTemplateVersion(templateContext, template) != nil || template.TemplateDigest != ref.TemplateDigest {
		return promptApplicationRuntimeAuthority{}, PromptApplicationRuntimeFailureAuthorityChanged
	}
	if expected != nil && (expected.CandidateID != candidate.CandidateID || expected.CandidateReviewVersion != candidate.ReviewVersion ||
		expected.DraftID != candidate.DraftID || expected.DraftVersion != candidate.DraftVersion || expected.DraftDigest != candidate.DraftDigest || expected.PromptTemplateRef != ref) {
		return promptApplicationRuntimeAuthority{}, PromptApplicationRuntimeFailureAuthorityChanged
	}
	return promptApplicationRuntimeAuthority{Application: application, Candidate: candidate, Draft: draft, Template: template}, ""
}

type promptApplicationRuntimeService struct {
	repository promptApplicationRuntimeRepository
	resolver   promptApplicationRuntimeAuthorityResolver
	now        func() time.Time
	newID      func(string) (string, error)
}

func newPromptApplicationRuntimeService(repository promptApplicationRuntimeRepository, resolver promptApplicationRuntimeAuthorityResolver) promptApplicationRuntimeService {
	return promptApplicationRuntimeService{repository: repository, resolver: resolver, now: func() time.Time { return time.Now().UTC() }, newID: newWorkflowRAGStableID}
}

func (service promptApplicationRuntimeService) Read(ctx PromptApplicationRuntimeContext) PromptApplicationRuntimeResult {
	if validatePromptApplicationRuntimeContext(ctx) != nil {
		return promptApplicationRuntimeFailure(PromptApplicationRuntimeFailureScopeDenied)
	}
	assignment, events, err := service.repository.Read(ctx)
	if err != nil {
		return promptApplicationRuntimeRepositoryFailure(err)
	}
	result := PromptApplicationRuntimeResult{Assignment: &assignment, Events: events, CurrentAssignmentVersion: assignment.AssignmentVersion, CurrentState: assignment.State}
	if assignment.State == "active" {
		if _, failure := service.resolver.Resolve(ctx, assignment.CandidateID, &assignment); failure != "" {
			result.FailureCode = PromptApplicationRuntimeFailureAuthorityChanged
		}
	}
	return result
}

func (service promptApplicationRuntimeService) Decide(ctx PromptApplicationRuntimeContext, input PromptApplicationRuntimeDecisionInput) PromptApplicationRuntimeResult {
	input.Action, input.CandidateID = strings.TrimSpace(input.Action), strings.TrimSpace(input.CandidateID)
	if validatePromptApplicationRuntimeContext(ctx) != nil {
		return promptApplicationRuntimeFailure(PromptApplicationRuntimeFailureScopeDenied)
	}
	if !ctx.WriteEnabled {
		return promptApplicationRuntimeFailure(PromptApplicationRuntimeFailureWriteDisabled)
	}
	if input.ExpectedAssignmentVersion < 0 || !promptApplicationRuntimeActionAllowed(input.Action) ||
		(input.Action == "revoke" && input.CandidateID != "") || (input.Action != "revoke" && !applicationDraftIdentifierPattern.MatchString(input.CandidateID)) {
		return promptApplicationRuntimeFailure(PromptApplicationRuntimeFailurePayload)
	}
	current, events, readErr := service.repository.Read(ctx)
	exists := readErr == nil
	if readErr != nil && !errors.Is(readErr, errPromptApplicationRuntimeNotFound) {
		return promptApplicationRuntimeRepositoryFailure(readErr)
	}
	if (!exists && (input.Action != "activate" || input.ExpectedAssignmentVersion != 0)) || (exists && current.AssignmentVersion != input.ExpectedAssignmentVersion) {
		result := promptApplicationRuntimeFailure(PromptApplicationRuntimeFailureVersionConflict)
		if exists {
			result.CurrentAssignmentVersion, result.CurrentState = current.AssignmentVersion, current.State
		}
		return result
	}
	if exists && input.Action == "activate" || !exists && input.Action != "activate" || exists && current.State != "active" ||
		exists && input.Action == "replace" && current.CandidateID == input.CandidateID {
		return PromptApplicationRuntimeResult{FailureCode: PromptApplicationRuntimeFailureTransition, CurrentAssignmentVersion: current.AssignmentVersion, CurrentState: current.State, Events: []PromptApplicationRuntimeAssignmentEvent{}}
	}
	var authority promptApplicationRuntimeAuthority
	var failure string
	if input.Action != "revoke" {
		authority, failure = service.resolver.Resolve(ctx, input.CandidateID, nil)
		if failure != "" {
			return promptApplicationRuntimeFailure(PromptApplicationRuntimeFailureCandidate)
		}
	}
	at := service.now().UTC().Format(time.RFC3339Nano)
	assignment := current
	if !exists {
		assignmentID, err := service.newID("ptra_")
		if err != nil {
			return promptApplicationRuntimeFailure(PromptApplicationRuntimeFailureStoreUnavailable)
		}
		assignment = PromptApplicationRuntimeAssignment{SchemaVersion: promptApplicationRuntimeAssignmentSchema, AssignmentID: assignmentID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, OwnerSubjectRef: ctx.OwnerSubjectRef, ActivatedAt: at, ActivatedByActorRef: ctx.ActorRef}
	}
	assignment.AssignmentVersion = input.ExpectedAssignmentVersion + 1
	assignment.UpdatedAt, assignment.UpdatedByActorRef, assignment.RequestID, assignment.AuditRef = at, ctx.ActorRef, ctx.RequestID, ctx.AuditRef
	assignment.State, assignment.RevokedAt = "active", nil
	if input.Action == "revoke" {
		assignment.State, assignment.RevokedAt = "revoked", &at
	} else {
		assignment.CandidateID, assignment.CandidateReviewVersion = authority.Candidate.CandidateID, authority.Candidate.ReviewVersion
		assignment.DraftID, assignment.DraftVersion, assignment.DraftDigest = authority.Draft.DraftID, authority.Draft.DraftVersion, authority.Candidate.DraftDigest
		assignment.PromptTemplateRef = *authority.Candidate.Configuration.PromptTemplateRef
	}
	digest, err := promptApplicationRuntimeAssignmentDigest(assignment)
	if err != nil {
		return promptApplicationRuntimeFailure(PromptApplicationRuntimeFailureStoreContract)
	}
	assignment.AssignmentDigest = digest
	eventID, err := service.newID("ptrae_")
	if err != nil {
		return promptApplicationRuntimeFailure(PromptApplicationRuntimeFailureStoreUnavailable)
	}
	event := PromptApplicationRuntimeAssignmentEvent{
		SchemaVersion: promptApplicationRuntimeAssignmentEventSchema, EventID: eventID, AssignmentID: assignment.AssignmentID,
		TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, OwnerSubjectRef: ctx.OwnerSubjectRef,
		EventSequence: len(events) + 1, Action: input.Action, ExpectedAssignmentVersion: input.ExpectedAssignmentVersion,
		ResultingAssignmentVersion: assignment.AssignmentVersion, CandidateID: assignment.CandidateID, CandidateReviewVersion: assignment.CandidateReviewVersion,
		DraftID: assignment.DraftID, DraftVersion: assignment.DraftVersion, DraftDigest: assignment.DraftDigest, PromptTemplateRef: assignment.PromptTemplateRef,
		AssignmentDigest: assignment.AssignmentDigest, OccurredAt: at, ActorRef: ctx.ActorRef, RequestID: ctx.RequestID, AuditRef: ctx.AuditRef,
	}
	if err := service.repository.Apply(ctx, input.ExpectedAssignmentVersion, assignment, event); err != nil {
		return promptApplicationRuntimeRepositoryFailure(err)
	}
	events = append(events, event)
	return PromptApplicationRuntimeResult{Assignment: &assignment, Events: events, CurrentAssignmentVersion: assignment.AssignmentVersion, CurrentState: assignment.State}
}

func validatePromptApplicationRuntimeContext(ctx PromptApplicationRuntimeContext) error {
	if ctx.RequestContext == nil || !validPromptApplicationScope(ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef) ||
		!validPromptApplicationRef(ctx.RequestID) || !validPromptApplicationRef(ctx.ActorRef) || !validPromptApplicationRef(ctx.AuditRef) {
		return errPromptApplicationRuntimeContract
	}
	return nil
}

func validatePromptApplicationRuntimeMutation(ctx PromptApplicationRuntimeContext, current promptApplicationRuntimeMemoryEntry, exists bool, assignment PromptApplicationRuntimeAssignment, event PromptApplicationRuntimeAssignmentEvent) error {
	if validatePromptApplicationRuntimeAssignment(assignment) != nil || validatePromptApplicationRuntimeAssignmentEvent(event) != nil ||
		assignment.TenantRef != ctx.TenantRef || assignment.WorkspaceID != ctx.WorkspaceID || assignment.ApplicationID != ctx.ApplicationID || assignment.OwnerSubjectRef != ctx.OwnerSubjectRef ||
		event.AssignmentID != assignment.AssignmentID || event.EventSequence != len(current.events)+1 || event.ResultingAssignmentVersion != assignment.AssignmentVersion || event.AssignmentDigest != assignment.AssignmentDigest ||
		event.CandidateID != assignment.CandidateID || event.CandidateReviewVersion != assignment.CandidateReviewVersion || event.DraftID != assignment.DraftID || event.DraftVersion != assignment.DraftVersion || event.DraftDigest != assignment.DraftDigest || event.PromptTemplateRef != assignment.PromptTemplateRef {
		return errPromptApplicationRuntimeContract
	}
	if !exists && (event.Action != "activate" || assignment.AssignmentVersion != 1) || exists && (current.assignment.State != "active" || current.assignment.AssignmentID != assignment.AssignmentID || assignment.AssignmentVersion != current.assignment.AssignmentVersion+1) {
		return errPromptApplicationRuntimeContract
	}
	want, err := promptApplicationRuntimeAssignmentDigest(assignment)
	if err != nil || want != assignment.AssignmentDigest {
		return errPromptApplicationRuntimeContract
	}
	return nil
}

func validatePromptApplicationRuntimeEntry(ctx PromptApplicationRuntimeContext, entry promptApplicationRuntimeMemoryEntry) error {
	if validatePromptApplicationRuntimeAssignment(entry.assignment) != nil || entry.assignment.TenantRef != ctx.TenantRef || entry.assignment.WorkspaceID != ctx.WorkspaceID || entry.assignment.ApplicationID != ctx.ApplicationID || entry.assignment.OwnerSubjectRef != ctx.OwnerSubjectRef || len(entry.events) != entry.assignment.AssignmentVersion {
		return errPromptApplicationRuntimeContract
	}
	want, err := promptApplicationRuntimeAssignmentDigest(entry.assignment)
	if err != nil || want != entry.assignment.AssignmentDigest {
		return errPromptApplicationRuntimeContract
	}
	for index, event := range entry.events {
		if validatePromptApplicationRuntimeAssignmentEvent(event) != nil || event.EventSequence != index+1 || event.ResultingAssignmentVersion != index+1 || event.AssignmentID != entry.assignment.AssignmentID || event.TenantRef != ctx.TenantRef || event.WorkspaceID != ctx.WorkspaceID || event.ApplicationID != ctx.ApplicationID || event.OwnerSubjectRef != ctx.OwnerSubjectRef {
			return errPromptApplicationRuntimeContract
		}
	}
	return nil
}

func promptApplicationRuntimeAssignmentDigest(value PromptApplicationRuntimeAssignment) (string, error) {
	value.AssignmentDigest = ""
	payload, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(digest[:]), nil
}

func promptApplicationRuntimeKey(ctx PromptApplicationRuntimeContext) string {
	return strings.Join([]string{ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef}, "\x00")
}

func promptApplicationRuntimeActionAllowed(action string) bool {
	return action == "activate" || action == "replace" || action == "revoke"
}

func promptApplicationRuntimePublishContext(ctx PromptApplicationRuntimeContext) ApplicationPublishContext {
	return ApplicationPublishContext{RequestContext: ctx.RequestContext, RequestID: ctx.RequestID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, ActorRef: ctx.ActorRef, OwnerSubjectRef: ctx.OwnerSubjectRef, AuditRef: ctx.AuditRef, PromptTemplateSourceReadEnabled: true}
}

func promptApplicationRuntimeDraftContext(ctx PromptApplicationRuntimeContext) ApplicationConfigurationDraftContext {
	return ApplicationConfigurationDraftContext{RequestContext: ctx.RequestContext, RequestID: ctx.RequestID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, ActorRef: ctx.ActorRef, OwnerSubjectRef: ctx.OwnerSubjectRef, AuditRef: ctx.AuditRef}
}

func promptApplicationRuntimeTemplateContext(ctx PromptApplicationRuntimeContext) PromptApplicationTemplateContext {
	return PromptApplicationTemplateContext{RequestContext: ctx.RequestContext, RequestID: ctx.RequestID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, ActorRef: ctx.ActorRef, OwnerSubjectRef: ctx.OwnerSubjectRef, AuditRef: ctx.AuditRef}
}

func promptApplicationRuntimeFailure(code string) PromptApplicationRuntimeResult {
	return PromptApplicationRuntimeResult{FailureCode: code, Events: []PromptApplicationRuntimeAssignmentEvent{}}
}

func promptApplicationRuntimeRepositoryFailure(err error) PromptApplicationRuntimeResult {
	switch {
	case errors.Is(err, errPromptApplicationRuntimeNotFound):
		return promptApplicationRuntimeFailure(PromptApplicationRuntimeFailureNotFound)
	case errors.Is(err, errPromptApplicationRuntimeVersionConflict):
		return promptApplicationRuntimeFailure(PromptApplicationRuntimeFailureVersionConflict)
	case errors.Is(err, errPromptApplicationRuntimeContract):
		return promptApplicationRuntimeFailure(PromptApplicationRuntimeFailureStoreContract)
	default:
		return promptApplicationRuntimeFailure(PromptApplicationRuntimeFailureStoreUnavailable)
	}
}
