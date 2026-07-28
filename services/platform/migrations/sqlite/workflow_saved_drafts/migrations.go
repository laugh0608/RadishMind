package workflowsaveddrafts

import (
	_ "embed"

	"radishmind.local/services/platform/internal/sqlitedev"
)

const (
	Component          = "workflow_saved_drafts"
	MigrationID        = "0003_saved_workflow_draft_library"
	StoreSchemaVersion = "saved_workflow_drafts_store_v1"
)

//go:embed 0001_saved_workflow_drafts.up.sql
var initialUpSQL string

//go:embed 0002_saved_workflow_draft_revisions.up.sql
var revisionUpSQL string

//go:embed 0003_saved_workflow_draft_library.up.sql
var libraryUpSQL string

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
			ID:                 MigrationID,
			StoreSchemaVersion: StoreSchemaVersion,
			UpSQL:              libraryUpSQL,
		},
	}
}
