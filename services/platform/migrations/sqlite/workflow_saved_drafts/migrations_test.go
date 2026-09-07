package workflowsaveddrafts

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/sqlitedev"
)

func TestSQLiteSavedWorkflowDraftMigrationChainIncludesStructuredInputs(t *testing.T) {
	migrations := Migrations()
	if len(migrations) != 5 || migrations[len(migrations)-1].ID != MigrationID {
		t.Fatalf("unexpected SQLite saved draft migration chain: %#v", migrations)
	}
	for _, literal := range []string{
		"saved_workflow_draft_lifecycle_events",
		"saved_workflow_drafts_owner_lifecycle_list_idx",
		"saved_workflow_drafts_validation_list_idx",
		"saved_workflow_drafts_provenance_list_idx",
		"saved_workflow_drafts_name_list_idx",
		"library_updated_at_unix_nano = updated_at_unix_nano",
		"provenance_kind = CASE",
	} {
		if !strings.Contains(libraryUpSQL, literal) {
			t.Fatalf("SQLite library lifecycle migration is missing %q", literal)
		}
	}
	for _, literal := range []string{
		"saved_workflow_draft.v2",
		"input_contract.contract_digest",
		"saved_workflow_drafts_pre_structured_inputs",
		"saved_workflow_draft_revisions_pre_structured_inputs",
		"saved_workflow_draft_lifecycle_events_pre_structured_inputs",
	} {
		if !strings.Contains(structuredInputUpSQL, literal) {
			t.Fatalf("SQLite structured input migration is missing %q", literal)
		}
	}
	for _, literal := range []string{
		"workspace_template_derivation",
		"CREATE TABLE workflow_template_candidates",
		"CREATE TABLE workflow_template_versions",
		"CREATE TABLE workflow_template_lineages",
		"workflow_template_candidates_workspace_list_idx",
		"workflow_template_versions_no_update",
	} {
		if !strings.Contains(workflowTemplateCatalogUpSQL, literal) {
			t.Fatalf("SQLite workflow template catalog migration is missing %q", literal)
		}
	}
}

func TestSQLiteSavedWorkflowDraftTemplateCatalogUpgradeRollbackAndReapply(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	databasePath := filepath.Join(t.TempDir(), "workflow-template-migration.db")
	legacyMigrations := Migrations()[:4]
	legacyRuntime, err := sqlitedev.Open(ctx, sqlitedev.Options{
		DatabasePath: databasePath,
		Migrations:   legacyMigrations,
	})
	if err != nil {
		t.Fatalf("apply SQLite saved draft migration through 0004: %v", err)
	}
	if err = legacyRuntime.Close(); err != nil {
		t.Fatalf("close SQLite 0004 runtime: %v", err)
	}

	failedMigrations := append([]sqlitedev.Migration(nil), Migrations()...)
	failedMigrations = append(failedMigrations, sqlitedev.Migration{
		Component:          Component,
		ID:                 "0006_intentional_failure",
		StoreSchemaVersion: StoreSchemaVersion,
		UpSQL:              "CREATE TABLE workflow_template_partial_write (id INTEGER); SELECT * FROM missing_workflow_template_source;",
	})
	if failedRuntime, openErr := sqlitedev.Open(ctx, sqlitedev.Options{
		DatabasePath: databasePath,
		Migrations:   failedMigrations,
	}); openErr == nil {
		_ = failedRuntime.Close()
		t.Fatal("SQLite failed migration unexpectedly committed")
	}

	rolledBack, err := sqlitedev.Open(ctx, sqlitedev.Options{
		DatabasePath: databasePath,
		Migrations:   legacyMigrations,
	})
	if err != nil {
		t.Fatalf("reopen SQLite after migration rollback: %v", err)
	}
	for _, table := range []string{"workflow_template_candidates", "workflow_template_partial_write"} {
		var count int
		if err = rolledBack.DB().QueryRowContext(ctx, `SELECT count(*) FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&count); err != nil || count != 0 {
			t.Fatalf("SQLite failed migration left %s: count=%d err=%v", table, count, err)
		}
	}
	if err = rolledBack.Close(); err != nil {
		t.Fatalf("close rolled-back SQLite runtime: %v", err)
	}

	reapplied, err := sqlitedev.Open(ctx, sqlitedev.Options{
		DatabasePath: databasePath,
		Migrations:   Migrations(),
	})
	if err != nil {
		t.Fatalf("reapply SQLite workflow template migration: %v", err)
	}
	defer reapplied.Close()
	if err = reapplied.VerifyMigrations(ctx, Migrations()); err != nil {
		t.Fatalf("verify SQLite workflow template migration markers: %v", err)
	}
	var tableSQL string
	if err = reapplied.DB().QueryRowContext(ctx, `SELECT sql FROM sqlite_master WHERE type='table' AND name='saved_workflow_drafts'`).Scan(&tableSQL); err != nil {
		t.Fatalf("read upgraded SQLite Saved Draft schema: %v", err)
	}
	if !strings.Contains(tableSQL, "workspace_template_derivation") {
		t.Fatalf("SQLite Saved Draft provenance constraint was not upgraded: %s", tableSQL)
	}
	var markerChecksum string
	if err = reapplied.DB().QueryRowContext(ctx, `SELECT migration_checksum FROM radishmind_schema_migrations WHERE component=? AND migration_id=?`, Component, MigrationID).Scan(&markerChecksum); err != nil {
		t.Fatalf("read SQLite workflow template migration marker: %v", err)
	}
	if markerChecksum != Migrations()[4].Checksum() {
		t.Fatalf("SQLite workflow template checksum drifted: %s", markerChecksum)
	}
}
