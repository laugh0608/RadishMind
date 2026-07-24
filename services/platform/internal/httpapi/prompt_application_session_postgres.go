package httpapi

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresPromptApplicationSessionRepository struct{ pool *pgxpool.Pool }

func newPostgresPromptApplicationSessionRepository(pool *pgxpool.Pool) *postgresPromptApplicationSessionRepository {
	return &postgresPromptApplicationSessionRepository{pool: pool}
}

func (repository *postgresPromptApplicationSessionRepository) Create(ctx ApplicationInteractionContext, session ApplicationInteractionSession) (ApplicationInteractionSession, error) {
	if repository == nil || repository.pool == nil || session.SchemaVersion != promptApplicationSessionV2Schema {
		return ApplicationInteractionSession{}, errApplicationSessionStore
	}
	payload, err := encodeApplicationInteractionSession(session)
	if err != nil {
		return ApplicationInteractionSession{}, err
	}
	updatedAt, err := applicationInteractionTimestamp(session.UpdatedAt)
	if err != nil {
		return ApplicationInteractionSession{}, err
	}
	command, err := repository.pool.Exec(applicationInteractionRequestContext(ctx), `INSERT INTO prompt_application_sessions
(tenant_ref,workspace_id,application_id,owner_subject_ref,session_id,session_state,record_version,updated_at,authority_digest,sanitized_session_payload)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) ON CONFLICT DO NOTHING`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, session.SessionID,
		session.State, session.RecordVersion, updatedAt, session.Authority.AuthorityDigest, payload)
	if err != nil {
		return ApplicationInteractionSession{}, errApplicationSessionStore
	}
	if command.RowsAffected() != 1 {
		return ApplicationInteractionSession{}, errApplicationSessionVersionConflict
	}
	return cloneApplicationInteractionSession(session), nil
}

func (repository *postgresPromptApplicationSessionRepository) Read(ctx ApplicationInteractionContext, sessionID string) (ApplicationInteractionSession, error) {
	if repository == nil || repository.pool == nil {
		return ApplicationInteractionSession{}, errApplicationSessionStore
	}
	return readPostgresPromptApplicationSession(repository.pool.QueryRow(applicationInteractionRequestContext(ctx), `SELECT
session_state,record_version,updated_at,authority_digest,sanitized_session_payload FROM prompt_application_sessions
WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 AND session_id=$5`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, sessionID), ctx)
}

func (repository *postgresPromptApplicationSessionRepository) List(ctx ApplicationInteractionContext, query applicationInteractionSessionListQuery) ([]ApplicationInteractionSession, error) {
	if repository == nil || repository.pool == nil {
		return nil, errApplicationSessionStore
	}
	if query.ExecutionProfile != "" && query.ExecutionProfile != applicationInteractionProfilePrompt {
		return []ApplicationInteractionSession{}, nil
	}
	var after any
	if query.AfterUpdatedAt != "" {
		parsed, err := applicationInteractionTimestamp(query.AfterUpdatedAt)
		if err != nil {
			return nil, errApplicationSessionContract
		}
		after = parsed
	}
	rows, err := repository.pool.Query(applicationInteractionRequestContext(ctx), `SELECT
session_state,record_version,updated_at,authority_digest,sanitized_session_payload FROM prompt_application_sessions
WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4
AND ($5='' OR session_state=$5) AND ($6::timestamptz IS NULL OR updated_at<$6 OR (updated_at=$6 AND session_id<$7))
ORDER BY updated_at DESC,session_id DESC LIMIT $8`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef,
		query.State, after, query.AfterSessionID, query.Limit)
	if err != nil {
		return nil, errApplicationSessionStore
	}
	defer rows.Close()
	result := make([]ApplicationInteractionSession, 0, query.Limit)
	for rows.Next() {
		session, scanErr := readPostgresPromptApplicationSession(rows, ctx)
		if scanErr != nil {
			return nil, scanErr
		}
		result = append(result, session)
	}
	if rows.Err() != nil {
		return nil, errApplicationSessionStore
	}
	return result, nil
}

