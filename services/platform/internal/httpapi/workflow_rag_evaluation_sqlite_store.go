package httpapi

import (
	"context"
	"database/sql"
	"errors"
)

type sqliteWorkflowRAGEvaluationDatasetRepository struct{ database *sql.DB }

func newSQLiteWorkflowRAGEvaluationDatasetRepository(database *sql.DB) *sqliteWorkflowRAGEvaluationDatasetRepository {
	return &sqliteWorkflowRAGEvaluationDatasetRepository{database: database}
}

func (repository *sqliteWorkflowRAGEvaluationDatasetRepository) Create(
	ctx WorkflowRAGSnapshotContext,
	resource WorkflowRAGEvaluationDatasetResource,
	version WorkflowRAGEvaluationDatasetVersion,
	audit WorkflowRAGEvaluationAudit,
) error {
	resourcePayload, versionPayload, auditPayload, err := encodeWorkflowRAGEvaluationCreatePayloads(ctx, resource, version, audit)
	if err != nil || repository == nil || repository.database == nil || resource.LatestVersion != 1 || audit.EventKind != "dataset_created" {
		return errWorkflowRAGEvaluationContract
	}
	createdAt, err := workflowRAGEvaluationUnixNano(resource.CreatedAt)
	if err != nil {
		return err
	}
	tx, err := repository.database.BeginTx(workflowRAGDatabaseContext(ctx), nil)
	if err != nil {
		return errWorkflowRAGEvaluationStore
	}
	defer func() { _ = tx.Rollback() }()
	result, err := tx.ExecContext(workflowRAGDatabaseContext(ctx), `INSERT INTO workflow_rag_evaluation_dataset_resources (
	 tenant_ref,workspace_id,application_id,dataset_id,dataset_key,lifecycle_state,latest_version,latest_digest,
	 created_at_unix_nano,updated_at_unix_nano,archived_at_unix_nano,sanitized_resource_payload
	) VALUES (?,?,?,?,?,?,?,?,?,?,NULL,?) ON CONFLICT DO NOTHING`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID,
		resource.DatasetID, resource.DatasetKey, resource.LifecycleState, resource.LatestVersion, resource.LatestDigest,
		createdAt, createdAt, string(resourcePayload))
	if err != nil {
		return errWorkflowRAGEvaluationStore
	}
	if affected, rowsErr := result.RowsAffected(); rowsErr != nil || affected != 1 {
		return workflowRAGEvaluationConflictError{}
	}
	if err = insertSQLiteWorkflowRAGEvaluationVersion(ctx, tx, version, versionPayload); err != nil {
		return err
	}
	if err = insertSQLiteWorkflowRAGEvaluationAudit(ctx, tx, audit, auditPayload); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return errWorkflowRAGEvaluationStore
	}
	return nil
}

func (repository *sqliteWorkflowRAGEvaluationDatasetRepository) List(ctx WorkflowRAGSnapshotContext, query workflowRAGEvaluationListQuery) ([]WorkflowRAGEvaluationDatasetResource, error) {
	if repository == nil || repository.database == nil {
		return nil, errWorkflowRAGEvaluationStore
	}
	rows, err := repository.database.QueryContext(workflowRAGDatabaseContext(ctx), `SELECT sanitized_resource_payload
	 FROM workflow_rag_evaluation_dataset_resources
	 WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND lifecycle_state=? AND (?='' OR dataset_key>?)
	 ORDER BY dataset_key ASC LIMIT ?`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, query.LifecycleState,
		query.AfterDatasetKey, query.AfterDatasetKey, query.Limit)
	if err != nil {
		return nil, errWorkflowRAGEvaluationStore
	}
	defer rows.Close()
	resources := make([]WorkflowRAGEvaluationDatasetResource, 0)
	for rows.Next() {
		resource, scanErr := scanSQLiteWorkflowRAGEvaluationResource(rows, ctx)
		if scanErr != nil {
			return nil, evaluationSQLiteStoreError(scanErr)
		}
		resources = append(resources, resource)
	}
	if rows.Err() != nil {
		return nil, errWorkflowRAGEvaluationStore
	}
	return resources, nil
}

