package workflowsaveddrafts

import (
	"strings"
	"testing"
)

func TestSQLiteSavedWorkflowDraftMigrationChainIncludesLibraryLifecycle(t *testing.T) {
	migrations := Migrations()
	if len(migrations) != 3 || migrations[len(migrations)-1].ID != MigrationID {
		t.Fatalf("unexpected SQLite saved draft migration chain: %#v", migrations)
	}
	for _, literal := range []string{
		"saved_workflow_draft_lifecycle_events",
		"saved_workflow_drafts_owner_lifecycle_list_idx",
		"saved_workflow_drafts_validation_list_idx",
		"saved_workflow_drafts_provenance_list_idx",
		"saved_workflow_drafts_name_list_idx",
		"library_updated_at_unix_nano = updated_at_unix_nano",
		"provenance_kind = CASE",
	} {
		if !strings.Contains(libraryUpSQL, literal) {
			t.Fatalf("SQLite library lifecycle migration is missing %q", literal)
		}
	}
}
