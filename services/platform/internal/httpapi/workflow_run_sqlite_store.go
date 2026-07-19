package httpapi

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	sqliteworkflowrunmigrations "radishmind.local/services/platform/migrations/sqlite/workflow_runs"
)

type sqliteWorkflowRunStore struct {
	database *sql.DB
}

func newSQLiteWorkflowRunStore(database *sql.DB) *sqliteWorkflowRunStore {
	return &sqliteWorkflowRunStore{database: database}
}

func (store *sqliteWorkflowRunStore) UpsertRun(
	runContext WorkflowRunContext,
	record *WorkflowRunRecord,
) error {
	if store == nil || store.database == nil || runContext.RequestContext == nil || record == nil {
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
	startedAtUnixNano, err := workflowRunUnixNano(startedAt)
	if err != nil {
		return errWorkflowRunStoreContract
	}
	completedAtUnixNano, err := optionalWorkflowRunUnixNano(completedAt)
	if err != nil {
		return errWorkflowRunStoreContract
	}

	var row sqliteWorkflowRunRow
	if record.RecordVersion == 0 {
		if record.Status != WorkflowRunStatusRunning {
			return errWorkflowRunStoreConflict
		}
		row = store.database.QueryRowContext(
			runContext.RequestContext,
			sqliteWorkflowRunInsertSQL,
			runContext.TenantRef,
			runContext.WorkspaceID,
			runContext.ApplicationID,
			next.RunID,
			sourceKind,
			sourceID,
			sourceVersion,
			next.RecordVersion,
			sqliteworkflowrunmigrations.RunRecordStoreSchemaVersion,
			next.SchemaVersion,
			next.Status,
			startedAtUnixNano,
			completedAtUnixNano,
			next.ActorRef,
			next.RequestID,
			next.AuditRef,
			next.FailureCode,
			workflowRunRecordFailureBoundary(next),
			next.SelectedProvider,
			next.SelectedModel,
			string(payload),
		)
	} else {
		row = store.database.QueryRowContext(
			runContext.RequestContext,
			sqliteWorkflowRunUpdateSQL,
			sourceKind,
			sourceID,
			sourceVersion,
			next.SchemaVersion,
			next.Status,
			completedAtUnixNano,
			next.ActorRef,
			next.RequestID,
			next.AuditRef,
			next.FailureCode,
			workflowRunRecordFailureBoundary(next),
			next.SelectedProvider,
			next.SelectedModel,
			string(payload),
			runContext.TenantRef,
			runContext.WorkspaceID,
			runContext.ApplicationID,
			next.RunID,
			record.RecordVersion,
			startedAtUnixNano,
		)
	}

	stored, err := scanSQLiteWorkflowRunRecord(runContext, row)
	if errors.Is(err, sql.ErrNoRows) {
		return errWorkflowRunStoreConflict
	}
	if err != nil {
		return normalizeSQLiteWorkflowRunStoreError(err)
	}
	*record = stored
	return nil
}

func (store *sqliteWorkflowRunStore) ReadRun(
	runContext WorkflowRunContext,
	runID string,
) (WorkflowRunRecord, bool, error) {
	if store == nil || store.database == nil || runContext.RequestContext == nil {
		return WorkflowRunRecord{}, false, errWorkflowRunStoreContract
	}
	record, err := scanSQLiteWorkflowRunRecord(runContext, store.database.QueryRowContext(
		runContext.RequestContext,
		sqliteWorkflowRunReadSQL,
		runContext.TenantRef,
		runContext.WorkspaceID,
		runContext.ApplicationID,
		runID,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return WorkflowRunRecord{}, false, nil
	}
	if err != nil {
		return WorkflowRunRecord{}, false, normalizeSQLiteWorkflowRunStoreError(err)
	}
	return record, true, nil
}

func (store *sqliteWorkflowRunStore) ListRuns(
	runContext WorkflowRunContext,
	filter WorkflowRunListFilter,
) (WorkflowRunListPage, error) {
	if store == nil || store.database == nil || runContext.RequestContext == nil {
		return WorkflowRunListPage{}, errWorkflowRunStoreContract
	}
	startedFrom, err := optionalWorkflowRunUnixNano(filter.StartedFrom)
	if err != nil {
		return WorkflowRunListPage{}, errWorkflowRunStoreContract
	}
	startedTo, err := optionalWorkflowRunUnixNano(filter.StartedTo)
	if err != nil {
		return WorkflowRunListPage{}, errWorkflowRunStoreContract
	}
	beforeTime, err := optionalWorkflowRunUnixNano(filter.BeforeTime)
	if err != nil {
		return WorkflowRunListPage{}, errWorkflowRunStoreContract
	}
	staleCutoff, err := workflowRunUnixNano(time.Now().UTC().Add(-workflowExecutorDefaultMaxRuntime))
	if err != nil {
		return WorkflowRunListPage{}, errWorkflowRunStoreContract
	}
	var staleRunning any
	if filter.StaleRunning != nil {
		if *filter.StaleRunning {
			staleRunning = 1
		} else {
			staleRunning = 0
		}
	}
	limit := workflowRunStoreListLimit(filter.Limit)
	rows, err := store.database.QueryContext(
		runContext.RequestContext,
		sqliteWorkflowRunListSQL,
		runContext.TenantRef,
		runContext.WorkspaceID,
		runContext.ApplicationID,
		string(filter.Status),
		string(filter.Status),
		filter.DraftID,
		filter.DraftID,
		string(filter.FailureCode),
		string(filter.FailureCode),
		string(filter.FailureBoundary),
		string(filter.FailureBoundary),
		filter.Provider,
		filter.Provider,
		filter.Model,
		filter.Model,
		staleRunning,
		staleCutoff,
		staleRunning,
		startedFrom,
		startedFrom,
		startedTo,
		startedTo,
		beforeTime,
		beforeTime,
		beforeTime,
		filter.BeforeRunID,
		limit+1,
	)
	if err != nil {
		return WorkflowRunListPage{}, errWorkflowRunStoreUnavailable
	}
	defer rows.Close()

	records := make([]WorkflowRunRecord, 0, limit+1)
	for rows.Next() {
		record, scanErr := scanSQLiteWorkflowRunRecord(runContext, rows)
		if scanErr != nil {
			return WorkflowRunListPage{}, normalizeSQLiteWorkflowRunStoreError(scanErr)
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

type sqliteWorkflowRunRow interface {
	Scan(dest ...any) error
}

func scanSQLiteWorkflowRunRecord(
	runContext WorkflowRunContext,
	row sqliteWorkflowRunRow,
) (WorkflowRunRecord, error) {
	var tenantRef string
	var workspaceID string
	var applicationID string
	var runID string
	var sourceKind string
	var sourceID string
	var sourceVersion int
	var recordVersion int
	var storeSchemaVersion string
	var schemaVersion string
	var status string
	var startedAtUnixNano int64
	var completedAtUnixNano sql.NullInt64
	var actorRef string
	var requestID string
	var auditRef string
	var failureCode string
	var failureBoundary string
	var selectedProvider string
	var selectedModel string
	var payload []byte
	if err := row.Scan(
		&tenantRef,
		&workspaceID,
		&applicationID,
		&runID,
		&sourceKind,
		&sourceID,
		&sourceVersion,
		&recordVersion,
		&storeSchemaVersion,
		&schemaVersion,
		&status,
		&startedAtUnixNano,
		&completedAtUnixNano,
		&actorRef,
		&requestID,
		&auditRef,
		&failureCode,
		&failureBoundary,
		&selectedProvider,
		&selectedModel,
		&payload,
	); err != nil {
		return WorkflowRunRecord{}, err
	}
	record, err := decodeWorkflowRunStorageRecord(runContext, payload)
	if err != nil {
		return WorkflowRunRecord{}, errWorkflowRunStoreContract
	}
	_, startedAt, completedAt, err := encodeWorkflowRunStorageRecord(record)
	if err != nil {
		return WorkflowRunRecord{}, errWorkflowRunStoreContract
	}
	decodedStartedAtUnixNano, err := workflowRunUnixNano(startedAt)
	if err != nil {
		return WorkflowRunRecord{}, errWorkflowRunStoreContract
	}
	decodedCompletedAtUnixNano, err := optionalWorkflowRunUnixNano(completedAt)
	if err != nil {
		return WorkflowRunRecord{}, errWorkflowRunStoreContract
	}
	decodedSourceKind, decodedSourceID, decodedSourceVersion, err := workflowRunStorageExecutionSource(record)
	if err != nil {
		return WorkflowRunRecord{}, errWorkflowRunStoreContract
	}
	if tenantRef != runContext.TenantRef || workspaceID != runContext.WorkspaceID || applicationID != runContext.ApplicationID || runID != record.RunID {
		return WorkflowRunRecord{}, fmt.Errorf("%w: scope projection", errWorkflowRunStoreContract)
	}
	if sourceKind != decodedSourceKind || sourceID != decodedSourceID || sourceVersion != decodedSourceVersion {
		return WorkflowRunRecord{}, fmt.Errorf("%w: execution source projection", errWorkflowRunStoreContract)
	}
	if recordVersion != record.RecordVersion || storeSchemaVersion != sqliteworkflowrunmigrations.RunRecordStoreSchemaVersion || schemaVersion != record.SchemaVersion || status != string(record.Status) {
		return WorkflowRunRecord{}, fmt.Errorf("%w: version projection", errWorkflowRunStoreContract)
	}
	if startedAtUnixNano != decodedStartedAtUnixNano || !sqliteWorkflowRunOptionalTimeMatches(completedAtUnixNano, decodedCompletedAtUnixNano) {
		return WorkflowRunRecord{}, fmt.Errorf("%w: timestamp projection", errWorkflowRunStoreContract)
	}
	if actorRef != record.ActorRef || requestID != record.RequestID || auditRef != record.AuditRef || failureCode != string(record.FailureCode) || failureBoundary != string(workflowRunRecordFailureBoundary(record)) || selectedProvider != record.SelectedProvider || selectedModel != record.SelectedModel {
		return WorkflowRunRecord{}, fmt.Errorf("%w: metadata projection", errWorkflowRunStoreContract)
	}
	return record, nil
}

func sqliteWorkflowRunOptionalTimeMatches(stored sql.NullInt64, decoded any) bool {
	if decoded == nil {
		return !stored.Valid
	}
	value, ok := decoded.(int64)
	return ok && stored.Valid && stored.Int64 == value
}

func normalizeSQLiteWorkflowRunStoreError(err error) error {
	if errors.Is(err, errWorkflowRunStoreContract) {
		return errWorkflowRunStoreContract
	}
	return errWorkflowRunStoreUnavailable
}

const sqliteWorkflowRunColumns = `
    tenant_ref,
    workspace_id,
    application_id,
    run_id,
    execution_source_kind,
    execution_source_id,
    execution_source_version,
    record_version,
    store_schema_version,
    schema_version,
    run_status,
    started_at_unix_nano,
    completed_at_unix_nano,
    actor_ref,
    request_id,
    audit_ref,
    failure_code,
    failure_boundary,
    selected_provider,
    selected_model,
    sanitized_run_record`

const sqliteWorkflowRunInsertSQL = `
INSERT INTO workflow_run_records (
    tenant_ref, workspace_id, application_id, run_id,
    execution_source_kind, execution_source_id, execution_source_version,
    record_version, store_schema_version, schema_version, run_status,
    started_at_unix_nano, completed_at_unix_nano, actor_ref, request_id, audit_ref,
    failure_code, failure_boundary, selected_provider, selected_model, sanitized_run_record
) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
ON CONFLICT (tenant_ref, workspace_id, application_id, run_id) DO NOTHING
RETURNING ` + sqliteWorkflowRunColumns

const sqliteWorkflowRunUpdateSQL = `
UPDATE workflow_run_records
   SET execution_source_kind=?, execution_source_id=?, execution_source_version=?,
       record_version=record_version+1, schema_version=?, run_status=?,
       completed_at_unix_nano=?, actor_ref=?, request_id=?, audit_ref=?, failure_code=?,
       failure_boundary=?, selected_provider=?, selected_model=?, sanitized_run_record=?
 WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND run_id=?
   AND record_version=? AND run_status='running' AND started_at_unix_nano=?
RETURNING ` + sqliteWorkflowRunColumns

const sqliteWorkflowRunReadSQL = `
SELECT ` + sqliteWorkflowRunColumns + `
  FROM workflow_run_records
 WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND run_id=?`

const sqliteWorkflowRunListSQL = `
SELECT ` + sqliteWorkflowRunColumns + `
  FROM workflow_run_records
 WHERE tenant_ref=? AND workspace_id=? AND application_id=?
   AND (?='' OR run_status=?)
   AND (?='' OR (execution_source_kind='workflow_draft' AND execution_source_id=?))
   AND (?='' OR failure_code=?)
   AND (?='' OR failure_boundary=?)
   AND (?='' OR selected_provider=?)
   AND (?='' OR selected_model=?)
   AND (? IS NULL OR CASE WHEN run_status='running' AND started_at_unix_nano < ? THEN 1 ELSE 0 END = ?)
   AND (? IS NULL OR started_at_unix_nano >= ?)
   AND (? IS NULL OR started_at_unix_nano <= ?)
   AND (? IS NULL OR started_at_unix_nano < ? OR (started_at_unix_nano = ? AND run_id < ?))
 ORDER BY started_at_unix_nano DESC, run_id DESC
 LIMIT ?`

var _ workflowRunStore = (*sqliteWorkflowRunStore)(nil)
