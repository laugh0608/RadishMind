package httpapi

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresWorkflowRAGEvaluationDatasetRepository struct{ pool *pgxpool.Pool }

func newPostgresWorkflowRAGEvaluationDatasetRepository(pool *pgxpool.Pool) *postgresWorkflowRAGEvaluationDatasetRepository {
	return &postgresWorkflowRAGEvaluationDatasetRepository{pool: pool}
}

func (repository *postgresWorkflowRAGEvaluationDatasetRepository) Create(
	ctx WorkflowRAGSnapshotContext,
	resource WorkflowRAGEvaluationDatasetResource,
	version WorkflowRAGEvaluationDatasetVersion,
	audit WorkflowRAGEvaluationAudit,
) error {
	resourcePayload, versionPayload, auditPayload, err := encodeWorkflowRAGEvaluationCreatePayloads(ctx, resource, version, audit)
	if err != nil || repository == nil || repository.pool == nil || resource.LatestVersion != 1 || audit.EventKind != "dataset_created" {
		return errWorkflowRAGEvaluationContract
	}
	createdAt, _ := time.Parse(time.RFC3339Nano, resource.CreatedAt)
	tx, err := repository.pool.Begin(workflowRAGDatabaseContext(ctx))
	if err != nil {
		return errWorkflowRAGEvaluationStore
	}
	defer func() { _ = tx.Rollback(workflowRAGDatabaseContext(ctx)) }()
	result, err := tx.Exec(workflowRAGDatabaseContext(ctx), `INSERT INTO workflow_rag_evaluation_dataset_resources (
	 tenant_ref,workspace_id,application_id,dataset_id,dataset_key,lifecycle_state,latest_version,latest_digest,
	 created_at,updated_at,archived_at,sanitized_resource_payload
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$9,NULL,$10) ON CONFLICT DO NOTHING`, ctx.TenantRef, ctx.WorkspaceID,
		ctx.ApplicationID, resource.DatasetID, resource.DatasetKey, resource.LifecycleState, resource.LatestVersion,
		resource.LatestDigest, createdAt, resourcePayload)
	if err != nil {
		return errWorkflowRAGEvaluationStore
	}
	if result.RowsAffected() != 1 {
		return workflowRAGEvaluationConflictError{}
	}
	if err = insertPostgresWorkflowRAGEvaluationVersion(ctx, tx, version, versionPayload); err != nil {
		return err
	}
	if err = insertPostgresWorkflowRAGEvaluationAudit(ctx, tx, audit, auditPayload); err != nil {
		return err
	}
	if err = tx.Commit(workflowRAGDatabaseContext(ctx)); err != nil {
		return errWorkflowRAGEvaluationStore
	}
	return nil
}

func (repository *postgresWorkflowRAGEvaluationDatasetRepository) List(ctx WorkflowRAGSnapshotContext, query workflowRAGEvaluationListQuery) ([]WorkflowRAGEvaluationDatasetResource, error) {
	if repository == nil || repository.pool == nil {
		return nil, errWorkflowRAGEvaluationStore
	}
	rows, err := repository.pool.Query(workflowRAGDatabaseContext(ctx), `SELECT sanitized_resource_payload
	 FROM workflow_rag_evaluation_dataset_resources
	 WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND lifecycle_state=$4 AND ($5='' OR dataset_key>$5)
	 ORDER BY dataset_key ASC LIMIT $6`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, query.LifecycleState,
		query.AfterDatasetKey, query.Limit)
	if err != nil {
		return nil, errWorkflowRAGEvaluationStore
	}
	defer rows.Close()
	resources := make([]WorkflowRAGEvaluationDatasetResource, 0)
	for rows.Next() {
		resource, scanErr := scanPostgresWorkflowRAGEvaluationResource(rows, ctx)
		if scanErr != nil {
			return nil, evaluationPostgresStoreError(scanErr)
		}
		resources = append(resources, resource)
	}
	if rows.Err() != nil {
		return nil, errWorkflowRAGEvaluationStore
	}
	return resources, nil
}

func (repository *postgresWorkflowRAGEvaluationDatasetRepository) ReadLatest(ctx WorkflowRAGSnapshotContext, datasetID string) (WorkflowRAGEvaluationDatasetResource, WorkflowRAGEvaluationDatasetVersion, error) {
	resource, err := repository.readResource(ctx, datasetID)
	if err != nil {
		return WorkflowRAGEvaluationDatasetResource{}, WorkflowRAGEvaluationDatasetVersion{}, err
	}
	version, err := readPostgresWorkflowRAGEvaluationVersion(ctx, repository.pool, datasetID, resource.LatestVersion)
	if err != nil {
		return WorkflowRAGEvaluationDatasetResource{}, WorkflowRAGEvaluationDatasetVersion{}, evaluationPostgresStoreError(err)
	}
	if !workflowRAGEvaluationResourceVersionMatch(resource, version) {
		return WorkflowRAGEvaluationDatasetResource{}, WorkflowRAGEvaluationDatasetVersion{}, errWorkflowRAGEvaluationContract
	}
	return resource, version, nil
}

