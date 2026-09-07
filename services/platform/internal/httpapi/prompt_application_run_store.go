package httpapi

import (
	"database/sql"
	"errors"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type combinedWorkflowRunStore struct {
	workflow workflowRunStore
	prompt   workflowRunStore
	agent    workflowRunStore
}

func newCombinedWorkflowRunStore(workflow, prompt workflowRunStore) workflowRunStore {
	return &combinedWorkflowRunStore{workflow: workflow, prompt: prompt}
}

func newCombinedWorkflowRunStoreWithAgent(workflow, prompt, agent workflowRunStore) workflowRunStore {
	return &combinedWorkflowRunStore{workflow: workflow, prompt: prompt, agent: agent}
}

func (store *combinedWorkflowRunStore) UpsertRun(ctx WorkflowRunContext, record *WorkflowRunRecord) error {
	if record != nil && record.SchemaVersion == workflowRunRecordPromptSchemaVersion {
		return store.prompt.UpsertRun(ctx, record)
	}
	if record != nil && record.SchemaVersion == agentCopilotRunV7Schema {
		return store.agent.UpsertRun(ctx, record)
	}
	return store.workflow.UpsertRun(ctx, record)
}

func (store *combinedWorkflowRunStore) ReadRun(ctx WorkflowRunContext, runID string) (WorkflowRunRecord, bool, error) {
	record, found, err := store.workflow.ReadRun(ctx, runID)
	if err != nil || found {
		return record, found, err
	}
	record, found, err = store.prompt.ReadRun(ctx, runID)
	if err != nil || found || store.agent == nil {
		return record, found, err
	}
	return store.agent.ReadRun(ctx, runID)
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
	agentHasMore := false
	if store.agent != nil {
		agentPage, agentErr := store.agent.ListRuns(ctx, filter)
		if agentErr != nil {
			return WorkflowRunListPage{}, agentErr
		}
		records = append(records, agentPage.Records...)
		agentHasMore = agentPage.HasMore
	}
	sort.Slice(records, func(left, right int) bool {
		if records[left].StartedAt == records[right].StartedAt {
			return records[left].RunID > records[right].RunID
		}
		return records[left].StartedAt > records[right].StartedAt
	})
	limit := workflowRunStoreListLimit(filter.Limit)
	hasMore := workflowPage.HasMore || promptPage.HasMore || agentHasMore || len(records) > limit
	if len(records) > limit {
		records = records[:limit]
	}
	return WorkflowRunListPage{Records: records, HasMore: hasMore}, nil
}

func (store *combinedWorkflowRunStore) ListWorkspaceRuns(
	ctx WorkflowWorkspaceRunListContext,
	filter WorkflowWorkspaceRunListFilter,
) (WorkflowRunListPage, error) {
	records := make([]WorkflowRunRecord, 0)
	hasMore := false
	for _, runStore := range []workflowRunStore{store.workflow, store.prompt, store.agent} {
		if runStore == nil {
			continue
		}
		projection, ok := runStore.(workflowWorkspaceRunProjection)
		if !ok {
			return WorkflowRunListPage{}, errWorkflowRunStoreUnavailable
		}
		page, err := projection.ListWorkspaceRuns(ctx, filter)
		if err != nil {
			return WorkflowRunListPage{}, err
		}
		records = append(records, page.Records...)
		hasMore = hasMore || page.HasMore
	}
	sort.Slice(records, func(left, right int) bool {
		if records[left].StartedAt != records[right].StartedAt {
			return records[left].StartedAt > records[right].StartedAt
		}
		if records[left].RunID != records[right].RunID {
			return records[left].RunID > records[right].RunID
		}
		return records[left].ApplicationID > records[right].ApplicationID
	})
	limit := workflowRunStoreListLimit(filter.Limit)
	hasMore = hasMore || len(records) > limit
	if len(records) > limit {
		records = records[:limit]
	}
	return WorkflowRunListPage{Records: records, HasMore: hasMore}, nil
}

func newAgentCopilotRunStoreForWorkflowRunStore(store workflowRunStore) (workflowRunStore, error) {
	switch typed := store.(type) {
	case *memoryWorkflowRunStore:
		return newMemoryWorkflowRunStore(typed.capacity), nil
	case *sqliteWorkflowRunStore:
		if typed.database == nil {
			return nil, errors.New("Agent Copilot run store requires the shared SQLite database")
		}
		return &sqlitePromptApplicationRunStore{
			database: typed.database, table: "agent_copilot_run_records", schema: agentCopilotRunV7Schema,
		}, nil
	case *postgresWorkflowRunStore:
		if typed.pool == nil {
			return nil, errors.New("Agent Copilot run store requires the Workflow PostgreSQL pool")
		}
		return &postgresPromptApplicationRunStore{
			pool: typed.pool, table: "agent_copilot_run_records", schema: agentCopilotRunV7Schema,
		}, nil
	default:
		return nil, errors.New("Agent Copilot run store requires a supported Workflow runtime backend")
	}
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

type sqlitePromptApplicationRunStore struct {
	database *sql.DB
	table    string
	schema   string
}

func (store *sqlitePromptApplicationRunStore) UpsertRun(ctx WorkflowRunContext, record *WorkflowRunRecord) error {
	if store == nil || store.database == nil || ctx.RequestContext == nil || record == nil ||
		validateWorkflowRunStoreRecord(ctx, record) != nil || record.SchemaVersion != store.expectedSchema() {
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
	safety, err := encodeActionSafetyRunSnapshot(next.ActionSafety)
	if err != nil {
		return errWorkflowRunStoreContract
	}
	safetySchema, safetyDigest, safetyPayload := safety.sqliteColumnValues()
	startedUnix, err := workflowRunUnixNano(startedAt)
	if err != nil {
		return errWorkflowRunStoreContract
	}
	completedUnix, err := optionalWorkflowRunUnixNano(completedAt)
	if err != nil {
		return errWorkflowRunStoreContract
	}
	var storedPayload string
	var storedSafetySchema, storedSafetyDigest sql.NullString
	var storedSafetyPayload []byte
	if store.hasActionSafetySnapshot() && record.RecordVersion == 0 {
		err = store.database.QueryRowContext(ctx.RequestContext, store.statement(`INSERT INTO prompt_application_run_records
(tenant_ref,workspace_id,application_id,run_id,record_version,run_status,template_id,template_version,authority_digest,started_at_unix_nano,completed_at_unix_nano,sanitized_run_payload,action_safety_schema_version,action_safety_projection_digest,sanitized_action_safety_snapshot)
VALUES(?,?,?,?,1,?,?,?,?,?,?,?,?,?,?) RETURNING sanitized_run_payload,action_safety_schema_version,action_safety_projection_digest,sanitized_action_safety_snapshot`),
			ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, next.RunID, next.Status,
			next.ExecutionSource.ID, next.ExecutionSource.Version, applicationProjectionRunAuthorityDigest(next),
			startedUnix, completedUnix, string(payload), safetySchema, safetyDigest, safetyPayload,
		).Scan(&storedPayload, &storedSafetySchema, &storedSafetyDigest, &storedSafetyPayload)
	} else if store.hasActionSafetySnapshot() {
		err = store.database.QueryRowContext(ctx.RequestContext, store.statement(`UPDATE prompt_application_run_records SET
record_version=record_version+1,run_status=?,completed_at_unix_nano=?,sanitized_run_payload=?,action_safety_schema_version=?,action_safety_projection_digest=?,sanitized_action_safety_snapshot=?
WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND run_id=? AND record_version=? AND run_status='running'
RETURNING sanitized_run_payload,action_safety_schema_version,action_safety_projection_digest,sanitized_action_safety_snapshot`),
			next.Status, completedUnix, string(payload), safetySchema, safetyDigest, safetyPayload,
			ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, next.RunID, record.RecordVersion,
		).Scan(&storedPayload, &storedSafetySchema, &storedSafetyDigest, &storedSafetyPayload)
	} else if record.RecordVersion == 0 {
		err = store.database.QueryRowContext(ctx.RequestContext, store.statement(`INSERT INTO prompt_application_run_records
(tenant_ref,workspace_id,application_id,run_id,record_version,run_status,template_id,template_version,authority_digest,started_at_unix_nano,completed_at_unix_nano,sanitized_run_payload)
VALUES(?,?,?,?,1,?,?,?,?,?,?,?) RETURNING sanitized_run_payload`),
			ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, next.RunID, next.Status,
			next.ExecutionSource.ID, next.ExecutionSource.Version, applicationProjectionRunAuthorityDigest(next),
			startedUnix, completedUnix, string(payload)).Scan(&storedPayload)
	} else {
		err = store.database.QueryRowContext(ctx.RequestContext, store.statement(`UPDATE prompt_application_run_records SET
record_version=record_version+1,run_status=?,completed_at_unix_nano=?,sanitized_run_payload=?
WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND run_id=? AND record_version=? AND run_status='running'
RETURNING sanitized_run_payload`),
			next.Status, completedUnix, string(payload), ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID,
			next.RunID, record.RecordVersion).Scan(&storedPayload)
	}
	if errors.Is(err, sql.ErrNoRows) {
		return errWorkflowRunStoreConflict
	}
	if err != nil {
		return errWorkflowRunStoreUnavailable
	}
	stored, err := decodeWorkflowRunStorageRecordWithActionSafety(ctx, []byte(storedPayload), actionSafetyStorageSnapshot{
		SchemaVersion: storedSafetySchema.String, ProjectionDigest: storedSafetyDigest.String, Payload: storedSafetyPayload,
	})
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
	var safetySchema, safetyDigest sql.NullString
	var safetyPayload []byte
	query := `SELECT sanitized_run_payload FROM prompt_application_run_records
WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND run_id=?`
	if store.hasActionSafetySnapshot() {
		query = `SELECT sanitized_run_payload,action_safety_schema_version,action_safety_projection_digest,sanitized_action_safety_snapshot FROM prompt_application_run_records
WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND run_id=?`
	}
	row := store.database.QueryRowContext(ctx.RequestContext, store.statement(query), ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, runID)
	var err error
	if store.hasActionSafetySnapshot() {
		err = row.Scan(&payload, &safetySchema, &safetyDigest, &safetyPayload)
	} else {
		err = row.Scan(&payload)
	}
	if errors.Is(err, sql.ErrNoRows) {
		return WorkflowRunRecord{}, false, nil
	}
	if err != nil {
		return WorkflowRunRecord{}, false, errWorkflowRunStoreUnavailable
	}
	record, err := decodeWorkflowRunStorageRecordWithActionSafety(ctx, []byte(payload), actionSafetyStorageSnapshot{
		SchemaVersion: safetySchema.String, ProjectionDigest: safetyDigest.String, Payload: safetyPayload,
	})
	return record, err == nil, err
}

func (store *sqlitePromptApplicationRunStore) ListRuns(ctx WorkflowRunContext, filter WorkflowRunListFilter) (WorkflowRunListPage, error) {
	if store == nil || store.database == nil || ctx.RequestContext == nil {
		return WorkflowRunListPage{}, errWorkflowRunStoreContract
	}
	query := `SELECT sanitized_run_payload FROM prompt_application_run_records
WHERE tenant_ref=? AND workspace_id=? AND application_id=? ORDER BY started_at_unix_nano DESC,run_id DESC`
	if store.hasActionSafetySnapshot() {
		query = `SELECT sanitized_run_payload,action_safety_schema_version,action_safety_projection_digest,sanitized_action_safety_snapshot FROM prompt_application_run_records
WHERE tenant_ref=? AND workspace_id=? AND application_id=? ORDER BY started_at_unix_nano DESC,run_id DESC`
	}
	rows, err := store.database.QueryContext(ctx.RequestContext, store.statement(query),
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID)
	if err != nil {
		return WorkflowRunListPage{}, errWorkflowRunStoreUnavailable
	}
	defer rows.Close()
	return collectPromptApplicationRuns(ctx, filter, func() ([]byte, actionSafetyStorageSnapshot, bool, error) {
		if !rows.Next() {
			return nil, actionSafetyStorageSnapshot{}, false, rows.Err()
		}
		var payload string
		var safetySchema, safetyDigest sql.NullString
		var safetyPayload []byte
		var scanErr error
		if store.hasActionSafetySnapshot() {
			scanErr = rows.Scan(&payload, &safetySchema, &safetyDigest, &safetyPayload)
		} else {
			scanErr = rows.Scan(&payload)
		}
		if scanErr != nil {
			return nil, actionSafetyStorageSnapshot{}, false, scanErr
		}
		return []byte(payload), actionSafetyStorageSnapshot{SchemaVersion: safetySchema.String, ProjectionDigest: safetyDigest.String, Payload: safetyPayload}, true, nil
	})
}

func (store *sqlitePromptApplicationRunStore) ListWorkspaceRuns(
	ctx WorkflowWorkspaceRunListContext,
	filter WorkflowWorkspaceRunListFilter,
) (WorkflowRunListPage, error) {
	if store == nil || store.database == nil || !validWorkflowWorkspaceRunListContext(ctx) {
		return WorkflowRunListPage{}, errWorkflowRunStoreContract
	}
	beforeTime, err := optionalWorkflowRunUnixNano(filter.BeforeTime)
	if err != nil {
		return WorkflowRunListPage{}, errWorkflowRunStoreContract
	}
	query := `SELECT application_id,sanitized_run_payload FROM prompt_application_run_records
WHERE tenant_ref=? AND workspace_id=? AND (?='' OR application_id=?)
AND (? IS NULL OR started_at_unix_nano < ? OR
    (started_at_unix_nano = ? AND (run_id < ? OR (run_id = ? AND application_id < ?))))
ORDER BY started_at_unix_nano DESC,run_id DESC,application_id DESC`
	if store.hasActionSafetySnapshot() {
		query = `SELECT application_id,sanitized_run_payload,action_safety_schema_version,action_safety_projection_digest,sanitized_action_safety_snapshot FROM prompt_application_run_records
WHERE tenant_ref=? AND workspace_id=? AND (?='' OR application_id=?)
AND (? IS NULL OR started_at_unix_nano < ? OR
    (started_at_unix_nano = ? AND (run_id < ? OR (run_id = ? AND application_id < ?))))
ORDER BY started_at_unix_nano DESC,run_id DESC,application_id DESC`
	}
	rows, err := store.database.QueryContext(ctx.RequestContext, store.statement(query),
		ctx.TenantRef, ctx.WorkspaceID, filter.ApplicationID, filter.ApplicationID,
		beforeTime, beforeTime, beforeTime, filter.BeforeRunID, filter.BeforeRunID, filter.BeforeApplicationID)
	if err != nil {
		return WorkflowRunListPage{}, errWorkflowRunStoreUnavailable
	}
	defer rows.Close()
	return collectWorkspaceApplicationProjectionRuns(ctx, filter, store.expectedSchema(), func() (string, []byte, actionSafetyStorageSnapshot, bool, error) {
		if !rows.Next() {
			return "", nil, actionSafetyStorageSnapshot{}, false, rows.Err()
		}
		var applicationID, payload string
		var safetySchema, safetyDigest sql.NullString
		var safetyPayload []byte
		var scanErr error
		if store.hasActionSafetySnapshot() {
			scanErr = rows.Scan(&applicationID, &payload, &safetySchema, &safetyDigest, &safetyPayload)
		} else {
			scanErr = rows.Scan(&applicationID, &payload)
		}
		if scanErr != nil {
			return "", nil, actionSafetyStorageSnapshot{}, false, scanErr
		}
		return applicationID, []byte(payload), actionSafetyStorageSnapshot{SchemaVersion: safetySchema.String, ProjectionDigest: safetyDigest.String, Payload: safetyPayload}, true, nil
	})
}

type postgresPromptApplicationRunStore struct {
	pool   *pgxpool.Pool
	table  string
	schema string
}

func (store *postgresPromptApplicationRunStore) UpsertRun(ctx WorkflowRunContext, record *WorkflowRunRecord) error {
	if store == nil || store.pool == nil || ctx.RequestContext == nil || record == nil ||
		validateWorkflowRunStoreRecord(ctx, record) != nil || record.SchemaVersion != store.expectedSchema() {
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
	safety, err := encodeActionSafetyRunSnapshot(next.ActionSafety)
	if err != nil {
		return errWorkflowRunStoreContract
	}
	safetySchema, safetyDigest, safetyPayload := safety.columnValues()
	var storedPayload, storedSafetyPayload []byte
	var storedSafetySchema, storedSafetyDigest sql.NullString
	if store.hasActionSafetySnapshot() && record.RecordVersion == 0 {
		err = store.pool.QueryRow(ctx.RequestContext, store.statement(`INSERT INTO prompt_application_run_records
(tenant_ref,workspace_id,application_id,run_id,record_version,run_status,template_id,template_version,authority_digest,started_at,completed_at,sanitized_run_payload,action_safety_schema_version,action_safety_projection_digest,sanitized_action_safety_snapshot)
VALUES($1,$2,$3,$4,1,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14) ON CONFLICT DO NOTHING
RETURNING sanitized_run_payload,action_safety_schema_version,action_safety_projection_digest,sanitized_action_safety_snapshot`),
			ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, next.RunID, next.Status,
			next.ExecutionSource.ID, next.ExecutionSource.Version, applicationProjectionRunAuthorityDigest(next),
			startedAt, completedAt, payload, safetySchema, safetyDigest, safetyPayload,
		).Scan(&storedPayload, &storedSafetySchema, &storedSafetyDigest, &storedSafetyPayload)
	} else if store.hasActionSafetySnapshot() {
		err = store.pool.QueryRow(ctx.RequestContext, store.statement(`UPDATE prompt_application_run_records SET
record_version=record_version+1,run_status=$1,completed_at=$2,sanitized_run_payload=$3,action_safety_schema_version=$4,action_safety_projection_digest=$5,sanitized_action_safety_snapshot=$6
WHERE tenant_ref=$7 AND workspace_id=$8 AND application_id=$9 AND run_id=$10 AND record_version=$11 AND run_status='running'
RETURNING sanitized_run_payload,action_safety_schema_version,action_safety_projection_digest,sanitized_action_safety_snapshot`),
			next.Status, completedAt, payload, safetySchema, safetyDigest, safetyPayload,
			ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, next.RunID, record.RecordVersion,
		).Scan(&storedPayload, &storedSafetySchema, &storedSafetyDigest, &storedSafetyPayload)
	} else if record.RecordVersion == 0 {
		err = store.pool.QueryRow(ctx.RequestContext, store.statement(`INSERT INTO prompt_application_run_records
(tenant_ref,workspace_id,application_id,run_id,record_version,run_status,template_id,template_version,authority_digest,started_at,completed_at,sanitized_run_payload)
VALUES($1,$2,$3,$4,1,$5,$6,$7,$8,$9,$10,$11) ON CONFLICT DO NOTHING RETURNING sanitized_run_payload`),
			ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, next.RunID, next.Status,
			next.ExecutionSource.ID, next.ExecutionSource.Version, applicationProjectionRunAuthorityDigest(next),
			startedAt, completedAt, payload).Scan(&storedPayload)
	} else {
		err = store.pool.QueryRow(ctx.RequestContext, store.statement(`UPDATE prompt_application_run_records SET
record_version=record_version+1,run_status=$1,completed_at=$2,sanitized_run_payload=$3
WHERE tenant_ref=$4 AND workspace_id=$5 AND application_id=$6 AND run_id=$7 AND record_version=$8 AND run_status='running'
RETURNING sanitized_run_payload`),
			next.Status, completedAt, payload, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID,
			next.RunID, record.RecordVersion).Scan(&storedPayload)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return errWorkflowRunStoreConflict
	}
	if err != nil {
		return errWorkflowRunStoreUnavailable
	}
	stored, err := decodeWorkflowRunStorageRecordWithActionSafety(ctx, storedPayload, actionSafetyStorageSnapshot{
		SchemaVersion: storedSafetySchema.String, ProjectionDigest: storedSafetyDigest.String, Payload: storedSafetyPayload,
	})
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
	var payload, safetyPayload []byte
	var safetySchema, safetyDigest sql.NullString
	query := `SELECT sanitized_run_payload FROM prompt_application_run_records
WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND run_id=$4`
	if store.hasActionSafetySnapshot() {
		query = `SELECT sanitized_run_payload,action_safety_schema_version,action_safety_projection_digest,sanitized_action_safety_snapshot FROM prompt_application_run_records
WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND run_id=$4`
	}
	row := store.pool.QueryRow(ctx.RequestContext, store.statement(query), ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, runID)
	var err error
	if store.hasActionSafetySnapshot() {
		err = row.Scan(&payload, &safetySchema, &safetyDigest, &safetyPayload)
	} else {
		err = row.Scan(&payload)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return WorkflowRunRecord{}, false, nil
	}
	if err != nil {
		return WorkflowRunRecord{}, false, errWorkflowRunStoreUnavailable
	}
	record, err := decodeWorkflowRunStorageRecordWithActionSafety(ctx, payload, actionSafetyStorageSnapshot{
		SchemaVersion: safetySchema.String, ProjectionDigest: safetyDigest.String, Payload: safetyPayload,
	})
	return record, err == nil, err
}

func (store *postgresPromptApplicationRunStore) ListRuns(ctx WorkflowRunContext, filter WorkflowRunListFilter) (WorkflowRunListPage, error) {
	if store == nil || store.pool == nil || ctx.RequestContext == nil {
		return WorkflowRunListPage{}, errWorkflowRunStoreContract
	}
	query := `SELECT sanitized_run_payload FROM prompt_application_run_records
WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 ORDER BY started_at DESC,run_id DESC`
	if store.hasActionSafetySnapshot() {
		query = `SELECT sanitized_run_payload,action_safety_schema_version,action_safety_projection_digest,sanitized_action_safety_snapshot FROM prompt_application_run_records
WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 ORDER BY started_at DESC,run_id DESC`
	}
	rows, err := store.pool.Query(ctx.RequestContext, store.statement(query),
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID)
	if err != nil {
		return WorkflowRunListPage{}, errWorkflowRunStoreUnavailable
	}
	defer rows.Close()
	return collectPromptApplicationRuns(ctx, filter, func() ([]byte, actionSafetyStorageSnapshot, bool, error) {
		if !rows.Next() {
			return nil, actionSafetyStorageSnapshot{}, false, rows.Err()
		}
		var payload, safetyPayload []byte
		var safetySchema, safetyDigest sql.NullString
		var scanErr error
		if store.hasActionSafetySnapshot() {
			scanErr = rows.Scan(&payload, &safetySchema, &safetyDigest, &safetyPayload)
		} else {
			scanErr = rows.Scan(&payload)
		}
		if scanErr != nil {
			return nil, actionSafetyStorageSnapshot{}, false, scanErr
		}
		return payload, actionSafetyStorageSnapshot{SchemaVersion: safetySchema.String, ProjectionDigest: safetyDigest.String, Payload: safetyPayload}, true, nil
	})
}

func (store *postgresPromptApplicationRunStore) ListWorkspaceRuns(
	ctx WorkflowWorkspaceRunListContext,
	filter WorkflowWorkspaceRunListFilter,
) (WorkflowRunListPage, error) {
	if store == nil || store.pool == nil || !validWorkflowWorkspaceRunListContext(ctx) {
		return WorkflowRunListPage{}, errWorkflowRunStoreContract
	}
	query := `SELECT application_id,sanitized_run_payload FROM prompt_application_run_records
WHERE tenant_ref=$1 AND workspace_id=$2 AND ($3='' OR application_id=$3)
AND ($4::timestamptz IS NULL OR (started_at,run_id,application_id) < ($4,$5,$6))
ORDER BY started_at DESC,run_id DESC,application_id DESC`
	if store.hasActionSafetySnapshot() {
		query = `SELECT application_id,sanitized_run_payload,action_safety_schema_version,action_safety_projection_digest,sanitized_action_safety_snapshot FROM prompt_application_run_records
WHERE tenant_ref=$1 AND workspace_id=$2 AND ($3='' OR application_id=$3)
AND ($4::timestamptz IS NULL OR (started_at,run_id,application_id) < ($4,$5,$6))
ORDER BY started_at DESC,run_id DESC,application_id DESC`
	}
	rows, err := store.pool.Query(ctx.RequestContext, store.statement(query),
		ctx.TenantRef, ctx.WorkspaceID, filter.ApplicationID,
		filter.BeforeTime, filter.BeforeRunID, filter.BeforeApplicationID)
	if err != nil {
		return WorkflowRunListPage{}, errWorkflowRunStoreUnavailable
	}
	defer rows.Close()
	return collectWorkspaceApplicationProjectionRuns(ctx, filter, store.expectedSchema(), func() (string, []byte, actionSafetyStorageSnapshot, bool, error) {
		if !rows.Next() {
			return "", nil, actionSafetyStorageSnapshot{}, false, rows.Err()
		}
		var applicationID string
		var payload, safetyPayload []byte
		var safetySchema, safetyDigest sql.NullString
		var scanErr error
		if store.hasActionSafetySnapshot() {
			scanErr = rows.Scan(&applicationID, &payload, &safetySchema, &safetyDigest, &safetyPayload)
		} else {
			scanErr = rows.Scan(&applicationID, &payload)
		}
		if scanErr != nil {
			return "", nil, actionSafetyStorageSnapshot{}, false, scanErr
		}
		return applicationID, payload, actionSafetyStorageSnapshot{SchemaVersion: safetySchema.String, ProjectionDigest: safetyDigest.String, Payload: safetyPayload}, true, nil
	})
}

func collectPromptApplicationRuns(ctx WorkflowRunContext, filter WorkflowRunListFilter, next func() ([]byte, actionSafetyStorageSnapshot, bool, error)) (WorkflowRunListPage, error) {
	records := make([]WorkflowRunRecord, 0)
	for {
		payload, safetySnapshot, ok, err := next()
		if err != nil {
			return WorkflowRunListPage{}, errWorkflowRunStoreUnavailable
		}
		if !ok {
			break
		}
		record, err := decodeWorkflowRunStorageRecordWithActionSafety(ctx, payload, safetySnapshot)
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

func collectWorkspaceApplicationProjectionRuns(
	ctx WorkflowWorkspaceRunListContext,
	filter WorkflowWorkspaceRunListFilter,
	expectedSchema string,
	next func() (string, []byte, actionSafetyStorageSnapshot, bool, error),
) (WorkflowRunListPage, error) {
	limit := workflowRunStoreListLimit(filter.Limit)
	records := make([]WorkflowRunRecord, 0, limit+1)
	for len(records) <= limit {
		applicationID, payload, safetySnapshot, ok, err := next()
		if err != nil {
			return WorkflowRunListPage{}, errWorkflowRunStoreUnavailable
		}
		if !ok {
			break
		}
		record, decodeErr := decodeWorkflowRunStorageRecordWithActionSafety(WorkflowRunContext{
			RequestContext: ctx.RequestContext, TenantRef: ctx.TenantRef,
			WorkspaceID: ctx.WorkspaceID, ApplicationID: applicationID,
		}, payload, safetySnapshot)
		if decodeErr != nil {
			return WorkflowRunListPage{}, decodeErr
		}
		if record.SchemaVersion != expectedSchema {
			return WorkflowRunListPage{}, errWorkflowRunStoreContract
		}
		if record.ActorRef == ctx.OwnerSubjectRef &&
			workflowRunMatchesFilter(record, filter.WorkflowRunListFilter) {
			records = append(records, record)
		}
	}
	hasMore := len(records) > limit
	if hasMore {
		records = records[:limit]
	}
	return WorkflowRunListPage{Records: records, HasMore: hasMore}, nil
}

func (store *sqlitePromptApplicationRunStore) expectedSchema() string {
	if store.schema != "" {
		return store.schema
	}
	return workflowRunRecordPromptSchemaVersion
}

func (store *sqlitePromptApplicationRunStore) hasActionSafetySnapshot() bool {
	return store.expectedSchema() == agentCopilotRunV7Schema
}

func (store *sqlitePromptApplicationRunStore) statement(query string) string {
	table := store.table
	if table == "" {
		table = "prompt_application_run_records"
	}
	return strings.ReplaceAll(query, "prompt_application_run_records", table)
}

func (store *postgresPromptApplicationRunStore) expectedSchema() string {
	if store.schema != "" {
		return store.schema
	}
	return workflowRunRecordPromptSchemaVersion
}

func (store *postgresPromptApplicationRunStore) hasActionSafetySnapshot() bool {
	return store.expectedSchema() == agentCopilotRunV7Schema
}

func (store *postgresPromptApplicationRunStore) statement(query string) string {
	table := store.table
	if table == "" {
		table = "prompt_application_run_records"
	}
	return strings.ReplaceAll(query, "prompt_application_run_records", table)
}

func applicationProjectionRunAuthorityDigest(record WorkflowRunRecord) string {
	if record.SchemaVersion == agentCopilotRunV7Schema && record.AgentCopilotAuthority != nil {
		return record.AgentCopilotAuthority.AuthorityDigest
	}
	if record.PromptApplication != nil {
		return record.PromptApplication.AuthorityDigest
	}
	return ""
}

var _ workflowRunStore = (*combinedWorkflowRunStore)(nil)
var _ workflowRunStore = (*sqlitePromptApplicationRunStore)(nil)
var _ workflowRunStore = (*postgresPromptApplicationRunStore)(nil)
var _ workflowWorkspaceRunProjection = (*combinedWorkflowRunStore)(nil)
var _ workflowWorkspaceRunProjection = (*sqlitePromptApplicationRunStore)(nil)
var _ workflowWorkspaceRunProjection = (*postgresPromptApplicationRunStore)(nil)
