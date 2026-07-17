package httpapi

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type sqliteWorkflowRAGSnapshotRepository struct{ database *sql.DB }

func newSQLiteWorkflowRAGSnapshotRepository(database *sql.DB) *sqliteWorkflowRAGSnapshotRepository {
	return &sqliteWorkflowRAGSnapshotRepository{database: database}
}

func (repository *sqliteWorkflowRAGSnapshotRepository) Create(ctx WorkflowRAGSnapshotContext, resource WorkflowRAGSnapshotResource, record WorkflowRAGSnapshotRecord, audit WorkflowRAGExecutionAudit) error {
	resourcePayload, recordPayload, auditPayload, err := encodeWorkflowRAGCreatePayloads(ctx, resource, record, audit)
	if err != nil || repository == nil || repository.database == nil {
		return errWorkflowRAGStoreContract
	}
	tx, err := repository.database.BeginTx(workflowRAGDatabaseContext(ctx), nil)
	if err != nil {
		return errWorkflowRAGStoreUnavailable
	}
	defer func() { _ = tx.Rollback() }()
	createdAt, _ := time.Parse(time.RFC3339Nano, resource.CreatedAt)
	result, err := tx.ExecContext(workflowRAGDatabaseContext(ctx), `INSERT INTO workflow_rag_snapshot_resources (
	 tenant_ref,workspace_id,application_id,snapshot_id,snapshot_key,lifecycle_state,latest_version,
	 latest_digest,created_at_unix_nano,updated_at_unix_nano,archived_at_unix_nano,sanitized_resource_payload
	) VALUES (?,?,?,?,?,?,?,?,?,?,NULL,?) ON CONFLICT DO NOTHING`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, resource.SnapshotID, resource.SnapshotKey,
		resource.LifecycleState, resource.LatestVersion, resource.LatestDigest, createdAt.UnixNano(), createdAt.UnixNano(), string(resourcePayload))
	if err != nil {
		return errWorkflowRAGStoreUnavailable
	}
	if affected, rowsErr := result.RowsAffected(); rowsErr != nil || affected != 1 {
		return workflowRAGVersionConflictError{}
	}
	if err = insertSQLiteWorkflowRAGVersion(ctx, tx, record, recordPayload); err != nil {
		return err
	}
	if err = insertSQLiteWorkflowRAGAudit(ctx, tx, audit, auditPayload); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return errWorkflowRAGStoreUnavailable
	}
	return nil
}

