package httpapi

import (
	"database/sql"
	"errors"
	"sort"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type combinedWorkflowRunStore struct {
	workflow workflowRunStore
	prompt   workflowRunStore
}

func newCombinedWorkflowRunStore(workflow, prompt workflowRunStore) workflowRunStore {
	return &combinedWorkflowRunStore{workflow: workflow, prompt: prompt}
}

func (store *combinedWorkflowRunStore) UpsertRun(ctx WorkflowRunContext, record *WorkflowRunRecord) error {
	if record != nil && record.SchemaVersion == workflowRunRecordPromptSchemaVersion {
		return store.prompt.UpsertRun(ctx, record)
	}
	return store.workflow.UpsertRun(ctx, record)
}

func (store *combinedWorkflowRunStore) ReadRun(ctx WorkflowRunContext, runID string) (WorkflowRunRecord, bool, error) {
	record, found, err := store.workflow.ReadRun(ctx, runID)
	if err != nil || found {
		return record, found, err
	}
	return store.prompt.ReadRun(ctx, runID)
}

func (store *combinedWorkflowRunStore) ListRuns(ctx WorkflowRunContext, filter WorkflowRunListFilter) (WorkflowRunListPage, error) {
	workflowPage, err := store.workflow.ListRuns(ctx, filter)
	if err != nil {
		return WorkflowRunListPage{}, err
	}
	promptPage, err := store.prompt.ListRuns(ctx, filter)
	if err != nil {
		return WorkflowRunListPage{}, err
	}
	records := append(workflowPage.Records, promptPage.Records...)
	sort.Slice(records, func(left, right int) bool {
		if records[left].StartedAt == records[right].StartedAt {
			return records[left].RunID > records[right].RunID
		}
		return records[left].StartedAt > records[right].StartedAt
	})
	limit := workflowRunStoreListLimit(filter.Limit)
	hasMore := workflowPage.HasMore || promptPage.HasMore || len(records) > limit
	if len(records) > limit {
		records = records[:limit]
	}
	return WorkflowRunListPage{Records: records, HasMore: hasMore}, nil
}

func newPromptApplicationRunStoreForWorkflowRunStore(store workflowRunStore) (workflowRunStore, error) {
	switch typed := store.(type) {
	case *memoryWorkflowRunStore:
		return newMemoryWorkflowRunStore(typed.capacity), nil
	case *sqliteWorkflowRunStore:
		if typed.database == nil {
			return nil, errors.New("Prompt application run store requires the shared SQLite database")
		}
		return &sqlitePromptApplicationRunStore{database: typed.database}, nil
	case *postgresWorkflowRunStore:
		if typed.pool == nil {
			return nil, errors.New("Prompt application run store requires the Workflow PostgreSQL pool")
		}
		return &postgresPromptApplicationRunStore{pool: typed.pool}, nil
	default:
		return nil, errors.New("Prompt application run store requires a supported Workflow runtime backend")
	}
}

type sqlitePromptApplicationRunStore struct{ database *sql.DB }

func (store *sqlitePromptApplicationRunStore) UpsertRun(ctx WorkflowRunContext, record *WorkflowRunRecord) error {
	if store == nil || store.database == nil || ctx.RequestContext == nil || record == nil ||
		validateWorkflowRunStoreRecord(ctx, record) != nil || record.SchemaVersion != workflowRunRecordPromptSchemaVersion {
		return errWorkflowRunStoreContract
	}
	next := cloneWorkflowRunRecord(*record)
	if record.RecordVersion == 0 && record.Status != WorkflowRunStatusRunning {
		return errWorkflowRunStoreConflict
	}
	next.RecordVersion++
	payload, startedAt, completedAt, err := encodeWorkflowRunStorageRecord(next)
	if err != nil {
		return err
	}
	startedUnix, err := workflowRunUnixNano(startedAt)
	if err != nil {
		return errWorkflowRunStoreContract
	}
	completedUnix, err := optionalWorkflowRunUnixNano(completedAt)
	if err != nil {
		return errWorkflowRunStoreContract
	}
	var storedPayload string
	if record.RecordVersion == 0 {
		err = store.database.QueryRowContext(ctx.RequestContext, `INSERT INTO prompt_application_run_records
(tenant_ref,workspace_id,application_id,run_id,record_version,run_status,template_id,template_version,authority_digest,started_at_unix_nano,completed_at_unix_nano,sanitized_run_payload)
VALUES(?,?,?,?,1,?,?,?,?,?,?,?) RETURNING sanitized_run_payload`,
			ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, next.RunID, next.Status,
			next.ExecutionSource.ID, next.ExecutionSource.Version, next.PromptApplication.AuthorityDigest,
			startedUnix, completedUnix, string(payload)).Scan(&storedPayload)
	} else {
		err = store.database.QueryRowContext(ctx.RequestContext, `UPDATE prompt_application_run_records SET
record_version=record_version+1,run_status=?,completed_at_unix_nano=?,sanitized_run_payload=?
WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND run_id=? AND record_version=? AND run_status='running'
RETURNING sanitized_run_payload`,
			next.Status, completedUnix, string(payload), ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID,
			next.RunID, record.RecordVersion).Scan(&storedPayload)
	}
	if errors.Is(err, sql.ErrNoRows) {
		return errWorkflowRunStoreConflict
	}
	if err != nil {
		return errWorkflowRunStoreUnavailable
	}
	stored, err := decodeWorkflowRunStorageRecord(ctx, []byte(storedPayload))
	if err != nil {
		return err
	}
	*record = stored
	return nil
}

func (store *sqlitePromptApplicationRunStore) ReadRun(ctx WorkflowRunContext, runID string) (WorkflowRunRecord, bool, error) {
	if store == nil || store.database == nil || ctx.RequestContext == nil {
		return WorkflowRunRecord{}, false, errWorkflowRunStoreContract
	}
	var payload string
	err := store.database.QueryRowContext(ctx.RequestContext, `SELECT sanitized_run_payload FROM prompt_application_run_records
WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND run_id=?`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, runID).Scan(&payload)
	if errors.Is(err, sql.ErrNoRows) {
		return WorkflowRunRecord{}, false, nil
	}
	if err != nil {
		return WorkflowRunRecord{}, false, errWorkflowRunStoreUnavailable
	}
	record, err := decodeWorkflowRunStorageRecord(ctx, []byte(payload))
	return record, err == nil, err
}

func (store *sqlitePromptApplicationRunStore) ListRuns(ctx WorkflowRunContext, filter WorkflowRunListFilter) (WorkflowRunListPage, error) {
	if store == nil || store.database == nil || ctx.RequestContext == nil {
		return WorkflowRunListPage{}, errWorkflowRunStoreContract
	}
	rows, err := store.database.QueryContext(ctx.RequestContext, `SELECT sanitized_run_payload FROM prompt_application_run_records
WHERE tenant_ref=? AND workspace_id=? AND application_id=? ORDER BY started_at_unix_nano DESC,run_id DESC`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID)
	if err != nil {
		return WorkflowRunListPage{}, errWorkflowRunStoreUnavailable
	}
	defer rows.Close()
	return collectPromptApplicationRuns(ctx, filter, func() ([]byte, bool, error) {
		if !rows.Next() {
			return nil, false, rows.Err()
		}
		var payload string
		if scanErr := rows.Scan(&payload); scanErr != nil {
			return nil, false, scanErr
		}
		return []byte(payload), true, nil
	})
}

type postgresPromptApplicationRunStore struct{ pool *pgxpool.Pool }

func (store *postgresPromptApplicationRunStore) UpsertRun(ctx WorkflowRunContext, record *WorkflowRunRecord) error {
	if store == nil || store.pool == nil || ctx.RequestContext == nil || record == nil ||
		validateWorkflowRunStoreRecord(ctx, record) != nil || record.SchemaVersion != workflowRunRecordPromptSchemaVersion {
		return errWorkflowRunStoreContract
	}
	next := cloneWorkflowRunRecord(*record)
	if record.RecordVersion == 0 && record.Status != WorkflowRunStatusRunning {
		return errWorkflowRunStoreConflict
	}
	next.RecordVersion++
	payload, startedAt, completedAt, err := encodeWorkflowRunStorageRecord(next)
	if err != nil {
		return err
	}
	var storedPayload []byte
	if record.RecordVersion == 0 {
		err = store.pool.QueryRow(ctx.RequestContext, `INSERT INTO prompt_application_run_records
(tenant_ref,workspace_id,application_id,run_id,record_version,run_status,template_id,template_version,authority_digest,started_at,completed_at,sanitized_run_payload)
VALUES($1,$2,$3,$4,1,$5,$6,$7,$8,$9,$10,$11) ON CONFLICT DO NOTHING RETURNING sanitized_run_payload`,
			ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, next.RunID, next.Status,
			next.ExecutionSource.ID, next.ExecutionSource.Version, next.PromptApplication.AuthorityDigest,
			startedAt, completedAt, payload).Scan(&storedPayload)
	} else {
		err = store.pool.QueryRow(ctx.RequestContext, `UPDATE prompt_application_run_records SET
record_version=record_version+1,run_status=$1,completed_at=$2,sanitized_run_payload=$3
WHERE tenant_ref=$4 AND workspace_id=$5 AND application_id=$6 AND run_id=$7 AND record_version=$8 AND run_status='running'
RETURNING sanitized_run_payload`,
			next.Status, completedAt, payload, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID,
			next.RunID, record.RecordVersion).Scan(&storedPayload)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return errWorkflowRunStoreConflict
	}
	if err != nil {
		return errWorkflowRunStoreUnavailable
	}
	stored, err := decodeWorkflowRunStorageRecord(ctx, storedPayload)
	if err != nil {
		return err
	}
	*record = stored
	return nil
}

func (store *postgresPromptApplicationRunStore) ReadRun(ctx WorkflowRunContext, runID string) (WorkflowRunRecord, bool, error) {
	if store == nil || store.pool == nil || ctx.RequestContext == nil {
		return WorkflowRunRecord{}, false, errWorkflowRunStoreContract
	}
	var payload []byte
	err := store.pool.QueryRow(ctx.RequestContext, `SELECT sanitized_run_payload FROM prompt_application_run_records
WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND run_id=$4`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, runID).Scan(&payload)
	if errors.Is(err, pgx.ErrNoRows) {
		return WorkflowRunRecord{}, false, nil
	}
	if err != nil {
		return WorkflowRunRecord{}, false, errWorkflowRunStoreUnavailable
	}
	record, err := decodeWorkflowRunStorageRecord(ctx, payload)
	return record, err == nil, err
}

func (store *postgresPromptApplicationRunStore) ListRuns(ctx WorkflowRunContext, filter WorkflowRunListFilter) (WorkflowRunListPage, error) {
	if store == nil || store.pool == nil || ctx.RequestContext == nil {
		return WorkflowRunListPage{}, errWorkflowRunStoreContract
	}
	rows, err := store.pool.Query(ctx.RequestContext, `SELECT sanitized_run_payload FROM prompt_application_run_records
WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 ORDER BY started_at DESC,run_id DESC`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID)
	if err != nil {
		return WorkflowRunListPage{}, errWorkflowRunStoreUnavailable
	}
	defer rows.Close()
	return collectPromptApplicationRuns(ctx, filter, func() ([]byte, bool, error) {
		if !rows.Next() {
			return nil, false, rows.Err()
		}
		var payload []byte
		if scanErr := rows.Scan(&payload); scanErr != nil {
			return nil, false, scanErr
		}
		return payload, true, nil
	})
}

func collectPromptApplicationRuns(ctx WorkflowRunContext, filter WorkflowRunListFilter, next func() ([]byte, bool, error)) (WorkflowRunListPage, error) {
	records := make([]WorkflowRunRecord, 0)
	for {
		payload, ok, err := next()
		if err != nil {
			return WorkflowRunListPage{}, errWorkflowRunStoreUnavailable
		}
		if !ok {
			break
		}
		record, err := decodeWorkflowRunStorageRecord(ctx, payload)
		if err != nil {
			return WorkflowRunListPage{}, err
		}
		if workflowRunMatchesFilter(record, filter) {
			records = append(records, record)
		}
	}
	limit := workflowRunStoreListLimit(filter.Limit)
	hasMore := len(records) > limit
	if hasMore {
		records = records[:limit]
	}
	return WorkflowRunListPage{Records: records, HasMore: hasMore}, nil
}

var _ workflowRunStore = (*combinedWorkflowRunStore)(nil)
var _ workflowRunStore = (*sqlitePromptApplicationRunStore)(nil)
var _ workflowRunStore = (*postgresPromptApplicationRunStore)(nil)
