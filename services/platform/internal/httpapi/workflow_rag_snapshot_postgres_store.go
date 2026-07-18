package httpapi

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresWorkflowRAGSnapshotRepository struct{ pool *pgxpool.Pool }

func newPostgresWorkflowRAGSnapshotRepository(pool *pgxpool.Pool) *postgresWorkflowRAGSnapshotRepository {
	return &postgresWorkflowRAGSnapshotRepository{pool: pool}
}

func (repository *postgresWorkflowRAGSnapshotRepository) Create(ctx WorkflowRAGSnapshotContext, resource WorkflowRAGSnapshotResource, record WorkflowRAGSnapshotRecord, audit WorkflowRAGExecutionAudit) error {
	resourcePayload, recordPayload, auditPayload, err := encodeWorkflowRAGCreatePayloads(ctx, resource, record, audit)
	if err != nil || repository == nil || repository.pool == nil {
		return errWorkflowRAGStoreContract
	}
	tx, err := repository.pool.Begin(workflowRAGDatabaseContext(ctx))
	if err != nil {
		return errWorkflowRAGStoreUnavailable
	}
	defer func() { _ = tx.Rollback(workflowRAGDatabaseContext(ctx)) }()
	createdAt, _ := time.Parse(time.RFC3339Nano, resource.CreatedAt)
	result, err := tx.Exec(workflowRAGDatabaseContext(ctx), `INSERT INTO workflow_rag_snapshot_resources (
	 tenant_ref,workspace_id,application_id,snapshot_id,snapshot_key,lifecycle_state,latest_version,
	 latest_digest,created_at,updated_at,archived_at,sanitized_resource_payload
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$9,NULL,$10) ON CONFLICT DO NOTHING`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, resource.SnapshotID, resource.SnapshotKey,
		resource.LifecycleState, resource.LatestVersion, resource.LatestDigest, createdAt, resourcePayload)
	if err != nil {
		return errWorkflowRAGStoreUnavailable
	}
	if result.RowsAffected() != 1 {
		return workflowRAGVersionConflictError{}
	}
	if err = insertPostgresWorkflowRAGVersion(ctx, tx, record, recordPayload); err != nil {
		return err
	}
	if err = insertPostgresWorkflowRAGAudit(ctx, tx, audit, auditPayload); err != nil {
		return err
	}
	if err = tx.Commit(workflowRAGDatabaseContext(ctx)); err != nil {
		return errWorkflowRAGStoreUnavailable
	}
	return nil
}

func (repository *postgresWorkflowRAGSnapshotRepository) List(ctx WorkflowRAGSnapshotContext, query workflowRAGSnapshotListQuery) ([]WorkflowRAGSnapshotResource, error) {
	if repository == nil || repository.pool == nil {
		return nil, errWorkflowRAGStoreUnavailable
	}
	rows, err := repository.pool.Query(workflowRAGDatabaseContext(ctx), `SELECT sanitized_resource_payload
	 FROM workflow_rag_snapshot_resources WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND lifecycle_state=$4
	 AND ($5='' OR snapshot_key>$5) ORDER BY snapshot_key ASC LIMIT $6`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, query.LifecycleState, query.AfterSnapshotKey, query.Limit)
	if err != nil {
		return nil, errWorkflowRAGStoreUnavailable
	}
	defer rows.Close()
	resources := make([]WorkflowRAGSnapshotResource, 0)
	for rows.Next() {
		resource, scanErr := scanWorkflowRAGResource(rows, ctx)
		if scanErr != nil {
			return nil, errWorkflowRAGStoreContract
		}
		resources = append(resources, resource)
	}
	if rows.Err() != nil {
		return nil, errWorkflowRAGStoreUnavailable
	}
	return resources, nil
}

