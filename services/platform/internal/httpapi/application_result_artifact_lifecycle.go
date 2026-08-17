package httpapi

import (
	"strings"
	"time"
)

const (
	applicationResultArtifactLifecycleSchemaVersion      = "application_result_artifact_lifecycle.v1"
	applicationResultArtifactLifecycleEventSchemaVersion = "application_result_artifact_lifecycle_event.v1"
)

type ApplicationResultArtifactLifecycleState string

const (
	ApplicationResultArtifactLifecycleActive   ApplicationResultArtifactLifecycleState = "active"
	ApplicationResultArtifactLifecycleArchived ApplicationResultArtifactLifecycleState = "archived"
)

type ApplicationResultArtifactLifecycleTransitionKind string

const (
	ApplicationResultArtifactLifecycleTransitionArchived   ApplicationResultArtifactLifecycleTransitionKind = "archived"
	ApplicationResultArtifactLifecycleTransitionUnarchived ApplicationResultArtifactLifecycleTransitionKind = "unarchived"
)

type ApplicationResultArtifactLifecycle struct {
	SchemaVersion     string                                  `json:"schema_version"`
	TenantRef         string                                  `json:"tenant_ref"`
	WorkspaceID       string                                  `json:"workspace_id"`
	ApplicationID     string                                  `json:"application_id"`
	OwnerSubjectRef   string                                  `json:"owner_subject_ref"`
	ArtifactID        string                                  `json:"artifact_id"`
	LifecycleState    ApplicationResultArtifactLifecycleState `json:"lifecycle_state"`
	LifecycleVersion  int                                     `json:"lifecycle_version"`
	ArchivedAt        *string                                 `json:"archived_at"`
	UpdatedAt         string                                  `json:"updated_at"`
	UpdatedByActorRef string                                  `json:"updated_by_actor_ref"`
	RequestID         string                                  `json:"request_id"`
	AuditRef          string                                  `json:"audit_ref"`
}

type ApplicationResultArtifactLifecycleEvent struct {
	SchemaVersion    string                                           `json:"schema_version"`
	TenantRef        string                                           `json:"tenant_ref"`
	WorkspaceID      string                                           `json:"workspace_id"`
	ApplicationID    string                                           `json:"application_id"`
	OwnerSubjectRef  string                                           `json:"owner_subject_ref"`
	ArtifactID       string                                           `json:"artifact_id"`
	LifecycleVersion int                                              `json:"lifecycle_version"`
	FromState        ApplicationResultArtifactLifecycleState          `json:"from_state"`
	ToState          ApplicationResultArtifactLifecycleState          `json:"to_state"`
	TransitionKind   ApplicationResultArtifactLifecycleTransitionKind `json:"transition_kind"`
	OccurredAt       string                                           `json:"occurred_at"`
	ActorRef         string                                           `json:"actor_ref"`
	RequestID        string                                           `json:"request_id"`
	AuditRef         string                                           `json:"audit_ref"`
}

type applicationResultArtifactStoredRecord struct {
	Artifact  ApplicationResultArtifact
	Lifecycle ApplicationResultArtifactLifecycle
}

type ApplicationResultArtifactLifecycleTransitionInput struct {
	SessionID                string
	ArtifactID               string
	ExpectedLifecycleVersion int
}

type ApplicationResultArtifactLifecycleTransitionResult struct {
	Lifecycle               *ApplicationResultArtifactLifecycle
	Event                   *ApplicationResultArtifactLifecycleEvent
	FailureCode             string
	CurrentLifecycleVersion int
	CurrentLifecycleState   ApplicationResultArtifactLifecycleState
}