func (repository *postgresWorkflowRAGEvaluationDatasetRepository) ReadVersion(ctx WorkflowRAGSnapshotContext, datasetID string, datasetVersion int) (WorkflowRAGEvaluationDatasetResource, WorkflowRAGEvaluationDatasetVersion, error) {
	resource, err := repository.readResource(ctx, datasetID)
	if err != nil {
		return WorkflowRAGEvaluationDatasetResource{}, WorkflowRAGEvaluationDatasetVersion{}, err
	}
	version, err := readPostgresWorkflowRAGEvaluationVersion(ctx, repository.pool, datasetID, datasetVersion)
	if errors.Is(err, pgx.ErrNoRows) {
		return WorkflowRAGEvaluationDatasetResource{}, WorkflowRAGEvaluationDatasetVersion{}, errWorkflowRAGEvaluationNotFound
	}
	if err != nil {
		return WorkflowRAGEvaluationDatasetResource{}, WorkflowRAGEvaluationDatasetVersion{}, evaluationPostgresStoreError(err)
	}
	return resource, version, nil
}

func (repository *postgresWorkflowRAGEvaluationDatasetRepository) CreateVersion(
	ctx WorkflowRAGSnapshotContext,
	datasetID string,
	expectedVersion int,
	resource WorkflowRAGEvaluationDatasetResource,
	version WorkflowRAGEvaluationDatasetVersion,
	audit WorkflowRAGEvaluationAudit,
) error {
	resourcePayload, versionPayload, auditPayload, err := encodeWorkflowRAGEvaluationCreatePayloads(ctx, resource, version, audit)
	if err != nil || repository == nil || repository.pool == nil || resource.DatasetID != datasetID ||
		version.Dataset.DatasetVersion != expectedVersion+1 || audit.EventKind != "dataset_versioned" {
		return errWorkflowRAGEvaluationContract
	}
	updatedAt, _ := time.Parse(time.RFC3339Nano, resource.UpdatedAt)
	tx, err := repository.pool.Begin(workflowRAGDatabaseContext(ctx))
	if err != nil {
		return errWorkflowRAGEvaluationStore
	}
	defer func() { _ = tx.Rollback(workflowRAGDatabaseContext(ctx)) }()
	result, err := tx.Exec(workflowRAGDatabaseContext(ctx), `UPDATE workflow_rag_evaluation_dataset_resources SET
	 lifecycle_state=$1,latest_version=$2,latest_digest=$3,updated_at=$4,sanitized_resource_payload=$5
	 WHERE tenant_ref=$6 AND workspace_id=$7 AND application_id=$8 AND dataset_id=$9 AND latest_version=$10 AND lifecycle_state='active'`,
		resource.LifecycleState, resource.LatestVersion, resource.LatestDigest, updatedAt, resourcePayload,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, datasetID, expectedVersion)
	if err != nil {
		return errWorkflowRAGEvaluationStore
	}
	if result.RowsAffected() != 1 {
		return repository.mutationConflict(ctx, tx, datasetID, expectedVersion)
	}
	if err = insertPostgresWorkflowRAGEvaluationVersion(ctx, tx, version, versionPayload); err != nil {
		return err
	}
	if err = insertPostgresWorkflowRAGEvaluationAudit(ctx, tx, audit, auditPayload); err != nil {
		return err
	}
	if err = tx.Commit(workflowRAGDatabaseContext(ctx)); err != nil {
		return errWorkflowRAGEvaluationStore
	}
	return nil
}

