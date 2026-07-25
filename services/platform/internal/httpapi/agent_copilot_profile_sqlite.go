package httpapi

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"time"
)

type sqliteAgentCopilotProfileRepository struct {
	database *sql.DB
}

type agentCopilotProfileRow interface {
	Scan(...any) error
}

func newSQLiteAgentCopilotProfileRepository(database *sql.DB) *sqliteAgentCopilotProfileRepository {
	return &sqliteAgentCopilotProfileRepository{database: database}
}

func (repository *sqliteAgentCopilotProfileRepository) SaveDraft(ctx AgentCopilotProfileContext, draft AgentCopilotProfileDraftV1, expectedVersion int) (AgentCopilotProfileDraftV1, error) {
	if repository == nil || repository.database == nil {
		return AgentCopilotProfileDraftV1{}, errAgentCopilotProfileStore
	}
	databaseContext := agentCopilotProfileDatabaseContext(ctx)
	connection, err := beginSQLiteAgentCopilotProfileWrite(databaseContext, repository.database)
	if err != nil {
		return AgentCopilotProfileDraftV1{}, errAgentCopilotProfileStore
	}
	defer connection.Close()
	defer func() { _, _ = connection.ExecContext(context.Background(), "ROLLBACK") }()
	if expectedVersion > 0 {
		current, readErr := scanStoredAgentCopilotProfileDraft(ctx, connection.QueryRowContext(databaseContext, `SELECT sanitized_draft_payload
			FROM agent_copilot_profile_drafts WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND profile_id=?`,
			ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, draft.ProfileID))
		if errors.Is(readErr, sql.ErrNoRows) {
			return AgentCopilotProfileDraftV1{}, agentCopilotProfileVersionConflictError{CurrentDraftVersion: 0}
		}
		if readErr != nil {
			return AgentCopilotProfileDraftV1{}, agentCopilotProfileStoredError(readErr)
		}
		draft.CreatedAt, draft.CreatedByActorRef = current.CreatedAt, current.CreatedByActorRef
	}
	if err := validateStoredAgentCopilotProfileDraft(ctx, draft); err != nil {
		return AgentCopilotProfileDraftV1{}, err
	}
	payload, err := json.Marshal(draft)
	updatedAt, timestampErr := time.Parse(time.RFC3339Nano, draft.UpdatedAt)
	if err != nil || timestampErr != nil {
		return AgentCopilotProfileDraftV1{}, errAgentCopilotProfileOwner
	}
	var row agentCopilotProfileRow
	if expectedVersion == 0 {
		row = connection.QueryRowContext(databaseContext, `INSERT INTO agent_copilot_profile_drafts
			(tenant_ref,workspace_id,application_id,owner_subject_ref,profile_id,draft_version,profile_digest,policy_digest,updated_at_unix_nano,sanitized_draft_payload)
			VALUES (?,?,?,?,?,?,?,?,?,?) ON CONFLICT DO NOTHING RETURNING sanitized_draft_payload`,
			ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, draft.ProfileID, draft.DraftVersion,
			draft.ProfileDigest, draft.PolicyDigest, updatedAt.UnixNano(), string(payload))
	} else {
		row = connection.QueryRowContext(databaseContext, `UPDATE agent_copilot_profile_drafts SET
			draft_version=?,profile_digest=?,policy_digest=?,updated_at_unix_nano=?,sanitized_draft_payload=?
			WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND profile_id=? AND draft_version=?
			RETURNING sanitized_draft_payload`, draft.DraftVersion, draft.ProfileDigest, draft.PolicyDigest, updatedAt.UnixNano(), string(payload),
			ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, draft.ProfileID, expectedVersion)
	}
	saved, scanErr := scanStoredAgentCopilotProfileDraft(ctx, row)
	if scanErr == nil {
		if _, err := connection.ExecContext(databaseContext, "COMMIT"); err != nil {
			return AgentCopilotProfileDraftV1{}, errAgentCopilotProfileStore
		}
		return saved, nil
	}
	if !errors.Is(scanErr, sql.ErrNoRows) {
		return AgentCopilotProfileDraftV1{}, agentCopilotProfileStoredError(scanErr)
	}
	currentVersion, found, queryErr := sqliteAgentCopilotProfileCurrentDraftVersion(databaseContext, connection, ctx, draft.ProfileID)
	if queryErr != nil {
		return AgentCopilotProfileDraftV1{}, errAgentCopilotProfileStore
	}
	if !found && expectedVersion > 0 {
		return AgentCopilotProfileDraftV1{}, agentCopilotProfileVersionConflictError{CurrentDraftVersion: 0}
	}
	if !found {
		return AgentCopilotProfileDraftV1{}, errAgentCopilotProfileNotFound
	}
	return AgentCopilotProfileDraftV1{}, agentCopilotProfileVersionConflictError{CurrentDraftVersion: currentVersion}
}