func (repository *sqliteWorkflowRAGEvaluationDatasetRepository) ReadLatest(ctx WorkflowRAGSnapshotContext, datasetID string) (WorkflowRAGEvaluationDatasetResource, WorkflowRAGEvaluationDatasetVersion, error) {
	resource, err := repository.readResource(ctx, datasetID)
	if err != nil {
		return WorkflowRAGEvaluationDatasetResource{}, WorkflowRAGEvaluationDatasetVersion{}, err
	}
	version, err := readSQLiteWorkflowRAGEvaluationVersion(ctx, repository.database, datasetID, resource.LatestVersion)
	if err != nil {
		return WorkflowRAGEvaluationDatasetResource{}, WorkflowRAGEvaluationDatasetVersion{}, evaluationSQLiteStoreError(err)
	}
	if !workflowRAGEvaluationResourceVersionMatch(resource, version) {
		return WorkflowRAGEvaluationDatasetResource{}, WorkflowRAGEvaluationDatasetVersion{}, errWorkflowRAGEvaluationContract
	}
	return resource, version, nil
}

func (repository *sqliteWorkflowRAGEvaluationDatasetRepository) ReadVersion(ctx WorkflowRAGSnapshotContext, datasetID string, datasetVersion int) (WorkflowRAGEvaluationDatasetResource, WorkflowRAGEvaluationDatasetVersion, error) {
	resource, err := repository.readResource(ctx, datasetID)
	if err != nil {
		return WorkflowRAGEvaluationDatasetResource{}, WorkflowRAGEvaluationDatasetVersion{}, err
	}
	version, err := readSQLiteWorkflowRAGEvaluationVersion(ctx, repository.database, datasetID, datasetVersion)
	if errors.Is(err, sql.ErrNoRows) {
		return WorkflowRAGEvaluationDatasetResource{}, WorkflowRAGEvaluationDatasetVersion{}, errWorkflowRAGEvaluationNotFound
	}
	if err != nil {
		return WorkflowRAGEvaluationDatasetResource{}, WorkflowRAGEvaluationDatasetVersion{}, evaluationSQLiteStoreError(err)
	}
	return resource, version, nil
}

func (repository *sqliteWorkflowRAGEvaluationDatasetRepository) CreateVersion(
	ctx WorkflowRAGSnapshotContext,
	datasetID string,
	expectedVersion int,
	resource WorkflowRAGEvaluationDatasetResource,
	version WorkflowRAGEvaluationDatasetVersion,
	audit WorkflowRAGEvaluationAudit,
) error {
	resourcePayload, versionPayload, auditPayload, err := encodeWorkflowRAGEvaluationCreatePayloads(ctx, resource, version, audit)
	if err != nil || repository == nil || repository.database == nil || resource.DatasetID != datasetID ||
		version.Dataset.DatasetVersion != expectedVersion+1 || audit.EventKind != "dataset_versioned" {
		return errWorkflowRAGEvaluationContract
	}
	updatedAt, err := workflowRAGEvaluationUnixNano(resource.UpdatedAt)
	if err != nil {
		return err
	}
	tx, err := repository.database.BeginTx(workflowRAGDatabaseContext(ctx), nil)
	if err != nil {
		return errWorkflowRAGEvaluationStore
	}
	defer func() { _ = tx.Rollback() }()
	result, err := tx.ExecContext(workflowRAGDatabaseContext(ctx), `UPDATE workflow_rag_evaluation_dataset_resources SET
	 lifecycle_state=?,latest_version=?,latest_digest=?,updated_at_unix_nano=?,sanitized_resource_payload=?
	 WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND dataset_id=? AND latest_version=? AND lifecycle_state='active'`,
		resource.LifecycleState, resource.LatestVersion, resource.LatestDigest, updatedAt, string(resourcePayload),
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, datasetID, expectedVersion)
	if err != nil {
		return errWorkflowRAGEvaluationStore
	}
	if affected, rowsErr := result.RowsAffected(); rowsErr != nil || affected != 1 {
		return repository.mutationConflict(ctx, tx, datasetID, expectedVersion)
	}
	if err = insertSQLiteWorkflowRAGEvaluationVersion(ctx, tx, version, versionPayload); err != nil {
		return err
	}
	if err = insertSQLiteWorkflowRAGEvaluationAudit(ctx, tx, audit, auditPayload); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return errWorkflowRAGEvaluationStore
	}
	return nil
}

