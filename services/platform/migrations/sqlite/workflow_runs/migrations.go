package workflowruns

import (
	_ "embed"

	"radishmind.local/services/platform/internal/sqlitedev"
)

const (
	Component          = "workflow_runs"
	MigrationID        = "0001_workflow_runs"
	StoreSchemaVersion = "workflow_runs_store_v1"
)

//go:embed 0001_workflow_runs.up.sql
var upSQL string

func Migrations() []sqlitedev.Migration {
	return []sqlitedev.Migration{{
		Component:          Component,
		ID:                 MigrationID,
		StoreSchemaVersion: StoreSchemaVersion,
		UpSQL:              upSQL,
	}}
}