func (repository *postgresWorkflowRAGSnapshotRepository) ReadVersion(ctx WorkflowRAGSnapshotContext, snapshotID string, version int) (WorkflowRAGSnapshotResource, WorkflowRAGSnapshotRecord, error) {
	if repository == nil || repository.pool == nil {
		return WorkflowRAGSnapshotResource{}, WorkflowRAGSnapshotRecord{}, errWorkflowRAGStoreUnavailable
	}
	resource, err := scanWorkflowRAGResource(repository.pool.QueryRow(workflowRAGDatabaseContext(ctx), `SELECT sanitized_resource_payload
	 FROM workflow_rag_snapshot_resources WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND snapshot_id=$4`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, snapshotID), ctx)
	if errors.Is(err, pgx.ErrNoRows) {
		return WorkflowRAGSnapshotResource{}, WorkflowRAGSnapshotRecord{}, repository.postgresMissingScopeResult(ctx, snapshotID)
	}
	if err != nil {
		return WorkflowRAGSnapshotResource{}, WorkflowRAGSnapshotRecord{}, workflowRAGStoreError(err)
	}
	record, err := readPostgresWorkflowRAGVersion(ctx, repository.pool, snapshotID, version)
	if errors.Is(err, pgx.ErrNoRows) {
		return WorkflowRAGSnapshotResource{}, WorkflowRAGSnapshotRecord{}, errWorkflowRAGNotFound
	}
	if err != nil {
		return WorkflowRAGSnapshotResource{}, WorkflowRAGSnapshotRecord{}, workflowRAGStoreError(err)
	}
	return resource, record, nil
}

func (repository *postgresWorkflowRAGSnapshotRepository) ReadByRAGRef(ctx WorkflowRAGSnapshotContext, ragRef string) (WorkflowRAGSnapshotResource, WorkflowRAGSnapshotRecord, error) {
	key, version, ok := parseWorkflowRAGRAGRef(ragRef)
	if !ok || repository == nil || repository.pool == nil {
		return WorkflowRAGSnapshotResource{}, WorkflowRAGSnapshotRecord{}, errWorkflowRAGNotFound
	}
	resource, err := scanWorkflowRAGResource(repository.pool.QueryRow(workflowRAGDatabaseContext(ctx), `SELECT sanitized_resource_payload
	 FROM workflow_rag_snapshot_resources WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND snapshot_key=$4`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, key), ctx)
	if errors.Is(err, pgx.ErrNoRows) {
		var count int
		if countErr := repository.pool.QueryRow(workflowRAGDatabaseContext(ctx), `SELECT count(*) FROM workflow_rag_snapshot_resources WHERE snapshot_key=$1`, key).Scan(&count); countErr != nil {
			return WorkflowRAGSnapshotResource{}, WorkflowRAGSnapshotRecord{}, errWorkflowRAGStoreUnavailable
		}
		if count > 0 {
			return WorkflowRAGSnapshotResource{}, WorkflowRAGSnapshotRecord{}, errWorkflowRAGScopeDenied
		}
		return WorkflowRAGSnapshotResource{}, WorkflowRAGSnapshotRecord{}, errWorkflowRAGNotFound
	}
	if err != nil {
		return WorkflowRAGSnapshotResource{}, WorkflowRAGSnapshotRecord{}, workflowRAGStoreError(err)
	}
	record, err := readPostgresWorkflowRAGVersion(ctx, repository.pool, resource.SnapshotID, version)
	if errors.Is(err, pgx.ErrNoRows) {
		return WorkflowRAGSnapshotResource{}, WorkflowRAGSnapshotRecord{}, errWorkflowRAGNotFound
	}
	if err != nil {
		return WorkflowRAGSnapshotResource{}, WorkflowRAGSnapshotRecord{}, workflowRAGStoreError(err)
	}
	return resource, record, nil
}

