package httpapi

import (
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresAgentCopilotProfileRepository struct {
	pool *pgxpool.Pool
}

func newPostgresAgentCopilotProfileRepository(pool *pgxpool.Pool) *postgresAgentCopilotProfileRepository {
	return &postgresAgentCopilotProfileRepository{pool: pool}
}

func (repository *postgresAgentCopilotProfileRepository) SaveDraft(ctx AgentCopilotProfileContext, draft AgentCopilotProfileDraftV1, expectedVersion int) (AgentCopilotProfileDraftV1, error) {
	if repository == nil || repository.pool == nil {
		return AgentCopilotProfileDraftV1{}, errAgentCopilotProfileStore
	}
	databaseContext := agentCopilotProfileDatabaseContext(ctx)
	transaction, err := repository.pool.Begin(databaseContext)
	if err != nil {
		return AgentCopilotProfileDraftV1{}, errAgentCopilotProfileStore
	}
	defer func() { _ = transaction.Rollback(databaseContext) }()
	if expectedVersion > 0 {
		current, readErr := scanStoredAgentCopilotProfileDraft(ctx, transaction.QueryRow(databaseContext, `SELECT sanitized_draft_payload
			FROM agent_copilot_profile_drafts WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 AND profile_id=$5 FOR UPDATE`,
			ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, draft.ProfileID))
		if errors.Is(readErr, pgx.ErrNoRows) {
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
	if err != nil {
		return AgentCopilotProfileDraftV1{}, errAgentCopilotProfileStore
	}
	var row agentCopilotProfileRow
	if expectedVersion == 0 {
		row = transaction.QueryRow(databaseContext, `INSERT INTO agent_copilot_profile_drafts
			(tenant_ref,workspace_id,application_id,owner_subject_ref,profile_id,draft_version,profile_digest,policy_digest,updated_at,sanitized_draft_payload)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) ON CONFLICT DO NOTHING RETURNING sanitized_draft_payload`,
			ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, draft.ProfileID, draft.DraftVersion,
			draft.ProfileDigest, draft.PolicyDigest, draft.UpdatedAt, payload)
	} else {
		row = transaction.QueryRow(databaseContext, `UPDATE agent_copilot_profile_drafts SET
			draft_version=$1,profile_digest=$2,policy_digest=$3,updated_at=$4,sanitized_draft_payload=$5
			WHERE tenant_ref=$6 AND workspace_id=$7 AND application_id=$8 AND owner_subject_ref=$9 AND profile_id=$10 AND draft_version=$11
			RETURNING sanitized_draft_payload`, draft.DraftVersion, draft.ProfileDigest, draft.PolicyDigest, draft.UpdatedAt, payload,
			ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, draft.ProfileID, expectedVersion)
	}
	saved, scanErr := scanStoredAgentCopilotProfileDraft(ctx, row)
	if scanErr == nil {
		if transaction.Commit(databaseContext) != nil {
			return AgentCopilotProfileDraftV1{}, errAgentCopilotProfileStore
		}
		return saved, nil
	}
	if !errors.Is(scanErr, pgx.ErrNoRows) {
		return AgentCopilotProfileDraftV1{}, agentCopilotProfileStoredError(scanErr)
	}
	var currentVersion int
	queryErr := transaction.QueryRow(databaseContext, `SELECT draft_version FROM agent_copilot_profile_drafts
		WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 AND profile_id=$5`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, draft.ProfileID).Scan(&currentVersion)
	if errors.Is(queryErr, pgx.ErrNoRows) {
		if expectedVersion > 0 {
			return AgentCopilotProfileDraftV1{}, agentCopilotProfileVersionConflictError{CurrentDraftVersion: 0}
		}
		return AgentCopilotProfileDraftV1{}, errAgentCopilotProfileNotFound
	}
	if queryErr != nil {
		return AgentCopilotProfileDraftV1{}, errAgentCopilotProfileStore
	}
	return AgentCopilotProfileDraftV1{}, agentCopilotProfileVersionConflictError{CurrentDraftVersion: currentVersion}
}

func (repository *postgresAgentCopilotProfileRepository) ReadDraft(ctx AgentCopilotProfileContext, profileID string) (AgentCopilotProfileDraftV1, error) {
	if repository == nil || repository.pool == nil {
		return AgentCopilotProfileDraftV1{}, errAgentCopilotProfileStore
	}
	draft, err := scanStoredAgentCopilotProfileDraft(ctx, repository.pool.QueryRow(agentCopilotProfileDatabaseContext(ctx), `SELECT sanitized_draft_payload
		FROM agent_copilot_profile_drafts WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 AND profile_id=$5`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, profileID))
	if errors.Is(err, pgx.ErrNoRows) {
		return AgentCopilotProfileDraftV1{}, errAgentCopilotProfileNotFound
	}
	if err != nil {
		return AgentCopilotProfileDraftV1{}, agentCopilotProfileStoredError(err)
	}
	return draft, nil
}

func (repository *postgresAgentCopilotProfileRepository) ListDrafts(ctx AgentCopilotProfileContext) ([]AgentCopilotProfileDraftSummary, error) {
	if repository == nil || repository.pool == nil {
		return nil, errAgentCopilotProfileStore
	}
	rows, err := repository.pool.Query(agentCopilotProfileDatabaseContext(ctx), `SELECT sanitized_draft_payload FROM agent_copilot_profile_drafts
		WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 ORDER BY updated_at DESC,profile_id ASC LIMIT 200`,
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

func (repository *postgresAgentCopilotProfileRepository) CreateVersion(ctx AgentCopilotProfileContext, version AgentCopilotProfileVersionV1) (AgentCopilotProfileVersionV1, error) {
	if repository == nil || repository.pool == nil {
		return AgentCopilotProfileVersionV1{}, errAgentCopilotProfileStore
	}
	databaseContext := agentCopilotProfileDatabaseContext(ctx)
	transaction, err := repository.pool.Begin(databaseContext)
	if err != nil {
		return AgentCopilotProfileVersionV1{}, errAgentCopilotProfileStore
	}
	defer func() { _ = transaction.Rollback(databaseContext) }()
	draft, err := scanStoredAgentCopilotProfileDraft(ctx, transaction.QueryRow(databaseContext, `SELECT sanitized_draft_payload
		FROM agent_copilot_profile_drafts WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 AND profile_id=$5 FOR UPDATE`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, version.ProfileID))
	if errors.Is(err, pgx.ErrNoRows) {
		return AgentCopilotProfileVersionV1{}, errAgentCopilotProfileNotFound
	}
	if err != nil {
		return AgentCopilotProfileVersionV1{}, agentCopilotProfileStoredError(err)
	}
	if draft.DraftVersion != version.SourceDraftVersion || draft.ProfileDigest != version.ProfileDigest || draft.PolicyDigest != version.PolicyDigest {
		return AgentCopilotProfileVersionV1{}, agentCopilotProfileVersionConflictError{CurrentDraftVersion: draft.DraftVersion}
	}
	var nextVersion int
	if err := transaction.QueryRow(databaseContext, `SELECT COALESCE(max(profile_version),0)+1 FROM agent_copilot_profile_versions
		WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 AND profile_id=$5`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, version.ProfileID).Scan(&nextVersion); err != nil {
		return AgentCopilotProfileVersionV1{}, errAgentCopilotProfileStore
	}
	var sourceExists bool
	if err := transaction.QueryRow(databaseContext, `SELECT EXISTS(SELECT 1 FROM agent_copilot_profile_versions
		WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 AND profile_id=$5 AND source_draft_version=$6)`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, version.ProfileID, version.SourceDraftVersion).Scan(&sourceExists); err != nil {
		return AgentCopilotProfileVersionV1{}, errAgentCopilotProfileStore
	}
	if sourceExists {
		return AgentCopilotProfileVersionV1{}, errAgentCopilotProfileImmutable
	}
	version.ProfileVersion = nextVersion
	if err := validateStoredAgentCopilotProfileVersion(ctx, version); err != nil {
		return AgentCopilotProfileVersionV1{}, err
	}
	payload, err := json.Marshal(version)
	if err != nil {
		return AgentCopilotProfileVersionV1{}, errAgentCopilotProfileStore
	}
	created, err := scanStoredAgentCopilotProfileVersion(ctx, transaction.QueryRow(databaseContext, `INSERT INTO agent_copilot_profile_versions
		(tenant_ref,workspace_id,application_id,owner_subject_ref,profile_id,profile_version,source_draft_version,profile_digest,policy_digest,published_at,immutable_version_payload)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING immutable_version_payload`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef,
		version.ProfileID, version.ProfileVersion, version.SourceDraftVersion, version.ProfileDigest, version.PolicyDigest, version.PublishedAt, payload))
	if err != nil {
		return AgentCopilotProfileVersionV1{}, agentCopilotProfileStoredError(err)
	}
	if transaction.Commit(databaseContext) != nil {
		return AgentCopilotProfileVersionV1{}, errAgentCopilotProfileStore
	}
	return created, nil
}

func (repository *postgresAgentCopilotProfileRepository) ReadVersion(ctx AgentCopilotProfileContext, profileID string, profileVersion int) (AgentCopilotProfileVersionV1, error) {
	if repository == nil || repository.pool == nil {
		return AgentCopilotProfileVersionV1{}, errAgentCopilotProfileStore
	}
	version, err := scanStoredAgentCopilotProfileVersion(ctx, repository.pool.QueryRow(agentCopilotProfileDatabaseContext(ctx), `SELECT immutable_version_payload
		FROM agent_copilot_profile_versions WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 AND profile_id=$5 AND profile_version=$6`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, profileID, profileVersion))
	if errors.Is(err, pgx.ErrNoRows) {
		return AgentCopilotProfileVersionV1{}, errAgentCopilotProfileNotFound
	}
	if err != nil {
		return AgentCopilotProfileVersionV1{}, agentCopilotProfileStoredError(err)
	}
	return version, nil
}

func (repository *postgresAgentCopilotProfileRepository) ListVersions(ctx AgentCopilotProfileContext, profileID string) ([]AgentCopilotProfileVersionSummary, error) {
	if repository == nil || repository.pool == nil {
		return nil, errAgentCopilotProfileStore
	}
	databaseContext := agentCopilotProfileDatabaseContext(ctx)
	var draftExists bool
	if err := repository.pool.QueryRow(databaseContext, `SELECT EXISTS(SELECT 1 FROM agent_copilot_profile_drafts
		WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 AND profile_id=$5)`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, profileID).Scan(&draftExists); err != nil {
		return nil, errAgentCopilotProfileStore
	}
	if !draftExists {
		return nil, errAgentCopilotProfileNotFound
	}
	rows, err := repository.pool.Query(databaseContext, `SELECT immutable_version_payload FROM agent_copilot_profile_versions
		WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 AND profile_id=$5 ORDER BY profile_version DESC`,
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
