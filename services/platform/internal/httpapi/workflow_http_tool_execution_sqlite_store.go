package httpapi

import (
	"database/sql"
	"encoding/json"
	"errors"
	"reflect"

	sqliteworkflowrunmigrations "radishmind.local/services/platform/migrations/sqlite/workflow_runs"
)

type sqliteWorkflowHTTPToolExecutionStore struct{ database *sql.DB }

func newSQLiteWorkflowHTTPToolExecutionStore(database *sql.DB) *sqliteWorkflowHTTPToolExecutionStore {
	return &sqliteWorkflowHTTPToolExecutionStore{database: database}
}

func (store *sqliteWorkflowHTTPToolExecutionStore) ReadApprovedConfirmation(
	ctx WorkflowHTTPToolActionContext,
	planID string,
	recordVersion int,
) (WorkflowHTTPToolConfirmationDecision, bool, error) {
	if store == nil || store.database == nil || validateWorkflowHTTPToolActionContext(ctx) != "" ||
		planID == "" || recordVersion < 1 || ctx.RequestContext == nil {
		return WorkflowHTTPToolConfirmationDecision{}, false, errWorkflowHTTPToolExecutionContract
	}
	var payload string
	err := store.database.QueryRowContext(ctx.RequestContext, `SELECT sanitized_confirmation_decision
	 FROM workflow_http_tool_confirmation_decisions
	 WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND plan_id=?
	   AND outcome='approve' AND resulting_record_version=?
	 ORDER BY decided_at_unix_nano DESC, confirmation_id DESC LIMIT 1`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, planID, recordVersion).Scan(&payload)
	if errors.Is(err, sql.ErrNoRows) {
		return WorkflowHTTPToolConfirmationDecision{}, false, nil
	}
	if err != nil {
		return WorkflowHTTPToolConfirmationDecision{}, false, errWorkflowHTTPToolExecutionUnavailable
	}
	decision, err := decodeWorkflowHTTPToolConfirmationDecision([]byte(payload))
	if err != nil || decision.PlanID != planID || decision.ResultingRecordVersion != recordVersion {
		return WorkflowHTTPToolConfirmationDecision{}, false, errWorkflowHTTPToolExecutionContract
	}
	return decision, true, nil
}

func (store *sqliteWorkflowHTTPToolExecutionStore) ReadClaimedExecution(
	ctx WorkflowHTTPToolActionContext,
	planID string,
) (WorkflowHTTPToolExecutionAttempt, WorkflowRunRecord, bool, error) {
	if store == nil || store.database == nil || validateWorkflowHTTPToolActionContext(ctx) != "" ||
		planID == "" || ctx.RequestContext == nil {
		return WorkflowHTTPToolExecutionAttempt{}, WorkflowRunRecord{}, false, errWorkflowHTTPToolExecutionContract
	}
	var attemptPayload, runPayload string
	err := store.database.QueryRowContext(ctx.RequestContext, `SELECT
	 a.sanitized_execution_attempt,r.sanitized_run_record
	 FROM workflow_http_tool_execution_attempts AS a
	 JOIN workflow_run_records AS r
	   ON r.tenant_ref=a.tenant_ref AND r.workspace_id=a.workspace_id
	  AND r.application_id=a.application_id AND r.run_id=a.run_id
	 WHERE a.tenant_ref=? AND a.workspace_id=? AND a.application_id=? AND a.plan_id=?
	   AND a.status='claimed' AND r.run_status='running'`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, planID).Scan(&attemptPayload, &runPayload)
	if errors.Is(err, sql.ErrNoRows) {
		return WorkflowHTTPToolExecutionAttempt{}, WorkflowRunRecord{}, false, nil
	}
	if err != nil {
		return WorkflowHTTPToolExecutionAttempt{}, WorkflowRunRecord{}, false, errWorkflowHTTPToolExecutionUnavailable
	}
	attempt, err := decodeWorkflowHTTPToolExecutionAttempt([]byte(attemptPayload))
	if err != nil {
		return WorkflowHTTPToolExecutionAttempt{}, WorkflowRunRecord{}, false, errWorkflowHTTPToolExecutionContract
	}
	run, err := decodeWorkflowRunStorageRecord(workflowRunContextFromToolAction(ctx), []byte(runPayload))
	if err != nil || run.PlanID != planID || run.ToolAttempt == nil || !reflect.DeepEqual(*run.ToolAttempt, attempt) {
		return WorkflowHTTPToolExecutionAttempt{}, WorkflowRunRecord{}, false, errWorkflowHTTPToolExecutionContract
	}
	return attempt, run, true, nil
}

