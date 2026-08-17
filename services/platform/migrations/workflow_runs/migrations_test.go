package workflowrunmigrations

import (
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestEmbeddedWorkflowRunMigration(t *testing.T) {
	for _, required := range []string{
		"CREATE TABLE workflow_run_records",
		"failure_boundary",
		"CREATE TABLE workflow_evaluation_cases",
		"CREATE TABLE workflow_evaluation_case_revisions",
		"ADD COLUMN current_version",
		"CREATE TABLE workflow_evaluation_suites",
		"CREATE TABLE workflow_evaluation_suite_decisions",
		"CREATE TABLE workflow_http_tool_action_plans",
		"CREATE TABLE workflow_http_tool_confirmation_decisions",
		"CREATE TABLE workflow_http_tool_execution_audits",
		"CREATE TABLE workflow_http_tool_execution_attempts",
		"CREATE TABLE workflow_rag_snapshot_resources",
		"CREATE TABLE workflow_rag_snapshot_versions",
		"CREATE TABLE workflow_rag_snapshot_fragments",
		"CREATE TABLE workflow_rag_execution_audits",
		"outcome_unknown",
		"workflow_http_tool_confirmation_decisions_append_only",
		"workflow_http_tool_execution_audits_append_only",
		"workflow_rag_snapshot_versions_append_only",
		"workflow_rag_snapshot_fragments_append_only",
		"workflow_rag_execution_audits_append_only",
		"retrieval_started",
		"CREATE TABLE workflow_rag_evaluation_dataset_resources",
		"CREATE TABLE workflow_rag_evaluation_dataset_versions",
		"CREATE TABLE workflow_rag_candidate_snapshot_reviews",
		"CREATE TABLE workflow_rag_evaluation_audits",
		"workflow_rag_candidate_snapshot_reviews_append_only",
		"CREATE TABLE workflow_rag_knowledge_promotion_candidates",
		"CREATE TABLE workflow_rag_knowledge_promotion_decisions",
		"CREATE TABLE workflow_rag_application_bindings",
		"CREATE TABLE workflow_rag_knowledge_promotion_audits",
		"workflow_rag_knowledge_promotion_decisions_append_only",
		"workflow_rag_application_bindings_append_only",
		"workflow_rag_knowledge_promotion_audits_append_only",
		"execution_source_kind",
		"workflow_run_record.v4",
		"CREATE TABLE workflow_rag_application_runtime_assignments",
		"CREATE TABLE workflow_rag_application_runtime_events",
		"CREATE TABLE workflow_rag_application_runtime_audits",
		"workflow_rag_application_runtime_events_append_only",
		"workflow_rag_application_runtime_audits_append_only",
		"CREATE TABLE workflow_definition_release_candidates",
		"CREATE TABLE workflow_definition_release_decisions",
		"CREATE TABLE workflow_definition_versions",
		"CREATE TABLE workflow_definition_activations",
		"CREATE TABLE workflow_definition_activation_events",
		"CREATE TABLE workflow_definition_release_audits",
		"workflow_definition_release_decisions_append_only",
		"workflow_definition_release_audits_append_only",
		"workflow_run_record.v5",
		"workflow_definition",
		"CREATE TABLE application_interaction_sessions",
		"CREATE TABLE application_interaction_session_turns",
		"application_interaction_sessions_controlled_update",
		"application_interaction_turns_no_delete",
		"CREATE TABLE prompt_application_runtime_assignments",
		"CREATE TABLE prompt_application_runtime_assignment_events",
		"CREATE TABLE prompt_application_sessions",
		"CREATE TABLE prompt_application_session_turns",
		"CREATE TABLE prompt_application_run_records",
		"workflow_run_record.v6",
		"CREATE TABLE agent_copilot_runtime_assignments",
		"CREATE TABLE agent_copilot_runtime_assignment_events",
		"agent_copilot_assignments_controlled_update",
		"CREATE TABLE agent_copilot_sessions",
		"CREATE TABLE agent_copilot_session_turns",
		"CREATE TABLE agent_copilot_run_records",
		"workflow_run_record.v7",
		"agent_copilot_runs_controlled_update",
		"workflow_run_record.v8",
		"input_contract_id",
		"input_contract_digest",
		"workflow_run_records_structured_input_projection_check",
		"application_session.v4",
		"application_session_turn.v4",
		"application_runtime_authority.v4",
		"application_interaction_sessions_payload_v4_check",
		"application_evaluation_plan.v2",
		"application_evaluation_plan_version.v2",
		"application_evaluation_campaign.v2",
		"application_evaluation_campaigns_payload_v2_check",
		"workflow_http_tool_action_plan.v2",
		"workflow_http_tool_confirmation_decision.v2",
		"workflow_http_tool_execution_audit.v2",
		"workflow_http_tool_action_plans_source_union_check",
		"workflow_http_tool_action_plans_definition_idx",
		"workflow_run_record.v9",
		"CREATE TABLE application_result_artifacts",
		"application_result_artifacts_session_history_idx",
		"application_result_artifacts_append_only",
		"CREATE TABLE application_result_artifact_lifecycles",
		"CREATE TABLE application_result_artifact_lifecycle_events",
		"application_result_artifact_lifecycles_controlled_mutation",
		"application_result_artifact_lifecycle_events_append_only",
	} {
		if !strings.Contains(upSQL, required) {
			t.Fatalf("workflow run up migration is missing %q", required)
		}
	}
	for _, required := range []string{
		"DROP TABLE IF EXISTS workflow_http_tool_execution_attempts",
		"DROP TABLE IF EXISTS workflow_rag_execution_audits",
		"DROP TABLE IF EXISTS workflow_rag_snapshot_fragments",
		"DROP TABLE IF EXISTS workflow_rag_snapshot_versions",
		"DROP TABLE IF EXISTS workflow_rag_snapshot_resources",
		"DROP TABLE IF EXISTS workflow_rag_evaluation_audits",
		"DROP TABLE IF EXISTS workflow_rag_candidate_snapshot_reviews",
		"DROP TABLE IF EXISTS workflow_rag_evaluation_dataset_versions",
		"DROP TABLE IF EXISTS workflow_rag_evaluation_dataset_resources",
		"DROP TABLE IF EXISTS workflow_rag_knowledge_promotion_audits",
		"DROP TABLE IF EXISTS workflow_rag_application_bindings",
		"DROP TABLE IF EXISTS workflow_rag_knowledge_promotion_decisions",
		"DROP TABLE IF EXISTS workflow_rag_knowledge_promotion_candidates",
		"DROP TABLE IF EXISTS workflow_rag_application_runtime_audits",
		"DROP TABLE IF EXISTS workflow_rag_application_runtime_events",
		"DROP TABLE IF EXISTS workflow_rag_application_runtime_assignments",
		"DROP TABLE IF EXISTS workflow_definition_release_audits",
		"DROP TABLE IF EXISTS workflow_definition_activation_events",
		"DROP TABLE IF EXISTS workflow_definition_activations",
		"DROP TABLE IF EXISTS workflow_definition_versions",
		"DROP TABLE IF EXISTS workflow_definition_release_decisions",
		"DROP TABLE IF EXISTS workflow_definition_release_candidates",
		"DROP TABLE IF EXISTS application_interaction_session_turns",
		"DROP TABLE IF EXISTS application_interaction_sessions",
		"DROP TABLE IF EXISTS prompt_application_run_records",
		"DROP TABLE IF EXISTS prompt_application_runtime_assignments",
		"DROP TABLE IF EXISTS agent_copilot_runtime_assignment_events",
		"DROP TABLE IF EXISTS agent_copilot_runtime_assignments",
		"DROP TABLE IF EXISTS agent_copilot_run_records",
		"DROP TABLE IF EXISTS agent_copilot_session_turns",
		"DROP TABLE IF EXISTS agent_copilot_sessions",
		"DROP TABLE IF EXISTS workflow_http_tool_confirmation_decisions",
		"DROP TABLE IF EXISTS workflow_http_tool_execution_audits",
		"DROP TABLE IF EXISTS workflow_http_tool_action_plans",
		"DROP TABLE IF EXISTS workflow_evaluation_suite_decisions",
		"DROP TABLE IF EXISTS workflow_evaluation_suites",
		"DROP TABLE IF EXISTS workflow_evaluation_case_revisions",
		"DROP TABLE IF EXISTS workflow_evaluation_cases",
		"DROP TABLE IF EXISTS workflow_run_records",
		"DROP CONSTRAINT application_interaction_sessions_payload_v4_check",
		"DROP CONSTRAINT application_evaluation_campaigns_payload_v2_check",
		"DROP CONSTRAINT workflow_http_tool_action_plans_source_union_check",
		"DELETE FROM workflow_run_records",
		"DROP TABLE IF EXISTS application_result_artifacts",
		"DROP FUNCTION IF EXISTS reject_application_result_artifact_mutation",
		"DROP TABLE IF EXISTS application_result_artifact_lifecycle_events",
		"DROP TABLE IF EXISTS application_result_artifact_lifecycles",
		"DROP FUNCTION IF EXISTS validate_application_result_artifact_lifecycle_mutation",
		"DROP FUNCTION IF EXISTS reject_application_result_artifact_lifecycle_event_mutation",
	} {
		if !strings.Contains(downSQL, required) {
			t.Fatalf("workflow run down migration is missing %q", required)
		}
	}
	if checksum := ExpectedChecksum(); !strings.HasPrefix(checksum, "sha256:") || len(checksum) != 71 {
		t.Fatalf("workflow run migration checksum is invalid: %s", checksum)
	}
	for _, forbidden := range []string{
		"workflow_http_tool_execution_attempts",
		"ALTER TABLE workflow_run_records",
		"workflow_run_record.v2",
	} {
		if strings.Contains(upSQLV6, forbidden) {
			t.Fatalf("batch A migration contains batch B storage %q", forbidden)
		}
	}
	if count := strings.Count(upSQLV6, "tool_version integer NOT NULL CHECK (tool_version = 1)"); count != 3 {
		t.Fatalf("PostgreSQL HTTP tool action migration must pin tool_version=1 on all three tables, got %d", count)
	}
}

func TestWorkflowRunPendingMigrationPaths(t *testing.T) {
	testCases := []struct {
		name              string
		migrationID       string
		requiredFragment  string
		forbiddenFragment string
	}{
		{name: "v1", migrationID: legacyMigrationID, requiredFragment: "ADD COLUMN failure_code"},
		{name: "v2", migrationID: diagnosticsMigrationID, requiredFragment: "CREATE TABLE workflow_evaluation_cases"},
		{name: "v3", migrationID: evaluationMigrationID, requiredFragment: "ADD COLUMN current_version"},
		{name: "v4", migrationID: caseVersioningMigrationID, requiredFragment: "CREATE TABLE workflow_evaluation_suites"},
		{name: "v5", migrationID: evaluationSuiteMigrationID, requiredFragment: "CREATE TABLE workflow_http_tool_action_plans", forbiddenFragment: "CREATE TABLE workflow_evaluation_suites"},
		{name: "v6", migrationID: toolActionsMigrationID, requiredFragment: "CREATE TABLE workflow_http_tool_execution_attempts", forbiddenFragment: "CREATE TABLE workflow_http_tool_action_plans"},
		{name: "v7", migrationID: toolExecutionMigrationID, requiredFragment: "CREATE TABLE workflow_rag_snapshot_resources", forbiddenFragment: "CREATE TABLE workflow_http_tool_execution_attempts"},
		{name: "v8", migrationID: ragSnapshotMigrationID, requiredFragment: "retrieval_started", forbiddenFragment: "CREATE TABLE workflow_rag_snapshot_resources"},
		{name: "v9", migrationID: ragExecutionAuditMigrationID, requiredFragment: "CREATE TABLE workflow_rag_evaluation_dataset_resources", forbiddenFragment: "retrieval_started"},
		{name: "v10", migrationID: ragEvaluationDatasetMigrationID, requiredFragment: "CREATE TABLE workflow_rag_knowledge_promotion_candidates", forbiddenFragment: "CREATE TABLE workflow_rag_evaluation_dataset_resources"},
		{name: "v11", migrationID: ragKnowledgePromotionMigrationID, requiredFragment: "CREATE TABLE workflow_rag_application_runtime_assignments", forbiddenFragment: "CREATE TABLE workflow_rag_knowledge_promotion_candidates"},
		{name: "v12", migrationID: applicationRuntimeMigrationID, requiredFragment: "CREATE TABLE workflow_definition_release_candidates", forbiddenFragment: "CREATE TABLE workflow_rag_application_runtime_assignments"},
		{name: "v13", migrationID: definitionReleaseMigrationID, requiredFragment: "workflow_run_record.v5", forbiddenFragment: "CREATE TABLE workflow_definition_release_candidates"},
		{name: "v14", migrationID: definitionExecutionMigrationID, requiredFragment: "CREATE TABLE application_interaction_sessions", forbiddenFragment: "CREATE TABLE workflow_definition_release_candidates"},
		{name: "v15", migrationID: applicationSessionMigrationID, requiredFragment: "CREATE TABLE prompt_application_runtime_assignments", forbiddenFragment: "CREATE TABLE application_interaction_sessions"},
		{name: "v16", migrationID: promptRuntimeMigrationID, requiredFragment: "CREATE TABLE agent_copilot_runtime_assignments", forbiddenFragment: "CREATE TABLE prompt_application_runtime_assignments"},
		{name: "v17", migrationID: agentRuntimeMigrationID, requiredFragment: "CREATE TABLE agent_copilot_sessions", forbiddenFragment: "CREATE TABLE agent_copilot_runtime_assignments"},
		{name: "v18", migrationID: agentInvocationMigrationID, requiredFragment: "CREATE TABLE application_evaluation_plans", forbiddenFragment: "CREATE TABLE agent_copilot_sessions"},
		{name: "v19", migrationID: applicationEvaluationMigrationID, requiredFragment: "input_contract_id", forbiddenFragment: "CREATE TABLE application_evaluation_plans"},
		{name: "v20", migrationID: structuredDefinitionMigrationID, requiredFragment: "application_session.v4", forbiddenFragment: "ADD COLUMN input_contract_id"},
		{name: "v21", migrationID: structuredSessionMigrationID, requiredFragment: "application_evaluation_plan.v2", forbiddenFragment: "application_session.v4"},
		{name: "v22", migrationID: structuredEvaluationMigrationID, requiredFragment: "workflow_http_tool_action_plan.v2", forbiddenFragment: "application_evaluation_plan.v2"},
		{name: "v23", migrationID: toolDefinitionSourcesMigrationID, requiredFragment: "workflow_run_record.v9", forbiddenFragment: "workflow_http_tool_action_plan.v2"},
		{name: "v24", migrationID: definitionHTTPToolExecutionMigrationID, requiredFragment: "CREATE TABLE application_result_artifacts", forbiddenFragment: "workflow_run_record.v9"},
		{name: "v25", migrationID: resultArtifactMigrationID, requiredFragment: "CREATE TABLE application_result_artifact_lifecycles", forbiddenFragment: "CREATE TABLE application_result_artifacts"},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			pendingSQL := pendingMigrationSQL(testCase.migrationID)
			if !strings.Contains(pendingSQL, testCase.requiredFragment) {
				t.Fatalf("pending migration path is incomplete for %s", testCase.migrationID)
			}
			if testCase.forbiddenFragment != "" && strings.Contains(pendingSQL, testCase.forbiddenFragment) {
				t.Fatalf("pending migration path replays %q for %s", testCase.forbiddenFragment, testCase.migrationID)
			}
		})
	}
	if pendingMigrationSQL("0000_unknown") != "" {
		t.Fatal("unknown pending migration must fail closed")
	}
}

