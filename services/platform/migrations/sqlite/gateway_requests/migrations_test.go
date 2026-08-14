package gatewayrequests

import (
	"context"
	"path/filepath"
	"testing"

	"radishmind.local/services/platform/internal/sqlitedev"
)

func TestGatewayRequestSQLiteMigrationUpgradesV2DataToV3(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "gateway-request-upgrade.db")
	legacyMigrations := Migrations()[:2]
	legacy, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: databasePath,
		Migrations:   legacyMigrations,
	})
	if err != nil {
		t.Fatalf("open Gateway request v2 database: %v", err)
	}
	if _, err = legacy.DB().ExecContext(context.Background(), `INSERT INTO gateway_request_records (
        tenant_ref, workspace_id, consumer_ref, application_id, request_id, record_version, schema_version,
        store_mode, request_route, protocol, request_status, started_at_unix_nano, completed_at_unix_nano,
        selected_provider, selected_profile, selected_model, failure_boundary, usage_availability,
        cost_availability, sanitized_request_record
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"tenant_demo", "workspace_demo", "consumer_demo", "", "request_legacy_v2", 1,
		"gateway_request_record.v2", "sqlite_dev", "/v1/chat/completions", "openai-chat-completions",
		"started", int64(1_800_000_000_000_000_000), nil, "mock", "profile", "model", "",
		"not_reported", "usage_not_reported", `{"schema_version":"gateway_request_record.v2"}`,
	); err != nil {
		t.Fatalf("seed Gateway request v2 data: %v", err)
	}
	if err = legacy.Close(); err != nil {
		t.Fatalf("close Gateway request v2 database: %v", err)
	}

	upgraded, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: databasePath,
		Migrations:   Migrations(),
	})
	if err != nil {
		t.Fatalf("upgrade Gateway request database to v3: %v", err)
	}
	t.Cleanup(func() { _ = upgraded.Close() })
	var schemaVersion string
	var attemptCount int
	var fallbackUsed bool
	var terminalProvider string
	var terminalProfile string
	if err = upgraded.DB().QueryRowContext(context.Background(), `SELECT schema_version,
        provider_attempt_count, fallback_used, terminal_provider, terminal_profile
        FROM gateway_request_records WHERE request_id=?`, "request_legacy_v2").
		Scan(&schemaVersion, &attemptCount, &fallbackUsed, &terminalProvider, &terminalProfile); err != nil {
		t.Fatalf("read upgraded Gateway request data: %v", err)
	}
	if schemaVersion != "gateway_request_record.v2" || attemptCount != 0 || fallbackUsed ||
		terminalProvider != "" || terminalProfile != "" {
		t.Fatalf("Gateway request v2 upgrade drifted: schema=%s attempts=%d fallback=%v provider=%s profile=%s",
			schemaVersion, attemptCount, fallbackUsed, terminalProvider, terminalProfile)
	}
}