func (repository *sqliteAgentCopilotProfileRepository) ReadDraft(ctx AgentCopilotProfileContext, profileID string) (AgentCopilotProfileDraftV1, error) {
	if repository == nil || repository.database == nil {
		return AgentCopilotProfileDraftV1{}, errAgentCopilotProfileStore
	}
	draft, err := scanStoredAgentCopilotProfileDraft(ctx, repository.database.QueryRowContext(agentCopilotProfileDatabaseContext(ctx), `SELECT sanitized_draft_payload
		FROM agent_copilot_profile_drafts WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND profile_id=?`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, profileID))
	if errors.Is(err, sql.ErrNoRows) {
		return AgentCopilotProfileDraftV1{}, errAgentCopilotProfileNotFound
	}
	if err != nil {
		return AgentCopilotProfileDraftV1{}, agentCopilotProfileStoredError(err)
	}
	return draft, nil
}

func (repository *sqliteAgentCopilotProfileRepository) ListDrafts(ctx AgentCopilotProfileContext) ([]AgentCopilotProfileDraftSummary, error) {
	if repository == nil || repository.database == nil {
		return nil, errAgentCopilotProfileStore
	}
	rows, err := repository.database.QueryContext(agentCopilotProfileDatabaseContext(ctx), `SELECT sanitized_draft_payload FROM agent_copilot_profile_drafts
		WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? ORDER BY updated_at_unix_nano DESC,profile_id ASC LIMIT 200`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef)
	if err != nil {
		return nil, errAgentCopilotProfileStore
	}
	defer rows.Close()
	summaries := make([]AgentCopilotProfileDraftSummary, 0)
	for rows.Next() {
		draft, scanErr := scanStoredAgentCopilotProfileDraft(ctx, rows)
		if scanErr != nil {
			return nil, agentCopilotProfileStoredError(scanErr)
		}
		summaries = append(summaries, agentCopilotProfileDraftSummary(draft))
	}
	if rows.Err() != nil {
		return nil, errAgentCopilotProfileStore
	}
	return summaries, nil
}

