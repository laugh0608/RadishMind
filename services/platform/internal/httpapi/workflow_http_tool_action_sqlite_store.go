package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

type sqliteWorkflowHTTPToolActionStore struct{ database *sql.DB }

func newSQLiteWorkflowHTTPToolActionStore(database *sql.DB) *sqliteWorkflowHTTPToolActionStore {
	return &sqliteWorkflowHTTPToolActionStore{database: database}
}

func (store *sqliteWorkflowHTTPToolActionStore) CreatePlan(ctx WorkflowHTTPToolActionContext, plan *WorkflowHTTPToolActionPlan, audit WorkflowHTTPToolExecutionAudit) error {
	if store == nil || store.database == nil || validateWorkflowHTTPToolActionContext(ctx) != "" || plan == nil || !workflowHTTPToolPlanMatchesContext(*plan, ctx) ||
		plan.RecordVersion != 1 || plan.Status != WorkflowHTTPToolActionStatusPending || !workflowHTTPToolCreateAuditMatchesContext(audit, *plan, ctx) {
		return errWorkflowHTTPToolActionContract
	}
	source, valid := workflowHTTPToolStorageSourceForPlan(*plan)
	if !valid {
		return errWorkflowHTTPToolActionContract
	}
	planJSON, auditJSON, createdAt, expiresAt, occurredAt, err := encodeWorkflowHTTPToolCreateStorage(*plan, audit)
	if err != nil {
		return errWorkflowHTTPToolActionContract
	}
	safety, err := encodeActionSafetyPlanSnapshot(plan.ActionSafety)
	if err != nil {
		return errWorkflowHTTPToolActionContract
	}
	safetySchema, safetyDigest, safetyPayload := safety.sqliteColumnValues()
	tx, err := store.database.BeginTx(ctx.RequestContext, nil)
	if err != nil {
		return errWorkflowHTTPToolActionUnavailable
	}
	defer func() { _ = tx.Rollback() }()
	result, err := tx.ExecContext(ctx.RequestContext, `INSERT INTO workflow_http_tool_action_plans (
	 tenant_ref,workspace_id,application_id,plan_id,schema_version,status,record_version,source_kind,draft_id,draft_version,
	 workflow_definition_id,workflow_definition_version,workflow_definition_digest,activation_pointer_version,
	 node_id,tool_id,tool_version,definition_digest,profile_id,profile_version,profile_digest,target_policy_key,tool_plan_digest,
		 method,credential_policy,timeout_ms,max_response_bytes,max_output_bytes,planned_by_actor_ref,audit_ref,
		 created_at_unix_nano,expires_at_unix_nano,sanitized_action_plan,
		 action_safety_schema_version,action_safety_projection_digest,sanitized_action_safety_snapshot
	) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
	 ON CONFLICT (tenant_ref,workspace_id,application_id,plan_id) DO NOTHING`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, plan.PlanID, plan.SchemaVersion, plan.Status,
		plan.RecordVersion, source.SourceKind, nullableWorkflowHTTPToolString(source.DraftID), nullableWorkflowHTTPToolInt(source.DraftVersion),
		nullableWorkflowHTTPToolString(source.WorkflowDefinitionID), nullableWorkflowHTTPToolInt(source.WorkflowDefinitionVersion),
		nullableWorkflowHTTPToolString(source.WorkflowDefinitionDigest), nullableWorkflowHTTPToolInt(source.ActivationPointerVersion),
		plan.NodeID, plan.ToolID, plan.ToolVersion, plan.DefinitionDigest,
		plan.ProfileID, plan.ProfileVersion, plan.ProfileDigest, plan.TargetPolicyKey, plan.ToolPlanDigest,
		plan.Method, plan.CredentialPolicy, plan.TimeoutMS, plan.MaxResponseBytes, plan.MaxOutputBytes,
		plan.PlannedByActorRef, plan.AuditRef, createdAt.UnixNano(), expiresAt.UnixNano(), string(planJSON),
		safetySchema, safetyDigest, safetyPayload)
	if err != nil {
		return errWorkflowHTTPToolActionUnavailable
	}
	inserted, err := result.RowsAffected()
	if err != nil || inserted != 1 {
		return errWorkflowHTTPToolActionConflict
	}
	if err := insertSQLiteWorkflowHTTPToolAudit(ctx.RequestContext, tx, *plan, audit, occurredAt, auditJSON); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return errWorkflowHTTPToolActionUnavailable
	}
	return nil
}

