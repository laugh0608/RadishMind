package workflowruns

import (
	_ "embed"

	"radishmind.local/services/platform/internal/sqlitedev"
)

const (
	Component                     = "workflow_runs"
	MigrationID                   = "0003_workflow_http_tool_execution"
	StoreSchemaVersion            = "workflow_runs_store_v3"
	RunRecordStoreSchemaVersion   = "workflow_runs_store_v3"
	legacyMigrationID             = "0001_workflow_runs"
	toolActionsMigrationID        = "0002_workflow_http_tool_actions"
	legacyRunStoreSchemaVersion   = "workflow_runs_store_v1"
	toolActionsStoreSchemaVersion = "workflow_runs_store_v2"
)

//go:embed 0001_workflow_runs.up.sql
var upSQLV1 string

//go:embed 0002_workflow_http_tool_actions.up.sql
var upSQLV2 string

//go:embed 0003_workflow_http_tool_execution.up.sql
var upSQLV3 string

func Migrations() []sqlitedev.Migration {
	return []sqlitedev.Migration{
		{
			Component:          Component,
			ID:                 legacyMigrationID,
			StoreSchemaVersion: legacyRunStoreSchemaVersion,
			UpSQL:              upSQLV1,
		},
		{
			Component:          Component,
			ID:                 toolActionsMigrationID,
			StoreSchemaVersion: toolActionsStoreSchemaVersion,
			UpSQL:              upSQLV2,
		},
		{
			Component:          Component,
			ID:                 MigrationID,
			StoreSchemaVersion: StoreSchemaVersion,
			UpSQL:              upSQLV3,
		},
	}
}