func (repository *sqliteWorkflowRAGEvaluationDatasetRepository) Archive(
	ctx WorkflowRAGSnapshotContext,
	datasetID string,
	expectedVersion int,
	resource WorkflowRAGEvaluationDatasetResource,
	audit WorkflowRAGEvaluationAudit,
) error {
	resourcePayload, err := encodeWorkflowRAGEvaluationResource(resource, ctx)
	if err != nil || repository == nil || repository.database == nil || resource.DatasetID != datasetID ||
		resource.LifecycleState != workflowRAGEvaluationArchived || resource.ArchivedAt == nil ||
		audit.EventKind != "dataset_archived" || !workflowRAGEvaluationAuditResourceMatch(audit, resource) {
		return errWorkflowRAGEvaluationContract
	}
	auditPayload, err := encodeWorkflowRAGEvaluationAudit(audit, ctx)
	if err != nil {
		return err
	}
	updatedAt, err := workflowRAGEvaluationUnixNano(resource.UpdatedAt)
	if err != nil {
		return err
	}
	archivedAt, err := workflowRAGEvaluationUnixNano(*resource.ArchivedAt)
	if err != nil {
		return err
	}
	tx, err := repository.database.BeginTx(workflowRAGDatabaseContext(ctx), nil)
	if err != nil {
		return errWorkflowRAGEvaluationStore
	}
	defer func() { _ = tx.Rollback() }()
	result, err := tx.ExecContext(workflowRAGDatabaseContext(ctx), `UPDATE workflow_rag_evaluation_dataset_resources SET
	 lifecycle_state='archived',updated_at_unix_nano=?,archived_at_unix_nano=?,sanitized_resource_payload=?
	 WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND dataset_id=? AND latest_version=? AND lifecycle_state='active'`,
		updatedAt, archivedAt, string(resourcePayload), ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, datasetID, expectedVersion)
	if err != nil {
		return errWorkflowRAGEvaluationStore
	}
	if affected, rowsErr := result.RowsAffected(); rowsErr != nil || affected != 1 {
		return repository.mutationConflict(ctx, tx, datasetID, expectedVersion)
	}
	if err = insertSQLiteWorkflowRAGEvaluationAudit(ctx, tx, audit, auditPayload); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return errWorkflowRAGEvaluationStore
	}
	return nil
}