func TestWorkflowRunPendingRollbackPathsDoNotDropUnappliedTables(t *testing.T) {
	for _, migrationID := range []string{
		legacyMigrationID,
		diagnosticsMigrationID,
		evaluationMigrationID,
		caseVersioningMigrationID,
		evaluationSuiteMigrationID,
	} {
		rollbackSQL := rollbackSQLThrough(migrationID)
		if !strings.Contains(rollbackSQL, "DROP TABLE IF EXISTS workflow_run_records") {
			t.Fatalf("rollback path does not remove v1 for %s", migrationID)
		}
		if strings.Contains(rollbackSQL, "workflow_http_tool_action_plans") {
			t.Fatalf("pending rollback tries to remove unapplied v6 for %s", migrationID)
		}
		if strings.Contains(rollbackSQL, "workflow_http_tool_execution_attempts") {
			t.Fatalf("pending rollback tries to remove unapplied v7 for %s", migrationID)
		}
	}
	toolActionRollback := rollbackSQLThrough(toolActionsMigrationID)
	if !strings.Contains(toolActionRollback, "workflow_http_tool_action_plans") || strings.Contains(toolActionRollback, "workflow_http_tool_execution_attempts") {
		t.Fatalf("v6 rollback must remove applied action tables without removing unapplied v7: %s", toolActionRollback)
	}
	toolExecutionRollback := rollbackSQLThrough(toolExecutionMigrationID)
	if !strings.Contains(toolExecutionRollback, "workflow_http_tool_execution_attempts") || strings.Contains(toolExecutionRollback, "workflow_rag_snapshot_resources") {
		t.Fatalf("v7 rollback must remove applied execution tables without removing unapplied v8: %s", toolExecutionRollback)
	}
	ragExecutionAuditRollback := rollbackSQLThrough(ragExecutionAuditMigrationID)
	if !strings.Contains(ragExecutionAuditRollback, "retrieval_started") || strings.Contains(ragExecutionAuditRollback, "workflow_rag_evaluation_dataset_resources") {
		t.Fatalf("v9 rollback must remove applied execution audit changes without removing unapplied v10: %s", ragExecutionAuditRollback)
	}
	ragEvaluationDatasetRollback := rollbackSQLThrough(ragEvaluationDatasetMigrationID)
	if !strings.Contains(ragEvaluationDatasetRollback, "workflow_rag_evaluation_dataset_resources") || strings.Contains(ragEvaluationDatasetRollback, "workflow_rag_knowledge_promotion_candidates") {
		t.Fatalf("v10 rollback must remove applied evaluation dataset resources without removing unapplied v11: %s", ragEvaluationDatasetRollback)
	}
	ragKnowledgePromotionRollback := rollbackSQLThrough(ragKnowledgePromotionMigrationID)
	if !strings.Contains(ragKnowledgePromotionRollback, "workflow_rag_knowledge_promotion_candidates") || strings.Contains(ragKnowledgePromotionRollback, "workflow_rag_application_runtime_assignments") {
		t.Fatalf("v11 rollback must remove applied promotion resources without removing unapplied v12: %s", ragKnowledgePromotionRollback)
	}
	applicationRuntimeRollback := rollbackSQLThrough(applicationRuntimeMigrationID)
	if !strings.Contains(applicationRuntimeRollback, "workflow_rag_application_runtime_assignments") || strings.Contains(applicationRuntimeRollback, "workflow_definition_release_candidates") {
		t.Fatalf("v12 rollback must remove applied application runtime resources without removing unapplied v13: %s", applicationRuntimeRollback)
	}
	definitionReleaseRollback := rollbackSQLThrough(definitionReleaseMigrationID)
	if !strings.Contains(definitionReleaseRollback, "workflow_definition_release_candidates") || strings.Contains(definitionReleaseRollback, "workflow_run_record.v5") {
		t.Fatalf("v13 rollback must remove applied definition release resources without removing unapplied v14: %s", definitionReleaseRollback)
	}
	definitionExecutionRollback := rollbackSQLThrough(definitionExecutionMigrationID)
	if !strings.Contains(definitionExecutionRollback, "workflow_run_record.v5") || strings.Contains(definitionExecutionRollback, "application_interaction_sessions") {
		t.Fatalf("v14 rollback must remove applied definition execution resources without removing unapplied v15: %s", definitionExecutionRollback)
	}
	applicationSessionRollback := rollbackSQLThrough(applicationSessionMigrationID)
	if !strings.Contains(applicationSessionRollback, "application_interaction_sessions") || strings.Contains(applicationSessionRollback, "prompt_application_run_records") {
		t.Fatalf("v15 rollback must remove applied sessions without removing unapplied v16: %s", applicationSessionRollback)
	}
	promptRuntimeRollback := rollbackSQLThrough(promptRuntimeMigrationID)
	if !strings.Contains(promptRuntimeRollback, "prompt_application_run_records") || strings.Contains(promptRuntimeRollback, "agent_copilot_runtime_assignments") {
		t.Fatalf("v16 rollback must remove Prompt runtime without removing unapplied v17: %s", promptRuntimeRollback)
	}
	agentRuntimeRollback := rollbackSQLThrough(agentRuntimeMigrationID)
	if !strings.Contains(agentRuntimeRollback, "agent_copilot_runtime_assignments") || strings.Contains(agentRuntimeRollback, "agent_copilot_run_records") {
		t.Fatalf("v17 rollback must remove Agent Copilot assignments without removing unapplied v18: %s", agentRuntimeRollback)
	}
	agentInvocationRollback := rollbackSQLThrough(agentInvocationMigrationID)
	if !strings.Contains(agentInvocationRollback, "agent_copilot_run_records") || strings.Contains(agentInvocationRollback, "application_evaluation_plans") {
		t.Fatalf("v18 rollback must remove Agent Copilot invocation projections without removing unapplied v19: %s", agentInvocationRollback)
	}
	applicationEvaluationRollback := rollbackSQLThrough(applicationEvaluationMigrationID)
	if !strings.Contains(applicationEvaluationRollback, "application_evaluation_plans") || strings.Contains(applicationEvaluationRollback, "input_contract_id") {
		t.Fatalf("v19 rollback must remove application evaluation campaigns without removing unapplied v20: %s", applicationEvaluationRollback)
	}
	structuredDefinitionRollback := rollbackSQLThrough(structuredDefinitionMigrationID)
	if !strings.Contains(structuredDefinitionRollback, "input_contract_id") || strings.Contains(structuredDefinitionRollback, "application_interaction_sessions_payload_v4_check") {
		t.Fatalf("v20 rollback must remove structured Definition columns without removing unapplied v21: %s", structuredDefinitionRollback)
	}
	structuredSessionRollback := rollbackSQLThrough(structuredSessionMigrationID)
	if !strings.Contains(structuredSessionRollback, "application_interaction_sessions_payload_v4_check") || strings.Contains(structuredSessionRollback, "application_evaluation_campaigns_payload_v2_check") {
		t.Fatalf("v21 rollback must remove structured sessions without removing unapplied v22: %s", structuredSessionRollback)
	}
	structuredEvaluationRollback := rollbackSQLThrough(structuredEvaluationMigrationID)
	if !strings.Contains(structuredEvaluationRollback, "application_evaluation_campaigns_payload_v2_check") || strings.Contains(structuredEvaluationRollback, "workflow_http_tool_action_plans_source_union_check") {
		t.Fatalf("v22 rollback must remove structured evaluation without removing unapplied v23: %s", structuredEvaluationRollback)
	}
	toolDefinitionSourcesRollback := rollbackSQLThrough(toolDefinitionSourcesMigrationID)
	if !strings.Contains(toolDefinitionSourcesRollback, "workflow_http_tool_action_plans_source_union_check") || strings.Contains(toolDefinitionSourcesRollback, "workflow_run_record.v9") {
		t.Fatalf("v23 rollback must remove Definition tool sources without removing unapplied v24: %s", toolDefinitionSourcesRollback)
	}
	definitionHTTPToolRollback := rollbackSQLThrough(definitionHTTPToolExecutionMigrationID)
	if !strings.Contains(definitionHTTPToolRollback, "workflow_run_record.v9") || strings.Contains(definitionHTTPToolRollback, "application_result_artifacts") {
		t.Fatalf("v24 rollback must remove Definition tool execution without removing unapplied v25: %s", definitionHTTPToolRollback)
	}
	resultArtifactRollback := rollbackSQLThrough(resultArtifactMigrationID)
	if !strings.Contains(resultArtifactRollback, "application_result_artifacts") || strings.Contains(resultArtifactRollback, "application_result_artifact_lifecycles") {
		t.Fatalf("v25 rollback must remove result artifacts without removing unapplied v26 lifecycle tables: %s", resultArtifactRollback)
	}
	if rollbackSQLThrough("0000_unknown") != "" {
		t.Fatal("unknown pending rollback must fail closed")
	}
}

func TestWorkflowRunDatabaseErrorsAreSanitized(t *testing.T) {
	secret := "postgresql://user:secret@example.invalid/db"
	if got := safeDatabaseError("connect", errors.New(secret)).Error(); strings.Contains(got, "secret") {
		t.Fatalf("connection material leaked: %s", got)
	}
	if got := safeDatabaseError("query", &pgconn.PgError{Code: "23505", Message: secret}).Error(); got != "query failed (SQLSTATE 23505)" {
		t.Fatalf("unexpected PostgreSQL error: %s", got)
	}
}
