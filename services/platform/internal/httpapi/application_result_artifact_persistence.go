package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"
)

func encodeApplicationResultArtifact(
	ctx ApplicationInteractionContext,
	artifact ApplicationResultArtifact,
) ([]byte, error) {
	if validateApplicationResultArtifact(ctx, artifact) != nil {
		return nil, errApplicationResultArtifactContract
	}
	payload, err := json.Marshal(artifact)
	if err != nil {
		return nil, errApplicationResultArtifactContract
	}
	return payload, nil
}

func decodeApplicationResultArtifact(
	ctx ApplicationInteractionContext,
	payload []byte,
) (ApplicationResultArtifact, error) {
	var artifact ApplicationResultArtifact
	decoder := json.NewDecoder(strings.NewReader(string(payload)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&artifact); err != nil {
		return ApplicationResultArtifact{}, errApplicationResultArtifactContract
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return ApplicationResultArtifact{}, errApplicationResultArtifactContract
	}
	if validateApplicationResultArtifact(ctx, artifact) != nil {
		return ApplicationResultArtifact{}, errApplicationResultArtifactContract
	}
	return artifact, nil
}

func applicationResultArtifactProjectionMatches(
	artifact ApplicationResultArtifact,
	artifactID string,
	sessionID string,
	turnID string,
	clientTurnKey string,
	executionProfile string,
	runID string,
	runSchemaVersion string,
	contentType string,
	contentBytes int,
	contentDigest string,
	createdAt time.Time,
	postgresTimestamp bool,
) bool {
	payloadCreatedAt := parseApplicationInteractionTimestamp(artifact.CreatedAt)
	if payloadCreatedAt == nil {
		return false
	}
	timestampMatches := payloadCreatedAt.Equal(createdAt)
	if postgresTimestamp {
		timestampMatches = applicationInteractionPostgresTimesEqual(*payloadCreatedAt, createdAt)
	}
	return artifact.ArtifactID == artifactID && artifact.SessionID == sessionID && artifact.TurnID == turnID &&
		artifact.ClientTurnKey == clientTurnKey && artifact.ExecutionProfile == executionProfile &&
		artifact.RunRef.RunID == runID && artifact.RunRef.SchemaVersion == runSchemaVersion &&
		artifact.ContentType == contentType && artifact.ContentBytes == contentBytes &&
		artifact.ContentDigest == contentDigest && timestampMatches
}