func (repository *sqliteWorkflowRAGSnapshotRepository) List(ctx WorkflowRAGSnapshotContext, query workflowRAGSnapshotListQuery) ([]WorkflowRAGSnapshotResource, error) {
	if repository == nil || repository.database == nil {
		return nil, errWorkflowRAGStoreUnavailable
	}
	rows, err := repository.database.QueryContext(workflowRAGDatabaseContext(ctx), `SELECT sanitized_resource_payload
	 FROM workflow_rag_snapshot_resources WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND lifecycle_state=?
	 AND (?='' OR snapshot_key>?) ORDER BY snapshot_key ASC LIMIT ?`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, query.LifecycleState, query.AfterSnapshotKey, query.AfterSnapshotKey, query.Limit)
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

func (repository *sqliteWorkflowRAGSnapshotRepository) ReadVersion(ctx WorkflowRAGSnapshotContext, snapshotID string, version int) (WorkflowRAGSnapshotResource, WorkflowRAGSnapshotRecord, error) {
	if repository == nil || repository.database == nil {
		return WorkflowRAGSnapshotResource{}, WorkflowRAGSnapshotRecord{}, errWorkflowRAGStoreUnavailable
	}
	resource, err := scanWorkflowRAGResource(repository.database.QueryRowContext(workflowRAGDatabaseContext(ctx), `SELECT sanitized_resource_payload
	 FROM workflow_rag_snapshot_resources WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND snapshot_id=?`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, snapshotID), ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return WorkflowRAGSnapshotResource{}, WorkflowRAGSnapshotRecord{}, repository.sqliteMissingScopeResult(ctx, snapshotID)
	}
	if err != nil {
		return WorkflowRAGSnapshotResource{}, WorkflowRAGSnapshotRecord{}, workflowRAGStoreError(err)
	}
	record, err := readSQLiteWorkflowRAGVersion(ctx, repository.database, snapshotID, version)
	if errors.Is(err, sql.ErrNoRows) {
		return WorkflowRAGSnapshotResource{}, WorkflowRAGSnapshotRecord{}, errWorkflowRAGNotFound
	}
	if err != nil {
		return WorkflowRAGSnapshotResource{}, WorkflowRAGSnapshotRecord{}, workflowRAGStoreError(err)
	}
	return resource, record, nil
}

func (repository *sqliteWorkflowRAGSnapshotRepository) ReadLatest(ctx WorkflowRAGSnapshotContext, snapshotID string) (WorkflowRAGSnapshotResource, WorkflowRAGSnapshotRecord, error) {
	if repository == nil || repository.database == nil {
		return WorkflowRAGSnapshotResource{}, WorkflowRAGSnapshotRecord{}, errWorkflowRAGStoreUnavailable
	}
	resource, err := scanWorkflowRAGResource(repository.database.QueryRowContext(workflowRAGDatabaseContext(ctx), `SELECT sanitized_resource_payload
	 FROM workflow_rag_snapshot_resources WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND snapshot_id=?`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, snapshotID), ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return WorkflowRAGSnapshotResource{}, WorkflowRAGSnapshotRecord{}, repository.sqliteMissingScopeResult(ctx, snapshotID)
	}
	if err != nil {
		return WorkflowRAGSnapshotResource{}, WorkflowRAGSnapshotRecord{}, workflowRAGStoreError(err)
	}
	record, err := readSQLiteWorkflowRAGVersion(ctx, repository.database, snapshotID, resource.LatestVersion)
	if err != nil {
		return WorkflowRAGSnapshotResource{}, WorkflowRAGSnapshotRecord{}, workflowRAGStoreError(err)
	}
	return resource, record, nil
}

func (repository *sqliteWorkflowRAGSnapshotRepository) CreateVersion(ctx WorkflowRAGSnapshotContext, snapshotID string, expectedVersion int, resource WorkflowRAGSnapshotResource, record WorkflowRAGSnapshotRecord, audit WorkflowRAGExecutionAudit) error {
	resourcePayload, recordPayload, auditPayload, err := encodeWorkflowRAGCreatePayloads(ctx, resource, record, audit)
	if err != nil || repository == nil || repository.database == nil || record.SnapshotVersion != expectedVersion+1 {
		return errWorkflowRAGStoreContract
	}
	tx, err := repository.database.BeginTx(workflowRAGDatabaseContext(ctx), nil)
	if err != nil {
		return errWorkflowRAGStoreUnavailable
	}
	defer func() { _ = tx.Rollback() }()
	updatedAt, _ := time.Parse(time.RFC3339Nano, resource.UpdatedAt)
	result, err := tx.ExecContext(workflowRAGDatabaseContext(ctx), `UPDATE workflow_rag_snapshot_resources SET
	 lifecycle_state=?,latest_version=?,latest_digest=?,updated_at_unix_nano=?,sanitized_resource_payload=?
	 WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND snapshot_id=? AND latest_version=? AND lifecycle_state='active'`,
		resource.LifecycleState, resource.LatestVersion, resource.LatestDigest, updatedAt.UnixNano(), string(resourcePayload),
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, snapshotID, expectedVersion)
	if err != nil {
		return errWorkflowRAGStoreUnavailable
	}
	if affected, rowsErr := result.RowsAffected(); rowsErr != nil || affected != 1 {
		return repository.sqliteMutationConflict(ctx, tx, snapshotID, expectedVersion)
	}
	if err = insertSQLiteWorkflowRAGVersion(ctx, tx, record, recordPayload); err != nil {
		return err
	}
	if err = insertSQLiteWorkflowRAGAudit(ctx, tx, audit, auditPayload); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return errWorkflowRAGStoreUnavailable
	}
	return nil
}