func (store *sqliteWorkflowHTTPToolActionStore) ReadPlan(ctx WorkflowHTTPToolActionContext, planID string) (WorkflowHTTPToolActionPlan, bool, error) {
	if store == nil || store.database == nil || validateWorkflowHTTPToolActionContext(ctx) != "" || planID == "" {
		return WorkflowHTTPToolActionPlan{}, false, errWorkflowHTTPToolActionContract
	}
	var payload string
	var safetyPayload []byte
	var safetySchema, safetyDigest sql.NullString
	var recordVersion, toolVersion int
	var status, digest string
	var source workflowHTTPToolStorageSource
	err := store.database.QueryRowContext(ctx.RequestContext, `SELECT sanitized_action_plan,record_version,status,tool_plan_digest,tool_version,
	 source_kind,coalesce(draft_id,''),coalesce(draft_version,0),coalesce(workflow_definition_id,''),
	 coalesce(workflow_definition_version,0),coalesce(workflow_definition_digest,''),coalesce(activation_pointer_version,0),
	 action_safety_schema_version,action_safety_projection_digest,sanitized_action_safety_snapshot
	 FROM workflow_http_tool_action_plans WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND plan_id=?`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, planID).Scan(
		&payload, &recordVersion, &status, &digest, &toolVersion, &source.SourceKind, &source.DraftID, &source.DraftVersion,
		&source.WorkflowDefinitionID, &source.WorkflowDefinitionVersion, &source.WorkflowDefinitionDigest, &source.ActivationPointerVersion,
		&safetySchema, &safetyDigest, &safetyPayload,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return WorkflowHTTPToolActionPlan{}, false, nil
	}
	if err != nil {
		return WorkflowHTTPToolActionPlan{}, false, errWorkflowHTTPToolActionUnavailable
	}
	plan, err := decodeWorkflowHTTPToolActionPlan([]byte(payload), ctx, recordVersion, status, digest, toolVersion, source,
		actionSafetyStorageSnapshot{SchemaVersion: safetySchema.String, ProjectionDigest: safetyDigest.String, Payload: safetyPayload})
	return plan, err == nil, err
}

