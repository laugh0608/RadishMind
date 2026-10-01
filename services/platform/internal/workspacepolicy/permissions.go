package workspacepolicy

import (
	"slices"
	"strings"
)

const (
	PermissionMembersRead      = "local_identity_members:read"
	PermissionMembershipsWrite = "local_identity_memberships:write"
	PermissionRolesRead        = "local_identity_roles:read"
	PermissionRolesAssign      = "local_identity_roles:assign"
)

var managementPermissions = []string{
	PermissionMembersRead,
	PermissionMembershipsWrite,
	PermissionRolesRead,
	PermissionRolesAssign,
}

var permissionAllowlist = map[string]struct{}{
	"applications:read":                        {},
	"applications:write":                       {},
	"applications:archive":                     {},
	"api_keys:read":                            {},
	"api_keys:write":                           {},
	"api_keys:revoke":                          {},
	"usage:read":                               {},
	"runs:read":                                {},
	"workflow_drafts:read":                     {},
	"workflow_drafts:write":                    {},
	"workflow_drafts:archive":                  {},
	"application_drafts:read":                  {},
	"application_drafts:write":                 {},
	"application_publish_candidates:read":      {},
	"application_publish_candidates:write":     {},
	"application_publish_candidates:review":    {},
	"workflow_definitions:write":               {},
	"workflow_definitions:review":              {},
	"workflow_definitions:activate":            {},
	"workflow_definitions:read":                {},
	"workflow_runs:execute":                    {},
	"workflow_runs:read":                       {},
	"application_sessions:read":                {},
	"application_sessions:write":               {},
	"application_sessions:execute":             {},
	"application_result_artifacts:archive":     {},
	"application_result_artifacts:export":      {},
	"prompt_application_templates:read":        {},
	"prompt_application_templates:read_source": {},
	"prompt_application_templates:write":       {},
	"prompt_application_templates:version":     {},
	"prompt_application_templates:bind":        {},
	"prompt_application_runtime:read":          {},
	"agent_copilot_profiles:read":              {},
	"agent_copilot_profiles:read_source":       {},
	"agent_copilot_profiles:write":             {},
	"agent_copilot_profiles:version":           {},
	"agent_copilot_profiles:bind":              {},
	"prompt_application_runtime:write":         {},
	"agent_copilot_runtime:write":              {},
	"workflow_rag_evaluation_datasets:read":    {},
	"workflow_rag_evaluation_datasets:write":   {},
	"workflow_rag_evaluation_datasets:review":  {},
	"workflow_rag_evaluation_datasets:archive": {},
	"workflow_rag_snapshots:read":              {},
	"workflow_rag_snapshots:write":             {},
	"workflow_rag_snapshots:archive":           {},
	"workflow_rag_promotions:read":             {},
	"workflow_rag_promotions:write":            {},
	"workflow_rag_promotions:review":           {},
	"workflow_rag_promotions:bind":             {},
	"workflow_rag_runtime:write":               {},
	"workflow_rag:execute":                     {},
	"workflow_tool_actions:plan":               {},
	"workflow_tool_actions:confirm":            {},
	"workflow_tool_actions:execute":            {},
	"workflow_evaluations:write":               {},
	"application_evaluations:read":             {},
	"application_evaluations:write":            {},
	"application_evaluations:execute":          {},
	"admin_gateway_quotas:read":                {},
	"admin_gateway_quotas:write":               {},
	"admin_gateway_pricing:read":               {},
	"admin_gateway_pricing:write":              {},
	PermissionMembersRead:                      {},
	PermissionMembershipsWrite:                 {},
	PermissionRolesRead:                        {},
	PermissionRolesAssign:                      {},
}

// NormalizeRequiredPermissions trims and deduplicates requests in first-seen order.
// Empty requests or any undeclared permission are rejected in full.
func NormalizeRequiredPermissions(required []string) ([]string, bool) {
	if len(required) == 0 {
		return nil, false
	}
	permissions := make([]string, 0, len(required))
	seen := make(map[string]struct{}, len(required))
	for _, raw := range required {
		permission := strings.TrimSpace(raw)
		if _, allowed := permissionAllowlist[permission]; !allowed {
			return nil, false
		}
		if _, duplicate := seen[permission]; duplicate {
			continue
		}
		seen[permission] = struct{}{}
		permissions = append(permissions, permission)
	}
	return permissions, len(permissions) > 0
}

// ValidPermissionGrants rejects duplicate grants after trimming; unlike request
// normalization, it must not silently repair a persisted or asserted grant list.
func ValidPermissionGrants(grants []string) bool {
	if len(grants) == 0 {
		return false
	}
	seen := make(map[string]struct{}, len(grants))
	for _, raw := range grants {
		grant := strings.TrimSpace(raw)
		if _, allowed := permissionAllowlist[grant]; !allowed {
			return false
		}
		if _, duplicate := seen[grant]; duplicate {
			return false
		}
		seen[grant] = struct{}{}
	}
	return true
}

// ContainsManagementPermission checks exact grants without normalization.
func ContainsManagementPermission(grants []string) bool {
	for _, permission := range managementPermissions {
		if slices.Contains(grants, permission) {
			return true
		}
	}
	return false
}

// PermissionNames returns a sorted copy of the declared workspace permissions.
func PermissionNames() []string {
	names := make([]string, 0, len(permissionAllowlist))
	for name := range permissionAllowlist {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

// HasManagementPermissions requires every management grant, using exact matches.
// Assignment lifecycle and role eligibility are checked by the identity owner.
func HasManagementPermissions(grants []string) bool {
	for _, required := range managementPermissions {
		if !slices.Contains(grants, required) {
			return false
		}
	}
	return true
}
