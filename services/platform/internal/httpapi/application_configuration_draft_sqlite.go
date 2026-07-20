package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
)

type sqliteApplicationConfigurationDraftRepository struct {
	database *sql.DB
}

func newSQLiteApplicationConfigurationDraftRepository(database *sql.DB) *sqliteApplicationConfigurationDraftRepository {
	return &sqliteApplicationConfigurationDraftRepository{database: database}
}

func (repository *sqliteApplicationConfigurationDraftRepository) Save(
	requestContext ApplicationConfigurationDraftContext,
	draft ApplicationConfigurationDraft,
	expectedVersion int,
) (ApplicationConfigurationDraft, error) {
	if repository == nil || repository.database == nil {
		return ApplicationConfigurationDraft{}, errApplicationDraftStoreUnavailable
	}
	if expectedVersion > 0 {
		current, err := repository.Read(requestContext, draft.DraftID)
		if err != nil {
			return ApplicationConfigurationDraft{}, err
		}
		draft.CreatedAt = current.CreatedAt
		draft.CreatedByActorRef = current.CreatedByActorRef
	}
	payload, err := json.Marshal(draft)
	if err != nil {
		return ApplicationConfigurationDraft{}, errApplicationDraftStoreUnavailable
	}
	ctx := applicationDraftDatabaseContext(requestContext)
	var row applicationConfigurationDraftRow
	if expectedVersion == 0 {
		row = repository.database.QueryRowContext(ctx, `INSERT INTO application_configuration_drafts
            (tenant_ref, workspace_id, application_id, owner_subject_ref, draft_id, draft_version,
             schema_version, sanitized_draft_payload, created_at, updated_at, created_by_actor_ref,
             updated_by_actor_ref, request_id, audit_ref)
            VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)
            ON CONFLICT DO NOTHING
            RETURNING sanitized_draft_payload`,
			requestContext.TenantRef, requestContext.WorkspaceID, requestContext.ApplicationID, requestContext.OwnerSubjectRef,
			draft.DraftID, draft.DraftVersion, draft.SchemaVersion, string(payload), draft.CreatedAt, draft.UpdatedAt,
			draft.CreatedByActorRef, draft.UpdatedByActorRef, draft.RequestID, draft.AuditRef,
		)
	} else {
		row = repository.database.QueryRowContext(ctx, `UPDATE application_configuration_drafts SET
            draft_version=?, schema_version=?, sanitized_draft_payload=?, updated_at=?,
            updated_by_actor_ref=?, request_id=?, audit_ref=?
            WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=?
              AND draft_id=? AND draft_version=?
            RETURNING sanitized_draft_payload`,
			draft.DraftVersion, draft.SchemaVersion, string(payload), draft.UpdatedAt, draft.UpdatedByActorRef, draft.RequestID, draft.AuditRef,
			requestContext.TenantRef, requestContext.WorkspaceID, requestContext.ApplicationID, requestContext.OwnerSubjectRef,
			draft.DraftID, expectedVersion,
		)
	}
	saved, err := scanApplicationConfigurationDraft(row)
	if err == nil {
		return saved, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return ApplicationConfigurationDraft{}, errApplicationDraftStoreUnavailable
	}
	currentVersion, found, queryFailed := repository.currentVersion(ctx, requestContext, draft.DraftID)
	if queryFailed {
		return ApplicationConfigurationDraft{}, errApplicationDraftStoreUnavailable
	}
	if !found {
		return ApplicationConfigurationDraft{}, errApplicationDraftNotFound
	}
	return ApplicationConfigurationDraft{}, applicationDraftVersionConflictError{CurrentVersion: currentVersion}
}

func (repository *sqliteApplicationConfigurationDraftRepository) Read(
	requestContext ApplicationConfigurationDraftContext,
	draftID string,
) (ApplicationConfigurationDraft, error) {
	if repository == nil || repository.database == nil {
		return ApplicationConfigurationDraft{}, errApplicationDraftStoreUnavailable
	}
	draft, err := scanApplicationConfigurationDraft(repository.database.QueryRowContext(applicationDraftDatabaseContext(requestContext), `
        SELECT sanitized_draft_payload FROM application_configuration_drafts
        WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND draft_id=?`,
		requestContext.TenantRef, requestContext.WorkspaceID, requestContext.ApplicationID, requestContext.OwnerSubjectRef, draftID,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return ApplicationConfigurationDraft{}, errApplicationDraftNotFound
	}
	if err != nil {
		return ApplicationConfigurationDraft{}, errApplicationDraftStoreUnavailable
	}
	return draft, nil
}

func (repository *sqliteApplicationConfigurationDraftRepository) List(
	requestContext ApplicationConfigurationDraftContext,
) ([]ApplicationConfigurationDraftSummary, error) {
	if repository == nil || repository.database == nil {
		return nil, errApplicationDraftStoreUnavailable
	}
	rows, err := repository.database.QueryContext(applicationDraftDatabaseContext(requestContext), `
        SELECT sanitized_draft_payload FROM application_configuration_drafts
        WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=?
        ORDER BY updated_at DESC, draft_id ASC LIMIT 200`,
		requestContext.TenantRef, requestContext.WorkspaceID, requestContext.ApplicationID, requestContext.OwnerSubjectRef,
	)
	if err != nil {
		return nil, errApplicationDraftStoreUnavailable
	}
	defer rows.Close()
	summaries := make([]ApplicationConfigurationDraftSummary, 0)
	for rows.Next() {
		draft, scanErr := scanApplicationConfigurationDraft(rows)
		if scanErr != nil {
			return nil, errApplicationDraftStoreUnavailable
		}
		summaries = append(summaries, applicationConfigurationDraftSummary(draft))
	}
	if rows.Err() != nil {
		return nil, errApplicationDraftStoreUnavailable
	}
	return summaries, nil
}

func (repository *sqliteApplicationConfigurationDraftRepository) currentVersion(
	ctx context.Context,
	requestContext ApplicationConfigurationDraftContext,
	draftID string,
) (int, bool, bool) {
	var version int
	err := repository.database.QueryRowContext(ctx, `SELECT draft_version FROM application_configuration_drafts
        WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND draft_id=?`,
		requestContext.TenantRef, requestContext.WorkspaceID, requestContext.ApplicationID, requestContext.OwnerSubjectRef, draftID,
	).Scan(&version)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, false
	}
	if err != nil {
		return 0, false, true
	}
	return version, true, false
}
