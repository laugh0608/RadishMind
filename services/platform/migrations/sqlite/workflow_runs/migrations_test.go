package workflowruns

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"radishmind.local/services/platform/internal/sqlitedev"
)

func TestWorkflowRunSQLiteMigrationsAreOrderedThroughRAGSnapshotBatchA(t *testing.T) {
	migrations := Migrations()
	if len(migrations) != 4 {
		t.Fatalf("unexpected workflow run SQLite migration count: %d", len(migrations))
	}
	if migrations[0].ID != legacyMigrationID || migrations[0].StoreSchemaVersion != legacyRunStoreSchemaVersion {
		t.Fatalf("legacy workflow run migration drifted: %#v", migrations[0])
	}
	if migrations[1].ID != toolActionsMigrationID || migrations[1].StoreSchemaVersion != toolActionsStoreSchemaVersion {
		t.Fatalf("HTTP tool action migration drifted: %#v", migrations[1])
	}
	if migrations[2].ID != toolExecutionMigrationID || migrations[2].StoreSchemaVersion != toolExecutionStoreSchemaVersion {
		t.Fatalf("HTTP tool execution migration drifted: %#v", migrations[2])
	}
	if migrations[3].ID != MigrationID || migrations[3].StoreSchemaVersion != StoreSchemaVersion {
		t.Fatalf("workflow RAG snapshot migration drifted: %#v", migrations[3])
	}
	for _, required := range []string{
		"CREATE TABLE workflow_http_tool_action_plans",
		"CREATE TABLE workflow_http_tool_confirmation_decisions",
		"CREATE TABLE workflow_http_tool_execution_audits",
		"workflow_http_tool_confirmation_decisions_append_only_update",
		"workflow_http_tool_execution_audits_append_only_delete",
	} {
		if !strings.Contains(upSQLV2, required) {
			t.Fatalf("SQLite HTTP tool action migration is missing %q", required)
		}
	}
	for _, required := range []string{
		"CREATE TABLE workflow_rag_snapshot_resources",
		"CREATE TABLE workflow_rag_snapshot_versions",
		"CREATE TABLE workflow_rag_snapshot_fragments",
		"CREATE TABLE workflow_rag_execution_audits",
		"workflow_rag_snapshot_versions_append_only_update",
		"workflow_rag_snapshot_fragments_append_only_delete",
		"workflow_rag_execution_audits_append_only_update",
	} {
		if !strings.Contains(upSQLV4, required) {
			t.Fatalf("SQLite workflow RAG snapshot migration is missing %q", required)
		}
	}
	if count := strings.Count(upSQLV2, "tool_version INTEGER NOT NULL CHECK (tool_version = 1)"); count != 3 {
		t.Fatalf("SQLite HTTP tool action migration must pin tool_version=1 on all three tables, got %d", count)
	}
	for _, forbidden := range []string{
		"workflow_http_tool_execution_attempts",
		"ALTER TABLE workflow_run_records",
		"workflow_run_record.v2",
	} {
		if strings.Contains(upSQLV2, forbidden) {
			t.Fatalf("batch A SQLite migration contains batch B storage %q", forbidden)
		}
	}
	for _, required := range []string{
		"CREATE TABLE workflow_run_records_v3",
		"workflow_run_record.v2",
		"outcome_unknown",
		"CREATE TABLE workflow_http_tool_execution_attempts",
	} {
		if !strings.Contains(upSQLV3, required) {
			t.Fatalf("SQLite HTTP tool execution migration is missing %q", required)
		}
	}
}

