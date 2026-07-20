package applicationdraftmigrations

import (
	"strings"
	"testing"
)

func TestApplicationConfigurationDraftMigrationContract(t *testing.T) {
	if checksum := ExpectedChecksum(); !strings.HasPrefix(checksum, "sha256:") || len(checksum) != 71 {
		t.Fatalf("unexpected migration checksum: %s", checksum)
	}
	for _, expected := range []string{
		"application_configuration_drafts",
		"tenant_ref",
		"workspace_id",
		"application_id",
		"owner_subject_ref",
		"draft_version",
		"sanitized_draft_payload",
	} {
		if !strings.Contains(upSQL, expected) {
			t.Fatalf("application draft migration is missing %s", expected)
		}
	}
	if !strings.Contains(downSQL, "DROP TABLE IF EXISTS application_configuration_drafts") {
		t.Fatal("application draft down migration must remove the draft table")
	}
}
