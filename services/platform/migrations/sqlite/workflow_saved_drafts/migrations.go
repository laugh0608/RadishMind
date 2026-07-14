package workflowsaveddrafts

import (
	_ "embed"

	"radishmind.local/services/platform/internal/sqlitedev"
)

const (
	Component          = "workflow_saved_drafts"
	MigrationID        = "0001_saved_workflow_drafts"
	StoreSchemaVersion = "saved_workflow_drafts_store_v1"
)

//go:embed 0001_saved_workflow_drafts.up.sql
var upSQL string

func Migrations() []sqlitedev.Migration {
	return []sqlitedev.Migration{{
		Component:          Component,
		ID:                 MigrationID,
		StoreSchemaVersion: StoreSchemaVersion,
		UpSQL:              upSQL,
	}}
}
