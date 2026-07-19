package httpapi

import (
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresWorkflowRunStore struct{ pool *pgxpool.Pool }

func newPostgresWorkflowRunStore(pool *pgxpool.Pool) *postgresWorkflowRunStore {
	return &postgresWorkflowRunStore{pool: pool}
}

func (store *postgresWorkflowRunStore) UpsertRun(runContext WorkflowRunContext, record *WorkflowRunRecord) error {
	if store == nil || store.pool == nil || runContext.RequestContext == nil {
		return errWorkflowRunStoreContract
	}
	if err := validateWorkflowRunStoreRecord(runContext, record); err != nil {
		return err
	}
	next := cloneWorkflowRunRecord(*record)
	next.RecordVersion++
	sourceKind, sourceID, sourceVersion, err := workflowRunStorageExecutionSource(next)
	if err != nil {
		return err
	}
	payload, startedAt, completedAt, err := encodeWorkflowRunStorageRecord(next)
	if err != nil {
		return err
	}
	var storedVersion int
	if record.RecordVersion == 0 {
		err = store.pool.QueryRow(runContext.RequestContext, `INSERT INTO workflow_run_records
 (tenant_ref,workspace_id,application_id,run_id,execution_source_kind,execution_source_id,execution_source_version,record_version,schema_version,run_status,started_at,completed_at,actor_ref,request_id,audit_ref,failure_code,failure_boundary,selected_provider,selected_model,sanitized_run_record)
 VALUES ($1,$2,$3,$4,$5,$6,$7,1,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)
 ON CONFLICT DO NOTHING RETURNING record_version`,
			runContext.TenantRef, runContext.WorkspaceID, runContext.ApplicationID, next.RunID, sourceKind,
			sourceID, sourceVersion, next.SchemaVersion, next.Status, startedAt, completedAt, next.ActorRef,
			next.RequestID, next.AuditRef, next.FailureCode, workflowRunRecordFailureBoundary(next), next.SelectedProvider,
			next.SelectedModel, payload).Scan(&storedVersion)
	} else {
		err = store.pool.QueryRow(runContext.RequestContext, `UPDATE workflow_run_records SET
			execution_source_kind=$1,execution_source_id=$2,execution_source_version=$3,
			record_version=record_version+1,schema_version=$4,run_status=$5,
			completed_at=$6,actor_ref=$7,request_id=$8,audit_ref=$9,failure_code=$10,failure_boundary=$11,
			selected_provider=$12,selected_model=$13,sanitized_run_record=$14
 WHERE tenant_ref=$15 AND workspace_id=$16 AND application_id=$17 AND run_id=$18
  AND record_version=$19 AND run_status='running' RETURNING record_version`,
			sourceKind, sourceID, sourceVersion, next.SchemaVersion, next.Status, completedAt, next.ActorRef,
			next.RequestID, next.AuditRef, next.FailureCode, workflowRunRecordFailureBoundary(next), next.SelectedProvider,
			next.SelectedModel, payload, runContext.TenantRef, runContext.WorkspaceID,
			runContext.ApplicationID, next.RunID, record.RecordVersion).Scan(&storedVersion)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return errWorkflowRunStoreConflict
	}
	if err != nil {
		return errWorkflowRunStoreUnavailable
	}
	record.RecordVersion = storedVersion
	return nil
}

func (store *postgresWorkflowRunStore) ReadRun(runContext WorkflowRunContext, runID string) (WorkflowRunRecord, bool, error) {
	if store == nil || store.pool == nil || runContext.RequestContext == nil {
		return WorkflowRunRecord{}, false, errWorkflowRunStoreContract
	}
	var sourceKind, sourceID string
	var sourceVersion int
	var payload []byte
	err := store.pool.QueryRow(runContext.RequestContext, `SELECT execution_source_kind,execution_source_id,execution_source_version,sanitized_run_record FROM workflow_run_records WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND run_id=$4`, runContext.TenantRef, runContext.WorkspaceID, runContext.ApplicationID, runID).Scan(&sourceKind, &sourceID, &sourceVersion, &payload)
	if errors.Is(err, pgx.ErrNoRows) {
		return WorkflowRunRecord{}, false, nil
	}
	if err != nil {
		return WorkflowRunRecord{}, false, errWorkflowRunStoreUnavailable
	}
	record, err := decodePostgresWorkflowRunStorageProjection(runContext, sourceKind, sourceID, sourceVersion, payload)
	if err != nil {
		return WorkflowRunRecord{}, false, err
	}
	return record, true, nil
}

func (store *postgresWorkflowRunStore) ListRuns(runContext WorkflowRunContext, filter WorkflowRunListFilter) (WorkflowRunListPage, error) {
	if store == nil || store.pool == nil || runContext.RequestContext == nil {
		return WorkflowRunListPage{}, errWorkflowRunStoreContract
	}
	limit := workflowRunStoreListLimit(filter.Limit)
	rows, err := store.pool.Query(runContext.RequestContext, `SELECT execution_source_kind,execution_source_id,execution_source_version,sanitized_run_record FROM workflow_run_records
 WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3
	AND ($4='' OR run_status=$4)
	AND ($5='' OR (execution_source_kind='workflow_draft' AND execution_source_id=$5))
	AND ($6='' OR execution_source_kind=$6) AND ($7='' OR execution_source_id=$7) AND ($8=0 OR execution_source_version=$8)
	AND ($9='' OR failure_code=$9) AND ($10='' OR failure_boundary=$10)
	AND ($11='' OR selected_provider=$11) AND ($12='' OR selected_model=$12)
	AND ($13::boolean IS NULL OR (run_status='running' AND started_at < $14)=$13)
	AND ($15::timestamptz IS NULL OR started_at >= $15) AND ($16::timestamptz IS NULL OR started_at <= $16)
	AND ($17::timestamptz IS NULL OR (started_at,run_id) < ($17,$18))
 ORDER BY started_at DESC, run_id DESC LIMIT $19`,
		runContext.TenantRef, runContext.WorkspaceID, runContext.ApplicationID, string(filter.Status), filter.DraftID,
		filter.ExecutionSourceKind, filter.ExecutionSourceID, filter.ExecutionSourceVersion,
		string(filter.FailureCode), string(filter.FailureBoundary), filter.Provider, filter.Model, filter.StaleRunning,
		time.Now().UTC().Add(-workflowExecutorDefaultMaxRuntime), filter.StartedFrom, filter.StartedTo,
		filter.BeforeTime, filter.BeforeRunID, limit+1)
	if err != nil {
		return WorkflowRunListPage{}, errWorkflowRunStoreUnavailable
	}
	defer rows.Close()
	records := make([]WorkflowRunRecord, 0, limit+1)
	for rows.Next() {
		var sourceKind, sourceID string
		var sourceVersion int
		var payload []byte
		if err = rows.Scan(&sourceKind, &sourceID, &sourceVersion, &payload); err != nil {
			return WorkflowRunListPage{}, errWorkflowRunStoreUnavailable
		}
		record, decodeErr := decodePostgresWorkflowRunStorageProjection(runContext, sourceKind, sourceID, sourceVersion, payload)
		if decodeErr != nil {
			return WorkflowRunListPage{}, decodeErr
		}
		records = append(records, record)
	}
	if rows.Err() != nil {
		return WorkflowRunListPage{}, errWorkflowRunStoreUnavailable
	}
	hasMore := len(records) > limit
	if hasMore {
		records = records[:limit]
	}
	return WorkflowRunListPage{Records: records, HasMore: hasMore}, nil
}

func decodePostgresWorkflowRunStorageProjection(runContext WorkflowRunContext, sourceKind, sourceID string, sourceVersion int, payload []byte) (WorkflowRunRecord, error) {
	record, err := decodeWorkflowRunStorageRecord(runContext, payload)
	if err != nil {
		return WorkflowRunRecord{}, err
	}
	decodedKind, decodedID, decodedVersion, err := workflowRunStorageExecutionSource(record)
	if err != nil || sourceKind != decodedKind || sourceID != decodedID || sourceVersion != decodedVersion {
		return WorkflowRunRecord{}, errWorkflowRunStoreContract
	}
	return record, nil
}

func workflowRunRecordFailureBoundary(record WorkflowRunRecord) WorkflowRunFailureBoundary {
	if record.Diagnostic == nil {
		return ""
	}
	if (record.SchemaVersion == workflowRunRecordRAGSchemaVersion || record.SchemaVersion == workflowRunRecordAppRAGSchemaVersion) && record.Diagnostic.FailureBoundary == "" {
		return WorkflowRunFailureBoundary("none")
	}
	return record.Diagnostic.FailureBoundary
}

var _ workflowRunStore = (*postgresWorkflowRunStore)(nil)
