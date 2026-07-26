package promptapplicationtemplatemigrations

import (
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestPromptApplicationTemplateMigrationContract(t *testing.T) {
	for _, fragment := range []string{
		"CREATE TABLE prompt_application_template_drafts",
		"CREATE TABLE prompt_application_template_versions",
		"prompt_application_template_drafts_controlled_update",
		"prompt_application_template_versions_no_update",
		"prompt_application_template_versions_no_delete",
		"prompt_application_template_drafts_no_delete",
		"prompt_application_template_draft.v1",
		"prompt_application_template_version.v1",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("PostgreSQL template migration is missing %q", fragment)
		}
	}
	for _, fragment := range []string{"DROP TABLE IF EXISTS prompt_application_template_versions", "DROP TABLE IF EXISTS prompt_application_template_drafts"} {
		if !strings.Contains(downSQL, fragment) {
			t.Fatalf("PostgreSQL template rollback is missing %q", fragment)
		}
	}
	if checksum := ExpectedChecksum(); !strings.HasPrefix(checksum, "sha256:") || len(checksum) != 71 {
		t.Fatalf("unexpected migration checksum: %s", checksum)
	}
}

func TestPromptApplicationTemplateDatabaseErrorsAreSanitized(t *testing.T) {
	secret := "postgresql://user:secret@example.invalid/database"
	if got := safeDatabaseError("connect", errors.New(secret)).Error(); strings.Contains(got, "secret") {
		t.Fatalf("database material leaked: %s", got)
	}
	if got := safeDatabaseError("query", &pgconn.PgError{Code: "23505", Message: secret}).Error(); got != "query failed (SQLSTATE 23505)" {
		t.Fatalf("unexpected sanitized PostgreSQL error: %s", got)
	}
}
