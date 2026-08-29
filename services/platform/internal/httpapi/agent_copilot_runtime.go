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
	AgentCopilotRuntimeFailureScopeDenied      = "agent_copilot_runtime_scope_denied"
	AgentCopilotRuntimeFailureNotFound         = "agent_copilot_runtime_assignment_not_found"
	AgentCopilotRuntimeFailureVersionConflict  = "agent_copilot_runtime_assignment_version_conflict"
	AgentCopilotRuntimeFailureCandidate        = "agent_copilot_runtime_candidate_ineligible"
	AgentCopilotRuntimeFailureAuthorityChanged = "agent_copilot_runtime_authority_changed"
	AgentCopilotRuntimeFailureTransition       = "agent_copilot_runtime_transition_invalid"
	AgentCopilotRuntimeFailurePayload          = "agent_copilot_runtime_payload_invalid"
	AgentCopilotRuntimeFailureWriteDisabled    = "agent_copilot_runtime_write_disabled"
	AgentCopilotRuntimeFailureStoreUnavailable = "agent_copilot_runtime_store_unavailable"
	AgentCopilotRuntimeFailureStoreContract    = "agent_copilot_runtime_store_contract_mismatch"
)

var (
	errAgentCopilotRuntimeNotFound        = errors.New(AgentCopilotRuntimeFailureNotFound)
	errAgentCopilotRuntimeVersionConflict = errors.New(AgentCopilotRuntimeFailureVersionConflict)
	errAgentCopilotRuntimeStore           = errors.New(AgentCopilotRuntimeFailureStoreUnavailable)
	errAgentCopilotRuntimeContract        = errors.New(AgentCopilotRuntimeFailureStoreContract)
)