func (repository *postgresWorkflowRAGSnapshotRepository) ReadLatest(ctx WorkflowRAGSnapshotContext, snapshotID string) (WorkflowRAGSnapshotResource, WorkflowRAGSnapshotRecord, error) {
	if repository == nil || repository.pool == nil {
		return WorkflowRAGSnapshotResource{}, WorkflowRAGSnapshotRecord{}, errWorkflowRAGStoreUnavailable
	}
	resource, err := scanWorkflowRAGResource(repository.pool.QueryRow(workflowRAGDatabaseContext(ctx), `SELECT sanitized_resource_payload
	 FROM workflow_rag_snapshot_resources WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND snapshot_id=$4`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, snapshotID), ctx)
	if errors.Is(err, pgx.ErrNoRows) {
		return WorkflowRAGSnapshotResource{}, WorkflowRAGSnapshotRecord{}, repository.postgresMissingScopeResult(ctx, snapshotID)
	}
	if err != nil {
		return WorkflowRAGSnapshotResource{}, WorkflowRAGSnapshotRecord{}, workflowRAGStoreError(err)
	}
	record, err := readPostgresWorkflowRAGVersion(ctx, repository.pool, snapshotID, resource.LatestVersion)
	if err != nil {
		return WorkflowRAGSnapshotResource{}, WorkflowRAGSnapshotRecord{}, workflowRAGStoreError(err)
	}
	return resource, record, nil
}

func (repository *postgresWorkflowRAGSnapshotRepository) CreateVersion(ctx WorkflowRAGSnapshotContext, snapshotID string, expectedVersion int, resource WorkflowRAGSnapshotResource, record WorkflowRAGSnapshotRecord, audit WorkflowRAGExecutionAudit) error {
	resourcePayload, recordPayload, auditPayload, err := encodeWorkflowRAGCreatePayloads(ctx, resource, record, audit)
	if err != nil || repository == nil || repository.pool == nil || record.SnapshotVersion != expectedVersion+1 {
		return errWorkflowRAGStoreContract
	}
	tx, err := repository.pool.Begin(workflowRAGDatabaseContext(ctx))
	if err != nil {
		return errWorkflowRAGStoreUnavailable
	}
	defer func() { _ = tx.Rollback(workflowRAGDatabaseContext(ctx)) }()
	updatedAt, _ := time.Parse(time.RFC3339Nano, resource.UpdatedAt)
	result, err := tx.Exec(workflowRAGDatabaseContext(ctx), `UPDATE workflow_rag_snapshot_resources SET
	 lifecycle_state=$1,latest_version=$2,latest_digest=$3,updated_at=$4,sanitized_resource_payload=$5
	 WHERE tenant_ref=$6 AND workspace_id=$7 AND application_id=$8 AND snapshot_id=$9 AND latest_version=$10 AND lifecycle_state='active'`,
		resource.LifecycleState, resource.LatestVersion, resource.LatestDigest, updatedAt, resourcePayload,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, snapshotID, expectedVersion)
	if err != nil {
		return errWorkflowRAGStoreUnavailable
	}
	if result.RowsAffected() != 1 {
		return repository.postgresMutationConflict(ctx, tx, snapshotID, expectedVersion)
	}
	if err = insertPostgresWorkflowRAGVersion(ctx, tx, record, recordPayload); err != nil {
		return err
	}
	if err = insertPostgresWorkflowRAGAudit(ctx, tx, audit, auditPayload); err != nil {
		return err
	}
	if err = tx.Commit(workflowRAGDatabaseContext(ctx)); err != nil {
		return errWorkflowRAGStoreUnavailable
	}
	return nil
}

func (repository *postgresWorkflowRAGSnapshotRepository) Archive(ctx WorkflowRAGSnapshotContext, snapshotID string, expectedVersion int, resource WorkflowRAGSnapshotResource, audit WorkflowRAGExecutionAudit) error {
	resourcePayload, err := encodeWorkflowRAGResource(resource)
	if err != nil || repository == nil || repository.pool == nil || resource.LifecycleState != workflowRAGSnapshotArchived || resource.ArchivedAt == nil {
		return errWorkflowRAGStoreContract
	}
	auditPayload, err := encodeWorkflowRAGAudit(audit, ctx)
	if err != nil {
		return errWorkflowRAGStoreContract
	}
	tx, err := repository.pool.Begin(workflowRAGDatabaseContext(ctx))
	if err != nil {
		return errWorkflowRAGStoreUnavailable
	}
	defer func() { _ = tx.Rollback(workflowRAGDatabaseContext(ctx)) }()
	updatedAt, _ := time.Parse(time.RFC3339Nano, resource.UpdatedAt)
	archivedAt, _ := time.Parse(time.RFC3339Nano, *resource.ArchivedAt)
	result, err := tx.Exec(workflowRAGDatabaseContext(ctx), `UPDATE workflow_rag_snapshot_resources SET
	 lifecycle_state='archived',updated_at=$1,archived_at=$2,sanitized_resource_payload=$3
	 WHERE tenant_ref=$4 AND workspace_id=$5 AND application_id=$6 AND snapshot_id=$7 AND latest_version=$8 AND lifecycle_state='active'`,
		updatedAt, archivedAt, resourcePayload, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, snapshotID, expectedVersion)
	if err != nil {
		return errWorkflowRAGStoreUnavailable
	}
	if result.RowsAffected() != 1 {
		return repository.postgresMutationConflict(ctx, tx, snapshotID, expectedVersion)
	}
	if err = insertPostgresWorkflowRAGAudit(ctx, tx, audit, auditPayload); err != nil {
		return err
	}
	if err = tx.Commit(workflowRAGDatabaseContext(ctx)); err != nil {
		return errWorkflowRAGStoreUnavailable
	}
	return nil
}

