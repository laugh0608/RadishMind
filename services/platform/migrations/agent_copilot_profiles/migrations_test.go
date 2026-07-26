package agentcopilotprofilemigrations

import (
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestAgentCopilotProfileMigrationContract(t *testing.T) {
	for _, fragment := range []string{
		"CREATE TABLE agent_copilot_profile_drafts",
		"CREATE TABLE agent_copilot_profile_versions",
		"agent_copilot_profile_drafts_controlled_update",
		"agent_copilot_profile_versions_no_update",
		"agent_copilot_profile_versions_no_delete",
		"agent_copilot_profile_drafts_no_delete",
		"agent_copilot_profile_draft.v1",
		"agent_copilot_profile_version.v1",
		"policy_digest",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("PostgreSQL profile migration is missing %q", fragment)
		}
	}
	for _, fragment := range []string{"DROP TABLE IF EXISTS agent_copilot_profile_versions", "DROP TABLE IF EXISTS agent_copilot_profile_drafts"} {
		if !strings.Contains(downSQL, fragment) {
			t.Fatalf("PostgreSQL profile rollback is missing %q", fragment)
		}
	}
	if checksum := ExpectedChecksum(); !strings.HasPrefix(checksum, "sha256:") || len(checksum) != 71 {
		t.Fatalf("unexpected migration checksum: %s", checksum)
	}
}

func TestAgentCopilotProfileDatabaseErrorsAreSanitized(t *testing.T) {
	secret := "postgresql://user:secret@example.invalid/database"
	if got := safeDatabaseError("connect", errors.New(secret)).Error(); strings.Contains(got, "secret") {
		t.Fatalf("database material leaked: %s", got)
	}
	if got := safeDatabaseError("query", &pgconn.PgError{Code: "23505", Message: secret}).Error(); got != "query failed (SQLSTATE 23505)" {
		t.Fatalf("unexpected sanitized PostgreSQL error: %s", got)
	}
}
