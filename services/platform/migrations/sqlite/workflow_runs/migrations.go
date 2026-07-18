package workflowruns

import (
	_ "embed"

	"radishmind.local/services/platform/internal/sqlitedev"
)

const (
	Component                       = "workflow_runs"
	MigrationID                     = "0006_workflow_evaluation_resources"
	StoreSchemaVersion              = "workflow_run_store_sqlite_v6"
	RunRecordStoreSchemaVersion     = "workflow_runs_store_v3"
	legacyMigrationID               = "0001_workflow_runs"
	toolActionsMigrationID          = "0002_workflow_http_tool_actions"
	toolExecutionMigrationID        = "0003_workflow_http_tool_execution"
	ragSnapshotMigrationID          = "0004_workflow_rag_snapshots"
	legacyRunStoreSchemaVersion     = "workflow_runs_store_v1"
	toolActionsStoreSchemaVersion   = "workflow_runs_store_v2"
	toolExecutionStoreSchemaVersion = "workflow_runs_store_v3"
	ragSnapshotStoreSchemaVersion   = "workflow_run_store_sqlite_v4"
	ragExecutionAuditMigrationID    = "0005_workflow_rag_execution_audits"
	ragExecutionAuditSchemaVersion  = "workflow_run_store_sqlite_v5"
)

//go:embed 0001_workflow_runs.up.sql
var upSQLV1 string

//go:embed 0002_workflow_http_tool_actions.up.sql
var upSQLV2 string

//go:embed 0003_workflow_http_tool_execution.up.sql
var upSQLV3 string

//go:embed 0004_workflow_rag_snapshots.up.sql
var upSQLV4 string

//go:embed 0005_workflow_rag_execution_audits.up.sql
var upSQLV5 string

//go:embed 0006_workflow_evaluation_resources.up.sql
var upSQLV6 string

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
			ID:                 toolExecutionMigrationID,
			StoreSchemaVersion: toolExecutionStoreSchemaVersion,
			UpSQL:              upSQLV3,
		},
		{
			Component:          Component,
			ID:                 ragSnapshotMigrationID,
			StoreSchemaVersion: ragSnapshotStoreSchemaVersion,
			UpSQL:              upSQLV4,
		},
		{
			Component:          Component,
			ID:                 ragExecutionAuditMigrationID,
			StoreSchemaVersion: ragExecutionAuditSchemaVersion,
			UpSQL:              upSQLV5,
		},
		{
			Component:          Component,
			ID:                 MigrationID,
			StoreSchemaVersion: StoreSchemaVersion,
			UpSQL:              upSQLV6,
		},
	}
}
