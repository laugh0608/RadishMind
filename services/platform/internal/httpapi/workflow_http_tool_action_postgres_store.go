package httpapi

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresWorkflowHTTPToolActionStore struct{ pool *pgxpool.Pool }

func newPostgresWorkflowHTTPToolActionStore(pool *pgxpool.Pool) *postgresWorkflowHTTPToolActionStore {
	return &postgresWorkflowHTTPToolActionStore{pool: pool}
}

func (store *postgresWorkflowHTTPToolActionStore) CreatePlan(ctx WorkflowHTTPToolActionContext, plan *WorkflowHTTPToolActionPlan, audit WorkflowHTTPToolExecutionAudit) error {
	if store == nil || store.pool == nil || validateWorkflowHTTPToolActionContext(ctx) != "" || plan == nil || !workflowHTTPToolPlanMatchesContext(*plan, ctx) ||
		plan.RecordVersion != 1 || plan.Status != WorkflowHTTPToolActionStatusPending || !workflowHTTPToolCreateAuditMatchesContext(audit, *plan, ctx) {
		return errWorkflowHTTPToolActionContract
	}
	planJSON, auditJSON, createdAt, expiresAt, occurredAt, err := encodeWorkflowHTTPToolCreateStorage(*plan, audit)
	if err != nil {
		return errWorkflowHTTPToolActionContract
	}
	tx, err := store.pool.Begin(ctx.RequestContext)
	if err != nil {
		return errWorkflowHTTPToolActionUnavailable
	}
	defer func() { _ = tx.Rollback(ctx.RequestContext) }()
	result, err := tx.Exec(ctx.RequestContext, `INSERT INTO workflow_http_tool_action_plans (
	 tenant_ref,workspace_id,application_id,plan_id,schema_version,status,record_version,draft_id,draft_version,
	 node_id,tool_id,tool_version,definition_digest,profile_id,profile_version,profile_digest,target_policy_key,tool_plan_digest,
	 method,credential_policy,timeout_ms,max_response_bytes,max_output_bytes,planned_by_actor_ref,audit_ref,
	 created_at,expires_at,sanitized_action_plan
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28)
	 ON CONFLICT (tenant_ref,workspace_id,application_id,plan_id) DO NOTHING`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, plan.PlanID, plan.SchemaVersion, plan.Status,
		plan.RecordVersion, plan.DraftID, plan.DraftVersion, plan.NodeID, plan.ToolID, plan.ToolVersion, plan.DefinitionDigest,
		plan.ProfileID, plan.ProfileVersion, plan.ProfileDigest, plan.TargetPolicyKey, plan.ToolPlanDigest,
		plan.Method, plan.CredentialPolicy, plan.TimeoutMS, plan.MaxResponseBytes, plan.MaxOutputBytes,
		plan.PlannedByActorRef, plan.AuditRef, createdAt, expiresAt, planJSON)
	if err != nil {
		return errWorkflowHTTPToolActionUnavailable
	}
	if result.RowsAffected() != 1 {
		return errWorkflowHTTPToolActionConflict
	}
	if err := insertPostgresWorkflowHTTPToolAudit(ctx.RequestContext, tx, *plan, audit, occurredAt, auditJSON); err != nil {
		return err
	}
	if err := tx.Commit(ctx.RequestContext); err != nil {
		return errWorkflowHTTPToolActionUnavailable
	}
	return nil
}

func (store *postgresWorkflowHTTPToolActionStore) ReadPlan(ctx WorkflowHTTPToolActionContext, planID string) (WorkflowHTTPToolActionPlan, bool, error) {
	if store == nil || store.pool == nil || validateWorkflowHTTPToolActionContext(ctx) != "" || planID == "" {
		return WorkflowHTTPToolActionPlan{}, false, errWorkflowHTTPToolActionContract
	}
	var payload []byte
	var recordVersion, toolVersion int
	var status, digest string
	err := store.pool.QueryRow(ctx.RequestContext, `SELECT sanitized_action_plan,record_version,status,tool_plan_digest,tool_version
	 FROM workflow_http_tool_action_plans WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND plan_id=$4`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, planID).Scan(&payload, &recordVersion, &status, &digest, &toolVersion)
	if errors.Is(err, pgx.ErrNoRows) {
		return WorkflowHTTPToolActionPlan{}, false, nil
	}
	if err != nil {
		return WorkflowHTTPToolActionPlan{}, false, errWorkflowHTTPToolActionUnavailable
	}
	plan, err := decodeWorkflowHTTPToolActionPlan(payload, ctx, recordVersion, status, digest, toolVersion)
	return plan, err == nil, err
}

