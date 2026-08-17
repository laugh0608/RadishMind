package httpapi

import (
	"database/sql"
	"errors"
	"time"
)

func (repository *sqliteApplicationResultArtifactRepository) ReadLifecycle(
	ctx ApplicationInteractionContext,
	artifactID string,
) (ApplicationResultArtifactLifecycle, error) {
	if repository == nil || repository.database == nil {
		return ApplicationResultArtifactLifecycle{}, errApplicationResultArtifactStore
	}
	return readSQLiteApplicationResultArtifactLifecycle(repository.database.QueryRowContext(
		applicationInteractionRequestContext(ctx),
		`SELECT artifact_id,lifecycle_state,lifecycle_version,archived_at_unix_nano,updated_at_unix_nano,
updated_by_actor_ref,request_id,audit_ref,lifecycle_payload
FROM application_result_artifact_lifecycles
WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND artifact_id=?`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, artifactID,
	), ctx)
}

func (repository *sqliteApplicationResultArtifactRepository) ListByLifecycle(
	ctx ApplicationInteractionContext,
	sessionID string,
	state ApplicationResultArtifactLifecycleState,
) ([]applicationResultArtifactStoredRecord, error) {
	if repository == nil || repository.database == nil {
		return nil, errApplicationResultArtifactStore
	}
	rows, err := repository.database.QueryContext(applicationInteractionRequestContext(ctx), `SELECT
a.artifact_id,a.session_id,a.turn_id,a.client_turn_key,a.execution_profile,a.run_id,a.run_schema_version,
a.content_type,a.content_bytes,a.content_digest,a.created_at_unix_nano,a.artifact_payload,
l.lifecycle_state,l.lifecycle_version,l.archived_at_unix_nano,l.updated_at_unix_nano,
l.updated_by_actor_ref,l.request_id,l.audit_ref,l.lifecycle_payload
FROM application_result_artifacts a
JOIN application_result_artifact_lifecycles l
  ON l.tenant_ref=a.tenant_ref AND l.workspace_id=a.workspace_id AND l.application_id=a.application_id
 AND l.owner_subject_ref=a.owner_subject_ref AND l.artifact_id=a.artifact_id
WHERE a.tenant_ref=? AND a.workspace_id=? AND a.application_id=? AND a.owner_subject_ref=?
  AND a.session_id=? AND l.lifecycle_state=?
ORDER BY a.created_at_unix_nano DESC,a.artifact_id DESC`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, sessionID, state)
	if err != nil {
		return nil, errApplicationResultArtifactStore
	}
	defer rows.Close()
	records := make([]applicationResultArtifactStoredRecord, 0)
	for rows.Next() {
		record, scanErr := readSQLiteApplicationResultArtifactStoredRecord(rows, ctx)
		if scanErr != nil {
			return nil, scanErr
		}
		records = append(records, record)
	}
	if rows.Err() != nil {
		return nil, errApplicationResultArtifactStore
	}
	return records, nil
}