func (repository *sqliteWorkflowRAGEvaluationDatasetRepository) CreateReview(ctx WorkflowRAGSnapshotContext, review WorkflowRAGCandidateSnapshotReview, audit WorkflowRAGEvaluationAudit) error {
	reviewPayload, err := encodeWorkflowRAGCandidateReview(review, ctx)
	if err != nil || repository == nil || repository.database == nil || !workflowRAGEvaluationAuditReviewMatch(audit, review) {
		return errWorkflowRAGEvaluationContract
	}
	auditPayload, err := encodeWorkflowRAGEvaluationAudit(audit, ctx)
	if err != nil {
		return err
	}
	createdAt, err := workflowRAGEvaluationUnixNano(review.CreatedAt)
	if err != nil {
		return err
	}
	tx, err := repository.database.BeginTx(workflowRAGDatabaseContext(ctx), nil)
	if err != nil {
		return errWorkflowRAGEvaluationStore
	}
	defer func() { _ = tx.Rollback() }()
	version, err := readSQLiteWorkflowRAGEvaluationVersion(ctx, tx, review.Dataset.DatasetID, review.Dataset.DatasetVersion)
	if errors.Is(err, sql.ErrNoRows) {
		return errWorkflowRAGEvaluationNotFound
	}
	if err != nil {
		return evaluationSQLiteStoreError(err)
	}
	if version.Dataset.DatasetDigest != review.Dataset.DatasetDigest || version.Dataset.Snapshot != review.BaselineSnapshot ||
		version.Dataset.ContentClassification != audit.ContentClassification || !workflowRAGEvaluationAuditVersionMatch(audit, version) {
		return errWorkflowRAGEvaluationContract
	}
	var reviewCount int
	if err = tx.QueryRowContext(workflowRAGDatabaseContext(ctx), `SELECT count(*) FROM workflow_rag_candidate_snapshot_reviews
	 WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND dataset_id=?`, ctx.TenantRef, ctx.WorkspaceID,
		ctx.ApplicationID, review.Dataset.DatasetID).Scan(&reviewCount); err != nil {
		return errWorkflowRAGEvaluationStore
	}
	if reviewCount >= workflowRAGEvaluationMaxReviews {
		return errWorkflowRAGEvaluationStore
	}
	_, err = tx.ExecContext(workflowRAGDatabaseContext(ctx), `INSERT INTO workflow_rag_candidate_snapshot_reviews (
	 tenant_ref,workspace_id,application_id,dataset_id,dataset_version,dataset_digest,review_id,created_at_unix_nano,review_payload
	) VALUES (?,?,?,?,?,?,?,?,?)`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, review.Dataset.DatasetID,
		review.Dataset.DatasetVersion, review.Dataset.DatasetDigest, review.ReviewID, createdAt, string(reviewPayload))
	if err != nil {
		return errWorkflowRAGEvaluationStore
	}
	if err = insertSQLiteWorkflowRAGEvaluationAudit(ctx, tx, audit, auditPayload); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return errWorkflowRAGEvaluationStore
	}
	return nil
}

func (repository *sqliteWorkflowRAGEvaluationDatasetRepository) ReadReview(ctx WorkflowRAGSnapshotContext, datasetID, reviewID string) (WorkflowRAGCandidateSnapshotReview, error) {
	if _, err := repository.readResource(ctx, datasetID); err != nil {
		return WorkflowRAGCandidateSnapshotReview{}, err
	}
	var payload string
	err := repository.database.QueryRowContext(workflowRAGDatabaseContext(ctx), `SELECT review_payload
	 FROM workflow_rag_candidate_snapshot_reviews WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND dataset_id=? AND review_id=?`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, datasetID, reviewID).Scan(&payload)
	if errors.Is(err, sql.ErrNoRows) {
		return WorkflowRAGCandidateSnapshotReview{}, errWorkflowRAGEvaluationNotFound
	}
	if err != nil {
		return WorkflowRAGCandidateSnapshotReview{}, errWorkflowRAGEvaluationStore
	}
	review, err := decodeWorkflowRAGCandidateReview([]byte(payload), ctx)
	if err != nil || review.Dataset.DatasetID != datasetID || review.ReviewID != reviewID {
		return WorkflowRAGCandidateSnapshotReview{}, errWorkflowRAGEvaluationContract
	}
	return review, nil
}

