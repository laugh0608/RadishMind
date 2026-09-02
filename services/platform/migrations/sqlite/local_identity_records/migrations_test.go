package localidentityrecords

import (
	"strings"
	"testing"
)

func TestLocalIdentitySQLiteAdministrationMigrationContract(t *testing.T) {
	migrations := Migrations()
	if len(migrations) != 5 || migrations[2].ID != AdministrationMigrationID ||
		migrations[2].StoreSchemaVersion != AdministrationSchemaVersion ||
		migrations[3].ID != SelfServiceMigrationID ||
		migrations[3].StoreSchemaVersion != SelfServiceSchemaVersion ||
		migrations[4].ID != WorkspaceInvitationMigrationID ||
		migrations[4].StoreSchemaVersion != WorkspaceInvitationSchemaVersion {
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
	for _, fragment := range []string{
		"CREATE TABLE local_workspace_invitations", "secret_digest BLOB NOT NULL", "workspace_reader",
		"local_workspace_invitations_directory_idx", "updated_at_unix_nano DESC", "invitation_id DESC",
		"local_workspace_invitations_pending_expiry_idx", "expires_at_unix_nano",
	} {
		if !strings.Contains(workspaceInvitationUpSQL, fragment) {
			t.Fatalf("SQLite workspace invitation migration missing %q", fragment)
		}
	}
}