func (repository *sqliteWorkflowRAGSnapshotRepository) Archive(ctx WorkflowRAGSnapshotContext, snapshotID string, expectedVersion int, resource WorkflowRAGSnapshotResource, audit WorkflowRAGExecutionAudit) error {
	resourcePayload, err := encodeWorkflowRAGResource(resource)
	if err != nil || repository == nil || repository.database == nil || resource.LifecycleState != workflowRAGSnapshotArchived || resource.ArchivedAt == nil {
		return errWorkflowRAGStoreContract
	}
	auditPayload, err := encodeWorkflowRAGAudit(audit, ctx)
	if err != nil {
		return errWorkflowRAGStoreContract
	}
	tx, err := repository.database.BeginTx(workflowRAGDatabaseContext(ctx), nil)
	if err != nil {
		return errWorkflowRAGStoreUnavailable
	}
	defer func() { _ = tx.Rollback() }()
	updatedAt, _ := time.Parse(time.RFC3339Nano, resource.UpdatedAt)
	archivedAt, _ := time.Parse(time.RFC3339Nano, *resource.ArchivedAt)
	result, err := tx.ExecContext(workflowRAGDatabaseContext(ctx), `UPDATE workflow_rag_snapshot_resources SET
	 lifecycle_state='archived',updated_at_unix_nano=?,archived_at_unix_nano=?,sanitized_resource_payload=?
	 WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND snapshot_id=? AND latest_version=? AND lifecycle_state='active'`,
		updatedAt.UnixNano(), archivedAt.UnixNano(), string(resourcePayload), ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, snapshotID, expectedVersion)
	if err != nil {
		return errWorkflowRAGStoreUnavailable
	}
	if affected, rowsErr := result.RowsAffected(); rowsErr != nil || affected != 1 {
		return repository.sqliteMutationConflict(ctx, tx, snapshotID, expectedVersion)
	}
	if err = insertSQLiteWorkflowRAGAudit(ctx, tx, audit, auditPayload); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return errWorkflowRAGStoreUnavailable
	}
	return nil
}

func encodeWorkflowRAGCreatePayloads(ctx WorkflowRAGSnapshotContext, resource WorkflowRAGSnapshotResource, record WorkflowRAGSnapshotRecord, audit WorkflowRAGExecutionAudit) ([]byte, []byte, []byte, error) {
	resourcePayload, err := encodeWorkflowRAGResource(resource)
	if err != nil {
		return nil, nil, nil, err
	}
	recordPayload, err := encodeWorkflowRAGSnapshotMetadata(record, ctx)
	if err != nil {
		return nil, nil, nil, err
	}
	auditPayload, err := encodeWorkflowRAGAudit(audit, ctx)
	return resourcePayload, recordPayload, auditPayload, err
}

func insertSQLiteWorkflowRAGVersion(ctx WorkflowRAGSnapshotContext, tx *sql.Tx, record WorkflowRAGSnapshotRecord, payload []byte) error {
	createdAt, err := time.Parse(time.RFC3339Nano, record.CreatedAt)
	if err != nil {
		return errWorkflowRAGStoreContract
	}
	_, err = tx.ExecContext(workflowRAGDatabaseContext(ctx), `INSERT INTO workflow_rag_snapshot_versions (
	 tenant_ref,workspace_id,application_id,snapshot_id,snapshot_version,snapshot_digest,created_at_unix_nano,sanitized_snapshot_payload
	) VALUES (?,?,?,?,?,?,?,?)`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, record.SnapshotID,
		record.SnapshotVersion, record.SnapshotDigest, createdAt.UnixNano(), string(payload))
	if err != nil {
		return errWorkflowRAGStoreUnavailable
	}
	for _, fragment := range record.Fragments {
		fragmentPayload, encodeErr := encodeWorkflowRAGFragment(fragment)
		if encodeErr != nil {
			return errWorkflowRAGStoreContract
		}
		_, err = tx.ExecContext(workflowRAGDatabaseContext(ctx), `INSERT INTO workflow_rag_snapshot_fragments (
		 tenant_ref,workspace_id,application_id,snapshot_id,snapshot_version,fragment_ref,content_digest,sanitized_fragment_payload
		) VALUES (?,?,?,?,?,?,?,?)`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, record.SnapshotID,
			record.SnapshotVersion, fragment.FragmentRef, fragment.ContentDigest, string(fragmentPayload))
		if err != nil {
			return errWorkflowRAGStoreUnavailable
		}
	}
	return nil
}