func (store *postgresWorkflowHTTPToolActionStore) DecidePlan(ctx WorkflowHTTPToolActionContext, plan *WorkflowHTTPToolActionPlan, decision WorkflowHTTPToolConfirmationDecision, audit WorkflowHTTPToolExecutionAudit) error {
	if store == nil || store.pool == nil || validateWorkflowHTTPToolActionContext(ctx) != "" || plan == nil || !workflowHTTPToolPlanMatchesContext(*plan, ctx) ||
		!workflowHTTPToolDecisionMatchesPlan(decision, *plan) || !workflowHTTPToolDecisionAuditMatchesContext(audit, decision, *plan, ctx) ||
		decision.ExpectedRecordVersion+1 != decision.ResultingRecordVersion {
		return errWorkflowHTTPToolActionContract
	}
	planJSON, decisionJSON, auditJSON, decidedAt, occurredAt, err := encodeWorkflowHTTPToolDecisionStorage(*plan, decision, audit)
	if err != nil {
		return errWorkflowHTTPToolActionContract
	}
	tx, err := store.pool.Begin(ctx.RequestContext)
	if err != nil {
		return errWorkflowHTTPToolActionUnavailable
	}
	defer func() { _ = tx.Rollback(ctx.RequestContext) }()
	result, err := tx.Exec(ctx.RequestContext, `UPDATE workflow_http_tool_action_plans
 SET status=$1,record_version=$2,sanitized_action_plan=$3
	 WHERE tenant_ref=$4 AND workspace_id=$5 AND application_id=$6 AND plan_id=$7 AND record_version=$8 AND tool_plan_digest=$9 AND audit_ref=$10 AND tool_version=$11
	   AND (status='pending'
	     OR (status='deferred' AND $12 IN ('approve','reject','cancel','expire','invalidate'))
	     OR (status='approved' AND $12 IN ('cancel','expire','invalidate')))`,
		plan.Status, plan.RecordVersion, planJSON, ctx.TenantRef, ctx.WorkspaceID,
		ctx.ApplicationID, plan.PlanID, decision.ExpectedRecordVersion, plan.ToolPlanDigest, plan.AuditRef,
		plan.ToolVersion, decision.Outcome)
	if err != nil {
		return errWorkflowHTTPToolActionUnavailable
	}
	if result.RowsAffected() != 1 {
		return errWorkflowHTTPToolActionConflict
	}
	if err := insertPostgresWorkflowHTTPToolAudit(ctx.RequestContext, tx, *plan, audit, occurredAt, auditJSON); err != nil {
		return err
	}
	_, err = tx.Exec(ctx.RequestContext, `INSERT INTO workflow_http_tool_confirmation_decisions (
	 tenant_ref,workspace_id,application_id,plan_id,confirmation_id,schema_version,outcome,draft_id,draft_version,
	 node_id,tool_id,tool_version,tool_plan_digest,expected_record_version,resulting_record_version,decided_by_actor_ref,
	 actor_source,reason_code,decided_at,audit_ref,sanitized_confirmation_decision
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21)`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, plan.PlanID, decision.ConfirmationID,
		decision.SchemaVersion, decision.Outcome, decision.DraftID, decision.DraftVersion, decision.NodeID,
		decision.ToolID, decision.ToolVersion, decision.ToolPlanDigest, decision.ExpectedRecordVersion, decision.ResultingRecordVersion,
		decision.DecidedByActorRef, decision.ActorSource, decision.ReasonCode, decidedAt, decision.AuditRef, decisionJSON)
	if err != nil {
		return errWorkflowHTTPToolActionUnavailable
	}
	if err := tx.Commit(ctx.RequestContext); err != nil {
		return errWorkflowHTTPToolActionUnavailable
	}
	return nil
}

func insertPostgresWorkflowHTTPToolAudit(ctx context.Context, tx pgx.Tx, plan WorkflowHTTPToolActionPlan, audit WorkflowHTTPToolExecutionAudit, occurredAt time.Time, payload []byte) error {
	_, err := tx.Exec(ctx, `INSERT INTO workflow_http_tool_execution_audits (
	 tenant_ref,workspace_id,application_id,plan_id,audit_id,schema_version,event_kind,tool_version,tool_plan_digest,
	 actor_ref,request_id,audit_ref,occurred_at,sanitized_execution_audit
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		plan.TenantRef, plan.WorkspaceID, plan.ApplicationID, plan.PlanID, audit.EventID,
		audit.SchemaVersion, audit.EventKind, audit.ToolVersion, plan.ToolPlanDigest, audit.ActorRef,
		audit.RequestID, audit.AuditRef, occurredAt, payload)
	if err != nil {
		return errWorkflowHTTPToolActionUnavailable
	}
	return nil
}

var _ workflowHTTPToolActionStore = (*postgresWorkflowHTTPToolActionStore)(nil)
