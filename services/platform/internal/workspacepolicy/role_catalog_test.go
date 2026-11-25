package workspacepolicy

import (
	"encoding/json"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"testing"
)

func TestBuiltInRoleLookupAndCatalogCopies(t *testing.T) {
	baseline := BuiltInRoleCatalog()
	for _, want := range baseline.Roles {
		got, exists := BuiltInRole(" " + want.RoleKey + "\t")
		if !exists || !reflect.DeepEqual(got, want) {
			t.Fatalf("role lookup mismatch: %q", want.RoleKey)
		}
		got.PermissionGrants[0] = "unknown:write"
		got.DisplayName = "mutated"
		fresh, _ := BuiltInRole(want.RoleKey)
		if !reflect.DeepEqual(fresh, want) {
			t.Fatal("single role lookup exposed canonical state")
		}
	}
	for _, key := range []string{"", " \t", "unknown", "WORKSPACE_ADMIN"} {
		got, exists := BuiltInRole(key)
		if exists || !reflect.DeepEqual(got, RoleDefinition{}) {
			t.Fatalf("unknown role accepted: %q", key)
		}
	}
	copy := BuiltInRoleCatalog()
	for i := range copy.Roles {
		copy.Roles[i].PermissionGrants[0] = "unknown:write"
		copy.Roles[i].RoleKey = "mutated"
	}
	copy.Roles = append(copy.Roles, RoleDefinition{RoleKey: "extra"})
	if !reflect.DeepEqual(BuiltInRoleCatalog(), baseline) {
		t.Fatal("catalog lookup exposed canonical role storage")
	}
}

func TestRoleCatalogJSONContract(t *testing.T) {
	encoded, err := json.Marshal(BuiltInRoleCatalog())
	if err != nil {
		t.Fatal(err)
	}
	var catalog map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &catalog); err != nil {
		t.Fatal(err)
	}
	assertJSONKeys := func(object map[string]json.RawMessage, keys ...string) {
		t.Helper()
		if len(object) != len(keys) {
			t.Fatalf("unexpected JSON fields: %s", encoded)
		}
		for _, key := range keys {
			if _, exists := object[key]; !exists {
				t.Fatalf("missing JSON field %q", key)
			}
		}
	}
	assertJSONKeys(catalog, "schema_version", "catalog_version", "definition_digest", "roles")
	var roles []map[string]json.RawMessage
	if err := json.Unmarshal(catalog["roles"], &roles); err != nil {
		t.Fatal(err)
	}
	if len(roles) != 5 {
		t.Fatalf("unexpected role count: %d", len(roles))
	}
	for _, role := range roles {
		assertJSONKeys(role, "catalog_version", "role_key", "display_name", "summary", "permission_grants", "definition_digest", "can_manage_local_identity")
	}
}