func (repository *sqliteApplicationResultArtifactRepository) TransitionLifecycle(
	ctx ApplicationInteractionContext,
	artifactID string,
	target ApplicationResultArtifactLifecycleState,
	expectedVersion int,
	now time.Time,
) (ApplicationResultArtifactLifecycle, ApplicationResultArtifactLifecycleEvent, error) {
	if repository == nil || repository.database == nil {
		return ApplicationResultArtifactLifecycle{}, ApplicationResultArtifactLifecycleEvent{}, errApplicationResultArtifactStore
	}
	current, err := repository.ReadLifecycle(ctx, artifactID)
	if err != nil {
		return ApplicationResultArtifactLifecycle{}, ApplicationResultArtifactLifecycleEvent{}, err
	}
	if current.LifecycleVersion != expectedVersion {
		return current, ApplicationResultArtifactLifecycleEvent{}, errApplicationResultArtifactLifecycleVersion
	}
	if current.LifecycleState == target || !validApplicationResultArtifactLifecycleState(target) {
		return current, ApplicationResultArtifactLifecycleEvent{}, errApplicationResultArtifactLifecycleState
	}
	updated, event := nextApplicationResultArtifactLifecycle(ctx, current, target, now)
	lifecyclePayload, err := encodeApplicationResultArtifactLifecycle(ctx, updated)
	if err != nil {
		return current, ApplicationResultArtifactLifecycleEvent{}, err
	}
	eventPayload, err := encodeApplicationResultArtifactLifecycleEvent(ctx, event)
	if err != nil {
		return current, ApplicationResultArtifactLifecycleEvent{}, err
	}
	updatedAt := parseApplicationInteractionTimestamp(updated.UpdatedAt)
	if updatedAt == nil {
		return current, ApplicationResultArtifactLifecycleEvent{}, errApplicationResultArtifactContract
	}
	archivedAt := any(nil)
	if updated.ArchivedAt != nil {
		parsed := parseApplicationInteractionTimestamp(*updated.ArchivedAt)
		if parsed == nil {
			return current, ApplicationResultArtifactLifecycleEvent{}, errApplicationResultArtifactContract
		}
		archivedAt = parsed.UnixNano()
	}
	transaction, err := repository.database.BeginTx(applicationInteractionRequestContext(ctx), nil)
	if err != nil {
		return current, ApplicationResultArtifactLifecycleEvent{}, errApplicationResultArtifactStore
	}
	defer transaction.Rollback()
	result, err := transaction.ExecContext(applicationInteractionRequestContext(ctx), `UPDATE application_result_artifact_lifecycles
SET lifecycle_state=?,lifecycle_version=?,archived_at_unix_nano=?,updated_at_unix_nano=?,
    updated_by_actor_ref=?,request_id=?,audit_ref=?,lifecycle_payload=?
WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND artifact_id=?
  AND lifecycle_state=? AND lifecycle_version=?`,
		updated.LifecycleState, updated.LifecycleVersion, archivedAt, updatedAt.UnixNano(), updated.UpdatedByActorRef,
		updated.RequestID, updated.AuditRef, string(lifecyclePayload), ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID,
		ctx.OwnerSubjectRef, artifactID, current.LifecycleState, expectedVersion)
	if err != nil {
		return current, ApplicationResultArtifactLifecycleEvent{}, errApplicationResultArtifactStore
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return current, ApplicationResultArtifactLifecycleEvent{}, errApplicationResultArtifactStore
	}
	if affected != 1 {
		_ = transaction.Rollback()
		latest, readErr := repository.ReadLifecycle(ctx, artifactID)
		if readErr != nil {
			return current, ApplicationResultArtifactLifecycleEvent{}, readErr
		}
		if latest.LifecycleVersion != expectedVersion {
			return latest, ApplicationResultArtifactLifecycleEvent{}, errApplicationResultArtifactLifecycleVersion
		}
		return latest, ApplicationResultArtifactLifecycleEvent{}, errApplicationResultArtifactLifecycleState
	}
	occurredAt := parseApplicationInteractionTimestamp(event.OccurredAt)
	if occurredAt == nil {
		return current, ApplicationResultArtifactLifecycleEvent{}, errApplicationResultArtifactContract
	}
	if _, err = transaction.ExecContext(applicationInteractionRequestContext(ctx), `INSERT INTO application_result_artifact_lifecycle_events
(tenant_ref,workspace_id,application_id,owner_subject_ref,artifact_id,lifecycle_version,from_state,to_state,
transition_kind,occurred_at_unix_nano,actor_ref,request_id,audit_ref,event_payload)
VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		event.TenantRef, event.WorkspaceID, event.ApplicationID, event.OwnerSubjectRef, event.ArtifactID,
		event.LifecycleVersion, event.FromState, event.ToState, event.TransitionKind, occurredAt.UnixNano(),
		event.ActorRef, event.RequestID, event.AuditRef, string(eventPayload)); err != nil {
		return current, ApplicationResultArtifactLifecycleEvent{}, errApplicationResultArtifactStore
	}
	if err = transaction.Commit(); err != nil {
		return current, ApplicationResultArtifactLifecycleEvent{}, errApplicationResultArtifactStore
	}
	return updated, event, nil
}

func readSQLiteApplicationResultArtifactLifecycle(
	row sqliteApplicationResultArtifactRow,
	ctx ApplicationInteractionContext,
) (ApplicationResultArtifactLifecycle, error) {
	var artifactID, state, updatedByActorRef, requestID, auditRef, payload string
	var version int
	var archivedAtUnixNano sql.NullInt64
	var updatedAtUnixNano int64
	if err := row.Scan(&artifactID, &state, &version, &archivedAtUnixNano, &updatedAtUnixNano,
		&updatedByActorRef, &requestID, &auditRef, &payload); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ApplicationResultArtifactLifecycle{}, errApplicationResultArtifactNotFound
		}
		return ApplicationResultArtifactLifecycle{}, errApplicationResultArtifactStore
	}
	lifecycle, err := decodeApplicationResultArtifactLifecycle(ctx, []byte(payload))
	var archivedAt *time.Time
	if archivedAtUnixNano.Valid {
		value := time.Unix(0, archivedAtUnixNano.Int64).UTC()
		archivedAt = &value
	}
	if err != nil || !applicationResultArtifactLifecycleProjectionMatches(
		lifecycle, artifactID, ApplicationResultArtifactLifecycleState(state), version, archivedAt,
		time.Unix(0, updatedAtUnixNano).UTC(), updatedByActorRef, requestID, auditRef, false,
	) {
		return ApplicationResultArtifactLifecycle{}, errApplicationResultArtifactContract
	}
	return lifecycle, nil
}

func readSQLiteApplicationResultArtifactStoredRecord(
	row sqliteApplicationResultArtifactRow,
	ctx ApplicationInteractionContext,
) (applicationResultArtifactStoredRecord, error) {
	var artifactID, sessionID, turnID, clientTurnKey, executionProfile string
	var runID, runSchemaVersion, contentType, contentDigest, artifactPayload string
	var contentBytes int
	var createdAtUnixNano int64
	var lifecycleState, updatedByActorRef, requestID, auditRef, lifecyclePayload string
	var lifecycleVersion int
	var archivedAtUnixNano sql.NullInt64
	var updatedAtUnixNano int64
	if err := row.Scan(&artifactID, &sessionID, &turnID, &clientTurnKey, &executionProfile, &runID,
		&runSchemaVersion, &contentType, &contentBytes, &contentDigest, &createdAtUnixNano, &artifactPayload,
		&lifecycleState, &lifecycleVersion, &archivedAtUnixNano, &updatedAtUnixNano,
		&updatedByActorRef, &requestID, &auditRef, &lifecyclePayload); err != nil {
		return applicationResultArtifactStoredRecord{}, errApplicationResultArtifactStore
	}
	artifact, err := decodeApplicationResultArtifact(ctx, []byte(artifactPayload))
	if err != nil || !applicationResultArtifactProjectionMatches(
		artifact, artifactID, sessionID, turnID, clientTurnKey, executionProfile, runID, runSchemaVersion,
		contentType, contentBytes, contentDigest, time.Unix(0, createdAtUnixNano).UTC(), false,
	) {
		return applicationResultArtifactStoredRecord{}, errApplicationResultArtifactContract
	}
	var archivedAt *time.Time
	if archivedAtUnixNano.Valid {
		value := time.Unix(0, archivedAtUnixNano.Int64).UTC()
		archivedAt = &value
	}
	lifecycle, err := decodeApplicationResultArtifactLifecycle(ctx, []byte(lifecyclePayload))
	if err != nil || !applicationResultArtifactLifecycleProjectionMatches(
		lifecycle, artifactID, ApplicationResultArtifactLifecycleState(lifecycleState), lifecycleVersion,
		archivedAt, time.Unix(0, updatedAtUnixNano).UTC(), updatedByActorRef, requestID, auditRef, false,
	) {
		return applicationResultArtifactStoredRecord{}, errApplicationResultArtifactContract
	}
	return applicationResultArtifactStoredRecord{Artifact: artifact, Lifecycle: lifecycle}, nil
}

var _ applicationResultArtifactRepository = (*sqliteApplicationResultArtifactRepository)(nil)