func (repository *sqliteAgentCopilotProfileRepository) CreateVersion(ctx AgentCopilotProfileContext, version AgentCopilotProfileVersionV1) (AgentCopilotProfileVersionV1, error) {
	if repository == nil || repository.database == nil {
		return AgentCopilotProfileVersionV1{}, errAgentCopilotProfileStore
	}
	databaseContext := agentCopilotProfileDatabaseContext(ctx)
	connection, err := beginSQLiteAgentCopilotProfileWrite(databaseContext, repository.database)
	if err != nil {
		return AgentCopilotProfileVersionV1{}, errAgentCopilotProfileStore
	}
	defer connection.Close()
	defer func() { _, _ = connection.ExecContext(context.Background(), "ROLLBACK") }()
	draft, err := scanStoredAgentCopilotProfileDraft(ctx, connection.QueryRowContext(databaseContext, `SELECT sanitized_draft_payload
		FROM agent_copilot_profile_drafts WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND profile_id=?`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, version.ProfileID))
	if errors.Is(err, sql.ErrNoRows) {
		return AgentCopilotProfileVersionV1{}, errAgentCopilotProfileNotFound
	}
	if err != nil {
		return AgentCopilotProfileVersionV1{}, agentCopilotProfileStoredError(err)
	}
	if draft.DraftVersion != version.SourceDraftVersion || draft.ProfileDigest != version.ProfileDigest || draft.PolicyDigest != version.PolicyDigest {
		return AgentCopilotProfileVersionV1{}, agentCopilotProfileVersionConflictError{CurrentDraftVersion: draft.DraftVersion}
	}
	var existingCount int
	if err := connection.QueryRowContext(databaseContext, `SELECT count(*) FROM agent_copilot_profile_versions
		WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND profile_id=? AND source_draft_version=?`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, version.ProfileID, version.SourceDraftVersion).Scan(&existingCount); err != nil {
		return AgentCopilotProfileVersionV1{}, errAgentCopilotProfileStore
	}
	if existingCount != 0 {
		return AgentCopilotProfileVersionV1{}, errAgentCopilotProfileImmutable
	}
	var nextVersion int
	if err := connection.QueryRowContext(databaseContext, `SELECT COALESCE(max(profile_version),0)+1 FROM agent_copilot_profile_versions
		WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND profile_id=?`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, version.ProfileID).Scan(&nextVersion); err != nil {
		return AgentCopilotProfileVersionV1{}, errAgentCopilotProfileStore
	}
	version.ProfileVersion = nextVersion
	if err := validateStoredAgentCopilotProfileVersion(ctx, version); err != nil {
		return AgentCopilotProfileVersionV1{}, err
	}
	payload, err := json.Marshal(version)
	publishedAt, timestampErr := time.Parse(time.RFC3339Nano, version.PublishedAt)
	if err != nil || timestampErr != nil {
		return AgentCopilotProfileVersionV1{}, errAgentCopilotProfileOwner
	}
	created, err := scanStoredAgentCopilotProfileVersion(ctx, connection.QueryRowContext(databaseContext, `INSERT INTO agent_copilot_profile_versions
		(tenant_ref,workspace_id,application_id,owner_subject_ref,profile_id,profile_version,source_draft_version,profile_digest,policy_digest,published_at_unix_nano,immutable_version_payload)
		VALUES (?,?,?,?,?,?,?,?,?,?,?) RETURNING immutable_version_payload`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef,
		version.ProfileID, version.ProfileVersion, version.SourceDraftVersion, version.ProfileDigest, version.PolicyDigest, publishedAt.UnixNano(), string(payload)))
	if err != nil {
		return AgentCopilotProfileVersionV1{}, agentCopilotProfileStoredError(err)
	}
	if _, err := connection.ExecContext(databaseContext, "COMMIT"); err != nil {
		return AgentCopilotProfileVersionV1{}, errAgentCopilotProfileStore
	}
	return created, nil
}

func (repository *sqliteAgentCopilotProfileRepository) ReadVersion(ctx AgentCopilotProfileContext, profileID string, profileVersion int) (AgentCopilotProfileVersionV1, error) {
	if repository == nil || repository.database == nil {
		return AgentCopilotProfileVersionV1{}, errAgentCopilotProfileStore
	}
	version, err := scanStoredAgentCopilotProfileVersion(ctx, repository.database.QueryRowContext(agentCopilotProfileDatabaseContext(ctx), `SELECT immutable_version_payload
		FROM agent_copilot_profile_versions WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND profile_id=? AND profile_version=?`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, profileID, profileVersion))
	if errors.Is(err, sql.ErrNoRows) {
		return AgentCopilotProfileVersionV1{}, errAgentCopilotProfileNotFound
	}
	if err != nil {
		return AgentCopilotProfileVersionV1{}, agentCopilotProfileStoredError(err)
	}
	return version, nil
}