func (repository *postgresPromptApplicationSessionRepository) Close(ctx ApplicationInteractionContext, sessionID string, expectedVersion int, updated ApplicationInteractionSession) (ApplicationInteractionSession, error) {
	var output ApplicationInteractionSession
	err := repository.mutate(ctx, func(tx pgx.Tx) error {
		current, err := readPostgresPromptApplicationSession(tx.QueryRow(applicationInteractionRequestContext(ctx), `SELECT
session_state,record_version,updated_at,authority_digest,sanitized_session_payload FROM prompt_application_sessions
WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 AND session_id=$5 FOR UPDATE`,
			ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, sessionID), ctx)
		if err != nil {
			return err
		}
		if current.State != applicationSessionStateActive {
			return errApplicationSessionClosed
		}
		if current.RecordVersion != expectedVersion {
			return errApplicationSessionVersionConflict
		}
		if validateApplicationInteractionSessionTransition(current, updated) != nil {
			return errApplicationSessionContract
		}
		payload, err := encodeApplicationInteractionSession(updated)
		if err != nil {
			return err
		}
		updatedAt, _ := applicationInteractionTimestamp(updated.UpdatedAt)
		command, err := tx.Exec(applicationInteractionRequestContext(ctx), `UPDATE prompt_application_sessions SET
session_state=$1,record_version=$2,updated_at=$3,sanitized_session_payload=$4 WHERE tenant_ref=$5 AND workspace_id=$6
AND application_id=$7 AND owner_subject_ref=$8 AND session_id=$9 AND record_version=$10 AND session_state='active'`,
			updated.State, updated.RecordVersion, updatedAt, payload, ctx.TenantRef, ctx.WorkspaceID,
			ctx.ApplicationID, ctx.OwnerSubjectRef, sessionID, expectedVersion)
		if err != nil {
			return errApplicationSessionStore
		}
		if command.RowsAffected() != 1 {
			return errApplicationSessionVersionConflict
		}
		output = cloneApplicationInteractionSession(updated)
		return nil
	})
	return output, err
}

func (repository *postgresPromptApplicationSessionRepository) ReserveTurn(ctx ApplicationInteractionContext, expectedVersion int, updated ApplicationInteractionSession, turn ApplicationInteractionTurn) (ApplicationInteractionSession, ApplicationInteractionTurn, bool, error) {
	var outputSession ApplicationInteractionSession
	var outputTurn ApplicationInteractionTurn
	var replay bool
	err := repository.mutate(ctx, func(tx pgx.Tx) error {
		current, err := readPostgresPromptApplicationSession(tx.QueryRow(applicationInteractionRequestContext(ctx), `SELECT
session_state,record_version,updated_at,authority_digest,sanitized_session_payload FROM prompt_application_sessions
WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 AND session_id=$5 FOR UPDATE`,
			ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, updated.SessionID), ctx)
		if err != nil {
			return err
		}
		existing, err := readPostgresApplicationInteractionTurn(tx.QueryRow(applicationInteractionRequestContext(ctx), `SELECT
turn_sequence,client_turn_key,turn_status,started_at,completed_at,sanitized_turn_payload FROM prompt_application_session_turns
WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 AND session_id=$5 AND client_turn_key=$6`,
			ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, updated.SessionID, turn.ClientTurnKey), ctx)
		if err == nil {
			if !applicationInteractionTurnReservationsEqual(existing, turn) {
				return errApplicationSessionIdempotency
			}
			outputSession, outputTurn, replay = current, existing, true
			return nil
		}
		if !errors.Is(err, errApplicationSessionNotFound) {
			return err
		}
		if current.State != applicationSessionStateActive {
			return errApplicationSessionClosed
		}
		if current.RecordVersion != expectedVersion {
			return errApplicationSessionVersionConflict
		}
		if validateApplicationInteractionSessionReserve(current, updated, turn) != nil {
			return errApplicationSessionContract
		}
		sessionPayload, err := encodeApplicationInteractionSession(updated)
		if err != nil {
			return err
		}
		turnPayload, err := encodeApplicationInteractionTurn(turn)
		if err != nil {
			return err
		}
		updatedAt, _ := applicationInteractionTimestamp(updated.UpdatedAt)
		startedAt, _ := applicationInteractionTimestamp(turn.StartedAt)
		command, err := tx.Exec(applicationInteractionRequestContext(ctx), `UPDATE prompt_application_sessions SET
record_version=$1,updated_at=$2,sanitized_session_payload=$3 WHERE tenant_ref=$4 AND workspace_id=$5
AND application_id=$6 AND owner_subject_ref=$7 AND session_id=$8 AND record_version=$9 AND session_state='active'`,
			updated.RecordVersion, updatedAt, sessionPayload, ctx.TenantRef, ctx.WorkspaceID,
			ctx.ApplicationID, ctx.OwnerSubjectRef, updated.SessionID, expectedVersion)
		if err != nil {
			return errApplicationSessionStore
		}
		if command.RowsAffected() != 1 {
			return errApplicationSessionVersionConflict
		}
		_, err = tx.Exec(applicationInteractionRequestContext(ctx), `INSERT INTO prompt_application_session_turns
(tenant_ref,workspace_id,application_id,owner_subject_ref,session_id,turn_id,turn_sequence,client_turn_key,turn_status,started_at,completed_at,sanitized_turn_payload)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,NULL,$11)`,
			ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, turn.SessionID,
			turn.TurnID, turn.Sequence, turn.ClientTurnKey, turn.Status, startedAt, turnPayload)
		if err != nil {
			return errApplicationSessionStore
		}
		outputSession, outputTurn = cloneApplicationInteractionSession(updated), cloneApplicationInteractionTurn(turn)
		return nil
	})
	return outputSession, outputTurn, replay, err
}

