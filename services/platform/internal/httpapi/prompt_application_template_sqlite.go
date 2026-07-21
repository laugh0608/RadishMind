package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
)

type sqlitePromptApplicationTemplateRepository struct {
	database *sql.DB
}

type promptApplicationTemplateRow interface {
	Scan(...any) error
}

func newSQLitePromptApplicationTemplateRepository(database *sql.DB) *sqlitePromptApplicationTemplateRepository {
	return &sqlitePromptApplicationTemplateRepository{database: database}
}

func (repository *sqlitePromptApplicationTemplateRepository) SaveDraft(ctx PromptApplicationTemplateContext, draft PromptApplicationTemplateDraft, expectedVersion int) (PromptApplicationTemplateDraft, error) {
	if repository == nil || repository.database == nil {
		return PromptApplicationTemplateDraft{}, errPromptApplicationTemplateStore
	}
	databaseContext := promptApplicationTemplateDatabaseContext(ctx)
	connection, err := beginSQLitePromptApplicationTemplateWrite(databaseContext, repository.database)
	if err != nil {
		return PromptApplicationTemplateDraft{}, errPromptApplicationTemplateStore
	}
	defer connection.Close()
	defer func() { _, _ = connection.ExecContext(context.Background(), "ROLLBACK") }()
	if expectedVersion > 0 {
		current, readErr := scanStoredPromptApplicationTemplateDraft(ctx, connection.QueryRowContext(databaseContext, `SELECT sanitized_draft_payload
			FROM prompt_application_template_drafts WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND template_id=?`,
			ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, draft.TemplateID))
		if errors.Is(readErr, sql.ErrNoRows) {
			return PromptApplicationTemplateDraft{}, promptApplicationTemplateVersionConflictError{CurrentDraftVersion: 0}
		}
		if readErr != nil {
			return PromptApplicationTemplateDraft{}, errPromptApplicationTemplateStore
		}
		draft.CreatedAt, draft.CreatedByActorRef = current.CreatedAt, current.CreatedByActorRef
	}
	if validateStoredPromptApplicationTemplateDraft(ctx, draft) != nil {
		return PromptApplicationTemplateDraft{}, errPromptApplicationTemplateContract
	}
	payload, err := json.Marshal(draft)
	if err != nil {
		return PromptApplicationTemplateDraft{}, errPromptApplicationTemplateStore
	}
	updatedAt := parsePromptApplicationTemplateTimestamp(draft.UpdatedAt)
	if updatedAt == nil {
		return PromptApplicationTemplateDraft{}, errPromptApplicationTemplateContract
	}
	var row promptApplicationTemplateRow
	if expectedVersion == 0 {
		row = connection.QueryRowContext(databaseContext, `INSERT INTO prompt_application_template_drafts
			(tenant_ref,workspace_id,application_id,owner_subject_ref,template_id,draft_version,template_digest,updated_at_unix_nano,sanitized_draft_payload)
			VALUES (?,?,?,?,?,?,?,?,?) ON CONFLICT DO NOTHING RETURNING sanitized_draft_payload`,
			ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, draft.TemplateID, draft.DraftVersion, draft.TemplateDigest, updatedAt.UnixNano(), string(payload))
	} else {
		row = connection.QueryRowContext(databaseContext, `UPDATE prompt_application_template_drafts SET
			draft_version=?,template_digest=?,updated_at_unix_nano=?,sanitized_draft_payload=?
			WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND template_id=? AND draft_version=?
			RETURNING sanitized_draft_payload`, draft.DraftVersion, draft.TemplateDigest, updatedAt.UnixNano(), string(payload),
			ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, draft.TemplateID, expectedVersion)
	}
	saved, scanErr := scanStoredPromptApplicationTemplateDraft(ctx, row)
	if scanErr == nil {
		if _, err := connection.ExecContext(databaseContext, "COMMIT"); err != nil {
			return PromptApplicationTemplateDraft{}, errPromptApplicationTemplateStore
		}
		return saved, nil
	}
	if !errors.Is(scanErr, sql.ErrNoRows) {
		return PromptApplicationTemplateDraft{}, errPromptApplicationTemplateStore
	}
	currentVersion, found, queryErr := sqlitePromptApplicationTemplateCurrentDraftVersion(databaseContext, connection, ctx, draft.TemplateID)
	if queryErr != nil {
		return PromptApplicationTemplateDraft{}, errPromptApplicationTemplateStore
	}
	if !found && expectedVersion > 0 {
		return PromptApplicationTemplateDraft{}, promptApplicationTemplateVersionConflictError{CurrentDraftVersion: 0}
	}
	if !found {
		return PromptApplicationTemplateDraft{}, errPromptApplicationTemplateNotFound
	}
	return PromptApplicationTemplateDraft{}, promptApplicationTemplateVersionConflictError{CurrentDraftVersion: currentVersion}
}

