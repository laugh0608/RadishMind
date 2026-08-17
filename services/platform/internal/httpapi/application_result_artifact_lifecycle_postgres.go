package httpapi

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

func (repository *postgresApplicationResultArtifactRepository) ReadLifecycle(
	ctx ApplicationInteractionContext,
	artifactID string,
) (ApplicationResultArtifactLifecycle, error) {
	if repository == nil || repository.pool == nil {
		return ApplicationResultArtifactLifecycle{}, errApplicationResultArtifactStore
	}
	return readPostgresApplicationResultArtifactLifecycle(repository.pool.QueryRow(
		applicationInteractionRequestContext(ctx),
		`SELECT artifact_id,lifecycle_state,lifecycle_version,archived_at,updated_at,
updated_by_actor_ref,request_id,audit_ref,lifecycle_payload
FROM application_result_artifact_lifecycles
WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 AND artifact_id=$5`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, artifactID,
	), ctx)
}

func (repository *postgresApplicationResultArtifactRepository) ListByLifecycle(
	ctx ApplicationInteractionContext,
	sessionID string,
	state ApplicationResultArtifactLifecycleState,
) ([]applicationResultArtifactStoredRecord, error) {
	if repository == nil || repository.pool == nil {
		return nil, errApplicationResultArtifactStore
	}
	rows, err := repository.pool.Query(applicationInteractionRequestContext(ctx), `SELECT
a.artifact_id,a.session_id,a.turn_id,a.client_turn_key,a.execution_profile,a.run_id,a.run_schema_version,
a.content_type,a.content_bytes,a.content_digest,a.created_at,a.artifact_payload,
l.lifecycle_state,l.lifecycle_version,l.archived_at,l.updated_at,
l.updated_by_actor_ref,l.request_id,l.audit_ref,l.lifecycle_payload
FROM application_result_artifacts a
JOIN application_result_artifact_lifecycles l
  ON l.tenant_ref=a.tenant_ref AND l.workspace_id=a.workspace_id AND l.application_id=a.application_id
 AND l.owner_subject_ref=a.owner_subject_ref AND l.artifact_id=a.artifact_id
WHERE a.tenant_ref=$1 AND a.workspace_id=$2 AND a.application_id=$3 AND a.owner_subject_ref=$4
  AND a.session_id=$5 AND l.lifecycle_state=$6
ORDER BY a.created_at DESC,a.artifact_id DESC`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, sessionID, state)
	if err != nil {
		return nil, errApplicationResultArtifactStore
	}
	defer rows.Close()
	records := make([]applicationResultArtifactStoredRecord, 0)
	for rows.Next() {
		record, scanErr := readPostgresApplicationResultArtifactStoredRecord(rows, ctx)
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

func (repository *postgresApplicationResultArtifactRepository) TransitionLifecycle(
	ctx ApplicationInteractionContext,
	artifactID string,
	target ApplicationResultArtifactLifecycleState,
	expectedVersion int,
	now time.Time,
) (ApplicationResultArtifactLifecycle, ApplicationResultArtifactLifecycleEvent, error) {
	if repository == nil || repository.pool == nil {
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
	updated, event := nextApplicationResultArtifactLifecycle(ctx, current, target, now.UTC().Round(time.Microsecond))
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
	var archivedAt *time.Time
	if updated.ArchivedAt != nil {
		archivedAt = parseApplicationInteractionTimestamp(*updated.ArchivedAt)
		if archivedAt == nil {
			return current, ApplicationResultArtifactLifecycleEvent{}, errApplicationResultArtifactContract
		}
	}
	transaction, err := repository.pool.Begin(applicationInteractionRequestContext(ctx))
	if err != nil {
		return current, ApplicationResultArtifactLifecycleEvent{}, errApplicationResultArtifactStore
	}
	defer transaction.Rollback(context.Background())
	command, err := transaction.Exec(applicationInteractionRequestContext(ctx), `UPDATE application_result_artifact_lifecycles
SET lifecycle_state=$1,lifecycle_version=$2,archived_at=$3,updated_at=$4,
    updated_by_actor_ref=$5,request_id=$6,audit_ref=$7,lifecycle_payload=$8
WHERE tenant_ref=$9 AND workspace_id=$10 AND application_id=$11 AND owner_subject_ref=$12 AND artifact_id=$13
  AND lifecycle_state=$14 AND lifecycle_version=$15`,
		updated.LifecycleState, updated.LifecycleVersion, archivedAt, updatedAt.UTC(), updated.UpdatedByActorRef,
		updated.RequestID, updated.AuditRef, lifecyclePayload, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID,
		ctx.OwnerSubjectRef, artifactID, current.LifecycleState, expectedVersion)
	if err != nil {
		return current, ApplicationResultArtifactLifecycleEvent{}, errApplicationResultArtifactStore
	}
	if command.RowsAffected() != 1 {
		_ = transaction.Rollback(applicationInteractionRequestContext(ctx))
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
	if _, err = transaction.Exec(applicationInteractionRequestContext(ctx), `INSERT INTO application_result_artifact_lifecycle_events
(tenant_ref,workspace_id,application_id,owner_subject_ref,artifact_id,lifecycle_version,from_state,to_state,
transition_kind,occurred_at,actor_ref,request_id,audit_ref,event_payload)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		event.TenantRef, event.WorkspaceID, event.ApplicationID, event.OwnerSubjectRef, event.ArtifactID,
		event.LifecycleVersion, event.FromState, event.ToState, event.TransitionKind, occurredAt.UTC(),
		event.ActorRef, event.RequestID, event.AuditRef, eventPayload); err != nil {
		return current, ApplicationResultArtifactLifecycleEvent{}, errApplicationResultArtifactStore
	}
	if err = transaction.Commit(applicationInteractionRequestContext(ctx)); err != nil {
		return current, ApplicationResultArtifactLifecycleEvent{}, errApplicationResultArtifactStore
	}
	return updated, event, nil
}

func readPostgresApplicationResultArtifactLifecycle(
	row postgresApplicationResultArtifactRow,
	ctx ApplicationInteractionContext,
) (ApplicationResultArtifactLifecycle, error) {
	var artifactID, state, updatedByActorRef, requestID, auditRef string
	var version int
	var archivedAt *time.Time
	var updatedAt time.Time
	var payload []byte
	if err := row.Scan(&artifactID, &state, &version, &archivedAt, &updatedAt,
		&updatedByActorRef, &requestID, &auditRef, &payload); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ApplicationResultArtifactLifecycle{}, errApplicationResultArtifactNotFound
		}
		return ApplicationResultArtifactLifecycle{}, errApplicationResultArtifactStore
	}
	lifecycle, err := decodeApplicationResultArtifactLifecycle(ctx, payload)
	if err != nil || !applicationResultArtifactLifecycleProjectionMatches(
		lifecycle, artifactID, ApplicationResultArtifactLifecycleState(state), version, archivedAt,
		updatedAt, updatedByActorRef, requestID, auditRef, true,
	) {
		return ApplicationResultArtifactLifecycle{}, errApplicationResultArtifactContract
	}
	return lifecycle, nil
}

func readPostgresApplicationResultArtifactStoredRecord(
	row postgresApplicationResultArtifactRow,
	ctx ApplicationInteractionContext,
) (applicationResultArtifactStoredRecord, error) {
	var artifactID, sessionID, turnID, clientTurnKey, executionProfile string
	var runID, runSchemaVersion, contentType, contentDigest string
	var contentBytes int
	var createdAt time.Time
	var artifactPayload []byte
	var lifecycleState, updatedByActorRef, requestID, auditRef string
	var lifecycleVersion int
	var archivedAt *time.Time
	var updatedAt time.Time
	var lifecyclePayload []byte
	if err := row.Scan(&artifactID, &sessionID, &turnID, &clientTurnKey, &executionProfile, &runID,
		&runSchemaVersion, &contentType, &contentBytes, &contentDigest, &createdAt, &artifactPayload,
		&lifecycleState, &lifecycleVersion, &archivedAt, &updatedAt,
		&updatedByActorRef, &requestID, &auditRef, &lifecyclePayload); err != nil {
		return applicationResultArtifactStoredRecord{}, errApplicationResultArtifactStore
	}
	artifact, err := decodeApplicationResultArtifact(ctx, artifactPayload)
	if err != nil || !applicationResultArtifactProjectionMatches(
		artifact, artifactID, sessionID, turnID, clientTurnKey, executionProfile, runID, runSchemaVersion,
		contentType, contentBytes, contentDigest, createdAt, true,
	) {
		return applicationResultArtifactStoredRecord{}, errApplicationResultArtifactContract
	}
	lifecycle, err := decodeApplicationResultArtifactLifecycle(ctx, lifecyclePayload)
	if err != nil || !applicationResultArtifactLifecycleProjectionMatches(
		lifecycle, artifactID, ApplicationResultArtifactLifecycleState(lifecycleState), lifecycleVersion,
		archivedAt, updatedAt, updatedByActorRef, requestID, auditRef, true,
	) {
		return applicationResultArtifactStoredRecord{}, errApplicationResultArtifactContract
	}
	return applicationResultArtifactStoredRecord{Artifact: artifact, Lifecycle: lifecycle}, nil
}

var _ applicationResultArtifactRepository = (*postgresApplicationResultArtifactRepository)(nil)
