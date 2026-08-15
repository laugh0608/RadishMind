package httpapi

import (
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresApplicationResultArtifactRepository struct{ pool *pgxpool.Pool }

func newPostgresApplicationResultArtifactRepository(pool *pgxpool.Pool) *postgresApplicationResultArtifactRepository {
	return &postgresApplicationResultArtifactRepository{pool: pool}
}

type postgresApplicationResultArtifactRow interface {
	Scan(...any) error
}

func (repository *postgresApplicationResultArtifactRepository) Create(
	ctx ApplicationInteractionContext,
	artifact ApplicationResultArtifact,
) (ApplicationResultArtifact, bool, error) {
	if repository == nil || repository.pool == nil {
		return ApplicationResultArtifact{}, false, errApplicationResultArtifactStore
	}
	payload, err := encodeApplicationResultArtifact(ctx, artifact)
	if err != nil {
		return ApplicationResultArtifact{}, false, err
	}
	createdAt := parseApplicationInteractionTimestamp(artifact.CreatedAt)
	if createdAt == nil {
		return ApplicationResultArtifact{}, false, errApplicationResultArtifactContract
	}
	command, err := repository.pool.Exec(applicationInteractionRequestContext(ctx), `INSERT INTO application_result_artifacts
(tenant_ref,workspace_id,application_id,owner_subject_ref,artifact_id,session_id,turn_id,client_turn_key,
execution_profile,run_id,run_schema_version,content_type,content_bytes,content_digest,created_at,artifact_payload)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16) ON CONFLICT DO NOTHING`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, artifact.ArtifactID,
		artifact.SessionID, artifact.TurnID, artifact.ClientTurnKey, artifact.ExecutionProfile,
		artifact.RunRef.RunID, artifact.RunRef.SchemaVersion, artifact.ContentType, artifact.ContentBytes,
		artifact.ContentDigest, createdAt.Round(time.Microsecond), payload)
	if err != nil {
		return ApplicationResultArtifact{}, false, errApplicationResultArtifactStore
	}
	if command.RowsAffected() == 1 {
		return cloneApplicationResultArtifact(artifact), false, nil
	}
	if command.RowsAffected() != 0 {
		return ApplicationResultArtifact{}, false, errApplicationResultArtifactStore
	}
	existing, err := repository.ReadByTurn(ctx, artifact.SessionID, artifact.TurnID)
	if err == nil && applicationResultArtifactsEquivalent(existing, artifact) {
		return existing, true, nil
	}
	if err != nil && !errors.Is(err, errApplicationResultArtifactNotFound) {
		return ApplicationResultArtifact{}, false, err
	}
	return ApplicationResultArtifact{}, false, errApplicationResultArtifactConflict
}

func (repository *postgresApplicationResultArtifactRepository) Read(
	ctx ApplicationInteractionContext,
	artifactID string,
) (ApplicationResultArtifact, error) {
	if repository == nil || repository.pool == nil {
		return ApplicationResultArtifact{}, errApplicationResultArtifactStore
	}
	return readPostgresApplicationResultArtifact(repository.pool.QueryRow(applicationInteractionRequestContext(ctx), `SELECT artifact_id,session_id,turn_id,client_turn_key,execution_profile,run_id,run_schema_version,
content_type,content_bytes,content_digest,created_at,artifact_payload FROM application_result_artifacts
WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 AND artifact_id=$5`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, artifactID), ctx)
}

func (repository *postgresApplicationResultArtifactRepository) ReadByTurn(
	ctx ApplicationInteractionContext,
	sessionID string,
	turnID string,
) (ApplicationResultArtifact, error) {
	if repository == nil || repository.pool == nil {
		return ApplicationResultArtifact{}, errApplicationResultArtifactStore
	}
	return readPostgresApplicationResultArtifact(repository.pool.QueryRow(applicationInteractionRequestContext(ctx), `SELECT artifact_id,session_id,turn_id,client_turn_key,execution_profile,run_id,run_schema_version,
content_type,content_bytes,content_digest,created_at,artifact_payload FROM application_result_artifacts
WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 AND session_id=$5 AND turn_id=$6`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, sessionID, turnID), ctx)
}

func (repository *postgresApplicationResultArtifactRepository) List(
	ctx ApplicationInteractionContext,
	sessionID string,
) ([]ApplicationResultArtifact, error) {
	if repository == nil || repository.pool == nil {
		return nil, errApplicationResultArtifactStore
	}
	rows, err := repository.pool.Query(applicationInteractionRequestContext(ctx), `SELECT artifact_id,session_id,turn_id,client_turn_key,execution_profile,run_id,run_schema_version,
content_type,content_bytes,content_digest,created_at,artifact_payload FROM application_result_artifacts
WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 AND session_id=$5
ORDER BY created_at DESC,artifact_id DESC`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, sessionID)
	if err != nil {
		return nil, errApplicationResultArtifactStore
	}
	defer rows.Close()
	artifacts := make([]ApplicationResultArtifact, 0)
	for rows.Next() {
		artifact, scanErr := readPostgresApplicationResultArtifact(rows, ctx)
		if scanErr != nil {
			return nil, scanErr
		}
		artifacts = append(artifacts, artifact)
	}
	if rows.Err() != nil {
		return nil, errApplicationResultArtifactStore
	}
	return artifacts, nil
}

func readPostgresApplicationResultArtifact(
	row postgresApplicationResultArtifactRow,
	ctx ApplicationInteractionContext,
) (ApplicationResultArtifact, error) {
	var artifactID, sessionID, turnID, clientTurnKey, executionProfile string
	var runID, runSchemaVersion, contentType, contentDigest string
	var contentBytes int
	var createdAt time.Time
	var payload []byte
	if err := row.Scan(&artifactID, &sessionID, &turnID, &clientTurnKey, &executionProfile, &runID,
		&runSchemaVersion, &contentType, &contentBytes, &contentDigest, &createdAt, &payload); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ApplicationResultArtifact{}, errApplicationResultArtifactNotFound
		}
		return ApplicationResultArtifact{}, errApplicationResultArtifactStore
	}
	artifact, err := decodeApplicationResultArtifact(ctx, payload)
	if err != nil || !applicationResultArtifactProjectionMatches(
		artifact, artifactID, sessionID, turnID, clientTurnKey, executionProfile, runID,
		runSchemaVersion, contentType, contentBytes, contentDigest, createdAt, true,
	) {
		return ApplicationResultArtifact{}, errApplicationResultArtifactContract
	}
	return artifact, nil
}