func insertPostgresWorkflowRAGVersion(ctx WorkflowRAGSnapshotContext, tx pgx.Tx, record WorkflowRAGSnapshotRecord, payload []byte) error {
	createdAt, err := time.Parse(time.RFC3339Nano, record.CreatedAt)
	if err != nil {
		return errWorkflowRAGStoreContract
	}
	_, err = tx.Exec(workflowRAGDatabaseContext(ctx), `INSERT INTO workflow_rag_snapshot_versions (
	 tenant_ref,workspace_id,application_id,snapshot_id,snapshot_version,snapshot_digest,created_at,sanitized_snapshot_payload
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, record.SnapshotID,
		record.SnapshotVersion, record.SnapshotDigest, createdAt, payload)
	if err != nil {
		return errWorkflowRAGStoreUnavailable
	}
	for _, fragment := range record.Fragments {
		fragmentPayload, encodeErr := encodeWorkflowRAGFragment(fragment)
		if encodeErr != nil {
			return errWorkflowRAGStoreContract
		}
		_, err = tx.Exec(workflowRAGDatabaseContext(ctx), `INSERT INTO workflow_rag_snapshot_fragments (
		 tenant_ref,workspace_id,application_id,snapshot_id,snapshot_version,fragment_ref,content_digest,sanitized_fragment_payload
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, record.SnapshotID,
			record.SnapshotVersion, fragment.FragmentRef, fragment.ContentDigest, fragmentPayload)
		if err != nil {
			return errWorkflowRAGStoreUnavailable
		}
	}
	return nil
}

