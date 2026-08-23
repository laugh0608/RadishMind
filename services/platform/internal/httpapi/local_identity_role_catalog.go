package httpapi

import (
	"crypto/sha256"
	"encoding/hex"
	"slices"
	"strings"
	"time"
)

const (
	localIdentityRoleCatalogSchemaVersion = "local_identity_role_catalog.v1"
	localIdentityRoleCatalogVersion       = "local_identity_builtin_roles_v1"

	localIdentityRoleWorkspaceReader   = "workspace_reader"
	localIdentityRoleWorkspaceBuilder  = "workspace_builder"
	localIdentityRoleWorkspaceReviewer = "workspace_reviewer"
	localIdentityRoleWorkspaceAdmin    = "workspace_admin"

	localIdentityPermissionMembersRead      = "local_identity_members:read"
	localIdentityPermissionMembershipsWrite = "local_identity_memberships:write"
	localIdentityPermissionRolesRead        = "local_identity_roles:read"
	localIdentityPermissionRolesAssign      = "local_identity_roles:assign"
)

var localIdentityManagementPermissions = []string{
	localIdentityPermissionMembersRead,
	localIdentityPermissionMembershipsWrite,
	localIdentityPermissionRolesRead,
	localIdentityPermissionRolesAssign,
}

// LocalIdentityRoleDefinition is an immutable server-owned role policy.
// Permission grants are copied whenever the catalog leaves this package.
type LocalIdentityRoleDefinition struct {
	CatalogVersion         string   `json:"catalog_version"`
	RoleKey                string   `json:"role_key"`
	DisplayName            string   `json:"display_name"`
	Summary                string   `json:"summary"`
	PermissionGrants       []string `json:"permission_grants"`
	DefinitionDigest       string   `json:"definition_digest"`
	CanManageLocalIdentity bool     `json:"can_manage_local_identity"`
}

type LocalIdentityRoleCatalog struct {
	SchemaVersion    string                        `json:"schema_version"`
	CatalogVersion   string                        `json:"catalog_version"`
	DefinitionDigest string                        `json:"definition_digest"`
	Roles            []LocalIdentityRoleDefinition `json:"roles"`
}

var builtInLocalIdentityRoleCatalog = buildLocalIdentityRoleCatalog()

func LocalIdentityBuiltInRoleCatalog() LocalIdentityRoleCatalog {
	return cloneLocalIdentityRoleCatalog(builtInLocalIdentityRoleCatalog)
}

func builtInLocalIdentityRole(roleKey string) (LocalIdentityRoleDefinition, bool) {
	for _, definition := range builtInLocalIdentityRoleCatalog.Roles {
		if definition.RoleKey == strings.TrimSpace(roleKey) {
			return cloneLocalIdentityRoleDefinition(definition), true
		}
	}
	return LocalIdentityRoleDefinition{}, false
}