func TestWorkflowRunSQLiteMigrationUpgradesWithoutChangingLegacyRuns(t *testing.T) {
	ctx := context.Background()
	databasePath := filepath.Join(t.TempDir(), "workflow-runs.db")
	migrations := Migrations()

	legacyRuntime, err := sqlitedev.Open(ctx, sqlitedev.Options{
		DatabasePath: databasePath,
		Migrations:   migrations[:1],
	})
	if err != nil {
		t.Fatalf("open legacy workflow run SQLite database: %v", err)
	}
	if _, err = legacyRuntime.DB().ExecContext(ctx, `INSERT INTO workflow_run_records (
		tenant_ref, workspace_id, application_id, run_id, draft_id, draft_version,
		record_version, store_schema_version, schema_version, run_status,
		started_at_unix_nano, completed_at_unix_nano, actor_ref, request_id, audit_ref,
		failure_code, failure_boundary, selected_provider, selected_model, sanitized_run_record
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NULL, ?, ?, ?, '', '', '', '', ?)`,
		"tenant_demo", "workspace_demo", "application_demo", "run_legacy",
		"draft_demo", 1, 1, legacyRunStoreSchemaVersion, "workflow_run_record.v1",
		"running", int64(1), "actor_demo", "request_demo", "audit_run_demo", `{}`,
	); err != nil {
		_ = legacyRuntime.Close()
		t.Fatalf("seed legacy workflow run: %v", err)
	}
	if err = legacyRuntime.Close(); err != nil {
		t.Fatalf("close legacy workflow run SQLite database: %v", err)
	}

	upgradedRuntime, err := sqlitedev.Open(ctx, sqlitedev.Options{
		DatabasePath: databasePath,
		Migrations:   migrations,
	})
	if err != nil {
		t.Fatalf("upgrade workflow run SQLite database: %v", err)
	}
	if err = upgradedRuntime.VerifyMigrations(ctx, migrations); err != nil {
		_ = upgradedRuntime.Close()
		t.Fatalf("verify upgraded workflow run SQLite migrations: %v", err)
	}
	var legacyRunCount, migrationCount, actionTableCount, triggerCount int
	if err = upgradedRuntime.DB().QueryRowContext(ctx, `SELECT count(*) FROM workflow_run_records WHERE run_id='run_legacy' AND store_schema_version=?`, RunRecordStoreSchemaVersion).Scan(&legacyRunCount); err != nil || legacyRunCount != 1 {
		_ = upgradedRuntime.Close()
		t.Fatalf("legacy workflow run changed during upgrade: count=%d err=%v", legacyRunCount, err)
	}
	if err = upgradedRuntime.DB().QueryRowContext(ctx, `SELECT count(*) FROM radishmind_schema_migrations WHERE component=?`, Component).Scan(&migrationCount); err != nil || migrationCount != 4 {
		_ = upgradedRuntime.Close()
		t.Fatalf("unexpected workflow run migration markers: count=%d err=%v", migrationCount, err)
	}
	if err = upgradedRuntime.DB().QueryRowContext(ctx, `SELECT count(*) FROM sqlite_master WHERE type='table' AND name IN (
		'workflow_http_tool_action_plans',
		'workflow_http_tool_confirmation_decisions',
		'workflow_http_tool_execution_audits',
		'workflow_http_tool_execution_attempts'
	)`).Scan(&actionTableCount); err != nil || actionTableCount != 4 {
		_ = upgradedRuntime.Close()
		t.Fatalf("HTTP tool action tables are incomplete: count=%d err=%v", actionTableCount, err)
	}
	if err = upgradedRuntime.DB().QueryRowContext(ctx, `SELECT count(*) FROM sqlite_master WHERE type='trigger' AND name LIKE 'workflow_http_tool_%_append_only_%'`).Scan(&triggerCount); err != nil || triggerCount != 4 {
		_ = upgradedRuntime.Close()
		t.Fatalf("append-only triggers are incomplete: count=%d err=%v", triggerCount, err)
	}
	var ragTableCount, ragTriggerCount int
	if err = upgradedRuntime.DB().QueryRowContext(ctx, `SELECT count(*) FROM sqlite_master WHERE type='table' AND name IN (
		'workflow_rag_snapshot_resources','workflow_rag_snapshot_versions','workflow_rag_snapshot_fragments','workflow_rag_execution_audits'
	)`).Scan(&ragTableCount); err != nil || ragTableCount != 4 {
		_ = upgradedRuntime.Close()
		t.Fatalf("workflow RAG snapshot tables are incomplete: count=%d err=%v", ragTableCount, err)
	}
	if err = upgradedRuntime.DB().QueryRowContext(ctx, `SELECT count(*) FROM sqlite_master WHERE type='trigger' AND name LIKE 'workflow_rag_%_append_only_%'`).Scan(&ragTriggerCount); err != nil || ragTriggerCount != 6 {
		_ = upgradedRuntime.Close()
		t.Fatalf("workflow RAG snapshot append-only triggers are incomplete: count=%d err=%v", ragTriggerCount, err)
	}

	insertSQLiteHTTPToolActionRecords(t, ctx, upgradedRuntime)
	var pinnedToolVersionCount int
	if err = upgradedRuntime.DB().QueryRowContext(ctx, `SELECT count(*) FROM (
		SELECT tool_version FROM workflow_http_tool_action_plans WHERE plan_id='plan_demo'
		UNION ALL
		SELECT tool_version FROM workflow_http_tool_confirmation_decisions WHERE plan_id='plan_demo'
		UNION ALL
		SELECT tool_version FROM workflow_http_tool_execution_audits WHERE plan_id='plan_demo'
	) WHERE tool_version=1`).Scan(&pinnedToolVersionCount); err != nil || pinnedToolVersionCount != 3 {
		_ = upgradedRuntime.Close()
		t.Fatalf("SQLite HTTP tool records did not persist tool_version=1: count=%d err=%v", pinnedToolVersionCount, err)
	}
	if _, err = upgradedRuntime.DB().ExecContext(ctx, `UPDATE workflow_http_tool_action_plans SET tool_version=2 WHERE plan_id='plan_demo'`); err == nil {
		_ = upgradedRuntime.Close()
		t.Fatal("SQLite action plan accepted tool_version other than 1")
	}
	assertSQLiteHTTPToolActionScopeForeignKey(t, ctx, upgradedRuntime)
	assertSQLiteHTTPToolDecisionVersionUnique(t, ctx, upgradedRuntime)
	if _, err = upgradedRuntime.DB().ExecContext(ctx, `UPDATE workflow_http_tool_confirmation_decisions SET outcome='reject' WHERE confirmation_id='confirmation_demo'`); err == nil {
		_ = upgradedRuntime.Close()
		t.Fatal("SQLite confirmation decision update must be rejected")
	}
	if _, err = upgradedRuntime.DB().ExecContext(ctx, `DELETE FROM workflow_http_tool_execution_audits WHERE audit_ref='audit_confirmation_demo'`); err == nil {
		_ = upgradedRuntime.Close()
		t.Fatal("SQLite execution audit delete must be rejected")
	}
	if _, err = upgradedRuntime.DB().ExecContext(ctx, sqliteActionPlanInsertSQL,
		"tenant_demo", "workspace_demo", "application_demo", "plan_bad_digest", invalidDigest,
	); err == nil {
		_ = upgradedRuntime.Close()
		t.Fatal("SQLite action plan accepted an invalid digest")
	}
	if err = upgradedRuntime.Close(); err != nil {
		t.Fatalf("close upgraded workflow run SQLite database: %v", err)
	}

	reopenedRuntime, err := sqlitedev.Open(ctx, sqlitedev.Options{
		DatabasePath: databasePath,
		Migrations:   migrations,
	})
	if err != nil {
		t.Fatalf("reopen upgraded workflow run SQLite database: %v", err)
	}
	t.Cleanup(func() { _ = reopenedRuntime.Close() })
	var planStatus string
	var recordVersion, toolVersion int64
	if err = reopenedRuntime.DB().QueryRowContext(ctx, `SELECT status, record_version, tool_version FROM workflow_http_tool_action_plans WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND plan_id=?`,
		"tenant_demo", "workspace_demo", "application_demo", "plan_demo",
	).Scan(&planStatus, &recordVersion, &toolVersion); err != nil || planStatus != "approved" || recordVersion != 2 || toolVersion != 1 {
		t.Fatalf("durable action plan did not survive restart: status=%s version=%d tool_version=%d err=%v", planStatus, recordVersion, toolVersion, err)
	}
}

