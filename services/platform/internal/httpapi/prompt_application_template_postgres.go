package httpapi

import (
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresPromptApplicationTemplateRepository struct {
	pool *pgxpool.Pool
}

func newPostgresPromptApplicationTemplateRepository(pool *pgxpool.Pool) *postgresPromptApplicationTemplateRepository {
	return &postgresPromptApplicationTemplateRepository{pool: pool}
}

func (repository *postgresPromptApplicationTemplateRepository) SaveDraft(ctx PromptApplicationTemplateContext, draft PromptApplicationTemplateDraft, expectedVersion int) (PromptApplicationTemplateDraft, error) {
	if repository == nil || repository.pool == nil {
		return PromptApplicationTemplateDraft{}, errPromptApplicationTemplateStore
	}
	databaseContext := promptApplicationTemplateDatabaseContext(ctx)
	transaction, err := repository.pool.Begin(databaseContext)
	if err != nil {
		return PromptApplicationTemplateDraft{}, errPromptApplicationTemplateStore
	}
	defer func() { _ = transaction.Rollback(databaseContext) }()
	if expectedVersion > 0 {
		current, readErr := scanStoredPromptApplicationTemplateDraft(ctx, transaction.QueryRow(databaseContext, `SELECT sanitized_draft_payload
			FROM prompt_application_template_drafts WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 AND template_id=$5 FOR UPDATE`,
			ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, draft.TemplateID))
		if errors.Is(readErr, pgx.ErrNoRows) {
			return PromptApplicationTemplateDraft{}, promptApplicationTemplateVersionConflictError{CurrentDraftVersion: 0}
		}
		if readErr != nil {
			return PromptApplicationTemplateDraft{}, promptApplicationTemplateStoredError(readErr)
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
	var row promptApplicationTemplateRow
	if expectedVersion == 0 {
		row = transaction.QueryRow(databaseContext, `INSERT INTO prompt_application_template_drafts
			(tenant_ref,workspace_id,application_id,owner_subject_ref,template_id,draft_version,template_digest,updated_at,sanitized_draft_payload)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) ON CONFLICT DO NOTHING RETURNING sanitized_draft_payload`,
			ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, draft.TemplateID, draft.DraftVersion, draft.TemplateDigest, draft.UpdatedAt, payload)
	} else {
		row = transaction.QueryRow(databaseContext, `UPDATE prompt_application_template_drafts SET
			draft_version=$1,template_digest=$2,updated_at=$3,sanitized_draft_payload=$4
			WHERE tenant_ref=$5 AND workspace_id=$6 AND application_id=$7 AND owner_subject_ref=$8 AND template_id=$9 AND draft_version=$10
			RETURNING sanitized_draft_payload`, draft.DraftVersion, draft.TemplateDigest, draft.UpdatedAt, payload,
			ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, draft.TemplateID, expectedVersion)
	}
	saved, scanErr := scanStoredPromptApplicationTemplateDraft(ctx, row)
	if scanErr == nil {
		if transaction.Commit(databaseContext) != nil {
			return PromptApplicationTemplateDraft{}, errPromptApplicationTemplateStore
		}
		return saved, nil
	}
	if !errors.Is(scanErr, pgx.ErrNoRows) {
		return PromptApplicationTemplateDraft{}, promptApplicationTemplateStoredError(scanErr)
	}
	var currentVersion int
	queryErr := transaction.QueryRow(databaseContext, `SELECT draft_version FROM prompt_application_template_drafts
		WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 AND template_id=$5`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, draft.TemplateID).Scan(&currentVersion)
	if errors.Is(queryErr, pgx.ErrNoRows) {
		if expectedVersion > 0 {
			return PromptApplicationTemplateDraft{}, promptApplicationTemplateVersionConflictError{CurrentDraftVersion: 0}
		}
		return PromptApplicationTemplateDraft{}, errPromptApplicationTemplateNotFound
	}
	if queryErr != nil {
		return PromptApplicationTemplateDraft{}, errPromptApplicationTemplateStore
	}
	return PromptApplicationTemplateDraft{}, promptApplicationTemplateVersionConflictError{CurrentDraftVersion: currentVersion}
}

func (repository *postgresPromptApplicationTemplateRepository) ReadDraft(ctx PromptApplicationTemplateContext, templateID string) (PromptApplicationTemplateDraft, error) {
	if repository == nil || repository.pool == nil {
		return PromptApplicationTemplateDraft{}, errPromptApplicationTemplateStore
	}
	draft, err := scanStoredPromptApplicationTemplateDraft(ctx, repository.pool.QueryRow(promptApplicationTemplateDatabaseContext(ctx), `SELECT sanitized_draft_payload
		FROM prompt_application_template_drafts WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 AND template_id=$5`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, templateID))
	if errors.Is(err, pgx.ErrNoRows) {
		return PromptApplicationTemplateDraft{}, errPromptApplicationTemplateNotFound
	}
	if err != nil {
		return PromptApplicationTemplateDraft{}, promptApplicationTemplateStoredError(err)
	}
	return draft, nil
}

func (repository *postgresPromptApplicationTemplateRepository) ListDrafts(ctx PromptApplicationTemplateContext) ([]PromptApplicationTemplateDraftSummary, error) {
	if repository == nil || repository.pool == nil {
		return nil, errPromptApplicationTemplateStore
	}
	rows, err := repository.pool.Query(promptApplicationTemplateDatabaseContext(ctx), `SELECT sanitized_draft_payload FROM prompt_application_template_drafts
		WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 ORDER BY updated_at DESC,template_id ASC LIMIT 200`,
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

func (repository *postgresPromptApplicationTemplateRepository) CreateVersion(ctx PromptApplicationTemplateContext, version PromptApplicationTemplateVersion) (PromptApplicationTemplateVersion, error) {
	if repository == nil || repository.pool == nil {
		return PromptApplicationTemplateVersion{}, errPromptApplicationTemplateStore
	}
	databaseContext := promptApplicationTemplateDatabaseContext(ctx)
	transaction, err := repository.pool.Begin(databaseContext)
	if err != nil {
		return PromptApplicationTemplateVersion{}, errPromptApplicationTemplateStore
	}
	defer func() { _ = transaction.Rollback(databaseContext) }()
	draft, err := scanStoredPromptApplicationTemplateDraft(ctx, transaction.QueryRow(databaseContext, `SELECT sanitized_draft_payload
		FROM prompt_application_template_drafts WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 AND template_id=$5 FOR UPDATE`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, version.TemplateID))
	if errors.Is(err, pgx.ErrNoRows) {
		return PromptApplicationTemplateVersion{}, errPromptApplicationTemplateNotFound
	}
	if err != nil {
		return PromptApplicationTemplateVersion{}, promptApplicationTemplateStoredError(err)
	}
	if draft.DraftVersion != version.SourceDraftVersion || draft.TemplateDigest != version.TemplateDigest {
		return PromptApplicationTemplateVersion{}, promptApplicationTemplateVersionConflictError{CurrentDraftVersion: draft.DraftVersion}
	}
	var nextVersion int
	if err := transaction.QueryRow(databaseContext, `SELECT COALESCE(max(template_version),0)+1 FROM prompt_application_template_versions
		WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 AND template_id=$5`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, version.TemplateID).Scan(&nextVersion); err != nil {
		return PromptApplicationTemplateVersion{}, errPromptApplicationTemplateStore
	}
	var sourceExists bool
	if err := transaction.QueryRow(databaseContext, `SELECT EXISTS(SELECT 1 FROM prompt_application_template_versions
		WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 AND template_id=$5 AND source_draft_version=$6)`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, version.TemplateID, version.SourceDraftVersion).Scan(&sourceExists); err != nil {
		return PromptApplicationTemplateVersion{}, errPromptApplicationTemplateStore
	}
	if sourceExists {
		return PromptApplicationTemplateVersion{}, errPromptApplicationTemplateImmutable
	}
	version.TemplateVersion = nextVersion
	if validateStoredPromptApplicationTemplateVersion(ctx, version) != nil {
		return PromptApplicationTemplateVersion{}, errPromptApplicationTemplateContract
	}
	payload, err := json.Marshal(version)
	if err != nil {
		return PromptApplicationTemplateVersion{}, errPromptApplicationTemplateStore
	}
	created, err := scanStoredPromptApplicationTemplateVersion(ctx, transaction.QueryRow(databaseContext, `INSERT INTO prompt_application_template_versions
		(tenant_ref,workspace_id,application_id,owner_subject_ref,template_id,template_version,source_draft_version,template_digest,created_at,immutable_version_payload)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING immutable_version_payload`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef,
		version.TemplateID, version.TemplateVersion, version.SourceDraftVersion, version.TemplateDigest, version.CreatedAt, payload))
	if err != nil {
		return PromptApplicationTemplateVersion{}, promptApplicationTemplateStoredError(err)
	}
	if transaction.Commit(databaseContext) != nil {
		return PromptApplicationTemplateVersion{}, errPromptApplicationTemplateStore
	}
	return created, nil
}

func (repository *postgresPromptApplicationTemplateRepository) ReadVersion(ctx PromptApplicationTemplateContext, templateID string, templateVersion int) (PromptApplicationTemplateVersion, error) {
	if repository == nil || repository.pool == nil {
		return PromptApplicationTemplateVersion{}, errPromptApplicationTemplateStore
	}
	version, err := scanStoredPromptApplicationTemplateVersion(ctx, repository.pool.QueryRow(promptApplicationTemplateDatabaseContext(ctx), `SELECT immutable_version_payload
		FROM prompt_application_template_versions WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 AND template_id=$5 AND template_version=$6`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, templateID, templateVersion))
	if errors.Is(err, pgx.ErrNoRows) {
		return PromptApplicationTemplateVersion{}, errPromptApplicationTemplateVersionNotFound
	}
	if err != nil {
		return PromptApplicationTemplateVersion{}, promptApplicationTemplateStoredError(err)
	}
	return version, nil
}

func (repository *postgresPromptApplicationTemplateRepository) ListVersions(ctx PromptApplicationTemplateContext, templateID string) ([]PromptApplicationTemplateVersionSummary, error) {
	if repository == nil || repository.pool == nil {
		return nil, errPromptApplicationTemplateStore
	}
	databaseContext := promptApplicationTemplateDatabaseContext(ctx)
	var draftExists bool
	if err := repository.pool.QueryRow(databaseContext, `SELECT EXISTS(SELECT 1 FROM prompt_application_template_drafts
		WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 AND template_id=$5)`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, templateID).Scan(&draftExists); err != nil {
		return nil, errPromptApplicationTemplateStore
	}
	if !draftExists {
		return nil, errPromptApplicationTemplateNotFound
	}
	rows, err := repository.pool.Query(databaseContext, `SELECT immutable_version_payload FROM prompt_application_template_versions
		WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 AND template_id=$5 ORDER BY template_version DESC`,
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
