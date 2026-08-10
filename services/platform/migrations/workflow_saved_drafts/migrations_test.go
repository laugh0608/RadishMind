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
		if !strings.Contains(initialUpMigrationSQL, literal) {
			t.Fatalf("embedded migration is missing %q", literal)
		}
	}
	for _, literal := range []string{
		"saved_workflow_draft.v2",
		"saved_workflow_drafts_payload_schema_check",
		"input_contract,contract_digest",
	} {
		if !strings.Contains(structuredInputUpMigrationSQL, literal) {
			t.Fatalf("embedded structured input migration is missing %q", literal)
		}
	}
	for _, literal := range []string{
		"CREATE TABLE saved_workflow_draft_revisions",
		"revision_kind",
		"backfilled_current",
		"saved_workflow_draft_revisions_owner_version_idx",
	} {
		if !strings.Contains(revisionUpMigrationSQL, literal) {
			t.Fatalf("embedded revision migration is missing %q", literal)
		}
	}
	for _, literal := range []string{
		"CREATE TABLE saved_workflow_draft_lifecycle_events",
		"saved_workflow_drafts_owner_lifecycle_list_idx",
		"saved_workflow_drafts_validation_list_idx",
		"saved_workflow_drafts_provenance_list_idx",
		"saved_workflow_drafts_name_list_idx",
		"library_updated_at = updated_at",
		"provenance_kind = CASE",
	} {
		if !strings.Contains(libraryUpMigrationSQL, literal) {
			t.Fatalf("embedded library migration is missing %q", literal)
		}
	}
	if !strings.Contains(initialDownMigrationSQL, "DROP TABLE IF EXISTS saved_workflow_drafts") {
		t.Fatalf("test rollback SQL does not remove saved workflow drafts")
	}
	if !strings.Contains(revisionDownMigrationSQL, "DROP TABLE IF EXISTS saved_workflow_draft_revisions") {
		t.Fatalf("test rollback SQL does not remove saved workflow draft revisions")
	}
	if !strings.Contains(
		libraryDownMigrationSQL,
		"DROP TABLE IF EXISTS saved_workflow_draft_lifecycle_events",
	) {
		t.Fatalf("test rollback SQL does not remove saved workflow draft lifecycle events")
	}
	if !strings.Contains(structuredInputDownMigrationSQL, "DROP CONSTRAINT IF EXISTS saved_workflow_drafts_payload_schema_check") {
		t.Fatal("test rollback SQL does not remove the structured input payload constraint")
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
