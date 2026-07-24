package httpapi

import (
	"context"
	"database/sql"
	"errors"
)

type sqlitePromptApplicationSessionRepository struct{ database *sql.DB }

func newSQLitePromptApplicationSessionRepository(database *sql.DB) *sqlitePromptApplicationSessionRepository {
	return &sqlitePromptApplicationSessionRepository{database: database}
}

func (repository *sqlitePromptApplicationSessionRepository) Create(ctx ApplicationInteractionContext, session ApplicationInteractionSession) (ApplicationInteractionSession, error) {
	if repository == nil || repository.database == nil || session.SchemaVersion != promptApplicationSessionV2Schema {
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
	result, err := repository.database.ExecContext(applicationInteractionRequestContext(ctx), `INSERT INTO prompt_application_sessions
(tenant_ref,workspace_id,application_id,owner_subject_ref,session_id,session_state,record_version,updated_at_unix_nano,authority_digest,sanitized_session_payload)
VALUES (?,?,?,?,?,?,?,?,?,?) ON CONFLICT DO NOTHING`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, session.SessionID,
		session.State, session.RecordVersion, updatedAt.UnixNano(), session.Authority.AuthorityDigest, string(payload))
	if err != nil {
		return ApplicationInteractionSession{}, errApplicationSessionStore
	}
	affected, err := result.RowsAffected()
	if err != nil || affected != 1 {
		return ApplicationInteractionSession{}, errApplicationSessionVersionConflict
	}
	return cloneApplicationInteractionSession(session), nil
}

func (repository *sqlitePromptApplicationSessionRepository) Read(ctx ApplicationInteractionContext, sessionID string) (ApplicationInteractionSession, error) {
	if repository == nil || repository.database == nil {
		return ApplicationInteractionSession{}, errApplicationSessionStore
	}
	return readSQLitePromptApplicationSession(repository.database.QueryRowContext(applicationInteractionRequestContext(ctx), `SELECT session_state,record_version,updated_at_unix_nano,authority_digest,sanitized_session_payload
FROM prompt_application_sessions WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND session_id=?`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, sessionID), ctx)
}

func (repository *sqlitePromptApplicationSessionRepository) List(ctx ApplicationInteractionContext, query applicationInteractionSessionListQuery) ([]ApplicationInteractionSession, error) {
	if repository == nil || repository.database == nil {
		return nil, errApplicationSessionStore
	}
	if query.ExecutionProfile != "" && query.ExecutionProfile != applicationInteractionProfilePrompt {
		return []ApplicationInteractionSession{}, nil
	}
	statement := `SELECT session_state,record_version,updated_at_unix_nano,authority_digest,sanitized_session_payload FROM prompt_application_sessions
WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=?`
	args := []any{ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef}
	if query.State != "" {
		statement += " AND session_state=?"
		args = append(args, query.State)
	}
	if query.AfterUpdatedAt != "" {
		after, err := applicationInteractionTimestamp(query.AfterUpdatedAt)
		if err != nil {
			return nil, errApplicationSessionContract
		}
		statement += " AND (updated_at_unix_nano<? OR (updated_at_unix_nano=? AND session_id<?))"
		args = append(args, after.UnixNano(), after.UnixNano(), query.AfterSessionID)
	}
	statement += " ORDER BY updated_at_unix_nano DESC,session_id DESC LIMIT ?"
	args = append(args, query.Limit)
	rows, err := repository.database.QueryContext(applicationInteractionRequestContext(ctx), statement, args...)
	if err != nil {
		return nil, errApplicationSessionStore
	}
	defer rows.Close()
	result := make([]ApplicationInteractionSession, 0, query.Limit)
	for rows.Next() {
		session, scanErr := readSQLitePromptApplicationSession(rows, ctx)
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

func (repository *sqlitePromptApplicationSessionRepository) Close(ctx ApplicationInteractionContext, sessionID string, expectedVersion int, updated ApplicationInteractionSession) (ApplicationInteractionSession, error) {
	var output ApplicationInteractionSession
	err := repository.mutate(ctx, func(connection *sql.Conn) error {
		current, err := readSQLitePromptApplicationSession(connection.QueryRowContext(applicationInteractionRequestContext(ctx), `SELECT session_state,record_version,updated_at_unix_nano,authority_digest,sanitized_session_payload
FROM prompt_application_sessions WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND session_id=?`,
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
		result, err := connection.ExecContext(applicationInteractionRequestContext(ctx), `UPDATE prompt_application_sessions SET
session_state=?,record_version=?,updated_at_unix_nano=?,sanitized_session_payload=?
WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND session_id=? AND record_version=? AND session_state='active'`,
			updated.State, updated.RecordVersion, updatedAt.UnixNano(), string(payload), ctx.TenantRef, ctx.WorkspaceID,
			ctx.ApplicationID, ctx.OwnerSubjectRef, sessionID, expectedVersion)
		if err != nil {
			return errApplicationSessionStore
		}
		affected, _ := result.RowsAffected()
		if affected != 1 {
			return errApplicationSessionVersionConflict
		}
		output = cloneApplicationInteractionSession(updated)
		return nil
	})
	return output, err
}

func (repository *sqlitePromptApplicationSessionRepository) ReserveTurn(ctx ApplicationInteractionContext, expectedVersion int, updated ApplicationInteractionSession, turn ApplicationInteractionTurn) (ApplicationInteractionSession, ApplicationInteractionTurn, bool, error) {
	var outputSession ApplicationInteractionSession
	var outputTurn ApplicationInteractionTurn
	var replay bool
	err := repository.mutate(ctx, func(connection *sql.Conn) error {
		current, err := readSQLitePromptApplicationSession(connection.QueryRowContext(applicationInteractionRequestContext(ctx), `SELECT session_state,record_version,updated_at_unix_nano,authority_digest,sanitized_session_payload
FROM prompt_application_sessions WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND session_id=?`,
			ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, updated.SessionID), ctx)
		if err != nil {
			return err
		}
		existing, err := readSQLiteApplicationInteractionTurn(connection.QueryRowContext(applicationInteractionRequestContext(ctx), `SELECT turn_sequence,client_turn_key,turn_status,started_at_unix_nano,completed_at_unix_nano,sanitized_turn_payload
FROM prompt_application_session_turns WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND session_id=? AND client_turn_key=?`,
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
		result, err := connection.ExecContext(applicationInteractionRequestContext(ctx), `UPDATE prompt_application_sessions SET
record_version=?,updated_at_unix_nano=?,sanitized_session_payload=? WHERE tenant_ref=? AND workspace_id=? AND application_id=?
AND owner_subject_ref=? AND session_id=? AND record_version=? AND session_state='active'`,
			updated.RecordVersion, updatedAt.UnixNano(), string(sessionPayload), ctx.TenantRef, ctx.WorkspaceID,
			ctx.ApplicationID, ctx.OwnerSubjectRef, updated.SessionID, expectedVersion)
		if err != nil {
			return errApplicationSessionStore
		}
		affected, _ := result.RowsAffected()
		if affected != 1 {
			return errApplicationSessionVersionConflict
		}
		_, err = connection.ExecContext(applicationInteractionRequestContext(ctx), `INSERT INTO prompt_application_session_turns
(tenant_ref,workspace_id,application_id,owner_subject_ref,session_id,turn_id,turn_sequence,client_turn_key,turn_status,started_at_unix_nano,completed_at_unix_nano,sanitized_turn_payload)
VALUES (?,?,?,?,?,?,?,?,?,?,NULL,?)`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef,
			turn.SessionID, turn.TurnID, turn.Sequence, turn.ClientTurnKey, turn.Status, startedAt.UnixNano(), string(turnPayload))
		if err != nil {
			return errApplicationSessionStore
		}
		outputSession, outputTurn = cloneApplicationInteractionSession(updated), cloneApplicationInteractionTurn(turn)
		return nil
	})
	return outputSession, outputTurn, replay, err
}

func (repository *sqlitePromptApplicationSessionRepository) ReadTurn(ctx ApplicationInteractionContext, sessionID, turnID string) (ApplicationInteractionTurn, error) {
	if repository == nil || repository.database == nil {
		return ApplicationInteractionTurn{}, errApplicationSessionStore
	}
	return readSQLiteApplicationInteractionTurn(repository.database.QueryRowContext(applicationInteractionRequestContext(ctx), `SELECT turn_sequence,client_turn_key,turn_status,started_at_unix_nano,completed_at_unix_nano,sanitized_turn_payload
FROM prompt_application_session_turns WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND session_id=? AND turn_id=?`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, sessionID, turnID), ctx)
}

func (repository *sqlitePromptApplicationSessionRepository) ReadTurnByClientKey(ctx ApplicationInteractionContext, sessionID, clientKey string) (ApplicationInteractionTurn, error) {
	if repository == nil || repository.database == nil {
		return ApplicationInteractionTurn{}, errApplicationSessionStore
	}
	return readSQLiteApplicationInteractionTurn(repository.database.QueryRowContext(applicationInteractionRequestContext(ctx), `SELECT turn_sequence,client_turn_key,turn_status,started_at_unix_nano,completed_at_unix_nano,sanitized_turn_payload
FROM prompt_application_session_turns WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND session_id=? AND client_turn_key=?`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, sessionID, clientKey), ctx)
}

func (repository *sqlitePromptApplicationSessionRepository) CompleteTurn(ctx ApplicationInteractionContext, terminal ApplicationInteractionTurn) (ApplicationInteractionSession, ApplicationInteractionTurn, bool, error) {
	var outputSession ApplicationInteractionSession
	var outputTurn ApplicationInteractionTurn
	var replay bool
	err := repository.mutate(ctx, func(connection *sql.Conn) error {
		current, err := readSQLiteApplicationInteractionTurn(connection.QueryRowContext(applicationInteractionRequestContext(ctx), `SELECT turn_sequence,client_turn_key,turn_status,started_at_unix_nano,completed_at_unix_nano,sanitized_turn_payload
FROM prompt_application_session_turns WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND session_id=? AND turn_id=?`,
			ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, terminal.SessionID, terminal.TurnID), ctx)
		if err != nil {
			return err
		}
		session, err := readSQLitePromptApplicationSession(connection.QueryRowContext(applicationInteractionRequestContext(ctx), `SELECT session_state,record_version,updated_at_unix_nano,authority_digest,sanitized_session_payload
FROM prompt_application_sessions WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND session_id=?`,
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
		result, err := connection.ExecContext(applicationInteractionRequestContext(ctx), `UPDATE prompt_application_session_turns SET
turn_status=?,completed_at_unix_nano=?,sanitized_turn_payload=? WHERE tenant_ref=? AND workspace_id=? AND application_id=?
AND owner_subject_ref=? AND session_id=? AND turn_id=? AND turn_status='running'`,
			terminal.Status, completedAt.UnixNano(), string(payload), ctx.TenantRef, ctx.WorkspaceID,
			ctx.ApplicationID, ctx.OwnerSubjectRef, terminal.SessionID, terminal.TurnID)
		if err != nil {
			return errApplicationSessionStore
		}
		affected, _ := result.RowsAffected()
		if affected != 1 {
			return errApplicationSessionIdempotency
		}
		outputSession, outputTurn = session, cloneApplicationInteractionTurn(terminal)
		return nil
	})
	return outputSession, outputTurn, replay, err
}

func (repository *sqlitePromptApplicationSessionRepository) ListTurns(ctx ApplicationInteractionContext, sessionID string) ([]ApplicationInteractionTurn, error) {
	if _, err := repository.Read(ctx, sessionID); err != nil {
		return nil, err
	}
	rows, err := repository.database.QueryContext(applicationInteractionRequestContext(ctx), `SELECT turn_sequence,client_turn_key,turn_status,started_at_unix_nano,completed_at_unix_nano,sanitized_turn_payload
FROM prompt_application_session_turns WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND session_id=? ORDER BY turn_sequence`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, sessionID)
	if err != nil {
		return nil, errApplicationSessionStore
	}
	defer rows.Close()
	result := make([]ApplicationInteractionTurn, 0)
	for rows.Next() {
		turn, scanErr := readSQLiteApplicationInteractionTurn(rows, ctx)
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

func (repository *sqlitePromptApplicationSessionRepository) mutate(ctx ApplicationInteractionContext, operation func(*sql.Conn) error) error {
	if repository == nil || repository.database == nil {
		return errApplicationSessionStore
	}
	requestContext := applicationInteractionRequestContext(ctx)
	connection, err := repository.database.Conn(requestContext)
	if err != nil {
		return errApplicationSessionStore
	}
	defer connection.Close()
	if _, err = connection.ExecContext(requestContext, "BEGIN IMMEDIATE"); err != nil {
		return errApplicationSessionStore
	}
	committed := false
	defer func() {
		if !committed {
			_, _ = connection.ExecContext(context.Background(), "ROLLBACK")
		}
	}()
	if err = operation(connection); err != nil {
		return err
	}
	if _, err = connection.ExecContext(requestContext, "COMMIT"); err != nil {
		return errApplicationSessionStore
	}
	committed = true
	return nil
}

func readSQLitePromptApplicationSession(row sqliteApplicationInteractionSessionRow, ctx ApplicationInteractionContext) (ApplicationInteractionSession, error) {
	var state, authorityDigest, payload string
	var version int
	var updatedAt int64
	if err := row.Scan(&state, &version, &updatedAt, &authorityDigest, &payload); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ApplicationInteractionSession{}, errApplicationSessionNotFound
		}
		return ApplicationInteractionSession{}, errApplicationSessionStore
	}
	session, err := decodeApplicationInteractionSession(ctx, []byte(payload))
	if err != nil || session.State != state || session.RecordVersion != version || session.Authority.AuthorityDigest != authorityDigest {
		return ApplicationInteractionSession{}, errApplicationSessionContract
	}
	timestamp, err := applicationInteractionTimestamp(session.UpdatedAt)
	if err != nil || timestamp.UnixNano() != updatedAt {
		return ApplicationInteractionSession{}, errApplicationSessionContract
	}
	return session, nil
}

var _ applicationInteractionSessionRepository = (*sqlitePromptApplicationSessionRepository)(nil)