func buildLocalIdentityRoleCatalog() LocalIdentityRoleCatalog {
	reader := []string{
		"applications:read",
		"api_keys:read",
		"application_drafts:read",
		"application_evaluations:read",
		"application_sessions:read",
		"agent_copilot_profiles:read",
		"prompt_application_templates:read",
		"runs:read",
		"usage:read",
		"workflow_definitions:read",
		"workflow_drafts:read",
		"workflow_rag_evaluation_datasets:read",
		"workflow_rag_promotions:read",
		"workflow_rag_snapshots:read",
		"workflow_runs:read",
	}
	builder := mergeLocalIdentityRoleGrants(reader, []string{
		"applications:write",
		"api_keys:write",
		"application_drafts:write",
		"application_evaluations:execute",
		"application_evaluations:write",
		"application_publish_candidates:write",
		"application_sessions:execute",
		"application_sessions:write",
		"agent_copilot_profiles:bind",
		"agent_copilot_profiles:read_source",
		"agent_copilot_profiles:version",
		"agent_copilot_profiles:write",
		"agent_copilot_runtime:write",
		"prompt_application_runtime:write",
		"prompt_application_templates:bind",
		"prompt_application_templates:read_source",
		"prompt_application_templates:version",
		"prompt_application_templates:write",
		"workflow_definitions:write",
		"workflow_drafts:write",
		"workflow_evaluations:write",
		"workflow_rag:execute",
		"workflow_rag_evaluation_datasets:write",
		"workflow_rag_promotions:bind",
		"workflow_rag_promotions:write",
		"workflow_rag_runtime:write",
		"workflow_rag_snapshots:write",
		"workflow_runs:execute",
		"workflow_tool_actions:execute",
		"workflow_tool_actions:plan",
	})
	reviewer := mergeLocalIdentityRoleGrants(builder, []string{
		"application_publish_candidates:review",
		"workflow_definitions:activate",
		"workflow_definitions:review",
		"workflow_rag_evaluation_datasets:review",
		"workflow_rag_promotions:review",
		"workflow_tool_actions:confirm",
	})
	administrator := mergeLocalIdentityRoleGrants(reviewer, []string{
		"admin_gateway_pricing:read",
		"admin_gateway_pricing:write",
		"admin_gateway_quotas:read",
		"admin_gateway_quotas:write",
		"api_keys:revoke",
		"application_result_artifacts:archive",
		"application_result_artifacts:export",
		"applications:archive",
		"workflow_drafts:archive",
		"workflow_rag_evaluation_datasets:archive",
		"workflow_rag_snapshots:archive",
	}, localIdentityManagementPermissions)
	roles := []LocalIdentityRoleDefinition{
		newLocalIdentityRoleDefinition(
			localIdentityRoleWorkspaceReader,
			"Workspace reader",
			"Read workspace applications, runs, workflows, sessions, evaluations, and usage metadata.",
			reader,
			false,
		),
		newLocalIdentityRoleDefinition(
			localIdentityRoleWorkspaceBuilder,
			"Workspace builder",
			"Create and execute workspace applications, workflows, sessions, profiles, and evaluations.",
			builder,
			false,
		),
		newLocalIdentityRoleDefinition(
			localIdentityRoleWorkspaceReviewer,
			"Workspace reviewer",
			"Build workspace resources and perform explicit review, activation, and confirmation actions.",
			reviewer,
			false,
		),
		newLocalIdentityRoleDefinition(
			localIdentityRoleWorkspaceAdmin,
			"Workspace administrator",
			"Manage the workspace, destructive lifecycle actions, policy surfaces, members, and role assignments.",
			administrator,
			true,
		),
	}
	slices.SortFunc(roles, func(left, right LocalIdentityRoleDefinition) int {
		return strings.Compare(left.RoleKey, right.RoleKey)
	})
	digests := make([]string, 0, len(roles)+2)
	digests = append(digests, localIdentityRoleCatalogSchemaVersion, localIdentityRoleCatalogVersion)
	for _, role := range roles {
		digests = append(digests, role.DefinitionDigest)
	}
	return LocalIdentityRoleCatalog{
		SchemaVersion:    localIdentityRoleCatalogSchemaVersion,
		CatalogVersion:   localIdentityRoleCatalogVersion,
		DefinitionDigest: localIdentityDigest(digests...),
		Roles:            roles,
	}
}

func newLocalIdentityRoleDefinition(
	roleKey string,
	displayName string,
	summary string,
	grants []string,
	canManageLocalIdentity bool,
) LocalIdentityRoleDefinition {
	grants = mergeLocalIdentityRoleGrants(grants)
	capability := "false"
	if canManageLocalIdentity {
		capability = "true"
	}
	digestParts := []string{
		localIdentityRoleCatalogVersion,
		roleKey,
		displayName,
		summary,
		capability,
	}
	digestParts = append(digestParts, grants...)
	return LocalIdentityRoleDefinition{
		CatalogVersion:         localIdentityRoleCatalogVersion,
		RoleKey:                roleKey,
		DisplayName:            displayName,
		Summary:                summary,
		PermissionGrants:       grants,
		DefinitionDigest:       localIdentityDigest(digestParts...),
		CanManageLocalIdentity: canManageLocalIdentity,
	}
}

