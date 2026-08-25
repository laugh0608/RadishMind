package localidentityrecords

import (
	"strings"
	"testing"
)

func TestLocalIdentitySQLiteAdministrationMigrationContract(t *testing.T) {
	migrations := Migrations()
	if len(migrations) != 4 || migrations[2].ID != AdministrationMigrationID ||
		migrations[2].StoreSchemaVersion != AdministrationSchemaVersion ||
		migrations[3].ID != SelfServiceMigrationID ||
		migrations[3].StoreSchemaVersion != SelfServiceSchemaVersion {
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
	for _, fragment := range []string{
		"local_web_sessions_self_service_list_idx", "user_id", "created_at_unix_nano DESC", "session_id DESC",
	} {
		if !strings.Contains(selfServiceUpSQL, fragment) {
			t.Fatalf("SQLite self-service migration missing %q", fragment)
		}
	}
}
