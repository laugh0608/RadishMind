package apikeymigrations

import (
	"strings"
	"testing"
)

func TestAPIKeyMigrationContract(t *testing.T) {
	for _, fragment := range []string{
		"CREATE TABLE api_key_records", "credential_digest bytea", "sanitized_record_payload jsonb",
		"UNIQUE (api_key_id)", "UNIQUE (credential_digest)", "api_key_records_owner_list_idx",
		"api_key_records_owner_application_list_idx", "api_key_records_authentication_idx",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("up migration missing %q", fragment)
		}
	}
	if !strings.Contains(downSQL, "DROP TABLE IF EXISTS api_key_records") {
		t.Fatal("down migration does not remove API key table")
	}
	if !strings.HasPrefix(ExpectedChecksum(), "sha256:") || len(ExpectedChecksum()) != 71 {
		t.Fatalf("unexpected migration checksum: %s", ExpectedChecksum())
	}
}
