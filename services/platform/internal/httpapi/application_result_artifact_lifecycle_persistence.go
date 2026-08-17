package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"
)

func encodeApplicationResultArtifactLifecycle(
	ctx ApplicationInteractionContext,
	lifecycle ApplicationResultArtifactLifecycle,
) ([]byte, error) {
	if validateApplicationResultArtifactLifecycle(ctx, lifecycle) != nil {
		return nil, errApplicationResultArtifactContract
	}
	payload, err := json.Marshal(lifecycle)
	if err != nil {
		return nil, errApplicationResultArtifactContract
	}
	return payload, nil
}

func decodeApplicationResultArtifactLifecycle(
	ctx ApplicationInteractionContext,
	payload []byte,
) (ApplicationResultArtifactLifecycle, error) {
	var lifecycle ApplicationResultArtifactLifecycle
	decoder := json.NewDecoder(strings.NewReader(string(payload)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&lifecycle); err != nil {
		return ApplicationResultArtifactLifecycle{}, errApplicationResultArtifactContract
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return ApplicationResultArtifactLifecycle{}, errApplicationResultArtifactContract
	}
	if validateApplicationResultArtifactLifecycle(ctx, lifecycle) != nil {
		return ApplicationResultArtifactLifecycle{}, errApplicationResultArtifactContract
	}
	return lifecycle, nil
}

func encodeApplicationResultArtifactLifecycleEvent(
	ctx ApplicationInteractionContext,
	event ApplicationResultArtifactLifecycleEvent,
) ([]byte, error) {
	if validateApplicationResultArtifactLifecycleEvent(ctx, event) != nil {
		return nil, errApplicationResultArtifactContract
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return nil, errApplicationResultArtifactContract
	}
	return payload, nil
}

func applicationResultArtifactLifecycleProjectionMatches(
	lifecycle ApplicationResultArtifactLifecycle,
	artifactID string,
	state ApplicationResultArtifactLifecycleState,
	version int,
	archivedAt *time.Time,
	updatedAt time.Time,
	updatedByActorRef string,
	requestID string,
	auditRef string,
	postgresTimestamp bool,
) bool {
	payloadUpdatedAt := parseApplicationInteractionTimestamp(lifecycle.UpdatedAt)
	if payloadUpdatedAt == nil {
		return false
	}
	updatedAtMatches := payloadUpdatedAt.Equal(updatedAt)
	if postgresTimestamp {
		updatedAtMatches = applicationInteractionPostgresTimesEqual(*payloadUpdatedAt, updatedAt)
	}
	archivedAtMatches := lifecycle.ArchivedAt == nil && archivedAt == nil
	if lifecycle.ArchivedAt != nil && archivedAt != nil {
		payloadArchivedAt := parseApplicationInteractionTimestamp(*lifecycle.ArchivedAt)
		archivedAtMatches = payloadArchivedAt != nil && payloadArchivedAt.Equal(*archivedAt)
		if postgresTimestamp && payloadArchivedAt != nil {
			archivedAtMatches = applicationInteractionPostgresTimesEqual(*payloadArchivedAt, *archivedAt)
		}
	}
	return lifecycle.ArtifactID == artifactID && lifecycle.LifecycleState == state &&
		lifecycle.LifecycleVersion == version && archivedAtMatches && updatedAtMatches &&
		lifecycle.UpdatedByActorRef == updatedByActorRef && lifecycle.RequestID == requestID && lifecycle.AuditRef == auditRef
}