func (repository *postgresWorkflowRAGEvaluationDatasetRepository) Archive(
	ctx WorkflowRAGSnapshotContext,
	datasetID string,
	expectedVersion int,
	resource WorkflowRAGEvaluationDatasetResource,
	audit WorkflowRAGEvaluationAudit,
) error {
	resourcePayload, err := encodeWorkflowRAGEvaluationResource(resource, ctx)
	if err != nil || repository == nil || repository.pool == nil || resource.DatasetID != datasetID ||
		resource.LifecycleState != workflowRAGEvaluationArchived || resource.ArchivedAt == nil ||
		audit.EventKind != "dataset_archived" || !workflowRAGEvaluationAuditResourceMatch(audit, resource) {
		return errWorkflowRAGEvaluationContract
	}
	auditPayload, err := encodeWorkflowRAGEvaluationAudit(audit, ctx)
	if err != nil {
		return err
	}
	updatedAt, _ := time.Parse(time.RFC3339Nano, resource.UpdatedAt)
	archivedAt, _ := time.Parse(time.RFC3339Nano, *resource.ArchivedAt)
	tx, err := repository.pool.Begin(workflowRAGDatabaseContext(ctx))
	if err != nil {
		return errWorkflowRAGEvaluationStore
	}
	defer func() { _ = tx.Rollback(workflowRAGDatabaseContext(ctx)) }()
	result, err := tx.Exec(workflowRAGDatabaseContext(ctx), `UPDATE workflow_rag_evaluation_dataset_resources SET
	 lifecycle_state='archived',updated_at=$1,archived_at=$2,sanitized_resource_payload=$3
	 WHERE tenant_ref=$4 AND workspace_id=$5 AND application_id=$6 AND dataset_id=$7 AND latest_version=$8 AND lifecycle_state='active'`,
		updatedAt, archivedAt, resourcePayload, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, datasetID, expectedVersion)
	if err != nil {
		return errWorkflowRAGEvaluationStore
	}
	if result.RowsAffected() != 1 {
		return repository.mutationConflict(ctx, tx, datasetID, expectedVersion)
	}
	if err = insertPostgresWorkflowRAGEvaluationAudit(ctx, tx, audit, auditPayload); err != nil {
		return err
	}
	if err = tx.Commit(workflowRAGDatabaseContext(ctx)); err != nil {
		return errWorkflowRAGEvaluationStore
	}
	return nil
}

