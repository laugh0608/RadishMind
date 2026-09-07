//go:build postgres_integration

package httpapi

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	adminproviderroutemigrations "radishmind.local/services/platform/migrations/admin_provider_routes"
)

func TestPostgresConfiguredWorkflowTemplateProductChainRestartNoFallbackAndHistory(t *testing.T) {
	adminPool, runtimeDatabaseURL, ctx := newConfiguredPostgresTestDatabase(t)
	for _, gate := range configuredPostgresMigrationGates(adminPool) {
		state, _, err := gate.apply(ctx)
		if err != nil || state != "applied" {
			t.Fatalf("apply %s migration for configured workflow template chain: state=%s err=%v", gate.name, state, err)
		}
	}
	if _, err := adminproviderroutemigrations.RollbackForDevTest(ctx, adminPool); err != nil {
		t.Fatalf("reset Admin Provider Route migration for workflow template chain: %v", err)
	}
	state, err := adminproviderroutemigrations.Apply(ctx, adminPool)
	if err != nil || state.MigrationState != adminproviderroutemigrations.MigrationStateApplied {
		t.Fatalf("apply Admin Provider Route migration for workflow template chain: state=%s err=%v", state.MigrationState, err)
	}
	t.Cleanup(func() {
		cleanupContext, cancel := contextWithWorkflowTemplateCleanupTimeout()
		defer cancel()
		_, _ = adminproviderroutemigrations.RollbackForDevTest(cleanupContext, adminPool)
	})

	runtimeUser := strings.TrimSpace(os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER"))
	constrainPostgresWorkflowTemplateHistoryPrivileges(t, ctx, adminPool, runtimeUser)

	cfg := workflowTemplateProductConfig(configuredPostgresProductConfig(runtimeDatabaseURL))
	cfg.AdminProviderRouteStoreMode = adminProviderRouteStoreModePostgresDevTest
	cfg.AdminProviderRouteDatabaseURL = runtimeDatabaseURL
	cfg.AdminProviderRouteDatabaseTimeout = time.Second
	first, err := NewServerWithError(cfg, Options{BuildVersion: "postgres-workflow-template-product-first"})
	if err != nil {
		t.Fatalf("start configured PostgreSQL workflow template product server: %v", err)
	}
	if first.localPersistenceRuntime != nil {
		t.Fatal("configured PostgreSQL workflow template product selected local SQLite persistence")
	}
	repository, ok := first.workflowTemplateCatalogRepository.(*postgresWorkflowTemplateCatalogRepository)
	if !ok {
		t.Fatalf("configured workflow template catalog did not select PostgreSQL: %T", first.workflowTemplateCatalogRepository)
	}
	evidence := exerciseWorkflowTemplateProductHTTPChain(t, first)
	assertWorkflowTemplateProductSideEffects(t, first, evidence.ProviderBridge)
	assertPostgresWorkflowTemplateProductRows(t, repository)
	assertPostgresWorkflowTemplateHistoryIsAppendOnly(t, ctx, repository.pool)

	closedRepository := first.workflowTemplateCatalogRepository
	first.Close()
	if _, err := closedRepository.ReadLineage(workflowTemplateProductContext(), workflowTemplateTestTemplate); !errors.Is(err, errWorkflowTemplateStoreUnavailable) {
		t.Fatalf("closed PostgreSQL template repository silently fell back: %v", err)
	}

	restarted, err := NewServerWithError(cfg, Options{BuildVersion: "postgres-workflow-template-product-restarted"})
	if err != nil {
		t.Fatalf("restart configured PostgreSQL workflow template product server: %v", err)
	}
	defer restarted.Close()
	attachWorkflowTemplateProductInventory(restarted)
	assertWorkflowTemplateProductRestored(t, restarted, evidence)
}

func assertPostgresWorkflowTemplateProductRows(t *testing.T, repository *postgresWorkflowTemplateCatalogRepository) {
	t.Helper()
	for table, expected := range map[string]int{
		"workflow_template_candidates": 1, "workflow_template_decisions": 1, "workflow_template_versions": 1,
		"workflow_template_lineages": 1, "workflow_template_listing_events": 1, "workflow_template_audits": 4,
		"workflow_run_records": 0, "workflow_http_tool_confirmation_decisions": 0, "workflow_rag_execution_audits": 0,
	} {
		var count int
		if err := repository.pool.QueryRow(workflowTemplateProductContext().RequestContext, "SELECT count(*) FROM "+table).Scan(&count); err != nil || count != expected {
			t.Fatalf("unexpected PostgreSQL product row count: table=%s count=%d expected=%d err=%v", table, count, expected, err)
		}
	}
}

func contextWithWorkflowTemplateCleanupTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 20*time.Second)
}