func (repository *sqliteWorkflowRAGEvaluationDatasetRepository) ListReviews(ctx WorkflowRAGSnapshotContext, datasetID string, query workflowRAGCandidateReviewListQuery) ([]WorkflowRAGCandidateSnapshotReview, error) {
	if _, err := repository.readResource(ctx, datasetID); err != nil {
		return nil, err
	}
	var beforeAt int64
	var err error
	if query.BeforeAt != "" {
		beforeAt, err = workflowRAGEvaluationUnixNano(query.BeforeAt)
		if err != nil {
			return nil, err
		}
	}
	rows, err := repository.database.QueryContext(workflowRAGDatabaseContext(ctx), `SELECT review_payload
	 FROM workflow_rag_candidate_snapshot_reviews
	 WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND dataset_id=?
	 AND (?=0 OR created_at_unix_nano<? OR (created_at_unix_nano=? AND review_id<?))
	 ORDER BY created_at_unix_nano DESC,review_id DESC LIMIT ?`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, datasetID,
		beforeAt, beforeAt, beforeAt, query.BeforeReviewID, query.Limit)
	if err != nil {
		return nil, errWorkflowRAGEvaluationStore
	}
	defer rows.Close()
	reviews := make([]WorkflowRAGCandidateSnapshotReview, 0)
	for rows.Next() {
		var payload string
		if err = rows.Scan(&payload); err != nil {
			return nil, errWorkflowRAGEvaluationStore
		}
		review, decodeErr := decodeWorkflowRAGCandidateReview([]byte(payload), ctx)
		if decodeErr != nil || review.Dataset.DatasetID != datasetID {
			return nil, errWorkflowRAGEvaluationContract
		}
		reviews = append(reviews, review)
	}
	if rows.Err() != nil {
		return nil, errWorkflowRAGEvaluationStore
	}
	return reviews, nil
}

type sqliteWorkflowRAGEvaluationReader interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

type sqliteWorkflowRAGEvaluationScanner interface{ Scan(...any) error }

func insertSQLiteWorkflowRAGEvaluationVersion(ctx WorkflowRAGSnapshotContext, tx *sql.Tx, version WorkflowRAGEvaluationDatasetVersion, payload []byte) error {
	createdAt, err := workflowRAGEvaluationUnixNano(version.CreatedAt)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(workflowRAGDatabaseContext(ctx), `INSERT INTO workflow_rag_evaluation_dataset_versions (
	 tenant_ref,workspace_id,application_id,dataset_id,dataset_version,dataset_digest,created_at_unix_nano,dataset_version_payload
	) VALUES (?,?,?,?,?,?,?,?)`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, version.Dataset.DatasetID,
		version.Dataset.DatasetVersion, version.Dataset.DatasetDigest, createdAt, string(payload))
	if err != nil {
		return errWorkflowRAGEvaluationStore
	}
	return nil
}

func insertSQLiteWorkflowRAGEvaluationAudit(ctx WorkflowRAGSnapshotContext, tx *sql.Tx, audit WorkflowRAGEvaluationAudit, payload []byte) error {
	occurredAt, err := workflowRAGEvaluationUnixNano(audit.OccurredAt)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(workflowRAGDatabaseContext(ctx), `INSERT INTO workflow_rag_evaluation_audits (
	 tenant_ref,workspace_id,application_id,dataset_id,dataset_version,event_id,event_kind,occurred_at_unix_nano,audit_payload
	) VALUES (?,?,?,?,?,?,?,?,?)`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, audit.DatasetID, audit.DatasetVersion,
		audit.EventID, audit.EventKind, occurredAt, string(payload))
	if err != nil {
		return errWorkflowRAGEvaluationStore
	}
	return nil
}

func readSQLiteWorkflowRAGEvaluationVersion(ctx WorkflowRAGSnapshotContext, reader sqliteWorkflowRAGEvaluationReader, datasetID string, version int) (WorkflowRAGEvaluationDatasetVersion, error) {
	var payload string
	err := reader.QueryRowContext(workflowRAGDatabaseContext(ctx), `SELECT dataset_version_payload
	 FROM workflow_rag_evaluation_dataset_versions
	 WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND dataset_id=? AND dataset_version=?`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, datasetID, version).Scan(&payload)
	if err != nil {
		return WorkflowRAGEvaluationDatasetVersion{}, err
	}
	decoded, err := decodeWorkflowRAGEvaluationVersion([]byte(payload), ctx)
	if err != nil || decoded.Dataset.DatasetID != datasetID || decoded.Dataset.DatasetVersion != version {
		return WorkflowRAGEvaluationDatasetVersion{}, errWorkflowRAGEvaluationContract
	}
	return decoded, nil
}

