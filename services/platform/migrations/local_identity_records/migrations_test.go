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
		if !strings.Contains(legacyUpSQL, fragment) {
			t.Fatalf("up migration missing %q", fragment)
		}
	}
	for _, fragment := range []string{"CREATE TABLE local_identity_oidc_authorization_transactions", "state_digest bytea NOT NULL UNIQUE", "code_verifier text NOT NULL"} {
		if !strings.Contains(oidcAuthorizationUpSQL, fragment) {
			t.Fatalf("OIDC authorization migration missing %q", fragment)
		}
	}
	for _, fragment := range []string{"ADD COLUMN role_catalog_version", "ADD COLUMN role_definition_digest", "local_workspace_memberships_directory_idx"} {
		if !strings.Contains(administrationUpSQL, fragment) {
			t.Fatalf("administration migration missing %q", fragment)
		}
	}
	for _, table := range []string{
		"local_workspace_memberships", "local_role_assignments", "local_web_sessions",
		"external_identity_bindings", "local_credentials", "local_user_accounts",
	} {
		if !strings.Contains(legacyDownSQL, "DROP TABLE IF EXISTS "+table) {
			t.Fatalf("down migration does not remove %s", table)
		}
	}
	if !strings.Contains(oidcAuthorizationDownSQL, "DROP TABLE IF EXISTS local_identity_oidc_authorization_transactions") {
		t.Fatal("OIDC authorization down migration does not remove its table")
	}
	if !strings.Contains(administrationDownSQL, "DROP INDEX IF EXISTS local_workspace_memberships_directory_idx") ||
		!strings.Contains(administrationDownSQL, "DROP COLUMN IF EXISTS role_catalog_version") {
		t.Fatal("administration down migration does not remove its index and columns")
	}
	if !strings.HasPrefix(ExpectedChecksum(), "sha256:") || len(ExpectedChecksum()) != 71 {
		t.Fatalf("unexpected migration checksum: %s", ExpectedChecksum())
	}
}