func TestLocalIdentityBuiltInRoleCatalogContract(t *testing.T) {
	catalog := BuiltInRoleCatalog()
	if catalog.DefinitionDigest != "sha256:7a9265bf9b1883c199d177e4fdc2014b420282e1d5b960138e4a7ac6df8225a1" {
		t.Fatalf("role catalog changed without an explicit catalog version decision: %s", catalog.DefinitionDigest)
	}
	if catalog.SchemaVersion != RoleCatalogSchemaVersion ||
		catalog.CatalogVersion != RoleCatalogVersion ||
		len(catalog.Roles) != 5 || !strings.HasPrefix(catalog.DefinitionDigest, "sha256:") {
		t.Fatalf("role catalog contract drifted: %#v", catalog)
	}
	byKey := make(map[string]RoleDefinition, len(catalog.Roles))
	for _, definition := range catalog.Roles {
		if definition.CatalogVersion != catalog.CatalogVersion ||
			!regexp.MustCompile(`^[a-z][a-z0-9_]{2,63}$`).MatchString(definition.RoleKey) ||
			definition.DisplayName == "" || definition.Summary == "" ||
			!slices.IsSorted(definition.PermissionGrants) ||
			len(definition.PermissionGrants) == 0 || !strings.HasPrefix(definition.DefinitionDigest, "sha256:") {
			t.Fatalf("invalid role definition: %#v", definition)
		}
		if !ValidPermissionGrants(definition.PermissionGrants) {
			t.Fatalf("invalid role grants: %#v", definition)
		}
		if _, duplicate := byKey[definition.RoleKey]; duplicate {
			t.Fatalf("duplicate role key: %s", definition.RoleKey)
		}
		byKey[definition.RoleKey] = definition
	}
	reader := byKey[RoleWorkspaceReader]
	builder := byKey[RoleWorkspaceBuilder]
	reviewer := byKey[RoleWorkspaceReviewer]
	administrator := byKey[RoleWorkspaceAdmin]
	for _, permission := range []string{"application_publish_candidates:read", "prompt_application_runtime:read"} {
		if !slices.Contains(reader.PermissionGrants, permission) {
			t.Fatalf("workspace_reader must cover owner read permission %s", permission)
		}
	}
	if !localIdentityGrantSubset(reader.PermissionGrants, builder.PermissionGrants) ||
		!localIdentityGrantSubset(builder.PermissionGrants, reviewer.PermissionGrants) ||
		!localIdentityGrantSubset(reviewer.PermissionGrants, administrator.PermissionGrants) {
		t.Fatal("built-in role grants are not cumulative")
	}
	allowed := slices.DeleteFunc(PermissionNames(), func(p string) bool { return strings.HasPrefix(p, "development_") })
	if !slices.Equal(administrator.PermissionGrants, allowed) {
		t.Fatalf("workspace_admin grants drifted:\nwant=%v\n got=%v", allowed, administrator.PermissionGrants)
	}
	wantDeveloperGrants := []string{
		"applications:read",
		"development_artifacts:export",
		"development_artifacts:read",
		"development_artifacts:save",
		"development_tasks:cancel",
		"development_tasks:execute",
		"development_tasks:read",
		"development_tasks:write",
	}
	if developer := byKey[RoleWorkspaceDeveloper]; !slices.Equal(developer.PermissionGrants, wantDeveloperGrants) {
		t.Fatalf("workspace_developer grant drift: %v", developer.PermissionGrants)
	}
	for _, definition := range catalog.Roles {
		wantManagement := definition.RoleKey == RoleWorkspaceAdmin
		if definition.CanManageLocalIdentity != wantManagement {
			t.Fatalf("identity management capability drifted for %s", definition.RoleKey)
		}
		for _, permission := range []string{PermissionMembersRead, PermissionMembershipsWrite, PermissionRolesRead, PermissionRolesAssign} {
			if slices.Contains(definition.PermissionGrants, permission) != wantManagement {
				t.Fatalf("management permission %s leaked into role %s", permission, definition.RoleKey)
			}
		}
	}
	catalog.Roles[0].PermissionGrants[0] = "applications:archive"
	if BuiltInRoleCatalog().Roles[0].PermissionGrants[0] == "applications:archive" {
		t.Fatal("role catalog caller mutated canonical grants")
	}
}

func TestHistoricalRoleCatalogsPreserveFrozenGrants(t *testing.T) {
	for _, test := range []struct {
		catalog RoleCatalog
		digest  string
	}{
		{builtInRoleCatalogV1, "sha256:d784ef5d5595f4fa3ed96f32c86f3fd12edbd4098da46668366f97ce42e2d4d0"},
		{builtInRoleCatalogV2, "sha256:44c8a3a41eb90b2da25859662abf13ba91cef00505eb5191767dc5b17eb4abae"},
	} {
		if test.catalog.DefinitionDigest != test.digest || len(test.catalog.Roles) != 4 {
			t.Fatalf("original catalog %s drifted: %s", test.catalog.CatalogVersion, test.catalog.DefinitionDigest)
		}
		if _, ok := BuiltInRoleAtVersion(test.catalog.CatalogVersion, RoleWorkspaceDeveloper); ok {
			t.Fatalf("developer exists in historical catalog %s", test.catalog.CatalogVersion)
		}
		for _, want := range test.catalog.Roles {
			got, ok := BuiltInRoleAtVersion(test.catalog.CatalogVersion, want.RoleKey)
			if !ok || !reflect.DeepEqual(got, want) {
				t.Fatalf("incorrect versioned lookup: %s/%s", test.catalog.CatalogVersion, want.RoleKey)
			}
			got.PermissionGrants[0] = "unknown:write"
			fresh, _ := BuiltInRoleAtVersion(test.catalog.CatalogVersion, want.RoleKey)
			if !reflect.DeepEqual(fresh, want) {
				t.Fatal("versioned lookup exposed canonical grant storage")
			}
			if test.catalog.CatalogVersion == roleCatalogVersionV2 {
				current, _ := BuiltInRole(want.RoleKey)
				if !slices.Equal(current.PermissionGrants, want.PermissionGrants) {
					t.Fatalf("v3 changed deployed grants for %s", want.RoleKey)
				}
			}
		}
	}
	if _, ok := BuiltInRoleAtVersion("local_identity_builtin_roles_v99", RoleWorkspaceAdmin); ok {
		t.Fatal("unknown version resolved a role")
	}
}

func localIdentityGrantSubset(subset []string, superset []string) bool {
	for _, grant := range subset {
		if !slices.Contains(superset, grant) {
			return false
		}
	}
	return true
}
