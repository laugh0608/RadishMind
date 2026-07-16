package workflowruns

import (
	_ "embed"

	"radishmind.local/services/platform/internal/sqlitedev"
)

const (
	Component                   = "workflow_runs"
	MigrationID                 = "0002_workflow_http_tool_actions"
	StoreSchemaVersion          = "workflow_runs_store_v2"
	RunRecordStoreSchemaVersion = "workflow_runs_store_v1"
	legacyMigrationID           = "0001_workflow_runs"
)

//go:embed 0001_workflow_runs.up.sql
var upSQLV1 string

//go:embed 0002_workflow_http_tool_actions.up.sql
var upSQLV2 string

func Migrations() []sqlitedev.Migration {
	return []sqlitedev.Migration{
		{
			Component:          Component,
			ID:                 legacyMigrationID,
			StoreSchemaVersion: RunRecordStoreSchemaVersion,
			UpSQL:              upSQLV1,
		},
		{
			Component:          Component,
			ID:                 MigrationID,
			StoreSchemaVersion: StoreSchemaVersion,
			UpSQL:              upSQLV2,
		},
	}
}