func (store *sqliteWorkflowHTTPToolActionStore) DecidePlan(ctx WorkflowHTTPToolActionContext, plan *WorkflowHTTPToolActionPlan, decision WorkflowHTTPToolConfirmationDecision, audit WorkflowHTTPToolExecutionAudit) error {
	if store == nil || store.database == nil || validateWorkflowHTTPToolActionContext(ctx) != "" || plan == nil || !workflowHTTPToolPlanMatchesContext(*plan, ctx) ||
		!workflowHTTPToolDecisionMatchesPlan(decision, *plan) || !workflowHTTPToolDecisionAuditMatchesContext(audit, decision, *plan, ctx) ||
		decision.ExpectedRecordVersion+1 != decision.ResultingRecordVersion {
		return errWorkflowHTTPToolActionContract
	}
	source, valid := workflowHTTPToolStorageSourceForPlan(*plan)
	if !valid {
		return errWorkflowHTTPToolActionContract
	}
	planJSON, decisionJSON, auditJSON, decidedAt, occurredAt, err := encodeWorkflowHTTPToolDecisionStorage(*plan, decision, audit)
	if err != nil {
		return errWorkflowHTTPToolActionContract
	}
	safety, err := encodeActionSafetyPlanSnapshot(plan.ActionSafety)
	if err != nil {
		return errWorkflowHTTPToolActionContract
	}
	_, safetyDigest, _ := safety.sqliteColumnValues()
	tx, err := store.database.BeginTx(ctx.RequestContext, nil)
	if err != nil {
		return errWorkflowHTTPToolActionUnavailable
	}
	defer func() { _ = tx.Rollback() }()
	result, err := tx.ExecContext(ctx.RequestContext, `UPDATE workflow_http_tool_action_plans
 SET status=?,record_version=?,sanitized_action_plan=?
	 WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND plan_id=? AND record_version=? AND tool_plan_digest=? AND audit_ref=? AND tool_version=?
	   AND action_safety_projection_digest IS ?
	   AND (status='pending'
	     OR (status='deferred' AND ? IN ('approve','reject','cancel','expire','invalidate'))
	     OR (status='approved' AND ? IN ('cancel','expire','invalidate')))`,
		plan.Status, plan.RecordVersion, string(planJSON), ctx.TenantRef, ctx.WorkspaceID,
		ctx.ApplicationID, plan.PlanID, decision.ExpectedRecordVersion, plan.ToolPlanDigest, plan.AuditRef,
		plan.ToolVersion, safetyDigest, decision.Outcome, decision.Outcome)
	if err != nil {
		return errWorkflowHTTPToolActionUnavailable
	}
	updated, err := result.RowsAffected()
	if err != nil || updated != 1 {
		return errWorkflowHTTPToolActionConflict
	}
	if err := insertSQLiteWorkflowHTTPToolAudit(ctx.RequestContext, tx, *plan, audit, occurredAt, auditJSON); err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx.RequestContext, `INSERT INTO workflow_http_tool_confirmation_decisions (
	 tenant_ref,workspace_id,application_id,plan_id,confirmation_id,schema_version,outcome,source_kind,draft_id,draft_version,
	 workflow_definition_id,workflow_definition_version,workflow_definition_digest,activation_pointer_version,
	 node_id,tool_id,tool_version,tool_plan_digest,expected_record_version,resulting_record_version,decided_by_actor_ref,
	 actor_source,reason_code,decided_at_unix_nano,audit_ref,sanitized_confirmation_decision
	) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, plan.PlanID, decision.ConfirmationID,
		decision.SchemaVersion, decision.Outcome, source.SourceKind, nullableWorkflowHTTPToolString(source.DraftID), nullableWorkflowHTTPToolInt(source.DraftVersion),
		nullableWorkflowHTTPToolString(source.WorkflowDefinitionID), nullableWorkflowHTTPToolInt(source.WorkflowDefinitionVersion),
		nullableWorkflowHTTPToolString(source.WorkflowDefinitionDigest), nullableWorkflowHTTPToolInt(source.ActivationPointerVersion), decision.NodeID,
		decision.ToolID, decision.ToolVersion, decision.ToolPlanDigest, decision.ExpectedRecordVersion, decision.ResultingRecordVersion,
		decision.DecidedByActorRef, decision.ActorSource, decision.ReasonCode, decidedAt.UnixNano(),
		decision.AuditRef, string(decisionJSON))
	if err != nil {
		return errWorkflowHTTPToolActionUnavailable
	}
	if err := tx.Commit(); err != nil {
		return errWorkflowHTTPToolActionUnavailable
	}
	return nil
}

func insertSQLiteWorkflowHTTPToolAudit(ctx context.Context, tx *sql.Tx, plan WorkflowHTTPToolActionPlan, audit WorkflowHTTPToolExecutionAudit, occurredAt time.Time, payload []byte) error {
	source, valid := workflowHTTPToolStorageSourceForAudit(audit)
	if !valid {
		return errWorkflowHTTPToolActionContract
	}
	_, err := tx.ExecContext(ctx, `INSERT INTO workflow_http_tool_execution_audits (
	 tenant_ref,workspace_id,application_id,plan_id,audit_id,schema_version,event_kind,source_kind,draft_id,draft_version,
	 workflow_definition_id,workflow_definition_version,workflow_definition_digest,activation_pointer_version,tool_version,tool_plan_digest,
	 actor_ref,request_id,audit_ref,occurred_at_unix_nano,sanitized_execution_audit
	) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		plan.TenantRef, plan.WorkspaceID, plan.ApplicationID, plan.PlanID, audit.EventID,
		audit.SchemaVersion, audit.EventKind, source.SourceKind, nullableWorkflowHTTPToolString(source.DraftID), nullableWorkflowHTTPToolInt(source.DraftVersion),
		nullableWorkflowHTTPToolString(source.WorkflowDefinitionID), nullableWorkflowHTTPToolInt(source.WorkflowDefinitionVersion),
		nullableWorkflowHTTPToolString(source.WorkflowDefinitionDigest), nullableWorkflowHTTPToolInt(source.ActivationPointerVersion),
		audit.ToolVersion, plan.ToolPlanDigest, audit.ActorRef, audit.RequestID,
		audit.AuditRef, occurredAt.UnixNano(), string(payload))
	if err != nil {
		return errWorkflowHTTPToolActionUnavailable
	}
	return nil
}

