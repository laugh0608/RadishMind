//go:build postgres_integration

package httpapi

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	controlplanereadmigrations "radishmind.local/services/platform/migrations/control_plane_admin_read"
)

func TestControlPlaneAdminReadPostgresLifecycle(t *testing.T) {
	databaseURL := postgresIntegrationDatabaseURL(t)
	runtimeUser := strings.TrimSpace(os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER"))
	if runtimeUser == "" {
		t.Fatal("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER is required")
	}
	runtimeDatabaseURL := postgresIntegrationDatabaseURLForCredentials(t, runtimeUser, os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_PASSWORD"))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	adminPool, err := controlplanereadmigrations.OpenPool(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open control plane read test database: %v", err)
	}
	defer adminPool.Close()
	assertPostgresIntegrationDatabaseIsDisposable(t, ctx, adminPool)
	_, _ = adminPool.Exec(ctx, "DROP TABLE IF EXISTS control_plane_audit_summary_projections, control_plane_tenant_summary_projections, control_plane_read_schema_versions")
	state, err := controlplanereadmigrations.Apply(ctx, adminPool)
	if err != nil || state.MigrationState != controlplanereadmigrations.MigrationStateApplied {
		t.Fatalf("apply control plane read migration: %#v %v", state, err)
	}
	configureControlPlaneRuntimeReadPrivileges(t, ctx, adminPool, runtimeUser)
	if _, err := adminPool.Exec(ctx, `INSERT INTO control_plane_tenant_summary_projections
        (tenant_ref, schema_version, projection_version, tenant_display_name, tenant_state, plan_ref,
         quota_summary_ref, deployment_status_ref, audit_summary_ref, projected_at)
        VALUES ('tenant_demo',$1,1,'Demo tenant','active','plan_internal','quota_demo','deployment_demo','audit_latest',now())`,
		controlplanereadmigrations.TenantProjectionSchemaVersion); err != nil {
		t.Fatalf("seed tenant projection: %v", err)
	}
	for _, row := range []struct {
		auditRef, tenantRef, eventKind, resourceRef, recordedAt string
	}{
		{"audit_003", "tenant_demo", "read", "application:demo", "2026-07-12T10:03:00Z"},
		{"audit_002", "tenant_demo", "review", "application:demo", "2026-07-12T10:02:00Z"},
		{"audit_001", "tenant_demo", "read", "tenant:demo", "2026-07-12T10:01:00Z"},
		{"audit_other", "tenant_other", "read", "application:other", "2026-07-12T10:04:00Z"},
	} {
		if _, err := adminPool.Exec(ctx, `INSERT INTO control_plane_audit_summary_projections
            (tenant_ref, audit_ref, schema_version, projection_version, actor_subject_ref, event_kind,
             resource_ref, decision, failure_code, trace_id, recorded_at, projected_at)
            VALUES ($1,$2,$3,1,'subject_admin',$4,$5,'allowed',NULL,'trace_' || $2,$6,now())`,
			row.tenantRef, row.auditRef, controlplanereadmigrations.AuditProjectionSchemaVersion, row.eventKind, row.resourceRef, row.recordedAt); err != nil {
			t.Fatalf("seed audit projection: %v", err)
		}
	}
	runtimePool, err := controlplanereadmigrations.OpenPool(ctx, runtimeDatabaseURL)
	if err != nil {
		t.Fatalf("open control plane runtime pool: %v", err)
	}
	defer runtimePool.Close()
	if _, err := controlplanereadmigrations.PreflightRuntime(ctx, runtimePool); err != nil {
		t.Fatalf("runtime preflight: %v", err)
	}
	repository := &postgresControlPlaneAdminReadRepository{pool: runtimePool, timeout: 5 * time.Second}
	repositoryContext := ReadRepositoryContext{RequestContext: ctx, TenantRef: "tenant_demo", AuditRef: "audit_request"}
	if result := repository.ReadTenantSummary(repositoryContext, ReadTenantSummaryRequest{}); len(result.Items) != 1 || result.Items[0].TenantRef != "tenant_demo" {
		t.Fatalf("read tenant projection: %#v", result)
	}
	first := repository.ListAuditSummaries(repositoryContext, ListAuditSummariesRequest{ReadRepositoryRequest: ReadRepositoryRequest{Limit: 2, Sort: "recorded_at_desc"}})
	if len(first.Items) != 2 || first.NextCursor == nil || first.Items[0].AuditRef != "audit_003" {
		t.Fatalf("read first audit page: %#v", first)
	}
	second := repository.ListAuditSummaries(repositoryContext, ListAuditSummariesRequest{ReadRepositoryRequest: ReadRepositoryRequest{Limit: 2, Sort: "recorded_at_desc", Cursor: *first.NextCursor}})
	if len(second.Items) != 1 || second.Items[0].AuditRef != "audit_001" || second.NextCursor != nil {
		t.Fatalf("read second audit page: %#v", second)
	}
	filtered := repository.ListAuditSummaries(repositoryContext, ListAuditSummariesRequest{ReadRepositoryRequest: ReadRepositoryRequest{Limit: 10, Filters: ReadRepositoryFilters{"event_kind": "review"}, Sort: "recorded_at_desc"}})
	if len(filtered.Items) != 1 || filtered.Items[0].AuditRef != "audit_002" {
		t.Fatalf("filter audit projection: %#v", filtered)
	}
	for _, statement := range []string{
		"INSERT INTO control_plane_tenant_summary_projections (tenant_ref) VALUES ('forbidden')",
		"UPDATE control_plane_tenant_summary_projections SET tenant_state='forbidden' WHERE tenant_ref='tenant_demo'",
		"DELETE FROM control_plane_tenant_summary_projections WHERE tenant_ref='tenant_demo'",
		"TRUNCATE control_plane_tenant_summary_projections",
		"CREATE TABLE control_plane_runtime_forbidden (id bigint)",
		"ALTER TABLE control_plane_tenant_summary_projections ADD COLUMN forbidden text",
		"DROP TABLE control_plane_tenant_summary_projections",
	} {
		if _, err := runtimePool.Exec(ctx, statement); err == nil {
			t.Fatalf("runtime role unexpectedly executed write or DDL: %s", statement)
		} else {
			var postgresError *pgconn.PgError
			if !errors.As(err, &postgresError) || postgresError.Code != "42501" {
				t.Fatalf("runtime denial did not return insufficient_privilege for %s", statement)
			}
		}
	}
	if _, err := controlplanereadmigrations.RollbackForDevTest(ctx, adminPool); err != nil {
		t.Fatalf("rollback control plane read migration: %v", err)
	}
	if _, err := controlplanereadmigrations.Apply(ctx, adminPool); err != nil {
		t.Fatalf("reapply control plane read migration: %v", err)
	}
	configureControlPlaneRuntimeReadPrivileges(t, ctx, adminPool, runtimeUser)
}

func configureControlPlaneRuntimeReadPrivileges(t *testing.T, ctx context.Context, adminPool interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}, runtimeUser string) {
	t.Helper()
	quotedRuntimeUser := pgx.Identifier{runtimeUser}.Sanitize()
	for _, statement := range []string{
		"REVOKE ALL ON control_plane_read_schema_versions, control_plane_tenant_summary_projections, control_plane_audit_summary_projections FROM " + quotedRuntimeUser,
		"GRANT SELECT ON control_plane_read_schema_versions, control_plane_tenant_summary_projections, control_plane_audit_summary_projections TO " + quotedRuntimeUser,
	} {
		if _, err := adminPool.Exec(ctx, statement); err != nil {
			t.Fatalf("configure control plane runtime read privileges: %v", err)
		}
	}
}