func (repository *postgresPromptApplicationSessionRepository) ReadTurn(ctx ApplicationInteractionContext, sessionID, turnID string) (ApplicationInteractionTurn, error) {
	if repository == nil || repository.pool == nil {
		return ApplicationInteractionTurn{}, errApplicationSessionStore
	}
	return readPostgresApplicationInteractionTurn(repository.pool.QueryRow(applicationInteractionRequestContext(ctx), `SELECT
turn_sequence,client_turn_key,turn_status,started_at,completed_at,sanitized_turn_payload FROM prompt_application_session_turns
WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 AND session_id=$5 AND turn_id=$6`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, sessionID, turnID), ctx)
}

func (repository *postgresPromptApplicationSessionRepository) ReadTurnByClientKey(ctx ApplicationInteractionContext, sessionID, clientKey string) (ApplicationInteractionTurn, error) {
	if repository == nil || repository.pool == nil {
		return ApplicationInteractionTurn{}, errApplicationSessionStore
	}
	return readPostgresApplicationInteractionTurn(repository.pool.QueryRow(applicationInteractionRequestContext(ctx), `SELECT
turn_sequence,client_turn_key,turn_status,started_at,completed_at,sanitized_turn_payload FROM prompt_application_session_turns
WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 AND session_id=$5 AND client_turn_key=$6`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, sessionID, clientKey), ctx)
}