type AgentCopilotRuntimeContext struct {
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

type AgentCopilotRuntimeDecisionInput struct {
	ExpectedAssignmentVersion int                                `json:"expected_assignment_version"`
	Action                    string                             `json:"action"`
	CandidateID               string                             `json:"candidate_id"`
	ActionSafetyCandidate     *ActionSafetyCandidateProjectionV1 `json:"-"`
}

type AgentCopilotRuntimeResult struct {
	Assignment               *AgentCopilotRuntimeAssignmentV1
	Events                   []AgentCopilotRuntimeAssignmentEventV1
	FailureCode              string
	CurrentAssignmentVersion int
	CurrentState             string
}

type agentCopilotRuntimeRepository interface {
	Read(AgentCopilotRuntimeContext) (AgentCopilotRuntimeAssignmentV1, []AgentCopilotRuntimeAssignmentEventV1, error)
	Apply(AgentCopilotRuntimeContext, int, AgentCopilotRuntimeAssignmentV1, AgentCopilotRuntimeAssignmentEventV1) error
}

type agentCopilotRuntimeMemoryEntry struct {
	assignment AgentCopilotRuntimeAssignmentV1
	events     []AgentCopilotRuntimeAssignmentEventV1
}

type memoryAgentCopilotRuntimeRepository struct {
	mu          sync.RWMutex
	entries     map[string]agentCopilotRuntimeMemoryEntry
	unavailable bool
}

func newMemoryAgentCopilotRuntimeRepository() *memoryAgentCopilotRuntimeRepository {
	return &memoryAgentCopilotRuntimeRepository{entries: make(map[string]agentCopilotRuntimeMemoryEntry)}
}

func newAgentCopilotRuntimeRepositoryForRunStore(store workflowRunStore) (agentCopilotRuntimeRepository, error) {
	switch typed := store.(type) {
	case *memoryWorkflowRunStore:
		return newMemoryAgentCopilotRuntimeRepository(), nil
	case *sqliteWorkflowRunStore:
		if typed.database == nil {
			return nil, errors.New("Agent Copilot runtime requires the shared SQLite database")
		}
		return newSQLiteAgentCopilotRuntimeRepository(typed.database), nil
	case *postgresWorkflowRunStore:
		if typed.pool == nil {
			return nil, errors.New("Agent Copilot runtime requires the Workflow PostgreSQL pool")
		}
		return newPostgresAgentCopilotRuntimeRepository(typed.pool), nil
	default:
		return nil, errors.New("Agent Copilot runtime requires a supported Workflow runtime backend")
	}
}

func (repository *memoryAgentCopilotRuntimeRepository) Read(ctx AgentCopilotRuntimeContext) (AgentCopilotRuntimeAssignmentV1, []AgentCopilotRuntimeAssignmentEventV1, error) {
	if repository == nil || repository.unavailable {
		return AgentCopilotRuntimeAssignmentV1{}, nil, errAgentCopilotRuntimeStore
	}
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	entry, ok := repository.entries[agentCopilotRuntimeKey(ctx)]
	if !ok {
		return AgentCopilotRuntimeAssignmentV1{}, nil, errAgentCopilotRuntimeNotFound
	}
	if validateAgentCopilotRuntimeEntry(ctx, entry) != nil {
		return AgentCopilotRuntimeAssignmentV1{}, nil, errAgentCopilotRuntimeContract
	}
	return cloneAgentCopilotRuntimeAssignment(entry.assignment), cloneAgentCopilotRuntimeEvents(entry.events), nil
}

func (repository *memoryAgentCopilotRuntimeRepository) Apply(
	ctx AgentCopilotRuntimeContext,
	expectedVersion int,
	assignment AgentCopilotRuntimeAssignmentV1,
	event AgentCopilotRuntimeAssignmentEventV1,
) error {
	if repository == nil || repository.unavailable {
		return errAgentCopilotRuntimeStore
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	key := agentCopilotRuntimeKey(ctx)
	current, exists := repository.entries[key]
	currentVersion := 0
	if exists {
		if validateAgentCopilotRuntimeEntry(ctx, current) != nil {
			return errAgentCopilotRuntimeContract
		}
		currentVersion = current.assignment.AssignmentVersion
	}
	if currentVersion != expectedVersion {
		return errAgentCopilotRuntimeVersionConflict
	}
	if validateAgentCopilotRuntimeMutation(ctx, current, exists, assignment, event) != nil {
		return errAgentCopilotRuntimeContract
	}
	current.assignment = cloneAgentCopilotRuntimeAssignment(assignment)
	current.events = append(current.events, cloneAgentCopilotRuntimeEvent(event))
	repository.entries[key] = current
	return nil
}

type agentCopilotRuntimeAuthority struct {
	Application ApplicationSummary
	Candidate   ApplicationPublishCandidate
	Draft       ApplicationConfigurationDraft
	Profile     AgentCopilotProfileVersionV1
}

type agentCopilotRuntimeAuthorityResolver struct {
	publishRepository applicationPublishCandidateRepository
	draftRepository   applicationConfigurationDraftRepository
	profileRepository agentCopilotProfileRepository
	readApplication   applicationPublishBaselineReader
}

func (resolver agentCopilotRuntimeAuthorityResolver) Resolve(
	ctx AgentCopilotRuntimeContext,
	candidateID string,
	expected *AgentCopilotRuntimeAssignmentV1,
) (agentCopilotRuntimeAuthority, string) {
	publishContext := agentCopilotRuntimePublishContext(ctx)
	application, err := resolver.readApplication(publishContext)
	if err != nil || application.ApplicationKind != "agent" || strings.TrimSpace(application.ApplicationRef) == "" {
		return agentCopilotRuntimeAuthority{}, AgentCopilotRuntimeFailureCandidate
	}
	candidate, err := resolver.publishRepository.Read(publishContext, candidateID)
	if err != nil || candidate.SchemaVersion != applicationPublishCandidateSchemaVersionV4 ||
		candidate.CandidateState != applicationPublishStateApproved || candidate.ReviewVersion < 1 ||
		candidate.Configuration.AgentCopilotProfileRef == nil {
		return agentCopilotRuntimeAuthority{}, AgentCopilotRuntimeFailureCandidate
	}
	candidates, err := resolver.publishRepository.List(publishContext)
	if err != nil {
		return agentCopilotRuntimeAuthority{}, AgentCopilotRuntimeFailureStoreUnavailable
	}
	if applicationPublishCandidateIsSuperseded(candidate, candidates) {
		return agentCopilotRuntimeAuthority{}, AgentCopilotRuntimeFailureCandidate
	}
	draftContext := agentCopilotRuntimeDraftContext(ctx)
	draft, err := resolver.draftRepository.Read(draftContext, candidate.DraftID)
	if err != nil {
		return agentCopilotRuntimeAuthority{}, AgentCopilotRuntimeFailureAuthorityChanged
	}
	digest, err := applicationConfigurationCanonicalDigest(applicationPublishSnapshotFromDraft(draft))
	if err != nil || draft.SchemaVersion != applicationConfigurationDraftSchemaVersionV4 ||
		draft.DraftVersion != candidate.DraftVersion || digest != candidate.DraftDigest ||
		strings.TrimSpace(draft.BaseApplicationUpdatedAt) != strings.TrimSpace(application.UpdatedAt) ||
		!draft.ValidationSummary.IsValid ||
		!validateApplicationConfigurationDraftPayload(draftContext, draft.ApplicationConfigurationDraftPayload).IsValid ||
		!agentCopilotProfileRefsEqual(draft.AgentCopilotProfileRef, candidate.Configuration.AgentCopilotProfileRef) {
		return agentCopilotRuntimeAuthority{}, AgentCopilotRuntimeFailureAuthorityChanged
	}
	ref := *candidate.Configuration.AgentCopilotProfileRef
	profileContext := agentCopilotRuntimeProfileContext(ctx)
	profile, err := resolver.profileRepository.ReadVersion(profileContext, ref.ProfileID, ref.ProfileVersion)
	if err != nil || validateStoredAgentCopilotProfileVersion(profileContext, profile) != nil ||
		!agentCopilotProfileRefsEqual(agentCopilotProfileRefFromVersion(profile), &ref) {
		return agentCopilotRuntimeAuthority{}, AgentCopilotRuntimeFailureAuthorityChanged
	}
	if expected != nil && (expected.CandidateID != candidate.CandidateID ||
		expected.CandidateReviewVersion != candidate.ReviewVersion ||
		expected.DraftID != candidate.DraftID || expected.DraftVersion != candidate.DraftVersion ||
		expected.DraftDigest != candidate.DraftDigest ||
		!agentCopilotProfileRefsEqual(&expected.AgentCopilotProfileRef, &ref)) {
		return agentCopilotRuntimeAuthority{}, AgentCopilotRuntimeFailureAuthorityChanged
	}
	return agentCopilotRuntimeAuthority{Application: application, Candidate: candidate, Draft: draft, Profile: profile}, ""
}

type agentCopilotRuntimeService struct {
	repository   agentCopilotRuntimeRepository
	resolver     agentCopilotRuntimeAuthorityResolver
	now          func() time.Time
	newID        func(string) (string, error)
	actionSafety *actionSafetyRuntimeV1
}

func newAgentCopilotRuntimeService(repository agentCopilotRuntimeRepository, resolver agentCopilotRuntimeAuthorityResolver) agentCopilotRuntimeService {
	return agentCopilotRuntimeService{
		repository:   repository,
		resolver:     resolver,
		now:          func() time.Time { return time.Now().UTC() },
		newID:        newWorkflowRAGStableID,
		actionSafety: newActionSafetyRuntimeV1("development"),
	}
}

func (service agentCopilotRuntimeService) Read(ctx AgentCopilotRuntimeContext) AgentCopilotRuntimeResult {
	if validateAgentCopilotRuntimeContext(ctx) != nil {
		return agentCopilotRuntimeFailure(AgentCopilotRuntimeFailureScopeDenied)
	}
	assignment, events, err := service.repository.Read(ctx)
	if err != nil {
		return agentCopilotRuntimeRepositoryFailure(err)
	}
	result := AgentCopilotRuntimeResult{
		Assignment: &assignment, Events: events,
		CurrentAssignmentVersion: assignment.AssignmentVersion, CurrentState: assignment.State,
	}
	if assignment.State == "active" {
		if _, failure := service.resolver.Resolve(ctx, assignment.CandidateID, &assignment); failure != "" {
			result.FailureCode = failure
		}
	}
	return result
}

func (service agentCopilotRuntimeService) Decide(ctx AgentCopilotRuntimeContext, input AgentCopilotRuntimeDecisionInput) AgentCopilotRuntimeResult {
	input.Action, input.CandidateID = strings.TrimSpace(input.Action), strings.TrimSpace(input.CandidateID)
	if validateAgentCopilotRuntimeContext(ctx) != nil {
		return agentCopilotRuntimeFailure(AgentCopilotRuntimeFailureScopeDenied)
	}
	if !ctx.WriteEnabled {
		return agentCopilotRuntimeFailure(AgentCopilotRuntimeFailureWriteDisabled)
	}
	if input.ExpectedAssignmentVersion < 0 || !agentCopilotRuntimeActionAllowed(input.Action) ||
		input.Action == "revoke" && input.CandidateID != "" ||
		input.Action != "revoke" && !applicationDraftIdentifierPattern.MatchString(input.CandidateID) {
		return agentCopilotRuntimeFailure(AgentCopilotRuntimeFailurePayload)
	}
	current, events, readErr := service.repository.Read(ctx)
	exists := readErr == nil
	if readErr != nil && !errors.Is(readErr, errAgentCopilotRuntimeNotFound) {
		return agentCopilotRuntimeRepositoryFailure(readErr)
	}
	if !exists && (input.Action != "activate" || input.ExpectedAssignmentVersion != 0) ||
		exists && current.AssignmentVersion != input.ExpectedAssignmentVersion {
		result := agentCopilotRuntimeFailure(AgentCopilotRuntimeFailureVersionConflict)
		if exists {
			result.CurrentAssignmentVersion, result.CurrentState = current.AssignmentVersion, current.State
		}
		return result
	}
	if exists && input.Action == "activate" || !exists && input.Action != "activate" ||
		exists && current.State != "active" ||
		exists && input.Action == "replace" && current.CandidateID == input.CandidateID {
		return AgentCopilotRuntimeResult{
			FailureCode:              AgentCopilotRuntimeFailureTransition,
			CurrentAssignmentVersion: current.AssignmentVersion,
			CurrentState:             current.State,
			Events:                   []AgentCopilotRuntimeAssignmentEventV1{},
		}
	}
	var authority agentCopilotRuntimeAuthority
	if input.Action != "revoke" {
		var failure string
		authority, failure = service.resolver.Resolve(ctx, input.CandidateID, nil)
		if failure != "" {
			return agentCopilotRuntimeFailure(failure)
		}
	}
	var actionSafetyAssignment *ActionSafetyAssignmentProjectionV1
	if input.ActionSafetyCandidate != nil {
		if input.Action == "revoke" {
			return agentCopilotRuntimeFailure(AgentCopilotRuntimeFailurePayload)
		}
		if _, memory := service.repository.(*memoryAgentCopilotRuntimeRepository); !memory || service.actionSafety == nil {
			return agentCopilotRuntimeFailure(AgentCopilotRuntimeFailureStoreContract)
		}
		projection, safetyFailure := service.actionSafety.ActivateCandidate(
			ctx, *input.ActionSafetyCandidate, authority.Profile, input.ExpectedAssignmentVersion+1, true,
		)
		if safetyFailure != "" {
			return agentCopilotRuntimeFailure(string(safetyFailure))
		}
		actionSafetyAssignment = &projection
	}
	at := service.now().UTC().Format(time.RFC3339Nano)
	assignment := current
	if !exists {
		assignmentID, err := service.newID("acra_")
		if err != nil {
			return agentCopilotRuntimeFailure(AgentCopilotRuntimeFailureStoreUnavailable)
		}
		assignment = AgentCopilotRuntimeAssignmentV1{
			SchemaVersion: agentCopilotRuntimeAssignmentSchema,
			AssignmentID:  assignmentID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID,
			ApplicationID: ctx.ApplicationID, OwnerSubjectRef: ctx.OwnerSubjectRef,
			ActivatedAt: at, ActivatedByActorRef: ctx.ActorRef,
		}
	}
	assignment.AssignmentVersion = input.ExpectedAssignmentVersion + 1
	assignment.UpdatedAt, assignment.UpdatedByActorRef = at, ctx.ActorRef
	assignment.RequestID, assignment.AuditRef = ctx.RequestID, ctx.AuditRef
	assignment.State, assignment.RevokedAt = "active", nil
	if input.Action == "revoke" {
		assignment.State, assignment.RevokedAt = "revoked", &at
		assignment.ActionSafety = nil
	} else {
		assignment.CandidateID = authority.Candidate.CandidateID
		assignment.CandidateReviewVersion = authority.Candidate.ReviewVersion
		assignment.DraftID, assignment.DraftVersion = authority.Draft.DraftID, authority.Draft.DraftVersion
		assignment.DraftDigest = authority.Candidate.DraftDigest
		assignment.AgentCopilotProfileRef = *authority.Candidate.Configuration.AgentCopilotProfileRef
		assignment.ActionSafety = cloneActionSafetyAssignmentProjection(actionSafetyAssignment)
	}
	digest, err := agentCopilotRuntimeAssignmentDigest(assignment)
	if err != nil {
		return agentCopilotRuntimeFailure(AgentCopilotRuntimeFailureStoreContract)
	}
	assignment.AssignmentDigest = digest
	eventID, err := service.newID("acrae_")
	if err != nil {
		return agentCopilotRuntimeFailure(AgentCopilotRuntimeFailureStoreUnavailable)
	}
	event := AgentCopilotRuntimeAssignmentEventV1{
		SchemaVersion: agentCopilotRuntimeAssignmentEventSchema,
		EventID:       eventID, AssignmentID: assignment.AssignmentID,
		TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID,
		OwnerSubjectRef: ctx.OwnerSubjectRef, EventSequence: len(events) + 1,
		Action: input.Action, ExpectedAssignmentVersion: input.ExpectedAssignmentVersion,
		ResultingAssignmentVersion: assignment.AssignmentVersion,
		CandidateID:                assignment.CandidateID, CandidateReviewVersion: assignment.CandidateReviewVersion,
		DraftID: assignment.DraftID, DraftVersion: assignment.DraftVersion, DraftDigest: assignment.DraftDigest,
		AgentCopilotProfileRef: assignment.AgentCopilotProfileRef,
		AssignmentDigest:       assignment.AssignmentDigest, OccurredAt: at, ActorRef: ctx.ActorRef,
		RequestID: ctx.RequestID, AuditRef: ctx.AuditRef,
		ActionSafety: cloneActionSafetyAssignmentProjection(assignment.ActionSafety),
	}
	if err := service.repository.Apply(ctx, input.ExpectedAssignmentVersion, assignment, event); err != nil {
		return agentCopilotRuntimeRepositoryFailure(err)
	}
	events = append(events, event)
	return AgentCopilotRuntimeResult{
		Assignment: &assignment, Events: events,
		CurrentAssignmentVersion: assignment.AssignmentVersion, CurrentState: assignment.State,
	}
}

func validateAgentCopilotRuntimeContext(ctx AgentCopilotRuntimeContext) error {
	if ctx.RequestContext == nil || !validAgentCopilotScope(ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef) ||
		!validPromptApplicationRef(ctx.RequestID) || !validPromptApplicationRef(ctx.ActorRef) || !validPromptApplicationRef(ctx.AuditRef) {
		return errAgentCopilotRuntimeContract
	}
	return nil
}

func validateAgentCopilotRuntimeMutation(
	ctx AgentCopilotRuntimeContext,
	current agentCopilotRuntimeMemoryEntry,
	exists bool,
	assignment AgentCopilotRuntimeAssignmentV1,
	event AgentCopilotRuntimeAssignmentEventV1,
) error {
	if validateAgentCopilotRuntimeAssignment(assignment) != nil || validateAgentCopilotRuntimeAssignmentEvent(event) != nil ||
		assignment.TenantRef != ctx.TenantRef || assignment.WorkspaceID != ctx.WorkspaceID ||
		assignment.ApplicationID != ctx.ApplicationID || assignment.OwnerSubjectRef != ctx.OwnerSubjectRef ||
		event.AssignmentID != assignment.AssignmentID || event.EventSequence != len(current.events)+1 ||
		event.ResultingAssignmentVersion != assignment.AssignmentVersion ||
		event.AssignmentDigest != assignment.AssignmentDigest ||
		event.CandidateID != assignment.CandidateID ||
		event.CandidateReviewVersion != assignment.CandidateReviewVersion ||
		event.DraftID != assignment.DraftID || event.DraftVersion != assignment.DraftVersion ||
		event.DraftDigest != assignment.DraftDigest ||
		event.AgentCopilotProfileRef != assignment.AgentCopilotProfileRef {
		return errAgentCopilotRuntimeContract
	}
	if (event.ActionSafety == nil) != (assignment.ActionSafety == nil) || event.ActionSafety != nil &&
		event.ActionSafety.ProjectionDigest != assignment.ActionSafety.ProjectionDigest {
		return errAgentCopilotRuntimeContract
	}
	if !exists && (event.Action != "activate" || assignment.AssignmentVersion != 1) ||
		exists && (current.assignment.State != "active" ||
			current.assignment.AssignmentID != assignment.AssignmentID ||
			assignment.AssignmentVersion != current.assignment.AssignmentVersion+1) {
		return errAgentCopilotRuntimeContract
	}
	want, err := agentCopilotRuntimeAssignmentDigest(assignment)
	if err != nil || want != assignment.AssignmentDigest {
		return errAgentCopilotRuntimeContract
	}
	return nil
}

func validateAgentCopilotRuntimeEntry(ctx AgentCopilotRuntimeContext, entry agentCopilotRuntimeMemoryEntry) error {
	if validateAgentCopilotRuntimeAssignment(entry.assignment) != nil ||
		entry.assignment.TenantRef != ctx.TenantRef || entry.assignment.WorkspaceID != ctx.WorkspaceID ||
		entry.assignment.ApplicationID != ctx.ApplicationID || entry.assignment.OwnerSubjectRef != ctx.OwnerSubjectRef ||
		len(entry.events) != entry.assignment.AssignmentVersion {
		return errAgentCopilotRuntimeContract
	}
	want, err := agentCopilotRuntimeAssignmentDigest(entry.assignment)
	if err != nil || want != entry.assignment.AssignmentDigest {
		return errAgentCopilotRuntimeContract
	}
	for index, event := range entry.events {
		if validateAgentCopilotRuntimeAssignmentEvent(event) != nil ||
			event.EventSequence != index+1 || event.ResultingAssignmentVersion != index+1 ||
			event.AssignmentID != entry.assignment.AssignmentID ||
			event.TenantRef != ctx.TenantRef || event.WorkspaceID != ctx.WorkspaceID ||
			event.ApplicationID != ctx.ApplicationID || event.OwnerSubjectRef != ctx.OwnerSubjectRef {
			return errAgentCopilotRuntimeContract
		}
	}
	latest := entry.events[len(entry.events)-1]
	if latest.ResultingAssignmentVersion != entry.assignment.AssignmentVersion ||
		latest.AssignmentDigest != entry.assignment.AssignmentDigest ||
		latest.CandidateID != entry.assignment.CandidateID ||
		latest.CandidateReviewVersion != entry.assignment.CandidateReviewVersion ||
		latest.DraftID != entry.assignment.DraftID ||
		latest.DraftVersion != entry.assignment.DraftVersion ||
		latest.DraftDigest != entry.assignment.DraftDigest ||
		latest.AgentCopilotProfileRef != entry.assignment.AgentCopilotProfileRef {
		return errAgentCopilotRuntimeContract
	}
	if (latest.ActionSafety == nil) != (entry.assignment.ActionSafety == nil) || latest.ActionSafety != nil &&
		latest.ActionSafety.ProjectionDigest != entry.assignment.ActionSafety.ProjectionDigest {
		return errAgentCopilotRuntimeContract
	}
	return nil
}

func agentCopilotRuntimeAssignmentDigest(value AgentCopilotRuntimeAssignmentV1) (string, error) {
	value.AssignmentDigest = ""
	payload, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(digest[:]), nil
}

func agentCopilotRuntimeKey(ctx AgentCopilotRuntimeContext) string {
	return strings.Join([]string{ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef}, "\x00")
}

func cloneAgentCopilotRuntimeAssignment(value AgentCopilotRuntimeAssignmentV1) AgentCopilotRuntimeAssignmentV1 {
	cloned := value
	cloned.ActionSafety = cloneActionSafetyAssignmentProjection(value.ActionSafety)
	if value.RevokedAt != nil {
		revokedAt := *value.RevokedAt
		cloned.RevokedAt = &revokedAt
	}
	return cloned
}

func cloneAgentCopilotRuntimeEvent(value AgentCopilotRuntimeAssignmentEventV1) AgentCopilotRuntimeAssignmentEventV1 {
	cloned := value
	cloned.ActionSafety = cloneActionSafetyAssignmentProjection(value.ActionSafety)
	return cloned
}

func cloneAgentCopilotRuntimeEvents(values []AgentCopilotRuntimeAssignmentEventV1) []AgentCopilotRuntimeAssignmentEventV1 {
	cloned := make([]AgentCopilotRuntimeAssignmentEventV1, len(values))
	for index, value := range values {
		cloned[index] = cloneAgentCopilotRuntimeEvent(value)
	}
	return cloned
}

func agentCopilotRuntimeActionAllowed(action string) bool {
	return action == "activate" || action == "replace" || action == "revoke"
}

func agentCopilotRuntimePublishContext(ctx AgentCopilotRuntimeContext) ApplicationPublishContext {
	return ApplicationPublishContext{
		RequestContext: ctx.RequestContext, RequestID: ctx.RequestID, TenantRef: ctx.TenantRef,
		WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, ActorRef: ctx.ActorRef,
		OwnerSubjectRef: ctx.OwnerSubjectRef, AuditRef: ctx.AuditRef, AgentProfileSourceReadEnabled: true,
	}
}

func agentCopilotRuntimeDraftContext(ctx AgentCopilotRuntimeContext) ApplicationConfigurationDraftContext {
	return ApplicationConfigurationDraftContext{
		RequestContext: ctx.RequestContext, RequestID: ctx.RequestID, TenantRef: ctx.TenantRef,
		WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, ActorRef: ctx.ActorRef,
		OwnerSubjectRef: ctx.OwnerSubjectRef, AuditRef: ctx.AuditRef,
	}
}

func agentCopilotRuntimeProfileContext(ctx AgentCopilotRuntimeContext) AgentCopilotProfileContext {
	return AgentCopilotProfileContext{
		RequestContext: ctx.RequestContext, RequestID: ctx.RequestID, TenantRef: ctx.TenantRef,
		WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, ActorRef: ctx.ActorRef,
		OwnerSubjectRef: ctx.OwnerSubjectRef, AuditRef: ctx.AuditRef,
	}
}

func agentCopilotRuntimeFailure(code string) AgentCopilotRuntimeResult {
	return AgentCopilotRuntimeResult{FailureCode: code, Events: []AgentCopilotRuntimeAssignmentEventV1{}}
}

func agentCopilotRuntimeRepositoryFailure(err error) AgentCopilotRuntimeResult {
	switch {
	case errors.Is(err, errAgentCopilotRuntimeNotFound):
		return agentCopilotRuntimeFailure(AgentCopilotRuntimeFailureNotFound)
	case errors.Is(err, errAgentCopilotRuntimeVersionConflict):
		return agentCopilotRuntimeFailure(AgentCopilotRuntimeFailureVersionConflict)
	case errors.Is(err, errAgentCopilotRuntimeContract):
		return agentCopilotRuntimeFailure(AgentCopilotRuntimeFailureStoreContract)
	default:
		return agentCopilotRuntimeFailure(AgentCopilotRuntimeFailureStoreUnavailable)
	}
}
