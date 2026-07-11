package workflowdraftmigrations

import (
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestEmbeddedMigrationIdentityIsStable(t *testing.T) {
	if !strings.HasPrefix(ExpectedChecksum(), "sha256:") || len(ExpectedChecksum()) != len("sha256:")+64 {
		t.Fatalf("unexpected migration checksum: %s", ExpectedChecksum())
	}
	for _, literal := range []string{
		"CREATE TABLE saved_workflow_drafts",
		"PRIMARY KEY (tenant_ref, workspace_id, application_id, draft_id)",
		"saved_workflow_drafts_owner_list_idx",
		"saved_workflow_drafts_status_list_idx",
	} {
		if !strings.Contains(upMigrationSQL, literal) {
			t.Fatalf("embedded migration is missing %q", literal)
		}
	}
	if !strings.Contains(downMigrationSQL, "DROP TABLE IF EXISTS saved_workflow_drafts") {
		t.Fatalf("test rollback SQL does not remove saved workflow drafts")
	}
}

func TestSafeDatabaseErrorDoesNotExposeConnectionMaterial(t *testing.T) {
	secretMaterial := "postgresql://operator:secret@example.invalid/private"
	message := safeDatabaseError("connect", errors.New(secretMaterial)).Error()
	if strings.Contains(message, secretMaterial) || strings.Contains(message, "secret") {
		t.Fatalf("generic database error leaked connection material: %s", message)
	}

	message = safeDatabaseError("query", &pgconn.PgError{
		Code:    "23505",
		Message: secretMaterial,
	}).Error()
	if message != "query failed (SQLSTATE 23505)" {
		t.Fatalf("PostgreSQL error should expose only SQLSTATE: %s", message)
	}
}