func (repository *sqliteAgentCopilotProfileRepository) ListVersions(ctx AgentCopilotProfileContext, profileID string) ([]AgentCopilotProfileVersionSummary, error) {
	if repository == nil || repository.database == nil {
		return nil, errAgentCopilotProfileStore
	}
	databaseContext := agentCopilotProfileDatabaseContext(ctx)
	var draftCount int
	if err := repository.database.QueryRowContext(databaseContext, `SELECT count(*) FROM agent_copilot_profile_drafts
		WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND profile_id=?`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, profileID).Scan(&draftCount); err != nil {
		return nil, errAgentCopilotProfileStore
	}
	if draftCount == 0 {
		return nil, errAgentCopilotProfileNotFound
	}
	rows, err := repository.database.QueryContext(databaseContext, `SELECT immutable_version_payload FROM agent_copilot_profile_versions
		WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND profile_id=? ORDER BY profile_version DESC`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, profileID)
	if err != nil {
		return nil, errAgentCopilotProfileStore
	}
	defer rows.Close()
	summaries := make([]AgentCopilotProfileVersionSummary, 0)
	for rows.Next() {
		version, scanErr := scanStoredAgentCopilotProfileVersion(ctx, rows)
		if scanErr != nil {
			return nil, agentCopilotProfileStoredError(scanErr)
		}
		summaries = append(summaries, agentCopilotProfileVersionSummary(version))
	}
	if rows.Err() != nil {
		return nil, errAgentCopilotProfileStore
	}
	return summaries, nil
}

func scanStoredAgentCopilotProfileDraft(ctx AgentCopilotProfileContext, row agentCopilotProfileRow) (AgentCopilotProfileDraftV1, error) {
	var payload []byte
	if err := row.Scan(&payload); err != nil {
		return AgentCopilotProfileDraftV1{}, err
	}
	var draft AgentCopilotProfileDraftV1
	if strictAgentCopilotProfileJSON(payload, &draft) != nil {
		return AgentCopilotProfileDraftV1{}, errAgentCopilotProfileOwner
	}
	if err := validateStoredAgentCopilotProfileDraft(ctx, draft); err != nil {
		return AgentCopilotProfileDraftV1{}, err
	}
	return cloneAgentCopilotProfileDraft(draft), nil
}

func scanStoredAgentCopilotProfileVersion(ctx AgentCopilotProfileContext, row agentCopilotProfileRow) (AgentCopilotProfileVersionV1, error) {
	var payload []byte
	if err := row.Scan(&payload); err != nil {
		return AgentCopilotProfileVersionV1{}, err
	}
	var version AgentCopilotProfileVersionV1
	if strictAgentCopilotProfileJSON(payload, &version) != nil {
		return AgentCopilotProfileVersionV1{}, errAgentCopilotProfileOwner
	}
	if err := validateStoredAgentCopilotProfileVersion(ctx, version); err != nil {
		return AgentCopilotProfileVersionV1{}, err
	}
	return cloneAgentCopilotProfileVersion(version), nil
}

func strictAgentCopilotProfileJSON(payload []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil || decoder.Decode(&struct{}{}) != io.EOF {
		return errAgentCopilotProfileOwner
	}
	return nil
}

func agentCopilotProfileStoredError(err error) error {
	if errors.Is(err, errAgentCopilotProfileDigest) {
		return errAgentCopilotProfileDigest
	}
	if errors.Is(err, errAgentCopilotProfileOwner) {
		return errAgentCopilotProfileOwner
	}
	return errAgentCopilotProfileStore
}

func sqliteAgentCopilotProfileCurrentDraftVersion(ctx context.Context, connection *sql.Conn, scope AgentCopilotProfileContext, profileID string) (int, bool, error) {
	var version int
	err := connection.QueryRowContext(ctx, `SELECT draft_version FROM agent_copilot_profile_drafts
		WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND profile_id=?`,
		scope.TenantRef, scope.WorkspaceID, scope.ApplicationID, scope.OwnerSubjectRef, profileID).Scan(&version)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	return version, err == nil, err
}

func beginSQLiteAgentCopilotProfileWrite(ctx context.Context, database *sql.DB) (*sql.Conn, error) {
	connection, err := database.Conn(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := connection.ExecContext(ctx, "BEGIN IMMEDIATE"); err != nil {
		_ = connection.Close()
		return nil, err
	}
	return connection, nil
}

func agentCopilotProfileDatabaseContext(ctx AgentCopilotProfileContext) context.Context {
	if ctx.RequestContext != nil {
		return ctx.RequestContext
	}
	return context.Background()
}
