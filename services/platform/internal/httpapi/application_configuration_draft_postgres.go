package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresApplicationConfigurationDraftRepository struct {
	pool *pgxpool.Pool
}

type applicationConfigurationDraftRow interface {
	Scan(...any) error
}

func newPostgresApplicationConfigurationDraftRepository(pool *pgxpool.Pool) *postgresApplicationConfigurationDraftRepository {
	return &postgresApplicationConfigurationDraftRepository{pool: pool}
}

func (repository *postgresApplicationConfigurationDraftRepository) Save(
	requestContext ApplicationConfigurationDraftContext,
	draft ApplicationConfigurationDraft,
	expectedVersion int,
) (ApplicationConfigurationDraft, error) {
	if repository == nil || repository.pool == nil {
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
		row = repository.pool.QueryRow(ctx, `INSERT INTO application_configuration_drafts
            (tenant_ref, workspace_id, application_id, owner_subject_ref, draft_id, draft_version,
             schema_version, sanitized_draft_payload, created_at, updated_at, created_by_actor_ref,
             updated_by_actor_ref, request_id, audit_ref)
            VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
            ON CONFLICT DO NOTHING
            RETURNING sanitized_draft_payload`,
			requestContext.TenantRef, requestContext.WorkspaceID, requestContext.ApplicationID, requestContext.OwnerSubjectRef,
			draft.DraftID, draft.DraftVersion, draft.SchemaVersion, payload, draft.CreatedAt, draft.UpdatedAt,
			draft.CreatedByActorRef, draft.UpdatedByActorRef, draft.RequestID, draft.AuditRef,
		)
	} else {
		row = repository.pool.QueryRow(ctx, `UPDATE application_configuration_drafts SET
            draft_version=$1, schema_version=$2, sanitized_draft_payload=$3, updated_at=$4,
            updated_by_actor_ref=$5, request_id=$6, audit_ref=$7
            WHERE tenant_ref=$8 AND workspace_id=$9 AND application_id=$10 AND owner_subject_ref=$11
              AND draft_id=$12 AND draft_version=$13
            RETURNING sanitized_draft_payload`,
			draft.DraftVersion, draft.SchemaVersion, payload, draft.UpdatedAt, draft.UpdatedByActorRef, draft.RequestID, draft.AuditRef,
			requestContext.TenantRef, requestContext.WorkspaceID, requestContext.ApplicationID, requestContext.OwnerSubjectRef,
			draft.DraftID, expectedVersion,
		)
	}
	saved, err := scanApplicationConfigurationDraft(row)
	if err == nil {
		return saved, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
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

func (repository *postgresApplicationConfigurationDraftRepository) Read(
	requestContext ApplicationConfigurationDraftContext,
	draftID string,
) (ApplicationConfigurationDraft, error) {
	if repository == nil || repository.pool == nil {
		return ApplicationConfigurationDraft{}, errApplicationDraftStoreUnavailable
	}
	draft, err := scanApplicationConfigurationDraft(repository.pool.QueryRow(applicationDraftDatabaseContext(requestContext), `
        SELECT sanitized_draft_payload FROM application_configuration_drafts
        WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 AND draft_id=$5`,
		requestContext.TenantRef, requestContext.WorkspaceID, requestContext.ApplicationID, requestContext.OwnerSubjectRef, draftID,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return ApplicationConfigurationDraft{}, errApplicationDraftNotFound
	}
	if err != nil {
		return ApplicationConfigurationDraft{}, errApplicationDraftStoreUnavailable
	}
	return draft, nil
}

func (repository *postgresApplicationConfigurationDraftRepository) List(
	requestContext ApplicationConfigurationDraftContext,
) ([]ApplicationConfigurationDraftSummary, error) {
	if repository == nil || repository.pool == nil {
		return nil, errApplicationDraftStoreUnavailable
	}
	rows, err := repository.pool.Query(applicationDraftDatabaseContext(requestContext), `
        SELECT sanitized_draft_payload FROM application_configuration_drafts
        WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4
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

func (repository *postgresApplicationConfigurationDraftRepository) currentVersion(
	ctx context.Context,
	requestContext ApplicationConfigurationDraftContext,
	draftID string,
) (int, bool, bool) {
	var version int
	err := repository.pool.QueryRow(ctx, `SELECT draft_version FROM application_configuration_drafts
        WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 AND draft_id=$5`,
		requestContext.TenantRef, requestContext.WorkspaceID, requestContext.ApplicationID, requestContext.OwnerSubjectRef, draftID,
	).Scan(&version)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, false, false
	}
	if err != nil {
		return 0, false, true
	}
	return version, true, false
}

func scanApplicationConfigurationDraft(row applicationConfigurationDraftRow) (ApplicationConfigurationDraft, error) {
	var payload []byte
	if err := row.Scan(&payload); err != nil {
		return ApplicationConfigurationDraft{}, err
	}
	var draft ApplicationConfigurationDraft
	if err := json.Unmarshal(payload, &draft); err != nil || strings.TrimSpace(draft.DraftID) == "" || draft.DraftVersion < 1 || !applicationConfigurationDraftSchemaSupported(draft.SchemaVersion) ||
		(draft.SchemaVersion == applicationConfigurationDraftSchemaVersionV1 && (draft.WorkflowRAGBindingRef != nil || draft.PromptTemplateRef != nil)) ||
		(draft.SchemaVersion == applicationConfigurationDraftSchemaVersionV2 && draft.PromptTemplateRef != nil) ||
		(draft.SchemaVersion == applicationConfigurationDraftSchemaVersionV3 && (draft.ApplicationKind != "prompt_application" || draft.WorkflowRAGBindingRef != nil || draft.PromptTemplateRef == nil)) ||
		(draft.WorkflowRAGBindingRef != nil && !validWorkflowRAGApplicationBindingRef(*draft.WorkflowRAGBindingRef)) ||
		(draft.PromptTemplateRef != nil && !validPromptApplicationTemplateRef(*draft.PromptTemplateRef)) {
		return ApplicationConfigurationDraft{}, errors.New("stored application draft contract mismatch")
	}
	draft.ApplicationConfigurationDraftPayload = normalizeApplicationConfigurationDraftPayload(draft.ApplicationConfigurationDraftPayload)
	digest, err := applicationConfigurationCanonicalDigest(applicationPublishSnapshotFromDraft(draft))
	if err != nil || draft.DraftDigest != "" && draft.DraftDigest != digest || draft.SchemaVersion != applicationConfigurationDraftSchemaVersionV1 && draft.DraftDigest == "" {
		return ApplicationConfigurationDraft{}, errors.New("stored application draft contract mismatch")
	}
	draft.DraftDigest = digest
	return draft, nil
}

func applicationDraftDatabaseContext(requestContext ApplicationConfigurationDraftContext) context.Context {
	if requestContext.RequestContext != nil {
		return requestContext.RequestContext
	}
	return context.Background()
}