const (
	validDigest   = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	invalidDigest = "sha256:not-a-digest"

	sqliteActionPlanInsertSQL = `INSERT INTO workflow_http_tool_action_plans (
		tenant_ref, workspace_id, application_id, plan_id, schema_version,
		status, record_version, draft_id, draft_version, node_id,
			tool_id, tool_version, definition_digest, profile_id, profile_version,
		profile_digest, target_policy_key, tool_plan_digest,
		method, credential_policy, timeout_ms, max_response_bytes, max_output_bytes,
		planned_by_actor_ref, audit_ref, created_at_unix_nano,
		expires_at_unix_nano, sanitized_action_plan
	) VALUES (?, ?, ?, ?, 'workflow_http_tool_action_plan.v1',
		'pending', 1, 'draft_demo', 1, 'node_http',
			'workflow.http.read_demo.v1', 1, '` + validDigest + `', 'profile_demo', 1,
		'` + validDigest + `', 'target_demo', ?,
		'GET', 'none', 5000, 65536, 16384, 'actor_planner', 'audit_plan_demo',
		1, 901, '{}')`
)

func insertSQLiteHTTPToolActionRecords(t *testing.T, ctx context.Context, runtime *sqlitedev.Runtime) {
	t.Helper()
	tx, err := runtime.DB().BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin SQLite HTTP tool action transaction: %v", err)
	}
	defer func() { _ = tx.Rollback() }()
	if _, err = tx.ExecContext(ctx, sqliteActionPlanInsertSQL,
		"tenant_demo", "workspace_demo", "application_demo", "plan_demo", validDigest,
	); err != nil {
		t.Fatalf("insert SQLite HTTP tool action plan: %v", err)
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO workflow_http_tool_execution_audits (
		tenant_ref, workspace_id, application_id, plan_id, audit_id,
			schema_version, event_kind, tool_version, tool_plan_digest,
		actor_ref, request_id, audit_ref, occurred_at_unix_nano,
		sanitized_execution_audit
	) VALUES (?, ?, ?, ?, ?, 'workflow_http_tool_execution_audit.v1',
			'confirmation_recorded', 1, ?, 'actor_confirmer', 'request_decide_demo',
		'audit_confirmation_demo', 2, '{}')`,
		"tenant_demo", "workspace_demo", "application_demo", "plan_demo",
		"audit_event_confirmation_demo", validDigest,
	); err != nil {
		t.Fatalf("insert SQLite HTTP tool execution audit: %v", err)
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO workflow_http_tool_confirmation_decisions (
		tenant_ref, workspace_id, application_id, plan_id, confirmation_id,
		schema_version, outcome, draft_id, draft_version,
			node_id, tool_id, tool_version, tool_plan_digest,
		expected_record_version, resulting_record_version,
		decided_by_actor_ref, actor_source, reason_code,
		decided_at_unix_nano, audit_ref,
		sanitized_confirmation_decision
	) VALUES (?, ?, ?, ?, ?, 'workflow_http_tool_confirmation_decision.v1',
			'approve', 'draft_demo', 1, 'node_http', 'workflow.http.read_demo.v1', 1, ?,
		1, 2, 'actor_confirmer', 'human', 'workflow_tool_confirmation_approved',
		2, ?, '{}')`,
		"tenant_demo", "workspace_demo", "application_demo", "plan_demo",
		"confirmation_demo", validDigest, "audit_confirmation_demo",
	); err != nil {
		t.Fatalf("insert SQLite HTTP tool confirmation decision: %v", err)
	}
	if result, updateErr := tx.ExecContext(ctx, `UPDATE workflow_http_tool_action_plans
		SET status='approved', record_version=2
		WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND plan_id=?
		  AND status='pending' AND record_version=1`,
		"tenant_demo", "workspace_demo", "application_demo", "plan_demo",
	); updateErr != nil {
		t.Fatalf("CAS SQLite HTTP tool action plan: %v", updateErr)
	} else if affected, rowsErr := result.RowsAffected(); rowsErr != nil || affected != 1 {
		t.Fatalf("SQLite HTTP tool action CAS count drifted: affected=%d err=%v", affected, rowsErr)
	}
	if err = tx.Commit(); err != nil {
		t.Fatalf("commit SQLite HTTP tool action transaction: %v", err)
	}
}

