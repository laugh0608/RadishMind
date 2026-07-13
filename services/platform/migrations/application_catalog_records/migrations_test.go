package applicationcatalogmigrations

import (
	"strings"
	"testing"
)

func TestApplicationCatalogMigrationContract(t *testing.T) {
	for _, fragment := range []string{
		"CREATE TABLE application_catalog_records", "sanitized_record_payload jsonb", "record_version bigint",
		"lifecycle_state text", "application_catalog_records_owner_list_idx", "application_catalog_records_owner_kind_list_idx",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("up migration missing %q", fragment)
		}
	}
	if !strings.Contains(downSQL, "DROP TABLE IF EXISTS application_catalog_records") {
		t.Fatal("down migration does not remove application catalog table")
	}
	if !strings.HasPrefix(ExpectedChecksum(), "sha256:") || len(ExpectedChecksum()) != 71 {
		t.Fatalf("unexpected migration checksum: %s", ExpectedChecksum())
	}
}