func initialApplicationResultArtifactLifecycle(artifact ApplicationResultArtifact) ApplicationResultArtifactLifecycle {
	return ApplicationResultArtifactLifecycle{
		SchemaVersion: applicationResultArtifactLifecycleSchemaVersion,
		TenantRef:     artifact.TenantRef, WorkspaceID: artifact.WorkspaceID, ApplicationID: artifact.ApplicationID,
		OwnerSubjectRef: artifact.OwnerSubjectRef, ArtifactID: artifact.ArtifactID,
		LifecycleState: ApplicationResultArtifactLifecycleActive, LifecycleVersion: 1,
		UpdatedAt: artifact.CreatedAt, UpdatedByActorRef: artifact.CreatedByActorRef,
		RequestID: artifact.RequestID, AuditRef: artifact.AuditRef,
	}
}

func (service applicationResultArtifactService) Archive(
	ctx ApplicationInteractionContext,
	input ApplicationResultArtifactLifecycleTransitionInput,
) ApplicationResultArtifactLifecycleTransitionResult {
	return service.transitionLifecycle(ctx, input, ApplicationResultArtifactLifecycleArchived)
}

func (service applicationResultArtifactService) Unarchive(
	ctx ApplicationInteractionContext,
	input ApplicationResultArtifactLifecycleTransitionInput,
) ApplicationResultArtifactLifecycleTransitionResult {
	return service.transitionLifecycle(ctx, input, ApplicationResultArtifactLifecycleActive)
}

func (service applicationResultArtifactService) transitionLifecycle(
	ctx ApplicationInteractionContext,
	input ApplicationResultArtifactLifecycleTransitionInput,
	target ApplicationResultArtifactLifecycleState,
) ApplicationResultArtifactLifecycleTransitionResult {
	if service.repository == nil {
		return applicationResultArtifactLifecycleFailure(ApplicationResultArtifactFailureStoreUnavailable)
	}
	if !ctx.WriteEnabled || validateApplicationInteractionContext(ctx) != nil ||
		!applicationSessionIDPattern.MatchString(strings.TrimSpace(input.SessionID)) ||
		!applicationResultArtifactIDPattern.MatchString(strings.TrimSpace(input.ArtifactID)) ||
		input.ExpectedLifecycleVersion < 1 || !validApplicationResultArtifactLifecycleState(target) {
		return applicationResultArtifactLifecycleFailure(ApplicationResultArtifactFailurePayloadInvalid)
	}
	artifact, err := service.repository.Read(ctx, strings.TrimSpace(input.ArtifactID))
	if err != nil {
		return applicationResultArtifactLifecycleRepositoryFailure(err, ApplicationResultArtifactLifecycle{})
	}
	if artifact.SessionID != strings.TrimSpace(input.SessionID) {
		return applicationResultArtifactLifecycleFailure(ApplicationResultArtifactFailureNotFound)
	}
	lifecycle, event, err := service.repository.TransitionLifecycle(
		ctx,
		artifact.ArtifactID,
		target,
		input.ExpectedLifecycleVersion,
		service.now().UTC(),
	)
	if err != nil {
		return applicationResultArtifactLifecycleRepositoryFailure(err, lifecycle)
	}
	if validateApplicationResultArtifactLifecycle(ctx, lifecycle) != nil ||
		validateApplicationResultArtifactLifecycleEvent(ctx, event) != nil ||
		lifecycle.ArtifactID != artifact.ArtifactID || event.ArtifactID != artifact.ArtifactID ||
		event.LifecycleVersion != lifecycle.LifecycleVersion || event.ToState != target ||
		lifecycle.LifecycleVersion != input.ExpectedLifecycleVersion+1 {
		return applicationResultArtifactLifecycleFailure(ApplicationResultArtifactFailureStoreContract)
	}
	return ApplicationResultArtifactLifecycleTransitionResult{
		Lifecycle: &lifecycle, Event: &event,
		CurrentLifecycleVersion: lifecycle.LifecycleVersion, CurrentLifecycleState: lifecycle.LifecycleState,
	}
}

func applicationResultArtifactLifecycleFailure(code string) ApplicationResultArtifactLifecycleTransitionResult {
	return ApplicationResultArtifactLifecycleTransitionResult{FailureCode: code}
}