func (store *sqliteWorkflowHTTPToolExecutionStore) ClaimExecution(
	ctx WorkflowHTTPToolActionContext,
	plan *WorkflowHTTPToolActionPlan,
	confirmation WorkflowHTTPToolConfirmationDecision,
	attempt *WorkflowHTTPToolExecutionAttempt,
	run *WorkflowRunRecord,
	audit WorkflowHTTPToolExecutionAudit,
) error {
	if store == nil || store.database == nil ||
		!validWorkflowHTTPToolExecutionClaim(ctx, plan, confirmation, attempt, run, audit) {
		return errWorkflowHTTPToolExecutionContract
	}
	planPayload, err := json.Marshal(plan)
	if err != nil {
		return errWorkflowHTTPToolExecutionContract
	}
	attemptPayload, claimedAt, _, err := encodeWorkflowHTTPToolExecutionAttempt(*attempt)
	if err != nil {
		return err
	}
	claimedRun := cloneWorkflowRunRecord(*run)
	claimedRun.RecordVersion = 1
	runPayload, startedAt, completedAt, err := encodeWorkflowRunStorageRecord(claimedRun)
	if err != nil || completedAt != nil {
		return errWorkflowHTTPToolExecutionContract
	}
	startedAtUnixNano, err := workflowRunUnixNano(startedAt)
	if err != nil {
		return errWorkflowHTTPToolExecutionContract
	}
	auditPayload, occurredAt, err := encodeWorkflowHTTPToolAuditStorage(audit)
	if err != nil {
		return errWorkflowHTTPToolExecutionContract
	}
	tx, err := store.database.BeginTx(ctx.RequestContext, nil)
	if err != nil {
		return errWorkflowHTTPToolExecutionUnavailable
	}
	defer func() { _ = tx.Rollback() }()
	if valid, err := sqliteWorkflowHTTPToolConfirmationMatches(ctx, tx, confirmation); err != nil {
		return err
	} else if !valid {
		return errWorkflowHTTPToolExecutionConflict
	}
	result, err := tx.ExecContext(ctx.RequestContext, `UPDATE workflow_http_tool_action_plans
	 SET status='consumed',record_version=?,sanitized_action_plan=?
	 WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND plan_id=?
	   AND status='approved' AND record_version=? AND tool_plan_digest=? AND tool_version=1`,
		plan.RecordVersion, string(planPayload), ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID,
		plan.PlanID, confirmation.ResultingRecordVersion, plan.ToolPlanDigest)
	if err != nil {
		return errWorkflowHTTPToolExecutionUnavailable
	}
	if updated, rowsErr := result.RowsAffected(); rowsErr != nil || updated != 1 {
		return errWorkflowHTTPToolExecutionConflict
	}
	_, err = tx.ExecContext(ctx.RequestContext, `INSERT INTO workflow_run_records (
	 tenant_ref,workspace_id,application_id,run_id,draft_id,draft_version,record_version,store_schema_version,
	 schema_version,run_status,started_at_unix_nano,completed_at_unix_nano,actor_ref,request_id,audit_ref,
	 failure_code,failure_boundary,selected_provider,selected_model,sanitized_run_record
	) VALUES (?,?,?,?,?,?,1,?,?,?,?,NULL,?,?,?,?,?,?,?,?)`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, claimedRun.RunID, claimedRun.DraftID, claimedRun.DraftVersion,
		sqliteworkflowrunmigrations.RunRecordStoreSchemaVersion, claimedRun.SchemaVersion, claimedRun.Status,
		startedAtUnixNano, claimedRun.ActorRef, claimedRun.RequestID, claimedRun.AuditRef, claimedRun.FailureCode,
		workflowRunRecordFailureBoundary(claimedRun), claimedRun.SelectedProvider, claimedRun.SelectedModel, string(runPayload))
	if err != nil {
		return errWorkflowHTTPToolExecutionUnavailable
	}
	_, err = tx.ExecContext(ctx.RequestContext, `INSERT INTO workflow_http_tool_execution_attempts (
	 tenant_ref,workspace_id,application_id,plan_id,confirmation_id,attempt_id,run_id,status,tool_plan_digest,
	 claimed_at_unix_nano,completed_at_unix_nano,failure_code,sanitized_execution_attempt
	) VALUES (?,?,?,?,?,?,?,?,?,?,NULL,'',?)`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, plan.PlanID, confirmation.ConfirmationID,
		attempt.AttemptID, run.RunID, attempt.Status, attempt.ToolPlanDigest, claimedAt.UnixNano(), string(attemptPayload))
	if err != nil {
		return errWorkflowHTTPToolExecutionUnavailable
	}
	if err := insertSQLiteWorkflowHTTPToolAudit(ctx.RequestContext, tx, *plan, audit, occurredAt, auditPayload); err != nil {
		return errWorkflowHTTPToolExecutionUnavailable
	}
	if err := tx.Commit(); err != nil {
		return errWorkflowHTTPToolExecutionUnavailable
	}
	run.RecordVersion = 1
	return nil
}