func (repository *sqlitePromptApplicationTemplateRepository) ReadDraft(ctx PromptApplicationTemplateContext, templateID string) (PromptApplicationTemplateDraft, error) {
	if repository == nil || repository.database == nil {
		return PromptApplicationTemplateDraft{}, errPromptApplicationTemplateStore
	}
	draft, err := scanStoredPromptApplicationTemplateDraft(ctx, repository.database.QueryRowContext(promptApplicationTemplateDatabaseContext(ctx), `SELECT sanitized_draft_payload
		FROM prompt_application_template_drafts WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND template_id=?`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, templateID))
	if errors.Is(err, sql.ErrNoRows) {
		return PromptApplicationTemplateDraft{}, errPromptApplicationTemplateNotFound
	}
	if err != nil {
		return PromptApplicationTemplateDraft{}, promptApplicationTemplateStoredError(err)
	}
	return draft, nil
}

func (repository *sqlitePromptApplicationTemplateRepository) ListDrafts(ctx PromptApplicationTemplateContext) ([]PromptApplicationTemplateDraftSummary, error) {
	if repository == nil || repository.database == nil {
		return nil, errPromptApplicationTemplateStore
	}
	rows, err := repository.database.QueryContext(promptApplicationTemplateDatabaseContext(ctx), `SELECT sanitized_draft_payload FROM prompt_application_template_drafts
		WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? ORDER BY updated_at_unix_nano DESC,template_id ASC LIMIT 200`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef)
	if err != nil {
		return nil, errPromptApplicationTemplateStore
	}
	defer rows.Close()
	summaries := make([]PromptApplicationTemplateDraftSummary, 0)
	for rows.Next() {
		draft, scanErr := scanStoredPromptApplicationTemplateDraft(ctx, rows)
		if scanErr != nil {
			return nil, promptApplicationTemplateStoredError(scanErr)
		}
		summaries = append(summaries, promptApplicationTemplateDraftSummary(draft))
	}
	if rows.Err() != nil {
		return nil, errPromptApplicationTemplateStore
	}
	return summaries, nil
}