func assertSQLiteHTTPToolActionScopeForeignKey(t *testing.T, ctx context.Context, runtime *sqlitedev.Runtime) {
	t.Helper()
	tx, err := runtime.DB().BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin SQLite scope foreign-key transaction: %v", err)
	}
	defer func() { _ = tx.Rollback() }()
	if _, err = tx.ExecContext(ctx, `INSERT INTO workflow_http_tool_execution_audits (
		tenant_ref, workspace_id, application_id, plan_id, audit_id,
			schema_version, event_kind, tool_version, tool_plan_digest,
		actor_ref, request_id, audit_ref, occurred_at_unix_nano,
		sanitized_execution_audit
	) VALUES (?, ?, ?, ?, ?, 'workflow_http_tool_execution_audit.v1',
			'confirmation_recorded', 1, ?, 'actor_confirmer', 'request_wrong_scope',
		'audit_wrong_scope', 3, '{}')`,
		"tenant_demo", "workspace_other", "application_demo", "plan_demo",
		"audit_event_wrong_scope", validDigest,
	); err != nil {
		t.Fatalf("stage deferred SQLite scope foreign-key violation: %v", err)
	}
	if err = tx.Commit(); err == nil {
		t.Fatal("SQLite execution audit crossed the action-plan scope foreign key")
	}
}