func mergeLocalIdentityRoleGrants(groups ...[]string) []string {
	seen := make(map[string]struct{})
	for _, group := range groups {
		for _, permission := range group {
			permission = strings.TrimSpace(permission)
			if permission != "" {
				seen[permission] = struct{}{}
			}
		}
	}
	grants := make([]string, 0, len(seen))
	for permission := range seen {
		grants = append(grants, permission)
	}
	slices.Sort(grants)
	return grants
}

func localIdentityDigest(parts ...string) string {
	digest := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return "sha256:" + hex.EncodeToString(digest[:])
}

func validLocalIdentityRoleCatalogMetadata(assignment LocalRoleAssignment) bool {
	version := strings.TrimSpace(assignment.RoleCatalogVersion)
	digest := strings.TrimSpace(assignment.RoleDefinitionDigest)
	if version == "" && digest == "" {
		return true
	}
	return version != "" && version == assignment.RoleCatalogVersion &&
		validControlPlaneReadAuthReference(version, false) &&
		len(digest) == len("sha256:")+sha256.Size*2 && strings.HasPrefix(digest, "sha256:") &&
		digest == assignment.RoleDefinitionDigest && isLowerHex(digest[len("sha256:"):])
}

func isLowerHex(value string) bool {
	if value == "" {
		return false
	}
	for _, character := range value {
		if character < '0' || character > '9' {
			if character < 'a' || character > 'f' {
				return false
			}
		}
	}
	return true
}

func localIdentityRoleDefinitionMatchesAssignment(
	definition LocalIdentityRoleDefinition,
	assignment LocalRoleAssignment,
) bool {
	return assignment.RoleKey == definition.RoleKey &&
		assignment.RoleCatalogVersion == definition.CatalogVersion &&
		assignment.RoleDefinitionDigest == definition.DefinitionDigest &&
		slices.Equal(assignment.PermissionGrants, definition.PermissionGrants)
}

func localIdentityAssignmentCanManage(assignment LocalRoleAssignment, now time.Time) bool {
	if assignment.LifecycleState != localIdentityStateActive ||
		assignment.ExpiresAt != nil && !assignment.ExpiresAt.After(now.UTC()) {
		return false
	}
	definition, exists := builtInLocalIdentityRole(assignment.RoleKey)
	if !exists || !definition.CanManageLocalIdentity {
		return false
	}
	grantSet := make(map[string]struct{}, len(assignment.PermissionGrants))
	for _, grant := range assignment.PermissionGrants {
		grantSet[grant] = struct{}{}
	}
	for _, required := range localIdentityManagementPermissions {
		if _, exists := grantSet[required]; !exists {
			return false
		}
	}
	return true
}

func localIdentityContainsManagementPermission(grants []string) bool {
	for _, permission := range localIdentityManagementPermissions {
		if slices.Contains(grants, permission) {
			return true
		}
	}
	return false
}

func cloneLocalIdentityRoleDefinition(definition LocalIdentityRoleDefinition) LocalIdentityRoleDefinition {
	definition.PermissionGrants = append([]string(nil), definition.PermissionGrants...)
	return definition
}

func cloneLocalIdentityRoleCatalog(catalog LocalIdentityRoleCatalog) LocalIdentityRoleCatalog {
	catalog.Roles = append([]LocalIdentityRoleDefinition(nil), catalog.Roles...)
	for index := range catalog.Roles {
		catalog.Roles[index] = cloneLocalIdentityRoleDefinition(catalog.Roles[index])
	}
	return catalog
}