func (store *sqliteWorkflowHTTPToolExecutionStore) CompleteExecution(
	ctx WorkflowHTTPToolActionContext,
	attempt *WorkflowHTTPToolExecutionAttempt,
	run *WorkflowRunRecord,
	audit WorkflowHTTPToolExecutionAudit,
) error {
	if store == nil || store.database == nil || !validWorkflowHTTPToolExecutionCompletion(ctx, attempt, run, audit) {
		return errWorkflowHTTPToolExecutionContract
	}
	attemptPayload, _, completedAt, err := encodeWorkflowHTTPToolExecutionAttempt(*attempt)
	if err != nil || completedAt == nil {
		return errWorkflowHTTPToolExecutionContract
	}
	completedRun := cloneWorkflowRunRecord(*run)
	completedRun.RecordVersion++
	runPayload, startedAt, runCompletedAt, err := encodeWorkflowRunStorageRecord(completedRun)
	if err != nil || runCompletedAt == nil {
		return errWorkflowHTTPToolExecutionContract
	}
	startedAtUnixNano, err := workflowRunUnixNano(startedAt)
	if err != nil {
		return errWorkflowHTTPToolExecutionContract
	}
	completedAtUnixNano, err := workflowRunUnixNano(*runCompletedAt)
	if err != nil {
		return errWorkflowHTTPToolExecutionContract
	}
	auditPayload, occurredAt, err := encodeWorkflowHTTPToolAuditStorage(audit)
	if err != nil {
		return errWorkflowHTTPToolExecutionContract
	}
	tx, err := store.database.BeginTx(ctx.RequestContext, nil)
	if err != nil {
		return errWorkflowHTTPToolExecutionUnavailable
	}
	defer func() { _ = tx.Rollback() }()
	result, err := tx.ExecContext(ctx.RequestContext, `UPDATE workflow_http_tool_execution_attempts
	 SET status=?,completed_at_unix_nano=?,failure_code=?,sanitized_execution_attempt=?
	 WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND plan_id=? AND attempt_id=? AND status='claimed'`,
		attempt.Status, completedAt.UnixNano(), attempt.FailureCode, string(attemptPayload),
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, run.PlanID, attempt.AttemptID)
	if err != nil {
		return errWorkflowHTTPToolExecutionUnavailable
	}
	if updated, rowsErr := result.RowsAffected(); rowsErr != nil || updated != 1 {
		return errWorkflowHTTPToolExecutionConflict
	}
	result, err = tx.ExecContext(ctx.RequestContext, `UPDATE workflow_run_records SET
	 draft_id=?,draft_version=?,record_version=record_version+1,schema_version=?,run_status=?,
	 completed_at_unix_nano=?,actor_ref=?,request_id=?,audit_ref=?,failure_code=?,failure_boundary=?,
	 selected_provider=?,selected_model=?,sanitized_run_record=?
	 WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND run_id=?
	   AND record_version=? AND run_status='running' AND started_at_unix_nano=?`,
		completedRun.DraftID, completedRun.DraftVersion, completedRun.SchemaVersion, completedRun.Status,
		completedAtUnixNano, completedRun.ActorRef, completedRun.RequestID, completedRun.AuditRef,
		completedRun.FailureCode, workflowRunRecordFailureBoundary(completedRun), completedRun.SelectedProvider,
		completedRun.SelectedModel, string(runPayload), ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID,
		completedRun.RunID, run.RecordVersion, startedAtUnixNano)
	if err != nil {
		return errWorkflowHTTPToolExecutionUnavailable
	}
	if updated, rowsErr := result.RowsAffected(); rowsErr != nil || updated != 1 {
		return errWorkflowHTTPToolExecutionConflict
	}
	plan := WorkflowHTTPToolActionPlan{TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, PlanID: run.PlanID, ToolPlanDigest: attempt.ToolPlanDigest}
	if err := insertSQLiteWorkflowHTTPToolAudit(ctx.RequestContext, tx, plan, audit, occurredAt, auditPayload); err != nil {
		return errWorkflowHTTPToolExecutionUnavailable
	}
	if err := tx.Commit(); err != nil {
		return errWorkflowHTTPToolExecutionUnavailable
	}
	run.RecordVersion = completedRun.RecordVersion
	return nil
}

func sqliteWorkflowHTTPToolConfirmationMatches(
	ctx WorkflowHTTPToolActionContext,
	tx *sql.Tx,
	expected WorkflowHTTPToolConfirmationDecision,
) (bool, error) {
	var payload string
	err := tx.QueryRowContext(ctx.RequestContext, `SELECT sanitized_confirmation_decision
	 FROM workflow_http_tool_confirmation_decisions
	 WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND plan_id=? AND confirmation_id=?
	   AND outcome='approve' AND resulting_record_version=?`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, expected.PlanID, expected.ConfirmationID,
		expected.ResultingRecordVersion).Scan(&payload)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, errWorkflowHTTPToolExecutionUnavailable
	}
	stored, err := decodeWorkflowHTTPToolConfirmationDecision([]byte(payload))
	if err != nil {
		return false, errWorkflowHTTPToolExecutionContract
	}
	return reflect.DeepEqual(stored, expected), nil
}

var _ workflowHTTPToolExecutionStore = (*sqliteWorkflowHTTPToolExecutionStore)(nil)
