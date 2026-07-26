package promptapplicationtemplates

import (
	"strings"
	"testing"
)

func TestPromptApplicationTemplateSQLiteMigrationContract(t *testing.T) {
	for _, fragment := range []string{
		"CREATE TABLE prompt_application_template_drafts",
		"CREATE TABLE prompt_application_template_versions",
		"prompt_application_template_drafts_controlled_update",
		"prompt_application_template_versions_no_update",
		"prompt_application_template_draft.v1",
		"prompt_application_template_version.v1",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("SQLite template migration is missing %q", fragment)
		}
	}
}
