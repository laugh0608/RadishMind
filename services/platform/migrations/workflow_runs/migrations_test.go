package workflowrunmigrations

import (
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestEmbeddedWorkflowRunMigration(t *testing.T) {
	if !strings.Contains(upSQL, "CREATE TABLE workflow_run_records") || !strings.Contains(upSQL, "failure_boundary") || !strings.Contains(upSQL, "CREATE TABLE workflow_evaluation_cases") || !strings.Contains(upSQL, "CREATE TABLE workflow_evaluation_case_revisions") || !strings.Contains(upSQL, "ADD COLUMN current_version") || !strings.Contains(upSQL, "CREATE TABLE workflow_evaluation_suites") || !strings.Contains(upSQL, "CREATE TABLE workflow_evaluation_suite_decisions") || !strings.Contains(downSQL, "DROP TABLE IF EXISTS workflow_evaluation_suite_decisions") || !strings.Contains(downSQL, "DROP TABLE IF EXISTS workflow_evaluation_suites") || !strings.Contains(downSQL, "DROP TABLE IF EXISTS workflow_evaluation_case_revisions") || !strings.Contains(downSQL, "DROP TABLE IF EXISTS workflow_evaluation_cases") || !strings.Contains(downSQL, "DROP TABLE IF EXISTS workflow_run_records") {
		t.Fatal("workflow run migration contract is incomplete")
	}
	if !strings.HasPrefix(ExpectedChecksum(), "sha256:") {
		t.Fatal("workflow run migration checksum is invalid")
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
