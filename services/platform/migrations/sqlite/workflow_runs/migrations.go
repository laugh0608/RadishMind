package workflowruns

import (
	_ "embed"

	"radishmind.local/services/platform/internal/sqlitedev"
)

const (
	Component                          = "workflow_runs"
	MigrationID                        = "0026_application_evaluation_schedules"
	StoreSchemaVersion                 = "workflow_run_store_sqlite_v26"
	RunRecordStoreSchemaVersion        = "workflow_runs_store_v7"
	legacyMigrationID                  = "0001_workflow_runs"
	toolActionsMigrationID             = "0002_workflow_http_tool_actions"
	toolExecutionMigrationID           = "0003_workflow_http_tool_execution"
	ragSnapshotMigrationID             = "0004_workflow_rag_snapshots"
	legacyRunStoreSchemaVersion        = "workflow_runs_store_v1"
	toolActionsStoreSchemaVersion      = "workflow_runs_store_v2"
	toolExecutionStoreSchemaVersion    = "workflow_runs_store_v3"
	ragSnapshotStoreSchemaVersion      = "workflow_run_store_sqlite_v4"
	ragExecutionAuditMigrationID       = "0005_workflow_rag_execution_audits"
	ragExecutionAuditSchemaVersion     = "workflow_run_store_sqlite_v5"
	evaluationResourcesMigrationID     = "0006_workflow_evaluation_resources"
	evaluationResourcesSchemaVersion   = "workflow_run_store_sqlite_v6"
	ragEvaluationMigrationID           = "0007_workflow_rag_evaluation_datasets"
	ragEvaluationSchemaVersion         = "workflow_run_store_sqlite_v7"
	ragPromotionMigrationID            = "0008_workflow_rag_knowledge_promotions"
	ragPromotionSchemaVersion          = "workflow_run_store_sqlite_v8"
	applicationRuntimeMigrationID      = "0009_workflow_rag_application_invocations"
	applicationRuntimeSchemaVersion    = "workflow_run_store_sqlite_v9"
	definitionReleaseMigrationID       = "0010_workflow_definition_releases"
	definitionReleaseSchemaVersion     = "workflow_run_store_sqlite_v10"
	definitionExecutionMigrationID     = "0011_workflow_definition_execution"
	definitionExecutionSchemaVersion   = "workflow_run_store_sqlite_v11"
	applicationSessionMigrationID      = "0012_application_interaction_sessions"
	applicationSessionSchemaVersion    = "workflow_run_store_sqlite_v12"
	promptRuntimeMigrationID           = "0013_prompt_application_runtime_projections"
	promptRuntimeSchemaVersion         = "workflow_run_store_sqlite_v13"
	agentRuntimeMigrationID            = "0014_agent_copilot_runtime_assignments"
	agentRuntimeSchemaVersion          = "workflow_run_store_sqlite_v14"
	agentInvocationMigrationID         = "0015_agent_copilot_invocation_projections"
	agentInvocationSchemaVersion       = "workflow_run_store_sqlite_v15"
	applicationEvaluationMigrationID   = "0016_application_evaluation_campaigns"
	applicationEvaluationSchemaVersion = "workflow_run_store_sqlite_v16"
	structuredDefinitionMigrationID    = "0017_workflow_definition_structured_inputs"
	structuredDefinitionSchemaVersion  = "workflow_run_store_sqlite_v17"
	structuredSessionMigrationID       = "0018_application_structured_sessions"
	structuredSessionSchemaVersion     = "workflow_run_store_sqlite_v18"
	structuredEvaluationMigrationID    = "0019_application_evaluation_structured_inputs"
	structuredEvaluationSchemaVersion  = "workflow_run_store_sqlite_v19"
	toolDefinitionSourcesMigrationID   = "0020_workflow_http_tool_definition_sources"
	toolDefinitionSourcesSchemaVersion = "workflow_run_store_sqlite_v20"
	definitionHTTPToolMigrationID      = "0021_workflow_definition_http_tool_execution"
	definitionHTTPToolSchemaVersion    = "workflow_run_store_sqlite_v21"
	resultArtifactMigrationID          = "0022_application_result_artifacts"
	resultArtifactSchemaVersion        = "workflow_run_store_sqlite_v22"
	resultArtifactLifecycleMigrationID = "0023_application_result_artifact_lifecycle"
	resultArtifactLifecycleVersion     = "workflow_run_store_sqlite_v23"
	resultArtifactHistoryMigrationID   = "0024_application_result_artifact_application_history"
	resultArtifactHistoryVersion       = "workflow_run_store_sqlite_v24"
	actionSafetyMigrationID            = "0025_action_safety_snapshots"
	actionSafetySchemaVersion          = "workflow_run_store_sqlite_v25"
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

//go:embed 0013_prompt_application_runtime_projections.up.sql
var upSQLV13 string

//go:embed 0014_agent_copilot_runtime_assignments.up.sql
var upSQLV14 string

//go:embed 0015_agent_copilot_invocation_projections.up.sql
var upSQLV15 string

//go:embed 0016_application_evaluation_campaigns.up.sql
var upSQLV16 string

//go:embed 0017_workflow_definition_structured_inputs.up.sql
var upSQLV17 string

//go:embed 0018_application_structured_sessions.up.sql
var upSQLV18 string

//go:embed 0019_application_evaluation_structured_inputs.up.sql
var upSQLV19 string

//go:embed 0020_workflow_http_tool_definition_sources.up.sql
var upSQLV20 string

//go:embed 0021_workflow_definition_http_tool_execution.up.sql
var upSQLV21 string

//go:embed 0022_application_result_artifacts.up.sql
var upSQLV22 string

//go:embed 0023_application_result_artifact_lifecycle.up.sql
var upSQLV23 string

//go:embed 0024_application_result_artifact_application_history.up.sql
var upSQLV24 string

//go:embed 0025_action_safety_snapshots.up.sql
var upSQLV25 string

//go:embed 0026_application_evaluation_schedules.up.sql
var upSQLV26 string

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
		{Component: Component, ID: applicationSessionMigrationID, StoreSchemaVersion: applicationSessionSchemaVersion, UpSQL: upSQLV12},
		{Component: Component, ID: promptRuntimeMigrationID, StoreSchemaVersion: promptRuntimeSchemaVersion, UpSQL: upSQLV13},
		{Component: Component, ID: agentRuntimeMigrationID, StoreSchemaVersion: agentRuntimeSchemaVersion, UpSQL: upSQLV14},
		{Component: Component, ID: agentInvocationMigrationID, StoreSchemaVersion: agentInvocationSchemaVersion, UpSQL: upSQLV15},
		{Component: Component, ID: applicationEvaluationMigrationID, StoreSchemaVersion: applicationEvaluationSchemaVersion, UpSQL: upSQLV16},
		{Component: Component, ID: structuredDefinitionMigrationID, StoreSchemaVersion: structuredDefinitionSchemaVersion, UpSQL: upSQLV17},
		{Component: Component, ID: structuredSessionMigrationID, StoreSchemaVersion: structuredSessionSchemaVersion, UpSQL: upSQLV18},
		{Component: Component, ID: structuredEvaluationMigrationID, StoreSchemaVersion: structuredEvaluationSchemaVersion, UpSQL: upSQLV19},
		{Component: Component, ID: toolDefinitionSourcesMigrationID, StoreSchemaVersion: toolDefinitionSourcesSchemaVersion, UpSQL: upSQLV20},
		{Component: Component, ID: definitionHTTPToolMigrationID, StoreSchemaVersion: definitionHTTPToolSchemaVersion, UpSQL: upSQLV21},
		{Component: Component, ID: resultArtifactMigrationID, StoreSchemaVersion: resultArtifactSchemaVersion, UpSQL: upSQLV22},
		{Component: Component, ID: resultArtifactLifecycleMigrationID, StoreSchemaVersion: resultArtifactLifecycleVersion, UpSQL: upSQLV23},
		{Component: Component, ID: resultArtifactHistoryMigrationID, StoreSchemaVersion: resultArtifactHistoryVersion, UpSQL: upSQLV24},
		{Component: Component, ID: actionSafetyMigrationID, StoreSchemaVersion: actionSafetySchemaVersion, UpSQL: upSQLV25},
		{Component: Component, ID: MigrationID, StoreSchemaVersion: StoreSchemaVersion, UpSQL: upSQLV26},
	}
}
