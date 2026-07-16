package httpapi

import (
	"encoding/json"
	"errors"
	"reflect"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresWorkflowHTTPToolExecutionStore struct{ pool *pgxpool.Pool }

func newPostgresWorkflowHTTPToolExecutionStore(pool *pgxpool.Pool) *postgresWorkflowHTTPToolExecutionStore {
	return &postgresWorkflowHTTPToolExecutionStore{pool: pool}
}

func (store *postgresWorkflowHTTPToolExecutionStore) ReadApprovedConfirmation(
	ctx WorkflowHTTPToolActionContext,
	planID string,
	recordVersion int,
) (WorkflowHTTPToolConfirmationDecision, bool, error) {
	if store == nil || store.pool == nil || validateWorkflowHTTPToolActionContext(ctx) != "" ||
		planID == "" || recordVersion < 1 || ctx.RequestContext == nil {
		return WorkflowHTTPToolConfirmationDecision{}, false, errWorkflowHTTPToolExecutionContract
	}
	var payload []byte
	err := store.pool.QueryRow(ctx.RequestContext, `SELECT sanitized_confirmation_decision
	 FROM workflow_http_tool_confirmation_decisions
	 WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND plan_id=$4
	   AND outcome='approve' AND resulting_record_version=$5
	 ORDER BY decided_at DESC, confirmation_id DESC LIMIT 1`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, planID, recordVersion).Scan(&payload)
	if errors.Is(err, pgx.ErrNoRows) {
		return WorkflowHTTPToolConfirmationDecision{}, false, nil
	}
	if err != nil {
		return WorkflowHTTPToolConfirmationDecision{}, false, errWorkflowHTTPToolExecutionUnavailable
	}
	decision, err := decodeWorkflowHTTPToolConfirmationDecision(payload)
	if err != nil || decision.PlanID != planID || decision.ResultingRecordVersion != recordVersion {
		return WorkflowHTTPToolConfirmationDecision{}, false, errWorkflowHTTPToolExecutionContract
	}
	return decision, true, nil
}

func (store *postgresWorkflowHTTPToolExecutionStore) ReadClaimedExecution(
	ctx WorkflowHTTPToolActionContext,
	planID string,
) (WorkflowHTTPToolExecutionAttempt, WorkflowRunRecord, bool, error) {
	if store == nil || store.pool == nil || validateWorkflowHTTPToolActionContext(ctx) != "" ||
		planID == "" || ctx.RequestContext == nil {
		return WorkflowHTTPToolExecutionAttempt{}, WorkflowRunRecord{}, false, errWorkflowHTTPToolExecutionContract
	}
	var attemptPayload, runPayload []byte
	err := store.pool.QueryRow(ctx.RequestContext, `SELECT
	 a.sanitized_execution_attempt,r.sanitized_run_record
	 FROM workflow_http_tool_execution_attempts AS a
	 JOIN workflow_run_records AS r
	   ON r.tenant_ref=a.tenant_ref AND r.workspace_id=a.workspace_id
	  AND r.application_id=a.application_id AND r.run_id=a.run_id
	 WHERE a.tenant_ref=$1 AND a.workspace_id=$2 AND a.application_id=$3 AND a.plan_id=$4
	   AND a.status='claimed' AND r.run_status='running'`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, planID).Scan(&attemptPayload, &runPayload)
	if errors.Is(err, pgx.ErrNoRows) {
		return WorkflowHTTPToolExecutionAttempt{}, WorkflowRunRecord{}, false, nil
	}
	if err != nil {
		return WorkflowHTTPToolExecutionAttempt{}, WorkflowRunRecord{}, false, errWorkflowHTTPToolExecutionUnavailable
	}
	attempt, err := decodeWorkflowHTTPToolExecutionAttempt(attemptPayload)
	if err != nil {
		return WorkflowHTTPToolExecutionAttempt{}, WorkflowRunRecord{}, false, errWorkflowHTTPToolExecutionContract
	}
	run, err := decodeWorkflowRunStorageRecord(workflowRunContextFromToolAction(ctx), runPayload)
	if err != nil || run.PlanID != planID || run.ToolAttempt == nil || !reflect.DeepEqual(*run.ToolAttempt, attempt) {
		return WorkflowHTTPToolExecutionAttempt{}, WorkflowRunRecord{}, false, errWorkflowHTTPToolExecutionContract
	}
	return attempt, run, true, nil
}

func (store *postgresWorkflowHTTPToolExecutionStore) ClaimExecution(
	ctx WorkflowHTTPToolActionContext,
	plan *WorkflowHTTPToolActionPlan,
	confirmation WorkflowHTTPToolConfirmationDecision,
	attempt *WorkflowHTTPToolExecutionAttempt,
	run *WorkflowRunRecord,
	audit WorkflowHTTPToolExecutionAudit,
) error {
	if store == nil || store.pool == nil ||
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
	auditPayload, occurredAt, err := encodeWorkflowHTTPToolAuditStorage(audit)
	if err != nil {
		return errWorkflowHTTPToolExecutionContract
	}
	tx, err := store.pool.Begin(ctx.RequestContext)
	if err != nil {
		return errWorkflowHTTPToolExecutionUnavailable
	}
	defer func() { _ = tx.Rollback(ctx.RequestContext) }()
	if valid, err := postgresWorkflowHTTPToolConfirmationMatches(ctx, tx, confirmation); err != nil {
		return err
	} else if !valid {
		return errWorkflowHTTPToolExecutionConflict
	}
	result, err := tx.Exec(ctx.RequestContext, `UPDATE workflow_http_tool_action_plans
	 SET status='consumed',record_version=$1,sanitized_action_plan=$2
	 WHERE tenant_ref=$3 AND workspace_id=$4 AND application_id=$5 AND plan_id=$6
	   AND status='approved' AND record_version=$7 AND tool_plan_digest=$8 AND tool_version=1`,
		plan.RecordVersion, planPayload, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID,
		plan.PlanID, confirmation.ResultingRecordVersion, plan.ToolPlanDigest)
	if err != nil {
		return errWorkflowHTTPToolExecutionUnavailable
	}
	if result.RowsAffected() != 1 {
		return errWorkflowHTTPToolExecutionConflict
	}
	_, err = tx.Exec(ctx.RequestContext, `INSERT INTO workflow_run_records (
	 tenant_ref,workspace_id,application_id,run_id,draft_id,draft_version,record_version,schema_version,
	 run_status,started_at,completed_at,actor_ref,request_id,audit_ref,failure_code,failure_boundary,
	 selected_provider,selected_model,sanitized_run_record
	) VALUES ($1,$2,$3,$4,$5,$6,1,$7,$8,$9,NULL,$10,$11,$12,$13,$14,$15,$16,$17)`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, claimedRun.RunID, claimedRun.DraftID,
		claimedRun.DraftVersion, claimedRun.SchemaVersion, claimedRun.Status, startedAt, claimedRun.ActorRef,
		claimedRun.RequestID, claimedRun.AuditRef, claimedRun.FailureCode,
		workflowRunRecordFailureBoundary(claimedRun), claimedRun.SelectedProvider, claimedRun.SelectedModel, runPayload)
	if err != nil {
		return errWorkflowHTTPToolExecutionUnavailable
	}
	_, err = tx.Exec(ctx.RequestContext, `INSERT INTO workflow_http_tool_execution_attempts (
	 tenant_ref,workspace_id,application_id,plan_id,confirmation_id,attempt_id,run_id,status,tool_plan_digest,
	 claimed_at,completed_at,failure_code,sanitized_execution_attempt
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,NULL,'',$11)`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, plan.PlanID, confirmation.ConfirmationID,
		attempt.AttemptID, run.RunID, attempt.Status, attempt.ToolPlanDigest, claimedAt, attemptPayload)
	if err != nil {
		return errWorkflowHTTPToolExecutionUnavailable
	}
	if err := insertPostgresWorkflowHTTPToolAudit(ctx.RequestContext, tx, *plan, audit, occurredAt, auditPayload); err != nil {
		return errWorkflowHTTPToolExecutionUnavailable
	}
	if err := tx.Commit(ctx.RequestContext); err != nil {
		return errWorkflowHTTPToolExecutionUnavailable
	}
	run.RecordVersion = 1
	return nil
}

func (store *postgresWorkflowHTTPToolExecutionStore) CompleteExecution(
	ctx WorkflowHTTPToolActionContext,
	attempt *WorkflowHTTPToolExecutionAttempt,
	run *WorkflowRunRecord,
	audit WorkflowHTTPToolExecutionAudit,
) error {
	if store == nil || store.pool == nil || !validWorkflowHTTPToolExecutionCompletion(ctx, attempt, run, audit) {
		return errWorkflowHTTPToolExecutionContract
	}
	attemptPayload, _, completedAt, err := encodeWorkflowHTTPToolExecutionAttempt(*attempt)
	if err != nil || completedAt == nil {
		return errWorkflowHTTPToolExecutionContract
	}
	completedRun := cloneWorkflowRunRecord(*run)
	completedRun.RecordVersion++
	runPayload, _, runCompletedAt, err := encodeWorkflowRunStorageRecord(completedRun)
	if err != nil || runCompletedAt == nil {
		return errWorkflowHTTPToolExecutionContract
	}
	auditPayload, occurredAt, err := encodeWorkflowHTTPToolAuditStorage(audit)
	if err != nil {
		return errWorkflowHTTPToolExecutionContract
	}
	tx, err := store.pool.Begin(ctx.RequestContext)
	if err != nil {
		return errWorkflowHTTPToolExecutionUnavailable
	}
	defer func() { _ = tx.Rollback(ctx.RequestContext) }()
	result, err := tx.Exec(ctx.RequestContext, `UPDATE workflow_http_tool_execution_attempts
	 SET status=$1,completed_at=$2,failure_code=$3,sanitized_execution_attempt=$4
	 WHERE tenant_ref=$5 AND workspace_id=$6 AND application_id=$7 AND plan_id=$8 AND attempt_id=$9 AND status='claimed'`,
		attempt.Status, completedAt, attempt.FailureCode, attemptPayload,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, run.PlanID, attempt.AttemptID)
	if err != nil {
		return errWorkflowHTTPToolExecutionUnavailable
	}
	if result.RowsAffected() != 1 {
		return errWorkflowHTTPToolExecutionConflict
	}
	result, err = tx.Exec(ctx.RequestContext, `UPDATE workflow_run_records SET
	 draft_id=$1,draft_version=$2,record_version=record_version+1,schema_version=$3,run_status=$4,
	 completed_at=$5,actor_ref=$6,request_id=$7,audit_ref=$8,failure_code=$9,failure_boundary=$10,
	 selected_provider=$11,selected_model=$12,sanitized_run_record=$13
	 WHERE tenant_ref=$14 AND workspace_id=$15 AND application_id=$16 AND run_id=$17
	   AND record_version=$18 AND run_status='running'`,
		completedRun.DraftID, completedRun.DraftVersion, completedRun.SchemaVersion, completedRun.Status,
		runCompletedAt, completedRun.ActorRef, completedRun.RequestID, completedRun.AuditRef,
		completedRun.FailureCode, workflowRunRecordFailureBoundary(completedRun), completedRun.SelectedProvider,
		completedRun.SelectedModel, runPayload, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID,
		completedRun.RunID, run.RecordVersion)
	if err != nil {
		return errWorkflowHTTPToolExecutionUnavailable
	}
	if result.RowsAffected() != 1 {
		return errWorkflowHTTPToolExecutionConflict
	}
	plan := WorkflowHTTPToolActionPlan{TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, PlanID: run.PlanID, ToolPlanDigest: attempt.ToolPlanDigest}
	if err := insertPostgresWorkflowHTTPToolAudit(ctx.RequestContext, tx, plan, audit, occurredAt, auditPayload); err != nil {
		return errWorkflowHTTPToolExecutionUnavailable
	}
	if err := tx.Commit(ctx.RequestContext); err != nil {
		return errWorkflowHTTPToolExecutionUnavailable
	}
	run.RecordVersion = completedRun.RecordVersion
	return nil
}

func postgresWorkflowHTTPToolConfirmationMatches(
	ctx WorkflowHTTPToolActionContext,
	tx pgx.Tx,
	expected WorkflowHTTPToolConfirmationDecision,
) (bool, error) {
	var payload []byte
	err := tx.QueryRow(ctx.RequestContext, `SELECT sanitized_confirmation_decision
	 FROM workflow_http_tool_confirmation_decisions
	 WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND plan_id=$4 AND confirmation_id=$5
	   AND outcome='approve' AND resulting_record_version=$6`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, expected.PlanID, expected.ConfirmationID,
		expected.ResultingRecordVersion).Scan(&payload)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, errWorkflowHTTPToolExecutionUnavailable
	}
	stored, err := decodeWorkflowHTTPToolConfirmationDecision(payload)
	if err != nil {
		return false, errWorkflowHTTPToolExecutionContract
	}
	return reflect.DeepEqual(stored, expected), nil
}

var _ workflowHTTPToolExecutionStore = (*postgresWorkflowHTTPToolExecutionStore)(nil)