func applicationResultArtifactLifecycleRepositoryFailure(
	err error,
	current ApplicationResultArtifactLifecycle,
) ApplicationResultArtifactLifecycleTransitionResult {
	result := applicationResultArtifactLifecycleFailure(applicationResultArtifactRepositoryFailure(err).FailureCode)
	if validateApplicationResultArtifactLifecycleWithoutContext(current) == nil {
		result.CurrentLifecycleVersion = current.LifecycleVersion
		result.CurrentLifecycleState = current.LifecycleState
	}
	return result
}

func (repository *memoryApplicationResultArtifactRepository) ReadLifecycle(
	ctx ApplicationInteractionContext,
	artifactID string,
) (ApplicationResultArtifactLifecycle, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	if repository.unavailable {
		return ApplicationResultArtifactLifecycle{}, errApplicationResultArtifactStore
	}
	artifact, found := repository.byID[strings.TrimSpace(artifactID)]
	lifecycle, lifecycleFound := repository.lifecycleByID[strings.TrimSpace(artifactID)]
	if !found || !lifecycleFound || !applicationResultArtifactScopeMatches(ctx, artifact) {
		return ApplicationResultArtifactLifecycle{}, errApplicationResultArtifactNotFound
	}
	if validateApplicationResultArtifactLifecycle(ctx, lifecycle) != nil {
		return ApplicationResultArtifactLifecycle{}, errApplicationResultArtifactContract
	}
	return cloneApplicationResultArtifactLifecycle(lifecycle), nil
}

func (repository *memoryApplicationResultArtifactRepository) ListByLifecycle(
	ctx ApplicationInteractionContext,
	sessionID string,
	state ApplicationResultArtifactLifecycleState,
) ([]applicationResultArtifactStoredRecord, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	if repository.unavailable {
		return nil, errApplicationResultArtifactStore
	}
	records := make([]applicationResultArtifactStoredRecord, 0)
	for _, artifact := range repository.byID {
		if artifact.SessionID != strings.TrimSpace(sessionID) || !applicationResultArtifactScopeMatches(ctx, artifact) {
			continue
		}
		lifecycle, found := repository.lifecycleByID[artifact.ArtifactID]
		if !found || validateApplicationResultArtifactLifecycle(ctx, lifecycle) != nil || lifecycle.ArtifactID != artifact.ArtifactID {
			return nil, errApplicationResultArtifactContract
		}
		if lifecycle.LifecycleState == state {
			records = append(records, applicationResultArtifactStoredRecord{
				Artifact: cloneApplicationResultArtifact(artifact), Lifecycle: cloneApplicationResultArtifactLifecycle(lifecycle),
			})
		}
	}
	return records, nil
}

