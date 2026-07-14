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
	payload, startedAt, completedAt, err := encodeWorkflowRunStorageRecord(next)
	if err != nil {
		return err
	}
	var storedVersion int
	if record.RecordVersion == 0 {
		err = store.pool.QueryRow(runContext.RequestContext, `INSERT INTO workflow_run_records
 (tenant_ref,workspace_id,application_id,run_id,draft_id,draft_version,record_version,schema_version,run_status,started_at,completed_at,actor_ref,request_id,audit_ref,failure_code,failure_boundary,selected_provider,selected_model,sanitized_run_record)
 VALUES ($1,$2,$3,$4,$5,$6,1,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)
 ON CONFLICT DO NOTHING RETURNING record_version`,
			runContext.TenantRef, runContext.WorkspaceID, runContext.ApplicationID, next.RunID, next.DraftID,
			next.DraftVersion, next.SchemaVersion, next.Status, startedAt, completedAt, next.ActorRef,
			next.RequestID, next.AuditRef, next.FailureCode, workflowRunRecordFailureBoundary(next), next.SelectedProvider,
			next.SelectedModel, payload).Scan(&storedVersion)
	} else {
		err = store.pool.QueryRow(runContext.RequestContext, `UPDATE workflow_run_records SET
 draft_id=$1,draft_version=$2,record_version=record_version+1,schema_version=$3,run_status=$4,
	 completed_at=$5,actor_ref=$6,request_id=$7,audit_ref=$8,failure_code=$9,failure_boundary=$10,
	 selected_provider=$11,selected_model=$12,sanitized_run_record=$13
 WHERE tenant_ref=$14 AND workspace_id=$15 AND application_id=$16 AND run_id=$17
	 AND record_version=$18 AND run_status='running' RETURNING record_version`,
			next.DraftID, next.DraftVersion, next.SchemaVersion, next.Status, completedAt, next.ActorRef,
			next.RequestID, next.AuditRef, next.FailureCode, workflowRunRecordFailureBoundary(next), next.SelectedProvider,
			next.SelectedModel, payload, runContext.TenantRef, runContext.WorkspaceID,
			runContext.ApplicationID, next.RunID, record.RecordVersion).Scan(&storedVersion)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return errWorkflowRunStoreConflict
	}
	if err != nil {
		return err
	}
	record.RecordVersion = storedVersion
	return nil
}

func (store *postgresWorkflowRunStore) ReadRun(runContext WorkflowRunContext, runID string) (WorkflowRunRecord, bool, error) {
	if store == nil || store.pool == nil || runContext.RequestContext == nil {
		return WorkflowRunRecord{}, false, errWorkflowRunStoreContract
	}
	var payload []byte
	err := store.pool.QueryRow(runContext.RequestContext, `SELECT sanitized_run_record FROM workflow_run_records WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND run_id=$4`, runContext.TenantRef, runContext.WorkspaceID, runContext.ApplicationID, runID).Scan(&payload)
	if errors.Is(err, pgx.ErrNoRows) {
		return WorkflowRunRecord{}, false, nil
	}
	if err != nil {
		return WorkflowRunRecord{}, false, err
	}
	record, err := decodeWorkflowRunStorageRecord(runContext, payload)
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
	rows, err := store.pool.Query(runContext.RequestContext, `SELECT sanitized_run_record FROM workflow_run_records
 WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3
 AND ($4='' OR run_status=$4) AND ($5='' OR draft_id=$5)
	AND ($6='' OR failure_code=$6) AND ($7='' OR failure_boundary=$7)
	AND ($8='' OR selected_provider=$8) AND ($9='' OR selected_model=$9)
	AND ($10::boolean IS NULL OR (run_status='running' AND started_at < $11)=$10)
	AND ($12::timestamptz IS NULL OR started_at >= $12) AND ($13::timestamptz IS NULL OR started_at <= $13)
	AND ($14::timestamptz IS NULL OR (started_at,run_id) < ($14,$15))
 ORDER BY started_at DESC, run_id DESC LIMIT $16`,
		runContext.TenantRef, runContext.WorkspaceID, runContext.ApplicationID, string(filter.Status), filter.DraftID,
		string(filter.FailureCode), string(filter.FailureBoundary), filter.Provider, filter.Model, filter.StaleRunning,
		time.Now().UTC().Add(-workflowExecutorDefaultMaxRuntime), filter.StartedFrom, filter.StartedTo,
		filter.BeforeTime, filter.BeforeRunID, limit+1)
	if err != nil {
		return WorkflowRunListPage{}, err
	}
	defer rows.Close()
	records := make([]WorkflowRunRecord, 0, limit+1)
	for rows.Next() {
		var payload []byte
		if err = rows.Scan(&payload); err != nil {
			return WorkflowRunListPage{}, err
		}
		record, decodeErr := decodeWorkflowRunStorageRecord(runContext, payload)
		if decodeErr != nil {
			return WorkflowRunListPage{}, decodeErr
		}
		records = append(records, record)
	}
	if rows.Err() != nil {
		return WorkflowRunListPage{}, rows.Err()
	}
	hasMore := len(records) > limit
	if hasMore {
		records = records[:limit]
	}
	return WorkflowRunListPage{Records: records, HasMore: hasMore}, nil
}

func workflowRunRecordFailureBoundary(record WorkflowRunRecord) WorkflowRunFailureBoundary {
	if record.Diagnostic == nil {
		return ""
	}
	return record.Diagnostic.FailureBoundary
}

var _ workflowRunStore = (*postgresWorkflowRunStore)(nil)
