package localidentitymigrations

import (
	"strings"
	"testing"
)

func TestLocalIdentityMigrationContract(t *testing.T) {
	for _, fragment := range []string{
		"CREATE TABLE local_user_accounts", "normalized_login_identifier text NOT NULL UNIQUE",
		"CREATE TABLE local_credentials", "local_credentials_active_user_idx",
		"CREATE TABLE external_identity_bindings", "external_identity_bindings_active_subject_idx",
		"CREATE TABLE local_web_sessions", "credential_digest bytea NOT NULL UNIQUE",
		"CREATE TABLE local_role_assignments", "local_role_assignments_active_scope_idx",
		"CREATE TABLE local_workspace_memberships", "local_workspace_memberships_active_scope_idx",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("up migration missing %q", fragment)
		}
	}
	for _, table := range []string{
		"local_workspace_memberships", "local_role_assignments", "local_web_sessions",
		"external_identity_bindings", "local_credentials", "local_user_accounts",
	} {
		if !strings.Contains(downSQL, "DROP TABLE IF EXISTS "+table) {
			t.Fatalf("down migration does not remove %s", table)
		}
	}
	if !strings.HasPrefix(ExpectedChecksum(), "sha256:") || len(ExpectedChecksum()) != 71 {
		t.Fatalf("unexpected migration checksum: %s", ExpectedChecksum())
	}
}
