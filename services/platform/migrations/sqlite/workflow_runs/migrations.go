package workflowruns

import (
	_ "embed"

	"radishmind.local/services/platform/internal/sqlitedev"
)

const (
	Component                        = "workflow_runs"
	MigrationID                      = "0012_application_interaction_sessions"
	StoreSchemaVersion               = "workflow_run_store_sqlite_v12"
	RunRecordStoreSchemaVersion      = "workflow_runs_store_v5"
	legacyMigrationID                = "0001_workflow_runs"
	toolActionsMigrationID           = "0002_workflow_http_tool_actions"
	toolExecutionMigrationID         = "0003_workflow_http_tool_execution"
	ragSnapshotMigrationID           = "0004_workflow_rag_snapshots"
	legacyRunStoreSchemaVersion      = "workflow_runs_store_v1"
	toolActionsStoreSchemaVersion    = "workflow_runs_store_v2"
	toolExecutionStoreSchemaVersion  = "workflow_runs_store_v3"
	ragSnapshotStoreSchemaVersion    = "workflow_run_store_sqlite_v4"
	ragExecutionAuditMigrationID     = "0005_workflow_rag_execution_audits"
	ragExecutionAuditSchemaVersion   = "workflow_run_store_sqlite_v5"
	evaluationResourcesMigrationID   = "0006_workflow_evaluation_resources"
	evaluationResourcesSchemaVersion = "workflow_run_store_sqlite_v6"
	ragEvaluationMigrationID         = "0007_workflow_rag_evaluation_datasets"
	ragEvaluationSchemaVersion       = "workflow_run_store_sqlite_v7"
	ragPromotionMigrationID          = "0008_workflow_rag_knowledge_promotions"
	ragPromotionSchemaVersion        = "workflow_run_store_sqlite_v8"
	applicationRuntimeMigrationID    = "0009_workflow_rag_application_invocations"
	applicationRuntimeSchemaVersion  = "workflow_run_store_sqlite_v9"
	definitionReleaseMigrationID     = "0010_workflow_definition_releases"
	definitionReleaseSchemaVersion   = "workflow_run_store_sqlite_v10"
	definitionExecutionMigrationID   = "0011_workflow_definition_execution"
	definitionExecutionSchemaVersion = "workflow_run_store_sqlite_v11"
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

//go:embed 0007_workflow_rag_evaluation_datasets.up.sql
var upSQLV7 string

//go:embed 0008_workflow_rag_knowledge_promotions.up.sql
var upSQLV8 string

//go:embed 0009_workflow_rag_application_invocations.up.sql
var upSQLV9 string

//go:embed 0010_workflow_definition_releases.up.sql
var upSQLV10 string

//go:embed 0011_workflow_definition_execution.up.sql
var upSQLV11 string

//go:embed 0012_application_interaction_sessions.up.sql
var upSQLV12 string

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
			ID:                 evaluationResourcesMigrationID,
			StoreSchemaVersion: evaluationResourcesSchemaVersion,
			UpSQL:              upSQLV6,
		},
		{Component: Component, ID: ragEvaluationMigrationID, StoreSchemaVersion: ragEvaluationSchemaVersion, UpSQL: upSQLV7},
		{Component: Component, ID: ragPromotionMigrationID, StoreSchemaVersion: ragPromotionSchemaVersion, UpSQL: upSQLV8},
		{Component: Component, ID: applicationRuntimeMigrationID, StoreSchemaVersion: applicationRuntimeSchemaVersion, UpSQL: upSQLV9},
		{Component: Component, ID: definitionReleaseMigrationID, StoreSchemaVersion: definitionReleaseSchemaVersion, UpSQL: upSQLV10},
		{Component: Component, ID: definitionExecutionMigrationID, StoreSchemaVersion: definitionExecutionSchemaVersion, UpSQL: upSQLV11},
		{Component: Component, ID: MigrationID, StoreSchemaVersion: StoreSchemaVersion, UpSQL: upSQLV12},
	}
}
