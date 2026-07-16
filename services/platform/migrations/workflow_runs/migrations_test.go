package workflowrunmigrations

import (
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestEmbeddedWorkflowRunMigration(t *testing.T) {
	for _, required := range []string{
		"CREATE TABLE workflow_run_records",
		"failure_boundary",
		"CREATE TABLE workflow_evaluation_cases",
		"CREATE TABLE workflow_evaluation_case_revisions",
		"ADD COLUMN current_version",
		"CREATE TABLE workflow_evaluation_suites",
		"CREATE TABLE workflow_evaluation_suite_decisions",
		"CREATE TABLE workflow_http_tool_action_plans",
		"CREATE TABLE workflow_http_tool_confirmation_decisions",
		"CREATE TABLE workflow_http_tool_execution_audits",
		"CREATE TABLE workflow_http_tool_execution_attempts",
		"outcome_unknown",
		"workflow_http_tool_confirmation_decisions_append_only",
		"workflow_http_tool_execution_audits_append_only",
	} {
		if !strings.Contains(upSQL, required) {
			t.Fatalf("workflow run up migration is missing %q", required)
		}
	}
	for _, required := range []string{
		"DROP TABLE IF EXISTS workflow_http_tool_execution_attempts",
		"DROP TABLE IF EXISTS workflow_http_tool_confirmation_decisions",
		"DROP TABLE IF EXISTS workflow_http_tool_execution_audits",
		"DROP TABLE IF EXISTS workflow_http_tool_action_plans",
		"DROP TABLE IF EXISTS workflow_evaluation_suite_decisions",
		"DROP TABLE IF EXISTS workflow_evaluation_suites",
		"DROP TABLE IF EXISTS workflow_evaluation_case_revisions",
		"DROP TABLE IF EXISTS workflow_evaluation_cases",
		"DROP TABLE IF EXISTS workflow_run_records",
	} {
		if !strings.Contains(downSQL, required) {
			t.Fatalf("workflow run down migration is missing %q", required)
		}
	}
	if checksum := ExpectedChecksum(); !strings.HasPrefix(checksum, "sha256:") || len(checksum) != 71 {
		t.Fatalf("workflow run migration checksum is invalid: %s", checksum)
	}
	for _, forbidden := range []string{
		"workflow_http_tool_execution_attempts",
		"ALTER TABLE workflow_run_records",
		"workflow_run_record.v2",
	} {
		if strings.Contains(upSQLV6, forbidden) {
			t.Fatalf("batch A migration contains batch B storage %q", forbidden)
		}
	}
	if count := strings.Count(upSQLV6, "tool_version integer NOT NULL CHECK (tool_version = 1)"); count != 3 {
		t.Fatalf("PostgreSQL HTTP tool action migration must pin tool_version=1 on all three tables, got %d", count)
	}
}

func TestWorkflowRunPendingMigrationPaths(t *testing.T) {
	testCases := []struct {
		name              string
		migrationID       string
		requiredFragment  string
		forbiddenFragment string
	}{
		{name: "v1", migrationID: legacyMigrationID, requiredFragment: "ADD COLUMN failure_code"},
		{name: "v2", migrationID: diagnosticsMigrationID, requiredFragment: "CREATE TABLE workflow_evaluation_cases"},
		{name: "v3", migrationID: evaluationMigrationID, requiredFragment: "ADD COLUMN current_version"},
		{name: "v4", migrationID: caseVersioningMigrationID, requiredFragment: "CREATE TABLE workflow_evaluation_suites"},
		{name: "v5", migrationID: evaluationSuiteMigrationID, requiredFragment: "CREATE TABLE workflow_http_tool_action_plans", forbiddenFragment: "CREATE TABLE workflow_evaluation_suites"},
		{name: "v6", migrationID: toolActionsMigrationID, requiredFragment: "CREATE TABLE workflow_http_tool_execution_attempts", forbiddenFragment: "CREATE TABLE workflow_http_tool_action_plans"},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			pendingSQL := pendingMigrationSQL(testCase.migrationID)
			if !strings.Contains(pendingSQL, testCase.requiredFragment) || !strings.Contains(pendingSQL, "CREATE TABLE workflow_http_tool_execution_attempts") {
				t.Fatalf("pending migration path is incomplete for %s", testCase.migrationID)
			}
			if testCase.forbiddenFragment != "" && strings.Contains(pendingSQL, testCase.forbiddenFragment) {
				t.Fatalf("pending migration path replays %q for %s", testCase.forbiddenFragment, testCase.migrationID)
			}
		})
	}
	if pendingMigrationSQL("0000_unknown") != "" {
		t.Fatal("unknown pending migration must fail closed")
	}
}

func TestWorkflowRunPendingRollbackPathsDoNotDropUnappliedTables(t *testing.T) {
	for _, migrationID := range []string{
		legacyMigrationID,
		diagnosticsMigrationID,
		evaluationMigrationID,
		caseVersioningMigrationID,
		evaluationSuiteMigrationID,
	} {
		rollbackSQL := rollbackSQLThrough(migrationID)
		if !strings.Contains(rollbackSQL, "DROP TABLE IF EXISTS workflow_run_records") {
			t.Fatalf("rollback path does not remove v1 for %s", migrationID)
		}
		if strings.Contains(rollbackSQL, "workflow_http_tool_action_plans") {
			t.Fatalf("pending rollback tries to remove unapplied v6 for %s", migrationID)
		}
		if strings.Contains(rollbackSQL, "workflow_http_tool_execution_attempts") {
			t.Fatalf("pending rollback tries to remove unapplied v7 for %s", migrationID)
		}
	}
	toolActionRollback := rollbackSQLThrough(toolActionsMigrationID)
	if !strings.Contains(toolActionRollback, "workflow_http_tool_action_plans") || strings.Contains(toolActionRollback, "workflow_http_tool_execution_attempts") {
		t.Fatalf("v6 rollback must remove applied action tables without removing unapplied v7: %s", toolActionRollback)
	}
	if rollbackSQLThrough("0000_unknown") != "" {
		t.Fatal("unknown pending rollback must fail closed")
	}
}

func TestWorkflowRunDatabaseErrorsAreSanitized(t *testing.T) {
	secret := "postgresql://user:secret@example.invalid/db"
	if got := safeDatabaseError("connect", errors.New(secret)).Error(); strings.Contains(got, "secret") {
		t.Fatalf("connection material leaked: %s", got)
	}
	if got := safeDatabaseError("query", &pgconn.PgError{Code: "23505", Message: secret}).Error(); got != "query failed (SQLSTATE 23505)" {
		t.Fatalf("unexpected PostgreSQL error: %s", got)
	}
}