func scanSQLiteWorkflowRAGEvaluationResource(scanner sqliteWorkflowRAGEvaluationScanner, ctx WorkflowRAGSnapshotContext) (WorkflowRAGEvaluationDatasetResource, error) {
	var payload string
	if err := scanner.Scan(&payload); err != nil {
		return WorkflowRAGEvaluationDatasetResource{}, err
	}
	return decodeWorkflowRAGEvaluationResource([]byte(payload), ctx)
}

func (repository *sqliteWorkflowRAGEvaluationDatasetRepository) readResource(ctx WorkflowRAGSnapshotContext, datasetID string) (WorkflowRAGEvaluationDatasetResource, error) {
	if repository == nil || repository.database == nil {
		return WorkflowRAGEvaluationDatasetResource{}, errWorkflowRAGEvaluationStore
	}
	resource, err := scanSQLiteWorkflowRAGEvaluationResource(repository.database.QueryRowContext(workflowRAGDatabaseContext(ctx), `SELECT sanitized_resource_payload
	 FROM workflow_rag_evaluation_dataset_resources WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND dataset_id=?`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, datasetID), ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return WorkflowRAGEvaluationDatasetResource{}, repository.missingScopeResult(ctx, datasetID)
	}
	if err != nil {
		return WorkflowRAGEvaluationDatasetResource{}, evaluationSQLiteStoreError(err)
	}
	return resource, nil
}

func (repository *sqliteWorkflowRAGEvaluationDatasetRepository) missingScopeResult(ctx WorkflowRAGSnapshotContext, datasetID string) error {
	var count int
	if err := repository.database.QueryRowContext(workflowRAGDatabaseContext(ctx), `SELECT count(*) FROM workflow_rag_evaluation_dataset_resources WHERE dataset_id=?`, datasetID).Scan(&count); err != nil {
		return errWorkflowRAGEvaluationStore
	}
	if count > 0 {
		return errWorkflowRAGEvaluationScopeDenied
	}
	return errWorkflowRAGEvaluationNotFound
}

func (repository *sqliteWorkflowRAGEvaluationDatasetRepository) mutationConflict(ctx WorkflowRAGSnapshotContext, tx *sql.Tx, datasetID string, expectedVersion int) error {
	resource, err := scanSQLiteWorkflowRAGEvaluationResource(tx.QueryRowContext(workflowRAGDatabaseContext(ctx), `SELECT sanitized_resource_payload
	 FROM workflow_rag_evaluation_dataset_resources WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND dataset_id=?`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, datasetID), ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return errWorkflowRAGEvaluationNotFound
	}
	if err != nil {
		return evaluationSQLiteStoreError(err)
	}
	if resource.LatestVersion != expectedVersion {
		return workflowRAGEvaluationConflictError{CurrentVersion: resource.LatestVersion, CurrentState: resource.LifecycleState}
	}
	if resource.LifecycleState != workflowRAGEvaluationActive {
		return errWorkflowRAGEvaluationArchived
	}
	return errWorkflowRAGEvaluationStore
}

func evaluationSQLiteStoreError(err error) error {
	if errors.Is(err, errWorkflowRAGEvaluationContract) {
		return errWorkflowRAGEvaluationContract
	}
	return errWorkflowRAGEvaluationStore
}

var _ workflowRAGEvaluationDatasetRepository = (*sqliteWorkflowRAGEvaluationDatasetRepository)(nil)
