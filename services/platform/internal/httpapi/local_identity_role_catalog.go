package httpapi

import (
	"crypto/sha256"
	"encoding/hex"
	"slices"
	"strings"
	"time"

	"radishmind.local/services/platform/internal/workspacepolicy"
)

func localIdentityDigest(parts ...string) string {
	digest := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return "sha256:" + hex.EncodeToString(digest[:])
}

func validLocalIdentityRoleCatalogMetadata(assignment LocalRoleAssignment) bool {
	version := strings.TrimSpace(assignment.RoleCatalogVersion)
	digest := strings.TrimSpace(assignment.RoleDefinitionDigest)
	if version == "" && digest == "" {
		return assignment.RoleKey != workspacepolicy.RoleWorkspaceDeveloper &&
			!slices.ContainsFunc(assignment.PermissionGrants, func(grant string) bool {
				return strings.HasPrefix(strings.TrimSpace(grant), "development_tasks:")
			})
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
	definition workspacepolicy.RoleDefinition,
	assignment LocalRoleAssignment,
) bool {
	return assignment.RoleKey == definition.RoleKey &&
		assignment.RoleCatalogVersion == definition.CatalogVersion &&
		assignment.RoleDefinitionDigest == definition.DefinitionDigest &&
		slices.Equal(assignment.PermissionGrants, definition.PermissionGrants)
}

func localIdentityAssignmentMatchesCatalog(assignment LocalRoleAssignment) bool {
	definition, known := workspacepolicy.BuiltInRoleAtVersion(assignment.RoleCatalogVersion, assignment.RoleKey)
	return known && localIdentityRoleDefinitionMatchesAssignment(definition, assignment)
}

func localIdentityAssignmentGrantsValid(assignment LocalRoleAssignment) bool {
	if assignment.RoleCatalogVersion == "" && assignment.RoleDefinitionDigest == "" {
		return validLocalIdentityRoleCatalogMetadata(assignment)
	}
	return (assignment.RoleKey != workspacepolicy.RoleWorkspaceDeveloper || assignment.WorkspaceID != "") &&
		localIdentityAssignmentMatchesCatalog(assignment)
}

func localIdentityAssignmentCanManage(assignment LocalRoleAssignment, now time.Time) bool {
	return assignment.RoleKey == workspacepolicy.RoleWorkspaceAdmin &&
		localIdentityRoleAssignmentEffective(assignment, now) &&
		localIdentityAssignmentGrantsValid(assignment) &&
		workspacepolicy.HasManagementPermissions(assignment.PermissionGrants)
}