func (repository *sqlitePromptApplicationTemplateRepository) CreateVersion(ctx PromptApplicationTemplateContext, version PromptApplicationTemplateVersion) (PromptApplicationTemplateVersion, error) {
	if repository == nil || repository.database == nil {
		return PromptApplicationTemplateVersion{}, errPromptApplicationTemplateStore
	}
	databaseContext := promptApplicationTemplateDatabaseContext(ctx)
	connection, err := beginSQLitePromptApplicationTemplateWrite(databaseContext, repository.database)
	if err != nil {
		return PromptApplicationTemplateVersion{}, errPromptApplicationTemplateStore
	}
	defer connection.Close()
	defer func() { _, _ = connection.ExecContext(context.Background(), "ROLLBACK") }()
	draft, err := scanStoredPromptApplicationTemplateDraft(ctx, connection.QueryRowContext(databaseContext, `SELECT sanitized_draft_payload
		FROM prompt_application_template_drafts WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND template_id=?`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, version.TemplateID))
	if errors.Is(err, sql.ErrNoRows) {
		return PromptApplicationTemplateVersion{}, errPromptApplicationTemplateNotFound
	}
	if err != nil {
		return PromptApplicationTemplateVersion{}, promptApplicationTemplateStoredError(err)
	}
	if draft.DraftVersion != version.SourceDraftVersion || draft.TemplateDigest != version.TemplateDigest {
		return PromptApplicationTemplateVersion{}, promptApplicationTemplateVersionConflictError{CurrentDraftVersion: draft.DraftVersion}
	}
	var existingCount int
	if err := connection.QueryRowContext(databaseContext, `SELECT count(*) FROM prompt_application_template_versions
		WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND template_id=? AND source_draft_version=?`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, version.TemplateID, version.SourceDraftVersion).Scan(&existingCount); err != nil {
		return PromptApplicationTemplateVersion{}, errPromptApplicationTemplateStore
	}
	if existingCount != 0 {
		return PromptApplicationTemplateVersion{}, errPromptApplicationTemplateImmutable
	}
	var nextVersion int
	if err := connection.QueryRowContext(databaseContext, `SELECT COALESCE(max(template_version),0)+1 FROM prompt_application_template_versions
		WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND template_id=?`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, version.TemplateID).Scan(&nextVersion); err != nil {
		return PromptApplicationTemplateVersion{}, errPromptApplicationTemplateStore
	}
	version.TemplateVersion = nextVersion
	if validateStoredPromptApplicationTemplateVersion(ctx, version) != nil {
		return PromptApplicationTemplateVersion{}, errPromptApplicationTemplateContract
	}
	payload, err := json.Marshal(version)
	createdAt := parsePromptApplicationTemplateTimestamp(version.CreatedAt)
	if err != nil || createdAt == nil {
		return PromptApplicationTemplateVersion{}, errPromptApplicationTemplateContract
	}
	created, err := scanStoredPromptApplicationTemplateVersion(ctx, connection.QueryRowContext(databaseContext, `INSERT INTO prompt_application_template_versions
		(tenant_ref,workspace_id,application_id,owner_subject_ref,template_id,template_version,source_draft_version,template_digest,created_at_unix_nano,immutable_version_payload)
		VALUES (?,?,?,?,?,?,?,?,?,?) RETURNING immutable_version_payload`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef,
		version.TemplateID, version.TemplateVersion, version.SourceDraftVersion, version.TemplateDigest, createdAt.UnixNano(), string(payload)))
	if err != nil {
		return PromptApplicationTemplateVersion{}, promptApplicationTemplateStoredError(err)
	}
	if _, err := connection.ExecContext(databaseContext, "COMMIT"); err != nil {
		return PromptApplicationTemplateVersion{}, errPromptApplicationTemplateStore
	}
	return created, nil
}

func (repository *sqlitePromptApplicationTemplateRepository) ReadVersion(ctx PromptApplicationTemplateContext, templateID string, templateVersion int) (PromptApplicationTemplateVersion, error) {
	if repository == nil || repository.database == nil {
		return PromptApplicationTemplateVersion{}, errPromptApplicationTemplateStore
	}
	version, err := scanStoredPromptApplicationTemplateVersion(ctx, repository.database.QueryRowContext(promptApplicationTemplateDatabaseContext(ctx), `SELECT immutable_version_payload
		FROM prompt_application_template_versions WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND template_id=? AND template_version=?`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, templateID, templateVersion))
	if errors.Is(err, sql.ErrNoRows) {
		return PromptApplicationTemplateVersion{}, errPromptApplicationTemplateVersionNotFound
	}
	if err != nil {
		return PromptApplicationTemplateVersion{}, promptApplicationTemplateStoredError(err)
	}
	return version, nil
}

