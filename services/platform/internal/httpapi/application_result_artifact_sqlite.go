package httpapi

import (
	"database/sql"
	"errors"
	"time"
)

type sqliteApplicationResultArtifactRepository struct{ database *sql.DB }

func newSQLiteApplicationResultArtifactRepository(database *sql.DB) *sqliteApplicationResultArtifactRepository {
	return &sqliteApplicationResultArtifactRepository{database: database}
}

type sqliteApplicationResultArtifactRow interface {
	Scan(...any) error
}

func (repository *sqliteApplicationResultArtifactRepository) Create(
	ctx ApplicationInteractionContext,
	artifact ApplicationResultArtifact,
) (ApplicationResultArtifact, bool, error) {
	if repository == nil || repository.database == nil {
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
	result, err := repository.database.ExecContext(applicationInteractionRequestContext(ctx), `INSERT INTO application_result_artifacts
(tenant_ref,workspace_id,application_id,owner_subject_ref,artifact_id,session_id,turn_id,client_turn_key,
execution_profile,run_id,run_schema_version,content_type,content_bytes,content_digest,created_at_unix_nano,artifact_payload)
VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT DO NOTHING`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, artifact.ArtifactID,
		artifact.SessionID, artifact.TurnID, artifact.ClientTurnKey, artifact.ExecutionProfile,
		artifact.RunRef.RunID, artifact.RunRef.SchemaVersion, artifact.ContentType, artifact.ContentBytes,
		artifact.ContentDigest, createdAt.UnixNano(), string(payload))
	if err != nil {
		return ApplicationResultArtifact{}, false, errApplicationResultArtifactStore
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return ApplicationResultArtifact{}, false, errApplicationResultArtifactStore
	}
	if affected == 1 {
		return cloneApplicationResultArtifact(artifact), false, nil
	}
	if affected != 0 {
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

func (repository *sqliteApplicationResultArtifactRepository) Read(
	ctx ApplicationInteractionContext,
	artifactID string,
) (ApplicationResultArtifact, error) {
	if repository == nil || repository.database == nil {
		return ApplicationResultArtifact{}, errApplicationResultArtifactStore
	}
	return readSQLiteApplicationResultArtifact(repository.database.QueryRowContext(applicationInteractionRequestContext(ctx), `SELECT artifact_id,session_id,turn_id,client_turn_key,execution_profile,run_id,run_schema_version,
content_type,content_bytes,content_digest,created_at_unix_nano,artifact_payload FROM application_result_artifacts
WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND artifact_id=?`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, artifactID), ctx)
}

func (repository *sqliteApplicationResultArtifactRepository) ReadByTurn(
	ctx ApplicationInteractionContext,
	sessionID string,
	turnID string,
) (ApplicationResultArtifact, error) {
	if repository == nil || repository.database == nil {
		return ApplicationResultArtifact{}, errApplicationResultArtifactStore
	}
	return readSQLiteApplicationResultArtifact(repository.database.QueryRowContext(applicationInteractionRequestContext(ctx), `SELECT artifact_id,session_id,turn_id,client_turn_key,execution_profile,run_id,run_schema_version,
content_type,content_bytes,content_digest,created_at_unix_nano,artifact_payload FROM application_result_artifacts
WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND session_id=? AND turn_id=?`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, sessionID, turnID), ctx)
}

func (repository *sqliteApplicationResultArtifactRepository) List(
	ctx ApplicationInteractionContext,
	sessionID string,
) ([]ApplicationResultArtifact, error) {
	if repository == nil || repository.database == nil {
		return nil, errApplicationResultArtifactStore
	}
	rows, err := repository.database.QueryContext(applicationInteractionRequestContext(ctx), `SELECT artifact_id,session_id,turn_id,client_turn_key,execution_profile,run_id,run_schema_version,
content_type,content_bytes,content_digest,created_at_unix_nano,artifact_payload FROM application_result_artifacts
WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND session_id=?
ORDER BY created_at_unix_nano DESC,artifact_id DESC`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, sessionID)
	if err != nil {
		return nil, errApplicationResultArtifactStore
	}
	defer rows.Close()
	artifacts := make([]ApplicationResultArtifact, 0)
	for rows.Next() {
		artifact, scanErr := readSQLiteApplicationResultArtifact(rows, ctx)
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

func readSQLiteApplicationResultArtifact(
	row sqliteApplicationResultArtifactRow,
	ctx ApplicationInteractionContext,
) (ApplicationResultArtifact, error) {
	var artifactID, sessionID, turnID, clientTurnKey, executionProfile string
	var runID, runSchemaVersion, contentType, contentDigest, payload string
	var contentBytes int
	var createdAtUnixNano int64
	if err := row.Scan(&artifactID, &sessionID, &turnID, &clientTurnKey, &executionProfile, &runID,
		&runSchemaVersion, &contentType, &contentBytes, &contentDigest, &createdAtUnixNano, &payload); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ApplicationResultArtifact{}, errApplicationResultArtifactNotFound
		}
		return ApplicationResultArtifact{}, errApplicationResultArtifactStore
	}
	artifact, err := decodeApplicationResultArtifact(ctx, []byte(payload))
	if err != nil || !applicationResultArtifactProjectionMatches(
		artifact, artifactID, sessionID, turnID, clientTurnKey, executionProfile, runID,
		runSchemaVersion, contentType, contentBytes, contentDigest, time.Unix(0, createdAtUnixNano).UTC(), false,
	) {
		return ApplicationResultArtifact{}, errApplicationResultArtifactContract
	}
	return artifact, nil
}