func assertSQLiteHTTPToolDecisionVersionUnique(t *testing.T, ctx context.Context, runtime *sqlitedev.Runtime) {
	t.Helper()
	tx, err := runtime.DB().BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin SQLite decision uniqueness transaction: %v", err)
	}
	defer func() { _ = tx.Rollback() }()
	if _, err = tx.ExecContext(ctx, `INSERT INTO workflow_http_tool_execution_audits (
		tenant_ref, workspace_id, application_id, plan_id, audit_id,
			schema_version, event_kind, tool_version, tool_plan_digest,
		actor_ref, request_id, audit_ref, occurred_at_unix_nano,
		sanitized_execution_audit
	) VALUES (?, ?, ?, ?, ?, 'workflow_http_tool_execution_audit.v1',
			'confirmation_recorded', 1, ?, 'actor_confirmer', 'request_duplicate_version',
		'audit_duplicate_version', 3, '{}')`,
		"tenant_demo", "workspace_demo", "application_demo", "plan_demo",
		"audit_event_duplicate_version", validDigest,
	); err != nil {
		t.Fatalf("insert SQLite audit for duplicate decision version: %v", err)
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO workflow_http_tool_confirmation_decisions (
		tenant_ref, workspace_id, application_id, plan_id, confirmation_id,
		schema_version, outcome, draft_id, draft_version,
			node_id, tool_id, tool_version, tool_plan_digest,
		expected_record_version, resulting_record_version,
		decided_by_actor_ref, actor_source, reason_code,
		decided_at_unix_nano, audit_ref, sanitized_confirmation_decision
	) VALUES (?, ?, ?, ?, ?, 'workflow_http_tool_confirmation_decision.v1',
			'approve', 'draft_demo', 1, 'node_http', 'workflow.http.read_demo.v1', 1, ?,
		1, 2, 'actor_confirmer', 'human', 'workflow_tool_confirmation_approved',
		3, 'audit_duplicate_version', '{}')`,
		"tenant_demo", "workspace_demo", "application_demo", "plan_demo",
		"confirmation_duplicate_version", validDigest,
	); err == nil {
		t.Fatal("SQLite confirmation decision accepted a duplicate resulting record version")
	}
}