func (repository *memoryApplicationResultArtifactRepository) TransitionLifecycle(
	ctx ApplicationInteractionContext,
	artifactID string,
	target ApplicationResultArtifactLifecycleState,
	expectedVersion int,
	now time.Time,
) (ApplicationResultArtifactLifecycle, ApplicationResultArtifactLifecycleEvent, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if repository.unavailable {
		return ApplicationResultArtifactLifecycle{}, ApplicationResultArtifactLifecycleEvent{}, errApplicationResultArtifactStore
	}
	artifact, found := repository.byID[strings.TrimSpace(artifactID)]
	current, lifecycleFound := repository.lifecycleByID[strings.TrimSpace(artifactID)]
	if !found || !lifecycleFound || !applicationResultArtifactScopeMatches(ctx, artifact) {
		return ApplicationResultArtifactLifecycle{}, ApplicationResultArtifactLifecycleEvent{}, errApplicationResultArtifactNotFound
	}
	if validateApplicationResultArtifactLifecycle(ctx, current) != nil || current.ArtifactID != artifact.ArtifactID {
		return ApplicationResultArtifactLifecycle{}, ApplicationResultArtifactLifecycleEvent{}, errApplicationResultArtifactContract
	}
	if current.LifecycleVersion != expectedVersion {
		return cloneApplicationResultArtifactLifecycle(current), ApplicationResultArtifactLifecycleEvent{}, errApplicationResultArtifactLifecycleVersion
	}
	if current.LifecycleState == target || !validApplicationResultArtifactLifecycleState(target) {
		return cloneApplicationResultArtifactLifecycle(current), ApplicationResultArtifactLifecycleEvent{}, errApplicationResultArtifactLifecycleState
	}
	updated, event := nextApplicationResultArtifactLifecycle(ctx, current, target, now)
	if validateApplicationResultArtifactLifecycle(ctx, updated) != nil || validateApplicationResultArtifactLifecycleEvent(ctx, event) != nil {
		return cloneApplicationResultArtifactLifecycle(current), ApplicationResultArtifactLifecycleEvent{}, errApplicationResultArtifactContract
	}
	events := repository.lifecycleEvents[artifact.ArtifactID]
	if events == nil {
		events = make(map[int]ApplicationResultArtifactLifecycleEvent)
	}
	if _, duplicate := events[event.LifecycleVersion]; duplicate {
		return cloneApplicationResultArtifactLifecycle(current), ApplicationResultArtifactLifecycleEvent{}, errApplicationResultArtifactContract
	}
	repository.lifecycleByID[artifact.ArtifactID] = cloneApplicationResultArtifactLifecycle(updated)
	events[event.LifecycleVersion] = event
	repository.lifecycleEvents[artifact.ArtifactID] = events
	return cloneApplicationResultArtifactLifecycle(updated), event, nil
}

func nextApplicationResultArtifactLifecycle(
	ctx ApplicationInteractionContext,
	current ApplicationResultArtifactLifecycle,
	target ApplicationResultArtifactLifecycleState,
	now time.Time,
) (ApplicationResultArtifactLifecycle, ApplicationResultArtifactLifecycleEvent) {
	occurredAt := now.UTC().Format(time.RFC3339Nano)
	updated := cloneApplicationResultArtifactLifecycle(current)
	updated.LifecycleState = target
	updated.LifecycleVersion++
	updated.UpdatedAt = occurredAt
	updated.UpdatedByActorRef = strings.TrimSpace(ctx.ActorRef)
	updated.RequestID = strings.TrimSpace(ctx.RequestID)
	updated.AuditRef = strings.TrimSpace(ctx.AuditRef)
	transition := ApplicationResultArtifactLifecycleTransitionArchived
	if target == ApplicationResultArtifactLifecycleArchived {
		updated.ArchivedAt = &occurredAt
	} else {
		updated.ArchivedAt = nil
		transition = ApplicationResultArtifactLifecycleTransitionUnarchived
	}
	event := ApplicationResultArtifactLifecycleEvent{
		SchemaVersion: applicationResultArtifactLifecycleEventSchemaVersion,
		TenantRef:     current.TenantRef, WorkspaceID: current.WorkspaceID, ApplicationID: current.ApplicationID,
		OwnerSubjectRef: current.OwnerSubjectRef, ArtifactID: current.ArtifactID,
		LifecycleVersion: updated.LifecycleVersion, FromState: current.LifecycleState, ToState: target,
		TransitionKind: transition, OccurredAt: occurredAt, ActorRef: strings.TrimSpace(ctx.ActorRef),
		RequestID: strings.TrimSpace(ctx.RequestID), AuditRef: strings.TrimSpace(ctx.AuditRef),
	}
	return updated, event
}

func validateApplicationResultArtifactLifecycle(
	ctx ApplicationInteractionContext,
	lifecycle ApplicationResultArtifactLifecycle,
) error {
	if lifecycle.TenantRef != ctx.TenantRef || lifecycle.WorkspaceID != ctx.WorkspaceID ||
		lifecycle.ApplicationID != ctx.ApplicationID || lifecycle.OwnerSubjectRef != ctx.OwnerSubjectRef {
		return errApplicationResultArtifactContract
	}
	return validateApplicationResultArtifactLifecycleWithoutContext(lifecycle)
}

