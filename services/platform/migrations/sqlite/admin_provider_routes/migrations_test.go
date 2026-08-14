package adminproviderroutes

import (
	"context"
	"path/filepath"
	"testing"

	"radishmind.local/services/platform/internal/sqlitedev"
)

func TestAdminProviderRouteSQLiteMigrationPreservesV1AndAcceptsV2(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "admin-provider-route-upgrade.db")
	legacy, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: databasePath,
		Migrations:   Migrations()[:1],
	})
	if err != nil {
		t.Fatalf("open admin provider route v1 database: %v", err)
	}
	v1Payload := `{"schema_version":"admin_provider_route_configuration_draft.v1","tenant_ref":"tenant_demo","workspace_id":"workspace_demo","environment":"test","configuration_id":"gateway-v1","draft_revision":1,"draft_digest":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`
	if _, err = legacy.DB().ExecContext(context.Background(), `INSERT INTO admin_provider_route_drafts (
        tenant_ref, workspace_id, environment, configuration_id, draft_revision, draft_digest,
        sanitized_draft_payload, updated_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, "tenant_demo", "workspace_demo", "test", "gateway-v1", 1,
		"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", v1Payload,
		"2026-08-13T10:00:00Z"); err != nil {
		t.Fatalf("seed admin provider route v1 draft: %v", err)
	}
	if err = legacy.Close(); err != nil {
		t.Fatalf("close admin provider route v1 database: %v", err)
	}

	upgraded, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: databasePath,
		Migrations:   Migrations(),
	})
	if err != nil {
		t.Fatalf("upgrade admin provider route database to v2: %v", err)
	}
	t.Cleanup(func() { _ = upgraded.Close() })
	var restoredPayload string
	if err = upgraded.DB().QueryRowContext(context.Background(), `SELECT sanitized_draft_payload
        FROM admin_provider_route_drafts WHERE configuration_id=?`, "gateway-v1").Scan(&restoredPayload); err != nil {
		t.Fatalf("read upgraded admin provider route v1 draft: %v", err)
	}
	if restoredPayload != v1Payload {
		t.Fatalf("admin provider route v1 payload drifted: %s", restoredPayload)
	}
	v2Payload := `{"schema_version":"admin_provider_route_configuration_draft.v2","tenant_ref":"tenant_demo","workspace_id":"workspace_demo","environment":"test","configuration_id":"gateway-v2","draft_revision":1,"draft_digest":"sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}`
	if _, err = upgraded.DB().ExecContext(context.Background(), `INSERT INTO admin_provider_route_drafts (
        tenant_ref, workspace_id, environment, configuration_id, draft_revision, draft_digest,
        sanitized_draft_payload, updated_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, "tenant_demo", "workspace_demo", "test", "gateway-v2", 1,
		"sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", v2Payload,
		"2026-08-13T10:01:00Z"); err != nil {
		t.Fatalf("insert admin provider route v2 draft after upgrade: %v", err)
	}
}
