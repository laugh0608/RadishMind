package workspacepolicy

import (
	"crypto/sha256"
	"encoding/hex"
	"slices"
	"strings"
)

const (
	RoleCatalogSchemaVersion = "local_identity_role_catalog.v1"
	RoleCatalogVersion       = "local_identity_builtin_roles_v2"

	RoleWorkspaceReader   = "workspace_reader"
	RoleWorkspaceBuilder  = "workspace_builder"
	RoleWorkspaceReviewer = "workspace_reviewer"
	RoleWorkspaceAdmin    = "workspace_admin"
)

// RoleDefinition is an immutable server-owned role policy.
// Permission grants are copied whenever the catalog leaves this package.
type RoleDefinition struct {
	CatalogVersion         string   `json:"catalog_version"`
	RoleKey                string   `json:"role_key"`
	DisplayName            string   `json:"display_name"`
	Summary                string   `json:"summary"`
	PermissionGrants       []string `json:"permission_grants"`
	DefinitionDigest       string   `json:"definition_digest"`
	CanManageLocalIdentity bool     `json:"can_manage_local_identity"`
}

type RoleCatalog struct {
	SchemaVersion    string           `json:"schema_version"`
	CatalogVersion   string           `json:"catalog_version"`
	DefinitionDigest string           `json:"definition_digest"`
	Roles            []RoleDefinition `json:"roles"`
}

var builtInRoleCatalog = buildRoleCatalog()

// BuiltInRoleCatalog returns an independent copy of the canonical directory.
func BuiltInRoleCatalog() RoleCatalog {
	return cloneRoleCatalog(builtInRoleCatalog)
}

// BuiltInRole looks up a trimmed, case-sensitive key and copies its grants.
func BuiltInRole(roleKey string) (RoleDefinition, bool) {
	for _, definition := range builtInRoleCatalog.Roles {
		if definition.RoleKey == strings.TrimSpace(roleKey) {
			return cloneRoleDefinition(definition), true
		}
	}
	return RoleDefinition{}, false
}

func buildRoleCatalog() RoleCatalog {
	reader := []string{
		"applications:read",
		"api_keys:read",
		"application_drafts:read",
		"application_evaluations:read",
		"application_publish_candidates:read",
		"application_sessions:read",
		"agent_copilot_profiles:read",
		"prompt_application_templates:read",
		"prompt_application_runtime:read",
		"runs:read",
		"usage:read",
		"workflow_definitions:read",
		"workflow_drafts:read",
		"workflow_rag_evaluation_datasets:read",
		"workflow_rag_promotions:read",
		"workflow_rag_snapshots:read",
		"workflow_runs:read",
	}
	builder := mergeRoleGrants(reader, []string{
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
	reviewer := mergeRoleGrants(builder, []string{
		"application_publish_candidates:review",
		"workflow_definitions:activate",
		"workflow_definitions:review",
		"workflow_rag_evaluation_datasets:review",
		"workflow_rag_promotions:review",
		"workflow_tool_actions:confirm",
	})
	administrator := mergeRoleGrants(reviewer, []string{
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
	}, managementPermissions)
	roles := []RoleDefinition{
		newRoleDefinition(
			RoleWorkspaceReader,
			"Workspace reader",
			"Read workspace applications, runs, workflows, sessions, evaluations, and usage metadata.",
			reader,
			false,
		),
		newRoleDefinition(
			RoleWorkspaceBuilder,
			"Workspace builder",
			"Create and execute workspace applications, workflows, sessions, profiles, and evaluations.",
			builder,
			false,
		),
		newRoleDefinition(
			RoleWorkspaceReviewer,
			"Workspace reviewer",
			"Build workspace resources and perform explicit review, activation, and confirmation actions.",
			reviewer,
			false,
		),
		newRoleDefinition(
			RoleWorkspaceAdmin,
			"Workspace administrator",
			"Manage the workspace, destructive lifecycle actions, policy surfaces, members, and role assignments.",
			administrator,
			true,
		),
	}
	slices.SortFunc(roles, func(left, right RoleDefinition) int {
		return strings.Compare(left.RoleKey, right.RoleKey)
	})
	digests := make([]string, 0, len(roles)+2)
	digests = append(digests, RoleCatalogSchemaVersion, RoleCatalogVersion)
	for _, role := range roles {
		digests = append(digests, role.DefinitionDigest)
	}
	return RoleCatalog{
		SchemaVersion:    RoleCatalogSchemaVersion,
		CatalogVersion:   RoleCatalogVersion,
		DefinitionDigest: catalogDigest(digests...),
		Roles:            roles,
	}
}

func newRoleDefinition(
	roleKey string,
	displayName string,
	summary string,
	grants []string,
	canManageLocalIdentity bool,
) RoleDefinition {
	grants = mergeRoleGrants(grants)
	capability := "false"
	if canManageLocalIdentity {
		capability = "true"
	}
	digestParts := []string{
		RoleCatalogVersion,
		roleKey,
		displayName,
		summary,
		capability,
	}
	digestParts = append(digestParts, grants...)
	return RoleDefinition{
		CatalogVersion:         RoleCatalogVersion,
		RoleKey:                roleKey,
		DisplayName:            displayName,
		Summary:                summary,
		PermissionGrants:       grants,
		DefinitionDigest:       catalogDigest(digestParts...),
		CanManageLocalIdentity: canManageLocalIdentity,
	}
}

func mergeRoleGrants(groups ...[]string) []string {
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

func catalogDigest(parts ...string) string {
	digest := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return "sha256:" + hex.EncodeToString(digest[:])
}

func cloneRoleDefinition(definition RoleDefinition) RoleDefinition {
	definition.PermissionGrants = append([]string(nil), definition.PermissionGrants...)
	return definition
}

func cloneRoleCatalog(catalog RoleCatalog) RoleCatalog {
	catalog.Roles = append([]RoleDefinition(nil), catalog.Roles...)
	for index := range catalog.Roles {
		catalog.Roles[index] = cloneRoleDefinition(catalog.Roles[index])
	}
	return catalog
}
