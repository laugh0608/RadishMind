package workflowsaveddrafts

import (
	_ "embed"

	"radishmind.local/services/platform/internal/sqlitedev"
)

const (
	Component          = "workflow_saved_drafts"
	MigrationID        = "0005_workspace_workflow_template_catalog"
	StoreSchemaVersion = "saved_workflow_drafts_store_v1"
)

//go:embed 0001_saved_workflow_drafts.up.sql
var initialUpSQL string

//go:embed 0002_saved_workflow_draft_revisions.up.sql
var revisionUpSQL string

//go:embed 0003_saved_workflow_draft_library.up.sql
var libraryUpSQL string

//go:embed 0004_saved_workflow_draft_structured_inputs.up.sql
var structuredInputUpSQL string

//go:embed 0005_workspace_workflow_template_catalog.up.sql
var workflowTemplateCatalogUpSQL string

func Migrations() []sqlitedev.Migration {
	return []sqlitedev.Migration{
		{
			Component:          Component,
			ID:                 "0001_saved_workflow_drafts",
			StoreSchemaVersion: StoreSchemaVersion,
			UpSQL:              initialUpSQL,
		},
		{
			Component:          Component,
			ID:                 "0002_saved_workflow_draft_revisions",
			StoreSchemaVersion: StoreSchemaVersion,
			UpSQL:              revisionUpSQL,
		},
		{
			Component:          Component,
			ID:                 "0003_saved_workflow_draft_library",
			StoreSchemaVersion: StoreSchemaVersion,
			UpSQL:              libraryUpSQL,
		},
		{
			Component:          Component,
			ID:                 "0004_saved_workflow_draft_structured_inputs",
			StoreSchemaVersion: StoreSchemaVersion,
			UpSQL:              structuredInputUpSQL,
		},
		{
			Component:          Component,
			ID:                 MigrationID,
			StoreSchemaVersion: StoreSchemaVersion,
			UpSQL:              workflowTemplateCatalogUpSQL,
		},
	}
}