func validateApplicationResultArtifactLifecycleWithoutContext(lifecycle ApplicationResultArtifactLifecycle) error {
	updatedAt := parseApplicationInteractionTimestamp(lifecycle.UpdatedAt)
	if lifecycle.SchemaVersion != applicationResultArtifactLifecycleSchemaVersion ||
		strings.TrimSpace(lifecycle.TenantRef) == "" || strings.TrimSpace(lifecycle.WorkspaceID) == "" ||
		strings.TrimSpace(lifecycle.ApplicationID) == "" || strings.TrimSpace(lifecycle.OwnerSubjectRef) == "" ||
		!applicationResultArtifactIDPattern.MatchString(lifecycle.ArtifactID) ||
		!validApplicationResultArtifactLifecycleState(lifecycle.LifecycleState) || lifecycle.LifecycleVersion < 1 ||
		updatedAt == nil || strings.TrimSpace(lifecycle.UpdatedByActorRef) == "" ||
		strings.TrimSpace(lifecycle.RequestID) == "" || strings.TrimSpace(lifecycle.AuditRef) == "" {
		return errApplicationResultArtifactContract
	}
	if lifecycle.LifecycleState == ApplicationResultArtifactLifecycleActive && lifecycle.ArchivedAt != nil {
		return errApplicationResultArtifactContract
	}
	if lifecycle.LifecycleState == ApplicationResultArtifactLifecycleArchived {
		if lifecycle.ArchivedAt == nil || parseApplicationInteractionTimestamp(*lifecycle.ArchivedAt) == nil {
			return errApplicationResultArtifactContract
		}
	}
	return nil
}

func validateApplicationResultArtifactLifecycleEvent(
	ctx ApplicationInteractionContext,
	event ApplicationResultArtifactLifecycleEvent,
) error {
	if event.SchemaVersion != applicationResultArtifactLifecycleEventSchemaVersion ||
		event.TenantRef != ctx.TenantRef || event.WorkspaceID != ctx.WorkspaceID ||
		event.ApplicationID != ctx.ApplicationID || event.OwnerSubjectRef != ctx.OwnerSubjectRef ||
		!applicationResultArtifactIDPattern.MatchString(event.ArtifactID) || event.LifecycleVersion < 2 ||
		!validApplicationResultArtifactLifecycleState(event.FromState) ||
		!validApplicationResultArtifactLifecycleState(event.ToState) || event.FromState == event.ToState ||
		parseApplicationInteractionTimestamp(event.OccurredAt) == nil || strings.TrimSpace(event.ActorRef) == "" ||
		strings.TrimSpace(event.RequestID) == "" || strings.TrimSpace(event.AuditRef) == "" {
		return errApplicationResultArtifactContract
	}
	if (event.TransitionKind == ApplicationResultArtifactLifecycleTransitionArchived &&
		event.FromState == ApplicationResultArtifactLifecycleActive && event.ToState == ApplicationResultArtifactLifecycleArchived) ||
		(event.TransitionKind == ApplicationResultArtifactLifecycleTransitionUnarchived &&
			event.FromState == ApplicationResultArtifactLifecycleArchived && event.ToState == ApplicationResultArtifactLifecycleActive) {
		return nil
	}
	return errApplicationResultArtifactContract
}

func validApplicationResultArtifactLifecycleState(state ApplicationResultArtifactLifecycleState) bool {
	return state == ApplicationResultArtifactLifecycleActive || state == ApplicationResultArtifactLifecycleArchived
}

func cloneApplicationResultArtifactLifecycle(lifecycle ApplicationResultArtifactLifecycle) ApplicationResultArtifactLifecycle {
	copy := lifecycle
	copy.ArchivedAt = cloneStringPointer(lifecycle.ArchivedAt)
	return copy
}

func cloneStringPointer(value *string) *string {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

var _ applicationResultArtifactRepository = (*memoryApplicationResultArtifactRepository)(nil)