func (repository *postgresPromptApplicationSessionRepository) CompleteTurn(ctx ApplicationInteractionContext, terminal ApplicationInteractionTurn) (ApplicationInteractionSession, ApplicationInteractionTurn, bool, error) {
	var outputSession ApplicationInteractionSession
	var outputTurn ApplicationInteractionTurn
	var replay bool
	err := repository.mutate(ctx, func(tx pgx.Tx) error {
		current, err := readPostgresApplicationInteractionTurn(tx.QueryRow(applicationInteractionRequestContext(ctx), `SELECT
turn_sequence,client_turn_key,turn_status,started_at,completed_at,sanitized_turn_payload FROM prompt_application_session_turns
WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 AND session_id=$5 AND turn_id=$6 FOR UPDATE`,
			ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, terminal.SessionID, terminal.TurnID), ctx)
		if err != nil {
			return err
		}
		session, err := readPostgresPromptApplicationSession(tx.QueryRow(applicationInteractionRequestContext(ctx), `SELECT
session_state,record_version,updated_at,authority_digest,sanitized_session_payload FROM prompt_application_sessions
WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 AND session_id=$5`,
			ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, terminal.SessionID), ctx)
		if err != nil {
			return err
		}
		if current.Status != string(WorkflowRunStatusRunning) {
			if !applicationInteractionTurnsIdempotentlyEqual(current, terminal) {
				return errApplicationSessionIdempotency
			}
			outputSession, outputTurn, replay = session, current, true
			return nil
		}
		if validateApplicationInteractionTurnTransition(current, terminal) != nil {
			return errApplicationSessionContract
		}
		payload, err := encodeApplicationInteractionTurn(terminal)
		if err != nil {
			return err
		}
		completedAt, err := applicationInteractionCompletedTimestamp(terminal.CompletedAt)
		if err != nil || completedAt == nil {
			return errApplicationSessionContract
		}
		command, err := tx.Exec(applicationInteractionRequestContext(ctx), `UPDATE prompt_application_session_turns SET
turn_status=$1,completed_at=$2,sanitized_turn_payload=$3 WHERE tenant_ref=$4 AND workspace_id=$5 AND application_id=$6
AND owner_subject_ref=$7 AND session_id=$8 AND turn_id=$9 AND turn_status='running'`,
			terminal.Status, completedAt, payload, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID,
			ctx.OwnerSubjectRef, terminal.SessionID, terminal.TurnID)
		if err != nil {
			return errApplicationSessionStore
		}
		if command.RowsAffected() != 1 {
			return errApplicationSessionIdempotency
		}
		outputSession, outputTurn = session, cloneApplicationInteractionTurn(terminal)
		return nil
	})
	return outputSession, outputTurn, replay, err
}

func (repository *postgresPromptApplicationSessionRepository) ListTurns(ctx ApplicationInteractionContext, sessionID string) ([]ApplicationInteractionTurn, error) {
	if _, err := repository.Read(ctx, sessionID); err != nil {
		return nil, err
	}
	rows, err := repository.pool.Query(applicationInteractionRequestContext(ctx), `SELECT
turn_sequence,client_turn_key,turn_status,started_at,completed_at,sanitized_turn_payload FROM prompt_application_session_turns
WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 AND session_id=$5 ORDER BY turn_sequence`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, sessionID)
	if err != nil {
		return nil, errApplicationSessionStore
	}
	defer rows.Close()
	result := make([]ApplicationInteractionTurn, 0)
	for rows.Next() {
		turn, scanErr := readPostgresApplicationInteractionTurn(rows, ctx)
		if scanErr != nil {
			return nil, scanErr
		}
		result = append(result, turn)
	}
	if rows.Err() != nil {
		return nil, errApplicationSessionStore
	}
	return result, nil
}

func (repository *postgresPromptApplicationSessionRepository) mutate(ctx ApplicationInteractionContext, operation func(pgx.Tx) error) error {
	if repository == nil || repository.pool == nil {
		return errApplicationSessionStore
	}
	requestContext := applicationInteractionRequestContext(ctx)
	tx, err := repository.pool.Begin(requestContext)
	if err != nil {
		return errApplicationSessionStore
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	if err = operation(tx); err != nil {
		return err
	}
	if err = tx.Commit(requestContext); err != nil {
		return errApplicationSessionStore
	}
	return nil
}

func readPostgresPromptApplicationSession(row postgresApplicationInteractionSessionRow, ctx ApplicationInteractionContext) (ApplicationInteractionSession, error) {
	var state, authorityDigest string
	var version int
	var updatedAt time.Time
	var payload []byte
	if err := row.Scan(&state, &version, &updatedAt, &authorityDigest, &payload); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ApplicationInteractionSession{}, errApplicationSessionNotFound
		}
		return ApplicationInteractionSession{}, errApplicationSessionStore
	}
	session, err := decodeApplicationInteractionSession(ctx, payload)
	storedAt, timeErr := applicationInteractionTimestamp(session.UpdatedAt)
	if err != nil || timeErr != nil || session.State != state || session.RecordVersion != version ||
		session.Authority.AuthorityDigest != authorityDigest || !applicationInteractionPostgresTimesEqual(storedAt, updatedAt) {
		return ApplicationInteractionSession{}, errApplicationSessionContract
	}
	return session, nil
}

var _ applicationInteractionSessionRepository = (*postgresPromptApplicationSessionRepository)(nil)