func (repository *postgresWorkflowRAGEvaluationDatasetRepository) CreateReview(ctx WorkflowRAGSnapshotContext, review WorkflowRAGCandidateSnapshotReview, audit WorkflowRAGEvaluationAudit) error {
	reviewPayload, err := encodeWorkflowRAGCandidateReview(review, ctx)
	if err != nil || repository == nil || repository.pool == nil || !workflowRAGEvaluationAuditReviewMatch(audit, review) {
		return errWorkflowRAGEvaluationContract
	}
	auditPayload, err := encodeWorkflowRAGEvaluationAudit(audit, ctx)
	if err != nil {
		return err
	}
	createdAt, _ := time.Parse(time.RFC3339Nano, review.CreatedAt)
	tx, err := repository.pool.Begin(workflowRAGDatabaseContext(ctx))
	if err != nil {
		return errWorkflowRAGEvaluationStore
	}
	defer func() { _ = tx.Rollback(workflowRAGDatabaseContext(ctx)) }()
	version, err := readPostgresWorkflowRAGEvaluationVersion(ctx, tx, review.Dataset.DatasetID, review.Dataset.DatasetVersion)
	if errors.Is(err, pgx.ErrNoRows) {
		return errWorkflowRAGEvaluationNotFound
	}
	if err != nil {
		return evaluationPostgresStoreError(err)
	}
	if version.Dataset.DatasetDigest != review.Dataset.DatasetDigest || version.Dataset.Snapshot != review.BaselineSnapshot ||
		version.Dataset.ContentClassification != audit.ContentClassification || !workflowRAGEvaluationAuditVersionMatch(audit, version) {
		return errWorkflowRAGEvaluationContract
	}
	var reviewCount int
	if err = tx.QueryRow(workflowRAGDatabaseContext(ctx), `SELECT count(*) FROM workflow_rag_candidate_snapshot_reviews
	 WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND dataset_id=$4`, ctx.TenantRef, ctx.WorkspaceID,
		ctx.ApplicationID, review.Dataset.DatasetID).Scan(&reviewCount); err != nil {
		return errWorkflowRAGEvaluationStore
	}
	if reviewCount >= workflowRAGEvaluationMaxReviews {
		return errWorkflowRAGEvaluationStore
	}
	_, err = tx.Exec(workflowRAGDatabaseContext(ctx), `INSERT INTO workflow_rag_candidate_snapshot_reviews (
	 tenant_ref,workspace_id,application_id,dataset_id,dataset_version,dataset_digest,review_id,created_at,review_payload
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, review.Dataset.DatasetID,
		review.Dataset.DatasetVersion, review.Dataset.DatasetDigest, review.ReviewID, createdAt, reviewPayload)
	if err != nil {
		return errWorkflowRAGEvaluationStore
	}
	if err = insertPostgresWorkflowRAGEvaluationAudit(ctx, tx, audit, auditPayload); err != nil {
		return err
	}
	if err = tx.Commit(workflowRAGDatabaseContext(ctx)); err != nil {
		return errWorkflowRAGEvaluationStore
	}
	return nil
}

func (repository *postgresWorkflowRAGEvaluationDatasetRepository) ReadReview(ctx WorkflowRAGSnapshotContext, datasetID, reviewID string) (WorkflowRAGCandidateSnapshotReview, error) {
	if _, err := repository.readResource(ctx, datasetID); err != nil {
		return WorkflowRAGCandidateSnapshotReview{}, err
	}
	var payload []byte
	err := repository.pool.QueryRow(workflowRAGDatabaseContext(ctx), `SELECT review_payload
	 FROM workflow_rag_candidate_snapshot_reviews WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND dataset_id=$4 AND review_id=$5`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, datasetID, reviewID).Scan(&payload)
	if errors.Is(err, pgx.ErrNoRows) {
		return WorkflowRAGCandidateSnapshotReview{}, errWorkflowRAGEvaluationNotFound
	}
	if err != nil {
		return WorkflowRAGCandidateSnapshotReview{}, errWorkflowRAGEvaluationStore
	}
	review, err := decodeWorkflowRAGCandidateReview(payload, ctx)
	if err != nil || review.Dataset.DatasetID != datasetID || review.ReviewID != reviewID {
		return WorkflowRAGCandidateSnapshotReview{}, errWorkflowRAGEvaluationContract
	}
	return review, nil
}

func (repository *postgresWorkflowRAGEvaluationDatasetRepository) ListReviews(ctx WorkflowRAGSnapshotContext, datasetID string, query workflowRAGCandidateReviewListQuery) ([]WorkflowRAGCandidateSnapshotReview, error) {
	if _, err := repository.readResource(ctx, datasetID); err != nil {
		return nil, err
	}
	var beforeAt any
	if query.BeforeAt != "" {
		parsed, err := time.Parse(time.RFC3339Nano, query.BeforeAt)
		if err != nil {
			return nil, errWorkflowRAGEvaluationContract
		}
		beforeAt = parsed
	}
	rows, err := repository.pool.Query(workflowRAGDatabaseContext(ctx), `SELECT review_payload
	 FROM workflow_rag_candidate_snapshot_reviews
	 WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND dataset_id=$4
	 AND ($5::timestamptz IS NULL OR created_at<$5 OR (created_at=$5 AND review_id<$6))
	 ORDER BY created_at DESC,review_id DESC LIMIT $7`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, datasetID,
		beforeAt, query.BeforeReviewID, query.Limit)
	if err != nil {
		return nil, errWorkflowRAGEvaluationStore
	}
	defer rows.Close()
	reviews := make([]WorkflowRAGCandidateSnapshotReview, 0)
	for rows.Next() {
		var payload []byte
		if err = rows.Scan(&payload); err != nil {
			return nil, errWorkflowRAGEvaluationStore
		}
		review, decodeErr := decodeWorkflowRAGCandidateReview(payload, ctx)
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

type postgresWorkflowRAGEvaluationReader interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

type postgresWorkflowRAGEvaluationScanner interface{ Scan(...any) error }

func insertPostgresWorkflowRAGEvaluationVersion(ctx WorkflowRAGSnapshotContext, tx pgx.Tx, version WorkflowRAGEvaluationDatasetVersion, payload []byte) error {
	createdAt, _ := time.Parse(time.RFC3339Nano, version.CreatedAt)
	_, err := tx.Exec(workflowRAGDatabaseContext(ctx), `INSERT INTO workflow_rag_evaluation_dataset_versions (
	 tenant_ref,workspace_id,application_id,dataset_id,dataset_version,dataset_digest,created_at,dataset_version_payload
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, version.Dataset.DatasetID,
		version.Dataset.DatasetVersion, version.Dataset.DatasetDigest, createdAt, payload)
	if err != nil {
		return errWorkflowRAGEvaluationStore
	}
	return nil
}

func insertPostgresWorkflowRAGEvaluationAudit(ctx WorkflowRAGSnapshotContext, tx pgx.Tx, audit WorkflowRAGEvaluationAudit, payload []byte) error {
	occurredAt, _ := time.Parse(time.RFC3339Nano, audit.OccurredAt)
	_, err := tx.Exec(workflowRAGDatabaseContext(ctx), `INSERT INTO workflow_rag_evaluation_audits (
	 tenant_ref,workspace_id,application_id,dataset_id,dataset_version,event_id,event_kind,occurred_at,audit_payload
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, audit.DatasetID,
		audit.DatasetVersion, audit.EventID, audit.EventKind, occurredAt, payload)
	if err != nil {
		return errWorkflowRAGEvaluationStore
	}
	return nil
}

func readPostgresWorkflowRAGEvaluationVersion(ctx WorkflowRAGSnapshotContext, reader postgresWorkflowRAGEvaluationReader, datasetID string, version int) (WorkflowRAGEvaluationDatasetVersion, error) {
	var payload []byte
	err := reader.QueryRow(workflowRAGDatabaseContext(ctx), `SELECT dataset_version_payload
	 FROM workflow_rag_evaluation_dataset_versions
	 WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND dataset_id=$4 AND dataset_version=$5`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, datasetID, version).Scan(&payload)
	if err != nil {
		return WorkflowRAGEvaluationDatasetVersion{}, err
	}
	decoded, err := decodeWorkflowRAGEvaluationVersion(payload, ctx)
	if err != nil || decoded.Dataset.DatasetID != datasetID || decoded.Dataset.DatasetVersion != version {
		return WorkflowRAGEvaluationDatasetVersion{}, errWorkflowRAGEvaluationContract
	}
	return decoded, nil
}

func scanPostgresWorkflowRAGEvaluationResource(scanner postgresWorkflowRAGEvaluationScanner, ctx WorkflowRAGSnapshotContext) (WorkflowRAGEvaluationDatasetResource, error) {
	var payload []byte
	if err := scanner.Scan(&payload); err != nil {
		return WorkflowRAGEvaluationDatasetResource{}, err
	}
	return decodeWorkflowRAGEvaluationResource(payload, ctx)
}

func (repository *postgresWorkflowRAGEvaluationDatasetRepository) readResource(ctx WorkflowRAGSnapshotContext, datasetID string) (WorkflowRAGEvaluationDatasetResource, error) {
	if repository == nil || repository.pool == nil {
		return WorkflowRAGEvaluationDatasetResource{}, errWorkflowRAGEvaluationStore
	}
	resource, err := scanPostgresWorkflowRAGEvaluationResource(repository.pool.QueryRow(workflowRAGDatabaseContext(ctx), `SELECT sanitized_resource_payload
	 FROM workflow_rag_evaluation_dataset_resources WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND dataset_id=$4`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, datasetID), ctx)
	if errors.Is(err, pgx.ErrNoRows) {
		return WorkflowRAGEvaluationDatasetResource{}, repository.missingScopeResult(ctx, datasetID)
	}
	if err != nil {
		return WorkflowRAGEvaluationDatasetResource{}, evaluationPostgresStoreError(err)
	}
	return resource, nil
}

func (repository *postgresWorkflowRAGEvaluationDatasetRepository) missingScopeResult(ctx WorkflowRAGSnapshotContext, datasetID string) error {
	var count int
	if err := repository.pool.QueryRow(workflowRAGDatabaseContext(ctx), `SELECT count(*) FROM workflow_rag_evaluation_dataset_resources WHERE dataset_id=$1`, datasetID).Scan(&count); err != nil {
		return errWorkflowRAGEvaluationStore
	}
	if count > 0 {
		return errWorkflowRAGEvaluationScopeDenied
	}
	return errWorkflowRAGEvaluationNotFound
}

func (repository *postgresWorkflowRAGEvaluationDatasetRepository) mutationConflict(ctx WorkflowRAGSnapshotContext, tx pgx.Tx, datasetID string, expectedVersion int) error {
	resource, err := scanPostgresWorkflowRAGEvaluationResource(tx.QueryRow(workflowRAGDatabaseContext(ctx), `SELECT sanitized_resource_payload
	 FROM workflow_rag_evaluation_dataset_resources WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND dataset_id=$4`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, datasetID), ctx)
	if errors.Is(err, pgx.ErrNoRows) {
		return errWorkflowRAGEvaluationNotFound
	}
	if err != nil {
		return evaluationPostgresStoreError(err)
	}
	if resource.LatestVersion != expectedVersion {
		return workflowRAGEvaluationConflictError{CurrentVersion: resource.LatestVersion, CurrentState: resource.LifecycleState}
	}
	if resource.LifecycleState != workflowRAGEvaluationActive {
		return errWorkflowRAGEvaluationArchived
	}
	return errWorkflowRAGEvaluationStore
}

func evaluationPostgresStoreError(err error) error {
	if errors.Is(err, errWorkflowRAGEvaluationContract) {
		return errWorkflowRAGEvaluationContract
	}
	return errWorkflowRAGEvaluationStore
}

var _ workflowRAGEvaluationDatasetRepository = (*postgresWorkflowRAGEvaluationDatasetRepository)(nil)
