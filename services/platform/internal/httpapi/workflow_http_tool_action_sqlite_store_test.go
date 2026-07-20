package httpapi

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestSQLiteWorkflowHTTPToolActionStorePersistsCASAndScope(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "workflow-http-tool-actions.db")
	runtime := openWorkflowRunSQLiteRuntimeWithoutCleanup(t, databasePath)
	ctx := workflowHTTPToolActionTestContext()
	store := newSQLiteWorkflowHTTPToolActionStore(runtime.DB())
	plan := workflowHTTPToolActionPlanForStoreTest(t, ctx, "wtap_0000000000000100")
	createdAudit := workflowHTTPToolAuditForStoreTest(plan, "wtae_0000000000000100", "confirmation_requested")
	if err := store.CreatePlan(ctx, &plan, createdAudit); err != nil {
		t.Fatalf("create SQLite tool action plan: %v", err)
	}
	if leaked, found, err := store.ReadPlan(workflowHTTPToolActionContextWithScope(ctx, "tenant_other", ctx.WorkspaceID, ctx.ApplicationID), plan.PlanID); err != nil || found || leaked.PlanID != "" {
		t.Fatalf("SQLite action plan leaked across scope: found=%v plan=%#v err=%v", found, leaked, err)
	}
	if err := runtime.Close(); err != nil {
		t.Fatalf("close SQLite action plan runtime: %v", err)
	}

	reopened := openWorkflowRunSQLiteRuntimeWithoutCleanup(t, databasePath)
	t.Cleanup(func() { _ = reopened.Close() })
	store = newSQLiteWorkflowHTTPToolActionStore(reopened.DB())
	restored, found, err := store.ReadPlan(ctx, plan.PlanID)
	if err != nil || !found || restored.RecordVersion != 1 || restored.ToolPlanDigest != plan.ToolPlanDigest {
		t.Fatalf("restore SQLite action plan: found=%v plan=%#v err=%v", found, restored, err)
	}

	updated := cloneWorkflowHTTPToolActionPlan(restored)
	updated.RecordVersion = 2
	updated.Status = WorkflowHTTPToolActionStatusApproved
	actor, decidedAt := ctx.ActorRef, "2026-07-16T09:01:00Z"
	updated.LastDecisionByActorRef, updated.LastDecisionAt = &actor, &decidedAt
	decision := workflowHTTPToolDecisionForStoreTest(updated, "wtcd_0000000000000100", WorkflowHTTPToolConfirmationApprove, actor)
	decision.AuditRef = "audit_workflow_http_tool_sqlite_decision"
	decisionCtx := ctx
	decisionCtx.AuditRef = decision.AuditRef
	audit := workflowHTTPToolAuditForStoreTest(updated, "wtae_0000000000000101", "confirmation_recorded", decision.ConfirmationID)
	audit.AuditRef = decision.AuditRef
	if err := store.DecidePlan(decisionCtx, &updated, decision, audit); err != nil {
		t.Fatalf("decide SQLite action plan: %v", err)
	}

	illegal := cloneWorkflowHTTPToolActionPlan(updated)
	illegal.RecordVersion = 3
	illegal.Status = WorkflowHTTPToolActionStatusRejected
	illegalDecision := workflowHTTPToolDecisionForStoreTest(illegal, "wtcd_0000000000000101", WorkflowHTTPToolConfirmationReject, actor)
	illegalDecision.AuditRef = "audit_workflow_http_tool_sqlite_loser"
	illegalAudit := workflowHTTPToolAuditForStoreTest(illegal, "wtae_0000000000000102", "confirmation_rejected", illegalDecision.ConfirmationID)
	illegalAudit.AuditRef = illegalDecision.AuditRef
	illegalCtx := ctx
	illegalCtx.AuditRef = illegalDecision.AuditRef
	if err := store.DecidePlan(illegalCtx, &illegal, illegalDecision, illegalAudit); !errors.Is(err, errWorkflowHTTPToolActionConflict) {
		t.Fatalf("illegal SQLite approved-to-rejected transition must conflict, got %v", err)
	}

	losing := cloneWorkflowHTTPToolActionPlan(updated)
	losing.Status = WorkflowHTTPToolActionStatusRejected
	losingDecision := workflowHTTPToolDecisionForStoreTest(losing, "wtcd_0000000000000102", WorkflowHTTPToolConfirmationReject, actor)
	losingDecision.AuditRef = "audit_workflow_http_tool_sqlite_loser"
	losingAudit := workflowHTTPToolAuditForStoreTest(losing, "wtae_0000000000000103", "confirmation_rejected", losingDecision.ConfirmationID)
	losingAudit.AuditRef = losingDecision.AuditRef
	losingCtx := ctx
	losingCtx.AuditRef = losingDecision.AuditRef
	if err := store.DecidePlan(losingCtx, &losing, losingDecision, losingAudit); !errors.Is(err, errWorkflowHTTPToolActionConflict) {
		t.Fatalf("stale SQLite decision must conflict, got %v", err)
	}

	invalidated := cloneWorkflowHTTPToolActionPlan(updated)
	invalidated.RecordVersion = 3
	invalidated.Status = WorkflowHTTPToolActionStatusInvalidated
	systemActor, invalidatedAt := "system:workflow_tool_action_policy", "2026-07-16T09:02:00Z"
	invalidated.LastDecisionByActorRef, invalidated.LastDecisionAt = &systemActor, &invalidatedAt
	invalidateDecision := workflowHTTPToolSystemDecisionForStoreTest(invalidated, "wtcd_0000000000000103", workflowHTTPToolConfirmationInvalidate)
	invalidateDecision.DecidedAt = invalidatedAt
	invalidateDecision.AuditRef = "audit_workflow_http_tool_sqlite_invalidate"
	invalidateCtx := ctx
	invalidateCtx.AuditRef = invalidateDecision.AuditRef
	invalidateAudit := workflowHTTPToolAuditForStoreTest(invalidated, "wtae_0000000000000104", "confirmation_invalidated", invalidateDecision.ConfirmationID)
	invalidateAudit.AuditRef = invalidateDecision.AuditRef
	if err := store.DecidePlan(invalidateCtx, &invalidated, invalidateDecision, invalidateAudit); err != nil {
		t.Fatalf("invalidate approved SQLite action plan: %v", err)
	}

	expireCtx := ctx
	expireCtx.AuditRef = "audit_workflow_http_tool_sqlite_expire_create"
	expired := workflowHTTPToolActionPlanForStoreTest(t, expireCtx, "wtap_0000000000000110")
	if err := store.CreatePlan(expireCtx, &expired, workflowHTTPToolAuditForStoreTest(expired, "wtae_0000000000000110", "confirmation_requested")); err != nil {
		t.Fatalf("create SQLite plan for expiry: %v", err)
	}
	expired.RecordVersion = 2
	expired.Status = WorkflowHTTPToolActionStatusExpired
	expiredAt := "2026-07-16T09:16:00Z"
	expired.LastDecisionByActorRef, expired.LastDecisionAt = &systemActor, &expiredAt
	expireDecision := workflowHTTPToolSystemDecisionForStoreTest(expired, "wtcd_0000000000000110", workflowHTTPToolConfirmationExpire)
	expireDecision.DecidedAt = expiredAt
	expireDecision.AuditRef = "audit_workflow_http_tool_sqlite_expire"
	expireCtx.AuditRef = expireDecision.AuditRef
	expireAudit := workflowHTTPToolAuditForStoreTest(expired, "wtae_0000000000000111", "confirmation_expired", expireDecision.ConfirmationID)
	expireAudit.AuditRef = expireDecision.AuditRef
	if err := store.DecidePlan(expireCtx, &expired, expireDecision, expireAudit); err != nil {
		t.Fatalf("expire pending SQLite action plan: %v", err)
	}
	if recoveredExpired, found, err := store.ReadPlan(expireCtx, expired.PlanID); err != nil || !found || recoveredExpired.Status != WorkflowHTTPToolActionStatusExpired {
		t.Fatalf("read expired SQLite action plan: found=%v plan=%#v err=%v", found, recoveredExpired, err)
	}

	deferCtx := ctx
	deferCtx.AuditRef = "audit_workflow_http_tool_sqlite_defer_create"
	deferred := workflowHTTPToolActionPlanForStoreTest(t, deferCtx, "wtap_0000000000000120")
	if err := store.CreatePlan(deferCtx, &deferred, workflowHTTPToolAuditForStoreTest(deferred, "wtae_0000000000000120", "confirmation_requested")); err != nil {
		t.Fatalf("create SQLite plan for defer transition: %v", err)
	}
	deferred.RecordVersion = 2
	deferred.Status = WorkflowHTTPToolActionStatusDeferred
	deferredAt := "2026-07-16T09:01:00Z"
	deferred.LastDecisionByActorRef, deferred.LastDecisionAt = &actor, &deferredAt
	deferDecision := workflowHTTPToolDecisionForStoreTest(deferred, "wtcd_0000000000000120", WorkflowHTTPToolConfirmationDefer, actor)
	deferDecision.AuditRef = "audit_workflow_http_tool_sqlite_defer"
	deferCtx.AuditRef = deferDecision.AuditRef
	deferAudit := workflowHTTPToolAuditForStoreTest(deferred, "wtae_0000000000000121", "confirmation_deferred", deferDecision.ConfirmationID)
	deferAudit.AuditRef = deferDecision.AuditRef
	if err := store.DecidePlan(deferCtx, &deferred, deferDecision, deferAudit); err != nil {
		t.Fatalf("defer pending SQLite action plan: %v", err)
	}
	repeatedDefer := cloneWorkflowHTTPToolActionPlan(deferred)
	repeatedDefer.RecordVersion = 3
	repeatedAt := "2026-07-16T09:02:00Z"
	repeatedDefer.LastDecisionAt = &repeatedAt
	repeatedDecision := workflowHTTPToolDecisionForStoreTest(repeatedDefer, "wtcd_0000000000000121", WorkflowHTTPToolConfirmationDefer, actor)
	repeatedDecision.DecidedAt = repeatedAt
	repeatedDecision.AuditRef = "audit_workflow_http_tool_sqlite_repeat_defer"
	repeatedCtx := deferCtx
	repeatedCtx.AuditRef = repeatedDecision.AuditRef
	repeatedAudit := workflowHTTPToolAuditForStoreTest(repeatedDefer, "wtae_0000000000000122", "confirmation_deferred", repeatedDecision.ConfirmationID)
	repeatedAudit.AuditRef = repeatedDecision.AuditRef
	if err := store.DecidePlan(repeatedCtx, &repeatedDefer, repeatedDecision, repeatedAudit); !errors.Is(err, errWorkflowHTTPToolActionConflict) {
		t.Fatalf("repeated SQLite deferred-to-deferred transition must conflict, got %v", err)
	}

	var status string
	var recordVersion, toolVersion, decisionCount, auditCount, decisionToolVersionMin, decisionToolVersionMax, auditToolVersionMin, auditToolVersionMax int
	if err := reopened.DB().QueryRow(`SELECT status,record_version,tool_version FROM workflow_http_tool_action_plans
	 WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND plan_id=?`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, plan.PlanID).Scan(&status, &recordVersion, &toolVersion); err != nil {
		t.Fatal(err)
	}
	if err := reopened.DB().QueryRow(`SELECT count(*),min(tool_version),max(tool_version) FROM workflow_http_tool_confirmation_decisions WHERE plan_id=?`, plan.PlanID).Scan(&decisionCount, &decisionToolVersionMin, &decisionToolVersionMax); err != nil {
		t.Fatal(err)
	}
	if err := reopened.DB().QueryRow(`SELECT count(*),min(tool_version),max(tool_version) FROM workflow_http_tool_execution_audits WHERE plan_id=?`, plan.PlanID).Scan(&auditCount, &auditToolVersionMin, &auditToolVersionMax); err != nil {
		t.Fatal(err)
	}
	if status != string(WorkflowHTTPToolActionStatusInvalidated) || recordVersion != 3 || toolVersion != workflowHTTPToolVersion ||
		decisionCount != 2 || decisionToolVersionMin != workflowHTTPToolVersion || decisionToolVersionMax != workflowHTTPToolVersion ||
		auditCount != 3 || auditToolVersionMin != workflowHTTPToolVersion || auditToolVersionMax != workflowHTTPToolVersion {
		t.Fatalf("SQLite CAS left partial or version-drifted evidence: status=%s version=%d tool_version=%d decisions=%d decision_tool_versions=%d..%d audits=%d audit_tool_versions=%d..%d",
			status, recordVersion, toolVersion, decisionCount, decisionToolVersionMin, decisionToolVersionMax,
			auditCount, auditToolVersionMin, auditToolVersionMax)
	}
	var deferredStatus string
	var deferredVersion, deferredDecisionCount, deferredAuditCount int
	if err := reopened.DB().QueryRow(`SELECT status,record_version FROM workflow_http_tool_action_plans WHERE plan_id=?`, deferred.PlanID).Scan(&deferredStatus, &deferredVersion); err != nil {
		t.Fatal(err)
	}
	if err := reopened.DB().QueryRow(`SELECT count(*) FROM workflow_http_tool_confirmation_decisions WHERE plan_id=?`, deferred.PlanID).Scan(&deferredDecisionCount); err != nil {
		t.Fatal(err)
	}
	if err := reopened.DB().QueryRow(`SELECT count(*) FROM workflow_http_tool_execution_audits WHERE plan_id=?`, deferred.PlanID).Scan(&deferredAuditCount); err != nil {
		t.Fatal(err)
	}
	if deferredStatus != string(WorkflowHTTPToolActionStatusDeferred) || deferredVersion != 2 || deferredDecisionCount != 1 || deferredAuditCount != 2 {
		t.Fatalf("repeated SQLite defer wrote partial evidence: status=%s version=%d decisions=%d audits=%d",
			deferredStatus, deferredVersion, deferredDecisionCount, deferredAuditCount)
	}
}