func insertPostgresWorkflowRAGAudit(ctx WorkflowRAGSnapshotContext, tx pgx.Tx, audit WorkflowRAGExecutionAudit, payload []byte) error {
	occurredAt, err := time.Parse(time.RFC3339Nano, audit.OccurredAt)
	if err != nil {
		return errWorkflowRAGStoreContract
	}
	_, err = tx.Exec(workflowRAGDatabaseContext(ctx), `INSERT INTO workflow_rag_execution_audits (
	 tenant_ref,workspace_id,application_id,snapshot_id,event_id,event_kind,snapshot_version,snapshot_digest,
	 actor_ref,request_id,audit_ref,occurred_at,sanitized_audit_payload
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID,
		audit.SnapshotID, audit.EventID, audit.EventKind, audit.SnapshotVersion, audit.SnapshotDigest, audit.ActorRef,
		audit.RequestID, audit.AuditRef, occurredAt, payload)
	if err != nil {
		return errWorkflowRAGStoreUnavailable
	}
	return nil
}

func (repository *postgresWorkflowRAGSnapshotRepository) AppendAudit(ctx WorkflowRAGSnapshotContext, audit WorkflowRAGExecutionAudit) error {
	payload, err := encodeWorkflowRAGAudit(audit, ctx)
	if err != nil || repository == nil || repository.pool == nil {
		return errWorkflowRAGStoreContract
	}
	tx, err := repository.pool.Begin(workflowRAGDatabaseContext(ctx))
	if err != nil {
		return errWorkflowRAGStoreUnavailable
	}
	defer func() { _ = tx.Rollback(workflowRAGDatabaseContext(ctx)) }()
	if err = insertPostgresWorkflowRAGAudit(ctx, tx, audit, payload); err != nil {
		return err
	}
	if err = tx.Commit(workflowRAGDatabaseContext(ctx)); err != nil {
		return errWorkflowRAGStoreUnavailable
	}
	return nil
}

type postgresWorkflowRAGVersionReader interface {
	QueryRow(context.Context, string, ...any) pgx.Row
	Query(context.Context, string, ...any) (pgx.Rows, error)
}

func readPostgresWorkflowRAGVersion(ctx WorkflowRAGSnapshotContext, reader postgresWorkflowRAGVersionReader, snapshotID string, version int) (WorkflowRAGSnapshotRecord, error) {
	var payload []byte
	err := reader.QueryRow(workflowRAGDatabaseContext(ctx), `SELECT sanitized_snapshot_payload FROM workflow_rag_snapshot_versions
	 WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND snapshot_id=$4 AND snapshot_version=$5`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, snapshotID, version).Scan(&payload)
	if err != nil {
		return WorkflowRAGSnapshotRecord{}, err
	}
	rows, err := reader.Query(workflowRAGDatabaseContext(ctx), `SELECT sanitized_fragment_payload FROM workflow_rag_snapshot_fragments
	 WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND snapshot_id=$4 AND snapshot_version=$5 ORDER BY fragment_ref`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, snapshotID, version)
	if err != nil {
		return WorkflowRAGSnapshotRecord{}, errWorkflowRAGStoreUnavailable
	}
	defer rows.Close()
	fragments := make([]WorkflowRAGFragment, 0)
	for rows.Next() {
		var fragmentPayload []byte
		if err = rows.Scan(&fragmentPayload); err != nil {
			return WorkflowRAGSnapshotRecord{}, errWorkflowRAGStoreUnavailable
		}
		fragment, decodeErr := decodeWorkflowRAGFragment(fragmentPayload)
		if decodeErr != nil {
			return WorkflowRAGSnapshotRecord{}, decodeErr
		}
		fragments = append(fragments, fragment)
	}
	if rows.Err() != nil {
		return WorkflowRAGSnapshotRecord{}, errWorkflowRAGStoreUnavailable
	}
	return decodeWorkflowRAGSnapshotMetadata(payload, fragments, ctx)
}

func (repository *postgresWorkflowRAGSnapshotRepository) postgresMissingScopeResult(ctx WorkflowRAGSnapshotContext, snapshotID string) error {
	var count int
	if err := repository.pool.QueryRow(workflowRAGDatabaseContext(ctx), `SELECT count(*) FROM workflow_rag_snapshot_resources WHERE snapshot_id=$1`, snapshotID).Scan(&count); err != nil {
		return errWorkflowRAGStoreUnavailable
	}
	if count > 0 {
		return errWorkflowRAGScopeDenied
	}
	return errWorkflowRAGNotFound
}

func (repository *postgresWorkflowRAGSnapshotRepository) postgresMutationConflict(ctx WorkflowRAGSnapshotContext, tx pgx.Tx, snapshotID string, expectedVersion int) error {
	resource, err := scanWorkflowRAGResource(tx.QueryRow(workflowRAGDatabaseContext(ctx), `SELECT sanitized_resource_payload
	 FROM workflow_rag_snapshot_resources WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND snapshot_id=$4`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, snapshotID), ctx)
	if errors.Is(err, pgx.ErrNoRows) {
		return errWorkflowRAGNotFound
	}
	if err != nil {
		return workflowRAGStoreError(err)
	}
	if resource.LatestVersion != expectedVersion {
		return workflowRAGVersionConflictError{CurrentVersion: resource.LatestVersion, CurrentState: resource.LifecycleState}
	}
	if resource.LifecycleState != workflowRAGSnapshotActive {
		return errWorkflowRAGArchived
	}
	return errWorkflowRAGStoreUnavailable
}

var _ workflowRAGSnapshotRepository = (*postgresWorkflowRAGSnapshotRepository)(nil)
