package workflowsaveddrafts

import (
	_ "embed"

	"radishmind.local/services/platform/internal/sqlitedev"
)

const (
	Component          = "workflow_saved_drafts"
	MigrationID        = "0002_saved_workflow_draft_revisions"
	StoreSchemaVersion = "saved_workflow_drafts_store_v1"
)

//go:embed 0001_saved_workflow_drafts.up.sql
var initialUpSQL string

//go:embed 0002_saved_workflow_draft_revisions.up.sql
var revisionUpSQL string

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
			ID:                 MigrationID,
			StoreSchemaVersion: StoreSchemaVersion,
			UpSQL:              revisionUpSQL,
		},
	}
}
