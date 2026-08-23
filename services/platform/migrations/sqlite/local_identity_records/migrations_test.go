package localidentityrecords

import (
	"strings"
	"testing"
)

func TestLocalIdentitySQLiteAdministrationMigrationContract(t *testing.T) {
	migrations := Migrations()
	if len(migrations) != 3 || migrations[2].ID != AdministrationMigrationID ||
		migrations[2].StoreSchemaVersion != AdministrationSchemaVersion {
		t.Fatalf("unexpected SQLite local identity migration order: %#v", migrations)
	}
	for _, fragment := range []string{
		"ADD COLUMN role_catalog_version", "ADD COLUMN role_definition_digest",
		"length(role_definition_digest) = 71", "local_workspace_memberships_directory_idx",
		"updated_at_unix_nano DESC", "membership_id DESC",
	} {
		if !strings.Contains(administrationUpSQL, fragment) {
			t.Fatalf("SQLite administration migration missing %q", fragment)
		}
	}
}