func insertSQLiteWorkflowRAGAudit(ctx WorkflowRAGSnapshotContext, tx *sql.Tx, audit WorkflowRAGExecutionAudit, payload []byte) error {
	occurredAt, err := time.Parse(time.RFC3339Nano, audit.OccurredAt)
	if err != nil {
		return errWorkflowRAGStoreContract
	}
	_, err = tx.ExecContext(workflowRAGDatabaseContext(ctx), `INSERT INTO workflow_rag_execution_audits (
	 tenant_ref,workspace_id,application_id,snapshot_id,event_id,event_kind,snapshot_version,snapshot_digest,
	 actor_ref,request_id,audit_ref,occurred_at_unix_nano,sanitized_audit_payload
	) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, audit.SnapshotID,
		audit.EventID, audit.EventKind, audit.SnapshotVersion, audit.SnapshotDigest, audit.ActorRef, audit.RequestID,
		audit.AuditRef, occurredAt.UnixNano(), string(payload))
	if err != nil {
		return errWorkflowRAGStoreUnavailable
	}
	return nil
}

type sqliteWorkflowRAGVersionReader interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func readSQLiteWorkflowRAGVersion(ctx WorkflowRAGSnapshotContext, reader sqliteWorkflowRAGVersionReader, snapshotID string, version int) (WorkflowRAGSnapshotRecord, error) {
	var payload string
	err := reader.QueryRowContext(workflowRAGDatabaseContext(ctx), `SELECT sanitized_snapshot_payload FROM workflow_rag_snapshot_versions
	 WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND snapshot_id=? AND snapshot_version=?`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, snapshotID, version).Scan(&payload)
	if err != nil {
		return WorkflowRAGSnapshotRecord{}, err
	}
	rows, err := reader.QueryContext(workflowRAGDatabaseContext(ctx), `SELECT sanitized_fragment_payload FROM workflow_rag_snapshot_fragments
	 WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND snapshot_id=? AND snapshot_version=? ORDER BY fragment_ref`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, snapshotID, version)
	if err != nil {
		return WorkflowRAGSnapshotRecord{}, errWorkflowRAGStoreUnavailable
	}
	defer rows.Close()
	fragments := make([]WorkflowRAGFragment, 0)
	for rows.Next() {
		var fragmentPayload string
		if err = rows.Scan(&fragmentPayload); err != nil {
			return WorkflowRAGSnapshotRecord{}, errWorkflowRAGStoreUnavailable
		}
		fragment, decodeErr := decodeWorkflowRAGFragment([]byte(fragmentPayload))
		if decodeErr != nil {
			return WorkflowRAGSnapshotRecord{}, decodeErr
		}
		fragments = append(fragments, fragment)
	}
	if rows.Err() != nil {
		return WorkflowRAGSnapshotRecord{}, errWorkflowRAGStoreUnavailable
	}
	return decodeWorkflowRAGSnapshotMetadata([]byte(payload), fragments, ctx)
}

func (repository *sqliteWorkflowRAGSnapshotRepository) sqliteMissingScopeResult(ctx WorkflowRAGSnapshotContext, snapshotID string) error {
	var count int
	if err := repository.database.QueryRowContext(workflowRAGDatabaseContext(ctx), `SELECT count(*) FROM workflow_rag_snapshot_resources WHERE snapshot_id=?`, snapshotID).Scan(&count); err != nil {
		return errWorkflowRAGStoreUnavailable
	}
	if count > 0 {
		return errWorkflowRAGScopeDenied
	}
	return errWorkflowRAGNotFound
}

func (repository *sqliteWorkflowRAGSnapshotRepository) sqliteMutationConflict(ctx WorkflowRAGSnapshotContext, tx *sql.Tx, snapshotID string, expectedVersion int) error {
	resource, err := scanWorkflowRAGResource(tx.QueryRowContext(workflowRAGDatabaseContext(ctx), `SELECT sanitized_resource_payload
	 FROM workflow_rag_snapshot_resources WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND snapshot_id=?`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, snapshotID), ctx)
	if errors.Is(err, sql.ErrNoRows) {
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

func workflowRAGDatabaseContext(ctx WorkflowRAGSnapshotContext) context.Context {
	if ctx.RequestContext != nil {
		return ctx.RequestContext
	}
	return context.Background()
}

var _ workflowRAGSnapshotRepository = (*sqliteWorkflowRAGSnapshotRepository)(nil)