func (repository *sqlitePromptApplicationTemplateRepository) ListVersions(ctx PromptApplicationTemplateContext, templateID string) ([]PromptApplicationTemplateVersionSummary, error) {
	if repository == nil || repository.database == nil {
		return nil, errPromptApplicationTemplateStore
	}
	databaseContext := promptApplicationTemplateDatabaseContext(ctx)
	var draftCount int
	if err := repository.database.QueryRowContext(databaseContext, `SELECT count(*) FROM prompt_application_template_drafts
		WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND template_id=?`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, templateID).Scan(&draftCount); err != nil {
		return nil, errPromptApplicationTemplateStore
	}
	if draftCount == 0 {
		return nil, errPromptApplicationTemplateNotFound
	}
	rows, err := repository.database.QueryContext(databaseContext, `SELECT immutable_version_payload FROM prompt_application_template_versions
		WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND template_id=? ORDER BY template_version DESC`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, templateID)
	if err != nil {
		return nil, errPromptApplicationTemplateStore
	}
	defer rows.Close()
	summaries := make([]PromptApplicationTemplateVersionSummary, 0)
	for rows.Next() {
		version, scanErr := scanStoredPromptApplicationTemplateVersion(ctx, rows)
		if scanErr != nil {
			return nil, promptApplicationTemplateStoredError(scanErr)
		}
		summaries = append(summaries, promptApplicationTemplateVersionSummary(version))
	}
	if rows.Err() != nil {
		return nil, errPromptApplicationTemplateStore
	}
	return summaries, nil
}

func scanStoredPromptApplicationTemplateDraft(ctx PromptApplicationTemplateContext, row promptApplicationTemplateRow) (PromptApplicationTemplateDraft, error) {
	var payload []byte
	if err := row.Scan(&payload); err != nil {
		return PromptApplicationTemplateDraft{}, err
	}
	var draft PromptApplicationTemplateDraft
	if json.Unmarshal(payload, &draft) != nil || validatePromptApplicationTemplateContractJSON(promptApplicationTemplateDraftSchemaVersion, payload) != nil || validateStoredPromptApplicationTemplateDraft(ctx, draft) != nil {
		return PromptApplicationTemplateDraft{}, errPromptApplicationTemplateContract
	}
	return clonePromptApplicationTemplateDraft(draft), nil
}

func scanStoredPromptApplicationTemplateVersion(ctx PromptApplicationTemplateContext, row promptApplicationTemplateRow) (PromptApplicationTemplateVersion, error) {
	var payload []byte
	if err := row.Scan(&payload); err != nil {
		return PromptApplicationTemplateVersion{}, err
	}
	var version PromptApplicationTemplateVersion
	if json.Unmarshal(payload, &version) != nil || validatePromptApplicationTemplateContractJSON(promptApplicationTemplateVersionSchemaVersion, payload) != nil || validateStoredPromptApplicationTemplateVersion(ctx, version) != nil {
		return PromptApplicationTemplateVersion{}, errPromptApplicationTemplateContract
	}
	return clonePromptApplicationTemplateVersion(version), nil
}

func promptApplicationTemplateStoredError(err error) error {
	if errors.Is(err, errPromptApplicationTemplateContract) {
		return errPromptApplicationTemplateContract
	}
	return errPromptApplicationTemplateStore
}

type sqlitePromptApplicationTemplateQueryRower interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func sqlitePromptApplicationTemplateCurrentDraftVersion(ctx context.Context, query sqlitePromptApplicationTemplateQueryRower, scope PromptApplicationTemplateContext, templateID string) (int, bool, error) {
	var version int
	err := query.QueryRowContext(ctx, `SELECT draft_version FROM prompt_application_template_drafts
		WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND template_id=?`,
		scope.TenantRef, scope.WorkspaceID, scope.ApplicationID, scope.OwnerSubjectRef, templateID).Scan(&version)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	return version, err == nil, err
}

func beginSQLitePromptApplicationTemplateWrite(ctx context.Context, database *sql.DB) (*sql.Conn, error) {
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

func promptApplicationTemplateDatabaseContext(ctx PromptApplicationTemplateContext) context.Context {
	if ctx.RequestContext != nil {
		return ctx.RequestContext
	}
	return context.Background()
}