func encodeWorkflowHTTPToolCreateStorage(plan WorkflowHTTPToolActionPlan, audit WorkflowHTTPToolExecutionAudit) ([]byte, []byte, time.Time, time.Time, time.Time, error) {
	planJSON, err := json.Marshal(plan)
	if err != nil {
		return nil, nil, time.Time{}, time.Time{}, time.Time{}, err
	}
	auditJSON, occurredAt, err := encodeWorkflowHTTPToolAuditStorage(audit)
	if err != nil {
		return nil, nil, time.Time{}, time.Time{}, time.Time{}, err
	}
	createdAt, err := time.Parse(time.RFC3339Nano, plan.CreatedAt)
	if err != nil {
		return nil, nil, time.Time{}, time.Time{}, time.Time{}, err
	}
	expiresAt, err := time.Parse(time.RFC3339Nano, plan.ExpiresAt)
	return planJSON, auditJSON, createdAt, expiresAt, occurredAt, err
}

func encodeWorkflowHTTPToolDecisionStorage(plan WorkflowHTTPToolActionPlan, decision WorkflowHTTPToolConfirmationDecision, audit WorkflowHTTPToolExecutionAudit) ([]byte, []byte, []byte, time.Time, time.Time, error) {
	planJSON, err := json.Marshal(plan)
	if err != nil {
		return nil, nil, nil, time.Time{}, time.Time{}, err
	}
	decisionJSON, err := json.Marshal(decision)
	if err != nil {
		return nil, nil, nil, time.Time{}, time.Time{}, err
	}
	auditJSON, occurredAt, err := encodeWorkflowHTTPToolAuditStorage(audit)
	if err != nil {
		return nil, nil, nil, time.Time{}, time.Time{}, err
	}
	decidedAt, err := time.Parse(time.RFC3339Nano, decision.DecidedAt)
	return planJSON, decisionJSON, auditJSON, decidedAt, occurredAt, err
}

func encodeWorkflowHTTPToolAuditStorage(audit WorkflowHTTPToolExecutionAudit) ([]byte, time.Time, error) {
	payload, err := json.Marshal(audit)
	if err != nil {
		return nil, time.Time{}, err
	}
	occurredAt, err := time.Parse(time.RFC3339Nano, audit.OccurredAt)
	return payload, occurredAt, err
}

func decodeWorkflowHTTPToolActionPlan(payload []byte, ctx WorkflowHTTPToolActionContext, recordVersion int, status, digest string, toolVersion int, source workflowHTTPToolStorageSource, safetySnapshot actionSafetyStorageSnapshot) (WorkflowHTTPToolActionPlan, error) {
	var plan WorkflowHTTPToolActionPlan
	if err := json.Unmarshal(payload, &plan); err != nil {
		return WorkflowHTTPToolActionPlan{}, errWorkflowHTTPToolActionContract
	}
	var err error
	plan.ActionSafety, err = decodeActionSafetyPlanSnapshot(safetySnapshot)
	if err != nil || !workflowHTTPToolPlanMatchesContext(plan, ctx) ||
		plan.RecordVersion != recordVersion || string(plan.Status) != status || plan.ToolPlanDigest != digest ||
		plan.ToolVersion != toolVersion || toolVersion != workflowHTTPToolVersion || !workflowHTTPToolStorageSourceMatchesPlan(source, plan) {
		return WorkflowHTTPToolActionPlan{}, errWorkflowHTTPToolActionContract
	}
	computed, err := workflowHTTPToolPlanDigest(plan)
	if err != nil || computed != plan.ToolPlanDigest {
		return WorkflowHTTPToolActionPlan{}, errWorkflowHTTPToolActionContract
	}
	return plan, nil
}

func nullableWorkflowHTTPToolString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func nullableWorkflowHTTPToolInt(value int) any {
	if value == 0 {
		return nil
	}
	return value
}

var _ workflowHTTPToolActionStore = (*sqliteWorkflowHTTPToolActionStore)(nil)
