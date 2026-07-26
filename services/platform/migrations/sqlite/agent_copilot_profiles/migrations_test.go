package agentcopilotprofiles

import (
	"strings"
	"testing"
)

func TestAgentCopilotProfileSQLiteMigrationContract(t *testing.T) {
	for _, fragment := range []string{
		"CREATE TABLE agent_copilot_profile_drafts",
		"CREATE TABLE agent_copilot_profile_versions",
		"agent_copilot_profile_drafts_controlled_update",
		"agent_copilot_profile_versions_no_update",
		"agent_copilot_profile_draft.v1",
		"agent_copilot_profile_version.v1",
		"policy_digest",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("SQLite profile migration is missing %q", fragment)
		}
	}
}
